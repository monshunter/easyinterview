# F3 Output Schema Contract: 语言无关 schema 真理源 + prompt 契约渲染 + Resolve 接线 + 校验闭环

> **版本**: 2.16
> **状态**: active
> **更新日期**: 2026-07-21

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [prompt-rubric-registry spec](../../spec.md) v2.20 §3.1 D-13 锁定的 `output_schema` 契约端到端落地，关闭 §6 C-12。完成后：

- `config/prompts/<feature_key>/v0.1.0.schema.json` 覆盖当前 **9 个 chat feature_key**，每个是**语言无关**的 JSON Schema 子集（校验关键字 `type` / `required` / `properties` / `items` / `enum`，允许 `description` 非校验注解），描述后端实际反序列化契约；当前 9 个 schema 顶层均为 `object`。
- 当前 9 个 canonical `multi` prompt body 的输出段使用 schema 渲染/校验的「输出契约 + complete example JSON output」写法；schema 是唯一字段真理源，prompt body 不手工维护第二份字段清单；example 必须是完整代表性 JSON output（覆盖 schema 声明的 required + optional 字段，使用业务形态值），且明确不是 JSON Schema / OpenAPI schema；改动的 `.md` 重算 `template_hash`。
- `config/prompts/README.md` 固化 output schema 约定、`description` 注解策略与 prompt body 输出契约渲染/校验写法。
- `scripts/lint/prompt_lint.py` 新增 schema gate：① schema 只用允许关键字（含 `description` 注解）；② 每个 chat feature_key 有 schema（语音 feature_key 豁免）；③ schema `required` ⊆ prompt body 声明的输出 key；④ schema 字段与后端反序列化 struct 的 json tag 一致（drift → exit 1）；⑤ prompt body 输出契约块与 schema 可重渲染结果一致，complete example JSON output 包含 required + optional 字段、使用业务形态值、明确不是 JSON Schema / OpenAPI schema，并通过 schema 校验。
- `backend/internal/ai/registry/` loader 加载语言无关 schema（不混入 per-language `template_hash`），`ResolveActive` 输出非空 `OutputSchema`。
- A3 `backend/internal/ai/aiclient/observability` 的 `validateOutputSchema` 扩展支持 `enum`；模型输出违反 `enum` 或缺 `required` → `AI_OUTPUT_INVALID` fail-close。
- 各当前 chat feature_key 的 caller 把 `resolution.OutputSchema` 透传进 `CallMetadata.OutputSchema`，运行时 fail-close 全量生效；语音（STT/TTS）路径跳过。
- 修复 `resume.tailor.bullet_suggestions` prompt 输出 key 与 struct json tag 不一致（若 Phase 0 确认为 drift）。

本 plan **不**切真实 Model Profile（推到 003）、**不**实现 LLM Judge（003）、**不**实现灰度（004）、**不**向 provider 下发 `response_format`（归 A3 后续）。schema 只描述业务输出形状，不含 provider / model / endpoint / SDK 私有字符串。

## 2 背景

spec v1.9 / D-12 起即规划 Resolve 可输出 provider-neutral `output_schema`，但 `001-baseline` 明确预留 `OutputSchema` 字段「不消费」（`backend/internal/ai/registry/resolver.go:62` 硬编码 `OutputSchema: nil`），prompt body 仅以 prose 描述输出形状。当前事实：

- 输出形状以自然语言写在每个语言变体 `.md`（prose + 双份重抄），无机器可读 schema，无 drift gate；BUG-0065 即此类 drift 的案例。002 修复不能把 prose drift 变成多份手写字段表 drift；schema 必须成为唯一字段真理源，prompt 只承载由 schema 渲染/校验的可读契约块。
- A3 `aiclient` observability 已有 `validateOutputSchema`（`decorator.go:538`），递归校验 `type` / `required` / `properties` / `items`，但**不支持 `enum`**；且仅当 `CallMetadata.OutputSchema` 非空时触发。002 负责把当前仍保留 caller 的 `OutputSchemaVersion` provenance 标签和实际 `OutputSchema` 区分清楚，并接通 fail-close。
- 当前 9 个 chat feature_key 反序列化对齐分三档：6 个完全对齐；`practice.*` 三个 prompt 多声明字段、parser 只取部分（有意）；`resume.parse` / `resume.tailor.gap_review` 停在 `json.RawMessage`；`resume.tailor.bullet_suggestions` prompt key（`rewrite` / `why_better` / `kept_facts`）与 struct json tag（`SuggestedBullet` / `Reason` / `OriginalBullet`）疑似不一致。

D-13 把 `output_schema` 从「可追加」升级为可机器校验的锁定契约。本 plan 在一个 vertical slice 内落地。

## 2.1 修订记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-21 | 2.16 | Reopen Phase 15 for the immutable practice interviewer-identity v0.3 prompt/schema pair, exact rollback coordinate, and gated dev/DB activation. |
| 2026-07-13 | 2.15 | Keep the complete report JSON example, pair it with synthetic candidate input, add anti-copy/current-context regeneration, and synchronize hash/migration/lint evidence. |
| 2026-07-13 | 2.14 | Replace one-shot completion/judge assumption with independent max4 evalkit budgets；preserve dynamic scope/full validation and typed judge retry boundary；marker pending. |
| 2026-07-13 | 2.13 | Require evalkit to reuse product full semantic validator；sole-label targeted repair，all other/mixed whole-report repair，full revalidation and one total budget；marker remains pending. |
| 2026-07-13 | 2.12 | Finalize A：action schema fuse200 code points；prompt semantic bound24 whitespace words/64 Unicode code points；targeted repair internal margin18/52；reopen marker. |
| 2026-07-13 | 2.11 | A-200：action schema maxLength200 fuse；prompt/UX remains14/40 and desktop+390 typed-invalid/no-raw gate. |
| 2026-07-13 | 2.10 | Separate wire fuse from14/40 quality；specific bound superseded by A-200. |
| 2026-07-13 | 2.9 | Correct action support by type；require one directly cited missing behavior per focused retry code and reject umbrella-only labels；forbid invented review artifact/gap/scenario/transfer task；seal focus to the exact-set table. |
| 2026-07-12 | 2.8 | Pin active report v0.2 Resolve provenance to report-context.v1 while practice v0.2 retains registry.v1. |
| 2026-07-12 | 2.7 | Extend Phase 14 with immutable practice semantic-focus v0.2 prompt/schema + content-identical rubric, eight-status dev snapshot activation and one 000019 transaction for both report/practice pairs. |
| 2026-07-12 | 2.6 | Enforce report v0.2 recursive object closure and locked string/array/numeric bounds in lint, loader and the shared runtime validator before re-emitting GREEN markers. |
| 2026-07-12 | 2.5 | Define rubric status metadata and separate dev file atomic-snapshot activation/rollback from staging/prod DB transaction; 002 may mutate only activation status on 004-owned rubric content. |
| 2026-07-12 | 2.4 | Split report v0.2 ownership: this plan owns prompt/schema, multi-version loading and gated final activation; 004 solely owns rubric/judge/eval. |
| 2026-07-12 | 2.3 | Reopen Phase 14 for grounded direct report prompt/schema/struct exactness and old-key rejection. |
| 2026-07-10 | 2.2 | Add a downstream cmd/api scenario-fixture gate for canonical parser changes after repairing the Practice hint fixture drift. |
| 2026-07-10 | 2.1 | 原地增加 Phase 12：删除 practice 模型输出 parser 的 `question` / `intent` / `hint` / `answer_summary` alias，并移除 F3 lint/Go 注释中的固定 spec 版本引用。 |
| 2026-07-10 | 2.0 | 原地增加 Phase 11：删除 resume tailor 无当前输入来源的 output alias，parser 只接受 schema canonical keys。 |
| 2026-07-10 | 1.9 | 将 prompt example 的 generic placeholder 表述收敛为 filler values，不改变 renderer 合同。 |
| 2026-07-10 | 1.8 | 删除范围外文本输入 STT schema 豁免口径；语音 schema 红线只承接电话模式 STT/TTS feature_key。 |
| 2026-07-07 | 1.7 | Wording cleanup：收敛 migration audit、seed net-state 与 parser alias 说明为当前 out-of-scope gate 口径，不改变 prompt/rubric 可执行契约。 |
| 2026-07-06 | 1.6 | D-16 / D-22 后裁剪复查：当前 prompt/rubric truth source 为 9 个 chat feature_key；Debrief / JD Match 仅保留 migration delete rows 与负向 lint 语境，不再作为 plan/context/checklist 当前正向 target surface。 |
| 2026-05-24 | 1.5 | L2 remediation：修复 `jd_match.recommendation` schema 将 `posted` 误标 required，而生产 jobs pool 与 generator contract 未提供该字段，导致 fail-close 路径可能迫使模型编造 freshness；同时修复 prompt lint 在 schema 缺少 `description` 时抛 traceback 而非返回 lint diagnostic 的工具缺口。 |
| 2026-05-24 | 1.4 | L2 remediation：修复 seed migration 静态门禁只覆盖固定 11 个 feature_key 的 false-green；新增动态 seed coverage test，从 `config/prompts` / `config/rubrics` 反推 active 坐标，补充 `jd_match.recommendation` / `jd_match.search` prompt/rubric seed migration，并用 `make migrate-check` 验证迁移链。 |
| 2026-05-24 | 1.3 | L2 remediation：将 prompt example 从最小 required-only JSON 升级为完整代表性 JSON output，覆盖 schema 声明的 optional 字段，使用业务形态值，并在 prompt block 中明确不是 JSON Schema / OpenAPI schema；同步 README、renderer/lint gate、26 个 prompt body、YAML hash 与 seed migration。 |
| 2026-05-24 | 1.2 | L2 remediation：`validateOutputSchema` 拒绝 schema-valid JSON 后追加 trailing prose / token 的模型输出。 |
| 2026-05-23 | 1.1 | 完成 output schema contract 主体交付并收口到 completed。 |

