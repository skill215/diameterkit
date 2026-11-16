// SPDX-AI-Disclosure: ai-generated
// SPDX-AI-Model: Unrecorded; see AI_DISCLOSURE.md
// SPDX-AI-Provider: OpenAI
// SPDX-AI-Scope: Library code generated and revised with Codex under maintainer direction.
// SPDX-AI-Date: 2026-11-16

// Package shdata explicitly interprets a bounded Ref11/12/24/25 subset of TS 29.328
// V18.1.0 Sh-Data. It is independent of Diameter decoding and business state.
// Parse checks XML and selected structural rules, not the entire Sh XSD.
package shdata

import (
	"errors"
	"fmt"
	"slices"

	"github.com/skill215/diameterkit/avp"
)

var (
	ErrXML         = errors.New("shdata: malformed XML")
	ErrStructure   = errors.New("shdata: invalid selected structure or value")
	ErrUnsupported = errors.New("shdata: unsupported representation")
	ErrLimit       = errors.New("shdata: resource limit exceeded")
	ErrArgument    = errors.New("shdata: invalid argument or limits")
)

// Error identifies a bounded path and byte offset, without retaining input or
// echoing payload values. Use errors.Is to distinguish the error kind.
type Error struct {
	Kind   error
	Path   string
	Offset int64
}

func (e *Error) Error() string { return fmt.Sprintf("%v at %s (byte %d)", e.Kind, e.Path, e.Offset) }
func (e *Error) Unwrap() error { return e.Kind }

// Limits are explicit per-call limits. Zero is invalid. MaxBytes bounds the
// input before tokenization (including a single token's temporary allocations).
// The other limits are checked as tokens arrive, before retaining their data.
// These logical limits are not a Go heap/RSS estimate or Node budget reservation.
type Limits struct{ MaxBytes, MaxDepth, MaxElements, MaxAttributes, MaxTextBytes, MaxOpaqueBytes int }

func DefaultLimits() Limits { return Limits{64 << 10, 24, 512, 256, 32 << 10, 32 << 10} }
func (l Limits) valid() bool {
	return l.MaxBytes > 0 && l.MaxBytes <= 1<<20 && l.MaxDepth > 0 && l.MaxDepth <= 64 && l.MaxElements > 0 && l.MaxElements <= 4096 && l.MaxAttributes > 0 && l.MaxAttributes <= 4096 && l.MaxTextBytes > 0 && l.MaxTextBytes <= l.MaxBytes && l.MaxOpaqueBytes > 0 && l.MaxOpaqueBytes <= l.MaxBytes
}

type Usage struct{ Bytes, MaxDepth, Elements, Attributes, TextBytes, OpaqueBytes int }

// Number is required inside a present SMSRegistration. An empty Number is an
// explicit empty element; it means deregistration only in SMSUpdate context.
// SCAddress and SBI preserve absence, empty text and explicit zero separately.
type SMSRegistration struct {
	Number    string
	SCAddress avp.Optional[string]
	SBI       avp.Optional[uint8]
}
type Reachability struct{ MME, SGSN, AMF3GPP, AMFNon3GPP avp.Optional[uint8] }

func (r Reachability) any() bool {
	return r.MME.Present || r.SGSN.Present || r.AMF3GPP.Present || r.AMFNon3GPP.Present
}

type Data struct {
	SMS          avp.Optional[SMSRegistration]
	Reachability avp.Optional[Reachability]
	IMSUserState avp.Optional[uint8]
	SCSCFName    avp.Optional[string]
}

// Fragment indexes an uninterpreted subtree in Raw(), including its start/end
// tags. It is not a standalone XML document: namespace declarations may be on an
// ancestor. Its contents have XML checks only, not Sh schema/semantic checks.
type Fragment struct {
	Path, Namespace, Local string
	Start, End             int
}
type Reason string

