# Backend Profile Candidate Profile and Experience Cards Baseline Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-21

**关联计划**: [plan](./plan.md)

## Phase 1: candidate profile handler + cmd/api wiring + B2 IK additive 准备

- [ ] 1.1 实现 `backend/internal/profile/handler/get.go`，generated server interface `GetMyProfile`（验证：编译 PASS + `go vet` PASS）
- [ ] 1.2 `GetMyProfile` 实现首次访问 seed 逻辑：`candidate_profiles` 无对应行时按 `user_settings` 默认值 seed；后续访问返回同一行（验证：unit test `TestGetMyProfileSeedAndReuse` PASS）
- [ ] 1.3 实现 `backend/internal/profile/handler/update.go`，generated server interface `UpdateMyProfile`（验证：编译 PASS）
- [ ] 1.4 `UpdateMyProfile` 实现 patch 语义：只更新 supplied fields；空字符串视为合法值（清空）；`yearsOfExperience` minimum 0 校验（违反返回 422 + VALIDATION_FAILED）（验证：unit test `TestUpdateMyProfilePatchAndValidation` PASS）
- [ ] 1.5 B2 cross-owner additive：修订 `openapi/openapi.yaml` 在 `createExperienceCard` / `updateExperienceCard` operation 添加 `IdempotencyKey` parameter $ref（验证：编辑后 fixture validator 与 generated artifacts 同步 PASS）
- [ ] 1.6 B2 cross-owner additive：修订 `openapi/fixtures/Profile/createExperienceCard.json` / `updateExperienceCard.json` 增补 IK header 示例（验证：`make validate-fixtures` PASS）
- [ ] 1.7 B2 cross-owner additive：修订 `docs/spec/openapi-v1-contract/spec.md` 新增 D-X "Profile experience card CUD adopt IK" 决策行；修订 `docs/spec/openapi-v1-contract/history.md` 追加 backend-profile/001 cross-owner additive 行（验证：sync-doc-index --check PASS）
- [ ] 1.8 重新生成 OpenAPI client/server artifacts；运行 `make codegen-check` + `make lint-openapi` + `make openapi-diff` PASS（验证：generated artifacts / inventory / additive-only gate 全 PASS）

## Phase 2: experience cards CRUD handler + IK + cross-user 隔离

- [ ] 2.1 实现 `backend/internal/profile/handler/list_cards.go`，generated server interface `ListExperienceCards`（验证：编译 PASS）
- [ ] 2.2 `ListExperienceCards` cursor pagination（`updated_at DESC, id DESC`）+ cross-user 过滤；cursor invalid 返回 422 + VALIDATION_FAILED（验证：unit test `TestListExperienceCardsPagination` + `TestListExperienceCardsCrossUser` + `TestListExperienceCardsInvalidCursor` PASS）
- [ ] 2.3 实现 `backend/internal/profile/handler/create_card.go`，generated server interface `CreateExperienceCard`（验证：编译 PASS）
- [ ] 2.4 `CreateExperienceCard`：IK 校验 + 强制 `source_type='manual'` + 默认 `confidence='medium'` + 必填字段校验（title/companyName/situation/task/action/result/language）（验证：unit test `TestCreateExperienceCardManualForce` + `TestCreateExperienceCardIKReplay` + `TestCreateExperienceCardIKConflict` + `TestCreateExperienceCardValidation` PASS）
- [ ] 2.5 实现 `backend/internal/profile/handler/update_card.go`，generated server interface `UpdateExperienceCard`（验证：编译 PASS）
- [ ] 2.6 `UpdateExperienceCard`：IK 校验 + cross-user 404 + RESOURCE_NOT_FOUND + patch 语义（验证：unit test `TestUpdateExperienceCardPatch` + `TestUpdateExperienceCardCrossUser404` + `TestUpdateExperienceCardIKReplay` PASS）

## Phase 3: store layer + cursor pagination + cross-user 隔离

- [ ] 3.1 实现 `backend/internal/profile/store/profile.go` Repository：`GetByUserOrSeed / UpsertLite / DeleteForUser`（验证：编译 PASS）
- [ ] 3.2 `UpsertLite` 实现 `profile_version += 1` + `updated_at = now()` + `SELECT FOR UPDATE` 防 race（验证：integration test `TestProfileVersionMonotonic` PASS）
- [ ] 3.3 实现 `backend/internal/profile/store/experience_cards.go` Repository：`ListByUser / Create / Update / DeleteForUser / CountBySource`（验证：编译 PASS）
- [ ] 3.4 cursor pagination 实现：按 `updated_at DESC, id DESC` 唯一稳定序（验证：integration test `TestExperienceCardsCursorPagination` 25+ 行 + 第二页 PASS）
- [ ] 3.5 integration test：candidate_profiles seed / patch / version bump / delete + experience_cards CRUD + cursor pagination + cross-user 隔离 + count by source（验证：`cd backend && go test ./internal/profile/store/... -tags=integration -count=1` PASS）

## Phase 4: privacy delete internal API + source counts internal API + cmd/api runtime wiring

