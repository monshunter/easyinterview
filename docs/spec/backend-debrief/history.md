# Backend Debrief History

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-16

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-16 | 1.0 | 初始创建 Backend Debrief owner spec：承接 engineering-roadmap §5.2 Debrief workstream 的后端域；锁定 18 条决策（D-1~D-18）覆盖 API 契约 / 失败语义 / DB 真理源 / Worker 拓扑 / AI 范围 / P1 字段处理 / 隐私红线 / cross-owner addendum 范围 / 复盘面试 handoff；Operation Matrix 包含 `createDebrief` (既有 B2) + `getDebrief` (既有 B2) + `suggestDebriefQuestions` (Phase 0 新增 B2) + `debrief_generate` worker handler；§6 验收标准 C-1~C-17 覆盖主路径 / IK / validation / draft→completed worker / cross-user / AI failure graceful / retry / 隐私 / 跨域 handoff；派 plan `001-debrief-record-and-analysis` v1.0 active；保留编号建议 `002-debrief-listing-and-update` / `003-debrief-voice-and-stt-integration` / `004-debrief-retention-and-cascade`。Phase 0 cross-owner pre-launch addendum 已授权：B1 新增 `DEBRIEF_NOT_FOUND` 错误码 + `DebriefRoundType` enum + `DebriefQuestionSource` enum；B2 新增 `POST /debriefs/question-suggestions` `suggestDebriefQuestions` operation + 修复 `Debrief.roundType` 引用；B3 修复 `shared/events.yaml` `debrief.created.roundType: $ref:b1.InterviewerRole` 漂移；F3 新增 `debrief.suggest_questions` feature_key + 基线 prompt v0.1.0。 | 001-debrief-record-and-analysis |
