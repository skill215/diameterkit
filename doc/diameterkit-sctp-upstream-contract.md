# skill215/sctp U2–U4接口契约与有界实施清单

> 第67批当前状态：独立adapter已锁定正式SCTP v0.8.0（bce1b3cf3ea3），源码与第66批候选一致，A–E、资源和Peer边界不变；正式版本回归完成，见[batch67](../spec/history/batch-67.json)。上游发布门已关闭，核心/adapter公开发布及现场验收仍单列；下文旧候选/未合并表述按历史时点阅读。

状态：设计基线2026-11-04，Draft 0.20的SCTP实施补充；2026-11-09第58–63批完成上游候选U1–U5所需实现、资源审计、本地/远端矩阵和真实模块引用，第64–65批已固定候选完成本库有限adapter/预算/Peer接线及托管拨号监听；第66批已按A–E核验上游新候选并完成通用能力下沉及薄适配迁移，历史交接见下。正式发布与部署验收仍单列。Go 1.26、纯Go/CGO_ENABLED=0、无持久化。沿用U1–U5、SCTP-01–06及24个SPEC ID，不增加工作包。本文是[部署§4.5–4.8](diameterkit-deployment-interop.md)的具体接口交接；同主题以本页输入/输出和竞态裁决为准，不改变Diameter与上游的责任分界。

> 最新实施：[第66批](../spec/history/batch-66.json)已核验上游A–E并固定bb739b612088，完成薄适配迁移。原生AcceptContext、有界建立、socket公共接口、资源查询和消息连续性已替代本库重复逻辑。[下沉需求](sctp-upstream-improvement-plan.md)及[session提示词](sctp-upstream-session-prompt.md)保留为已执行交接；后文旧版本/待办按对应时点阅读。

> 第63批历史交付：[第63批](../spec/history/batch-63.json)已创建上游PR #30、验证真实可安装commit/伪版本及14项远端矩阵；head/CI merge tree同候选。U5开发接线所需候选门已满足，正式review/合并/tag仍单列；下一批固定该commit做本库适配器，不据此宣称本库生产可用。

> 第62批历史实施：[第62批](../spec/history/batch-62.json)完成U5本地资源账目、ABI/100轮取消退休、可重建离线安装和发布门禁接线。8192为有限reader/writer方式的上游逻辑固定额度，不是总heap/RSS或adapter总量；正式上游版本、远端矩阵和本库接线仍未完成，U5整体未关闭。

> 第61批历史实施：[第61批](../spec/history/batch-61.json)完成ParseNotification五类typed解析、未知原值/借用Tail边界，并在当前Linux/amd64真实触发五类事件；长通知超限不消费重读、最小选项不一致与accept拒绝释放均已验。312条测试/子测试/fuzz种子/示例零skip；U4最小链路及对应U2原生缺口关闭，U5完整预算/候选发布/平台和本库适配仍未完。以下批次说明按历史时点阅读。

> 第60批历史实施：[第60批](../spec/history/batch-60.json)已实现Messages、ReadEvent/WriteMsgContext、IORequirements及所需选项设置/实际读回/accept核验，接通U3等待门。284条测试/子测试/示例零skip；U4 typed通知、真实大通知等联合出口、U5发布及本库适配/完整预算仍未完成。以下历史状态不代表当前API缺失。

> 第59批历史实施：[第59批](../spec/history/batch-59.json)已接通ResolveSCTPAddrContext/DialContext，Go resolver及相同ctx总预算、Control前置拒绝、失败释放、最终交接/后取消隔离经过九组新增和全量上游测试。仍未有Messages/ReadEvent/WriteMsgContext，U3须在U2接线后继续验证；U4/U5、本库适配器及发布未完成。下面第58批状态按历史时点阅读。

> 第58批实施：[第58批](../spec/history/batch-58.json)基于远端当前v0.7.4/62ad388，在skill215/sctp补公共标准deadline、内部context gate及hook退休；真实SCTP257条测试/子测试零skip、全量race及纯Go检查通过。U3公开Context解析/拨号尚未实现，U2/U4/U5及本库适配器未实现；没有新增Message API或发布版本。e415406及“本轮未改代码”的下文记录保留原设计时点，不代表最新实施状态。上游独立v1重设计文档未由本批实施。

