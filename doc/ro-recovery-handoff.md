# Application-owned Ro handoff and recovery

The existing recovery APIs cover named CCR profiles with explicit targets,
identity, evidence, deadlines and cumulative limits. SCAP1 remains SingleAttempt;
this does not implement general CC-Session-Failover or a durable charging ledger.
See the [public examples](../examples/ro/README.md),
[handoff examples](../ro_handoff_example_test.go) and
[TCP/TLS tests](../ro_handoff_test.go).

## Ownership and intent

PrepareRoRecoveryRequest freezes the selected request and original identity; it
neither sends nor creates a completion. The application owns the record and serial
business decisions. If durable intent is needed, the application stores the exact
record, plan and full reserved allowance, including import declarations/checkpoint
when applicable, before explicitly calling SubmitRoRecovery.

The example receipt/digest binds confirmation to the entire intent. It is a test
of exact-intent association, not a recommended storage format, authentication
mechanism or transaction protocol. In-memory applications need not invent a
storage acknowledgement layer. The library has no storage callback or worker.

Successful admission yields a RequestID and one completion. Rejection sends
nothing and yields no completion. If the application cannot durably establish
whether admission occurred, absence of a saved RequestID is not proof of rejection.
Reserve the full segment allowance up front; unused allowance is not refunded.
Consume the completion and preserve its actual answer, report, delivery evidence
and Checkpoint. A terminal outcome cannot be relabeled unknown to enable replay.

## Unknown state and import

Unknown storage commit before sending must be resolved in the application's own
storage protocol; do not submit without the required confirmation. After confirmed
intent with unknown admission/write outcome, isolate the old sender and account
for the entire reserved segment as possibly used. Preserve conservative uncertainty,
may-have-reached-peer and replay-pending facts; do not roll back cumulative limits.

Import requires explicit old-sender isolation, trustworthy historical facts and
time evidence. A new Node cannot detect fabricated prior progress or a deliberately
erased terminal result. Known success or permanent failure stays terminal across
Nodes. These limitations cannot be repaired by assuming a larger in-memory cache.

## Business time, protocol time and late answers

Business Tx and per-attempt ResponseWait/Tr are separate. Keep the original
business deadline and maximum recovery age across segments. A business timeout
does not complete the protocol exchange or authorize a retry before its guards.
The application can stop business waiting while continuing to consume the single
completion, or explicitly Cancel and consume cancellation. Late answers do not
create a second business notification or completion.

Error continuation preserves the original deadline and named result/target
policy. CanContinue describes protocol eligibility, not a decision to charge,
release traffic, compensate or retry. Remote DDFH is usable only in the applicable
validated normal debit answer; absence, explicit zero and explicit one differ.
No-answer, E envelopes and invalid answers cannot invent a remote value. Balance,
refund, ECUR and SCAP1 do not inherit a debit-specific business branch.

## Boundary and evidence

Public consumption tests exercise standard debit, named compatibility mappings,
late answers, explicit cross-target/no-answer continuation and import refusals.
Original identity, owned bytes, target authorization, writer retirement, budgets
and deadlines remain enforced. A local synthetic OCS does not validate private
PSP/token namespaces, target deduplication retention, live ECUR continuity or
financial settlement. Those are independent application/deployment evidence gates.
