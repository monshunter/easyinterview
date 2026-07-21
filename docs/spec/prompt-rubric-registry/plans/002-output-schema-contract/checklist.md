# F3 Output Schema Contract Checklist

> **版本**: 2.16
> **状态**: active
> **更新日期**: 2026-07-21

**关联计划**: [plan](./plan.md)

## Phase 0: 基线复核与缺口确认（只读 + handoff snapshot）

- [x] 0.1 复核 plan §3.1 operation matrix 当前 9 行的 prompt body 输出 key 与后端 struct json tag。验证: 9 行对照表写入 plan §8.2 Phase 0 handoff
- [x] 0.2 读取 `backend/internal/resume/jobs/tailor.go` 的 `decodeTailorAIResponse` / `normalizeSuggestions`，判定 `resume.tailor.bullet_suggestions` 当时的 prompt key（`rewrite`/`why_better`/`kept_facts`）与 canonical parser key（`originalBullet`/`suggestedBullet`/`reason`）属于新契约 drift。验证: 判定结论写入 handoff；Phase 2.3 统一 prompt/schema，Phase 11 删除 runtime alias
- [x] 0.3 读取 `backend/internal/ai/aiclient/observability/decorator.go:531-594`，记录 `outputSchema` 类型与 `validateAgainstSchema` 的 `enum` 缺口与扩展点。验证: 扩展点写入 handoff
- [x] 0.4 读取 `loader.go`（`WalkDir`/`computeTemplateHash`）、`resolver.go:62`、`types.go:15`，确认 schema 不混入 per-language `template_hash` 的加载设计可行。验证: 加载/接线现状写入 handoff
- [x] 0.5 列出直接型 caller（`review`×2 / `practice`-chat×3）与本地-resolution 型 caller（`targetjob` / `resume/jobs`×2），区分 `OutputSchemaVersion`（标签）vs `OutputSchema`（实际 schema），标注 `practice/voice_turn_service.go` STT/TTS 跳过点。验证: caller 清单写入 handoff

## Phase 1: README 契约 + prompt 输出契约渲染规则

- [x] 1.1 `config/prompts/README.md` 新增 output schema 约定（文件落点、语言无关、JSON Schema 校验子集、`description` 非校验注解、不混入 `template_hash`、与 prompt body / struct 一致性）。验证: README 含 output schema 章节；`make lint-prompts` 仍绿
- [x] 1.2 在 README 固化 prompt body 输出契约块规范：字段顺序来自 schema，字段说明来自 `description`，complete example JSON output 由 schema 生成并可解析，覆盖 schema 声明的 required + optional 字段，使用业务形态值，并明确不是 JSON Schema / OpenAPI schema；人不在当前 18 个 `.md` 手工维护第二份字段表。验证: README 明确 schema 是唯一字段真理源，并说明 renderer/lint 的 drift 行为
- [x] 1.3 在 README / handoff 规则中明确 alias / optional 字段策略：受 output schema 约束的 parser 只接受 canonical keys，out-of-scope alias 只作为 negative-test 输入或历史 drift 证据；prompt-only 可选字段必须有 `description` 说明评估价值。验证: README 或 plan §8 handoff 模板包含 alias/optional 判定项

## Phase 2: 9 份语言无关 output schema 文件

- [x] 2.1 为当前 9 个 chat feature_key 各落地 `config/prompts/<feature_key>/v0.1.0.schema.json`（语言无关，顶层 `object`，`required`=后端依赖字段，含 `enum`，允许 `description` 注解，只用允许关键字）。验证: `find config/prompts -mindepth 2 -name 'v0.1.0.schema.json' | wc -l` = 9；`jq empty` 解析全部通过；required/关键可选字段含 `description`
- [x] 2.2 人工核对每份 schema 的 `required`/字段名与后端 struct json tag / parser required key 一致。验证: 核对表写入 plan §8 handoff（Phase 3 lint 强制）
- [x] 2.3 修复 `resume.tailor.bullet_suggestions` schema/prompt canonical key，使用 `originalBullet` / `suggestedBullet` / `reason` 并补 round-trip 断言；Phase 11 删除 parser alias fallback。验证: `go test ./backend/internal/resume/jobs/... -race` 全绿

