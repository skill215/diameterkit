// Package codecengine is the single offline/runtime codec implementation.
package codecengine

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"io"
)

type Engine struct {
	bundle *schema.Bundle
	limits wire.Limits
}

func New(b *schema.Bundle, l wire.Limits) (*Engine, error) {
	if b == nil {
		return nil, schema.ErrDefinition
	}
	if err := l.Validate(); err != nil {
		return nil, err
	}
	return &Engine{b, l}, nil
}
func (e *Engine) Limits() wire.Limits { return e.limits }
func (e *Engine) operation(h pdu.Header) (schema.Command, error) {
	if h.Flags&pdu.RequestFlag != 0 && h.Flags&pdu.ErrorFlag != 0 || h.Flags&pdu.RequestFlag == 0 && h.Flags&pdu.RetransmittedFlag != 0 {
		return schema.Command{}, wire.ErrFrame
	}
	c, ok := e.bundle.Lookup(h.Key())
	if !ok {
		return c, validation.ErrUnsupported
	}
	if (h.Flags&pdu.ProxiableFlag != 0) != c.Proxiable {
		return c, wire.ErrFrame
	}
	if h.Flags&pdu.ErrorFlag != 0 {
		if c.Error == nil {
			return schema.Command{}, validation.ErrUnsupported
		}
		x := c.Error
		c.Decode = x.Decode
		c.Encode = x.Encode
		c.Validate = x.Validate
		c.Clone = x.Clone
		c.DecodeBound = x.DecodeBound
		c.Envelope = x.Envelope
		c.CompleteAnswer = x.CompleteAnswer
		c.CheckAnswer = nil // E grammar has no normal application body relation.
	}
	return c, nil
}
func (e *Engine) ValidateBody(b pdu.Body) error {
	if b == nil {
		return validation.ErrType
	}
	c, ok := e.bundle.Lookup(b.CommandKey())
	if !ok {
		return validation.ErrUnsupported
	}
	if _, isError := b.(*pdu.ErrorAnswer); isError {
		h := pdu.Header{CommandCode: c.Key.Code, ApplicationID: c.Key.ApplicationID, Flags: pdu.ErrorFlag}
		if c.Proxiable {
			h.Flags |= pdu.ProxiableFlag
		}
		var err error
		c, err = e.operation(h)
		if err != nil {
			return err
		}
	}
	if err := c.Validate(b, validation.Build); err != nil {
		return err
	}
	w := wire.NewEncoder(nil, e.limits)
	if err := c.Encode(w, b); err != nil {
		return err
	}
	if w.Err() != nil {
		return w.Err()
	}
	if w.Len() > e.limits.MaxFrameBytes-20 {
		return wire.ErrLimit
	}
	return nil
}

// MeasureBody bounds a Build-valid prototype before envelope completion clones it.
// It does not fill missing envelope fields or grant runtime admission.
func (e *Engine) MeasureBody(b pdu.Body) (int, error) {
	if err := e.ValidateBody(b); err != nil {
		return 0, err
	}
	c, _ := e.bundle.Lookup(b.CommandKey())
	encode := c.Encode
	if _, ok := b.(*pdu.ErrorAnswer); ok {
		encode = c.Error.Encode
	}
	w := wire.NewEncoder(nil, e.limits)
	if err := encode(w, b); err != nil {
		return 0, err
	}
	if err := w.Err(); err != nil {
		return 0, err
	}
	return 20 + w.Len(), nil
}

func (e *Engine) Validate(m pdu.Message) error { _, _, err := e.measure(m); return err }
func (e *Engine) measure(m pdu.Message) (schema.Command, int, error) {
	return e.measureMode(m, validation.Send)
}
func (e *Engine) measureMode(m pdu.Message, mode validation.Mode) (schema.Command, int, error) {
	if m.Body == nil {
		return schema.Command{}, 0, validation.ErrType
	}
	if mode != validation.Receive && m.Header.Flags&0x0f != 0 || m.Header.Flags&pdu.RequestFlag == 0 && m.Header.Flags&pdu.RetransmittedFlag != 0 || m.Header.Key() != m.Body.CommandKey() {
		return schema.Command{}, 0, wire.ErrFrame
	}
	c, err := e.operation(m.Header)
	if err != nil {
		return c, 0, err
	}
	if err = c.Validate(m.Body, mode); err != nil {
		return c, 0, err
	}
	w := wire.NewEncoder(nil, e.limits)
	if mode == validation.Receive {
		w = wire.NewReceiveSizer(e.limits)
	}
	if err = c.Encode(w, m.Body); err != nil {
		return c, 0, err
	}
	if err = w.Err(); err != nil {
		return c, 0, err
	}
	n := 20 + w.Len()
	if n > e.limits.MaxFrameBytes {
		return c, 0, wire.ErrLimit
	}
	return c, n, nil
}
func (e *Engine) Append(dst []byte, m pdu.Message) ([]byte, error) {
	c, n, err := e.measure(m)
	if err != nil {
		return dst, err
	}
	if n > int(^uint(0)>>1)-len(dst) {
		return dst, wire.ErrLimit
	}
	old := len(dst)
	dst = append(dst, make([]byte, n)...)
	wire.PutHeader(dst[old:old+20], m.Header, n)
	w := wire.NewEncoder(dst[old+20:], e.limits)
	err = c.Encode(w, m.Body)
	if err == nil {
		err = w.Err()
	}
	if err == nil && w.Len() != n-20 {
		err = wire.ErrFrame
	}
	if err != nil {
		return dst[:old], err
	}
	return dst, nil
}

