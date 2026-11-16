# 给skill215/sctp维护session的提示词

> **执行状态更新：** 本交接已在上游bb739b612088实现，并于[第66批](../spec/history/batch-66.json)完成diameterkit迁移和回归。以下原计划/提示词保留为历史交接依据，不再是未实施任务。

将下面整段复制到以`skill215/sctp`为工作目录的session。默认两个仓库位于同一父目录；不在同一目录时，将`../diameterkit`替换成实际路径。详细交接正文为[能力下沉需求](sctp-upstream-improvement-plan.md)，其中的API都是建议签名，必须核对当前源码后落实，不是已存在接口。

---

我维护skill215/sctp和diameterkit。请在当前sctp仓库实施一轮有界的通用能力增强，让diameterkit直接调用公开API，删除其目前自行实现的Accept轮询、地址解析/复制、socket syscall及与上游实现绑定的资源公式。基准Go1.26，生产CGO_ENABLED=0，无持久化。需要完成代码、文档、测试及可供另一session接入的交付摘要，不停留在建议。

首先读取：

- 当前仓库AGENTS.md（如有）、CONTEXT.md及现有工作流；检查git状态，保留所有已有改动，不能reset/覆盖未提交文件。
- `../diameterkit/doc/sctp-upstream-improvement-plan.md`，这是本次消费方契约，A–E及其验收为有界范围。
- `../diameterkit/adapters/sctp/{setup.go,listener_linux.go,conn_linux.go,config.go}`，以及`setup_linux_test.go`、`managed_linux_test.go`、`integration_linux_test.go`。
- 本仓库`context*.go`、`socket_deadline*.go`、`socket_io_linux.go`、`message*.go`、`notification*.go`和`docs/diameterkit-u2-message.md`、`diameterkit-u3-context.md`、`diameterkit-u5-resources.md`。
- 核对现有`docs/v1-design.md`及`openspec/changes/redesign-sctp-v1/`，区分拟议设计和实际API。本次不默认实施整个v1重写；按仓库工作流跟踪这个增量任务，不重复制造第二套同名消息栈。

基线提醒：diameterkit锁的是`v0.7.5-0.20261109044437-1649569dd9b8`，commit `1649569dd9b851b95664fef36f575c88a3547a78`。交接时本地sctp HEAD为62ad388但工作区含增量文件，10个关键文件与已安装候选一致。请重新核对你当前工作区，不用HEAD数字推断文件是否缺失，不重新实现已经存在的ReadEvent、WriteMsgContext、ParseNotification、conn标准deadline或Context拨号。

按三个阶段完成：

**阶段一：A+C，listener生命周期及公共socket能力。**

1. 增加`(*SCTPListener).AcceptSCTPContext(ctx)`及`SetDeadline(time.Time)`。复用现有socketIO/gate/poller和deadline更新机制；禁止20ms定时轮询、遗弃阻塞accept goroutine、隐藏预接纳队列或关闭重绑listener。旧Accept/AcceptSCTP与新接口共用所有权且保持兼容。
2. ctx覆盖排队、accept4、子socket初始化/选项核验和最终交接。取消/Close在交接门先胜出时关掉子socket返回nil；交接先成功后旧ctx取消或listener Close不影响返回conn。单次取消后listener可复用。返回前取消hook/缓冲借用实际退休。listener deadline作用于当前/后续Accept，0清除，不继承给conn；错误身份遵循既有U3。
3. 提供`SetSocketTimeouts(read,write time.Duration)`或等价公开清零能力，处理SOL_SOCKET超时，不能让消费方继续使用SyscallConn。拒绝负数/溢出，说明两次setsockopt可能部分生效，失败消费方Close；不改变标准SetDeadline或旧非零socket timeout语义。
4. 提供冻结bind快照`BoundAddr()`：保留显式IP/zone/wildcard、取得实际端口，返回owned值；不枚举全部接口。不改变旧Addr查询语义，FileListener缺少快照时明确报不可用。先在现有Listen路径保存输入，阶段二保证输入有界。

**阶段二：B+D，有界公共建立入口和可查询资源账目。**

