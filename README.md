# diameterkit
A pure-Go Diameter library targeting Go 1.26+, with generated, reflection-free codecs and versioned protocol definitions.

应用版本、Application-ID、JSON定义入口及Go生成方法见[应用定义指南](doc/diameterkit-application-definitions.md)。本库不加载运行时XML字典；Sh/S6c/SGd/Ro的输入在`spec/*.json`，生成类型在`app/*/zz_generated_messages.go`。

Run the standard Ro SMS IEC example with `CGO_ENABLED=0 go run ./examples/ro`, or add `-flow ecur` for no-UPDATE INITIAL/TERMINATION or `-flow abort` for independent OCS→CTF ASR/ASA; see [CCR/CCA results, application boundaries and cleanup](examples/ro/README.md).

Run the public SGd/Gdd MO/MT example with `CGO_ENABLED=0 go run ./examples/sgd` or add `-interface gdd`; see [direction, result and budget configuration](examples/sgd/README.md).

Run the public S6c GMSC/HSS routing example with `CGO_ENABLED=0 go run ./examples/s6c`; see [configuration, result namespaces and cleanup](examples/s6c/README.md).

Run the public Sh Ref24 TCP example with `CGO_ENABLED=0 go run ./examples/sh`; see [configuration, XML boundaries and cleanup](examples/sh/README.md).

Run the public SCAP1 bidirectional TCP example with `CGO_ENABLED=0 go run ./examples/scap1`; see [configuration, result modes and cleanup](examples/scap1/README.md).


Current development slices provide offline Base control messages and ordinary MSISDN S6c SRR/SRA with IMSI and bounded SGSN/MME/MSC routing nodes, RDR/RDA with separate or combined MME/MSC/SGSN/IP-SM-GW outcomes and HSS or explicitly role-bound MME/SGSN ALR/ALA success/failure pairs, plus ordinary IMSI SGd/Gdd OFR/OFA and MME/SGSN TFR/TFA with bounded retransmission-time pairing, explicit eras and application evidence. A bounded T4 MO charging slice adds OFA External-Identifier and explicit original-sender/SC/terminal evidence; S6c also supports bounded T4 no-MSISDN SRR/SRA, RDR/RDA and HSS ALR/ALA with explicit purpose, IMSI, SC and terminal evidence; native T4 and SCTP deployment acceptance remain pending. SRA now supports selected experimental failures 5550/5552/5556/5557, domain diagnostics with explicit zero/presence and bounded GMSC/IP-SM-GW role/UNRI evidence. Sh now adds ordinary Ref24 UDR/UDA and PUR/PUA, bounded single/mixed Ref11/12/25 subscription and unsubscription SNR/SNA and independent PNR/PNA, with required Sh VSAI and opaque User-Data. They include common normal/E error envelopes, explicit request/answer association, the JSON-to-Go generator, bounded codec, owned snapshots and frozen profiles. A bounded SingleAttempt Node/Peer runtime now includes budget admission, per-connection pending, completion ports, Base handshake/watchdog, local draining/shutdown, dual-connection election and cancel/close handling, tested with fake streams and real loopback TCP/TLS connections. Production TCP/TLS adapters now provide bounded framing, explicit TLS handshake, cancellation and close handling. A limited inbound path adds explicit no-cache command policies, Peer.Receive, one-time Inbound.Respond and an independent inbound budget. Opt-in bounded automatic errors and named Inbound.Reject share the response commit gate, with an independently reserved emergency path. Generated CanonicalV1 and an opt-in bounded Node-local duplicate cache now provide conflict detection and per-arrival response replay with fixed retention. Fixed-terminal Routes now support bounded transport failover or explicit original-identity retransmission of the SRR slice, with independent attempt answer budgets, late-answer association and one completion. Bounded Dialer/Maintain now adds TCP/TLS establishment, cancellation, capped backoff and reconnect suppression; reconnection never automatically replays business requests. Per-connection hierarchical byte DRR now schedules control, answers and requests with class weights, an emergency answer lane and independent hard budgets. Fixed-peer Listener/Serve now reserves candidates before accepting TCP sockets and shares a deadline across TLS and CER/CEA; cancellation waits for unfinished callbacks while established connections remain owned by Peer/Node. Bounded identifier spaces, custom application E2E sources, Session-Id generation and class guard quotas now track retirement across retries and I/O; HBH exhaustion drains the connection. Scoped validation reports now connect offline context/answer checks, immutable inbound evidence and runtime results; compiled NoContext permission prevents unaudited semantics from silently passing. Base and ordinary SGd without retransmission-time negotiation, ordinary S6c queries/reports and HSS alerts, and plain Sh retain zero-context checks; MME/SGSN alerts require bounded explicit terminal-role evidence; Sh feature messages use explicit bounded end-terminal evidence, with incomplete reports when evidence is missing. Sh XML/schema and business semantics are not included in its AVP-only report. The separate `app/sh/shdata` package now provides explicit bounded Ref11/12/24/25 XML builders, parsing and scoped checks, including explicit Ref11/12/25 notification combinations, preserving original bytes and uninterpreted extensions. It is not a full XSD validator or subscription store. S6c now checks Additional-Serving-Node against the original SRR GPRS flag, with bounded international-decimal SMEA helpers and optional success data. Registration history and VLR facts remain application responsibilities. Compiled normal-answer relations additionally check the original TFR target for MME-only TFA results and location fields. TPDU bytes remain opaque; no SMS retry or delivery state is managed. Selected Ro CCR recovery, independent ASR/ASA, DRA Relay negotiation and bounded request targets are now implemented. Complete rejection diagnostics, shared multi-peer admission and remaining application subsets still need acceptance; all four applications remain in the v1 scope.

