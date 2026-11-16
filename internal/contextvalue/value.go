// Package contextvalue holds the closed representation of protocol context.
package contextvalue

// Feature is a complete supported bitmap, not an intersection or a patch.
type Feature struct{ Vendor, List, Bits uint32 }

// Sh is bounded value evidence supplied by app/sh. It holds no callbacks, history
// or shared mutable collections. The strings are cloned by the named factory.
type Sh struct {
	Command                                                      uint32
	RequesterHost, RequesterRealm, ResponderHost, ResponderRealm string
	RequesterKnown, ResponderKnown                               bool
	Requester, Responder                                         [4]Feature
	RequesterCount, ResponderCount                               uint8
	NotificationRefs                                             uint8 // bit0 Ref11, bit1 Ref12, bit2 Ref25; explicit opaque-XML use
}

// Alert is an explicit endpoint-role binding, not notification history.
type Alert struct {
	Role                                                         uint8
	RequesterHost, RequesterRealm, ResponderHost, ResponderRealm string
}

// Retransmission is bounded evidence for a direct MME/SGSN or explicit Router leg.
type Retransmission struct {
	RequesterHost, RequesterRealm, ResponderHost, ResponderRealm string
	Target                                                       uint8
	SamePLMNKnown, SamePLMN, PowerSavingKnown, PowerSaving       bool
	RouterAlertKnown, RouterAlert                                bool
	MaximumEra, RequestedEra                                     uint32
	RequestedEraKnown                                            bool
}

// T4Charging binds one known sender/SC to an explicit IWMSC/MTC-IWF pair.
type T4Charging struct {
	RequesterHost, RequesterRealm, ResponderHost, ResponderRealm string
	IMSI, MSISDN, SCAddress, ExternalIdentifier                  string
	MSISDNPresent, ExternalKnown                                 bool
	Interface                                                    uint8 // kind7 only: SGd or Gdd; shares bounded sender storage with kind4
}

// SMSIdentity is bounded storage shared by disjoint S6c T4/IMS and SGd IMS context kinds.
// UserName is an IMSI for T4 and an HSS-ID for IMS; neither is inferred.
type SMSIdentity struct {
	URIsDigest [32]byte // kind11/12 bind two exact URI byte strings without retaining both

	Role                                                         uint8 // IMS routing only; zero for T4/report/alert
	UNRIKnown, UNRI                                              bool
	Command                                                      uint32
	RequesterHost, RequesterRealm, ResponderHost, ResponderRealm string
	UserName, SCAddress, DestinationSIPURI                       string
}

// Routing binds a declared GMSC/IP-SM-GW requester to its HSS and an external UNRI fact.
type Routing struct {
	RequesterHost, RequesterRealm, ResponderHost, ResponderRealm string
	Role                                                         uint8
	UNRIKnown, UNRI                                              bool
}
type Value struct {
	routing        Routing
	identity       SMSIdentity
	t4             T4Charging
	retransmission Retransmission
	sh             Sh
	alert          Alert
	kind           uint8
}

func NewSh(s Sh) Value                                 { return Value{sh: s, kind: 1} }
func NewAlert(a Alert) Value                           { return Value{alert: a, kind: 2} }
func NewRetransmission(r Retransmission) Value         { return Value{retransmission: r, kind: 3} }
func (v Value) Retransmission() (Retransmission, bool) { return v.retransmission, v.kind == 3 }
func NewT4Charging(t T4Charging) Value                 { return Value{t4: t, kind: 4} }
func (v Value) T4Charging() (T4Charging, bool)         { return v.t4, v.kind == 4 }
func NewS6cT4(s SMSIdentity) Value                     { return Value{identity: s, kind: 5} }
func (v Value) S6cT4() (SMSIdentity, bool)             { return v.identity, v.kind == 5 }
func NewRouting(r Routing) Value                       { return Value{routing: r, kind: 6} }
func (v Value) Routing() (Routing, bool)               { return v.routing, v.kind == 6 }
func NewMOForwarding(t T4Charging) Value               { return Value{t4: t, kind: 7} }
func (v Value) MOForwarding() (T4Charging, bool)       { return v.t4, v.kind == 7 }
func (v Value) Empty() bool                            { return v.kind == 0 }
func (v Value) Sh() (Sh, bool)                         { return v.sh, v.kind == 1 }
func (v Value) Alert() (Alert, bool)                   { return v.alert, v.kind == 2 }

