# Sh Ref24公开接入示例

在repo根目录运行（Go1.26+，纯Go，无外部服务）：

```sh
CGO_ENABLED=0 go run ./examples/sh
CGO_ENABLED=0 go run ./examples/sh -result 5012
CGO_ENABLED=0 go run ./examples/sh -result 3004
```

两个独立Node经127.0.0.1随机端口真实TCP握手，各发一笔UDR并处理另一端UDR。2001携带显式合成的Ref24 SMSRegistrationInfo，5012为正常失败UDA，3004为E错误答案。最后DPR/DPA关闭，并打印双方零资源快照。`-timeout`默认10秒；所有返回路径另有5秒有限Close/端口关闭/Wait清理。

这是单进程双端点API演示。为展示发送和应答，两端均授权UDR；实际部署应按AS/HSS角色配置允许方向，不据此推导HSS应向AS发UDR，也不提供HSS数据库、订阅管理或DRA部署发现。

[config.go](config.go)使用Base+Sh Profile、显式Vendor10415/Auth16777217能力、默认HeaderApplication绑定、固定终端/来源、NoCache、单次attempt。每Node只有一个完成端口/一个未消费槽；节点、传输、标识、class、Peer、入站/答案和字节预算都有限。预算是逻辑额度，示例预留值不是生产容量或RSS测量。HeaderApplication能力在握手协商，不能把NewNode成功视为已协商；缺Sh能力的双端测试验证握手失败。

[main.go](main.go)展示Serve/Dial、Submit、Peer.Receive、Inbound.Respond和CompletionPort.Receive。Session由公开NewSessionID分配；应答包络由Respond恢复。Submit的ctx控制接纳，接纳后需要Node.Cancel终止请求；Receive的ctx只停止等待。未消费完成仍保留额度，Node.Wait不代消费，必须接收结果或关闭端口。

成功UDA通过`shdata.MarshalSMSPull`构造；接收方显式调用`Parse`和`Check(SMSPull)`。XML使用独立512字节/12层等有限限制，不计入Node预算，也不等于完整XSD、订户权限或业务状态校验。所选号码/SC/SBI均由示例显式提供；不推导默认值。报文层Validation.Complete不意味着XML有效，收到答案也不意味着业务成功。

[main_test.go](main_test.go)只调用公开库API，覆盖双向三种结果、能力协商失败、策略与传输预算配置、UDA缺数据/失败带数据/错误VSAI拒绝、XML显式解析与限额、独立槽/字节背压、视图隔离、取消和晚答一次完成以及关闭后归零。测试不会自动扩大PUR/SNR/PNR、其他Data-Reference或TLS/SCTP示例范围。

应用版本和JSON生成方式见[定义指南](../../doc/diameterkit-application-definitions.md)；本批独立消费module和验证结果见[batch80](../../spec/history/batch-80.json)。
