// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Tests and benchmarks generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/base"
	"github.com/skill215/diameterkit/codec"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/transport"
)

func smallSchedule() SchedulerConfig {
	return SchedulerConfig{QuantumBytes: 64, ControlWeight: 4, AnswerWeight: 2, RequestWeight: 1, EmergencyWeight: 1}
}
func schedulerConfig(t *testing.T) NodeConfig {
	cfg := inboundConfig(t)
	cfg.Scheduler = smallSchedule()
	cfg.Peers[0].MaxPending = 16
	cfg.Resources.Classes[0].Slots = 32
	cfg.Peers[0].MaxBytes = 64 << 20
	return cfg
}
func TestDRRConfigurationAndOwnedInfo(t *testing.T) {
	cfg := inboundConfig(t)
	cfg.Scheduler.RequestClasses = []ClassWeight{{ID: "sms", Weight: 3}}
	n, _ := newInboundNode(t, cfg)
	info := n.SchedulerInfo()
	if info.QuantumBytes != 16384 || info.ControlWeight != 4 || info.AnswerWeight != 2 || info.RequestWeight != 1 || info.RequestClasses[0].Weight != 3 || info.AnswerClasses[0].Weight != 1 {
		t.Fatal(info)
	}
	cfg.Scheduler.RequestClasses[0].Weight = 63
	info.AnswerClasses[0].Weight = 50
	if n.SchedulerInfo().RequestClasses[0].Weight != 3 || n.SchedulerInfo().AnswerClasses[0].Weight != 1 {
		t.Fatal("mutable schedule escaped")
	}
	for _, mutate := range []func(*SchedulerConfig){
		func(c *SchedulerConfig) { c.QuantumBytes = 63 }, func(c *SchedulerConfig) { c.QuantumBytes = 1 << 21 }, func(c *SchedulerConfig) { c.ControlWeight = 0 }, func(c *SchedulerConfig) { c.AnswerWeight = 65 }, func(c *SchedulerConfig) { c.EmergencyWeight = 0 },
		func(c *SchedulerConfig) { c.RequestClasses = []ClassWeight{{"missing", 1}} }, func(c *SchedulerConfig) { c.RequestClasses = []ClassWeight{{"sms", 0}} }, func(c *SchedulerConfig) { c.AnswerClasses = []ClassWeight{{"sms", 65}} }, func(c *SchedulerConfig) { c.RequestClasses = []ClassWeight{{"sms", 1}, {"sms", 2}} },
	} {
		c := schedulerConfig(t)
		mutate(&c.Scheduler)
		if bad, e := NewNode(c); e == nil {
			_ = bad.Close()
			t.Fatal("invalid schedule accepted", c.Scheduler)
		}
	}
}
func TestDRRByteSharesAndNoStarvation(t *testing.T) {
	for _, q := range [][]uint64{{256, 128, 64}, {64, 192, 128}} {
		r := newDRR(q, 4096)
		costs := []uint64{72, 512, 2048}
		var sent [3]uint64
		var gap, maxGap [3]int
		for range 30000 {
			i := r.peek(costs)
			r.debit(i, costs[i])
			sent[i] += costs[i]
			for j := range costs {
				if i == j {
					gap[j] = 0
				} else {
					gap[j]++
					maxGap[j] = max(maxGap[j], gap[j])
				}
			}
		}
		for i := range costs {
			if sent[i] == 0 || maxGap[i] > 200 {
				t.Fatal("starvation", sent, maxGap)
			}
			a, b := int64(sent[i]*q[0]), int64(sent[0]*q[i])
			delta := a - b
			if delta < 0 {
				delta = -delta
			}
			if delta > int64(8192*max(q[0], q[i])) {
				t.Fatal("packet-count rather than byte share", sent, q)
			}
		}
	}
}