## 1. 固定范围与来源

只补Linux one-to-one、可靠逐消息发送、完整数据/通知接收、可更新deadline/Context取消及本库必需通知/选项。基础为本地`skill215/sctp e4154064dc9928274bcf4615c5382dd6fd81ad21`；源码/哈希见[部署§4.8](diameterkit-deployment-interop.md)。旧地址深拷贝、fd所有权、Close唤醒、reset检查及逐消息PPID实现复用并回归。本轮未在上游建分支、修改源码或执行测试。

接口采用下列拟议Go签名以便实施评审；上游可作等价命名调整，但不得省略语义或重新开放已裁决分支。没有Diameter类型、业务存储、重放队列或通用插件回调。one-to-many、peeloff的新事件模式、动态事件切换、PR-SCTP、level2、全RFC6458接口和新平台均不进入此次清单。

## 2. 最小接口面与模式绑定

| 接口 / 配置 | 输入输出契约 |
| --- | --- |
| `SocketConfig.Messages *MessageOptions` | 新增显式消息模式；nil保留原模式。非nil时在建连/listen前复制完整值并固定，accept继承。新模式与NotificationHandler和用户Control回调互斥，配置失败在创建fd前返回ErrInvalidConfig；复用的是库内部连接前设置阶段，不是让用户回调参与期限承诺 |
| `MessageIORequirements(opts MessageOptions) (IORequirements, error)` | 无I/O的固定布局查询，返回最小OOB/通知固定结构空间、peek及库内读/写scratch上界；依架构及选定布局计算，不伪称内核能力探测。供本库在拨号前计入TransportReserve，不依赖通用schema/运行时规则服务 |
| `(*SCTPConn).ReadEvent(ctx context.Context, data, notification, oob []byte) (ReadResult, error)` | 一次返回一个数据片段或一条通知，绝不跳过通知继续等业务。按§3在同一读所有权下peek分类再消费；ctx作用于此次调用，不影响另一方向。三块缓冲在调用期间由该调用独占，返回后不再访问 |
| `(*SCTPConn).WriteMsgContext(ctx context.Context, p []byte, info SndInfo) (int, error)` | 复用已有SndInfo的SID/Flags/PPID/Context/AssocID值形状；调用开始复制info。此次子集只允许Flags=0或UNORDERED、AssocID=0或当前one-to-one关联；显式编码SNDRCV及TTL=0，不继承隐藏默认TTL。空payload/其他发送控制位本次不支持，发送前报错；不是新增PR/ABORT接口 |
| `SetDeadline/SetReadDeadline/SetWriteDeadline(time.Time) error` | 实现现有签名，语义见§4；成功后影响已有及后续操作，零值清除标准deadline。返回后不能由旧操作清理撤销 |
| `ResolveSCTPAddrContext(ctx context.Context, network, addrs string) (*SCTPAddr, error)` | 与现有地址语法一致，Go resolver、总ctx及输出owned地址；不改变原无ctx入口签名，不依赖libc解析器 |
| `(*SocketConfig).DialContext(ctx context.Context, network string, laddr, raddr *SCTPAddr) (*SCTPConn, error)` | 输入为已解析地址，校验/深拷贝、选项、connect等待及连接交接均在ctx内；调用方解析和Dial传同一ctx才能覆盖端到端期限。此新入口不接受用户Control回调，避免无法取消的同步用户代码；旧Dial/Listen的原Control用法保留。历史最小切片未要求AcceptContext；2026-11-09下沉交接A已改为明确要求，见页首 |
| `(*SCTPConn).GetMessageOptions() (MessageOptionState, error)` | 返回有限选项实际读回值及选定接收/事件布局，非任意key/value能力表；没有相应内核值的事实按实现模式报告，不能假称getsockopt成功。readonly，不切模式 |
| `ParseNotification(p []byte) (Notification, error)` | 纯值解析器，只消费完整通知bytes（Header.Length须恰等于len(p)）；检查header和已知结构，输出§5的typed事实及未知原值。固定字段按值复制，变长tail明确借用p且不保留引用；需跨下一次接收保留时由调用方在预算内复制 |

