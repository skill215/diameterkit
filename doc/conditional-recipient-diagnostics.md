<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Conditional combined Recipient-Address diagnostics and acceptance.
SPDX-AI-Date: 2026-11-18
-->

# Conditional Recipient-Address diagnostics

Batch 111 adds a bounded `5005 DIAMETER_MISSING_AVP` diagnostic for one existing
condition in the combined sms-fw binding. IEC, INITIAL and TERMINATION requests
already require Recipient-Address inside their MMS-Information. The wire field
remains optional because the same binding also carries money templates.

## Selected condition

The existing SMS receive rule first checks the request sequence, two ordered
subscriptions, SMS template markers, identity relation and required SMS fields.
After confirming MMS-Information is present, it supplies a typed missing-group
proof only when Recipient-Address is absent and Submission-Time, Priority,
Message-Size (UserDataLength), SMSKey and Status-Report-Request are all present.

A second missing conditional MMS field keeps the earlier coarse missing error
and cannot produce this group example. Missing the entire MMS/SMS group and
other conditional omissions remain outside this slice. Money templates retain
their existing rules. No condition is inferred from a bare error sentinel.

Build/send still reject the missing recipient with their existing error path.
Receive refines the path to Vendor10415 `873 / 877 / 1201`. It does not change
which requests are valid or skip later business-value checks. A present empty
Recipient-Address retains its actual missing-child error; wrong Address-Type,
NUL-containing Address-Data, timestamp mismatch and excess groups retain their
existing validation. A present zero-length Address-Data remains permitted.

## Exact minimum and bounds

The private proof uses the existing reviewed V2 group convention and contains
only static schema metadata. It does not retain request buffers or address data.
As with other schema callbacks, the producer is trusted code; this metadata is
not authentication or a proof of a valid business request.

| Element | Bytes |
| --- | ---: |
| Service-Information ancestor header | 12 |
| MMS-Information ancestor header | 12 |
| Recipient-Address group header | 12 |
| Address-Type header and four zero bytes | 16 |
| Address-Data header with empty octets | 12 |
| Complete diagnostic subtree | 64 |

The zero Address-Type is a minimum-size example, not a valid address or an
inferred business role. No sibling values are copied. The existing consumer
checks the complete typed tree, flags, public error path and byte/field/frame,
depth and AVP budgets before allocating. The outer Failed-AVP adds one AVP and
one depth level: the selected subtree needs six AVPs and depth four including
that wrapper. The complete answer has its own envelope and byte limits.
A changed path cannot downgrade explicit evidence to an older placeholder.

TCP and TLS tests cover all three templates in ordinary and emergency slots,
exact bytes and Session/Proxy contents, invalid-request non-delivery, held charges,
valid follow-up traffic and retirement. Default-close and each additional missing
conditional MMS field are exercised over both transports.

## Reproduce and migrate

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/missingavp -run 'Test(RecipientMissing|RejectRecipientMissing|MinimumVariable|RejectMinimumVariable)' -count=20 -timeout=10m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/missingavp -run 'Test(RecipientMissing|RejectRecipientMissing|MinimumVariable|RejectMinimumVariable)' -count=20 -timeout=10m
CGO_ENABLED=0 go test ./internal/missingavp -run '^$' -fuzz '^FuzzMinimumVariableTree$' -fuzztime=30s -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Combined advances to `0.1.0-slice.7`, changing its definition digest and dependent
Bundle fingerprints. Public signatures/layouts, generator v23 and operations v9
remain unchanged. Keep original builds for old recovery records; published tags
remain immutable. A new publication and cold installation are separate gates.
The library stays pure Go and introduces no persistence.

See [remaining gaps](remaining-gaps.md), [typed minimum groups](variable-group-diagnostics.md)
and [the batch receipt](../spec/evidence/batch-111-recipient-missing.json).
