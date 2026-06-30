# 001 BDD Plan

> **版本**: 1.5
> **状态**: active
> **更新日期**: 2026-06-30

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 关联 Phase | 关联 Spec C-* | 关联 BDD-Gate（主 checklist） |
|---------|------|-----------|--------------|----------------------------|
| E2E.P0.014 | primary path · home 默认渲染 | Phase 1 + 2 | C-1, C-4 | Phase 1.5、Phase 2.6 |
| E2E.P0.015 | primary path · paste→import→parse 主路径 + alternate path（upload / URL variants） + failure path（4xx / failed） | Phase 3 + 4 | C-2, C-3, C-6 | Phase 3.6、Phase 4.10 |
| E2E.P0.016 | primary path · parse 编辑 + 绑定简历 + Save/Start handoff + failure（updateTargetJob 4xx）+ empty resume gate | Phase 4 + Phase 7 | C-5, C-7, C-17 | Phase 4.10 + Phase 7.4 |
| E2E.P0.017 | regression / legacy-negative · jd_match P1 placeholder smoke + 旧 prototype 业务 testid 反向 grep | Phase 5 | C-8 | Phase 5.5 |

---

## 1.1 Real Backend Overlay

E2E.P0.014-P0.016 的 UI 子用例继续使用 fixture-backed component transports，原因是这些场景要稳定覆盖 DOM、source variants、auth pending action、4xx/failed variants、privacy negative grep 与 responsive parity。2026-05-22 L2 remediation 在每个 trigger 前置运行 `src/api/targetJob.realApiMode.test.ts`，并显式设置 `VITE_EI_API_MODE=real` / `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`：该 gate 证明 production bootstrap/generated client 对 `listTargetJobs`、`createUploadPresign`、`importTargetJob`、`getTargetJob`、`updateTargetJob` 使用真实 backend base URL、cookie credentials、side-effect `Idempotency-Key` 与 TargetJob provenance roundtrip。真实 backend route/persistence/auth/IK/parse semantics 由 `backend-targetjob/001-targetjob-import-and-parse-bootstrap` 的 E2E.P0.010-P0.013 live scenarios 配对证明；upload presign route/handler 由 `backend-upload/001-file-objects-and-presign-baseline` focused tests 配对证明。E2E.P0.017 是 jd_match UI-only smoke，不属于 TargetJobs/import/parse real backend overlay。

## Phase 1 + 2: Home 默认渲染（含 Recent mocks 三态）

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.014 | Home 默认渲染（empty + non-empty + 12+） | 用户打开 frontend dev server，未登录，listTargetJobs fixture 配置三种 variant：empty / one-job / 12+jobs | 切换三种 fixture variant 分别加载 `/#route=home` | （1）Hero `home-hero-{label,title,sub}`、JD textarea + Submit、aux cards `home-aux-{jobpicks,debrief}`、Resume create CTA 全部渲染并 testid 命中；（2）empty variant 显示 `HomeEmptyState`，点击「回到 JD 输入」按钮 focus textarea；（3）one-job variant 渲染 1 张 `home-recent-mock-card-${id}`，status pill computed background 对应 D2 token，MiniRoundRail 当前轮次圆点位置正确；（4）12+jobs variant 仅渲染 12 张卡片，按 `updatedAt desc` 排序；（5）TopBar `topbar-nav-home` 高亮；（6）切换 zh/en、warm→dark→customAccent，关键文本 / computed background 出现可见变化 | `test/scenarios/e2e/p0-014-home-default-render/` |