Messages模式一旦创建不可切换。其`Read`、`SCTPRead`、`Write`、`SCTPWrite`及旧wrapped I/O必须在读写前返回ErrIOMode；旧模式的ReadEvent/WriteMsgContext同样拒绝，不调用原handler或发送。不能通过多个wrapper争抢同一fd。消息模式下改变事件/接收格式的旧setter必须拒绝；read-only状态查询和标准deadline仍可用。外部绕过所有权、直接读写fd或修改消息选项不属于安全契约；本库适配器不这样调用。

MessageOptions固定字段为`ReceiveInfo ReceiveInfoMode`（此次仅SNDRCV）、`Events uint32`（仅§5五类通知bit）、`NoDelay bool`、`FragmentInterleave uint8`（此次仅0）、`AllowFragmentation bool`；DATA_IO由ReceiveInfo派生，不在两个字段中重复配置。InitMsg继续使用原配置。IORequirements的OOBMin/NotificationMin为int，PeekScratchBytes/ReadScratchBytes/WriteScratchBytes为uint64且不重复计入调用方buffer；MessageOptionState返回这组实际核验选项和固定布局标识，不输出任意配置map。无业务引用，参数窄化前检查。普通旧入口签名/零值行为保持；新增公开字段对非具名struct literal的源码兼容风险须在上游API评审/发布说明记录，不能声称绝对零迁移。新模式的拒绝不改变旧模式的read/callback分派。

ReceiveInfoMode零值为Unspecified，Messages非nil但未显式选SNDRCV则配置失败；不从其他字段推模式。Events=0在通用上游表示不订阅这五类通知，但Diameter配置必须选择全部五类；NoDelay/AllowFragmentation按显式bool值处理，本库分别要求true。配置对象在创建期间不得被调用方并发修改，创建后修改原值不影响已复制的listener/connection。

本库适配接线按[运行时§3.5.1–3.5.3](diameterkit-runtime-contracts.md)及[Peer§3.6](diameterkit-state-consistency.md)：通知通过既有接收结果的固定事件分支交接，不能隐藏吞掉；不新增上游Peer回调。IORequirements不含调用方buffer，也不等于整个连接/监听器的内存总额；第60批scratch按结构大小和保守标量空间计算，Go调用栈/堆管理、error/closure及poller/context状态另计，不承诺零分配；固定连接状态、配置副本及timer/hook必须提供可核对的版本依据，由适配器纳入完整Info/构造期预算。同一上游scratch只计一次。上游仍允许非零socket timeout；Diameter适配器固定S=0并统一协调D/连接ctx，Tw不属于上游deadline。

## 3. U2：消息结果、分流与缓冲所有权

ReadResult固定字段为`Kind ReadKind`（None/Data/Notification/EOF）、`N/OOBN/RecvFlags int`、`Info SndRcvInfo`、`InfoPresent/Consumed/CloseRequired bool`和`RequiredNotificationBytes uint32`。N只指当前Kind对应data或notification区的有效长度，另一块payload区不写；实际使用范围按len而非cap。Info.Flags与RecvFlags独立，0流/0 PPID不等于Info缺失。N/OOBN均不得超出实传缓冲；收到大于缓冲的原长度只能放入独立required字段，不能拿它切slice。

调用前检查ctx非nil、模式、三块缓冲足够且不重叠；空data不做recv。缓冲最小值由IORequirements给出，通知最小值覆盖已订阅必需固定结构，不随网络长度扩张。调用期间不得由用户修改缓冲；库不保留它们，不为payload另分配整帧。ReadResult本身只含值，不持有输入buffer；ParseNotification的借用tail是另一个显式调用契约。

同一个读gate覆盖peek、判断和真实消费；仅一个当前调用占gate，其余按§4等待。peek使用固定8字节通知头空间，依据MSG_NOTIFICATION分类，不能根据payload内容猜类型；peek的EOR/CTRUNC、元数据和长度不冒充已消费结果。数据分支只消费到data区；通知分支先核对声明长度再消费到notification区。现选Linux行为下一次recv应交完整通知；若实际为不完整通知则显式失败，不隐式跨下一调用补组。新入口不跑用户handler，也不隐藏多轮通知等待。

