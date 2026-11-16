# 标准Ro SMS IEC / ECUR / ASR公开接入示例

在repo根目录运行（Go1.26+、纯Go，无外部服务）：

```sh
CGO_ENABLED=0 go run ./examples/ro
CGO_ENABLED=0 go run ./examples/ro -result 2001/4012
CGO_ENABLED=0 go run ./examples/ro -result 4012
CGO_ENABLED=0 go run ./examples/ro -result 5012
CGO_ENABLED=0 go run ./examples/ro -result 3004
```

IEC/ECUR模式的两个独立Node通过127.0.0.1随机端口真实TCP连接：CTF只发送CCR，合成OCS只接收并应答，不创建完成端口。ASR模式的方向相反，见下节。默认`-flow iec`每次运行发送一笔标准SMS IEC DIRECT_DEBITING，随后DPR/DPA关闭。`-timeout`默认10秒；所有返回路径另有5秒有限Close/端口Close/Wait，结束打印双方零资源快照。

[config.go](config.go)显式选择Base+`app/ro` Profile、Auth-Application-Id=4能力和HeaderApplication绑定。正常CCR使用`18.32274@3gpp.org`、EVENT_REQUEST=4、CC-Request-Number=0、显式存在的Requested-Action=0、MSI=1、单MSCC/Service-Identifier=1及请求单位1；SMS-Information选择IP-SM-GW/Submission。值全部来自示例输入，不通过订户查询、业务状态或本地时钟推导。

| `-result` | 命令结果 | 单服务结果/数据 |
| --- | --- | --- |
| 默认/2001 | CCA Result-Code=2001 | MSCC Result-Code=2001、显式授予单位1、合成opaque 2022 token |
| 2001/4012 | CCA Result-Code=2001 | MSCC Result-Code=4012，无本例成功grant/token |
| 4012 | CCA Result-Code=4012 | 无MSCC |
| 5012 | CCA Result-Code=5012 | 无MSCC |
| 3004 | 独立E错误包络 | 回显Session/头关联，不套正常CCA语法 |

输出保留顶层与MSCC结果。`AnswerReceived`和`Validation.Complete`表示已收到并通过所选协议检查，**不等于扣费成功**。顶层2001不能覆盖服务级4012；grant和Refund-Information也不能证明真实OCS账本状态、短信投递或退款资格。失败、超时、Cancel、关闭都不能证明请求未扣费；本例不会自动退款、重发、切换OCS或在重连后重放。

2022只作透明字节往返，不触发退款。单位使用精确uint64；测试以`2^53+1`验证请求/答案均不经float舍入，不意味着该数量适用于实际SMS计费。这里的OCS只构造测试答案，没有账本、鉴权、定价、持久化或接入真实OCS。

CTF固定Destination-Host/Realm为OCS，并独立授权答案终端；OCS配置CTF请求白名单。这里邻接就是终端，实际经DRA时须分别配置邻接和最终CTF/OCS，不能把相邻Origin视为计费终端授权。仅加载Profile不会宣告能力或授权方向；本例不装载sms-fw私有、legacy420、通用DCCA或SCAP1绑定，也不启用RO-06恢复。各App版本/定义入口见[应用指南](../../doc/diameterkit-application-definitions.md)。

节点、传输、标识、class、Peer、完成端口的逻辑资源均有上限。仅发送CTF不配置Inbound；OCS仍保留当前Node契约要求的公共出站class/Peer预算，但没有CCR发送策略和完成端口。NoCache和SingleAttempt不提供持久化幂等或财务去重。额度是示例逻辑上限，不是RSS或生产容量承诺。

[main.go](main.go)只使用公开Serve/Dial/Submit/Receive/Respond及关闭接口。Submit ctx只控制接纳；完成等待ctx只停止等待，实际取消使用Node.Cancel。Wait不代消费结果；未消费完成仍占额度，须Receive或关闭端口。真实应用若需保留结果，必须自行消费处理；本例不提供存储。

[main_test.go](main_test.go)覆盖五类结果、顶层成功/服务拒绝不产生跟进请求、CC类型/序号/服务键关联、缺子结果/缺grant/未选实验配对拒绝、显式零action缺失、错误目的及反方向CCR、请求/答案来源限制、精确单位、预算背压、可变视图、取消Respond重试、晚答一次完成和端口/Node生命周期。仅调用公开API，无fake transport或根包私有helper。

