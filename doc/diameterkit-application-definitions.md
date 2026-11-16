# 应用版本、定义入口与生成方式

本库使用**构建期JSON定义 → 生成Go类型/codec → 显式装入Profile/Bundle**。没有运行时XML字典加载器。找应用字段应从下表的`spec/*.json`进入；Go使用者直接导入生成的应用包，不需要部署JSON文件。这里的JSON承担命令和AVP结构定义；条件关系、上下文和业务边界还由手写Go规则约束，不能只看字典就判断完整协议支持。

## 四种“版本”不能混用

- **Application-ID**是Diameter头中的应用标识，如S6c为16777312；它不是3GPP Release编号，也不是Vendor-ID。AVP Vendor-ID、能力广告中的Vendor-ID和头Application-ID分别定义。
- **规范版本**是选定来源，如TS 29.338 V18.2.0（Release 18）。本库实现经过审核的子集，不承诺实现该版全部字段、过程或后续CR。逐项来源与修订说明在JSON的`sources`，完整依赖/分支在[主流程审计](diameterkit-mainflow-audit.md)和[字段清单](diameterkit-field-inventory.md)。
- **Profile版本**是本库切片版本，如`0.1.0-slice.22`，由JSON的`version`生成Go包的`Version`常量。名字中历史性的`ordinary`/`iec`不代表仍只实现最初切片；当前范围以规则、[支持表](../spec/README.md)和证据为准。
- **Go module版本**用于分发代码；当前核心batch85是离线验收候选，不是公开release。`spec/progress.json`的Draft版本是库行为规范进度，也不是3GPP版本。

## 首版标准应用：当前规范基线和入口

以下版本和命令取自本批实际生成输入，未因添加接入示例升级协议切片。

| 应用 | 规范基线 | 头Application-ID / Command-Code | Profile版本 | 定义输入 / 生成类型与Profile / 手写规则 |
| --- | --- | --- | --- | --- |
| Sh | TS 29.329 V18.0.0；Sh过程及Sh-Data TS 29.328 V18.1.0 | 16777217；UDR/UDA 306、PUR/PUA 307、SNR/SNA 308、PNR/PNA 309 | `0.1.0-slice.11` | [sh.json](../spec/sh.json) / [app/sh](../app/sh/zz_generated_messages.go) / [rules.go](../app/sh/rules.go) |
| S6c | TS 29.338 V18.2.0 | 16777312；SRR/SRA 8388647、RDR/RDA 8388649、ALR/ALA 8388648 | `0.1.0-slice.22` | [s6c.json](../spec/s6c.json) / [app/s6c](../app/s6c/zz_generated_messages.go) / [rules.go](../app/s6c/rules.go) |
| SGd/Gdd | TS 29.338 V18.2.0；两接口共用应用及命令定义 | 16777313；OFR/OFA 8388645、TFR/TFA 8388646 | `0.1.0-slice.15` | [sgd.json](../spec/sgd.json) / [app/sgd](../app/sgd/zz_generated_messages.go) / [rules.go](../app/sgd/rules.go) |
| 标准Ro SMS | TS 32.299 V18.0.0、RFC 4006；部分PS字段引用TS 29.061 V18.1.0等，见sources | 4；CCR/CCA 272 | `0.1.0-slice.6` | [ro.json](../spec/ro.json) / [app/ro](../app/ro/zz_generated_messages.go) / [rules.go](../app/ro/rules.go) |
| 通用DCCA（另一CCR绑定） | RFC 4006；有限单服务EVENT，不是完整RFC 8506实现 | 4；CCR/CCA 272 | `0.1.0-slice.1` | [dcca.json](../spec/dcca.json) / [app/ro/dcca](../app/ro/dcca/zz_generated_messages.go) / [rules.go](../app/ro/dcca/rules.go) |
| Ro独立会话中止 | RFC 6733 §§8.5.1–8.5.2及TS 32.299 §6.5.5引用；TS版次沿Ro审计基线，不另宣称整版支持 | 4；ASR/ASA 274 | `0.1.0-slice.1` | [roabort.json](../spec/roabort.json) / [app/ro/abort](../app/ro/abort/zz_generated_messages.go) / [rules.go](../app/ro/abort/rules.go) |
| Base | RFC 6733所选控制/公共错误规则 | 0；CER/CEA 257、DWR/DWA 280、DPR/DPA 282 | `0.1.0-slice.8` | [base.json](../spec/base.json) / [base](../base/zz_generated_messages.go) / [rules.go](../base/rules.go) |