| 情况 | 确定返回与后续行为 |
| --- | --- |
| 数据片段/完整消息 | Kind=Data、Consumed=true，N/OOBN/RecvFlags来自实际recv；EOR=false是合法部分消息，EOR=true才是消息尾。InfoPresent只表示通过所选CMSG结构校验，不表示已通过Diameter检查 |
| 完整通知 | Kind=Notification、Consumed=true、N=通知长度，实际EOR及头长度一致；不把CMSG当作该通知业务内容。单次只返回这条通知；上层自行决定继续读，消息总期限不能刷新 |
| 通知声明长度大于notification区 | peek后返回ErrNotificationTooLarge；Kind=Notification、Consumed=false、N/OOBN/RecvFlags=0，RequiredNotificationBytes给出经检查的需求，不保留peek EOR。上游允许调用方用已获预算的更大缓冲再读；本库NotificationMax固定，采用关闭策略，不循环重试或自动扩容 |
| 配置/参数/小于固定最小buffer | ErrInvalidConfig/ErrIOMode/ErrShortBuffer；不peek、不消费、N=0。不把buf不足当对端坏消息 |
| 实际recv返回TRUNC/CTRUNC、坏CMSG或不完整通知 | 保留实际N/OOBN/flags及Consumed，附对应错误；CloseRequired=true，连接进入读不可继续状态，后续ReadEvent不再recv。由调用方Close；本库关闭association。不能把已消费部分当成未消费而重复读取 |
| peek发现通知头长度非法、或peek/消费类型不一致 | 返回结构错误及已知消费状态，CloseRequired=true；不按声明长度分配或继续猜消息边界 |
| EOF、关闭、取消、超时 | EOF用Kind=EOF、N=0、io.EOF；无消费的其他失败Kind=None/N=0。若底层实际接受了零字节数据消息且有可靠消息标志，按Data/Consumed=true返回，不混成EOF；本库将其按Diameter短帧拒绝。ctx与成功I/O竞争按§4，不能一律归零 |
| 未知合法CMSG/未知通知 | 仍有界跳过未知CMSG；数据没得到所选信息时InfoPresent=false，不能补0假装有效。未知通知完整返回，ParseNotification返回Unknown及原type/flags/tail；本库按固定未知事件规则处理，不猜新SEND_FAILED布局 |

ReadEvent不提供“取消后回滚已消费字节”。成功数据片段可以先于同刻取消返回，调用方继续拼帧或关闭；如果真实消费后的元数据/边界已经无效，取消不能覆盖这个错误。需要CloseRequired的错误从此不再复用读边界，禁止忙循环或后台drain。

## 4. U3：期限、取消和一次结果的线性化

### 4.1 有效期与持有者

每个操作在登记时固定开始时刻t0和socket timeout快照S；S>0得到`t0+S`，gate等待也消耗它。有效截止为当前方向标准deadline D、ctx deadline C、固定socket timeout截止三者最早值；unset不参与。ctx取消另外唤醒。标准deadline可动态变化，C和S的剩余额度不能被D清零/延长重置。各次ReadEvent是不同操作；本库按原ReadPrefix/ReadFrameInto消息期限反复传同一ctx/绝对D，避免通知/分段刷新整体预算。

复用os.File poller和方向gate。持久D和变更代数由连接级短临界区管理；setter发布后同时唤醒相应poller/gate，不等待占用方向gate的I/O。SetDeadline在一个逻辑提交更新读写两方向，不能成功返回只更新一边；Close先提交则setter返回net.ErrClosed。底层应用deadline失败不得伪报成功，终止该连接的I/O并保留cause。

上游不创建每次排队的后台goroutine或无界请求队列；每个调用仅持有自己的固定timer/hook，连接只有一个当前读owner和一个写owner。调用方自行限制并发调用数；本库适配器保持一reader/一writer，固定临时资源纳入对应预算，不把任意外部并发堆积说成无限容量保证。

