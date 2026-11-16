---
disclosure-default: ai-generated
providers:
  - OpenAI
scope: |
  Project-authored library code, tests, benchmarks, examples, tooling and
  substantial documentation were generated and revised with OpenAI Codex under
  maintainer direction. Complete historical model identities were not recorded.
last-updated: 2026-11-16
---

# AI disclosure

diameterkit follows the disclosure structure used by
[smppkit](https://github.com/skill215/smppkit/blob/main/AI_DISCLOSURE.md).
The `SPDX-AI-*` comments are provenance declarations, not certification by SPDX,
GitHub or a standards body. They do not change the Apache-2.0 license.

## Tool, provider and scope

OpenAI Codex generated and revised the implementation in response to the
maintainer's requirements, constraints and corrections. Project-owned Go files
carry disclosure, model-record status, provider, scope and date fields, including
tests, benchmarks, examples and generated declarations. The generator emits the
same declarations deterministically. Build constraints and Go documentation are
preserved.

The complete historical model inventory is **unrecorded**. The current agent is
identified as Codex based on GPT-6, but this does not establish the identity of
every model used in earlier batches. File headers explicitly report that gap;
they do not copy smppkit's separately confirmed GPT-6.0 Astra attribution.
No precise AI percentage or complete per-edit provenance history is claimed.

## Human involvement and review limits

The maintainer supplied requirements, reviewed progress and directed changes.
This does not assert an independent line-by-line human audit. AI-authored tests
and agent review may share implementation mistakes. Independent wire fixtures
are independent of the production serializer, not necessarily independently
human-authored or proof of full standards conformance.

Maintainers remain responsible for accepted code and releases. The library has
no AI runtime dependency. Current support and deployment limitations remain in
the [support matrix](doc/diameterkit-support-matrix.md).

## Existing material and future changes

The standard [LICENSE](LICENSE), external protocol sources and upstream SCTP
code are not claimed as AI-authored project work. Raw test and benchmark results
are measured tool output; their explanations may be AI-authored. Historical
evidence and already-published tags are not rewritten by this disclosure.

Keep future declarations accurate. Do not attribute externally supplied or
human-authored files to AI merely by copying adjacent headers. Record known
models when evidence exists; preserve uncertainty when it does not.
