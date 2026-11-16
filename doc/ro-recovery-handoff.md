# RO-06：Go应用交接与有界验收

2026-11-09，第57批。本文承接[运行时§1.7](diameterkit-runtime-contracts.md)、[主流程§33](diameterkit-mainflow-audit.md)及[第56批导入契约](../spec/history/batch-56.json)。范围限定现有四个Ro CCR Profile；SCAP1仍为SingleAttempt，标准CC-Session-Failover未因具名TERM兼容而实现。本页不定义第二套恢复协议，也不增加生产API。

可运行示例见[ro_handoff_example_test.go](../ro_handoff_example_test.go)，真实TCP/TLS验收见[ro_handoff_test.go](../ro_handoff_test.go)。示例全部为测试代码，应用策略不会编译进发布库；它们只使用公开的库类型/方法，测试夹具负责建立连接。示例JSON及SHA256用于演示确认回执绑定完整意图，**不是**推荐的存储格式、认证、可靠提交或库提供的事务协议。无持久化需求的应用可仅持有内存记录，不需要这层确认门。

## 1. 一次有界交接

1. 应用配置固定Profile、原Origin、目标证据、恢复限制和预算，调用`PrepareRoRecoveryRequest`。返回的原记录不发送、不创建完成事件；保留其完整所有权，不在提交期间修改slice/计划。
2. 应用构造本段`RoRecoveryPlan`及`RoRecoveryProgress`，保存**原记录+确切计划+完整预留额度**的发送意图。导入时还包含Import声明/Checkpoint。应用保证只由一个owner串行处理本记录的业务决定。
3. 应用确认同一意图已可靠提交。示例`handoffReceipt`将确认绑定到完整意图摘要，拒绝未知、明确拒绝及任何记录/目标/截止/额度变化；换计划必须取得新的确认。摘要比较不能代替应用自己的存储/所有权协议。
4. 确认成功后显式调用`SubmitRoRecovery`。接纳成功才有新的RequestID和唯一Completion；接纳拒绝不发送、不交付Completion。应用收到确定的拒绝可以记录该事实；拒绝结果未可靠保存且随后崩溃时，仍按“可能接纳”保守处理，不能根据旧数据库里没有RequestID退回额度。
5. 活动段只使用既有`RequestInfo/Retransmit/Cancel`控制，Retransmit受Tr/传输状态、writer退休、目标许可及剩余额度检查。每段先预留其全部额度，未用额度不返还；库不启动存储回调或自动恢复worker。
6. 应用消费唯一Completion，分别保存真实协议答案/Report/发送事实与`RoRecoveryResult.Checkpoint()`。错误后的新段沿用原业务Tx截止及恢复年龄，重新保存意图；不把每个新段当作新的业务等待窗口。
7. 已知最终成功或不允许续发的失败，以NextTrigger=0保存终态。需要新的扣费/退款是新业务操作，不能把它改成Unknown复用原身份。

## 2. 提交未知与崩溃未知不能合并

| 状态 | 应用动作 | 库侧效果/本批证据 |
| --- | --- | --- |
| 存储提交未知，尚未获准发送 | 停止调用Submit；在应用存储协议内确认原意图或停止处理 | `ExampleNode_SubmitRoRecovery_commitUnknown`及`TestRoHandoffCommitUnknownAndExactIntent`：无Node调用、无Outbound/Outstanding；收到相同意图确认后才发送 |
| 已确认意图，但提交/写出结果未可靠保存 | 先隔离旧sender，再把该意图**整段**视为已使用，保存下一段意图 | `handoffCrashCheckpoint`保留SegmentNo/ReservedThrough/旧Delivery，FrameBytes取冻结最大帧，Uncertain/MayHaveReachedPeer/ReplayPending=true，NextTrigger=NoAnswer、Unknown=true |
| 确实未发送或确实已发送，但恢复方无法区分 | 两种情况均采用上述保守状态；允许浪费额度，不允许回退 | `TestRoHandoffCrashConsumesReservedIntent`在TCP/TLS分别模拟两种历史；两次额度均保留，下一次累计为3、T=1 |
| 原记录已结清成功或永久失败 | 保存原Checkpoint，不调用崩溃未知构造 | `TestRoHandoffKnownTerminalCannotRestart`：真实2001/4012终态在新Node仍拒绝导入，无保留槽和发送 |
| 结果已知但保存结果的提交未知 | 停止新段，由应用查询/确认；保留内存中的已知结果，不能为方便重放把它改成Unknown | 存储协议为应用证据；示例不声称验证了数据库或跨进程事务 |

`handoffCrashCheckpoint`只演示“已确认意图可能被接纳”的路径，不是任意记录修复器。只有应用确认历史事实、旧sender隔离与可信时间后，才能填写三项Import声明。新Node不能识别应用伪造的旧进度或被故意抹掉的终态；这些不是扩大库内历史缓存就能解决的事。

## 3. Tx、Tr和晚答

业务Tx截止依据首次段`RequestInfo.AcceptedAt+Tx`；配置要求0<Tx<Tr，恢复段保持原业务截止。`ResponseWait`是每个attempt从排队计起的Tr，`ReplayUntil`是原记录最大年龄。应用不得把业务Tx用作CompletionPort.Receive的最终停止条件而遗留协议结果，也不能用Tx重置库deadline。