// DecodeOwned adopts src. The caller must never mutate/reuse src after a successful
// call. Runtime reservations remain the caller's responsibility, not the engine's.
func (e *Engine) DecodeOwned(src []byte) (pdu.Message, error) {
	h, n, err := wire.Header(src, e.limits)
	if err != nil {
		return pdu.Message{}, &wire.DecodeError{Cause: err}
	}
	if n != len(src) {
		return pdu.Message{}, &wire.DecodeError{Cause: wire.ErrFrame}
	}
	c, err := e.operation(h)
	if err != nil {
		return pdu.Message{}, &wire.DecodeError{Offset: 20, FrameComplete: true, Cause: err}
	}
	if _, err = c.DecodeBound(n, e.limits); err != nil {
		return pdu.Message{}, err
	}
	d := wire.NewDecoder(src[20:], e.limits)
	b, err := c.Decode(d)
	if err == nil {
		err = c.Validate(b, validation.Receive)
	}
	if err == nil && (b == nil || b.CommandKey() != h.Key() || d.Remaining() != 0) {
		err = validation.ErrType
	}
	if err != nil {
		return pdu.Message{}, &wire.DecodeError{Offset: 20, FrameComplete: true, Cause: err}
	}
	return pdu.Message{Header: h, Body: b}, nil
}
func (e *Engine) Decode(src []byte) (pdu.Message, error) {
	_, n, err := wire.Header(src, e.limits)
	if err != nil {
		return pdu.Message{}, err
	}
	if n != len(src) {
		return pdu.Message{}, wire.ErrFrame
	}
	return e.DecodeOwned(append([]byte(nil), src...))
}
func (e *Engine) Read(r io.Reader) (pdu.Message, error) {
	b, err := ReadFrame(r, e.limits)
	if err != nil {
		return pdu.Message{}, err
	}
	return e.DecodeOwned(b)
}
func (e *Engine) CompleteRequest(b pdu.Body, env pdu.RequestEnvelope) (pdu.Message, error) {
	for _, v := range []string{env.OriginHost, env.OriginRealm, env.DestinationHost, env.DestinationRealm} {
		if v != "" && avp.Identity(v) != nil {
			return pdu.Message{}, validation.ErrEnvelope
		}
	}
	if err := e.ValidateBody(b); err != nil {
		return pdu.Message{}, err
	}
	c, _ := e.bundle.Lookup(b.CommandKey())
	if !c.Key.Request || c.CompleteRequest == nil {
		return pdu.Message{}, validation.ErrType
	}
	copy, err := c.CompleteRequest(b, env)
	if err != nil {
		return pdu.Message{}, err
	}
	flags := pdu.RequestFlag
	if c.Proxiable {
		flags |= pdu.ProxiableFlag
	}
	m := pdu.Message{Header: pdu.Header{Flags: flags, CommandCode: c.Key.Code, ApplicationID: c.Key.ApplicationID, HopByHop: env.HopByHop, EndToEnd: env.EndToEnd}, Body: copy}
	if err = e.Validate(m); err != nil {
		return pdu.Message{}, err
	}
	return m, nil
}
func ReadFrame(r io.Reader, l wire.Limits) ([]byte, error) {
	if err := l.Validate(); err != nil {
		return nil, err
	}
	var header [20]byte
	if _, err := io.ReadFull(r, header[:]); err != nil {
		return nil, err
	}
	_, n, err := wire.Header(header[:], l)
	if err != nil {
		return nil, err
	}
	b := make([]byte, n)
	copy(b, header[:])
	_, err = io.ReadFull(r, b[20:])
	if err != nil {
		return nil, err
	}
	return b, nil
}
func WriteFrame(w io.Writer, b []byte, l wire.Limits) (int, error) {
	_, n, err := wire.Header(b, l)
	if err != nil {
		return 0, err
	}
	if n != len(b) {
		return 0, wire.ErrFrame
	}
	written, err := w.Write(b)
	if written < 0 || written > len(b) {
		return 0, errors.New("writer returned invalid count")
	}
	if err == nil && written != len(b) {
		err = io.ErrShortWrite
	}
	return written, err
}
