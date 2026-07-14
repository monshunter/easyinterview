# Grounded Conversation Report BDD Plan

> **版本**: 2.20
> **状态**: active
> **更新日期**: 2026-07-14

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.REPORT.GENERATE.001` | completed session 具有 frozen context；provider output 也可能 invalid/truncated/retryable | 生成、repair/retry、持久化、读取或 replay report | 只使用 frozen context；合法 direct report 原子持久化，非法/过大输出 fail closed 且无 stale-worker/隐私副作用 | `backend/internal/review/conversation_report_test.go` + `report_generation_contract_test.go`，由根 `make test` 承接 |

## Real E2E handoff

| ID | Type | Phase | Given | When | Then |
|----|------|-------|-------|------|------|
| E2E.P0.099 | real report/generating API/UI | 8 | shared real frontend/backend/provider, current-run en/zh ready reports and one honest generating resource | browser captures exact six full-page images while runner binds authenticated report API and read-only PostgreSQL evidence | every row binds current DB/API state plus report/session/context/screenshot digests；390x844 ready images show complete action regions and generating images never claim ready |

## Evidence boundary

- Report validator、repair/retry、persistence、replay projection、canonical-round overview、配置默认值和 provider/judge reliability 均由 code/integration/eval gate 承接，不包装为 E2E。
- Exact 24/64 boundary belongs to code-level tests；P0.099 only proves that the current legal real content is fully visible at desktop/mobile.
- P0.099 must reach the host-run frontend/backend and current database/API. Fixture transport, route interception, dev mock, jsdom, package test output or provider CLI/eval cannot satisfy it.
- Root `make test` is the independent whole backend/frontend unit regression gate and never an E2E step or PASS marker.
