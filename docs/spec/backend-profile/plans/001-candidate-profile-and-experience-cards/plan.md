# Backend Profile Candidate Profile and Experience Cards Baseline

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-21

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [backend-profile spec](../../spec.md) §6 C-1..C-15 全部落到 backend Go handler + store + cmd/api wiring + cross-owner contract additive + privacy delete internal API + source counts internal API：

- 实现 `GET /api/v1/profiles/me` (getMyProfile) handler，首次访问 seed Lite profile 并以 `user_settings` 默认值填充（D-1）；后续访问返回同一行，不重复 seed；
- 实现 `PATCH /api/v1/profiles/me` (updateMyProfile) handler，patch 语义 + `profile_version` 自增 + cross-user 隔离（D-2 / D-12）；
- 实现 `GET /api/v1/profiles/me/experience-cards` (listExperienceCards) handler，cursor pagination + `updated_at DESC, id DESC` 唯一稳定序 + cross-user 隔离（D-8）；
- 实现 `POST /api/v1/profiles/me/experience-cards` (createExperienceCard) handler，IK + `source_type='manual'` 强制覆盖 + `confidence='medium'` 默认 + IK replay（D-5 / D-6 / D-7）；
- 实现 `PATCH /api/v1/profiles/me/experience-cards/{cardId}` (updateExperienceCard) handler，IK + patch 语义 + cross-user 404；
- 实现 `candidate_profiles` 与 `experience_cards` store Repository（CRUD + cross-user + cursor pagination + privacy delete + source counts）；
- 在 `cmd/api` 挂载 5 个 route + session middleware + IK middleware（仅 experience card CUD），并验证真实 HTTP runtime；
- 携带 [B2 cross-owner additive](../../../openapi-v1-contract/spec.md) 修订（在 `createExperienceCard` / `updateExperienceCard` 添加 `IdempotencyKey` parameter $ref + fixtures 增补 IK header 示例 + inventory 重算 + spec D-X 锁定），与本 plan Phase 1 同步落地；
- 暴露 `DeleteCandidateProfileForUser(userId)` internal API + `CountExperienceCardsBySource(userId)` internal API + `GetCandidateProfileForUser(userId)` internal API（[spec D-13](../../spec.md#31-已锁定决策)，read-only / 不触发 seed 副作用 / 缺失返回 nil），供 backend internal privacy runner + [backend-jobs-recommendations](../../../backend-jobs-recommendations/spec.md) `getJobMatchProfile` aggregation 消费；
- 通过 spec §6 C-1..C-15 验收 + 新增 E2E.P0.091 + E2E.P0.092 + E2E.P0.093 三个 BDD 场景；
- 不实现 AI 调用 / Insight Cards / 修正覆盖层 / 独立经历库 UI / cross-owner `AppendExperienceCardEvidence` write path（归 [spec §2.2](../../spec.md#22-out-of-scope) 与 plan 002 P1）。

## 2 背景

[engineering-roadmap §5.2](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) "App shell + auth + settings" workstream + [§6.3 S2](../../../engineering-roadmap/spec.md#63-s2--backend-domain-implementation) 标记 `backend-profile` 为身份 / 文件 / 画像基础三件套之一，是 [backend-jobs-recommendations](../../../backend-jobs-recommendations/spec.md) `getJobMatchProfile` aggregation 的画像证据底座；当前 5 个 Profile endpoint 在 [B2 §3.1.1](../../../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) 已 freeze、fixture 已就绪、但缺少真实 backend handler + cmd/api wiring。本 plan 是 backend-profile 第一批 plan，承担 P0 "已登录用户看到 / 修订 Lite candidate profile + 手动维护经历证据 + 隐私删除" 的 backend 端到端。

`candidate_profiles` 与 `experience_cards` 表已经在 [B4 baseline migration](../../../db-migrations-baseline/spec.md) 中创建，本 plan 不引入新 migration；只承接 store / handler / cmd/api wiring + cross-owner B2 IK additive。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就有 candidate profile read/write handler + cmd/api route；Phase 2 起来就有 experience cards CRUD handler + IK + cross-owner B2 additive；Phase 3 起来就有 store layer + cursor pagination + cross-user 隔离；Phase 4 起来就有 privacy delete + source counts internal API + cmd/api runtime wiring；Phase 5 收口 + BDD + 解锁 backend-jobs-recommendations / backend-internal-privacy-runner 后续 plan。

执行本 plan 前必须确认：

- [B2 §3.1.1](../../../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) Profile tag 5 个 operationId 已 freeze；fixture `openapi/fixtures/Profile/*.json` 5 个文件已存在并通过 `make validate-fixtures`。
- [B4 baseline migration](../../../db-migrations-baseline/spec.md) `candidate_profiles` / `experience_cards` 表已就位（baseline 000001）；本 plan 不修改 B4 schema。
- [backend-auth/001](../../../backend-auth/plans/001-passwordless-session-bootstrap/plan.md) completed：session middleware / current-user resolver / `user_settings` 表 seed 已可用（getMyProfile seed 依赖 user_settings 默认值）。
- [B1 D-5 error codes](../../../shared-conventions-codified/spec.md#31-已锁定决策) 提供 `VALIDATION_FAILED` / `RESOURCE_NOT_FOUND` / IK conflict envelope；本 plan 不私造错误码。
- [backend-runtime-topology](../../../backend-runtime-topology/spec.md) `cmd/api` in-process composition 模式已就绪（参考 backend-resume / backend-targetjob route wiring 同款）。

## 3 质量门禁分类

- **Plan 类型**: `code-internal + feature-behavior + contract`。本 plan 实现 backend handler / store / cmd/api wiring；用户可见 HTTP API 行为；含 B2 cross-owner additive 修订（contract）。
- **TDD 策略**: 适用。Red-Green-Refactor 入口：
  1. handler unit test：5 个 endpoint 各自参数校验 + IK + 422 / 404 / 跨用户隔离 + source_type 强制覆盖；
  2. store integration test：candidate_profiles seed / patch / version bump；experience_cards CRUD + cursor pagination + cross-user 隔离 + source counts；
  3. cross-owner B2 additive test：generated client / fixture / inventory 同步 PASS（`make codegen-check` / `make validate-fixtures` / `make lint-openapi` / `make openapi-diff`）；
  4. privacy delete internal API test：experience_cards → candidate_profiles 删除顺序 + audit tombstone 完整 + 无敏感字段泄漏；
  5. source counts internal API test：4 个 source_type 计数返回正确；
  6. candidate profile read internal API test（D-13）：已 seed 用户返回 `*CandidateProfile` 与 `getMyProfile` 字段集一致；未 seed 用户返回 nil（不写 audit_events / 不 bump profile_version / DB `candidate_profiles` 仍 0 行）；后续 `getMyProfile` 仍能按 D-1 seed；
  7. `cmd/api` route/runtime test：session middleware、IK middleware（experience CUD only）、route path params、5 个 route 真实可达 + cross-user 404 + IK replay。
  执行入口：`/implement backend-profile/001-candidate-profile-and-experience-cards` → `/tdd`。
- **BDD 策略**: 适用（Feature plan requires BDD）。E2E.P0.091 candidate-profile-seed-and-patch + E2E.P0.092 experience-cards-crud-with-ik + E2E.P0.093 profile-privacy-delete-lifecycle。详见 [bdd-plan.md](./bdd-plan.md) / [bdd-checklist.md](./bdd-checklist.md)。
- **替代验证 gate**:
  - `cd backend && go test ./...`
  - `cd backend && go test ./internal/profile/handler/... -count=1`
  - `cd backend && go test ./internal/profile/store/... -tags=integration -count=1`
  - `cd backend && go test ./cmd/api -run 'TestBuildProfileRuntime|TestProfileHTTPScenario' -count=1`
  - smoke：`curl -X GET /api/v1/profiles/me` 与 mock-server fixture 字节比对
  - `make codegen-check`（generated Go/TS artifacts + OpenAPI drift PASS）
  - `make validate-fixtures`（B2 additive fixture 同步 PASS）
  - `make lint-openapi`（IK parameter inventory 同步 PASS）
  - `make openapi-diff`（additive only，PASS）
  - grep `mistake|growth|drill|experiences|star` in `backend/internal/profile/` + outbox payload：0 命中（C-14 negative）
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `make docs-check`

### 3.1 Frontend / Backend Operation Matrix

| operationId | fixture | frontend consumer | backend handler | persistence | AI dependency | scenario coverage |
|-------------|---------|-------------------|-----------------|-------------|---------------|-------------------|
| `getMyProfile` | `openapi/fixtures/Profile/getMyProfile.json` `default` | 未来 `frontend-profile-and-settings` subspec；当前未被 frontend 消费（contract 就绪等待 UI 接入） | `backend/internal/profile/handler/get.go` real handler + `cmd/api` `GET /api/v1/profiles/me` route with session middleware | `candidate_profiles` (UPSERT by user_id) + 读取 `user_settings` 作为 seed 默认值 | none | E2E.P0.091 |
| `updateMyProfile` | `openapi/fixtures/Profile/updateMyProfile.json` `default` | 未来 `frontend-profile-and-settings` subspec；当前未被 frontend 消费 | `backend/internal/profile/handler/update.go` real handler + `cmd/api` `PATCH /api/v1/profiles/me` route with session middleware（无 IK，patch 幂等） | `candidate_profiles` UPDATE supplied fields + `profile_version += 1` | none | E2E.P0.091 |
| `listExperienceCards` | `openapi/fixtures/Profile/listExperienceCards.json` `default` | 未来 `frontend-profile-and-settings` subspec；当前未被 frontend 消费 | `backend/internal/profile/handler/list_cards.go` real handler + `cmd/api` `GET /api/v1/profiles/me/experience-cards` route with session middleware | `experience_cards` cursor pagination | none | E2E.P0.092 |
| `createExperienceCard` | `openapi/fixtures/Profile/createExperienceCard.json` `default` + new IK fixture variant via B2 additive | 未来 `frontend-profile-and-settings` subspec；当前未被 frontend 消费 | `backend/internal/profile/handler/create_card.go` real handler + `cmd/api` `POST /api/v1/profiles/me/experience-cards` route with session + IK middleware | `experience_cards` INSERT with `source_type='manual'` forced + `confidence='medium'` default | none | E2E.P0.092 |
| `updateExperienceCard` | `openapi/fixtures/Profile/updateExperienceCard.json` `default` + new IK fixture variant via B2 additive | 未来 `frontend-profile-and-settings` subspec；当前未被 frontend 消费 | `backend/internal/profile/handler/update_card.go` real handler + `cmd/api` `PATCH /api/v1/profiles/me/experience-cards/{cardId}` route with session + IK middleware + cross-user 404 | `experience_cards` UPDATE supplied fields | none | E2E.P0.092 |

### 3.2 B2 Cross-owner Additive

Phase 1 必须同时携带 [openapi-v1-contract](../../../openapi-v1-contract/spec.md) cross-owner additive 修订：

- `openapi/openapi.yaml`：在 `createExperienceCard` 与 `updateExperienceCard` operation 的 `parameters` 列表中追加 `- $ref: '#/components/parameters/IdempotencyKey'`（与 [B2 §3.1](../../../openapi-v1-contract/spec.md) 既有 side-effect IK 惯例一致）。
- `openapi/fixtures/Profile/createExperienceCard.json` / `updateExperienceCard.json`：在 default scenario 的 request 中追加 `headers.Idempotency-Key` 示例（或新增 `with-idempotency-key` scenario，保持 default 字节兼容）。
- B2 spec.md：新增 D-X "Profile experience card CUD adopt IK" 决策行，引用本 plan 与 backend-profile spec D-5；同步追加 B2 history.md 行（与已落地 D-18 / D-23 模式一致），并在 B2 `plans/INDEX.md` 行更新最近更新日期，保 `sync-doc-index --check` PASS。
- B2 fixture validator (`make validate-fixtures`) / inventory lint (`make lint-openapi`) / breaking-change gate (`make openapi-diff`) / generated server+client artifacts (`make codegen-check`) 全部 PASS。
- B2 endpoint 总数不变（仍 59 op），只是 2 个 op 新增 IK parameter（additive change，非破坏性）。

## 4 实施步骤

### Phase 1: candidate profile handler + cmd/api wiring + B2 IK additive 准备

#### 1.1 实现 `backend/internal/profile/handler/get.go`
- 实现 generated server interface `GetMyProfile`
- 调用 store `GetByUserOrSeed(userId)`：若 `candidate_profiles` 无对应行，按 `user_settings` 默认值 seed 一行（headline=null / yearsOfExperience=null / currentRole=null / preferredPracticeLanguage=user_settings.preferred_practice_language / uiLanguage=user_settings.ui_language / region=user_settings.region）；返回 `CandidateProfile` 响应

#### 1.2 实现 `backend/internal/profile/handler/update.go`
- 实现 generated server interface `UpdateMyProfile`
- patch 语义：只更新 supplied fields；`yearsOfExperience` minimum 0 校验（违反返回 422 + VALIDATION_FAILED）；空字符串视为合法值（清空）
- 调用 store `UpsertLite(userId, patch)`：UPDATE + `profile_version += 1` + `updated_at = now()`；store 内部用 `SELECT FOR UPDATE` 防并发 race
- 返回更新后的 `CandidateProfile`

#### 1.3 B2 cross-owner additive 修订（与 Phase 1 同步）
- 修订 `openapi/openapi.yaml` 在 `createExperienceCard` / `updateExperienceCard` operation 添加 `IdempotencyKey` parameter $ref
- 修订对应 fixture 增补 IK header 示例（保持 default scenario 字节兼容）
- 修订 [openapi-v1-contract spec.md](../../../openapi-v1-contract/spec.md) 新增 D-X 决策行
- 修订 [openapi-v1-contract history.md](../../../openapi-v1-contract/history.md) 追加 backend-profile/001 cross-owner additive 行
- 重新生成 OpenAPI client/server artifacts；运行 `make codegen-check` + `make validate-fixtures` + `make lint-openapi` + `make openapi-diff` PASS

#### 1.4 unit test
- `get_test.go`: seed 路径（首次访问） / get 路径（已存在） / cross-user 隔离（无）
- `update_test.go`: patch 单字段 / patch 多字段 / yearsOfExperience -1 返回 422 / 空字符串清空字段 / profile_version 自增

### Phase 2: experience cards CRUD handler + IK + cross-user 隔离

#### 2.1 实现 `backend/internal/profile/handler/list_cards.go`
- 实现 generated server interface `ListExperienceCards`
- cursor pagination（`updated_at DESC, id DESC`），cursor invalid 返回 422 + VALIDATION_FAILED
- cross-user 过滤（仅返回 `user_id = current_user_id`）
- 返回 `PaginatedExperienceCard{items, pageInfo{nextCursor, pageSize, hasMore}}`

#### 2.2 实现 `backend/internal/profile/handler/create_card.go`
- 实现 generated server interface `CreateExperienceCard`
- IK 校验（缺失 / 24h replay 返回首次 card / mismatch fingerprint 走 B1 generic IK conflict 409）
- 强制 `source_type='manual'`（即使 body 中携带其他 source_type 也覆盖）
- 默认 `confidence='medium'`
- 必填字段校验（title / companyName / situation / task / action / result / language），违反返回 422 + VALIDATION_FAILED
- 关联 `profile_id` = 用户的 `candidate_profiles.id`（若 profile 不存在则按 Phase 1 seed 逻辑创建）

#### 2.3 实现 `backend/internal/profile/handler/update_card.go`
- 实现 generated server interface `UpdateExperienceCard`
- IK 校验（同 2.2）
- cross-user 404 + RESOURCE_NOT_FOUND（cardId 属于其他用户时不暴露存在）
- patch 语义：只更新 supplied fields
- 返回更新后的 `ExperienceCard`

#### 2.4 unit test
- `list_cards_test.go`: 空列表 / 25 行 + cursor 第二页 / cross-user 不可见 / cursor invalid 返回 422
- `create_card_test.go`: 成功 / IK replay / IK conflict 409 / source_type 强制覆盖 / 必填字段校验 / confidence 默认 medium
- `update_card_test.go`: 成功 patch / cross-user 404 / IK replay / 不存在 cardId 返回 404

### Phase 3: store layer + cursor pagination + cross-user 隔离

#### 3.1 实现 `backend/internal/profile/store/profile.go`
- `GetByUserOrSeed(userId)`：先 SELECT；不存在则按 user_settings 默认值 INSERT + RETURNING；UNIQUE constraint 已在 B4 baseline 保证
- `UpsertLite(userId, patch)`：UPDATE + `profile_version += 1` + `updated_at = now()`；SELECT FOR UPDATE 防 race
- `DeleteForUser(userId)`：DELETE FROM candidate_profiles WHERE user_id = ...（FK CASCADE 删除 experience_cards 由 B4 baseline 保证，但 Phase 4 handler 层显式先删 experience_cards 保 audit 完整）

#### 3.2 实现 `backend/internal/profile/store/experience_cards.go`
- `ListByUser(userId, cursor, pageSize)`：cursor pagination，`updated_at DESC, id DESC` 唯一稳定序；返回 items + nextCursor + hasMore
- `Create(userId, profileId, attrs, source)`：INSERT；attrs 包含 title / company / situation / task / action / result / metrics / skills / language；source 包含 source_type / source_ref_id / confidence
- `Update(cardId, userId, patch)`：UPDATE WHERE id = cardId AND user_id = userId；返回 affected rows，0 行视为 cross-user 404
- `DeleteForUser(userId)`：DELETE FROM experience_cards WHERE user_id = ...；返回 affected count（用于 audit tombstone）
- `CountBySource(userId)`：SELECT source_type, COUNT(*) GROUP BY source_type；返回 `{manual, resume_parse, practice_report, debrief}`

#### 3.3 integration test
- `profile_integration_test.go`: seed / patch / version bump / delete / cross-user
- `experience_cards_integration_test.go`: list cursor pagination 边界（empty / single page / 25 行 + 第二页 / cross-user 不可见）+ create / update / delete / count by source

### Phase 4: privacy delete internal API + source counts internal API + cmd/api runtime wiring

#### 4.1 实现 `backend/internal/profile/service/privacy.go`
- `DeleteCandidateProfileForUser(userId)`：
  1. 调 store `CountBySource(userId)` 获取 experienceCardCount
  2. 调 store `experience_cards.DeleteForUser(userId)`
  3. 调 store `candidate_profiles.DeleteForUser(userId)`
  4. 写入 audit_events 表（tombstone：userId / experienceCardCount / 删除时间 / job_id；不含 title / situation / task / action / result / metrics / skills / headline / currentRole）
  5. 错误处理：任一步骤失败回滚事务并写 audit failure 记录

#### 4.2 实现 `backend/internal/profile/service/source_counts.go`
- `CountExperienceCardsBySource(userId)`：直接代理 store `CountBySource`；返回 `map[source_type]count`
- 为 [backend-jobs-recommendations](../../../backend-jobs-recommendations/spec.md) `getJobMatchProfile` aggregation 留出消费入口

#### 4.3 实现 `backend/internal/profile/service/profile_reader.go`（spec D-13）
- `GetCandidateProfileForUser(userId) -> (*CandidateProfile, error)`：调 store `GetByUser(userId)` 纯 SELECT 路径（**不复用 `GetByUserOrSeed`**，避免触发 D-1 seed 副作用）；缺失返回 `(nil, nil)`；caller 决定 fallback
- 返回字段集与 generated `CandidateProfile` schema 一致（headline / yearsOfExperience / currentRole / preferredPracticeLanguage / uiLanguage / region）；不返回 profile_version / created_at / updated_at（不在 B2 schema 内）
- 调用路径不写 audit_events / 不 bump profile_version / 不修改 DB；read-only
- 对外仅由 [backend-jobs-recommendations](../../../backend-jobs-recommendations/spec.md) `getJobMatchProfile` aggregation 调用；不暴露 HTTP；cross-user 隔离由 caller 提供合法 `userId` 保证

#### 4.4 `cmd/api` route wiring
- 新增 `buildProfileRuntime`（或等价 composition helper），组合 profile store / user_settings reader / idempotency middleware
- 挂载：
  - `GET /api/v1/profiles/me` → session middleware + `GetMyProfile`
  - `PATCH /api/v1/profiles/me` → session middleware + `UpdateMyProfile`（无 IK middleware）
  - `GET /api/v1/profiles/me/experience-cards` → session middleware + `ListExperienceCards`
  - `POST /api/v1/profiles/me/experience-cards` → session middleware + IK middleware + `CreateExperienceCard`
  - `PATCH /api/v1/profiles/me/experience-cards/{cardId}` → session middleware + IK middleware + path param adapter + `UpdateExperienceCard`
- `cmd/api` tests 断言 route 存在、缺 session 返回 auth error、缺 IK（create/update card）返回 generated error envelope、同 IK replay 不重复创建 `experience_cards`

#### 4.5 unit test
- `privacy_delete_test.go`: 完整删除链路 + audit tombstone 完整 + 无敏感字段泄漏
- `source_counts_test.go`: 4 个 source_type 计数返回正确（包括 0 计数）
- `profile_reader_test.go`: 已 seed 用户返回字段一致；未 seed 用户返回 nil（不写 audit / 不 bump version / DB 0 行）；调用后再走 `getMyProfile` 仍可正常 seed（无抢占副作用）

### Phase 5: 收口 + BDD + cross-owner handoff

#### 5.1 跨 gate 收口

按 §3 替代验证 gate 依序运行：
- `cd backend && go test ./...` PASS
- `cd backend && go test ./internal/profile/...` PASS
- `cd backend && go test ./cmd/api -run 'TestBuildProfileRuntime|TestProfileHTTPScenario' -count=1` PASS
- mock-first 对齐：5 个 endpoint 通过真实 route 响应与 [B2 fixtures](../../../openapi-v1-contract/spec.md) Profile/*.json default scenario 字节比对 PASS
- `make codegen-check` PASS（含 generated Go/TS artifacts drift）
- `make validate-fixtures` PASS（含 IK additive fixture）
- `make lint-openapi` PASS
- `make openapi-diff` PASS（additive only）
- grep `mistake|growth|drill|experiences|star` in `backend/internal/profile/` + outbox payload：0 命中（C-14 negative）
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` PASS
- `make docs-check` PASS

#### 5.2 BDD 场景验证

- 执行 `test/scenarios/e2e/p0-091-candidate-profile-seed-and-patch/` 全 PASS
- 执行 `test/scenarios/e2e/p0-092-experience-cards-crud-with-ik/` 全 PASS
- 执行 `test/scenarios/e2e/p0-093-profile-privacy-delete-lifecycle/` 全 PASS
- 在 `test/scenarios/e2e/INDEX.md` 追加 P0.091 / P0.092 / P0.093 行

#### 5.3 cross-owner handoff 信号

通知下游 owner：
- [backend-jobs-recommendations/001-jd-match-real-backend-baseline](../../../backend-jobs-recommendations/plans/001-jd-match-real-backend-baseline/plan.md) owner：`GetCandidateProfileForUser(userId)`（D-13，read-only / 不 seed）+ `CountExperienceCardsBySource(userId)`（D-11）2 个 cross-owner internal API 已可用；前者用于 `BuildJobMatchProfile` 的 headline / yearsOfExperience，后者作为内部质量信号与 P1 `experienceCards` additive 扩展锚点，不直接写入当前 `JobMatchProfileSourceCounts`
- backend internal privacy runner（[backend-runtime-topology](../../../backend-runtime-topology/spec.md)）owner：`DeleteCandidateProfileForUser(userId)` internal API 已可用，可纳入 privacy_delete job dispatcher 链路
- 未来 `frontend-profile-and-settings` subspec owner：5 个 Profile endpoint real backend 已就位，可启动 UI 集成（contract 已通过 generated client + fixture 字节级对齐）

本 plan 不直接修订下游 owner 文件，只在 5.3 完成 "可消费" 信号传递。

#### 5.4 spec / history / INDEX 同步

- backend-profile spec.md 本次维持 1.1 active（L1 plan-review 已在 spec 创建当天新增 D-13；plan 完成时若无新决策则保持 1.1，否则同步 bump）
- backend-profile history.md 在 plan 完成当天追加一行"plan 001 完成 + cross-owner GetCandidateProfileForUser / CountExperienceCardsBySource / DeleteCandidateProfileForUser internal API 全可用"
- `docs/spec/INDEX.md` §5 P0 Implementation 表中 `backend-profile` 行已在 spec 创建期写入（1.0 active 2026-05-21）；plan 完成时只需把 INDEX 行更新日期同步到 plan 完成日（如需），不再新增行
- [engineering-roadmap §5.2](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 行已在 spec 创建期改为 `backend-profile active（001 candidate profile + experience cards baseline）`，roadmap 已 bump 至 3.17；plan 完成时把状态描述从 "active（001 ... baseline）" 调整为 "active（001 ... baseline completed）"（或同时启动 002 时改写 002 descriptor），roadmap 版本号 bump 3.17 → 3.18，更新日期改为 plan 落地日；roadmap history.md 同步追加 3.18 行

## 5 验收标准

- 本计划列出的 §4 所有 Phase task 全部完成
- §3 替代验证 gate 全部通过
- spec §6 C-1..C-15 全部 PASS（含 C-15 cross-owner candidate profile internal API 验证）
- `cmd/api` route/runtime gate PASS：session middleware、IK middleware（experience CUD only）、5 个 route 真实可达、cross-user 404、IK replay 均有测试证据
- B2 cross-owner additive 修订 PASS：`createExperienceCard` / `updateExperienceCard` 已含 IK parameter；fixture / generated artifacts / inventory / breaking-change gate / B2 spec.md + history.md + plans/INDEX.md 同步 PASS
- BDD E2E.P0.091 + E2E.P0.092 + E2E.P0.093 全 PASS（E2E.P0.092 含 `GetCandidateProfileForUser` cross-owner internal API 验证）
- 下游 owner 已收到 `GetCandidateProfileForUser` + `CountExperienceCardsBySource` + `DeleteCandidateProfileForUser` 三个 internal API 可用信号
- `docs/spec/INDEX.md` + engineering-roadmap §5.2 + engineering-roadmap history.md 已同步本 subject `plan 001 完成` 状态描述 + roadmap 3.17 → 3.18 bump

## 6 风险与应对

| 风险 | 应对 |
|------|------|
| R1: B2 cross-owner IK additive 触发 breaking-change gate 误判 | additive parameter 只追加不删除；fixture 保留 default 字节兼容（IK header 作为可选示例）；gate 配置应允许 parameter additive 升级；若仍误判，先修订 B2 diff-config 允许 IK additive |
| R2: candidate profile seed 与 user_settings 默认值漂移 | Phase 1 seed 必须读 `user_settings` 当前值；如 user_settings 不存在则按 B4 baseline 默认值（zh-CN / en / null）；integration test 覆盖 seed 路径 |
| R3: cross-user 隔离漏洞导致越权 | handler + store 双层 `user_id` 过滤；integration test 强制覆盖 cross-user case；experience card update 必须 WHERE 同时匹配 cardId AND userId |
| R4: source_type taxonomy 被前端绕过 | handler 层强制覆盖 `source_type='manual'`（即使 body 携带其他值）；store 层 CHECK constraint 兜底；unit test 覆盖伪造 source_type |
| R5: privacy delete 顺序错误导致 FK 违反 | Phase 4 handler 层显式先删 experience_cards 再删 candidate_profiles；B4 baseline FK CASCADE 仅作兜底，不依赖 |
| R6: IK middleware 与无 IK PATCH 共存配置错误 | cmd/api 中 PATCH /profiles/me 路由显式不挂 IK middleware；其他 4 个 endpoint 中只 POST/PATCH experience card 挂 IK；route wiring test 验证 |
| R7: 与 backend-jobs-recommendations 并行开发导致 source counts API 形态错位 | 本 plan Phase 4.2 暴露的 internal API 签名 `CountExperienceCardsBySource(userId) -> map[source_type]count` 在 spec D-11 已锁定；jobs-recommendations 在 plan 设计时引用同一签名，发现不一致先回到 backend-profile spec 修订 |
| R8: backend-profile 与 future frontend-profile-and-settings 并行可能出现 contract 漂移 | 本 plan 严格按 B2 freeze 的 5 个 op 实现；任何字段扩展必须先 B2 additive；frontend 切真时引用同一 generated client + fixture |
| R9: `GetCandidateProfileForUser` 误调用 seed 路径导致 backend-jobs-recommendations aggregation 产生 backend-profile 副作用 | Phase 4.3 实现严格使用纯 SELECT 的 store `GetByUser`（**不复用** `GetByUserOrSeed`）；unit test `profile_reader_test.go` 必须包含 "调用后 DB 未创建新 candidate_profile 行" + "调用后 `getMyProfile` 仍可正常 seed" 双断言；spec D-13 与 §4.4 cross-owner 约束已显式锁定 read-only 语义 |