[批次验收](../../spec/evidence/batch-83-ro-example.json)记录独立消费module及完整测试。第83批覆盖标准SMS IEC、本机TCP公开接入；ECUR本批见下节，其他既有Ro能力、真实OCS/DRA和公开发行仍有独立出口。本例通过不授予未选SMS退款流程或完整规范符合性。

## 无UPDATE ECUR（第84批）

```sh
CGO_ENABLED=0 go run ./examples/ro -flow ecur
CGO_ENABLED=0 go run ./examples/ro -flow ecur -term-result 5002
CGO_ENABLED=0 go run ./examples/ro -flow ecur -result 2001/4012
```

[ecur.go](ecur.go)在同一真实TCP连接演示INITIAL(type=1/number=0)与TERMINATION(type=3/number=1)。两请求保持同一Session、CTF/OCS Host/Realm、Service-Context和Service-Identifier，分别分配新的RequestID、HBH、E2E；没有UPDATE。INITIAL没有Requested-Action；TERM不带Requested-Action/RSU，显式报告USU=1和MSCC Reporting-Reason=FINAL(2)。请求的完整Origin/目的来自显式端点配置，不从DRA推导。

`ro.CheckECURContinuation(initial, term)`要求完整请求体，只检查这两份请求的所选协议连续性；它既不读取INITIAL答案，也不证明已获额度、实际投递或财务授权。示例先消费INITIAL完成，确认AnswerReceived、Validation.Complete及原请求关联，再由应用代码检查顶层和服务结果均为2001且显式grant覆盖合成用量1，才调用TERM。这个有限演示策略不是所有部署的计费业务规则；真实应用负责用量、时间、失败处理和会话进度。本库不保存计费会话或业务账本。

`-result`选择INITIAL，仍支持上表五类结果，但ECUR不带2022 token。INITIAL服务拒绝、顶层失败或E错误输出`TERMINATION=not-sent`。`-term-result`支持2001、5002、5012、E3004，默认2001；正常TERM答案不带grant/token。TERM失败不被改写为业务成功，也不自动发新INITIAL或退款。取消、关闭、未完成或不确定交换不触发本例续发；它们不证明请求未计费。IEC模式保持默认，拒绝非默认`-term-result`以免静默忽略输入。

[ecur_test.go](ecur_test.go)覆盖14种连续性反例、显式USU零值、父组独立所有权、INITIAL结果未消费时的TERM背压、独立字节预算、旧INITIAL答案/错服务/TERM grant和token拒绝、两个阶段的取消/关闭/晚答及一次完成。关闭前消费正常完成，所有退出路径按原有限清理流程归零；Wait不代消费端口结果。

[第84批验收](../../spec/evidence/batch-84-ro-ecur-example.json)及[独立公开API消费](../../spec/evidence/batch-84-ro-ecur-public-api.json)记录68项示例测试、八种ECUR CLI结果、保留IEC结果及26项已用API查询。公开库和既有协议定义不变，仍无持久化、自动金融重试、OCS切换或RO-06恢复授权；本机合成OCS不构成真实OCS/DRA互通证明。

## 独立ASR/ASA（第85批）

```sh
CGO_ENABLED=0 go run ./examples/ro -flow abort
CGO_ENABLED=0 go run ./examples/ro -flow abort -result 5002
CGO_ENABLED=0 go run ./examples/ro -flow abort -result 5012
CGO_ENABLED=0 go run ./examples/ro -flow abort -result 3004
CGO_ENABLED=0 go run ./examples/ro -flow abort -session 'ctf.example;unknown'
```

[abort.go](abort.go)组合Base、一个标准CCR Profile与独立的[app/ro/abort](../../app/ro/abort/rules.go) Profile（版本`0.1.0-slice.1`、App4/Command274，源定义[roabort.json](../../spec/roabort.json)）。明确广告Auth-Application-Id=4；只授权OCS发送ASR、CTF接收ASR，OCS持有唯一完成端口。加载CCR定义不授权CCR，也不授权反方向ASR。这个示例复用有限资源配置和Serve/Dial/TCP清理；真实DRA邻接与最终OCS/CTF授权仍需分开配置。

