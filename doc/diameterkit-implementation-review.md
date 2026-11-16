# 实施进度与缺口扫描（第86批更新；保留第67批原扫描）

> 第86批当前结论：[支持矩阵](diameterkit-support-matrix.md)将实现、公开消费、外部目标分开。API快照24公开包/两平台，本批43外部契约检查通过；已发现S6c UX-02纯解释/构造和client/s6c仍未实现，应作为下一本库P1。Ro五操作/恢复的公开消费与私有目标材料另列，不能把签名检查当完整接线。库生产源码不变，仍用batch85候选；[证据](../spec/evidence/batch-86-public-surface.json)。后文第67批扫描中的“当前/未实现”属于当时结论，最新状态以本矩阵和progress为准。

> 第85批更新：独立ASR/ASA公开接入及应用决策、同Session CCR独立性、预算/取消/关闭通过，见[证据](../spec/evidence/batch-85-ro-abort-example.json)。无库内会话、自动TERM/恢复；下一有界出口为P7支持矩阵/API覆盖收尾，明确已选Ro剩余消费证据，不再扩展协议调查。完整发布与真实目标门保持。

> 第84批更新：无UPDATE ECUR公开INITIAL/TERMINATION、显式应用续发/原身份与新exchange关联、预算和取消关闭通过，见[证据](../spec/evidence/batch-84-ro-ecur-example.json)。库行为/协议定义不变，下一有界出口为独立ASR/ASA公开流程；其他P7和目标门保持。

> 第83批更新：标准Ro SMS IEC公开CCR/CCA、顶层/服务级结果、固定终端和精确单位/资源生命周期通过，见[证据](../spec/evidence/batch-83-ro-example.json)及[示例](../examples/ro/README.md)。既有协议/运行时不变，batch在忽略目录。下一有界出口为无UPDATE ECUR公开流程，P7其他门与真实OCS证据保持。


> 第82批更新：SGd/Gdd MO/MT公开TCP接入、方向/固定终端/结果/资源及独立消费module已验，见[证据](../spec/evidence/batch-82-sgd-example.json)和[示例](../examples/sgd/README.md)。协议与运行时不变，batch记录保存在忽略目录。下一有界出口为标准Ro SMS IEC公开例，其他P7与真实目标门保留。


> 第81批更新：S6c GMSC/HSS公开SRR/SRA例与四类结果、身份/目的/原请求关联、预算/取消/清理和独立消费module通过，见[batch81](../spec/evidence/batch-81-s6c-example.json)。协议字段/运行时行为不变；下一有界出口为SGd/Gdd公开例，剩余P7和真实目标门保持。


> 第80批更新：Sh Ref24公开双向TCP例、显式XML校验边界和独立消费module已验收，见[batch80](../spec/history/batch-80.json)。各应用版本/定义入口见[指南](diameterkit-application-definitions.md)。现有库行为/协议切片不变，下一有界出口为S6c SRR/SRA公开例，其余P7与目标证据门保持。


> 第79批更新：SCAP1完整公开配置及双向TCP运行例已交付，独立无replace消费module、结果/广告/预算/取消/清理验证完成，见[batch79](../spec/history/batch-79.json)。公开注释历史错误已修正，库运行行为未改变。其余应用示例、完整API/CI/fuzz/压力/性能与公开发行仍待办，不把SCAP1例通过记成整个P7结束。


> 第78批更新：SCAP1具名旧双Auth/Acct广告与已有ACR/ACA有限缓存/多Arrival机制已完成，见[batch78](../spec/history/batch-78.json)。默认标准校验不变，私有ACA/对应恢复及六模板目标材料仍未齐；下文“旧广告/多Arrival未实现”是第67批历史结论。1250wire/17XML，新离线核心batch78有126文件，SCTPv0.8.0不变；下一批推进P7双向示例/API核对，不猜私有字段或复制CCR恢复。


> 第77批更新：既有模拟器输入的有界接受/拒绝和修正路径已验，见[迁移矩阵](diameterkit-simulator-migration.md)及[batch77](../spec/history/batch-77.json)。生产代码/11定义不变，1233wire/17XML，核心仍离线batch76。此前“P4剩输入决策”项关闭；SIM-01旧应用迁移与真实目标对照未完成，不将整个P4标为结束。下一有界代码批次为SCAP1已明广告兼容/多Arrival；私有定义/六模板及P7门保留。下文第67批扫描及逐批状态按时点阅读。


