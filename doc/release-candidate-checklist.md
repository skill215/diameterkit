<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: RC publication scope and requirement completion conditions.
SPDX-AI-Date: 2026-11-19
-->

# v0.2.0-rc.1 release checklist

For the current published pair, see the [rc.2 release record](releases/v0.2.0-rc.2.md) and [verified gate checklist](../spec/release/v0.2.0-rc.2.json). This page retains the rc.1 scope and the broader SPEC completion conditions.

This candidate ships the verified v0 subset. It does not complete the full
first-release/v1 scope. The [machine checklist](../spec/release/v0.2.0-rc.1.json)
and [release receipt](releases/v0.2.0-rc.1.md) carry current gate status.
All 24 broad SPEC requirements keep their separate full-release assessment.

## Required for this RC

| Gate | Completion condition |
| --- | --- |
| RC-SCOPE | Freeze selected support, exclusions, external evidence and all 24 completion conditions. |
| RC-API | Review changes since core rc.2 and retain exact Linux/Windows API snapshot; no Go source change in this release batch. |
| RC-CORE-CI | Seven checkout suites pass for the exact commit selected for the core tag. |
| RC-CORE-PUBLISH | Publish immutable core tag, verify remote commit, authenticated download and independent core-only cold consumer. |
| RC-ADAPTER-PIN | Pin adapter to the already published core; generate real sums and pass GOWORK=off installed-core pure/race/vet/API checks. |
| RC-ADAPTER-CI | Seven checkout suites pass for the exact commit selected for the adapter tag. |
| RC-ADAPTER-PUBLISH | Publish the separate immutable adapter tag and verify its commit. |
| RC-COLD-LOCAL | Both independently empty-cache consumers pass exact graph/API/source/checksum/pure/race/native checks without replace. |
| RC-COLD-HOSTED | Published-install workflow passes for the exact new core/adapter pair with artifact validation. |
| RC-RECEIPTS | Publish final commits/checksums/CI/install receipts and current English installation/migration guidance. |

Publish core first, verify its real download, then change the adapter requirement
and real checksums. Publish the adapter only from its reviewed checkout.
Neither local checkout replacements nor an old pair's receipts pass the new
installation gate. Never move/reuse an existing tag. Each failed attempt keeps
its evidence; an availability or network retry starts a new empty consumer.

The two modules use independent tags and both select v0.2.0-rc.1 here. The
adapter version reflects its new exact-core dependency contract; this batch
changes no Go implementation or public declarations. Earlier core changes,
including Redirect fields and Bundle fingerprints, require the v0 minor step.

## Explicitly unsupported in this RC

- Full SCUR/PRICE_ENQUIRY, S6a, full standalone RFC8506 behavior and unselected deep protocol subsets.
- Full Sh XSD/data model and internationalization; transparent User-Data is not semantic support.
- DRA/relay product, MAP/SIP/TPDU engine, charging/accounting service, business retry scheduler or library persistence.
- Persistent exactly-once, unapproved recovery, automatic cross-target authorization or universal six-template replacement.
- Native SCTP on Windows/386; arm64 cross-build evidence is not native runtime acceptance; SCTP DTLS is not enabled.
- Ordinary no-IMSI success SRA, unreviewed Realm/time-equality construction and restricted standard SMS failure/sequence or experimental 5011 generation branches.
- Automatic diagnostic mapping for unreviewed ambiguous/mixed conditions; safe-close/unsupported boundaries remain.
- Production throughput/latency/capacity, long-soak, multi-tenant fairness or mixed-transport guarantees from finite tests.

## External evidence, excluded from this RC's claims

These items are required before claiming the affected full capability or target
deployment. Their absence does not authorize an implementation to guess values
or silently expand recovery permissions. Application durability remains outside
the library.

| Evidence | Owner | Required material |
| --- | --- | --- |
| EXT-TEMPLATE | Template/target maintainer | Original six-template inputs/outputs and deployed versions; PSP/private ACA/1292 mappings and 420/2018 acceptance; Independent SCAP1 recovery authorization |
| EXT-RECOVERY | Target and application operators | Deduplication keys/content/domain/retention; ECUR continuity and DRA/terminal roles; Durable intent, old-worker fencing and unknown-result reconciliation |
| EXT-LOAD | Deployment acceptance owner | Declared duration/saturation/tenant/transport workloads and thresholds; Native target capacity/security/shutdown evidence |
| EXT-PROTOCOL | Protocol/compatibility reviewer | Direct normative or named compatibility evidence for restricted result/identity/Realm/time combinations |

