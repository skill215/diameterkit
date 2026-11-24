<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Inbound service-unit group exclusions for the combined ECUR templates.
SPDX-AI-Date: 2026-11-24
-->

# Extra service-unit groups in ECUR requests

With `BaseErrorFailures`, a combined ECUR peer receives `5008` when an otherwise valid request contains the service-unit group prohibited for its operation. Both archived negative requests add exactly one group containing a single UInt64 CC-Service-Specific-Units (417). [Fixture generator](../interop/smsfw-ecur-fixtures.py)

| Request | Required MSCC content | Extra group quoted in Failed-AVP |
| --- | --- | --- |
| INITIAL (1/0) | Requested-Service-Unit with 417 equal to one | [Used-Service-Unit (446)](../testdata/codec/b43-init-usu.hex) |
| TERMINATION (3/1) | Used-Service-Unit and Reporting-Reason equal to FINAL (2) | [Requested-Service-Unit (437)](../testdata/codec/b43-term-rsu.hex) |

The error identifies the presence of the extra group, so its UInt64 value may span the full range. The response copies the original 24 bytes of that group under an eight-byte Multiple-Services-Credit-Control (456) header. A 32-byte quote budget is sufficient. A 31-byte budget or insufficient space for the complete answer closes the connection. The SMS and MSCC branches share the request decoder’s depth, field-size and AVP-count limits.

A missing required unit, an invalid required-unit value or reporting reason, a conflicting status/discharge combination, or a money-bearing requested unit prevents this mapping. Earlier validation errors keep their precedence. An extra Requested-Action together with an extra unit still closes at Requested-Action. `CloseOnFailure` closes the connection even for an isolated exclusion. Build and Send report the existing unsupported MSCC error.

Development slice.34 contains these diagnostics. The published rc.2 pair contains slice.29. Run the targeted checks from the repository root.

```sh
CGO_ENABLED=0 go test -count=1 -run '^Test(ECURUnit|InitialStatus|RejectUnitValue|ECUROriginator|UnitValue|ECURAction)' . ./compat/smsfwro/combined ./internal/excessavp
CGO_ENABLED=1 go test -race -count=1 -run '^Test(ECURUnit|InitialStatus|RejectUnitValue|ECUROriginator|UnitValue|ECURAction)' . ./compat/smsfwro/combined ./internal/excessavp
CGO_ENABLED=0 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -parallel=2 -fuzzminimizetime=5s -timeout=2m
```
