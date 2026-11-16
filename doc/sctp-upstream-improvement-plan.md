# SCTP通用能力下沉：diameterkit交接需求

> **执行状态更新：** 本交接已在上游bb739b612088实现，并于[第66批](../spec/history/batch-66.json)完成diameterkit迁移和回归。以下原计划/提示词保留为历史交接依据，不再是未实施任务。

原交接日期：2026-11-09。以下为第65批之后、当时尚未实施或升级依赖的设计记录；当前完成状态见页首和第66批。用于另一个session改进`skill215/sctp`，随后由diameterkit调用并删除重复逻辑。对应原U1–U5和SPEC-TRANSPORT-01 / SPEC-BUDGET-01 / SPEC-CLOSE-01，不增加新的库级需求编号。可直接使用[实施提示词](sctp-upstream-session-prompt.md)。

## 1. 结论和边界

上游负责SCTP socket生命周期、可取消等待、地址与选项、消息/通知及可查询的资源承诺；diameterkit保留一个薄适配器，负责Diameter报文、Peer和业务资源。目标是移除本库的SCTP等待/地址解析实现和直接socket syscall，不把Diameter协议搬进通用SCTP库。

当前最需要下沉的是：**原生Accept取消/deadline；有界建立入口；socket timeout及绑定地址公共接口；版本化资源查询；SCTP分片连续性和流边界检查。** 默认无序stream0和OrderedMultiStream/PPID0属于Diameter策略；上游已经能发送SID/Flags/PPID，不需要再造Diameter专用发送API。

本清单固定为下文A–E五项、三个交付阶段。A–D是下一次薄适配器迁移的必须项，E是移除重复消息一致性检查前的必须项，可独立晚于A–D交付。不等待SCTP全API、新平台或上游v1重写。本页P0/P1仅表示这次上游交接内部的优先级，不重开已结束的Diameter协议P0，也不阻塞本库其他独立能力。

## 2. 核对基线：已存在与真正缺少的能力

本库锁定`github.com/skill215/sctp v0.7.5-0.20261109044437-1649569dd9b8`，commit `1649569dd9b851b95664fef36f575c88a3547a78`。来源和既有验证见[第63批](../spec/history/batch-63.json)、[第64批](../spec/history/batch-64.json)、[第65批](../spec/history/batch-65.json)。本轮没有查询远端最新PR状态，不能据此重述PR目前是否合并。

本地兄弟目录`../sctp`的git HEAD为`62ad388`，有大量已修改/未跟踪文件，不能将工作区等同该HEAD或reset/覆盖它。本轮核对的10个关键文件与已安装1649569候选逐字节一致，清单和SHA256见[审查记录](../spec/reviews/2026-11-09-sctp-upstream-handoff.json)。这只证明这些文件相同，不声称全工作树无差异。上游`docs/v1-design.md`及`openspec/changes/redesign-sctp-v1/`仍须在上游session核实其状态；不自动将本次增量工作并入破坏性v1重设计。

| 能力 | 当前事实 | 本次动作 |
| --- | --- | --- |
| ReadEvent / WriteMsgContext | 已有消息/通知分流、caller buffers、EOR/raw flags、SNDRCV、单次send及不重发短写尾部 | 保留，E补连续性/流边界和可移植EOR访问 |
| ParseNotification | 已有五类typed通知、未知保真、借用tail | 保留；无需完整新通知族 |
| conn标准deadline和Context gate | 已有可更新deadline、Close解阻、hook退休；D/C/S取早 | A复用该内部机制，不复制第二个poller |
| ResolveSCTPAddrContext / DialContext | 已有纯Go解析、同ctx拨号、失败清理及交接门 | B增加有界公共门面；旧通用入口保留 |
| AcceptSCTP | 只有无ctx入口；listener没有公开SetDeadline | A新增；移除adapter的20ms SO_RCVTIMEO轮询 |
| SetSocketTimeouts / 有界绑定地址 | 缺少适合本库的公共入口；adapter直接操作fd | C补齐，不以通用raw-fd回调替代 |
| MessageIORequirements | 只有I/O布局/scratch查询；8192及地址账目主要在文档 | D增加机器可读的完整集成资源账目 |
| 多流 / PPID0 | SndInfo已有SID、PPID、UNORDERED；GetStatus已有实际入出流数 | E做实际流范围核验和证据；Diameter模式选择留在本库 |

