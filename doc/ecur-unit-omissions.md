<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Bounded conditional missing-group diagnostics for combined ECUR.
SPDX-AI-Date: 2026-11-24
-->

# Missing service-unit groups in combined ECUR

With `BaseErrorFailures`, the combined profile returns `5005` when the required service-unit group is the sole omission in an otherwise valid ECUR request. Both archived fixtures retain Multiple-Services-Credit-Control (MSCC, 456). [Fixture generator](../interop/smsfw-ecur-fixtures.py)

| Request | Missing group | Conditions at the unit check |
| --- | --- | --- |
| INITIAL (1/0) | [Requested-Service-Unit (437)](../testdata/codec/b43-init-no-rsu.hex) | MSCC is empty. SM-Status and Discharge-Time are absent. |
| TERMINATION (3/1) | [Used-Service-Unit (446)](../testdata/codec/b43-term-no-usu.hex) | MSCC contains only the FINAL Reporting-Reason. SM-Status and Discharge-Time are present. |

Failed-AVP contains a 32-byte minimum example. Eight bytes encode the MSCC header, eight encode the missing group header and sixteen encode CC-Service-Specific-Units (417) with an eight-byte zero placeholder. An actual INITIAL request still requires one unit. If the group exists but lacks 417, the existing structural diagnostic identifies the child. The resulting minimum bytes match, while the validation paths distinguish the two failures.

The group, field, frame, AVP-count and depth budgets must permit the complete example. A 31-byte quote budget or a one-byte-short answer budget closes the connection. A missing group paired with a contradictory unit, reporting reason or status keeps the original MSCC error. An earlier Requested-Action error retains priority. Build and Send keep their previous unsupported error. Receive adds missing-field evidence and preserves the unsupported sentinel. The default `CloseOnFailure` policy closes for the isolated omission.

This behavior ships in development slice.35. The published rc.2 pair contains slice.29. Run from the repository root.

```sh
CGO_ENABLED=0 go test -count=1 -run '^Test(ECURUnit|MinimumVariableTree|MMSGroupMissing|RejectUnitValue|ECUROriginator|UnitValue|ECURAction)' . ./compat/smsfwro/combined ./internal/missingavp
CGO_ENABLED=1 go test -race -count=1 -run '^Test(ECURUnit|MinimumVariableTree|MMSGroupMissing|RejectUnitValue|ECUROriginator|UnitValue|ECURAction)' . ./compat/smsfwro/combined ./internal/missingavp
CGO_ENABLED=0 go test ./internal/missingavp -run '^$' -fuzz '^FuzzMinimumVariableTree$' -fuzztime=1000000x -parallel=2 -fuzzminimizetime=5s -timeout=2m
```
