# F3 Real Model Profile and Evals: LLM Judge, offline eval set, judge profile activation

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-24

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [prompt-rubric-registry spec](../../spec.md) §7 列出的后续派生计划 `004-real-model-profile-and-evals` 落到可实施计划。完成后：

- F3 `internal/ai/registry` 不再只导出 `NotImplementedJudge`，而是提供真实 `LLMJudge` 实现：经 A3/F3 cross-owner additive 的 judge capability dispatch（`judge.default` model profile + `CompleteJudge`/等价窄接口）调 `AIClient`，载入 rubric（`GetRubric`）、output schema 与模型 output，按 rubric `dimensions[]` **逐维度**产出 `[]Score` + `Reasoning`。
- `Judge` 接口按 spec **D-9 v2.8** 从返回单个 `Score` 演进为返回 `[]Score`（每个 rubric dimension 一项）；`TestJudgeSignature` 断言新签名。
- `config/ai-profiles.yaml` 的 `judge.default` 从 `status: unsupported` 翻为 `active`，`config/ai-providers.yaml` 用非 placeholder `judge_compatible` provider ref（例如 `judge-deepseek`，不得保留 `judge-placeholder` / `judge-provider-required`）承接真实 judge 调用；`make lint-ai-profile-coverage` 新增断言 `judge.default` active 且 §3.1.1 的 13 个 chat profile 全部解析到**非 placeholder** provider。
- 落地 ≥ 50 题离线评估集（`config/evals/<feature_key>/`，覆盖 §3.1.1 chat feature_key），并用 **Promptfoo** 承载 runner；Promptfoo 必须经 `RegistryClient.ResolveActive` 消费同一份 `config/prompts` 真理源，禁止复制第二份 prompt 文本；`make eval-offline` 默认跑录制 fixture（确定性、零成本、CI 安全），`EVAL_LIVE=1` opt-in 真实 provider 调用。
- 评估维度复用现有 rubric `dimensions[]`（如 `followup_relevance` / `report_specificity` / `score_outlier` / `language_consistency`；report 业务仍可复用自身 `report_calibration` 维度），对齐 spec §3.2 默认质量指标口径，不重新发明同义维度。

本计划**不**新增用户可见 UI、**不**改 live API 请求路径（LLM Judge 仅用于离线评估，不进入业务 HTTP 请求链路）、**不**重复翻动 A3 已 `active` 的 13 个 chat 业务 profile 的 `status`（尊重 A3 ownership，只做 `judge.default` 激活与 coverage 断言）、**不**引入 PostHog 灰度分桶（推到 `005-grayscale-and-quality-feedback`）。

## 2 背景

`prompt-rubric-registry` spec 在 `001-baseline`（2026-05-09 completed）锁定了 `Judge` 接口签名并交付 `NotImplementedJudge` stub（每次返回 `ErrJudgeNotImplemented`）；`002-output-schema-contract`（completed）与 `003-language-coordinate-simplification`（completed）收敛了 output schema 契约与 canonical `multi` 坐标。spec §7 把「切真实 Model Profile + ≥ 50 题离线评估集 + LLM Judge 实现」推给本 plan，依赖 A3 `003-provider-registry-and-capability-profiles`（2026-05-08 completed）提供完整 profile coverage 与 judge capability profile。

深度核对当前事实（避免照搬历史口径）：