ASR引用应用提供的原Session，默认是合成引用`ctf.example;example;original`，不会为ASR调用NewSessionID。ASR带Auth-Application-Id=4；正常ASA按独立274语法构造，不附CCR类型/编号、Auth-Application-Id或MSCC。Session与头关联由公开运行时保留；不套CCA语法。

| 应用选择 `-result` | 答案 | 含义与限制 |
| --- | --- | --- |
| 2001（默认） | ASA 2001 | 合成应用报告中止成功；库不会执行会话业务动作 |
| 5002 | ASA 5002 | 应用报告未知会话 |
| 5012 | ASA 5012 | 应用报告不能执行 |
| 3004 | 独立E错误 | 不把3004写入正常ASA |

`decideAbort`只演示显式应用决策：收到的Session不匹配合成已知引用时，先返回5002；匹配时才采用所选结果。没有会话表、查询服务、持久化、账本或真正短信任务。真实CTF应在应用层执行相应业务动作并决定ASA结果，避免把协议Validation.Complete当作会话状态事实。失败/取消不证明OCS或CTF未执行动作，ASA成功也不证明扣费、退款或投递已完成。

ASR是独立exchange。成功ASA不会结束同Session的挂起CCR、自动产生TERM或获得RO-06重传许可；应用需显式协调现有操作、终止上报和重复业务决策。NoCache/SingleAttempt不提供跨重启去重。标准正常/E应答、未知可选扩展及可选本地Origin-State的完整范围仍以原Profile为准；本例省略本地Origin-State，不回填远端请求值。

[abort_test.go](abort_test.go)用真实TCP验证缺Auth4协商、方向/目的/来源授权、字段及结果拒绝、原Session/Proxy不可变应答关联、远端Origin-State不回显、有限槽/字节预算、等待/响应取消、晚答一次完成、结果保留到端口消费或关闭、Respond/Close竞态。混合CCR/ASR测试显式启用两方向策略，证明ASA后同Session CCR仍保持pending且没有自动TERM；清理时由应用显式Cancel。

IEC仍为默认；`-session`只用于abort，其他flow拒绝非默认值；abort拒绝非默认`-term-result`及未选结果。正常流程DPR/DPA关闭，所有出口均有独立有限Close/端口Close/Wait并检查零资源。

[第85批验收](../../spec/evidence/batch-85-ro-abort-example.json)和[公开API消费](../../spec/evidence/batch-85-ro-abort-public-api.json)包含100项累计Ro示例测试（本批新增32项）、20轮race2000、33项已用API查询及隔离模块运行。既有库协议定义/运行时不变，未新增独立协议栈互通或真实OCS接受声明。下一P7有界工作是公开支持矩阵/API覆盖收尾，明确既有Ro余额、退款、私有绑定、恢复入口尚需哪些消费证据；不重新开放全协议字段补齐。

## 公开恢复接线与五操作消费（第88批）

```sh
CGO_ENABLED=0 go run ./examples/ro -flow recovery
CGO_ENABLED=0 go run ./examples/ro -flow recovery -recovery-scenario restart
CGO_ENABLED=0 go run ./examples/ro -flow recovery -recovery-scenario no-answer
CGO_ENABLED=0 go run ./examples/ro -flow recovery -result 2001/4012
```

`recovery.go` 使用公开 API 实际执行 DIRECT_DEBITING：PrepareRoRecoveryRequest → 应用确认确切意图 → SubmitRoRecovery → 原 Completion → 检查 CanContinue/NextTrigger → 显式第二段。error 场景先收到 E3004，再在同 Node 续段；restart 在消费 E3004 后关闭并 Wait 旧 Node，再向新 Node 导入原记录；no-answer 让首段 Tr 到期，关闭旧端点后导入可能已发送的记录。重建端点仍是同一合成 OCS 身份，不是跨 OCS 迁移。

每段最多一次发送，最多三段/三次、记录最长一分钟；示例实际消费两段。初始错误仍保存在独立 Result 中；第二段不覆盖它。返回的 CanContinue 只代表协议资格，示例明确决定继续一次，没有无限重试。成功/拒绝/服务拒绝/E结果分别保留，顶层2001不改写服务4012。完整 CCR body、原 E2E 保持，错误后续发 T=0，无答可能已写后的续发 T=1，累计 ReservedThrough=2；新 Node 的连接/HBH 空间独立，不要求跨连接 HBH 数值不同。

