<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Reviewed IEC Originator-SCCP mixed-condition omission and acceptance.
SPDX-AI-Date: 2026-11-18
-->

# IEC Originator-SCCP omission

Batch 114 adds a bounded 5005 diagnostic for one unambiguous branch of the
combined IEC rule. That rule already rejects either an absent Originator-SCCP
address or an OriginatorInterface type other than 1.

| IEC input at this condition | Receive behavior |
| --- | --- |
| Address absent, interface type 1 | Typed missing example; both ErrMissing and ErrUnsupported |
| Address present, interface type 1 | Existing value and later request checks |
| Address present, interface type other than 1 | Existing unsupported failure; no missing example |
| Address absent, interface type other than 1 | Existing unsupported failure; no missing example |

The selected branch runs only after existing sequence, subscriptions, SMS/MMS,
timestamp and Requested-Action checks. It does not infer or repair an interface
value. INITIAL, TERMINATION and money-template rules are unchanged. Build/Send
keep their earlier ErrUnsupported and error path. Receive preserves
`errors.Is(err, validation.ErrUnsupported)` while adding
`errors.Is(err, validation.ErrMissing)` only for the selected isolated omission.

Private evidence contains the static Vendor10415 path `873 / 2000 / 2008`,
flags and an explicit zero-octet minimum. No address value or request buffer is
retained. The diagnostic is 36 bytes including both ancestor headers; its empty
address is a minimum-size example, not reconstructed subscriber information.
A present empty address remains permitted, and NUL-containing data retains its
existing failure.

The existing consumer checks metadata, public path and byte/field/frame/count/
depth limits before allocation. Including Failed-AVP needs four AVPs and depth
three; complete-answer envelope and byte budgets apply separately. Changed paths
cannot fall back to older untyped examples. BaseErrorFailures remains opt-in;
the default policy still closes.

Acceptance includes a 60-case direct mode matrix (IEC and non-IEC), a 12-case
wire matrix, invalid/empty values and earlier-failure precedence. Four positive
TCP/TLS ordinary/emergency cases verify exact diagnostic and Session/Proxy data,
no invalid delivery, held resources, valid follow-up and retirement. Two default
close and six mixed/earlier-unsupported network cases retain the close boundary.

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/missingavp -run 'Test(OriginatorMissing|RejectOriginatorMissing|SMSMissing|RejectSMSMissing|MMSMissing|RejectMMSMissing|RecipientMissing|RejectRecipientMissing|MinimumVariable|RejectMinimumVariable|VariableMinimum)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/missingavp -run 'Test(OriginatorMissing|RejectOriginatorMissing|SMSMissing|RejectSMSMissing|MMSMissing|RejectMMSMissing|RecipientMissing|RejectRecipientMissing|MinimumVariable|RejectMinimumVariable|VariableMinimum)' -count=20 -timeout=10m
CGO_ENABLED=0 go test ./internal/missingavp -run '^$' -fuzz '^FuzzVariableMinimum$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Combined advances to `0.1.0-slice.10`; its definition digest and dependent Bundle
fingerprints change. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Keep original builds for earlier recovery records. Published
tags remain immutable, and new publication/cold installation are separate gates.
The library stays pure Go with no persistence.

Wrong-interface diagnostics, simultaneous conditional omissions and unreviewed
whole groups remain separate. See [remaining gaps](remaining-gaps.md),
[SMS conditions](conditional-sms-diagnostics.md) and
[the receipt](../spec/evidence/batch-114-originator-missing.json).
