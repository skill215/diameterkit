package s6c

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

func hasIPSMGW(number avp.Optional[[]byte], name, realm avp.Optional[string], sbi avp.Optional[uint32]) bool {
	return number.Present || name.Present || realm.Present || sbi.Present
}

// Realm ambiguity never masks independent missing-number or mixed-node errors.
// Generated field checks enforce TBCD/identity/width and per-parent cardinality.
func checkIPSMGWShape(number avp.Optional[[]byte], other bool) error {
	if !number.Present {
		return validation.At(smsKey(3100), validation.ErrMissing)
	}
	if other {
		return validation.ErrField
	}
	return nil
}

func checkIPSMGWSelection(realm avp.Optional[string], sbi avp.Optional[uint32]) error {
	if realm.Present {
		return validation.At(smsKey(3112), validation.ErrUnsupported)
	}
	if sbi.Present && sbi.Value > 1 {
		return validation.At(smsKey(3348), validation.ErrUnsupported)
	}
	return nil
}

// This is a fixed profile boundary, not an application fact that Context can
// override. No runtime policy switch, Peer mutation or address resolution occurs.
func checkSRAIPSMGWSelection(a *SRA) error {
	if a.AdditionalServingNode.Present && a.AdditionalServingNode.Value.IPSMGWNumber.Present {
		return validation.At(smsKey(2406), validation.At(smsKey(3100), validation.ErrUnsupported))
	}
	if a.ServingNode.Present && a.ServingNode.Value.IPSMGWNumber.Present {
		n := a.ServingNode.Value
		if err := checkIPSMGWSelection(n.IPSMGWRealm, n.IPSMGWSBISupportIndicator); err != nil {
			return validation.At(smsKey(2401), err)
		}
		if a.AdditionalServingNode.Present {
			return validation.At(smsKey(2406), validation.ErrUnsupported)
		}
	}
	return nil
}
