# Backend Profile Spec

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-21

## 1 背景与目标

[engineering-roadmap §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选) 与 [engineering-roadmap §6.3 S2](../engineering-roadmap/spec.md#63-s2--backend-domain-implementation) 把 `backend-profile` 列为 P0 实施 backend 域中"身份、文件与画像基础"中的画像 owner，但在本 spec 创建之前 subject 尚未真正派生，导致 5 个 Profile tag operationId（`getMyProfile` / `updateMyProfile` / `listExperienceCards` / `createExperienceCard` / `updateExperienceCard`）在 [B2 §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) 已 freeze 但缺少真实 handler、store 与 cmd/api wiring。

本 subject 是 Candidate Profile 业务域的后端 owner：

1. **承载 candidate_profiles + experience_cards 两张表的 store 与 handler**。两张表在 [B4 baseline migration](../db-migrations-baseline/spec.md) 已存在；本 subject 不引入新 migration，只承接 store / handler / cmd/api wiring。
2. **实现 5 个 Profile endpoint 的真实业务逻辑**：candidate profile Lite 读写（M1-lite Progressive Profile，按 [product-scope §6.6](../product-scope/spec.md#65-m1-lite渐进式画像) 的最小可用画像形态）+ experience cards CRUD（作为画像证据原子层，按 [product-scope §5.1 M1](../product-scope/spec.md#51-产品能力层) Progressive Profile 的"经历证据"承载，**不恢复独立经历库 UI 入口**）。
3. **为 backend-jobs-recommendations 提供画像证据聚合源**：candidate_profiles 提供 `JobMatchProfile` headline / yearsOfExperience；experience_cards 当前作为内部质量信号与 P1 `experienceCards` 扩展锚点；当前 `sources` response 计数仍由简历 / JD / 模拟面试 / 复盘 owner 提供；本 subject 是这份聚合的"画像底座"。
4. **隐私删除链路 owner**：按 [B4 D-11 privacy deletion matrix](../db-migrations-baseline/spec.md) 与 [product-scope §9.3](../product-scope/spec.md#93-数据与隐私) 提供 `DeleteCandidateProfileForUser(userId)` internal API，hard delete `candidate_profiles` + `experience_cards`，并通过 audit tombstone 记录删除时间与 ID。
5. **mock-first 切真**：本 subject 实现的 handler 响应与 [B2 fixtures](../openapi-v1-contract/spec.md) `Profile/*.json` 字节比对，[mock-contract-suite C-9](../mock-contract-suite/spec.md#6-验收标准) 强制 enforce。

本 subject **不实现**：前端 UI（用户画像页归 [docs/ui-design/user-profile-and-settings.md](../../ui-design/user-profile-and-settings.md) + 未来 `frontend-profile-and-settings` subspec；当前 UI placeholder shell 暂由 [frontend-shell](../frontend-shell/spec.md) 承接）；岗位推荐 / 全球搜岗（归 [backend-jobs-recommendations](../backend-jobs-recommendations/spec.md)）；简历资产（归 [backend-resume](../backend-resume/spec.md)）；复盘（归 [backend-debrief](../backend-debrief/spec.md)）。

## 2 范围

### 2.1 In Scope

- **HTTP handler + runtime wiring**：实现 [B2 §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) Profile tag 5 个 operationId（`getMyProfile` / `updateMyProfile` / `listExperienceCards` / `createExperienceCard` / `updateExperienceCard`），在 `cmd/api` 按当前 session middleware / generated response envelope 口径挂载真实 route；experience card CUD 必带 `Idempotency-Key`（本 spec D-5 锁定，需 B2 cross-owner additive）。
- **store layer**：
  - `candidate_profiles` Repository：`GetByUser(userId)` / `UpsertLite(userId, fields...)` / `BumpProfileVersion(userId)` / `DeleteForUser(userId)`；`UpsertLite` 在 user 首次访问 `getMyProfile` 时按 [B4 baseline `user_settings`](../db-migrations-baseline/spec.md) 默认值 seed。
  - `experience_cards` Repository：`ListByUser(userId, cursor, pageSize)` / `Create(userId, profileId, attrs, source)` / `Update(cardId, userId, patch)` / `DeleteForUser(userId)`；cursor pagination 按 `updated_at DESC, id DESC` 唯一稳定序。
- **source taxonomy enforcement**：`experience_cards.source_type` ∈ `manual | resume_parse | practice_report | debrief`（[B4 enum source](../db-migrations-baseline/spec.md)）；外部 owner（resume.parse / report.generate / debrief.generate）通过 internal API `AppendExperienceCardEvidence(userId, profileId, source_type, source_ref_id, payload)` 写入，**本 subject 不在 P0 实现该 internal API**，但 store 与 schema 必须为后续 cross-owner write path 留出可调用入口（D-6）。
- **隐私删除接口**：暴露 `DeleteCandidateProfileForUser(userId)` internal API，由后续 backend internal privacy runner 调用；删除顺序 `experience_cards → candidate_profiles`；audit tombstone 写 user_id / 删除时间，不含原文字段。
- **mock-first 对齐**：本 subject 实现的 handler 响应字段集 / status code / 字段顺序与 [B2 fixtures](../openapi-v1-contract/spec.md) `Profile/getMyProfile.json` / `updateMyProfile.json` / `listExperienceCards.json` / `createExperienceCard.json` / `updateExperienceCard.json` 字节比对。
- **cross-user 隔离**：所有 endpoint 必须以 `user_id = current_user_id` 过滤；cross-user 访问返回 404 / `RESOURCE_NOT_FOUND`（不暴露存在）。
- **画像证据 source counts 暴露**：experience_cards `source_type` 计数能力（`CountExperienceCardsBySource(userId)` internal API），供 [backend-jobs-recommendations](../backend-jobs-recommendations/spec.md) `getJobMatchProfile` 聚合路径读取；当前 `JobMatchProfileSourceCounts` schema 仍只含 `resumes / jds / mocks / debriefs`，experience_cards 计数仅作为内部质量信号与 P1 `experienceCards` additive 扩展锚点，不直接写入当前 `sources` response。

### 2.2 Out of Scope

- **前端 UI**：用户画像页与设置页归 `docs/ui-design/user-profile-and-settings.md` + 未来 `frontend-profile-and-settings` subspec；本 subject 不实现 React 组件。
- **AI 画像生成（Insight Cards / Sections / Headline 推断）**：[user-profile-and-settings.md §3 §4](../../ui-design/user-profile-and-settings.md) 描述的"AI 当前推断 / 置信度 / 修正流"在 P0 不实现，归后续 `profile.insight_*` plan；本 subject 只承接 Lite identity（headline / yearsOfExperience / currentRole / preferredPracticeLanguage / uiLanguage / region）。
- **修正覆盖层**：[user-profile-and-settings.md §4](../../ui-design/user-profile-and-settings.md) 提到"用户修正不删除原始证据"的覆盖层机制在 P0 不实现，归后续 P1 plan；experience_cards `source_type='manual'` 是 P0 唯一允许的用户修正路径。
- **cross-owner evidence write path (AppendExperienceCardEvidence)**：resume.parse / report.generate / debrief.generate 自动写 experience_cards 的 internal API 在 P0 不实现；本 subject 只确保 schema / source_type enum / source_ref_id 字段为后续 cross-owner write 留出入口。
- **独立"经历库" UI 模块**：[product-scope §5.4](../product-scope/spec.md#54-移除或降级的能力) + [docs/ui-design/removed-modules-and-scope.md §5](../../ui-design/removed-modules-and-scope.md) 已锁定经历库 / STAR 编辑器为已丢弃 / 长期不恢复；本 subject 提供 experience cards CRUD endpoint 是为承接 B2 已 freeze 的契约 + 画像证据原子层，**不得**作为恢复独立经历库 UI 的依据。
- **完整 Progressive Profile sections / Insight Cards / 来源统计 UI 数据接口**：当前 OpenAPI Profile tag 已 freeze 在 5 个 endpoint；后续 UI 需要扩展时由 owner 修订 B2 + 本 spec。
- **岗位推荐相关 endpoint**：归 [backend-jobs-recommendations](../backend-jobs-recommendations/spec.md)；本 subject 只提供画像聚合证据源（internal API），不实现 JobMatch handler。
- **`/runtime-config` 用户偏好聚合**：归 [backend-auth](../backend-auth/spec.md) + [secrets-and-config](../secrets-and-config/spec.md)；本 subject `updateMyProfile` 更新的 `preferredPracticeLanguage` / `uiLanguage` 字段写入 `candidate_profiles`，不直接同步到 `user_settings`（user_settings 由 backend-auth + A4 owner 维护；如需双写由后续 cross-owner plan 处理）。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | candidate_profile 单例边界 | 每用户 1 行 `candidate_profiles`（`user_id` UNIQUE，[B4 baseline](../db-migrations-baseline/spec.md)）；首次 `getMyProfile` 调用以 `user_settings` 默认值 seed 行（headline=null / yearsOfExperience=null / currentRole=null / preferredPracticeLanguage=user_settings.preferred_practice_language / uiLanguage=user_settings.ui_language / region=user_settings.region） | `getMyProfile` 永远不返回 404（已登录用户必然有 profile）；新用户首次访问看到 empty Lite profile，不需要单独 onboarding step |
| D-2 | UpdateProfileRequest 字段语义 | 所有 6 个字段 optional；只有 supplied field 更新；空字符串视为合法值（清空）；`yearsOfExperience` minimum 0；language / region 字段不做白名单（由前端 i18n locale helper 决定合法集合）；`updated_at` + `profile_version` 自动 bump | 与 [B2 schema `UpdateProfileRequest`](../openapi-v1-contract/spec.md) "All fields optional — only supplied fields are updated" 一致；patch 语义清晰；前端可单独保存 headline 等单字段 |
| D-3 | candidate profile read 无 IK | `getMyProfile` / `listExperienceCards` 是纯 read，不需要 IK | 与 B2 schema 一致（read endpoint 无 `IdempotencyKey` parameter） |
| D-4 | candidate profile write IK | `updateMyProfile` 当前 [B2 OpenAPI](../openapi-v1-contract/spec.md) 未声明 IK；本 spec D-4 选择 P0 不强制 IK（patch 语义本身幂等：同字段 patch 两次结果相同；profile_version bump 由 store 层 advisory lock 保证单调） | 不为单字段 patch 引入不必要 IK 复杂度；如未来 PATCH 改为 PUT 全量替换语义，必须先修订 B2 加 IK |
| D-5 | experience card write IK | `createExperienceCard` / `updateExperienceCard` P0 需要 IK 防止网络抖动重复创建；当前 [B2 OpenAPI](../openapi-v1-contract/spec.md) 未声明 IK，需要本 subject 同步 cross-owner additive 修订 [openapi-v1-contract](../openapi-v1-contract/spec.md) 在 `createExperienceCard` / `updateExperienceCard` 上添加 `IdempotencyKey` parameter；fixture / generated client 同步增补 | side-effect operation 必带 IK 是 B2 §3.1 既有惯例；本 spec 与 B2 additive 修订同步落地，避免 frontend 在切真时遇到 contract 缺口 |
| D-6 | experience card source taxonomy enforcement | handler 层 `createExperienceCard` 强制 `source_type='manual'` 写入；非 manual 来源（resume_parse / practice_report / debrief）的 experience_cards 创建由后续 cross-owner `AppendExperienceCardEvidence(userId, profileId, source_type, source_ref_id, payload)` internal API 承接（P1，不在本 plan）；P0 store 层与 schema 必须保持 source_type enum 完整，为未来 cross-owner write 留出入口 | 防止前端通过 HTTP create endpoint 伪造 `source_type='resume_parse'` 等系统来源；保留 B4 schema 完整性；明确 manual 与系统来源的写入路径分离 |
| D-7 | confidence 默认值 | `experience_cards.confidence` P0 在 manual 创建路径默认 `medium`；用户可在 update 中改为 high / low；系统来源（resume_parse / practice_report / debrief）在未来 cross-owner write 中由源 owner 自行确定 | manual 用户输入证据置信度默认中性；高 / 低由用户自行评估 |
| D-8 | cross-user 隔离错误 | 所有 endpoint cross-user 访问返回 `404 + RESOURCE_NOT_FOUND`（不暴露 cardId / profileId 存在性）；audit_events 不写敏感字段 | 与 [backend-resume D-6](../backend-resume/spec.md#31-已锁定决策) cross-user 行为一致；防止枚举攻击 |
| D-9 | privacy delete 顺序 | `DeleteCandidateProfileForUser(userId)` 调用顺序：`experience_cards → candidate_profiles`（FK CASCADE 已存在 [B4 baseline](../db-migrations-baseline/spec.md)，handler 层显式删除以保证 audit 写入完整）；audit tombstone 写 userId / experienceCardCount / 删除时间，不含 title / situation / task / action / result / metrics / skills | 与 [backend-resume privacy](../backend-resume/spec.md#43-存储约束) 同款 hard delete；保证 audit 可观测但不泄漏内容 |
| D-10 | preferred_practice_language / ui_language 双写边界 | `updateMyProfile` 只写 `candidate_profiles.preferred_practice_language` / `ui_language`；不级联写 `user_settings`；如未来 UI 需要把画像偏好同步到 `user_settings`（影响 `/runtime-config` 返回），由 backend-auth + A4 owner 在 cross-owner plan 中实现 | 防止 backend-profile 越权写 backend-auth 拥有的表；保留扩展点；P0 user 可同时在画像页（写 candidate_profiles）与设置页（写 user_settings）维护语言偏好，UI owner 负责一致性提示 |
| D-11 | source counts internal API | 暴露 `CountExperienceCardsBySource(userId) -> {manual, resume_parse, practice_report, debrief}` internal API；P0 由 [backend-jobs-recommendations](../backend-jobs-recommendations/spec.md) `getJobMatchProfile` aggregation 读取，用于内部质量信号、trace 与 P1 `experienceCards` additive 扩展锚点；当前 response 的 `JobMatchProfileSourceCounts` 仍只含 `resumes / jds / mocks / debriefs`，由 backend-resume / backend-targetjob / backend-practice / backend-debrief 各自提供 | 画像证据聚合 owner matrix 清晰：每个 source counts owner 各自提供 internal API；不在 backend-profile 内部聚合所有计数，也不把 experience_cards 计数私自塞入当前 B2 schema |
| D-12 | profile_version 用途 | `candidate_profiles.profile_version` 在每次 `UpsertLite` 成功后 `+=1`；用作未来 P1 Insight Cards / 修正覆盖层 / cache invalidation 的版本锚点；P0 不在 API 响应中暴露（B2 `CandidateProfile` schema 未声明 `profileVersion` 字段，本 spec 不绕过 B2 私造响应字段） | 为 P1 留出版本演进入口；不污染 P0 contract |
| D-13 | candidate profile cross-owner 读取 internal API | 暴露 `GetCandidateProfileForUser(userId) -> *CandidateProfile` internal API，返回字段集与 [B2 `CandidateProfile` schema](../openapi-v1-contract/spec.md) / `getMyProfile` 响应一致；**不触发 D-1 seed 路径**：当 `candidate_profiles` 无对应行时返回 nil/Optional，由 caller 决定 fallback（[backend-jobs-recommendations](../backend-jobs-recommendations/spec.md) `getJobMatchProfile` aggregation 在缺 profile 时使用空字段默认值，避免聚合路径产生 backend-profile 副作用）；cross-user 隔离由 caller 提供合法 `userId` 保证（caller 必须已通过 session middleware 解析得到 current_user_id）；P0 由 [backend-jobs-recommendations](../backend-jobs-recommendations/spec.md) `getJobMatchProfile` 消费；不暴露 HTTP；调用不写 audit_events | 把 backend-jobs-recommendations 已在 plan 1.0 declare 的 `backend-profile.GetCandidateProfileForUser` cross-owner 内部 API 形态正式锁定在 owner spec 内；保持 read-only 语义（不 seed），避免聚合路径意外修改 backend-profile 状态；与 D-11 `CountExperienceCardsBySource` 一起构成 backend-profile 对外 read-only internal API 集合 |

### 3.2 待确认事项

| ID | 待确认事项 | 影响 | 默认处理 |
|----|------------|------|----------|
| Q-1 | experience cards 是否在 P0 frontend UI 中暴露 manual 创建入口 | [docs/ui-design/user-profile-and-settings.md](../../ui-design/user-profile-and-settings.md) 当前未展示独立经历卡 CRUD UI；frontend 现在不消费 experience cards endpoint | 默认 P0 backend 实现 5 个 endpoint + IK，但前端 UI 暴露与否由后续 `frontend-profile-and-settings` 设计决定；contract handoff 已就绪 |
| Q-2 | candidate profile 修正覆盖层（用户改写 AI 推断）何时进入 | [user-profile-and-settings.md §4](../../ui-design/user-profile-and-settings.md) 描述了修正流但需要 Insight Cards / AI 推断先落地 | P1：先在 backend-jobs-recommendations 或新 `backend-profile-insights` subspec 落地 AI 推断，再回到 backend-profile 加修正覆盖层 |
| Q-3 | preferred_practice_language / ui_language 双写到 user_settings 由谁负责 | 影响 `/runtime-config` 返回是否一致 | P0 默认不双写；UI owner 在画像页与设置页之间提示一致性；P1 由 backend-auth + A4 owner cross-owner plan 决定 |
| Q-4 | 是否需要 ExperienceCard 删除 endpoint | 当前 [B2](../openapi-v1-contract/spec.md) 未声明 `deleteExperienceCard`；用户可通过 `archived_at` 软删除（store 支持，但 B2 contract 当前未暴露） | 默认 P0 不新增 delete endpoint；如 UI 需要，需要先修订 B2 additive 添加 |

## 4 设计约束

### 4.1 契约约束

- 实现 [B2 §3.1.1](../openapi-v1-contract/spec.md#311-v100-freeze-endpoint-列表) Profile tag 5 个 operationId 的 generated server interface；不允许私造 handler 签名。
- 响应字段集 / status code / 字段顺序与 [B2 fixtures](../openapi-v1-contract/spec.md) `Profile/*.json` 字节比对；如 fixture 中出现的字段在当前 B2 schema 未声明，先修订 B2 fixture，不在 backend 端绕过 fixture 自定义。
- 错误码必须 `$ref` [B1 D-5](../shared-conventions-codified/spec.md#31-已锁定决策) 已锁定的常量集；本 spec 不私造未登记错误码（cross-user 用 `RESOURCE_NOT_FOUND`；参数错误用 `VALIDATION_FAILED`；IK conflict 走 B1 generic IK conflict）。
- D-5 锁定的 `createExperienceCard` / `updateExperienceCard` IK 必须作为 cross-owner additive 修订 [openapi-v1-contract](../openapi-v1-contract/spec.md) 同步落地（新增 `IdempotencyKey` parameter $ref + fixtures 增补 IK header 示例 + inventory 重算），本 spec plan 001 携带 B2 cross-owner addendum。
- 未实现 endpoint 时不得在 cmd/api 中返回 `501 NOT_IMPLEMENTED`；P0 5 个 endpoint 必须全部 wire（区分 P0 / P1 字段，不区分 P0 / P1 endpoint）。

### 4.2 AI 约束

- 本 subject P0 **不调用 AI**：所有 endpoint 是纯 CRUD + Lite identity；Insight Cards / Headline AI 推断 / 经历证据 AI 提取等归后续 `backend-profile-insights` 或 cross-owner plan。
- 不引入 [A3 AIClient](../ai-provider-and-model-routing/spec.md) 依赖；不在 [F3](../prompt-rubric-registry/spec.md) 注册新 feature_key（如 P1 需要 `profile.*` feature_key，由后续 plan 同步修订 F3 D-8 字典）。
- 当 cross-owner evidence write path（resume.parse 写 experience_card from resume）启用时，AI provenance 由 source owner（backend-resume）注入到 `experience_cards.metrics` jsonb 或新加列；本 subject 不在 P0 实现 AI provenance 字段映射。

### 4.3 存储约束

- 使用 [B4 baseline `candidate_profiles`](../db-migrations-baseline/spec.md) + `experience_cards` 表；不引入新 migration（如未来需要 `profile_corrections` / `profile_insight_cards` 表，由 P1 plan 携带 cross-owner B4 additive）。
- 跨用户隔离：所有 read endpoint 必须以 `user_id = current_user_id` 过滤；cross-user 访问返回 404（不暴露存在）。
- 隐私删除调用 `DeleteCandidateProfileForUser(userId)`：先 `experience_cards` 显式删除（保 audit 完整）→ `candidate_profiles` 删除；DB FK CASCADE 兜底；audit tombstone 仅保留 userId / 计数 / 删除时间，不含敏感字段（title / situation / task / action / result / metrics / skills 等）。
- experience_cards `source_type` 默认 `manual`；handler 强制覆盖（即使前端伪造 source_type='resume_parse' 也按 `manual` 落库）。
- `metrics` jsonb 与 `skills text[]` P0 接受任意结构 / 字符串数组；不做 schema 校验；P1 引入 AI extraction 时由 owner 加 schema lint。
- raw user-supplied text（`title` / `situation` / `task` / `action` / `result` / `headline` / `currentRole`）不出现在 audit_events / outbox / log 中（PII 红线）。

### 4.4 cross-owner 接口约束

- `DeleteCandidateProfileForUser(userId)` 仅由 backend internal privacy runner 调用；不暴露 HTTP。
- `CountExperienceCardsBySource(userId)` 仅由 [backend-jobs-recommendations](../backend-jobs-recommendations/spec.md) 内部读取；不暴露 HTTP；返回值不含 card content。
- `GetCandidateProfileForUser(userId)` 仅由 [backend-jobs-recommendations](../backend-jobs-recommendations/spec.md) `getJobMatchProfile` aggregation 内部读取；不暴露 HTTP；返回字段集与 `getMyProfile` 一致；**不触发 seed 路径**（D-13），缺失返回 nil；调用不写 audit_events / 不 bump profile_version；caller 必须保证 `userId` 是合法 current_user。
- 未来 `AppendExperienceCardEvidence(userId, profileId, source_type, source_ref_id, payload)` 在 P1 由 cross-owner plan 携带签名锁定；本 spec 仅保证 schema 支持。

### 4.5 BDD / TDD 约束

- 每个 endpoint 必须有 handler unit test（参数校验 + cross-user + IK where applicable）+ `cmd/api` route wiring test（session middleware / path params）+ store integration test（state transition + cross-user isolation + cursor pagination）。
- 用户可见行为（getMyProfile seed / updateMyProfile patch / createExperienceCard manual / listExperienceCards pagination）必须有 BDD scenario 覆盖；privacy delete 链路必须有专用场景验证 audit 完整与无敏感字段泄漏。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 5 个 Profile HTTP handler | backend-profile | 真实业务逻辑 |
| `candidate_profiles` / `experience_cards` 表 schema | [B4 db-migrations-baseline](../db-migrations-baseline/spec.md) | 字段 / 索引 / FK / check constraint 已在 baseline migration 落地 |
| `cmd/api` runtime wiring | backend-profile + [backend-runtime-topology](../backend-runtime-topology/spec.md) | 挂载 Profile route + session middleware + IK middleware（experience card CUD） |
| frontend 用户画像 UI | 未来 `frontend-profile-and-settings` subspec / [docs/ui-design/user-profile-and-settings.md](../../ui-design/user-profile-and-settings.md) | React 组件、修正流、Insight Cards UI |
| 隐私删除调用 | backend internal privacy runner（[backend-runtime-topology](../backend-runtime-topology/spec.md)） | 调用 `DeleteCandidateProfileForUser` |
| `user_settings` 双写 | [backend-auth](../backend-auth/spec.md) + [secrets-and-config](../secrets-and-config/spec.md) | 本 subject P0 不级联写 user_settings；如需双写由 cross-owner plan |
| `getJobMatchProfile` aggregation | [backend-jobs-recommendations](../backend-jobs-recommendations/spec.md) | 消费本 subject `GetCandidateProfileForUser` (D-13) + `CountExperienceCardsBySource` (D-11) + 其他 owner 计数；本 subject 不实现 JobMatch 聚合；candidate_profiles 表直接 SQL 读取在 backend-profile 边界外不被允许 |
| AI 画像生成（Insight Cards / Headline 推断） | 未来 `backend-profile-insights` 或 cross-owner P1 plan | 本 subject P0 不调用 AI；不引入 A3 / F3 依赖 |
| mock-first fixtures | [B2 fixtures](../openapi-v1-contract/spec.md) | backend-profile handler 响应字节比对 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | getMyProfile seed | 已登录用户 + `candidate_profiles` 表无对应行 | 调 `GET /api/v1/profiles/me` | 返回 200 + Lite profile（headline=null / yearsOfExperience=null / currentRole=null / preferredPracticeLanguage=user_settings.preferred_practice_language / uiLanguage=user_settings.ui_language / region=user_settings.region）；DB 新增 1 行 `candidate_profiles`（user_id UNIQUE）；再次调用返回同一行（不重复 seed） | 001-candidate-profile-and-experience-cards |
| C-2 | updateMyProfile patch | 已登录用户已有 `candidate_profiles` 行 | 调 `PATCH /api/v1/profiles/me` body `{ headline: "Senior frontend...", yearsOfExperience: 5 }` | 返回 200 + 更新后 CandidateProfile；DB 行 `headline` / `years_of_experience` 更新，其他字段保持；`profile_version` += 1；`updated_at` 更新；只提供的字段被更新（不提供的字段不变） | 001-candidate-profile-and-experience-cards |
| C-3 | updateMyProfile validation | 已登录 | 调 PATCH body `{ yearsOfExperience: -1 }` | 返回 422 + `error.code='VALIDATION_FAILED'`；DB 行不变；profile_version 不 bump | 001-candidate-profile-and-experience-cards |
| C-4 | listExperienceCards pagination | 用户 A 有 25 个 experience_card | 调 `GET /api/v1/profiles/me/experience-cards?pageSize=20` 然后 cursor 第二页 | 第一页返回 20 行 + `pageInfo.nextCursor` 非空；第二页返回 5 行 + `hasMore=false`；按 `updated_at DESC, id DESC` 唯一稳定序；cross-user 不可见 | 001-candidate-profile-and-experience-cards |
| C-5 | createExperienceCard manual | 已登录用户 + 有 candidate_profile + IK | 调 `POST /api/v1/profiles/me/experience-cards` body `CreateExperienceCardRequest{ title, companyName, situation, task, action, result, language, skills }` + `Idempotency-Key` | 返回 201 + `ExperienceCard`；DB 新增行 `source_type='manual'`（即使 body 中携带 source_type 也强制为 manual）、`confidence='medium'`、`profile_id=user.candidate_profile_id`；IK 二次重放返回首次 card 不创建新行 | 001-candidate-profile-and-experience-cards |
| C-6 | createExperienceCard IK conflict | 同 IK 不同 fingerprint | 调 create 2 次（body 不同） | 第二次返回 409 + B1 generic IK conflict envelope；DB 只有首次行 | 001-candidate-profile-and-experience-cards |
| C-7 | updateExperienceCard patch | 已登录 + 拥有 cardId + IK | 调 `PATCH /api/v1/profiles/me/experience-cards/{cardId}` body `{ result: "更新结果..." }` + `Idempotency-Key` | 返回 200 + 更新后 ExperienceCard；DB 行 `result` 更新，其他字段保持；`updated_at` 更新；IK 二次重放返回首次结果 | 001-candidate-profile-and-experience-cards |
| C-8 | cross-user 隔离 | 用户 A 有 experienceCard X；用户 B 调 `PATCH /experience-cards/X` | – | 返回 404 + `error.code='RESOURCE_NOT_FOUND'`；不暴露存在；DB 行不变；audit_events 不写敏感字段 | 001-candidate-profile-and-experience-cards |
| C-9 | mock-first 字节比对 | B2 fixture `getMyProfile.json` `default` scenario | 通过 `cmd/api` route 调真实 handler | 响应字段集 / status / header 字节一致；session middleware 不改变 generated response envelope；其他 4 个 endpoint 同样比对各自 fixture | 001-candidate-profile-and-experience-cards |
| C-10 | privacy 删除链路 | 用户 A 有 candidate_profile + 5 experience_cards | privacy_delete job 触发，调 `DeleteCandidateProfileForUser(userId)` | 删除顺序：experience_cards → candidate_profiles；audit tombstone 写 userId / experienceCardCount=5 / 删除时间；不含 title / situation / task / action / result / metrics / skills / headline / currentRole；后续 getMyProfile 调用按 D-1 seed 行为 re-seed 一行新 profile（旧 user 已删，cross-user 隔离）或不可达（取决于 user 是否同时被删除） | 001-candidate-profile-and-experience-cards |
| C-11 | source_type taxonomy enforcement | 前端 create 时 body 携带 `source_type='resume_parse'` | – | handler 强制覆盖为 `manual`；DB 行 `source_type='manual'`；不接受外部伪造系统来源 | 001-candidate-profile-and-experience-cards |
| C-12 | source counts internal API | 用户 A 有 3 个 manual + 2 个 resume_parse experience_card（resume_parse 通过 store layer 直接写入模拟 cross-owner future write） | 调 `CountExperienceCardsBySource(userId)` | 返回 `{manual: 3, resume_parse: 2, practice_report: 0, debrief: 0}` | 001-candidate-profile-and-experience-cards |
| C-13 | profile_version monotonic | 用户连续调 updateMyProfile 3 次 | 每次 patch 不同字段 | `profile_version` 从 1 → 2 → 3 → 4 严格单调递增；不在 API 响应中暴露；store 层 SELECT FOR UPDATE 防并发 race | 001-candidate-profile-and-experience-cards |
| C-14 | 旧模块负向断言 | grep `mistake|growth|drill|experiences|star` in `backend/internal/profile/` 与 outbox payload | – | 0 命中（与 [product-scope D-11 默认丢弃](../product-scope/spec.md#31-已锁定决策) + [engineering-roadmap D-6](../engineering-roadmap/spec.md#31-已锁定决策) 同步） | 001-candidate-profile-and-experience-cards |
| C-15 | cross-owner candidate profile internal API | 用户 A 已 seed candidate_profile（headline / yearsOfExperience / region 非空）；用户 C 未访问过 `getMyProfile`（candidate_profiles 无对应行）| 内部调 `GetCandidateProfileForUser(userA)` 与 `GetCandidateProfileForUser(userC)` | (1) userA 调用返回 `*CandidateProfile` 与 `getMyProfile` 字段集一致（含 headline / yearsOfExperience / currentRole / preferredPracticeLanguage / uiLanguage / region），不写 audit_events，不 bump profile_version；(2) userC 调用返回 nil/Optional（D-13 不 seed 副作用），DB `candidate_profiles` 仍 0 行 userC；(3) 后续 userC 调 `getMyProfile` 仍按 D-1 seed 出一行新 profile（profile_version=1），证明 internal API 不抢占 seed 行为 | 001-candidate-profile-and-experience-cards |

## 7 关联计划

- [001-candidate-profile-and-experience-cards](./plans/001-candidate-profile-and-experience-cards/plan.md)：第一批 plan，落地 5 个 Profile endpoint 的 handler + store + cmd/api wiring + cross-owner B2 IK additive + privacy delete internal API + source counts internal API；BDD 覆盖 getMyProfile seed → updateMyProfile patch → experience cards CRUD → privacy delete 主路径。
- `002-profile-insights-and-corrections`（P1 延后）：落地 Insight Cards / 修正覆盖层 / Headline 自动推断 / cross-owner `AppendExperienceCardEvidence` internal API；同步携带 F3 `profile.*` feature_key 与 cross-owner B4 additive。

## 8 关联文档

- 上游 spec：[`product-scope`](../product-scope/spec.md)、[`engineering-roadmap`](../engineering-roadmap/spec.md)、[`openapi-v1-contract`](../openapi-v1-contract/spec.md)、[`db-migrations-baseline`](../db-migrations-baseline/spec.md)、[`shared-conventions-codified`](../shared-conventions-codified/spec.md)、[`backend-runtime-topology`](../backend-runtime-topology/spec.md)、[`mock-contract-suite`](../mock-contract-suite/spec.md)
- 下游 spec：[`backend-jobs-recommendations`](../backend-jobs-recommendations/spec.md)（消费 `CountExperienceCardsBySource` 与 candidate_profiles 读取）
- UI 真理源：[`docs/ui-design/user-profile-and-settings.md`](../../ui-design/user-profile-and-settings.md)、[`ui-design/src/screen-profile.jsx`](../../../ui-design/src/screen-profile.jsx)
- 历史：[history.md](./history.md)
