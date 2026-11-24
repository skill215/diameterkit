<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Scope and acceptance of the combined money MMS exclusion.
SPDX-AI-Date: 2026-11-24
-->

# MMS fields in combined money requests

The development checkout can return 5008 with the original SMSKey for a general
debit or balance check. This applies to the named combined profile when a single
well-formed SMSKey appears in MMS-Information and the inbound policy enables
automatic Base errors. The published rc.2 pair contains the earlier slice.29.

## Evidence for each field

The general sender projection puts only Originator-Address in MMS-Information.
The balance operation reuses that template. The combined fixture generator also
adds a valid `sms-0001` key to an otherwise valid general request as an explicit
negative case. The existing combined Receive rule rejects its presence. These
sources establish a prohibition for the selected profile. Deployed OCS behavior
remains an external acceptance item. [General source projection](../interop/evidence/2026-11-07-smsfw-general-source.json) · [Combined fixtures and provenance](../interop/evidence/2026-11-07-smsfw-combined-source.json) · [Negative input](../testdata/codec/b47-general-mms-key.hex).

| Present child (Vendor 10415) | Isolated Receive outcome |
| --- | --- |
| SMSKey (1210) | Well-formed key, including empty, produces 5008 under the automatic error policy. A key containing NUL still reaches the earlier 5004 group check. |
| Recipient-Address (1201) | Existing unsupported group error and connection closure for an otherwise valid address. |
| Submission-Time (1202) | Existing unsupported group error and connection closure. |
| Priority (1209) | Existing unsupported group error and connection closure. |
| User-Data-Length (1212) | Existing unsupported group error and connection closure. |
| Status-Report-Request (1216) | Existing unsupported group error and connection closure. |

For the last five fields the archived projection establishes sender omission.
The receiver rejects these fields in the selected implementation, but the available
template evidence does not establish a field-specific wire prohibition for them.
When several otherwise valid excluded fields occur together, the receiver retains
the group-level error. Refund rejects the whole MMS group at its earlier rule.
Build and Send keep the original unsupported classification and group path.

## Validation and budgets

Receive identifies the exact `873/877/1210` instance. The bounded scanner
rechecks the money request selectors and a valid client address, the MMS group
and its required originator address. Each branch shares the original decoder
budget. The quote copies the entire key, including its original padding, under
the two ancestor headers. It needs `36 + align4(key length)` bytes. A short
quote or answer budget follows the existing close path. [RFC 6733 §7.1.5](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.1.5).

The tests exercise all 64 presence combinations per money operation, the
refund boundary, malformed-key and invalid-client precedence, exact and
one-short limits, the frozen negative fixture, and WSL TCP/TLS loopback under
both failure policies. The combined binding advances to slice.30. Dependent
Bundle fingerprints change. Keep original builds for earlier recovery records.

```sh
CGO_ENABLED=0 go test -count=1 -run 'Test(MoneyMMS|RejectMoneyMMS)' . ./compat/smsfwro/combined ./internal/excessavp
CGO_ENABLED=1 go test -race -count=1 -run 'Test(MoneyMMS|RejectMoneyMMS)' . ./compat/smsfwro/combined ./internal/excessavp
python3 spec/tools/release_check.py --suite all
```
