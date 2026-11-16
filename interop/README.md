# 离线独立对照

证据工具带`ignore` build tag，在临时module运行，不进入本库依赖或发布构建，不启动SCTP、不修改其他仓库。对照本地`github.com/ctrlzy/go-diameter/v4` checkout；目录有未提交修改，因此记录HEAD和源内容散列，不能只凭HEAD复现。所有结果均为有限离线字段解释/重编码证据，不等于现场抓包、DRA/HSS/OCS认证、网络互通、完整协议语义或全局VECTOR-01完成。

## 历史证据与具体边界

| 批次 | 已执行范围 | 工具/原始记录 |
| --- | --- | --- |
| 1 | Base DWR/DWA、S6c SRR、成功SRA、未知用户SRA，共5帧；选定头/字段/结果/MME组和精确重编码 | [工具](go-diameter-check.go)、[证据](evidence/2026-11-06-base-s6c.json) |
| 2 | 普通SRR、E28 3002/5012、两实例Proxy正常/E答案和5001引用诊断，共7帧；Session/结果/Proxy-Host与二进制或空Proxy-State | [工具](go-diameter-envelope-check.go)、[证据](evidence/2026-11-06-envelope.json) |
| 15 | 普通RDR/RDA、HSS ALR/ALA，共7帧；字段及精确重编码，不验证恢复、持久化或角色部署 | [工具](go-diameter-s6c-alert-check.go)、[证据](evidence/2026-11-06-s6c-alert.json) |
| 16 | SGd/Gdd普通MO/MT共13帧；选择字段、目标/结果、位置组和精确旧库重编码 | [工具](go-diameter-sgd-check.go)、[证据](evidence/2026-11-06-sgd.json) |
| 17 | Sh四组成对请求/成功/失败共12帧；VSAI、UserIdentity、Enum、Time、结果；702为opaque | [工具](go-diameter-sh-check.go)、[证据](evidence/2026-11-06-sh.json) |
| 18 | Ref24/25原始XSD/Xerces共30项，13份文档候选加17派生/控制；XML及预期存证据 | [XML证据](evidence/2026-11-06-shdata.json) |
| 19 | 单Ref11/12/25订阅/退订15帧；D21三帧匹配原摘要、12份明确投影；710/712存在性、枚举、结果/配对 | [工具](go-diameter-sh-subscription-check.go)、[线路证据](evidence/2026-11-06-sh-subscription.json) |
| 19 | Ref11/12 XML共15项，4份XML加11控制；空名/URI/顺序/状态枚举；单数据Check不允许多数据 | [XML证据](evidence/2026-11-06-sh-ims-data.json) |

历史证据保持对应批次时点的限制；当前范围见[spec](../spec/README.md)及各批契约。第三至十四批等未新增独立外部协议帧的运行时验收见[history](../spec/history)，不能将本表当作其网络测试替代。

第一批负例发现旧库接受payload长度3的Auth-Session-State。本库`codec/TestRejectMalformedAndBounded/bad_u32`拒绝；旧库宽松接收不构成合法性依据。只对旧库传入这个受控畸形标量，不向旧解码器输入无界fuzz。

第十六批旧库默认生成字典未将EPS解释为Grouped，工具改为读取其**既有、未修改的XML源字典**；旧字典M位并非本库规范依据。V-OFR-01原顶层顺序与本库定义顺序不同，核心检查语义roundtrip/Canonical等价，旧库检查原序精确重编码。该批未验证IMS/T4、无IMSI、重投时间及未选结果/角色。

第十八/十九批使用未修改官方29.328附件XSD，固定schema摘要，外部Java只在审计时使用。生产及常规测试只需Go和仓库fixture。XML schema接受URI不表示helper已完整解释SIP/IDNA；Ref11/12单数据Check会拒绝额外数据。第十七批旧库对702的opaque检查不能作为XML验证证据。混合数据的新增范围单列如下。

## 第二十批：Sh特性/5011及组合通知

