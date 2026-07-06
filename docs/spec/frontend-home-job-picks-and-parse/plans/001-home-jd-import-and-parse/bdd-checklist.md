# 001 BDD Checklist

> **版本**: 2.0
> **状态**: completed
> **更新日期**: 2026-07-06

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.014 Home 默认渲染（empty + non-empty + 12+）

- [x] 创建场景目录 `test/scenarios/e2e/p0-014-home-default-render/`，含 `README.md`（§6 baseline + §7 离线限制）
- [x] 准备 fixture variant：`listTargetJobs.json` 至少 3 个 variant（empty / one-job / 12+jobs），按 `mock-contract-suite` 规则配置；通过 `make validate-fixtures`
- [x] 实现 `scripts/setup.sh`（预检 frontend dist + chromium + fixture variant 切换入口）/ `scripts/trigger.sh`（运行 home spec + 三 variant 切换）/ `scripts/verify.sh`（断言 hero/textarea/aux cards/empty state/3-card cap + 更多跳转/`updatedAt desc` 排序/topbar 高亮/zh-en 切换/warm-dark-customAccent 切换）/ `scripts/cleanup.sh`
- [x] 2026-07-06 dropdown + recent cap gate：P0.014 trigger / verify / README / expected outcome 必须验证 `home-resume-select` 是下拉框而不是平铺 button 列表，`twelve-plus` 只渲染 3 张最近卡片，`home-recent-more` 点击跳转 `workspace`。
  - Evidence 2026-07-06: `test/scenarios/e2e/p0-014-home-default-render/scripts/setup.sh`, `trigger.sh`, `verify.sh`, and `cleanup.sh` exited 0 after README / seed / expected outcome updates; trigger includes focused Home tests and verify rejects missing dropdown / 3-card cap / `更多` workspace marker.
- [x] 2026-07-06 source layout gate：P0.014 trigger / verify / README / expected outcome 必须验证 `home-source-layout` 分离粘贴 JD 与上传文件，`home-upload-trigger` 不在 `home-jd-input-card` 内，`home-resume-select` 与 `home-resume-create` 同一行，`home-jd-submit` 位于 `home-resume-row` 下方且不在 textarea card 内。
  - Evidence 2026-07-06: `test/scenarios/e2e/p0-014-home-default-render/scripts/setup.sh`, `trigger.sh`, `verify.sh`, and `cleanup.sh` exited 0; trigger includes `HomeLayout.test.tsx`, and Home pixel parity passed 10 desktop/mobile layout tests.
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：spec 调用栈 + variant 切换日志 + 截图（baseline + 当前）+ retired-testid grep 0 命中
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.014 行（关联需求 `frontend-home-job-picks-and-parse C-1, C-4`，状态 Ready，automated）

## E2E.P0.015 Paste/Upload/URL → Import → Parse 主路径