Core code and tools build with `CGO_ENABLED=0`, with no third-party runtime dependencies or persistent storage. The separate Linux SCTP adapter uses kernel sockets through the published pure-Go skill215/sctp v0.8.0 module. The core/adapter public releases remain pending.

```sh
CGO_ENABLED=0 go test ./...
CGO_ENABLED=0 go run ./examples/codec
```

See [implementation progress](spec/progress.json), [implemented scope and checks](spec/README.md), [Sh ordinary slice](spec/history/batch-17.json), [Sh-Data helpers](spec/history/batch-18.json), [Sh features, mixed subscriptions and 5011](spec/history/batch-20.json), [S6c routing slice](spec/history/batch-21.json), [S6c multi-node delivery reports](spec/history/batch-22.json), [MME/SGSN alerts](spec/history/batch-23.json), [SGd/Gdd retransmission times](spec/history/batch-24.json), [T4 MO charging identity](spec/history/batch-25.json), [S6c T4 report/alert identities](spec/history/batch-26.json), [S6c T4 routing](spec/history/batch-27.json), [SRA failures and diagnostics](spec/history/batch-28.json), [SGd MO unknown-user interface evidence](spec/history/batch-29.json), [S6c IP-SM-GW nodes](spec/history/batch-30.json), [S6c IP-SM-GW delivery outcomes](spec/history/batch-31.json), [S6c IMS delivery reports](spec/history/batch-32.json), [S6c IMS HSS alerts](spec/history/batch-33.json), [S6c IMS routing](spec/history/batch-34.json), [IMS MO forwarding](spec/history/batch-35.json), [IMS MT forwarding](spec/history/batch-36.json), [IMS MO source identities](spec/history/batch-37.json), [T4 IP-SM-GW reports](spec/history/batch-38.json), [Ro SMS IEC](spec/history/batch-39.json), [Ro SMS ECUR](spec/history/batch-40.json), and the [offline example](examples/codec/main.go). Regenerate output with `go generate ./...`; tests detect stale generated files.

See the [implementation architecture and package layout](doc/diameterkit-architecture.md), [design](doc/diameterkit-design.md) and [library specification](doc/diameterkit-specification.md) for scope, boundaries, and acceptance requirements.

SGd OFA experimental5001 now has explicit interface, original MSISDN and terminal evidence. Gdd/T4 keep the owned answer with an unsupported report; zero context needs application evidence. See [batch29](spec/history/batch-29.json).

S6c SRA supports selected IP-SM-GW Number/Name/SBI combinations. Realm, reserved SBI and Additional IP-SM-GW values are retained on receive with an unsupported report; standard send rejects them. Named Realm compatibility is not enabled. See [batch30](spec/history/batch-30.json).

Ordinary-MSISDN S6c RDR now supports IP-SM-GW cause0/1/2 and optional cause1 diagnostic11/12. Unsupported well-shaped values are retained by the offline codec with an unsupported report; the current runtime rejects and closes unsupported inbound requests. Later batches add selected IMS and T4 IP outcomes; delivery scheduling remains an application responsibility. See [batch31](spec/history/batch-31.json).

