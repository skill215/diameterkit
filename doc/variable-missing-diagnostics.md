# Selected variable-length missing-field diagnostics

Batch 105 adds explicit minimum payload evidence for eleven generated fields.
It extends bounded `5005 DIAMETER_MISSING_AVP` construction without changing
accepted wire values, public Go layouts, profile slice versions or operations
`ops9`. These changes are in the development checkout, not the published tags.

## Selected fields

| Definition | Field | Minimum payload bytes |
| --- | --- | ---: |
| Base | CER / CEA `ProductName` | 0 |
| S6c | `UserIdentifier.MSISDN`, `SRR.SCAddress`, `SRR.DestinationRealm` | 1 |
| SGd | OFR / TFR `SMRPUI` | 0 |
| SGd | `OFR.SCAddress` | 1 |
| combined | `CCR.ServiceContextID` | 1 |
| combined | `SubscriptionID.SubscriptionIDData` | 0 |
| DCCA | `SubscriptionID.SubscriptionIDData` | 1 |

Each field explicitly declares `diagnostic_minimum_bytes` in its JSON definition.
Zero is a declaration, distinct from an omitted property. The generator accepts
only required fields with `min: 1` and an exact existing primitive minimum:
`identity` and `tbcd` require one byte; `utf8` and `octets` allow zero unless their
existing `nonempty` constraint requires one. Unsupported types, optional fields,
minimum counts above one, negative widths and mismatched widths are rejected.

The declaration supplies diagnostic metadata. It does not impose a new validation
rule. A present empty field keeps its existing acceptance or rejection behavior;
it never becomes a missing-field error. Contextual domain minima are not inferred:
for example, DCCA's one-byte diagnostic does not claim that a one-digit IMSI is
valid. Likewise, a zero-filled Identity example is diagnostic content, not a
usable routing identity. This follows the missing-AVP example convention in
[RFC 6733 section 7.5](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.5).

## Evidence and limits

Generated private errors carry the field key, flags, exact minimum and a versioned
wire-type marker. Existing ancestor evidence preserves the precise parent instance
through nested or repeated groups. External generated packages need no import of
a core internal package. These callbacks are trusted schema code, not proof that
untrusted input is authentic.

The consumer validates the complete path and metadata before allocation. Both
ancestors and the Failed-AVP wrapper count toward depth and AVP limits. Payload,
field, frame and complete-reply budgets all apply. One-byte payloads get three
zero padding bytes, while AVP Length includes only the header and actual payload.
Quotes own their storage and contain no unrelated siblings or subscriber values.
A failed explicit proof cannot fall back to an older, shorter untyped placeholder.

TCP/TLS tests cover six request cases through ordinary and emergency resources,
including the second repeated Subscription-ID, exact returned bytes, budget
failure, held charges, invalid-request non-delivery, subsequent valid work and
retirement. Base CER/CEA and the remaining selected binding fields have codec
tests; Base coverage does not add a handshake-specific response policy.

Selection belongs to a field in a particular definition. OFR's SC-Address opt-in
does not enable TFR's same AVP. Bare errors and undeclared variable fields remain
unsupported. A variable child opt-in does not implicitly enable its whole parent.
[Batch 106](variable-group-diagnostics.md) separately adds typed whole-group
plans for three reviewed groups, with exact minimum and named-rule audit declarations. Conditional and minimum-count-above-one omissions also remain
separate work.

## Reproduce and migrate

```sh
CGO_ENABLED=0 go test . ./base ./app/s6c ./app/ro/dcca ./internal/missingavp ./internal/groupavp ./internal/excessavp ./internal/codegen -run 'Test(Variable|RejectVariable|Minimum|RejectMinimum|Missing|RejectDeepMissing|Group|RejectGroup|Manual|RejectManual|Length|RejectLength|Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./base ./app/s6c ./app/ro/dcca ./internal/missingavp ./internal/groupavp ./internal/excessavp ./internal/codegen -run 'Test(Variable|RejectVariable|Minimum|RejectMinimum|Missing|RejectDeepMissing|Group|RejectGroup|Manual|RejectManual|Length|RejectLength|Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' -count=20 -timeout=10m
CGO_ENABLED=0 go test ./internal/missingavp -run '^$' -fuzz '^FuzzVariableMinimum$' -fuzztime=30s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Generator v21 changes all eleven generated digests and Bundle fingerprints.
Regenerate external bindings; retain original builds for earlier recovery records.
The planned next core candidate remains `v0.2.0-rc.1`. Publication of a new immutable
core/adapter pair and cold-cache installation with `GOWORK=off` and no `replace`
remain separate gates. Pure Go and the no-persistence boundary are unchanged.
See the [receipt](../spec/evidence/batch-105-variable-omissions.json),
[remaining gaps](remaining-gaps.md) and [version policy](version-install-policy.md).
