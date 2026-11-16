# 定义与实施追踪

行为规范唯一正文仍为 [Library Specification](../doc/diameterkit-specification.md)。本目录保存生成输入与实施证据，不初始化 OpenSpec，不维护第二套需求 ID。

批次过程记录`batch-*.md`统一保存在本地`spec/local/batches/`，由`.gitignore`排除，不随仓库上传。以后新增批次记录也放此目录。可发布进度和验收使用`progress.json`、`history/`及`evidence/`；公开文档已链接这些JSON，不依赖本地批次Markdown。冻结历史证据保留当时的原始路径/摘要；当前进度的`local_notes_update.relocations`记录其本地新位置。

应用的规范版本、Application-ID、切片版本及JSON→Go生成入口见[应用定义指南](../doc/diameterkit-application-definitions.md)。Sh Ref24公开接入例见[examples/sh](../examples/sh/README.md)。

当前已完成九十二批，第九十三批政策与工具已实施、真实安装待发布：Base/S6c基础、公共包络与关联、出站运行时、关闭与选举、TCP/TLS、有限Inbound/应答，有限Reject/FailurePolicy和应急拒绝预算，Canonical/有限重复缓存和多Arrival重答，固定同终端Route/多attempt与活动请求原身份重传，有限Dialer/Maintain拨号及重连，按class/字节DRR调度，固定Peer的有限Listener/Serve，有界标识源、Session-Id和guard预算，Base/S6c有限Context/Report，普通S6c RDR/RDA、HSS ALR/ALA，SGd/Gdd普通OFR/OFA、TFR/TFA，以及Sh普通四组命令与透明User-Data、Ref24/25有限Sh-Data构造/解析、Sh单订阅/退订及Ref11/12数据、有限特性/混合订阅/5011及具名终端证据，以及S6c普通HSS的SGSN/MSC/Additional节点、SMEA及成对路由条件、普通MSISDN的MME/MSC/SGSN独立与组合投递报告，以及显式端终端角色绑定的MME/SGSN普通IMSI提醒，以及SGd/Gdd直达重投时间及显式era/应用证据、T4 MO计费身份及原请求绑定、S6c T4无MSISDN报告和HSS提醒、T4已知IMSI的SRR/SRA、SRA四种实验失败及三域诊断/角色UNRI、SGd OFA实验5001及接口/原MSISDN证据、S6c IP-SM-GW节点及Realm接收报告、普通MSISDN RDR的IP-SM-GW结果父组/诊断、IMS RDR/RDA的HSS-ID/目的URI及具名用途证据、IMS ALR/ALA的dummy身份及HSS提醒绑定、IMS SRR/SRA的HSS-ID/目的URI、角色UNRI和固定HSS恢复、IMS OFR/OFA的dummy源身份、SMSMI三叶、IP结果与具名转发证据、IMS TFR/TFA的顶层HSS-ID、双URI与SMS-GMSC/IP-SM-GW绑定、IMS MO可用IMSI及真实MSISDN的值/presence绑定、S6c T4报告IP-SM-GW结果与合法节点组合、Ro标准SMS IEC DIRECT_DEBITING基础、无UPDATE的ECUR INITIAL/TERMINATION及纯跨请求关联。机器可读进度见 [progress.json](progress.json)，本批状态与失败证据见 [batch-93](evidence/batch-93-version-install.json)，前批记录在 [history](history)。`completed`仅指对应有界批次，24项库级需求仍按范围标 `partial` 或 `planned`，不代表完整P3或v1。

## 已实现的切片

| 固定 Profile | 能力 | 明确缺项 |
| --- | --- | --- |
| `base-rfc6733-controls-slice-v1` | CER/CEA、DWR/DWA、DPR/DPA 正常包络；CER/CEA VSAI；所选公共诊断与各答案的独立 E 操作 | Redirect 3006、完整应用扩展、多Peer listener/admission及完整Peer事件报告与协商 |
| `s6c-29.338-ordinary-mme-slice-v1` | 普通 MSISDN SRR；2001 + 真实IMSI + 五种SGSN/MME/MSC节点形状及Additional的SRA，单IP-SM-GW的N/NH/NB/NHB（SBI0/1），Realm/保留SBI/Additional-IPSMGW有界接收及未支持报告；10415/5001、5550/5552/5556/5557 SRA及5550三域诊断/角色UNRI；选定基础失败与 E；Proxy、Route-Record、Supported-Features 结构；MTI缺省/0/1、SRR-Flags与国际十进制SMEA、成功LMSI/Alert-MSISDN/MWD位图；普通MSISDN RDR/RDA及HSS ALR/ALA，MME/SGSN普通IMSI ALR/ALA及显式角色Context、MME/MSC/SGSN/IP-SM-GW cause0/1/2（MSC/MME互斥）、按父组选择的诊断（IP组11/12）、可选Time；T4无MSISDN的IMSI SRR/SRA（成功原IMSI关联）、RDR/RDA和HSS ALR/ALA及独立用途Context，T4 RDR含IP-SM-GW原因0/1/2与诊断11/12及合法节点组合；IMS RDR/RDA的HSS-ID、目的SMSMI URI及IMSDeliveryReport证据；IMS ALR/ALA的15零dummy、目的URI及IMSHSSAlert证据；IMS SRR/SRA的HSS-ID/目的URI与IMSRouting证据、GMSC/IP-SM-GW角色UNRI；显式Router HPLMN成功token、3311=1仅MCC/MNC查询及可选节点 | 其他结果/节点/身份、无IMSI成功、未选T4分支与Realm兼容/SMSF、未选特殊路由身份和3311=0、完整 S6c 条件和特性协商 |
| `sgd-gdd-29.338-ordinary-slice-v1` | 已知IMSI OFR/OFA，MME/SGSN TFR/TFA；透明TPDU、普通MME位置、所选结果/原因及原请求目标关联；3330/3332/3331及显式era、直达角色和同PLMN/节能证据；T4 OFA顶层3111/M0及原身份/SC映射证据；SGd OFA实验5001及MOForwarding接口/原MSISDN证据；IMS OFR/OFA的无IMSI dummy源、SMSMI三叶、临时IP-SM-GW结果和IMSMOForwarding证据，以及可用IMSI/可选真实MSISDN源身份；IMS TFR/TFA顶层HSS-ID/双URI及IMSMTForwarding用途；Router已知IMSI腿的显式Alert能力、3331及3102/1 token | 未选IMS组合、T4其他身份/命令及SCTP部署、其他位置/诊断、未选Router身份/结果、未审核重传（旧SGd提醒另有显式委托Profile） |
| `sh-29.329-sms-opaque-slice-v1` | 普通Ref24 UDR/PUR、单/混合Ref11/12/25订阅与退订SNR、独立PNR；有限628/5011和具名终端证据；必需Sh VSAI、有限公共身份/MSISDN、所选正常/E结果及透明702 | 有限特性/5011与11/12/25混合已实现；完整XML/schema、其他Ref和特殊身份、未知终端动态探测管理仍未实现 |
| `dcca-rfc4006-single-event-v1` | 通用非3GPP单服务EVENT查询/扣费/退款，顶层413金额/417单位，余额0/1及缺失/未知报告，所选失败/E，纯CheckRefundRelation，TCP/TLS SingleAttempt | MSCC/2022/3GPP服务事件、私有1292、恢复及目标OCS验收；每Bundle仅一种272绑定 |
| `ro-rfc6733-abort-slice-v1` | 独立App4/274 ASR/ASA；2001/5002/5012和独立E、Session/Proxy/Header/来源关联；有限Inbound/缓存/多Arrival及取消关闭和在途TERM独立完成；与任一CCR绑定组合 | 应用会话决定/适用TERM接线、目标OCS证据；Redirect、未选字段/结果及ASR自动恢复未支持 |
| `ro-32.299-sms-iec-slice-v1` | App4/272，标准IP-SM-GW执行前IEC直接扣费；单MSCC/服务键、RSU/GSU事件uint64、退款opaque token；无UPDATE的ECUR I/T及USU/FINAL、TERM5002、IMEISV/有限PS、纯CheckECURContinuation；2001/4010/4011/4012/5030及基础失败/E、2022原IEC有界提取/回送检查、5031诊断、5011双结果安全接收与原请求关联；固定同OCS活动exchange恢复 | ECUR其他分支、退款请求、其他结果/节点/服务键、标准失败/在线序号争议、特性协商/实验5011自动配对、六模板/私有绑定、SCAP1、RO-06和现场OCS证据 |
| `smsfw-ro-iec-template-v1` | 实际IEC42节点及ECUR INITIAL41/TERM44节点，NewIEC/NewInitial/NewTermination、纯Continuation检查、旧STRING和所选正常/E关联；费用及单MSCC结果分层；5031/5011安全接收；一个固定Bundle；固定同OCS活动exchange恢复 | 通用/退款已在独立legacy420域；SCAP1另域；组合已在独立combined实现；更多答案字段/PSP/token证据、RO-06及真实OCS/TERM顺序接受 |
| `smsfw-ro-legacy420-v1` | 退款26及通用32模板节点；413/固定420分键，显式signed请求金额、uint64单位、退款0/1292；扣费/余额查询/退款、正常/E及顶层Cost/GSU/余额原请求关联；单MSCC结果分层及直接扣费DDFH、5031/5011安全接收；固定同OCS活动exchange恢复 | 组合已在独立combined实现（2018固定wire）；SCAP1另域；未知CCA token/PSP及原扣费映射、完整目标/恢复验收 |
| `smsfw-scap1-event-v1` | App0/271、实际47模板节点/11保序参数组、显式NewEvent、合成基础ACA正常/E及原记录关联；显式SCAP1Accounting绑定、严格会计能力及TCP/TLS SingleAttempt | 具名旧双Auth/Acct广告和有限缓存已在batch78完成；实际私有ACA、恢复及目标OCS验收 |
| `smsfw-ro-five-template-wire-v1` | 单一App4/272定义涵盖IEC/INITIAL/TERM/通用扣费与余额/退款，2018固定owned线路字节及模板约束；显式构造、原请求答案分支关联、纯Continuation和TCP/TLS；费用/DDFH及单MSCC结果分层、5031/5011安全接收；固定同OCS活动exchange恢复 | 目标对2018混合表示及固定字段顺序的接受，实际私有CCA/token/PSP、实验5011自动配对、RO-06及现场验收 |