## 3 质量门禁分类

- **Plan 类型**: `truth-source + contract + tooling + code-internal（cross-domain wiring）`。落地 `config/` 语言无关 schema 真理源 + prompt body 契约统一 + `scripts/lint/` 静态 gate + registry 加载 / resolver 接线 + A3 校验器 `enum` 扩展 + 当前 4 个 domain caller 集合（targetjob / resume/jobs / review / practice）的 `OutputSchema` 透传。不引入用户可见 UI、新 HTTP API 行为或新业务流。
- **TDD 策略**: Code plan requires TDD。所有 checklist item 先红后绿：① schema lint 先写 negative fixtures（schema 非法关键字 / `required` 不在 prompt / schema 字段与 struct json tag 不一致 / prompt 输出契约块手工漂移）再实现；② prompt 输出契约 renderer 先写 fixture 断言（同一 schema 对 multi/en 输出同一字段顺序、complete example JSON output 覆盖 required + optional 字段、使用业务形态值、schema-valid 且不是 JSON Schema / OpenAPI schema）再落地；③ registry loader/resolver 先写「加载 schema + `ResolveActive` 输出非空 OutputSchema + 语言无关单份」断言再改实现；④ `validateOutputSchema` enum 先写「enum 违反 → error」focused test 再加字段；⑤ caller 透传先写「`metadata.OutputSchema` 非空 + 端到端 fail-close」断言再改 wiring；⑥ `bullet_suggestions` key 修复先写 round-trip 断言；⑦ canonical parser 删除 alias 前先写 alias-only fail-close 与 canonical success 测试，并同步执行消费该 parser 的 `cmd/api` 确定性场景，防止 downstream fixture 仍生成旧 key。每个 phase 退出 gate 都是可执行命令。
- **BDD 策略**: **BDD 不适用**。本 plan 落地内部契约（schema 真理源 + lint + registry 接线 + provider-neutral 校验器），不新增用户可见 UI、新 HTTP API 行为或端到端业务工作流；`validateOutputSchema` fail-close 复用既有 `AI_OUTPUT_INVALID` 错误路径，仅扩展生效范围与 `enum`，无新端到端用户行为。后续 P0 用户行为流由各 C 域 plan 维护 BDD/E2E gate。
- **替代验证 gate**: ① `make lint-prompts`（schema 子集 + prompt 契约块可重渲染 + complete example JSON output schema-valid + 三向一致性 drift gate + negative fixtures）；② `go test ./backend/internal/ai/registry/... -race`（loader 加载 schema + `ResolveActive` OutputSchema 非空 + 语言无关单份 + fallback 仍返回同一 schema + seed migration 覆盖全部 active prompt/rubric 坐标）；③ `go test ./backend/internal/ai/aiclient/... -race`（`validateOutputSchema` enum/required fail-close）；④ `go test ./backend/internal/{targetjob,resume/jobs,review,practice}/... -race`（caller `OutputSchema` 透传 + 端到端 fail-close）；⑤ canonical parser 变更后执行命中 consumer 的 `backend/cmd/api` 场景测试；⑥ schema↔struct drift negative fixture；⑦ grep red-line：语音 feature_key（STT/TTS）不落 schema、不接 `OutputSchema`、不发 `response_format`；⑧ out-of-scope feature_key negative lint；⑨ `python3 .agent-skills/implement/shared/scripts/validate_context.py` + `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`；⑩ 顶层 `make lint`；⑪ `DATABASE_URL=... make migrate-check`（Postgres dev-stack 下执行迁移 `up -> down -> up` + migration lint）。

### 3.1 Cross-layer Operation Matrix（9 chat feature_key）

本 plan 不新增 HTTP operation，[development.md §2.1](../../../../development.md) 标准字段状态：`operationId` = N/A（内部 feature_key，非 HTTP）；`fixture` = 本 plan 不新增/变更；`frontend consumer` = 无（不改 UI）；`backend handler` = 下表反序列化 struct 列；`persistence` = 沿用各 feature_key 既有 `ai_task_runs` provenance，本 plan 不新增/变更业务表；`AI dependency` = 各 feature_key §3.1.1 默认 model profile；`scenario coverage` = §3 替代验证 gate（BDD 不适用）。下表聚焦本 plan 真正改变的 cross-layer 维度（prompt ↔ schema ↔ struct ↔ caller）：

Schema canonical key policy：`required` 只覆盖后端实际依赖字段；受 output schema 约束的运行路径只接受 schema canonical keys。Phase 0 发现的 out-of-scope alias 只作为 drift 证据或 negative-test 输入，不得进入 prompt 或 runtime contract。若保留可选 prompt-only 字段，必须说明评估价值，并用 schema `description` + lint 证明不是第二份手写字段表。

| feature_key | 顶层 | schema 文件（语言无关） | 后端反序列化 struct（file:line） | caller 透传方式 | enum 字段 |
|---|---|---|---|---|---|
| `target.import.parse` | object | `target.import.parse/v0.1.0.schema.json` | `parseAIResponse` (`targetjob/parse_executor.go:90`) | 本地 `targetjob.PromptResolution` + `RegistryAdapter` | `kind`, `evidenceLevel` |
| `practice.session.first_question` | object | `.../v0.1.0.schema.json` | `firstQuestion` (`practice/session_starter.go:332`) | 直接 `registry.PromptResolution` | – |
| `practice.session.follow_up` | object | `.../v0.1.0.schema.json` | `firstQuestion` 复用 (`practice/append_session_event_service.go:263`) | 直接 | – |
| `practice.turn.lightweight_observe` | object | `.../v0.1.0.schema.json` | `hintAIResponse` (`practice/hint_ai.go:19`) | 直接 | `severity` |
| `report.generate` | object | `.../v0.1.0.schema.json` | `ReportContentDraft` (`review/generate_report.go:48`) | 直接 | – |
| `report.question_assessment` | object | `.../v0.1.0.schema.json` | `QuestionAssessmentDraft` (`review/generate_report.go:75`) | 直接 | `status`(由 `score_level` 映射) |
| `resume.parse` | object | `.../v0.1.0.schema.json` | `json.RawMessage` 仅查 6 顶层 key (`resume/jobs/parse.go:248`) | 本地 `resume/jobs.PromptResolution` + adapter | – |
| `resume.tailor.gap_review` | object | `.../v0.1.0.schema.json` | `decodedTailorOutput`(RawMessage) (`resume/jobs/tailor.go:287`) | 本地 + adapter | `severity` |
| `resume.tailor.bullet_suggestions` | object | `.../v0.1.0.schema.json` | `decodedTailorSuggestion` + `normalizeSuggestions`；只接受 `originalBullet` / `suggestedBullet` / `reason` canonical keys | 本地 + adapter | – |
> 语音 feature_key（`practice.voice.stt` / `practice.voice.tts`，对应 `doubao_speech` / `minimax_speech`）不产 JSON content，**不在本 plan 范围**，不落 schema、不接 `OutputSchema`。Debrief / JD Match feature_key 为 out-of-scope，不出现在当前 prompt/rubric truth source。

## 4 实施步骤

### Phase 0: 基线复核与缺口确认（只读 + handoff snapshot）

#### 0.1 9 个 chat feature_key 输出形态 + struct 测绘
复核 §3.1 operation matrix 每行的 prompt body 输出 key 与后端 struct json tag；记入 §8 Phase 0 handoff。

#### 0.2 `resume.tailor.bullet_suggestions` 映射核实
读取 `backend/internal/resume/jobs/tailor.go` 的 `decodeTailorAIResponse` / `normalizeSuggestions`，判定当时的 prompt key（`rewrite`/`why_better`/`kept_facts`）与 canonical parser key（`originalBullet`/`suggestedBullet`/`reason`）属于新契约 drift；Phase 2.3 统一 prompt/schema，Phase 11 删除 runtime alias，结论记入 handoff。

#### 0.3 校验器能力复核
读取 `backend/internal/ai/aiclient/observability/decorator.go:531-594`：确认 `outputSchema` 类型（`type`/`required`/`properties`/`items`）与 `validateAgainstSchema`，记录 `enum` 缺口与扩展点。

