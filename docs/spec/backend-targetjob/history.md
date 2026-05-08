# Backend TargetJob History

> **版本**: 1.2
> **状态**: active
> **更新日期**: 2026-05-08

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-05-08 | 1.2 | 完成 001 plan 交付：新增 E2E.P0.010 / 011 / 012 / 013 场景资产与脚本证据，补齐 generated TargetJob summary / fitSummary provenance 映射，并修复 `AI_PROVIDER_SECRET_MISSING` 合法 B1 错误码被 payload redline 误杀的问题。 | 001-targetjob-import-and-parse-bootstrap |
| 2026-05-08 | 1.1 | L1 plan-review remediation：按具体场景补齐 B1/B2/B3/F1 owner 契约前置、manual_form terminal job 语义、B3 sourceType 映射、TargetJob 错误码、F1 指标名与 BDD.P0.013 manual_form 场景。 | 001-targetjob-import-and-parse-bootstrap |
| 2026-05-08 | 1.0 | 初始创建：固定 4 个 TargetJob operation 的 backend owner 边界、4 类导入源处理、target_import 异步解析管线、隐私/观测红线、cross-user 隔离、URL fetch SSRF 守护、F3/A3 fail-closed、manual_form 同步路径与 idempotency dedupe；派生 001-targetjob-import-and-parse-bootstrap plan，BDD 占用 E2E.P0.010 / E2E.P0.011 / E2E.P0.012 三个场景 ID（接续 practice-voice-mvp 已预留的 P0.007-P0.009） | 001-targetjob-import-and-parse-bootstrap |
