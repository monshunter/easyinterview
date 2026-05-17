# Frontend Resume Workshop History

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-17

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-17 | 1.1 | L1 设计结晶：起草 plan 002 与 plan 003 全套资产（plan.md / checklist.md / bdd-plan.md / bdd-checklist.md / context.yaml）；§3.2 待确认事项中 guided 模式 / accept-reject inline action / 首页 "1 分钟创建" deep link 三项由"默认"升级为"已锁定 + 指向具体 plan"；§6 C-10 / C-11 由"（XXX 范围）"链接到对应 plan；§7 plan 002 / plan 003 由"未创建"升级为 active 计划链接 + BDD 场景编号 `E2E.P0.081-083` (002) / `E2E.P0.084-087` (003)；plan 002 显式 Phase 0 gate 等待 [backend-resume/002 Phase 1](../backend-resume/plans/002-versions-tailor-runs-and-save-v1/plan.md#phase-1-b2-d-18-additive-confirmresumestructuredmaster--b1-错误码增补) `confirmResumeStructuredMaster` 落地；plan 003 显式 Phase 0 gate 等待 backend-resume/002 Phase 4..8 落地及 `acceptResumeTailorSuggestion / rejectResumeTailorSuggestion` fixture 漂移收敛。 | 002-create-flow-and-onboarding, 003-branch-rewrites-and-edit |
| 2026-05-11 | 1.0 | 初始创建：从 engineering-roadmap 3.10 §5.2 派生 `frontend-resume-workshop` subject，作为 `resume_versions` 路由的前端 owner 并接管 frontend-shell PlaceholderScreen；锁定 D-1..D-7 决策（UI 真理源唯一性 / 术语 adapter / 路由参数语义 / mock-first / UI parity gate 强制 / PDF P0 stub / 旧入口负向 grep）；首批 plan `001-listing-routing-and-detail-readonly` 覆盖路由替换 + ResumeListView + ResumeDetailView Preview Tab；后续 002 / 003 列入关联计划但 P0 后启动。 | 001-listing-routing-and-detail-readonly |