#### 0.4 加载 / 接线现状复核
读取 `loader.go`（`WalkDir` 只认 `.yaml`、`computeTemplateHash` per-language 5 字段）、`resolver.go:62`（`OutputSchema: nil`）、`types.go:15`（`PromptResolution.OutputSchema *json.RawMessage`）；确认 schema 不混入 per-language `template_hash` 的设计可行。

#### 0.5 caller 接线清单
列出直接型 caller（`review`×2 / `practice`-chat×3）与本地-resolution 型 caller（`targetjob` / `resume/jobs`×2）；区分既有 `OutputSchemaVersion`（标签）与待接的 `OutputSchema`（实际 schema）；标注 `practice/voice_turn_service.go` STT/TTS 跳过点。记入 handoff。

### Phase 1: README 契约 + prompt 输出契约渲染规则

#### 1.1 `config/prompts/README.md` 契约扩写
新增 output schema 约定：文件落点 `<feature_key>/<version>.schema.json`、**语言无关**（每个 `(feature_key, version)` 一份）、JSON Schema 校验子集关键字、`description` 非校验注解、不混入 `template_hash`、与 prompt body 输出 key / 后端 struct json tag 的一致性要求。

#### 1.2 prompt body 输出契约块规范
在 README 固化 prompt body 输出段只能是 schema 渲染/校验的契约块：字段顺序来自 schema `required` + `properties`，字段说明来自 `description`，complete example JSON output 由 schema 生成并可解析，且必须覆盖 schema 声明的 required + optional 字段、使用业务形态值、明确不是 JSON Schema / OpenAPI schema；人手只编辑 schema 与业务 prompt 上下文，不在当前 18 个 `.md` 中独立维护第二份字段表。

#### 1.3 alias / optional 字段记录规则
在 README / handoff 规则中明确：受 output schema 约束的 parser 只接受 canonical keys；out-of-scope alias 只允许作为 negative-test 输入或历史 drift 证据。prompt-only 可选字段必须有 `description` 说明评估价值，并由 lint 证明与 schema/rendered block 一致。

### Phase 2: 9 份语言无关 output schema 文件

#### 2.1 落地 9 份 `v0.1.0.schema.json`
按 §3.1 为每个当前 chat feature_key 写语言无关 schema：顶层 `type` 为 `object`；`required` = 后端实际依赖字段；`enum` 覆盖矩阵列出的受限取值；`description` 覆盖 required 字段、重要可选字段与 enum 语义；只用允许关键字。

#### 2.2 schema ↔ struct canonical key 对齐
人工核对每份 schema 的 `required`/字段名与后端 struct json tag / parser 实际依赖字段一致（Phase 3 lint 强制）。对停在 `json.RawMessage` 的 feature_key（`resume.parse`、`resume.tailor.gap_review`，无强类型 struct json tag），三向对齐降级为「schema ↔ parser required key 一致」，并在 §8 handoff 标注；本 plan 不为这两个 feature_key 引入强类型 struct（留各 C 域 plan）。

#### 2.3 `resume.tailor.bullet_suggestions` canonical key 修复（条件项）
统一 schema/prompt canonical key 为 `originalBullet` / `suggestedBullet` / `reason`，并补 round-trip 断言；Phase 11 删除 parser alias fallback，并用 alias-only fail-close 回归锁定 canonical-only 合同。

### Phase 3: schema lint + prompt 契约 renderer（先红后绿）

#### 3.1 `prompt_lint.py` schema gate + 单测
扩展 `scripts/lint/prompt_lint.py`（或新增 `schema_lint.py` 复用同一 README 描述）：① schema 只含 `type`/`required`/`properties`/`items`/`enum` + `description`；② 每个当前 chat feature_key 有 schema 文件（语音 feature_key 豁免，维护豁免清单）；③ schema `required` ⊆ 对应 prompt body 声明输出 key；④ schema 字段 ↔ 后端 struct json tag / parser required key 一致。停在 `json.RawMessage` 的 `resume.parse` / `resume.tailor.gap_review` 因无 struct json tag，④ 降级为 schema ↔ parser required key 一致。negative fixtures ≥4（非法关键字 / required 不在 prompt / 字段与 struct 不一致 / prompt contract block 手工漂移），先红后绿。

#### 3.2 prompt output contract renderer
新增可测试 renderer（可内嵌于 `prompt_lint.py` 或独立 helper）：从 schema 生成稳定的 prompt 输出契约块和 complete representative example JSON output；同一 feature_key 的 multi/en body 必须渲染出同一字段结构，仅语言说明文本可不同。单测覆盖 object 顶层、array 顶层、enum、nested object/array、`description` 缺失报错、example JSON 覆盖 optional 字段、使用业务形态值且可被 schema gate 接受。

#### 3.3 Makefile 接线
`make lint-prompts` 覆盖 schema gate + rendered prompt contract gate；顶层 `make lint` 联动。

### Phase 4: 9 × 2 prompt body 输出段统一（由 schema 渲染/校验）

#### 4.1 生成/更新 prompt 输出契约块
把当前 9 个 feature_key 的每个 `.md` 的 “Return strict JSON …” prose 段替换为 schema 渲染/校验的「输出契约 + complete example JSON output」块，与 Phase 2 schema 对齐；只改输出契约段，不改角色/任务/业务规则措辞。每个改动的 `.md` 重算 `template_hash` 写回同名 `.yaml`。

#### 4.2 prompt body drift gate
`make lint-prompts` 证明 schema、prompt contract block、complete example JSON output、YAML hash 与 seed migration hash（若涉及）全部一致；故意改一个 prompt contract 字段名或 example key 必须失败。L2 remediation 要求 example 不再是最小 required-only JSON；renderer 必须覆盖 schema 声明的 optional 字段，并用业务形态值替代 generic `string` / `1` filler values。

### Phase 5: registry 加载 + resolver 接线 `OutputSchema`（先红后绿）

#### 5.1 `loader.go` 加载语言无关 schema
`readPrompt` 时按 `<feature_key>/<version>.schema.json` 加载一次（语言无关，所有 language 变体共享同一份），存入 snapshot；不混入 `computeTemplateHash`；chat feature_key 缺 schema → 启动 fail，语音 feature_key 豁免。focused tests 覆盖加载 / 缺失 / 语言无关共享。

#### 5.2 `resolver.go` 填充 `OutputSchema`
`ResolveActive` / `GetPrompt` 用加载的 schema 替换 `OutputSchema: nil`；language fallback 时仍返回同一份语言无关 schema。focused tests 断言非空 + 各 language 返回一致。

### Phase 6: A3 `validateOutputSchema` 扩展 `enum`（先红后绿）

#### 6.1 `decorator.go` outputSchema 加 `enum`
`outputSchema` struct 加 `Enum []any json:"enum"`；`validateAgainstSchema` 增加 enum 成员校验分支（值不在 enum → error）。`description` 属 schema annotation，A3 校验器不需要解析或影响校验结果。

#### 6.2 aiclient 测试
focused tests：enum 违反 → `AI_OUTPUT_INVALID`；缺 required 仍 fail；合法输出（含 array 顶层）通过。

#### 6.3 L2 strict JSON trailing token remediation
`validateOutputSchema` 必须把模型输出视为单个完整 JSON document 校验：第一个 schema-valid JSON value 后若还存在非空 trailing token / prose，仍返回 `AI_OUTPUT_INVALID`，避免「JSON + 解释文本」绕过 prompt output contract 的 strict JSON 要求。补 focused negative test 后最小实现。

### Phase 7: caller 端到端透传 + 收口

#### 7.1 `targetjob` 透传
`targetjob.PromptResolution` 加 `OutputSchema` 字段；`RegistryAdapter.Resolve` 映射；`parse_executor.go` 填 `CallMetadata.OutputSchema`。cross-layer test 断言透传 + fail-close。

#### 7.2 `resume/jobs` 透传
`resume/jobs.PromptResolution` 加 `OutputSchema` + adapter 映射；`parse.go` / `tailor.go` 填 metadata；测试。

#### 7.3 `review` / `practice`(chat) 透传
直接读 `resolution.OutputSchema` 填 `CallMetadata.OutputSchema`；`practice/voice_turn_service.go` STT/TTS 路径跳过，仅 chat（follow_up）接线；测试。

#### 7.4 Out-of-scope feature_key negative gate
确认 out-of-scope feature_key 不存在于 `config/prompts` / `config/rubrics`，且 `prompt_lint` / `rubric_lint` 对 out-of-scope key 或范围外模块字段保持 fail-fast。

#### 7.5 grep red-line
`grep` 确认语音 feature_key 不落 schema、不接 `OutputSchema`；业务包不出现 `response_format` / `json_schema` 请求字段（归 A3）。

#### 7.6 收口
`make lint-prompts` + `go test ./backend/internal/ai/registry/... ./backend/internal/ai/aiclient/... ./backend/internal/{targetjob,resume/jobs,review,practice}/... -race` + `validate_context.py` + `sync-doc-index --check`；§8 handoff 列 C-12 命令证据；plan/checklist Header 切 `completed`，同步 INDEX 与工作日志。

