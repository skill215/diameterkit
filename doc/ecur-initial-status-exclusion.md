<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Receive diagnostics for an extra SM-Status in a combined ECUR INITIAL request.
SPDX-AI-Date: 2026-11-24
-->

# SM-Status in a combined ECUR INITIAL request

With `BaseErrorFailures`, the named combined profile returns 5008 if an otherwise
valid ECUR INITIAL request contains SM-Status (Vendor 10415, code 2014). Its
payload is opaque and may be empty or contain NUL. The archived negative fixture
adds a single byte, `00`, to a valid INITIAL as the payload of a new SM-Status.
TERMINATION has its own required status and discharge fields.
[Fixture generator](../interop/smsfw-ecur-fixtures.py) · [INITIAL negative fixture](../testdata/codec/b43-init-status.hex)

| INITIAL request | Inbound outcome with automatic Base errors |
| --- | --- |
| Isolated extra SM-Status | 5008. Failed-AVP contains the original padded status inside Service-Information and SMS-Information. |
| Discharge-Time alone, or both fields in either order | Connection closes. |
| Extra status and an invalid CC-Service-Specific-Units value in an otherwise valid Requested-Service-Unit | The earlier 5004 quotes the unit value. |
| Extra status plus a Used-Service-Unit, forbidden Requested-Action or extra Originator-SCCP-Address | The earlier group or field check wins and the connection closes. |

The bounded scan rechecks the INITIAL type/number pair (1/0), the single
requested unit with value one, the SMS markers, the two interface types and
the unique status instance. Its decoder budget covers both the unit and SMS
branches. The quote needs 36 bytes plus the payload length rounded up to four
bytes. Insufficient quote or full-answer space closes the connection.
`CloseOnFailure` also closes it for the isolated status. Build and Send keep
their existing group-level unsupported error.

The change is in development slice.33. The published rc.2 pair contains slice.29.
Run these checks from the repository root.

```sh
CGO_ENABLED=0 go test -count=1 -run '^Test(InitialStatus|ECUROriginator|UnitValue|ECURAction)' . ./compat/smsfwro/combined ./internal/excessavp
CGO_ENABLED=1 go test -race -count=1 -run '^Test(InitialStatus|ECUROriginator|UnitValue|ECURAction)' . ./compat/smsfwro/combined ./internal/excessavp
```
