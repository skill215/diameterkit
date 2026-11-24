// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: GPT-6
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Exercise validation.Context through public consumer imports.
// SPDX-AI-Date: 2026-11-24

package contracts_test

import (
	"testing"

	"github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
)

type contextReader interface {
	Empty() bool
	Endpoints() (string, string, string, string, bool)
	IMSMOSender() (string, string, bool, bool)
}

var _ contextReader = validation.Context{}
var _ contextReader = (*validation.Context)(nil)

func TestPublicValidationContext(t *testing.T) {
	var ctx validation.Context
	if !ctx.Empty() {
		t.Fatal("zero Context has evidence")
	}
	if _, _, _, _, ok := ctx.Endpoints(); ok {
		t.Fatal("zero Context has endpoints")
	}
	request := pdu.Origin{Host: "gmsc.example", Realm: "example"}
	answer := pdu.Origin{Host: "hss.example", Realm: "example"}
	ctx, err := s6c.RoutingRole(s6c.RoutingRoleEvidence{Role: s6c.RoutingFromGMSC, Requester: request, Responder: answer})
	if err != nil {
		t.Fatal(err)
	}
	rh, rr, ah, ar, ok := ctx.Endpoints()
	if ctx.Empty() || !ok || rh != request.Host || rr != request.Realm || ah != answer.Host || ar != answer.Realm {
		t.Fatal("incorrect routing endpoints")
	}
	routing, ok := ctx.Routing()
	if !ok || routing.Role != uint8(s6c.RoutingFromGMSC) || routing.RequesterHost != rh {
		t.Fatal("incorrect routing evidence")
	}
	if _, _, _, ok := ctx.IMSMOSender(); ok {
		t.Fatal("unexpected IMS MO evidence")
	}
}
