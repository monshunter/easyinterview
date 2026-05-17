# 002 BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-17

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.074 resume confirm master and version reads

- [ ] 创建场景目录 `test/scenarios/e2e/p0-074-resume-confirm-master-and-version-reads/`，含 `README.md`（baseline + 离线限制）+ `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 B2 fixture scenario + 测试数据：`confirmResumeStructuredMaster.json` 全 4 scenario、`getResumeVersion.json` `default` / `not-found-404`、`listResumeVersions.json` `default` / `empty` / `paginated`；2 个测试用户（A / B）；用户 A 通过 backend-resume/001 register 路径 + resume.parse 后拥有 first ready resume_asset（带 parsed_summary 与 parsed_text_snapshot）+ 第二个 ready resume_asset（pagination empty 用）；准备 22 行版本（1 master + 21 branched targeted）覆盖 pagination；第三个 parse_status='processing' asset 用于 422 PARSE_NOT_READY
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + migration 000007 + `cmd/api` resume runtime + session / IK middleware + 用户登录 + 批量 register 三个 resume_asset + 推进 parse 完成）/ `scripts/trigger.sh`（依序触发 A1 confirm + A2 IK replay + A3 409 + A4 422 + A5 getResumeVersion + A6/A7 listResumeVersions + A8 empty + A9 invalid cursor + B1/B2 cross-user + C1 parse_not_ready 422，并运行 `cd backend && go test ./cmd/api -run 'TestResumeConfirmStructuredMasterHTTPScenario|TestResumeVersionReadHTTPScenario' -count=1 -v`）/ `scripts/verify.sh`（断言 `method=cmd-api-http` + no-op / skip 不可 PASS + DB state + partial UNIQUE INDEX 兜底 + 字节比对 fixture + cross-user 404 + cursor 序稳定性 + 隐私 grep + 旧口径 grep）/ `scripts/cleanup.sh`（清理 resume_versions + resume_assets + 用户 + 回滚 migration 000007 可选）
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-074-resume-confirm-master-and-version-reads/trigger.log` + `cmd/api` HTTP scenario log + verify 输出 + DB 行 dump + partial UNIQUE INDEX 兜底证据 + fixture byte diff 0 + cross-user 404 + cursor 序证据 + 隐私 grep 0 命中 + `method=cmd-api-http` 或等价 live runtime evidence + no no-op / no skip 证据
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.074 行（关联需求 `backend-resume C-6, C-14, C-15`，状态 Ready，automated）

## E2E.P0.075 resume update version merge and IK

- [ ] 创建场景目录 `test/scenarios/e2e/p0-075-resume-update-version-merge-and-ik/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 B2 fixture + 测试数据：`updateResumeVersion.json` `default` / `validation-error-422`，并由本 plan Phase 4.7 补齐 `idempotency-replay` fixture；用户 A 拥有 1 行 ready structured_master version；用户 B 已登录无版本；准备额外被 soft-delete 的 version 用于 deleted 404 case
- [ ] 实现 `scripts/setup.sh`（dev stack + 推进 confirm v1 + 用户登录）/ `scripts/trigger.sh`（A1 partial merge + A2 IK replay + A3 IK conflict + A4 422 不可编辑 + A5 focusAngle/matchScore + A6 cross-user + A7 deleted_at，并运行 `cd backend && go test ./cmd/api -run TestResumeUpdateVersionHTTPScenario -count=1 -v`）/ `scripts/verify.sh`（DB merge semantic 断言 + server-reset provenance 断言 + 字节比对 + 隐私 grep + `method=cmd-api-http`）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-075-resume-update-version-merge-and-ik/trigger.log` + DB merge 轨迹 + fixture byte diff 0 + cross-user 404 + deleted 404 + 隐私 grep 0 命中 + live runtime evidence
- [ ] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.075 行（关联需求 `backend-resume C-14`，状态 Ready，automated）

## E2E.P0.076 resume branch version sync paths

