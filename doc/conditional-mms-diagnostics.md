<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Isolated combined MMS field diagnostics and reproducible acceptance.
SPDX-AI-Date: 2026-11-18
-->

# Conditional MMS field diagnostics

Batch 112 extends the combined SMS receive condition to five isolated missing
MMS fields. IEC, INITIAL and TERMINATION already require these fields.
The wire fields remain optional for the shared money templates. Valid requests,
business-value checks and Build/Send behavior are unchanged.

| Missing field | Vendor10415 code | Minimum payload | Diagnostic including ancestors |
| --- | ---: | --- | ---: |
| Submission-Time | 1202 | Four zero Time bytes | 40 bytes |
| Priority | 1209 | Four zero Int32 bytes | 40 bytes |
| UserDataLength | 1212 | Four zero Uint32 bytes | 40 bytes |
| SMSKey | 1210 | Empty octets | 36 bytes |
| Status-Report-Request | 1216 | Four zero Int32 bytes | 40 bytes |

The existing condition still runs after sequence, subscription, SMS marker,
identity and required SMS checks. MMS-Information must be present. Exactly one
of its six conditional fields may be absent; the sixth field, Recipient-Address,
retains the [batch 111 group diagnostic](conditional-recipient-diagnostics.md).
Two or more missing fields keep the original coarse missing error and close
boundary. Missing the entire MMS group or earlier SMS fields remains separate.

Private evidence contains only schema keys, flags, exact primitive minima and
the ancestor chain `873 / 877 / field`. No request or subscriber values are
retained. Scalars use the existing fixed-width convention; SMSKey explicitly
uses the existing zero-octet convention. The producer remains trusted schema code.
A zero-filled example is not an inferred timestamp, priority, length or report
choice. A present empty SMSKey remains permitted; NUL bytes and invalid scalar
lengths retain their existing error classifications.

The consumer checks metadata, path and byte/field/frame/count/depth limits before
allocation. The diagnostic plus its Failed-AVP wrapper needs four AVPs and depth
three. Complete answers also enforce envelope and reply budgets. Changed paths
cannot downgrade explicit evidence to legacy placeholders.

Acceptance covers all 63 nonempty missing-field combinations for each of the
three SMS templates. TCP/TLS covers five fields by three templates in ordinary
and emergency slots (60 positive cases), exact diagnostic and Session/Proxy
contents, held charges, non-delivery, follow-up and retirement. Default-close
covers all five fields on both transports; earlier multi-omission and recipient
tests remain active. Build/Send retain their prior error path.

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/missingavp -run 'Test(MMSMissing|RejectMMSMissing|RecipientMissing|RejectRecipientMissing|MinimumVariable|RejectMinimumVariable|VariableMinimum)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/missingavp -run 'Test(MMSMissing|RejectMMSMissing|RecipientMissing|RejectRecipientMissing|MinimumVariable|RejectMinimumVariable|VariableMinimum)' -count=20 -timeout=10m
CGO_ENABLED=0 go test ./internal/missingavp -run '^$' -fuzz '^FuzzVariableMinimum$' -fuzztime=30s -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Combined advances to `0.1.0-slice.8`; its definition digest and dependent Bundle
fingerprints change. Public signatures/layouts, generator v23 and operations v9
are unchanged. Keep original builds for earlier recovery records. Published
tags remain immutable; a new release and cold installation remain separate.
The library stays pure Go and introduces no persistence.

See [remaining gaps](remaining-gaps.md) and [the receipt](../spec/evidence/batch-112-mms-missing.json).

[Batch 113](conditional-sms-diagnostics.md) separately adds common SMS and
termination-only isolated omissions, preserving the existing first-failure order.
