# 当前支持矩阵与有界交付清单

2026-11-16，更新至第93批。这里记录**当前实现、公开消费证据、目标验收**三个不同层次；不替代[行为规范](diameterkit-specification.md)，不新增SPEC需求或放大协议范围。版本/JSON入口见[应用定义指南](diameterkit-application-definitions.md)，机器表见[支持矩阵](../spec/release/support-matrix.json)。

## 公开API与平台

[API快照](../spec/release/public-api.json)由[工具](../spec/tools/public_api_snapshot.py)读取双module在Linux/amd64、Windows/amd64、CGO=0下实际选中的公开包及完整GoDoc：核心24包＋SCTP适配1包。排除internal、main和仅测试包；包含生成类型、结构字段、公开函数/方法、别名及常量文档。它是开发期审查基线，**不是稳定API承诺或自动兼容性裁决**；源码、注释、工具链或构建选择变化均可能使快照变化。

```sh
python3 spec/tools/public_api_snapshot.py --check
CGO_ENABLED=0 go test ./examples/contracts
```

先安装adapter/go.mod中锁定依赖；工具强制GOWORK=off、GOPROXY=off、GOSUMDB=off，不联网下载。缺依赖时报错，不悄悄换版本。只有经过审查才去掉`--check`更新快照；第86批复用的batch85离线候选是历史证据；第90批使用当前源码replace，不作为已发布包安装证据。

TCP/TLS为纯Go适配；SCTP调用正式skill215/sctp v0.8.0。SCTP的Linux/amd64有真实内核证据，Linux/arm64目前只有交叉构建证据；Windows及386使用不支持stub。Windows API可见/能编译不表示能使用SCTP。固定Peer的Listener/Serve、有限诊断与扩展边界保持，不能据本表宣传无限多Peer接纳或完整协议实现。

## 12个Profile的支持和消费层次

同一Bundle只能选择一个4/272 CCR绑定，五种绑定两两冲突均由外部消费测试验证。ASR/ASA另占4/274；SCAP1是0/271和显式Vendor193/Acct19302绑定。20对请求/答案按绑定计数，包含重叠键，不是20个独立标准命令。

| 应用/绑定 | 已实现的选定范围 | 公开消费证据 | 仍受限或待验 |
| --- | --- | --- | --- |
| Base | CER/CEA、DWR/DWA、DPR/DPA及公共正常/E | 所有双Node示例 | 不是通用代理、完整Redirect/所有诊断实现 |
| Sh | UDR/PUR/SNR/PNR，Ref11/12/24/25有限XML和具名特性 | [Sh例](../examples/sh/README.md)只演示Ref24 UDR/UDA；其他已选分支见既有测试 | 无订阅存储、完整XSD或全面国际化 |
| S6c | SRR/RDR/ALR及已选普通、IMS/T4、Router身份/节点 | [S6c例](../examples/s6c/README.md)普通SRR/SRA；其他分支有根/codec测试 | UX-02已完成，见[接入契约](s6c-routing-helpers.md)；普通无IMSI成功及Realm争议不扩大 |
| SGd/Gdd | OFR/TFR及已选IMS/T4、Router关联 | [SGd/Gdd例](../examples/sgd/README.md)普通MO/MT | TPDU透明、时间相等条件受限；标准提醒走S6c |
| 标准Ro SMS | IEC DIRECT_DEBITING、无UPDATE ECUR、所选结果及2022 helper | [Ro例](../examples/ro/README.md)及本批token三态消费 | 不支持标准SMS余额/退款发送；helper不授权退款 |
| 通用DCCA | 独立单服务EVENT debit/balance/refund及退款关联 | 本批离线完整请求/答案消费；[网络测试](../dcca_test.go)沿第70批 | 非3GPP SMS/完整RFC8506，无R33/RO-06许可 |
| Ro abort | ASR/ASA独立会话引用 | Ro例 `-flow abort` | 应用选结果，无自动TERM/CCR完成或恢复 |
| smsfwro | 旧IEC/ECUR及实际失败TERM字段 | [第90批公开IEC/ECUR及四场景恢复](../examples/ro/README.md) | 私有答案、目标接受及业务消费不自动成立 |
| legacy420 | 明确旧类型、余额/扣费/退款及精确金额 | [第90批公开扣费/余额/退款及四场景恢复](../examples/ro/README.md) | 不把1292当标准2022；实际映射仍需材料 |
| combined | 一个具名绑定的五CCR模板与独立表示 | [第89至90批五形态、双目标及导入](../examples/ro/README.md) | 不根据payload猜绑定；不能替代第六个SCAP1模板 |
| SCAP1 | ACR/ACA、严格/具名旧广告、有限重复应答 | [双向例](../examples/scap1/README.md) | 实际私有ACA/恢复和目标材料未齐；不继承CCR恢复 |
| 旧SGd Alert | 显式16777313/8388648委托S6c规则 | [原运行时测试](../sgdalert_test.go)、第73–74批证据 | 非标准SGd新增命令，目标绑定须明确 |

