# Provider Registry and Capability Profiles

> **版本**: 1.16
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [ai-provider-and-model-routing spec](../../spec.md) v1.9 的 provider registry 与 capability-scoped Model Profile 设计落到可实施计划中。完成后，业务仍只通过 F3 Resolve 得到 `model_profile_name` 并调用 `AIClient`，但 A3 runtime 能够按 profile 的 `capability`、`provider_ref`、model、参数与 fallback chain 选择真实 provider，不再把整个产品固定在一个全局 `AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 语义上。

本 plan 覆盖：

- `config/ai-providers.yaml` provider registry schema、loader、secret env ref 解析与 capability 校验；
- `config/ai-profiles.yaml` 单一 profile catalog 从 `task_type` / 全局 provider 口径迁移到 `capability` / `provider_ref` 口径，并收敛为单一 catalog；
- AIClient 中央路由与 profile fallback chain，业务代码不得自行 retry-with-different-model；
- A4 env/config 字典、B1 shared vocabulary、F3 6 个 baseline feature_key 覆盖与 drift gate；
- unsupported capability 的 fail-closed 行为，为后续 002 / C14 / F3 eval 打开 STT、realtime、judge adapter 留出安全边界。

本 plan 不实现完整 STT / realtime speech / judge provider 协议；这些 adapter 仍由 [002-tools-streaming-and-stt](../002-tools-streaming-and-stt/plan.md) 或对应业务 / eval plan 激活后承接。

## 2 背景

2026-05-08 基于当前个人开发阶段与用户决策重新收敛：easyinterview 当前只保留结构化抽取、对话生成、低延迟观察、长上下文报告、写作改写，以及 fail-closed 的 STT / realtime voice / judge-eval profile。当前阶段删除向量化 / 重排实现与基础设施，AIClient 只保留当前可执行 chat 能力和 fail-closed 的 speech / judge profile；repo-tracked 开发主力 provider 收敛为 DeepSeek V4 Flash/Pro。

本 plan 是 A3 001 之后的配置契约升级。它不重做已完成 bootstrap，而是在现有 `AIClient` / provider adapter / observability decorator 之上补齐 provider registry 与 capability profile。由于当前项目尚未上线，runtime 只接受 current schema / env keys；实施时直接迁移 repo-tracked fixtures、A4 bindings、B1 generated vocabulary 与相关 tests。

2026-05-05 本 plan 原地修订：用户反馈当前 per-profile YAML 目录对 17 个小 profile 来说文件碎片过多，维护与审查成本高于收益。经 change-intake 匹配，本主题仍由 003 承接，active truth source 改为单一 `config/ai-profiles.yaml` catalog；`AI_MODEL_PROFILE_PATH` env key 保留，但含义改为 catalog 文件路径。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + contract + platform-foundation`。本 plan 修改 AI provider runtime contract、profile/registry schema、配置绑定、共享字段与 lint gate；不直接引入用户可感知 UI、HTTP API 行为或业务工作流。
- **TDD 策略**: Code plan requires TDD。后续实施必须通过 `/implement` -> `/tdd` 执行。纯 profile 配置默认值、override 与非法值只在 typed loader owner package 保留一组表驱动契约，并由 canonical catalog/coverage lint 锁定当前坐标；消费者只为 fallback、协议读取上限、隐私/观测等非平凡分支保留 focused test，不逐 adapter、composition、domain 或 scenario 复制相同默认值与 `limit / limit+1` 断言。
- **BDD 策略**: BDD 不适用。本 plan 是内部 AI provider 配置与路由契约，不创建用户可见 UI、外部 API 行为或端到端业务流程。后续电话模式、report、practice 等用户行为 workstream 必须在自身 plan 维护 BDD gate。
- **替代验证 gate**: profile loader owner contract、active-budget floor lint、共享 response-cap 非平凡协议分支 focused test、backend-review 独立业务回归、config/env lint、B1 codegen drift check、privacy grep、docs/context/diff gates。纯配置不运行真实 provider、不做 bytes+tokens 公式，也不新增或复用 E2E gate；阶段收口由根 `make test` 执行前后端全量回归。

## 4 实施步骤

### Phase 1: Provider Registry schema 与 loader

#### 1.1 Registry schema truth source

