# OpenAPI v1 Contract Resume Workshop Additive Coverage Checklist

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-07-06

**关联计划**: [plan](./plan.md)

## Cross-plan prerequisite signals

- [x] B3 D-14 `ResumeTailorMode` 漂移修复已由 [event-and-outbox-contract/002](../../../event-and-outbox-contract/plans/002-resume-tailor-mode-drift-fix/plan.md) 落地；`shared/events.yaml` / baseline manifest / Go/TS generated events 均对齐 `[gap_review, bullet_suggestions]`。
- [x] B4 D-17 `resume_versions` / `resume_version_suggestions` 已由 [db-migrations-baseline/002](../../../db-migrations-baseline/plans/002-resume-versions-additive/plan.md) 落地；migration chain、enum-sources 与 privacy deletion matrix 已承接 Resume Workshop additive persistence。

## Phase 1: B1 D-10 vocabulary 真理源同步

- [x] 1.1 在 `shared/conventions.yaml` 添加 `ResumeVersionType` / `ResumeSeedStrategy` / `ResumeTailorSuggestionStatus` 3 个枚举与 `RESUME_EXPORT_NOT_AVAILABLE` 错误码常量（验证：YAML lint + B1 generator idempotent）
- [x] 1.2 运行 `make codegen-conventions && make codegen-check`，验证 Go side `errors.AllCodes` 与 TS side `as const` 字面量含 `RESUME_EXPORT_NOT_AVAILABLE`（验证：`git diff --exit-code`）
- [x] 1.3 运行 B1 lint（`UPPER_SNAKE_CASE` 错误码 / `lower_snake_case` 枚举值），断言新增 enum 与错误码命名通过（验证：B1 本地 lint gate）
- [x] 1.4 修订 `docs/spec/shared-conventions-codified/spec.md` D-10 中"声明阶段"措辞改为"落地"，shared-conventions-codified 枚举类型 14 → 17（验证：手工检查 + `sync-doc-index --check`）
- [x] 1.5 B1 spec.md 1.16 → 1.17，history.md 追加 1.17 行（关联本 plan）（验证：`sync-doc-index --check`）

## Phase 2: OpenAPI schema + operation 落地

- [x] 2.1 在 `openapi/openapi.yaml` `components.schemas` 新增 7 个 schema：`ResumeVersion` / `BranchResumeVersionAccepted` / `PaginatedResumeAsset` / `PaginatedResumeVersion` / `BranchResumeVersionRequest` / `UpdateResumeVersionRequest` / `ResumeTailorSuggestionStatus`（验证：`swagger-cli validate` + JSON Schema 校验）
- [x] 2.2 扩展 `RegisterResumeRequest` additive 字段（`sourceType` / `rawText` / `guidedAnswers` JSON object），保持现有 `fileObjectId` / `title` / `language` 向后兼容，并与 B4 `resume_assets.guided_answers` jsonb 持久化字段对齐（验证：`make openapi-diff` additive only）
- [x] 2.3 添加 9 个 op 到 `Resumes` tag（listResumes / listResumeVersions / getResumeVersion / branchResumeVersion / updateResumeVersion / acceptResumeTailorSuggestion / rejectResumeTailorSuggestion / archiveResumeAsset / exportResumeVersion），每个 op 含 request/response/error schema `$ref`；不得把 accept/reject suggestion 落到 `ResumeTailor` tag（验证：inventory lint negative case 通过）
- [x] 2.4 6 个 side-effect operation 必含 `Idempotency-Key` header schema（验证：inventory lint `IK_REQUIRED` 覆盖）
- [x] 2.5 `ResumeVersion` schema 含 `provenance` 字段 `$ref` `GenerationProvenance`（验证：`make validate-fixtures` AI provenance gate）
- [x] 2.6 `exportResumeVersion` P0 响应固定为 `501` + `ApiErrorResponse.error.code = "RESUME_EXPORT_NOT_AVAILABLE"`；本 plan 不落 P1 `202 + Job`，若未来执行 501 → 202 再按 D-12 privacy export 模式补 diff-config 白名单（验证：`make openapi-diff` additive only）
- [x] 2.7 同步 `scripts/lint/openapi_inventory.py` 501 allowlist：允许 `requestPrivacyExport` 与 `exportResumeVersion`，并保持其他 endpoint 501 负向拦截（验证：新增/更新 inventory lint negative case）

## Phase 3: Fixtures 与 mock-contract-suite 同步

