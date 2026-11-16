// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import "github.com/skill215/diameterkit/avp"

// ClassWeight changes service share, never the class admission budget.
type ClassWeight struct {
	ID     ClassID
	Weight uint32
}

// SchedulerConfig is frozen at Node construction. All-zero scalar fields select
// 16 KiB, Control:Answer:Request = 4:2:1, Emergency = 1. Otherwise every scalar
// must be valid: QuantumBytes is 64..1048576 and weights are 1..64.
// Omitted class overrides expand to weight 1; unknown or duplicate IDs fail.
// Request and answer classes refer to their respective resource partitions.
// EmergencyWeight controls an independent answer lane for emergency rejects.
type SchedulerConfig struct {
	QuantumBytes                                                int
	ControlWeight, AnswerWeight, RequestWeight, EmergencyWeight uint32
	RequestClasses, AnswerClasses                               []ClassWeight
}
type schedulerPlan struct {
	config                        SchedulerConfig
	requestIndex, answerIndex     map[ClassID]int
	requestQuantum, answerQuantum []uint64
	categoryQuantum               []uint64
	maxFrame                      uint64
}

func compileScheduler(cfg SchedulerConfig, r ResourcePlan, maxFrame int) (schedulerPlan, error) {
	p := schedulerPlan{requestIndex: make(map[ClassID]int), answerIndex: make(map[ClassID]int), maxFrame: uint64(maxFrame)}
	if len(r.Classes) > 256 || len(r.InboundClasses) > 256 {
		return p, ErrConfig
	}
	if cfg.QuantumBytes == 0 && cfg.ControlWeight == 0 && cfg.AnswerWeight == 0 && cfg.RequestWeight == 0 && cfg.EmergencyWeight == 0 {
		cfg.QuantumBytes = 16384
		cfg.ControlWeight = 4
		cfg.AnswerWeight = 2
		cfg.RequestWeight = 1
		cfg.EmergencyWeight = 1
	}
	if cfg.QuantumBytes < 64 || cfg.QuantumBytes > 1<<20 {
		return p, ErrConfig
	}
	for _, w := range []uint32{cfg.ControlWeight, cfg.AnswerWeight, cfg.RequestWeight, cfg.EmergencyWeight} {
		if w < 1 || w > 64 {
			return p, ErrConfig
		}
	}
	expand := func(classes []ClassConfig, overrides []ClassWeight, index map[ClassID]int) ([]ClassWeight, []uint64, error) {
		if len(overrides) > len(classes) {
			return nil, nil, ErrConfig
		}
		weights := make([]ClassWeight, len(classes))
		quantum := make([]uint64, len(classes))
		for i, c := range classes {
			if c.ID == "" || len(c.ID) > 128 {
				return nil, nil, ErrConfig
			}
			if _, ok := index[c.ID]; ok {
				return nil, nil, ErrConfig
			}
			id := ClassID(avp.CloneString(string(c.ID)))
			index[id] = i
			weights[i] = ClassWeight{id, 1}
		}
		seen := make(map[ClassID]bool)
		for _, w := range overrides {
			i, ok := index[w.ID]
			if !ok || seen[w.ID] || w.Weight < 1 || w.Weight > 64 {
				return nil, nil, ErrConfig
			}
			seen[w.ID] = true
			weights[i].Weight = w.Weight
		}
		for i, w := range weights {
			quantum[i] = uint64(cfg.QuantumBytes) * uint64(w.Weight)
		}
		return weights, quantum, nil
	}
	var err error
	cfg.RequestClasses, p.requestQuantum, err = expand(r.Classes, cfg.RequestClasses, p.requestIndex)
	if err != nil {
		return p, err
	}
	cfg.AnswerClasses, p.answerQuantum, err = expand(r.InboundClasses, cfg.AnswerClasses, p.answerIndex)
	if err != nil {
		return p, err
	}
	// Emergency rejects have their own answer lane; no user ClassID can collide.
	p.answerQuantum = append(p.answerQuantum, uint64(cfg.QuantumBytes)*uint64(cfg.EmergencyWeight))
	p.categoryQuantum = []uint64{uint64(cfg.QuantumBytes) * uint64(cfg.ControlWeight), uint64(cfg.QuantumBytes) * uint64(cfg.AnswerWeight), uint64(cfg.QuantumBytes) * uint64(cfg.RequestWeight)}
	p.config = cfg
	return p, nil
}