- [ ] 创建场景目录 `test/scenarios/e2e/p0-076-resume-branch-version-sync-paths/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 B2 fixture + 测试数据：`branchResumeVersion.json` `copy-master-sync` + `blank-sync`（本 plan Phase 5.7 补齐 fixture）；用户 A 拥有 ready structured_master version vmaster + ready targetJob tj1；用户 B 已登录有自己的 vmaster_b
- [ ] 实现 `scripts/setup.sh`（dev stack + register + confirm + targetJob 准备）/ `scripts/trigger.sh`（A1 copy_master + A2 IK replay + A3 blank + A4 parent 404 + A5 targetJob 404 + A6 422 invalid seed_strategy + B1/B2 cross-user，运行 `cd backend && go test ./cmd/api -run TestResumeBranchVersionHTTPScenario -count=1 -v`）/ `scripts/verify.sh`（DB structured_profile 拷贝断言 + blank 空字段 + 不入队 async_jobs 断言 + 字节比对 fixture + cross-user 404 + `method=cmd-api-http`）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-076-resume-branch-version-sync-paths/trigger.log` + DB resume_versions / async_jobs 行断言 + fixture byte diff 0 + 隐私 grep + live runtime evidence
- [ ] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.076 行（关联需求 `backend-resume C-10`，状态 Ready，automated）

## E2E.P0.077 resume tailor async dispatch and ready

- [ ] 创建场景目录 `test/scenarios/e2e/p0-077-resume-tailor-async-dispatch-and-ready/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture + stub provider + 测试数据：`branchResumeVersion.json` `ai-select-202-with-job`（本 plan Phase 5.7 补齐 fixture）+ `requestResumeTailor.json` `default`（Phase 6.8 补齐 `Idempotency-Key` 请求头）/ `idempotency-replay` + `getResumeTailorRun.json` `default` / `queued` / `generating` / `failed`；A3 AIClient stub 返回 success JSON（matchSummary + suggestions[3]）；F3 `resume.tailor.gap_review` + `resume.tailor.bullet_suggestions` feature_key ready；用户 A 拥有 ready resume_asset + structured_master + targetJob
- [ ] 实现 `scripts/setup.sh`（dev stack + `cmd/api` in-process resume_tailor drainer 启动 + stub provider 注入 + 用户登录 + register + confirm + targetJob 准备）/ `scripts/trigger.sh`（A1 ai_select branch + A2 requestTailor + A3 drainer `RunOnce(resume_tailor)` + A4 getTailorRun ready + A5 getTailorRun mid-state + A6 IK replay，运行 `cd backend && go test ./cmd/api -run 'TestResumeBranchVersionHTTPScenario|TestResumeTailorEndpointsHTTPScenario|TestResumeTailorDrainerHTTPScenario' -count=1 -v`）/ `scripts/verify.sh`（DB resume_tailor_runs / resume_version_suggestions / ai_task_runs / outbox_events 行断言 + outbox payload PII 边界（不含 suggested bullet / match_summary 文本）+ ready-only outbox + 字节比对 fixture + `method=cmd-api-http`）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-077-resume-tailor-async-dispatch-and-ready/trigger.log` + drainer scenario log + DB 4 张表轨迹 + outbox payload dump（仅 `tailorRunId/resumeAssetId/targetJobId/mode/status`）+ stub provider call log + live runtime evidence + no no-op / no skip
- [ ] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.077 行（关联需求 `backend-resume C-10, C-16`，状态 Ready，automated）

## E2E.P0.078 resume tailor failure and retry

- [ ] 创建场景目录 `test/scenarios/e2e/p0-078-resume-tailor-failure-and-retry/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 stub provider 三 variant（timeout / output_invalid / retry 成功）+ 测试数据：用户 A 3 个 queued resume_tailor_runs 对应三 variant；F3 feature_key ready
- [ ] 实现 `scripts/setup.sh`（dev stack + drainer + stub provider variant 注入 + 用户登录 + register + confirm + targetJob + requestTailor 3 次）/ `scripts/trigger.sh`（A drainer `RunOnce` 处理 M1 timeout + B `RunOnce` 处理 M2 output_invalid + C1 `RunOnce` 处理 M3 首次 timeout + C2 retry 后 `RunOnce` 处理 M3 success + D shutdown，运行 `cd backend && go test ./cmd/api -run TestResumeTailorDrainerFailureScenario -count=1 -v`）/ `scripts/verify.sh`（DB resume_tailor_runs `queued -> generating -> failed` / retry `failed -> generating -> ready` 状态机 + error_code + async_jobs retry metadata + ai_task_runs 多行 + outbox ready-only 断言 + shutdown 无 goroutine leak + 旧口径 grep + 隐私 grep + `method=cmd-api-http`）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-078-resume-tailor-failure-and-retry/trigger.log` + DB status 转换轨迹 + `async_jobs` attempt metadata + `ai_task_runs` 多行 dump + outbox 仅 1 行 completed event 断言 + shutdown 证据 + 隐私 grep 0 命中 + live runtime evidence
- [ ] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.078 行（关联需求 `backend-resume C-16`，状态 Ready，automated）

