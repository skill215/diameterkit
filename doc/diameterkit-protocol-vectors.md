# Diameter 候选协议样本

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

日期：2026-10-31（第二十三轮补逐叶/结果/文本边界），第十四轮建立，第十五轮补§6–7，第十六轮补§8，第十七轮补§9–10，第十八轮补§11–13，第二十轮补§15–17，第二十一轮补§18–20，第二十二轮补§21，第二十三轮补§22，第二十四轮补§23，第二十五轮补§24勘误。状态：人工按规则整理的候选；XML离线schema与字节算术已核对，协议运行验证Planned。本文为[主流程审计 §9–21](diameterkit-mainflow-audit.md)和[Library Specification](diameterkit-specification.md)提供可复核输入，不是库实现、抓包集或已通过的互操作报告。

> 2026-11-07第二十六批实施说明：D21的T4 RDR/ALR六份派生配方已原样物化并核对原摘要。D21-T4-RDR-R保留IP-SM-GW Outcome3320，是当前切片未支持控制；新增B26-T4-RDR-MME明确改变Origin-Host及Outcome子组为3317，另有5001和空/双身份控制。不改旧bytes或把旧候选声明为语义通过；详见[batch-26](../spec/history/batch-26.json)。

> 2026-11-07第二十七批实施：D21-T4-SRR-R/A/F按原配方物化并核对原长度/SHA，另有10份独立身份/节点/GPRS/E投影；已知IMSI路径已实现，无IMSI成功限制未撤销。详见[batch-27](../spec/history/batch-27.json)。

> 实施注（2026-11-07）：第三十八批已在T4DeliveryReport证据下启用D21-T4-RDR-R的IP结果，保留原字节/摘要；新GMSC投影与当前用途验收另记，不回写第26批的历史支持结论。见[batch-38](../spec/history/batch-38.json)。历史轮次当时状态保留。

## 1. 证据等级与使用方法

以下hex按固定RFC/3GPP条款排列；第二十轮用临时算术脚本排出候选字节，没有使用拟实现的diameterkit encoder。这里只检查长度、类型宽度及所列规则的预期；“不依赖新encoder”不等于“已独立验证”。晋升为验收fixture还需另一解析实现/专家逐字段复核、实现端正反测试，以及适用部署的对端证据。保留原候选、来源、复核人/工具版本与差异，不能把自编码自解码成功当唯一证据。

hex以网络字节序表示；空白仅用于阅读。AVP Length不含padding，Message Length含全部AVP及padding。`avp-hex`是一段AVP，不是完整消息；`message-hex`才含20字节Diameter头。除明确反例外，未列出的包络、父路径、原请求、角色和绑定都必须另外提供。片段通过不表示整个命令的Wire/Exchange/Application校验通过。

沿用Profile错误契约：非法宽度/值/缺必需字段在本地构造为FieldError；可信请求才能映射线路错误；应答不能反向生成错误答案。原始报文可关联而Wire/Exchange无效时ErrInvalidAnswer，不自动重传；Application未支持、冲突或缺依据不能变为类型化业务成功。

## 2. 身份、长度与数值片段

### V-HSS-01：保留前导零

来源：29.338 §6.3.3.14、23.003 §5.3；SMSMI-HSS、FIELD-03。HSS-ID为10415/3325，UTF8String `001011`，AVP Length18、padding2、总20。只证明数字结构合规，不确定MNC长度、不证明运营方分配或与IMSI不冲突。

```avp-hex
00000cfd 80 000012 000028af 303031303131 0000
```

反例：保留长度把payload一位改为`61`（a），格式非法；把M从0改为1得到`c0`，对该已理解SMS字段接收时按APP-MBIT忽略该位差异，不能误报未知AVP，Canonical相同；标准发送仍为`80`。

### V-LMSI-01：四字节值

来源：29.173 §6.4.2、23.003 §2.5；SMS-LMSI、FIELD-03。10415/2400，payload `01 02 03 04`，总16。

```avp-hex
00000960 c0 000010 000028af 01020304
```

反例：Length改15并将最后payload字节移为padding后，payload只有3字节，字段长度非法；不能把padding算成第四字节。全零四字节仍保留原值，其分配语义不得凭空升级为全局线路格式错误。

### V-PORT-01：U32线路与16位端口范围

来源：29.368 §6.4.11、23.040 §9.2.3.24.4；RO-TRIGGER-PORT、FIELD-04。10415/3010，值65535，线路payload必须4字节，总16。

```avp-hex
00000bc2 c0 000010 000028af 0000ffff
```

反例：改为`00010000`（65536）仍满足U32宽度，但超端口语义范围；改Length14并仅保留2字节`ffff`加2字节padding则违反U32宽度。两个父路径不能互相搬移字段来消除错误。

### V-SMSTATUS-01：显式零与缺值

来源：32.299 §7.2.209；RO-SMSTATUS、FIELD-04。10415/2014，Length13、1字节`00`和3字节padding，总16。

```avp-hex
000007de c0 00000d 000028af 00 000000
```

反例：删除整个AVP是Absent，不等于值0；Length16、4字节零payload不是合法SM-Status宽度。业务报告条件仍按当前请求检查。

### V-GSU-01：零额度不是缺额度

来源：RFC4006 §8.17/§8.26；RO-UNITS-COUNT、RESULT-03。GSU431包含CC-Service-Specific-Units417=0，子16、父24字节。

```avp-hex
000001af 40 000018
000001a1 40 000010 0000000000000000
```

反例：整个GSU缺失、空GSU（Length8）、本例显式零，各保留独立Presence。何者在原请求动作下允许由Application规则判断，不因空组符合通用ABNF就承诺授予成功。

### V-MONEY-01：有符号指数与币种

来源：RFC4006 §8.8/§8.9/§8.22；RO-MONEY-VALUE、RO-MONEY-CURRENCY、FIELD-06。CC-Money413包含Unit-Value445（Value-Digits447=1234、Exponent429=-2）与Currency-Code425=840；值为12.34，父56、Unit36字节。此例为通用DCCA片段，不宣称该字段可用于所有标准SMS路径。

```avp-hex
0000019d 40 000038
000001bd 40 000024
000001bf 40 000010 00000000000004d2
000001ad 40 00000c fffffffe
000001a9 40 00000c 00000348
```

反例：把Exponent改`7fffffff`仍是合法I32编码；解码保留二元组，不求巨大十次幂。显式转金额类型超范围时返回本地转换错误，不能溢出、耗尽内存或四舍五入而不报告。

### V-MONEY-02：Currency可选性由父路径决定

同源。只保留上例Unit-Value，CC-Money Length44，缺Currency；通用DCCA语法允许但未满足SHOULD出现的建议，应保留缺失并报告建议偏离，不默认币种。

```avp-hex
0000019d 40 00002c
000001bd 40 000024
000001bf 40 000010 00000000000004d2
000001ad 40 00000c fffffffe
```

反例：只把父Code改为`000001a7`（Cost-Information423），相同字节布局仍可安全解析，但其语法必需Currency，故缺字段。Remaining-Balance10415/2021也必需Currency，但必须另建12字节vendor头，不能只改Code套此样本。

### V-UNIT-01：缺指数与显式0

来源：RFC4006 §8.8/§8.9；RO-MONEY-VALUE、FIELD-06。Unit-Value24字节，digits=1234、Exponent absent，数值解释指数0。

```avp-hex
000001bd 40 000018
000001bf 40 000010 00000000000004d2
```

变体：追加`000001ad 40 00000c 00000000`并将父Length改36（`000024`），数值仍1234但Presence不同；当前CanonicalV1没有忽略该字段的规则，不能自动将两种内容合并。

## 3. 最小CCA与结果变体

### V-CCA-01：正常E=0包络

来源：RFC4006 §3.2及32.299 CCA语法；SPEC-CODEC-01、SPEC-RESULT-01。全长116（`0x74`），P=1、R/E/T=0，Command272、App4，HBH=`01020304`、E2E=`05060708`。依次为Session-Id `s;1`、Result-Code2001、Origin-Host `ocs.example`、Origin-Realm `example`、Auth-Application-Id4、CC-Request-Type4(EVENT)、CC-Request-Number0。

```message-hex
01 000074 40 000110 00000004 01020304 05060708
00000107 40 00000b 733b31 00
0000010c 40 00000c 000007d1
00000108 40 000013 6f63732e6578616d706c65 00
00000128 40 00000f 6578616d706c65 00
00000102 40 00000c 00000004
000001a0 40 00000c 00000004
0000019f 40 00000c 00000000
```

只声明这是基础语法候选；并非完整余额检查/直接扣费/退款用例。Exchange需匹配原请求、连接代及授权OCS；Application需原Requested-Action、Service-Context-Id等证据，本页未提供完整CCR，不能因此宣称“扣费成功”。测试身份只作语法输入，不代表可用真实域名/授权。

### V-EXP-01：实验5011片段

来源：32.299 §6.5.10/§7.1.8A；SPEC-RO-5011裁决记录、RESULT-04。Experimental-Result297包含Vendor-Id266=10415与Experimental-Result-Code298=5011，总32。

```avp-hex
00000129 40 000020
0000010a 40 00000c 000028af
0000012a 40 00000c 00001393
```

| 变体ID / 对V-CCA-01的变更 | 完整帧长度 | 预期 |
| --- | --- | --- |
| V-CCA-02：尾部加V-EXP-01，头长度改148 | `000094` | 顶层2001与实验失败冲突；即使其余Wire/Exchange通过，也不能解释为成功；特性相关条件需另行检查 |
| V-CCA-03：从02删除顶层Result-Code的12字节 | `000088`（136） | E=0正常CCA缺必需Result-Code，可关联时ErrInvalidAnswer；不以Experimental代替，不回Answer、不重传 |
| V-CCA-04：将02基础结果payload改`0000138b`（5003） | `000094`（148） | 保留基础/实验失败；此例不是标准5011发送配对裁决，也不授权自动生成该配对 |
| V-CCA-05：将02的实验Vendor改一个未支持值 | `000094`（148） | 保留未知命名空间，不从基础2001推导类型化成功；不能将Vendor未知解释成3GPP5011 |

### V-MSCC-01：每服务失败与顶层失败优先

来源：RFC4006 §5.1.2/§8.16；RO-MSCC-RESULT、RESULT-05。MSCC456、Service-Identifier439=1、Result-Code268=4012，总32。给定已协商多服务能力和匹配原请求服务，不能将此组解释为成功配额。

```avp-hex
000001c8 40 000020
000001b7 40 00000c 00000001
0000010c 40 00000c 00000fac
```

V-CCA-06：在01尾部追加本组并改全长148，顶层成功不抹掉服务失败。V-CCA-07：将06顶层结果改5003、组内结果改2001，顶层失败优先。两个变体都需给出原请求/能力及服务上下文；缺依据只得到未完成解释。重复此片段不能代表合法费率切换：该场景必须另补Tariff限定字段的独立样本，不能为了测试重复键而捏造有效双额度。

## 4. 文本与上下文场景

本节不提供完整AVP hex，专门核对应用语法/输入边界；每行执行状态仍Planned。来源及RuleID详见主流程§10，禁止把单字段预期扩大到整个命令。

| ID / 关联 | 输入与变体 | 预期 |
| --- | --- | --- |
| V-URI-01；ID-SIP-SYNTAX | `sip:alice@example.net`；`sip:alice%4G@example.net`；`sip:alice@` | 第一项基本语法候选；后两项分别为非法转义、缺host；不联网查域 |
| V-URI-02；ID-SIP-SYNTAX、5954 | `sip:alice@[2001:db8::7]`；`sip:alice@[2001:db8:::7]` | 合法IPv6文字地址与非法三冒号；裸IPv6不可绕方括号规则 |
| V-URI-03；ID-SH-CANON | `sip:alice@example.net:5060;transport=tcp` | 泛SIP语法可以合法，但不符合本节Sh公共身份canonical形式；不能自动删参数后接受，不能反向推断SMSMI也必拒绝 |
| V-URI-04；CANON-01 | 同键合法请求的URI host由`[2001:db8::7]`改为`[2001:db8:0:0:0:0:0:7]` | SIP显式地址比较可等价，但CanonicalV1保留内容差异而报冲突；本库没有因此执行第二次业务操作 |
| V-NAI-01；ID-EXTERNAL | `device42@example.net`；`@example.net`；`device42`；`a..b@example.net` | 第一项ASCII结构候选；中间两项虽能是泛NAI却缺外部标识必要组成；末项用户名点语法非法 |
| V-NAI-02；ID-EXTERNAL | `a\@b@example.net`，其中反斜杠是一个实际`0x5c`字节 | 按所选4282，转义@属于用户名；不能按两个@判错。不能把这条规则声称为7542语法；国际化完整验证不由本例覆盖 |
| V-NAI-03；FIELD-05 | 一串符合用户名/域语法的254字节ASCII外部标识，配置预算允许 | 不得仅因超过253报协议格式错；若超过实际本地额度则ResourceLimit。还须具备所有普通字段/帧边界检查 |
| V-CONTEXT-01；PROVISION-01 | CCR的Service-Context-Id为`18.32274@3gpp.org`，省略OM/OC，本地helper默认选择启用 | 接收不能用本地选择报远端Missing-AVP；最短形式的OM/OC可选，M/C及已出现字段仍检查 |
| V-CONTEXT-02；PROVISION-01 | 改为`001.001.18.32274@3gpp.org`，未提供该运营方的具名规则 | 应用检查缺依据，不猜OM/OC配置、不默认为完整通过；原请求证据不明时也不要求CCA伪造Service-Context-Id |
| V-UE-01；ID-UE-ENCODING | Type0、仅IMEI的15位数字；PS补F到16文本，IMS保留15十进制文本 | 规则依业务使用位置，不能互换；原样8字节TBCD不等价于这些UTF-8文本。SMS路径已在第十五轮细分，仅IMEI映射仍未裁决，不把任一形式标为完整短信成功样本；真实IMEISV16子集另见§8 |

## 5. 尚缺什么

VECTOR-01只建立来源/输入/预期的可复核候选。第十五轮§6补通用DCCA完整CCR/CCA候选，第十七轮§9已补标准SMS初始及S6c/SGd部分命令pair；仍缺其余标准SMS流程、Sh/S6c/SGd全命令、所有Grouped递归叶、真实六模板与错误答案、NAI国际化、完整URI/设备编码裁决、Tariff/池限定以及独立实现结果。5011标准发送配对保持未决。以上缺口按能力阻塞支持声明；不为填满表格制造线上证据，也不扩大库的业务职责。

## 6. 第十五轮：完整DCCA请求—应答候选

### V-PAIR-CCR-01：IEC direct-debit请求

采用RFC4006 §3.1/§6.3的通用DCCA基线；这是测试用服务文档`debit@example.net`，约定Service-Identifier=1、计量为一次事件。不是3GPP标准SMS服务标识，也不是实际部署抓包。仅用它验证已有DCCA能力和完整包络关联；标准SMS的不同父路径/category另行验收。

测试元数据：本地CTF身份`ctf.example`、realm `example.net`；授权终端OCS为`ocs.example`/`example.net`；邻接可以是已经授权的DRA，但成功答案Origin仍应是OCS。原请求P=1、R=1、E/T=0；Command272/App4；HBH=`11121314`、E2E=`21222324`；Session-Id=`ctf.example;1;1`仅在本隔离样本中使用，不作生产发号算法。type=EVENT(4)、number=0、Requested-Action=DIRECT_DEBITING(0)。Subscription-Id Type=END_USER_E164(0)、Data=`12025550123`；时间保留原始Time=`e0000000`，如需转换显式使用era0，非现场发生时间。RSU请求1个service-specific unit；Service-Identifier位于顶层，因此不声明多服务能力。

总长288（`0x120`）。依次为Session、OriginHost、OriginRealm、DestinationRealm、AuthApp、ServiceContext、CCType、CCNumber、DestinationHost、EventTimestamp、SubscriptionId、ServiceIdentifier、RSU、RequestedAction。所有补零为padding而非字段值。

```message-hex
01 000120 c0 000110 00000004 11121314 21222324
00000107 40 000017 6366742e6578616d706c653b313b31 00
00000108 40 000013 6366742e6578616d706c65 00
00000128 40 000013 6578616d706c652e6e6574 00
0000011b 40 000013 6578616d706c652e6e6574 00
00000102 40 00000c 00000004
000001cd 40 000019 6465626974406578616d706c652e6e6574 000000
000001a0 40 00000c 00000004
0000019f 40 00000c 00000000
00000125 40 000013 6f63732e6578616d706c65 00
00000037 40 00000c e0000000
000001bb 40 000028
000001c2 40 00000c 00000000
000001bc 40 000013 3132303235353530313233 00
000001b7 40 00000c 00000001
000001b5 40 000018
000001a1 40 000010 0000000000000001
000001b4 40 00000c 00000000
```

### V-PAIR-CCA-01：匹配的成功应答

给定上一请求及上述binding，答案保留HBH/E2E、Session、type/number，P=1、R/E/T=0，Result-Code=2001，Origin为授权OCS，顶层GSU给1个service-specific unit。总长156（`0x9c`）。单服务关联来自原请求，CCA不重复填Service-Context-Id或Service-Identifier。

```message-hex
01 00009c 40 000110 00000004 11121314 21222324
00000107 40 000017 6366742e6578616d706c653b313b31 00
0000010c 40 00000c 000007d1
00000108 40 000013 6f63732e6578616d706c65 00
00000128 40 000013 6578616d706c652e6e6574 00
00000102 40 00000c 00000004
000001a0 40 00000c 00000004
0000019f 40 00000c 00000000
000001af 40 000018
000001a1 40 000010 0000000000000001
```

预期仅为“已收到匹配的协议成功及所声明单元”，不是短信已投递、外部扣费账本已提交或退款恢复授权。对端余额充足是本测试预设结果，不由库查询/证明。没有业务持久化也能完成编解码、关联和纯解释。

| 变体 / PAIR-01 | 预期与所属层 |
| --- | --- |
| 原请求未知/已过期，无相同活动exchange | 不能仅因号码像合法响应就交付成功；按未匹配/迟到答案规则处理，不产生第二次Completion |
| CCA的CC-Request-Number改1，HBH/E2E仍可关联 | Exchange违规，可关联时ErrInvalidAnswer，不自动重传、不对答案再答复 |
| CCA的HBH改为另一值 | 不命中该pending；丢弃/诊断受限，原请求按自身期限继续，不能猜给最近请求 |
| 以DRA名替换成功CCA Origin-Host，终端策略仅授权ocs.example | 来源策略不通过；邻接CER或连接成功不能授权DRA冒充OCS成功 |
| 成功CCA中GSU缺失；或GSU显式0 | direct-debit程序所需单元缺失与0分开报告；零不是“自动拒绝解码”，也不能将它解释成足够提供本例1次服务；不自动再扣费 |
| 将CCR Service-Context-Id改成最短SMS标识，保留顶层RSU/Service-Identifier；相应长度全部修正 | 切换到固定SMS规则后，not-used路径不能由OM/OC最短形式例外放行；不能仅换字符串就成为标准SMS正例 |
| 将GSU挪入MSCC，未提供多服务能力/匹配服务依据 | 不允许解释器只取第一组而沿用单服务结论；按所选Profile/原请求报告相应条件失败 |