### Phase 8: L2 prompt example remediation（完整 JSON output）

#### 8.1 renderer red/green gate
为 `scripts/lint/prompt_lint.py` 增加 focused regression：example JSON 必须包含 schema-declared optional properties，且不能使用 `string` / `1` 这类 generic filler values；先观察 red，再最小实现。

#### 8.2 prompt truth source regeneration
重渲染 9 × 2 个 prompt body 的 contract block；刷新 18 个 YAML `template_hash`；同步 `migrations/000002_seed_baseline_prompt_rubric_versions.up.sql` 中已有 prompt seed row 的 body/hash；保持 seed migration 与 on-disk truth source 一致。

#### 8.3 documentation and lifecycle closure
同步 `config/prompts/README.md`、spec v2.6、plan/checklist v1.3、context discovery 与 INDEX；执行 `make lint-prompts`、prompt lint 单测、context validation、docs/index gate 与 `git diff --check` 后确认 plan/checklist 状态为 `completed`。

### Phase 9: L2 seed migration coverage remediation（active truth source 全量覆盖）

#### 9.1 seed coverage static gate
重写 `backend/internal/ai/registry/db_integration_test.go` 的 seed coverage 断言：不再写死固定 feature_key / seed row 数，而是从 `config/prompts/*/v*.yaml` 的 `status: active` 与 `config/rubrics/*/v*.yaml` 反推当前 truth-source 坐标，并扫描所有 `migrations/*seed_baseline_prompt_rubric*.up.sql` 及后续删除迁移的净效果。缺行、重复行、额外行或 prompt `template_hash` drift 都必须失败。

#### 9.2 out-of-scope feature_key seed net-zero gate
对后续删除迁移中的 out-of-scope feature_key 行执行净效果校验：out-of-scope seed 可以作为 migration audit rows 保留，但当前迁移链落库后的 active prompt/rubric coordinate 必须只剩 9 个 feature_key × 2 language，out-of-scope key 不得作为 active truth source 回流。

#### 9.3 migration/runtime gate closure
执行 focused red/green 与聚合 gate：`go test ./backend/internal/ai/registry -run TestSeedMigrationCoversBaselineFeatureKeys -count=1 -v`、`python3 scripts/lint/{prompt_lint,rubric_lint,migrations_lint}.py`、`go test ./backend/internal/ai/registry -count=1`、`DATABASE_URL=... make migrate-check`、context/index/docs/whitespace gates。完成后同步 checklist / INDEX / Bug 记录 / retrospective。

### Phase 10: L2 review follow-up（out-of-scope key rejection + lint diagnostic hardening）

#### 10.1 Out-of-scope feature_key rejection
删除当前 config truth source 中的 out-of-scope feature_key，并补 lint 回归测试：out-of-scope key 目录不存在；若测试 fixture 人为写入 out-of-scope key，lint 必须返回 clear diagnostic，不允许被 seed coverage 或 schema renderer 当作 active coordinate。

#### 10.2 prompt lint invalid schema diagnostic
修复 `prompt_lint.py` 在 schema subset / contract 已失败时仍调用 renderer 的流程，避免缺 `description` 等 invalid schema 触发 `KeyError` traceback。新增 integration-style negative test，要求 lint 输出 `missing non-empty description` 并拒绝 `Traceback`。

#### 10.3 verification and closeout
执行 focused tests、prompt/rubric/migration lint、registry seed coverage 与 adjacent Go gate；同步 BUG-0098、Pattern 5 与 retrospective 报告。

### Phase 11: Resume tailor output alias removal

#### 11.1 Red: canonical-only parser contract

在 `backend/internal/resume/jobs/tailor_test.go` 增加 handler-level 回归：root-level `strengths_to_amplify` / `gaps` 和 suggestion `original_bullet` / `suggested_bullet` / `rewrite` / `why_better` / `whyBetter` 组成的 alias-only 输出必须返回 `AI_OUTPUT_INVALID`；canonical payload 必须继续成功，且冲突 alias 不得覆盖 canonical 值。

#### 11.2 Green: 删除 alias fallback

`decodeTailorAIResponse` 只读取 root `matchSummary` 与 suggestion `originalBullet` / `suggestedBullet` / `reason`。删除 root-level match-summary 归一化、snake_case / rewrite / why-better fallback、缺失 `originalBullet` 时回填请求输入的行为，以及因此失去调用方的 helper。

#### 11.3 Focused and adjacent verification

运行 resume jobs focused tests、`make lint-prompts`、prompt/rubric lint 与生产代码负向搜索。`config/prompts` / schema / eval fixture 保持 canonical，生产 parser 中 alias key 0 命中；模板输入变量 `{{original_bullet}}` 不属于模型输出合同，不在本 gate 删除范围。

#### 11.4 Documentation and lifecycle closure

把本 plan/checklist 中的范围术语统一为 `out-of-scope` / current contract，context 投影同步到 spec 2.19。删除 registry preflight 对 spec 精确版本号的脆弱断言，改为验证 D-13 与当前 9-key 稳定语义；运行 context/index/docs/diff/pruning gates 后确认状态为 `completed`。

### Phase 12: Practice output alias removal and stable F3 references

#### 12.1 Current-contract lock

将 F3 spec 提升到 2.20，明确 output-schema runtime parser 只消费 canonical keys；`config/prompts/README.md` 同步 canonical parser policy，四个 plan context 投影到 spec 2.20。

#### 12.2 Red: practice canonical-only parser tests

在 practice focused tests 中锁定：`parseFirstQuestion` 的 `question` / `intent` alias-only 输出必须返回 `AI_OUTPUT_INVALID`；`parseHint` 不接受 `hint`；`parseTurnObservation` 不消费 `answer_summary`；canonical keys 继续成功，冲突 alias 不覆盖 canonical 值。

#### 12.3 Green: delete practice output aliases

从 `parseFirstQuestion` 删除 `Question` / `Intent` 字段和 fallback；从 `hintAIResponse` 删除 `Hint` / `AnswerSummarySnake`，只读取 `cue` / `answerSummary`。测试 fixture 的成功响应全部改用 canonical keys，范围外 alias 只留在 12.2 negative tests。

#### 12.4 Stable owner references and verification

删除 `prompt_lint.py` / `rubric_lint.py` docstring、registry Judge 注释和测试注释中的固定 spec 版本号，保留稳定 D-9 / F3 语义引用。运行 practice focused/package tests、registry tests、prompt/rubric lint、固定版本与 alias 负向搜索、context/index/docs/diff/pruning gates 后确认状态为 `completed`。

### Phase 13: Canonical parser downstream contract gate

`practice.turn.lightweight_observe` canonical parser 改动由 practice owner 的 focused parser/service tests 验证：成功 fixture 输出 `cue` / `answerSummary`，alias-only `hint` 只能留在 invalid-output 负测；canonical 成功路径不得降级为 `session_wait` 或额外写入 `AI_OUTPUT_INVALID` task run。该契约属于代码层测试，不创建 BDD/E2E 场景。

## 5 验收标准

- spec §6 C-12 通过 §8 handoff 命令证据验证；§3 替代 gate 全部通过。
- 9 个 chat feature_key 各有 1 份语言无关 schema；prompt body 输出契约块可由 schema 重渲染，complete example JSON output 可解析、schema-valid、覆盖 schema 声明的 required + optional 字段、使用业务形态值且明确不是 JSON Schema / OpenAPI schema；`ResolveActive` 输出非空 `OutputSchema`；故意制造 schema↔prompt↔struct 不一致 → `make lint-prompts` 失败。
- seed migration 覆盖从当前 `config/prompts` / `config/rubrics` truth source 反推的 9 个 active chat feature_key × 2 language 坐标；out-of-scope seed rows 仅作 migration audit，不得作为 active prompt/rubric coordinate 回流；`make migrate-check` 在 dev-stack Postgres 下通过。
- `validateOutputSchema` 对违反 `enum`/缺 `required` 的模型输出 fail-close（`AI_OUTPUT_INVALID`）。
- resume tailor parser 只接受 schema canonical output keys；alias-only 输出 fail-close 为 `AI_OUTPUT_INVALID`，canonical 输出与持久化行为保持不变。
- practice 首题/追问与轻量观察 parser 只接受 schema canonical output keys；范围外 alias-only 输出进入既有 `AI_OUTPUT_INVALID` / degrade 路径，canonical 输出行为保持不变。
- canonical parser 收紧后，相关 `cmd/api` 确定性成功 fixture 与 schema 同步，且 consumer HTTP 场景保持成功。
- 语音 feature_key 零 schema、零 `OutputSchema` 接线（grep red-line 0 命中）。
- `docs/spec/prompt-rubric-registry/plans/INDEX.md` 写入 002 行；四个 plan context 投影到 spec v2.20；`spec.md` Header v2.20 与 `docs/spec/INDEX.md` 一致。
- BDD 不适用声明固化在本 plan §3 + checklist BDD-Gate 区段，并附替代 gate 命令。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| schema 语言无关 vs `template_hash` per-language 语义错配 | schema **不**混入 per-language `template_hash`；用独立三向一致性 lint gate（schema↔rendered prompt contract↔struct）保证不漂移；loader 加载一份共享给所有 language。 |
| prompt body 可读性与重复维护冲突 | schema 加 `description`，prompt 输出段由 renderer 生成/校验；人只维护 schema 一处，`make lint-prompts` 拒绝手工漂移。 |
| caller 透传面广（5 domain） | 每处改动小（直接型填一行；本地-resolution 型加字段 + adapter 映射）；Phase 0.5 先出清单，Phase 7 逐个 + 测试。 |
| 语音路径误接 schema / OutputSchema | Phase 7.5 grep red-line 固化语音 feature_key 零 schema、零接线；lint schema gate 的 chat-only 要求维护豁免清单。 |
| schema 关键字与 Go 校验器子集漂移 | lint 限制 schema 校验关键字只用 `type`/`required`/`properties`/`items`/`enum`，`description` 只作为 annotation；加跨实现一致性测试。 |
| 删除 `bullet_suggestions` alias 影响当前输出 | Phase 11 先证明 prompt/schema/eval 只产 canonical keys，再用 alias-only fail-close 与 canonical success handler tests 锁定边界。 |
| prompt body 重写引入 `template_hash` drift | 每个改动 `.md` 重算 hash 写回 yaml；`loader` + `prompt_lint` 双校验；seed migration hash 一致性（若涉及）一并复核。 |
| 直接型 caller 误用 `OutputSchemaVersion` 当 schema | 明确区分：`OutputSchemaVersion`（字符串 provenance 标签，保留）vs `OutputSchema`（实际 schema，新接）；测试断言两者并存且语义不同。 |

