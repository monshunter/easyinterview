# E2E.P0.083 Profile Privacy Delete Lifecycle

> **场景 ID**: E2E.P0.083
> **执行方式**: automated
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **状态**: Ready

## 1 Given

A2 dev stack 已拉起；`cmd/api` 真实路由 + `backend/internal/profile` runtime + privacy service 就位；用户 A 已有 candidate_profile 行（headline 等字段非空）+ ≥1 experience_card 行（含 title / situation / task / action / result 真实内容）；其他用户 B / C 不受影响；privacy_delete job ID 已就位。缺少 live env、focused gate no-op 或 integration-tag 测试 skip 时本场景必须 fail。

## 2 When

`scripts/trigger.sh` 执行：
- `TestProfileHTTPScenario` 完整跑一遍（在末段调用 internal `DeleteCandidateProfileForUser(userA, "scenario-job")`）
- `TestPrivacyDeleteOrderAndAudit` 单元测试覆盖 fake-store 上的删除链路 + audit tombstone 写入 + 无敏感字段

底层入口：
- `cd backend && go test ./internal/profile/service -run TestPrivacyDeleteOrderAndAudit -count=1 -v`
- `cd backend && go test ./cmd/api -run TestProfileHTTPScenario -count=1 -v`

## 3 Then

- 删除顺序：先 `experience_cards.DeleteForUser` 再 `candidate_profiles.DeleteForUser`（service layer 显式实现）
- audit_events 新增行：`action='profile.privacy_delete'`，`resource_type='candidate_profile'`，metadata 仅含 `experienceCardCount` / `deletedAt` / `jobId`；不包含 raw card content / headline / currentRole / region
- DB state：user A 在 `candidate_profiles` / `experience_cards` 均为 0 行；其他用户不受影响
- 后续 `GET /profiles/me` 仍按 D-1 重新 seed 一行新 profile（profile_version=1）
- `CountExperienceCardsBySource(userA)` 在删除后返回 `{ manual:0, resume_parse:0, practice_report:0, debrief:0 }`
- 旧模块负向 grep 0 命中；trigger.log 中不出现 raw card title / situation / result 字符串

证据：`.test-output/e2e/p0-083-profile-privacy-delete-lifecycle/trigger.log` 包含两个 `PASS` 行以及 `ok  github.com/monshunter/easyinterview/backend/...`；`verify.sh` 检查 audit metadata 与 negative grep。
