// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: SMS client wire NUL and equality diagnostics with original bounded quotes.
// SPDX-AI-Date: 2026-11-24

package combined

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestSMSClientModes(t *testing.T) {
	for _, kind := range []string{"iec", "initial", "term"} {
		for sample, data := range [][]byte{nil, {}, []byte("ok"), []byte("different"), {0}, {1, 0, 2}, bytes.Repeat([]byte{1}, 1024)} {
			for _, fault := range []string{"plain", "sequence", "indicator", "equipment", "node", "type", "client-missing", "recipient-missing", "smsc-missing", "reply-missing", "header-missing", "later-mms", "later-action", "later-timestamp"} {
				for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
					t.Run(fmt.Sprintf("%s/%d/%s/%d", kind, sample, fault, mode), func(t *testing.T) {
						m := actionRequest(t, kind)
						s := &m.ServiceInformation.SMSInformation
						s.SMSCAddress = avp.Some([]byte("ok"))
						s.ClientAddressWire = avp.Some(data)
						switch fault {
						case "sequence":
							m.CCRequestNumber = 9
						case "indicator":
							m.MultipleServicesIndicator = avp.Some(int32(9))
						case "equipment":
							m.UserEquipmentInfo.Value.UserEquipmentInfoValue = []byte("other")
						case "node":
							s.SMSNode = 9
						case "type":
							s.SMMessageType = 9
						case "client-missing":
							s.ClientAddressWire.Present = false
						case "recipient-missing":
							s.RecipientSCCPAddress.Present = false
						case "smsc-missing":
							s.SMSCAddress.Present = false
						case "reply-missing":
							s.ReplyPathRequested.Present = false
						case "header-missing":
							s.SMUserDataHeader.Present = false
						case "later-mms":
							m.ServiceInformation.MMSInformation.Present = false
						case "later-action":
							m.RequestedAction = avp.Some(uint32(99))
						case "later-timestamp":
							m.ServiceInformation.MMSInformation.Value.SubmissionTime = avp.Some(m.EventTimestamp + 1)
						}
						e := validateCCRSubset(m, mode)
						var proof interface{ DiameterSMSClientValueV1() string }
						reached := fault == "plain" || fault == "later-mms" || fault == "later-action" || fault == "later-timestamp"
						invalid := !bytes.Equal(data, []byte("ok"))
						want := reached && invalid && mode == validation.Receive
						if errors.As(e, &proof) != want {
							t.Fatal("evidence order", e)
						}
						if reached && invalid {
							var f *validation.FieldError
							if !errors.Is(e, validation.ErrField) || errors.Is(e, validation.ErrUnsupported) || !errors.As(e, &f) {
								t.Fatal(e)
							}
							n := 1
							if want {
								n = 3
							}
							if len(f.Path) != n || f.Path[n-1] != (avp.Key{VendorID: 10415, Code: 2018}) {
								t.Fatal(f.Path)
							}
						}
						if fault == "plain" && (e == nil) != !invalid {
							t.Fatal("acceptance", e)
						}
						if want {
							name := "combined-sms-client-mismatch-v1"
							if bytes.IndexByte(data, 0) >= 0 {
								name = "combined-sms-client-nul-v1"
							}
							*m = CCR{}
							if proof.DiameterSMSClientValueV1() != name {
								t.Fatal("proof identity")
							}
						}
					})
				}
			}
		}
	}
}
func TestSMSClientMoneyBoundary(t *testing.T) {
	for _, kind := range []string{"general", "balance", "refund"} {
		for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
			m := actionRequest(t, kind)
			m.ServiceInformation.SMSInformation.ClientAddressWire = avp.Some([]byte{0})
			e := validateCCRSubset(m, mode)
			var proof interface{ DiameterSMSClientValueV1() string }
			if errors.As(e, &proof) {
				t.Fatal(kind, e)
			}
		}
	}
}

func TestSMSClientEmptyEqual(t *testing.T) {
	for _, kind := range []string{"iec", "initial", "term"} {
		for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
			m := actionRequest(t, kind)
			m.ServiceInformation.SMSInformation.ClientAddressWire = avp.Some([]byte{})
			m.ServiceInformation.SMSInformation.SMSCAddress = avp.Some([]byte{})
			if e := validateCCRSubset(m, mode); e != nil {
				t.Fatal(kind, e)
			}
		}
	}
}