- [x] 创建场景目录 `test/scenarios/e2e/p0-015-jd-import-and-parse/`，含 `README.md`
- [x] 准备 fixture variant：`Uploads/createUploadPresign.json` 至少 2 个 variant（target_job_attachment 成功 / 4xx），`importTargetJob.json` 至少 4 个 variant（manual_text 成功 / file 成功 / url 成功 / 422 invalid source），`getTargetJob.json` 至少 3 个 variant（queued → processing → ready 序列模拟 / failed / requirements & hidden signals 富数据）
- [x] 实现 `scripts/setup.sh`（含三种 source variant 切换入口）/ `scripts/trigger.sh`（按 A/B/C 三条路径运行）/ `scripts/verify.sh`（断言 upload 路径先调 `createUploadPresign` 且 `purpose=target_job_attachment`、ImportTargetJobRequest discriminator 三种 type、side-effect 请求均带 `Idempotency-Key`、route 跳 parse、loading 4 步节奏 ≥600ms、preview 字段映射、summary/fitSummary provenance 可追溯、failed UI、JD raw text 0 命中、4xx inline 错误、前端 network/client spy 无 LLM/provider/prompt-registry 调用）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS（三条 source 路径 + failed variant 共 4 子用例）
- [x] 记录验证证据：mockTransport 调用日志 spy + 隐私反查日志 + 4xx 路径截图
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.015 行（关联需求 C-2, C-3, C-6，状态 Ready，automated）
- [x] 2026-05-24 regression gate：ready fixture 不得直接进入 preview；P0.015 trigger 通过 `ParseFlow.test.tsx` 断言 loading step DOM 先出现，完成 `ui-design` tick 后才出现 `parse-basics-title`。 <!-- evidence: `.test-output/e2e/p0-015-jd-import-and-parse/trigger.log` includes `ParseFlow.test.tsx` PASS and 54 total home/parse tests PASS; verify.sh PASS -->
- [x] 2026-05-24 browser gate：P0.015 trigger 通过 Playwright `tests/pixel-parity/parse.spec.ts` 打开 `/parse?targetJobId=...`，fixture-backed ready response 下捕获 loading DOM screenshot，断言 preview 在 loading window 内缺席且 tick 完成后出现；verify.sh 要求 `E2E.P0.015 ready-response loading browser gate screenshotBytes=` marker。 <!-- evidence: `.test-output/e2e/p0-015-jd-import-and-parse/trigger.log` includes Playwright parse.spec ready-response browser gate + screenshotBytes marker; verify.sh PASS -->
- [x] 2026-05-24 same-route target switch regression：P0.015 trigger 的 `ParseFlow.test.tsx` 覆盖同一 mounted `ParseScreen` 在 preview 状态切换 `targetJobId` 的 rerender 路径，断言旧 preview 被清空、loading DOM 重新出现、tick 完成后 hydrate 新 TargetJob。 <!-- evidence: `.test-output/e2e/p0-015-jd-import-and-parse/trigger.log` includes ParseFlow.test.tsx PASS; focused local Red-Green reproduced stale preview before fix -->
- [x] 2026-07-06 home resume pre-bind gate：P0.015 trigger / verify / README / expected outcome 必须改为验证 Home 主路径先选择已有 ready 简历，再点击「立即面试」提交 import；verify.sh 必须拒绝旧 hero sub、旧「解析并确认面试」按钮文案、未选择简历仍可 import，以及 parse route 缺少真实 `resumeId` 的成功 marker。
  - Evidence 2026-07-06: `test/scenarios/e2e/p0-015-jd-import-and-parse/scripts/trigger.sh` and `verify.sh` exited 0; trigger includes `HomeResumeSelection.test.tsx`, `HomeImport.test.tsx`, `HomeAuthGate.test.tsx`, frontend build, and parse loading Playwright gate.
- [x] 2026-07-06 home resume dropdown gate：P0.015 trigger / verify / README / expected outcome 必须证明用户通过 `home-resume-select` 下拉框选择 ready 简历后，paste / upload / URL import 继续携带真实 `resumeId` 到 parse；不得要求点击平铺简历按钮。
  - Evidence 2026-07-06: `test/scenarios/e2e/p0-015-jd-import-and-parse/scripts/setup.sh`, `trigger.sh`, `verify.sh`, and `cleanup.sh` exited 0; trigger includes `HomeResumeSelection.test.tsx`, `HomeImport.test.tsx`, `HomeAuthGate.test.tsx`, frontend build, and parse loading Playwright gate.
- [x] 2026-07-06 home source separation gate：P0.015 trigger / verify / README / expected outcome 必须证明新布局下 paste、upload、URL 三种 source 仍使用同一个 ready resume gate；上传文件从独立 source panel 打开 modal，paste 的「立即面试」在简历行下方提交并携带真实 `resumeId`。
  - Evidence 2026-07-06: `test/scenarios/e2e/p0-015-jd-import-and-parse/scripts/setup.sh`, `trigger.sh`, `verify.sh`, and `cleanup.sh` exited 0; trigger includes `HomeLayout.test.tsx`, `HomeResumeSelection.test.tsx`, `HomeImport.test.tsx`, and `HomeAuthGate.test.tsx`, plus parse loading Playwright gate.

