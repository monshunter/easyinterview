# AI Tools, Streaming, and STT Extension Checklist

> **版本**: 0.2
> **状态**: draft
> **更新日期**: 2026-05-05

**关联计划**: [plan](./plan.md)

> 本 plan 处于 draft 状态；任何 phase 仅在对应 trigger 出现并完成 ADR / spec 修订后才可勾选。

## Phase 1: 触发条件复核与 ADR / spec 修订

- [ ] 1.1 在工作日志中归档触发证据（业务 spec id / plan id / 事故记录 / 上游版本号）
- [ ] 1.2 完成 ADR-Q6 修订或新增 supersession ADR（保留零 SDK / 隐私 / 唯一对外能力红线）
- [ ] 1.3 把 spec 版本从当前基线递增到下一版本（当前基线 1.9 时为 2.0+）并同步 history.md
- [ ] 1.4 把本 plan Header 切换为 `状态: active` + `版本: 1.0`，并同步 plans/INDEX.md

## Phase 2: Tools / function calling 实现

- [ ] 2.1 在 spec §4.1 锁定接口形态（独立 `Tools(...)` 或 `Complete` payload 扩展）
- [ ] 2.2 openai_compatible adapter + stub provider 落地 tool 调用与 deterministic 回放
- [ ] 2.3 `AICallMeta` 扩展 tool 相关字段，log / DB 守住 hash / 长度 / profile 红线

## Phase 3: Stream consumer 完整化

- [ ] 3.1 openai_compatible SSE / chunked 解析映射到 plan 001 锁定的 delta / error / done 事件
- [ ] 3.2 context cancellation 路径补齐 partial token meta 与 B1 错误码
- [ ] 3.3 HTTP wire（SSE 或 chunked）选型落地，并把决策写回 spec §3.1

## Phase 4: STT provider adapter

- [ ] 4.1 与 C14 spec 联合锁定 `Transcribe` 入参形态后，更新 spec §4.1
- [ ] 4.2 落地 openai_compatible `/v1/audio/transcriptions` 适配，`capability=stt` 从 unsupported profile 升级为可执行
- [ ] 4.3 校验或扩展 7 个 ai_* metric family 的 label 集合，确保 STT 可观测

## Phase 5: 接入 F1 / F3 / B1

- [ ] 5.1 F1 metric / log / dashboard 字段扩展同步
- [ ] 5.2 F3 profile schema 增量（tools / output_schema / stream_wire）先行落地，再被本 plan 消费
- [ ] 5.3 B1 共享常量 / 错误码扩展先行合入，再在本 plan 引用

## Phase 6: Verification

- [ ] 6.1 spec §6 AC 表为每个被激活 phase 追加 ≥ 1 条 AC（含正常 / 错误 / 隐私 / 观测）
- [ ] 6.2 单测 + 离线契约测试覆盖被激活的 tool / streaming / STT 协议子集
- [ ] 6.3 本地部署 + Kind 场景端到端 smoke 通过，无明文泄漏，埋点齐全
