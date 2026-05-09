# 001 — Plan and Session Orchestration Test Checklist

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-09

**关联 Test Plan**: [test-plan](./test-plan.md)

## Phase 0: 跨 spec 前置修订 + Preflight

- [x] Phase 0 本计划定义的测试项全部通过：shared types unit + drift；openapi-diff + fixture parity；codegen-events + lint-events；migrate up/down + CHECK 约束单元测试；F3 baseline preflight 单元测试；legacy-negative grep（PracticeMode 上下文）

## Phase 1: Plan + Session 主流程 (success path) + idempotency replay 基础

- [x] Phase 1 本计划定义的单元测试项全部通过：idempotency middleware 5 项（pending lock / success replay / per-user 隔离 / TTL expire / domain namespace）；createPracticePlan handler + repository + contract test；getPracticePlan handler（含越权 404）+ contract test；startPracticeSession 集成单元测试（fake F3 + fake AIClient + DB lock 检测断言外部 AI 调用不在 tx 内）+ contract test；getPracticeSession handler（含越权 404）+ contract test；outbox_emitter 序列化单元测试

## Phase 2: 错误路径 + Idempotency 完备性

- [x] Phase 2 本计划定义的单元测试项全部通过：error_mapping 每错误码一例（timeout / invalid output / secret missing / fallback exhausted）；reservation_retry 集成测试（fake AIClient 注入失败 → 重试成功）；conflict body mismatch 单元测试；cross-user 隔离 middleware + repository 集成测试；并发单执行者 goroutine 集成测试；同 plan 多 key 并发 partial UNIQUE INDEX 集成测试

## Phase 3: 观测 / 隐私 / 收尾

- [x] Phase 3 本计划定义的单元测试项全部通过：observability ai_task_runs typed columns 集成测试；audit_events 单元测试；redaction negative-fixture 单元测试；Phase 3 legacy-negative grep（retired 模块术语）