const (
	Uninterpreted        Reason = "uninterpreted"
	NumberRepresentation Reason = "number-representation"
	URIRepresentation    Reason = "uri-representation"
	MissingData          Reason = "missing-data"
	NotApplicable        Reason = "not-applicable"
)

type Issue struct {
	Path   string
	Reason Reason
}
type Use uint8

const (
	SMSUpdate                Use = iota + 1 // Ref24 PUR: SCAddress is not update data.
	SMSPull                                 // Ref24 successful UDA data.
	ReachabilityNotification                // Ref25 data in PNR; no subscription lookup.
	IMSUserStateData                        // Ref11 data in a successful SNA or independent PNR.
	SCSCFNameData                           // Ref12; present empty means unavailable/removed, not absent.
)
const Scope = "shdata-29.328-ref11-12-24-25-v1"

// Complete covers only this fixed subset and use. It never establishes user
// permissions, a remote state transition, full XSD conformance or reachability
// after the instant represented by the notification.
type Report struct {
	Scope    string
	Complete bool
	Issues   []Issue
}

// Document is immutable after Parse and safe for concurrent reads. Accessors
// return owned slices. It stores no subscription, clock, network or persistence.
type Document struct {
	raw       []byte
	data      Data
	fragments []Fragment
	issues    []Issue
	usage     Usage
}

func (d *Document) Raw() []byte           { return slices.Clone(d.raw) }
func (d *Document) Data() Data            { return d.data }
func (d *Document) Fragments() []Fragment { return slices.Clone(d.fragments) }
func (d *Document) Usage() Usage          { return d.usage }

const smsPath = "/Sh-Data/Sh-IMS-Data/Extension/Extension/Extension/SMSRegistrationInfo"
const reachPath = "/Sh-Data/Sh-IMS-Data/Extension/Extension/Extension/UEReachabilityForIP"

func (d *Document) Check(use Use) (Report, error) {
	if d == nil || use < SMSUpdate || use > SCSCFNameData {
		return Report{}, ErrArgument
	}
	return d.checkUses(use, 0)
}
func (d *Document) checkUses(use Use, refs uint8) (Report, error) {
	r := Report{Scope: Scope, Issues: slices.Clone(d.issues)}
	add := func(path string, reason Reason) { r.Issues = append(r.Issues, Issue{path, reason}) }

	if use == IMSUserStateData || refs&1 != 0 {
		if !d.data.IMSUserState.Present {
			add("/Sh-Data/Sh-IMS-Data/IMSUserState", MissingData)
		}
	} else if d.data.IMSUserState.Present {
		add("/Sh-Data/Sh-IMS-Data/IMSUserState", NotApplicable)
	}
	if use == SCSCFNameData || refs&2 != 0 {
		if !d.data.SCSCFName.Present {
			add("/Sh-Data/Sh-IMS-Data/SCSCFName", MissingData)
		}
	} else if d.data.SCSCFName.Present {
		add("/Sh-Data/Sh-IMS-Data/SCSCFName", NotApplicable)
	}
	if use == ReachabilityNotification || refs&4 != 0 {
		if !d.data.Reachability.Present || !d.data.Reachability.Value.any() {
			add(reachPath, MissingData)
		}
	} else if d.data.Reachability.Present {
		add(reachPath, NotApplicable)
	}
	if use == SMSUpdate || use == SMSPull {
		if !d.data.SMS.Present {
			add(smsPath, MissingData)
		}
		if use == SMSUpdate && d.data.SMS.Value.SCAddress.Present {
			add(smsPath+"/Extension/SCAddress", NotApplicable)
		}
	} else if d.data.SMS.Present {
		add(smsPath, NotApplicable)
	}

	r.Complete = len(r.Issues) == 0
	return r, nil
}
func number(s string) bool {
	if len(s) < 1 || len(s) > 15 || s[0] < '1' || s[0] > '9' {
		return false
	}
	for i := range s {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
