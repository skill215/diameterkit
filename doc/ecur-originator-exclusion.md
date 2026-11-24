<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Combined ECUR Originator-SCCP-Address exclusion and bounded wire diagnostics.
SPDX-AI-Date: 2026-11-24
-->

# Originator-SCCP-Address in combined ECUR requests

With `BaseErrorFailures`, the named combined profile returns 5008 when an
otherwise valid ECUR INITIAL or TERMINATION request contains an extra
Originator-SCCP-Address (Vendor 10415, code 2008). The originating interface
must have type 3 and the destination interface type 2. Both archived negative
fixtures add only this AVP to a valid request. [Fixture generator](../interop/smsfw-ecur-fixtures.py) · [INITIAL input](../testdata/codec/b43-init-origin-sccp.hex) · [TERMINATION input](../testdata/codec/b43-term-origin-sccp.hex)

| Received condition | Outcome under automatic Base errors |
| --- | --- |
| Sole extra address, empty or any NUL-free byte string | 5008 with the original padded 2008 inside 873/2000. |
| Address contains NUL and the earlier group rule identifies it | Existing 5004 for the malformed address. |
| Extra address plus a wrong interface type or a later unit/status error | Connection closes. |
| Extra address plus a timestamp mismatch or forbidden Requested-Action | The earlier comparison or action check takes precedence and the connection closes. |

The wire scan checks the request sequence, absence of Requested-Action,
SMS markers, both interface types and the exact address instance using shared
decoder limits. Failed-AVP embeds two ancestor headers and the original AVP,
including padding. This requires 36 bytes plus the address payload length
rounded up to a multiple of four. An insufficient quote or full-answer budget
closes the connection. `CloseOnFailure` closes the connection for an isolated
exclusion as well. Build and Send preserve their existing error path.

Run the focused checks from the repository root.

```sh
CGO_ENABLED=0 go test -count=1 -run '^Test(ECUROriginator|ECURInterfaceValue|ECURAction|SMSClient|SMSData)' . ./compat/smsfwro/combined ./internal/excessavp
CGO_ENABLED=1 go test -race -count=1 -run '^Test(ECUROriginator|ECURInterfaceValue|ECURAction|SMSClient|SMSData)' . ./compat/smsfwro/combined ./internal/excessavp
```

The change is in development slice.32. The published rc.2 pair contains slice.29.