// A deliberately slow round-by-round oracle checks the empty-round arithmetic,
// including changing frame sizes and class weights; it does not fast-forward.
func TestDRRFastForwardMatchesRoundOracle(t *testing.T) {
	q := []uint64{64, 192, 320}
	r := newDRR(q, 1<<20)
	def := make([]uint64, 3)
	cursor := 0
	credited := false
	rng := rand.New(rand.NewSource(215))
	for step := 0; step < 3000; step++ {
		costs := []uint64{uint64(20 + 4*rng.Intn(4000)), uint64(20 + 4*rng.Intn(300)), uint64(20 + 4*rng.Intn(1000))}
		want := -1
		for want < 0 {
			if !credited {
				def[cursor] += q[cursor]
				credited = true
			}
			if def[cursor] >= costs[cursor] {
				want = cursor
				def[cursor] -= costs[cursor]
			} else {
				cursor = (cursor + 1) % 3
				credited = false
			}
		}
		got := r.peek(costs)
		if got != want {
			t.Fatal(step, got, want)
		}
		r.debit(got, costs[got])
		for i := range def {
			if def[i] != r.deficit[i] {
				t.Fatal(step, def, r.deficit)
			}
		}
	}
	large := newDRR([]uint64{64}, 16<<20)
	for range 100 {
		if i := large.peek([]uint64{16 << 20}); i != 0 {
			t.Fatal(i)
		}
		large.debit(0, 16<<20)
	}
}
func schedulerHarness(t *testing.T) (*connection, *codec.Codec) {
	cfg := testConfig(t)
	cfg.Resources.Classes = []ClassConfig{{"a", 1 << 20, 4}, {"b", 1 << 20, 4}}
	cfg.Resources.InboundClasses = []ClassConfig{{"a", 1 << 20, 4}, {"b", 1 << 20, 4}}
	cfg.Scheduler = smallSchedule()
	cfg.Scheduler.RequestClasses = []ClassWeight{{"b", 3}}
	cfg.Scheduler.AnswerClasses = []ClassWeight{{"b", 3}}
	plan, e := compileScheduler(cfg.Scheduler, cfg.Resources, cfg.Limits.MaxFrameBytes)
	if e != nil {
		t.Fatal(e)
	}
	n := &Node{scheduler: plan}
	p := &Peer{node: n}
	co, e := codec.New(cfg.Bundle, codec.Config{Limits: cfg.Limits})
	if e != nil {
		t.Fatal(e)
	}
	return &connection{peer: p, schedule: plan.newWriter()}, co
}
func schedulingFrames(t *testing.T, co *codec.Codec) (control, request, answer []byte) {
	dwr := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 280, HopByHop: 1, EndToEnd: 2}, Body: &base.DWR{OriginHost: "local.example", OriginRealm: "example"}}
	control = reqBytes(t, co, dwr)
	body := requestBody()
	body.Extra = []avp.Raw{{Key: avp.Key{Code: 9999}, Data: make([]byte, 360)}}
	req, e := co.CompleteRequest(body, pdu.RequestEnvelope{OriginHost: "local.example", OriginRealm: "example", DestinationRealm: "example", HopByHop: 10, EndToEnd: 11})
	if e != nil {
		t.Fatal(e)
	}
	request = reqBytes(t, co, req)
	ans, e := co.CompleteAnswer(req, inboundFailure(), serverOrigin)
	if e != nil {
		t.Fatal(e)
	}
	answer = reqBytes(t, co, ans)
	return
}
func harnessRequest(c *connection, class int, frame []byte, tag uint32) *exchange {
	x := &exchange{scheduleClass: class, frame: frame, hbh: tag}
	c.enqueueRequest(x)
	return x
}
func harnessAnswer(c *connection, class ClassID, frame []byte, emergency bool) *Inbound {
	in := &Inbound{policy: inboundPolicy{InboundPolicy: InboundPolicy{Class: class}}, frame: frame, emergency: emergency}
	c.enqueueAnswer(in)
	return in
}
func harnessConsume(c *connection, kind, class int, cost uint64) {
	c.debitWrite(kind, class, cost)
	switch kind {
	case scheduleControl:
		c.popControl()
	case scheduleRequest:
		for _, x := range c.queue {
			if x.scheduleClass == class {
				c.removeRequest(x)
				break
			}
		}
	case scheduleAnswer:
		for _, in := range c.answers {
			if in.scheduleClass == class {
				c.removeAnswer(in)
				break
			}
		}
	}
}
func TestDRRHierarchyUsesEncodedBytes(t *testing.T) {
	c, co := schedulerHarness(t)
	ctrl, req, ans := schedulingFrames(t, co)
	for range 2 {
		c.enqueueControl(controlJob{prepared: &preparedControl{frame: ctrl}})
		for i, id := range []ClassID{"a", "b"} {
			harnessRequest(c, i, req, 0)
			harnessAnswer(c, id, ans, false)
		}
	}
	var categories [3]uint64
	var requests, answers [2]uint64
	for range 40000 {
		kind, class, cost := c.pickWrite()
		harnessConsume(c, kind, class, cost)
		categories[kind] += cost
		switch kind {
		case scheduleControl:
			c.enqueueControl(controlJob{prepared: &preparedControl{frame: ctrl}})
		case scheduleRequest:
			requests[class] += cost
			harnessRequest(c, class, req, 0)
		case scheduleAnswer:
			answers[class] += cost
			harnessAnswer(c, []ClassID{"a", "b"}[class], ans, false)
		}
	}
	check := func(a, b uint64, weight uint64) {
		t.Helper()
		delta := int64(a) - int64(b*weight)
		if delta < 0 {
			delta = -delta
		}
		if delta > 8192 {
			t.Fatal("weighted byte share", a, b, weight, categories, requests, answers)
		}
	}
	check(categories[0], categories[2], 4)
	check(categories[1], categories[2], 2)
	check(requests[1], requests[0], 3)
	check(answers[1], answers[0], 3)
}
func TestDRREmptyRefillCannotJumpOtherLanes(t *testing.T) {
	c, co := schedulerHarness(t)
	ctrl, req, ans := schedulingFrames(t, co)
	c.enqueueControl(controlJob{prepared: &preparedControl{frame: ctrl}})
	for i, id := range []ClassID{"a", "b"} {
		harnessRequest(c, i, req, 0)
		harnessAnswer(c, id, ans, false)
	}
	var count [3]int
	var reqCount, ansCount [2]int
	for range 300 {
		kind, class, cost := c.pickWrite()
		harnessConsume(c, kind, class, cost)
		count[kind]++
		// The selected lane empties and refills while the simulated writer holds it.
		switch kind {
		case scheduleControl:
			c.enqueueControl(controlJob{prepared: &preparedControl{frame: ctrl}})
		case scheduleRequest:
			reqCount[class]++
			harnessRequest(c, class, req, 0)
		case scheduleAnswer:
			ansCount[class]++
			harnessAnswer(c, []ClassID{"a", "b"}[class], ans, false)
		}
	}
	for _, n := range append(append(count[:], reqCount[:]...), ansCount[:]...) {
		if n == 0 {
			t.Fatal("refill starved another lane", count, reqCount, ansCount)
		}
	}
}
func TestDRRClassFIFOAndCancellationCredit(t *testing.T) {
	c, co := schedulerHarness(t)
	_, req, _ := schedulingFrames(t, co)
	var want [2]uint32
	for tag := uint32(1); tag <= 12; tag++ {
		for class := 0; class < 2; class++ {
			harnessRequest(c, class, req, tag)
		}
	}
	for len(c.queue) > 0 {
		kind, class, cost := c.pickWrite()
		if kind != scheduleRequest {
			t.Fatal(kind)
		}
		for _, x := range c.queue {
			if x.scheduleClass == class {
				want[class]++
				if x.hbh != want[class] {
					t.Fatal("class FIFO changed", class, x.hbh, want)
				}
				break
			}
		}
		harnessConsume(c, kind, class, cost)
	}
	if c.schedule.category.deficit[scheduleRequest] != 0 || c.schedule.request.deficit[0] != 0 || c.schedule.request.deficit[1] != 0 {
		t.Fatal("idle credit retained")
	}
	x := harnessRequest(c, 0, req, 1)
	_, _, _ = c.pickWrite()
	c.removeRequest(x)
	if c.schedule.category.deficit[scheduleRequest] != 0 || c.schedule.request.deficit[0] != 0 {
		t.Fatal("cancel retained credit")
	}
}
func TestDRREmergencyAndReplayCost(t *testing.T) {
	c, co := schedulerHarness(t)
	_, _, ans := schedulingFrames(t, co)
	ordinary := harnessAnswer(c, "a", ans, false)
	emergency := harnessAnswer(c, "", ans, true)
	if ordinary.scheduleClass == emergency.scheduleClass || emergency.scheduleClass != 2 {
		t.Fatal("emergency shared business lane")
	}
	replay := &Inbound{replay: &duplicateRecord{result: ans[20:]}, proxy: make([]byte, 160), policy: inboundPolicy{InboundPolicy: InboundPolicy{Class: "b"}}}
	c.enqueueAnswer(replay)
	if answerCost(replay) != uint64(len(ans)+160) {
		t.Fatal("proxy bytes omitted")
	}
	seen := map[int]bool{}
	for len(c.answers) > 0 {
		kind, class, cost := c.pickWrite()
		if kind != scheduleAnswer {
			t.Fatal(kind)
		}
		if class == 1 && cost != uint64(len(ans)+160) {
			t.Fatal(cost)
		}
		seen[class] = true
		harnessConsume(c, kind, class, cost)
	}
	if len(seen) != 3 {
		t.Fatal(seen)
	}
}
func TestDRRSelectionAllocatesNoPayloadOrQueue(t *testing.T) {
	c, co := schedulerHarness(t)
	ctrl, req, ans := schedulingFrames(t, co)
	c.enqueueControl(controlJob{prepared: &preparedControl{frame: ctrl}})
	harnessRequest(c, 0, req, 0)
	harnessAnswer(c, "b", ans, false)
	if got := testing.AllocsPerRun(1000, func() { _, _, _ = c.pickWrite() }); got != 0 {
		t.Fatal("selection allocated", got)
	}
}

