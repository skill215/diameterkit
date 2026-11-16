// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

package diameter

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/budget"
	"github.com/skill215/diameterkit/internal/codecengine"
	"github.com/skill215/diameterkit/pdu"
	"github.com/skill215/diameterkit/wire"
	"time"
)

// RoTargetConfig is an explicit deployment declaration, not verified remote proof.
// Operational limits/class come from the listed peers' existing CCR policies;
// only Destination and answer sources are replaced for recovery operations.
type RoTargetConfig struct {
	TargetID                 string
	Peers                    []PeerID
	Destination              pdu.Origin
	ErrorAgents              []pdu.Origin
	ReplayDomain, EvidenceID string
	Retention                time.Duration
	// Named legacy session continuity; standard CC-Session-Failover is not inferred.
	SessionEvidenceID string
}
type roTargetConfig struct {
	binding RoRecoveryBinding
	peers   []PeerID
}

// RoRecoveryPath lists one terminal and its ordered alternate transport paths.
// Targets cannot repeat within a segment plan.
type RoRecoveryPath struct {
	TargetID string
	Peers    []PeerID
}
type roPreparedTarget struct {
	index    int
	id       string
	peers    []*Peer
	policy   CommandPolicy
	frame    []byte
	snapshot *codecengine.RequestSnapshot
}