## 7 关联文档导航

> - [Prompt Rubric Registry Spec](../../spec.md)
> - [AI Provider and Model Routing Spec](../../../ai-provider-and-model-routing/spec.md)
> - [Observability Stack Spec](../../../observability-stack/spec.md)
> - [Backend Practice Spec](../../../backend-practice/spec.md)
> - [Backend Targetjob Spec](../../../backend-targetjob/spec.md)
> - [Shared Conventions Codified Spec](../../../shared-conventions-codified/spec.md)
> - [001-baseline plan](../001-baseline/plan.md)

## 8 Handoff / Evidence Log

### 8.1 Phase evidence log

| Phase | Evidence slot | Required command / artifact |
|---|---|---|
| Phase 0 | 基线 snapshot | 9 feature_key 输出↔struct 对照、`bullet_suggestions` 映射判定、校验器 enum 缺口、加载/接线现状、caller 清单 |
| Phase 1 | README + renderer contract | `make lint-prompts`（现有 hash gate 保持绿）+ README 记录 schema annotation / rendered block 规则 |
| Phase 2 | 9 schema 文件 | `find config/prompts -name 'v0.1.0.schema.json' \| wc -l` = 9 + schema descriptions present |
| Phase 3 | schema lint + renderer gate | `make lint-prompts` + negative fixtures 红→绿（含 prompt contract drift） |
| Phase 4 | prompt body contract blocks | `make lint-prompts`（hash 一致 + rendered block 一致 + example schema-valid） |
| Phase 5 | registry 接线 | `go test ./backend/internal/ai/registry/... -race`（OutputSchema 非空 / 语言无关单份） |
| Phase 6 | 校验器 enum | `go test ./backend/internal/ai/aiclient/... -race`（enum fail-close） |
| Phase 7 | caller 透传 + 收口 | `go test ./backend/internal/{targetjob,resume/jobs,review,practice}/... -race` + grep red-line + `make lint` + validate_context + sync-doc-index |
| Phase 13 | downstream HTTP fixture | relevant `cmd/api` Go tests; canonical success + alias-only negative coverage |

### 8.2 Phase 0 handoff snapshot

实施日期：2026-05-23。当前快照基于 `config/prompts/*/v0.1.0*.md`、`backend/internal/**` 调用端与 parser、`backend/internal/ai/registry/{loader,resolver,types}.go`、`backend/internal/ai/aiclient/observability/decorator.go` 读取结果。

#### 8.2.1 9 feature_key 输出契约与 consumer 对照

| feature_key | 当前 prompt 输出 key | 当前 consumer / parser 事实 | 002 schema 决策 |
|---|---|---|---|
| `target.import.parse` | `title`, `companyName`, `coreThemes`, `interviewHypotheses`, `strengths`, `gaps`, `riskSignals`, `requirements[].{kind,label,description,evidenceLevel}` | `parseAIResponse` / `parseAIResponseReq` 强类型消费同名 json tag；`title`、`companyName`、`kind`、`evidenceLevel` 是 parse success contract | schema 与当前 prompt / struct 同名对齐；`title`、`companyName` 必填并持久化到 TargetJob；`description`、`evidenceLevel` 可选但在 schema 中声明 |
| `practice.session.first_question` | `question`, `intent`, `focus_dimension`, `expected_signals`, `time_budget_seconds` | Phase 0 发现 alias fallback；Phase 12 后 `parseFirstQuestion` 只消费 `questionText` / `questionIntent` | prompt/schema/runtime 使用 canonical `questionText` / `questionIntent`；out-of-scope alias 只作为 negative-test 输入 |
| `practice.session.follow_up` | `follow_up_question`, `intent`, `branch_dimension`, `confidence` | 复用 canonical-only `parseFirstQuestion`；无效输出走既有 fallback follow-up action | prompt/schema/runtime 使用 `questionText` / `questionIntent`；可选 `branchDimension` / `confidence` 仅作 schema annotation |
| `practice.turn.lightweight_observe` | `cue`, `severity`, `dimension_hint` | Phase 0 发现 `hint` / `answer_summary` fallback；Phase 12 后 `hintAIResponse` 只消费 `cue` / `answerSummary` | schema required 使用 `cue`；`answerSummary` 保持 optional，缺失时走既有 degrade summary；`severity` / `dimensionHint` 仅作 optional annotation |
| `report.generate` | `summary`, `dimension_scores[]`, `highlights[]`, `issues[]`, `next_actions[]`, `retry_focus_turn_ids[]` | `ReportContentDraft` 强类型消费 snake_case json tag，normalize 后拒绝空报告 | schema 与 `ReportContentDraft` 同名对齐；嵌套 evidence / score / next action 使用当前 struct tag |
| `report.question_assessment` | `dimension_results`, `overall_status`, `confidence`, `strengths`, `gaps`, `recommended_framework`, `review_status` | `QuestionAssessmentDraft` 强类型消费同名 json tag；`dimension_results.*.status` / `overall_status` 是 shared status 值 | schema 与 struct 同名对齐；status enum 使用 `needs_work` / `meets_bar` / `strong` |
| `resume.parse` | `basics`, `experiences`, `projects`, `education`, `skills`, `languages` | `decodeResumeParseResponse` 只验证 6 个顶层 key 存在，保存 `json.RawMessage` | schema required 降级为 6 个 parser-required key；嵌套字段只作为 optional 描述，不宣称强类型 |
| `resume.tailor.gap_review` | `alignment_score`, `gaps[].{topic,why,severity}`, `strengths_to_amplify[].{topic,evidence}`, `risks` | Phase 0 发现 root-level alias fallback；Phase 11 后 `decodeTailorAIResponse` 只接受 canonical `matchSummary.{strengths,gaps}` | schema/prompt/runtime 使用 canonical `matchSummary.{strengths,gaps}` required；out-of-scope key 只作为 negative-test 输入 |
| `resume.tailor.bullet_suggestions` | `suggestions[].{rewrite,why_better,kept_facts}` | Phase 0 发现 suggestion alias fallback；Phase 11 后 `normalizeSuggestions` 只接受 `originalBullet` / `suggestedBullet` / `reason` | schema/prompt/runtime 使用 `suggestions[].{originalBullet,suggestedBullet,reason}`；out-of-scope alias 只作为 negative-test 输入 |

> D-16 / D-22 后，Debrief 与 JD Match feature_key 为 out-of-scope；本表只列当前 `config/prompts` / `config/rubrics` truth source 中存在的 9 个 feature_key。

#### 8.2.2 `resume.tailor.bullet_suggestions` 判定

- 结论：这是需要修复的新契约 drift，而不是继续写入 prompt contract 的兼容形态。
- canonical 新契约：`suggestions[].originalBullet`、`suggestions[].suggestedBullet`、`suggestions[].reason`。
- Runtime 合同：parser 只接受 canonical keys；alias-only 输出返回 `AI_OUTPUT_INVALID`，冲突 alias 不覆盖 canonical 值。
- 删除出新契约：`kept_facts` 当前未被 parser/store 消费，也没有独立 persisted 字段；不进入 schema/rendered prompt block。

#### 8.2.3 `validateOutputSchema` enum 扩展点

- 当前 `outputSchema` 只解析 `type` / `required` / `properties` / `items`。
- 当前 `validateAgainstSchema` 已支持 object required、nested properties、array items 与基础 type 检查，但不会读取 `enum`。
- Phase 6 变更点：在 `outputSchema` 增加 `Enum []any json:"enum"`，在 type check 后、nested traversal 前执行 `enum` 成员校验；`description` 是 annotation，不进入运行时结构。

