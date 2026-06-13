# Frontend Resume Workshop Create Flow and Onboarding Checklist

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-06-13

**关联计划**: [plan](./plan.md)

## Phase 0: 上游依赖 gate

- [x] 0.1 确认 [backend-resume/002 Phase 1](../../../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md#phase-1-b2-d-18-additive-confirmresumestructuredmaster--b1-错误码增补) 已落地：`openapi/openapi.yaml` 含 `POST /api/v1/resumes/{resumeAssetId}/structured-master` operation 与 `ConfirmResumeStructuredMasterRequest` schema；`openapi/fixtures/Resumes/confirmResumeStructuredMaster.json` 含 `default / idempotency-replay / already-exists-409 / validation-422` 四 scenario；frontend generated client 暴露 `confirmResumeStructuredMaster` method 与对应 request/response 类型（验证：`ls openapi/fixtures/Resumes/confirmResumeStructuredMaster.json` + `git grep confirmResumeStructuredMaster -- frontend/src/api/generated/`）
- [x] 0.2 确认 [backend-upload/001](../../../backend-upload/plans/001-file-objects-and-presign-baseline/plan.md) `createUploadPresign` handler + fixture 已就位；mock-first dev path 可直接消费（验证：`ls openapi/fixtures/Uploads/createUploadPresign.json` + spy generated client `createUploadPresign` 存在）
- [x] 0.3 确认 [backend-resume/001](../../../backend-resume/plans/001-asset-register-parse-and-listing/plan.md) `registerResume` + `getResume` handler 已就位；mock-first 路径与切真路径均可用（验证：`ls openapi/fixtures/Resumes/registerResume.json openapi/fixtures/Resumes/getResume.json`）
- [x] 0.4 确认 [frontend-resume-workshop/001](../001-listing-routing-and-detail-readonly/plan.md) completed 且 `flow=create` 在当前实现下渲染 `<NotImplementedPlaceholder>`（验证：`git grep -nE "NotImplementedPlaceholder|flow.*create" frontend/src/app/screens/resume-workshop/`）
- [x] 0.5 旧入口 grep baseline：`git grep -nE "OnboardingScreen|onboarding=true|welcome|mistake|growth|drill|followup|STAR|experiences|voice" frontend/src/app/screens/resume-workshop/`（baseline 0 命中前置，Phase 6.8 收口时再次验证）

## Phase 1: ResumeCreateFlow 容器 + stage / mode 状态机 + auth gate

- [x] 1.1 修订 `frontend/src/app/screens/resume-workshop/ResumeWorkshopScreen.tsx`：`flow === 'create'` 时渲染 `ResumeCreateFlow`，placeholder 仅保留给 `flow === 'branch'`（验证：Vitest 单测 + grep `flow.*create.*ResumeCreateFlow`）
- [x] 1.2 实现 `frontend/src/app/screens/resume-workshop/create/ResumeCreateFlow.tsx`：源级复刻 UI 真理源 Header + 主区两栏 + Card / Btn / Icon primitive 调用；stage / mode 状态机 8 + case Vitest PASS
- [x] 1.3 auth gate：未登录访问 `resume_versions?flow=create` 渲染 auth gate / 登录引导；pendingAction 仅保留 `{ route, params: { flow, createMode? } }`，不含 rawText / guidedAnswers / file binary（验证：Vitest mock client spy 0 个 protected API 请求 + pendingAction params 字段集合断言）
- [x] 1.4 Vitest 组件单测：stage / mode 切换 + back / cancel + auth gate 至少 ≥ 8 case PASS
- [x] 1.5 i18n key 空间脚手架：`resumeWorkshop.create.*` 命名空间初始化在 `frontend/src/app/i18n/en.ts` + `zh.ts`，未实际渲染的 key 可暂缺，但 namespace 必须存在（验证：Vitest i18n 模拟切换不报错）

## Phase 2: Upload tab 双步上传 + IK 契约

- [x] 2.1 实现 `frontend/src/app/screens/resume-workshop/create/UploadTab.tsx`：源级复刻 UI 真理源 dropzone / icon / file picker 行为；扩展名 / size 客户端 pre-check（验证：Vitest 渲染 + 客户端 pre-check 边界 case PASS）
- [x] 2.2 实现 `hooks/useResumePresignUpload.ts`：Step 1 `createUploadPresign` + IK；Step 2 `fetch(uploadUrl, ...)` 上传 binary；TTL 过期重 presign；retry 复用 fileObjectId（验证：Vitest happy / failure / TTL / retry 至少 ≥ 6 case PASS + request spy 断言 `Idempotency-Key` header）
- [x] 2.3 衔接到 register：UploadTab `handleSubmit` 不直接调 registerResume；由父 `ResumeCreateFlow` 拿到 `fileObjectId` 后触发 Phase 3 `useResumeRegistration`（验证：Vitest spy 调用顺序）
- [x] 2.4 错误映射：presign `VALIDATION_FAILED` inline / PUT 失败 toast + 重试 / TTL expired 自动重 presign（验证：Vitest）
- [x] 2.5 隐私：file binary 不出现在 console / URL / pendingAction / localStorage / mock transport log（验证：Vitest spy grep）

## Phase 3: Paste + Guided tab + `useResumeRegistration` hook

- [x] 3.1 实现 `create/PasteTab.tsx`：textarea + submit disabled when empty + i18n 文案与 UI 真理源对齐（验证：Vitest disabled / enabled 切换）
- [x] 3.2 实现 `create/GuidedTab.tsx`：左栏 5 step nav + 右栏 question + textarea + 上一步 / 下一步 / Generate v1 按钮；提交 mapping 到 `{ recentRole, direction, proofProject, metrics, target }` jsonb（验证：Vitest 5 step nav + payload shape）
- [x] 3.3 实现 `hooks/useResumeRegistration.ts`：sourceType × payload 三态 + `generateIdempotencyKey()` + generated client `registerResume`；成功后触发 parent `stage='parsing'` + 携带 `resumeAssetId`（验证：Vitest 三 sourceType payload mapper + IK header + happy / 422 / 5xx 至少 ≥ 9 case PASS）
- [x] 3.4 客户端 title 默认值派生：`create/util/title.ts` 提供 mode-specific 默认 title（如 "Pasted resume" / "粘贴的简历"）；不预填用户提交内容（验证：Vitest）
- [x] 3.5 fixture parity test：`registerResume.json` `default` (upload) / `paste-text` / `guided-answers` scenarios 与 hook payload 字节兼容（验证：fixture parity test PASS）

## Phase 4: Agent Parsing stage（`ResumeParseFlow` + `useResumeParsingPolling`）

- [x] 4.1 实现 `create/ResumeParseFlow.tsx`：源级复刻 UI 真理源 7 step ticker + "取消并返回修改" button（验证：Vitest 渲染 + step transition 动画基线）
- [x] 4.2 实现 `hooks/useResumeParsingPolling.ts`：指数退避轮询 `getResume(resumeAssetId)` (初始 1500ms / backoff 1.4x / max 8 attempt / ~30s 上限)；终态 ready / failed / timeout / cancel（验证：Vitest happy / failed / timeout / cancel 至少 ≥ 6 case PASS）
- [x] 4.3 mock harness：在 fixture 未覆盖 `queued / generating / failed` parseStatus scenario 时使用 mock client deterministic stepping；测试断言显式标注"mock attempt-aware"，retrospective 记录补 fixture 提议（验证：Vitest mock harness 切换断言 + retrospective 文档草稿）
- [x] 4.4 IK 与 header：`getResume` 不属 side-effect op，不要求 IK header；request spy 断言无 `Idempotency-Key` header（验证：Vitest spy）
- [x] 4.5 隐私：ParseFlow DOM 不渲染 parsedTextSnapshot / originalText / guidedAnswers / parsedSummary 字段；URL / pendingAction 不堆叠 polling 进度（验证：Vitest DOM grep + Playwright DOM/URL sniff）
- [x] 4.6 cancel 路径：返回 input stage 时保留 createMode + rawText / guidedAnswers / pickedFile，不触发新的 registerResume（验证：Vitest state preservation 断言）

## Phase 5: ResumePreviewConfirm + `confirmResumeStructuredMaster` 保存 v1

- [x] 5.1 Gate：确认 confirmResumeStructuredMaster fixture + generated client 已就位（Phase 0.1 PASS）；不满足则 PreviewConfirm 渲染 `<ComingSoonPreviewConfirm>` 并 fail Phase 5
- [x] 5.2 实现 `create/ResumePreviewConfirm.tsx`：源级复刻 UI 真理源 header + 两栏（左草稿主体 / 右 "会保存什么" + "解析备注"）+ 返回 / 确认按钮（验证：Vitest 渲染 + DOM testid 覆盖）
- [x] 5.3 adapter `mapParsedSummaryToStructuredProfileDraft`：把 `resume_assets.parsedSummary` 投影为 `ConfirmResumeStructuredMasterRequest.structuredProfile` 字段集（identity / summary / experience / projects / skills / education）（验证：Vitest 单测 ≥ 6 case）
- [x] 5.4 实现 `hooks/useResumeStructuredMasterConfirm.ts`：`generateIdempotencyKey()` + generated client `confirmResumeStructuredMaster` + 三态错误映射（验证：Vitest happy / replay / 409 / 422 / IK header 至少 ≥ 8 case PASS）
- [x] 5.5 409 fallback：触发 `listResumeVersions(resumeAssetId)` 找到已存在 master version → nav `resume_versions?versionId={masterId}&tab=preview`；如未找到则降级到 list（验证：Vitest fallback path）
- [x] 5.6 toast 文案：保存成功 toast + 409 提示 + 422 inline，仅使用 enum / generic 文案，不回显 raw error envelope（验证：Vitest）
- [x] 5.7 fixture parity test：`confirmResumeStructuredMaster.json` `default` / `idempotency-replay` / `already-exists-409` / `validation-422` 与 hook 行为字节匹配（验证：fixture parity test）
- [x] 5.8 隐私：DOM 渲染 structuredProfile 字段但 URL / pendingAction / localStorage / mock transport log 不含 structuredProfile 字符串内容（验证：Vitest + Playwright）

## Phase 6: Home / Workspace CTA 串通 + i18n + a11y + 隐私 + UI parity + BDD + 旧入口负向

- [x] 6.1 Home CTA 集成测试：从 `home` route 渲染 `还没有简历？1 分钟创建` button → click → `resume_versions?flow=create` → 渲染 `ResumeCreateFlow`（DOM testid `resume-create-flow` 命中）（验证：Vitest integration test 或场景内集成断言；BDD P0.083 中验证）
- [x] 6.2 Workspace CTA 集成测试：`WorkspaceMissingResumeState` → 点击 "创建简历" → `resume_versions?flow=create` → 渲染 `ResumeCreateFlow`；pendingAction 未登录态只携带 `{ flow: 'create' }`（验证：Vitest + 更新 P0.018 / P0.019 既有断言不退化）
- [x] 6.3 旧 `onboarding` route alias 仍折回 `resume_versions`：normalizeRoute 不复活旧 OnboardingScreen；测试断言访问 `/onboarding` 等价 `/resume_versions`（验证：Vitest normalizeRoute test）
- [x] 6.4 i18n：新增 `resumeWorkshop.create.*` / `.parsing.*` / `.preview.*` / `.errors.*` key 空间；EN / ZH 切换关键文案 + Accept-Language header 携带 createUploadPresign / registerResume / getResume / confirmResumeStructuredMaster 四个 op（验证：Vitest + integration test 验证 header）
- [x] 6.5 a11y：`role="tablist"` / `role="tab"` / `aria-selected` / focus 管理 / ESC 行为 / 键盘 ←/→ / aria-live toast；Playwright axe-core check PASS（验证：Playwright a11y spec）
- [x] 6.6 隐私红线 grep：raw text / guidedAnswers / pickedFile / parsedTextSnapshot / parsedSummary / structuredProfile 不出现在 console / URL / pendingAction / localStorage / mock transport log / error toast 内容（验证：Vitest spy grep + Playwright DOM/network sniff）
- [x] 6.7 UI parity gate：复用 frontend-shell/003 框架；新增 `frontend/tests/pixel-parity/resume-workshop-create.spec.ts` 覆盖 5 屏幕 desktop + mobile DOM anchor + computed style + bounding box + 非空截图 buffer；clean checkout PASS 不依赖未跟踪 baseline（验证：`pnpm --filter @easyinterview/frontend build && pnpm --filter @easyinterview/frontend test:pixel-parity` PASS；首次或新机器先跑 `pnpm --filter @easyinterview/frontend test:pixel-parity:install`）
- [x] 6.8 BDD-Gate: E2E.P0.081 resume-create-flow-upload-paste-guided-happy PASS（详见 [bdd-checklist.md](./bdd-checklist.md)）
- [x] 6.9 BDD-Gate: E2E.P0.082 resume-create-flow-parsing-failure-and-retry PASS
- [x] 6.10 BDD-Gate: E2E.P0.083 resume-create-flow-preview-confirm-and-cta-handoff PASS
- [x] 6.11 旧入口 grep（收口）：`git grep -nE "welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen|onboarding=true" -- frontend/src/app/screens/resume-workshop/create/` 0 命中（验证：CI lint）
- [x] 6.12 prototype import grep：`git grep -nE "ui-design/src/(data|screen-resume-workshop)" -- frontend/src/app/screens/resume-workshop/create/` 0 命中（验证：CI lint）
- [x] 6.13 在 `test/scenarios/e2e/INDEX.md` 追加 P0.081 + P0.082 + P0.083 行（关联需求 `frontend-resume-workshop C-10`，状态 Ready，automated）
- [x] 6.14 spec / history / INDEX 同步核对：frontend-resume-workshop spec.md / history.md / `docs/spec/INDEX.md` 已保持 1.1，§3.1 D-4 / §6 C-10 / §7 plan 002 行指向当前 active plan；`docs/spec/frontend-resume-workshop/plans/INDEX.md` 已包含 002 active 行；不为 checklist 收口重复 bump spec，除非发现新的设计事实（验证：`sync-doc-index --check` PASS）
- [x] 6.15 L2 real-backend generated-client gate：P0.081-P0.083 trigger 前置 `frontendOwners.realApiMode.test.ts`；verify 检查 `VITE_EI_API_MODE=real`、默认 backend base URL 与测试文件 marker，覆盖 upload/resume/structured-master generated client real-mode routing。 <!-- verified: 2026-05-23 method=focused-vitest evidence=frontendOwners.realApiMode.test.ts PASS; scenario scripts updated with shared real-backend gate/verify helpers -->

## Phase 7: D-20 简历扁平化适配（CreateFlow upload/paste）

> product-scope D-20 / spec D-8。

- [ ] 7.1 `ResumeCreateFlow` 删 guided tab；`ResumePreviewConfirm` 改调扁平保存（删 `confirmResumeStructuredMaster`，parse→preview→list）（验证：vitest + pixel parity PASS）
- [ ] 7.2 收口：full vitest + typecheck + build + 零 `guided`/`confirmResumeStructuredMaster`/`resumeAssetId` 残留 grep（验证：全 gate PASS + 负向 grep）