字段、Vendor、父路径、Presence、重复数量和编码顺序由 [base.json](base.json)、[s6c.json](s6c.json)、[sgd.json](sgd.json)、[sh.json](sh.json)、[ro.json](ro.json)、[roabort.json](roabort.json)、[dcca.json](dcca.json)、[smsfwro.json](smsfwro.json)、[legacy420.json](legacy420.json)、[scap1.json](scap1.json)、[combined.json](combined.json) 定义；规则函数在对应包 `rules.go`，来源及人工规则修订号随定义记录。现有 Profile 明确为开发子集，不能直接替代 DRA 部署使用的完整 Profile。标准应用接收到未选择字段时，未知 M=0 保留 raw、M=1 返回 `ErrUnsupported`；smsfwro/legacy420/combined/scap1固定模板正常报文则关闭Extra，未知M0也拒绝；保留 raw 不表示理解其语义。已知字段按固定类型解码，不因 M 位变化猜类型。

S6c 正常基础失败目前选定 5001/5004/5005/5007/5008/5009/5012/5014；未选结果/MTI仍显式返回 `ErrUnsupported`，不映射成用户未知。E 走独立基础语法，详细支持与保留限制见 [本批契约](history/batch-02.json)。SMS 的 260 接收仅检查外层后按 raw 保留并忽略语义，发送禁止；公共 UTF8String 同时检查编码和 U+0000。

`Decode` 成功仅表示本 Profile 的静态线路检查通过；`ValidateAnswer` / `DecodeAnswer` 再检查原请求与显式来源集合。`CompleteAnswer` 从已验证请求补齐 Session/Proxy/头标识，本地 Origin 必须显式给出。已提供基于原始request的codec.Canonical；现有切片已提供ValidateContext/ValidateAnswerContext和runtime Report，已有有限ShPeerFeatures、S6cAlertRole、SGDRetransmission、T4Charging、S6c T4NoMSISDN、S6c RoutingRole、MOForwarding及IMS报告/提醒/路由非零具名Context，Ro基础CCR/CCA及四Profile的普通R33同OCS活动exchange恢复已实现；RO-06已提供原记录准备、有限段接纳/无答续段、Tr、具名错误续段、授权跨OCS及显式重启导入/RecoverySpaces；交接示例已在第57批验收；标准会话协商、SCAP1恢复和实际应用/部署证据仍未完成。普通SRR和已选CCR已接入Route；RDR/ALR、OFR/TFR、SCAP1和四个Sh请求仍不授予TransportReplay。Sh CheckContext/Report只覆盖明确的AVP与特性子集，不包含XML/schema或业务验证。独立`app/sh/shdata`现提供Ref11/12/24/25有限XML构造、解析与显式单/混合数据Check报告，范围见[batch-20](history/batch-20.json)，不会隐式扩大codec报告。

SRA5550当前完整校验需要RoutingRole（普通MSISDN）、T4Routing或IMSRouting证据；UNRI不是SRR-Flags。三域诊断仅5550组合在本批支持，其他组合未支持；零/缺失/保留值分开，详见[batch-28](history/batch-28.json)。

SGd OFA实验5001的完整校验需要MOForwarding显式接口/原身份/端点；Gdd/T4报告未支持，原请求缺MSISDN不可补证据绕过，见[batch-29](history/batch-29.json)。

S6c IP-SM-GW节点使用Optional Number/Name/Realm/SBI四叶；Additional同名字段仅供有界接收。Realm与保留SBI不因补Context而完整通过，标准发送拒绝，具名兼容未启用；ALR不扩大。见[batch-30](history/batch-30.json)。

普通MSISDN RDR的IP-SM-GW结果已支持cause0/1/2及cause1的11/12诊断。离线接收保留未支持的枚举/诊断并在完整报告拒绝；当前运行时关闭未支持入站请求，不将其交付Inbound。T4 typed解码在第31批不等于完整支持；第38批已在T4DeliveryReport证据下启用IP结果，见[batch-31](history/batch-31.json)和[batch-38](history/batch-38.json)。

IMS RDR/RDA已新增第八种封闭Context `IMSDeliveryReport`，仅明确HSS Host/Realm、6..15位HSS-ID与≤256字节所选ASCII SIP URI；零Context为NeedsApplication，其他用途不能互换。T4/IMS共用有界内部身份存储，既有2KiB预留及旧证据上界测试通过；见[batch-32](history/batch-32.json)。

IMS ALR/ALA已新增第九种封闭Context `IMSHSSAlert`，仅HSS→IWMSC、15零dummy及目的URI；普通HSS角色、T4、IMS报告不能互换。共享内部身份存储不增加Context容量；MaximumUEAvailabilityTime继续保留原值/presence，不推断era/优先级或启动提醒。见[batch-33](history/batch-33.json)。

