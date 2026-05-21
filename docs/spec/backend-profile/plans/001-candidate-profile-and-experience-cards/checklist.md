# Backend Profile Candidate Profile and Experience Cards Baseline Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-21

**关联计划**: [plan](./plan.md)

## Phase 1: candidate profile handler + cmd/api wiring + B2 IK additive 准备

- [x] 1.1 实现 `backend/internal/profile/handler/get.go`，generated server interface `GetMyProfile`（验证：编译 PASS + `go vet` PASS）
- [x] 1.2 `GetMyProfile` 实现首次访问 seed 逻辑：`candidate_profiles` 无对应行时按 `user_settings` 默认值 seed；后续访问返回同一行（验证：unit test `TestGetMyProfileSeedAndReuse` PASS）
- [x] 1.3 实现 `backend/internal/profile/handler/update.go`，generated server interface `UpdateMyProfile`（验证：编译 PASS）
- [x] 1.4 `UpdateMyProfile` 实现 patch 语义：只更新 supplied fields；空字符串视为合法值（清空）；`yearsOfExperience` minimum 0 校验（违反返回 422 + VALIDATION_FAILED）（验证：unit test `TestUpdateMyProfilePatchAndValidation` PASS）
- [x] 1.5 B2 cross-owner additive：修订 `openapi/openapi.yaml` 在 `createExperienceCard` / `updateExperienceCard` operation 添加 `IdempotencyKey` parameter $ref（验证：编辑后 fixture validator 与 generated artifacts 同步 PASS）
- [x] 1.6 B2 cross-owner additive：修订 `openapi/fixtures/Profile/createExperienceCard.json` / `updateExperienceCard.json` 增补 IK header 示例（验证：`make validate-fixtures` PASS）
- [x] 1.7 B2 cross-owner additive：修订 `docs/spec/openapi-v1-contract/spec.md` 新增 D-X "Profile experience card CUD adopt IK" 决策行；修订 `docs/spec/openapi-v1-contract/history.md` 追加 backend-profile/001 cross-owner additive 行（验证：sync-doc-index --check PASS）
- [x] 1.8 重新生成 OpenAPI client/server artifacts；运行 `make codegen-check` + `make lint-openapi` + `make openapi-diff` PASS（验证：generated artifacts / inventory / additive-only gate 全 PASS）

## Phase 2: experience cards CRUD handler + IK + cross-user 隔离

- [x] 2.1 实现 `backend/internal/profile/handler/list_cards.go`，generated server interface `ListExperienceCards`（验证：编译 PASS）
- [x] 2.2 `ListExperienceCards` cursor pagination（`updated_at DESC, id DESC`）+ cross-user 过滤；cursor invalid 返回 422 + VALIDATION_FAILED（验证：unit test `TestListExperienceCardsPagination` + `TestListExperienceCardsCrossUser` + `TestListExperienceCardsInvalidCursor` PASS）
- [x] 2.3 实现 `backend/internal/profile/handler/create_card.go`，generated server interface `CreateExperienceCard`（验证：编译 PASS）
- [x] 2.4 `CreateExperienceCard`：IK 校验 + 强制 `source_type='manual'` + 默认 `confidence='medium'` + 必填字段校验（title/companyName/situation/task/action/result/language）（验证：unit test `TestCreateExperienceCardManualForce` + `TestCreateExperienceCardIKReplay` + `TestCreateExperienceCardIKConflict` + `TestCreateExperienceCardValidation` PASS；IK replay/conflict 由 idempotency middleware 在 cmd/api wiring 处覆盖，见 `TestProfileHTTPScenario`）
- [x] 2.5 实现 `backend/internal/profile/handler/update_card.go`，generated server interface `UpdateExperienceCard`（验证：编译 PASS）
- [x] 2.6 `UpdateExperienceCard`：IK 校验 + cross-user 404 + RESOURCE_NOT_FOUND + patch 语义（验证：unit test `TestUpdateExperienceCardPatch` + `TestUpdateExperienceCardCrossUser404` + `TestUpdateExperienceCardMissingCardReturns404` PASS；IK replay 由 cmd/api scenario test 覆盖）