func validRoReference(s string) bool { return len(s) > 0 && len(s) <= 128 }
func roTargetConfigReserve(cfg NodeConfig) (uint64, error) {
	if len(cfg.RoTargets) > 16 || len(cfg.RoTargets) > 0 && cfg.RoRecovery.MaxTargets == 0 {
		return 0, ErrConfig
	}
	var cost uint64
	for _, t := range cfg.RoTargets {
		if !validRoReference(t.TargetID) || !validRoReference(t.ReplayDomain) || !validRoReference(t.EvidenceID) || len(t.SessionEvidenceID) > 128 || len(t.Peers) < 1 || len(t.Peers) > 64 || len(t.ErrorAgents) > 63 || avp.Identity(t.Destination.Host) != nil || avp.Identity(t.Destination.Realm) != nil || t.Retention <= 0 || t.Retention > 365*24*time.Hour {
			return 0, ErrConfig
		}
		cost += 4096 + uint64(len(t.Peers))*256 + uint64(len(t.ErrorAgents))*1024 + uint64(len(t.Destination.Host)+len(t.Destination.Realm))*4
		for _, o := range t.ErrorAgents {
			if avp.Identity(o.Host) != nil || avp.Identity(o.Realm) != nil {
				return 0, ErrConfig
			}
		}
	}
	return cost, nil
}
func (n *Node) configureRoTargets(configs []RoTargetConfig) error {
	n.roTargets = make(map[string]roTargetConfig, len(configs))
	key := pdu.CommandKey{ApplicationID: 4, Code: 272, Request: true}
	for _, t := range configs {
		if _, exists := n.roTargets[t.TargetID]; exists {
			return ErrConfig
		}
		var policy CommandPolicy
		peers := make([]PeerID, 0, len(t.Peers))
		for _, id := range t.Peers {
			p := n.peers[id]
			if p == nil {
				return ErrConfig
			}
			v, ok := p.policies[key]
			if !ok || v.ApplicationBinding != HeaderApplication {
				return ErrConfig
			}
			for _, old := range peers {
				if old == id {
					return ErrConfig
				}
			}
			v.DestinationHost = avp.CloneString(t.Destination.Host)
			v.DestinationRealm = avp.CloneString(t.Destination.Realm)
			v.Sources = cloneRoSources(pdu.AnswerSources{Terminals: []pdu.Origin{t.Destination}, ErrorAgents: t.ErrorAgents})
			if len(peers) > 0 && !sameRoPolicy(v, policy) {
				return ErrConfig
			}
			policy = v
			peers = append(peers, PeerID(avp.CloneString(string(id))))
		}
		b := RoRecoveryBinding{AllowedPeers: peers, TargetID: avp.CloneString(t.TargetID), Retention: t.Retention, Policy: policy, ReplayDomain: avp.CloneString(t.ReplayDomain), EvidenceID: avp.CloneString(t.EvidenceID), SessionEvidenceID: avp.CloneString(t.SessionEvidenceID)}
		n.roTargets[b.TargetID] = roTargetConfig{b, peers}
	}
	return nil
}
func (n *Node) roBinding(t RoRecoveryTarget, key pdu.CommandKey) (RoRecoveryBinding, bool) {
	if c, ok := n.roTargets[t.ID]; ok {
		for _, id := range c.peers {
			if id == t.Peer && t.Retention <= c.binding.Retention {
				b := c.binding
				b.Retention = t.Retention
				return b, true
			}
		}
		return RoRecoveryBinding{}, false
	}
	p := n.peers[t.Peer]
	if p == nil {
		return RoRecoveryBinding{}, false
	}
	v, ok := p.policies[key]
	return RoRecoveryBinding{TargetID: t.ID, Retention: t.Retention, Policy: v}, ok
}
func (n *Node) roPathAllowed(b RoRecoveryBinding, id PeerID) bool {
	if c, ok := n.roTargets[b.TargetID]; ok {
		if !sameRoPeers(b.AllowedPeers, c.peers) || b.Retention > c.binding.Retention || b.ReplayDomain != c.binding.ReplayDomain || b.EvidenceID != c.binding.EvidenceID || b.SessionEvidenceID != c.binding.SessionEvidenceID || !sameRoPolicy(b.Policy, c.binding.Policy) {
			return false
		}
		for _, p := range c.peers {
			if p == id {
				return true
			}
		}
		return false
	}
	p := n.peers[id]
	return p != nil && len(b.AllowedPeers) == 0 && b.ReplayDomain == "" && b.EvidenceID == "" && b.SessionEvidenceID == "" && sameRoPolicy(p.policies[b.Policy.Command], b.Policy)
}
func roTargetIndex(r RoRecoveryRequest, id string) int {
	for i, t := range r.Targets {
		if t.TargetID == id {
			return i
		}
	}
	return -1
}
func roCrossAllowed(r RoRecoveryRequest, index int) bool {
	if index == 0 {
		return true
	}
	a, b := r.Targets[0], r.Targets[index]
	if a.ReplayDomain == "" || a.ReplayDomain != b.ReplayDomain || a.EvidenceID == "" || b.EvidenceID == "" {
		return false
	}
	// Original frozen CCR was already validated. TERM needs a per-session
	// application declaration and target continuity evidence under named legacy policy.
	d := wire.NewDecoder(r.Frame[20:], wire.DefaultLimits())
	for {
		v, ok, e := d.Next()
		if e != nil {
			return false
		}
		if !ok {
			break
		}
		if v.Key == (avp.Key{Code: 416}) {
			typ, e := wire.U32(v.Data)
			if e != nil {
				return false
			}
			if typ == 3 {
				return r.PolicyID == RoSMSFWRecoveryPolicy && r.SessionEvidenceID != "" && a.SessionEvidenceID != "" && b.SessionEvidenceID != ""
			}
		}
	}
	return true
}
func (n *Node) addRoTargetAttemptLocked(g *routeExchange) (AttemptID, error) {
	state := g.recovery
	cur := g.attempts[len(g.attempts)-1]
	last := error(ErrNoEligibleAlternate)
	for _, t := range state.targets {
		if state.visited&(uint16(1)<<t.index) != 0 && t.index != cur.roTarget {
			continue
		}
		for _, p := range t.peers {
			if !p.readyLocked() || !p.conn.apps.allows(g.route.cfg.Command, t.policy.ApplicationBinding) {
				continue
			}
			c := p.conn
			used := false
			for _, a := range g.attempts {
				if a.attemptConnection == c.id && a.roTarget == t.index {
					used = true
				}
				if a.writing {
					return AttemptID{}, ErrAttemptNotReplayable
				}
			}
			if used {
				continue
			}
			pc, e := budget.Reserve(g.peerBytes, 1, "route-peer", &p.account)
			if e != nil {
				last = ErrBackpressure
				continue
			}
			hbh, e := c.nextHBH(false)
			if e != nil {
				pc.Release()
				last = e
				continue
			}
			retransmit := state.replayPending
			for _, a := range g.attempts {
				retransmit = retransmit || a.delivery.MayHaveReachedPeer
			}
			snap := t.snapshot.ForAttempt(hbh, retransmit, t.policy.Sources)
			wire.PutHeader(t.frame, snap.Header(), len(t.frame))
			id := AttemptID{g.root.id, uint32(len(g.attempts) + 1)}
			a := &exchange{roTarget: t.index, scheduleClass: g.root.scheduleClass, id: g.root.id, conn: c, port: g.root.port, charge: g.root.charge, snapshot: snap, frame: t.frame, deadline: g.root.deadline, answerMax: t.policy.MaxAnswerBytes, answerBound: g.root.answerBound, hbh: hbh, delivery: DeliveryEvidence{FrameBytes: len(t.frame)}, group: g, attemptID: id, peerCharge: pc, attemptPeer: p, attemptConnection: c.id}
			startRoResponseWait(a, time.Now())
			state.visited |= uint16(1) << t.index
			g.attempts = append(g.attempts, a)
			g.phase = AttemptActive
			c.pending[hbh] = a
			c.enqueueRequest(a)
			c.signal()
			n.notify()
			return id, nil
		}
	}
	return AttemptID{}, last
}

func roAttemptTarget(g *routeExchange, a *exchange) string {
	if g.recovery != nil {
		for _, t := range g.recovery.targets {
			if t.index == a.roTarget {
				return t.id
			}
		}
	}
	return ""
}

func sameRoPeers(a, b []PeerID) bool {
	if len(a) != len(b) {
		return false
	}
	for i, id := range a {
		if id != b[i] {
			return false
		}
	}
	return true
}
