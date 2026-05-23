# F3 Output Schema Contract: 语言无关 schema 真理源 + prompt 契约渲染 + Resolve 接线 + 校验闭环

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-23

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [prompt-rubric-registry spec](../../spec.md) v2.5 §3.1 D-13 锁定的 `output_schema` 契约端到端落地，关闭 §6 C-12。完成后：

- `config/prompts/<feature_key>/v0.1.0.schema.json` 覆盖 **13 个 chat feature_key**，每个是**语言无关**的 JSON Schema 子集（校验关键字 `type` / `required` / `properties` / `items` / `enum`，允许 `description` 非校验注解），描述后端实际反序列化契约；`jd_match.recommendation` / `jd_match.search` 顶层为 `array`，其余 11 个为 `object`。
- 13 × 2 个 prompt body 的输出段从 prose 描述统一为 schema 渲染/校验的「输出契约 + example JSON」写法；schema 是唯一字段真理源，prompt body 不手工维护第二份字段清单；改动的 `.md` 重算 `template_hash`。
- `config/prompts/README.md` 固化 output schema 约定、`description` 注解策略与 prompt body 输出契约渲染/校验写法。
- `scripts/lint/prompt_lint.py` 新增 schema gate：① schema 只用允许关键字（含 `description` 注解）；② 每个 chat feature_key 有 schema（语音 feature_key 豁免）；③ schema `required` ⊆ prompt body 声明的输出 key；④ schema 字段与后端反序列化 struct 的 json tag 一致（drift → exit 1）；⑤ prompt body 输出契约块与 schema 可重渲染结果一致，example JSON 通过 schema 校验。
- `backend/internal/ai/registry/` loader 加载语言无关 schema（不混入 per-language `template_hash`），`ResolveActive` 输出非空 `OutputSchema`。
- A3 `backend/internal/ai/aiclient/observability` 的 `validateOutputSchema` 扩展支持 `enum`；模型输出违反 `enum` 或缺 `required` → `AI_OUTPUT_INVALID` fail-close。
- 各 chat feature_key 的 caller 把 `resolution.OutputSchema` 透传进 `CallMetadata.OutputSchema`，运行时 fail-close 全量生效；语音（STT/TTS）路径跳过。
- 修复 `resume.tailor.bullet_suggestions` prompt 输出 key 与 struct json tag 不一致（若 Phase 0 确认为 drift）。

本 plan **不**切真实 Model Profile（推到 003）、**不**实现 LLM Judge（003）、**不**实现灰度（004）、**不**向 provider 下发 `response_format`（归 A3 后续）。schema 只描述业务输出形状，不含 provider / model / endpoint / SDK 私有字符串。

## 2 背景

spec v1.9 / D-12 起即规划 Resolve 可输出 provider-neutral `output_schema`，但 `001-baseline` 明确预留 `OutputSchema` 字段「不消费」（`backend/internal/ai/registry/resolver.go:62` 硬编码 `OutputSchema: nil`），prompt body 仅以 prose 描述输出形状。当前事实：

- 输出形状以自然语言写在每个语言变体 `.md`（prose + 双份重抄），无机器可读 schema，无 drift gate；BUG-0065（`debrief.generate` prompt 旧形状 vs handler 新形状）即此类 drift 的历史案例。002 修复不能把 prose drift 变成 26 份手写字段表 drift；schema 必须成为唯一字段真理源，prompt 只承载由 schema 渲染/校验的可读契约块。
- A3 `aiclient` observability 已有 `validateOutputSchema`（`decorator.go:538`），递归校验 `type` / `required` / `properties` / `items`，但**不支持 `enum`**；且仅当 `CallMetadata.OutputSchema` 非空时触发——目前仅 `cmd/api/jdmatch_runtime.go:572` 透传，其余 caller 只填 `OutputSchemaVersion`（字符串标签，写 `ai_task_runs` provenance），**不填实际 schema** → 校验空转。
- 13 个 chat feature_key 反序列化对齐分三档：6 个完全对齐；`practice.*` 三个 prompt 多声明字段、parser 只取部分（有意）；`resume.parse` / `resume.tailor.gap_review` 停在 `json.RawMessage`；`resume.tailor.bullet_suggestions` prompt key（`rewrite` / `why_better` / `kept_facts`）与 struct json tag（`SuggestedBullet` / `Reason` / `OriginalBullet`）疑似不一致。

D-13 把 `output_schema` 从「可追加」升级为可机器校验的锁定契约。本 plan 在一个 vertical slice 内落地。

## 3 质量门禁分类