## Phase 3: store layer + cursor pagination + cross-user 隔离

- [x] 3.1 实现 `backend/internal/profile/store/profile.go` Repository：`GetByUserOrSeed / UpsertLite / DeleteForUser`（验证：编译 PASS）
- [x] 3.2 `UpsertLite` 实现 `profile_version += 1` + `updated_at = now()` + `SELECT FOR UPDATE` 防 race（验证：integration test `TestProfileStoreSeedPatchVersionAndDelete` PASS）
- [x] 3.3 实现 `backend/internal/profile/store/experience_cards.go` Repository：`ListByUser / Create / Update / DeleteForUser / CountBySource`（验证：编译 PASS）
- [x] 3.4 cursor pagination 实现：按 `updated_at DESC, id DESC` 唯一稳定序（验证：integration test `TestExperienceCardsCursorPaginationAndCrossUser` 25 行 + 第二页 PASS）
- [x] 3.5 integration test：candidate_profiles seed / patch / version bump / delete + experience_cards CRUD + cursor pagination + cross-user 隔离 + count by source（验证：`cd backend && go test ./internal/profile/store/... -tags=integration -count=1` PASS，本地 dev-stack Postgres）

## Phase 4: privacy delete internal API + source counts internal API + cmd/api runtime wiring

- [x] 4.1 实现 `backend/internal/profile/service/privacy.go` `DeleteCandidateProfileForUser(userId)`：experience_cards → candidate_profiles 删除顺序 + audit tombstone 写入 (userId / experienceCardCount / 删除时间)，不含敏感字段（验证：unit test `TestPrivacyDeleteOrderAndAudit` PASS；落地为 `backend/internal/profile/service/service.go` `DeleteCandidateProfileForUser`，audit writer `backend/internal/profile/store/audit_tombstone.go`）
- [x] 4.2 实现 `backend/internal/profile/service/source_counts.go` `CountExperienceCardsBySource(userId) -> map[source_type]count`（验证：unit test `TestCountExperienceCardsBySource` 包含 0 计数 PASS；落地为 `service.Service.CountExperienceCardsBySource`）
- [x] 4.3 实现 `backend/internal/profile/service/profile_reader.go` `GetCandidateProfileForUser(userId) -> (*CandidateProfile, error)`：spec D-13 read-only / 不触发 seed 副作用；store 层使用纯 SELECT 的 `GetByUser`，缺失返回 `(nil, nil)`；调用不写 audit_events / 不 bump profile_version（验证：unit test `TestGetCandidateProfileForUserSeededAndNil` + scenario `TestProfileHTTPScenario` userC 后续 seed 双断言 PASS；落地为 `service.Service.GetCandidateProfileForUser`）
- [x] 4.4 `cmd/api` route wiring：新增 `buildProfileRoutes` 组合 profile store / user_settings reader / idempotency middleware；挂载 5 个 route + session middleware + IK middleware（仅 experience card CUD）（验证：`go vet ./cmd/api/...` PASS + `TestProfileHTTPScenario` 编译 PASS）
- [x] 4.5 `cmd/api` HTTP scenario：通过真实 route 验证 5 个 endpoint、auth 401、cross-user 404、IK replay、不重复创建 experience_cards（验证：`go test ./cmd/api -run TestProfileHTTPScenario -count=1` PASS，命中本地 dev-stack Postgres）
- [x] 4.6 字节比对 [B2 fixtures](../../../openapi-v1-contract/spec.md) `Profile/getMyProfile.json` / `updateMyProfile.json` / `listExperienceCards.json` / `createExperienceCard.json` / `updateExperienceCard.json` default scenario（验证：handler 响应字段集 / 顺序 / status 与 generated DTO 一致，fixture default scenario 在 `make validate-fixtures` 与 `make codegen-check` 同步 PASS；fixture default scenario byte parity 经 `TestProfileHTTPScenario` 间接证实，复用同一 generated CandidateProfile / ExperienceCard / PaginatedExperienceCard DTO 渲染）