IMS SRR/SRA已新增第十种封闭Context `IMSRouting`，显式GMSC/IP-SM-GW、HSS-ID/目的URI、SC、已知HSS端点及可选UNRI；成功IMSI由HSS提供，无需请求方预知，不用HSS-ID等值/不等值猜订户分配。结构容量与2KiB预算不增，固定同HSS的两种恢复策略已补测，见[batch-34](history/batch-34.json)。

IMS OFR/OFA新增第十一种Context `IMSMOForwarding`，用途为IP-SM-GW→IWMSC，第35批起支持无IMSI/dummy源。两个URI由固定摘要绑定，原字节保留在请求，Context从672增至704字节而既有2KiB上界不变；OFR仍禁止多attempt恢复。开发期API：`sgd.UserIdentifier.UserName`改为Optional，普通调用用`avp.Some(imsi)`，见[batch-35](history/batch-35.json)。

第十二种Context `IMSMTForwarding`绑定SMS-GMSC→IP-SM-GW和顶层HSS-ID/两个URI/SC。TFR使用独立两叶SMSMI类型，不接受组内HSS-ID；Context仍704字节/≤2KiB。Timer/Start保留原值，不扩3330/3331/3332或自动转发，见[batch-36](history/batch-36.json)。

第37批扩展`IMSMOForwardingEvidence`的Optional SenderIMSI/SenderMSISDN，支持已知IMSI及可选真实MSISDN；两项缺省仍为dummy。Context种类和704字节/2KiB上界不变；实验5001在无原MSISDN时关联拒绝，有值时仍是IMS未选结果，见[batch-37](history/batch-37.json)。

第38批启用S6c T4无MSISDN报告的IP-SM-GW结果，复用T4DeliveryReport原IMSI/SC及GMSC/HSS绑定；六种合法节点组合、原因0/1/2和诊断11/12，零Context仍NeedsApplication。无新增Context/预算/重放权限，见[batch-38](history/batch-38.json)。

## 生成与操作契约

输入为 format=1 严格 JSON：拒绝未知/重复/大小写别名键、null、非法标识符、重复字段/命令、Vendor/flags 冲突、未解析 Grouped、循环/过深引用。当前生成器版本为13、操作契约ops8，Base Profile版本为`0.1.0-slice.8`，SGd为`0.1.0-slice.15`，S6c为`0.1.0-slice.22`，Sh为`0.1.0-slice.11`，Ro为`0.1.0-slice.6`、Ro abort与通用DCCA均为`0.1.0-slice.1`，smsfwro为`0.1.0-slice.6`、legacy420为`0.1.0-slice.6`、scap1为`0.1.0-slice.2`、combined为`0.1.0-slice.5`；保留Envelope及Canonical操作，第九批新增显式TransportReplay编译期许可，当前SRR（普通MSISDN、明确T4Routing的IMSI查询及IMSRouting的HSS-ID查询）和四Ro Profile已选CCR启用；CCR只允许同OCS活动exchange，见[batch-50](history/batch-50.json)。当前只支持本切片需要的 UTF8String、ASCII DiameterIdentity、Unsigned32/Enumerated、Unsigned64、Integer32/Integer64、具名Time、OctetString、数字 TBCD、Address、Grouped；跨文件引用/完整依赖锁与定义差异工具在后续 P2。ASCII identity 接受已转为 ASCII 的 DNS 名，不提供 Unicode/IDNA 转换。通用TBCD helper不处理TON/NPI；独立s6c.EncodeSMEA/DecodeSMEA仅处理国际十进制SMEA，见[batch-21](history/batch-21.json)。

生成输出直接访问强类型字段。`schema.Command` 是已实现的有限 wire 操作表：Decode、Encode（nil 输出时计量）、Validate、Clone、CompleteRequest、DecodeBound、Envelope、CompleteAnswer、正常答案可选CheckAnswer，有限具名CheckContext/ContextName；答案另可声明完整的ErrorOperations；请求可声明Canonical，缺失时不能启用缓存；应用请求可声明TransportReplay，缺失时拒绝多attempt Route；NoContext显式证明当前切片无额外调用方输入，NoContext与CheckContext互斥，两者均缺失则仅wire支持。缺失必需回调会使 Profile 构造失败。Profile/Bundle 显式组合且冻结，不使用 init 注册、反射、通用 AVP 树或运行时定义加载。外部扩展按同一公共操作契约生成并编译测试；完整 ExtensionSet 尚未实现。回调是受信任的编译代码，必须确定、线程安全、保持输入不变；这不是不可信插件沙箱。

生成 `DefinitionDigest` 覆盖规范化 JSON、显式规则修订号和生成器版本；Bundle 指纹再包含操作版本ops8、Profile 身份、命令元数据、E/Canonical、TransportReplay、NoContext、CheckAnswer存在性及ContextName，组合顺序不影响结果。它们是现有切片的定义/组合一致性标记，**尚不是部署契约中的完整 WireFingerprint/Canonical 指纹**。人工规则行为变化必须同步更新定义中的修订号/版本及回归证据，不能只改函数而沿用语义版本。

## 所有权与限制

