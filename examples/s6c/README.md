# S6c SRR/SRA公开接入示例

在repo根目录运行（Go1.26+、纯Go，无外部服务）：

```sh
CGO_ENABLED=0 go run ./examples/s6c
CGO_ENABLED=0 go run ./examples/s6c -result 5012
CGO_ENABLED=0 go run ./examples/s6c -result 5001
CGO_ENABLED=0 go run ./examples/s6c -result 3004
```

GMSC与HSS是两个独立Node，通过127.0.0.1随机端口真实TCP连接。GMSC只授权发送SRR，HSS只授权接收SRR并返回SRA；HSS不创建完成端口。示例只运行一笔查询，随后DPR/DPA关闭。`-timeout`默认10秒，错误/取消路径另有5秒有限Close/端口Close/Wait清理，最后双方资源归零。

| 参数 | 实际答案 | 所选数据 |
| --- | --- | --- |
| 默认或2001 | 正常SRA，Result-Code=2001 | 显式合成IMSI及MME Name/Realm/Number |
| 5012 | 正常SRA，Result-Code=5012 | 无成功路由数据 |
| 5001 | 正常SRA，Experimental-Result=10415/5001 | 与基础5001区分，输出带experimental命名空间 |
| 3004 | E错误包络，Result-Code=3004 | 回显原Session/头关联 |

所有号码、IMSI、节点与结果均由示例显式选择，库不查询订户、不决定未知用户、不从MSISDN生成IMSI。成功SRA不表示已投递SMS，也不证明返回MME可以连通。本例没有实现DRA进程或真实HSS服务。

[config.go](config.go)展示完整Base+S6c Profile、CER/CEA Vendor10415/Auth16777312能力及HeaderApplication绑定。命令键直接读取生成类型`(&s6c.SRR{}).CommandKey()`，不依赖不存在的`s6c.ApplicationID`常量。S6c正常报文不照抄Sh VSAI；能力广告与消息字段分开。GMSC固定Destination-Host/Realm并授权HSS答案来源，HSS授权GMSC请求来源。DRA部署须把相邻Peer与终端来源分开配置。

普通MSISDN SRR使用有界TBCD编码的MSISDN/SC，SM-RP-MTI显式为0；原请求未请求Additional节点，返回Additional会被关联检查拒绝。本例只选真实IMSI成功和MME节点，不启用Router token、MCC/MNC-only、IMS/T4、IP-SM-GW角色UNRI或重传。现有规则已支持的其他范围见[应用定义指南](../../doc/diameterkit-application-definitions.md)，不因本例省略而撤销。

节点、传输、标识、class、Peer、完成端口都有有限额度；仅发送端不配置入站额度。当前Node配置仍要求公共出站class/Peer预算，即便HSS没有出站业务策略；这不授予发送权限。预留值是示例逻辑上限，不是生产容量或RSS承诺。NoCache、SingleAttempt均显式沿默认边界；加载Profile不自动宣告能力或授权命令。

[main.go](main.go)只使用公开Serve/Dial、Submit、Receive、Respond、Shutdown/Close/Wait。Submit ctx只控制接纳，完成等待ctx只控制等待，实际取消用Node.Cancel；未消费完成仍占额度，Wait不代消费，必须Receive或关闭端口。

[main_test.go](main_test.go)覆盖四类结果、两端缺能力的握手失败、错误配置、缺IMSI/混合结果/失败带成功数据/不完整MME/未经原请求许可的Additional应答、缺请求身份/SC/SMEA与错误目的、槽/字节背压、取消/晚答一次完成及关闭归零。独立消费module验收见[spec进度](../../spec/progress.json)。新例只验证本机TCP公开接入，不宣称真实DRA/HSS或新增TLS/SCTP部署验收。

第87批新增[纯构造、解释器与薄客户端](../../doc/s6c-routing-helpers.md)。示例 newRouting 已使用 NewSRR；client_test.go 在真实 TCP 上验收 Route 的 QueryRouting/SubmitRouting、四类结果、取消及资源清零。它复用既有 Node/Route/CompletionPort，业务自行选择候选并发布结果。
