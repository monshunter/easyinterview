# F3 Language Coordinate Simplification Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-24

**关联计划**: [plan](./plan.md)

## Phase 0: 当前契约预读与坐标面清单

- [x] 0.1 读取 `docs/development.md` §2、`config/README.md`、`config/prompts/README.md`、`config/rubrics/README.md`、`backend/README.md`、`migrations/README.md`、`deploy/dev-stack/README.md`，确认本 plan 不涉及 OpenAPI operation matrix 或 UI parity。验证: 在 plan §8.2.1 记录预读结论
- [x] 0.2 用 `rg` / `find` 列出所有 `v0.1.0.en.*` 文件、seed migration `en` rows、loader/lint/test 中 `multi + en` / `>=2 language` / `26 coordinates` 断言、docs 当前正向要求。验证: 清单写入 plan §8.2.2
- [x] 0.3 锁定 baseline invariant：storage 只有 `language: multi`；`ResolveActive(featureKey, "en"|"zh-CN"|"fr")` fallback 到 `multi`；prompt 保留 `{{language}}` 输出语言指令；output schema 语言无关。验证: invariant 写入 plan §8.2.3，并在 Phase 2/4 测试覆盖

## Phase 1: spec 与 README 语义修订

- [x] 1.1 修订 `docs/spec/prompt-rubric-registry/spec.md` 到 v2.7，更新 D-1/D-2/D-6/D-7/D-12、§2.1、§4.1、§4.3、§6 C-1/C-6/C-13、§7 plan order，明确 baseline canonical `multi` only。验证: `sync-doc-index --check` 通过；spec/history/INDEX 版本一致
- [x] 1.2 修订 `config/prompts/README.md` 与 `config/rubrics/README.md`，删除“至少两个 language coordinates”要求，写明 exact-language override 只有语义差异时允许，rubric 默认语言无关。验证: README 语义已更新；后续 Phase 2/5 运行 prompt/rubric lint 读取新规则
- [x] 1.3 明确历史 completed plan evidence 不作为当前 truth；negative search gate 排除 `001`/`002` completed evidence 或在 plan §8 中标注历史范围。验证: plan §8.2.2 标注 completed `001`/`002` 为历史 evidence，Phase 4.3 继续记录 negative search 输出

## Phase 2: lint / tests 红灯调整

- [x] 2.1 先更新 `scripts/lint/prompt_lint_test.py`，证明 baseline 只需要 `multi`；重复 `en` variant 无 rationale 时失败；prompt body 保留 `{{language}}` 输出指令。随后修改 `scripts/lint/prompt_lint.py`。验证: focused pytest red/green；完整 `python3 -m pytest scripts/lint/prompt_lint_test.py -q` 在 Phase 5 通过
- [x] 2.2 更新/新增 rubric lint 测试，证明 baseline rubric 只需要 `multi`，不要求 `multi + en`；orphan/重复 override 失败。随后修改 `scripts/lint/rubric_lint.py`。验证: rubric lint focused test 或脚本负向 fixture red/green；完整 `python3 scripts/lint/rubric_lint.py` 在 Phase 5 通过
- [x] 2.3 更新 Go registry tests：`SnapshotSize` = 13；loader 要求每个 feature_key 有 `multi`；`ResolveActive(..., "en")` 与 unknown locale fallback 到 `multi`；OutputSchema 仍非空且共享。验证: focused `go test ./backend/internal/ai/registry -run 'Test(NewRegistryClientLoadsAllBaselines|LoadHappyPath|Resolve.*Language|ResolveActiveReturnsOutputSchema)' -count=1`
- [x] 2.4 更新 seed coverage tests，从 active `multi` truth source 反推期望坐标，扫描 seed migration 时拒绝 extra `en` rows。验证: focused `go test ./backend/internal/ai/registry -run TestSeedMigrationCoversBaselineFeatureKeys -count=1 -v` 先红后绿

## Phase 3: 删除 `en` truth source 与 seed rows

- [x] 3.1 删除 13 个 `config/prompts/*/v0.1.0.en.md`、13 个 `config/prompts/*/v0.1.0.en.yaml`、13 个 `config/rubrics/*/v0.1.0.en.yaml`。验证: `find config/prompts -name 'v0.1.0.en.*'` 与 `find config/rubrics -name 'v0.1.0.en.yaml'` 均无输出
- [x] 3.2 移除 baseline seed migrations 中 `language='en'` 的 prompt_versions / rubric_versions rows，保留 `multi` rows 和 idempotent `ON CONFLICT` 语义。验证: `python3 scripts/lint/migrations_lint.py` + focused seed coverage test 通过
- [x] 3.3 确认保留的 13 个 `multi` prompt / rubric 与 13 个 output schema 完整，prompt `template_hash` 不漂移。验证: `make lint-prompts` + `make lint-rubrics` 通过；3 个 `find ... | wc -l` 计数分别为 13

## Phase 4: loader / resolver / docs consumer 收敛

- [x] 4.1 修改 `backend/internal/ai/registry/loader.go` language parity rule：baseline `multi` 必须存在；任意 override 必须 prompt/rubric 成对；不要求 `en`。验证: loader focused tests 覆盖 missing multi / orphan override
- [x] 4.2 保持 resolver exact -> `multi` fallback，并更新 tests 证明 `en` 无 exact coordinate 时 fallback 到 `multi` 且 fallback counter 递增。验证: resolver focused tests 通过
- [x] 4.3 清理 current truth-source docs/README/lint/tests/migrations 中 `multi + en`、`26 coordinates`、`>=2 language coordinates`、`v0.1.0.en` 正向要求。验证: `rg -n 'multi \\+ en|26 coordinates|>=2 language|v0\\.1\\.0\\.en' config migrations` 无输出；code/scripts/spec/003 残留均为负向 fixture、删除 gate 或当前“不再要求”声明，分类记录在 plan §8.5；completed `001`/`002` 历史 evidence 可单独列出

## Phase 5: 验证、生命周期与收口

- [x] 5.1 运行 focused verification：prompt/rubric lint tests/scripts、registry focused/adjacent tests、seed coverage focused test、migration lint。验证: 命令输出记录到 plan §8
- [x] 5.2 在 dev-stack Postgres 可用时运行 `DATABASE_URL=postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable make migrate-check`；若环境不可用，记录 blocker 与 static gate 证据。验证: migrate-check pass 或明确 blocker
- [x] 5.3 运行 `validate_context.py`、`sync-doc-index --check`、`make docs-check`、`make lint`、`git diff --check`；如发现真实缺陷，按 `/bug-report` 建档；完成后将 plan/checklist 置为 `completed` 并同步 work journal/retrospective。验证: 全部 gate 通过，plans INDEX 显示 completed

## BDD-Gate

> **BDD 不适用**: 本 plan 收敛内部 registry truth source、lint、loader/resolver 和 seed migration 语义，不新增用户可见 UI、新 HTTP API 行为或端到端业务工作流。多语言用户体验由 runtime `language` 参数、prompt `{{language}}` 指令、各业务 caller 的 language provenance 和前端 i18n/BDD gate 承接。
>
> **替代验证 gate**:
>
> 1. prompt/rubric lint tests + `make lint-prompts` / `make lint-rubrics`
> 2. registry loader/resolver/seed coverage Go tests
> 3. migration lint + dev-stack `make migrate-check`
> 4. stale-contract negative search for `en` baseline requirements
> 5. context/index/docs/lint/whitespace gates