- **Plan 类型**: `truth-source + contract + tooling + code-internal（cross-domain wiring）`。落地 `config/` 语言无关 schema 真理源 + prompt body 契约统一 + `scripts/lint/` 静态 gate + registry 加载 / resolver 接线 + A3 校验器 `enum` 扩展 + 5 个 domain caller 的 `OutputSchema` 透传。不引入用户可见 UI、新 HTTP API 行为或新业务流。
- **TDD 策略**: Code plan requires TDD。所有 checklist item 先红后绿：① schema lint 先写 negative fixtures（schema 非法关键字 / `required` 不在 prompt / schema 字段与 struct json tag 不一致 / prompt 输出契约块手工漂移）再实现；② prompt 输出契约 renderer 先写 fixture 断言（同一 schema 对 multi/en 输出同一字段顺序与 schema-valid example）再落地；③ registry loader/resolver 先写「加载 schema + `ResolveActive` 输出非空 OutputSchema + 语言无关单份」断言再改实现；④ `validateOutputSchema` enum 先写「enum 违反 → error」focused test 再加字段；⑤ caller 透传先写「`metadata.OutputSchema` 非空 + 端到端 fail-close」断言再改 wiring；⑥ `bullet_suggestions` key 修复先写 round-trip 断言。每个 phase 退出 gate 都是可执行命令。
- **BDD 策略**: **BDD 不适用**。本 plan 落地内部契约（schema 真理源 + lint + registry 接线 + provider-neutral 校验器），不新增用户可见 UI、新 HTTP API 行为或端到端业务工作流；`validateOutputSchema` fail-close 复用既有 `AI_OUTPUT_INVALID` 错误路径，仅扩展生效范围与 `enum`，无新端到端用户行为。后续 P0 用户行为流由各 C 域 plan 维护 BDD/E2E gate。
- **替代验证 gate**: ① `make lint-prompts`（schema 子集 + prompt 契约块可重渲染 + example JSON schema-valid + 三向一致性 drift gate + negative fixtures）；② `go test ./backend/internal/ai/registry/... -race`（loader 加载 schema + `ResolveActive` OutputSchema 非空 + 语言无关单份 + fallback 仍返回同一 schema）；③ `go test ./backend/internal/ai/aiclient/... -race`（`validateOutputSchema` enum/required fail-close）；④ `go test ./backend/internal/{targetjob,resume/jobs,review,practice,debrief,jdmatch}/... -race`（caller `OutputSchema` 透传 + 端到端 fail-close）；⑤ schema↔struct drift negative fixture；⑥ grep red-line：语音 feature_key（STT/TTS）不落 schema、不接 `OutputSchema`、不发 `response_format`；⑦ `python3 .agent-skills/implement/shared/scripts/validate_context.py` + `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`；⑧ 顶层 `make lint`。

### 3.1 Cross-layer Operation Matrix（13 chat feature_key）

本 plan 不新增 HTTP operation，[development.md §2.1](../../../../development.md) 标准字段状态：`operationId` = N/A（内部 feature_key，非 HTTP）；`fixture` = 本 plan 不新增/变更；`frontend consumer` = 无（不改 UI）；`backend handler` = 下表反序列化 struct 列；`persistence` = 沿用各 feature_key 既有 `ai_task_runs` provenance，本 plan 不新增/变更业务表；`AI dependency` = 各 feature_key §3.1.1 默认 model profile；`scenario coverage` = §3 替代验证 gate（BDD 不适用）。下表聚焦本 plan 真正改变的 cross-layer 维度（prompt ↔ schema ↔ struct ↔ caller）：

Schema canonical key policy：`required` 只覆盖后端实际依赖字段；当前 parser 已兼容的 legacy alias（例如 `question`/`questionText`、`rewrite`/`suggestedBullet`、`id`/`jobMatchId`）只能作为 Phase 0 兼容事实记录，不得自动进入 prompt 新契约。若保留可选 prompt-only 字段，必须说明评估价值，并用 schema `description` + lint 证明不是第二份手写字段表。

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
| `resume.tailor.bullet_suggestions` | object | `.../v0.1.0.schema.json` | `decodedTailorSuggestion` + `normalizeSuggestions` (`resume/jobs/tailor.go:292` / `:348`，当前兼容 `originalBullet`/`original_bullet`、`suggestedBullet`/`suggested_bullet`/`rewrite`、`reason`/`why_better`/`whyBetter`) | 本地 + adapter | – |
| `debrief.generate` | object | `.../v0.1.0.schema.json` | `generateAIResponse` (`debrief/generate_handler.go:285`) | 直接 | `severity` |
| `debrief.suggest_questions` | object | `.../v0.1.0.schema.json` | `suggestQuestionsAIResponse` (`debrief/service.go:415`) | 直接 | `source` |
| `jd_match.recommendation` | **array** | `.../v0.1.0.schema.json` | `[]llmRecommendation` (`jdmatch/generators/recommendation.go:79`) | 直接（`jdmatch_runtime.go:572` 已透传） | `level` |
| `jd_match.search` | **array** | `.../v0.1.0.schema.json` | `SearchAIResult` 取 `jobMatchId` (`jdmatch/handler/search.go:52`) | 直接 | `level` |