## E2E.P0.079 resume suggestion accept reject terminal

- [ ] 创建场景目录 `test/scenarios/e2e/p0-079-resume-suggestion-accept-reject-terminal/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture + 测试数据：`acceptResumeTailorSuggestion.json` / `rejectResumeTailorSuggestion.json` `default` / `idempotency-replay` / `already-decided-409`（缺 variant 由本 plan Phase 8.7 补齐 fixture，并将当前 `conflict-409` + `TARGET_INVALID_STATE_TRANSITION` 漂移替换为 `VALIDATION_FAILED` + `detail.reason='SUGGESTION_ALREADY_DECIDED'`）；用户 A 拥有 1 行 ready targeted version + 5 个 pending suggestions（带初始 structured_profile 状态供 D-12 断言）；用户 B 已登录有自己 suggestion
- [ ] 实现 `scripts/setup.sh`（dev stack + 推进 ai_select tailor 流程产生 suggestion 或直接 seed 5 个 pending suggestion + 用户登录）/ `scripts/trigger.sh`（A1 accept s1 + A2 IK replay + A3 already-decided 409 + A4 reject s2 + A5 already-decided 409 + A6 not found 404 + B1/B2 cross-user，运行 `cd backend && go test ./cmd/api -run TestResumeSuggestionAcceptRejectHTTPScenario -count=1 -v`）/ `scripts/verify.sh`（DB suggestions status 状态机 + decided_at + **`resume_versions.structured_profile` 不被改动**（D-12）+ IK middleware replay + 字节比对 fixture + cross-user 404 + 隐私 grep + `method=cmd-api-http`）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-079-resume-suggestion-accept-reject-terminal/trigger.log` + DB suggestion 状态机轨迹 + `resume_versions.structured_profile` 不变断言 + IK replay 证据 + cross-user 404 + 隐私 grep 0 命中 + live runtime evidence
- [ ] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.079 行（关联需求 `backend-resume C-11`，状态 Ready，automated）

## E2E.P0.080 resume versions privacy and legacy negative

- [ ] 创建场景目录 `test/scenarios/e2e/p0-080-resume-versions-privacy-legacy/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备前置：E2E.P0.074-079 均已 PASS（依次写入 resume_versions / resume_tailor_runs / resume_version_suggestions / ai_task_runs / outbox_events / audit_events 行）；本场景只做 regression / privacy negative 反查
- [ ] 实现 `scripts/setup.sh`（dev stack 拉起 + 重放前序场景数据 fixture 或共享 DB state）/ `scripts/trigger.sh`（grep 检查：A `git grep -nE 'inline|rewrite|mirror' backend/internal/resume/` + B `git grep -nE 'mistakes|growth|drill|inline-debrief-record' backend/internal/resume/` + C `cd backend && go test ./internal/resume/... -run 'TestOutboxPrivacy|TestAuditPrivacy|TestAiTaskRunsPrivacy' -count=1 -v` + D outbox payload assertion unit test 重跑；当前 plan / BDD prose 与历史 out-of-scope 文档不纳入 zero-reference gate，避免说明文字自匹配）/ `scripts/verify.sh`（断言 grep 0 命中 + outbox payload 字段集严格匹配 + ai_task_runs 不含 raw text / suggested bullet + audit_events 不含 prompt body + `method=cmd-api-http`）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-080-resume-versions-privacy-legacy/trigger.log` + grep 输出 0 命中 + outbox payload 字段集合断言 + ai_task_runs / audit_events PII grep 输出 + live runtime evidence
- [ ] 在 `test/scenarios/e2e/INDEX.md` 追加 P0.080 行（关联需求 `backend-resume C-13`，状态 Ready，automated）
