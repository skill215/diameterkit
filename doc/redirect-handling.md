# Bounded Redirect 3006 handling

The unreleased batch 101 checkout can build, receive and snapshot selected
`pdu.ErrorAnswer` messages with Result-Code 3006. The common E grammar carries
redirect information through the existing codec, `Inbound.Respond` and completion
APIs. It does not resolve names, dial a URI, alter a route, replay a request or
maintain a redirect cache. Applications own those decisions and authorization.

## Fields and URI subset

| Field | Selected contract |
| --- | --- |
| `RedirectHost []string` | Required and nonempty for 3006; at most `base.MaxRedirectHosts` (256), or the lower configured repetition limit. Order and spelling are preserved. |
| `RedirectHostUsage avp.Optional[uint32]` | Optional singleton, values 0 through 6. Absent means no caching indication. |
| `RedirectMaxCacheTime avp.Optional[uint32]` | Optional singleton; required when usage is nonzero. Zero is valid. Presence with absent/zero usage grants no caching permission. |

These three fields are rejected with other result codes in the selected E
grammar. Duplicate singleton AVPs, wrong integer widths and unknown mandatory
AVPs fail validation. Source authorization, Session/Proxy association and command
header checks still apply. Malformed matched answers produce `ErrInvalidAnswer`
without a usable body or automatic alternate attempt.

`avp.DiameterURI` accepts `aaa://` and `aaas://`, the existing ASCII DNS-name
`Identity` subset (including single labels), an optional port 1–65535 with at
most five digits, optional `;transport=tcp` or `;transport=sctp`, then optional
`;protocol=diameter`. Protocol alone is allowed. Tokens are ASCII
case-insensitive. No normalization or default expansion occurs. For example:

```text
aaa://redirect.example
aaas://redirect.example:5658;transport=sctp;protocol=diameter
```

IP literals, trailing root dots, Unicode names, escapes, userinfo, paths, query
strings, fragments, UDP, other AAA protocols and unknown, reordered or repeated
parameters are outside this subset. The length is capped by
`avp.MaxDiameterURIBytes` before parsing. This is a deliberately narrower subset
of [RFC 6733 section 4.3](https://www.rfc-editor.org/rfc/rfc6733.html#section-4.3);
the field dependencies follow [sections 6.12–6.14](https://www.rfc-editor.org/rfc/rfc6733.html#section-6.12).

## Public use and ownership

Pass a keyed prototype to `Inbound.Respond`, supplying the usual response context:

```go
answer := &pdu.ErrorAnswer{
    ResultCode: 3006,
    RedirectHost: []string{"aaas://redirect.example;transport=tcp"},
    RedirectHostUsage: avp.Some(uint32(0)),
}
```

Use `codec.CompleteAnswer` when working without a Node; it supplies the request
envelope and origin before encoding. Accepted response/snapshot values own their
host strings and slice. Absent optional values are canonicalized to zero.
An invalid or oversized response does not commit the inbound operation; a valid
response can still follow. Successful duplicates reuse the frozen answer and
each arrival's envelope. Redirect cache hints do not change the Node's configured
volatile duplicate-cache deadline, slot count or byte accounting.

The fixed host cap, configured field/count/repetition/frame limits, full response
grammar and runtime reply reservations all apply. A valid URI does not grant a
new peer identity, cross-target recovery permission or a durable deduplication
guarantee. Pure Go and the no-persistence boundary are unchanged.

## Verification and compatibility

Independent raw frames check all usage values, TTL boundaries, required fields,
duplicate/wrong-width fields, envelope/source failures, ownership and exact limits.
TCP/TLS tests cover both route policies, terminal/agent answers, combined and
Sh/S6c/SGd cross-peer duplicate caches, failed response rollback and resource
retirement. These are local peer tests, not real-target redirect acceptance.

```sh
CGO_ENABLED=0 go test . ./avp ./codec ./schema -run 'Test(Redirect|DiameterURI|ProductionImport)' -count=20 -timeout=5m
CGO_ENABLED=1 go test -race . ./avp ./codec ./schema -run 'Test(Redirect|DiameterURI|ProductionImport)' -count=20 -timeout=5m
CGO_ENABLED=0 go test ./avp -run '^$' -fuzz '^FuzzDiameterURI$' -fuzztime=1000000x -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

This adds exported declarations and fields to `ErrorAnswer`; use keyed struct
literals. The shared error grammar changes Bundle operations from `ops8` to
`ops9`, invalidating earlier Bundle fingerprints. Generator v17, all eleven
generated files, DefinitionDigest constants and profile slice versions remain
unchanged in this batch. Settle old recovery records using their original build;
never relabel a fingerprint. The planned core candidate remains `v0.2.0-rc.1`.
No new immutable version is published here. See the [version policy](version-install-policy.md),
[batch receipt](../spec/evidence/batch-101-redirect.json) and [remaining gaps](remaining-gaps.md).