> 语音 feature_key（`practice.voice.stt` / `practice.voice.tts` / `practice.dictation.stt` / `debrief.voice.tts`，对应 `doubao_speech` / `minimax_speech`）不产 JSON content，**不在本 plan 范围**，不落 schema、不接 `OutputSchema`。

## 4 实施步骤

### Phase 0: 基线复核与缺口确认（只读 + handoff snapshot）

#### 0.1 13 个 chat feature_key 输出形态 + struct 测绘
复核 §3.1 operation matrix 每行的 prompt body 输出 key 与后端 struct json tag；记入 §8 Phase 0 handoff。

#### 0.2 `resume.tailor.bullet_suggestions` 映射核实
读取 `backend/internal/resume/jobs/tailor.go` 的 `decodeTailorAIResponse` / `normalizeSuggestions`，判定当前 prompt key（`rewrite`/`why_better`/`kept_facts`）与 canonical parser key（`originalBullet`/`suggestedBullet`/`reason`）是需要修复的新契约 drift，还是仅作为 parser alias 兼容保留；结论决定 Phase 2.3 是否修复，记入 handoff。

#### 0.3 校验器能力复核
读取 `backend/internal/ai/aiclient/observability/decorator.go:531-594`：确认 `outputSchema` 类型（`type`/`required`/`properties`/`items`）与 `validateAgainstSchema`，记录 `enum` 缺口与扩展点。

#### 0.4 加载 / 接线现状复核
读取 `loader.go`（`WalkDir` 只认 `.yaml`、`computeTemplateHash` per-language 5 字段）、`resolver.go:62`（`OutputSchema: nil`）、`types.go:15`（`PromptResolution.OutputSchema *json.RawMessage`）；确认 schema 不混入 per-language `template_hash` 的设计可行。

#### 0.5 caller 接线清单
列出直接型 caller（`debrief`×2 / `review`×2 / `practice`-chat×3 / `jdmatch`）与本地-resolution 型 caller（`targetjob` / `resume/jobs`×2）；区分既有 `OutputSchemaVersion`（标签）与待接的 `OutputSchema`（实际 schema）；标注 `practice/voice_turn_service.go` STT/TTS 跳过点。记入 handoff。

### Phase 1: README 契约 + prompt 输出契约渲染规则

#### 1.1 `config/prompts/README.md` 契约扩写
新增 output schema 约定：文件落点 `<feature_key>/<version>.schema.json`、**语言无关**（每个 `(feature_key, version)` 一份）、JSON Schema 校验子集关键字、`description` 非校验注解、不混入 `template_hash`、与 prompt body 输出 key / 后端 struct json tag 的一致性要求。

#### 1.2 prompt body 输出契约块规范
在 README 固化 prompt body 输出段只能是 schema 渲染/校验的契约块：字段顺序来自 schema `required` + `properties`，字段说明来自 `description`，example JSON 由 schema 生成并可解析；人手只编辑 schema 与业务 prompt 上下文，不在 26 个 `.md` 中独立维护第二份字段表。

#### 1.3 alias / optional 字段记录规则
在 README / handoff 规则中明确：parser 为兼容历史 prompt 保留的 alias 不等于新契约字段；prompt-only 可选字段必须有 `description` 说明评估价值，并由 lint 证明与 schema/rendered block 一致。

### Phase 2: 13 份语言无关 output schema 文件

#### 2.1 落地 13 份 `v0.1.0.schema.json`
按 §3.1 为每个 chat feature_key 写语言无关 schema：顶层 `type`（`jd_match.*` 为 `array` + `items`，其余 `object`）；`required` = 后端实际依赖字段；`enum` 覆盖矩阵列出的受限取值；`description` 覆盖 required 字段、重要可选字段与 enum 语义；只用允许关键字。

#### 2.2 schema ↔ struct canonical key 对齐
人工核对每份 schema 的 `required`/字段名与后端 struct json tag / parser 实际依赖字段一致（Phase 3 lint 强制）。对停在 `json.RawMessage` 的 feature_key（`resume.parse`、`resume.tailor.gap_review`，无强类型 struct json tag），三向对齐降级为「schema ↔ parser required key 一致」，并在 §8 handoff 标注；本 plan 不为这两个 feature_key 引入强类型 struct（留各 C 域 plan）。`jd_match.*` 顶层为 `array`，对齐作用于 `items` schema。

#### 2.3 `resume.tailor.bullet_suggestions` canonical key 修复（条件项）
若 0.2 确认当前 prompt 仍要求 parser 不再需要或评估价值不足的 legacy 字段：统一新 schema/prompt canonical key（默认使用 `originalBullet` / `suggestedBullet` / `reason`，保留 parser alias 兼容但不写入新 prompt contract），并补 round-trip 断言。

### Phase 3: schema lint + prompt 契约 renderer（先红后绿）

