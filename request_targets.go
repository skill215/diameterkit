package diameter

import (
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/pdu"
)

// RequestTarget authorizes one terminal for one existing command on a Peer.
// Class, limits, timeout and error agents come from that command's default policy.
// A selected target accepts normal answers only from Destination. This is not
// discovery, a replay permission or a Ro recovery target declaration.
type RequestTarget struct {
	ID          string
	Command     pdu.CommandKey
	Destination pdu.Origin
}
type requestTargetKey struct {
	command pdu.CommandKey
	id      string
}

// Charge the bounded compiled map, policies, terminal slices and owned text
// before allocating them. Existing immutable ErrorAgents are shared, not cloned.
func requestTargetReserve(cfg NodeConfig) (uint64, error) {
	var cost uint64
	for _, pc := range cfg.Peers {
		if len(pc.RequestTargets) > 64 {
			return 0, ErrConfig
		}
		if len(pc.RequestTargets) > 0 {
			cost += 1024
		}
		for _, t := range pc.RequestTargets {
			if len(t.ID) == 0 || len(t.ID) > 128 || !t.Command.Request || t.Command.ApplicationID == 0 || avp.Identity(t.Destination.Host) != nil || avp.Identity(t.Destination.Realm) != nil {
				return 0, ErrConfig
			}
			cost += 1024 + uint64(len(t.ID)+len(t.Destination.Host)+len(t.Destination.Realm))
		}
	}
	return cost, nil
}
func (p *Peer) configureRequestTargets(ts []RequestTarget) error {
	if len(ts) == 0 {
		return nil
	}
	p.targets = make(map[requestTargetKey]CommandPolicy, len(ts))
	for _, t := range ts {
		policy, ok := p.policies[t.Command]
		if !ok || policy.ApplicationBinding != HeaderApplication {
			return ErrConfig
		}
		key := requestTargetKey{t.Command, t.ID}
		if _, exists := p.targets[key]; exists {
			return ErrConfig
		}
		key.id = avp.CloneString(t.ID)
		policy.DestinationHost = avp.CloneString(t.Destination.Host)
		policy.DestinationRealm = avp.CloneString(t.Destination.Realm)
		policy.Sources.Terminals = []pdu.Origin{{Host: policy.DestinationHost, Realm: policy.DestinationRealm}}
		p.targets[key] = policy
	}
	return nil
}
func (p *Peer) requestPolicy(command pdu.CommandKey, target string) (CommandPolicy, bool) {
	if len(target) > 128 {
		return CommandPolicy{}, false
	}
	if target == "" {
		v, ok := p.policies[command]
		return v, ok
	}
	v, ok := p.targets[requestTargetKey{command, target}]
	return v, ok
}