## E2E.P0.016 Parse 编辑 + 绑定简历 + Save/Start handoff

- [x] 创建场景目录 `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/`，含 `README.md`
- [x] 准备 fixture variant：`updateTargetJob.json` 至少 2 个 variant（成功 / 4xx）；signed-in / signed-out 两种状态切换入口
- [x] 实现 `scripts/setup.sh`（含 signed-in/out 切换）/ `scripts/trigger.sh`（按 A 已登录 / B 未登录 / C 通用 三子场景运行，并追加 Playwright browser route/context gate）/ `scripts/verify.sh`（断言 updateTargetJob body schema 仅含 supplied fields 且不含 hit toggle / summary / fitSummary / hidden signals、`Idempotency-Key` header 存在、route 跳 workspace 携带 7 字段 interviewContext、auth pending action 触发与登录恢复、Re-parse 只重新轮询 `getTargetJob` 不直连 LLM、Cancel 行为、隐私反查、browser gate contextKeys + screenshotBytes marker）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS（A/B/C 共 ≥3 子用例）
- [x] 记录验证证据：updateTargetJob request body 截取 + auth pending action 路径流 + interviewContext 字段集合断言
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.016 行（关联需求 C-5, C-7，状态 Ready，automated）
- [x] 2026-05-24 browser route/context gate：P0.016 trigger 通过 Playwright `tests/pixel-parity/parse.spec.ts` 打开 `/parse?targetJobId=...`，mock generated API 返回 ready TargetJob，点击 Confirm 后验证 `/workspace` query 携带 `targetJobId` / `jobId` / `jdId` / `planId` / `resumeVersionId` / `roundId` / `roundName`，并捕获 `workspace-missing-resume` screenshot；verify.sh 要求 `E2E.P0.016 parse confirm workspace browser gate contextKeys=targetJobId,jobId,jdId,planId,resumeVersionId,roundId,roundName screenshotBytes=` marker。 <!-- evidence: `.test-output/e2e/p0-016-parse-confirm-to-workspace/trigger.log` includes focused Playwright parse.spec confirm gate desktop/mobile PASS and screenshotBytes markers; verify.sh PASS -->
- [x] 2026-06-30 resume binding gate：P0.016 trigger / verify / README / expected outcome 必须改为验证 Parse 成功出口携带真实 ready `resumeId`，`仅保存规划` 不再渲染 `workspace-missing-resume`，`立即面试` 通过 `workspace autoStartPractice=1` handoff 创建 session 后到 `practice`；verify.sh 必须拒绝 `resume-unbound` 成功 marker。
  - Evidence 2026-06-30: `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/setup.sh`, `trigger.sh`, `verify.sh`, and `cleanup.sh` all exited 0. Trigger ran `targetJob.realApiMode.test.ts`, focused Parse Vitest tests, frontend build, and Playwright markers `parse save-plan workspace browser gate ... resumeId=01918fa0-0000-7000-8000-000000001000` plus `parse start-interview autoStart browser gate ... route=practice`.
- [x] 2026-07-06 parse inherit home resume gate：P0.016 trigger / verify / README / expected outcome 必须新增 route `resumeId` 继承子用例，证明 Home 传入的 ready 简历在 Parse preview 已绑定；同时保留 route `resumeId` 无效时不默认选中最近简历、Save/Start disabled 的负向覆盖。
  - Evidence 2026-07-06: `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/trigger.sh` and `verify.sh` exited 0; verbose Vitest output contains `inherits a valid route resumeId from the Home immediate interview handoff`, followed by desktop/mobile Save/Start Playwright gates rejecting `resume-unbound`.

## E2E.P0.017 jd_match P1 Placeholder Smoke