[go-diameter-sh-features-check.go](go-diameter-sh-features-check.go)使用内置Sh字典检查20帧的628及266/629/630、外层flags和精确重编码；[证据](evidence/2026-11-06-sh-features.json)记录本次工具、Profile和外部源码散列。17份原候选、3份明确Ref24投影；原V25-SH-PARTIAL-R的UDR含Ref11/12仍不在本库Pull子集，不以旧库可解码升级为本库已支持。旧库保留原AVP顺序，本库生成器输出固定顺序，测试验证稳定重编码/请求Canonical等价，不声称所有帧字节顺序相同。字典对照不负责终端事实、完整集合、5011触发、XML或现场互通；本库规则另有测试。

[七种组合XML证据](evidence/2026-11-06-sh-notification-data.json)使用原XSD及相同安全配置Xerces，数据存于证据中。常规Go测试/生产没有Java、schema或外部module依赖。具体规则与未覆盖项见[batch-20](../spec/history/batch-20.json)。

## 重跑方式

替换绝对路径；所有replace只在临时证据module，不修改核心go.mod：

```sh
core_repo=/absolute/path/diameterkit
peer_repo=/absolute/path/go-diameter
check_dir=$(mktemp -d)
go -C "$check_dir" mod init diameterkit-local-interop-check
go -C "$check_dir" mod edit -go=1.26.0 \
  -require=github.com/ctrlzy/go-diameter/v4@v4.0.0 \
  -replace="github.com/ctrlzy/go-diameter/v4=$peer_repo"
CGO_ENABLED=0 GOWORK=off GOTOOLCHAIN=local \
  go -C "$check_dir" run -mod=mod "$core_repo/interop/go-diameter-check.go" "$core_repo/testdata/codec"
```

将工具名换成本表对应文件即可。若外部checkout含本地SCTP编译依赖，按其module声明在**临时module**添加相应replace；它不会进入本库依赖。当前审计使用已有cache和`GOPROXY=off GOSUMDB=off`，没有下载依赖；只描述此次离线条件，不建议关闭发布验证。XML证据记录oracle源代码、安全配置、schema摘要、输入与预期，可独立复查。核心验收不要求消费者具备邻仓；更广互通验收沿原发布要求。

## 第二十一批：S6c普通HSS路由节点与SMEA

[独立工具](go-diameter-s6c-routing-check.go)使用原内置S6c字典核对20份未修改候选的Grouped、所选类型/flags/值和原序精确重编码，见[2026-11-07证据](evidence/2026-11-07-s6c-routing.json)。包括7个SRR/SRA成功基础帧及13个请求控制帧。SGSN-Name/Realm的M0、LMSI四字节、MWD显式0和不同Alert-MSISDN均核对；go-diameter对SMEA只提供OctetString，因此非法及超子集控制帧也可解码，不代表通过本库格式或双节点条件。SMEA实际peer语义互通P1仍开放。使用上述临时module重跑方式；不改第三方仓库、不添加核心依赖。实现范围及新预算/API说明见[batch-21](../spec/history/batch-21.json)。

## 第二十二批：S6c MSC/SGSN投递报告

[独立工具](go-diameter-s6c-outcomes-check.go)和[执行证据](evidence/2026-11-07-s6c-outcomes.json)覆盖17帧：11份未改候选，3个明确TLV请求投影及3个未改配对答案。16帧原序精确重编码；N24-RDR-DIAG3-R记录**有损差异**，该go-diameter checkout的DecodeUnsigned32对非4字节返回0且无错误，重编码将叶长度15改为16。本库拒绝该坏包；工具精确断言旧库的已观察行为，不将其计为独立语义反例通过。旧库字典也未检查MSC/MME互斥、diag256或成功附诊断；实际节点语义另由本库测试覆盖。未修改对方仓库，生产无外部依赖；完整边界见[batch-22](../spec/history/batch-22.json)。

## 第二十三批：MME/SGSN ALR/ALA