> 第76批更新：SRA Router强制HPLMN身份及3311=1仅MCC/MNC查询已实现，见[batch76](../spec/history/batch-76.json)。S6c slice.22、1205wire/17XML；P4剩模拟器输入有界决策，以下旧批次“Router/MCCMNC未实现”结论以本更新为准。

> 第75批更新：SGd Router腿TFR/TFA时间/token输出已实现，显式目标/Alert能力与原请求关联已验，见[batch75](../spec/history/batch-75.json)。1178wire/17XML，SGd slice.15；剩余P4是SRA Router/MCCMNC用途和模拟器输入边界。下列旧结论按其批次阅读，不再把TFA token输出作为未实现项。

> 第74批更新：Router→GMSC ALR/ALA的原token/SC/端点证据已实现，标准与旧App均验证，见[batch74](../spec/history/batch-74.json)。1148wire/17XML，S6c slice.21；共享线路语法放宽后仍需角色/用途校验。剩余P4是SGd TFA token输出、SRA Router/MCCMNC用途及模拟器输入边界，不笼统重开全部Router提醒。

> 第73批更新：旧SGd ALR/ALA的显式固定绑定已实现，复用S6c字段/角色/正常与E操作，见[batch73](../spec/history/batch-73.json)。现11生成定义加1委托Profile、20对绑定、1125wire/17XML；标准SGd和S6c不变。旧SGd提醒的库绑定缺口关闭，Router token/MCCMNC、模拟器特殊输入、Ro/SCAP1及发布出口保持有界待办。后文未实现结论按原扫描时点阅读。

> 第72批更新：OrderedMultiStream/PPID0库策略及真实模式验收完成，见[batch72](../spec/history/batch-72.json)。默认行为、核心生产代码及上游v0.8.0保持；后文“具名模式未实现”为第67批历史结论。目标接受、公开发行与P4/Ro/SCAP1余量继续按原清单跟踪。

> 第71批更新：标准SMS IEC/I/T顶层及单MSCC4010/4011/5030、2022原IEC有界提取/回送检查已实现，见[batch71](../spec/history/batch-71.json)。现11定义、19组请求/答案定义、1099wire/17XML；单服务叶级检查不代表完整SMS退款发送。第67批扫描正文及后续批次说明按各时点阅读，最新状态见[progress](../spec/progress.json)。既定结果/2022项已关闭；下一项转P4/具名传输余量，未证SMS失败/5011/私有兼容不扩为前置spec调查。

> 第70批更新：通用DCCA单服务EVENT查询/扣费/退款落地，见[batch70](../spec/history/batch-70.json)。现11定义、19组请求/答案定义（含重叠272绑定）、1053wire/17XML。标准SMS Profile的未选结果、2022与服务父组、私有字段/目标证据和发布仍待办；本批未扩大旧恢复许可。

> 第69批更新：RO-05库机制已在独立`app/ro/abort`落地，定义/正常与E/关联、入站/多Arrival/取消关闭/TERM竞争见[batch69](../spec/history/batch-69.json)。当前为10定义、18组请求/答案定义（含重叠CCR绑定）、982wire/17XML；后文“ASR未实现”是第67批历史结论。应用TERM接线和目标OCS验收未关闭。下一优先级是既定标准Ro操作/结果，再处理P4/SCAP1及部署发布出口。

> 第68批更新：Relay-only协商和有限逐请求目标表已实现并验收，见[batch68](../spec/history/batch-68.json)。后文第67批扫描缺口为当时状态；目标表之外的动态发现和其他Ro/P4/发布缺口未由本批关闭。

日期：2026-11-09。范围：当前两module、九份生成定义及输出、运行时/传输/应用规则、测试与第1–67批证据、spec需求及既定首版用法。此次是实现评估，不是第68批实施或完整安全审计；未改Go代码、未重新认证外部规范/OCS，也未重跑刚完成的第67批suite。先核对397个源快照哈希不变，再引用对应测试证据。机器摘要见[审查记录](../spec/reviews/2026-11-09-global-progress.json)。

## 结论与完成尺度

