# Backend Profile Candidate Profile and Experience Cards 交付复盘报告

> **日期**: 2026-05-21
> **审查人**: Claude Opus 4.7 (1M context)

## 1 复盘范围与成功证据

- 交付范围：[`backend-profile/001-candidate-profile-and-experience-cards`](../spec/backend-profile/plans/001-candidate-profile-and-experience-cards/plan.md) Phase 1-5 完整实施。
- 验证证据：
  - `go test ./...` 全 PASS（含 handler unit、service unit、conventions parity）。
  - `DATABASE_URL=... go test -tags=integration ./internal/profile/store/... -count=1` PASS（命中本地 dev-stack Postgres）。
  - `DATABASE_URL=... go test ./cmd/api -run TestProfileHTTPScenario -count=1` PASS（real `auth.SessionMiddleware` + IK middleware + Postgres）。
  - 3 个 BDD 场景全 PASS：`p0-091-candidate-profile-seed-and-patch` / `p0-092-experience-cards-crud-with-ik` / `p0-093-profile-privacy-delete-lifecycle`（setup → trigger → verify → cleanup 四段）。
  - `make lint-openapi` / `make validate-fixtures` / `make openapi-diff`（additive only）/ `make docs-check` 全 PASS；`sync-doc-index --check` 0 drift。
  - C-14 negative grep `mistake|growth|drill|experiences|star` 在 `backend/internal/profile/` 0 命中。
  - 提交：`093e6c6 feat(backend-profile): land candidate profile and experience cards baseline`，分支 `feat/backend-profile-001-candidate-profile-and-experience-cards-0521`。

## 2 会话中的主要阻点/痛点

- **B1 / B2 cross-owner additive 未在 plan 中列出**：
  - **证据**：spec D-8 要求 cross-user 404 返回 `RESOURCE_NOT_FOUND`，但 `shared/conventions.yaml` 与 B1 D-5 前缀字典从未注册该 code；BDD-Plan E2E.P0.091 A1 期望 seed 后 `headline / yearsOfExperience / currentRole / region` 返回 JSON null，但 generated `CandidateProfile` 字段为非指针非可空。plan §3.2 仅列出 IK additive，未提及 `RESOURCE_NOT_FOUND` 与 `CandidateProfile` nullable 字段两次 additive。
  - **影响**：实施进入 Phase 1 后才暴露，迫使在 cross-owner 边界（B1 / B2）做计划外修订并通过 `AskUserQuestion` 由用户确认；如果按默认 plan 直接编码，将在 byte parity / 单元测试阶段集中爆炸。

- **BDD 场景 ID 与现网占用冲突**：
  - **证据**：plan / bdd-plan / bdd-checklist 全部使用 `E2E.P0.081` / `082` / `083`，但 `test/scenarios/e2e/INDEX.md` 中这三个 ID 已被 `frontend-resume-workshop/003` 占用；冲突在执行 BDD 阶段直接被 `INDEX.md` grep 命中。
  - **影响**：需要在实施末段统一改名为 P0.091 / 092 / 093，并跨 plan / bdd-plan / bdd-checklist / 场景目录 / 脚本 / 5 行 INDEX 文本同步重命名；如果在执行前未与 INDEX 对账，可能出现 owner 看到两个 P0.081 的混乱状态。

- **预先存在的 `make codegen-events-check` failure 干扰 gate 收口**：
  - **证据**：`make codegen-check` 报 `FAIL: frontend/src/api/devMockClient.ts: naked event/job literal 'debrief_generate'`；`git stash` 验证在干净 HEAD 上也复现，与本 plan 无关。
  - **影响**：plan §3 替代验证 gate 把 `make codegen-check` 整体列为收口项，没有声明 events 子项的已知缺口；实施时必须额外区分本 plan 引入的 OpenAPI / B1 portion（PASS）与遗留 frontend events failure（pre-existing），才能正确签收。

- **`current_role` PG 关键字未在 spec/plan 提示**：
  - **证据**：`migrations/000001_create_baseline.up.sql` 用 `"current_role"` 加引号建列，但 plan 未提示；初版 store SQL 直接写 `current_role`，integration test 立即报 `syntax error at or near "current_role"`。
  - **影响**：一次性 SQL 调试，影响有限但属于"plan 未把 cross-owner schema 引号约定显式化"的典型小坑。

## 3 根因归类

- **B1 / B2 cross-owner additive 缺漏（最重要）**
  - **类别**：spec-plan。Plan §3.2 只展开 IK additive，spec D-8 与 BDD A1 同时引入了 B1 + B2 nullable 两个未声明依赖；L1 plan-review 应该把所有 cross-owner contract 改动一次性列入 §3.2 / §3.3。