上述变体是完整pair上的最小修改说明；除明确给出的两帧外不声称已有独立二进制文件或已运行这些分支。

## 7. 第十五轮：路径与类型反例

### V-NODE-01：S6c的MME号码

来源29.338表5.3.3.1/2、29.272 §7.3.159；S6C-NODE-NUMBER、FIELD-07。10415/1645，S6c发送M=1覆盖来源定义的M=0。国际数字`12345`编码为`21 43 f5`，无TON/NPI；Length15、padding1，总16。

```avp-hex
0000066d c0 00000f 000028af 2143f5 00
```

反例：`21 f3 45`中间F非法；`21 43 a5`不是E.164数字；在前面追加`91`并更新长度不能被解码器自动当TON/NPI剥除，也不能仅因首字节91判为非法号码；其余数字/长度仍须检查。M改0的已理解S6c字段按APP-MBIT接收，标准发送仍按S6c表。只有号码片段不代表整个Serving-Node组合合法，MME组合还需Name/Realm。

### V-ADDR-01：非IP Address恰为16字节payload

来源RFC6733 §4.3.1、32.299 §7.2.128；RO-ADDRESS-FAMILY、FIELD-08。Originator-SCCP-Address10415/2008，family8加UTF8数字`12025550123456`，AVP payload恰16字节、Length28。用于验证类型保存，不证明该号码真实分配。

```avp-hex
000007d8 c0 00001c 000028af 0008 3132303235353530313233343536
```

重新编码必须仍是family8，不因payload长度16当IPv6；不能变成family2加原family8字节。反例：只把family改2而保持14字节地址值，则不满足IPv6所需16字节。该样本只覆盖数字Global Title，不声称覆盖SCCP Point Code/Global Title复合编码。

### V-REPORT-01：合法枚举放错父路径

来源32.299 §7.2.175；RO-REPORT-PLACEMENT、FIELD-06。Reporting-Reason10415/872、THRESHOLD(0)，总16。

```avp-hex
00000368 c0 000010 000028af 00000000
```

同一片段在CCR/MSCC/USU中是候选正确位置（仍需usage上下文）；放到CCR顶层、CCA或直接MSCC中均不能按“枚举0已知”通过相应使用规则。改FINAL(2)则应直接位于MSCC，不能继续放USU。不能由此证明业务真的触发了阈值/终止。

### V-TIME-MECH-01：必需子项与共存优先级

来源32.299 §7.2.228/230/29、§6.5.7；RO-TIME-MECHANISM、FIELD-06。Time-Quota-Mechanism10415/1270含Type1271=CONTINUOUS_TIME_PERIOD(1)、Base-Time-Interval1265=60，总44。

```avp-hex
000004f6 c0 00002c 000028af
000004f7 c0 000010 000028af 00000001
000004f1 c0 000010 000028af 0000003c
```

在允许该机制且有相应时间quota的CCA/MSCC中，与Quota-Consumption-Time共存并非互斥错误，解释时机制优先。删除Base-Time-Interval并把父Length改28仍缺必需子项；只保留外层长度不改则首先是framing错误。库返回机制描述，不启动60秒业务timer。

### V-APP-01：设备父路径、操作与not-used

文本场景，APP-01/PROVISION-01：同一Type0设备组放在普通SMS在线CCR顶层与Service-Information/PS-Information内，必须先区分位置；在线表允许前者、嵌套项标“-”。最短`18.32274@3gpp.org`不能解除这个限制。顶层允许仍须满足§8所选真实IMEISV16子集；仅IMEI表示尚未裁决；不自动将嵌套值搬到顶层，也不复制两个位置。

V-NODE-02文本场景：MSC-Number+MME-Name+MME-Realm是明确允许组合，缺MME-Number不因此失败；MME-Name单独出现则残缺。IP-SM-GW-Number单独出现合法，不能强加Realm。带IP-SM-GW-Realm的完整组合保留为待裁决样本，不能标“已经验证必拒绝”。

本轮新增两帧成对候选与四个AVP片段；所有运行、独立实现复核及真实对端验证仍Planned。标准SMS pair和其余主命令没有被通用DCCA pair替代。

## 8. 第十六轮：有限子集与未决规则候选

来源为[主流程审计 §12](diameterkit-mainflow-audit.md)。下面是文档级输入与预期，全部Planned；字节算术检查不证明完整SMS合规、身份真实或独立互操作。固定Profile不随对照PDF升级。

### V-UE-SUBSET-01：真实IMEISV16文本

RO-UE-IMEISV16 / FIELD-09。调用者明确声明真实IMEISV `4901542032375101`；Type=0，Value为16个ASCII数字，无Vendor，M选0（该组及成员的M位允许可选，不推广成其他AVP的flags策略）。User-Equipment-Info长度44，Type长度12，Value长度24。

```avp-hex
000001ca 00 00002c
000001cb 00 00000c 00000000
000001cc 00 000018 34393031353432303332333735313031
```

仅用于允许的顶层SMS路径及明确子集解释，不证明PS嵌套路径可用。前导零保留，不查TAC/设备库，不做Luhn或从IMSI推导。将Value改为15数字或15数字加F时，正确调整Length/padding后仍可作为原始OctetString保留；SMS helper应报告未支持映射，不能误报整数宽度错、静默省略或自动补SVN。缺必需Type/Value、父边界溢出则独立按已知规则失败。

### V-REALM-CAND-01：可解码不等于组合已裁决

S6C-NODE-REALM / FIELD-07 / DECISION-01。Serving-Node10415/2401内含IP-SM-GW-Number3100=`12345`的TBCD、Name3101=`gw.example`、Realm3112=`example`；沿用S6c应用覆盖V/M=1。父长度72，子AVP含padding占16、24、20字节。

```avp-hex
00000961 c0 000048 000028af
00000c1c c0 00000f 000028af 2143f5 00
00000c1d c0 000016 000028af 67772e6578616d706c65 0000
00000c28 c0 000013 000028af 6578616d706c65 00
```

这不是完整SRA或合法组合的正例。Wire可表示，Application必须报告Realm组合未裁决并保留原值；薄解释器不交付已验证路由。删除Realm整个20字节AVP并将父Length改为52时，得到明确的Number+Name组合候选；仍须完整SRA/原SRR条件才可判主流程通过。Realm改为非法DiameterIdentity、同成员重复或缺Number时，分别检查可证明的语法/次数/组合违规，不被“尚未裁决”掩盖。

### V-DECISION-01：规范冲突不是缺上下文

文本候选，APP-01/DECISION-01：输入标准SMS失败事件且需SMS-Result，或需SM-Sequence-Number的串接事件；CR0081去除offline-only，但32.274在线表与32.299使用表仍相冲突。标准helper在接纳前返回有RuleID的本地未支持，不产生Completion/写入；不能丢字段后伪装成功。接收保留有界原值及未决报告，不仅据表“-”生成5008。任意添加“已确认”业务facts不改变规则状态，也不能用FieldSelection跳过条件。

相关答案若Wire/Exchange通过，应结束网络等待；原值可交付但不宣称完整应用验证。非法帧、无效关联、已知字段违规仍独立失败；未决报告不授权重试、退款或查询业务存储。

### V-GT-SUBSET-01：显式GT与收到的数字文本

文本候选，FIELD-08：显式E.164 GT输入1至15个ASCII数字，以family8和UTF8数字编码；V-ADDR-01是其中一个字节候选。0位、16位、加号、空格及TBCD不通过该helper；不意味着所有family8 Address只能表示数字GT。收到`1234`不凭外形判为GT或Point Code。收到`pc=1234;gt=12025550123`时可保留受限原始payload，但没有已采用复合语法就不能解释/生成成功，也不能仅因helper不支持断言线路UTF8非法。

两端均使用本库的回环不能代替独立fixture；以上子集、删除Realm变体、未支持输入及真实节点/计费结果仍须逐项执行验收。

## 9. 第十七轮：标准SMS、S6c与SGd完整成对候选

本节按[主流程审计 §13](diameterkit-mainflow-audit.md)的固定版本与子集手工构造；8帧请求/成功答案和4帧替代失败答案都是候选，不是运行或对端验证结果。相同pair的两个答案代表独立测试场景；同一exchange只能完成一次。各帧Session-Id首位、R/P/E/T按头标示、答案继承本跳HBH和端到端E2E。成功解释同时要求已保存原请求、命令/角色/来源binding；不能只凭hex本身证明这些运行时前提。

共同环境：realm为`example.net`，DRA邻接预先授权；终端身份分别为`ocs.example`、`hss.example`、`mme.example`、`iwmsc.example`。样本假定已具备对应应用/传输能力，未编码CER或SCTP建链。无Proxy-Info的测试路径不代表实际DRA不能加入代理字段；若加入须按规范回显并重算长度。本地部署将样本Service-Identifier=1约定为短信事件单位；不是标准给它分配了“短信业务码1”。所有标识仅用于隔离fixture，不作为生产发号算法。

### V-SMS-CCR-01：标准SMS ECUR INITIAL

32.274 §5.3.2.1、§6.2.1.2.1与在线表6.3.3.1，32.299 §6.4.2/7.1.9。Command272/App4，HBH31323334/E2E41424344；Session=`ctf.example;17;1`，Origin为充当IP-SM-GW的`ctf.example`，Destination为`ocs.example`。CC-Type=INITIAL(1)、Number=0、Service-Context=`18.32274@3gpp.org`；Multiple-Services-Indicator=1，MSCC中RSU请求1个service-specific unit并带Service-Identifier=1。INITIAL不带Requested-Action，不套用IEC或顶层RSU/Service-Identifier。

Subscription-Id为Type0/MSISDN `12025550123`；Time保留e0000000（转换显式era0）。真实IMEISV取§8的16位子集。Service-Information下PS-Information含RAT=06、Ro ULI=`80 00f110 0001`（type128、MCC001/MNC01、TAC1）、时区`23 00`（+8小时，无DST）；SMS-Information含Node1、MessageType0、Originator-Interface Type1，以及Recipient-Info内Destination-Interface Type2和显式GT `12025550123`。这些值由应用明确提供，不从DRA CER、IMSI或系统时区推导。

构造选择只启用本例所列可选项。最短Service-Context允许未选择的OM/OC缺省；M/C及已给字段仍校验。没有MMS地址、失败结果、串接序号、UDH或Device Trigger事件；不得把实际需要这些信息的输入静默删掉以匹配本例。总长556（0x22c）。

```message-hex
01 00 02 2c c0 00 01 10 00 00 00 04 31 32 33 34 41 42 43 44
00 00 01 07 40 00 00 18 63 74 66 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 31
00 00 01 08 40 00 00 13 63 74 66 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 02 40 00 00 0c 00 00 00 04
00 00 01 cd 40 00 00 19 31 38 2e 33 32 32 37 34 40 33 67 70 70 2e 6f 72 67 00 00 00
00 00 01 a0 40 00 00 0c 00 00 00 01
00 00 01 9f 40 00 00 0c 00 00 00 00
00 00 01 25 40 00 00 13 6f 63 73 2e 65 78 61 6d 70 6c 65 00
00 00 00 37 40 00 00 0c e0 00 00 00
00 00 01 bb 40 00 00 28
00 00 01 c2 40 00 00 0c 00 00 00 00
00 00 01 bc 40 00 00 13 31 32 30 32 35 35 35 30 31 32 33 00
00 00 01 c7 40 00 00 0c 00 00 00 01
00 00 01 c8 40 00 00 2c
00 00 01 b5 40 00 00 18
00 00 01 a1 40 00 00 10 00 00 00 00 00 00 00 01
00 00 01 b7 40 00 00 0c 00 00 00 01
00 00 01 ca 00 00 00 2c
00 00 01 cb 00 00 00 0c 00 00 00 00
00 00 01 cc 00 00 00 18 34 39 30 31 35 34 32 30 33 32 33 37 35 31 30 31
00 00 03 69 c0 00 00 d8 00 00 28 af
00 00 03 6a c0 00 00 40 00 00 28 af
00 00 00 15 80 00 00 0d 00 00 28 af 06 00 00 00
00 00 00 16 80 00 00 12 00 00 28 af 80 00 f1 10 00 01 00 00
00 00 00 17 80 00 00 0e 00 00 28 af 23 00 00 00
00 00 07 d0 c0 00 00 8c 00 00 28 af
00 00 07 e0 c0 00 00 10 00 00 28 af 00 00 00 01
00 00 07 d7 c0 00 00 10 00 00 28 af 00 00 00 00
00 00 07 d9 c0 00 00 1c 00 00 28 af
00 00 07 d6 c0 00 00 10 00 00 28 af 00 00 00 01
00 00 07 ea c0 00 00 44 00 00 28 af
00 00 07 d2 c0 00 00 1c 00 00 28 af
00 00 07 d6 c0 00 00 10 00 00 28 af 00 00 00 02
00 00 07 da c0 00 00 19 00 00 28 af 00 08 31 32 30 32 35 35 35 30 31 32 33 00 00 00
```

### V-SMS-CCA-01：ECUR初始授权

同Session/CC类型1/编号0与HBH/E2E；顶层Result2001，Origin为授权OCS。MSCC带GSU的1个事件单位、Service-Identifier1及组内Result2001；总长188（0xbc）。按原请求服务键关联，授权单元由本例明确给出；初始授权不等于整个ECUR会话、短信投递或最终扣费已完成。后续TERMINATION由应用发起，本节不伪造它。

```message-hex
01 00 00 bc 40 00 01 10 00 00 00 04 31 32 33 34 41 42 43 44
00 00 01 07 40 00 00 18 63 74 66 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 31
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 08 40 00 00 13 6f 63 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 02 40 00 00 0c 00 00 00 04
00 00 01 a0 40 00 00 0c 00 00 00 01
00 00 01 9f 40 00 00 0c 00 00 00 00
00 00 01 c8 40 00 00 38
00 00 01 af 40 00 00 18
00 00 01 a1 40 00 00 10 00 00 00 00 00 00 00 01
00 00 01 b7 40 00 00 0c 00 00 00 01
00 00 01 0c 40 00 00 0c 00 00 07 d1
```

### V-SMS-CCA-02：信用额度不足

替代上一成功答案：基础4012，R/E/T=0、P=1；无GSU/MSCC，保留匹配CC身份，总长132（0x84）。不得把缺GSU当这个失败答案的成功字段缺失；协议拒绝不授权自动再扣费、改服务键或切OCS。

```message-hex
01 00 00 84 40 00 01 10 00 00 00 04 31 32 33 34 41 42 43 44
00 00 01 07 40 00 00 18 63 74 66 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 31
00 00 01 0c 40 00 00 0c 00 00 0f ac
00 00 01 08 40 00 00 13 6f 63 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 02 40 00 00 0c 00 00 00 04
00 00 01 a0 40 00 00 0c 00 00 00 01
00 00 01 9f 40 00 00 0c 00 00 00 00
```

### V-SRR-01：普通有MSISDN的投递路由查询

29.338 §5.2.1.1与§5.3.2.3，Command8388647/App16777312，HBH51525354/E2E61626364，Session=`gmsc.example;17;1`。Origin为`gmsc.example`，目标`hss.example`；Auth-Session-State277=1。MSISDN701=`12025550123`、SC-Address3300=`12025550100`均只有TBCD数字，不含TON/NPI。SM-RP-MTI3308=0明确普通MT，不请求两号码、不用MSISDN-less、不设SM-Delivery-Not-Intended、SMSMI或相关条件；总长196（0xc4）。

```message-hex
01 00 00 c4 c0 80 00 27 01 00 00 60 51 52 53 54 61 62 63 64
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 31 00 00 00
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 14 67 6d 73 63 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 bd c0 00 00 12 00 00 28 af 21 20 55 05 21 f3 00 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c ec c0 00 00 10 00 00 28 af 00 00 00 00
```

### V-SRA-01：完整MME节点组合

29.338表5.2.1.1-2、§5.3.2.4/5.3.3.6。Origin为`hss.example`，Result2001，User-Name=`001010123456789`明确是本例IMSI；Serving-Node2401包含MME-Name2402=`mme.example`、MME-Realm2408=`example.net`、MME-Number1645=`12025550999`。三叶在S6c覆盖中M=1。总长216（0xd8）。

原SRR意图投递且无SMS Router MT相关标识替换；因此不能仅保留IMSI删除Serving-Node。HSS获知注册信息是本例外部前提，不由库查询。拿到节点数据后是否发送TFR及路由授权归应用，本库不自动连MME；同组数据也不证明该MME此刻可达。

```message-hex
01 00 00 d8 40 80 00 27 01 00 00 60 51 52 53 54 61 62 63 64
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 31 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 00 01 40 00 00 17 30 30 31 30 31 30 31 32 33 34 35 36 37 38 39 00
00 00 09 61 c0 00 00 50 00 00 28 af
00 00 09 62 c0 00 00 17 00 00 28 af 6d 6d 65 2e 65 78 61 6d 70 6c 65 00
00 00 09 68 c0 00 00 17 00 00 28 af 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 06 6d c0 00 00 12 00 00 28 af 21 20 55 05 99 f9 00 00
```

### V-SRA-02：未知用户

替代成功答案，Experimental-Result内Vendor10415/Code5001，正常答案E=0；无基础Result-Code，Auth-Session-State1及Origin仍必需；总长132（0x84）。不要求失败答案带成功IMSI或节点，不把10415/5001混成基础5001，不对答案生成错误答案。

```message-hex
01 00 00 84 40 80 00 27 01 00 00 60 51 52 53 54 61 62 63 64
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 31 00 00 00
00 00 01 29 40 00 00 20
00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 2a 40 00 00 0c 00 00 13 89
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-OFR-01：普通MO，附EPS位置子集

29.338 §6.2.1.1/6.3.2.3。Command8388645/App16777313，HBH71727374/E2E81828384，Session=`mme.example;17;1`；MME→SMS-IWMSC，Destination=`iwmsc.example`，Auth-Session-State1。SC-Address为`12025550100`；User-Identifier3102同时含已知IMSI与MSISDN（与上例一致），不能只因组语法可选就丢掉当前已知身份。无SMSMI，故未触发其SM-Delivery-Outcome条件；本例不是T4。

EPS-Location1496/MME-Location1600包含ECGI1602=`00f11000000001`及TAI1603=`00f1100001`，M均0；无Ro location tag、ASN.1/SGs/NAS包头。SM-RP-UI3301为14字节`01 00 0b 91 2120550521f3 00 04 01 41`：SMS-SUBMIT、MR0、11位目标地址、TP-PID0、8-bit DCS04、UDL1/数据41，无VP/UDH。TP地址里的91属于TPDU，不能复制到SC-Address。总长308（0x134）。

TPDU结构参考23.040 §9.2.2.2，DCS参考23.038 §4；这里只校对fixture，库的SM-RP-UI仍是最大200字节的opaque载荷，不替应用解释消息内容。

```message-hex
01 00 01 34 c0 80 00 25 01 00 00 61 71 72 73 74 81 82 83 84
00 00 01 07 40 00 00 18 6d 6d 65 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 31
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 6d 6d 65 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 15 69 77 6d 73 63 2e 65 78 61 6d 70 6c 65 00 00 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c 1e c0 00 00 38 00 00 28 af
00 00 00 01 40 00 00 17 30 30 31 30 31 30 31 32 33 34 35 36 37 38 39 00
00 00 02 bd c0 00 00 12 00 00 28 af 21 20 55 05 21 f3 00 00
00 00 05 d8 80 00 00 40 00 00 28 af
00 00 06 40 80 00 00 34 00 00 28 af
00 00 06 42 80 00 00 13 00 00 28 af 00 f1 10 00 00 00 01 00
00 00 06 43 80 00 00 11 00 00 28 af 00 f1 10 00 01 00 00 00
00 00 0c e5 c0 00 00 1a 00 00 28 af 01 00 0b 91 21 20 55 05 21 f3 00 04 01 41 00 00
```

### V-OFA-01：MO提交协议成功

Origin=`iwmsc.example`、基础2001、Auth-Session-State1；不带可选TPDU响应或仅T4条件要求的External-Identifier。总长112（0x70）。代表对该提交返回成功，不推导目标UE已收到短信，也不查询SMS-SC数据库。

```message-hex
01 00 00 70 40 80 00 25 01 00 00 61 71 72 73 74 81 82 83 84
00 00 01 07 40 00 00 18 6d 6d 65 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 31
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 15 69 77 6d 73 63 2e 65 78 61 6d 70 6c 65 00 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-OFA-02：服务中心拥塞