- [x] 3.1 创建 `openapi/fixtures/Resumes/listResumes.json`（含 `default` / `empty` / `paginated` 3 variant）（验证：`make validate-fixtures` schema-valid）
- [x] 3.2 创建 `openapi/fixtures/Resumes/listResumeVersions.json`（含 `default` / `master-only` / `with-targeted-branches`）（验证：`make validate-fixtures`）
- [x] 3.3 创建 `openapi/fixtures/Resumes/getResumeVersion.json`（含 `master-default` / `targeted-with-suggestions` / `not-found-404`）（验证：同上）
- [x] 3.4 创建 `openapi/fixtures/Resumes/branchResumeVersion.json`（含 `copy-master-sync` / `blank-sync` / `ai-select-202-with-job` / `validation-error-422` / `idempotent-replay`）（验证：同上 + IK 字段断言）
- [x] 3.5 创建 `openapi/fixtures/Resumes/updateResumeVersion.json`（含 `default` / `validation-error-422`）（验证：同上）
- [x] 3.6 创建 `openapi/fixtures/Resumes/acceptResumeTailorSuggestion.json` 与 `rejectResumeTailorSuggestion.json`（含 `default` / `conflict-409`）（验证：同上 + `ResumeTailor` fixture 目录不新增这两个 operation）
- [x] 3.7 创建 `openapi/fixtures/Resumes/archiveResumeAsset.json`（含 `default-202` / `already-archived-409`）（验证：同上）
- [x] 3.8 创建 `openapi/fixtures/Resumes/exportResumeVersion.json`（仅 `p0-501-not-available` variant，含 `error.code = RESUME_EXPORT_NOT_AVAILABLE`）（验证：同上）
- [x] 3.9 不强制生成 Resume Workshop `prototype-baseline`：先验证 `openapi/fixtures/PROTOTYPE_MAPPING.md` 尚未声明 `ui-design/src/screen-resume-workshop.jsx` 数据源；如本 plan 决定新增映射，必须同时修订同步工具并通过 `make sync-fixtures-from-prototype` 幂等（验证：mapping gap 0 或明确 N/A 记录；`rg screen-resume-workshop|listResumes|Resume Workshop|resume openapi/fixtures/PROTOTYPE_MAPPING.md scripts/codegen/sync_fixtures_from_prototype.py` 0 命中）
- [x] 3.10 原地修订 `docs/spec/mock-contract-suite/spec.md` 1.5 → 1.6，inventory 章节升级到 55 op + 新 variant 计数；同步 `openapi/fixtures/README.md`（验证：`sync-doc-index --check`）

## Phase 4: Inventory lint + spec §3.1.1 回填

- [x] 4.1 `scripts/lint/openapi_inventory.py` 修改：`EXPECTED_OPERATIONS = 55`、`IK_REQUIRED` 追加 6 项、`AI_PROVENANCE_SCHEMAS` 追加 `ResumeVersion`（验证：`python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml` PASS）
- [x] 4.2 `scripts/lint/validate_fixtures.py` 注释从 46 → 55 同步（验证：`make validate-fixtures` PASS）
- [x] 4.3 `openapi/README.md` validator 描述同步（验证：手工检查 + `sync-doc-index --check`）
- [x] 4.4 回填 B2 spec §3.1.1 endpoint 列表 #47-55 完整行（参照 #35-46 JobMatch 模式：`# | Tag | Method | Path | OperationId | 关联 schema`）（验证：`sync-doc-index --check`）
- [x] 4.5 B2 spec §3.1.1 末尾 "总计 46 个 endpoint" → "总计 55 个 endpoint"；§2.1 "46 端点" → "55 端点"；3.1.1 "Resume Workshop additive (47–55)" 备注从"声明阶段"改为"已纳入 freeze"（验证：手工检查）
- [x] 4.6 同步 B2 spec §4.1 P0 exception matrix（privacy export + resume version export）、§4.6 AI provenance schema 列表（追加 `ResumeVersion`）与 §6 C-12 `resume export 501` 验收行（验证：grep `RESUME_EXPORT_NOT_AVAILABLE` / `exportResumeVersion` / `ResumeVersion`）
- [x] 4.7 B2 spec.md 1.16 → 1.17，history.md 追加 1.17 行（"D-18 落地阶段：§3.1.1 回填 + §2.1 总数升 55 + resume export 501 例外"，关联本 plan）（验证：`sync-doc-index --check`）
- [x] 4.8 同步 `docs/spec/engineering-roadmap/spec.md` 中 "46 endpoint / 46 operation" 文字描述升级到 55；spec.md 3.12 → 3.13，history.md 追加 3.13 行（验证：grep negative search 旧数字）

## Phase 5: 验收 + 解锁下游

- [x] 5.1 运行 `npx -p @apidevtools/swagger-cli@4.0.4 swagger-cli validate openapi/openapi.yaml` PASS（验证：exit 0）
- [x] 5.2 运行 `make codegen-openapi && make codegen-check` PASS，无 generated drift（验证：`git diff --exit-code`）
- [x] 5.3 运行 `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml` PASS（断言 55 op / 13 tag / 6 新 IK / AI_PROVENANCE_SCHEMAS 含 `ResumeVersion` / 501 allowlist 含 `exportResumeVersion`）
- [x] 5.4 运行 `make validate-fixtures` PASS（含 GenerationProvenance 完整性 + provenance 字段在 ai-select fixture 中存在）
- [x] 5.5 运行 `make openapi-diff` PASS（additive only；确认本 plan 只新增 P0 `exportResumeVersion` 501 stub，未提前落 202）
- [x] 5.6 运行 `sync-doc-index --check` PASS（B1 / B2 / mock-contract-suite / engineering-roadmap spec/history/INDEX 同步）
- [x] 5.7 修订 `docs/spec/INDEX.md` openapi-v1-contract / shared-conventions-codified 版本与日期（验证：`sync-doc-index --check`）
- [x] 5.8 通知 `frontend-workspace-and-practice/001-workspace-and-interview-context` owner：`listResumes` operation + fixtures 已就位，可启动 disabled-list → active-list 原地修订（验证：在 workspace 001 plan 中追加引用本 plan 的 unblock 链接）