只有当前方向owner可把该操作的effective deadline应用到poller；每次setter/ctx唤醒、取gate、EAGAIN/EINTR返回均重读有效状态。旧timer或取消回调必须核对操作代数/owner；清理仅移除自己的临时ctx/S，不得无条件`SetReadDeadline(time.Time{})`清除新D。返回前移除并结清自己的timer/取消hook；不得有后来触发的旧hook去唤醒或终止下一操作。外部ctx超时回调是唤醒来源，不是独立的结果交付线程。

### 4.2 竞态裁决表

“提交”指在连接/操作状态的受控点记录终态；时间相近不靠goroutine调度顺序重复交付。非阻塞syscall的准入与结束分别是两个点，取消不能撤回已准入内核调用的真实结果。

| 竞争窗口 | 唯一裁决 |
| --- | --- |
| ctx/Close已提交，尚未获得syscall准入 | 不再开始recv/send/connect步骤，N=0并返回对应原因；排队者取消只移除自己，不关闭共享连接 |
| 已获gate，deadline更新先于下一次syscall准入 | 按新D判断；缩短到已过期则不调用，清零/延长按C/S重新取早。不得使用最初timer缓存继续睡或误超时 |
| 标准deadline曾到期，但尚未提交timeout时被延长/清除 | 重新计算后可继续；如果timeout已提交，更新只影响后续操作，不能复活已返回调用 |
| syscall已准入，随后ctx取消/Close/到期，同时内核返回真实数据或完整发送 | 保留真实n/flags/消费事实。成功且无独立I/O/结构错误时返回成功，不用ctx错误抹掉进度；未来操作再观察取消/关闭。保留这一结果不授权再发一条消息 |
| syscall返回EAGAIN/EINTR且无进度 | 重查取消/关闭/期限后才允许重试等待；取消胜出则退出，不以EINTR无限循环绕过截止。非重试的内核错误保留原cause，不改写成碰巧发生的timeout |
| 正数短写、n>0伴随真实I/O或元数据错误 | 返回n和实际错误；WriteMsgContext的无错误正数短写补io.ErrShortWrite，不发送尾部。本库按现有发送证据终结/关闭，不能改为NotWritten或重发整帧 |
| 多个无进度终止原因同时可见 | 已提交终态不变；尚未提交时检查顺序为Close、ctx取消/截止、有效D/S截止。D与S同时过期取最早截止，同刻取标准D；已确定的非重试内核/结构错误优先保留。规则用于稳定结果，不声称判定两个线程的物理先后 |
| Dial成功与ctx取消 | 完成连接/选项检查后，在同一交接裁决点检查ctx：取消先提交则关闭所有新资源且返回nil连接；交接成功后撤销/结清dial取消hook，ctx以后取消不关闭交给调用方的连接 |
| Close、旧清理、新操作交错 | Close只释放拥有的fd，唤醒全部等待者；已成功I/O仍可如实返回。旧timer/cleanup不得改新owner或复用fd。后续调用稳定返回closed，不能以后台还在读写而先返buffer |

只有ctx取消不自动关闭连接。若通知/数据已经发生不可恢复的部分消费，仍带CloseRequired交给调用方；本库可按连接级期限关闭，但单exchange取消不碰共享reader。DialContext输入已经解析，不承诺涵盖调用之前的DNS时间；适配器须将ResolveSCTPAddrContext、DialContext和Diameter握手放进同一总预算。

### 4.3 错误兼容与可判定性

错误须可通过errors.Is识别ctx.Err（Canceled/DeadlineExceeded）、net.ErrClosed、os.ErrDeadlineExceeded、ErrSocketTimeout、ErrIOMode及缓冲/结构错误；包装保留原errno。标准D超时匹配os.ErrDeadlineExceeded；socket timeout匹配ErrSocketTimeout并保留EAGAIN；ctx截止匹配context.DeadlineExceeded。timeout类满足相应Timeout语义，cancel不伪装timeout。旧API保留原socket timeout返回兼容，但新增标准deadline不能被旧EAGAIN映射误吞。旧NotificationHandler属于用户同步代码，标准deadline只控制I/O等待，不能强行打断已进入的handler；新消息模式禁止handler，不能把旧模式的回调时间也声称为硬期限保证。没有任何错误字符串参与Diameter重试决定。

