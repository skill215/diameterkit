# 既有模拟器输入与迁移边界

日期：2026-11-10，第77批。完成前批约定的SIM-01输入接受/拒绝决策；本表是有界源码路径清单，不追加全字段规范调查。现有库规则足以承接修正后的选定路径，本批只增加测试、报文和证据，不修改生产代码、Profile、预算或重放权限。Go1.26、纯Go、无持久化边界保持。

## 来源与证据强度

只读检查`ip-sm-gw`的三个既有模拟器：`go/svc/mme-simulator/internal/diameter-handler`、`go/svc/diameter-smsc-poc/internal/handler`、`sim/sh_server/server.go`，并固定相关配置、号码编码及module共19个文件的实际工作树SHA256。HEAD不是干净发布证明。准确路径、摘要与独立go-diameter参考源码快照见[机器证据](../interop/evidence/2026-11-10-simulator-inputs.json)。

[独立生成工具](../interop/go-diameter-simulator-fixtures.go)在冻结的go-diameter工作树中执行，使用其`smparser`、`Marshal`、`Answer`和encoder构造28帧；没有调用diameterkit生成器。没有执行三个旧服务、配置加载器、旧BCD包或`fkgi/sms` TPDU构造器，也没有连接旧服务或目标DRA。它们是**源码构造模式的投影**，不是抓包或原二进制全链路结果：

- 固定Session、HBH/E2E；应答的原请求采用已有`v-*`报文，Origin/Realm采用测试授权端点。ALR也先换成合法测试Origin，单独检查标志/身份，不声称字节复现其默认监听地址。
- OFR保留默认`localhost:3868`、17位拼接IMSI及号码字段问题；每步修正均单列。号码投影只含已观察到的数字字面值，未实现完整旧BCD行为。
- TPDU用明确的6字节opaque测试数据（含0/255及非ASCII字节）替代，证明库保留字节，不证明日本语短信编码。日本语UCS2编码/分片由应用的短信库负责。
- `<Sh-Data/>`只用于成功UDA必需User-Data的线路正例；不以此证明Ref24深层XML业务有效，深层规则继续使用`app/sh/shdata`的既有测试。
- ULR/AIR只有App/Command头部投影，证明当前Bundle不接纳S6a，不是完整S6a报文或解析能力证明。

