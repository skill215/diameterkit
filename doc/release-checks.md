# 双 module checkout 发布检查

第91批提供可重复的本地/CI入口：[release_check.py](../spec/tools/release_check.py) 和 [GitHub Actions](../.github/workflows/ci.yml)。它验证当前checkout，尚不代表已发布版本、远端工作流已通过或完整发布验收。

## 运行

在Linux/amd64安装 [.go-version](../.go-version) 指定的Go（当前1.26.10）、Python 3和race所需的C编译器，并保证内核支持本地SCTP socket。从仓库根目录运行：

```sh
python3 spec/tools/release_check.py
# 基础四组如下；第92批另有fuzz/stress/bench：
python3 spec/tools/release_check.py --suite static
python3 spec/tools/release_check.py --suite pure
python3 spec/tools/release_check.py --suite race
python3 spec/tools/release_check.py --suite cross
```

默认日志和机器结果位于忽略目录 `.cache/release-check/<suite>/`，也可用`--output`指定目录。每次命令失败立即返回非零，记录失败步骤；Go测试必须实际通过且不能跳过测试。无测试文件的包不算跳过测试。SCTP内核不可用时验收失败，不能据Windows stub或`go build`声称真实收发通过。测试均`-count=1`，不使用缓存的成功结果或自动重试。

| 组 | 执行内容 | 证明边界 |
| --- | --- | --- |
| static | 工具失败路径测试；11份生成输出逐字节`-check`；生成确定性；Linux/Windows各25包API快照；两个module的vet | 漂移检测，不自动接受API变化；故意损坏的临时生成文件/API副本必须失败且不被重写 |
| pure | 核心和adapter分别`CGO_ENABLED=0 go test -json ./...` | 当前Linux/amd64真实运行，包括SCTP；根目录测试不会自动包含嵌套adapter |
| race | 核心和adapter分别`go test -race -json ./...` | race工具需要CGO_ENABLED=1及C编译器；生产代码依然不引入cgo，pure检查独立证明可无CGO运行 |
| cross | 两module分别CGO=0构建linux/arm64、linux/386、windows/amd64 | 仅编译，不执行；Windows/386 SCTP仍是unsupported stub |

`GOTOOLCHAIN=local`和精确工具链检查防止静默下载另一个Go版本。更新Go时应同时审查API快照，不直接让CI刷新它。源码生成仍由`go generate ./...`显式执行；验收只比较，且确认生成文件清单没有漏项。

## 当前源码与独立发行的区别

adapter的go.mod仍锁定未发布的核心`v0.0.0-diameterkit.batch85`。干净CI不能把它当作公开可下载版本，也不能用旧缓存冒充当前核心。

检查入口在临时目录创建go.work，显式纳入两个当前module，并只为adapter要求的那个核心版本添加指向checkout的替换。仅use两个目录仍可能查询旧版本go.mod，故需要该版本化替换。随后检查module graph：核心与adapter都必须是当前目录的Main module，SCTP必须为go.mod锁定版本且无替换，新增依赖必须经审查。临时workspace在退出时删除，不写仓库go.work、不改两个go.mod、adapter go.sum或API基线。

SCTP依赖由已提交go.sum校验。CI使用官方Go代理和校验数据库；本地可设`GOPROXY=off`使用已校验缓存，缺依赖就失败。本批验收使用仅预置SCTPv0.8.0下载缓存的新module cache，确认不含batch85核心。未重新打包或冒用旧候选版本。

API快照工具新增显式`--workspace /absolute/path/go.work`选项供此入口使用；不传时保持原`GOWORK=off`独立module检查行为。它仍离线读取，不下载依赖。workspace选择不改变快照格式或现有基线内容。

发布前仍须将离线核心依赖替换成真正可引用的不可变版本，并以GOWORK=off、无replace、干净缓存验收安装。该步骤和版本政策、较长fuzz/压力/峰值资源/性能、实际目标接受均未在本批关闭。

## CI与证据

工作流在push、pull_request和手动触发时运行四组Ubuntu24.04检查；仅授予contents:read，checkout不保留凭据，三个Actions固定到审核过的提交。各组限时20分钟，组间不因另一组失败而取消；同一ref的新run可取消旧run。每组无论成功失败都上传该组日志和summary，保留14天。仓库分支保护可选择`checkout-static`、`checkout-pure`、`checkout-race`、`checkout-cross`四个检查。

本批只在WSL执行了同一入口并校验workflow语法；没有推送、触发远端run或修改分支保护，不能把本地通过描述成GitHub托管CI已通过。SCTP取消/到达测试曾捕获队列跨轮干扰：现按轮隔离监听队列并增加真正fresh accept收发，保留同监听器复用的独立测试，未放宽库行为或错误许可。

实现依据：[Go workspace规则](https://go.dev/ref/mod#workspaces)、[setup-go配置](https://github.com/actions/setup-go)、[checkout v6.1.0](https://github.com/actions/checkout/releases/tag/v6.1.0)。Actions固定提交和实际本地结果见第91批验收证据。

## Bounded load checks (batch92)

The same runner now accepts `--suite fuzz`, `--suite stress` and `--suite bench`; `all` includes them. Five fuzz targets use30s/two workers each, pressure workloads have fixed windows/counts and pure/race repetitions, and eight benchmarks use200 iterations/five samples. See [workload definitions and measurement limits](load-checks.md). The workflow now has seven suites; this adds local measurement evidence, not a hosted-run, production-capacity or performance-improvement claim.

独立发行须另跑第93批[真实版本安装门禁](version-install-policy.md)。它使用官方代理/校验库、两个空缓存消费者和GOWORK=off，禁止replace，与本文临时workspace检查分开。目前核心/adapter候选未发布，实际下载失败，不宣称安装通过。
