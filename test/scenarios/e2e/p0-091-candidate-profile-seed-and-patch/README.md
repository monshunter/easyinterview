# E2E.P0.091 Candidate Profile Seed and Patch

> **场景 ID**: E2E.P0.091
> **执行方式**: automated
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

A2 dev stack 已拉起；`cmd/api` 真实路由 + `auth.SessionMiddleware` + 真实 `backend/internal/profile` runtime（store / settings reader / service）就位。三个用户（A / B / C）有效 session cookie + `user_settings` 行（preferred_practice_language=`en` / ui_language=`zh-CN` / region=`CN-SH`），用户 A 起始无 candidate_profile 行，用户 B 起始无 profile，用户 C 起始无 profile。缺少 live env、focused gate no-op 或 integration-tag 测试 skip 时本场景必须 fail。

## 2 When

`scripts/trigger.sh` 通过 `cmd/api` 路由依序触发：
- A1: `GET /api/v1/profiles/me` 用户 A 首次访问（seed）
- A2: `GET /api/v1/profiles/me` 用户 A 二次访问（不重复 seed）
- A3: `PATCH /api/v1/profiles/me` 用户 A `{ headline, yearsOfExperience }`
- A5: `PATCH /api/v1/profiles/me` 用户 A `{ yearsOfExperience: -1 }`（违反 minimum）
- B1: `GET /api/v1/profiles/me` 用户 B 首次访问（独立 seed）
- C1: `GET /api/v1/profiles/me` 用户 C 首次访问（独立 seed）

底层入口：`cd backend && go test ./cmd/api -run TestProfileHTTPScenario -count=1 -v`，命中本地 dev-stack Postgres。

## 3 Then

- A1 返回 200，CandidateProfile JSON `{ headline:null, yearsOfExperience:null, currentRole:null, preferredPracticeLanguage:"en", uiLanguage:"zh-CN", region:"CN-SH" }`；DB `candidate_profiles` 新增 user_A 行 profile_version=1。
- A2 返回 200，DB 仍 1 行 user_A，profile_version=1。
- A3 返回 200，DB user_A `headline`/`years_of_experience` 更新；profile_version=2；其他字段保持。
- A5 返回 422 + `error.code='VALIDATION_FAILED'`；DB user_A profile_version 不变。
- B1 / C1 返回 200，DB 各新增独立 user_B / user_C 行；与 user_A 完全隔离。
- audit_events 不写敏感字段；`grep mistake|growth|drill|experiences|star backend/internal/profile/` 0 命中。
- backend log / outbox payload / URL / cookie 不携带 raw profile 字段值。

证据：`.test-output/e2e/p0-091-candidate-profile-seed-and-patch/trigger.log` 必须包含 `ok  github.com/monshunter/easyinterview/backend/cmd/api`；`verify.sh` 强校验 method 与 no-op 排除。
