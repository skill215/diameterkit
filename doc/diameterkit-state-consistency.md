# 协议上下文、Peer 与进程内标识契约

> 第67批当前状态：独立adapter已锁定正式SCTP v0.8.0（bce1b3cf3ea3），源码与第66批候选一致，A–E、资源和Peer边界不变；正式版本回归完成，见[batch67](../spec/history/batch-67.json)。上游发布门已关闭，核心/adapter公开发布及现场验收仍单列；下文旧候选/未合并表述按历史时点阅读。

> **当前范围（2026-11-04，Draft 0.20）：** 第一阶段=首版/v1，包含Sh、S6c、SGd/Gdd、Ro/DCCA及sms-fw实际所需六模板、SCAP1和RO-06恢复能力；Go 1.26、纯Go实现（含依赖与SCTP适配器，发布路径CGO_ENABLED=0）、公开Go库、无持久化边界不变。明确撤销第十九轮“Ro后置”分期；带日期的旧分期/统计仅作历史证据，不再授予Ro跳过首版验收的例外。限定S6a、完整SCUR/PRICE_ENQUIRY等未承诺能力仍后置。范围及剩余缺口见[Specification §3/§7](diameterkit-specification.md)，顺序见[总设计§14](diameterkit-design.md)，验收见[覆盖§39](diameterkit-usage-coverage.md)。2026-11-06已完成前十九批有界实施（Base/S6c基础、公共包络与关联、出站运行时、关闭与选举、TCP/TLS生产适配、有限Inbound/应答、有限Reject/应急预算、Canonical/有限重复缓存及多Arrival重答、固定同终端Route/多attempt与活动请求原身份重传、有限Dialer/Maintain拨号及重连、按class/编码字节DRR调度、固定Peer的有限Listener/Serve、有界标识源/Session/guard、Base/S6c有限Context/Report、S6c普通RDR/RDA及HSS ALR/ALA、SGd/Gdd普通OFR/OFA与TFR/TFA、Sh普通四组命令及透明User-Data、Ref24/25有限Sh-Data helper、Sh单订阅/退订及Ref11/12有限数据）；已执行范围及未完成项以[Specification §6](diameterkit-specification.md#6-需求与证据的维护)和[spec进度](../spec/progress.json)为准，未列明的运行验收仍Planned。

> **第十三批实施注记（2026-11-06）：** 已接通有界IdentifierConfig/EndToEndSource、NewSessionID、内置分区轮次guard、自定义class配额及固定跟踪表，原身份重传共用ID，HBH保留区触发连接排空轮换。实际API、预算和关闭边界见[batch-13](../spec/history/batch-13.json)，检查见[spec进度](../spec/progress.json)。跨进程唯一性、RO-06恢复导入、具名上下文/报告、完整四应用及SCTP仍沿原边界，未由本批宣称完成。

日期：2026-10-31（第二十一轮补号码/订阅与特殊身份切片）。第十二轮按 [主流程审计](diameterkit-mainflow-audit.md) 明确字段选择与结果边界；第十一轮按 [Profile 规则契约](diameterkit-profile-rules.md) 收敛协议输入；第十轮收敛 R32/R33/R36；第九轮复核公开库边界；第八轮已按用户明确定位修订：这是公开 Go Diameter 库，**不提供或要求持久化组件**。保留 R24 的协议上下文和 R29 的 Peer/标识能力；撤销 R28 的 Store/Resolve、持久 owner、标识 provider/区间封存以及 managed charging 设计。撤销表示超出库边界，不表示这些机制已经实现或验证。

[总体设计 §1](diameterkit-design.md) 是职责边界来源。Go 1.26、DRA 场景、SCTP 能力和既有消息/字段兼容义务保持；业务工作流、业务幂等、数据库、重启恢复及跨进程协调由应用负责。本库运行不需要数据目录、数据库、WAL 或存储适配器。

第十八轮更新的库级需求与验收索引见 [Library Specification](diameterkit-specification.md)；本文保留状态、时序和预算细节，不因规范草案建立而视为实现通过。

## 1. R24 / R31：具名协议上下文，不建业务事实框架

### 1.1 保留的输入与职责

第十一轮将调用方 Context 与内部校验证据分开：Role/Direction 来自 binding，SRR 用途和 CC 类型/编号从当前/原请求快照推导，不允许公开输入覆盖。validation.Context 是由应用生成构造器产生的不可变有界值，首版保留有规范用例的sh.PeerFeatures及ro.SessionFeatures；无额外输入时用零值。每项须有 RuleID、条款、关联/来源、缺失行为和空间上界，没有通用 ApplicationFacts 容器或业务状态准入门。Ro SMS 的 OM/OC 字段选择属于纯构造参数，不扩充 validation.Context 或接收 Ready 的业务前置条件；对端配置未知不能以本地字段省略选择代替。最短Service-Context-Id的OM/OC可选例外由当前/原CCR判定；带运营方形式所需规则未明时报告应用检查缺依据，不增加在线查询或第三类Context。

库从自己的原请求快照取得关联字段，复制确需保留的有界上下文，拒绝类型/身份冲突。Inbound.Check 只检查已声明的协议条件；无待补条件时直接 Ready。Ready 表示这些协议检查完成，不表示获准投递、订阅有效、余额足够或数据库提交成功。库不要求普通 PNR/OFR/ASR 先查询数据库才能答复。

### 1.2 本轮撤回的泛化能力

第七、八轮的通用 FactSet/FactSubject、SendFacts、RespondWithFacts、TemporalUse、Revision/ObservedAt/NotAfter 和 ClockUncertainty/ClockUntrusted 框架尚无逐规则必要性证据，**从当前首版公开 API 撤回**。撤回不是已完成验收，也不删除规范本身要求的时间 AVP。数据库版本、订阅 Found/NotFound、业务授权期限、发送许可和提交确认由应用处理；不换成另一个通用回调框架。

保留 RequestOptions.Validation、Inbound.Check、Inbound.Respond 和纯结果解释器。成功 Check 冻结该次协议检查输入；再次相同输入幂等，不同输入报 ContextConflict，不能修改原请求关联。应答结果分支由应用明确构造，Respond 检查它与原请求及固定规则是否一致；不要求应用把同一结果再包装成业务 facts。

若逐 Profile 审计证明某条消息确实需要回答时的额外协议上下文，先提供条款和最小反例，优先在对应应用的纯构造/解释器中增加具名参数；只有运行时必须持有时再评审最小 endpoint 扩展。不能先公开空泛框架再等待需求。未完成这类规则映射时明确列为不支持组合，而不是静默跳过校验。

### 1.3 时间与空间边界

请求 Deadline、watchdog、缓存期限和写入/关闭期限仍由 runtime 管理。Event-Timestamp、Expiry-Time、SM-Delivery-Timer 等线路字段按固定规则编码/解释；Time era 单独审计，不等于通用业务 facts TTL。Check 的历史输入不因墙钟推进而自动失效，重复应答复用已提交结果，不重新查询业务状态。

出站上下文计入 Q/M，入站检查快照计入逻辑 I；C 保存规范内容、已提交结果及必要元数据，不再预留独立 AnswerFactsBound。没有通用事实审计日志、刷新任务或时钟可信度服务。Node 结束不恢复这些状态。

### 1.4 应用最小职责表（四应用首版）

| 用例 | 库检查/返回 | 应用负责 |
| --- | --- | --- |
| S6c SRR/SRA | 原查询用途、身份语义、节点组合和结果解释 | 查询授权、最终投递路由策略 |
| SGd/Gdd OFR/OFA | 请求关联、结果分支及诊断字段组合 | 短信投递、外部队列、幂等、选择成功或失败结果 |
| Sh PNR/PNA | 标识/载荷/结果语法及有条款依据的关联条件 | 订阅记录是否存在、是否消费通知、状态更新 |
| Ro CCR/CCA | 正常CCA的原请求CC类型/编号、Session及服务级结果；E错误按独立包络关联，不强制CC字段 | 会话工作流、退款、计量与业务结果 |
| Ro ASR/ASA | 原ASR的头/Session/来源关联与ASA结果；不要求CC类型/编号。2001表达应用报告已终止，库不独立核验外部状态 | 实际会话终止、2001/5002/5012结果选择及与TERMINATION的业务竞争 |

第十一轮已在 Profile 规则契约 §4–5 给出首批 RuleID、具名输入、无导入环依赖方向与四流程推演。R24/R31 的接口边界设计收敛；全量应用规则、实际编译原型/fixture 仍为 P1，不能把文档签名视为已通过的 API。

## 2. R28：从库范围撤销持久化协议

不交付 charging.Engine、Store、IntentRef/Receipt、Resolve、Claim/Release、outbox、ScanSnapshot、owner fencing、数据库/provider 适配器或持久状态机；它们也不作为公开库的可选首版组件。第七轮有关提交确认的具体接口及故障矩阵撤回，不再阻塞本库 API 冻结。

公开 Route.Prepare/Stage/StagedAttempt.Send/Discard 是为该持久化操作流程引入的接口，本轮同时撤回；内部 attempt 仍由 Request/Submit 的 exchange 引擎管理。app/ro 的纯消息构造器（如 PrepareSMSDebit）保留，它们不联网、不发线路标识、不保存记录，与撤回的 Route.Prepare 不同。

应用仍可保存库返回的 owned 消息、RequestIdentity、Attempts 和发送证据；何时写入存储、如何确认提交、是否重发或补偿，均是应用责任。库不规定业务存储格式，也不在启动时扫描旧记录。**当前没有恢复旧进程 exchange 或导入持久 operation 的公开契约**；第十轮已给出运行中 Node 内的受限原身份重传契约，见运行时契约 §1.6；这不恢复跨进程业务 operation 导入。再次 Request（新 exchange）不能声称恢复原操作。

晚到结果仍采用唯一 CompletionPort：应用使用 Submit 设置覆盖所需收答窗口的 Deadline；业务自己的 Tx 到期不必调用 Cancel。exchange 未完成时到达的有效 CCA 可正常完成；exchange 已终结、Port 关闭或 Node 退出后，不承诺继续交付。重复/无匹配应答有界丢弃并计数，没有持久证据邮箱或第二次完成事件。

## 3. R29：Peer、watchdog 与管理意图

### 3.1 状态维度与分用途发送门（R32）

Peer 是 Node 内稳定对象。PeerInfo 同时展示 `Lifecycle(Active/Draining/Closed)`、`ProtocolState`、`WatchdogState`、`ReconnectState(Allowed/Suppressed)` 和 ConnectionID/原因；不把全部组合平铺成巨大公开枚举。Closed 的协议状态可重连，Closed 的 Lifecycle 不可复活。

`Ready = Lifecycle.Active && ProtocolState∈{I-Open,R-Open} && WatchdogState==OKAY && 已完成本地 CEA 写屏障/对端 CEA 校验 && 准入/能力满足`。新请求选路及未进入排空阶段的 Request writer 共同检查 Ready；应用能力不足只关闭对应应用资格，不借 Relay 广告改变 Profile。Peer 事件、timer 和 writer 完成携带 ConnectionID/事件序号，在同一状态提交门处理；旧代或已取消 timer 事件丢弃，锁内不执行网络/业务 I/O。

Ready 不是所有帧的写入总开关。以下门限在 writer 真正开始时复核，均要求对应 ConnectionID 仍有效、帧预算已取得且未超过写/关闭期限：

| 用途 | 额外门限 |
| --- | --- |
| 新应用 Request / 新 failover attempt | 必须 Ready；Draining 不接纳新 exchange，也不启动新 failover |
| 本地 Shutdown 前已接纳的首次 Request | 仅在 DrainDeadline 前、尚未发出/收到 DPR，且 Protocol Open、watchdog OKAY 时可继续；其余 Ready 条件仍适用 |
| 应用 Answer / 已收到副本的缓存重答 | 必须有本代真实 Arrival；允许 SUSPECT 和本地 Draining 的已有工作；进入 Closing 后不再开始应用写；REOPEN 没有新业务发送权；不因 Ready=false 一概丢弃 |
| CER/CEA/DWR/DWA/DPR/DPA | 由各自 Peer/watchdog 状态迁移授权，使用控制预算；允许握手、REOPEN、Closing 的必要帧，不能要求先 Ready |

Connect/Accept 成功只表示该稳定 Peer 的活动连接已完成准入、选举和能力交换；不保证返回时 Ready。重连可能仍在 REOPEN。SCTP 的 CEA 后模式屏障只限制切换无序/多流，未切换时使用有序 stream 0；它不阻止必要 DWR，也不额外要求首次握手完成三轮探测。通过 Info/Snapshot/WatchPeers 观察当前可用性；Request 在无可用连接时正常拒绝。Connect/Accept 的 ctx 只控制本次连接建立，不等待三轮 watchdog，也不在成功返回后接管连接寿命。

### 3.2 连接与选举

内部保留 RFC 的 Closed、Wait-Conn-Ack、Wait-I-CEA、Wait-Conn-Ack/Elect、Wait-Returns、I-Open、R-Open、Closing 状态。下面按实现分支归并事件；未列的规范事件按 RFC 6733 §5.6 表处理，非预期应用帧不“帮忙”跳到 Open。测试必须覆盖该表全部事件，而非只测顺利握手。

| 场景 | diameterkit 的动作和资源边界 |
| --- | --- |
| 初始/重连主动连接 | Closed→Wait-Conn-Ack；在 ConnectDeadline 内完成传输/TLS。成功发送 CER→Wait-I-CEA；合法成功 CEA 后 I-Open，失败或非 CEA/超时关闭该候选 |
| 被动连接 | 独立有界握手槽先读 CER/准入；没有竞争时发送 CEA→R-Open；本地 CEA 完整写出前不发应用帧，失败不得发布 Ready |
| 主动尚未成功又收到 CER | 保留一个被动候选，进入 Wait-Conn-Ack/Elect；主动失败则完成被动握手，主动成功发送 CER 后进入 Wait-Returns 并选举；不再收第三条候选 |
| 等 CEA 时收到 CER | 进入 Wait-Returns；按经过准入的两端 Origin-Host 做 ASCII 不区分大小写比较。字典序较大的一方获胜，关闭自己主动发起的连接、保留被动连接；另一方保留主动连接。相等身份拒绝自连接/身份冲突 |
| 等选举结果时 | 主动 CEA 获胜路径关闭被动候选；主动断开可完成已有被动路径；被动断开继续等主动 CEA；总握手期限不因竞争刷新。候选关闭只释放候选资源，不关闭稳定 Peer |
| 已有 I/R-Open 又来连接 | 拒绝新候选，不热替换活动 ConnectionID；明确管理关闭后才重建。临时选举的两条连接计入握手上限，不是两条活动业务连接 |
| 错误/超时 | 按原因关闭、释放握手/控制资源；协议、证书、身份或配置失败转 Reconnect Suppressed，普通传输失败可重试；显式 Close 永久结束 Lifecycle |

选举方向依据 [RFC 6733 §5.6.4](https://www.rfc-editor.org/rfc/rfc6733.html#section-5.6.4)。完整对端状态表为该 RFC §5.6；本项目补充候选资源、代次、超时和 CEA 写屏障，不能只按 TCP 建立先后决定赢家。

### 3.3 Watchdog 的确定行为

采用 INITIAL、OKAY、SUSPECT、DOWN、REOPEN；每个代次最多一个排队或在途 DWR，关联必须匹配本地发出的 HBH/E2E/命令。默认 Twinit=30 s，允许配置但不得低于 6 s，每次重设加独立 ±2 s 抖动；不因普通请求超时而直接判 Peer dead。参考 [Erlang/OTP watchdog 配置](https://www.erlang.org/doc/apps/diameter/diameter.html) 与 [RFC 3539 §3.4/附录 A](https://www.rfc-editor.org/rfc/rfc3539.html#section-3.4)。

| 当前状态/事件 | 动作 / 下一状态 |
| --- | --- |
| INITIAL，首次能力交换成功 | OKAY，启动 Tw；首次业务可用，不要求额外三个探测 |
| DOWN，重新完成能力交换 | REOPEN，NumDWA=0，按下述发送子步骤安排一个新 DWR；此时仍不可选路 |
| OKAY，合法消息到达 | 重设 Tw；仅匹配 DWA 清 Pending，其他消息不假装回答了 DWR。无效帧/未匹配 DWA 不刷新存活证据 |
| OKAY，Tw 到期且无 Pending | 排入控制 writer 一个 DWR；writer 开始前处于 ProbeQueued，不把本地排队当未答探测；发送子步骤见下文，不重传这个 DWR |
| OKAY，Tw 到期且 Pending | SUSPECT；移出 Ready，通知 exchange 传输不可用一次，重设 Tw；暂不只因这一次未答探测立即关闭连接 |
| SUSPECT，合法消息到达 | 回 OKAY、重设 Tw；匹配 DWA 清 Pending。已启动的其他 attempt 不撤销历史证据，也不把旧请求重新自动迁回 |
| SUSPECT，再次 Tw 到期 | 关闭连接→DOWN；清本代控制 pending，关闭 writer 后结清其证据 |
| REOPEN，匹配 DWA | 清 Pending，NumDWA++；达到 3 才 OKAY 并重设 Tw，未达到时保持原 Tw，等其到期发送下一探测 |
| REOPEN，Tw 到期无 Pending | 按发送子步骤安排下一 DWR |
| REOPEN，Tw 到期仍 Pending | NumDWA≥0 时置 -1 并再等一轮；NumDWA<0 时关闭→DOWN。不重复发送 pending DWR；迟到匹配 DWA 从 -1 加到 0，不立即恢复 |
| REOPEN，其他消息 | DWR/DPR 等基础控制正常处理但不增加 NumDWA；应用帧有界丢弃，不交业务、不缓存待就绪。此处不能为探测期间请求产生第二次业务处理 |
| 已建立代次收到终止性传输事件 | DOWN，发布不可用一次；首次能力交换尚未成功的候选失败保持 INITIAL。Close/Stop 后不再安排 watchdog，状态事件与协议错误原因分开 |

DWR 发送子步骤：创建控制匹配记录及 ProbeQueued 后暂停本次 Tw，ControlWriteDeadline 包含排队期限；writer 开始时原子置 Pending、取消 ProbeQueued 并启动新的 Tw，然后写帧。匹配记录先于字节可见，快速 DWA 可清 Pending，迟到 writer 完成不得重新置位。ProbeQueued 失败为本地控制进展错误，不伪装成 DWA 超时；实际传输写失败仍按 R19 分类。每个连接最多一个 ProbeQueued/Pending；ProbeQueued 期间合法消息只更新存活记录，不重新安排 Tw 或第二个探测，由控制任务截止或 writer 开始结束此阶段；旧 Tw 事件按序号忽略。

SUSPECT 仍可收答或回复已取得发送权的 Arrival，不能接纳新的出站业务写；完整合法入站消息使其回 OKAY 后才按普通入站流程处理。REOPEN 没有继承旧 ConnectionID 的 pending/Arrival；旧代已经读完的候选按 R19 完成门处理。watchdog 不对业务错误码/应用存储超时负责。

进入 SUSPECT 是 R19 的确定传输不可用触发，不要求等待 TCP EOF。TransportFailover 可在结清当前 writer 后选备用；SingleAttempt 结束为相应传输证据，ExplicitRetransmit 在 writer 结清后进入 AwaitingRetransmit；Ro 的模式边界见运行时契约 §1.6。若结清正在写的帧需要关闭连接，那是独立的写入安全原因，不把它伪装成“第一次 DWA 超时就关闭”的 watchdog 规则。

### 3.4 Timer、DPR 与重连所有者

> 第十批已实现本节的有限管理切片，见[batch-10](../spec/history/batch-10.json)。实际Dial使用既有Timeouts.Handshake作为传输/TLS/CER共用总上限；MaintainPolicy逐Peer显式指定退避/抖动/稳定期，全零不启用Maintain，不隐式采用下表候选30秒。第一次失败等InitialBackoff，按此前连续失败数计算指数；固定Peer的Listener/Serve已在第十二批实现，完整事件表及多Peer admission仍未完成。

| Timer/事件 | 所有者与截止行为 |
| --- | --- |
| ConnectDeadline / HandshakeDeadline | Peer 管理者拥有；每个连接候选绝对期限，TLS/CER/选举共用已配置总上限，不因收到半帧刷新；完成/候选关闭取消 |
| Tw | 当前连接的 watchdog；只由上表事件重设；关闭取消，代次/序号检查过滤迟到触发 |
| FrameRead / FrameWrite / ControlWriteDeadline | transport/runtime；部分帧不能无限续期，失败结清 I/O 并关闭；控制排队等待也有界，不依赖业务释放额度 |
| DrainDeadline / DPADeadline | Lifecycle 管理者；Shutdown 先禁止新工作，在预留关闭窗口之前停止排空，再发 DPR 并等 DPA，使用下述同一总关闭上限；DPR/DPA 不证明远端完成业务 |
| Tc / 重连退避 | 只有 Maintain 可自动调度；Tc 候选初值 30 s，退避 `min(Tc×2^失败次数, MaxReconnect)` 加非负有界抖动；首次显式连接可立即尝试，连续 OKAY 达 StablePeriod 才清失败计数 |
| 收到 DPR REBOOTING | 停止新发送，回复 DPA 后结束本连接；Maintain 可在 Tc 后重连，保留原业务证据 |
| 收到 DPR BUSY / DO_NOT_WANT_TO_TALK_TO_YOU | 回复 DPA、关闭，置 Reconnect Suppressed；不自动定时探测。停止原 Maintain 后由调用方显式 Connect/重新 Maintain 才解除抑制；来源仍须准入 |
| 远端 DPR / 本地 DPR 交叉 | 本代 DPR/DPA 分别关联；匹配 DPA/EOF/截止先到即关闭，未完的本地 DPA 仅结清真实证据；不刷新截止或等待业务 |
| Maintain ctx 结束 / Peer.Close | 前者停止管理和当前连接，Peer 保留；后者 Lifecycle 永久 Closed。两者都使已安排的重连事件失效，不自动 failover 主动取消的 exchange |

Shutdown 在调用时固定 TotalDeadline（ctx 截止与配置总上限取早），并从剩余时间中预留 CloseReserve。`DrainDeadline = max(now, TotalDeadline - CloseReserve)`；排空提早结束即可提前进入 Closing。到 DrainDeadline 停止尚未开始的应用 Request；进入 Closing 后已有 Arrival 也不再开始新的应用写。已开始帧只能结清，必要时关闭，不能插入 DPR 破坏 TCP 帧或 SCTP 消息。可安全发送时 DPR/DPA 的排队、写入和等待均截到 TotalDeadline；收到匹配 DPA 立即关闭，响应方只在剩余 ClosingGrace 内等 EOF；截止直接关闭，不刷新预算，也不承诺时间不足时一定送达 DPR。Node 级关闭的所有 Peer 共享该总截止，不能串行各耗一次预算。Close 继续表示立即关闭。

收到远端 DPR 立即停止新应用 Request（包括本地排队者），及时发送 DPA；已开始写入先结清或关闭，不再开始应用应答写；ClosingGrace 只用于已发送 DPA 后等待 EOF，收到匹配 DPA 则立即关闭。此处与本地主动排空的首次 Request 例外分开。

Reconnect Suppressed 仅禁止自动主动拨号；已配置且通过准入的远端新 CER 仍可建立被动候选，除非 Lifecycle 已关闭。被动连成不自动清除主动拨号抑制；下一次断线仍按原管理意图处理。

收到 DPR 时不再接纳新的业务；已接纳 pending 保留真实证据，若无法完成则终结/交由应用判断业务后续，不把所有 DPR 自动归为可重放的传输故障。无公共 Resume API：受抑制管理任务停止后的一次明确 Connect/Maintain 是管理意图变更，Node.Close/Peer.Close 则不能恢复。

### 3.5 R32：状态提交、写入与关闭竞争的最终规则

本节收敛的是库的并发语义，不新增一套公开 FSM 插件。ProtocolState 继续使用 §3.2 的 RFC 状态，Lifecycle 与 watchdog 分开；`Draining` 是发送 DPR 前的本地阶段，不能与协议 `Closing` 混同。

**三个线性化点。** Admission 原子登记 exchange/Arrival 及额度；BeginWrite 原子复核 Node/Peer 生命周期、ConnectionID、exchange/Arrival 发送权、deadline 和预算并取得唯一 writer 使用权；Complete 原子选定一个最终结果。排队不是 BeginWrite，拿到快照也不是发送授权。实现若分锁，固定次序为 Node 接纳门→Peer 门→exchange/Inbound 门→额度门；禁止反向持锁调用，锁内不做源回调、编解码或 I/O。纯 Peer 事件不在持 Peer 锁时同步调用 exchange；输出带代次的事件再提交。下一 attempt 与旧应答仍竞争同一 exchange 门。

取得 BeginWrite 后在锁外写入；Close 可以并发关闭 fd。若库尚未调用底层 Write，可证明 NotWritten；进入底层后只能采用其实际返回证据。先撤销所有未开始发送权，再关闭 transport，最后等正在借用缓冲区的读写退出才释放；不能靠取消 ctx 提前归还 I/W/A/Q。transport.Close 必须幂等、解除读写，独立适配器用相同契约验收。

| Peer/连接阶段 | 新应用 Request | 已接纳首次 Request | 应用 Answer | 基础控制/接收 |
| --- | --- | --- | --- | --- |
| 握手/选举 | 禁止 | 禁止 | 无 Arrival | 仅状态机允许的 CER/CEA；非预期帧失败，候选有固定期限 |
| Active + Open + OKAY | 按绑定接纳 | 可开始 | 可答真实 Arrival | DWR/DWA 正常；每代一个探测 |
| SUSPECT | 禁止；发布一次传输不可用 | 尚未开始者不得写 | 可答已存在 Arrival | 可收答/处理控制；有效入站恢复 OKAY 后才交新业务 |
| REOPEN | 禁止 | 禁止 | 不创建应用 Arrival | 只做恢复探测/控制；应用帧有界丢弃 |
| 本地 Draining，协议仍 Open | 禁止新 exchange、新自动/显式重传 | 仅首次 attempt、OKAY、DrainDeadline 前可开始 | 已有逻辑工作及已接纳拒绝可答；同键副本仍受原期限/额度 | 不接纳新逻辑业务，按固定 FailurePolicy 有界拒绝；继续匹配已有应答和必要 DWR/DWA |
| Closing（已决定发 DPR 或收到 DPR） | 禁止 | 禁止 | 不再开始任何应用应答写；逻辑结果仍可提交到 Node 缓存 | 不创建新 Arrival；仅处理关闭控制及已经完整收到的匹配应答，见下表；不能重新回到 Ready |
| 连接终止/Lifecycle Closed | 禁止 | 禁止 | 本代发送权终止；其他 Peer 的真实副本另判断 | 旧事件仅用于结清引用和证据，不能复活状态 |

SCTP CEA 后模式屏障独立于本表：首次 Open/OKAY 可先在有序 stream 0 发业务，取得屏障证据后才切无序/多流；REOPEN 仍须自己的探测完成。DWR/DPR 不经业务重传调度，DPA/CEA 不能被业务队列等待条件卡住。

Closing 对已有匹配应用应答和交叉 DPR 的处理是本库针对在途事件的有限竞态规则，不增加 RFC 状态或开放新的业务发送。新应用请求、DWR、无关 DWA/DPA 有界丢弃/计数，不新建 Arrival、不安排 DWA/探测；CER/CEA 等握手帧出现在已关闭握手阶段则按协议异常直接关闭。坏长度/损坏 framing 始终直接关闭，不能为继续等 DPA 越过帧边界。

**关闭时序。** 首次 Shutdown 固定 `TotalDeadline=min(ctx截止, now+ShutdownTimeout)`、`DrainDeadline=max(now, TotalDeadline-CloseReserve)`；无 ctx 截止仍使用有限配置。只有 exchange/Arrival 已结清且本地可发送队列为空才算提前排空；不能因队列暂空而忽略尚在等待的应答/worker。到 DrainDeadline 进入 Closing，停止所有未开始应用写，准备 DPR。已经开始的帧只允许结清，期限取原写期限和 TotalDeadline 的较早值；不能在半帧中插入 DPR，无法结清就直接关闭。

| 事件/竞争 | 确定结果 |
| --- | --- |
| Shutdown 与新 Submit/入站接纳 | 接纳门先赢者属于原工作；Shutdown 先赢则拒绝新工作。新逻辑入站拒绝只在 Draining 的紧急预算内进行；Closing 不再排新拒绝帧 |
| Shutdown 与已经排队的 failover/Retransmit | Node 关闭或对应 Peer Draining 使未 BeginWrite 的该 attempt 终止；不能伪装成 transport failure 再绕到其他 Peer。首次 attempt 才有排空例外 |
| 收到 DPR | 立即进入 Closing；在收到时固定 `TotalDeadline=now+CloseReserve`，已有本地截止则取早，不能等 DPA 写完才起算。撤销排队应用写/新探测，使用控制槽及时回复 DPA；不等待 worker。已开始帧在控制/关闭上限内结清或关闭；DPA 无法安全发送则直接关闭 |
| 本地 DPR 已写完 | 在 `min(TotalDeadline, DPR写完时刻+DPAWait)` 内等匹配 DPA/EOF；不发新 DWR，不启动业务写。既有匹配应答可先赢 Complete |
| 收到匹配 DPA（成功或错误结果） | 立即发起 transport.Close，不再等待 ClosingGrace 或 worker。错误结果仅保留诊断，不取消关闭意图；这遵循 RFC 6733 §5.4.2/§5.6 的关闭路径 |
| 对远端 DPR 的 DPA 已写完 | 只等远端 EOF，最多 `min(TotalDeadline, 写完时刻+ClosingGrace)`；ClosingGrace 仅为连接清理上限，不是业务发送窗口。不刷新收到 DPR 时确定的 TotalDeadline |
| 交叉 DPR | 本地 DPR 和对远端 DPR 的 DPA 使用不同控制关联；各最多一次，deadline 不刷新。匹配 DPA/EOF/截止任一先到即关闭；若自己的 DPA 尚未写完，只结清实际证据，不延迟关闭保证对端收到 |
| 未匹配 DPA、迟到 DWA、旧代 timer | 有界丢弃/计数，不改变关闭阶段、不重置期限、不清另一个控制 pending |
| AnswerCandidate 与 Close/Deadline | 已通过完整校验并先赢 Complete 的应答保留；关闭先封住完成门则按已有发送证据终结。仅“已进入 socket/已读一部分”不等于候选已选中 |
| Respond 与连接关闭 | 逻辑提交和 Arrival 写分开。Committed 可为 true 而 PrimaryWrite=NotWritten；已开始写需结清。Peer.Close 不撤销其他 Peer 副本的逻辑工作，Node.Close 才终止全部未提交工作 |
| 重复 Shutdown/Close | 截止只能缩短，不能延长；ctx 提前取消缩短到当前时刻。Close 直接升级为立即 I/O 终止。首次 Shutdown 已永久停止对应管理域重连，取消调用不恢复 Active |

Node.Shutdown 将所有 Peer 与握手候选纳入同一总截止，不能逐 Peer 累加超时。握手中无可用 Open 连接时直接关闭候选，不伪造 DPR；Maintain/Accept 已完成的旧回调不得在关闭后安装连接。Shutdown 到期可返回 DeadlineExceeded，但只有底层 I/O 结清才释放借用对象；Wait 继续提供库 goroutine 已退出的屏障，CompletionPort 消费不在 Wait 内。

Timeouts 显式展开 ShutdownTimeout、CloseReserve、DPAWait、ClosingGrace，要求均为正且 CloseReserve≤ShutdownTimeout；实际 ctx 可使剩余时间更短，只缩短各子截止。Closing 一进入即取消 Tw、ProbeQueued 和未开始的非关闭控制写；旧 watchdog 事件只结清关联，不再安排探测/重连。Close 幂等发起 I/O 终止，Shutdown 等到本地清理完成或调用截止；需要确认所有库任务结束时使用 Wait。仅需停止等待 Shutdown 而不缩短它的调用方式本版不另设；其 ctx 取消就是立即结束排空的意图。

关闭原因分别决定管理与请求后果：普通 EOF/reset/write failure/watchdog 可进入 R19/R33；LocalClose/Shutdown/MaintainCanceled、PeerProtocolViolation、LocalResourceFailure、RemoteDPR 不自动触发重传。证书/身份/固定配置失败抑制自动重连；普通网络失败按 Tc；远端 DPR 原因沿用 §3.4。原因一旦提交不被随后 net.ErrClosed 覆盖，避免把人为关闭误判为可重传失败。

本轮纠正第九轮“ClosingGrace 内还可开始已有 Arrival 应答”的设计。参考 [RFC 6733 §5.4/§5.6](https://www.rfc-editor.org/rfc/rfc6733.html#section-5.4)；freeDiameter [p_psm.c](https://github.com/freeDiameter/freeDiameter/blob/9d4bc47584bfeced9de708fe00f5629ac4689db5/libfdcore/p_psm.c) 另有 CLOSING_GRACE 来处理无序传输竞争，本库不照搬其扩展为 DPA 后业务窗口。DPR/DPA 都不证明远端业务已执行完。

### 3.6 SCTP事件与连接代的固定映射（2026-11-04）

上游ReadEvent/ParseNotification只报告传输事实；适配器按[运行时§3.5.1–3.5.3](diameterkit-runtime-contracts.md)返回固定TransportEvent。核心按ConnectionID/事件序号提交下表动作；内核assoc ID不替代本地连接代，one-to-one中未提供可用assoc ID不能伪装为新代或据0误关联请求。通知不经业务worker或可能合并丢失的WatchPeers传播，日志与公开观察只是提交后的副作用。

| 输入 | 本库确定动作 / 原因边界 |
| --- | --- |
| COMM_UP、可用路径变化 | 更新有限传输事实；仍先验选项/流数、CER/CEA、选举及watchdog。不是Ready/合法Diameter消息，不刷新Tw、不清DWR、不解除CEA后模式屏障；未知state也不猜成UP |
| COMM_LOST、CANT_STR_ASSOC、RESTART、SHUTDOWN_COMP、EOF/reset | 终结对应候选或活动连接代；RESTART必须Close后新建连接、重新准入/握手，不在原fd上更换ConnectionID继续读。保持本代真实发送证据，普通association失效按原TransportFailure/R19资格判断，不保证重传 |
| SCTP_SHUTDOWN_EVENT | 收到后禁止开始任何新写，包括DWR/DPR/DPA，结清已有I/O并Close；按远端传输终止处理。它不是Diameter DPR，不生成DPA、不进入DPR/DPA的ClosingGrace；此前已提交RemoteDPR或本地关闭原因则保持原原因 |
| PARTIAL_DELIVERY_ABORTED | 本库level0策略一律终止该association，即使未观察到部分数据；不尝试猜哪个消息可继续。归传输中止而非业务失败；未知indication只记录原值，不假装ABORTED |
| CloseRequired、坏元信息、长度/EOR矛盾、截断 | 封住transport后续写准入并Close，实际进度保留；Peer原因记录接收边界/能力失败，不凭net.ErrClosed降成普通网络失败，也不无证据指责远端业务。未归入已批准R19触发类别的完整性错误不自动重传 |
| 通知超NotificationMax、通知速率/候选槽/控制额度不足 | LocalResourceFailure；关闭受影响连接或拒绝候选，不自动重传绕过本地上限。与未消费/已消费的上游结果同时记录 |
| 单路径不可达、SEND_FAILED、未知合法通知 | 固定诊断或有限路径事实；SEND_FAILED仅按原attempt关联补证据，不单独触发failover/业务补偿，不覆盖Written或已完成结果。无界payload/新布局不进入Peer状态 |

终止动作只发布一次不可用、撤销未开始写、取消本代timer，再结清正在进行的I/O。尚未握手成功的候选保持INITIAL；已建立代次进入DOWN。不得用传输错误把Peer Lifecycle置为永久Closed，是否自动重连仍由Maintain/抑制原因决定；已关闭Lifecycle永不复活。晚到writer成功只补真实证据，已读完合法答案按原Complete竞争；二者都不重新安装旧代。

Tw独立于socket读deadline，按§3.3推进OKAY/SUSPECT/DOWN；部分帧超时仍可独立关闭连接。默认无序/多流屏障仅控制发送模式：CEA完整写出后的首次Open/OKAY允许有序stream0业务，必要DWR不得受业务门阻塞；REOPEN仍须自己的探测，通知不替代DWA或合法握手后消息。

依据：[RFC6458 §6.1.5](https://www.rfc-editor.org/rfc/rfc6458.html#section-6.1.5)规定SCTP shutdown通知后的发送限制；[RFC3539附录A](https://www.rfc-editor.org/rfc/rfc3539.html#appendix-A)给出watchdog状态迁移。上表的资源拒绝、事件交接及保守关闭是本库实施裁决，不声称RFC规定本库内部接口。实现与运行验收仍Planned。

## 4. R29 / R36：标识接口、作用域与内存预算

### 4.1 最小公开面

第十三批已实现普通请求标识切片，精确公开字段、源忙即失败、固定guard预算及证据见[batch-13](../spec/history/batch-13.json)。下列仍为最小接口设计示意，实际IdentifierConfig另含MaxSessionIDBytes；ResourcePlan沿用IdentifierBytes并新增IdentifierGuardSlots/IdentifierClasses，未扩充wire.Limits。RO-06恢复部分仍待实施。撤回含义不明的统一 IdentifierSource，区分调用方可提供的业务 Session-Id、可选 E2E 候选源及库内部 HBH/运行标识。

```go
type EndToEndSource interface {
    NextEndToEnd() (uint32, error)
}
type SequenceSpace struct {
    Prefix      uint32
    CounterBits uint8
}
type IdentifierConfig struct {
    Space       SequenceSpace
    ReuseGuard  time.Duration
    EndToEnd    EndToEndSource // nil 使用内置顺序源
}
// NodeConfig.Identifiers 的类型为 IdentifierConfig。
// Source 不是 Store；没有 Acquire/Commit/Release、文件或 I/O 契约。
(*Node).NewSessionID() (string, error)
```

零值 Space 由构造器展开为 Prefix=0、CounterBits=32；显式空间要求 2≤N≤32、Prefix<2^(32-N)。线上的 E2E 为 `(Prefix << N) | suffix`；suffix 的最高一位划分应用/控制，两者各有 2^(N-1) 个值，应用为 0、控制为 1。这是本库局部发号约定，不是 Diameter 位字段。配置报告必须显示有效应用位数，不能仍按 32 位宣传容量。

可选源只供应**应用 E2E**完整候选值，必须落在配置的应用半区。控制 CER/DWR/DPR 一律走独立内置源，入站应答回显远端标识不分配本地 E2E。一个原身份重传沿用原 E2E，不调用源。外部源没有 HBH 覆盖权，也不能把已经终结的 RequestID 变回活 exchange。

应用可直接提供自己的唯一 Session-Id，因此不再增加 SessionIDSource 接口。NewSessionID 提供便利默认值；Session 在多个 CCR 中复用是正常协议行为，不能把重复 Session-Id 当作新的 E2E 碰撞。请求类型/编号的会话管理属于应用。

NewSessionID 与 Node 关闭共用接纳门：关闭先赢则返回 NodeClosed，不消耗新号；已接纳生成须结清后再释放格式化 scratch。启动 nonce/BootID 的安全随机初始化失败使 NewNode 失败；不退回墙钟单独充当唯一性来源。格式/长度上界在构造期先验证，不能持续生成超出 MaxSessionIDBytes 的结果。

### 4.2 内置分配与回绕

| 身份 | 生成/保留方案 | 耗尽或失败 |
| --- | --- | --- |
| ConnectionID/RequestID/AttemptID | Node 随机 128 位 BootID + 各自 64 位递增序号；只有本 Node 可解析其活记录 | 随机失败或序号耗尽拒绝，不重置、不导入旧句柄 |
| HBH | 每连接、本地方向控制/应用共享 32 位顺序空间；随机起点，本代最多走一整轮；末尾保留候选 1024 个控制号 | 应用在余量前停接纳并排空轮换；控制余量用尽则直接关闭；不按 session 发号，也不因取消复用 |
| 内置应用/控制 E2E | 两个半区分别随机/时间播种、顺序递增，每轮每值只用一次；各保存轮次、已分配数、活记录计数和最晚退休时刻 | 一轮耗尽后，只有该半区活记录为零且最晚退休+ReuseGuard 已过才开启新轮；否则 IdentifiersUnavailable。控制无法发号导致连接进展不可维持时关闭，不借应用半区 |
| NewSessionID | Origin-Host + 64 位计数的高/低位十进制 + 128 位安全随机启动 nonce；返回 owned string | 计数耗尽、随机失败或格式/长度不满足时拒绝；无永久历史集合 |

ReuseGuard 构造期展开为至少 240 s，部署可增大；这只约束**本次 Node 运行**。一个 E2E 从成功分配到最后可发 attempt 结束始终 Active；exchange 完成且所有 writer 已结清后才 Retired。ExplicitRetransmit 的 AwaitingRetransmit 阶段仍 Active，不能以“当前没有连接”提前释放。控制请求在回答/超时/关闭且 writer 结清后退休。未写出且未完成接纳的候选可以直接放弃，已成功接纳的请求采用同一退休规则，无需靠发送概率决定回收。

默认源用顺序轮次证明本轮不复用，不保留每个已完成 E2E 的四分钟 tombstone；活引用包含在 exchange/控制记录中。轮次等待采用单调时钟且从最后退休起计时，墙钟回拨不缩短等待。分区会减少可用位数；不得在耗尽时重新随机一个起点绕过当前轮次保护。

### 4.3 自定义源的调用与窗口跟踪

每次 admission 先预留 exchange 资源及一个 GuardSlot，再在**不持状态/额度锁**的情况下调用 NextEndToEnd；同一 Node 内的源调用串行、有界、无 I/O。源返回后重新进接纳门核验 Node 状态、值域、active/quarantine 冲突并原子安装。Close 已先发生则拒绝，不注册 pending、不写出；候选是否被源消耗由源自己决定，库不调用回滚。冲突或源失败立即报 IdentifierCollision/IdentifierSourceFailed，不在内部无限重抽候选；额度不足在调用源前返回 IdentifiersUnavailable。

跟踪表按 E2E 精确匹配，容量固定。Active 记录不能淘汰；退休时令 `ReuseAfter=RetiredAt+ReuseGuard`，转入到期索引；只有超过该单调截止才可删除。共享同一条记录的所有 attempts 不重复收费；Completion 被消费不提前释放 quarantine，未消费 Completion 也不延长 ID 窗口。完成取消、旧应答赢、新 attempt 尚在写等情况，都以最后 writer 结清及禁止后续发送之后为退休点。

建议内部采用预分配槽数组、固定负载因子的索引和到期堆/队列，每条槽只出现一次；固定操作界限/表满拒绝，禁止无界 map 增长、每个 ID 一个 goroutine 或按 2^32 建表。到期表是 allocator 的内存元数据，不是持久日志。其确切 Go 字节 bound 须由原型验证，而不是把 `uint32` 的 4 字节当作全部成本。

EndToEndSource 是受信组件，须满足有界执行；ctx 不能强制打断任意 Go 函数。若同一实例被多个 Node 使用，源还须并发安全，并保证它们联合生命周期内不重复候选或自行满足联合窗口；各 Node 的局部表不能互相认证。共享源只供应应用 E2E，不能协调各 Node 的内置控制 E2E；因此共享 Origin 的并行 Node 仍须使用互不重叠的 Space，或由部署另行证明包含控制半区的完整隔离。只有应用候选不重复不足以关闭此项。库不提供全局注册表、后台号段服务或共享数据库。

### 4.4 单账本预算与容量

ResourcePlan 增加 **IdentifierReserve**，仍属于同一 Node 总账本；只包含固定 allocator/索引/scratch，不是恢复后的 Evidence 分区。配置由 builder 展开并可检查：

| 字段/计量 | 确定规则 |
| --- | --- |
| IdentifierReserve.FixedBytes | 内置两个 E2E 轮次状态、BootID、Session 计数/串行格式化 scratch 和固定索引头的保守上界 |
| IdentifierReserve.GuardSlots / GuardBytes | 自定义源或RO-06恢复入口启用；`GuardBytes ≥ TrackerBound(GuardSlots)`，覆盖 active+quarantine、恢复摘要、哈希空槽、到期索引和结构对齐；仅内置源且未启用恢复时为Disabled，不将零解释为无限 |
| ResourceClass.IdentifierSlots | 跟踪槽按 class 硬配额，配额之和≤GuardSlots；active、quarantine和恢复保留均归原 class，不互借。物理表只记一次 Node 字节，避免大 Sh 的完成速率耗光 SGd 的 ID 窗口 |
| 与 Q/A/M 的关系 | ID 跟踪槽本体只记在 IdentifierReserve；exchange 的 M 仅计引用/退休标记。原身份重传不新增 GuardSlot，新 attempt 的 HBH/匹配仍计 M/A |
| 控制隔离 | 自定义源 GuardSlots 用满只阻止新应用 exchange；控制 E2E 不用它。HBH/控制 active 记录计入各连接 TransportReserve，既有重传/收答不要求新应用 ID |
| NewSessionID | 有限 NodeConfig.Limits.MaxSessionIDBytes 下测量并序列化生成；一次固定 scratch，结果返回后归调用者；不把每个已生成 Session 永久保存 |

Node 构造检查 `TransportReserve + IdentifierReserve + ΣClassPartitions ≤ NodeLimit`。固定预留不会因为一个请求完成就归还 Node；GuardSlot 到期只恢复 allocator/class 可用槽，Snapshot 只报告 Active/Quarantine/Free 数量和容量，不枚举 ID。Node.Close 后等待所有 writer、仍在进行的源调用/发号及接纳检查结清再释放 allocator；既有 CompletionPort 的结果另有自己的 charge，不需要保留已停止 Node 的跟踪表。外部自定义源自行拥有的内存须在部署容量报告单列，不能声称由 Node 预算管住任意用户代码。

自定义源最低容量按 `最大同时活 exchange + Guard 时间内峰值完成数` 估算，不能只用 pending 上限。例如 1000 新请求/s、最长活跃 30 s、Guard=240 s，稳态约需 270000 槽，另加突发余量；若测得 TrackerBound 平摊 128 字节/槽，仅跟踪就约 33 MiB，128 MiB Node 候选总额必须减去这部分。128 是算例，不是尚未实现的 Go 对象承诺。默认顺序源不承担该逐完成记录成本，但会在一轮空间耗尽时保守等待。

### 4.5 重启、多 Node 和 Origin-State-Id

默认随机 nonce/播种提供概率隔离，不宣称证明了 RFC 的跨重启 E2E/Session 唯一性。每个部署须选择真实独立 Origin、互不重叠号域/启动世代或自身发号规则，并说明快速重启限制。共享 Origin 的多个 Node 不能各用同一默认 Space 就认定无碰撞；旧 Node 停止后局部保护不自动延续到新 Node。静态掩码也不能独自解决同号域快速重启。

Origin-State-Id 默认省略或调用方显式配置；Profile 必需时缺失报错，库不维护持久重启计数。它描述Origin-Host实体，CEA及应用答案使用本地值或省略，不回显请求方状态；显式0与缺省保留Presence差异，但均不作旧会话失效推断。非零重启单调保证由部署负责，不能把启动秒/随机值当证明。远端值变化只形成既有有界证据/事件，不跨Origin比较，不单凭此取消其他pending或删除业务会话/账本，也不新增远端历史表。RFC要求与部署保证分别记入符合性矩阵；输入及R17比较关系见主流程§26.1，无存储仍可使用全部上述API。

参考 [OTP sequence 配置](https://www.erlang.org/doc/apps/diameter/diameter.html) 的有限位域/速率约束和 freeDiameter [sessions.c](https://github.com/freeDiameter/freeDiameter/blob/9d4bc47584bfeced9de708fe00f5629ac4689db5/libfdproto/sessions.c) 的受锁高低位计数。本项目选择控制/应用半区和有界 quarantine，不声称参考库提供相同保证；普通源接口/轮次与guard预算已在第十三批实现；生产压力和跨部署唯一性仍待验证。

### 4.6 不随边界修订取消的规范工作

固定 Profile 仍需锁定修订/依赖、字段/条件/错误分支及 fixture。Time 保留原始 32 位，转换需显式 era；Supported-Features 缺失为 Unknown，最终节点能力不能从 DRA 广告推导。移除持久化不等于放松编解码、协议状态机或来源校验。

### 4.7 RO-06原身份接纳的例外与资源（2026-11-04，首版内P5）

> 最新实施（第五十六批）：重启显式导入、默认本轮普通已用环形区间、自定义active/quarantine和独立RecoverySpaces已接通；导入暂占guard失败回滚、准备/收发期间固定保留、已有本地终态不能Import复活，见[batch56](../spec/history/batch-56.json)。最多16个恢复域与整个普通应用/控制Space不交叠；默认32位全域需先显式缩小才能另划RecoverySpace。Checkpoints和三项声明不替代应用历史进度/隔离/时间保证。

> 历史实施（第五十二批）：同Node prepared guard复用及原E2E段接纳、SegmentNo/额度高水位、同段幂等/完成防复活已实现；正在准备/活动的guard不被到期扫描清除，最后borrower结清后再保守延长ReuseGuard。重启导入、本轮普通已用区间检查和RecoverySpace仍待实施。[batch52](../spec/history/batch-52.json)限定原目标、本地无答续段；下方旧实施注按历史时点阅读。

> 第五十一批已实现Prepare的有界本地摘要guard、默认/自定义E2E冲突检查、配额及关闭归还；默认源碰撞只消耗一个候选并返回，不重抽。下述导入号域、同RecordID段接纳/高水位和RecoverySpace仍待实施；[batch51](../spec/history/batch-51.json)不宣称这些已通过。

[运行时§1.7](diameterkit-runtime-contracts.md)允许保存的RoRecoveryRequest在重启后保留线上E2E；这不是EndToEndSource反复返回同一候选号，也不是恢复旧RequestID。Prepare从本Node正常应用号域发号；SubmitRoRecovery导入先校验版本、不可变内容、进度、Origin授权和配置允许的恢复号域，再占GuardSlot。普通新请求与控制请求不开放原始头覆盖。

Guard以本地Origin/E2E为冲突键，保存RecordID/摘要、Prepared/Active/BetweenSegments/Final状态、SegmentNo和已用次数高水位。重复同一活动段只引用原exchange；完成后旧段号拒绝，合法后继只在旧writer/pending/candidate全部结清后占同一槽。另一内容/RecordID或当前仍活的普通请求冲突一律拒绝；不能仅凭Session相同豁免。已成功/永久终结的本地记录不能通过导入旧进度重开。进度表有期限、有配额，不是完成历史查询服务。

默认顺序源也须在发号接纳门检查恢复保留槽；候选号碰撞立即返回IdentifiersUnavailable，不无界扫描/重抽。导入值若已落在本轮普通分配的已用区间、控制域或其他active/quarantine中，拒绝；本轮已用区间可由轮次起点/计数确定，不增加所有普通完成请求的逐条历史。只有已有同RecordID/内容的恢复槽允许原身份续发。新Node没有旧进度，应用须在开始新的普通流量前导入待恢复记录并协调新号域；库不能查出尚未导入的磁盘记录或另一进程使用的号。

旧C++队列记录若E2E不在本Node应用半区，不能重写E2E假装恢复。迁移方须配置专门的有限RecoverySpace及承接原Origin的Node，RecoverySpace与该Node普通应用/控制Space完全不交叠；旧新发送者先隔离，域外记录明确拒绝。不能默认接受任意32位值覆盖控制空间。旧文件解码、分批恢复/排空和原始身份完整性核验均在应用迁移工具完成，库只接收同一受检记录格式。

恢复槽从Prepare/导入到max(ReplayUntil,最后writer退休时刻)+ReuseGuard保留，包含段间等待和终态；计入IdentifierReserve.GuardBytes及class IdentifierSlots。新段复用同槽，另预留当前Q/A/M/Completion；默认源未启用RO-06时仍沿§4.2常数轮次模型。恢复槽容量按“准备速率×最长恢复年龄 + Guard内退休数 + 突发”估计，不能套普通30秒pending算例；表满明确拒绝，不能提前清掉成功标记来腾槽。

恢复记录不是远端去重保证。跨重启时间、进度防回滚、Origin唯一性、旧进程网络隔离和OCS去重保留期由应用/部署承担；禁止把内存guard解释成全局ownership或持久幂等。验收沿REPLAY-01–04/RO-06，不新增标识服务或数据库。

## 5. 参考与取舍

[Erlang/OTP diameter_app](https://www.erlang.org/doc/apps/diameter/diameter_app.html) 的请求/重传及 Peer 回调边界、[diameter](https://www.erlang.org/doc/apps/diameter/diameter.html) 的 watchdog/sequence 配置用于协议设计参照。freeDiameter 的 [p_psm.c](https://github.com/freeDiameter/freeDiameter/blob/9d4bc47584bfeced9de708fe00f5629ac4689db5/libfdcore/p_psm.c)、[p_dw.c](https://github.com/freeDiameter/freeDiameter/blob/9d4bc47584bfeced9de708fe00f5629ac4689db5/libfdcore/p_dw.c) 与 [sessions.c](https://github.com/freeDiameter/freeDiameter/blob/9d4bc47584bfeced9de708fe00f5629ac4689db5/libfdproto/sessions.c) 提供状态机/计数实现参考；不从这些局部实现推导应用幂等或持久保证。

第七轮新增存储框架超出了公开 Diameter 库的定位；本轮按用户指示删除，而不是换名为可选 persistence 插件继续交付。库保留必要的协议状态，不追求无状态网络端点，也不管理业务持久状态。

## 6. 当前验收与撤销项

四应用首版均使用R32/Peer、R36标识及通用上下文/内存边界。计费Tx、CCR/CCA、R33/RO-06用例也须在首版执行；公共超时/关闭/发号/预算复用验证，并补Ro影响部分。

| ID | 当前验收边界 |
| --- | --- |
| FACT-01/02 | 改为逐规则具名上下文、原请求快照/类型/空间、Ready 不代表业务授权；原通用实时 FactSet/时钟框架验收撤回，不记通过 |
| PEER-01/02 | 连接选举、CEA 屏障、watchdog 全表、DPR 原因、代次和关闭/重连竞态 |
| PEER-03/04 | 分用途 BeginWrite、DPA 即关、交叉关闭/候选安装/旧 timer、总截止及引用结清；详细用例见覆盖基线 §11/§12 |
| ID-01/02 | 默认内存发号、控制余量、进程内不冲突/回绕、随机源失败、可选源与部署责任；无 provider/Seal/持久恢复测试 |
| ID-03/04 | R36 自定义源的窗口记录/预算/拒绝和重传后释放；接口/保留规则已确定，容量与实现仍待验证 |
| STORE-01/02、REV-07 | 按用户明确边界撤销库验收，不记为测试通过；应用自行承担存储一致性 |
| SCOPE-01 | 不配置存储、不创建数据文件即可运行；无 charging/Store/provider 公开面；停止 Node 不等待外部业务事务 |
| SCOPE-02 | 业务 Tx 与 exchange Deadline 分开；一次 Completion、晚到/超期丢弃规则明确；Node 重启不恢复 pending 或去重缓存 |

以上为文档设计，未执行实现测试。R24/Peer/标识的进程内契约继续待验证；R29 逐 Profile 审计、R03 真实部署与 R18 SCTP 能力仍需闭环。应用数据库、跨进程接管和业务恢复不再是这个公开库的发布阻塞项。

第十八轮补充：[主流程§14](diameterkit-mainflow-audit.md)的TERMINATION沿用业务Session但建立新exchange；Number1仅适用于候选中没有UPDATE的INITIAL→TERMINATION序列，不是Node的固定发号策略。独立IEC/refund及RDR/ALR分别使用各自请求标识。完成后不为这些场景新增历史存储或跨进程恢复；原请求关联仍指当前pending快照，已完成业务阶段的输入由调用方提供。


第二十轮[主流程§16](diameterkit-mainflow-audit.md)不新增进程状态种类：Sh SNR与后续PNR分别分配exchange标识，不能以SNR的HBH/Session关联通知完成；XML的业务值不替换重复请求raw内容。ALR与后续可能由应用发出的SRR/TFR也各自发号。Expiry、用户注册和服务节点映射不进入库状态存储；已有pending、去重、Peer关闭及标识预算契约继续适用。


第二十一轮来源数据的“permanent/temporary subscriber data”属于23.008描述的网络实体，不成为库持久化职责。Sh退订与到达PNR竞争只服从各自exchange/Inbound一次完成，不维护持久订阅状态；IMS HSS-ID、dummy MSISDN与Router token不使用HBH/E2E发号器。状态种类/预算不增加，详见[主流程§17](diameterkit-mainflow-audit.md)。


第二十二轮的[Sh混合语义](diameterkit-mainflow-audit.md)只描述协议字段，不建立订阅状态机：25的一次通知删除归HSS，持续11/12不由库消费；PNR早于SNA或退订后的到达不触发历史查表。应用显式拆分的两个SNR各自新建exchange与标识；PNR不复用SNR标识完成它。兼容Profile指纹隔离沿用已有去重契约，SRA返回的节点/用户身份不作为发号输入，也不持久化。


第二十三轮[文本边界](diameterkit-field-inventory.md)不改变Canonical或Peer身份用途：前者不做NFC/NFKC/URI等价化/XML重编码，后者保留既有ASCII身份比较规则。Sh XML SequenceNumber与Diameter716仅为各自消息字段，不使用HBH/E2E/Session发号器，不保存跨请求版本。诊断Presence和结果命名空间属于owned消息，不引入业务状态或持久化。

第二十五轮[Sh特性与T4规则](diameterkit-mainflow-audit.md)不增加状态所有者：本次终端特性/用途来自既有具名输入与原请求；库不自动缓存能力、查询外部身份或保存T4计费状态。5011/Base5001处理不改变Peer状态、pending完成点或发号契约，不隐式另发降级请求。全部逐叶候选仍是文档证据。

第二十八轮[九答案族关联](diameterkit-mainflow-audit.md)澄清候选查找与Complete不同：ConnectionID/HBH候选还需App/Command/E2E匹配，之后才校验Session和来源；未知/错头不消费别的pending，已匹配无效答案按既有ErrInvalidAnswer完成。已完成/取消/超期和无映射的新连接同字节输入有界丢弃；旧代已合法完整读入的候选仍携原AttemptID按R19门竞争。R=1的PNR与SNR及其答案独立，CorrelationID/Session不能代替线路关联域，不新增持久历史或全局索引。

第二十九轮[公共边界](diameterkit-mainflow-audit.md)不改变完成门：已关联答案未知M1/外层坏包按原ErrInvalidAnswer；Failed-AVP仅内部引述坏长度不据此重新等待或重试；ResourceLimit依原资源失败路径。诊断父路径和SF键不成为新的pending索引。

第三十一轮[目的/去重前置校验§27.3](diameterkit-mainflow-audit.md)保持每次Arrival独立准入；明确授权的本地别名不自动成为Canonical内容等价，Destination仍保留。Session接收缺失不得进入发号器补值；非法回显快照的终止沿用既有发送门和关闭契约，不增加业务会话或持久状态。

## 第64批Peer接线

COMM_UP/路径/SEND_FAILED不刷新Tw；终止事件封住transport新写后进既有完成/关闭门，旧代事件不复活。SEND_FAILED Context0仅诊断。CEA/DWR控制有序stream0；响应方固定Handshake时限屏障在合法握手后消息后解除，REOPEN原规则不变。NotificationLimit归StopBackpressure，关闭但不自动重传。具体范围/证据见[batch64](../spec/history/batch-64.json)，历史未实施表述按对应时点阅读。

## 第65批托管建立接线

SCTP真实Maintain断链复用既有退避和REOPEN，重连生成新ConnectionID；三次DWA前不Ready，不继承旧pending，不自动重放SRR。原请求按WrittenNoAnswer完成一次，旧Inbound晚答不得进入新代。Serve取消仅撤销listener和未交接候选；建立完成的Peer保持可用。双向Dial/Serve选举及双端Shutdown经过真实SCTP验证，候选在借用退出后才归零。见[batch65](../spec/history/batch-65.json)。

## 第66批终止通知映射

上游RESTART现在可同时返回raw通知、ErrAssociationChanged和CloseRequired；部分投递中止也可带CloseRequired。本库在合法结果下先保留typed终止事件，坏/未消费结果不可续调；writer先观察restart时保留真实字节并分类为association终止。首因、ConnectionID、一次完成和Wait退休保持，见[batch66](../spec/history/batch-66.json)。
