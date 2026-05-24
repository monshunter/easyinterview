# F3 Language Coordinate Simplification: canonical multi truth source

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-24

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 F3 prompt / rubric registry 从“每个 baseline feature_key 同时维护 `multi` + `en` 两套几乎相同 truth source”收敛为“每个 baseline feature_key 只维护 canonical `multi` truth source，运行时 `language` 仍作为用户目标输出语言和 provenance 字段”。完成后：

- `config/prompts/<feature_key>/v0.1.0.{yaml,md}` 与 `config/rubrics/<feature_key>/v0.1.0.yaml` 是 baseline 唯一 active truth source。
- `config/prompts/*/v0.1.0.en.{yaml,md}` 与 `config/rubrics/*/v0.1.0.en.yaml` 被删除，不再进入 seed migration、loader snapshot、lint 期望或测试断言。
- `ResolveActive(featureKey, requestedLanguage)` 仍接受 `en` / `zh-CN` / `fr` 等运行时语言请求；当没有 exact coordinate 时统一 fallback 到 `multi`，prompt body 通过 `{{language}}` 控制模型输出语言。
- rubric 评估标准保持语言无关；用户可见本地化由前端 i18n / UI locale 负责，不通过复制 rubric 版本实现。
- 后续如果某语言确有语义差异（例如地区法规、面试习惯、英文专项训练），可新增显式 language override，但必须有 spec 决策、业务 rationale 和 lint 证明，不能仅为“英语版”复制 baseline。

本计划不切真实 Model Profile、不改 A3 provider contract、不新增用户可见 UI/API/业务流程。

## 2 背景

当前 repo 已完成 `002-output-schema-contract`，`output_schema` 已是语言无关 truth source；prompt body 输出契约由 schema 渲染，rubric 在忽略 `language` 字段后完全一致。继续维护 `multi` + `en` baseline 会制造重复维护面：

- prompt body 和 YAML hash 双份变更，seed migration 双份同步；
- rubric 评估标准双份复制，容易出现权重/description drift；
- loader / lint / tests 把“至少两个语言坐标”当成正确性，实际只验证重复配置存在；
- 用户真正关心的是输出语言、内容准确性和评估稳定性，不关心隐藏 prompt 文件是否有 `en` 副本。

因此本 plan 将 `language` 的职责改清楚：它是 runtime request / provenance / output-language target；baseline config storage 默认只有 `multi`。多语言用户体验由 `{{language}}`、调用方传入语言、frontend i18n 与 language consistency gate 保证，而不是靠复制 prompt/rubric 文件。

## 3 质量门禁分类

- **Plan 类型**: `contract + tooling + migration + code-internal + docs`。本 plan 修改 F3 spec、`config/prompts` / `config/rubrics` README、lint 脚本、registry loader/resolver tests、seed migration 和 active config truth source。
- **TDD 策略**: Code plan requires TDD。通过 `/implement` -> `/tdd` 串行执行：先改/新增失败断言（prompt/rubric lint 单坐标规则、registry snapshot/fallback、seed migration coverage、negative grep），再删除 `en` 文件和更新实现，最后跑 focused + adjacent gates。
- **BDD 策略**: BDD 不适用。本 plan 不新增用户可见 UI、新 HTTP API 行为或端到端业务工作流；它收敛内部 registry truth source 和运行时 fallback 语义。用户语言体验的可见文本仍由各业务/前端 plan 的 i18n/BDD gate 承接。
- **替代验证 gate**: `python3 -m pytest scripts/lint/prompt_lint_test.py -q`、`python3 scripts/lint/prompt_lint.py`、`python3 scripts/lint/rubric_lint.py`、`go test ./backend/internal/ai/registry -count=1`、focused registry fallback tests、`python3 scripts/lint/migrations_lint.py`、`DATABASE_URL=... make migrate-check`、negative search for `v0.1.0.en` / `13 feature_keys * 2 languages` / `>=2 language coordinates` in current truth sources, `make lint-prompts`, `make lint-rubrics`, `make lint`, `validate_context.py`, `sync-doc-index --check`, `make docs-check`, `git diff --check`.

## 4 实施步骤

### Phase 0: 当前契约预读与坐标面清单

#### 0.1 前后端契约预读

