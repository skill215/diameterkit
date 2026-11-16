# SGd/Gdd MO与MT公开接入示例

在repo根目录运行（Go1.26+、纯Go，无外部服务）：

```sh
CGO_ENABLED=0 go run ./examples/sgd
CGO_ENABLED=0 go run ./examples/sgd -interface gdd
CGO_ENABLED=0 go run ./examples/sgd -interface sgd -result 5555
CGO_ENABLED=0 go run ./examples/sgd -interface gdd -result 5012
CGO_ENABLED=0 go run ./examples/sgd -interface gdd -result 3004
```

两个独立Node通过127.0.0.1随机端口真实TCP握手。默认SGd使用`mme.example`，Gdd使用`sgsn.example`；另一端`sms.example`在本例同时承担SMS-IWMSC和SMS-GMSC角色。先MO、后MT，各一次独立exchange，没有内部业务循环或自动重试。

| 流程 | 请求方向与策略 | 答案方向 |
| --- | --- | --- |
| MO OFR/OFA，8388645 | MME或SGSN → SMS-IWMSC | IWMSC → 原请求接入节点 |
| MT TFR/TFA，8388646 | SMS-GMSC → MME或SGSN | 接入节点 → 原请求GMSC |

SGd/Gdd共用Application-ID 16777313及同一生成Profile。`-interface`是示例应用的部署选择，不是额外Diameter头字段或库自动探测。SGd TFR明确携带MME-Number-for-MT-SMS，Gdd明确携带SGSN-Number；MO使用共同的已知IMSI形状，不拿OFR-Flags推断接口或邻接角色。合设IWMSC/GMSC只是本例配置，实际部署可以分别绑定不同终端。

| `-result` | OFA/TFA及所选字段 |
| --- | --- |
| 默认/2001 | 基础成功，带显式合成SM-RP-UI |
| 5012 | 基础失败，无本例成功载荷 |
| 5555 | 实验结果10415/5555，MO原因3、MT原因0；保留零值及原因父组存在性 |
| 3004 | E错误包络，回显原Session/头关联 |

输出明确区分base/experimental/E。TPDU为包含0和255的合成透明字节，只验证字节保留，不宣称有效SMS TPDU、实际发送或终端投递。IMSI、SC和节点号码均由示例显式提供，不查询订户或推导身份。没有持久化、业务恢复、重投时间协商、IMS/T4/Router语义或计费副作用。

[config.go](config.go)展示Base+SGd Profile、显式Vendor10415/Auth16777313能力、每端一个允许发送命令及另一个允许接收命令。目的Host/Realm、答案终端来源和请求来源分别限定，加载Profile不自动授权。这里邻接Peer就是授权终端；经过DRA时须明确区分DRA与最终IWMSC/GMSC/MME/SGSN，不把相邻Origin当成终端授权。

两个Node都有独立出站/入站class、Peer及完成端口额度，端口各一个未消费槽，Node/传输/标识/字节预算均有限。NoCache和SingleAttempt不借Profile的其他能力自动扩大。预留额度是示例逻辑上限，不是RSS或生产吞吐承诺。

[main.go](main.go)使用公开Serve/Dial/Submit/Receive/Respond及完成端口。`-timeout`默认10秒；成功以DPR/DPA收尾，所有返回路径另有5秒有限Close/端口Close/Wait。Submit ctx只控制接纳；完成等待ctx只停止等待，实际取消用Node.Cancel。Wait不代消费结果，未消费完成仍计费，最终需消费或关闭端口。结束打印两个零资源快照。

[main_test.go](main_test.go)覆盖两接口×两方向×四结果、错误方向/目的/身份及超长载荷、请求/答案终端白名单、5555缺原因/反方向原因/混合结果、Gdd拒绝MME专用TFA结果，以及双端独立槽/字节背压、可变视图、取消/晚答一次完成、未消费完成与关闭归零。没有私有helper或fake transport。独立消费module、完整回归及所用API证据见[批次验收](../../spec/evidence/batch-82-sgd-example.json)和[当前进度](../../spec/progress.json)。本例只新增真实TCP接入证据，不是实际DRA/接入节点或新的TLS/SCTP部署验收。

第89批回归修正：双向示例在 Dial 后有界等待两端 Peer.Ready，再发起任意方向的业务请求。Dial 的返回不替代接受端 CEA 写入屏障；等待受调用 ctx 限制，失败沿既有关闭路径清理。Ready 仍是快照，后续 Submit 失败必须照常处理。来源/方向/预算校验未放宽。