## Phase 3: schema lint + prompt 契约 renderer（先红后绿）

- [x] 3.1 扩展 `scripts/lint/prompt_lint.py`（或新增 `schema_lint.py`）+ 单测：① schema 只含 `type`/`required`/`properties`/`items`/`enum` + `description`；② 每个当前 chat feature_key 有 schema（语音豁免清单）；③ schema `required` ⊆ prompt body 输出 key；④ schema 字段 ↔ 后端 struct json tag / parser required key 一致（`resume.parse`/`resume.tailor.gap_review` 因停在 `json.RawMessage`，④ 降级为 schema ↔ parser required key 一致）。验证: `python3 -m pytest scripts/lint/prompt_lint_test.py` 全绿；≥4 negative fixtures（非法关键字 / required 不在 prompt / 字段与 struct 不一致 / prompt contract block 手工漂移）红→绿
- [x] 3.2 新增 prompt output contract renderer（可内嵌于 `prompt_lint.py` 或独立 helper）：从 schema 生成稳定的输出契约块和 complete representative example JSON output，覆盖 object、array、enum、nested object/array、`description` 缺失报错、example JSON schema-valid、optional 字段完整示例与业务形态值。验证: renderer 单测全绿
- [x] 3.3 `make lint-prompts` 覆盖 schema gate + rendered prompt contract gate；顶层 `make lint` 联动。验证: `make lint-prompts && make lint` 全绿

## Phase 4: 9 × 2 prompt body 输出段统一（由 schema 渲染/校验）

- [x] 4.1 把当前 9 × 2 个 `.md` 的 “Return strict JSON …” prose 段替换为 schema 渲染/校验的「输出契约 + complete example JSON output」块，重算 `template_hash` 写回 `.yaml`。验证: `make lint-prompts`（hash 一致 + rendered block 一致）全绿
- [x] 4.2 prompt body drift gate 负向验证：故意改一个 prompt contract 字段名或 example key，`make lint-prompts` 失败；还原故意改动后通过。验证: negative fixture 或测试覆盖

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
- [x] 7.3 `review`/`practice`(chat) 直接读 `resolution.OutputSchema` 填 metadata；`practice/voice_turn_service.go` STT/TTS 跳过，仅 chat 接线。验证: `go test ./backend/internal/{review,practice}/... -race` 全绿
- [x] 7.4 Out-of-scope feature_key negative gate：Debrief / JD Match 不出现在当前 `config/prompts` / `config/rubrics` truth source，且 lint 对 out-of-scope key fail-fast。验证: prompt/rubric out-of-scope-key negative tests 全绿
- [x] 7.5 grep red-line：语音 feature_key 不落 schema、不接 `OutputSchema`；业务包无 `response_format`/`json_schema` 请求字段。验证: `find config/prompts -name '*.schema.json' | grep -E 'voice|stt|tts|dictation'` 返回 0 行（语音 feature_key 无 schema）；`! grep -rnE '"response_format"|json_schema' backend/internal` provider 请求字段命中为 0
- [x] 7.6 收口：`make lint-prompts` + `go test ./backend/internal/ai/registry/... ./backend/internal/ai/aiclient/... ./backend/internal/{targetjob,resume/jobs,review,practice}/... -race` + `validate_context.py` + `sync-doc-index --check`；§8 handoff 列 C-12 证据；Header 切 `completed` 同步 INDEX 与工作日志。验证: 全部命令绿；INDEX 显示 completed

## Phase 8: L2 prompt example remediation（完整 JSON output）