func drrBlockedWriter(t *testing.T, n *Node, co *codec.Codec, tls bool) (*Peer, transport.Conn, *inboundWriteBarrier, func()) {
	t.Helper()
	a, b := networkAdapters(t, tls)
	barrier := &inboundWriteBarrier{Conn: a, request: true, entered: make(chan struct{}), release: make(chan struct{})}
	var once sync.Once
	release := func() { once.Do(func() { close(barrier.release) }) }
	t.Cleanup(release)
	p := connectInboundTransport(t, n, co, barrier, b)
	return p, b, barrier, release
}
func waitBarrier(t *testing.T, b *inboundWriteBarrier) {
	t.Helper()
	select {
	case <-b.entered:
	case <-time.After(time.Second):
		t.Fatal("writer barrier")
	}
}
func TestDRRRealMixedFramesAndDrain(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		for _, shutdown := range []bool{false, true} {
			t.Run(fmt.Sprint(shutdown), func(t *testing.T) {
				n, co := newInboundNode(t, schedulerConfig(t))
				p, remote, barrier, release := drrBlockedWriter(t, n, co, tls)
				port := portTest(t, n, 8)
				_, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
				if e != nil {
					t.Fatal(e)
				}
				first := routeRead(t, co, remote)
				waitBarrier(t, barrier)
				for i := 0; i < 4; i++ {
					body := requestBody()
					body.SessionID = fmt.Sprintf("local;%d", i)
					if i%2 == 0 {
						body.Extra = []avp.Raw{{Key: avp.Key{Code: 9999}, Data: make([]byte, 512)}}
					}
					if _, e = p.Submit(ctxTest(t), body, RequestOptions{}, port); e != nil {
						t.Fatal(e)
					}
				}
				responses := make(chan error, 3)
				for i := 0; i < 3; i++ {
					in, _ := receiveInbound(t, n, p, remote, co, uint32(100+i))
					go func() { _, err := in.Respond(ctxTest(t), inboundFailure()); responses <- err }()
					<-in.Done()
				}
				for i := 0; i < 2; i++ {
					dwr := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 280, HopByHop: uint32(800 + i), EndToEnd: uint32(900 + i)}, Body: &base.DWR{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm}}
					if e = writeWire(ctxTest(t), remote, co, dwr); e != nil {
						t.Fatal(e)
					}
					want := i + 1
					waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return len(p.conn.controls) == want })
				}
				var closed chan error
				if shutdown {
					closed = make(chan error, 1)
					go func() { closed <- n.Shutdown(ctxTest(t)) }()
					waitFor(t, func() bool { return p.Info().Lifecycle == Draining })
				}
				release()
				routeReply(t, co, remote, first)
				var got [3]int
				for range 9 {
					msg := routeRead(t, co, remote)
					if msg.Header.ApplicationID == 0 {
						if msg.Header.CommandCode != 280 {
							t.Fatal("control bypassed drain", msg)
						}
						got[0]++
					} else if msg.Header.Flags&pdu.RequestFlag != 0 {
						got[2]++
						routeReply(t, co, remote, msg)
					} else {
						got[1]++
					}
				}
				if got != [3]int{2, 3, 4} {
					t.Fatal(got)
				}
				for range 3 {
					if e = <-responses; e != nil {
						t.Fatal(e)
					}
				}
				for range 5 {
					c, e := port.Receive(ctxTest(t))
					if e != nil || c.Result.Outcome != AnswerReceived {
						t.Fatal(c, e)
					}
				}
				if shutdown {
					dpr := routeRead(t, co, remote)
					if dpr.Header.CommandCode != 282 {
						t.Fatal(dpr)
					}
					dpa, e := co.CompleteAnswer(dpr, &base.DPA{ResultCode: 2001}, draOrigin)
					if e != nil {
						t.Fatal(e)
					}
					if e = writeWire(ctxTest(t), remote, co, dpa); e != nil {
						t.Fatal(e)
					}
					if e = <-closed; e != nil {
						t.Fatal(e)
					}
				}
				waitFor(t, func() bool { s := n.Resources(); return s.OutboundBytes == 0 && s.InboundBytes == 0 })
			})
		}
	})
}
func TestDRRCancelAndClosePreserveWriterBudget(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, schedulerConfig(t))
		p, remote, barrier, release := drrBlockedWriter(t, n, co, tls)
		port := portTest(t, n, 4)
		first, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, remote)
		waitBarrier(t, barrier)
		id, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		n.mu.Lock()
		held := n.requests[first].charge.Capacity
		n.mu.Unlock()
		_ = n.Cancel(id)
		got, e := port.Receive(ctxTest(t))
		if e != nil || got.RequestID != id || got.Result.Outcome != NotWritten {
			t.Fatal(got, e)
		}
		n.mu.Lock()
		zero := len(p.conn.queue) == 0 && p.conn.schedule.category.deficit[scheduleRequest] == 0 && p.conn.schedule.request.deficit[0] == 0
		n.mu.Unlock()
		if !zero {
			t.Fatal("canceled lane retained credit")
		}
		in, _ := receiveInbound(t, n, p, remote, co, 300)
		answer := make(chan ResponseResult, 1)
		go func() { r, _ := in.Respond(ctxTest(t), inboundFailure()); answer <- r }()
		<-in.Done()
		_ = n.Close()
		if r := <-answer; r.PrimaryWrite.State != ResponseNotWritten {
			t.Fatal(r)
		}
		if n.Resources().OutboundBytes != held {
			t.Fatal("writer borrower lost charge", n.Resources(), held)
		}
		release()
		got, e = port.Receive(ctxTest(t))
		if e != nil || got.RequestID != first || got.Result.Outcome != WrittenNoAnswer {
			t.Fatal(got, e)
		}
		if e = n.Wait(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		if n.Resources().OutboundBytes != 0 || n.Resources().InboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}