// SchedulerInfo returns expanded owned configuration, not queue contents or payloads.
func (n *Node) SchedulerInfo() SchedulerConfig {
	v := n.scheduler.config
	v.RequestClasses = append([]ClassWeight(nil), v.RequestClasses...)
	v.AnswerClasses = append([]ClassWeight(nil), v.AnswerClasses...)
	return v
}
func schedulerReserve(r ResourcePlan, peers int) uint64 {
	classes := uint64(len(r.Classes) + len(r.InboundClasses))
	return 2048 + 256*classes + 2*uint64(peers)*(1024+64*(classes+1))
}

const (
	scheduleControl = iota
	scheduleAnswer
	scheduleRequest
)

// drrRing retains credit only for a nonempty lane. A visit may send multiple
// whole frames; peek never debits. Empty rounds are skipped arithmetically,
// so even a large frame with a small quantum takes at most two lane scans.
type drrRing struct {
	quantum, deficit []uint64
	cursor           int
	credited         bool
	maxFrame         uint64
}

func newDRR(q []uint64, maxFrame uint64) drrRing {
	return drrRing{quantum: q, deficit: make([]uint64, len(q)), maxFrame: maxFrame}
}
func (r *drrRing) reset(i int) {
	r.deficit[i] = 0
	if r.cursor == i {
		r.credited = false
	}
}

// Emptying the current lane ends its visit immediately. A producer that
// refills a one-frame lane during I/O must not get a fresh visit ahead of peers.
func (r *drrRing) empty(i int) {
	r.deficit[i] = 0
	if r.cursor == i {
		r.next()
	}
}
func (r *drrRing) clear() { clear(r.deficit); r.cursor = 0; r.credited = false }
func (r *drrRing) next()  { r.cursor = (r.cursor + 1) % len(r.deficit); r.credited = false }
func (r *drrRing) peek(cost []uint64) int {
	if len(cost) == 0 {
		return -1
	}
	for pass := 0; pass < 2; pass++ {
		for range cost {
			i := r.cursor
			if cost[i] == 0 {
				r.reset(i)
				r.next()
				continue
			}
			if !r.credited {
				r.deficit[i] = min(r.deficit[i]+r.quantum[i], r.maxFrame+r.quantum[i])
				r.credited = true
			}
			if r.deficit[i] >= cost[i] {
				return i
			}
			r.next()
		}
		rounds := uint64(0)
		for i, c := range cost {
			if c == 0 {
				continue
			}
			needed := (c - r.deficit[i] + r.quantum[i] - 1) / r.quantum[i]
			if rounds == 0 || needed < rounds {
				rounds = needed
			}
		}
		if rounds == 0 {
			return -1
		}
		for i, c := range cost {
			if c != 0 {
				r.deficit[i] = min(r.deficit[i]+(rounds-1)*r.quantum[i], r.maxFrame+r.quantum[i])
			}
		}
	}
	panic("DRR head exceeds its configured frame bound")
}
func (r *drrRing) debit(i int, cost uint64) { r.deficit[i] -= cost }

type writerSchedule struct {
	category, request, answer   drrRing
	requestCounts, answerCounts []int
	requestCosts, answerCosts   []uint64
}