- [x] 8.1 为 renderer 增加回归测试：example JSON 必须包含 schema-declared optional properties，且不能使用 `string` / `1` 这类 generic filler values。验证: red `python3 -m pytest scripts/lint/prompt_lint_test.py -q -k 'rendered_example_includes_optional_properties'` 失败于 optional 字段缺失；green `python3 -m pytest scripts/lint/prompt_lint_test.py -q -k 'rendered_example'` → `2 passed`
- [x] 8.2 重渲染当前 9 × 2 prompt body，刷新 18 个 YAML `template_hash`，同步 `migrations/000002_seed_baseline_prompt_rubric_versions.up.sql` 中已有 prompt seed row 的 body/hash，并补充 `resume.parse` project/education optional schema 字段。验证: `make lint-prompts` → `prompt_lint: 9 files clean`；`rg -n '"string"|: 1,|Example JSON:' config/prompts -g '*.md'` → 0 matches
- [x] 8.3 同步 spec/plan/checklist/context/index 与收口验证。验证: `validate_context.py` + `sync-doc-index --check` + `git diff --check` 通过；Header 确认为 `completed`

## Phase 9: L2 seed migration coverage remediation（active truth source 全量覆盖）

- [x] 9.1 重写 seed migration 静态覆盖测试，从 `config/prompts` active YAML 与 `config/rubrics` YAML 反推出期望坐标，扫描 seed / delete migrations 的净效果，拒绝 missing / extra / duplicate row 与 prompt `template_hash` drift。验证: focused seed coverage red→green
- [x] 9.2 Out-of-scope feature_key seed net-zero gate：out-of-scope seed rows 只作为 migration audit rows，当前迁移链最终不得留下 out-of-scope active prompt/rubric coordinate。验证: prompt/rubric/migration lint 与 migrate-check 全绿
- [x] 9.3 执行 migration/runtime 收口 gate。验证: `go test ./backend/internal/ai/registry -count=1` → pass；`python3 scripts/lint/prompt_lint.py` → `prompt_lint: 9 files clean`；`python3 scripts/lint/rubric_lint.py` → `rubric_lint: 9 files clean`；`python3 scripts/lint/migrations_lint.py` → `migration lint: ok`；`DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable make migrate-check` → pass

## Phase 10: L2 review follow-up（out-of-scope key rejection + lint diagnostic hardening）

- [x] 10.1 Out-of-scope feature_key rejection：当前 config truth source 中无 Debrief / JD Match prompt/rubric 目录；人为写入 out-of-scope key 时 lint 返回 clear diagnostic。验证: `python3 -m pytest scripts/lint/prompt_lint_test.py -q -k 'jd_match or missing_schema_description'` → pass
- [x] 10.2 修复 `prompt_lint.py` 在 invalid schema 已产生 subset/contract error 时仍调用 renderer 导致 traceback 的缺口。验证: `test_missing_schema_description_reports_lint_error_without_traceback` 先红后绿，断言 stderr 包含 `missing non-empty description` 且不含 `Traceback`
- [x] 10.3 执行收口验证并同步 Bug / retrospective 文档。验证: `python3 -m pytest scripts/lint/prompt_lint_test.py -q` → pass；`python3 scripts/lint/prompt_lint.py` → `prompt_lint: 9 files clean`；`python3 scripts/lint/rubric_lint.py` → `rubric_lint: 9 files clean`；`python3 scripts/lint/migrations_lint.py` → `migration lint: ok`；`DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable make migrate-check` → `migration lint: ok`；`go test ./backend/internal/ai/registry -count=1` → pass；out-of-scope feature key negative tests → pass；`go test ./backend/internal/ai/aiclient/observability -run TestDecorator_OutputSchema -count=1` → pass；`validate_context.py --context docs/spec/prompt-rubric-registry/plans/002-output-schema-contract/context.yaml --docs-root docs --target backend` → pass；`sync-doc-index.py --check` → zero drift；`git diff --check` → pass

## Phase 11: Resume tailor output alias removal

