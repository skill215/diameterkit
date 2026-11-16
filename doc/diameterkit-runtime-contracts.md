# Diameter 运行时契约：故障转移、分阶段校验与预算转移

> 第68批更新：Relay-only协商和有限逐请求目标表已实现并验收，见[batch68](../spec/history/batch-68.json)。后文第67批扫描缺口为当时状态；目标表之外的动态发现和其他Ro/P4/发布缺口未由本批关闭。

> 第67批当前状态：独立adapter已锁定正式SCTP v0.8.0（bce1b3cf3ea3），源码与第66批候选一致，A–E、资源和Peer边界不变；正式版本回归完成，见[batch67](../spec/history/batch-67.json)。上游发布门已关闭，核心/adapter公开发布及现场验收仍单列；下文旧候选/未合并表述按历史时点阅读。

> 实施注（2026-11-07，第四十六批）：SCAP1显式SCAP1Accounting绑定、严格193/Acct19302能力及App0/271普通预算/pending路径已接通，真实TCP/TLS验证SingleAttempt、双向能力、正常/E、取消晚答、关闭和重连。旧双Auth/Acct广告仍拒绝，实际私有ACA与现场OCS未验；不自动重放。见[batch46](../spec/history/batch-46.json)和[progress](../spec/progress.json)。下方旧批记录保留历史时点。