func (p *schedulerPlan) newWriter() writerSchedule {
	return writerSchedule{category: newDRR(p.categoryQuantum, p.maxFrame), request: newDRR(p.requestQuantum, p.maxFrame), answer: newDRR(p.answerQuantum, p.maxFrame), requestCounts: make([]int, len(p.requestQuantum)), answerCounts: make([]int, len(p.answerQuantum)), requestCosts: make([]uint64, len(p.requestQuantum)), answerCosts: make([]uint64, len(p.answerQuantum))}
}
func (c *connection) enqueueRequest(x *exchange) {
	s := &c.schedule
	i := x.scheduleClass
	if len(c.queue) == 0 {
		s.category.reset(scheduleRequest)
	}
	if s.requestCounts[i] == 0 {
		s.request.reset(i)
	}
	s.requestCounts[i]++
	c.queue = append(c.queue, x)
}
func (c *connection) removeRequest(x *exchange) {
	for j, v := range c.queue {
		if v == x {
			copy(c.queue[j:], c.queue[j+1:])
			c.queue[len(c.queue)-1] = nil
			c.queue = c.queue[:len(c.queue)-1]
			s := &c.schedule
			i := x.scheduleClass
			s.requestCounts[i]--
			if s.requestCounts[i] == 0 {
				s.request.empty(i)
			}
			if len(c.queue) == 0 {
				s.category.empty(scheduleRequest)
			}
			return
		}
	}
}
func (c *connection) enqueueAnswer(in *Inbound) {
	s := &c.schedule
	i := len(s.answerCounts) - 1
	if !in.emergency {
		i = c.peer.node.scheduler.answerIndex[in.policy.Class]
	}
	in.scheduleClass = i
	if len(c.answers) == 0 {
		s.category.reset(scheduleAnswer)
	}
	if s.answerCounts[i] == 0 {
		s.answer.reset(i)
	}
	s.answerCounts[i]++
	c.answers = append(c.answers, in)
}
func (c *connection) removeAnswer(in *Inbound) {
	for j, v := range c.answers {
		if v == in {
			copy(c.answers[j:], c.answers[j+1:])
			c.answers[len(c.answers)-1] = nil
			c.answers = c.answers[:len(c.answers)-1]
			s := &c.schedule
			i := in.scheduleClass
			s.answerCounts[i]--
			if s.answerCounts[i] == 0 {
				s.answer.empty(i)
			}
			if len(c.answers) == 0 {
				s.category.empty(scheduleAnswer)
			}
			return
		}
	}
}
func (c *connection) clearAnswers() {
	c.answers = nil
	clear(c.schedule.answerCounts)
	c.schedule.answer.clear()
	c.schedule.category.empty(scheduleAnswer)
}
func (c *connection) enqueueControl(job controlJob) {
	if len(c.controls) == 0 {
		c.schedule.category.reset(scheduleControl)
	}
	c.controls = append(c.controls, job)
}
func (c *connection) popControl() controlJob {
	job := c.controls[0]
	copy(c.controls, c.controls[1:])
	c.controls[len(c.controls)-1] = controlJob{}
	c.controls = c.controls[:len(c.controls)-1]
	if len(c.controls) == 0 {
		c.schedule.category.empty(scheduleControl)
	}
	return job
}
func answerCost(in *Inbound) uint64 {
	if in.replay != nil {
		return uint64(20 + len(in.replay.result) + len(in.proxy))
	}
	return uint64(len(in.frame))
}
func (c *connection) pickWrite() (kind, class int, cost uint64) {
	s := &c.schedule
	clear(s.requestCosts)
	clear(s.answerCosts)
	for _, x := range c.queue {
		if s.requestCosts[x.scheduleClass] == 0 {
			s.requestCosts[x.scheduleClass] = uint64(len(x.frame))
		}
	}
	for _, in := range c.answers {
		if s.answerCosts[in.scheduleClass] == 0 {
			s.answerCosts[in.scheduleClass] = answerCost(in)
		}
	}
	ri, ai := s.request.peek(s.requestCosts), s.answer.peek(s.answerCosts)
	var costs [3]uint64
	if len(c.controls) > 0 {
		costs[scheduleControl] = uint64(len(c.controls[0].prepared.frame))
	}
	if ri >= 0 {
		costs[scheduleRequest] = s.requestCosts[ri]
	}
	if ai >= 0 {
		costs[scheduleAnswer] = s.answerCosts[ai]
	}
	kind = s.category.peek(costs[:])
	if kind < 0 {
		return -1, 0, 0
	}
	if kind == scheduleRequest {
		class = ri
	} else if kind == scheduleAnswer {
		class = ai
	}
	return kind, class, costs[kind]
}
func (c *connection) debitWrite(kind, class int, cost uint64) {
	s := &c.schedule
	s.category.debit(kind, cost)
	if kind == scheduleRequest {
		s.request.debit(class, cost)
	} else if kind == scheduleAnswer {
		s.answer.debit(class, cost)
	}
}