- [x] 11.1 Red: handler tests 证明 alias-only gap-review / bullet-suggestion 输出返回 `AI_OUTPUT_INVALID`，canonical 输出保持成功且冲突 alias 不覆盖 canonical 值。验证: focused test 在现有 fallback parser 上先失败
  <!-- verified: 2026-07-10 red="cd backend && go test ./internal/resume/jobs -run 'TestTailorHandlerModeRoutingAndFailurePaths|TestTailorHandlerBulletSuggestionsCanonicalKeysRoundTrip' -count=1" result="alias-only root match summary and suggestion cases were incorrectly persisted as success" -->
- [x] 11.2 Green: `decodeTailorAIResponse` / `normalizeMatchSummary` / `normalizeSuggestions` 只消费 schema canonical keys，删除 root-level 与 snake_case/rewrite/why-better fallback、请求输入回填和无调用 helper。验证: 11.1 focused tests 全绿
  <!-- verified: 2026-07-10 green="cd backend && go test ./internal/resume/jobs -run 'TestTailorHandlerModeRoutingAndFailurePaths|TestTailorHandlerBulletSuggestionsCanonicalKeysRoundTrip' -count=1" result="pass; production parser alias search returned zero matches" -->
- [x] 11.3 Verify: `go test ./internal/resume/jobs -count=1`、`make lint-prompts`、prompt/rubric lint 与生产 parser alias 负向搜索通过；`{{original_bullet}}` 输入模板变量作为明确例外
  <!-- verified: 2026-07-10 commands="cd backend && go test ./internal/resume/jobs -count=1; make lint-prompts; python3 scripts/lint/{prompt_lint,rubric_lint}.py; production parser alias rg" result="pass; 9 prompt and 9 rubric files clean; only input template variable remains" -->
- [x] 11.4 Docs/closure: 统一本 owner 的 out-of-scope/current-contract 术语，四个 plan context 对齐 spec 2.19；registry preflight 从精确 spec 版本号改为 D-13 + 当前 9-key 稳定语义断言；运行 context/index/docs/diff/pruning gates 并确认状态为 `completed`
  <!-- verified: 2026-07-10 red="registry preflight expected spec 2.17 while current spec was 2.19" green="focused preflight and full registry package pass" gates="four F3 contexts; resume jobs; output-schema observability; prompt/rubric/migration lint; docs-check; diff-check; pruning surface" result="pass; real_residuals=0" -->

## Phase 12: Practice output alias removal and stable F3 references

- [x] 12.1 Current contract: F3 spec 2.20 和 `config/prompts/README.md` 明确 runtime parser canonical-only；四个 plan context 投影到 2.20，plan/checklist/INDEX 以 2.1 active 原地承接
  <!-- verified: 2026-07-10 commands="validate_context.py --target backend for F3 plans 001-004; sync-doc-index --check; git diff --check" result="pass; zero drift" -->
- [x] 12.2 Red: practice focused tests 锁定 `question` / `intent` / `hint` / `answer_summary` alias-only 输出失败或被忽略，canonical 输出成功且冲突 alias 不覆盖 canonical 值
  <!-- verified: 2026-07-10 command="cd backend && go test ./internal/practice -run 'TestParseFirstQuestionUsesCanonicalOutputKeys|TestParseHintUsesCanonicalCue|TestParseTurnObservationUsesCanonicalAnswerSummary' -count=1" result="failed on all four existing alias paths" -->
- [x] 12.3 Green: 删除 `parseFirstQuestion` 与 `hintAIResponse` 的四类 output alias/fallback；成功 fixture 统一 canonical keys，范围外 alias 只留在 12.2 negative tests
  <!-- verified: 2026-07-10 command="cd backend && go test ./internal/practice -run 'TestParseFirstQuestionUsesCanonicalOutputKeys|TestParseHintUsesCanonicalCue|TestParseTurnObservationUsesCanonicalAnswerSummary' -count=1" result="pass; production parser alias search returned zero" -->