- [ ] 4.1 实现 `backend/internal/profile/service/privacy.go` `DeleteCandidateProfileForUser(userId)`：experience_cards → candidate_profiles 删除顺序 + audit tombstone 写入 (userId / experienceCardCount / 删除时间)，不含敏感字段（验证：unit test `TestPrivacyDeleteOrderAndAudit` PASS）
- [ ] 4.2 实现 `backend/internal/profile/service/source_counts.go` `CountExperienceCardsBySource(userId) -> map[source_type]count`（验证：unit test `TestCountExperienceCardsBySource` 包含 0 计数 PASS）
- [ ] 4.3 实现 `backend/internal/profile/service/profile_reader.go` `GetCandidateProfileForUser(userId) -> (*CandidateProfile, error)`：spec D-13 read-only / 不触发 seed 副作用；store 层使用纯 SELECT 的 `GetByUser`，缺失返回 `(nil, nil)`；调用不写 audit_events / 不 bump profile_version（验证：unit test `TestGetCandidateProfileForUserSeededAndNil` 含"未 seed 返回 nil 且 DB 0 行" + "调用后 getMyProfile 仍可正常 seed" 双断言 PASS）
- [ ] 4.4 `cmd/api` route wiring：新增 `buildProfileRuntime` 组合 profile store / user_settings reader / idempotency middleware；挂载 5 个 route + session middleware + IK middleware（仅 experience card CUD）（验证：`cd backend && go test ./cmd/api -run TestBuildProfileRuntime -count=1` PASS）
- [ ] 4.5 `cmd/api` HTTP scenario：通过真实 route 验证 5 个 endpoint、auth 401、cross-user 404、IK replay、不重复创建 experience_cards（验证：`cd backend && go test ./cmd/api -run TestProfileHTTPScenario -count=1` PASS）
- [ ] 4.6 字节比对 [B2 fixtures](../../../openapi-v1-contract/spec.md) `Profile/getMyProfile.json` / `updateMyProfile.json` / `listExperienceCards.json` / `createExperienceCard.json` / `updateExperienceCard.json` default scenario（验证：fixture parity test PASS）

## Phase 5: 收口 + BDD + cross-owner handoff

- [ ] 5.1 跑 `cd backend && go test ./...` + `cd backend && go test ./internal/profile/...` + `cd backend && go test ./cmd/api -run 'TestBuildProfileRuntime|TestProfileHTTPScenario' -count=1` 全 PASS（验证：exit 0）
- [ ] 5.2 mock-first 对齐：5 个 endpoint 通过 `cmd/api` 真实 route 响应与对应 fixture default scenario 字节比对 PASS
- [ ] 5.3 grep `mistake|growth|drill|experiences|star` in `backend/internal/profile/` + outbox payload：0 命中（C-14 negative）（验证：`git grep` 输出）
- [ ] 5.4 BDD-Gate: E2E.P0.081 candidate-profile-seed-and-patch PASS（详见 [bdd-checklist.md](./bdd-checklist.md)）
- [ ] 5.5 BDD-Gate: E2E.P0.082 experience-cards-crud-with-ik PASS
- [ ] 5.6 BDD-Gate: E2E.P0.083 profile-privacy-delete-lifecycle PASS（含 audit tombstone 验证 + 无敏感字段泄漏）
- [ ] 5.7 在 `test/scenarios/e2e/INDEX.md` 追加 P0.081 + P0.082 + P0.083 行（关联需求 `backend-profile C-1..C-15`，含 C-15 cross-owner candidate profile internal API）
- [ ] 5.8 校准 `docs/spec/INDEX.md` §5 P0 Implementation 表 `backend-profile` 行：spec 创建期已写入 1.0 active 2026-05-21；本步骤只在 plan 完成时把行更新日期同步到 plan 落地日（如有变动）；如本 plan-review 已把 spec bump 至 1.1，需同步 INDEX 行版本字段；不新增行（验证：`sync-doc-index --check` PASS）
- [ ] 5.9 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `backend-profile` 状态描述从 "active（001 candidate profile + experience cards baseline）" 调整为 "active（001 candidate profile + experience cards baseline completed）"（或在启动 002 时改写 002 descriptor）；roadmap 版本号 bump 3.17 → 3.18；更新日期改为 plan 落地日；同步追加 engineering-roadmap history.md 3.18 行（验证：`sync-doc-index --check` PASS）
- [ ] 5.10 cross-owner handoff 信号：通知 [backend-jobs-recommendations/001](../../../backend-jobs-recommendations/plans/001-jd-match-real-backend-baseline/plan.md) owner（`GetCandidateProfileForUser` + `CountExperienceCardsBySource` 双 internal API 可用）+ backend internal privacy runner owner（`DeleteCandidateProfileForUser` 可用）+ 未来 `frontend-profile-and-settings` owner（5 个 Profile endpoint real backend 已就位）
- [ ] 5.11 收尾 spec/history 同步：plan 与 checklist 全部勾选后追加 backend-profile history.md "plan 001 完成" 行；把 backend-profile/plans/INDEX.md 001 行从"进行中（Active）"表移动到"已完成（Completed）"表并写完成日期；plan 落地最后一刻把 plan/checklist/bdd-plan/bdd-checklist 状态从 active → completed，并由 `/sync-doc-index --check` 校验 plans/INDEX 投影一致