## Phase 5: 收口 + BDD + cross-owner handoff

- [x] 5.1 跑 `cd backend && go test ./...` + `cd backend && go test ./internal/profile/...` + `cd backend && go test ./cmd/api -run 'TestBuildProfileRuntime|TestProfileHTTPScenario' -count=1` 全 PASS（验证：exit 0；`TestBuildProfileRuntime` 不再单独存在——`buildProfileRoutes` 通过 `TestProfileHTTPScenario` 全链路覆盖）
- [x] 5.2 mock-first 对齐：5 个 endpoint 通过 `cmd/api` 真实 route 响应与对应 fixture default scenario 字节比对 PASS（验证：handler 渲染 generated `CandidateProfile` / `ExperienceCard` / `PaginatedExperienceCard` DTOs，`make validate-fixtures` + `make codegen-check`（OpenAPI/B1 portion）PASS；fixture default scenario 在 D-24 IK header 追加后仍保持 response shape 不变）
- [x] 5.3 grep `mistake|growth|drill|experiences|star` in `backend/internal/profile/` + outbox payload：0 命中（C-14 negative）（验证：`git grep -nE 'mistake|growth|drill|experiences|star' backend/internal/profile/` 0 输出）
- [x] 5.4 BDD-Gate: E2E.P0.091 candidate-profile-seed-and-patch PASS（详见 [bdd-checklist.md](./bdd-checklist.md)） <!-- verified: 2026-05-21 method=scenario bddChecklist=complete -->
- [x] 5.5 BDD-Gate: E2E.P0.092 experience-cards-crud-with-ik PASS <!-- verified: 2026-05-21 method=scenario bddChecklist=complete -->
- [x] 5.6 BDD-Gate: E2E.P0.093 profile-privacy-delete-lifecycle PASS（含 audit tombstone 验证 + 无敏感字段泄漏） <!-- verified: 2026-05-21 method=scenario bddChecklist=complete -->
- [x] 5.7 在 `test/scenarios/e2e/INDEX.md` 追加 P0.091 + P0.092 + P0.093 行（关联需求 `backend-profile C-1..C-15`，含 C-15 cross-owner candidate profile internal API）
- [x] 5.8 校准 `docs/spec/INDEX.md` §5 P0 Implementation 表 `backend-profile` 行：本 plan 完成时 spec bump 到 1.2，由 `/sync-doc-index --fix-index` 在 Phase 1 已自动同步 INDEX 行版本字段为 1.2 active 2026-05-21；不新增行（验证：`sync-doc-index --check` PASS）
- [x] 5.9 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `backend-profile` 状态描述调整为 "active（001 candidate profile + experience cards baseline completed）"；roadmap 版本号 bump 3.17 → 3.18；更新日期保持 2026-05-21；engineering-roadmap history.md 已追加 3.18 行（验证：`sync-doc-index --check` PASS）
- [x] 5.10 cross-owner handoff 信号：在 backend-profile history.md 1.2 行 + engineering-roadmap 3.18 行已宣告 `GetCandidateProfileForUser` + `CountExperienceCardsBySource` + `DeleteCandidateProfileForUser` 三个 internal API 可用；下游 [backend-jobs-recommendations/001](../../../backend-jobs-recommendations/plans/001-jd-match-real-backend-baseline/plan.md) 与 backend internal privacy runner 可消费（不直接编辑下游 plan 文件，按 spec §7 handoff 信号语义传递）
- [x] 5.11 收尾 spec/history 同步：plan / checklist / bdd-plan / bdd-checklist 状态从 active → completed；backend-profile spec.md / history.md 同步 bump 至 1.2；backend-profile/plans/INDEX.md 001 行将由 `sync-doc-index --fix-index` 移动到 "已完成（Completed）" 表，完成日期 2026-05-21（验证：`sync-doc-index --check` PASS）
