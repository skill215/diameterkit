# Combined binding: volatile duplicate acceptance

The fixed `compat/smsfwro/combined` binding now has explicit server-side duplicate
acceptance tests on real loopback TCP and TLS. This complements its existing
independent request/answer and client recovery tests. It introduces no new
protocol fields, runtime policy, database, automatic business retry or durable
charging deduplication.

## Verified cases

| Coverage | Cases on each transport | Required behavior |
| --- | ---: | --- |
| IEC debit, general debit, balance, refund, INITIAL and TERMINATION; each with normal 2001, normal 5012 and E/3004 | 18 | One application delivery, one answer per arrival, then cached replay |
| Each construction path as the original; all five other paths reuse its Origin/E2E identity | 6 scenarios, 30 conflicts | Reject conflicting content with E/5012; keep the original record and result authority |
| Arrival limit, cache limit, expiry and concurrent Respond/Close | 4 | No record eviction, bounded refusal/retirement and zero final resources |

The two debit constructions are separate inputs within the existing five-template
binding. These tests do not establish equivalence with the original private
six-template application or validate an actual OCS.

Processing duplicates have different HBH and Proxy-Info values. Answers are
associated by HBH and checked against each arrival's original envelope; the first
arrival's Proxy chain must never leak into another answer. A later T=0 request
with the same identity/content still uses the cached result. T alone does not
determine whether a request is a duplicate.

Content conflicts modify the second Subscription-Id and, where required, its UEI
copy so the conflicting request still satisfies the fixed template. The test
checks conflicts both during processing and after commitment. Neither a hit nor
a conflict may extend the original execution deadline or retention deadline,
replace the result, or create a second application delivery. Template changes
under the same identity are conflicts, not permission to start a new operation.

Mutating the delivered request view cannot change response authority. Invalid
responses and canceled preparation leave the response right available; after a
valid commit, mutating the caller's answer cannot change queued or cached bytes.
Normal and E results keep their separate envelopes and result codes.

## Budgets and lifecycle

The workload declares a 1 GiB Node limit, a 128 MiB duplicate class, a 256 MiB
inbound class/peer byte limit, at most four active arrivals, a 2,048-byte canonical
bound, 700 result bytes and 300 proxy-envelope bytes. Default test execution and
replay-retention bounds are two and five seconds respectively. These are logical
reservations and finite test settings, not production memory sizing advice.

Limit tests reduce the cache to one record and arrivals to one. A new identity
cannot evict the promised cached result, and another arrival cannot exceed the
arrival limit. The existing fail-closed path is required. Expiry is exercised by
advancing the existing internal timer input to the captured retention deadline;
it is not a long wall-clock soak. After expiry, the same request can become new
application work, which illustrates why this cache is not a durable ledger.

The lifecycle cases require the entire ResourceSnapshot to be zero after Close
and Wait. They also race response preparation with Close. Existing generic tests
continue to cover writer-borrow retirement; this batch does not replace them.

## Reproduction and evidence

Run from the repository root with the pinned Go toolchain:

```sh
CGO_ENABLED=0 go test . -run '^TestCombinedDuplicate' -count=20 -timeout=4m
CGO_ENABLED=1 go test -race . -run '^TestCombinedDuplicate' -count=20 -timeout=4m
python3 spec/tools/release_check.py --suite static
python3 spec/tools/release_check.py --suite pure
python3 spec/tools/release_check.py --suite race
```

The ordinary pure/race CI suites include these tests. See the
[test source](../smsfw_combined_duplicate_test.go) and
[batch 95 evidence](../spec/evidence/batch-95-combined-duplicates.json) for actual
results, commands and log hashes. The libraries remain pure Go; race uses CGO only
for instrumentation. Process notes remain under ignored `spec/local/batches/`.

Subsequent batches implement [selected 5005](missing-avp-diagnostics.md),
[5009](excess-avp-diagnostics.md) and [5004/5001](invalid-avp-diagnostics.md)
diagnostics. [Batch 99](cross-peer-duplicate-acceptance.md) adds combined cross-peer
and four-profile shared/independent cache-budget acceptance. Other diagnostics,
larger tenant/mixed-transport workloads, private ACA/PSP/1292 mappings, real target
recovery and deployment capacity remain separate. See the
[current gap list](remaining-gaps.md). No published version
tag or measured benchmark baseline is changed by this acceptance work.