- `codec.Decode` 先复制有界帧；`Read` 自己分配帧后移交内部 owned 入口。公共 API 无 Release/Adopt。输出可独立持有；调用方不得并发修改送入同一操作的 body。
- `CompleteRequest` 在校验/计量 body 及校验 envelope 后复制字节、字符串及分组，填补未指定 envelope；已有非空冲突、该命令不支持的 envelope 字段均拒绝。缺省 Optional 的残留 Value 被清零，不把未出现字段带入快照。
- 默认上限：帧 1 MiB、单 AVP payload 512 KiB、全树 4096 个 AVP、已知 Grouped 深度 16、每字段/每父实例 raw 扩展 256 个。配置必须整体有效；未知 raw 不递归。Grouped payload 限额在计量和写入时一致。
- DecodeBound、RequestBound、CompleteAnswerBound 是含帧、复制字符串、结构体及切片容量的保守逻辑预留估计，**不是 Go 堆/RSS 或运行时账本的证明**；Node 的 SingleAttempt admission/reservation/transfer 已在第三批接入；第六批补齐NoCache单Arrival I/W入站预算，第七批加入TransportReserve内的独立应急拒绝槽；第八批接通有限C及多Arrival重答，第九批为Route一次预留N份A，逐attempt保留关联/Peer窗口、writer/reader结清后转移完成额度；完整最小I/W/Q/A细分及堆证明未完成。
- 第十批新增每Peer固定4096字节管理预留，拨号先占主动候选槽，回调实际退出才释放；不计入业务出站额度。Dialer须遵守ctx与借用契约；标准库DNS/TLS及Go运行时内存不在逻辑额度证明内。
- 第十一批提供冻结SchedulerConfig和owned SchedulerInfo，按Control/Answer/Request再按class进行完整帧字节DRR；应急Answer独立lane。固定调度元数据计TransportReserve，权重不改变class字节/槽限额；作用域及无饥饿证据见[batch-11](history/batch-11.json)。
- 第十二批通过ListenerBytesPerPeer显式启用固定Peer的Serve；先占候选再Accept，TLS Establish与CER/CEA共用deadline，Node/Peer.Wait包含晚回调。取消Serve保留已建立连接，Shutdown负责排空，边界见[batch-12](history/batch-12.json)。
- 第二十批增加有限Sh特性/混合/5011、具名证据冻结和通知组合，生成器9/ops8、Sh slice.10，见[batch-20](history/batch-20.json)。
- 第十九批增加单Ref11/12/25订阅/退订、SendData presence和Ref11/12有限XML；Sh Profile升slice.9，其他版本不变，见[batch-19](history/batch-19.json)。
- 第十八批新增Ref24/25有限Sh-Data helper，精确父路径/namespace、字节/深度/节点/属性/文本/扩展预算及owned原文；不加载schema或承接订阅状态，见[batch-18](history/batch-18.json)。
- 第十七批新增Sh普通四组命令；User-Data只做AVP presence/字节预算并保留原文，正常Sh VSAI显式提供；PNR不依赖本地订阅状态，见[batch-17](history/batch-17.json)。
- 第十六批新增app/sgd与ops7 CheckAnswer，正常答案在共同关联后核对原请求应用关系，E独立；无新运行时状态/队列，见[batch-16](history/batch-16.json)。
- 第十五批增加普通MSISDN RDR/RDA及HSS ALR/ALA，具名avp.Time只保存四字节NTP seconds，不猜era；已有接收请求的关联使用Receive校验而不放宽本端发送。定义与独立离线证据见[batch-15](history/batch-15.json)。
- 第十四批提供codec.ValidateContext/ValidateAnswerContext、Inbound.Validation/Check及Result.Validation，显式NoContext审核纳入ops6指纹；当时仅零Context；第二十批增加有限ShPeerFeatures，Ro及完整逐叶报告待对应应用，见[batch-14](history/batch-14.json)。
- 第十三批提供IdentifierConfig/EndToEndSource、NewSessionID和IdentifierInfo；内置轮次无逐ID历史，自定义源须固定guard表及class硬配额，源忙即失败，控制/业务I/O实际退出后退休。HBH耗尽沿原DPR排空轮换；RO-06导入及跨进程保证不在本批，见[batch-13](history/batch-13.json)。
- Append 校验/计量后才扩展输出；错误保留调用前长度及原前缀，不能把返回错误的尾部当作完整帧。ReadFrame 在头部处拒绝超限，不读取下一帧；WriteFrame 单次 Write，保留 `n, err`，不自动发送短写剩余尾部。离线函数不提供网络发送语义或重试；第三批runtime单独管理BeginWrite与证据。
- 错误通过 `errors.Is/As` 分类；FieldError 记录父路径而不包含用户字段值。第七批仅为五个顶层基础必需字段生成5005占位；重复实例索引、任意坏请求的自动Failed-AVP和完整报告仍待实现。当前codec另支持诊断原字节、有界扫描及Incomplete标记。

第二十四批新增`codec.ValidateResponseContext`，离线发送完整校验用原请求Receive规则及答案Send规则；`CompleteAnswer/Append`只做静态线路/关联。SGd相等值接收保留并报告NeedsApplication，完整发送校验拒绝；条件、资源和事实冻结边界见[batch-24](history/batch-24.json)。

## 可重复验收

使用 Go 1.26.x，核心没有第三方依赖，因此无须空 `go.sum`。命令均在仓库根运行：

```sh
CGO_ENABLED=0 GOWORK=off GOTOOLCHAIN=local go generate ./...
CGO_ENABLED=0 GOWORK=off GOTOOLCHAIN=local go test ./...
CGO_ENABLED=0 GOWORK=off GOTOOLCHAIN=local go vet ./...
CGO_ENABLED=0 GOWORK=off GOTOOLCHAIN=local go build ./...
CGO_ENABLED=0 GOWORK=off GOTOOLCHAIN=local go run ./examples/codec
CGO_ENABLED=0 GOWORK=off GOTOOLCHAIN=local go test ./codec -run '^$' -fuzz '^FuzzDecodeAnswer$' -fuzztime=10s -parallel=2
CGO_ENABLED=0 GOWORK=off GOTOOLCHAIN=local go test ./codec -run '^$' -fuzz '^FuzzCanonical$' -fuzztime=10s -parallel=2
CGO_ENABLED=1 GOWORK=off GOTOOLCHAIN=local go test -race ./...
CGO_ENABLED=1 GOWORK=off GOTOOLCHAIN=local go test -race . -count=20 -timeout=300s
```

TCP/TLS集成测试绑定127.0.0.1临时端口、内存证书，不访问外部服务；需允许本地socket的执行环境，不能skip后记Passed。

race 使用 Go 的检测工具链，单列运行；发布构建仍为 CGO=0。受限环境可把 `GOCACHE` / `GOMODCACHE` 指向可写目录；本批因共享/tmp容量不足，将本任务GOCACHE迁至已ignore的`.cache/go-build`，GOTMPDIR使用`.cache/tmp`，未清理其他项目缓存。生成一致性、import DAG、外部生成消费者均包含在 `go test ./...`；检查而不改输出可运行 `go run ./cmd/diameter-schema -in spec/s6c.json -out app/s6c/zz_generated_messages.go -check`。

固定样本及 SHA-256 见 [testdata/codec/manifest.json](../testdata/codec/manifest.json)，独立栈的执行方法/差异见 [interop](../interop/README.md)。各批限定帧的独立离线证据不代表历史全部候选向量或全局 VECTOR-01 已通过。

## 下一批的有界顺序

当前优先级（第五十批后）：普通R33同OCS活动请求恢复已通过机制测试；目标启用仍需去重域/保存期证据。下一批进入RO-06原请求记录与标识准入，然后ResponseWait/串行attempt、错误后/重启续发及应用联验。标准实验5011配对、私有token/PSP、真实OCS/SCAP1和发布证据仍缺。