#### 3.1 `prompt_lint.py` schema gate + 单测
扩展 `scripts/lint/prompt_lint.py`（或新增 `schema_lint.py` 复用同一 README 描述）：① schema 只含 `type`/`required`/`properties`/`items`/`enum` + `description`；② 每个 chat feature_key 有 schema 文件（语音 feature_key 豁免，维护豁免清单）；③ schema `required` ⊆ 对应 prompt body 声明输出 key；④ schema 字段 ↔ 后端 struct json tag / parser required key 一致。array 顶层（`jd_match.*`）时 ③④ 作用于 `items` schema；停在 `json.RawMessage` 的 `resume.parse` / `resume.tailor.gap_review` 因无 struct json tag，④ 降级为 schema ↔ parser required key 一致。negative fixtures ≥4（非法关键字 / required 不在 prompt / 字段与 struct 不一致 / prompt contract block 手工漂移），先红后绿。

#### 3.2 prompt output contract renderer
新增可测试 renderer（可内嵌于 `prompt_lint.py` 或独立 helper）：从 schema 生成稳定的 prompt 输出契约块和 minimal example JSON；同一 feature_key 的 multi/en body 必须渲染出同一字段结构，仅语言说明文本可不同。单测覆盖 object 顶层、array 顶层、enum、nested object/array、`description` 缺失报错、example JSON 可被 schema gate 接受。

#### 3.3 Makefile 接线
`make lint-prompts` 覆盖 schema gate + rendered prompt contract gate；顶层 `make lint` 联动。

### Phase 4: 13 × 2 prompt body 输出段统一（由 schema 渲染/校验）

#### 4.1 生成/更新 prompt 输出契约块
把每个 `.md` 的 “Return strict JSON …” prose 段替换为 schema 渲染/校验的「输出契约 + example JSON」块，与 Phase 2 schema 对齐；只改输出契约段，不改角色/任务/业务规则措辞。每个改动的 `.md` 重算 `template_hash` 写回同名 `.yaml`。

#### 4.2 prompt body drift gate
`make lint-prompts` 证明 schema、prompt contract block、example JSON、YAML hash 与 seed migration hash（若涉及）全部一致；故意改一个 prompt contract 字段名或 example key 必须失败。

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

### Phase 7: caller 端到端透传 + 收口

#### 7.1 `targetjob` 透传
`targetjob.PromptResolution` 加 `OutputSchema` 字段；`RegistryAdapter.Resolve` 映射；`parse_executor.go` 填 `CallMetadata.OutputSchema`。cross-layer test 断言透传 + fail-close。

#### 7.2 `resume/jobs` 透传
`resume/jobs.PromptResolution` 加 `OutputSchema` + adapter 映射；`parse.go` / `tailor.go` 填 metadata；测试。

#### 7.3 `debrief` / `review` / `practice`(chat) 透传
直接读 `resolution.OutputSchema` 填 `CallMetadata.OutputSchema`（`jdmatch_runtime.go` 范式）；`practice/voice_turn_service.go` STT/TTS 路径跳过，仅 chat（follow_up）接线；测试。

#### 7.4 `jdmatch` 验证
确认 `jdmatch_runtime.go:572` 既有透传在 schema 接通后仍有效；array 顶层 schema 校验通过。

#### 7.5 grep red-line
`grep` 确认语音 feature_key 不落 schema、不接 `OutputSchema`；业务包不出现 `response_format` / `json_schema` 请求字段（归 A3）。

#### 7.6 收口
`make lint-prompts` + `go test ./backend/internal/ai/registry/... ./backend/internal/ai/aiclient/... ./backend/internal/{targetjob,resume/jobs,review,practice,debrief,jdmatch}/... -race` + `validate_context.py` + `sync-doc-index --check`；§8 handoff 列 C-12 命令证据；plan/checklist Header 切 `completed`，同步 INDEX 与工作日志。

## 5 验收标准

