# Backend Debrief History

> **版本**: 1.3
> **状态**: active
> **更新日期**: 2026-05-21

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-21 | 1.3 | 登记 backend-jobs-recommendations/001 cross-owner additive：新增 `CountDebriefsForUser(ctx, db, userID) (int, error)` 内部 API（`backend/internal/debrief/count.go`），read-only `SELECT COUNT(*) FROM debriefs WHERE user_id = $1`；cross-user 隔离由 caller userId 保证；不写 audit_events。单元测试 `count_test.go` 覆盖 happy / cross-user / nil-db / empty-userId。 | backend-jobs-recommendations/001-jd-match-real-backend-baseline Phase 0.17 |
| 2026-05-16 | 1.2 | 完成 plan 001 `backend-debrief/001-debrief-record-and-analysis`：落地 createDebrief / getDebrief / suggestDebriefQuestions API、`debrief_generate` worker、debrief store/service/handler、idempotency mismatch 统一错误码、隐私/观测/retry/legacy negative gates 与 E2E.P0.060-064 场景资产；全局 backend、codegen、fixture、events、migration、lint、docs 与 diff gates 通过。 | 001-debrief-record-and-analysis completion |
| 2026-05-16 | 1.1 | Phase 0.6 反查 backend-practice 真理源后修正复盘面试 handoff 口径：`debrief` 是 `PracticeGoal`，不是 `PracticeMode`；mode 继续二值 `assisted` / `strict`，goal 派生能力仍依赖 backend-practice/004 或等价 addendum。 | 001-debrief-record-and-analysis Phase 0.6 |
| 2026-05-16 | 1.0 | 初始创建 Backend Debrief owner spec：承接 engineering-roadmap §5.2 Debrief workstream 的后端域；锁定 18 条决策（D-1~D-18）覆盖 API 契约 / 失败语义 / DB 真理源 / Worker 拓扑 / AI 范围 / P1 字段处理 / 隐私红线 / cross-owner addendum 范围 / 复盘面试 handoff；Operation Matrix 包含 `createDebrief` (既有 B2) + `getDebrief` (既有 B2) + `suggestDebriefQuestions` (Phase 0 新增 B2) + `debrief_generate` worker handler；§6 验收标准 C-1~C-17 覆盖主路径 / IK / validation / draft→completed worker / cross-user / AI failure graceful / retry / 隐私 / 跨域 handoff；派 plan `001-debrief-record-and-analysis` v1.0 active；保留编号建议 `002-debrief-listing-and-update` / `003-debrief-voice-and-stt-integration` / `004-debrief-retention-and-cascade`。Phase 0 cross-owner pre-launch addendum 已授权：B1 新增 `DEBRIEF_NOT_FOUND` / `IDEMPOTENCY_KEY_MISMATCH` 错误码 + `DebriefRoundType` enum + `DebriefQuestionSource` enum，并复用 canonical `AI_*` 失败码；B2 新增 `POST /debriefs/question-suggestions` `suggestDebriefQuestions` operation + 修复 `Debrief.roundType` 引用 + 扩展既有 create/get fixtures；B3 修复 `shared/events.yaml` `debrief.created.roundType: $ref:b1.InterviewerRole` 漂移；F3 新增 `debrief.suggest_questions` feature_key + 基线 prompt v0.1.0。 | 001-debrief-record-and-analysis |
