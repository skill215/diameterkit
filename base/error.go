// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package base

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/schema"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// ErrorOptions is compiled Profile metadata, not per-message policy.
type ErrorOptions struct {
	// Rule narrows the E grammar for a compiled application profile. It must be pure.
	Rule              func(*pdu.ErrorAnswer, validation.Mode) error
	DiagnosticGroups  []avp.Key
	AuthSessionState  avp.Optional[uint32]
	IgnoreVSAI        bool
	UnsupportedFields []avp.Key
}

// NewErrorOperations binds the common E grammar to one normal answer identity.
// Redirect 3006 remains explicitly unsupported until the DiameterURI subset lands.
func NewErrorOperations(key pdu.CommandKey, options ErrorOptions) *schema.ErrorOperations {
	groups := append(DiagnosticGroups(), options.DiagnosticGroups...)
	options.DiagnosticGroups = nil
	options.UnsupportedFields = append([]avp.Key(nil), options.UnsupportedFields...)
	check := func(b pdu.Body, mode validation.Mode) error {
		m, ok := b.(*pdu.ErrorAnswer)
		if !ok || m == nil {
			return validation.ErrType
		}
		if m.Command != key && !(mode == validation.Build && m.Command == (pdu.CommandKey{})) {
			return validation.ErrType
		}
		if err := validateError(m, mode, options); err != nil {
			return err
		}
		if options.Rule != nil {
			return options.Rule(m, mode)
		}
		return nil
	}
	return &schema.ErrorOperations{
		Validate: check,
		Decode: func(d *wire.Decoder) (pdu.Body, error) {
			m, err := decodeError(d, key, groups, options)
			if err != nil {
				return nil, err
			}
			return m, check(m, validation.Receive)
		},
		Encode: func(w *wire.Encoder, b pdu.Body) error {
			m, ok := b.(*pdu.ErrorAnswer)
			if !ok || m == nil {
				return validation.ErrType
			}
			return encodeError(w, m, groups)
		},
		Clone: func(b pdu.Body) (pdu.Body, error) {
			m, ok := b.(*pdu.ErrorAnswer)
			if !ok || m == nil {
				return nil, validation.ErrType
			}
			return cloneError(m), nil
		},
		DecodeBound: func(n int, l wire.Limits) (uint64, error) { return wire.DecodeBound(n, l, 4096) },
		Envelope: func(b pdu.Body) (pdu.Envelope, error) {
			m, ok := b.(*pdu.ErrorAnswer)
			if !ok || m == nil {
				return pdu.Envelope{}, validation.ErrType
			}
			return pdu.Envelope{SessionID: m.SessionID, Origin: pdu.Origin{Host: m.OriginHost, Realm: m.OriginRealm}, ProxyInfo: m.ProxyInfo, AuthSessionState: m.AuthSessionState}, nil
		},
		CompleteAnswer: func(b pdu.Body, env pdu.AnswerEnvelope) (pdu.Body, error) {
			m, ok := b.(*pdu.ErrorAnswer)
			if !ok || m == nil {
				return nil, validation.ErrType
			}
			if m.Command != (pdu.CommandKey{}) && m.Command != env.Command {
				return nil, validation.ErrEnvelope
			}
			c := cloneError(m)
			c.Command = env.Command
			if err := CompleteEnvelopeValues(&c.SessionID, &c.OriginHost, &c.OriginRealm, &c.ProxyInfo, env); err != nil {
				return nil, err
			}
			return c, nil
		},
	}
}

var errorKeys = []avp.Key{{Code: 263}, {Code: 264}, {Code: 296}, {Code: 268}, {Code: 278}, {Code: 281}, {Code: 294}, {Code: 279}, {Code: 297}, {Code: 284}, {Code: 277}, {Code: 260}, {Code: 282}}