正常答案Experimental10415/5555，带必需SM-Delivery-Failure-Cause3303，其必需Enum3304=4（SC-CONGESTION），可选诊断省略。总长160（0xa0）；不另带基础2001。原因4属于MO允许子集；库保留原因供应用处理，不启动重投递计时。

```message-hex
01 00 00 a0 40 80 00 25 01 00 00 61 71 72 73 74 81 82 83 84
00 00 01 07 40 00 00 18 6d 6d 65 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 31
00 00 01 29 40 00 00 20
00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 2a 40 00 00 0c 00 00 15 b3
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 15 69 77 6d 73 63 2e 65 78 61 6d 70 6c 65 00 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e7 c0 00 00 1c 00 00 28 af
00 00 0c e8 c0 00 00 10 00 00 28 af 00 00 00 04
```

### V-TFR-01：向MME发送普通MT

29.338 §6.2.2.1/6.3.2.5。Command8388646/App16777313，HBH91929394/E2Ea1a2a3a4，Session=`gmsc.example;17;2`；Origin=`gmsc.example`、必需Destination-Host=`mme.example`。本例给定原SRA节点数据与独立部署授权，User-Name为IMSI，SC-Address同上，MME-Number1645为`12025550999`；**本SGd使用位置M=0**，不能复制S6c的M=1覆盖。SM-Delivery-Timer3306=30秒、Start-Time3307=ed2723f0（era0：2026-01-30 12:34:56 UTC）；由应用提供明确时间，库不据此调度业务重传。无更多待发消息，TFR-Flags可省；可选SMS-GMSC-Address/最大重传时间未选择。

SM-RP-UI为20字节`04 0b 91 2120550521f3 00 04 62100321436500 01 41`：SMS-DELIVER、无后续消息、源号码11位、PID0/DCS04、SCTS=2026-01-30 12:34:56 UTC、UDL1/41。SCTS与Diameter Start-Time在本例指向同一时刻但分别保留；测试时钟须固定在该发送时刻，本隔离fixture不声称现实交付时限仍有效；时间解释不得用当前系统年份猜era。总长268（0x10c）。

```message-hex
01 00 01 0c c0 80 00 26 01 00 00 61 91 92 93 94 a1 a2 a3 a4
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 32 00 00 00
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 14 67 6d 73 63 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 13 6d 6d 65 2e 65 78 61 6d 70 6c 65 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 00 01 40 00 00 17 30 30 31 30 31 30 31 32 33 34 35 36 37 38 39 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c e5 c0 00 00 20 00 00 28 af 04 0b 91 21 20 55 05 21 f3 00 04 62 10 03 21 43 65 00 01 41
00 00 06 6d 80 00 00 12 00 00 28 af 21 20 55 05 99 f9 00 00
00 00 0c ea c0 00 00 10 00 00 28 af 00 00 00 1e
00 00 0c eb c0 00 00 10 00 00 28 af ed 27 23 f0
```

### V-TFA-01：MT协议成功

Origin=`mme.example`，基础2001、Auth-Session-State1；成功可省可选SM-RP-UI、User-Identifier和位置。总长112（0x70）。不以成功答案缺User-Identifier判缺必需AVP；库仅解释协议结果，应用决定业务状态如何更新。

```message-hex
01 00 00 70 40 80 00 26 01 00 00 61 91 92 93 94 a1 a2 a3 a4
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 32 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 6d 6d 65 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-TFA-02：接收设备容量不足

正常答案Experimental10415/5555，SM-Delivery-Failure-Cause/Enum=0（MEMORY_CAPACITY_EXCEEDED），总长160（0xa0）。原因0在MT失败子集中；不能因数值为0判“缺原因”，也不能与MO的SC拥塞原因交换。Requested-Retransmission-Time未给，库不猜重试时刻。

```message-hex
01 00 00 a0 40 80 00 26 01 00 00 61 91 92 93 94 a1 a2 a3 a4
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 32 00 00 00
00 00 01 29 40 00 00 20
00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 2a 40 00 00 0c 00 00 15 b3
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 6d 6d 65 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e7 c0 00 00 1c 00 00 28 af
00 00 0c e8 c0 00 00 10 00 00 28 af 00 00 00 00
```

### V-PAIR-NEG-01：完整pair的最小负向修改

以下均保留完整原帧；涉及长度改变时必须同步父AVP/消息Length及padding。只变身份字段时也须遵守其固定长度。语义预期仍Planned。

| 输入变体 | 可观察预期 / 层 |
| --- | --- |
| SMS CCR将MSCC的RSU/Service-Identifier搬到顶层 | 固定SMS应用路径不通过；不是通用DCCA通过就可放行 |
| SMS CCA把CC编号改1或服务键改2；或顶层2001但MSCC4012 | 前者原请求关联失败；后者不解释为可用额度。失败顶层4012缺GSU本身合法 |
| 将Ro RAT06改为00000006；或把GTP RAT10当NR | 前者已知宽度违规；后者本29.061字典未采用该值，报告未支持，不沿用GTP枚举 |
| Ro ULI删type128、把SGd TAI加type128；EPS叶加入SGs TLV头 | 已知父路径值长度/布局错误；不能互相自动修复 |
| 时区2300改成2301 | 仍保留表示；offset已含DST，不再加1小时；不能借纯转换修改原比较字节 |
| SRA成功缺Serving-Node；MME组合仅保留Name | 原SRR意图投递/节点组合检查失败；不能强行降级成“不用于投递” |
| SRA改为IP-SM-GW-Number+Name且无Realm；再加合法Realm | 前者为§12明确组合候选，后者报告未裁决；须更新业务元数据为IP-SM-GW节点，不能继续沿用MME目的解释 |
| OFR缺已知MSISDN或IMSI；加入SMSMI但不补Delivery-Outcome | 分别验证当前输入的身份条件、SMSMI联动条件；不由字典的可选成员掩盖 |
| TFR删除Destination-Host或去掉发往MME必需的号码 | 静态/角色条件不通过；不从邻接DRA名补目标 |
| OFA实验5555删Cause；把OFA原因改0、TFA原因改4 | 缺必需失败原因或命令原因子集不符；不能将所有5555解释成同一种错误 |
| 任一失败答案额外加入基础2001、E置1后仍按正常应用格式解释 | 结果冲突或错误语法路径不符；不交付成功、不重传、不再回答案 |
| 正确body但错HBH、旧连接代回调、未授权DRA成功Origin | 分别按pending、代际与来源约束处理，不猜原请求、不产生第二完成 |

## 10. 第十七轮：深层依赖片段候选

### V-FURNISH-01：定义可复用但不自动进入SMS

32.299 §7.2.155–157、29.061项目2；Furnish865内Charging-Id2为OctetString四字节00000001、Free-Format-Data866为00ff、Append867=0。父长60（0x3c）；仅在允许的CCA/MSCC位置解释。省867表示Overwrite指示，但库不执行存储；删必需字段与保留零长度值不同。不能把RADIUS的0206子头加到Charging-Id，也不能把该值当Session-Id或E2E。

```avp-hex
00 00 03 61 c0 00 00 3c 00 00 28 af
00 00 00 02 80 00 00 10 00 00 28 af 00 00 00 01
00 00 03 62 c0 00 00 0e 00 00 28 af 00 ff 00 00
00 00 03 63 c0 00 00 10 00 00 28 af 00 00 00 00
```

### V-MMS-ADDR-01：明确选择PLMN文本格式

32.299 §7.2.7/9/125与OMA MMS Encapsulation1.3 §8。Originator-Address886含Type899=1、Data897=`+12025550123/TYPE=PLMN`；所有键Vendor10415、V/M=1，父长64（0x40）。只定义一个有限构造子集，不将旧模板不含后缀的输入静默改写成它，也不据本例宣称所有邮箱/URI/IPv6地址已校验。

```avp-hex
00 00 03 76 c0 00 00 40 00 00 28 af
00 00 03 83 c0 00 00 10 00 00 28 af 00 00 00 01
00 00 03 81 c0 00 00 22 00 00 28 af 2b 31 32 30 32 35 35 35 30 31 32 33 2f 54 59 50 45 3d 50 4c 4d 4e 00 00
```

FIELD-10/11/12、DEP-06、PAIR-02见覆盖基线§19。独立fixture/运行、真实DRA/OCS、所有命令和六模板兼容仍未通过；文档中的字节算术只证明可复核布局。

## 11. 第十八轮：ECUR终止、IEC与退款结构候选（后续版本）

来源：固定32.299 §6.3.3/6.3.4、7.1.9、7.2.171/175，32.274 §6.2.1/6.3，RFC4006 §5.4/6.3/6.4/8.3；规则见[主流程§14.1](diameterkit-mainflow-audit.md)。这里有三个不同exchange；只有终止请求沿用§9 INITIAL的业务Session。新exchange重新分配HBH/E2E，T=0；活动请求的原身份重传另按R33，不把“发TERMINATION”当作重传INITIAL。

所有Ro请求仍选最短`18.32274@3gpp.org`，终端OCS=`ocs.example`，邻接DRA须另外授权。OM/OC只选择样本中的字段；不声称当前有真实设备、位置或delivery-report输入，所以不自动复制INITIAL可选字段，也不自动加入SM-Status/Discharge-Time。M/C与实际事件用途仍检查；不能用本例的选择隐藏真实输入触发的未决SMS-Result条件。TERMINATION的Event-Timestamp为`e000003c`，比§9 INITIAL的`e0000000`晚60秒；其余新Ro请求使用`ed2723f0`（era0：2026-01-30 12:34:56 UTC）。均为固定场景时间，不读取运行时钟。

### V-TERM-CCR-01：同一ECUR会话的最终用量

沿用V-SMS-CCR-01的Session=`ctf.example;17;1`，CC-Type=3、CC-Number=1；本例没有UPDATE。HBH=b1b2b3b4/E2E=c1c2c3c4，与INITIAL不同。MSCC内Service-Identifier=1、USU446/CC-Service-Specific-Units417=1，Reporting-Reason10415/872=FINAL(2)直接位于MSCC。无RSU和Requested-Action。应用提供已用量，不由库推算或扣账。 总长464（0x1d0）。

```message-hex
01 00 01 d0 c0 00 01 10 00 00 00 04 b1 b2 b3 b4 c1 c2 c3 c4
00 00 01 07 40 00 00 18 63 74 66 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 31
00 00 01 08 40 00 00 13 63 74 66 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 02 40 00 00 0c 00 00 00 04
00 00 01 cd 40 00 00 19 31 38 2e 33 32 32 37 34 40 33 67 70 70 2e 6f 72 67 00 00 00
00 00 01 a0 40 00 00 0c 00 00 00 03
00 00 01 9f 40 00 00 0c 00 00 00 01
00 00 01 25 40 00 00 13 6f 63 73 2e 65 78 61 6d 70 6c 65 00
00 00 00 37 40 00 00 0c e0 00 00 3c
00 00 01 bb 40 00 00 28
00 00 01 c2 40 00 00 0c 00 00 00 00
00 00 01 bc 40 00 00 13 31 32 30 32 35 35 35 30 31 32 33 00
00 00 01 c7 40 00 00 0c 00 00 00 01
00 00 01 c8 40 00 00 3c
00 00 01 be 40 00 00 18
00 00 01 a1 40 00 00 10 00 00 00 00 00 00 00 01
00 00 01 b7 40 00 00 0c 00 00 00 01
00 00 03 68 c0 00 00 10 00 00 28 af 00 00 00 02
00 00 03 69 c0 00 00 98 00 00 28 af
00 00 07 d0 c0 00 00 8c 00 00 28 af
00 00 07 e0 c0 00 00 10 00 00 28 af 00 00 00 01
00 00 07 d7 c0 00 00 10 00 00 28 af 00 00 00 00
00 00 07 d9 c0 00 00 1c 00 00 28 af
00 00 07 d6 c0 00 00 10 00 00 28 af 00 00 00 01
00 00 07 ea c0 00 00 44 00 00 28 af
00 00 07 d2 c0 00 00 1c 00 00 28 af
00 00 07 d6 c0 00 00 10 00 00 28 af 00 00 00 02
00 00 07 da c0 00 00 19 00 00 28 af 00 08 31 32 30 32 35 35 35 30 31 32 33 00 00 00
```

### V-TERM-CCA-01：终止协议成功

与原TERM的Session/CC/HBH/E2E配对，基础2001、E=0；无GSU、MSCC或Cost-Information仍可成立。该结果不能当作新额度，也不是本库对外部账本提交或短信送达的确认。 总长132（0x84）。

```message-hex
01 00 00 84 40 00 01 10 00 00 00 04 b1 b2 b3 b4 c1 c2 c3 c4
00 00 01 07 40 00 00 18 63 74 66 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 31
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 08 40 00 00 13 6f 63 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 02 40 00 00 0c 00 00 00 04
00 00 01 a0 40 00 00 0c 00 00 00 03
00 00 01 9f 40 00 00 0c 00 00 00 01
```

### V-TERM-CCA-02：对端不认识会话

V-TERM-CCR-01的替代答案，基础5002 DIAMETER_UNKNOWN_SESSION_ID、E=0；不强制GSU。库报告失败并结束该exchange，不自动补发INITIAL、不变更CC编号、不重建计费会话。 总长132（0x84）。

```message-hex
01 00 00 84 40 00 01 10 00 00 00 04 b1 b2 b3 b4 c1 c2 c3 c4
00 00 01 07 40 00 00 18 63 74 66 2e 65 78 61 6d 70 6c 65 3b 31 37 3b 31
00 00 01 0c 40 00 00 0c 00 00 13 8a
00 00 01 08 40 00 00 13 6f 63 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 02 40 00 00 0c 00 00 00 04
00 00 01 a0 40 00 00 0c 00 00 00 03
00 00 01 9f 40 00 00 0c 00 00 00 01
```

### V-IEC-CCR-01：IEC直接扣费请求

独立Session=`ctf.example;18;1`，CC-Type=4、CC-Number=0、Requested-Action=0 DIRECT_DEBITING；HBH=b1b2b3b5/E2E=c1c2c3c5。MSCC含RSU/事件单位1和Service-Identifier=1，不使用顶层单位或金额。本例为服务执行前的请求，尚无失败SM事务需要报告；不是refund。 总长460（0x1cc）。

```message-hex
01 00 01 cc c0 00 01 10 00 00 00 04 b1 b2 b3 b5 c1 c2 c3 c5
00 00 01 07 40 00 00 18 63 74 66 2e 65 78 61 6d 70 6c 65 3b 31 38 3b 31
00 00 01 08 40 00 00 13 63 74 66 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 02 40 00 00 0c 00 00 00 04
00 00 01 cd 40 00 00 19 31 38 2e 33 32 32 37 34 40 33 67 70 70 2e 6f 72 67 00 00 00
00 00 01 a0 40 00 00 0c 00 00 00 04
00 00 01 9f 40 00 00 0c 00 00 00 00
00 00 01 25 40 00 00 13 6f 63 73 2e 65 78 61 6d 70 6c 65 00
00 00 00 37 40 00 00 0c ed 27 23 f0
00 00 01 bb 40 00 00 28
00 00 01 c2 40 00 00 0c 00 00 00 00
00 00 01 bc 40 00 00 13 31 32 30 32 35 35 35 30 31 32 33 00
00 00 01 c7 40 00 00 0c 00 00 00 01
00 00 01 b4 40 00 00 0c 00 00 00 00
00 00 01 c8 40 00 00 2c
00 00 01 b5 40 00 00 18
00 00 01 a1 40 00 00 10 00 00 00 00 00 00 00 01
00 00 01 b7 40 00 00 0c 00 00 00 01
00 00 03 69 c0 00 00 98 00 00 28 af
00 00 07 d0 c0 00 00 8c 00 00 28 af
00 00 07 e0 c0 00 00 10 00 00 28 af 00 00 00 01
00 00 07 d7 c0 00 00 10 00 00 28 af 00 00 00 00
00 00 07 d9 c0 00 00 1c 00 00 28 af
00 00 07 d6 c0 00 00 10 00 00 28 af 00 00 00 01
00 00 07 ea c0 00 00 44 00 00 28 af
00 00 07 d2 c0 00 00 1c 00 00 28 af
00 00 07 d6 c0 00 00 10 00 00 28 af 00 00 00 02
00 00 07 da c0 00 00 19 00 00 28 af 00 08 31 32 30 32 35 35 35 30 31 32 33 00 00 00
```

### V-IEC-CCA-01：IEC授权与opaque退款凭据

基础2001、匹配服务级2001、GSU事件单位1；同MSCC带Refund-Information10415/2022=`00 ff`。token首字节为零且不是UTF8文字，必须按长度保留。GSU位置采用32.299的IEC过程及MSCC规则，32.274表的用语差异见§14.1。此答案是该操作的协议结果，不自动产生短信投递任务或退款历史。 总长204（0xcc）。

```message-hex
01 00 00 cc 40 00 01 10 00 00 00 04 b1 b2 b3 b5 c1 c2 c3 c5
00 00 01 07 40 00 00 18 63 74 66 2e 65 78 61 6d 70 6c 65 3b 31 38 3b 31
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 08 40 00 00 13 6f 63 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 02 40 00 00 0c 00 00 00 04
00 00 01 a0 40 00 00 0c 00 00 00 04
00 00 01 9f 40 00 00 0c 00 00 00 00
00 00 01 c8 40 00 00 48
00 00 01 af 40 00 00 18
00 00 01 a1 40 00 00 10 00 00 00 00 00 00 00 01
00 00 01 b7 40 00 00 0c 00 00 00 01
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 07 e6 c0 00 00 0e 00 00 28 af 00 ff 00 00
```

### V-IEC-CCA-02：IEC拒绝额度

V-IEC-CCR-01的替代答案：基础4012 CREDIT_LIMIT_REACHED，E=0，无MSCC/GSU/token。失败分支不因缺成功字段判坏包，不允许因其无额度而自动换OCS或业务重试。 总长132（0x84）。

```message-hex
01 00 00 84 40 00 01 10 00 00 00 04 b1 b2 b3 b5 c1 c2 c3 c5
00 00 01 07 40 00 00 18 63 74 66 2e 65 78 61 6d 70 6c 65 3b 31 38 3b 31
00 00 01 0c 40 00 00 0c 00 00 0f ac
00 00 01 08 40 00 00 13 6f 63 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 02 40 00 00 0c 00 00 00 04
00 00 01 a0 40 00 00 0c 00 00 00 04
00 00 01 9f 40 00 00 0c 00 00 00 00
```

下面两帧是**32.299退款结构与原凭据关联的候选，不是已收敛的标准SMS应用正例**。退款示意涉及既往服务失败，32.274的SMS-Result条件冲突仍按DECISION-01：标准SMS helper接纳前拒绝受影响分支，不能因为本候选没有3409就宣布可发送。它们用于验证线路结构、CC关联和token原样传递；完整Application预期是未支持/未裁决，待规范裁决或具名对端证据后才能转为正向支持样本。没有新增持久化接口。

### V-REFUND-CCR-01：退款结构及原token关联

应用明确提供V-IEC-CCA-01的服务键、原token和退款意图。新Session=`ctf.example;18;2`、Type=4、Number=0、Action=1 REFUND_ACCOUNT，HBH=b1b2b3b6/E2E=c1c2c3c6。MSCC中Service-Identifier=1及10415/2022原字节00ff；未选择可选RSU。不是0/1292，不把退款当作已完成IEC的T位重传。 总长452（0x1c4）。

```message-hex
01 00 01 c4 c0 00 01 10 00 00 00 04 b1 b2 b3 b6 c1 c2 c3 c6
00 00 01 07 40 00 00 18 63 74 66 2e 65 78 61 6d 70 6c 65 3b 31 38 3b 32
00 00 01 08 40 00 00 13 63 74 66 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 02 40 00 00 0c 00 00 00 04
00 00 01 cd 40 00 00 19 31 38 2e 33 32 32 37 34 40 33 67 70 70 2e 6f 72 67 00 00 00
00 00 01 a0 40 00 00 0c 00 00 00 04
00 00 01 9f 40 00 00 0c 00 00 00 00
00 00 01 25 40 00 00 13 6f 63 73 2e 65 78 61 6d 70 6c 65 00
00 00 00 37 40 00 00 0c ed 27 23 f0
00 00 01 bb 40 00 00 28
00 00 01 c2 40 00 00 0c 00 00 00 00
00 00 01 bc 40 00 00 13 31 32 30 32 35 35 35 30 31 32 33 00
00 00 01 c7 40 00 00 0c 00 00 00 01
00 00 01 b4 40 00 00 0c 00 00 00 01
00 00 01 c8 40 00 00 24
00 00 01 b7 40 00 00 0c 00 00 00 01
00 00 07 e6 c0 00 00 0e 00 00 28 af 00 ff 00 00
00 00 03 69 c0 00 00 98 00 00 28 af
00 00 07 d0 c0 00 00 8c 00 00 28 af
00 00 07 e0 c0 00 00 10 00 00 28 af 00 00 00 01
00 00 07 d7 c0 00 00 10 00 00 28 af 00 00 00 00
00 00 07 d9 c0 00 00 1c 00 00 28 af
00 00 07 d6 c0 00 00 10 00 00 28 af 00 00 00 01
00 00 07 ea c0 00 00 44 00 00 28 af
00 00 07 d2 c0 00 00 1c 00 00 28 af
00 00 07 d6 c0 00 00 10 00 00 28 af 00 00 00 02
00 00 07 da c0 00 00 19 00 00 28 af 00 08 31 32 30 32 35 35 35 30 31 32 33 00 00 00
```

### V-REFUND-CCA-01：退款应答结构

对应V-REFUND-CCR-01的基础2001、Type=4/Number=0；32.299退款应答过程不要求GSU。本候选只证明该结果形状可表示，不能消除请求端SMS语义未决，也不能用成功码伪造完整支持或外部事务证明。 总长132（0x84）。

```message-hex
01 00 00 84 40 00 01 10 00 00 00 04 b1 b2 b3 b6 c1 c2 c3 c6
00 00 01 07 40 00 00 18 63 74 66 2e 65 78 61 6d 70 6c 65 3b 31 38 3b 32
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 08 40 00 00 13 6f 63 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 02 40 00 00 0c 00 00 00 04
00 00 01 a0 40 00 00 0c 00 00 00 04
00 00 01 9f 40 00 00 0c 00 00 00 00
```

## 12. 第十八轮：S6c报告与提醒成对候选

来源：29.338 §5.2.2/5.2.3、5.3.2.5–8、5.3.3.14–23、7.3.1–2/7.3.11，RFC6733 §7.1/7.2。均为S6c App16777312、Auth-Session-State=1的普通有MSISDN场景。消息间关联由应用场景说明提供；库不把RDR成功挂起直到ALR，也不把不同Session拼成跨请求业务事务。

### V-RDR-01：报告MME容量不足

Command8388649、HBH=d1d2d3d4/E2E=e1e2e3e4、Session=`gmsc.example;18;1`。Origin=`gmsc.example`、Destination=`hss.example`；User-Identifier3102/MSISDN701=`12025550123`，SC3300=`12025550100`，号码均TBCD。3316/3317/3321三层为SM-Delivery-Outcome/MME-SM-Delivery-Outcome/UE_MEMORY_CAPACITY_EXCEEDED(0)。本例可由应用在V-TFA-02之后显式选择，不由库自动触发；非AbsentUser，未给诊断；未选择Single-Attempt-Delivery。 总长232（0xe8）。

```message-hex
01 00 00 e8 c0 80 00 29 01 00 00 60 d1 d2 d3 d4 e1 e2 e3 e4
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 31 38 3b 31 00 00 00
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 14 67 6d 73 63 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c 1e c0 00 00 20 00 00 28 af
00 00 02 bd c0 00 00 12 00 00 28 af 21 20 55 05 21 f3 00 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c f4 c0 00 00 28 00 00 28 af
00 00 0c f5 c0 00 00 1c 00 00 28 af
00 00 0c f9 c0 00 00 10 00 00 28 af 00 00 00 00
```

### V-RDA-01：报告成功且Alert MSISDN不同

基础2001；应用侧HSS给定Alert MSISDN=`12025550124`，不同于请求123，因此返回User-Identifier/MSISDN，不能由库根据号码递增构造此差异。此答案只确认本次协议操作；本库不建立MWD记录。若Alert号相同，另验不返回该可选组的正常变体。 总长144（0x90）。

```message-hex
01 00 00 90 40 80 00 29 01 00 00 60 d1 d2 d3 d4 e1 e2 e3 e4
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 31 38 3b 31 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c 1e c0 00 00 20 00 00 28 af
00 00 02 bd c0 00 00 12 00 00 28 af 21 20 55 05 21 f4 00 00
```

### V-RDA-02：MWD列表满

V-RDR-01的替代答案，Experimental-Result Vendor10415/Code5558，E=0；不附成功的Alert MSISDN。5558是RDA适用失败，不是SGd的5555+SM-Delivery-Failure-Cause。库不扩容或持久化MWD，不据此调度再次RDR。 总长132（0x84）。

```message-hex
01 00 00 84 40 80 00 29 01 00 00 60 d1 d2 d3 d4 e1 e2 e3 e4
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 31 38 3b 31 00 00 00
00 00 01 29 40 00 00 20
00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 2a 40 00 00 0c 00 00 15 b6
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-RDA-03：未知用户

