# S6c 路由便利层

本批固定使用当前 `app/s6c.Profile()` 的有界切片，新增 `NewSRR`、`InterpretSRA` 和 `client/s6c`。无持久化、订户查询、候选优先级、投递、后台队列或自动业务重试。

```go
q := s6c.RoutingQuery{
    SessionID: session, // 调用者通过 Node.NewSessionID 或自己的受控来源提供
    MSISDN: avp.Some("12025550123"),
    SCAddress: "12025550100",
}
r, err := s6c.NewSRR(q)
// 离线使用：Codec.CompleteRequest / ValidateAnswerContext 后调用 InterpretSRA。
c, err := s6cclient.New(route)
reply, err := c.QueryRouting(ctx, q, diameter.RequestOptions{})
// 无论 err 是否为空，都保留 reply.Exchange 的发送证据和原答案。
```

号码是国际十进制数字，不接受 `+`、TON/NPI 或非法字符。`Fields` 提供生成 SRR 的可选字段；禁止与顶层 SessionID/MSISDN/SCAddress/固定 AuthSessionState 重复。T4 的 IMSI 与 IMS 的 HSS-ID/目的 URI 使用原有 UserName/SMSMICorrelationID 字段；不会从字符串猜用途。`Fields` 的切片由构造器复制。

具名 T4/IMS/Router/MCCMNC/RoutingRole 上下文仍通过 `RequestOptions.Validation` 显式提供；同时在 `Fields` 提供与证据一致的 Origin/Destination，以供纯解释器复核。普通请求可由运行时补齐端点。零上下文不能授权非普通用途。当前接口绑定固定 Profile，不接受任意 Profile 或扩大扩展语义。

解释器接收原始请求和答案 body；它检查体结构、Session/Proxy、原请求成对条件和具名上下文。调用者必须先用 codec 或运行时检查头标识和应答来源。薄客户端仅解释 `AnswerReceived`、无运行时错误且 `Validation.Complete` 的结果。异步路径使用 `SubmitRouting` 与调用者原有 CompletionPort，取回结果后调用 `InterpretRouting`；不新增通道、不接管发布业务事件。

`RoutingResult` 区分基础、10415 实验、E 答案，保留 code/vendor；业务失败为 `RoutingRejected`。上下文不足或所选接收边界不支持时返回 `RoutingUnsupported` 与 Issues，不提供可投递候选；结构错误、未知命名空间等返回错误。原答案仍在 Exchange，成功解释的 Answer 是独立副本。

候选保留 Primary/Additional 来源与十进制号码。MSC+MME 是带 Companion 的单一候选。IP-SM-GW 不伪造 Realm：普通形态为 NeedsResolution，SBI=1 为 SBIAccess，不表示本库实现 SBI。Realm/保留 SBI、Additional IP-SM-GW 按现有 Profile 报告未支持。三域诊断与 MWDStatus 保留 presence/零值；原始生成答案保留 LMSI、Alert identity、未知 raw 等字段。请求中的 HSS-ID 不会被当成应答 IMSI。

`Client.New` 检查非 nil Route；命令/Profile/目标绑定由运行时提交时校验。QueryRouting 使用运行时同步 Request 所需的完成资源；SubmitRouting 使用调用者的现有端口，端口容量仍由 Node 预算管理。

验收覆盖节点形态、组合/Additional、命名空间/诊断/缺上下文、非法输入、所有权、IMS/T4/Router/MCCMNC、真实 TCP Query/Submit/取消、保留原始 Result，以及全部核心回归。公开发行与真实目标部署仍为独立未完成项。