## 3. A：listener Context、deadline和交接（P0，U3）

建议新增签名；允许上游按惯例等价命名，但必须给出最终调用对照：

```go
func (ln *SCTPListener) AcceptSCTPContext(ctx context.Context) (*SCTPConn, error)
func (ln *SCTPListener) SetDeadline(t time.Time) error
```

- `SetDeadline`只作用于listener的当前/后续Accept，零值清除；不继承给返回conn。不需要额外的SetReadDeadline/SetWriteDeadline别名。语义参考[Go 1.26 TCPListener.SetDeadline](https://pkg.go.dev/net@go1.26.0#TCPListener.SetDeadline)，AcceptContext是本库所需扩展，并非Go TCPListener已有同名方法。
- AcceptContext复用socketIO读gate、可更新deadline和poller唤醒。无周期性20ms等待、无每次启动一个阻塞accept worker再遗弃、无隐藏预接纳队列。空闲时没有定时轮询唤醒。
- 旧Accept/AcceptSCTP和新入口共享同一fd/gate；旧接口行为保持。通用上游可继续允许有ctx可取消的排队调用，diameterkit依旧仅一个Accept caller；固定资源承诺不包括任意数量的上游等待者。
- 同ctx覆盖等待gate、accept4、子socket初始化/选项读回和最终交接。任一步失败负责关闭新fd；取消单次Accept不能关闭listener。nil/pre-canceled ctx在I/O前拒绝。
- **最终交接规则**：accept4成功不等于返回成功。交接门观察到Close、ctx终止或已到期listener截止时，关掉尚未交接的子socket并返回nil；成功交接先胜出后，旧ctx取消/后来的listener Close不再关闭该conn。两者真并发时按门的胜出顺序裁决，不声称return与取消能物理同时原子发生。
- 错误身份沿已有U3：显式Close→net.ErrClosed，ctx取消/截止→对应context错误，标准listener deadline→os.ErrDeadlineExceeded；兼容socket timeout仍保留原身份。已提交的结果不被晚取消改写。固定SO timeout若存在，仍按一次Accept起点计算、不随重试刷新；新B门面固定S=0。
- SetDeadline延长/清除只可拯救尚未提交的超时；旧hook/cleanup不能清除新deadline或唤醒下一个owner。公开方法返回前，已启动的回调/借用必须退休。

验收至少包括：空闲取消后复用；已有/后续Accept的deadline缩短/延长/清除；gate排队取消；旧hook竞争；accept4前后/选项失败/交接前后取消；Close竞态；两个Accept隔离；成功后取消旧ctx连接仍可用；旧Accept回归。用可控barrier验证窗口，真实SCTP验证解阻和资源释放，不把sleep或全局goroutine计数作为唯一证据。

## 4. B：有界解析、拨号和监听门面（P0，U1/U3）

建议在现有底层API之上提供一个通用入口族，而不是让每个应用复制N/Z/D解析器：

```go
type AddressLimits struct {
    MaxAddresses, MaxZoneBytes, MaxAddressBytes int
}
type SetupConfig struct {
    Socket SocketConfig
    Limits AddressLimits
}
func NewDialer(network, local, remote string, cfg SetupConfig) (*Dialer, error)
func (d *Dialer) DialContext(ctx context.Context) (*SCTPConn, error)
func ListenContext(ctx context.Context, network, local string, cfg SetupConfig) (*SCTPListener, error)
```

这些是**拟议新增API**，不是当前上游已提供的签名；不要直接拷贝diameterkit的transport.Info、Node类型或业务policy。上游已有等价门面时复用一个，不并行保留多套有界入口。显式有界入口与旧通用解析器分开，避免悄悄改变原语法和零值兼容。

- N/Z/D按local+remote合计；先检查再Split/Clone/DNS/创建fd。支持本库当前默认16/256/4096及范围N=1..64、Z=0..4096、D=1..16384；可支持更大经审计范围，但本次验收不要求扩大。所有非optional字段显式；不以0表示无限。
- 允许sctp/sctp4/sctp6及现有slash地址语法；有界门面固定numeric port，拒绝溢出、符号、service name、空远端/空multihoming host，远端port0拒绝，local允许空/wildcard/port0。网络、文本、zone、IP和MessageOptions都冻结为owned值；小substring不得保留大输入底层数组。解析后再核对N/Z/IP家族。
- 一次显式host只选一个合适IP，显式多host才构成multihoming；不能把DNS返回服务集合自动拼成多宿主。DNS内部结果、缓存/平台resolver资源明确排除在逻辑字节保证外；若要强上界部署可只配置数字IP，而不是虚称DNS总内存受N限制。
- Dial解析两端、建立及交接共用原ctx绝对截止，不为各阶段重建完整timeout，不自动重试多个OCS/其他目的。ListenContext同ctx覆盖解析、bind/listen和交接，取消成功listener不影响已交接对象。
- 新门面拒绝用户Control/NotificationHandler，避免不可取消回调，也不偷偷覆盖输入回调；旧原模式保留。不能通过ctx检查后遗弃后台bind/resolve工作冒充完整退休。
- 新门面创建/接纳的message socket固定SO_RCVTIMEO/SO_SNDTIMEO=0，I/O走已有D/ctx。配置message options和实际选项验证仍走现有实现，不复制两套。
- builder构造是调用方资源；Node必须先取得D的Requirements并接纳，再调用Dial/Accept。上游不创建总预算manager或限制整个进程连接数。

验收包括输入边界±1、combined local/remote上限、IPv4/IPv6/zone/显式multihoming、substring/配置修改隔离、阻塞DNS取消、解析+connect同一剩余期限、迟到成功清理、失败fd清理、listener选项继承和旧入口兼容。

## 5. C：socket选项和有界绑定信息（P0，U1/U4）

推荐提供以下小型公共能力；不要求万能socket option表：

```go
func (c *SCTPConn) SetSocketTimeouts(read, write time.Duration) error
func (ln *SCTPListener) BoundAddr() (*SCTPAddr, error)
```

SetSocketTimeouts作用于SOL_SOCKET的SO_RCVTIMEO/SO_SNDTIMEO，**不同于**标准SetDeadline。负值/溢出在修改前拒绝；0清除。两个setsockopt不是原子事务，任一步失败返回可识别错误并说明可能部分生效；diameterkit在takeover失败时Close。无须热改活动操作的S：每次操作仍按进入时snapshot，方法供独占建立/接管阶段调用。无论是此方法还是等价的专用清零方法，完成后本库不得再为清零socket timeout使用SyscallConn。旧conn标准deadline可继续直接调用已有SetDeadline(time.Time{})。

BoundAddr针对B创建的listener返回**冻结的显式bind地址副本+实际分配端口**，wildcard保留wildcard，不枚举全部接口。IP/zone/端口正确；有界状态和返回副本计入D。缓存只含配置/绑定结果，不动态追踪接口。内部getsockname及ABI留在上游。原Addr/SCTPLocalAddr查询实际地址集合的语义不改变；FileListener等没有冻结输入的入口可明确返回“快照不可用”，不得悄悄返回无上界全接口集合。不要把BoundAddr误当对外可路由地址公告。

GetMessageOptions/GetStatus已经存在，本次继续返回真实配置、关联ID和实际Instreams/Ostreams；无需增加“Diameter-ready”或全局能力缓存。选项不支持/继承核验失败必须真实返回错误并关闭相应候选，不能降级假成功。

## 6. D：可查询的资源账目（P0，U5）

不能继续让diameterkit长期硬编码上游8192或复制与上游实现绑定的地址公式。保留MessageIORequirements的布局查询，新增一个**纯查询、版本化、无I/O**的集成账目入口，例如`SetupRequirements(SetupConfig) (SetupResourceRequirements, error)`。

至少给出以下独立量，名称可调整，包含/排除关系不能省略：

| 返回项 | 必须表达的含义 |
| --- | --- |
| LedgerVersion | 可记录/对比的账目版本，按具体实现/架构计算；不是协议特性版本 |
| DialerBytes | 冻结配置/端点及构造临时值的保守峰值额度；调用方共享builder不导致隐藏免费副本 |
| ListenerBytes | listener固定状态、冻结绑定信息、构造峰值、一个同步Accept owner及其hook；不含accepted conn |
| EstablishBytes | 每主动/被动候选最坏建立峰值；含解析/地址副本及**正在建立的连接同时存在**的总工作集，不含共享listener/Dialer项 |
| ConnectionBytes | 建立后上游连接固定状态+单reader/单writer的scratch/operation/hook峰值，含一个固定parsed通知值 |
| IORequirements | OOBMin/NotificationMin等现有布局；已含在ConnectionBytes的上游scratch明确标记，禁止重复相加 |

diameterkit可保守按`候选上游额 = DialerBytes + max(EstablishBytes, ConnectionBytes)`计入每候选，共享builder重复计费可接受；listener另加ListenerBytes。本库prefix/notification/OOB/discard、适配器值和核心事件/控制/业务额度另计。accepted候选可使用同一保守上限，不必为节省少量逻辑字节增加复杂转移API。

单reader/单writer、单Accept、无排队上游I/O、受控标准context等前提必须列明；额外并发等待者/无限deadline更新不由固定数值担保。原8192及第65批57728/61872是旧版本账目，迁移时重新计算，不要求新API伪造相同数值。对算术溢出/不支持平台/非法配置拒绝，不能返回0假装已支持。

验收：sizeof及scratch/调用栈生命周期分析支持逻辑系数；默认/边界值查询无网络I/O；并发峰值含listener+旧conn退休+新候选+parser；建立失败/取消及hook退出后才允许应用释放；实测分配作为证据而非RSS承诺。Go runtime/allocator/栈、内核队列和DNS内部继续单列。

## 7. E：SCTP分片连续性和流边界（P1，U2/U4）

ReadEvent已校验单次接收的边界/CMSG/截断；当前diameterkit仍在conn.read保存并比较跨片段的stream/assoc/PPID/ordered/SSN。该比较是SCTP消息完整性，应优先在上游既有one-to-one、fragment-interleave=0消息路径维护一个固定元数据游标：

- 第一个data记录建立游标；后续片段到EOR必须属于同一关联/stream/PPID/ordered属性，有序时SSN一致；无序SSN不当作标识。InfoPresent不可由stream0/PPID0猜测。不同完整消息之间可合法切换流和PPID。
- 通知不重置正在接收的用户消息。只保存固定值，不保存payload或完整消息、不按流建无限reassembly map。interleave>0继续明确不支持。
- 元信息缺失/矛盾、已消费后失去边界时保留实际N/flags及CloseRequired；读方向永久不可续读，沿现有ErrReadUnusable路径。不借此新增Diameter业务重试或自动丢包。
- 可增加`ReadResult.EndOfRecord() bool`等只读方法，屏蔽本库对syscall.MSG_EOR的依赖；原raw flags仍保留。无需一次返回整条消息的自动扩容Receive接口。错误消耗后副作用和取消未消费的区别保持。
- WriteMsgContext使用实际Ostreams核验SID，入流用实际Instreams核验；不能拿配置请求值或另一方向计数代替。实际流数快照及其有效期、重启/动态变化边界要说明；本次不支持stream reset/动态扩容。上游只报告restart事实，diameterkit终止该连接代。
- PPID仍是host-order uint32且只转换一次；0是合法通用值，是否兼容0、要求46或拒绝47由Diameter配置决定。发送SID>0、有序/无序、非对称实际流数、PPID0/46原样收发均需真实证明。已有能力仅补证，不另造“OrderedMultiStream”上游模式。

[接收EOR的依据是RFC6458 §3.1.4](https://www.rfc-editor.org/rfc/rfc6458.html#section-3.1.4)；上述校验是在既有level0限制下的本次集成设计，不是宣称RFC要求所有API都只能单游标。兼容处理不得静默破坏合法消息；如原Messages API必须保留宽松行为，应明确可选strict模式及独占读所有者，不同时引入两套独立reader。最终API必须给出模式兼容说明及新的资源账目。

## 8. 保留在diameterkit的最小逻辑

| 保留职责 | 理由 |
| --- | --- |
| 20字节头、24位Message-Length、prefix/body/discard与长度对EOR检查 | 上游只知道SCTP消息，不认识Diameter格式或申请业务预算的时点 |
| 整帧绝对截止从首data开始、通知速率上限/超限关闭 | 属于调用方协议及资源政策；上游执行传入ctx/deadline，不决定Tw |
| transport.Event投影、固定Sequence、ConnectionID绑定、首个关闭原因 | 连接代和完成语义属于Node；ParseNotification已在上游 |
| CEA/DWR发送屏障、控制流0、Session哈希选业务流、PPID兼容开关 | Diameter协议和具名部署策略；依据[RFC6733 §2.1.1](https://www.rfc-editor.org/rfc/rfc6733.html#section-2.1.1) |
| Peer握手/选举、Watchdog/REOPEN、DPR/DPA、Maintain退避 | 不能让SCTP自动重连导致绕过Diameter状态机 |
| pending/attempt、一次完成、Node配额、Arrival、取消退休 | 上游报实际传输事实；不理解会计/业务请求，也不拥有Node预算 |
| transport.Dialer/Listener/Incoming的薄接口包装 | 返回上游conn后接本库类型/所有权；不得为包装重新实现accept等待或地址算法 |

SEND_FAILED的sent/unsent/Context只是原始事实；上游不得解析返回payload为HBH、替应用确认投递或重试。通知终止后的Diameter新写准入封闭、与业务/超限首因的竞争仍在本库完成门。保留这些逻辑不违背“直接调用上游”，它们是两层协议必需边界。

## 9. 三阶段交付、删除清单和结束条件

1. **上游阶段一：A+C。** AcceptContext/SetDeadline，公开socket timeout清理及有界绑定快照；真实生命周期测试和旧入口回归。可先提交独立可审查变更。
2. **上游阶段二：B+D。** 有界门面复用第一阶段，交付资源查询和外部module示例。A–D同一候选通过后，diameterkit可迁移建立链路；不等待E。
3. **上游阶段三：E。** 固定消息连续性/流界限及非对称多流证据；更新账目。同一最终候选交付后，diameterkit删除对应接收检查，再实施具名选流/PPID策略。

C中的BoundAddr快照先在现有Listen路径支持，第二阶段B补输入上限和构造峰值；原没有冻结信息的入口返回明确不可用。这样阶段一不依赖尚未完成的B实现。

| 本库迁移位置 | 上游就绪后删除/替换 | 必须保留 |
| --- | --- | --- |
| adapters/sctp/setup.go | validateEndpoints/resolveEndpoint/validateResolved、地址冻结和S账目公式；转调B/D | 配置传值、Node需求额度映射和薄Dial接口 |
| adapters/sctp/listener_linux.go | 20ms常量/timeout轮询、直接setsockopt/getsockname、自建绑定表示 | transport.Listener/Incoming接口及Node交接契约；一caller约束可保留 |
| adapters/sctp/conn_linux.go New | SyscallConn清零SO timeout；转调C或B已保证的行为 | 本库选项策略、原标准deadline清理、adapter构造失败Close |
| conn.read | E证明后去掉重复元信息连续性/原生flag解释 | Diameter长度、PPID允许集合、帧deadline、有限缓冲 |
| config.go Requirements | 硬编码UpstreamReserve及从上游文档抄来的系数 | 本库自己的StateReserve及业务/控制/通知buffer额度 |

退出必须同时满足：同候选Go1.26/CGO0原生SCTP、独立race和既有架构/ABI矩阵；新增测试无隐藏skip；旧消息/通知/短写/关闭语义回归；外部module GOWORK=off消费、公开最终API/限制/资源账目；明确commit、版本、go.sum以及“远端可取”还是“仅本地候选”。正式发布与远端CI不能用本地replace成功替代。

随后diameterkit锁新候选、迁移上述位置，并重跑原第64/65批Base/S6c、预算、取消/Close、选举、无自动重放和REOPEN测试。**上游测试通过不是本库迁移通过**，不能先删除旧实现或提前改依赖。另一个session只做sctp，返回交付摘要后再在本库接线。

本轮不进入：one-to-many、peeloff、PR-SCTP、level2/I-DATA重组、全通知/任意选项、DNS服务发现、自动重连/业务重放、DTLS或新平台。已有支持平台真实回归不能省；新平台不是完成条件。历史第65批20ms路径继续作为当前可运行基线；本交接显式取代旧文档“无需AcceptContext”和“长期由adapter维护N/Z/D公式”的设计结论。