公共运行时、codec及默认TCP/TLS/SCTP已形成可用的开发基线，四应用都有实质实现；**尚不能宣称四应用完整首版或完全替换sms-fw现有用法**。主风险已从SCTP上游转到DRA接入边界、应用子集、Ro/SCAP1实际兼容及发布验收。上游v0.8.0已发布和接入，不再阻塞。

67个完成批次不等于67个完整功能，也不等于整体完成率。24个SPEC是跨度很大的不变量，当前23个partial、1个planned，无法据此计算可信百分比。测试项数包含子测试及fuzz种子，也不是协议用法分母。建议按下表及已有覆盖§37十二行、§38十六组分别判断库机制/目标证据/应用接线。

| 原里程碑 | 当前判断 | 已有能力与未关闭出口 |
| --- | --- | --- |
| P1/P2 模型与codec | 主干已实现，应用子集未齐 | 9定义、17组请求/答案定义（包含重叠兼容Profile），强类型生成、owned、关联、E包络、Canonical、有限Report；不是17种互不重叠的标准命令。ASR/ASA及部分已承诺应用范围仍缺 |
| P3 运行时/传输 | 主干已实现，DRA场景有具体缺口 | 预算/DRR、一次完成、pending、Peer/选举/关闭、Route、去重/多Arrival、Dial/Maintain/Serve、TCP/TLS及SCTP v0.8.0；Relay-only协商和同Peer按请求目的绑定需补 |
| P4 Sh/S6c/SGd | 普通及多种IMS/T4子集已实现 | Sh四组、S6c三组、SGd两组及Ref11/12/24/25 helper；既定Router token/MCCMNC、旧SGd提醒/模拟器和适用兼容仍须按原清单落实 |
| P5 Ro/兼容 | 多数基础机制已实现，实际全覆盖未完成 | 标准SMS IEC/无UPDATE ECUR、旧五模板组合及SCAP1线路、四CCR Profile受检恢复；ASR/ASA、部分标准helper/结果/特性及SCAP1恢复、私有接收定义和目标证据未闭合 |
| P6 独立互通 | 有离线解释和本库真实网络证据 | 929wire/17XML，独立go-diameter字典/overlay检查；本库对本库真实TCP/TLS/SCTP不等于真实DRA/OCS或旧栈等价证明 |
| P7 发布 | 尚未完成 | 核心/adapter公开可安装版本、API/依赖锁、可复现CI、最终四应用fuzz/压力/性能、双向示例和支持矩阵仍需闭合 |

第67批实际结果：核心CGO0/race各2603项，adapter各71项，有限生命周期31项×20=620；零失败/测试skip。全量adapter含长REOPEN。两module vet/build、无Cgo依赖和三个目标交叉build通过。见[原始摘要](../spec/evidence/batch-67-sctp-v080.json)。该证据证明声明切片，不证明当前缺失能力。

## 必须优先处理的代码缺口

### P1：DRA Relay-only协商

[peer.go](../peer.go)的`commonApps`只把远端具体App-ID放入集合，再与本地具体App-ID求交；没有将远端Relay广告映射为本地显式启用应用。`handshake`在无交集时主动失败、被动返回5010。因此，对端仅广告Relay而未逐个广告Sh/S6c/SGd/Ro时，现有路径不能建立应用Ready。此结论来自代码分支，未在本轮新增网络复现；它对应既定DRA-01，不是新增DRA产品范围。

有界出口：分别验证主动/被动Relay-only、具体应用交集、真正不相交、本地未启用应用拒绝；固定Profile及终端Origin授权保持。不能由Relay广告推导终端特性，也不能顺带放开SCAP1特殊广告。应优先于可选流模式开发。

### P1：同一DRA连接上的逐请求目的绑定

[config.go](../config.go)的CommandPolicy固定DestinationHost/Realm，[node.go](../node.go)拒绝同Peer同命令重复policy；[exchange.go](../exchange.go)按policy完成并核对请求目的。[route.go](../route.go)也要求相同终端绑定。因此当前普通发送只能表达固定命令目的；无法在同一Peer/命令下用两份SRA结果为两次TFR选择不同MME Host/Realm。

这不是“邻接Peer和终端字段完全混淆”：二者已有分离；缺的是受限的逐请求选择。RoTarget有独立受检恢复表，不能据此认为普通SMS路径已经具备。绕过方法不应要求应用伪造Peer身份、修改冻结配置或复制无界连接。