S6c IMS RDR/RDA now supports HSS-ID at User-Identifier/User-Name and destination-only SMSMI correlation, with explicit IMSDeliveryReport evidence binding the original identities, SC and terminal endpoints. The selected ASCII SIP URI subset is bounded to256 bytes; no HSS discovery, SIP normalization or delivery history is managed. See [batch32](spec/history/batch-32.json).

S6c IMS ALR/ALA now supports the explicit fifteen-zero dummy MSISDN, destination-only SMSMI URI and optional raw availability Time, with IMSHSSAlert evidence for the HSS/IWMSC endpoints and SC. Reports and alerts retain independent pending/completion; no MWD storage, alert scheduling or automatic SMS forwarding is added. See [batch33](spec/history/batch-33.json).

S6c IMS SRR/SRA now supports request HSS-ID and destination SIP URI with explicit IMSRouting evidence for the role, SC and known HSS. The HSS supplies the successful answer IMSI; the requester need not know it beforehand. Fixed same-HSS active-request recovery preserves the original identity and URI. No subscriber lookup, no-IMSI success fallback or persistent replay is added. See [batch34](spec/history/batch-34.json).

IMS OFR/OFA now supports the selected no-IMSI dummy source, all three SMSMI identities and temporary IP-SM-GW delivery outcomes, with explicit IMSMOForwarding evidence. The library does not forward SMS or retry delivery automatically. Development API change: `sgd.UserIdentifier.UserName` is now `avp.Optional[string]`; existing known-IMSI callers use `avp.Some(imsi)`. See [batch35](spec/history/batch-35.json).

IMS TFR/TFA now supports top-level HSS-ID, separate originating/destination SMSMI URIs, and explicit IMSMTForwarding evidence for SMS-GMSC/IP-SM-GW endpoints. Delivery Timer/Start retain raw optional values; no delivery scheduling or OFR-to-TFR conversion is added. See [batch36](spec/history/batch-36.json).

IMS MO also supports a known originating IMSI with an optional real MSISDN. Set `IMSMOForwardingEvidence.SenderIMSI` and `SenderMSISDN` explicitly; both absent retain the prior dummy-source contract. Values and presence are bound to the owned request, without expanding Context budgets or enabling automatic retry. See [batch37](spec/history/batch-37.json).

S6c T4 reports without MSISDN now support the selected IP-SM-GW outcomes and legal combinations with MME/MSC/SGSN, using the existing T4DeliveryReport identity/SC/endpoint evidence. No extra Context capacity, MWD store or replay permission is added. See [batch38](spec/history/batch-38.json).

Ro now provides a bounded standard SMS IEC DIRECT_DEBITING CCR/CCA profile: one MSCC service, exact uint64 event units, opaque refund bytes, paired success/refusal/E answers and original-request association. TCP/TLS tests exercise both directions and cancellation. This was the first Ro slice; ECUR is added in batch40, while legacy templates/private bindings and RO-06 recovery remain v1 work. Generator10 changes all generated definition fingerprints; existing Base/S6c/SGd/Sh definitions are unchanged. See [batch39](spec/history/batch-39.json).

Ro slice.2 adds standard IP-SM-GW SMS ECUR INITIAL/TERMINATION for the no-UPDATE path, explicit final usage and paired 2001/5002 answers. `ro.CheckECURContinuation` checks two caller-supplied requests; the runtime retains no charging session and sends no automatic TERM or retry. Frozen INITIAL equipment/location fields are supported in bounded form. See [batch40](spec/history/batch-40.json) for scope, tests and the correction that SMS-Node1 means IP-SM-GW.

The isolated [`compat/smsfwro`](compat/smsfwro/iec.go) profile now supports the complete source-defined IEC request template, including legacy string/address layouts and signed fields, plus selected normal/E answers. It cannot share an App4 Bundle with standard `app/ro`. [Batch41](spec/history/batch-41.json) records source projections, tests and the remaining OCS/private-field, refund420 and recovery boundaries.

The named [`legacy420`](compat/smsfwro/legacy420/refund.go) refund slice adds fixed Grouped420 money, exact signed amounts and explicit binary token input. Standard Ro and IEC-only profiles remain isolated. [Batch42](spec/history/batch-42.json) records template coverage, tests and the remaining original-debit/OCS evidence requirements.