[第86批消费测试](../examples/contracts/contracts_test.go)不读仓库fixtures、不调用internal或根包私有helper，可原样复制到独立module。43项包括12 Profile、五CCR两两冲突、DCCA结果与退款新Session、标准SMS token Absent/Present-empty/Unknown边界、旧金额/token与ECUR。这组第86批测试只检查恢复签名、无效记录及Checkpoint值；第88至90批另有真实TCP恢复流程和独立module消费证据，见Ro示例。

## sms-fw Ro原十二行用法：当前落点

沿[原覆盖§37–38](diameterkit-usage-coverage.md)核对；库能力、目标、业务接线不合并计“完成”。现有恢复根测试覆盖四个受检CCR Profile；通用DCCA、ASR和SCAP1不能继承许可。

| 原用法 | 库实现/入口与证据 | 公开消费和剩余门 |
| --- | --- | --- |
| 余额检查 | dcca.CCR；legacy420/combined.NewBalanceCheck；第70批及本批离线成对测试 | 旧PSP operation_result及目标业务消费仍需证据；余额结果不预留额度 |
| IEC预留/直接扣费 | ro.CCR、smsfwro/combined.NewIEC；标准公开例已运行 | 原reserve/charge不同入口的私有答案/token/槽映射另验 |
| ECUR预留 | ro.CCR；smsfwro/combined.NewInitial、CheckECURContinuation | 标准公开流程已验；第90批具名INITIAL/TERM公开恢复已验；不是实际OCS连续性证明 |
| 投递成功的结束 | ro.CCR TERM；smsfwro/combined.NewTermination | 标准公开TERM结果已验；用量/投递事实由应用输入 |
| 失败ECUR结束 | 具名私有TERM、0/非0用量/Status/DischargeTime | 根/codec有覆盖，本批独立消费0用量；旧业务差异/目标接受仍待，标准3409不放宽 |
| IEC退款或不退款 | dcca.CheckRefundRelation；legacy420/combined.NewRefund；ro.RefundInformationForService仅检查2022 | 新Session与显式token已验；旧1292来源/金额/是否退款由已授权应用流程决定 |
| ASR/ASA | abort.ASR/ASA；第85批公开流程 | 原Session CCR不受自动处置，应用TERM义务不下沉 |
| 六模板/Getter | 五CCR组合＋SCAP1，显式生成类型 | 源码投影不是实际六模板OCS收发证明；私有ACA/PSP门保持 |
| 答案消费 | typed CCA、top/MSCC/费用/单位/DDFH/透明token及Report | 公共例只覆盖已声明结果；完整私有Getter映射仍须限定证据 |
| 路由/槽/版本 | Bundle、CommandPolicy、RequestTargets、RoTargetConfig、显式SCAP1广告 | 本机DRA机制已有；实际槽/终端/去重域配置另验 |
| 多OCS/Tx/Tr/晚答/重启 | PrepareRoRecoveryRequest、SubmitRoRecovery、Plan/Progress/Import；[原恢复交接测试](../ro_handoff_example_test.go) | 第88批五业务操作消费、第89至90批具名双目标/无答/导入运行已验；真实目标许可/去重/连续性仍待 |
| 可用性/业务映射 | Peer、RequestInfo、Result/DeliveryEvidence、一次完成 | 旧业务错误/保存策略和迁移演练属应用；所需协议入口仍需验收，不能借边界删义务 |

