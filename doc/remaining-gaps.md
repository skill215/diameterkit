# Current remaining gaps

This is the current implementation overview for **batch 125, Draft 0.126**
(2026-11-19). [Progress](../spec/progress.json) records actual check completion;
the broad release status remains `not_ready`. The 24 SPEC IDs span implementation,
verification and external acceptance; their `partial` labels are not a percentage
of unfinished code. Dated specification sections and frozen batch records retain
their historical meaning.

## Current RC and full-release work

[The RC checklist](release-candidate-checklist.md) separates required publication
gates from unsupported and external-evidence claims, with explicit completion
conditions for all 24 SPEC requirements. Batch 121 verified the published pair through ordered publication and cold installation.
Batches 122 through 125 add unpublished SMS, unit, indicator and equipment-type diagnostics; the earlier RC receipt does not
certify the newer checkout. A future release needs fresh publication/install evidence.
The broad not_ready status refers to the full first-release scope, not a failed
checkout or a requirement to complete every diagnostic before a v0 candidate.

Batch 107 repairs eight stale support-matrix definition hashes and adds
[executable consistency checks](support-matrix-checks.md) for all twelve bindings.
This closes a release-record integrity gap; protocol and external acceptance
boundaries below remain unchanged.

| Area | Completed selected scope | Remaining work |
| --- | --- | --- |
| Error diagnostics | Selected Base and generated fixed/variable-field 5005; reviewed whole-group minima, including three groups with variable children; first-excess 5009; isolated-value 5004, unknown-M 5001 and typed-length 5014; bounded Redirect 3006; selected Proxy/VSAI leaf, condition and omission diagnostics; selected SRR conditional SM-RP-SMEA omission; combined two-subscription shortage and SCAP1 ordered-parameter shortage; isolated conditional combined Recipient-Address, five MMS fields, five common SMS fields, two termination fields isolated IEC Originator-SCCP omission conditional whole MMS-Information omission and EVENT Requested-Action omission; selected IEC/money EVENT action-value 5004; isolated IEC and INITIAL/TERMINATION interface-type 5004; isolated SMS-Node and SM-Message-Type 5004; isolated Reporting-Reason and CC-Service-Specific-Units 5004; isolated Multiple-Services-Indicator 5004; isolated User-Equipment-Info-Type 5004 | Other conditional/cardinality and unreviewed group omissions, domain/handshake mappings and malformed-header length diagnostics |
| Duplicate acceptance | Generic and combined cross-peer TCP/TLS tests; four-profile identity, simultaneous arrivals and shared/independent slot/byte cache acceptance | Larger tenant and mixed-transport workloads, real-target deduplication domains and durable application behavior |
| Load acceptance | Seventeen count-bounded fuzz targets with bounded observed work and retained failure receipts, pure/race finite TCP/SCTP stress, four-application TCP isolation and measured benchmarks | Declared long-soak, saturation, tenant fairness and mixed-transport workloads |
| Publication | Reviewed API/generated output and exact dual-module checkout CI; core/adapter v0.2.0-rc.1 published and cold-installed locally and on hosted CI | Full-first-release scope and external acceptance; future versions need their own publication receipts |

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
[Batch 108](conditional-missing-diagnostics.md) adds receive-only evidence for the
SRR flag-conditioned SM-RP-SMEA omission and its three-byte minimum. [Batch 109](cardinality-diagnostics.md) adds combined two-subscription shortage
proofs and bounded examples. Batch 110 adds SCAP1 ordered-parameter shortages,
including all 2047 proper ordered subsets. [Batch 111](conditional-recipient-diagnostics.md)
adds isolated conditional Recipient-Address omissions for combined SMS templates.
[Batch 112](conditional-mms-diagnostics.md) adds the five remaining isolated MMS
field omissions, with all 63 nonempty missing-field combinations checked per template.
[Batch 113](conditional-sms-diagnostics.md) adds the five common SMS fields and
the two termination-only fields, preserving first-failure order between conditions.
[Batch 114](conditional-originator-diagnostics.md) adds IEC Originator-SCCP absence
only when the existing interface condition is satisfied, retaining ErrUnsupported.
[Batch 115](conditional-mms-group-diagnostics.md) adds the whole MMS-Information
minimum for SMS templates, including structural Originator-Address and all six
conditional children.
[Batch 116](conditional-action-diagnostics.md) adds absent Requested-Action in
IEC and money EVENT paths, retaining the prior unsupported classification and
present-value behavior.
[Batch 118](conditional-action-values.md) adds selected IEC and money EVENT
action-value diagnostics with exact original-instance quotes, preserving validation order.
[Batch 119](conditional-interface-values.md) adds isolated IEC interface-type
value diagnostics when the Originator-SCCP address is present, with a precise
Receive path and a 52-byte original-instance quote.
[Batch 120](ecur-interface-values.md) adds the isolated INITIAL/TERMINATION
originator and destination interface-value branches, checking the valid
counterpart with shared source counts. [Batch 122](conditional-sms-values.md) adds isolated SMS-Node and SM-Message-Type
value diagnostics for the three SMS templates. Simultaneous bad values, money
relations and other conditions retain their separate boundaries.
[Batch 123](conditional-unit-values.md) adds isolated TERMINATION Reporting-Reason
and IEC/INITIAL requested-unit value diagnostics. Mixed presence/value failures
and validation order remain unchanged.
[Batch 124](conditional-indicator-values.md) adds isolated Multiple-Services-Indicator
value diagnostics under the existing SMS marker conditions. It also closes batch
123 hosted artifact content verification; the original access failure remains
historical evidence.
[Batch 125](conditional-equipment-values.md) adds isolated User-Equipment-Info-Type
value diagnostics with NUL-free Value and no extra group fields, preserving
mixed group errors and structural wire-order precedence.
Other conditional, cardinality, domain,
handshake and malformed-header mappings remain separate;
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

Generator v23 and common operations `ops9` fingerprints differ from earlier builds. Existing published tags
remain immutable, and earlier recovery records require their original build.
See the [version policy](version-install-policy.md),
[5005 guide](missing-avp-diagnostics.md), [5009 guide](excess-avp-diagnostics.md),
[5004/5001 guide](invalid-avp-diagnostics.md), [5014 guide](length-avp-diagnostics.md),
[Redirect guide](redirect-handling.md) and [manual-group guide](manual-group-diagnostics.md).

[Batch 117](count-bounded-fuzz.md) changes the default gate to one million
executions per target and preserves partial receipts, timeout logs and failure
corpora. The original timed failures in batches 114 and 116 remain recorded;
their upstream root cause is unproven. Counted acceptance avoids selecting that
timed-stop path without suppressing failures or claiming equivalent coverage.

Batch 118 corrects the counted gate's exact-equality assumption using the pinned
Go minimization bound: up to31 additional calls per worker,62 for two workers.
The original hosted failure remains recorded; nonzero exits are never waived.