按[Go net.Conn](https://pkg.go.dev/net#Conn)约束deadline影响已有/后续I/O；上表ctx/gate/原因优先级是本项目新增契约，不能归因于Go标准库。IO完成后普通库无fd/后台操作泄漏的保证不覆盖用户在旧Control回调中自造的后台任务；此次新消息/Context入口直接拒绝该回调组合，避免不可能兑现的硬取消承诺。

## 5. U4：最小通知和选项集合

### 5.1 固定接收与通知布局

本轮选定Linux现有`SCTP_SNDRCV` ancillary和旧`SCTP_EVENTS`订阅路径、旧`SCTP_SEND_FAILED`通知。它们与当前上游的SndRcvInfo/订阅基础一致；不以引入新版SCTP_RCVINFO/SCTP_SEND_FAILED_EVENT为前置，也不在收到消息后自动切布局。若锁定内核无法提供所选必需能力则不启用；未来增加另一布局须具名配置、独立正反例和版本记录，不当本轮待办。

Diameter适配器需要DATA_IO元数据和下面五类通知；DATA_IO是信息订阅而非第六种业务通知。ParseNotification的header/type/state/assoc按Linux ABI验证，PPID只做一次网络到host转换；地址按family解码，端口/地址及scope保留，不解析DNS。未知state保留数字，不猜成成功。固定结构字段owned；可变tail借用且不得自动解码成Diameter。 第61批固定PARTIAL_DELIVERY为当前Linux24字节布局，stream/sequence分别标Presence；16/20字节前缀拒绝而不伪造0。ParseNotification错误返回零值与ErrMessageBoundary；它不持有SCTPConn，已消费的坏通知由适配器关闭。

| 通知 | 上游必须交付的事实 | 本库固定使用方式 |
| --- | --- | --- |
| ASSOC_CHANGE | type/flags/length、state/error、in/out streams、assoc ID及有界tail | COMM_UP仍须能力交换/状态机；COMM_LOST、CANT_STR_ASSOC、关闭和RESTART走既定连接代终止处理，不自动重放 |
| PEER_ADDR_CHANGE | 远端地址/family/port/scope、state/error、assoc ID；不以unsafe直接投影未校验的sockaddr | 仅更新路径事实；单路径不可达不直接判整个association或OCS失效 |
| SHUTDOWN_EVENT | assoc ID及完整结构 | 停止所有新写并结清/Close；不是Diameter DPR，不回DPA、不启动ClosingGrace；不解释成对端已处理此前业务 |
| PARTIAL_DELIVERY_EVENT | indication、assoc ID；已选ABI提供的stream/sequence有独立Presence | 本库对ABORTED一律关闭association；one-to-one/level0不依赖是否已观察到部分帧或缺失的stream/sequence猜身份，不能补0伪装在位 |
| 旧SEND_FAILED | error、DATA_UNSENT/DATA_SENT原值、完整sndrcv info（含Context/PPID/stream）、assoc ID、借用的返回payload | 只对可靠唯一关联的attempt补传输事实，无法关联给有界诊断；UNSENT不把历史成功send调用改成NotWritten，SENT不证明对端已处理；不重复完成或自动重试 |

REMOTE_ERROR、AUTH、ADAPTATION、SENDER_DRY、动态stream/reset及新版SEND_FAILED_EVENT不列必需、不默认订阅。意外收到合法未知通知仍返回Unknown，不能为了声称“typed全覆盖”误解码或无限缓存。notification缓冲不足由§3处理，未知事件不授权扩大容量。

### 5.2 选项、设置时点和失败行为

| 最小项 | 设置/核验责任 |
| --- | --- |
| ReceiveInfo=SNDRCV，事件集合 | 消息模式创建前固定；Diameter配置DATA_IO及五类事件，读回核对，不用EVENT_ALL。每次Data还检查InfoPresent，设置成功不免除实际接收检验 |
| FragmentInterleave=0 | 此次新接收契约只验level0；dial/listen在connect/listen前设置并读回，accept后核对继承。其他级别不静默取0或据此宣称支持 |
| NoDelay=true | Diameter明确请求并读回；上游bool选项可表达false，但本库配置规则不默用它。上游不决定吞吐/公平性策略 |
| AllowFragmentation=true | 对应DISABLE_FRAGMENTS=0，允许内核网络分片；不能为了user-message边界禁网络分片。初始化/accept核对；不设置并不存在的EXPLICIT_EOR选项 |
| 可靠逐消息发送 | 新WriteMsgContext固定发送完整SNDRCV信息、TTL=0且无PR/EOF/ABORT/地址覆盖等控制位；不依赖可变默认发送参数。本库选PPID46或具名0/stream，库上游只做值校验/编码。一次成功send=一条user message，无尾部重发 |
| InitMsg与Status | 复用InitMsg、GetStatus；记录期望及实际流数，窄化前校验，不把期望Ostreams当实际值。listener不查association状态冒充已连接；dial完成/accept后再查询状态及实际选项 |

MessageOptions是创建时不可变输入，GetMessageOptions返回实际核验事实；不提供热更新SetMessageOptions。Messages模式的旧默认发送参数setter也拒绝变更，避免绕过可靠发送契约；只读查询不受此限制。配置不识别的位/枚举直接报错，不悄悄剔除。连接前设置不支持/失败则清fd且不返回连接；accept核验失败关闭这个新连接，监听器可继续服务，必须保留可识别错误。本库还需在CER前完成自身准入；上游不执行CER。每条连接保存有限配置/布局版本用于诊断，无数据库或全局能力缓存。

所选ABI依据[Linux v6.12 UAPI](https://github.com/torvalds/linux/blob/v6.12/include/uapi/linux/sctp.h)；实际消费/发送路径对照[Linux sctp socket实现](https://github.com/torvalds/linux/blob/v6.12/net/sctp/socket.c)，通知语义参考[RFC6458 §6.1](https://www.rfc-editor.org/rfc/rfc6458.html#section-6.1)。这些是字段/接口依据，不证明所有目标内核都通过本契约。

## 6. 上游有界实施清单与退出标准

以下五行就是交接清单，复用U1–U5，不新增“先补全SCTP”的第六包。

| 归属 / 顺序 | 必须完成的产物 | 退出标准 |
| --- | --- | --- |
| U1 基线先验 | 保留地址/Close/reset/PPID和旧接口回归；记录锁定源码及公开API差异，Messages零值兼容设计 | 原用法无回归；后续每个新入口纳入同样生命周期检查。不是重写底层库 |
| U3 I/O生命周期 | 标准deadline状态/代数、gate与poller唤醒、操作ctx、原因映射、Context解析/拨号交接 | §4九种窗口均可用barrier/可控时钟重复验证，无旧hook串扰/无fd或goroutine泄漏；旧timeout测试与新deadline同时通过 |
| U2 接收与发送入口 | 不可变模式、IORequirements、ReadEvent的分流/结果/所有权和WriteMsgContext | §3各返回分支及真实通知>20字节、超限/peek/取消均通过；实际bytes/flags不丢；无整帧预分配、无回调吞通知 |
| U4 有限解析/配置 | 五类typed通知、未知保真、固定SNDRCV/旧SEND_FAILED、最小选项创建/读回/accept继承 | §5集合内每项有正例、坏长度/缺元数据/不支持反例，真实内核核验；不以dummy/stub跳过必需能力 |
| U5 同候选发布 | U1–U4在一个可安装版本/commit上集成，Go1.26.x/CGO=0真实I/O/ABI/泄漏、race另跑；API与限制说明 | §7全部适用上游列Passed且无隐藏skip，记录内核/架构/源码/命令和结果。上游不等DRA/OCS；本库锁该候选后另做P3/P6/P7 |

U2/U4可在契约固定后与U3并行开发，但合并验收不能各用不同候选。此表保留原设计实施清单；第58批已在用户后续授权下实施等待层基础，未将整行U3标为完成。真实SCTP适配器必须等上述候选；模型/codec、TCP/TLS、可控适配器测试仍可先行。

## 7. 验收矩阵与文档结束边界

复用[覆盖§42](diameterkit-usage-coverage.md)的十项，增加具体断言而不另建编号体系。所有用例必须有确定预期；竞态不能靠只sleep碰运气，纯函数/fault injection与真实Linux association各自记录。CGO=0普通测试和CGO=1 race工具检查分开；不写新生产C依赖。

| §42子项 | 本次补充的必过断言 |
| --- | --- |
| 1 基线 | 旧模式无行为迁移；新旧reader互斥发生在消费前；Messages+handler/Control在fd创建前失败；Close/reset及地址副本覆盖新入口 |
| 2 flags | N/OOBN上界、stream0/PPID0 Presence、peek flags不泄漏成实际EOR；合法partial Data与CTRUNC/坏CMSG分开，n>0错误不被ctx归零 |
| 3 分流 | data len=1/20、通知大于20、相邻通知与数据；RequiredNotificationBytes报告后未消费原通知、无busy retry；短/重叠buffer前置拒绝；返回后不再触碰缓冲 |
| 4 通知 | 五类型各自字段/长度/字节序，未知state和未知新版event不误解码；巨大SEND_FAILED借用/超限策略，匹配与不匹配Context不误归请求 |
| 5 deadline | §4.2九个窗口逐项，包括过期尚未提交再延长、旧cleanup清新D、旧ctx hook打到新owner、gate排队ctx取消；D/C/S取早及同刻优先级稳定，读写方向隔离 |
| 6 Dial | Resolve和Dial同ctx总预算，取消先于交接得nil并清资源；交接成功后ctx取消连接仍可用；Control拒绝，不以遗弃回调伪造取消 |
| 7 选项 | DATA_IO+五通知及实际读回，level0/NODELAY/网络分片配置；accept继承失败只关新conn；GetStatus实际流数，缺能力不回假成功 |
| 8 写入 | flags/TTL/SNDRCV明确、PPID转换一次；零/短写/正数伴错及取消前后结果，绝不发送尾部；旧SCTPWrite与新入口回归分别记录 |
| 9 资源 | IORequirements对实际scratch峰值成立，三缓冲不重叠；通知超限不扩容，排队/ctxhook释放、通知洪泛不刷新整体消息截止；本库再验TransportReserve和HOL |
| 10 发布 | 同候选纯Go构建、最低工具链、真实I/O及独立race/ABI；每项有证据链接，失败/skip不记Passed；本库互通另验，24个SPEC ID保持 |

至此U2–U4的接口、返回分支、竞态、最小集合及上游出口已给定。下一步是按清单实施和验证；文档只在实现证明契约矛盾或目标内核有具体反例时定点修订。不追加新版事件、全部选项或跨平台栈调查作为实施前置；不把文档完成记为API已存在或SCTP已可启用。

## 第65批消费说明

本库继续锁定第63批真实候选1649569；不增加上游API或改动该版本。上游Context解析/拨号用于NewDialer；其listener无公开AcceptContext/deadline setter，本库以固定SO_RCVTIMEO=20ms使用已有poller等待，循环前后检查ctx，并关闭迟到socket。不是raw accept替代、后台队列或持续忙轮询；连接交接前清零继承timeout。该可取消Accept是本库adapter能力，不宣称上游新增同名接口。N/Z/D前置限制、listener与候选额度及生命周期由本库实施，DNS内部内存与应用监听器绑定并发不在Node承诺内。完整边界和真实验证见[batch65](../spec/history/batch-65.json)。

## 后续职责调整（设计交接，未实施）

第65批没有因本次下沉计划失效；当前版本继续运行旧接纳实现。A–D经同候选验证后，再由adapter转调公开上游API，删除20ms循环、直接SO timeout/getsockname syscall和N/Z/D复制/公式；E就绪后移除重复SCTP分片元信息检查。Diameter长度、CEA屏障、选流策略、Peer/预算/一次完成继续留在本库。具体五项边界、三阶段出口和版本迁移见[交接需求](sctp-upstream-improvement-plan.md)，不增加U6或第二套SPEC。

## 第66批消费完成

A–E及迁移证据见[batch66](../spec/history/batch-66.json)。本库生产代码已无20ms循环、socket syscall、N/Z/D解析或上游8192固定额；上游CloseRequired/ErrAssociationChanged与有效通知并存的情况已接入Peer原终止门。候选公开可取不等于正式release；Diameter策略与现场验收边界保持。
