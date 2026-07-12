# 001 - OpenAPI v1 Contract Bootstrap

> **版本**: 1.12
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

落地当前 B2 OpenAPI v1 contract bootstrap：

- `openapi/openapi.yaml` is the single HTTP contract truth source for current 37 operations / 10 tags.
- Go generated server/types live in `backend/internal/api/generated/`.
- TypeScript generated client/types live in `frontend/src/api/generated/`.
- Root Make targets provide `codegen-openapi`, `codegen-check`, `lint-openapi`, and `docs-openapi`.
- B1 shared conventions are referenced through generated/shared types and error envelope rules.
- Fixtures and breaking-change gates consume this bootstrap output through sibling B2 plans.

This completed owner plan is now an executable evidence index. It does not preserve staged remediation prose.

## 2 Current Contract

| Surface | Current contract | Gate |
|---------|------------------|------|
| OpenAPI inventory | 10 tags, 37 operations, `/api/v1` prefix, session-cookie auth, public/protected operation security | `make lint-openapi`, inventory tests |
| Error envelope | B1 `ApiError` inner object + B2 `ApiErrorResponse` wire envelope | generator tests, codegen-check |
| Shared types | B1 enum/page/error conventions are reused; OpenAPI does not duplicate shared enum ownership | conventions drift and generated tests |
| Codegen | Go server/types and TS client/types are reproducible from `openapi/openapi.yaml` | `make codegen-openapi`, `make codegen-check` |
| Local docs | Redocly CLI renders `openapi/dist/index.html` without committing generated docs | `make docs-openapi` |
| Downstream handoff | 002 owns fixtures/mock source; 003 owns breaking-change baseline/gate; 004 owns resume additive coverage | plans INDEX and context validation |

## 3 Current Operation Inventory

| Tag | Operations |
|-----|------------|
| Auth | `getMe`, `completeMyProfile`, `deleteMe`, `startAuthEmailChallenge`, `verifyAuthEmailChallenge`, `logout`, `getRuntimeConfig` |
| Uploads | `createUploadPresign` |
| Resumes | `listResumes`, `registerResume`, `getResume`, `getResumeSource`, `updateResume`, `duplicateResume`, `archiveResume`, `exportResume` |
| TargetJobs | `importTargetJob`, `listTargetJobs`, `getTargetJob`, `updateTargetJob`, `archiveTargetJob` |
| PracticePlans | `createPracticePlan`, `getPracticePlan` |
| PracticeSessions | `listPracticeSessions`, `startPracticeSession`, `getPracticeSession`, `sendPracticeMessage`, `completePracticeSession`, `createPracticeVoiceTurn` |
| Reports | `getFeedbackReport`, `listTargetJobReports` |
| ResumeTailor | `requestResumeTailor`, `getResumeTailorRun` |
| Jobs | `getJob` |
| Privacy | `requestPrivacyExport`, `requestPrivacyDelete`, `getPrivacyRequest` |

## 4 Completed Implementation Scope

- OpenAPI 3.1 document with fixed server prefix, tags, security schemes, shared components, idempotency headers, request/response schemas, and default error envelope.
- OpenAPI inventory lint for operation/tag count, operation IDs, idempotency rules, privacy export exception, and schema provenance requirements.
- Go and TS codegen pipeline with reproducible generated artifacts.
- Local API docs renderer using the current Redocly CLI target.
- Codegen and inventory validation integrated into root Make targets.
- Handoff to fixture/mock source and breaking-change gate owner plans.

## 2.3 Make 入口

Current root Make targets owned by this plan:

- `make lint-openapi`: validates `openapi/openapi.yaml` and the current 10-tag / 37-operation inventory.
- `make codegen-openapi`: regenerates Go and TypeScript OpenAPI artifacts.
- `make codegen-check`: verifies generated OpenAPI artifacts are reproducible and not drifted.
- `make docs-openapi`: renders the local OpenAPI HTML document.

## 5 Verification Commands

```bash
make lint-openapi
make codegen-openapi
make codegen-check
cd backend && go test ./cmd/codegen/openapi -count=1
make docs-openapi
python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/openapi-v1-contract/plans/001-bootstrap/context.yaml --target contract
python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check
make docs-check
git diff --check
```

## 6 BDD Applicability

BDD is not applicable. This plan owns internal API contract and generated artifact reproducibility, not a user-facing workflow. User-visible API behavior is covered by the backend/frontend/scenario owners that consume the generated contract.

