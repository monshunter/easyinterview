# 001 Workspace + InterviewContext + Start Practice Contract

> **版本**: 1.49
> **状态**: completed
> **更新日期**: 2026-07-19

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)
**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)
**关联 BDD Checklist**: [bdd-checklist](./bdd-checklist.md)

## 1 目标

Phase 26 显式 supersede v1.19 / Phase 14 的 pure-list/card-to-Parse 结论：`/workspace` 无 `targetJobId` 时展示 ready 规划列表；`/workspace?targetJobId` 时展示统一只读规划详情；ready card 直接进入 workspace detail。Parse 只属于刚导入规划的 command/progress，ready replace/back 由 frontend-shell/004 Phase 12 owning。
本次 v1.13 原地修订追加面试规划列表卡片化视觉收口：列表项必须是清晰可感知的卡片，而不是无容器的文本列。
本次 v1.14 原地修订追加面试规划列表卡片简化：卡片不得展示 `sourceType` / `targetLanguage` 等低价值技术字段，底部只保留主题色进入 CTA，并使用现有卡片 / 边框 / elevation token 拉开与页面背景的层次。
本次 v1.15 原地修订修复 `进入规划` 回归：列表卡片必须消费后端投影的当前 practice plan / bound resume 信息，导航到统一面试规划详情时携带真实 `planId` / `resumeId`，不得回落到缺简历空态或合成不存在的 plan/resume id。
本次 v1.16 原地修订修复方案 A 数据持久化缺口：创建 JD/规划时选中的简历必须作为 `target_jobs.resume_id` 持久化；`listTargetJobs` / `getTargetJob` 的 `resumeId` 优先来自 target job 级绑定，列表卡片重新进入规划时即使尚未创建 `practice_plans` 也必须携带真实 `resumeId`。
本次 v1.17 原地修订曾将 workspace 普通回访复用 `frontend-home-job-picks-and-parse` 的 Parse-derived 母版；该过渡合同已被 v1.19 取代，当前 workspace 不再渲染详情或执行 create/start session。
本次 v1.18 原地修订修复面试列表准入与顶栏导航回归：no-context workspace 必须只以当前 route params 判定，不得继承 stale `InterviewContext`；列表请求 `listTargetJobs` 必须带 `analysisStatus=ready`，并过滤 failed / 空标题 TargetJob，防止解析失败脏数据进入面试列表。
历史 v1.19 曾把 `workspace` 收敛为纯列表并让卡片导航 Parse；该结论已由本计划 Phase 26 明确 supersede，不再作为当前实现合同。
本次 v1.20 原地修订修复面试列表卡片规格回归：desktop plan-list grid 必须使用固定最大列宽，1/2/3 张卡片的规格保持稳定，不得因单卡数量被拉伸为整行宽卡。
本次 v1.21 原地修订融合 Home 最近模拟面试与 workspace 面试列表卡片：workspace 卡片必须复用 Home recent card 的主体结构、公司、岗位、可选真实地点和 mini round rail。本次 v1.22 原地修订把列表卡片的 `进入规划` 可见 footer CTA 改为点击卡片主体承接，并增加 `立即面试` 主按钮和使用简历列表 trash 图标样式的删除能力；Home recent 复用同一卡片动作模型但不展示删除按钮。本次 v1.45 原地修订移除同一 `TargetJob.status` 的重复展示和空地点 `Location not set` 占位，真实地点仍按原层级展示。
本次 v1.46 原地重开 Phase 30，修复共享 `startPracticeFromParams` 在等待 session opening LLM 时仅禁用入口按钮、页面看似卡死的实现漂移：四类正式入口必须共享同一全屏面试准备过渡态，并在成功/失败、可访问性与 reduced-motion 路径中一致收敛。
本次 v1.47 原地重开 Phase 31：Practice 不再隐藏 App chrome；全局 TopBar 与独立 Practice Session Header 同时存在，并复用 frontend-shell 的内存 runtime/display context，route 切换零额外 `/me`。
本次 v1.48 原地重开 Phase 32：按用户提供的 1916×821 面试列表参考稿重构 query-free Workspace 主体、双列宽卡、轮次 rail、footer 与响应式几何；共享 TopBar 维持 frontend-shell owner，业务 route、API、归档与启动合同不变。
本次 v1.49 补充 Phase 32 背景覆盖与对齐修正：全视口页面背景与 1508px 居中内容必须拆分为两层，禁止背景伪元素被限宽容器的 overflow 裁剪后在右侧留下空白带；header 与 card grid 共享 1456px 右边界，使新建 CTA 与第二列卡片右侧对齐。
本次 v1.38 原地修订收口结构化轮次目录与时长一致性：`TargetJob.summary.interviewRounds[]` 是轮次顺序和规划时长的唯一来源；`PracticePlan.timeBudgetMinutes` 保存所选轮次时长快照，Practice Top Bar 从 plan 读取预算；报告下一轮只取有序列表的紧邻后一项，末轮、空/未知轮次、加载失败和重复点击 fail closed，不再使用固定 `25:00`、固定轮次表或默认回退。当前轮的持久化事实由 v1.39 `practiceProgress` 接管。
本次 v1.39 按方案 A 把轮次进度事实收回后端：Home/Workspace/Parse/Report 只消费 `TargetJob.practiceProgress`，不再用 TargetJob lifecycle `status`、自由文本、时长或浏览器状态猜测当前轮；计划只按 exact round pair 复用，全部完成后启动 fail closed。