- spec §6 C-12 通过 §8 handoff 命令证据验证；§3 替代 gate 全部通过。
- 13 个 chat feature_key 各有 1 份语言无关 schema；prompt body 输出契约块可由 schema 重渲染，example JSON 可解析且 schema-valid；`ResolveActive` 输出非空 `OutputSchema`；故意制造 schema↔prompt↔struct 不一致 → `make lint-prompts` 失败。
- `validateOutputSchema` 对违反 `enum`/缺 `required` 的模型输出 fail-close（`AI_OUTPUT_INVALID`）。
- 语音 feature_key 零 schema、零 `OutputSchema` 接线（grep red-line 0 命中）。
- `docs/spec/prompt-rubric-registry/plans/INDEX.md` 写入 002 行；`history.md` v2.5 entry；`spec.md` Header v2.5 与 `docs/spec/INDEX.md` 一致。
- BDD 不适用声明固化在本 plan §3 + checklist BDD-Gate 区段，并附替代 gate 命令。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| schema 语言无关 vs `template_hash` per-language 语义错配 | schema **不**混入 per-language `template_hash`；用独立三向一致性 lint gate（schema↔rendered prompt contract↔struct）保证不漂移；loader 加载一份共享给所有 language。 |
| prompt body 可读性与重复维护冲突 | schema 加 `description`，prompt 输出段由 renderer 生成/校验；人只维护 schema 一处，`make lint-prompts` 拒绝手工漂移。 |
| caller 透传面广（5 domain） | 每处改动小（直接型填一行；本地-resolution 型加字段 + adapter 映射）；Phase 0.5 先出清单，Phase 7 逐个 + 测试。 |
| 语音路径误接 schema / OutputSchema | Phase 7.5 grep red-line 固化语音 feature_key 零 schema、零接线；lint schema gate 的 chat-only 要求维护豁免清单。 |
| schema 关键字与 Go 校验器子集漂移 | lint 限制 schema 校验关键字只用 `type`/`required`/`properties`/`items`/`enum`，`description` 只作为 annotation；加跨实现一致性测试。 |
| `bullet_suggestions` 修复影响既有行为 | Phase 0.2 先判定 canonical key 与 alias 兼容事实；修复走 round-trip 断言护栏，最小改动对齐。 |
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
| Phase 0 | 基线 snapshot | 13 feature_key 输出↔struct 对照、`bullet_suggestions` 映射判定、校验器 enum 缺口、加载/接线现状、caller 清单 |
| Phase 1 | README + renderer contract | `make lint-prompts`（现有 hash gate 保持绿）+ README 记录 schema annotation / rendered block 规则 |
| Phase 2 | 13 schema 文件 | `find config/prompts -name 'v0.1.0.schema.json' \| wc -l` = 13 + schema descriptions present |
| Phase 3 | schema lint + renderer gate | `make lint-prompts` + negative fixtures 红→绿（含 prompt contract drift） |
| Phase 4 | prompt body contract blocks | `make lint-prompts`（hash 一致 + rendered block 一致 + example schema-valid） |
| Phase 5 | registry 接线 | `go test ./backend/internal/ai/registry/... -race`（OutputSchema 非空 / 语言无关单份） |
| Phase 6 | 校验器 enum | `go test ./backend/internal/ai/aiclient/... -race`（enum fail-close） |
| Phase 7 | caller 透传 + 收口 | `go test ./backend/internal/{targetjob,resume/jobs,review,practice,debrief,jdmatch}/... -race` + grep red-line + `make lint` + validate_context + sync-doc-index |

### 8.2 Phase 0 handoff snapshot

实施日期：2026-05-23。当前快照基于 `config/prompts/*/v0.1.0*.md`、`backend/internal/**` 调用端与 parser、`backend/internal/ai/registry/{loader,resolver,types}.go`、`backend/internal/ai/aiclient/observability/decorator.go` 读取结果。

#### 8.2.1 13 feature_key 输出契约与 consumer 对照