协议边界依据固定的[TS29.329 V18.0.0 §6.1](https://www.etsi.org/deliver/etsi_ts/129300_129399/129329/18.00.00_60/ts_129329v180000p.pdf)与[TS29.338 V18.2.0 §5/6](https://www.etsi.org/deliver/etsi_ts/129300_129399/129338/18.02.00_60/ts_129338v180200p.pdf)，以及既有Profile子集。go-diameter能编码某值不代表该值满足本库完整规则。

## 实际构造与决策

下列报文前缀均为`testdata/codec/b77-sim-`，完整哈希见[manifest](../testdata/codec/manifest.json)。错误标签表示库检查层，不要求一次Decode列出所有缺陷；例如同时含坏IMSI/Origin的报文先报IMSI。

| 来源路径 / 行为 | 当前输入决策与修正 | 可执行证据 |
| --- | --- | --- |
| MME `handle-ofa.go` → OFR | 参考`NewRequest`仅设R，源路径未补P；`mme-ofr-original`拒绝。补P后17位`46031013800000000`仍拒绝；Origin中的`host:port`也独立拒绝。应用显式使用合法DiameterIdentity和已配置≤15位IMSI；库不截断、去前缀或补身份 | `mme-ofr-header-fixed`、`mme-ofr-host-fixed`、`mme-ofr-host-invalid`分别定位；`mme-ofr-fixed`通过，普通MO允许已选IMSI与MSISDN同在 |
| MME `getTBCD` | 源码先删除`+`再编码号码；库`EncodeTBCD`只接受数字且不含TON/NPI。明确的国际数字格式转换由应用负责，不自动删除任意字符；不能将该helper用于SMEA | 修正OFR的SC解码为`138888888`；直接传`+138888888`拒绝；opaque TPDU保真 |
| MME `handle-ala.go` → ALR | 源路径也缺P；修正P后User-Identifier同时IMSI/MSISDN不满足当前ALR身份选择。修正为原IMSI、原SC/Event，显式`AlertRole(AlertFromMME)`与端点证据；零Context保留NeedsApplication | `mme-alr-original`、`mme-alr-both-identities`拒绝，`mme-alr-fixed`静态通过并完成角色校验；库不代替应用丢弃身份 |
| MME `handle-srr.go` → SRA | 2001附原数值IMSI和MME节点通过；原5012仍附成功身份/节点，拒绝。修正失败包去掉生产字段，保留正确结果和包络 | `mme-sra-2001`、`mme-sra-5012`、`mme-sra-failure-fixed`；TCP/TLS拒绝后可修正Respond，首次失败未提交 |
| MME `handle-tfr.go` / `handle-rdr.go` | 授权端点、正确原请求包络下2001/5012通过；不把所有请求目的复制为有权响应的Origin | `mme-tfa-2001/-5012`、`mme-rda-2001/-5012`成对关联；错授权源拒绝 |
| SMSC POC `handle-ofr.go` → OFA | 原Auth-Session-State=0不在选定SGd包络，拒绝；显式改为1通过。不能用成功Result=2001掩盖错误状态 | `poc-ofa-state0/-state1`；真实连接注入非法匹配答案只完成一次，随后正确晚答不再次完成 |
| Sh server UDR → UDA | 原2001没有User-Data，拒绝；5012无数据通过。成功必须由应用提供相应数据，不由库生成空业务结果 | `sh-uda-2001/-5012/-fixed`；真实连接拒绝缺失数据，再用合法应答完成 |
| Sh server PUR → PUA / SNR → SNA | 本批固定普通PUR/SNR请求的2001/5012成对通过；不把这一结果扩展为所有含Expiry/Send-Data/混合特性的SNR输入 | `sh-pua-2001/-5012`、`sh-sna-2001/-5012`；其他订阅条件沿用既有Sh测试和Context |
| Sh server RDR → RDA | 原RDA含SMS VSAI。现有SMS接收忽略、发送禁止政策保持：可解码及关联，不可直接原样Append作本库发送；重建合法RDA时不带该字段 | `sh-rda-vsai`分别校验receive与send，不能把接收成功记成双向透传支持 |
| 三个模拟器的手工答案包络 | 旧handler未回显Proxy；MME还复制请求Destination作为Origin。Proxy存在时旧答案不能完整关联，未知/错误授权Origin也拒绝。迁移用当前`Inbound.Respond`或离线`CompleteAnswer`，保留每个Arrival自己的Session/Proxy/头标识 | `sh-pua-no-proxy`反例及重建正例；TCP/TLS两个不同Proxy Arrival与第三个缓存重答，无重复业务交付 |
| Sh server ULR/AIR | 属COMPAT-02后续S6a，不因模拟器注册而进入本批；显式拒绝。PNR构造仅注释，未列为活动源码路径，库既有PNR/PNA能力不变 | `s6a-316/-318`头部拒绝；无完整S6a认证主张 |

`Marshal`会替换AVP列表：旧`Answer(code)`再`Marshal(body)`并不自动造成两个Result-Code。`Answer`只清R并沿用原头标识，不会自行补Session/Proxy；上述差异来自读取冻结实现并实际构造报文，不从函数名猜测行为。

## 配置与应用接线

以下是读取源码得到的迁移要求，不冒充运行过旧进程的测试结果：

- 分离监听地址与Origin-Host；`localhost:3868`可以是网络地址，不能直接作为当前Profile的DiameterIdentity。MME `Run`当前走TCP且传空证书，虽然另有通用TLS监听helper；不能据helper存在宣称旧MME已启用TLS。新库TCP/TLS验收来自本库对本库连接。
- SMSC POC的VSAI将Vendor-ID填成SGd App-ID 16777313；迁移配置用Vendor10415及正确Auth-App，不从数字猜类型。Base的uint32 Vendor-ID语法不禁止其他值，这里不登记为通用codec非法报文。
- 原响应写到保存的连接指针、解引用可缺省Destination-Host及非`/v4`导入混用需要在应用迁移清除。库提供Inbound及owned请求证据，不承接旧指针生命周期、配置加载、发包TPS循环或业务失败比例。
- 使用有限配置的目标/来源策略与Route；邻接DRA不等于MME/HSS/SMSC终端身份。订阅内容、ALR角色、号码归属及重投调度继续由Go应用提供。未添加通用模拟器服务器框架、存储、自动业务重放或任意App-ID重映射。

## 验收出口及仍未完成项

[codec测试](../codec/simulator_test.go)覆盖11个拒绝用例、14个接收成对答案（其中1个SMS VSAI仅接收）、修正请求、角色、owned答案、错E2E、Proxy重建与错源。[运行时测试](../simulator_test.go)覆盖真实TCP/TLS上的Sh UDR/PUR/SNR、SGd OFR/TFR及S6c失败应答修正、取消的Respond不提交、多Arrival/缓存重答、非法匹配答案与晚答一次完成和预算归零。本批未增加SCTP模拟器场景，SCTP运行原有完整回归。

输入决策到此结束；测试/构建摘要见[batch77](../spec/history/batch-77.json)。SIM-01整体仍需独立的旧应用迁移与同输入新旧进程对照，真实DRA/目标接受也未完成；不据28个合成报文宣称这些门已经通过。后续工作只按已列门交付，不再重复扩张本表或逐叶规范。Ro私有字段/SCAP1证据门、公开依赖发布、API示例与最终压力性能验收保持独立。
