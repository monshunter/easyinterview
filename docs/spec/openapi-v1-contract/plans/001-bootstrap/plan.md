# 001 - OpenAPI v1 Contract Bootstrap

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

落地当前 B2 OpenAPI v1 contract bootstrap：

- `openapi/openapi.yaml` is the single HTTP contract truth source for current 36 operations / 10 tags.
- Go generated server/types live in `backend/internal/api/generated/`.
- TypeScript generated client/types live in `frontend/src/api/generated/`.
- Root Make targets provide `codegen-openapi`, `codegen-check`, `lint-openapi`, and `docs-openapi`.
- B1 shared conventions are referenced through generated/shared types and error envelope rules.
- Fixtures and breaking-change gates consume this bootstrap output through sibling B2 plans.

This completed owner plan is now an executable evidence index. It does not preserve staged remediation prose.

## 2 Current Contract

| Surface | Current contract | Gate |
|---------|------------------|------|
| OpenAPI inventory | 10 tags, 36 operations, `/api/v1` prefix, session-cookie auth, public/protected operation security | `make lint-openapi`, inventory tests |
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
| Resumes | `listResumes`, `registerResume`, `getResume`, `updateResume`, `duplicateResume`, `archiveResume`, `exportResume` |
| TargetJobs | `importTargetJob`, `listTargetJobs`, `getTargetJob`, `updateTargetJob` |
| PracticePlans | `createPracticePlan`, `getPracticePlan` |
| PracticeSessions | `listPracticeSessions`, `startPracticeSession`, `getPracticeSession`, `appendSessionEvent`, `completePracticeSession`, `createPracticeVoiceTurn` |
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

- `make lint-openapi`: validates `openapi/openapi.yaml` and the current 10-tag / 36-operation inventory.
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
| 2026-07-07 | 1.7 | Compress owner plan to current 36-operation / 10-tag OpenAPI contract and executable evidence index. |
| 2026-05-04 | 1.6 | Complete OpenAPI v1 bootstrap delivery. |