[工具](go-diameter-s6c-access-alert-check.go)和[证据](evidence/2026-11-07-s6c-access-alert.json)记录14帧原字典解码及精确重编码：7份未改候选（含Router token负控制）、7份仅改Event/节点Presence的明确TLV投影。检查类型、flags及配对标识，不认证端终端角色、事件条件、可达性或实际业务历史。旧库接受的零/缺事件及token不意味着本库支持；本批角色规则及范围见[batch-23](../spec/history/batch-23.json)。此前U32有损负例和SMEA互通缺口保留，未修改第三方仓库。

## 第二十四批：SGd/Gdd重投时间

[工具](go-diameter-sgd-retransmission-check.go)和[证据](evidence/2026-11-07-sgd-retransmission.json)覆盖17份未改逐叶候选：16帧原字典Time/OctetString、flags、配对标识及精确重编码；N24-TFA-TIME3则精确记录旧DecodeTime对3字节载荷返回零*Time且nil错误、随后改写为4字节的行为，不计作负例拒绝。旧库高位推测era不用于本库的成对语义认证。工具先序列化当前帧再读下一帧，避免对方读取缓冲复用影响Grouped原bytes。字典对照不证明同PLMN、实际节能状态、部署角色或相等规范裁决，第三方仓库及本库生产依赖未变。实现与测试见[batch-24](../spec/history/batch-24.json)。

## 第二十五批：T4 OFA External-Identifier

[工具](go-diameter-sgd-t4-check.go)和[证据](evidence/2026-11-07-sgd-t4.json)覆盖12帧：10份未改候选、2份仅清顶层3111 M位的发送投影。11帧精确往返，错误App16777311被原字典拒绝。原字典虽把3111列为must M,V，实际解码接受M0/M1，也接受坏UTF8、重复及不符合NAI/计费条件的字段；这些不构成语义负例通过。发送M0以29.338表6.3.3.1/2为准。未修改第三方字典/源码，核心无外部依赖。T4 SCTP与现场证据未完成，范围见[batch-25](../spec/history/batch-25.json)。

## 第二十六批：S6c T4无MSISDN报告与HSS提醒

[工具](go-diameter-s6c-t4-check.go)和[证据](evidence/2026-11-07-s6c-t4.json)覆盖10帧，全部精确往返。六份D21派生配方按原长度/摘要冻结，另有MME报告投影、RDA5001及双/空身份反例。原D21 IP-SM-GW报告在本库仍未支持；旧字典接受双/空身份，不能算业务语义通过。只核对离线字典类型/头与Session配对，不证明T4用途、身份事实、原生T4或SCTP部署。核心无第三方依赖，未修改其他仓库；有界范围见[batch-26](../spec/history/batch-26.json)。

## 第二十七批：T4用途的S6c路由

[工具](go-diameter-s6c-t4-routing-check.go)及[证据](evidence/2026-11-07-s6c-t4-routing.json)覆盖13帧，全部精确往返。D21-T4-SRR三份派生样本保留原摘要，另10份为独立原始TLV投影。旧字典接受缺/双/重复/非法身份、错误答案IMSI及缺/额外节点，不能代替原请求/用途/节点条件校验。无第三方源码修改、核心外部依赖或SCTP部署认证；见[batch-27](../spec/history/batch-27.json)。

## 第二十八批：SRA实验失败和三域诊断

[工具](go-diameter-s6c-sra-results-check.go)及[证据](evidence/2026-11-07-s6c-sra-results.json)共21帧：20份精确往返，1份三字节U32被旧字典接受并把AVP长度15改为16。后者是明确差异，不计精确往返或合规；重复/保留/错域/越255值的字典往返也不证明语义通过。旧V24无SMEA请求保留为控制，另补独立正向投影。没有第三方源码修改或SCTP部署声明；见[batch-28](../spec/history/batch-28.json)。

## 第二十九批：SGd OFA实验5001及接口边界

[工具](go-diameter-sgd-mo-interface-check.go)及[证据](evidence/2026-11-07-sgd-mo-interface.json)对8帧均精确往返；2份旧候选原样冻结，6份原始TLV投影。旧字典也接受Gdd5001、错误Vendor/重复/双结果，不能作为语义通过。端点来源、原MSISDN及SGd/Gdd事实由本库完整校验/应用证据负责；无抓包/SCTP部署声明，见[batch-29](../spec/history/batch-29.json)。

