<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Conditional whole MMS omission and reproducible acceptance.
SPDX-AI-Date: 2026-11-18
-->

# Conditional whole MMS omission

Batch 115 adds a bounded 5005 diagnostic when MMS-Information is wholly absent
from a combined INITIAL, TERMINATION or IEC SMS request. It runs at the existing
Receive validation branch, after earlier sequence, subscription and SMS checks.
The error still satisfies `errors.Is(err, validation.ErrMissing)`. Build/Send
retain their prior error and path. Money templates are unchanged.

The static V2 minimum describes Service-Information / MMS-Information and all
required children for this condition:

| MMS child | Minimum data |
| --- | --- |
| Originator-Address (structurally required) | Address-Type: four zero bytes; Address-Data: zero octets |
| Recipient-Address (conditionally required) | Address-Type: four zero bytes; Address-Data: zero octets |
| Submission-Time | Four zero bytes |
| Priority | Four zero bytes |
| User-Data-Length | Four zero bytes |
| SMS-Key | Zero octets |
| Status-Report-Request | Four zero bytes |

These are minimum-size diagnostic examples, not reconstructed business values
or a valid replacement MMS object. The proof retains no request, address, key or
timestamp. The complete example is **180 bytes**, including the Service ancestor.
It contains 13 AVPs; including Failed-AVP requires **14 AVPs and depth 4**.
Byte, field, frame, count and depth limits are checked before allocation.
The complete answer has separate envelope and reply budgets.

An existing empty MMS group still fails its structural Originator-Address
requirement. A present MMS group with multiple missing conditional children
retains its existing coarse failure. This change does not infer other group
minima, alter first-failure order or add interface-value diagnostics.
Changed public paths cannot fall back to an untyped example.
`BaseErrorFailures` is opt-in; default close behavior remains unchanged.

Acceptance covers 36 direct mode/scope combinations, all three SMS templates,
27 exact/below-limit/path checks and earlier/later failure precedence. Twelve
TCP/TLS ordinary/emergency cases verify exact diagnostic and Session/Proxy
contents, no invalid delivery, held resources, follow-up requests and retirement.
Six default-policy network cases still close.

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/missingavp -run 'Test(MMSGroupMissing|RejectMMSGroupMissing|OriginatorMissing|RejectOriginatorMissing|SMSMissing|RejectSMSMissing|MMSMissing|RejectMMSMissing|RecipientMissing|RejectRecipientMissing|MinimumVariable|RejectMinimumVariable|VariableMinimum)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/missingavp -run 'Test(MMSGroupMissing|RejectMMSGroupMissing|OriginatorMissing|RejectOriginatorMissing|SMSMissing|RejectSMSMissing|MMSMissing|RejectMMSMissing|RecipientMissing|RejectRecipientMissing|MinimumVariable|RejectMinimumVariable|VariableMinimum)' -count=20 -timeout=10m
CGO_ENABLED=0 go test ./internal/missingavp -run '^$' -fuzz '^FuzzMinimumVariableTree$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Combined advances to `0.1.0-slice.11`; its definition digest and dependent Bundle
fingerprints change. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Keep original builds for earlier recovery records and never
relabel published tags. New immutable publication and cold installation remain
separate gates. The library stays pure Go with no persistence.

See [remaining gaps](remaining-gaps.md), [MMS leaf omissions](conditional-mms-diagnostics.md)
and [the acceptance receipt](../spec/evidence/batch-115-mms-group-missing.json).