[`smsfwro`](compat/smsfwro/ecur.go) slice.2 now also builds the legacy ECUR INITIAL/TERMINATION templates in the same fixed profile as IEC, with explicit usage/status/time and a pure continuation check. [Batch43](spec/history/batch-43.json) documents the Optional-field API migration, the original TERM MMS ordering difference, and the limits of offline compatibility evidence.

The existing `legacy420` profile now covers the actual general sms-fw template for direct debit and balance check, alongside refund, with explicit amounts and top-level cost/unit/balance results. [Batch44](spec/history/batch-44.json) records the Optional API migration and the remaining Client-Address type boundary before combining all five CCR templates.

[`compat/scap1`](compat/scap1/event.go) provides the actual47-node SCAP1 template as an explicit App0/271 profile, ordered parameters and synthetic base ACA/error pairs. [Batch45](spec/history/batch-45.json) records the original offline foundation and independent dictionary evidence. Adding this profile does not enable SCAP1 network dispatch or capability advertisement.

[Batch46](spec/history/batch-46.json) adds SCAP1 SingleAttempt over TCP/TLS with explicit `SCAP1Accounting` command policies and strict Vendor193/Acct19302 capability binding. App0/271 uses ordinary application budgets/pending while Base control traffic remains separate. Reconnection rechecks capabilities and never replays business requests. The legacy dual Auth/Acct advertisement and actual private OCS answers remain outside the verified subset.

[Batch47](spec/history/batch-47.json) adds the explicit [`combined`](compat/smsfwro/combined/builders.go) App4 profile for all five sms-fw CCR templates, including balance use of the general template. Client-Address2018 has one fixed opaque wire representation with source-specific value constraints; no type guessing or automatic conversion is used. Original profiles remain unchanged. Paired TCP/TLS and cross-template rejection tests pass; target OCS acceptance of the mixed legacy representations and fixed field ordering remains unverified.

[Batch48](spec/history/batch-48.json) preserves command and service-level CCA results separately in all three legacy CCR profiles, adds selected4010/4011/5030 branches, IEC/ECUR cost details and general direct-debit DDFH. Original-request checks prevent balance/refund requests from acquiring DDFH behavior. Fields remain owned protocol evidence; no automatic charging, retry or persistence is introduced. Private token/PSP namespaces remain unverified.

第四十九批已完成Ro命令级5031诊断与5011安全接收：保留基础/实验结果及支持列表，冲突和未知语义不报告完整通过，自动实验配对仍禁用。65份新报文、63份原字典对照和全量测试通过；继续保持纯Go、无持久化，下一批进入RO-06。见[批次说明](spec/history/batch-49.json)与[实施进度](spec/progress.json)。

第五十批已启用四个Ro Profile的同OCS活动请求恢复：复用Route的显式重传或传输故障转移，保留原身份和内容、共享总期限与预算；有效答案结束请求。27份新报文及真实TCP/TLS竞态测试通过，保持纯Go、无持久化。目标去重证据及完整RO-06仍未完成；见[批次说明](spec/history/batch-50.json)与[进度](spec/progress.json)。

第五十一批已实现Ro原请求准备：冻结已选CCR、目标绑定和恢复上限，保留原E2E，返回应用独占的版本化记录；有界scratch/guard及取消关闭测试通过。准备不发送，记录校验不授权恢复；下一批接有限段接纳及累计额度。纯Go、无持久化边界保持，见[批次说明](spec/history/batch-51.json)和[进度](spec/progress.json)。

第五十二批接通`SubmitRoRecovery`：本Node/原OCS的首段与无答续段保持原E2E，复用Route并原子扣除累计额度；活动同段幂等、旧段晚答隔离和收发资源屏障已验。`Result.RoRecovery`交付累计证据，仍无持久化。第53批已补本地ResponseWait/Tr；第54批补具名本地错误续段；第55批补有限跨OCS；第56批补显式原记录导入，见[批次说明](spec/history/batch-52.json)与[进度](spec/progress.json)。

第五十三批加入可选`RoRecoveryPlan.ResponseWait`：从排队接纳计Tr，健康连接超时后显式恢复；保留旧pending和答案预算，旧答与后继收发竞争仍只完成一次。业务Tx由应用控制，writer结清前不开放后继，见[批次说明](spec/history/batch-53.json)。

