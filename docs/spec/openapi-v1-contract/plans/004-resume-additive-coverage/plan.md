# OpenAPI v1 Contract Resume Workshop Additive Coverage

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-06-14

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [openapi-v1-contract spec](../../spec.md) §3.1 D-18 Resume Workshop additive 升级声明阶段落到仓库可执行 artifact：

- 在 `openapi/openapi.yaml` 中扩容 `Resumes` tag（不新建 `ResumeVersions` tag），新增 9 operationId（`listResumes` / `listResumeVersions` / `getResumeVersion` / `branchResumeVersion` / `updateResumeVersion` / `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` / `archiveResumeAsset` / `exportResumeVersion`）；
- 新增 7 个 schema（`ResumeVersion` / `BranchResumeVersionAccepted` / `PaginatedResumeAsset` / `PaginatedResumeVersion` / `BranchResumeVersionRequest` / `UpdateResumeVersionRequest` / `ResumeTailorSuggestionStatus` enum）+ `RegisterResumeRequest` additive 扩展（`sourceType` ∈ `upload | paste | guided`、`rawText`、`guidedAnswers` JSON object）；
- 通过 `$ref` 引用 [B1 spec](../../../shared-conventions-codified/spec.md) D-10 锁定的 3 个新 enum（`ResumeVersionType` / `ResumeSeedStrategy` / `ResumeTailorSuggestionStatus`）+ 1 个新错误码（`RESUME_EXPORT_NOT_AVAILABLE`），并在 `shared/conventions.yaml` 同步落地；
- `branchResumeVersion` / `updateResumeVersion` / `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` / `archiveResumeAsset` / `exportResumeVersion` 6 个 side-effect operation 必带 `Idempotency-Key`；`ResumeVersion` 列入 `AI_PROVENANCE_SCHEMAS`，`structuredProfile` 与 tailor suggestion 字段必带 `GenerationProvenance`；
- 同步 `scripts/lint/openapi_inventory.py` EXPECTED_OPERATIONS 46 → 55、IK_REQUIRED 追加 6 项、`AI_PROVENANCE_SCHEMAS` 追加 `ResumeVersion`；同步 `scripts/lint/validate_fixtures.py` 注释；
- 每个新 operation 落地 fixtures（含 default / paginated / empty / processing / failed / not-found 多 variant）；
- 回填 B2 spec §3.1.1 endpoint 列表 #47-55 + §2.1 "46 端点" → "55 端点"；
- 通过 spec §6 C-1 / C-2 / C-3 / C-6 / C-8 / C-11 验收；
- **直接解除** [frontend-workspace-and-practice/001-workspace-and-interview-context](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) Phase 3.3 的 `listResumes` disabled-list 模式阻塞（workspace 001 原地修订由其 owner 在本 plan Phase 5 完成后单独触发，不属于本 plan 范围）。

本 plan 不实现 backend handler（归未来 `backend-resume` / `backend-upload` subspec）；不实现 frontend Resume Workshop 业务页面（归未来 `frontend-resume-workshop`）；不改 `Resumes` tag 之外的契约。

## 2 背景

[engineering-roadmap §5.2](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) Resume Workshop 候选 subject 当前依赖 B2、B3、B4、A3、`C2 backend-upload`（roadmap 3.10 已显式化）。3.10 同步阶段 0 contract additive 升级路径：在创建 `backend-upload` / `backend-resume` / `frontend-resume-workshop` 三个新 subspec 之前，必须先扩容 B2 契约，避免下游 frontend mock-first plan 重复遭遇 `listResumes` 缺契约的二次阻塞（workspace 001 spec §3.2 / plan Phase 3.3 已显式记录"disabled-list 模式"）。