读取 `docs/development.md` §2、`config/README.md`、`config/prompts/README.md`、`config/rubrics/README.md`、`backend/README.md`、`migrations/README.md`、`deploy/dev-stack/README.md`，确认本 plan 不涉及 OpenAPI operation matrix 或 UI parity。

#### 0.2 language coordinate surface inventory

用 `rg` / `find` 列出当前所有 `v0.1.0.en.*` 文件、seed migration `en` rows、loader/lint/test 中 `multi + en` / `>=2 language` / `26 coordinates` 断言、docs 中需要从当前 truth 语义移除的正向要求。

#### 0.3 baseline invariant lock

锁定本 plan 的 runtime invariant：baseline storage 必须只有 `language: multi`；`ResolveActive(featureKey, "en")` / `"zh-CN"` / `"fr"` must fallback to `multi`；returned prompt must still include `{{language}}` or equivalent output-language instruction; output schema stays language-independent and shared.

### Phase 1: spec 与 README 语义修订

#### 1.1 F3 spec v2.7

修订 `docs/spec/prompt-rubric-registry/spec.md`：D-1/D-2/D-6/D-7/D-12、§2.1、§4.1、§4.3、§6 C-1/C-6/C-13、§7 plan order 明确 baseline canonical `multi` only，language variants optional 且必须有语义差异 rationale。

#### 1.2 config prompt/rubric README

修订 `config/prompts/README.md` 与 `config/rubrics/README.md`：删除“每个 feature_key 至少两个 language coordinates”的 baseline 要求；写明 exact-language override 只有在语义差异真实存在时才允许；rubric 默认语言无关，UI 本地化不通过 rubric variant 实现。

#### 1.3 历史完成计划边界

保留 `001`/`002` 已完成计划中的历史 evidence 语句作为历史记录，不把它们当作当前 truth；只更新当前 active spec、README、003 plan/checklist/context 与索引。若当前 lint/search gate 需要排除历史 completed evidence，必须在 checklist 中显式说明排除范围。

### Phase 2: lint / tests 红灯调整

#### 2.1 prompt lint 单坐标规则

先更新 `scripts/lint/prompt_lint_test.py`，增加/调整用例证明 baseline 只需要 `multi`；`en` variant 若存在但与 `multi` 仅重复或没有 rationale，应失败；prompt body 必须保留按 `{{language}}` 输出的指令。随后修改 `scripts/lint/prompt_lint.py`。

#### 2.2 rubric lint 单坐标规则

先更新 rubric lint 测试，证明 baseline rubric 只需要 `multi`，不要求 prompt/rubric language set 等于 `multi + en`；若出现 language override，必须与 prompt override policy 一致且不能只是复制 `language` 字段。随后修改 `scripts/lint/rubric_lint.py`。

#### 2.3 registry tests

先更新 Go focused tests：`SnapshotSize` 期望 13；loader 要求每个 feature_key 有 `multi` prompt/rubric；`ResolveActive(..., "en")` 与 unknown locale 都 fallback 到 `multi`，fallback counter 递增；output schema 仍 language-independent。

#### 2.4 seed coverage tests

先更新 `backend/internal/ai/registry/db_integration_test.go` 期望 active truth-source 坐标只来自 `multi` 文件；扫描所有 seed migration 时拒绝 extra `en` rows 和 prompt hash drift。

### Phase 3: 删除 `en` truth source 与 seed rows

#### 3.1 删除 prompt/rubric `en` 文件

删除 13 个 `config/prompts/*/v0.1.0.en.md`、13 个 `config/prompts/*/v0.1.0.en.yaml`、13 个 `config/rubrics/*/v0.1.0.en.yaml`。保留 13 个 `multi` prompt、13 个 `multi` rubric 和 13 个语言无关 output schema。

#### 3.2 更新 seed migration

按当前 pre-launch clean-break 规则，移除 baseline seed migrations 中 `language='en'` 的 prompt_versions / rubric_versions rows；确保 `up` 只写 13 个 active prompt rows + 13 个 active rubric rows，`down` 与当前 rows 对齐，`ON CONFLICT` 幂等语义保留。

#### 3.3 hash 与 generated prompt contract 保持

确认保留下来的 `multi` prompt body 与 `template_hash` 不漂移；output schema 文件不参与 `template_hash`；prompt contract block 仍由 schema renderer 校验。

### Phase 4: loader / resolver / docs consumer 收敛

#### 4.1 loader parity rule

