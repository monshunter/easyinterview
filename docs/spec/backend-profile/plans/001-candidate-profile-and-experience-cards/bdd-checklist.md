# 001 BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-21

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.081 candidate-profile-seed-and-patch

- [ ] 创建场景目录 `test/scenarios/e2e/p0-081-candidate-profile-seed-and-patch/`，含 `README.md`（背景 + baseline + 离线限制）+ `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 B2 fixture scenario + 测试数据：`getMyProfile` `default`、`updateMyProfile` `default`；2 个测试用户（A / B），各自 user_settings 行已就位；用户 A 起始无 candidate_profile；缺少 live env、integration-tag 测试 skip 或 focused gate no-op 时本场景必须 fail
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + `cmd/api` profile route + session middleware + 用户登录 + user_settings seed）/ `scripts/trigger.sh`（依序触发 A1/A2/A3/A4/A5/B1，并运行 `cd backend && go test ./cmd/api -run TestProfileHTTPScenario -count=1 -v`）/ `scripts/verify.sh`（断言 `method=cmd-api-http` 或等价 live runtime evidence + no-op / skip 不可 PASS + DB state + profile_version 单调 + seed 默认值 + cross-user 隔离 + validation 直接断言 + fixture scenario 字节比对 + 隐私 grep + 旧口径 grep）/ `scripts/cleanup.sh`（清理 candidate_profiles + experience_cards + 用户）
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-081-candidate-profile-seed-and-patch/trigger.log` + `cmd/api` HTTP scenario log + verify 输出 + DB profile_version 轨迹 + getMyProfile / updateMyProfile fixture byte diff 0 + validation error direct assertion + 隐私 grep 0 命中 + cross-user seed 验证 + `method=cmd-api-http` 或等价 live runtime evidence + no no-op / no skip 证据
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.081 行（关联需求 `backend-profile C-1, C-2, C-3, C-13, C-14`，状态 Ready，automated）

## E2E.P0.082 experience-cards-crud-with-ik

- [ ] 创建场景目录 `test/scenarios/e2e/p0-082-experience-cards-crud-with-ik/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 B2 fixture scenario + 测试数据：`listExperienceCards` `default`、`createExperienceCard` `default` + B2 cross-owner IK additive 后的 IK header 示例、`updateExperienceCard` `default` + IK；3 个测试用户（A / B / C，C 用于 D-13 未 seed 路径验证）；用户 A 已通过 store 直接写入 25 个 experience_card 行（混合 source_type：12 manual + 8 resume_parse + 3 practice_report + 2 debrief）+ 用户 A 已有 candidate_profile（headline / yearsOfExperience / region 真实非空）；用户 C `candidate_profiles` 行 0；缺少 live env、integration-tag 测试 skip 或 focused gate no-op 时本场景必须 fail
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + `cmd/api` profile route + session middleware + IK middleware + 3 用户登录 + 批量 store-write 25 个 experience_card + user_A candidate_profile seed + user_C 保持未 seed）/ `scripts/trigger.sh`（依序触发 A1/A2/A3/A4/A5/A6/A7/A8/B1/C1/C2/C3/C4/D1，并运行 `cd backend && go test ./cmd/api -run TestProfileHTTPScenario -count=1 -v` + `cd backend && go test ./internal/profile/service -run 'TestCountExperienceCardsBySource|TestGetCandidateProfileForUserSeededAndNil' -count=1 -v`）/ `scripts/verify.sh`（断言 `method=cmd-api-http+internal-api` 或等价 live runtime evidence + no-op / skip 不可 PASS + DB state + source_type 强制 manual + IK replay vs conflict + cross-user 404 + cursor pagination 边界 + `CountExperienceCardsBySource` 返回正确 + `GetCandidateProfileForUser` 已 seed 返回字段一致 / 未 seed 返回 nil / DB 仍 0 行 userC / 后续 userC `GET /profiles/me` 仍能 seed + fixture scenario 字节比对 + 隐私 grep + 旧口径 grep）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-082-experience-cards-crud-with-ik/trigger.log` + `cmd/api` HTTP scenario log + service-layer internal API log + verify 输出 + DB source_type force 证据 + IK replay vs conflict 证据 + cursor 序稳定性 + CountBySource 输出 + GetCandidateProfileForUser seeded vs nil 输出 + userC re-seed 证据（cross-owner read 不抢占 seed）+ cross-user 404 验证 + experience cards 5 个 fixture byte diff 0 + 隐私 grep 0 命中 + `method=cmd-api-http+internal-api` 或等价 live runtime evidence + no no-op / no skip 证据
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.082 行（关联需求 `backend-profile C-4, C-5, C-6, C-7, C-8, C-9, C-11, C-12, C-14, C-15`，状态 Ready，automated）

## E2E.P0.083 profile-privacy-delete-lifecycle

- [ ] 创建场景目录 `test/scenarios/e2e/p0-083-profile-privacy-delete-lifecycle/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备测试数据：用户 A 已登录 + candidate_profile 行（headline / currentRole 非空 真实内容）+ 5 experience_card 行（含 title / situation / task / action / result / metrics / skills 真实内容）；其他用户 B / C 不受影响；privacy_delete job 模拟入口已就位；缺少 live env、no-op、skip 时本场景必须 fail
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + `cmd/api` runtime + profile service + 用户登录 + 写入 user_A 真实内容）/ `scripts/trigger.sh`（调用 internal `DeleteCandidateProfileForUser(user_A)` + 触发完整删除链路 + 运行 `cd backend && go test ./internal/profile/service -run TestPrivacyDeleteOrderAndAudit -count=1 -v` + `cd backend && go test ./cmd/api -run TestProfileHTTPScenario -count=1 -v`）/ `scripts/verify.sh`（断言 `method=internal-api+cmd-api-http` 或等价 live runtime evidence + no-op / skip 不可 PASS + 删除顺序 experience_cards → candidate_profiles + audit tombstone 完整无敏感字段 + DB 状态 user_A 0 行 + 其他用户不受影响 + 后续 getMyProfile re-seed + CountBySource 全 0 + 失败回滚 + 隐私 grep + 旧口径 grep）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-083-profile-privacy-delete-lifecycle/trigger.log` + 删除链路调用 log + audit tombstone payload dump（验证无敏感字段）+ DB state 删除前后对比 + cross-user 不影响验证 + 失败回滚证据 + `method=internal-api+cmd-api-http` 或等价 live runtime evidence + no no-op / no skip 证据
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.083 行（关联需求 `backend-profile C-10, C-14`，状态 Ready，automated）
