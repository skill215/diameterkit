package s6c_test

import (
	"context"
	"errors"
	diameter "github.com/skill215/diameterkit"
	app "github.com/skill215/diameterkit/app/s6c"
	"github.com/skill215/diameterkit/avp"
	client "github.com/skill215/diameterkit/client/s6c"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/validation"
	"testing"
)

func TestNoAnswerAndIncompletePreserveExchange(t *testing.T) {
	c := new(client.Client)
	for _, tc := range []diameter.Result{
		{Err: context.Canceled, Outcome: diameter.WrittenNoAnswer},
		{Outcome: diameter.AnswerReceived, Answer: pdu.Message{Body: &app.SRA{}}, Validation: validation.Report{State: validation.NeedsApplication}},
		{Outcome: diameter.NotWritten},
	} {
		tc.RequestID.Sequence = 12
		r, e := c.InterpretRouting(app.RoutingContext{}, tc)
		if e == nil || r.Routing != nil || r.Exchange.RequestID.Sequence != 12 || r.Exchange.Answer.Body != tc.Answer.Body || r.Exchange.Outcome != tc.Outcome {
			t.Fatal(r, e)
		}
		if tc.Err != nil && !errors.Is(e, tc.Err) {
			t.Fatal(e)
		}
	}
}
func TestInvalidClient(t *testing.T) {
	if _, e := client.New(nil); e == nil {
		t.Fatal("nil route")
	}
	var c *client.Client
	q := app.RoutingQuery{SessionID: "s;1", MSISDN: avp.Some("123"), SCAddress: "456"}
	if _, e := c.QueryRouting(context.Background(), q, diameter.RequestOptions{}); !errors.Is(e, validation.ErrContext) {
		t.Fatal(e)
	}
	if _, e := c.SubmitRouting(context.Background(), q, diameter.RequestOptions{}, nil); !errors.Is(e, validation.ErrContext) {
		t.Fatal(e)
	}
}
