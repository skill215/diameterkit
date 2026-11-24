# Current remaining gaps

This is the current implementation overview for **batch 135, Draft 0.136**
(2026-11-24). [Progress](../spec/progress.json) records actual check completion;
the broad release status remains `not_ready`. The 24 SPEC IDs span implementation,
verification and external acceptance; their `partial` labels are not a percentage
of unfinished code. Dated specification sections and frozen batch records retain
their historical meaning.

## Current RC and full-release work

[The RC checklist](release-candidate-checklist.md) separates required publication
gates from unsupported and external-evidence claims, with explicit completion
conditions for all 24 SPEC requirements. Batch 121 verified the published pair through ordered publication and cold installation.
Batches 122 through 135 add unpublished conditional/value diagnostics, including money SMS-Node, Service-Identifier/Timezone, client bytes, MMS SMSKey, equipment identity, address, subscription and SMS address data values, plus SMS client wire conditions; the earlier RC receipt does not
certify the newer checkout. A future release needs fresh publication/install evidence.
The broad not_ready status refers to the full first-release scope, not a failed
checkout or a requirement to complete every diagnostic before a v0 candidate.

Batch 107 repairs eight stale support-matrix definition hashes and adds
[executable consistency checks](support-matrix-checks.md) for all twelve bindings.
This closes a release-record integrity gap; protocol and external acceptance
boundaries below remain unchanged.

| Area | Completed selected scope | Remaining work |
| --- | --- | --- |
| Error diagnostics | Selected Base and generated fixed/variable-field 5005; reviewed whole-group minima, including three groups with variable children; first-excess 5009; isolated-value 5004, unknown-M 5001 and typed-length 5014; bounded Redirect 3006; selected Proxy/VSAI leaf, condition and omission diagnostics; selected SRR conditional SM-RP-SMEA omission; combined two-subscription shortage and SCAP1 ordered-parameter shortage; isolated conditional combined Recipient-Address, five MMS fields, five common SMS fields, two termination fields isolated IEC Originator-SCCP omission conditional whole MMS-Information omission and EVENT Requested-Action omission; selected IEC/money EVENT action-value 5004; isolated IEC and INITIAL/TERMINATION interface-type 5004; isolated SMS-Node and SM-Message-Type 5004; isolated Reporting-Reason and CC-Service-Specific-Units 5004; isolated Multiple-Services-Indicator 5004; isolated User-Equipment-Info-Type 5004; isolated Address-Type 5004; isolated money SMS-Node 5004; isolated Service-Identifier/Timezone 5004; bounded money ClientAddressWire 5004; isolated MMS SMSKey NUL 5004; isolated equipment Value NUL 5004; isolated Address-Data NUL 5004; repeated Subscription-Id-Data NUL 5004; isolated SMS address NUL 5004; SMS ClientAddressWire NUL/mismatch 5004 | Other conditional/cardinality and unreviewed group omissions, domain/handshake mappings and malformed-header length diagnostics |
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
[Batch 126](conditional-address-values.md) adds isolated Originator/Recipient
Address-Type value diagnostics with NUL-free Address-Data and no extra selected
group children, preserving mixed errors and cross-group/wire validation order.
[Batch 127](conditional-money-node-values.md) adds isolated money SMS-Node value
diagnostics when all seven forbidden SMS fields are absent. Earlier action/unit
conditions, mixed errors and later money conditions retain their order.
[Batch 128](conditional-money-parameter-values.md) adds isolated Service-Identifier
and Timezone value diagnostics for general/balance requests. Both-bad values and
mixed presence conditions keep bare ErrField; earlier rules and later client/MMS
checks retain their order.
[Batch 129](conditional-money-client-values.md) adds bounded wrong client-byte
diagnostics for general/balance requests, including variable lengths and original
padding. Earlier conditions and structural/group rules retain priority; quotes
fit completely or the connection closes without a truncated answer.
[Batch 130](conditional-mms-key-values.md) adds isolated MMS SMSKey NUL-value
diagnostics with original padded-key quotes. Mixed extra children and the existing
structural/group/request validation order retain their behavior.
[Batch 131](conditional-equipment-data-values.md) adds isolated equipment Value
NUL diagnostics with Type0 and no extra children. Mixed type/value errors,
structural failures, wire order and later request relations retain their behavior.
[Batch 132](conditional-address-data-values.md) adds Originator/Recipient Address-Data
NUL diagnostics with Type1 and no extras. Mixed type/data errors and structural,
wire and group order retain their behavior; an earlier address error keeps
priority over a later MMS key error.
[Batch 133](conditional-subscription-data-values.md) adds Subscription-Id-Data NUL
diagnostics, resolving the first failing repeated group in wire order. Extra
children and earlier structural failures retain priority; later type pairing
and request relations remain separate.
[Batch 134](conditional-sms-data-values.md) adds isolated Originator-SCCP,
Recipient-SCCP and SMSC address NUL diagnostics. Multiple bad address fields
and extra children keep their original group error; earlier structural/wire
failures and later request relations retain their order.
[Batch 135](conditional-sms-client-values.md) adds SMS ClientAddressWire NUL
and NUL-free SMSC mismatch diagnostics for IEC, INITIAL and TERMINATION.
Prior presence/node/group errors, NUL-before-equality order and later MMS
relations remain unchanged; money binary client rules stay independent.
Other conditional, cardinality, domain,
handshake and malformed-header mappings remain separate;
do not infer them from bare sentinels.
Keep malformed/untrusted/resource-limit failures closed when no safe reply exists.

## Verification observation

Batch 133 observed one existing legacy420 refund/restart import refusal during
local full pure acceptance. The exact case passed 100 pure and 100 race
repetitions; its cause remains unconfirmed. The failed receipt and final suite
results are retained in [batch evidence](../spec/evidence/batch-133-subscription-data-values.json).
No recovery authorization or clock guard was weakened. Investigate this case if
it recurs; a successful rerun is not a root-cause fix.

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
