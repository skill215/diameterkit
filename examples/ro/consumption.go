package main

import (
	"errors"
	diameter "github.com/skill215/diameterkit"
	"github.com/skill215/diameterkit/compat/smsfwro/combined"
)

// These are the five original business operations, not Diameter command types.
type businessOperation uint8

const (
	businessCheck businessOperation = iota
	businessCharge
	businessReserve
	businessCommit
	businessCancel
)

type localReason uint8

const (
	txTimeout localReason = iota
	connectionDown
	temporaryError
	systemFailure
)

type fallbackConfig struct {
	TxTimeout, ConnectionDown, TemporaryError int
	ErrorCause                                uint32
	LogLSI                                    bool
}
type businessDecision struct {
	Status     string
	Reason     localReason
	ErrorCause uint32
	LogLSI     bool
	Exchange   diameter.Result // Never rewrite protocol code or DeliveryEvidence.
}

// Compatibility mapping from mainflow-audit §33.1, confined to the example.
// SYSTEM_FAILURE is conservatively reported, not mapped to financial success;
// its application-supplied diagnostics are retained as an explicit migration choice.
func consumeLocal(op businessOperation, reason localReason, c fallbackConfig, result diameter.Result) (businessDecision, error) {
	d := businessDecision{Reason: reason, ErrorCause: c.ErrorCause, LogLSI: c.LogLSI, Exchange: result}
	if op > businessCancel || reason > systemFailure {
		return d, errors.New("unsupported business operation/reason")
	}
	success := [5]string{"CHECK_SUCCESSFUL", "COMMITTED", "RESERVATION_SUCCESSFUL", "COMMITTED", "CANCELLED"}
	failure := [5]string{"CHECK_FAILED", "COMMIT_FAILED", "OTHER_ERROR", "COMMIT_FAILED", "CANCEL_FAILED"}
	fallback := [5]string{"NOT_CHARGED", "NOT_CHARGED", "NOT_CHARGED", "COMMITTED", "CANCELLED"}
	switch reason {
	case txTimeout:
		switch c.TxTimeout {
		case 0:
			d.Status = failure[op]
		case 1:
			d.Status = success[op]
		default:
			d.Status = fallback[op]
		}
	case connectionDown:
		if c.ConnectionDown == 0 {
			failure[1] = "OTHER_ERROR"
			d.Status = failure[op]
		} else {
			d.Status = fallback[op]
		}
	case temporaryError:
		if c.TemporaryError == 1 {
			d.Status = fallback[op]
		} else {
			d.Status = success[op]
		}
	case systemFailure:
		d.Status = "SYSTEM_FAILURE"
	}
	return d, nil
}

// Single application owner; a late protocol result still needs consumption even
// when Tx already delivered the sole business notification.
type businessConsumer struct{ notified bool }

func (c *businessConsumer) notify(d businessDecision) bool {
	if c.notified {
		return false
	}
	c.notified = true
	return true
}

// DDFH interpretation is limited to a validated DIRECT_DEBITING request/result.
// Nonanswers and E answers use configuration, never a fabricated remote value.
func debitDDFH(request *combined.CCR, result diameter.Result, configured uint32) (uint32, error) {
	if request == nil || request.CCRequestType != 4 || (!request.RequestedAction.Present || request.RequestedAction.Value != 0) || configured > 1 {
		return 0, diameter.ErrConfig
	}
	if result.Outcome != diameter.AnswerReceived {
		return configured, nil
	}
	if result.Err != nil || !result.Validation.Complete {
		return 0, errors.New("unvalidated answer")
	}
	if a, ok := result.Answer.Body.(*combined.CCA); ok {
		if a.DirectDebitingFailureHandling.Present {
			if a.DirectDebitingFailureHandling.Value > 1 {
				return 0, diameter.ErrConfig
			}
			return a.DirectDebitingFailureHandling.Value, nil
		}
	}
	return configured, nil
}

// A request to the surrounding application, never a claim of successful saving.
// Named sms-fw DIRECT_DEBITING policy only, after eligibility/phase checks.
func debitNoBackup(ddfh uint32, qualifiedError bool) (requestSave bool, status string, err error) {
	if ddfh > 1 {
		return false, "", diameter.ErrConfig
	}
	if qualifiedError {
		if ddfh == 1 {
			return true, "temporary-error", nil
		}
		return true, "actual-error", nil
	}
	return ddfh == 0, "no-answer", nil
}
