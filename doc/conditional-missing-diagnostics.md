<!--
SPDX-AI-Disclosure: ai-generated
SPDX-AI-Model: GPT-6
SPDX-AI-Provider: OpenAI
SPDX-AI-Scope: Selected conditional omission behavior and reproducible verification.
SPDX-AI-Date: 2026-11-18
-->

# Conditional S6c omission diagnostics

Batch 108 adds a bounded 5005 diagnostic for one existing S6c SRR rule: when
`SRR-Flags` bit 0 is set, `SM-RP-SMEA` must be present. This extends error
explanation; it does not introduce a new presence rule or broaden accepted SMEA
address syntax.

The receive validator emits private evidence naming this exact rule and retaining
only the decoded flag word. Build/send validation still rejects missing fields
without producing receive evidence. The omission consumer accepts only the
reviewed top-level key (Vendor 10415, Code 3309), flags `0xc0`, active condition
and three-byte minimum. It rejects unknown rule names, altered paths, nested
uses, mixed variable-type evidence and arbitrary widths. Schema callbacks remain
trusted code; the convention is not peer authentication.

The existing selected `DecodeSMEA` syntax requires a digit count, type-of-address
and at least one packed digit byte. The diagnostic therefore contains three zero
bytes, an AVP Length of 15 and one padding byte, for 16 bytes total. These zeros
are a minimum-size missing-field example, not a usable address. Subscriber values
are never copied into the proof or synthesized from MSISDN or SC-Address.

An absent flag or a cleared bit does not require SMEA. Present empty, short or
unsupported SMEA content remains a value error, not a missing-field proof.
Reserved receive-side flag bits retain their existing behavior. Malformed or
duplicate flag AVPs retain the earlier decoder's 5014/5009 evidence. Ambiguous
subscriber identity and bare missing-field errors do not grant a guessed quote.

Automatic network replies require `BaseErrorFailures`, a valid return envelope
and sufficient field/count/depth/frame/reply budgets. TCP and TLS tests cover
normal and emergency slots, exact diagnostic bytes, Session/Proxy preservation,
no invalid business delivery, held resource charges, subsequent valid requests
and retirement. `CloseOnFailure` still closes without an automatic reply.

```sh
CGO_ENABLED=0 go test . ./app/s6c ./internal/missingavp -run 'Test(Conditional|RejectConditional)' -count=20
CGO_ENABLED=1 go test -race . ./app/s6c ./internal/missingavp -run 'Test(Conditional|RejectConditional)' -count=20
CGO_ENABLED=0 go test ./internal/missingavp -run '^$' -fuzz '^FuzzConditionalMinimum$' -fuzztime=30s -parallel=2 -timeout=2m
python3 spec/tools/release_check.py --suite all
```

S6c advances to slice `0.1.0-slice.23`. Its definition digest and any Bundle
containing it change; `sgdalert` also incorporates the delegated S6c digest.
Generator v22, operations v9 and public type layouts remain unchanged. Old
recovery records require their original build and must not be relabeled.
Published versions remain immutable; a new core/adapter pair and cold-install
receipt are separate release work.

Other conditional groups, combined subscription cardinalities and SCAP1 ordered
service parameters remain outside this selected diagnostic. The library stays
pure Go and introduces no persistence.