## 7 Revision Log

| Date | Version | Change |
|------|---------|--------|
| 2026-07-12 | 1.12 | Reopen Phase 11 for additive practice round identity and TargetJob progress projection contract. |
| 2026-07-10 | 1.11 | Remove the unconsumed frontend raw OpenAPI snapshot output and its dedicated generator surface. |
| 2026-07-10 | 1.10 | Remove the unreferenced provenance ref constant from the inventory linter. |
| 2026-07-10 | 1.9 | Move the test-only snapshot hash calculation out of the production codegen package. |
| 2026-07-10 | 1.8 | Align owner inventory with the current 37-operation contract including `getResumeSource` and `archiveTargetJob`. |
| 2026-07-07 | 1.7 | Compress owner plan to the 2026-07-07 36-operation / 10-tag OpenAPI contract and executable evidence index. |
| 2026-05-04 | 1.6 | Complete OpenAPI v1 bootstrap delivery. |

## 8 Test-only snapshot hash cleanup

删除只被 `run_test.go` 使用的 production `sha256.go`。幂等测试在 snapshot traversal 内直接计算 SHA-256，保持 byte-identical generated artifact 断言不变。

## 9 Inventory linter dead constant cleanup

删除 `scripts/lint/openapi_inventory.py` 中无读取方的 `PROVENANCE_REF`；现有 `GenerationProvenance` schema shape 与可达性检查继续由真实 schema-name traversal 承担。

## 10 Frontend raw-spec snapshot removal

TypeScript codegen 只输出正式消费的 `client.ts` 与 `types.ts`。删除没有 import、未进入 Vite bundle、也不被 docs/mock tooling 读取的 raw OpenAPI 字符串快照，同时删除专用 TS template 与只服务该快照的字符串转义 helper；保留 `openapi/openapi.yaml`、backend generated spec 镜像、Redocly 文档和所有 wire/API contract 不变。

## 11 Practice round identity and progress projection

### 11.1 Additive wire contract

- `CreatePracticePlanRequest` 新增可选 `roundId`，只表达客户端选择的结构化轮次；`roundSequence` 必须由服务端从 TargetJob summary 推导，客户端不得提交。
- `PracticePlan` 新增可选 `roundId` / `roundSequence`。新创建记录必须成对返回；字段可选只用于读取 legacy null identity，不授权新路径省略。
- `TargetJob` 新增可选 `practiceProgress: PracticeProgress`；`PracticeProgress` 包含 `status=not_started|in_progress|completed`、有序去重的 `completedRounds: PracticeRoundRef[]` 与 nullable `currentRound`。

### 11.2 Compatibility and generated artifacts

变更只新增 schema / optional property，不新增 endpoint、不修改现有 required 字段或状态码。同步 `openapi/baseline/openapi-v1.0.0.yaml`、Go/TS generated artifacts，并运行 `make lint-openapi`、`make codegen-openapi`、`make codegen-check`、`make openapi-diff`；diff 必须分类为 additive。

### 11.3 Consumer invariant

TargetJob lifecycle `status` 只表示岗位生命周期，不能解释为面试轮次。backend-targetjob 只从完成 session 事实投影 `practiceProgress`；frontend 只消费 `practiceProgress.currentRound` 选择卡片当前轮和 quick-start，legacy null plan 不得按时长碰撞复用。

### 11.4 Operation matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `createPracticePlan` | `PracticePlans/createPracticePlan.json` round variants | shared parse/workspace/report start helper | backend-practice generated adapter/service/store | insert `practice_plans.round_id/round_sequence` + idempotency/audit | none | P0.022/P0.070/P0.072/P0.098 |
| `getPracticePlan` | `PracticePlans/getPracticePlan.json` current + legacy-null | shared start exact-pair reuse; Practice budget | backend-practice generated adapter/store | read nullable legacy/current plan identity | none | P0.022/P0.070/P0.098 |
| `listTargetJobs` | `TargetJobs/listTargetJobs.json` zero/partial/final progress | Home/Workspace cards and quick-start | backend-targetjob list handler/store/service projection | TargetJob summary + plans/sessions/completion events; no mutable progress column | none | P0.018/P0.098 |
| `getTargetJob` | `TargetJobs/getTargetJob.json` zero/partial/final progress | Parse/Report/current-round handoff | backend-targetjob get handler/store/service projection | same ledger projection as list | none after persisted JD parse | P0.057/P0.098 |
