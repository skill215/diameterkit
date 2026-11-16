# Diameter 主流程字段、结果分支与依赖审计

> 实施注（2026-11-07，第四十八批）：已证旧CCA顶层/单MSCC结果分别保留，新增4010/4011/5030、IEC/ECUR费用及通用直接扣费DDFH；见[batch48](../spec/history/batch-48.json)。67新wire累计837，66原字典对照；token接收Vendor/父路径与PSP仍缺证，不猜字段或默认成功。下一批5031/5011结果安全，随后RO-06；最新[progress](../spec/progress.json)。

> 实施注（2026-11-07，第四十七批）：五CCR模板的独立combined Profile已接通；2018固定线路字节及模板约束、原请求答案分支隔离、六用法TCP/TLS见[batch47](../spec/history/batch-47.json)。原八Profile及样本不变；新增36wire累计770。固定发送顺序和2018混合表示的实际OCS接受、私有答案/5011/RO-06仍未验收。最新状态见[progress](../spec/progress.json)。

> 实施注（2026-11-07，第四十六批）：SCAP1显式SCAP1Accounting绑定、严格193/Acct19302能力及App0/271普通预算/pending路径已接通，真实TCP/TLS验证SingleAttempt、双向能力、正常/E、取消晚答、关闭和重连。旧双Auth/Acct广告仍拒绝，实际私有ACA与现场OCS未验；不自动重放。见[batch46](../spec/history/batch-46.json)和[progress](../spec/progress.json)。下方旧批记录保留历史时点。

> 实施注（2026-11-07，第四十五批）：SCAP1实际47节点/11保序参数的离线构造、codec与合成基础ACA配对已完成；见[batch45](../spec/history/batch-45.json)。App0/271运行时和193/19302能力/旧广告绑定、真实私有ACA仍未启用；六源模板构造基础不关闭完整COMPAT/RO-06。最新状态见[progress](../spec/progress.json)。

> 实施注（2026-11-07，第四十四批）：通用32节点、DIRECT_DEBITING/CHECK_BALANCE与顶层Cost/GSU/余额结果已在legacy420 slice.2完成；见[batch44](../spec/history/batch-44.json)。五CCR源基础分布于两个Profile，2018 Address/STRING差异仍须固定组合验收；SCAP1/未知私有答案/RO-06及现场证据未关闭。最新状态见[progress](../spec/progress.json)。

> 实施注（2026-11-07，第四十三批）：旧INITIAL41/TERMINATION44节点的构造/codec/有限TCP-TLS已完成，与IEC共用smsfwro固定Profile；见[batch-43](../spec/history/batch-43.json)。TERM两MMS子叶的原顺序与sender差异明确保留，未取得目标接受证据；四模板基础不关闭全部兼容/CCA依赖/RO-06。最新状态以[progress](../spec/progress.json)为准。

> 实施注（2026-11-07，第四十二批）：旧退款26模板节点及CCR/456/437/420固定Grouped、signed金额与顶层0/1292已完成构造/codec/有限TCP-TLS，见[batch-42](../spec/history/batch-42.json)。原扣费关联、未知CCA token/PSP、目标接受及恢复仍未关闭；早期Planned记录保留历史，当前状态以[progress](../spec/progress.json)为准。

> 实施注（2026-11-07，第四十一批）：sms-fw IEC源模板42节点已有独立具名Go构造/codec和TCP/TLS验收，见[batch-41](../spec/history/batch-41.json)。31份新wire是独立源投影/合成结果，非C++执行输出或OCS抓包；旧模板中没有420，退款Grouped420留下一批。本文早期Planned记录保留历史含义，当前Partial进度以[progress](../spec/progress.json)为准，六模板/私有答案/RO-06整体未关闭。

> 实施注（2026-11-07，第四十批）：V-SMS/V-TERM各三份完整报文已原样物化并完成无UPDATE ECUR I/T基础验收，含原INITIAL的IMEISV16/PS字段及纯跨请求校验。SMS-Node1正确名称为IP-SM-GW，第39批说明中的SMS-SC误称已更正，原字节不变。范围及未实现边界见[batch-40](../spec/history/batch-40.json)，不关闭SMS失败/在线序号、六模板/私有绑定或RO-06。

> 实施注（2026-11-07，第三十九批）：V-IEC-CCR-01/CCA-01/CCA-02已原样物化，完成标准SMS-SC执行前IEC DIRECT_DEBITING生成codec及成对验收，另有38份独立投影。支持范围、uint64/退款凭据、结果/关联和未选边界见[batch-39](../spec/history/batch-39.json)。退款候选、ECUR、六模板、私有绑定、5011及RO-06不因本注转为已实现；历史分析结论保留。

> **当前范围（2026-11-04，Draft 0.20）：** 第一阶段=首版/v1，包含Sh、S6c、SGd/Gdd、Ro/DCCA及sms-fw实际所需六模板、SCAP1和RO-06恢复能力；Go 1.26、纯Go实现（含依赖与SCTP适配器，发布路径CGO_ENABLED=0）、公开Go库、无持久化边界不变。明确撤销第十九轮“Ro后置”分期；带日期的旧分期/统计仅作历史证据，不再授予Ro跳过首版验收的例外。限定S6a、完整SCUR/PRICE_ENQUIRY等未承诺能力仍后置。范围及剩余缺口见[Specification §3/§7](diameterkit-specification.md)，顺序见[总设计§14](diameterkit-design.md)，验收见[覆盖§39](diameterkit-usage-coverage.md)。2026-11-06已完成前十九批有界实施（Base/S6c基础、公共包络与关联、出站运行时、关闭与选举、TCP/TLS生产适配、有限Inbound/应答、有限Reject/应急预算、Canonical/有限重复缓存及多Arrival重答、固定同终端Route/多attempt与活动请求原身份重传、有限Dialer/Maintain拨号及重连、按class/编码字节DRR调度、固定Peer的有限Listener/Serve、有界标识源/Session/guard、Base/S6c有限Context/Report、S6c普通RDR/RDA及HSS ALR/ALA、SGd/Gdd普通OFR/OFA与TFR/TFA、Sh普通四组命令及透明User-Data、Ref24/25有限Sh-Data helper、Sh单订阅/退订及Ref11/12有限数据）；已执行范围及未完成项以[Specification §6](diameterkit-specification.md#6-需求与证据的维护)和[spec进度](../spec/progress.json)为准，未列明的运行验收仍Planned。

日期：2026-10-31（第三十一轮协议切片修订），第十二轮建立，第十三轮补§9，第十四轮补§10，第十五轮补§11；第十六轮§12核对CR/后续版本；第十七轮§13补29.061、实际编码依赖及标准主流程pair；第十八轮§14补ECUR/IEC、RDR/ALR和六模板源码证据，针对 R29/R34。定位仍是 Go 1.26 的公开、无持久化端点库。本页补充 [Profile 规则契约](diameterkit-profile-rules.md)，接受 [库级 specification](diameterkit-specification.md) 的需求与验收索引，不增加业务会话、计费状态机或配置服务。

截至第三十一轮，主命令顶层字段、关键 Grouped/叶字段及结果条件已盘点；§16新增Sh四组完整pair、Sh-Data及MME/SGSN提醒、Gdd角色条件。**字段被列出、定义被核对、条件已收敛、独立 fixture 通过是不同状态。** 本页不声称全套 Rel-18 AVP 或所有可选功能已经审核完毕；历史缺口见§8；§17补号码/订阅/特殊身份；§18收敛SRA无IMSI、Realm具名兼容与Sh混合订阅，第二十三轮逐叶/结果/深层边界见§19及独立清单；§20补独立完整帧及跨组/时间条件，§21补Sh特性/T4/逐叶反例及勘误，§22展开命令/父路径矩阵并复核P1，§23补实际应用路径反例、Note9及四项P1边界，§24补九答案族正常/E包络与关联，§25补公共父实例/未知扩展/诊断边界和CR0069线索，§26补Origin-State/Route-Record/结果父组/错误文本和P1交叉依据，§27补公共父叶/请求必需项/目的绑定并修正ALR过程条件。首版前置分析止于§27；2026-11-03收尾后的任务归属见Specification §7。§28为2026-11-04用户授权的首版Ro有界收敛；§29继续审查结果、ASR和错误包络；§30补服务集合、凭据证据、特性生命周期及ASR终止报告职责；§31集中裁决三项受限能力与兼容证据门槛，不重开首版P0、不开始实施。

首版Ro的实施准入、五包与验收机制见[覆盖§36](diameterkit-usage-coverage.md)；用户进一步明确以sms-fw C/C++实际用法全覆盖为交付终点，必需用法、顺序及完成判定见该文§37和本页§32。最小标准子集只是中间交付，不能替代此终点。本页§28–31保留协议依据与具体限制，不新增实施代码。

> 实施注（2026-11-07，第三十批）：§18.3中的标准无Realm组合及§12.4的有界Realm接收/未支持报告已落地；Additional与ALR不继承Serving组合。具名s6c-ip-sm-gw-realm-v1仍未启用，CR0069未取得新的正文/批准证据。旧go-diameter的3112为OctetString、3348为Unknown，仅作字节往返对照，见[batch-30](../spec/history/batch-30.json)。本注不覆盖各历史轮次的当时状态。

> 实施注（2026-11-07）：第三十一批落实普通MSISDN RDR的IP-SM-GW结果父组、cause0/1/2与11/12诊断；诊断13是未选分支，不记为规范禁止。可解析未支持请求仍按现有入站策略关闭，不能套用答案交付报告契约；MWD/SC历史及提醒属于应用。旧IMS/T4候选与规范未决项未因typed组新增而整体通过。见[batch-31](../spec/history/batch-31.json)。本注不改历史轮次当时状态。

> 实施注（2026-11-07）：第三十二批落地IMS RDR/RDA的3102/1 HSS-ID、3324/3327目的URI、禁止3325/3326及明确Host/Realm终端绑定。组条件来自当前消息；新增具名IMSDeliveryReport只断言本次用途/身份，不是实时facts或数据库接口。URI选ASCII≤256字节，不宣布全SIP/国际化支持，旧3325字典类型差异保留。见[batch-32](../spec/history/batch-32.json)。历史轮次当时状态保留。

> 实施注（2026-11-07）：第三十三批落实IMS HSS ALR/ALA：15零dummy MSISDN、目的SMSMI URI及具名IMSHSSAlert；HSS-ID/源URI和接入Event/节点禁止。可选Time只保留原值，提醒重试/MWD清理及SC转发仍属应用；不以ALA2001宣称已投递。见[batch-33](../spec/history/batch-33.json)。历史轮次当时状态保留。

> 实施注（2026-11-07）：第三十四批落实IMS SRR/SRA的HSS-ID/目的URI及显式角色/终端/UNRI绑定，保留真实IMSI成功语义；T4原IMSI关联与IMS HSS-ID不同，不能凭字节判断订户分配。两种固定HSS活动请求恢复已验证，不扩业务重试或真正无IMSI成功。见[batch-34](../spec/history/batch-34.json)。历史轮次当时状态保留。

> 实施注（2026-11-07）：第三十五批落实IMS OFR/OFA无IMSI dummy源、SMSMI三叶和临时IP-SM-GW结果；IMSMOForwarding显式绑定IP-SM-GW/IWMSC和原身份，普通5001/T4计费证据不能借dummy授权。原投递历史与实际SC转发仍属应用。见[batch-35](../spec/history/batch-35.json)。历史轮次当时状态保留。

> 实施注（2026-11-07）：第三十六批落实IMS TFR/TFA顶层HSS-ID、双URI及GMSC/IP-SM-GW证据；Timer/Start保留原值，MME/SGSN重投时间和未选失败不自动继承。见[batch-36](../spec/history/batch-36.json)。历史轮次当时状态保留。

> 实施注（2026-11-07）：第三十七批落实IMS MO的可用IMSI/可选真实MSISDN及原值/presence绑定，保留无IMSI dummy。5001缺原MSISDN关联拒绝，有号码仍不授权IMS5001用途；无自动转发。见[batch-37](../spec/history/batch-37.json)。历史轮次当时状态保留。

> 实施注（2026-11-07）：第三十八批落实T4无MSISDN报告的IP-SM-GW结果及合法节点组合，仍以T4DeliveryReport绑定原IMSI/SC与GMSC/HSS，投递节点不等于请求方角色。见[batch-38](../spec/history/batch-38.json)。历史轮次当时状态保留。

## 1. 采用依据与真实文件证据

取证日期为 2026-10-30。下表哈希针对实际下载字节；PDF、ZIP、解压成员分别计算。文件用于本次只读分析，未把整本规范或依赖源码复制进仓库。哈希固定取证对象，不表示规范符合性或对端互操作通过。

| 原件与固定版本 | SHA-256 | 本轮用途/状态 |
| --- | --- | --- |
| [TS 29.329 V18.0.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129329/18.00.00_60/ts_129329v180000p.pdf) | `f57118a4ba21ac86fd129cbbe8010038d82a246731ddb008e44f9fc9579405ac` | Sh 命令、AVP 与结果 |
| [TS 29.328 V18.1.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129328/18.01.00_60/ts_129328v180100p.pdf) | `485d8f22da249e7bbf41277c8ed70e0fbf9c31fcac7ed5195715edb03c7033f5` | Sh 程序、数据与附件引用 |
| [TS 29.338 V18.2.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf) | `b22bc1341f1432be91838a84f2f163d6b5e42f27ca19d3bdbaf0995a9b09096d` | S6c/SGd/Gdd |
| [TS 29.229 V18.0.0](https://www.etsi.org/deliver/etsi_ts/129200_129299/129229/18.00.00_60/ts_129229v180000p.pdf) | `3a3d6581afb7ab3533e47b2ad9cedf7ebfc1c74226f2bca69c10a8f690ee5bbf` | 公共特性/结果 |
| [TS 29.272 V18.3.0](https://www.etsi.org/deliver/etsi_ts/129200_129299/129272/18.03.00_60/ts_129272v180300p.pdf) | `6a96a93926decce8f31ff7a7b4f7254316e2bf76c53d3fb5a37996816586d64b` | EPS 位置及复用字段 |
| [TS 32.299 V18.0.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132299/18.00.00_60/ts_132299v180000p.pdf) | `584c47bf152e7ba150a2047559a65d2f7352c07b1c719ef812836d6c3f13b3e9` | Ro 命令、Grouped、类型与应用覆盖 |
| [TS 32.274 V18.0.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132274/18.00.00_60/ts_132274v180000p.pdf) | `a7bac496be7d624f8fb1fb99f2383f3591c5a635c3afbce8ee3b4262f3e716a7` | SMS charging §6.2/6.3；其余范围不自动纳入 |
| [TS 32.240 V18.6.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132240/18.06.00_60/ts_132240v180600p.pdf) | `e738759daeb5cd6e971e3302d9a468ad681cee744eaceb1d6a8c4c1af2163931` | §5.4 category |
| [TS 29.336 V18.2.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129336/18.02.00_60/ts_129336v180200p.pdf) | `fdbd864b79deb3c7b8a6d023489633ec0d97141645e7745e889aa57185910e61` | §6.4.2/6.4.11身份；第二十二轮追加§6.4.12/14/15/18节点组合与Realm |
| [TS 29.173 V18.0.0](https://www.etsi.org/deliver/etsi_ts/129100_129199/129173/18.00.00_60/ts_129173v180000p.pdf) | `90520af30dfc8a4ccd8e26d6dd5c70edd5f2a5bcee8d3731d4212876b55d7dfc` | 节点、LMSI；应用覆盖优先 |
| [TS 32.270 V18.2.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132270/18.02.00_60/ts_132270v180200p.pdf) | `b01dcfd2a3a6bb902364055349439e22f4195175223df601f21e773c94cb1126` | MMS 子模型引用已定位，非全本审核 |
| [TS 29.002 V18.0.0](https://www.etsi.org/deliver/etsi_ts/129000_129099/129002/18.00.00_60/ts_129002v180000p.pdf) | `79eeee45fb74c532fe13d2653914a5f9ca29ea775634d859534b31420b56dd06` | 地址/编码引用，非 MAP 实现 |
| [TS 23.040 V18.0.0](https://www.etsi.org/deliver/etsi_ts/123000_123099/123040/18.00.00_60/ts_123040v180000p.pdf) | `3d624f5fa179e40cdfa96c0b9124f7115bb6b28310574168a3d0a953287e96fe` | 短信字段引用，非 TPDU 引擎 |
| [29.328 同版本 ZIP](https://www.etsi.org/deliver/etsi_ts/129300_129399/129328/18.01.00_60/ts_129328v180100p0.zip) | `d52149a944f455c16ecb123d3ebfb022f680330216838bc96a4c6ba3287e04d5` | 取得官方附件 |
| [ZIP 内 ShDataType_Rel18.xsd](https://www.etsi.org/deliver/etsi_ts/129300_129399/129328/18.01.00_60/ts_129328v180100p0.zip) | `8e01bf2537df9460ede73dbcb39a275243681f7c1cda8d19061c924824371e79` | 55518 字节；无外部 schema 依赖 |

本轮新增固定依赖：29.336 V18.2.0 的 User-Identifier/External-Identifier，29.173 V18.0.0 的节点字段，32.240 V18.6.0 的字段 category，32.270 V18.2.0 的 MMS 子模型。版本来自官方目录与原件封面；不是把不存在的 V18.0.0 URL 当成已选版本。只有实际引用的字段进入本库，不由此声明 S6m/S6t、SLh、MMS、CDR 或 charging 系统支持。

Sh 的 DRMP、OC、Load 分别补入 [RFC 7944](https://www.rfc-editor.org/rfc/rfc7944.html)、[RFC 7683](https://www.rfc-editor.org/rfc/rfc7683.html)、[RFC 8583](https://www.rfc-editor.org/rfc/rfc8583.html) 依赖。定义其有界字段表示，不等于自动启用路由优先级、DOIC 限流或负载选路。未实现算法不得广告支持或生成暗示算法已执行的报告；收到的相关信息也不能直接改变 DRA/OCS 路由策略。

### 1.1 规范冲突与勘误裁决

| 记录 | 证据与本库决定 | 验收/未决状态 |
| --- | --- | --- |
| SPEC-SMSMI-CODE | 29.338 表 6.3.3.1/1 指定 SMSMI-Correlation-ID 为 3324，§6.3.3.13 的示例却写 3308；3308 已用于 SM-RP-MTI。本库固定采用表中 `(10415,3324)`，不按负载猜两种类型 | 这是有记录的本地裁决，不冒称官方勘误；须用独立报文和另一个栈交叉验收 |
| SPEC-SH-XSD | 29.328 Annex C 的正文仍提及 Rel15 文件名；同版本官方 ZIP 实际含 `ShDataType_Rel18.xsd` | 采用同版本附件及其哈希，不从其他版本补一个同名文件；下述结构已检查；第二十轮XML离线schema结果见§16，库fixture未实施 |
| SPEC-RO-5011 | 32.299 §6.4.3 要求正常 CCA 有 Result-Code，§6.5.10/§7.1.8A 又指定特性不足时使用实验结果 5011，却未在这些条款给出伴随基础结果的唯一取值 | 保留两个结果和冲突诊断；绝不因基础 2001 抹掉实验失败。自动构造该分支的取值仍为 P1，不能自造一个“标准配对” |
| ERR-6733-4615 | [Verified 4615](https://www.rfc-editor.org/errata/eid4615)：5001 的诊断使用一个 Failed-AVP 容器 | 采用；容器内可有多个有界错误项，嵌套路径和最小占位值受检 |
| ERR-6733-4887 | [Verified 4887](https://www.rfc-editor.org/errata/eid4887)：应答填本地 Origin-Realm | 采用；与 Origin-Host 一起由 endpoint 注入，不复制请求端身份 |
| ERR-6733-3805/3997/4808 | [6733 勘误表](https://errata.rfc-editor.org/search/?rfc_number=6733)：长度单位、保护传输端口、取消“完全向后兼容”措辞 | 长度按字节；保护端口 5868；不承诺 3588 的重复 Failed-AVP/Vendor-Id 与本 Profile 通用兼容 |
| ERR-4006-3329 | [Held for Document Update 3329](https://www.rfc-editor.org/errata/eid3329)：未出现 GSU 与出现零值 GSU 不等价 | 作为必测反例，保留 Presence；不把 Held 状态写成 Verified，不引入自动终止/充值流程 |

其余 RFC 更新、勘误和 3GPP CR 尚未逐项裁决；已看勘误索引不等于全部采用。特别是 RFC 7075/8553、完整 8506 差异与非主流程依赖仍在清单中。

## 2. 顶层字段全集与出现次数

下列字段清单以所选版本的命令语法为盘点依据，**还要叠加 §3–5 的应用条件**。`!` 为 1 次，`?` 为 0..1，`+` 为 1..N，`*` 为 0..N；N 受 MaxRepeated/MaxAVPs/MaxFrameBytes 限制。`E` 是仅用于本表展开的公共项，不是公开接口或运行时继承框架：

- Sh/SMS 的 E：Session-Id!（首 AVP）、Auth-Session-State!、Origin-Host!、Origin-Realm!、DRMP?、Supported-Features*、Proxy-Info*、Route-Record*、扩展 AVP*。
- Sh 再加 Vendor-Specific-Application-Id!；SMS 的该 AVP 虽出现在 ABNF 可选项中，实际发送禁带、接收忽略，沿用 SMS-VSAI。
- Sh/SMS 请求加 Destination-Realm!；除 PNR/TFR 为 Destination-Host! 外，其余请求为 Destination-Host?。所有正常答案加 Result-Code?、Experimental-Result?、Failed-AVP?；结果存在性和互斥按应用条件检查。
- 未标出的 AVP 不因为处于 `扩展 AVP*` 就免于类型、M 位、路径、数量和资源检查；反过来，也不能把扩展槽误作“只准白名单”的理由。

| 命令 / 来源 | E 与目标/结果规则之外的全部具名顶层字段 |
| --- | --- |
| UDR / 29.329 §6.1 | User-Identity!、Wildcarded-Public-Identity?、Wildcarded-IMPU?、Server-Name?、Service-Indication*、Data-Reference+、Identity-Set*、Requested-Domain?、Current-Location?、DSAI-Tag*、Session-Priority?、User-Name?、Requested-Nodes?、Serving-Node-Indication?、Pre-paging-Supported?、Local-Time-Zone-Indication?、UDR-Flags?、Call-Reference-Info?、OC-Supported-Features? |
| UDA / 29.329 §6.1 | Wildcarded-Public-Identity?、Wildcarded-IMPU?、User-Data?、OC-Supported-Features?、OC-OLR?、Load* |
| PUR / 29.329 §6.1 | User-Identity!、Wildcarded-Public-Identity?、Wildcarded-IMPU?、User-Name?、Data-Reference+、User-Data!、OC-Supported-Features? |
| PUA / 29.329 §6.1 | Wildcarded-Public-Identity?、Wildcarded-IMPU?、Repository-Data-ID?、Data-Reference?、OC-Supported-Features?、OC-OLR?、Load* |
| SNR / 29.329 §6.1 | User-Identity!、Wildcarded-Public-Identity?、Wildcarded-IMPU?、Service-Indication*、Send-Data-Indication?、Server-Name?、Subs-Req-Type!、Data-Reference+、Identity-Set*、Expiry-Time?、DSAI-Tag*、One-Time-Notification?、User-Name?、OC-Supported-Features? |
| SNA / 29.329 §6.1 | Wildcarded-Public-Identity?、Wildcarded-IMPU?、User-Data?、Expiry-Time?、OC-Supported-Features?、OC-OLR?、Load* |
| PNR / 29.329 §6.1 | User-Identity!、Wildcarded-Public-Identity?、Wildcarded-IMPU?、User-Name?、User-Data! |
| PNA / 29.329 §6.1 | 无额外具名字段 |
| SRR / 29.338 §5.3.2 | MSISDN?、User-Name?、SMSMI-Correlation-ID?、SC-Address?、SM-RP-MTI?、SM-RP-SMEA?、SRR-Flags?、SM-Delivery-Not-Intended? |
| SRA / 29.338 §5.3.2 | User-Name?、Serving-Node?、Additional-Serving-Node?、LMSI?、User-Identifier?、MWD-Status?、MME-Absent-User-Diagnostic-SM?、MSC-Absent-User-Diagnostic-SM?、SGSN-Absent-User-Diagnostic-SM?、SMSF-3GPP-Address?、SMSF-Non-3GPP-Address?、SMSF-3GPP-Absent-User-Diagnostic-SM?、SMSF-Non-3GPP-Absent-User-Diagnostic-SM? |
| ALR / 29.338 §5.3.2 | SC-Address!、User-Identifier!、SMSMI-Correlation-ID?、Maximum-UE-Availability-Time?、SMS-GMSC-Alert-Event?、Serving-Node? |
| ALA / 29.338 §5.3.2 | 无额外具名字段 |
| RDR / 29.338 §5.3.2 | User-Identifier!、SMSMI-Correlation-ID?、SC-Address!、SM-Delivery-Outcome!、RDR-Flags? |
| RDA / 29.338 §5.3.2 | User-Identifier? |
| OFR / 29.338 §6.3.2 | SC-Address!、OFR-Flags?、User-Identifier!、EPS-Location-Information?、SM-RP-UI!、SMSMI-Correlation-ID?、SM-Delivery-Outcome? |
| OFA / 29.338 §6.3.2 | SM-Delivery-Failure-Cause?、SM-RP-UI?、External-Identifier? |
| TFR / 29.338 §6.3.2 | User-Name!、SMSMI-Correlation-ID?、SC-Address!、SM-RP-UI!、MME-Number-for-MT-SMS?、SGSN-Number?、TFR-Flags?、SM-Delivery-Timer?、SM-Delivery-Start-Time?、Maximum-Retransmission-Time?、SMS-GMSC-Address? |
| TFA / 29.338 §6.3.2 | Absent-User-Diagnostic-SM?、SM-Delivery-Failure-Cause?、SM-RP-UI?、Requested-Retransmission-Time?、User-Identifier?、EPS-Location-Information? |

旧字典与固定版本的差异已具体定位：旧 Sh UDA/PUA/SNA 缺 Load；旧 SGd TFA 缺 EPS-Location-Information；旧 TFR 规则未列扩展 AVP 槽；旧 ALR 使用 Serving-Node-T4 名称，而所选命令语法使用 Serving-Node，必须核对实际 Code 和组定义，不按别名直接继承。这些发现来自字典与原件逐项比较，不表示所有旧版本报文都错误。

### 2.1 Ro 顶层全集及应用覆盖

Ro 不复用上面的 E。CCR/CCA 均固定 Session-Id 首位；`Auth-Application-Id=4`，原请求/答案的 CC 类型与编号相符。语法依据 32.299 §6.4.2–3，以下是字段盘点，发送集合须再按其表格及 32.274 覆盖收紧。

| 命令 | 必需字段（各 1） | 0..1 字段 | 0..N 字段 |
| --- | --- | --- | --- |
| CCR | Session-Id、Origin-Host、Origin-Realm、Destination-Realm、Auth-Application-Id、Service-Context-Id、CC-Request-Type、CC-Request-Number | Destination-Host、User-Name、CC-Sub-Session-Id、Acct-Multi-Session-Id、Origin-State-Id、Event-Timestamp、Service-Identifier、Termination-Cause、Requested-Service-Unit、Requested-Action、AoC-Request-Type、Multiple-Services-Indicator、CC-Correlation-Id、User-Equipment-Info、OC-Supported-Features、Service-Information | Subscription-Id、Used-Service-Unit、Multiple-Services-Credit-Control、Service-Parameter-Info、Proxy-Info、Route-Record、扩展 AVP |
| CCA | Session-Id、Result-Code、Origin-Host、Origin-Realm、Auth-Application-Id、CC-Request-Type、CC-Request-Number | Experimental-Result、User-Name、CC-Session-Failover、CC-Sub-Session-Id、Acct-Multi-Session-Id、Origin-State-Id、Event-Timestamp、Granted-Service-Unit、Cost-Information、Low-Balance-Indication、Remaining-Balance、Final-Unit-Indication、Check-Balance-Result、Credit-Control-Failure-Handling、Direct-Debiting-Failure-Handling、Validity-Time、OC-Supported-Features、OC-OLR、Redirect-Host-Usage、Redirect-Max-Cache-Time、Failed-AVP、Service-Information | Multiple-Services-Credit-Control、Redirect-Host、Proxy-Info、Route-Record、扩展 AVP |

Supported-Features 虽未单列在上述 CCR/CCA ABNF 中，仍受 §6.5.10 约束，经扩展槽承载；不能据此省掉特性规则。ASR/ASA 使用 RFC 6733 §8.5.1–2 的独立语法：ASR 的 Session、Origin、Destination、Auth-App 必需；ASA 用正常结果及可选错误诊断，不要求 CCR 的 CC 编号或 Service-Context。ASA 2001具有已终止的协议含义；外部会话事实由应用负责，库不独立核验，不能仅因接纳ASR自动成功。5002/5012分支见§29。

**应用覆盖的具体结论：** 所选 32.299 的顶层 CC-Sub-Session-Id、Acct-Multi-Session-Id、Service-Identifier、RSU/USU/GSU、FUI、Validity-Time、Check-Balance-Result 等有 “not used” 项；RSU/USU/GSU 主要走 MSCC，且 MSCC 内 CC-Money 也标为 not used。32.274 又将 SMS 的 Operation Failover 标为不使用。标准发送构造器不得只按通用 ABNF 放行这些路径。接收仍安全解码并报告应用不支持/不符合，不凭 “not used” 自动猜一个线路错误，也不静默丢弃后报完整成功。

原有 RO-01 余额检查、金额/退款旧模板仍须交付：它们属于明确的 DCCA/旧对端组合 Profile 与纯构造解释能力，不能被新标准线删除，也不能冒称该固定 SMS Ro 线全部支持。相同 App-ID 的不相容解释继续遵守 DRA 部署绑定限制。

## 3. Grouped 与主流程叶字段闭包

### 3.1 公共、Sh 与 SMS

各行给出字段路径、基础类型和出现条件；同名字段按完整父路径审计。U32/U64/I32/I64 为对应固定宽度整数；Enum 为有符号 32 位枚举；Octets 不作 UTF-8 或电话号码推断。标识、重复容器和原始字节均有大小上限。

| 父路径 / 定义 | 必须保留或检查的子字段与类型 | 规则/反例 |
| --- | --- | --- |
| VSAI / RFC 6733 §6.11 | Vendor-Id U32!；Auth-Application-Id U32?；Acct-Application-Id U32? | 两种 App-ID 按该语法的互斥/存在规则；Sh 使用认证应用。SMS 忽略例外不扩展到 CER/CEA |
| Experimental-Result / §7.6 | Vendor-Id U32!；Experimental-Result-Code U32! | 与顶层基础结果区分命名空间；10415 的 5001 不是基础 5001 |
| Proxy-Info / §6.7.3 | Proxy-Host Identity!；Proxy-State Octets!；扩展* | 保留重复组与顺序；不能把 state 转字符串或归并不同代理 |
| Supported-Features / 29.229、29.329 应用覆盖 | Vendor-Id U32!；Feature-List-ID U32!；Feature-List U32!；扩展* | 按 Vendor/List-ID 解释位；重复键冲突不能后写覆盖；Sh 与 Ro M 位及协商规则分别处理 |
| Sh User-Identity / 29.329 §6.3.1 | Public-Identity UTF8?；MSISDN Octets/TBCD?；External-Identifier UTF8?；扩展* | 三种身份选择不能套用 SMS User-Identifier 的多身份规则；额外 IMSI 不从此处凭空合成 |
| Sh Repository-Data-ID、Call-Reference-Info | 前者 Service-Indication Octets!、Sequence-Number U32!；后者 Call-Reference-Number Octets!、AS-Number Octets!；各自扩展槽按定义 | 与 User-Data XML 中同义字段的编码不同；不做更新序号的数据库校验 |
| Sh 主体叶字段 / §6.3 | Data-Reference、Subs-Req-Type、Send-Data-Indication、One-Time-Notification、Identity-Set 等 Enum；Expiry-Time 为 Time；Service-Indication/DSAI-Tag/User-Data 为 Octets；Requested-Nodes/UDR-Flags 为 U32；身份/Server-Name 为相应 UTF8/Identity | 完整顶层见 §2；枚举/bitmask 分开；SNR 一次通知条件、PUR 多引用条件沿用既定规则 |
| SMS User-Identifier / 29.336 §6.4.2 | User-Name UTF8?；MSISDN Octets?；External-Identifier UTF8?；LMSI Octets?；扩展* | 至少有身份；可同时有 IMSI/MSISDN。不要导入 S6t §8.4.36 的另一套定义和 Type-Of-External-Identifier |
| SMSMI-Correlation-ID / 29.338 §6.3.3.13 | HSS-ID UTF8?；Originating-SIP-URI UTF8?；Destination-SIP-URI UTF8?；扩展* | 语法可选不替代 OFR 场景要求三项齐全；相关 OFR 必须带 IP-SM-GW delivery outcome。Code 采用 §1.1 裁决 |
| Serving-Node / §5.3.3.6 | SGSN-Name/Realm、MME-Name/Realm、IP-SM-GW-Name/Realm 为 Identity；SGSN-Number、MME-Number-for-MT-SMS、MSC-Number、IP-SM-GW-Number 为 Octets；IP-SM-GW-SBI-Support-Indicator Enum；各 0..1，扩展* | 组合表分别校验，允许 MSC+MME；IP-SM-GW 缺 Realm 不可一律拒绝；电话号码不能作为 Diameter host |
| Additional-Serving-Node / §5.3.3.7 | SGSN Name/Realm/Number；MME Name/Realm/Number-for-MT-SMS；MSC-Number；各 0..1，扩展* | 不是 Serving-Node 的完全别名，不自动加 IP-SM-GW 成员；保留 primary/additional 区别 |
| SMSF-3GPP/Non-3GPP-Address / §5.3.3.35–38 | 各自 Number Octets、Name/Realm Identity、SBI-Support-Indicator Enum，按对应组定义 | 可保留/解释能力不等于库会执行 SBI；不得伪造成 MME 目标，使用原 SRR 特性判断答案条件 |
| SM-Delivery-Outcome / §5.3.3.14–20、27–28 | MME/MSC/SGSN/IP-SM-GW/SMSF-3GPP/SMSF-Non-3GPP 六种子组；每组 Cause Enum?、Absent-User-Diagnostic-SM U32? | 结构容许缺项不代表业务报告有意义；按 RDR/OFR 条件检查，不能丢掉当前调用未消费的分支 |
| SM-Delivery-Failure-Cause / §6.3.3.5–7 | SM-Enumerated-Delivery-Failure-Cause Enum!；SM-Diagnostic-Info Octets?；扩展* | OFA/TFA 的 5555 要求此组；MO/MT 允许原因子集不同，诊断不得当 UTF8 |
| EPS-Location-Information / 29.272 §7.3.111、115–116 | MME/SGSN-Location-Information 各 0..1；前者 ECGI/TAI、地理/测地信息、当前位置标识、年龄、User-CSG、eNodeB/Extended-eNodeB；后者 CGI/LAI/SAI/RAI、地理/测地信息、当前位置标识、年龄、User-CSG | Octets、Enum、U32 按各叶定义；不可将整个组声明为 RawAVP 后声称语义完整。§13.3已补MME下ECGI/TAI值链；地理/CSG/其他节点编码仍列P1 |
| SMS 主体叶字段 / 29.338 表 5.3.3.1/1、6.3.3.1/1 | SC-Address、SM-RP-UI、SM-RP-SMEA、LMSI、SMS-GMSC-Address 为 Octets；MTI/Not-Intended 为 Enum；flags/MWD/诊断/Delivery-Timer 为 U32；Start/Maximum-Retransmission/Requested-Retransmission/Availability 为 Time | SC-Address 与 SMS-GMSC-Address 均只含 TBCD 数字，不带 TON/NPI；不要套用 MAP AddressString 的前导字节，也不能套用 Ro 的 Diameter Address 编码 |

### 3.2 Ro SMS 主路径与剩余可选闭包

| 路径 | 字段/类型和条件 | 本库设计决定 |
| --- | --- | --- |
| Subscription-Id | Type Enum!、Data UTF8!；可多组 | IMSI 与 MSISDN 分开，顺序和重复不能以 map 丢失 |
| User-Equipment-Info | Type Enum!、Value Octets! | Type=0 是 IMEISV，不是 IMSI；未知设备信息不由订阅身份补造 |
| MSCC 单位与关联 | RSU?、GSU?、USU*；Service-Identifier U32*、Rating-Group U32?、Result-Code U32?、Validity-Time U32?、Reporting-Reason Enum*、Refund-Information Octets? | 请求/答案各自允许的单位方向受检；服务匹配用原 CCR，不能“取第一个 MSCC”或把所有结果折叠成顶层 2001 |
| RSU/GSU/USU 叶字段 | CC-Time U32；CC-Total/Input/Output-Octets 与 CC-Service-Specific-Units U64；GSU 的 Tariff-Time-Change 为 Time；USU 的 Tariff-Change-Usage 为 Enum | 各字段 Presence 独立。CC-Money 的通用定义与固定 3GPP 使用限制分开；空 RSU、缺 GSU、零单位分别验收 |
| Money/Cost/Remaining-Balance | Unit-Value：Value-Digits I64!、Exponent I32?；Currency-Code U32；Cost-Unit UTF8? | CC-Money 的 Currency-Code 在 RFC 4006 语法中可选，Cost/Remaining 的条件不同；不得用一个 Grouped 的必填规则覆盖另一个；业务金额不用 float |
| G-S-U-Pool-Reference / FUI | Pool-Identifier U32!、CC-Unit-Type Enum!、Unit-Value!；FUI 的 Final-Unit-Action Enum!，可有 Restriction-Filter-Rule/Filter-Id*、Redirect-Server?（类型 Enum!、地址 UTF8!） | 能解释配额/重定向指示，不自动安装过滤器、充值跳转或启动业务重试；不支持的配额特性明确报给应用 |
| Service-Information | SMS-Information、MMS-Information；另有 PS/IMS 等可选分支 | SMS 实际使用 MMS 子模型有 32.274 依据，不把 MMS 全部删除；PS/IMS 的全部深层依赖尚未闭合，不声称完整 charging 支持 |
| SMS-Information 直接叶字段 | SMS-Node Enum、Client-Address Address、Originator-SCCP-Address Address、SMSC-Address Address、Data-Coding-Scheme I32、SM-Discharge-Time Time、SM-Message-Type Enum、SM-Protocol-ID Octets、Reply-Path-Requested Enum、SM-Status Octets、SM-User-Data-Header Octets、Number-Of-Messages-Sent U32、SM-Sequence-Number U32、SM-Service-Type Enum、SMS-Result U32、SM-Device-Trigger-Indicator Enum、MTC-IWF-Address Address、Application-Port-Identifier U32、External-Identifier UTF8；各 0..1 | 另有 Originator-Interface?、Recipient-Info*、Originator-Received-Address?、SM-Device-Trigger-Information?；这是 §7.2.211 的全部具名直接成员。端口定义已沿 29.337→29.368 核对，见 §9；消息/角色/category 条件仍须叠加 |
| Recipient-Info | Destination-Interface?、Recipient-Address*、Recipient-Received-Address*、Recipient-SCCP-Address Address?、SM-Protocol-ID Octets? | 每组对应一个接收方；PID 的 MO/MT 路径不同，不能在归一化时移父节点 |
| Originator/Destination-Interface | Interface-Id UTF8?、Interface-Text UTF8?、Interface-Port UTF8?、Interface-Type Enum? | 不能只保留 Type；应用短号码与接口上的应用标识不是同一字段 |
| Originator/Recipient/Received-Address | Address-Type Enum?、Address-Data UTF8?、Address-Domain?；Recipient-Address 另有 Addressee-Type Enum?；Address-Domain 内 Domain-Name UTF8?、3GPP-IMSI-MCC-MNC UTF8?，出现该组时至少有一项 | 电话号码字符串与 Diameter Address、TBCD 严格区分；地址变换由应用显式给出，库不做号码归一化 |
| MMS 中 SMS 使用的叶字段 | Originator-Address、Submission-Time Time、Priority Enum、Message-ID UTF8、Message-Size U32、Message-Class（Class-Identifier Enum?/Token-Text UTF8?）、Delivery-Report-Requested Enum | 按 32.274 category/用途检查；Delivery Report 与普通提交分开，Priority 不任意接受 low，退款/重传不改原 Submission-Time |
| SM-Device-Trigger-Information | MTC-IWF-Address、Reference-Number、Serving-Node、Validity-Time、Priority-Indication、Application-Port-Identifier | 作为明确可选分支列入闭包；§9 已补端口/引用号/优先级定义和 Serving-Node 的应用差异，完整节点叶及样本仍需补；不因有容器就承诺 T4 业务能力 |

完整 MSCC 还包括 Tariff-Change-Usage、Time/Volume/Unit-Quota-Threshold、Quota-Holding/Consumption-Time、Trigger、PS-Furnish-Charging-Information、AF-Correlation-Information、Envelope/Envelope-Reporting、Time-Quota-Mechanism、Service-Specific-Info、QoS-Information、Announcement-Information、3GPP-RAT-Type、Related-Trigger 与扩展槽。已盘点这些入口，尚未完成所有递归叶和条件；它们不得混入“已审计 SMS 主路径”标记。资源和级联数量同样必须计入，不允许未知深层组无限展开。

主路径已经可明确的叶级长度也进入验收：SM-RP-UI 上限为 200 octets（29.338 §6.3.3.3），SM-Status 恰为 1 octet（32.299 §7.2.209）；U32/I32/Enum/Time 为 4 octets，U64/I64 为 8 octets。协议上限与更小的本地资源限额分开报告；不能用一个宽泛 MaxFieldBytes 放过 201 字节 SM-RP-UI。Address 必须保留两字节 family 与对应内容，SCCP E.164 使用 family 8 和 UTF8 地址内容；纯文本、IPv4 四字节和完整 Address 不是可交换表示。

容量计量先检查 AVP 8/12 字节头、声明 payload、四字节对齐及父边界，再检查具体字段上限；Group 子节点、重复列表、诊断和 XML 扩展分别计数。保留 raw 与 typed 两份内容时都计入 DecodeBound，不能只按线路长度计一次。实际 Go 对象系数和所有可选深层组的 bound 仍待实现验证，本轮不虚构 heap 数字。

### 3.3 OM/OC 条件的有限输入

32.240 §5.4 的 M/C 与 OM/OC 不是 AVP 的 M 位：OM/OC 涉及运营方是否配置生成该项。第十四轮修正：32.299 §7.1.12 规定，Service-Context-Id 采用最短 `Release.service-context@domain` 时，所有 OM/OC 均为可选；包含运营方标识时才结合已知运营方规则检查配置后的出现条件。32.274 的应用 category 与该例外必须共同考虑，不得统一把 OM 变成必填或无条件可省。具体检查顺序见 §10.3。

设计补充仅落在 `app/ro` 纯构造入口：`SMSBuildOptions.FieldSelection` 表达一份有限、不可变的具名字段选择（构造器候选名，非已实现 API）。构造 helper 可默认选择启用支持范围内的 OM/OC，这是本地输入约定，不是最短 Service-Context-Id 的标准必填规则；显式省略由应用配置输入，库不替运营方作选择。父组被省略而子字段启用、已禁用字段仍提供值等组合报本地 FieldError；对 C/M 不提供省略开关。收费事件的 MO/MT、IEC/ECUR、提交/报告和已知输入由原有事件值表达，不新增通用 facts。

该选择只服务本地消息构造，不是公共规则注册器、数据库、热配置或对端事实认证；不进入 `validation.Context`。低层 codec 仍区分线路语法通过与 SMS 应用条件通过，不能在未给定构造条件时宣称后者完整。接收方不能用自己的省略配置拒绝对端缺少 OM 项；若部署约定了对端规则，应是明确的具名应用策略。只影响发送选择的配置不改变 WireFingerprint；改变入站接受语义的约束必须进入 InboundSemanticsFingerprint。

## 4. 结果分支和条件字段

| RuleID / 分支 | 必须检查与保留 | 禁止推导 / Planned 反例 |
| --- | --- | --- |
| RESULT-ENVELOPE | 先验证请求/答案关联和 E 位语法，再解释基础/实验结果；保留 Vendor、Code、层级和 Presence | 基础 5001（不支持 AVP）与 10415/5001（用户未知）混淆；答案再生成答案；错误代理冒充成功终端 |
| SH-RESULT | 基础成功/错误与 29.329 §6.2、29.229 的实验失败分开；实验失败不携带基础 Result-Code。5100–5108、5008、5011、5002、4100/4101 及 29.229 引用的结果按具体命令/数据操作解释 | 不用“结果非零即失败”或全局 XOR；SNR 成功不保证每次 PNR 对应本进程订阅；PUA 成功不是数据库提交回执 |
| S6C-SRA-RESULT | 2001 与 User-Name/Serving 节点条件关联；标准实验错误含 5001、5550、5556、5557；非投递查询、MSISDN-less 查询与常规投递查询分别校验 | 不能要求所有成功 SRA 一定有 Serving-Node，不能把 dummy IMSI 当可投递身份；保留各域诊断 |
| S6C-RDA-RESULT | 结果确认本次报告；5001/5558 等失败及可选 User-Identifier 按 §5.2.3 检查 | RDA 不证明短信已投递；不触发库内 MWD/业务重试状态 |
| S6C-ALA-RESULT | 结果确认 alert 请求；保留 §5.2.2/§7 允许的基础/实验失败 | 不因 ALA 成功就自动发送待投递短信，ALR/ALA 不创造持久订阅 |
| SGD-OFA-RESULT | 5555 必须有 Failure-Cause；MO 侧原因子集为 3..6，诊断可选；5552 及 §7.3.2 明确的来源/用户未知分支分别处理 | 只有 Experimental-Result 而无必需 Cause；将 0..2 的 MT 原因套给 MO；以计时超时自动选择一种业务原因 |
| SGD-TFA-RESULT | 5555 要有 Cause（MT 原因 0..2）；5550 可带缺席诊断。Requested-Retransmission-Time 只用于相应缺席分支、原 TFR 有 Maximum-Retransmission-Time，且不得晚于该上限；相等边界与严格早于的构造子集见§20.3 | 缺原 TFR、跨 era 时未经解释直接比较 U32、成功结果带仅失败分支字段；SMS Router 的 User-Identifier 条件另检查 |
| RO-EVENT | EVENT 编号为 0；Requested-Action 区分 Direct Debit、Refund、Check Balance、Price Enquiry；仅固定支持线/兼容线允许的操作进入对应纯解释器 | 成功 CCA 不等于“已授信/有余额”；余额检查的 NO_CREDIT 不转成成功扣费；未支持的动作不偷换成 Direct Debit |
| RO-RESULT | 保留顶层及每个服务结果；4010、4011、4012、5003、5030、5031 有不同含义；5031 需要 Failed-AVP；不认识的枚举/结果按固定规则保留/报告 | 4011 是单独的“不再适用信用控制”指示，不能一概当重试失败；顶层成功+服务失败、错服务/不合适用条件的MSCC重复（不把合法费率双组误判冲突，见§10）、缺/零 GSU 分别验收 |
| RO-REFUND | 标准 10415/2022 Refund-Information 走 MSCC，保留 bytes；若应用声明原 IEC 返回该凭据，退款构造必须原样带回 | 不重新编码为字符串、不制造凭据、不把 0/1292 当标准字段；是否退款、原扣费记录和恢复归应用 |
| RO-FEATURE-RESULT | 初始请求无特性、特性均支持、部分不支持、完全不识别 Supported-Features 四分支分别保留；5001+Failed-AVP 与实验 5011 不混淆 | 5011 的伴随 Result-Code 存在规范未决，见 §1.1；库不猜成功，不在这里新增降级重试引擎 |
| RO-FAILURE-HANDLING | CCFH/DDFH、Redirect、失败结果作为协议输出；与传输失败证据分开 | 业务失败不授权 R33 原身份重传；DDFH 值不授权库持久排队，Redirect 不授权跨 OCS 自动迁移 |

错误或未知结果的完整原始值须能随 Completion 返回给应用；解释器的“不支持/不完整/冲突”与对端实际返回失败不能混成同一个本地 error。必须先取得可信 envelope，诊断超限不扩大资源预算。

退款依据还须分层记录：32.274 §6.2.1 的表后仍有退款信息承载机制待研究的编辑注释；本页 Refund-Information 的类型和回送规则直接依据 32.299 §7.2.171 及其 IEC 退款程序。不能据该字段已定义便宣布 32.274 的退款细节全部收敛，也不能用编辑注释抹掉 32.299 已明确的字段义务。具体 OCS 是否返回凭据、旧模板如何承载及配套 CCA 仍需独立证据；库不补造完整退款业务流程。

## 5. Sh XML 主路径

官方 ZIP 的 XSD 无 import/include/redefine，无 targetNamespace。短信数据位于 `Sh-Data/Sh-IMS-Data/Extension/Extension/Extension` 的具名子元素下；不能扫描任意层同名元素作为合法数据。该位置的 SMSRegistrationInfo 与 UEReachabilityForIP 均为 0..1。

- SMSRegistrationInfo 的 IP-SM-GW-Number 在 XSD 中必须出现，但其 string 类型允许空串；注销语义应结合 29.328 §6.1.2.1 的步骤 4b 和 §7.6.16，不能用“必需字符串非空”误拒绝注销。SCAddress 位于下一层 Extension；SBI indication 在再下一层。
- UEReachabilityForIP 的 MME、SGSN、AMF3GPP/AMFnon3GPP 信息处于不同扩展层，分别保留 Presence；不把 absent 当 false，不用深度搜索把未来字段覆盖成旧字段。
- schema 使用 sequence 和 lax 扩展；只在允许位置保留有界未知元素，不能以 lax 为由放宽已知字段类型、顺序或 multiplicity。有限 XML helper 不自动解析所有 Sh-Data，原 User-Data bytes 仍可保留。

XSD 本身没有网络依赖，库也不按 xsi:schemaLocation 联网取 schema。DTD/外部实体、深度、节点数、文本字节、未知扩展字节均有硬上限；这是解析器边界，不是新增 XML 服务。Diameter 重复比较继续按 User-Data 原字节，不做 XML C14N、元素换序或号码归一化。待实施样本包括注册、空值注销、一次可达通知、错扩展深度、重复元素、外部命名空间及超限。第二十轮已补前述数据/结构正反候选并做离线schema检查（§16），库级预算/安全执行仍待实施。

## 6. 六个旧模板的实际差异

本轮只读重查 `sms-fw/c-cpp/libsrc/kernel/dialib/` 六个仓库模板；不以模板等同现场报文，也不改旧系统。

| 模板 | 已逐项发现的具体问题/覆盖 | 迁移决定 |
| --- | --- | --- |
| yppdia_avptemplate.cf | 顶层 Service-Identifier；私有 10425/23；MSCC/RSU 的 CC-Money；Client-Address 使用 IPv4 写法 | 标准 3GPP 路径和私有/旧用法分开；§14已确认普通IPv4编码入口加family、draft分支另有布局；实际构建/模式与对端bytes仍待核对 |
| yppdia_iec_avptemplate.cf | Type=0 的 User-Equipment-Info 却填 IMSI；Client/SMSC/SCCP 地址用 STRING；Destination-Interface 和 Recipient-SCCP-Address 直接放 SMS-Information | 标准新构造使用真实设备身份、Address 编码和 Recipient-Info 父路径；保持旧布局需具名兼容和对端样本 |
| yppdia_init_avptemplate.cf | 与 IEC 相似，方向常量不同，省去 Originator-SCCP-Address | INITIAL 条件、MO/MT 角色和实际输入分别审核；不能拿 IEC 模板重命名证明 ECUR 支持 |
| yppdia_term_avptemplate.cf | USU/Reporting-Reason、SM-Status、SM-Discharge-Time；仍沿用上述设备身份与地址布局 | 用量与结束条件归应用；Discharge-Time 的报告条件不能仅凭 TERMINATION 判定满足；保留原 Session/编号关联 |
| yppdia_refund_avptemplate.cf | CC_MONEY 标为 0/420（实际标准为 CC-Time）；私有退款 token 为 0/1292；标准 CC-Money 是 0/413，Refund-Information 是 10415/2022 | 不按名称覆盖数值定义，不自动“修好”发给旧 OCS；同一 Vendor/Code 的冲突不能靠加载顺序解决 |
| yppdia_avptemplate_scap1.cf | Vendor 193、Acct-App 19302、553/554/555、606 等自有编码及 Service-Parameter-Info 重复项 | 独立兼容 Profile；仍保留验收义务，不映射成标准 Ro 或对接 C ABI |

六模板共有的 Subscription-Id、Event-Timestamp、单位/金额、SMS/MMS/接口/地址字段已与 §3 建立分组映射；§14已补解析/Fetcher/编码/选择链及条件化字节推导；仍缺旧encoder执行后的完整字节、部署模板版本、私有字段规范和独立CCA/ACA样本。固定标准 Profile 不吸收这些不确定性；公开扩展机制也不能绕过已知 Code 的类型冲突。退款关联可由应用向纯构造器提供原结果/凭据，不需要库保存扣费历史。

## 7. 参考库的具体取舍

- [go-diameter 本地字典](../../go-diameter/diam/dict/testdata/tgpp_sh.xml)（本轮 HEAD `f26e3528006fcf0d33cb663994a0ec0ccbf7ef55`）用于发现缺项和类型差异。其 [credit_control.xml](../../go-diameter/diam/dict/testdata/credit_control.xml) 将 CC-Money 内 Currency-Code 标 required，与 RFC 4006 的可选语法不同；不能把该字典直接升级标签后称为固定 Rel-18 支持。
- [OTP diameter_dict](https://www.erlang.org/doc/apps/diameter/diameter_dict.html) 的生成记录区分单值/列表、继承和固定语法。本库参考 Presence、次数和显式来源；协议程序条件仍单独审计，字典生成不等于应用结果正确。
- [freeDiameter 固定源码](https://github.com/freeDiameter/freeDiameter/blob/9d4bc47584bfeced9de708fe00f5629ac4689db5/libfdproto/messages.c) 分开检查 min/max、禁止项、固定位置。本库借鉴检查维度；其诊断构造附近仍有 TODO，故不把参考库错误答案直接作为唯一标准 golden fixture。

参考实现、规范样本和真实对端样本应形成互相独立的证据；新库自编码再自解码只证明内部一致。静态资料审查不算运行过上述库的互操作。

## 8. R29/R34 当前门槛

递归依赖按下表继续核对；第十三至十五轮新依赖和规则见§9–11，其他未闭合项不声称已冻结。只有字段表示/程序条件实际用到的部分进入库，不把引用整本规范当作新增功能：

| 字段/用途 | 尚需核对的定义链 | 关闭证据 |
| --- | --- | --- |
| HSS-ID、External-Identifier、LMSI、IMEISV | 已固定 23.003 V18.5.0；External-Identifier 已核对 RFC 4282/23.682 的基础范围，国际化依赖待补；Sh地址23.008已由§17取得并闭合所选表示子集 | HSS-ID 格式/LMSI 长度已核对；外部标识国际化、SMS设备使用位置和独立验证仍待补 |
| SMSMI 三身份与 OFR delivery outcome | 已固定 23.204 V18.0.0 的 MSISDN-less 程序 | §9 已按五种请求展开条件；§10补通用URI/接口边界及候选，完整使用位置和独立验证仍待补，不实现IMS业务流程 |
| 地理/测地、CSG、eNodeB/Extended-eNodeB | 29.272 对应叶条款实际引用的 23.032、23.003、36.331 等 | 每个所用编码的位宽/范围/版本；不把 ASN.1 或地理库整体引入 |
| Ro 地址子模型、PS/IMS 可选字段 | §13已固定OMA MMS1.3、29.061 V18.1.0及RAT/位置/时区实际依赖；32.251 V18.0.0、32.260 V18.3.0已取得 | 接口/PLMN地址、PS LTE值子集见§13；完整OMA地址、IMS/QoS等仍P1，与32.270分开记录 |
| Ro Device Trigger/Port/Reference/Serving | 已固定 29.337 V18.0.0→29.368 V18.0.0；端口继续引用已取 23.040 | §9 已核对三个叶及 Ro/T4 Serving-Node 差异；节点深层编码、全部条件和独立样本仍阻塞完整分支 |
| DRMP/OC/Load 的算法能力 | RFC 7944/7683/8583 及实际选择的扩展 | 字段解析与行为支持分别验收；没有实现算法就不广告该能力 |

| 优先级 / 工作项 | 已完成的文档证据 | 尚需补齐，不能提前关闭 |
| --- | --- | --- |
| P1 / 主命令与结果 | Sh/S6c/SGd 顶层盘点、Ro 顶层及应用覆盖、结果命名空间、主条件和模板差异 | 每个叶字段的 Code/Vendor/flags/错误/数量/预算到 fixture 的机器可核对清单；不得把本页分组表当其替代 |
| 首版Ro P1 / Ro明确未决 | 已识别5011配对空白，修正最短Service-Context-Id例外，补核心单位和结果优先级 | 5011 分支裁决与独立样本；所有用到的深层 MSCC、PS/IMS、地址/业务字段依赖；不能凭任选基础结果关闭 |
| P1 / 依赖与 Sh | 累计30份PDF、Sh ZIP/XSD、4份RFC TXT及4份CR材料共40项哈希；其中3份PDF仅作版本对照，XSD依赖为空且主路径已核对 | 新取得依赖只审所列字段；29.061及选定地址/位置值链已补§13；外部标识及其余地址/节点依赖继续闭合；32.251/260已取但只审部分字段，完整勘误/CR/8506 差异未审完 |
| 后续兼容阶段P1 / 兼容发布 | 六模板冲突与§14的14文件源码链/条件化片段已登记；RO-01/SCAP1/S6a义务保留 | 私有定义、最终 bytes、实际部署模板和错误 CCA；不能用 raw 透传标已完成 |
| P1 / 验证 | 下述 FIELD/RESULT/DEP/XML/TEMPLATE 验收项已定义 | 独立正反 bytes/XML、实际 decoder/encoder、资源/竞态及双端互操作均未运行；实现仍待用户指令 |

验收条目：FIELD-01 顶层枚举与漏项；FIELD-02 Grouped 递归/类型/父路径；RESULT-01 命名空间与 E 位；RESULT-02 OFA/TFA 条件；RESULT-03 Ro 顶层/MSCC/5011/缺与零；DEP-01 原件/依赖/勘误；XML-01 固定附件与短信路径；TEMPLATE-01 六模板标准/私有/冲突；PROVISION-01 OM/OC 及发送/接收边界。均为 Planned，不是已通过测试。

下一轮应围绕上表具体缺项补证据与规则，不再新增持久化或泛化框架。R32/R33/R36 与 SCTP 工作包不因本次字段审计而自动完成；公共库能力发布与具体 DRA/OCS 部署验收继续分开。

## 9. 第十三轮：身份/触发字段与结果处理切片

### 9.1 新取得的依赖及采用范围

以下原件于 2026-10-30 下载并核对封面、相关条款及 SHA-256。选择固定 Rel-18 修订用于所列字段，不声称最新；四份原件加在 §1 的 13 份 PDF 之后，不替换主应用规范。没有将整本规范复制进仓库。

| 固定来源 | SHA-256 | 审查范围 |
| --- | --- | --- |
| [TS 23.003 V18.5.0](https://www.etsi.org/deliver/etsi_ts/123000_123099/123003/18.05.00_60/ts_123003v180500p.pdf) | `a5b25cccc2081e5483607a0677663d2757e83c801f53c91caeb4902b88122f91` | §2.5 LMSI、§5.3 HSS-ID、§6.2.2 IMEISV、§19.7.2 外部标识；非全本符合性审核 |
| [TS 23.204 V18.0.0](https://www.etsi.org/deliver/etsi_ts/123200_123299/123204/18.00.00_60/ts_123204v180000p.pdf) | `72bfaa13f1a9fd5f7bdf847431cf8630a3f92e686597f4c2f80d37021b72e709` | §6.3a.1/2 的失败后转 SMSC 与重投递程序，用于判断字段来源和职责 |
| [TS 29.337 V18.0.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129337/18.00.00_60/ts_129337v180000p.pdf) | `d8c3ba536a4a0888f247e985d1bdbc4e4d78f745a72aba3c8f542fbbd0d02503` | 表 6.3.1/2 的复用链、§6.3.3/4 节点组；不加入 T4 命令支持 |
| [TS 29.368 V18.0.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129368/18.00.00_60/ts_129368v180000p.pdf) | `ae2535c18df292ab0fd4ef0fc16d306092386b67b9e6ad7872b3b45fecef34b7` | 表 6.4.1/1、§6.4.7/8/11 的优先级、引用号和端口；不加入 Tsp 业务 |

依赖审查发现两类不能靠名称继承的问题：32.299 表中的 Application-Port-Identifer 拼写、正文 Application-Port-Identifier 与 29.337 的 SMS-Application-Port-ID，最终按 29.368 的 `(10415,3010)` 定义对齐，生成模型只用一个规范名；29.337 的 Serving-Node 则含 SMSF 名称/realm/号码，不能直接复制 29.338 的节点组内容。前者是本库有来源记录的命名裁决，后者必须生成按应用/路径区分的组规则，不以相同 `(10415,2401)` 合并接受集合。

### 9.2 可追踪字段规则

这是 RULE-01 的已展开切片，仍非全量字典。以下输入只来自当前消息及固定 binding；没有数据库、特性服务或任意 facts。每项需生成独立正反 fixture，统一由 FIELD-03/04 验收，目前均 Planned。

共同处理约定：发送检查下表类型/次数/条件；接收按 APP-MBIT 等应用覆盖，不对已理解的 Sh/SMS AVP 误用严格 M 位相等。P/保留位按 BASE-FLAGS；Canonical 保留有效 payload、完整父路径及同键序，仅屏蔽已审核的忽略位。缺必需项/禁止项/超协议次数分别按 5005/5008/5009，已安全识别的非法值/长度按 5004/5014；本地构造用 FieldError，本地容量限制用 ResourceLimit。以上线路错误只对可信请求 envelope 适用；应答按 SPEC-ERROR-01，不回错误答案。

单字段线路上界记 `W(v,n)=align4((v?12:8)+n)`，n 是 payload 字节数；此处上界包含 padding，不把 padding 计进 AVP Length。所有乘加须溢出检查。整组再加父头与其各子项的 W，受 MaxDepth/MaxAVPs/MaxRepeated/MaxFieldBytes/MaxFrameBytes 限制；Decode、Canonical、Diagnostic 的实际 Go 空间仍由各 bound 计量，不能从 W 推导零分配。

| RuleID / 完整使用位置 | Vendor/Code、发送 flags、类型/次数 | 来源、检查及反例 | 线路上界 / 比较 |
| --- | --- | --- | --- |
| SMSMI-GROUP；SRR/ALR/RDR/OFR/TFR 顶层 | 10415/3324，V=1 M=0，Grouped 0..1 | 29.338 表 6.3.3.1/1 与 §6.3.3.13；组内三子项各 0..1、扩展*；命令条件见下一表，空组不能通过适用条件 | 父头 12 + 所有有界子项 W；默认 Grouped 顺序，不跨父节点搬移 |
| SMSMI-HSS；上述组/HSS-ID | 10415/3325，V=1 M=0，UTF8String 0..1 | 29.338 §6.3.3.14→23.003 §5.3；十进制 MCC(3)+MNC(2或3)+非空 index，总长≤15，即格式长度6..15；非数字、5/16位为反例 | 最大 W(1,15)=28；保留前导零，不变整数，不猜 MNC 位数 |
| SMSMI-ORIGIN；上述组/Originating-SIP-URI | 10415/3326，V=1 M=0，UTF8String 0..1 | 29.338 §6.3.3.15；按命令的出现规则检查；§10.2补语法/接口边界，完整使用位置与独立验证未闭合，不以UTF8合法标完整URI通过 | W(1,MaxFieldBytes)；原字节保留，不折叠大小写/转义 |
| SMSMI-DEST；上述组/Destination-SIP-URI | 10415/3327，V=1 M=0，UTF8String 0..1 | 29.338 §6.3.3.16；缺目的 URI、重复、错误位置分别验收；URI 完整语法同上 | 同上；不得拿 Originating 的值补空 |
| SMS-ID-GROUP；ALR/RDR/OFR 的 User-Identifier 等已声明路径 | 10415/3102，V=1 M=1，Grouped；父命令次数见 §2 | 29.336 §6.4.2；User-Name、MSISDN、External-Identifier、LMSI 各0..1；至少一个具名身份，只有扩展项不满足；不导入 S6t Type-Of-External-Identifier | 父头+有界子项；保留多身份，四个子定义不能都当同一种 string |
| SMS-LMSI；SRA/LMSI、User-Identifier/LMSI | 10415/2400，V=1 M=1，OctetString 0..1 | 29.173 §6.4.2→23.003 §2.5；恰4字节；3/5字节非法。全零不能作为有效 VLR 分配号，但无当前应用额外依据时不全局误判为线路格式错 | W(1,4)=16；保留全零原值/诊断，不自行分配或查 VLR |
| SMS-RPUI；OFR/TFR 的 SM-RP-UI，OFA/TFA 可选同字段 | 10415/3301，V=1 M=1，OctetString；请求1、答案0..1 | 29.338 §6.3.3.3；payload≤200字节，201字节非法；仅有界承载，不宣称已解析 TPDU 内容 | 最大 W(1,200)=212；内容改变必须比较出差异 |
| RO-SMSTATUS；CCR/Service-Information/SMS-Information/SM-Status | 10415/2014，V=1 M=1，OctetString 0..1 | 32.299 §7.2.209；恰1字节，报告/category 条件另叠加；缺值与0x00不同 | W(1,1)=16；Presence 和原值均保留 |
| RO-TRIGGER-REF；CCR/Service-Information/SMS-Information/SM-Device-Trigger-Information/Reference-Number | 10415/3007，V=1 M=1，U32 0..1 | 32.299 §7.2.205B→29.337表6.3.1/2→29.368 §6.4.8；事务引用由应用提供，不借用 HBH/E2E 分配器 | W(1,4)=16；保留值，不验证外部事务历史 |
| RO-TRIGGER-PRIORITY；同上/Priority-Indication | 10415/3006，V=1 M=1，Enum 0..1 | 同引用链→29.368 §6.4.7；发送支持0/1；未知值不默认成普通优先级，按固定枚举规则报告 | W(1,4)=16；与 MMS Priority 分开，不能同名强转 |
| RO-TRIGGER-PORT；SMS-Information/Application-Port-Identifier 及其 SM-Device-Trigger-Information/Application-Port-Identifier 两路径 | 10415/3010，V=1 M=1，U32，各0..1 | 29.368 §6.4.11→23.040 §9.2.3.24.4；线路4字节，应用端口语义16位，65535/65536边界；不能编码成2字节 AVP，也不解析/重写 TPDU | W(1,4)=16；两个路径分别存在，不互相复制或归并 |

HSS-ID 的数字结构不能证明“不与某个 UE 的完整 IMSI 相同”，该分配约束归运营方；库不查询用户全集，也不把任意 IMSI 自动标成 HSS-ID。External-Identifier的字符/长度和版本边界、IMEISV在PS/IMS中的表示差异由§10继续细化；国际化、SMS设备实际使用位置及独立报文仍未闭合。不能仅用一次`@`分割完成验证，也不从旧STRING模板臆测与TBCD可互换。

### 9.3 SMSMI 的命令条件与业务边界

依据 29.338 §5.2.1、§5.2.2、§5.2.3、§6.2.1、§6.2.2 的信息元素表，不能把 §6.3.3.13 的全可选 ABNF 直接当有效业务组合：

| 请求且 SMSMI 组存在 | HSS-ID | Originating-SIP-URI | Destination-SIP-URI | 额外条件 |
| --- | --- | --- | --- | --- |
| OFR | 必需 | 必需 | 必需 | 必须同时有 SM-Delivery-Outcome 内的 IP-SM-GW outcome；不补造投递失败原因 |
| TFR | 禁止 | 必需 | 必需 | 不能直接复制 OFR 的整个组 |
| SRR | 禁止 | 禁止 | 必需 | MSISDN-less 重试时 User-Name 可承载 HSS-ID；不能要求它一定是普通用户 IMSI |
| ALR | 禁止 | 禁止 | 必需 | 关联标识来自已有 MWD 程序，但库不拥有 MWD 表 |
| RDR | 禁止 | 禁止 | 必需 | 仍独立校验 SC-Address 与 delivery outcome；不能与前次 TFR 共用 pending |

23.204 §6.3a.1/2 解释了失败相关信息进入 SMSC 后的再投递来源；这些程序不能被搬成库内存储/计时/调度引擎。以上条件可以从当前请求判定，无需第三类 validation.Context。S6c 解释 HSS-ID/IMSI 的区别仍用原 SRR 的已验证用途；只有字符长度相同不足以决定身份类型。

### 9.4 Ro 5011：先固定本库处理，不伪造标准配对

§1.1 的标准发送配对未决仍保留，但接收结果、错误和支持声明现在可以明确。下面所有应答先完成安全关联、E 位及正常/错误语法选择；编号和服务不匹配不是这个分支的“另一种合法配对”。

| 实际输入 / 操作 | 确定行为 |
| --- | --- |
| E=0 CCA 只有 Experimental 10415/5011，缺必需 Result-Code | 违反所选正常 CCA 语法；可关联时按运行时 §2.4 以 ErrInvalidAnswer 终结、保留有界诊断，不返回成功对象、不自动重传、不回 Answer |
| E=0 CCA Result-Code=2001 且 Experimental=10415/5011，其他 Wire/Exchange 规则通过 | 原 CCA 可作为带应用报告的接收结果；解释器报告结果冲突/特性失败，MUST NOT 给出扣费成功或完整应用校验通过 |
| E=0 CCA 基础失败且 Experimental=10415/5011，其他规则通过 | 保留两个失败及特性列表，可报告收到特性失败；不称该具体配对已获标准裁决，不由其自动降级或重试 |
| 未识别的 Vendor/Experimental 值或未知特性位 | 保留/报告未支持语义，不能只凭基础2001生成类型化成功；没有新增通用“忽略实验结果”开关 |
| 标准 helper/Respond 自动生成实验5011分支 | 在伴随基础码得到裁决前不提供“标准自动配对”；该分支不能标完整支持，完整 OCS 特性错误生成能力仍为 P1。低层数据表示可表达两字段，不等于业务发送校验授权 |
| 明确旧对端要求一种配对 | 只有独立来源/报文、具名兼容 Profile、发送/接收规则及指纹已登记后才允许；不能由业务传入任意 Result-Code 绕过校验 |

与正常主流程结果一样，字段解释局限不改变传输发送证据。CTF 已收到可关联失败，不因为缺少“理想配对”就重新扣费。新增 RESULT-04 对上述分支逐一验收；本轮没有宣称解决规范本身的空白。

### 9.5 本轮关闭与保留

已完成文档审查：四个依赖版本/哈希，HSS-ID 格式与分配职责，LMSI 宽度，端口/引用号/优先级定义，五种 SMSMI 请求条件，以及 5011 的安全接收/生成限制。第十三轮当时未完成项包括全部递归叶、SIP URI/外部标识/设备编码、Serving-Node全部应用组合、地理/PS/IMS依赖、完整勘误/8506差异及独立fixture。第十四轮具体进展与仍未闭合范围见§10；候选不等于独立fixture通过。

参考 [OTP 字典继承及生成记录](https://www.erlang.org/doc/apps/diameter/diameter_dict.html)，将基础定义和使用位置的规则分开；不能用通用继承吞掉 Ro/T4/S6c 组差异。本地 go-diameter 的 Ro 字典引用了 Application-Port-Identifier 却未在本轮检查的 testdata 字典集合找到同名 AVP 定义，说明“有引用”不是依赖已闭合。保持 [SPEC-PROFILE-01/02](diameterkit-specification.md) 的未解析引用失败门槛，并用独立样本核对，不能靠生成器跳过该字段获得通过。

## 10. 第十四轮：身份语法、服务上下文与核心额度规则

本节继续 R29/R34；只收敛下列规则，不把全部 SIP、NAI、PS/IMS 或 charging 功能纳入支持承诺。新增的[候选协议样本](diameterkit-protocol-vectors.md)由规范手工推导，尚未经独立实现或真实对端验证；不是抓包，也不是已通过的 golden fixture。

### 10.1 新证据与采用范围

| 原件 | 实际取得 / SHA-256 | 本轮采用范围 |
| --- | --- | --- |
| [TS 23.682 V18.0.0](https://www.etsi.org/deliver/etsi_ts/123600_123699/123682/18.00.00_60/ts_123682v180000p.pdf) | PDF，1,407,723 bytes；`e756ad237e11f07762330834a11baa7e5c225c56e283feaf14b90f7c123aa68a` | §4.6.2 External Identifier 的分配/归属职责；不导入 MTC 系统实现 |
| [RFC 3261 文本](https://www.rfc-editor.org/rfc/rfc3261.txt) | TXT，647,976 bytes；`d513777f77fea01a4de9c0a2d9d6713cb53b8231f1b7a2ab56705f8d51b066dc` | §19.1、§25.1 SIP URI 语法及显式比较；不实现 SIP 信令 |
| [RFC 4282 文本](https://www.rfc-editor.org/rfc/rfc4282.txt) | TXT，34,421 bytes；`7274047459eac68c52416cae94fbc178f786dc7153a9bac3a887fa7c92450a57` | §2.1–2.4 与 23.003 §19.7.2 的 External-Identifier 引用关系 |
| [RFC 4006 文本](https://www.rfc-editor.org/rfc/rfc4006.txt) | TXT，288,794 bytes；`036788a26ad4f82bfb71134d21af2fdd185fa95c84ad22cb56499b7f8dce426a` | 复核 §5.1.2、§8.7–8.9、§8.16–8.32、§8.49–8.51；固定 DCCA 基线不自动替换成8506 |

截至第十四轮，累计18份PDF、Sh ZIP/XSD两件与上述3份RFC TXT有原件哈希。另在线核对 [RFC 5954 §4](https://www.rfc-editor.org/rfc/rfc5954.html#section-4) 的 SIP IP-literal 修正，以及 [RFC 7542](https://www.rfc-editor.org/rfc/rfc7542.html) 对4282的替代关系；这两件没有本轮归档文件哈希。取得文件不等于全本/全部更新/勘误审计完毕。

### 10.2 身份检查分层，不让规范化改变请求内容

| RuleID / 位置与来源 | 本轮确定的发送与接收检查 | 保留、失败和剩余限制 |
| --- | --- | --- |
| ID-SIP-SYNTAX；SMSMI 的3326/3327与采用 SIP URI 的身份字段；3261 §19.1/§25.1、5954 §4 | 对整个字段解析 URI：scheme、userinfo、host、参数/转义各用 SIP 语法；IPv6 文字地址使用方括号及5954修正规则。错误 `%HH`、缺host、裸空白/控制符和 `Display Name <...>` 头字段包装不是合法 URI 值；不能复用 HTTP URL 解析成功作为充分条件 | 仍先查 UTF8String、字段长度和父路径。合法未知 URI 参数不因库不使用而一律丢弃/拒绝；解析不做 DNS、网络 I/O 或身份授权。SIPS/tel 在每个 SMSMI 使用位置的允许范围仍需逐条依据，不能从通用 parser 能读出就宣布可用 |
| ID-SH-CANON；Sh/Public-Identity；23.003 §13.4、29.329 身份条款 | Sh 使用指定 canonical 公共身份形式；SIP 形式为 `sip:username@domain`，Tel 形式是无参数/视觉分隔的全局号码且最多15数字。通用 SIP URI 合法不保证是合法 Sh 公共身份 | 不把 Sh 的接口约束复制到 SMSMI。纯检查报告不符合该使用位置的值，不默删参数/端口“修好”；显式身份等价比较若提供，须按对应 URI 标准单独验收 |
| ID-EXTERNAL；User-Identifier/External-Identifier（10415/3111，UTF8String）；23.003 §19.7.2、23.682 §4.6.2、4282 §2 | 采用所选3GPP明确引用的4282规则，External-Identifier 必须有非空 LocalIdentifier 与 DomainIdentifier；泛 NAI 允许的仅用户名或 `@realm` 不足以构成外部标识。用户名存在反斜杠转义，不能按所有 `@` 字符计数分割；必须整串检查 | 保留原值。72/253是NAI最低/建议支持长度，不是Diameter的协议最大长度；实际按字段/消息预算拒绝超额，不报虚构的253字节格式限制。国际化/SASLprep/IDNA完整依赖仍P1，ASCII样本通过不代表完整NAI支持 |
| ID-UE-ENCODING；使用位置上的 User-Equipment-Info/Type459、Value460；32.299 §7.1.17、23.003 §6.2.2 | Type=IMEISV(0) 的Value仍是OctetString：PS规定UTF-8十六进制文本，只有IMEI时以F补到16；IMS规定UTF-8十进制文本，只有IMEI时15位。不能统一编码成8字节TBCD，也不能把PS的F规则强加给IMS | 真实IMEISV为16位数字，不代表每个业务路径都只允许16数字。SMS的顶层与嵌套在线适用性已在§11.2细分，32.251/32.260已取得；顶层SMS真实16位IMEISV子集已由§12.3收敛，仅IMEI映射仍未裁决；不能用“IMS承载SMS”自动选IMS计费规则；旧模板把IMSI装成Type0仍非合法设备身份依据 |

External-Identifier 的全球唯一性、域所属运营方和 LocalIdentifier 分配由应用/部署负责；库不读 HSS 或注册中心证明它们。RFC7542移除了旧反斜杠转义并调整国际化处理，不能无声替换所选3GPP的4282引用，也不能合并两套接受语言后称为同一 Profile。若以后采用修订规范，须登记依据、版本、兼容差异和指纹。

以上检查与 CanonicalV1 的重复请求比较是不同操作：该比较保留身份内容原字节，不做URI等价化、Unicode正规化、域名折叠或号码转换。比如两种等价IPv6拼写，在显式 SIP 身份比较中可以相等，在本库同键请求比较中仍可能是内容冲突；这是保守的合并边界，不应称为完整 SIP URI 等价算法。符合性检查要求的 canonical 输入也不意味着库有权改写收到的身份。解析用既有 MaxFieldBytes/MaxAVPs 约束线性扫描和有界诊断，不新建URI缓存或国际化服务。

### 10.3 RO-SERVICE-CONTEXT：OM/OC 的最短格式例外

32.299 §7.1.12 的 `Service-Context-Id`（0/461，M=1、UTF8String，CCR必需一次）包含服务文档标识。SMS 标识末尾为 `32274@3gpp.org`；`18.32274@3gpp.org` 是Rel-18最短形式，`001.001.18.32274@3gpp.org` 则带MNC、MCC。数字顺序是MNC后MCC，不能从 Origin-Realm 或DRA的CER推算运营方规则。

检查顺序固定如下：

1. 按当前 Profile/使用位置检查服务标识、Release及允许格式；不把任意后缀匹配当支持证明，也不因库用Rel-18字典就替调用方修改其Release。
2. 对最短形式，OM/OC均可选；仍检查实际出现的字段、M/C条件、父路径和命令规则；§11.2明确的在线not-used限制不会因此放开。不能因本地FieldSelection启用了某项，就把远端缺少该项报成Missing-AVP。
3. 对带运营方部分的形式，按已明确的具名规则检查OM/OC条件；未知规则应报告应用检查未完成/缺协议依据，不能推断远端启用、报告违规或默认为已全部通过。该依据固定在已审核的Profile/绑定策略中，不增加运行时Map、回调或第三类Context。
4. FieldSelection只决定纯构造helper接受的本地输入；选择生成而缺输入可以是本地FieldError，须与标准必填错误区分。CCA侧需要服务上下文时取已保存的原CCR证据，不要求CCA携带一个其语法未要求的Service-Context-Id。

本地旧配置 `sms-fw/c-cpp/libsrc/kernel/dialib/yppdia_dcca.cf:138` 使用 `version1.12645.000.000.6.32274@3gpp.org`。这条静态配置证明迁移时必须审计扩展/运营方/Release含义，不证明现网实际发出了这些字节；不能静默改为Rel-18最短形式来绕过旧对端约定。Service-Context-Id原值属于请求内容；本地helper默认选择本身不改变WireFingerprint，但修正规范接受条件须更新规则版本/语义指纹。

### 10.4 RO-UNITS / RO-MONEY：核心叶字段与有界数值

下列基础AVP均Vendor=0、发送M=1/V=0；次数是所在已允许父路径内的语法次数，命令/服务条件另外叠加。W表示含padding线路字节，不含Go对象开销。通用DCCA金额能力保留；32.274 SMS标为not-used的字段不能因此自动进入标准SMS构造。

| RuleID / 使用位置 | Code / 类型 / 次数 | 来源、检查及预算 |
| --- | --- | --- |
| RO-UNITS-GROUP；CCR/CCA已允许的直接或MSCC下GSU/RSU/USU | 431/437/446，Grouped；MSCC内前两者各0..1、USU为0..* | RFC4006 §8.16–19、32.299 §7.1.14。下面每一种核心unit在单个组中各0..1；USU重复不能压成单值。整体由MaxRepeated、深度及总字节预算约束 |
| RO-UNITS-TIME；GSU/RSU/USU的CC-Time | 420，U32，0..1 | §8.21，秒；W=12；不以Go duration的范围替换线路范围 |
| RO-UNITS-COUNT；相同父组的CC-Total-Octets/Input-Octets/Output-Octets/Service-Specific-Units | 421/412/414/417，U64，各0..1 | §8.23–26；每项W=16，缺与显式0不同；解码不转浮点、不静默截断到有符号整数 |
| RO-MONEY-VALUE；CC-Money/Cost-Information/Remaining-Balance内Unit-Value | 445 Grouped必需1；447 Value-Digits为I64必需1，429 Exponent为I32可选1 | §8.8–9；值为digits×10^exponent；省略exponent按0解释但保留Presence。W(Unit)=24或36；有负指数完全合法。只保留有符号二元组；显式数值转换检查范围/精度和计算预算，不分配极大十次幂 |
| RO-MONEY-CURRENCY；GSU/RSU/USU内CC-Money | 413 Grouped 0..1；425 Currency-Code U32 0..1 | §8.22，Currency SHOULD出现但非语法必需；未知币种不猜默认单位。无扩展时W=32–56；缺Currency不能自动变成零币种或沿用上一答 |
| RO-COST-CURRENCY；CCA/Cost-Information | 423 Grouped 0..1；Unit-Value与Currency-Code各必需1，424 Cost-Unit UTF8String 0..1 | §8.7；不同于CC-Money，缺Currency违规。无Cost-Unit时W=44或56；可选文本按字段预算 |
| RO-BALANCE-CURRENCY；CCA/Remaining-Balance | 10415/2021，V=1 M=1 Grouped 0..1；Unit-Value、Currency-Code各必需1 | 32.299 §7.2.172；W=48或60（不含额外扩展）。不能套CC-Money的Currency可选规则 |

相关分支按原CCR和当前CCA检查。空RSU、缺RSU、显式零GSU是三种输入，不归并为默认0；额度、消耗和余额是不同含义。结构/次数/宽度违规走既定Wire错误映射；可安全关联的无效答案走ErrInvalidAnswer且不重传。业务语义不支持或缺必要原请求证据则给Application报告，不返回类型化的已授予/已扣费成功。未知M=0扩展的安全保留不能证明未知unit语义可用；也不能把组内任意未知扩展都猜成unit。

仍待逐行补齐：32.299对MSCC/USU的全部额外字段、tariff/time语义、池及最终单元动作、PS/IMS/节点/地址的递归叶。上述核心子集不是完整MSCC支持声明。

### 10.5 RO-MSCC-RESULT：关联与结果优先级

依据RFC4006 §5.1.2、§8.16及§8.28–30，固定下列纯解释行为；不新增配额账户、计时、业务终止或自动CCR-UPDATE引擎。

| 输入 / 分支 | 库必须保持的区别与动作 |
| --- | --- |
| 顶层Result-Code不为2001，但某MSCC为2001 | 顶层结果优先，不能被子组“救回成功”；原结果全部保留，不交付可用配额结论。4011仍保留“不再需要信用控制”的独立语义，不能把它降成普通拒绝或已扣费；见§29 |
| 顶层2001，某MSCC失败、缺结果或未知 | 逐服务解释；失败不变成功，缺/未知不能机械填2001。缺组内结果按已审核服务程序判断，缺依据就报告解释未完成；顶层成功不足以证明所有请求服务都获额度 |
| Service-Identifier与Rating-Group同时出现 | 单元目标仍是列出的服务；只有Rating-Group时才代表该组。关联还需原Service-Context-Id；不跨服务上下文混合相同数字。一个MSCC中的额度面向该目标集合，不复制为每个服务各自获同额；匹配规则见§30.1 |
| 相同服务键出现多个MSCC | 费率切换存在合法双额度；保留有序列表、原组位置及限定字段，不能直接按service/rating键去重或最后值覆盖。未支持该限定语义时报告不支持，不判为通用5009；超本地数量限额才是ResourceLimit |
| G-S-U-Pool-Reference指定一种CC-Unit-Type | 相应实际unit必须存在，例如TIME要求CC-Time；不能只凭Pool-ID生成零额度。不由库追踪跨消息池余额 |
| 原服务请求未带RSU，答案带新quota | 在§8.16对应流程中违规；核对原请求而非只校验CCA树。ECUR终止及其他动作仍按其具名流程，不能概括为所有EVENT均要求RSU |
| 接收不支持的unit/tariff语义 | 按RFC4006 §8.17/§8.20报告不可用CCA/应用失败；既定Wire/Exchange与Application分层仍适用。标准客户端后续的会话终止及DIAMETER_BAD_ANSWER由应用执行，公共库不自动发送终止请求或替应用修改计费会话 |

作为参考实现交叉检查，本地 [go-diameter信用控制测试字典](../../go-diameter/diam/dict/testdata/credit_control.xml) 将MSCC内USU、Service-Identifier和Pool-Reference声明为max=1，而RFC4006允许重复。这里仅记录该文件的差异，不推断所有版本/实际运行行为；本库的生成定义不能盲目复制。继续采用OTP“定义与使用位置分离”的思路；核心数值编码、Group cardinality与应用结果解释分别验收。

本轮新增FIELD-05/06、RESULT-05、DEP-03、VECTOR-01，并扩展PROVISION-01，见[覆盖基线 §16](diameterkit-usage-coverage.md)。R29/R34保持P1：已有具体候选样本不等于独立验证完成，5011标准配对也没有被本节解决。

## 11. 第十五轮：在线使用位置、节点/地址与MSCC条件

本轮将字段“有定义”“语法可出现”“某在线流程允许”分开审查。继续使用公开、无持久化端点边界；所有新增规则和成对样本仍待实现验证，不新增charging工作流。

### 11.1 固定依赖与审查范围

| 原件 | 实际取得 / SHA-256 | 本轮用途 |
| --- | --- | --- |
| [TS 32.251 V18.0.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132251/18.00.00_60/ts_132251v180000p.pdf) | PDF，1,167,606 bytes；`306fba1fe03829508f4ecbac2999d114f07044a60150cc8c5453368f0a077d38` | §6.3.1.2 PS信息中的设备字段；不导入PS承载、CDR或配额消耗引擎 |
| [TS 32.260 V18.3.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132260/18.03.00_60/ts_132260v180300p.pdf) | PDF，2,077,421 bytes；`b2c4c20595f28c58e3a87137718cc516d9a45f92de84cce27f5bffbb893a380a` | §6.3.1.1中PS信息/Subscriber Equipment Number的复用关系；不导入IMS计费程序 |

版本来自官方目录和PDF封面，32.260使用实际存在的18.3.0，不猜18.0.0地址。截至第十五轮累计20份PDF、Sh ZIP/XSD及3份RFC TXT有哈希；第十六轮增量见§12。另复读已归档32.274 §6.2.1/§6.3/§6.4、32.299 §6.4/§7、29.338 §5.3.3.6–7及其引用。32.274印刷页90的在线字段表已对照原PDF图像，不能把它与§6.3.1.3离线表或§6.2a收敛计费表混读。

### 11.2 RO-SMS-APPLICABILITY：在线使用规则先于可选性

§10.3的最短Service-Context-Id例外只使OM/OC可省，**不会使not-used、错误操作或错误父路径变为允许**。固定顺序：确定应用/服务/操作及父路径 → 应用已核对且无冲突的在线允许规则 → M/C或OM/OC出现条件 → 值/跨字段检查。通用ABNF不能自动覆盖在线用途约束；证据冲突按§12.2处理，不能仅凭表更具体就定对端违规。

| 使用位置 / 依据 | 确定行为与反例 |
| --- | --- |
| CCR/User-Equipment-Info，32.274表6.2.1.2.1.1的Subscriber Equipment Number、表6.3.3.1、32.299 CCR语法 | SMS在线I/T/E允许顶层设备信息，是否生成仍受OC条件及最短标识例外影响；未知设备不得用IMSI填Type0。此位置不因PS子组的禁用而一起删除 |
| CCR/Service-Information/PS-Information/User-Equipment-Info，32.274表6.3.1.1、表6.3.3.1及§6.4 | 通用信息表/映射确实列出该项，但在线表将嵌套User Equipment Info在SMS、T4 Device Triggering和MO-SMS T4均标“-”。所选标准SMS在线构造不生成该路径；接收可安全表示并报告应用不适用，不自动搬到顶层、不仅因通用定义存在就报完整通过 |
| SMS-Information/SMS-Result、SM-Sequence-Number、External-Identifier等，32.274表6.3.3.1 | 前两项存在CR与在线表/32.299字段表的冲突，构造与接收采用§12.2的未支持分支策略。External-Identifier在该表仅MO-SMS T4的T可用；不等于禁止OFA等其他接口的同名字段 |
| 顶层/嵌套设备编码，32.251 §6.3.1.2、32.260 §6.3.1.1、32.299 §7.1.17 | 父路径不是PS/IMS计费类别的充分判据：IMS也复用PS信息中的设备项。PS与IMS的IMEI填充差异见§10.2；SMS顶层Type0先采用§12.3真实16位IMEISV子集，仅IMEI映射仍未裁决。不能因为IP-SM-GW有IMS接入就自动选15位规则，也不能因出现PS父组就自动补F |

第十五轮曾暂以表6.3.3.1优先处理差异；第十六轮核对已批准CR0081后撤回这一笼统决定。`SPEC-RO-SMS-TABLE`的现行证据与逐字段行为见§12.2：冲突尚未消除，不能自动判违规、静默省略或伪造成功。

两条设备路径的用途检查仍有效；SMS顶层表示仅收敛§12.3有限子集，上述规范差异仍未解决。32.251原文还对Terminal Information替代User Equipment Info的在线用途保留编辑说明；不得顺势替换成Terminal-Information。标准线上不适用与具名旧对端兼容分别记录；兼容只能凭实际依据启用，不提供“忽略全部category”开关。

### 11.3 S6C-NODE-COMBINATION / S6C-NODE-NUMBER

29.338 §5.3.3.6–7在S6c使用2401/2406的节点组；每个具名成员0..1，扩展槽不意味着任意已知节点组合都合法。以下是其列明的组合，不是Ro/T4 Serving-Node的公共别名：

| Serving-Node允许组合 | Additional-Serving-Node | 解释和所需证据 |
| --- | --- | --- |
| SGSN-Number | 允许 | 只有号码时不生成Diameter host |
| SGSN-Name + SGSN-Realm + SGSN-Number | 允许 | 规范条件为HSS支持Gdd且已收到SGSN的Gdd指示；接收方检查组合，但不联网确认HSS历史。发送方负责其真实业务依据，不把额外事实变成公共Context |
| MME-Name + MME-Realm + MME-Number-for-MT-SMS | 允许 | MME名/域为DiameterIdentity，号码是TBCD；不能互填 |
| MSC-Number | 允许 | 号码本身不授权新建Diameter连接 |
| MSC-Number + MME-Name + MME-Realm | 允许 | 明确合法；不能用“一个组只能一种节点”或“出现MME就必须其号码”的粗规则拒绝 |
| IP-SM-GW-Number；或再加Name；或再加SBI-Support-Indicator；或Name与Indicator同时加入 | 不列入该组的具名成员/组合 | 不强制补Realm，不把Indicator当成要执行SBI；Additional组不能复用Serving的全体字段 |

IP-SM-GW-Realm在Serving语法中列出，但组合清单未明确其伴随方式；本轮登记为`S6C-NODE-REALM`待裁决，不能既声称完整关闭又私自把所有带Realm的输入判5008。固定标准helper先只生成明确组合；接收保留有界值并标明该条件未完成审查，阻塞该能力完整符合性声明。不存在这个歧义的明显残缺组合（例如仅MME-Name）仍可据明确规则拒绝/报告。扩展的M位和路径规则另外检查。

| 叶字段 / Vendor=10415 | 发送flags与编码 | 边界和反例 |
| --- | --- | --- |
| MSC-Number2403、SGSN-Number1489、MME-Number-for-MT-SMS1645、IP-SM-GW-Number3100 | 本S6c使用位置均V=1/M=1。29.173 §6.4.5、29.272 §7.3.102/159、29.336 §6.4.14：国际号码的TBCD数字串，非UTF8、非Diameter Address，不附TON/NPI | 低半字节先放数字；奇数位仅最后高半字节F补齐。E.164数字最多15位；不接受中间F、A–E；不能把前缀91当TON/NPI剥掉，它也可能是合法TBCD内容，不可据首字节自动判定调用方多加了前缀。仅做格式检查，不验证号码分配；最大payload8字节、W=20 |
| MME-Name2402、MME-Realm2408、SGSN-Name2409、SGSN-Realm2410 | DiameterIdentity；MME二项在S6c均M=1，SGSN名/域按引用定义M=0；V均1。29.338表5.3.3.1/2显式将MME-Realm、MME-Number提升为Must，不能抄29.173/29.272原表M位 | 字符/长度按BASE Identity规则和本地额度；不是号码或SIP URI。接收已理解AVP按APP-MBIT；准入仍按DRA绑定，不由这些数据直接授权路由 |

以上规则与§9.3的SMSMI命令条件独立。路由解释返回声明的节点/协议能力，应用再按允许的目的地选择路由；HSS答案里的节点名称不是DRA邻接Peer身份。

### 11.4 RO-ADDRESS-FAMILY：显式family，禁止按长度猜IP

RFC6733 §4.3.1的Address包含2字节网络序family及该family的值；不同于上述TBCD号码，也不同于MMS Address-Type/Address-Data组。候选公开值应分别保存family和owned payload，IP构造器显式生成family1/2；不能把两者压成一个长度驱动的byte slice后再猜类型。

| 使用位置 / 来源 | 发送/接收规则与预算 |
| --- | --- |
| 已审核允许IP的Address字段，如Client-Address2018/PDP-Address1227；32.299 §7.2.41及对应定义 | IPv4 family1后恰4字节、IPv6 family2后恰16；vendor AVP的W分别20/32。PDP用途按实际IP类型检查；Client/SMSC等字段其他family的允许范围不能仅凭它们都叫Address就互抄 |
| SMS-Information/Originator-SCCP-Address2008、Recipient-Info/Recipient-SCCP-Address2010；32.299 §7.2.128/170 | 本条款明确family8，地址信息UTF8编码；不能发裸TBCD、自动变IP或丢family。E.164数字实例只是一类输入，条款也提及Point Code/Global Title，复合表示未明确部分保持待裁决，不凭空制定分隔符或全面限为15字节 |
| 本地未知/尚未支持family | 可安全保留family与有界payload不等于实现其语义；应用未支持不能给出有效目标。畸形已知IP长度走字段错误；本地资源限额不是无效地址语法。Canonical保留family与有效payload，不把IPv4映射IPv6、UTF8、TBCD互转 |

本地[go-diameter Address实现](../../go-diameter/diam/datatype/address.go)对非IP解码保留含family的slice，但Serialize再按`net.IP(slice)`推断：family8加14个ASCII数字正好16字节，会被该代码路径重新视为IPv6。此结论来自固定本地源码检查，未运行其测试；新库以明确family避免该歧义。必须以非IP总长4/16及正常IP各做独立正反例，不能只测IPv4/IPv6 round-trip。

### 11.5 MSCC扩展：表示、方向和跨字段条件

此表补32.299的可复用规则，不表示标准SMS自动启用全部配额机制。具体流程仍先过§11.2和32.299表6.4.2.1/6.4.3.1。下列vendor字段均10415、V/M=1；U32/Enum/Time的W=16。错误、Canonical及本地限额沿用§9.2/§10，不新增计时或业务状态。

| RuleID / 父路径 | 类型、次数与条款 | 条件、错误反例及边界 |
| --- | --- | --- |
| RO-REPORT-PLACEMENT；CCR/MSCC/Reporting-Reason872及CCR/MSCC/USU/同字段 | Enum；MSCC为0..*、单USU为0..1；§7.2.175 | 报告usage时须带适用reason，不能仅因ABNF可选就省略；只随usage report发送，禁止顶层和CCA。0/3/5仅USU；1/2/4/6/7/9直接MSCC；8在USU，全部unit同池时可在MSCC。不可把所有合法枚举当所有父路径可用；是否真实触发/同池由应用提供依据，无需库读账本 |
| RO-THRESHOLD-UNIT；CCA/MSCC/Time-Quota-Threshold868、Volume-Quota-Threshold869、Unit-Quota-Threshold1226 | U32各0..1；§7.2.229/243/240 | 分别要求同组GSU有CC-Time、至少一种Octets、CC-Service-Specific-Units。额度是U64不意味着阈值也是U64；不能只因阈值高于额度就无依据判格式错。重新授权由应用执行 |
| RO-QUOTA-TIMERS；CCA/MSCC/Quota-Holding-Time871、Quota-Consumption-Time881 | U32各0..1；§7.2.160/159、§6.5.4 | 前者0表示不启用，缺值涉及客户端本地默认，库不代填；后者需时间quota。数字为秒，仍保留0/缺区别，不启动timer，不把缺业务配置报远端Missing-AVP |
| RO-TIME-MECHANISM；CCA/MSCC/Time-Quota-Mechanism1270 | Grouped 0..1；Type1271 Enum必需1、Base-Time-Interval1265 U32必需1；§7.2.228/230/29 | Type0/1；用于时间quota。与Quota-Consumption-Time共存时机制优先（§6.5.7），不是互斥字段错误；保留两原值。最小无扩展W=44，基准间隔0的业务可用性不靠猜测规定 |
| RO-ENVELOPE；CCR/MSCC/Envelope1266 | Grouped 0..*；Start1269 Time必需1、End1267 Time可选1，核心Octets及ServiceSpecificUnits各0..1；§7.2.59–62 | 尚未关闭的envelope缺End合法；Start即使早于本次区间也保留。多组须时间顺序，时间era缺失时不以uint32排序跨2036边界；依BASE-TIME显式参数检查，不增加通用Context。无扩展W为28–108，数量/深度受总预算 |
| RO-ENVELOPE-REPORT；CCA[INITIAL]/MSCC/Envelope-Reporting1268 | Enum0..1，值0–4；§7.2.61 | 首答缺失表示不报告individual envelopes；volume/event条件依已选报告模式。纯解释只返回指示，跨CCR的envelope累计与保留归应用；不新增会话仓库 |

Reporting-Reason的未知值按固定枚举策略报告，不能当FINAL或0；有些条件只涉及当前报文，可以Wire检查，有些依赖原请求/应用已持有的协议证据，不能用“通用Context不足”新增facts容器。需要历史状态才能证明的业务真实性不作为普通codec使用前提；支持声明须区分字段解释与完整CTF执行。

§13已补PS-Furnish和Service-Specific-Info直接叶及用途边界；剩余MSCC递归闭包仍有Trigger/Related-Trigger、AF-Correlation、QoS、Announcement及其他扩展的实际使用条件。单独字段核对不代表全部MSCC完成。

### 11.6 成对样本与本轮门槛

[候选样本 §6–7](diameterkit-protocol-vectors.md)新增完整DCCA IEC direct-debit CCR/CCA，以及S6c号码、非IP Address、Reporting-Reason和时间机制片段。成对样本明确采用测试用服务文档语义与RFC4006基线，服务标识不冒充`32274@3gpp.org`；它服务于已有DCCA/旧模板义务，不用于宣布标准SMS主流程符合性。

标准SMS pair按§12明确的子能力分别补齐条件与输入，§13已给ECUR INITIAL候选；不涉及未决字段的路径无需等待全部争议解决，也不能为制作“成功样本”省略实际输入触发的条件。新增FIELD-07/08、APP-01、DEP-04、PAIR-01，扩展FIELD-06及PROVISION-01。R29/R34保持P1；节点Realm歧义、SMS仅IMEI映射、冲突字段、SCCP复合表示、5011配对和剩余递归叶继续逐项登记，不以框架或业务存储解决。

## 12. 第十六轮：四项裁决的证据、可用子集与阻塞范围

本轮区分“库已决定如何处理”和“标准歧义已有权威解答”。下列处理策略可供实现与测试使用；没有把尚存歧义改称协议必错，也没有要求调用方接入在线facts、HSS查询或持久化服务。固定Rel-18基线不因对照后续版本自动升级。

### 12.1 新取得的原件及复核结果

| 取证对象 | bytes / SHA-256 | 结论与用途 |
| --- | --- | --- |
| [32.274 V18.1.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132274/18.01.00_60/ts_132274v180100p.pdf) | 802313 / `54ff72e7dfb27fa0f8a7bfcd22f005e8c03da6385a3a3f41056502d6648a1bb8` | 对照版本；§6.3.3.1在线SMS列仍将SMS result、SM Sequence Number、嵌套User Equipment Info标“-”；印刷页90已核对图像。变更历史记录CR0089为引用修正，不能据版本更新推断本问题解决 |
| [29.338 V19.3.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/19.03.00_60/ts_129338v190300p.pdf) | 382765 / `19ca6f17f1598e16ebc7f14a346c02db107a2ed9c1d2100256fc360c43144cef` | 对照§5.3.3.6：Realm仍在语法中、仍未列入IP-SM-GW组合。第十八轮纠正：18.2.0 §5.3.3.13已含不支持SMSF特性时的地址回填说明，不能称为19.3才新增；相关组合仍按固定版本单独审核 |
| [32.299 V19.0.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132299/19.00.00_60/ts_132299v190000p.pdf) | 1256608 / `31f71799ebeb0d0c449d0d09aa7073a4fd76fbca1ab3a4e1fdb9977e68b75d50` | 对照§7.1.17、§7.2.128/170：PS/IMS的设备表示分支、SCCP family8/UTF8及Point Code/Global Title描述仍在；未提供新的SMS IMEI映射或复合地址分隔格式 |
| [S5-213635，32.274 CR0081 rev1](https://www.3gpp.org/ftp/TSG_SA/WG5_TM/TSGS5_137e/Docs/S5-213635.zip) | ZIP 47284 / `1fe95f6eb91c534242e646a7522485779925005a2a0583ed4cc5b63a5839b0fb` | 已读取DOCX及修订标记；删除了SMS result和SM Sequence Number描述中的离线专用限制，不能只根据CR标题推断数值编码已明确 |
| 上项ZIP内DOCX原件 | `3ca08e459986e68ca0dd373310e08949771243a254f59f4d97f40337c2bb91d5` | 成员名为 `S5-213635 Rel-16 CR 32.274 Correcting SMS Result coding.docx`；删除文本单独核对，没有把删除内容当最终正文 |
| [S5-130729，32.299 CR0528](https://www.3gpp.org/ftp/TSG_SA/WG5_TM/TSGS5_88/Docs/S5-130729.zip) | ZIP 22307 / `9e5e75833bbd4fd3c2bcab4c21ca600181be8383febb8e09ae5d27ac31043df3` | 历史编码线索；已核对DOC正文片段中的UTF8十进制表示，不能用历史通用文字覆盖当前PS/IMS分支 |
| 上项ZIP内DOC原件 | `8dd7b81031f9ba5990bf84ba66a30ff156b91367ae489c02031e3497f12164c8` | 成员名为 `S5-130729 R12 CR 32299 Correction of User-Equipment-Info-Value encoding.doc`；用OLE正文片段复核，不声明完整修订版式审计 |

[3GPP CR登记页](https://portal.3gpp.org/ChangeRequests.aspx?q=1&specnumber=32.274)将0081 rev1列为SA5 agreed、SA#92批准并纳入16.3.0；现行18.0.0变更历史也包含它。这里的残余问题不是“CR未批准”，而是批准后的描述、在线矩阵及32.299字段表没有完全一致。截至第十六轮累计23份PDF（其中该轮3份仅作对照）、Sh ZIP/XSD、3份RFC文本、2份CR ZIP及2份成员，共32项原件哈希；目录/Portal检索不等于穷尽所有勘误或CR。

### 12.2 SPEC-RO-SMS-TABLE：撤回笼统的在线表优先

第十五轮“总以在线表更具体为由拒绝”不足以处理已批准CR，故只对无矛盾条目继续采用在线用途表。将以下两个字段单列，禁止从一行裁决推广到整张表或整个SMS Profile：

| 项目 | 交叉证据 | 本库当前处理 |
| --- | --- | --- |
| SMS-Result（10415/3409，U32，V/M=1） | 32.274 CR0081删除offline-only；§6.3.1.2仍是失败时C；普通在线SMS表标“-”；32.299表7.2.1将其列为ACR使用、CCR列“-”，但§7.2.212A描述一般SM提交/投递结果 | Wire类型/父组可表示；标准SMS语义对受此冲突影响的输入返回未支持/未裁决报告，不能仅据某张表声称“已证明对端违规”。标准helper暂不自动生成该失败事件分支；不能自动丢掉SMS-Result后生成成功，也不擅选值或塞入CCA Result-Code |
| SM-Sequence-Number（10415/3408，U32，V/M=1） | 同CR亦删除该字段的offline-only描述；32.274在线表仍“-”，32.299 §7.2.208A与字段表仍明确限离线 | 采用与上述相同的未裁决处理；不默许在线发送，也不声称CR已经修正32.299。串接短信本身可由应用处理，但库不能靠把序号改放TPDU或另一个字段来绕过冲突 |

两项保留在固定定义中，不因使用语义未决删掉AVP；同码合法宽度不证明某条消息完全符合标准。Pure builder在接纳前返回有界本地“不支持该构造分支”错误并指向该裁决ID，区别于值/长度错误。通用接收先完成Wire/Exchange，随后在Application报告中标记该具体规则未解决；没有未支持字段和条件的其他合法分支不受影响。

未知的SM交易结果不是本库失败码。应用向纯helper提供已知的当前事件用途/状态即可，库不读取投递历史；如果当前操作要报告的字段受上述冲突影响，就明确限制该helper能力，不能要求“再给一个Context”来消除标准矛盾。安全收到的原字段仍可供应用诊断，未支持状态不触发自动重传、退款或第二次业务交付。

### 12.3 RO-UE-IMEISV16：先支持真实16位设备身份

32.299 §7.1.17中，真实16位十进制IMEISV在PS的UTF8十六进制文本表示与IMS的UTF8十进制文本表示中具有相同字节。历史CR0528也提供UTF8文本的依据。本库将**调用方明确提供的真实IMEISV、恰16个ASCII数字**作为SMS顶层设备helper的首个有限子集；这是基于共同表示作出的库支持选择，不宣称标准为SMS单独规定了所有编码分支。

| 输入/动作 | 构造与接收决定 |
| --- | --- |
| 顶层User-Equipment-Info，Type=0，明确的16位IMEISV | 按原数字UTF8写Value，保留前导零；不做Luhn校验、不验证TAC归属。Type/Value均必需；无vendor头，所选标准发送M=0（表7.1允许M），无扩展总W=44 |
| 只有15位IMEI，或15位加F的PS形式 | 该SMS helper不自动补SVN、补F、去F或截断；报告该输入转换尚未支持。泛OctetString仍可表示，接收不把它伪称非法整数宽度；SMS专用解释保持未支持。已有PS/IMS独立已审核用法不因此被删除 |
| 输入IMSI、8字节TBCD或任意opaque值 | 不自动视为IMEISV；形状与类型分别检查。16数字也不证明真实设备归属，真实来源由调用方负责 |
| 设备信息缺失 | 在当前路径/出现条件允许时省略；不能为满足默认FieldSelection造假值，不能在明确给了未支持IMEI后静默当作缺失成功 |
| 设备位于PS-Information嵌套路径 | 仍独立执行§11.2在线位置约束；数值属于16位子集也不能搬移或放开错误位置 |

decode保留Type、原Value、Presence；纯解释必须说明适用的子集。接收默认owned语义及MaxFieldBytes不变，不引入设备编码探测器或每条消息的动态策略。真实IMEISV16不再等待“SMS所有设备分支统一解决”才可做原型；仅IMEI映射仍为单独P1。

### 12.4 S6C-NODE-REALM：可安全接收，不猜规范组合或路由

已核对18.2.0与19.3.0的语法/组合差异，本轮未找到将Realm加入组合的权威修正；不能将别的接口对Realm的规则直接移用。当前固定行为如下：

- 无Realm且满足§11.3明确组合：正常检查，不要求应用补Realm，不连带阻塞MME/SGSN等节点分支。
- 已知Realm存在且本身是合法DiameterIdentity：Wire保留该字段，应用组合报告`S6C-NODE-REALM`未裁决；不能仅因此报未知M=1 AVP、5005或5008。薄路由helper不输出“全部通过”的目标，原始节点数据与Report仍可取得。
- Realm自身格式/次数错误、Number缺失等与歧义无关的已证明违规：照既定规则处理，不能用未裁决状态掩盖错误。若带Realm且把它去掉后是明确组合，仍不是Realm组合已经标准验证的证明。
- 标准构造只提供无歧义组合；若调用方明确要求带Realm的该分支，本地报告未支持，不静默去掉字段。具名兼容须补独立报文/对端要求、精确组合与验收后才能启用；第十六轮不创建无证据的宽松Profile；第二十二轮§18.3补有限具名三元组的设计，运行/部署证据门槛仍保留。

报告中的“不知道标准组合”与“缺调用方协议输入”不同。补充validation.Context、用DRA CER中的Realm或从Name后缀猜Realm都不能关闭该问题。该报告不会保持网络pending等待业务查询，也不授权按Realm新建连接。

### 12.5 RO-SCCP-GT：明确E.164 Global Title子集

32.299 §7.2.128/170明确family8、UTF8；同段提及Point Code/Global Title但未给复合分隔规则，19.0.0仍相同。确定公开纯helper先覆盖**显式E.164 Global Title输入**：1..15个ASCII数字，family=8、内容UTF8，不含加号、空格、TON/NPI或TBCD；这些字符限制是该数字构造入口的契约，不是对所有family8 opaque输入作新的全局规定。vendor地址AVP的W为16..32。

构造者明确选择GT语义，库不猜`1234`是点码还是号码。通用接收保存family与owned payload；只有调用方显式选择GT解释，或固定具名子能力已约定GT语义时，才按该数字子集解释。该解释不证明号码归属，也不把SCCP地址变为可直接拨号的Diameter Peer。

Point-Code-only、Point Code+GT、含分隔符的字符串继续保持有界原值和未支持语义，不提供自动拼接/拆解helper，不将CDR的AddressString/ASN.1编码搬到Diameter Address。已审核的GT子集不受复合格式P1阻塞；相反，复合格式也不能因raw能往返而被标为符合性通过。

### 12.6 分层失败、发布门槛与参考库

| 状态 | 运行时/公开行为 | 阻塞范围 |
| --- | --- | --- |
| 证明了线路类型/边界违规 | 原Wire错误映射；无效且可关联答案ErrInvalidAnswer，不回答案、不重传 | 该输入 |
| 有报文但标准规则未解决/实现语义未支持 | 有界Application报告含RuleID、路径、已检查范围及原值；不能称ContextRequired或完整成功。Wire/Exchange已通过的答案按既定生命周期结束网络等待 | 该字段、组合或helper分支；不是所有SMS/所有传输 |
| 上下文确实由已定义接口提供但缺失 | 保留既有ContextRequired语义；不能用任意值覆盖规范分歧 | 依赖该输入的检查 |
| 已确定表示的有限子集 | 按明确构造/接收规则实施和验收，仍须独立fixture/运行测试后才可发布为已验证 | 子集自身；不代替整个Profile或v1 A/B义务 |

这不新建运行时报告类型、facts或规则插件；用现有Report的分类细化原因，WireFingerprint/InboundSemanticsFingerprint随接受与解释规则变更，不能只改文案。当前24个SPEC需求保持，新增DECISION-01、FIELD-09、DEP-05并扩展APP-01/FIELD-07/08。

参考[OTP字典接口](https://www.erlang.org/doc/apps/diameter/diameter_dict.html)及本地[go-diameter Ro/Rf字典](../../go-diameter/diam/dict/testdata/tgpp_ro_rf.xml)：字段定义可复用，但字典的类型/次数不裁决这四项程序语义。另查阅[Mobius Accounting AVPs说明](https://www.mobius-software.com/documentation/Mobius%2BDiameter/Accounting%2BAVPs)，其SCCP项也描述family8/UTF8及点码/GT；所查条目没有给出可据以冻结本库复合格式的线路实例，因此不将第三方描述提升为标准补丁。

## 13. 第十七轮：实际字段依赖闭合与标准主流程pair

本轮补齐下面明确使用的编码子集及引用链，保留字段未支持与完整Profile未完成的区别。新文档样本包含SMS ECUR INITIAL、S6c SRR、SGd MO/MT的完整请求和成功/失败答案；不是现网抓包，也未运行协议实现。

### 13.1 固定原件与采用范围

| 原件 | bytes / SHA-256 | 本轮实际核对范围 |
| --- | --- | --- |
| [TS 29.061 V18.1.0](https://www.etsi.org/deliver/etsi_ts/129000_129099/129061/18.01.00_60/ts_129061v180100p.pdf) | 1191674 / `21e9d89331d74b8bdd01695edd286aee4410ffdd4f01c45e2062c75757fb930c` | §16.4.7.2、§16a.5表9a：RADIUS派生值、Diameter类型/flags；不导入Gi/Sgi程序 |
| [TS 29.274 V18.6.0](https://www.etsi.org/deliver/etsi_ts/129200_129299/129274/18.06.00_60/ts_129274v180600p.pdf) | 2790366 / `c9eefb4832b1549b7fdb9d3abdf9bf7afe4f7da611942ae6e3149e0a807d3be2` | §8.17、8.21.4–5、8.44：RAT、TAI/ECGI与时区值；不导入GTP头/会话 |
| [TS 24.008 V18.5.0](https://www.etsi.org/deliver/etsi_ts/124000_124099/124008/18.05.00_60/ts_124008v180500p.pdf) | 4695712 / `16e7826226a4f27b6c15ab2463e8e4c3516ebcc1d722975899f4ba80dfc2ab96` | §10.5.3.8→已取23.040 §9.2.3.11：时区半字节与符号 |
| [TS 29.118 V18.0.0](https://www.etsi.org/deliver/etsi_ts/129100_129199/129118/18.00.00_60/ts_129118v180000p.pdf) | 564409 / `9a3fa9dffb39877fd516afe004d8be2f23f70fa5fddb5ac794fe6c66c897d341` | §9.4.3a、9.4.21a：经29.002引用的ECGI/TAI值，不带SGs TLV |
| [TS 24.301 V18.6.0](https://www.etsi.org/deliver/etsi_ts/124300_124399/124301/18.06.00_60/ts_124301v180600p.pdf) | 3947930 / `2d0c90b4f9e64b97dda298064306ddde4d1653d108af4cc9a4eb3e8e6cca7c79` | §9.9.3.32：TAI五字节、两/三位MNC、删除值；不导入NAS状态机 |
| [TS 23.038 V18.0.0](https://www.etsi.org/deliver/etsi_ts/123000_123099/123038/18.00.00_60/ts_123038v180000p.pdf) | 300034 / `c0eacd73a76acc1b37f757d013a75117dd11b70bc5d1ce997d1412bc244fd047` | §4、6.2.2：样本DCS=04表示8-bit用户数据；不是新增TPDU codec |
| [OMA MMS Encapsulation Approved 1.3 / 2011-09-13](https://www.openmobilealliance.org/release/MMS/V1_3-20110913-A/OMA-TS-MMS_ENC-V1_3-20110913-A.pdf) | 898789 / `8dd6b13d459dc56345b55019e74507b1351967181574f2d3067c3c28fa7280be` | §8：32.299所引MMS地址表示；仅选择电话号码文本子集 |
| [RFC 7155](https://www.rfc-editor.org/rfc/rfc7155.txt) | 154559 / `f100a5a47def22bda012369e8926f5d17c57b9a73a0fe4daeac16bb01b272cd5` | §1.2：删除旧RADIUS互操作章节但保留派生值格式；不能误称该RFC要求实现转换网关 |

版本经原件封面确认；这些是固定依赖选择，不宣称最新。累计30份PDF、Sh ZIP/XSD、4份RFC文本及4份CR材料，共40项原件哈希。原有32.299/32.274/29.338基线不升级。下面的依赖闭合指选定字段的编码依据可追溯，不等于全部标准、勘误或实现验收关闭。

### 13.2 29.061派生值：父路径、类型与封装必须分开

32.299表7.2.1把若干定义转引29.061。这里选其Gi/Sgi派生定义（表9a，Vendor10415、V=1/M=0，发送P=0），不混用同文Gmb表的M位。RADIUS的Type/Length及外层Vendor-Specific头、GTP IE头都不进入Diameter payload；Length仍按Diameter头计算。未知值保留原始Octets并报告支持范围，不把未知值默认成E-UTRAN或本地位置。以下W均含AVP头/padding。

| RuleID / 完整使用位置 | 键、次数、值规则及来源 | 错误、比较与预算 |
| --- | --- | --- |
| RO-CHARGING-ID；CCA/MSCC/PS-Furnish-Charging-Information/3GPP-Charging-Id；PS-Information同叶另按其使用条件 | 10415/2，OctetString；Furnish内1，PS内0..1；恰4字节，按无符号32位网络序解释；29.061 §16.4.7.2项目2。不能因数值可转U32就改字典类型，PDN-Connection-Charging-ID2050另有定义 | W=16；3/5字节长度违规，0与缺失不同；保留原bytes，不作为库发号器/账本键 |
| RO-RAT-OCTET；CCR/Service-Information/PS-Information/3GPP-RAT-Type | 10415/21，OctetString 0..1，恰1字节；29.061项目21→29.274 §8.17。首子集采用0–9范围中已定义的1–9，6为E-UTRAN；0保留为reserved，不宣称可用接入。29.061自己的10–50为spare，不能照抄29.274的10=NR；51–58须引用规范显式采用 | W=16、Length13；4字节Enum/ASCII“6”不等同单字节06。保留未知原值并报告未支持，Canonical不归并；MSCC直接同字段仅在其NBIFOM条件下使用，SMS不能随意复制到MSCC |
| RO-ULI-LTE；CCR/Service-Information/PS-Information/3GPP-User-Location-Info | 10415/22，OctetString 0..1；29.061项目22→29.274 §8.21.4–5。首个常用子集为type128+TAI(5字节)、type129+ECGI(7字节)，payload分别6/8；PLMN三字节，TAC两字节大端，ECI28位，高4位spare发送0。两位MNC用指定F，不按普通TBCD奇数补F处理 | W均20；已知tag长度不足/过长按值编码错误；未知tag安全保留但不猜布局。多个PLMN表示/原spare位不经Canonical重写；不带RADIUS/GTP头，禁止把5字节TAI本身当完整Ro ULI |
| RO-MS-TIMEZONE；同PS父组/3GPP-MS-TimeZone | 10415/23，OctetString 0..1、2字节；29.061项目23→29.274 §8.44→24.008 §10.5.3.8→23.040 §9.2.3.11。第一字节为15分钟单位的半字节表示，bit3是符号；第二字节低2位DST=0/1/2，3为spare；发送高6位0，接收忽略其语义但保存原字节 | W=16、Length14；2300表示UTC+8且无DST，不是有符号小时整数。时区值已含DST调整，不再次加小时；不从系统时区/当前时间补值。未知DST不猜本地规则；字节长度错与语义未支持分开 |

PLMN/TAC/ECI可纯解析，不证明网络分配或位置新鲜度。正常构造子集要求十进制MCC/MNC和对应spare规则；24.301明确保留异常/已删除TAI的表示，收到非正常PLMN或TAC=0000/FFFE时不能一概判Diameter帧坏，也不能交付可用位置结论。这里不引入网络查询、位置缓存或规范之外的号码正规化。Ro ULI的130及131–139组合、GGSN旧CGI/SAI/RAI与NR的额外采用条件仍单列P1，不以本轮LTE两种形式代替全部位置支持。

### 13.3 SGd的EPS位置链与Ro不能互换

SGd OFR/TFA → 29.338 §6.3.3.1 → 29.272 §7.3.111/115/117/118 → 29.002类型E-UTRAN-CGI/TA-Id → 29.118 §9.4.3a/9.4.21a → 29.274 §8.21.5与24.301 §9.9.3.32。这条链已核对到值布局；29.274的TAI布局也作交叉核对，不能跳过实际规范引用。

| RuleID / 路径 | 键 / flags / 次数 | 已支持表示与边界 |
| --- | --- | --- |
| SGD-EPS-LTE；OFR或TFA/EPS-Location-Information/MME-Location-Information | 10415/1496→1600；两层均V=1/M=0、各0..1 | MME内1602 ECGI、1603 TAI各0..1且V=1/M=0；payload为7/5字节，不含Ro的128/129 tag、ASN.1 TLV、SGs IEI/Length或NAS IEI。每叶W=20，两个叶齐备的两层组W=64 |
| SGD-EPS-LTE-VALUE；上述两个叶 | OctetString；29.002尺寸分别7/5 | PLMN前导零与两位MNC F保留；正常ECI高4位发送0、28位数值；TAC大端。已知叶错误宽度独立失败。异常/删除身份与正常位置分开解释，不从节点名称推导PLMN |

OFR位置表示发起短信的服务小区；TFA另按MT结果/出现条件使用。该值是调用者或对端提供的协议数据，不证明实时位置，不要求Current-Location-Retrieved或Age由库编造。Geo/Geodetic、CSG、eNodeB、SGSN分支仍有独立依赖；本轮不把只含ECGI/TAI的样本宣称全EPS组完成。相同原始TAI/ECGI叶可复用纯值函数，但Ro tagged值、SGd的AVP父路径/flags/条件必须分开。

### 13.4 主流程深层组的实际支持范围

复读六模板及SMS标准主流程后，区分：现有用法确实涉及接口/收发地址/单位；PS位置为标准SMS可选且与SGd共享值依赖；AF/QoS等仅因通用MSCC引用而存在，不因用户场景包含IP-SM-GW就强行导入全IMS/PCC。按第十九轮分期保留A/B义务，Ro现于首版实施；未在模板发现某项不等于以后禁止该项。

| RuleID / 路径 | 直接叶闭合与采用规则 | 预算/反例/边界 |
| --- | --- | --- |
| RO-SMS-INTERFACE；CCR/Service-Information/SMS-Information/Originator-Interface2009；Recipient-Info2026/Destination-Interface2002 | 各0..1 Grouped；其2003 Id、2004 Port、2005 Text均UTF8，2006 Type为Enum0..4，各0..1；32.299 §7.2.71–74/126/168，所有键10415、V/M=1 | 文本受MaxFieldBytes，Type W16；仅Type的组W28。Port是文本，不能与3010 U32 Application-Port混用。Destination-Interface不得直接放SMS-Information，重复Recipient-Info保持顺序及逐接收者关联 |
| RO-MMS-PHONE；Service-Information/MMS-Information/Originator-Address886或Recipient-Address1201；SMS-Information/Recipient-Info/Recipient-Address亦按其父路径使用 | Grouped；899 Type Enum与897 Data UTF8各0..1；本库电话号码构造子集两者都给，Type=1，Data选OMA §8完整PLMN文本，例`+12025550123/TYPE=PLMN`。这是32.299 §7.2.7引用的表示子集，不把Diameter Address family8或TBCD搬入 | 899 W16；897按UTF8字节算W；示例组W64。号码前导零、加号、分隔符和后缀不在重复比较时正规化。旧模板纯数字Data保留为待核对兼容表示，不静默加/剥后缀；接收不能仅因未覆盖的地址形式判Wire错误 |
| RO-ADDRESS-DOMAIN；上述地址组/898 | Grouped0..1，至少一项；Domain-Name1200 UTF8?、3GPP-IMSI-MCC-MNC8 UTF8?；前者10415 V/M=1并按§7.2.55承载FQDN，后者10415 V=1/M=0。后者按29.061项目8为MCC3+MNC2或3数字，不补0统一长度 | 列表/字符串受总预算；空组违反已知条件；把UTF8字段改成DiameterIdentity或域名解析不是本规范授权。其他email/IPv6/短码与IMSI地址类型按能力逐项验证，不声称全OMA地址解析完成 |
| RO-FURNISH；CCA/MSCC/PS-Furnish-Charging-Information865 | Grouped0..1，Charging-Id2必需1、PS-Free-Format-Data866 Octets必需1、PS-Append-Free-Format-Data867 Enum0..1；32.299 §7.2.155–157。865/866/867 V/M=1，2沿§13.2为M=0；867为0 Append或1 Overwrite，缺省语义为Overwrite | 无扩展W=40+4×ceil(n/4)，有867再加16，n为opaque数据字节；空值与缺字段不同。只返回指示，不写CDR/Rf、不建立free-format存储，不把CCA数据自动复制回CCR；标准SMS未选此能力，不插入SMS pair |
| RO-SERVICE-SPECIFIC；CCR/MSCC/Service-Specific-Info1249及各已允许父路径 | Grouped可重复；Data863 UTF8?、Type1257 U32?，V/M=1；32.299 §7.1.9、7.2.194–196及表6.4.2.1。语法未把两叶设为必需，不无依据补缺Type错误；是否有可用含义由明确服务定义决定 | 每组W12+可选叶W；重复顺序保留，不按Type覆盖。Data受MaxFieldBytes/总预算；不发明通用业务脚本或把任意文本当SMS专属语义 |

以上父组发生时，缺成员、次数、类型、已知值格式按既有Wire/Application分层；不支持的业务解释仍按§12.6报告。固定类型已知不代表服务用途已启用，FieldSelection不放开not-used。Canonical保留原字段与同键顺序，既有spare处理只按各规则指纹执行，不以本轮纯解释器改写入站比较字节。原值和解释副本均计入DecodeBound，不能只以W替代Go内存账本。

AF-Correlation/Flows→29.214/29.212、QoS/ARP→29.212、Trigger/Related-Trigger、完整IMS SIP字段、地理/测地/CSG及完整tariff仍是相应扩展能力P1；它们未在本轮有限pair中启用。已找到编码依据的四组与尚未审核的扩展必须分别记录，不能再次笼统写“29.061尚未取得”或“所有深层已补齐”。

### 13.5 成对报文的边界与结果

候选全文见[协议样本 §9–10](diameterkit-protocol-vectors.md)。每组分别给完整请求、成功答案及失败答案，保留同一请求的Session/HBH/E2E；失败答案是替代场景，不是允许同一exchange完成两次。

| 主流程 | 本轮选定子集 | 解释边界 |
| --- | --- | --- |
| 标准SMS Ro | `18.32274@3gpp.org`、ECUR INITIAL、MSCC/RSU→GSU事件单位、顶层真实IMEISV16、SMS接口/显式GT、PS的RAT6/TAI/UTC+8 | 适用32.274/32.299而非通用DCCA顶层单位。成功是初始授权；TERMINATION/用量/实际扣费完成归后续应用流程。本轮完整pair不含未裁决SMS-Result/Sequence，也不宣称覆盖IEC失败或refund |
| S6c | 有MSISDN、意图投递的SRR；SRA给IMSI与完整MME Name/Realm/Number；失败实验5001 | 原SRR未设NotIntended，不可省Serving-Node；IMSI不是HSS-ID。MME数据只供应用选择目的，DRA/目标准入仍独立，不自动发TFR |
| SGd MO | 已知IMSI/MSISDN的普通OFR、选定EPS LTE位置、14字节SMS-SUBMIT；成功OFA；失败实验5555+SC-CONGESTION4 | 无SMSMI的普通路径，不偷用MSISDN-less/T4；OFA成功不表示最终接收UE已收到 |
| SGd MT | 向MME发TFR，含MME号码与计时输入、20字节SMS-DELIVER；成功TFA；失败实验5555+MEMORY_CAPACITY_EXCEEDED0 | 发送/结果按原请求与角色检查；计时值/失败指示不启动业务重投递，SM-RP-UI由库作为有界opaque保留 |

所有pair的DRA是授权邻接而非成功答案Origin；本例不带Proxy-Info，代理若添加必须另按规则回显，不能把本例无代理字段当DRA禁止代理的要求。无CER/CEA、SCTP握手/stream/PPID或安全传输实测证据；这些均属于另行验收。TPDU的DCS/字段仅为样本人工校对，不增加库内短信编码/解析产品。

### 13.6 参考实现与剩余优先级

本地[go-diameter Ro字典](../../go-diameter/diam/dict/testdata/tgpp_ro_rf.xml)将TGPP-Charging-Id、TGPP-RAT-Type、TGPP-MS-TimeZone和TGPP-User-Location-Info均列为OctetString且M=0，与本轮类型选择相符；这类字典没有证明一字节RAT、tag相关位置长度、时区符号、使用条件或实际互操作已校验。延续[OTP字典接口](https://www.erlang.org/doc/apps/diameter/diameter_dict.html)的定义/使用位置分离，不把它当程序语义的替代依据。

第十七轮列出的TERMINATION/IEC、RDR/ALR候选与六模板源码推导已由§14补充；第十九轮将这些P1按版本区分：三应用分支继续推进，当时后置的Ro争议与计费兼容证据现纳入首版对应门槛。SMS冲突字段、仅IMEI、IP-SM-GW-Realm、SCCP复合表示、5011保持单项证据门槛。新增DEP-06、FIELD-10/11/12及PAIR-02，全部运行状态Planned；本轮不新增公开Context、持久化、位置服务或业务恢复接口。

## 14. 第十八轮：终止、单次计费、报告提醒与六模板证据

固定Profile仍为32.299 V18.0.0、32.274 V18.0.0、29.338 V18.2.0；沿用40项规范/附件哈希，不更换Release。下面补已选子集的规则和完整候选，另外固定14项本地源码/模板证据。源码推导、独立执行和部署验证分别记账；库仍为公开、无持久化的Go库。

### 14.1 Ro：原请求关联和不同动作的结果不能混用（首版）

依据：[TS32.299](https://www.etsi.org/deliver/etsi_ts/132200_132299/132299/18.00.00_60/ts_132299v180000p.pdf) §6.3.3/6.3.4、7.1.9、7.2.171/175，[TS32.274](https://www.etsi.org/deliver/etsi_ts/132200_132299/132274/18.00.00_60/ts_132274v180000p.pdf) §6.2.1/6.3，[RFC4006](https://www.rfc-editor.org/rfc/rfc4006.html) §5.4/6.3/6.4/8.3。未审的8506差异不借此关闭。

| RuleID / 路径 | 发送与接收、关联及结果 | 比较/预算/边界 |
| --- | --- | --- |
| RO-ECUR-FINAL；CCR/CC-Type、Number、MSCC/USU | 本例INITIAL=1/0后无UPDATE，TERMINATION=3/1；Session与OCS不变，Service-Context/服务键沿用，HBH/E2E属于新exchange。已发生UPDATE时编号由应用按实际请求序列提供，不硬编码所有TERM=1；一次原身份重传不递增编号。报告原授予的每种单位，本事件子集USU446/CC-SSU417 U64；不用RSU再次申请 | USU0不等于Absent；FINAL872 Enum=2必须直接MSCC，不能在USU/顶层；总量由应用计量，库不计算扣费。每USU仅事件单位W24，MSCC=头8+USU24+Service-ID12+Reason16=60；重复/未知叶继续计DecodeBound |
| RO-TERM-RESULT；CCA对原TERMINATION | 基础2001不要求GSU，允许没有MSCC；可选Cost/Balance按已采用字段检查，不当新额度。示例5002 UNKNOWN_SESSION_ID是基础永久失败、E=0；有效失败结束exchange，无自动INITIAL、STR或OCS迁移 | Session/CC-Type/Number和来源检查不因无GSU省略。接收其他未知结果保留并报告，不因不存在专用样本改判成功 |
| RO-IEC-DEBIT；CCR EVENT及CCA/MSCC | Type4/Number0、Requested-Action436=0；可给RSU非货币事件单位。成功CCA按32.299直接扣费过程返回GSU，本例放MSCC并匹配Service-ID；顶层/服务级结果均检查。4012失败可无MSCC/GSU；成功缺GSU、显式0、服务级失败各独立解释 | 不把IEC授权与ECUR预留共用一个“扣账已提交”结论；源业务允许服务前或后计费，本例仅服务前、无失败SM事件。SMS额度仍在MSCC，不借通用RFC语法启用顶层金额 |
| RO-IEC-REFUND；CCA/CCR的MSCC/Refund-Information | Vendor10415/Code2022 OctetString0..1、V/M=1；原IEC CCA若给token，退款CCR须原样包含；明确未给、已给空值和原答未知分别处理（§30.2），不跨MSCC借token。本例关联同OCS/Service-Context/服务键，由应用显式给原结果和退款意图；新的EVENT/Action1/Number0、新exchange/Session，不能重传已完成IEC。32.299退款CCA不要求GSU；可选RSU不凭空制造 | W=12+4×ceil(n/4)，n由字段及总预算限制；空字节值和缺AVP不同，不能用strlen或解码为文字。库不持久保存token、不验证退款资格、不读取账本、不执行补偿 |
| RO-SMS-REPORT-SCOPE；SMS-Information/SM-Status2014、Discharge-Time2012、SMS-Result3409 | 32.274的SM-Status适用delivery-report或ECUR；Discharge-Time仅delivery-report。二者OC，不能从Type3自动推出必填；其单位/值定义沿用已有规则。refund涉及失败事件时仍触发§12.2的SMS-Result规范冲突 | 本轮refund两帧仅为32.299线路/关联结构候选，完整SMS Application预期未支持；不以删3409构成合法正例，也不把冲突交给泛化Context解决 |

**IEC的MSCC采用说明。** 32.274表6.2.1.2.2.1的Multiple Unit Operation说明提到“IEC退款及ECUR”，没有把普通IEC标成not-used；§6.2.1.0又明确Debit/Reserve采用32.299及其字段描述。这里依据32.299 §6.3.3的直接扣费GSU要求、§7.1.9及SMS顶层单位限制，将非货币GSU放MSCC，显式记录此采用解释，不把一句说明扩为禁止普通IEC。独立规范复核/对端测试仍需覆盖该选择；与§12.2确有相反表格/CR的SMS-Result争议分开。

失败应答只给当前exchange的协议结果；5002/4012不授权任何业务重试。TERMINATION对先前INITIAL的业务关联来自调用方消息，不要求Node保留已完成INITIAL快照。refund原凭据同样由应用显式传入有界值；纯helper不查询历史、没有Store/Resolve或新的facts类型。原身份重传只操作运行中exchange，不能借本轮成对示例恢复已完成/跨进程请求。应用CTF在适用ASR过程中另有§30.4的终止报告义务；库不自动发送不等于该义务被撤销。

### 14.2 S6c RDR/RDA与ALR/ALA的字段及结果闭合

依据：[TS29.338 V18.2.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf) §5.2.2/5.2.3、5.3.2.5–8、5.3.3.14–23、7.3；User-Identifier沿29.336。选定有MSISDN、MME结果和HSS发起的普通路径；T4/IMS无MSISDN、MME/SGSN发起提醒等保留独立条件，不由缺字段自动切换。

| RuleID / 路径 | 键、flags、次数及已收敛条件 | 失败/比较/预算 |
| --- | --- | --- |
| S6C-RDR-ENVELOPE；RDR8388649/App16777312 | Session首AVP，Auth-Session-State1，Origin Host/Realm及Destination-Realm必需，Destination-Host可选（样本显式HSS）；User-Identifier3102、SC-Address3300、SM-Delivery-Outcome3316各1。请求和正常答案P=1/E=0 | 基础结果与Experimental分开；新命令不是SGd TFR的后半帧。DRA邻接不替代HSS目的或答案Origin |
| S6C-RDR-IDENTITY；User-Identifier、SMSMI | 普通路径有MSISDN则使用；无MSISDN且T4才使用IMSI；IMS无MSISDN重试为User-Name内HSS-ID，SMSMI仅目的URI，禁止origin URI/HSS-ID子叶。沿用SMSMI五请求规则 | 身份类型与原意图关联，保留TBCD/文本前导零，不用请求名猜业务模式；不新增HSS查询 |
| S6C-RDR-CAUSE；3316/节点outcome/3321、3322 | 3316 Grouped1；MME3317、MSC3318、SGSN3319、IP-SM-GW3320各0..1，另有SMSF分支不在本例。上述四组和Cause3321、Diagnostic3322均10415、V/M=1。3321 Enum：0容量不足、1用户不在、2成功；3322 U32。节点子叶语法各0..1，但整个RDR至少一个cause，MSC与MME cause不能同时存在；AbsentUser时关联已知诊断 | 不能用全可选字典接受空outcome；未知cause报告语义未支持。一个MME/cause组总W40，有诊断加16；叶与未知扩展计总预算。接收保留原顺序/原bits，不合并不同节点 |
| S6C-RDR-FLAGS；3323 | U32 0..1、V=1/M=0；bit0 Single-Attempt-Delivery，未定义位发送0、接收不解释其语义且保留原值 | 标志影响HSS业务处理，不让库写MWD或限制应用队列；Canonical按既有字段规则，不能擅自删原bytes |
| S6C-RDA-RESULT；RDA | 正常成功基础2001；适用实验10415/5001 UnknownUser、5558 MWD_LIST_FULL。成功时Alert MSISDN与请求号不同则User-Identifier/MSISDN必含；相同可省，未知用户/列表满失败不强制成功字段 | 不将10415/5001当基础5001；不塞SGd失败原因组。2001不是库已存MWD或保证未来ALR，返回号只是协议结果 |
| S6C-ALR-HSS；ALR8388648/App16777312 | 同类必需包络、SC3300/User-Identifier3102各1。Alert MSISDN存在用该值；T4无号用IMSI、IMS无号重试用dummy MSISDN且SMSMI只目的URI。3329 Maximum-UE-Availability-Time为Time 0..1、V=1/M=0；若可用则含。普通HSS样本不含3333 Alert-Event或Serving-Node | Time线路4字节、W16；明确era和调用方提供时间，不由库读实时可达性/安排提醒。MME/SGSN通知的3333/Serving-Node组合须单独审核，不因为ABNF可选就覆盖所有ALR |
| S6C-ALA-RESULT；ALA | 正常2001；§5.2.2.1明确没有为本过程定义S6c应用错误。不创造5001/5555映射；基础协议错误仍合法。E3004样本按RFC6733 §7.2通用错误格式，不能强加正常ALA的Auth-Session-State | E位先选语法，结果命名空间再解释；未知实验结果可安全保留但不作为本已支持结果分支。收到答案不回答案、不自动重试 |

ALR不是RDR的最终应答：两者有各自Session/HBH/E2E/pending，RDA完成即结束其exchange。HSS管理MWD、判断UE恢复、选择提醒时间/次数以及服务中心重投递归应用；本库只构造/校验/收发已有协议消息。上述条件以当前输入及原请求快照检查，不为每次RDR/ALR引入数据库准入步骤。

### 14.3 六模板与编码链的可复核快照（后续版本证据保留）

只读对象为相邻sms-fw仓库，HEAD=`fda641246c507efe7f45067991a08e159622d2f8`。以下14个文件在读取时`git status --porcelain -- <files>`无变更；哈希取工作树原字节。该事实只固定审查输入，不能证明现场binary由此commit构建、include搜索路径相同、配置实际启用或运行正确。完整仓库其他文件状态不在该断言内。

| 文件 | bytes / SHA-256 |
| --- | --- |
| [yppdia_avptemplate.cf](../../sms-fw/c-cpp/libsrc/kernel/dialib/yppdia_avptemplate.cf) | 2297 / `84d9ab24998028541ecac23b1c856346326d7a298c27991835dcb39f739585f1` |
| [yppdia_iec_avptemplate.cf](../../sms-fw/c-cpp/libsrc/kernel/dialib/yppdia_iec_avptemplate.cf) | 2927 / `8da958f71cc8deead116dc6bf9c3b7b84abb1aa49ba2457f8bb968679b5f0da3` |
| [yppdia_init_avptemplate.cf](../../sms-fw/c-cpp/libsrc/kernel/dialib/yppdia_init_avptemplate.cf) | 2903 / `de7c84a7ddd7246746b212b8761dc319a4d6c58b881931f0fe6b4597afa227fb` |
| [yppdia_term_avptemplate.cf](../../sms-fw/c-cpp/libsrc/kernel/dialib/yppdia_term_avptemplate.cf) | 3076 / `b62afa1dfbe9416f9b481d675af7140f3e38c43537c57cef85f22d0061c63720` |
| [yppdia_refund_avptemplate.cf](../../sms-fw/c-cpp/libsrc/kernel/dialib/yppdia_refund_avptemplate.cf) | 1999 / `6a26a7037badfce37202cfb35917ec27770740ecaf67e957e8e7d8a9314edfed` |
| [yppdia_avptemplate_scap1.cf](../../sms-fw/c-cpp/libsrc/kernel/dialib/yppdia_avptemplate_scap1.cf) | 3495 / `a8744c3f7705460baadb9e6ab15b142e54ed4e49d36eeb15f86de4b1ec07f4cc` |
| [AvpsTemplate.cc](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/AvpsTemplate.cc) | 20390 / `063ed87a7a5d807ce7d68a4d3ec4c890f8fe84b18e24f0701f6fb938175b75d0` |
| [AvpImpl.cc](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/AvpImpl.cc) | 9378 / `58e363aa1d08da12a6fff155a0d87d4298702fb8867b18a61450d4235b6ff828` |
| [DccaClientWrapper.cc](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc) | 125443 / `8c6c650c87304bbbf7072a1f290ab708f101a8a921a6723d275e39dc989859ff` |
| [dia-cc.cc](../../sms-fw/c-cpp/libsrc/kernel/dialib/dia-cc/dia-cc.cc) | 61029 / `3254a1456d2591467a245f29c73174aed95892994174dfaedff0d5452a1e62be` |
| [dia-message.cc](../../sms-fw/c-cpp/libsrc/kernel/dialib/dia-base/dia-message.cc) | 35675 / `9400dac621c906379c006bfc0cf80c68b35dcd974719de1a97b1ce9a7b09d268` |
| [dia-attr.hpp](../../sms-fw/c-cpp/libsrc/kernel/dialib/dia-base/dia-attr.hpp) | 19198 / `33d05e1f954833eb73d70f25334b8d904ef2c6cc5374e5cb09311b0674a5d364` |
| [yppdia_dcca.cf](../../sms-fw/c-cpp/libsrc/kernel/dialib/yppdia_dcca.cf) | 10036 / `dbe6d5c86c7edd223b2b29f9ecc9e3a83bc09a9f07061cab24bdeb7e7a9427e9` |
| [yppdiaengmx.cc](../../sms-fw/c-cpp/appsrc/kernel/yppdia/yppdiaengmx.cc) | 101096 / `70319a363062d9681132881c2d10eb8ab901ab2bb05ca800f3cf022fa5285ec8` |

以下静态链均能在上述快照定位；不编译旧C++、不执行计费请求。

| 层 / 文件定位 | 可证明的行为 | 新库采用边界 |
| --- | --- | --- |
| 模板解析；AvpsTemplate.cc:218–356、357–552 | Grouped按声明成员数递归消费后续有效行，检查成员名字存在，不按名字列表重新排序；缩进不起分组作用。重复检查按code/name，不是完整Vendor/父路径。类型由模板文字分派 | 不导入旧配置解释器或其code-only键。生成定义按Vendor/Code/父路径及binding隔离；冲突不能靠加载顺序覆盖 |
| 变量与值；AvpsTemplate.cc:64–161、AvpImpl.cc:20–97、yppdiaengmx.cc:198/754/763/1044/1088/1164 | 类型与变量名共同决定Getter；STRING调用UTF8 writer，Octets携带长度。SUBSCRIPTIONID_IMSI还受chargedParty条件影响，可为空；SMSCIPV4Addr名称虽像转换器，实际只复制21字节。PID Getter用`%4d`生成十进制文本；refund取旧reservationId内部布局 | 不能仅按变量名称假设IMSI/地址/TPDU字节。新Go调用方给明确owned值，不复制S1内部结构、业务选费方或持久reservationId协议；迁移适配归业务binary |
| 基础编码；AvpImpl.cc:20–305、dia-message.cc:387–486/533–605 | M位由模板writer设1，非零Vendor加V/12字节头；UTF8通过strlen到首NUL，Octets按显式len；Time加NTP偏移。普通IPv4入口有2字节family，draft支路不同；STRING入口不经过Address转换 | 旧字符串地址不能自动变标准Address，原bytes须另验。新库opaque保留零字节与长度；预算先检查再分配，不继承旧缓冲实现 |
| 选择与包络；DccaClientWrapper.cc:1991–2160/3364–3404、yppdiaengmx.cc:2257–2359/2461–2640 | 分route/global的普通、init、term、refund四个模板槽；业务选择ECUR/IEC及终止/退款。encodeCCR给TERM原Session且Number1，其他Number0；配置给Service-Context/版本，故文件名不是最终发送模式 | 没有独立“iec文件自动启用”槽；必须记录实际槽→文件→版本/route映射。库不复制旧reservation标记或TERM编号永远1策略 |
| 消息构造；dia-cc.cc:975–1125、AvpsTemplate.cc:34–61 | 先写baseAVPs，再按文件顺序逐项构造。SCAP1选ACR，库内dia-attr.hpp常量为Command271/App0；模板VSAI却带Vendor193/Acct19302。include与构建产物仍需执行证据确认 | App0不能全部当控制消息；按明确CommandKey分派。不是Ro272/4，也不能将19302直接当header；SCAP1仍SingleAttempt，未放开Ro重传 |
| 答案解析；dia-cc.cc:1447–1530、yppdia_dcca.cf:312 | 示例配置退款code=1292；接收器在顶层或MSCC按code寻找，并有CMD4.0服务级Result路径。局部匹配未校验完整Vendor/父路径 | 这只解释旧接受行为，不证明OCS实际回包或标准允许。新结果解释保留命名空间、顶层/服务级和冲突，不复制code-only错误接受 |

### 14.4 六模板逐个证据和兼容决定

计数为剔除注释后按上述消费规则机械核对的顶层自定义AVP数/含子叶总记录数，不含代码自动补的baseAVPs；不是执行旧parser成功的断言。六份Grouped成员消费/名字核对均完成。片段见候选§13，输入值明确且全部为手工/源码推导。

| 模板 / 记录数 | 本轮已补的证据 | 标准与兼容决定 / 仍缺 |
| --- | --- | --- |
| 通用yppdia_avptemplate.cf；7/32 | 私有10425/23实际INT32；MSCC/RSU有413金额及事件单位；IPv4常量127.0.0.1进入专用Address writer | 顶层Service-ID、金额、Destination-Interface直接SMS等不能当固定SMS标准线；普通IPv4与draft输出需分别锁实际模式。V-LEGACY-GENERIC-01只证明私有键布局 |
| IEC；7/42 | STRING地址不补family；UE-Type0来自IMSI；相同变量可能因Getter走不同编码。它通过普通模板槽才能被选用，无专用IEC槽 | IMSI不能当真实IMEISV，直接2010/2002父路径不能静默移进Recipient-Info；V-LEGACY-IEC-01固定错误地址形状。需部署文件映射/实际变量及成功失败CCA |
| INITIAL；7/41 | init槽，单位RSU=1；比IEC少Originator-SCCP，设备/地址同类问题；V-LEGACY-INIT-01为15字节IMSI填设备的具体片段 | 字段/角色不同不能用IEC样本替代；标准新构造使用真实设备或合法省略，旧形状仅具名兼容候选，不新增身份猜测 |
| TERMINATION；7/44 | term槽；USU来自CC_SSU，Reason2在MSCC正确。MMS声明名中UDL在SMS_KEY前，实际记录/输出顺序反向；附SM-Status/Discharge-Time | V-LEGACY-TERM-01只覆盖正确用量片段，整模板仍有设备/地址/父路径及Discharge-Time条件问题；同键顺序及应用输入关联须完整fixture，不默认所有TERM是delivery-report |
| REFUND；6/26 | 420按Grouped实际构造，不是由名称修成413；0/1292按显式Octets取旧reservation内容；接收code来自配置，非固定标准映射 | V-LEGACY-REFUND-01为已知标准类型冲突。只可在独立具名wire解释域承接，经对端证据决定；同连接/App/命令不能同时启用标准U32和旧Grouped解释，无法唯一绑定时拒绝配置；不能让标准420同键同时有两种类型，也不实现旧token内部格式/退款存储 |
| SCAP1；18/47 | header代码分支271/0与VSAI193/19302分开；553/554/555及606/611/612/617私有定义；607组11次重复，Type序列0,1,5,8,14,16,17,18,2,3,13 | V-LEGACY-SCAP-01只给声明/单位；不得按608排序或用map覆盖607重复。完整header、全部字符串Getter、ACA/错误和部署分派仍待执行/对端证据，不冒称标准Ro支持 |

### 14.5 兼容门槛、参考库与下一步

六模板支持声明需逐项具备：固定模板及代码/构建版本；实际route→模式→模板映射（含Service-Context、proxiable、地址draft模式）；脱敏后仍保持字段结构和二进制长度的输入；旧encoder完整请求与独立解码结果；同一请求的成功/失败CCA或SCAP1 ACA；新Go实现与预期差异报告。抓包可替代部分现场证据但不能替代标准规则；缺现场资料不阻止标准纯codec开发，也不能据文件存在标TEMPLATE-01通过。本轮补齐的是六项源码证据链与条件化片段，尚无执行/现场互操作证据。

参考本地[go-diameter S6c字典](../../go-diameter/diam/dict/testdata/tgpp_s6c.xml)：RDR中Outcome必需，但Outcome的各节点与Cause叶仍可选；因此类型/ABNF能表达结构，却不能代替“至少一个cause、MSC/MME互斥、已知AbsentUser诊断”的程序条件。参考[OTP字典接口](https://www.erlang.org/doc/apps/diameter/diameter_dict.html)：按应用组织定义/消息规则，重复定义/继承有约束；采用隔离定义的方向，不因成熟库能加载某AVP就判定模板业务兼容。没有引入通用规则DSL或运行时旧模板解释器。

新增PAIR-03（ECUR/IEC/退款结构）、PAIR-04（S6c RDR/ALR）、TEMPLATE-02（六模板源码证据与反例）；24个SPEC需求ID保持。第十九轮调整优先级：首版P1优先Sh成对主流程，以及S6c/SGd的T4/IMS无MSISDN、MME/SGSN提醒、Realm等实际分支；余额检查/ASR、SMS计费冲突、Ro5011、仅IMEI、计费SCCP/深层字段和六模板完整执行/部署证据全部保留为首版Ro P1。不是再增加库层级、业务Session管理、MWD服务或存储。

## 15. 当前四应用首版的审计适用范围（2026-11-04修订）

本页Sh/S6c/SGd/Gdd及Ro/DCCA、IEC/ECUR/ASR、六模板/SCAP1和RO-06分析均为首版对应能力的依据。§14.1–14.4全部进入首版；混合章节按实际字段用途检查，不能整体跳过Ro，也不能因此启用未承诺的深层功能。历史第十九轮的Ro后置安排明确撤销。

本页保留原规范、源码和候选字节证据；缺失实际私有定义、目标答案及未决标准条款不因分期改变而关闭。当前有界缺口见Specification§7.1，Ro验收沿覆盖§36–39；439行原三应用矩阵不能代替Ro验收，不新增全Ro逐叶穷举任务。

## 16. 第二十轮：Sh成对结果、Sh-Data与SMS角色分支

### 16.1 本轮证据与收敛范围

2026-10-31重新取得§1固定的29.329 V18.0.0、29.328 V18.1.0、29.338 V18.2.0及Sh ZIP/XSD，五项哈希与既有记录一致；未升级Profile，也未重新获取其余原件。ETSI下载使用同目录的`etsi_TS`路径。依据为29.329 §6.1–6.3，29.328 §6.1.1–6.1.4、表7.6.1、§7.6.16/17及附件，29.338 §5.3.2/5.3.3、§6.2.2/6.2.3和§6.3。来源链接/哈希仍集中在§1。

本轮增加[候选样本§15–17](diameterkit-protocol-vectors.md)：Sh四组各一请求、成功答案和失败答案，共12帧；MME移动提醒与Gdd MT各3帧。它们是静态候选，不是抓包或已运行的库fixture。XSD样本采用官方附件做离线校验；不要求未来Go库集成通用XSD运行时。OpenJDK25.0.4.1内置Xerces直接装载原件；lxml6.0.2/libxml2 2.15.2对第1000/1039行两个type QName的前导空白报错，临时副本仅规范化这两处后结果一致。保留原件/哈希，不把工具差异列为官方勘误；工具与用例范围见候选§16。验收新增PAIR-05/06、XML-02、FIELD-13，见[覆盖基线§22](diameterkit-usage-coverage.md)。Ro现纳入首版，R29/R34未整体关闭。

### 16.2 Sh普通短信子集：命令、结果与关联

本表限定单Data-Reference、普通Public-Identity，无wildcard、Private Identity、多数据效率特性或透明RepositoryData。User-Identity700含Public-Identity601；二者V=1/M=1，Vendor10415。URI按已有规则检查，本例为`sip:alice@example.net`，不把它改成MSISDN或IMSI。外部用户/权限事实由调用方处理，不进入库的RequiredContext。

| RuleID / 来源 | 请求与成功条件 | 失败、接收与边界 |
| --- | --- | --- |
| SH-PAIR-ENVELOPE / 29.329 §6.1 | 四组App16777217，Command306/307/308/309；每组正常包络带VSAI260（Vendor266=10415、AuthApp258=16777217）、Auth-Session-State277；样本取1。Session-Id首个；请求R/P=1，正常答案R=0/P=1/E=0；答案Session/HBH/E2E匹配本跳请求 | 正常答案按所选Sh规则只取一种结果：基础Result-Code或Experimental-Result（Vendor10415）；实验失败不能另加2001，不能因5xxx置E=1。真正基础协议E错误另走RFC错误语法，不硬套Sh正常答案的VSAI/User-Data条件 |
| SH-PULL-SMS / 29.328 §6.1.1、表7.6.1 | UDR306：Data-Reference703 Enumerated=24；本例UDA2001且有SMSRegistrationInfo User-Data。数据存在且AS有读取权限时应返回该数据 | UDA实验5102表示无读取权限；无User-Data也可形成完整失败答案。不得把“所有UDA2001都必有本例SMS数据”提升成全DataRef规则；其他合法空数据/缺数据分支须按原请求解释 |
| SH-UPDATE-SMS / §6.1.2.1步骤4b | PUR307：DataRef24、必需User-Data702；非空IP-SM-GW-Number表示注册/更新，显式空元素表示注销。PUA2001不要求User-Data、Repository-Data-ID或DataRef回显 | PUA实验5103表示无修改权限；4101可表示更新进行中，暴露结果但不自动重试。SCAddress仅适用Pull，PUR标准短信helper不生成它；原始接收保留并报告不适用，不能声称HSS会更新SC。HSS对预配置记录的保护属于HSS业务责任 |
| SH-SUBSCRIBE-REACH / §6.1.3、29.329 §6.3.16/17/22 | SNR308：Subs-Req-Type705=0，DataRef25，One-Time-Notification712=0且存在；不含Send-Data-Indication710。Expiry-Time709是Time而非时长，可省。成功SNA2001可带HSS确认的更早到期时间 | 失败样本实验5104表示无通知权限。不能要求SNA回显请求的Expiry或必须有User-Data；成功SNA缺Expiry按规范表示无限期，不能回填请求值伪造成HSS承诺。一次通知仍只通知一次，不因无Expiry变成持续订阅 |
| SH-NOTIFY-REACH / §6.1.4、表6.1.4.1.1 | PNR309：HSS→AS、Destination-Host必需、User-Identity与User-Data必需；没有具名顶层Data-Reference要求。PNA2001确认该协议请求。PNR是独立入站exchange，用自己的Session/HBH/E2E，不能挂在已完成SNR上等第二个答案 | PNA实验5100用于AS不认识/不支持收到的数据；另有实验5001/5008/5107或基础5012等过程结果。库不因本进程没有订阅记录就生成5107，也不把应用拒绝当XML必然损坏。应用明确选择业务失败；失败结果不会触发库续订或重发短信 |

DataRef24在本固定子集中用于Pull/Update，25用于Subs-Notif及相应通知数据；不能因为四组ABNF可扩展就把25当Pull或把24当可订阅短信数据。SNR退订用Subs-Req-Type=1，结果及不在订阅列表也成功的语义来自§6.1.3.1；完整退订/多身份/多DataRef候选仍待补，不以本轮12帧冒充全Sh功能。

原请求保留命令、DataRef、身份、请求特性等已有有界数据，供答案解释；不是新的订阅数据库。PNR可在SNA前后到达，应按独立入站消息处理，不等待本地SNR状态。仅凭同一用户名不足以授权Origin：Sh的HSS终端身份由部署binding核验，DRA CER身份只是相邻Peer。PUR的Origin是本AS/IP-SM-GW身份，不可由DRA元数据替换；PNR的目标AS也不是DRA本身。

### 16.3 Sh-Data：附件优先、Presence和适用位置

令`P = Sh-Data/Sh-IMS-Data/Extension/Extension/Extension`。已选helper只解释下表短信与可达性路径；其他数据保留raw，不能由全局同名元素搜索返回成功。

| RuleID / 路径 | 附件类型、次数与值 | 应用规则 / 不得推导 |
| --- | --- | --- |
| XML-SMS-NUMBER / P/SMSRegistrationInfo/IP-SM-GW-Number | SMSRegistrationInfo为0..1；其号码为1，xs:string/minLength0，无XSD最大长度。显式空与缺元素不同，预算必须另限 | PUR空串是注销；缺元素无效。样本数字是XML文本，不是TBCD、base64或Diameter Address。XSD通过不证明号码业务语义；§17.1已核对23.008并明确首版表示子集 |
| XML-SMS-SC / P/SMSRegistrationInfo/Extension/SCAddress | 0..1，string长度0..20；Extension也0..1 | 仅Pull数据使用；空值保留，不改写为“无字段”。这是XML字符长度限制，与本库UTF-8字节预算分别检查，不能把Diameter SC-Address3300的TBCD编码套进XML |
| XML-SMS-SBI / P/SMSRegistrationInfo/Extension/Extension/IP-SM-GW-SBISupportIndication | 0..1，unsignedByte枚举0/1 | **元素名采用附件的`IP-SM-GW-SBISupportIndication`**；PDF结构表出现的`IP-SM-GW-SBI-SupportIndication`不是该XSD的已知元素。不能自动别名修正；未知扩展策略也不能在无wildcard位置接收拼错的字段 |
| XML-REACH-MME / P/UEReachabilityForIP/UEIPReachabilityMME | 0..1，unsignedByte枚举仅0 | 0=REACHABLE；1虽满足maxInclusive1却不在枚举中。缺字段不等于不可达，不构造bool(false) |
| XML-REACH-SGSN / P/UEReachabilityForIP/Extension/UEIPReachabilitySGSN | 0..1，枚举仅0 | 与MME分别保留Presence，不移到MME同层 |
| XML-REACH-AMF / 上一Extension/Extension/UEIPReachabilityAMF3GPP、UEIPReachabilityAMFnon3GPP | 各0..1，枚举仅0，3GPP在non3GPP之前 | 字段解析不等于库实现5G接入/注册；缺MME时也不能丢弃SGSN/AMF |
| XML-PATH / 固定schema | 无targetNamespace、无外部schema依赖；已知元素按namespace+local-name+完整路径识别；sequence不可乱序 | 同层同时出现时UEReachabilityForIP在SMSRegistrationInfo之前。`##other`只允许在实际wildcard位置；带外部namespace的同名元素仅为扩展，不是标准号码/可达性 |

首版helper的安全边界：原User-Data为owned、有界OctetString；通用Diameter解码不隐式展开XML。调用方显式调用短信helper或选择应用校验时，按固定路径验证；不要求网络reader运行通用schema引擎。解析前限原字节；解析中累计深度/节点/属性/文本与未知扩展字节，覆盖DOM/结构与raw并存时的预算，不在解析完后才计费。禁止DTD/外部实体，忽略`xsi:schemaLocation`的取件作用，绝不联网。超限/不支持编码/非法XML应有明确本地错误，不能转成空数据；采用UTF-8候选，其他XML编码的支持声明和测试仍需单列。

helper不能把任意合法Sh-Data都宣称为所请求短信数据：空Sh-Data可过XSD，却不满足本例注册PUR；UEReachabilityForIP空容器可过XSD，却没有本例所声明的可达通知项。对未知但schema允许的内容返回“未解释/扩展”，不称为完整语义支持。库不比较注册号码与业务配置、不访问HSS数据库、不记订阅或发续订。

原始User-Data仍按字节做重复请求内容比较；解析产生的值不替代重复比较输入。`<IP-SM-GW-Number/>`与显式成对空标签在XML语义上相同，raw却不同；元素缩进、namespace前缀、实体转义也不做C14N或重编码。解析后重建XML不能保证原字节一致；需要原样转交时直接使用保留的owned raw。

### 16.4 S6c提醒与Gdd MT按发送角色校验

| RuleID / 条款 | 已确定行为 | 剩余边界 |
| --- | --- | --- |
| ALERT-ROLE / 29.338 §5.3.2.5、§6.2.3 | HSS→SMS-IWMSC与MME/SGSN→SMS-GMSC（可能经SMS Router）都使用ALR/ALA8388648、App16777312。MME/SGSN分支必需SC-Address、User-Identifier内User-Name=IMSI，以及SMS-GMSC-Alert-Event3333；V=1/M=0、Unsigned32位图 | 不能用邻居DRA类型推断发起角色；部署授权与实际消息角色一致才解释为对应分支。旧SGd App-ID Alert仍需具名兼容，不能标准模式按payload猜App |
| ALERT-EVENT / §5.3.3.23 | bit0=UE可投递（值1），bit1=转移到新服务节点（值2）。发送未定义位清零，接收raw保留、语义忽略未知位；不是Enum0/1。发送本例移动事件用2 | 不凭两项描述推导它们互斥：值3保留两个已知事实，不能标成非法枚举；不替应用决定动作顺序。全零没有本例所需事件，strict角色helper拒绝无事件；仅未知位不能解释为可达 |
| ALERT-NODE / §6.2.3.1表 | 移动事件且新节点已知时包含Serving-Node2401；MME为Name2402+Realm2408+Number1645。新节点为SGSN时包含Number1489，Name2409/Realm2410在可用时成对提供 | “移动事件必有Serving-Node”过强；不可用时允许省略，应向应用暴露无新路由信息，由应用决定是否另发SRR。MME/SGSN提醒条件不要求读取HSS的“Gdd in SGSN”存储证据；不能套用SRA的HSS角色条件 |
| ALERT-IDENTITY / §6.2.3.1/4 | MME/SGSN直发时User-Identifier/User-Name为IMSI；经SMS Router发给GMSC时为该Router此前TFA给出的User Identifier Alert；SC-Address来自此前TFR的SMS-GMSC-Address | Router可能提供其映射值，不能对所有User-Name无条件套IMSI数字规则。库暴露标识和角色，不建立映射表/待投递列表；跨消息来源事实由应用提供，缺资料不伪装完整验证 |
| GDD-TFR-NODE / §6.2.2.1、§6.3.2.5、表6.3.3.1/2 | SGd与Gdd都用App16777313、OFR8388645/TFR8388646。到MME的TFR含MME-Number1645，到SGSN含SGSN-Number1489；二者在该使用位置均M=0。Destination-Host必需，S6c返回的节点字段不能不看使用位置直接搬flags | 已知SGSN目标的strict helper生成SGSN号码，不用MME号码替代；数字/TBCD边界沿用已选规则。IP-SM-GW目标不能假装MME来补1645，须按独立角色分支解释 |
| GDD-TFR-RESULT / §6.2.2.2/3 | 成功TFA2001；本例MT失败为实验10415/5555且Failure-Cause3303内Enum3304=0。Maximum-Retransmission-Time一旦提供，SMS-GMSC-Address条件必需；时间为绝对时间，与Delivery-Timer时长分开 | Requested-Retransmission-Time以及后来ALR只提供协议事实，不授权库定时重投递、维护MWD、删除旧请求或自动跟随新节点。失败原因与MO命令按各自子集验收 |

ALR普通ALA2001及基础E3004各提供一帧；本过程没有专用ALA应用错误，不借用SGd5555。样本移动事件的Origin是旧MME，Serving-Node指新MME；不得将二者强制相等。绑定只授权收到的终端/角色，不将报文携带的新节点自动变成可连接目标。

### 16.5 对参考实现的具体取舍

本地[go-diameter Sh字典](../../go-diameter/diam/dict/testdata/tgpp_sh.xml)提供类型/Grouped/命令定义，适合做独立字段对照；但不能替代DataRef操作限制或Sh-Data schema。实际[ip-sm-gw SNR构造](../../ip-sm-gw/go/svc/isg-dia/internal/handlers/diameter-handler/sn.go)已选择DataRef25/OneTime0并注释掉SendData，保留该过程选择；其Expiry/OneTime设置M=1与标准发送M=0不同，不能复制成新库默认。已知字段接收M差异仍按APP-MBIT忽略，不做对端误拒绝。

[UDR构造](../../ip-sm-gw/go/svc/isg-dia/internal/handlers/diameter-handler/ud.go)把DataReference作为Unsigned32传入，与Enumerated线路同为4字节，但新typed API须表达规范类型；不靠字节同宽掩盖类型错误。UDR/SNR/PUR将相邻Peer metadata用作Destination的做法不能直接继承到DRA拓扑。新库使用已配置目标/受验证binding；本轮不修改旧服务代码。

[go-diameter S6c字典](../../go-diameter/diam/dict/testdata/tgpp_s6c.xml)与[SGd/Gdd字典](../../go-diameter/diam/dict/testdata/tgpp_sgd_gdd.xml)按应用分开SGSN号码及Alert定义；新库同样以固定使用位置约束，另外补程序角色条件。没有执行这些库的编码器，以上仅源码对照；不能算独立互操作通过。

### 16.6 第二十轮末剩余项（第二十一轮增量见§17；当前优先级见§18.5）

| 优先级 | 具体剩余项 | 完成门槛 |
| --- | --- | --- |
| 首版P1 / 下一份文档优先 | Sh号码的23.008依赖；Sh退订、SNA带数据的非一次订阅、允许的多身份/Private Identity与多数据特性分支 | 固定引用版本/条款，区分XML词法、数据语义和业务状态；补相应正反候选，不把本轮单DataRef覆盖当全Sh |
| 首版P1 / S6c、SGd/Gdd | T4/IMS无MSISDN完整pair、SMSMI三身份/结果；IP-SM-GW发RDR；SGSN/Router Alert完整身份及新节点分支；Gdd MO；具名旧SGd Alert | 从已有§9–14条件继续，逐分支列来源/字段/原请求/失败；新组合独立bytes。IP-SM-GW-Realm规范争议保留，不能由本轮MME样本化解 |
| 首版P1 / 验证 | 新增18帧及XML正反例的独立协议复核；三应用codec/Peer/预算/关闭/故障转移和SCTP验证 | 当前仅静态检查；用户另行授权写代码后实施，不能用同一算术脚本充当独立codec或真实DRA证据 |
| 首版P2 | XML辅助API命名/默认预算、全量支持表和性能材料 | 以实际三应用数据及测量收敛，不建立XML配置服务或订阅引擎 |
| 首版内Ro阶段 | 所有既有Ro未决规则、六模板、专用重传 | 按§15及覆盖§21保持留档，不阻塞本轮三应用设计 |

## 17. 第二十一轮：号码依赖、订阅分支与特殊身份流程

### 17.1 固定来源与号码表示

2026-10-31新增[TS 23.008 V18.1.0](https://www.etsi.org/deliver/etsi_TS/123000_123099/123008/18.01.00_60/ts_123008v180100p.pdf)，实际PDF 855050字节，SHA-256为`89f1ffbb6b2f601a198233a9739d450de3bf61fe8ae599d279f5d8587532c8ca`。核对§2.4.20、§2.4.20A及§2.4.23；同时重新取得既定23.003 V18.5.0、23.204 V18.0.0、29.002 V18.0.0，三项哈希与§1/§9一致。沿用第二十轮取得的29.328/329/338和原XSD。累计规范原件记录由40增至41项；14项旧源码哈希不变，不算新增规范。

23.008的存储描述针对HLR/HSS/IP-SM-GW等网络实体，**不是Diameter库的持久化要求**。IP-SM-GW号码是E.164地址；Diameter Identity是另外的Name/Realm。SC地址引用29.002；其条件存储、注册来源和预配置数据保护仍由应用负责。取到该依赖不意味着实现MAP或HSS数据模型。

| RuleID / 路径 | 已闭合的定义及首版构造契约 | 接收与边界 |
| --- | --- | --- |
| SH-NUMBER-E164 / XML IP-SM-GW-Number | 23.008 §2.4.20→E.164；XSD仍是string。首版显式号码构造入口接受1..15个ASCII数字、首位1–9，直接写入XML；注销单独使用显式空值。该数字入口是本库有限表示策略，不从XSD臆造正则 | 不自动去`+`、空白、TON/NPI、前导零或解析hex/base64。原XML可保留schema合法但不在数字入口子集内的文本，报告未解释的号码表示；不把“helper不支持”写成Diameter线路损坏。不查询号码实际分配/可路由性 |
| SH-SC-TEXT / XML SCAddress | 23.008 §2.4.23→29.002 §17.7.8 MAP-CommonDataTypes的AddressString：MAP为1..20 **octets**，首字节是地址性质/编号计划，随后TBCD。Sh附件却定义SCAddress为0..20 **字符**的string；本库数字文本构造子集同上，仅用于Pull | 不把MAP AddressString字节转成XML文本当作标准要求；23.008没有规定这种转换。保留Sh原文本/Presence，有限数字解释与完整MAP地址域分开；20字符与20octets不能互换 |
| SMS-NUMBER-WIRE / Diameter SC-Address、MSISDN及节点号码 | 29.338 §6.3.3.2、29.329 §6.3.2和已固定节点条款：各自使用位置采用TBCD数字，奇数尾高半字节F；SC-Address无TON/NPI前缀 | 例如数字`12025550100`：XML为11个字符；Diameter数字payload为`21 20 55 05 01 f0`；MAP国际E.164 AddressString才是`91 21 20 55 05 01 f0`。不得通用转换后覆盖原payload |
| SMS-DUMMY-MSISDN / 已选IMS兼容位置 | 23.003 §3.3的例外为**15个0**；Diameter MSISDN701 payload恰`00 00 00 00 00 00 00 f0`，不是空串/单个0/IMSI。仅在29.338明确的缺MSISDN分支使用 | dummy不代表可路由E.164用户，不能被普通号码校验全局拒绝，也不能由库自动替任意缺身份补零；与真实身份、HSS-ID及本地业务关联分别保留 |

NUMBER-01的反例：非空号码16位、非ASCII数字、带`+`/空白、误加91、错误TBCD filler、dummy14/16个0；分别验证有限helper、XML schema或具体线路规则，不用同一个“非法号码”抹掉层次。号码/URI原值继续参与原有Canonical规则，XML整段按raw比较。

### 17.2 Sh剩余订阅分支的具体规则

依据29.328 §6.1.3、§6.1.4、§7.1、§7.6及表7.6.1，29.329 §6.1.5/6、§6.3.6/17/22、§7.1表7.1.1。这里补的是端点消息与纯校验，不实现订阅集合、续订、超时删除或HSS事务。

| RuleID / 分支 | 请求与对应响应 | 可判定范围 / 反例 |
| --- | --- | --- |
| SH-ACCESS-XOR / 访问身份 | 对已选24/25等表7.6.1的OR访问键，只选择一个Public-Identity601、MSISDN701或External-Identifier3111；各自0..1且至少一个。11只接受普通IMS Public User Identity；12接受对应公共用户/服务身份 | 表7.6.1前的定义明确OR为one and only one；不能从User-Identity组的全可选ABNF推导三种身份可同时使用。接收保留原字段并报告违反已选访问键，不按优先级选一个或自动丢弃另一个。与SMS User-Identifier允许的多项条件不同 |
| SH-PRIVATE-KEY / 可选Private Identity | 24/25可带顶层User-Name1（UTF8String、M=1）作为私有身份；不嵌入User-Identity700。请求中公共/私有身份的真实对应关系由HSS判断，实验5002表示不匹配 | 本地检查适用DataRef、位置和语法，不要求查询用户库才能发送；通知按§6.1.4表要求仅在关联订阅带Private Identity时携带它。独立PNR仍可交付handler，不因本库无历史记录而拒绝或虚构已验证关联 |
| SH-UNSUBSCRIBE / 退订 | SNR308的Subs-Req-Type705=1；访问键、DataRef、适用Service-Indication/Private Identity与欲取消订阅一致。25仍带OneTime712=0且无SendData；仅持续11/12的SNR不带OneTime。成功SNA2001；不存在该AS订阅也成功。失败仍可为5104、5001或5002等实际检查结果 | 默认构造不附SendData或请求Expiry；不靠Expiry=0表示退订。后续通知/迟到PNR与退订exchange独立，库不做全局屏障，不吞已授权通知，也不根据2001推导远端所有并发工作已停止 |
| SH-SUBSCRIBE-DATA / 持续订阅并取当前数据 | 示例DataRef11，SubsReq0、SendData710=1、M=0，无OneTime；成功SNA通常含User-Data，XML IMSUserState值0..3。Expiry可省；请求有Expiry时HSS应确认实际到期值，可更早，省略表示无限期 | SendData=0不解释为已请求数据；1的SHOULD返回不能提升成所有成功SNA必含User-Data的MUST。拒绝在DataRef25一次订阅加入SendData。请求单Ref11不需要多数据特性上下文 |
| SH-NOTIF-EFF / 多数据订阅 | 例DataRef11+12共用一个Public-Identity；双方终端特性明确支持Notif-Eff（Vendor10415、Feature-List-ID1、bit0，即值1），SNR列两个703；SNA可返回一个包含两个具名数据片段的User-Data文档，PNR可合并通知 | 本地特性来自binding，远端来自已绑定HSS的PeerFeatures，不能用DRA CER代替。Known不支持→UnsupportedFeature，Unknown→已有具名上下文要求；不是数据库facts。多项全成功才2001；任一失败返回相应错误，恢复原订阅状态是HSS责任，库不发补偿退订 |
| SH-SUB-KEYS / 其他数据引用 | Ref0需要Service-Indication；Ref10订阅按过程表携带Identity-Set；Ref13/19按表带Server-Name，19另需DSAI-Tag。多个Service-Indication/Identity-Set的效率路径同样受Notif-Eff约束 | 不能靠通用DataRef数组声明支持所有数据模型。本轮正向XML新增11/12；RepositoryData、iFC/DSAI等本轮只明确选择键与失败门槛，深层完整helper另按实际需求与覆盖矩阵验收，不混成Ro纳入首版 |

IMPU/PSI分类是调用方明确的使用前提，不能仅由URI外形识别，也不要求库查身份库。

11/12合并XML的顺序由附件决定：`Sh-Data/Sh-IMS-Data/SCSCFName`在`IMSUserState`前。IMSUserState取0未注册、1注册、2有未注册服务的注册态、3认证处理中；它们与可达性的唯一枚举0不是同一个类型。12对应SCSCFName string/SIP URL，空值移除语义按29.328 §6.1.4保留。不得把所有Sh-Data字段都塞到短信的三层Extension里。

连续订阅状态变化、认证瞬态抑制、组合通知以及OneTime后删除订阅是HSS/应用程序行为；库仅编码、解释和报告。本轮选择多Ref11+12作为已审核正向组合；25与持续订阅数据混合的一次通知作用域尚无充分推演，第二十一轮仅允许raw保留、完整语义helper未支持；第二十二轮§18.4已将限制收窄到明确子集之外，不能把历史限制当现行全局禁令，也不能直接判对端违法或引入任意Context消除未知。相同DataRef重复也不能悄悄去重改变raw。

### 17.3 S6c、SGd/Gdd的特殊身份按命令展开

身份由消息用途和字段存在条件决定，不从“15位数字”猜IMSI/HSS-ID。23.204 §6.3a.1/2给出失败后进入SMS-SC的来源流程，具体字段以29.338信息元素表为准。

| RuleID / 位置 | T4无MSISDN子集 | IMS无MSISDN子集 |
| --- | --- | --- |
| SMS-SRR-ID / SRR，表5.2.1.1-1 | 顶层User-Name=用户IMSI，MSISDN缺席，无SMSMI；SC-Address仍必需 | 顶层User-Name=HSS-ID；SMSMI3324只含Destination-SIP-URI3327；不能在组内另带3325或Originating-SIP-URI3326 |
| SMS-SRA-ID / SRA，表5.2.1.1-2及§5.2.1.3 | 成功User-Name按实际IMSI或已选Router MT关联语义；投递用途需要对应节点 | HSS不把SRR转发给IP-SM-GW，而返回其节点号码。成功示例显式选**HSS已知目标IMSI**、User-Name为该IMSI，Serving-Node仅IP-SM-GW-Number3100，避开未决Realm组合。不能假设请求HSS-ID必原样回显；目标IMSI也不存在时的成功User-Name取值仍P1，禁止伪造标准答案 |
| SMS-RDR-ID / RDR，表5.2.3.1-1 | User-Identifier3102/User-Name=IMSI；无SMSMI | User-Identifier/User-Name=HSS-ID；SMSMI仅目的URI；不是顶层User-Name，不用dummy MSISDN替代。IP-SM-GW发送时Outcome3316/IP-SM-GW-Outcome3320/Cause3321按实际原因给出，不能放MME cause |
| SMS-ALR-HSS-ID / HSS ALR，表5.2.2.1-1 | User-Identifier/User-Name=IMSI，无SMSMI | User-Identifier/MSISDN701=15个0；SMSMI仅目的URI。HSS-ID不在ALR中，不能因RDR用HSS-ID就照搬；MaxUEAvailability若已知按原规则携带。HSS ALR不强制MME/SGSN的Event3333 |
| SMS-OFR-ID / OFR，表6.2.1.1/1 | 普通SGSN MO：User-Identifier带可用IMSI和存在的MSISDN；两者不是Sh的互斥访问键。T4回执/External-Identifier条件另按其实际接口位置 | SMSMI含HSS-ID、源URI、目的URI三个子项；Outcome必须有IP-SM-GW原因。若源用户IMSI可用则使用它；不存在IMSI/MSISDN时才使用dummy MSISDN，源用户身份来自Originating-SIP-URI，不能把目的HSS-ID写成源用户IMSI |
| SMS-TFR-ID / TFR，表6.2.2.1 | User-Name=IMSI，按MME/SGSN目标带相应号码；不新增T4命令栈 | User-Name=HSS-ID，SMSMI含源/目的URI、禁止组内HSS-ID。目的IP-SM-GW路由身份由已知部署/查询结果提供；不假装MME/SGSN填1645/1489。不能从SRA回的IMSI覆盖本分支明确要求的HSS-ID |

RDR2001仅确认报告，失败10415/5001或5558继续成立；无成功Alert号时不强制答案带User-Identifier。OFA/TFA的实验5555必须按各自命令带Failure-Cause；MO用SC拥塞等MO原因，MT用容量不足等MT原因，不把Outcome3321和Failure-Cause3304混成一个枚举。ALR没有本过程专用应用错误，失败候选沿用基础E3004。

IMS重投递的外部关联只描述数据来源：OFR三元组→SRR目的URI/HSS-ID→应用显式TFR→应用按结果决定RDR；HSS后续ALR为新的exchange，ALR仅有dummy+目的URI，不足以从当前帧恢复OFR全部信息。应用自行查找待发业务数据；库不提供关联表、补URI、存MWD或安排重投递。T4触发的生成/递送也不进入本库。

### 17.4 SGSN与SMS Router提醒、具名旧Alert

| RuleID | 设计收敛 | 验收边界 |
| --- | --- | --- |
| ALERT-SGSN | 与MME相同S6c App16777312/Command8388648；User-Identifier/User-Name=IMSI、SC-Address及Event必需。可达用1；移动用2，有新SGSN时Serving-Node含1489，2409/2410若可用成对带上；S6c中1489 M1、2409/2410 M0 | 新SGSN未知可以无Serving-Node，不因此查HSS；新节点号码没有强制匹配发送方Origin。号码/名字组合按所选角色检查，不能引用SRA过程的HSS订阅事实当必需Context |
| ALERT-ROUTER | SGSN/MME→Router一跳的标识为IMSI；Router→GMSC一跳的User-Name为该Router此前TFA给出的User Identifier Alert。示例opaque token为`router-alert-21`，不是IMSI。两跳分别有独立Hop标识和可验证的来源 | 这是应用侧转交消息流程，不声明本库为Diameter透明代理；本库不存token→IMSI映射，也不从SessionId推导token。对于真正代理转发的HBH/E2E规则须按既有代理规范，由外部DRA处理，不能套本例新endpoint请求 |
| ALERT-LEGACY-SGD | 固定兼容名`legacy-sgd-alert-v1`：仅App16777313、Command8388648的ALR/ALA切片，沿用声明的S6c字段/角色规则，答案回原请求的App-ID；标准S6c仍为16777312 | 必须在构造期binding中显式选择；不重写收到的App，不全局注册同名alias，不从Event/来源猜标准还是旧格式；同一连接/App定义冲突仍拒绝。命令选择明确，不能把所有SGd消息替换成S6c解释 |

兼容来源为[sms-fw双App Alert分派](../../sms-fw/go/svc/zdimanmx/internal/handler/diameter-interface-handler/diameter-client.go)的S6c/SGd两条handler，和[go-diameter ALR模型](../../go-diameter/diam/sm/smparser/alr.go)。它们证明存在兼容调用面，不证明线上每个字段或分支都被使用；配置、实际bytes和对端接受仍须取证。旧handler用请求Destination充当答案Origin且SGd路径处理与S6c不同，不作为新库协议要求；新库用已授权本地endpoint身份响应，由应用决定接收到提醒后的动作。

[ip-sm-gw RDR构造](../../ip-sm-gw/go/svc/isg-dia/internal/handlers/diameter-handler/rd.go)已使用IP-SM-GW outcome，但标准S6c仍不能继承其VSAI发送、相邻Peer目标或“RDR-Flags=0表示Single-Attempt”的注释。位图值0表示未设bit0，值1才设置Single-Attempt；样本不依靠该注释推导投递策略。Sh旧SNR可同时填IMPU/MSISDN的做法按SH-ACCESS-XOR修正；不修改旧服务。

既有V-SRR-01候选勘误：SM-RP-MTI的Code由误写2400改为3308（29.338表5.3.3.1/1），原值0及长度不变；2400属于LMSI。候选§18保留旧值的错误记录，不能以旧样本长度校验通过阻止纠错。

### 17.5 第二十一轮末的证据与未关闭项（本轮更新见§18.5）

[候选§18–20](diameterkit-protocol-vectors.md)给出8组完整请求/成功/失败，共24帧；另外给出从具名基帧重建的分支替换清单，逐帧长度/标识/摘要用于复核，不把派生清单计成独立对端证据。新增NUMBER-01、SUB-01、IDENTITY-01、ALERT-02及PAIR-07，见[覆盖§23](diameterkit-usage-coverage.md)。所有库执行验收仍Planned。

| 优先级 | 本轮后具体剩余项 | 阻塞范围 |
| --- | --- | --- |
| 首版P1 / 精确语义 | IMS目标IMSI也不可用时SRA成功User-Name；IP-SM-GW-Realm组合；Sh混合25与持续订阅的OneTime作用域、实际使用的Repository/iFC/DSAI深层模型 | 仅相应helper/完整支持声明；不以空值/HSS-ID回显/任意Context兜底。新23.008已取得，不能再笼统列为未取得 |
| 首版P1 / 独立验证 | 本轮24帧和派生分支、Sh XML/号码策略、旧Alert实际bytes/配置；三应用全量逐叶机器清单；Peer/FO/VAL/BUD/SCTP | 静态候选不关闭协议执行、资源安全及现网投产门槛；未开始实现 |
| 首版P2 | 已选数字文本/多身份错误诊断的API表达，有限helper覆盖矩阵和测量后的预算默认 | 不增加通用号码归一化、订阅状态机、业务事务或兼容推断框架 |
| 首版Ro | 保留所有计费分析、未决项、六模板及专用恢复设计 | 不阻塞三应用独立开发，但首版完整交付必须验收Ro，不把范围纳入记作通过 |

## 18. 第二十二轮：成功身份、Realm与混合订阅的有限裁决

### 18.1 证据范围与参考库

本轮复核固定29.328 V18.1.0、29.329 V18.0.0、29.338 V18.2.0及23.204 V18.0.0，重新取得[29.336 V18.2.0](https://www.etsi.org/deliver/etsi_TS/129300_129399/129336/18.02.00_60/ts_129336v180200p.pdf)与[29.338 V19.3.0对照件](https://www.etsi.org/deliver/etsi_TS/129300_129399/129338/19.03.00_60/ts_129338v190300p.pdf)；两PDF哈希分别与§1、§12原记录一致，不升级Rel-18基线，不重复计为新来源。尤其29.336 §6.4.12(g)明确列出Number+Name+Realm，§6.4.18及表6.4.1/1规定3112为DiameterIdentity、M/V；这比单看29.338的ABNF多了一条组合证据，但29.336的S6m/S6n用途不能自动覆盖S6c的应用规则；T4依赖29.337，见§21.2。

| 参考实现 / 固定范围 | 实际观察 | 本库采用与不采用的部分 |
| --- | --- | --- |
| [freeDiameter 9d4bc475字典](https://github.com/freeDiameter/freeDiameter/blob/9d4bc47584bfeced9de708fe00f5629ac4689db5/extensions/dict_dcca_3gpp/dict_dcca_3gpp.csv) | 3100为OctetString，3101/3112为DiameterIdentity；712为Enumerated、V且禁止M | 参考基础类型和flags；字典存在不证明实现了SRA身份回填、节点组合或混合订阅生命周期 |
| 本地go-diameter f26e3528006fcf0d33cb663994a0ec0ccbf7ef55：[S6c字典](../../go-diameter/diam/dict/testdata/tgpp_s6c.xml)、[SRA parser](../../go-diameter/diam/sm/smparser/sra.go)、[SNR parser](../../go-diameter/diam/sm/smparser/snr.go) | 3112被写成OctetString；SRA的UserName可选，sanityCheck只查Origin；SNR有DataReference数组和单个OneTime指针，但sanityCheck未按DataRef解释OneTime | 不继承3112类型差异；结构能承载/Parse成功不等于应用条件通过。不要照搬SNR Parse中Unmarshal报错却返回nil的处理；本库保留分阶段错误。此处仅源码核对，未运行旧栈 |
| [Erlang/OTP diameter_app](https://www.erlang.org/doc/apps/diameter/diameter_app.html)，2026-10-31所见diameter 2.7.2 | 消息字典与应用callback分工；handle_answer同时收到答案及原请求，handle_request处理新入站请求 | 参考原请求关联和独立通知入口；不把库callback提升成HSS订阅数据库，也不据通用OTP文档声称这些3GPP特殊分支已互通 |

本轮没有找到能授权SRA成功回显HSS-ID的直接条款；这只描述所核对材料，不声称已穷尽全部CR或实现。29.338 V19.3.0的成功身份枚举及Realm组合遗漏仍在，不能记为后续版已修复。

### 18.2 SMS-SRA-ID：区分查询方不知道与目标确无IMSI

依据29.338表5.2.1.1-1/2、§5.2.1.2/3、§5.4。SRR用HSS-ID只说明请求定位方式，不能推导HSS没有目标IMSI。成功SRA中的User-Name仍是必需项；外层UTF8String和可选ABNF不能取消该条件。

| 情形 / 原请求与角色 | 构造与解释 | 支持范围 |
| --- | --- | --- |
| SMS-GMSC不知道IMSI，以HSS-ID+目的URI查询；HSS知道真实IMSI | 成功User-Name填写IMSI，V21-IMS-SRR-A继续成立；原请求HSS-ID仍保留供后续IMS TFR使用 | 标准已选子集；接收按协议期望标注IMSI，但不声称核验了用户数据库 |
| SMS-GMSC↔SMS Router且已明确启用HPLMN强制路由 | 成功User-Name可为真实的MT Correlation ID；该值由应用提供，不从HBH/E2E/Session或HSS-ID生成 | 标准条件分支；不能把HSS直答伪装成Router路径来绕过无IMSI问题 |
| 原请求SM-Delivery-Not-Intended=1（只取MCC+MNC） | 允许MCC+MNC+dummy MSIN；这不是普通投递请求可用的IMSI或通用兜底值 | 按原请求目的解释；不能自动把投递SRR改为只取网络码 |
| 目标确无IMSI，且上述Router/只取网络码条件均不满足 | 标准成功builder不提供HSS-ID回显、空值、dummy MSISDN或自造IMSI；在Respond提交前返回本地语义未支持并保留入站应答权，供应用另作明确处理 | `SMS-SRA-ID`规范缺口仍在；不是用户不存在或服务未订阅的证据，库不自动构造实验5001/5003/AbsentUser |
| 收到2001且缺User-Name | 已知成功条件违反；按Application报告，不能因字段在ABNF中为可选而判完整成功 | 与“字段存在、特殊取值尚无标准依据”分开；不向答案再发错误答案 |
| 收到2001、User-Name存在，但明确场景/声明的身份种类与标准分支不相容 | 保留原值、结果、原SRR及RuleID；报告语义未支持/冲突，不输出已验证投递身份 | Wire/Exchange合法的答案按既有契约终结网络等待，不等新答案、不自动重发、不查身份库 |

不能仅因答案等于请求字符串就证明它是非法HSS-ID；IMSI、HSS-ID、关联token的字节形状会重叠。离线解释缺原请求/角色时使用既有MissingContext；缺规范定义则是未支持，两者不能互换。无需新增“目标有无IMSI”的实时facts接口：构造方显式提供所选身份值/用途，接收方报告协议含义与核验范围；业务真实身份由应用负责。标准无IMSI成功分支保留P1证据门槛，不把库边界决定伪装为3GPP已裁决。

### 18.3 S6C-NODE-REALM：标准域保留冲突，限定兼容三元组

字段类型现在明确：`SRA/Serving-Node(10415/2401)/IP-SM-GW-Realm(10415/3112)`为DiameterIdentity、0..1、V/M均置1；Number3100为TBCD，Name3101为DiameterIdentity。§12.4的标准域行为保留。结合29.338响应信息表/ABNF与29.336 §6.4.12(g)，本轮决定可设计一项**显式选择的具名兼容规则**`s6c-ip-sm-gw-realm-v1`，只补三元组，不称官方勘误或放开所有节点组合。它是现有固定Profile/binding机制的一项规则，无新策略回调、插件系统或逐消息开关；在实现及独立验证前仍为Planned。

记N=Number3100、H=Name3101、R=Realm3112、B=SBI3348。下表只讨论IP-SM-GW节点；其他节点仍按原规则。

| 组合 | 标准域 | 选择兼容规则后 |
| --- | --- | --- |
| N；N+H；N+B；N+H+B | 沿用29.338明确组合与各叶条件 | 不改变；没有R不能由H后缀、SRA Origin-Realm或相邻DRA补齐 |
| N+H+R，三叶合法 | 接收保留，报告`S6C-NODE-REALM`未裁决；标准builder拒绝该未支持组合 | 只在S6c App16777312、SRA Command8388647、直接Serving-Node2401采用此组合；报告所用兼容ID与覆盖范围，可以显式构造/提取Name、Realm，不声称标准完整符合性 |
| N+R；N+H+R+B；跨节点混放 | 不靠Realm遗漏推导对端一定违规 | 本轮兼容子集不支持，不自动删字段或补Name；独立已证明的组合错误照常报告 |
| 缺N、重复单值叶或R本身非法 | 分别按已明确的组合、数量、DiameterIdentity规则报告 | 同样检查；兼容不是关闭M位/数量/类型校验，不把已知3112报未知AVP |
| Additional-Serving-Node、ALR或其他应用中的同名字段 | 各自过程规则 | 本规则不覆盖，不由相同Code自动扩散 |

兼容ID及规则版本纳入既有Profile/Canonical指纹；同一解释域不能按内容猜是否启用。默认标准域保持不变。选择兼容后，`N+H+R`的通过仅表示该具名规则允许这个结构；身份来源授权、真实节点归属与路由仍须符合DRA部署binding。提取字段不是自动连接、信任或选择IP-SM-GW，库不将SRA节点地址写入Peer表。三元组+SBI若实际需要，先补独立条款/样本再扩大规则，而非现在预建任意组合。

### 18.4 SH-MIXED-SUB：按DataRef限定一次语义，不增加订阅引擎

依据[29.328 V18.1.0](https://www.etsi.org/deliver/etsi_TS/129300_129399/129328/18.01.00_60/ts_129328v180100p.pdf) §6.1.3.1步骤7、表6.1.3.1、§6.1.4.1，以及[29.329 V18.0.0](https://www.etsi.org/deliver/etsi_TS/129300_129399/129329/18.00.00_60/ts_129329v180000p.pdf) §6.3.22、§7.1：多DataRef分别形成通知请求，712只适用于25。**由这两条合并推导**，不能把一个SNR中出现712解释为全部DataRef都变一次订阅。本轮采纳这个有限解释，不声称获得新的官方勘误；§17.2中“所有25混合helper未支持”的临时限制由下列支持子集替换。

| 项目 | 本轮采纳的有限规则 | 未扩大的范围 |
| --- | --- | --- |
| 构造子集 | DataRef集合为{11,25}、{12,25}或{11,12,25}，各一次；双方终端Notif-Eff已知支持；共同访问键为一个普通IMPU；SubsReq0或1；712=0、M0恰一次 | 不从DRA CER推断特性；无特性/未知特性仍按SH-NOTIF-EFF报告。不自动去重703或支持任意DataRef笛卡尔积 |
| 字段适用性 | 本子集不带顶层Private Identity、Service-Indication、Identity-Set、Server-Name或DSAI-Tag；公共字段须适用于全部所选Ref | 不能因Private Identity对25合法就用于11/12；需要25的私有身份限定时，由应用另发相应单Ref请求 |
| OneTime与SendData | 712仅限定Ref25；11/12保持持续订阅含义。本子集整个SNR不带710，包括值0；避免把全局SendData字段强行投影成只取11/12数据 | “混合+SendData”不是本轮支持分支，不自动删710或判定对端必然违法；需要初值时应用明确选择独立Pull或拆分SNR |
| Expiry与退订 | 请求709表达共同期望到期时间，不是保证的最终上界；成功SNA应返回实际值（可更早），未回709则表示无限期，不能回填请求值。25仍按一次通知更早结束。SubsReq1取消列出的对应订阅，保留25所需712，无SendData | 不承诺多个订阅拥有独立到期值；不记录计时器，不将退订成功当PNR屏障 |
| SNA | 一次混合SNR只有一个结果；全部成功才2001，任一失败由HSS恢复请求前状态。未请求SendData的正例不带User-Data | 不返回库虚构的每Ref成功清单；不自动补偿、重订阅或把失败解释为可安全重试 |
| PNR只含11/12数据 | 独立通知exchange；显式XML解释器按字段路径返回数据，不推断25已触发/被删除 | 不因最初SNR带712而关闭所有订阅；库不保存最初SNR来拦截PNR |
| PNR含25，单独或与11/12合并 | 一次语义只涉及相应的25订阅；HSS在发送相关通知后删除它，11/12不因同一PNR而变一次。合并XML仍按XSD sequence，PNR不复制SNR的712/705/703 | 这是对相关订阅范围的上述推导及HSS责任；库不根据PNA成功/失败实现消费、回滚或恢复，也不靠XML字段推测历史订阅 |

没有承诺“通知与订阅建立顺序可由本库验证”。PNR可早于SNA、与退订交叉、合并来自不同SNR建立的订阅，接收授权沿用部署边界。单次PNR的多个数据片段只形成一个入站完成/答案；同E2E重传由现有去重处理，不以UEReachability值0或DataRef自行建业务去重键。

若应用选择拆成“25一次订阅”和“11/12持续订阅并取初值”两次SNR，必须显式提交两次：分别分配exchange/Session/HBH/E2E、期限与预算，结果分别交付。单Ref11与单Ref25不需要多数据Notif-Eff，11+12仍需要。两次请求不继承单个混合请求的全成全败语义，任一接纳失败/超时/拒绝不能由库回滚另一请求，也不能假造一个总2001。无需新增BatchSubscribe、订阅事务或全局锁。

### 18.5 第二十二轮末状态（当前优先级见§19.3）

[候选§21](diameterkit-protocol-vectors.md)给出确定性派生消息及正反决策输入；[覆盖§24](diameterkit-usage-coverage.md)新增IDENTITY-02、NODE-01、SUB-02。三个问题不笼统同标“已关闭”：SRA是冻结不造身份的库行为、标准缺口保留；Realm是明确默认与具名兼容的差别、标准冲突保留；Sh有限混合语义已形成可实现契约，运行验证未做。没有新增SPEC需求ID、持久状态或通用facts。

| 优先级 | 剩余工作 | 关闭门槛 |
| --- | --- | --- |
| 首版P1 / 下一轮文档 | 已用三应用的逐叶清单及结果分支交叉复核；Sh实际需要的Repository/iFC/DSAI数据模型与国际化子集，避免将类型可解码当语义支持 | 明确实际采用范围/父路径/flags/预算/错误，并给独立候选；不扩大成全Rel-18 |
| 首版P1 / 分支证据 | 目标确无IMSI的标准成功SRA；Realm的标准组合矛盾 | 直接规范澄清或具名部署证据；前者不得伪造回显，后者只能按已明确兼容子集声明，不能因设计有方案而标标准通过 |
| 首版P1 / 独立验证 | 本轮派生样本和Sh混合XML、兼容三元组的独立栈与实际对端；所有既有codec/Peer/FO/VAL/BUD/SCTP门槛 | 静态文档核对不能代替执行；写代码仍待用户指令。兼容子集无部署证据时不承诺现网可用 |
| 首版P2 | 有限解释结果/诊断API与测量后默认预算 | 沿用现有入口，不增加身份注册表、订阅引擎或动态兼容框架 |
| 首版Ro | 原计费规则、未决项、六模板/SCAP1及专用恢复 | 分期现修订为首版，保留全部验收义务 |


## 19. 第二十三轮：实际逐叶清单、结果闭包与Sh数据边界

### 19.1 单一清单与本轮发现

新增[三应用字段清单](diameterkit-field-inventory.md)，它是本页§2–3的逐叶补充：111行AVP叶定义、10项已采用XML叶，包含明确父路径/用途、Vendor/Code、类型、发送V/M、次数、值/条件、预算和样本缺口。每个字段保留A/B调用或D设计来源；不是整个Rel-18字典冻结，也没有新增生成DSL。已用公共字段和各主命令的原始/解释字段纳入，未实际启用的可选功能按清单§3.6登记而不虚构使用。

两个业务仓库HEAD已变化，本轮按当前工作树重新读取指定handler及模拟器；14份关键源文件有本轮哈希，详见清单§1，旧历史摘要不覆盖当前观察。复取23.040 V18.0.0原件并核对既有哈希，表1a闭合3313/3314/3315/3322原因的0..255值域和已分配0..13；值0是MSC无寻呼响应，不是“缺省”。未知/保留值、节点适用性和缺席结果条件分别检查，库不复制业务错误映射。

结果按9个答案分别列成功数据义务、15个Sh实验结果与10个SMS实验结果及通用基础错误；见清单§4。补上PNA的3GPP/5001与5008，以及TFA的5552（§6.2.2.2）/MME角色5557（§7.3.10）。这几项不能只凭单处简表或已有成功/失败示例漏掉。未知结果仍有界保存，命名空间/结果冲突不能被顶层2001覆盖。

### 19.2 Sh深层模型与国际化的有限支持

本轮A/B调用证据聚焦Sh Ref24/25；Ref11/12及有限混合仍为前轮已选D能力。RepositoryData/iFC/DSAI完整helper未发现实际业务启用，不能从模拟器结构声明推断支持需要。清单§5明确其结构、序号/选择键及还缺的语义证据：XML SequenceNumber限0..65535，不等于AVP716 U32；ServiceData要求子元素，不是任意裸base64；iFC和DSAI不创建业务规则执行器、版本存储或HSS事务。此项从笼统“首版P1所有深层模型”收窄为按实际需求启用的扩展能力，不是归入Ro延期。

清单§6区分UTF8字节校验、ASCII DiameterIdentity、SIP/固定NAI语法及XML合法字符/转义。通用UTF8不承诺IDNA/SASLprep完整语义；未知国际化范围不通过新Context伪装成已验证。原值不被Unicode正规化/URI解码/XML重编码改变；Peer身份比较与请求内容Canonical比较仍是两种不同用途。

### 19.3 第二十三轮末验收（当前优先级见§20.5）

新增FIELD-14、RESULT-06、XML-03、TEXT-01，见[覆盖§25](diameterkit-usage-coverage.md)；[候选§22](diameterkit-protocol-vectors.md)含5个schema正例、5个负例及具体字段/结果/文本变体。正例中的未启用模型仍不能由typed helper假称支持，schema验证不证明业务语义或运行安全。

| 优先级 | 下一份交付 / 门槛 |
| --- | --- |
| 首版P1 / 已采用字段 | 清单逐行标注的待补独立bytes、零诊断/父路径、完整结果变体；A/B与已选D各自对照，不能以111行当测试通过 |
| 首版P1 / 标准与部署 | 真正无IMSI成功SRA直接依据；Realm标准矛盾与具名兼容对端证据；旧Alert配置/bytes、Sh混合独立核对及DRA身份/保护 |
| 首版P1 / 实现验证 | API/codec、R32/R36、FO/VAL/BUD、SCTP及独立栈验证，仍待写代码指令 |
| 按需求启用 / Sh扩展 | Repository/iFC/DSAI完整helper、全IDNA/SASLprep、其他可选深层字段：先提供真实调用/语义范围，再补依赖和独立样本；当前raw保留不宣称完整支持 |
| 首版P2 / 首版Ro | 前者为默认预算/诊断API/测量；后者仍保留计费分析、六模板/SCAP1及专用恢复，不作为当前三应用前置 |

R29/R34整体仍未关闭；本轮没有业务存储、订阅服务、路由缓存、自动重试或额外通用facts接口。


## 20. 第二十四轮：独立完整帧、跨组条件与时间边界

### 20.1 候选交付与参考实现

新增[逐叶完整报文](diameterkit-leaf-vectors.md)：16组请求场景、136条完整hex，其中49条是应用失败的合法替代答案，其余是请求、成功及负例/控制。每帧独立给bytes、长度、SHA-256、原请求与预期，不依赖旧候选变换才能提取。它们由规范手工定义、临时排字节和另一检查器核对，**不是独立Diameter栈证据**。旧96段hex和33条派生记录不变，文档合计232段hex；没有新实现、持久状态或规范版本升级。

本轮重新核对既有原件29.328/29.329/29.338/23.040及其摘要；规范证据仍按原固定版本，不因网页新版本而替换。其他库只用于其可证明的范围：

| 参考材料 | 本轮观察 | 采用及限制 |
| --- | --- | --- |
| [freeDiameter固定提交add_avps.c](https://github.com/freeDiameter/freeDiameter/blob/9d4bc47584bfeced9de708fe00f5629ac4689db5/extensions/dict_dcca_3gpp/add_avps.c) | 3329/3330/3331引用Time类型、V1/M0；3314为U32、V1/M1；2406为Grouped | 对照类型/发送flags；声明不存在时间与原请求、双节点或业务结果联动的证明 |
| [OTP diameter_app](https://www.erlang.org/doc/apps/diameter/diameter_app.html#handle_answer/4) | handle_answer收到原Request及解码结果/原bytes；一次call至多一次answer/error完成 | 参考请求关联与结构/应用callback分工；不移入OTP应用状态，不靠重试另拿一个“更好”的结果 |
| [本地go-diameter Time](../../go-diameter/diam/datatype/time.go)与[SGd字典](../../go-diameter/diam/dict/testdata/tgpp_sgd_gdd.xml) | 以高位处理2036回绕，但DecodeTime长度不为4时返回空Time及nil错误；字典将3330/3331列可选Time | 参考时间表示，不继承错误宽度默许；补3字节Time完整负例，以及跨era的有Presence零Time。字典可选性不替代条件必需性 |

### 20.2 S6c：Additional组合、跨组限制和诊断

`S6C-ADDITIONAL-PAIR`依据29.338 §5.2.1.1/§5.2.1.3、§5.3.3.7，叠加于各叶定义，而非新增通用规则引擎。

| 条件 | 本轮明确行为与候选 |
| --- | --- |
| Additional单组允许形式 | 仅SGSN Number；SGSN Name+Realm+Number（HSS支持且收到Gdd指示）；MME Name+Realm+Number；仅MSC Number；MSC Number+MME Name+Realm。完整SRA正例覆盖五种形式；缺Realm、空组、重复叶分别给负例 |
| 普通HSS回答两个节点 | 原SRR bit0 GPRS-Indicator必须置位；两节点为SGSN与MSC/MME之一。两个Grouped各自合法不足以证明整份SRA合法；N24-SRA-MSCMME明确拒绝分成两个节点的MSC+MME。单组MSC Number+MME Name/Realm仍是条款允许的组合，不能按两个节点误拒绝 |
| 两组不能重复信息 | N24-ADD-DUPNODE重复同一SGSN，不能因为父AVP不同就通过；检查直接可见的相同信息，不查询真实节点别名注册表。IP-SM-GW不是Additional已选组合，不能继承Realm兼容；接收保留/报告未支持，不伪报已知叶为未知必需AVP |
| LMSI及Alert MSISDN | 顶层2400恰4octets；HSS成功答案含MSC且VLR已用LMSI时应携带。缺LMSI只有在该外部事实已知时能作条件违例；接收端不能由MSC存在猜VLR事实。SRA/RDA的3102中Alert MSISDN与请求号码不同的例子已补，不保存号码映射 |
| MWD和零诊断 | SRA MWD0存在；SRA5550/3314=0、MSC RDR3318/cause1/3322=0完整正例与缺诊断控制并存。零、缺失、保留值与越255分别解释；N24-RDR-DIAG256等不因U32可承载而通过。MME/MSC报告仍互斥 |

上述双节点约束限本轮明确的普通HSS流程；SMSF、IP-SM-GW特殊投递及Router改写按各自过程，不能用通用“节点最多两个/必须不同host”的新规则覆盖它们。Router可能将两组改为自己的地址，必须先按可信角色/过程选规则，不从Origin字符串猜角色。

### 20.3 SGd/Gdd：时间条件与成功附加字段

依据29.338 §6.2.2.1–3、§6.3.3.17–19以及[RFC6733 §4.3.1](https://www.rfc-editor.org/rfc/rfc6733.html#section-4.3.1)。复用已有原请求证据和有界用途输入，不增加PLMN数据库、时钟服务或新通用Context。

| 条件 | 具体契约 / 候选 |
| --- | --- |
| 发送3330 | 目的用户属于GMSC的PLMN；若有SMS Router，其Alert能力已知；有3330就必有3332（E.164 TBCD，无TON/NPI）。这些是发送方场景事实，收到3330不能据此证明PLMN关系；本库不查HSS/配置服务 |
| 接收3331 | 仅5550且原TFR有3330；适用节能/不可达过程。缺原请求报告缺输入，已知原请求无3330是条件违例，两者不混淆。超原上限是应用错误，正确宽度/关联不代表业务可用 |
| 上限相等 | 表6.2.2.1/2规定不得超过，§6.2.2.2过程用prior to。严格早于满足两者，首版完整builder采用该子集；相等C24-TFA-EQUAL保留值并报告该边界未完全裁决，不自动截减一秒或断言对端违规。不将整类时间功能标未支持，也不另造兼容Profile |
| Router提醒身份 | Router→GMSC的5550答案含3331时，3102内User-Name的MT Correlation ID必需；只给MSISDN不能替代。直达MME/SGSN不强制此token。此身份不属于成功2001的必需附加字段，不以IMSI格式检查token |
| 成功附加字段 | TFA2001可携带3301报告TPDU、MME EPS中的ECGI/TAI；OFA2001可携带3301报告TPDU。完整bytes和ECGI6字节/TPDU201字节负例已补。TFA2001附3331是失败专属字段条件违例。OFA的T4 External-Identifier不是普通SGd/Gdd成功必需字段 |
| Time与本地期限 | 3329/3330/3331/709均4字节绝对Time；3306是U32持续秒数，3307是绝对起点。2036回绕例Start=fffffff0、Requested=00000000、Max=00000010分别为06:28:00/16/32Z，不能裸U32比较或把0丢成缺省。已过的ALR可达时间保留原文，由应用决定是否仍投递；本库不延长pending、不安排重投递 |

本节场景将角色和PLMN等事实写入fixture说明，不等于设计一个远端事实验证服务。原文/有界Report交付与完整业务语义检查范围分开；没有事实时不能把“库不知道”升级为对端错误。

### 20.4 Sh与结果分支的纠错

Sh SNR Ref11请求710=1及709，成功SNA分别给更早到期/省略无限期/没有702三个控制：缺702属于29.328 SHOULD偏离，不升级成无例外MUST；省709不能回填请求。宽度3字节和重复709单列静态负例。UDA在明确请求数据存在的Ref24场景缺702，才按该场景报告成功载荷不足。

纠正PUA附加字段用途：29.328表6.1.2.1及29.329 §6.1.4，Repository-Data-ID715用于已协商Update-Eff/Enhance的多Repository更新失败定位；Data-Reference703用于已协商Enhance的多DataRef失败定位。它们不是普通Ref24成功回显；N24-PUA-DATAREF明确未协商Enhance的负例。相应深层正向过程仍未启用，不用错误正例扩大支持。

结果完整帧包括Sh已选流程的5001、5101–5104（5101另配HSS已知PSI访问键不适用的请求，不随意替换合法IMPU请求的答案）、4100/4101及PNA5008/5100/5107；5002沿用已有带Private的完整退订pair。SMS实验结果在各自角色/命令补齐：§7.3.2还明确SMS-IWMSC在SGd按MSISDN返回OFA5001，本轮补完整帧；Gdd同值只列接口边界控制，不能自动扩大该条款。TFA5552不因信息简表遗漏而拒绝，5557仅MME为已证明正例；SGSN同码保留并报告角色范围。MO每种cause3..6、MT每种cause0..2均有独立5555帧和透明3305；缺父/缺必需cause、方向错用、双结果/缺结果/未知值有单独负例或控制。所有失败仍只完成一次，不自动重试、删除订阅或修改MWD。

Sh5011的完整特性触发、PUA5008透明数据容量、5105/5106/5108的Repository/DSAI合法触发仍与当前有限helper范围分开：本轮只给部分结果识别边界记录，不将不匹配的Ref当成正向协议流程。结果覆盖不声称穷举全命令×角色×数据模型。

### 20.5 本轮关闭范围与下一优先级

新增PAIR-08、NODE-02、TIME-01、RESULT-07；继续细化FIELD-14/RESULT-06，24个SPEC ID不变。独立完整bytes的文档缺口已补本轮优先子集；独立栈、库执行、资源和现场门槛仍Planned。

| 优先级 | 剩余工作 |
| --- | --- |
| 首版P1 / 候选审查 | 对本轮完整帧作独立协议实现复核；Sh5011完整特性触发、已选D的T4 OFA External-Identifier整条pair、各待补单字段重复/预算/类型反例与静态规则映射。不能把136条算成全部111叶已闭合 |
| 首版P1 / 有限标准分支 | 真正无IMSI成功SRA、Realm冲突、重传时间相等边界；分别保留不伪造/有限兼容/严格早于子集，独立部署与规范证据各自关闭 |
| 首版P1 / 实现与部署 | 用户授权写代码后再做codec/API、Peer/关闭/发号、FO/VAL/BUD/SCTP；Sh混合、旧Alert真实bytes和DRA身份/保护证据仍需验证 |
| 按实际需求启用 | PUA多数据失败定位及Repository/iFC/DSAI完整helper、其他非实际使用深层字段；不因结构可承载宣称全流程支持 |
| 首版P2 / 首版Ro | 默认预算/诊断API和测量仍P2；Ro及六模板/SCAP1完整分析现纳入首版对应验收 |

R29/R34不整体关闭。没有增加业务存储、订阅状态、Router token映射、自动业务重发或规范争议的运行时猜测。

## 21. 第二十五轮：Sh特性失败、T4 OFA与逐叶反例收敛

本节为当前规则；§20及前轮候选的具体勘误见§21.4。新增251条[完整候选](diameterkit-leaf-vectors.md)，每条有完整bytes、长度、摘要及预期，111叶均有精确父路径静态反例。与旧136条合计387条逐叶帧，加旧候选96段message/avp-hex，文档总计483段hex。数量不代表独立实现或全Rel-18功能通过。24个SPEC ID不变，新增FEATURE-01、PAIR-09、FIELD-15，执行状态Planned。

### 21.1 Sh 5011的完整触发与发送/接收区别

依据[29.329 V18.0.0 §7.1、表7.1.1](https://www.etsi.org/deliver/etsi_ts/129300_129399/129329/18.00.00_60/ts_129329v180000p.pdf)及其明确采用的[29.229 V18.0.0 §7.2–7.2.1](https://www.etsi.org/deliver/etsi_ts/129200_129299/129229/18.00.00_60/ts_129229v180000p.pdf)。重新下载29.229，SHA-256与§1一致。5011不是任意业务失败占位码，也不同于首版Ro中尚未裁决的双结果问题。

`SH-FEATURE-TRIGGER`：特性键为应用+终端+命令对+Vendor-Id+Feature-List-ID；Sh的Notif-Eff为10415/List1/bit0。请求必须广告实际使用的特性；使用Notif-Eff合并11/12或混合11/25时，628的发送M必须为1。请求本身不使用扩展、仅愿接受特性化答案时，M可以为0。答案628发送M始终0，所列是答方完整支持集，不能只回双方交集；答案具体内容只能使用双方共同特性。629、630各必需一次、U32、V1/M0；266为0/266、U32、M1，不能和外层Vendor头混为一项。

| 对端与原请求事实 | 答案与本库行为 | 完整候选 |
| --- | --- | --- |
| 认识628，支持请求全部特性 | 不得用5011；按实际过程答2001或其他适用失败，附完整支持集；本例返回Ref11/12数据 | V25-SH-NEED-R/A |
| 认识628，但M1组中至少一项不支持 | Experimental-Result=10415/5011，无Result-Code；回答方支持集。语义失败交给调用方，不自动改Ref/拆批/重新订阅 | F25-SH-5011；F25-SH-PARTIAL-5011 |
| 全部不支持与部分不支持 | NEED例受限旧端点支持集0；PARTIAL例UDR广告List1=9（Notif-Eff+Additional-MSISDN），Rel-18答方支持集1，仍必须5011；另给支持集9的成功替代答案 | V25-SH-PARTIAL-R/A；bit3未用于Ref11/12的返回内容，不新增Ref17 helper |
| 不认识628且原请求628 M1 | 基础Result-Code=5001 AVP_UNSUPPORTED；一个Failed-AVP内包含至少一个原始628；不是实验5001 USER_UNKNOWN | F25-SH-BASE5001；缺Failed-AVP反例 |
| 原请求无628且不使用扩展 | 不得仅因对端支持集不同返回5011；答方可广告完整支持集，无特性则可不带 | V25-SH-PLAIN-R/A、N25-SH-NOFEATURE5011 |
| 原请求本身未使用扩展，628 M0只是接收能力 | 未支持可选报价不构成“缺M1要求”的5011触发；按共同支持集构造答案，不凭M0补出多Ref支持 | V25-SH-OFFER-R/A |
| 只知道邻接DRA能力或缺原请求 | 不能当成终端特性证据；接收保留对端报告，限制应用检查范围；不猜“这是标准5011触发” | 既有binding/RequiredContext，不新增facts接口 |

表7.1.1的M/O描述特性在相应版本的实现义务，与628 AVP头的M位不是同一概念。NEED失败场景特意模拟不支持mandatory Notif-Eff的旧/受限端点，不宣称它是符合Rel-18的HSS；PARTIAL以optional bit3展示符合mandatory基线的部分不支持。支持集0是明确4字节零位图，不是遗漏630。未知Vendor/List-ID不能因位0相同被解释为Notif-Eff；重复键冲突仍按既定报告，不能OR或最后写入覆盖。未知特性范围不自动授权公开helper生成未支持功能。

`APP-MBIT`仍适用：上述M是**发送义务**。已理解628的接收端按AVP表注释忽略单纯M不匹配，不能把C25-SH-ANSWER-M1/NEED-M0都映射成线路5004；若应用本身使用的特性不支持，再按原请求M组和相应规则解释。不能为了宽容接收继续生成已知错误flags。基础5001的未知AVP场景采用受限对端字典，不改变本库Sh Profile认识628的事实。

公开库提供有界特性表示、纯校验/结果解释；具体终端支持集由应用给定且只能广告实际支持能力。已有KnownUnsupported/Unknown helper门槛不因这些低层收包候选自动放宽，不新增自动探测、跨请求能力缓存、持久订阅或自动降级重试。规范允许发送方选择另发共同特性请求，是应用决定；不能借原pending延长deadline/改变原身份与预算。本节不启用Update-Eff/Repository模型，也不改变Ro延期。

### 21.2 T4的SGd-OFR/OFA、External-Identifier和边界

新增固定原件：[TS 29.337 V18.0.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129337/18.00.00_60/ts_129337v180000p.pdf)，SHA-256 `d8c3ba536a4a0888f247e985d1bdbc4e4d78f745a72aba3c8f542fbbd0d02503`。§1/§5.1明确T4上的SGd-OFR/OFA由29.338定义；§6.1.2规定SCTP。由此补齐接口依赖，不把整个T4设备触发功能纳入首版。

`T4-OFA-EXTERNAL`依据[29.338 V18.2.0 表6.2.1.1/1、/2及§6.2.1.3](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf)：本例为SMS-IWMSC→MTC-IWF一跳，App16777313、Command8388645；SC-Address为已配置的MTC-IWF国际号码，仍是无TON/NPI的TBCD。OFR携带已知发送用户IMSI/MSISDN及原TPDU；OFA反向返回。顶层3111/V1/M0（2026-11-07实施勘误：29.338表6.3.3.1/2覆盖定义规范；旧文误写M1）在经T4向SMS-IWMSC提供计费所需发送者External-Identifier时必需；3102内同名叶不能代替顶层项。供计费用途不意味着本库计算费用或提前实现Ro。

| 配对/反例 | 明确外部前提与结论 |
| --- | --- |
| V25-T4-OFR-R / V25-T4-OFA-A | 外部映射、用户身份均已知；2001附`device-001@example.net`。这不是MME→IWMSC普通SGd一跳，也不自动把该叶转发下游 |
| 同请求 / F25-T4-OFA-5555 | 身份仍已知且本次供计费用途成立；MTC-IWF拥塞cause4，5555+3303/3304+3111。未知用户、身份取不到或不供计费场景未在此伪造固定结果 |
| N25-T4-NOEXT、N25-T4-NESTED、N25-T4-NAI | 分别缺条件必需顶层、放错父组、ASCII NAI缺local part；都不可凭同名或“是UTF8”放行 |
| N25-T4-WRONGAPP、N25-L092-TEXT/DUP | 错用原生T4 App16777311不能匹配原请求；另给非法UTF8和重复顶层3111的静态反例 |

2026-11-07第二十五批实现记录：V25/F25原M1字节保留为接收兼容候选，新增M0发送投影；有限T4Charging及ASCII身份、独立对照已实现，范围和SCTP未完成项见[batch-25](../spec/history/batch-25.json)。原候选不再作为M1发送依据。

Native T4的Device-Trigger/Delivery-Report使用App16777311、8388643/8388644，不能因都叫T4就给OFR/OFA换App。29.337中的DRA还可能指Delivery-Report-Answer，不能与本repo的Diameter Routing Agent混同。29.336是S6m/S6n/S6t依赖，不是T4定义；§18.1旧“T4用途”措辞已纠正。29.336的User-Identifier/External-Identifier和Realm证据仍有效，但不会自动裁决S6c组合。

T4是否启用及MTC-IWF目标由具名部署绑定/当前调用语义明确提供；库不由Origin字符串猜接口、不从DRA CER取终端特性、不查询HSS/S6m或维护号码→MTC-IWF/External-Identifier映射。本切片采用SCTP；已存在的通用TCP能力不表示本T4部署已获规范证明。完整NAI国际化仍沿用TEXT-01有限范围：本例只证明ASCII子集，UTF8可承载不等于全IDNA/SASLprep支持。

### 21.3 叶反例与新发现的SRR条件

`S6C-SMEA`：29.338表5.2.1.1/1规定，SMS-GMSC支持接收两个号码时SRR必须带SM-RP-SMEA；§5.3.3.4的GPRS-Indicator bit0恰表达该支持。不能只加bit0和Additional答案而漏掉3309。§5.3.3.3引用[23.040 V18.0.0 §9.1.2.5](https://www.etsi.org/deliver/etsi_ts/123000_123099/123040/18.00.00_60/ts_123040v180000p.pdf)地址格式。本库本次有限子集为十进制国际号码：Address-Length（有效半字节数）+Type-of-Address=0x91+数字TBCD；最多15位E.164数字，整个地址仍受规范12octets上限。11位示例为`0b 91 21 20 55 05 21 f3`。不剥成SC-Address、不引入MAP/RP IEI或外层Length；库不从SC号码或用户MSISDN自动推算原始SME。

这只是23.040地址的数字子集；alphanumeric、其他TON/NPI按未支持表示报告，不能一律说非法UTF8或用数字解码器误解析。源文件首段另转引SM-RL地址规则，不把采用的此处23.040完整地址解释升级为所有MAP/RP地址共用编码。其他部署表示仍须具名证据；没有新增完整短信协议解析器。完整正向构造的格式与独立对端确认继续保留P1。

| 叶/路径 | 补齐的候选与规则 |
| --- | --- |
| L046 / SRR3309；L047 bit0 | V25-SRR-TWO-R补正确SMEA；六个V25-SRA候选覆盖原五种Additional形式及基线变体，并与新请求关联。缺3309、长度不一致、TOA最高位、13octets及重复分别有完整输入 |
| L054/L056 | V25-DIAG-MME=13、SGSN=5的5550答案；14保留值、256越界、3字节和重复分开。0仍按Presence保存，但不会把MSC原因0默认填入其他域 |
| L080 | 29.338 §5.3.3.6(h)/(i)、§5.3.3.39：N+B、N+H+B；B=0不支持SBI、1支持。未知2报告，不能当false；不额外拼Realm或启用HTTP/SBI |
| L093/L097 | OFR bit0明确SGd0/Gdd1；不据bit重选App。TFR显式0、原1及保留位31控制，发送清未定义位，接收语义忽略但保留raw；不生成投递任务 |
| L110/L111 | 保留原ECGI6反例，补TAI4/6及两叶重复；位置层始终1496/1600，ECGI7/TAI5，不套Ro tag |
| L001–L111 | 每叶独立完整静态负例及路径索引；数值/Time用错误宽度，文本用非法UTF8，单值Octets用重复；数值/文本单值另给重复。可重复字段不误判为单值，opaque bytes不强套UTF8 |
| 资源 | FIELD64/65与REPEAT2/3成对边界；64字节/2项只是fixture配置，合法编码超本地额度为ResourceLimit，不是标准5004/5008；原文保留与解释结果仍受现有owned预算 |

静态层输入可在未实现深层语义时验证类型/数量，但不能冒称Ref0/10/13/19、多Repository失败定位或Realm标准组合已经完整支持。索引明确这类输入的范围；每行一个父路径见证不是覆盖别名集合全部命令和所有值。后续把它们编成可执行fixture时，须继续验证层次优先级和完整路径，不默认所有错误都在线回Answer。

### 21.4 旧候选勘误与其他库对照

旧bytes、哈希和历史评审全部保留，当前证据分类修正如下，禁止在新测试中照旧标整体通过：

| 旧记录 | 纠正与替代 |
| --- | --- |
| V21-SUBMULTI-R、D22-MIX-SNR-R、D22-MIX-PNR-R | 实际使用Notif-Eff却628 M0；降为发送规则负例/宽容接收控制。替代V25-SH-NEED-R及V25-MIX-SNR/PNR-R均M1。D22答案未回已支持Notif-Eff集合，也不能作为完整特性正例；新答案均补集合 |
| C24-SH-SN-5011 | 旧单Ref请求没有完整特性触发且答案无SF；仅保留结果码表示边界。完整实验5011见F25-SH-5011/PARTIAL-5011，不把旧记录重标成功 |
| V24-S6C-SR-R；对应V24-SRA双节点pair | 原请求bit0=1但漏3309；请求降为应用条件负例，对应答案仍可单独验证叶/跨组形状，整pair不合格。以V25-SRR-TWO-R和V25-SRA系列替代；其他以此为基线的反例只能证明目标静态问题，不能证明“仅有一个完整应用缺陷” |
| 旧Sh无SF答案 | 只证明既列字段/结果；不能据此证明本地已经支持Notif-Eff等功能的完整端点广告合规。若场景明确支持集非空，答案必须补完整集合；无特性旧端点控制不冒称全Rel-18实现 |

参考[本地go-diameter Sh字典](../../go-diameter/diam/dict/testdata/tgpp_sh.xml)，HEAD仍为`f26e3528006fcf0d33cb663994a0ec0ccbf7ef55`，工作树SHA-256 `cb08c2e82f12c18714bf9e4ceea97317cd3795c156a4d0690c9f585cf13d89be`：628声明M可选、三个子项各必需一次，629/630禁止M。它能证明结构，不能决定本次请求是否实际使用Notif-Eff；不继承“may M就一律M0”的错误推导。[OTP diameter_dict 2.7.2](https://www.erlang.org/doc/apps/diameter/diameter_dict.html)将应用ID、类型/发送flags、命令与Grouped次数分开声明；可借鉴表示分层，不能据通用字典推定5011/T4业务条件已实现。本轮没有运行其他Diameter栈，也不把文档对照当互通证据。

### 21.5 剩余优先级

| 优先级 | 具体下一门槛 |
| --- | --- |
| 首版P1 / 独立复核 | 251新增帧及旧候选勘误的独立解析/专家或对端证据；将111叶索引扩展到实际使用父路径/命令矩阵。Sh特性与T4不再笼统写“缺完整pair”，但尚未成为可执行fixture |
| 首版P1 / 标准证据 | 无IMSI成功SRA、IP-SM-GW-Realm标准组合、Requested=Maximum时间边界未决；SMEA所选数字编码须独立对端确认。仍只输出有限构造子集/范围报告 |
| 首版P1 / 实现与部署 | 待用户启动代码工作后实施codec/规则、资源上限、Peer/关闭/发号、FO/VAL/BUD/SCTP；T4按具名绑定验证SCTP与终端身份。库发布证据与具体DRA投产证据分开 |
| 按功能需求 / P2 | 未启用Sh深层功能、完整国际化、其他SMEA表示及所有取值笛卡尔积不强行扩为首版；默认预算和诊断易用性待测量。Ro及六模板原分析纳入首版验收 |

R29/R34保持P1，原因转为上述明确门槛，不再用新增接口或持久化解决证据缺口。

## 22. 第二十六轮：实际命令/父路径矩阵与剩余P1复核

第二十六轮快照（当前增量见§23）为[命令与父路径矩阵](diameterkit-command-path-matrix.md)：111叶展开439个正常命令/精确Vendor:Code路径条目，18种命令；253条有在位候选、142条有本表口径的定向反例。其余297条没有纳入同命令同路径的定向反例，其中公共叶254条、应用叶43条；不是说这些类型未定义或整个应用从未测试。上一轮111行各有一个反例，仅证明叶定义层的见证，不能作为439个命令路径的全部证据。

新增[30条完整输入](diameterkit-leaf-vectors.md)针对跨命令、禁止位置和父实例/Vendor隔离；旧387帧保持，逐叶语料共417帧，连同旧96段hex共513段。没有实现库、生成器、测试框架或业务服务；24个SPEC ID不变，新增MATRIX-01，执行状态仍Planned。

### 22.1 覆盖单位与实际使用证据

`APP-PATH-INSTANCE`：校验与覆盖键必须包含App、Command、R/E形状、每层Vendor/Code、直接父实例和当前业务条件。V=0的Code3111、SF下的0/266、Experimental-Result下的0/266不是可互替的值；重复父组的子项数量分别统计。发送/接收、静态/应用/预算、未知语义/缺原请求仍使用既有分层，不建立新的Context或通用facts服务。

依据[RFC6733 §4.1、§4.4及§7.6](https://www.rfc-editor.org/rfc/rfc6733.html#section-4.1)，AVP身份由Code与Vendor联合确定，Grouped子项属于其所在组。不能先用全树Find取得一个同Code叶，再宣称当前父路径必需项已满足。相同类型可以复用底层检查，但每个命令的生成规则、数量、使用位置仍须独立验证；只测试公共解码函数不能证明命令描述正确。

矩阵保留13个明确禁止的应用路径：IMS OFR中的MME/MSC/SGSN outcome六叶；HSS-ID在SRR/RDR/ALR/TFR的SMSMI组内四项；Originating-SIP-URI在SRR/RDR/ALR的SMSMI组内三项。它们不能机械继承允许在其他命令中的组定义。13项均有定向应用反例；在位字节不是成功样本，未知扩展的安全保留也不能满足已选分支的必需条件。

26个条目为未启用语义/结构边界，另有Realm具名兼容条目。它们不能因为出现在矩阵中就变成首版必须实现的业务helper。特别是UI-SMS各叶不能把“Grouped中可表示”解释为所有六个命令都允许用该身份完成当前过程；本轮区分SRA/RDA的Alert MSISDN、TFA的Router User-Name token和OFA顶层External-Identifier。

实际源码证据在[逐叶清单§1](diameterkit-field-inventory.md)追加S15–S19，原S1–S14摘要不变：S15发送SRR只填MSISDN、SC和MTI，没有声明双节点bit0，不应把新增SMEA条件强加到它的所有普通请求；S16 UDA入口只Parse/日志，没有逐叶业务消费证据；S17/S18的PUA/SNA读取两类结果但未证明User-Data、Expiry或SF处理完整。S19把应用超时映为OFA5555/cause4，是业务层决定，不迁入公开库的网络超时处理。命令配对要求与实际handler消费层次分开，不把一条注册函数或字典字段当生产启用证明。

### 22.2 新增报文复核与库实现对照

本轮逐帧重查上一轮251条输入的header、完整父边界、padding、长度/摘要、关联字段和目标类型/数量；3份正向Sh XML仍复用原附件。语义审查特别保留以下限制：PARTIAL5011使用UDR的optional Additional-MSISDN bit3，不能移到PUR；NEED中不支持Notif-Eff者是旧/受限端点；静态N25输入中C类模型/兼容Realm不被升级为完整应用正例；SMEA请求与SRA配对只证明所选数字子集，不证明外部地址真实归属。

| 30条输入的用途 | 具体预期与验收范围 |
| --- | --- |
| 跨命令14项 | SNR的Expiry宽度；PUR/PNR/UDA的User-Data重复；RDR/ALR/TFA的User-Name非法UTF8；RDR/ALR/SRA/RDA内MSISDN重复；OFR的ECGI/TAI长度另补2条，及OFA diagnostic重复1条。按实际ID和矩阵计数，不共享另一命令的结论 |
| IMS禁止位置10条 | 三个错误outcome域分别覆盖cause/diagnostic两叶（6行），四个HSS-ID禁位、三个OriginURI禁位，共13行应用路径；不是一律静态5004 |
| 精确定位6条 | OFA同Code不同Vendor、SF内Vendor-Id假冒、实验结果缺Vendor但其他父组有同名叶、SF父实例借叶；两个Proxy组合法控制和第一组缺State负例。未知可选字段可保留，不能充当所需已知叶 |

上述分组按输入数为14+10+6=30；一条outcome负例有两个明确应用叶目标。反例的实际目标在矩阵登记，不把它所携带的其他正确公共字段也计为负例覆盖。E=1错误包络不纳入正常语法439行分母，九种答案族的E形状/关联验证仍要单列，不能把E3004当成E0应用失败。

| 参考实现/材料 | 已核对行为 | 对本库的约束 |
| --- | --- | --- |
| [go-diameter message.go](../../go-diameter/diam/message.go)，HEAD f26e3528006fcf0d33cb663994a0ec0ccbf7ef55；工作树SHA-256 `ad182fd58a6fdb4ea4b08e0999a6b405aa2cd9952a0d827d31cf8dcccdf22e21` | `FindAVP(s)`查字典时使用Vendor，但`findFromAVP`递归只比Code；`FindAVPsWithPath`先转换成Code路径，`avpsWithPath`也只比Code，并汇总多个父组命中 | 不把这些搜索结果用于必需叶/唯一性或特性/结果命名空间判断；按每层Vendor/Code和父实例校验。这里是具体源码风险，不声称所有go-diameter API都如此，也没有修改该仓库 |
| [OTP diameter_dict 2.7.2](https://www.erlang.org/doc/apps/diameter/diameter_dict.html) | 应用ID、类型/flags、messages和Grouped成员分别定义，重复项与单项有不同表示 | 借鉴独立描述与次数表示；不能据字典存在就宣称本矩阵业务条件通过 |
| [OTP diameter_app 2.7.2](https://www.erlang.org/doc/apps/diameter/diameter_app.html) | 应用回调接收消息/错误，答案处理有原请求上下文 | 结果解释使用原请求，业务分支仍归应用；不把回调变成存储查询或自动重试政策 |

当前环境未提供tshark或OTP可执行文件。本轮使用临时文档检查器和源码/规范复核，不把它包装为独立Diameter栈互通，不安装新依赖或写Go原型。不同检查脚本只提高字节核对可信度，仍不能自证协议实现正确。

### 22.3 剩余P1条款再核对：结论与关闭条件

本轮复核固定[29.338 V18.2.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf)、已存[29.338 V19.3.0对照件](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/19.03.00_60/ts_129338v190300p.pdf)及[29.002 V18.0.0](https://www.etsi.org/deliver/etsi_ts/129000_129099/129002/18.00.00_60/ts_129002v180000p.pdf)，既有原件摘要不变；不升级Rel18基线，也不声称已穷尽CR。

| 缺口 | 本轮证据/裁决 | 能关闭什么、仍需什么 |
| --- | --- | --- |
| 真正无IMSI的成功SRA | Rel18和所查Rel19表5.2.1.1/2仍要求成功含User-Name，允许IMSI、特定Router MT Correlation ID、仅MCCMNC目的下的dummy MSIN；没有通用HSS-ID回显授权 | 不造身份、不改请求用途、不自动答用户不存在的库行为保持明确。要扩展标准成功分支须直接条款/CR；具体部署非标准行为只能具名兼容，不影响普通真实IMSI分支实施 |
| IP-SM-GW-Realm | Rel19 §5.3.3.6的ABNF仍列3112，允许组合仍无Realm；29.336的另接口三元组不能自动覆盖 | 标准域继续报告未决，s6c-ip-sm-gw-realm-v1只覆盖SRA/2401/N+H+R。独立对端可验证此兼容能力，但不能关闭标准矛盾；不能借矩阵扩到Additional/ALR或N+H+R+B |
| Requested=Maximum | Rel19表6.2.2.1/2仍“不超过”，§6.2.2.2仍“prior to”；字典Time类型不能裁决相等 | 标准完整builder取严格早于子集；接收相等保留并报告边界，不伪判已证明违规。独立实现采取其中一边只说明该实现选择，不代替规范裁决 |
| SME地址表示/出现条件 | 29.002 §7.6.8.17也引用23.040地址；ASN.1 SM-RP-SMEA为1..12octets，支持既选长度上限，但不能把ASN.1外层TLV塞入Diameter。29.002 §12.1.3的MAP出现条件是HPLMN过滤功能且已有对应RP参数；29.338的SRR表则写双节点能力 | 两接口条件不同不能互相覆盖；继续按Diameter已采用条款要求bit0场景提供3309，不从SC补值。数字完整地址构造仍需独立对端输入/复核；未支持其他TON/NPI不是未知字段错误。S15普通SRR没有启用bit0，不因此被判漏SMEA |

这些缺口按具体能力阻塞完整声明，不应阻塞无关的普通Sh、真实IMSI的S6c或SGd/Gdd实现设计；也不能因不阻塞所有能力就标为Resolved。已有具名兼容必须单独验收版本/来源/终端授权，保持无持久化边界。

### 22.4 下一门槛与优先级收敛

| 优先级 | 明确工作与退出条件 |
| --- | --- |
| 首版P1 / 公共路径正确性 | 先验证Vendor、父路径、父实例、未知扩展和正常/E错误形状；相同Code或其他组不能补必需叶。矩阵公共叶254项未纳入定向反例，应在实现阶段以共享类型测试+逐命令规则检查补齐，不能只再堆同类手写报文 |
| 首版P1 / 应用路径 | 43项无本表定向反例，其中16项为按需结构边界；其余27项优先Sh访问键/Private Identity/DataRef、IMS OFR outcome、各请求SC/TPDU、OFA报告及SMSMI跨命令URI。先有正确场景基线，再派生一项缺陷，不把未启用模型作为“正例” |
| 首版P1 / 独立协议证据 | 30新增及251前轮候选仍需独立栈/专家或对端复核；失败预期应包含层次、完整路径、是否终结exchange及是否可回Answer。路径字节检查通过不等于所有语义条件通过 |
| 首版P1 / 规范或兼容 | 按§22.3分别处理四项门槛；没有新增facts、存储确认或恢复服务。Ro争议按独立Ro验收处理，不扩写本三应用矩阵 |
| 实现授权后 / P2 | 待用户启动代码工作后编制可执行规则/codec/Peer/预算/SCTP验证；预算默认和性能须实测，矩阵不指定新运行架构。未启用深层功能/全国际化按实际需求扩展 |

R29/R34保留P1；本轮完成的是精确覆盖分母、实例隔离规则和证据分类，不把下一阶段代码工作提前实施。

## 23. 第二十七轮：应用路径缺例、身份组合和四项P1的关闭边界

本轮只更新文档。公开Go1.26、无持久化的端点库边界不变，首版Sh/S6c/SGd-Gdd，Ro全部纳入首版对应验收。沿用24个SPEC ID和MATRIX-01、IDENTITY-02、NODE-01、FIELD-13/15等验收，不引入新API、通用facts、订阅表或业务重试。§22是上一轮统计快照，本节及[矩阵](diameterkit-command-path-matrix.md)记录当前增量。

### 23.1 27条应用路径缺例：先保证基线适用，再派生目标缺陷

439行分母不变，定向反例从142行增至169行；剩余270行=254公共行+16个未启用语义边界行。上一轮43个应用缺例中的27个非边界条目均有本命令/父路径输入；这仅关闭“没有一个定向反例”的文档缺口，不关闭每行的全部次数、值域、角色、原请求条件、独立实现与互操作验收。10条新增在位控制将有在位候选行数从253提升到263。

[逐叶§10](diameterkit-leaf-vectors.md)新增56份完整帧：27份补矩阵缺例、10份应用基线控制、3份Sh身份组合反例、10份SMEA控制、4份Realm控制和2份跨era的零值时间控制。旧417帧原字节保留；逐叶共473帧，加旧96段message/AVP hex共569段。本轮不复制16个未启用深层能力的全套流程，也不以增加帧数替代独立验证。

| 精确路径与命令 | 基线/缺陷与证据 | 证据等级和范围 |
| --- | --- | --- |
| UDR/PUR/PNR的700/601 | 普通IMPU基线；N27-*-PUBLIC-UTF8 | A场景的逐命令UTF8反例；不从SIP字面区分IMPU/PSI |
| UDR/PUR/PNR的700/701、700/3111、顶层0/1 | V27-*-MSISDN/EXTERNAL/PRIVATE先提供9条基线，后分别破坏TBCD、NAI结构或UTF8 | Ref24 Pull/Update、Ref25独立通知的D补齐；不冒称现有handler消费所有键 |
| UDR/PUR的703 | N27-*-NOREF删除唯一必需叶 | 缺必需与未知枚举、更新失败或特性未支持分开 |
| IMS OFR的3316/3320/3321、3322 | V27-IMS-OFR-ABSENT为cause1/diag11；N27-OFR-NOCAUSE和DIAG256 | 不复用RDR父路径结论；诊断不能推导cause，IP-SM-GW不借MSC diag0 |
| SRR/RDR/ALR/TFR的3300 | N27-*-NOSC | 按各命令独立检查；SMEA、3332或用户号码不能补SC |
| TFR/OFA的3301 | N27-*-TPDU201 | 201octets字段上限反例；预算充足、非UTF8错误；OFA可选不等于在位后无限长 |
| TFR的3324/3326、3327；SRR/RDR/ALR的3324/3327 | N27-*-URI3326/3327 | IMS子集按准确命令/父路径；既有禁止位置反例仍有效 |

原矩阵L021.PUR、L025.PUR改用V24-SH-PU-R作为在位基线；旧V24-SH-PU-KEY-R具有已知PSI不适用的应用缺陷，不能拿它派生“仅有一个错误”的新反例。原bytes/原失败应答仍保留。新消息复用来源关联标识，仅作互斥的离线变体，不是可按顺序发送的多答案会话。

### 23.2 Sh访问键：Note9与通知上下文不得遗漏

依据[29.328 V18.1.0表7.6.1及Notes5/9、§6.1.4](https://www.etsi.org/deliver/etsi_ts/129300_129399/129328/18.01.00_60/ts_129328v180100p.pdf)。Ref24/25采用三种备选访问键，但有以下限制；它们是既有SH-PRIVATE-KEY/访问键校验的细化，不增加身份数据库：

- Public Identity / MSISDN / External-Identifier三选一，不能把本轮独立基线拼成同时带多键的合法消息。
- Public Identity或MSISDN指向多Private时，操作指定其中一个Private的场景须显式带相应Private；对外部身份的访问，Note9排除额外Private。无身份关联证据时，不猜“相同用户”或自动产生5002。
- `N27-UDR/PUR/PNR-EXTERNAL-PRIVATE`中两个字段各自格式合法，违规是组合条件。普通非空ASCII NAI正例不升级为完整国际化；非法UTF8、错误NAI格式和不允许的组合分别报告。
- PNR的Private来自原订阅用途；本库不依赖本进程曾经发送SNR，也不维护订阅表。单Ref25基线、原订阅含Private是fixture前提；独立接收缺该外部前提时只承诺字段/当前消息可判条件，不能伪造“原订阅不匹配”的结论。PNR不增703来猜DataRef，XML由显式解释得出。

混合{11,25}/{12,25}/{11,12,25}仍限定共同IMPU、无Private、无SendData的已选子集。本轮替代身份只用于单Ref24/25；不能由§23.1行覆盖推导这些混合场景也成立。SNR的既有Private字段规则同步遵守Note9，但本轮三份组合反例只覆盖所列UDR/PUR/PNR。

### 23.3 无IMSI成功SRA：条件与身份种类，不由字符串外形裁决

复核固定[29.338 V18.2.0表5.2.1.1-2及§5.2.1.2/3](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf)，并对照已登记V19.3.0同表。仍未发现“目标确无IMSI时可在普通成功SRA回显HSS-ID”的直接授权；不声称穷尽全部CR。以下是已收敛的本地行为，不是新的线路身份约定。

| 可复核场景 | 期望与剩余门槛 |
| --- | --- |
| V21-IMS-SRR-R/A：请求方不知道，HSS知道真实IMSI | 正常成功解释；请求HSS-ID与答案IMSI分别保留，不要求字符串不同 |
| N22-SRA-HSSID：同一数字形状明确被提供为HSS-ID，目标确无IMSI，普通HSS投递用途 | 标准完整builder拒绝未支持语义；接收保留2001和原值并报告SMS-SRA-ID未支持，不输出“已验证IMSI” |
| 复用N22-SRA-HSSID的相同bytes，但应用明确其值就是真实IMSI | 字符串相等不是违规证据；按普通IMSI语义解释，库不承诺核验外部用户库。只改变场景说明，不新增一份假不同bytes |
| Router强制HPLMN路由，或原请求仅MCC+MNC用途 | 只有原请求和明确角色分别授权MT Correlation ID、dummy MSIN分支；缺角色/原请求报告缺输入，不按Origin名称猜Router，不为绕过问题改请求目的 |
| N22-SRA-MISSING/EMPTY与N27-REALM-NOUSER | 成功缺身份/空身份是独立条件问题，不被Realm未决遮蔽；不能向答案回错误答案 |

发送方本地拒绝发生在Respond提交前，保留已有应答权；库不自动变更为5001、5550或伪造身份。接收方Wire/Exchange有效的关联答案终结原等待，语义未支持与原结果一同交付，不等另一答案。已有MissingContext/ContextRequired、Unsupported和已证明条件违规保持区分，以上名称引用既有分类，不新增公共错误类型。关闭标准无IMSI分支仍需直接规范/CR；对端“能够接受”仅可成为具体非标准部署证据，不新增一个空泛no-IMSI兼容开关。

### 23.4 Realm：复用固定兼容，不扩大为任意组合规则

§18.3规则保持：标准域对N+H+R保留未裁决报告；`s6c-ip-sm-gw-realm-v1`只覆盖S6c SRA直接2401的三元组。固定[29.338](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf)的信息表/ABNF与允许组合仍有差异；所查[19.3.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/19.03.00_60/ts_129338v190300p.pdf)没有消除该差异。兼容能力和标准问题分别关闭。

| 输入 | 默认标准域 / 固定兼容域的预期 |
| --- | --- |
| D22-REALM-A，与D22-REALM-R关联 | 相同bytes在默认域报告未决，在具名域只放行N+H+R结构；切换的是事先固定的规则环境，不能按包内容切Profile |
| C27-REALM-NR | N+R不由兼容授权；不从Name后缀、Origin或DRA补字段 |
| C27-REALM-NHRB | SBI=1也是已知叶，但N+H+R+B未被三元组规则授权；不自动删R/B降级到另一合法组合 |
| N27-REALM-TEXT | 兼容也不豁免DiameterIdentity语法；含空格必须报告 |
| N27-REALM-NOUSER | 独立成功身份缺项；不能因Realm未裁决而停止所有可判字段检查 |

Additional/ALR、其他应用及跨节点混放保持原范围。应用可明确选择标准允许的N或N+H输入，但库不得把收到的三元组静默改写成它们。兼容退出门槛为已固定Profile指纹/终端binding、对端可复核样本及以上正反输入；通过后也不关闭标准矛盾，不产生新的Peer、路由或节点身份授权。

### 23.5 Requested=Maximum：时间存在性、era和业务条件分开

沿用§20.3：表6.2.2.1/2的上限措辞与§6.2.2.2更严格的过程措辞尚未裁决。标准完整builder选择严格早于子集；接收相等保留/报告未决，不减一秒、不自动执行重投递。相等与不相等的比较先使用显式era解释，不能比较裸U32。

| 依序判定的输入 | 本地结论 |
| --- | --- |
| 3331重复、宽度不为4，或结果不是5550 | 分别为静态/已证明应用条件问题；不进入“相等未决”掩盖已有问题 |
| 5550有3331、原请求缺失 / 已知原请求无3330 | 前者缺输入；后者已知条件违规，不混作同一失败 |
| 原请求和两个Time都有明确era，Requested < Maximum / > Maximum | 分别处于已选上限子集 / 超上限违规；PLMN、Router身份等其他条件仍需各自满足 |
| C24-TFA-EQUAL或C27-TIME-ZERO-R/A | 相等：标准完整builder拒绝该未裁决分支，接收交付原5550和值及范围报告；不判为已获标准证明的非法时间 |
| 缺era、无法解释绝对时间 | 保留各4字节值，报告缺转换输入；不使用运行机器当前年份或网络deadline代替 |

C27中Start=fffffff0/era0，Maximum=Requested=00000000/era1，即2036-02-07 06:28:16Z；零是有Presence的真实时间。答案bytes复用V24-ERA-MT-WAIT但原请求上限从00000010变成0，说明仅看答案无法判断关系。Router→GMSC的5550/3331另需3102/0:1 token，MME直达不强加；外部角色未知仍报告检查范围。本次没有新相等兼容Profile或定时器；规范澄清关闭标准问题，对端选择只证明其具体行为。

### 23.6 SMEA：数字子集格式设计具备直接依据，互操作仍待验证

本轮进一步核对[29.002 V18.0.0 §17.7.6的SM-RP-SMEA ASN.1注释](https://www.etsi.org/deliver/etsi_ts/129000_129099/129002/18.00.00_60/ts_129002v180000p.pdf)，其1..12octets内容明确包含三部分。结合29.338 §5.3.3.3和[23.040 V18.0.0 §9.1.2.3/5](https://www.etsi.org/deliver/etsi_ts/123000_123099/123040/18.00.00_60/ts_123040v180000p.pdf)，**所选数字编码的三段格式及长度单位不再作为无依据设计缺口；独立实现/对端证据仍P1**。此结论限已选国际十进制子集，不是全TON/NPI或号码有效性认证。

| 层次 | 明确契约 / 本轮输入 |
| --- | --- |
| 内容与长度 | 原3309 payload为有效数字数d的一字节Length、0x91、低半字节优先数字；d为1..15，内容长度恰2+ceil(d/2)，奇数末高半字节F，其他数字位置0..9。结构12octets上限与E.164的15位上限分开；所选格式最多10octets |
| 正向结构边界 | V25的11位原值，加C27-SMEA-EVEN(12)、MAX15(15)、MIN1(1)。MIN1只证明格式，不声称该号码实际分配或可路由；零长不作为本helper有效号码 |
| 定向反例 | N27-SMEA-PAD、DIGIT、STRIPPED、TLV分别隔离填充、非数字、少Length和错误MAP外层；C27-SMEA-16DIGITS是E.164值域反例，即使总长度未超过12也不通过 |
| 范围控制 | C27-SMEA-NATIONAL为另一已定义TON，当前helper未启用；不能只因为TOA不是0x91就报未知AVP或笼统格式坏包。其他TON/NPI先保留raw及范围报告，格式可判错误仍单列 |
| 出现条件 | SRR双节点能力bit0要求3309，N25-SRR-NOSMEA仍有效。C27-SMEA-BIT0-CLEAR只证明反方向不成立：bit0未置不据此禁止在位SMEA；不得将它配成双节点SRA正例。MAP的HPLMN过滤条件不迁入Diameter |

SC-Address无TON/NPI，SMEA有Length+TOA，两者不能共享“去前缀”的构造规则。源SME由应用提供，不从SC、用户MSISDN、MAP字节长度或DRA地址推算；无值时本地报告所缺条件，不自动降bit0或填零。独立验收至少需一条已知数字的SRR3309原payload以及奇偶/边界的解读，核对三段、半字节顺序和对端角色；仅看到解析为OctetString不能关闭此门槛。

### 23.7 参考实现与剩余优先级

| 来源 / 固定证据 | 本轮观察及设计取舍 |
| --- | --- |
| 本地go-diameter HEAD f26e3528006fcf0d33cb663994a0ec0ccbf7ef55；[default.go](../../go-diameter/diam/dict/default.go) SHA-256 `ca8a13b7df2e86ea564b114c073aaf2852a7e82520a06a4cf1ffe24fa7886c6c` | 3309为OctetString，3330/3331为Time；类型声明不能证明SMEA内部三段或两个时间的关系。未改该仓库 |
| [sra.go](../../go-diameter/diam/sm/smparser/sra.go) SHA-256 `2cfb3d295f934fb9e94b423267d406b332dd869c0c03fca0271719dac928193e` | UserName为可选字段且sanityCheck只要求Origin信息，故Parse成功不能覆盖成功SRA身份条件。不是所有go-diameter能力的泛化结论 |
| [datatype/time.go](../../go-diameter/diam/datatype/time.go) SHA-256 `45c8d4a64caa8e2c19c35ce300cd12c753f8e18d54eb6eba200fc98d11af7a92` | 当前DecodeTime宽度不为4返回零值指针且nil错误，按高位选1900/2036转换；本库需明确宽度错误，原U32与显式era转换分开。只读源码，无运行兼容性结论 |
| [OTP diameter_dict](https://www.erlang.org/doc/apps/diameter/diameter_dict.html)、[diameter_app](https://www.erlang.org/doc/apps/diameter/diameter_app.html)，所见diameter2.7.2 | 字典负责消息/Grouped描述，handle_answer带原请求；借鉴分层与关联，不将通用callback扩成身份库、订阅存储或业务调度器 |
| [Wireshark的Rel18 MAP-SM-DataTypes](https://raw.githubusercontent.com/wireshark/wireshark/master/epan/dissectors/asn1/gsm_map/MAP-SM-DataTypes.asn)，2026-10-31查看 | 文件注明29.002 V18.0.0并保留三段注释；只是独立项目保存的规范定义，非运行解析器/抓包证据，也不是Diameter实现认证；以固定ETSI原件为规范依据 |

| 优先级 | 下一步与关闭条件 |
| --- | --- |
| 首版P1 / 公共规则与证据 | 补9个答案族正常/E错误包络、结果选择与原请求关联、父实例；254公共行缺例用共享类型检查加逐命令规则覆盖，避免复制254份等价手写报文 |
| 首版P1 / 已选应用条件 | 27条行缺例已补，但缺必需/重复/值域/错误父组并非每行齐全；优先身份组合、Known/MissingContext、角色限定及结果条件，不把169行当Passed。其中3份Note9组合反例也需独立复核 |
| 首版P1 / 标准未决 | 无IMSI成功身份、Realm标准组合、Requested=Maximum各保留直接规范/CR门槛；库本地行为已经明确，不能靠对端容忍度宣布标准问题关闭 |
| 首版P1 / 独立互操作 | SMEA数字格式设计已有依据，仍须独立解读及所用DRA/终端证据；Realm具名兼容另验固定binding和反例。56新帧和既有语料仍是候选，执行状态Planned |
| 实现授权后 / P1与P2 | codec/规则、Peer/发号/关闭、预算与SCTP验证仍未执行；16个未启用应用边界按需扩展，性能/默认预算实测。Ro不加入本轮分母或首版门槛 |

R29/R34不整体关闭。本轮验收范围是可审核文档、静态字节/关联/来源及XML一致性；不增加业务持久化、任意组合引擎或实时facts服务。

本轮静态复核结果：56新帧长度/padding/摘要与关联检查通过，27项定向输入各只改变一个目标叶；旧30项Vendor/父实例控制复跑通过；2份复用XML通过原始Rel18 XSD。旧417帧与96段hex/33条派生记录保持，14份文档434个本地链接、24个SPEC ID、19份业务源码及3份参考库源码摘要核对通过。这些检查不包括独立Diameter栈、网络互通、库实现或性能验证。

## 24. 第二十八轮：九答案族正常/E包络、结果选择与请求关联

本轮入口为[九答案族覆盖](diameterkit-answer-envelope.md)，以148份新增完整输入及线路外场景覆盖UDA/PUA/SNA/PNA、SRA/RDA/ALA、OFA/TFA。旧473份逐叶帧、96段早期hex及33条派生记录保持；全doc现在717段message/AVP hex。正常E0矩阵仅登记明确叶目标，新增53条原缺例行，现222行有定向反例，剩217=201公共行+16按需应用边界。E1及连接/完成竞争另列，不纳入439行分母。

公开Go1.26、无持久化端点库；Ro与旧计费分析纳入首版实施。所有输入仍是文档候选，没有codec、网络、并发或独立栈运行。沿用BASE-ERROR、RESULT-ENVELOPE、BASE-FLAGS及既有VAL/RESULT/MATRIX验收，不增加公共API、错误策略插件或业务状态。

### 24.1 正常与E语法先分开，条件随后验证

依据[RFC6733 §3、§6.2、§7.1–7.7](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.2)、[29.329 V18.0.0 §6.1/6.2](https://www.etsi.org/deliver/etsi_ts/129300_129399/129329/18.00.00_60/ts_129329v180000p.pdf)、[29.338 V18.2.0 §5.3.2/6.3.2/7](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf)。九族实际索引见新覆盖§2/3，不以普通ALA的基础错误冒充专用3GPP应用失败。

| 检查项 | 正常E0答案 | E1基础错误答案 |
| --- | --- | --- |
| Header | R=0、T=0；P回原请求；App/Command/HBH/E2E来自同一attempt请求 | 同样回原请求；不把App改0或Command改通用值；E不是来源授权 |
| Session | 各正常命令必需，固定首位，与原请求原字节相同 | §7.2虽写0..1，§6.2要求原请求有就回；本九族请求都有，因此全九族E也必需。不能用ABNF可选取消该条件 |
| Origin | 264/296各一次，本地身份/realm；收到时按绑定来源校验 | 同样各一次；可以是授权错误代理，不要求等于成功终端；不能从请求Destination或相邻Peer自动补值 |
| Auth-Session-State / VSAI | 九族正常277必需且所选值1；Sh正常260必需且App16777217；SMS沿用不发/接收忽略VSAI语义 | 不强制277/260，也不强制正常应用成功载荷；在位已知扩展的类型、外层边界与采用的应用覆盖仍适用 |
| 顶层结果 | 三应用正常答案恰选一个268或297；重复同值也不覆盖；结果与角色/用途另检 | 必需一个0/268；297不替代268。可选297的结构与应用含义见§24.2，不套正常XOR |
| Destination / Proxy | 不带293/283；终端对本跳收到请求的Proxy-Info按原顺序复制 | 完全相同；不能因是错误回答就漏Proxy或交换顺序 |
| 附加诊断 | 错误文本/报告来源/Failed-AVP按结果触发，不能以文本猜结果 | 同样有界；Failed-AVP内引述内容不作为应用载荷再次执行 |

所有接收输入按已有资源限制处理；超过MaxAnswerBytes不是普通字段错。正常成功载荷条件（如UDA数据、SRA身份/节点）仍按Application处理；头/结构/次数及请求关联已经无效时不交付伪成功对象。Sh/SMS认识AVP时的M位接收例外、SMS-VSAI及头保留位忽略规则继续有效。

### 24.2 结果选择：基础5xxx的E例外与补充实验结果

| 结果形状 | 本轮明确行为 / 输入 |
| --- | --- |
| E0基础2001，或所选正常应用失败 | 沿用原pair；N28九族N-NORESULT/DUAL/DUPRESULT分别验证缺失/两类冲突/重复。ALA E0用F28-ALA-N-5012，不新增实验5001业务映射 |
| E1基础3xxx | 必需E；九族E28-*-3002是授权DRA无法投递。正常完整UDA带3002但E=0仍违规，N28-UDA-N-3002不因普通载荷齐全而通过 |
| E1基础5xxx | RFC6733 §7.1.5允许无法或不便形成应用答案时用E语法；九族E28-*-5012证明不能仅因E1+5xxx拒绝。发送由既有具名FailurePolicy决定，接收不能臆测远端是否“足够高效” |
| E1基础2001或4xxx | N28-UDA-E-SUCCESS/TRANSIENT不得按合法E错误交付；4xxx的瞬时性不等于允许E位，也不授权业务重试 |
| E1只有实验结果 | N28-UDA-E-EXPERIMENT-ONLY缺必需268，ErrInvalidAnswer；不能把实验5001误作基础5001 |
| E1同时有268和297 | RFC错误语法明确有可选297，通用解码不能仅因并存就按正常XOR拒绝；基础结果决定协议失败，补充实验结果保留Vendor/Code及解释范围，不能覆盖为成功。标准完整ErrorAnswer构造当前只产生有依据的基础结果，不默认添加297 |
| E1补充未知实验Vendor | C28-UDA-E-EXTRA-UNKNOWN结构可表示，报告未知补充含义并保留原值；不是“已支持该Vendor应用失败” |
| E1补充已知Sh实验5102 | C28-UDA-E-EXTRA-3GPP是确定反例：29.329 §6.2明确要求使用其已定义实验结果时Result-Code缺省；E语法必需268，故此组合不能成为合法Sh错误。通用E结构可解析，固定Sh Profile结果约束失败；头已关联时ErrInvalidAnswer终结，保留有界原始诊断，不交付有效业务结果。不需要业务Context，也不是新的规范歧义 |
| E0未知实验Vendor/值 | C28-UDA-N-UNKNOWNEXP仅一个结果，不能误判缺结果；保留命名空间与未知语义。基础未知结果类别按RFC永久失败语义处理，不靠未知数字产生成功对象 |

以上E5权限只针对基础Result-Code，不是把Sh 510x、SMS555x等实验失败直接置E。合法错误只终结当前exchange；3002/3004的后续选择交由显式应用策略，不触发原身份传输恢复或隐式重投递。Ro的结果配对仍以其专门分析为准，不从本轮三应用正常XOR泛化。

### 24.3 关联与完成：先辨认候选，再借用预算

补明确既有R19/R21、VAL-03和SPEC-ERROR-01中的“可关联”：它不是“某个Session相同”，也不是“只命中HBH就消费pending”。不增加一个公共Correlation接口。

1. 已准入传输以固定有界头scratch验证可安全读取的版本/长度，R=1走请求路径；不得因与本地HBH相同当答案。头/帧边界不可信按既有关闭策略，不猜关联。
2. R=0用**实际ConnectionID + 本地方向HBH**找候选attempt，再比较保存的App、Command和E2E。此时只找候选，不移除pending、不提交完成、不转移其A预算。跨Peer/连接不按Session/E2E全局搜索。
3. 未命中，或App/Command/E2E不匹配，用已有MaxDiscardBytes/时限和固定scratch有界丢弃；超限关闭连接并按真实关闭原因结束受影响请求。丢弃本身不消耗其他pending的完成权或A。N28九族E-E2E及UDA E-HBH/APP/COMMAND是独立输入。
4. 完整头命中后，才使用该attempt预留的A读取有界完整答案。按固定R/E语法、数量/宽度/Session位置与原值、P/T、Origin授权以及适用Proxy回显关系检查；已关联的结构/来源/Exchange错误以ErrInvalidAnswer终结一次，不重试、不回答案。N28正常SESSION和E-NOSESSION属于这一类，而不是第3步。
5. Wire/Exchange通过后交付原结果和Application报告；无IMSI、Realm、时间未决或缺业务输入不等待另一答案。已完成/取消/超期的迟到重复只做有界丢弃；完成门与取消/关闭/其他attempt竞争仍产生一个Completion。

ConnectionID是本地来源事实，不是线路字段。旧代已完整读入并携带原AttemptID的合法候选，若其原匹配记录/截止仍满足R19，可在完成门竞争；不能改标新代，也不能仅因新连接建立就丢掉已合法读入的候选。新连接没有原attempt映射时，相同HBH不复活旧pending。这里澄清已有状态/预算归属，不扩大跨重启恢复或持久历史。

### 24.4 公共字段、错误来源和诊断隔离

| 项目 | 具体边界及证据 |
| --- | --- |
| Origin授权 | 成功来源查TerminalOrigins；具名E错误查对应Terminal/ErrorAgent授权。N28-UDA-N-AGENT场景明确dra-a仅有错误权限；N28-UDA-E-ORIGIN来源evil不在任何集合。配置可明确同一节点同时承担两角色，但不能由错误权限自动推导成功权限 |
| Error-Reporting-Host | 294用于诊断，不授权264、不新增Peer。RFC §7.1/7.4的条件是设置非成功结果者不同于Origin（3006例外按其规则），不能强制每个原端产生的失败都带294。接收方缺“谁改写结果”证据时不能臆断该项缺失；本库不是改写答案的代理产品 |
| Failed-AVP | 一个279父组可含多个错误AVP；不能用重复279模拟多错误。E28-SNA-5001与NOFAILED/DUPFAILED/EMPTYFAILED分别检查必需、数量、非空。父组/原bytes计资源预算，内部是证据而非第二份应用消息 |
| 非法内容在不同位置 | E28-PNA-5004引述原N27-PNR-PUBLIC-UTF8的700/601，不能因引述非法UTF8拒绝整个错误答案；N28-PNA-E-BADTEXT把非法UTF8放到容器外281则必须报告。恶意父边界仍按有界诊断读取，不执行子AVP业务规则/加载扩展 |
| Proxy-Info | V28-UDR-PROXY-R有两个直接父实例，V28-UDA-PROXY-A与E28-UDA-PROXY-A保留顺序和opaque State；N28-UDA-N/E-PROXYORDER各自字段合法但回显顺序错误。不把某DRA在远端腿插入又移除的Proxy误当本地请求快照的一部分 |
| Header flags | 发送从原请求保留P，明确清R/T；N28-UDA-E-T/P分别验证。C28-UDA-E-RESERVED接收忽略保留位，发送仍置0；C28-UDA-RFLAG必须分到请求入口，不结束出站UDR |
| 其余公共叶 | Origin-State、Supported-Features、Route-Record、Error-Message/Reporting-Host的逐命令宽度/重复/未知扩展仍有201项矩阵缺例；本轮没有因共享类型已有测试就标全覆盖 |

服务端Reply/Reject只使用已有Arrival的原关联/发送权与固定身份配置；不能让调用方把Destination/Session/HBH随意填入错误答案来绕过绑定。若原请求的必要Session/Proxy无法在拒绝预算内可信提取，仍关闭/终止，不生成猜测关联的错误。

### 24.5 规范勘误及其他库对照

本轮只采用与本切片直接相关的已核对勘误，不声称完整勘误审计：

| 来源 | 本轮结论 |
| --- | --- |
| [RFC6733 erratum4615](https://errata.rfc-editor.org/eid4615/)，Verified | 基础5001只含一个Failed-AVP父组；采纳至BASE-ERROR/Failed-AVP数量，多个被拒绝AVP放同一父组 |
| [erratum4887](https://errata.rfc-editor.org/eid4887/)，Verified | 答案除本地Origin-Host还需本地Origin-Realm；九族正常NOREALM及E代表输入覆盖，不复制请求目的地作身份 |
| [erratum6832](https://errata.rfc-editor.org/eid6832/)，本次仍Reported | 关于4002适用面，未采用；不将外部存储失败引入公开库，也不影响E4禁止规则 |
| [OTP diameter_app 2.7.2](https://www.erlang.org/doc/apps/diameter/diameter_app.html) | answer_message允许3xxx/5xxx且区分6733/3588字典；handle_answer保留原请求。参考语法分流，不移植任意结果注入API |
| [OTP answer_errors配置](https://www.erlang.org/doc/apps/diameter/diameter.html) | 解码错误可callback/report/discard；表明答案错误不能直接等同业务结果。本库沿用确定的ErrInvalidAnswer完成策略，不增加三种公共模式或以默认超时掩盖已匹配无效答案 |
| 本地go-diameter [message.go Answer](../../go-diameter/diam/message.go)，HEAD f26e3528006fcf0d33cb663994a0ec0ccbf7ef55；工作树SHA-256 `ad182fd58a6fdb4ea4b08e0999a6b405aa2cd9952a0d827d31cf8dcccdf22e21` | Answer复制Header时只清R，未主动清T/选择E；可选加基础结果，但不自动补Session/Origin/Proxy。是低层构造便利函数，不是完整合规答案builder；本库不可直接复制其行为作为Response契约。只读源码，未运行/修改该仓库 |

### 24.6 剩余P1：缩小范围，不以新报文代替验收

| 优先级 | 已明确 / 下一退出条件 |
| --- | --- |
| 首版P1 / 九族包络 | 正常/E基线、每族结果选择/关键字段/Session及E2E反例已补；还需独立解析器/对端复核，再在实现阶段验证一次完成、绑定和预算。正常/E类型检查不能只看Result-Code数字 |
| 首版P1 / 公共字段 | 201公共行仍无本表定向反例；优先重复父实例、未知M/扩展、Failed-AVP有界截断/深度、Origin/Proxy的跨角色条件。共享规则和逐命令检查结合，不要求复制201份等价帧 |
| 首版P1 / 结果分层验收 | 通用E可承载297、未知Vendor保留语义范围、已知Sh实验码受§6.2禁止与268并存，这三层设计已明确；下一步是独立解码/实现验收，不把已知禁止组合继续列作规范待定。其他未启用扩展按需，不能扩大为全Vendor研究门槛 |
| 首版P1 / 原协议缺口 | 真正无IMSI成功SRA、Realm标准组合、Requested=Maximum仍保留直接规范门槛；E语法不能绕过正常2001的身份/节点条件。SMEA数字格式已有依据，仍待独立解读和实际DRA/终端验证 |
| 实现授权后 / P1与按需P2 | codec/分阶段验证、Peer/发号/关闭、资源/SCTP尚未实施；16个未启用应用边界按实际需求推进，Ro纳入首版实施 |

R29/R34不整体关闭，24个SPEC ID不增删；本轮更新为Draft0.16。没有持久化、存储确认、业务重试、DRA代理服务或新通用facts接口。

### 24.7 本轮静态交付复核记录

2026-10-31复核15份文档、467个本地链接、717段message/AVP hex及24个稳定SPEC ID。新增148帧逐项检查长度、padding、AVP边界、摘要、原请求头/Session、九族正常成功/失败/E3002/E5012与各负例目标；额外核对Vendor命名空间、Proxy顺序、诊断引述原bytes和Sh5102禁止组合。回归既有30项精确Vendor/父路径控制；439行矩阵重新统计为263行在位候选、222行定向反例，新增53行、剩余201公共+16按需应用边界。

旧473份逐叶帧、96段早期hex、33条派生记录及19份业务源码摘要保持；go-diameter被引用文件摘要保持。本轮未修改XML载荷，未重新声明XSD执行结果。上述均为文档和字节静态检查，不是diameterkit validator、独立Diameter栈、网络、并发、预算或真实DRA/SCTP验证；运行验收继续Planned。

## 25. 第二十九轮：公共父实例、未知扩展、诊断边界和P1证据收敛

新增81份[完整输入§12](diameterkit-leaf-vectors.md)，覆盖18种正常命令的父实例、未知M处理、诊断错误pair及内外长度边界。旧473逐叶帧与148答案帧保持；正常矩阵新增39条原缺例行、72条在位候选，当前335行在位、261行定向反例，剩178=162公共+16按需应用边界。E/资源/未知项场景不计入439行分母。统计仍不是通过率；所有运行验收Planned。公开Go1.26无持久化端点库，Ro纳入首版实施。

### 25.1 父实例隔离：有独立输入，不以全消息搜索结果判次数

依据[RFC6733 §4.4及§6.7](https://www.rfc-editor.org/rfc/rfc6733.html#section-4.4)、[29.329 §6.3.11（转引29.229 §6.3.29；第三十一轮修正章节号）](https://www.etsi.org/deliver/etsi_ts/129300_129399/129329/18.00.00_60/ts_129329v180000p.pdf)及已选29.338命令语法。沿用固定Vendor:Code路径和直接父实例，不新增对象模型。

| 输入族 | 具体预期 / 不能据此声称的能力 |
| --- | --- |
| C29-18种命令-PARENTS | 每条附两Proxy父组和两SF父组。两个Proxy各一次Host/State，State分别为00ff41、8042；跨父同Code正常，不要求UTF8。每组SF各有266/629/630；新增Vendor424242、List7/8仅结构控制，不启用未知特性、不冒称这些私有键已获得全球分配 |
| N29-18种命令-PROXY-STATE | 第一Proxy缺33，第二组仍有33；对应L014逐命令反例。不能将第二组借给第一组；完整头已匹配的答案以ErrInvalidAnswer结束，请求按固定缺叶拒绝路径 |
| N29-18种命令-SF-ID | 末SF父内有两个629，其他SF正常；对应L017逐命令反例。未知Vendor含义不豁免已知628的结构；不能后写覆盖，也不能误判“只因另一父有629就重复” |
| N29-UDR-PARENT-HOST/VENDOR/BITS | 分别补284内280重复、末628内266缺失、末628内630缺失。仍按本父计数；这些代表例不自动覆盖所有18命令的其他叶 |
| N29-UDR-VENDOR-SHADOW | 260中的424242/266 M0不能满足0/266；允许保留未知项与必需叶缺失可以同时成立 |

C29正常答案均关联同族C29请求，复制本跳两Proxy原顺序，不是库充当代理。SF新增未知报价不改变原请求已选特性、原答案结果或角色前提。运行规则共享，但文档矩阵只计已经物化的命令/父路径证据；余下162公共行仍需按宽度、次数、位置与结果条件补齐。

### 25.2 未知扩展：描述符、M位、边界及完成后果分开

[RFC6733 §4.1/4.4](https://www.rfc-editor.org/rfc/rfc6733.html#section-4.1)不提供“Grouped标志位”；是否递归由固定字典的准确Vendor:Code描述符决定。本轮未知Code固定为424242/0x00ffff01和0x00ffff02，仅指测试Profile没有定义它们。所有受测未知输入仍占接收字节和owned预算。

| 场景 / 完整输入 | 接收结论 |
| --- | --- |
| UDR、正常UDA、E UDA顶层UNKNOWN-M0 | 外层头/长度有效后保留opaque并忽略其应用含义；负UTF8字节也不按文本检查。答案已有包络/结果/关联检查仍照常 |
| 同三位置UNKNOWN-M1 | 请求拒绝基础5001，E29-UDA-UNKNOWN-5001保留原AVP。两个答案输入若完整头关联，则ErrInvalidAnswer一次；不向答案再答，也不等待超时替代明确错误 |
| C29-UDR-OPAQUE-NEST | 未知M0载荷看似包含坏宽度278和未知M1；无描述符就不递归，不因看起来像Grouped而报内部错。只计一个可见AVP及其全部字节，不虚构不可识别子项数量 |
| N29-UDR-KNOWN-M0 | 已知Proxy父即使M0仍检查其成员，未知M1子项不能藏在其中；对应E29-UDA-NESTED-5001可含父路径。已知Sh/SMS字段的M位接收例外不扩为未知子项豁免 |
| N29-UDR-UNKNOWN-VSHORT | M0但V1、Length8不足12字节Vendor头：外层结构不成立，不能按unknown/opaque接受；先判头边界再读取Vendor，不切片越界 |
| N29-UDR-TWO-UNKNOWN / E29-UDA-TWO-UNKNOWN | 选择首个检测错误类别5001，单一279可携多个同类证据；不生成多个Result-Code。多种不同错误的优先选择沿用有界遍历的首个错误，不许无界收集全部错误 |

未知项在已知父内是否允许，还受该父语法扩展槽/位置限制；M0并不覆盖显式禁止位置。保留raw不是增加应用能力，也不触发插件加载、DRA路径发现、终端能力缓存或重试。

### 25.3 Failed-AVP：被引用的坏长度与诊断容器自身损坏不同

依据[RFC6733 §7.1.5、§7.5](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.5)及已采用erratum4615。本节细化§24的“有界解析”：**引述内部长度可能就是错误证据**；不能要求它像正常业务Grouped一样全部可递归解码。Message和顶层279自身必须有可信边界。

| 层次 / 输入 | 确定行为 |
| --- | --- |
| E29-UDA-LENGTH-SHORT/LONG | 各自关联N29-UDR-LENGTH-SHORT/LONG。279外层有效，内278声明Length4/0xffffff，实际只含8字节头+4字节零payload；保留该错误头，绝不依声明长度分配、跳读、整数回绕或跨279取字节。交付原基础5014和有界诊断，不重判为业务278坏包 |
| E29-UDA-MISSING-5005 | 原UDR缺277；诊断填预期头和4字节零payload。零只是缺失占位，不能套所选Auth-Session-State=1拒绝该错误答案；不从占位值重建请求 |
| N29-UDA-DIAG-OUTER-OVER/UNDER | 改的是279自身Length；它不能越过外层message或短于自身header。完整头已关联时ErrInvalidAnswer，若输送帧本身不可信则按原关闭策略；不能以“内部诊断可以坏”接受外围破坏 |
| N29-UDA-DIAG-TINY | 279内仅4字节Code；不满足本库5014构造时补齐最小header的要求。接收只保留不可进一步解读的短证据并报告诊断不完整；它不是本库标准完整builder正例，不能靠猜零/猜类型使取证视图看似完整 |
| C29-UDA-DIAG-DEPTH | 八层已知284用于已识别诊断深度压力，不能按普通Proxy缺Host/State报业务错。它不是规范构造器正例，也不证明原请求存在此路径；仅检验有界诊断读取和资源分类 |
| Failed-AVP内未知M1、UTF8或叶值错误 | 不作为当前答案的新应用失败；保留原bytes。容器外281/264/296等照常校验，沿用E28-PNA-5004与N28-PNA-E-BADTEXT的对照 |

读取步骤固定：验证外围长度及配额；把279内容作为拥有所有权的有界诊断；只有完整且边界可证的内层头才继续记录路径。碰到坏内层长度或短头就停止该诊断分支，保留剩余raw并报告无法继续定位，不重同步扫描、不把剩余bytes猜成兄弟AVP。279后的顶层字段从可信279外层末尾继续检查。已知可展开节点仍计MaxAVPs/MaxDepth；未知opaque不展开。日志/格式化也受同一限额，不在String/日志路径做另一轮无界递归。

同一输入的预算对照（数值只是fixture前提，不是默认配置）：

| 预算场景 | 预期 |
| --- | --- |
| C29-UDA-DIAG-DEPTH：279内容80octets，MaxFieldBytes=80且其他额度足够 / =79 | 前者能保留原诊断；后者ResourceLimit。恰等上限合法；不把本地上限报成5014/5009，也不截断成“完整”诊断 |
| C29-UDA-DIAG-DEPTH：根消息深度0、279为1、八层284最深9；MaxDepth=9 / 8，其他足够 | 前者有界读取，后者ResourceLimit并停止；不是普通Proxy缺叶。叶本身不增加Grouped深度 |
| C29-UDR-PARENTS：重复集合限额允许2个Proxy / 只允许1 | 前者结构可读，后者本地ResourceLimit；标准`*`不等于内存无限，不能将配置上限伪作协议5009 |
| 请求拒绝预算不足以容纳Session/Proxy、单个错误证据及必须包络 | 沿用TransportReserve失败/关闭路径，不截断必要字段、不另建紧急无界队列；能按规范使用最短缺失/坏长度诊断时无需复制整份请求 |

没有新增公开诊断遍历API或容器级预算。诊断错误不影响关联的基础失败类别，但真正的外层坏包和资源耗尽仍有各自终止结果；不得把所有诊断问题合并成ErrInvalidAnswer或全部忽略。

### 25.4 三项P1：从重复查字典转为有条件的证据清单

固定[29.338 V18.2.0](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf)，本轮复看[19.3.0对照件](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/19.03.00_60/ts_129338v190300p.pdf)与[3GPP CR登记表](https://portal.3gpp.org/ChangeRequests.aspx?q=1&specnumber=29.338)。未升级基线，也不声称穷尽全部CR正文。最新登记表出现新的具体线索，不能只沿用上一轮“未发现”措辞。

| P1 | 本轮增量 / 固定行为 | 下一可关闭证据 |
| --- | --- | --- |
| 无IMSI成功SRA | 表5.2.1.1/2继续分别约束真实IMSI、特定Router token、仅MCCMNC用途dummy。CR0050原版对应HSS-ID类型修正已纳入18.2.0，登记表的rev1反而withdrawn；不能用“修订号更高”替代已批准版本。标题/类型修正均不授权普通成功SRA回显HSS-ID | 需要明确修改该成功身份条件的条款/CR正文及批准/纳入状态。现有N22/V21身份场景继续验收；仅“请求方不知道IMSI”不是“目标确无IMSI”。数字同形仅凭bytes不可判；库不查询用户库 |
| IP-SM-GW-Realm | 新定位**CR0069，C4-264035，IP-SM-GW address in SRA，类别F，受影响19.3.0、目标Rel20、CT4#137工作组agreed**；当前门户未列TSG批准或新版本。只凭标题不能断言它修正Realm、取消Number或扩Additional。默认标准域/既有三元组兼容保持§23.4 | 优先取得0069正文，核对修改段落、g/h/i组合、Realm/Number及接口范围，再查TSG批准/最终版本/Rel18回移植。独立对端通过只关闭具名兼容；未批准的Rel20候选不自动进入Rel18标准域 |
| Requested=Maximum | Rel18/所查Rel19的上限表与更严格过程措辞仍不一致。登记表把eDRX源头定位到CR0017 rev3、0018 rev1（13.0.0），及0019提醒/0021补TFA字段；条目标题不能裁决相等。旧历史表与登记表分类等元数据不能代替正文 | 定向追踪eDRX相关正文或后续澄清，找明确边界语义。标准构造仍取严格早于；接收相等保留5550/Time和范围报告。era缺失、原3330缺失、>上限先各自报告，不被“未决”掩盖 |

本轮C4-264035下载页可访问并指向CT4_137_Prague-2026-10会议ZIP，但公开下载端返回403，未取得正文；因此只登记可核对元数据，不给正文摘要或SHA，不宣布标准矛盾关闭。来源入口为[C4-264035](https://portal.3gpp.org/ngppapp/DownloadTDoc.aspx?contributionUid=C4-264035)。没有因此请求业务账号、增加自动追踪任务或暂停其他文档工作。

三项均为**具体能力**的标准证据P1，不阻塞普通真实IMSI、已明确无Realm组合、严格早于子集的实现设计；库本地行为已定。未决不是调用方提供任意facts就能解除，不增加no-IMSI开关、相等时间兼容或宽松Realm排列组合。SMEA数字子集仍缺独立解读/真实对端证据，Ro争议按首版受限分支处理。

### 25.5 其他库对照与剩余优先级

| 来源 | 本轮读到的行为 / 取舍 |
| --- | --- |
| [OTP diameter_dict 2.7.2](https://www.erlang.org/doc/apps/diameter/diameter_dict.html) | messages、Grouped和成员次数分别描述；借鉴逐父结构，不让全消息字段搜索替代父实例检查 |
| [OTP diameter_codec 2.7.2](https://www.erlang.org/doc/apps/diameter/diameter_codec.html) | 文档明确解码产物存在不表示没有errors；本库同样区分原字节、可读字段与错误报告，不把结构控制当全部语义Passed。未运行OTP，不据文档宣称其坏长度诊断行为与本库相同 |
| 本地go-diameter [avp.go](../../go-diameter/diam/avp.go)，HEAD f26e3528006fcf0d33cb663994a0ec0ccbf7ef55；SHA-256 `8f6568755364bac33ccfbb51b8ffa7d6959357b2679808751baa41502f6caea2` | DecodeFromBytes查字典失败直接返回，已知Grouped进入递归；该低层路径未见Failed-AVP特殊上下文。V1读取Vendor前只验8字节最低头，也提示必须独立验证12字节界限；仅源码审查，不报告已运行的漏洞复现 |
| [group.go](../../go-diameter/diam/group.go)，SHA-256 `2f172e3d6ed88b915f0ec3d34a95f5516d6e7785cee5a0ed829a0c89392d457c` | DecodeGrouped对子项调用普通DecodeAVP，子错误即返回；本库需要诊断上下文及有界raw退路，不能把同一递归直接用于279。未修改/运行该库，不泛化为所有API均不支持未知项 |

优先顺序：①公共诊断/opaque边界的独立解读与实现时的内存/深度验证；②剩162公共缺例按命令规则和共享类型补齐，优先Origin-State、Route-Record、结果父组/错误文本及数量边界；③先取得CR0069正文，再跟踪无IMSI与eDRX精确条款；④SMEA和Realm具名兼容独立对端证据。16个按需应用边界不扩首版，codec/Peer/标识/预算/SCTP实施需后续用户授权。R29/R34保持P1，24个SPEC ID不变；本轮Draft0.17，不引入持久化、订阅表、业务恢复或DRA代理功能。


### 25.6 本轮静态交付复核

2026-10-31：15份文档、497个本地链接、798段message/AVP hex和24个稳定SPEC ID复核通过。81新增帧的Message长度/摘要、18命令父实例和答案关联、精确Vendor/M位、诊断引用原bytes及声明长度均已核对；5份故意破坏外层AVP边界的输入按其目标确认失败，不伪称全部hex都能常规解析。诊断12octets坏长度引用、80octets深度载荷及九层Grouped的预算场景尺寸已确认，尚未运行真实资源限额逻辑。

439行矩阵重新统计为335行在位、261行定向反例；相较上轮新增72在位/39定向，剩162公共+16按需应用边界。旧473逐叶帧、148答案帧、96段早期hex及33条派生记录保持；30项旧Vendor/父路径静态控制复跑，19份业务源码和本轮两份参考库源码摘要一致。未改XML，不重新声明XSD执行结果；临时检查器只在/tmp，仓库只改文档，未增加实现/测试代码或依赖，未修改业务或SCTP仓库。独立Diameter栈、网络互通、Peer竞态、实际内存/性能与SCTP验收继续Planned。

## 26. 第三十轮：Origin-State、Route-Record、结果父组与错误文本

本轮新增145份[完整输入§13](diameterkit-leaf-vectors.md)：18种正常命令的状态/路由控制与反例、九答案族的错误文本/报告身份及实验结果父组，另补共享边界和E语法代表例。旧554逐叶帧、148答案帧及96段早期hex保持；全doc943段hex。正常矩阵新增58条原缺例行和54条在位候选，当前389行在位、319行定向反例，剩120=104公共+16按需应用边界。字段存在、语法控制、应用符合性与独立互操作仍是不同证据。

仅更新文档，基准Go1.26；公开、无持久化端点库，Ro纳入首版实施。没有实现状态缓存、重试调度、代理转发、持久计数或新通用facts。R29/R34未整体关闭，24个SPEC ID不变，specification升为Draft0.18。

### 26.1 Origin-State-Id：属于Origin，不属于请求回显或连接代次

依据[RFC6733 §8.16](https://www.rfc-editor.org/rfc/rfc6733.html#section-8.16)。既有Peer/标识§4.5的默认省略策略保持；非零重启单调保证由显式配置该值的部署负责，库不声称凭随机数、当前秒或进程计数即可保证跨重启单调。

| 场景 / 输入 | 行为与范围 |
| --- | --- |
| C30-18命令-STATE-ROUTE | 请求278=1、答案278=2；两者分别对应各自Origin-Host的已配置状态。不能复制请求值填到本地Origin的答案；收到不同值不是关联失败 |
| N30-18命令-STATE3/STATE-DUP；N30-UDR-STATE5 | U32恰4字节、顶层0..1；3/5字节及重复同值均为静态错误。已关联答案ErrInvalidAnswer，请求按固定拒绝路径；不能取最大值/最后值“纠错” |
| C30-UDR-STATE-ZERO/MAX | 0和0xffffffff均有在位U32字节；0不作旧会话失效推断，max不是Time，不用era解释。max后的非零单调生成能力不由该样本承诺 |
| C30-UDA-NO-STATE-ROUTE | 原请求有278，答案仍可省略；它不是Session/Proxy那样的回显字段。默认不配置就不发；显式0与缺省在线路和owned对象中保留Presence差异 |
| 同一DRA连接先收到HSS-A/278=10，再收到HSS-B/278=2 | Origin不同，不能比较成“远端回退/重启”；DRA CER的278仅描述DRA，不为后面的HSS代言 |
| 已绑定同Origin的值变化、迟到低值或0 | 可保留当前消息证据/既有有界事件；不凭单包建立无限Origin历史表、终结其他pending、删除业务订阅或更新业务会话。Peer/Transport真实故障仍按原状态机处理 |

重复键仍是接收域+Origin+E2E；Origin-State不进入键来拆分一次业务执行，但它的Presence/值继续进入Canonical内容比较。同一键下缺省→0或1→2是内容冲突，不能因“0与省略都不推断会话失效”就把原始内容合并。新状态也不是重用E2E的许可。

### 26.2 Route-Record：保序与路径准入，不是Proxy回显或新路由授权

依据[RFC6733 §6.1.3/6.1.9、§6.2、§6.7.1](https://www.rfc-editor.org/rfc/rfc6733.html#section-6.7.1)。所选三应用正常CCF列出重复282；本库仍是消费请求的端点，不新增代理循环检测/转发产品。

| 输入 / 检查 | 本轮确定行为 |
| --- | --- |
| C30-18命令-STATE-ROUTE | 每个请求有两个顺序明确的282，正常答案有不同的一个282；只证明接收结构/保留。不能要求答案282与请求相等，不能默认照抄整份路由历史；按具体命令语法接受在位项 |
| N30-18命令-ROUTE-TEXT | 首项含空格，第二项正常也不能抵消；逐实例检查DiameterIdentity，不把282当任意UTF8说明文字 |
| C30-UDR-ROUTE-REPEAT/REORDER | 重复远端名称不是单值次数违规；重排必须在raw视图中可见。是否实际存在转发环路另取决于代理角色/本地身份，不据“列表有重复”自动返回3005 |
| 请求未经过代理或未携282 | 不为了填字段而加入本地Origin/DRA名；字节可选不构成可信路径证明。已有部署入站准入仍需满足，不能因字段类型通过就认为路径可授权 |
| 282提到一个未授权节点 | 不新增Peer、终端来源或重试目的地，不从字段名反推相邻CER关系。若部署已有路径约束，由既有准入机制检查；不让解析器默认相信自述路径 |

RFC的代理加282和逐跳loop规则不迁成库端点的自动转发逻辑。接收方不拥有每个历史hop的CER，不能逐项强求与当前邻接DRA的Origin一致。原字节保序与Canonical规则不冲突：**顶层Vendor0/282仍按R17逐跳字段剥离**，因此仅改变这些字段的重复请求可以在独立准入/校验通过后合并；嵌套或不同Vendor同Code不剥离。Proxy-Info的本跳应答回显继续单独检查。

### 26.3 结果父组：结构先于结果解释，正常与E分母仍分开

依据[RFC6733 §7.6/7.7](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.6)及所选应用规则。九族正常结果选择沿用§24；“父组存在”不代表其两个直接子项已满足要求。

| 输入族 | 预期 |
| --- | --- |
| N30-九族-EXP-NOVENDOR / EXP-CODE3 | 297内分别缺0/266、0/298宽度3；缺命名空间与未知合法命名空间不同。别的VSAI/SF中的266不能补；完整头已关联后静态失败终结 |
| N30-九族-EXP-SPLIT | 两个297分别只有Vendor和Code，是明确复合负例：父次数与各父必需叶均错误。不能扁平搜索后拼成一个结果，不指定运行必须优先报告哪一项 |
| N30-UDA-EXP-EMPTY/DUPPARENT/DUPVENDOR/DUPCODE/VENDOR3 | 共享代表输入分别覆盖空父、双父、同父重复Vendor/Code、Vendor宽度3。完全重复也不能后写覆盖；这些代表例不自动给其他命令补全所有错误维度 |
| C30-ALA-UNKNOWN-RESULT | 用Vendor424242/5999替换成功基础结果，只是未知语义接收控制；没有创建ALA的3GPP/5001标准分支，也不把未知结果视为成功。L009/L010的在位证据是结构范围 |
| N30-UDA-E-EXP-NOVENDOR | E1仍有基础5012，但附加297缺266；通用E允许可选297不豁免已知Grouped结构。区别于完整未知Vendor补充结果，也区别于§24已禁止的Sh实验码+268组合 |

消息头、次数、宽度、父实例先检查，随后才解释Vendor/Code与应用条件。已关联静态失败不向答案再发错误；结构有效而语义未支持按原结果+范围报告交付，不再等待另一个答案。没有结果插件或任意Result-Code注入API。

### 26.4 Error-Message与Error-Reporting-Host：UTF8限制、表示与语义分开

依据[RFC6733 §4.3.1、§7.3/7.4](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.3)及其引用的[RFC3629](https://www.rfc-editor.org/rfc/rfc3629.html)。UTF8String不等于“任意Go string”：除了编码有效，RFC6733排除U+0000；其历史上界不能被用来接受RFC3629已禁止的超长序列、surrogate或大于U+10FFFF值。这里只收敛基本类型，不引入完整国际化helper。

| 输入 / 边界 | 结果 |
| --- | --- |
| C30-九族-ERROR-TEXT | 原应用失败码不变；281含中文和CR/LF，294是接收的诊断Host。控制只证明字段可表示，不能证明其自述改写来源；本地失败默认使用本地Origin，无需伪造另一个reporter |
| N30-九族-TEXT-UTF8 / REPORTER-TEXT | 281为c328非法UTF8、294含空格非法Identity；两者在Failed-AVP容器外，完整关联后ErrInvalidAnswer，不因“只是日志”忽略类型错误 |
| N30-UDA-TEXT-DUP/REPORTER-DUP | 各0..1字段重复同值仍错误；不串接多值形成新的错误文案，也不以第二个报告身份覆盖第一个 |
| C30-UDA-TEXT-EMPTY | 空字符串在位，长度0不等于缺结果；无需凭文案猜错误码。空文本/字段缺省在raw视图仍不同 |
| N30-UDA-TEXT-NUL | NUL可以由通用UTF8编码，但不是Diameter UTF8String允许的码点；静态违规。不能用C字符串截断或替换字符使输入通过；有界原始证据可保存 |
| C30-UDA-E-TEXT / N30-UDA-E-TEXT | E1同样遵守外围281类型检查；不能用“E错误”或“诊断例外”放行非法文字 |

控制码/前后空白在规范中另有SHOULD约束，不能全部升级为接收MUST拒绝。默认构造文案避免控制字符，需要换行用CRLF；显示和日志使用有界转义，不能直接把对端文本当format string、终端指令或结构化日志边界。此显示转换不修改原字段，也不自动Unicode正规化。字段/owned预算按UTF8**字节**计，不按汉字数量；超本地额度为ResourceLimit，不伪造线路结果。

281中的`retry=false`或任何自然语言不是协议调度输入。294既不替代Origin授权，也不授予DRA成功回答权限；接收方缺“谁设置结果”的证据时，不把所有失败统一要求294。本轮不改§24的条件规则。Failed-AVP中引用原非法文本继续遵守诊断例外；不能把U+0000限制扩散到OctetString的二进制Proxy-State、TPDU或XML原bytes类型。

### 26.5 CR0069正文与无IMSI、相等时间：本轮取得和未取得的证据

| 问题 | 本轮核对 | 当前结论 / 下一证据 |
| --- | --- | --- |
| CR0069 / C4-264035 | 再查[官方登记表](https://portal.3gpp.org/ChangeRequests.aspx?q=1&specnumber=29.338)：仍见CT4#137 agreed、目标Rel20，未列TSG批准/新版本；重新请求[官方会议ZIP](https://www.3gpp.org/ftp/tsg_ct/WG4_protocollars_ex-CN4/CT4_137_Prague-2026-10/Docs/C4-264035.zip)仍HTTP403，正文未取得 | 不能把本轮写成“正文确认Realm修正”。不从标题推断g/h/i或Number存在条件。保留既定Rel18标准域和s6c-ip-sm-gw-realm-v1；下一步只追这份正文及批准/回移植证据，无需重复泛查同一字典 |
| 无IMSI成功SRA | 复查[23.204 V18.0.0 §6.0a.2、§6.1步骤4、§6.4步骤3b/3c](https://www.etsi.org/deliver/etsi_ts/123200_123299/123204/18.00.00_60/ts_123204v180000p.pdf)：MSISDN-less过程仍描述取得IMSI，MSISDN-less correlation包含HLR标识不等于它成为成功SRA的IMSI | 这组Stage2证据支持“无MSISDN不能推导无IMSI”的区分，并未解除[29.338 V18.2.0表5.2.1.1/2](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf)的成功身份条件。真正无IMSI普通成功仍未支持；请求方未知/服务端已知、Router token与仅MCCMNC用途仍按既有分支。不引入身份查询 |
| Requested=Maximum | 新增交叉检查[29.118 V18.0.0 §5.1.3.1、§8.21.1、§9.4.32/33](https://www.etsi.org/deliver/etsi_ts/129100_129199/129118/18.00.00_60/ts_129118v180000p.pdf)：SGs过程提供原Maximum依赖和两个时间IE含义/编码引用，未在所查条款给出可覆盖SGd相等问题的明确裁决 | SGs接口不能替代29.338表6.2.2.1/2与§6.2.2.2的条件。继续严格早于构造子集；接收相等带原5550/时间及范围报告，era缺失与原上限缺失分别处理。下一证据仍是eDRX修订正文/直接澄清，不增加SGs/MAP栈、调度器或相等兼容开关 |

29.118这里只作跨接口语义对照，不增加首版字段依赖、解析模型或规范实现声明；本轮通过官方网页读取，未新增本地原件摘要。以上能力P1不阻塞已明确普通分支，也不因无法下载一份文档而停止公共字段补齐。Ro全部纳入首版对应验收。

### 26.6 其他库参考与剩余优先级

| 来源 / 固定证据 | 观察与取舍 |
| --- | --- |
| [OTP diameter 2.7.2的Origin-State配置](https://www.erlang.org/doc/apps/diameter/diameter.html) | 可选本地值用于控制消息，提供启动时刻派生函数；可参考可选配置，但不能当跨重启单调证明。本库raw保留显式0/缺省差异，不增加持久provider |
| 本地go-diameter [sm/cer.go](../../go-diameter/diam/sm/cer.go)，SHA-256 `31ff40b12553f8ade638073370c84a1b7600fc1ea4200aaf4f50fc8dac391513`；[sm/sm.go](../../go-diameter/diam/sm/sm.go)，`71a488a6aa088dd37025428879a9d31442192f32698e51c65e265c744e02d338`；HEAD f26e3528006fcf0d33cb663994a0ec0ccbf7ef55 | CEA路径写本地Origin但复制CER中的OriginStateID，Settings注释亦说明此行为。本库不能照搬，应使用本地Origin所属配置或省略；不是需要库持久化的新理由。只读源码，未运行/修改 |
| [datatype/utf8string.go](../../go-diameter/diam/datatype/utf8string.go)，SHA-256 `6f8e996145b5d1b272e140c9e205b7c55e7bc35da9817d29439619baef16379a` | DecodeUTF8String直接包装原bytes且返回nil错误；不能据其解码成功证明UTF8或Diameter码点合法。String原文格式化也不等于安全的日志边界。本库在字段校验和有界显示分别处理，不修改该库 |

剩余104公共缺例为：L013 Proxy-Host、L016 SF Vendor、L018 SF bits各16；L001 Session、L003 Origin-Host、L004 Origin-Realm、L006 Destination-Host、L007 Destination-Realm各8；L002 Auth-Session-State为7；L011 Sh VSAI Vendor为6；L012 Sh VSAI Auth-App为3。下一轮优先补这些必需/次数与目的绑定反例，沿用共享类型检查加逐命令证据，不用同一帧附带的无关叶虚增覆盖。父组复合反例的首错报告顺序不冻结为公共API。

独立解析/对端、实际codec与内存/日志预算验证仍P1；Origin-State/CER本地归属应纳入Peer实施验收。16项未启用应用边界按需，SMEA/Realm具名兼容仍缺独立互通。无需为本轮新建持久状态、代理模式或业务恢复服务。

### 26.7 本轮静态交付复核

2026-10-31：15份文档、526个本地链接、943段message/AVP hex和24个稳定SPEC ID复核通过。145份新增帧的Message/TLV边界、长度、SHA-256、原请求头及Session关联已核对；18命令的状态宽度/次数与路由实例、九答案族的文字/报告身份及结果父组均检查了实际bytes。U+0000与非法UTF8分别验证；零值、max、缺省、空文本、重复/分裂父组及E外围字段控制与表述一致。这里的通过仅指静态交付检查，不是实现的接收结果。

439行正常矩阵重算为389行在位、319行定向反例，本轮增加54条在位/58条原缺例行，余104公共+16按需应用边界。旧554逐叶帧、148答案帧、96段早期hex及33条派生记录保持；第二十九轮81帧（含5份故意破坏外层边界的输入）和更早30项Vendor/父实例检查复跑。19份业务源码及5份参考库源码摘要一致；未修改XML，不重新声明XSD运行结果。

临时生成/检查器仅在/tmp，仓库只更新文档；没有新增实现/测试代码或依赖，没有修改业务、go-diameter或SCTP仓库。尚未执行独立Diameter栈、真实DRA对端、Peer并发、实际预算/日志或性能验收，运行状态继续Planned。CR0069正文仍未取得，不计入已读规范正文。

## 27. 第三十一轮：公共父叶、请求必需项与目的绑定

新增214份[完整输入§14](diameterkit-leaf-vectors.md)，补齐此前104条公共叶缺例；[正常矩阵](diameterkit-command-path-matrix.md)当前389行在位、423/439行有定向反例，剩16条按需应用边界。L001–L020每个适用命令/路径至少有一条反例，不等于每种宽度、次数、编码、角色、错误答案和资源维度都已覆盖；也不表示首版已通过符合性或互通验证。逐叶累计913帧，全doc1157段hex；旧报文不改字节。

Go1.26公开、无持久化端点库，Ro纳入首版实施。没有业务订阅表、目的发现、自动代理、持久特性缓存或新facts/provider。Specification Draft0.19细化原需求，24个SPEC ID不变；R29/R34仍是实现和独立验证P1。

### 27.1 Proxy-Host与Supported-Features：从父组定位到直接子项

规范入口为[RFC6733 §6.7.2–4](https://www.rfc-editor.org/rfc/rfc6733.html#section-6.7.2)、[29.329 V18.0.0 §6.3.11–13](https://www.etsi.org/deliver/etsi_ts/129300_129399/129329/18.00.00_60/ts_129329v180000p.pdf)所引用的[29.229 V18.0.0 §6.3.29–31](https://www.etsi.org/deliver/etsi_ts/129200_129299/129229/18.00.00_60/ts_129229v180000p.pdf)。修正§25.1把Supported-Features依据误写为29.329 §6.3.2的问题：那一条定义MSISDN，不是特性父组。没有升级修订或引入Cx/Dx支持。

| 完整输入 | 固定目标与范围 |
| --- | --- |
| N31-18命令-PROXY-HOST-MISSING/TEXT | 基于C29同命令双Proxy：只移除第一父0/280，或将其设为含空格Identity；第二父仍完整。不能全树查到一个Host就通过，也不能依据该Host授权业务Origin；原33仍是opaque |
| N31-18命令-SF-VENDOR-MISSING | 仅末10415/628父缺0/266；此前父及Sh VSAI的Vendor存在不补本父。未知Vendor/List含义与已知628结构分开 |
| N31-18命令-SF-BITS3 | 末父10415/630仅3字节，其他组完整；不能把短值填零或借另一组位图。零位图需实际4字节 |
| N31-UDR-SF-VENDOR3/VENDOR-DUP/BITS-MISSING/BITS-DUP/ID-MISSING | 共享代表宽度、直接子项缺失及重复。仅UDR增加这些维度，不假称18命令已逐一覆盖同样维度 |
| N31-UDR-SF-VENDOR-SHADOW | 末父只有424242/266 M0，不能满足0/266；保留未知项不等于满足必需叶 |
| C31-UDR-SF-ZERO | 未知Vendor424242/List8的完整零位图，仅结构控制；不得用其位0推导Notif-Eff，不当成已知3GPP终端明确不支持，也不等于缺628/630 |

继承C29多父控制及第二十五轮已知Sh支持集控制：解析键是Header App+父内Vendor值+List-ID，AVP头的Vendor不是父内Vendor-Id的替代。不同列表不能混合，同键冲突不能OR或后写覆盖；本轮不重新定义重复键的线路错误。Sh特性使用、完整报价、M位与5011条件仍按§21，不把SMS的未知列表套入Sh协商，也不把DRA CER视为终端广告。对静态无效的完整关联答案仍ErrInvalidAnswer；请求能否形成合法拒绝另见§27.4。

### 27.2 请求必需项与Sh应用父组：先校验，再绑定，不靠补值修复接收

复核所选29.329 §6.1与[29.338 V18.2.0 §5.3.2、§6.3.2](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf)的CCF。九请求UDR/PUR/SNR/PNR/SRR/RDR/ALR/OFR/TFR均要求Session、Auth-Session-State、Origin-Host/Realm和Destination-Realm；PNR/TFR的Destination-Host为必需，其余七种在CCF为可选，具体过程与部署可以收紧。

| 输入族 | 预期与不能推广的结论 |
| --- | --- |
| N31-九请求-NO-SESSION/ORIGIN-HOST/ORIGIN-REALM/AUTH-STATE/DEST-REALM | 每帧只删除一个顶层必需字段；CER的邻接身份、其他AVP或本地默认值不能修复接收消息。不能把缺Session变成新Session后继续执行业务 |
| N31-九请求-DEST-HOST-DUP | 必需/可选均最多一次；重复同值仍不通过，不能以最后值选择目标 |
| N31-PNR/TFR-NO-DEST-HOST | 违反命令必需项，尚未到realm-only准入；允许realm-only策略也不能覆盖CCF |
| C31-UDR/PUR/SNR/SRR/RDR/OFR-NO-DEST-HOST | 仅结构允许，角色/过程仍沿用原请求；在明确允许realm-only且目标realm本地服务的策略下进入后续校验，策略禁止时相同bytes拒绝 |
| N31-ALR-NO-DEST-HOST | ALR的CCF虽可选293，本HSS→IWMSC基线却受29.338 §5.3.1.2约束，需要Host和Realm。缺Host为过程负例，不能因realm-only准入通过就作为合法应用请求；“否则走MAP”是应用职责，不让库提供MAP回落 |
| N31-八种Sh正常命令-VSAI-NOVENDOR/NOAUTHAPP/WRONGAPP | 260内必需叶缺失或258与Header不符；不从Header补齐缺叶，也不能由错误258改变固定decoder/pending。正常答案沿用各自原请求，静态失败终结一次 |
| N31-UDR-VSAI-VENDOR-VALUE/AUTHAPP3/NOVSAI | 分别为非3GPP Vendor、258宽度错误、整个必需父组缺失；纯父组缺失不重复给两叶计数。SMS-VSAI忽略例外不迁移到Sh |
| N31-UDR-SESSION-DUP/NUL/POS、ORIGIN-HOST-DUP、ORIGIN-REALM-DUP、DEST-REALM-DUP | 重复、Diameter UTF8String禁用码点或Session非首项分别校验；不先排序、裁剪或补默认。保持有界原bytes用于诊断 |
| N31-UDR-AUTH-STATE0 | 基础枚举0存在，但29.329 §5转引29.229 §5.3要求无状态值1；不得把“子集不允许”写成“基础协议未定义”。不为该值建立持久会话管理 |

出站CompleteRequest/运行时注入Origin、默认Destination是已有构造流程；接收校验不能复用这种补值行为。SMS命令收到260的例外继续遵从§5.3.2.2；本轮Sh260反例不扩大到SMS。E错误仍走独立语法，不能强迫E3002具备Sh正常260/277或应用成功载荷。

### 27.3 目的绑定：合法字段、准入策略和完整错误pair分别验收

[RFC6733 §6.1.4](https://www.rfc-editor.org/rfc/rfc6733.html#section-6.1.4)规定本地消费条件；本项目既定LocalDestinationPolicy/DestinationPolicy另外约束部署允许的host/realm配对。**本地Host匹配但realm不在许可配对**属于项目准入拒绝，不能写成RFC定义的坏Identity。目的判断不建立DNS路由器，也不从用户身份/IMSI推断realm。

每族以原请求的Host/Realm为一组许可配对，再显式加入`alternate.example / alternate.example.net`；`outside.example`和`outside.example.net`未授权。原正常答案的Origin代表本地身份配置；SourcePolicy允许原请求来源，邻接admission已通过，预算充足。所有变体独立重置，不能把同E2E变体依次投入缓存后混淆目的准入与内容冲突。

| 场景 | 入站 / 出站结果 |
| --- | --- |
| N31-九请求-DEST-FOREIGN-HOST | Host语法合法但不是本地目标；入站3002/E1，不调用业务、不转发。出站DestinationNotAllowed，无发送、无Completion；不是先提交再异步失败 |
| N31-九请求-DEST-FOREIGN-REALM | Realm语法合法但未授权，按同一既定准入路径；不伪造5004或自动改成DRA realm |
| N31-九请求-DEST-CROSS-PAIR | Host与realm分别出现在许可集合，交叉配对却未授权；不能把允许Host集合与realm集合做任意组合 |
| E31-九答案族-DEST-3002 | 完整关联FOREIGN-HOST反例，使用本地配置Origin和原Session/HBH/E2E；P保留、R/T清零、E1、基础3002，不附正常应用载荷。不从错误Destination复制Origin；出站拒绝不生成该帧 |
| C31-UDR-DEST-CASE | Identity按既定ASCII大小写无关比较通过，raw仍不同；不顺带接受尾点、Unicode归一、IDNA或DNS别名推导 |
| C31-UDR-DEST-ALIAS | 显式配置(hss-alias.example,example.net)才可用；撤去别名的相同bytes为准入反例。别名不能从相邻CER、Route-Record、Proxy-Host或SRA动态加入 |
| 六个C31-NO-DEST-HOST条件控制 | 允许本地realm-only与禁止realm-only两种固定策略分别验收；不因无Host跳过准入。PNR/TFR为CCF必需，ALR本HSS场景为过程必需，均不得放宽 |
| 同一重复键在两条DRA路径先后到达 | 每次Arrival独立执行Source/目的准入，后做R17内容比较；缓存命中不能绕过准入。改目的仍不在Canonical剥离表，不靠缓存复用吞掉DRA改写差异 |

3002只是现有端点拒绝映射；不新增3003/重定向/路由发现策略。异步来源或目的字段不改变预绑定的App/Command/Profile。部署别名用于准入，不表示它们在R17内容比较中自动合并；除已有Identity大小写折叠外，不能因两个Host同属本地配置就把不同请求内容当相同。答案来源仍按已保存TerminalOrigins/ErrorAgentOrigins验证，不从Destination别名推导所有成功来源授权。

### 27.4 缺叶诊断与可发送错误包络：不把每个负例都算成一份有效5005

沿用[RFC6733 §6.2、§7.2/7.5](https://www.rfc-editor.org/rfc/rfc6733.html#section-7.2)、§24.3/25.3和现有拒绝预算。以下明确区分静态检测结果与传输结果，不增错误API或保存协议状态。

| 原请求问题 | 本库路径与验证边界 |
| --- | --- |
| 277、Origin、Destination-Realm或SF子项缺失，原Session/Proxy均可信 | 进入已有缺叶拒绝路径；基础5005及有界Failed-AVP缺失占位，必要时带父路径。Origin来自本地配置；不需要用坏Origin查询用户库。完整5005成对输出还需按错误族继续补证据 |
| Session完全缺失、其余边界可信 | 可确认“原请求没有263”不同于“有263但不能可信提取”；基础E错误允许省略原本不存在的Session，诊断记录缺失，不生成新Session。不能套正常Sh/SMS必需Session而断言根本无法回复；接收方若此前确有pending且发出的请求含Session，则无Session错误答案仍按该原请求关联规则失败 |
| Session重复/非法，或Proxy中必需Host缺失/非法 | 静态负例成立，但当前快照无法同时满足合法回显与标准错误包络时，沿用不可安全构造拒绝的关闭/终止路径；不能丢掉一组、猜Host/Session、把坏Proxy修好或原样回显后宣称标准有效。此为本库保守策略，不声称RFC要求所有实现关闭；真实对端可接受行为仍属P1 |
| 头/外层边界不可信或拒绝额度不足 | 既定关闭/终止，不新建无界紧急队列，也不尝试多份错误回答 |
| 完整关联答案含上述父叶错误 | ErrInvalidAnswer一次完成，无再答、无自动重传；合法错误结果与结构非法答案不可混同 |

对§25.1“请求按固定缺叶拒绝路径”的补充是：进入拒绝路径不保证一定能发出答案，仍受原Session/Proxy可提取性与合法包络约束。Failed-AVP内部引述坏字段的例外不能用于顶层Proxy回显。本轮E31只覆盖可合法构造的目的拒绝，不以这九份3002证明所有缺叶/坏Proxy拒绝都已闭合。

### 27.5 剩余P1规范证据：精确条款仍优先于跨命令类推

| 能力 | 本轮核对与结果 | 关闭条件 / 当前边界 |
| --- | --- | --- |
| CR0069与IP-SM-GW-Realm | [3GPP登记表](https://portal.3gpp.org/ChangeRequests.aspx?q=1&specnumber=29.338)仍列C4-264035/CT4#137 agreed，目标Rel20、受影响19.3.0，未列TSG批准/纳入版本。再次打开下载入口及官方ZIP，工具未能取得正文；上轮HTTP403记录保持，本轮不伪造新的HTTP状态或正文摘要 | 取得修改页并核对Number/Host/Realm、Additional及接口范围，再确认批准/Rel18回移植；固定标准域及s6c-ip-sm-gw-realm-v1不扩展 |
| 无IMSI成功SRA | 重新比对29.338表5.2.1.1-2的成功身份与表6.2.1.1-1的MO身份：后者明确提供IMS、IMSI不可用时的dummy MSISDN+Originating SIP URI路径，但它是OFR发送方身份，不是SRA成功User-Name的替代规则 | 不把一个命令明确的无IMSI例外迁到另一命令；普通成功仍真实IMSI，Router token及仅MCCMNC用途按既定特殊分支。需要直接修改SRA条件的依据，不能靠“MSISDN-less”名称或HSS-ID格式同形关闭 |
| Requested=Maximum | 复核29.338表6.2.2.1/2的上限关系与§6.2.2.2的严格早于过程，差异仍在。CR0017 rev3已定位C4-152173/CP-150776并批准纳入13.0.0，但本轮下载入口未取得正文；登记元数据不能裁决等号 | 标准构造仍严格早于，接收相等保留5550/原Time并带范围报告；原Maximum或era缺失单独处理。待eDRX修改正文/明确澄清，不新增调度器或持久提醒 |
| SMEA与具名Realm兼容 | 本轮公共字段样本不是SMEA编码或Realm组合的独立证据 | 保留所选有限子集与既有compat范围；仍需具名对端/独立解析，不靠矩阵423行推导已互通 |

第一项下载入口为[C4-264035](https://portal.3gpp.org/ngppapp/DownloadTDoc.aspx?contributionUid=C4-264035)，时间源头入口为[C4-152173](https://portal.3gpp.org/ngppapp/DownloadTDoc.aspx?contributionUid=C4-152173)。这些未取得正文的条目不增加本地规范原件摘要，也不增加需要业务账号或自动追踪服务的库依赖。已明确普通子集的设计工作不被这些能力P1整体阻断。

### 27.6 其他库对照与下一门槛

| 来源 | 观察 / 本库取舍 |
| --- | --- |
| [OTP diameter_dict 2.7.2](https://www.erlang.org/doc/apps/diameter/diameter_dict.html) | @messages与@grouped分别规定命令和直接父组形状；借鉴分层，不用一个全消息字段集合代替每父次数。没有采用新的运行时字典DSL |
| [OTP diameter_app 2.7.2](https://www.erlang.org/doc/apps/diameter/diameter_app.html) | 请求解码errors与应用回调/答案构造区分；可参考错误报告与处理阶段分离，不据此宣称其对坏Proxy的返回策略与本库一致；未运行OTP |
| 本地go-diameter [message.go](../../go-diameter/diam/message.go)，HEAD f26e3528006fcf0d33cb663994a0ec0ccbf7ef55，SHA-256 `ad182fd58a6fdb4ea4b08e0999a6b405aa2cd9952a0d827d31cf8dcccdf22e21` | FindAVP先查字典后按Code递归查找，FindAVPsWithPath返回多个父路径的扁平结果，内部路径比较仅Code；不能用这类结果证明精确Vendor或父实例的必需/次数。只读该固定实现，不泛化为全部API均不能正确验证 |

第三十一轮提出的后续工作于2026-11-03完成分流：§27.4错误出口进入P2/P3实现测试；基线和新增输入的独立解读、SMEA/具名Realm互通进入P4/P6发布验证；三项外部规范未决维持既定受限行为，不作为全部实施的前置。Peer/标识/预算/SCTP按P3实施。当前队列以[Specification §7](diameterkit-specification.md)为准；不再追加前置报文穷举，Ro现纳入首版。

未有定向反例的16行精确为：L032–L035.SNR；L038.SRA/RDA；L040.OFR/RDR/SRA/RDA/TFA；L041.OFR/RDR/SRA/RDA/TFA。它们均在既定按需边界，不为追求439/439而新增完整业务helper。已有反例的423行仍需错误维度、应用前提、完整错误输出与运行证据审查；不将R29/R34标为完成。

### 27.7 本轮静态交付复核

2026-10-31：15份文档、555个本地链接、1157段message/AVP hex及24个稳定SPEC ID复核通过。214份新增帧的Message/TLV边界、长度、SHA-256、头和原请求Session关联已检查；逐父验证只修改指定实例/子项，其余父组保持。九E3002的本地Origin、P/R/E/T、基础结果及正常应用字段省略均与表述一致；六个条件realm-only输入与ALR过程反例分别登记。静态检查未执行真实目的策略或错误发送逻辑。

正常矩阵439行重算为389在位、423定向，本轮补104条原缺例；所有公共行已有定向输入，16条按需缺例逐项核对。旧699逐叶帧、148答案帧、96段早期hex及33条派生记录保持；前轮Vendor/父实例、未知扩展/诊断、状态/路由/结果/文本静态检查复跑。19份业务源码、6份参考库源码摘要保持；既有29.229 V18.0.0原件SHA-256 `3a3d6581afb7ab3533e47b2ad9cedf7ebfc1c74226f2bca69c10a8f690ee5bbf`复核一致。

临时生成/检查器仅在/tmp，仓库只改文档；未修改实现、测试代码、依赖或其他仓库，未运行库/独立Diameter栈、真实DRA、Peer竞态、内存/日志预算或SCTP验收。未改XML，不重新声明XSD运行结果。CR0069及C4-152173正文未取得；规范缺口、保守关闭的对端后果和完整错误出口仍按P1保留，运行验收继续Planned。

## 28. 首版Ro：操作边界、规范争议和有限退出条件（2026-11-04）

本节继续Ro协议收敛，只细化既有RO-01–06、RESULT-04、PAIR-03及TEMPLATE-01/02。首版Sh/S6c/SGd-Gdd范围、Draft 0.19及24个SPEC ID不变；不新增全字段矩阵、报文全集或持久化框架。首版Ro交付仍是消息构造、校验、结果解释和有界exchange，不是完整CTF/OCS业务系统。

### 28.1 本轮核对及证据限度

重新核对[32.299 V18.0.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132299/18.00.00_60/ts_132299v180000p.pdf) §6.4.3、§6.5.10.3、§7.1.8A、§7.2.171及§7.2.208A：正常CCA的基础Result-Code仍必需；特性协商要求实验5011，却没有在这些条款中给出唯一伴随基础码；2022是退款凭据；3408描述仍指向离线。对照[32.274 V18.1.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132274/18.01.00_60/ts_132274v180100p.pdf) §6.3.1.2与§6.3.3.1，失败事件描述与普通SMS在线列仍存在§12.2所记张力；表中T4 Device Triggering和MO-SMS via T4列不能混作普通SMS列；本节原将第三列误称投递报告，§31已据原表纠正。此次为在线文本复核，未新增PDF原件哈希或完整表格视觉审计；不更换固定V18.0.0 Profile。

CR0081 rev1的批准状态及已读修订正文沿用§12.1证据；[CR登记页](https://portal.3gpp.org/ChangeRequests.aspx?q=1&specnumber=32.274)仍可核对该记录。本轮没有重新下载CR原件，也未获得直接解除两规范冲突的新裁决。不能再把问题写成“等CR批准”，也不以“没有搜到”证明所有勘误均不存在。

[RFC4006 §6.2/§6.4/§8.6](https://www.rfc-editor.org/rfc/rfc4006.html)明确了余额查询、退款和余额结果的操作条件。其通用DCCA过程与32.274的完整SMS用途必须分别声明；RFC8506全量迁移不作为本次已选4006子集的前置，亦不得宣称8506符合性。

### 28.2 操作与结果的固定契约

以下各行均先满足既有Wire/Exchange校验、原请求关联、正常/E包络及结果命名空间规则，再作操作解释。基础2001只说明对应协议操作的结果，不能代替服务级结果、余额结果或外部账本事实。表中的“可实施”表示设计可进入测试，不是当前已有实现。

| 既有覆盖项 | 操作与必要关联 | 结果/拒绝分支及支持边界 |
| --- | --- | --- |
| RO-01：余额检查 | 通用DCCA或已明确的旧对端Profile；EVENT=4、Number=0、Action=CHECK_BALANCE(2)，Subscription-Id识别用户，Service-Context明确 | 成功答案须取得Check-Balance-Result：0为ENOUGH_CREDIT、1为NO_CREDIT。2001+NO_CREDIT是查询完成且余额不足，不能解释为扣费失败或扣费成功；成功缺值不能默认为有余额，按应用条件缺失报告；未知枚举保留并报告未支持。有效基础失败不要求伪造余额值；原动作不匹配时不能用该值得出余额结论。查询不预留、不保证后续扣费成功 |
| RO-02：IEC直接扣费 | 已选标准SMS成功事件子集沿用§14.1：EVENT=4/Number=0、Action=DIRECT_DEBITING(0)，服务键与MSCC单位对应 | 基础/服务结果、GSU缺失/显式0分别解释；4012失败可无GSU，4011按不再需要信用控制单独解释；4010/4012不能证明先前报告用量未扣除（§29）。不能把事件单位任意换成顶层金额，不能从2001推导外部事务完成。通用DCCA金额及旧模板走各自解释域 |
| RO-03：退款 | 新EVENT/Action=REFUND_ACCOUNT(1)/Number=0与新Session/exchange；应用提供原操作关联及退款意图。2022若曾由原IEC CCA提供则原字节回送，含零字节与Presence均保留 | 通用退款构造和2022表示可以实施；32.274实际失败SMS退款受§28.3限制。退款CCA不因无GSU判失败；不伪造token、不查账、不推导资格、不自动补偿。结构候选不能标为标准SMS退款全流程通过 |
| RO-04：ECUR | INITIAL=1/0至TERMINATION=3/1仅适用于中间无UPDATE的已选路径；同Session/OCS、不同exchange。应用提供实际编号、单位和事件状态 | INITIAL授予额度不等于投递完成；TERM可2001且无GSU，5002不自动重建INITIAL。USU=0与缺失分开。涉及失败SMS/3408或3409的分支受§28.3限制，不能因成功终止样本存在就声明失败终止完整。通用UPDATE表达不等于承诺完整SCUR/UPDATE工作流 |
| RO-05：ASR/ASA | 按§2.1的独立命令语法和固定来源绑定接收入站ASR；ASA关联该ASR，不补CCR的CC-Type/Number。会话是否存在、能否中止及答复结果由应用显式决定 | ASA 2001对应应用报告已终止，5002对应会话不活跃，5012对应其他未停止原因；库不独立核验外部状态，也不能因无业务session表一律返回5002或因收到ASR自动2001。并行TERMINATION结果不能代替该决定；每个入站请求仍受既有应答/重复/关闭契约约束，不新增会话表 |
| RO-06：原身份重传 | 仅已审核操作、同终端OCS/去重域、同活动exchange；双DRA只是可选传输路径。原E2E/Session/CC编号/内容保持，HBH更新，T位按累计发送证据 | 沿用REPLAY-01–04的互斥调度、截止时间、次数、预算和一次完成。有效失败、业务Tx超时、已完成请求、重启、realm-only或换OCS都不能自动授权该能力；业务重新发起是新操作，不能伪装成原身份恢复 |

余额检查的条件校验是对§4 RO-EVENT的细化，不按字典中“全局可选”省略成功查询所需结果。未知/未支持与已证明缺失分开；上述表不新增公开结果枚举或facts接口，命名仍在实现阶段确定。

### 28.3 SMS失败/序号：保留限制，退款分四层验收

本轮没有获得足以撤销SPEC-RO-SMS-TABLE的直接依据；§12.2的3409/3408类型、保留方式和发送限制继续生效。只限制受冲突影响的事件，不把整个Ro标为不可实现。具体发布边界固定为：

| 层次 | 可以承诺的内容（通过对应测试后） | 不能据此推导的内容 |
| --- | --- | --- |
| RFC4006通用DCCA退款 | Action、单位、服务关联及结果的已选纯构造/解释 | 不自动具有32.274失败SMS语义；不能通过换Profile名称规避实际SMS条件 |
| 32.299退款凭据 | MSCC中10415/2022的opaque原样传递、Presence、边界及相应退款结果 | 有token不证明具备退款资格、退款金额正确或失败SMS报告完整 |
| 32.274标准SMS失败事件 | 安全接收原字段并报告具体未裁决规则；helper在接纳前明确拒绝受影响构造 | 暂不能声明失败退款/失败终止及在线序号涉及的分支完整支持；禁止删3409、伪造成功事件或搬运序号以绕过限制 |
| 具名旧OCS兼容 | 只有固定模板/版本/部署绑定、字段语义及成功失败证据齐备，才启用该具名操作 | 0/1292私有token不自动等价于10415/2022；旧金额420 Grouped不覆盖标准420 CC-Time/413 CC-Money；SCAP1不冒称标准Ro |

RO-03仍是实际A/B替换义务。如果目标sms-fw依赖受限失败退款，标准子集可单独发布，但**不能宣称该旧计费路径已替换完成**；必须取得直接标准裁决或完成目标OCS具名兼容验收。仅IMEI、SCCP复合地址及未选深层数据继续按§12已有有限子集报告，不为本轮再开全集清查。

### 28.4 5011：客户端接收与服务端生成分别关闭

§9.4的接收矩阵已足以定义CTF客户端行为：正常CCA只含实验5011而缺必需基础结果，按无效答案；基础2001与实验5011并存则保留冲突/特性失败，不能导出计费成功；基础失败与5011并存保留两者，不宣称其配对已获标准裁决；未知实验结果或特性保持范围报告。有效失败不触发重新扣费、降级重试或修改Peer能力缓存。

**标准5011自动生成仍受限，限制定位到OCS应答分支。** 发送方所需特性、答方支持集以及实验结果都明确，也不能据此编造伴随基础码；不能套用Sh的结果XOR规则绕过CCA必需268。普通CCA成功/已明确失败的构造及上述接收处理可以分别进入验收，但不宣传完整OCS特性协商支持。若以后确有目标对端要求特定配对，按已有具名兼容证据和固定binding审查，不新增任意Result-Code覆盖开关。

因此这项不是“等5011解决才能实现整个Ro”的门槛。扩大生成支持只接受直接规范依据，或带明确非标准标签的目标对端兼容证据；当前将受限行为本身作为测试预期。

### 28.5 其他库的借鉴及本轮结束条件

只读核对本地go-diameter的`diam/dict/testdata/credit_control.xml`（SHA-256 `c432c3a84f5ea36c0b699d0b584c76957ee3f0e2fa6838f0075416b853f9c5ad`）：CCA字典要求Result-Code，但把Check-Balance-Result列为全局可选；因此可借鉴结构描述，不能以字典代替原请求动作和成功条件。`tgpp_ro_rf.xml`（`1d111f092169d086e0fa404d0a7a1d73766d36528e63891aca639c62079caef7`）定义2022为OctetString，只能佐证表示方式，不能证明失败SMS退款过程已解决。本轮未运行该库或旧encoder；其他库不是规范冲突的裁决来源。

首版Ro只保留下列五包工作，沿用原验收ID；优先级是Ro阶段内部顺序，不回流为首版P0任务：

| 顺序 / 工作包 | 固定退出条件 | 状态及阻塞范围 |
| --- | --- | --- |
| 1 / 已选操作与结果 | RO-01–05、RESULT、PAIR-03：余额足/不足/缺失/未知与原动作关联；IEC成功/4012/服务失败；TERM成功无GSU/5002/0单位；退款token原样；ASR成功/未知及并行TERM。使用既有候选，在实现测试中补余额/ASR等必要输入，不先扩写全hex库 | 已定行为，待实施验证；受限SMS分支测试拒绝/报告，不伪装成功 |
| 2 / 5011接收及构造限制 | RESULT-04：§9.4各结果组合、支持列表与原请求条件；拒绝未裁决标准配对，未知命名空间不能降为成功 | 客户端行为可实施；完整标准OCS 5011生成受限 |
| 3 / 活动exchange恢复 | REPLAY-01–04：固定OCS双DRA、竞争旧CCA、取消/关闭、部分写/后继未写、次数及额度耗尽；一次完成且原期限不延长 | 已定行为，待运行及目标去重域验证；无业务恢复引擎 |
| 4 / 目标兼容 | TEMPLATE-01/02、COMPAT-01：六模板分别关联实际使用配置、源码/构建/模板版本、完整旧编码请求、成功及失败答案、独立解码与新旧差异；核对结果、单位、token及端点绑定。未部署模板仍逐项标未验证，不扩大或默删原义务 | 现仅源码和条件化片段；完整替换声明受目标证据阻塞，标准codec可继续 |
| 5 / 受限能力登记 | 保持§28.3/§28.4及§12的支持/不支持清单；只在获得直接依据或明确目标需求时重新开启相应分支 | 本轮边界已固定；不要求当前穷尽CR、8506、所有AVP或所有OCS |

本轮文档结束条件是：三项争议有确定的发送/接收处理，RO-01–06有分层的操作与结果契约，剩余项有可执行退出条件和明确证据责任。本节已经给出这些条件；后续普通推进不再安排“下一轮全Ro specification补齐”。测试发现与这些契约矛盾时定点修订；直接证据改变受限结论时只审对应分支。运行与互操作均仍Planned，不把协议边界已定记为符合性通过。

## 29. 首版Ro交叉审查：结果含义、ASR与错误包络（2026-11-04）

本节按用户要求继续文档/specification审查，尚未开始实施。范围固定在§28已有五包内：修正容易导致错误实现的契约，不追加全字段/全报文矩阵。§28的SMS失败/序号、5011生成和六模板证据限制保持，未取得新依据，不宣称这些争议已解决。

### 29.1 审查发现与原位修正

| 优先级 / 已有契约 | 问题与影响 | 本轮决定 |
| --- | --- | --- |
| Ro P1 / RESULT、RO-02/04 | §4已经区分4011，§28操作摘要却只突出2001/4012，容易实现成“非2001均拒绝”；也未明确拒绝与既有用量扣除的关系 | 原位补§28和§10.5；4011独立解释，4010/4012不作为零扣费证明。接口保留码、层级、动作、Presence和报告，不新增账本查询 |
| Ro P1 / RO-05 | “ASA不证明已取消”未区分协议陈述与库的核验职责；状态职责表又将CC编号要求合并到ASR/ASA | 修正§2.1/§28、Profile、API、状态/运行时正文；明确2001/5002/5012以及ASA不要求CC字段 |
| Ro P1 / RESULT-01、BIND、PAIR-01/03 | Profile已有E错误独立语法，但摘要没有给Ro的可核对反例；实现容易用正常CCA结构拒绝DRA错误 | §29.3明确三种包络、错误来源和终结结果，沿用同一完成门与现有错误分类 |
| Ro P1 / REPLAY-01–04 | 固定Ro引用4006，运行时§1.6却直接用8506解释当前恢复范围；“库不做”易被读为“规范不允许” | 改回4006直接依据，标明库的有限交付与完整客户端程序之别；不扩大重传、不导入持久状态 |

### 29.2 结果码：协议报告与业务事实分开

依据[RFC4006 §6.3、§9.1–9.2](https://www.rfc-editor.org/rfc/rfc4006.html#section-9.1)及[TS32.299 V18.0.0 §7.1.11](https://www.etsi.org/deliver/etsi_ts/132200_132299/132299/18.00.00_60/ts_132299v180000p.pdf)。本轮仅复核所列条款，不升级Profile或声称全结果符合性通过。

| 原答案含义 | 解释器交付 / 不允许的推断 |
| --- | --- |
| DIRECT_DEBITING收到4011，未授予GSU | 报告信用控制不再适用，不能套2001的授予条件判“缺GSU”；保留非2001原码。是否继续服务由应用决定；不生成假额度，不自动开离线计费 |
| CHECK_BALANCE或REFUND收到4011 | 保留原动作和该指示，不从4011生成ENOUGH_CREDIT或退款完成。原动作的专用结果条件仍适用；不能把DIRECT_DEBITING的分支机械套到其他操作 |
| 4010或4012，原CCR报告了USU | 请求被拒绝不等于报告的用量未扣除；可发生用量扣除。库不能证明具体账务结果，不据失败自动生成反向退款、清零USU或重发计费 |
| 顶层2001，某服务4011/4012 | 保留每个MSCC实例及服务关联；无信用控制与额度不足分别解释，不将任一服务的结论提升到全消息或其他服务 |
| 顶层失败，某服务2001；或5011实验失败与基础2001冲突 | 仍按原优先级/冲突规则，不由服务成功或基础码抹掉失败。4011的特殊含义不解除结构、来源、特性及原请求关联检查 |
| 5030或5031 | 用户未知与计价失败分别报告；5031仍要求相应Failed-AVP，诊断用已有有界规则。两者不是重新发送授权，也不为解释结果查询外部用户或计价服务 |

这是结果解释而非业务授权API；不添加一个吞掉层级差异的`Success`/`Charged`布尔快捷结论，也不改变缺/零GSU和合法费率双组的既有规则。

### 29.3 正常CCA、E错误和ASA的关联检查

依据[RFC6733 §6.2、§7.2、§8.5.1–2](https://www.rfc-editor.org/rfc/rfc6733.html#section-8.5.2)及已有§24.3完成顺序。先找ConnectionID/HBH并比对App/Command/E2E，再使用原binding和预算；本节不增加Session全局索引。

| 包络 | 要求及反例 | 完成与业务边界 |
| --- | --- | --- |
| 正常CCA，E=0，Command272/App4 | 按固定语法要求Session、Auth-Application-Id、CC-Request-Type/Number和基础结果；CC字段须匹配原CCR。2001、4011、4012或正常5002均不免除这些正常包络要求 | 完整头关联后的缺失/错值按既有ErrInvalidAnswer；合法答案按其动作解释，不等第二个“更成功”的答案 |
| 对CCR的E错误，Command272/App4 | 使用基础错误语法。原CCR有Session故仍须回显；可没有Auth-App/CC类型/编号/MSCC，不能报这些正常字段缺失。合法3002/3004来自已授权ErrorAgentOrigins时可被接收。E=1的4011/4012不属于允许的E结果类别，不能作为该正常分支 | 不给额度/扣费成功结论；合法错误终结原exchange，不触发本库原身份恢复。错误Agent不得因能返回3002而被授权返回OCS成功或4011 |
| ASA，E=0，Command274/App4 | 正常ASA的Session、基础结果和Origin按自身语法/原ASR关联；不要求Auth-App或CC类型/编号，不把原ASR的Auth-App盲目复制成必需答案字段 | 2001表达应用报告已终止；5002为会话不活跃；5012为其他未停止原因。库不独立验证外部事实，Committed/Written也不是额外业务证明 |
| 对ASR的E错误，Command274/App4 | 按基础错误语法与原ASR头/Session/来源关联；不套CCA或正常ASA专属条件。E中的附加应用字段仍做适用的结构/资源检查，不能用于导出正常业务成功 | 已有一次完成、诊断和来源策略保持；不自动发CCR、STR、退款或成功ASA |

E包络若带扩展中的CC字段，不因此切回正常CCA或获得额度语义；这些字段不能覆盖头/Session匹配或拓宽来源。正常CCA的缺GSU、错误服务结果及未决语义继续由对应规则分层报告，不能为了统一接口都伪装成传输失败。

ASA业务决定应由调用方依据当前会话事实给出，不要求向库提交业务记录。不同请求的ASR与TERMINATION没有库内跨命令原子性；同一ASR的重复副本仍使用既有内容比较与缓存响应，不能因业务状态后来改变便改写已提交结果。ASA成功的协议含义和公开库无持久化边界可以同时成立。适用32.299 §6.5.5时，应用CTF另须发送CCR TERMINATION关闭关联信用控制会话，ASA不能替代该报告；细化见§30.4。

### 29.4 恢复范围与参考库核对

[RFC4006 §5.7、§6.5](https://www.rfc-editor.org/rfc/rfc4006.html#section-6.5)区分路径传输恢复、会话迁移和事件失败过程，也有错误后换服务器及缓冲处理。普通R33只承担已审核同OCS活动exchange恢复；RO-06扩展见§33/运行时§1.7，不实现的过程必须以支持限制公开，不能倒写成RFC禁止。该限制也不能由DDFH/CCFH、Redirect或DRA就绪自动解除。RFC8506仍是独立后续目标。

只读对照[go-diameter基础字典](../../go-diameter/diam/dict/testdata/base.xml:117)的ASR/ASA定义：ASA没有CCR的CC字段；字典描述字段形状，不能核验外部会话是否终止。该文件SHA-256为`b3bfce2703e97e02aea3b4ee85bc1d11af3a6be8c415b3831926c7674dc404fc`；没有运行该实现，也不把整个旧字典视作现行RFC的完整符合性证明。

本轮四项文档问题均已有确定修正，落入既有五包，不新增第六包或SPEC编号。仍只做文档；后续审查可继续针对明确问题，不自动要求转实施。SMS冲突、5011标准配对、六模板完整执行/对端证据是保留项；没有直接新证据时不重复穷举这些外部问题。

## 30. 首版Ro：服务集合、凭据证据和会话特性（2026-11-04）

本轮只审查既有五包内的四个问题：MSCC目标集合、退款条件输入、特性协商生命周期，以及ASR后的CTF协议义务。首版范围、固定Profile、24个SPEC ID和候选字节不扩展；没有启动实现或重新穷举规范冲突。

### 30.1 MSCC不是按下标配对的服务字典

依据[RFC4006 §5.1.2、§8.16](https://www.rfc-editor.org/rfc/rfc4006.html#section-8.16)，一个MSCC可以对应多个服务或一个rating group，服务端可选择服务粒度或组粒度授权。§10.5已有服务键优先和合法费率双组规则，本轮补足解释边界：

- 以原Service-Context、服务标识/Rating-Group及已审核的限定字段关联，不能按CCR/CCA的MSCC出现次序逐一配对；合法换序不等于换服务。Canonical的重复内容比较仍保留原有顺序规则，不能把解释时关联等同于去重时重排。
- 同组含Service-Identifier列表时，额度关联该集合；不能把GSU的N个单位复制成列表中每个服务各N单位，也不按列表长度自动平分。Rating-Group同时存在不取消显式服务目标。
- 只给Rating-Group时，不仅凭数字相同猜出全部成员。若原请求及已选规则不足以关联，则保留组和范围报告；不增加在线rating映射查询，也不把“无法唯一解释”伪造成线上的5009。
- 多实例、重叠目标和费率/池限定不能在map中最后写入覆盖，也不能无依据累加额度。已审核语义按对应程序解释，范围外保留原组并报告；未出现的服务不自动得到额度或一个伪造的失败码。

本轮没有从RFC推导“缺MSCC Result-Code必等于2001”。缺结果仍按§10.5的已审核程序处理；全局ABNF可选不能证明任意业务成功。这里明确关系和不允许的推断，不承诺完整多服务池管理。

### 30.2 退款：条件未知不是原答未携带

[TS32.299 V18.0.0 §6.3.3、§7.2.171](https://www.etsi.org/deliver/etsi_ts/132200_132299/132299/18.00.00_60/ts_132299v180000p.pdf)将2022回送义务限定于之前IEC答案已携带的情况。沿用纯helper已有的原结果/凭据输入，固定以下解释，不新增公共facts或凭据仓库：

| 调用方已有证据 | 纯构造/检查结果 |
| --- | --- |
| 原IEC答案明确没有2022 | 仅就此条件可省略；不升级为退款资格证明，不解除其他字段条件及SMS失败事件限制 |
| 原对应服务的2022存在，payload为空或含零字节 | Presence为存在，须原样回送；不能因空值转成缺失，不能按文本处理 |
| 原答案未知、不完整或未说明是否返回2022 | 完整回送条件尚不能验证；helper返回本地缺少构造依据的报告，不能默认为“未返回”。低层raw表达不等于该历史条件已通过 |
| 多MSCC、同字节token或另一个服务的token | 依据原操作和服务父实例关联；相同bytes不证明相同业务凭据，不能“取第一个2022”或按token去重/跨组搬移。若所选关联无法确定，保持本地未完成报告 |

这里的输入来自应用已经持有的有界证据，库不验证其外部真实性。无2022不被全局定义为协议禁止退款；只有目标具名兼容确有更强要求时，才按该规则限制，不能把旧OCS经验写成标准。原请求未知也不自动触发退款；新退款和活动exchange原身份重传仍分开。

### 30.3 服务器报价、共同集合与本次需要集合

32.299 §6.5.10.3要求初始答案在适用条件下报告服务器能力，且无特性初始请求的失败答案也可有完整支持集；会话计费则按建立时双方共同能力使用特性。因此现有SessionFeatures不能仅由“收到一个有Supported-Features的CCA”生成。

| 输入/阶段 | 固定处理 |
| --- | --- |
| 失败初始CCA含服务器支持列表 | 保留为服务器报价及来源证据，不宣称信用控制会话已建立；不能创建可供后续TERM使用的成功协商依据 |
| 成功建立且必要协商依据齐备 | 共同集合取本地支持与该终端建立时支持集的交集，绑定Profile、OCS、Session及Service-Context；特性键含Vendor-Id/Feature-List-ID，不能跨服务域只按bit合并 |
| 发出后续会话请求 | 仅声明本次实际使用、需要对端支持的特性，并检查它们属于已知共同集合；不复制整个共同集合，不因新一次报价扩大建立时的能力 |
| 缺报价、未知位/未知列表或输入不匹配 | 未知与已知空集合不同；缺当前必需证据给既有本地上下文/范围报告，不推导全部支持。未使用这些特性的基础请求不因此强制查外部状态 |
| 独立IEC事件、换Session或换OCS、DRA重连 | 不继承另一个事件/会话的SessionFeatures；DRA能力和连接恢复不能授权特性。库不维护全局能力缓存或隐式更新调用方输入 |

这是32.299请求使用规则的具体化，不把Sh的特性分支或结果XOR直接移入Ro。5011接收/生成限制仍按§9.4/§28.4；对端报价也不证明某个未审核SMS深层字段已获本库语义支持。

### 30.4 ASR后终止报告的职责不能漏写

32.299 §6.5.5明确：适用的CTF收到ASR后以CCR TERMINATION关闭关联信用控制会话；用户会话行为由middle-tier规定。前轮写明“库不自动发CCR”是正确边界，但单独这句话不足以描述调用方的协议义务。

应用负责协调ASA决定、已在途TERM、实际用量及CC编号；若需要发送TERM，使用普通构造/请求入口建立独立exchange。库不替应用从ASR猜出用量，不把ASR的HBH/E2E复制成新CCR身份，不以ASA Committed/Written代替TERM结果，也不强迫另一业务操作等待ASA完成。两条消息的业务一致性和重复触发控制归应用，不新增跨命令事务、后台worker或持久队列。

此补充不改变RFC6733的ASA 2001/5002/5012含义；也不据此规定一收到任何ASR都另发第二份TERM。没有活跃关联、已有TERM在途等情况由应用结合对应程序处理。未核对的middle-tier用户行为不得扩成短信撤回、重投递或自动退款。

### 30.5 证据与本轮退出记录

本轮在线核对RFC4006上述条款、32.299 V18.0.0 §6.3.3/§6.5.5/§6.5.10.3/§7.2.171；32.274 V18.0.0在线获取超时，未据此新作SMS用户会话或失败事件裁决，既有原件/CR证据保留。未新增PDF原件哈希。

只读复核[go-diameter信用控制字典](../../go-diameter/diam/dict/testdata/credit_control.xml:251)与[Ro/Rf字典](../../go-diameter/diam/dict/testdata/tgpp_ro_rf.xml:2481)：前者MSCC重复字段仍有§10.5登记的max=1差异，后者定义特性组三个必需子项，但都不能代替跨消息关联、条件Presence或会话生命周期。两文件SHA-256分别为`c432c3a84f5ea36c0b699d0b584c76957ee3f0e2fa6838f0075416b853f9c5ad`、`1d111f092169d086e0fa404d0a7a1d73766d36528e63891aca639c62079caef7`，与既有快照一致；未运行参考库。

本轮新增判断均归既有RO/RESULT/PROVISION/PAIR验收；见[覆盖§35.2](diameterkit-usage-coverage.md)。四项契约已具体化，外部规范冲突与六模板证据缺口未关闭，运行仍Planned；不新增第六工作包、字段全集或必需存储设施。

## 31. 首版Ro三项重点：标准未决与兼容证据的最终门槛（2026-11-04）

本轮限定SPEC-RO-SMS-TABLE、SPEC-RO-5011和TEMPLATE-01/02，不扩字段或hex全集。收敛的是**确定的库行为、证据缺口和解除条件**，不是把外部矛盾强行判为标准已解决。公开Go-only、无持久化及Ro现纳入首版。

### 31.1 新版与原件核对

本轮从ETSI目录核对并取得下列对照原件，不替换固定Rel-18 Profile。目录和CR检索只是本轮可见证据，不能宣称穷尽所有3GPP勘误。

| 对照原件 | bytes / SHA-256 | 实际核对范围 |
| --- | --- | --- |
| [32.274 V19.1.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132274/19.01.00_60/ts_132274v190100p.pdf) | 797963 / `95a7eb876f2554f0d968e584adb85877df9766ad684dbed38c490f7d29c56888` | §6.3.1.2.1定义、§6.3.3.1在线表及§6.5收敛计费表；已读取保留版式文本并视觉核对印刷页90的在线表 |
| [32.299 V19.0.0](https://www.etsi.org/deliver/etsi_ts/132200_132299/132299/19.00.00_60/ts_132299v190000p.pdf) | 1256608 / `31f71799ebeb0d0c449d0d09aa7073a4fd76fbca1ab3a4e1fdb9977e68b75d50` | 与§12既有原件哈希一致；重读§6.4.3、§6.5.10.3、§7.1.8A、表7.2.1、§7.2.208A/212A及SMS组定义 |

[32.274 CR登记页](https://portal.3gpp.org/ChangeRequests.aspx?q=1&specnumber=32.274)仍列0081 rev1/S5-213635已批准、纳入16.3.0；此前已读取的修订正文证据保持，本轮未重下CR。32.299登记页可见1122条、6页，本轮仅核对当前返回页与相关标题检索，不以第一页无相关标题宣称不存在修复CR。未获得指定基础结果配对或统一SMS在线字段表的新直接裁决。

### 31.2 SMS失败与序号：冲突拆到具体用途

新版仍有：32.274定义表将失败SMS result列为C、序号列为串接条件下OC；在线用途表的普通SMS两项却均为“-”。32.299的3409/3408使用表仍未列CCR使用，3408正文仍限定offline。CR0081删除32.274描述中的offline-only，并未在本轮所查版本中同步消除这些差异。

**纠正旧§28.1列名：** 在线表三列为SMS、T4 Device Triggering、MO-SMS via T4；第三列不是投递报告。其3409行分别为“- / E / IT”，3408为“- / - / -”。§6.5的converged charging表另列ITE，不能移植成Ro CCR许可。

| 具体输入 | 当前决定 | 能解除限制的最小证据 |
| --- | --- | --- |
| 普通SMS失败事件，需表达3409 | 标准helper仍在接纳前报告该分支未支持；接收保留原值与未裁决规则，不仅据“-”指控对端违规；不删除3409伪造正例 | 直接统一定义/在线用途/AVP使用关系的规范或批准勘误，或只对目标部署生效的具名兼容证据 |
| 普通串接SMS需报告3408 | 与失败结果分开登记；成功串接也可能触发该字段限制。只会携带UDH/TPDU不等于支持3408语义 | 明确在线序号用途及对应命令/父路径的直接依据；不能只凭CR标题、U32宽度或另一个协议的序号推断 |
| T4 Device Triggering或MO-SMS via T4的3409 | 32.274对应列的许可是已确认的局部依据，但仍不能抹去32.299 CCR表差异；不把它合并成“所有SMS均禁止”或“普通SMS退款已解决” | 该T4用途的跨规范闭合和目标流程证据；本库当前不因此新增原生T4计费能力 |
| 不触发上述字段的已选普通成功路径 | 继续按原规则构造/解释；标准子集不等待外部争议关闭 | 对应既有操作验收；不扩大为完整SMS计费声明 |

SMS-Result只是值表示，并非本库可按失败原因随意赋值的全局枚举；结果数值来源亦须随具体事件/规范或具名对端登记。现有退款token、旧模板没有3409/3408、成功终止样本均不能替代该证明。标准限定下的DECISION-01保持，受影响旧业务路径的完整替换声明仍受阻。

### 31.3 5011：发送配对不能从相邻分支类推

对照32.299 V19.0.0仍是正常CCA必需268、可选297，特性不足指定10415/5011而未在所查条款指定伴随268。定义一个5011常量、按Sh使用单实验结果、或把失败改装成E包络，都不是该正常CCA分支的标准解法。

| 候选做法 | 文档决定 |
| --- | --- |
| 仅Experimental-Result=10415/5011，无268 | 正常CCA无效；不以“特性结果已明确”豁免必需基础码 |
| 268=2001并附实验5011 | 接收保留冲突/失败，禁止计费成功；该接收能力不授权标准helper生成此配对 |
| 268=5012/5003等基础失败并附实验5011 | 接收按已知双结果规则保留，不能从“更保守的失败”推导标准指定了该值；不自动选择一个通用失败码 |
| 268=5001作为特性不足的统一替代 | 5001+Failed-AVP对应不识别必需AVP的另一条件；已经理解Supported-Features但不支持其中功能不能无依据套此分支 |
| 268=5011或E错误作为绕过 | 基础5011是另一命名空间的版本错误；E语法不得被用作未决应用配对的自动修复，也不能缺268 |

因此，客户端安全接收契约已收敛，标准OCS自动生成此分支仍明确不支持。解除标准限制需要能指明正常CCA的基础码、实验码及适用条件的直接依据；只有真实对端特定配对证据时，只能关闭对应具名兼容门槛，不能宣称标准问题已解决。未支持时是本地构造限制，不因要应答而伪造线路码。

### 31.4 六模板：本轮补齐的证据分类

相邻sms-fw本轮HEAD为`80fa48c4126800cfa5c655e78889dffe6c03bd6c`；§14.3原14个文件哈希全部一致。以下新增三项只读源码证据；本轮17个审查路径的`git status --porcelain`均为空，不代表其他文件或现场binary状态。

| 新核对文件 | bytes / SHA-256 | 可证明及不能证明 |
| --- | --- | --- |
| [tmplverify.cc](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/tmplverify.cc:9) | 1313 / `386f6c1ce63cf4ecdf884f8b80f221de1fee50e823b0381ef7d13d0774b95422` | 读取模板并打印toString；没有实际Getter输入、完整CCR编码或对端答案。即使以后运行通过，也仅关闭模板加载检查 |
| [serversimulator/Worker.cc](../../sms-fw/c-cpp/libsrc/kernel/dialib/diameter-impl/serversimulator/Worker.cc:236) | 8938 / `509ff4c70a3bd26e7528148536c1dfc3ccf5cae0f47db6738778d02a1f323d89` | fillInAnswerFromRequest设置结果和关联字段，允许按配置省Session；这是模拟器行为，不是六模板逐字段接受或完整5011协商证据；同实现模拟端不是独立OCS |
| [dialib/CMakeLists.txt](../../sms-fw/c-cpp/libsrc/kernel/dialib/CMakeLists.txt:44) | 5972 / `3db7104d83d4b08dedfe4b73ad076d6f009e847face3f4fd9cf927ef69086ad7` | yppdiatmplverify链接旧dialib及依赖；本轮未构建/运行，构建定义不证明目标binary使用该include/模板版本 |

源码答案链还有确定的信息丢失边界：[DIACCParseAnswer](../../sms-fw/c-cpp/libsrc/kernel/dialib/dia-cc/dia-cc.cc:1447)要求找到基础Result-Code；循环中多个MSCC的Result-Code写同一临时值，最后在顶层为2001/2002时覆盖单一结果（1725–1731行）。未见该函数对297/298的专门解释。这不能证明旧系统正确处理5011；新库必须保留顶层/逐服务/实验结果，不能为“输出相同”重现覆盖或丢失。退款的code-only/top-or-MSCC查找也不能成为新库跨Vendor/父路径接受规则。

### 31.5 六行证据缺口与有限验收包

示例配置实际列出普通/init/term/refund四槽；普通默认指向通用模板，没有专用IEC槽，SCAP1示例版本开关为注释。此为仓库样例，不是实际部署映射。每行都保留原义务；以下“缺”指本轮未取得的证据，不声称目标环境不存在。

| 模板 | 已固定的源码依据 | 仍缺的具体差异证据 |
| --- | --- | --- |
| 通用 | 普通槽样例；413金额与10425/23；Address writer | 实际普通槽/版本、draft地址开关及Getter输入；完整请求和同操作成功/失败答案，证明顶层/服务结果、金额/单位解释 |
| IEC | 需将普通槽实际映射至IEC文件；STRING地址、IMSI填设备及旧父路径 | 实际槽映射与对端接受这些旧字段的依据；完整EVENT扣费请求、成功/拒绝答案和token来源，不由通用样本替代 |
| INITIAL | init槽、RSU与旧设备/地址字段 | 真实INITIAL请求/答案及授予服务键；与TERM同Session/OCS的关联证据，不只看独立成功码 |
| TERMINATION | term槽、USU/FINAL及Status/Discharge-Time | 与INITIAL配套的完整TERM请求/答案、0/非0用量及实际事件用途；不能以所有TERM均为delivery-report解释字段 |
| REFUND | refund槽、420 Grouped及0/1292；原IEC凭据关系 | 原扣费答案→退款请求→退款答案的关联链，token bytes/长度/Vendor/父路径、私有金额定义和成功/失败结果；单份退款请求不足 |
| SCAP1 | 分支271/App0、VSAI193/19302、11个重复607 | 实际版本/普通槽映射与最终header；全部重复组/Getter输出及相应成功/失败ACA。不是Ro272/4，不能用CCA替代证据 |

每行采用同一有限证据包：**部署绑定记录 + 旧编码完整请求及输入 + 同操作替代成功/失败答案 + 独立解码/语义来源 + 未来新旧差异验证**。一条请求不要求真的收到两个答案；成功/失败是相同输入条件族的独立替代测试。答案可来自可复现的目标OCS测试环境或带来源/版本的脱敏现场材料，不强制生产抓包，也不要求提供业务账本。没有真实对端资料时，可以先完成文档/离线golden基线，但须标相应互操作未验证。

退款420与标准CC-Time同键异类型是部署约束，不只是缺一份pcap：同连接/App4/相同父路径的固定解释不能同时容纳两种类型。目标DRA若只能把二者放在同一不相容binding，则该部署下的组合支持被阻塞；解决方向是兼容解释域/连接隔离或经证据确认的单一组合Profile，不能按Requested-Action动态切decoder、运行时忽略冲突或把420偷偷改成413。

### 31.6 三项收敛状态与解除条件

| 项目 | 本轮结论 | 下次开启条件 |
| --- | --- | --- |
| SMS失败/序号 | 普通SMS/T4/收敛计费用途已分开，纠正列名；标准受影响分支继续受限 | 获得能统一字段用途的直接依据，或指定目标的完整具名兼容包；不因再找到同类字典重复开启 |
| 5011 | 安全接收已定；标准生成仍无配对依据，候选替代做法逐项排除 | 直接配对条款/勘误，或明确非标准的目标配对证据；不得任选码“关闭” |
| 六模板 | 静态证据补到17个文件，六行缺项和部署冲突明确；完整编码/互操作未验证 | 按上表提交缺失材料或后续获准执行对应验证；解析器/同栈模拟器通过不能代替这些材料 |

本轮只修订文档，未新增规范实现、模板解释器、持久化、计费状态机或业务补偿。现有五包保持；未取得的外部证据不靠新增框架解决，不把文档收敛记作协议符合性或完整替换通过。

## 32. sms-fw C/C++实际Ro用法作为完整交付基线（2026-11-04）

用户明确要求完全覆盖sms-fw C/C++的Ro用法。本轮重读实际yppdia调用、dialib编码/答案/恢复及模板配置，按[覆盖§37](diameterkit-usage-coverage.md)登记必需验收。标准子集可作中间成果，但不能以“后续可选”“未见生产启用”或标准冲突为理由从最终Ro目标删除源码已接线的A/B用法。公开Go-only、无持久化以及Ro现纳入首版；不承诺复制全部旧业务引擎。

### 32.1 本轮源码证据

只读sms-fw HEAD=`80fa48c4126800cfa5c655e78889dffe6c03bd6c`；§14.3的14项源码/模板哈希逐项复核一致。本轮另固定以下4项，合计18个审查路径的工作树状态为空；§31的其他3项历史证据保持，不能把本次18项与累计证据数混用。未构建、运行或改动业务仓库，也未证明生产配置与仓库样例一致。

| 文件 | bytes / SHA-256 | 本轮用途 |
| --- | --- | --- |
| [yppdia/CMakeLists.txt](../../sms-fw/c-cpp/appsrc/kernel/yppdia/CMakeLists.txt:1) | 1369 / `1b2379df26c41c9019bbd0f67188aa29523f7210cfecedf5c1e01f7751e3d8c8` | yppdiaengine_m链接dialibth_static，include指向libsrc/kernel/dialib；不是仅凭同名导出头/工具样例认定调用 |
| [diaalarmhandler.cc](../../sms-fw/c-cpp/appsrc/kernel/yppdia/diaalarmhandler.cc:4) | 2108 / `cee9d0363e0664b08144f11f3c90b7f81e4daf475e9928a70deef017a4f149ac` | Peer上下线、单路由主备和全部接口可用状态的业务接线 |
| [dia-peer.cc](../../sms-fw/c-cpp/libsrc/kernel/dialib/dia-base/dia-peer.cc:611) | 79301 / `6ee63636e459c4104e96db5112103381e841fc43017b8b341955e7438b90a4da` | CER版本分支、ASR接收及旧ASA默认成功；新库按既定应用决定契约纠正 |
| [WCreditControlAnswer.hpp](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/WCreditControlAnswer.hpp:10) | 858 / `d72750cffe2fb9d1e8fee2864f034b26e2314d3a14406ffab2943b8982d672c2` | 费用有符号值/指数、币种、服务单位、refund bytes与私有operation_result的交付形状 |

搜索范围包括c-cpp/appsrc与libsrc的Ro调用和模板文件；本轮沿实际构建链审查yppdia与dialib，tools/serversim及include/include同名副本不自动当成主binary实现。未发现主调用发PRICE_ENQUIRY或SCUR UPDATE的新证据，不因此扩展全Ro过程。

### 32.2 相对标准子集不能遗漏的实际输入/输出

| 源码事实 | 对全覆盖的影响 |
| --- | --- |
| [PP_SUBMIT_CHECK](../../sms-fw/c-cpp/appsrc/kernel/yppdia/yppdiaengmx.cc:2204)发CHECK_BALANCE；精确条件为2000<result<3000时检查operation_result，非0改业务结果，最终只有2001走CHECK成功；wrapper同时交付check_result | 标准Check-Balance-Result与旧私有operation_result分开。旧初始化为1、非wire默认；库保留Absent，应用兼容策略及无答案分支见§34.2 |
| [CCA私有结果解析](../../sms-fw/c-cpp/libsrc/kernel/dialib/dia-cc/dia-cc.cc:1629)意图按873→20400→20430查Integer32，旧查找未校验Vendor；内层首项迭代还使用外层列表 | 逻辑父路径/读取类型已明，三层实际接收Vendor/flags和OCS缺省意义未证；不能从发送模板873的10415推导接收子叶Vendor。固定定义与剩余材料见§34，不复制错误迭代 |
| [IEC reserve](../../sms-fw/c-cpp/appsrc/kernel/yppdia/yppdiaengmx.cc:2353)调用handleCCAInfo；[费用与单位交付](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:222)包含Currency/Value/Exponent/service_unit及refund token | 必须覆盖Cost-Information、GSU及供退款使用的完整有符号费用表示；只交付Result-Code/2022不够。旧token默认0/1292、顶层/MSCC位置与标准2022分别建模，不改成float或统一成功布尔值 |
| [handleCCAInfo](../../sms-fw/c-cpp/appsrc/kernel/yppdia/yppdiaengmx.cc:1458)将费用和token装入reservationId并截断；[refund_info](../../sms-fw/c-cpp/appsrc/kernel/yppdia/yppdiaengmx.cc:763)解其内部布局 | 旧业务数据迁移归Go binary；库交付owned完整值/显式超限，不能暴露旧S1内存布局、按strlen裁token或复制静默截断 |
| [投递成功](../../sms-fw/c-cpp/appsrc/kernel/yppdia/yppdiaengmx.cc:2450)仅ECUR标记分支发送TERM；[投递失败](../../sms-fw/c-cpp/appsrc/kernel/yppdia/yppdiaengmx.cc:2538)的ECUR分支也实际发送TERM，虽注释说等ASR | 两种TERM及IEC成功不发第二次CCR、失败token不足不退款都需迁移用例；不能按旧注释删失败TERM。旧代码在失败TERM后强制业务成功，不证明OCS退款/撤销成功，应用须保留真实协议结果 |
| [encodeCCR](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:1991)按route组Session前缀，TERM复用计数、编号1，其余0；[route profile](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:3364)覆盖四模板槽、Service-Context/P位等 | 覆盖完整Session和应用明确给出的阶段/编号，不能照搬裸计数拼接或把新请求认成原identity恢复。多路由、四槽覆盖/默认选择与六模板是两层映射，需一起验收 |
| [模板变量映射](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/AvpsTemplate.cc:64)和[yppdia Getter](../../sms-fw/c-cpp/appsrc/kernel/yppdia/yppdiaengmx.cc:1088)有同名不同类型/不同来源分支 | 固定兼容定义必须覆盖已接线变量的实际输出，不按宏名字猜类型；未被某个样例模板引用的配置分支仍属B证据，不能默认废弃，也不宣称任意未来模板都已支持 |

Getter迁移至少按以下已有族核对：订阅方/收发方IMSI与号码及AS/V变体；SMSC/VMSC的STRING/Octets/E164形态和fakeVMSC；DCS/PID/UDH/content/status、接口/消息类型、长度/分段数/序号、时间/优先级/回执；tariff/service-description/audit/message-id；费用/币种/指数、GSU/CC_SSU及refund。它们进入已有FIELD/TEMPLATE验收，不要求在库中保留S1读取、号码业务选取、计价或模板DSL。应用给明确值，库负责定义、编码、校验和无损结果；实际配置引用但尚无固定描述的项是兼容缺口，不能仅靠raw承载标完成。

### 32.3 恢复与旧缺陷：不能用职责划分假装覆盖

[processAnswer](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:1310)、[Tx与晚答](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:1555)以及Tr/持久队列分支包含临时错误转backup、业务等待结束后处理答案、DDFH和缓冲。必须把每条实际需要的行为映射到“库协议能力 + Go应用编排 + 目标验证”，不能只写“归应用”便标RO-06完成。

- 活动exchange、同OCS、获授权的传输恢复沿R33；业务Tx可以早于库deadline，由应用持有请求关联并消费最终Completion，不能在Tx时提前终结又要求库无损接收后续答案。
- 多OCS初次选路可用固定Route/Peer；已有业务错误后的新操作、跨OCS会话迁移、持久重启恢复不自动获得R33授权。该缺口现由§33及运行时§1.7的受检原请求续发契约承接；不能伪装成普通新Request，也不能恢复库内Store/Engine。机制、应用接线和具体目标证据仍须分别验收，未通过前不宣称完全替换。
- Peer/Route状态与发送证据须足够供应用实现告警、Tx/Tr/DDFH策略和未知扣费处理；越过最终期限后的无损晚到结果、持久队列与业务账本不属于现有库保证，若迁移依赖它们仍须在应用侧完成并验证。
- 旧ASA固定2001并回填请求Origin-State、MSCC覆盖顶层、缺省私有结果、截断token、缓冲成功当Diameter成功等不能成为新库正确性标准。保留能表达实际用途的输入/输出，缺陷修正必须列差异及迁移回归，不作悄悄不兼容。

### 32.4 标准与旧用法的优先关系

标准规范约束标准Profile；现有六模板/私有字段的兼容证据约束具名Profile。32.274失败/序号争议不禁止按已确认旧OCS路径完成失败TERM/退款兼容，也不能用普通成功路径代替这些必需用法。标准5011自动生成不是yppdia作为CTF使用全部现有路径的前置；5011安全接收仍归既定结果验收。

库可先实现源码已明确的表示、构造/解释及运行机制，未知Vendor/语义、真实绑定冲突和目标答案仍须取证；“必须cover”是交付目标，不把证据缺失变成已经通过。最终判定以覆盖§37逐行的协议能力与应用接线证据为准，§36最小子集只记中间进度。本轮不写代码，不新增全量标准审查、工作包或验收ID。


## 33. RO-06恢复契约收敛（2026-11-04）

本节只审sms-fw已有恢复分支，沿§32源码基线，不扩大到全部Ro/SCUR或重开首版specification。规范性机制只有[运行时§1.7](diameterkit-runtime-contracts.md)一份，原R33普通Route限制保留；新增原请求准备/受检续发是支持实际用法所必需的协议入口，非持久计费引擎。状态：设计已明确；实现、Go应用迁移及目标启用证据均未通过。

### 33.1 逐项源代码追溯与迁移判定

| 旧实际分支 / 源码 | 承接方案与明确差异 |
| --- | --- |
| [首次路由](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:1762)：HA首可用、非HA轮询；route的四模板槽/目的Host/Realm | Go应用按固定目标表选新操作，库校验TargetID和Ready；route全down返回NotWritten/不可用，旧流程不因此自动入队。实际四槽/版本绑定仍按§32/六模板验收 |
| [Peer上下线](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:2193)、[FailoverHandler/FailbackHandler](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:2608) | 库给稳定Peer/连接代次及可用快照；业务汇总路由可用性/告警。每个exchange独立检查目的及恢复域；不照搬按队列首条route搬全部请求；failback不自动重发在途消息 |
| [MessageTimeoutHandler](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:2512) | 普通CCR与DIA_SCAP1的ACR均进入计时分支；SCAP1独立固定恢复规则/关联，不能因不是标准CCR而遗漏；forceMessageDestroy按取消/终结处理，不自动续发 |
| [准备Tx/Tr](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:966)、[底层到期判断](../../sms-fw/c-cpp/libsrc/kernel/dialib/dia-base/dia-peer.cc:2112) | prepareForTr不重置tSend，首次Tr是t0+Tr；备用排队重置起点。新契约按AcceptedAt+ResponseWait计时，应用业务Tx独立；Tr≤Tx配置作为明确迁移差异拒绝 |
| [Tx开关](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:1621) | Tx返回DIAMETER_TX_TIMEOUT；开关关则Cancel并接受丢晚答，开则继续后台收答至Tr；不能把业务函数返回当整个请求已结束 |
| [Tr1/Tr2](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:1684) | 首次Tr显式选择合法备用，第二次Tr或无备用按配置DDFH保存/停止；库提供响应等待门，不必靠断开健康连接制造失败。所有次数/年龄有总上限 |
| [Tx前错误分类与backup](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:1311)、[错误码定义](../../sms-fw/c-cpp/libsrc/kernel/dialib/dia-base/dia-attr.hpp:205) | 标准临时集合3002/3004/3005；旧额外集合4002/5001/5004/5005/5008/5012仅具名兼容。实际错误先交Completion，再由应用续发新段；不在库内隐去原失败 |
| [无备用DDFH分支](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:1420) | 0返回实际答并保存；1也保存，decodeCCA(NULL)后又dealTmpErr，不能按日志写成“成功放行”。配置默认与答案覆盖保留Presence，应用决定业务状态 |
| [Tx后答案](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:1555) | 在原窗口内成功/永久失败结束恢复，3002/3004/3005保存；不复用Tx前额外六码集合。新库仍交唯一真实结果给应用后台消费，业务不二次完成 |
| [持久输入/答案/超时](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:1900)、[processPersistentAnswer](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:1061)、[processPersistentTimeoutTr](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:1212) | 应用每轮首选+至多一备；无Ready/临时错误/超时保留，最终2xxx及其他非临时3xxx、4/5xxx结束项但不抹掉失败。旧bool表示已处理，不是已扣费 |
| [队列删除/计数](../../sms-fw/c-cpp/libsrc/kernel/dialib/persqueue/PersistentQueueManager.cc:170)、[storeMsg](../../sms-fw/c-cpp/libsrc/kernel/dialib/persqueue/PersistentQueueManager.cc:298) | Go应用实现有限容量、间隔、最大次数、禁用/满/超限处理及明确保存失败；库不持久化。旧false会减少发送计数，不能照搬成跨重启无限额度 |
| [重启读队列](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:993) | 旧逻辑重置未保存的tCreated；新应用保存原CreatedAt/ReplayUntil和全局次数，禁止通过重启续命；旧文件由迁移工具转成受检协议数据，Node不加载文件 |
| [答案Session修补](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:2359) | 旧代码会以请求Session补空/错Session；新库按正常CCA或E错误各自语法和原请求关联检查，不静默修补。非法答案不授权续发为成功，差异进入迁移回归 |


wrapper返回本地错误后还有[yppdia MapError2Data](../../sms-fw/c-cpp/appsrc/kernel/yppdia/yppdiaengmx.cc:1581)这一层，不能遗漏PREPAID_FALLBACK配置。以下均是**Go应用的兼容业务映射**，不进入库Result-Code，也不证明OCS已扣费/退款；原协议结果/发送证据另行保留。五操作是CHECK/CHARGE/RESERVE/COMMIT/CANCEL，不把业务COMMIT与库应答Committed混淆。

| 本地原因与旧配置 | CHECK / CHARGE / RESERVE | COMMIT / CANCEL |
| --- | --- | --- |
| TX_TIMEOUT=0 | CHECK_FAILED / COMMIT_FAILED / OTHER_ERROR | COMMIT_FAILED / CANCEL_FAILED |
| TX_TIMEOUT=1 | CHECK_SUCCESSFUL / COMMITTED / RESERVATION_SUCCESSFUL | COMMITTED / CANCELLED |
| TX_TIMEOUT其他分支 | 三者NOT_CHARGED | COMMITTED / CANCELLED |
| CONNECTION_DOWN=0 | CHECK_FAILED / OTHER_ERROR / OTHER_ERROR | COMMIT_FAILED / CANCEL_FAILED |
| CONNECTION_DOWN非0 | 三者NOT_CHARGED | COMMITTED / CANCELLED |
| DIAMETER_REV_TMPERR且DIAMETER_TMP_ERROR=1 | 三者NOT_CHARGED | COMMITTED / CANCELLED |
| DIAMETER_REV_TMPERR且上述开关非1 | 旧switch落入成功分支，CHECK_SUCCESSFUL / COMMITTED / RESERVATION_SUCCESSFUL | COMMITTED / CANCELLED |

Tx/route-down分支的errorCause使用配置_logErrorIACC，拒绝分支另设_logLSI；临时错误fallback也保留其诊断。其他未单列错误进入SYSTEM_FAILURE分支，该参数影响LSI/errorCause设置，不能只按注释把所有失败改成成功。Go应用可纠正误导性业务状态名或配置行为，但必须显式记录迁移差异和回归，禁止把旧“成功状态”返写为协议2001。

### 33.2 jDiameter、OTP及规范的实际参考边界

- [jDiameter ClientRoSessionImpl，固定提交7f390934](https://github.com/RestComm/jdiameter/blob/7f3909342cca30ffc7e69b7d8b4111f5fde52ece/core/jdiameter/impl/src/main/java/org/jdiameter/client/impl/app/ro/ClientRoSessionImpl.java)提供Ro客户端事件/会话状态机、Tx处理以及按Requested-Action/DDFH区分的失败动作；CHECK、DEBIT、REFUND并非统一retry。其[LocalImpl](https://github.com/RestComm/jdiameter/blob/7f3909342cca30ffc7e69b7d8b4111f5fde52ece/core/jdiameter/impl/src/main/java/org/jdiameter/client/impl/app/ro/ClientRoSessionDataLocalImpl.java)把state、Tx和buffer放在内存字段，说明协议状态与持久化可以分离；不能据此宣称默认磁盘恢复、sms-fw六模板或本库预算已经得到验证。本库采用动作/原因分离，业务grant/deny和buffer动作由Go应用执行，不复制完整Java业务回调层。
- [OTP diameter_app](https://www.erlang.org/doc/apps/diameter/diameter_app.html)将选Peer、准备首次/重传和答案/错误交付分开；链路失败后的旧答案仍可能先完成调用，单次call至多交付一次answer或error。本库借鉴唯一完成和显式策略边界；OTP是通用Diameter接口，这份文档不证明它自带Ro Tx/Tr/DDFH或持久重放。新增ResponseWait与原身份导入是本库设计，不能冒称照搬OTP现成API。
- 固定Ro依据仍为[RFC4006 §5.7、§6.5、§7](https://www.rfc-editor.org/rfc/rfc4006.html#section-6.5)；一次事件换服务器与已有会话迁移分别处理，后者需会话许可。按[RFC6733 §3](https://www.rfc-editor.org/rfc/rfc6733.html#section-3)区别无答重传与错误后续发的T位。标准许可不证明具体OCS集群共享去重；旧额外错误集合和DDFH跨动作复用属于具名兼容，不上升为标准默认。

### 33.3 本次关闭与仍待证据

已明确：多OCS选择/迁移门槛、Tx/Tr时点、DDFH逐阶段分支、原请求发送前固定、错误后续发/重启重放身份、累计预算及旧/新答案竞争。现有§32“没有等价入口”的设计阻塞由运行时§1.7承接，禁止只改成“应用负责”后标覆盖。

仍待三类有限证据：①OCS目标集合的去重键/保存期/重启后原身份答案，以及ECUR会话连续性；②具名旧错误/DDFH/SCAP1规则与六模板实际binding；③Go应用保存失败、崩溃窗口、单发送者切换和最终未知结果处理。前两类阻塞对应目标启用/兼容声明，第三类阻塞应用迁移完成；公共库不提供它们的数据库实现。机制可按覆盖§38先实施和故障注入，本轮仍不写代码。

本次只读来源仍为sms-fw HEAD `80fa48c4126800cfa5c655e78889dffe6c03bd6c`，上述已登记源码哈希保持；补核`c-cpp/libsrc/kernel/dialib/persqueue/PersistentQueueManager.cc`，SHA-256 `0ee3dd8659fc700b23e33b49f9017e56f9d8176ce0f750c1beaeda624d0418b3`。这只证明本次静态来源，不证明现场binary或OCS行为。

## 34. Ro私有字段定义与420固定绑定（2026-11-04）

本节收敛Specification§7.1第一项，不扩展全部供应商字典。范围限§32已接线输入/输出、六模板及其配置；定义可实施与目标兼容已通过分别报告。源码基线仍为§32.1 HEAD；补核[DccaClientWrapperConfig.cc](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapperConfig.cc:943)，30195 bytes / SHA-256 `e6523f9adcd6bf9854b4c73486eda900743c7b915d28d4301dfd61dcf1fa3a67`：退款配置Getter只给一个Code，不给Vendor或路径。本次检索主dialib、yppdia及仓库C/C++配置/源码，20400/20430只找到常量与答案解析链，未找到能补足接收Vendor的定义；此结论不外推到生产OCS。

### 34.1 完整键、路径与来源

以下`V/C`为Vendor/Code；`0/C`表示V=0且**不编码Vendor-ID字段**，不是V=1、Vendor-ID=0。每层Vendor独立，不能从父Grouped、VSAI、CER厂商或宏名继承。源码模板writer给自定义AVP置M=1，非零Vendor置V=1、P=0；这是所列模板的发送形状，不证明CCA应收到相同flags。RFC标准420是Unsigned32，413才是Grouped金额；兼容字段必须有自己的具名定义，不能污染标准字典。[RFC6733 §4.1–4.1.1](https://www.rfc-editor.org/rfc/rfc6733.html#section-4.1)、[RFC4006 §8.21–8.22](https://www.rfc-editor.org/rfc/rfc4006.html#section-8.21)

| 实际字段 / 方向 | 已确认的完整父路径与类型 | 值/缺省及当前定义边界 |
| --- | --- | --- |
| PSP operation_result；正常CCA接收 | `CCA / ?/873 Grouped / ?/20400 Grouped / ?/20430 Integer32`；问号逐层表示未证Vendor，**不是通配符**。来自dia-cc.cc:1629–1654的查找意图 | 类型化值是有符号32位，零与缺失分开；源码不足以证明这条坏迭代路径实际读到了叶。三层Vendor/flags、出现次数、0/非0及缺失的OCS意义须由§34.4材料固定后才能启用该接收解释 |
| 旧退款token；CCR发送 | `CCR / 0/1292 OctetString`，顶层；refund模板:46，M=1/V=0/P=0 | `$REFUND_INFO`经Octets Getter取应用数据；不是文本，不转2022。接收配置改Code不修改模板发送Code，不能自动双向替换 |
| 旧退款token；CCA接收 | 旧逻辑支持`CCA / ?/configured-code`及`CCA / ?/456 / ?/configured-code`，均读OctetString；配置样例Code=1292。每个MSCC实例独立 | 接收Vendor/flags未查；不等于已证明0/1292。固定接收路径集合可包含顶层和MSCC，但须逐条有依据；重复值全部保留来源，不按遍历顺序覆盖，跨实例不自动选token |
| 标准退款凭据 | `CCR或CCA / 0/456 / 10415/2022 OctetString`，按既有标准过程/出现条件 | 与旧token并列、有独立Presence；不存在1292↔2022默认别名、复制或fallback。退款输入由应用明确关联原答案，不读业务存储 |
| 旧退款金额；CCR发送 | `CCR / 0/456 / 0/437 / 0/420 Grouped`；其`0/445 Grouped / {0/447 Integer64, 0/429 Integer32}`和直接`0/425 Integer32` | 此模板各成员各1个；字面值digits=1、exponent=0、currency=840，RSU另有`0/417 Unsigned64=1`。这是样例常量，不是库缺省、不证明真实退款金额来自CCA；旧425 INT32与标准Unsigned32定义分别记录 |
| 标准费用及单位输出 | `CCA / 0/423 / 0/445 / {0/447 Integer64,0/429 Integer32}`；`CCA / 0/423 / 0/425 Unsigned32`；`CCA / 0/431 / 0/417 Unsigned64` | 旧解析确实读顶层Cost-Information与顶层GSU，构造初值全0；标准逐MSCC费用/额度支持仍按原规则，不能将此旧读取路径误当已读每个MSCC。字段缺失保持Absent，0不是缺失标记，金额不变float |
| 通用模板TIMEZONE；CCR发送 | `CCR / 10425/23 Integer32`，字面0；模板:25 | 保留供应商号10425，不能“纠正”为10415或按名称改成字符串/时区偏移。0的供应商业务含义未证；字节构造已明，启用仍归原六模板证据包 |
| SCAP1已接线私有组 | ACR顶层`193/553 Grouped / {193/555 Integer32,193/554 UTF8String}`；`193/606 Grouped / {193/611 Enumerated,193/612 Grouped / 193/617 Unsigned64}`；`193/615 Integer32`、`193/610 Time`；重复`193/607 Grouped / {193/608 Integer32,193/609 UTF8String}` | Vendor/父路径/类型来自scap1模板；各607实例保序配对，不以608值覆盖。值与Getter沿§14.3–14.5；271/App0及VSAI193/19302分别固定，不借Ro App4的420规则。缺参数显式失败，不默认空串/0成功；实际ACA仍归SCAP1证据包 |

PSP未定义时，通用有界raw接收仍按既有未知AVP/M位和错误规则处理；它不授予PSP语义、不绕过未知必需字段拒绝，也不能完成依赖operation_result的CHECK。未证数据不进入标准或具名typed builder，不靠先填Vendor=0等现场“再看”实现覆盖。

表中标准费用/单位的`0/Code`是新标准Profile的定义；旧接收器同样按Code读取，不能据此宣称已经测到这些字段的实际Vendor。标准定义本身可实施，目标回包若采用其他命名空间须按上述证据包具名列明，不能复刻旧code-only宽松接受。

### 34.2 缺省、父实例与应用消费

1. [DIACCAnswer_t构造](../../sms-fw/c-cpp/libsrc/kernel/dialib/dia-cc/dia-cc.cc:1761)设operation_result=1、check_balance_result=-1、费用/单位=0、token长度=0。[真实CCA交付](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/DccaClientWrapper.cc:222)再逐项复制；因此“旧CHECK缺私有值默认成功”不成立。新库返回Absent/Present值及字段诊断，绝不伪造在位1或0；标准余额结果不替代PSP结果。
2. 旧PP_SUBMIT_CHECK在2000<base<3000时要求operation_result==0，最终只有2001成功；这只证明**应用判断**，不是供应商对所有整数的定义。Go迁移若保留该策略：有效CCA、base=2001且有效在位私有值=0才满足旧私有放行条件；Absent/畸形/未知语义不放行并带缺字段原因，非零原值保留，不自动映射DDFH、重试或Diameter错误码。标准结果和兼容校验仍分别成立；以后OCS证明缺失有其他含义也只能改具名应用策略，不改库Presence。
3. [WCreditControlAnswer构造](../../sms-fw/c-cpp/libsrc/kernel/dialib/dcca-wrapper/WCreditControlAnswer.hpp:26)只初始化token长度，未初始化operation_result/check_result；decodeCCA(NULL)还会提前返回“已入队成功”。这些路径不能套用DIACCAnswer默认1，更不能复制未初始化读取。新库以无Answer/发送事实交付，应用执行RO-06分支，不把入队或零值当协议成功。
4. token输出为owned bytes加Vendor/完整父路径/父实例来源，区分Absent、Present(empty)、Present(bytes)。库不使用strlen、旧1000字节wrapper截断或reservationId容量；沿现有MaxFieldBytes、MaxRepeated及答案预算，超限显式失败。空token是否可退款由具名过程规则/目标依据决定；不把旧`reservationId长度>25`变成协议最小token长度。多个候选即使bytes相同，也不能丢父实例；应用显式给原token及所属服务，无选择或相互冲突则不得构造假定退款。
5. 退款Getter链为CCA→owned费用/token→应用保存/关联→明确退款输入。旧CI_DIGIT_VALUE/CI_EXPONENT/CI_CURRENCY_CODE/GSU_SERVICE_UNIT确从reservationId中的refund_factors读取，见yppdiaengmx.cc:558/568/653/663；这些Getter的存在**不表示样例退款模板使用它们**，该模板金额/单位仍为常量。新库接收signed digits/exponent、uint32 currency、uint64 units及raw token；原S1布局、金额选择和是否退款归Go应用。旧无符号临时变量/INT64复用Uint64 writer不改变新公共金额的有符号类型，负值及边界进入编码验收。
6. 每个私有叶只在固定父实例内查找。多个20400、多个20430或顶层/MSCC token并存时，保留有界证据，按固定cardinality报告；未有“选首个/选最后个”依据时不提供单值成功解释。解析失败、错误Vendor、未知M=1、协议拒绝和业务不放行保持不同结果；不能让私有0覆盖CCA顶层失败或服务级失败。

### 34.3 420绑定的确定方案

沿既有Node→PeerPlan→ApplicationBinding→CommandPolicy和WireFingerprint，不增加运行时模板/字典注册器。启动前生成具名固定规则：App4、Command272、R/E分支、完整父路径及Vendor/Code确定类型；报文负载长度、Requested-Action、Origin/Destination和失败重试都不选择decoder。名称以下仅为文档中的解释域标签，不新增公开API类型。

| 固定解释域 | `CCR / 0/456 / 0/437 / 0/420`的唯一解释 | 413、其他父路径与限制 |
| --- | --- | --- |
| 标准Ro | Unsigned32 CC-Time，payload恰4字节 | 0/413为标准金额；旧420 Grouped不能构造/不能当合法CC-Time接收 |
| `smsfw-ro-legacy420-v1` | Grouped旧金额，按§34.1逐子字段固定；所有该CCR路径均如此，不仅REFUND时 | 同路径0/413可继续承载原通用模板金额，417承载事件单位；这只使已列五个CCR模板的金额路径可组合，不能证明整体类型已兼容：第四十四批确认2018在通用为Address、IEC/ECUR为STRING，仍需固定组合契约。该路径标准CC-Time不在支持集，若请求时间计费须明确拒绝/另域；不能宣称全标准Ro |
| 该兼容域其他CCR父路径/CCA | 无证据不扩展旧420别名；采用已列标准路径定义或明确不支持 | 不能把GSU/USU、顶层RSU或CCA的420全部改Grouped。后续实际证据若显示这些路径也需要变更，按同一份固定定义修订并重验，不靠每次消息猜型 |

这是有意选定的**非标准路径例外**，不是把不相容标准与兼容字典直接并集；生成器须在兼容域显式排除冲突位置的标准420，再形成该域的唯一规则。未知420不能降级raw来宣称支持两种类型。正常/E答案语法分别固定，退款CCA不推定要echo请求420。

部署选择只有两条：①目标确只需要上述窄SMS组合，所有Route候选引用同一经审计的固定组合Profile和答案定义；②同一父路径确需两种420类型，则采用真正独立、预先可鉴别的解释域/连接及Peer准入。第二条不能只改Route名、Requested-Action或PeerID；同邻接身份多连接仍遵守Peer重复连接/选举，必要时用独立Node/网络端点和真实身份管理。同身份多Node还须解决既有E2E/Session/Origin-State部署义务。DRA只给单一不可区分binding而又要求两型时返回BindingConflict，阻塞该部署组合；这是类型不可兼容，不是多抓一个包就能解除。

WireFingerprint包含这些固定路径/类型/flags/出现规则与版本。活动请求、RO-06原记录和备用目标不得切到另一种420解释；仅改变目的Host/Realm不能改变冻结的私有字段。标准与兼容builder可以不同，但不能让同一固定接收键有两个解释。类型组合可先离线实现；目标接受窄组合、实际槽映射和答案支持仍按原TEMPLATE-02验收。

### 34.4 收敛状态、外部证据与结束边界

**本轮已关闭设计歧义：** Vendor不继承、发送1292与接收配置分开、PSP逻辑父路径/有符号类型、旧默认1及无答案例外、token父实例/Presence/预算、常量金额与Getter区别、420窄组合或真实隔离。它们可实施；没有声称PSP Vendor已获得或任何目标互通已通过。其规范要求见[Specification §3.7](diameterkit-specification.md)，有限验收见[覆盖§40](diameterkit-usage-coverage.md)。

剩余仅收集三组材料，合并进§31.5六模板证据包，**不新增工作包、SPEC ID或全AVP矩阵**：

| 材料 | 必须回答的问题 / 退出条件 | 缺失时的准确影响 |
| --- | --- | --- |
| PSP答案定义 | 指定OCS/版本的字典或字段说明，配CHECK成功/非零失败/省略字段的关联请求答案；锁定873/20400/20430各层Vendor、flags、父路径、cardinality、signed值意义与缺省。多值/异常规则可按已定拒绝策略，不要求穷举所有整数 | 缺任一层Vendor不得启用PSP typed解释；省略的业务意义未证可保留明确的“不放行”应用策略，不能称已复现OCS省略语义。阻塞依赖该私有结果的完整CHECK兼容，标准余额检查与公共codec不受阻 |
| 旧token及费用链 | 指定实际route/global槽文件和接收Code配置；原IEC CCA→REFUND CCR→CCA成功/失败链，锁定接收Vendor/父实例及到发送0/1292的明确映射、空值规则、实际金额/单位是常量还是CI/GSU Getter | 源码明确的顶层0/1292及420构造可实施；接收定义/映射缺失不得猜Vendor、改码或选最后一个token，不能宣称完整原扣费→退款覆盖 |
| 420部署组合 | 指明同一解释域是否需要冲突路径CC-Time；确认窄SMS组合或真实隔离方案、全部候选Peer/目标、固定答案定义。复用六模板现有完整旧编码、独立解读、目标成功失败证据 | 无证据不宣称该组合互通；若两型必须共用不可区分binding则明确不支持该部署，需要部署变更，不能用动态decoder延期掩盖 |

一次抓包只能证明出现的形状，不能证明省略/枚举/重复字段的普遍语义；来源说明与成对报文互补。其余模板的TIMEZONE/SCAP1等已明字段继续原六包验收，不因本节加入新的全供应商调查。所选字段定义收敛即结束此项文档工作，之后只接收上述材料或实现发现的具体反例。

参考实现取舍：本轮核对[OTP diameter_dict](https://www.erlang.org/doc/apps/diameter/diameter_dict.html)的应用字典、逐AVP Vendor、类型/flags及Grouped定义；采用显式生成定义/构造期冲突检查的思路，不把OTP的字典配置当成它已支持sms-fw私有协议的证据。本库按路径的420兼容例外是本项目固定Profile设计，不能归因于OTP或用其他库字典推断未知供应商号。库只承担协议字段、固定绑定、校验、owned结果与有界资源；凭据保存、旧业务布局迁移、计费判断及重放持久记录仍归应用。

## 第四十九批执行更新：Ro结果安全

既有§9.4的正常CCA矩阵已落实到标准Ro和三个兼容Profile：5031要求诊断，基础5011/实验10415-5011分开；双结果/未知特性经过原请求关联后以不完整应用报告交付。E实验配对拒绝，自动正常配对仍禁用。只验证命令级诊断存在性/引用边界，不宣称rating或完整特性触发已验。详见[本批契约](../spec/history/batch-49.json)、[执行证据](../interop/evidence/2026-11-07-ro-result.json)和[进度](../spec/progress.json)。

## 第五十批执行更新：同OCS原身份恢复

§28.2/R33的受限恢复已在四个Ro Profile实施：传输停止触发，同一活动exchange、固定终端/来源和总期限，不因有效失败、实验冲突或DDFH另发请求。§33/运行时§1.7的专用恢复交接仍待实施，不能称完整RO-06已替换。详见[本批契约](../spec/history/batch-50.json)、[独立证据](../interop/evidence/2026-11-07-ro-recovery.json)和[进度](../spec/progress.json)。