同一RDR的另一替代答案，Experimental10415/5001，E=0。与基础5001 AVP_UNSUPPORTED命名空间不同；不因为码相同就合并。 总长132（0x84）。

```message-hex
01 00 00 84 40 80 00 29 01 00 00 60 d1 d2 d3 d4 e1 e2 e3 e4
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 31 38 3b 31 00 00 00
00 00 01 29 40 00 00 20
00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 2a 40 00 00 0c 00 00 13 89
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-ALR-01：HSS提醒服务中心

Command8388648、Session=`hss.example;18;1`，Origin=`hss.example`、Destination=`iwmsc.example`，HBH=d1d2d3d5/E2E=e1e2e3e5。SC-Address仍为100，User-Identifier给显式Alert MSISDN124。应用给定UE已恢复可达；Max-UE-Availability-Time10415/3329=`ed27251c`（固定era0时刻12:39:56 UTC），V=1/M=0，已知则包含。HSS普通路径不加入MME/SGSN通知分支的SMS-GMSC-Alert-Event/Serving-Node，不带SMSMI。 总长208（0xd0）。

```message-hex
01 00 00 d0 c0 80 00 28 01 00 00 60 d1 d2 d3 d5 e1 e2 e3 e5
00 00 01 07 40 00 00 18 68 73 73 2e 65 78 61 6d 70 6c 65 3b 31 38 3b 31
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 15 69 77 6d 73 63 2e 65 78 61 6d 70 6c 65 00 00 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c 1e c0 00 00 20 00 00 28 af
00 00 02 bd c0 00 00 12 00 00 28 af 21 20 55 05 21 f4 00 00
00 00 0d 01 80 00 00 10 00 00 28 af ed 27 25 1c
```

### V-ALA-01：提醒协议成功

Origin=`iwmsc.example`，基础2001，E=0。ALA不表示后续短信一定投递成功；库不执行Service Centre队列/提醒重试或清除MWD。 总长112（0x70）。

```message-hex
01 00 00 70 40 80 00 28 01 00 00 60 d1 d2 d3 d5 e1 e2 e3 e5
00 00 01 07 40 00 00 18 68 73 73 2e 65 78 61 6d 70 6c 65 3b 31 38 3b 31
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 15 69 77 6d 73 63 2e 65 78 61 6d 70 6c 65 00 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-ALA-02：基础协议忙错误

替代V-ALA-01：Result-Code=3004 DIAMETER_TOO_BUSY，E=1、P=1（flags60），保留请求Session/HBH/E2E，采用RFC6733通用错误答案结构，省略正常ALA所需Auth-Session-State。不是S6c专用业务失败；29.338明确该ALA过程没有定义应用错误，不能随意加Experimental5001/5555当作正常ALA失败。 总长100（0x64）。

```message-hex
01 00 00 64 60 80 00 28 01 00 00 60 d1 d2 d3 d5 e1 e2 e3 e5
00 00 01 07 40 00 00 18 68 73 73 2e 65 78 61 6d 70 6c 65 3b 31 38 3b 31
00 00 01 08 40 00 00 15 69 77 6d 73 63 2e 65 78 61 6d 70 6c 65 00 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 0c 40 00 00 0c 00 00 0b bc
```

### V-CYCLE-NEG-01：新pair的条件与关联反例

以下语义场景仍Planned；修改字段或删除AVP时同步所有父Length及Message Length。每个失败答案是同请求的替代场景，绝不交付第二次完成。

| 输入变体 | 预期 / 边界 |
| --- | --- |
| TERM复制INITIAL的CC-Type/Number；把合法TERMINATION用量当一次重传INITIAL提交 | 原请求CC关联失败；新业务消息不能使用活动exchange重传接口 |
| TERM USU的事件量改显式0；删除该单位；FINAL移到USU或命令顶层 | 零用量与缺值分开；本事件单位分支必须表达用量；FINAL父路径错误，不默认从原GSU推算用量 |
| TERM/退款成功CCA无GSU；IEC直接扣费成功CCA无GSU或显式零 | 前者不缺新额度；后者必须按原动作/单位解释，不能把缺值与零值合并；4012失败无GSU合法 |
| 把SM-Discharge-Time当成所有TERM的必填；已知失败SMS事件删3409再构造 | 前者无依据，Discharge-Time只适用delivery-report；后者不能绕过DECISION-01 |
| 退款复制原IEC HBH/E2E/T或把token00ff转成字符串；把1292改名当2022 | 新exchange不能借原身份；token保留长度/原bytes；标准与私有路径分开 |
| RDR仅有空3316、空3317，或同时有MSC与MME cause | “≥1cause”及MSC/MME互斥失败，即使字典每个成员均可选也不能通过 |
| RDR cause1且应用确有诊断却未提供；把诊断挂到另一节点 | 当前选择的数据不满足已知条件/归属；codec不联网补诊断。cause0与缺cause不同 |
| RDA5001/5558再加基础2001；ALA套用实验5555；E3004仍按正常ALA必填检查 | 结果冲突、没有该应用分支支持，或错误语法选择失败；不回答案、不自动重传 |
| ALR发往SGd App16777313；用邻接DRA当HSS Origin；更换成功答案Origin未经授权 | 标准命令/绑定不符；旧SGd alert须具名兼容，DRA不替代最终节点身份 |
| RDR无MSISDN就无条件填IMSI；ALR的SMSMI加HSS-ID/Originating-SIP-URI | T4/IMS无MSISDN分支不能靠空值自动选择；五请求SMSMI规则仍逐位置检查 |

## 13. 第十八轮：六模板的源码推导兼容片段（后续版本）

来源与14文件SHA-256见[主流程§14.3–14.5](diameterkit-mainflow-audit.md)。以下每个片段均为**条件化源码推导**，未编译/执行旧encoder，没有真实CCA/ACA或抓包。输入采用明确的假定变量值，不声称来自现网。只覆盖一个关键差异，不代替完整模板、base头、所有变量或独立对端证据。标准decoder应报告相应冲突；它们不是标准正例。

### V-LEGACY-GENERIC-01：通用模板的私有时区键

yppdia_avptemplate.cf的TIMEZONE：Vendor10425/Code23、INT32=0，M=1。与标准10415/23两字节OctetString不同，不以名字相同覆盖。 总长16（0x10）。

```avp-hex
00 00 00 17 c0 00 00 10 00 00 28 b9 00 00 00 00
```

### V-LEGACY-IEC-01：IEC模板STRING地址

假定$SMSC_ADDR在第一个NUL前为127.0.0.1；Client-Address10415/2018的STRING路径调用strlen写9字节ASCII，Length21、padding3。前两字节3132不是IPv4 family0001；不会自动补family或转换成四字节IPv4。真实Fetcher从21字节字段取值，不证明现场始终为IPv4文本。 总长24（0x18）。

```avp-hex
00 00 07 e2 c0 00 00 15 00 00 28 af 31 32 37 2e 30 2e 30 2e 31 00 00 00
```

### V-LEGACY-INIT-01：INITIAL模板用IMSI填设备号

假定$SUBSCRIPTIONID_IMSI取001010123456789；UE458/Type459=0/Value460为15字节ASCII，所有M=1。旧代码也可能取空字符串，缺输入不是自动省略AVP。新标准IMEISV16 helper拒绝这项替换；不能补数字造设备身份。 总长44（0x2c）。

```avp-hex
00 00 01 ca 40 00 00 2c
00 00 01 cb 40 00 00 0c 00 00 00 00
00 00 01 cc 40 00 00 17 30 30 31 30 31 30 31 32 33 34 35 36 37 38 39 00
```

### V-LEGACY-TERM-01：TERMINATION模板的正确用量片段

假定$CC_SSU=1；MSCC/USU事件单位1与直接Reporting-Reason=FINAL2，与所选标准位置相符。其他设备、地址和Discharge-Time条件仍未因此通过，不能把一个正确片段升级成整个模板合规。 总长48（0x30）。

```avp-hex
00 00 01 c8 40 00 00 30
00 00 01 be 40 00 00 18
00 00 01 a1 40 00 00 10 00 00 00 00 00 00 00 01
00 00 03 68 c0 00 00 10 00 00 28 af 00 00 00 02
```

### V-LEGACY-REFUND-01：退款模板按420编码Grouped

模板MSCC/RSU/CC_MONEY的Code实际为0/420，内含Unit-Value445（Value-Digits447=1、Exponent429=0）和Currency425=840；源编码器按模板类型构组。标准420是四字节CC-Time，此56字节组不合标准类型；仅该具名诊断片段按Grouped展开。不能让标准字典变成Grouped，也不能自动改413后宣称旧对端兼容。另一个顶层私有token仍为0/1292，不与标准2022归并。 总长56（0x38）。

```avp-hex
00 00 01 a4 40 00 00 38
00 00 01 bd 40 00 00 24
00 00 01 bf 40 00 00 10 00 00 00 00 00 00 00 01
00 00 01 ad 40 00 00 0c 00 00 00 00
00 00 01 a9 40 00 00 0c 00 00 03 48
```

### V-LEGACY-SCAP-01：SCAP1声明与私有单位

两个顶层组：VSAI260含Vendor193和Acct-App19302；Vendor193/RSU606含Unit-Type611=2与Unit-Value612/Value-Digits617 U64=1。这里不含消息头；本地源码SCAP1分支另选Command271/App0，不能把内部19302抄成header App，也不能按Ro272/4解释。Service-Parameter-Info607的11次重复及原顺序仍须完整兼容fixture验证。 总长92（0x5c）。

```avp-hex
00 00 01 04 40 00 00 20
00 00 01 0a 40 00 00 0c 00 00 00 c1
00 00 01 03 40 00 00 0c 00 00 4b 66
00 00 02 5e c0 00 00 3c 00 00 00 c1
00 00 02 63 c0 00 00 10 00 00 00 c1 00 00 00 02
00 00 02 64 c0 00 00 20 00 00 00 c1
00 00 02 69 c0 00 00 14 00 00 00 c1 00 00 00 00 00 00 00 01
```

模板终止MMS组实际行中SMS_KEY1210先于UDL1212，虽然声明名列表顺序相反；成员按后续行构造，不能按声明列表排序“修正”。字符值的首NUL截断、Time的Unix→NTP换算及PID十进制文本等转换是旧业务链证据；新库只按明确协议类型接收owned值，不复制旧S1/reservationId格式或其业务恢复逻辑。

新增PAIR-03/04与TEMPLATE-02见覆盖基线§20。累计候选只完成静态算术/场景登记，协议运行、独立复核、真实六模板及DRA/OCS互操作仍Planned。

## 14. 候选样本按当前实施范围使用（2026-11-04修订）

本次范围修订不修改任何hex、输入值或既有规则结论。首版使用Sh/S6c/SGd/Ro及通用结构的适用候选：包括V-HSS、V-LMSI、V-SRR/SRA、V-OFR/OFA、V-TFR/TFA、V-RDR/RDA、V-ALR/ALA等。共享类型片段只用于实际首版定义，不据此广告另一个应用。

Ro专用V-PORT、V-SMSTATUS、额度/金额/MSCC/CCA/Service-Context、V-PAIR-CCR/CCA、V-SMS-CCR/CCA、V-TERM、V-IEC、V-REFUND、V-FURNISH、V-MMS-ADDR及V-LEGACY系列均纳入首版Ro/兼容候选验收。V-PAIR-NEG、V-CYCLE-NEG等混合反例按应用分开验收；首版必须执行其承诺范围内的Ro fixture并保留SMS-Result标准限制/具名兼容义务。样本所属版本不改变其候选/未验证状态，54段累计hex不是首版支持数量。

## 15. 第二十轮：Sh四组完整成功/失败pair