## 第三十批：S6c IP-SM-GW节点与Realm边界

[工具](go-diameter-s6c-ipsmgw-check.go)及[证据](evidence/2026-11-07-s6c-ipsmgw.json)对18帧精确往返，5份旧候选原SHA冻结、13份显式原始TLV投影。旧字典的10415/3112为OctetString，3348没有定义且以Unknown保留；不能据此认证DiameterIdentity、枚举值或Realm组合。它也接受缺号码/IMSI、格式/重复/跨节点反例，未支持语义不能记通过。未修改第三方库、未启用具名兼容、无SCTP部署声明，见[batch-30](../spec/history/batch-30.json)。

## 第三十一批：S6c IP-SM-GW投递结果

[工具](go-diameter-s6c-ipsmgw-outcome-check.go)及[证据](evidence/2026-11-07-s6c-ipsmgw-outcome.json)对16份独立原始TLV投影精确往返，与既有RDA2001/5001/5558保留相同Session/关联标识。旧字典将10415/3320、3321、3322分别解为Grouped、Enumerated、Unsigned32，但也接受缺cause、重复、非法值及错误父路径。这里只证明线路类型及字节保留，不认证语义、IMS/T4或部署互通；未修改第三方仓库，见[batch-31](../spec/history/batch-31.json)。

## 第三十二批：S6c IMS RDR/RDA

[工具](go-diameter-s6c-ims-report-check.go)及[证据](evidence/2026-11-07-s6c-ims-report.json)对20帧精确往返：7份旧候选原字节、13份独立原始TLV投影，核对原Session/HBH/E2E。3324/3327分别为Grouped/UTF8String；有效RDR中的HSS-ID位于3102/1。旧字典3325仍为OctetString，与29.338 V18.2.0有差异，本批仅在禁止位置负例中遇到；其对坏UTF8/重复/身份冲突的容忍不认证协议语义。未修改第三方仓库，无SIP栈、IMS其他命令或现场互通声明，见[batch-32](../spec/history/batch-32.json)。

## 第三十三批：S6c IMS ALR/ALA

[工具](go-diameter-s6c-ims-alert-check.go)及[证据](evidence/2026-11-07-s6c-ims-alert.json)对20帧精确往返：6份旧文档原字节、14份独立TLV投影，保留原Session/HBH/E2E。确认3324/3327/701/3329线路类型；旧字典容忍错误dummy/身份/父路径，Time往返也不验证era或实际可达性。3325旧类型差异仅出现于禁止位置负例，未修改第三方仓库，无现场或SCTP声明，见[batch-33](../spec/history/batch-33.json)。

## 第三十四批：S6c IMS SRR/SRA

[工具](go-diameter-s6c-ims-routing-check.go)及[证据](evidence/2026-11-07-s6c-ims-routing.json)对22帧精确往返：6份旧文档原字节、16份独立TLV投影，核对原Session/HBH/E2E及3324/3327/顶层User-Name类型。旧字典接受错误父路径/身份/URI不证明语义；同字节HSS-ID/IMSI不代表HSS-ID可替代真实IMSI。固定HSS恢复由本库TCP/TLS测试另证，不是旧字典互通结论。第三方仓库未修改，无现场/SCTP认证，见[batch-34](../spec/history/batch-34.json)。

## 第三十五批：IMS OFR/OFA

[工具](go-diameter-sgd-ims-mo-check.go)与[证据](evidence/2026-11-07-sgd-ims-mo.json)核对31帧：6份旧文档原字节和25份独立TLV投影，均精确往返并保留原pair关联。旧SGd/Gdd HSS-ID为UTF8String，不能套用其S6c字典的OctetString结论。错误dummy/URI/父路径/结果值域的字节容忍不是语义认证；本库用途、冻结和无自动重试由codec/TCP/TLS另证。第三方仓库未修改，无现场/SCTP声明，见[batch-35](../spec/history/batch-35.json)。

