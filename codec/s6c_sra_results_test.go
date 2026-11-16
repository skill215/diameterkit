package codec_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

func routingRoleFacts(r pdu.Message, role s6c.RoutingOriginRole, unri avp.Optional[bool]) s6c.RoutingRoleEvidence {
	b := r.Body.(*s6c.SRR)
	return s6c.RoutingRoleEvidence{Role: role, Requester: pdu.Origin{Host: b.OriginHost, Realm: b.OriginRealm}, Responder: pdu.Origin{Host: b.DestinationHost.Value, Realm: b.DestinationRealm}, UNRI: unri}
}
func routingRoleContext(t *testing.T, e s6c.RoutingRoleEvidence) validation.Context {
	t.Helper()
	c, err := s6c.RoutingRole(e)
	if err != nil {
		t.Fatal(err)
	}
	return c
}
func TestS6cSRAResultFrozenPairs(t *testing.T) {
	c := newCodec(t, codec.Config{})
	pairs := [][2]string{{"b28-srr-smea", "v24-sra-zero"}, {"b28-srr-smea", "c24-sra-nodiag"}, {"b28-srr-smea", "f24-s6c-sr-5552"}, {"b28-srr-smea", "f24-s6c-sr-5556"}, {"b28-srr-smea", "f24-s6c-sr-5557"}, {"v25-srr-two-r", "v25-diag-mme"}, {"v25-srr-two-r", "v25-diag-sgsn"}, {"b28-srr-smea", "b28-sra-absent-adjuncts"}, {"b28-srr-smea", "b28-sra-diag-all"}}
	for _, name := range []string{"b28-t4-sra-5550", "b28-t4-sra-5552", "b28-t4-sra-5556", "b28-t4-sra-5557"} {
		pairs = append(pairs, [2]string{"d21-t4-srr-r", name})
	}
	for _, pair := range pairs {
		t.Run(pair[1], func(t *testing.T) {
			r, a := decoded(t, c, pair[0]), decoded(t, c, pair[1])
			e := routingRoleFacts(r, s6c.RoutingFromGMSC, avp.Optional[bool]{})
			ctx := routingRoleContext(t, e)
			if r.Body.(*s6c.SRR).UserName.Present {
				ctx = s6cT4Context(t, t4RoutingFacts(r))
			}
			src := pdu.AnswerSources{Terminals: []pdu.Origin{e.Responder}}
			for _, p := range []validation.Report{c.ValidateAnswerContext(r, a, src, ctx), c.ValidateResponseContext(r, a, src, ctx)} {
				if !p.Complete {
					t.Fatal(p)
				}
			}
			for _, m := range []pdu.Message{r, a} {
				raw := encoded(t, c, m)
				owned, err := c.Decode(raw)
				if err != nil {
					t.Fatal(err)
				}
				clear(raw)
				if !bytes.Equal(encoded(t, c, m), encoded(t, c, owned)) {
					t.Fatal("borrowed bytes")
				}
				n, err := c.Measure(m)
				if err != nil || n != len(encoded(t, c, m)) {
					t.Fatal(n, err)
				}
				if b, err := c.DecodeBound(m.Header, n); err != nil || b < uint64(n) {
					t.Fatal(b, err)
				}
			}
			a.Header.HopByHop++
			if err := c.ValidateAnswer(r, a, src); !errors.Is(err, validation.ErrUnmatchedAnswer) {
				t.Fatal(err)
			}
		})
	}
	zero := decoded(t, c, "v24-sra-zero").Body.(*s6c.SRA)
	absent := decoded(t, c, "c24-sra-nodiag").Body.(*s6c.SRA)
	if !zero.MSCAbsentUserDiagnosticSM.Present || zero.MSCAbsentUserDiagnosticSM.Value != 0 || absent.MSCAbsentUserDiagnosticSM.Present {
		t.Fatal("presence lost")
	}
}
func TestS6cSRAAbsentRoleAndUNRI(t *testing.T) {
	c := newCodec(t, codec.Config{})
	r := decoded(t, c, "b28-srr-smea")
	src := routingSources()
	for _, hasIMSI := range []bool{false, true} {
		for _, unri := range []avp.Optional[bool]{{}, avp.Some(false), avp.Some(true)} {
			a := decoded(t, c, "v24-sra-zero")
			if hasIMSI {
				a = decoded(t, c, "b28-sra-ipsm-imsi")
			}
			ctx := routingRoleContext(t, routingRoleFacts(r, s6c.RoutingFromIPSMGW, unri))
			p := c.ValidateAnswerContext(r, a, src, ctx)
			switch {
			case hasIMSI || unri.Present && unri.Value:
				if !p.Complete {
					t.Fatal(hasIMSI, unri, p)
				}
			case !unri.Present:
				if p.State != validation.NeedsApplication {
					t.Fatal(p)
				}
			default:
				if !errors.Is(p.Err(), validation.ErrMissing) {
					t.Fatal(p)
				}
			}
		}
	}
	// None of the three SRR flags supplies the external UNRI fact.
	for flag := uint32(0); flag < 8; flag++ {
		r.Body.(*s6c.SRR).SRRFlags = avp.Some(flag)
		if p := c.ValidateAnswerContext(r, decoded(t, c, "v24-sra-zero"), src, routingRoleContext(t, routingRoleFacts(r, s6c.RoutingFromIPSMGW, avp.Optional[bool]{}))); p.State != validation.NeedsApplication {
			t.Fatal(flag, p)
		}
	}
	if p := c.ValidateAnswerContext(r, decoded(t, c, "v24-sra-zero"), src, validation.Context{}); p.State != validation.NeedsApplication {
		t.Fatal(p)
	}
	for _, name := range []string{"f24-s6c-sr-5552", "f24-s6c-sr-5556", "f24-s6c-sr-5557"} {
		if p := c.ValidateAnswerContext(r, decoded(t, c, name), src, validation.Context{}); !p.Complete {
			t.Fatal(p)
		}
	}
	base := routingRoleFacts(r, s6c.RoutingFromGMSC, avp.Optional[bool]{})
	for _, edit := range []func(*s6c.RoutingRoleEvidence){func(e *s6c.RoutingRoleEvidence) { e.Role = 0 }, func(e *s6c.RoutingRoleEvidence) { e.UNRI = avp.Some(false) }, func(e *s6c.RoutingRoleEvidence) { e.Responder.Host = strings.Repeat("x", 254) }} {
		e := base
		edit(&e)
		if _, err := s6c.RoutingRole(e); !errors.Is(err, validation.ErrContext) {
			t.Fatal(err)
		}
	}
	for _, edit := range []func(*s6c.RoutingRoleEvidence){func(e *s6c.RoutingRoleEvidence) { e.Requester.Host = "dra.example" }, func(e *s6c.RoutingRoleEvidence) { e.Responder.Realm = "other.example" }} {
		e := base
		edit(&e)
		if p := c.ValidateContext(r, routingRoleContext(t, e)); !errors.Is(p.Err(), validation.ErrContext) {
			t.Fatal(p)
		}
	}
	ctx := routingRoleContext(t, base)
	r.Body.(*s6c.SRR).DestinationHost.Present = false
	if p := c.ValidateContext(r, ctx); !errors.Is(p.Err(), validation.ErrContext) {
		t.Fatal(p)
	}
	if p := c.ValidateContext(decoded(t, c, "d21-t4-srr-r"), ctx); p.Complete {
		t.Fatal("ordinary role authorizes T4", p)
	}
}
func TestS6cSRADiagnosticsReceiveAndSend(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, pair := range [][2]string{{"v25-srr-two-r", "c25-diag-mme-reserved"}, {"b28-srr-smea", "b28-sra-diag-wrongdomain"}} {
		r, a := decoded(t, c, pair[0]), decoded(t, c, pair[1])
		ctx := routingRoleContext(t, routingRoleFacts(r, s6c.RoutingFromGMSC, avp.Optional[bool]{}))
		if err := c.ValidateAnswer(r, a, routingSources()); err != nil {
			t.Fatal(err)
		}
		if p := c.ValidateAnswerContext(r, a, routingSources(), ctx); p.Complete || !errors.Is(p.Err(), validation.ErrUnsupported) {
			t.Fatal(p)
		}
		if _, err := c.Append(nil, a); !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"n25-diag-mme-256", "b28-sra-diag-dup", "b28-sra-diag-width", "v24-s6c-sr-r"} {
		if _, err := c.Decode(fixture(t, name)); err == nil {
			t.Fatal(name)
		}
	}
	// Exercise each selected domain and reserved/bad value without generating preconditions from codec.
	for domain := range 3 {
		for value := uint32(0); value <= 256; value++ {
			a := decoded(t, c, "c24-sra-nodiag")
			b := a.Body.(*s6c.SRA)
			v := avp.Some(value)
			masks := []uint16{1 << 13, 1<<0 | 1<<1 | 1<<2 | 1<<3 | 1<<4 | 1<<9 | 1<<13, 1<<2 | 1<<5 | 1<<6 | 1<<7 | 1<<8 | 1<<10 | 1<<13}
			switch domain {
			case 0:
				b.MMEAbsentUserDiagnosticSM = v
			case 1:
				b.MSCAbsentUserDiagnosticSM = v
			case 2:
				b.SGSNAbsentUserDiagnosticSM = v
			}
			err := c.Validate(a)
			valid := value <= 13 && masks[domain]&(1<<value) != 0
			if (err == nil) != valid {
				t.Fatal(domain, value, err)
			}
		}
	}
}
func TestS6cSRAResultIsolation(t *testing.T) {
	c := newCodec(t, codec.Config{})
	for _, mutate := range []func(*s6c.SRA){
		func(a *s6c.SRA) { a.ResultCode = avp.Some(uint32(2001)) }, func(a *s6c.SRA) { a.ExperimentalResult.Value.VendorID = 0 },
		func(a *s6c.SRA) { a.ExperimentalResult.Value.ExperimentalResultCode = 5551 }, func(a *s6c.SRA) { a.UserName = avp.Some("opaque-token") },
		func(a *s6c.SRA) { a.ServingNode = avp.Some(routingNode(0)) }, func(a *s6c.SRA) { a.LMSI = avp.Some([]byte{1, 2, 3, 4}) },
	} {
		a := decoded(t, c, "v24-sra-zero")
		mutate(a.Body.(*s6c.SRA))
		if err := c.Validate(a); err == nil {
			t.Fatal("invalid result branch")
		}
	}
	for _, code := range []uint32{5001, 5552, 5556, 5557} {
		a := decoded(t, c, "v24-sra-zero")
		a.Body.(*s6c.SRA).ExperimentalResult.Value.ExperimentalResultCode = code
		if err := c.Validate(a); !errors.Is(err, validation.ErrUnsupported) {
			t.Fatal("diagnostic outside selected result", err)
		}
		a.Body.(*s6c.SRA).MSCAbsentUserDiagnosticSM.Present = false
		a.Body.(*s6c.SRA).UserName = avp.Some("001010123456789")
		if err := c.Validate(a); !errors.Is(err, validation.ErrField) {
			t.Fatal("unrelated failure identity", err)
		}
	}
	// SRA result support must not leak into the other S6c commands.
	for _, code := range []uint32{5550, 5552, 5556, 5557} {
		for _, name := range []string{"d21-t4-rdr-f", "d21-t4-alr-a"} {
			m := decoded(t, c, name)
			switch a := m.Body.(type) {
			case *s6c.RDA:
				a.ExperimentalResult = avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: code})
			case *s6c.ALA:
				a.ResultCode.Present = false
				a.ExperimentalResult = avp.Some(s6c.ExperimentalResult{VendorID: 10415, ExperimentalResultCode: code})
			}
			if err := c.Validate(m); !errors.Is(err, validation.ErrUnsupported) {
				t.Fatal(code, err)
			}
		}
	}
}