修改 `backend/internal/ai/registry/loader.go` 的 prompt/rubric language parity 检查：要求每个 feature_key 至少有 `multi`，不要求 `en`；若 exact language override 存在，prompt/rubric 必须成对存在。

#### 4.2 resolver fallback behavior

保持 `selectByLanguage` 的 exact -> `multi` 逻辑；更新测试证明 `en` 请求在无 `en` coordinate 时使用 `multi` prompt/rubric/schema，且 caller 仍拿到 requested language 用于 prompt interpolation / provenance。

#### 4.3 stale contract negative search

清理当前 truth-source docs、README、lint/test 中 `multi + en`、`26 coordinates`、`>=2 language coordinates`、`v0.1.0.en` 的正向要求。Completed plan 历史 evidence 可保留，但 active spec/README/current tests 不得继续要求 `en` baseline。

### Phase 5: 验证、生命周期与收口

#### 5.1 focused verification

运行 prompt/rubric lint 单测与脚本、registry focused/adjacent tests、seed coverage focused test、migration lint。

#### 5.2 integration verification

在 dev-stack Postgres 可用时运行 `DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable make migrate-check`；如环境不可用，记录 blocker 并至少完成 static migration/lint gates。

#### 5.3 docs / lifecycle closeout

运行 `validate_context.py`、`sync-doc-index --check`、`make docs-check`、`make lint`、`git diff --check`；完成后将 003 plan/checklist 置为 `completed`，同步 plans INDEX、work journal、retrospective。如发现真实缺陷，按 `/bug-report` 建档。

## 5 验收标准

- 当前 active F3 spec 与 config README 明确 baseline canonical `multi` only；language variant 仅作为有语义差异的显式 override。
- `find config/prompts -name 'v0.1.0.en.*'` 与 `find config/rubrics -name 'v0.1.0.en.yaml'` 均为 0；`find config/prompts -mindepth 2 -name 'v0.1.0.yaml' | wc -l` = 13；`find config/rubrics -mindepth 2 -name 'v0.1.0.yaml' | wc -l` = 13；`find config/prompts -mindepth 2 -name 'v0.1.0.schema.json' | wc -l` = 13。
- `ResolveActive(featureKey, "en")` / `"zh-CN"` / `"fr"` fallback 到 `multi`，仍返回非空 `OutputSchema`，fallback counter 行为可测试。
- `scripts/lint/prompt_lint.py` / `rubric_lint.py` 不再要求每个 feature_key 至少两个 language coordinates；但仍拒绝 orphan override、hash drift、schema/prompt/struct drift、rubric weight/allowlist drift。
- seed migration 与 static coverage gate 不再包含 active `en` rows；migration lint 与 migrate-check 通过。
- Negative search 证明 current truth sources 不再正向要求 `multi + en` / `26 coordinates` / `>=2 language coordinates`；历史 completed evidence 若保留，必须不被当前 gate 当作 active truth。
- BDD 不适用声明与替代验证 gate 已固化；plan/checklist/context/index 生命周期闭环。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 删除 `en` 文件后用户英文输出退化 | prompt `multi` body 必须保留 `{{language}}` 输出语言指令；resolver fallback tests 覆盖 `en` 请求；业务调用仍传 requested language 给 template interpolation / provenance。 |
| 未来真的需要 per-language semantic override | spec/README 允许 override，但必须有业务 rationale、prompt/rubric 成对文件和 lint gate；禁止无差异复制。 |
| seed migration 修改影响已有本地库 | 当前项目未上线且 AGENTS 明确无需历史兼容；使用 dev-stack `migrate-check` 验证 clean-break migration chain。必要时提示本地开发库 reset。 |
| 历史 completed plan 中仍有 `13 × 2` evidence | 历史 evidence 保留为交付记录；current truth negative search 只对 spec/README/config/lint/tests/migrations 生效，避免改写历史造成审计失真。 |
| loader 误删 parity 检查导致 orphan override | parity 规则改成“baseline `multi` 必须存在；任意 override 必须 prompt/rubric 成对”，并加 negative test。 |

## 7 关联文档导航

> - [Prompt Rubric Registry Spec](../../spec.md)
> - [001-baseline plan](../001-baseline/plan.md)
> - [002-output-schema-contract plan](../002-output-schema-contract/plan.md)
> - [AI Provider and Model Routing Spec](../../../ai-provider-and-model-routing/spec.md)
> - [Development Contract Workflow](../../../../development.md)

