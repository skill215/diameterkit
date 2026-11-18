<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Count-bounded fuzz acceptance and retained failure evidence.
SPDX-AI-Date: 2026-11-18
-->

# Count-bounded fuzz acceptance

Batch 117 makes the default checkout fuzz gate run **1,000,000 executions per
target** across the same seventeen targets. Batch 118 corrects the original
exact-count assumption: successful observed counts must be between 1,000,000
and 1,000,062 inclusive, with the actual overrun recorded. Every target must finish successfully;
any assertion, panic, nonzero exit, timeout or incomplete receipt fails the suite.
There is no automatic retry and no conversion of deadline failures into success.

## Why the stopping policy changed

The original timed runs in
[batch 114](../spec/evidence/batch-114-originator-missing.json) and
[batch 116](../spec/evidence/batch-116-action-missing.json) ended with
`context deadline exceeded` after their 30-second workloads, without a reported
assertion or saved counterexample. Their original failures and follow-up runs
remain historical evidence.

The pinned Go 1.26.10 source in `src/internal/fuzz/fuzz.go` has distinct
coordinator stopping paths: a positive time budget creates a deadline context,
whereas reaching a positive execution limit requests a normal stop. This
motivates selecting the count limit; it does **not** establish the root cause of
those intermittent exits or claim an upstream Go fix.

One million executions is a new finite acceptance workload. It is not equivalent
to 30 seconds of coverage: the old execution counts varied by target and host.
Mutation paths, seed/cache state and runtime duration still vary. Exact work
counts do not prove exhaustive coverage or deterministic input sequences.

## Bounds and receipts

| Setting | Current gate |
| --- | --- |
| Targets | The existing seventeen entries in `FUZZ_TARGETS` |
| Requested work | `-fuzztime=1000000x` per target; 17,000,000 total |
| Accepted observed count | 1,000,000–1,000,062 per target; actual counts recorded |
| Concurrency | Two workers and `GOMAXPROCS=2` |
| Minimization | Five seconds |
| Go package timeout | Two minutes per target |
| Outer command timeout | Fifteen minutes, including command/build overhead |
| Hosted suite timeout | Twenty minutes |

The runner requires a final execution count within that finite range, the two-worker fuzz marker,
and successful Go test markers, in addition to exit code zero. Seed-only output,
less than the requested work, more than the reviewed maximum, and failure text
cannot satisfy the gate.

`fuzz.json` is initialized for each invocation and updated after every target.
A failed target remains in the partial receipt with its log hash and failure
classification. Outer process timeouts retain the flushed log and its hash.
Files under a target's `testdata/fuzz` directory are copied, with hashes, into
the output's `fuzz-corpus` directory even on failure; original files are not
removed. A fresh output directory keeps separate attempts independently auditable.

The Python negative controls cover nonzero exits, timeout, misleading success
text, incomplete counts, stale receipts and corpus preservation. A separate
local real Go fuzzer deliberately discovers a counterexample, exits nonzero,
archives the minimized input and reproduces the failure from that input.
[Batch 117 evidence](../spec/evidence/batch-117-counted-fuzz.json) records the
commands, fixture, results and hashes.

## Reproduce

Use the repository's pinned Go toolchain and available module dependencies:

```sh
python3 spec/tools/release_check.py --suite fuzz --output .cache/fuzz-counted
python3 -m unittest discover -s spec/tools -p 'test_release_fuzz.py'
# The two targets involved in the historical timed exits:
CGO_ENABLED=0 GOMAXPROCS=2 go test ./internal/missingavp -run '^$' -fuzz '^FuzzMissingQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
CGO_ENABLED=0 GOMAXPROCS=2 go test ./internal/cardinalityavp -run '^$' -fuzz '^FuzzCardinalityQuote$' -fuzztime=1000000x -fuzzminimizetime=5s -parallel=2 -timeout=2m
# Replay a saved input in its original package:
go test <package> -run '^<Target>/<hash>$' -count=1
```

A failed target stops the suite. Investigate its original artifacts before any
separately recorded rerun. This policy does not authorize suppressing failures.

## Scope

This is release tooling only. No Go library source, public API, protocol rules,
profile versions, generator v23/ops9, Bundle fingerprints, module manifests or
published tags change. The libraries remain pure Go with no persistence service.
New immutable publication, deployment-specific long soaks and real-peer/private
compatibility acceptance remain separate. Diagnosis of arbitrary timed Go fuzz
exits also remains an external tooling follow-up.


## Batch 118 count correction

The first batch 118 hosted DiameterURI run exited zero with PASS after
1,000,019 executions. The original exact-count parser rejected it; the original
failed suite and log remain in the [batch 118 receipt](../spec/evidence/batch-118-action-values.json).

In pinned Go 1.26.10, `src/internal/fuzz/minimize.go` checks the stop predicate
before each byte's printable-candidate loop, then may attempt all 32 candidates
without another check. A worker can therefore exceed its allocated limit by
up to 31 calls in that loop. With two workers the reviewed allowance is 62.
This finite bound is tied to the inspected toolchain and worker policy; review it
when either changes. It is not a generic percentage tolerance or a waiver for
nonzero exits. The parser now accepts the observed count range and records the
actual overrun; 999,999 and 1,000,063 are failures.

The original hosted log is a regression fixture. Controls also cover both
accepted boundaries and still reject failure text, incomplete work and nonzero
exits. The fixture is measured tool output, not authored protocol data.
