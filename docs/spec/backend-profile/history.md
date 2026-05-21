# Backend Profile History

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-21

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-21 | 1.0 | 初始创建：从 [engineering-roadmap 3.16 §5.2](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)（App shell + auth + settings workstream）与 [§6.3 S2](../engineering-roadmap/spec.md#63-s2--backend-domain-implementation) 派生 `backend-profile` subject，作为 Candidate Profile 业务域后端 owner；锁定 D-1..D-12 决策（candidate_profile 单例 / UpdateProfileRequest patch 语义 / experience card IK / source_type taxonomy enforcement / privacy 删除顺序 / source counts internal API 等）；首批 plan `001-candidate-profile-and-experience-cards` 覆盖 5 个 Profile endpoint 真实 backend 实现 + cross-owner B2 additive (IK on experience card CUD) + privacy delete + source counts internal API；明确 P0 不实现 AI Insight Cards / 修正覆盖层 / 独立经历库 UI；后续 plan 002 `profile-insights-and-corrections` 列入关联计划但 P0 后启动。 | 001-candidate-profile-and-experience-cards |
| 2026-05-21 | 1.1 | L1 plan-review 收口：新增 D-13 锁定 `GetCandidateProfileForUser(userId) -> *CandidateProfile` cross-owner read-only internal API，由 [backend-jobs-recommendations](../backend-jobs-recommendations/spec.md) `getJobMatchProfile` aggregation 消费；明确不触发 D-1 seed 副作用，缺失返回 nil，调用不写 audit_events / 不 bump profile_version；同步在 §4.4 cross-owner 接口约束、§5 module boundary、§6 验收矩阵新增 C-15 验证场景。修复与 backend-jobs-recommendations/001 plan 中已 declare 的 `backend-profile.GetCandidateProfileForUser` cross-owner internal API 形态在 owner spec 缺位的漂移。 | 001-candidate-profile-and-experience-cards |