// The current slice has one application request command per binding. Two peers
// exercise independent hard class budgets without inventing new wire commands.
func TestDRRClassBudgetIsolation(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := twoDuplicatePeers(schedulerConfig(t))
		cfg.Resources.TransportBytes = 128 << 20
		cfg.Resources.NodeBytes = 512 << 20
		cfg.Resources.Classes = []ClassConfig{{"a", 32 << 20, 1}, {"b", 32 << 20, 1}}
		cfg.Resources.InboundClasses = []ClassConfig{{"a", 32 << 20, 1}, {"b", 32 << 20, 1}}
		cfg.Scheduler.RequestClasses = []ClassWeight{{"a", 64}, {"b", 1}}
		cfg.Scheduler.AnswerClasses = []ClassWeight{{"a", 64}, {"b", 1}}
		for i, class := range []ClassID{"a", "b"} {
			cfg.Peers[i].Commands = append([]CommandPolicy(nil), cfg.Peers[i].Commands...)
			cfg.Peers[i].Commands[0].Class = class
			cfg.Peers[i].InboundCommands[0].Class = class
		}
		n, co := newInboundNode(t, cfg)
		p1, r1 := duplicateConnect(t, n, co, tls, "dra")
		p2, r2 := duplicateConnect(t, n, co, tls, "dra2")
		port1, port2 := portTest(t, n, 2), portTest(t, n, 2)
		if _, e := p1.Submit(ctxTest(t), requestBody(), RequestOptions{}, port1); e != nil {
			t.Fatal(e)
		}
		req1 := routeRead(t, co, r1)
		routeReply(t, co, r1, req1)
		waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return port1.count == 1 })
		before := n.Resources()
		if _, e := p1.Submit(ctxTest(t), requestBody(), RequestOptions{}, port1); !errors.Is(e, ErrBackpressure) {
			t.Fatal(e)
		}
		if n.Resources() != before {
			t.Fatal("failed admission changed budget")
		}
		if _, e := p2.Submit(ctxTest(t), requestBody(), RequestOptions{}, port2); e != nil {
			t.Fatal(e)
		}
		routeReply(t, co, r2, routeRead(t, co, r2))
		if c, e := port2.Receive(ctxTest(t)); e != nil || c.Result.Outcome != AnswerReceived {
			t.Fatal(c, e)
		}
		n.mu.Lock()
		isolated := n.classes["a"].UsedSlots == 1 && n.classes["a"].UsedBytes > 0 && n.classes["b"].UsedSlots == 0 && n.classes["b"].UsedBytes == 0
		n.mu.Unlock()
		if !isolated {
			t.Fatal("class quota borrowed or held completion lost its charge")
		}
		if _, e := port1.Receive(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		in1, _ := receiveInbound(t, n, p1, r1, co, 100)
		in2, req2 := receiveInbound(t, n, p2, r2, co, 200)
		if _, e := in2.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		ans := routeRead(t, co, r2)
		if e := co.ValidateAnswer(req2, ans, pdu.AnswerSources{Terminals: []pdu.Origin{n.cfg.Origin}}); e != nil {
			t.Fatal(e)
		}
		n.mu.Lock()
		isolated = n.inboundClasses["a"].UsedSlots == 1 && n.inboundClasses["a"].UsedBytes > 0 && n.inboundClasses["b"].UsedSlots == 0 && n.inboundClasses["b"].UsedBytes == 0
		n.mu.Unlock()
		if !isolated {
			t.Fatal("inbound classes borrowed quota")
		}
		if _, e := in1.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, r1)
		if s := n.Resources(); s.InboundBytes != 0 || s.OutboundBytes != 0 {
			t.Fatal(s)
		}
	})
}
func TestDRRRealEmergencyLane(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := rejectConfig(t)
		cfg.Scheduler = smallSchedule()
		cfg.Resources.InboundClasses[0].Slots = 1
		n, co := newInboundNode(t, cfg)
		p, remote, barrier, release := drrBlockedWriter(t, n, co, tls)
		port := portTest(t, n, 2)
		if _, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); e != nil {
			t.Fatal(e)
		}
		first := routeRead(t, co, remote)
		waitBarrier(t, barrier)
		in, _ := receiveInbound(t, n, p, remote, co, 100)
		answer := make(chan error, 1)
		go func() { _, e := in.Respond(ctxTest(t), inboundFailure()); answer <- e }()
		<-in.Done()
		req := inboundRequest(t, n, co, 101)
		if e := writeWire(ctxTest(t), remote, co, req); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return len(p.conn.answers) == 2 })
		n.mu.Lock()
		counts := append([]int(nil), p.conn.schedule.answerCounts...)
		n.mu.Unlock()
		if len(counts) != 2 || counts[0] != 1 || counts[1] != 1 {
			t.Fatal(counts)
		}
		s := n.Resources()
		if s.InboundOutstanding != 1 || s.RejectOutstanding != 1 {
			t.Fatal(s)
		}
		release()
		seenNormal, seenEmergency := false, false
		for range 2 {
			msg := routeRead(t, co, remote)
			if msg.Header.Flags&pdu.ErrorFlag != 0 {
				e, ok := msg.Body.(*pdu.ErrorAnswer)
				if !ok || e.ResultCode != 3004 || !pdu.MatchesAnswerHeader(req.Header, msg.Header) {
					t.Fatal(msg)
				}
				seenEmergency = true
			} else {
				seenNormal = true
			}
		}
		if !seenNormal || !seenEmergency {
			t.Fatal("lane starved")
		}
		if e := <-answer; e != nil {
			t.Fatal(e)
		}
		routeReply(t, co, remote, first)
		if _, e := port.Receive(ctxTest(t)); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool {
			s := n.Resources()
			return s.InboundBytes == 0 && s.RejectBytes == 0 && s.OutboundBytes == 0
		})
	})
}
func TestDRRControlDeadlineWhileWriterBorrowed(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		n, co := newInboundNode(t, schedulerConfig(t))
		p, remote, barrier, release := drrBlockedWriter(t, n, co, tls)
		port := portTest(t, n, 2)
		id, e := p.Submit(ctxTest(t), requestBody(), RequestOptions{}, port)
		if e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, remote)
		waitBarrier(t, barrier)
		dwr := pdu.Message{Header: pdu.Header{Flags: pdu.RequestFlag, CommandCode: 280, HopByHop: 90, EndToEnd: 91}, Body: &base.DWR{OriginHost: draOrigin.Host, OriginRealm: draOrigin.Realm}}
		if e = writeWire(ctxTest(t), remote, co, dwr); e != nil {
			t.Fatal(e)
		}
		waitFor(t, func() bool { n.mu.Lock(); defer n.mu.Unlock(); return len(p.conn.controls) == 1 })
		n.mu.Lock()
		conn := p.conn
		held := n.requests[id].charge.Capacity
		conn.controls[0].deadline = time.Now().Add(-time.Second)
		n.mu.Unlock()
		n.tick(time.Now())
		n.mu.Lock()
		stopped := conn.closed && len(conn.controls) == 0 && conn.schedule.category.deficit[scheduleControl] == 0
		n.mu.Unlock()
		if !stopped || n.Resources().OutboundBytes != held {
			t.Fatal(stopped, n.Resources())
		}
		release()
		c, e := port.Receive(ctxTest(t))
		if e != nil || c.Result.Outcome != WrittenNoAnswer || !errors.Is(c.Result.Err, context.DeadlineExceeded) {
			t.Fatal(c, e)
		}
		if n.Resources().OutboundBytes != 0 {
			t.Fatal(n.Resources())
		}
	})
}
func TestDRRFixedReserve(t *testing.T) {
	cfg := schedulerConfig(t)
	n, _ := newInboundNode(t, cfg)
	r := cfg.Resources
	// Public minimum includes both physical candidates, bounded rings and the
	// frozen plan; service weights cannot buy extra admission capacity.
	minimum := 2*(n.controlBound+r.AdapterBytesPerConnection) + 4096 + uint64(cfg.Peers[0].MaxPending)*1024 + uint64(cfg.Peers[0].MaxInbound)*512 + uint64(r.MaxPorts*r.MaxPortOutstanding)*512 + uint64(r.MaxPorts)*512 + schedulerReserve(r, 1)
	cfg.Resources.TransportBytes = minimum - 1
	if bad, e := NewNode(cfg); !errors.Is(e, ErrConfig) {
		if bad != nil {
			_ = bad.Close()
		}
		t.Fatal(e)
	}
	cfg.Resources.TransportBytes = minimum
	cfg.Scheduler.RequestWeight = 64
	cfg.Scheduler.RequestClasses = []ClassWeight{{"sms", 64}}
	exact, _ := newInboundNode(t, cfg)
	if exact.Resources().OutboundBytes != 0 || exact.Resources().InboundBytes != 0 {
		t.Fatal(exact.Resources())
	}
}

