<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated INITIAL/TERMINATION interface diagnostics and acceptance.
SPDX-AI-Date: 2026-11-18
-->

# INITIAL and TERMINATION interface values

Batch 120 adds bounded 5004 diagnostics for two isolated combined ECUR Receive
failures. Originator-Interface must have Interface-Type 3; Destination-Interface
must have Interface-Type 2. These are existing template constraints.

| Conditions at this rule | Receive behavior |
| --- | --- |
| No Requested-Action or Originator-SCCP; destination type 2; originator type other than 3 | Quote the original originator type with 5004 |
| No Requested-Action or Originator-SCCP; originator type 3; destination type other than 2 | Quote the original destination type with 5004 |
| Both interface types correct and forbidden fields absent | Continue existing checks |
| Both interface types wrong, or either forbidden field present | Retain unsupported failure; no new diagnostic |

Enable `BaseErrorFailures` for these replies. `CloseOnFailure` remains the
default. The two selected Receive errors retain `ErrUnsupported` and add
`ErrField`. Their Vendor 10415 paths are:

- Originator: `873 / 2000 / 2009 / 2006`.
- Destination: `873 / 2000 / 2002 / 2006`.

Build and Send retain the earlier unsupported classification and single-key
2008 path. The earlier Requested-Action failure retains its own 436 path.
IEC, money-template rules, mixed failures and structural checks keep their
boundaries. Earlier sequence, subscription, SMS/MMS and timestamp errors retain
precedence; later service-unit failures do not replace these selected errors.

## Original-instance and resource checks

Private named evidence contains only a static branch discriminator. It retains
no request, address or interface value. The consumer verifies request type,
template markers, absence of both forbidden fields, unique ancestors and both
interface instances, flags and Integer32 widths. It also verifies that the
counterpart is valid and the selected value is invalid.

Both interface child decoders share the source AVP counter. A budget that
covers the selected branch but excludes the counterpart fails. Each containing
AVP area is scanned completely; unrelated sibling payloads remain opaque.
Missing, duplicate, malformed or ambiguous selected/counterpart instances cannot
produce this quote. Altered paths cannot fall back to an untyped example.

The owned diagnostic is **52 bytes**: three minimal 12-byte ancestor headers and
the exact original 16-byte leaf. It copies no sibling and substitutes no value.
Including Failed-AVP requires five AVPs and depth four. Source scanning has
separate limits: the minimal consumer fixture uses nine AVPs, a 132-byte frame
and a 68-byte largest field. All source and quote checks precede allocation;
the complete reply envelope and byte budget are checked separately.

Tests cover signed extremes, the full interface/address/action/mode matrix,
source and output budgets, counterpart corruption, exact instance selection,
validation order, and TCP/TLS ordinary/emergency/default handling. They verify
Session/Proxy envelopes, owned output, no invalid application delivery,
preservation of held resources, valid follow-up and retirement.

## Reproduce

Use the pinned toolchain from the core module:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(ECURInterfaceValue|RejectECURInterfaceValue|InterfaceValue|RejectInterfaceValue)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(ECURInterfaceValue|RejectECURInterfaceValue|InterfaceValue|RejectInterfaceValue)' -count=20 -timeout=10m
CGO_ENABLED=0 GOMAXPROCS=2 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Combined advances to `0.1.0-slice.15`, changing its definition digest and
dependent Bundle fingerprints. Public signatures/layouts, generator v23 and
operations v9 remain unchanged. Keep original builds for older recovery records.
Published tags remain immutable; publication and cold installation of a new
reviewed pair are separate gates. The library stays pure Go with no persistence.

See [IEC interface values](conditional-interface-values.md),
[remaining gaps](remaining-gaps.md) and
[batch evidence](../spec/evidence/batch-120-ecur-interface-values.json).
