<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated combined SMS condition diagnostics and reproducible acceptance.
SPDX-AI-Date: 2026-11-18
-->

# Conditional SMS field diagnostics

Batch 113 adds bounded 5005 examples at two existing conditions in the combined
SMS receive rule. Accepted requests, wire optionality and Build/Send errors
remain unchanged. The single wire decoder keeps ClientAddressWire opaque.

| Condition | Missing field | Vendor10415 code | Minimum payload |
| --- | --- | ---: | --- |
| IEC / INITIAL / TERMINATION | ClientAddressWire | 2018 | Empty octets |
| IEC / INITIAL / TERMINATION | RecipientSCCPAddress | 2010 | Empty octets |
| IEC / INITIAL / TERMINATION | SMSCAddress | 2017 | Empty octets |
| IEC / INITIAL / TERMINATION | ReplyPathRequested | 2011 | Four zero Int32 bytes |
| IEC / INITIAL / TERMINATION | SMUserDataHeader | 2015 | Empty octets |
| TERMINATION only | SMStatus | 2014 | Empty octets |
| TERMINATION only | DischargeTime | 2012 | Four zero Time bytes |

Exactly one field must be missing within the condition currently being checked.
The common five-field condition runs after sequence, subscriptions, template
markers, identity relation and SMS node/message type checks. The termination
pair runs only after the existing MMS, interface and final-usage checks.
Two or more absent fields within one condition retain the coarse missing error
and close boundary. Originator-SCCP mixed conditions and whole-group absence
remain outside this slice.

First-failure order is unchanged: a common SMS omission precedes a later MMS or
termination omission; a missing MMS field precedes a termination omission.
A diagnostic identifies only that selected earlier failure, not every fault in
the request. It does not assert that the remaining request is business-valid.
An earlier unsupported or multiple-missing condition cannot be skipped to obtain
a later diagnostic. Non-termination status/time fields remain unsupported.

Private metadata records the fixed path `873 / 2000 / field`, flags and exact
primitive minimum. No source values are retained, copied or reconstructed.
ClientAddressWire is never parsed or converted; neither an SMSC string nor
general-template address bytes are inferred. Existing string NUL checks,
ClientAddressWire/SMSC equality, scalar lengths and opaque binary header/status
behavior are preserved. Present empty octets remain governed by those rules.

The ancestor chain and leaf total 36 bytes for octets and 40 for scalars.
The outer Failed-AVP brings the minimum to four AVPs and depth three.
The existing consumer checks metadata/path and byte/field/frame/count/depth
budgets before allocation; complete-answer budgets apply separately.
Zero-filled examples do not supply actual addresses, report choices or times.

Tests cover all 31 nonempty common-condition subsets per SMS template, all 3
termination-pair subsets, cross-condition precedence, mode boundaries and present
values. TCP/TLS covers 68 positive cases (17 applicable template/field pairs,
ordinary/emergency slots), 14 default-close cases and 4 multiple-omission close
cases. Exact diagnostic/Session/Proxy data, no invalid delivery, held resources,
valid follow-up and retirement are verified.

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/missingavp -run 'Test(SMSMissing|RejectSMSMissing|MMSMissing|RejectMMSMissing|RecipientMissing|RejectRecipientMissing|MinimumVariable|RejectMinimumVariable|VariableMinimum)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/missingavp -run 'Test(SMSMissing|RejectSMSMissing|MMSMissing|RejectMMSMissing|RecipientMissing|RejectRecipientMissing|MinimumVariable|RejectMinimumVariable|VariableMinimum)' -count=20 -timeout=10m
CGO_ENABLED=0 go test ./internal/missingavp -run '^$' -fuzz '^FuzzVariableMinimum$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Combined advances to `0.1.0-slice.9`; its definition digest and dependent Bundle
fingerprints change. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Retain original builds for earlier recovery records. Published
tags remain immutable; new publication and cold installation are separate gates.
The library stays pure Go with no persistence.

See [remaining gaps](remaining-gaps.md), [MMS diagnostics](conditional-mms-diagnostics.md)
and [the receipt](../spec/evidence/batch-113-sms-missing.json).

[Batch 114](conditional-originator-diagnostics.md) separately adds isolated IEC
Originator-SCCP absence with a valid existing interface condition. Wrong-interface
and combined mixed failures remain unsupported without a missing example.