- TopBar `workspace` 文案改为 `面试` / `Interview`，route/testid 仍保持 `workspace`。
- `/workspace` 无 target 时使用 generated `listTargetJobs(analysisStatus=ready)` 渲染列表；有 target 时使用 generated `getTargetJob` 渲染统一只读详情。
- 面试规划列表每个 plan item 必须具备独立卡片容器，并以 Home 最近模拟面试卡片主体为主：公司、岗位、可选真实地点和 mini round rail 保持同源；desktop 使用固定最大列宽的响应式多列，1/2/3 张卡片规格保持稳定，mobile 折叠为单列；workspace 只在同一卡片底部追加 `立即面试` 主按钮，删除图标固定在卡片右上角。
- 面试规划列表卡片只展示对继续规划有决策价值的信息：岗位、公司、非空真实地点和 backend `practiceProgress` 轮次；不展示 TargetJob lifecycle `status`、`Location not set`、`手动输入` / 来源类型 / 目标语言等导入元信息；不展示可见的 `进入规划` footer button。
- 面试规划列表只展示已解析成功且具备岗位标题的 TargetJob：generated `listTargetJobs` 请求必须带 `analysisStatus=ready`，UI 层必须防御性排除 failed / processing / queued / 空标题记录。
- 顶栏 `面试` 进入 query-free `/workspace` 列表；卡片主体进入 `/workspace?targetJobId=...` 详情。`planId`/`resumeId`/auto-start 等非安全 query 被 shell 剔除。
- 卡片详情 route 只携带 `targetJobId`；绑定 resume/plan/round 事实由 detail `getTargetJob` response 恢复，不从 list item/query 复制。
- Workspace detail 复用统一只读母版，不拥有 `autoStartPractice` route side effect；列表 quick-start 仍使用 shared generated practice handoff。
- Workspace detail 删除独立 Interview Launch/绑定简历大卡片：标题旁的“绑定简历”只使用 `getTargetJob` 保存的 `resumeId` 进入 `resume_versions` 详情；标题下首行动作行从左依次展示“立即面试”和“面试报告”，desktop 同排、mobile 同序换行；不得新增 `getResume` 预读、route resume authority 或 in-place rebind。
- 列表页删除图标使用 generated `archiveTargetJob` 和 `Idempotency-Key` 持久软归档 TargetJob；成功后从当前列表移除，失败时不导航、不隐藏卡片，并展示错误；不得继续使用本地-only hidden set 作为删除合同。
- `InterviewContext` 不在 `workspace` route carry；`practice / generating / report` owner route 按各自最小上下文携带稳定 ID 与 `practiceGoal`，不携带 mode/modality/hint 状态。
- Workspace runtime 保留 list + read-only detail 两态；不包含 Plan Switcher、Resume Picker 或 route-side 启动副作用。
- 当前规划记录区只展示 typed records static affordance，不从 `TargetJob` fixture extension、`any` 或 report API 拼接记录行。
- JD 原文、简历正文、题目文本、答案、提示、prompt/response 不进入 URL、localStorage、console 或 fixture transport 日志。

## 2 当前合同

### 2.1 Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|----------|
| `listTargetJobs` | current list/progress fixtures | Workspace list rail | backend-targetjob list owner | TargetJob + completion projection | none | `E2E.P0.098` 仅 progress refresh |
| `getTargetJob` | current detail/progress fixtures | Workspace detail | backend-targetjob detail owner | TargetJob requirements/progress | none | `E2E.P0.098` 仅 progress/detail read |
| `createPracticePlan` / `getPracticePlan` | current plan fixtures | quick-start/start helpers | backend-practice plan owner | practice plans | none | 当前无真实 E2E owner；root `make test` |
| `getResume` | `openapi/fixtures/Resumes/getResume.json` | Resume detail owner only；Workspace detail 不消费 | `backend-resume/001` | `resumes` | none | external owner gates |
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` | Home selector + Resume Workshop；Workspace/Parse detail 不消费 | `backend-resume/001` | `resumes` summary projection | none | Home/Resume owner gates |
| `startPracticeSession` | `openapi/fixtures/PracticeSessions/startPracticeSession.json` | Home recent、Workspace list/detail、Report replay/next-round 通过 shared start 启动；等待期间共享 UI-only transition，成功后导航 `practice` | `backend-practice/001` | `practice_sessions` + opening `practice_messages` row | backend-only `practice.session.chat` | 当前无真实 E2E owner；Phase 30 domain behavior + root `make test` |
| `getFeedbackReport` | N/A | 本 plan 不消费；report owner handles replay/next-round CTA | external owner | external | none | external owner gates |

### 2.2 UI / Route Boundary

- `workspace` 与 `practice` 保留 App chrome；只有 `generating` 由下游 owner 隐藏 chrome。Practice 会话控件属于独立 Practice Session Header。
- TopBar 显示 `面试` / `Interview`；query-free workspace 是列表，可选 `targetJobId` 是详情；Parse 不是 ready-card route。
- 当前面试规划不展示练习模式卡片、成长中心、单题深钻、专项练习、独立 voice route 或独立公司信号页面。
- `resumeId` 是当前简历绑定键；`resumeVersionId` 不作为本 plan 正向 route/context 字段。
- Records static affordance 只说明当前规划下的模拟面试记录区域存在；真实记录行必须来自 typed records contract owner。

## 3 质量门禁分类

- **Plan 类型**: `feature-behavior + contract + frontend-ui + BDD`。
- **TDD 策略**: 适用。Vitest 覆盖 route hydration、InterviewContext reducer、ordered round resolver、plan time-budget create/reuse、Practice plan budget display、report next-round/last-round/unknown-round/double-click handoff、四类启动入口的 pending/success/failure transition、generated client body/header、auth pendingAction、privacy and out-of-scope negative gates。
- **BDD 策略**: `BDD.WORKSPACE.CONTEXT.001` 由代码层 owner tests 验证 list/detail、后端 progress 投影、exact-plan reuse 与 fail-closed 行为；`BDD.WORKSPACE.CARD.003` 验证卡片元信息；`BDD.PRACTICE.LAUNCH.004` 由四类 caller domain behavior tests 验证启动等待反馈。三者由仓库根 `make test` 统一回归；`E2E.P0.098` 仅作为 completion/progress refresh 的独立真实环境 handoff，只有显式运行后才产生 PASS，且不承接 quick-start/session start/next-round。
- **替代验证 gate**:
  - `pnpm --filter @easyinterview/frontend test src/app/screens/workspace src/app/screens/parse/ParseResumeBinding.test.tsx src/app/screens/report/__tests__/ReplayCta.test.tsx src/app/App.test.tsx`
  - `pnpm --filter @easyinterview/frontend test`
  - `make validate-fixtures`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/context.yaml --target frontend`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