> **当前范围（2026-11-04，Draft 0.20）：** 第一阶段=首版/v1，包含Sh、S6c、SGd/Gdd、Ro/DCCA及sms-fw实际所需六模板、SCAP1和RO-06恢复能力；Go 1.26、纯Go实现（含依赖与SCTP适配器，发布路径CGO_ENABLED=0）、公开Go库、无持久化边界不变。明确撤销第十九轮“Ro后置”分期；带日期的旧分期/统计仅作历史证据，不再授予Ro跳过首版验收的例外。限定S6a、完整SCUR/PRICE_ENQUIRY等未承诺能力仍后置。范围及剩余缺口见[Specification §3/§7](diameterkit-specification.md)，顺序见[总设计§14](diameterkit-design.md)，验收见[覆盖§39](diameterkit-usage-coverage.md)。2026-11-06已完成前十九批有界实施（Base/S6c基础、公共包络与关联、出站运行时、关闭与选举、TCP/TLS生产适配、有限Inbound/应答、有限Reject/应急预算、Canonical/有限重复缓存及多Arrival重答、固定同终端Route/多attempt与活动请求原身份重传、有限Dialer/Maintain拨号及重连、按class/编码字节DRR调度、固定Peer的有限Listener/Serve、有界标识源/Session/guard、Base/S6c有限Context/Report、S6c普通RDR/RDA及HSS ALR/ALA、SGd/Gdd普通OFR/OFA与TFR/TFA、Sh普通四组命令及透明User-Data、Ref24/25有限Sh-Data helper、Sh单订阅/退订及Ref11/12有限数据）；已执行范围及未完成项以[Specification §6](diameterkit-specification.md#6-需求与证据的维护)和[spec进度](../spec/progress.json)为准，未列明的运行验收仍Planned。

> **第十四批实际落地边界：** 当前Base/S6c仅零Context；codec.ValidateContext/ValidateAnswerContext、Inbound.Validation/Check及Result.Validation已接入。Report是限定切片的阶段级/首失败报告，NoContext未审核不视作缺facts。下文Sh/Ro工厂、非零输入和完整NeedsApplication转换仍为目标契约，随对应应用实施；详见[batch-14](../spec/history/batch-14.json)。
> **第十三批实施注记（2026-11-06）：** 已接通有界IdentifierConfig/EndToEndSource、NewSessionID、内置分区轮次guard、自定义class配额及固定跟踪表，原身份重传共用ID，HBH保留区触发连接排空轮换。实际API、预算和关闭边界见[batch-13](../spec/history/batch-13.json)，检查见[spec进度](../spec/progress.json)。跨进程唯一性、RO-06恢复导入、具名上下文/报告、完整四应用及SCTP仍沿原边界，未由本批宣称完成。

日期：2026-10-31（第二十一轮补号码/订阅与特殊身份切片）。状态：第四轮建立、第十轮收敛 Peer/重传/标识、第十一轮补固定 Profile/输入与接收语义、第十二轮补应用字段和结果限制、第十三轮接入库级 specification、第十四轮同步身份/结果细化索引，仅文档，没有实现或验证结果。本文是 R19–R21 的契约来源，取代早期文档中“普通 Route 永不自动重传”及这些问题的待决表述；函数拼写和数值默认值仍须 v0.x 调用原型验证。范围为 Go 1.26+ 端点库，不增加跨语言桥、DRA 产品或业务账本。

第八轮 [协议上下文、Peer 与进程内标识契约](diameterkit-state-consistency.md) 撤销 Store/Resolve、managed charging、持久标识 provider 以及公开 Stage 流程；本页只规定运行中 Node 的协议能力。资源“账本”只是内存额度计数，不是持久账本。

第五轮补充：[部署与互操作契约](diameterkit-deployment-interop.md) 固定 R03/R17/R18；其逻辑 Inbound/逐连接 Arrival 和 C 计量规则同时修订本文 §2.3/§3.4，其他 R19–R21 不变。

关联：[总体设计](diameterkit-design.md)、[API 与流程](diameterkit-api-workflows.md)、[覆盖基线](diameterkit-usage-coverage.md)、[评审记录](diameterkit-design-review.md)。协议的逐字段规则、实际 DRA/终端拓扑、底层 SCTP 能力仍须分别验证；本轮收敛的是运行时模型及其可测试行为。

第十八轮更新的库级需求与验收索引见 [Library Specification](diameterkit-specification.md)；本文保留状态、时序和预算细节，不因规范草案建立而视为实现通过。主流程§10的服务上下文/单位/结果检查沿用Wire、Exchange、Application分层；候选报文不表示已跑decoder，业务终止/重新计费不因Application失败而自动执行。

2026-11-06代码归属见[实施架构§4–5](diameterkit-architecture.md)：root拥有运行时状态和单账本转移，internal/codecengine共享离线/网络编解码；transport/adapters不接触reservation。RO-06的Node关联准备/接纳入口归root，纯协议规则由Profile传入，app/ro不导入Node。不改变本文的阶段、结果或预算算法。

## 1. R19：唯一调度者与有界故障转移

### 1.1 调用入口与策略

RouteConfig 新增 ExchangePolicy，必须显式选择，零值配置报错；RequestOptions 不能逐次放宽该策略。Profile 描述各命令允许的恢复类别，构造期检查策略与 Profile 相容。

| 模式 / 入口 | 调度所有者 | 确定行为 |
| --- | --- | --- |
| SingleAttempt / Route.Request、Submit | 普通 exchange 引擎 | 选择一次就绪 Peer，接纳后不切换；用于无合法备用路径或需要应用自行决定的用例，明确不宣称已覆盖 failover |
| TransportFailover / Route.Request、Submit | 普通 exchange 引擎 | 仅按 §1.3 的触发条件切换邻接；一个 RequestID、固定逻辑身份、有界 attempts、一个最终 Completion |
| ExplicitRetransmit / Route.Submit + Node.Retransmit（首版Ro版本） | 同一 exchange 引擎，应用决定是否继续 | 只在库确认的可重传失败后接纳下一 attempt；固定身份/期限/次数与唯一 Completion，不同时自动恢复；详见 §1.6 |
| Peer.Request、Submit | 普通 exchange 引擎 | 永远 SingleAttempt；仍执行 CommandPolicy.OutboundPolicy 的目的地、应答来源、facts 和 class 约束，不会因某个 Route 的配置而隐式跳到其他 Peer |

首版同时实现SingleAttempt、经逐命令审核的TransportFailover和Ro显式模式所需RequestInfo/Retransmit/AwaitingRetransmit。普通Ro CCR仅对固定OCS启用自动/显式传输恢复，详见§1.6；SCAP1及跨目标/跨段恢复只能按§1.7的独立审核规则接纳，不随标准CCR许可自动启用。库不调度计费业务会话迁移、持久缓冲或补偿；提供必要协议入口不等于完整DCCA业务状态机。

普通Route的TransportFailover/ExplicitRetransmit配置含MaxAttempts（含首次，至少2；RO-06分段例外见§1.7.2）、AllowedPeers、ReplayDomain、MaxReplayAge。ReplayDomain 记录固定 App/Profile、允许的目的地范围、固定终端或共享去重的终端集合、已验证去重期限与证据引用；它是部署契约，不是网络协商，也不能仅用一个任意字符串证明安全。具名标准 Profile 标记可参与传输恢复的命令，运营商/未知扩展命令默认不可参与。普通 Sh/SGd/S6c 命令仍需逐项规则及样本，不能因它们不计费就假定无副作用。

AllowedPeers 必须是 Route.Peers 的子集，首次选择也受该集合/域约束；不会先发给域外 Peer 再尝试把它恢复到域内。策略、class 和候选 PeerID 在接纳时固定，候选的当前连接状态在每次发送前核对。

带 Destination-Host 的请求保持该目标；允许更换通向它的 DRA，不允许改成另一个最终 Host。realm-only 请求只有在该范围内的所有潜在终端具备已验证的重复识别边界时才能启用恢复。路由配置校验每次 Destination override；超出范围在 admission 前失败。自动模式配置的候选路径不足两条、Profile 不相容、期限超过去重边界或规则未支持时拒绝 TransportFailover 配置，而不是静默降级。

### 1.2 逻辑 exchange 与 attempt

逻辑 exchange 保存 RequestID、CorrelationID、Origin、Session-Id（若有）、End-to-End、App/Command、Destination、ProfileFingerprint、原业务内容和协议上下文快照、PolicySnapshot、绝对 Deadline。它们在 admission 后不可修改；Deadline 取调用期限、Route 上限及首次接纳时间加 MaxReplayAge 的最早值。重试不续期，不重新生成事件时间或业务字段。

每个 attempt 独立保存 AttemptID、PeerID、ConnectionID、Hop-by-Hop、T 位、开始时间、写入证据、终止原因及应答摘要。attempt 在 writer 接纳前注册；writer 接纳即消耗一次 MaxAttempts，即使随后确认未写出。尚未注册/接纳的候选扫描不消耗次数，但一次选择仅遍历有限 AllowedPeers。每个连接代次最多被本 exchange 选中一次。

冻结的请求编码主体可以复用，当前 attempt 的头单独生成；前一次 writer 尚未结束时不得修改其输入或开始下一次写入。首次发送 T=0；已发送、部分写入或无法证明未发送且尚无应答的重传按规则设 T=1；只有从未开始发送的本地排队重选仍为 T=0。收到合法错误应答后不进入此类恢复，不能沿用 T=1 盲目重发。基础依据：[RFC 6733 §3](https://www.rfc-editor.org/rfc/rfc6733.html#section-3)、[§5.5.4](https://www.rfc-editor.org/rfc/rfc6733.html#section-5.5.4)。

### 1.3 触发、状态与选择

| 事件 | 普通 TransportFailover 的动作 |
| --- | --- |
| watchdog 进入 SUSPECT 或 Peer 确认传输不可用、连接 reset/EOF、写入终止性错误或写期限导致连接关闭 | 先取得当前 writer 的最终证据，再在策略范围内选择当前就绪备用；不等待重连、不进行平行 hedging |
| 已接纳但尚未写入时连接失效 | 保存 NotWritten，允许转移；没有备用则终结 |
| 仅 exchange Deadline 到期、Node.Cancel、Port.Close、Node/Peer 主动关闭或 shutdown 排空到期 | 停止所有后续发送；取消不是恢复触发；保留此前发送证据 |
| 本地配置、校验、资源超限或协议违规导致断连 | 不自动换 Peer；错误不能伪装为普通传输失败触发循环 |
| 已收到相关且可接受的成功、业务拒绝或 E 位错误应答 | 终结普通 exchange；TOO_BUSY、UNABLE_TO_DELIVER、重定向均原样交付，不自动延时重试或重新发现目标 |
| 错误关联头 | 不归属该 exchange；按协议异常策略丢弃/关闭，不把它当作业务失败应答 |

状态顺序为 Admitted → AttemptActive → SelectingAlternate → AttemptActive，或进入 Completing → Completed。SelectingAlternate 同时检查 deadline、取消、候选能力、ReplayDomain 和资源；从轮询/主备策略给出的有限候选中选取。没有就绪合法目标、次数耗尽或所有候选资源不足时立即结束，分别记录 NoEligibleAlternate、AttemptsExhausted、FailoverBackpressure；不后台无限等待。已有接纳不保证备用 Peer 的窗口永远可用。

普通故障转移的触发理由来自结构化终止原因，不能只检查 errors.Is(err, net.ErrClosed)。恢复与新请求选路都复用同一 Peer 状态快照，不能因队列满强行发送到 Suspect/Closing 连接。控制协议有自己的状态机，CER/DWR/DPR 不进入普通应用重传。

### 1.4 应答、取消与一次完成

每个逻辑 exchange 有一个串行提交门，仅串行化状态变更和 Send 接纳点，不在锁内等待网络或业务。旧 attempt 的匹配记录保留到逻辑 exchange 终结或原连接关闭；新旧 attempt 各有独立应答预算。即使旧连接已判不可用，已经读完并送达提交门的有效应答仍可赢得完成。

1. reader 完成帧/静态/关联及可得上下文检查后提交 AnswerCandidate；需要应用状态的应答携带 ValidationReport，不因此触发重传。
2. 第一个被提交门接受的 AnswerCandidate 固定选中结果，阻止尚未接纳的下一次 Send。成功与失败不按“优先成功”重排；它们都是有效终端结果。
3. 若下一 attempt 已接纳，取消其尚未写入部分；已开始写入的路径必须取得最终写入证据，必要时关闭其连接。普通完成可等待该本地清理，但不得超过受控写入/关闭期限；保留已选应答，不能被后来写失败覆盖。
4. Cancel/Deadline 先进入门时禁止新 Send；若尚未选中应答，则完成为取消/超时。门前已选中的应答不被取消降级。后续到达的普通应答只计数，不产生第二次 Completion，也不改变已交付对象。

一个 exchange 只交付一次最终 Completion；清理期间观察到额外冲突可计数，不承诺交付全部重复/晚到结果。Ro 适用同样规则。业务等待期可短于 Submit 的 exchange Deadline；只要未取消/完成，后续有效 CCA 仍可交付，详见状态与一致性契约 §2。

Result 新增有界 Attempts 摘要、SelectedAttempt、StopReason 和 DeliveryEvidence。各次 attempt 的错误和字节进度分别保留，不把它们相加成一帧。Outcome 取选中应答则 AnswerReceived；否则任一确认完整写出为 WrittenNoAnswer，再否则有不确定写入为 WriteUncertain，再否则有部分写出为 PartlyWritten；仅全部有证据未写出才为 NotWritten。DeliveryEvidence.MayHaveReachedPeer 对任意 Written/Partial/Unknown 都为 true。最后一次 NotWritten 绝不能覆盖之前的风险证据。

### 1.5 有界传输恢复与应用恢复的区别

例：S6c SRR 经 DRA-A 完整写出后连接失效，DRA-B 在同一 ReplayDomain 且有预算；内部新 attempt 使用新 HBH、相同 E2E/Session/业务内容和 T=1，SRA 完成原 RequestID。第二次未写出而整体期限耗尽，结果仍保留 WrittenNoAnswer。普通错误应答、期限和取消不启动额外重试。

普通Ro CCR按§1.6所选模式自动切同终端路径、等待应用显式重传或直接返回发送证据；再次Request的新exchange不能当成原操作恢复。RO-06的原请求准备/受检续发另见§1.7：只交付可由应用保存的协议数据，不代办存储提交、不导入旧进程pending。

每个exchange的协议调度只发生在当前Node和固定deadline内；重连只恢复连接可用性。§1.7允许应用在原总恢复上限内创建新段，但不复活已终结请求或自动恢复外部业务记录。

### 1.6 R33：活动 exchange 内的 Ro 原身份重传（首版契约）

> 第九批已用普通S6c SRR验证本节的通用活动请求调度/查询/幂等重传机制，实际入口见[batch-09](../spec/history/batch-09.json)。第五十批已为标准Ro及三个兼容Profile的已选CCR实现本节许可和真实TCP/TLS验收，见[batch-50](../spec/history/batch-50.json)；§1.7的原记录/跨段/重启/Tr等专用入口仍未实现，目标去重域仍待部署证据。

**范围修订（2026-11-04）：** 本节继续规定普通 Route 的受限模式；§1.7另补RO-06专用的响应等待到期、跨OCS及应用保存原请求后的续发。下文“不导入/不跨OCS/完成后不再发”均指本节入口，不能再用来排除§1.7已明确的必需能力。旧RequestID始终不可复活。

第十轮撤回“Ro 一律禁止 TransportFailover”的过宽限制，收敛为**同一终端、同一活动 exchange、无应答的传输恢复**。新增 ExplicitRetransmit 供应用显式决定重传时点，和自动 TransportFailover 二选一；两者共用原 exchange、预算、调度门和唯一 Completion。没有新的持久操作、ReplayHandle 导入或业务恢复引擎。

#### 1.6.1 命令与目标范围

Ro 的两种重传模式都要求首次发送前具有明确 Destination-Host/Realm，绑定固定 OCS 身份；AllowedPeers 只是通向该 OCS 的已配置邻接路径。DRA 的 relay 广告、相同 realm 或共享负载均衡地址不能证明同终端。路由到同名集群实例时，其去重/会话一致性仍由部署提供 ReplayDomain 证据；未证明则使用 SingleAttempt。首版Ro首个交付子集不为realm-only 发送启用原身份重传，也不从重传接口改变 Destination。

| 协议操作 | 当前允许的原身份重传 | 留在应用或支持范围外 |
| --- | --- | --- |
| IEC CHECK_BALANCE | 固定 OCS，经传输失败触发；自动或显式模式 | 余额结果的业务决定；错误答案后重选 OCS 不走本接口 |
| IEC DIRECT_DEBITING | 同上；Session/编号/Requested-Action/单位/事件时间均不变 | DDFH 的放行/缓冲/补偿策略，Tx 后是否继续业务 |
| IEC REFUND_ACCOUNT | 同上，保留原退款请求及关联凭据；不是重新构造一笔退款 | 原扣费的核对、退款决定、持久缓冲及进程重启后恢复 |
| ECUR INITIAL / TERMINATION | 同一 OCS、同一 Session/CC 类型/编号的无应答请求可重传；已有会话不因改走 DRA 被视作迁移 | CCFH、CC-Session-Failover 所约束的另一 OCS 会话迁移；TERMINATION 与 ASR 的业务竞争 |
| ASR/ASA、SCAP1、未知扩展 | 首版Ro/兼容交付保持SingleAttempt；标准基础控制也不进入此调度器 | 未审计重传规则不能借 App-ID 4 或“计费”标签自动放开 |

固定 Profile 必须逐命令声明上述允许组合并有规则/fixture；所选TS32.299引用RFC4006，依据使用[RFC4006 §5.7、§6.5](https://www.rfc-editor.org/rfc/rfc4006.html#section-5.7)：传输路径恢复与换计费服务器分开，事件处理按Requested-Action区分。RFC允许的某些错误后换服务器、Tx/DDFH处理及持久缓冲没有纳入本库；本表是更窄的受限能力，不能写成规范禁止这些过程，也不能宣称完整DCCA客户端。RFC8506仅保留独立差异目标，不作为本表的替代规范依据，见[Profile §1](diameterkit-profile-rules.md)及[主流程§29](diameterkit-mainflow-audit.md)。

#### 1.6.2 公开入口与唯一调度者

```go
// 签名示意；只对本 Node 内、以 ExplicitRetransmit 模式 Submit 的请求有效。
(*Node).RequestInfo(id RequestID) (RequestInfo, error)
(*Node).Retransmit(ctx context.Context, id RequestID,
    after AttemptID) (AttemptID, error)
```

RouteConfig.ExchangePolicy 新增 ExplicitRetransmit，配置同样固定 MaxAttempts、AllowedPeers、ReplayDomain、MaxReplayAge，且不运行自动 failover；首版Ro版本仅为上表已审核的CCR组合启用，其余命令在发号/接纳前返回 ReplayUnsupported。首次 Submit 正常选路/发送，不是 Stage；同步 Request 对该模式返回 ModeRequiresSubmit，不接纳、不发号，避免调用者拿不到 RequestID 却必须控制重传。Peer.Request/Submit 仍只支持 SingleAttempt。自动 TransportFailover 的 request 不允许再调用 Retransmit，以免双重调度。

ExplicitRetransmit 可配置一个 Peer 供它重连后的新 ConnectionID 使用，也可配置多条 DRA 路径；每个连接代次只用一次。调用仅扫描**当前 Ready**候选，不负责等待重连；没有候选/预算返回 NoEligibleAlternate/Backpressure 并继续保留 AwaitingRetransmit 至原 deadline。后台 Maintain 可独立恢复连接，但连接恢复不会自动发出业务重传。自动 TransportFailover 继续要求至少两条配置路径，失败时不等待重连。

单 Peer 重连可能经历 Tc、握手及 REOPEN 的多轮 Tw，候选 30 s 请求期限通常不能覆盖全部过程。应用若选择此模式，须在首次 Submit 明确给出与 ReplayDomain 相容的窗口；库不为了等重连续期，窗口不足就正常完成超时。双 DRA 已有就绪备用不需要等待这些恢复步骤。

RequestInfo 返回有限 owned 摘要：ID、Phase、CurrentAttempt、ReplayEligible、TransportStopReason、AttemptsUsed、Deadline、DeliveryEvidence；不返回可改写的请求或开放原始标识覆盖。Phase 为 AttemptActive / AwaitingRetransmit / Completing；完成从活表移除后返回 RequestNotActive，结果只从原 CompletionPort 取得，不建完成历史数据库。业务可按自己的定时器或 PeerWatch 唤醒后查询；不增加第二条可靠事件流、不承诺每个状态边沿都有通知。

#### 1.6.3 允许触发及状态迁移

```text
Submit → AttemptActive ── 合法应答 ─────────────────→ Completing → 一个 Completion
             │                                          ↑
             └─ 可重传传输失败，writer已结清 ─→ AwaitingRetransmit
                                                │       │
                       显式 Retransmit 接纳 ──────┘       └─ Cancel/Deadline/关闭
                              ↓
                         AttemptActive
```

只有 runtime 观察到 EOF/reset、传输写终止性错误、导致断连的写期限或 watchdog SUSPECT，并结清该 attempt 的 writer 后，才能置 ReplayEligible。没有应答是必要条件，单纯应用 Tx、读取 RequestInfo、旧状态快照或传入 after 不是故障证明；尚在排队/写入/正常等答时返回 AttemptNotReplayable。这样不会增加基于任意超时的并行 hedging。协议/配置/本地资源错误、远端 DPR、本地主动关闭均直接终结，不进入 AwaitingRetransmit。

可重传失败后，若仍有次数和原 deadline，ExplicitRetransmit 保留 Q、全部所需 A/M 与活跃 E2E，等待应用决定；没有自动新 attempt，也不先交付“失败 Completion”。次数用尽则完成 AttemptsExhausted；应用不重传也不 Cancel 时，最终仍在原 Deadline 完成。普通业务 Tx 只影响应用决定，不能隐式重置库 deadline。

Retransmit 的 ctx 仅约束此次接纳，规则与 Submit 相同：提交门原子校验 mode、after、ReplayEligible、目标/资源、Node 生命周期及剩余次数，成功登记下一 AttemptID 并进入有预算的 writer 队列后返回；ctx 随后取消不取消整个 exchange。没有网络字节即可接纳成功，不代表写出；已接纳 attempt 即消耗次数，包括随后未写。拒绝不消耗次数、不生成第二个 Completion、不重新分配 E2E。

after 是预期的失败 AttemptID，不是业务幂等键。同一 after 在同一活 exchange 中最多建立一个“前次→下次”边：并发/重复调用若该边已建立，返回既有下一 AttemptID，绝不再次排队；若不是当前可重传前驱且无已建立边，返回 StaleAttempt。边只占 MaxAttempts 有界摘要；exchange 完成后不保留供将来调用的幂等服务。成功登记与 ctx 取消竞争时，登记赢必须返回成功 ID，不能把已接纳重传伪装成失败供调用者重试。

#### 1.6.4 身份、应答和关闭竞争

重传只引用库已拥有的原请求：Origin、E2E、Session、App/Command、Destination、CC 类型/编号、Requested-Action、业务/扩展内容、Profile 和历史上下文不变；新 attempt 分配新 HBH。任何已知可能写出的前驱使 T=1；所有前驱都证明 NotWritten 时 T=0。新 Route.Request/Submit 是新 exchange，不能用相同 Session 冒充原身份重传。

| 竞争/失败 | 确定结果 |
| --- | --- |
| 旧 CCA 在 AwaitingRetransmit 到达 | 仍按原 attempt 匹配及独立 A 校验；先赢完成门就交付唯一结果，后续 Retransmit 被拒绝 |
| 旧 CCA 与 Retransmit 接纳竞争 | 应答先赢则无新 attempt；接纳先赢但尚未 BeginWrite 则停止新写，已开始则结清实际证据后发布已选应答 |
| 新 attempt 未写，旧 attempt 已 Written | 总 DeliveryEvidence 仍可能已处理；不以末次 NotWritten 覆盖 |
| 任一相关有效成功/失败/E 位错误答案 | 结束 exchange，不自动也不显式原身份再发。3002/3004、临时错误或重定向的后续属于另外的应用失败流程；不能在收到错误后机械保留 T=1 |
| Cancel/Port.Close/Node.Shutdown/所选 Peer 管理关闭 | 停止未开始 attempts；AwaitingRetransmit 没有当前 Peer 发送权，由 Node 取消/总 deadline 管理；Node.Shutdown 不为等待应用决定延长排空期限 |
| 已完成、跨 Node、重启后的 RequestID | RequestNotActive/ForeignRequest；不导入结果或原始头恢复发送，不创建第二次完成 |

所选 Peer 在最后 attempt 仍附着时被显式 Close，结束该 exchange；如果它之前已因传输失败脱离并进入 AwaitingRetransmit，关闭该旧 Peer 只移除候选，不取消 Node 所有的等待。事件顺序由提交门决定并记录原因；应用要取消请求应调用 Node.Cancel，不能依赖关闭一个历史 Peer 作为全局取消。

#### 1.6.5 最小调用与预算

文档调用顺序：配置固定 OCS 的 ExplicitRetransmit Route → Submit CCR 得 RequestID → 从 RequestInfo 观察 AwaitingRetransmit/失败 AttemptID → 应用决定继续后调用 Retransmit(id, after) → 消费原 CompletionPort 一次。若不需要应用介入选择时点，配置同终端 TransportFailover 并只消费 Completion；不要再调用 Retransmit。业务持久化、Tx/Tr、DDFH/CCFH 和补偿均不在这条库内状态链中。

接纳仍预留 `Q + MaxAttempts×A + M` 和一个完成槽，M 加有界 attempt 边/状态摘要。进入 AwaitingRetransmit 不归还用于后续重传的 Q/未用 A；旧连接仍可能收答则保留对应 A。只有旧连接不再收答且 reader/decode/AnswerCandidate 的引用都结清、没有选中结果占用时，才可释放该次 A 并仅保留发送摘要；不能因断连先于完成门事件到达而释放候选。已释放的旧 A 不会用作新增预算，也不在再接纳时分配超过最初 MaxAttempts 的槽。E2E 始终 Active，内置轮次不能回收；自定义源的一条 GuardSlot 在最终 writer 结清、exchange 禁止后续发送后才转 quarantine，见状态契约 §4。

应用不消费、长期等待或不决定都会占用已承诺额度，由最初 Deadline/Port.Close/Node.Close 结束，不引入无限期 retry queue。明确不支持：完成后同 ID 再开、跨进程/重启重传、业务错误后改内容重发、跨 OCS 会话迁移及持久缓冲。设计已给出上述有限能力与排除项；实际 Profile/对端去重及实现竞态仍须 REPLAY-01–04 验证。

参考 [OTP prepare_retransmit](https://www.erlang.org/doc/apps/diameter/diameter_app.html) 的原请求和旧应答竞争边界、freeDiameter [p_sr.c](https://github.com/freeDiameter/freeDiameter/blob/9d4bc47584bfeced9de708fe00f5629ac4689db5/libfdcore/p_sr.c) 的 pending/T 位重排。ExplicitRetransmit/after 及固定 OCS 是本库选择，不声称参考库已有相同 API。

### 1.7 RO-06：应用驱动的有界恢复契约（2026-11-04，首版内P5）

本节是sms-fw恢复能力的实施契约，替换此前“跨OCS/重启原身份恢复尚无入口”的设计缺口；**设计收敛不等于实现、应用接线或目标OCS验证完成**。原始调用分支见[主流程§33](diameterkit-mainflow-audit.md)，有限验收见[覆盖§38](diameterkit-usage-coverage.md)。仍是公开无持久化库：只管理本Node内的请求、匹配、时限和资源；没有Store、Resolve、charging.Engine、后台持久队列或业务事务回调。

#### 1.7.1 三种恢复关系与唯一发送者

> 第55批的迁移许可采用保守接纳：新目标计划即要求与原目标共同ReplayDomain、各目标EvidenceID及保存期；不等待“可能写出”后再补授权。TERM限smsfw-ro-recovery-v1及目标/本请求会话证据，不把它称作标准CC-Session-Failover支持。一次段为有限访问轮，离开目标不返回；新段可重访但不重置总次数/年龄，跨段业务轮次由应用控制。主目标未Ready时不暗中换目标；已接纳后仅应用显式Retransmit遍历固定计划。

| 关系 | 库提供的机制 | 应用/部署必须提供的决定和依据 |
| --- | --- | --- |
| 同OCS换DRA/连接 | §1.6原exchange的自动或显式恢复 | 固定终端、去重域、次数/期限；不用计费服务器迁移条件代替传输条件 |
| 无答、Tr到期，仍在本次exchange | 下述响应等待门；显式选定已授权终端/路径，串行建立后继attempt；旧答案仍可竞争 | 是否继续、下一OCS、DDFH/业务Tx决定；自动TransportFailover不能同时运行 |
| 已收到可续发的错误、窗口结束后缓冲、进程重启 | 应用提交保存的不可变Ro请求及进度，库创建**新本地exchange**，恢复线上的原请求身份；每段各一次Completion | 保存、调度、去重/所有权、跨段累计限额、消费结果；旧exchange不得仍能发送。普通Request不是此入口 |

选择初始OCS与换传输路径不同。恢复配置列出有限的 `TargetID → Destination-Host/Realm、AllowedPeers、AnswerSource、ReplayDomain`，同一操作的目标集合在首次准备时冻结。HA按优先顺序、非HA按轮询选**新操作**；Go应用选择TargetID，库再次检查Ready/绑定/额度，状态快照不授予发送权。恢复不得把整个Peer队列按第一条消息的route搬到备用，不因主节点重新上线自动搬回在途计费。

IEC CHECK/DEBIT/REFUND可按固定操作规则选择另一OCS；有可能已扣/退时，部署须证明整个目标集合的共同去重及保留期限。ECUR尚未发送的新INIT可初选备用；未取得成功CCA的INIT无答后换OCS须证明初始请求的共同去重/预留状态处理，保持原INIT身份，不能假定原OCS未建会话。成功会话的后续TERM换OCS必须给出关联会话的CC-Session-Failover许可及目标会话连续性。缺省/NOT_SUPPORTED不允许标准会话迁移。所选SMS规则若不使用该AVP，不能凭一般DCCA许可自动放开：须走独立具名旧OCS兼容及会话/重复处理证据。相同Realm、同一DRA、相同Session或T位都不是证明；无依据时该目标迁移未启用，RO-06对应部署仍未通过。

SCAP1实际进入旧timeout/queue分支，不能漏掉。标准CCR许可不外溢到App=0/Command=271：在已审定的SCAP1固定binding内单独声明允许的动作、身份/编号字段、答案关联与恢复规则，复用本节调度器；缺该模板证据时保持禁用并列为全覆盖阻塞。ASR/ASA及任意未知命令仍不进入此恢复入口。

#### 1.7.2 原请求记录与最小入口

> 第五十七批交接验收：现有API的确切意图确认、未知崩溃、已知终态、Tx/晚答及DIRECT_DEBITING DDFH示例已通过七组真实TCP/TLS测试，见[应用交接清单](ro-recovery-handoff.md)。生产API不变；本地示例不证明应用存储/隔离或真实OCS行为。已选CCR库侧交接示例工作收尾，剩余应用/部署证据单独验收。

> 最新实施更新（第五十六批）：显式AllowRoRecoveryImport/Plan.Import及Checkpoint、原记录重验、未知崩溃T位、普通号域/RecoverySpaces和准备/关闭回滚已实现，见[batch56](../spec/history/batch-56.json)。无自动重放；持久意图确认、旧发送者隔离和历史进度真实性仍归应用。标准CC-Session-Failover、SCAP1恢复及实际OCS/部署证据仍未关闭；以下批注按历史时点阅读。

> 历史实施更新（第五十五批）：本地准备/无答/Tr/具名错误续段及有限授权跨OCS已接通；RoTargets、记录v2、AllowedPeers/目标证据冻结和同DRA多目标例外见[batch55](../spec/history/batch-55.json)。标准CC-Session-Failover协商、SCAP1恢复及重启导入仍未开放；旧批注保留历史时点。

> 实施更新（第五十一批）：PrepareRoRecoveryRequest的四Ro CCR准备/冻结、本地E2E保留及scratch/guard配额已实现；记录CheckIntegrity只是结构/摘要校验。SubmitRoRecovery、累计段高水位、重启导入与RecoverySpace仍未实现，下文仍规定这些后续验收。详见[batch51](../spec/history/batch-51.json)。

仅增加两个候选入口概念，最终Go拼写在调用原型确定；不恢复通用Stage/Commit框架：

1. **PrepareRoRecoveryRequest**：对已审核Ro/具名兼容请求做Build检查，冻结内容、目标集合和恢复上限，使用Node发号器分配原E2E，返回owned `RoRecoveryRequest`。不选连接、不注册pending、不写网络，也不调用应用存储。准备前预留有界scratch和一个恢复标识槽；准备成功后标识保留到记录恢复期结束及ReuseGuard，避免应用保存期间被普通新请求取用。
2. **SubmitRoRecovery**：接收该记录、本段有限计划、应用提供的进度及CompletionPort；同一入口用于首次发送、错误后续发和重启导入。接纳前验证记录版本/长度/完整性、固定Profile和binding、身份、目标/触发、原恢复期限、剩余次数、标识冲突及Node/Port状态。成功返回新的本Node RequestID；拒绝不发字节，不产生Completion。活动段用原RequestInfo/Retransmit/Cancel控制，显式模式独占调度，不另开可靠事件流。本段计划包含有序TargetID及路径；Retransmit(id, after)只在该计划中选当前Ready且本轮未访的候选，不增任意目标参数。需要改变计划须先结束本段，再在原冻结集合/总限额内提交获准的新段。RO-06的单段MaxAttempts允许为1：此段只有一次发送，响应等待到期直接完成无答结果，不进入AwaitingRetransmit；以后是否接纳新段仍检查原总限额和原因。不能把普通Route多attempt模式的“至少2”约束用于拒绝这种单次段，也不能因此给每段重置总次数。

| 输入部分 | 必须包含 / 不得包含 |
| --- | --- |
| 不可变请求 | 记录版本、随机RecordID、Origin-Host/Realm、原E2E、完整Session、App/Command/P位、CC类型/编号/action或SCAP1对应身份、完整AVP及私有字节、Profile/WireFingerprint、固定校验上下文与RecoveryPolicyID；原事件时间/单位/退款凭据不变 |
| 恢复边界 | 有限TargetID集合及binding摘要、CreatedAt、不可延长的ReplayUntil、MaxTotalAttempts、MaxSegments、每段MaxAttempts/最长等待上限；时间/次数为正且受部署配置更严上限约束 |
| 应用进度 | 本段递增SegmentNo、累计已预留attempt数、累计发送证据、已访目标/轮次、触发原因、上一段真实结果/明确无答或崩溃未知、业务Tx是否已结束；ECUR提供关联会话许可/历史上下文。进度不是可执行业务脚本 |
| 排除 | 旧RequestID/AttemptID/ConnectionID、旧HBH、socket、pending指针、数据库句柄、Owner租约或存储回调；这些不能从磁盘复活 |

记录是有版本的owned协议数据，应用自行序列化/保存；库只接受完整受限字段，不执行旧C++队列文件加载。摘要用于完整性与本地一致性，**不是来源认证或数据库提交证明**。不能通过改摘要重新构造同一E2E的另一条消息。首次请求冻结除Destination-Host/Realm和逐跳头外的全部内容；换目标只按已冻结TargetID映射替换目的字段，重新验证该attempt的AnswerSource。旧Proxy-Info/Route-Record不可作为本地新发出的逐跳状态导入；具名兼容若实际携带此类字段须有明确本地构造规则，否则拒绝。

Origin必须是当前Node被配置承接的本地身份；重启时当前CER/新控制请求的Origin-State仍反映本次启动，重放请求中原有Origin-State保持原请求值，不能顺手改写正文。此差异须纳入OCS重启恢复互通。改变Origin、CC编号、动作、服务或费用是新协议操作，由应用另行构造，不能伪装成原请求恢复。

#### 1.7.3 Tx、Tr与有界等待

> 第53批已实施本节的**本地原OCS同段**Tr门。Plan.ResponseWait=0禁用；正值不超过原MaxResponseWait。AcceptedAt固定首次排队，AttemptAcceptedAt/ResponseWaitUntil跟随当前attempt；AttemptResult保留逐次时间。到期但writer未结清时只有ReplayReason，ReplayEligible仍false；实际FrameWrite截止仍有效。单次/最后Tr以ErrResponseWaitExpired/StopResponseWait完成，整体deadline优先；健康连接、旧pending和独立A保留，后来的真实关闭/协议错误不能被Tr屏蔽。业务Receive的ctx超时不改Tr。不同目标共用同一DRA的下述例外仍待跨OCS实现；当前同一Connection代次不回访。详见[本批接口及验收](../spec/history/batch-53.json)。

**Tx属于Go业务等待；Tr属于本次attempt响应等待；ReplayUntil属于整个原请求的最大恢复年龄。三者不能复用一个context deadline。** Tx到期可向业务交付超时/按策略放行，但继续消费原CompletionPort；不Cancel就不结束协议收答。旧TIMEOUT_RETRY_ENABLED=0映射为Tx时Cancel，明确接受随后晚答被丢弃，不伪称仍能无损收答。

为RO-06显式模式增加固定 `ResponseWait` 门：首次及备用等待值在本段接纳时确定；从attempt进入有界发送队列的单调时点计时，RequestInfo增加AcceptedAt、ResponseWaitUntil及ReplayReason。等待到期只在writer已终止且仍无有效答案时进入AwaitingRetransmit，原因是ResponseWaitExpired；不要求把健康连接伪装成断连。写入尚未完成时按写期限结清/必要时关闭连接，绝不并行启动第二个writer。Tx自身、用户任意调用和未到期快照仍不授予重传权。不同目标共用同一DRA时，到期后可沿该健康Connection向下一已授权Target发送新HBH；这是本入口对§1.6“每连接代次只用一次”的明确例外。仍不回访本轮已访Target/路径，不作并行hedging；下一持久轮可重访，但须另获新段准入。

旧 `prepareForTr`只改Tanswer、不改tSend，结合dia-peer超时检查，**首次Tr以原排队时点t0为基准，即t0+Tr，不是t0+Tx+Tr**；备用排队才重置tSend。本迁移固定首次ResponseWait=Tr、备用=Tr，Tx应用计时以本段AcceptedAt为基准；要求配置0<Tx<Tr。旧Tr≤Tx的立即连锁超时配置显式报配置差异，不静默改成新的等待语义。排队占用等待时间，不能借拥塞无期限续期。错误答案在Tx前导致新段时，应用沿用原业务Tx截止，不重新给予一次业务等待。

Tr1选择当前可用、未访且授权的备用；Tr2无答/无备用时由应用按下表结束、保存或以后重放，库不轮询等重连。旧普通流程的Tx→Tr1→备用Tr2映射为至多两次网络发送；错误结果走多备用时按有限目标数扣次数。重放队列一轮首选+至多一个备用，随后交回应用决定下轮间隔；本轮visited集合只用于防止同轮循环，下一轮可重访主节点，但累计attempt/段数和年龄不能清零。

#### 1.7.4 结果、DDFH及T位

> 第54批实现本节的本地原OCS有限错误续段。冻结PolicyID为`ro-transient-errors-v1`或`smsfw-ro-recovery-v1`且Plan显式选RoErrorTransient，才允许3002/3004/3005；后一个名字配RoErrorSMSFWBeforeTx和ErrorContinuationUntil，才在校验答案时点早于该截止时加入4002/5001/5004/5005/5008/5012。截止后/相等只保留临时集合，不取消请求。默认模式仍所有答案终态。NextTrigger由获选真实答案及本地guard授予，外部声明不能替代；错误后T清零仅清最后一次有效错误之后的重传依据，累计Delivery不清零。4002仅接收且报告语义不完整。后文跨OCS/持久队列映射仍为应用/后续实现契约，不能据本批声称已实现。


库返回真实Result-Code/Experimental-Result、顶层/服务级结果、DDFH/CCFH Presence与值、发送证据及关联报告；不把队列接纳、Written、超时或应用放行变成2001。标准DDFH用于DIRECT_DEBITING，不能替代ECUR的CCFH/会话许可；REFUND临时失败的保存义务与CHECK失败处理由应用按动作落实。旧通用wrapper把DDFH用于多动作的行为只在具名迁移策略中保留，不能宣称标准语义。

| 原分支 | Go应用承接；库的边界 |
| --- | --- |
| 首发所有route down | 旧代码返回ROUTE_DOWN且不入持久队列；应用映射该业务错误。库给NoEligiblePeer/NotWritten，不能自动制造缓冲成功 |
| Tx前3002/3004/3005 | 真实错误结束本地exchange；应用选择未访备用并以ErrorAnswer原因提交下一段。正常结果/永久失败不统一归“成功” |
| Tx前4002/5001/5004/5005/5008/5012 | 旧代码也轮转备用；仅 `smsfw-ro-recovery-v1`具名兼容策略允许这个有限集合，仍保持原请求，须有目标处理依据；库不补AVP/改动作规避错误 |
| 上述错误且无备用 | 旧DDFH=0返回实际失败并尝试保存；DDFH=1先decodeCCA(NULL)，随后dealTmpErr改成业务临时错误，且也保存。应用保留这种分支差异，不把日志里的“success”当最终业务结果 |
| Tr1无备用或Tr2无答 | 旧分支用配置DDFH：0保存、1结束且不保存。应用保存/停止；库给无答与累计可能发送证据。没有CCA时不能伪造远端DDFH覆盖 |
| Tx后、最终窗口内晚答 | 2xxx/永久结果仅结束后台恢复；3002/3004/3005保存。这里不采用Tx前六个额外错误的backup集合；旧DDFH=1也可能走该保存分支。业务Tx只完成一次 |
| 持久队列一轮 | 初始无Ready、备用耗尽/超时/3002/3004/3005保留记录；2xxx以及其余3xxx、全部4/5xxx结束该队列项，但保存真实结果。旧msgProcessed=true表示处理完毕，**不表示扣费成功**；未知结果/关联错误不成为成功或自动删除依据 |
| 队列关闭/满/记录超限 | 应用明确保存失败、告警/拒绝或按批准策略继续；旧日志后丢弃不是新系统的成功保证。库不接管队列、不承诺补偿成功 |

DDFH远端合法值在有该字段的适用回答中覆盖配置；保留Absent/0/1，不能照搬旧解析器以wire+1表示Presence的内部技巧。其作用范围由应用本次事件/会话策略固定，不跨用户或route继承。yppdia的TX_TIMEOUT、CONNECTION_DOWN、DIAMETER_TMP_ERROR及SYSTEM_FAILURE后续业务映射也必须由Go应用承接，具体五操作矩阵见主流程§33.1；wrapper临时码不是最终业务决定。已获最终成功或不在允许续发集合的失败不允许SubmitRoRecovery再次发送；应用需要新扣费/退款时走新操作，不能为其复用原身份。

T位由**本次续发原因**而非“来自磁盘”决定：首次及所有前驱可证明NotWritten时T=0；无答/响应超时/崩溃未知且可能写出时T=1；已收到有效错误答案后的显式续发T=0，遵守RFC6733 §3，之后若该新段又无答，其重传T=1。历史可能发送证据仍保留，T=0不等于从未扣费。应用不能直接覆盖T/HBH；不得把上段错误藏成超时来取得另一种规则。

#### 1.7.5 应用保存、重启和晚答的交接

只需进程内恢复时，应用可内存持有原请求并直接SubmitRoRecovery，不要求任何存储。**声明跨重启恢复时**，首次发送顺序固定为：准备原请求 → 应用可靠保存原请求及**本段发送意图/次数额度** → SubmitRoRecovery → 消费唯一Completion并由应用更新记录。每段先预留其整个MaxAttempts额度：进度明确携带ReservedBefore、SegmentAllowance和ReservedThrough，必须满足ReservedThrough=ReservedBefore+SegmentAllowance≤MaxTotalAttempts，且本段MaxAttempts≤SegmentAllowance。首次ReservedBefore=0；后段至少继承上段ReservedThrough，本Node按guard高水位阻止回退。进程在接纳/写出前后崩溃且结果未知，该额度视为已用，不能凭缺少Completion退回。未使用额度可保守浪费；不引入存储提交确认接口。存储提交未知时应用不得发送，必须在自己的存储/所有权协议内确认或停止。

重启流程：应用隔离旧发送进程/worker → 读取原记录和累计进度 → 检查年龄、目标去重保留期、Profile/绑定版本与原Origin授权 → 为下一段保存意图 → 新Node接纳。新BootID/RequestID/AttemptID/HBH，原E2E/Session/CC字段和正文保持。库只校验当前Node内冲突及记录约束；外部多worker互斥、历史进度未回滚、对端记录仍存在由应用/部署保证，内存guard和RecordID不能证明跨进程恰好一次。

同Node每个RecordID至多一个活动段；同SegmentNo重复提交在该段活动时返回既有RequestID、不重复排队，绑定不同Port/内容则拒绝。该段完成后不再次返回历史Completion，返回SegmentAlreadyFinished；更大序号仅在上一段writer/pending/candidate已结清、原因获准时接纳。同一记录在本Node内保存有限进度高水位和终态拒绝标记，不能用旧进度撤销成功/重置次数；重启后库没有这些事实，只能由应用提供已确认进度。

| 答案到达时机 | 必须行为 |
| --- | --- |
| Tx后、原exchange未结束 | 合法匹配仍交付同一个Completion，应用处理后台结果，不第二次唤醒已结束的业务调用 |
| Tr到期/后继接纳/BeginWrite竞争 | 所有仍有预算的旧attempt都保留匹配；旧答可赢完成门，停止未开始的新写，已开始写需结清证据；仍只有一次Completion |
| 某段已完成，新段已接纳 | 旧HBH/Connection的答案不得按Session附到新段；旧段答案只计有界丢弃/诊断。新段靠自己的请求匹配取得OCS原结果/去重答案 |
| 最终期限、Cancel/Port.Close/Node.Close后或重启旧连接 | 不承诺无损晚答，不开永久监听/结果邮箱。应用保留“未知”，在允许期限内重放/查询对账或人工处理；不能根据本地丢弃推导未扣费 |

错误后续发和重启不会延长旧exchange的Deadline；新段有新本地deadline，但不得超过不可变ReplayUntil和剩余全局次数。进程内使用单调时钟；恢复记录保存绝对时间及原最大年龄，应用必须提供可信的跨重启时间检查。时钟回拨/年龄无法确认时拒绝自动重放，不能重置CreatedAt。OCS去重过期后的原身份请求也不得靠更换E2E重发，转应用核对。

#### 1.7.6 标识、预算和退出条件

恢复记录占一个有界GuardSlot，即使Node使用默认顺序源也一样；持有RecordID/内容摘要、原E2E、进度高水位、终态和到期元数据，不存完整历史答案。准备但未发、段间缓冲均占槽，保留至 `max(ReplayUntil, LastWriterRetiredAt)+ReuseGuard`；不能按LRU提前淘汰。槽/每class配额不足时Prepare/导入明确拒绝。普通新E2E分配必须检查这些保留；导入已被普通请求使用或另一记录占用的E2E时拒绝，不能借同Session豁免。详见[标识契约§4.7](diameterkit-state-consistency.md)。

每段接纳仍预留 `Q + MaxAttempts×A + M + 一个完成槽`，M计有界目标/attempt进度；Q的编码长度上界必须取冻结目标集合中各Destination-Host/Realm替换后的最大值，并计入原请求owned对象与编码缓冲可能同时存在的内存，不能只按首选目标帧长预留。Prepare/导入先检查每个允许目标的最大帧长与MaxRecordBytes/消息上限，任一超限在发号/接纳前拒绝；后继编码不得临时借其他class额度。等待Tr/显式下一步不释放潜在收答所需A。Prepare临时复制/校验也受最大记录长度和class scratch限制，返回的owned记录内存归应用；调用方无限保留副本不算Node可控内存。Node只保留guard摘要，段完成归还Q/A/M，guard按上限退休。MaxSegments、MaxTotalAttempts、MaxReplayAge、MaxRecordBytes和目标数均须有限展开；不允许用新段重置Node总资源或原请求次数。

本轮关闭的是**恢复入口、计时/身份/晚答/预算及责任分工的设计缺口**。实施退出按覆盖§38有限矩阵；OCS去重/会话连续性、六模板恢复授权、Go应用持久保存及切换验证仍是对应交付门槛。证据未到不扩大标准承诺，不阻塞已明确机制的离线实现，也不再追加全量Ro规范调查。

## 2. R20：协议校验阶段与上下文所有者

### 2.1 一个规则模型，五个阶段

生成规则记录 RuleID、Profile、适用命令/方向/角色、Phase、RequiredFacts、错误分类及线路错误映射。条件表达式采用 True/False/Unknown 三态；缺协议上下文产生 Unknown/MissingContext，不当作 false 跳过规则。只有带条款依据的协议输入才进入 RequiredFacts；订阅/账务是否存在等业务决定不自动变成库的 Check 前置条件。定义编译拒绝无法提供所需协议输入的启用组合。

| 阶段 | 输入 / 执行者 | 成功边界 |
| --- | --- | --- |
| Build | 操作输入或生成 body；应用构造器/生成 ValidateBody | 所有业务必填字段和可得条件通过；只允许缺声明为 runtime-owned 的 envelope 字段，不允许缺任意必填 AVP |
| Wire | 完整 pdu.Message；CompleteRequest 后的 Codec.Validate/Append，或 Decode | framing、类型、数量、命令语法、静态条件通过；离线 API 不声称业务状态已验证 |
| Exchange | Wire 通过的快照；Node/Peer/exchange | 本地角色和方向、授权来源/目的、能力、关联以及已绑定的应用 facts 满足；无业务 I/O |
| Application | 不可变请求/应答 + 生成的应用 facts；业务 worker/纯解释器 | 声明的协议上下文规则完成；业务副作用是否允许由应用决定 |
| Answer | 已验证入站请求与完整应答快照；Respond/Reject | 应答 envelope、请求关联和相应结果分支的字段/条件通过，随后才取得发送提交权 |

同一规则只定义一次，按所需 facts 在最早合法阶段执行；context 改变时按依赖重算。Codec.Validate/Decode 仍严格执行 Wire 阶段，不增加 Strict=false。完整消息必填 Origin 缺失仍报错，ValidateBody 仅允许指定 envelope 缺省；数值 0、空 OctetString 和缺失继续分别表示。

### 2.2 Context、角色与公开入口

ApplicationBinding 将单个 LocalRole 改为 EndpointRoles 集合及 CommandPolicies。每个 policy 以 CommandKey + 本地发起/接收方向绑定一个确定的本地角色、UseCasePolicy、所需具名协议输入和允许结果分支；同一命令方向若有多个角色且无法由固定规则唯一选择则构造失败。多个角色共享同一固定 codec，不因角色选择改变字段的线路类型。

内部校验输入由 binding 的 Role/Direction、当前/原请求推导的 UseCase/RequestEvidence 和调用方 validation.Context 组成。Context 是有界不可变值，当前由 sh.ValidationContext、ro.ValidationContext 构造终端/会话特性输入；零值表示无额外输入，不使用 map[string]any、通用 ApplicationFacts 或业务记录。CommandPolicy 固定 MaxContextBytes 与生成的输入类型约束，按规则核验类型/关联、先计量再复制。DRA CER 不能证明终端特性；Check 的快照空间计入逻辑 I。第九轮收缩 R24，详见状态契约 §1：不再提供通用实时 FactSet、Revision/期限或业务授权门。

| 入口 / 返回 | 明确契约 |
| --- | --- |
| codec.ValidateBody(body) → error | Build 阶段；返回有效业务 body 不代表完整线路消息 |
| codec.Validate(msg) / Append / Decode | Wire 阶段；规则无上下文时留给上下文校验，不静默宣称全通过 |
| codec.ValidateContext(msg, context) → Report | 校验全部适用上下文规则；Report 含 Scope、CheckedRules、Violations、MissingContext、LocalErrors、Complete；Scope 由入口固定，Complete 要求本 scope 无违规、缺失或本地输入错误，不能由调用方选择空规则集伪造 |
| RequestOptions.Validation | 仅提供具名的额外协议输入；能从 body/binding/原请求取得的不重复填写。普通 S6c/SGd/PNR/IEC 无此需求时用零值；依赖特性的规则缺输入才报 ContextRequired |
| Inbound.Validation() → Report | 返回复制的当前报告及 NeedsApplication/Ready/Rejected/Expired 校验状态；Receive 不等于 Application 完成 |
| Inbound.Check(context) → Report | worker 中同步、有界、无 I/O 地校验库保存的原请求；成功冻结 facts 快照并进入 Ready；不读取业务修改过的 Message 副本 |

直接 Request/Submit 在成功接纳前完成所有发送所需 Build/Wire/Exchange/Application 检查；未知的远端未来结果不属于发送前必需事实，不能用它拒绝所有请求。缺 facts 返回 ContextRequired（含 RuleID/FactKind），不产生 Completion、不写入。Route 的固定 facts 与调用输入冲突时 ContextConflict，不按最后覆盖者决定。

Request/Submit 保存消息及协议上下文快照，接纳和每次实际写入前核对固定发送策略、deadline、连接能力及对应写入资格。所有入口执行 CommandPolicy.OutboundPolicy，Route 只能收紧；内部故障转移复用原输入。库不在 writer 查询或刷新订阅、授权、账务状态。

历史请求事实冻结后用于关联/解释，不因墙钟推进而撤销。时间 AVP 的专属规则由固定 Profile 描述；业务事务版本、授权期限和“是否允许发送”由应用在调用库前处理。没有具名规范用例前，不公开 SendFacts、RespondWithFacts 或通用时钟可信度框架。

### 2.3 Receive、Check、Respond 的门槛

完整入站帧先通过 Wire/Exchange；静态或邻接授权失败不交付业务。只剩具名协议上下文缺失时，可交付状态为 NeedsApplication 的 Inbound 给有界 worker，用于补齐协议输入和调用 Check；其报告列出具体缺失项。不存在缺输入或未支持语义、且适用检查全部通过时才直接Ready。规范未裁决/语义未支持不进入补Context的等待流程，也不标Ready或作为已验证业务请求交付；在接纳边界按固定FailurePolicy处理，只有已审核的合法映射才能应答，否则终止该处理路径。原因保留为本地未支持，不能伪装成对端FieldViolation。reader不查数据库，不调用应用回调。

Check 使用内部不可变请求及绑定 Profile；Message 是交给业务的可修改视图，不是校验凭据。MissingContext 可再次补齐；facts 类型错误、超限、冲突等记入 LocalErrors，可修正输入，不能把本地错误归咎于对端。确认原请求违反协议时标记 Rejected，不能 Respond 普通应答，只能 Reject 或等到超时。第一次成功 Check 固定协议上下文快照，再次传相同 facts 幂等，传不同 facts 报 ContextConflict；需要更改业务结论时应基于业务事务处理并使用合法应答，不能篡改已验证请求。Rejected 是校验状态，不等于应答已提交；逻辑句柄的 Open/Committed/Expired 生命周期独立，Done 在不可再提交结果时关闭；单个 Arrival 断线不关闭逻辑句柄。

Ready 仅证明给定输入下的协议检查完成，应用自行负责业务事务和副作用。逻辑句柄超时或 Node 关闭后 Respond 失败；单个 Arrival 断线不阻止在逻辑期限内提交结果供合法副本使用。应用无法作出业务决定时可选择合法失败分支，不能用假的成功答复代替未知结果。

Respond 要求 Ready，验证应用构造的结果分支及完整应答 envelope，并使用原请求关联和已冻结的协议上下文。失败不消耗提交权；成功后取得一次提交权，缓存结果，为各 Arrival 编码。ResponseResult.Committed 与 PrimaryWrite 分开。库检查例如失败 ASA 的字段语法，但不判断外部 Session 是否存在或自动确认会话已取消。

Ro适用TS32.299 §6.5.5时，应用CTF仍负责收到ASR后的CCR TERMINATION；无自动发送不表示该协议义务可省略。ASA提交与TERM结果是两个独立exchange，库不等待另一笔完成来制造原子提交；应用协调在途TERM和编号/用量，不能因ASA成功而跳过终止报告。参见[主流程§30.4](diameterkit-mainflow-audit.md)。

Reject 允许在 NeedsApplication/Rejected/Ready 状态调用，但仅接受具名 ProtocolFailure；不接受任意成功结果或借此绕过 Application。由 binding 固定的 FailurePolicy 把 ContextUnavailable、Overloaded、DeadlineExceeded 及 FieldViolation 映射为相应命令合法的普通错误或 E 位错误；未定义合法映射时关闭/终止，不伪造成功或“缺 AVP”。构造期必须为启用命令指定这些分支。码值来自固定规范，不能把全部错误统一映射为 UNABLE_TO_COMPLY。

### 2.4 应答和失败处理

相关应答通过 Wire/Exchange 后可结束网络等待；若请求保存的 facts 足够，则同时执行应用解释规则，否则 Result.Validation 明示 NeedsApplication，仍交付 owned Answer 和发送证据。不得把收到应答直接解释为可投递或已扣费。第十二轮 [主流程结果表](diameterkit-mainflow-audit.md) 要求保留基础/实验命名空间、顶层/MSCC 结果及缺/零单位；协议失败不会变成传输重传授权。Ro SMS 的 FieldSelection 在纯构造阶段应用，不成为 runtime 的业务事实查询或额外会话表；其配置由调用者持有，接纳后只保留有界消息及原已声明的必要协议输入。薄 S6c 客户端及 Ro 纯解释器必须完成适用 Application 检查后才给出类型化结果；失败保留原应答和 Report。主流程审计§12的规范未裁决/语义未支持与可补输入的NeedsApplication原因分开记录，不能伪称ContextRequired或等待业务查询；Wire/Exchange通过的相关答案结束网络等待并交付原值/未完成语义报告，不自动重传。发送helper在接纳前拒绝受影响分支。入站请求若受未支持语义影响，不得标Ready或据此Respond成功；按已有FailurePolicy处理，缺合法映射则终止，不猜5008、不为此增加状态或泛化facts。

可关联指实际ConnectionID/HBH及App/Command/E2E完整头匹配；仅HBH命中但其余头不匹配走有界discard，不消费原pending的A/完成权。完整头匹配后，静态/Session/来源无效的应答以 ErrInvalidAnswer 终结，不自动重传；只保留有预算的原始诊断，不伪造成功的强类型对象。无效应答不产生错误应答。Ro 5011 含缺基础码、结果冲突等具体处理见 [主流程审计 §9.4](diameterkit-mainflow-audit.md)；只有 Wire/Exchange 通过的应答才能作为带应用报告的 Answer 交付。纯 ctx 取消与收到正常业务拒绝分别记录，MissingContext 也不是对端拒绝。

| 用例 | facts 所有者 / 必需验证 |
| --- | --- |
| SRR→SRA | QueryRouting 保存查询 UseCase 与 Subscriber 语义；解释器验证 HSSID/IMSI、节点组合，不从 DRA 推导终端特性 |
| OFR→延迟 OFA | binding 检查角色/方向与具名协议条件；投递决定属于 worker，应答仍使用原内部关联 |
| SNR、独立 PNR→PNA | 库检查各自消息/结果和对应请求关联；不把 PNR 关联到 SNR 的 pending；订阅查询与通知消费在应用，不要求 Found/NotFound 或数据库版本才能答 PNA |
| CCR/ASR | 正常CCA的CC编号/类型从原CCR快照关联；ASR/ASA不要求这些字段，E错误使用独立语法。只有依赖会话协商的特性规则才需要具名SessionFeatures；普通ASR/ASA协议准入不查询业务会话。ASA成功表达应用报告已终止，库不独立确认外部状态 |

## 3. R21：单账本、保留额度与所有权转移

### 3.1 额度模型与资源分区

Node 内部只有一个资源账本。每个 reservation 记录 Capacity、Held（实际占用）、Owner 和 ChargePath；`0 ≤ Held ≤ Capacity`，已占用是 Capacity 的组成部分，不另收一遍。Node 的 committed 为所有 reservation Capacity 之和；Peer、ResourceClass 和 Port 是同一 reservation 的约束维度，不是再次分配的内存。申请必须同时满足所有适用限额，否则原子失败并撤销，不持有部分额度等待另一维度。

使用内部 reservation/split/transfer/release 操作，均不暴露给业务。transfer 保留同一份 charge；split 的容量和等于原容量；确需并存的复制才申请新增容量。release 幂等且发生在最后一个库持有者释放后。库拥有对象的结构、切片/字符串头和保守分配开销计入各 bound；Go GC 元数据、分配器保留空间及 TLS/内核外部缓冲另行报告，该账本不是进程 RSS 硬上限。

v1 使用硬分区，不允许借用其他分区空闲的保留额度：

| 分区 | 归属与用途 |
| --- | --- |
| TransportReserve | Node 构造时为最大连接/握手数预留固定头、SCTP 前缀暂存、单 writer scratch、控制接收/发送与有界拒绝槽；业务不能借用 |
| IdentifierReserve | Node 的固定发号状态及自定义源的有限 active/quarantine 索引；默认顺序源无逐完成记录；与 Q/A/M 的引用不重复收费，见状态契约 §4 |
| ResourceClass.Outbound | 按应用/固定 Profile/业务组分配：新请求、进程内故障转移 attempts、完成结果；多个 Route 可指向同一 class |
| ResourceClass.Inbound | 入站快照、worker 句柄，以及所有 Arrival（包括重复副本）的 I/W/M 和发送槽 |
| ResourceClass.Duplicate | 进行中去重记录、规范内容和完成业务结果缓存 C；逐 Arrival 重答空间归 Inbound，不重复预留 |

ResourceClass 配置每个分区的 Bytes/Slots；Route 指定 ClassID，入站 CommandPolicy 也指定唯一 ClassID。NodeConfig.ResourcePlan 显式提供这些分区、IdentifierReserve 及 TransportReserve 的连接/握手槽、ControlMax、EmergencyRequestMax、EmergencyReplyMax、MaxDiscardBytes；builder 可生成候选配置，但 NewNode 校验的是全部展开后的具体值。无法分类、分区未配置或有未展开的零值时构造失败；仅内置发号源且未启用RO-06时，Guard跟踪显式Disabled是允许的例外，不代表无限容量。自定义源另配 ResourceClass.IdentifierSlots，active/quarantine 共用原 class 配额，配额之和不超过 Node GuardSlots；物理字节只在 IdentifierReserve 记一次。多个类分区容量加 TransportReserve、IdentifierReserve 必须不超过 Node 上限。TransportReserve 按允许的最大连接/握手并发预留固定前缀、独立控制 RX/TX 及拒绝槽；不得只按启动时的连接数计算。Peer 限额额外限制实际 attached 资源及 pending 数。当前 16 MiB/Peer、128 MiB/Node 仍为容量原型，不保证任意四应用配置都能通过构造或满并发；不自动压缩保留值使配置“能启动”。

分区不会因慢 Sh 完成消费者而侵占 SGd/控制的预算，但同一连接的网络读写仍可能受阻。保证的是不等待业务释放额度的处理/失败路径，不能保证故障网络或无限对端洪泛下业务成功。

### 3.2 如何计量并在分配前申请

MaxAnswerBytes 表示线路帧长度上限，不等于完整 owned decode 的内存上限。定义生成器输出保守的 `DecodeBound(maxWire, maxAVPs, depth)`，包含 backing frame、字符串/字段复制、最大生成对象数、校验报告和有界临时空间；每个扩展 codec 必须声明可校验的界限，否则不能进入受限 runtime。实际使用少于预留可在结果确定后缩减，不能解码到一半再无界申请。

Q 是请求不可变主体、协议上下文和本次 exchange 所需数据的完整计量；A 是一个 attempt 的 DecodeBound；M 是有界摘要/定时器/匹配记录。TransportFailover/ExplicitRetransmit 的接纳预算为 `Q + MaxAttempts × A + M`，同时预留一个逻辑完成槽和 MaxAttempts 个内部 attempt/接收槽。SingleAttempt 取 1。每个 attempt 各有 A，避免旧/新应答同时到达争用唯一结果空间；最终 Completion 只保留选中应答，其他槽在清理后释放。同一 attempt 的 A 在 Receiving/Decoding/Candidate/Port持有期间不可被第二帧再次占用；重复 HBH/E2E 不产生第二份隐式 A。重复应答均按 §3.5 有界丢弃，不分配第二份 A。

Port 的 MaxOutstandingBytes 对上述完整 reservation 做 admission 上限检查，Node 只记一次；它不是额外的一份内存。保守预留使高 MaxAttempts/大 Sh 应答降低并发，这个成本必须在配置和 benchmark 中公开。失败转移时目标 Peer 的槽和字节须再检查：原 Peer 已终止资源 detach 后可转移；原连接仍保留 late slot 时只能 split 到不同 Peer，不得从账本抹掉旧资源。

初始测量/校验只允许有界扫描和固定 scratch；计算 Q 或编码长度发生溢出、超过节点单消息限制时立即拒绝。拿到额度后才深拷贝可变输入和编码。配置/定义对象是构造期资源；离线 Codec/纯消息构造器返回的调用者拥有对象受单消息限制，但不承诺替使用方限制其长期保存量。

### 3.3 出站预算转移表

| 事件 | Owner / 额度动作 |
| --- | --- |
| Submit 接纳 | 取得 Node/class/Port 和首次 Peer 约束、完成槽、Q/全部 A/M 及必要 ID 槽；有界源调用后再次进接纳门核验并提交；失败全数撤销且无 Completion，不在锁内执行源 |
| 入队、写入、等待 | exchange 持有 Q，writer 同步借用；各 attempt 已预留 A 仍计入 committed，尚未分配完整应答 buffer |
| 固定头命中 attempt | 在读取大 body 前核对可信范围的长度/连接/Hop-by-Hop/End-to-End/App/Command；将该 attempt 的 A 转为 Receiving，不再申请另一份 frame budget |
| decoder 消费 owned frame | 同一 A 从 Receiving 转为 Decoding，再转为 Candidate；frame、生成对象及 scratch 的共存已包含在 A |
| failover | Q 不变；下一 A 从 Reserved 转为新 attempt，先检查目标 Peer 约束；旧 A 直到其接收/匹配生命周期结束才能释放 |
| 完成提交 | 先固定选中候选，再停止发送并结清 writer；释放 Q 和未使用/落败的 A，保留获选 Answer 实际保留量及 M 中的结果摘要；Owner 转为 CompletionPort，解除已无连接持有者的 Peer 约束，Node/class/Port 仍计费 |
| Receive 成功 | 结果所有权转业务并释放该结果 charge；独立 ID quarantine 仍保留至其截止，不依赖 Port；业务长期持有消息不计库预算 |
| Node.Close | 将已接纳请求转为完成/清理；port 中结果仍有 charge，Node.Wait 不等待消费；后续 Receive/Close 释放 |
| Port.Close | 取消关联 exchange，等待正在使用的 I/O 结清，再释放已丢弃结果和 reservation；绝不提前回收 writer 正在借用的 buffer |

同步 Request 使用私有单完成槽，规则相同。Ro 和其他应用使用同一 Outbound 额度及 CompletionPort，不设持久证据邮箱或 Evidence 分区。exchange 的收答窗口由原 Deadline 控制，业务的单独等待超时不自动延长或终止它。

完成后不再为这个 exchange 保留新的应答接收资格；未消费的已选 Answer 仍有 charge，直到 Receive 转移所有权或 Port.Close 丢弃。应用另行保存结果不再占 Node 额度，也不构成库承诺的持久证据。

### 3.4 入站、缓存与拒绝

入站读取完整头后，按已固定的 App/Command/方向取得 class；在读大 body 前申请 `I + W + M`：I 覆盖原始不可变证据、内部校验 facts 及业务视图的同时存活；W 覆盖该命令的 ReplyWireMax、编码临时峰值及一个发送槽。ReplyWireMax 含自动回填的 Proxy-Info/Session 等 envelope，而不只是业务 answer body。超过可用包络上限的请求不交给业务后再发现无法回答。

I = DecodeBound(请求上限) + EvidenceSnapshotBound + CanonicalScratchBound，W = EncodePeakBound(ReplyWireMax)，C 按部署与互操作契约 §3.3 包含记录、CanonicalBound、CachedResultBound(MaxResultBytes)及必要编码临时上界；W/C 包括必要的不可变字段副本和结构开销，不能把线路字节直接当作全部内存。三种 bound 均由生成描述符和扩展的上界契约计算。解析发现实际 envelope 无法容纳时，走拒绝/关闭而不交付业务。

通过 Wire/Exchange 并建立去重键后，必须在交付业务前另外从 Duplicate 分区预留固定期限内的记录和最坏业务结果缓存 C。初次请求或重复副本任一必需额度不足，都不执行业务。重复副本的解析 I 和 envelope/W/M/发送槽都从 Inbound 取得，并有数目上限；共享 C 内的规范内容和业务结果，不共享逐跳头。结果缓存不含第一次代理链。DuplicatePolicy 的 MaxResultBytes 与 MaxEnvelopeBytes 在构造期证明可共同容纳于 ReplyWireMax；原始规范内容也必须有 MaxCanonicalBytes 上界，不能只按 answer 大小预留 C。若不启用缓存支持，相应 Profile/角色不能宣称承诺期限内可重复重答。

重复键的查找/建立与 C reservation 在同一提交门内完成，避免两个同时入站请求分别执行。命中已有记录时不再申请第二份 C，按 CanonicalV1 核对内容后释放重复解析对象，仅保留其 Arrival envelope 和重答资源；未通过内容一致性检查的请求不能共享结果。C 到期释放前必须检查仍在写入/生成重复应答的引用，不能先归还再让 writer 借用。

业务视图在 Peer.Receive 成功时转给使用方，runtime 只保留独立的 I 内部证据；重复 Message() 访问同一调用者视图，不为每次调用无限克隆。Respond 测量完整结果并校验不超过 W/C；超限返回 ResponseTooLarge，未消耗句柄，可改答或 Reject。成功后从既有 W/C 生成帧及可复用业务结果；逻辑 I 在提交/超时后释放，各 Arrival 的 envelope/W 在自身发送结束后释放；C 转为缓存 owner，到固定 KeepUntil 且引用结清后释放。首连接断开不能提前释放仍在处理的逻辑凭据；超时或 Node.Close 与 Respond 争同一决定。业务完成与传输成功分开：已得业务结果但写失败仍可按去重契约保存 C。

若入站业务额度不足，可非阻塞取得 TransportReserve 的一个 Reject 槽。该槽同时包含 EmergencyRequestMax 范围内的接收/诊断预算、EmergencyReplyMax 和发送槽；仅生成具名错误，不交付业务。请求太大、代理链/Session 等必需字段无法在拒绝限额内保留、槽耗尽或边界损坏时关闭连接；不得截断必需 envelope 后声称已合法拒绝，也不在 reader 等待业务消费。

控制消息使用独立限额；超出 ControlMax 或控制槽耗尽即终止失去可控进展的连接，不能借用其他业务的额度。对新请求的资源拒绝可能关闭整条共享 DRA 连接，此影响必须出现在指标及所有受影响 pending 的结果中；保留额度不等于保证连接永不被过载关闭。

### 3.5 大报文、未匹配应答与 SCTP

所有 runtime transport 使用总体设计的 ReadPrefix → 账本申请 → ReadFrameInto 调用者缓冲区边界；同一连接只能有一个未结清的prefix/帧游标（前缀字节已经消费），取消/失败必须释放对 dst 的使用后才返回。适配器不分配第二份完整帧，不导入核心内部账本。TCP/TLS 先从每连接固定 scratch 读取 20 字节头。匹配应答长度超过 MaxAnswerBytes 或无法按已预留 A 安全解码时，完成为超限/无效应答，DeliveryEvidence 仍保留；尚未读完则关闭连接，不在 stream 中假装跳过 body。若帧已经在原 reservation 内完整消费，可释放并结束本次请求，不强制关闭仍有边界的连接；这不授权自定义传输先无预算分配完整帧。

对未知/已完成 attempt 的应答，只允许用固定 discard scratch 在 MaxDiscardBytes 和部分帧时限内消耗完整帧；超过该限额或边界不可信就关闭，不分配最大消息或无界 drain。错误的关联头不能消耗另一个合法 pending 的 A。每条接收路径必须有初始额度或明确关闭策略。

SCTP数据首块进入TransportReserve的小型前缀暂存后分类、申请相应reservation，随后的分段直接归入该token。**消费前先区分数据/通知**，由上游通用接收入口在同一读所有权下执行peek分类及对应缓冲消费；本库不能先用20字节数据区收通知再尝试补救。每连接固定prefix、NotificationMax、ancillary和上游固定scratch都纳入TransportReserve/Info上界，通知数据不借业务A/I/W、不生成第二份完整业务帧。notify超限/损坏/失去边界直接关闭；不随SEND_FAILED携带载荷的长度扩大内存。未完成消息数、暂存字节和不刷新的期限均有限，超额或交织无法无歧义组装则关闭association，不等其他应用让空间。合同依赖部署§4.5–4.7的U2/U3/U4，不能用旧SCTPRead或socket timeout冒充flags/动态deadline能力。

适配器给连接reader/writer期限，核心负责单exchange取消和Shutdown总截止；上游只处理generic deadline、Context及fd，不知道RequestID/RO-06。可更新deadline的清理不得覆盖新值，通知/peek/分段共用有效期限。连接RESTART/终止通过本库Peer路径结清pending；单路径不可达及未能关联attempt的SEND_FAILED只给有限诊断，不能产生业务重试或第二次Completion。责任及实施前置见[部署§4.6–4.8](diameterkit-deployment-interop.md)。

上游具体入口沿[接口契约§2–4](diameterkit-sctp-upstream-contract.md)：结合IORequirements与适配器总上界预留prefix/通知/OOB及底层固定占用，完整帧dst仍取对应业务/控制reservation，再用独占Messages模式ReadEvent/WriteMsgContext。通知too-large可能尚未消费，但本库仍按固定NotificationMax关闭，不忙循环/无预算扩容；CloseRequired表示读边界不得复用。n>0/实际消费与ctx取消竞争时保留I/O事实，不能仅凭err把请求降成NotWritten；上游的Data/Notification成功也不代替本库完整帧/协议校验。ParseNotification变长tail借用通知区，跨读保存必须另取owned预算或只留小型值诊断；本库正常事件交接只使用固定值，不保留tail。

对已终结 exchange 的重复/晚到应答执行同一有界丢弃/关闭规则，不向已交付 Completion 追加内容，不提供无损审计通道。需要较长收答窗口的应用在 Submit 时显式配置 Deadline，并承担相应 pending/额度占用。

#### 3.5.1 上游事件到两阶段接收的交接（2026-11-04）

ReadEvent不是本库ReadPrefix的同义词。本库三个接收方法的结果须能区分PrefixReady、FrameComplete/DiscardComplete与TransportEvent；这是既有transport契约的结果细化，不增加业务通知API、第二reader或后台事件队列。TCP/TLS可只返回原有完成/错误结果。SCTP单reader按下表驱动，事件不冒充零长度Diameter帧或EOF。

| 所处阶段 | 适配器与runtime的确定衔接 |
| --- | --- |
| ReadPrefix | 用固定20字节区累计数据，ReadEvent的data仅传剩余非空部分；不足20且EOR即短帧失败。到20字节后返回不可变头、已收长度、真实EOR及首片元数据，runtime才分类取额度。20字节且EOR并不在传输层无条件拒绝，交头声明长度和后续协议校验判断 |
| ReadFrameInto | 将已收前缀复制入获准dst，后续直接填剩余区。累计到声明长度时必须已EOR；提前EOR、长度满而无EOR直接失败，不能再跨消息探读。头已构成完整帧时直接完成，不调用禁止空data的ReadEvent；仅FrameComplete且无错误才进入完整帧decoder |
| DiscardFrame | 以固定discard scratch消耗已确认的同一帧，计入原MaxDiscardBytes和FrameRead截止；同样检查长度/EOR/stream/PPID，不能因不交业务就跳过边界验证；无额外大帧分配 |
| 任一阶段收到完整非终止通知 | 就地ParseNotification，只将所需固定值转为TransportEvent并带ConnectionID/接收序号返回；不跨返回借用notification tail。runtime在短状态提交门处理后，继续原接收方法和原期限。ReadFrameInto续调使用同一dst及累计偏移，不重新复制前缀/申请额度；适配器只保存标量游标和元数据，返回后不保存dst引用 |
| 通知/数据错误或终止 | 先封住此连接的后续transport写准入，再返回带固定原因和实际进度的终止结果；runtime提交Peer事件、幂等Close并结清已有I/O。不会将需要Close的事件放进可丢失的WatchPeers队列，不能通过重入ReadPrefix“复位”坏边界 |

中间TransportEvent的字节进度是本帧累计值，不能每次再累加；FrameComplete=false时不交decoder。runtime在事件返回间持有原reservation/dst，同一连接最多一个未结清接收游标，不能让另一个调用接管。不继续时必须Close；不得释放额度后继续旧帧。续调必须匹配原阶段、目标长度和累计进度；调用者负责保持同一dst内容，错误调用不消费新数据。事件是单reader的同步返回，不在持Peer锁时执行I/O/用户回调；终止、writer完成和timer仍竞争原状态提交门。已完整读入的应答仍按原ConnectionID进入Complete竞争，不因socket已关闭就无条件抹除。

#### 3.5.2 连接期限与发送上下文

适配器统一协调上游D；每次ReadEvent使用连接生命期ctx，叠加当前握手/FrameRead/关闭的绝对上限。首个Data字节开始FrameRead计时（短头同样计时），从前缀、分类、事件返回到read/discard结束都不刷新；握手和关闭原总截止只取早。Ready且尚无数据时不凭等待ReadPrefix的时长启动部分帧超时。Tw是Peer独立timer，不直接设置为ReadDeadline；Tw到期按watchdog表发DWR、进入SUSPECT或最终关闭，通知和字节片段不刷新Tw。按[状态契约§3.3](diameterkit-state-consistency.md)处理完整合法Diameter消息的存活证据。

本库SCTP连接明确使用SO_RCVTIMEO/SO_SNDTIMEO=0，期限统一走D和ctx；上游非零S组合仍是U3兼容验收项，不能用逐ReadEvent重起的S实现整帧时限。适配器固定配置`NotificationWindow>0`和`MaxNotificationsPerWindow>0`，以每连接单调时钟窗口和一个计数器限制所有通知；达到上限后的下一通知按LocalResourceFailure关闭，不排队延迟消费。窗口仅随时间推进，不因Data/通知重置；该限制不替代watchdog和帧截止。

writer在BeginWrite后只调用WriteMsgContext，以独立FrameWrite/ControlWrite和关闭上限取早，并使用连接生命期ctx；单exchange ctx取消只撤销排队或收答资格，不接管已开始I/O。完整接纳只记Written；正数短写/不确定失败结清并关闭，零进度也须以底层明确未接纳证据才记NotWritten。上游成功与Close并发可返回真实成功，但不能据此复活Peer、重新开放发送或覆盖已选Completion。

SEND_FAILED可选关联只使用本代单调不复用的非零uint32 Context，0表示不跟踪；在BeginWrite后、调用send前登记，固定关联项计入既有attempt M，控制关联计入控制槽，不另建历史payload表。完成/释放后删除关联项但本代不复用值，耗尽后后续发送用0仅保留连接诊断，不绕回、也不为诊断强制重连。只关联仍有owned记录的本代attempt；先于writer返回的通知只合并固定事实，不能提前释放借用帧。无匹配、未知或已完成者有界计数，不按payload猜HBH，不改变重试资格或增加第二次Completion。

#### 3.5.3 固定预算、候选与关闭尾部

Node管理的传输在构造期先从已锁定适配器版本/配置得到无I/O资源需求，再分配/拨号；不能先建连接再靠实例Info决定是否有内存。Connect/Accept允许移交调用方预建conn，但此前外部分配不在Node预算承诺内；接管仍须匹配构造期需求/Info、先取得额度，失败Close，不允许未计量conn安装到Peer，见[架构§5](diameterkit-architecture.md)。IORequirements提供的是最小buffer与上游scratch，不是完整TransportReserve。适配器声明的每连接总上界须包含：prefix、NotificationMax、OOB、discard、同时存在的read/write/peek scratch、上游与适配器固定连接状态/选项副本、timer/取消hook、一个固定事件值/计数器、控制RX/TX及拒绝槽；地址等可变配置副本按已限制的输入计量。实例Info报告实际核验值且不得超过预留；版本变化使上界变化须重新校验配置，不能无预算补分配。上游只需提供可核对的固定占用/布局依据，不导入本库账本；Go运行时/内核额外开销继续单列。

prefix与完整dst在复制时并存：20字节固定区计TransportReserve，含头完整dst计原A/I/控制reservation；这不是第二份完整帧，也不把同一backing重复计费。控制RX、控制TX、notification及writer scratch可能同时使用，不能以“单reader”把它们错误取max。必要诊断只保留固定摘要/计数，通知payload不进入Completion或异步日志队列。

固定容量按`ListenerReserve + ConnSlots×ConnBound + HandshakeSlots×CandidateBound`展开；三个项是互不重叠的预留，已有控制/拒绝区只计一次。ConnSlots包含Open及已脱离Peer但I/O尚未结清的旧代；HandshakeSlots包含主动拨号、被动候选、选举败者和取消后尚未回收的候选。候选升格须原子取得活动槽并转移同一资源所有权，不能先归还候选槽再悬空持有fd/buffer。槽不足则不安装/不新拨；关闭后的旧代必须等reader/writer/hook退出才释放槽，新代同时存在须有额外真实槽。不能用稳定Peer数替代实际连接代数，也不增加无界retired列表。listener使用有界accept循环；无候选额度时先等待可取消的额度再Accept，内核backlog另行报告，不能无限accept后才记账。

第十二批固定Peer的Listener/Serve实现了先占两候选之一再Accept、无容量等待、两阶段TLS/共同截止及迟到回调屏障；空闲Accept预留不作为已收到CER的选举参与者。具体ListenerReserve公式、所有权和有限失败退避见[batch-12](../spec/history/batch-12.json)。该切片尚不提供多Peer listener/admission或动态发现。

上述计数释放与业务reservation分别判断：传输资源只在底层借用退出后释放，已交付Inbound、Duplicate缓存及Completion继续按各自最后引用计量。Node/Peer.Wait确认库I/O任务退出，不要求业务消费完成结果；不提供持久事件邮箱或重启接管。

### 3.6 调度、关闭和容量算例

writer 在帧边界调度，采用三类有正权重的 deficit round robin：Control、Answer、Request，再在各类内部按 ResourceClass 权重轮转。候选默认 QuantumBytes=16 KiB、三类权重=4:2:1、各 ResourceClass 权重=1；权重可显式修改但不得为 0，Info 返回实际展开值。每轮增加 deficit，额度足够才发一整帧；空队列清零，非空 deficit 有 `MaxFrame + quantum×weight` 上限。控制权重较高但不能永远跳过 Answer/Request；不声称能抢占已开始的 TCP/TLS 帧。常量是可测试的首版调度选择，性能默认值冻结仍需混合负载测量。

第十一批已按本节实现`NodeConfig.Scheduler`及`Node.SchedulerInfo()`：全零标量展开16 KiB/4:2:1/应急1；量子64..1 MiB、所有权重1..64；按出站/入站class分别覆盖，应急拒绝有独立Answer lane。精确API、每候选固定reserve公式、空队列立即补入的防饥饿规则、证据及性能边界见[batch-11](../spec/history/batch-11.json)。旧逐帧交替仅为前批历史；DRR不改变硬预算、原期限或一次完成。

每帧 WriteDeadline 和关闭解除 I/O 的能力提供失败上限。慢 writer 最终失败关闭，其请求按 §1 的原因分类处理；Close/Shutdown 不等待业务 handler 或完成消费者退出。Node 内 reservation 是轻量账本对象，port 可在 Node.Wait 后继续持有完成结果；这不要求遗留 reader/writer goroutine。

容量算例仅验证账本算术，不是性能默认值：若某 class 的 Outbound=16 MiB、Q+M=64 KiB、每 attempt A=2 MiB（含 owned decode 峰值）、MaxAttempts=2，则一个 exchange 预留 4.0625 MiB，最多同时接纳 3 个；不能仍宣传 128 个。读取一个 1 MiB 应答只是其 A 内从未使用额度转实际持有，不额外再扣 1 MiB。假设完成后结果仅保留 1.25 MiB，则其余额度释放；取走结果再释放最后 1.25 MiB，前提是没有库内缓存继续共享该 backing。

缓存容量另按峰值到达率 × (MaxProcessing + ReplayRetention) × 最坏 C 估算，并加入同时 Arrival 的 I/W/M。即使平均业务只需 10 ms，固定 KeepUntil 仍包含配置的 MaxProcessing；小应答不能掩盖规范内容/结果预留的容量成本。未测量前不增加弹性借用或提前淘汰。

R21 的设计保证以每个 codec 的可验证空间上界为前提；具体 Go 对象开销、默认分区和吞吐仍须实现后测量。预算失败必须可解释为哪一分区/维度不足，指标不能带用户标识或无界标签。

## 4. 设计验收与未完成事项

当前首版覆盖四应用及通用FO/VAL/BUD/生命周期条目；§1.6普通Ro重传与§1.7受检恢复、CCR/CCA及必要API全部纳入首版。发送证据、旧答案竞争、attempt/恢复记录预算和一次完成复用公共验证，并执行Ro专用16组；旧第十九轮后置豁免撤销。

| 覆盖 ID | 必测反例 / 不变量 | 阶段 |
| --- | --- | --- |
| FO-01 | 三模式互斥；Ro 只允许已审核的固定 OCS 重传；目的地越域、自动模式无合法备用、规则未支持时拒绝 | P0/P3 |
| FO-02 | Written→备用 NotWritten 仍为可能已处理；次数/总 deadline 不刷新；T 与逻辑身份正确 | P3 |
| FO-03 | 旧应答与下一 Send/取消竞态；只一次完成，保留选中应答与全部有界 attempts | P3 |
| FO-04 | Ro 不在配置策略之外补发/换 OCS；业务 Tx 不等于 exchange Deadline；有效窗口内 CCA 只完成一次，窗口后有界丢弃 | P3/P5/P6 |
| VAL-01 | Build 可缺 Origin 而不可缺业务必填；Wire 拒绝不完整消息；MissingContext 为 Unknown | P1/P2 |
| VAL-02 | NeedsApplication→Check→Ready；修改业务视图不改变凭据；未 Ready 不能 Respond 成功 | P3/P4 |
| VAL-03 | 逐规则协议上下文类型/关联/上界；业务决定不成为默认 Check 前提；E 位错误语法 | P4/P5 |
| BUD-01 | 每次 reserve/split/transfer/release 容量守恒；预算申请失败无部分持有；各种关闭路径归还 | P2/P3 |
| BUD-02 | 所有 admission 占满仍能用原 A 收回应答；新旧 attempts 同时收答，无重复计费或覆盖 | P3 |
| BUD-03 | Sh 慢消费者不侵占其他 class；OFA/控制/拒绝额度及 writer 公平；无法应答时确定关闭 | P3/P7 |
| BUD-04 | 入站 I/W/C、两代理链重复请求、SCTP 分段与重复/晚到应答丢弃；没有无预算的克隆/缓存 | P3/P5/P7 |

这些是 DRA-04/05 和 R19–R21 的细化，不取消原有覆盖 ID。第五轮又给出 R03/R17/R18 的 BIND/DUP/SCTP 细项，并补齐逻辑结果/Arrival 及规范内容预算。设计收敛不关闭实际部署、逐 Profile 规则、底层能力或实现/压力/互操作门槛。

第十八轮补充：[成对结果规则](diameterkit-mainflow-audit.md)不改变R19–R21：TERMINATION的5002、IEC的4012、RDA的实验5001/5558、ALA的基础E3004都是当前exchange的结果，不自动授权重传或下一业务操作。合法终止/退款应答无GSU不能按初始授予规则判坏包；错误答案先依据E位选择语法。原token、outcome叶和未知扩展占用既有owned内存/解码预算，不另建持久缓存；未决SMS退款Application语义不会延长网络等待。


第二十轮[Sh/SMS候选](diameterkit-protocol-vectors.md)给首版FO/VAL/BUD提供具体输入：每组成功/失败是替代答案，不能重复完成；SNR与PNR是独立exchange，PNR处理不等待SNA。XML解析显式执行，原bytes与解析结构并存计入既有预算，不能绕过MaxAnswerBytes/owned内存边界。ALA/TFA失败及重传时间不启动业务重试；通用TransportFailover仍严格受本契约已有条件约束。


第二十一轮[订阅及特殊身份候选](diameterkit-protocol-vectors.md)不改变发送/完成契约：退订2001不构成PNR全局屏障，退订期间的PNR按独立入站exchange处理；多Ref失败不启动库内补偿请求。Sh身份冲突、短信URI父路径和结果错误在对应校验阶段报告；raw/XML及多个字段的累计owned预算继续受既有上限控制，无订阅或MWD资源池。


第二十二轮[裁决](diameterkit-mainflow-audit.md)不改变网络完成点：合法关联SRA即使应用身份语义未支持也结束等待，保留原结果和Report而不自动重发。无法构造有依据的成功SRA在Respond提交前失败，既有入站应答权不因本地构造失败被消费。Sh混合SNR是一个exchange，组合PNR也是一个独立exchange；应用主动拆分的两个SNR分别计入既有预算/期限与一次完成，没有全成全败、自动补偿或额外订阅池。XML解释的多个片段与原bytes仍按既有有界owned预算计量。


第二十三轮[字段/结果边界](diameterkit-field-inventory.md)继续使用已有预算：111行叶、所有父组/未知扩展、原XML与解释结果按字节/数量/深度计量，无每个业务字段独立资源池。合法关联的4100/4101/5550等结果不能延长pending或自动重试；HSS序号/订阅变更归应用。UTF8/类型/字段条件错误按既有VAL结束/拒绝规则处理，raw能保留不表示应用语义通过。


第二十四轮[时间/结果候选](diameterkit-leaf-vectors.md)不改变运行时完成点：合法关联答案的业务失败或应用条件报告不延长等待，也不自动安排3331指定的重投递；Time宽度错误按既有静态无效答案路径终结。ALR业务时间已过与网络deadline不同，原值/报告保留。136完整帧仍未验证codec、资源上限或并发运行行为。

第二十五轮[特性/T4候选](diameterkit-leaf-vectors.md)不改变完成点：实验5011与基础5001按各自结果交付；规范允许的另发共同特性请求由应用决定，不延长原pending或隐式迁移预算。T4一跳不增加转发/MTC业务状态；静态无效答案与应用报告仍走既有分阶段契约。

第二十六轮[路径/实例反例](diameterkit-command-path-matrix.md)继续遵循既有校验完成点：错Vendor导致缺必需叶、错误父实例及应用禁位分别报告，不用其他组中的值修复消息。业务OFA5555/cause4不由本库网络超时自动生成；不新增持久状态或重试。

第二十七轮[四项P1边界](diameterkit-mainflow-audit.md)不改变完成点：Sh组合违例、SRA身份范围、Realm和时间未决报告不触发持久查询/自动重发；有效关联答案照常终结等待，不对答案回错误答案。Respond前的本地构造失败不消费应答权；位于标准子集之外不等于远端业务失败。原值、字段Presence、显式era及Report仍按既有owned/预算契约保留，不自动建定时器或Peer。

第二十八轮[九答案族](diameterkit-answer-envelope.md)补具体完成/预算输入，顺序见[主流程§24.3](diameterkit-mainflow-audit.md)：R分类与完整头匹配先于借用attempt的A；不匹配头使用固定discard额度，超限关闭按真实关闭原因处理。正常与E的有效基础/应用失败都完成一次，不启动原身份重传；E5允许不改变业务失败分层。旧代已完整读入且带原AttemptID的候选按原R19门竞争，不重新标为新ConnectionID。E错误省应用成功字段不豁免Session/Origin/来源/Proxy规则。

第二十九轮[诊断与扩展预算](diameterkit-mainflow-audit.md)沿用既有A/owned和Reject槽：未知opaque计全部字节但不虚构子项；279已知可读层计深度/节点，遇坏内部长度停止分支并保留有界raw。不得依错误Length分配或跨容器读；外围坏包、内部诊断不完整、本地ResourceLimit分开。日志也受同一界限。无法在拒绝预算保留必需包络/证据则沿用关闭路径，不扩展持久缓存或自动恢复。

第三十轮[公共字段覆盖](diameterkit-mainflow-audit.md)不改变A/Reject槽和完成门：已关联答案的278宽度/重复、297父子结构、外围281/294类型错误按ErrInvalidAnswer；合法未知结果交付范围报告。UTF8文本按octets计空间，显示转义也必须有界；不按对端文字启动恢复，不维护Origin状态历史库。

第三十一轮[请求拒绝边界§27.4](diameterkit-mainflow-audit.md)区分确认缺Session与非法/重复Session、坏Proxy：前者基础E语法可不带原本不存在的263，后者若无法构成可信合法回显则沿用关闭/终止，不修复原值。出站DestinationNotAllowed发生在接纳前，无发送/Completion；入站合法包络的目的拒绝才产生E3002。父叶静态失败仍不绕过预算，不新增应急队列或持久恢复。

## 第五十批执行更新：普通R33恢复已接入Ro

本节§1.6范围已由四Ro Profile编译许可和既有Route实现，根运行时无需新增计费状态。2..16次/固定终端/原deadline/一次完成保持；每连接内分配HBH，不要求跨连接的HBH数值全局不同。§1.7单段一次发送、ResponseWait、换OCS/错误后续发/重启入口仍须独立实施。见[本批契约](../spec/history/batch-50.json)及[进度](../spec/progress.json)。

## 第五十一批执行更新：准备与发送分界

原记录的准备入口已落地，必须显式配置RoRecoveryLimits和RoRecoverySlots/Classes；默认顺序源也检查恢复保留。准备不触发网络/完成，原请求帧归应用持有，Node仅保留摘要guard。多目标预检覆盖最长帧及每目标绑定，目的字段外逐字节保持；时间/次数上限冻结，但未运行Tr或消费段额度。后续SubmitRoRecovery必须先补齐记录重新校验、标识接纳/高水位、段预算和一次完成，不能直接复用普通Submit的重新发号。见[batch51](../spec/history/batch-51.json)。

## 第五十二批执行更新：有限段复用公共exchange

段接纳只选择本Node准备记录的第一个目标及完全一致绑定的有限Peer路径；一次attempt段不受普通Route构造器“至少两次”的限制。原身份由guard受检复用，绝不调用普通Submit重新发号。可续段范围仅本地已结清的无答、传输/看门狗/取消/段截止；所有答案及关闭在本切片终态。固定元数据保留SegmentNo/额度高水位和累计事实，依赖现有收发借用屏障清理后再交付唯一Completion。

ResponseWait/Tr尚未启用，Plan.Deadline仅为本段总期限。第1.7节规定的有限错误续发、跨OCS/重启等需求不撤销；下一批先接Tr和同段晚答，见[batch52](../spec/history/batch-52.json)。

## 第62批资源实施补充（2026-11-09）

[U5账目](../spec/history/batch-62.json)提供当前上游候选的8192字节逻辑固定集成额度，适用单reader/单writer、无上游排队、受控标准context。含源值、收发scratch、一个通知值及有界channel/context/file槽策略，不等于heap/RSS上限。adapter还须叠加自身值、prefix、NotificationMax/OOB/discard、独立控制/拒绝槽；Go/内核额外开销单列。建立地址副本按N/Z/D前置限制另计，listener和每候选独立占槽，handoff只转移。此补充是接线依据，当前尚未新增SCTP Info/TransportReserve实现或发布版本。

## 第64批实现注记

实现原§3.5.1三阶段续调、无tail/无第二帧、固定通知大小/速率及不刷新的帧期限。默认Info16432=8192上游+4096adapter+4096通知+48OOB，Node每候选另计2048事件快照，原控制/应急槽独立。仅预建socket接管，建立N/Z/D与listener额度仍待下一批。具体范围/证据见[batch64](../spec/history/batch-64.json)，历史未实施表述按对应时点阅读。

## 第65批实现注记

已补SCTP托管建立的N/Z/D前置限制与复制隔离。设A=64+96N+Z+2D+16(N+2)，建立额度S=4096+4A；默认N16/Z256/D4096，listener=8192+4096+S=57728，候选=原Config.Requirements()+S=61872。listener及每物理候选相加，不以Ready连接数代替接纳/退休中连接；核心2048事件额度及控制/业务额度另计。连接交接后继续保守占候选声明额，不退还后又透支；DNS内部、allocator/栈、内核队列不计入逻辑承诺。

Dial/Resolve/CER共用同一绝对期限。固定上游没有AcceptContext/listener deadline setter，adapter将listener SO_RCVTIMEO设20ms，通过现有poller分段等待；ctx取消响应受该片段及调度影响，不承诺硬实时20ms。每次返回后复查取消/Close，迟到socket关闭。New清零已接纳socket的SO timeout；无旧ctx钩子污染建立后读写。Serve/Incoming交接、Close与Wait的真实验收见[batch65](../spec/history/batch-65.json)。

## 第65批后上游交接（未实施）

第65批57728/61872额度及20ms接受等待是当前代码基线，不再作为长期上游/adapter分工。待上游A–D交付后改用其原生Accept取消及资源查询，建立预算按新账目重新计算；不得既加旧8192/地址公式又重复加入新上游总额。Diameter帧期限、通知限额、控制/业务预算和Wait退休不变。计划/边界见[上游下沉需求](sctp-upstream-improvement-plan.md)。

## 第66批实现注记

A–E上游增强及消费方迁移已完成。默认amd64 Config.Info17756、listener52620、每候选99676，来自上游bounded-setup-v2查询加本库固定状态/通知/OOB/建立包装额度；每候选核心2048及控制/业务额度另计。旧8192和N/Z/D公式不再参与当前计算。Accept通过上游poller取消，无20ms循环；RESTART/partial-abort有效通知即使同时携带终止标记也保留事件事实。旧版本数字/待实施表述按历史时点阅读，当前契约及证据见[batch66](../spec/history/batch-66.json)。
