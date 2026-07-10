# 001 — Report Screen and Generating Handoff BDD Plan

> **版本**: 1.14
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md) / [checklist](./checklist.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 0 BDD 框架与编号

本 plan 的 4 个 BDD 场景保留 `E2E.P0.xxx` 编号与 Given / When / Then 语义；001 本次代码交付的可执行入口落在 vitest + jsdom 主套件 + Playwright pixel parity，同步派生 4 个 `test/scenarios/e2e/p0-NNN-*` 目录，trigger 跑对应 Vitest / Playwright 套件，verify 反查 testid / nav payload / 负向 grep / i18n 完整性。

- 套件: `e2e`
- 阶段: `P0`
- 已占用编号现状（[`test/scenarios/e2e/INDEX.md`](../../../../../test/scenarios/e2e/INDEX.md)）：`001-006`, `010-047`；backend-practice/003 已通过 Go HTTP scenario 预留 `048-051`（hint 四场景）；backend-review/001 预留 `052-055`（report 四场景，已在 backend-review/001 plan 内固化）。本 plan 在空闲号段 `056-059` 中分配 4 个场景
- 编号分配: `E2E.P0.056` / `E2E.P0.057` / `E2E.P0.058` / `E2E.P0.059`
- 执行入口: vitest + jsdom 主套件 `pnpm --filter @easyinterview/frontend test` + Playwright `pnpm --filter @easyinterview/frontend test:pixel-parity` + 4 个 scenario 目录的 verify.sh
- 外部 shell 场景资产: 001 派生 `test/scenarios/e2e/p0-{056,057,058,059}-*` 本地 runner 目录与 backend-practice/002 + frontend-workspace-and-practice/002 同模式

每个场景的执行证据在 [bdd-checklist](./bdd-checklist.md) 跟踪；本文件只记录场景的 Given / When / Then 与覆盖范围，不出现执行 checkbox。

## 1 场景矩阵

| 场景 ID | 名称 | 类别 | 关联 Plan Phase | 关联 spec AC / D |
|---------|------|------|----------------|-------------------|
| `E2E.P0.056` | 五个 focused owner tests：preflight + poll hook + GeneratingScreen + ReportScreen + DetailSurface；real-mode 配置与 fixture-backed 行为分层 | primary contract composition | Phase 1 + Phase 2 + Phase 3 + Phase 10 | C-1, C-2, C-5, C-8, C-11 |
| `E2E.P0.057` | 复练 CTA 路径 A retry_current_round + 路径 B next_round direct-start + 未登录 report auth gate + replay pendingAction round-trip + payload 完整性 | alternate | Phase 4+9 | C-9, C-10, D-5 |
| `E2E.P0.058` | 六个 focused failure contracts：preflight + failure/missing components + report hook/route + poll hook | failure contract composition | Phase 1 + Phase 2 + Phase 11 | C-3, C-4, C-6, C-7, D-6 |
| `E2E.P0.059` | Playwright desktop/mobile DOM、状态、geometry、overflow 与非空内存截图 + i18n zh/en 完整性 + 旧 reportLayout / 5 档 readiness / 报告时间线 / mistakes / drill / growth_center 负向 grep | regression + UX | Phase 5+8+12 | C-12, C-13, D-7, D-10, D-11, D-12, D-13 |

## 2 Phase 1 + 2 + 3 + 4 — GeneratingScreen → ReportScreen happy path

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.056` | Generating/report focused owner contract composition | shared real-mode bootstrap contract；五个独立 Vitest 文件各自使用 deterministic client/fixture：preflight、poll hook、GeneratingScreen、ReportScreen、DetailSurface | trigger 运行五个 focused 文件；verify 读取每个文件的 pass marker，并执行 implementation testid inventory、out-of-scope lint 与 listTargetJobReports 负向搜索 | preflight 验证共享 schema/fixture/owner 合同；poll hook 覆盖状态、backoff、ready/failed callback、取消与 read-only header；GeneratingScreen 覆盖 DOM、缺参、i18n 和 route handoff；ReportScreen 覆盖 dashboard/error/missing、getTargetJob/getResume labels、read-only header 与 raw body avoidance；DetailSurface 覆盖五 tab 和本地 replay marker。real-mode gate只证明 production client 配置，这组 focused tests 不声明单一 browser/live-backend journey或固定跨文件请求序列 | `test/scenarios/e2e/p0-056-generating-to-report-happy-path/` |

## 3 Phase 4 — 复练 CTA 路径 A + 路径 B

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.057` | 复练 CTA 路径 A retry_current_round + 路径 B next_round | （A）已登录用户 + ready ReportDashboard，report 含 retryFocusTurnIds/issues；（B）同一用户选择下一轮；（C）未登录用户直接进入 report route；（D）独立 `replay_practice` pendingAction fixture | （A）点击唯一 Header「复练当前轮」；（B）点击唯一 Header「进入下一轮」；（C）观察 App auth gate；（D）encode/decode pendingAction | （A）generated `createPracticePlan` / `startPracticeSession` 创建 retry_current_round fresh session并直接进入 practice，sourceReportId/sourceSessionId/replayItems/evidenceGaps 完整；（B）创建 next_round fresh plan/session，roundId 按 canonical ladder 轮转并直接进入 practice；（C）CTA 挂载前进入 auth_login，不读取 report、不启动 session；（D）pendingAction 保留 report route 与安全上下文，鉴权后回 report；payload 无 raw answer/question/hint/prompt/model 内容，CTA 不重复读取 report 或调用 listTargetJobReports | `test/scenarios/e2e/p0-057-replay-cta-paths-a-and-b/` |

## 4 Phase 2 + 11 — Focused failure contracts

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.058` | Report/poll focused failure contract composition | shared real-mode bootstrap contract；deterministic component/hook clients for failed status, missing ids, HTTP 404, HTTP 5xx and persistent generating | trigger 运行 owner preflight、ReportFailureState、ReportMissingSessionState、useFeedbackReport、ReportScreen、useReportGenerationPoll 六个 focused 文件；verify 要求每个 marker 和 typed i18n key | failure component 覆盖 AI_* matrix、REPORT_NOT_FOUND 独立 copy、retry/back handlers 与 UNKNOWN fallback；missing component 覆盖 session/report variants；report hook/route 覆盖 ready/404/5xx/refresh 与 rendered not-found state；poll hook覆盖 failed/404/timeout/backoff/cancel。该 runner 不挂载 GeneratingScreen，不声明 timeout UI、重复 timeout fallback、live-backend sequence 或宽泛 route/storage/telemetry 隐私证据 | `test/scenarios/e2e/p0-058-report-failure-and-missing-session/` |

## 5 Phase 5 — Pixel parity + i18n + out-of-scope negative

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| `E2E.P0.059` | Playwright browser smoke + i18n + 范围外输入负向 | real-mode contract gate；owner/browser preflight；i18n 与 out-of-scope tests；fixture-backed generating/report routes；frontend build；Playwright web server | （A）运行 owner/browser preflight、i18n tests 与 scoped lint；（B）构建 frontend；（C）Playwright 依次覆盖 generating desktop 主屏、缺 reportId、mobile overflow，以及 report desktop dashboard、缺 sessionId、failed state、mobile overflow | preflight 反查 active spec、六份 plan artifact、浏览器源码与 scenario claims，只允许当前七个状态的显式证据；关键 DOM、主屏起始坐标或 390px viewport overflow 断言通过，每个状态取得非空内存截图；`report` desktop 保持 TopBar 可见；i18n、负向 lint、build 与 Playwright 均通过 | `test/scenarios/e2e/p0-059-report-pixel-parity-i18n-and-out-of-scope-negative/` |

## 6 数据隔离与污染恢复

每个场景按 `test/scenarios/e2e/README.md` §5 / §3 / §6 / §8 约定：

- 数据隔离：每个 scenario 使用独立的 `user_id` / `session_id` / `report_id` / `target_job_id` 命名空间；不复用 `E2E.P0.018 ~ E2E.P0.055` 已占用的资源（含 workspace + frontend-workspace-and-practice + backend-practice + backend-review 范围）
- Vitest 套件：每个 test 文件用 fresh `setupTests.ts` mockTransport + InterviewContext 独立 hydrate；不污染其他文件状态
- Playwright：每个 spec 使用独立 fixture variant；截图只保留在内存中并断言字节非空，不写入仓库或场景输出目录
- 污染恢复：场景失败时按 README §8 顺序：① 清理场景自身资源；② 定位并恢复 shared 组件（mockTransport spy buffer、Playwright browser context）；③ 仅在 ① ② 失败时 `test/scenarios/env-cleanup.sh && env-setup.sh` 全量重建
- 不预设 Helm chart / 外部 Git 平台名称；所有命令以本仓库脚本为真理源

## 7 与单元测试边界

本 BDD plan 验证用户可见行为切片（HTTP API 调用 + UI 渲染 + nav payload + 隐私红线 + pixel parity + scoped out-of-scope grep）；不重复 useReportGenerationPoll / useFeedbackReport 内部 7 态 4 态、ReadinessTier 算法、retry_focus 选择、next_action 决策等单元测试覆盖（详见 [test-plan](./test-plan.md)）。001 阶段不存在 runtime AI 调用（前端不直连 LLM），"AI 失败测试" 通过 fixture variant 模拟，不在 BDD plan 直接测试 backend AI 失败。

## 8 与 spec AC 映射

| spec AC | 覆盖场景 |
|---------|----------|
| C-1（两条 owner route 接管） | `E2E.P0.056`（generating + report 都渲染正式 Screen） |
| C-2（GeneratingScreen 轮询 happy path） | `E2E.P0.056` poll hook ready callback + GeneratingScreen route handoff focused assertions（不锁固定轮询次数） |
| C-3（GeneratingScreen 失败处理） | `E2E.P0.056` GeneratingScreen failed route handoff + `E2E.P0.058` poll-hook onFailed/errorCode |
| C-4（GeneratingScreen 超时） | `E2E.P0.056` GeneratingScreen timeout UI + `E2E.P0.058` poll-hook max-attempts state |
| C-5（ReportDashboard ready 渲染） | `E2E.P0.056` ReportScreen + DetailSurface focused assertions |
| C-6（ReportFailureState） | `E2E.P0.058` failure component + ReportScreen route-state focused assertions |
| C-7（ReportMissingSessionState） | `E2E.P0.058` missing component + ReportScreen no-fetch focused assertions |
| C-8（5 detail tab 切换） | `E2E.P0.056` 子断言（5 tab + ARIA tablist） |
| C-9（复练 CTA 路径 A） | `E2E.P0.057` （A） + （C） |
| C-10（复练 CTA 路径 B） | `E2E.P0.057` （B） |
| C-11（UI source structure parity） | `E2E.P0.056` 子断言（testid 命中）+ `E2E.P0.059` |
| C-12（UI visual geometry parity） | `E2E.P0.059` |
| C-13（UI stale-contract negative） | `E2E.P0.059` 子断言（scoped out-of-scope grep） |
| C-14（BDD 主流程 + 关键分支） | `E2E.P0.056` + `E2E.P0.057` + `E2E.P0.058` + `E2E.P0.059` |
| C-15（Privacy 红线） | `E2E.P0.056` raw resume/JD body avoidance + `E2E.P0.057` payload/URL/storage；P0.058 只验证 typed error mapping，不扩大为全链路隐私证据 |
| D-3（GeneratingScreen 轮询节奏） | `E2E.P0.056` happy/route + `E2E.P0.058` failed/404/timeout hook states |
| D-4（状态分支三态） | `E2E.P0.056` screen states + `E2E.P0.058` report/poll failure states |
| D-5（复练 CTA payload） | `E2E.P0.057` |
| D-6（报告失败状态语义） | `E2E.P0.058` ① |
| D-7 / D-13（i18n + 隐私红线） | `E2E.P0.059` + `E2E.P0.056-058` 子断言 |
| D-10 / D-11（4 档 readiness + 维度三态映射） | `E2E.P0.056` 子断言 + `E2E.P0.059` |
| D-12（out-of-scope 术语） | `E2E.P0.059` 子断言（scoped out-of-scope grep） |
