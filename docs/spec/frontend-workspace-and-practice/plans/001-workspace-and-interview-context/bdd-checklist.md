# 001 BDD Checklist

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-23

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

> L2 code review fix pass 已在 2026-05-09 复核并加固 P0.018-P0.021 场景脚本。证据包括对应 `setup -> trigger -> verify -> cleanup` 日志、full Vitest、full pixel parity、frontend build 与 runtime negative grep。

## E2E.P0.018 Workspace 默认渲染 + Plan Switcher / Resume Picker

- [x] 创建场景目录 `test/scenarios/e2e/p0-018-workspace-default-render/`，含 `README.md`（§6 baseline + §7 离线限制）+ `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 fixture variant：`getTargetJob.json` 至少 2 个 variant（`default` + `with-rounds`）；`getResume.json` `default`；`listTargetJobs.json` 至少 3 个 variant（`default` 多 plan + `single-plan` + `prototype-baseline`）；`getPracticePlan.json` `default`；按 `mock-contract-suite` 规则配置；通过 fixture/schema 验证
- [x] 实现 `scripts/setup.sh`（fixture variant + signed-in 状态）/ `scripts/trigger.sh`（运行 App route hydration、Workspace modal integration、Modal a11y 等覆盖）/ `scripts/verify.sh`（断言 ≥ 20 个 runtime testid 命中、modal tests 已运行、retired testid / `listResumes` runtime negative grep 0 命中）/ `scripts/cleanup.sh`
- [x] 执行 `setup -> trigger -> verify -> cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-018-workspace-default-render/trigger.log` + verify 输出 + full pixel parity workspace desktop/mobile baseline + retired-testid grep 0 命中
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.018 行（关联需求 `frontend-workspace-and-practice C-2, C-7, C-8, C-9`，状态 Ready，automated）

## E2E.P0.019 Workspace context loading + 空态 + getPracticePlan refresh

- [x] 创建场景目录 `test/scenarios/e2e/p0-019-workspace-context-loading/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 fixture variant：`getTargetJob.json` 至少 3 个 variant（`with-rounds` 成功 / `not-found` / `5xx`）；`getResume.json` 至少 2 个 variant（`default` / `not-found`）；`getPracticePlan.json` 至少 4 个 variant（`default(ready)` / `archived` / `not-found` / `default(ready)` 第二份）
- [x] 实现 `scripts/setup.sh`（fixture 切换入口）/ `scripts/trigger.sh`（运行 App route hydration、target/resume/practice plan hook 覆盖）/ `scripts/verify.sh`（断言 route hydration 与 getPracticePlan refresh 测试已运行、runtime 隐私负向 grep 0 命中）/ `scripts/cleanup.sh`
- [x] 执行 `setup -> trigger -> verify -> cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-019-workspace-context-loading/trigger.log` + verify 输出 + `InterviewContext` hydration / reducer 测试断言 + 隐私反查日志
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.019 行（关联需求 `frontend-workspace-and-practice C-2, C-3, C-8, C-9`，状态 Ready，automated）

## E2E.P0.020 立即面试主路径 + 错误重试 + 未登录恢复

- [x] 创建场景目录 `test/scenarios/e2e/p0-020-workspace-start-practice/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 fixture variant：`createPracticePlan.json` 至少 3 个 variant（`default` 成功 / `missing-resume` 422 / `validation-422` 通用 422）；`startPracticeSession.json` 至少 2 个 variant（`default` 成功 / `ai-timeout-502`）；`getPracticePlan.json` `default(ready)` + `archived`；signed-in / signed-out 两种状态切换入口
- [x] 实现 `scripts/setup.sh`（signed-in/out + plan 存在/不存在切换）/ `scripts/trigger.sh`（运行 WorkspaceStartPractice 与 WorkspaceAuthGate 覆盖）/ `scripts/verify.sh`（断言双步启动、plan refresh fallback、pendingAction auto-start、隐私负向 grep 与敏感字段约束）/ `scripts/cleanup.sh`
- [x] 执行 `setup -> trigger -> verify -> cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-020-workspace-start-practice/trigger.log` + verify 输出 + generated client body/header 断言 + `pendingAction.params` 字段集合断言 + retry/idempotency 测试断言
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.020 行（关联需求 `frontend-workspace-and-practice C-1, C-3, C-12`，状态 Ready，automated）

## E2E.P0.021 Workspace handoff + 隐私红线 + 旧入口反向 grep

- [x] 创建场景目录 `test/scenarios/e2e/p0-021-workspace-handoff/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 fixture variant：`getTargetJob.json` `with-rounds` 不新增未声明 `recentSessions[]` extension；`getResume.json` `default`；`getPracticePlan.json` `default(ready)`；signed-in 状态；history 验证固定为 `EmptyHistory` / disabled placeholder
- [x] 实现 `scripts/setup.sh`（fixture variant 切换）/ `scripts/trigger.sh`（运行 WorkspaceHandoff 与相关 workspace regression 覆盖）/ `scripts/verify.sh`（断言 handoff、history placeholder、runtime 敏感字段、旧 testid / route alias / prototype import / forbidden generated calls grep 0 命中）/ `scripts/cleanup.sh`
- [x] 执行 `setup -> trigger -> verify -> cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-021-workspace-handoff/trigger.log` + verify 输出 + handoff nav params 断言 + runtime negative grep 0 命中日志
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.021 行（关联需求 `frontend-workspace-and-practice C-7, C-9, C-10, C-12`，状态 Ready，automated）

## 整体 Regression（Phase 6 收口）

- [x] D1+D2+D3 Regression 当前等价 gate：`E2E.P0.001 / 002 / 004 / 005 / 006` 由 full Vitest + full pixel parity 覆盖；workspace `E2E.P0.018 / 019 / 020 / 021` 已全部 `setup -> trigger -> verify -> cleanup` PASS；home plan `E2E.P0.014 / 015 / 016 / 017` 属上游 active plan 条件 gate，当前不作为本 plan completion blocker
- [x] `pnpm --filter @easyinterview/frontend test` 全量 Vitest PASS（66 files, 423 tests，含本 plan 新增测试文件）
- [x] `pnpm --filter @easyinterview/frontend test:pixel-parity` 在 D2/D3 + home plan 现有基础上累加 workspace 新增 spec，总数 96/96 PASS
- [x] `pnpm --filter @easyinterview/frontend build` PASS（包含 `tsc --noEmit` + `vite build`）
- [x] 文档与索引收口：本 checklist、主 checklist 与 plans INDEX 已同步；`make docs-check` / `/sync-doc-index` 作为 post-fix drift gate 执行
- [x] 2026-05-23 real-backend gate：P0.018-P0.021 trigger 前置 `frontendOwners.realApiMode.test.ts`，verify 检查 `VITE_EI_API_MODE=real` / `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1` / 测试文件 marker。 <!-- evidence: focused real-mode Vitest PASS; scenario scripts updated -->
