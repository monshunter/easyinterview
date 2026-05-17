# Frontend Resume Workshop Branch, Rewrites and Edit Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-17

**关联计划**: [plan](./plan.md)

## Phase 0: 上游依赖 gate + retired drift baseline

- [ ] 0.1 确认 [backend-resume/002 Phase 4..8](../../../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md) 当前事实仍成立；`branchResumeVersion / requestResumeTailor / getResumeTailorRun / acceptResumeTailorSuggestion / rejectResumeTailorSuggestion / updateResumeVersion` 6 个 generated client/server surface + handler + `cmd/api` route 真实可用（验证：`rg` 读 `frontend/src/api/generated/client.ts`、`backend/internal/api/generated/server.gen.go`、`backend/internal/resume/handler/`、`backend/cmd/api/main.go`）
- [ ] 0.2 确认 `acceptResumeTailorSuggestion.json` / `rejectResumeTailorSuggestion.json` 为 `default / idempotency-replay / already-decided-409`，且 409 body 为 `error.code='VALIDATION_FAILED'` + `error.details.reason='SUGGESTION_ALREADY_DECIDED'`；如回到旧 `conflict-409` / `TARGET_INVALID_STATE_TRANSITION`，本 plan Phase 4 暂停并升级 regression blocker（验证：`jq` 读 fixture scenario keys + body）
- [ ] 0.3 确认 `requestResumeTailor.json default / idempotency-replay` 请求 header 均包含 `Idempotency-Key`，`getResumeTailorRun.json` 含 `queued / generating / default(ready) / failed` 四态；如缺失，本 plan Phase 5 / E2E.P0.085 暂停并转回 backend-resume/002 修复，不能以 synthetic schema 收口（验证：`jq` 读 fixture scenario keys + request headers）
- [ ] 0.4 确认 [frontend-resume-workshop/001](../001-listing-routing-and-detail-readonly/plan.md) 容器已就位，当前分支 [002](../002-create-flow-and-onboarding/plan.md) 实现已把 `flow=create` 替换为 `ResumeCreateFlow`，而 `flow=branch` 与 Rewrites / Edit tab 仍分别是 `<NotImplementedPlaceholder>` / `<ComingSoonTab>`（验证：grep + Vitest）
- [ ] 0.5 retired drift baseline：`git grep -nE "(^|[^A-Za-z0-9_])(inline|rewrite|mirror)([^A-Za-z0-9_]|$)" -- frontend/src/app/screens/resume-workshop/` 0 命中；`git grep -nE "welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen|onboarding=true" -- frontend/src/app/screens/resume-workshop/` 0 命中

## Phase 1: ResumeBranchFlow 容器 + 路由 + auth gate

- [ ] 1.1 修订 `frontend/src/app/screens/resume-workshop/ResumeWorkshopScreen.tsx`：`flow === 'branch'` 时渲染 `ResumeBranchFlow`，传入 `original` + `master` 上下文（验证：Vitest + grep `flow.*branch.*ResumeBranchFlow`）
- [ ] 1.2 实现 `frontend/src/app/screens/resume-workshop/branch/ResumeBranchFlow.tsx`：源级复刻 UI 真理源 Header + BRANCHING FROM 卡 + form fields + actions；表单 state + 校验 + canSubmit；至少 ≥ 8 case Vitest PASS
- [ ] 1.3 auth gate：未登录访问 `resume_versions?flow=branch&branchOriginalId={id}` 渲染 auth gate；pendingAction 仅携带 `{ flow: 'branch', branchOriginalId }`，不含 form draft（验证：Vitest mock client 0 个 protected API + pendingAction 字段集合断言）
- [ ] 1.4 originalId 解析：复用 plan 001 `listResumes` + `listResumeVersions` 拿到 `original` 与 MASTER `version`；cross-user / not-found → NotFound CTA（验证：Vitest）
- [ ] 1.5 i18n key 空间脚手架 `resumeWorkshop.branch.*` 初始化在 en/zh（验证：切换不报错）

