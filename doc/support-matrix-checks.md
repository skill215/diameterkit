<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Release metadata consistency and reproducible acceptance documentation.
SPDX-AI-Date: 2026-11-18
-->

# Support matrix consistency

The current [support matrix](../spec/release/support-matrix.json) must describe
the checked-out definitions. Batch 107 corrects eight stale definition hashes
and adds an executable gate for all eleven generated bindings and the delegated
`sgdalert` binding.

Run these commands from the repository root:

```sh
python3 spec/tools/check_support_matrix.py
python3 -m unittest discover -s spec/tools -p 'test_release*.py' -v
python3 spec/tools/release_check.py --suite static
```

The standalone check needs only Python's standard library. The static suite
also needs the pinned Go toolchain and module dependencies. Both local static
acceptance and the hosted `checkout-static` job execute the check.

The gate verifies the exact binding inventory, reviewed repository-relative
definition paths, SHA-256 of the definition file bytes, profile names, versions,
ordered request identities and aggregate counts. Duplicate JSON keys, missing
or duplicate bindings, missing files and paths escaping the repository fail.
Generated metadata comes from JSON definitions. The handwritten `sgdalert`
metadata uses narrowly recognized literal Go declarations, including matching
request/answer command identities; changes to these declaration shapes require
review of the extractor. This is not a general Go parser or proof of delegated
runtime semantics. Its S6c dependency is independently checked as a matrix row.

The gate is check-only: it never refreshes hashes or repairs a baseline. Tests
change definitions, hashes, identities, inventories and paths in temporary
fixtures. Static acceptance also executes the real CLI against a stale matrix
and verifies a nonzero exit without rewriting it. The checked-in matrix is
protected against mutation throughout the release suite.

Raw file hashes are distinct from generated `DefinitionDigest` values, Bundle
fingerprints, API snapshots and immutable published-package checksums. Updating
the current matrix does not rewrite historical receipts or establish real-target
interoperability, release installation or production capacity.
