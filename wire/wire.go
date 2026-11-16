// Package wire provides bounded byte primitives for generated Diameter codecs.
// Views borrow their input for the duration of the enclosing codec operation.
package wire

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"io"
)

var (
	ErrFrame  = errors.New("invalid Diameter framing")
	ErrAVP    = errors.New("invalid AVP framing")
	ErrLimit  = errors.New("resource limit exceeded")
	ErrConfig = errors.New("invalid codec limits")
)

const MaxWireSize = 0xffffff

type Limits struct{ MaxFrameBytes, MaxFieldBytes, MaxAVPs, MaxDepth, MaxRepeated int }

func DefaultLimits() Limits { return Limits{1 << 20, 1 << 19, 4096, 16, 256} }
func (l Limits) Validate() error {
	if l.MaxFrameBytes < 20 || l.MaxFrameBytes > MaxWireSize || l.MaxFieldBytes < 1 || l.MaxFieldBytes > l.MaxFrameBytes || l.MaxAVPs < 1 || l.MaxDepth < 1 || l.MaxDepth > 64 || l.MaxRepeated < 1 || l.MaxRepeated > l.MaxAVPs {
		return ErrConfig
	}
	return nil
}
func Uint24(b []byte) uint32   { return uint32(b[0])<<16 | uint32(b[1])<<8 | uint32(b[2]) }
func Put24(b []byte, v uint32) { b[0] = byte(v >> 16); b[1] = byte(v >> 8); b[2] = byte(v) }
func Header(b []byte, l Limits) (pdu.Header, int, error) {
	if err := l.Validate(); err != nil {
		return pdu.Header{}, 0, err
	}
	if len(b) < 20 {
		return pdu.Header{}, 0, io.ErrUnexpectedEOF
	}
	n := int(Uint24(b[1:4]))
	if b[0] != 1 || n < 20 || n%4 != 0 {
		return pdu.Header{}, 0, ErrFrame
	}
	if n > l.MaxFrameBytes {
		return pdu.Header{}, 0, ErrLimit
	}
	h := pdu.Header{Flags: b[4], CommandCode: Uint24(b[5:8]), ApplicationID: binary.BigEndian.Uint32(b[8:12]), HopByHop: binary.BigEndian.Uint32(b[12:16]), EndToEnd: binary.BigEndian.Uint32(b[16:20])}
	if h.Flags&pdu.RequestFlag != 0 && h.Flags&pdu.ErrorFlag != 0 || h.Flags&pdu.RequestFlag == 0 && h.Flags&pdu.RetransmittedFlag != 0 {
		return h, n, ErrFrame
	}
	return h, n, nil
}
func PutHeader(b []byte, h pdu.Header, n int) {
	b[0] = 1
	Put24(b[1:4], uint32(n))
	b[4] = h.Flags
	Put24(b[5:8], h.CommandCode)
	binary.BigEndian.PutUint32(b[8:12], h.ApplicationID)
	binary.BigEndian.PutUint32(b[12:16], h.HopByHop)
	binary.BigEndian.PutUint32(b[16:20], h.EndToEnd)
}

// DecodeError records framing progress without retaining the input frame.
type DecodeError struct {
	Offset        int
	FrameComplete bool
	Cause         error
}

func (e *DecodeError) Error() string { return fmt.Sprintf("decode at %d: %v", e.Offset, e.Cause) }
func (e *DecodeError) Unwrap() error { return e.Cause }

type View struct {
	Key    avp.Key
	Flags  byte
	Data   []byte
	Offset int
}
type counter struct{ avps int }
type Decoder struct {
	data          []byte
	offset, depth int
	limits        Limits
	count         *counter
}

