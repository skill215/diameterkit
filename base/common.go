// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package base

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
)

// DiagnosticGroups is a fresh list of known Base Grouped keys. Profiles append
// their compiled application groups; unknown payloads are never guessed to be groups.
func DiagnosticGroups() []avp.Key {
	return []avp.Key{{Code: 279}, {Code: 284}, {Code: 297}, {Code: 260}}
}

func ValidateExtra(extra []avp.Raw, mode validation.Mode, known ...avp.Key) error {
	for _, v := range extra {
		for _, k := range known {
			if v.Key == k {
				return validation.At(k, validation.ErrDuplicate)
			}
		}
		if v.Flags&avp.MandatoryFlag != 0 {
			return validation.At(v.Key, validation.ErrUnsupported)
		}
		if (v.Key.VendorID != 0) != (v.Flags&avp.VendorFlag != 0) || mode != validation.Receive && v.Flags&0x1f != 0 {
			return validation.At(v.Key, wire.ErrAVP)
		}
	}
	return nil
}
func appendExtra(dst *[]avp.Raw, v wire.View, l wire.Limits) error {
	if v.Flags&avp.MandatoryFlag != 0 {
		return validation.At(v.Key, validation.ErrUnsupported)
	}
	if len(*dst) >= l.MaxRepeated {
		return wire.ErrLimit
	}
	*dst = append(*dst, avp.Raw{Key: v.Key, Flags: v.Flags, Data: v.Data})
	return nil
}
func encodeExtra(w *wire.Encoder, extra []avp.Raw) error {
	if len(extra) > w.Limits().MaxRepeated {
		return wire.ErrLimit
	}
	for _, v := range extra {
		w.Bytes(v.Key, v.Flags, v.Data)
	}
	return w.Err()
}
func cloneExtra(v []avp.Raw) []avp.Raw {
	out := make([]avp.Raw, len(v))
	for i := range v {
		out[i] = v[i].Clone()
	}
	return out
}

func ValidateProxy(m *pdu.ProxyInfo, mode validation.Mode) error {
	if m == nil {
		return validation.ErrType
	}
	if err := avp.Identity(m.ProxyHost); err != nil {
		return validation.At(avp.Key{Code: 280}, err)
	}
	return ValidateExtra(m.Extra, mode, avp.Key{Code: 280}, avp.Key{Code: 33})
}
func EncodeProxy(w *wire.Encoder, m *pdu.ProxyInfo) error {
	w.String(avp.Key{Code: 280}, 0x40, m.ProxyHost)
	w.Bytes(avp.Key{Code: 33}, 0x40, m.ProxyState)
	return encodeExtra(w, m.Extra)
}
func DecodeProxy(d *wire.Decoder) (*pdu.ProxyInfo, error) {
	m := new(pdu.ProxyInfo)
	host, state := false, false
	for {
		v, ok, err := d.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		switch v.Key {
		case avp.Key{Code: 280}:
			if host {
				if d.Limits().MaxRepeated < 2 {
					return nil, wire.ErrLimit
				}
				return nil, profileTooMany(v, 1)
			}
			host = true
			m.ProxyHost = string(v.Data)
			if err := avp.Identity(m.ProxyHost); err != nil {
				return nil, profileBadValue(v, 5004, err)
			}
		case avp.Key{Code: 33}:
			if state {
				if d.Limits().MaxRepeated < 2 {
					return nil, wire.ErrLimit
				}
				return nil, profileTooMany(v, 1)
			}
			state = true
			m.ProxyState = v.Data
		default:
			if v.Flags&avp.MandatoryFlag != 0 {
				return nil, profileBadValue(v, 5001, validation.ErrUnsupported)
			}
			if err := appendExtra(&m.Extra, v, d.Limits()); err != nil {
				return nil, err
			}
		}
	}
	if !host {
		return nil, constraintFailure("proxy-host-missing", avp.Key{Code: 280}, validation.ErrMissing)
	}
	if !state {
		return nil, constraintFailure("proxy-state-missing", avp.Key{Code: 33}, validation.ErrMissing)
	}
	return m, ValidateProxy(m, validation.Receive)
}
func CloneProxy(m *pdu.ProxyInfo) *pdu.ProxyInfo {
	c := *m
	c.ProxyHost = avp.CloneString(m.ProxyHost)
	c.ProxyState = append([]byte(nil), m.ProxyState...)
	c.Extra = cloneExtra(m.Extra)
	return &c
}
func CloneProxies(v []pdu.ProxyInfo) []pdu.ProxyInfo {
	out := make([]pdu.ProxyInfo, len(v))
	for i := range v {
		out[i] = *CloneProxy(&v[i])
	}
	return out
}

// EqualProxies compares complete instances in order; State remains binary.
func EqualProxies(a, b []pdu.ProxyInfo) bool { return pdu.EqualProxies(a, b) }