## Phase 2: branchResumeVersion 三 seedStrategy 提交 + IK + nav 行为

- [ ] 2.1 实现 `branch/hooks/useResumeBranchSubmit.ts`：`generateIdempotencyKey()` + generated client `branchResumeVersion` + 三态响应处理 + 错误映射；至少 ≥ 8 case Vitest PASS
- [ ] 2.2 BranchFlow "创建版本" 触发 hook；nav target 按 seedStrategy 三态分发：copy_master → rewrites tab；blank → edit tab；ai_select → rewrites tab + polling 启动（验证：Vitest nav target 断言）
- [ ] 2.3 mapper `mapBranchFormToBranchResumeVersionRequest`：表单字段 → `BranchResumeVersionRequest`；focus enum 字面量映射（验证：Vitest mapper 至少 ≥ 6 case）
- [ ] 2.4 fixture parity test：`branchResumeVersion.json` `default` / `copy-master-sync` / `blank-sync` / `ai-select-202-with-job` / `idempotent-replay` / `validation-error-422` 与 hook 字节匹配（验证：fixture parity test PASS）
- [ ] 2.5 IK request spy：`Idempotency-Key` header 出现且同一表单 retry 复用至成功或 422 inline（验证：Vitest spy）
- [ ] 2.6 失败路径：422 inline / 404 parent / 404 targetJob / IK conflict 409 toast（验证：Vitest）

## Phase 3: Rewrites Tab UI + getResumeVersion 投影

- [ ] 3.1 实现 `frontend/src/app/screens/resume-workshop/tabs/ResumeRewritesTab.tsx`：源级复刻 UI 真理源 scope banner + 左侧列表 + 右侧 diff Card + 顶部计数 chip；至少 ≥ 8 testid（`resume-rewrites-scope-banner` / `-bullet-list` / `-bullet-row-{id}` / `-diff-card` / `-action-{accept,reject,edit}` / `-status-chip-{accepted,pending,rejected}` / `-rerun-tailor`）（验证：Vitest）
- [ ] 3.2 数据来源：plan 001 `useResumeVersion(versionId)` 返回的 `version.suggestions[]` + adapter `mapBulletSuggestionToUi` 扩展含 `status / decidedAt / source`（验证：Vitest adapter 至少 ≥ 6 case）
- [ ] 3.3 计数派生：accepted / pending / rejected 三类计数从 `suggestions[]` 派生，不写死（验证：Vitest 数量断言）
- [ ] 3.4 选中切换：`selectedBulletId` 切换时取消任何 inline edit；列表行截断展示 90 字符，完整字段仅在 diff Card 渲染（验证：Vitest）
- [ ] 3.5 隐私：DOM 渲染 original / rewritten 文本；URL / pendingAction / localStorage / mock transport log / telemetry 不含 originalBullet / suggestedBullet 文本（验证：Vitest spy grep）
- [ ] 3.6 fixture parity test：`getResumeVersion.json` `targeted-with-suggestions` scenario 与 RewritesTab 渲染字节匹配（验证：fixture parity test）

## Phase 4: 单条 suggestion accept / reject / manual edit 终态

- [ ] 4.1 实现 `tabs/hooks/useAcceptResumeTailorSuggestion.ts`：`generateIdempotencyKey()` + generated client `acceptResumeTailorSuggestion` + 三态错误映射；至少 ≥ 8 case Vitest PASS
- [ ] 4.2 实现 `tabs/hooks/useRejectResumeTailorSuggestion.ts`：同形态调 `rejectResumeTailorSuggestion`；至少 ≥ 8 case Vitest PASS
- [ ] 4.3 inline manual edit：UI 真理源 Edit / Cancel / Save manual edit 三按钮 + textarea；保存触发 `updateResumeVersion` patch `structuredProfile.manualEdits[]`，成功后再调用 bodyless `acceptResumeTailorSuggestion` 标记终态；update 成功但 accept 失败时显示 saved-manual-pending retry（验证：Vitest update→accept、update 422 不触发 accept、accept failure retry 三路径）
- [ ] 4.4 状态机断言：terminal 状态 accept / reject 都是终态；accept / reject request body 为 `undefined`；再次 accept / reject 走 IK replay 或返回 409 already-decided；不同 fingerprint 同 key 409 generic IK conflict；不自动 patch `version.structured_profile`（D-12 同步）（验证：Vitest 终态 + IK replay/conflict + structured_profile DOM 不变 + request body spy）
- [ ] 4.5 fixture parity test：`acceptResumeTailorSuggestion.json` `default / idempotency-replay / already-decided-409`、`rejectResumeTailorSuggestion.json` 同形态 scenario 与 hook 字节匹配（验证：fixture parity test PASS；如 fixture 回到旧 envelope，本步骤保持 blocked，转回 backend-resume/002 修复）
- [ ] 4.6 cross-user / 404 / 422：toast generic + inline error；不暴露原 envelope（验证：Vitest）

