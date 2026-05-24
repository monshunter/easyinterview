# F3 Output Schema Contract Checklist

> **版本**: 1.4
> **状态**: completed
> **更新日期**: 2026-05-24

**关联计划**: [plan](./plan.md)

## Phase 0: 基线复核与缺口确认（只读 + handoff snapshot）

- [x] 0.1 复核 plan §3.1 operation matrix 13 行的 prompt body 输出 key 与后端 struct json tag。验证: 13 行对照表写入 plan §8.2 Phase 0 handoff
- [x] 0.2 读取 `backend/internal/resume/jobs/tailor.go` 的 `decodeTailorAIResponse` / `normalizeSuggestions`，判定 `resume.tailor.bullet_suggestions` 当前 prompt key（`rewrite`/`why_better`/`kept_facts`）与 canonical parser key（`originalBullet`/`suggestedBullet`/`reason`）是需要修复的新契约 drift，还是仅作为 parser alias 兼容保留。验证: 判定结论（canonical key / alias 兼容 / drift 修复范围）写入 handoff，决定 2.3 是否执行
- [x] 0.3 读取 `backend/internal/ai/aiclient/observability/decorator.go:531-594`，记录 `outputSchema` 类型与 `validateAgainstSchema` 的 `enum` 缺口与扩展点。验证: 扩展点写入 handoff
- [x] 0.4 读取 `loader.go`（`WalkDir`/`computeTemplateHash`）、`resolver.go:62`、`types.go:15`，确认 schema 不混入 per-language `template_hash` 的加载设计可行。验证: 加载/接线现状写入 handoff
- [x] 0.5 列出直接型 caller（`debrief`×2 / `review`×2 / `practice`-chat×3 / `jdmatch`）与本地-resolution 型 caller（`targetjob` / `resume/jobs`×2），区分 `OutputSchemaVersion`（标签）vs `OutputSchema`（实际 schema），标注 `practice/voice_turn_service.go` STT/TTS 跳过点。验证: caller 清单写入 handoff

## Phase 1: README 契约 + prompt 输出契约渲染规则

- [x] 1.1 `config/prompts/README.md` 新增 output schema 约定（文件落点、语言无关、JSON Schema 校验子集、`description` 非校验注解、不混入 `template_hash`、与 prompt body / struct 一致性）。验证: README 含 output schema 章节；`make lint-prompts` 仍绿
- [x] 1.2 在 README 固化 prompt body 输出契约块规范：字段顺序来自 schema，字段说明来自 `description`，complete example JSON output 由 schema 生成并可解析，覆盖 schema 声明的 required + optional 字段，使用业务形态值，并明确不是 JSON Schema / OpenAPI schema；人不在 26 个 `.md` 手工维护第二份字段表。验证: README 明确 schema 是唯一字段真理源，并说明 renderer/lint 的 drift 行为
- [x] 1.3 在 README / handoff 规则中明确 alias / optional 字段策略：parser legacy alias 不自动进入新 prompt contract；prompt-only 可选字段必须有 `description` 说明评估价值。验证: README 或 plan §8 handoff 模板包含 alias/optional 判定项

## Phase 2: 13 份语言无关 output schema 文件

- [x] 2.1 为 13 个 chat feature_key 各落地 `config/prompts/<feature_key>/v0.1.0.schema.json`（语言无关，顶层 `object`/`array`，`required`=后端依赖字段，含 `enum`，允许 `description` 注解，只用允许关键字）。验证: `find config/prompts -mindepth 2 -name 'v0.1.0.schema.json' | wc -l` = 13；`jq empty` 解析全部通过；required/关键可选字段含 `description`
- [x] 2.2 人工核对每份 schema 的 `required`/字段名与后端 struct json tag / parser required key 一致。验证: 核对表写入 plan §8 handoff（Phase 3 lint 强制）
- [x] 2.3 （条件项，依 0.2 结论）修复 `resume.tailor.bullet_suggestions` schema/prompt canonical key，默认使用 `originalBullet` / `suggestedBullet` / `reason`，保留 parser alias 兼容但不写入新 prompt contract，补 round-trip 断言。验证: `go test ./backend/internal/resume/jobs/... -race` 全绿

## Phase 3: schema lint + prompt 契约 renderer（先红后绿）

