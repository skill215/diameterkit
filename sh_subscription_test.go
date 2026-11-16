package diameter

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/skill215/diameterkit/app/sh"
	"github.com/skill215/diameterkit/app/sh/shdata"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
)

func subscription(ref, action uint32) *sh.SNR {
	b := shBody(308).(*sh.SNR)
	b.SessionID = fmt.Sprintf("as.example;subscription;%d;%d", ref, action)
	b.DataReference = []uint32{ref}
	b.SubsReqType = action
	b.ExpiryTime = avp.Optional[avp.Time]{}
	if ref != 25 {
		b.OneTimeNotification = avp.Optional[uint32]{}
	}
	return b
}
func imsData(t *testing.T, ref uint32) []byte {
	t.Helper()
	var b []byte
	var e error
	switch ref {
	case 11:
		b, e = shdata.MarshalIMSUserState(1, shdata.DefaultLimits())
	case 12:
		b, e = shdata.MarshalSCSCFName("", shdata.DefaultLimits())
	default:
		b, e = shdata.MarshalReachability(shdata.Reachability{MME: avp.Some(uint8(0))}, shdata.DefaultLimits())
	}
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func TestShSubscriptionInboundResultsAndData(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		hbh := uint32(1900)
		for _, ref := range []uint32{11, 12, 25} {
			for _, action := range []uint32{0, 1} {
				for _, kind := range []string{"success", "no-data", "experimental", "E"} {
					req := shRequest(t, n, co, 308, hbh)
					hbh++
					b := req.Body.(*sh.SNR)
					b.DataReference = []uint32{ref}
					b.SubsReqType = action
					if ref != 25 {
						b.OneTimeNotification = avp.Optional[uint32]{}
						b.SendDataIndication = avp.Some(uint32(1))
					}
					if e := writeWire(ctxTest(t), remote, co, req); e != nil {
						t.Fatal(e)
					}
					in, e := p.Receive(ctxTest(t))
					if e != nil {
						t.Fatal(e)
					}
					got := in.Message().Body.(*sh.SNR)
					if len(got.DataReference) != 1 || got.DataReference[0] != ref || got.SubsReqType != action || got.SendDataIndication != b.SendDataIndication {
						t.Fatal(got)
					}
					var answer pdu.Body = shAnswer(308, 2001, false)
					switch kind {
					case "success":
						a := answer.(*sh.SNA)
						a.UserData = avp.Some(imsData(t, ref))
						a.ExpiryTime = avp.Some(avp.Time(0))
					case "experimental":
						answer = shAnswer(308, 5104, true)
					case "E":
						answer = &pdu.ErrorAnswer{ResultCode: 3004}
					}
					if r, e := in.Respond(ctxTest(t), answer); e != nil || !r.Committed {
						t.Fatal(r, e)
					}
					a := duplicateAnswer(t, n, co, remote, req)
					if kind == "no-data" {
						b := a.Body.(*sh.SNA)
						if b.UserData.Present || b.ExpiryTime.Present {
							t.Fatal("copied SHOULD data/expiry")
						}
					}
					if kind == "success" {
						d, e := shdata.Parse(a.Body.(*sh.SNA).UserData.Value, shdata.DefaultLimits())
						if e != nil {
							t.Fatal(e)
						}
						use := shdata.ReachabilityNotification
						if ref == 11 {
							use = shdata.IMSUserStateData
						}
						if ref == 12 {
							use = shdata.SCSCFNameData
						}
						r, e := d.Check(use)
						if e != nil || !r.Complete {
							t.Fatal(r, e)
						}
					}
					waitFor(t, func() bool { return n.Resources().InboundOutstanding == 0 })
				}
			}
		}
	})
}
func TestShUnsubscribeDoesNotFenceNotifications(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, false))
		p, remote := networkConnectNode(t, n, co, tls)
		port := portTest(t, n, 3)
		ids := map[uint32]RequestID{}
		reqs := map[uint32]pdu.Message{}
		for _, ref := range []uint32{11, 12, 25} {
			id, e := p.Submit(ctxTest(t), subscription(ref, 1), RequestOptions{}, port)
			if e != nil {
				t.Fatal(e)
			}
			ids[ref] = id
			reqs[ref] = routeRead(t, co, remote)
		}
		notify := func(hbh, ref uint32) {
			req := shRequest(t, n, co, 309, hbh)
			req.Body.(*sh.PNR).UserData = imsData(t, ref)
			if e := writeWire(ctxTest(t), remote, co, req); e != nil {
				t.Fatal(e)
			}
			in, e := p.Receive(ctxTest(t))
			if e != nil {
				t.Fatal(e)
			}
			if r, e := in.Respond(ctxTest(t), shAnswer(309, 2001, false)); e != nil || !r.Committed {
				t.Fatal(r, e)
			}
			duplicateAnswer(t, n, co, remote, req)
		}
		notify(2000, 11) // Already in flight before any unsubscribe answer.
		if e := n.Cancel(ids[11]); e != nil {
			t.Fatal(e)
		}
		done, e := port.Receive(ctxTest(t))
		if e != nil || done.RequestID != ids[11] || !errors.Is(done.Result.Err, context.Canceled) {
			t.Fatal(done, e)
		}
		for _, ref := range []uint32{25, 12, 11} {
			ans, e := co.CompleteAnswer(reqs[ref], shAnswer(308, 2001, false), serverOrigin)
			if e != nil {
				t.Fatal(e)
			}
			if e = writeWire(ctxTest(t), remote, co, ans); e != nil {
				t.Fatal(e)
			}
			if ref != 11 {
				done, e = port.Receive(ctxTest(t))
				if e != nil || done.RequestID != ids[ref] || done.Result.Outcome != AnswerReceived {
					t.Fatal(done, e)
				}
			}
		}
		notify(2001, 12)
		notify(2002, 11) // Successful or canceled SNR creates no PNR filter.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, e = port.Receive(ctx); !errors.Is(e, context.DeadlineExceeded) {
			t.Fatal("late SNA completed twice", e)
		}
		waitFor(t, func() bool { return n.Resources().Outstanding == 0 && n.Resources().InboundOutstanding == 0 })
	})
}
func TestShSubscriptionDuplicatePresenceConflict(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, shConfig(t, true))
		p, remote := networkConnectNode(t, n, co, tls)
		req := shRequest(t, n, co, 308, 2100)
		b := req.Body.(*sh.SNR)
		b.DataReference = []uint32{11}
		b.OneTimeNotification = avp.Optional[uint32]{}
		b.SendDataIndication = avp.Some(uint32(0))
		if e := writeWire(ctxTest(t), remote, co, req); e != nil {
			t.Fatal(e)
		}
		in, e := p.Receive(ctxTest(t))
		if e != nil {
			t.Fatal(e)
		}
		dup, e := co.Snapshot(req)
		if e != nil {
			t.Fatal(e)
		}
		dup.Header.HopByHop++
		dup.Header.Flags |= pdu.RetransmittedFlag
		if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
			t.Fatal(e)
		}
		noDuplicateDelivery(t, p)
		if r, e := in.Respond(ctxTest(t), shAnswer(308, 2001, false)); e != nil || !r.Committed {
			t.Fatal(r, e)
		}
		for range 2 {
			ans := routeRead(t, co, remote)
			original := req
			if ans.Header.HopByHop == dup.Header.HopByHop {
				original = dup
			}
			if e = co.ValidateAnswer(original, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
				t.Fatal(e)
			}
		}
		// Original identity is unchanged: absence is not equivalent to explicit 0,
		// and changing an existing request into unsubscribe is a content conflict.
		for _, change := range []func(*sh.SNR){func(b *sh.SNR) { b.SendDataIndication = avp.Optional[uint32]{} }, func(b *sh.SNR) { b.SubsReqType = 1 }} {
			dup, e = co.Snapshot(req)
			if e != nil {
				t.Fatal(e)
			}
			dup.Header.HopByHop += 10
			change(dup.Body.(*sh.SNR))
			if e = writeWire(ctxTest(t), remote, co, dup); e != nil {
				t.Fatal(e)
			}
			a := duplicateAnswer(t, n, co, remote, dup)
			if a.Body.(*pdu.ErrorAnswer).ResultCode != 5012 {
				t.Fatal(a)
			}
			noDuplicateDelivery(t, p)
		}
	})
}