依据[主流程审计§16.2](diameterkit-mainflow-audit.md)。单DataRef普通Public-Identity子集，全部消息含完整Diameter头和AVP。成功与失败答案是同一请求的**替代结果**，不是允许完成两次。所有Sh请求/正常答案都带VSAI，不能复制S6c的省略VSAI规则。样本未带Proxy-Info/Route-Record；经DRA后本跳HBH、代理信息仍按运行时契约处理。

### V-SH-UDR-01：UDR请求

Command306/App16777217，Session=`as.example;20;1`，HBH=20200001/E2E=30300001；Origin=`as.example`，Destination=`hss.example`，Realm均为`example.net`。DataRef24；读取普通短信注册数据。 总长232（0xe8）。

```message-hex
01 00 00 e8 c0 00 01 32 01 00 00 01 20 20 00 01 30 30 00 01
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 31 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 12 61 73 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 bc c0 00 00 30 00 00 28 af 00 00 02 59 c0 00 00 21 00 00 28 af
73 69 70 3a 61 6c 69 63 65 40 65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00
00 00 02 bf c0 00 00 10 00 00 28 af 00 00 00 18
```

### V-SH-UDA-01：UDA成功

基础2001，Origin为请求目标；VSAI/AuthState仍在，结果与本组请求匹配。User-Data逐字采用X-SH-PULL-01；本例数据已知存在且有读取权限。 总长416（0x1a0）。

```message-hex
01 00 01 a0 40 00 01 32 01 00 00 01 20 20 00 01 30 30 00 01
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 31 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 be c0 00 01 11 00 00 28 af 3c 53 68 2d 44 61 74 61 3e 3c 53 68
2d 49 4d 53 2d 44 61 74 61 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 45 78
74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 53 4d 53 52
65 67 69 73 74 72 61 74 69 6f 6e 49 6e 66 6f 3e 3c 49 50 2d 53 4d 2d 47
57 2d 4e 75 6d 62 65 72 3e 31 32 30 32 35 35 35 30 39 39 39 3c 2f 49 50
2d 53 4d 2d 47 57 2d 4e 75 6d 62 65 72 3e 3c 45 78 74 65 6e 73 69 6f 6e
3e 3c 53 43 41 64 64 72 65 73 73 3e 31 32 30 32 35 35 35 30 31 30 30 3c
2f 53 43 41 64 64 72 65 73 73 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c
2f 53 4d 53 52 65 67 69 73 74 72 61 74 69 6f 6e 49 6e 66 6f 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c
2f 53 68 2d 44 61 74 61 3e 00 00 00
```

### V-SH-UDA-02：UDA失败

替代成功答案，Experimental10415/5102：无读取权限。E=0，基础Result-Code缺席，无User-Data。此处错误原因是显式场景前提，不由库查询权限/用户/订阅数据库。 总长160（0xa0）。

```message-hex
01 00 00 a0 40 00 01 32 01 00 00 01 20 20 00 01 30 30 00 01
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 31 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 2a
40 00 00 0c 00 00 13 ee
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-SH-PUR-01：PUR请求

Command307/App16777217，Session=`as.example;20;2`，HBH=20200002/E2E=30300002；Origin=`as.example`，Destination=`hss.example`，Realm均为`example.net`。DataRef24；User-Data逐字采用X-SH-REG-01，SCAddress不用于更新。 总长448（0x1c0）。

```message-hex
01 00 01 c0 c0 00 01 33 01 00 00 01 20 20 00 02 30 30 00 02
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 32 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 12 61 73 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 bc c0 00 00 30 00 00 28 af 00 00 02 59 c0 00 00 21 00 00 28 af
73 69 70 3a 61 6c 69 63 65 40 65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00
00 00 02 bf c0 00 00 10 00 00 28 af 00 00 00 18
00 00 02 be c0 00 00 d8 00 00 28 af 3c 53 68 2d 44 61 74 61 3e 3c 53 68
2d 49 4d 53 2d 44 61 74 61 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 45 78
74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 53 4d 53 52
65 67 69 73 74 72 61 74 69 6f 6e 49 6e 66 6f 3e 3c 49 50 2d 53 4d 2d 47
57 2d 4e 75 6d 62 65 72 3e 31 32 30 32 35 35 35 30 39 39 39 3c 2f 49 50
2d 53 4d 2d 47 57 2d 4e 75 6d 62 65 72 3e 3c 2f 53 4d 53 52 65 67 69 73
74 72 61 74 69 6f 6e 49 6e 66 6f 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e
3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e
3c 2f 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c 2f 53 68 2d 44 61 74 61 3e
```

### V-SH-PUA-01：PUA成功

基础2001，Origin为请求目标；VSAI/AuthState仍在，结果与本组请求匹配。没有User-Data，不因省略请求payload而判失败。 总长140（0x8c）。

```message-hex
01 00 00 8c 40 00 01 33 01 00 00 01 20 20 00 02 30 30 00 02
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 32 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-SH-PUA-02：PUA失败

替代成功答案，Experimental10415/5103：无修改权限。E=0，基础Result-Code缺席，无User-Data。此处错误原因是显式场景前提，不由库查询权限/用户/订阅数据库。 总长160（0xa0）。

```message-hex
01 00 00 a0 40 00 01 33 01 00 00 01 20 20 00 02 30 30 00 02
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 32 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 2a
40 00 00 0c 00 00 13 ef
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-SH-SNR-01：SNR请求

Command308/App16777217，Session=`as.example;20;3`，HBH=20200003/E2E=30300003；Origin=`as.example`，Destination=`hss.example`，Realm均为`example.net`。DataRef25、SubsReq0、OneTime0，后者M=0；Expiry709的M=0、Time=ee90ff80，era0为2026-11-01 00:00:00 UTC，样本时钟2026-10-31 00:00:00 UTC。没有SendData。 总长280（0x118）。

```message-hex
01 00 01 18 c0 00 01 34 01 00 00 01 20 20 00 03 30 30 00 03
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 33 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 12 61 73 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 bc c0 00 00 30 00 00 28 af 00 00 02 59 c0 00 00 21 00 00 28 af
73 69 70 3a 61 6c 69 63 65 40 65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00
00 00 02 bf c0 00 00 10 00 00 28 af 00 00 00 19
00 00 02 c1 c0 00 00 10 00 00 28 af 00 00 00 00
00 00 02 c5 80 00 00 10 00 00 28 af ee 90 ff 80
00 00 02 c8 80 00 00 10 00 00 28 af 00 00 00 00
```

### V-SH-SNA-01：SNA成功

基础2001，Origin为请求目标；VSAI/AuthState仍在，结果与本组请求匹配。本例返回同一到期时间；更早或省略均按主流程§16解释，不要求User-Data。 总长156（0x9c）。

```message-hex
01 00 00 9c 40 00 01 34 01 00 00 01 20 20 00 03 30 30 00 03
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 33 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 c5 80 00 00 10 00 00 28 af ee 90 ff 80
```

### V-SH-SNA-02：SNA失败

替代成功答案，Experimental10415/5104：无通知权限。E=0，基础Result-Code缺席，无User-Data。此处错误原因是显式场景前提，不由库查询权限/用户/订阅数据库。 总长160（0xa0）。

```message-hex
01 00 00 a0 40 00 01 34 01 00 00 01 20 20 00 03 30 30 00 03
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 33 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 2a
40 00 00 0c 00 00 13 f0
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-SH-PNR-01：PNR请求

Command309/App16777217，Session=`hss.example;20;4`，HBH=20200004/E2E=30300004；Origin=`hss.example`，Destination=`as.example`，Realm均为`example.net`。User-Data逐字采用X-SH-MME-01；无DataRef顶层AVP；与SNR为两个独立exchange，来源HSS需另有binding授权。 总长432（0x1b0）。

```message-hex
01 00 01 b0 c0 00 01 35 01 00 00 01 20 20 00 04 30 30 00 04
00 00 01 07 40 00 00 18 68 73 73 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 34
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 12 61 73 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 bc c0 00 00 30 00 00 28 af 00 00 02 59 c0 00 00 21 00 00 28 af
73 69 70 3a 61 6c 69 63 65 40 65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00
00 00 02 be c0 00 00 d6 00 00 28 af 3c 53 68 2d 44 61 74 61 3e 3c 53 68
2d 49 4d 53 2d 44 61 74 61 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 45 78
74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 55 45 52 65
61 63 68 61 62 69 6c 69 74 79 46 6f 72 49 50 3e 3c 55 45 49 50 52 65 61
63 68 61 62 69 6c 69 74 79 4d 4d 45 3e 30 3c 2f 55 45 49 50 52 65 61 63
68 61 62 69 6c 69 74 79 4d 4d 45 3e 3c 2f 55 45 52 65 61 63 68 61 62 69
6c 69 74 79 46 6f 72 49 50 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f
45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f
53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c 2f 53 68 2d 44 61 74 61 3e 00 00
```

### V-SH-PNA-01：PNA成功

基础2001，Origin为请求目标；VSAI/AuthState仍在，结果与本组请求匹配。没有User-Data，不因省略请求payload而判失败。 总长140（0x8c）。

```message-hex
01 00 00 8c 40 00 01 35 01 00 00 01 20 20 00 04 30 30 00 04
00 00 01 07 40 00 00 18 68 73 73 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 34
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 12 61 73 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-SH-PNA-02：PNA失败

替代成功答案，Experimental10415/5100：AS不认识或不支持所通知的数据。E=0，基础Result-Code缺席，无User-Data。此处错误原因是显式场景前提，不由库查询权限/用户/订阅数据库。 总长160（0xa0）。

```message-hex
01 00 00 a0 40 00 01 35 01 00 00 01 20 20 00 04 30 30 00 04
00 00 01 07 40 00 00 18 68 73 73 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 34
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 2a
40 00 00 0c 00 00 13 ec
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 12 61 73 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

| PAIR-05反例 | 检查层与预期 |
| --- | --- |
| UDA带基础2001又带实验5102；或把5102放基础Result-Code | Sh结果规则判冲突/错误命名空间，不交付typed成功；不回错误答案 |
| 正常Sh答案删VSAI、VSAI AuthApp与header不符、答案Session/HBH/E2E不匹配 | 分别为Wire/Exchange错误；不得因有2001忽略 |
| PUR数据只含SCAddress、漏User-Data或号码元素；UDR改DataRef25；SNR改DataRef24 | 所选短信helper的结构/适用位置失败；无业务数据库要求 |
| SNR25删除OneTime、设置OneTime=1或附SendData | 本固定过程构造失败；OneTime0是显式请求，不是未选择 |
| SNA2001省Expiry或更早Expiry | 合法响应变体，不补写请求时间；按确认语义暴露 |
| PNR删除Destination-Host；将PNR强行关联已完成SNR；DRA Origin冒充HSS | 分别按Wire、独立exchange、binding拒绝错误关联 |
| 合法PNR但本进程无订阅记录 | 允许交给已授权handler；库不自动返回5107或查存储 |
| Expiry/OneTime改M=1 | 标准构造仍发M=0；接收已知AVP按APP-MBIT忽略差异，不能把该变体当未知M1 |

## 16. 第二十轮：Sh-Data固定附件样本与反例

以下XML每块为**UTF-8单行，无BOM、XML声明及结尾换行**；围栏换行不进入User-Data。与报文中的702 payload逐字一致。只选择短信/可达性子集；XML字段名按附件，不从PDF示意拼写推断。XSD静态通过不代表业务权限、订阅、号码分配或端到端协议已经验证。

### X-SH-REG-01：PUR注册；只给号码，不带SCAddress

User-Data为204字节。

```xml
<Sh-Data><Sh-IMS-Data><Extension><Extension><Extension><SMSRegistrationInfo><IP-SM-GW-Number>12025550999</IP-SM-GW-Number></SMSRegistrationInfo></Extension></Extension></Extension></Sh-IMS-Data></Sh-Data>
```

### X-SH-PULL-01：UDA读取号码和SC地址

User-Data为261字节。

```xml
<Sh-Data><Sh-IMS-Data><Extension><Extension><Extension><SMSRegistrationInfo><IP-SM-GW-Number>12025550999</IP-SM-GW-Number><Extension><SCAddress>12025550100</SCAddress></Extension></SMSRegistrationInfo></Extension></Extension></Extension></Sh-IMS-Data></Sh-Data>
```

### X-SH-DEREG-01：PUR注销；必需号码元素显式为空

User-Data为193字节。

```xml
<Sh-Data><Sh-IMS-Data><Extension><Extension><Extension><SMSRegistrationInfo><IP-SM-GW-Number></IP-SM-GW-Number></SMSRegistrationInfo></Extension></Extension></Extension></Sh-IMS-Data></Sh-Data>
```

### X-SH-MME-01：PNR的MME可达性，值0

User-Data为202字节。

```xml
<Sh-Data><Sh-IMS-Data><Extension><Extension><Extension><UEReachabilityForIP><UEIPReachabilityMME>0</UEIPReachabilityMME></UEReachabilityForIP></Extension></Extension></Extension></Sh-IMS-Data></Sh-Data>
```

### X-SH-SGSN-01：SGSN可达性必须在下一层Extension

User-Data为227字节。

```xml
<Sh-Data><Sh-IMS-Data><Extension><Extension><Extension><UEReachabilityForIP><Extension><UEIPReachabilitySGSN>0</UEIPReachabilitySGSN></Extension></UEReachabilityForIP></Extension></Extension></Extension></Sh-IMS-Data></Sh-Data>
```

### X-SH-AMF-01：两个AMF字段按附件顺序出现；仅验证数据表示

User-Data为314字节。

```xml
<Sh-Data><Sh-IMS-Data><Extension><Extension><Extension><UEReachabilityForIP><Extension><Extension><UEIPReachabilityAMF3GPP>0</UEIPReachabilityAMF3GPP><UEIPReachabilityAMFnon3GPP>0</UEIPReachabilityAMFnon3GPP></Extension></Extension></UEReachabilityForIP></Extension></Extension></Extension></Sh-IMS-Data></Sh-Data>
```

### X-SH-SBI-01：采用附件的准确SBI元素名；显式1

User-Data为314字节。

```xml
<Sh-Data><Sh-IMS-Data><Extension><Extension><Extension><SMSRegistrationInfo><IP-SM-GW-Number>12025550999</IP-SM-GW-Number><Extension><Extension><IP-SM-GW-SBISupportIndication>1</IP-SM-GW-SBISupportIndication></Extension></Extension></SMSRegistrationInfo></Extension></Extension></Extension></Sh-IMS-Data></Sh-Data>
```

### X-SH-EXT-01：允许wildcard位置的外部namespace同名元素；标准号码仍是12025550999

User-Data为275字节。

```xml
<Sh-Data><Sh-IMS-Data><Extension><Extension><Extension><SMSRegistrationInfo><IP-SM-GW-Number>12025550999</IP-SM-GW-Number><x:IP-SM-GW-Number xmlns:x="urn:example:ext">opaque</x:IP-SM-GW-Number></SMSRegistrationInfo></Extension></Extension></Extension></Sh-IMS-Data></Sh-Data>
```

| XML-02反例（在上述XML上单项修改） | 静态XSD预期 / 库级语义 |
| --- | --- |
| REG删除IP-SM-GW-Number但保留SMSRegistrationInfo | XSD失败；不能作为注销 |
| REG少一层包裹Extension | XSD失败；不能按任意后代搜索接受 |
| REG同一路径出现两个号码 | XSD失败；不能取first/last覆盖 |
| MME将0改1 | XSD失败；枚举只有0，不按maxInclusive单独放行 |
| REG给Sh-Data加默认namespace `urn:example:wrong` | XSD失败；namespace不能整体剥离 |
| SBI将元素名改成`IP-SM-GW-SBI-SupportIndication` | XSD失败；附件无该别名 |
| PULL的SCAddress改21个ASCII字符 | XSD失败；字符限制不替代内存字节预算 |
| 同层SMSRegistrationInfo放在UEReachabilityForIP之前 | XSD失败；正序且值合法才能通过 |
| `<Sh-Data/>`或合法路径下空UEReachabilityForIP | XSD允许，但没有本例注册/可达语义；helper不能返回“已注册/可达” |
| DEREG号码改为自闭合标签；或添加缩进 | XML语义可能相同，raw bytes不同；重复请求比较不做C14N |
| 外部namespace的同名号码处于允许wildcard位置 | X-SH-EXT-01可过XSD，但扩展不是标准号码；不得覆盖具名字段 |
| DTD/外部实体、超深/超节点/超文本/超扩展预算、恶意schemaLocation | 库级解析器必须有界拒绝/不取件；XSD结果不是安全验收，运行测试仍Planned |

本轮用OpenJDK 25.0.4.1内置Xerces直接装载官方原XSD，离线检查8个正例、前8项非法变体及2个空容器控制项，结果符合表中预期。lxml6.0.2/libxml2 2.15.2直接装载时对两处type QName前导空白报错；仅在临时副本规范化第1000行` tTADSinformationExtension2`、第1039行` tPSUserState`后得到相同结果。原附件/哈希未改，不把工具兼容差异判为官方勘误。raw比较/安全/应用语义边界尚为运行验收要求；没有运行diameterkit或其他Diameter库的Sh解码。

## 17. 第二十轮：MME移动提醒与Gdd MT成对候选

标准ALR继续用S6c，Gdd TFR用SGd/Gdd应用；不能因同一物理Peer共用连接就混用App-ID。字段和角色条件见[主流程§16.4](diameterkit-mainflow-audit.md)。三帧为请求及两个替代答案，不含业务重试状态。

### V-ALR-MOVE-01：旧MME通知已迁移至新MME

S6c Command8388648/App16777312；Session=`mme.example;20;5`，HBH20200005/E2E30300005。Origin=`mme.example`、Destination=`gmsc.example`；SC-Address=`12025550100`是此前TFR的GMSC地址；User-Identifier3102内User-Name=IMSI `001010123456789`。Event3333=2、M=0；Serving-Node指向`mme2.example`/`example.net`/`12025550888`，MME三叶使用S6c的M=1。外部前提是旧MME已有此次提醒所需事实；库不保存TFR调度状态。 总长288（0x120）。

```message-hex
01 00 01 20 c0 80 00 28 01 00 00 60 20 20 00 05 30 30 00 05
00 00 01 07 40 00 00 18 6d 6d 65 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 35
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 6d 6d 65 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 14 67 6d 73 63 2e 65 78 61 6d 70 6c 65
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c 1e c0 00 00 24 00 00 28 af 00 00 00 01 40 00 00 17 30 30 31 30
31 30 31 32 33 34 35 36 37 38 39 00
00 00 0d 05 80 00 00 10 00 00 28 af 00 00 00 02
00 00 09 61 c0 00 00 50 00 00 28 af 00 00 09 62 c0 00 00 18 00 00 28 af
6d 6d 65 32 2e 65 78 61 6d 70 6c 65 00 00 09 68 c0 00 00 17 00 00 28 af
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 06 6d c0 00 00 12 00 00 28 af
21 20 55 05 88 f8 00 00
```

### V-ALA-MOVE-01：移动提醒成功

匹配ALR-MOVE-01，Origin=`gmsc.example`，基础2001；不表示已重投递SMS或已更新路由数据库。 总长108（0x6c）。

