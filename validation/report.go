// Package validation defines bounded, protocol-neutral validation results.
package validation

import (
	"errors"
	"fmt"
	"github.com/skill215/diameterkit/avp"
	"github.com/skill215/diameterkit/internal/contextvalue"
)

type Context = contextvalue.Value
type Mode uint8

const (
	Build Mode = iota
	Send
	Receive
)

var (
	ErrResultConflict  = errors.New("conflicting application results")
	ErrField           = errors.New("invalid field")
	ErrMissing         = errors.New("missing required AVP")
	ErrDuplicate       = errors.New("AVP cardinality exceeded")
	ErrUnsupported     = errors.New("unsupported protocol capability")
	ErrType            = errors.New("body type does not match command")
	ErrEnvelope        = errors.New("conflicting envelope")
	ErrUnmatchedAnswer = errors.New("answer does not match request header")
	ErrInvalidAnswer   = errors.New("matched answer is invalid")
	ErrSource          = errors.New("answer origin is not authorized")
)

// FieldError never includes subscriber data. Path is ordered outermost first.
type FieldError struct {
	Path  []avp.Key
	Rule  string
	Cause error
}

func (e *FieldError) Error() string { return fmt.Sprintf("%s at %v: %v", e.Rule, e.Path, e.Cause) }
func (e *FieldError) Unwrap() error { return e.Cause }
func At(k avp.Key, err error) error {
	if err == nil {
		return nil
	}
	var f *FieldError
	if errors.As(err, &f) {
		p := make([]avp.Key, 1, len(f.Path)+1)
		p[0] = k
		p = append(p, f.Path...)
		return &FieldError{p, f.Rule, f.Cause}
	}
	return &FieldError{[]avp.Key{k}, "wire", err}
}

// State describes validation, independently of response submission or business success.
type State uint8

const (
	Unchecked State = iota
	NeedsApplication
	Ready
	Rejected
	Expired
)

var (
	ErrContext         = errors.New("invalid protocol context")
	ErrContextRequired = errors.New("protocol context required")
	ErrExpired         = errors.New("validation handle expired")
)

// Report is an owned, first-failure report for its explicit Scope. Complete does
// not assert full protocol coverage, peer authority or business acceptance.
// Current engine reports contain at most four checked rules, two missing names
// and one diagnostic. Diagnostics contain fixed classifications, never payloads.
type Report struct {
	Scope          string
	State          State
	Complete       bool
	CheckedRules   []string
	Violations     []error
	MissingContext []string
	LocalErrors    []error
}

// Clone owns all mutable report collections. Engine diagnostics are immutable
// sentinel errors; no callback error, FieldError pointer or message is retained.
func (r Report) Clone() Report {
	r.CheckedRules = append([]string(nil), r.CheckedRules...)
	r.Violations = append([]error(nil), r.Violations...)
	r.MissingContext = append([]string(nil), r.MissingContext...)
	r.LocalErrors = append([]error(nil), r.LocalErrors...)
	return r
}

// Failure records one bounded classification. Detailed paths remain available
// through the existing error APIs; this summary deliberately retains no input.
func (r *Report) Failure(err error) {
	r.MissingContext = nil
	r.Violations = nil
	r.LocalErrors = nil
	r.Complete = false
	r.State = Rejected
	if errors.Is(err, ErrContext) {
		r.LocalErrors = []error{ErrContext}
		return
	}
	if errors.Is(err, ErrUnsupported) {
		r.LocalErrors = []error{ErrUnsupported}
		return
	}
	for _, kind := range []error{ErrResultConflict, ErrUnmatchedAnswer, ErrSource, ErrEnvelope, ErrMissing, ErrDuplicate, ErrType, ErrInvalidAnswer} {
		if errors.Is(err, kind) {
			r.Violations = []error{kind}
			return
		}
	}
	r.Violations = []error{ErrField}
}

// Err returns the first bounded report classification, including missing context.
func (r Report) Err() error {
	if r.Complete {
		return nil
	}
	if len(r.LocalErrors) > 0 {
		return r.LocalErrors[0]
	}
	if len(r.Violations) > 0 {
		return r.Violations[0]
	}
	return ErrContextRequired
}