有界出口：固定本地授权策略/有限目标选择，冻结每次请求目的和答案来源；同DRA连续/并发向两个已授权终端发送、越界拒绝、晚答错源拒绝、重传目的不漂移、预算接纳及取消通过。具体接口在对应实现批次确定，不新增通用发现/路由框架。

### P1：Ro实际必需操作尚不完整

九份`spec/*.json`没有ASR/ASA或274命令；标准[app/ro/rules.go](../app/ro/rules.go)只允许EVENT的DIRECT_DEBITING及选定INITIAL/TERMINATION。标准退款/余额动作不能因为旧兼容Profile支持而标成标准支持；标准4011等结果也不能从兼容Profile的处理推断已覆盖。特性/实验5011是安全接收子集，尚无完整共同特性/构造选择能力。多MSCC服务匹配、未选结果和深层诊断按实际首版输入有界补齐，不扩成全DCCA。

ASR/ASA先做：完整请求/成功和失败应答、原请求关联、有限Inbound、由应用明确选择2001/5002/5012、重复/取消/关闭/在途TERM。库不自动终止业务、退款或创建会话存储。标准退款需要原答2022的Absent/Present-empty/Unknown三态，不能由raw字节存在推断退款授权。

### P1：P4实际兼容余量

当前S6c代码和定义明确不支持Router token重写语义；SGd定义也排除Router token。旧SGd Alert未进入现有SGd两组命令定义。它们需对照[首版边界](diameterkit-specification.md#31-首版支持边界2026-11-04修订)和原用法逐项落实，不能因为普通/IMS成功路径已多次扩展就自动关闭P4。SMEA国际数字子集已经实现，不再笼统列“SMEA尚未实现”。完整Sh XML/国际化、普通无IMSI成功SRA及争议规范分支不自动扩大此次首版义务。

### P1：SCAP1及旧计费兼容

[application_binding.go](../application_binding.go)只允许严格Vendor193/Acct19302；旧双Auth/Acct广告不是已实现兼容。[compat/scap1](../compat/scap1)有ACR/ACA和Canonical，不能写成“没有codec/没有Canonical”；但实际私有ACA、对应恢复许可和多Arrival运行验收未齐。四CCR Profile的RO-06不能自动授权App0/271恢复。

PSP逐父层Vendor/flags/Absent语义、旧接收token与退款字段映射、六模板实际槽绑定及目标对2018混合类型/字段顺序的接受，需要限定证据包。它们同时包含定义待证和接入待验；不能猜值实现，也不能全丢给“应用职责”后标完成。

### P1/P2：传输与交付余量

- 默认SCTP路径已通过正式版本验收。OrderedMultiStream和PPID0具名模式仍未实现；模式/协商流范围、固定映射/顺序、CEA/DWR及关闭竞态属于本库。上游A–E无需重开。
- Serve目前一个listener绑定一个预配置Peer；多Peer共享入口仍是明确的API限制。优先按实际模拟器/服务部署验收，不扩成动态发现框架。
- 深层父实例诊断/安全Reject的完整适用集及SCAP1/combined多Arrival需定向测试；已有Canonical/有限Reject不等于全部应用出口已验。
- 生成freshness和Bundle指纹已实现；完整来源/依赖锁与定义差异发布工具未齐，不能写成“完全没有版本指纹”。
- 仓库当前未发现`.github` CI流程或Go benchmark函数；有4个fuzz入口及历史短时运行证据，不能称“从未fuzz”，也不能用上游SCTP fuzz结果替代当前九Profile的发布级覆盖。最终负载/峰值资源、性能对比、API快照及四应用双向接入示例需要P7证据。

## sms-fw Ro十二行用法的当前落点

本表评估原覆盖§37，不增加需求ID；“部分”不等于阻止已明子集使用。

| 实际用法 | 当前落点 | 剩余 |
| --- | --- | --- |
| 余额检查 | legacy420/combined请求和结果已实现 | PSP operation_result定义及实际消费闭环 |
| IEC预留/扣费 | 标准及smsfw兼容已实现 | reserve/charge各自旧路由/答案/token目标验证 |
| ECUR预留 | INITIAL和结果/纯关联已实现 | 旧完整会话输入/目标语义核验 |
| 投递成功 | TERM及结果已实现 | 实际业务消费及旧字段/顺序对照 |
| 失败ECUR结束 | 旧兼容字段/TERM已有 | 0/非0用量、旧OCS/晚答及显式业务差异闭环；标准3409仍受限 |
| IEC退款/不发退款 | legacy420/combined退款已有 | 原扣费答案→旧token/费用→退款请求/答案链及不发分支 |
| ASR/ASA | 未进入定义/运行时切片 | codec、Inbound、应答、竞态和应用TERM义务 |
| 六模板/Getter | 明确字段离线模板、五模板组合、SCAP1已有 | 部署槽/版本、完整旧编码及真实答案/私有语义 |
| 答案消费 | 层级结果、费用/单位/DDFH及部分退款字段已有 | PSP/旧token、未选多服务及结果分支 |
| 路由/槽/版本 | 固定Bundle/绑定及RoTarget已实现 | 实际映射/广告兼容；普通DRA逐请求目的另有上述缺口 |
| 主备/Tx/Tr/晚答 | Prepare/段/Tr/具名错误/跨OCS/导入及示例已实现 | SCAP1与完整16组适用库检查、目标许可/去重/连续性及五操作应用联验 |
| 可用性/业务映射 | Peer/连接代次、结果/发送证据及有限示例已有 | 五操作映射、旧新切换/故障演练；告警/账本仍属应用 |

## 外部证据、应用职责与后续能力

**证据门**：限定六模板输入输出/槽版本、私有接收定义、实际OCS对格式接受、恢复域及ECUR连续性。材料未知时只禁止相关分支/恢复，不停止普通库机制开发。独立测试环境或有来源脱敏材料可验库兼容，不把生产上线作为全部库发布的唯一门槛。

**应用职责**：业务Tx/DDFH决策、短信投递、订阅状态、账本、保存意图、跨进程唯一发送者及重启调度。库需提供足够记录/结果/接纳边界并验证契约，但不实现Store/WAL/provider、C++队列解码或业务恢复引擎。第57批示例只覆盖有限DIRECT_DEBITING，不能代表五操作全适配。

**明确后置/受限**：S6a、全SCUR/PRICE_ENQUIRY、完整Sh深层/国际化、全部3GPP/RFC8506、无IMSI普通SRA成功、未裁决Realm/时间相等标准扩展。维持已选接收报告和禁用规则；不因本次扫描重启全量specification补齐。代码中ErrUnsupported多为刻意范围边界，不应一律删除。

## 有界后续顺序与结束标准

沿原P3–P7收敛，以下是工作顺序，不是新一轮设计阶段，也不保证每项只需一个批次：

1. **DRA接入闭环**：先Relay-only协商，再逐请求终端绑定；用同一DRA两终端、来源拒绝/重传/预算矩阵结束。
2. **Ro必需操作**：ASR/ASA优先，随后实际首版缺失的标准退款/结果/特性和服务关联；按覆盖§36/37适用行结束，不追求完整DCCA。
3. **P4/传输既定余量**：Router token/MCCMNC、旧SGd提醒/模拟器、具名多流/PPID0及必要监听边界；逐已承诺用法验收，不自动扩域。
4. **旧计费/SCAP1联验**：所需私有定义/六模板证据与前述开发同步收集，取得证据后补绑定/恢复/消费闭环；按十二行、十六组及原十项私有验收分别记录库/目标/应用状态。
5. **发布收尾**：最终支持矩阵、双module公开依赖、API/示例、CI及最终fuzz/race/压力/性能。可先交付明确限制的v0.x；v1必须关闭既定四应用A/B用法，不允许通过改名或删目标绕过。

不建议继续用“所有SPEC仍partial”或“再补一批报文”来描述进度。下一批建议改为DRA接入，而非仅按旧next_batch优先做SCTP模式；其余外部材料不妨碍这项执行。当前审查没有提供可承诺的日历工期，也没有以测试数量估计完成百分比。

## 追踪偏差

旧覆盖§37仍写十二行均Planned，§38的当前段停在第52批；实际第53–57批已有Tr/跨目标/导入/交接。部分历史文案写“尚未实施”，部分progress.remaining把应用/目标证明混到库不变量状态。这些文字不能作为缺代码的单独依据。当前状态以本次扫描、第67批和现有代码为准；后续按原验收行补证据映射，保留历史批次，不新增第二套SPEC编号、不把本次评估记成实施批次。