## 第三十六批：IMS TFR/TFA

[工具](go-diameter-sgd-ims-mt-check.go)与[证据](evidence/2026-11-07-sgd-ims-mt.json)核对31帧：3份D21原派生摘要相符，3份N26/N27原反例及25份独立TLV投影。保持原Session/HBH/E2E并检查UTF8/Grouped、3306 Unsigned32、3307/3330 Time，全部精确往返。旧库的坏父路径/值域容忍不代替语义验收；无字典覆盖、第三方修改或现场认证，见[batch-36](../spec/history/batch-36.json)。

## 第三十七批：IMS MO源身份

[工具](go-diameter-sgd-ims-identities-check.go)及[证据](evidence/2026-11-07-sgd-ims-identities.json)核对21份新独立TLV投影及2份原答案，共23项精确往返；保持原Session/HBH/E2E，检查源User-Name UTF8与MSISDN OctetString。坏值/重复/Vendor容忍只说明字节保留；源身份presence、dummy规则与5001关联/用途区别由codec和TCP/TLS验证，见[batch-37](../spec/history/batch-37.json)。无第三方仓库或字典修改，无现场认证。

## 第三十八批：S6c T4 IP-SM-GW结果

[工具](go-diameter-s6c-t4-ip-outcomes-check.go)及[证据](evidence/2026-11-07-s6c-t4-ip-outcomes.json)核对20份新独立TLV帧和3份D21原帧，共23项精确往返；验证IMSI UTF8、结果Grouped/Enumerated/U32及原关联。D21字节不变、当前完整用途支持与旧批未支持状态分开。坏父实例和值域容忍不代表语义合规，第三方仓库和字典未改，见[batch-38](../spec/history/batch-38.json)。

## Ro SMS IEC（第三十九批）

[工具](go-diameter-ro-sms-iec-check.go)使用现有App4字典和隔离临时module，核对35份冻结/独立投影帧的Unsigned64、Grouped、OctetString、Address、原Session/头标识及精确重编码；未使用本库encoder生成输入、未改旧库字典。另6份错误宽度只由本库拒绝测试覆盖。完整帧及来源摘要见[证据](evidence/2026-11-07-ro-sms-iec.json)，环境/模块替换沿前批离线流程。旧库保留反例不证明语义合法，也不代表OCS网络、余额、去重/恢复或SCTP通过。Ro目前仅标准SMS IEC单服务子集，六模板/私有绑定/RO-06仍属后续首版。

## Ro SMS ECUR（第四十批）

[工具](go-diameter-ro-sms-ecur-check.go)以原App4字典检查44份V-SMS/V-TERM原帧及独立投影的精确重编码、各自HBH/E2E、同Session和U64/Grouped/Enumerated/OctetString类型；一份坏U64宽度仅作本库拒绝测试。完整输入来源及源码摘要见[证据](evidence/2026-11-07-ro-sms-ecur.json)。无字典/第三方修改，无现场OCS或恢复认证。第39批证据将SMS-Node1误称SMS-SC是文字错误；本批按TS32.299 §7.2.212更正为IP-SM-GW，历史证据不回写，原字段值未变。

## sms-fw IEC具名模板（第四十一批）

[Python工具](smsfw-iec-fixtures.py)直接读取实际C++模板，按42节点顺序和固定宏输入独立构造31帧，并记录[来源映射](evidence/2026-11-07-smsfw-iec-source.json)。[go-diameter工具](go-diameter-smsfw-iec-check.go)用未修改的App4字典精确重编码29帧，另2份只作本库负例。见[执行证据](evidence/2026-11-07-smsfw-iec.json)。这些证据没有执行C++、没有生产OCS抓包；字典对旧字符串地址保留bytes，不证明legacy字段语义/父路径/支付结果。完整支持边界见[batch-41](../spec/history/batch-41.json)。

## sms-fw退款与Grouped420（第四十二批）