```message-hex
01 00 00 6c 40 80 00 28 01 00 00 60 20 20 00 05 30 30 00 05
00 00 01 07 40 00 00 18 6d 6d 65 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 35
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 14 67 6d 73 63 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-ALA-MOVE-02：移动提醒基础忙错误

同请求的替代答案，基础3004、E=1/P=1；按基础错误语法，仅Session/Result/Origin，无VSAI、AuthState或实验5555。库不据此自动安排短信重投递。 总长96（0x60）。

```message-hex
01 00 00 60 60 80 00 28 01 00 00 60 20 20 00 05 30 30 00 05
00 00 01 07 40 00 00 18 6d 6d 65 2e 65 78 61 6d 70 6c 65 3b 32 30 3b 35
00 00 01 0c 40 00 00 0c 00 00 0b bc
00 00 01 08 40 00 00 14 67 6d 73 63 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-GDD-TFR-01：向SGSN发送普通MT

Command8388646/App16777313，与SGd同一应用。Session=`gmsc.example;20;6`，HBH20200006/E2E30300006；Origin=`gmsc.example`，必需Destination-Host=`sgsn.example`。IMSI `001010123456789`，SC地址`12025550100`，SGSN-Number1489=`12025550777`、V=1/M=0；无MME-Number。TPDU、Timer30秒及Start-Time沿用V-TFR-01，固定样本时钟仍为2026-01-30 12:34:56 UTC。部署授权SGSN且明确路由号码，不由DRA CER推导。 总长268（0x10c）。

```message-hex
01 00 01 0c c0 80 00 26 01 00 00 61 20 20 00 06 30 30 00 06
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 32 30 3b
36 00 00 00
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 14 67 6d 73 63 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 14 73 67 73 6e 2e 65 78 61 6d 70 6c 65
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 00 01 40 00 00 17 30 30 31 30 31 30 31 32 33 34 35 36 37 38 39 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c e5 c0 00 00 20 00 00 28 af 04 0b 91 21 20 55 05 21 f3 00 04 62
10 03 21 43 65 00 01 41
00 00 05 d1 80 00 00 12 00 00 28 af 21 20 55 05 77 f7 00 00
00 00 0c ea c0 00 00 10 00 00 28 af 00 00 00 1e
00 00 0c eb c0 00 00 10 00 00 28 af ed 27 23 f0
```

### V-GDD-TFA-01：Gdd MT成功

匹配GDD-TFR-01，Origin=`sgsn.example`，基础2001；无需可选User-Identifier或TPDU回显。 总长112（0x70）。

```message-hex
01 00 00 70 40 80 00 26 01 00 00 61 20 20 00 06 30 30 00 06
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 32 30 3b
36 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 14 73 67 73 6e 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### V-GDD-TFA-02：Gdd MT容量不足

同请求的替代答案，正常E=0、实验10415/5555；SM-Delivery-Failure-Cause3303内3304=0。原因0存在，不能当缺值；本例没有Requested-Retransmission-Time。 总长160（0xa0）。

```message-hex
01 00 00 a0 40 80 00 26 01 00 00 61 20 20 00 06 30 30 00 06
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 32 30 3b
36 00 00 00
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 2a
40 00 00 0c 00 00 15 b3
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 14 73 67 73 6e 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e7 c0 00 00 1c 00 00 28 af 00 00 0c e8 c0 00 00 10 00 00 28 af
00 00 00 00
```

| PAIR-06 / FIELD-13反例 | 预期 |
| --- | --- |
| ALR-MOVE将App改16777313 | 标准Profile拒绝；旧SGd Alert仅在具名兼容binding中另验 |
| 移动事件删Serving-Node，并明确新节点未知 | 合法条件变体；不自动SRR或选择DRA当新节点 |
| 移动事件已知新节点却省字段，或MME节点缺三叶之一 | strict已选角色条件失败；不得靠补默认节点通过 |
| Event=1、2、3，或0/仅未知位 | 前三者保留已知位语义；后两者不满足本例已知事件，raw仍保留；不把3判非法Enum |
| Alert的Origin与Serving-Node不同 | 本例合法；旧节点报告新节点，不强制相等 |
| ALR无Event或直接把User-Name放顶层、User-Identifier缺IMSI | MME/SGSN分支必需条件/父路径失败；不能套HSS普通MSISDN分支 |
| Gdd TFR把1489换成MME1645、删Destination-Host；接收1489的M=1 | 前两项违反已选发送角色/包络；M差异对已知字段按应用接收规则忽略 |
| 给TFR加Maximum-Retransmission-Time却不带SMS-GMSC-Address | 对应条件缺字段；库不生成地址或业务重传计划 |
| ALR/ALA错误自动改成实验5555，或TFA5555缺原因 | 分别不符合提醒过程结果/MT失败条件；不能靠顶层失败跳过所有字段校验 |

累计72段hex（旧54+本轮18）；当前四应用首版归属按§14及Specification§3处理，不以数量宣布支持完成。字节边界、flags、配对和XML附件检查只是静态证据；PAIR-05/06、XML-02、FIELD-13的实现/独立对端状态均Planned。

## 18. 第二十一轮：号码与订阅的增量证据

来源和规则见[主流程§17](diameterkit-mainflow-audit.md)。下列8组各给请求、成功和失败替代答案，共24帧。与旧候选相同，静态排字节不等于独立Diameter实现验证。省略的外部事实均在场景中明确；库不读取任何业务存储。

**既有样本勘误：** V-SRR-01曾把SM-RP-MTI写成2400（`00 00 09 60`），该Code实际为LMSI。已改为3308（`00 00 0c ec`），总长、值0、Vendor及其余字段不变。旧值只保留为本条错误记录，不再作为正向fixture；以往长度算术通过并不能证明Code语义正确。

号码表示例见主流程§17.1。dummy MSISDN严格为15个0，TBCD为8字节`00 00 00 00 00 00 00 f0`。不能把它作为普通注册IP-SM-GW号码或当作真实用户路由目标。

### X-SH-CONT-01：单数据持续订阅的当前值

UTF-8单行，无BOM/声明/结尾换行；IMSUserState1=REGISTERED，与可达0无关。

```xml
<Sh-Data><Sh-IMS-Data><IMSUserState>1</IMSUserState></Sh-IMS-Data></Sh-Data>
```

### X-SH-MULTI-01：Notif-Eff订阅返回两项当前值

XML中SCSCFName在IMSUserState之前，二者是Sh-IMS-Data直接子项。UTF-8单行，无结尾换行。

```xml
<Sh-Data><Sh-IMS-Data><SCSCFName>sip:scscf.example.net</SCSCFName><IMSUserState>1</IMSUserState></Sh-IMS-Data></Sh-Data>
```

### UNSUB：Sh一次可达订阅的退订

DataRef25、SubsReq1、OneTime0/M0；Public-Identity访问键，顶层User-Name为Private Identity；不带Expiry或SendData。成功也覆盖HSS已无该订阅；失败实验5002表示HSS判断公共/私有身份不匹配，并非库预先查询所得。

Session=`as.example;21;1`；HBH=21210001、E2E=31310001，仅供隔离fixture，不是生产发号策略。

**V21-UNSUB-R / 请求**：总长304（0x130），Command308/App16777217。

```message-hex
01 00 01 30 c0 00 01 34 01 00 00 01 21 21 00 01 31 31 00 01
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 31 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 12 61 73 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 bc c0 00 00 30 00 00 28 af 00 00 02 59 c0 00 00 21 00 00 28 af
73 69 70 3a 61 6c 69 63 65 40 65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00
00 00 00 01 40 00 00 27 30 30 31 30 31 30 31 32 33 34 35 36 37 38 39 40
69 6d 73 2e 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 bf c0 00 00 10 00 00 28 af 00 00 00 19
00 00 02 c1 c0 00 00 10 00 00 28 af 00 00 00 01
00 00 02 c8 80 00 00 10 00 00 28 af 00 00 00 00
```

**V21-UNSUB-A / 成功答案**：总长140（0x8c），Command308/App16777217。

```message-hex
01 00 00 8c 40 00 01 34 01 00 00 01 21 21 00 01 31 31 00 01
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 31 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

**V21-UNSUB-F / 失败替代答案**：总长160（0xa0），Command308/App16777217。

```message-hex
01 00 00 a0 40 00 01 34 01 00 00 01 21 21 00 01 31 31 00 01
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 31 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 2a
40 00 00 0c 00 00 13 8a
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### SUBMULTI：Sh持续订阅两种数据并返回当前值

DataRef依次11、12；SubsReq0、SendData1/M0、无OneTime。Supported-Features628/V1/M0内Vendor10415、FeatureListID1、FeatureList1。场景前提：本地和目标HSS均已明确支持Notif-Eff，不能由DRA元数据推断。成功User-Data为X-SH-MULTI-01；失败实验5104表示无相应通知权限，无部分成功数据；HSS回滚既有状态不是库责任。省Expiry表示未设有限到期，不启动续订定时器。

Session=`as.example;21;2`；HBH=21210002、E2E=31310002，仅供隔离fixture，不是生产发号策略。

**V21-SUBMULTI-R / 请求**：总长336（0x150），Command308/App16777217。

```message-hex
01 00 01 50 c0 00 01 34 01 00 00 01 21 21 00 02 31 31 00 02
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 32 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 12 61 73 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 74 80 00 00 38 00 00 28 af 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 02 75 80 00 00 10 00 00 28 af 00 00 00 01 00 00 02 76 80 00 00 10
00 00 28 af 00 00 00 01
00 00 02 bc c0 00 00 30 00 00 28 af 00 00 02 59 c0 00 00 21 00 00 28 af
73 69 70 3a 61 6c 69 63 65 40 65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00
00 00 02 bf c0 00 00 10 00 00 28 af 00 00 00 0b
00 00 02 bf c0 00 00 10 00 00 28 af 00 00 00 0c
00 00 02 c1 c0 00 00 10 00 00 28 af 00 00 00 00
00 00 02 c6 80 00 00 10 00 00 28 af 00 00 00 01
```

**V21-SUBMULTI-A / 成功答案**：总长328（0x148），Command308/App16777217。

```message-hex
01 00 01 48 40 00 01 34 01 00 00 01 21 21 00 02 31 31 00 02
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 32 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 74 80 00 00 38 00 00 28 af 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 02 75 80 00 00 10 00 00 28 af 00 00 00 01 00 00 02 76 80 00 00 10
00 00 28 af 00 00 00 01
00 00 02 be c0 00 00 84 00 00 28 af 3c 53 68 2d 44 61 74 61 3e 3c 53 68
2d 49 4d 53 2d 44 61 74 61 3e 3c 53 43 53 43 46 4e 61 6d 65 3e 73 69 70
3a 73 63 73 63 66 2e 65 78 61 6d 70 6c 65 2e 6e 65 74 3c 2f 53 43 53 43
46 4e 61 6d 65 3e 3c 49 4d 53 55 73 65 72 53 74 61 74 65 3e 31 3c 2f 49
4d 53 55 73 65 72 53 74 61 74 65 3e 3c 2f 53 68 2d 49 4d 53 2d 44 61 74
61 3e 3c 2f 53 68 2d 44 61 74 61 3e
```

**V21-SUBMULTI-F / 失败替代答案**：总长160（0xa0），Command308/App16777217。

```message-hex
01 00 00 a0 40 00 01 34 01 00 00 01 21 21 00 02 31 31 00 02
00 00 01 07 40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 32 00
00 00 01 04 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 02
40 00 00 0c 01 00 00 01
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 2a
40 00 00 0c 00 00 13 f0
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

## 19. 第二十一轮：特殊身份与提醒完整pair

所有SMS候选的标准发送均无VSAI；Sh正常包络保留VSAI。Realm统一`example.net`。IMS各身份的来源与命令位置不能相互替代。

### IMS-SRR：IMS无MSISDN的路由查询

请求顶层User-Name为HSS-ID，不带MSISDN；SMSMI只有目的URI；SC-Address和SM-RP-MTI3308=0。成功场景明确HSS知道目的用户IMSI，答案User-Name为该IMSI、Serving-Node2401只含IP-SM-GW-Number3100=`12025550999`。不带Name/Realm以避开未决组合；若应用随后使用Diameter TFR，目标isg.example另由已授权部署配置提供，不能从号码猜host。失败实验5001不要求成功节点/IMSI。

Session=`gmsc.example;21;3`；HBH=21210003、E2E=31310003，仅供隔离fixture，不是生产发号策略。

**V21-IMS-SRR-R / 请求**：总长244（0xf4），Command8388647/App16777312。

```message-hex
01 00 00 f4 c0 80 00 27 01 00 00 60 21 21 00 03 31 31 00 03
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 32 31 3b
33 00 00 00
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 14 67 6d 73 63 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 00 01 40 00 00 17 30 30 31 30 31 30 30 30 30 30 30 30 30 30 31 00
00 00 0c fc 80 00 00 2c 00 00 28 af 00 00 0c ff 80 00 00 1f 00 00 28 af
73 69 70 3a 62 6f 62 40 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c ec c0 00 00 10 00 00 28 af 00 00 00 00
```

**V21-IMS-SRR-A / 成功答案**：总长168（0xa8），Command8388647/App16777312。

```message-hex
01 00 00 a8 40 80 00 27 01 00 00 60 21 21 00 03 31 31 00 03
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 32 31 3b
33 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 00 01 40 00 00 17 30 30 31 30 31 30 31 32 33 34 35 36 37 38 39 00
00 00 09 61 c0 00 00 20 00 00 28 af 00 00 0c 1c c0 00 00 12 00 00 28 af
21 20 55 05 99 f9 00 00
```

**V21-IMS-SRR-F / 失败替代答案**：总长132（0x84），Command8388647/App16777312。

```message-hex
01 00 00 84 40 80 00 27 01 00 00 60 21 21 00 03 31 31 00 03
00 00 01 07 40 00 00 19 67 6d 73 63 2e 65 78 61 6d 70 6c 65 3b 32 31 3b
33 00 00 00
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 2a
40 00 00 0c 00 00 13 89
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### IMS-RDR：IP-SM-GW报告IMS用户投递失败

User-Identifier3102内User-Name为HSS-ID；SMSMI仅目的URI。Outcome3316内IP-SM-GW-Outcome3320/Cause3321=0，明确容量不足，无需AbsentUser诊断。成功RDA不要求Alert号码；失败实验5558为MWD列表满。无VSAI、无RDR-Flags；不由此决定重试。

Session=`isg.example;21;4`；HBH=21210004、E2E=31310004，仅供隔离fixture，不是生产发号策略。

**V21-IMS-RDR-R / 请求**：总长276（0x114），Command8388649/App16777312。

```message-hex
01 00 01 14 c0 80 00 29 01 00 00 60 21 21 00 04 31 31 00 04
00 00 01 07 40 00 00 18 69 73 67 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 34
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 69 73 67 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c 1e c0 00 00 24 00 00 28 af 00 00 00 01 40 00 00 17 30 30 31 30
31 30 30 30 30 30 30 30 30 30 31 00
00 00 0c fc 80 00 00 2c 00 00 28 af 00 00 0c ff 80 00 00 1f 00 00 28 af
73 69 70 3a 62 6f 62 40 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c f4 c0 00 00 28 00 00 28 af 00 00 0c f8 c0 00 00 1c 00 00 28 af
00 00 0c f9 c0 00 00 10 00 00 28 af 00 00 00 00
```

**V21-IMS-RDR-A / 成功答案**：总长108（0x6c），Command8388649/App16777312。

```message-hex
01 00 00 6c 40 80 00 29 01 00 00 60 21 21 00 04 31 31 00 04
00 00 01 07 40 00 00 18 69 73 67 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 34
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

**V21-IMS-RDR-F / 失败替代答案**：总长128（0x80），Command8388649/App16777312。

```message-hex
01 00 00 80 40 80 00 29 01 00 00 60 21 21 00 04 31 31 00 04
00 00 01 07 40 00 00 18 69 73 67 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 34
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 2a
40 00 00 0c 00 00 15 b6
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### IMS-ALR：HSS提醒IMS无MSISDN的目的用户

User-Identifier内MSISDN为15个0；SMSMI只有目的URI，没有HSS-ID或源URI，无MME/SGSN Event。SC地址来自应用已知事实；MaxUEAvailability本例未知而省略。2001不表示已重新投递；失败基础E3004，无AuthState。

Session=`hss.example;21;5`；HBH=21210005、E2E=31310005，仅供隔离fixture，不是生产发号策略。

**V21-IMS-ALR-R / 请求**：总长232（0xe8），Command8388648/App16777312。

```message-hex
01 00 00 e8 c0 80 00 28 01 00 00 60 21 21 00 05 31 31 00 05
00 00 01 07 40 00 00 18 68 73 73 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 35
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 12 69 77 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c 1e c0 00 00 20 00 00 28 af 00 00 02 bd c0 00 00 14 00 00 28 af
00 00 00 00 00 00 00 f0
00 00 0c fc 80 00 00 2c 00 00 28 af 00 00 0c ff 80 00 00 1f 00 00 28 af
73 69 70 3a 62 6f 62 40 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

**V21-IMS-ALR-A / 成功答案**：总长108（0x6c），Command8388648/App16777312。

```message-hex
01 00 00 6c 40 80 00 28 01 00 00 60 21 21 00 05 31 31 00 05
00 00 01 07 40 00 00 18 68 73 73 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 35
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 12 69 77 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

**V21-IMS-ALR-F / 失败替代答案**：总长96（0x60），Command8388648/App16777312。

```message-hex
01 00 00 60 60 80 00 28 01 00 00 60 21 21 00 05 31 31 00 05
00 00 01 07 40 00 00 18 68 73 73 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 35
00 00 01 0c 40 00 00 0c 00 00 0b bc
00 00 01 08 40 00 00 12 69 77 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### IMS-OFR：IP-SM-GW将直接投递失败的SMS交给IWMSC

源用户IMSI/MSISDN均不可用，User-Identifier只含dummy MSISDN；SMSMI有HSS-ID及源/目的URI。Outcome包含IP-SM-GW/Cause0，来自此前目的网络返回的信息，不能用OFA结果倒填。SM-RP-UI沿用V-OFR-01的人工MO TPDU，仅有界承载，不假装TPDU内号码就是IMS SIP身份。成功2001，失败实验5555且Failure-Cause=4（SC拥塞）。

Session=`isg.example;21;6`；HBH=21210006、E2E=31310006，仅供隔离fixture，不是生产发号策略。

**V21-IMS-OFR-R / 请求**：总长364（0x16c），Command8388645/App16777313。

```message-hex
01 00 01 6c c0 80 00 25 01 00 00 61 21 21 00 06 31 31 00 06
00 00 01 07 40 00 00 18 69 73 67 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 36
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 69 73 67 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 12 69 77 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c e5 c0 00 00 1a 00 00 28 af 01 00 0b 91 21 20 55 05 21 f3 00 04
01 41 00 00
00 00 0c 1e c0 00 00 20 00 00 28 af 00 00 02 bd c0 00 00 14 00 00 28 af
00 00 00 00 00 00 00 f0
00 00 0c fc 80 00 00 6c 00 00 28 af 00 00 0c fd 80 00 00 1b 00 00 28 af
30 30 31 30 31 30 30 30 30 30 30 30 30 30 31 00 00 00 0c fe 80 00 00 21
00 00 28 af 73 69 70 3a 61 6c 69 63 65 40 65 78 61 6d 70 6c 65 2e 6e 65
74 00 00 00 00 00 0c ff 80 00 00 1f 00 00 28 af 73 69 70 3a 62 6f 62 40
65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c f4 c0 00 00 28 00 00 28 af 00 00 0c f8 c0 00 00 1c 00 00 28 af
00 00 0c f9 c0 00 00 10 00 00 28 af 00 00 00 00
```

