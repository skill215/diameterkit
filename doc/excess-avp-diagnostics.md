# First-excess AVP diagnostics

This guide applies to the unreleased batch 97 development checkout. Production
code remains pure Go and has no persistence boundary. Published core
`v0.1.0-rc.2` does not contain this extension.

## Behavior

With `BaseErrorFailures` enabled for an inbound application command, a generated
decoder can now produce an E answer with Result-Code 5009 when a known AVP exceeds
its definition's finite maximum. Failed-AVP contains the **first excess instance**,
including its original header, data and padding. Earlier permitted instances and
later excess instances are not substituted. This follows
[RFC 6733 sections 7.1.5 and 7.5](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.1.5).

The decoder records relative offsets, keys, flags and the compiled maximum in
private error metadata. Known Grouped parents extend this exact instance path.
The rejection builder walks the original request at those offsets and verifies
that the leaf is occurrence `maximum + 1` in its direct parent. It copies the
offending instance inside minimal ancestor headers. It neither searches another
parent for the same key nor copies unrelated parent contents.

Quoted data is diagnostic evidence: it is not decoded again as an ordinary
application value. A duplicate Grouped AVP with malformed internal bytes can be
quoted when its outer frame is trustworthy; the existing diagnostic scanner
reports incomplete interpretation. Known diagnostic descendants still consume
the full response's depth and AVP-count budget.

## Safety and scope

| Condition | Behavior |
| --- | --- |
| Generated finite-maximum violation, trusted envelope, sufficient bounds | E/5009 with one first-excess diagnostic hierarchy |
| Ordinary inbound slots exhausted, emergency rejection reserve available | Same finite decode check before the existing overload decision |
| `CloseOnFailure` policy | Existing close behavior |
| Local `MaxRepeated`, depth, AVP-count or byte limit exceeded | Resource failure; no conversion into a protocol cardinality error |
| Duplicate/untrusted Session-Id, invalid Proxy-Info or illegal ancestor flags | Close without inventing or repairing the envelope |
| Bare `FieldError`, manual cardinality rule or unproven offset/path | Close; the new mapping requires generated instance evidence |

At most 64 path nodes are processed. Original frame/field/walker limits bound
inspection; quote size is checked before allocation. The complete response must
also pass the existing selected E grammar, diagnostic traversal, response size
and reservation checks. If any check fails, no partial diagnostic is sent.

The error stores no payload or request-buffer reference. The quoted result owns
its copied bytes. Generated extensions remain usable outside the module: private
metadata types are emitted with public dependencies only. Schema callbacks remain
trusted code; their metadata is not remote-peer authentication.

Manual Proxy/VSAI cardinality and handshake-specific 5009 handling remain outside
this slice. The later [batch 98 extension](invalid-avp-diagnostics.md) adds generated
5001/5004. The [batch 100 extension](length-avp-diagnostics.md) adds selected typed-length 5014.
[Batch 101](redirect-handling.md) adds bounded advisory Redirect 3006.
Manual/domain/malformed-header mappings remain separate, as do the
unsupported omission kinds listed in the [missing-AVP guide](missing-avp-diagnostics.md).
This extension does not grant a new business retry or recovery permission.

## Compatibility

Generator v15 changes generated DefinitionDigest constants and Bundle
fingerprints. Public signatures and struct layouts are unchanged. Earlier recovery
records must be drained or settled using their original reviewed build; do not
rewrite fingerprints to pass import checks. The planned next core candidate
remains `v0.2.0-rc.1`, pending separate publication and cold installation acceptance.
This batch does not change published tags or the adapter's dependency pin.

## Acceptance

Tests distinguish three different instances, retain binary padding, locate a
child within the second repeated parent, and exercise four nested money groups.
TCP and TLS cover ordinary and emergency rejection, no application delivery,
resource retirement, envelope refusal, local repetition limits and reply budgets.
Independent metadata tests cover maximum greater than one, offset/path changes,
truncation, ownership and bounds. The new quote fuzzer is the seventh CI target.

```sh
CGO_ENABLED=0 go test . ./internal/excessavp ./internal/codegen \
  -run 'Test(Excess|RejectExcess|ExternalGenerated|Generated|ProductionImport)' \
  -count=20 -timeout=5m
CGO_ENABLED=1 go test -race . ./internal/excessavp ./internal/codegen \
  -run 'Test(Excess|RejectExcess|ExternalGenerated|Generated|ProductionImport)' \
  -count=20 -timeout=5m
python3 spec/tools/release_check.py --suite all
```

See [batch 97 evidence](../spec/evidence/batch-97-excess-diagnostics.json) and the
[current remaining gaps](remaining-gaps.md). Local loopback tests do not establish
real target interoperability.