func validateError(m *pdu.ErrorAnswer, mode validation.Mode, options ErrorOptions) error {
	if m.ResultCode == 3006 {
		return validation.ErrUnsupported
	}
	if !(m.ResultCode >= 3000 && m.ResultCode < 4000 || m.ResultCode >= 5000 && m.ResultCode < 6000) {
		return validation.At(avp.Key{Code: 268}, validation.ErrField)
	}
	if m.SessionID.Present {
		if m.SessionID.Value == "" {
			return validation.At(avp.Key{Code: 263}, validation.ErrField)
		}
		if err := avp.UTF8(m.SessionID.Value); err != nil {
			return validation.At(avp.Key{Code: 263}, err)
		}
	}
	for _, f := range []struct {
		k     uint32
		value string
	}{{264, m.OriginHost}, {296, m.OriginRealm}} {
		if mode == validation.Build && f.value == "" {
			continue
		}
		if err := avp.Identity(f.value); err != nil {
			return validation.At(avp.Key{Code: f.k}, err)
		}
	}
	if m.ErrorMessage.Present {
		if err := avp.UTF8(m.ErrorMessage.Value); err != nil {
			return validation.At(avp.Key{Code: 281}, err)
		}
	}
	if m.ErrorReportingHost.Present {
		if err := avp.Identity(m.ErrorReportingHost.Value); err != nil {
			return validation.At(avp.Key{Code: 294}, err)
		}
	}
	if err := ValidateFailureDiagnostic(m.ResultCode, m.FailedAVP.Present, mode); err != nil {
		return err
	}
	if m.FailedAVP.Present {
		if err := ValidateDiagnostic(&m.FailedAVP.Value, mode); err != nil {
			return validation.At(avp.Key{Code: 279}, err)
		}
	}
	if m.ExperimentalResult.Present {
		if err := ValidateExtra(m.ExperimentalResult.Value.Extra, mode, avp.Key{Code: 266}, avp.Key{Code: 298}); err != nil {
			return validation.At(avp.Key{Code: 297}, err)
		}
	}
	for i := range m.ProxyInfo {
		if err := ValidateProxy(&m.ProxyInfo[i], mode); err != nil {
			return validation.At(avp.Key{Code: 284}, err)
		}
	}
	if m.AuthSessionState.Present && (m.AuthSessionState.Value > 1 || options.AuthSessionState.Present && m.AuthSessionState.Value != options.AuthSessionState.Value) {
		return validation.At(avp.Key{Code: 277}, validation.ErrField)
	}
	if m.VendorSpecificApplicationID.Present {
		if options.IgnoreVSAI {
			return validation.ErrUnsupported
		}
		if err := ValidateVSAI(&m.VendorSpecificApplicationID.Value, mode); err != nil {
			return validation.At(avp.Key{Code: 260}, err)
		}
	}
	for _, v := range m.Extra {
		for _, k := range options.UnsupportedFields {
			if k == v.Key {
				return validation.At(v.Key, validation.ErrUnsupported)
			}
		}
		if v.Key == (avp.Key{Code: 283}) || v.Key == (avp.Key{Code: 293}) {
			return validation.At(v.Key, validation.ErrField)
		}
		if options.IgnoreVSAI && mode == validation.Receive && v.Key == (avp.Key{Code: 260}) {
			continue
		}
		if err := ValidateExtra([]avp.Raw{v}, mode, errorKeys...); err != nil {
			return err
		}
	}
	for _, v := range m.RouteRecord {
		if err := avp.Identity(v); err != nil {
			return validation.At(avp.Key{Code: 282}, err)
		}
	}
	return nil
}
func encodeError(w *wire.Encoder, m *pdu.ErrorAnswer, groups []avp.Key) error {
	if m.SessionID.Present {
		w.String(avp.Key{Code: 263}, 0x40, m.SessionID.Value)
	}
	w.String(avp.Key{Code: 264}, 0x40, m.OriginHost)
	w.String(avp.Key{Code: 296}, 0x40, m.OriginRealm)
	w.Uint32(avp.Key{Code: 268}, 0x40, m.ResultCode)
	if m.OriginStateID.Present {
		w.Uint32(avp.Key{Code: 278}, 0x40, m.OriginStateID.Value)
	}
	if m.ErrorMessage.Present {
		w.String(avp.Key{Code: 281}, 0, m.ErrorMessage.Value)
	}
	if m.ErrorReportingHost.Present {
		w.String(avp.Key{Code: 294}, 0, m.ErrorReportingHost.Value)
	}
	if m.FailedAVP.Present {
		w.Diagnostic(avp.Key{Code: 279}, 0x40, m.FailedAVP.Value.Data, groups)
	}
	if m.ExperimentalResult.Present {
		mark := w.BeginGroup(avp.Key{Code: 297}, 0x40)
		v := m.ExperimentalResult.Value
		w.Uint32(avp.Key{Code: 266}, 0x40, v.VendorID)
		w.Uint32(avp.Key{Code: 298}, 0x40, v.ExperimentalResultCode)
		if err := encodeExtra(w, v.Extra); err != nil {
			return err
		}
		w.EndGroup(mark)
	}
	if len(m.ProxyInfo) > w.Limits().MaxRepeated {
		return wire.ErrLimit
	}
	for i := range m.ProxyInfo {
		mark := w.BeginGroup(avp.Key{Code: 284}, 0x40)
		if err := EncodeProxy(w, &m.ProxyInfo[i]); err != nil {
			return err
		}
		w.EndGroup(mark)
	}
	if m.AuthSessionState.Present {
		w.Uint32(avp.Key{Code: 277}, 0x40, m.AuthSessionState.Value)
	}
	if m.VendorSpecificApplicationID.Present {
		mark := w.BeginGroup(avp.Key{Code: 260}, 0x40)
		if err := EncodeVSAI(w, &m.VendorSpecificApplicationID.Value); err != nil {
			return err
		}
		w.EndGroup(mark)
	}
	if len(m.RouteRecord) > w.Limits().MaxRepeated {
		return wire.ErrLimit
	}
	for _, v := range m.RouteRecord {
		w.String(avp.Key{Code: 282}, 0x40, v)
	}
	return encodeExtra(w, m.Extra)
}
func decodeExperimental(d *wire.Decoder) (pdu.ExperimentalResult, error) {
	var m pdu.ExperimentalResult
	vendor, code := false, false
	for {
		v, ok, err := d.Next()
		if err != nil {
			return m, err
		}
		if !ok {
			break
		}
		switch v.Key {
		case avp.Key{Code: 266}:
			if vendor {
				return m, validation.At(v.Key, validation.ErrDuplicate)
			}
			vendor = true
			m.VendorID, err = wire.U32(v.Data)
		case avp.Key{Code: 298}:
			if code {
				return m, validation.At(v.Key, validation.ErrDuplicate)
			}
			code = true
			m.ExperimentalResultCode, err = wire.U32(v.Data)
		default:
			err = appendExtra(&m.Extra, v, d.Limits())
		}
		if err != nil {
			return m, validation.At(v.Key, err)
		}
	}
	if !vendor || !code {
		return m, validation.ErrMissing
	}
	return m, nil
}
func decodeError(d *wire.Decoder, key pdu.CommandKey, groups []avp.Key, options ErrorOptions) (*pdu.ErrorAnswer, error) {
	m := &pdu.ErrorAnswer{Command: key}
	var seen [13]bool
	for {
		v, ok, err := d.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		index := -1
		for i, k := range errorKeys {
			if k == v.Key {
				index = i
				break
			}
		}
		if index >= 0 && v.Key.Code != 284 && v.Key.Code != 282 && !(options.IgnoreVSAI && v.Key.Code == 260) {
			if seen[index] {
				return nil, validation.At(v.Key, validation.ErrDuplicate)
			}
			seen[index] = true
		}
		switch v.Key {
		case avp.Key{Code: 263}:
			if v.Offset != 0 {
				return nil, validation.At(v.Key, validation.ErrField)
			}
			m.SessionID = avp.Some(string(v.Data))
		case avp.Key{Code: 264}:
			m.OriginHost = string(v.Data)
		case avp.Key{Code: 296}:
			m.OriginRealm = string(v.Data)
		case avp.Key{Code: 268}:
			m.ResultCode, err = wire.U32(v.Data)
		case avp.Key{Code: 278}:
			var n uint32
			n, err = wire.U32(v.Data)
			m.OriginStateID = avp.Some(n)
		case avp.Key{Code: 281}:
			m.ErrorMessage = avp.Some(string(v.Data))
		case avp.Key{Code: 294}:
			m.ErrorReportingHost = avp.Some(string(v.Data))
		case avp.Key{Code: 277}:
			var n uint32
			n, err = wire.U32(v.Data)
			m.AuthSessionState = avp.Some(n)
		case avp.Key{Code: 279}:
			var complete bool
			complete, err = d.Diagnostic(v.Data, groups)
			m.FailedAVP = avp.Some(pdu.FailedAVP{Data: v.Data, Incomplete: !complete})
		case avp.Key{Code: 297}:
			var child *wire.Decoder
			child, err = d.Child(v.Data)
			if err == nil {
				var x pdu.ExperimentalResult
				x, err = decodeExperimental(child)
				m.ExperimentalResult = avp.Some(x)
			}
		case avp.Key{Code: 284}:
			if len(m.ProxyInfo) >= d.Limits().MaxRepeated {
				return nil, wire.ErrLimit
			}
			var child *wire.Decoder
			child, err = d.Child(v.Data)
			if err == nil {
				var x *pdu.ProxyInfo
				x, err = DecodeProxy(child)
				if err == nil {
					m.ProxyInfo = append(m.ProxyInfo, *x)
				}
			}
		case avp.Key{Code: 260}:
			if options.IgnoreVSAI {
				if len(m.Extra) >= d.Limits().MaxRepeated {
					return nil, wire.ErrLimit
				}
				m.Extra = append(m.Extra, avp.Raw{Key: v.Key, Flags: v.Flags, Data: v.Data})
				break
			}
			var child *wire.Decoder
			child, err = d.Child(v.Data)
			if err == nil {
				var x *pdu.VendorSpecificApplicationID
				x, err = DecodeVSAI(child)
				if err == nil {
					m.VendorSpecificApplicationID = avp.Some(*x)
				}
			}
		case avp.Key{Code: 283}, avp.Key{Code: 293}:
			return nil, validation.At(v.Key, validation.ErrField)
		case avp.Key{Code: 282}:
			if len(m.RouteRecord) >= d.Limits().MaxRepeated {
				return nil, wire.ErrLimit
			}
			m.RouteRecord = append(m.RouteRecord, string(v.Data))
		default:
			err = appendExtra(&m.Extra, v, d.Limits())
		}
		if err != nil {
			return nil, validation.At(v.Key, err)
		}
	}
	for _, i := range []int{1, 2, 3} {
		if !seen[i] {
			return nil, validation.At(errorKeys[i], validation.ErrMissing)
		}
	}
	return m, nil
}
func cloneError(m *pdu.ErrorAnswer) *pdu.ErrorAnswer {
	c := *m
	c.OriginHost = avp.CloneString(m.OriginHost)
	c.OriginRealm = avp.CloneString(m.OriginRealm)
	if c.SessionID.Present {
		c.SessionID.Value = avp.CloneString(c.SessionID.Value)
	}
	if c.ErrorMessage.Present {
		c.ErrorMessage.Value = avp.CloneString(c.ErrorMessage.Value)
	}
	if c.ErrorReportingHost.Present {
		c.ErrorReportingHost.Value = avp.CloneString(c.ErrorReportingHost.Value)
	}
	if !c.SessionID.Present {
		c.SessionID = avp.Optional[string]{}
	}
	if !c.ErrorMessage.Present {
		c.ErrorMessage = avp.Optional[string]{}
	}
	if !c.ErrorReportingHost.Present {
		c.ErrorReportingHost = avp.Optional[string]{}
	}
	if !c.OriginStateID.Present {
		c.OriginStateID = avp.Optional[uint32]{}
	}
	if !c.AuthSessionState.Present {
		c.AuthSessionState = avp.Optional[uint32]{}
	}
	c.Extra = cloneExtra(m.Extra)
	c.ProxyInfo = CloneProxies(m.ProxyInfo)
	c.RouteRecord = append([]string(nil), m.RouteRecord...)
	for i := range c.RouteRecord {
		c.RouteRecord[i] = avp.CloneString(c.RouteRecord[i])
	}
	if c.FailedAVP.Present {
		c.FailedAVP.Value = *CloneDiagnostic(&m.FailedAVP.Value)
	} else {
		c.FailedAVP = avp.Optional[pdu.FailedAVP]{}
	}
	if c.ExperimentalResult.Present {
		c.ExperimentalResult.Value.Extra = cloneExtra(m.ExperimentalResult.Value.Extra)
	} else {
		c.ExperimentalResult = avp.Optional[pdu.ExperimentalResult]{}
	}
	if c.VendorSpecificApplicationID.Present {
		c.VendorSpecificApplicationID.Value = *CloneVSAI(&m.VendorSpecificApplicationID.Value)
	} else {
		c.VendorSpecificApplicationID = avp.Optional[pdu.VendorSpecificApplicationID]{}
	}
	return &c
}