[来源投影](smsfw-refund-fixtures.py)直接消费实际退款模板，生成38帧并记录[26节点映射](evidence/2026-11-07-smsfw-refund-source.json)。[go-diameter工具](go-diameter-smsfw-refund-check.go)先记录标准字典拒绝，再加载[显式离线覆盖](smsfw-legacy420-offline.xml)，33帧精确往返。执行时在原临时模块命令末尾增加此XML的绝对路径。覆盖是工具进程内的全局键替换，不是本库路径规则或生产依赖；未修改第三方仓库。范围和实际观察见[证据](evidence/2026-11-07-smsfw-refund.json)及[batch-42](../spec/history/batch-42.json)，不声称C++执行/现场OCS接受。

## sms-fw ECUR原模板（第四十三批）

[独立投影](smsfw-ecur-fixtures.py)按实际C++顺序消费算法读取INITIAL41/TERM44节点，生成49帧并记录[源映射](evidence/2026-11-07-smsfw-ecur-source.json)。[旧库工具](go-diameter-smsfw-ecur-check.go)使用原App4字典对47帧精确往返，无覆盖；原TERM1210在1212前，另有固定sender顺序投影，两帧不混称同一原输入。见[执行证据](evidence/2026-11-07-smsfw-ecur.json)及[batch-43](../spec/history/batch-43.json)。这些仍非C++执行输出/现场OCS证明。

## sms-fw通用模板（第四十四批）

[独立投影](smsfw-general-fixtures.py)读取实际通用模板32节点，按默认false的IPv4 writer输出RFC Address，产生55个输入。旧库[独立工具](go-diameter-smsfw-general-check.go)对47个输入使用原字典精确往返，另8个坏宽度/冲突仅为本库反例；见[执行证据](evidence/2026-11-07-smsfw-general.json)和[batch44](../spec/history/batch-44.json)。字典425的unsigned解释不证明旧请求signed语义，字节一致也不证明业务/父路径/原Action校验。无字典覆盖、无第三方修改；CCA合成，非C++执行/OCS抓包。

## SCAP1离线基础（第四十五批）

[源投影](smsfw-scap1-fixtures.py)读取47节点及11重复607组，61输入含合成基础ACA。[独立工具](go-diameter-smsfw-scap1-check.go)先记录原字典193/607为Unknown，再显式加载[私有定义](smsfw-scap1-offline.xml)，56帧精确重编码，5个宽度/头域反例仅本库检验。见[源映射](evidence/2026-11-07-smsfw-scap1-source.json)、[执行证据](evidence/2026-11-07-smsfw-scap1.json)及[batch45](../spec/history/batch-45.json)。过程内字典不验证父路径/业务条件，不是生产运行依赖；没有C++执行、现场ACA或SCAP1网络收发证明。

## SCAP1运行时与离线复验（第四十六批）

本机TCP/TLS的显式会计协商、ACR/ACA和关闭/重连证据在[batch46](../spec/history/batch-46.json)及对应测试；仅代表本库对本库测试端点。[离线复验](evidence/2026-11-07-smsfw-scap1-runtime.json)仍是原56帧显式字典重编码，更新scap1 slice.2元数据摘要，既有61帧和源投影不变。两种证据均不代表旧双Auth/Acct广告、私有ACA或实际DRA/OCS现场验收。

## 五CCR固定组合（第四十七批）

[独立投影](smsfw-combined-fixtures.py)读取冻结源帧，人工固定次序产生六请求及配对答案/跨模板反例，共36帧。[旧go-diameter工具](go-diameter-smsfw-combined-check.go)先观察原420冲突，再以[显式进程内字典](smsfw-combined-offline.xml)固定420/425/1292/2018，36帧精确重编码。见[投影](evidence/2026-11-07-smsfw-combined-source.json)、[执行证据](evidence/2026-11-07-smsfw-combined.json)和[batch47](../spec/history/batch-47.json)。字典能往返本库拒绝的跨模板输入，不能证明父路径/分支、请求与答案币种类型、OCS格式/顺序接受或业务扣费。