所有确认仅在本进程内：摘要绑定确切 record/plan/progress；没有文件、存储接口或后台队列。示例先关闭并加入旧 Node，再声明 SenderFenced/ProgressConfirmed/ClockConfirmed。这是可检查的本地演示，不证明跨进程隔离、可靠提交或真实 OCS 去重。真实应用须提供自己的确认、可信历史与时间保证。

`consumption.go` 的“五操作”严格对应原业务 CHECK/CHARGE/RESERVE/COMMIT/CANCEL，依据主流程审计§33.1处理 TX_TIMEOUT、CONNECTION_DOWN、DIAMETER_TMP_ERROR。它们不是 CHECK_BALANCE/DIRECT_DEBITING/REFUND/INITIAL/TERMINATION 的别名。配置产生的业务状态保存在独立 decision，完整 Exchange/DeliveryEvidence 不改写。SYSTEM_FAILURE 采用明确保守迁移：返回 SYSTEM_FAILURE，保留应用提供的 ErrorCause/LogLSI，不臆造未确认的旧私有成功映射。

DDFH helper 仅演示具名 combined DIRECT_DEBITING 的消费边界：缺失使用配置、显式0/1覆盖、非法值拒绝；未验证答案拒绝，E/无答不虚构远端值；CHECK/REFUND/ECUR不继承这项策略。它是本批单元消费验收，原第57批的真实线路验证保持有效，不声称本批新增私有OCS互通。无备用时返回的是“请求应用保存”的决定，未执行保存。

`recovery_test.go` 验证真实TCP三场景×五种最终结果、Tx后晚答/显式取消/一次Completion与业务通知、意图未确认/变更、导入声明缺失、终态、回退、损坏记录和额度拒绝；所有退出路径归零。`consumption_test.go` 独立核对五操作×八组本地原因/配置、诊断与原始证据保留、DDFH边界。

公开恢复示例不启用通用DCCA/ASR/SCAP1恢复，不新增库API、协议字段、Profile或持久化。其他绑定/跨目标的外部消费、六模板私有映射、真实部署、发行包安装与持久化可靠性仍按spec单列。独立模块验收使用本地源码replace，不能冒充安装已发布版本。

## combined 双目标公开消费（第89批）

```sh
CGO_ENABLED=0 go run ./examples/ro -flow compat -compat-action balance
CGO_ENABLED=0 go run ./examples/ro -flow compat -compat-action refund -result 4012
CGO_ENABLED=0 go run ./examples/ro -flow compat -compat-action term
```

`compat_recovery.go` 显式编译 Base+combined，与标准Ro的4/272解释互斥。本例创建一个CTF、两个直达合成OCS以及两个独立Peer；未实现DRA转发或目标发现。支持 debit/balance/refund/initial/term 五种**协议请求形态**，区别于第88批 CHECK/CHARGE/RESERVE/COMMIT/CANCEL 五业务操作。固定终端、应答来源、共同ReplayDomain、每目标EvidenceID和TERM连续性声明均由示例配置提供。

第一目标返回E3004，应用先消费原始Result，再显式将下一段指向冻结的backup目标。保持原CCR完整body（仅Destination可变）、E2E、Session、CC类型/编号、金额/负指数/币种、显式零单位及opaque token；错误后续发T=0，累计额度2。每段一个attempt，无自动failback。最终2001/4012/5012/E3004分别保留；余额2001+NO_CREDIT(1)表示查询完成但余额不足，不改成额度授权。

TERM使用显式合成已有Session `ctf.example;synthetic-existing-session`，不是新建Session；本例不执行此前INITIAL，也不证明实际OCS会话连续性。应用与两个目标的SessionEvidenceID仅用于演示受检声明。refund是新的业务操作，具有独立Session和显式合成token；该退款的后续恢复保持原Session，不推导真实退款资格或1292业务映射。

组合消息的解码预留大于标准Ro示例。服务器的有限入站预算设为128MiB、Node预算512MiB；`compat_limits_test.go`按公开DecodeBound验证所选请求的预留大于原32MiB且落在新额度内。这是逻辑容量，不是RSS承诺。初版复用32MiB导致服务端接纳拒绝，现已在示例配置修正，无核心预算规则变更。

