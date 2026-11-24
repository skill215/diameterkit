<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: SMS client wire conditions, bounded diagnostics and acceptance.
SPDX-AI-Date: 2026-11-24
-->

# SMS client wire values

Batch 135 adds bounded 5004 diagnostics for the two existing SMS request-level
ClientAddressWire (Vendor 10415, code 2018) conditions. It does not parse the
bytes as an IP address or change the accepted template values.

| Condition reached by the request validator | Receive evidence |
| --- | --- |
| Client bytes contain NUL | `combined-sms-client-nul-v1` |
| Client bytes are NUL-free but differ from SMSC-Address | `combined-sms-client-mismatch-v1` |
| Client and SMSC bytes are equal and NUL-free, including empty values | No new error |

Both errors retain `ErrField`, without adding `ErrUnsupported`. Receive supplies
the exact path `873/2000/2018`; Build/Send retain their original single-field
`2018` path. Static private evidence holds no request or address data.
`BaseErrorFailures` enables the reply; `CloseOnFailure` keeps connection closure.

## Validation order

The proof is issued at the existing failure branch. Sequence/subscription
constraints, SMS markers, equipment equality, SMS-Node/SM-Message-Type and the
five common required SMS fields remain earlier. NUL is checked before equality.
A NUL-containing client therefore uses NUL evidence even when it also differs
from SMSC-Address. Group structural errors and nested callback errors are earlier
than this request callback, including an SMSC NUL or malformed MMS scalar.

MMS presence, MMS template relations, action/interface and unit rules stay later.
A missing later MMS group does not suppress the client error. Optional root
extras are checked later and also do not suppress it. SMS group extras remain an
earlier failure. Money direct-debit/balance binary client values and refund
presence rules are independent and receive no SMS client evidence.

## Bounded consumption

The consumer scans the complete root and service/SMS areas with shared counters.
It verifies request kind/number, indicator and equipment markers, then closes
the fourteen-field SMS set: unique Vendor 10415 instances, correct flags,
eleven required fields, five scalar widths, zero node/message-type values and
NUL-free address counterparts. It then rechecks the specific NUL or mismatch
condition. Unknown, duplicate, missing, malformed or mismatched proof/path data
cannot produce a partial quote. Bare `ErrField` is insufficient.

Earlier subscription/equipment relationships and nested interface payloads are
trusted-decoder preconditions, not reinterpreted by this consumer. Unrelated
root/service payloads remain opaque. This named-evidence convention is not
proof of authentication or a general-purpose raw CCR validator.

For `n` client bytes, the owned quote is **`36 + align4(n)` bytes**: two minimal
12-byte vendor group headers and the entire original padded client AVP. It
copies no SMSC or other sibling bytes. Original padding is preserved and excluded
from the NUL/equality comparisons. Empty mismatch data produces a 36-byte quote;
NUL data requires at least one byte and a 40-byte quote.

The minimal consumer fixture uses seventeen scanned source AVPs, source depth
two, `252 + align4(n)` frame bytes and `176 + align4(n)` bytes of largest field
payload. It uses opaque empty equipment/interface payloads and is not a fully
valid CCR. Failed-AVP requires four output AVPs and depth three. Source, quote
and full-answer budgets are checked independently before allocation; insufficient
budget closes without truncation or a partial answer.

## Acceptance and reproduction

Tests cover IEC, INITIAL and TERMINATION; both errors; all optional SMS-field
subsets; exact/one-less budgets; malformed markers, flags, widths and duplicates;
original padding; empty equality; ownership; mixed errors; and validation order.
TCP/TLS ordinary and emergency tests verify exact envelopes, non-delivery,
held charges, valid follow-up, short-budget closure and resource retirement.

Use the pinned Go toolchain:

```sh
CGO_ENABLED=0 go test . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(SMSClient|RejectSMSClient)' -count=20 -timeout=15m
CGO_ENABLED=1 go test -race . ./compat/smsfwro/combined ./internal/excessavp -run 'Test(SMSClient|RejectSMSClient)' -count=20 -timeout=15m
GOMAXPROCS=2 CGO_ENABLED=0 go test ./internal/excessavp -run '^$' -fuzz '^FuzzInvalidQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

Local acceptance includes twenty-round diagnostic regression, full dual-module
pure/race, static consistency and one affected counted fuzz target. Hosted CI
runs all seventeen fuzz targets plus static, pure, race, cross, stress and
benchmarks. Actual status, failed attempts and artifact content checks are in
[batch evidence](../spec/evidence/batch-135-sms-client-values.json).

## Version boundary

Combined advances to `0.1.0-slice.29`, changing its definition digest and Bundle
fingerprints. Public signatures/layouts, generator v23 and operations v9 stay
unchanged. Keep original builds for earlier records; never relabel fingerprints.
Pure Go and application-owned persistence remain the boundary.

The frozen rc.1 publication receipt does not cover this checkout. The rc.2 pair
remains planned and unpublished and requires fresh immutable publication and
cold-install acceptance. See [remaining gaps](remaining-gaps.md) and the
[release checklist](release-candidate-checklist.md).
