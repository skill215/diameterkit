# Bounded handwritten Proxy/VSAI diagnostics

The unreleased batch 102 checkout extends decode-time evidence into the common
handwritten `Proxy-Info` and `Vendor-Specific-Application-Id` (VSAI) decoders.
Generated parents preserve the actual instance offsets across package boundaries.
The runtime still uses the existing opt-in `BaseErrorFailures` policy and bounded
Failed-AVP construction. Pure Go and the no-persistence boundary are unchanged.

## Selected mappings

| Error | Selected evidence | Result |
| --- | --- | --- |
| Invalid Proxy-Host | The actual failing Identity value | 5004 |
| Unknown mandatory child in Proxy/VSAI | Exact child key, flags and offset | 5001 |
| Repeated Proxy-Host or Proxy-State | First instance beyond the singleton maximum | 5009 |
| Repeated Vendor-Id, Auth-Application-Id or Acct-Application-Id | First instance beyond the singleton maximum | 5009 |
| Wrong width for a VSAI integer | Exact instance plus independently checked Unsigned32 width | 5014 |
| Missing VSAI Vendor-Id | Compiled four-byte minimum and ancestor keys/flags | 5005 |

Present-field diagnostics retain the offending leaf's original header, payload
and padding, with minimal ancestor headers. Earlier valid groups, earlier valid
leaves and later invalid duplicates cannot replace the selected instance.
Omission diagnostics contain a zero-filled placeholder; they do not claim to
copy an instance that was never present. The rules follow the selected Base
grammar and [RFC 6733 Failed-AVP semantics](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.5).

Each known singleton checks the configured repetition ceiling before reporting
its schema maximum. Complete AVP framing is required. Resource errors, malformed
headers and bare sentinels never become guessed protocol diagnostics. The
bounded walker checks paths, offsets, keys, flags, type width, count, depth and
reply limits before copying bytes. The Failed-AVP wrapper and the complete answer
must also fit their normal limits and runtime reservations.

## Safe response boundary

VSAI errors in an application request can produce an E answer when its full
Session/Proxy envelope is valid. TCP/TLS tests exercise ordinary and emergency
rejection, preserve held charges, prevent invalid request delivery, accept later
valid work and retire resources after Close/Wait.

A malformed Proxy is different: its required return envelope cannot be safely
echoed. The decoder provides exact diagnostic evidence, but automatic network
rejection still closes without a reply. It never drops a Proxy, invents a host
or state, or copies only the valid siblings to force an answer through.

The following remain outside this selected mapping:

- Other whole-group, conditional and variable-length omissions beyond the
  [batch 103 VSAI/Proxy rules](group-constraint-diagnostics.md).
- Application-specific identity/domain constraints and handshake-specific answer
  mappings. Repeated CER VSAI tests provide offline codec evidence only.
- Profiles that deliberately treat VSAI as opaque, including the selected SMS
  ignore rule; this extension does not reinterpret those bytes.

An empty binary Proxy-State remains valid. Optional unknown extensions retain
their existing opaque behavior. No new authentication, authorization, routing,
recovery or durable deduplication claim is introduced.

## Verification and migration

Tests cover the second repeated Proxy/VSAI parent, exact duplicate instances and
padding, all three VSAI integer keys, source-byte ownership, tampered paths and
offsets, resource ceilings, the original conditional exclusions and TCP/TLS close boundaries. Batch 103
adds explicit conditional and Proxy-omission evidence as described in its guide.
A generated external module checks evidence through two repeated ancestor levels
using only public imports. `FuzzManualDiagnostics` exercises real handwritten
decoders and bounded copying; batch 102 added the eleventh target. Batch 103 adds a twelfth constraint target.

```sh
CGO_ENABLED=0 go test . ./base ./internal/excessavp ./internal/codegen -run 'Test(Manual|RejectManual|Length|RejectLength|Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' -count=20 -timeout=5m
CGO_ENABLED=1 go test -race . ./base ./internal/excessavp ./internal/codegen -run 'Test(Manual|RejectManual|Length|RejectLength|Invalid|RejectInvalid|ExternalGenerated|Generated|ProductionImport)' -count=20 -timeout=5m
CGO_ENABLED=0 go test ./base -run '^$' -fuzz '^FuzzManualDiagnostics$' -fuzztime=30s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Generator v18 changes the eleven generated DefinitionDigest constants and Bundle
fingerprints. Public signatures and struct layouts, common operations `ops9`,
protocol JSON inputs and profile slice versions stay unchanged. Regenerate and
review external bindings to obtain the cross-package evidence bridge. Settle
earlier recovery records with their original build; never relabel fingerprints.
The planned core candidate remains `v0.2.0-rc.1`; this batch publishes no tag.
See the [version policy](version-install-policy.md), [batch receipt](../spec/evidence/batch-102-manual-diagnostics.json)
and [current remaining gaps](remaining-gaps.md).