## 4 实施结果

### Phase 0: Contract preflight

- Confirmed `docs/development.md` §2 frontend/backend contract workflow.
- Confirmed UI design document: `docs/ui-design/module-job-workspace.md`, `frontend/src`.
- Confirmed generated client operations and fixtures listed in §2.1.

### Phase 1: Workspace shell and InterviewContext

- Added `InterviewContextProvider`, reducer actions, route hydration and owner-route carry/clear behavior.
- Replaced the `workspace` route fallback shell with `WorkspaceScreen`; kept `practice` / `generating` for their owner plans.
- Added zh/en `workspace.*` messages and DOM anchor coverage.

### Phase 2: TargetJob, resume and workspace data

- Wired `getTargetJob` into header, launcher, JD breakdown, preparation signals and empty/error states.
- Wired `getResume` into bound resume summary and missing-resume state.
- Preserved source-level layout parity for desktop and mobile.

### Phase 3: Plan and resume switching (moved out of workspace owner)

- Historical implementation added workspace-local plan/resume modals.
- v1.19 removes those runtime files from workspace owner; parse detail owns resume selection and workspace cards only open `parse`.
- Modal keyboard / focus behavior is no longer a workspace acceptance gate.

### Phase 4: Start practice and auth recovery (moved out of workspace)

- Historical implementation added workspace-local start-practice hooks.
- v1.19 moves the two-step launch to parse/report handoff owners through shared `startPracticeFromParams`.
- `workspace(autoStartPractice=1)` is no longer a valid side-effect contract; URL codec strips all workspace context params.

### Phase 5: Workspace boundary and privacy

- Workspace runtime does not own company insight or records implementations.
- Report and record data remain with their typed owner contracts.
- Negative coverage remains for sensitive fields and out-of-scope runtime routes/testids.

### Phase 6: Verification closeout

- Focused tests remain development feedback；phase completion uses repository-root `make test`, with formal component/responsive assertions, fixture validation, docs/index checks and negative grep as separate gates. Only `E2E.P0.098` owns the real completion/progress-refresh flow described in the BDD plan.

### Phase 7: Interview nav and plan-list landing revision

- Update product/UI design documents and static prototype so TopBar uses `面试` / `Interview` and `workspace` separates plan-list landing from current-plan detail.
- Add `WorkspacePlanList` backed by generated `listTargetJobs`; v1.19 changes plan-card navigation to `parse?targetJobId=...` without fabricating resume or report data.
- Detail, not-found and missing-resume states are parse owner responsibilities, not workspace list states.
- Update source-structure, i18n, route, scenario and responsive-browser tests so TopBar click proves the list landing and hydrated route still proves the current-plan detail.

### Phase 8: Plan-list card visual hardening

- Reopen the completed owner after screenshot review because Phase 7 verified list presence but did not assert visual card affordance.
- Update `docs/ui-design/module-job-workspace.md` and `frontend/src` so the plan list card contract includes visible card background, 1px border, subtle elevation, internal body/footer sections and responsive card grid.

### Phase 10: Plan-list bound resume navigation remediation

- Reopen the completed owner after runtime bug report because Phase 7-9 preserved card visuals but allowed list card navigation to omit the existing practice-plan resume binding.
- Add regression coverage that `WorkspacePlanList` uses declared TargetJob projection fields for current plan id and bound resume id, and does not synthesize `plan-${targetJobId}` or `resume-unbound`.
- Keep the no-binding fallback explicit: route carries only target job identity and detail page owns missing-resume recovery.
- Verified focused frontend suites and typecheck on 2026-07-08.

### Phase 11: TargetJob-level resume binding remediation

- Reopen the completed owner because runtime evidence showed imported target jobs can have no `practice_plans` rows yet, so deriving `resumeId` only from latest ready practice plan still loses the resume selected during Home import.
- Require `importTargetJob` to persist the selected `resumeId` to `target_jobs.resume_id`; `TargetJob.resumeId` must expose this target job binding even before a practice plan exists.
- Update `WorkspacePlanList` tests so a list item with `resumeId` but no `currentPracticePlanId` opens current-plan detail with the real resume binding and does not show the missing-resume state.
- Update formal `WorkspacePlanList` to source-level mirror that card treatment while keeping generated `listTargetJobs` data flow and safe plan navigation unchanged.
- Add jsdom and Playwright responsive-browser assertions that fail when plan items render as loose text columns or lose card elevation/sectioning.

