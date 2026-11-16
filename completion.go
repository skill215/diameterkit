package diameter

import (
	"context"
	"github.com/skill215/diameterkit/internal/budget"
)

type CompletionLimits struct {
	MaxOutstanding      int
	MaxOutstandingBytes uint64
}
type completionEntry struct {
	value  Completion
	charge *budget.Reservation
}
type CompletionPort struct {
	node        *Node
	account     budget.Account
	queue       []completionEntry
	head, count int
	closed      bool
	changed     chan struct{}
}

func (n *Node) NewCompletionPort(l CompletionLimits) (*CompletionPort, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.closed || n.shutting {
		return nil, ErrClosed
	}
	if l.MaxOutstanding <= 0 || l.MaxOutstanding > n.cfg.Resources.MaxPortOutstanding || l.MaxOutstandingBytes == 0 {
		return nil, ErrConfig
	}
	if len(n.ports) >= n.cfg.Resources.MaxPorts {
		return nil, ErrBackpressure
	}
	p := &CompletionPort{node: n, account: budget.Account{LimitBytes: l.MaxOutstandingBytes, LimitSlots: uint64(l.MaxOutstanding)}, queue: make([]completionEntry, l.MaxOutstanding), changed: make(chan struct{})}
	n.ports[p] = true
	return p, nil
}
func (p *CompletionPort) signal() { close(p.changed); p.changed = make(chan struct{}) }
func (p *CompletionPort) Receive(ctx context.Context) (Completion, error) {
	for {
		p.node.mu.Lock()
		if p.count > 0 {
			entry := p.queue[p.head]
			p.queue[p.head] = completionEntry{}
			p.head = (p.head + 1) % len(p.queue)
			p.count--
			entry.charge.Release()
			p.signal()
			p.node.mu.Unlock()
			return entry.value, nil
		}
		if p.closed {
			p.node.mu.Unlock()
			return Completion{}, ErrClosed
		}
		ch := p.changed
		p.node.mu.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			return Completion{}, ctx.Err()
		}
	}
}
func (p *CompletionPort) Close(ctx context.Context) error {
	n := p.node
	n.mu.Lock()
	if !p.closed {
		p.closed = true
		for _, x := range n.requests {
			if x.port == p {
				n.chooseLocked(x, Result{Err: ErrClosed})
			}
		}
		for p.count > 0 {
			p.queue[p.head].charge.Release()
			p.queue[p.head] = completionEntry{}
			p.head = (p.head + 1) % len(p.queue)
			p.count--
		}
		p.signal()
		n.notify()
	}
	for p.account.UsedSlots != 0 {
		ch := p.changed
		n.mu.Unlock()
		select {
		case <-ch:
		case <-ctx.Done():
			return ctx.Err()
		}
		n.mu.Lock()
	}
	p.queue = nil
	delete(n.ports, p)
	if n.stopped && len(n.ports) == 0 {
		n.fixed = 0
	}
	n.mu.Unlock()
	return nil
}
