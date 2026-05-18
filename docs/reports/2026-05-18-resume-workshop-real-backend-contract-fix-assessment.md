# Resume Workshop Real Backend Contract Fix 交付复盘报告

> **日期**: 2026-05-18
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `main` 相较 `checkpoint/016` review 中确认的 Resume Workshop real-backend drift，覆盖 create structured master、edit/manual save、branch target ID、tailor rerun version binding、version suggestions hydration。
- 关联 Bug：[BUG-0076](../bugs/BUG-0076.md)。
- 成功证据：
  - `pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/create/adapters/mapParsedSummaryToStructuredProfileDraft.test.ts src/app/screens/resume-workshop/tabs/hooks/useUpdateResumeVersion.test.tsx src/app/screens/resume-workshop/tabs/hooks/useRequestResumeTailor.test.tsx src/app/screens/resume-workshop/branch/ResumeBranchFlow.test.tsx src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx src/app/screens/resume-workshop/ResumeWorkshopScreen.test.tsx src/app/screens/resume-workshop/ResumeWorkshopAuthGate.test.tsx src/app/screens/resume-workshop/ResumeWorkshopPrivacy.test.ts`
  - `pnpm --filter @easyinterview/frontend typecheck`
  - `cd backend && go test ./internal/resume/...`
  - `cd backend && go test -tags=integration ./internal/resume/store -run 'TestResumeTailorRunStoreStateTransitionsIsolationAndClaim|TestCompleteTailorRunSuccessWritesSuggestionsAndReadyOnlyOutbox'`
  - `cd backend && go test ./...`
  - `make validate-fixtures`
  - `make lint-openapi`
  - `make docs-check`
  - `git diff --check`

## 2 会话中的主要阻点/痛点

- Fixture-backed UI tests were green while real backend validators would reject create/update requests.
  - **证据**：create mapper lacked `structuredProfile.provenance`; update mapper forwarded read-only provenance; focused frontend tests had to be added for both.
  - **影响**：real backend create/save paths could fail despite UI-level tests.
- Tailor rerun used broad asset/target inference instead of the current version identity.
  - **证据**：`RequestResumeTailorRequest` originally had no `resumeVersionId`; store async payload did not carry the version to mutate.
  - **影响**：multiple targeted versions under the same target could receive suggestions on the wrong sibling.
- Backend read-model tests did not assert suggestions hydration.
  - **证据**：`CompleteTailorRunSuccess` wrote rows, but `GetVersionByID` and `ListVersionsByAsset` did not load `resume_version_suggestions` until this fix.
  - **影响**：Rewrites tab could remain empty on real backend after successful generation.
- `make openapi-diff` produced noisy failure from older baseline drift.
  - **证据**：the run classified this change as additive but still failed on pre-existing unrelated breaking findings against `openapi-v1.0.0`.
  - **影响**：local contract verification needs manual interpretation instead of a clean pass/fail signal for this branch.

## 3 根因归类

- Real-backend request mapper assertions were too weak.
  - **类别**：spec-plan
- Async job mutation identity was under-specified in the OpenAPI request contract.
  - **类别**：spec-plan
- Store read model regression coverage did not mirror the UI consumption path.
  - **类别**：spec-plan
- OpenAPI diff baseline drift is an existing process/tooling issue, not introduced by this session.
  - **类别**：README

## 4 对流程资产的改进建议

- Add a Resume Workshop plan/review gate requiring every generated-client side-effect call to assert required fields, read-only stripped fields, and server-bound IDs in focused tests.
  - **落点**：spec-plan
  - **优先级**：high
- Add a backend resume store checklist item for read-model hydration whenever a write path persists child rows consumed by frontend tabs.
  - **落点**：spec-plan
  - **优先级**：medium
- Refresh or scope the OpenAPI diff baseline so additive branch changes are not obscured by unrelated historical drift.
  - **落点**：README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：在对应 Resume Workshop spec/plan gate 中固化 real-backend generated-client request assertions，避免 fixture-only green 再次漏掉真实 validator。
- 次优先级：处理 OpenAPI diff baseline 噪声，让 `make openapi-diff` 能重新成为分支级清晰信号。
- 本次不建议直接修改 skill 或 AGENTS.md；问题集中在当前 Resume Workshop owner gate 与 OpenAPI baseline 维护。
