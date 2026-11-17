# Versioned protocol definitions

Diameter's wire version, Application-ID, specification edition, profile slice
version and Go module SemVer are different identifiers. A supported Application-ID
does not imply full support for all procedures in that application.

The library compiles JSON definitions at build time into concrete Go messages and
profile descriptions. It does not load runtime XML dictionaries or reflect over
message values. Handwritten rules enforce selected contextual and cross-field
contracts. Sh-Data XML is an explicit AVP payload layer, not the definition format.

## Definition entry points

- **base**: [`spec/base.json`](../spec/base.json), profile `base-rfc6733-controls-slice-v1`, slice `0.1.0-slice.8`.
- **sh**: [`spec/sh.json`](../spec/sh.json), profile `sh-29.329-sms-opaque-slice-v1`, slice `0.1.0-slice.11`.
- **s6c**: [`spec/s6c.json`](../spec/s6c.json), profile `s6c-29.338-ordinary-mme-slice-v1`, slice `0.1.0-slice.22`.
- **sgd**: [`spec/sgd.json`](../spec/sgd.json), profile `sgd-gdd-29.338-ordinary-slice-v1`, slice `0.1.0-slice.15`.
- **ro**: [`spec/ro.json`](../spec/ro.json), profile `ro-32.299-sms-iec-slice-v1`, slice `0.1.0-slice.6`.
- **dcca**: [`spec/dcca.json`](../spec/dcca.json), profile `dcca-rfc4006-single-event-v1`, slice `0.1.0-slice.1`.
- **roabort**: [`spec/roabort.json`](../spec/roabort.json), profile `ro-rfc6733-abort-slice-v1`, slice `0.1.0-slice.1`.
- **smsfwro**: [`spec/smsfwro.json`](../spec/smsfwro.json), profile `smsfw-ro-iec-template-v1`, slice `0.1.0-slice.6`.
- **legacy420**: [`spec/legacy420.json`](../spec/legacy420.json), profile `smsfw-ro-legacy420-v1`, slice `0.1.0-slice.6`.
- **combined**: [`spec/combined.json`](../spec/combined.json), profile `smsfw-ro-five-template-wire-v1`, slice `0.1.0-slice.5`.
- **scap1**: [`spec/scap1.json`](../spec/scap1.json), profile `smsfw-scap1-event-v1`, slice `0.1.0-slice.2`.
- **sgdalert**: [`compat/sgdalert/profile.go`](../compat/sgdalert/profile.go), profile `legacy-sgd-alert-v1`, slice `legacy-sgd-alert-v1.1`.

The [support matrix](diameterkit-support-matrix.md) lists application/command IDs
and limitations. Sources and rule identifiers in each JSON record the precise
normative references; legacy bindings cite their reviewed template evidence rather
than pretending to be another 3GPP version. Multiple incompatible 4/272 bindings
cannot share a bundle. The abort profile is independent 4/274; SCAP1 is 0/271 with
an explicit accounting capability convention.

## Generate and review

```sh
go generate ./...
python3 spec/tools/release_check.py --suite static
```

Generation is explicit. Eleven generated files must match their definitions and
generator output byte-for-byte, including AI provenance headers. CI checks rather
than repairs stale files, and exercises an intentionally stale negative control.
Generated public declarations participate in the reviewed API snapshot.

Generators v14–v17 add private decode-time omission, cardinality, value and length
metadata without requiring an internal-package import in external generated code.
Their DefinitionDigest changes
invalidate earlier Bundle fingerprints; see the [diagnostic and recovery migration
scope](missing-avp-diagnostics.md#compatibility-and-recovery). Definition fields,
wire grammars and profile slice versions are unchanged in these batches.

Batch 101 adds [bounded Redirect handling](redirect-handling.md) to the common E
grammar. Operations `ops9` invalidates earlier Bundle fingerprints independently
of generator v17, unchanged generated files and profile slice versions. The
common `ErrorAnswer` API grows three fields; use keyed literals.

Batch 102 [manual-group diagnostics](manual-group-diagnostics.md) uses generator
v18 to preserve handwritten evidence through generated parents. Eleven digests
and Bundle fingerprints change; public shapes, ops9 and profile slices stay
unchanged. Regenerate external bindings to gain the evidence bridge.

Change definitions, handwritten rules, independent wire vectors and contextual
negative tests together when expanding a slice. Review ownership, envelope/source
association, error namespace, role evidence and resource limits. Update the slice
version and public scope when semantics change; follow the independent
[module/API policy](version-install-policy.md). Header-only provenance changes do
not expand protocol support or change wire bytes.

The Sh payload helper covers selected Ref11/12/24/25 forms with separate byte/depth
bounds. A valid Diameter answer does not imply valid XML, subscriber permission or
business success. No definition grants automatic delivery, charging, retry,
persistence or target authorization.
