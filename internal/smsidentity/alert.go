package smsidentity

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
)

// MaxAlertTokenBytes is a selected library limit, not a 3GPP IMSI length rule.
const MaxAlertTokenBytes = 256

// AlertToken preserves UTF8 bytes exactly; it performs no identity lookup or
// Unicode/case normalization. Shared by Router TFA output and later ALR input.
func AlertToken(s string) error {
	if len(s) == 0 || len(s) > MaxAlertTokenBytes || avp.UTF8(s) != nil {
		return validation.ErrField
	}
	return nil
}