定义 `config/ai-providers.yaml` schema：`providers[]` 每项包含 `name`、`protocol`、`base_url_env`、`api_key_env`、`capabilities[]`、`version`。`protocol` 当前允许 `stub`、`openai_compatible`、`realtime_audio`、`judge_compatible`；`capabilities[]` 当前允许 `chat`、`stt`、`realtime`、`judge`。tracked registry 只保存 env key 名，不保存 secret 明文；`stub` provider 允许 `base_url_env` / `api_key_env` 为空，其他需要网络出站的 protocol 必须声明 env ref。

#### 1.2 Registry loader + secret resolution

在 `backend/internal/ai/aiclient/` 增加 provider registry loader，启动时读取 `AI_PROVIDER_REGISTRY_PATH`，使用 A4 SecretSource 解析 `base_url_env` / `api_key_env`。loader 必须校验 provider name 唯一、protocol 合法、capabilities 非空；secret 校验按 protocol 和实际选中 provider 生效：`stub` 不要求 secret，网络出站 provider 必须声明 env ref，且非 test 环境中被 profile primary 或 fallback chain 选中时缺实际 secret 必须 fail-fast。

#### 1.3 Registry hot reload snapshot

registry 与 profile loader 使用同一 snapshot 语义：变更后 ≤ 30s 热加载，新调用使用新 registry/profile 快照，进行中的调用使用原快照完成。热加载失败不得污染当前有效快照，必须输出结构化 warn。

#### 1.4 Registry negative fixtures

新增负向 fixtures / tests 覆盖：重复 provider name、未知 protocol、capability 拼写错误、网络出站 provider 缺 env ref、profile 引用不存在 provider ref、profile capability 与 provider capabilities 不匹配、被选中真实 provider secret 缺失、fallback 超 2 跳；同时补正向 fixture 证明 `stub` provider 不需要伪造 secret。

#### 1.5 L2 remediation: runtime registry/profile wiring

补齐生产可用的 registry/profile bootstrap：非 test AIClient 构造必须实际读取 `AI_PROVIDER_REGISTRY_PATH` 与 `AI_MODEL_PROFILE_PATH`，通过 A4 SecretSource 校验 active profile 选中的 provider primary/fallback secret，按 provider ref materialize adapter，并确保 `ResolveSelectedProviders` 不只存在于单测路径。

#### 1.6 L2 remediation: reload warning evidence

profile hot reload 失败时必须保持当前快照并通过 `OnWarn` 输出结构化 warn；focused test 需覆盖失败 reload 不污染快照且 warning 可观测。

### Phase 2: Capability-scoped Model Profile schema

#### 2.1 Profile schema migration

将 Model Profile schema 从 `task_type` 迁移到 `capability`，将 `default.provider` 语义明确为 `default.provider_ref`，并将 `fallback[]` 扩展为 provider-aware chain。Profile 字段集必须对齐 spec §2.1，新增 `status=active|disabled|unsupported` 与 `unsupported_reason` 校验；loader 只接受 current schema keys，并由负向测试拒绝 out-of-scope keys。

#### 2.2 F3 baseline profile fixtures

在 `config/ai-profiles.yaml` 的 `profiles[]` catalog 中覆盖 F3 `prompt-rubric-registry` §3.1.1 的 6 个当前 feature_key 对应默认 profile 引用；其中 `resume.tailor.gap_review` 与 `resume.tailor.bullet_suggestions` 共享 `resume.tailor.default`。当前 chat profile 集合覆盖 `target.import.default`、`practice.chat.default`、`report.generate.default`、`resume.parse.default` 与 `resume.tailor.default`，并保留必要的 `status=disabled` / `status=unsupported` profile 表达 P1/P2 能力；不可执行 profile 必须写明 `unsupported_reason`。

同时补齐 spec §4.5 的非 F3 baseline Product/UI profiles：`target.intel.default`、`practice.voice.stt.default`、`practice.voice.tts.default`、`practice.voice.realtime.default`、`judge.default`。这些 profile 在对应 adapter / eval plan 激活前必须以 `disabled` / `unsupported` 状态存在并写明 `unsupported_reason`，不能缺 catalog entry，不能静默降级到 chat / stub；`judge.default` 已由 Phase 9 按独立 judge capability/protocol 激活。

