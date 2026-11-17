# VSAI conditions and variable-length Proxy omissions

Batch 103 adds bounded diagnostics for selected handwritten group constraints.
The implementation is pure Go, retains no request bytes in error metadata, and
adds no persistence. Automatic answers still require `BaseErrorFailures`, a
complete valid return envelope, and sufficient request/reply reservations.

## Selected rules

| Decoded group | Condition | Result and Failed-AVP content |
| --- | --- | --- |
| VSAI | Both Auth-Application-Id and Acct-Application-Id | 5009; both received fields, in their original order with their original flags and values |
| VSAI | Neither application ID, with Vendor-Id present | 5005; Auth-Application-Id and Acct-Application-Id, each with a four-byte zero value |
| Proxy-Info | Missing Proxy-Host | 5005; a one-byte zero value for the selected Identity minimum, with zero padding |
| Proxy-Info | Valid Host but missing Proxy-State | 5005; a zero-length OctetString placeholder |

VSAI conditions follow [RFC 6733 section 6.11](https://www.rfc-editor.org/rfc/rfc6733.html#section-6.11).
The Proxy grammar and placeholder rules are specified in
[section 6.7.2](https://www.rfc-editor.org/rfc/rfc6733.html#section-6.7.2) and
[section 7.5](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.5).
The Host minimum derives from this release's ASCII DNS Identity validator; the
zero byte is diagnostic content, not a usable Identity. An empty State that was
actually present remains valid. If both Proxy fields are absent, the first
missing field is Host. If VSAI Vendor-Id is absent, its existing missing-field
diagnostic takes precedence over the conditional application-ID rule.

## Evidence and bounds

Generator v19 propagates versioned private constraint evidence through generated
parents and external packages. Metadata names a rule and the actual ancestor
keys, flags and offsets. The quote builder independently walks those instances,
checks the entire selected group, and verifies the rule again against the source.
Repeated parents cannot silently select an earlier valid sibling. Unknown
mandatory fields, duplicate known children, wrong integer widths, malformed
framing and resource exhaustion do not establish a conditional proof.

Diagnostics contain only the selected leaves and minimal ancestor headers.
They own their bytes. Input frame, field, count, depth and repetition limits,
output field/frame size, and the caller's byte budget are checked before output
allocation. The Failed-AVP wrapper and the complete answer must fit too. Bare
sentinels, mismatched paths and unsupported rules produce no guessed quote.
The convention trusts schema callbacks; it does not authenticate a peer.

VSAI conditional failures retain `errors.Is(err, validation.ErrField)` and also
identify the selected `ErrDuplicate` or `ErrMissing` condition. Build/send-time
validation still rejects invalid VSAI combinations. Public signatures and struct
layouts stay unchanged.

## Network boundary

VSAI conditional application requests are verified over TCP and TLS using normal
and emergency rejection slots. Tests check both field orders, omission examples,
envelope preservation, held resource charges, non-delivery of invalid work,
subsequent valid requests and resource retirement.

A damaged Proxy still cannot supply a safe return envelope. Its decoder can
provide a bounded omission diagnostic, but the automatic network path closes
without a reply. It never invents a return Host/State or drops an invalid Proxy.
Repeated CER VSAI and nested external-module checks establish codec/evidence
behavior, not handshake-specific answer mappings. Profiles deliberately treating
VSAI as opaque retain that behavior.

## Reproduce and migrate

```sh
CGO_ENABLED=0 go test . ./base ./internal/groupavp ./internal/excessavp ./internal/codegen -run 'Test(Group|RejectGroup|Manual|RejectManual|Length|RejectLength|Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' -count=20 -timeout=5m
CGO_ENABLED=1 go test -race . ./base ./internal/groupavp ./internal/excessavp ./internal/codegen -run 'Test(Group|RejectGroup|Manual|RejectManual|Length|RejectLength|Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' -count=20 -timeout=5m
CGO_ENABLED=0 go test ./internal/groupavp -run '^$' -fuzz '^FuzzGroupQuote$' -fuzztime=30s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

The release suite has twelve bounded fuzz targets. Generator v19 changes all
eleven DefinitionDigest constants and Bundle fingerprints; protocol JSON inputs,
profile slice versions and common operations `ops9` remain unchanged. Regenerate
and review external bindings. Settle earlier recovery records with their original
build; never relabel fingerprints. The planned core remains `v0.2.0-rc.1`, and
this batch publishes no immutable version. See the [version policy](version-install-policy.md),
[batch receipt](../spec/evidence/batch-103-group-constraints.json) and
[remaining gaps](remaining-gaps.md). [Batch 104](minimum-group-diagnostics.md) adds selected whole-group minima.
Other domain/handshake diagnostics, malformed-header length failures, unreviewed
whole-group omissions and cardinality minima greater than one remain separate.