- [x] 3.1 扩展 `scripts/lint/prompt_lint.py`（或新增 `schema_lint.py`）+ 单测：① schema 只含 `type`/`required`/`properties`/`items`/`enum` + `description`；② 每个 chat feature_key 有 schema（语音豁免清单）；③ schema `required` ⊆ prompt body 输出 key；④ schema 字段 ↔ 后端 struct json tag / parser required key 一致（array 顶层 `jd_match.*` 作用于 `items`；`resume.parse`/`resume.tailor.gap_review` 因停在 `json.RawMessage`，④ 降级为 schema ↔ parser required key 一致）。验证: `python3 -m pytest scripts/lint/prompt_lint_test.py` 全绿；≥4 negative fixtures（非法关键字 / required 不在 prompt / 字段与 struct 不一致 / prompt contract block 手工漂移）红→绿
- [x] 3.2 新增 prompt output contract renderer（可内嵌于 `prompt_lint.py` 或独立 helper）：从 schema 生成稳定的输出契约块和 complete representative example JSON output，覆盖 object、array、enum、nested object/array、`description` 缺失报错、example JSON schema-valid、optional 字段完整示例与业务形态值。验证: renderer 单测全绿
- [x] 3.3 `make lint-prompts` 覆盖 schema gate + rendered prompt contract gate；顶层 `make lint` 联动。验证: `make lint-prompts && make lint` 全绿

## Phase 4: 13 × 2 prompt body 输出段统一（由 schema 渲染/校验）

- [x] 4.1 把 13 × 2 个 `.md` 的 “Return strict JSON …” prose 段替换为 schema 渲染/校验的「输出契约 + complete example JSON output」块，重算 `template_hash` 写回 `.yaml`。验证: `make lint-prompts`（hash 一致 + rendered block 一致）全绿
- [x] 4.2 prompt body drift gate 负向验证：故意改一个 prompt contract 字段名或 example key，`make lint-prompts` 失败；恢复后通过。验证: negative fixture 或测试覆盖

## Phase 5: registry 加载 + resolver 接线 OutputSchema（先红后绿）

- [x] 5.1 `loader.go` 按 `<feature_key>/<version>.schema.json` 加载语言无关 schema（所有 language 共享一份），不混入 `template_hash`；chat feature_key 缺 schema → 启动 fail，语音豁免。验证: `go test ./backend/internal/ai/registry -run TestLoad -race` 全绿（加载 / 缺失 fail / 语言无关共享）
- [x] 5.2 `resolver.go` `ResolveActive` / `GetPrompt` 用加载的 schema 替换 `OutputSchema: nil`；fallback 仍返回同一份。验证: `go test ./backend/internal/ai/registry -run TestResolve -race` 全绿（OutputSchema 非空 / 各 language 一致）

## Phase 6: A3 validateOutputSchema 扩展 enum（先红后绿）

- [x] 6.1 `decorator.go` `outputSchema` 加 `Enum []any` + `validateAgainstSchema` enum 成员校验分支；`description` 不影响运行时校验。验证: 编译通过；enum 字段被解析
- [x] 6.2 aiclient focused tests：enum 违反 → `AI_OUTPUT_INVALID`；缺 required 仍 fail；合法输出（含 array 顶层）通过。验证: `go test ./backend/internal/ai/aiclient/... -race` 全绿，enum negative test 红→绿
- [x] 6.3 L2 remediation：`validateOutputSchema` 必须拒绝 schema-valid JSON 后追加非空 trailing token / prose 的模型输出，返回 `AI_OUTPUT_INVALID`。验证: focused negative test 红→绿；`go test ./backend/internal/ai/aiclient/observability -run 'TestDecorator_OutputSchemaRejectsTrailingTokens' -count=1` pass；`go test ./backend/internal/ai/aiclient/observability -run 'TestDecorator_OutputSchema' -count=1` pass

## Phase 7: caller 端到端透传 + 收口

- [x] 7.1 `targetjob.PromptResolution` 加 `OutputSchema` 字段 + `RegistryAdapter` 映射 + `parse_executor.go` 填 `CallMetadata.OutputSchema`。验证: `go test ./backend/internal/targetjob/... -race` 全绿，cross-layer test 断言透传 + fail-close
- [x] 7.2 `resume/jobs.PromptResolution` 加 `OutputSchema` + adapter 映射 + `parse.go`/`tailor.go` 填 metadata。验证: `go test ./backend/internal/resume/jobs/... -race` 全绿
- [x] 7.3 `debrief`/`review`/`practice`(chat) 直接读 `resolution.OutputSchema` 填 metadata；`practice/voice_turn_service.go` STT/TTS 跳过，仅 chat 接线。验证: `go test ./backend/internal/{debrief,review,practice}/... -race` 全绿
- [x] 7.4 验证 `jdmatch_runtime.go:572` 既有透传在 schema 接通后仍有效（array 顶层校验通过）。验证: `go test ./backend/internal/jdmatch/... ./backend/cmd/api/... -race` 全绿
- [x] 7.5 grep red-line：语音 feature_key 不落 schema、不接 `OutputSchema`；业务包无 `response_format`/`json_schema` 请求字段。验证: `find config/prompts -name '*.schema.json' | grep -E 'voice|stt|tts|dictation'` 返回 0 行（语音 feature_key 无 schema）；`! grep -rnE '"response_format"|json_schema' backend/internal` provider 请求字段命中为 0
- [x] 7.6 收口：`make lint-prompts` + `go test ./backend/internal/ai/registry/... ./backend/internal/ai/aiclient/... ./backend/internal/{targetjob,resume/jobs,review,practice,debrief,jdmatch}/... -race` + `validate_context.py` + `sync-doc-index --check`；§8 handoff 列 C-12 证据；Header 切 `completed` 同步 INDEX 与工作日志。验证: 全部命令绿；INDEX 显示 completed