#### 2.3 Product/UI capability coverage

为 spec §4.5 的 Product/UI AI Capability Catalog 建立覆盖检查：表内每个默认 profile 都必须是具体 profile name，并存在于 profile catalog；暂不可执行的 capability 必须登记为 `disabled` / `unsupported` fail-closed profile；新增 AI 场景必须能映射到 F3 feature_key 或 fail-closed profile；不得只在业务代码中 hardcode 新 profile 名。

#### 2.4 Schema docs and README

同步 `backend/internal/ai/aiclient/README.md`、`config/README.md` 与 profile fixture 注释，说明 provider registry、profile capability、fallback chain、unsupported capability 与 secret redaction 规则。

#### 2.5 Catalog consolidation remediation

新增单一 `config/ai-profiles.yaml` catalog schema：顶层只允许 `profiles[]`，每个 entry 复用 2.1 profile 字段集；loader、bootstrap、tracked catalog tests、F3/Product UI profile coverage lint、A4 默认配置、`.env.example` 与 README 全部改为读取 catalog 文件路径。删除 per-profile YAML directory active truth source，负向搜索确认 active scope 不再依赖该目录。

### Phase 3: AIClient routing, fallback, and fail-closed behavior

#### 3.1 Provider ref routing

AIClient 在每次调用前解析 profile snapshot，再根据 `capability` 与 `provider_ref` 选择 provider adapter。Chat 走 `openai_compatible` adapter；`status=disabled` / `status=unsupported` 或 adapter 未激活的 capability 返回 B1-owned `AI_UNSUPPORTED_CAPABILITY`（或同义且已由 B1 批准的 `AI_*` 错误码）并写 meta/log，不得降级到 chat 或 stub。

#### 3.2 Central fallback chain

实现 profile fallback chain：primary 失败且 `when[]` 条件命中时，AIClient 最多执行 2 跳 fallback；每一跳必须记录 provider/model/capability、错误码、latency、tokens（如可得）与最终结果。业务代码不得感知 fallback 细节，也不得自行切换 model。

#### 3.3 Observability and privacy regression

更新 `AICallMeta` / metric label / log 字段中的 capability 表达；fallback metric 使用 `from_provider` / `from_model_family` / `to_provider` / `to_model_family` 或 F1 最终确认的字段集。隐私测试必须证明 registry/profile 内容不会导致 prompt / response 明文落 log、metric label、DB metadata 或 audit metadata。

#### 3.4 Existing adapter compatibility

重构 openai_compatible adapter 的 base URL / API key 注入来源，从全局 config 切到 provider ref secret。保留 root endpoint 与已含 `/v1` endpoint 的归一化测试；不引入厂商 SDK。

### Phase 4: A4 / B1 / F3 integration

#### 4.1 A4 env and config dictionary

把 A4 env/config 真理源扩展为 `AI_PROVIDER_REGISTRY_PATH` + `AI_MODEL_PROFILE_PATH` + provider-specific secret env refs。`AI_PROVIDER_BASE_URL` / `AI_PROVIDER_API_KEY` 可作为默认 provider ref 引用的 env 名，但不得再被描述成全局唯一 provider contract。同步 `.env.example`、`config/config.yaml`、bindings、validator、redaction、`make lint-config` tests。

#### 4.2 B1 shared vocabulary

在 B1 shared vocabulary 中新增或迁移 AI capability enum、provider registry field names、profile field names、meta field names 与 provider/profile routing 错误码。至少需要确认或新增 `AI_UNSUPPORTED_CAPABILITY`、`AI_PROVIDER_CONFIG_INVALID`、`AI_PROVIDER_SECRET_MISSING` 三类 B1-owned `AI_*` code，并更新 Go/TS generated artifacts 与 parity tests；禁止 A3 私自导出跨语言常量或私造跨边界 error string。

#### 4.3 F3 profile coverage lint

F3 Resolve 字典中的默认 `model_profile_name` 与 spec §4.5 Product/UI AI Capability Catalog 中的默认 profile 必须全部存在于 A3 profile catalog。新增 lint gate：缺 profile、profile capability 不合法、provider ref 不存在、P1/P2/002+ profile 未标 `status=disabled` / `status=unsupported` 或缺 `unsupported_reason` 均失败。

