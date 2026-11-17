# Cross-peer and mixed-profile duplicate acceptance

Batch 99 adds finite acceptance tests for the existing Node-local volatile cache.
Production code, generated definitions, the public API snapshot, module pins and
published tags are unchanged. The library remains pure Go with no persistence.

## Identity and cross-peer behavior

The cache key is the case-normalized Origin-Host plus End-to-End identifier.
Peer, HBH, T flag and Proxy-Info do not create a new logical request. The selected
command/profile and business content participate in canonical comparison:
reusing an existing key for a different profile is a conflict, not independent
work. Source and destination policy still run before cache lookup. A different
authorized Origin-Host can use the same End-to-End identifier independently.

On real loopback TCP and TLS, the tests cover:

| Scenario | Leaf cases per transport | Checks |
| --- | ---: | --- |
| Six combined construction paths × normal 2001/5012 and E/3004 × connected/disconnected primary | 36 | One delivery, one cache charge, cross-peer processing and cached replay, per-arrival HBH/Proxy-Info, content-conflict refusal |
| Source, duplicate-policy and arrival-limit refusal on the second peer | 3 | No join or second delivery; original work and result remain usable |
| Each of combined, S6c, Sh and SGd owns a key | 4 | Other profiles conflict before and after commitment; an authorized independent origin creates new work |
| Shared/independent cache classes × each of four delayed applications | 8 | Four logical requests and eight arrivals coexist; three profiles complete and replay while one stays pending |
| Shared/independent classes × slot/byte exhaustion | 4 | Cache hits succeed at capacity; new records cannot evict results; isolation, expiry and re-admission |

There are 110 leaf cases across both transports, or 125 Go test/subtest pass
events per run. Both pure and race repetitions use the same assertions.

The combined cases retain the original six construction inputs: IEC debit,
general debit, balance, refund, INITIAL and TERMINATION. Both debit constructions
belong to the existing five-template binding. These are not evidence of complete
replacement for the original private six-template application.

Content conflicts alter the second Subscription-Id and its UEI copy where
required, preserving valid input while changing canonical content. Hits and
conflicts must not change canonical bytes, execution or retention deadlines,
record identity or committed-result authority. After the primary connection is
lost, `Respond` must report its primary write failure while the admitted second
peer still receives the committed result.

## Cache budgets and retirement

The mixed workload uses a 2 GiB logical Node limit, a 512 MiB inbound class with
16 slots, and either one shared 128 MiB cache class or four independent 128 MiB
classes. Each peer allows 16 inbound arrivals. Cache policies retain at most four
arrivals per logical request, 2,048 canonical bytes, 700 result bytes and 300
Proxy-Info bytes. Processing and retention are bounded at five and thirty seconds;
the Node request limit is ten seconds and watchdog interval is two minutes.
These are finite test reservations, not production memory-sizing guidance.

Byte-limit cases obtain each profile's compiled reservation size before creating
the tested Node. They configure exactly one reservation per independent class,
or the sum of all four in a shared class, while leaving slot capacity available.
Slot-limit cases leave byte headroom. No account limits are changed after Node
construction. Aggregate `DuplicateBytes` must equal the sum of admitted profile
reservations; processing duplicates and cached hits must not increase it.

With a shared full cache, all four retained results replay without eviction. With
independent classes, a full combined class cannot prevent the other three from
admitting work. Refusal closes the affected peer with backpressure while the
other peer can still use retained results. Expiry advances the existing internal
timer input to the captured retention deadline, then requires zero cache charges
and a new logical record for the old identity. This is not a wall-clock soak.

Close/Wait must retire all charged resources. The monotonic conflict counter is
checked separately and is not mistaken for retained memory. The tests do not
promise durable charging deduplication, fairness ratios, tenant capacity, mixed
transport interoperability or real OCS recovery permission.

## Reproduction

```sh
CGO_ENABLED=0 go test . -run 'Test(CombinedCrossPeer|MixedDuplicate)' \
  -count=20 -timeout=5m
CGO_ENABLED=1 go test -race . -run 'Test(CombinedCrossPeer|MixedDuplicate)' \
  -count=20 -timeout=5m
python3 spec/tools/release_check.py --suite all
```

The ordinary checkout pure/race CI suites include these tests. See the
[test source](../mixed_duplicate_test.go),
[batch 99 receipt](../spec/evidence/batch-99-cross-profile-duplicates.json),
[earlier same-peer acceptance](combined-duplicate-acceptance.md) and
[remaining gaps](remaining-gaps.md). Process notes remain ignored under
`spec/local/batches/`. Historical benchmark measurements remain unchanged.