**V21-IMS-OFR-A / 成功答案**：总长108（0x6c），Command8388645/App16777313。

```message-hex
01 00 00 6c 40 80 00 25 01 00 00 61 21 21 00 06 31 31 00 06
00 00 01 07 40 00 00 18 69 73 67 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 36
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 12 69 77 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

**V21-IMS-OFR-F / 失败替代答案**：总长156（0x9c），Command8388645/App16777313。

```message-hex
01 00 00 9c 40 80 00 25 01 00 00 61 21 21 00 06 31 31 00 06
00 00 01 07 40 00 00 18 69 73 67 2e 65 78 61 6d 70 6c 65 3b 32 31 3b 36
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 2a
40 00 00 0c 00 00 15 b3
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 12 69 77 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e7 c0 00 00 1c 00 00 28 af 00 00 0c e8 c0 00 00 10 00 00 28 af
00 00 00 04
```

### SGSN-ALR：SGSN向Router报告移动

Origin=sgsn.example，目标router.example，User-Identifier/User-Name为IMSI，SC=`12025550100`是本例此前TFR提供的Router地址。Event3333=2/M0；新节点sgsn2.example、realm example.net、号码12025550777。SGSNName/Realm M0，S6c号码1489 M1。成功2001或基础E3004；两答案互斥。

Session=`sgsn.example;21;7`；HBH=21210007、E2E=31310007，仅供隔离fixture，不是生产发号策略。

**V21-SGSN-ALR-R / 请求**：总长300（0x12c），Command8388648/App16777312。

```message-hex
01 00 01 2c c0 80 00 28 01 00 00 60 21 21 00 07 31 31 00 07
00 00 01 07 40 00 00 19 73 67 73 6e 2e 65 78 61 6d 70 6c 65 3b 32 31 3b
37 00 00 00
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 14 73 67 73 6e 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 16 72 6f 75 74 65 72 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c 1e c0 00 00 24 00 00 28 af 00 00 00 01 40 00 00 17 30 30 31 30
31 30 31 32 33 34 35 36 37 38 39 00
00 00 0d 05 80 00 00 10 00 00 28 af 00 00 00 02
00 00 09 61 c0 00 00 54 00 00 28 af 00 00 09 69 80 00 00 19 00 00 28 af
73 67 73 6e 32 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 09 6a 80 00 00 17
00 00 28 af 65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 05 d1 c0 00 00 12
00 00 28 af 21 20 55 05 77 f7 00 00
```

**V21-SGSN-ALR-A / 成功答案**：总长116（0x74），Command8388648/App16777312。

```message-hex
01 00 00 74 40 80 00 28 01 00 00 60 21 21 00 07 31 31 00 07
00 00 01 07 40 00 00 19 73 67 73 6e 2e 65 78 61 6d 70 6c 65 3b 32 31 3b
37 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 16 72 6f 75 74 65 72 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

**V21-SGSN-ALR-F / 失败替代答案**：总长104（0x68），Command8388648/App16777312。

```message-hex
01 00 00 68 60 80 00 28 01 00 00 60 21 21 00 07 31 31 00 07
00 00 01 07 40 00 00 19 73 67 73 6e 2e 65 78 61 6d 70 6c 65 3b 32 31 3b
37 00 00 00
00 00 01 0c 40 00 00 0c 00 00 0b bc
00 00 01 08 40 00 00 16 72 6f 75 74 65 72 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### ROUTER-ALR：Router以此前给GMSC的标识通知

这是调用方在Router角色显式创建的新endpoint请求；Origin=router.example，目标gmsc.example，User-Name为opaque `router-alert-21`。外部前提：Router此前TFA已向GMSC给出该User Identifier Alert。不是复制SGSN的一跳报文，不按IMSI规则校验token；库不建立映射。SC为本例已知GMSC地址，Event/新SGSN仍保留。

Session=`router.example;21;8`；HBH=21210008、E2E=31310008，仅供隔离fixture，不是生产发号策略。

**V21-ROUTER-ALR-R / 请求**：总长300（0x12c），Command8388648/App16777312。

```message-hex
01 00 01 2c c0 80 00 28 01 00 00 60 21 21 00 08 31 31 00 08
00 00 01 07 40 00 00 1b 72 6f 75 74 65 72 2e 65 78 61 6d 70 6c 65 3b 32
31 3b 38 00
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 16 72 6f 75 74 65 72 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 14 67 6d 73 63 2e 65 78 61 6d 70 6c 65
00 00 01 1b 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e4 c0 00 00 12 00 00 28 af 21 20 55 05 01 f0 00 00
00 00 0c 1e c0 00 00 24 00 00 28 af 00 00 00 01 40 00 00 17 72 6f 75 74
65 72 2d 61 6c 65 72 74 2d 32 31 00
00 00 0d 05 80 00 00 10 00 00 28 af 00 00 00 02
00 00 09 61 c0 00 00 54 00 00 28 af 00 00 09 69 80 00 00 19 00 00 28 af
73 67 73 6e 32 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 09 6a 80 00 00 17
00 00 28 af 65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 05 d1 c0 00 00 12
00 00 28 af 21 20 55 05 77 f7 00 00
```

**V21-ROUTER-ALR-A / 成功答案**：总长112（0x70），Command8388648/App16777312。

```message-hex
01 00 00 70 40 80 00 28 01 00 00 60 21 21 00 08 31 31 00 08
00 00 01 07 40 00 00 1b 72 6f 75 74 65 72 2e 65 78 61 6d 70 6c 65 3b 32
31 3b 38 00
00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 14 67 6d 73 63 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

**V21-ROUTER-ALR-F / 失败替代答案**：总长100（0x64），Command8388648/App16777312。

