# 三应用实际字段、结果与Sh数据边界清单

> **当前范围（2026-11-04，Draft 0.20）：** 第一阶段=首版/v1，包含Sh、S6c、SGd/Gdd、Ro/DCCA及sms-fw实际所需六模板、SCAP1和RO-06恢复能力；Go 1.26、纯Go实现（含依赖与SCTP适配器，发布路径CGO_ENABLED=0）、公开Go库、无持久化边界不变。明确撤销第十九轮“Ro后置”分期；带日期的旧分期/统计仅作历史证据，不再授予Ro跳过首版验收的例外。限定S6a、完整SCUR/PRICE_ENQUIRY等未承诺能力仍后置。范围及剩余缺口见[Specification §3/§7](diameterkit-specification.md)，顺序见[总设计§14](diameterkit-design.md)，验收见[覆盖§39](diameterkit-usage-coverage.md)。2026-11-06已完成前十九批有界实施（Base/S6c基础、公共包络与关联、出站运行时、关闭与选举、TCP/TLS生产适配、有限Inbound/应答、有限Reject/应急预算、Canonical/有限重复缓存及多Arrival重答、固定同终端Route/多attempt与活动请求原身份重传、有限Dialer/Maintain拨号及重连、按class/编码字节DRR调度、固定Peer的有限Listener/Serve、有界标识源/Session/guard、Base/S6c有限Context/Report、S6c普通RDR/RDA及HSS ALR/ALA、SGd/Gdd普通OFR/OFA与TFR/TFA、Sh普通四组命令及透明User-Data、Ref24/25有限Sh-Data helper、Sh单订阅/退订及Ref11/12有限数据）；已执行范围及未完成项以[Specification §6](diameterkit-specification.md#6-需求与证据的维护)和[spec进度](../spec/progress.json)为准，未列明的运行验收仍Planned。

日期：2026-10-31，第三十一轮，Draft。适用于Go1.26公开、无持久化的Diameter端点库；Ro纳入首版实施。本页是[主流程审计](diameterkit-mainflow-audit.md)的可逐行核对清单，不是新字典DSL或已运行的测试。上位要求见[specification](diameterkit-specification.md)，能力状态仍Planned。

## 1. 范围、证据与读表方法

范围为本轮重新读取的isg-dia、isg-kernel注册模板、zdimanmx各Diameter handler及两个模拟器的三应用路径，加上前几轮明确选定的协议补齐分支。A/B沿用[业务覆盖基线](diameterkit-usage-coverage.md)：A为调用/消费，B为配置/模拟器；D表示本库已设计的标准补齐，不能冒称现网使用。C仅声明或未发现实际调用，不由此扩张支持。未证明两个仓库的所有运行配置或所有分支生产启用。

本轮实际HEAD：ip-sm-gw `01e18d514db8e1b957ff9e86b651f7bc9844dba5`；sms-fw `fda641246c507efe7f45067991a08e159622d2f8`。均按当前工作树只读审查，不能把旧覆盖基线的HEAD或当前HEAD当作未提交文件的版本保证；下表另固定关键文件摘要。规范继续使用主流程已登记的29.328 V18.1.0、29.329 V18.0.0、29.338 V18.2.0、29.336 V18.2.0、29.173 V18.0.0及29.272 V18.3.0。复取23.040 V18.0.0哈希与原记录一致。没有升级Profile或改业务仓库。

| 证据键 | 读取入口与用途 | 本轮文件SHA-256 |
| --- | --- | --- |
| S1 | [UDR](../../ip-sm-gw/go/svc/isg-dia/internal/handlers/diameter-handler/ud.go) | `6ee4737e1f6ade0dee039abd2ef2d43d020105cbf6fda62e47c9c2af3202560f` |
| S2 | [PUR](../../ip-sm-gw/go/svc/isg-dia/internal/handlers/diameter-handler/pu.go) | `218236b76ef0820c6517e5c840df430d09266b17531a40dd77db1c3d2ced26e0` |
| S3 | [SNR](../../ip-sm-gw/go/svc/isg-dia/internal/handlers/diameter-handler/sn.go) | `38523533fe709cdf0d96aaf7a3e9b53672c9b5e8a248b5ca6a87720c9fc4b9b7` |
| S4 | [PNR/PNA](../../ip-sm-gw/go/svc/isg-dia/internal/handlers/diameter-handler/pnr.go) | `e5f5013753b1972a837ae6ad47cf0f68ac56f4db6e7c2baa67768be8c310dd47` |
| S5 | [注册/注销XML](../../ip-sm-gw/go/svc/isg-kernel/internal/handler/diameter-handler/stream-handler.go) | `5d652639d78de79b1a26d345bb395283d63d30434c02558fe72fc7887f5c49af` |
| S6 | [SRA节点/结果](../../sms-fw/go/svc/zdimanmx/internal/handler/diameter-interface-handler/sra-handler.go) | `873e17f69d6174d9027c7798f19813d26b6450014da0e95126889f28a434f176` |
| S7 | [TFR](../../sms-fw/go/svc/zdimanmx/internal/handler/diameter-interface-handler/tfr-handler.go) | `0a04f9e79431af720566b66ff08aa39b08d788047ad134e00de5d90c82534cf0` |
| S8 | [TFA](../../sms-fw/go/svc/zdimanmx/internal/handler/diameter-interface-handler/tfa-handler.go) | `27bb0ca1d9835db6fa66f4f4bc7eacace43c3d04b91c4b5866141dd5e6259fb0` |
| S9 | [RDR](../../sms-fw/go/svc/zdimanmx/internal/handler/diameter-interface-handler/rdr-handler.go) | `4955dee4fbce047488502d5a8b75947372344c526afe57e428c228f45a4dd6fd` |
| S10 | [IP-SM-GW RDR](../../ip-sm-gw/go/svc/isg-dia/internal/handlers/diameter-handler/rd.go) | `cadf56bf54c254e6a122f02c91d1c0be43298f81c5c866c108eaa20319062a4a` |
| S11 | [OFR](../../sms-fw/go/svc/zdimanmx/internal/handler/diameter-interface-handler/ofr-handler.go) | `f17dea204232668f09a2820ded609fe2ef7a50bdec5e4f28d550ff023922c9e9` |
| S12 | [OFA业务错误](../../sms-fw/go/svc/zdimanmx/internal/handler/submit-handler/submit-reaad-queue.go) | `41a45a3a55d1fc16b47c18f21aafb2bd31727c423dbcdc73128018c2e8a23ba9` |
| S13 | [ALR模拟](../../ip-sm-gw/go/svc/mme-simulator/internal/diameter-handler/handle-ala.go) | `774fcd409fb0ada9034c869e464d9cb1eb20c97f4939cf85e83d1eb1e9f364bf` |
| S14 | [Sh/S6c模拟](../../ip-sm-gw/sim/sh_server/server.go) | `48a05ae5491d4b26300269d7ee76d16d1668ebf20de89892b119308b96dd8969` |
| S15 | [SRR构造](../../sms-fw/go/svc/zdimanmx/internal/handler/diameter-interface-handler/srr-handler.go) | `050b893595e3b97df8ba1347408cf7bbb6ae5c19d9858d2ea0e8225dd8c592f9` |
| S16 | [UDA接收](../../ip-sm-gw/go/svc/isg-dia/internal/handlers/diameter-handler/uda.go) | `cc62a939e5fd63b9b04d27817229b62a67145018cfa6c2088c27a8d8b0fc822e` |
| S17 | [PUA接收](../../ip-sm-gw/go/svc/isg-dia/internal/handlers/diameter-handler/pua.go) | `5cded27ec82e5847e40338e95bd533d3e83d4e991257c2ca33f0e9b57f9b261d` |
| S18 | [SNA接收](../../ip-sm-gw/go/svc/isg-dia/internal/handlers/diameter-handler/sna.go) | `3a048e7745fbf94b4c2162b342a6673a8b4ee9e73875de10e4fcdfe6903e27b9` |
| S19 | [OFA业务应答](../../sms-fw/go/svc/zdimanmx/internal/handler/diameter-interface-handler/ofa-handler.go) | `7014ba562254f045df2c4a5607afea33b74f25d8aa68ecd8b800f6b89dd4fec2` |

表中`Q`为三应用的请求集合，`A`为相应正常答案，`M`为两者；`ShQ`为UDR/PUR/SNR/PNR。每个集合都是有限展开，不表示该字段可出现在任意命令；E位基础错误按RFC6733错误语法单列。`UI-Sh`=ShQ/User-Identity700；`UI-SMS`=OFR/RDR/ALR及SRA/RDA/TFA适用时的User-Identifier3102。大写A证据与路径集合A分属不同列。

每行只列一个叶定义；同一叶在不同父组/不同flags下分行。父组数量与必需性见§2，叶次数相对其直接父实例，不能误认为整个消息次数。已选分支的条件比ABNF可选性严格；标准/兼容差异沿用主流程§18；跨组/时间见§20，特性/T4/SMEA及当前勘误见§21。基本字段若缺失、重复、类型宽度错误，由固定规则报告，不能拿业务数据库事实补足。

预算记号：`T`为原始文本字节，`O`为octets，二者均先计入MaxFieldBytes/MaxMessageBytes；`N`为4字节数值/Time，AVP尺寸为`align4(8+4V+payload)`。每个可识别叶、父组及扩展计AVP总数/已知Grouped深度；未知opaque不猜子项，诊断例外见主流程§25.3，重复集合同时受数量上限；源raw与解释结果并存计owned预算。XML字节再受显式XML解析预算，不把字符数当字节数。不新增每叶一套默认额度，沿用现有Q/M/I资源分类；新预算默认仍待测量。

错误分层：类型/宽度/UTF8等静态违规、已知应用条件违规、未知枚举/未支持语义、缺原请求/终端特性四者分开。Sh/SMS对**已知**AVP的M位接收例外遵守各AVP表注释；下列V/M是发送要求，不据M位不匹配一律拒绝。未知M=1按原规则处理，安全保留不等于语义通过。候选列引用[旧报文样本](diameterkit-protocol-vectors.md)、[第二十四/二十五轮完整帧](diameterkit-leaf-vectors.md)或验收ID；未有完整bytes写明“待补”，不因列表齐全宣称独立fixture齐全。

## 2. 容器和命令条件索引

公共叶依据RFC6733 §6–8；Sh命令29.329 §6.1、过程29.328 §6.1；SMS命令29.338 §5.3.2/6.3.2、过程§5.2/6.2。完整顶层ABNF盘点仍唯一维护于主流程§2，本页不重复一份。

| 父容器 / Vendor-Code / 发送V/M | 在所选命令中的数量与条件 |
| --- | --- |
| Vendor-Specific-Application-Id 0/260，0/1 | Sh正常请求/答案各1，认证App16777217；S6c/SGd正常发送不带、接收按SMS-VSAI忽略而仍计预算；不是所有App的通用禁令 |
| Experimental-Result 0/297，0/1；Failed-AVP 0/279，0/1 | 正常答案结果容器0..1，与268按三应用结果规则；错误涉及字段时构建有界Failed-AVP原路径，不递归执行其内容 |
| Proxy-Info 0/284，0/1；Supported-Features 10415/628，V1/M按角色 | 可重复；Proxy顺序保留；Sh实际用特性的请求628发M1，纯接收报价可M0，答案M0；子629/630仍M0。按应用/终端/命令对及Vendor/List-ID解释，见主流程§21.1 |
| User-Identity 10415/700，1/1 | ShQ各1；所选访问键三选一，Private Identity为顶层User-Name，不是本组成员 |
| User-Identifier 10415/3102，1/1 | OFR/RDR/ALR各1；SRA/RDA/TFA按过程可选；各叶数量不表示可以空组，SMS允许的双身份不套Sh互斥规则 |
| Serving-Node 10415/2401；Additional-Serving-Node 10415/2406，均1/1 | SRA各0..1；成功投递查询需要Serving-Node，非投递用途例外；ALR移动且新节点已知时带2401。两组不得重复同一节点信息；普通HSS双节点受SRR bit0及SGSN+MSC/MME之一限制，见主流程§20.2；Realm兼容仅覆盖SRA直接2401 |
| SM-Delivery-Outcome 10415/3316，1/1 | RDR必需；IMS特殊OFR条件必需；各子节点3317/3318/3319/3320各0..1且1/1。RDR至少一cause、MSC/MME互斥；不能由全可选组语法放行空报告 |
| SMSMI-Correlation-ID 10415/3324，1/0 | SRR/RDR/ALR/OFR/TFR各0..1，按IMS无MSISDN条件出现；三叶在不同命令的位置条件不同 |
| SM-Delivery-Failure-Cause 10415/3303，1/1 | OFA/TFA各0..1；实验5555时必需，子cause必需，不能把MO原因用于MT |
| EPS-Location-Information10415/1496→MME-Location-Information10415/1600，均1/0 | OFR/TFA各0..1；本轮既有D子集仅ECGI/TAI。其他深层位置组不是现有A/B消费证据，不把这两叶当全EPS支持 |

## 3. 逐叶字段

### 3.1 公共叶

来源：RFC6733 §6.3–6.13、§7；M位与容器规则按§2。

| LeafID | 完整路径或已定义路径集合 | Vendor/Code | 类型/预算 | V/M发送值 | 次数 | 条件/检查 | 使用证据 | 候选/验收 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| L001 | M/Session-Id | 0/263 | UTF8/T | 0/1 | 1 | 首位；回原请求，不能沿用旧UDR任意加前缀作为库策略 | A S1–S12；B S13/S14 | PAIR-05/07 |
| L002 | M/Auth-Session-State | 0/277 | Enum/N | 0/1 | 1 | 三应用所选值1；E错误不强制 | A S1–S12；B S13/S14 | PAIR-05/07 |
| L003 | M/Origin-Host | 0/264 | Identity/T | 0/1 | 1 | 授权终端身份，不复制请求Destination | A S1–S12；B S13/S14 | PAIR-05/07 |
| L004 | M/Origin-Realm | 0/296 | Identity/T | 0/1 | 1 | 原字节保留；不从相邻DRA获得HSS业务身份 | A S1–S12；B S13/S14 | PAIR-05/07 |
| L005 | M/Origin-State-Id | 0/278 | U32/N | 0/1 | 0..1 | 当前Origin的状态值；缺省/显式0保留，答案不回显；非零单调归部署，无库内持久计数 | B S14 | PAIR-05/07 |
| L006 | Q/Destination-Host | 0/293 | Identity/T | 0/1 | 0..1 | PNR/TFR为1；HSS发ALR过程也必需，其他按目的绑定/过程 | A S1–S3/S7 | PAIR-05/07 |
| L007 | Q/Destination-Realm | 0/283 | Identity/T | 0/1 | 1 | 目标终端realm；不由DRA邻接自动替换 | A S1–S3/S7 | PAIR-05/07 |
| L008 | A/Result-Code | 0/268 | U32/N | 0/1 | 0..1 | 成功2001；与实验结果不覆盖彼此 | A公共；代理为D | RESULT-01/02、PAIR-05/07 |
| L009 | A/Experimental-Result/Vendor-Id | 0/266 | U32/N | 0/1 | 1 | 采用3GPP错误时10415 | A公共；代理为D | RESULT-01/02、PAIR-05/07 |
| L010 | A/Experimental-Result/Experimental-Result-Code | 0/298 | U32/N | 0/1 | 1 | 值/应用/命令联合解释，见§4 | A公共；代理为D | RESULT-01/02、PAIR-05/07 |
| L011 | Sh M/Vendor-Specific-Application-Id/Vendor-Id | 0/266 | U32/N | 0/1 | 1 | 10415 | A公共；代理为D | RESULT-01/02、PAIR-05/07 |
| L012 | Sh M/Vendor-Specific-Application-Id/Auth-Application-Id | 0/258 | U32/N | 0/1 | 1 | 16777217；不得附Acct-App形成双用途 | A公共；代理为D | RESULT-01/02、PAIR-05/07 |
| L013 | M/Proxy-Info/Proxy-Host | 0/280 | Identity/T | 0/1 | 1 | 每个父组保留一次 | D 公共DRA字段，非业务handler消费 | RESULT-01/02、PAIR-05/07 |
| L014 | M/Proxy-Info/Proxy-State | 0/33 | Octets/O | 0/1 | 1 | 透明octets，非UTF8；不依文本格式去重 | D 公共DRA字段，非业务handler消费 | RESULT-01/02、PAIR-05/07 |
| L015 | M/Route-Record | 0/282 | Identity/T | 0/1 | 0..N | raw保序和重复；答案不强制回显；顶层0/282按R17剥离，准入独立 | D 公共DRA字段，非业务handler消费 | RESULT-01/02、PAIR-05/07 |
| L016 | M/Supported-Features/Vendor-Id | 0/266 | U32/N | 0/1 | 1 | 按App+Vendor+List解释，重复键冲突不后写覆盖；Sh双方终端特性，不是CER | D SH-NOTIF-EFF | V21-SUBMULTI、SUB-02 |
| L017 | M/Supported-Features/Feature-List-ID | 10415/629 | U32/N | 1/0 | 1 | 按App+Vendor+List解释，重复键冲突不后写覆盖；Sh双方终端特性，不是CER | D SH-NOTIF-EFF | V21-SUBMULTI、SUB-02 |
| L018 | M/Supported-Features/Feature-List | 10415/630 | U32/N | 1/0 | 1 | 按App+Vendor+List解释，重复键冲突不后写覆盖；Sh双方终端特性，不是CER | D SH-NOTIF-EFF | V21-SUBMULTI、SUB-02 |
| L019 | A/Error-Message | 0/281 | UTF8/T | 0/0 | 0..1 | UTF8有效且禁止U+0000；无路由/重试含义，显示有界转义 | D基础错误 | F25-SH-BASE5001；N25-L019-TEXT/DUP |
| L020 | A/Error-Reporting-Host | 0/294 | Identity/T | 0/0 | 0..1 | 错误来源，非自动信任节点 | D基础错误 | F25-SH-BASE5001；N25-L020-TEXT/DUP |

### 3.2 Sh采用字段与深层功能选择键

来源：29.329 §6.1/6.3及29.328 §6.1/表7.6.1。C边界行只说明结构/诊断，不宣布Ref0/10/13/19完整helper支持。

| LeafID | 完整路径或已定义路径集合 | Vendor/Code | 类型/预算 | V/M发送值 | 次数 | 条件/检查 | 使用证据 | 候选/验收 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| L021 | UI-Sh/Public-Identity | 10415/601 | UTF8/T | 1/1 | 0..1 | 所选公共身份/普通IMPU；语法与实际身份授权分开 | A S1–S4 | SUB-01、FIELD-05 |
| L022 | UI-Sh/MSISDN | 10415/701 | Octets/O | 1/1 | 0..1 | TBCD；与其他访问键按所选数据互斥 | A S3/S4 | SUB-01、FIELD-05 |
| L023 | UI-Sh/External-Identifier | 10415/3111 | UTF8/T | 1/1 | 0..1 | 固定23.003/4282子集；非空local@domain，见§6 | D FIELD-05 | SUB-01、FIELD-05 |
| L024 | ShQ/User-Name | 0/1 | UTF8/T | 0/1 | 0..1 | Private Identity只在表7.6.1适用数据；Note9排除External键叠加；混合11/12/25不带 | D SH-PRIVATE-KEY | V21-UNSUB、SUB-02 |
| L025 | UDR/PUR/SNR/Data-Reference | 10415/703 | Enum/N | 1/1 | 1..N | A为24注册/25可达；11/12与混合为D；枚举不当U32语义 | A S1–S3；D | PAIR-05、SUB-01/02 |
| L026 | PUA/Data-Reference | 10415/703 | Enum/N | 1/1 | 0..1 | 仅协商Update-Eff-Enhance且多DataRef更新失败时定位错误项；非成功回显 | D | N24-PUA-DATAREF；完整多Ref失败按需求补 |
| L027 | PUR/PNR/UDA/SNA/User-Data | 10415/702 | Octets/O | 1/1 | 0..1 | PUR/PNR为1；UDA按返回数据，SNA按SendData/结果；显式XML解析见§5 | A S2/S4/S5；D | XML-02/03 |
| L028 | SNR/Subs-Req-Type | 10415/705 | Enum/N | 1/1 | 1 | 0订阅/1退订，非Expiry=0 | A S3；D | SUB-01/02 |
| L029 | SNR/SNA/Expiry-Time | 10415/709 | Time/N | 1/0 | 0..1 | NTP绝对秒/era；SNA缺省无限期，不回填请求；更早值/缺702的SHOULD见主流程§20.4 | A S3；D | V24-SH-SN-A、C24-SNA-*、N24-SNA-* |
| L030 | SNR/One-Time-Notification | 10415/712 | Enum/N | 1/0 | 0..1 | 25所需值0；只限定25 | A S3；D | SUB-02 |
| L031 | SNR/Send-Data-Indication | 10415/710 | Enum/N | 1/0 | 0..1 | 0/1；本轮混合25子集不带，包括0 | D | SUB-01/02 |
| L032 | UDR/SNR/Service-Indication | 10415/704 | Octets/O | 1/1 | 0..N | Ref0选择键；不是XML字符串/base64自动转换 | C S14；D边界 | XML-03 |
| L033 | UDR/SNR/Server-Name | 10415/602 | UTF8/T | 1/1 | 0..1 | Ref13/19选择键；SIP URL，不是DiameterIdentity | C S14；D边界 | XML-03 |
| L034 | UDR/SNR/Identity-Set | 10415/708 | Enum/N | 1/0 | 0..N | Ref10适用；多值需Notif-Eff相应规则 | C S14；D边界 | N25-L034-WIDTH3；Ref10完整helper未启用 |
| L035 | UDR/SNR/DSAI-Tag | 10415/711 | Octets/O | 1/1 | 0..N | Ref19键；XML同名元素为string，不隐式字节转换 | C S14；D边界 | XML-03 |
| L036 | PUA/Repository-Data-ID/Service-Indication | 10415/704 | Octets/O | 1/1 | 1 | 父715为1/0且0..1；非XML ServiceIndication；715仅相应多Repository更新失败定位，非成功必备 | D边界 | XML-03；完整失败定位按需求补 |
| L037 | PUA/Repository-Data-ID/Sequence-Number | 10415/716 | U32/N | 1/0 | 1 | AVP U32；XML序号只有0..65535，不能混用；715仅相应多Repository更新失败定位，非成功必备 | D边界 | XML-03；完整失败定位按需求补 |

### 3.3 S6c与跨应用身份

来源：29.338 §5.2/5.3.3、29.336 §6.4.2、29.173 LMSI定义；用途解释沿用SMS-SRA-ID和各SMS-ID规则。

| LeafID | 完整路径或已定义路径集合 | Vendor/Code | 类型/预算 | V/M发送值 | 次数 | 条件/检查 | 使用证据 | 候选/验收 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| L038 | UI-SMS/User-Name | 0/1 | UTF8/T | 0/1 | 0..1 | OFR可用IMSI；IMS RDR为HSS-ID；ALR的Router token独立语义 | A S11；B S13；D IMS身份 | IDENTITY-01/02 |
| L039 | UI-SMS/MSISDN | 10415/701 | Octets/O | 1/1 | 0..1 | OFR普通源号码；IMS特定OFR/ALR才允许15个0 dummy | A S9–S11；D dummy身份 | IDENTITY-01/02 |
| L040 | UI-SMS/External-Identifier | 10415/3111 | UTF8/T | 1/1 | 0..1 | T4等已选位置，NAI有限解释 | D 身份补齐；所列A/B handler未直接消费 | IDENTITY-01/02 |
| L041 | UI-SMS/LMSI | 10415/2400 | Octets/O | 1/1 | 0..1 | 四octets，非整数；存在条件按身份/过程 | D 身份补齐；所列A/B handler未直接消费 | IDENTITY-01/02 |
| L042 | SRR/MSISDN | 10415/701 | Octets/O | 1/1 | 0..1 | 普通查询用MSISDN，无MSISDN按用途用User-Name | A SRR构造 | V-SRR-01 |
| L043 | SRR/User-Name | 0/1 | UTF8/T | 0/1 | 0..1 | T4用IMSI；IMS重试用HSS-ID，不从位数猜用途 | D | D21-T4-SRR、V21-IMS-SRR |
| L044 | SRA/User-Name | 0/1 | UTF8/T | 0/1 | 0..1 | 成功条件必需；IMSI/Router关联/只取MCCMNC；无IMSI缺口保留 | A S6；D | IDENTITY-02 |
| L045 | SRR/SM-RP-MTI | 10415/3308 | Enum/N | 1/1 | 0..1 | 0 DELIVER、1 STATUS-REPORT，不能使用2400 | A SRR构造 | V-SRR-01 |
| L046 | SRR/SM-RP-SMEA | 10415/3309 | Octets/O | 1/1 | 0..1 | 支持接收双节点时条件必需；长度+TOA+数字TBCD的有限子集，不同于SC裸TBCD | D边界 | V25-SRR-TWO-R；N25-SMEA-*；数字子集见主流程§21.3 |
| L047 | SRR/SRR-Flags | 10415/3310 | U32/N | 1/1 | 0..1 | bit0允许双节点、bit1 SM-RP-PRI、bit2 SingleAttempt；未知位发送0，接收raw保留但不解释 | D | V24-S6C-SR-R、C24-SRR-NOGPRS-R |
| L048 | SRR/SM-Delivery-Not-Intended | 10415/3311 | Enum/N | 1/1 | 0..1 | 0仅IMSI、1仅MCCMNC，不能更改投递请求意图 | D | IDENTITY-02 |
| L049 | SRA/LMSI | 10415/2400 | Octets/O | 1/1 | 0..1 | 恰4字节；成功含MSC且VLR已用LMSI时出现；接收端无VLR事实不猜违例 | D | V24-SRA-SGSN-MSC、N24-SRA-LMSI3、C24-SRA-NOLMSI |
| L050 | SRA/MWD-Status | 10415/3312 | U32/N | 1/1 | 0..1 | bit0 SC未加入、1 MNRF、2 MCEF、3 MNRG、4 MNR5G；0存在，未知位不解释，不作库内MWD状态 | D | V24-SRA-SGSN-MSC |
| L051 | RDR/RDR-Flags | 10415/3323 | U32/N | 1/0 | 0..1 | bit0值1为SingleAttempt；值0不是SingleAttempt | A S10 | PAIR-04/07 |
| L052 | ALR/Maximum-UE-Availability-Time | 10415/3329 | Time/N | 1/0 | 0..1 | 已知时带UTC可达截止；过去时间不自动编码失败，不建调度任务 | D | V24-S6C-AL-R、N24-ALR-TIME3-R、C24-ALR-EXPIRED-R |
| L053 | ALR/SMS-GMSC-Alert-Event | 10415/3333 | U32/N | 1/0 | 0..1 | MME/SGSN必需；bit0可达、bit1移动，HSS角色不强制 | B S13；D | ALERT-02 |
| L054 | SRA/MME-Absent-User-Diagnostic-SM | 10415/3313 | U32/N | 1/1 | 0..1 | 各域已知缺席诊断；0存在≠缺省；值链见§3.6 | A S6 | V25-DIAG-MME；C25/N25-DIAG-MME-*；N25-L054-* |
| L055 | SRA/MSC-Absent-User-Diagnostic-SM | 10415/3314 | U32/N | 1/1 | 0..1 | 各域已知缺席诊断；0存在≠缺省；值链见§3.6 | A S6 | V24-SRA-ZERO、C24-SRA-NODIAG |
| L056 | SRA/SGSN-Absent-User-Diagnostic-SM | 10415/3315 | U32/N | 1/1 | 0..1 | 各域已知缺席诊断；0存在≠缺省；值链见§3.6 | A S6 | V25-DIAG-SGSN；C25/N25-DIAG-SGSN-*；N25-L056-* |

### 3.4 节点和Outcome逐父路径展开

来源：29.338 §5.2.1/2/3及§5.3.3.6–20；节点复用29.173/272/336，S6c覆盖的1645/2408等M位不能继承其他App默认。

| LeafID | 完整路径或已定义路径集合 | Vendor/Code | 类型/预算 | V/M发送值 | 次数 | 条件/检查 | 使用证据 | 候选/验收 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| L057 | SRA/Serving-Node/MME-Name | 10415/2402 | Identity/T | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | A S6 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L058 | SRA/Serving-Node/MME-Realm | 10415/2408 | Identity/T | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | A S6 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L059 | SRA/Serving-Node/MME-Number-for-MT-SMS | 10415/1645 | Octets/O | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | A S6 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L060 | SRA/Serving-Node/MSC-Number | 10415/2403 | Octets/O | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | A S6 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L061 | SRA/Serving-Node/SGSN-Name | 10415/2409 | Identity/T | 1/0 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | A S6 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L062 | SRA/Serving-Node/SGSN-Realm | 10415/2410 | Identity/T | 1/0 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | A S6 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L063 | SRA/Serving-Node/SGSN-Number | 10415/1489 | Octets/O | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | A S6 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L064 | SRA/Additional-Serving-Node/MME-Name | 10415/2402 | Identity/T | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | S6有解析函数，成功主路未调用；D保留 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L065 | SRA/Additional-Serving-Node/MME-Realm | 10415/2408 | Identity/T | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | S6有解析函数，成功主路未调用；D保留 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L066 | SRA/Additional-Serving-Node/MME-Number-for-MT-SMS | 10415/1645 | Octets/O | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | S6有解析函数，成功主路未调用；D保留 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L067 | SRA/Additional-Serving-Node/MSC-Number | 10415/2403 | Octets/O | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | S6有解析函数，成功主路未调用；D保留 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L068 | SRA/Additional-Serving-Node/SGSN-Name | 10415/2409 | Identity/T | 1/0 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | S6有解析函数，成功主路未调用；D保留 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L069 | SRA/Additional-Serving-Node/SGSN-Realm | 10415/2410 | Identity/T | 1/0 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | S6有解析函数，成功主路未调用；D保留 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L070 | SRA/Additional-Serving-Node/SGSN-Number | 10415/1489 | Octets/O | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | S6有解析函数，成功主路未调用；D保留 | V24-SRA-*、N24-ADD-*；NODE-02 |
| L071 | ALR/Serving-Node/MME-Name | 10415/2402 | Identity/T | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | D ALERT-SGSN/MME | 既有V-ALR-MOVE/V21-SGSN-ALR；ALERT-02 |
| L072 | ALR/Serving-Node/MME-Realm | 10415/2408 | Identity/T | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | D ALERT-SGSN/MME | 既有V-ALR-MOVE/V21-SGSN-ALR；ALERT-02 |
| L073 | ALR/Serving-Node/MME-Number-for-MT-SMS | 10415/1645 | Octets/O | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | D ALERT-SGSN/MME | 既有V-ALR-MOVE/V21-SGSN-ALR；ALERT-02 |
| L074 | ALR/Serving-Node/SGSN-Name | 10415/2409 | Identity/T | 1/0 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | D ALERT-SGSN/MME | 既有V-ALR-MOVE/V21-SGSN-ALR；ALERT-02 |
| L075 | ALR/Serving-Node/SGSN-Realm | 10415/2410 | Identity/T | 1/0 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | D ALERT-SGSN/MME | 既有V-ALR-MOVE/V21-SGSN-ALR；ALERT-02 |
| L076 | ALR/Serving-Node/SGSN-Number | 10415/1489 | Octets/O | 1/1 | 0..1 | 号码按TBCD，Name/Realm按Identity；按该过程成组出现，不逐叶任选；不从Name补Realm | D ALERT-SGSN/MME | 既有V-ALR-MOVE/V21-SGSN-ALR；ALERT-02 |
| L077 | SRA/Serving-Node/IP-SM-GW-Number | 10415/3100 | Octets/O | 1/1 | 0..1 | N/N+H明确；N+H+R仅s6c-ip-sm-gw-realm-v1，3112非OctetString | A S6；D兼容 | NODE-01 |
| L078 | SRA/Serving-Node/IP-SM-GW-Name | 10415/3101 | Identity/T | 1/1 | 0..1 | N/N+H明确；N+H+R仅s6c-ip-sm-gw-realm-v1，3112非OctetString | A S6；D兼容 | NODE-01 |
| L079 | SRA/Serving-Node/IP-SM-GW-Realm | 10415/3112 | Identity/T | 1/1 | 0..1 | N/N+H明确；N+H+R仅s6c-ip-sm-gw-realm-v1，3112非OctetString | A S6；D兼容 | NODE-01 |
| L080 | SRA/Serving-Node/IP-SM-GW-SBI-Support-Indicator | 10415/3348 | Enum/N | 1/0 | 0..1 | 按SBI定义；不与Realm兼容三元组任意组合，不实现HTTP/SBI | D边界 | V25-SBI-N0/NH1；C25-SBI-UNKNOWN；N25-L080-* |
| L081 | RDR（IP-SM-GW另含IMS OFR）/SM-Delivery-Outcome/MME-SM-Delivery-Outcome/SM-Delivery-Cause | 10415/3321 | Enum/N | 1/1 | 0..1 | 0容量、1缺席、2成功；至少一个cause；MME/MSC互斥 | A S9 | PAIR-04/07、RESULT-06 |
| L082 | RDR（IP-SM-GW另含IMS OFR）/SM-Delivery-Outcome/MME-SM-Delivery-Outcome/Absent-User-Diagnostic-SM | 10415/3322 | U32/N | 1/1 | 0..1 | 只与对应节点缺席cause配套；0也有效；值链见§3.6 | A S9 | PAIR-04/07、RESULT-06 |
| L083 | RDR（IP-SM-GW另含IMS OFR）/SM-Delivery-Outcome/MSC-SM-Delivery-Outcome/SM-Delivery-Cause | 10415/3321 | Enum/N | 1/1 | 0..1 | 0容量、1缺席、2成功；至少一个cause；MME/MSC互斥 | D | V24-S6C-RD-R、C24/N24-RDR-* |
| L084 | RDR（IP-SM-GW另含IMS OFR）/SM-Delivery-Outcome/MSC-SM-Delivery-Outcome/Absent-User-Diagnostic-SM | 10415/3322 | U32/N | 1/1 | 0..1 | 只与对应节点缺席cause配套；0也有效；值链见§3.6 | D | V24-S6C-RD-R、C24/N24-RDR-* |
| L085 | RDR（IP-SM-GW另含IMS OFR）/SM-Delivery-Outcome/SGSN-SM-Delivery-Outcome/SM-Delivery-Cause | 10415/3321 | Enum/N | 1/1 | 0..1 | 0容量、1缺席、2成功；至少一个cause；MME/MSC互斥 | A S9 | PAIR-04/07、RESULT-06 |
| L086 | RDR（IP-SM-GW另含IMS OFR）/SM-Delivery-Outcome/SGSN-SM-Delivery-Outcome/Absent-User-Diagnostic-SM | 10415/3322 | U32/N | 1/1 | 0..1 | 只与对应节点缺席cause配套；0也有效；值链见§3.6 | A S9 | PAIR-04/07、RESULT-06 |
| L087 | RDR（IP-SM-GW另含IMS OFR）/SM-Delivery-Outcome/IP-SM-GW-SM-Delivery-Outcome/SM-Delivery-Cause | 10415/3321 | Enum/N | 1/1 | 0..1 | 0容量、1缺席、2成功；至少一个cause；MME/MSC互斥 | A S9/S10；D IMS-OFR | PAIR-04/07、RESULT-06 |
| L088 | RDR（IP-SM-GW另含IMS OFR）/SM-Delivery-Outcome/IP-SM-GW-SM-Delivery-Outcome/Absent-User-Diagnostic-SM | 10415/3322 | U32/N | 1/1 | 0..1 | 只与对应节点缺席cause配套；0也有效；值链见§3.6 | A S9/S10；D IMS-OFR | PAIR-04/07、RESULT-06 |

### 3.5 SGd/Gdd与SMSMI、位置叶

来源：29.338 §6.2、§6.3.3及表6.3.3.1/2；位置链见主流程§13.3。TFR号码与EPS父/子M位按SGd覆盖，不能照搬旧字典。

| LeafID | 完整路径或已定义路径集合 | Vendor/Code | 类型/预算 | V/M发送值 | 次数 | 条件/检查 | 使用证据 | 候选/验收 |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| L089 | SRR/RDR/ALR/OFR/TFR/SC-Address | 10415/3300 | Octets/O | 1/1 | 1 | 所选短信主过程必需；TBCD无TON/NPI前缀 | A S7/S9–S13 | PAIR-02/04/07 |
| L090 | OFR/TFR/SM-RP-UI | 10415/3301 | Octets/O | 1/1 | 1 | TPDU原文≤200octets，无UTF8/TPDU改写 | A S7/S11 | FIELD-04、PAIR-02 |
| L091 | OFA/TFA/SM-RP-UI | 10415/3301 | Octets/O | 1/1 | 0..1 | 报告TPDU按结果可用性；不自动拼ACK；≤200octets | D | V24-SGD/GDD-MO/MT-A；N24-TFA-TPDU201 |
| L092 | OFA/External-Identifier | 10415/3111 | UTF8/T | 1/0 | 0..1 | T4的SGd-OFA顶层；供计费用途成立时必需；2026-11-07按29.338复用表纠正发送M0，旧M1仅接收兼容；不是原生T4 App或Ro调用 | D | V25-T4-OFR-R/OFA-A；F25-T4-OFA-5555；N25-T4-* |
| L093 | OFR/OFR-Flags | 10415/3328 | U32/N | 1/0 | 0..1 | bit0为S6a/S6d-Indicator，接收不据标志自动改App | D | V25-SGD/GDD-OFR-FLAGS-R；N25-L093-* |
| L094 | TFR/User-Name | 0/1 | UTF8/T | 0/1 | 1 | 普通IMSI，IMS有SMSMI时HSS-ID | A S7；D | D21-IMS-TFR |
| L095 | TFR/MME-Number-for-MT-SMS | 10415/1645 | Octets/O | 1/0 | 0..1 | 目标MME时必需；不能给IP-SM-GW号码套此叶 | A S7 | PAIR-02/06 |
| L096 | TFR/SGSN-Number | 10415/1489 | Octets/O | 1/0 | 0..1 | 目标SGSN时必需；M与S6c节点不同 | A S7 | PAIR-06 |
| L097 | TFR/TFR-Flags | 10415/3302 | U32/N | 1/1 | 0..1 | bit0=MoreMessages；不由队列状态隐式生成 | A S7 | C25-TFR-FLAGS-ZERO/RESERVED-R；N25-L097-* |
| L098 | TFR/SM-Delivery-Timer | 10415/3306 | U32/N | 1/1 | 0..1 | SHOULD携带，单位秒；采用AVP表Unsigned32，正文Integer不另建I32类型 | A S7 | PAIR-02 |
| L099 | TFR/SM-Delivery-Start-Time | 10415/3307 | Time/N | 1/1 | 0..1 | SHOULD携带UTC/NTP；由调用方给定，重传不更新 | A S7 | PAIR-02 |
| L100 | TFR/Maximum-Retransmission-Time | 10415/3330 | Time/N | 1/0 | 0..1 | 同PLMN；若经Router需其Alert能力已知；3330有则3332必需，不建重投递队列 | D | V24-SGD/GDD-MT-R、N24-TFR-NOGMSC-R、V24-ERA-MT-R |
| L101 | TFR/SMS-GMSC-Address | 10415/3332 | Octets/O | 1/0 | 0..1 | 3330存在则本叶必需；GMSC/Router E.164 TBCD，不带TON/NPI | D | V24-SGD/GDD-MT-R、N24-TFR-NOGMSC-R |
| L102 | TFA/Requested-Retransmission-Time | 10415/3331 | Time/N | 1/0 | 0..1 | 仅5550且原TFR有3330、节能条件；完整builder严格早于上限，相等边界单列；Router→GMSC还需3102/1 | D | V24-*-MT-WAIT、N24-TFA/ROUTER-*、C24-TFA-EQUAL |
| L103 | TFA/Absent-User-Diagnostic-SM | 10415/3322 | U32/N | 1/1 | 0..1 | 实验5550时可有；0存在保留 | A S8 | V24-SGD/GDD-MT-WAIT；RESULT-07 |
| L104 | OFA/SM-Delivery-Failure-Cause/SM-Enumerated-Delivery-Failure-Cause | 10415/3304 | Enum/N | 1/1 | 1 | 实验5555要求父组；MO只选3/4/5/6 | A S12/超时OFA | F24-SGD/GDD-MO/MT-5555-C*；N24-OFA/TFA-* |
| L105 | TFA/SM-Delivery-Failure-Cause/SM-Enumerated-Delivery-Failure-Cause | 10415/3304 | Enum/N | 1/1 | 1 | 实验5555要求父组；MT只选0/1/2 | A S8 | F24-SGD/GDD-MO/MT-5555-C*；N24-OFA/TFA-* |
| L106 | OFA/TFA/SM-Delivery-Failure-Cause/SM-Diagnostic-Info | 10415/3305 | Octets/O | 1/1 | 0..1 | 不保证文本，空/二进制保留，不全量日志 | A S8/S12/超时OFA | F24-SGD/GDD-MO/MT-5555-C*；N24-OFA/TFA-* |
| L107 | IMS OFR/TFR/SRR/RDR/ALR/SMSMI-Correlation-ID/HSS-ID | 10415/3325 | UTF8/T | 1/0 | 0..1 | 仅IMS OFR组中必需；SRR/RDR/ALR/TFR的本组禁此叶 | D特殊身份 | IDENTITY-01/02 |
| L108 | IMS OFR/TFR/SRR/RDR/ALR/SMSMI-Correlation-ID/Originating-SIP-URI | 10415/3326 | UTF8/T | 1/0 | 0..1 | IMS OFR/TFR必需；SRR/RDR/ALR禁此叶 | D特殊身份 | IDENTITY-01/02 |
| L109 | IMS OFR/TFR/SRR/RDR/ALR/SMSMI-Correlation-ID/Destination-SIP-URI | 10415/3327 | UTF8/T | 1/0 | 0..1 | 这五种IMS特殊请求均需；原URI不做百分号解码重写 | D特殊身份 | IDENTITY-01/02 |
| L110 | OFR/TFA/EPS-Location-Information/MME-Location-Information/E-UTRAN-Cell-Global-Identity | 10415/1602 | Octets/O | 1/0 | 0..1 | 恰7字节，PLMN/ECI或TAC布局按主流程§13.3；无Ro tag | D；A/B未消费位置 | V24-SGD-MT-A；N24-TFA-ECGI6；N25-L110-DUP |
| L111 | OFR/TFA/EPS-Location-Information/MME-Location-Information/Tracking-Area-Identity | 10415/1603 | Octets/O | 1/0 | 0..1 | 恰5字节，PLMN/ECI或TAC布局按主流程§13.3；无Ro tag | D；A/B未消费位置 | V24-SGD-MT-A；N25-TAI-4/6；N25-L111-DUP |

### 3.6 缺席诊断和非本轮启用的扩展

`SMS-DIAG-PRESENCE`：29.338的3313/3314/3315/3322→[23.040 V18.0.0](https://www.etsi.org/deliver/etsi_TS/123000_123099/123040/18.00.00_60/ts_123040v180000p.pdf)表1/1a，4字节U32承载0..255的原因。表1a已分配0..13：0 MSC无寻呼响应；1 IMSI detached；2漫游限制；3/4非GPRS注销/清除；5 SGSN无寻呼响应；6 GPRS detached；7/8 GPRS注销/清除；9/10 MSC/SGSN未知订户；11 IMS未注册；12 IP-SM-GW无响应；13暂不可用。其余保留，不把14..255命名成新原因。超255违反本值域；保留值可有界接收并报告未解释，正常语义builder只生成已定义值，不映射为成功/默认缺席。节点归属/结果条件另检查，不能把枚举表当所有父路径均可随意取值。零值必须有Presence，禁止继承S9的`!=0`判断。该诊断不是3304的失败原因或3305的透明diagnostic bytes。

以下不因列在字典或消息结构里而升级成实际使用：Sh Requested-Domain/Current-Location/Session-Priority/Requested-Nodes/Serving-Node-Indication/Pre-paging/Local-Time-Zone/UDR-Flags/Call-Reference-Info，wildcard身份，OC/Load；SMSF专用组、非LTE位置/地理/CSG/eNodeB等。命令顶层和其定义仍保留在主流程§2–3，结构接收/未知扩展遵守既有规则；对应功能helper、语义符合性和特性广告均须单独列未支持/Planned。不能静默丢失未知可选字段，或让未实现功能阻塞无关已选字段的解释。若业务配置实际启用这些路径，先将其叶级依赖、结果及独立fixture加入本表再宣称支持，不预建全规范对象树。

## 4. 每个命令的结果分支

`Base/n`表示268，`3GPP/n`表示297内Vendor266=10415与298=n，不是可互换的同一整数。三应用正常答案采用其规定的结果选择规则；2001与实验失败并存、双268、缺结果均不能被“后一个覆盖前一个”或“先检查2001”误判成功。E位基础错误先选错误语法，必需基础268，允许的补充297及其语义边界见主流程§24.2；不套正常结果XOR或正常成功字段。未知Vendor/未知合法数值保留命名空间与原值，报告未知含义，不自动归为成功/重试；基础结果类别按RFC解释仍不授权业务重发。结果语义也不推导身份、外部提交或远端存储已经完成。

### 4.1 成功载荷与通用错误

| 答案 | Base/2001的条件和附带数据 | 非成功的处理 / 来源 |
| --- | --- | --- |
| UDA | 按原UDR所请求的已选数据返回702；Ref24使用SMS注册数据路径，合法的空/未配置表示仍按29.328和XSD | 不能把“无702却2001”的旧模拟器响应当完整返回数据能力；29.328 §6.1.1 |
| PUA | 确认本次更新；Ref24空号码为注销，不要求回User-Data | 不证明库保存注册关系；§6.1.2 |
| SNA | 只确认本次订阅/退订；709实际值或省略无限期；710=1时SHOULD带702，不提升为无例外MUST | 多Ref全成全败为HSS过程，未知订阅的退订也成功；§6.1.3 |
| PNA | 确认本次通知；没有DataRef或Sh-Data回显要求 | AS业务决定认可数据/订阅；本库没有订阅表，不能自动选择5107；§6.1.4 |
| SRA | 必需成功User-Name；投递用途带合规Serving-Node；原请求非投递用途例外，LMSI/附加节点按条件 | 无IMSI/Realm规则分别执行，2001不豁免字段条件；29.338 §5.2.1及主流程§18 |
| RDA | 确认本次报告；仅在适用时提供不同Alert身份 | 不保证后续Alert或短信已投递；§5.2.3 |
| ALA | 确认本次Alert；不要求回用户身份/节点 | 所选ALR过程没有专用应用错误，采用基础错误；旧SGd Alert须保持原App的兼容规则；§5.2.2/6.2.3 |
| OFA | 确认本次MO处理；SM-RP-UI/External-Identifier仅按适用流程提供 | 延迟应答/超时互斥，SMSC超时可由应用明确给5555/cause4，但库不把所有Deadline自动变此结果；§6.2.1 |
| TFA | 本次MT传递成功；若适用可含报告TPDU和位置；Router Alert身份是5550/3331条件分支，不是2001必需字段 | 不自动发RDR/安排重试；§6.2.2 |

基础错误依现有BASE-ERROR/VAL契约：可信请求的3001不支持命令、3007不支持应用、3002无法投递、3004过载、5001未知必需AVP、5004非法值、5005缺AVP、5008不允许的AVP、5009重复过多、5012无法执行、5014非法AVP长度等分别按实际原因生成；Failed-AVP遵循具体错误要求。**不是本地任何失败都能发这些码**：不可信/截断包、超预算、错误答案按原策略丢弃/关闭或本地报告；本地接纳失败不伪造远端答案。3GPP/5001是用户未知，3GPP/5008是TooMuchData，不能与基础码混淆。

### 4.2 Sh实验结果清单

来源：29.329 §6.2、29.328 §6.1.1–4过程及表6.1.4.1.1；5001/5002/5008/5011复用29.229。这里列已采用的直接程序含义，不把表做成“其他合法扩展结果一概坏包”的白名单。

| 3GPP码 | 名称简写 | 适用答案/触发 | 候选与缺口 |
| --- | --- | --- | --- |
| 5001 | USER_UNKNOWN | UDA/PUA/SNA：HSS不知道访问用户；PNA：AS不认识该用户 | F24-SH-UD/PU/SN/PN-5001；各命令完整替代答案 |
| 5002 | IDENTITIES_DONT_MATCH | UDA/PUA/SNA中适用Private Identity与公共访问键不匹配 | 既有V21-UNSUB-F；不凭字符串不同产生该码 |
| 5008 | TOO_MUCH_DATA | PUA透明数据容量；PNA收到过多数据，29.328表6.1.4.1.1明确允许 | F24-SH-PN-5008；PUA透明数据容量正向触发仍按扩展范围补 |
| 5011 | FEATURE_UNSUPPORTED | 所需终端特性不支持的对应Sh交换；按采用特性条款 | M1组有不支持特性才触发；须回支持集。见主流程§21.1；Known不支持的helper仍本地拒绝，不伪造答案 |
| 5100 | USER_DATA_NOT_RECOGNIZED | PNA不认识/不支持所通知数据 | F24-SH-PN-5100 |
| 5101 | OPERATION_NOT_ALLOWED | UDA/PUA/SNA的数据操作/访问键不适用 | F24-SH-UD/PU/SN-KEY-5101；另配已知PSI不适用的请求，不从URI猜身份类别 |
| 5102 | USER_DATA_CANNOT_BE_READ | UDA缺读权限 | F24-SH-UD-5102 |
| 5103 | USER_DATA_CANNOT_BE_MODIFIED | PUA缺写权限 | F24-SH-PU-5103 |
| 5104 | USER_DATA_CANNOT_BE_NOTIFIED | SNA缺通知权限 | F24-SH-SN-5104 |
| 5105 | TRANSPARENT_DATA_OUT_OF_SYNC | PUA RepositoryData序号/创建冲突 | 结果可识别；Ref0更新事务未纳入typed helper，不查询/递增本地版本 |
| 5106 | SUBS_DATA_ABSENT | SNA订阅不存在的RepositoryData | 结果可识别；Ref0完整订阅未启用，不从空XML推导 |
| 5107 | NO_SUBSCRIPTION_TO_DATA | PNA表示AS未订阅此数据 | F24-SH-PN-5107；由应用决定，不查库 |
| 5108 | DSAI_NOT_AVAILABLE | PUA/SNA的目标DSAI未配置 | 结果可识别；Ref19完整模型/业务尚未启用 |
| 4100 | USER_DATA_NOT_AVAILABLE | UDA当前无法提供数据 | F24-SH-UD-4100；不自动重试 |
| 4101 | PRIOR_UPDATE_IN_PROGRESS | PUA并发更新进行中 | F24-SH-PU-4101；不是本库锁错误 |

PNA允许3GPP/5001、5008是本轮补充，不能只实现5100/5107。HSS收到PNA用户未知后删除相应订阅的程序后果属于HSS；库返回结果但不执行该动作。各Sh成功使用Base/2001，不造3GPP/2001。基础/实验结果的头App、命令、Session/HBH/E2E仍先通过关联检查。


### 4.3 S6c、SGd/Gdd实验结果清单

来源：29.338 §5.2.1/3、§6.2.1/2、§7.3，结合信息元素表与结果定义，不能只抄一处列表。

| 3GPP码 | 名称简写 | 适用答案/条件 | 附加字段 / 候选 |
| --- | --- | --- | --- |
| 5001 | USER_UNKNOWN | SRA/RDA；TFA（MME/SGSN不认识用户）；SGd OFA（IWMSC按MSISDN，§7.3.2） | F24-S6C-SR/RD-5001、F24-SGD/GDD-MT-5001、F24-SGD-MO-5001；Gdd OFA同码单列接口边界 |
| 5550 | ABSENT_USER | SRA；TFA | V24-SRA-ZERO、V24-SGD/GDD/ROUTER/ERA-MT-WAIT；特殊IP-SM-GW失败身份另按既有条件 |
| 5551 | USER_BUSY_FOR_MT_SMS | TFA | F24-SGD/GDD-MT-5551 |
| 5552 | FACILITY_NOT_SUPPORTED | SRA、OFA、TFA | F24-S6C-SR-5552、F24-SGD/GDD-MO/MT-5552 |
| 5553 | ILLEGAL_USER | TFA（MME/SGSN鉴权失败） | F24-SGD/GDD-MT-5553 |
| 5554 | ILLEGAL_EQUIPMENT | TFA（MME/SGSN设备检查失败） | F24-SGD/GDD-MT-5554 |
| 5555 | SM_DELIVERY_FAILURE | OFA/TFA | F24-SGD/GDD-MO-5555-C3..6、MT-C0..2；含3303/3304及透明3305；错方向/缺字段负例已补 |
| 5556 | SERVICE_NOT_SUBSCRIBED | SRA（HSS/Router） | F24-S6C-SR-5556 |
| 5557 | SERVICE_BARRED | SRA（HSS/Router）；TFA（§7.3.10明确的MME分支） | F24-S6C-SR-5557、F24-SGD-MT-5557；C24-GDD-SERVICEBARRED为未支持角色控制 |
| 5558 | MWD_LIST_FULL | RDA（HSS） | F24-S6C-RD-5558；不是本地队列满 |

`SMS-RESULT-CLOSURE`：表内未列的结果值仍能被bounded decode/结果容器保存；按实际定义报告未识别/不适用，不能覆盖原值。三应用全部真实业务错误到SMSC/SIP内部错误码的映射归应用。基础5xxx或3GPP永久失败与23.040表1的业务暂时/永久分类不是同一个维度，不能直接套为本库TransportFailover或自动重试策略。


## 5. Sh深层数据的支持边界

### 5.1 实际使用与已选XML叶

XML定义依据29.328 Annex C/D和已固定原ShDataType_Rel18.xsd；字段位置既不是AVP code，也不因本地同名Go结构而成立。Sh User-Data是OctetString，显式helper才解析XML。当前S5构造Ref24注册/注销；S1/S2可传DataReference参数不证明任意Ref已用。S4读取身份后发业务通知，没有读取PNR User-Data的业务值；不能据此省略标准PNR的702或把任意PNR当可达通知。

记`I=/Sh-Data/Sh-IMS-Data`，`E=I/Extension/Extension/Extension`，这是展开后的完整路径别名；均无默认namespace。元素数量为XSD约束，操作规则另施加，例：Ref24更新时SMSRegistrationInfo的号码必须显式给出，空值是注销。

| XML叶ID / 完整路径 | XSD值/次数 | 本库有限表示与来源 | 证据/候选 |
| --- | --- | --- | --- |
| X01 / /Sh-Data/PublicIdentifiers/IMSPublicIdentity | tSIP_URL/tTEL_URL联合类型（基于string），0..N | S5实际生成；保留顺序，文本须XML转义；不表示同时更新Ref10。正常Ref24 helper只构造相应SMSRegistrationInfo，额外身份是原XML内容，不凭它改变访问键 | A S5；X23-IDENT |
| X02 / E/SMSRegistrationInfo/IP-SM-GW-Number | string，每组1 | 非空数字helper采用既定E.164子集；显式空为注销；不把TBCD/MAP地址写入文本 | A S5；X-SH-REG/DEREG-01 |
| X03 / E/SMSRegistrationInfo/Extension/SCAddress | string，0..1，≤20字符 | 已选Pull字段；数字helper与任意schema文本分开，不用于PUR设置SC地址 | D；X-SH-PULL-01 |
| X04 / E/SMSRegistrationInfo/Extension/Extension/IP-SM-GW-SBISupportIndication | unsignedByte枚举0/1，0..1 | 固定XSD准确拼写；仅表示数据，不实现SBI连接或改变Diameter路由 | D；X-SH-SBI-01 |
| X05 / E/UEReachabilityForIP/UEIPReachabilityMME | unsignedByte枚举0，0..1 | 0为可达；缺失、空、0不同，不解释为持续可达承诺 | D Ref25；X-SH-MME-01 |
| X06 / E/UEReachabilityForIP/Extension/UEIPReachabilitySGSN | unsignedByte枚举0，0..1 | 与MME路径不同，不按同名递归匹配 | D；X-SH-SGSN-01 |
| X07 / E/UEReachabilityForIP/Extension/Extension/UEIPReachabilityAMF3GPP | unsignedByte枚举0，0..1 | 正确扩展层，先于non3GPP | D；X-SH-AMF-01 |
| X08 / E/UEReachabilityForIP/Extension/Extension/UEIPReachabilityAMFnon3GPP | unsignedByte枚举0，0..1 | 不因同值0合并两接入域 | D；X-SH-AMF-01 |
| X09 / I/IMSUserState | unsignedByte枚举0..3，0..1 | 仅Ref11已选读取/通知，非SMS reachability枚举 | D；X-SH-CONT/MIX-01 |
| X10 / I/SCSCFName | string/SIP URL，0..1 | Ref12；空移除与缺失分开；位于IMSUserState之前 | D；X-SH-MULTI-01 |

PublicIdentifiers容器及IMSPublicIdentity的schema合法性不能证明PUR Ref24的每个额外数据都有更新权限。**首版不复制旧模板把额外身份当Ref24更新数据**：通用raw入口保留原文，操作解释不称完整通过；有限SMS helper仅生成所选SMS字段。是否需要保留额外PublicIdentifiers的部署行为，另由旧/新对端样本核对，不凭XSD通过建立兼容特例。PublicIdentifiers/MSISDN、Extension下身份集合也不从此扩大为Ref10 helper。

2026-11-06第十八批已实施本表X02–X08的有限构造/解析与独立检查，X01仍按未解释子树保留，X09/10于第十九批追加有限实现（见[batch-19](../spec/history/batch-19.json)）；对应13份冻结XML和深层反例见[batch-18](../spec/history/batch-18.json)。此状态不把完整Sh或全表其他字段改标为Passed。

### 5.2 RepositoryData、iFC、DSAI：具体结构已定位，业务模型不预建

本轮对A/B入口未找到Ref0/13/19实际构造或消费；S14的ServiceIndication、DSAI字段仅结构声明，不代表HSS实现。它们不属于Ro延期，也不被删除：保持可承载原User-Data、识别相应结果/选择键，**暂不提供完整typed数据helper，不阻塞已选24/25及11/12子集**。若后续发现真实调用，再按证据纳入功能范围，而非把整个Sh数据树当首版前置。

| DataRef / 精确结构 | 已核对的边界 | 扩大typed支持前还需什么 |
| --- | --- | --- |
| 0 / /Sh-Data/RepositoryData[]/ServiceIndication、SequenceNumber、ServiceData | ServiceIndication为string且必需；SequenceNumber为xs:int限制0..65535且必需；ServiceData可省，存在时包含一个xs:any/lax子元素，不是任意裸base64文本。无ServiceData有删除语义；XML序号不能套AVP716的U32宽度；多RepositoryData不能按键覆盖 | 透明负载schema/应用合同、选择键与序号规则/创建删除/5105/4101的独立样本。HSS比较并更新序号、并发事务归应用；库不保存或分配序号 |
| 13 / I/IFCs/InitialFilterCriteria[] | Priority≥0及ApplicationServer必需；TriggerPoint可省，存在时ConditionTypeCNF及SPT列表必需。SPT含可选ConditionNegated、Group列表及RequestURI/Method/SIPHeader/SessionCase/SessionDescription五选一；ApplicationServer含ServerName、可选DefaultHandling/ServiceInfo；ProfilePartIndicator及Extension另按XSD | 实际iFC变体、29.228/29.229语义与所需regex/条件模型；不实现SIP消息匹配/业务路由/规则执行。允许空IFCs表示移除不能被自动补空过滤器 |
| 19 / I/Extension/Extension/DSAI[]/DSAI-Tag、DSAI-Value | 两叶均必需；Tag为string，Value为0 ACTIVE或1 INACTIVE；不位于第三层短信扩展，不与AVP711的OctetString强制互转 | 实际DSAI选择键/标签编码及5108样本；触发iFC mask/unmask、Cx推送是HSS责任，不成为库行为 |

XSD能接受上述结构只证明schema层。未知**标准namespace**的子树不能在typed helper中被悄悄删掉后宣称完整解释；允许的外部namespace扩展按原XML保留并报告未解释范围。Schema的lax不能触发在线取schema或外部实体；不预建插件化XML执行器。`XML-03`候选专门验证“schema通过但超出typed支持”的差别。

## 6. 国际化支持分层

依据[RFC6733 §4.3.1/Appendix D](https://www.rfc-editor.org/rfc/rfc6733.html#appendix-D)、[RFC3629 §3–4](https://www.rfc-editor.org/rfc/rfc3629.html#section-3)、既定23.003及[RFC4282 §2.4](https://www.rfc-editor.org/rfc/rfc4282.html#section-2.4)。[RFC7542](https://www.rfc-editor.org/rfc/rfc7542.html)已取代4282，但固定3GPP引用的External-Identifier不因此被悄悄换成另一套NAI规范；7542仅作为差异对照。本节是首版有限能力声明，不宣称全SASLprep/IDNA/IRI已实现。

| 输入层 / RuleID | 首版规则 | 明确不做的转换或声明 |
| --- | --- | --- |
| UTF8String / TEXT-UTF8 | 按RFC3629验证1–4字节Unicode标量编码，拒绝过长编码、孤立续字节、代理项及>U+10FFFF；6733取值范围排除U+0000。有效非ASCII文本可承载，预算按octets | 不把Go string转换当验证；6733旧描述的较大上界不授权5/6字节UTF8；不默认NFC/NFKC、trim、大小写折叠或替换非法字节后再发送 |
| OctetString / TEXT-OCTETS | 702/3301/3305/704/711及Proxy-State保留原字节；字段自己的长度/语义再检查 | 不全局强制UTF8，不把SM-Diagnostic-Info/Service-Indication当可打印字符串；不能把payload未解析归为编码失败 |
| DiameterIdentity / ID-ASCII | Host/Realm线路值为非空ASCII FQDN/realm，检查所选语法；国际域名需调用方提供ASCII编码形式。原wire表示保留，Peer身份比较沿用既有ASCII不区分大小写规则 | 不自动对U-label做IDNA转换；ASCII `xn--`字符串通过字节/标签检查不等于已验证完整IDNA名称、DNS存在或授权。IDNA生成/完整语义验证作为额外能力未支持，报告检查范围 |
| SIP URI / ID-SIP-SYNTAX | 沿用3261/5954的整串语法，ASCII scheme/host/IPv6字面量与合法百分号转义；非ASCII UTF8通过基础层不等于合法SIP URI。%编码保持原样 | 不把IRI直接当SIP URI，不自动URL解码、拼realm、改大小写/参数顺序；百分号表达的octets不擅自按某种字符集归一化 |
| External-Identifier / ID-EXTERNAL-ASCII | 既定23.003 §19.7.2、4282；首版完整解释限ASCII local+domain，保留合法转义及非空两段。非ASCII用户名可过UTF8层，但其SASLprep/禁止字符/Bidi/未分配码点检查不完整，报告语义未支持 | 不把`split("@")`当NAI解析，不把72/253当协议最大长度；不静默采用PRECIS/7542替代4282，不加通用实时facts绕过 |
| Sh Private Identity | 顶层User-Name按所用23.003身份语法与上述ASCII子集解释，IMSI/NAI/IMS私有身份按用途分别检查 | 不套SRA的IMSI约束，不要求库验证IMPU与私有身份的数据库对应关系 |
| Sh XML / XML-TEXT | 原702按bytes存储；显式有限helper生成UTF8 XML并正确转义文本，按XML合法字符和XSD值检查；非ASCII扩展文本可保留 | UTF8有效不代表XML允许U+0000等字符；不拼接未转义模板、不C14N、不改XML前缀/空白来消除重复冲突。完整多编码XML兼容不由UTF8 helper承诺 |

已知非ASCII DiameterIdentity违反线路表示规则，与合法UTF8但未支持完整NAI语义是不同报告；不能统一判5004，也不能统一当raw成功。未支持国际化语义的相关typed builder在提交前失败，不将其结果强行解释为“合法可用身份”；通用有界结构保留用于调用方处理，并准确标注检查范围。现有身份准入/binding与重复请求Canonical契约不因本轮改变。

`TEXT-01`必须覆盖：中文UTF8的字节长度；U+0000与过长/代理项编码；组合字符与预组合字符原文不合并；U-label与ASCII编码身份；转义NAI；`sip:a%2Fb@example.net`不改为`a/b`；XML文本`&`转义及外部namespace Unicode。不能用一例ASCII正向证明国际化全支持。

## 7. 参考库差异、迁移纠错和关闭门槛

| 已读实现/证据 | 有用模式与局限 | 本库处理 |
| --- | --- | --- |
| [OTP diameter_dict](https://www.erlang.org/doc/apps/diameter/diameter_dict.html)，本轮所见diameter2.7.2 | 字典分messages/grouped/types/enum；文档说明UTF8校验，但DiameterIdentity的数据表示只说明非空octets | 参考定义分层，不把有类型的record视为全字段语义校验；不照搬字典即可产生的应用成功判断 |
| [freeDiameter固定提交的类型函数](https://github.com/freeDiameter/freeDiameter/blob/9d4bc47584bfeced9de708fe00f5629ac4689db5/libfdproto/dictionary_functions.c) | UTF8String_dump处理显示末端字节；这是dump函数，不是全UTF8/IDNA验证证据 | 不将展示能输出当格式合法；保留原值与报告，日志另外限额 |
| [本地go-diameter UTF8 decoder](../../go-diameter/diam/datatype/utf8string.go)，HEAD f26e3528006fcf0d33cb663994a0ec0ccbf7ef55 | 直接转换OctetString并返回nil，没有验证UTF8；旧S6c字典3112类型/SGd M位差异前轮已确认 | 新库明确TEXT-UTF8、父路径覆盖和应用层校验；不把全部字段降为raw来逃避痛点 |
| S1/S2/S3 | DataRef24使用U32构造而标准为Enum；Expiry/OneTime发送M=1；SNR可同时填公共身份/MSISDN；目标取邻接元数据 | 位宽相同不等于类型/条件相同；修正发送flags、访问键和DRA终端binding，不兼容化明显错误 |
| S6/S7 | SRA成功注释套用了SRR身份描述；Additional解析函数的SGSN Realm取Name；TFR IP-SM-GW分支填MME号码 | 分别保留真实字段；SRA成功不回显HSS-ID；IP-SM-GW不冒充MME；业务路由选择与缓存不进入库 |
| S9/S10 | S9诊断只在非零时填；S10按输入cause==1而非映射后的diaCause决定诊断 | 原因/Presence由明确协议字段决定，0不省略；不把SMSC→Diameter业务映射移入公开库 |
| S4/S14、MME模拟器 | PNR不消费702；模拟UDA可能2001但不含所需数据；Origin复制Destination；旧成功构造可能重复结果AVP | 模拟器通过不替代规范；正常/错误条件与单值数量分别验收，身份来自已授权endpoint |

新增FIELD-14（111行叶清单）、RESULT-06（逐命令结果）、XML-03（深层边界）、TEXT-01（国际化分层）。它们细化既有24个SPEC ID，没有新增公开类型或持久状态。所有实现/独立栈验收仍Planned；表中“待补”必须逐项补独立bytes/结果变体，不能以本页覆盖行数作通过证据。

优先级：P1先补本页标记的实际A/B与已选D分支的独立正反fixture、零诊断/结果/条件组合，以及命令字段清单到实现的闭合；真正无IMSI成功SRA和Realm标准矛盾保留分支证据门槛。Repository/iFC/DSAI完整helper及全IDNA/SASLprep为尚未纳入的扩展能力，需实际需求触发，不作为当前已选短信/Sh子集的无条件P1。通用Peer/FO/VAL/BUD/SCTP、DRA身份/保护和部署证据仍P1；Ro/六模板/SCAP1纳入首版实施。


## 8. 第二十四轮覆盖状态

[完整候选](diameterkit-leaf-vectors.md)给出136帧（49应用失败替代答案）；FIELD-14/RESULT-06进一步由PAIR-08、NODE-02、TIME-01、RESULT-07细化。本轮优先字段现有可独立提取的正反bytes，执行与独立栈证据仍Planned；表中ID缩写如SGD/GDD、通配符仅是索引，实际唯一ID及每帧摘要在候选§4。

纠错包括：普通HSS双节点须结合SRR bit0和跨组SGSN/MSC/MME限制；PUA703/715属于条件化失败定位；TFA Router身份与5550/3331关联；Time相等边界保留未裁决，跨2036不比较裸U32。证据不足的外部事实不新建查询接口，不成为库业务状态。

剩余不笼统写“所有字段待补”：已补五种Additional形式、零MSC诊断、LMSI/Alert/MWD零值、Sh实际到期/取数据、MO/MT报告和MME位置、时间依赖及全部MO/MT失败cause。仍待逐项补MME/SGSN诊断值变体、其他单叶数量/预算/宽度反例、SM-RP-SMEA、其余位图与身份变体、Sh5011完整特性触发，以及独立实现验证。T4 OFA External-Identifier整条pair仍按已选D保留P1；多Repository/多DataRef失败定位与未启用深层模型按实际功能需求单列，不从本轮补样本扩大为全Rel-18支持。Ro纳入首版实施。

## 9. 第二十五轮更新与剩余范围

当前规则以[主流程§21](diameterkit-mainflow-audit.md)为准；§8是上一轮交付状态。新增251条完整帧及[§7.1逐叶反例索引](diameterkit-leaf-vectors.md)：L001–L111每叶均有独立bytes，不再笼统写单叶反例待补。数值宽度、非法文本、单值重复、可重复项本地预算分开；父路径集合尚非全组合矩阵，应用语义未支持的静态样本明确标注。X01–X10沿用既有schema正反记录，未新增深层helper。

Sh5011的完整特性失败/成功/缺特性与基础5001分支、T4 OFR/OFA顶层External-Identifier成功/失败pair、MME/SGSN诊断、SBI0/1、OFR/TFR位图及TAI宽度已补。剩余是独立实现/对端复核、实际使用命令×父路径展开、无IMSI/Realm/时间相等证据，以及未启用模型的按需工作。SMEA采用数字完整地址子集，其他TON/NPI或字母编码不冒充已支持。

旧V21-SUBMULTI、D22混合请求628 M0与V24双节点SRR漏3309已登记勘误，见主流程§21.4；叶列引用旧样本只证明其列示字段，整pair当前资格必须看勘误。所有本库执行验收仍Planned，Ro纳入首版实施，不新增持久化或业务查询。

## 10. 第二十六轮命令与父路径展开

[439行矩阵](diameterkit-command-path-matrix.md)把111叶展开到18种正常命令，分列精确Vendor/Code、父路径、在位候选和定向反例；第二十六轮快照为253行有在位候选、142行有本表定向反例（当前增量见§11）。其余297行没有该口径反例，其中254为公共路径、43为应用路径；不能用另一个命令/父组的反例补计。矩阵中13个禁止位置与26个未启用语义边界分别标注，禁止位置均有应用反例。

新增S15–S19补命令另一侧的实际入口。S15未声明双节点能力，不能因现有普通SRR未填SMEA就认为违反新增条件；S16 UDA仅解析/日志，不证明业务消费User-Data。S17/S18主要消费结果，SNA到期/User-Data/特性仍属于设计补齐；S19的OFA超时5555/cause4来自业务决策，库不自动复制该映射。S1–S19的工作树摘要是只读证据，不代表部署启用。

L081–L086在IMS OFR不适用；只有L087/88为所选IP-SM-GW outcome。L107/108在不同SMSMI请求中的禁位必须按矩阵处理，不能照字段集合机械复制。111叶定义不增删，X01–X10不变；完整支持状态仍Planned，Ro现纳入首版实施。

## 11. 第二十七轮：27条应用路径缺例和组合边界

[矩阵](diameterkit-command-path-matrix.md)的27个已选应用路径缺例已补；剩余16个应用缺例全部属于按需结构边界。111叶定义和X01–X10不增删，A/B/D/C来源不变。新增Sh MSISDN/External/Private在位控制用于Ref24的UDR/PUR与Ref25的PNR，不声称业务handler已消费这些全部字段。Private通知的订阅用途由场景/应用提供，库无订阅表。

表7.6.1 Note9补充L023/L024组合限制：External-Identifier访问键不与Private Identity叠加；这不是二者类型本身错误。新3条组合反例及27条定向反例、10条基线见[逐叶§10](diameterkit-leaf-vectors.md)。L046数字SMEA的三段格式、15位与12octets的不同上限，以及单向bit0条件见[主流程§23.6](diameterkit-mainflow-audit.md)；其他TON/NPI未启用不等于未知AVP。L044成功身份、L079 Realm和L102时间仍按对应未决边界报告。独立验证仍Planned。

## 12. 第二十八轮：答案公共叶与E语法边界

[九答案族](diameterkit-answer-envelope.md)覆盖正常结果选择、Session/Origin/Auth-Session-State、Sh VSAI及E语法；L001–L004、L008/L010/L012的正常叶反例按明确目标登记，矩阵当轮222行有定向反例、217行缺例，其中201公共行与16按需应用边界。E1并未加入439行分母，不能因为E错误无277/260就反向放宽正常答案。

Failed-AVP依据Verified4615只一个父组、可含多个错误证据；其内部被引述的非法字段不按应用规则再次拒绝。外层仍有界、非空且计资源；容器外281的UTF8仍检。Session在九族E错误也须回原请求、Origin-Realm采用4887，本地Origin不由请求Destination推导。294是诊断来源而非新的来源授权。具体边界见[主流程§24](diameterkit-mainflow-audit.md)。111叶与X01–X10定义数量不变，Ro纳入首版实施。

## 13. 第二十九轮：公共父实例与诊断字段解释

L014/L017在18正常命令各有本父缺失/重复输入，L013/L016/L018/L011/L002另补UDR代表边界；精确计数以[矩阵§8](diameterkit-command-path-matrix.md)为准。公共未知项不是新增LeafID，Failed-AVP中的引述不是业务叶。已识别诊断树计深度/节点，未知opaque只计可见外层和全部字节；坏引述长度停止分支，不能用它分配内存。L044无IMSI、L079 Realm、L102时间相等的P1和新CR0069证据线索见[主流程§25.4](diameterkit-mainflow-audit.md)，不改变既定发送/接收子集。

## 14. 第三十轮：状态、路由、结果父组与错误文本

L005/L015各18命令、L009/L019/L020各9答案族均有定向输入，L010增加直接宽度反例；父组分裂为显式复合场景，不给每叶重复增分。ALA未知Vendor结果仅接收结构控制，不新增标准业务错误。L005零值语义不抹去Presence，L015保序不改变R17剥离；L019区别无效UTF8、NUL和控制码SHOULD。具体依据、CR0069重试结果及无IMSI/时间跨接口证据见[主流程§26](diameterkit-mainflow-audit.md)。

## 15. 第三十一轮：公共缺例与过程约束

L001–L020每个适用命令/路径现有定向反例，新增104条原缺例行，总423/439；不是全错误维度。L013/L016/L018补18命令直接父实例输入；L001/L002/L003/L004/L006/L007补九请求缺失/次数/目的绑定；L011/L012补八Sh命令。L006必须区分CCF可选与过程必需：本HSS发ALR缺Host仍不满足29.338 §5.3.1.2。16条按需边界精确列表与剩余规范P1见[主流程§27.6](diameterkit-mainflow-audit.md)，无新LeafID或业务helper。

2026-11-07第二十五批已实现L092有限ASCII/原请求身份证据与M0发送，旧3111 M1候选保留接收控制；详见[batch-25](../spec/history/batch-25.json)。此记录不升级本清单其余Planned项或T4 SCTP部署证据。