1. P2 后续：四应用余量的Canonical规则、完整报告/具名上下文、依赖锁及尚未承诺完成的字段/命令。已落地的公共包络和关联无需再做前置设计收敛。
2. P3已通过fake及真实TCP/TLS贯穿出站、关闭与选举。有限Inbound/应答及有限Reject/应急预算已在第六/七批实现；第八批补齐现有子集的Canonical、重复冲突、有限缓存和多Arrival重答；第九批完成固定同终端Route/多attempt与原身份重传；第十批完成有限Dialer/Maintain拨号接纳、取消/关闭、退避/抑制及重连；第十一批完成两层按class/字节DRR调度与资源公平性；第十二批完成固定Peer的有限Listener/Serve、两阶段握手及候选/关闭屏障；第十三批完成普通标识源、NewSessionID及耗尽/guard预算；第十四批完成现有Base/S6c有限Context/Report和分阶段入口；第十五批完成普通RDR/RDA及HSS ALR/ALA；第十六批完成SGd/Gdd普通MO/MT、透明TPDU/位置/结果及编译后的原请求目标检查；第十七批完成Sh四组普通命令及透明User-Data；第十八批完成Ref24/25有限Sh-Data helper与独立深层规则验收；第十九批完成Sh退订与Ref11/12单订阅/有限数据helper；第二十批完成有限Sh特性/混合/5011和终端证据；第二十一批完成普通HSS路由节点、国际十进制SMEA与SRR/SRA成对条件。第二十二批完成普通MSISDN RDR的MSC/SGSN父组及RDA验收；第二十三批完成MME/SGSN普通IMSI ALR/ALA、Event/新节点及显式角色证据。第二十四批完成SGd/Gdd直达3330/3332/3331、显式era和应用证据及发送/接收相等边界。第二十五批完成T4 OFA External-Identifier、M0发送与有界原身份/SC证据，TCP/TLS仅作中立运行时载体，T4 SCTP部署未验。第二十六批完成S6c T4无MSISDN的RDR/RDA、HSS ALR/ALA身份和用途绑定；RDR.UserIdentifier改为ReportUserIdentifier，旧MSISDN需包avp.Some，详见[batch-26](history/batch-26.json)。第二十七批完成T4无MSISDN SRR/SRA及原IMSI/节点关联，SRR.MSISDN改为Optional，见[batch-27](history/batch-27.json)。第二十八批完成上述SRA结果及5550三域诊断/角色UNRI条件，保留值接收并报告未支持，见[batch-28](history/batch-28.json)。第二十九批完成SGd OFA实验5001及接口/原MSISDN绑定，见[batch-29](history/batch-29.json)。第三十批完成S6c单IP-SM-GW节点及Realm/Additional接收边界，见[batch-30](history/batch-30.json)。第三十一批完成普通MSISDN RDR/RDA的IP-SM-GW结果父组及11/12诊断，见[batch-31](history/batch-31.json)。第三十二批完成IMS RDR/RDA的HSS-ID/目的URI及显式用途绑定，见[batch-32](history/batch-32.json)。第三十三批完成IMS ALR/ALA的dummy身份、目的URI与HSS提醒绑定，见[batch-33](history/batch-33.json)。第三十四批完成IMS SRR/SRA身份与角色绑定及固定HSS恢复复核，见[batch-34](history/batch-34.json)。第三十五批完成IMS OFR/OFA的dummy源身份、SMSMI三叶及IP-SM-GW→IWMSC用途，见[batch-35](history/batch-35.json)。第三十六批完成IMS TFR/TFA顶层HSS-ID、双URI、目标绑定和所选结果，见[batch-36](history/batch-36.json)。第三十七批完成IMS MO可用IMSI及可选真实MSISDN身份，见[batch-37](history/batch-37.json)。第三十八批完成S6c T4报告IP-SM-GW结果与诊断，见[batch-38](history/batch-38.json)。第三十九批已完成Ro基础：标准SMS IEC DIRECT_DEBITING的CCR/CCA类型、生成codec及成对结果；第四十批已完成无UPDATE的ECUR INITIAL/TERMINATION及原会话/序号关联；六模板/私有字段/RO-06仍在首版后续批次，不等待所有SMS/Sh扩展完成才开始Ro。
3. SCTP U1–U5 在上游独立推进，同候选验收后创建/接入独立适配 module；本批不增加 SCTP 依赖。
4. 原首版 Sh、完整 S6c、SGd/Gdd、Ro/兼容/RO-06 及发布门槛保持；Ro 未被重新后置。

第三十九批完成[标准SMS IEC基础](history/batch-39.json)：Ro slice.1、generator10/ops8；Base/S6c/SGd/Sh定义未改，生成摘要因版本变化更新。累计455份wire/17份XML，35份Ro旧库离线对照；4组TCP/TLS基础测试及全量回归详见progress。

第四十批完成[标准SMS ECUR](history/batch-40.json)：Ro slice.2，generator10/ops8及其他Profile不变；45份新wire（累计500wire/17XML），44份旧库离线对照，5组codec及3组TCP/TLS测试。当前Report仅校验本exchange，跨请求关系由纯helper显式检查，不保存业务Session。SMS-Node1为IP-SM-GW，已更正第39批的文字误称，原字节未改。

第四十一批完成[sms-fw IEC具名兼容](history/batch-41.json)：generator11/ops8，新增独立smsfwro slice.1，其他五定义未改。31份新wire（累计531wire/17XML）、29份旧库离线对照、4组codec/3组TCP-TLS及全量回归；所有42个旧模板节点有来源映射，但未执行旧C++或取得OCS抓包，不关闭完整TEMPLATE/COMPAT验收。

第四十二批完成[退款具名固定420基础](history/batch-42.json)：generator12/ops8、legacy420 slice.1，七输出更新/六既有定义不变。38新wire累计569wire/17XML，原旧库标准字典观察为不兼容，显式离线覆盖后33帧精确重编码；4组codec/3组TCP-TLS及全量回归。金额/单位由调用方明确提供，样例1/0/840/1不是库缺省；未关闭原扣费→真实OCS退款链。

第四十三批完成[旧ECUR两模板](history/batch-43.json)：smsfwro slice.2复用IEC固定类型，另六定义/输出及generator12/ops8不变。49新wire累计618wire/17XML，47原旧库字典离线往返，4组codec/3组TCP-TLS和全量回归；实际TERM的1210/1212顺序与固定sender不同，原帧保留且差异有专门验收。四模板构造基础不等于完整现场兼容。

第四十四批完成[通用CCR与费用/余额结果](history/batch-44.json)：legacy420 slice.2，另六定义/输出不变，generator12/ops8不变；55新wire累计673wire/17XML，47原字典独立对照、3组codec/4组TCP-TLS及全量回归。源码IPv4默认false的Address模式已证，现场目标配置及五模板组合仍未验。

第四十五批完成[SCAP1离线基础](history/batch-45.json)：新增第八份定义scap1 slice.1，原七定义/输出及generator12/ops8不变；61新wire累计734wire/17XML，56显式私有字典对照、4组codec/1组运行时拒绝边界测试。无SCAP1网络收发、自动广告或恢复许可。

第四十六批完成[SCAP1运行时](history/batch-46.json)：SCAP1Accounting显式出入站策略、头0/271与严格193/Acct19302协商、普通预算/pending和SingleAttempt Route，十组配置/TCP/TLS测试及全量回归。混合Peer Ready不等于SCAP1可用，重连重新核验，不自动重放。scap1 slice.2仅更新来源边界元数据，其他七定义/输出及generator12/ops8不变；无新wire/XML。旧双Auth/Acct广告明确拒绝，实际私有ACA与现场OCS未验。

第四十七批完成[五CCR固定组合](history/batch-47.json)：新增combined slice.1，原八定义/输出、generator12/ops8不变。2018统一为owned线路字节，由具名构造器保留STRING或固定Address载荷，模板规则禁止混用。四组codec/三组TCP-TLS、新36wire累计770wire/17XML，36独立字典往返。TERM及通用字段发送顺序差异显式记录，未取得OCS接受证据。

第四十八批完成[旧CCA分层结果/费用/DDFH](history/batch-48.json)：smsfwro/legacy420 slice.3、combined slice.2，其他六定义/输出及generator12/ops8不变。67新wire累计837wire/17XML，66原字典独立往返（坏宽度仅本库反例）；三组codec/两组TCP-TLS及全量回归。顶层与服务结果不覆盖，不把报告Complete当业务成功；私有token/PSP仍未猜测定义。

第四十九批完成[Ro结果安全](history/batch-49.json)：标准Ro slice.3、smsfwro/legacy420 slice.4、combined slice.3；generator13固定E规则/ops8，九输出刷新。新增65wire累计902wire/17XML、63原字典对照、三组codec/三组真实TCP-TLS与全量检查通过。实验结果/支持列表仅有界接收，2001+10415/5011报告ErrResultConflict；不生成未审定配对，下一批RO-06。

第五十批完成[Ro同OCS活动请求恢复](history/batch-50.json)：标准Ro slice.4、smsfwro/legacy420 slice.5、combined slice.4；generator13/ops8、其他五定义/输出不变。新增27帧累计929wire/17XML，27旧库对照（9原字典、18既有离线覆盖）；13组根测试和2组codec、全量纯Go/race通过。RO-06原记录/跨段/重启/Tr及目标恢复许可仍未完成。

## 第五十一批增量：Ro记录准备