### Phase 12: Unified detail route remediation

- Reopen the completed owner because user review confirmed that the standalone workspace detail and Parse result page duplicate the same JD / resume / round confirmation job.
- Keep `WorkspacePlanList` as the `面试` landing; v1.19 supersedes this phase by routing ordinary detail traffic to `parse` instead of `workspace?targetJobId=...`.
- Start-practice side effects are now owned by parse/report handoff code, not workspace `autoStartPractice`.
- Add regression coverage that the out-of-scope independent workspace detail anchors (`workspace-header`, `workspace-launcher`, `workspace-jd-card`, `workspace-prep-card`, `workspace-history-card`) do not appear on ordinary detail re-entry, while `workspace-plan-list` and start-practice gates remain live.

### Phase 13: Plan-list admission and stale-context navigation remediation

- Reopen the completed owner because runtime evidence showed parse-failed / blank TargetJob rows could appear as interview-plan cards, and TopBar `面试` from a detail page reused stale context.
- Require `useWorkspaceTargetJobs` consumers to request `analysisStatus=ready` and require `WorkspacePlanList` to ignore non-ready or blank-title records if stale data still appears.
- Align route split with `frontend/src`: workspace ignores stale `InterviewContext` and out-of-scope route params, so it cannot produce “缺少目标岗位 ID”.

### Phase 14: Workspace route purity remediation

- Reopen the completed owner because `workspace` is a REST list page, but historical detail/start hooks left it acting as a parameterized context route.
- Remove workspace detail/start/modal runtime files and tests from the current owner; keep only `WorkspacePlanList` + `useWorkspaceTargetJobs`.
- Change plan-card navigation and static UI design document to open `parse`, not `workspace?targetJobId=...`.
- Remove `workspace` from `INTERVIEW_CONTEXT_ROUTES`; App route sync clears context whenever route name is `workspace`.
- Move start-practice side effects to parse/report handoff owners through generated `getPracticePlan` / `createPracticePlan` / `startPracticeSession`, with `Idempotency-Key` on side effects and no `autoStartPractice` workspace hop.

### Phase 15: Plan-list card size stability

- Reopen the completed owner because screenshot review showed the desktop list grid stretched a single plan card across the full content width.
- Update `docs/ui-design/module-job-workspace.md` and `frontend/src` so the plan-list grid uses `auto-fill` with a fixed maximum column width instead of `auto-fit + 1fr`.
- Update formal `WorkspacePlanList` and focused tests so the grid contract rejects `1fr` desktop stretching while keeping compact mobile single-column behavior.
- Verify with focused Vitest, source/UI design document tests, typecheck, build, and browser screenshots.

### Phase 16: Home recent card / workspace list card fusion

- Reopen the completed owner because user review asked the Home "最近模拟面试" card and Interview list card to become one visual object instead of two separate card systems.
- Update `docs/ui-design/module-job-workspace.md` and `frontend/src` so workspace plan cards use the Home recent card body, including mini round rail driven by `TargetJob.summary.interviewRounds[]`.
- Update formal `WorkspacePlanList` to reuse the Home recent card component/body while appending the workspace-specific footer CTA; the card grid keeps the fixed `360px` desktop max width from Phase 15.
- Add focused regression coverage that fails when workspace cards lose the home recent mini rail or reintroduce a separate workspace-only body.

### Phase 17: Plan-list action row and card-click planning

- Reopen the completed owner because user review asks the visible `进入规划` footer CTA to become invisible and be replaced by clicking the card itself.
- Update `docs/ui-design/module-job-workspace.md` and `frontend/src` so workspace cards append `立即面试` and a top-right trash icon delete action; the card root remains the planning-detail navigation control.
- Update formal `WorkspacePlanList` to start practice directly through shared generated practice handoff with structured `roundId/roundName` when `立即面试` is clicked, keep delete isolated from card navigation, and let Phase 18 own backend-persistent archive behavior.
- Add focused regression coverage that fails when `进入规划` appears as a visible footer button, when Home recent shows a delete action, when delete triggers navigation/backend deletion, or when `立即面试` opens the planning detail instead of starting practice.

### Phase 18: Persistent TargetJob archive integration

- Reopen the completed owner because backend-targetjob now owns `archiveTargetJob`, so workspace delete must no longer be local-only.
- Update `WorkspacePlanList` to call generated `archiveTargetJob(targetJobId)` with an `Idempotency-Key`; only remove the card after the backend returns success.
- Preserve card-click planning and quick-start propagation boundaries: delete and quick-start must stop bubbling to card navigation.
- Add focused regression coverage that fails when delete is implemented via local-only hidden state, when `archiveTargetJob` is missing from generated client usage, when delete success does not remove the card, when delete failure hides the card, or when Home recent renders a delete control.
- Add real-backend browser smoke and screenshot proof that an archived TargetJob disappears from workspace after refresh.

### Phase 19: Remove non-executable workspace detail/start scenarios

- Current workspace runtime is a pure plan-list owner; detail loading moved to parse, while parse/report focused gates own direct practice start and auth recovery.
- Add a repository-wide scenario contract test requiring every explicit frontend Vitest/Playwright path in trigger scripts to resolve to a tracked file.

### Phase 20: Remove obsolete auto-start context state and implicit route-param carry

- Delete `autoStartPractice` from `InterviewContextState`, defaults and route hydration, and delete the unreferenced `CLEAR_AUTO_START` action.
- Replace `startPracticeFromParams` arbitrary input spreading plus one-key stripping with an explicit current practice-route output assembled from normalized context, new plan/session IDs and optional language.

