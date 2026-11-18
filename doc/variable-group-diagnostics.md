# Missing groups with variable-length children

Batch 106 extends bounded `5005 DIAMETER_MISSING_AVP` examples to three explicitly
reviewed groups. Their required children can mix fixed-width values with zero- or
one-byte variable payloads. Wire validation, public Go layouts and operations
`ops9` are unchanged. This describes the development checkout, not published tags.

## Selected scope

| Definition and missing group | Required diagnostic children | Variable minimum |
| --- | --- | --- |
| DCCA CCR `SubscriptionID` (443) | Type (450), Data (444) | One UTF8 byte, using the existing `nonempty` declaration |
| SCAP1 ACR `SubscriptionID` (193/553) | Type (193/555), Data (193/554) | Zero octets |
| combined MMS `OriginatorAddress` (10415/886) | Type (10415/899), Data (10415/897) | Zero octets |

Each group opts in with `diagnostic_minimum: true`. Every required variable child
must also declare the exact `diagnostic_minimum_bytes` already permitted by its
primitive and `nonempty` constraint. An omitted declaration is never inferred from
the AVP code, a sibling binding or the Go field's zero value.

A variable-bearing group with a handwritten `rule` must additionally declare
`diagnostic_minimum_rule` equal to that rule's identifier. This requirement applies
to each reviewed group above a variable descendant in the minimum tree. Missing,
stale or misplaced audit declarations are rejected. The marker records an author's
review; it cannot prove what a Go function does. Authors must recheck rule bodies
and enclosing command constraints when changing a definition. Rules with extra
conditional presence or required multiplicity must not opt in.

The compiler still rejects undeclared or mismatched variable minima, unsupported
types, missing child-group opt-ins, empty group minima, captured Mandatory flags,
required counts above one, cycles and plans exceeding 64 nodes or 16 levels.
Optional children are excluded. Existing fixed-only opt-ins retain their earlier
review contract.

## Evidence and encoding

Generated private errors now use `DiameterMissingGroupV2`. Each immutable preorder
record identifies a group or an exact primitive kind, so a zero-byte UTF8/octet
leaf cannot be mistaken for a group. Runtime validation checks the whole plan,
keys, flags, duplicate sibling keys, ancestor path, child counts, types and widths
before allocating output. Ancestors and the Failed-AVP wrapper consume depth and
AVP limits; field, frame and complete-reply budgets remain enforced.

For a one-byte child, AVP Length includes its header and one payload byte. Three
zero padding bytes follow it; the next sibling starts after that padding. Group
lengths include their padded children. All example payloads and padding are zero,
and the returned quote owns its storage. It contains only the minimum subtree and
its ancestors, with no unrelated sibling or subscriber data.

The consumer still accepts valid fixed-only V1 evidence. V1 rejects typed records;
mixed-version chains are rejected, and malformed V2 evidence never falls back to
V1 or an older untyped placeholder. External generated packages need no internal
core import. As with other schema callbacks, evidence comes from trusted code.

These are diagnostic examples under
[RFC 6733 section 7.5](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.5),
not valid application objects. Zero-filled Type/Data does not choose a subscription
domain, invent an identity or satisfy business rules. Present empty groups still
produce their actual missing-child error; present values retain existing checks.

## Conditions that remain separate

- Batches 109 and 110 separately added [combined subscription and SCAP1 parameter
  shortages](cardinality-diagnostics.md) through reviewed cardinality evidence.
- Batch 111 separately adds [isolated conditional Recipient-Address omissions](conditional-recipient-diagnostics.md)
  for combined SMS templates. Other conditional fields, simultaneous conditional
  omissions and absent MMS/SMS groups remain outside these selected slices.
- Unreviewed groups, unsafe envelopes and the default close policy remain closed.

TCP/TLS tests cover all three selected cases through ordinary and emergency slots,
exact returned bytes, present-group versus whole-group absence, held resource
charges, invalid-request non-delivery, subsequent valid work and retirement.
Negative network cases exercise the boundaries above. Synthetic external bindings
cover six variable kinds, nested mixed trees, optional-child exclusion and the
second repeated parent. Bounded fuzz checks metadata, framing, padding and limits.

## Reproduce and migrate

```sh
CGO_ENABLED=0 go test . ./base ./app/s6c ./app/ro/dcca ./internal/missingavp ./internal/groupavp ./internal/excessavp ./internal/codegen -run 'Test(Variable|RejectVariable|Minimum|RejectMinimum|Missing|RejectDeepMissing|Group|RejectGroup|Manual|RejectManual|Length|RejectLength|Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' -count=20 -timeout=15m
CGO_ENABLED=1 go test -race . ./base ./app/s6c ./app/ro/dcca ./internal/missingavp ./internal/groupavp ./internal/excessavp ./internal/codegen -run 'Test(Variable|RejectVariable|Minimum|RejectMinimum|Missing|RejectDeepMissing|Group|RejectGroup|Manual|RejectManual|Length|RejectLength|Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' -count=20 -timeout=15m
CGO_ENABLED=0 go test ./internal/missingavp -run '^$' -fuzz '^FuzzMinimumVariableTree$' -fuzztime=1000000x -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Generator v22 changes all eleven definition digests and Bundle fingerprints.
Regenerate external bindings and keep original builds for earlier recovery records.
Published core `v0.1.0-rc.2` and adapter `v0.1.0-rc.1` remain immutable. Publishing
and cold-installing the next pair with `GOWORK=off` and no `replace` are separate
gates. The library remains pure Go with no persistence boundary.

See the [receipt](../spec/evidence/batch-106-variable-groups.json),
[remaining gaps](remaining-gaps.md) and [version policy](version-install-policy.md).
