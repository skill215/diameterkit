# SCAP1双向公开API示例

需要Go1.26+。在仓库根目录运行：

```sh
CGO_ENABLED=0 go run ./examples/scap1
CGO_ENABLED=0 go run ./examples/scap1 -legacy-advertisement
CGO_ENABLED=0 go run ./examples/scap1 -result 5012
CGO_ENABLED=0 go run ./examples/scap1 -legacy-advertisement -result 3004
```

程序在`127.0.0.1:0`监听，通过`Node.Serve`和`Node.Dial`建立真实TCP连接。两个独立Node使用`left.example`与`right.example`身份，双方各发送一笔ACR并处理一笔请求；同一个进程内完成后退出。它不需要OCS、证书、数据库或C/C++库，也不运行旧sms-fw进程。TLS/SCTP适配器的验收仍见原批次，本示例只演示TCP公共接入。

正常输出先显示两个方向的结果，关闭后左右ResourceSnapshot均为零。`-timeout`默认为10s，参数必须为正；错误路径仍以独立5s期限关闭端口、Node和listener。

## 接入流程

[config.go](config.go)给出完整配置：Base与SCAP1 Profile显式编译，LocalCapabilities保留严格的Vendor193/Acct19302，每方向独立声明SCAP1Accounting及固定对端来源/目的。监听地址不充当DiameterIdentity。此例为直接端点；接DRA时，应分别配置邻接Peer与授权终端，不能照搬它们相等的假设。

零值广告是标准模式。`-legacy-advertisement`在**两端**Peer上选择已审核旧双组布局；不能直接把双Auth/Acct塞进LocalCapabilities。标准端点拒绝旧双组，选定兼容端点可接收标准或审核旧布局；一端选旧模式并不意味着另一端会自动兼容。

[main.go](main.go)展示`NewSessionID → scap1.NewEvent → Peer.Submit → Peer.Receive → Inbound.Respond → CompletionPort.Receive`。`NewSessionID`只分配标识；EventInput使用明确的合成号码、文本和固定wire Time，库不查询订户、推断收费方或生成业务时间。

处理方读取独立请求视图并明确选择结果：2001与5012构造ACA，3004使用`pdu.ErrorAnswer`产生E包络。正常ACA填写原Accounting-Record-Type/Number；Session、Proxy、头标识和本地Origin由Respond依据私有原请求补齐。`AnswerReceived`和`Validation.Complete`只说明协议应答/关联完成，5012或E/3004不代表业务成功，更不代表扣费已发生。

本例没有缓存、自动重试或计费状态，显式`NoCache:true`。每个到达的ACR独立交付；需要有限重复重答时按既有DuplicatePolicy配置，但不能把进程内缓存当作持久账本。SCAP1仍SingleAttempt；私有ACA和相应恢复仍受原证据门限制。

## 有限预算与清理

| 项目 | 示例上限 / 行为 |
| --- | --- |
| 帧、单字段、AVP、深度、重复 | 4096B、1024B、64、8、16；请求/答案各1024B |
| 每Node总额、控制/传输、标识 | 256MiB、32MiB、4096B；TCP连接与listener使用适配器公开BufferBytes声明 |
| 出站/入站class | 各64MiB、各2槽；每Peer出站/入站各32MiB、各2槽，独立分区 |
| 完成端口 | 每Node1个，1槽、32MiB；队列中未消费的结果继续占额度 |
| 重放/持久化 | 无Route恢复、无Ro恢复记录、无DuplicateClasses、无存储 |

这些是可运行示例的保守**逻辑承诺上限**，不是吞吐建议、启动时堆分配量或RSS/kernel内存上限。公开API要求显式给额度，本例没有最小预算估算器或生产容量预测。

成功路径调用Shutdown，使用DPR/DPA结束连接；所有路径最终执行Close、CompletionPort.Close和Wait。Node.Wait只等待库内部清理，不替调用方消费端口结果；只有端口也关闭，固定及完成预算才全部释放。关闭端口会丢弃未消费完成结果，实际应用应先处理需要保留的结果。本示例在成功路径已消费全部结果，错误路径直接终止演示，不试图恢复业务。

## 可执行API核对

[main_test.go](main_test.go)只调用公开库API，不使用根包私有字段、fake transport或测试helper，覆盖：

- 两种广告模式×2001/5012/E3004，真实双向请求及关联、DPR/DPA与最终资源清零。
- 缺能力/错误Vendor/双用途本地能力、未绑定方向、未固定终端、未知广告和不足控制预算拒绝；严格/旧广告混配握手失败。
- 完成槽满与独立字节预算不足时Submit拒绝且不交付远端；Consume前后额度变化。
- Submit的ctx仅管接纳；接纳后取消该ctx不会取消请求。CompletionPort.Receive的ctx只管等待；实际取消调用Node.Cancel。
- 错误记录号、取消的Respond不提交；修改公开请求视图不改变答案关联；取消后晚答不产生第二次完成。
- Node.Close/Wait后未消费完成仍占预算，关闭端口后释放；已取消运行ctx仍可完成有界清理。

```sh
go test ./examples/scap1 -race -count=20
```

本批还把这些源文件复制进一个只依赖离线候选、无replace且GOWORK=off的独立消费module，编译并运行相同验收；具体证据见[batch79](../../spec/history/batch-79.json)。这验证已使用的公开API，不是整个库API冻结或公开版本发布。