默认示例：Tx到期只结束业务等待，继续消费Completion。单owner的`handoffBusiness.finishOnce`演示业务只通知一次；后台真实CCA保留其结果，不第二次唤醒调用方。`TestRoHandoffTxDoesNotCompleteProtocol`验证Tx后、Tr前仍不授予重传权，晚2001仍交付原Completion。

旧TIMEOUT_RETRY_ENABLED=0作为独立应用选择：Tx时显式`Cancel`，继续消费取消Completion，之后的晚答不承诺保留。测试验证两种选择都只有一个Completion。Node.Close/Port.Close不应用来代替业务Tx，也不被视为旧请求未发送的证明。

`RoErrorSMSFWBeforeTx`的ErrorContinuationUntil必须沿用原业务截止。`TestRoHandoffLateErrorKeepsOriginalTxDeadline`验证截止后3004仍可给出ErrorAnswer，5012不再取得Tx前额外集合的续发许可。库的CanContinue只说明协议上允许，应用仍自行决定是否保存或继续。

## 4. DDFH与旧业务分支

本批示例仅覆盖具名sms-fw **DIRECT_DEBITING**；不把DDFH应用到余额查询、退款、ECUR或SCAP1。只有适用且已通过原请求/来源校验的正常CCA，才能提供远端DDFH；缺省与显式0/1区别保留。无答案、E包络或非法答案不能捏造远端值覆盖配置。库返回真实字段，不决定收费、放行、排队或补偿。

| 旧分支 | 应用承接 | 证据/边界 |
| --- | --- | --- |
| 首发无Ready | 映射route-down；旧普通流程不自动入队 | batch52/56已有接纳拒绝与回滚；应用不能将拒绝当缓冲成功 |
| Tx前合格错误、无备用，DDFH=0 | 保留实际失败并尝试保存 | 本批真实5012、Absent/0/1×配置0/1测试 |
| 同上，DDFH=1 | 映射业务临时错误，仍尝试保存 | 与上一分支一样不把真实5012改成2001；保存失败须独立报告 |
| Tr无备用/最终无答 | 用配置DDFH：0保存，1结束不保存 | `TestRoHandoffNoAnswerUsesConfiguredDDFH`验证真实Tr无答及未自动创建新请求 |
| Tx后晚答 | 3002/3004/3005保存；合法终态只结束后台处理 | 保持原Tx、真实结果和唯一业务通知；DDFH=1不取消该临时失败保存分支 |
| 旧队列一轮 | 无Ready/无答/3002/3004/3005保留；其他已验证终态结束该项 | “处理完”不等于“扣费成功”；未知结果/关联失败交由应用处理，不自动删除；无队列实现 |
| REFUND、CHECK、ECUR及wrapper错误映射 | 按[主流程§33.1](diameterkit-mainflow-audit.md)五操作矩阵交接 | 保留原义务；本DIRECT_DEBITING示例不宣称承接全部应用业务 |

应用自行将真实协议结果、传输无答、保存失败以及业务决定分开记录。示例`handoffNoBackup`只返回“请求应用保存”的决定，不执行保存，也不以测试中的内存结果证明持久队列成功。

## 5. 退出清单与剩余工作

库侧交接示例的有限退出条件已满足：

- 三个可执行Example与七组真实TCP/TLS验收通过；确认门绑定确切意图、未知历史不退额度、已知终态不能恢复。
- Tx后正常收答/显式Cancel及一次业务通知得到验证；原业务截止、DDFH Presence和两种无备用分支未混淆。
- 未新增生产类型、依赖、持久化provider、业务worker、协议字段或报文模板。第50–56批的原身份、预算、段、Tr、错误、目标及导入证据继续有效。

应用/部署上线验收仍需提供：确切意图和结果提交未知时的确认协议；旧sender隔离及多worker所有权；历史进度防回退和可信时间；原Origin/号域与启动流量隔离；目标去重/保留期及TERM会话连续性；Tx/DDFH及五操作业务映射、保存失败处理和真实OCS故障演练。**这些是应用/部署交付证据，不是待加入本库的存储功能。**

RO-06库侧的已选CCR本地机制和交接示例在此结束，不再因应用存储实现缺失而扩张本轮。首版仍未就绪：SCAP1恢复、标准会话failover、私有字段/实际OCS证据、SCTP和原有协议余量/压力验收仍单独追踪。下一批转入已冻结的[SCTP上游实施清单](diameterkit-sctp-upstream-contract.md)：先核对U1纯Go基础及U3 deadline/取消生命周期，在同一候选完成有界测试；U2/U4及本库适配器随后接入，不重复开展全局规范设计。

第88批新增[公开Go恢复接线](../examples/ro/README.md)及[验收](../spec/evidence/batch-88-ro-recovery-consumption.json)，补齐同OCS DIRECT_DEBITING恢复、五业务操作本地映射、Tx晚答/取消与导入拒绝。SYSTEM_FAILURE采用明确保守迁移并保留诊断。内存交接不证明持久性，其他绑定/跨目标/私有与现场门保持；24项partial/not_ready不变。
