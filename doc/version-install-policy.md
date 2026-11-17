# Version, API and immutable installation policy

The [machine policy](../spec/release/version-policy.json) applies to two independently
versioned modules. The project is in v0 development; this is not a stable-v1 or
full protocol-conformance promise. See the [publication record](releases/v0.1.0-rc.md)
for available tags, exact commits, checksums and acceptance results.

| Module | Git tag format | Candidate |
| --- | --- | --- |
| github.com/skill215/diameterkit | v0.MINOR.PATCH[-rc.N] | v0.1.0-rc.2 |
| github.com/skill215/diameterkit/adapters/sctp | adapters/sctp/v0.MINOR.PATCH[-rc.N] | v0.1.0-rc.1 |

Versions need not advance together. The adapter must require the exact tested,
already-published core and pinned upstream SCTP v0.8.0. An MVS upgrade to a different
core is a gate failure. Offline batch candidates are historical evidence only.

## Compatibility and API review

Patch releases contain compatible corrections. Under this project's v0 policy,
new API or intentional breaking changes require a minor increment and migration
notes. Changed release candidates increment rc.N. Never move, overwrite or reuse
a published tag. Retract a bad version in a subsequent release while preserving
its original content. v1 requires a separate readiness and support decision.

Review every exported non-internal package, including generated declarations,
fields, interfaces, errors, ownership, cancellation and resource contracts.
Record the affected module/packages, old/new behavior or signature, compatibility
classification, version choice, migration, support limits and tests. Protocol
slice versions and specification editions are separate from module SemVer.

The Linux/Windows [API snapshot](../spec/release/public-api.json) is a review gate,
not a semantic compatibility oracle. Source/comment/toolchain drift requires
review; CI never auto-refreshes it. The rc.2 provenance-only source changes retain
the same declarations and behavior. Snapshot source hashes are reviewed accordingly.

## Reproducible installation

The unreleased batch 96 [omission extension](missing-avp-diagnostics.md) and
batch 97 [cardinality extension](excess-avp-diagnostics.md) and batch 98
[value/unsupported extension](invalid-avp-diagnostics.md) change generated
DefinitionDigest constants and Bundle fingerprints through generator v14/v15/v16. Public signatures
remain unchanged, but earlier recovery records cannot be imported under the new
Bundle. The planned next core candidate is `v0.2.0-rc.1`, with no publication in
this batch. Published core/adapter versions and their dependency pins stay intact.

Use the pinned Go version, Python 3, Linux/amd64 with SCTP, and a C compiler for
the race tool. For the recorded rc.2/rc.1 pair, run from publication checkout
`416d66fce4ef4109462582ef6f1e66a80d86a647`, whose reviewed API baseline matches
those versions. The current development baseline intentionally rejects that older
pair. Choose a nonexistent output directory:

```sh
python3 spec/tools/release_install.py \
  --core-version v0.1.0-rc.2 \
  --adapter-version v0.1.0-rc.1 \
  --output .cache/published-install-rc2
```

The gate rejects branch/latest/pseudo-version/offline batch inputs, v1, build
metadata and malformed rc versions. It never publishes or edits repository module
files. Failed downloads are failures, not skips. A previous output directory
cannot be reused to retain stale green artifacts.

Two consumers are created outside the checkout: core-only, then core+adapter.
Each has independent initially empty GOMODCACHE, GOCACHE and GOPATH. Inherited
workspace, GOENV, private proxy/checksum exemptions, vendor flags and cross-target
settings are removed. Acquisition uses GOWORK=off, GOTOOLCHAIN=local, CGO=0,
`https://proxy.golang.org` and `sum.golang.org`, with no local/direct fallback.

Every condition must pass:

1. Download exact versions with authenticated Sum/GoModSum. Resolve remote tags
to commits; require matching Origin repository/hash and, when provided, tag Ref.
Go's Origin.Ref is optional. Recheck tag commits after testing.
2. Reject replace/exclude in consumer and downloaded manifests. Core has no
external dependency; adapter requires precisely core and upstream SCTP. All graph
directories must be in this consumer's cache, with no MVS upgrade.
3. Match the installed public package/file inventory, source hashes and GoDoc to
the reviewed snapshot on both build selections.
4. Copy only consumer tests: public binding contracts and Ro TCP/recovery examples,
plus bidirectional SCTP frames for the adapter consumer. Execute pure/race tests
with no skips. Library implementation must come from the downloaded modules.
5. After acquisition, disable Go network fetching, reject cgo dependencies and
checkout source paths, run go mod verify and ensure tests do not alter manifests/sums.

The final independent GitHub tag check still uses the network. Race needs CGO=1
for instrumentation; production modules remain pure Go. Build caches and evidence
logs do not create a library persistence boundary.

## Publication order and CI

Publish the reviewed immutable core first. Download and verify it, then update the
adapter's core requirement and regenerate real checksums. Recheck both modules,
publish the adapter's separate tag, and run the cold consumers above. The core and
adapter tags may point to different commits. Never amend a published core tag to
add later adapter/docs changes.

The [install workflow](../.github/workflows/release-install.yml) runs on adapter tags or manual
explicit versions, disables setup-go caching, uses read-only permissions and
uploads failure evidence too. Checkout CI and local probe rehearsals are distinct
from installed-release acceptance. A proxy may temporarily cache pre-publication
404s; preserve those failures and retry a fresh consumer after availability changes.

Module/tag rules follow the [Go Modules Reference](https://go.dev/ref/mod#vcs-version)
and [module authentication](https://go.dev/ref/mod#authenticating).

## Benchmarks

Release measurements are summarized at the end of the [README](../README.md#benchmarks).
They do not turn successful installation into a production capacity or latency guarantee.
