# Resume Flatten Follow-up Review 交付复盘报告

> **日期**: 2026-06-15
> **审查人**: Codex (GPT-5)

## 1 复盘范围与成功证据

本次交付范围是修复 reviewer 对 resume flatten 后续 patch 提出的 4 条 P2 问题：flat resume 创建 display name 初始化、duplicate explicit empty profile 语义、`000015` rollback table shape，以及 retired contracts 后 scripts/lint pytest suite 漂移。

已闭环的范围：

- `CreateWithParseJob` 创建 resume 时同步写入 `display_name`。
- `duplicateResume` 从 handler 到 store 保留 `structuredProfile` 字段存在语义，显式 `{}` 不再回退复制源 profile。
- `000015_resume_flatten.down.sql` 回滚后恢复 post-000014 的 `resume_tailor_runs` provenance columns。
- `scripts/lint` suite 与当前 conventions/OpenAPI/prompt/rubric contract 对齐，并增加 retired JD-Match 负向断言。

通过证据：

- `python3 -m pytest scripts/lint -q`
- `go test ./backend/internal/resume/store -run 'TestCreateWithParseJobInsertsResumeAndJobAtomically|TestDuplicateResumePersistsExplicitEmptyProfile' -count=1`
- `go test ./backend/internal/resume -run 'TestDuplicateResumeAllocatesNewIDAndAppliesProfile|TestDuplicateResumePreservesExplicitEmptyProfile' -count=1`
- `go test ./backend/internal/resume/handler -run 'TestDuplicateResumeReturns201|TestDuplicateResumePreservesExplicitEmptyStructuredProfile|TestDuplicateResumeAllowsEmptyBody' -count=1`
- `go test ./backend/internal/migrations -run TestResumeFlattenMigrationContract -count=1`
- `go test ./backend/internal/resume/... ./backend/internal/migrations -count=1`
- `go test ./backend/cmd/api -run 'Test.*Resume|Test.*Tailor' -count=1`
- `python3 scripts/lint/migrations_lint.py --repo-root .`
- `make validate-fixtures`
- `make docs-check`
- `make codegen-check`
- `make test`
- `make build`
- `make lint`
- `git diff --check`
- Retired expectation negative search over scripts/lint, shared conventions, OpenAPI, prompt/rubric config, generated API artifacts, backend and frontend generated clients.

## 2 会话中的主要阻点/痛点

- **Required display field did not have create-path coverage**
  - **证据**：`display_name` existed as an API/UI field, but the register insert path only populated `title`.
  - **影响**：new resumes could render blank headers, toasts, or duplicate names until manually edited.

- **JSON field presence semantics were lost across layers**
  - **证据**：handler decoded `structuredProfile`, but service used `len(map) > 0`, making explicit `{}` indistinguishable from omitted field.
  - **影响**：a user-cleared duplicate draft could silently restore the source profile.

- **Rollback contract test did not assert the target version schema**
  - **证据**：`000015` down migration recreated `resume_tailor_runs` without `language`, `feature_flag`, and `data_source_version`.
  - **影响**：one-step rollback to migration 14 would leave version-14 code reading missing columns.

- **Retired contracts were removed from source inventory but not from lint tests**
  - **证据**：initial `python3 -m pytest scripts/lint -q` failed in conventions, OpenAPI diff, prompt lint, and rubric lint tests.
  - **影响**：repo verification could not stay green after the intended contract retirements.

## 3 根因归类

- **Create / readback parity gap**
  - **类别**：spec-plan / test
  - Required API/UI fields need create-path and readback assertions, not only DTO/schema assertions.

- **Field-presence contract gap**
  - **类别**：spec-plan / no repo change needed
  - Explicit empty object is meaningful for resume drafts, so handler/service/store contracts must carry presence separately from map length.

- **Down migration target-shape gap**
  - **类别**：test
  - Migration rollback tests asserted that a down file existed, but did not assert the post-target-version column shape.

- **Retired-contract fixture gap**
  - **类别**：test / README
  - Removing a contract surface requires updating both positive inventory and tests that encode old positive examples.

## 4 对流程资产的改进建议

- **Resume create/update/duplicate review checklist should include required-field readback**
  - **落点**：resume owner plan gate or backend resume review checklist
  - **优先级**：medium

- **JSON patch/copy contracts should require field-presence tests**
  - **落点**：backend service test conventions
  - **优先级**：medium

- **Down migration contract should assert rollback target schema for recreated tables**
  - **落点**：backend migration test conventions
  - **优先级**：high

- **Contract retirement should include positive fixture removal and negative lint assertions**
  - **落点**：scripts/lint README or owner plan gate
  - **优先级**：high

## 5 建议优先级与后续动作

下一步最高价值动作是给 D-20 resume flatten owner plan 增加一个 targeted rollback/readback hardening gate：覆盖 create/display readback、duplicate explicit empty object、down migration target schema、retired contract negative search。该 gate 可以防止同类问题在后续 resume flatten follow-up 中再次靠 reviewer 才暴露。

备选动作是单独沉淀一个 scripts/lint retirement checklist，先保证未来删除 contract surfaces 时 lint tests 与 source inventory 原子更新。