#### 8.2.4 loader / resolver 接线现状

- `loadPrompts` 只遍历 `.yaml`，`readPrompt` 只加载同名 `.md` 并计算 per-language `template_hash`。
- `snapshot.prompts` 当前是 `featureKey -> language -> promptEntry`；`promptEntry` 没有 schema 字段。
- `ResolveActive` 当前构造 `PromptResolution{OutputSchema: nil}`；`types.go` 已保留 `OutputSchema *json.RawMessage` 字段。
- Phase 5 设计可行：按 `<feature_key>/<version>.schema.json` 加载语言无关 schema 到 prompt entry 或独立 `(featureKey, version)` map；schema 不参与 `computeTemplateHash`，fallback language 继续返回同一份 schema。

#### 8.2.5 caller 透传清单

- 本地 resolution 型：`targetjob.PromptResolution` 与 `resume/jobs.PromptResolution` 当前缺 `OutputSchema` 字段，adapter 未映射，`parse_executor.go` / `parse.go` / `tailor.go` 的 `CallMetadata` 未填实际 schema。
- 直接 registry 型：`review` 两处、`practice` chat 三处（first question / follow-up / lightweight observe）当前只填 provenance 字段，未填 `CallMetadata.OutputSchema`。
- 跳过：`practice.voice.stt` / `practice.voice.tts` 不产 JSON content，不落 schema，不填 `OutputSchema`；`voiceFollowUpPayload` 使用 chat `practice.session.follow_up`，应随 chat 路径填 schema。

#### 8.2.6 Phase 2 schema 文件核对

验证命令：

- `find config/prompts -mindepth 2 -name 'v0.1.0.schema.json' | sort | wc -l` → `9`
- `find config/prompts -mindepth 2 -name 'v0.1.0.schema.json' -print0 | xargs -0 -n1 jq empty` → pass
- schema 子集与 description 检查：`schema_count=9` / `schema_subset_and_descriptions=PASS`
- `make lint-prompts` → `prompt_lint: 9 files clean`

| feature_key | schema top-level | schema required / consumer alignment |
|---|---|---|
| `target.import.parse` | object | required 覆盖当前 prompt/struct 顶层 8 key（含 `title` / `companyName`）；`requirements[].kind` / `label` 与 `parseAIResponseReq` 依赖一致，`kind` / `evidenceLevel` enum 写入 |
| `practice.session.first_question` | object | required `questionText` / `questionIntent`，与 canonical-only parser 一致；out-of-scope `question` / `intent` 只用于 negative tests |
| `practice.session.follow_up` | object | required `questionText` / `questionIntent`，修复当前 `follow_up_question` prompt drift 的目标契约 |
| `practice.turn.lightweight_observe` | object | required `cue` 与 canonical-only parser 一致；`answerSummary` optional；`severity` optional enum |
| `report.generate` | object | required 与 `ReportContentDraft` 顶层 json tag 一致，嵌套 evidence / score / next action 字段与 struct tag 一致 |
| `report.question_assessment` | object | required 与 `QuestionAssessmentDraft` 顶层 json tag 一致；status enum 使用 shared status 值 |
| `resume.parse` | object | required 6 个 `decodeResumeParseResponse` parser-required key；嵌套字段保持 optional annotation |
| `resume.tailor.gap_review` | object | required `matchSummary.{strengths,gaps}`，对齐 canonical parser path；out-of-scope `strengths_to_amplify` / root-level `gaps` 不进 schema |
| `resume.tailor.bullet_suggestions` | object | required `suggestions[].{originalBullet,suggestedBullet,reason}`，对齐 canonical parser/store 字段；2.3 继续修 prompt/test |

#### 8.2.7 Phase 2.3 `bullet_suggestions` canonical 修复证据

- 新增 focused test：`TestTailorHandlerBulletSuggestionsCanonicalKeysRoundTrip` 覆盖 `suggestions[].originalBullet/suggestedBullet/reason` 到 `CompleteTailorRunSuccessInput.Suggestions` 的 round-trip。
- Red phase note：新增 test 立即通过，说明 parser 已有 canonical key 兼容；本项实现侧修复集中在 prompt/schema truth source 对齐。
- Prompt 修复：`config/prompts/resume.tailor.bullet_suggestions/v0.1.0.md` 与 `v0.1.0.en.md` 输出字段改为 canonical `originalBullet` / `suggestedBullet` / `reason`；out-of-scope `rewrite` / `why_better` / `kept_facts` 不作为 prompt 输出字段。
- Hash 同步：两个 YAML `template_hash` 与 `migrations/000002_seed_baseline_prompt_rubric_versions.up.sql` seed row 已同步。
- 验证：`go test ./backend/internal/resume/jobs/... -race` → pass；`make lint-prompts` → `prompt_lint: 9 files clean`。

#### 8.2.8 Phase 3/4 schema lint、renderer 与 prompt body 统一证据

- Red phase：新增 schema lint / renderer 后先运行 `python3 -m pytest scripts/lint/prompt_lint_test.py -q`，基线失败于当时全部 prompt body 缺少 schema-rendered output contract block，证明 drift gate 已接入真实 prompt body；当前 D-22 后 truth source 为 18 个 prompt body。
- Negative fixtures：`scripts/lint/prompt_lint_test.py` 覆盖非法 schema keyword、schema `required` 未出现在 prompt contract、schema required path 与 feature contract 不一致、prompt contract block 手工漂移、required 字段缺 `description`；renderer 覆盖 nested array / enum / example schema-valid。
- Prompt body 统一：9 个 chat feature_key 的 multi/en 共 18 个 `.md` 已替换为 `<!-- output-schema-contract:start -->` / `end` 包围的 schema-rendered 输出契约块；18 个 YAML `template_hash` 已重算，seed migration 已同步现有 baseline rows。
- Drift red-line：`rg -n "Return strict JSON with keys|Return a strict JSON array|Return strict JSON only|follow_up_question|why_better|kept_facts|strengths_to_amplify|alignment_score" config/prompts` 无命中，out-of-scope prose / prompt key 在 prompt truth source 中为零。
- `make lint-prompts` 已覆盖 schema 子集、description、feature contract、rendered block、example JSON 与 migration hash gate：输出 `prompt_lint: 9 files clean`。
- 顶层 `make lint` pass。过程中 `lint-secrets-pattern` 暴露脚本实现偏差：本地忽略的 `deploy/dev-stack/.env` 被全树扫描；已将 `scripts/lint/gitleaks.sh` 调整为扫描 git tracked + unignored candidate mirror，符合 Makefile “staged + tracked” 说明；`bash scripts/lint/gitleaks.sh .` 输出 no leaks，随后 `make lint` 全绿。
- 验证命令：
  - `python3 -m pytest scripts/lint/prompt_lint_test.py -q` → `10 passed`
  - `make lint-prompts` → `prompt_lint: 9 files clean`
  - `make lint` → pass

#### 8.2.9 Phase 5 registry loader / resolver `OutputSchema` 接线证据

- Red phase：新增 loader/resolver focused tests 后先运行，编译失败于 `promptEntry.outputSchema` 与 `PromptMeta.OutputSchema` 缺失，证明当前实现尚未加载/暴露 schema。
- Loader 实现：`loadPrompts` 按 `<feature_key>/<version>.schema.json` 读取语言无关 schema，缓存同一 `json.RawMessage` 指针并挂到 multi/en prompt entry；schema 不参与 `computeTemplateHash`。chat prompt 缺 schema 时启动 fail；语音 STT/TTS feature_key 保留豁免清单。
- Resolver 实现：`ResolveActive` 透出 `PromptResolution.OutputSchema`；language fallback 返回同一份 schema；`GetPrompt` 的 `PromptMeta.OutputSchema` 用于 debug/backfill 场景暴露同一份 schema。
- 附带 preflight 修复：`backend_review_preflight_test.go` 的 prompt-rubric-registry spec Header 断言从 v2.4 同步到当时 v2.5；本次 v2.6 remediation 同步刷新该 Header 断言。
- 验证命令：
  - `go test ./backend/internal/ai/registry -run 'TestLoad(OutputSchemaLanguageIndependent|MissingOutputSchemaRejected)' -count=1` → pass
  - `go test ./backend/internal/ai/registry -run 'TestResolveActiveReturnsOutputSchema|TestGetPromptExact' -count=1` → pass
  - `go test ./backend/internal/ai/registry/... -race` → pass

#### 8.2.10 Phase 6 A3 `validateOutputSchema` enum 扩展证据

