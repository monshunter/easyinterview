# 001 BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-08

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.018 Workspace 默认渲染 + Plan Switcher / Resume Picker

- [ ] 创建场景目录 `test/scenarios/e2e/p0-018-workspace-default-render/`，含 `README.md`（§6 baseline + §7 离线限制）+ `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`getTargetJob.json` 至少 2 个 variant（`default` + `with-rounds`）；`getResume.json` `default`；`listTargetJobs.json` 至少 3 个 variant（`default` 多 plan + `single-plan` + `prototype-baseline`）；`getPracticePlan.json` `default`；按 `mock-contract-suite` 规则配置；通过 `make validate-fixtures`
- [ ] 实现 `scripts/setup.sh`（含三种 fixture variant 切换入口 + signed-in 状态）/ `scripts/trigger.sh`（按主路径运行 + 打开/关闭两 modal + 切换 plan）/ `scripts/verify.sh`（断言 ≥ 20 个 testid 命中、TopBar 高亮、两 Modal a11y 行为、`listResumes` 调用 0、zh/en + warm/light → dark → customAccent 切换、mobile 不溢出）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：spec 调用栈 + variant 切换日志 + a11y focus 路径日志 + 截图（baseline + 当前）+ retired-testid grep 0 命中
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.018 行（关联需求 `frontend-workspace-and-practice C-2, C-7, C-8, C-9`，状态 Ready，automated）

## E2E.P0.019 Workspace context loading + 空态 + getPracticePlan refresh

- [ ] 创建场景目录 `test/scenarios/e2e/p0-019-workspace-context-loading/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`getTargetJob.json` 至少 3 个 variant（`with-rounds` 成功 / `not-found` / `5xx`）；`getResume.json` 至少 2 个 variant（`default` / `not-found`）；`getPracticePlan.json` 至少 4 个 variant（`default(ready)` / `archived` / `not-found` / `default(ready)` 第二份）
- [ ] 实现 `scripts/setup.sh`（含 4 子场景 fixture 切换入口）/ `scripts/trigger.sh`（按 A/B/C/D 四子场景 + E 通用错误路径运行）/ `scripts/verify.sh`（断言 A 主路径完整渲染、B `WorkspaceEmptyState` testid 命中 + CTA → home + textarea focus、C `WorkspaceMissingResumeState` testid 命中 + CTA → resume_versions、D plan refresh 状态 transition、E header 退化为只读 + 重试按钮、JD 原文 0 命中）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS（A/B/C/D/E 共 5 子用例）
- [ ] 记录验证证据：mockTransport 调用日志 + `InterviewContext` reducer 状态日志 + 隐私反查日志 + 5xx 路径截图
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.019 行（关联需求 `frontend-workspace-and-practice C-2, C-3, C-8, C-9`，状态 Ready，automated）

## E2E.P0.020 立即面试主路径 + 错误重试 + 未登录恢复

- [ ] 创建场景目录 `test/scenarios/e2e/p0-020-workspace-start-practice/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`createPracticePlan.json` 至少 3 个 variant（`default` 成功 / `missing-resume` 422 / `validation-422` 通用 422）；`startPracticeSession.json` 至少 2 个 variant（`default` 成功 / `ai-timeout-502`）；`getPracticePlan.json` `default(ready)` + `archived`；signed-in / signed-out 两种状态切换入口
- [ ] 实现 `scripts/setup.sh`（含 signed-in/out 切换 + plan 存在/不存在切换）/ `scripts/trigger.sh`（按 A1/A2/A3/B1/C1 五子场景运行）/ `scripts/verify.sh`（断言：A1 createPracticePlan + startPracticeSession 双步调用、body schema 完整、`Idempotency-Key` 双键稳定、nav practice 携带 InterviewContext + `PracticeDisplayContext` 字段、practice route 仍 PlaceholderScreen、`questionText` 不在 DOM；A2 422 inline 错误 + focus `更换简历`、不进入 startPracticeSession；A3 502 重试 `Idempotency-Key` 复用 + 3 次失败 fallback CTA；B1 跳过 createPracticePlan；C1 `requestAuth` 触发 + `auth_login` 路由携带 `pendingRoute=workspace` / `pendingType=start_practice` / `autoStartPractice=1` + verify 后自动 startPractice + nav practice、`pendingAction.params` 不含敏感字段；workspace 不产出 `debrief_replay`；StrictMode 下 generated client 调用次数 ≤ 2）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS（A1/A2/A3/B1/C1 共 5 子用例）
- [ ] 记录验证证据：generated client 请求 body 截取 + `Idempotency-Key` 双键日志 + `pendingAction.params` 字段集合断言 + nav practice params 截取 + retry 路径流
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.020 行（关联需求 `frontend-workspace-and-practice C-1, C-3, C-12`，状态 Ready，automated）

## E2E.P0.021 Workspace handoff + 隐私红线 + 旧入口反向 grep

- [ ] 创建场景目录 `test/scenarios/e2e/p0-021-workspace-handoff/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`getTargetJob.json` `with-rounds` 不新增未声明 `recentSessions[]` extension；`getResume.json` `default`；`getPracticePlan.json` `default(ready)`；signed-in 状态；history 验证固定为 `EmptyHistory` / disabled placeholder
- [ ] 实现 `scripts/setup.sh`（含 fixture variant 切换）/ `scripts/trigger.sh`（按 A/B 两子场景运行 + 隐私反查 D + 负向反向 grep E + regression rerun F）/ `scripts/verify.sh`（断言：A `nav("company_intel", { targetJobId, jdId })` + `getCompanyIntel` 调用 0；B history 区域为 `EmptyHistory` / disabled placeholder，点击不触发 `nav("report", ...)`，不读取 `TargetJob.recentSessions`，`getFeedbackReport` 调用 0；D JD 原文/简历正文/questionText/answerText/hintText/promptHash 在 console/URL/localStorage/telemetry 0 命中；E 旧 testid + 旧 route alias + prototype data import + `listResumes` + `getCompanyIntel` grep 0 命中；F D1+D2+D3 已存在 `P0.001/002/004/005/006` regression PASS，home plan `P0.014/015/016/017` 仅在场景资产存在且 INDEX Ready 时执行）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS（A/B/D/E/F 共 5 子用例）
- [ ] 记录验证证据：handoff nav params 截取 + retired-testid / route / data import grep 0 命中日志 + 隐私反查日志 + regression scenario PASS marker
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.021 行（关联需求 `frontend-workspace-and-practice C-7, C-9, C-10, C-12`，状态 Ready，automated）

## 整体 Regression（Phase 6 收口）

- [ ] D1+D2+D3 Regression 重跑：`E2E.P0.001 / 002 / 004 / 005 / 006` 全部 setup→trigger→verify→cleanup PASS；home plan `E2E.P0.014 / 015 / 016 / 017` 仅在场景资产存在且 INDEX Ready 时执行，否则记录为“上游 active plan 条件 gate，当前不适用”
- [ ] `pnpm --filter @easyinterview/frontend test` 全量 Vitest PASS（含本 plan 新增测试文件）
- [ ] `pnpm --filter @easyinterview/frontend test:pixel-parity` 在 D2/D3 + home plan 现有基础上累加 workspace 新增 spec，总数全 PASS
- [ ] `pnpm --filter @easyinterview/frontend typecheck` + `pnpm --filter @easyinterview/frontend build` + `make build` 全 PASS
- [ ] `make docs-check` zero drift；`/sync-doc-index --fix-index` post-fix zero drift；`check_md_links` 双 OK