// Endpoints permits the protocol-neutral runtime to enforce its actual binding.
func (v Value) Endpoints() (string, string, string, string, bool) {
	if v.kind == 6 {
		r := v.routing
		return r.RequesterHost, r.RequesterRealm, r.ResponderHost, r.ResponderRealm, true
	}
	if v.kind == 5 || v.kind == 8 || v.kind == 9 || v.kind == 10 || v.kind == 11 || v.kind == 12 || v.kind == 13 || v.kind == 14 || v.kind == 15 {
		s := v.identity
		return s.RequesterHost, s.RequesterRealm, s.ResponderHost, s.ResponderRealm, true
	}
	if v.kind == 4 || v.kind == 7 {
		t := v.t4
		return t.RequesterHost, t.RequesterRealm, t.ResponderHost, t.ResponderRealm, true
	}
	if v.kind == 3 {
		r := v.retransmission
		return r.RequesterHost, r.RequesterRealm, r.ResponderHost, r.ResponderRealm, true
	}
	if v.kind == 2 {
		a := v.alert
		return a.RequesterHost, a.RequesterRealm, a.ResponderHost, a.ResponderRealm, true
	}
	s := v.sh
	return s.RequesterHost, s.RequesterRealm, s.ResponderHost, s.ResponderRealm, v.kind == 1
}

func NewS6cIMSReport(s SMSIdentity) Value         { return Value{identity: s, kind: 8} }
func (v Value) S6cIMSReport() (SMSIdentity, bool) { return v.identity, v.kind == 8 }

func NewS6cIMSAlert(s SMSIdentity) Value         { return Value{identity: s, kind: 9} }
func (v Value) S6cIMSAlert() (SMSIdentity, bool) { return v.identity, v.kind == 9 }

func NewS6cIMSRouting(s SMSIdentity) Value         { return Value{identity: s, kind: 10} }
func (v Value) S6cIMSRouting() (SMSIdentity, bool) { return v.identity, v.kind == 10 }

// IMS MO reuses the inactive T4 source slots without sharing T4 purpose or endpoints.
func NewIMSMO(s SMSIdentity, imsi, msisdn string, msisdnPresent bool) Value {
	return Value{identity: s, t4: T4Charging{IMSI: imsi, MSISDN: msisdn, MSISDNPresent: msisdnPresent}, kind: 11}
}
func (v Value) IMSMOSender() (imsi, msisdn string, msisdnPresent, ok bool) {
	return v.t4.IMSI, v.t4.MSISDN, v.t4.MSISDNPresent, v.kind == 11
}
func (v Value) IMSMO() (SMSIdentity, bool) { return v.identity, v.kind == 11 }

func NewIMSMT(s SMSIdentity) Value         { return Value{identity: s, kind: 12} }
func (v Value) IMSMT() (SMSIdentity, bool) { return v.identity, v.kind == 12 }

// Router Alert reuses inactive identity slots; UserName is an exact opaque
// token, not an IMSI/HSS-ID. No prior-exchange objects or lookup callbacks remain.
func NewRouterAlert(s SMSIdentity) Value         { return Value{identity: s, kind: 13} }
func (v Value) RouterAlert() (SMSIdentity, bool) { return v.identity, v.kind == 13 }

// Special SRR purposes reuse identity storage. Kind14 asserts HPLMN Router
// routing; kind15 uses UserName/UNRIKnown for an optional exact MCCMNC prefix.
func NewRouterRouting(s SMSIdentity) Value         { return Value{identity: s, kind: 14} }
func (v Value) RouterRouting() (SMSIdentity, bool) { return v.identity, v.kind == 14 }
func NewMCCMNCRouting(s SMSIdentity) Value         { return Value{identity: s, kind: 15} }
func (v Value) MCCMNCRouting() (SMSIdentity, bool) { return v.identity, v.kind == 15 }