## 8 Handoff / Evidence Log

### 8.1 Phase evidence log

| Phase | Evidence slot | Required command / artifact |
|---|---|---|
| Phase 0 | 当前契约与坐标清单 | docs/development §2 + module README preflight；`rg` / `find` language-coordinate surface inventory；baseline invariant lock |
| Phase 1 | spec + README 语义 | spec/history/index v2.7；config prompt/rubric README current-language policy |
| Phase 2 | lint/test 红绿 | prompt/rubric lint tests；registry focused tests；seed coverage focused test |
| Phase 3 | truth source deletion | `find` no `v0.1.0.en.*`; 13 prompt/rubric/schema counts; migrations no active `en` rows |
| Phase 4 | loader/resolver/current truth 收敛 | loader/resolver tests；stale-contract negative search |
| Phase 5 | 验证与生命周期 | lint/test/migrate/docs gates；BUG/report/work-journal/retrospective closeout |

### 8.2 Phase 0 handoff snapshot

实施日期：2026-05-24。当前快照基于 `docs/development.md` §2、`config/README.md`、`config/prompts/README.md`、`config/rubrics/README.md`、`backend/README.md`、`migrations/README.md`、`deploy/dev-stack/README.md`、`config/prompts` / `config/rubrics` truth source、`migrations/*seed_baseline_prompt_rubric*.up.sql`、`scripts/lint` 与 `backend/internal/ai/registry` 搜索结果。

#### 8.2.1 契约预读结论

- 本 plan 不新增或修改 HTTP API、OpenAPI fixture、generated client/server artifact、用户可见 UI 或端到端业务流；`docs/development.md` §2.1 operation matrix 对本 plan 为 `N/A`，替代 gate 是 lint / registry unit test / migration check。
- `config/README.md` 要求 F3 prompt/rubric registry truth source 位于 `config/` 且不得持有 provider/model secret；本 plan 只改 F3 truth source 和 seed rows，不新增 env key 或 secret。
- `backend/README.md` 要求 backend workstream 保持 AI provider boundary；本 plan 不调用真实 LLM，Go tests 使用 registry/static fixtures。
- `migrations/README.md` 指向 B4 ownership；本 plan 在 pre-launch clean-break 语义下收敛 baseline seed rows，并用 migration lint / migrate-check 验证。
- `deploy/dev-stack/README.md` 确认 local integration 只需 Docker Compose 外部依赖，默认不引入 Kind/K8s/Helm；`make migrate-check` 以 dev-stack Postgres 为 live gate。

#### 8.2.2 language coordinate surface inventory

- 当前待删除文件：`config/prompts/*/v0.1.0.en.md` 13 个、`config/prompts/*/v0.1.0.en.yaml` 13 个、`config/rubrics/*/v0.1.0.en.yaml` 13 个。
- 当前保留基线计数：`config/prompts` 下 `v0.1.0.yaml` / `v0.1.0.md` / `v0.1.0.schema.json` 加 `en` 文件共 65；`config/rubrics` 下 `v0.1.0.yaml` + `v0.1.0.en.yaml` 共 26。完成后目标为 prompt 39（13 yaml + 13 md + 13 schema）与 rubric 13。
- Seed rows：`migrations/000002_seed_baseline_prompt_rubric_versions.up.sql` 仍包含 11 个历史 feature_key 的 `en` prompt/rubric rows；`migrations/000010_seed_baseline_prompt_rubric_versions_jd_match.up.sql` 仍包含 2 个 JD-Match feature_key 的 `en` prompt/rubric rows。
- 当前正向要求：`config/prompts/README.md` / `config/rubrics/README.md` 要求每个 feature_key 至少两个 language coordinates；`backend/internal/ai/registry/{loader_test.go,registry_test.go}` 断言 `multi + en` / 26 coordinates；`migrations/000002` 注释写 2 language coordinates。Completed `001` / `002` plan 与 history 中的 `13 × 2` / `26` 为历史 evidence，Phase 4 negative search 只将它们列为历史范围。

#### 8.2.3 baseline invariant