5. 实施AddressLimits(N/Z/D) + SetupConfig + NewDialer/DialContext + ListenContext，或一个等价且不重复的公共入口族。复制/解析/fd创建前限制combined local+remote地址数、zone字节和文本字节；保留现有slash语法，有界入口只接受numeric port。需要支持默认N16/Z256/D4096及文档范围，不能用零表示无限。
6. 冻结所有输入，防止substring保留大底层存储；解析后再次检查。一个显式host选择一个合适IP，不进行DNS服务发现或自动OCS failover。解析、连接/监听、交接共用一个绝对ctx期限。新门面拒绝用户Control/NotificationHandler，旧入口保留。新建/接纳message socket使用SO_RCVTIMEO/SO_SNDTIMEO=0。
7. 增加无I/O、带LedgerVersion的资源查询。分别表达Dialer冻结/构造峰值、Listener含一个Accept的峰值、每候选Establish总峰值、建立后Connection单reader/writer峰值，以及现有IORequirements。明确包含关系，不把scratch或accepted conn重复计费。diameterkit可用`DialerBytes + max(EstablishBytes,ConnectionBytes)`保守声明候选额，listener另计。
8. 不强行保留旧8192/57728/61872数值，用实现和测试更新账目。明确单reader/writer/Accept、无上游排队的前提；DNS内部、Go运行时/allocator/栈、内核队列不在逻辑字节保证内。拒绝溢出/非法配置/不支持平台，禁止返回0假成功。

A–D同候选通过后即可交付第一份薄适配器迁移候选，不等阶段三；随后继续完成E并交付最终候选。

**阶段三：E，消息连续性与多流基础证明。**

9. 复用既有one-to-one/fragment-interleave0消息接收路径，用一个固定元数据游标检查跨片段assoc/stream/PPID/ordered及有序SSN一致。通知不重置游标，EOR结束消息；不保存payload、不做全消息复制或按stream无限重组。缺元信息/已消费边界损坏保留实际N/flags并CloseRequired，沿既有读不可用门。未消费取消不伪造EOR或清除进行中的消息。
10. 提供可移植的EOR访问方式，例如ReadResult.EndOfRecord()，同时保留raw flags。模式和错误兼容性要说明；如需可选strict模式，仍只允许一个实际reader，不另造消息队列。
11. 发送SID用实际Ostreams、接收流用实际Instreams核验，不能用配置请求值。保留host-order PPID的一次转换、单次send、部分/不确定发送证据，不能重发短写尾部。补真实非对称流数、SID>0、ordered/unordered、PPID0/46以及非法SID测试；已有基础可直接补证，不需要新造Diameter专用模式。

必须保留在diameterkit、不移入sctp的职责：Diameter头/长度、prefix/body/discard预算时点、整帧deadline策略、通知速率、Peer/Watchdog/REOPEN、CEA屏障、DPR/DPA、Session哈希选流、46/0兼容策略、pending/attempt/一次完成、业务预算、重连与业务恢复。sctp不导入diameterkit，不解析SEND_FAILED payload为HBH，不自动重连或重放。

测试要求：

- 第一阶段用barrier/fault injection明确Accept排队、deadline更新、取消/Close、初始化失败和交接前后竞态；真实SCTP证明取消复用、关闭解阻、成功后旧ctx不污染。
- 第二阶段验证N/Z/D边界±1、IPv4/IPv6/zone/multihoming、输入所有权、DNS取消/总期限、fd清理、实际选项继承及资源峰值。不得把DNS内部所有分配说成受N上界控制。
- 第三阶段验证通知插入各片段、连续性反例、合法完整消息间切换流/PPID、metadata缺失、EOR/截断及写入实际字节；复用现有消息和五通知测试。
- 运行现有完整CGO_ENABLED=0原生SCTP测试、独立race、必要的重复竞态、vet/build，以及仓库已有ABI/架构矩阵。测试未运行/环境不支持/skip分别报告，不能记Passed；交叉编译不等于真实网络验收。
- 提供独立外部module消费者示例，证明最终API可构造、查询额度、dial/listen/accept、发送/接收/取消/关闭；最终交付使用精确候选、GOWORK=off、无replace。开发时的replace只算开发便利，不是发布证据。

其他约束：不引入one-to-many、peeloff、PR-SCTP、level2重组、全通知/任意选项框架、DTLS、新平台或持久化；不修改diameterkit生产代码或提前升级它的依赖。公开struct增加字段/改变模式的兼容风险要说明，旧零值/旧入口按既有承诺回归。发现具体契约冲突时定点修正并解释，其他细节自主完成，不重开无限范围的前置规范设计。

交付时给出：

- A–E逐项完成/未完成表、最终导出API及与交接拟议签名的对应关系。
- commit、模块版本和go.sum；明确哪些已在远端可取，哪些只是本地候选，CI与发布状态不混淆。
- 资源公式/数值及包含关系、接口兼容说明、真实运行命令/结果/skip和证据路径。
- diameterkit迁移步骤：删20ms Accept循环、直接setsockopt/getsockname、N/Z/D解析器和上游硬编码账目；E通过后再删重复分片元信息检查。列出应保留的Diameter逻辑。
- 消费方仍须重跑第64/65批真实Base/S6c、预算、取消/关闭、选举、REOPEN及无业务自动重放测试；上游成功不能冒充diameterkit已迁移完成。

完成A–E及上述交付后结束本轮，不自行扩展为整个SCTP v1项目。