## Phase 3 + 4: Paste/Upload/URL → Import → Parse 主路径

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.015 | Paste/Upload/URL → import → parse loading → preview | 用户已登录，importTargetJob fixture 返回 `{ targetJobId: "uuid", job }`，createUploadPresign fixture 返回 `fileObjectId`，getTargetJob fixture 配置 `analysisStatus` 序列 `queued → processing → ready` 与 `failed` 两种 variant | 用户在 home 分别走三条路径：（A）粘贴 JD 文本点 Submit；（B）打开 upload modal 拖入 placeholder 文件后 Continue；（C）打开 URL modal 输入 URL 后 Continue；并在解析过程中切换 `failed` variant 触发失败态 | （1）A/C 路径分别提交 `source.type=manual_text|url` 的 `ImportTargetJobRequest`；B 路径先调用 `createUploadPresign` `purpose=target_job_attachment` 并把返回 `fileObjectId` 写入 `source.type=file`；side-effect 调用均带 `Idempotency-Key`；（2）成功响应后 route 跳 `parse?targetJobId=…&source=…`；（3）Parse 屏先渲染 `parse-loading-step-${0..3}` + `parse-loading-footer`，按 ≥600ms 节奏推进，footer 只展示 backend parse metadata / fixture metadata，不触发任何 LLM/provider 请求；（4）`analysisStatus=ready` 后切到 preview，渲染 fixture/backend response 中 title/companyName/locationText/requirements/summary.interviewHypotheses/summary.coreThemes/fitSummary.riskSignals，且 summary/fitSummary `GenerationProvenance` 可追溯；（5）`analysisStatus=failed` variant 下显示 failed UI（重新解析 / 返回首页 2 button），不展示伪造 preview；（6）JD raw text 在 console / URL / localStorage / telemetry 全部 0 命中；（7）4xx import / presign 响应触发 inline 错误并保留 textarea/modal 输入；（8）network/client spy 只允许 generated `createUploadPresign`、TargetJobs client + existing shell runtime/auth 调用，不允许前端直连 AI provider、prompt registry 或 provider-specific endpoint | `test/scenarios/e2e/p0-015-jd-import-and-parse/` |

## Phase 4 / 7: Parse 编辑 + 绑定简历 + Save/Start handoff

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.016 | Parse 编辑 + 显式绑定简历 + Start/Save handoff | 用户在 parse 屏 preview 阶段，`listResumes` 返回 ready 简历；另有 empty/failed 变体 | 用户编辑 title 字段；选择前 Save/Start disabled；用户显式点击一份简历后分别点击 `仅保存规划` 与 `立即面试` | （A 选择前）（1）页面不展示默认绑定简历名称；（2）`仅保存规划` / `立即面试` disabled；（B 保存规划）（1）调 `updateTargetJob(targetJobId, body, { idempotencyKey })` body 仅含 supplied fields，不含 hit toggle 状态、summary、fitSummary 或 hidden signals；（2）成功后 route 跳 `workspace?targetJobId=&jobId=&jdId=&planId=&resumeId=&roundId=&roundName=`，`resumeId` 为用户点击的真实 ready 简历 id，禁止 `resume-unbound`；（3）不渲染 `workspace-missing-resume` 成功态；（C 立即面试）（1）调同一保存路径后进入 `workspace` 并携带 `autoStartPractice=1`，由 workspace `useStartPractice` 创建 session 后进入 `practice`；（2）handoff / pendingAction params 携带真实 `resumeId`；（D 无 ready 简历）（1）`立即面试` 与 `仅保存规划` disabled；（2）`parse-resume-create` 导航 `resume_versions?flow=create`；（E 通用）Re-parse / Cancel / 隐私负向保持原有要求；browser gate 输出真实 resumeId context marker，并拒绝 `workspace-missing-resume` / `resume-unbound` 成功 marker | `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/` |

## Phase 5: jd_match P1 Placeholder Shell

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.017 | jd_match P1 placeholder smoke | 用户已登录或未登录均可 | 通过 TopBar Job Picks 入口与 home 的 `home-aux-jobpicks` aux card 分别进入 `jd_match` 路由 | （1）route 渲染 `jd_match` 命中 `JDMatchScreen` 而非 PlaceholderScreen；（2）testid `jdmatch-hero-{label,title,sub}`、`jdmatch-profile-chip-{title,years,location,skills,sources}`、`jdmatch-tab-{recommended,search,watchlist}`、`jdmatch-placeholder`、`jdmatch-placeholder-cta` 全部命中；（3）TopBar `topbar-nav-jd_match` 高亮；（4）i18n zh/en placeholder 文案切换；（5）旧 prototype 业务 testid（`jdmatch-card-*` / `jdmatch-saved-search-*` / `jdmatch-watchlist-*` / `jdmatch-market-signal-*` / `jdmatch-search-bar` / `jdmatch-search-results` / `jdmatch-jd-detail-*` / `jdmatch-agent-status`）grep 0 命中；（6）warm/light → dark → customAccent 三态切换 hero 与 profile chip computed background 出现可见变化；（7）mobile (390×844) viewport 下 hero / profile chip / 三 tab 不溢出 | `test/scenarios/e2e/p0-017-jd-match-placeholder/` |
