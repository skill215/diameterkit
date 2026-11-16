# 核心与 SCTP adapter 版本/API 政策及安装验收

第93批的机器政策见 [version-policy.json](../spec/release/version-policy.json)，安装入口见 [release_install.py](../spec/tools/release_install.py)。当前是 **v0 开发期、尚未发布**，不是 v1 稳定 API 或生产投产承诺。

## 两个 module 独立版本

| module | Git 标签 | 第一组提议版本（未发布） |
| --- | --- | --- |
| `github.com/skill215/diameterkit` | `v0.MINOR.PATCH[-rc.N]` | `v0.1.0-rc.1` |
| `github.com/skill215/diameterkit/adapters/sctp` | `adapters/sctp/v0.MINOR.PATCH[-rc.N]` | `v0.1.0-rc.1` |

两者不要求永远同步版本。adapter 必须精确依赖已发布并验收过的核心版本；消费时 Go MVS 选择了另一核心版本也应失败。上游 SCTP 保持 `github.com/skill215/sctp v0.8.0`。当前 adapter/go.mod 中 batch85 是历史离线占位，不能作为 release；在真实核心可下载前不伪造新的 go.sum，也不把未发布候选写成已安装依赖。

在本项目 v0 政策下：patch 只作兼容修正；新增 API 或有意破坏兼容均升 minor，并列明迁移；rc 内容变化必须递增 rc.N。所有已发布 tag/version 禁止移动、覆盖或复用，错误版本在后续版本 retract，保留旧内容。第一组候选只是待审发行顺序，不是实际标签或下载保证。版本与 module 子目录标签规则依据 [Go Modules Reference](https://go.dev/ref/mod#vcs-version)，校验依据 [Go module authentication](https://go.dev/ref/mod#authenticating)。

## API 评审政策

核心与 adapter 的所有公开非 internal 库包都在范围内，包括生成声明、结构体字段、接口、error 语义、所有权/取消/资源契约。协议定义里的 slice version 独立于 module SemVer；协议行为改变也必须记录兼容性，不能仅改 module 版本遮蔽。

任何 API/行为变化记录：受影响 module/package、前后签名或行为、兼容性分类与版本选择、迁移和支持限制、测试与审核后的 API 快照。增加接口方法、字段类型变化、要求调用方新增必需配置等按潜在破坏处理。v1 需要独立的范围与支持承诺审查。

`spec/release/public-api.json` 是 Linux/Windows 构建选择下的文档及源码基线，漂移必须显式评审；CI 不自动刷新。相等不证明语义兼容，文档/实现变动导致不相等也不自动等于 breaking change。安装门禁使用当前审核基线核对安装包，不能借测试一个较旧版本代表当前实现。

## 可重复的真实安装验收

在原仓库根目录，使用 `.go-version` 的 Go、Python 3、race 所需 C 编译器及支持 SCTP 的 Linux/amd64 内核。输出目录必须尚不存在，防止旧成功记录混入本次失败：

```sh
python3 spec/tools/release_install.py \
  --core-version v0.1.0-rc.1 \
  --adapter-version v0.1.0-rc.1 \
  --output .cache/published-install-rc1
```

截至第93批，上述版本不存在，命令应非零退出并保留实际失败日志。发布后使用同一精确版本和新的输出目录重跑。`latest`、分支、pseudo-version、batch85、v1、构建元数据、非法 rc 均被当前政策拒绝。工具不会发布、提交、打标签或改仓库 go.mod/go.sum。

验收分两个完全独立的临时消费者：核心单独消费、核心＋adapter 消费。每个都在仓库外创建新的 GOMODCACHE/GOCACHE/GOPATH，记录初始空目录；不继承已有 workspace、私有代理豁免、vendor 模式或 GOENV。固定 `GOWORK=off`、`GOTOOLCHAIN=local`、`CGO_ENABLED=0`，下载仅用 `https://proxy.golang.org` 和 `sum.golang.org`，无本地代理/direct 回退。每个消费者都重新下载依赖，不能共享上一消费者的预热缓存。

必须全部通过：

1. 精确版本真实下载，记录模块 Sum、GoModSum；核心/adapter 必须提供与远端标签一致的 Origin commit/ref；GitHub 标签解析到 commit，验收末尾再次核对标签未移动。
2. 消费者及下载模块的 manifest 没有 replace/exclude；核心无外部依赖，adapter 只有精确核心与上游 SCTP，图中所有依赖来自本次缓存且没有 MVS 升级。
3. 安装包公开 API 的文件集合、源码哈希和 go doc 与审核基线一致；核心消费者不含 adapter 依赖。
4. 从源码复制的是**消费者测试**，不是库实现：外部绑定契约与 Ro TCP/恢复测试；adapter 消费者另有双向 SCTP 20/256/4096 字节收发及预算检查。纯 Go/race 都实际执行，零 skip。
5. 下载完成后 Go 验证阶段禁网（最终标签复核仍访问 GitHub），检查依赖不需要 cgo、无 checkout 源文件参与；`go mod verify` 成功，测试不修改消费者 go.mod/go.sum。

race 使用 CGO=1 仅用于 Go race 工具。运行库仍纯 Go、无持久化；缓存、日志、测试夹具是构建验收材料。进程外 Go 缓存不属于库的业务持久化边界。

[Published module installation workflow](../.github/workflows/release-install.yml) 通过显式版本输入手动运行，关闭 setup-go cache，失败也上传证据；不与 checkout 七组 CI 的绿色状态混用。人工审核基线的 checkout 和所验发行必须一致。工具自身的离线单元测试或 checkout 消费探针通过，只证明工具和探针可运行，不代表真实发行安装通过。

## 从当前状态到首个候选

1. 审核当前源码、API/支持限制和待提交文件，完成 checkout CI，准备不可变核心候选提交；发布核心 `v0.1.0-rc.1` 标签。
2. 从官方代理/校验数据库确认核心下载后，才把 adapter/go.mod 的 batch85 换为真实核心版本，生成真实 go.sum 并移除历史离线校验和；重新运行双 module 验收。
3. 在单独的 adapter 发行提交上发布 `adapters/sctp/v0.1.0-rc.1` 标签。两个标签可指向不同提交，adapter 的依赖必须指向前一步已存在的核心。
4. 执行上述两个独立冷消费者门禁，保留 commit/checksum/API/测试证据；任一失败保持未验收。不得移动已发布核心标签来补 adapter 或修复文档。

2026-11-16 的只读查询显示远端没有标签，Go 代理核心列表为空、adapter 无版本，旧 batch85 返回 unknown revision。当前真实安装出口缺少发布动作；本批不把工具完成标成安装通过。托管 CI、私有/目标证据与部署级容量仍保持各自范围。