Sh的XML是**User-Data AVP中的业务载荷**，与协议定义的JSON不是同一层。显式XML构造/解析在[app/sh/shdata](../app/sh/shdata/data.go)，独立范围`shdata-29.328-ref11-12-24-25-v1`，不是完整XSD校验器。Diameter codec不会自动解析该XML。Ro当前标准SMS支持IEC DIRECT_DEBITING及无UPDATE的ECUR INITIAL/TERMINATION等已选分支；不能从Application-ID=4推出支持所有DCCA/SMS动作。

## 兼容绑定不是另一个3GPP版本

| 显式绑定 | 当前版本 / 头身份 | 来源与入口 |
| --- | --- | --- |
| sms-fw IEC | `0.1.0-slice.6`；4/272 | 旧C/C++实际模板与writer来源；[smsfwro.json](../spec/smsfwro.json) → [compat/smsfwro](../compat/smsfwro/zz_generated_messages.go) |
| sms-fw legacy420 | `0.1.0-slice.6`；4/272 | 独立承接旧420类型，不修改标准Ro定义；[legacy420.json](../spec/legacy420.json) → [compat/smsfwro/legacy420](../compat/smsfwro/legacy420/zz_generated_messages.go) |
| sms-fw五CCR模板组合 | `0.1.0-slice.5`；4/272 | 具名部署绑定；[combined.json](../spec/combined.json) → [compat/smsfwro/combined](../compat/smsfwro/combined/zz_generated_messages.go) |
| SCAP1 | `0.1.0-slice.2`；头0/271 ACR/ACA | Vendor193/Acct19302能力及显式`SCAP1Accounting`，不是Ro App4；[scap1.json](../spec/scap1.json) → [compat/scap1](../compat/scap1/zz_generated_messages.go) |
| 旧SGd提醒 | `legacy-sgd-alert-v1.1`；16777313/8388648 ALR/ALA | 唯一无独立JSON的委托Profile：[compat/sgdalert/profile.go](../compat/sgdalert/profile.go)，委托S6c字段和上下文；不是标准SGd新增命令 |

标准Ro、通用DCCA及三个sms-fw CCR绑定占用相同4/272命令键，**同一Bundle只能选一个**，不能把所有Profile一起Compile。ASR/ASA是另一个命令键，可以与所选CCR Profile组合。`schema.Compile`拒绝冲突，不根据payload猜测私有模板。当前11份生成定义加1个委托Profile，按绑定计20对命令，含重叠命令键，不能称为20个唯一标准命令对。

## JSON如何定义应用与字段

实际schema结构在[internal/codegen/generate.go](../internal/codegen/generate.go)的`Definition`、`Group`、`Command`、`Field`；不是任意JSON字典。顶层`package/profile/version/sources`声明包、Profile身份、切片版本和审计依据；`groups`声明父组类型，`commands`声明头身份及字段。字段按**当前命令或父组**定义，不靠一个全局AVP名表消除父路径差异。

例如[s6c.json](../spec/s6c.json)的SRR包含以下实际片段（省略其余字段，不能单独作为完整定义输入）：

```json
{
  "name": "SRR",
  "application": 16777312,
  "code": 8388647,
  "request": true,
  "proxiable": true,
  "rule": "validateSRRSubset",
  "context_rule": "checkRoutingContext",
  "context_name": "S6cRoutingPurpose",
  "fields": [
    {"name":"MSISDN", "code":701, "vendor":10415, "flags":192, "type":"tbcd", "min":0, "max":1}
  ]
}
```

- `application/code/request`生成`CommandKey()`；`vendor/code`是AVP键。`flags=192`是0xC0，即V/M位；具体允许值按定义和规则检查。
- `type`选择有界原语（如`u32/u64/time/tbcd/octets/identity`）或命名Grouped类型。`min/max`定义出现次数；0..1生成`avp.Optional[T]`，1..1生成必需值，重复字段生成slice并受运行时Limits限制。例中MSISDN可缺省不是“请求可无身份”，其与其他身份的互斥/用途由手写规则检查。
- `rule`指向同Go包内手写函数；`answer_rule`检查原请求关系，`context_rule/context_name`检查调用者显式提供的有限证据；`no_context`是审核过的无额外证据声明。字段结构可生成，跨字段/结果/原请求关系仍须显式规则。
- `envelope/correlation`标记包络及关联；`error_answer/error_rule`选择独立E错误语法，`transport_replay`只声明已审核的传输恢复能力，不能赋予业务自动重试权。完整含义以生成器、运行时和行为spec为准。
- 未知可选AVP可按边界保留为`Extra []avp.Raw`，不代表语义已支持；已知但未选字段可能被手写规则明确拒绝。不能借清M或Raw绕开所选标准/私有类型绑定。

