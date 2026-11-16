# Diameter 部署绑定、重复请求与 SCTP 契约

> 第68批更新：Relay-only协商和有限逐请求目标表已实现并验收，见[batch68](../spec/history/batch-68.json)。后文第67批扫描缺口为当时状态；目标表之外的动态发现和其他Ro/P4/发布缺口未由本批关闭。

> 第67批当前状态：独立adapter已锁定正式SCTP v0.8.0（bce1b3cf3ea3），源码与第66批候选一致，A–E、资源和Peer边界不变；正式版本回归完成，见[batch67](../spec/history/batch-67.json)。上游发布门已关闭，核心/adapter公开发布及现场验收仍单列；下文旧候选/未合并表述按历史时点阅读。

> **当前范围（2026-11-04，Draft 0.20）：** 第一阶段=首版/v1，包含Sh、S6c、SGd/Gdd、Ro/DCCA及sms-fw实际所需六模板、SCAP1和RO-06恢复能力；Go 1.26、纯Go实现（含依赖与SCTP适配器，发布路径CGO_ENABLED=0）、公开Go库、无持久化边界不变。明确撤销第十九轮“Ro后置”分期；带日期的旧分期/统计仅作历史证据，不再授予Ro跳过首版验收的例外。限定S6a、完整SCUR/PRICE_ENQUIRY等未承诺能力仍后置。范围及剩余缺口见[Specification §3/§7](diameterkit-specification.md)，顺序见[总设计§14](diameterkit-design.md)，验收见[覆盖§39](diameterkit-usage-coverage.md)。2026-11-06已完成前十九批有界实施（Base/S6c基础、公共包络与关联、出站运行时、关闭与选举、TCP/TLS生产适配、有限Inbound/应答、有限Reject/应急预算、Canonical/有限重复缓存及多Arrival重答、固定同终端Route/多attempt与活动请求原身份重传、有限Dialer/Maintain拨号及重连、按class/编码字节DRR调度、固定Peer的有限Listener/Serve、有界标识源/Session/guard、Base/S6c有限Context/Report、S6c普通RDR/RDA及HSS ALR/ALA、SGd/Gdd普通OFR/OFA与TFR/TFA、Sh普通四组命令及透明User-Data、Ref24/25有限Sh-Data helper、Sh单订阅/退订及Ref11/12有限数据）；已执行范围及未完成项以[Specification §6](diameterkit-specification.md#6-需求与证据的维护)和[spec进度](../spec/progress.json)为准，未列明的运行验收仍Planned。

日期：2026-10-31（第二十一轮补号码/订阅与特殊身份切片）。第五轮建立 R03、R17、R18；第六轮修订指纹/策略边界、错误路径及应答生命周期，第八轮收回持久化职责、第九轮复核公开库边界、第十轮收敛 R32/R33/R36、第十一轮补 [Profile 规则与语义比较](diameterkit-profile-rules.md)、第十二轮补 [主流程与模板差异](diameterkit-mainflow-audit.md)，[协议上下文与 Peer 契约](diameterkit-state-consistency.md) 只保留进程内状态和可选调用方发号策略；与 [运行时契约](diameterkit-runtime-contracts.md) 的 R19–R21 配套。本文规定拟实现行为，不是实现或互操作通过记录。仅面向 Go 1.26 库，不增加 C/C++ 调用层；本轮不修改底层 SCTP 仓库。

第十八轮更新的库级需求与验收索引见 [Library Specification](diameterkit-specification.md)；本文保留状态、时序和预算细节，不因规范草案建立而视为实现通过。

## 1. 参考实现与取舍

参考实现用于发现成熟做法和反例；规范要求、其他库现状、diameterkit 的选择分开记录。以下是第五轮的源码观察记录；第九轮只重查评审中列出的直接来源，未重审此表所有分支。分支链接可能变化；实施基准须另锁 commit，不能将这些项目或业务 fork 的版本混为一谈。

| 来源 | 本轮观察 | 对本设计的影响 |
| --- | --- | --- |
| [freeDiameter routing_dispatch.c](https://github.com/freeDiameter/freeDiameter/blob/master/libfdcore/routing_dispatch.c) | 根据 Destination 与候选邻接进行匹配/评分，区分本地处理与转发 | 邻接和终端目的地分离；本库只选已配置端点路由，不复制完整代理路由引擎 |
| [freeDiameter sctp.c](https://github.com/freeDiameter/freeDiameter/blob/master/libfdcore/sctp.c) | 接收检查 EOR，读取 SCTP_STATUS；明确设置 fragment interleave，未启用 level 2 的并行分段处理 | 采用消息边界 API 和明确的内核选项；不能只包装 io.Reader；本库额外约束预算、截断及取消 |
| [go-diameter network_sctp.go](https://github.com/fiorix/go-diameter/blob/main/diam/network_sctp.go) | 提供按 stream 的缓冲/读取和发送接口；使用 PPID 46；Write 可沿用读取 stream | 保留 stream 元信息，但不把对端入流直接用作本地出流；不继承依赖隐式当前 stream 的可变连接状态 |
| [jDiameter PeerImpl.java](https://github.com/RestComm/jdiameter/blob/master/core/jdiameter/impl/src/main/java/org/jdiameter/client/impl/controller/PeerImpl.java) | 可选重复保护路径在发送应答时，按原请求路由信息和 End-to-End 保存结果；排除 CEx/DWx/DPx | 重复表跨单次读写存在，控制消息另走状态机；这一局部实现不能证明内容冲突校验、正在执行的去重或跨代理 envelope 正确 |

本库新增的 CanonicalV1、逻辑入站结果提交和硬预算是本项目的设计决定，不声称上述库已实现。后续 R19–R21、Profile 和性能分析也沿用“规范 → 上游源码 → 本地业务证据 → 选择/反例”的方法；不从 README 的支持声明推导一致性或性能结论。

## 2. R03：DRA 部署绑定

### 2.1 绑定对象及构造检查

固定链为 `Node → PeerPlan → (App-ID, Command, R/E) → Binding → CommandPolicy`；Route 只能引用这个已存在的解释。最终 Origin、Destination、Session 和 pending 都不能更换 decoder。通过 DRA 的邻接身份是 DRA，HSS/MME/OCS 是业务来源或目标；Relay 广告只证明邻接宣告的转发能力，不证明后端版本、特性或去重域。[RFC 6733 §2.4、§5.3](https://www.rfc-editor.org/rfc/rfc6733.html#section-2.4)

| 对象/拟定字段 | 确定契约 |
| --- | --- |
| Node.Identity / Domain | 一个真实本地 Origin 身份和网络管理上下文；同身份多实例由部署负责 Session/End-to-End 唯一性与 Origin-State-Id；库不提供或要求持久 provider；Domain 字符串不提供隔离证明 |
| PeerPlan.ExpectedPeer / AdmissionPolicy | ExpectedPeer 检查 CER/CEA 邻接身份；AdmissionPolicy 固定 listener/网络上下文、TLS 身份或明确受保护的网络来源；两者均通过才使用预构建 Bundle |
| ApplicationBinding.CommandPolicies | 每个命令方向固定角色、可推导用途/所需具名协议输入、ClassID、FailurePolicy，以及对应的 InboundPolicy / OutboundPolicy；基础控制命令另列，不以 App-ID 0 一概判断 |
| CommandPolicy.InboundPolicy | 固定 SourcePolicy、LocalDestinationPolicy、DuplicatePolicy；同 Node 同命令在允许跨 DRA 去重的入口使用一致的语义指纹 |
| CommandPolicy.OutboundPolicy | 所有 Peer/Route 请求入口的目的地、应答来源、具名协议输入、class 基线；Route 的对应策略只能收紧，不能扩大授权域或另换 class |
| RouteConfig.DestinationPolicy | 固定允许的 host/realm 配对、是否允许 realm-only，以及动态目标的受限范围；Route 默认值和每次 override 都必须满足 |
| RouteConfig.AnswerSourcePolicy | 分列 TerminalOrigins 与 ErrorAgentOrigins；后者只授权对应 E 位/错误分支，不能以 DRA 身份返回终端成功结果 |
| WireFingerprint | 固定类型/路径、命令语法、校验规则版本和扩展描述符；用于确认同一消息的解释，不含网络地址、来源许可及资源 class |
| InboundSemanticsFingerprint | WireFingerprint 加角色/UseCase、校验和失败策略、CanonicalV1/排除表、DuplicatePolicy 与 ClassID；用于同键重复结果的语义一致性 |
| BindingFingerprint / PolicySnapshot | 全绑定的诊断/配置版本摘要；不是 Route 候选必须整体相等的条件。比较不能只看 Profile 名称，实际比较固定描述符/策略，不以短摘要证明相容 |

第十二轮补充：Ro 字段选择只影响本地生成时，不改变 WireFingerprint；若部署确需约束对端OM/OC出现条件，该具名入站策略进入InboundSemanticsFingerprint；最短Service-Context-Id下不得将标准可选项误报必需。任何额外部署限制必须与标准校验区别，不能冒充通用3GPP约束。第十四轮修正规范接受条件须更新规则版本/语义指纹，区别于仅修改本地构造选择。标准 Recipient-Info 父路径、Address family、私有退款 Vendor/Code 和错误配对都是语义差异，不能由 Canonical 移动字段、猜类型或择一结果“修复”。SMS 的 CC-Session-Failover 不使用与 R33 同 OCS、同活动 exchange 的传输恢复是不同能力，不能据前者开放跨 OCS 切换。

SourcePolicy 选一种显式模式：`ExactOrigins`（host/realm 配对集合），或 `TrustedRelayRealm`（指定可信 Peer 集合和最终 Origin-Realm 集合，接受这些 realm 内的 host）。不允许默认通配所有 realm；后者依赖 DRA 的来源治理，库无法从 Origin AVP 证明端到端身份。双方策略始终检查合法 ASCII DiameterIdentity，比较采用 ASCII 大小写折叠，不做 DNS 查询、后缀猜测或任意别名归并。Session-Id、用户标识和 opaque 内容不采用该折叠。第十四轮明确：SIP身份的显式等价比较不等于CanonicalV1请求内容比较；URI等价但原字节不同仍可能保守报冲突，不做URI/Unicode正规化来合并业务请求，见主流程§10.2。

LocalDestinationPolicy 固定本地 host/realm 及明确别名；缺 Destination-Host 仅在命令/过程允许 realm-only、策略明确授权且 realm 属于本地服务域时接受；HSS发起ALR的过程必需Host不能被CCF可选性覆盖。面向其他 host 的请求不转发，走 UnableToDeliver。出站 SRA 提供的 MME 目标只是业务输入，需通过 DestinationPolicy，不能因来自一个合法 SRA 就任意建立连接或绕过路由范围。

构造期按所有候选 Peer 展开每个 Route，逐项检查 ApplicationRef、命令/角色、目的地、广告、资源 class、失败映射和去重策略。WireFingerprint 相同且分别满足角色/策略约束的候选可共享 Route；来源许可、地址不同不要求制造新的协议 Profile。WireFingerprint 不同不做运行时“相容猜测”：先离线审计字段/规则差异，构建一个具名、固定的组合 Profile，再让路由共同引用。可选字段扩展仍须验证未知 M 位、Grouped 路径及条件规则，不能简单取字段并集或放宽所有必填。不可合并则 BindingConflict，输出冲突的 Peer/App/Command/路径及规则 ID。

不相容 Profile 必须使用对端认可的真实独立部署（不同 Diameter 实例身份，或已经验证的独立进程/网络接入上下文）。v1 常规 Node 不接受“同身份再建一个 PeerID”作为绕过方式。特殊平行连接仍属具名兼容部署，未取得对端配置和主动入站证据前不能启用；IP 多地址通常属于一个 DialPlan/association。基础协议的单邻接连接及多实例例外见 [RFC 6733 §2.1](https://www.rfc-editor.org/rfc/rfc6733.html#section-2.1)。

2026-11-04私有字段收敛：旧退款420 Grouped的确切位置为`CCR/0/456/0/437/0/420`。具名`smsfw-ro-legacy420-v1`在此处唯一解释为Grouped，明确不支持同位置CC-Time，保留413/417，可承接已列五个CCR模板在类型层的组合；其他父路径/CCA不自动获得旧420别名。它不是全标准Ro，也不能按Requested-Action、payload长度或Origin/Destination选择decoder。若目标必须同时支持冲突位置的标准CC-Time，则采用真正可鉴别的独立解释域/连接和Peer准入；同邻接身份多连接仍遵守选举，必要时独立Node/真实端点，并满足共享Origin时标识唯一性义务。单改PeerID/Route名不构成隔离；DRA只有一个不可区分binding时返回BindingConflict。

私有定义在启动前完整固定，各层Vendor、flags、路径、类型/cardinality及版本进入WireFingerprint。PSP三层接收Vendor、token接收定义未证时不启用相应typed能力；不能继承父Vendor或按配置只改接收Code就重写发送1292。Route所有候选和RO-06原记录须保持同一解释，不能在备用目标把420改413或把1292换2022。定义表和准确证据出口见[主流程§34](diameterkit-mainflow-audit.md)，验收见[覆盖§40](diameterkit-usage-coverage.md)。三组私有材料合并进§31.5六模板已有包；样例槽映射和某次被接受不能替代完整请求/成功失败答案、独立解读及固定规则依据。库不保存验收档案，不查询凭据或校验业务账本。

### 2.2 校验顺序与错误归属

以下接收顺序适用于请求；答案先按主流程§24.3核对ConnectionID/HBH及完整关联头，再使用已保存binding与A预算，不由收到的App/Origin另选解释器。请求接收顺序固定：传输/邻接 admission → 有界头分类 → 已知命令的固定 binding Wire → Source/本地 Destination → 去重查找与内容比较 → Application Check → 副作用 → 结果提交。App/Command 未知或已知请求的 Wire 失败，走下述独立拒绝路径，不要求先成功取得应用 codec。每个重复副本都重新做来源/目的地授权，不能因命中缓存绕过授权；reader 不查询数据库。

| 失败 | 本地结果 / 线路行为 |
| --- | --- |
| 重复 Peer、不相容解释、策略缺失或歧义 | 构造失败 DuplicatePeer / BindingConflict / InvalidPolicy；不等第一条业务消息才决定 |
| CER 身份或传输 admission 不符 | 不进入 Ready；可合法回答 CER 时使用对应握手失败路径，随后关闭，不交付业务 |
| 未支持 App / Command | 分别使用 3007 APPLICATION_UNSUPPORTED / 3001 COMMAND_UNSUPPORTED 的基础错误分支 |
| 本地目的地不匹配 | 3002 UNABLE_TO_DELIVER；本库不尝试充当代理 |
| 入站来源不在 SourcePolicy | 本地 SourceDenied；固定策略映射 3002，且诊断不泄露缓存/订阅信息；这是本项目的路由准入拒绝映射，不冒充用户授权结果 |
| 出站目的地越界 | Submit/Request admission 失败 DestinationNotAllowed，无发送、无 Completion |
| 匹配应答的终端或错误代理来源不符 | ErrInvalidAnswer，保留有界诊断和已发送证据；不重试、不向 answer 再发错误 |

未知 App/Command 使用固定基础 ErrorAnswer builder；它不依赖未支持的应用 Profile。对 framing 仍可信的失败请求，只在紧急预算内提取必需关联、Session、Proxy-Info 和允许的 Failed-AVP 诊断；失败的业务字段不再按完整成功消息语法校验后才允许拒绝。无可信头或必需 envelope 无法安全保留时关闭，不猜测应答身份。已知命令的普通业务拒绝仍使用对应 Profile 的分支。

所有应答明确清除 R/T，保留请求 P；E 位错误用 RFC 6733 §7.2 的基础语法，要求 Result-Code，不能用“Result-Code/Experimental-Result 二选一”统一替代；可选Experimental-Result的语义范围见主流程§24.2，RFC6733允许基础5xxx采用E语法；普通应用结果规则按固定 Profile。Failed-AVP 内诊断内容不能再次按普通业务字段完整性拒绝。未知非 M 位 Grouped 作为 opaque 保留，不猜测其中的 M 位；已知 Grouped 依描述符与具体规范规则处理。[RFC 6733 §3、§4.4、§7.2](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.2)

错误受相应基础/应用 Error grammar、完整 envelope 和 R21 拒绝预算约束；不能合法编码或拒绝槽耗尽则关闭。TerminalOrigins 与 ErrorAgentOrigins 均要显式配置；E 位错误不强求成功应答的终端字段，但仍检查关联、Origin 和该错误分支语法。未知后端来源集合不能通过无条件信任 DRA 的成功应答补齐。

### 2.3 部署交付表与未取得的证据

部署材料按下表一行一个“本地身份 × 邻接集合 × App/命令方向 × 终端域”，保存配置快照与报文 fixture 引用。下列 `.example` 只是模型例子，Profile 名不是已选定的 3GPP 修订。

| 用途 | 本地/邻接 | 固定解释和终端约束 | 当前状态 |
| --- | --- | --- | --- |
| sms-fw 查询 HSS | sms.example → dra-a.example / dra-b.example | S6c 固定 Profile；HSS realm-only 查询；成功来源限定 HSS 域 | 拓扑方向已知，真实身份、广告、来源和版本待取证 |
| sms-fw 与 MME 双向消息 | 同上 | SGd/Gdd 固定 Profile；目标来自已校验 SRA 且在允许域；入站 OFR 等按命令角色约束 | 需要双向样本与所有主动入站命令清单 |
| ip-sm-gw 的 HSS 交互 | isg.example → DRA 集合 | Sh/S6c；包括无 pending 的 PNR/ALR；不能按请求曾经走过的 Route 选择解释 | DRA 数目、路由及 Profile 差异待取证 |
| Ro / SCAP1（首版） | 本地身份与 OCS 邻接路径未知 | 直连则 OCS 是 Peer；经 DRA 则绑定该 DRA 的 App 4/具名兼容命令域；恢复策略另核对 | 不推定直连，也不自动加入现有 DRA 的恢复池 |
| 未来 ip-sm-gw ↔ sms-fw | 真实双方身份待定 | 复用 Node/Peer/Route；先确定具体协议、命令方向与角色 | 不预设一定使用 SGd 或复用 HSS 策略 |

每行必须附：本地/远端身份与地址、Dial/Listen 上下文、保护方式、CER/CEA 广告、Profile/策略指纹、请求/成功/错误样本、允许 Origin 与 Destination、代理改写清单、Class 与容量、MaxProcessing/ReplayRetention、恢复域证据，以及调用方标识策略的作用域/多实例隔离/ReuseGuard、Peer timer 和重连抑制配置。TransportFailover/ExplicitRetransmit 额外要求各候选通向同终端或共享业务去重域，且可证明的保留期覆盖 MaxReplayAge；“同一 realm”或“双 DRA 均就绪”不能替代证明。

配置结构合法与部署验证通过分开：构造器能查确定性和冲突，不能认证人工填写的终端关系。部署验收记录分别标记 `Unknown / Declared / Verified`，附采集时间、配置版本及双向测试；不是运行时自动签发的安全证书。DRA 将重传分到不同业务进程时，Node 内存去重不跨进程；由应用选择业务幂等或经证明的分流约束，不能向库增加共享数据库要求。进程重启和热升级也不继承内存重复表。

第九轮 R35：上述完整交付表用于特定部署的互通/投产审查。公开 NewNode 只校验显式配置的一致性、范围和资源，不要求抓包文件、人工 Verified 状态或现场审批才能构造。普通 SingleAttempt 不要求填写 ReplayDomain 证明；启用 TransportFailover/ExplicitRetransmit 才需其域/期限约束。终端可达性和共享去重的真实性由部署负责，库不能认证人工声明。真实现场材料未齐不等于 codec 或独立对端测试不能开展。

第十轮的普通Ro Route模式仍固定Destination-Host/Realm及已审核命令，DRA可换、OCS不可换；realm-only保持SingleAttempt。ExplicitRetransmit可在Peer重连/另一DRA Ready后显式调用，不后台重放。RO-06专用跨目标/跨段续发按下节授权，不扩大普通Route权限。部署记录应附模式、原deadline/MaxReplayAge、终端去重期限、E2E Space及自定义源/恢复GuardSlots和class配额。

第十一轮符合性补充：TS 29.338 §4.6 为 S6c/SGd/Gdd 指定 SCTP。库保留 TCP/TLS 的基础能力及现有模拟器/部署兼容入口，但使用 TCP 的这些应用路径须列为具名兼容/偏差，不能因经 DRA 就自动宣称满足该传输条款；现场逐 hop 记录实际保护与承载。CER/CEA 的 VSAI 广告不能复制到这些应用消息中，发送/接收差异见 Profile 规则契约 §3。

### 2.4 RO-06多OCS与重启的最小部署材料（2026-11-04）

[运行时§1.7](diameterkit-runtime-contracts.md)冻结有限TargetID集合，每项独立给出终端Host/Realm、可用邻接、固定Profile、成功/错误来源及ReplayDomain；应用route/四模板槽映射必须落在此集合。初选可HA/轮询，续发逐请求授权。DRA并不自动代理计费会话迁移，也不能把邻接ready当终端或退款凭据已就绪。

本项只等待四组定点材料，不要求库调用在线facts或上传业务账本：

1. 目标OCS/DRA实际映射、同终端不同路径与不同OCS的区分；跨目标允许改写的Destination-Host/Realm配对，私有字段/版本解释一致性。
2. 去重键、重复内容比较、共享处理点及保存期，重复成功/失败答案的行为；保存期覆盖整个ReplayUntil，含错误后T=0续发、T=1无答重放和重启后的原Origin/E2E。普通CanonicalV1仍保留Destination及Origin-State；出站许可变更目的地**不放宽入站比较**，必须证明目标系统接受该差异而不重复扣/退，不能只以“共享去重”四字带过。
3. ECUR已有会话的许可/会话连续性；SMS未使用CC-Session-Failover或SCAP1时分别提供具名兼容依据，不能继承IEC或标准CCR许可。旧额外错误码/DDFH处理、六模板/四槽匹配也逐项验证。
4. 应用重启承接原Origin、记录年龄/次数、旧发送者网络隔离及号域方案；旧C++域外E2E按标识§4.7配置RecoverySpace，旧文件格式迁移归应用。当前CER状态与原重放请求Origin-State的差异必须验证，不承诺旧连接答案跨进程送达。

库构造只检查配置的确定性、授权边界和资源，不认证上述声明真实性、不要求生产抓包/人工审批才能跑测试。材料可来自受控OCS环境和可复现配置；缺证时该目标恢复不启用，RO-06全覆盖保留，普通SingleAttempt/机制开发继续。不能靠库内持久化弥补OCS去重域缺失。

## 3. R17：内容规范化与重复结果

> **第八批实现注记：** 本节当前Base/S6c子集已由[batch-08](../spec/history/batch-08.json)接通。可选显式DuplicatePolicy与NoCache并存，跨DRA保证只覆盖所有实际入口均启用缓存的相容绑定；完整角色/上下文指纹仍未实现。本节早期“提交后ctx取消停止未开始首份发送”的设计已由第六批采用的实现契约覆盖：提交后ctx不撤销发送，等待固定FrameWrite/关闭结清；第八批继续这一行为，见batch-08。

### 3.1 身份和 CanonicalV1

应用请求的重复键为 `Node 接收域 + ASCII-fold(Origin-Host) + End-to-End`。不加入 Peer、连接、App-ID、Command、T、Origin-State-Id 或 Profile 来把同身份冲突拆成两次业务执行；这些相关语义进入记录的内容/上下文比较。控制 CEx/DWx/DPx 归各自状态机，不用此应用缓存。T=0 也查重复表。以 Origin/End-to-End 检查重复的协议依据见 [RFC 6733 §3](https://www.rfc-editor.org/rfc/rfc6733.html#section-3)；以下字节规范化是本库的保守比较契约。

CanonicalV1 对通过 Wire 的内部原始证据产生有界、长度前缀编码的比较序列，不能从可修改的 Message 重新编码后比较。比较序列与 Profile/策略版本一同保存；哈希只加速查找，命中后做精确序列比较，不以哈希相等作为唯一凭据。

| 内容 | 比较规则 |
| --- | --- |
| Diameter 头 | 保留 Version、App-ID、Command、R/P/E；忽略接收规则要求忽略的保留位；忽略已经校验的总长度、Hop-by-Hop、T；End-to-End 在键内 |
| 传输 envelope | 连接代次、Peer、SCTP stream/PPID 不参与业务比较，保存在每次到达的 envelope |
| 顶层逐跳 AVP | 仅剥离标准 Vendor=0 的 Proxy-Info(284)、Route-Record(282)；仍独立校验、授权并保存本次应答所需代理链；嵌套或其他 Vendor 同码 AVP 不能被递归剥离 |
| Origin / Destination | Origin-Realm、Origin-State-Id、Destination-Host/Realm 均保留，包括“缺失”和“存在”的差异；已知 DiameterIdentity 按 ASCII 折叠；不把 Destination 自动归并为 DRA 或 realm |
| Session/业务值 | 保留 Session-Id、用户身份、计费编号/单位、TPDU、User-Data 等；XML 不做重排/空白折叠，TBCD/UTF-8/opaque bytes 不做“等价内容”转换，浮点也不改写其位表示 |
| 已知 AVP | 编码 Vendor/Code、有效标志、存在性、有效值；排除接收规则要求忽略的保留位和已审核路径的 M 差异；对具名应用 bitmask 使用相同有效位掩码。忽略已检查边界的 padding，已知 Grouped 递归处理 |
| 应用规定忽略的项 | 固定 29.338 Profile 的应用消息顶层 VSAI 不比较；只检查 framing/资源后保留诊断。不是逐跳剥离规则，不适用于 CER/CEA、Sh 或任意同码扩展；没有公共 ignore list |
| AVP 顺序与重复 | Grouped/扩展默认 Sequence；本轮标准命令顶层在固定位置核验后采用 IndependentKeys，可按 Vendor/Code 排列不同已知键，同键多次出现始终保持原序；固定位置字段先校验位置。禁止排序全部 repeated AVP 或用 map 覆盖重复项 |
| 未知 AVP | 未知 M 位先按 Wire 拒绝；允许保留的未知项按 raw payload/有效 flags 比较（规范要求忽略的保留位除外），不猜 Grouped。未知项是排序屏障：保留自身顺序和两侧已知项分段，只在段内按已声明规则排序 |
| 扩展/兼容 | 扩展描述符须提供 bounded canonical 规则；没有规则则保持原序/raw。禁止运行时 arbitrary ignore list、用户回调或根据 Origin 切换算法 |

因此：HBH/T/合法代理链、规定忽略的保留位/已知 M 差异/应用 payload 位，以及指定应用顶层 VSAI 的有无可合并；不同 TPDU、Session、CC 编号或未知 AVP 有效内容不能合并；已审核的独立字段换序可合并，同键 repeated 换序默认冲突。含未知项的合法换序可能保守拒绝，不能宣称支持任意代理改写。排除的逐跳字段仅可用于每次到达的准入/返回路由，不能再作为同一逻辑业务决定的输入；如业务确实依赖这些字段，须重新审计比较规则。DRA 若会加/改 Destination 或业务 AVP，必须先提交双路径样本，形成一个固定、具名且经语义审核的兼容 Profile；v1 默认不忽略这类差异。部署据此选择可共用的 DRA 路径，不以放宽去重掩盖不一致。

逐 Arrival 的 SourcePolicy/传输准入分别通过即可，不要求 DRA-A/B 的来源许可集合相同；这些许可不是规范内容或共享业务决定的输入。相反，角色、结果规则、去重期限或 class 的差异不能仅因线上字段相同就共享记录。

同一 Origin-Host 不同 ASCII 大小写会命中同键；Origin-State-Id 的变化不能绕过去重，发送者重启必须继续满足标识唯一性。

同键但 App/Command/WireFingerprint/InboundSemanticsFingerprint 或规范序列不一致，产生本地 `DuplicateConflict`，不交付第二次业务、不回放首次成功、不覆盖或延长旧记录。对可安全应答的副本，固定失败分支使用 5012 UNABLE_TO_COMPLY；优先应用错误语法，无法表达时使用规范允许的 E 位永久失败语法。此映射是本项目策略，RFC 没有专门的 DuplicateConflict 码；不能误用 5007 表示两个不同请求之间的冲突。记录有界的路径/规则诊断，不输出两份用户载荷。[RFC 6733 §7.1.5、§7.2](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.1.5)

### 3.2 一次业务决定，多份独立到达

每个记录拥有一个逻辑工作项和有界 Arrival 集合。每个 Arrival 保存自己真实的 ConnectionID、HBH、E2E、P、Session、Proxy-Info、传输信息、发送槽与期限；只能在该代连接写一次。业务只 Receive 第一个工作项；后续相同副本不重新调用 handler/Check，不延长处理期限。每个 Arrival 的来源/目的地准入仍独立执行。

本轮明确调整早期 Inbound 的生命周期：**Inbound 是首个请求的逻辑结果提交句柄，Info.ConnectionID 仍是首个 Arrival 的连接；Arrival 才是绑定连接的发送权。** Respond/Reject 提交一次不可变结果，再由库给各个已经实际收到的合法副本生成应答。没有新请求的重连连接绝不接收旧应答；重新完成 CER/CEA 还须通过 watchdog REOPEN 探测后才恢复新业务资格。这不是把旧 Arrival 重新绑定到新连接。

| 状态 / 事件 | 固定行为 |
| --- | --- |
| Absent → Processing | 原子插入键、规范内容和 C reservation，预留 I/W/M 后交付一次；任一失败不执行业务 |
| Processing 收到相同副本 | 预留该 Arrival 的 envelope/W/槽，加入等待集合；超过 MaxArrivals 或预算则只拒绝/关闭本次入口，不取消原业务 |
| 原连接断开 | 原 Arrival 终止，保留逻辑工作项至 ExecutionDeadline；已有/未来同键合法副本可等待同一结果；不重新交付业务 |
| Respond / Reject | 先检查逻辑状态、facts 及业务结果上限，再 CAS 提交；缓存业务结果并给可发送 Arrival 构造各自帧；失败的 Arrival 不影响其他副本 |
| Completed 收到相同副本 | 不读业务状态、不重做副作用；用缓存结果与本次 envelope 回答 |
| ExecutionDeadline 到期 | 与 Respond 争同一提交门；提交具名超时拒绝结果；不能编码时保留 Suppressed 标记并结束 Arrival；随后不接纳迟到业务结果 |
| Suppressed 收到相同副本 | 不重新执行；只能使用固定失败分支或关闭，不能构造成功 |
| KeepUntil 到期 | 停止复用；待 writer/编码引用结清后归还 C；库不承诺窗口外或进程重启后的去重 |

`ExecutionDeadline = FirstSeen + MaxProcessing`；`KeepUntil = ExecutionDeadline + ReplayRetention`，二者是本地单调时钟上的固定期限，均不被重传刷新。DuplicatePolicy 必选 MaxProcessing、ReplayRetention、MaxArrivals、MaxCanonicalBytes、MaxResultBytes、MaxEnvelopeBytes；不从 End-to-End 的四分钟唯一性要求推断业务缓存期限。MaxArrivals 限制同时保留的到达；关闭的 Arrival 结清引用后可释放其槽，重复洪泛仍受 Peer admission 速率和字节限制。

Inbound.Done 在逻辑提交、ExecutionDeadline 或 Node 关闭时关闭，不因单个 Arrival 断线关闭。Check 的内部凭据继续保留；处理超时仍会阻止成功提交；Check 冻结协议上下文；Respond 校验应用构造的应答分支。重复副本复用已提交结果，不重查业务状态；不再有通用实时 FactSet/RespondWithFacts，详见状态与一致性契约 §1。ResponseResult 分别报告 `Committed`（逻辑结果已固定）与 `PrimaryWrite`（首个 Arrival 的发送证据）；首连接已断开时允许 `Committed=true, PrimaryWrite=NotWritten`，其他副本由各自 Arrival 重答。Respond 不等待未来副本，不返回无限增长的发送列表；Respond 提交后若 ctx 取消，停止尚未开始的首份发送；已经开始则在固定写入/关闭期限内结清证据后返回，可能晚于调用 ctx 到期。WriteUncertain 仅表示底层证据确实无法判定，不用于“调用者不等了”而提前返回。当前到达数/失败计数和逐连接诊断由有界快照/指标表达。同一逻辑句柄的 Respond/Reject 校验临时区串行使用；并发调用返回 ResponseInProgress，不为每次竞争复制一份最大应答。校验结束后再竞争结果提交门，超时/Close 可能先赢；校验失败释放临时区但不消耗提交权。取得提交权后，即使 ctx 取消或首连接写失败也不能再次提交；提交前 ctx 已取消则不消耗句柄。

Peer 是跨重连的稳定接收对象；某 Peer 显式 Close 停止自身连接和 Receive，但已交付的逻辑 Inbound 仍由 Node 管理至原期限，可服务另一 Peer 已实际收到的合法副本。尚未交付的工作项由 Node 拒绝/抑制，不在关闭的队列里遗留无人处理的业务。Node.Close 终止逻辑句柄及其 Arrival，缓存随 Node 释放，不保留后台 goroutine。业务函数可能忽略取消，因此该协议缓存不提供数据库副作用恰好一次；延迟投递、计费和跨进程接管仍需业务幂等。超时拒绝只表示协议工作项未及时完成，不能被解释为外部投递或扣费必然未发生。

### 3.3 envelope、尺寸和 R21 的补充

缓存只存业务结果，不含首条 Proxy-Info、HBH 或连接状态。每个 Arrival 使用自己的 envelope 回填后再校验/编码；P/Session 等参与规范比较的字段相同也不能借用业务修改后的视图。语义结果和本地 Origin 固定，允许变化的只有逐到达路由/关联包络。[RFC 6733 §6.2](https://www.rfc-editor.org/rfc/rfc6733.html#section-6.2)

为避免“大代理链副本把已经完成的结果挤出上限”，binding 在构造期证明 `ReplyWireMax ≥ MaxResultBytes + MaxEnvelopeBytes + 固定头/必填开销`；上式各项按去重后的字段所有权分解，不能重复计算。缓存结果不占用 Arrival 的代理额度；超大 envelope 在准入时拒绝，不能临时截短 Proxy-State。结果超限在取得逻辑提交权前返回 ResponseTooLarge，可修正后提交。

R21 的 C 更新为 `记录元数据 + CanonicalBound(MaxCanonicalBytes) + CachedResultBound(MaxResultBytes) + 必要编码临时上界`，不是仅缓存 answer 大小。规范化 scratch 计入 I；命中比较可流式执行，若生成第二份序列也必须计入 I。首次保存的规范序列从 I 转移到 C 后再释放对应 I，不同时漏计或双计。C 取自首次命令的 Duplicate class；另一 class/策略命中同键必须冲突，不能借缓存跨分区绕过准入。

每个 Arrival 的解析 I、保留的 envelope/W/M 均归 Inbound 分区，Duplicate 分区只承担共享 C，不另收一份重答 W；逻辑请求证据和 facts 只保留一份至结果提交/超时，首连接断开不提前释放；校验/编码仍借用凭据时须结清引用后释放。C 在所有连接断开时仍保留至 KeepUntil。缓存不持有无预算的原始整帧，业务 Message 所有权交付与库内部凭据继续分开。故意耗尽缓存时拒绝新工作，不能 LRU 淘汰承诺窗口内的活记录。

## 4. R18：SCTP 补齐方案

### 4.1 依赖基线与上下游交付

历史核查为本地 `skill215/sctp` 提交 `f325fe38834ff20bcaa68e776bea4e00484bd33e`，审计基于release `0.7.1`。当时**GetStatus及Instreams/Ostreams已存在**，不是全部要新增；SCTPRead未暴露recvmsg flags，标准deadline返回EOPNOTSUPP。终止等待/共享地址缺陷来自当时复现，不能未经复核继续宣称存在于新版本。[历史接收实现](https://github.com/skill215/sctp/blob/f325fe38834ff20bcaa68e776bea4e00484bd33e/sctp_linux.go)、[历史状态/deadline](https://github.com/skill215/sctp/blob/f325fe38834ff20bcaa68e776bea4e00484bd33e/sctp.go)、[历史审计](https://github.com/skill215/sctp/blob/f325fe38834ff20bcaa68e776bea4e00484bd33e/docs/release-audit-0.7.1.md)。

2026-11-04纯Go复审读取本地快照`e4154064dc9928274bcf4615c5382dd6fd81ad21`：[README](../../sctp/README.md)、[go.mod](../../sctp/go.mod)、[sctp.go](../../sctp/sctp.go)、[sctp_linux.go](../../sctp/sctp_linux.go)、[socket_io_linux.go](../../sctp/socket_io_linux.go)、[sendmsg_linux.go](../../sctp/sendmsg_linux.go)，上述路径工作树无改动。其模块基线Go1.26.0、实现使用Go syscall/unsafe，不要求C库或lksctp-tools；README明确Linux内核SCTP且不支持Linux/386。当前三个标准deadline入口仍返回EOPNOTSUPP，因此纯Go依赖方向可用不等于满足U3。本轮未重跑上游测试、未认证其全部能力或认定该commit为最新发布；邻库其他未跟踪审查材料不作通过证据。

适配器与上游补齐均须保持CGO=0；ABI结构布局、socklen_t宽度、字节序和fd所有权仍需实测，不能以“无cgo”免除。宿主内核/网络支持由部署提供，库不自动加载内核模块、安装系统包或调用外部SCTP服务；容器也不提供独立于宿主的内核能力。未满足平台/内核/必需API时显式不可用。核心module不得间接导入Linux-only实现；SCTP独立module自身必须构建/测试，跨平台stub、交叉编译、QEMU ABI验证及真实association结果分开，详见[覆盖§41](diameterkit-usage-coverage.md)。

| 上游工作包 | 必需契约 | diameterkit 适配器责任 / 交付门槛 |
| --- | --- | --- |
| U1 稳定性基线/回归 | 新快照已有地址深拷贝、fd所有权/Close唤醒及reset后状态查询实现和测试；复用并回归，不重新当作全缺。新消息/取消路径也须满足同样保证 | 终止传播全部pending；路径故障不伪装答案。已有源码不等于本轮测试通过；接入SCTP-01 |
| U2 消息接收API；P1前置 | 当前SCTPRead内部得到但不返回recv flags，callback可能消费通知。需新增完整flags/元数据有效性、调用方数据与通知缓冲的通用接收入口；消费前安全分流，见§4.5 | 每连接一个reader仅用新入口；20字节前缀不吞通知，不混读旧SCTPRead/包装Read。适配器负责Diameter长度/EOR组装；SCTP-02 |
| U3 取消与deadline；P1前置 | 当前仅socket timeout快照，三个标准deadline返回EOPNOTSUPP，无DialContext。补持久可更新deadline、同ctx的解析/connect、关闭/取消结清，见§4.6 | 适配器给握手/帧/write绝对期限，核心给exchange/Shutdown总期限；不靠后台遗弃阻塞操作。SCTP-03/06 |
| U4 有限通知/选项；P1前置 | GetStatus/订阅及逐消息SCTPWrite已有；补所需typed通知、设置/读回选项、连接前配置入口及元数据解析有效性。仅当前使用子集，不要求全RFC6458接口 | 明确事件集合、level0/NODELAY/可靠消息模式、实际出流、RESTART/partial-abort处理；协议屏障归本库。SCTP-02/04/05 |
| U5 发布和可复现验证 | Go1.26.x固定工具链、CGO=0完整依赖/构建/普通测试/真实I/O，Linux内核/架构记录、版本标签；race单独启用检测工具所需cgo，ABI/泄漏独立证据 | adapters/sctp独立module锁定通过门槛的版本/提交；核心module不依赖平台SCTP；根目录测试不代替独立module验收 |

历史[v0.7.2计划](https://github.com/skill215/sctp/blob/f325fe38834ff20bcaa68e776bea4e00484bd33e/docs/release-plan-v0.7.2.md)不能当作当前进度表。本地[v0.7.3发布说明](../../sctp/releases/v0.7.3.md)明确该次主要修CLI出流选择，标准deadline/DialContext仍为独立功能工作；不因版本号认为U2–U4齐备。实施时评估可用最新候选并锁定已通过U1–U5的版本/提交，不保留旧API静默降级路径。上游发布不以Diameter应用验收为前置，避免与本库P3/P6形成循环依赖。

### 4.2 接收、截断与预算

v1采用Linux one-to-one socket，一个ConnectionID对应一个association；不实现one-to-many共享fd调度。显式设置并验证`SCTP_FRAGMENT_INTERLEAVE=0`，每次send构成一个user message，禁止多次send拼一条；因此每个association最多一个数据消息处于部分交付。这里要求的是发送语义，不能盲目要求Linux实现RFC里的每个选项：核对的[Linux v6.12 UAPI](https://github.com/torvalds/linux/blob/v6.12/include/uapi/linux/sctp.h)未定义SCTP_EXPLICIT_EOR，不将强行设置该选项成功作为门槛；也不能用SCTP_DISABLE_FRAGMENTS替代，它涉及网络分片，会破坏大消息能力。实际内核仍须证明一次send/一次message及level0行为。不能从unordered的SSN猜消息身份。[RFC6458 §8.1.20](https://www.rfc-editor.org/rfc/rfc6458.html#section-8.1.20)、[§8.1.26](https://www.rfc-editor.org/rfc/rfc6458.html#section-8.1.26)

一条 Diameter 帧作为一个 SCTP user message 收发，这是本适配器的互操作契约：适配器的 ReadPrefix 用固定小缓冲拿到 20 字节头，核心 runtime 按 R21 分类并取得 reservation，再将目标缓冲交给 ReadFrameInto，后续片段直接写入该缓冲直到 EOR；适配器不接收或访问核心 token；累积长度必须恰好等于头声明长度。提前 EOR、到达长度但未 EOR、超长、跨 stream/PPID 的不一致片段、TRUNC/CTRUNC 或缺失必要元数据均关闭 association。禁止仅因缓冲区填满就宣布完成，或把下一条 message 拼进上一帧。遗留端点若把一个 Diameter 帧拆为多个 SCTP user message，属于需独立审核的兼容差异，默认不接受。

通知须在消费前按NOTIFICATION分流到独立固定NotificationMax暂存，不能先用20字节数据缓冲消费、再期待旧回调补回完整通知；见§4.5。它不交给Diameter decoder，不得附到数据帧。检查通知头长度和完整性，未知但有合法边界的通知有界记录/跳过；截断、损坏、超NotificationMax及partial-delivery-aborted终止当前association，不随通知长度分配完整报文。数据组装与通知使用同一不刷新的部分消息绝对期限，限制通知处理速率；能力/通知缺失不能被日志掩盖。

level 0 降低了部分交付的组装复杂度，也意味着大消息部分交付期间存在接收 HOL；多流/无序不能消除此限制。通过 MaxFrame、部分消息时限及节点预算限制影响；不宣传“大 Sh 完全不影响控制消息”。若混合负载不能满足 DWR/OFA 延迟，再独立设计 level 2 的每流有界组装或具备明确 message ID 的方案，重新做容量证明，不能直接打开内核开关。

### 4.3 发送、stream 与连接屏障

默认 `UnorderedStream0`：握手屏障完成后，应用请求/应答用 stream 0、unordered；接收仍接受协商范围内任意 stream。具名兼容模式 `OrderedMultiStream`：控制用 0，业务使用本地 Ostreams 的 1..N-1，按固定 Session-Id hash 选流，无 Session 的命令按稳定请求身份选流；N=1 时使用有序 stream 0 并公开 HOL 限制。不取配置请求流数或 Instreams 作为发送上限；应答也独立选本地合法流。[RFC 6733 §2.1.1](https://www.rfc-editor.org/rfc/rfc6733.html#section-2.1.1)

| 阶段 | 发送/状态规则 |
| --- | --- |
| 建连 | 获取实际 Instreams/Ostreams，校验所需元信息/选项；不合法或缺能力则握手前失败 |
| CER/CEA | 使用有序 stream 0；回应方发完 CEA 后仍只使用有序 stream 0，立即排入 DWR |
| CEA 后屏障 | 发起方收到合法 CEA 后才发送业务；回应方收到对应 DWA 或其他合法的握手后消息，才切到配置的无序/多流模式；屏障期限不刷新，失败关闭 |
| Ready | 每个 Diameter 帧一次消息发送；所有发送经过同一有界 writer，stream 选择是每帧参数，不改变连接上的共享“当前 stream” |
| 本地 Shutdown | 停止新请求/新重传 admission，在 DrainDeadline 前排空原工作；进入 Closing 后不再开始应用写，结清当前帧再发有序 stream 0 DPR；匹配 DPA 到达即关闭，不追加 grace |
| 对端 DPR/DPA | 收到 DPR 停止所有未开始应用写，DPA 不等 worker；自己发完 DPA 后可在剩余 ClosingGrace 内等 EOF，不能开始业务写；收到匹配本地 DPR 的 DPA 则立即关闭 |

ClosingGrace 仅为 DPA 发送方等待 EOF 的有限清理期限，受 TotalDeadline 约束；收到匹配 DPA 的一方不使用它继续等待。它不构成跨 stream 的远端确认屏障，DPR/DPA 不能证明此前业务均已到达。库不为退出无限等待业务，也不能在旧连接关闭后把队列中的应答迁到新连接。Arrival 丢失后的逻辑结果仍按 §3.2 保存，只有新收到的副本能触发新路径重答。

发送默认可靠传输，不启用 PR-SCTP/TTL 丢弃。本库只调用Messages模式的WriteMsgContext；旧SCTPWrite仅作上游已有基础/兼容回归。完整接纳只记 Written，不证明对端交付；正数短写/不确定错误记录相应证据并终止 association，不将剩余尾部作为第二条 SCTP message 重发。SEND_FAILED 通知仅补充原 attempt 的证据，不自动调度应用重试。单路径地址不可达先作为多宿主路径状态处理；这类路径通知本身不进入R19；association终止及独立watchdog不可用仍按既有触发分类处理。RESTART 不沿用旧 Diameter 连接代，关闭并由管理者建立新代连接/握手。

PPID 在 Go API 内为 host order 46，底层负责一次网络字节序转换；未知 PPID 默认拒绝，历史 PPID=0 仅能具名配置，47 在未实现 DTLS/SCTP 时拒绝。NODELAY 明确开启并检查结果；不照搬其他库的立即 SACK 性能调优，SACK/RTO/心跳参数先保留内核默认，改动须测量。

### 4.4 取消与底层验收

标准 deadline 表示绝对时刻；非零 socket timeout 表示单次操作预算，两者并存时取较早时刻。单次ReadEvent内peek/重试不重新计算S；跨调用的整帧期限由本库绝对D/ctx固定，不能依赖每次重起的S；更新 deadline 应唤醒既有阻塞调用，清零/延长也须有确定行为。底层返回可识别的 timeout/canceled/closed/association-lost 分类；适配器只在明确知道尚未接纳时返回 NotWritten。

单个 exchange 的 ctx 取消不设置共享连接 ReadDeadline，也不必关闭 association；移除等待资格按 R19 完成。排队发送可撤销；已进入 Sendmsg 的发送必须在独立 WriteDeadline 内结清证据，必要时关闭整个 association，其他 pending 同时得到传输失败。连接reader使用握手、部分帧和关闭期限；Tw由Peer独立调度，不能直接当作读超时跳过SUSPECT；连接生命周期取消和 Node.Close 必须解除 I/O。无法取消的解析/拨号不能仅后台丢弃其结果，DialContext 失败后不得遗留连接/fd。

能力探测只能尽早报告不满足，不能替代 Linux 内核实测。验收先跑 U1 原缺陷红/绿回归，再跑 U2/U3/U4 每项行为，首版最后跑双DRA大Sh + 延迟OFA + S6c请求的混合负载；Ro实际模板/重放/guard负载同时纳入首版；记录 Go 补丁版、依赖 commit、内核/架构、socket 选项、真实协商流数和抓包。普通 SCTP 未交付前不以 TCP 测试代替；仍不宣称首版具有 DTLS/SCTP 或完整 RFC 6733 安全符合性。

### 4.5 U2：先区分数据/通知，再消费（纯Go，P1）

2026-11-04具体输入/输出已收敛为[上游接口契约§2–3](diameterkit-sctp-upstream-contract.md)：显式Messages模式、ReadEvent/WriteMsgContext、固定IORequirements、ReadResult消费/错误状态及缓冲所有权。以下为责任概要；同主题以该契约的精确返回表为准。它仍是待实现接口，不表示当前上游已有API。

本次源码审查固定在§4.1的新快照。`sctp_linux.go:198`的SCTPRead内部读取recvflags，公开返回只有n/info/error；`info.Flags`是SCTP ancillary flags，不能替代MSG_EOR/MSG_NOTIFICATION/MSG_TRUNC/MSG_CTRUNC。handler非空时通知在内部调用回调后继续读，handler为空也没有足够标志让上层可靠识别。故现有入口即使能传输字节，也不能满足本库前缀/预算契约。

**上游交付一个通用有界消息入口**（现采用ReadEvent形状，等价命名可在上游API评审调整，不新增本库公开API）：

1. 输入分别为调用方拥有的data目标区、固定notification区及有界ancillary空间；输出tagged Data/Notification/EOF、实际n、recv flags与独立的元数据有效性。stream/PPID为0与未取得元数据不同；malformed CMSG、CTRUNC、未知合法CMSG分开。已有SNDRCV格式足以提供所需字段时可复用，不要求先实现所有RCVINFO/NXTINFO变体；新入口声明实际选择并严格检查。
2. 在一个上游读所有权区间内先非消费分类，再对合适缓冲消费；本项目采用小型MSG_PEEK分类→同一读锁下recv的实现方向，最多读取固定通知头作分类，不分配整条数据。新旧reader不并用，禁止通知callback抢先吞事件。peek flags只作分类提示，最终长度/EOR/元数据以真正消费结果为准；peek、等待、消费共用期限，失败不能留下使用调用方缓冲的后台操作。
3. 小数据缓冲可以产生有界分段，不能仅n==len(dst)便推断完整。通知在独立固定空间内完整返回；不足以接收/组装时显式报告超限或不完整，由本库关闭association。上游不识别Diameter的20字节头、App-ID、pending或预算class。
4. NotificationMax至少容纳所选必需通知的固定头/结构，由启动检查；SEND_FAILED等可携带原消息大payload，不能因此自动扩大暂存到任意Frame大小。超限保持保守发送证据并关闭，不复制返回的整份业务载荷到日志/Completion。大小及通知频率上限仍由适配器设置；通知不递归解析Diameter payload，无额外深度解析器。

该设计方向依据[Linux v6.12 sctp_recvmsg](https://github.com/torvalds/linux/blob/v6.12/net/sctp/socket.c#L2090)的MSG_PEEK及分段路径：小缓冲可能返回尚未EOR的通知片段；现文“前缀20字节/独立通知区”若没有消费前分流便不闭合。这是本库发现的接口依赖，不是声称当前上游违反其已承诺API；锁定目标内核上仍须验证peek/消费、通知夹数据、Close和短读组合。[RFC6458 §9.8](https://www.rfc-editor.org/rfc/rfc6458.html#section-9.8)也区分输入peek与返回msg flags。

### 4.6 U3：复用poller，补绝对deadline与取消（P1）

具体契约见[上游接口§4](diameterkit-sctp-upstream-contract.md)：有效期取D/C/S最早、setter/owner代数、九类竞态及errors.Is映射。Resolve与Dial传同一ctx才覆盖端到端解析/建连；新Context拨号拒绝用户Control回调，防止不可取消同步回调破坏承诺。成功I/O与取消相遇保留实际进度，交接成功后dial ctx不再拥有连接。

当前`socket_io_linux.go:98–150`把SO_RCVTIMEO/SO_SNDTIMEO变成一次操作的绝对时点，用方向gate和os.File poller等待，操作结束清临时deadline；通知循环不刷新这次预算。三个公开SetDeadline仍返回EOPNOTSUPP；`sctp_linux.go:603`的connect等待使用ppoll及socket timeout，无ctx唤醒；ResolveSCTPAddr也无ctx。因此不能用SetSockopt timeout或后台调用旧Dial来冒充新契约。

| 上游负责的机制 | 确定行为 / 本库调用边界 |
| --- | --- |
| 持久read/write deadline状态 | SetDeadline同时更新两方向；单方向setter不影响另一方。更新/清零/延长须唤醒poller及排队gate并重新计算有效期，setter不得等待读锁/写锁释放。旧操作defer清理不得清掉后来设置的deadline |
| deadline与socket timeout组合 | 有效期=min(当前标准deadline，ctx截止，操作开始时固定的socket timeout截止)；unset项不参与。标准deadline动态更新，旧timeout快照不因通知/EINTR/gate唤醒而刷新；清标准deadline不解除仍有效的timeout。旧入口错误兼容另列，新入口区分deadline exceeded、socket timeout、cancel、closed和association lost，保留cause |
| Context解析与Dial | 通用的Resolve…Context与Dial…Context贯通同一总ctx；预解析地址仍深拷贝，DNS/各地址/创建/连接等待/返回交接都计入总期限。多地址不是给每个地址重开一份期限；返回连接前交接与取消只一个结果，失败清所有fd/取消唤醒资源。业务目的Host/Realm不由这里推导 |
| 取消与资源生命周期 | 基于现有非阻塞fd/Go poller增加可唤醒状态，或使用同样可证明关闭的有限唤醒机制；不fork一套C阻塞栈，不遗弃goroutine。I/O所有权结清后返回；EINTR/状态查询循环不得无限绕过取消检查，Close唤醒读、写和gate排队者，不误关复用fd |
| Listener | 已有Accept/Close唤醒能力复用并回归；本库连接管理者拥有listener及关闭。首版不要求为了此需求新增通用AcceptContext/one-to-many管理器；独立连接的ctx不能关闭其他共享listener使用者 |

本库适配器是连接deadline的唯一协调者：握手使用总截止，首个Data字节启动的FrameRead覆盖ReadPrefix/ReadFrameInto/DiscardFrame及中间事件，通知不刷新；writer使用独立写期限。适配器固定socket timeout为0，Tw不直接设置读deadline，详见[运行时§3.5.2](diameterkit-runtime-contracts.md)。核心runtime管理每个exchange的ctx/Tx/Tr及Shutdown总截止；单请求取消不改共享ReadDeadline，进入发送后按原写证据契约结清。上游不接收RequestID、Tx/Tr、业务重试或RO-06记录。

### 4.7 U4及适配器责任清单（P1，有限子集）

| 能力 | skill215/sctp负责 / 当前状态 | diameterkit负责 |
| --- | --- | --- |
| 地址/终止/Close | `address.go:26`深拷贝；dial/listen调用该函数；RawConn pin、Close唤醒、reset后状态查询已有。U1复用并将新入口加入回归，不默认重写 | 将真实association终止交给Peer/所有pending，保持一次完成/资源释放；不把某条路径down等同整个OCS或业务失败 |
| 消息发送与PPID | SCTPWrite已有逐消息info拷贝、PPID host→network转换和单次SendmsgN；返回准确n/error，不对正数短写自动发送尾部。与U3共同保证等待/取消边界；所有权及无静默分配放大可验证 | PPID46/具名0、禁止未支持47；每帧完整user message、可靠发送/TTL=0及合法flags；选本地Ostreams范围内stream，记录Written/Unknown而非对端已处理 |
| 状态与初始化 | GetStatus/Status流数、InitMsg、SubscribeEvents/读回已有。补所需选项的安全有类型设置/查询及准确长度/错误；复用库内部连接前选项阶段；Messages模式拒绝用户Control回调，不能向本库暴露必须复制的unsafe ABI实现 | 配置值在int→uint16等窄化前检查；连接前设置所需事件和level0/发送语义，accept后核对继承值、GetStatus及实际流数，必要配置不满足则在CER前拒绝。内核ESTABLISHED/COMM_UP不直接等于Diameter Ready |
| 通知解码 | 当前仅通知头/常量/raw handler，缺完整typed入口。具体集合固定为ASSOC_CHANGE、PEER_ADDR_CHANGE、SHUTDOWN、PARTIAL_DELIVERY及旧SEND_FAILED，SNDRCV + SCTP_EVENTS接收路径；按[接口§5](diameterkit-sctp-upstream-contract.md)检查长度/assoc/字段，raw tail明确借用。新版EVENT不作为前置或自动fallback | RESTART终结旧连接代；COMM_LOST/关闭/partial-abort明确失败；路径变化仅更新路径信息；SEND_FAILED只有可靠关联时补原attempt，否则有界连接诊断，不自动重试或变更已完成结果 |
| 选项支持与证据 | 上游提供NODELAY、fragment-interleave、必要订阅/接收info及可靠发送配置的有限安全接口；set/get失败明确。没有的内核选项不编造常量或返回假成功；不要求全RFC选项框架 | 选择所需集合、验证实际生效，启动能力快照用于准入；不把所有可选事件/优化变必需。SACK/RTO调优、PR-SCTP、level2、DTLS/SCTP仍不在此次补齐范围 |
| 协议与预算 | 上游不依赖Diameter；只保证调用方缓冲/元数据/通知上界与I/O完成，无协议持久化 | 20字节头、消息长度与EOR一致、Reader/Writer、公平性、CER/CEA屏障、DPR/DPA由本库实现；TransportReserve计入prefix、notification、ancillary及底层固定scratch上界，A/I/W等归核心账本 |

不得由本库另用SyscallConn混读recvmsg、复制上游fd/poller来“临时完成”U2/U3；有限选项的底层封装也归上游维护。上游测试只验一般SCTP行为，不引入DRA/计费用例或等待本库通过后才能发布。用户授权允许以后修上游，但本次仍只更本文档。

### 4.8 实施依赖、优先级与结束条件

本节细化既有U1–U5/P3/P6/P7，不新增第六个上游工作包，不扩大成全部RFC6458、用户态SCTP或通用持久恢复。完整有限验收见[覆盖§42](diameterkit-usage-coverage.md)。

交给skill215/sctp的可执行范围现固定为[接口契约§6五行清单及§7验收](diameterkit-sctp-upstream-contract.md)；已有范围内可等价调整API拼写，不再把通知布局、取消竞态或缓冲所有权留给实现临时猜测。消息模式与旧handler/用户Control互斥，使用库内部选项阶段；原非消息模式Control接口保留。

| 顺序 | 交付 / 可以先做什么 | 出口 / 阻塞范围 |
| --- | --- | --- |
| 1 基线与契约；U1/U2–U4 | 核对候选源码及原U1回归；固定消息/通知缓冲、deadline和有限事件/选项契约。当前文档已给行为要求，API拼写可在实现定 | U1已有实现不记新功能缺口，但改动后必须回归；不把源码/上游旧测试报告当本库通过 |
| 2 上游通用能力；U2/U3/U4 | U3以当前poller/gate为基础，U2/U4可按固定契约开发，但最终必须在同一候选提交组合验证 | 缺任一必要flags/分流、取消/deadline或通知/选项，真实SCTP adapter保持不可启用。不能只发布一个deadline补丁便宣布可接入 |
| 3 上游发布门槛；U5 | Go1.26.x/CGO=0真实SCTP、ABI、取消/Close/泄漏及旧新API回归；race独立。锁定可安装版本/提交，保留来源及目标内核/架构 | 上游门槛不等DRA/OCS；版本标签不替代功能验收。本库不浮动依赖latest，也不要求先发布正式tag才能锁定已验收提交 |
| 4 本库适配器；P3 | 契约明确后可先以可控传输实现分帧/预算/策略测试；正式socket接入等上一步候选。核心codec、TCP/TLS及四应用构造/校验可继续 | 真实association证明数据/通知事件续调、stream、PPID、握手屏障、关闭及预算组合；包括Tw不截断reader、SCTP shutdown不发DPR/DPA、旧代I/O退出后才还槽。本库不扩上游到协议策略 |
| 5 独立互通/发布；P6/P7 | 四应用及Ro模板/恢复与DRA混合负载，超限通知/大Sh HOL/延迟OFA、资源回收和正常CGO=0产物测试 | 通过既有SCTP-01–06/BUD/PEER及发布门槛才声明首版SCTP完成；上游通过不是本库互通通过，生产部署另验 |

**本轮交叉审查的实施阻塞表（不新增工作包）：**

| 优先级 / 归属 | 当前阻塞与解除条件 | 不阻塞的工作 |
| --- | --- | --- |
| P1，上游U2–U4→U5 | 新消息/通知、动态deadline/ctx及有限选项仍须在同一候选实现并通过上游十组验收；只有文档不能启用真实SCTP | 核心codec、TCP/TLS、四应用已明确字段及可控transport测试 |
| P1，本库P3 | 接收结果事件分支、Peer原因映射、BeginWrite/关闭、构造期总预算及旧代槽持有尚未实现；按运行时§3.5和Peer§3.6验证，不能只有上游测试 | 与上游并行实现/验证模型和可控I/O接线；通过候选后立即真实集成，无下一轮文档前置 |
| P1，P6/P7 | 目标Linux内核/架构真实association、ABI/纯Go/race、四应用混合负载及独立互通证据尚缺；对应检查Passed才解除首版SCTP阻塞 | 上游候选独立发布；其他已验收子能力的具名v0.x发布 |
| 部署条件 | 真实DRA/OCS绑定、保护与恢复许可仍按既有条目取证，缺失仅阻止对应启用/投产或兼容全覆盖声明 | 不转为库数据库/业务重放任务，不暂停无关库机制 |

实施顺序固定为U1回归→U3生命周期与U2/U4有限能力→U5同候选；本库P3模型接线可并行，真实接入随后，P6/P7收尾。所有行当前仍Planned；新版通知布局、level2、全选项和跨平台栈不是阻塞项。

源码追溯：新增核对[address.go](../../sctp/address.go)、[生命周期测试](../../sctp/sctp_lifecycle_linux_test.go)、[地址测试](../../sctp/sctp_stability_address_linux_test.go)、[终止测试](../../sctp/sctp_stability_liveness_linux_test.go)、[timeout测试](../../sctp/sctp_timeout_linux_test.go)及[v0.7.3验证记录](../../sctp/docs/release-validation-v0.7.3.md)。对应13个读取路径工作树干净，本轮未运行测试。关键SHA-256：sctp.go=`f1d185f6c4b788f7649272911d4b226ba825df27ff13d053bbf59327015e79f8`，sctp_linux.go=`28719daca46157aefb93aedec361a509e929bb59002121142fabfc80513a814c`，socket_io_linux.go=`43e967204e132b7cde8e36e10b4b64ea260a0b1305927b0c44c051c48a668780`，address.go=`0688d588996d95a330ce825c8b810ca4b11801796a1000add13c163ad1f68cf6`。这是静态基线，不能替代U1–U5的新候选执行报告。

## 5. 新增验收细项与剩余缺口

首版DRA绑定、去重和SCTP能力仍需完整验收；OCS身份/重复域、六模板和SCAP1部署表现在同属首版P5/P6门槛。非Ro能力的实现可独立开展，v0.x子集须明确缺项；不能跳过Ro而宣布v1完成。

| ID | 必测行为 | 关联/阶段 |
| --- | --- | --- |
| BIND-01 | 同 App 不相容定义拒绝；固定组合 Profile；主动 PNR/ALR/ASR 无 pending 仍唯一解释；App 0 兼容命令不进控制保留区 | R03、DRA-03；P0/P2/P6 |
| BIND-02 | DRA 不是最终 Origin；SourcePolicy 每份副本都执行；动态 MME 目标越界、非法成功来源、合法代理 E 位错误 | DRA-01/03；P3/P6 |
| BIND-03 | 真实多实例/连接映射；相同 realm 不证明恢复域；重传跨进程时应用提供幂等或分流证据，库无共享存储 | R03/R19；P0/P5/P6 |
| DUP-01 | HBH/T/代理链变化合并；P/App/Command/Session/TPDU/CC 编号/Origin-State/未知 AVP 变化冲突；同码异 Vendor 不误剥离 | R17、DRA-02；P2/P3 |
| DUP-02 | IndependentKeys 与 repeated 顺序、Grouped 递归、未知屏障、XML 原字节、hash 碰撞后精确比较 | R17；P1/P2 |
| DUP-03 | 双 DRA 同时到达只执行一次；首连接断开仍提交缓存；无新入站不得给新连接发旧应答；超时/Respond/Close 争一次决定 | DRA-02、VAL-02；P3/P6 |
| DUP-04 | 内容/结果/代理链上界；缓存满不淘汰承诺记录；多副本资源释放、固定保留期、跨 class 冲突与重启边界 | R21、BUD-04；P3/P7 |
| SCTP-05 | 默认无序 stream 0 与具名有序多流；非对称/单流；CEA 屏障和 DPR/DPA 超越；level 0 大消息 HOL 上限 | SCTP-02/04；P3/P6/P7 |
| SCTP-06 | deadline 更新/清除/延长，socket timeout 组合，通知洪泛不刷新；拨号取消、短写无尾部重发、RESTART 和路径故障区分 | SCTP-01/03/04；P3/P6 |

R03、R17、R18 现在有明确配置边界、算法、状态和补齐工作包；仍不标记验收关闭。剩余工作是：收集真实 DRA/OCS 部署证据，逐 Profile 审核 IndependentKeys/兼容改写/错误分支，验证新的 Inbound 语义和资源上界，实施 SCTP 上游能力及端到端互操作。具体协议修订、Sh-Data 覆盖、Ro 消息/结果解释与性能门槛仍按既有 P0–P7 推进，不在本轮假定完成。

第十八轮兼容证据见[主流程§14.3–14.5](diameterkit-mainflow-audit.md)：六模板及编码/选择链固定14项本地文件哈希，部署还须明确route/global四槽实际选用的文件、版本和对端身份；仓库文件名不证明线上启用。标准RDR/ALR仍绑定S6c，旧SGd alert另具名；SCAP1源码的271/0头与VSAI193/19302分开，不把App0所有命令都路由到基础控制。模板420类型冲突要求独立解释域及证据，不能在标准字典中覆盖。完整报文/答案、DRA/SCTP和真实保护条件仍须独立验证，本轮没有修改SCTP库。


第二十轮三应用部署切片见[主流程§16](diameterkit-mainflow-audit.md)：Sh的目标/Origin不能从相邻DRA CER照抄，反向PNR由HSS终端授权及AS目标决定；SNR是否在本进程发过不作为PNR认证。MME/SGSN发ALR仍用S6c App16777312，SGd/Gdd MO/MT用16777313；旧SGd Alert须在具名兼容解释域验证。移动ALR的Origin与Serving-Node可不同，后者是数据，不自动授权新连接。原DRA/SCTP工作包和发布/部署证据边界不变。


第二十一轮把旧Alert解释域具体命名为`legacy-sgd-alert-v1`，仅对App16777313/Command8388648的ALR/ALA生效；标准域不启用时不接受该别名，答案保留请求App。本地Origin由endpoint注入，不从请求Destination复制。SGSN→Router与Router→GMSC候选是两个由应用创建的endpoint exchange；它们不代表库实现转发代理或token映射。IMS-SRA仅号码节点不能自动决定Diameter目标；后续TFR必须另有获授权的目标binding，来源和剩余现场证据见[主流程§17](diameterkit-mainflow-audit.md)。


第二十二轮设计的`s6c-ip-sm-gw-realm-v1`只覆盖S6c SRA的直接Serving-Node三元组Number+Name+Realm；默认标准域仍报告组合未决。部署选用须固定解释域/规则版本、端点身份与独立对端证据，不按DRA来源、字段有无或重试自动切换。收到的Name/Realm不修改Peer/授权表，后续TFR仍经过已授权binding。Sh混合11/12/25须有目标HSS的Notif-Eff依据；HSS-ID查询不证明HSS缺IMSI，真正无IMSI成功SRA仍不得以部署配置伪造标准取值。详见[主流程§18](diameterkit-mainflow-audit.md)。


第二十三轮[逐叶清单](diameterkit-field-inventory.md)固定本轮读取的三应用工作树入口与关键文件摘要；旧模拟器成功不等于HSS数据/错误分支正确。迁移重点核对零诊断、TFR节点号码归属、Sh结果命名空间、旧PUR额外PublicIdentifiers和PNR真实User-Data。Diameter节点国际域名在线路上须为ASCII表示；库不自动IDNA转换或由可解析Name授权路由。现场采用的深层Sh/国际化范围另行明确，默认不宣称全模型支持。

第二十五轮[主流程§21](diameterkit-mainflow-audit.md)补T4的SGd-OFR/OFA具名切片：IWMSC/MTC-IWF角色、App16777313、SCTP、目的MTC-IWF号码和供计费身份条件由部署/调用语义明确提供，不能因App相同混用MME→IWMSC语义；同连接/App仍必须有唯一相容解释，不能按3111是否存在猜接口。DRA CER不证明终端Sh特性，5011不授权自动改绑定或重试。新增T4证据不扩大到原生T4设备触发，Ro首版范围另按当前Specification执行，不由这条T4证据扩大。

第二十八轮[包络与关联](diameterkit-mainflow-audit.md)不增加部署类型：ErrorAgentOrigins按固定错误分支授权，不能因为DRA能报E3002就接受其终端2001；294也不授予来源。正常与E答案的Origin来自应答方已配置身份，原请求Session与本跳Proxy顺序保留。错误App/Command/E2E不能改换pending绑定；跨连接HBH同值不是跨Peer恢复依据。九族具体输入见[覆盖](diameterkit-answer-envelope.md)，兼容及独立DRA验收仍Planned。

第二十九轮[CR与覆盖](diameterkit-mainflow-audit.md)定位CR0069/C4-264035为Realm后续证据线索，只有门户元数据、无正文和TSG批准记录，不自动升级固定Rel18 Profile。s6c-ip-sm-gw-realm-v1仍只覆盖SRA/2401既定三元组，其他组合/Additional/ALR不扩展。C29两Proxy控制表示本跳输入，未知SF报价不授予终端能力、来源身份或新路由。

第三十轮[主流程§26](diameterkit-mainflow-audit.md)区分raw与Canonical：Origin-State缺省/0/值变化仍属于内容比较，不改重复键；顶层0/282仍剥离，但每次Arrival的字段校验/入站准入不能跳过。282中的自述身份不授权来源或新Peer，294不证明结果实际改写者，DRA CER的状态不代表其后HSS。CR0069公开ZIP仍403，固定Realm域不升级。

第三十一轮[目的绑定§27.2–4](diameterkit-mainflow-audit.md)给九请求合法Identity但非法Host/Realm配对反例及九E3002。LocalDestinationPolicy的配对限制是项目准入，不能冒称RFC把本地Host/异realm定义成坏AVP。七种命令CCF允许省略Host不等于所有角色均可realm-only：29.338 §5.3.1.2下HSS发ALR仍要求双目的字段；库不实现其MAP回落。六个条件控制、大小写/显式别名、交叉配对和每Arrival准入见完整输入；本地别名不授权新答案来源、不自动归并Canonical。

## 第五十批机制与部署门槛

普通R33的Ro同OCS机制可测试/使用，实际部署启用仍须固定终端、目标去重键/内容比较/保留期及路径证据。ReplayDomain Target/Retention是调用方声明，不是库向远端认证的结果；证据未知保持SingleAttempt。完整跨段/重启/多OCS能力未由本批提供。详见[本批契约](../spec/history/batch-50.json)、[独立证据](../interop/evidence/2026-11-07-ro-recovery.json)和[进度](../spec/progress.json)。

## 第五十八批实施证据更新（2026-11-09）

上游实际基线已推进至远端v0.7.4/62ad388，原e415406分析仅为历史参考。标准deadline和内部ctx等待基础已在真实Linux SCTP执行全量回归，257条测试/子测试零skip；源码补丁、hash、race/纯Go及交叉编译边界见[batch58](../spec/history/batch-58.json)。U3公开Context解析/拨号、U2/U4、同候选发布及本库SCTP适配/真实DRA互通仍未关闭；不把本地上游测试当部署互通证据。

## 第五十九批实施证据更新（2026-11-09）

上游同候选已补Context解析/拨号总预算及成功交接前后的取消边界，真实连接拒绝/占用、fd回收和accept竞争均有测试；受控DNS及局部待连接模型与实际association证据分别记录，见[batch59](../spec/history/batch-59.json)。补丁依赖第58批，尚未发布；消息分流、通知选项及本库适配/真实DRA验收仍不能跳过。

## 第60批SCTP候选状态（2026-11-09）

上游已接通消息模式及U3取消门，真实Linux/amd64回环验证分片/通知分流和最小选项继承；284条测试/子测试/示例零skip。受控通知/短写故障注入只证明局部边界；五类typed通知、实际长通知/地址/失败/partial事件、其他native ABI以及现场DRA/OCS仍待验。精确候选为62ad388加58→59→60补丁，未发布/未锁依赖；本库无SCTP适配器，不能据此启用生产。证据见[batch60](../spec/history/batch-60.json)。

## 第61批SCTP候选状态（2026-11-09）

当前Linux/amd64已原生验证五类typed通知，含65532字节旧SEND_FAILED超限不消费后重读、148字节地址和24字节partial-abort。全量312条测试/子测试/fuzz种子/示例零skip，真实事件和选项/accept用例20轮race通过，见[batch61](../spec/history/batch-61.json)。这些是本地内核事件，不是DRA互通或所有路径故障；当前关闭U4最小链路，U5完整预算/平台/发布仍未验。精确候选62ad388加58→59→60→61，未提交/发布、未锁依赖，本库仍没有SCTP适配器。

## 第62批U5交付状态（2026-11-09）

资源账目、ABI和取消退休检查、确定性离线模块及独立consumer真实收发已验，见[batch62](../spec/history/batch-62.json)。8192逻辑固定额度不代替实际adapter预算或RSS容量规划。候选为62ad388加58→59→60→61→62，离线标签v0.0.0-diameterkit.u5.62不是真实发布；上游既有Ubuntu/Go/架构及强制namespace远端矩阵仍需对这份精确候选执行。未新增本库SCTP依赖/适配器，不据此宣称DRA/OCS互通。

## 第63批远端候选状态（2026-11-09）

[PR #30](https://github.com/skill215/sctp/pull/30)的14项远端矩阵已通过，真实候选伪版本可下载且430文件匹配，独立消费者真实收发通过，见[batch63](../spec/history/batch-63.json)。Ubuntu原生、Windows/Linux386所有权和QEMU ABI的证明范围分别记录，QEMU不替代其他架构原生SCTP网络。候选满足开发接线条件，但draft PR未合并、无正式tag，本库尚无SCTP适配器，DRA/OCS验收仍未通过。

## 第64批本机SCTP状态

已实现PPID46/default stream0、通知续调/关闭和CEA屏障；真实Linux/amd64 Base/S6c、取消、关闭通过。仅本机端点证据，非DRA/OCS现场认证；OrderedMultiStream、PPID0兼容、SCTP Dialer/Listener及正式核心/adapter发布仍待后续。具体范围/证据见[batch64](../spec/history/batch-64.json)，历史未实施表述按对应时点阅读。

## 第65批后上游增强交接

有限SCTP Dialer/Maintain和Listener/Serve已于[第65批](../spec/history/batch-65.json)完成本机验证。本次维护者选择先增强skill215/sctp，再迁移本库以删除底层重复逻辑；具名多流/PPID0后续仍由Diameter配置绑定，上游只负责通用SID/PPID/实际流数与消息完整性。交接和验收见[需求](sctp-upstream-improvement-plan.md)。未取得新候选和消费方回归前，不改变现有版本锁或宣称现场已验。

## 第66批候选与迁移状态

已核对上游bb739b612088的A–E、14项远端CI和70份Go源哈希，消费真实伪版本v0.7.5-0.20261109072213-bb739b612088并完成适配器迁移。真实Base/S6c、取消/关闭、选举/REOPEN和无自动业务重放回归通过；上游仍draft未合并，核心离线版本未公开，DRA/OCS现场验收未由本机测试代替。具体证据见[batch66](../spec/history/batch-66.json)。

## 第72批实施更新：OrderedMultiStream与PPID0（2026-11-10）

§4.3两种流模式已实现于[adapter](../adapters/sctp/README.md)。配置StreamMode明确选UnorderedStream0或OrderedMultiStream；PPIDPolicy独立选默认46或PPID0Compatibility（发送0、接收0/46，均拒47）。不跟随远端包自动切换模式，也不把PPID0当成缺元数据；元数据Presence与片段连续性继续由上游判定。

业务按本地实际Ostreams选择，固定FNV-1a64的域1+Session原字节，或域2+AppID/CommandCode/E2E线序字节；排除HBH/R/T，不按对端SID回送。请求流数零采用默认1/16出流、16入流，允许显式请求，最终使用协商值。N=1有序stream0回退及实际双向流数通过ModeReporter公开。接收不限制为本地选流规则；任意合法入流均可。

CEA后屏障与DPR/DPA生命周期复用核心，直接WriteFrame也固定有序0。新模式不保证跨流DPA之前的业务已交付，不改变总关闭期限、一次完成或重放授权。真实非对称协商、本机Base/S6c、短写/取消/关闭和预算验收见[batch72](../spec/history/batch-72.json)；上游A–E不重开，模式仍需目标DRA/OCS部署接受。核心/adapter公开发行、压力/性能和其他应用缺口继续单列。