- [x] 创建场景目录 `test/scenarios/e2e/p0-017-jd-match-placeholder/`，含 `README.md`
- [x] 准备：无需新 fixture（placeholder 不消费数据）；脚本入口校验 D1 generated client 不会被 jd_match 屏触发额外 API 调用
- [x] 实现 `scripts/setup.sh` / `scripts/trigger.sh`（通过 TopBar 与 home aux card 双入口进入 jd_match）/ `scripts/verify.sh`（断言 hero / profile chip / 三 tab / placeholder testid 全命中、TopBar `topbar-nav-jd_match` 高亮、旧业务 testid grep 0 命中、i18n zh/en 切换、warm-dark-customAccent 切换、mobile 不溢出、generated client 调用次数为 0 或仅限于 `getMe` 等已存在 D1 调用）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：retired-testid grep 0 命中日志 + generated client spy + mobile 截图
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.017 行（关联需求 C-8，状态 Ready，automated）

## Real Backend Overlay（2026-05-22）

- [x] P0.014-P0.016 trigger scripts export/default `VITE_EI_API_MODE=real` and `VITE_EI_API_BASE_URL=http://localhost:8080/api/v1`, then run `src/api/targetJob.realApiMode.test.ts` before fixture-backed UI sub-cases. <!-- evidence: 2026-05-22 P0.014/P0.015/P0.016 trigger logs each include VITE_EI_API_MODE=real, VITE_EI_API_BASE_URL=http://localhost:8080/api/v1, and targetJob.realApiMode.test.ts PASS -->
- [x] P0.014-P0.016 verify scripts reject missing real-mode markers, so fixture-backed UI PASS alone cannot satisfy TargetJobs/import/parse completion. <!-- evidence: 2026-05-22 verify scripts require VITE_EI_API_MODE=real, VITE_EI_API_BASE_URL=http://localhost:8080/api/v1, and targetJob.realApiMode.test.ts markers -->
- [x] Backend TargetJob owner evidence paired: E2E.P0.010 / P0.011 / P0.012 / P0.013 setup→trigger→verify→cleanup PASS through `backend/cmd/api` live HTTP harness. <!-- evidence: 2026-05-22 backend scenarios P0.010-P0.013 all PASS; verify scripts accepted generated result artifacts -->
- [x] Backend upload support evidence paired: `POST /api/v1/uploads/presign` route and handler focused tests PASS, covering the plan001 `createUploadPresign` supporting operation. <!-- evidence: 2026-05-22 `go test ./cmd/api -run TestBuildUploadRoutesAlignsIdempotencyTTLWithPresignTTL -count=1` PASS; `go test ./internal/upload/handler -run 'TestCreateUploadPresignReturnsCreatedResponse|TestCreateUploadPresignIdempotencyReplayAndTTL' -count=1` PASS -->
- [x] E2E.P0.017 remains jd_match UI-only smoke and intentionally has no TargetJobs real-mode overlay. <!-- evidence: plan001 P0.017 does not consume TargetJobs/import/parse operations -->

## 整体 Regression（Phase 6 收口）

- [x] D1+D2+D3 Regression 重跑：`E2E.P0.001 / 002 / 004 / 005 / 006` setup→trigger→verify→cleanup 全部 PASS（D2 视觉系统不被 home/parse/jd_match 改动破坏）
- [x] `pnpm --filter @easyinterview/frontend test` 全量 Vitest PASS（含本 plan 新增测试文件）
- [x] `pnpm --filter @easyinterview/frontend test:pixel-parity` 在 D2/D3 现有 21 spec × 2 viewport = 42 项基础上累加 home/parse/jd_match 新增 spec，总数全 PASS，并确认 parse loading footer 与 `ui-design` 源级结构一致但无前端 LLM/provider 请求
- [x] `pnpm --filter @easyinterview/frontend typecheck` + `pnpm --filter @easyinterview/frontend build` + `make build` 全 PASS
- [x] `make docs-check` zero drift；`/sync-doc-index --fix-index` post-fix zero drift；`check_md_links` 双 OK