### Phase 21: Remove test-only InterviewContext reducer actions

- Delete `MERGE_TARGET_JOB`, `MERGE_RESUME`, `MERGE_PRACTICE_PLAN`, `CLEAR_RESUME` and `CLEAR_PRACTICE_PLAN`; repository search proves they have no production dispatch sites.
- Keep only runtime-used `HYDRATE_FROM_ROUTE`, `MERGE_SESSION`, `INCREMENT_HINT_COUNT` and `CLEAR` behavior.
- Replace unit tests that manufacture unused actions with a source-surface negative assertion plus the existing runtime-action behavior tests.
- BDD is not applicable because no executable user path uses the removed actions. Alternative gates: focused reducer red/green, production dispatch inventory, full frontend tests/typecheck/build and owner/global gates.

### Phase 22: Remove unconsumed InterviewContext hook

- Delete `useStartPracticeContext`; repository-wide AST and text inventory prove the export has no production or test consumer.
- Keep `useInterviewContext` as the single runtime access API and leave `InterviewContextState` behavior unchanged.
- BDD is not applicable because the removed hook has no executable caller. Alternative gates: focused source-surface red/green, repository symbol inventory, full frontend tests/typecheck and owner/global gates.

### Phase 9: Plan-list card simplification and theme consistency

- Reopen the completed owner after screenshot review because Phase 8 still rendered low-value source/language metadata and a secondary CTA that visually competed with the theme.
- Update `docs/ui-design/module-job-workspace.md` and `frontend/src` so no-context plan cards are concise: status + updated date, title, company/location, and a theme accent `进入规划` / `Open plan` CTA only.
- Update formal `WorkspacePlanList` to remove source/language display, keep footer as a minimal action row, and strengthen separation from page background via existing card/rule/elevation tokens.

### Phase 23: Remove unreachable static Workspace detail branch

- Replace the static prototype `WorkspaceScreen` constant-false context split with the current pure `WorkspacePlanList` entrypoint.
- Delete the unreachable detail DOM, its exclusive context/history/modal/requirement helpers, and the now-unconsumed `screen-workspace-insight.jsx` source/script entry; move the one live Parse binding-pill consumer into its owner as a smaller local component, and retain `getWorkspaceResumeOptions` because Home and Parse still consume it.
- Replace UI contract assertions that kept the old detail branch alive with explicit zero-residual assertions for `hasPlanContext`, `PlanSwitcherModal`, `ResumePickerModal` and the old detail helpers.
- Reconcile the active workspace/practice spec from the stale embedded-insight contract to the current pure list boundary.

### Phase 24: Structured round runtime consistency

- Update the Practice/report UI design document so the visible budget comes from the selected structured round and next-round behavior uses the same ordered list.
- Add RED tests proving the current hard-coded `25:00`, fixed `ROUND_ORDER`, unknown-round fallback and repeated-click behavior are incorrect.
- Resolve the selected round once through the shared round assumptions; write its `durationMinutes` into `CreatePracticePlanRequest.timeBudgetMinutes`, and display the persisted plan budget in Practice. Phase 25 supersedes the old duration-only reuse predicate with exact persisted round-pair reuse.
- Resolve next round as the immediate existing successor in the fetched TargetJob round list. Sequence must be positive int32, unique and strictly increasing, but gaps such as `1,2,4` are valid: round `2` advances to round `4`, never a fabricated `3`. Disable next-round while round data is loading, when derived round IDs are duplicated, when the current round is missing/unknown, at the final/single round, or while a start is in flight; never fall back to the first or a fabricated default round.
- Keep elapsed time as informational budget progress: no automatic completion, no TargetJob status mutation, no new OpenAPI/schema field.
- Closeout evidence: [BUG-0161](../../../../bugs/BUG-0161.md) and [structured-round runtime consistency assessment](../../../../reports/2026-07-12-structured-round-runtime-consistency-assessment.md).

### Phase 25: Backend-persisted round progress and exact plan reuse

- Update `frontend/src` and shared round helpers so Home/Workspace use `practiceProgress.completedRounds/currentRound`; delete `nextRound` and lifecycle-status/text fallbacks. Keep `MiniRoundRail` DOM, style tokens, bounding boxes and responsive geometry unchanged. All completed renders every node done and disables quick-start; invalid/missing projection renders no false current state and fails closed.
- Replace `roundIndexFromTargetJobStatus` with a strict progress mapper. It validates exact round pair, positive int32 strictly increasing/unique canonical sequences without requiring contiguity, completed prefix, current first-incomplete and final completed/null state. Navigation derives `roundId/roundName` only from valid current progress; changing `TargetJob.status` cannot change the result.
- `buildCreatePlanRequest` sends `roundId` only; the backend derives sequence. Shared start reuses a plan only when ready target/resume/roundId/roundSequence exactly match, with duration as an integrity check. Equal-duration adjacent rounds and legacy null identity must create/validate a new plan. A new response with a mismatched pair cannot start a session.
- Home/Workspace/Parse quick-start must target `practiceProgress.currentRound`; final/invalid progress disables start with zero plan/session calls. Report next-round is enabled only when the next existing canonical array item exactly equals the backend current round; it must not compare with `sequence + 1`. Retry-current-round remains allowed and server-validated.
- Real PostgreSQL projection tests and focused frontend tests are development feedback. Code-level `BDD.WORKSPACE.CONTEXT.001` closes this owner through repository-root `make test`. When explicitly run, `E2E.P0.098` validates only the real completion API plus Home/Workspace/TargetJob progress-refresh contract；Parse、quick-start、session start 与 next-round plan creation 均在其范围外。本轮未运行，状态保持 `Ready`。