## Phase 5: requestResumeTailor + tailor run polling

- [ ] 5.1 实现 `tabs/hooks/useRequestResumeTailor.ts`：`generateIdempotencyKey()` + generated client `requestResumeTailor`；mode ∈ `{gap_review, bullet_suggestions}` 与 backend D-5 对齐（验证：Vitest IK + 错误映射）
- [ ] 5.2 实现 `tabs/hooks/useResumeTailorRunPolling.ts`：指数退避轮询 `getResumeTailorRun(tailorRunId)`（初始 1500ms / backoff 1.4x / max 12 attempt / ~60s 上限）；终态 ready → refetch getResumeVersion；failed / timeout → 失败 banner（验证：Vitest happy / failed / timeout / cancel 至少 ≥ 6 case PASS）
- [ ] 5.3 fixture-backed polling harness：使用 `getResumeTailorRun.json queued / generating / default(ready) / failed` 组成 deterministic sequence；只允许 mock 调度顺序，不 mock response schema（验证：Vitest harness 切换断言）
- [ ] 5.4 Rewrites Tab UI 集成：ai_select branch 完成后渲染 polling banner；用户 "重新运行改写" 触发同 banner；ready 后消失 + 列表刷新；failed 后红 banner + 重试 CTA（验证：Vitest 状态切换 + 断言）
- [ ] 5.5 IK 行为：requestResumeTailor 携带 IK；getResumeTailorRun 不携带 IK（验证：Vitest spy 双向断言）
- [ ] 5.6 cleanup：组件 unmount 时取消 polling；切换 tab 时 polling 在父 detail container 维持或在 unmount 取消，避免泄漏（验证：Vitest cleanup + Playwright network sniff）

## Phase 6: Edit Tab + updateResumeVersion 保存

- [ ] 6.1 实现 `frontend/src/app/screens/resume-workshop/tabs/ResumeEditTab.tsx`：源级复刻 UI 真理源 top banner + headline input + summary textarea + experience section placeholder + skills section placeholder + 保存改动按钮（验证：Vitest DOM ≥ 10 testid + master vs targeted scope banner i18n）
- [ ] 6.2 实现 `tabs/hooks/useUpdateResumeVersion.ts`：`generateIdempotencyKey()` + generated client `updateResumeVersion` + 错误映射；mapper 过滤不可编辑字段（验证：Vitest mapper + happy / 422 / 409 至少 ≥ 8 case PASS）
- [ ] 6.3 P0 实际可编辑字段：headline + summary；experience / skills section 仅 placeholder 渲染，Add 按钮 toast `敬请期待`（验证：Vitest Add click + toast 断言）
- [ ] 6.4 保存后行为：toast + 触发 `getResumeVersion(versionId)` refetch；不刷新整页路由（验证：Vitest）
- [ ] 6.5 fixture parity test：`updateResumeVersion.json` `default / idempotency-replay / validation-error-422` 与 hook 字节匹配（验证：fixture parity test PASS）
- [ ] 6.6 隐私：DOM 渲染 structuredProfile fields 但 URL / pendingAction / localStorage / mock transport log 不含字段内容（验证：Vitest spy grep）