`compat_recovery_test.go`包含五形态×四结果的真实TCP验收；另验证去重域不同、TERM目标/应用连续性证据缺失、未知目标、篡改记录、回退额度拒绝且备用不收请求。每个退出路径关闭唯一CTF和两个OCS，消费/关闭原CompletionPort，加入两个Serve并验证三个Node资源归零。

全流程仅公开API、纯Go，无持久化或共享OCS账本。当前范围是combined绑定的显式错误后跨目标续段；不自动扩展为全部Profile、无答跨目标/重启迁移、SCAP1恢复或现场去重证据。独立消费模块仍使用本地源码replace，不是发行包安装验收。

## 独立绑定、无答跨目标及导入（第90批）

```sh
CGO_ENABLED=0 go run ./examples/ro -flow compat -compat-binding smsfwro -compat-action term -recovery-scenario no-answer-import
CGO_ENABLED=0 go run ./examples/ro -flow compat -compat-binding legacy420 -compat-action refund -recovery-scenario no-answer -result 4012
CGO_ENABLED=0 go run ./examples/ro -flow compat -compat-binding combined -compat-action balance -recovery-scenario restart
```

`compat-binding`在创建Node前显式选择一个4/272解释，默认combined。绑定之间不能在同一恢复记录上切换，也不通过收到的AVP猜绑定。

| 绑定 | 可选compat-action | 明确边界 |
| --- | --- | --- |
| smsfwro | debit、initial、term | debit使用原IEC构造；不接受balance/refund |
| legacy420 | debit、balance、refund | 原EVENT货币/余额/退款构造；不接受initial/term |
| combined | debit、balance、refund、initial、term | 第89批五形态，保留组合绑定自己的校验 |

| recovery-scenario | 首段及继续方式 | 备用段T位 |
| --- | --- | --- |
| error | 消费E3004，同Node显式续段 | 0 |
| restart | 消费E3004，关闭/Wait全部旧Node，在新Node导入 | 0 |
| no-answer | 原OCS收到请求但不应答，Tr到期后同Node显式续段 | 1 |
| no-answer-import | 同样产生WrittenNoAnswer，关闭/Wait全部旧Node，在新Node导入 | 1 |

这四种场景均显式选择冻结的backup目标，每段一次尝试、累计ReservedThrough=2。保留完整原请求（只改变已授权Destination）、E2E、Session、CC类型/编号和opaque字节；不要求独立连接的HBH数值互异。无答的MayHaveReachedPeer/ReplayPending不被改成未发送；同Node无答场景在备用完成后还接收原OCS晚答，验证没有额外Completion。

导入使用首段真实Result.Checkpoint和进程内记录，先关闭并加入旧发送者，再确认SenderFenced、ProgressConfirmed、ClockConfirmed并设置CheckedAt。这里的restart仅演示Node重建，没有进程崩溃恢复、文件、数据库、可靠历史或时钟服务。两个OCS仍为合成直达端点，未实现共享账本/远端去重，不证明生产迁移安全；TERM仍使用显式已有Session声明，未推导之前的INITIAL授权。

`compat_bindings_test.go`新增156个正常矩阵用例，连同第89批20个用例覆盖11个具名操作×4场景×4最终结果（2001/4012/5012/E3004）。余额NO_CREDIT原样保留。另在三个绑定各验证八种新Node导入拒绝：缺发送者隔离、进度或时钟确认，检查点回退、目标去重域或证据变化、Bundle绑定变化、禁用导入。拒绝时没有Outstanding和备用入站请求，正常及失败路径均检查所有Node资源归零。

本批公开消费保持纯Go，仅复用PrepareRoRecoveryRequest/SubmitRoRecovery等既有API。源代码复制到独立module以本地replace验收，不冒充已发布包安装。第88批标准Ro恢复和五业务映射、第89批授权反例继续回归；私有六模板/真实OCS接受、业务账本和发布门保持独立。

第90批独立模块race还暴露既有混合ASR/CCR示例的握手时序：Dial返回时，接收端可能尚未完成CEA写屏障。共享connectPair现以有限ctx等待两端Ready，再交给双向调用者；没有放宽库许可，Ready仍只是可变化的状态快照。已有混合测试重复50轮race复核，首次失败与最终回归日志保留在验收证据中。