- [x] 12.4 Stable references/closure: 删除 F3 lint/registry 注释中的固定 spec 版本引用；运行 practice/registry tests、prompt/rubric lint、alias/version negative search、四个 context、index/docs/diff/pruning gates 并确认状态为 `completed`
  <!-- verified: 2026-07-10 commands="make lint; practice and registry package tests; 22 lint tests; F3 contexts 001-004; docs-check; diff-check; pruning surface" result="pass; prompt_lint=9 clean; rubric_lint=9 clean; real_residuals=0" -->

## Phase 13: Canonical parser downstream scenario fixture gate

- [x] 13.1 `cmd/api` Practice 确定性成功 fixture 输出 canonical `cue` / `answerSummary`，alias-only `hint` 仅保留在 invalid-output 负测；focused Practice/internal AI/cmd-api tests 从 `session_wait` / 额外 `AI_OUTPUT_INVALID` task run 红态恢复为成功。
- [x] 13.2 将 downstream consumer 场景执行加入 F3 canonical parser 变更的 TDD/替代验证 gate；验证: F3/002 context、INDEX、docs/diff/pruning gates 通过。

## Phase 14: Grounded report + practice semantic-focus version pairs

- [x] 14.1 RED: prompt/schema/preflight tests require direct summary/code+label/dimensionCode/sourceMessageSeqNos/actions/retryFocusDimensionCodes, recursive `additionalProperties=false`, locked min/max/pattern/unique bounds, and reject numeric/old/unknown keys.
  <!-- verified: 2026-07-12 red="candidate v0.2 files absent; then shared validator accepted unknown fields and ignored all bounds while loader accepted open objects" green="prompt lint exact-shape/closed/bounds tests, loader open-schema negative, shared runtime unknown/length/pattern/count/unique/numeric negatives and backend preflight all pass" -->
- [x] 14.2 RED-GREEN: refactor prompt/rubric snapshots to `(feature_key, language, version)`, add exact rubric `status=active|inactive` metadata, and reject duplicate versions, unknown status, zero/multiple active entries, missing active pairs and language/version parity drift; `GetPrompt` / `GetRubric` retrieve exact active or inactive versions.
  <!-- verified: 2026-07-12 red="loader map indexing and RubricSchema.Status tests failed to compile; status lint accepted retired" green="registry focused tests cover duplicate version, unknown status, zero/multiple active, version parity and exact v0.1/v0.2 getters; rubric status lint passes" -->
- [x] 14.3 GREEN: create only immutable-content v0.2.0 prompt/schema, preserve v0.1 prompt/schema/rubric rollback files and rows, and preserve trusted policy/untrusted context/session language; do not create/edit F3/004-owned v0.2 rubric content or the backend-review-owned ReportContent struct. Emit `REPORT_PROMPT_V020_READY` while v0.1 remains active.
  <!-- verified: 2026-07-12 marker=REPORT_PROMPT_V020_READY commands="python3 -m pytest scripts/lint/prompt_lint_test.py scripts/lint/rubric_lint_test.py -q; python3 scripts/lint/prompt_lint.py; python3 scripts/lint/rubric_lint.py; cd backend && go test ./internal/ai/aiclient/outputschema ./internal/ai/aiclient/observability ./internal/ai/registry -count=1" result="27 lint tests pass; 7 prompt and 7 rubric files clean; closed/bounds runtime and registry pass; v0.1 active and v0.2 draft/inactive" -->
- [x] 14.4 RED: exact-version prompt lint/hash and registry/backend-practice candidate fixtures require immutable `practice.session.chat/v0.2.0` draft prompt + closed schema and inactive rubric parity; v0.2 input uses only structured `semanticFocus/{{semantic_focus_json}}`, v0.2 rubric content/weights equal v0.1, and v0.1/000002 remain unchanged.
  <!-- verified: 2026-07-12 method=tdd-red evidence="Python focused tests fail because practice v0.2 assets/validator are absent; registry exact GetPrompt(v0.2.0) fails unsupported; backend-practice exact-candidate test fails before F3 assets exist. v0.1 prompt/hash and 000002 are unchanged." -->
