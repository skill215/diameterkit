# diameterkit Library Specification

> **当前范围（2026-11-10，Draft 0.87）：** 第一阶段=首版/v1，包含Sh、S6c、SGd/Gdd、Ro/DCCA及sms-fw实际所需六模板、SCAP1和RO-06恢复能力；Go 1.26、纯Go实现（含依赖与SCTP适配器，发布路径CGO_ENABLED=0）、公开Go库、无持久化边界不变。明确撤销第十九轮“Ro后置”分期；带日期的旧分期/统计仅作历史证据，不再授予Ro跳过首版验收的例外。限定S6a、完整SCUR/PRICE_ENQUIRY等未承诺能力仍后置。范围及剩余缺口见[Specification §3/§7](diameterkit-specification.md)，顺序见[总设计§14](diameterkit-design.md)，验收见[覆盖§39](diameterkit-usage-coverage.md)。2026-11-10已完成前八十六批有界实施（Base/S6c基础、公共包络与关联、出站运行时、关闭与选举、TCP/TLS生产适配、有限Inbound/应答、有限Reject/应急预算、Canonical/有限重复缓存及多Arrival重答、固定同终端Route/多attempt与活动请求原身份重传、有限Dialer/Maintain拨号及重连、按class/编码字节DRR调度、固定Peer的有限Listener/Serve、有界标识源/Session/guard、Base/S6c有限Context/Report、S6c普通RDR/RDA及HSS ALR/ALA、SGd/Gdd普通OFR/OFA与TFR/TFA、Sh普通四组命令及透明User-Data、Ref24/25有限Sh-Data helper、Sh单订阅/退订及Ref11/12有限数据、Sh特性/混合订阅/5011、S6c普通HSS节点及SRR/SRA成对路由条件、S6c普通多节点投递报告、MME/SGSN普通提醒及角色证据、SGd/Gdd直达重投时间及显式era/应用证据、T4 MO顶层计费身份及原请求绑定、S6c T4无MSISDN报告与HSS提醒、T4已知IMSI路由查询、SRA实验失败/诊断及角色UNRI、SGd OFA实验5001及接口证据、S6c IP-SM-GW节点及Realm接收边界、普通RDR的IP-SM-GW结果、IMS RDR/RDA身份绑定、IMS ALR/ALA提醒用途、IMS SRR/SRA身份及固定HSS恢复、IMS OFR/OFA转发用途及IMS TFR/TFA目标绑定、IMS MO可用IMSI与真实MSISDN、T4报告IP-SM-GW结果、Ro标准SMS IEC直接扣费基础及无UPDATE的ECUR I/T）；已执行范围及未完成项以[Specification §6](diameterkit-specification.md#6-需求与证据的维护)和[spec进度](../spec/progress.json)为准，未列明的运行验收仍Planned。

状态：Draft 0.87，2026-11-04将Ro并入第一阶段/首版；保留2026-11-03收尾的行为基线并显式修订范围；协议分析截至2026-10-31第三十一轮。本文规定拟交付库的可观察行为与验收要求；2026-11-10前八十六批实现状态见§6.1–6.86，尚不表示稳定 Go API 或完整协议符合性认证。基准为 Go 1.26；仅提供 Go 库。

## 1. 为什么需要这份 specification

需要。已有设计已覆盖协议、并发、资源和兼容，但“调用方能依赖什么”原先散落在九份设计文档中，评审历史还记录了被撤回的方案。实现前需要统一需求编号、支持声明和关闭门槛，才能发现遗漏并判断实现是否满足约定。

本文是库级行为契约；不是新的 Diameter 线路标准，也不抄写全部 RFC/3GPP。逐字段规则仍只有一份来源，见 [Profile 规则契约](diameterkit-profile-rules.md) 和 [主流程审计](diameterkit-mainflow-audit.md)。生成字典描述字段；本 specification 还约束所有权、失败、并发、期限和资源，因此字典不能代替它。

本轮采用仓库内 Markdown，不初始化 OpenSpec、不增加工作流依赖，也不制定第二种 schema DSL。以后若采用 OpenSpec，将同一组需求迁移到能力 spec，保留 ID 和证据链接，不同时维护两份规范性正文。当前以一份文件开始，只有篇幅或独立维护需要明确时才拆分。

## 2. 术语、效力与文档分工

本文的 **MUST / MUST NOT / SHOULD / MAY** 使用 [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119.html) 与 [RFC 8174](https://www.rfc-editor.org/rfc/rfc8174.html) 的需求强度；只约束本文明确的支持范围。SHOULD 的偏离须说明原因与影响。中文说明、示例签名和候选默认值不会自动变成额外 MUST。

| 术语 | 含义 |
| --- | --- |
| Profile | 构建期固定的协议定义、应用覆盖、已采用勘误及具名兼容规则；不是运行时随意注册的解析策略 |
| Node / Peer / Connection | 库运行实例 / 稳定邻接对象 / 一代具体连接；DRA 邻接身份不等于 HSS、MME 或 OCS 身份 |
| exchange / attempt | 一次已接纳的逻辑出站请求 / 其中一次传输尝试；多个 attempt 只有一个最终完成结果 |
| Inbound / Arrival | 进程内逻辑入站请求 / 某条连接实际收到的一份合法副本；只有 Arrival 赋予该连接的应答发送权 |
| owned | 调用方可在约定交付点后独立持有的数据；后续读写、缓冲复用和关闭不得改写它 |
| 支持声明 | 固定 Profile、命令/方向/角色、功能子集、传输/平台、限制及测试证据的组合；不是单一“支持 Ro”标签 |

文档分工如下。本文已编号条目规定外部不变量；细节引用指定章节，不复制整张状态表。发现规范依据或细化契约与条目冲突时，登记受影响 ID 并阻止该能力冻结，先修正文档；不能用“最新一轮优先”掩盖冲突。

| 材料 | 责任 |
| --- | --- |
| 本文 | 需求、不变量、支持与验收门槛；不固定所有类型拼写 |
| [实施架构](diameterkit-architecture.md)、[总体设计](diameterkit-design.md)、[API 与流程](diameterkit-api-workflows.md) | 架构/包与module依赖、候选签名、调用方式和取舍 |
| [运行时契约](diameterkit-runtime-contracts.md)、[Peer/标识契约](diameterkit-state-consistency.md)、[部署契约](diameterkit-deployment-interop.md) | 对应编号需求引用的状态、时序、预算和部署限制 |
| [Profile 规则](diameterkit-profile-rules.md)、[字段审计](diameterkit-mainflow-audit.md) | 固定规范依据、逐规则行为、裁决及尚缺的字段证据 |
| [使用覆盖基线](diameterkit-usage-coverage.md) | A/B 需求来源、验收 ID 和执行证据；已有验收 ID 不另行重编号 |
| [逐叶与结果清单](diameterkit-field-inventory.md) | 实际字段/父路径/结果及Sh深层、国际化支持边界，区分调用与设计证据 |
| [候选协议样本](diameterkit-protocol-vectors.md)、[逐叶完整帧](diameterkit-leaf-vectors.md) | 手工推导的字节/文本输入、预期及来源；候选不等于独立验证完成 |
| [评审记录](diameterkit-design-review.md) | 问题、优先级和历史；折叠历史不是当前契约 |

需求 ID 是稳定引用，不是包名或新增公开类型；本文使用两位序号，例如 SPEC-RESULT-01。主流程审计原有 SPEC-RO-5011 等是规范裁决记录，继续保留原 ID，不作为本表需求。当前下列需求作为Draft 0.87实施基线；已覆盖范围为Partial，其余仍Planned，细分证据见§6.1–6.86。条文明确不等于测试通过；§7 列出的未决项不通过一条高层 MUST 自动变为已解决。

## 3. 目标支持范围与版本分期

2026-11-04按用户最新指令：**第一阶段=首版/v1，包含四应用及实际Ro兼容/恢复；并非要求工程P1模型阶段一次完成网络栈。** 撤销2026-10-31的Ro后置安排；旧A/B证据与需求ID保留，版本号/日期和实际实现状态不因范围变化而自动确定。

| 版本 | 支持线 | 必需目标与验收 |
| --- | --- | --- |
| 第一阶段/首版 | Base endpoint与公共能力 | CER/CEA、DWR/DWA、DPR/DPA、codec、owned数据、分阶段校验、Peer/标识、预算、一次完成、进程内去重；不交付DRA/Relay产品 |
| 第一阶段/首版 | Sh | UDR/UDA、PUR/PUA、SNR/SNA、PNR/PNA及已选Sh-Data；SH/XML/SUB/NUMBER/FLOW-01等原义务 |
| 第一阶段/首版 | S6c | SRR/SRA、RDR/RDA、ALR/ALA及已选身份/节点/结果；S6C/ALERT/IDENTITY/NODE等原义务 |
| 第一阶段/首版 | SGd/Gdd | OFR/OFA、TFR/TFA、延迟OFA及S6c Alert组合；旧SGd App-ID Alert具名兼容；SGD/FLOW/PAIR等原义务 |
| 第一阶段/首版 | Ro SMS / DCCA | IEC余额/扣费/退款、ECUR INITIAL/TERMINATION、ASR/ASA、分层结果及RO-06恢复；RO-01–06、REPLAY-01–04、PAIR-01/03、FLOW-04、RESULT/PROVISION/APP等 |
| 第一阶段/首版 | 计费兼容 | sms-fw全部实际A/B用法、六模板、SCAP1、四槽/私有字段与结果、失败TERM/退款、受检恢复；覆盖§37/38及TEMPLATE-01/02、COMPAT-01 |
| 第一阶段/首版 | Transport与互操作 | TCP/TLS、Linux SCTP adapter、四应用适用的DRA/端点/OCS和模拟器路径；SCTP/DRA/SIM/PEER/ID/FO/VAL/BUD等 |
| 后续版本 | 限定S6a及未承诺能力 | COMPAT-02、完整SCUR/PRICE_ENQUIRY、独立RFC8506完整支持及已登记未启用深层子集；Ro纳入不隐含这些目标 |

首版MUST完成以上首版A/B义务及公共安全/资源契约。creditcontrol/app/ro、具名计费compat、适用SessionFeatures/FieldSelection、R33与RO-06必要入口MUST纳入首版实现及API验收，MUST NOT以空包/占位方法或SingleAttempt子集标完整交付。App-ID4及SCAP1仅在明确配置相应binding时注册/广告，MUST NOT因为链接包而自动广告所有应用。库无需配置OCS才能单独使用Sh等其他能力。

v0.x MAY先发布已验证子能力，但MUST列明Ro或其他首版目标缺项，不能宣称v1完成。限定S6a等明确后置项仍为未支持而非Passed。字段按四应用实际依赖导入，不扩成Tsp/MAP/SIP/TPDU、完整IMS/PCC或计费业务引擎。验收归属见[覆盖§21/§39](diameterkit-usage-coverage.md)。

### 3.1 首版支持边界（2026-11-04修订）

§3版本表及覆盖基线§21的首版A/B义务保持。下表固定所选子集与限制，将既有Ro义务并入首版，不新增无调用依据的协议功能，不以收尾撤销实际使用字段或兼容验收。

| 范围 | 首版承诺边界 | 明确不承诺 / 后续范围 |
| --- | --- | --- |
| 公共endpoint | Base控制消息、owned codec、分阶段校验、Peer、标识、一次完成、预算、进程内去重及已审核TransportFailover | DRA/Relay产品、任意透明转发、持久化、跨进程恰好一次、业务恢复 |
| Sh | 四组命令；Sh-Data Ref24/25及11/12有限子集、已选订阅/通知/特性规则；接收PNR不依赖库内订阅表 | Repository/iFC/DSAI完整helper、全Sh数据模型、全IDNA/SASLprep；raw承载不升级语义支持 |
| S6c | 三组命令；真实IMSI普通成功、已明确Router token/MCCMNC用途、实际节点/报告/提醒、选定SMEA数字子集；HSS发ALR保留目的Host过程要求 | 普通无IMSI成功SRA、任意HSS-ID替代、未裁决Realm组合的标准完整支持 |
| SGd/Gdd | 两组命令、延迟OFA、既定角色/身份/结果、标准S6c提醒及已选IMS/T4字段规则；无需Ro即可校验相应OFA身份 | 业务重投递/提醒调度、原生T4设备触发、MAP/SIP/TPDU业务引擎 |
| 时间与Realm边界 | 重传时间标准构造严格早于原上限；相等接收保留原结果/Time和范围报告。保留既定Realm标准域及显式s6c-ip-sm-gw-realm-v1解释范围 | 不将等号或争议Realm组合宣称为已裁决标准能力；兼容不授权新节点或扩大Additional/其他命令 |
| 传输/实际兼容 | TCP/TLS、Linux SCTP、三应用DRA绑定和模拟器路径；旧SGd Alert等既定具名兼容按原归属验收 | 不把SCTP、SMEA或首版A/B兼容义务转成可选项；现场投产证据与库发布分别记录 |
| Ro及计费兼容 | §3.2–3.6与覆盖§37/38的实际用法、六模板/SCAP1和受检恢复 | 不宣称完整DCCA业务引擎/OCS账务服务；标准冲突及具名兼容按各自门槛验收 |
| 后续能力 | 保留现有分析与证据 | 限定S6a、16条未启用应用边界、完整SCUR/PRICE_ENQUIRY等；不为达到439/439扩首版helper |

固定规范版本及已采用勘误沿用现有Profile，不自动升级到最新Release。本次没有重新联网核验CR状态；2026-10-31的正文未取得记录保留为历史证据。CR0069、无IMSI成功和时间等号的外部澄清不作为普通已明确子集开始实施的前置；已有受限接收/报告行为仍必须实现和测试。

### 3.2 首版Ro的结果与边界补充（2026-11-04）

按用户要求继续文档审查，尚未切换实施；首版范围现按Draft 0.87；固定规范版本和24条需求ID不变。以下是SPEC-RESULT-01、SPEC-APP-01、SPEC-ERROR-01、SPEC-BIND-01及SPEC-REPLAY-01在首版Ro的具体解释；依据与反例见[主流程§29](diameterkit-mainflow-audit.md)，沿用[覆盖§35](diameterkit-usage-coverage.md)五包。

- Ro解释器MUST保留4011的“不再需要信用控制”语义，MUST NOT按4xxx一律解释为服务拒绝，也MUST NOT把它当2001、已扣费、已退款或余额充足。4010/4012的拒绝结果MUST NOT作为先前报告用量未扣除的保证。顶层/服务级与基础/实验命名空间不得互相覆盖；未知与不完整不得静默成功。
- 正常CCA的Session、Auth-App及CC类型/编号MUST按相应语法和原CCR关联；合法E错误MUST按独立基础包络校验，MUST NOT因缺正常CCA的Auth-App/CC字段而拒绝。ASA MUST NOT被要求提供CC类型/编号。DRA错误来源仅能在预配置错误分支授权范围内使用，MUST NOT冒充OCS成功或4011业务指示。
- ASA 2001 MUST表达应用报告会话已终止，5002表达会话不活跃，5012表达其他未停止原因。库MUST NOT把接纳ASR自动解释成2001，也MUST NOT因为不存业务会话而一律生成5002；外部状态和并发业务决策仍由应用负责，库不新增查询/会话表。Committed仅是本地应答提交事实。
- 同OCS、同活动exchange的原身份恢复是本库受限能力；MUST NOT以该限制宣称RFC4006禁止其他故障处理，也MUST NOT宣称已实现完整DCCA客户端状态机。旧六模板替换、SMS失败/序号和5011生成限制保持；此轮审查不新建泛化兼容开关或持久恢复接口。

同日继续审查补充，依据和预期见[主流程§30](diameterkit-mainflow-audit.md)、[覆盖§35.2](diameterkit-usage-coverage.md)，仍细化既有SPEC-APP/RESULT/PROFILE/SCOPE要求：

- MSCC解释MUST保留实例、目标服务集合和额度之间的关系；MUST NOT按请求/答案数组下标匹配，或将同组额度重复分配给每个Service-Identifier。缺少可靠关联或未支持费率/池限定时，MUST保留范围报告，不猜服务归属，也不将合法多组一律判为重复AVP。
- 标准退款helper MUST区分原IEC明确未返回2022、已返回2022（包括空值）与原答情况未知。已有凭据MUST按原服务关联原字节回送；未知MUST NOT默认为未返回，MUST NOT跨服务父组取token。完整历史验证缺输入须在本地说明；这不授权持久查询，也不解除标准SMS失败事件限制。
- SessionFeatures MUST与服务器支持报价分开：失败答案不能建立成功会话的协商依据；共同集合、请求实际需要集合和未知/空集MUST分别处理。MUST NOT从DRA广告、其他Session或其他IEC事件推导本次共同特性，不新增会话能力缓存。
- 对适用TS32.299 §6.5.5的ASR过程，文档及使用契约MUST明确应用CTF的CCR TERMINATION义务；库MUST NOT把ASA提交等同于终止报告完成，也MUST NOT自动引入跨命令事务或业务恢复。未有middle-tier依据时不得推导短信撤回或自动退款。

### 3.3 首版Ro受限能力与兼容证据门槛（2026-11-04）

本节细化既有SPEC-PROFILE/APP/RESULT/BIND/RELEASE要求，依据见[主流程§31](diameterkit-mainflow-audit.md)，验收归[覆盖§35.3](diameterkit-usage-coverage.md)；不新增需求ID或工作包。

- SMS失败3409与在线序号3408 MUST分用途报告支持范围。32.274在线表的SMS、T4 Device Triggering、MO-SMS via T4三列及收敛计费表MUST NOT互相替代；新版对照MUST NOT自动升级固定Profile。受影响标准helper MUST在接纳前明确未支持；接收MUST保留原值及未裁决范围，MUST NOT仅据冲突表格判对端违规或删字段伪造支持。
- 正常CCA的10415/5011安全接收MUST保持§3.2和主流程既定结果规则。标准自动生成MUST NOT擅选2001、5012或5003作伴随基础码，MUST NOT以Sh单实验结果、基础5011、未知必需AVP的5001分支或E包络绕过未决配对。只有目标配对证据时，MUST限于具名兼容声明，MUST NOT标标准冲突已解决。
- 六模板兼容MUST逐项记录部署/版本绑定、旧编码完整请求及输入、同操作的成功/失败答案、独立解读依据和新旧差异；模板加载、字典、同栈模拟器MUST NOT替代完整编码或独立对端证据。目标测试环境或有来源的脱敏现场材料均可，不要求生产抓包或业务账本。离线基线与目标互操作MUST分别标状态。
- 旧退款420与标准CC-Time在同连接/App/父路径的类型冲突MUST在固定binding中解决；MUST NOT按Requested-Action动态切decoder。旧解析器合并结果的行为MUST NOT使新库丢失顶层、逐服务及实验结果。上述证据由项目验收维护，MUST NOT引入库内持久记录、业务凭据查询或计费恢复引擎。

三项的解除条件已固定；无新直接依据或目标证据时保留限制，不继续扩写同类字典/报文来冒充关闭。普通已选Ro子集可按原五包在首版内实施，六模板完整替换必须另满足全部对应兼容义务。

### 3.4 首版Ro实施准入与交付声明（2026-11-04）

实施机制及检查组见[覆盖§36](diameterkit-usage-coverage.md)，sms-fw C/C++实际用法全覆盖的必需目标和执行优先级见该文§37及本页§3.5，仍归原五个Ro工作包、24个需求ID及首版内P5。本节的最小子集是中间交付，不是Ro目标完成；不改变§3.2/3.3协议规则。

| 类别 | 当前可实施能力 | 支持声明门槛 |
| --- | --- | --- |
| 不等外部裁决 | 已选codec/关联与结果、普通成功IEC/ECUR、通用余额/退款及2022、ASR/ASA、特性和5011安全接收、受限分支拒绝/报告 | 自身实现依赖及对应独立/运行验收通过；初始SingleAttempt，不宣称完整SMS失败流程或完整OCS服务 |
| 机制可做、目标启用待证据 | 同OCS恢复及§3.6受检Ro续发机制；已明确字段的六模板固定兼容定义/隔离和离线验证 | 目标恢复须ReplayDomain/路径依据，具名模板须逐项完整编码、绑定及答案证据；不明必需语义不靠猜测实现 |
| 标准语义继续受限 | 受冲突影响的3408/3409生成与完整SMS失败流程，正常CCA实验5011自动配对；其他已登记未支持语义 | 直接规范依据才能解除对应标准限制；目标证据仅授权经审查的具名兼容，不能更名为标准支持 |

未取得目标恢复证据的部署MUST保持SingleAttempt；机制测试通过MUST NOT代替终端去重域证明。最小Ro子集MUST完成操作/结果、5011及限制测试和适用发布验收，MAY明确未交付恢复或六模板；原RO-06与TEMPLATE义务仍保留，MUST NOT记为Passed。部分子集发布MUST NOT使用“六模板完整替换”声明。当前仍仅文档，不因准入明确而自动开始编码；首版范围已扩大到四应用，最小Ro子集不是v1完成。

### 3.5 sms-fw C/C++ Ro全覆盖要求（2026-11-04）

按用户明确要求，首版Ro完整交付MUST以sms-fw实际yppdia→dialib构建链中的全部Ro A/B用法为基线，逐项满足[覆盖§37](diameterkit-usage-coverage.md)；源码证据和差异见[主流程§32](diameterkit-mainflow-audit.md)。这是SPEC-APP/PROFILE/RESULT/BIND/REPLAY/RELEASE既有目标的具体化，不新增编号；Ro现已纳入首版。

- 必需范围MUST包含余额检查、IEC reserve/charge、ECUR INITIAL及投递成功/失败TERMINATION、失败退款与不发送分支、ASR/ASA、六模板/SCAP1、已接线Getter与配置差异、完整答案消费及路由/恢复/可用性信息。B类未证生产启用MUST NOT默认为可删除；不扩大为无调用依据的全部DCCA过程。
- 标准最小子集、单模板或SingleAttempt中间交付MUST NOT记为Ro全覆盖。标准未决MUST保留标准限制，但MUST NOT成为删除旧失败TERM/退款等必需用法的理由；按固定具名兼容取得输入/输出与目标证据。源码/模板存在MUST NOT等同于互操作通过。
- 旧业务读取的费用值/指数/币种、服务单位、退款字节、DDFH和私有operation_result MUST能无损交付并保留来源/Presence。未知Vendor/父路径/语义MUST作为实际兼容缺口，MUST NOT靠code-only查找、默填0、截断或只提供raw来宣称该用法已完整解释。
- 每项旧恢复/业务接线MUST说明库保证、应用职责和验证。普通同OCS恢复MUST NOT冒充跨OCS/重启重放；后两者按§3.6/运行时§1.7接纳，未实现或证据未通过MUST保留阻塞，MUST NOT以“归应用”标已覆盖。库无持久化、无账本/业务恢复引擎、Go-only边界保持。
- 旧缺陷可以纠正，MUST记录行为差异及迁移回归；不得以兼容为由重新引入无条件成功ASA、结果覆盖、token截断或缓冲成功冒充协议成功。库完整协议覆盖与业务binary/现场迁移MUST分别报告。

当前只固定必需范围和验收，尚未实现或通过全覆盖验证。原标准审查与六模板义务保留，新增发现只补相应用法缺口，不重开无界全规范分析。

### 3.6 RO-06恢复契约修订（2026-11-04，首版内P5）

本节显式扩展SPEC-REPLAY-01在Ro的范围，普通§3.2/R33入口继续受限；完整行为见[运行时§1.7](diameterkit-runtime-contracts.md)，来源/有限验收见[主流程§33](diameterkit-mainflow-audit.md)和[覆盖§38](diameterkit-usage-coverage.md)。保持24个需求ID及原五个Ro工作包，首版范围按Draft 0.87修订，不表示运行通过。

- 库MUST提供发送前固定原身份/内容的owned Ro请求记录，以及受检首次/续发入口；MUST NOT依赖文件、数据库或应用提交回调。普通Request和旧RequestID MUST NOT用作重启恢复。应用声明跨重启恢复时MUST负责可靠保存原记录/发送意图、累计额度与唯一发送者；仅进程内使用不要求存储，库只保证本Node接纳与冲突检查。
- 多OCS MUST按冻结目标表逐请求授权；ECUR已有会话迁移MUST有适用许可和连续性依据，SMS不适用AVP及SCAP1 MUST分别走具名兼容审核。DRA可用或同Realm MUST NOT当作去重/会话证明；缺证保持对应恢复未启用。
- 业务Tx、attempt响应等待及原请求总恢复期MUST分离。RO-06显式入口MAY在预定Tr到期且writer结清后建立后继；健康连接不必关闭，MUST NOT并行写。有效错误MUST先完成本段，应用只能按固定允许集合另开受检新段；未知/失败MUST NOT自动变成成功或补偿。
- 跨段MUST保留原E2E/Origin/Session/CC身份和业务字节；仅按冻结目标映射变更Destination，HBH/运行ID重新分配。无答可能发送与错误后续发MUST采用各自T位规则。ReplayUntil及累计次数/段数MUST有界且不可刷新，导入MUST经过标识guard、版本/绑定、年龄/来源和预算检查。
- 每个exchange MUST只完成一次；原段在活窗口内的旧答可赢，完成/关闭后的旧HBH答案MUST NOT附到新段。库MUST NOT承诺跨重启或最终窗口外无损晚答、磁盘队列、跨进程恰好一次。完整迁移MUST覆盖覆盖§38的应用接线与差异；这些不是库内持久化要求。

### 3.7 Ro私有字段与420兼容绑定（2026-11-04）

本节细化SPEC-PROFILE-01/02、SPEC-VALID-01、SPEC-RESULT-01、SPEC-BIND-01及既有资源/重放要求，24个ID不变。完整来源、字段表及有限证据出口见[主流程§34](diameterkit-mainflow-audit.md)，验收见[覆盖§40](diameterkit-usage-coverage.md)；以下是协议库契约，不提供模板DSL、计费决策器或存储。

- 私有字段MUST固定命令/方向、每层Vendor/Code、父路径及实例、类型、flags、出现规则和规则版本；Vendor MUST NOT从父组、CER或VSAI继承。`0/Code` MUST编码为V=0且省略Vendor-ID。未证Vendor MUST NOT填0/10415或作为通配符；未完成定义的具名typed能力MUST在绑定构造期拒绝启用，不影响其他已明能力。
- PSP的`873→20400→20430`及Integer32只由旧解析意图支持，三层接收Vendor/flags尚未证明。库MUST区分Absent、在位0/非0及Invalid/Uninterpreted，MUST NOT把DIACCAnswer的默认1当wire值，或以标准Check-Balance-Result替代私有结果；无Answer时MUST NOT产生虚构私有成功。旧CHECK的缺值不放行策略属于应用，MUST NOT改写Diameter结果或引发库自动重试。
- 旧退款发送MUST按已定顶层`0/1292 OctetString`解释，接收Code配置MUST NOT隐式修改发送Code；接收Vendor/路径须逐项锁定。标准`MSCC/10415/2022`与旧凭据MUST分开。库MUST交付owned完整bytes、Presence及父实例来源，MUST NOT截断、选最后一项或自动跨路径/Code转换。退款helper须消费应用明确关联的凭据；必要输入缺失/选择不明确MUST显式失败，不读取业务账本。
- 费用digits/exponent MUST保持有符号，currency/units保持各自协议类型与Presence；旧初值0及模板常量1/0/840 MUST NOT变成库缺省。具名420金额子字段425的旧INT32与标准Unsigned32 MUST分开定义，MUST NOT由类型名或相同正数字节推定完全等价。旧Getter/业务存储布局MUST由应用迁移，库负责受检值与编码。
- 标准Ro MUST保持420=Unsigned32、413=Grouped。`smsfw-ro-legacy420-v1` MAY在固定`CCR/0/456/0/437/0/420`唯一采用Grouped并明确排除该位置标准CC-Time；413仍可存在。该窄SMS组合MUST NOT声称全标准Ro，MUST NOT向其他父路径/CCA扩散例外。所有R/E分支的规则MUST预先固定；MUST NOT按Requested-Action、负载长度或消息Origin/Destination猜decoder。
- 相同固定键需要两型时MUST以真正可鉴别的部署解释域隔离，无法隔离MUST报BindingConflict；改Route名/PeerID不构成隔离。规则版本/定义MUST进入WireFingerprint，Route候选及RO-06原记录MUST保持解释一致。未知字段仍遵守M位/资源/错误规则，raw能力MUST NOT用来宣布私有语义已覆盖。

已明字段表示、构造/解释框架及冲突检测可实施；PSP定义、旧token答案链和420目标部署三组材料未齐时只限制对应启用/兼容声明，仍是首版完整Ro的门槛。它们并入原六模板证据包，不扩大为下一轮全量specification补齐。

### 3.8 纯Go实现约束（2026-11-04）

此约束适用于核心、Sh/S6c/SGd-Gdd/Ro、已承诺兼容Profile、shdata、生成器、公开工具/示例和独立SCTP module，不只约束公开API。保持现有24个SPEC ID，验收细化见[覆盖§41](diameterkit-usage-coverage.md)。

- 上述组件及其发布依赖MUST以Go实现，在Go 1.26.x与`CGO_ENABLED=0`下完整构建、普通测试及运行；MUST NOT依赖C/C++编译/链接、FFI加载的原生库或外部代理进程来实现已声明功能，MUST NOT以关闭功能的stub通过构建后宣称支持。正式Go生成器MUST可离线基于固定输入重生成，消费者MUST NOT需要旧C++ encoder或规范下载工具。
- Linux SCTP MAY通过Go syscall/受控ABI封装使用内核SCTP，MUST NOT要求libsctp/lksctp-tools。Go运行时/标准库汇编与OS接口不属于C/C++库依赖；平台相关unsafe/ABI布局MUST局限于底层边界并受架构测试约束，不扩为协议对象的任意指针转换。纯Go MUST NOT被宣传为跨平台SCTP或独立用户态SCTP栈。
- 缺内核/平台/必需API能力MUST显式不可用，MUST NOT静默切TCP或C实现。独立SCTP module和核心MUST分别检查完整依赖、构建与运行；目标GOOS/GOARCH、真实传输、交叉编译和ABI测试MUST分别报告，编译/模拟检查MUST NOT替代association验收。
- TLS和Sh-Data所选功能MUST由Go实现；外部OpenSSL/libxml2/Xerces等MUST NOT成为运行或正式生成依赖。独立规范/互通验证工具MAY在专门测试环境运行，不随库启动，不自动扩展Sh/XSD支持范围。
- 发布/普通测试的CGO=0路径与工具链race检查MUST分开；race MAY仅为Go检测工具启用cgo及编译器，MUST NOT将其视为产品C依赖例外或替代CGO=0验收。最低Go版本验证MUST固定1.26.x工具链，不得由自动下载更高Go版本掩盖失败。

“Go API”与“纯Go实现”间的约束遗漏已修正；测试工具与发布依赖、Linux内核与用户态C库、根module与SCTP独立module验收已分开。不增加协议/业务框架，不改变首版四应用或无持久化边界。2026-11-06核心CGO=0构建切片已开始验证（§6.1），SCTP构建/运行证据仍Planned；SCTP deadline等既有功能缺口继续归P3/U1–U5，Ro三组私有证据继续归原六模板包。

### 3.9 SCTP剩余P1与上下游边界（2026-11-04）

上下游责任与实施依赖见[部署§4.1–4.8](diameterkit-deployment-interop.md)，有限验收见[覆盖§42](diameterkit-usage-coverage.md)。skill215/sctp新快照已有U1生命周期/地址修复基础；U2完整消息信息、U3标准deadline/Context和U4所需安全通知/选项仍有设计要求未交付，MUST分别记录，不能将“全缺”或“无cgo即齐备”作为进度结论。

U2–U4的具体接口/返回表、九类竞争窗口及最小集合由[上游接口契约§2–7](diameterkit-sctp-upstream-contract.md)细化。Messages模式MUST在创建时固定并与旧reader/handler及用户Control互斥；缓冲不足、未消费通知过大、已消费结构错误MUST分别返回，实际n/flags MUST NOT被取消抹除。标准D、ctx C、固定timeout S MUST取早，旧timer/cleanup MUST NOT作用于新owner；Dial交接成功后ctx MUST NOT继续关闭连接。此次MUST固定SNDRCV/SCTP_EVENTS及五类通知/旧SEND_FAILED，MUST NOT把新版事件或全选项实现增为前置。解析器借用tail、I/O缓冲归还及上游五行实施清单均按该契约验收；不新增SPEC ID。

本库接线MUST满足[运行时§3.5.1–3.5.3](diameterkit-runtime-contracts.md)及[Peer§3.6](diameterkit-state-consistency.md)：接收结果MUST区分中间TransportEvent与完整帧，并沿原游标/预算/截止续调，MUST NOT把通知丢进可合并的观察队列。Tw MUST独立调度，MUST NOT直接当读超时；适配器固定S=0，首个Data字节开始部分帧期限。SCTP_SHUTDOWN_EVENT MUST停止所有新写并结清/Close，MUST NOT生成DPR/DPA；COMM_UP/路径通知MUST NOT成为Ready或watchdog存活证据。终止原因MUST先固定再关闭，晚到成功仅补实际发送证据。预算MUST在创建前覆盖上下游完整固定占用、并行候选及尚未退出I/O的旧代；MUST NOT仅按Peer数或IORequirements scratch计费。没有新的事件总线、持久恢复或SPEC编号。

- 上游MUST提供通用的有界数据/通知接收、完整recv flags和元数据有效性、可更新绝对deadline/可取消解析拨号、可靠fd生命周期及有限typed通知/选项。MUST在消费前区分数据与通知，MUST NOT让20字节前缀或旧callback先吞通知；peek不是一次已消费消息的EOR证据。U1回归MUST覆盖新增入口。
- 上游MUST NOT依赖Diameter规则、预算class、pending或重试策略。本库MUST负责Diameter组帧/长度/EOR、预算、事件策略、PPID/stream、CER/CEA与关闭屏障以及Peer/Completion传播；MUST NOT自行复制上游recv/poller或混用reader补缺能力。声明SCTP依赖时MUST计入底层固定暂存上界。
- 标准deadline更新/清除/延长MUST影响既有poller及gate等待，MUST NOT被旧操作清理覆盖；socket timeout的既有操作截止不刷新。单exchange取消MUST NOT改变共享reader deadline。Context解析与拨号MUST共用总期限且失败资源结清，MUST NOT后台遗弃阻塞I/O。所需行为以纯Go实现，保留§3.8。
- 上游U1–U5组合验收与本库P3/P6/P7 MUST分开，MUST NOT互相作为对方开始/发布的前置。可控适配器测试及codec/TCP/TLS可先行；真实SCTP启用MUST等待必要U2–U4及已验收候选。Linux选项按真实能力核对，MUST NOT编造SCTP_EXPLICIT_EOR支持或用禁止网络分片代替“一次send一条消息”；必需能力缺失明确拒绝，不C库/TCP降级。

本节未增加新SPEC ID、通用SCTP状态机或用户态SCTP实现；跨协议互通、平台运行与具体依赖版本仍待执行证明。

### 3.10 实施架构与代码边界（2026-11-06最终文档审查）

[实施架构](diameterkit-architecture.md)固定目录、包DAG、生成操作与运行时所有权；本节细化既有SPEC-GEN/OWN/SCOPE/TRANSPORT/RELEASE要求，不新增SPEC编号或协议支持范围。设计已进入实施，前四十八批离线/fake及真实TCP/TLS运行时、DRR、Listener/Serve和有界标识证据见§6.1–6.50；本节完整验收尚未完成。

- MUST采用核心module与独立SCTP适配module的既定边界。核心MUST NOT依赖SCTP module/上游；adapter MUST仅通过公共transport契约接入，MUST NOT访问核心账本或复制Peer。逐module发布MUST在GOWORK=off和固定依赖下验证，不靠本地replace通过。
- pdu/avp/wire/validation等中立层MUST NOT反向依赖应用/codec/Node；schema MUST NOT导入具体应用或codec。生成应用MUST经schema有限操作契约显式组合，MUST NOT通过init/全局注册、反射或热加载发现规则。仓库外生成扩展MUST可仅用公共契约编译，不依赖internal路径；生成IR保持私有。
- 公开codec与网络路径MUST共用internal/codecengine；公开Decode保持复制/owned承诺，内部帧转交MUST由root完成预算转移，engine MUST NOT依赖reservation或网络。完整envelope补齐归Codec/engine而非pdu叶包；仅FrameComplete输入可交完整decoder。
- Node/Peer/exchange/Arrival/完成/RO-06接纳状态MUST有唯一所有者，不按同步/异步或应用复制运行时。纯app包MUST NOT依赖Node；RO-06的Node关联入口归root，协议规则由固定Profile提供，MUST NOT借此引入存储/通用恢复插件。
- 标准类型、人工规则/helper、生成输出及具名兼容MUST分开维护；legacy420等不相容定义MUST独立类型/解释域，MUST NOT修改标准类型或运行时猜型。具名Context构造MUST保持封闭、有界及无包环，按架构§4的内部值/公开别名方案实施，不开放任意fact注入。

包内文件名、接口拼写及测量后的容量可在实现中微调；上述不变量、原24个SPEC ID和既定验收义务不能改变。架构验收并入原GEN-01/OWN-01/BOUNDARY-01/SPEC-01及相关PEER/BUD/SCTP，首批可执行切片和出口见架构§6；不追加实施前全规范调查。

## 4. 库级需求

24个SPEC ID保持稳定，通用条款及实际Ro/计费子句均在首版实施。§5的Ro场景纳入首版验收；普通非Ro调用无需计费配置。混合条目按§3与覆盖§21识别适用范围。

### 4.1 边界与定义

| ID | 可观察要求 | 细化来源 / 验收 |
| --- | --- | --- |
| SPEC-SCOPE-01 | 库及发布依赖 MUST 满足§3.8纯Go/CGO=0约束；MUST可在不配置数据库、数据目录、WAL、Store、持久标识provider的条件下使用；MUST NOT自行持久化协议或业务运行状态。进程内pending、去重缓存、预算和发号元数据允许存在 | 总体 §1；BOUNDARY-01、覆盖§41 |
| SPEC-SCOPE-02 | 库 MUST NOT 决定短信投递、订阅存储、扣费/退款、跨进程幂等或恢复；协议成功、Written 和 Respond.Committed MUST NOT 表示外部业务事务已完成 | API §5、运行时 §2；BOUNDARY-01、RO-01–06 |
| SPEC-PROFILE-01 | 每项声明支持的命令/字段 MUST 关联固定来源、完整父路径、Vendor/Code、类型、flags、出现条件及发送/接收行为；通用定义/ABNF MUST NOT 自动覆盖无冲突的已采用在线用途限制；规范证据冲突 MUST 逐规则登记，MUST NOT 仅凭表更具体判对端违规；不相容定义 MUST 在构造/生成阶段显式失败，MUST NOT 后加载覆盖 | Profile §1–3；PROF-01、RULE-01、FIELD-01/02、FIELD-10–14、NUMBER-01、SUB-01/02、IDENTITY-01/02、TEXT-01 |
| SPEC-PROFILE-02 | 标准、应用覆盖、本库策略、具名兼容 MUST 分别记录。未知扩展的安全保留 MUST NOT 被宣传为已实现其语义；未决规则 MUST 阻止相应完整支持声明 | 主流程 §1、§8–21；DEP-01、TEMPLATE-01/02、XML-02/03、TEXT-01 |
| SPEC-GEN-01 | 固定定义的生成 MUST 可复现，并报告未解析引用和冲突；包依赖及生成接入MUST满足§3.10；生成的强类型 codec MUST 不依赖运行时反射/tag 扫描，MUST NOT 要求先构造通用 AVP 树。性能收益须另有等价工作量证据 | 总体 §6、§13；GEN-01、PERF-01 |

### 4.2 编解码、校验与结果

| ID | 可观察要求 | 细化来源 / 验收 |
| --- | --- | --- |
| SPEC-CODEC-01 | 解码 MUST 在分配/访问前检查帧、AVP、父边界、整数溢出及资源限制；MUST 保留 Presence、同键重复顺序和规则要求的原值。Encode/Decode 的发送/接收约束 MUST 分开 | Profile §3；WIRE-01/02、FIELD-02、CODEC-01 |
| SPEC-OWN-01 | 默认 Decode/消息交付 MUST 返回 owned 数据；提交接纳后的发送快照 MUST 不受调用方后续修改影响。库拥有的快照与业务视图 MUST 分开，借用数据的期限须显式约定 | 总体所有权、API 请求/应答契约；OWN-01 |
| SPEC-VALID-01 | 校验报告 MUST 说明已检查范围、违规、缺输入和本地错误；缺输入 MUST NOT 被当成通过或对端缺 AVP；规范未裁决/语义未支持 MUST 与可补齐的 ContextRequired 区分，MUST NOT 用虚构上下文消除。Role/Direction/原请求证据 MUST 来自其所有者，调用方额外输入只使用具名有界值 | Profile §4、运行时 §2；CTX-01、VAL-01–03 |
| SPEC-RESULT-01 | 结果 MUST 区分基础/实验命名空间、顶层/服务级、未知/失败/冲突、缺值/零值；便利解释器 MUST NOT 从顶层成功推导全部服务成功。协议失败 MUST NOT 自动授权传输恢复或业务重试 | 主流程 §4、§9–10、§14/16–21；逐叶清单§4；RESULT-01–07、PAIR-03–07 |
| SPEC-ERROR-01 | 错误 MUST 可分类并提供有界、安全的路径/规则诊断。无可信 envelope MUST NOT 猜测生成答案；静态无效但可关联的应答 MUST 以 ErrInvalidAnswer 结束且不重传，MUST NOT 对答案再发错误答案 | Profile §3.3、运行时 §2.4、主流程§24；VAL-03、RESULT-01/04 |
| SPEC-APP-01 | 常用构造/解释 helper MUST 无网络和业务 I/O；SRA 解释需原 SRR 用途，Ro 需原 CC/服务关联，顶层失败 MUST NOT 被服务级成功覆盖；FieldSelection 只表达本地 OM/OC 构造选择，MUST NOT 当成远端事实或第三种 Context。最短 Service-Context-Id 下 MUST NOT 仅因缺OM/OC判对端违规 | API §2/§5、Profile §4、主流程 §10/16–21；FLOW-01–04、PROVISION-01、PAIR-05/06、XML-02/03 |

### 4.3 端点、重传与入站处理

| ID | 可观察要求 | 细化来源 / 验收 |
| --- | --- | --- |
| SPEC-BIND-01 | 绑定 MUST 区分邻接 Peer 与最终 Origin/Destination；同连接/App-ID MUST 有唯一相容解释。每份入站副本 MUST 重新检查来源/目的授权；DRA CER MUST NOT 证明终端特性 | 部署 §2、Profile §2；DRA-01/03、CTX-01 |
| SPEC-EXCHANGE-01 | 接纳失败 MUST 不产生异步 Completion；已接纳 exchange MUST 只产生一个最终结果，Port.Close 可按契约丢弃。迟到/重复答案、取消和新 attempt 竞争 MUST 不产生第二次完成 | 运行时 §1.2/1.4、API 完成端口；FO-03、REPLAY-03 |
| SPEC-REPLAY-01 | SingleAttempt、TransportFailover、ExplicitRetransmit MUST由同一调度者互斥执行；重传MUST保留原身份/业务内容、更新HBH，T位依无答发送证据或错误后续发原因确定。普通Ro恢复限同OCS活动exchange；RO-06受检跨目标/跨段续发MUST遵守§3.6，MUST NOT复活旧运行句柄、延长原恢复期限或重置累计次数 | 运行时§1.6/1.7、Peer/标识§4.7；FO-01–04、RO-06、REPLAY-01–04 |
| SPEC-EVIDENCE-01 | 结果 MUST 区分未写、部分写、完整写及真实不确定；后继 NotWritten MUST NOT 覆盖前驱可能到达的证据。ctx 取消 MUST NOT 被当成底层尚未结清的发送证据 | 运行时 §1.4、API 发送结果；FO-02/03、REPLAY-03 |
| SPEC-INBOUND-01 | 接纳的逻辑 Inbound MUST 只允许一次结果提交；Respond/Reject 校验失败不消耗提交权，提交后写失败不恢复提交权。应答 MUST 使用对应 Arrival 的 envelope，MUST NOT 无请求向重连连接发送旧答案 | 部署 §3、运行时 §2.3；DUP-01–04、VAL-02 |
| SPEC-DUP-01 | 去重 MUST 受明确域、语义指纹、内容比较、期限和预算约束；冲突 MUST 不触发第二次业务交付或回放首次成功。MUST NOT 通过 XML/号码/URI等价化/Unicode正规化/字段搬移“修复”内容，或承诺窗口外、重启后恰好一次 | 部署 §3；CANON-01、DUP-01–04 |

### 4.4 生命周期、传输与资源

| ID | 可观察要求 | 细化来源 / 验收 |
| --- | --- | --- |
| SPEC-PEER-01 | 传输连接成功 MUST NOT 等同应用 Ready；新请求、已有应答和控制发送门 MUST 分别受 Peer/watchdog/lifecycle 约束。旧代 timer/I/O MUST NOT 恢复新代状态 | Peer/标识 §3；PEER-01–04 |
| SPEC-CLOSE-01 | Shutdown MUST 使用一个有界总截止且只能缩短；Closing MUST 不开始新的应用写，匹配 DPA MUST 触发立即关闭。库 MUST 在 I/O 结清后释放借用对象，Wait MUST 不等待业务消费 Completion | Peer/标识 §3.5；PEER-04、BUD-01 |
| SPEC-ID-01 | HBH、E2E、Session 与内部句柄 MUST 分别满足其作用域和耗尽行为；原身份重传 MUST 不新领 E2E。内置保护只保证本 Node 生命周期，MUST NOT 声称已协调共享 Origin 的多进程/快速重启 | Peer/标识 §4；ID-04 |
| SPEC-BUDGET-01 | 库 MUST 在接纳/分配前申请全部所需有界额度，保持单账本转移且最后持有者释放后回收；控制保留额 MUST 不被业务借用。账本 MUST NOT 被宣传为 Go 进程 RSS 硬上限 | 运行时 §3；BUD-01–04 |
| SPEC-TRANSPORT-01 | runtime transport MUST 实现前缀分类后接收至调用者缓冲的边界、受控deadline/Close和发送证据；SCTP MUST在消费前分流通知，正确处理EOR、截断、元数据有效性、流数、PPID与模式屏障；上下游责任及组合前置MUST满足§3.9。缺必需能力MUST显式不可用，不静默降级 | 部署 §4–5、运行时 §3.5；SCTP-01–06、覆盖§42 |
| SPEC-SECURITY-01 | 身份/来源准入、受保护传输或明确受保护网络配置 MUST 有可审核边界；不能从连接可达推导授权。默认错误/观测 MUST 不输出完整用户载荷或无界标识集合 | 部署 §2、总体安全/可观测性；SEC-01、OBS-01 |
| SPEC-RELEASE-01 | 发布 MUST 逐能力记录版本、角色、平台/传输、限制、测试和偏差；MUST分别验证每个module的Go1.26/CGO=0发布路径与独立race检查，并分开声明库能力验证与现网DRA/OCS投产。v1前 MUST完成§3明确的首版A/B范围，Ro及计费兼容MUST纳入首版门槛，仅限定S6a等明确后置项保留待实施，未验证部分不得以“完整支持”发布 | 总体 §14、覆盖基线；SPEC-01、PROF-01、TEMPLATE-01、覆盖§41 |

上述需求不添加新的恢复引擎、全局特性缓存、通用规则回调或持久化接口。资源系数、配置值和候选API按§6的有界实施切片验证，已验部分见对应批次与公开示例；未验部分及最终API冻结仍需后续验收。实现选择可以变，但不能默改外部不变量。

## 5. 最小可验收场景

这些是分版本的规范场景，不是已执行测试；纯Ro及混合场景中属于四应用的部分均阻塞首版验收。每行须保留输入、动作、可观察结果和对应验收 ID；实现后附独立 fixture/时序记录。

| 场景 | 给定与动作 | 必须观察的结果 / 关联 |
| --- | --- | --- |
| 无持久化启动 | 仅提供内存配置和允许的传输；处理 Sh/SGd 请求后关闭并新建 Node | 不要求业务存储；新 Node 不导入旧 pending/去重；BOUNDARY-01 |
| 接纳失败与输入修改 | 将 class/Port 额度耗尽后 Submit；另一次成功 Submit 后修改原 body | 前者无 Completion/发送，后者保持接纳快照；OWN-01、BUD-01 |
| 原身份恢复与旧答案 | 固定 OCS、双 DRA；首 attempt 完整写后失联，后继接纳与旧 CCA 竞争 | 原 E2E/Session/CC 编号不变、HBH 更新、只有一个结果；REPLAY-03 |
| 双入口重复与冲突 | 同请求经不同 Proxy 链到达；再到达一份改变有效 TPDU 的副本 | 合法副本只交付一次且各自回 envelope；冲突不重做、不回放成功；DUP、CANON-01 |
| 关闭竞争 | 应用帧排队后进入 Closing；收到匹配 DPA，同时旧 writer 回调到达 | 不开始该排队帧；立即关 transport，结清已开始 I/O 且不复活；PEER-04 |
| 命名空间和结果冲突 | CCA 顶层 2001，但 MSCC 失败或 Experimental 10415/5011 | 不给出计费成功，不自动重传；保留原值与规则报告；RESULT-03/04 |
| SMSMI 按命令检查 | OFR、TFR、SRR 分别带相同三子项相关组 | 按各命令条件判断，不能“一组合法则处处合法”；FIELD-03 |
| OM/OC最短形式 | CCR使用最短Service-Context-Id并省略OM/OC，本地发送默认启用这些字段 | 不误报远端缺AVP，M/C及已有字段仍检查；PROVISION-01 |
| 金额与缺币种 | 解码负指数；CC-Money与Cost-Information分别缺Currency | 精确保留有符号二元组，父路径决定必需性；不猜币种、不分配巨大十次幂；FIELD-06 |
| 同服务多组与顶层失败 | 同服务键带已审核费率限定的多个MSCC；另给顶层失败/组内成功 | 不按键覆盖；不支持限定语义时报告不可用，顶层失败优先；RESULT-05 |
| 身份与内容比较 | 接收两份同键请求，URI地址等价但字节不同 | 按固定CanonicalV1保守报内容冲突，不执行第二次业务；FIELD-05、CANON-01 |
| 在线用途与证据冲突 | 最短SMS服务标识，分别在顶层和PS子组放设备信息，另放已采用表标not-used的项 | 不由OM/OC可选例外放行错误位置；不自动搬移；SMS-Result/Sequence冲突不猜线路错误码；APP-01、DECISION-01 |
| 有限设备/地址子集 | 明确真实16位IMEISV或数字GT；另给仅IMEI、未说明PC/GT含义的数字 | 已支持子集按明确规则构造；其余原值保留并报告未支持，不自动补F/判GT；FIELD-09、FIELD-08 |
| Realm标准与兼容域 | 同一合法Number+Name+Realm分别在标准域与显式s6c-ip-sm-gw-realm-v1解释；另给非法Realm | 标准域报告未裁决；兼容只覆盖SRA/Serving-Node三元组并标注兼容ID，不授权新路由；两域均检查非法语法；NODE-01、FIELD-07、DECISION-01 |
| 非IP地址重编码 | family8加14字节UTF8数字形成16字节Address payload | family仍为8，不按总长度猜IPv6；FIELD-08 |
| 标准主流程pair | SMS ECUR初始、S6c普通投递、SGd MO/MT及各自成功/失败答案 | 按原请求、角色和父路径校验；不将初始授权当ECUR完成，不执行投递或重试；PAIR-02 |
| 后续计费操作 | TERM同Session但新exchange；无GSU的TERM/退款结果；IEC4012；refund token含零字节 | 区分原动作和结果，不自动重建/退款；SMS失败分支仍按未决规则处理；PAIR-03 |
| 号码、订阅与特殊身份 | XML/TBCD/MAP地址分别表示；Sh访问键互斥/Private Identity、退订及Notif-Eff；IMS各命令HSS-ID/dummy/URI位置；Router token | 仅核对协议与有限表示，不查订阅或MWD；NUMBER-01、SUB-01、IDENTITY-01、ALERT-02、PAIR-07 |
| Sh四组与Sh-Data | UDR/PUR24、SNR25/OneTime0、独立PNR；2001和各自实验失败；空注销/可达0/错深度/错namespace | VSAI与结果关联、XML raw不规范化；库不管理订阅；PAIR-05、XML-02 |
| MME/SGSN角色分支 | ALR移动位图/有或无新节点；Gdd TFR1489与5555原因0 | S6c Alert App与SGd/Gdd App分开；不自动SRR/TFR或信任新节点；PAIR-06、FIELD-13 |
| S6c报告/提醒 | 空outcome、MSC/MME双cause、RDA实验5558、ALA E3004及错误SGd App | 程序条件/结果命名空间/E语法分别检查，不管理MWD或重投递；PAIR-04 |
| 模板源码与实际兼容 | 420 Grouped、STRING地址、SCAP1头与内部App及重复607 | 条件化源码推导不能代替完整执行/对端答案，标准同键类型不覆盖；TEMPLATE-02 |
| 深层值封装 | 同一TAI用于Ro tagged ULI和SGd EPS叶，另给4字节RAT及带RADIUS子头值 | 按Vendor/Code/完整位置校验，不跨封装自动转换；FIELD-10/11 |
| 完整DCCA pair | 给定匹配原CCR/授权OCS和单事件服务语义；再修改答案编号 | 正常配对可解释协议结果，错编号不交付成功；不证明外部账本已提交；PAIR-01 |
| SCTP 分段或截断 | 收到分段消息夹通知、CTRUNC 或与 Diameter 长度矛盾的 EOR | 不跨消息拼帧、不绕预算；确定失败/关闭且所有 pending 结清；SCTP-02/03 |

## 6. 需求与证据的维护

不建立新的验收编号体系替代原清单。`SPEC-*` 指库需求，已有 `RuleID` 指协议规则，覆盖基线 ID 指测试/用法。它们的关系为：需求 → 指定细化条款/RuleID → 正反场景 → fixture/测试证据。该关系可以多对多，不能仅靠文件存在认定已覆盖。第十四至十八轮候选样本的手工来源、字节长度检查与独立实现验证分别记录；VECTOR-01未通过前不能把候选称为已验证golden。

新增字段或兼容规则时，MUST 同步记录标准来源、发送/接收/Canonical 差异、结果行为、预算、验收和影响的 SPEC ID。改变旧输入接受语义或重复比较语义，MUST 更新相应语义指纹及兼容说明；不以 schema 包名或 Go API 未变为理由省略。指纹算法仍按部署契约验证，本轮不另造版本服务。

区分两次冻结：2026-11-03固定的是**实施范围和行为基线**，不要求实现证据先于实现产生；发布时才按capability核对规则、独立样本和运行验证，记录发布版本/提交。没有能力因本次收尾被标为Passed或获得符合性认证。最终模块路径、API拼写、默认资源数值和性能门槛在相应实施阶段确定，不能据这些常规实现选择重启整轮前置协议分析。

实施基线只因实现发现契约矛盾、安全/协议正确性问题、或已承诺实际流程缺项而修订；记录受影响需求、行为差异和回归用例。新可选能力、额外Release、兼容扩展及完整深层helper进入后续清单，不自动加入当前里程碑。现有规范冲突仍须显式解决或限制对应能力，不能靠“封版”忽略。文档不再按轮次穷举追加报文；实施测试按缺陷和验收需要增补。

### 6.1 第一批实施记录（2026-11-06）

本节保留第一批完成时点的范围与缺项；原始结果/散列在 [batch-01历史记录](../spec/history/batch-01.json)。当前追踪入口为 [spec/progress.json](../spec/progress.json)，最新变化见§6.10；切片/限制及复跑命令见 [spec/README.md](../spec/README.md)。需求ID保持原24项；ledger中 `completed` 仅指对应批次，需求级 `partial` 不自动升级为Passed。历史段落“仅文档/Planned”描述对应评审时点。

| 实施范围 | 代码与证据 | 未完成边界 |
| --- | --- | --- |
| P1：Go1.26工程、基础值、命令身份、有限操作契约及Profile冻结 | go.mod；avp/pdu/wire/validation/schema；import DAG与外部生成消费者测试 | Node/Peer/Transport、完整Report/Context/ExtensionSet及运行时操作尚未实现 |
| P2：严格JSON→强类型代码→统一codecengine→公开codec | base/app/s6c生成文件；生成一致性；Base六条正常控制消息与普通SRR/SRA | E包络、完整公共字段、四应用余量、Canonical及关联仍待后续P2/P3 |
| 基础所有权与资源 | owned输入复制/内部移交、深复制补全、Presence、Vendor/父路径、Grouped计量/写入一致、单次短写证据；codec/wire回归 | 不等于Node发送快照、预算转移、一次完成或取消竞态已完成 |
| 固定报文与独立解释 | DWR/DWA及V-SRR-01、V-SRA-01/02共5帧；go-diameter离线逐字段/Grouped/重编码对照；结果见interop | 仅5帧证据，不表示全局VECTOR-01、所有历史候选或真实DRA/HSS互通通过 |

核心保持纯Go、无持久化、无第三方运行依赖；工具链为Go1.26.10。CGO=0单元/外部消费者、vet/build、race（单列工具环境）、短时fuzz及示例的具体结果以ledger为准，不以本文陈述替代执行记录。当前Bundle指纹仅为已实现定义/组合的一致性标记，完整部署WireFingerprint及人工规则依赖锁仍待P2；不得用此切片接纳尚未实现的部署Binding。

第一批结束时确定的下一步为P2公共正常/E包络及操作/关联所需契约（现已实施，见§6.2），再接P3预算/完成/Peer和fake transport；SCTP上游U1–U5独立推进。Sh、完整S6c、SGd/Gdd、Ro/六模板/RO-06仍是首版内后续实施，不因本批范围有限而后置或删除。不再新增一轮全量前置specification审查。

### 6.2 第二批实施记录（2026-11-06）

本节保留第二批完成时点的范围，原记录/散列见[batch-02历史](../spec/history/batch-02.json)，最新实施见§6.20。第二批已完成公共正常/E包络、请求关联及运行时需要的有限owned操作，详细契约/限制见 [batch-02](../spec/history/batch-02.json)，实际检查与来源散列见 [progress.json](../spec/progress.json)。Profile升级为 `0.1.0-slice.2`，生成器/操作版本为2；它们仍是开发子集，不是四应用完整支持声明。

| 实施范围 | 验收证据 | 仍未完成 |
| --- | --- | --- |
| Base/S6c公共正常字段、独立E操作 | 正常/E分派、基础/实验结果命名空间、Proxy/Route-Record、CER/CEA VSAI、S6c Supported-Features结构及SMS 260忽略规则 | 完整应用分支；3006 Redirect及其必需字段；E下已知但未实现的应用字段明确Unsupported |
| 请求关联与来源 | 先匹配App/Command/HBH/E2E，再校验P/T、Session、代理实例、正常277和Origin；终端与E错误代理权限分开；未匹配/匹配后无效答案分类；调用方变更与并发读取回归 | ConnectionID/AttemptID选择、真实Binding、pending、一次完成、重试与晚答处置 |
| owned操作与诊断边界 | Measure/Snapshot、内部不可变RequestSnapshot、DecodeBound/RequestBound/CompleteAnswerBound；单个Failed-AVP引用多AVP，畸形引用停止分支并保留证据，内层与整帧共享数量/深度限制 | 实际预算预留/转移、坏请求拒绝路径、完整Report/Context、Canonical；估计不是RSS保证 |
| 生成器与对照 | 外部生成消费者正常/E配对、操作表冻结、固定输入新增11帧；其中7帧经go-diameter独立解读与重编码；另记坏长度诊断差异 | 仍非全局VECTOR-01、真实DRA/HSS/OCS网络互通或完整生成依赖锁 |

CGO=0测试/vet/build、生成一致性、依赖审计、示例、race和请求关联短时fuzz均已执行通过，具体次数及命令以ledger为准。[独立离线证据](../interop/evidence/2026-11-06-envelope.json)记录工具与对端源码散列。SPEC-BIND/EXCHANGE/BUDGET的相关条款仅标Partial，不能把纯函数或保守上界当成对应运行时验收完成。

第二批结束时确定的下一入口（已在§6.3实施）为P3预算/一次完成/每连接pending与Peer的有界fake transport切片：Base握手→单次SRR/SRA→取消/关闭；通过后接TCP/TLS。尚缺的Canonical/Context/完整应用定义在其依赖阶段实施，不回到无界前置分析。SCTP U1–U5及四应用其余流程、Ro/兼容/RO-06仍保留原首版义务；核心继续纯Go、无第三方运行依赖、无持久化。

### 6.3 第三批实施记录（2026-11-06）

本节保留第三批完成时点。该批已贯穿固定Peer的SingleAttempt出站运行时，具体API/预算/状态边界见[batch-03](../spec/history/batch-03.json)，当时检查及源文件散列见[batch-03历史](../spec/history/batch-03.json)，最新实施见§6.20。公共Envelope新增目的字段Presence/值提取，修复realm-only绑定可能保留调用方Destination-Host的问题；Profile版本为`0.1.0-slice.3`，生成器/操作版本为3。原请求/答案线路定义未因此增加应用范围。

| 本批已执行 | 验收证据 | 未完成边界 |
| --- | --- | --- |
| 单账本四维准入、结果转移和释放 | 原子失败/溢出、peer/port窗口、未消费结果、控制独立预留及取消/关闭借用屏障；internal/budget与runtime_test | 完整Inbound/Duplicate硬分区、多attempt split/transfer、目标平台内存/负载计量 |
| 每连接pending与一次完成 | 先登记后写、错头不消耗A、超限头读body前失败、快答/Cancel/Close竞争、短写/不确定写、结果消费释放 | Route/failover/显式重传、完整attempt摘要、应用入站及重复缓存 |
| Peer/Base/fake transport | 主动/被动CER/CEA、CEA完整写屏障、DWR/DWA、SUSPECT与REOPEN三探测、远端DPR/DPA、旧连接不继承pending | 双连接选举、Maintain、本地Draining/Shutdown与完整交叉关闭、TCP/TLS/SCTP生产适配 |
| 标识与固定绑定 | BootID/连接和请求序列、HBH不复用及控制余量、应用/控制E2E半区；目的和终端/E代理来源分别受检 | 自定义源/guard后回绕/NewSessionID/恢复号域及完整部署Binding |

核心CGO=0、全量测试/vet/build、race、关键并发/预算20轮重复、生成一致性及无外部依赖检查的实际结果记入ledger。fake对端使用同一codec，**不属于新增独立协议栈互通证据**；前两批已有离线对照仍保留历史归属。SPEC-BUDGET/EXCHANGE/PEER/CLOSE/ID/TRANSPORT等仅按已测范围标Partial，不关闭完整P3或v1。

第三批确定的下一入口已在§6.4补齐本地Draining/Shutdown/DPR-DPA及候选选举的有限事件，接下来进入TCP/TLS。入站/Canonical/去重、多attempt恢复、SCTP U1–U5、四应用余量和Ro/六模板/RO-06仍属原首版义务；不再增加全规范前置轮次。

### 6.4 第四批实施记录（2026-11-06）

本节保留第四批完成时点。该批完成本地关闭和已建立连接的双候选选举，实际API、状态及截止语义见[batch-04](../spec/history/batch-04.json)，当时检查和源文件散列见[batch-04历史](../spec/history/batch-04.json)，最新实施见§6.20。Profile/生成操作仍为slice.3/ops3，没有扩大应用协议定义。第三批原始ledger已保存于[历史记录](../spec/history/batch-03.json)。

| 本批已执行 | 验收证据 | 未完成边界 |
| --- | --- | --- |
| Peer/Node Draining与Shutdown | 关闭新接纳、已接纳首次写、排空/总截止、多Peer共用截止、重复缩短/取消、Close升级；shutdown_test | 应用Inbound拒绝与完整FailurePolicy、生产fd关闭 |
| DPR/DPA与一次完成 | 独立控制关联、正常/E DPA立即关闭、交叉/重复/错头控制、部分帧不插DPR、快答与晚writer、DPA等待超时 | 生产传输下相同契约的执行证据 |
| 双连接选举与退休屏障 | 双Node交叉连接同向收敛、主动失败被动接管、被动EOF、第三候选/热替换拒绝、身份先校验、共享握手期限、Shutdown阻止CEA安装；election_test | Dialer/Wait-Conn-Ack、Maintain和完整RFC事件表 |
| 资源与借用退出 | 两候选固定预留、选举前缀读取移交、槽退出前不复用、Shutdown超时与Wait清理分离、未消费结果仍计费 | Inbound/Duplicate及多attempt账本、目标平台内存/压力验证 |

新增12组关闭、7组选举测试；全量CGO=0 test/vet/build、race以及root 20轮race重复通过。生成一致性、外部生成消费者与导入DAG继续通过；详细命令及纯Go依赖审计以ledger为准。fake对端和双Node使用同一codec，不新增独立互通证据。24项需求ID不变，SPEC-CLOSE/PEER/OWN/BUDGET等仍按有界覆盖标Partial，不宣称完整P3/v1通过。

第四批确定的TCP/TLS生产适配及真实连接期限/关闭测试已在§6.5实施。SCTP U1–U5、入站/Canonical/重复缓存、多attempt恢复、四应用余量及Ro/六模板/RO-06仍保留首版义务；不恢复无界前置规范审查。

### 6.5 第五批实施记录（2026-11-06）

本节保留第五批完成时点。该批完成TCP/TLS生产适配及真实本地连接验收。公开入口、所有权、TLS资源排除项与关闭策略见[batch-05](../spec/history/batch-05.json)，当时检查及源散列见[batch-05历史](../spec/history/batch-05.json)，最新实施见§6.20。已有Profile slice.3/ops3和24项SPEC ID不变。

| 本批已执行 | 验收证据 | 剩余边界 |
| --- | --- | --- |
| TCP/TLS单reader/writer、两阶段收帧和固定scratch | 分片/粘包/discard、错误长度/游标、预算不足接管拒绝 | SCTP事件续调与U1–U5、目标平台资源测量 |
| 绝对deadline/取消/关闭 | 首字节计时、prefix→body不续期、旧回调退出、部分写证据、并发Close解除I/O | 完整生产负载/跨平台验收 |
| 有限TLS握手与证书认证 | TLS 1.2/1.3 mTLS、未知CA/名称/缺证书、双方timeout/cancel/EOF | 部署级证书与Diameter身份映射、接入并发策略仍由应用负责 |
| 真实Node路径 | CER/CEA→SRR/SRA→DPR/DPA、晚答、取消/关闭竞争、双Node选举和交叉Shutdown | Inbound/重复缓存、自动Dialer/Maintain、Route/multi-attempt与四应用余量 |

新增17组真实连接测试；全量CGO=0 test/vet/build、race和20轮root race重复的实际记录在ledger。修复调用方截止与握手取消转发竞争导致DeadlineExceeded被标Canceled的问题。root只依赖transport契约，TCP/TLS通过内部共享stream实现复用有限帧cursor；核心无第三方依赖、无持久化。SPEC-TRANSPORT/CLOSE/PEER/SECURITY等仅按已执行范围Partial；不得将本地真实socket测试记为独立Diameter互通或完整P3/v1通过。

第五批确定的有限Inbound/应答已在§6.6实施，有限拒绝在§6.7完成，余下按依赖接Canonical/重复缓存及多attempt；SCTP上游/适配独立推进。Ro及其六模板/兼容/RO-06的首版义务保持。

### 6.6 第六批实施记录（2026-11-06）

本节保留第六批完成时点：单Arrival、显式NoCache的有限Inbound/应答，实际接口/失败边界见[batch-06](../spec/history/batch-06.json)，当时执行检查/散列见[batch-06历史](../spec/history/batch-06.json)，最新实施见§6.20。保持24项SPEC ID和既有线路Profile，不将显式无缓存处理记为去重实现。

| 本批已执行 | 验收证据 | 剩余边界 |
| --- | --- | --- |
| Peer.Receive与入站接纳 | 精确终端来源、本地目的、已协商命令、prefix后预留再读body、独立视图 | 完整Binding/Relay许可、NeedsApplication/Check及逐阶段报告 |
| Inbound.Respond一次提交 | 私有Session/Proxy/头补齐、SRA成功/普通失败/实验失败/E、校验失败可修正、提交前后取消区分 | 自动Reject/FailurePolicy、应急拒绝预算、多Arrival重答 |
| 独立Inbound预算 | Node/class/Peer同charge、与出站硬隔离、Base控制继续、满额不等待业务 | C缓存及完整split/transfer、目标平台内存/压力验证 |
| 关闭/期限/旧代 | 准备和writer借用退出、Closing撤销队列、Draining答案先于DPR、断线句柄与新连接隔离 | Canonical/重复缓存、完整DRR、完整Peer/Dialer/Maintain |

新增17组真实TCP/TLS入站测试，并用受控回调/写返回延迟检验预算屏障。普通构建保持CGO=0，race单列工具链；全量test/vet/build、race和20轮root race记录在ledger。SPEC-INBOUND从Planned变为Partial；SPEC-DUP仍Planned。所有Committed仅表示本句柄提交，不表示缓存/持久化、远端或业务成功；完整P3/v1仍未通过。

第六批确定的可信envelope与有限自动Reject/应急预算已在§6.7完成，后续接Canonical/重复缓存；Ro及六模板/兼容/RO-06继续在首版，SCTP沿U1–U5独立推进。

### 6.7 第七批实施记录（2026-11-06）

本节保留第七批完成时点：有限Reject/FailurePolicy与应急拒绝预算，接口、固定映射和失败出口见[batch-07](../spec/history/batch-07.json)，当时检查/散列见[batch-07历史](../spec/history/batch-07.json)，最新实施见§6.20。继续24个SPEC ID、slice.3/ops3；SPEC-ERROR/INBOUND/BUDGET/CLOSE等仅按以下范围Partial，SPEC-DUP仍Planned。

| 本批已执行 | 验收证据 | 剩余边界 |
| --- | --- | --- |
| 显式FailurePolicy与Inbound.Reject | 已知请求来源/目的3002、限定3004、5012超时及具名失败；与Respond共用准备/提交权 | 完整Context/Check、报告与应用专用失败映射 |
| 可信E包络与固定基础诊断 | 3001/3007、五个顶层必需字段5005占位、Session/Proxy完整关联；坏包络关闭；不回答answer | 深层/实例诊断、完整5001/5004/5009与3006 |
| 独立应急预算 | prefix后先预留；固定Transport额度内RX/TX槽；满额/超长关闭，不等待业务或重复扣Node动态额度 | C缓存与完整I/W/C、多attempt转移、堆/压力证据 |
| 自动超时与关闭竞态 | 超时帧交付前预生成；手动/自动唯一提交；读取/准备/写入借用退出后释放；已完成写入不被旧截止改写 | 多Arrival重答、完整DRR与Peer事件余量 |

新增19组真实TCP/TLS测试；全量CGO=0 test/vet/build、race、root 20轮race和生成/依赖审计按ledger记录。没有新增第三方运行依赖、持久化或独立Diameter互通声明。BaseErrorFailures必须显式启用；零值保留先前关闭/终止行为。3004只用于请求明确指定本地Host，realm-only过载不伪造3004。已知命令的超时预生成帧无法在现有包络/预算内构造时，关闭且不交付业务。

第七批确定的Base/S6c子集Canonical、重复冲突、有限结果缓存、多Arrival重答和C额度/保留期限已在§6.8实施；不重开全量前置分析或穷举诊断。完整S6c、Sh、SGd/Gdd、Ro及六模板/兼容/RO-06保持v1义务，SCTP按U1–U5依赖继续。

### 6.8 第八批实施记录（2026-11-06）

本节保留第八批完成时点。该批完成现有Base/S6c请求的CanonicalV1操作，以及普通S6c的有限Node内存重复缓存和多Arrival重答；详细配置/生命周期/预算边界见[batch-08](../spec/history/batch-08.json)，当时检查和源散列见[batch-08历史](../spec/history/batch-08.json)，最新实施见§6.20。保持24个SPEC ID；生成器/操作升级ops4，Profile slice.4，不扩充线路字段或命令范围。

| 本批已执行 | 验收证据 | 剩余边界 |
| --- | --- | --- |
| 原始request规范化 | 已知Identity折叠、顶层已知键稳定排序、未知屏障、同键重复/组内Sequence、原始Presence及固定SMS 260忽略 | 四应用余量规则、完整部署/角色指纹和上下文比较 |
| 同身份查重/冲突 | Node+Origin/E2E唯一键、T=0、跨Peer原子创建、来源独立准入、内容/有限策略冲突5012 | NoCache路径不提供去重，跨进程/业务幂等由应用负责 |
| C及多Arrival | 交业务前预留C、每副本I/W、结果去首份Proxy、独立头/代理链重答及组合后codec校验 | 最小I/W拆分转移、完整DRR、压力/堆证明 |
| 固定期限与关闭 | KeepUntil不刷新、超时缓存、原连接断开、REOPEN后新副本、Suppressed、到期/Close借用和实际写前关闭门 | 完整Inbound报告/具名Context、更多传输与四应用余量 |

新增21组root测试和Canonical fuzz入口，CGO=0全量test/vet/build、全量race及root 20轮race、短时fuzz以ledger为准。SPEC-DUP从Planned变Partial；SPEC-REPLAY的出站显式恢复尚未由入站缓存关闭。只验证本机TCP/TLS，同源codec不等于独立DRA/HSS/OCS互通。完整P3/v1仍未完成。

第八批确定的固定同终端Route/多attempt及原身份重传已在§6.9实施；完整四应用与Ro/RO-06首版义务保持。

### 6.9 第九批实施记录（2026-11-06）

本节保留第九批完成时点。该批完成普通SRR的固定同终端Route、自动传输故障转移和活动请求内显式原身份重传，详见[batch-09](../spec/history/batch-09.json)，当时执行/散列见[batch-09历史](../spec/history/batch-09.json)，最新实施见§6.20。生成器/操作升级ops5、Profile slice.5，显式TransportReplay许可纳入指纹；线路字段/命令和24个SPEC ID不变。

| 本批已执行 | 验收证据 | 剩余边界 |
| --- | --- | --- |
| 固定Route与单一调度者 | 三种显式模式、有限候选/次数/固定目标/保留窗口；Peer始终SingleAttempt | 完整角色/部署指纹、动态接入管理未完成 |
| 原身份与实际T位 | 共享不可变请求，保留Origin/Destination/Session/E2E；新连接内HBH，旧writer最终返回后才接纳后继 | 仅普通SRR许可；Ro/RO-06、跨OCS及重启恢复入口未完成 |
| 逐attempt预算/关联 | 接纳前预留N个A，独立pending/SourcePolicy；旧连接晚答、新旧reader同时借用、全部借用退出才转移完成额度 | 保守预留，非最小Q/A/R拆分或堆/RSS证明 |
| 显式恢复与一次完成 | RequestInfo/幂等after边、无备用/Peer预算不足保留原期限；单Peer新代REOPEN、Cancel/关闭/Shutdown、聚合风险与一个Completion | 无完成历史或业务幂等；无自动等待重连/业务错误重试 |

新增23组Route测试及schema/生成器许可用例；CGO=0全量test/vet/build、全量race、root 20轮race（97.494秒）、生成与依赖检查通过。未新增fuzz或独立网络Diameter互通证据。SPEC-REPLAY由Planned转Partial；SPEC-EXCHANGE/OWN/BUDGET/ID/PEER/CLOSE/EVIDENCE等按本批范围仍Partial，完整P3/v1仍未完成。

第九批确定的有限Dialer/Maintain接纳、取消/关闭、退避及重连已在§6.10实施，重连不自动重放业务；其他首版义务保持。

### 6.10 第十批实施记录（2026-11-06）

本节保留第十批完成时点：有限Dialer/Maintain与TCP/TLS拨号实现，具体接口、共同截止、预算与回调边界见[batch-10](../spec/history/batch-10.json)，当时检查/源散列见[batch-10历史](../spec/history/batch-10.json)，最新实施见§6.20。24个SPEC ID、ops5及Profile slice.5不变；未增加应用线路字段或命令。

| 本批已执行 | 验收证据 | 剩余边界 |
| --- | --- | --- |
| 拨号前接纳 | Node.Dial、transport.Dialer与TCP/TLS固定端点；预占主动候选、声明/实际adapter检查、共同TCP/TLS/CER期限 | Listener/Serve、动态发现与完整跨平台接入未实施 |
| 管理所有权/退避 | 显式MaintainPolicy、每Peer唯一管理者、有限指数退避/抖动、连续OKAY重置、协议/TLS与DPR抑制 | 完整Peer事件报告、安全/压力和独立对端验证仍待完成 |
| 候选/取消/关闭 | WaitConnAck/Elect、主动失败接管、被动胜出/晚返回隔离；Node/Peer.Wait保留回调与I/O屏障；Shutdown继续排空 | 任意不合作用户回调不保证可强制终止；逻辑预算不是RSS证明 |
| 重连与请求边界 | 新ConnectionID、REOPEN三DWA、SingleAttempt不回放；Explicit仍Awaiting直到显式Retransmit；管理取消不触发Route恢复 | Ro/RO-06、业务幂等和重启恢复未由本批完成 |

新增24组测试，具体纯Go全量test/vet/build、全量race及重复race结果以ledger为准；socket deadline与context timer先后不改变总截止的DeadlineExceeded分类。每Peer新增固定4096字节管理预留，复用两候选槽；没有每业务请求goroutine或持久记录。相关需求仍Partial，完整P3/v1未完成。

第十批确定的按class/字节DRR调度与资源公平性已在§6.11完成。Listener/Serve、具名Context/Report、标识源、完整四应用与Ro/兼容/RO-06、SCTP U1–U5及适配器沿原依赖继续，不重开无界前置文档审查。

### 6.11 第十一批实施记录（2026-11-06）

本节保留第十一批完成时点：每连接按类别/class的编码字节DRR。实际API、调度/预算及证据边界见[batch-11](../spec/history/batch-11.json)，当时执行检查和源散列见[batch-11历史](../spec/history/batch-11.json)，最新实施见§6.20。24个SPEC ID、ops5和Profile slice.5保持，不新增线路字段或运行时依赖。

| 本批已执行 | 验收证据 | 剩余边界 |
| --- | --- | --- |
| 冻结配置/两层DRR | SchedulerConfig/SchedulerInfo；类别4:2:1及class正权重，完整帧cost，FIFO、空队列清零/结束访问、整数跳轮及有界deficit | 默认性能调优和完整四应用混合压力未验收；非跨Peer全局调度 |
| 请求/应答/控制公平性 | Base DWR与S6c SRR/SRA不同帧长的字节份额与无饥饿测试；应急Answer独立lane，缓存重答含各Arrival Proxy | 不承诺墙钟延迟或抢占；当前单应用命令限制下，同连接多class由内部真实编码帧测试覆盖 |
| 资源硬隔离 | 固定TransportReserve覆盖计划/两候选调度状态；class字节/槽超限不借用，完成消费者停滞不阻塞另一class | 最小I/W/Q/A拆分、堆/RSS及目标平台压力仍未完成 |
| 取消/关闭/控制截止 | TCP/TLS混合帧、Draining→DPR/DPA、队列撤销清credit、writer实际退出才归还预算、排队Control超时关闭 | Listener/Serve、完整事件报告、SCTP及独立对端验证未完成 |

新增15组测试，全量CGO=0 test/vet/build、全量race和20轮root race通过；重复race报告113.592秒。选择器不分配payload/队列、不增加每请求goroutine/timer。已有Route/多Arrival/选举/Maintain回归保持。SPEC-BUDGET/PEER/CLOSE/TRANSPORT等仍Partial，release_status仍not_ready；本机同源codec不作为新增独立Diameter互通证据。

第十一批确定的有限Listener/Serve已在§6.12完成；具名Context/Report、标识源、完整四应用及Ro/兼容/RO-06、SCTP U1–U5/适配器沿原依赖实施，不将其遗漏或重新后置到首版之外。

### 6.12 第十二批实施记录（2026-11-06）

本节保留第十二批完成时点：固定Peer的有限Listener/Serve，实际API、两阶段握手、预算和所有权见[batch-12](../spec/history/batch-12.json)，当时执行检查/源散列见[batch-12历史](../spec/history/batch-12.json)，最新实施见§6.20。24个SPEC ID、ops5与Profile slice.5保持；无第三方运行时依赖或持久化。

| 本批已执行 | 验收证据 | 剩余边界 |
| --- | --- | --- |
| Listener/Incoming与Serve | TCP/TLS包装预绑定listener，固定Peer唯一Serve；先占被动候选再Accept，Open时等待容量 | 多Peer listener/admission、动态发现与完整角色绑定未实现 |
| 两阶段握手/选举 | 空闲Accept不耗握手期限；TLS/CER/CEA共用绝对deadline；双Node选举、主动失败被动接管、Maintain共存 | 完整RFC事件报告、跨平台与独立对端验收未完成 |
| 取消/关闭 | 停监听和未安装候选；已建立连接保持Peer所有权；Shutdown继续排空/DPR；迟到Accept/Establish/CEA不能安装且Wait保留借用 | 不合作自定义回调不能强制杀死；完整压力证明未完成 |
| 资源/错误 | ListenerBytesPerPeer显式启用，固定listener/管理reserve；声明/实际adapter检查、同方向独占、握手失败有限间隔再接纳 | 标准库/内核/RSS不在逻辑账本证明内；无多Peer服务调度承诺 |

新增21组测试，全部CGO=0 test/vet/build、全量race及20轮root race通过；重复race报告137.877秒。被撤销但未退出的Accept仍占候选，Maintain不能绕过该槽反复空转拨号。真实TCP/TLS仅本机和同源codec，不作为新增独立协议互通；既有DRR/Route/多Arrival/一次完成回归保留。SPEC-TRANSPORT/PEER/CLOSE/BUDGET等仍Partial，release_status仍not_ready。

第十二批确定的普通标识分配/NewSessionID/guard已在§6.13完成；具名Context/Report、诊断、完整四应用及Ro/兼容/RO-06、SCTP U1–U5/适配器和发布门槛沿原依赖实施，不重新开启无界前置文档审查。

### 6.13 第十三批实施记录（2026-11-06）

本批完成普通请求的有界标识分配、NewSessionID、guard预算及HBH耗尽排空。精确接口和保证边界见[batch-13](../spec/history/batch-13.json)，检查/源散列见[progress.json](../spec/progress.json)。24个SPEC ID、ops5及Profile slice.5保持；没有增加存储、全局ID服务或线路定义。

| 本批已执行 | 验收证据 | 剩余边界 |
| --- | --- | --- |
| 内置E2E空间/轮次 | 配置Prefix/CounterBits，应用/控制半区独立；Active=0且最后退休+guard已到才重开轮次 | 多实例共享Origin/快速重启的联合唯一性仍由部署负责 |
| 自定义源及硬配额 | 锁外串行，源忙/配额不足即失败；固定slot/hash/到期堆，class隔离、错误候选一次拒绝 | 受信回调不能强杀，任意源私有内存/跨Node协调不由库预算证明 |
| Session和退休 | NewSessionID有界owned格式；Route共用lease，Awaiting仍Active，业务/控制writer结清后退休 | Ro会话推进、RO-06恢复准备/导入及其号域未完成 |
| HBH耗尽 | 停止新业务接纳，有限排空，用1024控制保留号DPR/DPA关闭该代；稳定Peer保持Active | 完整Peer事件报告、目标负载/平台与SCTP验收未完成 |

新增15组测试，包含guard索引6000步对照、时间边界、真实TCP/TLS请求/应答、迟到writer及源关闭屏障。全量CGO=0 test/vet/build、全量race和20轮root race通过，重复race报告140.821秒；原Route/缓存/DRR/Listener/Maintain回归保持。SPEC-ID/BUDGET/CLOSE等仍Partial，release_status仍not_ready。轮次/到期测试使用内部状态加速，不以本地nonce声明跨重启确定唯一。

下一批优先现有Base/S6c的具名Context/Report及分阶段校验；完整四应用及Ro/兼容/RO-06、SCTP上游/适配器和发布压力证据继续保留首版义务。

### 6.14 第十四批实施记录（2026-11-06）

已完成现有Base/S6c切片的有限Context/Report与分阶段入口，契约见[batch-14](../spec/history/batch-14.json)，检查与散列见[progress.json](../spec/progress.json)，前批冻结证据见[batch-13历史](../spec/history/batch-13.json)。ops6/slice.6显式NoContext审核纳入指纹，不增加线路命令/字段或完整应用支持声明。

- codec.ValidateContext及ValidateAnswerContext明确区分wire、原请求/来源关联、缺输入和未支持语义；Report Scope绑定Bundle/命令，CheckedRules目前是阶段级，非逐叶trace。
- 当前Context只有零值；普通S6c不需额外facts。Inbound.Validation/Check使用私有已验证证据，返回owned报告；取消准备/坏答案不消费提交权，过期/关闭不能恢复Ready。成功提交保留历史验证结论。
- Result.Validation随正常/E答案及关联失败发布；wire/关联通过但语义未审时保留owned Answer和未完成报告，不自动重传。NoContext未审请求在接纳前拒绝，不能用facts跳过。
- 报告为固定小集合和首失败分类，使用原bookkeeping预留，不引入持久化、外部查询或无界诊断。Sh/Ro具名非零输入、实际NeedsApplication→Check→Ready与冲突验收随对应应用实现；SPEC-VALID/RESULT/INBOUND等仍Partial。

本批4组离线和5组真实TCP/TLS测试及全量检查见ledger。下一批进入S6c RDR/RDA、ALR/ALA的明确子集及必要字段/结果，以成对codec/运行时测试结束，不重开P0或全量文档补齐。

### 6.15 第十五批实施记录（2026-11-06）

已完成S6c普通MSISDN报告和HSS提醒子集，详细范围见[batch-15](../spec/history/batch-15.json)，当前检查/散列见[progress.json](../spec/progress.json)，前批冻结证据见[batch-14历史](../spec/history/batch-14.json)。生成器7、Profile slice.7，操作契约保持ops6；24个SPEC ID不变。

- 新增RDR/RDA、ALR/ALA及User-Identifier、单MME Outcome、具名Time线路类型。RDR显式cause 0/1/2，诊断仅MME已审核13；RDA基础成功/选定失败、实验5001/5558，ALA基础成功/失败及独立E。HSS ALR强制目的Host/Realm，未启用其他身份/节点/提醒角色。
- 原请求关联使用正确的Receive校验；本端出站仍用Send，避免合法收到的未知RDR flag导致Respond失败。NoContext只证明声明切片无额外facts，不代表完整S6c。
- RDR与ALR有独立pending/标识/完成；沿既有有限预算、一次提交、Canonical及多Arrival重答，未给它们授予TransportReplay。不建MWD，不自动发提醒或业务重试。
- 6组codec、7组runtime测试及7份冻结候选覆盖正反包络、深父路径、Time、保留位、独立交换、缓存冲突、预算、关闭。全量CGO=0检查、race及20轮root race通过；7帧go-diameter独立离线解读/精确重编码已记录，不等于现场互通。

本批已结束。下一批为SGd/Gdd OFR/OFA、TFR/TFA普通路径的定义、TPDU/结果分支及成对运行验收；完整S6c/Sh/Ro、SCTP和发布证据按原阶段继续，不重开前置P0。SPEC-APP/VALID/PROFILE/RELEASE等仍Partial/Planned。

### 6.16 第十六批实施记录（2026-11-06）

已完成SGd/Gdd普通IMSI MO及MME/SGSN MT子集，范围/限制见[batch-16](../spec/history/batch-16.json)，机器进度见[progress.json](../spec/progress.json)，前批冻结在[batch-15历史](../spec/history/batch-15.json)。生成器8、ops7、Profile slice.8；24个SPEC ID不变。

- app/sgd生成OFR/OFA、TFR/TFA，Application-ID 16777313；透明TPDU最多200字节，普通IMSI/TBCD、目标号码、Timer/Time、MME ECGI/TAI。选定基础/实验结果、MO cause 3..6与MT 0..2、诊断13、独立E；特殊身份/重投时间/其他诊断及仅SGd OFA实验5001未启用。
- ops7增加可选、编译后的CheckAnswer；静态和共同关联后核对原TFR目标，TFA 5557/MME位置不能用于SGSN。CompleteAnswer在提交前检查，pending解答失败返回ErrInvalidAnswer；独立E不调用正常答案规则。没有通用facts接口、额外缓存或root对应用的依赖。
- 6组codec和6组真实TCP/TLS测试、schema/generator契约检查，覆盖独立exchange、私有原请求、合法重试、一次完成、缓存冲突、预算和取消/关闭。13帧独立离线证据记录旧库默认字典缺EPS、改用其原有XML源字典的执行差异，不冒充网络互通或全协议验证。
- 未授予OFR/TFR TransportReplay；不自动转发/重投、不产生RDR/ALR或管理SC/MWD。应用持久化及业务流程边界不变。24项需求仍Partial/Planned，发布仍not_ready。

下一批为Sh四组命令普通子集与User-Data透明边界的成对codec/运行时实施；现有未完成首版义务保留，不重开前置P0。

### 6.17 第十七批实施记录（2026-11-06）

已完成Sh普通短信AVP子集：UDR/UDA、PUR/PUA、SNR/SNA、PNR/PNA。范围与验收见[batch-17](../spec/history/batch-17.json)，进度见[progress.json](../spec/progress.json)，前批冻结于[batch-16历史](../spec/history/batch-16.json)。生成器8/ops7不变，新增Sh Profile slice.8；24个SPEC ID不变。

- 正常Sh明确要求Vendor10415/AuthApp16777217的VSAI；单Ref24读/更新，单Ref25一次订阅和独立通知；有限普通SIP/TEL/MSISDN身份、所选基础/实验结果和独立E。特性协商、5011、多Ref/退订及私有/特殊身份未启用。
- User-Data只做AVP级字节/presence/预算检查，完整原文保留；不解析XML，不因Report.Complete声称Sh-Data有效或业务成功。有限XML helper及深层规则是下一批任务。Expiry保留原始Time和缺省，不回填SNA、不管理订阅。
- 7组codec与7组真实TCP/TLS测试验证必需字段/15个独立TLV反例、独立PNR早于SNA、四组pending/一次完成、应答重试、缓存冲突、预算及取消/关闭。12帧独立go-diameter离线字典解读/重编码记录为有限证据，不等于XML或现场互通验收。
- 无新增root应用依赖或业务状态；四请求均未授予TransportReplay。现有Full Sh/SMS/Ro/SCTP/发布义务不变，Partial/Planned和not_ready继续有效。

下一批实现Ref24/25有限Sh-Data helper及独立验证，沿已明确范围推进，不重开前置P0。

### 6.18 第十八批实施记录（2026-11-06）

已完成Ref24/25有限Sh-Data构造、显式解析与深层规则，范围见[batch-18](../spec/history/batch-18.json)，实际进度见[progress.json](../spec/progress.json)，前批冻结在[batch-17历史](../spec/history/batch-17.json)。生成器8、ops7、四Profile/线路定义与24个SPEC ID不变。

新增独立app/sh/shdata，显式构造Ref24注册/注销/Pull与Ref25 MME/SGSN/AMF数据。固定无namespace父路径、sequence、Presence、枚举及SCAddress字符限制；接收保留超数字子集文本并报告未解释。允许位置的外部namespace及其他已知Sh子树有界保留raw范围，不按全局同名搜索提取。Parse并非全XSD引擎，Check仅声明选定用途/范围完整；不承接HSS授权、注册/订阅状态或持久化。

Limits先限制输入，随后逐token计深度/元素/属性/文本/opaque；DTD/外部实体和schema取件无入口，单token临时内存由输入硬界限制，不伪称Node自动计费或已获RSS证明。Raw与各切片owned、只读并发；Canonical继续使用原702，不改XML空白/转义/前缀。codec和runtime不隐式解析XML，原AVP-only NoContext/Report不因此扩大。

8组helper测试、codec及真实TCP/TLS集成、fuzz和30项独立原XSD对照进入证据；具体执行结果见本批ledger。全部库实现/常规测试仍纯Go，外部XSD oracle只作离线审计。完整Sh、SMS剩余分支、Ro/RO-06、SCTP及v1发布门槛仍未完成。

### 6.19 第十九批实施记录（2026-11-06）

已完成Sh单Ref11/12/25订阅与退订、SendData可选字段及Ref11/12有限XML helper，边界见[batch-19](../spec/history/batch-19.json)，机器进度见[progress.json](../spec/progress.json)，前批冻结于[batch-18历史](../spec/history/batch-18.json)。仅Sh Profile改为slice.9；生成器8/ops7、其他Profile和24个SPEC ID保持。

710缺失/0/1、712仅Ref25且存在0、SubsReqType0/1与原Expiry分别保存。成功SNA的SHOULD数据不改为MUST，退订不建立本地订阅状态或通知屏障。PNR继续独立交付；取消与晚答沿既有一次完成。Private/特殊身份、多Ref、特性/5011仍未启用，不将单请求支持扩大为混合许可。

shdata新增IMSUserState0..3和SCSCFName有限构造/解析；显式空URI与缺失不同，其他XML文本有界保留并报告未解释，单数据Check不能将额外数据称为完整对应。原702字节仍用于Canonical，XML不隐式进入codec/runtime。15帧离线独立库对照、15项原XSD对照及新增codec/XML/真实TCP/TLS测试进入证据；不冒充现场互通或完整Sh认证。

### 6.20 第二十批实施记录（2026-11-06）

完成有限Sh Supported-Features/Notif-Eff、四种11/12/25混合SNR、5011原请求触发与完整集合校验，以及显式组合通知XML helper。契约见[batch-20](../spec/history/batch-20.json)，前批冻结于[batch-19历史](../spec/history/batch-19.json)，当前执行证据见[progress](../spec/progress.json)。生成器9/ops8，Sh slice.10，其他Profile输入/版本保持slice.8但生成摘要更新；SNR.DataReference是开发期API变化，改为1..3元素切片。

- SPEC-GEN/PROFILE/VALID：编译CheckContext与NoContext互斥，ShPeerFeatures封闭具名工厂；上下文按command pair及两个端终端绑定，最多各4个bitmap集合。未知证据报告NeedsApplication，已知不支持与证据矛盾分别分类，不以DRA CER替代HSS能力。
- SPEC-CODEC/APP：628的M位保留、发送/接收区别、重复Vendor/List拒绝、5011差集及完整应答集合；含25的混合禁止SendData包括0，不建订阅历史。通知XML与Diameter报告分开，不把opaque字节自动认证为混合Sh-Data。
- SPEC-EXCHANGE/INBOUND/CLOSE/BUDGET：出站包络绑定后检查/冻结；入站可交付NeedsApplication，Check成功后冻结且不接受更换，Respond核对答案。检查在状态锁外并占用同一preparer/borrower屏障；关闭不能提前释放额度，Node.Wait等检查退出。沿原逻辑预留，不新增协商队列/存储或Sh自动重传。
- SPEC-DUP：普通已知M位忽略存在一项明确例外：Supported-Features的M位影响5011，必须参与Canonical，防止原身份切换失败前提。已缓存结果及每Arrival来源/包络保持原契约。
- 20帧独立旧库字典对照中，原多数据UDR仍作为未支持控制；Ref24投影明确标识。七种通知组合经原XSD审计；这些不是现场互通或全Sh认证。全库24项ID仍Partial/Planned，release_status仍not_ready。

下一批限定S6c有IMSI的SGSN/MSC/Additional节点成功SRA与SRR成对条件；其他SMS、Sh特殊身份/多数据UDR、Ro/RO-06、SCTP及发布验收继续实施，不重新打开前置规范收尾。

### 6.21 第二十一批实施记录（2026-11-07）

完成普通MSISDN/有IMSI S6c的SGSN、MME、MSC及组合节点，Additional与原SRR GPRS位配对，SRR-Flags/MTI0或1/国际十进制SMEA和成功LMSI/Alert-MSISDN/MWD。详细契约与验收见[batch-21](../spec/history/batch-21.json)，前批冻结于[batch-20历史](../spec/history/batch-20.json)，执行证据见[progress](../spec/progress.json)。生成器9/ops8不变，S6c升slice.9；ServingNode三项MME字段改为Optional是开发期API变化。

- SPEC-PROFILE/VALID/RESULT：五种单节点形状；双节点必须一SGSN、一MSC/MME，且原请求GPRS=1。Supported-Features SMSF bit不能代替GPRS或Gdd注册事实。已有CheckAnswer执行配对，NoContext仅覆盖列明线路及原请求规则，不证明应用的Gdd/VLR数据来源。
- SPEC-CODEC/GEN/OWN：两个独立生成父组、拥有值、Presence、测长和新DecodeBound；SMEA helper只提供1..15位国际十进制格式，其他TON/NPI返回Unsupported。接收合法超子集地址的透明保留/未解释报告仍待实现。
- SPEC-EXCHANGE/INBOUND/DUP/BUDGET：Route仍保留SMEA/flags/原身份；SMS投递Single-Attempt位不是传输重试策略。新字段沿原预算、缓存、私有证据、取消/晚答和一次完成门；没有新增运行时应用状态或持久化。预算扩大由定义上界自然反映，不降低保守系数。
- 20份原候选经独立go-diameter解码/精确重编码，4组codec、helper和4组TCP/TLS验收；旧库SMEA透明字节不构成格式或现场互通证明，相关P1保持开放。全量检查结果记录于progress，24项需求保持Partial/Planned，release_status=not_ready。

下一批限定普通身份的RDR SGSN/MSC结果及对应RDA；无IMSI、IP-SM-GW/SMSF/Router、SGd特殊身份/时间、其他Sh、Ro/RO-06与SCTP继续分批。

### 6.22 第二十二批实施记录（2026-11-07）

完成普通MSISDN RDR的独立MSC/SGSN节点父组，与已有MME形成五种单/组合报告，以及对应RDA验收。详细条件见[batch-22](../spec/history/batch-22.json)，前批冻结于[batch-21历史](../spec/history/batch-21.json)，本次检查见[progress](../spec/progress.json)。S6c升slice.10；生成器9/ops8、Base/SGd/Sh不变，现有Go字段不改类型。

- SPEC-PROFILE/VALID/RESULT：每个已带节点要求自己的cause；MME/MSC互斥，SGSN可与其中之一共存。诊断仅配cause1，MSC零诊断有Presence，各父组选定范围独立；RDA正常2001、实验5001/5558、基础/E与来源关联继续隔离。普通身份要求MSISDN，不以缺字段启用T4/IMS角色。
- SPEC-GEN/CODEC/OWN/DUP：新增两个独立生成父类型；测长、拥有值、界限及Canonical覆盖新叶与父实例；修改任一节点内容或零值Presence均为原身份内容冲突。
- SPEC-INBOUND/EXCHANGE/CLOSE/BUDGET：复用已有冻结、预算、重答、取消/晚答/一次提交；没有新Context或运行时应用状态。现有3584字节结构估计仍涵盖扩展，不调整配额。RDA不证明MWD已存、业务已投递或将发ALR；RDR未获TransportReplay许可。
- 17帧独立对照：16帧精确重编码，三字节诊断在旧go-diameter中变成0/四字节的已知有损行为被明确记录，本库拒绝该坏包。4组codec及4组真实TCP/TLS测试通过；独立字典不能替代节点语义、应用数据来源或现场互通证据。全24项需求仍Partial/Planned，release_status=not_ready。

下一批限定MME/SGSN发起的普通IMSI ALR/ALA及Event/节点条件；其他应用和SCTP沿既定首版范围分批实施。

### 6.23 第二十三批实施记录（2026-11-07）

完成MME/SGSN普通IMSI ALR/ALA、Event位图与新节点条件，具名S6cAlertRole显式端终端证据。详细契约见[batch-23](../spec/history/batch-23.json)，前批冻结于[batch-22历史](../spec/history/batch-22.json)，执行结果见[progress](../spec/progress.json)。生成器9/ops8不变，S6c升slice.11；Sh因错误Context kind隔离修正也升slice.11，线路范围不变。

- SPEC-PROFILE/VALID/APP：Event1/2/3，移动可无已知节点，旧MME/SGSN可跨类型通知新节点；Serving-Node不是发送者身份。ALR.UserIdentifier改为独立AlertUserIdentifier，Optional IMSI/MSISDN互斥；RDR/SRA身份范围未扩大。
- SPEC-BIND/INBOUND/EXCHANGE：接入节点分支必须有绑定Requester/Responder的显式角色Context，未知证据NeedsApplication；原allowlist/唯一端终端目的仍受检。HSS旧MSISDN调用保留零Context，其保证仍限线路切片。Check成功冻结，普通ALA检查Responder；E走原基础包络。
- SPEC-OWN/DUP/CLOSE/BUDGET：复用原私有请求、Check/Respond借用屏障与缓存，Event/节点变更成为原身份内容冲突；没有新队列、路由学习、TFR记录、存储或重投递调度。两种封闭Context使用kind隔离，已有bookkeeping及结构预算覆盖。
- 14帧独立字典精确重编码、4组codec与4组真实TCP/TLS测试覆盖身份/角色、跨节点、结果及关闭竞态。字典接受负控制不能视为语义符合性。24项需求继续Partial/Planned，release_status=not_ready。

下一批限定SGd/Gdd重投时间字段及原TFR配对，包含era/相等条件；Router等其他身份、Ro/RO-06、SCTP及发布验收继续分批。

### 6.24 第二十四批实施记录（2026-11-07）

完成SGd/Gdd直达MME/SGSN的TFR3330/3332与TFA3331、原请求条件及具名SGDRetransmission证据。范围见[batch-24](../spec/history/batch-24.json)，前批冻结于[batch-23历史](../spec/history/batch-23.json)，执行结果见[progress](../spec/progress.json)。SGd升slice.9，Base slice.8、S6c/Sh slice.11、generator9/ops8不变。

- 3330存在必有GMSC地址；3331仅10415/5550且原请求有3330。三字段typed/Presence、重复、Time宽度、TBCD、克隆、计量和Canonical均覆盖；保持已选普通IMSI切片。
- 显式端终端MME/SGSN角色、同PLMN事实、最大/请求era及省电不可达事实；缺事实与已知false分开，不依赖DRA角色推断或时钟服务。旧无3330普通请求可继续零Context，E独立。
- era/seconds字典序比较处理2036回绕及显式零。严格早于通过；相等接收原值+NeedsApplication、完整发送ErrUnsupported；晚于上限报告条件错误。相等规范措辞尚未获得新增裁决，不以更多本地事实伪装关闭该问题。
- 新增公开codec.ValidateResponseContext以复用runtime已有发送检查；CompleteAnswer/Append仍为静态入口。未知答案事实可于Submit留空，保留完整答案后离线补查；Inbound.Check成功即冻结，Respond前需预备拟发答案所需事实。
- 3组codec及4组真实TCP/TLS测试，17帧独立对照（16帧精确往返，1帧记录旧库Time有损差异），全量测试/race/vet/build及证据检查通过。无新队列/状态存储、SMS定时器或TFR自动重放。

24个需求ID仍Partial/Planned，release_status=not_ready；下一批限定T4 OFR/OFA External-Identifier及原请求身份关联。其他Sh/SMS、Ro/RO-06、SCTP及发布验收继续首版内实施，不重复P0。

### 6.25 第二十五批实施记录（2026-11-07）

完成T4 SMS-IWMSC→MTC-IWF的OFR/OFA有限计费身份切片；契约见[batch-25](../spec/history/batch-25.json)，前批冻结于[batch-24历史](../spec/history/batch-24.json)，执行结果见[progress](../spec/progress.json)。SGd升slice.10，其他三个Profile及generator9/ops8不变。

- OFA顶层External-Identifier为10415/3111 UTF8，发送M0；根据29.338复用表纠正旧L092和主流程M1。旧M1候选保留接收兼容证据，新增两份M0投影，不篡改旧摘要。
- T4Charging绑定原IMSI/可选MSISDN Presence、SC和两端Origin，计费用途下正常答案需顶层3111；映射由应用提供。未知映射交付owned答案和NeedsApplication，错误映射拒绝；普通MO零Context不会自动识别T4，E独立。
- RFC4282有限ASCII解析处理转义而非Split，保留原文/大小写；坏UTF8/重复/错误父路径拒绝，子集外合法UTF8接收保留并报告未支持，发送受限。无身份服务、存储、Ro调用或自动重放。
- 3组codec及4组真实TCP/TLS中立运行时测试、12帧独立对照、全量test/race/vet/build及证据校验通过；11帧精确往返、错误App被拒绝，旧字典接受的语义负例不计通过。T4 SCTP部署仍待实施验收，不用TCP/TLS运行测试替代。

24个SPEC ID仍Partial/Planned，release_status=not_ready。下一批限定S6c T4无MSISDN的RDR/RDA与HSS ALR/ALA身份分支；Ro/RO-06、SCTP及发布门槛继续首版内推进。

### 6.26 第二十六批实施记录（2026-11-07）

完成S6c的T4无MSISDN RDR/RDA与HSS ALR/ALA；见[batch-26](../spec/history/batch-26.json)、[前批冻结](../spec/history/batch-25.json)及[progress](../spec/progress.json)。S6c升slice.12，其他Profile、generator9/ops8不变。

- RDR采用独立ReportUserIdentifier，Optional IMSI/MSISDN互斥；普通MSISDN构造需改用avp.Some。SRA/RDA的UserIdentifier语义不变，RDA成功可选Alert-MSISDN不等于IMSI回显。
- T4NoMSISDN明确报告/提醒用途，绑定原IMSI、SC和终端；无证据报告NeedsApplication。HSS IMSI提醒不强加Event，MME/SGSN角色仍要求Event，无法从缺Event推断用途。库不验证T4历史或存储MWD。
- 3组codec、4组真实TCP/TLS测试及10份独立字典精确往返通过。D21原IP-SM-GW报告保留未支持控制，新增MME投影；旧字典接受双/空身份不算语义通过。全量test/race/vet/build及证据审计通过。

24项SPEC需求仍Partial/Planned，release_status=not_ready。下一批限S6c T4无MSISDN SRR/SRA的已知IMSI和原请求/节点条件；无IMSI不确定成功身份、IMS/Router、Ro/RO-06、SCTP和发布门槛继续保留。

### 6.27 第二十七批实施记录（2026-11-07）

完成S6c T4无MSISDN的SRR/SRA；见[batch-27](../spec/history/batch-27.json)、[前批冻结](../spec/history/batch-26.json)及[progress](../spec/progress.json)。S6c升slice.13；其余Profile、generator9/ops8不变。

- SRR顶层UserName/MSISDN改为Optional互斥身份，普通调用点需用avp.Some。新增T4Routing用途，复用原kind5的终端/IMSI/SC证据；无额外Context容量或root应用依赖。
- 成功SRA的IMSI与原请求已知IMSI相等；Additional及GPRS/SMEA关系沿原规则。SMSMI、IMS/Router及未知IMSI成功仍未支持；结果范围不扩张。
- 3组codec、5组真实TCP/TLS测试验证完整用途校验、私有快照、重答、预算、关闭和固定同终端双attempt的原身份/晚答。13份独立字典精确往返，不把旧字典接受反例当语义认证。
- 全量test/race/vet/build和证据审计通过；SRR/SRA由NoContext改为CheckContext，移除校验许可的反向测试继续保留。纯Go、无持久化；SCTP部署及整体v1未完成。

下一批限定SRA实验5550/5552/5556/5557及身份/诊断条件；24项SPEC需求仍Partial/Planned，release_status=not_ready，Ro/RO-06和SCTP继续属于首版。

### 6.28 第二十八批实施记录（2026-11-07）

完成SRA实验5550/5552/5556/5557及5550三域诊断；见[batch-28](../spec/history/batch-28.json)、[前批冻结](../spec/history/batch-27.json)及[progress](../spec/progress.json)。S6c升slice.14，其余Profile与generator9/ops8不变。

- 增加Optional U32的MME/MSC/SGSN顶层诊断，零与缺失分开；0..255接收保留，保留/未选择域值通过Report报告未支持，本端发送只准已选域集。其他结果带诊断仍为切片外。
- RoutingRole(kind6)绑定GMSC/IP-SM-GW与HSS；5550的IMSI义务按外部UNRI核对，绝不把SRR-Flags当UNRI。未知必要事实交付owned答案和NeedsApplication，应用可离线补查；不保存UNRI/MWD或做重试。
- 4组codec、5组真实TCP/TLS、全量test/race/vet/build及证据审计通过。21份独立帧中20份精确往返，旧字典接受三字节诊断且补宽重编码单独记录，不计合规。原缺SMEA请求保留控制，另补明确正向投影。

下一批限定SGd OFA实验5001的接口/原MSISDN边界。24项SPEC维持Partial/Planned、release_status=not_ready；IMS/Router、Ro/RO-06、SCTP和发布验收继续属于首版后续工作。

### 6.29 第二十九批实施记录（2026-11-07）

完成普通已知IMSI的SGd OFA实验5001、SGd/Gdd接口边界及原MSISDN关联；见[batch-29](../spec/history/batch-29.json)、[前批冻结](../spec/history/batch-28.json)及[progress](../spec/progress.json)。SGd升slice.11，其余Profile与generator9/ops8不变。

- MOForwarding(kind7)显式绑定接口、原IMSI/MSISDN presence/SC和实际终端；复用有界身份存储，Context≤2KiB不改。不能由共同应用ID或DRA推导接口。
- 实验5001完整校验仅支持SGd且原请求含MSISDN；Gdd/T4接收owned答案并报告未支持，零Context报告NeedsApplication。缺MSISDN的原请求关系失败，不可以外部事实补成合法。基础5001/E仍独立。
- 3组codec、4组真实TCP/TLS和全量test/race/vet/build通过。8份新增独立帧均由旧字典精确往返；反例的旧字典接受行为不计语义符合。没有SCTP部署或现实业务事实认证。

下一批限定S6c已知IMSI成功SRA的IP-SM-GW节点/Realm、父路径及原SRR条件。24项SPEC仍Partial/Planned、release_status=not_ready；Ro仍属v1，后续部署/压力验收不以本批完成替代。

### 6.30 第三十批实施记录（2026-11-07）

完成已知IMSI成功SRA的单IP-SM-GW节点；见[batch-30](../spec/history/batch-30.json)、[前批冻结](../spec/history/batch-29.json)及[progress](../spec/progress.json)。S6c升slice.15，其余Profile与generator9/ops8不变。

- ServingNode/AdditionalServingNode增加3100/3101/3112/3348四叶。标准发送仅直接Serving的N/NH/NB/NHB、SBI0/1；Additional同名叶供接收保留，不能继承Serving的组合。原SRR Additional许可及T4原IMSI关联保留，ALR不扩大。
- Realm、保留SBI、Additional-IPSMGW及本切片未支持的跨节点组合接收保留并报告ErrUnsupported，不是缺应用证据；字段格式/次数/号码缺项独立拒绝。未启用s6c-ip-sm-gw-realm-v1，不能用Context替代固定Profile兼容。
- 4组codec、6组真实TCP/TLS、全量test/race/vet/build及新增叶级codec/race通过。18份独立帧往返；旧字典把Realm当OctetString、SBI当Unknown，记录为差异，不认证其语义或现场互通。生成资源计量包含新字段，无新Context/状态/业务回调。

下一批限定普通MSISDN RDR/RDA的IP-SM-GW投递结果父组、原因与已选诊断。24项SPEC仍Partial/Planned、release_status=not_ready；Realm标准争议/兼容、Ro/RO-06及SCTP部署继续保留各自门槛。

### 6.31 第三十一批实施记录（2026-11-07）

完成普通MSISDN RDR/RDA的IP-SM-GW投递结果，见[batch-31](../spec/history/batch-31.json)、[前批冻结](../spec/history/batch-30.json)及[progress](../spec/progress.json)。S6c升slice.16，其他Profile、generator9/ops8及七种具名Context不变。

- 3316/3320组及3321/3322叶采用独立父路径；本子集每个结果父组都有cause，选0/1/2，IP组仅cause1可带诊断11/12。四节点组合仍排除MSC+MME，Presence不补默认值，RDA2001/10415-5001/5558与E包络沿用原关联规则。
- 未支持的四字节cause/0..255诊断可静态保留，完整报告ErrUnsupported；诊断13只是未选分支，不声称规范禁止。宽度、重复、缺cause、>255及cause0/2带诊断独立拒绝。当前运行时对语义Rejected请求关闭并释放预算，不交付Inbound；不套用已关联答案的报告交付契约。
- 旧T4候选可解码新typed组，但完整校验/发送仍未支持；IMS/SMSF、MWD/SC存储及投递调度未扩大。3组codec、5组真实TCP/TLS、全量test/race/vet/build通过；16帧独立旧字典往返，反例字节保留不算语义通过。

下一批限定S6c IMS RDR/RDA的SMSMI-Correlation-ID、HSS-ID/Destination-SIP-URI父路径及显式身份/用途绑定，不扩无IMSI成功、IMS OFR/TFR或Router改写。24项SPEC仍Partial/Planned、release_status=not_ready；Ro/RO-06仍在首版，SCTP部署及独立互通/压力仍待验收。

### 6.32 第三十二批实施记录（2026-11-07）

完成S6c IMS RDR/RDA的HSS-ID、SMSMI目的URI及显式用途/终端绑定，见[batch-32](../spec/history/batch-32.json)、[前批冻结](../spec/history/batch-31.json)及[progress](../spec/progress.json)。S6c升slice.17；其他Profile、generator9/ops8不变。

- 新增RDR专用3324/3327 typed组；HSS-ID仍在3102/1，数字6..15，关联组禁止3325/3326。发送URI仅≤256字节ASCII sip:user@DNS子集，原字节参与绑定和Canonical；UTF8接收不等于完整SIP支持，未选范围显式报告。
- 第八种封闭Context IMSDeliveryReport绑定HSS-ID、URI、SC和真实终端，不做HSS解析、身份分配或业务历史核实。与T4复用有界内部存储但保持类型隔离，新增及旧Context最大证据均≤2KiB；无新运行时队列或重放权限。
- 4组codec及5组真实TCP/TLS通过，覆盖NeedsApplication/Check冻结、原请求关联、预算/多Arrival/冲突、取消晚答与一次完成。20帧独立旧字典精确往返；其3325 OctetString及语义负例容忍度作为差异记录，不当作符合性证据。全量test/race/vet/build及证据审计通过。

下一批限定S6c IMS ALR/ALA的dummy MSISDN、目的URI及HSS→IWMSC绑定；IMS SRR成功身份、OFR/TFR、Router和业务调度不并入。Ro/RO-06仍属首版，24项SPEC仍Partial/Planned、release_status=not_ready。

### 6.33 第三十三批实施记录（2026-11-07）

完成S6c IMS ALR/ALA的dummy MSISDN、目的URI和HSS→IWMSC用途绑定，见[batch-33](../spec/history/batch-33.json)、[前批冻结](../spec/history/batch-32.json)及[progress](../spec/progress.json)。S6c升slice.18，其他Profile、generator9/ops8不变。

- ALR专用3324/3327 typed组；身份恰15零MSISDN，不含User-Name/HSS-ID或源URI，MME/SGSN Event/节点不能混入。选定ASCII URI≤256字节，沿用原字节比较；Time原值及presence保留，不猜era。
- 第九种Context IMSHSSAlert绑定HSS/IWMSC、URI和SC，与报告/T4/普通AlertRole隔离。复用内部存储不扩大Context容量，既有2KiB预算继续通过；报告和提醒不共享pending或自动触发。
- 4组codec及5组真实TCP/TLS、全量test/race/vet/build及证据审计通过；20份独立帧与旧字典精确往返。反例容忍和Time转换不认证应用语义/可达性；无新持久化、MWD或调度职责。

下一批限定IMS SRR/SRA请求HSS-ID/URI和已知真实IMSI答案，复核SRR重传许可；真正无IMSI成功仍未支持，OFR/TFR和Router分开。24项SPEC仍Partial/Planned、release_status=not_ready，Ro/RO-06仍属首版。

### 6.34 第三十四批实施记录（2026-11-07）

完成IMS SRR/SRA请求HSS-ID、目的URI及角色/已知HSS绑定，见[batch-34](../spec/history/batch-34.json)、[前批冻结](../spec/history/batch-33.json)及[progress](../spec/progress.json)。S6c升slice.19；Base/SGd/Sh、generator9/ops8不变。

- SRR专用3324/3327 typed组，顶层User-Name为HSS-ID；T4仍要求答案原IMSI相等，IMS答案真实IMSI由HSS提供，不要求与HSS-ID相等或不等。库不证明订户分配、不回显HSS-ID造成功身份，请求方不必预先知道IMSI。
- 第十种Context IMSRouting绑定角色/HSS-ID/URI/SC/终端及可选IP-SM-GW UNRI，5550缺身份保持Ready/NeedsApplication/Rejected的分支。复用已有内部存储，2KiB预留和旧证据边界不增；公共Check冻结及完成报告沿用。
- 两种固定同HSS活动请求恢复策略补测原身份、URI/Context、每attempt答案预算、旧attempt晚答胜出和一次完成。DRA路径可换、HSS不可换，无新持久化/重连自动重放/业务重试。
- 4组codec、6组真实TCP/TLS、全量test/race/vet/build及审计通过；22份独立帧旧字典精确往返，数字同字节及负例接受仅是语法边界，不能认证无IMSI成功。

下一批限定IMS OFR/OFA的dummy源身份、SMSMI三叶、所选结果与IP-SM-GW→IWMSC用途；TFR及业务调度分开。24项SPEC仍Partial/Planned、release_status=not_ready，Ro/RO-06仍属首版，不重复前置P0。

### 6.35 第三十五批实施记录（2026-11-07）

完成IMS OFR/OFA的dummy源身份、SMSMI三叶及临时IP-SM-GW结果，见[batch-35](../spec/history/batch-35.json)、[前批冻结](../spec/history/batch-34.json)及[progress](../spec/progress.json)。SGd升slice.12；Base/S6c/Sh和generator9/ops8不变。

- UserIdentifier.UserName改Optional，普通MO仍要求IMSI；本IMS子集恰15零MSISDN且无IMSI。3324必需3325/3326/3327；3316必需3320及cause0/1，cause1所选诊断11/12；不由OFA结果倒填请求Outcome。
- 第十一种Context IMSMOForwarding绑定IP-SM-GW/IWMSC、HSS-ID、双URI摘要和SC。原URI保留于请求，Context增加32字节仍满足既有2KiB上界；有界身份语法抽取到内部共享helper，应用包不互相依赖。
- 4组codec和5组TCP/TLS、全量test/race/vet/build与证据审计通过，验证冻结/重复冲突/关闭/晚答和禁止Route恢复。31帧旧字典精确往返，旧SGd3325为UTF8，与旧S6c字典差异分别记录。

下一批限定IMS TFR/TFA身份与IP-SM-GW目标；其他IMS MO组合、业务历史/调度不并入。24项SPEC仍Partial/Planned、release_status=not_ready，Ro/RO-06仍属首版。

### 6.36 第三十六批实施记录（2026-11-07）

完成IMS TFR/TFA顶层HSS-ID、两叶SMSMI及SMS-GMSC→IP-SM-GW用途，见[batch-36](../spec/history/batch-36.json)、[前批冻结](../spec/history/batch-35.json)及[progress](../spec/progress.json)。SGd升slice.13；Base/S6c/Sh及generator9/ops8不变。

- TFR UserName在SMSMI存在时为6..15位HSS-ID；3324只含必需3326/3327，不接受组内3325或MME/SGSN号码。Timer3306/Start3307保留可选原32位值；3330/3331/3332不扩到本IMS用途。
- 第十二种Context IMSMTForwarding绑定GMSC/IP-SM-GW、HSS-ID、SC及域隔离的双URI摘要，Context仍704字节/≤2KiB；MO/T4/重投证据不能互换，无新增队列或存储。
- 4组codec与5组真实TCP/TLS覆盖成对结果、身份/父路径/时间、冻结/重复/预算/取消/晚答/关闭及禁止Route恢复。31帧通过旧库精确往返，D21三帧原摘要相符；完整执行记录见progress。

下一批限定IMS MO的可用IMSI及可选真实MSISDN分支。其余首版需求不因本批完成而关闭；24项SPEC仍Partial/Planned、release_status=not_ready，Ro/RO-06仍属首版。没有自动OFR→TFR、业务关联表或持久化重放。

### 6.37 第三十七批实施记录（2026-11-07）

完成IMS MO可用IMSI及可选真实MSISDN的值/presence绑定，见[batch-37](../spec/history/batch-37.json)、[前批冻结](../spec/history/batch-36.json)及[progress](../spec/progress.json)。SGd升slice.14，其他Profile及generator9/ops8不变。

- IMSMOForwardingEvidence新增两项Optional源身份；两项缺省保留dummy分支，有IMSI时不使用dummy。目的HSS-ID仍在SMSMI父组，源IMSI不从其推导；真实号码保留前导零。
- 沿kind11复用闲置源身份存储槽，Context仍704字节，新增分支最大1776字节，既有2KiB预算不增。缺省残留Value不复制；MO/T4/MT用途不互换。
- 3组新增codec测试、原5组TCP/TLS扩展为三种源身份共30个网络组合；21份新独立TLV帧及2份原答案通过旧库23项精确往返。5001缺原MSISDN为关联失败，有MSISDN仍报告IMS用途未支持；均无自动重试或二次完成。

下一批限定S6c T4无MSISDN投递报告的IP-SM-GW结果。24项SPEC仍Partial/Planned、release_status=not_ready；Ro/RO-06仍属首版，无持久化、业务存储或自动OFR→TFR。

### 6.38 第三十八批实施记录（2026-11-07）

完成S6c T4无MSISDN RDR/RDA的IP-SM-GW结果，见[batch-38](../spec/history/batch-38.json)、[前批冻结](../spec/history/batch-37.json)及[progress](../spec/progress.json)。S6c升slice.20，Base/SGd/Sh与generator9/ops8不变。

- 复用3320/3321/3322及既有cause0/1/2、cause1诊断11/12规则，解除T4身份限制；支持IP单组及五种与MME/MSC/SGSN的合法组合，仍要求每个父实例自己的cause，MSC/MME互斥。
- T4DeliveryReport绑定GMSC/HSS与原IMSI/SC，零Context仍NeedsApplication；不从投递节点或主机名推断请求方角色。Context种类/704字节结构及2KiB预算不变，无存储/调度/RDR重放。
- 4组codec测试及普通/T4共20个TCP/TLS网络组合，覆盖结果/父路径/用途/预算/冻结/取消/关闭；20份新独立帧加3份D21原帧通过23项旧库精确往返，历史字节保持不变。

下一批转入标准SMS IEC DIRECT_DEBITING CCR/CCA基础codec及成对验收。六模板、私有字段、RO-06、剩余SMS/Sh、Realm兼容、SCTP及现场验收仍属首版；24项SPEC继续Partial/Planned、release_status=not_ready。

### 6.39 第三十九批实施记录（2026-11-07）

完成[Ro标准SMS IEC DIRECT_DEBITING基础](../spec/history/batch-39.json)，新增Ro slice.1及app/ro，前批冻结于[history/batch-38](../spec/history/batch-38.json)。generator10新增精确u64；ops8与Base/S6c/SGd/Sh定义不变，全部生成摘要因生成器版本更新。

- App4/272、CC-Type4/Number0/Action0、标准SMS上下文及一个MSCC服务；RSU/GSU417完整uint64，0/缺项/空组区分，2022退款凭据owned原字节。选IP-SM-GW/Submission执行前子集，不宣称完整sms-fw兼容。
- 成对2001/4012/既有基础失败/E，CC-Type/Number/服务键与公共请求关联；顶层失败与子级成功分别保留，Ready不表示可以扣费。子级Result-Code缺省只保留并报告未支持，不填2001。
- 41份独立完整输入（累计455wire/17XML），35份旧库App4精确往返；错误宽度只作为本库反例。4组TCP/TLS验证双向消息、准备重试、错误关联一次完成、取消晚答和资源释放。纯Go测试/构建、race、依赖/摘要审计见[progress](../spec/progress.json)。
- 无新Context/持久化/计费会话或重试引擎；CCR无TransportReplay许可。ECUR、退款、六模板/私有420绑定、5011、RO-06及现场OCS继续首版后续验收；不关闭全局SPEC项。

下一批限定标准SMS ECUR INITIAL/TERMINATION及原会话/序号配对，避开未裁决SMS失败分支。24项SPEC仍Partial/Planned，release_status=not_ready。

### 6.40 第四十批实施记录（2026-11-07）

完成[标准SMS ECUR INITIAL/TERMINATION](../spec/history/batch-40.json)，Ro升slice.2，generator10/ops8不变，其他Profile源码/生成输出不变。前批原进度冻结于[history/batch-39](../spec/history/batch-39.json)。

- 无UPDATE的INITIAL1/0、TERM3/1；USU446/417和直接MSCC的FINAL872，0/缺项/空组区分。TERM2001无需GSU，5002不自动重建INITIAL，初始授权不等同最终扣费完成。
- 纯`ro.CheckECURContinuation`检查调用方两个完整请求的Session、端点、服务键及序号；不保存计费会话，不证明初始授权/投递/额度，不替代每个CCA与原请求的关联。当前exchange Report与跨请求helper范围分别声明。
- 原INITIAL中的顶层IMEISV16和PS的RAT/ULI/时区子集已实现；已知错父路径/未支持单位或SMS失败/在线序号不因M0放行。第39批SMS-Node1文字名称更正为IP-SM-GW，原数值和字节不变。
- 45份独立输入（累计500wire/17XML）、44份旧App4字典精确往返；5组codec、3组TCP/TLS及全量纯Go/race等证据见[progress](../spec/progress.json)。错误宽度不计第三方通过，未取得现场OCS证据。

下一批限定sms-fw IEC单模板具名兼容及私有Vendor/父路径/420绑定。其他模板、SCAP1、5011、RO-06及部署验收仍属首版；24项SPEC维持Partial/Planned，release_status=not_ready。

### 6.41 第四十一批实施记录（2026-11-07）

完成[sms-fw IEC单模板具名兼容基础](../spec/history/batch-41.json)。前批进度冻结于[history/batch-40](../spec/history/batch-40.json)。新增`compat/smsfwro`静态Profile，完整源模板42节点、显式输入builder、旧STRING/int32/二实例Subscription与复制关系；标准Ro规则不放宽，同App4两Profile在Compile时冲突。

正常CCA选择成功/4012/基础失败及独立E，顶层可选GSU、DDFH保留缺省/0/1；每个答案关联原请求。未知私有字段不猜定义。本IEC没有420或1292，Grouped420退款绑定移至下一独立模板，不能把未出现字段记成已实现。源投影、旧库字节往返、真实TCP/TLS为三类有限证据，均不等于现场OCS接受或旧C++执行结果。

generator11新增i32，ops8、12种Context及预算机制不变；五份原定义未改、六输出重新生成。31新wire累计531wire/17XML，29旧库对照，4组codec/3组TCP-TLS及全量检查见[progress](../spec/progress.json)。完整六模板/SCAP1、私有答案证据、5011与RO-06仍按原首版义务；24项SPEC仍Partial/Planned，release_status=not_ready，不扩大全量规范审查范围。

### 6.42 第四十二批实施记录（2026-11-07）

完成[旧退款固定Grouped420基础](../spec/history/batch-42.json)，前批冻结于[history/batch-41](../spec/history/batch-41.json)。独立`compat/smsfwro/legacy420` Profile选择EVENT4/Number0/Action1，完整26模板节点、精确i64金额/i32指数及旧币种、uint64单位、显式顶层0/1292 token；金额由应用给值，不将原模板常量当缺省。

固定420例外仅在CCR/456/437，其他父路径/413别名/标准420及未知私有字段拒绝；与标准Ro、IEC-only的同Bundle编译冲突。正常CCA无GSU也可成功，关联本次请求；E独立。没有原账本、资格推断、持久化或自动恢复。未知CCA token/PSP及原扣费映射仍需真实证据，不能由请求定义反推。

generator12/ops8，七输出刷新、六原定义未改；38新wire累计569wire/17XML，33份使用明确离线字典覆盖的旧库往返（原标准字典失败），4组codec/3组TCP-TLS与全量检查见[progress](../spec/progress.json)。不是标准420兼容、C++执行输出或OCS抓包。其余四模板、同域组合、SCAP1、5011、RO-06和现场验收仍在原首版范围；24项SPEC维持Partial/Planned，release_status=not_ready。

### 6.43 第四十三批实施记录（2026-11-07）

完成[旧ECUR INITIAL/TERMINATION基础](../spec/history/batch-43.json)，前批冻结于[history/batch-42](../spec/history/batch-42.json)。smsfwro slice.2在IEC同一固定类型/Bundle上增加1/0→3/1无UPDATE、RSU/USU/FINAL、原接口2/3、opaque状态及放行Time；新增显式纯构造和跨请求检查，不保存INITIAL、不验证grant/业务投递或自动发送TERM。

真实源模板分别41/44节点。TERM的MMS实际1210→1212顺序保留为原始证据，静态sender仍1212→1210，测试将差异限定于这一交换；不能声称旧TERM逐字节发送或现场OCS接受已完成。三项开发字段改Optional，NewIEC签名/原帧保持；完整迁移说明见本批契约。

generator12/ops8及其他六定义/输出不变，49新wire累计618wire/17XML，47原App4字典对照、4组codec/3组TCP-TLS与全量验收见[progress](../spec/progress.json)。四个旧模板构造基础已实现，通用模板/SCAP1、同域退款组合、更多CCA/私有答案/5011、RO-06和目标验收仍在原首版义务内。24项SPEC维持Partial/Planned，release_status=not_ready。

### 6.44 第四十四批实施记录（2026-11-07）

完成[通用CCR及直接费用/余额结果](../spec/history/batch-44.json)，前批冻结于[history/batch-43](../spec/history/batch-43.json)。legacy420 slice.2共用固定413/420类型，增加通用32节点、显式NewDirectDebit/NewBalanceCheck、topCost/GSU/余额结果及原Action关联。请求signed币种与答案unsigned币种分父路径，缺省指数与显式零分开。原退款Money/token改Optional；NewRefund原字节保持，历史GSU反例升级为新增已支持结果。

五CCR模板基础已有实现，但两个Profile尚未共域：通用Client-Address2018是RFC Address，旧IEC/ECUR是STRING。固定金额420解决方案不等于完整类型组合已完成；组合包必须补此绑定，不按Action/长度猜型。SCAP1、未知CCA token/PSP、5011、RO-06及目标验收仍在原v1义务。库不把余额查询、2001或缺省金额当业务授权，不保存账本/重启记录。

55新wire累计673wire/17XML，47旧库原字典离线对照、3组codec/4组TCP-TLS及全量检查见[progress](../spec/progress.json)。generator12/ops8和六份其他定义/输出不变，无新增Context种类/预算类。24项SPEC维持Partial/Planned，release_status=not_ready。

### 6.45 第四十五批实施记录（2026-11-07）

完成[SCAP1源模板离线基础](../spec/history/batch-45.json)，前批冻结于[history/batch-44](../spec/history/batch-44.json)。compat/scap1 slice.1固定App0/271及内部193/19302、47节点、11保序607父实例、显式Getter结果输入；正常ACA为参考基础语法的合成子集及原记录关联，不冒充实际私有ACA。缺宏与空串/零Time分开，无S1/时钟/账本/重试业务。

本批没有开放SCAP1运行时。现有App0控制归类和能力协商需下一批单独绑定，测试明确拒绝未审核业务接纳且不自动广告；旧双Auth/Acct VSAI冲突必须保留具名门槛。六模板构造基础不等于全部六模板互通完成，SCAP1 runtime、五CCR同域2018、私有答案/5011、RO-06及目标证据仍在原v1范围。

61新wire累计734wire/17XML，56份显式离线字典对照、4组codec/1组运行时边界测试及全量回归见[progress](../spec/progress.json)。八份生成输入、原七份定义/输出不变，generator12/ops8不变；24项SPEC维持Partial/Planned，release_status=not_ready。

### 6.46 第四十六批实施记录（2026-11-07）

完成[SCAP1 SingleAttempt运行时](../spec/history/batch-46.json)，前批冻结于[history/batch-45](../spec/history/batch-45.json)。每方向显式ApplicationBinding=SCAP1Accounting，头0/271与严格Vendor193/Acct19302分开，固定目标Host/Realm及答案来源；配置Profile不会自动广告。旧双Auth/Acct VSAI仍拒绝，这是具名兼容启用条件而非通用Vendor应用命名空间规则。

显式配置的App0/271进入普通预算/pending/Inbound路径，CER/DWR/DPR不变；正常/E、原记录关联、方向隔离、混合S6c能力、预算、取消晚答、关闭、SingleAttempt Route和重连能力清除均有真实TCP/TLS证据。混合Peer可Ready而SCAP1仍ErrNotReady；重连不自动重放。当前入站验收限定NoCache，不宣称财务去重或恢复。非法双字段CER尚无完整5009握手诊断，旧广告、Relay-only DRA和实际私有ACA/OCS证据仍未覆盖。

十组新配置/网络测试、全量纯Go test/vet/build/race及四组20次race通过；56份显式字典对照再次通过。scap1 slice.2仅更新启用边界元数据/摘要，wire规则及其他七份定义/输出不变；734wire/17XML、generator12/ops8及十二Context种类不变。下一步五CCR固定组合2018/TERM顺序，随后私有答案/5011及RO-06；24项SPEC仍Partial/Planned，release_status=not_ready，见[progress](../spec/progress.json)。

### 6.47 第四十七批实施记录（2026-11-07）

完成[五CCR固定组合](../spec/history/batch-47.json)，前批冻结于[history/batch-46](../spec/history/batch-46.json)。新增compat/smsfwro/combined slice.1的单一App4/272类型/操作集，覆盖五源模板及通用余额用法；原八Profile/生成输出不变。2018固定owned线路字节，不按Action/长度切换decoder；显式构造器保留旧STRING或固定普通Address载荷，完整模板规则拒绝混合形态。420仍固定Grouped，413分键；CCA联合字段必须按原请求约束，不自动扩大各模板答案许可。

本地发送采用单一确定顺序，明确记录TERM1210/1212、通用2018及439/10425:23的次序差异；原模板字节/节点证据保留。接收和Canonical保持既有契约，目标顺序容忍/2018格式接受仍未证明，不承诺旧C++或现场字节级等价。

36新wire累计770wire/17XML；四组codec、三组真实TCP/TLS、36显式字典对照和全量纯Go test/vet/build/race通过。九生成定义、generator12/ops8、十二Context种类，运行时队列/class/重放许可不变。显式Continuation不保存业务会话。下一批实际私有答案/结果依赖，随后RO-06；24项SPEC仍Partial/Planned、release_status=not_ready，见[progress](../spec/progress.json)。

### 6.48 第四十八批实施记录（2026-11-07）

完成[旧CCA分层结果、费用与DDFH](../spec/history/batch-48.json)，前批冻结于[history/batch-47](../spec/history/batch-47.json)。三个兼容CCR Profile均新增单个无服务键的结果专用MSCC，顶层与子结果分别保留，缺省不补成功；新增4010/4011/5030分支。IEC/ECUR可交付所选成功费用，通用直接扣费可交付DDFH；原余额/退款/ECUR不能借联合字段获得直接扣费策略。当前仅报文字段，无策略执行/持久化/自动业务。

源码token接收只按可配置code匹配而缺Vendor/父实例规则，PSP逐层Vendor和省略语义未证；保持未知私有字段拒绝，不复制旧截断、结果覆盖、缺省operation_result1或空答案假成功。2002、5031完整诊断、多MSCC/服务键/逐服务grant和5011仍未完成；下一批限定5031/5011结果安全，随后RO-06，外部证据只阻止相应目标兼容声明。

smsfwro/legacy420 slice.3、combined slice.2，其他六定义/输出和generator12/ops8不变。67新wire累计837wire/17XML，66原字典对照（无覆盖）、三组codec及两组真实TCP/TLS和全量纯Go test/vet/build/race通过。九定义/十二Context种类/207生产依赖不变；24项SPEC仍Partial/Planned，release_status=not_ready，见[progress](../spec/progress.json)。

### 6.49 第四十九批实施记录（2026-11-07）

完成[Ro 5031诊断与5011安全接收](../spec/history/batch-49.json)，前批冻结于[history/batch-48](../spec/history/batch-48.json)。标准Ro及三个旧CCR兼容Profile支持命令级5031/必需Failed-AVP与基础5011；正常CCA保持必需268和可选实验297、支持列表628。顶层2001+实验10415/5011交付原值并报告ErrResultConflict；其他实验命名空间/支持列表的语义报告未支持。缺268、坏结构/关联仍为无效答案，不因接收失败自动重试。

标准自动配对MUST继续拒绝；构造/编码/Respond不接受本批仅供接收的实验结果和特性列表，E实验配对亦拒绝。命令级5031的诊断引用保持有界owned，正常CCA残缺引用不报告完整语义通过。单诊断容器、子MSCC5031、完整特性协商与rating语义不在本批；没有持久化/业务放行/自动重放。

generator13新增固定E规则入口，ops8与十二Context种类不变；标准Ro slice.3、smsfwro/legacy420 slice.4、combined slice.3，其他五定义不变、九输出刷新。65新wire累计902wire/17XML、63原字典对照、三组codec/三组真实TCP-TLS及全量纯Go test/vet/build/race通过。下一批RO-06固定同OCS活动exchange恢复；24项SPEC仍Partial/Planned，release_status=not_ready，见[progress](../spec/progress.json)。

### 6.50 第五十批实施记录（2026-11-07）

完成[普通R33的Ro同OCS活动请求恢复](../spec/history/batch-50.json)，前批冻结于[history/batch-49](../spec/history/batch-49.json)。标准Ro slice.4、smsfwro/legacy420 slice.5、combined slice.4为已选CCR启用固定TransportReplay；原动作/序号/字段校验不变。复用Route的ExplicitRetransmit或TransportFailover、RequestInfo/Retransmit及总期限/预算/一次完成，不新增API或根运行时状态。其他五定义/输出、generator13/ops8及十二Context种类不变。

恢复MUST受固定Destination/唯一终端来源、ReplayDomain声明、2..16次及原恢复年龄限制；部署去重证据仍须另验。原E2E/Session/CC字段/内容及Profile不变，HBH按连接分配，T按累计发送证据。有效正常/E答案及冲突报告终结exchange；业务Tx或DDFH不授权恢复，Peer直接提交仍SingleAttempt。匹配错误、取消/关闭不能触发业务重试。

27新wire累计929wire/17XML；9原字典和18既有显式覆盖对照、13组根测试（含12组TCP/TLS）及2组codec、全量纯Go test/vet/build/race通过。覆盖旧/新答案竞争、每attempt A、reader/writer屏障、原身份、反压/幂等及重连REOPEN。下一批原请求记录/标识准入，随后Tr/跨段/重启等专用RO-06；当前不关闭覆盖§38的16组或目标OCS验收。24项SPEC仍Partial/Planned，release_status=not_ready。

### 6.51 第五十一批实施记录（2026-11-07）

完成[Ro原请求准备与本地标识保留](../spec/history/batch-51.json)，前批冻结于[history/batch-50](../spec/history/batch-50.json)。`Node.PrepareRoRecoveryRequest`仅接纳四Ro Profile的已审核CCR，在scratch/恢复槽配额内冻结原帧、全部目标绑定、Bundle指纹/空Context和恢复上限；所有目标检查后才发原E2E。HBH=0/T=0为准备占位，不选连接、不创建pending/完成、不写网络。原Origin-State及所有非目的字段不改写，Proxy/Route-Record拒绝。

Node仅保存固定摘要guard到ReplayUntil+ReuseGuard；默认及自定义源都不能将该E2E发给普通请求。`CheckIntegrity`只查版本/结构/期限关系/摘要，不能当成当前Profile、原Origin授权、恢复年龄、累计额度、OCS去重或存储确认。配置和有界owned表示见本批契约；Node.Wait包含准备借用者，Node关闭不等待恢复期到期。

十二组测试含四Profile/15用途与一组真实TCP/TLS，覆盖长目标、所有权、校验、配额、发号碰撞/回绕、预检、取消/关闭和source并发。全量纯Go test/vet/build、race及生成检查通过。九定义/输出、generator13/ops8、十二Context、929wire/17XML不变。下一批有限`SubmitRoRecovery`段接纳及累计额度高水位，随后Tr/晚答/续发/重启；完整导入/RecoverySpace和RO-06目标验收仍未完成。24项SPEC继续Partial/Planned，release_status=not_ready。

### 6.52 第五十二批实施记录（2026-11-07）

完成[Ro有限段接纳及累计额度](../spec/history/batch-52.json)，前批冻结于[history/batch-51](../spec/history/batch-51.json)。SubmitRoRecovery只接受本Node准备的原记录及第一个OCS目标，同段使用已有Route显式传输恢复；允许一次attempt段，以及本地已结清的无答续段。版本/指纹/Origin/全部绑定/最长目标/期限/guard/额度须重新验证，原E2E不重新发号，HBH在当前连接分配。

SegmentNo及ReservedBefore/SegmentAllowance/ReservedThrough在提交门原子接纳，未用额度不返还；活动同段幂等、不延长期限，准备/活动并发拒绝，完成段号和终态防复活。每段保留Q+全部A+M/一个完成槽；借用者结清才更新guard并交付完成。Result.RoRecovery单独携带累计发送证据，原Result.Delivery保持本段语义；历史可能写出不得因本段零字节失败而清除T。

十三组测试（含十一组TCP/TLS）和全量纯Go/race/vet/build/生成检查通过。第51批combined INITIAL测试标签遗漏已修正，并增加实际CC类型断言；历史快照不改写。九定义/输出、929wire/17XML和generator13/ops8不变。下一批ResponseWait/Tr；错误答案后/跨OCS/重启/RecoverySpace及目标交接仍未完。24项SPEC继续Partial/Planned，release_status=not_ready。

### 6.53 第五十三批实施记录（2026-11-09）

完成[Ro ResponseWait与同段晚答](../spec/history/batch-53.json)，前批冻结于[history/batch-52](../spec/history/batch-52.json)。RoRecoveryPlan增加可选ResponseWait，排队即计Tr，段内首发/备用同值且不超过原上限；业务Tx/Receive超时不延长或替代Tr。整体deadline优先，单次/最后Tr给独立错误与停止原因。

健康连接保持状态和旧pending/A；writer仍借用时只报告到期，结清才开放显式后继；旧答可撤下尚未开始的后继，已开始的writer及借用A的reader均结清才完成一次。无备用/反压不刷新计时或增加attempt，当前Tr之后的真正关闭/协议失败仍生效。复用单Node timer和整段Q/逐attempt A/M，无持久化或自动业务重试。

十二组TCP/TLS测试覆盖四Profile、配置/计时/幂等、排队未写出、晚答、读写屏障、预算和关闭；全量纯Go/race/vet/build/生成及证据审计见progress。九定义/输出、generator13/ops8、十二Context及929wire/17XML不变。下一批具名错误答案后的有限本地续段，再接跨OCS及重启导入/RecoverySpace；完整RO-06/目标兼容与发布仍未完。24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.54 第五十四批实施记录（2026-11-09）

完成[具名错误答案后的本地有限续段](../spec/history/batch-54.json)，前批冻结于[history/batch-53](../spec/history/batch-53.json)。默认所有答案终态；冻结策略名及Plan显式模式只授予有限基础错误集合，SMSFW额外集合受应用给定的Tx截止限制。经关联/来源校验的获选答案结束当前段，NextTrigger由本地guard决定；调用方不能伪造结果、混淆NoAnswer或回退累计额度。

错误续段T=0而历史CumulativeDelivery不清除；最后有效错误之后的可能写出才决定后续无答T。原帧身份/正文保持、新HBH、旧段晚答隔离及收发结清屏障复用；关闭期间不授予错误续段资格。根包不新增应用类型依赖、业务计时器、持久化或自动重试。

四Profile补正常CCA4002接收专用、不完整语义报告，标准Ro/combined slice.5、smsfwro/legacy420 slice.6及四输出指纹更新，其他五定义/输出、generator13/ops8、十二Context及929wire/17XML不变。十组真实TCP/TLS测试含288种模式/结果组合及期限/额度/关闭验收，全量检查见progress。下一批授权跨OCS，再接重启导入/RecoverySpace；旧独立证据不冒充新增恢复认证。24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.55 第五十五批实施记录（2026-11-09）

完成[有限授权跨OCS及同DRA恢复](../spec/history/batch-55.json)，前批冻结于[history/batch-54](../spec/history/batch-54.json)。Node恢复目标表限制TargetID/Peer/目的/来源及部署引用；记录v2冻结权限，v1拒绝自动升级。计划Alternates固定有序不同目标；缺共同去重域/保存期、TERM具名会话及本请求证据时拒绝，声明不等于现场认证。

每目标帧/快照在预留Q后构造，只改Destination；同DRA连接不同目标可分配新HBH，旧pending/A和来源仍独立，旧答案可赢一次完成。离开目标不回访，最大16目标/总64路径；新段仍受原总额度及年龄约束。RequestInfo/AttemptResult报告TargetID；普通Route不获得迁移权限，标准CC-Session-Failover协商/SCAP1仍未开放。

十组测试含九组TCP/TLS，四Profile/15用途新旧答案竞争共60条主流程及权限/预算/收发退休/关闭/最长目的边界。全量验证见progress；九定义/输出、generator13/ops8、十二Context及929wire/17XML不变。下一批原记录导入/RecoverySpace，应用保存/单发送者/历史进度与实际OCS证据继续归应用/部署。24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.56 第五十六批实施记录（2026-11-09）

完成[原记录导入/RecoverySpaces](../spec/history/batch-56.json)，前批冻结于[history/batch-55](../spec/history/batch-55.json)。AllowRoRecoveryImport默认关闭，Plan.Import携带应用隔离/进度/时钟声明及有限Checkpoint；仍复核原Origin/Bundle/冻结目标、原期限及剩余额度。未知崩溃保守扣全段并设T=1；合格错误历史保留T=0谱系。声明不是存储、单发送者或历史真实性证明。

原E2E位于普通应用半区时检查本轮已用区间或自定义active/quarantine；域外只接纳与普通应用/控制全域分离的最多16个RecoverySpaces。导入不调用发号源，复用有界class guard和Q/A；失败接纳回滚，活动幂等/终态防复活和准备/收发关闭屏障通过验收。记录保持v2，计划摘要v5。

十一组测试包含四Profile/15用途×TCP/TLS、新Boot/原正文、真实错误结果跨目标、未知发送、恢复域所有权/配额及关闭回滚。全量检查见progress；九定义/输出、generator13/ops8、十二Context及929wire/17XML不变。下一批限RO-06应用交接示例/验收，标准会话协商、SCAP1恢复、SCTP及现场/发布验收仍未完；24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.57 第五十七批实施记录（2026-11-09）

完成[RO-06应用交接示例与验收](../spec/history/batch-57.json)，前批冻结于[history/batch-56](../spec/history/batch-56.json)。三个可执行Example、七组真实TCP/TLS测试涵盖确切意图确认、提交未知不发送、未知崩溃保守扣整段、真实2001/4012终态禁止重启恢复、Tx后收答/Cancel、DDFH Presence/配置及原业务截止不重置。

应用策略全部为测试示例，未新增生产API/依赖、存储provider或重放worker。JSON/摘要回执不构成持久化/认证保证；库和应用退出条件见[交接清单](ro-recovery-handoff.md)。本地交接示例收尾，实际应用提交确认、隔离及OCS证据仍需部署验收，不能据此加入库内数据库职责。

全量检查见progress；九定义/输出、generator13/ops8、十二Context及929wire/17XML不变。下一批转SCTP上游U1核验/U3生命周期；标准会话协商、SCAP1恢复、私有/现场及发布验收仍独立追踪。24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.58 第五十八批实施记录（2026-11-09）

完成[上游SCTP U1核验/U3等待层基础](../spec/history/batch-58.json)，前批冻结于[history/batch-57](../spec/history/batch-57.json)。上游基线62ad388已与origin HEAD核对；支持Linux平台的标准deadline、排队/当前操作唤醒、内部ctx/固定S、hook退休及EINTR受控边界已实施，原非支持平台和旧socket timeout兼容。未修改上游独立v1设计任务，未发布依赖。

九组新增测试、真实SCTP全量257条测试/子测试零skip、race及CGO=0 vet/build通过；111项上游生产依赖纯Go且无外部模块；arm64/Windows仅交叉编译。精确补丁和源码/命令证据见batch58；本库生产源码、定义、生成器与929wire/17XML未变化，沿用第57批证据。

U3还需公开Context解析/拨号交接及后续消息入口接线，U2/U4/U5和本库适配器仍未完成。下一批仅补解析/拨号总预算与交接取消竞态，不以等待层通过声明SCTP生产可用。24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.59 第五十九批实施记录（2026-11-09）

完成[上游SCTP Context解析与拨号](../spec/history/batch-59.json)，前批冻结于[history/batch-58](../spec/history/batch-58.json)。Go resolver保留旧地址语法并复制结果，DialContext复制输入/config、在fd前拒绝Control，使用同一ctx及原poller完成总预算和最终交接检查。真实setup/connect errno保留；失败释放资源，交接成功后ctx取消不影响连接。旧入口、第58批等待层和上游v1提案均未改。

九组新增测试及外部包示例、上游完整纯Go/race/vet/build和交叉编译记录见本批证据；受控DNS/局部交接检查不冒充外部网络故障演练。增量补丁按58→59顺序应用，未发布版本；本库源码/定义/929wire/17XML不变，沿用原证据。

U3仍需在U2消息操作中完成集成与边界验证；下一批消息模式和ReadEvent/WriteMsgContext。U4/U5、本库SCTP适配/Peer预算及实际部署仍未完成，24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.60 第六十批实施记录（2026-11-09）

完成[上游SCTP显式消息链路](../spec/history/batch-60.json)，前批冻结于[history/batch-59](../spec/history/batch-59.json)。Messages在创建fd前复制，ReadEvent在同一个读gate内peek/分类/consume，WriteMsgContext显式TTL0且不重发短写尾部；新旧模式互斥。必需选项在建连/监听和交接点核验，GetMessageOptions返回实际内核读值。通知超限不消费；损坏已消费边界锁存CloseRequired。配置新增字段的非具名literal迁移影响已记录。

十组新增、六个子测试和一个外部包示例使全量达到284条、零skip；完整race/CGO0vet/build及交叉编译见证据。故障注入不等于全部通知真实触发，IORequirements逻辑scratch不等于连接总内存。本库代码/依赖/929wire/17XML未变。58→59→60补丁尚未发布；独立上游v1提案不由本批实施。

下一批U4五类typed通知和实际大通知/失败/地址/partial事件证据及选项拒绝检查，后续U5发布/平台和本库适配器/Peer/完整预算。U2最小链路完成不表示联合出口或SCTP生产接线完成。24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.61 第六十一批实施记录（2026-11-09）

完成[上游SCTP五类通知解析与真实事件](../spec/history/batch-61.json)，前批冻结于[history/batch-60](../spec/history/batch-60.json)。ParseNotification检查完整长度、固定布局和地址family，固定字段owned、Tail显式借用，未知type/state/flags原值保留；SNDRCV共用解码且PPID只转换一次。已消费通知解析失败由适配器关闭，不追加Peer/重放/持久化职责。

五类通知在当前Linux/amd64原生触发，补齐长SEND_FAILED（超限不消费后重读）、地址和partial-abort证据。选项真实读回不一致和20次accept拒绝的fd清理/监听器继续服务已验。十组新增/十子测试/六fuzz种子及一个Example，全量312条零skip；全量race、20轮定向race、10秒327046次fuzz及CGO0构建检查通过。完整证据及编译-only平台限制见批次记录。

本批关闭当前主机U4最小链路及对应U2原生通知缺口。下一步限定U5完整连接/解析结果预算、同候选发布与native平台出口，之后本库适配器/Peer接线。58→59→60→61补丁未发布，不新增依赖；本库929wire/17XML及生产代码不变。24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.62 第六十二批实施记录（2026-11-09）

完成[U5本地资源、候选安装及门禁整合](../spec/history/batch-62.json)，前批冻结于[history/batch-61](../spec/history/batch-61.json)。审计每连接单reader/单writer/无上游排队的源值和对象槽数，建议8192上游逻辑固定额度；该值不等于RSS或最终adapter预算，变量缓冲、建立地址副本和Go/内核额外开销单列。新增资源漂移、ABI和100轮取消后owner/gate退休检查。

同候选确定性打包、独立module cache且无replace的真实安装收发已验；发布脚本接入CGO0和新通知fuzz，架构/consumer脚本引用新API。具体执行证据与430文件清单见批次记录，离线版本标签不是上游发布。远端矩阵、正式commit/版本及本库adapter/Peer/实际预算未完成，整个U5保持未关闭。生产Go/定义/929wire/17XML不变，24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.63 第六十三批实施记录（2026-11-09）

完成[上游精确候选PR/矩阵/模块引用](../spec/history/batch-63.json)，前批冻结于[history/batch-62](../spec/history/batch-62.json)。PR #30 head1649569及CI临时merge的tree与第62批候选一致，430文件真实下载核对；伪版本v0.7.5-0.20261109044437-1649569dd9b8可由Go解析，独立无replace消费者真实收发通过。

远端14项原生/所有权/QEMU任务全部成功，原生强制namespace与完整脚本、各架构新增ABI/parser执行证据归档。可安装commit及同候选矩阵条件已满足，下一批可固定该候选进入SCTP适配器开发。PR仍为draft、未合并/未正式release，不等于本库发布或现场互通通过。本库生产代码、依赖、929wire/17XML不变；24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.64 第六十四批实施记录（2026-11-09）

完成[独立SCTP适配/预算/Peer接线](../spec/history/batch-64.json)。固定上游1649569伪版本，核心无第三方依赖；通知同步续调覆盖握手/body/discard，固定事实与终止进原状态门。默认adapter16432逻辑字节，Node另计每候选2048事件额度；无整帧副本、无持久化。补齐有序CEA/DWR及响应方固定时限屏障，合法握手后消息后才转无序业务。

真实双Node Base/S6c、取消/晚答/关闭和资源退场通过；核心CGO0/race各2603项，adapter35项及20轮race700项通过。模块无replace安装验证使用本地核心生产源码候选，不宣称核心版本已公开；SCTP托管Dialer/Listener、具名多流、正式发布和DRA/OCS验收仍待完成。24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.65 第六十五批实施记录（2026-11-09）

完成[有限SCTP Dialer/Maintain与Listener/Serve](../spec/history/batch-65.json)。同候选/同核心版本；SetupConfig在复制和解析前限制显式地址N、zone字节Z、端点文本D，默认listener57728、每物理候选61872逻辑字节，两者并存，控制/事件/业务额度另计。拨号解析/连接/CER共享绝对期限；listener用上游poller的20ms等待片段检查取消，不增后台accept worker；已接纳socket清除继承timeout。Serve取消不关闭已交接Peer；Close与Establish竞争，Wait在回调真正返回后才释放槽。

核心CGO0回归2603项；独立模块全量CGO0及race各55项，新接纳/关闭/选举范围20轮race380项，零测试skip。真实重连经过REOPEN三次DWA，旧请求一次WrittenNoAnswer、晚答被拒、没有自动重放，新显式请求正常。Listener/候选一字节不足在调用前拒绝，双向拨号选举及DPR/DPA关闭通过。核心114份生产文件与第64批包完全一致，不因adapter变化重新发布同版本不同内容。正式依赖发布、OrderedMultiStream/PPID0具名兼容与实际DRA/OCS仍未关闭；24项SPEC保持Partial/Planned，release_status=not_ready。

### 6.65.1 后续上游职责收敛（2026-11-09，仅设计）

按维护者要求，下一步调整为[skill215/sctp通用能力下沉](sctp-upstream-improvement-plan.md)，由另一session实施。A–D补原生listener取消/deadline、有界建立、socket公共能力和机器可查资源账目；E补单消息连续性/流边界证明。第65批代码及测试有效，上游新增工作均Planned，未更新依赖或运行新能力测试，不记为第66批实施完成。现有24个SPEC、Go1.26/纯Go/无持久化边界不变；A–D新候选通过后本库迁移删除底层重复逻辑，再验证原Base/S6c与生命周期。OrderedMultiStream/PPID0的Diameter策略保留在本库，不以全SCTP重写为前置。

### 6.66 第六十六批实施记录（2026-11-09）

完成[上游A–E核验、固定候选及薄适配迁移](../spec/history/batch-66.json)。70源哈希、上游日志及14项远端CI一致，adapter消费真实bb739b612088伪版本；移除20ms Accept轮询、socket syscall、重复N/Z/D算法、硬编码上游额和跨片段元数据比较。listener/候选/既有socket分别52620/99676/17756默认amd64逻辑字节，由上游资源查询加本库值计算。

合法RESTART/partial-abort与错误/CloseRequired并存时保留typed事实，writer先见restart不误判本地协议失败；Diameter长度/EOR、PPID、期限/通知限额、Peer/预算仍在本库。核心2603 CGO0、adapter71 CGO0/71race及70项20轮race1400通过，零测试skip；核心114生产文件不变，完整核心race沿用第64批。本批闭合先前A–E交接及消费迁移，不代表具名多流/PPID0、正式模块发布或现场互通已完成。24项SPEC仍Partial/Planned，release_status=not_ready。

### 6.67 第六十七批实施记录（2026-11-09）

完成[正式SCTP v0.8.0接入与回归](../spec/history/batch-67.json)，精确commit bce1b3cf3ea32a32a86262c7f655d565bb40055c、真实下载sum与70源文件身份均已核验；A–E接口/薄适配和默认资源账目保持，无Go代码变化。核心CGO0/race各2603、adapter各71、有限生命周期20轮620项通过；独立正式包consumer、vet/build、依赖验证及交叉build通过，零测试skip。上游已正式发布，不再将上游merge/tag列为未完成；核心仍为离线batch64候选，核心/adapter公开发布、具名多流/PPID0及现场DRA/OCS验收仍单列。24个SPEC仍Partial/Planned，release_status=not_ready。

### 6.68 第六十八批实施记录（2026-11-09）

完成[DRA Relay协商和逐请求终端绑定](../spec/history/batch-68.json)：远端顶层Relay广告只展开本地显式具体应用，SCAP1特殊广告不放开；PeerConfig.RequestTargets最多64条owned授权，RequestOptions.TargetID按请求选择，RouteConfig.TargetID固定跨attempt目标及来源。新目标表计入Node固定预算，接纳失败/错源答案/取消晚答/重传经TCP/TLS及真实SCTP验收。核心变化后adapter升级离线batch68候选，SCTP仍v0.8.0；不新增发现/存储/业务重试。计数及工具链见批次证据，其他全局缺口、v1及现场验收仍未完成。

### 6.69 第六十九批实施记录（2026-11-10）

独立`app/ro/abort` Profile 实现 App4/274 ASR/ASA，可与标准Ro或三种旧CCR绑定之一组合；不修改这些CCR的字段或兼容版本。ASR要求Session/Origin/目的Host+Realm/AuthApplicationId4，支持User-Name/Origin-State/Proxy/Route-Record；ASA正常结果仅2001/5002/5012，E沿独立基础语法。关联检查Header/Session/Proxy/来源，不要求CCA字段，不回显请求Origin-State。应用明确选择结果，`NoContext`只表示所选协议无需额外facts，不能证明业务已经终止。Redirect、实验结果及已知未选计费字段不在本切片内。

复用现有Inbound、按字节预算、Canonical及有限重复缓存，多Arrival各自补包络；相同ASR重放已提交结果，内容冲突E/5012。响应准备失败/调用取消可重试；提交后主Arrival写失败不能重新作业务决定，可由其他Arrival取得原答案。ASR与同Session在途TERM独立完成；应用仍负责32.299 §6.5.5适用TERM、用量、序号及跨命令协调，库不自动创建第二份TERM或跨重启去重。ASR不授予TransportReplay。

定义/正反样本/真实TCP/TLS测试及独立go-diameter解释见[batch69](../spec/history/batch-69.json)，机器验收见[证据](../spec/evidence/batch-69-ro-abort.json)。核心离线候选更新为batch69，独立adapter在固定SCTPv0.8.0上回归；尚非公开发行或目标OCS认证。RO-05库机制有界完成，其他Ro/P4/发布出口不自动关闭，24项SPEC仍partial/planned。

### 6.70 第七十批实施记录（2026-11-10）

新增`app/ro/dcca`及`spec/dcca.json`，提供RFC4006单服务EVENT（4/0）余额查询、直接扣费及通用退款。每个Bundle选择一种App4/272绑定；该绑定与独立ASR/ASA组合，不能与其他272 Profile同时编译。显式非3GPP Service-Context、顶层Service-Identifier、1–2个MSISDN/IMSI Subscription、可选413标准金额或417事件单位；金额保持int64 digits/Optional int32 exponent/uint32 Currency，未选MSCC、2022/1292及3GPP服务组拒绝。

余额0/1保留，成功缺422、未知值保留owned答案并给出不完整条件报告；发送/Inbound.Respond拒绝未完成条件，不提交响应权。2001+NO_CREDIT仍为查询成功。直接扣费2001要求GSU，4011独立保留；退款不要求GSU。原动作不符的余额字段、错误CC编号/类型拒绝关联；失败、DDFH和Cost保持独立，库不推导扣费/退款或发起后续操作。纯`CheckRefundRelation`检查新Session及原服务/订阅/端点关系，不要求已知原成功答案、不判断金额或资格、不承诺跨进程唯一性。

本批是通用DCCA子集，不宣称32.274失败SMS退款或旧六模板替换完成。现有十份定义、Ro恢复授权和运行时生产代码不改；新绑定SingleAttempt，无新增facts种类/worker/持久化。71新独立帧和33份go-diameter离线解释、真实TCP/TLS收发/报告/取消关闭与预算，以及独立SCTP模块在新核心batch70上的回归见[记录](../spec/history/batch-70.json)和[证据](../spec/evidence/batch-70-dcca-event.json)。剩余标准SMS结果、2022原答案/服务父组回送检查、应用/目标OCS及发布门仍保留。

### 6.71 第七十一批实施记录（2026-11-10）

标准SMS Ro选定IEC/INITIAL/TERMINATION的顶层及单MSCC结果4010/4011/5030，分别保留原码、层级和用量；失败不强制GSU、不变成零扣费或退款授权。2002、子MSCC5031及多服务仍未开放；原顶层5031诊断和5011安全接收保持。

`ro.RefundInformationForService(original, answer, serviceID, maxBytes)`和`ro.CheckRefundInformationEcho(...)`是无状态的2022叶级检查器。输入必须先经codec.ValidateAnswer完成Header/授权来源关联，或使用runtime已关联答案；离线codec没有连接状态，原请求与实际连接的选择归调用方。helper另验Session/Proxy、原IEC/单服务键和显式原DestinationHost/Realm与答案Origin。不能以裸body相同代替原交换证据。范围限定原标准SMS IEC DIRECT_DEBITING；缺原请求/答案/答案服务父组返回ErrContextRequired，多父组/非IEC/未完成语义拒绝，错误不返回token。

原服务父组已知无2022返回Absent；已出现空值仍Present；二进制按完整长度和字节比较。maxBytes为0..MaxWireSize，超限复制前返回ErrLimit；返回owned副本，不保留输入。诊断中的2022不计入所选父组，不跨服务/OCS借用，不把私有1292转换为2022。Check只比较拟回送叶的presence和bytes；不构造退款CCR、不验证新退款身份/金额/资格或历史真实性。独立helper的输出由调用方管理，maxBytes限制单次复制；Node原收答预算仍由现有资源账目承担。

46新独立wire（36结果组合、token状态及5反例），41份go-diameter字典/原字节重编码，真实TCP/TLS和全量检查见[batch71](../spec/history/batch-71.json)。Ro slice.6指纹更新，旧恢复记录必须继续匹配精确Bundle；没有记录迁移旁路。其他十定义、generator13/ops8和所有恢复许可不变。原SMS失败3409/在线序号3408限制、完整SMS退款发送、私有/目标OCS证据及其他发布出口保持；不以本批leaf检查宣称RO-03整条流程完成。

### 6.72 第七十二批实施记录（2026-11-10）

独立SCTP适配器完成显式OrderedMultiStream和PPID0Compatibility。默认无序stream0/PPID46不变；PPID0模式发送0、接收0/46，47及其他值拒绝，配置不由收包自动推导。Config.SocketConfig供预建socket，托管SetupConfig传播相同配置；流请求值不作为实际I/O上限，GetStatus冻结实际双向流数。

ordered=true或Base控制固定有序stream0；屏障完成后的有序多流按精确顶层Session-Id的固定FNV-1a64选择1..Ostreams-1，无Session用AppID/CommandCode/E2E，排除HBH/R/T。N=1回退有序0并通过ModeInfo.SingleStreamFallback公开。响应只按本地出流，接收接受任意协商入流。无会话表、无新增流队列、无业务排序/去重承诺。

复用公共wire顶层边界遍历和上游消息I/O，新增固定值在既有逻辑预算内；不重复通用SCTP解析/取消/建立能力。完整帧只send一次，短写不重发尾部，CEA/DWR和关闭时限仍由原Peer门控制。原level0 HOL、DPR/DPA非跨流交付确认、RESTART关闭旧代及关闭不自动重放的边界保持。实际Base/S6c、非对称流协商、预算、fuzz/race和关闭验收见[batch72](../spec/history/batch-72.json)。

本批只改adapter生产代码及核心架构依赖测试；核心生产121文件、11定义、1099wire/17XML保持，因此仍固定离线核心batch71和正式SCTPv0.8.0。具名模式的库机制已完成，目标部署接受、公开发行及P4/Ro/SCAP1余量未由此关闭；保持纯Go无持久化。

### 6.73 第七十三批实施记录（2026-11-10）

`compat/sgdalert.Profile()`提供既定`legacy-sgd-alert-v1`固定绑定，仅新增App16777313/8388648 ALR/ALA。两个独立Go消息类型复用已实施S6c的字段、深层校验、角色/IMS/T4 Context、正常/E语法和owned操作；不复制第二套生成定义，不按收到的Header改写应用身份。原标准S6c与SGd OFR/TFR可同时编入Bundle，未显式选择该Profile时仍拒绝旧绑定。

答案必须保持原请求的Application-Id/标识/Session/Proxy并使用明确的本地Origin；不延续旧handler将Destination直接当作Origin的行为。有限来源/目的绑定、一次完成、取消晚答、预算、重复缓存和关闭全部走已有公共运行时。相同Origin/E2E但不同App的请求是内容冲突，不能重放另一绑定的缓存答案。ALR仍不授予TransportReplay。

[TS29.338 V18.2.0 §6.3.2.2](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf)明确ALR/ALA即使经SGd/Gdd也使用S6c App16777312；本绑定因此保留具名兼容身份，来源为sms-fw已注册的旧SGd handler。原go-diameter默认SGd字典拒绝该命令，本批独立解释只在进程内显式加入命令声明，不能据此声称其默认支持或实际部署认证。

26份独立Header投影报文、正反规则继承、真实TCP/TLS及回归见[batch73](../spec/history/batch-73.json)。现11份生成定义、额外1个固定委托Profile、20组请求/答案绑定（含重叠CCR）、1125wire/17XML。新的122文件离线核心batch73供adapter回归，上游仍为正式SCTPv0.8.0。Router opaque token/MCCMNC、模拟器未证行为、Ro/SCAP1及公开发布出口仍单列；不扩大本批结束条件。

### 6.74 第七十四批实施记录（2026-11-10）

`s6c.RouterAlert(RouterAlertEvidence)`明确绑定Router/GMSC两端、原TFA的User-Identifier/User-Name及原TFR的SMS-GMSC-Address，支持标准S6c及显式旧SGd Alert委托。token选定1–256字节合法UTF8/无NUL子集，精确比较，不做IMSI推导、大小写折叠或Unicode正规化；号码按现有数字TBCD入口。应用负责历史真实性和token映射，库只冻结/校验给定证据。

ALR共享User-Name的线路语法相应放宽，未给用途仍NeedsApplication；MME/SGSN必须数字IMSI、T4/IMS原条件不变，不能把静态codec成功当作已验证Router请求。Event/新节点、正常/E答案及源授权、Canonical和有限缓存、取消关闭继续沿用既有公共契约。Context复用固定存储，最大证据仍在2KiB预留内，无新状态机/预算种类或自动SRR/SMS。

S6c slice.21与旧委托Bundle指纹更新，其他十定义不变；23新独立帧、原字典/显式旧App overlay及真实TCP/TLS验收见[batch74](../spec/history/batch-74.json)。现1148wire/17XML、11生成定义加1委托Profile、20对绑定。核心离线候选batch74，上游仍正式SCTPv0.8.0。仅完成Router提醒分支；SGd TFA token输出、SRA Router/MCCMNC用途、模拟器边界及公开发布/部署验收不由本批关闭。

### 6.75 第七十五批实施记录（2026-11-10）

`sgd.RetransmissionToRouter`复用原Retransmission Context，带Maximum-Retransmission-Time的请求MUST提供同PLMN及RouterAlertSupported已知true断言；缺失事实MUST报告NeedsApplication，显式false或错误端点MUST拒绝。直达MME/SGSN仍要求相应接入号码，并拒绝Router专属能力字段。无接入号码的TFR必须使用明确Router用途，不能从DRA/CER猜测角色。

Router正常5550+Requested-Retransmission-Time答案MUST含专用User-Identifier/User-Name token；选定1–256字节UTF8/无NUL原字节子集，与S6c RouterAlert共用规则。token不能替代原TFR IMSI，不能自动生成映射或下一条ALR。无3331的token当前ErrUnsupported。原请求最大时间、节能/era断言、正常/E关联、相等值接收NeedsApplication及完整发送拒绝均沿用既定契约。数字同形token不赋予角色证据。

本批SGd slice.15，其他十定义及S6c/旧Alert委托摘要不变，Context仍在固定2KiB预算内；30新独立帧，1178wire/17XML。真实TCP/TLS、独立go-diameter字典解释、完整核心/adapter回归及候选batch75证据见[batch75](../spec/history/batch-75.json)。P4仅余SRA Router/MCCMNC用途及模拟器有界输入；公开发布、Ro/SCAP1证据门仍单列，未宣称现场或全协议认证。

### 6.76 第七十六批实施记录（2026-11-10）

`RouterRouting`MUST显式断言GMSC/Router及HPLMN强制路由，`MCCMNCRouting`MUST匹配原MSISDN SRR的SM-Delivery-Not-Intended=1和固定端点。两种用途MUST NOT借用邻接DRA、T4/IMS身份或相同数字字节推导角色。MCCMNC前缀事实缺失时，成功答案MUST保持NeedsApplication；不得自动查归属、补dummy或改活动上下文。3311=0及其余值当前明确Unsupported。

成功SRA的User-Name采用有界UTF8共享语法，普通HSS/T4/IMS完整校验仍限制IMSI；Router用途允许1–256字节关联标识。原请求没有3311时成功Serving-Node仍MUST存在，MCC/MNC查询允许省略。后者身份MUST匹配显式5/6位MCCMNC，最多15位数字且含MSIN；不认证IMSI分配或dummy真实性。正常失败/E及Additional/GPRS等既定关系不变。

kind14/15复用固定存储/2KiB，S6c slice.22和旧Alert委托指纹更新，另外十定义不变。27新独立帧、1205wire/17XML、真实TCP/TLS和完整回归见[batch76](../spec/history/batch-76.json)。原P4余模拟器输入接受/拒绝，其他Ro/SCAP1及发布验证门单列；不重开逐叶前置分析。

### 6.77 第七十七批实施记录（2026-11-10）

SIM-01既有模拟器输入接受/拒绝范围完成，见[迁移矩阵](diameterkit-simulator-migration.md)。三个源码路径的坏P标志、Origin/IMSI/ALR身份、OFA状态、UDA数据及失败SRA附加字段均有拒绝/修正证据；旧手工Proxy/Origin包络由既有CompleteAnswer/Inbound.Respond承接。SMS VSAI只接收忽略、S6a后置、业务号码/角色/TPDU生成边界保持。

28份独立go-diameter编码投影及真实TCP/TLS验收见[batch77](../spec/history/batch-77.json)。现1233wire/17XML；生产代码、11定义及1委托Profile不变，125文件核心继续使用离线batch76，SCTPv0.8.0不变。纯Go/无持久化要求保持；本批不创建新版候选或新需求ID。输入决策结束，但SIM-01旧进程迁移/对照、目标验收与Ro/SCAP1及P7门仍分别未完，不再以本批为名启动全字段调查。

### 6.78 第七十八批实施记录（2026-11-10）

SCAP1旧广告模式由PeerConfig.SCAP1Advertisement显式选择，默认标准行为不变。具名LegacyDual模式只在正常CER/CEA承接源码固定相邻双VSAI，接收原地移除冗余组再走严格Base及原请求/来源关联；发送按同布局追加。必须已有SCAP1Accounting策略及严格本地能力，不从Auth19302派生会计授权，不扩大App0分派或恢复权限。

原始帧/AVP/重复/深度预算在移除前计费；发送44字节有界预留，CER/CEA预检。现有ACR/ACA缓存完成多Arrival独立Proxy、内容冲突、取消/关闭、C耗尽关闭与资源归零验收，见[batch78](../spec/history/batch-78.json)。17新帧、1250wire/17XML，11定义及1委托Profile不变；126文件离线核心batch78，SCTPv0.8.0不变。旧广告及库缓存机制结束；私有ACA、SCAP1恢复、六模板目标与公开发布仍有原证据门。

### 6.79 第七十九批实施记录（2026-11-10）

SCAP1双向接入示例及所用公开API核对完成，见[batch79](../spec/history/batch-79.json)和[运行说明](../examples/scap1/README.md)。两个独立Node用真实TCP/Serve/Dial，验证标准/旧广告、ACA2001/5012及E3004、槽/字节预算、取消晚答和端口/Node关闭。独立无replace消费module编译运行并核对安装版公开文档；Submit ctx与Receive ctx、完成预算生命周期明确，不依赖根包私有helper。

修正公开注释的历史措辞，未改变库声明、运行逻辑、定义或语料；1250wire/17XML、11生成定义+1委托保持。128文件新离线核心batch79含示例，SCTPv0.8.0不变。此项不关闭私有ACA、SCAP1恢复、目标OCS或整个P7发布门。

### 6.80 第八十批实施记录（2026-11-10）

[batch80](../spec/history/batch-80.json)完成Sh Ref24 UDR/UDA双向真实TCP公开例、显式Sh-Data构造/解析及有限资源/取消/清理验收；独立消费module验证25个所用公开API、CGO0/race各16项。示例20轮race320项，核心各3338、SCTP适配各113；既有定义/协议语料/运行时行为不变。新离线核心batch80共130文件，SCTPv0.8.0保持。

[应用定义指南](diameterkit-application-definitions.md)明确各应用规范基线、头ID、切片版本及JSON→生成Go→手写规则→Profile/Bundle入口。版本和操作表不得被当作完整Release符合性；Profile加载不得推导自动广告/授权，报文Validation.Complete不得代替Sh XML/业务校验。来源清单为派生证据，不增另一套规范或需求ID。其余应用例、完整API/发行和真实目标门未完成，24项状态仍Partial，release_status仍not_ready。

### 6.81 第八十一批实施记录（2026-11-10）

[第81批证据](../spec/evidence/batch-81-s6c-example.json)完成S6c SRR/SRA真实TCP公开例：GMSC发送、HSS应答，普通MSISDN/SC与真实IMSI/MME成功形状、基础失败、10415实验失败及E错误分别处理；身份/目的/原SRR Additional许可、资源/取消/清理均经公开API验证。独立消费module核对24个所用公开入口及CGO0/race各25项；示例20轮race500、核心各3363和SCTP适配各113通过。

既有库源码、协议定义和1250wire/17XML未变，核心离线候选batch81共132文件，SCTPv0.8.0保持。实验结果不得丢失Vendor命名空间，应用提供的IMSI和路由节点不得被解释成库查询或真实投递证据。此例完成不代表其他S6c流程、SGd/Gdd/Ro公开例、完整P7或真实目标验收结束；24项仍Partial，release_status=not_ready。

### 6.82 第八十二批实施记录（2026-11-10）

[第82批证据](../spec/evidence/batch-82-sgd-example.json)完成SGd/Gdd真实TCP公开OFR/OFA、TFR/TFA例。明确接入端MO发送/MT接收与短信端反向策略、固定终端及请求/答案来源、MME/SGSN目标、基础结果/实验5555方向原因/E包络和透明字节。反向命令、错误目的/来源、无效应答及资源/取消/晚答/清理经公开API验证。独立消费module八种CLI、25个安装版公开入口及CGO0/race各71项通过；示例20轮race1420、核心各3434和SCTP适配各113通过。

既有库Go源码/协议定义/1250wire/17XML不变，离线核心batch82共134文件，SCTPv0.8.0保持。载荷保留不得解释成SMS投递，合设IWMSC/GMSC的示例配置不得推导实际DRA或终端角色发现。batch Markdown仅本地保存；公开进度和验收证据继续维护。下一步为标准Ro SMS IEC公开例，完整P7/真实目标门未闭合，24项仍Partial，release_status=not_ready。

### 6.83 第八十三批实施记录（2026-11-10）

[第83批证据](../spec/evidence/batch-83-ro-example.json)完成标准SMS IEC DIRECT_DEBITING真实TCP公开CCR/CCA例：CTF只发请求、OCS只收请求并应答；显式Auth4/固定终端、五种命令/服务结果、透明token及精确单位，原CC类型/序号/服务关联和预算/取消/关闭已验。独立消费module五种CLI、25个公开入口、CGO0/race各33项；例20轮race660、核心各3467和SCTP适配各113通过。

协议AnswerReceived/Validation.Complete不得被当作扣费成功，顶层2001不得覆盖服务级4012，失败/取消不得推导未扣费或自动退款。当前例没有持久化、业务重试或RO-06授权。既有Go源码、定义与1250wire/17XML不变，离线核心batch83共136文件，SCTPv0.8.0保持。batch Markdown只保留本地；下一步为无UPDATE ECUR公开流程，其余P7及真实目标门仍未完成，24项Partial、release_status=not_ready。

### 6.84 第八十四批实施记录（2026-11-10）

[第84批证据](../spec/evidence/batch-84-ro-ecur-example.json)完成标准SMS无UPDATE ECUR真实TCP公开INITIAL/TERMINATION例。应用显式保留Session、CTF/OCS、服务及编号关联，两个exchange标识独立；先消费有效INITIAL答案并检查显式服务grant，才由示例应用发TERM。`CheckECURContinuation`是纯请求对检查，不表示已获额度或完成投递，也不启动库内计费会话。

INITIAL成功/服务拒绝/顶层失败/E与TERM2001/5002/5012/E分支、旧答案/错服务/不允许的TERM grant/token、独立预算、取消/关闭/晚答通过。示例68项、20轮race1360，核心CGO0/race各3502、适配各113；独立消费各68项和26项公开API查询。既有库Go源码、11定义及1250wire/17XML不变；扩展示例CLI、增加ECUR文件，离线核心batch84共137文件。无持久化、自动金融重试或RO-06授权；batch Markdown保持本地忽略。下一有界出口为既有独立ASR/ASA公开流程，P7其余项和真实目标仍待，24项Partial、release_status=not_ready。

### 6.85 第八十五批实施记录（2026-11-10）

[第85批验收](../spec/evidence/batch-85-ro-abort-example.json)完成独立Ro ASR/ASA公开真实TCP例。Base+单CCR+独立abort Profile组合、显式Auth4/OCS发送/CTF接收、应用原Session引用和正常2001/5002/5012或E3004选择均通过。ASA不附CCR类型/序号/授予单位；连续性/业务状态由应用负责，不以协议报告代替中止动作。未授权反向ASR、未选CCR及ASR显式重传入口被拒绝。

原Session/Proxy应答关联、远端Origin-State不回填、结果/字段拒绝、槽/字节预算、取消/晚答/Respond关闭竞态已验；同Session挂起CCR在ASA后仍等待，不自动TERM或Cancel。Ro例累计100项（本批32项）、20轮race2000、核心CGO0/race各3534、适配各113，独立消费各100项及33API查询通过。库Go源码、11定义/1250wire/17XML不变；只扩展示例和CLI，离线核心batch85共138文件，formal SCTPv0.8.0保持。

无持久化、业务会话管理或新增恢复许可；batch记录保持本地忽略。下一有界出口为P7公开支持矩阵/API覆盖收尾：按既有范围标明Ro余额/退款/私有绑定/恢复还缺哪些消费证据，不新增协议调查轮次；其他发布和目标门保持，24项Partial、release_status=not_ready。

### 6.86 第八十六批实施记录（2026-11-10）

[第86批证据](../spec/evidence/batch-86-public-surface.json)新增双module公开API文档快照（核心23包＋SCTP1包，Linux/Windows各24包）及12 Profile/20对绑定命令支持矩阵。快照明确为开发期基线，不授予稳定API或平台运行保证；SCTP Windows/386仍是stub。五CCR绑定两两冲突、DCCA debit/balance/refund、2022三态/受限退款、私有构造/ECUR及恢复值/签名消费共43项，隔离module无replace运行通过；核心CGO0/race各3577，独立消费各43。

本批只增工具、外部包测试和文档。原生产Go、11定义/1250wire/17XML不变，复用batch85的138文件核心候选与formalSCTPv0.8.0；原适配器各113项证据沿用，未在本批重跑。不生成无源码变化的新发行候选，不声称新增协议互通。

本次审查纠正“协议路径已验=使用便利层已完成”的遗漏：覆盖UX-02和架构明确的S6c纯构造/结果解释及client/s6c尚未实现，下一批优先补此本库P1；不得借“应用负责路由”删除纯协议解释义务。Ro恢复的可访问签名/Checkpoint仍不代替五操作外部应用接线。具体有限顺序和出口见[支持矩阵](diameterkit-support-matrix.md)，只归并原义务，不增加需求ID或重开规范调查。24项Partial，release_status=not_ready；batch Markdown继续忽略。

## 7. 遗留项归属与P0结束条件

2026-11-03完成这一次收尾，P0前置文档设计结束。后续没有“补完所有P1再开始实现”的门槛。历史R编号、严重度和验收ID保留，但必须按下表理解其归属；不把全部历史P1合成新的文档待办。

| 归属 | 有界工作 / 关联记录 | 关闭位置及证据 |
| --- | --- | --- |
| 实现任务 | 生成模型、codec、所有权、静态/应用校验及选定字段/结果；R24/R29/R31/R34的首版切片 | 总设计P1/P2及P4/P5：编译/生成、codec及四应用行为测试；规则和测试对应，不以文件存在算通过 |
| 实现任务 | 缺Session、必需叶、非法/重复Session或Proxy、拒绝额度不足等错误出口；主流程§27.4 | P2/P3：验证可发错误、不可安全构造时的既定关闭/终止及一次完成；不要求先在文档为所有命令物化全部错误pair |
| 实现任务 | Peer/关闭、发号、去重、故障转移、预算、TCP/TLS、SCTP；R03/R17–R21/R32/R36及R33通用切片 | P3：时序/竞态、内存/额度、发送证据及真实Linux association；SCTP U1–U5/原验收义务保持 |
| 发布验证 | 固定首版A/B流程、选定SMEA、Sh混合订阅、具名兼容、DRA/独立端点、压力与性能；R29/R34/R35的验证部分 | P4/P5/P6/P7：可复现测试与独立解释/对端证据覆盖声明能力；真实sms-fw/ip-sm-gw现网投产另验，不作为所有库能力的外部认证服务 |
| 受限能力 / 后续证据 | CR0069涉及的Realm争议、普通无IMSI成功SRA、Requested=Maximum的直接规范裁决 | 保持§3.1边界及既定接收报告；外部正文不可得不阻塞普通子集。扩展支持前再取得证据并审查；不自动轮询或持续扩大规范搜索 |
| 后续能力 | 16条未启用应用边界、完整Sh深层/国际化、限定S6a、完整SCUR/PRICE_ENQUIRY及独立RFC8506全实现 | 不纳入首版完整语义支持；保留分析、字节、来源及原验收ID，进入对应后续版本时处理 |
| 实现阶段定值 | API命名、默认容量/资源系数、测量后的性能目标和发布材料 | 在P1/P2、P3及P7分别确定；不能以“待定”为由增加公开框架或持久provider |

2026-11-04按用户要求继续首版Ro的协议收敛，具体边界见[主流程§28](diameterkit-mainflow-audit.md)，固定五包验收见[覆盖§35](diameterkit-usage-coverage.md)。此次范围修订已将Ro前移到v1，不重开全量首版P0调查。Ro操作与结果、5011接收、固定OCS活动exchange恢复转入对应实施验收；标准SMS失败/在线序号及标准5011自动生成保持明确限制；六模板兼容待完整编码与对端证据。旧RO-01–06义务保留，不以标准子集可发布宣称旧计费路径已完整替换。

本基线有24个SPEC需求、111叶/439行矩阵、423行定向反例（389行在位），公共适用行都有反例；剩16行为按需范围。现有1157段hex是候选语料，非1157项已通过的测试。既有错误样本勘误继续生效，不要求实现前对每段语料独立认证，不用439/439或所有外部CR关闭作为P0退出条件；发布时仍必须对声明能力完成对应验收。

P0退出的三项交付已经具备：§3.1固定范围；§7完成遗留项分流；[总设计§14](diameterkit-design.md)固定六个首版内实施里程碑与退出条件。**本次收尾后不再安排下一轮前置specification补齐。** 2026-11-06已按授权开始编码，最后架构审查固定的[包依赖及首批切片](diameterkit-architecture.md)正在落地，实际证据见§6.1–6.86；不重复P0。实施里程碑用通过的可执行验收结束，未通过时修复对应范围；新增可选需求不自动延长当前阶段。

### 7.1 四应用首版复审：剩余有限缺口（2026-11-04）

此次复审覆盖15份文档的范围、API、运行时、Profile/字段、样本、部署与验收索引；没有重新认证规范或执行协议栈。阶段标签不是实现证据；现有439行矩阵/1157段hex也不是四应用全部用法的覆盖证明。Ro按覆盖§36–38单独验收，不重新扩写一张全Ro逐叶矩阵。

| 优先级 / 既有归属 | 当前缺口及状态 | 有界解除条件 / 阻塞范围 |
| --- | --- | --- |
| P1；R29/R34、FIELD/TEMPLATE | **设计已收敛，接收/目标定义待证**：§3.7/主流程§34固定发送Vendor/路径、PSP逻辑路径及默认1的应用边界、token父实例、Getter/常量区别与420窄组合/隔离。PSP三层Vendor/flags/OCS语义、旧token接收定义/费用链和实际420部署仍未证 | 仅收主流程§34.4三组材料，并入原六模板包；覆盖§40十项子验收。已明codec/表示/绑定可实施；未证PSP不能启用，token不能猜映射，420不能动态切型；对应兼容及v1完整声明仍受阻，不再全量扩写字段矩阵 |
| P1；TEMPLATE-01/02、COMPAT-01 | **六模板实际兼容证据未齐**：标准/ENHANCE/INITIAL/TERMINATION/REFUND/SCAP1的真实槽/版本映射、完整旧编码和答案；SCAP1恢复规则不能继承CCR | 每模板按主流程§31.5一个有限证据包；独立解读及目标测试验证。阻塞相应兼容/v1完整声明，不阻止已明codec开发；同栈模拟器/模板加载不替代证据 |
| P1；RO-06、REPLAY、BIND | **恢复契约已定，目标条件未证**：共同去重键/内容比较/保存期、错误后T=0续发、重启原身份，以及ECUR连续性/许可 | 运行时§1.7 + 部署§2.4 + 覆盖§38的16组；库机制先离线实施，缺证禁对应目标恢复。不能以同Realm/DRA广告或新Request替代，不能标RO-06完成 |
| P1（标准完整声明）；DECISION/RESULT/RO-03/04 | **既有标准未决**：SMS失败3409/序号3408跨表、正常CCA实验5011伴随基础码 | 标准解除只接受直接依据/勘误；目标材料只能授权具名兼容。失败TERM/退款实际用法仍须兼容承接；5011安全接收可做，标准OCS自动生成可明确受限，不阻塞现有CTF全部实现 |
| P1（对应能力）；S6C/NODE/TIME/DECISION | **原三应用限制仍在**：真正无IMSI普通成功SRA、Realm标准争议组合、Requested=Maximum时间等号 | 维持已定的受限构造/保真接收及具名Realm兼容；直接依据或指定兼容证据仅解除对应分支。若真实必需调用命中受限分支，须解决其用法，不能泛称三应用完整已覆盖 |
| P1；PEER/ID/FO/VAL/BUD/SCTP | **机制未实现/未测**：公共栈和RO-06原记录/标识/Tr/竞态/资源；SCTP上游U1已有基础待回归，U2消费前分流/flags、U3更新deadline/Context、U4有限通知/选项未齐；本库组帧/预算/屏障/Peer传播另需实现 | 依部署§4.8先验上游U1–U5候选，再验本库P3/P6/P7；codec/TCP及可控适配器测试可先行。覆盖§42有限用例，Go1.26.x/CGO=0逐module门槛仍按§41；不扩大前置文档调查，也不以底层测试代替Diameter互通 |
| P2；PERF/UX/RELEASE | **发布证据未测**：四应用混合负载、Ro MSCC/模板、恢复guard驻留、峰值内存/公平性、API与示例 | P6/P7等价基线、fuzz/race/压力/互通及资源实测；不得为“先发布”移除Ro负载/关闭校验。API拼写与数值默认在实现原型定，不新增协议框架 |
| 应用迁移门槛；BOUNDARY/RO-06 | **Go binary接线/保存/切换未验证**：业务fallback、队列失败、发送意图、防回滚、旧worker隔离及最终未知核对 | 应用按主流程§33.1/覆盖§38演练；库必须提供并测试必要入口/结果/示例，但不提供数据库。应用真实投产与库v1协议覆盖分别报告，不能把业务存储完成设为公开库的必需依赖 |

本轮已修正三类交叉缺项：Ro专用API/广告与首版范围的冲突；P5未进入P6/P7及性能/示例的漏项；RO-06资源/分段边界未明（Q现按所有可选目标的最大编码长度预留，单段MaxAttempts=1不误套普通多attempt模式的至少2限制，见运行时§1.7.2/§1.7.6）。未发现需要引入持久化或推翻现有架构的新P0问题；这不表示实现正确性已经证明。

**结束边界：** 本表就是当前待办，不追加全Release/全AVP/全Ro状态机审查。已有确定规则可实施；私有定义不明只停该分支，目标证据缺失只停该启用/声明；v1整体仍须覆盖已承诺四应用。新直接证据或实现暴露具体矛盾时定点回改，当前不写代码。

## 8. 已采纳的协议细化与历史统计

以下细化规则保留效力，按§3.1范围应用；其中轮次统计仅为当时快照，不是当前待办或发布进度。规范事实未在本次收尾重新核验。现行任务归属统一见§7。

第二十二轮的细化不新增SPEC ID：SPEC-PROFILE-01/02、SPEC-VALID-01、SPEC-RESULT-01及SPEC-SCOPE-01/02继续约束这些分支。标准SRA成功MUST NOT自动回显请求HSS-ID或补造IMSI；不支持该成功分支MUST NOT自动映射为用户未知。Realm兼容MUST显式固定在binding/Profile中，MUST NOT按收到的内容自动切换；它只授权所列结构解释，不授权连接节点。Sh有限混合子集MUST将712只解释为Ref25的一次语义；MUST NOT删除/管理持续订阅，或把两个显式SNR合成原子操作。验收见[覆盖§24](diameterkit-usage-coverage.md)，场景和负例见[候选§21](diameterkit-protocol-vectors.md)。


第二十三轮以[逐叶与结果清单](diameterkit-field-inventory.md)细化SPEC-PROFILE-01/02、SPEC-VALID-01、SPEC-RESULT-01、SPEC-APP-01和SPEC-DUP-01，24个ID保持。所列三应用字段MUST保留零值Presence与结果命名空间；Sh完整typed深层能力、国际化语义和raw承载MUST分开声明。首版已选Sh数据为24/25及11/12有限子集；Repository/iFC/DSAI完整helper及全IDNA/SASLprep未纳入，不作为该子集的无条件前置，不能因raw能往返宣传全Sh或全国际化。旧调用中额外PublicIdentifiers不自动变成Ref24更新能力；有限builder按所选字段构造，通用raw不删改。验收FIELD-14、RESULT-06、XML-03、TEXT-01均Planned。


第二十四轮以[主流程§20](diameterkit-mainflow-audit.md)细化既有SPEC-PROFILE/VALID/RESULT/APP/CODEC要求，不新增SPEC ID。普通HSS双节点解释MUST结合原SRR与跨组条件；MUST NOT把单组合法当整答完整通过。Time零值MUST保留Presence并按NTP回绕解释，MUST NOT以错误宽度生成零时间或用绝对协议时间自动延长exchange。未知VLR等业务事实MUST NOT 被视为对端缺字段；本库MUST NOT为补这些事实增加持久化/查询服务。相等重传时间和未启用Sh触发保持范围报告，完整候选不等于已完成实现或独立互操作。

第二十五轮以[主流程§21](diameterkit-mainflow-audit.md)细化同一组SPEC要求：Sh实际使用特性的请求628 MUST发M1，纯接收报价例外与答案M0分别处理；M位宽容接收MUST NOT变成错误发送默认。5011 MUST有对应M1特性失败和答方支持集证据，不认识628的基础5001 MUST保留Failed-AVP。库MUST NOT自行降级/拆分/重试或缓存终端能力。T4的OFR/OFA MUST保持SGd应用身份，供计费条件成立时顶层External-Identifier必需；MUST NOT因此调用Ro或查询身份存储。双节点能力SRR MUST带SM-RP-SMEA，数字地址格式与SC-Address分开。局部字段超配置额度 MUST归ResourceLimit，不伪造协议结果。独立语义/运行验证之前，候选及111叶反例MUST NOT记作完整符合性通过；24个SPEC ID不变。

第二十六轮[命令与父路径矩阵](diameterkit-command-path-matrix.md)细化既有SPEC-PROFILE/VALID/CODEC/RESULT要求：校验MUST逐层使用Vendor+Code并在直接父实例内统计数量，MUST NOT用递归全树搜索、其他命令或相邻父组的值补足当前必需叶。覆盖报告MUST区分在位字节、定向反例、应用条件和独立执行证据；正常/E错误形状、禁止路径、未启用语义和具名兼容分开。矩阵MUST NOT 被视为新公开API/DSL或全规范支持承诺，24个SPEC ID保持。

第二十七轮[应用路径与P1契约](diameterkit-mainflow-audit.md)继续细化现有需求，不新增SPEC ID。SH访问键校验 MUST 按Ref与组合执行：External-Identifier作访问键时MUST NOT 附加Private Identity；三种访问键的独立正例MUST NOT 被合成为同时带多键的有效消息。PNR验证MUST NOT 要求库内存在订阅记录。SRA身份MUST NOT 由请求/答案字符串相等或数字外形推断其真实业务种类；Realm兼容MUST NOT 掩盖其他已证明的字段错误。时间相等与Time=0存在性分开，MUST NOT 静默改时间或延长exchange。SMEA已选数字子集的三段格式、数字数与长度、填充及1..15位规则见主流程§23.6，其他TON/NPI MUST 按现有支持范围分类；此设计依据不等于互操作Passed。矩阵169行有定向反例，270行仍待补，其中16个未启用应用边界按需，不能宣称439行全部符合。验收沿用[覆盖§29](diameterkit-usage-coverage.md)，Ro纳入首版实施。

第二十八轮[九答案族契约](diameterkit-mainflow-audit.md)细化SPEC-ERROR-01、SPEC-RESULT-01、SPEC-BIND-01、SPEC-EXCHANGE-01，不增需求ID。接收MUST先按实际ConnectionID/HBH找候选并比较App/Command/E2E，未匹配的头MUST NOT消费其他pending的完成权或A；完整头匹配后的无效包络/Session/来源按ErrInvalidAnswer终结。九族原请求有Session，E错误MUST回显它；答案MUST清R/T并回原P，禁止Destination。E语法MUST有基础结果，MUST NOT套正常XOR或以“只允许3xxx”拒绝RFC6733允许的基础5xxx；补充实验结果不得覆盖基础失败；Sh已定义实验码与268并存MUST按固定Profile拒绝，未知Vendor补充含义保留范围报告，见§24.2。Failed-AVP引述内容MUST NOT再次作为应用载荷验证，外围字段/边界/预算仍须验证。已合法读入的旧attempt候选继续服从R19完成门。验收沿用[覆盖§30](diameterkit-usage-coverage.md)，实现及互通Planned。

第二十九轮[主流程§25](diameterkit-mainflow-audit.md)细化SPEC-CODEC-01、SPEC-PROFILE-01、SPEC-VALID-01、SPEC-ERROR-01和SPEC-BUDGET-01，需求ID不增删。已知Grouped MUST按直接父实例检查；未知M0 MUST NOT猜测内部类型或递归，未知M1按请求/答案已有路径处理。Failed-AVP的外层边界MUST可信；内部坏长度/短证据MUST有界保留并报告可解读范围，MUST NOT按其声明长度分配或扫描重同步。明确的外层无效答案、诊断不完整和ResourceLimit MUST区分；MaxDepth/MaxAVPs仍约束已识别诊断树，opaque字节仍计owned与字段预算。CR0069只登记元数据；取得正文/批准状态前MUST NOT扩大Rel18 Realm标准域。验收见[覆盖§31](diameterkit-usage-coverage.md)，运行Planned。

第三十轮[主流程§26](diameterkit-mainflow-audit.md)细化SPEC-CODEC-01、SPEC-VALID-01、SPEC-ERROR-01、SPEC-RESULT-01、SPEC-ID-01及SPEC-DUP-01。Origin-State MUST归属于当前Origin-Host；答案MUST NOT盲目复制请求值，缺省/显式0的Presence MUST保留，非零跨重启保证归部署。Route-Record MUST保序/逐实例检查，MUST NOT被当成Proxy回显或新增来源授权；顶层逐跳剥离仍按R17。结果父组MUST逐实例验证，不能拼合分裂的Vendor/Code。Error-Message MUST检查UTF8及Diameter禁止的U+0000，MUST NOT作为重试/路由指令；有界显示不改raw。CR0069正文未取得，身份/相等时间能力边界不扩大。验收见[覆盖§32](diameterkit-usage-coverage.md)，24个ID不变，运行Planned。

第三十一轮[主流程§27](diameterkit-mainflow-audit.md)细化SPEC-PROFILE-01、SPEC-CODEC-01、SPEC-VALID-01、SPEC-ERROR-01、SPEC-BIND-01及SPEC-DUP-01，不增需求ID。Proxy/SF子项MUST按精确Vendor/父实例检查；接收必需项MUST NOT用CER/本地默认填补。Destination MUST同时满足命令、过程和固定配对策略；ALR的CCF可选Host MUST NOT解除所选HSS过程的必需Host，realm-only/别名MUST显式授权，MUST NOT自动改写为DRA目的。准入别名MUST NOT自动并入Canonical等价类。可确认Session缺失与已有Session/Proxy无法合法回显MUST区分；无法安全构造错误时沿用保守关闭路径，不声称所有负例均能生成有效5005。验收见[覆盖§33](diameterkit-usage-coverage.md)，运行Planned。

第87批已完成原UX-02的S6c纯构造/解释与Route薄客户端，详见[接入契约](s6c-routing-helpers.md)及[验收](../spec/evidence/batch-87-s6c-helpers.json)。保留纯Go、无持久化；其他24项库级需求仍partial，release_status=not_ready。此前第86批的缺口描述为历史状态。

第88批新增[公开Go恢复接线](../examples/ro/README.md)及[验收](../spec/evidence/batch-88-ro-recovery-consumption.json)，补齐同OCS DIRECT_DEBITING恢复、五业务操作本地映射、Tx晚答/取消与导入拒绝。SYSTEM_FAILURE采用明确保守迁移并保留诊断。内存交接不证明持久性，其他绑定/跨目标/私有与现场门保持；24项partial/not_ready不变。

第89批完成[combined公开双目标消费](../examples/ro/README.md)：debit/balance/refund/initial/term真实TCP、E3004后显式backup续段、原身份及结果保留、去重域/TERM连续性/记录和额度拒绝；见[验收](../spec/evidence/batch-89-combined-cross-target.json)。这是具名绑定及错误续段切片，不覆盖所有Profile/无答跨目标/重启或现场去重。既有恢复测试补本地writer释放等待，SGd双向示例等待两端Ready；不修改运行时许可。24项partial/not_ready保持。

第90批完成[独立绑定及无答跨目标/导入公开消费](../examples/ro/README.md)：smsfwro IEC/INITIAL/TERM、legacy420扣费/余额/退款，以及这两个绑定和combined的错误/无答同Node跨目标与新Node导入；保留原身份、T位、累计额度、晚答隔离与24项导入拒绝。见[验收](../spec/evidence/batch-90-binding-recovery-consumption.json)。仅使用进程内记录，先关闭并加入旧Node；不证明持久化崩溃恢复或真实OCS共享账本/会话连续性。所选本地消费项完成，私有/现场/发布门和24项partial/not_ready保持。

第91批完成[双module checkout发布检查](release-checks.md)：固定Go/workflow、11份生成一致性、两平台API快照及漂移拒绝、双module纯Go/race/vet和六项跨构建；见[验收](../spec/evidence/batch-91-release-checks.json)。临时workspace验证当前核心与adapter，不下载未发布batch85核心、不改依赖锁。已完成WSL本地与workflow语法验收，未推送或触发托管CI；独立发行安装、长fuzz/压力峰值/性能与现场/私有证据保持开放。24项partial/not_ready不变。

第92批完成[有界fuzz、压力与性能基线](load-checks.md)：5个目标各30秒/2worker，固定TCP窗口和SCTP工作量的纯Go/race各20轮，分列逻辑峰值/采样HeapAlloc/进程VmHWM，8项benchmark按200次×5样本记录。见[证据](../spec/evidence/batch-92-bounded-load.json)和[初始基线](../spec/release/performance-baseline.json)。这关闭所选本地有界验收切片，不证明穷尽安全、生产容量或性能提升；托管CI、真实发行安装、现场/私有门及24项partial/not_ready保持。

第93批实现[版本/API政策与真实安装门禁](version-install-policy.md)：v0双module独立标签、API变更评审、两个空缓存消费者、GOWORK=off、无replace、精确依赖图及commit/checksum/API/公开消费检查。工具和checkout消费测试通过；实际官方代理下载核心/adapter候选均404，远端无标签，见[证据](../spec/evidence/batch-93-version-install.json)。本批保持partial_pending_publication，不把门禁实现计为安装验收完成；先发布核心，再更新adapter真实依赖并发布，最后重跑冷安装。纯Go、无持久化，24项partial/not_ready保持。

Batch93 publication completion: immutable core v0.1.0-rc.2 and adapter v0.1.0-rc.1, real GOWORK=off/no-replace independent cold consumers and hosted CI passed. See [publication record](releases/v0.1.0-rc.md) and [evidence](../spec/evidence/batch-93-published-install.json). English publication guides and AI headers add no protocol behavior or persistence. Broader24 requirements remain partial.

Batch94 mixed application acceptance: Sh/S6c/SGd/Ro share one TCP connection with independent finite classes/ports. Each application stalls in turn while healthy classes complete work; cancellation, late answers, refusal accounting and zero cleanup pass in 20 pure/race repetitions. See [workload and commands](load-checks.md#mixed-application-isolation) and [evidence](../spec/evidence/batch-94-mixed-pressure.json). Production sources, API and persistence boundary are unchanged. Private/target and deployment-capacity gates remain open; all 24 requirements remain partial.

Batch95 combined duplicate acceptance: the fixed combined binding now has six construction paths with normal/E cached results, per-arrival envelopes, content/cross-template conflicts, immutable request/result ownership, finite admission/expiry and Respond/Close coverage on TCP/TLS. Twenty pure/race repetitions and dual-module regression passed. See [English guide](combined-duplicate-acceptance.md) and [evidence](../spec/evidence/batch-95-combined-duplicates.json). No runtime/API/persistence change; deep diagnostics, broader mixed-profile/cross-peer cache and private/target/capacity gates remain open. All 24 requirements remain partial.