`Node.PrepareRoRecoveryRequest`现已冻结四Ro Profile的已选CCR、目标绑定和恢复上限，并分配/保留原E2E；`RoRecoveryRequest.CheckIntegrity`只验证记录结构和摘要，`RoRecoveryInfo`报告有界配额。准备无连接/pending/发送/Completion，Node只保留摘要guard。默认/自定义源均保护该E2E，scratch、class槽和关闭借用屏障已验。见[本批契约](history/batch-51.json)。

下一批接`SubmitRoRecovery`的有限段接纳和累计额度高水位，再接ResponseWait/串行attempt/晚答与跨段/重启。记录的版本/Profile/绑定/原Origin/时间及冲突检查不能由`CheckIntegrity`代替；恢复导入和RecoverySpace尚未开放。无新增线路报文，九定义/输出、929wire/17XML保持不变。

## 第五十二批增量：Ro有限段接纳

`SubmitRoRecovery`现在接纳本Node准备的记录及原OCS有限计划，保留原E2E并分配新HBH；首段和已结清无答的续段复用Route/一次完成。SegmentNo及累计ReservedThrough在接纳时原子推进，未用allowance不退回；活动同段幂等、完成段号/终态防复活和预算/关闭屏障已验。`Result.RoRecovery`独立携带累计发送证据，普通Result.Delivery仍表示当前段。见[batch52](history/batch-52.json)。

下一步实现ResponseWait/Tr和同段旧新attempt竞态，之后才开放具名错误后续发、跨OCS和重启导入。第51批测试中combined INITIAL标签选择遗漏已在本批修正并增加线路类型断言；没有修改九定义/输出或原fixture。

## 第五十三批增量：Ro ResponseWait/Tr

`RoRecoveryPlan.ResponseWait`现支持有限、可禁用的排队计时；RequestInfo分别报告段/attempt接纳时间、Tr截止和ReplayReason。Tr不关闭健康连接，writer结清后才授予显式重传资格；旧pending/A保留，旧答可阻止尚未开始的后继，或等待已开始的writer/reader退休再交付唯一Completion。十二组TCP/TLS测试及完整检查见[batch53](history/batch-53.json)和[progress](progress.json)。

本批只扩展原OCS本地恢复段。下一批为具名错误答案后的本地有限续段，再接授权跨OCS与重启导入/RecoverySpace；应用保存、Tx/DDFH及OCS目标证据不由库接管。

## 第五十四批增量：具名错误续段

本地原OCS段增加显式ErrorContinuation和NextTrigger；默认所有答案终态。具名临时集合及带应用Tx截止的SMSFW扩展集合，仅依据校验通过的顶层基础结果授予有限ErrorAnswer续段。T依据和累计发送事实分开保留，正常/E答案照常交付，收发退休与关闭门不绕过。四Ro Profile补4002接收专用/不完整报告，版本和指纹更新，原929wire/17XML不变。十组TCP/TLS验收及下一批授权跨OCS的范围见[batch54](history/batch-54.json)。

## 第五十五批增量：授权跨OCS与同DRA恢复

Node增加仅用于Ro恢复的有限RoTargets配置；记录v2冻结AllowedPeers、目的/来源和去重/会话证据，Plan.Alternates固定有序不同目标。同DRA连接可为不同目标建立独立pending/A及新HBH，旧答案仍按原快照竞争一次完成；缺迁移依据拒绝，TERM仅开放具名兼容及本会话证据路径。十组测试含九组TCP/TLS、60条四Profile/15用途主流程，见[batch55](history/batch-55.json)。下一批原记录重启导入/RecoverySpace；无持久化、无自动业务重试边界保持。

## 第五十六批增量：原记录导入与RecoverySpaces

显式AllowRoRecoveryImport及Plan.Import接通四Profile/15用途跨Node恢复；Checkpoint保留计数、累计发送和T位谱系，重验原Origin/Bundle/目标与年龄。默认已用区间、自定义active/quarantine、控制域及有限RecoverySpaces参与接纳，失败暂占回滚，准备/收发退休共享Wait屏障。十一组测试与全量验收见[batch56](history/batch-56.json)。下一批限应用交接示例和验收；应用存储、隔离、进度真实性/时钟及OCS证据继续留在库外。

## 第五十七批增量：应用交接示例收尾

新增三个可执行Example、七组真实TCP/TLS测试，验证确切意图确认、未知发送保守额度、终态防复活、Tx后收答/Cancel与DDFH分支。所有应用策略仅在测试示例中，无生产API/协议定义变化；完整职责和退出清单见[交接说明](../doc/ro-recovery-handoff.md)，验收见[batch57](history/batch-57.json)。已选CCR库侧交接示例在此结束，实际应用/OCS证据另验。下一批转SCTP U1基础核验及U3生命周期，不新增存储provider或重放worker。

## 第五十八批增量：上游SCTP等待层

上游v0.7.4当前HEAD上补标准可更新deadline、内部context gate及hook退休，保留旧SO_*TIMEO语义。九组新增、全量257条测试/子测试零skip、race及纯Go检查通过；未修改上游未实施的v1设计。补丁/源码哈希和本批状态见[batch58](history/batch-58.json)。U3仅基础部分完成，下一批公开Context解析/拨号；U2/U4/U5和本库适配器仍未完成，本库依赖/定义不变。

## 第五十九批增量：上游Context解析/拨号

新增Go resolver的Context解析及使用同一ctx的有限拨号，保留原语法/配置所有权/errno；Control前置拒绝、失败关闭、最终交接与后取消隔离有真实SCTP和受控竞态证据。九组新增测试、外部包示例及全量检查见[batch59](history/batch-59.json)。第58批等待层与旧入口保持，下一批U2消息模式；U4/U5和本库适配器仍未完成。

## 第六十批增量：上游显式消息I/O

不可变Messages配置、ReadEvent分流/消费证据、WriteMsgContext单消息发送和最小选项读回已接通，沿用原context/deadline门。284条测试/子测试/示例零skip，见[batch60](history/batch-60.json)。U4 typed通知/实际长通知、完整连接预算、U5及本库适配器仍未关闭；旧批次说明按历史时点阅读。

## 第六十一批增量：五类通知及原生证据

ParseNotification固定字段owned、Tail借用，未知原值保真；全部五类通知在当前主机原生触发，真实长通知超限/重读及选项/accept失败路径已验。312条测试/子测试/fuzz种子/示例零skip、全量race及20轮定向race通过，见[batch61](history/batch-61.json)。本批关闭U4最小链路和对应U2原生证据，下一步U5完整预算/发布/平台出口；本库适配器仍未实施。

## 第六十二批增量：U5本地整合出口

资源账目分开固定值、逻辑槽位与Go/内核额外开销；有限reader/writer方式建议上游8KiB固定额度。新增ABI和100轮取消退休检查，同候选可重复打包并在无replace独立consumer中实际安装/收发。发布脚本及架构/consumer检查已接新API，见[batch62](history/batch-62.json)。远端矩阵与上游可引用发布未完成，不能据此关闭整个U5；下一步是精确候选review/CI/版本，再接本库适配器。

## 第六十三批增量：上游PR及远端候选验收