func TestDRRWeightsCannotBorrowClassBytes(t *testing.T) {
	eachNetwork(t, func(t *testing.T, tls bool) {
		cfg := twoDuplicatePeers(rejectConfig(t))
		cfg.Scheduler = smallSchedule()
		cfg.Scheduler.RequestClasses = []ClassWeight{{"a", 64}}
		cfg.Scheduler.AnswerClasses = []ClassWeight{{"a", 64}}
		cfg.Resources.Classes = []ClassConfig{{"a", 1, 16}, {"b", 64 << 20, 16}}
		cfg.Resources.InboundClasses = []ClassConfig{{"a", 1, 4}, {"b", 64 << 20, 4}}
		for i, class := range []ClassID{"a", "b"} {
			cfg.Peers[i].Commands = append([]CommandPolicy(nil), cfg.Peers[i].Commands...)
			cfg.Peers[i].Commands[0].Class = class
			cfg.Peers[i].InboundCommands[0].Class = class
		}
		n, co := newInboundNode(t, cfg)
		p1, r1 := duplicateConnect(t, n, co, tls, "dra")
		p2, r2 := duplicateConnect(t, n, co, tls, "dra2")
		port := portTest(t, n, 4)
		before := n.Resources()
		if _, e := p1.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); !errors.Is(e, ErrBackpressure) {
			t.Fatal(e)
		}
		if n.Resources() != before {
			t.Fatal("failed byte admission charged resources")
		}
		if _, e := p2.Submit(ctxTest(t), requestBody(), RequestOptions{}, port); e != nil {
			t.Fatal(e)
		}
		routeReply(t, co, r2, routeRead(t, co, r2))
		if c, e := port.Receive(ctxTest(t)); e != nil || c.Result.Outcome != AnswerReceived {
			t.Fatal(c, e)
		}
		req := inboundRequest(t, n, co, 500)
		if e := writeWire(ctxTest(t), r1, co, req); e != nil {
			t.Fatal(e)
		}
		readReject(t, r1, req.Header, 3004)
		in, _ := receiveInbound(t, n, p2, r2, co, 600)
		if _, e := in.Respond(ctxTest(t), inboundFailure()); e != nil {
			t.Fatal(e)
		}
		_ = routeRead(t, co, r2)
		waitFor(t, func() bool {
			s := n.Resources()
			return s.InboundBytes == 0 && s.RejectBytes == 0 && s.OutboundBytes == 0
		})
	})
}