## 旧CCA已证字段（第四十八批）

[独立构造](smsfw-answer-fixtures.py)产生67份分层结果/费用/DDFH及边界CCA，见[源消费清单](evidence/2026-11-07-smsfw-answer-source.json)。[旧go-diameter工具](go-diameter-smsfw-answer-check.go)使用未覆盖的原字典，66帧精确重编码；一个坏u32宽度反例仅本库校验。原字典DDFH为Enumerated，旧C++读取为Unsigned32，本子集只选0/1的共同四字节表示；不声明全域signed语义一致。见[执行证据](evidence/2026-11-07-smsfw-answer.json)与[batch48](../spec/history/batch-48.json)。字典不校验原动作/父实例，不证明未知token/PSP命名空间或现场扣费结果。

## 第四十九批：原字典结果命名空间对照

工具`go-diameter-ro-result-check.go`用原App4字典读取63份新增输入并检查结果/实验父组/支持列表/诊断类型及精确重编码；无字典覆盖。两份坏宽度/残缺引文保留为本库反例。可以重编码并不证明结果配对、父实例/必需项或OCS接受；无C++执行或外部网络互通。详见[本批契约](../spec/history/batch-49.json)、[执行证据](../interop/evidence/2026-11-07-ro-result.json)和[进度](../spec/progress.json)。

## 第五十批：恢复帧的头与正文保持

go-diameter-ro-recovery-check.go独立读取27份头投影帧：先用原字典核对9份标准帧，再加载既有smsfw-combined-offline.xml核对18份旧模板帧。均精确重编码；覆盖只用于离线工具，不进入生产或修改其他仓库。字典不验证发送证据、目标去重、恢复许可或财务状态。详见[本批契约](../spec/history/batch-50.json)、[独立证据](../interop/evidence/2026-11-07-ro-recovery.json)和[进度](../spec/progress.json)。

## Batch69: Ro ASR/ASA

`python3 interop/ro-abort-fixtures.py` independently constructs 53 RFC6733
App4/274 positive/negative frames (982 wire fixtures in total).
`go-diameter-ro-abort-check.go` checks eight positive frames using the local
go-diameter dictionary with no diameterkit import, including the absence of
CCA-specific mandatory fields in ASA. The exact local source hashes and
limitations are recorded in [evidence](evidence/2026-11-10-ro-abort.json).
This is offline dictionary evidence, not live OCS or legacy C++ execution.

## Batch70: generic DCCA single-service EVENT

`python3 interop/dcca-event-fixtures.py` constructs71 new independent frames;
`go-diameter-dcca-event-check.go` verifies33 positive frames against the local
dictionary, including signed money/exponent, unsigned currency/units and
normal/E answers. See [source snapshot and output](evidence/2026-11-10-dcca-event.json).
This covers generic RFC4006 representation, not SMS failure rules,2022 tokens,
legacy templates or liveOCS acceptance.

## 第77批：模拟器源码模式投影

[工具](go-diameter-simulator-fixtures.go)在冻结go-diameter工作树执行，用其encoder构造28份既有模拟器输入/修正帧；源应用19文件与参考编码器摘要见[证据](evidence/2026-11-10-simulator-inputs.json)。不是旧服务运行或抓包：固定Session/头/端点，TPDU用opaque测试数据替代，S6a只有头部。实际接受/拒绝及迁移责任见[矩阵](../doc/diameterkit-simulator-migration.md)，不以字典编码成功证明条件语义合法。

## 第78批：SCAP1旧广告

[独立Python投影](scap1-advertisement-fixtures.py)依照冻结C++配置和编码顺序生成17帧；[go-diameter检查](go-diameter-scap1-advertisement-check.go)确认严格/旧双组CER/CEA四帧的typed VSAI与精确往返。[证据](evidence/2026-11-10-scap1-advertisement.json)记录两份C++及参考Go/XML/module摘要。字典容纳双Auth/Acct不证明RFC合法；标准codec拒绝，只有具名Peer握手模式承接固定源布局。未执行旧C++或目标网络。