### Phase 26: Workspace list/detail route split

- Supersede v1.19 / Phase 14 pure-list/card-to-Parse: query-free `/workspace` loads the ready list; `/workspace?targetJobId` loads one read-only detail through generated `getTargetJob`. The card body carries only targetJobId and directly opens detail.
- Workspace detail reuses the unified read-only detail component and API-owned resume/round/progress facts. It must not call `importTargetJob`, start a Parse poll scheduler, render Parse progress animation, or run route-side auto-start. Missing/invalid/mismatched detail fails closed without leaking another TargetJob.
- Under React StrictMode, shell/001 safe-read single-flight and stable loader dependencies must yield exactly one same-key initial underlying `listTargetJobs` for list and one `getTargetJob` for detail. Request-count evidence comes from bottom transport spies; hook/effect call count is insufficient.
- Ready import transition and Back behavior remain shell/004 Phase 12 responsibility: Parse ready uses replace to workspace detail; Back cannot replay the animation. Workspace only consumes the canonical route state.

### Phase 27: Workspace detail round-state affordance

- Prototype-first: `frontend/src` consumes the already-validated `eiResolvePracticeProgress` result and renders round-assumption cards as `done/current/pending`; Workspace target detail starts in preview state and never runs the Parse animation.
- Formal frontend uses only `resolveTargetJobPracticeProgress(targetJob)`. Indexes before `completedCount` are done, the exact valid `currentIndex` is current, and later indexes are pending. Each valid card exposes a localized visible label plus `data-round-state`; invalid/missing projections keep neutral cards with no fabricated state and disable Start.
- Reuse existing palette tokens: done uses `okSoft/ok`, current uses `accentSoft/accent`, pending uses `bgSoft/rule-strong`. Do not add API/schema/store fields, theme tokens, lifecycle-status fallbacks, URL state or browser persistence.

### Phase 28: Workspace detail leading resume link and action row

- EXECUTION OWNER：共享 ready-detail 组件位于 `frontend-home-job-picks-and-parse/001`，其 Phase 23 是唯一 UI RED/GREEN 实施 owner；本 Phase 28 只承接 Workspace route、saved TargetJob 事实源、Start/Report handoff 与跨 owner 验收，不建立第二套组件、测试树或重复实现。
- RED：扩展 Workspace detail / shared `ParseScreen` component tests 与 UI source/responsive contract，先证明当前标题右侧 report、独立 `parse-launch`/`parse-resume-binding` block、页尾 Start 与缺失 resume 行为不符合新信息层级；保留 `getTargetJob` 单次实际 transport、Start/Report route 及轮次三态回归断言。
- GREEN：删除独立 Interview Launch/绑定简历大卡片和页尾 action 区；标题 cluster 在“面试规划详情”旁渲染“绑定简历”链接，点击 `navigate({ name: "resume_versions", params: { resumeId: targetJob.resumeId } })`。缺失/空绑定渲染非链接状态并禁用 Start；不得调用 `getResume`、`listResumes` 或从 URL/list item/最近简历推断。同步删除仅供旧 block 使用的 `parse.launch*`、`parse.resumeBound*`、`parse.footerHint` locale key/test 断言，保留新链接与缺失态所需文案。
- ACTIONS：标题下方首行动作行左对齐“立即面试” primary 与“面试报告” secondary；desktop 同排，mobile 保持 DOM/阅读顺序并在不足时换行。Report 只携带可信 `targetJobId`；Start 继续使用 saved resume + strict current progress。启动错误紧邻 action row，不阻断报告入口。
- CLOSEOUT：focused Vitest 只作开发反馈；执行根 `make test`、frontend typecheck/build、desktop/mobile DOM/style/bbox/no-overflow、owner contexts、`sync-doc-index --check`、`make docs-check`、`git diff --check` 与旧标题右侧/独立 binding/footer action/孤儿 locale key 负向搜索，完成后恢复 `completed`。

### Phase 29: Interview-plan card metadata pruning

- RED-GREEN：扩展 `MockInterviewCard.test.tsx`，对 lifecycle `status` 使用非默认值锁定卡片无对应状态文案/徽标；分别验证非空 `locationText` 仍显示、空值与空白值不产生 `Location not set` 或空地点节点。随后从共享卡片删除 `statusLabel`、`statusTone`、状态 badge 与 fallback 文案，不改 generated `TargetJob` 类型、后端状态机或接口字段。
- BDD：`BDD.WORKSPACE.CARD.003` 使用共享卡片组件测试作为 domain behavior test，覆盖 Home 最近面试与 Workspace 规划列表的同源可见行为；不新增 E2E。
- CLOSEOUT：执行 focused card/workspace/home tests、仓库根 `make test`、frontend typecheck、owner context、文档/index 与 diff gate；确认共享卡片源中无 lifecycle 状态展示和 `Location not set` 残留后恢复 `completed`。

## 5 验收标准