```message-hex
01 00 00 64 60 80 00 28 01 00 00 60 21 21 00 08 31 31 00 08
00 00 01 07 40 00 00 1b 72 6f 75 74 65 72 2e 65 78 61 6d 70 6c 65 3b 32
31 3b 38 00
00 00 01 0c 40 00 00 0c 00 00 0b bc
00 00 01 08 40 00 00 14 67 6d 73 63 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

## 20. 第二十一轮：具名派生分支与负例

为避免重复整段AVP，以下是可重建的**完整消息派生清单**。取指定基帧，按表替换/删除整AVP，不存在的删除项无操作；未列AVP、顺序、Vendor及flags全部保留。递归重算AVP Length、padding和Message Length，Session替换后不复用原padding；每组三帧统一新HBH/E2E，R/A/F分别为请求/成功/失败替代答案。表中长度和SHA-256是派生后整帧，不是基帧摘要。它们仅有静态重建证据，没有增加本文hex块数量或独立互操作覆盖。

公共值：IMSI=`001010123456789`，HSS-ID=`001010000000001`，源URI=`sip:alice@example.net`，目的URI=`sip:bob@example.net`；Realm不变。T4身份样本限定触发后已有短信协议事实，不实现T4 Submit Trigger。IMS-TFR目标IP-SM-GW必须已有独立授权binding。

### D21-SUB11

Session改as.example;21;9；删除628及DataRef12；若有User-Data替换为X-SH-CONT-01；其他不变。 HBH=21210009/E2E=31310009。

| 派生ID | 基帧 | 总长 | 整帧SHA-256 |
| --- | --- | --- | --- |
| D21-SUB11-R | V21-SUBMULTI-R | 264 | `4400e71fc2df6f6440a9a5e50019724ea7fd8952a7ee49e5f6bc630aa225fbe3` |
| D21-SUB11-A | V21-SUBMULTI-A | 228 | `ca43fcbc1b3c5d4cda626b830f16ae59a8f538a99fb3dfc8d039bd0feee14946` |
| D21-SUB11-F | V21-SUBMULTI-F | 160 | `99e578fb16f62b99ac16362b39fd89ed958b95204f7313b30e9903563137e0c9` |

### D21-T4-SRR

Session改gmsc.example;21;10；全部删SMSMI3324；请求身份改IMSI（SRR顶层1，其余3102/1）；T4-SRR成功Serving-Node改MME三叶2402=mme.example、2408=example.net、1645=TBCD12025550999/M1；其余字段不变。 HBH=2121000a/E2E=3131000a。

| 派生ID | 基帧 | 总长 | 整帧SHA-256 |
| --- | --- | --- | --- |
| D21-T4-SRR-R | V21-IMS-SRR-R | 200 | `a3022d13c463b0e402f03812397d3f3b8f50f6397df6ab1825705911a545e04b` |
| D21-T4-SRR-A | V21-IMS-SRR-A | 216 | `12d46cdb7e8a5603b4bef5677ff7e80026ae90612ca67e8b02a6cd7266c18c2b` |
| D21-T4-SRR-F | V21-IMS-SRR-F | 132 | `49d463e6387069e251e26112d5f527268988ebe7eff179eb30feb252bbbab87a` |

### D21-T4-RDR

Session改isg.example;21;11；全部删SMSMI3324；请求身份改IMSI（SRR顶层1，其余3102/1）；T4-SRR成功Serving-Node改MME三叶2402=mme.example、2408=example.net、1645=TBCD12025550999/M1；其余字段不变。 HBH=2121000b/E2E=3131000b。

| 派生ID | 基帧 | 总长 | 整帧SHA-256 |
| --- | --- | --- | --- |
| D21-T4-RDR-R | V21-IMS-RDR-R | 236 | `6f8dbc2007dcd66ab3752237436646aed6293fa9ddfbbd7b0e1246dd8b6a44ee` |
| D21-T4-RDR-A | V21-IMS-RDR-A | 112 | `d0fac404f95462c1a781a70b18dbf1816f6ca96e4c94e5265aae5c388e21d2f8` |
| D21-T4-RDR-F | V21-IMS-RDR-F | 132 | `69ce06fca38d9625c7291a00f30b932ed3cc7c4416fd0279159348e26647c4cd` |

### D21-T4-ALR

Session改hss.example;21;12；全部删SMSMI3324；请求身份改IMSI（SRR顶层1，其余3102/1）；T4-SRR成功Serving-Node改MME三叶2402=mme.example、2408=example.net、1645=TBCD12025550999/M1；其余字段不变。 HBH=2121000c/E2E=3131000c。

| 派生ID | 基帧 | 总长 | 整帧SHA-256 |
| --- | --- | --- | --- |
| D21-T4-ALR-R | V21-IMS-ALR-R | 196 | `933eaa9398bd56129afd66942dfc46834e6e4f61e1b090780c52498dffd38d00` |
| D21-T4-ALR-A | V21-IMS-ALR-A | 112 | `9faa9ed255223754acd605fc0caedd82c27ead3cc07831a360fa241f9429f7e6` |
| D21-T4-ALR-F | V21-IMS-ALR-F | 100 | `c565ec3d85bded2f9b38057154aa1ee814e765492dfb46d8e09d543d5e819c40` |

### D21-GDD-OFR

Session改sgsn.example;21;13；请求Origin改sgsn.example、Destination改iw.example，User-Identifier改IMSI+MSISDN12025550123；答案Origin改iw.example；删除可选EPS-Location-Information1496，其余不变。 HBH=2121000d/E2E=3131000d。

| 派生ID | 基帧 | 总长 | 整帧SHA-256 |
| --- | --- | --- | --- |
| D21-GDD-OFR-R | V-OFR-01 | 244 | `dd0d26d311f59cc68673c442ed5831cbaa41306487b4414d01a0c5dbb828a735` |
| D21-GDD-OFR-A | V-OFA-01 | 112 | `13eae0026770d2bfad6838e40c639c3629adae5e6227e993a94a399703c1e838` |
| D21-GDD-OFR-F | V-OFA-02 | 160 | `854fd01420907082ca5a9d0f70fd5762b063d83c95bebdeff84ca96f4b6a3f91` |

### D21-IMS-TFR

Session改gmsc.example;21;14；请求Destination改isg.example、顶层User-Name改HSS-ID、删1645，末尾加入SMSMI3324(V1/M0)，依次3326=源URI、3327=目的URI（均V1/M0）；答案Origin改isg.example；其余不变，固定时钟仍沿用V-TFR-01。 HBH=2121000e/E2E=3131000e。

| 派生ID | 基帧 | 总长 | 整帧SHA-256 |
| --- | --- | --- | --- |
| D21-IMS-TFR-R | V-TFR-01 | 328 | `eefd0e5a0df30e7027526e418891017a678217a37e0d0ddde680d825dcd45590` |
| D21-IMS-TFR-A | V-TFA-01 | 112 | `2fa5a486ba79370e5212aa80977e8feb1b4251abafc3885e9c50f94ecd87165b` |
| D21-IMS-TFR-F | V-TFA-02 | 160 | `d58cfd920273ea03bda0b726203e7bcf66a9a219a72c6973cc8f03116afc4523` |

### D21-LEGACY-ALR

Session改router.example;21;15，header App改16777313；字段/flags不变；仅legacy-sgd-alert-v1解释域可用，答案保持该App。 HBH=2121000f/E2E=3131000f。

| 派生ID | 基帧 | 总长 | 整帧SHA-256 |
| --- | --- | --- | --- |
| D21-LEGACY-ALR-R | V21-ROUTER-ALR-R | 300 | `70fc7f7af0119a2ec1aed54acfd95628fb1a4917844eacc4710c8723905399e3` |
| D21-LEGACY-ALR-A | V21-ROUTER-ALR-A | 112 | `6583446ee3bb599301e25f729697fc331ab245b7d992da04962e832453618edf` |
| D21-LEGACY-ALR-F | V21-ROUTER-ALR-F | 100 | `a95ede7d7a4668b0e1ede874b51bb9eb34503ceee6967b1a9804962be7303d5e` |

| NUMBER-01 / SUB-01 / IDENTITY-01 / ALERT-02反例 | 预期与层次 |
| --- | --- |
| Sh请求User-Identity同时放Public-Identity和MSISDN；Private Identity嵌入700 | 已选访问键/位置检查失败；不能为兼容旧构造自动选择其中一个；SMS User-Identifier的IMSI+MSISDN不适用此互斥规则 |
| UNSUB把705=1改0；以Expiry=0代替退订 | 前者变成订阅请求，后者不是退订；禁止helper静默改意图 |
| 多DataRef无终端Notif-Eff依据、仅DRA广告或HSS明确不支持 | 分别报告具名特性输入缺失、身份来源错误或特性不支持；不伪造已协商 |
| 仅持续Ref11带OneTime；一次Ref25带SendData；Ref11带不适用Private Identity | 当前具体过程/使用位置失败；与raw字段可以解码分开 |
| SNA2001无User-Data、SNA失败含错误码 | 数据缺失不得一律升级为协议错误；按SHOULD和实际结果解释，不自动重订阅 |
| IMS-SRR/RDR的SMSMI填HSS-ID或源URI；TFR复制OFR三元组 | 命令条件禁止项，不按相同Grouped定义全放行 |
| IMS-ALR把dummy换HSS-ID、删目的URI；OFR源身份错用目的HSS-ID | 不能解释为所选流程的有效身份，不通过查询业务库“补齐”当前消息 |
| IMS-RDR的IP-SM-GW cause放入MME节点；OFR缺3316 | 角色/条件失败；合法枚举值不证明父路径正确 |
| SGSN移动Alert缺已知新节点号码；Router token按IMSI拒绝 | 前者违背已选已知节点条件；后者是错误的全局身份校验。新节点确实未知时允许省略整组 |
| 旧SGd Alert在标准域启用，或成功ALA换回S6c App | 分派/关联失败；兼容域不能自动降级或改写header |
| T4/IMS/RDR成功后自动安排ALR或TFR | 超出公开库边界；每个后继请求由应用显式决定，不能用协议2001推导业务完成 |

本轮直接新增24段message-hex，累计96段hex；其中旧V-SRR-01修正Code，其他旧71段保持。另有21条完整派生消息描述和摘要。新增2个XML正例，以及IMSUserState改4、SCSCFName放在IMSUserState之后这2个非法变体；原XSD/Xerces离线校验分别接受/拒绝，schema正确不代表订阅业务状态正确。完整Diameter执行、独立栈/真实DRA及号码策略/预算/竞态测试仍Planned。


## 21. 第二十二轮：三项裁决的派生候选

来源与支持边界见[主流程§18](diameterkit-mainflow-audit.md)，验收见[覆盖§24](diameterkit-usage-coverage.md)。沿用§20的派生方法：未列字段、顺序、flags及padding不变；替换后重算所有受影响AVP及消息长度、补零padding。新派生不是已执行的独立栈样本；原96段hex及21条派生记录保留不变。

### 21.1 可重建的消息与XML

- `D22-MIX-SNR-R/A/F`：基于V21-SUBMULTI三帧，Session统一为`as.example;22;1`，HBH=22220001/E2E=32320001；请求Ref12原位改25，删除710，在末尾加712=0/V1/M0；成功删除702，失败仍实验5104。703顺序11、25，Notif-Eff字段和共同IMPU保留，无Private Identity/Expiry。成功与失败是同请求的替代结果。
- `D22-MIX-PNR-R/A/F`：基于V-SH-PNR-01/PNA-01/PNA-02；Session统一为`hss.example;22;2`，HBH=22220002/E2E=32320002；请求702替换为下述X-SH-MIX-01，并在末尾添加628/V1/M0，子项依次Vendor266=10415、629=1/V1/M0、630=1/V1/M0。答案沿用2001或实验5100。双方Notif-Eff是场景前提；该PNR是独立exchange，不共享SNR标识，不声称由库关联原订阅。
- `D22-REALM-R/A/F`：基于V21-IMS-SRR三帧，Session统一为`gmsc.example;22;3`，HBH=22220003/E2E=32320003；成功2401原3100后依次追加3101=`isg.example`、3112=`example.net`，两叶V1/M1，余项不变。仍假定HSS知道目标IMSI，不与无IMSI未决分支捆绑；成功三元组仅在`s6c-ip-sm-gw-realm-v1`作为兼容正例，标准域仍报告未裁决。失败仍实验5001，不能因没有IMSI而自动选择该错误。
- `N22-SRA-HSSID/EMPTY/MISSING`：只把V21-IMS-SRR-A的User-Name分别替换为`001010000000001`、零长度UTF8String或删除；保留原V21的Session/HBH/E2E。HSSID变体的场景明确HSS也没有真实IMSI，故不能从数字外形假称标准成功；另外两项分别不满足成功身份内容/存在要求。它们都是负向/未支持决策输入，不是新成功pair。

`X-SH-MIX-01`：Ref11的IMSUserState=1与Ref25的MME可达值0同时通知；UTF8原文无BOM、XML声明及末尾换行，共232字节。该XML不含“订阅已删除”事实。

```xml
<Sh-Data><Sh-IMS-Data><IMSUserState>1</IMSUserState><Extension><Extension><Extension><UEReachabilityForIP><UEIPReachabilityMME>0</UEIPReachabilityMME></UEReachabilityForIP></Extension></Extension></Extension></Sh-IMS-Data></Sh-Data>
```

| 派生ID | 基帧 | 总长 | 整帧SHA-256 |
| --- | --- | --- | --- |
| D22-MIX-SNR-R | V21-SUBMULTI-R | 336 | `5f6536d4cd1bf817dfff01f084bee15d1be54ee5dff080f7188af7422ae2b181` |
| D22-MIX-SNR-A | V21-SUBMULTI-A | 196 | `db5b29efdd0911cb0f993d0d301ec74a32f8bdb99929f579b9c2bca8e8c5b84b` |
| D22-MIX-SNR-F | V21-SUBMULTI-F | 160 | `bf0d8488daf31d63bbcaec6ea6144eb439828fc17345dff1ac302e9a78d39f85` |
| D22-MIX-PNR-R | V-SH-PNR-01 | 516 | `66d0c3803e27f492d4adf81f40ba367cd4e106f367521ae486f923d6e17e275a` |
| D22-MIX-PNR-A | V-SH-PNA-01 | 140 | `0dafc0e8ec2815b3d8ddfab2d5efa33c9f2bf95e269460781dca03e5e36e0245` |
| D22-MIX-PNR-F | V-SH-PNA-02 | 160 | `9298b93ad3efbf30b091e8d15915b710edcecb595a1c2d1ee4e6d2b5a1ea3e0d` |
| D22-REALM-R | V21-IMS-SRR-R | 244 | `1f3da068cb5201dcdd423be860fb2166db9b3bcf16a7ff3eafbeedf17bfb63a4` |
| D22-REALM-A | V21-IMS-SRR-A | 216 | `740ea1c0b262d2832621cd8edef61a04c452d9ee04098858090fc5a5f63a9d8b` |
| D22-REALM-F | V21-IMS-SRR-F | 132 | `d76d52a18822ab9053e17f6b1f8d4d4a6a235b6f19a80da908313983ec34e779` |
| N22-SRA-HSSID | V21-IMS-SRR-A | 168 | `5510f184c63ea1845224944e80660ec7dd5f8da4e3cebc8bd56c9cc895a3d146` |
| N22-SRA-EMPTY | V21-IMS-SRR-A | 152 | `3922accec038b7c6b4ea39cd1665c85afce585bfc7b03bd37c9ac1e85b2ce6f9` |
| N22-SRA-MISSING | V21-IMS-SRR-A | 144 | `b5caf6d90514bc2e95deb834e28330af7b95c3047f2439a37e671d6bb13b6abc` |

### 21.2 逐分支预期与负例

| 候选变体 / 输入 | 预期 |
| --- | --- |
| MIX-SNR中11+25/712=0，无710、共同IMPU且双方Notif-Eff已知 | 已选有限子集可构造；712只限定25，SNA只有一个总结果 |
| 用12替11；或11后追加12形成11+12+25 | 同一有限规则，保留各一次及共同访问键；未给额外字节摘要的这些变体只作为决策用例 |
| MIX-SNR删712、712改1，或11/12单独却带712 | 分别违反25的必需项、枚举域、字段适用性；不能全局忽略712 |
| MIX-SNR加710=1或0、顶层Private Identity、重复703=25 | 不在本轮混合helper支持范围；按已证明的字段条件和未支持子集分别报告，不悄悄丢弃/去重或一律归为线路损坏 |
| MIX-SNR缺终端特性证据，或只给DRA CER | 现有PeerFeatures检查失败，不虚构HSS支持；单Ref拆分不复用多Ref假设 |
| MIX-SNR带有限709，SNA2001未回709 | 按无限期解释，不将请求时间补成最终上界；25仍只有一次通知语义，无库内定时删除 |
| SNA实验5104；或两次显式拆分SNR一成功一失败/超时 | 前者按HSS原子过程解释；后者交付两个结果，无总成功/自动补偿，不由库保证远端回滚 |
| MIX-PNR删除可达字段只留X-SH-CONT-01 | 通知11；不能推断25已触发或结束；独立入站正常交付 |
| MIX-PNR只保留X-SH-MME-01；或同时包含两片段 | 相关一次语义只涉及25；库不记消费状态，不以PNA结果控制远端订阅删除 |
| MIX-PNR先于SNA、退订后到达、来自另一次SNR | 无本地订阅存在性门槛；仍检查终端授权/命令/XML/预算，是否接受通知由应用决定 |
| X-SH-MIX-01将IMSUserState放到Extension后；或将UEIPReachabilityMME改1 | XSD sequence/枚举反例；与业务订阅是否存在分开 |
| REALM-A仅标准域；同bytes显式兼容域 | 分别报告标准组合未决、具名兼容结构可解释；无自动binding切换或连接新节点 |
| REALM-A删3100、重复3112、3112改为`bad realm` | 独立组合/数量/身份语法错误仍检查，兼容不能绕过 |
| REALM-A删3101形成N+R，或加3348形成N+H+R+B | 未被具名规则覆盖，不继承兼容正例；不是把所有Realm输入统一报5008 |
| N22-SRA-HSSID与原SRR配对 | 在明确无IMSI场景保留2001与原值、报告SMS-SRA-ID未支持；不转换成IMSI，不补dummy/查库，不等待另一份答案 |
| N22-SRA-EMPTY/MISSING | 成功内容/存在条件不满足；不因ABNF可选或UTF8零长可编码而交付完整成功 |
| 本地无IMSI成功Respond构造失败 | 提交前失败，无bytes提交；入站应答权按既有期限继续有效，应用决定后续答案；不自动代填实验5001 |

本节新增9条正向/失败替代派生消息及3条负向派生答案，共12条完整重建记录；不增加message-hex段数。静态长度/flags/配对/摘要与原XSD离线校验可核对，库执行、独立Diameter栈、HSS订阅行为与DRA互操作仍Planned。


## 22. 第二十三轮：字段、结果及文本边界候选

规则入口为[逐叶清单§3–6](diameterkit-field-inventory.md)，FIELD-14/RESULT-06/XML-03/TEXT-01。本节保留原96段hex与33条既有派生记录；不增加库代码或声称这些候选经过独立Diameter栈执行。以下XML原文无BOM、声明、换行；长度/摘要按UTF8 octets计算。**schema正例不等于本库typed能力正例**：X23-IDENT含旧PUR额外身份，Ref24有限builder不生成它；Repository/iFC/DSAI完整helper未启用，原文可有界承载/报告未解释范围。

### X23-IDENT

312字节；SHA-256 `8bd59790617c0825452efc9458cf718f7b55c037915b92e4ad1fbbbcebf511e4`。

```xml
<Sh-Data><PublicIdentifiers><IMSPublicIdentity>sip:alice&amp;team@example.net</IMSPublicIdentity></PublicIdentifiers><Sh-IMS-Data><Extension><Extension><Extension><SMSRegistrationInfo><IP-SM-GW-Number>12025550999</IP-SM-GW-Number></SMSRegistrationInfo></Extension></Extension></Extension></Sh-IMS-Data></Sh-Data>
```

### X23-REPO-DELETE

138字节；SHA-256 `50a0eec9ae110109e8f21b582acbf6aa3d0a1bdf3e37f79b4b40809490f9a831`。

```xml
<Sh-Data><RepositoryData><ServiceIndication>service-a</ServiceIndication><SequenceNumber>65535</SequenceNumber></RepositoryData></Sh-Data>
```

### X23-REPO-DATA

210字节；SHA-256 `ef2f5851608f1f2f1aeafe7f4861afbf2880b3799c0719b5ae378e99350c489c`。

```xml
<Sh-Data><RepositoryData><ServiceIndication>service-a</ServiceIndication><SequenceNumber>0</SequenceNumber><ServiceData><x:data xmlns:x="urn:example:app">你好</x:data></ServiceData></RepositoryData></Sh-Data>
```

### X23-IFC

210字节；SHA-256 `798acbb1cc468fd27b2ab82d5bd426455ebe4195a8f6942bafd36b28bb9e7271`。

```xml
<Sh-Data><Sh-IMS-Data><IFCs><InitialFilterCriteria><Priority>0</Priority><ApplicationServer><ServerName>sip:as.example.net</ServerName></ApplicationServer></InitialFilterCriteria></IFCs></Sh-IMS-Data></Sh-Data>
```

### X23-DSAI

161字节；SHA-256 `b8ebe4ba029b8f558e66b03f72575859a7d9b72fab5f88622c06a421e0198029`。

```xml
<Sh-Data><Sh-IMS-Data><Extension><Extension><DSAI><DSAI-Tag>service-a</DSAI-Tag><DSAI-Value>0</DSAI-Value></DSAI></Extension></Extension></Sh-IMS-Data></Sh-Data>
```


| 负例 / 相对上述原文的唯一变更 | 预期层次 |
| --- | --- |
| N23-IDENT-ESCAPE：X23-IDENT的`&amp;`改为裸`&` | XML非良构；不能按旧字符串替换模板生成 |
| N23-REPO-RANGE：X23-REPO-DELETE的65535改65536 | XSD拒绝；AVP716可编码65536不授权XML序号越界 |
| N23-REPO-TEXT：X23-REPO-DATA将整个x:data元素替换为裸文本YWJj | XSD拒绝，ServiceData不是裸base64字符串槽 |
| N23-IFC-SERVER：X23-IFC删除整个ApplicationServer子树 | XSD拒绝；仅Priority不足以构成iFC |
| N23-DSAI-VALUE：X23-DSAI的DSAI-Value改2 | XSD拒绝；0/1语义不能套为任意bool整数 |

### 22.1 逐叶与结果变体的具体输入

以下是确定性字段变更说明，重建时须重算AVP/消息长度与padding；尚未给整帧摘要的项不是新增完整bytes fixture。结果按namespace/vendor区分；成功/失败是替代答案，不能连续提交给同一exchange声称完成多次。

| 输入与变更 | 预期 / 对应缺口 |
| --- | --- |
| V21-IMS-RDR-R中3321改1，并在同一3320父组末尾添加3322/V1/M1，U32值0；与不带3322比较 | 0为存在的诊断，非缺省；“cause1+零诊断”的具体节点适用性仍按23.040检查，不转其他原因；FIELD-14/RESULT-06 |
| 同一诊断改13、14、255、256 | 13已定义，14/255保留值不冒充已定义原因，256越0..255语义域；四者线路宽度均4字节，不因U32能容纳就语义通过 |
| V-SH-PNA-02的实验5100分别改5001、5008、5107 | 均有29.328表6.1.4.1.1的应用含义，不能只接受原5100；5001/5107由应用事实决定，库不查订阅库 |
| 同一Sh实验5008改为基础268=5008且删除297 | 含义变为基础AVP_NOT_ALLOWED，不再是TooMuchData；不能使用相同错误标签 |
| V-TFA-02删除3303，实验码改5552 | MME/SGSN过程明确允许FacilityNotSupported；不因简表漏列拒绝；RESULT-06 |
| 上项实验码改5557，分别声明MME/SGSN应答角色 | MME有§7.3.10依据；SGSN不由MME条款自动授权标准生成；接收保留未知/不适用范围 |
| V-TFA-02保留5555但cause改4；V21-IMS-OFR-F改cause0 | MO/MT原因子集错误，即使都是合法枚举也不通过 |
| V-SH-UDA-01删除702仍2001；同一答案同时带实验5102 | 已选Ref24成功数据缺失、结果冲突分别报告，不被基础2001遮蔽 |
| D22-REALM-A仅3112改为`bad realm`；另一变体仅删除成功User-Name | 字段语法与成功条件独立检查；不会被Realm组合未决掩盖，沿用NODE-01/IDENTITY-02 |
| 任一节点/身份叶超过MaxFieldBytes；111行中每个单值叶重复/每个N叶改3或5字节 | 静态负例清单待实现逐行展开；总消息未超额也不能绕过单字段/数量/宽度预算 |

零诊断变体只是字段Presence的结构检查，不为IP-SM-GW定义MSC特有语义：对IP-SM-GW专属构造应选择已适用的11/12/13；若要验证MSC原因0的完整标准正例，父组必须相应改为MSC-SM-Delivery-Outcome3318。这一区分防止“值有定义”被误当任意节点都能生成。

### 22.2 国际化与raw不改写

| 字节/文本输入 | 预期 |
| --- | --- |
| UTF8String值“你好”，UTF8为`e4 bd a0 e5 a5 bd` | 6octets，非2octets；基础UTF8有效，字段若是SIP/NAI还要按用途检查 |
| UTF8String `00`、`c0 af`、`ed a0 80`、`f4 90 80 80`、单独`80` | 分别为NUL、过长、代理项、越Unicode上界、孤立续字节，不作为合法UTF8String交付；同bytes放OctetString不能套此错误 |
| `é`的`c3 a9`与`é`的`65 cc 81` | 均为UTF8且原文不同；不默认NFC合并Canonical内容 |
| Origin-Host=`hss.例子.example` | 非ASCII DiameterIdentity；不自动转IDNA后继续 |
| Origin-Host=`hss.xn--fsqu00a.example` | 可检查ASCII标签形式，但不能仅据`xn--`前缀宣称完整IDNA验证或DNS授权 |
| External-Identifier=`alice\@ops@example.net`与`用户@example.net` | 前者按4282转义解析，非按所有@计数；后者UTF8可承载但完整SASLprep/NAI语义未支持 |
| `sip:a%2Fb@example.net`与`sip:a/b@example.net` | 原表示不同，不解码重写来消除重复冲突 |
| X23-REPO-DATA中的中文业务文本 | XSD lax允许子元素不证明应用schema已验证；原文保持，Repository typed功能仍未支持 |

XML五个正例与五个负例以原XSD/Xerces静态验证；文本/结果表是验收预期，不能记为已运行库校验。独立栈/实际对端/资源测试仍Planned。


## 23. 第二十四轮：优先叶字段的完整帧补充

新增[逐叶完整报文候选](diameterkit-leaf-vectors.md)，136帧均有独立完整hex、长度/摘要、请求关联及正例/负例/边界分类；与本页既有96段hex、33条派生记录并存。优先零诊断、成功附加字段、五种Additional组合、跨组GPRS限制、时间依赖/2036回绕及MO/MT结果。规则见[主流程§20](diameterkit-mainflow-audit.md)，不把完整bytes误称独立Diameter实现已验证。§22的若干变更说明已有对应新完整帧，旧说明作为来源保留；完整范围与剩余项以逐叶清单§8为准。

## 24. 第二十五轮勘误与新完整候选

[逐叶候选§7–8](diameterkit-leaf-vectors.md)新增251条完整帧，含Sh完整特性协商/5011、T4的SGd-OFR/OFA成功/5555、SMEA完整数字地址和111叶静态反例。规则及证据见[主流程§21](diameterkit-mainflow-audit.md)。本页96段hex与33条派生记录原字节/哈希均保留。

**当前资格修正：** V21-SUBMULTI-R、D22-MIX-SNR-R、D22-MIX-PNR-R实际使用Notif-Eff却628 M0，改为发送规则负例/接收宽容控制，不能再称完整标准发送正例。其D22答案未回完整已支持特性集合也须补；新配对为V25-SH-NEED和V25-MIX-SNR/PNR。V24-S6C-SR-R的bit0=1但缺3309，同样不能再作为完整双节点正例请求；替代V25-SRR-TWO-R/V25-SRA系列。C24-SH-SN-5011仍仅结果识别边界。旧文“成功”只保留其当时字段验证范围，不能跳过此勘误作完整fixture认证。

## 25. 第二十六轮矩阵与定位反例

[命令与父路径矩阵](diameterkit-command-path-matrix.md)精确展开111叶，普通命令/E错误形状分开。新30条完整输入在[逐叶候选§9](diameterkit-leaf-vectors.md)，本页96段hex与33条派生记录不变。矩阵“在位候选”不是整pair合规认证，上一轮M位、SMEA及无特性答案勘误仍有效；定向反例必须命中同一App/Command/父路径/直接父实例。

## 26. 第二十七轮增量入口

本页96段hex和33条派生记录保持。新增56完整输入见[逐叶§10](diameterkit-leaf-vectors.md)，规则见[主流程§23](diameterkit-mainflow-audit.md)，[矩阵](diameterkit-command-path-matrix.md)增加27条应用路径定向证据。Sh基线限定Ref24/25；External访问键加Private为独立组合反例。N22-SRA-HSSID可复用同bytes作“值恰好是真实IMSI”的另一明确场景，不能仅按字符串外形裁决；这不把原无IMSI场景改为成功正例。Realm及相等时间控制保持未裁决分类，SMEA数字格式具备直接依据仍待独立验证。Ro不改实施分期。

## 27. 第二十八轮答案包络入口

本页96段hex/33条派生记录保持。[九答案族](diameterkit-answer-envelope.md)复用正常请求/结果并补148份独立可提取输入；E3002/E5012和已匹配无效/未关联输入分开，Failed-AVP引述非法内容不作为业务字段再次拒绝。完整规则见[主流程§24](diameterkit-mainflow-audit.md)，E错误不是普通ALA应用失败，也不能绕过成功SRA的身份条件。

## 28. 第二十九轮公共字段边界入口

[逐叶§12](diameterkit-leaf-vectors.md)新增81份父实例/未知扩展/诊断输入；[主流程§25](diameterkit-mainflow-audit.md)给出失败层次和资源场景。此页96段hex及33条派生记录不改字节，旧统计为历史快照。

## 29. 第三十轮公共字段入口

[逐叶§13](diameterkit-leaf-vectors.md)新增145完整帧，正常/E及仅接收结构控制分列；早期96段hex/33条派生记录不改。规则、源码参考和P1证据见[主流程§26](diameterkit-mainflow-audit.md)。

## 30. 第三十一轮公共字段入口

[逐叶§14](diameterkit-leaf-vectors.md)补214完整帧，包含父叶/请求缺项、Sh父组、目的准入及九E3002。旧96段hex和33派生记录不变；过程与策略前提、错误可构造性和规范证据见[主流程§27](diameterkit-mainflow-audit.md)。

## 第四十九批执行更新：有界结果向量

新增65份独立合成CCA/E，覆盖标准IEC/I/T及旧六用法的5031、版本5011和实验失败/冲突，附结构/命名空间/父路径/引文反例；累计902wire/17XML。63份原go-diameter字典精确往返，两份坏宽度/残缺引用只作本库测试。原837份字节保持，b48基础5011及实验5011的当前接收期望由本批测试接替旧负例，不修改历史证据。详见[本批契约](../spec/history/batch-49.json)、[执行证据](../interop/evidence/2026-11-07-ro-result.json)和[进度](../spec/progress.json)。

## 第五十批执行更新：原身份帧投影

新增27份头投影帧：标准IEC/I/T与旧六用途各一对T0/T1请求和新HBH答案；正文/E2E原样保留，原902帧不变，累计929wire/17XML。9份原字典、18份既有离线覆盖往返通过；不构成OCS去重或目标恢复接受证明。详见[本批契约](../spec/history/batch-50.json)、[独立证据](../interop/evidence/2026-11-07-ro-recovery.json)和[进度](../spec/progress.json)。
