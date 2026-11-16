# 九答案族：正常/E错误包络、结果与原请求关联覆盖

> **当前范围（2026-11-04，Draft 0.20）：** 第一阶段=首版/v1，包含Sh、S6c、SGd/Gdd、Ro/DCCA及sms-fw实际所需六模板、SCAP1和RO-06恢复能力；Go 1.26、纯Go实现（含依赖与SCTP适配器，发布路径CGO_ENABLED=0）、公开Go库、无持久化边界不变。明确撤销第十九轮“Ro后置”分期；带日期的旧分期/统计仅作历史证据，不再授予Ro跳过首版验收的例外。限定S6a、完整SCUR/PRICE_ENQUIRY等未承诺能力仍后置。范围及剩余缺口见[Specification §3/§7](diameterkit-specification.md)，顺序见[总设计§14](diameterkit-design.md)，验收见[覆盖§39](diameterkit-usage-coverage.md)。2026-11-06已完成前十九批有界实施（Base/S6c基础、公共包络与关联、出站运行时、关闭与选举、TCP/TLS生产适配、有限Inbound/应答、有限Reject/应急预算、Canonical/有限重复缓存及多Arrival重答、固定同终端Route/多attempt与活动请求原身份重传、有限Dialer/Maintain拨号及重连、按class/编码字节DRR调度、固定Peer的有限Listener/Serve、有界标识源/Session/guard、Base/S6c有限Context/Report、S6c普通RDR/RDA及HSS ALR/ALA、SGd/Gdd普通OFR/OFA与TFR/TFA、Sh普通四组命令及透明User-Data、Ref24/25有限Sh-Data helper、Sh单订阅/退订及Ref11/12有限数据）；已执行范围及未完成项以[Specification §6](diameterkit-specification.md#6-需求与证据的维护)和[spec进度](../spec/progress.json)为准，未列明的运行验收仍Planned。

日期：2026-10-31，第二十八轮，Draft。规则依据与裁决唯一维护于[主流程§24](diameterkit-mainflow-audit.md)，验收见[覆盖§30](diameterkit-usage-coverage.md)。这是可审核的文档输入，不是新DSL、公开API或运行测试系统。Go1.26公开无持久化端点库；Ro纳入首版实施。

## 1. 范围与读表

九族为UDA/PUA/SNA/PNA、SRA/RDA/ALA、OFA/TFA；旧SGd-App的Alert兼容不混入标准ALA。每族复用原请求及正常成功/失败，分别补E3002与E5012；正常业务成功字段不复制进E错误。所有变体是互斥的离线场景，不是同一exchange连续多答。

线路外前提：当前Node内活动attempt，实际连接代次C1，原请求bytes已保存，deadline未过；终端Origin为基线成功答案的host/realm；ErrorAgentOrigins明确允许dra-a.example/example的E基础3002；绝不因Peer CER或294字符串新增授权。E5012默认来自相应终端。没有持久pending、身份/订阅查询或自动重试。

## 2. 九族包络与结果基线

| 答案族 | 原请求 | 正常E0成功 | 正常E0失败 | E1基础3002 | E1基础5012 |
| --- | --- | --- | --- | --- | --- |
| UDA | V24-SH-UD-R | V24-SH-UD-A | F24-SH-UD-5001 | E28-UDA-3002 | E28-UDA-5012 |
| PUA | V24-SH-PU-R | V24-SH-PU-A | F24-SH-PU-5001 | E28-PUA-3002 | E28-PUA-5012 |
| SNA | V24-SH-SN-R | V24-SH-SN-A | F24-SH-SN-5001 | E28-SNA-3002 | E28-SNA-5012 |
| PNA | V24-SH-PN-R | V24-SH-PN-A | F24-SH-PN-5100 | E28-PNA-3002 | E28-PNA-5012 |
| SRA | V25-SRR-TWO-R | V25-SRA-SGSN-MSC | V25-DIAG-MME | E28-SRA-3002 | E28-SRA-5012 |
| RDA | V24-S6C-RD-R | V24-S6C-RD-A | F24-S6C-RD-5001 | E28-RDA-3002 | E28-RDA-5012 |
| ALA | V24-S6C-AL-R | V24-S6C-AL-A | F28-ALA-N-5012 | E28-ALA-3002 | E28-ALA-5012 |
| OFA | V24-SGD-MO-R | V24-SGD-MO-A | F24-SGD-MO-5555-C4 | E28-OFA-3002 | E28-OFA-5012 |
| TFA | V24-SGD-MT-R | V24-SGD-MT-A | F24-SGD-MT-5555-C0 | E28-TFA-3002 | E28-TFA-5012 |

正常失败沿用已选应用结果：ALA是基础5012而非虚构实验用户未知；SRA为原双节点请求关联的5550/MME诊断控制。E1的3xxx/5012不是复制上述业务失败；E5语法允许不代表任意实验5xxx可直接置E。

## 3. 每族独立负例索引

`N28-族-N-后缀`使用正常成功基线，`N28-族-E-后缀`使用该族E3002基线。每个单元均在§6给完整hex。NORESULT/DUAL/DUPRESULT检查结果选择和数量；SESSION是完整头命中后的Session冲突；NOHOST/NOREALM/NOSTATE分别缺字段。E-NOSESSION与E-NORESULT为已关联无效答案；E-E2E为未关联输入，不终结原pending。矩阵不能只按相同Code搜索判通过。

| 族 | 正常结果缺失/双类/重复 | 正常Session/Origin/State | E关联/必需结果 | Sh VSAI Auth-App |
| --- | --- | --- | --- | --- |
| UDA | N28-UDA-N-NORESULT / DUAL / DUPRESULT | N28-UDA-N-SESSION / NOHOST / NOREALM / NOSTATE | N28-UDA-E-NOSESSION / NORESULT / E2E | N28-UDA-N-VSAIAPP |
| PUA | N28-PUA-N-NORESULT / DUAL / DUPRESULT | N28-PUA-N-SESSION / NOHOST / NOREALM / NOSTATE | N28-PUA-E-NOSESSION / NORESULT / E2E | N28-PUA-N-VSAIAPP |
| SNA | N28-SNA-N-NORESULT / DUAL / DUPRESULT | N28-SNA-N-SESSION / NOHOST / NOREALM / NOSTATE | N28-SNA-E-NOSESSION / NORESULT / E2E | N28-SNA-N-VSAIAPP |
| PNA | N28-PNA-N-NORESULT / DUAL / DUPRESULT | N28-PNA-N-SESSION / NOHOST / NOREALM / NOSTATE | N28-PNA-E-NOSESSION / NORESULT / E2E | N28-PNA-N-VSAIAPP |
| SRA | N28-SRA-N-NORESULT / DUAL / DUPRESULT | N28-SRA-N-SESSION / NOHOST / NOREALM / NOSTATE | N28-SRA-E-NOSESSION / NORESULT / E2E | SMS发送不带，接收沿用SMS-VSAI |
| RDA | N28-RDA-N-NORESULT / DUAL / DUPRESULT | N28-RDA-N-SESSION / NOHOST / NOREALM / NOSTATE | N28-RDA-E-NOSESSION / NORESULT / E2E | SMS发送不带，接收沿用SMS-VSAI |
| ALA | N28-ALA-N-NORESULT / DUAL / DUPRESULT | N28-ALA-N-SESSION / NOHOST / NOREALM / NOSTATE | N28-ALA-E-NOSESSION / NORESULT / E2E | SMS发送不带，接收沿用SMS-VSAI |
| OFA | N28-OFA-N-NORESULT / DUAL / DUPRESULT | N28-OFA-N-SESSION / NOHOST / NOREALM / NOSTATE | N28-OFA-E-NOSESSION / NORESULT / E2E | SMS发送不带，接收沿用SMS-VSAI |
| TFA | N28-TFA-N-NORESULT / DUAL / DUPRESULT | N28-TFA-N-SESSION / NOHOST / NOREALM / NOSTATE | N28-TFA-E-NOSESSION / NORESULT / E2E | SMS发送不带，接收沿用SMS-VSAI |

## 4. 共享维度与线路外场景

共享代码可以复用规则，但以下代表输入不冒称九族每个变体均已有完整bytes：UDA覆盖E=2001/4001、E缺Origin/重复Session/Session非首位/结果宽度、禁止Destination、P/T/保留位、错误HBH/App/Command及未授权Origin；SNA覆盖E5001的Failed-AVP存在/重复/空组。Proxy有服务端本跳请求与正常/E答案两条完整回显控制。

| 输入/执行前提（分别应用于§2九族对应E3002帧） | 可观察结果 |
| --- | --- |
| C1、活动attempt、头与Session/Origin全部满足 | 原pending完成一次，交付基础失败；无恢复/业务重发 |
| 相同bytes在C2到达，C2没有该attempt映射 | 不借C1的HBH命中；有界丢弃，不消耗C1的A/完成权 |
| 相同bytes在另一个Peer到达，数值HBH碰巧相同 | 每连接关联域隔离；不跨Peer找Session或E2E来“修复” |
| 当前exchange已经完成/取消/超期，映射退休 | 有界丢弃或既有超限关闭，不新增第二个Completion/历史库 |
| 旧连接已完整读入的候选，在切换后进入完成门 | 必须带原ConnectionID/AttemptID，不改标为新代；仍保留匹配记录且旧候选符合R19截止/关闭条件时才可竞争；不能一律把旧代合法候选误杀 |
| SNR待答时到达独立PNR，或本地有同Session的其他exchange | PNR走请求入口，不完成SNR；Session不是pending索引，业务CorrelationID也不是线路键 |

每个上下文行对九族逐项验收；无需复制相同bytes九次。未知/不匹配头的丢弃预算与已匹配无效答案的A预算不同，具体处理顺序见主流程§24.3。所有路径不得向答案再发错误答案。

## 5. 完整新增输入清单

每帧列出来源和原请求。结果/来源/头改变时以目标说明为准，不能默认所有列出的请求都应得到成功答案；字节相同而场景不同也不重复算独立编码验证。旧基线见[逐叶语料](diameterkit-leaf-vectors.md)和[早期报文](diameterkit-protocol-vectors.md)。

| ID | 来源 | 原请求 | 目标与预期 | 字节 | SHA-256 |
| --- | --- | --- | --- | --- | --- |
| E28-UDA-3002 | V24-SH-UD-A | V24-SH-UD-R | E1基础3002；已授权DRA无法投递，Origin=dra-a.example/example。无277/260/业务成功载荷，保留原Session和全部头关联。 | 96 | `5fe0c826b077fcba8311d6eaff2858082545d88d3eea58fd3c870f1cb8e01d3a` |
| E28-UDA-5012 | V24-SH-UD-A | V24-SH-UD-R | E1基础5012；终端不能高效组成应用答案，RFC6733允许该永久错误用基础错误语法。不按实验5012解释。 | 96 | `c77fc7da1fb7dd8119fe66172f021f5886c6aafbc72228d21f92f1d35c1c28f9` |
| N28-UDA-N-NORESULT | V24-SH-UD-A | V24-SH-UD-R | 正常E0两类结果均缺；不能把默认0视为成功。 | 404 | `4cab8df56b24a144ff2c2e5d192921a4bd631679bf8ebc3e33c8b6efd4609f6b` |
| N28-UDA-N-DUPRESULT | V24-SH-UD-A | V24-SH-UD-R | 正常E0基础2001重复，即使相同也不后写覆盖。 | 428 | `bef4e1cf1202e6d391c358b4a9af60cff869a9734fe9f240b11ee15e8586fce4` |
| N28-UDA-N-SESSION | V24-SH-UD-A | V24-SH-UD-R | 头关联完整命中，但Session不等于原请求；Exchange违规，ErrInvalidAnswer。 | 424 | `a84287557a056a7ae726b8ae6cd77111b99daf7d2b5127920a5095bfba980d04` |
| N28-UDA-N-NOHOST | V24-SH-UD-A | V24-SH-UD-R | 正常E0缺Origin-Host；不能用Destination/Peer身份补。 | 396 | `08b2ae6bb222deaab3d5970370678590e68745199f7e3b869e58e1b769f85f57` |
| N28-UDA-N-NOREALM | V24-SH-UD-A | V24-SH-UD-R | 正常E0缺Origin-Realm；按已采用4887及字段语法检查。 | 396 | `7193fa0fbf00d16a90ac4c4de2369980f57892861c680893b55502017953241a` |
| N28-UDA-N-NOSTATE | V24-SH-UD-A | V24-SH-UD-R | 正常E0缺本应用必需Auth-Session-State；不能套E语法放行。 | 404 | `11ea3e05060e3dc303401a2079b5b2fcfa75aa17dca530ee08d5bd2af6ffe065` |
| N28-UDA-N-DUAL | V24-SH-UD-A | V24-SH-UD-R | 正常E0基础2001与实验10415/5001并存；不产生业务成功。ALA本身亦无此已选应用失败分支。 | 448 | `75991cf5f676cd844e967a7e2bb74f00d6d063418774b1c8468d2414b9008ebf` |
| N28-UDA-E-NOSESSION | E28-UDA-3002 | V24-SH-UD-R | 原请求有Session，E错误不能凭0..1语法省略；完整头命中后ErrInvalidAnswer。 | 72 | `24cc76c32a737a62165e3989939815843d297af2029b95c3b4f1eadcfa509582` |
| N28-UDA-E-NORESULT | E28-UDA-3002 | V24-SH-UD-R | E1缺必需基础Result-Code；不从实验结果或E位猜码值。 | 84 | `9552a5610cbd38ac45bebf43949426b2f419f075aacf2e3e647e4159ce256eac` |
| N28-UDA-E-E2E | E28-UDA-3002 | V24-SH-UD-R | HBH候选命中但E2E不同；完整头未关联，有界丢弃，不消费原pending的A/完成权。 | 96 | `9ce7a42c9a14a6c8a997d641b0fb2e47ddd9c6617e53c4332f90ca815c4f05ce` |
| N28-UDA-N-VSAIAPP | V24-SH-UD-A | V24-SH-UD-R | Sh正常答案VSAI的Auth-App错为S6c，不能被Header的正确Sh App补足。 | 416 | `b41393cfc01f62a756aaa2d465a115771c4ad31c553088aa7639e6880be2dc63` |
| E28-PUA-3002 | V24-SH-PU-A | V24-SH-PU-R | E1基础3002；已授权DRA无法投递，Origin=dra-a.example/example。无277/260/业务成功载荷，保留原Session和全部头关联。 | 96 | `8b0feeb675ef4916ecb7a13938e12a283635604d23b6135c4d4941885696eaf5` |
| E28-PUA-5012 | V24-SH-PU-A | V24-SH-PU-R | E1基础5012；终端不能高效组成应用答案，RFC6733允许该永久错误用基础错误语法。不按实验5012解释。 | 96 | `572de8a84f13128eabd94d79affaa38c40453ab08282c0cc235778bcb000d2e0` |
| N28-PUA-N-NORESULT | V24-SH-PU-A | V24-SH-PU-R | 正常E0两类结果均缺；不能把默认0视为成功。 | 128 | `5e8e0f9892b028c8f6338e1caea3ced35316db443e8e9d3a1bfda4d220abb389` |
| N28-PUA-N-DUPRESULT | V24-SH-PU-A | V24-SH-PU-R | 正常E0基础2001重复，即使相同也不后写覆盖。 | 152 | `2e4602d476cac9473d1a36caf6a55054b881c78e902d351666debf454c52331e` |
| N28-PUA-N-SESSION | V24-SH-PU-A | V24-SH-PU-R | 头关联完整命中，但Session不等于原请求；Exchange违规，ErrInvalidAnswer。 | 148 | `51ee1af937397a251f9118271f4bc0922f18e81b93fe5c1ad6560596f3737d4f` |
| N28-PUA-N-NOHOST | V24-SH-PU-A | V24-SH-PU-R | 正常E0缺Origin-Host；不能用Destination/Peer身份补。 | 120 | `5034196c47e2d4594b0b984beb2ff3af335302ef52d259c94ac606d0071c3917` |
| N28-PUA-N-NOREALM | V24-SH-PU-A | V24-SH-PU-R | 正常E0缺Origin-Realm；按已采用4887及字段语法检查。 | 120 | `92131fa7ef95446bffeabce0071a4e5f3e3a7fea7aeb88164b2a973ad7e695dd` |
| N28-PUA-N-NOSTATE | V24-SH-PU-A | V24-SH-PU-R | 正常E0缺本应用必需Auth-Session-State；不能套E语法放行。 | 128 | `39c2156e490f109bd42e0dcea46be1d223d6c5e0b01dd3e6b68536c39c4de555` |
| N28-PUA-N-DUAL | V24-SH-PU-A | V24-SH-PU-R | 正常E0基础2001与实验10415/5001并存；不产生业务成功。ALA本身亦无此已选应用失败分支。 | 172 | `c75a66bb693938070686968c6bb47279e423737d4bd4addfcd92ff26d9a6c198` |
| N28-PUA-E-NOSESSION | E28-PUA-3002 | V24-SH-PU-R | 原请求有Session，E错误不能凭0..1语法省略；完整头命中后ErrInvalidAnswer。 | 72 | `f7b61767cbc2494c0d0663d25bf968d03ed9f0af0a0d4d7bfed51a44abcc3383` |
| N28-PUA-E-NORESULT | E28-PUA-3002 | V24-SH-PU-R | E1缺必需基础Result-Code；不从实验结果或E位猜码值。 | 84 | `0522b21a2c612a104fd42a90eec65c33915a5ee526de44a47656846e9c51bf51` |
| N28-PUA-E-E2E | E28-PUA-3002 | V24-SH-PU-R | HBH候选命中但E2E不同；完整头未关联，有界丢弃，不消费原pending的A/完成权。 | 96 | `0bbd8194496c662cbd8813d1951c1436ba4275eb00d0c121f19f8d723b756c47` |
| N28-PUA-N-VSAIAPP | V24-SH-PU-A | V24-SH-PU-R | Sh正常答案VSAI的Auth-App错为S6c，不能被Header的正确Sh App补足。 | 140 | `77779ae2ebbf9929e29d2d60edddcd807b536db4c8a5371a60d67830677e4bd5` |
| E28-SNA-3002 | V24-SH-SN-A | V24-SH-SN-R | E1基础3002；已授权DRA无法投递，Origin=dra-a.example/example。无277/260/业务成功载荷，保留原Session和全部头关联。 | 96 | `a07e838b11cb47144bd74a4784307f14ed67ee9e665f8cbd436abd468b18e120` |
| E28-SNA-5012 | V24-SH-SN-A | V24-SH-SN-R | E1基础5012；终端不能高效组成应用答案，RFC6733允许该永久错误用基础错误语法。不按实验5012解释。 | 96 | `08641d5c9d1de7da6b8350afb1972f01731520834036b8e8206920a2198a93b1` |
| N28-SNA-N-NORESULT | V24-SH-SN-A | V24-SH-SN-R | 正常E0两类结果均缺；不能把默认0视为成功。 | 232 | `1162a8bbb525f6131bf1b24b2d39f37a54779d2e574156cf43559e13a5fb43e0` |
| N28-SNA-N-DUPRESULT | V24-SH-SN-A | V24-SH-SN-R | 正常E0基础2001重复，即使相同也不后写覆盖。 | 256 | `b38983c3299cc1549d4d0ff78140d005ac6575f4ada828bf05dcb5f543144029` |
| N28-SNA-N-SESSION | V24-SH-SN-A | V24-SH-SN-R | 头关联完整命中，但Session不等于原请求；Exchange违规，ErrInvalidAnswer。 | 252 | `7f4b27876a7b0342c03f8504731a8af50eb061b662b4343e9c48965075ff76c8` |
| N28-SNA-N-NOHOST | V24-SH-SN-A | V24-SH-SN-R | 正常E0缺Origin-Host；不能用Destination/Peer身份补。 | 224 | `c919034a3a7b89d70aa230ebec17534339aaaabe81ff3c530e5de323e0820aeb` |
| N28-SNA-N-NOREALM | V24-SH-SN-A | V24-SH-SN-R | 正常E0缺Origin-Realm；按已采用4887及字段语法检查。 | 224 | `a35cc3d17e5ed0d81342b2b5f915580ad64914a318d05ac930a9618a20ce70fe` |
| N28-SNA-N-NOSTATE | V24-SH-SN-A | V24-SH-SN-R | 正常E0缺本应用必需Auth-Session-State；不能套E语法放行。 | 232 | `cfddd28a967ad84191b93c918f59ef6bda0903cbea9c661470a808dcee54d892` |
| N28-SNA-N-DUAL | V24-SH-SN-A | V24-SH-SN-R | 正常E0基础2001与实验10415/5001并存；不产生业务成功。ALA本身亦无此已选应用失败分支。 | 276 | `61d25c1f50a72d5697e81cbddd1a90b71208b1ff3d08091c85a56f5d93f3e1b6` |
| N28-SNA-E-NOSESSION | E28-SNA-3002 | V24-SH-SN-R | 原请求有Session，E错误不能凭0..1语法省略；完整头命中后ErrInvalidAnswer。 | 72 | `cc9308d15e13e3a24f037c7f9c4532aba24d4c27480949637ef8231c6bf2384b` |
| N28-SNA-E-NORESULT | E28-SNA-3002 | V24-SH-SN-R | E1缺必需基础Result-Code；不从实验结果或E位猜码值。 | 84 | `45d26d649c3e3dccbb4b44c076ecc380153f0dd73c43bb040c3402b79f7c388d` |
| N28-SNA-E-E2E | E28-SNA-3002 | V24-SH-SN-R | HBH候选命中但E2E不同；完整头未关联，有界丢弃，不消费原pending的A/完成权。 | 96 | `a657dd139e084a24fa86f37591e6ffbafb4e1fa64d25af483174498c8e2b9cdd` |
| N28-SNA-N-VSAIAPP | V24-SH-SN-A | V24-SH-SN-R | Sh正常答案VSAI的Auth-App错为S6c，不能被Header的正确Sh App补足。 | 244 | `73c64b1da97f9517cd01b3e6c5789c91f738aef51840d72a0a91289ec779bfcb` |
| E28-PNA-3002 | V24-SH-PN-A | V24-SH-PN-R | E1基础3002；已授权DRA无法投递，Origin=dra-a.example/example。无277/260/业务成功载荷，保留原Session和全部头关联。 | 96 | `8ca19263aa953cc2bb3455f40d887362a549e11645cb559ffe533e547a24a488` |
| E28-PNA-5012 | V24-SH-PN-A | V24-SH-PN-R | E1基础5012；终端不能高效组成应用答案，RFC6733允许该永久错误用基础错误语法。不按实验5012解释。 | 96 | `bebe0d851ded7da0b0a385c5fb74ec6184ce7d0eac330a3c60bb0d75b3d26b17` |
| N28-PNA-N-NORESULT | V24-SH-PN-A | V24-SH-PN-R | 正常E0两类结果均缺；不能把默认0视为成功。 | 128 | `6f0bc99ed664091e66f70fb5b4b24c5eb1862280c8a6d896f4236ff0ff44f2aa` |
| N28-PNA-N-DUPRESULT | V24-SH-PN-A | V24-SH-PN-R | 正常E0基础2001重复，即使相同也不后写覆盖。 | 152 | `d901c5fd997d8958f195630659345edac8aa071ef7953fdc47599520a13cd300` |
| N28-PNA-N-SESSION | V24-SH-PN-A | V24-SH-PN-R | 头关联完整命中，但Session不等于原请求；Exchange违规，ErrInvalidAnswer。 | 148 | `732b65058fb2b8b16170b4f09a16fdc4e177cc8dcd443d783d3598f635621404` |
| N28-PNA-N-NOHOST | V24-SH-PN-A | V24-SH-PN-R | 正常E0缺Origin-Host；不能用Destination/Peer身份补。 | 120 | `822457a2dd51fe258fbf9f9af2d0dfd940f4541427b228975af0cca6feeec54d` |
| N28-PNA-N-NOREALM | V24-SH-PN-A | V24-SH-PN-R | 正常E0缺Origin-Realm；按已采用4887及字段语法检查。 | 120 | `1a9a238d71dc16d08dbe7d8db3bfc09368a57a6aca534874c68d9ece943ec5ab` |
| N28-PNA-N-NOSTATE | V24-SH-PN-A | V24-SH-PN-R | 正常E0缺本应用必需Auth-Session-State；不能套E语法放行。 | 128 | `6176a07e60edbea239328e0515469f1cac0883701c84bc900a5570aa1a8d9595` |
| N28-PNA-N-DUAL | V24-SH-PN-A | V24-SH-PN-R | 正常E0基础2001与实验10415/5001并存；不产生业务成功。ALA本身亦无此已选应用失败分支。 | 172 | `da30880b35e803faa978acf13d79e09815ff53128642b4bc1d52e57c19d1ea8a` |
| N28-PNA-E-NOSESSION | E28-PNA-3002 | V24-SH-PN-R | 原请求有Session，E错误不能凭0..1语法省略；完整头命中后ErrInvalidAnswer。 | 72 | `6a381842028827ca689a0b3dffdc89dcd06c5f15f8086501c95b73c20a721a33` |
| N28-PNA-E-NORESULT | E28-PNA-3002 | V24-SH-PN-R | E1缺必需基础Result-Code；不从实验结果或E位猜码值。 | 84 | `9006ccb6e97b4d9407c765e24b6e392b16b6b812ccb18bc19c63f713ba3c52f0` |
| N28-PNA-E-E2E | E28-PNA-3002 | V24-SH-PN-R | HBH候选命中但E2E不同；完整头未关联，有界丢弃，不消费原pending的A/完成权。 | 96 | `e1396f0bfbc32ef77cc48fa7b5ac6871830096f7409e0fc4e2ec55475d9d6b82` |
| N28-PNA-N-VSAIAPP | V24-SH-PN-A | V24-SH-PN-R | Sh正常答案VSAI的Auth-App错为S6c，不能被Header的正确Sh App补足。 | 140 | `0c3c5bf78a8df76101006ef2ca06cff3600e014363c706722433adcefc45e897` |
| E28-SRA-3002 | V25-SRA-SGSN-MSC | V25-SRR-TWO-R | E1基础3002；已授权DRA无法投递，Origin=dra-a.example/example。无277/260/业务成功载荷，保留原Session和全部头关联。 | 100 | `29aec485533ddd52051f44215d636262ade20cc856502789fef7954b9a5a9840` |
| E28-SRA-5012 | V25-SRA-SGSN-MSC | V25-SRR-TWO-R | E1基础5012；终端不能高效组成应用答案，RFC6733允许该永久错误用基础错误语法。不按实验5012解释。 | 100 | `c956b190120fdbec34820f9a6671e94c6747bec67cc7ccb4860e895f64e9f24a` |
| N28-SRA-N-NORESULT | V25-SRA-SGSN-MSC | V25-SRR-TWO-R | 正常E0两类结果均缺；不能把默认0视为成功。 | 252 | `c3762c0253f74b59025a095d57b2695fdc0c646f6dec9581daa3cc72e8952011` |
| N28-SRA-N-DUPRESULT | V25-SRA-SGSN-MSC | V25-SRR-TWO-R | 正常E0基础2001重复，即使相同也不后写覆盖。 | 276 | `11f41fe128d16d69f03427ed99e0bcb129ea7b10cbc6d96d68cb4a606a08ed19` |
| N28-SRA-N-SESSION | V25-SRA-SGSN-MSC | V25-SRR-TWO-R | 头关联完整命中，但Session不等于原请求；Exchange违规，ErrInvalidAnswer。 | 268 | `9f17e3bf06d2e8c87dfae18b0c26700b8e5d13409f29799bff3a48399940ecd3` |
| N28-SRA-N-NOHOST | V25-SRA-SGSN-MSC | V25-SRR-TWO-R | 正常E0缺Origin-Host；不能用Destination/Peer身份补。 | 244 | `c0fedcb3a587f8130a307ee859f32406fab9ff1b85880bc260289e06848bc1ef` |
| N28-SRA-N-NOREALM | V25-SRA-SGSN-MSC | V25-SRR-TWO-R | 正常E0缺Origin-Realm；按已采用4887及字段语法检查。 | 244 | `d4edc8cdc48dc8ad1341d62f16b56dd5337f93303d739389f628f1f42af5b2d2` |
| N28-SRA-N-NOSTATE | V25-SRA-SGSN-MSC | V25-SRR-TWO-R | 正常E0缺本应用必需Auth-Session-State；不能套E语法放行。 | 252 | `7c718c71eded618d9e7b6fd2be90c15db332d27582bfb0cca292cfdfee8d65d5` |
| N28-SRA-N-DUAL | V25-SRA-SGSN-MSC | V25-SRR-TWO-R | 正常E0基础2001与实验10415/5001并存；不产生业务成功。ALA本身亦无此已选应用失败分支。 | 296 | `a8a455a65e1f0277c75bf4fd3332c6002fc9b4308a31496756fc9e8f67798582` |
| N28-SRA-E-NOSESSION | E28-SRA-3002 | V25-SRR-TWO-R | 原请求有Session，E错误不能凭0..1语法省略；完整头命中后ErrInvalidAnswer。 | 72 | `785a90733e543580ee643bdfb0731c85e455e2262c590bed252897fb7a31972e` |
| N28-SRA-E-NORESULT | E28-SRA-3002 | V25-SRR-TWO-R | E1缺必需基础Result-Code；不从实验结果或E位猜码值。 | 88 | `53b0078ace17e1700988a264826d312e1be6fce9dedb90ca8d01e01453829782` |
| N28-SRA-E-E2E | E28-SRA-3002 | V25-SRR-TWO-R | HBH候选命中但E2E不同；完整头未关联，有界丢弃，不消费原pending的A/完成权。 | 100 | `5acb284b5ede34590e6d069467ff432bd60fd12dcc567e8c47878f4b9bd6e325` |
| E28-RDA-3002 | V24-S6C-RD-A | V24-S6C-RD-R | E1基础3002；已授权DRA无法投递，Origin=dra-a.example/example。无277/260/业务成功载荷，保留原Session和全部头关联。 | 100 | `21336ba77c29067bc5ae4ad8de84ea506df48a727c7c6595ea03313de3f98978` |
| E28-RDA-5012 | V24-S6C-RD-A | V24-S6C-RD-R | E1基础5012；终端不能高效组成应用答案，RFC6733允许该永久错误用基础错误语法。不按实验5012解释。 | 100 | `e9092218c484a8a45fb51214b784a76be8e49fbbac6242b1a5f515be702ea5e6` |
| N28-RDA-N-NORESULT | V24-S6C-RD-A | V24-S6C-RD-R | 正常E0两类结果均缺；不能把默认0视为成功。 | 132 | `b8948ff13c9705e4432d3741dc036d014f2fc898206a52dfc2c080a9a51ddfb7` |
| N28-RDA-N-DUPRESULT | V24-S6C-RD-A | V24-S6C-RD-R | 正常E0基础2001重复，即使相同也不后写覆盖。 | 156 | `473c275050b059147dd8f2d1513541464b98f77ac9c953db1d168f84cc6e8f91` |
| N28-RDA-N-SESSION | V24-S6C-RD-A | V24-S6C-RD-R | 头关联完整命中，但Session不等于原请求；Exchange违规，ErrInvalidAnswer。 | 148 | `587e7783511582e3f12ee811f36eb41fa8add014f05ca9525cb889fc897a59b0` |
| N28-RDA-N-NOHOST | V24-S6C-RD-A | V24-S6C-RD-R | 正常E0缺Origin-Host；不能用Destination/Peer身份补。 | 124 | `c1fdb4f350c3c88f2298bfd14878a79a356ef61d0e00684254b7c35e481fbda3` |
| N28-RDA-N-NOREALM | V24-S6C-RD-A | V24-S6C-RD-R | 正常E0缺Origin-Realm；按已采用4887及字段语法检查。 | 124 | `96a2a1b2f9a9237f80bd857870e8f1d34c330f1e8787b03cab642ded603ceebf` |
| N28-RDA-N-NOSTATE | V24-S6C-RD-A | V24-S6C-RD-R | 正常E0缺本应用必需Auth-Session-State；不能套E语法放行。 | 132 | `9b27fe56f5729108ba533dd64bd290670d5eec31b31c064592ea2ab981bc3c9d` |
| N28-RDA-N-DUAL | V24-S6C-RD-A | V24-S6C-RD-R | 正常E0基础2001与实验10415/5001并存；不产生业务成功。ALA本身亦无此已选应用失败分支。 | 176 | `203fe996ad396052d25c72b315552e6c6ed9279803e07a02f55006cba00ccf44` |
| N28-RDA-E-NOSESSION | E28-RDA-3002 | V24-S6C-RD-R | 原请求有Session，E错误不能凭0..1语法省略；完整头命中后ErrInvalidAnswer。 | 72 | `f036d4dd64298704faa34bd8f53effa2ee91e846604c4b1ef215a036d1bc951e` |
| N28-RDA-E-NORESULT | E28-RDA-3002 | V24-S6C-RD-R | E1缺必需基础Result-Code；不从实验结果或E位猜码值。 | 88 | `66dd5a69b5e28e44f1da43c4ae4132cee5c3a1bd4c616158f0491fbabe130230` |
| N28-RDA-E-E2E | E28-RDA-3002 | V24-S6C-RD-R | HBH候选命中但E2E不同；完整头未关联，有界丢弃，不消费原pending的A/完成权。 | 100 | `638eca6de8b3b5fbdb2b035d6d77ea1a0010b7210b20e85b0fd5bca7ef3f8084` |
| F28-ALA-N-5012 | V24-S6C-AL-A | V24-S6C-AL-R | ALA没有专用3GPP应用错误；E0基础5012，场景为终端不能完成该请求，不造用户未知。 | 108 | `9374b37dfb884688c179163c965682a161a82e879ef10258fa0f297a5d46bf37` |
| E28-ALA-3002 | V24-S6C-AL-A | V24-S6C-AL-R | E1基础3002；已授权DRA无法投递，Origin=dra-a.example/example。无277/260/业务成功载荷，保留原Session和全部头关联。 | 96 | `a5e7d7b72d008e26fecc4a4c63ac39767fd687911019688489bed5b54d5a9fa5` |
| E28-ALA-5012 | V24-S6C-AL-A | V24-S6C-AL-R | E1基础5012；终端不能高效组成应用答案，RFC6733允许该永久错误用基础错误语法。不按实验5012解释。 | 96 | `9f3843d5b1dde22a2d7a33ac968b1300597429db445c1baeca1383dee2f15dc1` |
| N28-ALA-N-NORESULT | V24-S6C-AL-A | V24-S6C-AL-R | 正常E0两类结果均缺；不能把默认0视为成功。 | 96 | `0035d6c48465530f22c560b323ca945b912168e5be53b07f62c9f5a0df2fdb1b` |
| N28-ALA-N-DUPRESULT | V24-S6C-AL-A | V24-S6C-AL-R | 正常E0基础2001重复，即使相同也不后写覆盖。 | 120 | `992847eff8822731e6280f9daf47d0a13aa4848cf0c57f22e52db321c42b1b1a` |
| N28-ALA-N-SESSION | V24-S6C-AL-A | V24-S6C-AL-R | 头关联完整命中，但Session不等于原请求；Exchange违规，ErrInvalidAnswer。 | 116 | `16fe6ceabf0d23a3b342d58b6ac3e4d67d127cf26804eaea5d8e80d22dad0c7a` |
| N28-ALA-N-NOHOST | V24-S6C-AL-A | V24-S6C-AL-R | 正常E0缺Origin-Host；不能用Destination/Peer身份补。 | 88 | `6fd5c5c2e2f699b011b4a4d5a535760be62553cfd2156a359227f1fec16e9ecd` |
| N28-ALA-N-NOREALM | V24-S6C-AL-A | V24-S6C-AL-R | 正常E0缺Origin-Realm；按已采用4887及字段语法检查。 | 88 | `0e11979efd0ce9bbfe86da73e7f01dbd2739dfb2666f864e8e6a192d85cd6abd` |
| N28-ALA-N-NOSTATE | V24-S6C-AL-A | V24-S6C-AL-R | 正常E0缺本应用必需Auth-Session-State；不能套E语法放行。 | 96 | `ac1f76aebc9e6fd9516e87e5d187b061cd91c2b3085a8b518dee03b982211270` |
| N28-ALA-N-DUAL | V24-S6C-AL-A | V24-S6C-AL-R | 正常E0基础2001与实验10415/5001并存；不产生业务成功。ALA本身亦无此已选应用失败分支。 | 140 | `278ff5f571bfae484aeeb7c907d636e29a8f70abccf3f0e7ec0283819f1a0f63` |
| N28-ALA-E-NOSESSION | E28-ALA-3002 | V24-S6C-AL-R | 原请求有Session，E错误不能凭0..1语法省略；完整头命中后ErrInvalidAnswer。 | 72 | `3e264ffa4abe080e36fe8f1633bca89fd680b1ad1dd952a100e80148f9f69818` |
| N28-ALA-E-NORESULT | E28-ALA-3002 | V24-S6C-AL-R | E1缺必需基础Result-Code；不从实验结果或E位猜码值。 | 84 | `11b512af5dac97f4de16ac30cedca065dbc0c2e8c735b79dd4055332141b9515` |
| N28-ALA-E-E2E | E28-ALA-3002 | V24-S6C-AL-R | HBH候选命中但E2E不同；完整头未关联，有界丢弃，不消费原pending的A/完成权。 | 96 | `b3cf4f65c8159c7b7ce863b1d1c0e1332a8e07b1f95745c35b06b037de508fe8` |
| E28-OFA-3002 | V24-SGD-MO-A | V24-SGD-MO-R | E1基础3002；已授权DRA无法投递，Origin=dra-a.example/example。无277/260/业务成功载荷，保留原Session和全部头关联。 | 100 | `57dddfe7bac9d52d92756e865a4e528887b92f3b4e13c3ba0ffd372fb0bfa2b8` |
| E28-OFA-5012 | V24-SGD-MO-A | V24-SGD-MO-R | E1基础5012；终端不能高效组成应用答案，RFC6733允许该永久错误用基础错误语法。不按实验5012解释。 | 100 | `7cd125db8d7d49702828d1966425ece3a3fa18322465f2c77dfcb37d08b50f86` |
| N28-OFA-N-NORESULT | V24-SGD-MO-A | V24-SGD-MO-R | 正常E0两类结果均缺；不能把默认0视为成功。 | 124 | `5992f62c6a3267b7f8c0b99627950d195e08150939727c6d8b934af61f824644` |
| N28-OFA-N-DUPRESULT | V24-SGD-MO-A | V24-SGD-MO-R | 正常E0基础2001重复，即使相同也不后写覆盖。 | 148 | `35071c354446ec969a09ce611ea32176f1d3dfb2bcdb6a18c18edeee9ce4fc92` |
| N28-OFA-N-SESSION | V24-SGD-MO-A | V24-SGD-MO-R | 头关联完整命中，但Session不等于原请求；Exchange违规，ErrInvalidAnswer。 | 140 | `4377ccade8b1ccd19709d32a978e9a7ab856d680f42a9f998b58cfe1c14b19af` |
| N28-OFA-N-NOHOST | V24-SGD-MO-A | V24-SGD-MO-R | 正常E0缺Origin-Host；不能用Destination/Peer身份补。 | 116 | `c520ca240e826714e3b26d87c23a1e5b157073b389449c1c6bbecd400b7e5200` |
| N28-OFA-N-NOREALM | V24-SGD-MO-A | V24-SGD-MO-R | 正常E0缺Origin-Realm；按已采用4887及字段语法检查。 | 116 | `9752f2bdb2dd3eb9d389a25fadf76753b8844747aef68925983bb5b9aac8ec2c` |
| N28-OFA-N-NOSTATE | V24-SGD-MO-A | V24-SGD-MO-R | 正常E0缺本应用必需Auth-Session-State；不能套E语法放行。 | 124 | `1ecb97bd857167211e0b4eb196489cc2e43f7c3a72b37f6d85ef4d62f475e0dc` |
| N28-OFA-N-DUAL | V24-SGD-MO-A | V24-SGD-MO-R | 正常E0基础2001与实验10415/5001并存；不产生业务成功。ALA本身亦无此已选应用失败分支。 | 168 | `4cde0bdf0f1d78e4115167ef8f31a84e61b48ca0d42437f6b10bd50063faedbe` |
| N28-OFA-E-NOSESSION | E28-OFA-3002 | V24-SGD-MO-R | 原请求有Session，E错误不能凭0..1语法省略；完整头命中后ErrInvalidAnswer。 | 72 | `7b4837e1d18941537bb368f2cfcb84c4637d74700c42f0a9a506c361aaeb1533` |
| N28-OFA-E-NORESULT | E28-OFA-3002 | V24-SGD-MO-R | E1缺必需基础Result-Code；不从实验结果或E位猜码值。 | 88 | `ca27861a823ab71d928f2c86de94cca0acdec6b14ec4679308567ba16bce0f44` |
| N28-OFA-E-E2E | E28-OFA-3002 | V24-SGD-MO-R | HBH候选命中但E2E不同；完整头未关联，有界丢弃，不消费原pending的A/完成权。 | 100 | `5050bb7e835c7420e60397aee7d8aa9c5c4c15a8b83ab579fd729a9978595a6e` |
| E28-TFA-3002 | V24-SGD-MT-A | V24-SGD-MT-R | E1基础3002；已授权DRA无法投递，Origin=dra-a.example/example。无277/260/业务成功载荷，保留原Session和全部头关联。 | 100 | `5db748832a9108e9b62331b33c7dbbc74af95a8065e822962d8ee4ef960faa25` |
| E28-TFA-5012 | V24-SGD-MT-A | V24-SGD-MT-R | E1基础5012；终端不能高效组成应用答案，RFC6733允许该永久错误用基础错误语法。不按实验5012解释。 | 100 | `3fdc0ba493e4717917e3c2ff7e5c503de2c91c08f5583e1a972a61d9838a2a4b` |
| N28-TFA-N-NORESULT | V24-SGD-MT-A | V24-SGD-MT-R | 正常E0两类结果均缺；不能把默认0视为成功。 | 180 | `43d0ff427160681cfe869144a8ff9530387be78b8cc84bfa68f9d87e28d4a822` |
| N28-TFA-N-DUPRESULT | V24-SGD-MT-A | V24-SGD-MT-R | 正常E0基础2001重复，即使相同也不后写覆盖。 | 204 | `79c435278671b9d26ddd8addc512455561bf2c8379cec204ddd2c71283c171a1` |
| N28-TFA-N-SESSION | V24-SGD-MT-A | V24-SGD-MT-R | 头关联完整命中，但Session不等于原请求；Exchange违规，ErrInvalidAnswer。 | 196 | `d916a710135313bcc3506a6d4f783c98d7f47adab08e925c6311bad24fe57cb5` |
| N28-TFA-N-NOHOST | V24-SGD-MT-A | V24-SGD-MT-R | 正常E0缺Origin-Host；不能用Destination/Peer身份补。 | 172 | `5da5fd4554a89dee569c8b0a9ad431f9ffdd82e9ab94fd542a0a3a33829ca46c` |
| N28-TFA-N-NOREALM | V24-SGD-MT-A | V24-SGD-MT-R | 正常E0缺Origin-Realm；按已采用4887及字段语法检查。 | 172 | `d4109cd866bc89bcb5cbf4b79a1097d42c31cf6506c179138d9fa432261b2a26` |
| N28-TFA-N-NOSTATE | V24-SGD-MT-A | V24-SGD-MT-R | 正常E0缺本应用必需Auth-Session-State；不能套E语法放行。 | 180 | `d290e5b05a1269b67a8479ef5a54c97dd056fb7ae2648db34795b46fd18f4e81` |
| N28-TFA-N-DUAL | V24-SGD-MT-A | V24-SGD-MT-R | 正常E0基础2001与实验10415/5001并存；不产生业务成功。ALA本身亦无此已选应用失败分支。 | 224 | `7a041b4307f6c7da8a065626eaae0dfee82e7c4a8a3c4cec738c410be66314e8` |
| N28-TFA-E-NOSESSION | E28-TFA-3002 | V24-SGD-MT-R | 原请求有Session，E错误不能凭0..1语法省略；完整头命中后ErrInvalidAnswer。 | 72 | `40c631db74cc20e3e8a5fbdcefc6f51442b3f01c2327921d996dfb47b261dd16` |
| N28-TFA-E-NORESULT | E28-TFA-3002 | V24-SGD-MT-R | E1缺必需基础Result-Code；不从实验结果或E位猜码值。 | 88 | `303568823a273a27414e683e0dc3d34948b457b541483b742f4e6818d34ab2e3` |
| N28-TFA-E-E2E | E28-TFA-3002 | V24-SGD-MT-R | HBH候选命中但E2E不同；完整头未关联，有界丢弃，不消费原pending的A/完成权。 | 100 | `fd4fa7a26f22a2c720bbb061c0957a3f07ff41e685aea90b0c7122cf3d8a17f9` |
| N28-UDA-E-SUCCESS | E28-UDA-3002 | V24-SH-UD-R | E1不能携带本基础结果类别；不是把所有5xxx一并拒绝。 | 96 | `d04fdfb58aec7bd4d8ee6acf519cd2542b51b17a4e952446567a160dc9763cf8` |
| N28-UDA-E-TRANSIENT | E28-UDA-3002 | V24-SH-UD-R | E1不能携带本基础结果类别；不是把所有5xxx一并拒绝。 | 96 | `ef1b925cba609dccbc43a43ffe3c837d33726deb968614580b628f4494a0fc3d` |
| N28-UDA-E-DUPRESULT | E28-UDA-3002 | V24-SH-UD-R | E1基础结果只一次。 | 108 | `c1762eeb569ae2862fdf3a3d0d80f2c672529efdcae855321d7579ca5699a004` |
| N28-UDA-E-EXPERIMENT-ONLY | E28-UDA-3002 | V24-SH-UD-R | 只有Experimental-Result，不能满足E1必需基础结果。 | 116 | `80a4ec2f9a1032f717cac273546f7f01b7fd13d773a7cef4fea9a3334bb95285` |
| N28-UDA-E-NOHOST | E28-UDA-3002 | V24-SH-UD-R | E1必需公共字段的缺失/次数/宽度，见ID；不套业务成功载荷规则。 | 72 | `04059ffbe4c7dd656fe6270eec5b74cc79e435578169ddb73d89d02775510d39` |
| N28-UDA-E-NOREALM | E28-UDA-3002 | V24-SH-UD-R | E1必需公共字段的缺失/次数/宽度，见ID；不套业务成功载荷规则。 | 80 | `cdf0e1c8f07cba8974cbdf4460b03ff1379984ccf66d6c33732a3f8fcb804dc5` |
| N28-UDA-E-RESULT3 | E28-UDA-3002 | V24-SH-UD-R | E1必需公共字段的缺失/次数/宽度，见ID；不套业务成功载荷规则。 | 96 | `bb5ca7c054d5f61d20bf88de07e5b5f78d27f268b2b7ea05a97b6b7ec1b946e6` |
| N28-UDA-E-SESSION-DUP | E28-UDA-3002 | V24-SH-UD-R | E1必需公共字段的缺失/次数/宽度，见ID；不套业务成功载荷规则。 | 120 | `358dc1fdf3c01bb14240cb3a40d1107e378ab80e1645526ab13ca33b04b79e29` |
| N28-UDA-E-SESSION-LATE | E28-UDA-3002 | V24-SH-UD-R | E错误有Session时仍在固定首位，移到最后违反固定位置。 | 96 | `4aed2eabbc9aa4edae9d405e5f6fd8514d1447e99e9cf8cefe931191874b76ae` |
| N28-UDA-E-DESTHOST | E28-UDA-3002 | V24-SH-UD-R | 答案禁止Destination字段；E错误也不豁免。 | 116 | `6372a4a02831a0be3df548c9824d7e80f99e728d4884939481201987a01eef26` |
| N28-UDA-E-DESTREALM | E28-UDA-3002 | V24-SH-UD-R | 答案禁止Destination字段；E错误也不豁免。 | 112 | `c7452e585a41cab9d9028c2ff51c7193332ea2de91eba8b1d93b1e4f08bcdb8c` |
| N28-UDA-E-T | E28-UDA-3002 | V24-SH-UD-R | 答案T必须清零。 | 96 | `706628cace5832890a4454d10b55749712bc285bf90cd0443eb8ac0e1c3c6630` |
| N28-UDA-E-P | E28-UDA-3002 | V24-SH-UD-R | 答案P必须与请求一致；不是收到后自动修复。 | 96 | `5d46f43e7c106d3ac334baacc539b8c0189f8131f9008f4d6f45421b8c999a9d` |
| C28-UDA-E-RESERVED | E28-UDA-3002 | V24-SH-UD-R | 保留头位发送应为0，接收按BASE-FLAGS忽略该位；不是未知E语法。 | 96 | `792f432cf639abde75c9c0b238668de8a6fc39189a7f2a5fe8ed9c5ad422a186` |
| N28-UDA-E-HBH | E28-UDA-3002 | V24-SH-UD-R | 只变HBH；完整头不属于此attempt，不能消费原pending或其答案预算。 | 96 | `bdd594583d4eada445db4fc878fef7eead1d37c0b9085f38ed841d5fc3d2dd22` |
| N28-UDA-E-APP | E28-UDA-3002 | V24-SH-UD-R | 只变APP；完整头不属于此attempt，不能消费原pending或其答案预算。 | 96 | `d6ace3ea95b08c7833114408dfc5f493a1b49b1789f8572d9072f720b0a2129e` |
| N28-UDA-E-COMMAND | E28-UDA-3002 | V24-SH-UD-R | 只变COMMAND；完整头不属于此attempt，不能消费原pending或其答案预算。 | 96 | `0dda6515a43f8f6ea730ca4b0a2e8214e9a2ec1c054b1cfcc97c032d7f821439` |
| C28-UDA-RFLAG | V24-SH-UD-A | V24-SH-UD-R | R=1必须走入站请求分类，不能因HBH相等当成答案。其body不是合法UDR，按请求拒绝路径独立处理；不终结原出站exchange。 | 416 | `873e527f9ee1e398ffb665cbbf0de20adc45f8afa4da6ba33d5504b9d03426b2` |
| N28-UDA-N-3002 | V24-SH-UD-A | V24-SH-UD-R | 所有正常字段齐全仍不能以E0发送基础协议错误3002；不只按body判断。 | 416 | `3617af73fbd1c58945176bd10f41fb82a70097446badcbe9af52d5ab3318cb07` |
| C28-UDA-N-UNKNOWNEXP | V24-SH-UD-A | V24-SH-UD-R | E0恰一个未知Vendor实验结果；保留命名空间/原值并报告未知应用含义，不制造成功。 | 436 | `b7c208ec16980bf43ebcce2f9e55290111eaf1fb03691ea3b43b09cbff9bbc0f` |
| C28-UDA-E-EXTRA-UNKNOWN | E28-UDA-3002 | V24-SH-UD-R | E语法明确列可选297，不能套正常XOR；基础3002仍为协议失败。未知补充含义保留/报告，不当新成功。 | 128 | `7d9e49f3f5c84a7bb5ef4658bd79596d0c8a45da16e4d8198ecb122a0f8eac71` |
| C28-UDA-E-EXTRA-3GPP | E28-UDA-3002 | V24-SH-UD-R | E语法明确列可选297，不能套正常XOR；基础3002仍为协议失败。Sh5102触发29.329 §6.2的Result-Code必须缺省约束；固定Profile结果校验失败，已关联输入以ErrInvalidAnswer终结。通用E语法可承载不代表Sh合法。 | 128 | `dacfa465460e5768672994256a48fe8ca70b6f023db7f1fd9de75e2a4dd061da` |
| N28-UDA-N-AGENT | V24-SH-UD-A | V24-SH-UD-R | 已授权ErrorAgent不因此成为成功Terminal；此场景TerminalOrigins排除dra-a，匹配后来源失败ErrInvalidAnswer。 | 420 | `65437ebfece12b5ea00f8d8be76ba904152acdb3029a41b34c9d732f890cbd6a` |
| N28-UDA-E-ORIGIN | E28-UDA-3002 | V24-SH-UD-R | 头匹配但Origin不属于TerminalOrigins/ErrorAgentOrigins，ErrInvalidAnswer；不授予来源。 | 92 | `9f2a0e646b628faa40b30f751ce81902e01a8b038275e71dbf61deb60c793e9c` |
| E28-SNA-5001 | F25-SH-BASE5001 | V25-SH-NEED-R | E1基础5001与原628 M1特性请求配对；一个279含原AVP。诊断内部不能再次按业务必需叶/特性拒绝。 | 204 | `b0d3e28f7ad09101e4a3b76f3e8b06ee98587e5034912b82696e740863dbc5de` |
| N28-SNA-E-NOFAILED | E28-SNA-5001 | V25-SH-NEED-R | 基础5001要求Failed-AVP；不能用281错误文本替代。 | 140 | `a5238bc24801888a1e777339f272867e0ab5e14403ace3dba254de38fe3e8fd6` |
| N28-SNA-E-DUPFAILED | E28-SNA-5001 | V25-SH-NEED-R | 按Verified4615，基础5001只一个Failed-AVP父组；多错误应放同组。 | 268 | `47ad45df52419b85756518bea3589c1724b80241d3ad02ae03819dbfc36472f5` |
| N28-SNA-E-EMPTYFAILED | E28-SNA-5001 | V25-SH-NEED-R | Failed-AVP要求至少一项证据；诊断内不重校业务不意味着空父组合法。 | 148 | `8f3b04655174094a7f5e7aacaf07b08e4b3453ed92550a6a0dab24efad37a1eb` |
| V28-UDR-PROXY-R | V24-SH-UD-R | — | 服务端本跳收到的UDR含两个Proxy-Info；保持原顺序和opaque State。 | 320 | `94fb11f4daa674f4f2692bf259eef1a951accf7c0e515552030f78bb3a92fb0f` |
| V28-UDA-PROXY-A | V24-SH-UD-A | V28-UDR-PROXY-R | 终端生成正常答案复制两个Proxy-Info；不是库充当DRA。 | 504 | `19e2162c6fb9bed80f867a7e7955dab7bbd0417bf34cf4573b948d1a877856d6` |
| E28-UDA-PROXY-A | E28-UDA-5012 | V28-UDR-PROXY-R | 终端生成E错误同样复制两个Proxy-Info。 | 184 | `3128b754fdc8bd3c8ad4138bed88e81b7fafc7da54b7680e89f2b0933faf8ce0` |
| N28-UDA-N-PROXYORDER | V28-UDA-PROXY-A | V28-UDR-PROXY-R | 两个父组各自合法但原请求顺序颠倒；Exchange回显条件失败。 | 504 | `d48bacbadd73295e8f9468107572537c83ebe4b90ec8d4b7d5fdbb66c237b7a1` |
| N28-UDA-E-PROXYORDER | E28-UDA-PROXY-A | V28-UDR-PROXY-R | 两个父组各自合法但原请求顺序颠倒；Exchange回显条件失败。 | 184 | `54fbc4ee53ae52ec4c6615b8b0e6ed5906a40ae5d3aafb33616ac9a1ee6dea9a` |
| E28-PNA-5004 | V24-SH-PN-A | N27-PNR-PUBLIC-UTF8 | 原PNR的700/601含非法UTF8；E5004只在Failed-AVP中原样引述它，不按正常User-Identity再次拒绝此答案。 | 132 | `38a50d4d562da7b0bd02f4021efe9ca1a3d57243f991c9f5211c176b8fc0eda3` |
| N28-PNA-E-BADTEXT | E28-PNA-5004 | N27-PNR-PUBLIC-UTF8 | 同样非法UTF8放在诊断容器外的Error-Message必须报告；Failed-AVP例外不能扩到整个答案。 | 144 | `4a75afc9e2e55de8af5a9dad1df40a6642c78cfde6fecf216cb8019cee033463` |

## 6. 完整hex

### E28-UDA-3002

E错误控制；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### E28-UDA-5012

E错误控制；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 01 0c 40 00 00 0c 00 00 13 94
```

### N28-UDA-N-NORESULT

无效答案；404字节；关联及预期见§5。

```message-hex
01 00 01 94 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 be c0 00 01 11 00 00 28 af 3c 53 68 2d
44 61 74 61 3e 3c 53 68 2d 49 4d 53 2d 44 61 74
61 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 45 78
74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73 69
6f 6e 3e 3c 53 4d 53 52 65 67 69 73 74 72 61 74
69 6f 6e 49 6e 66 6f 3e 3c 49 50 2d 53 4d 2d 47
57 2d 4e 75 6d 62 65 72 3e 31 32 30 32 35 35 35
30 39 39 39 3c 2f 49 50 2d 53 4d 2d 47 57 2d 4e
75 6d 62 65 72 3e 3c 45 78 74 65 6e 73 69 6f 6e
3e 3c 53 43 41 64 64 72 65 73 73 3e 31 32 30 32
35 35 35 30 31 30 30 3c 2f 53 43 41 64 64 72 65
73 73 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c
2f 53 4d 53 52 65 67 69 73 74 72 61 74 69 6f 6e
49 6e 66 6f 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e
3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 53 68 2d 49 4d
53 2d 44 61 74 61 3e 3c 2f 53 68 2d 44 61 74 61
3e 00 00 00
```

### N28-UDA-N-DUPRESULT

无效答案；428字节；关联及预期见§5。

```message-hex
01 00 01 ac 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 0c 40 00 00 0c
00 00 07 d1 00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d
70 6c 65 00 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00 00 00 02 be c0 00 01 11
00 00 28 af 3c 53 68 2d 44 61 74 61 3e 3c 53 68
2d 49 4d 53 2d 44 61 74 61 3e 3c 45 78 74 65 6e
73 69 6f 6e 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e
3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 53 4d 53 52
65 67 69 73 74 72 61 74 69 6f 6e 49 6e 66 6f 3e
3c 49 50 2d 53 4d 2d 47 57 2d 4e 75 6d 62 65 72
3e 31 32 30 32 35 35 35 30 39 39 39 3c 2f 49 50
2d 53 4d 2d 47 57 2d 4e 75 6d 62 65 72 3e 3c 45
78 74 65 6e 73 69 6f 6e 3e 3c 53 43 41 64 64 72
65 73 73 3e 31 32 30 32 35 35 35 30 31 30 30 3c
2f 53 43 41 64 64 72 65 73 73 3e 3c 2f 45 78 74
65 6e 73 69 6f 6e 3e 3c 2f 53 4d 53 52 65 67 69
73 74 72 61 74 69 6f 6e 49 6e 66 6f 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 45 78 74 65 6e
73 69 6f 6e 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e
3e 3c 2f 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c
2f 53 68 2d 44 61 74 61 3e 00 00 00
```

### N28-UDA-N-SESSION

无效答案；424字节；关联及预期见§5。

```message-hex
01 00 01 a8 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 1d 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 2d 77 72 6f 6e
67 00 00 00 00 00 01 04 40 00 00 20 00 00 01 0a
40 00 00 0c 00 00 28 af 00 00 01 02 40 00 00 0c
01 00 00 01 00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 02 be c0 00 01 11 00 00 28 af
3c 53 68 2d 44 61 74 61 3e 3c 53 68 2d 49 4d 53
2d 44 61 74 61 3e 3c 45 78 74 65 6e 73 69 6f 6e
3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 45 78 74
65 6e 73 69 6f 6e 3e 3c 53 4d 53 52 65 67 69 73
74 72 61 74 69 6f 6e 49 6e 66 6f 3e 3c 49 50 2d
53 4d 2d 47 57 2d 4e 75 6d 62 65 72 3e 31 32 30
32 35 35 35 30 39 39 39 3c 2f 49 50 2d 53 4d 2d
47 57 2d 4e 75 6d 62 65 72 3e 3c 45 78 74 65 6e
73 69 6f 6e 3e 3c 53 43 41 64 64 72 65 73 73 3e
31 32 30 32 35 35 35 30 31 30 30 3c 2f 53 43 41
64 64 72 65 73 73 3e 3c 2f 45 78 74 65 6e 73 69
6f 6e 3e 3c 2f 53 4d 53 52 65 67 69 73 74 72 61
74 69 6f 6e 49 6e 66 6f 3e 3c 2f 45 78 74 65 6e
73 69 6f 6e 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e
3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 53
68 2d 49 4d 53 2d 44 61 74 61 3e 3c 2f 53 68 2d
44 61 74 61 3e 00 00 00
```

### N28-UDA-N-NOHOST

无效答案；396字节；关联及预期见§5。

```message-hex
01 00 01 8c 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00 00 00 02 be c0 00 01 11
00 00 28 af 3c 53 68 2d 44 61 74 61 3e 3c 53 68
2d 49 4d 53 2d 44 61 74 61 3e 3c 45 78 74 65 6e
73 69 6f 6e 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e
3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 53 4d 53 52
65 67 69 73 74 72 61 74 69 6f 6e 49 6e 66 6f 3e
3c 49 50 2d 53 4d 2d 47 57 2d 4e 75 6d 62 65 72
3e 31 32 30 32 35 35 35 30 39 39 39 3c 2f 49 50
2d 53 4d 2d 47 57 2d 4e 75 6d 62 65 72 3e 3c 45
78 74 65 6e 73 69 6f 6e 3e 3c 53 43 41 64 64 72
65 73 73 3e 31 32 30 32 35 35 35 30 31 30 30 3c
2f 53 43 41 64 64 72 65 73 73 3e 3c 2f 45 78 74
65 6e 73 69 6f 6e 3e 3c 2f 53 4d 53 52 65 67 69
73 74 72 61 74 69 6f 6e 49 6e 66 6f 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 45 78 74 65 6e
73 69 6f 6e 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e
3e 3c 2f 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c
2f 53 68 2d 44 61 74 61 3e 00 00 00
```

### N28-UDA-N-NOREALM

无效答案；396字节；关联及预期见§5。

```message-hex
01 00 01 8c 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 02 be c0 00 01 11
00 00 28 af 3c 53 68 2d 44 61 74 61 3e 3c 53 68
2d 49 4d 53 2d 44 61 74 61 3e 3c 45 78 74 65 6e
73 69 6f 6e 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e
3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 53 4d 53 52
65 67 69 73 74 72 61 74 69 6f 6e 49 6e 66 6f 3e
3c 49 50 2d 53 4d 2d 47 57 2d 4e 75 6d 62 65 72
3e 31 32 30 32 35 35 35 30 39 39 39 3c 2f 49 50
2d 53 4d 2d 47 57 2d 4e 75 6d 62 65 72 3e 3c 45
78 74 65 6e 73 69 6f 6e 3e 3c 53 43 41 64 64 72
65 73 73 3e 31 32 30 32 35 35 35 30 31 30 30 3c
2f 53 43 41 64 64 72 65 73 73 3e 3c 2f 45 78 74
65 6e 73 69 6f 6e 3e 3c 2f 53 4d 53 52 65 67 69
73 74 72 61 74 69 6f 6e 49 6e 66 6f 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 45 78 74 65 6e
73 69 6f 6e 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e
3e 3c 2f 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c
2f 53 68 2d 44 61 74 61 3e 00 00 00
```

### N28-UDA-N-NOSTATE

无效答案；404字节；关联及预期见§5。

```message-hex
01 00 01 94 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 be c0 00 01 11 00 00 28 af 3c 53 68 2d
44 61 74 61 3e 3c 53 68 2d 49 4d 53 2d 44 61 74
61 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 45 78
74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73 69
6f 6e 3e 3c 53 4d 53 52 65 67 69 73 74 72 61 74
69 6f 6e 49 6e 66 6f 3e 3c 49 50 2d 53 4d 2d 47
57 2d 4e 75 6d 62 65 72 3e 31 32 30 32 35 35 35
30 39 39 39 3c 2f 49 50 2d 53 4d 2d 47 57 2d 4e
75 6d 62 65 72 3e 3c 45 78 74 65 6e 73 69 6f 6e
3e 3c 53 43 41 64 64 72 65 73 73 3e 31 32 30 32
35 35 35 30 31 30 30 3c 2f 53 43 41 64 64 72 65
73 73 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c
2f 53 4d 53 52 65 67 69 73 74 72 61 74 69 6f 6e
49 6e 66 6f 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e
3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 53 68 2d 49 4d
53 2d 44 61 74 61 3e 3c 2f 53 68 2d 44 61 74 61
3e 00 00 00
```

### N28-UDA-N-DUAL

无效答案；448字节；关联及预期见§5。

```message-hex
01 00 01 c0 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 02 be
c0 00 01 11 00 00 28 af 3c 53 68 2d 44 61 74 61
3e 3c 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c 45
78 74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73
69 6f 6e 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c
53 4d 53 52 65 67 69 73 74 72 61 74 69 6f 6e 49
6e 66 6f 3e 3c 49 50 2d 53 4d 2d 47 57 2d 4e 75
6d 62 65 72 3e 31 32 30 32 35 35 35 30 39 39 39
3c 2f 49 50 2d 53 4d 2d 47 57 2d 4e 75 6d 62 65
72 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 53 43
41 64 64 72 65 73 73 3e 31 32 30 32 35 35 35 30
31 30 30 3c 2f 53 43 41 64 64 72 65 73 73 3e 3c
2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 53 4d 53
52 65 67 69 73 74 72 61 74 69 6f 6e 49 6e 66 6f
3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 45 78 74 65 6e
73 69 6f 6e 3e 3c 2f 53 68 2d 49 4d 53 2d 44 61
74 61 3e 3c 2f 53 68 2d 44 61 74 61 3e 00 00 00
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c
00 00 28 af 00 00 01 2a 40 00 00 0c 00 00 13 89
```

### N28-UDA-E-NOSESSION

无效答案；72字节；关联及预期见§5。

```message-hex
01 00 00 48 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 08 40 00 00 15 64 72 61 2d
61 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 01 28
40 00 00 0f 65 78 61 6d 70 6c 65 00 00 00 01 0c
40 00 00 0c 00 00 0b ba
```

### N28-UDA-E-NORESULT

无效答案；84字节；关联及预期见§5。

```message-hex
01 00 00 54 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00
```

### N28-UDA-E-E2E

未关联输入；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 00 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### N28-UDA-N-VSAIAPP

无效答案；416字节；关联及预期见§5。

```message-hex
01 00 01 a0 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 60 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 02 be
c0 00 01 11 00 00 28 af 3c 53 68 2d 44 61 74 61
3e 3c 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c 45
78 74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73
69 6f 6e 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c
53 4d 53 52 65 67 69 73 74 72 61 74 69 6f 6e 49
6e 66 6f 3e 3c 49 50 2d 53 4d 2d 47 57 2d 4e 75
6d 62 65 72 3e 31 32 30 32 35 35 35 30 39 39 39
3c 2f 49 50 2d 53 4d 2d 47 57 2d 4e 75 6d 62 65
72 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 53 43
41 64 64 72 65 73 73 3e 31 32 30 32 35 35 35 30
31 30 30 3c 2f 53 43 41 64 64 72 65 73 73 3e 3c
2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 53 4d 53
52 65 67 69 73 74 72 61 74 69 6f 6e 49 6e 66 6f
3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 45 78 74 65 6e
73 69 6f 6e 3e 3c 2f 53 68 2d 49 4d 53 2d 44 61
74 61 3e 3c 2f 53 68 2d 44 61 74 61 3e 00 00 00
```

### E28-PUA-3002

E错误控制；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 33 01 00 00 01 24 24 00 02
34 34 00 02 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 32 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### E28-PUA-5012

E错误控制；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 33 01 00 00 01 24 24 00 02
34 34 00 02 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 32 00 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 01 0c 40 00 00 0c 00 00 13 94
```

### N28-PUA-N-NORESULT

无效答案；128字节；关联及预期见§5。

```message-hex
01 00 00 80 40 00 01 33 01 00 00 01 24 24 00 02
34 34 00 02 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 32 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### N28-PUA-N-DUPRESULT

无效答案；152字节；关联及预期见§5。

```message-hex
01 00 00 98 40 00 01 33 01 00 00 01 24 24 00 02
34 34 00 02 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 32 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 0c 40 00 00 0c
00 00 07 d1 00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d
70 6c 65 00 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00
```

### N28-PUA-N-SESSION

无效答案；148字节；关联及预期见§5。

```message-hex
01 00 00 94 40 00 01 33 01 00 00 01 24 24 00 02
34 34 00 02 00 00 01 07 40 00 00 1d 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 32 2d 77 72 6f 6e
67 00 00 00 00 00 01 04 40 00 00 20 00 00 01 0a
40 00 00 0c 00 00 28 af 00 00 01 02 40 00 00 0c
01 00 00 01 00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00
```

### N28-PUA-N-NOHOST

无效答案；120字节；关联及预期见§5。

```message-hex
01 00 00 78 40 00 01 33 01 00 00 01 24 24 00 02
34 34 00 02 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 32 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00
```

### N28-PUA-N-NOREALM

无效答案；120字节；关联及预期见§5。

```message-hex
01 00 00 78 40 00 01 33 01 00 00 01 24 24 00 02
34 34 00 02 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 32 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00
```

### N28-PUA-N-NOSTATE

无效答案；128字节；关联及预期见§5。

```message-hex
01 00 00 80 40 00 01 33 01 00 00 01 24 24 00 02
34 34 00 02 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 32 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### N28-PUA-N-DUAL

无效答案；172字节；关联及预期见§5。

```message-hex
01 00 00 ac 40 00 01 33 01 00 00 01 24 24 00 02
34 34 00 02 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 32 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 01 29
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 2a 40 00 00 0c 00 00 13 89
```

### N28-PUA-E-NOSESSION

无效答案；72字节；关联及预期见§5。

```message-hex
01 00 00 48 60 00 01 33 01 00 00 01 24 24 00 02
34 34 00 02 00 00 01 08 40 00 00 15 64 72 61 2d
61 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 01 28
40 00 00 0f 65 78 61 6d 70 6c 65 00 00 00 01 0c
40 00 00 0c 00 00 0b ba
```

### N28-PUA-E-NORESULT

无效答案；84字节；关联及预期见§5。

```message-hex
01 00 00 54 60 00 01 33 01 00 00 01 24 24 00 02
34 34 00 02 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 32 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00
```

### N28-PUA-E-E2E

未关联输入；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 33 01 00 00 01 24 24 00 02
34 34 00 03 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 32 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### N28-PUA-N-VSAIAPP

无效答案；140字节；关联及预期见§5。

```message-hex
01 00 00 8c 40 00 01 33 01 00 00 01 24 24 00 02
34 34 00 02 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 32 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 60 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### E28-SNA-3002

E错误控制；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 34 01 00 00 01 24 24 00 03
34 34 00 03 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 33 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### E28-SNA-5012

E错误控制；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 34 01 00 00 01 24 24 00 03
34 34 00 03 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 33 00 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 01 0c 40 00 00 0c 00 00 13 94
```

### N28-SNA-N-NORESULT

无效答案；232字节；关联及预期见§5。

```message-hex
01 00 00 e8 40 00 01 34 01 00 00 01 24 24 00 03
34 34 00 03 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 33 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 be c0 00 00 58 00 00 28 af 3c 53 68 2d
44 61 74 61 3e 3c 53 68 2d 49 4d 53 2d 44 61 74
61 3e 3c 49 4d 53 55 73 65 72 53 74 61 74 65 3e
31 3c 2f 49 4d 53 55 73 65 72 53 74 61 74 65 3e
3c 2f 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c 2f
53 68 2d 44 61 74 61 3e 00 00 02 c5 80 00 00 10
00 00 28 af ee 8f b5 08
```

### N28-SNA-N-DUPRESULT

无效答案；256字节；关联及预期见§5。

```message-hex
01 00 01 00 40 00 01 34 01 00 00 01 24 24 00 03
34 34 00 03 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 33 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 0c 40 00 00 0c
00 00 07 d1 00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d
70 6c 65 00 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00 00 00 02 be c0 00 00 58
00 00 28 af 3c 53 68 2d 44 61 74 61 3e 3c 53 68
2d 49 4d 53 2d 44 61 74 61 3e 3c 49 4d 53 55 73
65 72 53 74 61 74 65 3e 31 3c 2f 49 4d 53 55 73
65 72 53 74 61 74 65 3e 3c 2f 53 68 2d 49 4d 53
2d 44 61 74 61 3e 3c 2f 53 68 2d 44 61 74 61 3e
00 00 02 c5 80 00 00 10 00 00 28 af ee 8f b5 08
```

### N28-SNA-N-SESSION

无效答案；252字节；关联及预期见§5。

```message-hex
01 00 00 fc 40 00 01 34 01 00 00 01 24 24 00 03
34 34 00 03 00 00 01 07 40 00 00 1d 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 33 2d 77 72 6f 6e
67 00 00 00 00 00 01 04 40 00 00 20 00 00 01 0a
40 00 00 0c 00 00 28 af 00 00 01 02 40 00 00 0c
01 00 00 01 00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 02 be c0 00 00 58 00 00 28 af
3c 53 68 2d 44 61 74 61 3e 3c 53 68 2d 49 4d 53
2d 44 61 74 61 3e 3c 49 4d 53 55 73 65 72 53 74
61 74 65 3e 31 3c 2f 49 4d 53 55 73 65 72 53 74
61 74 65 3e 3c 2f 53 68 2d 49 4d 53 2d 44 61 74
61 3e 3c 2f 53 68 2d 44 61 74 61 3e 00 00 02 c5
80 00 00 10 00 00 28 af ee 8f b5 08
```

### N28-SNA-N-NOHOST

无效答案；224字节；关联及预期见§5。

```message-hex
01 00 00 e0 40 00 01 34 01 00 00 01 24 24 00 03
34 34 00 03 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 33 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00 00 00 02 be c0 00 00 58
00 00 28 af 3c 53 68 2d 44 61 74 61 3e 3c 53 68
2d 49 4d 53 2d 44 61 74 61 3e 3c 49 4d 53 55 73
65 72 53 74 61 74 65 3e 31 3c 2f 49 4d 53 55 73
65 72 53 74 61 74 65 3e 3c 2f 53 68 2d 49 4d 53
2d 44 61 74 61 3e 3c 2f 53 68 2d 44 61 74 61 3e
00 00 02 c5 80 00 00 10 00 00 28 af ee 8f b5 08
```

### N28-SNA-N-NOREALM

无效答案；224字节；关联及预期见§5。

```message-hex
01 00 00 e0 40 00 01 34 01 00 00 01 24 24 00 03
34 34 00 03 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 33 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 02 be c0 00 00 58
00 00 28 af 3c 53 68 2d 44 61 74 61 3e 3c 53 68
2d 49 4d 53 2d 44 61 74 61 3e 3c 49 4d 53 55 73
65 72 53 74 61 74 65 3e 31 3c 2f 49 4d 53 55 73
65 72 53 74 61 74 65 3e 3c 2f 53 68 2d 49 4d 53
2d 44 61 74 61 3e 3c 2f 53 68 2d 44 61 74 61 3e
00 00 02 c5 80 00 00 10 00 00 28 af ee 8f b5 08
```

### N28-SNA-N-NOSTATE

无效答案；232字节；关联及预期见§5。

```message-hex
01 00 00 e8 40 00 01 34 01 00 00 01 24 24 00 03
34 34 00 03 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 33 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 be c0 00 00 58 00 00 28 af 3c 53 68 2d
44 61 74 61 3e 3c 53 68 2d 49 4d 53 2d 44 61 74
61 3e 3c 49 4d 53 55 73 65 72 53 74 61 74 65 3e
31 3c 2f 49 4d 53 55 73 65 72 53 74 61 74 65 3e
3c 2f 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c 2f
53 68 2d 44 61 74 61 3e 00 00 02 c5 80 00 00 10
00 00 28 af ee 8f b5 08
```

### N28-SNA-N-DUAL

无效答案；276字节；关联及预期见§5。

```message-hex
01 00 01 14 40 00 01 34 01 00 00 01 24 24 00 03
34 34 00 03 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 33 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 02 be
c0 00 00 58 00 00 28 af 3c 53 68 2d 44 61 74 61
3e 3c 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c 49
4d 53 55 73 65 72 53 74 61 74 65 3e 31 3c 2f 49
4d 53 55 73 65 72 53 74 61 74 65 3e 3c 2f 53 68
2d 49 4d 53 2d 44 61 74 61 3e 3c 2f 53 68 2d 44
61 74 61 3e 00 00 02 c5 80 00 00 10 00 00 28 af
ee 8f b5 08 00 00 01 29 40 00 00 20 00 00 01 0a
40 00 00 0c 00 00 28 af 00 00 01 2a 40 00 00 0c
00 00 13 89
```

### N28-SNA-E-NOSESSION

无效答案；72字节；关联及预期见§5。

```message-hex
01 00 00 48 60 00 01 34 01 00 00 01 24 24 00 03
34 34 00 03 00 00 01 08 40 00 00 15 64 72 61 2d
61 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 01 28
40 00 00 0f 65 78 61 6d 70 6c 65 00 00 00 01 0c
40 00 00 0c 00 00 0b ba
```

### N28-SNA-E-NORESULT

无效答案；84字节；关联及预期见§5。

```message-hex
01 00 00 54 60 00 01 34 01 00 00 01 24 24 00 03
34 34 00 03 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 33 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00
```

### N28-SNA-E-E2E

未关联输入；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 34 01 00 00 01 24 24 00 03
34 34 00 02 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 33 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### N28-SNA-N-VSAIAPP

无效答案；244字节；关联及预期见§5。

```message-hex
01 00 00 f4 40 00 01 34 01 00 00 01 24 24 00 03
34 34 00 03 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 33 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 60 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 02 be
c0 00 00 58 00 00 28 af 3c 53 68 2d 44 61 74 61
3e 3c 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c 49
4d 53 55 73 65 72 53 74 61 74 65 3e 31 3c 2f 49
4d 53 55 73 65 72 53 74 61 74 65 3e 3c 2f 53 68
2d 49 4d 53 2d 44 61 74 61 3e 3c 2f 53 68 2d 44
61 74 61 3e 00 00 02 c5 80 00 00 10 00 00 28 af
ee 8f b5 08
```

### E28-PNA-3002

E错误控制；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### E28-PNA-5012

E错误控制；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 08
40 00 00 12 61 73 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 01 0c 40 00 00 0c 00 00 13 94
```

### N28-PNA-N-NORESULT

无效答案；128字节；关联及预期见§5。

```message-hex
01 00 00 80 40 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 12
61 73 2e 65 78 61 6d 70 6c 65 00 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### N28-PNA-N-DUPRESULT

无效答案；152字节；关联及预期见§5。

```message-hex
01 00 00 98 40 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 0c 40 00 00 0c
00 00 07 d1 00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 12 61 73 2e 65 78 61 6d 70
6c 65 00 00 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00
```

### N28-PNA-N-SESSION

无效答案；148字节；关联及预期见§5。

```message-hex
01 00 00 94 40 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 1e 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 2d 77 72 6f
6e 67 00 00 00 00 01 04 40 00 00 20 00 00 01 0a
40 00 00 0c 00 00 28 af 00 00 01 02 40 00 00 0c
01 00 00 01 00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 12 61 73 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00
```

### N28-PNA-N-NOHOST

无效答案；120字节；关联及预期见§5。

```message-hex
01 00 00 78 40 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00
```

### N28-PNA-N-NOREALM

无效答案；120字节；关联及预期见§5。

```message-hex
01 00 00 78 40 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 12 61 73 2e 65
78 61 6d 70 6c 65 00 00
```

### N28-PNA-N-NOSTATE

无效答案；128字节；关联及预期见§5。

```message-hex
01 00 00 80 40 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 08 40 00 00 12
61 73 2e 65 78 61 6d 70 6c 65 00 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### N28-PNA-N-DUAL

无效答案；172字节；关联及预期见§5。

```message-hex
01 00 00 ac 40 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 12 61 73 2e 65
78 61 6d 70 6c 65 00 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 01 29
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 2a 40 00 00 0c 00 00 13 89
```

### N28-PNA-E-NOSESSION

无效答案；72字节；关联及预期见§5。

```message-hex
01 00 00 48 60 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 08 40 00 00 15 64 72 61 2d
61 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 01 28
40 00 00 0f 65 78 61 6d 70 6c 65 00 00 00 01 0c
40 00 00 0c 00 00 0b ba
```

### N28-PNA-E-NORESULT

无效答案；84字节；关联及预期见§5。

```message-hex
01 00 00 54 60 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00
```

### N28-PNA-E-E2E

未关联输入；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 35 01 00 00 01 24 24 00 04
34 34 00 05 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### N28-PNA-N-VSAIAPP

无效答案；140字节；关联及预期见§5。

```message-hex
01 00 00 8c 40 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 60 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 12 61 73 2e 65
78 61 6d 70 6c 65 00 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### E28-SRA-3002

E错误控制；100字节；关联及预期见§5。

```message-hex
01 00 00 64 60 80 00 27 01 00 00 60 25 25 00 20
35 35 00 20 00 00 01 07 40 00 00 1a 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 35 3b 33 32 00 00
00 00 01 08 40 00 00 15 64 72 61 2d 61 2e 65 78
61 6d 70 6c 65 00 00 00 00 00 01 28 40 00 00 0f
65 78 61 6d 70 6c 65 00 00 00 01 0c 40 00 00 0c
00 00 0b ba
```

### E28-SRA-5012

E错误控制；100字节；关联及预期见§5。

```message-hex
01 00 00 64 60 80 00 27 01 00 00 60 25 25 00 20
35 35 00 20 00 00 01 07 40 00 00 1a 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 35 3b 33 32 00 00
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d
70 6c 65 00 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00 00 00 01 0c 40 00 00 0c
00 00 13 94
```

### N28-SRA-N-NORESULT

无效答案；252字节；关联及预期见§5。

```message-hex
01 00 00 fc 40 80 00 27 01 00 00 60 25 25 00 20
35 35 00 20 00 00 01 07 40 00 00 1a 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 35 3b 33 32 00 00
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 00 01 40 00 00 17 30 30 31 30
31 30 31 32 33 34 35 36 37 38 39 00 00 00 09 61
c0 00 00 20 00 00 28 af 00 00 05 d1 c0 00 00 12
00 00 28 af 21 20 55 05 77 f7 00 00 00 00 09 66
c0 00 00 20 00 00 28 af 00 00 09 63 c0 00 00 12
00 00 28 af 21 20 55 05 66 f6 00 00 00 00 09 60
c0 00 00 10 00 00 28 af 01 23 45 67 00 00 0c 1e
c0 00 00 20 00 00 28 af 00 00 02 bd c0 00 00 12
00 00 28 af 21 20 55 05 21 f4 00 00 00 00 0c f0
c0 00 00 10 00 00 28 af 00 00 00 00
```

### N28-SRA-N-DUPRESULT

无效答案；276字节；关联及预期见§5。

```message-hex
01 00 01 14 40 80 00 27 01 00 00 60 25 25 00 20
35 35 00 20 00 00 01 07 40 00 00 1a 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 35 3b 33 32 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 00 01
40 00 00 17 30 30 31 30 31 30 31 32 33 34 35 36
37 38 39 00 00 00 09 61 c0 00 00 20 00 00 28 af
00 00 05 d1 c0 00 00 12 00 00 28 af 21 20 55 05
77 f7 00 00 00 00 09 66 c0 00 00 20 00 00 28 af
00 00 09 63 c0 00 00 12 00 00 28 af 21 20 55 05
66 f6 00 00 00 00 09 60 c0 00 00 10 00 00 28 af
01 23 45 67 00 00 0c 1e c0 00 00 20 00 00 28 af
00 00 02 bd c0 00 00 12 00 00 28 af 21 20 55 05
21 f4 00 00 00 00 0c f0 c0 00 00 10 00 00 28 af
00 00 00 00
```

### N28-SRA-N-SESSION

无效答案；268字节；关联及预期见§5。

```message-hex
01 00 01 0c 40 80 00 27 01 00 00 60 25 25 00 20
35 35 00 20 00 00 01 07 40 00 00 20 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 35 3b 33 32 2d 77
72 6f 6e 67 00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 00 01 40 00 00 17 30 30 31 30
31 30 31 32 33 34 35 36 37 38 39 00 00 00 09 61
c0 00 00 20 00 00 28 af 00 00 05 d1 c0 00 00 12
00 00 28 af 21 20 55 05 77 f7 00 00 00 00 09 66
c0 00 00 20 00 00 28 af 00 00 09 63 c0 00 00 12
00 00 28 af 21 20 55 05 66 f6 00 00 00 00 09 60
c0 00 00 10 00 00 28 af 01 23 45 67 00 00 0c 1e
c0 00 00 20 00 00 28 af 00 00 02 bd c0 00 00 12
00 00 28 af 21 20 55 05 21 f4 00 00 00 00 0c f0
c0 00 00 10 00 00 28 af 00 00 00 00
```

### N28-SRA-N-NOHOST

无效答案；244字节；关联及预期见§5。

```message-hex
01 00 00 f4 40 80 00 27 01 00 00 60 25 25 00 20
35 35 00 20 00 00 01 07 40 00 00 1a 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 35 3b 33 32 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 00 01
40 00 00 17 30 30 31 30 31 30 31 32 33 34 35 36
37 38 39 00 00 00 09 61 c0 00 00 20 00 00 28 af
00 00 05 d1 c0 00 00 12 00 00 28 af 21 20 55 05
77 f7 00 00 00 00 09 66 c0 00 00 20 00 00 28 af
00 00 09 63 c0 00 00 12 00 00 28 af 21 20 55 05
66 f6 00 00 00 00 09 60 c0 00 00 10 00 00 28 af
01 23 45 67 00 00 0c 1e c0 00 00 20 00 00 28 af
00 00 02 bd c0 00 00 12 00 00 28 af 21 20 55 05
21 f4 00 00 00 00 0c f0 c0 00 00 10 00 00 28 af
00 00 00 00
```

### N28-SRA-N-NOREALM

无效答案；244字节；关联及预期见§5。

```message-hex
01 00 00 f4 40 80 00 27 01 00 00 60 25 25 00 20
35 35 00 20 00 00 01 07 40 00 00 1a 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 35 3b 33 32 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 00 01
40 00 00 17 30 30 31 30 31 30 31 32 33 34 35 36
37 38 39 00 00 00 09 61 c0 00 00 20 00 00 28 af
00 00 05 d1 c0 00 00 12 00 00 28 af 21 20 55 05
77 f7 00 00 00 00 09 66 c0 00 00 20 00 00 28 af
00 00 09 63 c0 00 00 12 00 00 28 af 21 20 55 05
66 f6 00 00 00 00 09 60 c0 00 00 10 00 00 28 af
01 23 45 67 00 00 0c 1e c0 00 00 20 00 00 28 af
00 00 02 bd c0 00 00 12 00 00 28 af 21 20 55 05
21 f4 00 00 00 00 0c f0 c0 00 00 10 00 00 28 af
00 00 00 00
```

### N28-SRA-N-NOSTATE

无效答案；252字节；关联及预期见§5。

```message-hex
01 00 00 fc 40 80 00 27 01 00 00 60 25 25 00 20
35 35 00 20 00 00 01 07 40 00 00 1a 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 35 3b 33 32 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 00 01 40 00 00 17 30 30 31 30
31 30 31 32 33 34 35 36 37 38 39 00 00 00 09 61
c0 00 00 20 00 00 28 af 00 00 05 d1 c0 00 00 12
00 00 28 af 21 20 55 05 77 f7 00 00 00 00 09 66
c0 00 00 20 00 00 28 af 00 00 09 63 c0 00 00 12
00 00 28 af 21 20 55 05 66 f6 00 00 00 00 09 60
c0 00 00 10 00 00 28 af 01 23 45 67 00 00 0c 1e
c0 00 00 20 00 00 28 af 00 00 02 bd c0 00 00 12
00 00 28 af 21 20 55 05 21 f4 00 00 00 00 0c f0
c0 00 00 10 00 00 28 af 00 00 00 00
```

### N28-SRA-N-DUAL

无效答案；296字节；关联及预期见§5。

```message-hex
01 00 01 28 40 80 00 27 01 00 00 60 25 25 00 20
35 35 00 20 00 00 01 07 40 00 00 1a 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 35 3b 33 32 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 00 01 40 00 00 17 30 30 31 30 31 30 31 32
33 34 35 36 37 38 39 00 00 00 09 61 c0 00 00 20
00 00 28 af 00 00 05 d1 c0 00 00 12 00 00 28 af
21 20 55 05 77 f7 00 00 00 00 09 66 c0 00 00 20
00 00 28 af 00 00 09 63 c0 00 00 12 00 00 28 af
21 20 55 05 66 f6 00 00 00 00 09 60 c0 00 00 10
00 00 28 af 01 23 45 67 00 00 0c 1e c0 00 00 20
00 00 28 af 00 00 02 bd c0 00 00 12 00 00 28 af
21 20 55 05 21 f4 00 00 00 00 0c f0 c0 00 00 10
00 00 28 af 00 00 00 00 00 00 01 29 40 00 00 20
00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 2a
40 00 00 0c 00 00 13 89
```

### N28-SRA-E-NOSESSION

无效答案；72字节；关联及预期见§5。

```message-hex
01 00 00 48 60 80 00 27 01 00 00 60 25 25 00 20
35 35 00 20 00 00 01 08 40 00 00 15 64 72 61 2d
61 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 01 28
40 00 00 0f 65 78 61 6d 70 6c 65 00 00 00 01 0c
40 00 00 0c 00 00 0b ba
```

### N28-SRA-E-NORESULT

无效答案；88字节；关联及预期见§5。

```message-hex
01 00 00 58 60 80 00 27 01 00 00 60 25 25 00 20
35 35 00 20 00 00 01 07 40 00 00 1a 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 35 3b 33 32 00 00
00 00 01 08 40 00 00 15 64 72 61 2d 61 2e 65 78
61 6d 70 6c 65 00 00 00 00 00 01 28 40 00 00 0f
65 78 61 6d 70 6c 65 00
```

### N28-SRA-E-E2E

未关联输入；100字节；关联及预期见§5。

```message-hex
01 00 00 64 60 80 00 27 01 00 00 60 25 25 00 20
35 35 00 21 00 00 01 07 40 00 00 1a 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 35 3b 33 32 00 00
00 00 01 08 40 00 00 15 64 72 61 2d 61 2e 65 78
61 6d 70 6c 65 00 00 00 00 00 01 28 40 00 00 0f
65 78 61 6d 70 6c 65 00 00 00 01 0c 40 00 00 0c
00 00 0b ba
```

### E28-RDA-3002

E错误控制；100字节；关联及预期见§5。

```message-hex
01 00 00 64 60 80 00 29 01 00 00 60 24 24 00 06
34 34 00 06 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 36 00 00 00
00 00 01 08 40 00 00 15 64 72 61 2d 61 2e 65 78
61 6d 70 6c 65 00 00 00 00 00 01 28 40 00 00 0f
65 78 61 6d 70 6c 65 00 00 00 01 0c 40 00 00 0c
00 00 0b ba
```

### E28-RDA-5012

E错误控制；100字节；关联及预期见§5。

```message-hex
01 00 00 64 60 80 00 29 01 00 00 60 24 24 00 06
34 34 00 06 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 36 00 00 00
00 00 01 08 40 00 00 13 68 73 73 2e 65 78 61 6d
70 6c 65 00 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00 00 00 01 0c 40 00 00 0c
00 00 13 94
```

### N28-RDA-N-NORESULT

无效答案；132字节；关联及预期见§5。

```message-hex
01 00 00 84 40 80 00 29 01 00 00 60 24 24 00 06
34 34 00 06 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 36 00 00 00
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 0c 1e c0 00 00 20 00 00 28 af
00 00 02 bd c0 00 00 12 00 00 28 af 21 20 55 05
21 f4 00 00
```

### N28-RDA-N-DUPRESULT

无效答案；156字节；关联及预期见§5。

```message-hex
01 00 00 9c 40 80 00 29 01 00 00 60 24 24 00 06
34 34 00 06 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 36 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 0c 1e
c0 00 00 20 00 00 28 af 00 00 02 bd c0 00 00 12
00 00 28 af 21 20 55 05 21 f4 00 00
```

### N28-RDA-N-SESSION

无效答案；148字节；关联及预期见§5。

```message-hex
01 00 00 94 40 80 00 29 01 00 00 60 24 24 00 06
34 34 00 06 00 00 01 07 40 00 00 1f 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 36 2d 77 72
6f 6e 67 00 00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 0c 1e c0 00 00 20 00 00 28 af
00 00 02 bd c0 00 00 12 00 00 28 af 21 20 55 05
21 f4 00 00
```

### N28-RDA-N-NOHOST

无效答案；124字节；关联及预期见§5。

```message-hex
01 00 00 7c 40 80 00 29 01 00 00 60 24 24 00 06
34 34 00 06 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 36 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 0c 1e
c0 00 00 20 00 00 28 af 00 00 02 bd c0 00 00 12
00 00 28 af 21 20 55 05 21 f4 00 00
```

### N28-RDA-N-NOREALM

无效答案；124字节；关联及预期见§5。

```message-hex
01 00 00 7c 40 80 00 29 01 00 00 60 24 24 00 06
34 34 00 06 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 36 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 0c 1e
c0 00 00 20 00 00 28 af 00 00 02 bd c0 00 00 12
00 00 28 af 21 20 55 05 21 f4 00 00
```

### N28-RDA-N-NOSTATE

无效答案；132字节；关联及预期见§5。

```message-hex
01 00 00 84 40 80 00 29 01 00 00 60 24 24 00 06
34 34 00 06 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 36 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 0c 1e c0 00 00 20 00 00 28 af
00 00 02 bd c0 00 00 12 00 00 28 af 21 20 55 05
21 f4 00 00
```

### N28-RDA-N-DUAL

无效答案；176字节；关联及预期见§5。

```message-hex
01 00 00 b0 40 80 00 29 01 00 00 60 24 24 00 06
34 34 00 06 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 36 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c 1e c0 00 00 20 00 00 28 af 00 00 02 bd
c0 00 00 12 00 00 28 af 21 20 55 05 21 f4 00 00
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c
00 00 28 af 00 00 01 2a 40 00 00 0c 00 00 13 89
```

### N28-RDA-E-NOSESSION

无效答案；72字节；关联及预期见§5。

```message-hex
01 00 00 48 60 80 00 29 01 00 00 60 24 24 00 06
34 34 00 06 00 00 01 08 40 00 00 15 64 72 61 2d
61 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 01 28
40 00 00 0f 65 78 61 6d 70 6c 65 00 00 00 01 0c
40 00 00 0c 00 00 0b ba
```

### N28-RDA-E-NORESULT

无效答案；88字节；关联及预期见§5。

```message-hex
01 00 00 58 60 80 00 29 01 00 00 60 24 24 00 06
34 34 00 06 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 36 00 00 00
00 00 01 08 40 00 00 15 64 72 61 2d 61 2e 65 78
61 6d 70 6c 65 00 00 00 00 00 01 28 40 00 00 0f
65 78 61 6d 70 6c 65 00
```

### N28-RDA-E-E2E

未关联输入；100字节；关联及预期见§5。

```message-hex
01 00 00 64 60 80 00 29 01 00 00 60 24 24 00 06
34 34 00 07 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 36 00 00 00
00 00 01 08 40 00 00 15 64 72 61 2d 61 2e 65 78
61 6d 70 6c 65 00 00 00 00 00 01 28 40 00 00 0f
65 78 61 6d 70 6c 65 00 00 00 01 0c 40 00 00 0c
00 00 0b ba
```

### F28-ALA-N-5012

正常失败控制；108字节；关联及预期见§5。

```message-hex
01 00 00 6c 40 80 00 28 01 00 00 60 24 24 00 07
34 34 00 07 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 37 00 00 01 0c
40 00 00 0c 00 00 13 94 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 12 69 77 2e 65
78 61 6d 70 6c 65 00 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### E28-ALA-3002

E错误控制；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 80 00 28 01 00 00 60 24 24 00 07
34 34 00 07 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 37 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### E28-ALA-5012

E错误控制；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 80 00 28 01 00 00 60 24 24 00 07
34 34 00 07 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 37 00 00 01 08
40 00 00 12 69 77 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 01 0c 40 00 00 0c 00 00 13 94
```

### N28-ALA-N-NORESULT

无效答案；96字节；关联及预期见§5。

```message-hex
01 00 00 60 40 80 00 28 01 00 00 60 24 24 00 07
34 34 00 07 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 37 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 12
69 77 2e 65 78 61 6d 70 6c 65 00 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### N28-ALA-N-DUPRESULT

无效答案；120字节；关联及预期见§5。

```message-hex
01 00 00 78 40 80 00 28 01 00 00 60 24 24 00 07
34 34 00 07 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 37 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 0c 40 00 00 0c
00 00 07 d1 00 00 01 15 40 00 00 0c 00 00 00 01
00 00 01 08 40 00 00 12 69 77 2e 65 78 61 6d 70
6c 65 00 00 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00
```

### N28-ALA-N-SESSION

无效答案；116字节；关联及预期见§5。

```message-hex
01 00 00 74 40 80 00 28 01 00 00 60 24 24 00 07
34 34 00 07 00 00 01 07 40 00 00 1e 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 37 2d 77 72 6f
6e 67 00 00 00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 12 69 77 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00
```

### N28-ALA-N-NOHOST

无效答案；88字节；关联及预期见§5。

```message-hex
01 00 00 58 40 80 00 28 01 00 00 60 24 24 00 07
34 34 00 07 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 37 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00
```

### N28-ALA-N-NOREALM

无效答案；88字节；关联及预期见§5。

```message-hex
01 00 00 58 40 80 00 28 01 00 00 60 24 24 00 07
34 34 00 07 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 37 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 12 69 77 2e 65
78 61 6d 70 6c 65 00 00
```

### N28-ALA-N-NOSTATE

无效答案；96字节；关联及预期见§5。

```message-hex
01 00 00 60 40 80 00 28 01 00 00 60 24 24 00 07
34 34 00 07 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 37 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 08 40 00 00 12
69 77 2e 65 78 61 6d 70 6c 65 00 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
```

### N28-ALA-N-DUAL

无效答案；140字节；关联及预期见§5。

```message-hex
01 00 00 8c 40 80 00 28 01 00 00 60 24 24 00 07
34 34 00 07 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 37 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 12 69 77 2e 65
78 61 6d 70 6c 65 00 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 01 29
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 2a 40 00 00 0c 00 00 13 89
```

### N28-ALA-E-NOSESSION

无效答案；72字节；关联及预期见§5。

```message-hex
01 00 00 48 60 80 00 28 01 00 00 60 24 24 00 07
34 34 00 07 00 00 01 08 40 00 00 15 64 72 61 2d
61 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 01 28
40 00 00 0f 65 78 61 6d 70 6c 65 00 00 00 01 0c
40 00 00 0c 00 00 0b ba
```

### N28-ALA-E-NORESULT

无效答案；84字节；关联及预期见§5。

```message-hex
01 00 00 54 60 80 00 28 01 00 00 60 24 24 00 07
34 34 00 07 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 37 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00
```

### N28-ALA-E-E2E

未关联输入；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 80 00 28 01 00 00 60 24 24 00 07
34 34 00 06 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 37 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### E28-OFA-3002

E错误控制；100字节；关联及预期见§5。

```message-hex
01 00 00 64 60 80 00 25 01 00 00 61 24 24 00 0b
34 34 00 0b 00 00 01 07 40 00 00 19 6d 6d 65 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 31 31 00 00 00
00 00 01 08 40 00 00 15 64 72 61 2d 61 2e 65 78
61 6d 70 6c 65 00 00 00 00 00 01 28 40 00 00 0f
65 78 61 6d 70 6c 65 00 00 00 01 0c 40 00 00 0c
00 00 0b ba
```

### E28-OFA-5012

E错误控制；100字节；关联及预期见§5。

```message-hex
01 00 00 64 60 80 00 25 01 00 00 61 24 24 00 0b
34 34 00 0b 00 00 01 07 40 00 00 19 6d 6d 65 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 31 31 00 00 00
00 00 01 08 40 00 00 12 69 77 2e 65 78 61 6d 70
6c 65 00 00 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00 00 00 01 0c 40 00 00 0c
00 00 13 94
```

### N28-OFA-N-NORESULT

无效答案；124字节；关联及预期见§5。

```message-hex
01 00 00 7c 40 80 00 25 01 00 00 61 24 24 00 0b
34 34 00 0b 00 00 01 07 40 00 00 19 6d 6d 65 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 31 31 00 00 00
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 12 69 77 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 0c e5 c0 00 00 15 00 00 28 af
01 00 62 01 13 00 00 00 00 00 00 00
```

### N28-OFA-N-DUPRESULT

无效答案；148字节；关联及预期见§5。

```message-hex
01 00 00 94 40 80 00 25 01 00 00 61 24 24 00 0b
34 34 00 0b 00 00 01 07 40 00 00 19 6d 6d 65 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 31 31 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 12 69 77 2e 65
78 61 6d 70 6c 65 00 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 0c e5
c0 00 00 15 00 00 28 af 01 00 62 01 13 00 00 00
00 00 00 00
```

### N28-OFA-N-SESSION

无效答案；140字节；关联及预期见§5。

```message-hex
01 00 00 8c 40 80 00 25 01 00 00 61 24 24 00 0b
34 34 00 0b 00 00 01 07 40 00 00 1f 6d 6d 65 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 31 31 2d 77 72
6f 6e 67 00 00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 12 69 77 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 0c e5 c0 00 00 15 00 00 28 af
01 00 62 01 13 00 00 00 00 00 00 00
```

### N28-OFA-N-NOHOST

无效答案；116字节；关联及预期见§5。

```message-hex
01 00 00 74 40 80 00 25 01 00 00 61 24 24 00 0b
34 34 00 0b 00 00 01 07 40 00 00 19 6d 6d 65 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 31 31 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 0c e5
c0 00 00 15 00 00 28 af 01 00 62 01 13 00 00 00
00 00 00 00
```

### N28-OFA-N-NOREALM

无效答案；116字节；关联及预期见§5。

```message-hex
01 00 00 74 40 80 00 25 01 00 00 61 24 24 00 0b
34 34 00 0b 00 00 01 07 40 00 00 19 6d 6d 65 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 31 31 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 12
69 77 2e 65 78 61 6d 70 6c 65 00 00 00 00 0c e5
c0 00 00 15 00 00 28 af 01 00 62 01 13 00 00 00
00 00 00 00
```

### N28-OFA-N-NOSTATE

无效答案；124字节；关联及预期见§5。

```message-hex
01 00 00 7c 40 80 00 25 01 00 00 61 24 24 00 0b
34 34 00 0b 00 00 01 07 40 00 00 19 6d 6d 65 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 31 31 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 08
40 00 00 12 69 77 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 0c e5 c0 00 00 15 00 00 28 af
01 00 62 01 13 00 00 00 00 00 00 00
```

### N28-OFA-N-DUAL

无效答案；168字节；关联及预期见§5。

```message-hex
01 00 00 a8 40 80 00 25 01 00 00 61 24 24 00 0b
34 34 00 0b 00 00 01 07 40 00 00 19 6d 6d 65 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 31 31 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 12
69 77 2e 65 78 61 6d 70 6c 65 00 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e5 c0 00 00 15 00 00 28 af 01 00 62 01
13 00 00 00 00 00 00 00 00 00 01 29 40 00 00 20
00 00 01 0a 40 00 00 0c 00 00 28 af 00 00 01 2a
40 00 00 0c 00 00 13 89
```

### N28-OFA-E-NOSESSION

无效答案；72字节；关联及预期见§5。

```message-hex
01 00 00 48 60 80 00 25 01 00 00 61 24 24 00 0b
34 34 00 0b 00 00 01 08 40 00 00 15 64 72 61 2d
61 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 01 28
40 00 00 0f 65 78 61 6d 70 6c 65 00 00 00 01 0c
40 00 00 0c 00 00 0b ba
```

### N28-OFA-E-NORESULT

无效答案；88字节；关联及预期见§5。

```message-hex
01 00 00 58 60 80 00 25 01 00 00 61 24 24 00 0b
34 34 00 0b 00 00 01 07 40 00 00 19 6d 6d 65 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 31 31 00 00 00
00 00 01 08 40 00 00 15 64 72 61 2d 61 2e 65 78
61 6d 70 6c 65 00 00 00 00 00 01 28 40 00 00 0f
65 78 61 6d 70 6c 65 00
```

### N28-OFA-E-E2E

未关联输入；100字节；关联及预期见§5。

```message-hex
01 00 00 64 60 80 00 25 01 00 00 61 24 24 00 0b
34 34 00 0a 00 00 01 07 40 00 00 19 6d 6d 65 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 31 31 00 00 00
00 00 01 08 40 00 00 15 64 72 61 2d 61 2e 65 78
61 6d 70 6c 65 00 00 00 00 00 01 28 40 00 00 0f
65 78 61 6d 70 6c 65 00 00 00 01 0c 40 00 00 0c
00 00 0b ba
```

### E28-TFA-3002

E错误控制；100字节；关联及预期见§5。

```message-hex
01 00 00 64 60 80 00 26 01 00 00 61 24 24 00 08
34 34 00 08 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 38 00 00 00
00 00 01 08 40 00 00 15 64 72 61 2d 61 2e 65 78
61 6d 70 6c 65 00 00 00 00 00 01 28 40 00 00 0f
65 78 61 6d 70 6c 65 00 00 00 01 0c 40 00 00 0c
00 00 0b ba
```

### E28-TFA-5012

E错误控制；100字节；关联及预期见§5。

```message-hex
01 00 00 64 60 80 00 26 01 00 00 61 24 24 00 08
34 34 00 08 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 38 00 00 00
00 00 01 08 40 00 00 13 6d 6d 65 2e 65 78 61 6d
70 6c 65 00 00 00 01 28 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00 00 00 01 0c 40 00 00 0c
00 00 13 94
```

### N28-TFA-N-NORESULT

无效答案；180字节；关联及预期见§5。

```message-hex
01 00 00 b4 40 80 00 26 01 00 00 61 24 24 00 08
34 34 00 08 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 38 00 00 00
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 13 6d 6d 65 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 0c e5 c0 00 00 0e 00 00 28 af
00 00 00 00 00 00 05 d8 80 00 00 40 00 00 28 af
00 00 06 40 80 00 00 34 00 00 28 af 00 00 06 42
80 00 00 13 00 00 28 af 00 f1 10 00 00 00 01 00
00 00 06 43 80 00 00 11 00 00 28 af 00 f1 10 00
01 00 00 00
```

### N28-TFA-N-DUPRESULT

无效答案；204字节；关联及预期见§5。

```message-hex
01 00 00 cc 40 80 00 26 01 00 00 61 24 24 00 08
34 34 00 08 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 38 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 6d 6d 65 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 0c e5
c0 00 00 0e 00 00 28 af 00 00 00 00 00 00 05 d8
80 00 00 40 00 00 28 af 00 00 06 40 80 00 00 34
00 00 28 af 00 00 06 42 80 00 00 13 00 00 28 af
00 f1 10 00 00 00 01 00 00 00 06 43 80 00 00 11
00 00 28 af 00 f1 10 00 01 00 00 00
```

### N28-TFA-N-SESSION

无效答案；196字节；关联及预期见§5。

```message-hex
01 00 00 c4 40 80 00 26 01 00 00 61 24 24 00 08
34 34 00 08 00 00 01 07 40 00 00 1f 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 38 2d 77 72
6f 6e 67 00 00 00 01 0c 40 00 00 0c 00 00 07 d1
00 00 01 15 40 00 00 0c 00 00 00 01 00 00 01 08
40 00 00 13 6d 6d 65 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 0c e5 c0 00 00 0e 00 00 28 af
00 00 00 00 00 00 05 d8 80 00 00 40 00 00 28 af
00 00 06 40 80 00 00 34 00 00 28 af 00 00 06 42
80 00 00 13 00 00 28 af 00 f1 10 00 00 00 01 00
00 00 06 43 80 00 00 11 00 00 28 af 00 f1 10 00
01 00 00 00
```

### N28-TFA-N-NOHOST

无效答案；172字节；关联及预期见§5。

```message-hex
01 00 00 ac 40 80 00 26 01 00 00 61 24 24 00 08
34 34 00 08 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 38 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 0c e5
c0 00 00 0e 00 00 28 af 00 00 00 00 00 00 05 d8
80 00 00 40 00 00 28 af 00 00 06 40 80 00 00 34
00 00 28 af 00 00 06 42 80 00 00 13 00 00 28 af
00 f1 10 00 00 00 01 00 00 00 06 43 80 00 00 11
00 00 28 af 00 f1 10 00 01 00 00 00
```

### N28-TFA-N-NOREALM

无效答案；172字节；关联及预期见§5。

```message-hex
01 00 00 ac 40 80 00 26 01 00 00 61 24 24 00 08
34 34 00 08 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 38 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 13
6d 6d 65 2e 65 78 61 6d 70 6c 65 00 00 00 0c e5
c0 00 00 0e 00 00 28 af 00 00 00 00 00 00 05 d8
80 00 00 40 00 00 28 af 00 00 06 40 80 00 00 34
00 00 28 af 00 00 06 42 80 00 00 13 00 00 28 af
00 f1 10 00 00 00 01 00 00 00 06 43 80 00 00 11
00 00 28 af 00 f1 10 00 01 00 00 00
```

### N28-TFA-N-NOSTATE

无效答案；180字节；关联及预期见§5。

```message-hex
01 00 00 b4 40 80 00 26 01 00 00 61 24 24 00 08
34 34 00 08 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 38 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 08
40 00 00 13 6d 6d 65 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 0c e5 c0 00 00 0e 00 00 28 af
00 00 00 00 00 00 05 d8 80 00 00 40 00 00 28 af
00 00 06 40 80 00 00 34 00 00 28 af 00 00 06 42
80 00 00 13 00 00 28 af 00 f1 10 00 00 00 01 00
00 00 06 43 80 00 00 11 00 00 28 af 00 f1 10 00
01 00 00 00
```

### N28-TFA-N-DUAL

无效答案；224字节；关联及预期见§5。

```message-hex
01 00 00 e0 40 80 00 26 01 00 00 61 24 24 00 08
34 34 00 08 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 38 00 00 00
00 00 01 0c 40 00 00 0c 00 00 07 d1 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 13
6d 6d 65 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 0c e5 c0 00 00 0e 00 00 28 af 00 00 00 00
00 00 05 d8 80 00 00 40 00 00 28 af 00 00 06 40
80 00 00 34 00 00 28 af 00 00 06 42 80 00 00 13
00 00 28 af 00 f1 10 00 00 00 01 00 00 00 06 43
80 00 00 11 00 00 28 af 00 f1 10 00 01 00 00 00
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c
00 00 28 af 00 00 01 2a 40 00 00 0c 00 00 13 89
```

### N28-TFA-E-NOSESSION

无效答案；72字节；关联及预期见§5。

```message-hex
01 00 00 48 60 80 00 26 01 00 00 61 24 24 00 08
34 34 00 08 00 00 01 08 40 00 00 15 64 72 61 2d
61 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 01 28
40 00 00 0f 65 78 61 6d 70 6c 65 00 00 00 01 0c
40 00 00 0c 00 00 0b ba
```

### N28-TFA-E-NORESULT

无效答案；88字节；关联及预期见§5。

```message-hex
01 00 00 58 60 80 00 26 01 00 00 61 24 24 00 08
34 34 00 08 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 38 00 00 00
00 00 01 08 40 00 00 15 64 72 61 2d 61 2e 65 78
61 6d 70 6c 65 00 00 00 00 00 01 28 40 00 00 0f
65 78 61 6d 70 6c 65 00
```

### N28-TFA-E-E2E

未关联输入；100字节；关联及预期见§5。

```message-hex
01 00 00 64 60 80 00 26 01 00 00 61 24 24 00 08
34 34 00 09 00 00 01 07 40 00 00 19 67 6d 73 63
2e 65 78 61 6d 70 6c 65 3b 32 34 3b 38 00 00 00
00 00 01 08 40 00 00 15 64 72 61 2d 61 2e 65 78
61 6d 70 6c 65 00 00 00 00 00 01 28 40 00 00 0f
65 78 61 6d 70 6c 65 00 00 00 01 0c 40 00 00 0c
00 00 0b ba
```

### N28-UDA-E-SUCCESS

无效答案；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 07 d1
```

### N28-UDA-E-TRANSIENT

无效答案；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0f a1
```

### N28-UDA-E-DUPRESULT

无效答案；108字节；关联及预期见§5。

```message-hex
01 00 00 6c 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### N28-UDA-E-EXPERIMENT-ONLY

无效答案；116字节；关联及预期见§5。

```message-hex
01 00 00 74 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 29 40 00 00 20 00 00 01 0a
40 00 00 0c 00 00 28 af 00 00 01 2a 40 00 00 0c
00 00 13 89
```

### N28-UDA-E-NOHOST

无效答案；72字节；关联及预期见§5。

```message-hex
01 00 00 48 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 28
40 00 00 0f 65 78 61 6d 70 6c 65 00 00 00 01 0c
40 00 00 0c 00 00 0b ba
```

### N28-UDA-E-NOREALM

无效答案；80字节；关联及预期见§5。

```message-hex
01 00 00 50 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### N28-UDA-E-RESULT3

无效答案；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0b 00 00 00 00
```

### N28-UDA-E-SESSION-DUP

无效答案；120字节；关联及预期见§5。

```message-hex
01 00 00 78 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 07
40 00 00 17 61 73 2e 65 78 61 6d 70 6c 65 3b 32
34 3b 31 00 00 00 01 08 40 00 00 15 64 72 61 2d
61 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 01 28
40 00 00 0f 65 78 61 6d 70 6c 65 00 00 00 01 0c
40 00 00 0c 00 00 0b ba
```

### N28-UDA-E-SESSION-LATE

无效答案；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 08 40 00 00 15 64 72 61 2d
61 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 01 28
40 00 00 0f 65 78 61 6d 70 6c 65 00 00 00 01 0c
40 00 00 0c 00 00 0b ba 00 00 01 07 40 00 00 17
61 73 2e 65 78 61 6d 70 6c 65 3b 32 34 3b 31 00
```

### N28-UDA-E-DESTHOST

无效答案；116字节；关联及预期见§5。

```message-hex
01 00 00 74 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
00 00 01 25 40 00 00 13 68 73 73 2e 65 78 61 6d
70 6c 65 00
```

### N28-UDA-E-DESTREALM

无效答案；112字节；关联及预期见§5。

```message-hex
01 00 00 70 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
00 00 01 1b 40 00 00 0f 65 78 61 6d 70 6c 65 00
```

### N28-UDA-E-T

无效答案；96字节；关联及预期见§5。

```message-hex
01 00 00 60 70 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### N28-UDA-E-P

无效答案；96字节；关联及预期见§5。

```message-hex
01 00 00 60 20 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### C28-UDA-E-RESERVED

接收控制；96字节；关联及预期见§5。

```message-hex
01 00 00 60 61 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### N28-UDA-E-HBH

未关联输入；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 32 01 00 00 01 ee ee ee ee
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### N28-UDA-E-APP

未关联输入；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 32 01 00 00 60 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### N28-UDA-E-COMMAND

未关联输入；96字节；关联及预期见§5。

```message-hex
01 00 00 60 60 00 01 33 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### C28-UDA-RFLAG

方向分类控制；416字节；关联及预期见§5。

```message-hex
01 00 01 a0 c0 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 02 be
c0 00 01 11 00 00 28 af 3c 53 68 2d 44 61 74 61
3e 3c 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c 45
78 74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73
69 6f 6e 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c
53 4d 53 52 65 67 69 73 74 72 61 74 69 6f 6e 49
6e 66 6f 3e 3c 49 50 2d 53 4d 2d 47 57 2d 4e 75
6d 62 65 72 3e 31 32 30 32 35 35 35 30 39 39 39
3c 2f 49 50 2d 53 4d 2d 47 57 2d 4e 75 6d 62 65
72 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 53 43
41 64 64 72 65 73 73 3e 31 32 30 32 35 35 35 30
31 30 30 3c 2f 53 43 41 64 64 72 65 73 73 3e 3c
2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 53 4d 53
52 65 67 69 73 74 72 61 74 69 6f 6e 49 6e 66 6f
3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 45 78 74 65 6e
73 69 6f 6e 3e 3c 2f 53 68 2d 49 4d 53 2d 44 61
74 61 3e 3c 2f 53 68 2d 44 61 74 61 3e 00 00 00
```

### N28-UDA-N-3002

无效答案；416字节；关联及预期见§5。

```message-hex
01 00 01 a0 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 0b ba 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 02 be
c0 00 01 11 00 00 28 af 3c 53 68 2d 44 61 74 61
3e 3c 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c 45
78 74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73
69 6f 6e 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c
53 4d 53 52 65 67 69 73 74 72 61 74 69 6f 6e 49
6e 66 6f 3e 3c 49 50 2d 53 4d 2d 47 57 2d 4e 75
6d 62 65 72 3e 31 32 30 32 35 35 35 30 39 39 39
3c 2f 49 50 2d 53 4d 2d 47 57 2d 4e 75 6d 62 65
72 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 53 43
41 64 64 72 65 73 73 3e 31 32 30 32 35 35 35 30
31 30 30 3c 2f 53 43 41 64 64 72 65 73 73 3e 3c
2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 53 4d 53
52 65 67 69 73 74 72 61 74 69 6f 6e 49 6e 66 6f
3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 45 78 74 65 6e
73 69 6f 6e 3e 3c 2f 53 68 2d 49 4d 53 2d 44 61
74 61 3e 3c 2f 53 68 2d 44 61 74 61 3e 00 00 00
```

### C28-UDA-N-UNKNOWNEXP

未知语义控制；436字节；关联及预期见§5。

```message-hex
01 00 01 b4 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 29
40 00 00 20 00 00 01 0a 40 00 00 0c 00 06 79 32
00 00 01 2a 40 00 00 0c 00 00 17 6f 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 be c0 00 01 11 00 00 28 af 3c 53 68 2d
44 61 74 61 3e 3c 53 68 2d 49 4d 53 2d 44 61 74
61 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 45 78
74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73 69
6f 6e 3e 3c 53 4d 53 52 65 67 69 73 74 72 61 74
69 6f 6e 49 6e 66 6f 3e 3c 49 50 2d 53 4d 2d 47
57 2d 4e 75 6d 62 65 72 3e 31 32 30 32 35 35 35
30 39 39 39 3c 2f 49 50 2d 53 4d 2d 47 57 2d 4e
75 6d 62 65 72 3e 3c 45 78 74 65 6e 73 69 6f 6e
3e 3c 53 43 41 64 64 72 65 73 73 3e 31 32 30 32
35 35 35 30 31 30 30 3c 2f 53 43 41 64 64 72 65
73 73 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c
2f 53 4d 53 52 65 67 69 73 74 72 61 74 69 6f 6e
49 6e 66 6f 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e
3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 53 68 2d 49 4d
53 2d 44 61 74 61 3e 3c 2f 53 68 2d 44 61 74 61
3e 00 00 00
```

### C28-UDA-E-EXTRA-UNKNOWN

E补充结果边界；128字节；关联及预期见§5。

```message-hex
01 00 00 80 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c
00 06 79 32 00 00 01 2a 40 00 00 0c 00 00 17 6f
```

### C28-UDA-E-EXTRA-3GPP

E补充结果边界；128字节；关联及预期见§5。

```message-hex
01 00 00 80 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 15 64 72 61 2d 61 2e 65 78 61 6d 70 6c
65 00 00 00 00 00 01 28 40 00 00 0f 65 78 61 6d
70 6c 65 00 00 00 01 0c 40 00 00 0c 00 00 0b ba
00 00 01 29 40 00 00 20 00 00 01 0a 40 00 00 0c
00 00 28 af 00 00 01 2a 40 00 00 0c 00 00 13 ee
```

### N28-UDA-N-AGENT

无效答案；420字节；关联及预期见§5。

```message-hex
01 00 01 a4 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 15 64 72 61 2d
61 2e 65 78 61 6d 70 6c 65 00 00 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 02 be c0 00 01 11 00 00 28 af 3c 53 68 2d
44 61 74 61 3e 3c 53 68 2d 49 4d 53 2d 44 61 74
61 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 45 78
74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73 69
6f 6e 3e 3c 53 4d 53 52 65 67 69 73 74 72 61 74
69 6f 6e 49 6e 66 6f 3e 3c 49 50 2d 53 4d 2d 47
57 2d 4e 75 6d 62 65 72 3e 31 32 30 32 35 35 35
30 39 39 39 3c 2f 49 50 2d 53 4d 2d 47 57 2d 4e
75 6d 62 65 72 3e 3c 45 78 74 65 6e 73 69 6f 6e
3e 3c 53 43 41 64 64 72 65 73 73 3e 31 32 30 32
35 35 35 30 31 30 30 3c 2f 53 43 41 64 64 72 65
73 73 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c
2f 53 4d 53 52 65 67 69 73 74 72 61 74 69 6f 6e
49 6e 66 6f 3e 3c 2f 45 78 74 65 6e 73 69 6f 6e
3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 53 68 2d 49 4d
53 2d 44 61 74 61 3e 3c 2f 53 68 2d 44 61 74 61
3e 00 00 00
```

### N28-UDA-E-ORIGIN

无效答案；92字节；关联及预期见§5。

```message-hex
01 00 00 5c 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 14 65 76 69 6c 2e 65 78 61 6d 70 6c 65
00 00 01 28 40 00 00 0f 65 78 61 6d 70 6c 65 00
00 00 01 0c 40 00 00 0c 00 00 0b ba
```

### E28-SNA-5001

E错误控制；204字节；关联及预期见§5。

```message-hex
01 00 00 cc 60 00 01 34 01 00 00 01 25 25 00 01
35 35 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 35 3b 31 00 00 00 01 0c
40 00 00 0c 00 00 13 89 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 17 40 00 00 40 00 00 02 74 c0 00 00 38
00 00 28 af 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 02 75 80 00 00 10 00 00 28 af 00 00 00 01
00 00 02 76 80 00 00 10 00 00 28 af 00 00 00 01
00 00 01 19 00 00 00 17 55 6e 73 75 70 70 6f 72
74 65 64 20 41 56 50 00 00 00 01 26 00 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00
```

### N28-SNA-E-NOFAILED

无效答案；140字节；关联及预期见§5。

```message-hex
01 00 00 8c 60 00 01 34 01 00 00 01 25 25 00 01
35 35 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 35 3b 31 00 00 00 01 0c
40 00 00 0c 00 00 13 89 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 19 00 00 00 17 55 6e 73 75 70 70 6f 72
74 65 64 20 41 56 50 00 00 00 01 26 00 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00
```

### N28-SNA-E-DUPFAILED

无效答案；268字节；关联及预期见§5。

```message-hex
01 00 01 0c 60 00 01 34 01 00 00 01 25 25 00 01
35 35 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 35 3b 31 00 00 00 01 0c
40 00 00 0c 00 00 13 89 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 17 40 00 00 40 00 00 02 74 c0 00 00 38
00 00 28 af 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 02 75 80 00 00 10 00 00 28 af 00 00 00 01
00 00 02 76 80 00 00 10 00 00 28 af 00 00 00 01
00 00 01 17 40 00 00 40 00 00 02 74 c0 00 00 38
00 00 28 af 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 02 75 80 00 00 10 00 00 28 af 00 00 00 01
00 00 02 76 80 00 00 10 00 00 28 af 00 00 00 01
00 00 01 19 00 00 00 17 55 6e 73 75 70 70 6f 72
74 65 64 20 41 56 50 00 00 00 01 26 00 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00
```

### N28-SNA-E-EMPTYFAILED

无效答案；148字节；关联及预期见§5。

```message-hex
01 00 00 94 60 00 01 34 01 00 00 01 25 25 00 01
35 35 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 35 3b 31 00 00 00 01 0c
40 00 00 0c 00 00 13 89 00 00 01 08 40 00 00 13
68 73 73 2e 65 78 61 6d 70 6c 65 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 17 40 00 00 08 00 00 01 19 00 00 00 17
55 6e 73 75 70 70 6f 72 74 65 64 20 41 56 50 00
00 00 01 26 00 00 00 13 68 73 73 2e 65 78 61 6d
70 6c 65 00
```

### V28-UDR-PROXY-R

请求控制；320字节；关联及预期见§5。

```message-hex
01 00 01 40 c0 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 15
40 00 00 0c 00 00 00 01 00 00 01 08 40 00 00 12
61 73 2e 65 78 61 6d 70 6c 65 00 00 00 00 01 28
40 00 00 13 65 78 61 6d 70 6c 65 2e 6e 65 74 00
00 00 01 25 40 00 00 13 68 73 73 2e 65 78 61 6d
70 6c 65 00 00 00 01 1b 40 00 00 13 65 78 61 6d
70 6c 65 2e 6e 65 74 00 00 00 02 bc c0 00 00 30
00 00 28 af 00 00 02 59 c0 00 00 21 00 00 28 af
73 69 70 3a 61 6c 69 63 65 40 65 78 61 6d 70 6c
65 2e 6e 65 74 00 00 00 00 00 02 bf c0 00 00 10
00 00 28 af 00 00 00 18 00 00 01 1c 40 00 00 2c
00 00 01 18 40 00 00 17 70 72 6f 78 79 2d 61 2e
65 78 61 6d 70 6c 65 00 00 00 00 21 40 00 00 0b
00 ff 41 00 00 00 01 1c 40 00 00 2c 00 00 01 18
40 00 00 17 70 72 6f 78 79 2d 62 2e 65 78 61 6d
70 6c 65 00 00 00 00 21 40 00 00 0a 80 42 00 00
```

### V28-UDA-PROXY-A

正常控制；504字节；关联及预期见§5。

```message-hex
01 00 01 f8 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 02 be
c0 00 01 11 00 00 28 af 3c 53 68 2d 44 61 74 61
3e 3c 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c 45
78 74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73
69 6f 6e 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c
53 4d 53 52 65 67 69 73 74 72 61 74 69 6f 6e 49
6e 66 6f 3e 3c 49 50 2d 53 4d 2d 47 57 2d 4e 75
6d 62 65 72 3e 31 32 30 32 35 35 35 30 39 39 39
3c 2f 49 50 2d 53 4d 2d 47 57 2d 4e 75 6d 62 65
72 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 53 43
41 64 64 72 65 73 73 3e 31 32 30 32 35 35 35 30
31 30 30 3c 2f 53 43 41 64 64 72 65 73 73 3e 3c
2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 53 4d 53
52 65 67 69 73 74 72 61 74 69 6f 6e 49 6e 66 6f
3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 45 78 74 65 6e
73 69 6f 6e 3e 3c 2f 53 68 2d 49 4d 53 2d 44 61
74 61 3e 3c 2f 53 68 2d 44 61 74 61 3e 00 00 00
00 00 01 1c 40 00 00 2c 00 00 01 18 40 00 00 17
70 72 6f 78 79 2d 61 2e 65 78 61 6d 70 6c 65 00
00 00 00 21 40 00 00 0b 00 ff 41 00 00 00 01 1c
40 00 00 2c 00 00 01 18 40 00 00 17 70 72 6f 78
79 2d 62 2e 65 78 61 6d 70 6c 65 00 00 00 00 21
40 00 00 0a 80 42 00 00
```

### E28-UDA-PROXY-A

E错误控制；184字节；关联及预期见§5。

```message-hex
01 00 00 b8 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 01 0c 40 00 00 0c 00 00 13 94
00 00 01 1c 40 00 00 2c 00 00 01 18 40 00 00 17
70 72 6f 78 79 2d 61 2e 65 78 61 6d 70 6c 65 00
00 00 00 21 40 00 00 0b 00 ff 41 00 00 00 01 1c
40 00 00 2c 00 00 01 18 40 00 00 17 70 72 6f 78
79 2d 62 2e 65 78 61 6d 70 6c 65 00 00 00 00 21
40 00 00 0a 80 42 00 00
```

### N28-UDA-N-PROXYORDER

无效答案；504字节；关联及预期见§5。

```message-hex
01 00 01 f8 40 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 04
40 00 00 20 00 00 01 0a 40 00 00 0c 00 00 28 af
00 00 01 02 40 00 00 0c 01 00 00 01 00 00 01 0c
40 00 00 0c 00 00 07 d1 00 00 01 15 40 00 00 0c
00 00 00 01 00 00 01 08 40 00 00 13 68 73 73 2e
65 78 61 6d 70 6c 65 00 00 00 01 28 40 00 00 13
65 78 61 6d 70 6c 65 2e 6e 65 74 00 00 00 02 be
c0 00 01 11 00 00 28 af 3c 53 68 2d 44 61 74 61
3e 3c 53 68 2d 49 4d 53 2d 44 61 74 61 3e 3c 45
78 74 65 6e 73 69 6f 6e 3e 3c 45 78 74 65 6e 73
69 6f 6e 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c
53 4d 53 52 65 67 69 73 74 72 61 74 69 6f 6e 49
6e 66 6f 3e 3c 49 50 2d 53 4d 2d 47 57 2d 4e 75
6d 62 65 72 3e 31 32 30 32 35 35 35 30 39 39 39
3c 2f 49 50 2d 53 4d 2d 47 57 2d 4e 75 6d 62 65
72 3e 3c 45 78 74 65 6e 73 69 6f 6e 3e 3c 53 43
41 64 64 72 65 73 73 3e 31 32 30 32 35 35 35 30
31 30 30 3c 2f 53 43 41 64 64 72 65 73 73 3e 3c
2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 53 4d 53
52 65 67 69 73 74 72 61 74 69 6f 6e 49 6e 66 6f
3e 3c 2f 45 78 74 65 6e 73 69 6f 6e 3e 3c 2f 45
78 74 65 6e 73 69 6f 6e 3e 3c 2f 45 78 74 65 6e
73 69 6f 6e 3e 3c 2f 53 68 2d 49 4d 53 2d 44 61
74 61 3e 3c 2f 53 68 2d 44 61 74 61 3e 00 00 00
00 00 01 1c 40 00 00 2c 00 00 01 18 40 00 00 17
70 72 6f 78 79 2d 62 2e 65 78 61 6d 70 6c 65 00
00 00 00 21 40 00 00 0a 80 42 00 00 00 00 01 1c
40 00 00 2c 00 00 01 18 40 00 00 17 70 72 6f 78
79 2d 61 2e 65 78 61 6d 70 6c 65 00 00 00 00 21
40 00 00 0b 00 ff 41 00
```

### N28-UDA-E-PROXYORDER

无效答案；184字节；关联及预期见§5。

```message-hex
01 00 00 b8 60 00 01 32 01 00 00 01 24 24 00 01
34 34 00 01 00 00 01 07 40 00 00 17 61 73 2e 65
78 61 6d 70 6c 65 3b 32 34 3b 31 00 00 00 01 08
40 00 00 13 68 73 73 2e 65 78 61 6d 70 6c 65 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 01 0c 40 00 00 0c 00 00 13 94
00 00 01 1c 40 00 00 2c 00 00 01 18 40 00 00 17
70 72 6f 78 79 2d 62 2e 65 78 61 6d 70 6c 65 00
00 00 00 21 40 00 00 0a 80 42 00 00 00 00 01 1c
40 00 00 2c 00 00 01 18 40 00 00 17 70 72 6f 78
79 2d 61 2e 65 78 61 6d 70 6c 65 00 00 00 00 21
40 00 00 0b 00 ff 41 00
```

### E28-PNA-5004

E诊断控制；132字节；关联及预期见§5。

```message-hex
01 00 00 84 60 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 08
40 00 00 12 61 73 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 01 0c 40 00 00 0c 00 00 13 8c
00 00 01 17 40 00 00 24 00 00 02 bc c0 00 00 1c
00 00 28 af 00 00 02 59 c0 00 00 0e 00 00 28 af
c3 28 00 00
```

### N28-PNA-E-BADTEXT

无效答案；144字节；关联及预期见§5。

```message-hex
01 00 00 90 60 00 01 35 01 00 00 01 24 24 00 04
34 34 00 04 00 00 01 07 40 00 00 18 68 73 73 2e
65 78 61 6d 70 6c 65 3b 32 34 3b 34 00 00 01 08
40 00 00 12 61 73 2e 65 78 61 6d 70 6c 65 00 00
00 00 01 28 40 00 00 13 65 78 61 6d 70 6c 65 2e
6e 65 74 00 00 00 01 0c 40 00 00 0c 00 00 13 8c
00 00 01 17 40 00 00 24 00 00 02 bc c0 00 00 1c
00 00 28 af 00 00 02 59 c0 00 00 0e 00 00 28 af
c3 28 00 00 00 00 01 19 00 00 00 0a c3 28 00 00
```

## 7. 第二十九轮公共边界补充

本页148帧原样保留；新增未知M1的正常/E答案、5014坏长度诊断pair和279外层越界在[逐叶§12](diameterkit-leaf-vectors.md)。诊断内部长度可能就是错误证据，接收不能要求它像普通Grouped一样全部可递归；具体容器边界、短证据、深度和配额对照见[主流程§25.3](diameterkit-mainflow-audit.md)。本页计数仍为148，不将新81帧重计。

## 8. 第三十轮公共结果与诊断入口

本页148帧继续保留；[逐叶§13](diameterkit-leaf-vectors.md)补九族结果父组、错误文本及E外围281/附加297缺Vendor控制。Origin-State与Route-Record不按Session/Proxy规则回显；错误文本中的U+0000虽可UTF8编码，仍不满足Diameter规则。唯一规则正文见[主流程§26](diameterkit-mainflow-audit.md)。

## 9. 第三十一轮目的拒绝与缺叶出口

本页148帧保持；[逐叶§14](diameterkit-leaf-vectors.md)新增九E3002关联目的Host不匹配的请求，Origin取本地配置而非错误Destination。缺Session可确认不存在与原有Session/Proxy无法合法回显不同；后者不能靠诊断例外生成非法顶层字段。细则与待补5005/坏回显出口见[主流程§27.4](diameterkit-mainflow-audit.md)，九3002不代表全部错误出口闭合。