- Red phase：新增 `TestDecorator_OutputSchemaEnumMismatchEmitsAIOutputInvalid` 后先运行 focused test，失败于 enum mismatch 未返回错误，证明 A3 校验器未消费 schema `enum`。
- 实现：`outputSchema` 增加 `Enum []any`；`validateAgainstSchema` 在 type check 后执行 enum 成员校验；`description` 未进入 runtime struct/logic。保留 `json.Number` 与 schema numeric enum 的可比较性。
- 覆盖：enum 违反 → `AI_OUTPUT_INVALID` + `ValidationStatusInvalid` + validation failure counter；缺 required 既有 negative test 仍通过；合法 array 顶层 + item enum 输出通过。
- 验证命令：
  - `go test ./backend/internal/ai/aiclient/observability -run 'TestDecorator_OutputSchema(EnumMismatchEmitsAIOutputInvalid|ArrayEnumValidPasses|RequiredFieldMismatchEmitsAIOutputInvalid)' -count=1` → pass
  - `go test ./backend/internal/ai/aiclient/... -race` → pass

#### 8.2.11 Phase 7 caller `OutputSchema` 透传证据

- Red phase：为 targetjob / resume/jobs 本地 `PromptResolution` 与 direct caller metadata 增加断言后先运行 focused tests：targetjob/resume 编译失败于本地 resolution 缺 `OutputSchema` 字段；review/practice 失败于 `CallMetadata.OutputSchema` 为空。Out-of-scope caller 不属于当前 red/green surface。
- targetjob：`PromptResolution` 增加 `OutputSchema *json.RawMessage`；`RegistryAdapter` 映射 registry schema；`ParseExecutor` 填 `CallMetadata.OutputSchema`。focused tests 覆盖 adapter 与 executor metadata。
- resume/jobs：本地 `PromptResolution` 增加 `OutputSchema`；`RegistryAdapter` 映射；`ParseHandler` 与 `TailorHandler` 填 metadata。tests 覆盖 parse、gap_review、bullet_suggestions。
- direct caller：`review` report/assessment、`practice` first question / follow-up / hint 均从 `registry.PromptResolution.OutputSchema` 填 `CallMetadata.OutputSchema`；`voice_turn_service.go` 仅给 chat follow-up 填 schema，STT/TTS metadata 保持无 schema。
- out-of-scope feature key guard：Debrief / JD Match 不属于当前 caller surface；lint 和 config 目录负向断言防止其回流。
- 红线：语音 feature_key 未落 schema；业务包未出现 provider 请求字段 `response_format` / `json_schema`。
- 验证命令：
  - `go test ./backend/internal/targetjob/... -race` → pass
  - `go test ./backend/internal/resume/jobs/... -race` → pass
  - `go test ./backend/internal/review/... ./backend/internal/practice/... -race` → pass
  - out-of-scope feature key negative tests → pass
  - `find config/prompts -name '*.schema.json' | grep -E 'voice|stt|tts|dictation'` → 0 lines (`voice_schema_redline=PASS`)
  - `grep -rnE '"response_format"|json_schema' backend/internal` → 0 lines (`provider_request_redline=PASS`)

#### 8.2.12 Phase 7.6 final gate / lifecycle closure

- Final code gates:
  - `make lint-prompts` → `prompt_lint: 9 files clean`
  - `go test ./backend/internal/ai/registry/... ./backend/internal/ai/aiclient/... ./backend/internal/targetjob/... ./backend/internal/resume/jobs/... ./backend/internal/review/... ./backend/internal/practice/... -race` → pass
  - `make lint` → pass
  - `git diff --check` → pass