本 plan 是 [openapi-v1-contract spec §7 关联计划](../../spec.md#7-关联计划) 列出的第 4 个，承担 D-18 additive 升级的执行落地：把 spec 1.16 声明阶段（D-18 决策项 + §3.1.1 备注）实际投影到 `openapi/openapi.yaml`、`openapi/fixtures/`、`scripts/lint/`、generated client、`docs/spec/mock-contract-suite`、`docs/spec/engineering-roadmap` 等多个真理源；落地后 spec.md 同步升级到 1.17（§3.1.1 回填 #47-55 完整行、§2.1 endpoint 总数升 55）。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就有 B1 D-10 enum 与 B2 schema 真理源；Phase 2 起来就有 9 个 operation 进入 openapi.yaml + inventory lint pass；Phase 3 起来就有 fixtures default scenario 可被 mock-contract-suite 消费；Phase 4 把 spec §3.1.1 回填到 #47-55 完整冻结；Phase 5 收口验收 + 移交下游修订入口。

执行本 plan 前必须确认：

- [001-bootstrap](../001-bootstrap/plan.md) Phase 4 已完成：`openapi/openapi.yaml` v1.0.0 baseline 锁定形态。
- [002-fixtures-and-mock-source](../002-fixtures-and-mock-source/plan.md) Phase 6 已完成：fixture 同步工具与 `make validate-fixtures` 入口可用。
- [003-breaking-change-gate](../003-breaking-change-gate/plan.md) Phase 5 已完成：`make openapi-diff` 与 `openapi/baseline/openapi-v1.0.0.yaml` 就位；本 plan additive 升级时由 `make openapi-diff` 验证不触发 breaking。
- [B1 spec §3.1 D-10](../../../shared-conventions-codified/spec.md#31-已锁定决策)（声明阶段已完成；本 plan Phase 1 落地 `shared/conventions.yaml` 字面量与 generated 类型）。

## 3 质量门禁分类

- **Plan 类型**: `contract + tooling`。本 plan 落地 OpenAPI schema/operation、`shared/conventions.yaml` 枚举与错误码、generated Go/TS client、inventory lint 规则、fixtures；不实现业务 handler、前端 UI、用户 workflow。
- **TDD 策略**: 适用（Code plan requires TDD）。Red-Green-Refactor 入口：
  1. 每个新 schema 与 enum 写 OpenAPI lint + JSON Schema 校验断言；
  2. 每个新 operation 写 inventory lint negative case（缺 IK / 缺 provenance / 缺 fixture）；
  3. `make codegen-openapi` 后 `make codegen-check` 必须显示 Go/TS generated 类型对齐；
  4. `make openapi-diff` 验证 additive only，breaking 路径返回 exit 1；
  5. fixtures 通过 `make validate-fixtures` schema-valid + GenerationProvenance 完整。
  执行入口：`/implement openapi-v1-contract/004-resume-additive-coverage` → `/tdd`。
- **BDD 策略**: 不适用。本 plan 是纯 HTTP 契约 + tooling additive，不实现 runtime API handler 或用户行为流；用户可见的 Resume Workshop 流程由 `frontend-resume-workshop` / `backend-resume` / `backend-upload` subspec 各自维护 BDD gate，本 plan 以 operation matrix + fixture/schema/codegen gates 替代。
- **替代验证 gate**:
  - `npx @apidevtools/swagger-cli@4.0.4 swagger-cli validate openapi/openapi.yaml`
  - `make codegen-openapi && make codegen-check`
  - `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml`（断言 55 operation / 13 tag / 6 新 IK / AI_PROVENANCE_SCHEMAS 含 ResumeVersion）
  - `make validate-fixtures`（断言新 fixtures schema-valid + provenance 完整）
  - `make openapi-diff`（断言 additive only）
  - `sync-doc-index --check`（断言 spec/history/INDEX 同步）

### 3.1 Frontend / Backend Operation Matrix

本 plan 只落契约、fixtures、codegen 和 lint，不实现 runtime handler。矩阵用于区分 fixture-backed 可消费状态与真实 backend 状态，避免下游把 contract-ready 误判为 handler-ready。

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `listResumes` | `openapi/fixtures/Resumes/listResumes.json` `default` / `empty` / `paginated` | `frontend-resume-workshop/001` list view；`frontend-workspace-and-practice` ResumePicker unblock | `backend-resume/001` not-yet-implemented at this plan exit | `resume_assets` | none | current substitute gate: `make validate-fixtures` + mock-contract-suite 55-op inventory；downstream E2E.P0.034 / E2E.P0.036 |
| `listResumeVersions` | `openapi/fixtures/Resumes/listResumeVersions.json` `default` / `master-only` / `with-targeted-branches` | `frontend-resume-workshop/001` Tree/Flat views | future `backend-resume/002-versions-and-tailor-runs` | `resume_versions` | none | current substitute gate: `make validate-fixtures` + fixture scenario coverage；downstream E2E.P0.036 |
| `getResumeVersion` | `openapi/fixtures/Resumes/getResumeVersion.json` `master-default` / `targeted-with-suggestions` / `not-found-404` | `frontend-resume-workshop/001` detail preview | future `backend-resume/002-versions-and-tailor-runs` | `resume_versions` / `resume_version_suggestions` | none for read; provenance fields fixture-backed | current substitute gate: `make validate-fixtures` + `not-found-404` fixture；downstream E2E.P0.037 |
| `branchResumeVersion` | `openapi/fixtures/Resumes/branchResumeVersion.json` `copy-master-sync` / `blank-sync` / `ai-select-202-with-job` / `validation-error-422` / `idempotent-replay` | `frontend-resume-workshop/003` BranchFlow | future `backend-resume/002-versions-and-tailor-runs` | `resume_versions` + `async_jobs` for `ai_select` | `resume.tailor.*` only for `ai_select` | current substitute gate: `Idempotency-Key` inventory lint + `make validate-fixtures` over success/422/idempotent replay；TS client return type must union `ResumeVersion | BranchResumeVersionAccepted` |
| `updateResumeVersion` | `openapi/fixtures/Resumes/updateResumeVersion.json` `default` / `validation-error-422` | `frontend-resume-workshop/003` Edit Tab | future `backend-resume/002-versions-and-tailor-runs` | `resume_versions` | none | current substitute gate: `Idempotency-Key` inventory lint + `make validate-fixtures` over success/422 |
| `acceptResumeTailorSuggestion` | `openapi/fixtures/Resumes/acceptResumeTailorSuggestion.json` `default` / `conflict-409` | `frontend-resume-workshop/003` Rewrites Tab | future `backend-resume/002-versions-and-tailor-runs` | `resume_version_suggestions` | none | current substitute gate: `Idempotency-Key` inventory lint + `make validate-fixtures` over success/409 |
| `rejectResumeTailorSuggestion` | `openapi/fixtures/Resumes/rejectResumeTailorSuggestion.json` `default` / `conflict-409` | `frontend-resume-workshop/003` Rewrites Tab | future `backend-resume/002-versions-and-tailor-runs` | `resume_version_suggestions` | none | current substitute gate: `Idempotency-Key` inventory lint + `make validate-fixtures` over success/409 |
| `archiveResumeAsset` | `openapi/fixtures/Resumes/archiveResumeAsset.json` `default-202` / `already-archived-409` | `frontend-resume-workshop` archive action (future) | future `backend-resume/003-export-and-archive-and-delete` | `resume_assets.deleted_at` / status projection | none | current substitute gate: `Idempotency-Key` inventory lint + `make validate-fixtures` over 202/409 |
| `exportResumeVersion` | `openapi/fixtures/Resumes/exportResumeVersion.json` `p0-501-not-available` | `frontend-resume-workshop/001` toast fallback and future export action | P0 no real export handler beyond 501 stub; future `backend-resume/003` | none in P0; future file output requires separate upload purpose additive | none in P0 | current substitute gate: `make validate-fixtures` + B2 spec C-12 + inventory 501 allowlist；downstream E2E.P0.037 fallback/toast |

## 4 实施步骤

### Phase 1: B1 D-10 vocabulary 真理源同步

#### 1.1 落地 `shared/conventions.yaml` Resume vocabulary

在 `shared/conventions.yaml` 添加：
- `ResumeVersionType` 枚举：`structured_master` / `targeted`
- `ResumeSeedStrategy` 枚举：`copy_master` / `blank` / `ai_select`
- `ResumeTailorSuggestionStatus` 枚举：`pending` / `accepted` / `rejected`
- `RESUME_EXPORT_NOT_AVAILABLE` 错误码常量（前缀 `RESUME_*`）

同步 generator 输出 `backend/internal/shared/types/resume.go`（或同等位置）与 `frontend/src/lib/conventions/resume.ts`；shared-conventions-codified 14 枚举类型 → 17。

#### 1.2 同步 generated artifacts + B1 lint

运行 `make codegen-conventions && make codegen-check`，验证：
- Go side `errors.AllCodes` 含 `RESUME_EXPORT_NOT_AVAILABLE`
- TS side `frontend/src/lib/conventions/errors.ts` `as const` 字面量含 `RESUME_EXPORT_NOT_AVAILABLE`
- 3 个新 enum 在 Go/TS 双端生成 idempotent

### Phase 2: OpenAPI schema + operation 落地

#### 2.1 新增 7 个 schema + `RegisterResumeRequest` 扩展

在 `openapi/openapi.yaml` `components.schemas` 添加：
- `ResumeVersion`（含 `structuredProfile` 字段必带 `GenerationProvenance`）
- `PaginatedResumeAsset`（沿用 D-5 cursor pagination）
- `PaginatedResumeVersion`
- `BranchResumeVersionRequest`（必含 `parentVersionId` / `targetJobId` / `seedStrategy`、optional `focusAngle` / `displayName`）
- `UpdateResumeVersionRequest`（all fields optional, PATCH 语义）
- `ResumeTailorSuggestionStatus` 引用 B1 D-10

扩展 `RegisterResumeRequest`：新增 optional `sourceType` ∈ `upload | paste | guided`、`rawText`（paste 路径）、`guidedAnswers`（guided 路径，schema 为 JSON object，后端持久化到 B4 `resume_assets.guided_answers` jsonb），保持现有 `fileObjectId` / `title` / `language` 向后兼容。

#### 2.2 新增 9 个 operationId

在 `Resumes` tag 下新增 9 个 op；suggestion accept/reject 虽处理 tailor suggestion，但 URL 仍挂在 resume version 资源下，按 D-18 统一归 `Resumes` tag 与 `openapi/fixtures/Resumes/`，不改 `ResumeTailor` tag：

| operationId | Tag | Method | URL | IK | Async/Sync |
|-------------|-----|--------|-----|----|------------|
| `listResumes` | Resumes | GET | /api/v1/resumes | — | sync |
| `listResumeVersions` | Resumes | GET | /api/v1/resumes/{resumeAssetId}/versions | — | sync |
| `getResumeVersion` | Resumes | GET | /api/v1/resume-versions/{resumeVersionId} | — | sync |
| `branchResumeVersion` | Resumes | POST | /api/v1/resume-versions | ✓ | sync（`seedStrategy=ai_select` 在后端入队 tailor job） |
| `updateResumeVersion` | Resumes | PATCH | /api/v1/resume-versions/{resumeVersionId} | ✓ | sync |
| `acceptResumeTailorSuggestion` | Resumes | POST | /api/v1/resume-versions/{resumeVersionId}/suggestions/{suggestionId}/accept | ✓ | sync |
| `rejectResumeTailorSuggestion` | Resumes | POST | /api/v1/resume-versions/{resumeVersionId}/suggestions/{suggestionId}/reject | ✓ | sync |
| `archiveResumeAsset` | Resumes | POST | /api/v1/resumes/{resumeAssetId}/archive | ✓ | sync |
| `exportResumeVersion` | Resumes | POST | /api/v1/resume-versions/{resumeVersionId}/exports | ✓ | **P0**: `501` + `error.code = "RESUME_EXPORT_NOT_AVAILABLE"`；**P1**: `202 + Job(jobType=resume_export)` |

每个 operation 必须含 request body（若有）、success response、error response `$ref` `ApiErrorResponse`，且 §6 C-1 inventory lint 通过。

`exportResumeVersion` 引入第二个 P0 `501 Not Implemented` 例外时，必须同步：
- B2 spec §4.1 P0 exception matrix，从 privacy-only 改为 privacy export + resume version export；
- B2 spec §6 新增 C-12 `resume export 501` 验收行；
- `scripts/lint/openapi_inventory.py` 的 501 allowlist / negative fixture gate，避免继续把 501 规则写死为 privacy-only；
- B2 history.md 明确记录 `RESUME_EXPORT_NOT_AVAILABLE` 与 `exportResumeVersion` 501 stub 属 D-18 plan 004 落地。

### Phase 3: Fixtures 与 mock-contract-suite 同步

#### 3.1 落地 9 个 operation 的 fixtures

在 `openapi/fixtures/Resumes/` 新增（参照 §4.7 fixtures 约束；本 plan 不在其它 tag 目录新增这 9 个 operation 的 fixture 文件）：

- `listResumes.json`：`default`（多条）/ `empty`（空列表）/ `paginated`（带 nextCursor）
- `listResumeVersions.json`：`default` / `master-only` / `with-targeted-branches`
- `getResumeVersion.json`：`master-default` / `targeted-with-suggestions` / `not-found-404`
- `branchResumeVersion.json`：`copy-master-sync` / `blank-sync` / `ai-select-202-with-job` / `validation-error-422` / `idempotent-replay`
- `updateResumeVersion.json`：`default` / `validation-error-422`
- `acceptResumeTailorSuggestion.json` / `rejectResumeTailorSuggestion.json`：`default` / `conflict-409`（已 accept 再 reject）
- `archiveResumeAsset.json`：`default-202` / `already-archived-409`
- `exportResumeVersion.json`：`p0-501-not-available`（带 `error.code = RESUME_EXPORT_NOT_AVAILABLE`）

每个 fixture 必须遵守 `openapi/fixtures/README.md` 形态：顶层 `operationId` + `scenarios` map，且 `scenarios.default` 是第一项。Resume Workshop 的原型数据当前定义在 `ui-design/src/screen-resume-workshop.jsx`，不在 `ui-design/src/data.jsx`；本 plan 不强制生成 `prototype-baseline`，除非先修订 `openapi/fixtures/PROTOTYPE_MAPPING.md` 与同步工具支持该源文件。

#### 3.2 同步 mock-contract-suite spec

原地修订 `docs/spec/mock-contract-suite/spec.md`（1.5 → 1.6），inventory 章节加入 9 个新 op 与对应 fixture variant 计数；同步 `openapi/fixtures/README.md` tag/operation 计数（13 tag / 55 op）。

### Phase 4: Inventory lint + spec §3.1.1 回填

#### 4.1 更新 inventory lint 真理源

- `scripts/lint/openapi_inventory.py`:
  - `EXPECTED_OPERATIONS = 55`
  - `IK_REQUIRED` 追加 6 项（branch/update/accept/reject/archive/export）
  - `AI_PROVENANCE_SCHEMAS` 追加 `ResumeVersion`
  - P0 `501` allowlist 追加 `exportResumeVersion`，并保留其他 endpoint 501 负向拦截
- `scripts/lint/validate_fixtures.py` 注释升级（46 → 55）
- `openapi/README.md` validator 描述同步

#### 4.2 回填 B2 spec §3.1.1 endpoint 列表

把 9 个新 operation 作为 #47-55 行追加到 §3.1.1 endpoint 列表（参照 #35-46 JobMatch additive 模式）；§3.1.1 末尾 "总计 46 个 endpoint" → "总计 55 个 endpoint"；§2.1 "46 端点" → "55 端点"。同步 §4.1 P0 例外、§4.6 AI provenance schema 列表（追加 `ResumeVersion`）与 §6 C-12 `resume export 501` 验收行。

修订 `docs/spec/openapi-v1-contract/spec.md` 版本 1.16 → 1.17，`history.md` 追加 1.17 行（"D-18 落地阶段：§3.1.1 回填 #47-55 + §2.1 总数升 55"）。

#### 4.3 同步顶层 roadmap 表述

`docs/spec/engineering-roadmap/spec.md` §4.3 / §5.1 中如出现 "46 endpoint / 46 operation" 文字描述，同步升级到 55；版本 3.12 → 3.13，`history.md` 追加 3.13 行。

### Phase 5: 验收 + 解锁下游

#### 5.1 跨 gate 收口

按 §3 替代验证 gate 依序运行：
- `npx @apidevtools/swagger-cli@4.0.4 swagger-cli validate openapi/openapi.yaml` PASS
- `make codegen-openapi && make codegen-check` PASS（无 diff）
- `python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml` PASS（55 op / 13 tag）
- `make validate-fixtures` PASS（含 GenerationProvenance 完整性）
- `make openapi-diff` PASS（additive only）
- `sync-doc-index --check` PASS

#### 5.2 解锁 workspace 001 阻塞

通知 [frontend-workspace-and-practice/001-workspace-and-interview-context](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) owner 启动原地修订：
- workspace spec §3.2 第 1 条待确认事项标 "已通过 B2 D-18 additive 升级解除"
- workspace plan 001 Phase 3.3 disabled-list 模式 → active-list 模式
- workspace bdd-checklist Resume Picker 场景从 disabled-list → active-list
- `frontend/src/app/screens/workspace/modals/ResumePickerModal.tsx` 启用 listResumes generated client
- workspace i18n `workspace.resumePicker.disabledNote` 移除

本 plan 不直接修订 workspace 文件，只在 Phase 5.2 完成 "可解锁" 信号传递；workspace owner 在阶段 2 独立完成修订。

### Phase 6: L2 remediation - generated client response shape closure

#### 6.1 修复 `branchResumeVersion` 202 response type

将 `seedStrategy=ai_select` 的 `202` response body 提升为命名 schema `BranchResumeVersionAccepted`，并更新 TS generator 让多成功状态码生成 union return type，避免 frontend caller 把 `{resumeVersionId, version, job}` 误读为 `ResumeVersion`。

#### 6.2 修复 P0 export 501 typed response path

更新 TS generated client request plumbing：显式声明在 OpenAPI 中的 P0 `501` response（`requestPrivacyExport` / `exportResumeVersion`）应按 typed response 解析，不应被 generic non-OK throw 提前截断；默认 / 未声明的 4xx/5xx 仍必须 throw。

#### 6.3 补齐 dev fixture consumer 同步

将 Resume Workshop 9 个新 operation fixture 接入 `frontend/src/api/devMockClient.ts`，并用 focused Vitest 覆盖 generated operationId 完整性与 `exportResumeVersion` typed 501 fallback。

### Phase 7: D-20 简历扁平化 contract collapse

> product-scope D-20 / B2 本地 D-26。把 Resume Workshop 版本树契约坍缩为扁平 `resumes` 资产；同步完成 product-scope D-17 残留的 §2.1/§3.1.1 endpoint 计数 reconcile。Red 优先：先把 `scripts/lint/openapi_inventory.py` EXPECTED_OPERATIONS 改 48→43（lint 先红），再改 `openapi.yaml`。本 phase 与 `db-migrations-baseline/002` Phase 6（migration）+ `backend-resume` D-20 phase 同属 D-20 contract impl，原子提交（契约与生成物互依）。

#### 7.1 删除 6 个版本/suggestion operation

从 `openapi/openapi.yaml` 删除 `confirmResumeStructuredMaster` / `listResumeVersions` / `getResumeVersion` / `branchResumeVersion` / `acceptResumeTailorSuggestion` / `rejectResumeTailorSuggestion` 的 path + operation；删除对应 6 个 `openapi/fixtures/Resumes/*.json`。

（验证：`npx -p @apidevtools/swagger-cli@4.0.4 swagger-cli validate openapi/openapi.yaml` PASS；`grep -c "operationId:" openapi/openapi.yaml` = 43）

#### 7.2 重命名 3 op + resumeId 路径统一

`updateResumeVersion`→`updateResume`（`PATCH /api/v1/resumes/{resumeId}`）、`archiveResumeAsset`→`archiveResume`（`POST /api/v1/resumes/{resumeId}/archive`）、`exportResumeVersion`→`exportResume`（`POST /api/v1/resumes/{resumeId}/exports`，保留 `501` + `RESUME_EXPORT_NOT_AVAILABLE`）；`getResume` 路径参数 `resumeAssetId`→`resumeId`；`RequestResumeTailorRequest.resumeVersionId`→`resumeId`（required，tailor 作用于扁平 resume）。重命名 / 移动对应 fixtures。

（验证：`swagger-cli validate` PASS；`rg "resumeAssetId|resumeVersionId|resume-versions" openapi/openapi.yaml` 0 命中）

#### 7.3 新增 duplicateResume operation

新增 `POST /api/v1/resumes/{resumeId}/duplicate` operationId `duplicateResume`（IK 必带），request `DuplicateResumeRequest`（optional `structuredProfile` 覆盖 + `displayName`），success `201 + Resume`；新增 fixture `openapi/fixtures/Resumes/duplicateResume.json`（含 GenerationProvenance）。

（验证：fixture schema-valid；`make validate-fixtures` PASS）

#### 7.4 schema 扁平化

`ResumeAsset`→`Resume`（新增 `structuredProfile` object + `displayName`；保留只读 `sourceType`∈{`upload`,`paste`} / `rawText` / `parsedTextSnapshot` / `fileObjectId`）、`ResumeAssetWithJob`→`ResumeWithJob`、`PaginatedResumeAsset`→`PaginatedResume`；新增 `UpdateResumeRequest` / `DuplicateResumeRequest`；删除 `ResumeVersion` / `BranchResumeVersionRequest` / `BranchResumeVersionAccepted` / `PaginatedResumeVersion` / `ConfirmResumeStructuredMasterRequest` / `UpdateResumeVersionRequest`；`RegisterResumeRequest.sourceType` 收敛 {`upload`,`paste`}（删除 `guided` + `guidedAnswers`）；删除 `ResumeVersionType` / `ResumeSeedStrategy` / `ResumeTailorSuggestionStatus` 三个 enum 的 `$ref`（随 [B1 D-20](../../../shared-conventions-codified/spec.md) 退役）。

（验证：`swagger-cli validate` PASS；`$ref` 无悬空；`rg "ResumeVersion|BranchResumeVersion" openapi/openapi.yaml` 0 命中）

#### 7.5 inventory + fixture lint 真理源

`scripts/lint/openapi_inventory.py`：`EXPECTED_OPERATIONS` 48→43、`EXPECTED_TAGS` 维持 12、operation set 删 6 + 改名 3 + 加 `duplicateResume`、`IK_REQUIRED` 调整（删 branch/accept/reject/confirm，加 `updateResume`/`duplicateResume`/`archiveResume`/`exportResume`）、`AI_PROVENANCE_SCHEMAS` `ResumeVersion`→`Resume`、501 allowlist `exportResumeVersion`→`exportResume`；`scripts/lint/validate_fixtures.py` 注释计数 48→43。

（验证：`python3 scripts/lint/openapi_inventory.py openapi/openapi.yaml` PASS 43 op/12 tag；`make validate-fixtures` PASS；未登记/计数不符时 lint 先红后绿）

#### 7.6 codegen + baseline re-freeze

`make codegen-openapi`（Go server/types + TS client/types 重生）；`make codegen-check`（提交后 `git diff --exit-code` 清）；`openapi/baseline/openapi-v1.0.0.yaml` 原地 re-freeze（pre-launch correction，参考 D-25 模式）；`openapi/diff-config.yaml` `endpointCount` 同步 43；`make openapi-diff` PASS（删除/重命名属 pre-launch correction，按 D-25 同款 re-freeze 处理，非 breaking gate violation）。

（验证：`make codegen-openapi && make codegen-check` exit 0；`make openapi-diff` PASS）

#### 7.7 spec / README / roadmap 计数同步

B2 spec 1.29→1.30（本次 doc 修订已完成）：§1 / §2.1 / §3.1.1（43 op/12 tag + D-17 残留 reconcile）/ §4.1 / §4.2 / D-9 / 新增 D-26 / 退役 D-18/D-23/D-24；history 1.30；`openapi/README.md` + `openapi/fixtures/README.md` 计数；`docs/spec/mock-contract-suite` + `docs/spec/engineering-roadmap` resume op / count 表述。

（验证：`sync-doc-index --check` 零漂移；README 计数与 inventory 一致）

#### 7.8 零残留收口

负向 grep：`rg "resumeVersionId|resumeAssetId|ResumeVersion|ResumeAsset|branchResume|listResumeVersions|getResumeVersion|confirmResumeStructuredMaster|acceptResumeTailorSuggestion|rejectResumeTailorSuggestion|ResumeSeedStrategy|ResumeTailorSuggestionStatus" openapi/ scripts/lint/ backend/internal/api/generated frontend/src/api/generated` 0 命中（除 `history.md` 历史行 + 负向断言）。

（验证：负向 grep 0 命中 + 7.1-7.7 全 gate PASS）

#### 7.10 L2 hardening - retired fixture key gate

`scripts/lint/validate_fixtures.py` 必须递归扫描 fixture request / response key，拒绝 D-20 退役的 `resumeAssetId` / `resumeVersionId` 字段；`openapi/fixtures/Debriefs/suggestDebriefQuestions.json` 等下游 fixture 必须使用 `resumeId`。该 gate 与 `make validate-fixtures` 同步执行，防止 OpenAPI schema 已坍缩但 fixture / mock consumer 继续携带旧字段。

（验证：`python3 -m unittest scripts.lint.validate_fixtures_cli_test` PASS；`make validate-fixtures` PASS；`rg -n "resumeVersionId|resumeAssetId" openapi/fixtures openapi/openapi.yaml backend/internal/api/generated frontend/src/api/generated` 0 命中）

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成
- §3 替代验证 gate 全部通过
- B2 spec §3.1.1 已回填 #47-55 endpoint 完整行
- B2 spec §2.1 endpoint 总数已升级到 55
- 3 个 enum + 1 个错误码已在 B1 generated artifacts 中出现
- 9 个新 operation 的 fixtures 全部 schema-valid
- `frontend-workspace-and-practice/001` owner 已收到 listResumes 解锁信号

**D-20 简历扁平化 contract collapse（Phase 7）验收**（product-scope D-20 / B2 D-26）：

- `openapi/openapi.yaml` 43 operation / 12 tag：删除 6 个版本/suggestion op、3 个重命名（updateResume/archiveResume/exportResume）、新增 `duplicateResume`、`resumeAssetId`/`resumeVersionId`→`resumeId`
- schema `ResumeAsset`→`Resume`（含 `structuredProfile`/`displayName`）、`PaginatedResumeAsset`→`PaginatedResume`、新增 `UpdateResumeRequest`/`DuplicateResumeRequest`、删除 `ResumeVersion` 等 6 个版本 schema；`RequestResumeTailorRequest.resumeVersionId`→`resumeId`；`RegisterResumeRequest.sourceType` 收敛 {upload,paste}
- `scripts/lint/openapi_inventory.py` EXPECTED_OPERATIONS 48→43、`AI_PROVENANCE_SCHEMAS` `ResumeVersion`→`Resume`；`make validate-fixtures` / `make codegen-openapi && make codegen-check` / `make openapi-diff` 全 PASS；baseline 原地 re-freeze
- B2 spec 1.29→1.30 + history 1.30 + README/roadmap/mock-contract-suite 计数同步；`sync-doc-index --check` 零漂移
- 负向 grep `resumeVersionId|resumeAssetId|ResumeVersion|branchResume|...` 在 openapi/lint/generated 0 命中（除 history 历史行 + 负向断言）；`validate_fixtures.py` 递归拒绝 fixture request / response 中的 D-20 退役 key
- 同步修正 product-scope D-17 残留的 §2.1/§3.1.1 endpoint 计数漂移（60→48→43）

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: `RegisterResumeRequest` 扩展引入 optional 字段，可能引起 frontend 已实现的 ResumePickerModal generated client 版本漂移 | additive only 不破坏现有调用；frontend codegen 同步重生即可；workspace 001 修订时再启用新字段 |
| R2: `archiveResumeAsset` 与 `deleteResumeAsset` 语义冲突（删除已由 `DELETE /api/v1/me` privacy 链路实现） | 明确 archive 是 user-level soft hide（不入 privacy_delete job），`deleted_at` 仅承载 admin 硬删；具体 store 行为留 backend-resume/001 锁定 |
| R3: `exportResumeVersion` P0 走 501，未来 P1 切到 202+Job 时可能被误判为 breaking | 本 plan 只锁 P0 `501 + RESUME_EXPORT_NOT_AVAILABLE`，不落 202；未来 backend-resume/003 或对应 B2 follow-up 执行 501 → 202 时，再按 D-12 privacy export 模式补 history + diff-config 白名单 |
| R4: B1 D-10 与 B2 D-18 落地不同步 | 本 plan Phase 1 + Phase 2 在同一 PR 提交；`make codegen-check` cross-language drift gate 强制双端对齐 |
| R5: fixtures provenance 覆盖不全（特别是 ai_select 路径） | Phase 3 fixture 设计 explicit `provenance` block；`make validate-fixtures` 通过 `AI_PROVENANCE_SCHEMAS` enforce |
| R6: workspace 001 原地修订需额外授权 | 本 plan 只负责契约升级；workspace owner 在收到 5.2 信号后独立启动 plan 1.2 → 1.3 修订；不创建 sibling |