func NewDecoder(b []byte, l Limits) *Decoder { return &Decoder{data: b, limits: l, count: &counter{}} }
func (d *Decoder) Limits() Limits            { return d.limits }
func (d *Decoder) Remaining() int            { return len(d.data) - d.offset }
func (d *Decoder) Child(b []byte) (*Decoder, error) {
	if d.depth+1 > d.limits.MaxDepth {
		return nil, ErrLimit
	}
	return &Decoder{data: b, depth: d.depth + 1, limits: d.limits, count: d.count}, nil
}
func (d *Decoder) Next() (View, bool, error) {
	if d.offset == len(d.data) {
		return View{}, false, nil
	}
	if err := d.limits.Validate(); err != nil {
		return View{}, false, err
	}
	p := d.data[d.offset:]
	if len(p) < 8 {
		return View{}, false, ErrAVP
	}
	n := int(Uint24(p[5:8]))
	head := 8
	vendor := uint32(0)
	if p[4]&avp.VendorFlag != 0 {
		head = 12
		if len(p) < head {
			return View{}, false, ErrAVP
		}
		vendor = binary.BigEndian.Uint32(p[8:12])
		if vendor == 0 {
			return View{}, false, ErrAVP
		}
	}
	if n < head || n > len(p) {
		return View{}, false, ErrAVP
	}
	size := (n + 3) &^ 3
	if size > len(p) {
		return View{}, false, ErrAVP
	}
	if n-head > d.limits.MaxFieldBytes || d.count.avps >= d.limits.MaxAVPs {
		return View{}, false, ErrLimit
	}
	v := View{avp.Key{VendorID: vendor, Code: binary.BigEndian.Uint32(p[:4])}, p[4], p[head:n:n], d.offset}
	d.offset += size
	d.count.avps++
	return v, true, nil
}
func U32(b []byte) (uint32, error) {
	if len(b) != 4 {
		return 0, avp.ErrValue
	}
	return binary.BigEndian.Uint32(b), nil
}

// I32 decodes an exact-width Diameter Integer32, preserving the signed value.
func I32(b []byte) (int32, error) {
	v, err := U32(b)
	return int32(v), err
}

// U64 decodes an exact-width Diameter Unsigned64 without truncation.
func U64(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, avp.ErrValue
	}
	return binary.BigEndian.Uint64(b), nil
}

// I64 decodes an exact-width Diameter Integer64 without float conversion.
func I64(b []byte) (int64, error) { v, err := U64(b); return int64(v), err }

// Encoder uses nil output to measure. Non-nil output is caller-sized and never grown.
type Encoder struct {
	dst            []byte
	n, avps, depth int
	limits         Limits
	err            error
	allowReserved  bool
}

func NewEncoder(dst []byte, l Limits) *Encoder {
	e := &Encoder{dst: dst, limits: l}
	e.err = l.Validate()
	return e
}
func (e *Encoder) Len() int       { return e.n }
func (e *Encoder) Err() error     { return e.err }
func (e *Encoder) Limits() Limits { return e.limits }
func (e *Encoder) reserve(n int) (int, bool) {
	if e.err != nil {
		return 0, false
	}
	if n < 0 || n > e.limits.MaxFrameBytes-e.n {
		e.err = ErrLimit
		return 0, false
	}
	start := e.n
	e.n += n
	if e.dst != nil {
		if e.n > len(e.dst) {
			e.err = io.ErrShortBuffer
			return 0, false
		}
		clear(e.dst[start:e.n])
	}
	return start, true
}
func (e *Encoder) header(k avp.Key, flags byte, payload int) (int, int, bool) {
	if e.err != nil {
		return 0, 0, false
	}
	if !e.allowReserved && flags&0x1f != 0 || (k.VendorID != 0) != (flags&avp.VendorFlag != 0) {
		e.err = ErrAVP
		return 0, 0, false
	}
	if payload < 0 || payload > e.limits.MaxFieldBytes || e.avps >= e.limits.MaxAVPs {
		e.err = ErrLimit
		return 0, 0, false
	}
	h := 8
	if k.VendorID != 0 {
		h = 12
	}
	n := h + payload
	if n > MaxWireSize {
		e.err = ErrLimit
		return 0, 0, false
	}
	start, ok := e.reserve((n + 3) &^ 3)
	if !ok {
		return 0, 0, false
	}
	e.avps++
	if e.dst != nil {
		p := e.dst[start:]
		binary.BigEndian.PutUint32(p, k.Code)
		p[4] = flags
		Put24(p[5:8], uint32(n))
		if h == 12 {
			binary.BigEndian.PutUint32(p[8:12], k.VendorID)
		}
	}
	return start, h, true
}
func (e *Encoder) Bytes(k avp.Key, f byte, b []byte) {
	s, h, ok := e.header(k, f, len(b))
	if ok && e.dst != nil {
		copy(e.dst[s+h:], b)
	}
}
func (e *Encoder) String(k avp.Key, f byte, s string) {
	p, h, ok := e.header(k, f, len(s))
	if ok && e.dst != nil {
		copy(e.dst[p+h:], s)
	}
}
func (e *Encoder) Uint32(k avp.Key, f byte, v uint32) {
	s, h, ok := e.header(k, f, 4)
	if ok && e.dst != nil {
		binary.BigEndian.PutUint32(e.dst[s+h:], v)
	}
}

