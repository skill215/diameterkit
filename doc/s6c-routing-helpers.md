# S6c routing helpers

`app/s6c.NewSRR`, `InterpretSRA` and `client/s6c` operate on the existing bounded
profile. They add no subscriber lookup, candidate ranking, delivery, background
queue, automatic business retry or persistence.

```go
q := s6c.RoutingQuery{
    SessionID: session, // supplied by Node.NewSessionID or another controlled source
    MSISDN: avp.Some("12025550123"),
    SCAddress: "12025550100",
}
request, err := s6c.NewSRR(q)
// Handle err; offline callers validate envelope/source with the Codec first.
client, err := s6cclient.New(route)
reply, err := client.QueryRouting(ctx, q, diameter.RequestOptions{})
// Preserve reply.Exchange evidence even when err is non-nil.
```

Numbers are international decimal digits, without a plus sign or TON/NPI prefix.
Optional generated fields must not duplicate SessionID/MSISDN/SCAddress or fixed
AuthSessionState. Input slices are copied. IMS/T4/Router/MCCMNC/role contexts remain
explicit through RequestOptions.Validation and matching endpoint fields; zero
context does not authorize a nonordinary procedure. Strings do not select a role.

Interpretation checks original request/answer bodies, Session/Proxy pairing and
selected context. The caller must validate header identifiers and answer source
through the codec/runtime first. The thin client interprets only AnswerReceived
with no runtime error and complete validation. Async SubmitRouting uses the
caller's CompletionPort; consume it and call InterpretRouting. No channel or
business-event publisher is introduced.

Results preserve base/10415 experimental/E namespaces. Business failure is
RoutingRejected. Missing context or unsupported selected semantics yield
RoutingUnsupported and issues, without deliverable candidates. Invalid structure
or unknown namespaces return errors. Exchange retains the original result; the
interpreted Answer is independently owned.

Candidates retain Primary/Additional provenance. MSC+MME is one candidate with a
Companion. IP-SM-GW does not invent a Realm: ordinary forms require resolution;
SBI=1 reports SBIAccess, not an implemented SBI client. Unsupported Realm/reserved
SBI/Additional-IPSMGW shapes remain unsupported. Diagnostics preserve presence
and zero values. Request HSS-ID is not answer IMSI.

Tests cover construction/ownership, node shapes, result namespaces, named context,
real TCP query/submit/cancel and retention of original exchange evidence. See
[the runnable example](../examples/s6c/README.md) and
[batch87 evidence](../spec/evidence/batch-87-s6c-helpers.json).
