<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Attribution and validation order for combined SMS request equality checks.
SPDX-AI-Date: 2026-11-24
-->

# Unequal timestamps and identities in combined SMS requests

The named combined binding requires two pairs of fields to agree in IEC EVENT,
ECUR INITIAL and ECUR TERMINATION requests. If an equality check fails, the
receiver closes the connection without an answer under both inbound failure
policies. Callers inspecting validation locally still get `ErrField`.

| Relation | Source input written to both fields | Existing local error path |
| --- | --- | --- |
| Event-Timestamp (55) equals MMS Submission-Time (10415, 1202) | `$TIMESTAMP` | 10415:1202 |
| The second Subscription-Id (type 1) has the same data (444) as User-Equipment-Info-Value (458/460) | `$SUBSCRIPTIONID_IMSI` | 458/460 |

The sender writes both fields from one input. The receiver has neither that input
nor a template rule identifying an authoritative wire instance. ECUR TERMINATION
uses a separate input for Discharge-Time. The local error path identifies the
comparison that failed. [IEC template projection](../interop/evidence/2026-11-07-smsfw-iec-source.json) · [ECUR template projections](../interop/evidence/2026-11-07-smsfw-ecur-source.json)

The archived negative fixtures mutate Submission-Time or the equipment value.
Their generators know which bytes they changed. An incoming frame carries no such
history. Starting with two valid requests containing `(A,A)` and `(B,B)`, changing
one field in each can produce exactly the same `(A,B)` frame, including headers
and padding. The acceptance tests demonstrate both directions for each pair in
all three operations. Neither difference supplies evidence for a field-specific
`Failed-AVP`. [IEC fixture generator](../interop/smsfw-iec-fixtures.py) · [ECUR fixture generator](../interop/smsfw-ecur-fixtures.py)

## Mixed failures

Group validation runs before request-level comparisons. An independently invalid
NUL-containing subscription or equipment value retains its bounded 5004 answer
when the automatic policy is enabled. In the request validator, the equipment
comparison precedes the SMS client checks, which precede the timestamp comparison.
The timestamp comparison precedes the unit-value rules. With two mismatches, the
equipment comparison wins. [Validation order and loopback tests](../reject_combined_equality_test.go)

Run the focused tests in pure Go and with the race detector from the repository root.

```sh
CGO_ENABLED=0 go test -count=1 -run '^TestCombinedEquality' . ./compat/smsfwro/combined
CGO_ENABLED=1 go test -race -count=1 -run '^TestCombinedEquality' . ./compat/smsfwro/combined
```

The acceptance covers the development checkout and local TCP/TLS peers. The
combined binding remains at slice.30. The published rc.2 pair ships slice.29.