## Phase 8: L2 prompt example remediation（完整 JSON output）

- [x] 8.1 为 renderer 增加回归测试：example JSON 必须包含 schema-declared optional properties，且不能使用 `string` / `1` 这类最小占位。验证: red `python3 -m pytest scripts/lint/prompt_lint_test.py -q -k 'rendered_example_includes_optional_properties'` 失败于 optional 字段缺失；green `python3 -m pytest scripts/lint/prompt_lint_test.py -q -k 'rendered_example'` → `2 passed`
- [x] 8.2 重渲染 13 × 2 prompt body，刷新 26 个 YAML `template_hash`，同步 `migrations/000002_seed_baseline_prompt_rubric_versions.up.sql` 中已有 prompt seed row 的 body/hash，并补充 `resume.parse` project/education optional schema 字段。验证: `make lint-prompts` → `prompt_lint: 26 files clean`；`rg -n '"string"|: 1,|Example JSON:' config/prompts -g '*.md'` → 0 matches
- [x] 8.3 同步 spec/plan/checklist/context/index 与收口验证。验证: `validate_context.py` + `sync-doc-index --check` + `git diff --check` 通过；Header 恢复 `completed`

## Phase 9: L2 seed migration coverage remediation（active truth source 全量覆盖）

- [x] 9.1 重写 seed migration 静态覆盖测试，从 `config/prompts` active YAML 与 `config/rubrics` YAML 反推出期望坐标，扫描所有 `migrations/*seed_baseline_prompt_rubric*.up.sql`，拒绝 missing / extra / duplicate row 与 prompt `template_hash` drift。验证: red `go test ./backend/internal/ai/registry -run TestSeedMigrationCoversBaselineFeatureKeys -count=1 -v` 失败于缺失 `jd_match.recommendation` / `jd_match.search` en/multi rows
- [x] 9.2 新增 `migrations/000010_seed_baseline_prompt_rubric_versions_jd_match.{up,down}.sql`，补齐 `jd_match.recommendation` / `jd_match.search` × `en` / `multi` 的 prompt_versions 与 rubric_versions seed rows。验证: green `go test ./backend/internal/ai/registry -run TestSeedMigrationCoversBaselineFeatureKeys -count=1 -v` → pass
- [x] 9.3 执行 migration/runtime 收口 gate。验证: `go test ./backend/internal/ai/registry -count=1` → pass；`python3 scripts/lint/prompt_lint.py` → `prompt_lint: 26 files clean`；`python3 scripts/lint/rubric_lint.py` → `rubric_lint: 26 files clean`；`python3 scripts/lint/migrations_lint.py` → `migration lint: ok`；`DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable make migrate-check` → pass

## BDD-Gate

> **BDD 不适用**: 本 plan 落地内部契约（语言无关 output schema 真理源 + `scripts/lint/` 静态 gate + registry 加载/resolver 接线 + A3 provider-neutral 校验器 enum 扩展 + caller 透传），不新增用户可见 UI、新 HTTP API 行为或端到端业务工作流。`validateOutputSchema` fail-close 复用既有 `AI_OUTPUT_INVALID` 错误路径，仅扩展生效范围与 `enum`，无新端到端用户行为。后续 P0 用户行为流由各 C 域 plan 维护 BDD/E2E gate。
>
> **替代验证 gate**:
>
> 1. `make lint-prompts`（schema 子集 + schema-rendered prompt contract + complete example JSON output schema-valid + schema↔prompt↔struct 三向一致性 + negative fixtures）+ 顶层 `make lint`
> 2. `go test ./backend/internal/ai/registry/... -race`（loader 加载 schema + `ResolveActive` OutputSchema 非空 + 语言无关单份 + fallback 一致 + seed migration 覆盖全部 active prompt/rubric 坐标）
> 3. `go test ./backend/internal/ai/aiclient/... -race`（`validateOutputSchema` enum / required fail-close）
> 4. `go test ./backend/internal/{targetjob,resume/jobs,review,practice,debrief,jdmatch}/... -race`（caller `OutputSchema` 透传 + 端到端 fail-close）
> 5. grep red-line：语音 feature_key（STT/TTS）零 schema、零 `OutputSchema`、业务包零 `response_format`
> 6. `DATABASE_URL=... make migrate-check`（migration lint + Postgres dev-stack `up -> down -> up`）
> 7. `python3 .agent-skills/implement/shared/scripts/validate_context.py` + `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
