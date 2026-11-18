# Typed AVP length diagnostics

The unreleased batch 100 checkout adds bounded 5014 diagnostics to generated
decoders. A known inbound application command must opt into `BaseErrorFailures`.
`CloseOnFailure` remains the default. The implementation is pure Go and adds no
persistence, application retry or durable recovery state.

## Supported scope

| Generated wire type | Length failure mapped to 5014 |
| --- | --- |
| Unsigned32, Integer32 and Time | Payload is not exactly four bytes |
| Unsigned64 and Integer64 | Payload is not exactly eight bytes |
| Address | Fewer than two family bytes; IPv4 payload not six bytes; IPv6 payload not eighteen bytes, including family bytes |
| Current TBCD type | Empty payload or more than eight bytes |

[RFC 6733 section 7.1.5](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.1.5)
requires Failed-AVP for an invalid AVP length. This implementation selects fully
framed AVPs whose payload length contradicts the generated type. It does not
implement the separate malformed-header reconstruction cases described there.

TBCD nibble/digit failures within its byte bounds keep their existing 5004
mapping. Address family zero, empty opaque-family data and manual application
family restrictions do not acquire a length mapping. Other Address families
remain opaque; no IP width is inferred for them. The TBCD limit is the current
library type contract, not a claim about every protocol field named TBCD.

## Evidence and bounds

Generated private error metadata carries the actual key, flags, relative offset,
ancestor chain and a fixed type name. The existing `avp.ErrValue` sentinel and
`validation.FieldError` path remain available. No request buffer or subscriber
payload is retained in the metadata. External generated packages continue to use
public imports only.

The instance walker requires both invalid-instance and typed-length evidence,
locates the exact instance and checks the named length rule again against the
original bytes. A bare sentinel, unknown type name, correct-length value or
mismatched path/offset/flags cannot authorize a 5014 quote. The owned quote keeps
the leaf's original header, data and padding, plus minimal ancestor headers. A
later field or repeated parent cannot replace the offending instance.

Frame, field, count, depth and reply-byte bounds apply before allocation. The
Failed-AVP wrapper and full response grammar consume their ordinary limits.
Normal inbound handling and the emergency reserve share this path. Trusted
Session-Id and per-request Proxy-Info are preserved; malformed envelopes, framing,
illegal ancestor flags, resource exhaustion and insufficient reply space close
without a partial answer or application delivery. A typed-invalid but fully
framed quote is not marked `Incomplete` merely because its scalar width is wrong.

## Acceptance and reproduction

TCP/TLS tests exercise twenty typed failures per transport, including deep money
fields, a second Supported-Features parent, vendor headers, nonzero padding and
Address/TBCD boundaries. Two emergency cases and fifteen refusal cases per
transport cover preserved held reservations, no unsafe reply, subsequent valid
work and resource retirement. Independent evidence tests check lengths 0–20 and
invalid provenance. A generated external consumer checks all seven selected type
names and preserves opaque Address-family behavior. A ninth bounded fuzz target
checks exact-instance copying and output limits.

```sh
CGO_ENABLED=0 go test . ./internal/excessavp ./internal/codegen \
  -run 'Test(Length|RejectLength|Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' \
  -count=20 -timeout=5m
CGO_ENABLED=1 go test -race . ./internal/excessavp ./internal/codegen \
  -run 'Test(Length|RejectLength|Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' \
  -count=20 -timeout=5m
CGO_ENABLED=0 go test ./internal/excessavp -run '^$' \
  -fuzz '^FuzzLengthQuote$' -fuzztime=1000000x -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Generator v17 changes the eleven DefinitionDigest constants and Bundle
fingerprints. Public signatures and struct layouts remain unchanged. Existing
recovery records require their original reviewed build; never relabel a stored
fingerprint. Published core `v0.1.0-rc.2` and adapter `v0.1.0-rc.1` are unchanged.
The planned core `v0.2.0-rc.1` publication and immutable cold-install gate remain
separate from this checkout acceptance.

See the [batch 100 receipt](../spec/evidence/batch-100-length-diagnostics.json),
[earlier 5004/5001 scope](invalid-avp-diagnostics.md) and
[remaining gaps](remaining-gaps.md). Manual/domain, Proxy/VSAI internals,
handshake-specific mappings, malformed-header 5014 and remaining
omission diagnostics are separate work. These tests establish no real-target
interoperability or application recovery authorization.

[Batch 101](redirect-handling.md) subsequently adds bounded advisory Redirect 3006
and advances common operations to `ops9`; generator v17 remains unchanged.

[Batch 102](manual-group-diagnostics.md) extends selected mappings into handwritten
Proxy/VSAI decoders and preserves exact-instance evidence across generated
parents. [Batch 103](group-constraint-diagnostics.md) adds conditional VSAI pairs
and bounded Proxy omission minima. Unsafe Proxy envelopes still close; other
conditional and variable omissions remain separate. Generator v19 changes
generated digests and Bundle fingerprints.