- **BDD 场景 ID 冲突**
  - **类别**：spec-plan + skill。Plan / bdd-plan / bdd-checklist 自动从一个起始号开始数，但 e2e INDEX 是真实占用台账；`/plan-review` 与 `/design` 应该在生成 BDD 编号前对 e2e INDEX 做 reserve gate。

- **`make codegen-check` 含已知失败子项未在 plan 中标注**
  - **类别**：spec-plan + README。Plan §3 替代验证 gate 抄写 root Makefile target，未区分本 plan 真正应该携带的子部分（OpenAPI / B1 / events）；`backend/README.md` 或 plan §3 应该给"复合 gate 拆分"的最小颗粒度。

- **`current_role` 等 PG 保留字引号约定未在 store/data model 文档中显式**
  - **类别**：no repo change needed（一次性执行误差）。Migration 文件本身已有正确做法；只需要在写新 store 时直接看 migration。

## 4 对流程资产的改进建议

- 建议在 `/plan-review` 的 cross-owner additive 检查列表中补充 "non-trivial response field semantics（nullable / oneOf / discriminator）" 与 "新错误码"两项强制审查点，并把 plan §3.2 的"B2 cross-owner additive"段落要求显式列出每一项 schema 改动（field-level，不只是 endpoint-level）。
  - **落点**：`/plan-review` skill + 可选地 `AGENTS.md` §2.1.3 操作矩阵补充字段
  - **优先级**：high

- 在 `/design` / `/plan-review` 生成 BDD scenario ID 前增加"对 `test/scenarios/*/INDEX.md` 已占用 ID 的 reserve 检查"，避免新 plan 重复占用历史已分配的 P0.* 段。
  - **落点**：`/design` skill 与 `/plan-review` skill（共享 helper 脚本）
  - **优先级**：high

- 把 plan §3 "替代验证 gate" 模板改为按子目标（`make lint-openapi` / `make validate-fixtures` / `make openapi-diff` / `make codegen-conventions` / `make codegen-events` / `make codegen-openapi`）分行列出，并显式给出"已知 pre-existing failure"专门子节，让实施者能在 plan-review 阶段就识别哪些 gate 是本 plan 需要落地、哪些是遗留缺口。
  - **落点**：`/design` skill 模板 + `docs/spec/TEMPLATES.md`（如果存在 plan 模板）
  - **优先级**：medium

- 在 `backend/README.md` 增加一段"PG reserved keyword 列约定"，列出 `current_role` 等 baseline migration 用 `"..."` 引号的列名，提醒新 store 实现引用列名时同步加引号。
  - **落点**：`backend/README.md`
  - **优先级**：low

## 5 建议优先级与后续动作

- 下一轮最值得实施的改进：
  - **`/plan-review` skill 字段级 additive 审查**（high）：把"新错误码 + nullable 字段"作为 cross-owner additive 必查项，避免下次实施再次出现 plan-review 通过、实施期才发现 B1 / B2 缺口的事故。
  - **BDD scenario ID reserve gate**（high）：阻止 plan 用已被占用的 P0.* ID，省掉实施末段的批量重命名。
- 可以延后处理的优化项：
  - `make codegen-check` 子目标拆分与 events failure 已知缺口的台账（medium）。
  - `backend/README.md` PG 保留字列约定补充（low）。

## 6 关联资产

- 计划：[plan](../spec/backend-profile/plans/001-candidate-profile-and-experience-cards/plan.md) / [checklist](../spec/backend-profile/plans/001-candidate-profile-and-experience-cards/checklist.md) / [bdd-plan](../spec/backend-profile/plans/001-candidate-profile-and-experience-cards/bdd-plan.md) / [bdd-checklist](../spec/backend-profile/plans/001-candidate-profile-and-experience-cards/bdd-checklist.md)
- Owner spec：[backend-profile/spec.md](../spec/backend-profile/spec.md)（升至 1.2）
- Cross-owner additive：[`shared-conventions-codified` 1.20](../spec/shared-conventions-codified/spec.md) + [`openapi-v1-contract` 1.26](../spec/openapi-v1-contract/spec.md)
- 路线图：[`engineering-roadmap` 3.18](../spec/engineering-roadmap/spec.md)
- BDD 场景：`test/scenarios/e2e/p0-091-candidate-profile-seed-and-patch/` / `p0-092-experience-cards-crud-with-ik/` / `p0-093-profile-privacy-delete-lifecycle/`
- 提交：`093e6c6 feat(backend-profile): land candidate profile and experience cards baseline`
