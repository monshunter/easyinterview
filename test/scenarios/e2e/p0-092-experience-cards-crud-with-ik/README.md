# E2E.P0.092 Experience Cards CRUD with IK

> **场景 ID**: E2E.P0.092
> **执行方式**: automated
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

A2 dev stack 已拉起；`cmd/api` 真实路由 + `auth.SessionMiddleware` + 真实 `backend/internal/profile` runtime + idempotency middleware（仅 experience card POST/PATCH）就位。三个用户（A / B / C）已登录；user_settings 已就位；B2 IK additive 已落地（`createExperienceCard` / `updateExperienceCard` operation 含 `Idempotency-Key` parameter）。缺少 live env、focused gate no-op 或 integration-tag 测试 skip 时本场景必须 fail。

## 2 When

`scripts/trigger.sh` 通过 `cmd/api` 路由依序触发：
- 用户 A `POST /experience-cards` (IK ikCreate)：first create
- 用户 A `POST /experience-cards` (IK ikCreate, same body)：IK replay
- 用户 A `POST /experience-cards` 无 IK：422
- 用户 A `GET /experience-cards?pageSize=20`：list
- 用户 B `PATCH /experience-cards/{A.id}` (IK ikCross)：cross-user 404 `RESOURCE_NOT_FOUND`
- 内部 API `CountExperienceCardsBySource(userA)`
- 内部 API `GetCandidateProfileForUser(userC)` 返回 nil（D-13）
- 用户 C `GET /profiles/me`：仍按 D-1 seed
- 内部 API `GetCandidateProfileForUser(userC)` post-seed 返回 *CandidateProfile

底层入口：`cd backend && go test ./cmd/api -run TestProfileHTTPScenario -count=1 -v`（命中本地 dev-stack Postgres）。

## 3 Then

- 创建成功 201 + `source_type=manual`（即使 body 携带其他 source_type 也覆盖）、`confidence=medium`、profile_id 关联到 user_A 的 candidate_profile
- IK replay 返回首次 card，DB 不重复
- 缺 IK 返回 422 + `error.code=VALIDATION_FAILED`
- cross-user PATCH 返回 404 + `error.code=RESOURCE_NOT_FOUND`
- `CountExperienceCardsBySource` 返回 manual:>=1 + 其它 source_type 默认 0
- `GetCandidateProfileForUser(userC)` D-13：第一次返回 nil；userC 调 `GET /profiles/me` 后才创建 candidate_profile；第二次内部读取返回非 nil
- 旧模块负向 grep 0 命中；audit_events / outbox 不携带 raw card 内容
