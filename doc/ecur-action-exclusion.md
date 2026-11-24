<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Receive behavior for a prohibited ECUR Requested-Action.
SPDX-AI-Date: 2026-11-24
-->

# Requested-Action in combined ECUR requests

In the named combined profile, an otherwise valid INITIAL (type 1, number 0)
or TERMINATION (type 3, number 1) request containing Requested-Action (436)
can receive `DIAMETER_AVP_NOT_ALLOWED` (5008). Select `BaseErrorFailures` to
enable this answer. `CloseOnFailure` closes the connection. This behavior is in
development slice.31. The published rc.2 pair contains slice.29.

Both archived ECUR negative fixtures add exactly one Requested-Action, with value
zero, to the original request. The combined validator prohibits its presence for
both operations regardless of the four-byte value. RFC 8506 describes the purpose
of Requested-Action for EVENT requests. [Fixture generator](../interop/smsfw-ecur-fixtures.py) · [INITIAL input](../testdata/codec/b43-init-action.hex) · [TERMINATION input](../testdata/codec/b43-term-action.hex) · [RFC 8506 §8.41](https://www.rfc-editor.org/rfc/rfc8506.html#section-8.41)

The receiver scans the complete top-level AVP area within the configured limits.
It checks the type/number pair, SMS markers and one mandatory four-byte action
before copying the original 12-byte AVP into Failed-AVP. The full answer has a
separate size check. Insufficient space takes the existing close path.

Earlier errors still take precedence. For example, a timestamp mismatch closes
the connection and an independently malformed equipment value can produce 5004.
A typed-length failure on 436 retains its 5014 mapping. A later operation error,
such as an invalid interface or unit, leaves the action error unquoted.
Build and Send retain their existing unsupported-field error. EVENT value and
omission handling continue to use their existing checks.

```sh
CGO_ENABLED=0 go test -count=1 -run '^Test(ECURAction|ActionValue|ActionMissing|CombinedEquality)' . ./compat/smsfwro/combined ./internal/excessavp
CGO_ENABLED=1 go test -race -count=1 -run '^Test(ECURAction|ActionValue|ActionMissing|CombinedEquality)' . ./compat/smsfwro/combined ./internal/excessavp
```