- Baseline storage target：每个 active feature_key 只保留 `language: multi` prompt/rubric coordinate；output schema 继续每 `(feature_key, version)` 一份，语言无关。
- Runtime target：`ResolveActive(featureKey, "en")`、`ResolveActive(featureKey, "zh-CN")`、`ResolveActive(featureKey, "fr")` 在没有 exact override 时 fallback 到 `multi`；fallback counter 应递增；返回的 `OutputSchema` 非空且与 `multi` coordinate 共享。
- Prompt output-language target：13 个保留的 `v0.1.0.md` 已包含 `{{language}}` 或等价 output-language instruction；JSON keys 保持 ASCII，用户可见输出语言由 prompt 变量和 caller 传入语言控制。

### 8.3 Phase 2 lint TDD evidence

- Prompt lint red: `python3 -m pytest scripts/lint/prompt_lint_test.py -q -k 'language_override or runtime_language'` initially failed because duplicate `en` prompt coordinates and missing `{{language}}` body instruction were still accepted as clean.
- Prompt lint green: same focused command passed after `scripts/lint/prompt_lint.py` added canonical `multi` storage, explicit language override allowlist, and `{{language}}` runtime-language instruction checks.
- Rubric lint red: `python3 -m pytest scripts/lint/rubric_lint_test.py -q -k 'language_override'` initially failed because duplicate `en` rubric coordinates were still accepted as clean.
- Rubric lint green: same focused command passed after `scripts/lint/rubric_lint.py` added canonical `multi` storage and explicit language override allowlist checks.

### 8.4 Phase 2-4 registry / seed / deletion evidence

- Go registry red: focused `go test ./backend/internal/ai/registry -run 'Test(NewRegistryClientLoadsAllBaselines|LoadHappyPath|LoadMissingCanonicalMultiRejected|LoadOrphanLanguageOverrideRejected|Resolve.*Language|ResolveActiveReturnsOutputSchema|SeedMigrationCoversBaselineFeatureKeys)' -count=1 -v` initially failed on `SnapshotSize: want 13, got 26`, extra `en` seed rows, and missing canonical multi not rejected.
- Go registry green: same focused command passed after deleting `en` truth-source files, filtering seed rows to canonical `multi`, adding loader canonical-multi validation, and updating resolver tests for `en -> multi` fallback.
- Deletion/count proof: `find config/prompts -name 'v0.1.0.en.*'` and `find config/rubrics -name 'v0.1.0.en.yaml'` produced no output; retained counts are 13 prompt YAML, 13 prompt Markdown, 13 prompt schema, and 13 rubric YAML.
- Static lint proof: `make lint-prompts` returned `prompt_lint: 13 files clean`; `make lint-rubrics` returned `rubric_lint: 13 files clean`; `python3 scripts/lint/migrations_lint.py` returned `migration lint: ok`.

### 8.5 Phase 4 stale-contract search classification

- `rg -n 'multi \+ en|26 coordinates|>=2 language|v0\.1\.0\.en' config migrations` produced no output, proving current config truth source and seed migrations no longer carry stale `en` baseline requirements.
- `rg -n 'multi \+ en|26 coordinates|>=2 language' backend/internal/ai/registry scripts/lint docs/spec/prompt-rubric-registry/spec.md` produced only spec C-13 wording that explicitly says active README/spec/lint/test no longer require `multi + en` or `>=2 language coordinates`.
- Remaining `v0.1.0.en` strings in `backend/internal/ai/registry/*_test.go` and `scripts/lint/*_test.py` are negative fixtures that synthesize non-allowlisted language overrides to prove the new guard rejects them.
- Remaining `v0.1.0.en` / `multi + en` strings in this 003 plan/checklist/context are cleanup scope, inventory, deletion proof, or negative-search gate text, not active storage requirements.

### 8.6 Phase 5 final verification evidence

- `python3 -m pytest scripts/lint/prompt_lint_test.py -q` passed: 15 tests.
- `python3 -m pytest scripts/lint/rubric_lint_test.py -q` passed: 6 tests.
- `go test ./backend/internal/ai/registry -count=1` passed.
- `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/prompt-rubric-registry/plans/003-language-coordinate-simplification/context.yaml --docs-root docs --target backend` passed and resolved the plan/checklist/spec/reference files.
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` passed with zero drift.
- `make docs-check` passed.
- `git diff --check` passed.
- `DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable make migrate-check` passed with current target output `migration lint: ok`.
- `make lint` passed, including prompt/rubric lint, OpenAPI validation, fixture validation, runtime topology, docs lint, and frontend placeholder lint.