- [x] 14.5 GREEN: add the practice v0.2 prompt/schema/rubric pair, prove prompt/rubric version parity and exact getters while v0.1 remains active, and reject legacy focus tokens in the candidate/runtime fixture without treating immutable v0.1 rollback/history/negative literals as current positive consumers.
  <!-- verified: 2026-07-12 method=tdd-green evidence="practice v0.2 prompt draft hash=0b49cacb..., closed messageText schema, and inactive content-identical rubric load as an exact pair; focused Python 3/3, prompt/rubric lint 8+8 clean, registry candidate preflight and backend-practice exact-candidate test pass; old token positives are limited to immutable v0.1 rollback plus explicit lint negatives." -->
- [x] 14.6 OWNER-GATE: storage、F3/004 rubric 与 context-aware eval prerequisites 均通过；任一 prerequisite 不成立、active state 非唯一、version parity drift 或 immutable-content drift 时，在 status/DB mutation 前失败。
  <!-- verified: 2026-07-12 method=owner-marker-preflight evidence="TestV020ActivationOwnerMarkersReady reads verified owner checklist comments and fails if storage/rubric/context-aware-eval/report-ready marker is absent; all four markers pass before status or migration mutation. Red phase note: owner markers already existed, so behavioral assertion passed after correcting the test-only repo-root helper." -->
- [x] 14.7 GREEN: after all owner markers, flip exactly 8 dev status values across report/practice v0.1/v0.2 prompt+rubric files and prove activate→rollback→re-activate with full-snapshot validation + atomic pointer swap. Create `000019_activate_report_and_practice_prompt_rubric_v020` only via `make migrate-create`; independent PostgreSQL up/down/up activates/restores both pairs in one transaction. Active report v0.2 Resolve provenance must equal `report-context.v1`, while practice v0.2 remains `registry.v1`. Do not edit 000002 or claim cross-media atomicity.
  <!-- verified: 2026-07-12 method=coordinated-activation evidence="TestCoordinatedV020ActivationRollbackReactivate proves final-v0.2 snapshot, partial-edit rejection without pointer replacement, eight-status rollback to v0.1 and reactivation. 000019 was generated through make migrate-create; disposable PostgreSQL integration migrated v18->v19->v18->v19 and emitted both release markers; DATABASE_URL=<disposable> APP_ENV=test make migrate-check passed; final schema_migrations=19 dirty=false with exactly both v0.2 prompt/rubric pairs active." -->
- [x] 14.8 Verify final `ResolveActive` returns v0.2 prompt/rubric for report and practice, exact getters still retrieve both v0.1 coordinates, report/practice data-source coordinates are exact, latest-migration seed/hash parity remains strict, schema/prompt/example/struct gates pass, and runtime rubric injection is absent；backend semantic grounding still requires its own validator/repair gates。
- [x] 14.9 MAX4 RED-GREEN: schema200/semantic24-64/targeted18-52 unchanged。Evalkit generation and judge use independent max4 budgets；generation full-validates and dynamically selects action_labels/whole_report per round；judge retries only provider/protocol invalid and terminally rejects valid negative。Schema/example/seed/export/manifest contract records aggregate usage/latency + attempt_count/retry_count/reason/scope；UI UX 由下游 owner 独立验证。
- [x] 14.10 EXAMPLE-GROUNDING RED-GREEN: retain the complete report JSON example, place a synthetic candidate input immediately before it, make every example fact/action traceable to that input, and require an anti-copy/current-context regeneration instruction. Synchronize template hash, migration body/hash, resolved prompt and active dev DB; do not add an example-omission mode.

## Phase 15: Practice interviewer-identity v0.3 pair

- [x] 15.1 RED: prompt/registry/migration tests require the v0.3 identity-source policy and exact v0.2 rollback while the new coordinate is absent.
  <!-- verified: 2026-07-21 method=tdd-red evidence="active resolver tests returned v0.2, candidate metadata was draft/inactive, and the exact 000023 activation migration was absent; v0.3 owner-marker preflight also failed before verified markers existed" -->
