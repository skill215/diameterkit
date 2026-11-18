<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Conditional EVENT Requested-Action omission and acceptance.
SPDX-AI-Date: 2026-11-18
-->

# EVENT Requested-Action omission

Batch 116 adds a bounded 5005 diagnostic for an absent Requested-Action in the
combined IEC and money-template EVENT paths. These paths already required this
field; they previously shared an unsupported failure with present invalid values.

| Input at the existing condition | Receive behavior |
| --- | --- |
| IEC or money EVENT, Requested-Action absent | 5005 evidence; both ErrMissing and ErrUnsupported |
| IEC, present value other than 0 | Existing unsupported failure |
| Money EVENT, present value greater than 2 | Existing unsupported failure |
| Money EVENT, present value 0, 1 or 2 | Existing operation-specific checks |
| INITIAL or TERMINATION | Existing absence/forbidden-presence rules |

The money paths cover direct debit, balance and refund requests. The diagnostic
does not choose or reconstruct an operation. Its four zero data bytes express
the minimum UInt32 width only; they are not an instruction to debit or a fallback
for a missing refund/balance action.

The private proof contains only key 436, Mandatory flags and width 4. It owns no
request buffer, action value or financial state. Its **12-byte** example requires
**two AVPs and depth 1** including Failed-AVP. The consumer checks byte, field,
frame, count, depth and exact public-path consistency before allocation.
Complete-answer envelope, size and resource limits apply separately.

Receive retains `errors.Is(err, validation.ErrUnsupported)` and adds
`errors.Is(err, validation.ErrMissing)` only for the selected absence.
Build/Send keep their original unsupported classification and key-436 path.
Bare errors and altered explicit paths cannot downgrade to a guessed example.
`BaseErrorFailures` remains opt-in; the default policy still closes.

Validation order is unchanged. Structural and typed-length decoding precedes
the subset rules. Sequence and subscriptions still precede both action checks.
IEC SMS/MMS and timestamp checks still precede its action check; the money
action check remains before its later template checks. An action omission can
therefore be the first reported fault even if a later condition is also invalid.

Acceptance includes 180 direct combinations across six templates, five values,
presence and Build/Send/Receive; a 60-case wire matrix; 32 exact/below-limit/path
checks; and 27 earlier/later-failure cases. Sixteen positive TCP/TLS ordinary/
emergency cases check exact diagnostic and Session/Proxy data, no invalid
delivery, held charges, follow-up requests and retirement. Eight default-close
and eight present-invalid-value network cases preserve their close behavior.

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/missingavp -run 'Test(ActionMissing|RejectActionMissing|MMSGroupMissing|RejectMMSGroupMissing|OriginatorMissing|RejectOriginatorMissing|SMSMissing|RejectSMSMissing|MMSMissing|RejectMMSMissing|RecipientMissing|RejectRecipientMissing|MinimumVariable|RejectMinimumVariable|VariableMinimum)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/missingavp -run 'Test(ActionMissing|RejectActionMissing|MMSGroupMissing|RejectMMSGroupMissing|OriginatorMissing|RejectOriginatorMissing|SMSMissing|RejectSMSMissing|MMSMissing|RejectMMSMissing|RecipientMissing|RejectRecipientMissing|MinimumVariable|RejectMinimumVariable|VariableMinimum)' -count=20 -timeout=10m
CGO_ENABLED=0 go test ./internal/missingavp -run '^$' -fuzz '^FuzzMissingQuote$' -fuzztime=30s -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Combined advances to `0.1.0-slice.12`; definition digests and dependent Bundle
fingerprints change. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Earlier recovery records need their original build. Published
tags remain immutable; new publication and cold installation are separate gates.
The library stays pure Go with no persistence.

See [remaining gaps](remaining-gaps.md),
[the omission overview](missing-avp-diagnostics.md) and
[the acceptance receipt](../spec/evidence/batch-116-action-missing.json).