#### 4.4 ADR and active docs sync

同步 ADR-Q6、A3 history、A4/F3 spec、engineering-roadmap A3 职责描述与 docs/spec INDEX。完成后 active docs 不得再把“单一 provider base URL/API key”描述成完整目标架构。

#### 4.5 L2 remediation: active profile anti-stub gate

补强 `make lint-ai-profile-coverage`：repo-tracked `status=active` 默认 profile 不得指向 `stub` protocol provider，除单元测试 / 离线 mock fixture 外不得把 P0 active profile 绑定到 `unit-test-stub`。同时将当前 active default profiles 切到真实 provider ref 可解析的 OpenAI-compatible provider 配置。

### Phase 5: Verification and handoff

#### 5.1 Focused tests

运行 registry loader tests、profile schema tests、AIClient routing/fallback tests、openai_compatible adapter tests、observability/privacy tests、A4 config tests、B1 vocabulary parity tests、F3 profile coverage lint。

#### 5.2 Global gates

运行 `make lint-config`、`make codegen-check`、`make docs-check`、`make lint`、`make test`、`make build`。如果后续实施改动只影响文档，仍至少运行 `make docs-check` 与 context validation。

#### 5.3 Negative search

执行 active-scope 负向搜索，确认代码、配置、deploy、active spec/plan、generated artifacts 中不存在 out-of-scope schema key，也不存在把 AI provider 描述为独立 provider-proxy 业务语义或单一全局 endpoint 的当前口径。`docs/work-journal/`、`docs/reports/`、`docs/bugs/` 记录不参与 active runtime / contract surface 判定。

#### 5.4 Header closeout

全部 gate 通过后，将本 plan / checklist Header 切到 `completed`，同步 `plans/INDEX.md` 与 `docs/spec/INDEX.md`，记录工作日志。向 002 / C14 / practice / report / resume / F3 eval owner 留出 handoff：可直接引用 capability profile，不需要新增业务侧 provider 配置。

#### 5.5 L2 remediation verification

复跑 L2 remediation focused tests、profile coverage lint、config lint、context validation、negative search 与必要全局 gate；完成后确认 plan/checklist Header 与 active spec 中的 003 状态投影均为 `completed`。

#### 5.7 L2 remediation: deploy/profile drift gate hardening

修复 catalog consolidation L2 review 发现的 active drift：`deploy/dev-stack/.env.example` 必须使用 `AI_PROVIDER_REGISTRY_PATH=config/ai-providers.yaml` 与 `AI_MODEL_PROFILE_PATH=config/ai-profiles.yaml`；active product / repo owner docs 不得再把 `config/ai-profiles/` 描述为 Model Profile active truth source；`make lint-ai-profile-coverage` 或同级 gate 必须覆盖 dev-stack profile catalog path 与 Product/UI capability catalog 对 `config/ai-profiles.yaml` 的 capability 语义一致性。

### Phase 6: DeepSeek baseline and retrieval cleanup

#### 6.1 Capability surface cleanup

删除当前 active scope 中的向量化与重排 capability、profile、provider protocol、AIClient 方法、adapter wire、stub、job type、migration 表/索引/dev-stack 依赖；未来如产品确有个人知识库或大规模资料检索需求，必须另开 spec/plan 重新设计。

#### 6.2 DeepSeek provider baseline

将 repo-tracked 开发期 provider ref 收敛为 `deepseek`，chat profile 只使用 `deepseek-v4-flash` 与 `deepseek-v4-pro`。低延迟交互、解析、轻量观察默认 Flash；报告、评估、简历改写默认 Pro。out-of-scope 模型别名必须被配置 lint 拒绝。Phase 6 当时保留的 judge fail-closed 命名空间仅由本 plan 后续 Phase 9 激活。

#### 6.3 Cross-contract drift repair

同步 A3 / B1 / B3 / B4 / F3 active spec、README、lint、generated artifacts、fixtures 与 config，使当前代码、配置、迁移和文档都只暴露现阶段真实可用能力。STT / realtime 在 Phase 6 继续保留 fail-closed 命名空间；judge 只允许由 Phase 9 以独立 capability/protocol 激活；DeepSeek 不承担 STT。