第五十四批补具名错误答案后的本地有限续段：显式策略、应用Tx截止、真实顶层结果和NextTrigger控制准入；错误后T=0与历史发送事实分离。4002仅接收并报告语义不完整，普通请求及默认策略保持答案终态，见[批次说明](spec/history/batch-54.json)。

第五十五批补授权跨OCS恢复：独立RoTargets表冻结目的、路径及去重/会话证据，有限Alternates支持同DRA连接切换目标；旧答案保留原来源和独立预算。恢复记录升为v2，旧v1拒绝自动转换；第56批接通显式导入，见[批次说明](spec/history/batch-55.json)。

第五十六批接通显式原记录导入：确认应用隔离/进度/时钟声明，保留原E2E及正文，检查普通已用区间、active/quarantine与独立RecoverySpaces；Checkpoint区分未知发送和已合格错误后的T位。十一组测试覆盖真实TCP/TLS及关闭/回滚，见[批次说明](spec/history/batch-56.json)。应用仍负责保存、所有权与历史真实性，库不自动重放。

第五十七批完成RO-06应用交接示例：确切意图确认、提交未知停止、崩溃未知保守计数、终态禁止恢复、Tx/晚答与DDFH边界；三个可运行Example及七组TCP/TLS验收见[交接清单](doc/ro-recovery-handoff.md)与[批次说明](spec/history/batch-57.json)。生产API和协议定义不变，应用存储/隔离及实际OCS证据继续由部署提供。

第五十八批在`skill215/sctp`完成标准deadline和内部取消等待层基础，真实SCTP全量257条测试/子测试零skip、race及纯Go构建通过。上游补丁尚未发布，本库仍未接入SCTP；U3解析/拨号与U2/U4后续范围见[批次说明](spec/history/batch-58.json)。

第五十九批在`skill215/sctp`接通`ResolveSCTPAddrContext`与`SocketConfig.DialContext`：共用总预算、创建fd前拒绝Control、失败释放及交接后取消隔离。九组新增测试和外部包示例通过，见[批次说明](spec/history/batch-59.json)。候选未发布，消息模式/通知及Diameter适配器仍待实施。

第六十批在`skill215/sctp`实现显式消息模式、`ReadEvent/WriteMsgContext`、固定缓冲查询及必需选项读回。284条测试/子测试/示例零skip，全量race及纯Go检查通过，见[批次说明](spec/history/batch-60.json)。下一批U4 typed通知与其真实事件证据；上游发布和本库SCTP适配器仍未完成。

第六十一批在`skill215/sctp`补齐五类typed通知解析及原生事件：长SEND_FAILED不消费重读、地址变更、partial-abort、建立和关闭均有真实回环证据。312条测试/子测试/fuzz种子/示例零skip，见[批次说明](spec/history/batch-61.json)。下一步U5完整预算与同候选发布/平台验收，本库SCTP适配器尚未接入。

第六十二批完成SCTP U5本地资源核算、ABI/取消退休检查、可重复的离线模块安装验证及发布门禁接线，见[批次说明](spec/history/batch-62.json)。8KiB是有限使用方式下的上游逻辑固定额度，不是RSS或最终adapter预算；远端矩阵、正式版本和本库SCTP适配仍待完成。