## Completion conditions for every SPEC requirement

The RC column is a bounded acceptance contract, not a replacement for the full
first-release requirement. RC status becomes verified only after every mandatory
release gate passes. A full requirement may be marked complete only after its
full exit has evidence and review; a design boundary is not an unimplemented
storage or business feature.

| SPEC ID | This RC must demonstrate | Full first-release completion condition | External evidence / owner |
| --- | --- | --- | --- |
| SPEC-SCOPE-01 | Build both modules with CGO_ENABLED=0; disclose the exact profile/platform inventory and v0 support limits. | Approve the complete first-release A/B scope inventory and demonstrate every included capability; make a separate stable-v1 compatibility decision. | Project maintainer; target owners for claimed native platforms. |
| SPEC-SCOPE-02 | Public consumers keep subscriber, charging, retry authority and durable state outside the library; no persistence dependency is introduced. | Trace every first-release workflow to explicit library/application responsibility and exercise the corresponding public API without moving durable state into the library. | Application owners provide durable intent, fencing and reconciliation evidence; no library database is required. |
| SPEC-PROFILE-01 | All 12 reviewed binding identities, fixed wire types and definition hashes match the support matrix; reject conflicting bindings. | Resolve every required first-release field/type conflict through explicit definitions or named compatibility bindings, backed by source/target evidence and negative controls. | Original six-template source/version and target wire samples for disputed private fields. |
| SPEC-PROFILE-02 | Publish command, result, field and platform limits separately from Application-ID advertisement; verify independent public consumers. | Map the full first-release usage inventory to supported profiles and enforce every declaration/advertisement boundary with executable acceptance. | Target capability advertisements and configured endpoint roles where those claims are included. |
| SPEC-GEN-01 | Reproduce all 11 generated files offline, reject drift and preserve the reviewed generator 23 / ops 9/profile version ledger. | Cover all definitions required by the declared first-release inventory and audit handwritten-rule changes against explicit version/migration decisions; do not treat DefinitionDigest as a whole-source signature. | No external evidence for the existing generation mechanism; additional private definitions need authoritative inputs. |
| SPEC-CODEC-01 | Pass owned codec, malformed input, canonicalization and bounded fuzz tests for the declared slices, including exact diagnostic instances. | Complete the codec/vector/negative-case inventory for every required first-release field and binding, including authoritative private layouts. | Canonical target captures for private or disputed wire layouts. |
| SPEC-OWN-01 | Pass buffer mutation, snapshot, handoff, duplicate-arrival envelope and Close/Wait resource retirement tests in pure/race modes. | Provide a requirement-to-test inventory covering ownership and lifetime at every first-release public message, transport handoff and recovery boundary. | Application integration evidence for external buffer/worker ownership; library must not own application storage. |
| SPEC-VALID-01 | Keep framing/type/presence checks separate from domain and explicit Sh XML interpretation; reject unsupported cases. | Resolve and verify all validation rules within the declared first-release scope, while retaining explicitly excluded full-XSD/internationalization behavior. | Authoritative rules or named compatibility evidence for disputed Realm, timing and private relations. |
| SPEC-RESULT-01 | Public S6c/Ro consumers distinguish command, service, experimental and business outcomes; selected failure mappings remain bounded. | Complete result interpretation and paired request/answer tests for every first-release operation and required private result mapping. | PSP/private ACA/1292 and target result semantics; protocol success alone is insufficient. |
| SPEC-ERROR-01 | Selected 5001/5004/5005/5009/5014/3006 diagnostics pass bounds and original-instance tests; unreviewed failures retain safe close/unsupported behavior. | Enumerate each remaining promised condition/value/cardinality/group/domain/handshake/header mapping and either verify a bounded unambiguous diagnostic or record a justified safe-close classification. | Direct protocol or named target evidence where error attribution is ambiguous; do not guess repairs. |
| SPEC-APP-01 | Installed public consumers exercise selected Sh/S6c/SGd/Ro, S6c helpers and named recovery bindings without library persistence. | Cover the declared first-release A/B application calls, six-template operations, required outcomes and recovery entry points through public API acceptance. | Original complete template inputs/outputs, deployment versions and target application mappings. |
| SPEC-BIND-01 | Use explicit fixed bindings and reject wrong-binding imports or ungranted targets; preserve independent advertised applications. | Validate all first-release DRA/terminal roles, compatibility binding selections and target protection conditions against deployment fixtures. | DRA versus ultimate HSS/MME/OCS roles and target recovery permissions. |
| SPEC-EXCHANGE-01 | Verify correlation, deadlines, cancel/late-answer races, one completion and single-attempt default through public consumers. | Complete first-release flow/paired-message and fault-injection coverage across all claimed roles/transports; demonstrate actual deployment interoperation. | Named target request/answer traces and end-to-end fault scenarios. |
| SPEC-REPLAY-01 | Verify selected same-target/cross-target/import recovery with original identity, cumulative limits and explicit permission; no automatic durable retry. | Close the required RO-06/template recovery matrix with target deduplication domains, retention and ECUR continuity plus application fencing/unknown-result drills. | Real deduplication keys/content/retention, session continuity, independent SCAP1 authorization and application durable history. |
| SPEC-EVIDENCE-01 | Preserve write-stage uncertainty, final protocol result and one business notification; do not convert timeout or receipt into financial success. | Exercise every first-release uncertain-send/cancel/late-answer/failure transition and document its application interpretation with corresponding evidence. | Target financial/delivery outcomes and application reconciliation records. |
| SPEC-INBOUND-01 | Verify bounded receive/respond/reject, explicit result selection, Session/Proxy preservation and retirement without business session decisions. | Cover every inbound first-release operation/result and authorization boundary through the public responder API and target role fixtures. | Application-owned session decisions and target authorization mappings. |
| SPEC-DUP-01 | Pass combined cross-peer and four-profile identity/cache budget/expiry/conflict tests; preserve per-arrival envelopes. | Complete declared tenant/transport duplicate workloads and verify real target deduplication boundaries, without promising persistent exactly-once semantics. | Target identity/retention domains and declared tenant workload distributions. |
| SPEC-PEER-01 | Pass capability negotiation, watchdog, election, reconnect and named advertisement tests with no implicit replay. | Close all first-release peer/role/platform interoperation scenarios, including target capability and failover behavior. | Real peer advertisements, DRA topology and configured failure/protection policies. |
| SPEC-CLOSE-01 | Verify drain/close/wait, disconnected-primary handling and zero charged resources after retirement in pure/race tests. | Demonstrate shutdown and worker handover under declared long-soak/failure workloads with no stale ownership or hidden charged resources. | Deployment shutdown, fencing and process handover acceptance. |
| SPEC-ID-01 | Verify bounded ID exhaustion, active/quarantine intervals and explicit recovery-space import; reject conflicting imports. | Demonstrate first-release deployment-wide uniqueness and startup isolation across workers/restarts using application-provided identity history. | Boot/worker allocation, historical ID intervals and traffic isolation; no library durable allocator. |
| SPEC-BUDGET-01 | Pass finite admission/cache/diagnostic limits, peak-resource checks and isolation stress; distinguish logical bytes, heap and RSS. | Declare load duration, saturation levels, tenant mixes and mixed transports, then meet explicit resource/fairness/latency thresholds with reproducible measurements. | Deployment workload and capacity targets; finite benchmarks are not production sizing. |
| SPEC-TRANSPORT-01 | Pass native Linux/amd64 TCP/TLS/SCTP pure/race and pressure tests; separate cross-build evidence and unsupported stubs. | Obtain native and target interoperation evidence for every first-release advertised platform/transport combination, with explicit unsupported failures elsewhere. | Native target platforms, kernel capabilities and real DRA/OCS/HSS paths. |
| SPEC-SECURITY-01 | Verify frozen TLS policy, mTLS rejection/admission, no plaintext fallback and bounded malformed-input/resource handling. | Complete the first-release security/resource acceptance matrix for certificate ownership/callbacks, Diameter identity mapping, tenant isolation and configured trust boundaries. | Deployment certificates, identity authorization and security workload/threat assumptions. |
| SPEC-RELEASE-01 | Pass exact-candidate checkout CI, immutable core-first/adapter-second publication, independent local and hosted cold consumers, and publish truthful receipts. | Approve full-first-release scope/support policy, close required implementation and external gates, and record native/load/security/migration acceptance before any v1 claim. | Target and application acceptance for claims included in the full-first-release scope. |

The corresponding criteria are also stored on each requirement in
[progress.json](../spec/progress.json). Existing detailed evidence remains linked;
[history](../spec/history) preserves the earlier broad partial descriptions.