#### 6.4 Verification

运行 focused Go tests、B1/B3 codegen、profile coverage lint、config lint、migration lint、docs check 与 active-scope negative search。负向搜索应证明当前代码/配置/基础设施中没有向量化 / 重排实现、profile、provider ref、job type、migration 表或 dev-stack 依赖。

### Phase 7: Provider startup error contract cleanup

#### 7.1 Remove the test-only shared error mapper

删除没有运行时消费者、仅由自测维持的 `providerregistry.SharedErrorCode` 及其断言。Provider registry/bootstrap 启动失败继续使用 `ErrProviderConfigInvalid` 与 `ErrProviderSecretMissing` 哨兵错误，通过 `errors.Is` 保持可判定性；运行时业务边界继续由各 owner 映射 B1 `AI_*` 错误，不在启动配置层保留重复映射 API。

### Phase 8-9: Superseded budget migration cleanup

早期 report/judge budget 调整及其 exact-coordinate、边界 fixture、bytes+tokens 公式、activation marker 与真实 provider smoke 已被当前 typed-default 模型取代。当前只保留：

- loader owner 对 default / override / invalid 的单一契约；
- 六个 active profile `max_tokens >= 16384` 的 floor lint；
- judge adapter 的 non-thinking JSON request 与 reasoning-only/empty final fail-close；
- backend-review 自身的业务错误与 provider call/no-call 回归。

旧数值迁移不再作为当前 gate，也不得在 composition、domain、scenario 或 live provider 层复制。

### Phase 10: report generation non-thinking structured output

#### 10.1 RED: reproduce missing report thinking wire

Add openai-compatible request contract tests that require profile `thinking=enabled|disabled` to become the official `{"thinking":{"type":"..."}}` object while an object `output_schema` independently produces `response_format=json_object`. Loader owner tests reject invalid thinking values；an invalid runtime profile must fail with `AI_PROVIDER_CONFIG_INVALID` before any provider request，不再由 exact-profile lint 复制同一断言。

#### 10.2 GREEN: explicit non-thinking report profile and fail close

Set only `report.generate.default.default.params.thinking=disabled`, keeping its provider/model/route/fallback/capacity/output/timeout/TPM/version coordinates unchanged. Extend the openai-compatible adapter's shared Complete/Stream request builder with the official thinking object and a strict `enabled|disabled` allowlist. Keep report `response_format` out of profile params so F3's object output schema remains the only JSON-mode trigger.

#### 10.3 Verification and handoff

Run focused provider/profile tests, race tests, active-budget floor lint, owner context/index/docs gates and `git diff --check`; then run root `make test` for full frontend/backend regression。不要求真实 provider smoke。

### Phase 11: typed profile defaults and provider response cap

#### 11.1 RED: canonical active budgets and one owner contract

Canonical catalog/coverage lint must require all six active profiles (`judge.default`, `practice.chat.default`, `report.generate.default`, `resume.parse.default`, `resume.tailor.default`, `target.import.default`) to keep `max_tokens >= 16384`。Current typed defaults use `16384`, and report context default remains `1000000`。Keep one table-driven loader-owner contract for typed defaults, legal override and explicit invalid capacity; do not lock exact profile coordinates or repeat the same matrix for every profile/composition layer.

#### 11.2 GREEN: shared defaults and injected response limit

Encode matching `16384` per-profile code defaults for every active profile and keep the report context default at `1000000`. Route `ai.maxResponseBodyBytes=4194304` from A4 typed config into openai-compatible, judge-compatible, Doubao and MiniMax adapters; delete adapter-local 4MiB constants and replace unbounded response reads with a shared bounded reader. Test the shared reader and any genuinely distinct streaming/protocol truncation branch once; configuration injection itself is covered by types, construction and code review rather than four duplicate adapter matrices.

#### 11.3 Config/business boundary handoff

配置层不把 bytes 与 tokens 直接相加，不维护 report budget 测试或默认尺寸材料。Backend-review 的历史业务回归只证明其自身 provider path，不作为 profile 配置传播证据。No `input-*.json` boundary files are committed；E2E 与真实 provider smoke 均不是配置 gate。

#### 11.4 Verification and handoff

