# 002 BDD Plan

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 关联 Phase | 关联 Spec C-* | 关联 BDD-Gate（主 checklist） |
|---------|------|------------|--------------|-------------------------------|
| E2E.P0.074 | primary + boundary + auth · flat `getResume` / `listResumes` + removed route/catalog 404 | Phase 1 | C-6, C-13 | Phase 1 |
| E2E.P0.075 | primary + failure · `updateResume` IK + editable field overwrite + server-owned field 422 + cross-user 404 | Phase 2 | C-17 | Phase 2 |
| E2E.P0.076 | primary + boundary · `duplicateResume` save-as-new + IK + source isolation + rollback | Phase 3 | C-18 | Phase 3 |
| E2E.P0.077 | primary · `requestResumeTailor` + `getResumeTailorRun` queued/ready + `ai_task_runs` + ready-only outbox | Phase 4 + 5 | C-16 | Phase 4 + 5 |
| E2E.P0.078 | failure/recovery · resume.tailor timeout retryable + output_invalid terminal + retry-to-ready + ready-only outbox | Phase 4 + 5 | C-16 failure path | Phase 4 + 5 |
| E2E.P0.079 | regression + boundary · removed suggestion route inputs + flat save fixture parity + read-only frontend detail | Phase 1 + 2 + 3 + 6 | C-17, C-18 | Phase 6 |
| E2E.P0.080 | regression + privacy · outbox / audit / `ai_task_runs` privacy + runtime vocabulary negative | Phase 5 + 6 | C-13, C-16 | Phase 5 |

## 2 场景明细

### E2E.P0.074 flat resume reads and removed route boundary

| Given | When | Then | 验证入口 |
|-------|------|------|----------|
| 用户 A 拥有 flat ready resumes，用户 B 无访问权；B2 `getResume` / `listResumes` fixtures 使用 `resumeId`；generated route catalog 与当前 flat API 对齐 | 运行 `make validate-fixtures`、removed route/catalog tests、flat read handler/service/store focused tests | `getResume` / `listResumes` fixture parity PASS；cross-user 404；cursor order 稳定；removed route inputs 返回 404；generated route catalog/session policy 不含 removed operation；场景证据不泄漏 raw resume 或 suggestion 文本 | `test/scenarios/e2e/p0-074-resume-flat-read-api/` |

### E2E.P0.075 flat resume update and IK

| Given | When | Then | 验证入口 |
|-------|------|------|----------|
| 用户 A 拥有 flat resume，用户 B 无访问权；B2 `updateResume` fixture 有 `default` / `idempotency-replay` / `validation-error-422` | 运行 fixture validation、`cmd/api` route gate、handler fixture parity、service/store update focused tests | `PATCH /api/v1/resumes/{resumeId}` 覆盖 `structuredProfile` / `displayName`；IK replay/mismatch 语义保持；server-owned fields 422；cross-user / missing row 404；证据无 raw profile leak | `test/scenarios/e2e/p0-075-resume-update-flat-fields-and-ik/` |

### E2E.P0.076 flat resume duplicate save-as-new

| Given | When | Then | 验证入口 |
|-------|------|------|----------|
| 用户 A 拥有 flat resume；B2 `duplicateResume` fixture 有 `default` / `idempotency-replay` / `validation-error-422` | 运行 fixture validation、`cmd/api` route gate、handler fixture parity、service/store duplicate focused tests | `POST /api/v1/resumes/{resumeId}/duplicate` 创建新 resume，复制只读来源快照并应用可编辑 overlay；IK replay 不重复创建；invalid input 422；cross-user source 404；rollback 无 orphan | `test/scenarios/e2e/p0-076-resume-duplicate-save-as-new/` |

### E2E.P0.077 flat resume tailor async dispatch and ready

| Given | When | Then | 验证入口 |
|-------|------|------|----------|
| 用户 A 拥有 flat resume + target job；A3 AIClient stub 返回 success JSON；F3 `resume.tailor.*` feature_key ready；B2 request/get tailor fixtures 使用 `resumeId` | 运行 `requestResumeTailor` / `getResumeTailorRun` handler fixture parity、service/store tailor focused tests、`cmd/api` drainer ready path、job handler ready path | `requestResumeTailor` 创建 `async_jobs(job_type='resume_tailor')` with `payload.resumeId`；`getResumeTailorRun` 从 async job status/result 返回 queued/generating/ready/failed；success 写 typed `ai_task_runs`；`resume.tailor.completed` ready-only，payload 仅含 `tailorRunId` / `resumeId` / `targetJobId` / `mode` / `status` | `test/scenarios/e2e/p0-077-resume-tailor-async-dispatch-and-ready/` |

### E2E.P0.078 resume.tailor failure and retry

| Given | When | Then | 验证入口 |
|-------|------|------|----------|
| 三个 deterministic tailor async jobs 覆盖 timeout、output_invalid、timeout-then-success；`cmd/api` in-process drainer 可 `RunOnce` | 运行 `TestResumeTailorDrainerFailureScenario`、`TestTailorHandlerModeRoutingAndFailurePaths`、live store ready-only outbox integration | timeout 为 retryable `AI_PROVIDER_TIMEOUT`；invalid output 为 terminal `AI_OUTPUT_INVALID`；retry 可回到 generating 并最终 ready；每次 AI attempt 写 `ai_task_runs`；只有 final ready 发 `resume.tailor.completed` | `test/scenarios/e2e/p0-078-resume-tailor-failure-and-retry/` |

### E2E.P0.079 flat save fixture parity and read-only detail boundary

| Given | When | Then | 验证入口 |
|-------|------|------|----------|
| flat fixtures、generated route catalog、frontend detail read-only surface 已更新 | 运行 fixture validation、removed route/catalog tests、flat save fixture parity、frontend read-only detail negative Vitest | Removed suggestion route inputs 保持 404 且 generated catalog absent；`updateResume` / `duplicateResume` / `requestResumeTailor` fixture parity green；frontend detail 不渲染 Rewrites/Edit/export/copy/original surfaces，不通过详情调用 save/tailor/export 操作 | `test/scenarios/e2e/p0-079-resume-rewrites-accept-only-save/` |

### E2E.P0.080 tailor privacy and runtime vocabulary negative

| Given | When | Then | 验证入口 |
|-------|------|------|----------|
| P0.074-P0.079 已覆盖 flat API / persistence / drainer paths；privacy fixtures inject private markers | 运行 job privacy tests、live store ready-only outbox privacy gate、cmd/api drainer privacy gates、runtime vocabulary negative greps | outbox payload 只含 IDs/mode/status；`ai_task_runs` 和 audit metadata 不持久化 prompt/model/raw resume/match summary/suggested bullet；backend resume runtime 0 命中 `inline|mirror|mistakes|growth|drill|inline-debrief-record` | `test/scenarios/e2e/p0-080-resume-tailor-privacy-negative/` |