- A3 `config/ai-profiles.yaml` 中 §3.1.1 的 13 个 chat 业务 profile **已经**是 `status: active` + `provider_ref: deepseek`。因此「切真实 Model Profile」在 catalog 层对 chat 业务已基本满足；本 plan 不再重复翻动它们，只新增 coverage 断言锁定它们解析到非 placeholder provider。
- 唯一仍是 `status: unsupported` 的是 `judge.default`（`provider_ref: judge-placeholder`、`model: judge-provider-required`、`unsupported_reason: "LLM judge adapter is reserved for F3 eval implementation"`），其 capability 与 provider protocol 名称（`aiclient.CapabilityJudge` / `ProviderProtocolJudgeCompatible = "judge_compatible"`）在 A3 aiclient 已存在；但当前 `AIClient.Complete` 只按 `CapabilityChat` dispatch，`bootstrap.ResolveProvider` 对 `judge_compatible` 仍直接返回 unsupported。因此本 plan 负责同时激活 profile、替换 provider ref、接线 judge dispatch/adapter，并证明 chat `Complete` 与 judge `CompleteJudge` 边界不会互相绕过。
- `backend/internal/ai/aiclient/profile/catalog_test.go` 当前断言 `judge.default` 为 `ProfileStatusUnsupported`；Phase 3 翻 active 时必须同步更新该断言。
- 当前仓库无 `config/evals/` 与任何 eval runner；`make test` 显式约定 “AI tests use stub/fixture only, no provider secrets”，因此 `EVAL_LIVE` live 模式必须排除在 `make test` 与默认 `make eval-offline` 之外。
- 当前根 `package.json` 仅锁 `pnpm`，`pnpm-workspace.yaml` 仅包含 `frontend`；Promptfoo runner 必须在实施时选择明确的 repo-owned 依赖落点（根 devDependency 或新增 workspace package），写入 lockfile，不允许用未固定版本的 `pnpm dlx`/临时全局安装作为完成证据。
- spec §3.2 把「Promptfoo / OpenAI Evals 选型、prompt versioning 语义命名、LLM Judge 使用哪个 profile、rubric 维度是否对齐质量指标」显式 defer 给本评估 plan；本会话已确认：Promptfoo（pnpm Node，仓库已有前端 Node 运行时）、`judge.default` profile、复用现有 rubric 维度。