## 修正历史API文档的误读

[API工作流§1–6](diameterkit-api-workflows.md)保存设计契约，里面的类型名和签名不能直接复制当成当前API：

| 历史示意 | 当前代码事实 |
| --- | --- |
| NewSRR / InterpretSRA / RoutingResult、client/s6c | 第87批已实现；query.go/interpret.go/client/s6c为实际入口；固定Profile、显式上下文和原Result保留见接入契约 |
| 包级NewCompletionPort | 实际是Node.NewCompletionPort，预算归Node |
| Completion.ID / CorrelationID / Err | 实际是Completion.RequestID和Result；错误在Result.Err，没有CorrelationID |
| RequestOptions.Destination / MaxAnswerBytes | 实际可选TargetID、Deadline、Validation；目标来自有限配置，答案上限在CommandPolicy |
| 泛化facts/存储/会话接口 | 当前只有具名有限Context；没有Store、数据库provider或业务会话引擎 |

没有CorrelationID时，应用可由单一所有者完成Submit、登记返回RequestID映射后再分派完成；并发完成消费者必须自行协调登记与消费顺序，不能假定完成晚于Submit返回。本批未新增或删除这些运行时接口；是否保留历史便利字段应在API冻结前明确，不默默宣布旧草案签名已实现。

## 有界剩余顺序与结束标准

下面六组是既有义务的归并，不是新增需求ID或每组承诺一个批次；公开库仍无持久化、无C/C++ ABI。第1组在第87批完成；第2组所选本地消费在第88至90批完成，私有/现场和未选分支仍按对应门跟踪。

| 顺序/优先级 | 归属 | 结束标准 |
| --- | --- | --- |
| 第87批已完成 | 本库S6c UX-02实现 | 在已实现SRR/SRA范围补纯构造/解释和薄Route接入，保留身份、主/Additional/组合节点、结果/诊断/原始证据；公开消费正反例通过。不增加路由策略/缓存/投递 |
| 第88至90批所选本地项已完成 | 既有Ro公开消费验收 | 标准同目标恢复、五业务映射及三个具名兼容绑定的错误/无答双目标/导入已验；不把本地切片扩大为原16组全部业务故障/现场验收 |
| 3 / P1，部分完成 | 发布工程 | 第91至92批workflow及双module检查、有界fuzz/压力峰值/初始性能基线本地通过；托管run与部署特定长时容量/统计比较另验，不按测试条数承诺性能 |
| 4 / P1 | API/版本发行 | 第93批已明确v0/API政策并实现真实安装门禁；实际核心/adapter下载404，待发布核心、更新adapter真实依赖并发布后重验 |
| 5 / P1特定能力门 | 外部六模板/私有证据 | 固定原槽/版本/输入输出、PSP/ACA/1292、420/2018接受与恢复域/保留期/ECUR连续性材料，再验相应用法；不阻塞无关已明能力 |
| 6 / 部署验收 | 目标与Go业务应用 | DRA/HSS/MME/OCS授权、保护条件、实际恢复许可与应用迁移；本机双Node不代替现场，业务账本/存储始终归应用 |

普通无IMSI成功SRA、争议Realm组合、时间相等、完整Sh国际化/深层模型、完整SCUR/PRICE_ENQUIRY和S6a继续按既定受限/后置范围，不为本次收尾重开规范调查。固定Peer监听、诊断覆盖等已知限制必须留在支持声明并在发行审查核对原承诺，不能由API快照自动判为全量通过。