生成文件提供强类型报文、encode/decode、校验调用、owned clone、包络补全、Canonical、预算上界和`Profile()`操作表，不使用反射运行时字典。[schema/schema.go](../schema/schema.go)组合固定Profile；[codec](../codec/codec.go)使用Bundle查找命令。

## 如何装载和再生成

先调用各应用的`Profile()`，检查错误后`schema.Compile(baseProfile, s6cProfile, sgdProfile, shProfile, selectedRoProfile)`；离线codec使用`codec.New(bundle, codec.Config{...})`，Node使用`NodeConfig.Bundle`。`base.Profile()`提供Node所需握手和关闭控制。不存在`init()`全局自动注册。

加载Profile不会自动宣告能力或授权命令。应用还必须配置`LocalCapabilities`、出站`CommandPolicy`/入站`InboundPolicy`、固定终端或显式目标、可信来源及有限预算。Sh采用Vendor10415/Auth16777217的VSAI，且正常Sh报文本身也带VSAI；S6c/SGd的正常报文VSAI规则不能照抄Sh。DRA邻接身份与HSS/MME/OCS终端身份分别配置。完整可运行配置见[标准Ro IEC例](../examples/ro/config.go)、[SGd/Gdd例](../examples/sgd/config.go)、[S6c例](../examples/s6c/config.go)、[Sh例](../examples/sh/config.go)、[SCAP1例](../examples/scap1/config.go)，通用流程见[API工作流](diameterkit-api-workflows.md)。

在repo根目录执行：

```sh
# 只再生成S6c
CGO_ENABLED=0 go run ./cmd/diameter-schema -in spec/s6c.json -out app/s6c/zz_generated_messages.go
# 按doc.go的11条go:generate再生成全部
CGO_ENABLED=0 go generate ./...
# 检查确定性和生成物没有落后于输入
CGO_ENABLED=0 go test ./internal/codegen
```

维护入口是[doc.go](../doc.go)的`go:generate`、[CLI](../cmd/diameter-schema/main.go)及JSON源，不直接编辑`zz_generated_messages.go`。新增/修改字段时同时更新对应规范依据、切片版本、手写条件和成对正反报文/关联测试，再生成并回归。新增应用还须添加生成指令、定义清单及测试覆盖，不能仅新增一个文件就宣称Node支持。

运行时可查询应用包`Version`、`DefinitionDigest`以及`Profile.Name()/Version()/Digest()`，`Bundle.Fingerprint()`给出组合身份。生成Digest由generator版本与规范化JSON计算，Bundle包含Profile及ops契约；**它不自动hash手写Go函数体**，所以手写语义变化也必须更新JSON的`version/sources`修订说明并再生成。全源摘要另由[progress.json](../spec/progress.json)追踪，不把Profile摘要当完整软件供应链签名。

本批指南的版本/命令/路径核对见[机器清单](../spec/evidence/batch-80-application-definitions.json)。它是JSON/生成物的派生证据，不是第二套定义源。完整库状态仍以[行为规范](diameterkit-specification.md)与[spec进度](../spec/README.md)为准。

第85批公开ASR/ASA例见[Ro示例](../examples/ro/README.md)：`-flow abort`显式组合`ro.Profile()`与`abort.Profile()`，但只授权274；Profile/能力广告/业务方向各自配置。定义和生成版本未变，不能把独立ASR当成CCR的新类型或自动计费会话管理器。

第86批的[支持矩阵](diameterkit-support-matrix.md)按12个Profile分别记录实现/消费/限制，并附原Ro十二行用法。24个公开包的[API快照](../spec/release/public-api.json)可重复检查；它不修改本页规范或Profile版本。生产源码未变，仍复用batch85离线候选，没有人为生成batch86发布版本。