| feature_key | 当前 prompt 输出 key | 当前 consumer / parser 事实 | 002 schema 决策 |
|---|---|---|---|
| `target.import.parse` | `coreThemes`, `interviewHypotheses`, `strengths`, `gaps`, `riskSignals`, `requirements[].{kind,label,description,evidenceLevel}` | `parseAIResponse` / `parseAIResponseReq` 强类型消费同名 json tag；`kind`、`evidenceLevel` 是受限枚举 | schema 与当前 prompt / struct 同名对齐；`description`、`evidenceLevel` 可选但在 schema 中声明 |
| `practice.session.first_question` | `question`, `intent`, `focus_dimension`, `expected_signals`, `time_budget_seconds` | `parseFirstQuestion` 实际消费 `questionText` / `questionIntent`，兼容 `question` / `intent` alias；其他字段不消费 | 新 prompt/schema 使用 canonical `questionText` / `questionIntent`；旧 `question` / `intent` 只保留 parser alias，不进入新契约；其余 prompt-only 字段仅在有评估价值时作为 optional + `description` |
| `practice.session.follow_up` | `follow_up_question`, `intent`, `branch_dimension`, `confidence` | 复用 `parseFirstQuestion`，当前不消费 `follow_up_question`，失败时走 fallback follow-up action | 当前 prompt key 与 parser drift；新 prompt/schema 使用 `questionText` / `questionIntent`，可选 `branchDimension` / `confidence` 需有 `description` |
| `practice.turn.lightweight_observe` | `cue`, `severity`, `dimension_hint` | `hintAIResponse` 消费 `hint` / `cue`，未消费 `severity` / `dimension_hint`；空 cue/hint 触发 degrade | schema required 使用 `cue`；`severity` 保留 optional enum `info` / `nudge` / `alert` 作为 UI/评估信号；`dimensionHint` 若保留则只作 optional annotation |
| `report.generate` | `summary`, `dimension_scores[]`, `highlights[]`, `issues[]`, `next_actions[]`, `retry_focus_turn_ids[]` | `ReportContentDraft` 强类型消费 snake_case json tag，normalize 后拒绝空报告 | schema 与 `ReportContentDraft` 同名对齐；嵌套 evidence / score / next action 使用当前 struct tag |
| `report.question_assessment` | `dimension_results`, `overall_status`, `confidence`, `strengths`, `gaps`, `recommended_framework`, `review_status` | `QuestionAssessmentDraft` 强类型消费同名 json tag；`dimension_results.*.status` / `overall_status` 是 shared status 值 | schema 与 struct 同名对齐；status enum 使用 `needs_work` / `meets_bar` / `strong` |
| `resume.parse` | `basics`, `experiences`, `projects`, `education`, `skills`, `languages` | `decodeResumeParseResponse` 只验证 6 个顶层 key 存在，保存 `json.RawMessage` | schema required 降级为 6 个 parser-required key；嵌套字段只作为 optional 描述，不宣称强类型 |
| `resume.tailor.gap_review` | `alignment_score`, `gaps[].{topic,why,severity}`, `strengths_to_amplify[].{topic,evidence}`, `risks` | `decodeTailorAIResponse` 优先消费 canonical `matchSummary.{strengths,gaps}`；否则把 legacy `strengths_to_amplify` / `gaps` 归一为 `matchSummary`；`alignment_score` / `risks` 不消费 | 新 schema/prompt 使用 canonical `matchSummary.{strengths,gaps}` required；legacy key 只作 parser compatibility；若保留 severity，仅作为 optional legacy annotation，不作为 required |
| `resume.tailor.bullet_suggestions` | `suggestions[].{rewrite,why_better,kept_facts}` | `normalizeSuggestions` canonical 消费 `originalBullet` / `suggestedBullet` / `reason`，兼容 `original_bullet`、`rewrite`、`why_better`；`kept_facts` 不消费 | 判定为新契约 drift；新 schema/prompt 使用 `suggestions[].{originalBullet,suggestedBullet,reason}`，旧 alias 保留在 parser，`kept_facts` 不进入新契约 |
| `debrief.generate` | `questions[].{questionText,myAnswerSummary,interviewerReaction,aiAnalysis}`, `riskItems[].{label,severity}` | `parseGenerateResponse` 消费 debrief question/risk item shape；severity 受限于 debrief risk 语义 | schema 与 current parser/output shape 对齐；`riskItems[].severity` 使用 `low` / `medium` / `high` |
| `debrief.suggest_questions` | `suggestions[].{questionText,whyLikelyAsked,source,stage}` | `suggestQuestionsAIResponse` / `suggestedQuestionAI` 强类型消费同名 tag；`source` 经 `validDebriefQuestionSource` 校验 | schema 与 struct 同名对齐；`source` enum 使用 `jd` / `resume` / `mock_report` / `manual` |
| `jd_match.recommendation` | 顶层 array；items 包含 `jobMatchId`, `title`, `company`, `companyTag`, `level`, `location`, `comp`, `posted`, `score`, `fit`, `reasons`, `risks`, `highlights`, `sourceUrl`, `sourceLabel`, `networkNote`, `similarInterviewers`, `interviewHypotheses` | `[]llmRecommendation` 强类型消费 items；`level` 是可选枚举；generator 校验 required / score / fit 后 upsert | 顶层 `array`；items required 取后端实际依赖字段；nullable 字段用 optional 字段表达，不引入 `null` type |
| `jd_match.search` | 顶层 array；计划文档沿用 recommendation item contract | `parseJDMatchSearchIDs` 实际只依赖每项 string 或 object 的 `jobMatchId` / `id`；`jdMatchPayload` 已在 resolution 非空时透传 `OutputSchema` | 顶层 `array`；新契约 item canonical 使用 `jobMatchId` required，`id` 仅 alias；其他 recommendation 字段不得作为 search required |

#### 8.2.2 `resume.tailor.bullet_suggestions` 判定

- 结论：这是需要修复的新契约 drift，而不是继续写入 prompt contract 的兼容形态。
- canonical 新契约：`suggestions[].originalBullet`、`suggestions[].suggestedBullet`、`suggestions[].reason`。
- 兼容保留：parser 继续接受 `original_bullet`、`suggested_bullet`、`rewrite`、`why_better`、`whyBetter`，用于历史输出/fixture 回放。
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
- 直接 registry 型：`review` 两处、`practice` chat 三处（first question / follow-up / lightweight observe）、`debrief` 两处当前只填 provenance 字段，未填 `CallMetadata.OutputSchema`。
- 已接线：`backend/cmd/api/jdmatch_runtime.go` 的 `jdMatchPayload` 已在 `resolution.OutputSchema != nil` 时填 `metadata.OutputSchema`；Phase 7 仍需覆盖 search/recommendation array 顶层测试。
- 跳过：`practice.voice.stt` / `practice.voice.tts` / dictation STT / debrief TTS 不产 JSON content，不落 schema，不填 `OutputSchema`；`voiceFollowUpPayload` 使用 chat `practice.session.follow_up`，应随 chat 路径填 schema。

#### 8.2.6 Phase 2 schema 文件核对

验证命令：

