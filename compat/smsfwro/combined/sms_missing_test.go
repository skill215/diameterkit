// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Reviewed isolated combined SMS condition diagnostics and acceptance.
// SPDX-AI-Date: 2026-11-18

package combined

import (
	"errors"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/missingavp"
	"github.com/skill215/diameterkit/validation"
	"github.com/skill215/diameterkit/wire"
	"testing"
)

func TestSMSMissingModes(t *testing.T) {
	for _, code := range []uint32{2018, 2010, 2017, 2011, 2015, 2014, 2012} {
		for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
			m := NewTermination(TerminationInput{})
			s := &m.ServiceInformation.SMSInformation
			switch code {
			case 2018:
				s.ClientAddressWire = avp.Optional[[]byte]{}
			case 2010:
				s.RecipientSCCPAddress = avp.Optional[[]byte]{}
			case 2017:
				s.SMSCAddress = avp.Optional[[]byte]{}
			case 2011:
				s.ReplyPathRequested = avp.Optional[int32]{}
			case 2015:
				s.SMUserDataHeader = avp.Optional[[]byte]{}
			case 2014:
				s.SMStatus = avp.Optional[[]byte]{}
			case 2012:
				s.DischargeTime = avp.Optional[avp.Time]{}
			}
			e := validateCCRSubset(m, mode)
			if !errors.Is(e, validation.ErrMissing) {
				t.Fatal(code, mode, e)
			}
			q, ok := missingavp.Quote(e, wire.DefaultLimits(), 40)
			if ok != (mode == validation.Receive) {
				t.Fatal(code, mode, e, q)
			}
			var f *validation.FieldError
			if !errors.As(e, &f) {
				t.Fatal(e)
			}
			if mode != validation.Receive && (len(f.Path) != 1 || f.Path[0].Code != 2000) {
				t.Fatal("local path", f)
			}
			*s = SMSInformation{}
			if mode == validation.Receive {
				if _, ok := missingavp.Quote(e, wire.DefaultLimits(), 40); !ok {
					t.Fatal("borrowed values")
				}
			}
		}
	}
	for _, m := range []*CCR{NewIEC(IECInput{}), NewInitial(ECURInput{})} {
		for _, code := range []uint32{2014, 2012} {
			for _, mode := range []validation.Mode{validation.Build, validation.Send, validation.Receive} {
				m.ServiceInformation.SMSInformation.SMStatus = avp.Optional[[]byte]{}
				m.ServiceInformation.SMSInformation.DischargeTime = avp.Optional[avp.Time]{}
				if code == 2014 {
					m.ServiceInformation.SMSInformation.SMStatus = avp.Some([]byte{})
				} else {
					m.ServiceInformation.SMSInformation.DischargeTime = avp.Some(avp.Time(0))
				}
				e := validateCCRSubset(m, mode)
				if !errors.Is(e, validation.ErrUnsupported) {
					t.Fatal("nonterm status scope", code, mode, e)
				}
				if _, ok := missingavp.Quote(e, wire.DefaultLimits(), 40); ok {
					t.Fatal("nonterm acquired proof")
				}
			}
		}
	}
}