## Phase 6: L2 remediation - generated client response shape closure

- [x] 6.1 `branchResumeVersion` 202 response 使用命名 schema + generated TS union return type（验证：codegen unit test + `make codegen-openapi`）
- [x] 6.2 generated TS client 对显式 P0 `501` response 走 typed parse path，未声明 4xx/5xx 仍 throw（验证：codegen unit test + frontend focused Vitest）
- [x] 6.3 dev mock client 导入 9 个 Resume Workshop fixture，operationId coverage 与 `exportResumeVersion` typed fallback 通过（验证：`pnpm --filter @easyinterview/frontend test src/api/devMockClient.test.ts`）

## Phase 7: D-20 简历扁平化 contract collapse

> product-scope D-20 / B2 D-26。Red 优先：先改 inventory lint 期望值再改 `openapi.yaml`。与 `db-migrations-baseline/002` Phase 6 + `backend-resume` D-20 同属 contract impl 原子提交。

- [ ] 7.0 Red：`scripts/lint/openapi_inventory.py` `EXPECTED_OPERATIONS` 48→43 + operation set 调整，确认 `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml` 先红（验证：lint 退出码非 0）
- [ ] 7.1 删除 6 个 op（`confirmResumeStructuredMaster` / `listResumeVersions` / `getResumeVersion` / `branchResumeVersion` / `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion`）+ 6 fixtures（验证：`swagger-cli validate` PASS + `grep -c operationId` = 43）
- [ ] 7.2 重命名 `updateResume` / `archiveResume` / `exportResume` + `resumeId` 路径 + `RequestResumeTailorRequest.resumeVersionId`→`resumeId` + 移动 fixtures（验证：`swagger-cli validate` PASS + `rg "resumeAssetId|resumeVersionId|resume-versions" openapi/openapi.yaml` 0 命中）
- [ ] 7.3 新增 `duplicateResume` op + `DuplicateResumeRequest` schema + fixture（验证：fixture schema-valid + `make validate-fixtures` PASS）
- [ ] 7.4 schema 扁平化：`ResumeAsset`→`Resume`(+`structuredProfile`/`displayName`) / `PaginatedResume` / `UpdateResumeRequest`，删 `ResumeVersion` 等 6 schema，`RegisterResumeRequest.sourceType` {upload,paste}，删 3 enum `$ref`（验证：`swagger-cli validate` PASS + 无悬空 `$ref`）
- [ ] 7.5 inventory + `validate_fixtures` lint 真理源 48→43、`AI_PROVENANCE_SCHEMAS` `ResumeVersion`→`Resume`、`IK_REQUIRED` 调整（验证：`openapi_inventory.py` + `make validate-fixtures` PASS）
- [ ] 7.6 `make codegen-openapi` + `make codegen-check` + baseline 原地 re-freeze + `make openapi-diff` PASS（验证：codegen-check exit 0 + openapi-diff PASS）
- [x] 7.7 B2 spec 1.30 + history 1.30 + `openapi/README.md` / `openapi/fixtures/README.md` / roadmap / mock-contract-suite 计数同步（本次 doc 修订已完成 spec/history；README/roadmap 计数随 impl）（验证：`sync-doc-index --check` 零漂移）
- [ ] 7.8 零残留负向 grep `resumeVersionId|resumeAssetId|ResumeVersion|branchResume|listResumeVersions|getResumeVersion|confirmResumeStructuredMaster|accept+rejectResumeTailorSuggestion|ResumeSeedStrategy|ResumeTailorSuggestionStatus`（openapi/lint/generated；除 history 历史行 + 负向断言）（验证：0 命中）
- [ ] 7.9 下游信号：`frontend-resume-workshop` / `frontend-workspace-and-practice` / `backend-resume` 已收到 resume 契约坍缩 + `resumeId` 信号；Debrief 旧 subject 已随 product-scope D-22 删除，不再作为正向下游（验证：cross-plan 引用）
- [x] 7.10 L2 hardening: `validate_fixtures.py` 递归拒绝 fixture request / response 中的 D-20 退役 key `resumeAssetId` / `resumeVersionId`，并修正 Debriefs fixture 为 `resumeId`；验证：`python3 -m unittest scripts.lint.validate_fixtures_cli_test`、`make validate-fixtures`、`rg -n "resumeVersionId|resumeAssetId" openapi/fixtures openapi/openapi.yaml backend/internal/api/generated frontend/src/api/generated` 0 命中
