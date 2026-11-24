<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: RC publication scope and requirement completion conditions.
SPDX-AI-Date: 2026-11-19
-->

# Version, API and immutable installation policy

The [machine policy](../spec/release/version-policy.json) governs two independent
v0 modules. Core-only consumers can install v0.2.0-rc.2. The latest pair with
completed installation receipts is [core and adapter v0.2.0-rc.1](releases/v0.2.0-rc.1.md).
The [rc.2 release decision](releases/v0.2.0-rc.2.md) records each module's status.

| Module | Published version | rc.2 status |
| --- | --- | --- |
| github.com/skill215/diameterkit | v0.2.0-rc.2 | Core tag and public cold install verified |
| github.com/skill215/diameterkit/adapters/sctp | v0.2.0-rc.1 | Candidate pinned to core rc.2 and checkout verified |

## Version and compatibility rules

Versions need not advance together. Compatible corrections may use a patch;
new API or intentional v0 breaking behavior requires a minor increment.
Changed candidates use a new rc number. Never move, overwrite or reuse a
published tag; retract a bad version in a newer release while retaining its
content. v1 requires a separate readiness and support review.

The core minor step includes prior diagnostic/Redirect API and behavior changes,
generator v23, operations v9 and profile digest changes through combined slice.15.
The adapter minor step adopts the exact new-core dependency contract; its Go API
and implementation are unchanged. Upstream SCTP remains pinned to v0.8.0.

Review exported signatures, fields, errors, ownership, cancellation, resource
behavior and migration needs. The [Linux/Windows API snapshot](../spec/release/public-api.json)
checks exact package/source/documentation drift; it is not a semantic
compatibility proof and CI never auto-refreshes it. Record the changed module,
old/new behavior, compatibility classification, version and evidence.

For `validation.Context`, the snapshot includes the exported methods of both
the value and pointer on each target. It resolves the public alias through Go
compiler export data. The checkout gate compiles a consumer in a separate
module with `GOWORK=off` and checks that adding a pointer method to the alias
target causes drift. The immutable-install gate compares the downloaded
module against the same method sets.

Regenerate external bindings. Keep original builds for earlier recovery records;
never relabel those records or bypass Bundle fingerprint checks. Use keyed
struct literals for extended error-answer fields. Application migration, durable
intent, worker fencing and reconciliation stay outside the library.

## Required publication order

The [RC checklist](release-candidate-checklist.md) defines all required gates and
the separate full-release completion conditions for the 24 SPEC IDs.

1. Review the release scope and migration record; run exact-candidate checkout CI.
2. Publish the immutable core tag and verify its authenticated public download,
   tag commit and independent core-only consumer.
3. Change the adapter to require exactly that published core, regenerate real
   checksums and pass independent GOWORK=off/no-replace checks.
4. Run adapter-candidate checkout CI; publish its separate immutable tag.
5. Pass local and hosted independent cold consumers for the exact pair, then
   record commits, checksums, manifests, API, tests and failures.

Core and adapter tags may point to different commits. Never amend the core tag
to incorporate later adapter/docs changes. Checkout workspaces cannot stand in
for published installation evidence.

## Reproducing a published pair

The current checkout may advance its API/profile fingerprints after publication.
For the verified v0.2.0-rc.1 pair, use release-record commit
`6897bebc71e6ebd291b0b15573567a17b56a8259` before running its install gate.
Core v0.2.0-rc.2 has its own installation receipt. Pair acceptance follows publication of the rc.2 adapter.

## Independent cold installation

Use the pinned Go toolchain, Python3, Linux/amd64 with SCTP and a compiler for the
race instrumentation. From the final release-record checkout choose a nonexistent
output directory:

```sh
python3 spec/tools/release_install.py \
  --core-version v0.2.0-rc.1 \
  --adapter-version v0.2.0-rc.1 \
  --output .cache/published-install-v0.2.0-rc.1
```

The gate never publishes or edits module manifests. It rejects branch/latest,
pseudo/offline versions, unsupported v1 inputs, malformed versions and reused
output directories. Failed downloads are failures, not skips. A retry after a
publication-availability or network correction starts fresh and preserves the earlier failure.

Core-only and core+adapter consumers each start outside the checkout with empty
private GOMODCACHE, GOCACHE and GOPATH. Inherited workspace/private-proxy/checksum
exemptions and architecture flags are removed. Acquisition uses GOWORK=off,
GOTOOLCHAIN=local, CGO=0, proxy.golang.org and sum.golang.org without local/direct
fallback. Every condition must pass:

- Exact versions, authenticated Sum/GoModSum and remote tags matching download
  origins; recheck tags after execution.
- No replace/exclude in consumers, downloaded manifests or graph; no MVS upgrade.
  Adapter requires exactly the tested core and upstream SCTP.
- Linux/Windows public inventory/source hashes/GoDoc match the reviewed API.
- Public binding and Ro TCP/recovery consumers pass pure/race with no skips;
  adapter consumption includes real bidirectional SCTP.
- Library sources come only from downloaded modules. After acquisition, Go
  fetching is disabled, cgo dependencies are rejected and go mod verify passes
  without manifest/checksum changes.

Race may enable CGO for the Go detector; product modules remain pure Go.
Build caches and acceptance logs are not a library persistence facility.
The [installation workflow](../.github/workflows/release-install.yml) runs on
adapter tags or explicit manual versions, disables setup-go caching and retains
success/failure artifacts. Full protocol, target and production-capacity claims
retain the separate evidence gates in the checklist.

## Benchmarks

See the dated [README baseline](../README.md#benchmarks) and current checkout
receipts. Installation success does not establish production capacity or latency.
