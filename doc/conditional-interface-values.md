<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated IEC Originator-Interface value diagnostics and acceptance.
SPDX-AI-Date: 2026-11-18
-->

# IEC Originator-Interface values

Batch 119 adds a bounded 5004 reply for one existing combined IEC Receive
failure: Originator-SCCP-Address is present, but Originator-Interface's
Interface-Type is not 1. Enable the existing `BaseErrorFailures` policy to send
this diagnostic. The default `CloseOnFailure` policy still closes the connection.

| Input at the IEC condition | Receive behavior |
| --- | --- |
| Address present, type 1 | Continue existing checks |
| Address absent, type 1 | Existing batch 114 conditional 5005 |
| Address present, type other than 1 | Selected batch 119 exact-instance 5004 |
| Address absent, type other than 1 | Unsupported; no new diagnostic |

The selected failure retains `ErrUnsupported` and adds `ErrField`.
Its Receive `FieldError.Path` now identifies Vendor 10415
`873 / 2000 / 2009 / 2006` (Service-Information / SMS-Information /
Originator-Interface / Interface-Type). Build and Send retain their previous
unsupported classification and single-key 2008 path. INITIAL, TERMINATION,
money templates, other interface fields and mixed conditions are unchanged.

Earlier sequence, subscription, SMS/MMS, timestamp and Requested-Action failures
keep their precedence. A later unit failure does not replace this earlier
selected error. Typed length, duplicate instance and malformed framing errors
retain their existing classifications.

## Bounded original-instance quote

A private, static named proof carries no request, address or interface value.
The consumer rechecks the precise path, EVENT/action/template selectors, unique
ancestors and leaf, flags, signed Integer32 width/value, address presence and
complete framing of each containing AVP area. Child decoders share the source
AVP counter. Unrelated sibling payloads remain opaque.

The quote contains three minimal 12-byte ancestor headers and the exact
original 16-byte Interface-Type AVP: **52 bytes**. It never substitutes 1 or
copies the address or siblings. Negative and extreme signed values are retained
as their original wire bytes. All source and quote checks precede allocation.

Including Failed-AVP requires five AVPs and depth four; source scanning has its
own count requirement (nine in the minimal consumer fixture). Byte, field,
frame, count and depth limits apply, and the full reply envelope must also fit.
A 51-byte quote budget is rejected. Altered paths cannot fall back to an
untyped example. This proof is a trusted-schema convention, not authentication
of arbitrary errors.

TCP/TLS tests cover ordinary and emergency admission, exact Session/Proxy
envelopes, owned output, no invalid application delivery, preservation of held
resources, valid follow-up and cleanup. The existing invalid-value fuzzer also
checks the selected original instance and bounded output.

## Reproduce

Run from the core module using the pinned toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(InterfaceValue|RejectInterfaceValue|ActionValue|RejectActionValue|OriginatorMissing|RejectOriginatorMissing)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(InterfaceValue|RejectInterfaceValue|ActionValue|RejectActionValue|OriginatorMissing|RejectOriginatorMissing)' -count=20 -timeout=10m
CGO_ENABLED=0 GOMAXPROCS=2 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Combined advances to `0.1.0-slice.14`; the definition digest and dependent Bundle
fingerprints change. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Keep original builds for earlier recovery records. Published
tags stay immutable; new publication and cold installation remain separate.
The library remains pure Go with no persistence.

See [remaining gaps](remaining-gaps.md), [Originator omission](conditional-originator-diagnostics.md),
[EVENT action values](conditional-action-values.md) and
[the batch receipt](../spec/evidence/batch-119-interface-values.json).
