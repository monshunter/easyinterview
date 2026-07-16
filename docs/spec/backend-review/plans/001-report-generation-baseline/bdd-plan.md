# Grounded Conversation Report BDD Plan

> **版本**: 2.25
> **状态**: active
> **更新日期**: 2026-07-16

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.REPORT.GENERATE.001` | completed session 具有 frozen context；provider output 也可能 invalid/truncated/retryable，且可能同时违反多个语义 family | 生成、repair/retry、持久化、读取或 replay report | 只使用 frozen context；每个可达错误得到明确且合并的修复意图，anchor 只使用可信 user seq allowlist，unknown code 与 marker collision fail closed；合法 direct report 原子持久化，非法/过大输出无 stale-worker/隐私副作用 | `backend/internal/review/conversation_report_test.go` + `report_generation_contract_test.go`，由根 `make test` 承接 |
| `BDD.REPORT.CONVERSATION.API.001` | owned report 已由 reportable completion 创建，状态为 queued/generating/ready/failed，消息可为空或严格有序 | 以 reportId 读取会话记录 | 从现有唯一 report-session 关系返回 closed transcript；空 `messages` 数组仍为 200，跨用户/缺失/empty identity/blank-content/order/role corruption fail closed，且零内部 locator/AI/write/new table | `backend/internal/store/review/report_conversation_test.go` + `backend/internal/api/reports/report_conversation_test.go`，由根 `make test` 承接 |
| `BDD.REPORT.REGENERATE.001` | owned report 为非超限 terminal failed，完成会话 transcript 与 frozen context 仍存在；请求也可能并发、重放或命中非法状态 | 用户携带 Idempotency-Key 请求重新生成 | 同一 report row 原子回到 queued、只创建一个 fresh job、同 key 重放同一响应且 transcript 仍可读；非 failed/active old job/oversize/cross-user typed fail closed 并零重复 job/内容泄漏 | `backend/internal/review/regenerate_report_service_test.go` + `backend/internal/store/review/regenerate_report_test.go` + `backend/internal/api/reports/regenerate_feedback_report_test.go`，由根 `make test` 承接 |

## Real E2E handoff

| ID | Type | Phase | Given | When | Then |
|----|------|-------|-------|------|------|
| E2E.P0.099 | real report/generating/conversation API/UI | 8/12 | shared real frontend/backend/provider, current-run en/zh ready reports and one honest generating resource | browser preserves exact six full-page images and performs Report -> Conversation -> Back while runner binds authenticated conversation API and read-only PostgreSQL evidence | transcript belongs to the same report and ordered DB messages；exact-six report visual contract stays unchanged and no internal IDs enter tracked evidence |

## Evidence boundary

- Report validator、repair/retry、persistence、replay projection、canonical-round overview、配置默认值和 provider/judge reliability 均由 code/integration/eval gate 承接，不包装为 E2E。
- Exact 24/64 boundary belongs to code-level tests；P0.099 only proves that the current legal real content is fully visible at desktop/mobile.
- P0.099 must reach the host-run frontend/backend and current database/API. Fixture transport, route interception, dev mock, jsdom, package test output or provider CLI/eval cannot satisfy it.
- Root `make test` is the independent whole backend/frontend unit regression gate and never an E2E step or PASS marker.
