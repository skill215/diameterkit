// Package scap1 implements a fixed sms-fw SCAP1 event dialect.
// Its legacy header is App0/271, with a distinct Vendor193/Acct19302 declaration.
// This is not standard Ro or Base Accounting. Runtime use requires an explicit
// SCAP1Accounting binding and accounting capability; importing this package does
// not advertise or enable it. No storage, retry or Getter behavior is installed.
package scap1

import (
	"bytes"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

const ApplicationID uint32 = 0
const CommandCode uint32 = 271
const VendorID uint32 = 193
const AccountingApplicationID uint32 = 19302

func extras(xs []avp.Raw) error {
	if len(xs) > 0 {
		return validation.At(xs[0].Key, validation.ErrUnsupported)
	}
	return nil
}
func at(code, vendor uint32, e error) error {
	return validation.At(avp.Key{Code: code, VendorID: vendor}, e)
}
func validateVendorApplicationSubset(m *VendorApplication, _ validation.Mode) error {
	if m.VendorID != 193 || m.AcctApplicationID != 19302 {
		return validation.ErrUnsupported
	}
	return extras(m.Extra)
}
func validateSubscriptionIDSubset(m *SubscriptionID, _ validation.Mode) error {
	if m.SubscriptionIDType != 0 {
		return at(555, 193, validation.ErrUnsupported)
	}
	if bytes.IndexByte(m.SubscriptionIDData, 0) >= 0 {
		return at(554, 193, validation.ErrField)
	}
	return extras(m.Extra)
}
func validateUnitValueSubset(m *UnitValue, _ validation.Mode) error {
	if m.ValueDigits != 1 {
		return at(617, 193, validation.ErrUnsupported)
	}
	return extras(m.Extra)
}
func validateRequestedUnitSubset(m *RequestedUnit, _ validation.Mode) error {
	if m.UnitType != 2 {
		return at(611, 193, validation.ErrUnsupported)
	}
	return extras(m.Extra)
}
func validateServiceParameterSubset(m *ServiceParameter, _ validation.Mode) error {
	if bytes.IndexByte(m.ParameterValue, 0) >= 0 {
		return at(609, 193, validation.ErrField)
	}
	return extras(m.Extra)
}
func validateACRSubset(m *ACR, _ validation.Mode) error {
	if m.AccountingRecordType != 1 {
		return at(480, 0, validation.ErrUnsupported)
	}
	if m.AccountingRecordNumber != 0 {
		return at(485, 0, validation.ErrUnsupported)
	}
	if m.RequestedAction != 0 {
		return at(615, 193, validation.ErrUnsupported)
	}
	types := [...]int32{0, 1, 5, 8, 14, 16, 17, 18, 2, 3, 13}
	if len(m.ServiceParameterInfo) != len(types) {
		return at(607, 193, validation.ErrField)
	}
	for i, v := range m.ServiceParameterInfo {
		if v.ParameterType != types[i] {
			return at(607, 193, validation.ErrUnsupported)
		}
	}
	if !bytes.Equal(m.ServiceParameterInfo[0].ParameterValue, []byte("1")) || !bytes.Equal(m.ServiceParameterInfo[7].ParameterValue, []byte("4")) || !bytes.Equal(m.ServiceParameterInfo[10].ParameterValue, []byte("+0600")) || !bytes.Equal(m.ServiceParameterInfo[2].ParameterValue, m.ServiceParameterInfo[5].ParameterValue) {
		return at(607, 193, validation.ErrField)
	}
	return extras(m.Extra)
}
func validateACASubset(m *ACA, mode validation.Mode) error {
	if mode != validation.Receive && (m.AccountingRecordType != 1 || m.AccountingRecordNumber != 0) {
		return validation.ErrUnsupported
	}
	switch m.ResultCode {
	case 2001:
	case 5001, 5004, 5005, 5007, 5008, 5009, 5012, 5014:
		if e := base.ValidateFailureDiagnostic(m.ResultCode, m.FailedAVP.Present, mode); e != nil {
			return e
		}
	default:
		return at(268, 0, validation.ErrUnsupported)
	}
	return extras(m.Extra)
}
func validateACARelation(r *ACR, a *ACA) error {
	if r.AccountingRecordType != a.AccountingRecordType {
		return at(480, 0, validation.ErrInvalidAnswer)
	}
	if uint32(r.AccountingRecordNumber) != a.AccountingRecordNumber {
		return at(485, 0, validation.ErrInvalidAnswer)
	}
	return nil
}
func checkEvent(request, answer pdu.Body, ctx validation.Context, _ validation.Mode) error {
	if !ctx.Empty() {
		return validation.ErrContext
	}
	if _, ok := request.(*ACR); !ok {
		return validation.ErrContextRequired
	}
	return nil
}
