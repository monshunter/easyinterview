# Backend Resume Asset Register Parse and Listing Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-11

**关联计划**: [plan](./plan.md)

## Phase 1: register / get handler skeleton + sourceType 三路

- [ ] 1.1 实现 `backend/internal/resume/handler/register.go`，generated server interface `RegisterResume`（验证：编译 PASS + `go vet` PASS）
- [ ] 1.2 sourceType 三路参数校验：`upload` 必带 fileObjectId / `paste` 必带 rawText / `guided` 必带 guidedAnswers；其他组合 422（验证：unit test `TestRegisterSourceType` 3 路 + 错误组合 PASS）
- [ ] 1.3 upload 路径：调用 [backend-upload `RegisterFileObject(fileObjectId, expectedPurpose=resume, ownerUserId)`](../../../backend-upload/spec.md) internal API 校验后，写入 `resume_assets.file_object_id` FK（验证：integration test verify FK 建立）
- [ ] 1.4 IK 校验（缺失 / 24h replay 返回首次 resumeAssetId / mismatch 422）（验证：unit test `TestRegisterIdempotency` PASS）
- [ ] 1.5 返回 202 + `ResumeAssetWithJob{resumeAssetId, job(jobType=resume_parse, status=queued)}`，与 [B2 fixture `registerResume.json` `default` scenario](../../../mock-contract-suite/spec.md) 字节一致（验证：fixture parity test）
- [ ] 1.6 实现 `backend/internal/resume/handler/get.go`，generated server interface `GetResume`（验证：编译 PASS）
- [ ] 1.7 getResume：cross-user 返回 404 + 不暴露存在（验证：integration test cross-user PASS）

## Phase 2: resume_assets store + state machine

- [ ] 2.1 实现 `backend/internal/resume/store/assets.go` Repository：`Create / Get / List(cursor, pageSize) / MarkParsing / MarkReady / MarkFailed / DeleteForUser`（验证：编译 PASS）
- [ ] 2.2 parse_status state machine：`queued → processing → ready | failed`；非法转换拒绝（验证：unit test `TestParseStatusTransition` PASS）
- [ ] 2.3 cursor pagination 实现：按 `updated_at DESC, id DESC` 唯一稳定序（验证：integration test 25 行 + 第二页 cursor PASS）
- [ ] 2.4 integration test：CRUD + state transition + cross-user isolation + FK 约束（验证：`go test ./internal/resume/store/... -tags=integration` PASS）

## Phase 3: resume.parse async job + AIClient 集成

- [ ] 3.1 实现 `backend/internal/resume/jobs/parse.go`，注册到 backend internal runner（job_type=resume_parse, dotted=resume.parse）（验证：runner registry 测试）
- [ ] 3.2 从 `resume_assets` 读 file_object（upload）/ original_text（paste）/ guided_answers jsonb（guided）作为 prompt input；guided 不从 `original_text` 反序列化（验证：unit test verify 三路 input 路径）
- [ ] 3.3 通过 [A3 AIClient](../../../ai-provider-and-model-routing/spec.md) 调 [F3 `resume.parse` feature_key](../../../prompt-rubric-registry/spec.md)；不 hardcode prompt 正文（验证：unit test stub AIClient verify profile / feature_key 路由）
- [ ] 3.4 解析 LLM JSON 输出 → 写 `parsed_summary` + `parse_status='ready'`（验证：unit test `TestResumeParseHappyPath`）
- [ ] 3.5 失败路径：AI timeout / output_invalid → `parse_status='failed'` + `error_code`；retryable 信息落在 `async_jobs` retry metadata，不向 `resume_assets.parse_status` 私加 `failed_retryable`（验证：unit test `TestResumeParseFailureRetryable`）
- [ ] 3.6 写入 `ai_task_runs` typed columns：model_profile_name / model_profile_version / prompt_version / rubric_version / route / validation_status（验证：integration test verify `ai_task_runs` 行）
- [ ] 3.7 outbox `resume.parse.completed`：envelope 字段（resumeAssetId / userId / parseStatus）写入 outbox_events；PII 边界断言不含 raw text / guided answers / parsed_summary（验证：unit test + payload assertion）
- [ ] 3.8 PII leak negative：log / audit / outbox payload 写入路径不序列化 raw resume content、`guided_answers` 内容、prompt body 或 model raw response；允许 SQL/store 层出现必要列名（如 `original_text` / `guided_answers` / `parsed_summary`），禁止把列值写入日志或事件（验证：专用 lint / unit test 覆盖 log sink 与 outbox payload）

## Phase 4: listResumes handler

- [ ] 4.1 实现 `backend/internal/resume/handler/list.go`，generated server interface `ListResumes`（验证：编译 PASS）
- [ ] 4.2 cursor pagination 实现 + 返回 `PaginatedResumeAsset{items, pageInfo}`（验证：integration test 25+ 行 + 第二页 PASS）
- [ ] 4.3 cross-user 过滤：仅返回 `user_id = current_user_id` 行（验证：integration test cross-user PASS）
- [ ] 4.4 字节比对 [B2 fixture `listResumes.json`](../../../mock-contract-suite/spec.md) `default` / `empty` / `paginated` 三 variant（验证：fixture parity test）

## Phase 5: 收口 + BDD + 解锁 workspace 001

- [ ] 5.1 跑 `make backend-test` + `go test ./internal/resume/...` 全 PASS（验证：exit 0）
- [ ] 5.2 mock-first 对齐：3 个新 handler 响应与对应 fixture `default` scenario 字节比对 PASS
- [ ] 5.3 grep `inline|rewrite|mirror` in `backend/internal/resume/`：0 命中（C-13 negative）（验证：`git grep` 输出）
- [ ] 5.4 BDD-Gate: E2E.P0.034 resume-register-and-list PASS（详见 [bdd-checklist.md](./bdd-checklist.md)）
- [ ] 5.5 BDD-Gate: E2E.P0.035 resume-parse-async-job-lifecycle PASS（含 stub AIClient + outbox event 验证）
- [ ] 5.6 在 `test/scenarios/e2e/INDEX.md` 追加 P0.034 + P0.035 行（关联需求 `backend-resume C-1..C-8, C-13`）
- [ ] 5.7 同步 `docs/spec/engineering-roadmap/spec.md` §5.2 `backend-resume` 状态从 "未创建" 改为 "active"（与 backend-upload 同步行）（验证：`sync-doc-index --check`）
- [ ] 5.8 通知 [frontend-workspace-and-practice/001](../../../frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md) owner：`listResumes` 已就位，可启动 disabled-list → active-list 原地修订（验证：cross-plan 引用 commit + workspace 001 plan checklist 追加 unblock 引用）