| ID | 验收点 | 验证 |
|----|--------|------|
| A-2 | InterviewContext route hydration and carry/clear behavior are deterministic | `InterviewContext.test.tsx`, `App.test.tsx` |
| A-3 | Workspace has no Plan Switcher / Resume Picker; workspace detail displays API-saved binding readonly | source negative + detail tests |
| A-4 | Start practice runs from parse/report owner through `getPracticePlan` / `createPracticePlan` / `startPracticeSession` with idempotency | `ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx` |
| A-5 | Workspace owner no longer contains embedded insight / records static affordance runtime | source negative gate |
| A-6 | Privacy and out-of-scope route/module gates have zero runtime residuals | scenario verify scripts, pruning-surface lint |
| A-8 | Workspace plan-list cards keep stable desktop width regardless of 1/2/3 card count | `WorkspaceScreen.test.tsx`, browser screenshot |
| A-10 | Target job import persists selected resume binding; workspace detail recovers it from `getTargetJob`, never query/list-item authority | backend targetjob + detail tests |
| A-13 | Parse/report handoff owners start practice directly and do not route through `workspace(autoStartPractice=1)` | `ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx` |
| A-14 | Workspace card click opens planning detail while footer provides quick start carrying structured `roundId/roundName`, and top-right delete performs persistent `archiveTargetJob`; Home recent reuses quick start and omits delete | `MockInterviewCard.test.tsx`, `HomeRecentMocks.test.tsx`, `WorkspaceScreen.test.tsx`, `WorkspaceEmptyState.test.tsx`, browser screenshots |
| A-18 | Home/Workspace/Parse/Report consume backend-persisted progress, reuse only exact current round plans, and fail closed after final/invalid progress without browser business-state persistence | repository-root `make test`; focused mapper/start/Parse/Report tests only for development feedback; storage negative gate and UI parity |
| A-20 | Real login plus completion survives Home/Workspace/TargetJob refresh and detail read | `E2E.P0.098`; explicitly excludes Parse, chat, plan creation and session start |
| A-21 | Workspace detail starts with title-adjacent bound-resume link and a left-aligned Start/Reports action row; no standalone binding/launch block or footer Start remains | `ParseScreen.test.tsx`, `ParseResumeBinding.test.tsx`, `App.test.tsx`, responsive/a11y owner gates; root `make test` |
| A-22 | Shared Home/Workspace interview-plan cards omit TargetJob lifecycle status and empty-location placeholders while preserving real locations and the persisted-progress rail | `MockInterviewCard.test.tsx`, `HomeRecentMocks.test.tsx`, `WorkspaceScreen.test.tsx`; root `make test` |
| A-23 | Home recent、Workspace list/detail 与 Report replay/next-round 在 session opening LLM 等待期间共享诚实、可访问、阻断交互且 reduced-motion 兼容的全屏 transition；失败回到原入口错误 | shared transition contract + `HomeRecentMocks.test.tsx`, `WorkspaceScreen.test.tsx`, `ParseResumeBinding.test.tsx`, `ReplayCta.test.tsx`; root `make test` |

### Phase 30: Shared practice-launch transition

#### 30.1 RED：锁定四类正式入口的 pending 行为

在 Home recent、Workspace list、Workspace detail/Parse owner 与 Report replay/next-round caller tests 中延迟 `startPracticeSession`，先断言当前实现不能立即提供统一全屏 `role=status` / `aria-busy` 反馈；同时锁定重复点击、auth redirect 与 API/route/idempotency 不变。

#### 30.2 GREEN：实现共享诚实过渡态

在 `frontend/src/app/interview-context` 增加单一共享 `PracticeLaunchTransition`，由现有 caller 的 in-flight state 驱动。过渡层覆盖 viewport、阻断背景交互，使用本地化标题/说明和 indeterminate 装饰，不展示百分比、伪阶段或 opening message；CSS 在 `prefers-reduced-motion` 下停用非必要循环动画。成功导航 `practice`；失败卸载过渡层并保留 caller-owned error。

#### 30.3 REFACTOR / POST-PASS

保持 `startPracticeFromParams`、generated client、fixtures、OpenAPI、backend、persistence、route params 与 idempotency 不变；执行 focused caller/component/a11y tests、frontend typecheck/build、仓库根 `make test`、owner context/docs/index/diff gates，并完成必要的真实浏览器 desktop/mobile pending-state检查后恢复 completed lifecycle。

### Phase 31: Practice global App chrome

从 shell no-chrome allowlist 删除 `practice`，保留 `generating`。`PracticeScreen` 上方渲染共享 App TopBar，其下现有公司/岗位/角色/计时/暂停/电话/结束控件明确作为 Practice Session Header；不复制导航、主题或账号状态。component/router tests 锁定全局与会话两层 header、TopBar 导航/设置可用、进入/离开 Practice 的 `/me` 请求增量为 0；desktop/mobile responsive 与 Chrome 截图证明关键动作可达且无 document overflow。

### Phase 32: Interview plan-list reference alignment

#### 32.1 RED：锁定参考稿页面级几何

扩展 `WorkspaceScreen.test.tsx`、`MockInterviewCard.test.tsx` 与 `WorkspaceVisual.test.ts`，先要求全宽背景层 + 约 1508px desktop 内容层、header/grid 共享 1456px 右边界、两列等宽宽卡、公司/岗位/轮次 rail、上次保存 footer、52px 级删除触控区和参考级开始 CTA；旧 inline layout、360px 窄卡、CTA 右侧错位以及会裁剪全宽背景的单层容器合同必须先失败。

#### 32.2 GREEN：重构 Workspace list 与 card presentation

把 Workspace list 拆成全宽 `.ei-workspace-plan-list` 背景层和居中 `.ei-workspace-plan-inner` 内容层，并将 header/loading/error/empty/grid 从 inline style 收敛为 `.ei-workspace-*` 页面作用域；`MockInterviewCard` 的 Workspace presentation 改为参考稿的公司 icon、岗位标题、动态轮次 rail、分隔 footer 和本地化更新时间。保持卡片主体打开详情、右上角 `archiveTargetJob`、footer `startPracticeFromParams`、progress fail-closed 与 Home record presentation 不变。