// Int32 encodes the two's-complement Integer32 representation.
func (e *Encoder) Int32(k avp.Key, f byte, v int32) { e.Uint32(k, f, uint32(v)) }

func (e *Encoder) Uint64(k avp.Key, f byte, v uint64) {
	s, h, ok := e.header(k, f, 8)
	if ok && e.dst != nil {
		binary.BigEndian.PutUint64(e.dst[s+h:], v)
	}
}

// Int64 preserves the complete two's-complement Integer64 representation.
func (e *Encoder) Int64(k avp.Key, f byte, v int64) { e.Uint64(k, f, uint64(v)) }

func (e *Encoder) Address(k avp.Key, f byte, a avp.Address) {
	if err := a.Validate(); err != nil {
		e.err = err
		return
	}
	s, h, ok := e.header(k, f, 2+len(a.Data))
	if ok && e.dst != nil {
		binary.BigEndian.PutUint16(e.dst[s+h:], a.Family)
		copy(e.dst[s+h+2:], a.Data)
	}
}

// GroupMark belongs to the encoder that created it; generated code closes groups LIFO.
type GroupMark struct{ start, head, depth int }

func (e *Encoder) BeginGroup(k avp.Key, f byte) GroupMark {
	if e.depth >= e.limits.MaxDepth {
		e.err = ErrLimit
		return GroupMark{start: -1}
	}
	s, h, ok := e.header(k, f, 0)
	if !ok {
		return GroupMark{start: -1}
	}
	e.depth++
	return GroupMark{s, h, e.depth}
}
func (e *Encoder) EndGroup(mark GroupMark) {
	if e.err != nil {
		return
	}
	if mark.start < 0 || mark.depth != e.depth || mark.start > e.n {
		e.err = ErrAVP
		return
	}
	e.depth--
	n := e.n - mark.start
	if n > MaxWireSize || n-mark.head > e.limits.MaxFieldBytes {
		e.err = ErrLimit
		return
	}
	if e.dst != nil {
		Put24(e.dst[mark.start+5:mark.start+8], uint32(n))
	}
}

// DecodeBound estimates retained frame, copied strings, generated structs and
// slice capacity. It is a conservative logical reservation, not a Go heap/RSS
// guarantee. Generated structureBytes includes inline Grouped fields.
func DecodeBound(n int, l Limits, structureBytes int) (uint64, error) {
	if err := l.Validate(); err != nil {
		return 0, err
	}
	if n < 20 || n > l.MaxFrameBytes || structureBytes < 1 || structureBytes > 1<<20 {
		return 0, ErrLimit
	}
	count := min(l.MaxAVPs, (n-20)/8)
	return 4*uint64(n) + 4*uint64(count+1)*uint64(structureBytes) + 4096, nil
}

// NewReceiveSizer measures validated received values while ignoring reserved
// flag bits. It cannot write bytes. Sending still uses NewEncoder and Send rules.
func NewReceiveSizer(l Limits) *Encoder { e := NewEncoder(nil, l); e.allowReserved = true; return e }