Run the active-budget floor lint, the single loader default/override/invalid contract, focused shared-response-cap protocol tests, config lint, context/docs validation, root `make test` and `git diff --check`. Do not start a scenario environment or real provider smoke solely to prove profile defaults or wiring.

## 5 验收标准

- Provider registry schema、loader、secret env ref 解析与热加载已落地，负向 fixtures 覆盖重复 provider、未知 protocol、capability mismatch、网络出站 provider secret 缺失与 fallback 超限；`stub` provider 不需要伪造 secret。
- Model Profile schema 已迁移到 `capability` / `provider_ref` / `status`；repo-tracked active profiles 只使用 current schema keys；F3 6 个 baseline feature_key 的默认 profile 引用与 spec §4.5 Product/UI fail-closed profiles 均存在，且 `disabled` / `unsupported` profile 带 `unsupported_reason`。
- Repo-tracked Model Profile active truth source 已收敛为单一 `config/ai-profiles.yaml`；loader、coverage lint、bootstrap、A4 env/config 默认值和 README 均使用 catalog 文件路径，active scope 不再引用 per-profile YAML files。
- AIClient 路由与 fallback 由 A3 中央执行；业务代码没有 retry-with-different-model 循环；fallback meta / metric / log 完整。
- Unsupported capability fail-closed；STT / realtime / judge 在 adapter 激活前不会静默降级，并通过 B1-owned `AI_UNSUPPORTED_CAPABILITY` 或同义 approved `AI_*` code 对外表达。
- 当前 active scope 不含向量化 / 重排代码与基础设施；chat profiles 全部指向 `deepseek` provider ref 且模型 ID 只使用 `deepseek-v4-flash` / `deepseek-v4-pro`。
- A4 env/config 字典、B1 shared vocabulary、F3 + Product/UI profile coverage lint、A3 docs/README/fixtures 全部同步。
- 隐私红线与零厂商 SDK 红线保持；全局 gate 与 context validation 通过。
- 六个 active profile 的 `max_tokens` 均不低于 16384；coverage lint 只锁定 active 集合与 floor，一处 loader default/override/invalid owner contract 通过，不存在跨 composition/domain/scenario 的重复配置断言。
- `report.generate.default` 使用 1M context、16K output typed defaults 与 thinking disabled；不做 bytes+tokens 公式、exact-profile lint、report budget test 或真实 provider smoke。
- 四个 provider adapter 统一消费 `ai.maxResponseBodyBytes=4194304` 注入值；共享 bounded reader 与必要的 streaming/protocol 截断分支有 focused evidence，不存在 adapter-local 4MiB 或无界 response `ReadAll`，也不按 adapter 复制配置 propagation matrix。
- `judge.default` 使用至少 16K output 与 non-thinking JSON；reasoning-only 响应在 adapter 内脱敏 fail-closed。下游业务若验证 judge 行为，也不承担 profile 默认值或容量配置证明。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 迁移 profile schema 时业务或测试继续使用 `task_type` | Phase 2.1 只接受 current keys；profile loader negative test 与 active-scope search 同时拦截 out-of-scope key |
| Provider registry 引入 secret 明文泄漏风险 | Registry 只保存 env key 名；A4 SecretSource 解析实际值；redaction tests 覆盖 config dump / error wrapping / log |
| Central fallback 变成无界重试，放大成本或延迟 | Spec 锁最多 2 跳；fallback 只在 profile `when[]` 命中时执行；metric/log 必须记录每跳 |
| STT / realtime profile 已存在但 adapter 未实现，业务误以为可用 | Unsupported profile 必须 disabled 或 fail-closed；UI voice workstream 在 adapter 未激活前必须 feature-gated |
| F3 新增 feature_key 或 Product/UI 新增 AI 场景但 A3 profile catalog 未跟进 | Phase 2.3 / 4.3 profile coverage lint 拦截；新增 AI 场景必须同步 spec §4.5、F3 字典与 profile catalog |
| A4 env 字典与 A3 registry schema 漂移 | Phase 4.1 将 env/config 字典、bindings、validator 与 lint-config 作为同一阶段交付 |
| 单一 catalog 文件变大导致未来多人冲突 | 当前 17 个 profile 规模优先降低文件碎片；若未来 profile 数量或 owner 并发显著增加，再由 A3/F3 plan 显式重新评估目录型 catalog |
| Judge 默认 thinking 先耗尽 max tokens，final content 为空 | profile 显式关闭 thinking 并将 final JSON budget 提升到 16,384；adapter 不使用或泄漏 reasoning，profile 坐标由 catalog/lint 锁定，业务行为验证不重复承担配置容量证明 |