#### 32.3 BDD / A11Y / RESPONSIVE

`BDD.WORKSPACE.LIST.VISUAL.006` 使用 Workspace/Card owner tests 验证可见层级与交互行为；desktop 1916×821 和 mobile 390×844 通过真实 Chrome bbox、screenshot、keyboard、theme、console 与 no-overflow 验收，不创建伪 E2E。

#### 32.4 POST-PASS

执行 focused Vitest、frontend typecheck/build、仓库根 `make test`、owner context、`sync-doc-index --check`、`make docs-check` 与 `git diff --check`；证据同步后恢复 completed lifecycle。

## 6 变更记录

| 日期 | 版本 | 变更 |
|------|------|------|
| 2026-07-19 | 1.49 | Keep the Workspace canvas full-viewport and align the header CTA with the card-grid right edge. |
| 2026-07-19 | 1.48 | Reopen Phase 32 to align the Workspace plan-list body and wide-card geometry with the supplied desktop reference while preserving all runtime contracts. |
| 2026-07-19 | 1.47 | Reopen Phase 31 so Practice retains the global App TopBar above its session header with zero route-level account refetch. |
| 2026-07-18 | 1.46 | Reopen Phase 30 to add one shared accessible launch transition across every formal practice-session entry while the opening LLM request is pending. |
| 2026-07-17 | 1.45 | Reopen Phase 29 to remove duplicated lifecycle status and empty-location placeholders from the shared interview-plan card. |
| 2026-07-15 | 1.44 | Reopen Phase 28 to replace the standalone resume-binding launch block with a title-adjacent resume link and a leading Start/Reports action row. |
| 2026-07-14 | 1.43 | Separate code-owned Workspace behavior BDD from the Ready-only P0.098 real progress-refresh handoff. |
| 2026-07-14 | 1.42 | Add Phase 27 rail-consistent done/current/pending treatments to Workspace detail round assumptions. |
| 2026-07-14 | 1.41 | Add Phase 26 `/workspace` list plus targetJobId detail, direct card routing and exact detail GET count. |
| 2026-07-12 | 1.40 | Accept non-contiguous canonical sequences and require actual live-browser reload/quick-start evidence before closing the persisted-progress BDD gate. |
| 2026-07-12 | 1.39 | Reopen Phase 25 for backend-persisted round progress, exact plan reuse and browser business-state negative gates. |
| 2026-07-12 | 1.38 | Reopen the active owner for structured-round time budgets and fail-closed next-round progression. |
| 2026-07-10 | 1.36 | Remove the unreachable static Workspace detail/insight sources, localize the Parse binding pill, and reconcile the active spec to a pure list. |
| 2026-07-10 | 1.35 | Remove the unconsumed useStartPracticeContext export. |
| 2026-07-10 | 1.34 | Remove five test-only InterviewContext reducer actions and keep the runtime action surface explicit. |
| 2026-07-10 | 1.33 | Remove obsolete auto-start context state and replace implicit practice route-param carry with an explicit allowlist. |
| 2026-07-10 | 1.30 | Align workspace route/list negative wording to out-of-scope/stale terminology without behavior changes. |
| 2026-07-09 | 1.25 | Review remediation: preserve structured round params in workspace list quick-start practice handoff. |
| 2026-07-09 | 1.24 | Move the workspace card delete icon to the card top-right; keep footer for `立即面试` only. |
| 2026-07-09 | 1.23 | Reopen owner plan to integrate generated `archiveTargetJob` for persistent workspace card delete. |
| 2026-07-09 | 1.22 | Reopen owner plan to replace visible Open plan footer CTA with card-click planning plus quick-start and delete actions. |
| 2026-07-09 | 1.21 | Reopen owner plan to fuse Home recent mock cards and workspace plan-list cards into one shared card body with a workspace footer CTA. |
| 2026-07-09 | 1.16 | Reopen owner plan for target job-level resume binding persistence so workspace plan-list re-entry no longer loses the resume selected during JD import. |
| 2026-07-09 | 1.17 | Reopen owner plan to route workspace current-plan detail into the unified Parse-derived Interview Plan Detail / Context Confirm mother page while preserving workspace start-practice ownership. |
| 2026-07-10 | 1.29 | Rename workspace out-of-scope-param route purity wording while preserving canonical `/workspace` behavior. |
| 2026-07-10 | 1.28 | Rename records-area wording to records static affordance while preserving the current workspace empty records behavior and negative gates. |
| 2026-07-09 | 1.19 | Reopen owner plan to make workspace a pure list route, delete out-of-scope detail/start/modal runtime, route cards to parse, and move practice start side effects to parse/report owners. |
| 2026-07-09 | 1.18 | Reopen owner plan for parse-failure dirty-data admission defense and stale InterviewContext-free TopBar workspace navigation. |
| 2026-07-08 | 1.14 | Reopen owner plan for plan-list card simplification, metadata removal and theme-consistent CTA styling after screenshot review. |
| 2026-07-08 | 1.13 | Reopen owner plan for interview plan-list card visual hardening after screenshot review. |
| 2026-07-08 | 1.12 | Reopen owner plan for Interview nav naming and workspace plan-list landing revision. |
| 2026-07-07 | 1.11 | Compress owner docs to the current workspace, flat Resume Picker, start-practice, embedded insight, records static area and privacy contract. |
| 2026-07-07 | 1.10 | Reconcile workspace owner handoff, completion, fixture and active-list boundary wording. |