- [x] 15.2 GREEN: add v0.3 prompt/schema/hash and consume the F3 `004` v0.3 rubric; named TargetJob, anonymous TargetJob, Resume-company isolation and assistant-history correction policies are explicit.
  <!-- verified: 2026-07-21 method=prompt-registry-contract evidence="active v0.3 prompt hash=9fff2605..., closed messageText schema and role_identity rubric load as an exact pair; v0.2 remains retrievable as draft/inactive rollback; focused registry/Practice tests and prompt/rubric lint pass" -->
- [x] 15.3 OWNER-GATE: F3 `004` 的 verified `PRACTICE_INTERVIEWER_IDENTITY_V030_PASS` 与 backend-practice behavior gate 均存在，且失败说明/普通文本提及 marker 不得通过，随后才允许 active status 或 DB mutation。
  <!-- verified: 2026-07-21 method=explicit-owner-marker-preflight evidence="preflight RED while marker absent, then GREEN only for verified marker attributes; failure/evidence/unchecked mentions remain rejected" -->
- [x] 15.4 ACTIVATE: flip only practice v0.2/v0.3 prompt/rubric dev statuses with validated snapshot rollback/re-activate; create `000023_activate_practice_interviewer_identity_v030` via `make migrate-create` and pass PostgreSQL up/down/up.
  <!-- verified: 2026-07-21 method=file-snapshot+postgres evidence="practice-only v0.3 activate/v0.2 rollback/reactivate cache test PASS; migration lint and make migrate-check PASS; disposable PostgreSQL integration 22->23->22->23 PASS with report v0.2 untouched and role_identity weight 0.4; dev DB version=23 dirty=false" -->
- [x] 15.5 VERIFY: exact getters retain v0.2/v0.3, active Resolve returns practice v0.3 + `registry.v1`, report v0.2 is untouched, strict parity/lint/focused/root/docs/diff gates pass.
  <!-- verified: 2026-07-21 method=full-closeout evidence="exact getters and active resolver/cache/DB integration PASS with registry.v1 and report v0.2 isolation; all-migration content parity, prompt/rubric lint and focused tests PASS; make test/build/lint/docs-check, three context validators, index and diff gates PASS" -->

## BDD-Gate

> **BDD 不适用**: 本 plan 落地内部契约（语言无关 output schema 真理源 + `scripts/lint/` 静态 gate + registry 加载/resolver 接线 + A3 provider-neutral 校验器 enum 扩展 + caller 透传），不新增用户可见 UI、新 HTTP API 行为或端到端业务工作流。`validateOutputSchema` fail-close 复用既有 `AI_OUTPUT_INVALID` 错误路径，仅扩展生效范围与 `enum`，无新端到端用户行为。后续 P0 用户行为流由各 C 域 plan 维护 BDD/E2E gate。
>
> **替代验证 gate**:
>
> 1. `make lint-prompts`（schema 子集 + schema-rendered prompt contract + complete example JSON output schema-valid + schema↔prompt↔struct 三向一致性 + negative fixtures）+ 顶层 `make lint`
> 2. `go test ./backend/internal/ai/registry/... -race`（loader 加载 schema + `ResolveActive` OutputSchema 非空 + 语言无关单份 + fallback 一致 + seed migration 覆盖全部 active prompt/rubric 坐标）
> 3. `go test ./backend/internal/ai/aiclient/... -race`（`validateOutputSchema` enum / required fail-close）
> 4. `go test ./backend/internal/{targetjob,resume/jobs,review,practice}/... -race`（caller `OutputSchema` 透传 + 端到端 fail-close）
> 5. canonical parser 变更执行相关 `backend/cmd/api` consumer 场景，保证成功 fixture 与 alias-only 负测同时有效
> 6. grep red-line：语音 feature_key（STT/TTS）零 schema、零 `OutputSchema`、业务包零 `response_format`
> 7. `DATABASE_URL=... make migrate-check`（migration lint + Postgres dev-stack `up -> down -> up`）
> 8. `python3 .agent-skills/implement/shared/scripts/validate_context.py` + `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
