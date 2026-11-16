# 有界 fuzz、压力与性能基线

第92批复用[发布检查入口](release-checks.md)，新增三个suite；生产库仍纯Go、无持久化。所有命令在WSL/Linux amd64原仓库根目录执行，使用 `.go-version` 指定Go和已校验的SCTP依赖。

```sh
python3 spec/tools/release_check.py --suite fuzz
python3 spec/tools/release_check.py --suite stress
python3 spec/tools/release_check.py --suite bench
# 包含双module static/pure/race/cross及上述三组：
python3 spec/tools/release_check.py --suite all
```

每组在`.cache/release-check/<suite>/`输出日志和summary；可用`--output`指定新目录。fuzz.json、pressure.json、performance.json分别记录执行量、资源样本和性能样本。CI矩阵使用同一命令并归档这些文件；本批只报告WSL本地执行，没有托管run声明。

## 固定工作量与停止条件

| 检查 | 固定参数 | 失败条件 |
| --- | --- | --- |
| codec FuzzDecode / FuzzDecodeAnswer / FuzzCanonical | 每目标30s，GOMAXPROCS=2，parallel=2，单次最小化5s，包超时2m | panic、性质违反、超时、无实际fuzz执行或命令失败 |
| Sh-Data FuzzParse、adapter FuzzSCTPModeSelection | 同上；5目标共150s目标执行预算 | 同上；不把仅运行种子算成fuzz验收 |
| TCP Ro压力 | 窗口1/8/16，各16波；每波并发提交至满窗、验证拒绝、应答/取消、保留完成结果、消费并清零；纯Go/race各20轮 | 超出配置逻辑容量/槽、错误接纳、重复/缺失结果、未清零或超时 |
| SCTP压力 | 4对真实连接、每对128帧，20/256/4096字节轮换；纯Go/race各20轮 | 字节不一致、写入错误、固定BufferBytes超过1MiB、关闭后仍可写或超时 |
| benchmark | 每样本200次，5份样本，cpu=1、CGO=0，报告ns/op、B/op、allocs/op及适用的MB/s | 缺样本、工作量变化、非有限测量值、命令失败；没有凭空设定跨主机耗时阈值 |

TCP每轮包含400个请求、96次满窗拒绝和104次确定的Cancel；同一轮的窗口只使用一个Peer/一个CompletionPort。每个Node配置1GiB逻辑上限，出/入站class与port为256MiB，窗口最多16。结果尚未消费时，已完成请求仍必须占用port；消费后出/入站动态账目为零，Close/Wait后整个ResourceSnapshot为零。这里的GiB是有限逻辑配置，不表示实际分配了该数目的Go内存。

SCTP每轮合计512帧和8个transport，写/读由4个工作goroutine并行推进；没有新增业务重试、缓存或持久化。原SCTP取消/复用和其他完整测试仍在pure/race套件中执行。真实SCTP不能跳过；跨平台build不替代原生收发。

Go fuzz使用原有种子和本机Go fuzz缓存，变异次序和执行量可能不同。可用独立GOCACHE复核冷缓存行为，但构建时间也会变化；不能承诺相同输入顺序。失败时Go产生的`testdata/fuzz/<Target>/<hash>`会同时复制进输出目录的`fuzz-corpus/`，即使命令失败也保留供CI artifact取回；源回归输入不会被删除。修复后可用`go test <包> -run '<Target>/<hash>'`定向重放，Go库不接管运行期存储。

## 资源度量的含义

`pressure.json`区分三种数值：

- **逻辑预留峰值**：在满窗和完成保留等受控阶段读取Node.Resources，含Fixed/Outbound/Inbound等承诺；不是RSS。
- **SampledHeapAlloc**：阶段采样的Go进程HeapAlloc最大值，不是连续采样所得的绝对堆峰值，也不是请求独占内存。
- **ProcessRSSHighWater**：Linux `/proc/self/status` 的VmHWM，包含整个测试进程、runtime、race开销和此前轮次。测试设置1GiB进程RSS护栏；它是这组测试的保守失败上限，不是库的生产容量承诺。重复窗口的HWM可能单调上升。

资源断言与测量在压力测试内进行，排除Go编译器进程；配置的逻辑限制和RSS护栏不能互相替代。此次没有长期驻留/多租户容量、慢下游或生产网卡峰值声明。

## 性能样本的比较方式

测量项为5个codec操作、Ro TCP请求/应答完整一次消费、256/4096字节SCTP单向本机发送加owned frame接收，共8项。连接建立、Profile编译和fixture读取不计入计时；热路径的构造/身份分配/验证等是否包含以各benchmark源码为准。codec Append重用调用者dst，Decode/Canonical含owned输出分配；SCTP包含接收frame分配，单向本机测试不能称为远端RTT。

`performance.json`保留全部5份样本，以及各单位的min/median/max、CPU/内核/Go/cpu参数和工作量源码哈希。ns/op是每份benchmark的平均操作时间，样本median不是单次请求p50，更不是p99。首次结果作为后续对照基线，不宣称加速或稳定SLO；比较时使用相同工作量、硬件、Go、GOMAXPROCS/CPU亲和与负载，并另行审查统计差异，不自动覆盖已审核基线。

这些检查关闭第92批所选的有限验收切片；更长时fuzz、部署级饱和曲线/峰值/吞吐、真实网络和发行安装仍需独立证据。
