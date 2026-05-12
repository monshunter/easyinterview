# Backend Resume History

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-12

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-12 | 1.1 | L1 plan-review 修订：补齐首次创建保存边界，明确 `registerResume` / `resume.parse` 只创建 source 与解析草稿，用户 Preview Confirm 前不得创建正式 `structured_master` ResumeVersion；对齐 B2 fixture 场景名和 backend-upload active blocker；收敛 privacy delete 与 B4 matrix 的跨域顺序。 | 001-asset-register-parse-and-listing |
| 2026-05-11 | 1.0 | 初始创建：从 engineering-roadmap 3.10 §5.2 派生 `backend-resume` (C7) subject，作为 Resume 业务域后端 owner；锁定 D-1..D-8 决策（术语映射 / seed_strategy 三路 / suggestion 终态状态机 / parse 路径区分 / tailor mode / RESUME_EXPORT_NOT_AVAILABLE 行为 / pagination / IK 必带）；首批 plan `001-asset-register-parse-and-listing` 覆盖 register / get / list + resume.parse async job + sourceType 三路，并明确 guided answers 持久化到 `resume_assets.guided_answers` jsonb；后续 plan 002 / 003 列入关联计划但 P0 后启动。 | 001-asset-register-parse-and-listing |