spec v2.8 已在设计阶段固化本 plan 的接口与决策（D-9 演进、新增 D-15、§6 C-7/C-10 改写 + C-14/C-15、§5 边界、§7）；本 plan 只实施代码与配置以匹配 spec，不在 plan 内反向改写 spec 设计。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + contract + tooling + config`。涉及 Go（`internal/ai/registry` Judge 接口与实现）、Python（`scripts/lint/ai_profile_coverage.py` 门禁扩展）、Node/Promptfoo（eval runner）、`config/`（evals 用例 + `ai-profiles.yaml` judge.default 激活）。
- **TDD 策略**: Code plan requires TDD，经 `/implement` → `/tdd` 串行执行。每个实现项先红后绿：
  - Phase 1：先重写 `backend/internal/ai/registry/types_test.go` 的 `TestJudgeSignature`（期望 `[]Score` 返回）→ 红；改 `types.go` Judge 接口 → 绿。
  - Phase 2：先写 A3 judge dispatch/adapter 红灯（`Complete` 继续拒绝 judge profile，`CompleteJudge`/等价窄接口只接受 `CapabilityJudge`，`judge_compatible` bootstrap 不再 unsupported）+ `LLMJudge` 单测（录制 fixture judge 响应，断言逐维度 `[]Score` + `Reasoning` + fail-close）→ 红；实现 judge dispatch/adapter 与 `judge.go` `LLMJudge` → 绿。
  - Phase 3：先改 `catalog_test.go` 与 `scripts/lint/ai_profile_coverage.py` 断言（judge.default active + 非 placeholder provider_ref/model + 负向 placeholder 拒绝）→ 红；改 `config/ai-profiles.yaml`、`config/ai-providers.yaml` 与 lint 脚本 → 绿。
  - Phase 4：先写 eval count≥50 断言、registry-single-source drift 断言、`EVAL_LIVE` 未设不打网络断言 → 红；落地 `config/evals/` + Promptfoo 配置 + `make eval-offline` → 绿。
- **BDD 策略**: **BDD 不适用**。本 plan 不新增用户可见 UI、不新增/改变 HTTP API 请求路径行为、不引入端到端用户业务流；LLM Judge 与 eval 都是离线质量治理工具，judge.default 激活只影响离线评估而非 live 请求。用户可见的 AI 输出质量仍由各业务/前端 plan 的既有 BDD/i18n gate 承接。
- **替代验证 gate**（BDD 不适用的替代闭环）：
  - `go test ./backend/internal/ai/registry -count=1`（`TestJudgeSignature` 新签名 + `LLMJudge` 逐维度 + fail-close）。
  - `go test ./backend/internal/ai/aiclient/... -run 'Test.*Judge|Test.*judge' -count=1`（judge dispatch/adapter + bootstrap provider protocol + chat/judge capability 边界）。
  - `go test ./backend/internal/ai/aiclient/profile -count=1`（`catalog_test.go` judge.default active）。
  - `make lint-ai-profile-coverage`（judge.default active + judge provider/profile 非 placeholder + 13 chat profile 非 placeholder；负向 placeholder fail）。
  - `make eval-offline`（默认录制 fixture 跑通 ≥ 50 题；count≥50 断言；`EVAL_LIVE` 未设不打网络）。
  - registry-single-source drift check（Promptfoo 消费的 prompt == registry resolved；漂移 exit 1）+ `make lint-prompts-hardcode`（仍 green，证明未复制第二份 prompt）。
  - active-scope zero-reference grep（完整旧编号、短写旧编号与 stale spec version 引用在 active specs、F3 plans index、README、代码/配置/脚本 truth source 中 = 0；排除本 plan 自身 gate 定义目录、completed plan、history、`docs/reports/`、`docs/work-journal/`、`docs/bugs/` 历史记录）。
  - `make lint`、`validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check`。

## 4 实施步骤

### Phase 0: 当前契约预读与现状快照

#### 0.1 前后端契约预读

读取 `docs/development.md` §2、`config/README.md`、`config/prompts/README.md`、`config/rubrics/README.md`、`backend/README.md`、`ai-provider-and-model-routing/spec.md`、`observability-stack/spec.md`，确认本 plan 不涉及 OpenAPI operation matrix 或 UI parity，替代 gate 为 lint / registry unit test / eval-offline / drift。

Operation matrix: `N/A`。本 plan 不新增或改变 HTTP operation、OpenAPI fixture、frontend consumer、backend handler 或 persistence schema；所有新增行为在 Go 内部接口、provider/profile config、eval fixture 和 Make/script gate 内闭环。跨层 AI dependency 由 `judge.default` + `judge_compatible` provider ref + eval runner 记录，scenario coverage 由替代 gate `make eval-offline` 与 focused Go/lint gates 承接。

#### 0.2 Judge 与 profile 现状快照

固化实施起点快照：`judge.default` 当前 `status: unsupported` + placeholder provider；`registry.NotImplementedJudge` 返回 `ErrJudgeNotImplemented`；`types.go` Judge 当前返回单个 `Score`；`catalog_test.go` 断言 judge.default unsupported；13 个 chat profile 当前 active + deepseek；`ai_profile_coverage.py` 当前断言 entry 存在 + capability/provider_ref/status 合法。

额外快照：`AIClient.Complete` 当前以 `CapabilityChat` 为唯一 expected capability，不能直接调用 `judge.default`；`bootstrap.ResolveProvider` 当前对 `ProviderProtocolJudgeCompatible` 返回 `AI_UNSUPPORTED_CAPABILITY`；`config/ai-providers.yaml` 当前只有 `judge-placeholder`，不得作为 active judge provider ref 交付。

#### 0.3 eval 维度映射锁定

从现有 `config/rubrics/<feature_key>/v0.1.0.yaml` 的 `dimensions[]` 提取真实维度名，映射 spec §3.2 默认质量指标（追问相关率→`followup_relevance`、报告空泛率→`report_specificity`、异常高分率/离群评分→`score_outlier`、语言混乱率→`language_consistency`；report 业务校准仍复用自身 `report_calibration`），锁定 eval 只复用现有 rubric 维度、不新增同义维度。

#### 0.4 Promptfoo single-source 与执行模式约束确认

锁定不变量：Promptfoo 只能经 `RegistryClient.ResolveActive` 取 prompt（或由 registry 导出的解析产物驱动），不得在 eval 资产中复制 prompt 正文；默认执行模式跑录制 fixture（golden transcript），`EVAL_LIVE=1` 才发起真实 provider 调用；live 模式不进 `make test`、不进默认 `make eval-offline`。

### Phase 1: Judge 接口演进为逐维度 []Score

#### 1.1 TestJudgeSignature 红灯

重写 `backend/internal/ai/registry/types_test.go` 的 `TestJudgeSignature`：断言 `Judge.Judge` 第 1 返回值为 `[]Score`（而非单个 `Score`），保留 5 入参（ctx/featureKey/promptVersion/output/rubricVersion）与 `(., Reasoning, error)` 形状；先期红灯。

#### 1.2 types.go Judge 接口演进

按 spec D-9 v2.8 修改 `backend/internal/ai/registry/types.go`：`Judge` 接口返回 `([]Score, Reasoning, error)`；`Score{Dimension, Value}` 语义改为「每个 rubric dimension 一项」；`Reasoning` 保留 `Summary` + `EvidenceQuotes`。更新 doc.go 注释。

#### 1.3 NotImplementedJudge 与 caller 同步

更新 `judge.go` 的 `NotImplementedJudge.Judge` 返回 `(nil, Reasoning{}, ErrJudgeNotImplemented)`；更新 `judge_test.go`、`types_test.go` 中 `stubJudge` 与任何 import `Judge` 的 caller，保证全包编译通过。

#### 1.4 全包 build / test 绿灯

`go build ./backend/...` + `go test ./backend/internal/ai/registry -count=1` 通过；确认接口演进未破坏既有 resolver/loader/cache 测试。

### Phase 2: judge capability dispatch + 真实 LLMJudge 实现

#### 2.1 A3 judge dispatch / adapter 红灯单测

新增或扩展 A3 aiclient tests：`Complete(ctx, "judge.default", ...)` 必须继续因 capability mismatch fail-close；新增 `CompleteJudge(ctx, "judge.default", payload)`（或等价窄接口）只接受 `CapabilityJudge`；`CompleteJudge` 调 chat profile 必须 fail-close；`bootstrap` 对 `judge_compatible` provider ref 不再返回 “protocol not implemented”；provider/profile capability 不匹配仍 fail。先期红灯。

#### 2.2 judge_compatible adapter / AIClient 接线

实现 judge capability dispatch：A3 `AIClient` 保持现有 `Complete` 的 chat-only 语义，新增 `CompleteJudge`/等价窄接口走 `CapabilityJudge`；`bootstrap` 为 `ProviderProtocolJudgeCompatible` materialize repo-owned `providers/judge_compatible` adapter；adapter 使用明确的 JSON completion wire（可复用 OpenAI-compatible Chat Completions 子集，但 protocol/capability/meta 必须保持 `judge`），并延续 A3 secret fail-fast、fallback、observability 与 privacy red-line。`LLMJudge` 后续只依赖该窄接口，不直接绕过 A3 provider/profile/observability。

#### 2.3 LLMJudge 红灯单测

新增 `backend/internal/ai/registry/judge_llm_test.go`：用录制 fixture judge 响应，断言 `LLMJudge.Judge` 对给定 output + rubric 返回 `len([]Score) == len(rubric.dimensions)`、每项 `Dimension` 与 rubric 维度名一致、`Value ∈ [0,1]`、`Reasoning.Summary` 非空；先期红灯。

#### 2.4 LLMJudge 实现

实现 `LLMJudge`：构造函数注入 `RegistryClient`（取 rubric + output schema）+ A3 judge model client（经 `judge.default` profile 调用）；按 rubric `dimensions[]` 组织 judge prompt，解析 judge 模型输出为逐维度 `[]Score` + `Reasoning`；`Value` 按 rubric `score_levels[].threshold` 可映射回 label。judge 评分指令模板是 eval-harness 资产，从 `config/evals/`（F3 `004` 命名空间）按文件加载，不在代码内 hardcode 字面量（遵守 `lint-prompts-hardcode`），也不占用 §3.1.1 业务 feature_key 坐标、不复用 `report.question_assessment` 等业务 judge prompt；被评估的 13 个 business prompt 仍经 `RegistryClient.ResolveActive` single-source 取得。

#### 2.5 fail-close 路径

断言并实现 fail-close：`judge.default` profile 不可用 / 非 active → 返回 error；judge 模型输出无法解析或维度缺失 → 返回 error（不静默补零）；被评估 output 不满足该 feature_key/promptVersion 的 output schema → fail-close。output schema 校验必须复用或提取 A3 当前 `validateOutputSchema` 的同一子集语义，禁止在 F3 另写不等价校验器。对应负向单测。

#### 2.6 Judge 接线

提供 `var _ Judge = (*LLMJudge)(nil)` 编译期断言；保留 `NotImplementedJudge` 作为未注入 judge 依赖时的安全默认；记录 eval runner 如何获得 `LLMJudge` 实例（构造注入，非全局单例）。

### Phase 3: judge.default 激活 + coverage 门禁扩展

#### 3.1 catalog / coverage 红灯

先改 `backend/internal/ai/aiclient/profile/catalog_test.go`：`judge.default` 期望 `ProfileStatusActive`；先改 `scripts/lint/ai_profile_coverage.py` 测试期望：断言 judge.default active + 非 placeholder provider_ref/model，且 §3.1.1 的 13 个 chat profile 全部解析到非 placeholder（拒绝 `judge-placeholder` / `unit-test-stub` / `*-provider-required` / `*-placeholder`）。先期红灯。

#### 3.2 judge.default 激活

修改 `config/ai-providers.yaml`：新增或替换为非 placeholder judge provider ref（建议 `judge-deepseek`，`protocol: judge_compatible`，`capabilities: [judge]`，secret env refs 走 A4 字典中现有 provider env 或经 A4 spec 修订后新增）；删除或停用 active scope 对 `judge-placeholder` 的依赖。修改 `config/ai-profiles.yaml`：`judge.default` `status: unsupported → active`，`provider_ref` 指向该非 placeholder provider，移除 `unsupported_reason`，`model` 填真实可用 judge model；保持 `capability: judge` 与 `route: judge.default` 不变。**不**修改任何已 active 的 13 个 chat 业务 profile 的 status。

#### 3.3 coverage lint 扩展实现

扩展 `scripts/lint/ai_profile_coverage.py`：新增 placeholder 黑名单断言（judge.default 与 13 个 chat profile 的 default provider_ref/model 必须非 placeholder），并断言 `judge.default` 选中的 provider protocol/capability 与 `CapabilityJudge` 匹配；保留既有 entry/capability/status/unsupported_reason 校验；接线进顶层 `make lint`。

#### 3.4 coverage 绿灯 + 负向断言

`make lint-ai-profile-coverage` 返回 OK；负向 fixture/单测证明把任一 chat 或 judge profile/provider 改回 placeholder 时 gate exit 非 0；`go test ./backend/internal/ai/aiclient/profile -count=1` 与 focused judge dispatch/adapter tests 通过。

### Phase 4: ≥50 题离线评估集 + Promptfoo runner

#### 4.1 评估用例集

落地 `config/evals/<feature_key>/`（覆盖 §3.1.1 chat feature_key），评估题总量 ≥ 50；用例引用 feature_key + 输入 + 期望维度信号；至少 1 个用例覆盖 `en → multi` fallback resolve（请求语言无 exact coordinate 时仍可评估）。

#### 4.2 Promptfoo 依赖与 runner 落点

新增 repo-owned Promptfoo 依赖落点：根 `package.json` devDependency 或新增 workspace package（二选一，实施时按 repo 约定选择并同步 `pnpm-workspace.yaml`/`pnpm-lock.yaml`），命令必须通过 pinned dependency 执行（如 `pnpm exec promptfoo` 或 workspace script）。禁止以未固定版本的 `pnpm dlx promptfoo`、全局安装或外部临时目录作为验收证据。

#### 4.3 Promptfoo registry-driven 配置

落地 Promptfoo 配置：custom provider 经 `RegistryClient.ResolveActive` 取 prompt 三元组与 output_schema（或消费由 registry 导出的解析产物）+ `AIClient` 调 deepseek-backed profile；`LLMJudge` 作为 Promptfoo grader 产出逐维度评分。Promptfoo 资产不复制业务 prompt 正文；`LLMJudge` grader 的评分指令来自 `config/evals/` eval 资产（非业务 prompt 复制）。

#### 4.4 录制 fixture 默认 + EVAL_LIVE opt-in

落地 golden transcript 录制 fixture 作为默认评估输入；`EVAL_LIVE=1` 时才发起真实 provider/judge 调用；实现并断言 `EVAL_LIVE` 未设时 runner 不发起网络请求（确定性、零成本）。

#### 4.5 make eval-offline 与 single-source drift gate

新增 `make eval-offline` target（`.PHONY` + help 描述；默认 fixture 模式；**不**纳入 `make test`）；落地 count≥50 断言与 registry-single-source drift check（Promptfoo 实际使用的 prompt 必须等于 registry resolved，漂移 exit 1）。

#### 4.6 hardcode-green 负向闭环

确认 `make lint-prompts-hardcode` 仍 green（eval 引入未复制第二份 prompt 正文）；drift gate 与 hardcode gate 共同锁定 single-source 不变量。

### Phase 5: 验证、生命周期与收口

#### 5.1 focused + 聚合验证

运行 §3 替代验证 gate 全集：registry / profile go tests、`make lint-ai-profile-coverage`、`make eval-offline`、`make lint`、`make test`、`make build`。

#### 5.2 active-scope zero-reference grep 门禁

`rg -n '002-real-model|003-grayscale|F3 后续 002|prompt-rubric-registry/spec\.md.*v2\.[0-8]' docs/spec/*/spec.md docs/spec/prompt-rubric-registry/plans/INDEX.md config backend scripts --glob '!docs/spec/prompt-rubric-registry/plans/004-real-model-profile-and-evals/**'` 必须为 0；历史审计资产允许保留旧引用，completed plan/history 不作为 active-scope gate 输入。该 gate 证明 active specs、F3 plans index、README 与代码/配置/脚本 truth source 不再传播旧编号 shorthand 或旧 spec 版本引用。

#### 5.3 profile status 范围负向断言

`git diff config/ai-profiles.yaml config/ai-providers.yaml` 证明本 plan 只改 `judge.default` 一个 profile 的 status，并只新增/替换 judge provider ref；13 个 chat 业务 profile 的 `status` 与 `provider_ref` 未被翻动（尊重 A3 ownership）。

#### 5.4 docs / lifecycle 收口

运行 `validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check`；完成后将 004 plan/checklist 置为 `completed`，同步 plans INDEX、docs/spec INDEX、work journal、retrospective；如发现真实缺陷按 `/bug-report` 建档。

## 5 验收标准

- `TestJudgeSignature` 断言 `Judge` 返回 `[]Score`（逐维度），与 spec D-9 v2.8 一致；`go build ./backend/...` 通过。
- `LLMJudge` 经 `judge.default` 产出 `len == len(rubric.dimensions)` 的 `[]Score` + `Reasoning`；judge profile 不可用 / 输出不可解析 / 被评估 output schema invalid 均 fail-close（负向单测覆盖）。
- A3/F3 judge dispatch 可执行：chat `Complete` 不接受 judge profile，judge `CompleteJudge`/等价窄接口只接受 `CapabilityJudge`，`judge_compatible` adapter 不再 unsupported，secret/observability/privacy 规则延续 A3。
- `config/ai-profiles.yaml` `judge.default` 为 `active` + 非 placeholder provider/model；`config/ai-providers.yaml` 提供非 placeholder `judge_compatible` provider ref；`make lint-ai-profile-coverage` 断言 judge.default active 且 13 个 chat profile 解析到非 placeholder；负向 placeholder fail。
- `config/evals/` ≥ 50 题；`make eval-offline` 默认录制 fixture 跑通；count≥50 断言通过；`EVAL_LIVE` 未设不打网络；`EVAL_LIVE=1` opt-in live 可运行。
- Promptfoo 经 registry 解析消费同一份 prompt（single-source）；registry-single-source drift gate 与 `make lint-prompts-hardcode` 均通过；无第二份 prompt 副本。
- active-scope zero-reference grep：完整旧编号、短写旧编号与 stale spec version 引用在 active specs、F3 plans index、README 与代码/配置/脚本 truth source 中 = 0（排除本 plan 自身 gate 定义目录与历史 completed plan/history/reports/work-journal/bugs）。
- 本 plan 未翻动 A3 已 active 业务 profile 的 status（`git diff` 范围断言）。
- BDD 不适用声明与替代验证 gate 已固化；plan/checklist/context/index 生命周期闭环。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| `Judge` 接口演进破坏既有 caller / 测试 | Phase 1 先红 `TestJudgeSignature`，再改接口，全包 `go build` + registry tests 验证；`NotImplementedJudge` 同步演进保证安全默认。 |
| `judge.default` active 后仍因 AIClient chat-only dispatch / unsupported protocol 不可执行 | Phase 2 先写 A3 judge dispatch/adapter 红灯；`Complete` 保持 chat-only，新增 judge 窄接口走 `CapabilityJudge`；bootstrap materialize `judge_compatible` adapter；focused tests 证明两个能力边界。 |
| judge.default 激活后真实 judge 不可用导致 eval 失败 | 默认 eval 跑录制 fixture，不依赖 live judge；`EVAL_LIVE` opt-in 才打真实 provider；`LLMJudge` 对 profile 不可用 fail-close 而非静默。 |
| Promptfoo 复制 prompt 造成第二份真理源 | custom provider 强制经 registry 解析；registry-single-source drift gate + `lint-prompts-hardcode` 双重锁定；漂移即 exit 1。 |
| Promptfoo 工具版本漂移 / 本地全局安装造成不可复现 | Phase 4.2 固定 repo-owned dependency + lockfile；`make eval-offline` 只调用仓库内 pinned runner。 |
| live eval 误入 CI / `make test` 产生 API 成本与不确定性 | `EVAL_LIVE` 默认未设；`make eval-offline` 默认 fixture 模式且不纳入 `make test`；runner 在无 `EVAL_LIVE` 时断言不发网络。 |
| 误翻动 A3 已 active 业务 profile 越界 | Phase 3 只改 judge.default；Phase 5.3 用 `git diff` 范围断言锁定 chat profile status 未变；coverage gate 只做断言不改写业务 profile。 |
| eval 维度自造同义指标偏离 rubric | Phase 0.3 锁定只复用现有 rubric `dimensions[]`；`LLMJudge` 维度名来自 `GetRubric`，不在 eval 配置内新造维度。 |
| eval/judge live 调用污染 production `ai_task_runs` 业务指标 | eval/judge 经 AIClient 的 live 调用必须标记 eval 来源或排除出 production 业务 observability；默认 fixture 模式不产生 live 调用；Phase 4 实施时确认 `ai_task_runs` 埋点边界，不把 eval 流量并入业务 feature_key 统计。 |
| 旧 plan 编号 shorthand / stale spec version 残留导致审计漂移 | Phase 5.2 固化 active-scope zero-reference grep（排除本 plan gate 定义目录与历史 completed plan/history/reports/work-journal/bugs）；`/design` 修 F3-owned 残留，`/plan-review --fix` 修跨 subspec 残留。 |

## 7 关联文档导航

> - [Prompt Rubric Registry Spec](../../spec.md)
> - [001-baseline plan](../001-baseline/plan.md)
> - [002-output-schema-contract plan](../002-output-schema-contract/plan.md)
> - [003-language-coordinate-simplification plan](../003-language-coordinate-simplification/plan.md)
> - [AI Provider and Model Routing Spec](../../../ai-provider-and-model-routing/spec.md)
> - [Observability Stack Spec](../../../observability-stack/spec.md)
> - [ADR-Q6 AI Provider and Model Routing](../../../engineering-roadmap/decisions/ADR-Q6-ai-provider-and-model-routing.md)
> - [Development Contract Workflow](../../../../development.md)

## 8 Handoff / Evidence Log

### 8.1 Phase evidence log

> 本 plan 为新派生 active plan，evidence 待 `/implement` → `/tdd` 执行时逐 phase 回填，禁止预填 PASS。

| Phase | Evidence slot | Required command / artifact | 状态 |
|---|---|---|---|
| Phase 0 | 契约预读与现状快照 | development §2 + module README preflight；judge/profile 现状快照；eval 维度映射；single-source/执行模式不变量 | PASS：契约 §2 Operation matrix N/A 确认；现状快照（judge.default unsupported+judge-placeholder、types.go 单 Score、AIClient Complete chat-only、bootstrap judge_compatible→unsupported、catalog_test judge.default unsupported、13 chat active+deepseek、ai_profile_coverage OK）；eval 维度映射（followup_relevance/report_specificity/score_outlier/language_consistency + report_calibration，均来自现有 rubric dimensions[]）；single-source/fixture-default/EVAL_LIVE-opt-in 不变量锁定。附带修复 `backend_review_preflight_test.go` stale spec 版本断言 2.7→2.9（main 预存红，gate 5.2 范围内）。 |
| Phase 1 | Judge []Score 演进 | `TestJudgeSignature` 新签名红→绿；`go build ./backend/...`；registry tests | 待回填 |
| Phase 2 | judge dispatch + 真实 LLMJudge | A3 judge dispatch/adapter red→green；`judge_llm_test.go` 逐维度 + fail-close 红→绿；`go test ./backend/internal/ai/{aiclient,registry}` focused | 待回填 |
| Phase 3 | judge.default 激活 + coverage | `config/ai-providers.yaml` 非 placeholder judge provider；`catalog_test.go` active；`ai_profile_coverage.py` 非 placeholder 断言 + 负向；`make lint-ai-profile-coverage` | 待回填 |
| Phase 4 | eval 集 + Promptfoo | pinned Promptfoo dependency + lockfile；`config/evals` count≥50；`make eval-offline` fixture；drift gate；`EVAL_LIVE` 无网络断言；`lint-prompts-hardcode` green | 待回填 |
| Phase 5 | 验证与生命周期 | lint/test/build/eval/docs gates；active-scope zero-reference grep；profile status 范围断言；INDEX/work-journal/retrospective | 待回填 |

### 8.2 Owner handoff

- **A3 ai-provider-and-model-routing**: `judge.default` 由本 plan cross-owner additive 从 `unsupported` 翻 `active`（真实 provider）；`backend/internal/ai/aiclient/profile/catalog_test.go` 同步；A3 已 active 的 chat 业务 profile 不被本 plan 翻动。
- **F1 observability-stack**: rubric 维度复用现有命名，对齐 F1/F3 质量指标口径；本 plan 不扩展 F1 metric label。
- **F3 005-grayscale-and-quality-feedback**: PostHog 灰度分桶 + 报告页质量主观评分回流在本 baseline 评估闭环验证后承接。

### 8.3 Retrospective candidates

- Promptfoo registry-driven custom provider 与 Go registry 解析产物的接线方式（导出产物 vs in-process 调用）。
- 录制 fixture 的更新节奏与 `EVAL_LIVE` live 回归的触发条件。
- judge.default 激活后是否需要独立 judge provider rate-limit / 成本观测。