[第86批证据](../spec/evidence/batch-86-public-surface.json)：核心CGO0/race各3577项，独立消费各43项；API快照重复检查一致。生产源码/定义/1250wire/17XML未变，沿用第85批SCTP各113项证据而未重跑。24个SPEC仍Partial，release_status=not_ready；这个结论不等于完成率。

第87批验收关闭上述S6c本库实现项；其余五组仍待办。当前API快照为核心24包＋SCTP1包。已有batch85离线候选不包含新helper，不作为本批消费证据。

第88批完成[公开恢复接线与消费](../examples/ro/README.md)：同Node错误续段、新Node导入及Tr无答重放，CHECK/CHARGE/RESERVE/COMMIT/CANCEL是原业务五操作；不等同于五种协议请求。原Result/发送证据不改写。其他绑定/跨目标的公开消费及外部私有材料继续单列；见[验收](../spec/evidence/batch-88-ro-recovery-consumption.json)。

第89批完成[combined公开双目标消费](../examples/ro/README.md)：debit/balance/refund/initial/term真实TCP、E3004后显式backup续段、原身份及结果保留、去重域/TERM连续性/记录和额度拒绝；见[验收](../spec/evidence/batch-89-combined-cross-target.json)。这是具名绑定及错误续段切片，不覆盖所有Profile/无答跨目标/重启或现场去重。既有恢复测试补本地writer释放等待，SGd双向示例等待两端Ready；不修改运行时许可。24项partial/not_ready保持。

第90批完成[独立绑定及无答跨目标/导入公开消费](../examples/ro/README.md)：smsfwro IEC/INITIAL/TERM、legacy420扣费/余额/退款，以及这两个绑定和combined的错误/无答同Node跨目标与新Node导入；保留原身份、T位、累计额度、晚答隔离与24项导入拒绝。见[验收](../spec/evidence/batch-90-binding-recovery-consumption.json)。仅使用进程内记录，先关闭并加入旧Node；不证明持久化崩溃恢复或真实OCS共享账本/会话连续性。所选本地消费项完成，私有/现场/发布门和24项partial/not_ready保持。

第91批完成[双module checkout发布检查](release-checks.md)：固定Go/workflow、11份生成一致性、两平台API快照及漂移拒绝、双module纯Go/race/vet和六项跨构建；见[验收](../spec/evidence/batch-91-release-checks.json)。临时workspace验证当前核心与adapter，不下载未发布batch85核心、不改依赖锁。已完成WSL本地与workflow语法验收，未推送或触发托管CI；独立发行安装、长fuzz/压力峰值/性能与现场/私有证据保持开放。24项partial/not_ready不变。

第92批完成[有界fuzz、压力与性能基线](load-checks.md)：5个目标各30秒/2worker，固定TCP窗口和SCTP工作量的纯Go/race各20轮，分列逻辑峰值/采样HeapAlloc/进程VmHWM，8项benchmark按200次×5样本记录。见[证据](../spec/evidence/batch-92-bounded-load.json)和[初始基线](../spec/release/performance-baseline.json)。这关闭所选本地有界验收切片，不证明穷尽安全、生产容量或性能提升；托管CI、真实发行安装、现场/私有门及24项partial/not_ready保持。

第93批实现[版本/API政策与真实安装门禁](version-install-policy.md)：v0双module独立标签、API变更评审、两个空缓存消费者、GOWORK=off、无replace、精确依赖图及commit/checksum/API/公开消费检查。工具和checkout消费测试通过；实际官方代理下载核心/adapter候选均404，远端无标签，见[证据](../spec/evidence/batch-93-version-install.json)。本批保持partial_pending_publication，不把门禁实现计为安装验收完成；先发布核心，再更新adapter真实依赖并发布，最后重跑冷安装。纯Go、无持久化，24项partial/not_ready保持。