- `find config/prompts -mindepth 2 -name 'v0.1.0.schema.json' | sort | wc -l` → `13`
- `find config/prompts -mindepth 2 -name 'v0.1.0.schema.json' -print0 | xargs -0 -n1 jq empty` → pass
- schema 子集与 description 检查：`schema_count=13` / `schema_subset_and_descriptions=PASS`
- `make lint-prompts` → `prompt_lint: 26 files clean`

| feature_key | schema top-level | schema required / consumer alignment |
|---|---|---|
| `target.import.parse` | object | required 覆盖当前 prompt/struct 顶层 6 key；`requirements[].kind` / `label` 与 `parseAIResponseReq` 依赖一致，`kind` / `evidenceLevel` enum 写入 |
| `practice.session.first_question` | object | required `questionText` / `questionIntent`，与 parser 优先消费字段一致；旧 `question` / `intent` 不进 schema |
| `practice.session.follow_up` | object | required `questionText` / `questionIntent`，修复当前 `follow_up_question` prompt drift 的目标契约 |
| `practice.turn.lightweight_observe` | object | required `cue` 与 parser fallback 字段一致；`severity` optional enum |
| `report.generate` | object | required 与 `ReportContentDraft` 顶层 json tag 一致，嵌套 evidence / score / next action 字段与 struct tag 一致 |
| `report.question_assessment` | object | required 与 `QuestionAssessmentDraft` 顶层 json tag 一致；status enum 使用 shared status 值 |
| `resume.parse` | object | required 6 个 `decodeResumeParseResponse` parser-required key；嵌套字段保持 optional annotation |
| `resume.tailor.gap_review` | object | required `matchSummary.{strengths,gaps}`，对齐 canonical parser path；legacy `strengths_to_amplify` / `gaps` 不进新 schema |
| `resume.tailor.bullet_suggestions` | object | required `suggestions[].{originalBullet,suggestedBullet,reason}`，对齐 canonical parser/store 字段；2.3 继续修 prompt/test |
| `debrief.generate` | object | required `questions` / `riskItems`；question item required 与 parser 非空检查一致；risk severity enum 写入 |
| `debrief.suggest_questions` | object | required `suggestions[].{questionText,whyLikelyAsked,source}`，与 parser 非空 + source enum 校验一致 |
| `jd_match.recommendation` | array | items required 覆盖 generator upsert 依赖字段；`level` optional enum |
| `jd_match.search` | array | items required `jobMatchId`，与 search parser canonical ID path 一致；`id` 仅历史 alias |

#### 8.2.7 Phase 2.3 `bullet_suggestions` canonical 修复证据

- 新增 focused test：`TestTailorHandlerBulletSuggestionsCanonicalKeysRoundTrip` 覆盖 `suggestions[].originalBullet/suggestedBullet/reason` 到 `CompleteTailorRunSuccessInput.Suggestions` 的 round-trip。
- Red phase note：新增 test 立即通过，说明 parser 已有 canonical key 兼容；本项实现侧修复集中在 prompt/schema truth source 对齐。
- Prompt 修复：`config/prompts/resume.tailor.bullet_suggestions/v0.1.0.md` 与 `v0.1.0.en.md` 输出字段改为 canonical `originalBullet` / `suggestedBullet` / `reason`；旧 `rewrite` / `why_better` / `kept_facts` 不再作为 prompt 输出字段。
- Hash 同步：两个 YAML `template_hash` 与 `migrations/000002_seed_baseline_prompt_rubric_versions.up.sql` seed row 已同步。
- 验证：`go test ./backend/internal/resume/jobs/... -race` → pass；`make lint-prompts` → `prompt_lint: 26 files clean`。

#### 8.2.8 Phase 3/4 schema lint、renderer 与 prompt body 统一证据

- Red phase：新增 schema lint / renderer 后先运行 `python3 -m pytest scripts/lint/prompt_lint_test.py -q`，基线失败于 26 个 prompt body 缺少 schema-rendered output contract block，证明 drift gate 已接入真实 prompt body。
- Negative fixtures：`scripts/lint/prompt_lint_test.py` 覆盖非法 schema keyword、schema `required` 未出现在 prompt contract、schema required path 与 feature contract 不一致、prompt contract block 手工漂移、required 字段缺 `description`；renderer 覆盖 nested array / enum / example schema-valid。
- Prompt body 统一：13 个 chat feature_key 的 multi/en 共 26 个 `.md` 已替换为 `<!-- output-schema-contract:start -->` / `end` 包围的 schema-rendered 输出契约块；26 个 YAML `template_hash` 已重算，seed migration 已同步现有 baseline rows。
- Drift red-line：`rg -n "Return strict JSON with keys|Return a strict JSON array|Return strict JSON only|follow_up_question|why_better|kept_facts|strengths_to_amplify|alignment_score" config/prompts` 无命中，旧 prose / legacy prompt key 已从 prompt truth source 清除。
- `make lint-prompts` 已覆盖 schema 子集、description、feature contract、rendered block、example JSON 与 migration hash gate：输出 `prompt_lint: 26 files clean`。
- 顶层 `make lint` pass。过程中 `lint-secrets-pattern` 暴露脚本实现偏差：本地忽略的 `deploy/dev-stack/.env` 被全树扫描；已将 `scripts/lint/gitleaks.sh` 调整为扫描 git tracked + unignored candidate mirror，符合 Makefile “staged + tracked” 说明；`bash scripts/lint/gitleaks.sh .` 输出 no leaks，随后 `make lint` 全绿。
- 验证命令：
  - `python3 -m pytest scripts/lint/prompt_lint_test.py -q` → `10 passed`
  - `make lint-prompts` → `prompt_lint: 26 files clean`
  - `make lint` → pass

