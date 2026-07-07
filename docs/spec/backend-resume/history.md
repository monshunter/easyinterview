# Backend Resume History

> **版本**: 1.8
> **状态**: active
> **更新日期**: 2026-07-07

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-07 | 1.8 | 将 `resume.parse` 完成态命名收敛为 LLM-derived `display_name`：parse 成功后从结构化输出派生可识别名称，避免 ready 简历继续显示通用上传 / 粘贴标题。 | 001-asset-register-parse-and-listing |
| 2026-07-06 | 1.6 | 将 active spec 决策、存储边界、验收标准和关联计划收敛为 D-20 flat Resume 合同：9 个 operation、`resumes` 单表、`ai_task_runs` 承载 tailor 输出、`updateResume` / `duplicateResume` 落盘。 | product-scope/001-core-loop-module-pruning |
| 2026-07-06 | 1.5 | 删除已随 JD Match 模块移除的 `CountResumesForUser` cross-owner internal API 模块边界；当前 backend-resume 只承接扁平 Resume 业务域、parse/tailor 编排与删除链路，不再提供 jobs-recommendations aggregation helper。 | product-scope/001-core-loop-module-pruning |
| 2026-05-21 | 1.3 | 登记 backend-jobs-recommendations/001 cross-owner additive：新增 `CountResumesForUser(ctx, db, userID) (int, error)` 内部 API（`backend/internal/resume/count.go`），read-only `SELECT COUNT(*) FROM resume_assets WHERE user_id = $1 AND deleted_at IS NULL`；cross-user 隔离由 caller userId 保证；不写 audit_events，不改 store state。单元测试 `count_test.go` 覆盖 happy / cross-user / nil-db / empty-userId 4 项。模块边界表追加 cross-owner internal API 行。 | backend-jobs-recommendations/001-jd-match-real-backend-baseline Phase 0.11 |
| 2026-05-17 | 1.2 | L1 设计结晶：锁定 D-10 `confirmResumeStructuredMaster` 新 operationId（B2 D-18 additive cross-owner change，新增 `POST /api/v1/resumes/{resumeAssetId}/structured-master` + IK）、D-11 structured_master 唯一性（partial UNIQUE INDEX + handler 双层）、D-12 accept suggestion 不自动改 structured_profile；新增 C-14 confirmResumeStructuredMaster 主路径、C-15 唯一性 409、C-16 `resume.tailor.completed` envelope；§1 op 总数 13 → 14；§7 plan 002 行落地命名 `002-versions-tailor-runs-and-save-v1` 与 7 个 BDD 场景 `E2E.P0.074 – E2E.P0.080`。 | 002-versions-tailor-runs-and-save-v1 |
| 2026-05-12 | 1.1 | L1 plan-review 修订：补齐首次创建保存边界，明确 `registerResume` / `resume.parse` 只创建 source 与解析草稿，用户 Preview Confirm 前不得创建正式 `structured_master` ResumeVersion；对齐 B2 fixture 场景名和 backend-upload active blocker；收敛 privacy delete 与 B4 matrix 的跨域顺序。 | 001-asset-register-parse-and-listing |
| 2026-05-11 | 1.0 | 初始创建：从 engineering-roadmap 3.10 §5.2 派生 `backend-resume` (C7) subject，作为 Resume 业务域后端 owner；锁定 D-1..D-8 决策（术语映射 / seed_strategy 三路 / suggestion 终态状态机 / parse 路径区分 / tailor mode / RESUME_EXPORT_NOT_AVAILABLE 行为 / pagination / IK 必带）；首批 plan `001-asset-register-parse-and-listing` 覆盖 register / get / list + resume.parse async job + sourceType 三路，并明确 guided answers 持久化到 `resume_assets.guided_answers` jsonb；后续 plan 002 / 003 列入关联计划但 P0 后启动。 | 001-asset-register-parse-and-listing |
