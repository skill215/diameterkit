# Diameter 固定 Profile、协议规则与四流程审查

> **当前范围（2026-11-04，Draft 0.20）：** 第一阶段=首版/v1，包含Sh、S6c、SGd/Gdd、Ro/DCCA及sms-fw实际所需六模板、SCAP1和RO-06恢复能力；Go 1.26、纯Go实现（含依赖与SCTP适配器，发布路径CGO_ENABLED=0）、公开Go库、无持久化边界不变。明确撤销第十九轮“Ro后置”分期；带日期的旧分期/统计仅作历史证据，不再授予Ro跳过首版验收的例外。限定S6a、完整SCUR/PRICE_ENQUIRY等未承诺能力仍后置。范围及剩余缺口见[Specification §3/§7](diameterkit-specification.md)，顺序见[总设计§14](diameterkit-design.md)，验收见[覆盖§39](diameterkit-usage-coverage.md)。2026-11-06已完成前十九批有界实施（Base/S6c基础、公共包络与关联、出站运行时、关闭与选举、TCP/TLS生产适配、有限Inbound/应答、有限Reject/应急预算、Canonical/有限重复缓存及多Arrival重答、固定同终端Route/多attempt与活动请求原身份重传、有限Dialer/Maintain拨号及重连、按class/编码字节DRR调度、固定Peer的有限Listener/Serve、有界标识源/Session/guard、Base/S6c有限Context/Report、S6c普通RDR/RDA及HSS ALR/ALA、SGd/Gdd普通OFR/OFA与TFR/TFA、Sh普通四组命令及透明User-Data、Ref24/25有限Sh-Data helper、Sh单订阅/退订及Ref11/12有限数据）；已执行范围及未完成项以[Specification §6](diameterkit-specification.md#6-需求与证据的维护)和[spec进度](../spec/progress.json)为准，未列明的运行验收仍Planned。

> **第十四批实际落地边界：** 当前Base/S6c仅零Context；codec.ValidateContext/ValidateAnswerContext、Inbound.Validation/Check及Result.Validation已接入。Report是限定切片的阶段级/首失败报告，NoContext未审核不视作缺facts。下文Sh/Ro工厂、非零输入和完整NeedsApplication转换仍为目标契约，随对应应用实施；详见[batch-14](../spec/history/batch-14.json)。
日期：2026-10-31（第二十八轮补九答案族正常/E包络）。第十一轮收敛 R24/R31 的上下文入口和 R29/R34 的规则表达；第十二轮补充 [主流程字段与依赖审计](diameterkit-mainflow-audit.md)，更新原件、XSD、应用覆盖和结果分支状态。第十三轮接入 [库级 specification](diameterkit-specification.md)，补充主流程审计 §9 的字段切片。第十四轮补§10的身份语法、最短服务上下文、核心单位/结果规则及候选样本；均未完成运行验证。第十五轮§11补在线用途优先级、S6c组合/号码、显式Address family及MSCC扩展条件。第十六轮§12补CR证据、有限设备/地址子集和未决语义报告。第十七轮§13补29.061/位置/时区/地址依赖及标准pair。第十八轮§14补ECUR/IEC、RDR/ALR及六模板源码证据。本文是设计契约，不是完整字典、逐 AVP 审计完成报告或已通过的测试。只面向 Go 1.26 的公开、无持久化库。

[总体设计](diameterkit-design.md) 继续定义架构；本文集中定义规则来源和解释边界，不增加运行时规则服务、持久化或业务会话引擎。四流程使用现有 Request/Submit、Inbound/Respond、CompletionPort；[运行时契约](diameterkit-runtime-contracts.md) 和 [Peer/标识契约](diameterkit-state-consistency.md) 继续负责并发和资源。

## 1. 固定支持线与证据状态

### 1.1 主规范及已选直接依赖

选择固定的 Rel-18 支持线，不追随 Latest。以下版本选择已明确；“条款核对”只覆盖本文引用的部分，不表示整本规范或全部依赖已审核。版本选择、取得文件及哈希、规则审计、fixture/互操作是四个不同状态。

| 文档 | 固定版本 / 直接来源 | 使用范围与本轮证据 |
| --- | --- | --- |
| 基础 Diameter | [RFC 6733](https://www.rfc-editor.org/rfc/rfc6733.html)、[RFC 3539](https://www.rfc-editor.org/rfc/rfc3539.html) | 头/AVP/错误/Peer/Watchdog；已核对本文涉及条款，更新和勘误仍须逐项登记 |
| Sh 线路 | [TS 29.329 V18.0.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129329/18.00.00_60/ts_129329v180000p.pdf) | §6.1 命令、§6.2 结果、§6.3 AVP；已核对 |
| Sh 程序/数据 | [TS 29.328 V18.1.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129328/18.01.00_60/ts_129328v180100p.pdf) | §6.1.3/6.1.4、§7.6/7.16；已核对一次通知和短信数据引用，第十二轮已取得同版本 XSD，主短信路径和依赖已核对 |
| S6c、SGd/Gdd | [TS 29.338 V18.2.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf) | §4、§5.2/5.3、§6.2/6.3；已核对命令及本文所列差异 |
| 共用特性/标识定义 | [TS 29.229 V18.0.0](https://www.etsi.org/deliver/etsi_TS/129200_129299/129229/18.00.00_60/ts_129229v180000p.pdf) | 只导入被应用引用的 AVP/特性规则，不顺带声明 Cx/Dx 支持 |
| 复用短信/EPS 字段与限定 S6a | [TS 29.272 V18.3.0](https://www.etsi.org/deliver/etsi_ts/129200_129299/129272/18.03.00_60/ts_129272v180300p.pdf) | 版本已确认；复用字段逐项审计及 ULR/AIR 模拟器兼容仍未完成 |
| Ro（首版） | [TS 32.299 V18.0.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132299/18.00.00_60/ts_132299v180000p.pdf) | §6.3/6.4、§6.5.10、§7.1/7.2；本轮确认其引用 RFC 4006，不能默换为 RFC 8506 |
| SMS charging（首版） | [TS 32.274 V18.0.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132274/18.00.00_60/ts_132274v180000p.pdf) | 第十二轮取得原件并审查 §6.2/6.3 的 SMS 字段、节点角色和 category；不代表全本审核 |
| Credit-Control | [RFC 4006](https://www.rfc-editor.org/rfc/rfc4006.html) 与 [RFC 8506](https://www.rfc-editor.org/rfc/rfc8506.html) 分别标识 | 4006 为所选 3GPP 线的引用基线；8506 是取代它的现代 DCCA 规范，独立记录差异，不静默合并 |
| 字段编码引用 | [TS 29.002 V18.0.0](https://www.etsi.org/deliver/etsi_ts/129000_129099/129002/18.00.00_60/ts_129002v180000p.pdf)、[TS 23.040 V18.0.0](https://www.etsi.org/deliver/etsi_ts/123000_123099/123040/18.00.00_60/ts_123040v180000p.pdf) | 版本已确认；仅用于所需 TBCD/诊断/短信字段定义，不引入 MAP 栈或 TPDU 业务引擎 |

取得状态：第十二轮已下载 13 份主规范/直接依赖 PDF、Sh ZIP 及 XSD 成员并计算真实 SHA-256，见 [原件清单](diameterkit-mainflow-audit.md)。第十三轮再取得 TS 23.003 V18.5.0、23.204 V18.0.0、29.337 V18.0.0、29.368 V18.0.0 的 4 份 PDF 并记录哈希。第十四轮再取得23.682 V18.0.0及3份RFC文本，第十五轮再取得32.251 V18.0.0、32.260 V18.3.0，第十六轮增加3份对照PDF及2份CR ZIP/成员，累计23份PDF、Sh ZIP/XSD、3份RFC TXT及4份CR材料有哈希；对照版本不自动升级固定Profile。第十七轮再取得7份PDF及RFC7155，累计30份PDF与附件/文本/CR共40项哈希；选定编码闭包见主流程§13。第十八轮另外固定六模板及编码链14项本地文件哈希；这些是兼容源码证据，不算新增规范原件。递归规则与独立fixture仍未全部完成。

### 1.2 依赖与勘误怎样锁定

每条启用规则必须指向 `文档版本 + 条款/表/注 + 使用路径`。导入 AVP 要同时保留定义来源和当前应用的覆盖规则，例如复用 AVP 的 M 位可被当前应用规范覆盖。不能只用 AVP 名称、同属 Rel-18 或某个旧 XML 字典作为兼容证明。

依赖闭包按**实际使用的定义**展开，不为了一个电话号码把整套 MAP/短信程序纳入库。第十二轮固定 TS 32.240 V18.6.0、29.336 V18.2.0、29.173 V18.0.0、32.270 V18.2.0，并确认实际 Sh XSD 无 import/include/redefine。DRMP/OC/Load 补入 RFC 7944/7683/8583。深层地理/地址/PS/IMS/Device Trigger 的实际依赖及全部勘误仍需闭合；主流程审计 §9 已明确 SMSMI 五请求条件、HSS-ID 格式/LMSI 长度、Ro 端口/引用号/优先级与 T4 节点组差异；第十四轮§10补身份语法边界、PS/IMS设备表示差异、服务上下文与核心额度规则；第十五轮已区分SMS顶层/嵌套在线用途；真实IMEISV16子集已明确；§13已闭合RAT/LTE位置/时区及接口/PLMN地址、Furnish/Service-Specific叶；完整国际化、仅IMEI映射及其余深层字段仍待补。附表登记到具体字段后再冻结对应功能。不支持项必须在能力矩阵出现，不能悄悄删去 A/B 兼容义务。

勘误采用显式清单：编号、状态/核对日期、影响 RuleID、采用或不适用理由及回归 fixture。未审查不等于“无勘误”；不自动拉入后续 CR 或所有 Verified errata。第十二轮已记录 6733/4006 的部分勘误裁决及三项规范冲突，未声称完整勘误审计完成。运行时不联网拉规范或升级规则。

2026-11-04重点复核新增32.274 V19.1.0对照，并重读32.299 V19.0.0；原件哈希及具体条款见[主流程§31.1–3](diameterkit-mainflow-audit.md)。固定Rel-18不变：在线表第三列是MO-SMS via T4，不能按投递报告推广3409；3408/3409跨表冲突、正常CCA实验5011的基础配对均未得到直接裁决。类型已定义、局部用途许可、具名对端接受和标准完整支持须分别记录；不通过新增宽容开关消除这些区别。

### 1.3 Ro 的 4006 / 8506 选择（首版依据）

固定 Ro 线以 TS 32.299 的应用覆盖为入口，其 §6.4 明确引用 RFC 4006。公共 `creditcontrol` 包保留独立 RFC 8506 目标；不能因旧引用就把新 DCCA 的要求全部排除，也不能用更新 RFC 覆盖 3GPP 的差异。两种解释若在相同 App-ID 4 上不相容，不能装在同一 Peer binding 中。

差异审核至少包括：命令/Grouped 增量、Service-Context/扩展 rating 输入、特性与错误分支、基础 RFC 更新、failover/事件失败行为。R33的固定OCS活请求重传是受限库能力；RO-06按运行时§1.7逐命令/具名兼容扩展恢复规则，两者均不代表已实现任一完整DCCA客户端。未审完差异前，文档可以固定版本及接口设计，不能发布“Ro 完全符合 RFC 8506”的声明。

## 2. 命令与角色清单

本表锁定需要展开规则的命令集合，所有请求/应答两个方向都进入 codec 验收；库 endpoint 只允许绑定角色实际使用的方向。应用名不能替代 App-ID/Command/R 位。基本控制、E 位错误按独立基础语法处理。

| Profile / App-ID | 命令对与 Code | 发起侧 → 应答侧；来源 |
| --- | --- | --- |
| Sh / 16777217 | UDR/UDA 306；PUR/PUA 307；SNR/SNA 308 | AS → HSS；29.329 §6.1.1–6 |
| Sh / 16777217 | PNR/PNA 309 | HSS → AS；29.329 §6.1.7–8 |
| S6c / 16777312 | SRR/SRA 8388647 | SMS-GMSC/SMS Router → HSS/SMS Router；29.338 §5.3.2.3–4，具体允许角色按 §5.2 绑定 |
| S6c / 16777312 | ALR/ALA 8388648 | HSS → SMS-IWMSC；MME/SGSN → SMS-GMSC 的相应 alert 路径；§5.3.2.5–6/§6.2.3 |
| S6c / 16777312 | RDR/RDA 8388649 | SMS-GMSC/IP-SM-GW → HSS；§5.3.2.7–8 |
| SGd/Gdd / 16777313 | OFR/OFA 8388645；TFR/TFA 8388646 | MO：MME/SGSN → SMS-IWMSC；MT：SMS-GMSC → MME/SGSN；§6.3.2.3–6。T4 方向另有规范定义，不自动变成短信部署的角色 |
| Ro / 4 | CCR/CCA 272；ASR/ASA 274 | CCR为CTF→OCS；ASR反向交给应用。ASA 2001表达已终止，5002表示会话不活跃，5012表示其他未停止原因；外部状态由应用负责，库不独立核验；32.299 §6.4、RFC6733 §8.5 |
| 具名旧兼容 | SGd App-ID 下 ALR/ALA；SCAP1；限定 S6a ULR/ULA、AIR/AIA | 保留原覆盖清单；分别审计绑定/字段/样本。SCAP1 不冒用标准 Ro 解释，S6a 不扩大为完整产品支持 |

SGd/Gdd 标准 alert 使用 S6c App-ID，沿用单独绑定。DRA 是邻接，不改变本表终端角色；未来 ip-sm-gw 与 sms-fw 直连须先确定实际接口/角色，不能仅因能建立连接就开放所有方向。

## 3. 一个规则模型，分别生成发送、接收和比较行为

### 3.1 规则行的最小内容

规则清单是一份构建期数据，生成多个检查点，不是公开插件框架。每个叶字段、Grouped、命令和跨字段条件各有稳定 RuleID；一行至少包含：

| 列 | 约束 |
| --- | --- |
| 来源/覆盖 | 固定版本、条款/表/注；标准、应用覆盖、命名兼容或本库策略不得混写 |
| 位置/类型 | App/Command/方向/角色、Vendor/Code 和完整父路径；基础类型与使用位置分开 |
| 出现规则 | fixed-position、min/max、必需/禁止/条件；M 位不代替 cardinality；规范无上限与本地资源上限分开 |
| 条件输入 | 当前报文、原请求快照、binding、具名外部协议输入之一；不允许隐含数据库查询 |
| Send / Receive | 本地生成约束与对端接收动作分别列出；拒绝、忽略语义、保留未解释值不是同一结果 |
| 错误/比较 | 本地 FieldError 或对应线路失败分支；Failed-AVP 路径；Canonical 规则和原始证据边界 |
| 空间/验收 | Wire/Decode/Context/Canonical/Diagnostic 上界；正例、反例、缺输入、换序/位差异的 fixture ID 与状态 |

`*` 不等于无限内存；由 MaxAVPs/MaxDepth/MaxFieldBytes/MaxRepeated 等配置收紧。合法但超过本地上限记录 ResourceLimit，不伪造为 AVP 数量违反规范。复杂条件只允许有界生成检查；代码尚未实现时 fixture 状态保持 Planned。

Ro私有字段的本轮固定定义见[主流程§34](diameterkit-mainflow-audit.md)与[Specification §3.7](diameterkit-specification.md)：发送0/1292、10425/23及SCAP1字段有源码定义；PSP接收三层Vendor和token接收Vendor不能由发送模板推导。缺定义的具名能力在构造期拒绝启用，不注册问号/通配Vendor。逐叶Presence与解释状态分开，旧operation_result初值1、费用0及退款常量都不是协议缺省。

`smsfw-ro-legacy420-v1`是显式非标准的路径例外：仅固定`CCR/0/456/0/437/0/420`为Grouped，移除该位置标准CC-Time定义，保留413金额及417单位；不是后加载字典覆盖，也不是按Requested-Action切换。生成检查先选择具名规则集，再验证每个命令/R/E/父路径键唯一；其他父路径/CCA不继承例外。实际槽映射、接收定义和目标接受证据仍必需。构造器区分模板字面量和应用传入值，不引入S1 Getter解释器或按金额推导业务动作。

### 3.2 首批通用规则及修正

下表是本项目的检查/测试设计。规范依据为 [RFC 6733 §3–4、§7](https://www.rfc-editor.org/rfc/rfc6733.html#section-3) 及各应用列出的覆盖；不是把这些要求归因于参考库。

| RuleID | 发送 / 接收决定 | 对比较、错误和预算的影响 |
| --- | --- | --- |
| BASE-FRAME | 长度、V 位头长度、嵌套边界及整数宽度先检查；发送 padding 清零，接收忽略 padding 的内容但仍检查边界 | 消息/传输帧边界不可信时关闭，不试图重同步；帧已完整可信而AVP长度错误，可在可信包络和Reject预算内构造5014。279内引述坏长度按主流程§25.3保留，不据其声明分配/跳读 |
| BASE-FLAGS | 本地头保留位清零；接收忽略规定无语义的位；Answer 清 R/T、P 按原请求，E 选择错误语法 | WireEvidence 可保留原位，Canonical 不比较被忽略位；不将所有 flags 整字节参与比较 |
| BASE-MBIT | 端点不认识且 M=1 的 AVP 按规则拒绝；M=0 未知项允许保留但不执行业务语义；可识别不等于所有命令位置都允许 | 未知 Grouped 不猜类型；已知 Grouped 的内层仍逐条处理；Failed-AVP 是有界诊断例外 |
| APP-MBIT | Sh/短信规范指定范围内，发送按当前应用表设置 M；接收已理解 AVP 时忽略 M 与表不一致 | 对这些已知路径 Canonical 屏蔽 M 差异；不扩展到未知 AVP、任意扩展或所有 Ro 字段。依据 29.329 §6.3 表注、29.338 §5.3.3.1/§6.3.3.1 表注 |
| SMS-VSAI | 29.338 §5.3.2.2/§6.3.2.2：标准应用消息发送不带 Vendor-Specific-Application-Id（VSAI）；收到时忽略其语义 | 顶层该项不进入 Canonical；仍检查其外层 framing/空间，不把它当能力授权。CER/CEA 的 VSAI 广告、Sh 必需 VSAI 不受此规则影响 |
| APP-BITMASK | 对固定表明确要求接收忽略的未定义 payload 位，发送拒绝显式非法值，接收取有效位；如 RDR-Flags | Canonical 使用相同已审计 mask；不能把所有 Unsigned32 或未知枚举都按此处理 |
| BASE-ORDER | 固定位置先验证，其余允许位置按语法接收；标准命令顶层不同已知键按 IndependentKeys 比较 | 同键重复仍保留原序，未知项是排序屏障；Grouped 默认 Sequence，只有单独审核才变更；保守冲突边界在支持矩阵公开 |
| BASE-TIME | 线路 Time 保存原 32 位；转换到 time.Time 必须显式给 era，反向转换检查范围，不取运行机器当前年份猜测 | 规范要求时间比较时缺 era 返回本地 ContextRequired；通过纯转换参数提供 era，不新增通用时钟 facts。必须验证 2036 回绕及扩展到 2104 的支持，只有原始 uint32 保留不等于完成规范要求；Canonical 比较原字节 |
| BASE-ENUM | 枚举的未知值按该使用规则处理；可忽略值保留 raw/Unsupported，不解释成零值成功 | framing 与整数宽度始终检查；M 位/特性规则要求拒绝时走具名失败；新值未审核不能绕过条件检查 |
| BASE-ERROR | 未知 App/Command 用基础错误入口；有效 E 位答案不要求正常成功字段，但须回原请求Session且必需基础268；RFC6733基础5xxx可用E；无效答案不再生成 Answer | Report 区分 Wire、Exchange、Application；ProtocolFailure 不是任意 Result-Code 注入入口 |

发送归一化与接收证据分开：生成的已知字段由 encoder 写规定 flags，不能把收到的旧 flags 原样塞回标准发送路径。需要原样取证使用 owned 原始帧；正常重编码不承诺字节还原。被规范忽略的输入只供诊断，不能再参与本库结果解释、业务条件门或去重决定；调用方据 raw 诊断自行改变业务结论不在相同请求合并保证内。

### 3.3 错误分类的确定映射

| 情况 | 请求接收侧 / 本地发送侧 |
| --- | --- |
| 未支持应用/命令、已知支持命令的未知 M=1 AVP | 分别为 3007、3001、5001；先满足可信 envelope 和预算条件；本地发送在 admission 前失败 |
| 必需 AVP 缺失、明令不允许的 AVP、超过协议最大次数 | 分别为 5005、5008、5009；区分 max=0、有限 max 与本地上限；Failed-AVP 构造缺失/违规项的有限诊断 |
| 值非法、AVP 长度非法 | 5004、5014 只在安全识别字段/可构造错误时使用；损坏 framing 或无法保留必要 envelope 则关闭 |
| 规范未裁决/实现语义未支持 | Application报告含RuleID与检查范围，不是ContextRequired；不猜线路错误，不能由调用方补任意facts放行。发送helper在接纳前拒绝受影响分支；答案Wire/Exchange通过即可结束网络等待 |
| 缺调用方协议上下文/错误上下文 | 本地 ContextRequired/ContextConflict；不报告对端 Missing-AVP，不虚构 3GPP 业务失败 |
| 订阅不存在、短信投递失败、余额不足、会话未找到 | 应用选择合法协议结果；库校验分支，不查询或认证业务状态 |

结果码字段本身依 Profile：Sh 使用其应用定义的 Result/Experimental 分支；所选 Ro CCA 语法含必需 Result-Code 和可选 Experimental-Result，不能套全局 XOR。特性失败等组合还须结合 32.299 §6.5.10/§7.1.8A 逐分支核对。第十三轮主流程 §9.4 固定安全处理：正常 CCA 缺必需基础结果是无效应答；2001 与实验 5011 并存不能解释为成功；标准自动生成配对仍未决，不能擅自选择一个结果覆盖另一个。

## 4. 应用规则切片与协议输入（三应用首版，Ro首版）

这是一组可审查的规则和验收切片，**不是只实现这些字段就叫完整 Profile**。通用字段、所有命令 ABNF、复用 Grouped 的递归叶字段及旧模板的逐项展开仍按 §7 收尾。

| RuleID / 源条款 | 约束和最小反例 | 输入 / 检查点 / Planned fixture |
| --- | --- | --- |
| SH-SNR-ONCE；29.328 §6.1.3 表/§7.6.17/§7.16 | UE reachability for IP 的订阅有一次通知条件；Send-Data-Indication 对该一次订阅不适用；不能只按 SNR ABNF 的可选项处理 | 由 SNR Data-Reference/字段即可判断；Build/Wire；缺一次通知、错用发送数据、正常 SNR 三例 |
| SH-PNR-BODY；29.329 §6.1.7–8 | PNR 的必需载荷/身份与 PNA 结果、原请求关联受检；订阅 Found/NotFound 不成为 Ready 条件 | 当前 PNR/原请求和 binding；Wire/Answer；无数据库也可生成合法 PNA，缺 User-Data 仍拒绝 |
| SH-PUR-FEATURE；29.329 §6.1.3–4 注 | PUR 多 Data-Reference 需要双方 Update-Eff-Enhance 支持；单引用不因无对端特性证据被拒绝 | 本地固定特性+目标 HSS 的具名证据；发送前/应答解释；Unknown、错误 HSS、DRA 广告、支持双方分别验收 |
| SH-DATA-SCOPE；29.328 §7.6.16/附录数据模型 | SMSRegistrationInfo/UE reachability 的数据引用与 XML 子模型分开；不得把 well-formed XML 当完整数据语义通过 | shdata 纯工具，原字节保留；XSD/嵌套 Extension 已审查，注册/注销/未知扩展的独立样本仍待补，不进入默认 Diameter reader |
| SMS-SESSION；29.338 §4.4–4.7 | S6c/SGd 采用其隐式结束会话约束；禁止把它们套进授权/计费会话引擎；控制广告与应用 VSAI 规则分开 | binding+当前消息；Build/Wire；Auth-Session-State、禁用会话计时字段和广告分别测试 |
| S6C-SRA-USE；29.338 §5.2.1 | SRA 身份/节点信息按原 SRR 用途及结果解释；非投递查询可能无 Serving-Node，User-Name 不能统一当普通 IMSI | 原 SRR 快照及角色；Exchange/纯 Interpret；无请求离线解释返回缺上下文，不猜用途 |
| S6C-NODE；29.338 §5.3.3.6–7 | 保存 primary/additional 和组合节点，不能泛化成“多个节点字段互斥”；未知 SMSF/SBI 能力不转成伪 SGd 地址 | 当前字段+固定 Profile；Wire/Interpret；组合节点、无 realm、附加节点与未支持候选 |
| SGD-OFR-REPLY；29.338 §6.3.2.3–4/§6.3.3.5–7 | OFR 的 SC-Address、User-Identifier、SM-RP-UI 及 OFA 结果/失败诊断按语法；TPDU 不交 Diameter 库解析 | 消息/原请求；Wire/Answer；延迟答复、字节改变冲突、原连接断开但另一 Arrival 可答 |
| SGD-TFR-TIME；29.338 §6.3.2.5/§6.3.3 | TFR 的目标和时间字段按其固定类型/条件检查；不能把单位为秒的 timer 与绝对 Time 混为一类 | 当前消息/显式 era（仅绝对时间解释）；Build/Wire；长度、范围、回绕和原值保留 |
| RO-CCR-KIND；RFC 4006 §8.2/8.3/8.41，32.299 §6.4 | IEC EVENT 的 Requested-Action 必须表达，事件/INITIAL/TERMINATION 分支分开；重传不改变类型/编号 | 当前 CCR；Build/Wire；缺 action、错分支、原身份重传。跨 exchange 编号演进由应用负责 |
| RO-CCA-CORRELATE；32.299 §6.4.3，RFC 4006 §3.2 | 正常 CCA 关联 Session/CC 类型/编号及服务，顶层结果与 MSCC 结果分开；E 位错误另用基础语法 | runtime 原请求快照；Exchange/Interpret；顶层成功但服务失败、错编号、有效通用错误 |
| RO-FEATURES；32.299 §6.5.10 | 发起请求使用的服务特性与会话建立时共同特性不同；只声明本次需要的特性，不复制 DRA CER；不强制初次请求先拿未来 CCA | 初始由 body/local binding；后续依赖已建立协商时由应用提供具名 SessionFeatures；目标/Profile/Session/Service-Context 关联受检 |
| RO-SMS-FIELDS；32.299 §7.2.192/§7.2.211、32.274 §6.2/6.3及主流程§11 | 无冲突的在线用途先于OM/OC可选性；设备顶层与PS嵌套分开。SMS-Result/Sequence的CR/表间冲突按主流程§12.2报告未裁决，不凭表优先自动判对端违规；真实IMEISV16子集见§12.3。SMS、MMS旧模板字段保留义务不变；标准字段、旧对端附加字段和私有 refund 逐项归类 | 纯构造/解释；金额用整数/指数，opaque token 不转字符串；六模板的字段差异已登记；最终 bytes、私有规范和完整条件仍是 P1 |

### 4.1 收敛后的 Context 入口与包依赖

现行首版同时实施sh.PeerFeatures、ro.SessionFeatures及通用Context边界；FieldSelection仍是ro构造参数。撤销这里残留的第十九轮“不要求首版包依赖”表述，不使用占位类型；当前/原请求内已有协议输入仍须检查。

`validation.Context` 作为已有 RequestOptions.Validation / Inbound.Check 的不可变有界值保留，零值表示没有额外调用方输入；不表示跳过规则。根包和 codec 依赖 validation，应用包依赖 pdu/schema/validation，validation 不导入 app 或根包。应用生成的具名构造器创建 Context；内部表示由生成器管理，不公开 map、any、FactKind 任意注册或按字符串设置字段的入口。

Go包可见性按[架构§4](diameterkit-architecture.md)实现：validation.Context别名指向internal/contextvalue封闭Value；sh/ro具名工厂在同仓库调用其有限构造入口，外部代码不能任意注册facts。生成规则通过schema的中立操作契约接入，不导入codec/root；规则定义仍只有本页及spec锁定来源一份。

Role/Direction 来自 binding，RequestEvidence 来自 runtime 拥有的原请求，UseCase 能由报文推出时直接推导；这些不是可供用户覆盖的公开字段。离线纯解释器明确接收原请求/构造事件作为参数。原先“调用方填写整份 Role/Direction/RequestEvidence”的理解撤回。

离线 codec.ValidateContext(msg, ctx) 没有 endpoint 的隐式 binding/原请求；遇到依赖它们的规则时必须在 Report 中列出 MissingContext，不能空跑后报 Complete。完整离线解释通过对应应用的纯构造/解释器传入原请求和明确角色，内部复用规则。新增上下文构造器不能伪造已接纳 exchange 的请求证据。

只为已经有规范用例的输入给出候选签名，函数体尚未实现：

```go
sh.ValidationContext(in sh.PeerFeatures) (validation.Context, error)
ro.ValidationContext(in ro.SessionFeatures) (validation.Context, error)
```

Sh PeerFeatures 只记录目标终端身份、Profile、已知/未知状态及有界特性集合；本地特性来自 binding。Ro SessionFeatures 另绑定 Session-Id 和 Service-Context-Id，并表达建立时双方协商的集合。二者不携带 DB revision、TTL、余额、订阅记录或 owner；它们是调用方声明的协议输入，不是库验证过的业务事实或端到端身份认证。错误目标/会话/类型返回本地 ContextConflict，特性不足按实际 RuleID 报缺失/不支持。

Ro的服务器支持报价与已建立会话的共同集合必须区分：失败答案中的Supported-Features可以保留，但不据此构造已建立的SessionFeatures。共同集合来自成功建立时双方支持集的交集；每次请求只声明本次实际需要的特性，不机械发送全部共同集合。缺报价/未识别特性不等于全部支持；独立IEC事件不沿用其他Session的协商，DRA重连也不能隐式扩充集合。未知输入和有效空集分别报告；沿用现有具名输入，不增加缓存或第三种Context，见[主流程§30.3](diameterkit-mainflow-audit.md)。

需要预先证明目标特性的 Sh 多引用请求应绑定明确 HSS；若 realm-only 的候选终端不具有同一已声明特性边界，就不能拿某次 PUA 证明所有候选支持。单引用基础调用不受此附加限制。Ro SessionFeatures 只在本次确实使用相应会话特性时必需，不强制所有 ECUR 或普通 IEC 先查询外部状态。

第十二轮新增的 Ro SMS FieldSelection 仅是纯构造参数，表达已查明的 OM/OC 配置；不是第三种 validation.Context，也不是远端事实。32.299 §7.1.12的最短Service-Context-Id使OM/OC可选，不能把本地默认选择变为接收必需项；带运营方形式而缺规则时报告检查未完成，细节见主流程§10.3。对于 S6c 路由解释、SGd OFR/OFA、普通 Sh PNR/PNA 和不依赖会话协商的 IEC，不增加一套空的 Context factory。低层直接调用与薄客户端都使用同一规则；信息能从当前报文/原请求获得时不要求调用方重复提交。未知的将来 CCA 不作为首次 CCR 的发送前置条件。

Check 在仍缺输入时可补齐；第一次成功后冻结快照，再传不同输入报冲突。Respond 从原请求、冻结输入和本次应答分支即可校验；本轮没有证据要求新增 RespondWithFacts。特性缺失只影响依赖它的规则，不阻止不使用该特性的消息。

### 4.2 有界表示和比较

MaxContextBytes 包含身份字符串、Profile/Session/Service-Context 绑定及特性条目；规则给 MaxFeatureLists/条目宽度和重复键处理。复制前计量，重复且冲突的特性键报本地输入错误，不靠 map 后写覆盖。出站记 Q/M，入站记 I，重传复用原快照，缓存只保存已提交结果需要的有限数据；无后台刷新、全局特性缓存或持久历史。

规范化顺序固定为：安全 framing → 应用 Receive 规则 → 有效语义视图 → Canonical。明确忽略的顶层 VSAI、已知 M 差异和应用 bitmask 掩码由生成的有限规则决定，不能增加公共 ignore list。对未知 M=0 内容仍保守按 raw 比较，避免将未来扩展造成的业务差异吞掉；这是已公开的保守重复策略。

忽略路径、有效位掩码和顺序规则都进入 Profile/规范化指纹；修正规则后不同指纹不能共享去重记录。本文修改尚未实现的 CanonicalV1 草案，不宣称已有持久记录可迁移。对完全忽略的 VSAI 仍计入接收字节/AVP 数与诊断上限，不能因不参与比较而绕过容量检查。

## 5. 代表流程的文档推演（四流程首版）

| 流程 | 最小调用链及库输入 | 反例与职责边界 |
| --- | --- | --- |
| Sh 订阅/通知 | 构造 UE reachability SNR → Request/Submit → 解释 SNA；另一路 Peer.Receive PNR → 应用选择结果 → Respond PNA | PNR 是新请求，不伪装成 SNR 的迟到答案；不要求存在原进程的 SNR/订阅表。一次通知字段受检，实际订阅删除/消费由 HSS/AS 应用负责 |
| S6c 路由查询 | NewSRR 从 RoutingQuery 形成消息 → Route.Request → InterpretSRA 使用保存的原查询；低层调用使用 runtime 原 SRR 快照 | 不为低层用户增加重复的 QueryPurpose facts。缺投递节点不一律错；用户决定 MME/MSC/SMSF 路径，库不查 HLR、DNS 或投递策略 |
| SGd 延迟 OFA | Receive OFR（无待补协议输入则 Ready）→ 应用有限队列处理 → 构造 OFA → Respond | 缺数据库或消息尚未投递都不是缺协议上下文；应用选结果。断线不丢逻辑 Inbound，副本只有实际 Arrival 才可答；超时/关闭沿用原提交门 |
| Ro 固定 OCS 重传 | PrepareSMSDebit → ExplicitRetransmit Route.Submit → 真实传输失败后 RequestInfo/Retransmit → 原 CompletionPort → InterpretSMSDebit | 默认 IEC 无 SessionFeatures 前置查询；使用特性/ECUR 后续时只补相应协议输入。Tx 单独到期不授权重传，原 CCA 可赢完成；库不确认扣费、决定退款或管理会话恢复 |

上述是接口和反例推演，不是四条流程已经通过互操作。最小路径不需要数据库、持久化回执、Stage、全局 facts 服务；特殊规则所需的具名参数也不能被“无持久化”口号掩盖。

## 6. 参考其他库的取舍

[OTP diameter_dict](https://www.erlang.org/doc/apps/diameter/diameter_dict.html) 将生成字典、消息/Grouped 结构及显式继承分开，适合参考本库的单一规则来源；其 [strict_mbit 配置](https://www.erlang.org/doc/apps/diameter/diameter.html) 说明接收策略应有明确语义，本库选择逐 Profile 规则而非全局关闭 M 检查。

[freeDiameter messages.c，固定提交](https://github.com/freeDiameter/freeDiameter/blob/9d4bc47584bfeced9de708fe00f5629ac4689db5/libfdproto/messages.c) 分开检查未知必需 AVP、次数及固定位置；本库参考检查维度，不复制它的诊断对象或把字典校验视为全部 3GPP 程序校验。源代码也不能代替实际版本的表注、字段条件及本项目空间证明。

## 7. 实施基线与遗留项归属

2026-11-03完成P0文档收尾；Profile按现有固定版本、已采纳勘误及所选子集进入实施。这里固定设计输入，不宣称Profile符合性或运行验收已通过。

| 记录 | 现行归属 | 不再作为前置文档门槛的内容 |
| --- | --- | --- |
| R24/R31 | P1/P2/P4/P5：候选签名、具名输入、分阶段校验及四应用流程测试 | 不再增加泛化facts或为已明确零输入路径补框架 |
| R29/R34 | P1/P2/P4/P5实现与P6/P7验证；受限能力按Specification §3.1 | 不要求穷尽可选深层字段、439/439或全部候选独立认证才开始实现；不能删掉实际首版义务 |
| R03/R17–R21/R32/R36及R33通用部分 | P3实施；绑定、Canonical、生命周期、发号、预算和Linux SCTP验收 | 不把已定行为重复列作“尚待设计”；Ro专用切片纳入首版P5 |
| 具名兼容/SMEA/实际角色 | P4/P6首版对应能力验收 | 固定范围不扩展；独立证据不提前写作Passed，也不无故延期必需兼容 |
| Ro及实际计费兼容 | 首版P5及P6/P7 | 固定操作/结果、六模板/SCAP1、R33/RO-06按既有义务实现和验收，不引入业务存储框架 |
| 其他未承诺能力 | 后续版本 | 限定S6a、完整SCUR/PRICE_ENQUIRY和已登记未启用深层功能保留，不因Ro纳入自动扩大 |

唯一队列见[Specification §7](diameterkit-specification.md)，退出条件见[总设计§14](diameterkit-design.md)。不再要求另一轮字段/CR穷举作为实施前置；相等时间、无IMSI普通SRA及Realm争议保持已有受限行为。当前仍只授权文档；以后切换实施直接从P1/P2开始。

首版Ro的2026-11-04收敛见[主流程§28](diameterkit-mainflow-audit.md)：通用DCCA、32.299凭据、32.274失败SMS和旧OCS分别声明；固定4006子集不自动升级8506。5011接收按既定矩阵执行，标准自动生成仍受限；不能套Sh结果XOR或任意补基础码。余额成功条件须结合原CHECK_BALANCE动作，不从字典全局可选推导可缺结果。仍只使用既有Profile/binding/校验层和验收ID，不新增兼容开关、facts或存储框架；五包验收固定在[覆盖§35](diameterkit-usage-coverage.md)。

以下为已采纳细化和历史快照；规范规则按固定范围继续生效，旧“下一门槛”不构成新增的前置工作。

第十八轮补充：[主流程§14](diameterkit-mainflow-audit.md)的RO-ECUR-FINAL/TERM-RESULT、RO-IEC-DEBIT/REFUND及S6C-RDR/ALR各RuleID进一步细化现有规则；不新增Context、规则插件或业务状态机。TERMINATION成功不要求GSU；refund结构候选受DECISION-01限制；RDR的cause条件超出Grouped全可选语法，ALA通用E错误与正常应用结果分别校验。六模板在独立解释域保留既有兼容义务，但不允许标准420改成Grouped、1292冒充2022或SCAP1头App0被当成普通Ro。验收见PAIR-03/04、TEMPLATE-02。


第二十轮补充：[主流程§16](diameterkit-mainflow-audit.md)与[候选§15–17](diameterkit-protocol-vectors.md)固定Sh普通短信单DataRef、完整四组结果和XML使用位置，及MME/SGSN Alert、Gdd TFR角色条件。Schema合法、选定应用语义和业务事实分别报告；XML helper是显式有界解析，不成为默认Diameter decoder的隐式步骤。Sh正常答案有VSAI，S6c/SGd标准发送无VSAI；两套规则不能合并。ALR移动节点未知不是通用ContextRequired，更不要求读MWD。PAIR-05/06、XML-02、FIELD-13仍待实现验收。


第二十一轮补充：[主流程§17](diameterkit-mainflow-audit.md)固定23.008依赖及SH-NUMBER-E164/SH-SC-TEXT/SMS-DUMMY-MSISDN、SH-ACCESS-XOR/PRIVATE-KEY/UNSUBSCRIBE/SUBSCRIBE-DATA/NOTIF-EFF与逐命令特殊身份条件。OR访问键与SMS多身份组不同；Sh特性仍用既有PeerFeatures，不引入订阅状态Context。标准/有限helper/raw/具名legacy-sgd-alert-v1分别报告，既有样本Code错误要修正而非升格为兼容。原source记录累计41项规范及14项旧源码；Ro延期不影响首版这些规则。


第二十二轮[主流程§18](diameterkit-mainflow-audit.md)细化SMS-SRA-ID、S6C-NODE-REALM与SH-MIXED-SUB。前者冻结“无依据不造成功身份”的本地行为，并保留无IMSI标准分支缺口；Realm默认标准域不变，设计具名s6c-ip-sm-gw-realm-v1仅覆盖S6c SRA/Serving-Node的Number+Name+Realm。其启用需既有固定binding、规则指纹及NODE-01证据，不是动态放宽未知规则。Sh支持共同IMPU、终端Notif-Eff已知且无SendData/Private的{11,25}/{12,25}/{11,12,25}；一次语义仅25，超出子集仍分别报告未支持或已证明的违规。XML/raw的Canonical规则不变，不能按解释结果删字段、去重或排序重写原值。新增IDENTITY-02、NODE-01、SUB-02，不增加Context种类。


第二十三轮新增[逐叶和结果清单](diameterkit-field-inventory.md)，将本页规则落到111行AVP叶和10项XML叶、实际A/B与设计D来源。SMS-DIAG-PRESENCE、SMS-RESULT-CLOSURE、TEXT-UTF8/ID-ASCII/ID-EXTERNAL-ASCII是既有校验层的细化，无新Context或规则引擎。Repository/iFC/DSAI的原XML保留、schema结构和完整typed语义分开声明；未见实际启用的完整helper不作为已选Sh24/25/11/12前置。结果按命令/角色/namespace，不能从结构声明/泛Parse推导完整成功。FIELD-14、RESULT-06、XML-03、TEXT-01仍待实现/独立验证。


第二十四轮[主流程§20](diameterkit-mainflow-audit.md)细化S6C-ADDITIONAL-PAIR和SGD-TFA-RESULT：单组语法与普通HSS双节点过程、LMSI外部条件、结果附加字段、Time与原请求分别校验。相等重传时间的条款边界保留报告，完整builder选严格早于；不把原请求缺失/未知VLR状态判为远端违规。Sh PUA失败定位与成功回显分开，SNA缺702的SHOULD不升级为MUST。新增136条[完整候选](diameterkit-leaf-vectors.md)不升级能力状态，也不增加Context类别/动态Profile或业务查询。

第二十五轮[主流程§21](diameterkit-mainflow-audit.md)补SH-FEATURE-TRIGGER、T4-OFA-EXTERNAL和S6C-SMEA条件，纠正Sh628固定M0与双节点请求漏3309；已理解AVP的接收M例外保持。静态反例、应用不支持和ResourceLimit分层；251新增候选不升级实现状态，不新增动态Profile或深层helper。当轮门槛记录为§21.5；现行归属见本页§7。

第二十六轮[主流程§22](diameterkit-mainflow-audit.md)及[439行矩阵](diameterkit-command-path-matrix.md)明确APP-PATH-INSTANCE：规则逐层匹配Vendor/Code、按直接父实例统计，不能由其他组/命令补必需叶；正常/E错误、应用禁位、C边界和具名兼容分别验证。共享类型规则不代表逐命令生成正确，当轮门槛为§22.4，当前增量见§23.7。

第二十七轮[主流程§23](diameterkit-mainflow-audit.md)收敛既有规则：SH-PRIVATE-KEY补Note9（External访问键不得加Private）；SMS-SRA-ID按身份用途/原请求解释，不能据相同数字判非法回显；S6C-NODE-REALM兼容只放行既定三元组且独立字段错误仍报告；SGD-TFA-RESULT对相等未决、零值存在性、缺era/原请求分别处理。S6C-SMEA数字三段依据由29.002 ASN.1注释补强，d=1..15、TOA91、恰2+ceil(d/2)、奇数高半字节F；bit0为单向必带条件，其他TON/NPI按支持范围而非笼统坏包分类。27条矩阵缺例已补，剩余门槛见§23.7；不新增Profile种类、公共API或运行时规则服务。

第二十八轮[主流程§24](diameterkit-mainflow-audit.md)细化BASE-ERROR/RESULT-ENVELOPE：正常/E语法先分流，E的可选297不得由正常XOR一律拒绝；Sh已定义实验码按29.329 §6.2禁止与268并存，固定Profile直接拒绝；未知Vendor补充结果保留范围报告，不扩标准完整builder。采用Verified4615（一个Failed-AVP父组）和4887（本地Origin-Realm），Reported6832不采用。Failed-AVP引述字段不执行应用语义，容器外字段及有界解析照常；错误来源不取得成功权限。公共叶新增53条定向覆盖，剩余门槛见§24.6，不增加动态规则/API。

第二十九轮[主流程§25](diameterkit-mainflow-audit.md)细化BASE-MBIT/BASE-ERROR：已知Grouped按父实例计数，未知M0是opaque而非猜测Grouped；未知M1在已知M0父内仍处理。279本身边界有效后，其内坏长度/短证据保留raw并停止该分支，不能按普通Grouped要求全部子项可解码；外层越界仍错误，已识别深度/节点超限归ResourceLimit。新增81输入，矩阵261定向/335在位；剩162公共+16按需。CR0069正文/批准尚未确认，Realm标准域和具名兼容不变，不增加规则开关。

第三十轮[主流程§26](diameterkit-mainflow-audit.md)补145输入：L005/L015/L009/L019/L020在各实际命令均有定向反例；矩阵319定向/389在位，剩104公共+16按需。Origin-State属于当前Origin且非回显；Route-Record保序与顶层Canonical剥离分别执行。297须逐父检查。UTF8String的有效编码和禁止U+0000同时成立；控制码的SHOULD不升级为一律拒绝，诊断容器例外与外围文本分开。剩余清单见§26.6，不增规则插件。

第三十一轮[主流程§27](diameterkit-mainflow-audit.md)补214输入，L001–L020全部适用行有定向反例，矩阵423/439，余16按需。SF引用修正为29.329 §6.3.11–13→29.229 §6.3.29–31。Sh260逐父必需项/固定App与SMS忽略规则分开；ALR HSS过程的Host约束高于单看CCF可选性。目的类型通过不等于准入通过，坏回显字段的拒绝可构造性另按§27.4，不能以行覆盖代替完整错误出口。