#### 8.2.9 Phase 5 registry loader / resolver `OutputSchema` 接线证据

- Red phase：新增 loader/resolver focused tests 后先运行，编译失败于 `promptEntry.outputSchema` 与 `PromptMeta.OutputSchema` 缺失，证明当前实现尚未加载/暴露 schema。
- Loader 实现：`loadPrompts` 按 `<feature_key>/<version>.schema.json` 读取语言无关 schema，缓存同一 `json.RawMessage` 指针并挂到 multi/en prompt entry；schema 不参与 `computeTemplateHash`。chat prompt 缺 schema 时启动 fail；语音 STT/TTS feature_key 保留豁免清单。
- Resolver 实现：`ResolveActive` 透出 `PromptResolution.OutputSchema`；language fallback 返回同一份 schema；`GetPrompt` 的 `PromptMeta.OutputSchema` 用于 debug/backfill 场景暴露同一份 schema。
- 附带 preflight 修复：`backend_review_preflight_test.go` 的 prompt-rubric-registry spec Header 断言从 v2.4 同步到当前 v2.5。
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

- Red phase：为 targetjob / resume/jobs 本地 `PromptResolution` 与 direct caller metadata 增加断言后先运行 focused tests：targetjob/resume 编译失败于本地 resolution 缺 `OutputSchema` 字段；debrief/review/practice 失败于 `CallMetadata.OutputSchema` 为空；jdmatch 既有 `jdMatchPayload` 透传已通过。
- targetjob：`PromptResolution` 增加 `OutputSchema *json.RawMessage`；`RegistryAdapter` 映射 registry schema；`ParseExecutor` 填 `CallMetadata.OutputSchema`。focused tests 覆盖 adapter 与 executor metadata。
- resume/jobs：本地 `PromptResolution` 增加 `OutputSchema`；`RegistryAdapter` 映射；`ParseHandler` 与 `TailorHandler` 填 metadata。tests 覆盖 parse、gap_review、bullet_suggestions。
- direct caller：`debrief.generate` / `debrief.suggest_questions`、`review` report/assessment、`practice` first question / follow-up / hint 均从 `registry.PromptResolution.OutputSchema` 填 `CallMetadata.OutputSchema`；`voice_turn_service.go` 仅给 chat follow-up 填 schema，STT/TTS metadata 保持无 schema。
- jdmatch：`jdMatchPayload` 既有 `resolution.OutputSchema` 透传在 search/recommendation adapter test 中断言保持有效，array 顶层 schema 透传。
- 红线：语音 feature_key 未落 schema；业务包未出现 provider 请求字段 `response_format` / `json_schema`。
- 验证命令：
  - `go test ./backend/internal/targetjob/... -race` → pass
  - `go test ./backend/internal/resume/jobs/... -race` → pass
  - `go test ./backend/internal/debrief/... ./backend/internal/review/... ./backend/internal/practice/... -race` → pass
  - `go test ./backend/internal/jdmatch/... ./backend/cmd/api/... -race` → pass
  - `find config/prompts -name '*.schema.json' | grep -E 'voice|stt|tts|dictation'` → 0 lines (`voice_schema_redline=PASS`)
  - `grep -rnE '"response_format"|json_schema' backend/internal` → 0 lines (`provider_request_redline=PASS`)

#### 8.2.12 Phase 7.6 final gate / lifecycle closure

- Final code gates:
  - `make lint-prompts` → `prompt_lint: 26 files clean`
  - `go test ./backend/internal/ai/registry/... ./backend/internal/ai/aiclient/... ./backend/internal/targetjob/... ./backend/internal/resume/jobs/... ./backend/internal/review/... ./backend/internal/practice/... ./backend/internal/debrief/... ./backend/internal/jdmatch/... -race` → pass
  - `make lint` → pass
  - `git diff --check` → pass
- Lifecycle:
  - plan / checklist Header `状态` 切为 `completed`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index` → plans INDEX 迁移 002 从 Active 到 Completed，post-fix zero drift
  - `docs/work-journal/2026-05-23.md` 与 `docs/work-journal/INDEX.md` 记录 `feat(prompt-rubric): close output schema contract`