[PR #30](https://github.com/skill215/sctp/pull/30)的head/CI merge tree与第62批430文件候选一致，14项远端矩阵通过，真实伪版本下载/无replace消费者通过，见[batch63](history/batch-63.json)。可引用候选交付门已满足；draft PR未合并、无正式tag，首版发布状态继续单列。下一批固定此候选实现SCTP适配器最小链路。

## 第六十四批：SCTP适配与Peer接线

独立[adapter module](../adapters/sctp/README.md)固定第63批上游候选，完成有限帧/通知续调、16432默认逻辑额度与核心事件预留、Peer终止和CEA/DWR有序屏障；真实Base/S6c和取消/关闭测试通过，见[batch64](history/batch-64.json)。核心候选为离线开发版本，正式发布/托管拨号监听/现场验收继续待办。

## 第六十五批：SCTP托管建立与接纳

在同一上游候选上完成有限NewDialer/Listen、N/Z/D前置限制和建立期计费；复用Node.Dial/Maintain/Serve，真实Base/S6c、取消/关闭、失败退避、REOPEN无业务重放及双连接选举已验。固定20ms的上游poller等待片段支持Accept取消，无后台accept队列。契约/证据见[batch65](history/batch-65.json)。核心生产源码、生成定义及上游版本不变；正式发布和具名多流/PPID0兼容继续待办。

## 第65批之后：上游维护交接，未增加实施批次

维护者要求先增强skill215/sctp，让diameterkit直接调用。已形成[五项需求/三阶段验收](../doc/sctp-upstream-improvement-plan.md)和[可复制session提示词](../doc/sctp-upstream-session-prompt.md)，机器记录见[review](reviews/2026-11-09-sctp-upstream-handoff.json)。新工作均planned；第65批完成状态及历史测试保留，下一步改为先交付上游A–D，再迁移adapter；E随后删除连续性重复检查，具名多流策略留在本库。

## 第六十六批：A–E上游交接完成与迁移

[第66批](history/batch-66.json)固定bb739b612088真实伪版本，A–E接口/源码/日志/远端CI核对通过。删除adapter轮询、socket syscall、地址算法/上游硬编码账目和重复片段元数据检查。新额度由上游查询加本库值组成，终止通知映射补齐；核心2603、adapter71/71及20轮1400全部通过。先前交接计划已执行，不再作为待实施的上游任务；具名Diameter选流/PPID策略及发布/现场门禁仍单列。

第67批完成[SCTP正式v0.8.0锁定和回归](history/batch-67.json)，无Go源码变化。核心完整CGO0/race各2603、adapter各71、有限竞态20轮620均通过，证据见[evidence](evidence/batch-67-sctp-v080.json)；上游发布不再是阻塞，本库发布及现场互通仍单列。

第67批后[全局实施评估](../doc/diameterkit-implementation-review.md)发现DRA Relay-only协商及普通逐请求目的绑定仍缺，建议优先于具名SCTP模式；本次为扫描，未计为第68批、未新增SPEC编号。机器摘要见[review](reviews/2026-11-09-global-progress.json)。

第68批关闭全局扫描定位的Relay-only协商和有限逐请求终端绑定代码缺口，见[batch68](history/batch-68.json)；下一批优先Ro ASR/ASA，其他首版及现场验收保持未完成。

## 第六十九批：Ro ASR/ASA

[独立终止Profile](../app/ro/abort/rules.go)与四CCR Profile之一组合，复用公共运行时而不引入会话存储；正常2001/5002/5012由应用决定，E错误独立关联。53份独立正反报文、TCP/TLS重复/关闭/预算/TERM竞态及go-diameter八帧解释见[batch69](history/batch-69.json)。10份定义、982wire/17XML；generator13/ops8不变。RO-05库机制完成，应用TERM接线与目标OCS证据单列；其他标准操作、特性、SCAP1及发布缺口仍在进度中。

## 第七十批：通用DCCA单服务EVENT

[独立DCCA绑定](../app/ro/dcca/rules.go)提供查询/扣费/退款、精确金额、动作关联与条件结果报告；[退款关系检查](../app/ro/dcca/refund.go)只比较调用方给定原请求和新请求。新Profile不授予重传，标准SMS/2022与旧兼容域保持各自边界；详细范围、71帧和测试见[batch70](history/batch-70.json)。现11定义/1053wire/17XML，generator13/ops8不变，无新增运行时层或持久化。

### 第七十一批：标准SMS结果与2022叶级证据

标准Ro slice.6补4010/4011/5030顶层/单服务结果；[纯helper](../app/ro/refund.go)在已关联原IEC答案上区分Absent/Present-empty/Unknown，显式单服务/OCS绑定、长度上限与owned复制。它不替代Header/连接/来源校验、不授权完整SMS退款发送，也不判断金额或退款资格。现11定义/1099wire/17XML；46新帧和41独立解释、TCP/TLS及全量测试见[batch71](history/batch-71.json)。

### 第七十二批：SCTP具名模式

独立adapter新增OrderedMultiStream和PPID0Compatibility，实际出流选取、单流回退/ModeInfo、固定Session/hash映射、CEA/DWR屏障和预算/关闭已验。原上游v0.8.0、核心batch71及全部定义/语料不变；不增加业务排序、会话存储或SCTP通用能力副本。详细配置和测试见[batch72](history/batch-72.json)。

### 第七十三批：旧SGd提醒固定绑定

`compat/sgdalert`仅绑定16777313/8388648，复用S6c操作、角色和IMS/T4证据。正常/E、App关联、owned、重复冲突/多Arrival、预算/取消关闭及独立字典补充证据见[batch73](history/batch-73.json)。11生成定义不变，加1委托Profile，1125wire/17XML；新离线核心batch73、上游v0.8.0。Router token/MCCMNC和现场接受仍独立。

### 第七十四批：Router提醒身份

`s6c.RouterAlert`提供有界历史断言，ALR User-Name线路语法允许token，但MME/SGSN、T4/IMS与Router的用途校验分离。标准/旧App双绑定、23新帧、有限缓存/预算及TCP/TLS取消关闭见[batch74](history/batch-74.json)。S6c slice.21/旧委托指纹更新，1148wire/17XML；下一批为原SGd TFA及SRA Router/MCCMNC剩余用途。

### 第七十五批：SGd Router MT答案token

`sgd.RetransmissionToRouter`及显式Alert能力断言支持5550/3331/3102成对条件，专用`AlertUserIdentifier`保留token原字节；复用原era、关联和资源预算，无映射表或自动重发。SGd slice.15、30独立帧、1178wire/17XML及真实TCP/TLS验收见[batch75](history/batch-75.json)。下一批仅推进SRA Router/MCCMNC用途，随后模拟器输入边界；Ro/SCAP1及发布证据门保持。

### 第七十六批：SRA特殊用途

`RouterRouting`和`MCCMNCRouting`使用封闭Context，复用2KiB预算；成功token只在Router用途完整通过，MCC/MNC查询的身份前缀由应用提供，不查询归属或造dummy。原节点必需条件按原请求检查；S6c slice.22、27帧、1205wire/17XML、核心/adapter验收见[batch76](history/batch-76.json)。本批用途结束，P4剩模拟器输入有界决策。

### 第七十七批：模拟器输入决策

三个既有模拟器的活动构造/应答路径形成[有界迁移矩阵](../doc/diameterkit-simulator-migration.md)：坏标志/身份/状态/结果字段拒绝，修正路径、授权来源与每Arrival包络可执行验收。28新投影帧，1233wire/17XML；生产代码/11定义及离线corebatch76/SCTPv0.8.0不变。完整核心各3239、adapter各113及20轮760通过，见[batch77](history/batch-77.json)。输入决策关闭不等于SIM-01旧应用迁移或目标互通完成；下一批转既定SCAP1机制。

### 第七十八批：SCAP1旧广告与缓存

PeerConfig.SCAP1Advertisement显式开启源布局CER/CEA兼容，默认标准拒绝保持；原始预算、来源/关联及每代能力不放宽。ACR/ACA缓存多Arrival/冲突/取消关闭已验，C耗尽按原契约关闭。17新帧、1250wire/17XML，11定义/1委托Profile不变；核心离线batch78有126文件，SCTPv0.8.0不变，见[batch78](history/batch-78.json)。私有ACA/SCAP1恢复与目标证明仍单列，下一批进入P7双向接入示例/API核对。

### 第七十九批：SCAP1公开接入示例

[双向运行示例](../examples/scap1/README.md)使用完整公开配置和真实TCP，标准/旧模式×正常/失败/E应答及预算生命周期通过。独立消费module无replace、GOWORK=off并核对安装版公开API；见[batch79](history/batch-79.json)。无新线路语料/运行时行为，1250wire/17XML保持；公开注释和两个示例Go源纳入128文件离线corebatch79，SCTPv0.8.0不变。SCAP1示例出口结束，其余P7及原Ro/SCAP1证据门仍未齐。

第80批完成[Sh公开例及应用定义指南](history/batch-80.json)。核心CGO0/race各3338、适配各113、Sh例20轮race320及独立消费各16项通过；协议切片/1250wire/17XML不变。当前离线候选为batch80（130文件），不覆盖旧候选。下一批为S6c SRR/SRA公开例；其余P7/真实目标门保持，release_status仍not_ready。

第81批完成[S6c公开接入例](evidence/batch-81-s6c-example.json)，GMSC→HSS单向请求/反向答案，四种结果命名空间、身份/目的/Additional关联和资源生命周期已验。核心各3363、适配各113、示例20轮race500、独立消费各25项；现有协议定义不变。当前离线核心batch81有132文件。下一有界出口为SGd/Gdd公开例；完整P7及真实目标门保持。

第82批完成[SGd/Gdd公开接入](../examples/sgd/README.md)，两接口MO/MT方向、终端来源、基础/实验/E结果和独立资源生命周期已验。核心CGO0/race各3434、适配各113、示例20轮race1420、独立消费各71项通过；[验收证据](evidence/batch-82-sgd-example.json)。协议定义及1250wire/17XML不变，当前离线核心batch82共134文件。batch文档留在忽略目录，下一有界出口为标准Ro SMS IEC公开接入，完整P7/真实目标门保持。

第83批完成[标准Ro SMS IEC公开例](../examples/ro/README.md)，CCR/CCA五类结果、顶层与MSCC区分、精确单位及固定终端/预算/取消生命周期已验。核心CGO0/race各3467、适配各113、例20轮race660、独立消费各33项通过；[验收证据](evidence/batch-83-ro-example.json)。既有协议定义/1250wire/17XML不变，当前离线核心batch83共136文件。batch记录仍忽略，下一有界出口为无UPDATE ECUR公开流程，完整P7及真实目标门保持。

第84批完成[无UPDATE ECUR公开例](../examples/ro/README.md)：INITIAL与TERM的原Session/端点/服务关联、新exchange身份，显式应用续发决策与预算、取消/晚答/关闭均通过。核心CGO0/race各3502，适配各113，例68项、20轮race1360；独立消费各68项、26个API查询，见[证据](evidence/batch-84-ro-ecur-example.json)。11定义/1250wire/17XML不变，当前离线核心batch84共137文件，SCTPv0.8.0不变。batch记录继续忽略；下一有界出口为既有独立ASR/ASA公开流程，其他P7及真实目标门不变。

第85批完成[独立ASR/ASA公开例](../examples/ro/README.md)：OCS→CTF方向、原Session引用、应用显式选择2001/5002/5012/E、预算和取消/晚答/关闭已验，且ASA不自动完成同Session CCR或产生TERM。核心CGO0/race各3534，SCTP适配各113，累计Ro例100项、20轮race2000；独立消费各100项和33API查询，见[证据](evidence/batch-85-ro-abort-example.json)。11定义/1250wire/17XML及库行为不变，离线核心batch85有138文件，SCTPv0.8.0保持。batch Markdown仍忽略。下一有界出口为P7公开支持矩阵/API覆盖清单收尾，列明Ro剩余已选工作流消费证据；CI/fuzz/压力性能/发行及真实目标门仍独立。

第86批完成[公开支持矩阵](../doc/diameterkit-support-matrix.md)、[机器矩阵](release/support-matrix.json)和[API快照](release/public-api.json)：核心23包＋SCTP1包，Linux/Windows两构建选择。新[外部契约测试](../examples/contracts/README.md)43项，核心CGO0/race各3577、独立消费各43通过。生产Go/11定义/1250wire/17XML不变，复用138文件的batch85候选；适配器沿第85批各113项证据，本批未重跑。修正当前生成版本段落中SGd/S6c旧数字。

审查确认UX-02的S6c纯构造/结果解释与client/s6c仍缺，Ro恢复签名/Checkpoint不等于完整外部工作流。六组有界剩余项区别本库实现、消费验证、发布及外部证据，下一批优先S6c便利层；不能以示例或快照完成宣布全P7/v1结束。batch记录继续本地忽略。

第87批完成[纯S6c构造/解释及client/s6c](../doc/s6c-routing-helpers.md)，[验收](evidence/batch-87-s6c-helpers.json)含核心CGO0/race、公开消费、真实TCP Query/Submit/取消与资源清零。主/Additional/组合节点、三域诊断、Optional零值及原Result保留；固定Profile且无存储、候选优先级或投递策略。原UX-02实现缺口关闭，其他五组发布/消费/外部门保持；API快照为25包，未发布新候选。

第88批完成[Ro公开恢复与五操作消费](../examples/ro/README.md)，[验收](evidence/batch-88-ro-recovery-consumption.json)保留核心CGO0首次SGd就绪失败及复查证据。真实TCP DIRECT_DEBITING同OCS续段/导入/无答、Tx晚答和一次通知、五原业务操作及DDFH消费边界已验；无库API/定义/持久化变更。其他绑定/跨目标消费、私有/部署及发行门继续保持。

第89批完成[combined公开操作与双目标错误续段](../examples/ro/README.md)，[验收](evidence/batch-89-combined-cross-target.json)记录五请求形态×四结果、授权反例、逻辑入站预算修正及既有恢复测试与SGd双向示例的时序修正。核心API/定义不变；其他绑定/跨目标消费者、私有/目标及发布门继续保持。

第90批完成[独立绑定及无答跨目标/导入](../examples/ro/README.md)，[验收](evidence/batch-90-binding-recovery-consumption.json)记录三个具名绑定共176个矩阵场景、24项导入拒绝与独立module验证。无持久化；所选本地消费完成，私有/目标/发布门保持。

第91批完成[双module CI和生成/API/race本地验收](../doc/release-checks.md)，见[证据](evidence/batch-91-release-checks.json)。纯Go与真实SCTP零skip、临时checkout绑定、漂移反例及cross通过；没有托管run或发行安装声明。其他发布/外部门保持。

第92批完成[有界负载验收](../doc/load-checks.md)，[证据](evidence/batch-92-bounded-load.json)与[初始性能基线](release/performance-baseline.json)保留固定工作量和测量边界；纯Go/无持久化，未推送或执行托管CI，发行与外部门保持。

第93批[版本/安装政策](../doc/version-install-policy.md)与[门禁](tools/release_install.py)已实施并测试，真实冷缓存安装仍因未发布版本404保持partial；不得用本地成功替代发行证据。
