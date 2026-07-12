# 001 BDD Plan

> **版本**: 1.12
> **状态**: completed
> **更新日期**: 2026-07-12

**关联 Plan**: [plan](./plan.md)

## 1 场景矩阵

| 场景 ID | 类别 | 关联 Phase | 关联 Spec C-* | 关联 BDD-Gate（主 checklist） |
|---------|------|-----------|--------------|----------------------------|
| E2E.P0.034 | primary + alternate · register upload/paste + active count limit + getResume + getResumeSource + listResumes pagination + cross-user 隔离 | Phase 1 + 2 + 4 + 11 + 12 | C-1, C-2, C-5, C-6, C-7, C-8, C-10, C-14 | Phase 5.4 / 11.3 / 12.3 |
| E2E.P0.035 | primary + failure / recovery · resume.parse async job lifecycle + deterministic full-resume snapshot + long-input tail marker + output truncation fail-closed + outbox event + AI failure retryable + DOCX rejection | Phase 3 + 5 + 11 + 12 + 13 + 14 | C-3, C-4, C-13 | Phase 5.5 / 11.1 / 11.2 / 12.1 / 12.2 / 13.4 / 14.6 |

---

## Phase 1 + 2 + 4: register / get / list 主路径 + 双 sourceType

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.034 | resume register + get + list 全链路 + 双 sourceType + IK replay + cross-user | A2 dev stack 拉起；`backend-upload/001` completed，且 `RegisterFileObject` 已具备 object `Stat` + size mismatch rejection；`cmd/api` 使用真实 session middleware、IK middleware 与 resume route；用户 A 已登录（有效 session cookie）；用户 A 通过 `createUploadPresign` 使用 `purpose=resume` 上传 1 个 PDF 取得 fileObjectId；用户 A 调 paste 路径以 5KB rawText 创建；继续创建 25 个 resume 用于 pagination；用户 B 已登录但无 resume；缺少 live env、integration-tag 测试 skip 或 focused gate no-op 时本场景必须 fail | 用户 A 通过 `cmd/api` route 分别调：（A1）`POST /api/v1/resumes` upload + IK / paste 各 1 次；（A2）同 IK replay upload；（A3）`GET /api/v1/resumes/{A1.upload}`；（B1）用户 B 调 `GET /api/v1/resumes/{A1.upload}`；（C1）用户 A 调 `GET /api/v1/resumes?pageSize=20`；（C2）用户 A 调 `GET /api/v1/resumes?pageSize=20&cursor={C1.nextCursor}`；（D1）参数非法 sourceType=unknown / unsupported；（D2）upload object missing / size mismatch | （A1）upload / paste 返回 202 + `ResumeWithJob{resumeId, job(jobType=resume_parse, status=queued)}`；DB `resumes` 行 `parse_status='queued'`，且同事务存在 `async_jobs(job_type=resume_parse, resource_type=resume_asset)`；upload 路径通过 backend-upload `RegisterFileObject(fileObjectId, resume, userId)` 校验 object exists + actual size 后建立 `file_object_id` 引用 + `source_type='upload'`；paste 路径 `original_text` 写入 + `source_type='paste'` + `file_object_id` NULL；（A2）IK replay 返回首次 resumeId + 不创建新 DB 行 / async_jobs / outbox side effect；（A3）返回 200 + `Resume` 字段集；（B1）返回 404，不暴露存在；audit_events 不写敏感字段；（C1）返回 20 行 + `pageInfo.nextCursor` 非空 + `pageSize=20` + 按 `updated_at DESC, id DESC` 排序；（C2）返回 5 行 + `hasMore=false`；（D1/D2）返回 422 + `error.code="VALIDATION_FAILED"`，且不创建 `resumes`；（E 字节比对）`registerResume` 对齐 B2 fixture `default` / `paste-text`，`getResume` 对齐 `default` / `not-found`，`listResumes` 对齐 `default` / `empty` / `paginated`；validation 错误用直接断言，不声称存在 B2 error fixture；（F 隐私）raw text / parsed_summary 不出现在 console / URL / localStorage / log / outbox payload；（G 当前范围负向）grep `mistake|growth|drill` 在 `backend/internal/resume/` 0 命中 | `test/scenarios/e2e/p0-034-resume-register-and-list/`；trigger 必须包含 `cd backend && go test ./cmd/api -run TestResumeRegisterListHTTPScenario -count=1 -v` |

## Phase 3 + 5: resume.parse async job lifecycle

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.035 | resume.parse async job + outbox event + AI failure / retryable + LLM-derived displayName + deterministic full-resume snapshot | A2 dev stack 拉起；`cmd/api` in-process resume runner kernel 启动并注册 resume.parse handler；AIClient stub 提供 structured-only success / output_invalid / `finish_reason=length` / timeout；长输入正文末尾带唯一 marker；PDF / Markdown / text seed 可读，DOCX 在前置 gate 被拒绝；profile `max_tokens >= 8192`；缺少 live env、runner no-op 或 focused gate skip 必须 fail | （A）runner 消费 structured-only success；（B）output_invalid；（C）`finish_reason=length`；（D）timeout retry；（E）正文提取与长输入尾 marker；（F）DOCX 拒绝；（G）profile budget regression | （A1）完整 AI prompt 包含输入尾 marker；（A2）success 写 ready + structured profile + LLM-derived displayName，`parsed_text_snapshot` 由后端从完整正文确定性构建并包含 marker，模型输出不含 `markdownText`；（A3）typed ai_task_runs 与 ready-only completed outbox 正确且无 PII；（B1）invalid JSON 写 failed `AI_OUTPUT_INVALID` + 完整 snapshot，无 completed outbox；（C1）`finish_reason=length` 在 decode 前写 failed `AI_OUTPUT_INVALID` + 完整 snapshot，无 completed outbox；（D1）timeout retry metadata 正确并可重试到 ready；（E1）PDF / Markdown / text prompt 不含文件名、`%PDF`、二进制片段，且无字符/token 截断；（F1）DOCX 不进入 prompt；（G1）profile budget gate PASS；（H）范围外 tailor mode 负向 0 命中；（I）prompt/model raw output 不进入 log/outbox/audit | `test/scenarios/e2e/p0-035-resume-parse-async-job-lifecycle/`；trigger 必须包含真实 runner 场景、DOCX rejection、profile budget、long-input tail-marker、structured-only 与 finish-reason focused tests |