第六十三批将精确SCTP候选交付为[上游PR #30](https://github.com/skill215/sctp/pull/30)，远端14项矩阵通过；真实伪版本下载及独立消费者收发验证完成，见[批次说明](spec/history/batch-63.json)。下一批可固定该候选开发SCTP适配器；PR仍为draft，尚未合并或正式发布，本库依赖未改。

第六十四批已实现独立[SCTP适配器](adapters/sctp/README.md)、预算计费、通知/Peer关闭与CEA发送屏障；真实Base/S6c及取消/关闭验证通过，见[批次记录](spec/history/batch-64.json)。核心不依赖SCTP；生产依赖发布及SCTP托管拨号/监听仍待后续。

第六十五批已补齐SCTP有限Dialer/Maintain与Listener/Serve：建立前N/Z/D限制、listener/候选独立计费、取消交接、真实重连与选举；重连不自动重放业务。见[本批记录](spec/history/batch-65.json)与[进度](spec/progress.json)。纯Go、无持久化；正式模块发布和现场互通仍独立验收。

第65批后维护交接（现已完成于第66批）：在skill215/sctp补齐通用接口，再删除本库20ms Accept轮询、底层syscall和重复地址/资源逻辑。见[需求文档](doc/sctp-upstream-improvement-plan.md)及[另一session提示词](doc/sctp-upstream-session-prompt.md)。该历史交接计划已执行，最新依赖和证据以下述第67批为准。

第六十六批完成A–E核对及SCTP薄适配迁移，固定真实候选`bb739b612088`。上游14项CI和源码/日志验证通过，本库核心2603项、adapter完整纯Go/race各71项及20轮1400项通过。见[本批记录](spec/history/batch-66.json)与[spec进度](spec/progress.json)。保持纯Go、无持久化；具名多流/PPID0及正式发布/现场验收仍单列。

第六十七批已固定正式发布的SCTP `v0.8.0`（`bce1b3cf3ea3`），70个上游Go源码与第66批候选相同，无Go代码改动。核心纯Go/race各2603、adapter各71、有限生命周期20轮620项通过。见[正式版本接入记录](spec/history/batch-67.json)。上游发布门已关闭；本库核心/adapter公开版本、具名多流/PPID0和现场验收仍待完成。

第六十八批实现DRA Relay-only能力协商及有限RequestTargets逐请求终端绑定，Route固定目标并保留原身份/来源/预算；独立SCTP模块切换新离线核心batch68，上游仍v0.8.0。见[本批记录](spec/history/batch-68.json)。

第六十九批实现独立 Ro ASR/ASA（`app/ro/abort`），可与标准或兼容CCR Profile组合。由应用显式选择ASA结果，库负责关联、入站预算和有限重复重答，不自动终止业务或发送TERM。见[本批记录](spec/history/batch-69.json)和[实施进度](spec/progress.json)。

第七十批新增通用DCCA单服务EVENT（`app/ro/dcca`）：余额查询、直接扣费、退款、标准413精确金额及原请求/条件结果检查。每Bundle选择一种272绑定，可与独立ASR组合；3GPP SMS、MSCC/2022与旧私有token仍按各自范围验收。见[本批记录](spec/history/batch-70.json)。

第七十一批补齐标准SMS Ro的4010/4011/5030顶层及服务级结果，以及原IEC答案2022的有界提取/回送检查，区分缺省、空值与未知，绑定原服务和显式OCS。保持纯Go、无持久化；完整SMS退款发送及旧OCS证据仍按原限制验收。见[本批记录](spec/history/batch-71.json)和[进度](spec/progress.json)。

第七十二批实现SCTP具名有序多流及PPID0兼容：按实际出流数固定选流，公开单流回退，保留CEA/DWR屏障和原预算。继续使用上游v0.8.0与核心batch71；真实Base/S6c及取消/关闭验收见[本批记录](spec/history/batch-72.json)。

第七十三批完成显式`compat/sgdalert`旧SGd ALR/ALA绑定，复用S6c规则并保持请求App/身份与授权本地Origin；26新报文和TCP/TLS验收见[本批记录](spec/history/batch-73.json)。现11生成定义加1委托Profile，离线核心batch73，上游SCTPv0.8.0；Router token/MCCMNC及发布出口仍待办。

第七十四批完成`s6c.RouterAlert`：显式绑定原TFA token、原TFR地址和Router/GMSC端点，复用标准/旧App ALR/ALA运行时。23新报文及TCP/TLS验收见[本批记录](spec/history/batch-74.json)。不保存映射或自动投递；SGd TFA token输出和SRA Router/MCCMNC查询仍单列。

第七十五批完成SGd Router腿TFR/TFA的显式Alert能力、时间条件和User-Identifier token输出，与前批Router ALR衔接；30新报文及真实TCP/TLS验收见[本批记录](spec/history/batch-75.json)。SGd slice.15，离线核心batch75，SCTPv0.8.0不变。下一批是SRA Router/MCCMNC用途，token映射和业务重发仍由应用负责。

第七十六批完成SRA Router强制HPLMN关联身份与SRR仅MCC/MNC查询：显式用途入口、3311=1、成功身份和可选节点的原请求关联；27新报文和TCP/TLS验收见[本批记录](spec/history/batch-76.json)。S6c slice.22及旧Alert委托指纹更新，离线核心batch76，SCTPv0.8.0不变；下一批为既有模拟器输入的有界接受/拒绝。

第七十七批完成既有模拟器输入决策：28份独立编码投影、拒绝/显式修正矩阵及真实TCP/TLS迁移验收，见[本批记录](spec/history/batch-77.json)与[迁移边界](doc/diameterkit-simulator-migration.md)。生产代码和11定义不变，1233wire/17XML；核心仍为离线batch76、SCTPv0.8.0。旧程序迁移/真实目标验收未完成，下一批优先SCAP1已明绑定与多Arrival机制。

第七十八批完成按Peer显式启用的SCAP1旧双Auth/Acct广告模式，以及现有ACR/ACA多Arrival、缓存冲突、取消关闭和预算验收，见[本批记录](spec/history/batch-78.json)。标准Base保持严格、SCAP1仍SingleAttempt；17新帧，1250wire/17XML。核心更新离线batch78，SCTPv0.8.0不变；私有ACA、恢复及目标证据仍未齐。

第七十九批完成SCAP1双向公开API示例：独立Node/真实TCP、标准与旧广告、ACA正常/失败及E答案、预算/取消/端口清理，见[本批记录](spec/history/batch-79.json)。独立无replace消费module通过，公开注释已修正；库运行逻辑、1250wire/17XML与Profile不变。示例纳入新的128文件离线核心batch79，SCTPv0.8.0不变；私有ACA/恢复/目标与其他P7验收仍单列。

Batch80 adds the [public Sh Ref24 example](examples/sh/README.md) and [application version/JSON definition guide](doc/diameterkit-application-definitions.md). Independent consumer, finite XML/budget/lifecycle checks and unchanged protocol scope are recorded in [batch80](spec/history/batch-80.json). Core/adapter releases and remaining P7 acceptance are still pending.

Batch81 adds the [S6c GMSC/HSS public routing example](examples/s6c/README.md), with distinct base/experimental/E results, fixed terminal/identity checks and finite resource cleanup. [Progress](spec/progress.json) tracks independent consumer evidence and remaining P7 gates; protocol scope is unchanged.

Batch82 adds the [SGd/Gdd MO/MT public example](examples/sgd/README.md): explicit directions and terminal bindings, distinct base/experimental/E results, opaque payload preservation and finite resource cleanup. [Evidence](spec/evidence/batch-82-sgd-example.json) records isolated consumer and regression checks; existing protocol scope is unchanged. Local batch Markdown remains ignored.

Batch83 adds the [standard Ro SMS IEC public example](examples/ro/README.md), separating command and service outcomes with explicit terminal/identity checks, exact units and bounded cleanup. [Evidence](spec/evidence/batch-83-ro-example.json) records isolated consumer and regression checks; no automatic financial retry, persistence or protocol expansion. Batch Markdown stays local and ignored.

Batch84 adds the [no-UPDATE Ro ECUR public flow](examples/ro/README.md): explicit application-owned continuation, fixed Session/endpoints/service, new exchange identities, independent result budgets and bounded cancel/close handling. [Evidence](spec/evidence/batch-84-ro-ecur-example.json) records 68 example checks per suite and isolated API consumption. Existing library behavior and protocol definitions remain unchanged; complete P7/release and target OCS evidence remain pending.

Batch85 adds the [independent Ro ASR/ASA public flow](examples/ro/README.md): explicit OCS→CTF authorization, original session reference, application-selected success/unknown/unable/E outcomes, and bounded completion/cancel/close handling. A successful ASA neither completes an outstanding CCR for the same session nor generates TERM. [Evidence](spec/evidence/batch-85-ro-abort-example.json) records 100 cumulative example checks and isolated API consumption. Library behavior and definitions remain unchanged; support/API coverage, release checks and target acceptance remain separate gates.

Current public APIs, profile limits and the bounded remaining acceptance work are in the [support matrix](doc/diameterkit-support-matrix.md). Batch86 adds a reproducible two-module API snapshot and [external binding contract tests](examples/contracts/README.md), with no production-source changes. The audit identifies the previously promised S6c result interpreter and thin client as missing implementation; existing Ro recovery signatures alone do not prove a complete public application workflow. See [evidence](spec/evidence/batch-86-public-surface.json).

当前checkout的双module CI与生成/API/race检查见[发布检查指南](doc/release-checks.md)：`python3 spec/tools/release_check.py`。临时workspace明确使用当前核心与adapter，不依赖未发布batch85核心；不等于已发布版本安装验收。

核心/adapter的[v0版本与API政策、真实安装验收](doc/version-install-policy.md)已提供；第93批实际冷缓存下载因尚无发布标签失败，不能将checkout CI等同于发行安装通过。