## 7 Owner Handoff

- **002 / C14**：可直接基于 `capability=stt|realtime` profile 激活 speech adapter；adapter 未实现前保持 `status=unsupported` 或 `disabled`，不得在业务侧新增 provider 配置。
- **practice / report / resume**：业务代码继续只消费 F3 Resolve 返回的 `model_profile_name`，由 A3 `AIClient` 解析 provider ref、capability、fallback 与 secret。
- **F3 eval**：新增 judge / eval 场景时先同步 F3 feature_key 字典与 profile catalog；A3 profile coverage lint 会拦截缺失或 out-of-scope schema key。

## 8 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-14 | 1.16 | Raise all six active profile budgets to a 16K floor and replace exact-coordinate, formula, scenario and real-provider capacity gates with one loader owner contract. | A4 Phase 13 |
| 2026-07-14 | 1.15 | Reopen Phase 11 for typed profile code defaults, shared 4MiB provider response cap and 896KiB report capacity handoff. | A4 Phase 13 + backend-review/001 |
| 2026-07-13 | 1.14 | Add Phase 10 for report generation non-thinking structured output: official openai-compatible thinking wire, loader/adapter/lint fail-close, and output-schema-owned JSON mode. | backend-review/001 |
| 2026-07-12 | 1.13 | Historical judge budget migration; current gate retains non-thinking JSON and privacy-safe empty-final failure only. | F3/004 |
| 2026-07-12 | 1.12 | Historical report capacity exploration; superseded by typed defaults and loader owner validation. | backend-review/001 |
| 2026-07-12 | 1.11 | Historical report budget migration; old exact values are not current gates. | backend-review/001 |
| 2026-07-10 | 1.10 | 删除仅由测试自证的 provider 启动错误码映射层，保留哨兵错误合同。 | tech-debt pruning |
| 2026-07-10 | 1.9 | 统一 schema、model alias 与 profile directory 的 out-of-scope 术语，修正 completed-state 表述并对齐 checklist 版本。 | tech-debt pruning |
| 2026-07-10 | 1.8 | 将 Product/UI capability 描述收敛为 fail-closed profile，并对齐当前电话模式术语。 | tech-debt pruning |
| 2026-07-10 | 1.7 | 删除范围外文本输入 STT profile coverage，电话模式 STT/TTS profile 继续作为 Product/UI capability coverage。 | tech-debt pruning |
| 2026-07-07 | 1.6 | 对齐当前 9 个 baseline feature_key、DeepSeek chat baseline、speech / judge profile 边界和 out-of-scope schema wording。 | product-scope cleanup |
| 2026-05-08 | 1.5 | 原地修订：删除向量化 / 重排当前实现与基础设施，开发期 AI provider 收敛到 DeepSeek V4 Flash/Pro。 | change-intake user-approved revision |
| 2026-05-05 | 1.4 | 原地修订 L2 remediation：修复 dev-stack / product owner matrix out-of-scope profile directory 漂移，并补强 deploy/profile semantic drift gate。 | plan-code-review --fix |
| 2026-05-05 | 1.3 | 原地修订 catalog consolidation：将 per-profile YAML directory active truth source 收敛为单一 `config/ai-profiles.yaml`，并同步 loader、lint、A4/F3 docs 与验证 gate。 | change-intake user-approved revision |
| 2026-05-05 | 1.2 | 原地修订 L2 remediation：补 runtime registry/profile wiring、profile reload warn、active profile anti-stub gate 与 post-fix verification。 | plan-code-review --fix |
| 2026-05-05 | 1.1 | Phase 5 完成：全局 gate 与 active-scope 负向搜索通过，plan 生命周期切为 completed，并补充后续 owner handoff。 | implementation closeout |
| 2026-05-05 | 1.0 | 初始创建：承接 A3 spec v1.9，规划 provider registry、capability profile、central fallback、A4/B1/F3 联动与验证门禁。 | design crystallization |