## Phase 7: i18n + a11y + 隐私 + UI parity + BDD + 旧入口负向

- [ ] 7.1 i18n key 空间完整：`resumeWorkshop.branch.*` / `.rewrites.*` / `.edit.*` / `.tailor.*` namespace 在 en/zh 落齐；EN / ZH 切换关键文案 + Accept-Language header 携带 7 个 op 请求（验证：Vitest + integration test）
- [ ] 7.2 a11y：BranchFlow form a11y + Rewrites Tab listbox/option + Edit Tab labels + scope banner aria-live；Playwright axe-core check PASS（验证：Playwright a11y spec）
- [ ] 7.3 隐私红线 grep：originalBullet / suggestedBullet / matchSummary / structuredProfile / manual edit text / form draft 不出现在 console / URL / pendingAction / localStorage / mock transport log / telemetry / toast；pendingAction 仅 route + 必要 params（验证：Vitest spy grep + Playwright DOM/network sniff）
- [ ] 7.4 UI parity gate：新增 `frontend/tests/pixel-parity/resume-workshop-branch-rewrites-edit.spec.ts` 覆盖 BranchFlow / Rewrites Tab / Edit Tab desktop 1440px + mobile 390x844 DOM anchor + computed style + bounding box + 非空截图 buffer；clean checkout PASS 不依赖未跟踪 baseline（验证：`pnpm --filter @easyinterview/frontend build && pnpm --filter @easyinterview/frontend test:pixel-parity` PASS）
- [ ] 7.5 Export PDF / copyText 一致性：在 Rewrites / Edit Tab 顶 header 保留 plan 001 Export PDF / 复制纯文本按钮；切换 tab 不影响 IK header / 501 toast / clipboard 行为（验证：Vitest + Playwright 重跑 P0.037 关键断言）
- [ ] 7.6 BDD-Gate: E2E.P0.084 resume-branch-flow-three-seed-strategies PASS（详见 [bdd-checklist.md](./bdd-checklist.md)）
- [ ] 7.7 BDD-Gate: E2E.P0.085 resume-rewrites-tab-tailor-run-polling PASS
- [ ] 7.8 BDD-Gate: E2E.P0.086 resume-suggestion-accept-reject-edit-and-update-version PASS
- [ ] 7.9 BDD-Gate: E2E.P0.087 resume-detail-export-copy-consistency-and-parity PASS
- [ ] 7.10 旧入口 grep（收口）：`git grep -nE "welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen|onboarding=true" -- frontend/src/app/screens/resume-workshop/branch/ frontend/src/app/screens/resume-workshop/tabs/` 0 命中（验证：CI lint）
- [ ] 7.11 retired tailor mode grep：`git grep -nE "(^|[^A-Za-z0-9_])(inline|rewrite|mirror)([^A-Za-z0-9_]|$)" -- frontend/src/app/screens/resume-workshop/branch/ frontend/src/app/screens/resume-workshop/tabs/` 0 命中（验证：CI lint）
- [ ] 7.12 prototype import grep：`git grep -nE "ui-design/src/(data|screen-resume-workshop)" -- frontend/src/app/screens/resume-workshop/branch/ frontend/src/app/screens/resume-workshop/tabs/` 0 命中（验证：CI lint）
- [ ] 7.13 在 `test/scenarios/e2e/INDEX.md` 追加 P0.084 + P0.085 + P0.086 + P0.087 行（关联需求 `frontend-resume-workshop C-11`，状态 Ready，automated）
- [ ] 7.14 spec / history / INDEX 同步核对：frontend-resume-workshop spec.md / history.md / `docs/spec/INDEX.md` 已保持 1.1，§6 C-11 / §7 plan 003 行指向当前 active plan，§3.2 accept/reject 口径为 UI 真理源 inline action + terminal-state feedback；`docs/spec/frontend-resume-workshop/plans/INDEX.md` 已包含 003 active 行；不为 checklist 收口重复 bump spec，除非发现新的设计事实（验证：`sync-doc-index --check` PASS）