func ValidateVSAI(m *pdu.VendorSpecificApplicationID, mode validation.Mode) error {
	if m == nil {
		return validation.ErrType
	}
	if m.AuthApplicationID.Present == m.AcctApplicationID.Present {
		return validation.ErrField
	}
	return ValidateExtra(m.Extra, mode, avp.Key{Code: 266}, avp.Key{Code: 258}, avp.Key{Code: 259})
}
func EncodeVSAI(w *wire.Encoder, m *pdu.VendorSpecificApplicationID) error {
	w.Uint32(avp.Key{Code: 266}, 0x40, m.VendorID)
	if m.AuthApplicationID.Present {
		w.Uint32(avp.Key{Code: 258}, 0x40, m.AuthApplicationID.Value)
	}
	if m.AcctApplicationID.Present {
		w.Uint32(avp.Key{Code: 259}, 0x40, m.AcctApplicationID.Value)
	}
	return encodeExtra(w, m.Extra)
}
func DecodeVSAI(d *wire.Decoder) (*pdu.VendorSpecificApplicationID, error) {
	m := new(pdu.VendorSpecificApplicationID)
	vendor := false
	for {
		v, ok, err := d.Next()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		if v.Key.VendorID == 0 && (v.Key.Code == 266 || v.Key.Code == 258 || v.Key.Code == 259) {
			duplicate := v.Key.Code == 266 && vendor || v.Key.Code == 258 && m.AuthApplicationID.Present || v.Key.Code == 259 && m.AcctApplicationID.Present
			if duplicate {
				if d.Limits().MaxRepeated < 2 {
					return nil, wire.ErrLimit
				}
				return nil, profileTooMany(v, 1)
			}
			n, err := wire.U32(v.Data)
			if err != nil {
				return nil, profileBadLength(v, "u32", err)
			}
			switch v.Key.Code {
			case 266:
				vendor = true
				m.VendorID = n
			case 258:
				m.AuthApplicationID = avp.Some(n)
			case 259:
				m.AcctApplicationID = avp.Some(n)
			}
		} else {
			if v.Flags&avp.MandatoryFlag != 0 {
				return nil, profileBadValue(v, 5001, validation.ErrUnsupported)
			}
			if err := appendExtra(&m.Extra, v, d.Limits()); err != nil {
				return nil, err
			}
		}
	}
	if !vendor {
		return nil, profileMissing(avp.Key{Code: 266}, avp.MandatoryFlag, 4)
	}
	if m.AuthApplicationID.Present && m.AcctApplicationID.Present {
		return nil, constraintFailure("vsai-both", avp.Key{}, validation.ErrDuplicate)
	}
	if !m.AuthApplicationID.Present && !m.AcctApplicationID.Present {
		return nil, constraintFailure("vsai-neither", avp.Key{}, validation.ErrMissing)
	}
	return m, ValidateVSAI(m, validation.Receive)
}
func CloneVSAI(m *pdu.VendorSpecificApplicationID) *pdu.VendorSpecificApplicationID {
	c := *m
	c.Extra = cloneExtra(m.Extra)
	if !c.AuthApplicationID.Present {
		c.AuthApplicationID = avp.Optional[uint32]{}
	}
	if !c.AcctApplicationID.Present {
		c.AcctApplicationID = avp.Optional[uint32]{}
	}
	return &c
}

func ValidateDiagnostic(m *pdu.FailedAVP, mode validation.Mode) error {
	if m == nil {
		return validation.ErrType
	}
	if len(m.Data) == 0 || mode != validation.Receive && len(m.Data) < 8 {
		return validation.ErrField
	}
	return nil
}
func CloneDiagnostic(m *pdu.FailedAVP) *pdu.FailedAVP {
	c := *m
	c.Data = append([]byte(nil), m.Data...)
	return &c
}

// ValidateFailureDiagnostic applies Base MUST requirements. 5005 is SHOULD on
// receive; local construction follows that recommendation without inventing evidence.
func ValidateFailureDiagnostic(code uint32, present bool, mode validation.Mode) error {
	required := code == 5001 || code == 5004 || code == 5007 || code == 5008 || code == 5009 || code == 5014 || mode != validation.Receive && code == 5005
	if required && !present {
		return validation.At(avp.Key{Code: 279}, validation.ErrMissing)
	}
	return nil
}

func mergeOrigin(current *string, want string) error {
	if *current != "" && *current != want {
		return validation.ErrEnvelope
	}
	*current = avp.CloneString(want)
	return nil
}

// CompleteEnvelopeValues is used by generated answer functions after cloning.
// Nonempty caller values must agree; missing values are supplied from the request.
func CompleteEnvelopeValues(session *avp.Optional[string], host, realm *string, proxy *[]pdu.ProxyInfo, env pdu.AnswerEnvelope) error {
	if session != nil {
		if session.Present && *session != env.SessionID {
			return validation.ErrEnvelope
		}
		*session = env.SessionID
		if session.Present {
			session.Value = avp.CloneString(session.Value)
		}
	} else if env.SessionID.Present {
		return validation.ErrEnvelope
	}
	if err := mergeOrigin(host, env.Origin.Host); err != nil {
		return err
	}
	if err := mergeOrigin(realm, env.Origin.Realm); err != nil {
		return err
	}
	if proxy != nil {
		if len(*proxy) > 0 && !EqualProxies(*proxy, env.ProxyInfo) {
			return validation.ErrEnvelope
		}
		*proxy = CloneProxies(env.ProxyInfo)
		for i := range *proxy {
			for j := range (*proxy)[i].Extra {
				(*proxy)[i].Extra[j].Flags &^= 0x1f
			}
		}
	} else if len(env.ProxyInfo) > 0 {
		return validation.ErrEnvelope
	}
	return nil
}
