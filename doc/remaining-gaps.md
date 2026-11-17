# Current remaining gaps

This is the current implementation overview for **batch 106, Draft 0.107**
(2026-11-17). [Progress](../spec/progress.json) records actual check completion;
the broad release status remains `not_ready`. The 24 SPEC IDs span implementation,
verification and external acceptance; their `partial` labels are not a percentage
of unfinished code. Dated specification sections and frozen batch records retain
their historical meaning.

## Next local work

| Area | Completed selected scope | Remaining work |
| --- | --- | --- |
| Error diagnostics | Selected Base and generated fixed/variable-field 5005; reviewed whole-group minima, including three groups with variable children; first-excess 5009; isolated-value 5004, unknown-M 5001 and typed-length 5014; bounded Redirect 3006; selected Proxy/VSAI leaf, condition and omission diagnostics | Other conditional/cardinality and unreviewed group omissions, domain/handshake mappings and malformed-header length diagnostics |
| Duplicate acceptance | Generic and combined cross-peer TCP/TLS tests; four-profile identity, simultaneous arrivals and shared/independent slot/byte cache acceptance | Larger tenant and mixed-transport workloads, real-target deduplication domains and durable application behavior |
| Load acceptance | Fifteen bounded fuzz targets, pure/race finite TCP/SCTP stress, four-application TCP isolation and measured benchmarks | Declared long-soak, saturation, tenant fairness and mixed-transport workloads |
| Publication | Reviewed API/generated output and dual-module checkout CI; immutable rc.2/rc.1 installation receipts | Publish and cold-install a new reviewed core/adapter pair after the generator fingerprint transitions; planned core candidate `v0.2.0-rc.1` |

The selected combined cross-peer and mixed-profile cache slice is now covered by
[batch 99 acceptance](cross-peer-duplicate-acceptance.md). The [batch 100 extension](length-avp-diagnostics.md) now covers selected generated
typed-length 5014 errors. [Batch 101](redirect-handling.md) adds bounded Redirect
3006 URI/required-field handling. [Batch 102](manual-group-diagnostics.md) adds
selected handwritten Proxy/VSAI diagnostics and cross-package exact-instance
evidence. [Batch 103](group-constraint-diagnostics.md) adds conditional Auth/Acct
pairs and minimum Proxy omission examples. [Batch 104](minimum-group-diagnostics.md)
adds explicitly reviewed whole-group minima. [Batch 105](variable-missing-diagnostics.md)
adds eleven explicit variable-field minima. [Batch 106](variable-group-diagnostics.md)
adds typed whole-group plans for three reviewed groups with variable children.
Next, prioritize selected conditional missing fields with explicit rule evidence. Other conditional,
cardinality, domain, handshake and malformed-header mappings remain separate;
do not infer them from bare sentinels.
Keep malformed/untrusted/resource-limit failures closed when no safe reply exists.

## Evidence-dependent work

- Original six-template inputs, complete requests/answers and deployment versions;
  PSP/private ACA/1292 mappings; 420/2018 target acceptance; independent SCAP1
  recovery authorization.
- Target recovery deduplication keys/content/retention, ECUR continuity and
  DRA/ultimate-terminal roles. Local simulated peers cannot provide these claims.
- Application-owned durable intent, old-worker fencing and unknown-result
  reconciliation. These are deployment/migration duties, not library persistence.
- Existing restricted protocol branches: standard SMS failure/sequence fields,
  normal CCA experimental 5011 generation pairing, ordinary no-IMSI success SRA,
  disputed Realm combinations and requested/maximum-time equality. Extend only
  with the required direct specification or named compatibility evidence.

Full SCUR/PRICE_ENQUIRY, S6a and full Sh XSD/internationalization are deferred
capabilities, not newly added blockers for the selected first-release scope.

## Already implemented

S6c `NewSRR`, `InterpretSRA` and `client/s6c` were completed in batch 87. Selected
Ro recovery and independent binding/no-answer/import public consumption were
completed in batches 88–90. Dual-module CI, generated/API/race checks, bounded
load tooling and historical immutable cold installs have passed recorded gates.
Do not convert dated statements that these were missing into new implementation
tasks. Native SCTP and platform stubs remain explicitly distinguished in the
[support matrix](diameterkit-support-matrix.md).

Generator v22 and common operations `ops9` fingerprints differ from earlier builds. Existing published tags
remain immutable, and earlier recovery records require their original build.
See the [version policy](version-install-policy.md),
[5005 guide](missing-avp-diagnostics.md), [5009 guide](excess-avp-diagnostics.md),
[5004/5001 guide](invalid-avp-diagnostics.md), [5014 guide](length-avp-diagnostics.md),
[Redirect guide](redirect-handling.md) and [manual-group guide](manual-group-diagnostics.md).
