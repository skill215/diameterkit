# Bounded fuzz, pressure and performance acceptance

Run from the repository root on Linux/amd64 with the pinned toolchain:

```sh
python3 spec/tools/release_check.py --suite fuzz
python3 spec/tools/release_check.py --suite stress
python3 spec/tools/release_check.py --suite bench
```

The runner writes `fuzz.json`, `pressure.json`, `performance.json`, command logs
and a summary under `.cache/release-check/<suite>/`. CI uses the same commands.
The libraries remain pure Go and introduce no persistence service.

## Fixed work and stopping conditions

| Workload | Bound | Failure conditions |
| --- | --- | --- |
| FuzzDecode, FuzzDecodeAnswer, FuzzCanonical | 30s each, two workers, 5s minimization, 2m package timeout | Property failure, panic, timeout, no actual executions, nonzero exit |
| Sh-Data FuzzParse and SCTP mode selection | Same bounds; five targets total | Same; seed-only tests are not fuzz execution |
| Ro TCP pressure | Windows 1/8/16, 16 waves each, 20 repetitions per pure/race mode | Admission beyond limits, missing/duplicate outcome, resource retention, timeout |
| Mixed Sh/S6c/SGd/Ro TCP pressure | Four classes on one connection, windows 1/4 per class, each application stalled in turn, eight waves, 20 repetitions per mode | Saturated class bypass, blocked healthy classes, wrong/duplicate result, resource retention, timeout |
| SCTP pressure | Four connection pairs, 128 frames per pair, sizes 20/256/4096, 20 repetitions per mode | Corrupt frames, I/O failure, over 1MiB fixed buffer reserve, successful writes after Close |

Each Ro repetition has 400 requests, 96 full-window refusals and 104 explicit
cancellations. Completed-but-unconsumed results still hold port capacity. Dynamic
outbound/inbound accounting must return to zero after consumption; the entire
ResourceSnapshot must be zero after Close/Wait. Nodes have 1GiB logical limits,
with 256MiB class/port limits. These numbers are reservations, not preallocated Go heaps.

SCTP uses eight transports and 512 frames per repetition. Tests require real kernel
sockets. Existing cancellation, reuse and other regressions remain in full tests.

Fuzz mutations and execution counts vary with the seed/cache state. A separate
GOCACHE can exercise a cold start but also changes build costs. On failure the
minimized `testdata/fuzz/<Target>/<hash>` is copied into the artifact's `fuzz-corpus/`.
Checked-in regression inputs are never deleted. Replay with
`go test <package> -run '<Target>/<hash>'` after fixing the defect.

## Resource measurements

- **Logical peak:** controlled-stage Node.Resources samples including fixed,
  outbound, inbound and duplicate reservations; not RSS.
- **SampledHeapAlloc:** maximum sampled Go heap allocation across the process;
  not a continuous absolute peak or per-request ownership measurement.
- **ProcessRSSHighWater:** Linux VmHWM for the entire test process, including
  runtime/race costs and earlier repetitions. A 1GiB guard is a test failure
  threshold, not a production sizing recommendation.

The compiler process is outside these measurements. Kernel queues, deployment
saturation curves, long soaks, slow downstreams and multi-tenant capacity require
separate declared workloads and evidence.

## Mixed application isolation

`TestBoundedMixedApplicationPressure` compiles Base, Sh, S6c, SGd and Ro into one
bundle and advertises each application explicitly. It uses one shared TCP
connection between two Nodes, four resource classes and four completion ports.
A fifth empty port probes class limits independently of a full completion port;
the peer has one spare slot so its slot limit cannot explain that refusal.

Each of Sh Ref24 UDR, S6c SRR, SGd OFR and Ro IEC CCR becomes the stalled class in
turn. Other classes must complete fresh requests while the stalled class first
holds unanswered inbound work, then holds completed but unconsumed outcomes.
Alternating waves cancel selected stalled requests before a late answer. Every
admitted request must have exactly one correctly correlated outcome; Sh/S6c
return explicit 5012 failures and SGd/Ro return selected 2001 answers. These are
synthetic protocol outcomes, not HSS, delivery or charging decisions.

One repetition covers eight scenarios, 1,600 requests, 1,440 healthy-class
completions, 256 expected refusals and 48 cancellations. Both pure and race modes
repeat it 20 times. The runner requires all 160 scenario records per mode and
rejects missing work, missing measurements or exceeded resource guards.

The declared test limits are 1 GiB per Node, 64 MiB per outbound class/port,
128 MiB per inbound class and 256 MiB per peer direction. With the selected codec
limits, four S6c inbound reservations need 72,400,896 bytes: a 64 MiB inbound
class would reject this workload before its four-slot window fills. No codec
bound is reduced to make the test pass. Logical reservations, sampled HeapAlloc
and process VmHWM are recorded separately. Dynamic charges must return to zero
after every wave, and all resources must be zero after Close, port drain and Wait.

```sh
CGO_ENABLED=0 go test ./examples/ro -run '^TestBoundedMixedApplicationPressure$' -count=20 -timeout=3m
CGO_ENABLED=1 go test -race ./examples/ro -run '^TestBoundedMixedApplicationPressure$' -count=20 -timeout=3m
# Includes mixed, original Ro and native SCTP pressure, with machine checks:
python3 spec/tools/release_check.py --suite stress
```

This establishes finite application-class isolation with available shared
capacity. It does not establish tenant security isolation, byte-weighted fairness
ratios, network head-of-line freedom, long-soak stability or production capacity.
The stall is controlled by application progress, not a timing sleep or a real
remote outage; no wall-clock latency percentile is inferred.

The [batch 94 evidence](../spec/evidence/batch-94-mixed-pressure.json) records
the actual 32,000 requests per mode, resource samples, commands and log hashes.
Earlier failed budget probes remain diagnostic failures, separate from the
correctly provisioned passing workload.

## Benchmarks

Eight cases cover five codec operations, a complete Ro TCP request/answer
consumption, and 256/4096-byte SCTP one-way loopback send plus owned receive.
Each case uses 200 iterations, five samples, cpu=1 and CGO=0. The parser rejects
missing work, missing allocations, nonfinite values or incomplete samples.

Connection establishment, profile compilation and fixture loading are outside
timing. Append reuses caller storage; Decode/Canonical include owned allocation.
Ro includes request construction, identity allocation, codec/network work and result
consumption. SCTP is not a remote round-trip benchmark.

The [baseline](../spec/release/performance-baseline.json) records all samples,
min/median/max, CPU/kernel/toolchain and workload hashes. Each ns/op is a sample's
mean; a median across five samples is not a request p50/p99. Compare only equivalent
workloads, hardware, toolchain, CPU settings and load. No cross-host time threshold
or automatic baseline overwrite is used. The [README table](../README.md#benchmarks)
reports these existing measurements, not new results from a documentation edit.
