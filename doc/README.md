# Documentation

Start with the [project README](../README.md) for installation and runnable examples.
The guides below describe the current development checkout. Published candidates
have their own [release record](releases/v0.1.0-rc.md).

## Build an integration

- [Runtime, resource ownership and request outcomes](runtime-basics.md)
- [Supported protocols and platforms](diameterkit-support-matrix.md)
- [Versioned protocol definitions and generation](diameterkit-application-definitions.md)
- [S6c request construction and result interpretation](s6c-routing-helpers.md)
- [Application-owned Ro recovery and handoff](ro-recovery-handoff.md)
- [SCTP installation and connection setup](../adapters/sctp/README.md)
- [Independent public binding contracts](../examples/contracts/README.md)

## Diagnose and verify behavior

| Topic | Guides |
| --- | --- |
| Missing fields | [Fixed-width 5005](missing-avp-diagnostics.md), [minimum Grouped examples](minimum-group-diagnostics.md), [variable-field minima](variable-missing-diagnostics.md) |
| Invalid fields | [First-excess 5009](excess-avp-diagnostics.md), [5004/5001](invalid-avp-diagnostics.md), [typed-length 5014](length-avp-diagnostics.md) |
| Common groups | [Handwritten Proxy/VSAI](manual-group-diagnostics.md), [VSAI conditions and Proxy omissions](group-constraint-diagnostics.md) |
| Redirects | [Bounded 3006 and URI validation](redirect-handling.md) |
| Volatile duplicates | [Combined binding](combined-duplicate-acceptance.md), [cross-peer and mixed-profile budgets](cross-peer-duplicate-acceptance.md) |

## Release and contribute

- [Version/API policy and immutable installation](version-install-policy.md)
- [Published candidate record](releases/v0.1.0-rc.md)
- [Current remaining gaps](remaining-gaps.md)
- [Dual-module CI and local acceptance](release-checks.md)
- [Bounded fuzz, stress and benchmark methodology](load-checks.md)
- [Specification, progress and historical evidence](../spec/README.md)
- [AI authorship and review limits](../AI_DISCLOSURE.md)
