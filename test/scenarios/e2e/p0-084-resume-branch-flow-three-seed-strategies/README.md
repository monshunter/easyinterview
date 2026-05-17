# E2E.P0.084 Resume Branch Flow Three SeedStrategy + IK + 422 + Cross-User + UI Parity

> **场景 ID**: E2E.P0.084
> **执行方式**: automated (vitest jsdom)
> **隔离级别**: in-process (vitest worker)
> **状态**: Ready

## 1 Given

- Fixture-backed mock-first client：`Resumes/listResumes.json default` + `Resumes/listResumeVersions.json default` + `Resumes/branchResumeVersion.json` 六 scenario (`default / copy-master-sync / blank-sync / ai-select-202-with-job / idempotent-replay / validation-error-422`)。
- 用户：未登录 → 登录态，lang 默认。
- Phase 0 real-backend preflight 已固化：`branchResumeVersion` generated TS client + Go server interface + handler + cmd/api route 真实可用（plan 003 checklist 0.1 evidence）。

## 2 When

- 未登录加载 `resume_versions?flow=branch&branchOriginalId={id}` → 显示 auth gate。
- 登录恢复后渲染 `ResumeBranchFlow`，填写 name + target + focus + seed='copy_master' 提交。
- 同 IK 再次提交（replay）。
- 切到 seed='blank' / seed='ai_select' 重复提交。
- 422 路径：提交缺 displayName 字段触发 `validation-error-422`。
- 切到不存在的 `branchOriginalId` → NotFound CTA。
- 用户 B 用 A 的 parentVersionId + 自己的 IK → 404 cross-user。

## 3 Then

- pendingAction 仅携带 `{ flow: 'branch', branchOriginalId }`，不含 form draft (`name/target/focus/seed`) 或 wire 字段 (`parentVersionId/displayName/focusAngle/seedStrategy`)。
- `branchResumeVersion` 请求带 `Idempotency-Key` header（v1 wire format）；同 fingerprint retry 复用 IK；422 重置 IK 缓存。
- 三 seedStrategy 分发 nav：copy_master/version → tab=rewrites；blank/version → tab=edit；ai_select/accepted → tab=rewrites + `tailorRunId` 写入 nav。
- 422 → `resume-branch-error` in-form alert；不 nav；不创建行。
- 404 cross-user / 不存在 branchOriginalId → NotFound 或 generic 404 toast；不暴露原 envelope。
- DOM testid ≥ 12（resume-branch-flow / -back / -from-card / -from-original / -from-master / -field-name / -field-target / -focus-chip-{platform,collaboration,fullstack,leadership,custom} / -seed-card-{copy_master,blank,ai_select} / -cancel / -submit / -submit-hint / -error）。
- 旧入口 grep：`welcome|mistake|growth|drill|followup|STAR|experiences|voice|OnboardingScreen|onboarding=true` 在 `branch/` + `tabs/` 0 命中；retired tailor mode `(inline|rewrite|mirror)` 0 命中；prototype import `ui-design/src/(data|screen-resume-workshop)` 0 命中。
- `method=fixture-backed-frontend` 显式标注，附 backend real route preflight evidence。

## 4 Verification Entry

`scripts/trigger.sh` 通过 Vitest 调用：

- `src/app/screens/resume-workshop/branch/ResumeBranchFlow.test.tsx`
- `src/app/screens/resume-workshop/branch/hooks/useResumeBranchSubmit.test.tsx`
- `src/app/screens/resume-workshop/branch/adapters/mapBranchFormToRequest.test.ts`
- `src/app/screens/resume-workshop/ResumeWorkshopAuthGate.test.tsx`

## 5 Output

- `.test-output/e2e/p0-084-resume-branch-flow-three-seed-strategies/trigger.log` Vitest pass output。
- verify.sh 断言 trigger.log 含 vitest RUN 标记 + `Test Files .* passed` + `Tests .* passed`，并显式 grep 每个 spec 文件被执行。

## 6 Baseline

- `make codegen-check` 已通过的 generated TS client（`branchResumeVersion` 签名 + `Idempotency-Key` header）。
- `openapi/fixtures/Resumes/branchResumeVersion.json` 六 scenario 字节。

## 7 离线限制

本场景纯 fixture-backed Vitest 路径，无需 Docker Compose / Kind 或外网；离线运行 PASS。

## 8 方法标注

`method=fixture-backed-frontend`（plan 003 §3 fixture-backed 路径）。Backend real route preflight evidence：plan 003 checklist 0.1 - 0.4 已记录。
