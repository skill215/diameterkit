# Minimum Grouped examples for missing AVPs

Batch 104 extends bounded 5005 diagnostics to whole missing Grouped AVPs whose
minimum child structure has been explicitly reviewed. It keeps pure Go and the
no-persistence boundary. `BaseErrorFailures` and a valid return envelope remain
prerequisites for automatic network replies.

## Selected scope

The generator accepts `diagnostic_minimum: true` on a group. This declaration
asserts that its required fields describe the complete minimum structure; it
does not infer conditional requirements from handwritten validation rules.

| Binding | Emitted missing-group examples |
| --- | --- |
| Generic DCCA | UnitValue inside Money or CostInformation; required ValueDigits only |
| legacy420 and combined | UnitValue inside Money; ValueDigits and Exponent |
| SCAP1 | VendorApplication, RequestedUnit, and its nested UnitValue |

Money and DCCA CostInformation also declare their reviewed minimum structures.
An optional field's absence never triggers this mapping: generated omission
evidence is emitted only where the field has `min: 1` and no instance was seen.
Declarations do not create new mandatory fields or change accepted wire grammar.

Scalar diagnostic values are zero-filled at their declared four/eight-byte
widths. Group and child headers remain structured so the existing bounded
diagnostic walker can inspect them. These are diagnostic examples, not valid
business values or invented request data. The selected construction follows the
missing-AVP minimum-length and hierarchy guidance in
[RFC 6733 section 7.5](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.5).

## Bounds and exclusions

The compiled plan is immutable schema metadata, with at most 64 nodes and 16
levels. External generated packages use a versioned private error convention,
without importing core internal packages. Errors retain no request bytes.

Before allocating output, the consumer freezes ancestor metadata and checks the
plan, field path, keys, flags, scalar widths, child counts, unique sibling keys,
depth, total AVP count and byte budgets. Ancestors and the Failed-AVP wrapper
count toward runtime limits. Quotes own their bytes and omit existing siblings.
The complete answer must also fit encoder and runtime reservations.

The generator rejects opted-in groups with empty minima, unreviewed child groups,
captured Mandatory flags, undeclared variable-length required children or required counts
greater than one. Optional children are excluded. Custom rules that impose extra
presence conditions must not opt in. Bare sentinels and unreviewed groups do not
gain inferred diagnostics. RequestMSCC and unsafe Proxy envelopes retain their
existing close boundary.

TCP/TLS acceptance covers combined UnitValue and SCAP1 groups through ordinary
and emergency rejection slots. It checks complete quote bytes, original return
envelopes, held charges, no invalid delivery, later valid work and retirement.
Synthetic external-module tests cover a missing parent with multiple nested
children. Negative controls and bounded fuzz exercise malformed plans and limits.

## Reproduce and migrate

```sh
CGO_ENABLED=0 go test . ./base ./internal/missingavp ./internal/groupavp ./internal/excessavp ./internal/codegen -run 'Test(Minimum|RejectMinimum|Missing|RejectDeepMissing|Group|RejectGroup|Manual|RejectManual|Length|RejectLength|Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./base ./internal/missingavp ./internal/groupavp ./internal/excessavp ./internal/codegen -run 'Test(Minimum|RejectMinimum|Missing|RejectDeepMissing|Group|RejectGroup|Manual|RejectManual|Length|RejectLength|Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' -count=20 -timeout=10m
CGO_ENABLED=0 go test ./internal/missingavp -run '^$' -fuzz '^FuzzMinimumTree$' -fuzztime=1000000x -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

The release suite now has thirteen bounded fuzz targets. Generator v20 changes
all eleven DefinitionDigest constants and Bundle fingerprints. Four definition
files add diagnostic opt-ins; their existing fields, rules and slice versions
are unchanged. Public Go signatures/layouts and common operations `ops9` remain
unchanged. Regenerate external bindings and settle earlier recovery records with
their original build. No immutable version is published by this batch.

See [evidence](../spec/evidence/batch-104-minimum-groups.json),
[version policy](version-install-policy.md) and [remaining gaps](remaining-gaps.md).

Batch 105 adds separately selected [variable-field minimum diagnostics](variable-missing-diagnostics.md).
Field opt-ins alone do not widen the whole-group grammar. See that guide
for the current generator v21 fingerprint and migration boundary.

[Batch 106](variable-group-diagnostics.md) adds separately reviewed V2 whole-group
plans with variable children. Fixed-only V1 consumption remains supported;
unreviewed conditional and cardinality rules remain outside this mapping.