- Lifecycle:
  - plan / checklist Header `状态` 切为 `completed`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index` → plans INDEX 迁移 002 从 Active 到 Completed，post-fix zero drift
  - `docs/work-journal/2026-05-23.md` 与 `docs/work-journal/INDEX.md` 记录 `feat(prompt-rubric): close output schema contract`

#### 8.2.13 L2 remediation: strict JSON trailing token fail-close

- Finding: `validateOutputSchema` previously decoded only the first JSON document before schema validation. A model output such as `{"answer":"valid"} trailing prose` could satisfy the schema and avoid `AI_OUTPUT_INVALID`, contradicting the schema-rendered prompt contract's strict JSON requirement.
- Fix: after the first `json.Decoder.Decode`, the validator performs a second decode and accepts only `io.EOF`; any second JSON value or non-whitespace trailing token returns validation error and is wrapped by the decorator as `AI_OUTPUT_INVALID`.
- Red/green evidence:
  - Red: `go test ./backend/internal/ai/aiclient/observability -run 'TestDecorator_OutputSchemaRejectsTrailingTokens' -count=1` failed before the implementation because no error was returned.
  - Green: `go test ./backend/internal/ai/aiclient/observability -run 'TestDecorator_OutputSchemaRejectsTrailingTokens' -count=1` → pass
  - Adjacent: `go test ./backend/internal/ai/aiclient/observability -run 'TestDecorator_OutputSchema' -count=1` → pass
  - Focused no-op guard: `go test ./backend/internal/ai/aiclient/observability -list 'TestDecorator_OutputSchema'` lists `TestDecorator_OutputSchemaRejectsTrailingTokens` plus the existing output-schema tests.
  - Regression: `go test ./backend/internal/ai/registry/... ./backend/internal/ai/aiclient/... ./backend/internal/targetjob/... ./backend/internal/resume/jobs/... ./backend/internal/review/... ./backend/internal/practice/... ./backend/cmd/api/... -race` → pass.
  - Repo lint: `make lint` → pass.
  - Bug record: [BUG-0095](../../../../bugs/BUG-0095.md).

#### 8.2.14 L2 remediation: complete prompt example JSON output

- Finding: prompt contract blocks rendered schema-valid examples, but examples were too thin: required-only JSON with generic `string` / `1` filler values. This gave models a weak target and did not explicitly distinguish desired output from JSON Schema / OpenAPI schema.
- Fix:
  - `scripts/lint/prompt_lint.py` now renders complete representative JSON output, including every schema-declared required and optional property in stable schema order.
  - Renderer examples use business-shaped values for common prompt fields and include an explicit instruction: produce a complete JSON value, not JSON Schema or an OpenAPI schema.
  - `resume.parse` schema now declares project and education optional fields so the rendered full example has inspectable structure instead of `{}`.
  - Current 9 chat feature_key × 2 language coordinates were re-rendered; 18 YAML `template_hash` values and existing prompt seed rows in `migrations/000002_seed_baseline_prompt_rubric_versions.up.sql` were refreshed.
- Red/green evidence:
  - Red: `python3 -m pytest scripts/lint/prompt_lint_test.py -q -k 'rendered_example_includes_optional_properties'` failed because `example_for_schema` omitted optional fields.
  - Green: `python3 -m pytest scripts/lint/prompt_lint_test.py -q -k 'rendered_example'` → `2 passed`.
  - Regression: `python3 -m pytest scripts/lint/prompt_lint_test.py -q` → `11 passed`.
  - Prompt gate: `make lint-prompts` → `prompt_lint: 9 files clean`.
  - Filler-value negative search: `rg -n '"string"|: 1,|Example JSON:' config/prompts -g '*.md'` → 0 matches.
  - Focused registry preflight: `go test ./backend/internal/ai/registry -run TestF3ReportGenerateAndAssessmentPreflight -count=1` → pass.
  - Context/index: `validate_context.py --context docs/spec/prompt-rubric-registry/plans/002-output-schema-contract/context.yaml --docs-root docs --target backend` + `sync-doc-index.py --check` → zero drift.
  - Docs gate: `make docs-check` → sync-doc-index zero drift + markdown links OK.
  - Repo lint: `make lint` → pass, including `prompt_lint: 9 files clean`.
  - Whitespace gate: `git diff --check` → pass.
  - Bug record: [BUG-0096](../../../../bugs/BUG-0096.md).

#### 8.2.15 L2 remediation: active seed migration coverage and out-of-scope key net-zero

- Finding: output schema truth source and runtime parser cover the current 9 active chat feature_keys, but the seed gate must be derived from current `config/` truth source rather than hardcoded counts. Out-of-scope feature_keys may appear in older up/down migrations as audit rows, but the final migration chain must exclude them from active prompt/rubric tables.
- Root cause: the original coverage test asserted fixed counts instead of deriving expected rows from current `config/` truth source. `prompt_lint.py` only verified hash drift for seed rows it could see, so a completely missing or extra feature_key could remain false-green.
- Fix:
  - `backend/internal/ai/registry/db_integration_test.go` now derives expected prompt rows from active prompt YAML files and expected rubric rows from rubric YAML files, scans seed migrations plus deletion migrations, and fails on missing/extra/duplicate coordinates or prompt `template_hash` drift.
  - Out-of-scope key migration rows are treated as migration audit rows only; current net state is validated from active config plus delete migrations.
- Red/green evidence:
  - Red: `go test ./backend/internal/ai/registry -run TestSeedMigrationCoversBaselineFeatureKeys -count=1 -v` failed before deriving current coordinates dynamically.
  - Green: `go test ./backend/internal/ai/registry -run TestSeedMigrationCoversBaselineFeatureKeys -count=1 -v` → pass.
  - Registry package: `go test ./backend/internal/ai/registry -count=1` → pass.
  - Static lint: `python3 scripts/lint/prompt_lint.py` → `prompt_lint: 9 files clean`; `python3 scripts/lint/rubric_lint.py` → `rubric_lint: 9 files clean`; `python3 scripts/lint/migrations_lint.py` → `migration lint: ok`.
  - Migration chain: initial `make migrate-check` without `DATABASE_URL` failed as expected; rerun with dev-stack Postgres `DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable make migrate-check` → pass.

#### 8.2.16 Phase 11 canonical-only resume tailor output

- Finding: `resume.tailor.gap_review` and `resume.tailor.bullet_suggestions` prompts, schemas, and eval fixture only produce canonical keys, but `decodeTailorAIResponse` still accepted root-level and suggestion aliases and could refill a missing `originalBullet` from request input.
- Fix: the parser now only consumes `matchSummary` and `suggestions[].{originalBullet,suggestedBullet,reason}`; alias fallback, request-input refill, and the orphan normalization helper were deleted. Alias-only output fails closed with `AI_OUTPUT_INVALID`, while conflicting aliases cannot override canonical values.
- Preflight hardening: `TestF3ReportGenerateAndAssessmentPreflight` now checks the stable D-13 and current 9-key contract instead of an exact spec version number.
- Verification:
  - `go test ./internal/resume/jobs -count=1` → pass.
  - `go test ./internal/ai/registry -count=1` → pass.
  - `go test ./internal/ai/aiclient/observability -run TestDecorator_OutputSchema -count=1` → pass.
  - `make lint-prompts` → `prompt_lint: 9 files clean`; rubric lint → `rubric_lint: 9 files clean`; migration lint → `migration lint: ok`.
  - All four F3 contexts validate with `--target backend`; `make docs-check`, `git diff --check`, and `make lint-core-loop-pruning-surface` pass with `real_residuals (0)`.
  - Parser-region alias search returns zero; `{{original_bullet}}` remains only as the request-side prompt template variable.

#### 8.2.17 Phase 12 canonical-only practice output

- Finding: `practice.session.first_question`, `practice.session.follow_up`, and `practice.turn.lightweight_observe` schema/prompt contracts use canonical camelCase keys, while production parsers still accepted `question` / `intent` / `hint` / `answer_summary` aliases.
- Red evidence: focused parser tests failed on all four alias paths; conflicting `hint` also overrode canonical `cue`.
- Fix: `parseFirstQuestion` now only reads `questionText` / `questionIntent`; `hintAIResponse` only declares `cue` / `answerSummary`; successful fixtures use canonical keys. Alias literals remain only in focused negative tests.
- Stable references: F3 lint docstrings, prompt README, registry Judge comments, and Judge test comments no longer pin historical spec versions; semantic D-9 / F3 references remain.
- Verification: practice and registry package tests pass; 22 prompt/rubric lint tests pass; prompt/rubric lint each report 9 clean files; production parser alias and fixed-version searches return zero. Top-level `make lint`, all four F3 contexts, `make docs-check`, `git diff --check`, and pruning surface pass with `real_residuals (0)`.

#### 8.2.18 Phase 13 downstream scenario fixture gate

- Finding: after Phase 12 removed the `hint` alias, the `cmd/api` deterministic Practice AI success fixture still emitted `hint`, so canonical success tests degraded to `session_wait` and wrote an extra `AI_OUTPUT_INVALID` task run.
- Fix: the success fixture emits canonical `cue` / `answerSummary`; alias-only `hint` remains in the invalid-output negative case.
- Verification: focused Practice/internal AI/cmd-api tests pass；阶段收口统一执行根 `make test`。Canonical parser changes require downstream consumer code regression；代码测试不得进入 `test/scenarios/e2e/`。

### Phase 14: versioned grounded report + practice semantic-focus pairs and gated activation

- **Owner boundary**: this plan is the sole owner of `config/prompts/report.generate/v0.2.0.{yaml,md,schema.json}`, the complete `practice.session.chat/v0.2.0` prompt/schema+rubric pair, registry multi-version parsing, seed/current-migration parity and final activation. It must not create or edit the dimensions, weights, score levels or other immutable content in `config/rubrics/report.generate/v0.2.0.yaml`; that content, the context-aware judge and eval suite are solely owned by F3 `004`. The practice v0.2 rubric is not a new evaluation design: copy v0.1 dimensions/weights/levels unchanged and change only version/status. backend-practice/004 owns runtime semantic-focus construction, not these registry assets.
- Create v0.2 schema with action `maxLength=200` code-point fuse only。Prompt accepts English<=24 whitespace words/zh-CN<=64 Unicode code points；targeted action-label repair instructions use internal18/52 generation margin。Evalkit uses generation budget=4：each output full-validates and each retry recomputes targeted/whole scope。Judge uses independent budget=4 and separates retryable protocol/schema invalid from terminal content rejection。Both aggregate usage/latency and emit redacted attempt_count/retry_count/reason/scope manifest coordinates。Desktop/mobile wrapping and typed-invalid/no-raw UX 由下游 UI owner 在真实浏览器验证。
- Add immutable `practice.session.chat/v0.2.0` as a draft prompt with a closed copy of the current `{messageText}` schema. Replace the positive input contract with structured `"semanticFocus": {{semantic_focus_json}}`; explain empty generic focus versus server-resolved report-local code/label/issues without injecting rubric content or raw report transcript/anchors. Add an inactive v0.2 rubric whose immutable dimensions, weights and levels equal v0.1. Keep all v0.1 files and 000002 byte-immutable.
- Change the disk snapshot from `featureKey -> language -> entry` to `featureKey -> language -> version -> entry` for both prompts and rubrics. Add exact rubric `status: active|inactive` metadata, backfill current v0.1 rubrics as active, require exactly one active prompt and one active rubric per `(feature_key, language)`, and keep `GetPrompt` / `GetRubric` exact-version access for active or inactive coordinates. Duplicate versions, unknown status, zero/multiple active entries, missing active pair or language/version parity drift fail startup.
- Keep both feature keys' v0.1.0 prompt/schema/rubric files and historical 000002 seed rows as rollback coordinates. Versioned prompt body/schema and rubric dimensions are immutable after publication; activation metadata is the only mutable field. Before final activation, both v0.2 pairs are structurally GREEN while v0.1 remains selected.
- Final activation is ordered: report/practice candidate、storage、rubric/eval prerequisites 均通过后，才可激活各环境的 exclusive truth substrate。For dev files, one release change flips exactly 8 status values across report/practice v0.1/v0.2 prompt+rubric files, then loader full-validation publishes one atomic snapshot. Dev rollback restores all 8 and reloads v0.1 for both before re-activation. For DB, create reversible `000019_activate_report_and_practice_prompt_rubric_v020` only through `make migrate-create`; up inserts/validates both v0.2 pairs and activates them while deactivating both v0.1 pairs in one transaction, and down restores both v0.1 pairs atomically. There is no cross-filesystem/DB atomicity claim.
- Extend lint/preflight/migration parity so current coordinates are derived through the latest migration without editing 000002 or weakening missing/extra/hash checks. Verify final `ResolveActive` returns report and practice prompt/rubric v0.2.0, exact getters still retrieve both v0.1.0 coordinates, report provenance is exactly `report-context.v1` while practice remains `registry.v1`, practice candidate contains no positive legacy focus token, and runtime business prompts contain no rubric injection。

### Phase 15: practice interviewer-identity v0.3 pair and activation

- **Owner boundary**: F3 `002` owns `config/prompts/practice.session.chat/v0.3.0.{yaml,md,schema.json}`, template hash, multi-version registry exactness and activation/rollback. F3 `004` exclusively owns the immutable v0.3 rubric dimensions/weights/levels and identity eval content. backend-practice `001` owns the user behavior contract and real-provider acceptance. No HTTP operation, business table or runtime payload shape changes.
- RED prompt/registry tests require a v0.3 system policy that makes persisted TargetJob/round the only hiring-side identity source, classifies Resume companies as candidate history only, rejects assistant-history identity as evidence, and mandates company-neutral wording when the target employer is anonymous/ambiguous. v0.2 remains byte-immutable and exact-readable.
- GREEN creates the v0.3 prompt with the existing closed `{messageText}` schema, consumes F3 `004`'s v0.3 rubric, and switches only the four practice v0.2/v0.3 prompt/rubric status values after structural gates and the verified `PRACTICE_INTERVIEWER_IDENTITY_V030_PASS` owner marker pass. Partial status edits must fail snapshot validation without replacing the active registry pointer.
- Create `000023_activate_practice_interviewer_identity_v030` only through `make migrate-create`. Up inserts and validates the v0.3 prompt/rubric pair, deactivates practice v0.2 and activates v0.3 in one transaction; down restores practice v0.2 atomically. Report v0.2 remains active and untouched. Latest-migration parity must reject missing/extra/hash/content drift.
- Final verification requires exact `GetPrompt/GetRubric` for practice v0.2 and v0.3, `ResolveActive(practice.session.chat)=v0.3.0`, unchanged `registry.v1` provenance, dev activate/rollback/re-activate, PostgreSQL up/down/up, prompt/rubric/migration lint, focused registry/Practice tests, offline eval and root gates.
