# 公开绑定消费契约

运行 `CGO_ENABLED=0 go test ./examples/contracts`。此目录只有外部包测试，无运行时服务，可把[contracts_test.go](contracts_test.go)原样复制到独立module，仅依赖已安装的diameterkit核心版本。无仓库fixture、internal或根包私有helper。

43项检查覆盖12个Profile入口、五种4/272绑定两两拒绝混装、独立abort组合；DCCA单服务EVENT余额/扣费/退款成对消息及退款新Session；标准SMS原2022 token的absence/empty/unknown、复制和回显边界；legacy420/combined显式私有token/金额构造与smsfwro/combined ECUR连续性。

所有报文路径使用公开codec完成包络、encode/decode及原请求/来源/Context检查。DCCA成功余额含NO_CREDIT不等于额度授权；标准SMS的token helper不允许发送标准退款；1292不能冒充2022。这是合成离线消费，不是独立Diameter栈/OCS互通或真实财务流程。

RO-06在此仅核对方法签名、无效记录拒绝和Checkpoint值，不启动恢复段、确认持久状态或管理会话。可运行恢复机制仍由原根测试证明，公开应用五操作接线仍单列缺口。

[支持矩阵](../../doc/diameterkit-support-matrix.md)和[第86批证据](../../spec/evidence/batch-86-public-surface.json)区分实现/公开消费/目标验收。本批生产源码未变，隔离消费者使用原batch85离线候选，无replace；不是公开发行或稳定API冻结。
