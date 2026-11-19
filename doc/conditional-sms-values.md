<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated SMS template value diagnostics and development release scope.
SPDX-AI-Date: 2026-11-19
-->

# SMS-Node and SM-Message-Type values

Batch 122 adds bounded 5004 diagnostics for two existing combined SMS template
constraints. IEC, INITIAL and TERMINATION require both SMS-Node and
SM-Message-Type to be zero. These rules already exist in this named compatibility
binding; this change adds diagnostics and makes no universal protocol claim.

| Values at this rule | Receive behavior |
| --- | --- |
| SMS-Node nonzero, SM-Message-Type zero | Quote the original SMS-Node |
| SMS-Node zero, SM-Message-Type nonzero | Quote the original SM-Message-Type |
| Both zero | Continue existing validation |
| Both nonzero | Retain unsupported failure without this diagnostic |

The selected Receive errors retain `ErrUnsupported` and add `ErrField` with
Vendor 10415 paths `873 / 2000 / 2016` or `873 / 2000 / 2007`.
Build and Send retain the single-key 2000 path and earlier classification.
Money templates retain their existing behavior; they do not use this proof.
Earlier sequence, subscription, template-marker and identity errors retain
precedence. Later optional SMS/MMS, timestamp and unit failures do not displace
this earlier rule. Structural decode failures retain their own diagnostics.

## Original instance and bounds

Enable `BaseErrorFailures` to receive a reply. `CloseOnFailure` is the default.
The private evidence contains only a static branch discriminator. It stores no
request bytes, subscriber values or persistent state. The consumer independently
checks the request kind, template markers, exact path, unique ancestors and both
leaf instances, flags, Integer32 widths and values. Each containing AVP area is
scanned completely with a shared source counter. Unrelated sibling payloads stay
opaque. Missing, duplicate, malformed or ambiguous selected/counterpart instances
cannot produce this quote; altered evidence cannot fall back to an untyped reply.

The owned quote is exactly **40 bytes**: two minimal 12-byte Grouped headers and
the original 16-byte leaf. It copies neither sibling data nor a replacement value.
Including Failed-AVP requires four AVPs and depth three. The minimal source
fixture separately requires seven AVPs, a 108-byte frame and a 44-byte largest
field. Source and output checks precede allocation. The complete answer envelope
and exact reply-byte limit are checked separately.

Tests cover signed extremes, the two-value/mode/template matrix, corruption at
every selected level, counterpart absence/duplication/width, shared budgets,
ordering, ownership and TCP/TLS ordinary/emergency/default handling. Invalid
requests are not delivered; held charges, follow-up and retirement are checked.
Six seeds extend the existing `FuzzInvalidQuote` target without adding a new
unbounded fuzz path.

## Reproduce

Use the pinned toolchain from the current checkout:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(SMSValue|RejectSMSValue)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(SMSValue|RejectSMSValue)' -count=20 -timeout=10m
python3 spec/tools/release_check.py --suite all
```

## Version and migration

Combined advances to `0.1.0-slice.16`; its definition digest and dependent Bundle
fingerprints change. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Keep original builds for earlier recovery records; never
relabel records or bypass their fingerprints. The library remains pure Go with
application-owned persistence and recovery authorization.

This development change is **not included in published v0.2.0-rc.1**. Its verified
[release receipt](releases/v0.2.0-rc.1.md) remains historical and immutable.
A future candidate requires a new version and its own installation acceptance.
Use commit `6897bebc71e6ebd291b0b15573567a17b56a8259` to reproduce the published
pair's original API/install checks; the current API baseline intentionally differs.

See [remaining gaps](remaining-gaps.md), [completion conditions](release-candidate-checklist.md)
and [batch evidence](../spec/evidence/batch-122-sms-values.json).
