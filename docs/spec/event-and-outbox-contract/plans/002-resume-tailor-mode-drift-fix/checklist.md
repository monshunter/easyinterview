# Event and Outbox Contract Resume Tailor Mode Drift Fix Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: `shared/events.yaml` 修订与 codegen drift 验证

- [x] 1.1 修订前跑 `make codegen-events && make codegen-check` 确认 baseline PASS（验证：exit 0, 无 diff）
- [x] 1.2 修订 `shared/events.yaml` `eventLocalEnums.ResumeTailorMode` 字面量为 `[gap_review, bullet_suggestions]`（验证：yaml lint + 手工检查）
- [x] 1.3 运行 `make codegen-events` 重新生成 Go/TS 类型（验证：`backend/internal/shared/events/` 与 `frontend/src/lib/events/` 出现新字面量、non-current literals 已清除）
- [x] 1.4 运行 `make codegen-check` PASS（验证：`git diff --exit-code` 0）
- [x] 1.5 新增或更新 `backend/internal/shared/events/resume_tailor_mode_test.go`，断言 `ResumeTailorModeGapReview` / `ResumeTailorModeBulletSuggestions` 存在且 non-current `Inline` / `Rewrite` / `Mirror` 不在允许集合（验证：Red → Green）
- [x] 1.6 更新 `frontend/src/lib/events/events.test.ts`，断言 `ResumeTailorMode` typed payload 接受 `gap_review` / `bullet_suggestions` 并拒绝 non-current `inline` / `rewrite` / `mirror`（验证：Red → Green）
- [x] 1.7 运行 `cd backend && go test ./internal/shared/events/...` 与 `pnpm --filter @easyinterview/frontend test src/lib/events/events.test.ts` PASS（验证：相关 type-narrowing 测试通过）

## Phase 2: Baseline manifest 同步

- [x] 2.1 修订 `shared/events/baseline/events.v1.json` 中 `eventLocalEnums.ResumeTailorMode` 字面量集合（验证：JSON lint + 手工检查）
- [x] 2.2 运行 `make lint-events` PASS（验证：无 breaking 报警；如分类为 breaking，加入 `ResumeTailorMode drift-fix` 白名单配置）
- [x] 2.3 跨仓库 baseline `make codegen-events && make lint-events` 同时 PASS（验证：双重门禁）

## Phase 3: 跨仓库 grep negative search

- [x] 3.1 运行事件契约 artifact 精准负向搜索：`git grep -nE 'ResumeTailorMode(Inline|Rewrite|Mirror)|"(inline|rewrite|mirror)"' -- shared/events.yaml shared/events/refs/ResumeTailorMode.json shared/events/baseline/events.v1.json shared/events/schemas/resume.tailor.completed.v1.json backend/internal/shared/events frontend/src/lib/events openapi/openapi.yaml openapi/fixtures`（验证：0 命中；history / 本 plan diff 表述不纳入 executable/generated/source truth gate）
- [x] 3.2 修订 `docs/spec/event-and-outbox-contract/spec.md` §3.1.4 `resume.tailor.completed.mode` 列描述，移除"声明阶段 → 落地阶段"措辞，简化为 `[gap_review, bullet_suggestions]`（验证：手工检查）
- [x] 3.3 B3 spec.md 2.3 → 2.4，history.md 追加 2.4 行（关联本 plan，标记 "落地阶段"）（验证：`sync-doc-index --check`）

## Phase 4: 验收与下游同步

- [x] 4.1 运行 §3 全部替代验证 gate PASS：`make codegen-events && make codegen-check && make lint-events` + grep + `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml` + `sync-doc-index --check`
- [x] 4.2 在 `openapi-v1-contract/004-resume-additive-coverage` plan checklist 中追加 "B3 D-14 `ResumeTailorMode` 漂移修复已落地" 引用（验证：cross-plan 引用 commit）
- [x] 4.3 在 `db-migrations-baseline/002-flat-resume-migration` plan checklist 中追加同步引用（验证：cross-plan 引用 commit）
- [x] 4.4 修订 `docs/spec/INDEX.md` event-and-outbox-contract 版本与日期 2.3 → 2.4（验证：`sync-doc-index --check`）
- [x] 4.5 B4 flat Resume schema 已由 [db-migrations-baseline/002](../../../db-migrations-baseline/plans/002-flat-resume-migration/plan.md) 落地；events `ResumeTailorMode` 与 task-output suggestion owner chain 使用 `gap_review` / `bullet_suggestions`。
