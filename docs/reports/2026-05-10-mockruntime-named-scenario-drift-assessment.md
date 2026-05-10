# Mockruntime Named Scenario Drift 交付复盘报告

> **日期**: 2026-05-10
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `internal/api/mockruntime` named scenario 测试复制过期 fixture response expectation，恢复 backend full-suite。
- 成功证据：`go test ./internal/api/mockruntime -run TestHandlerSelectsNamedSeedScenariosAndFailsUnknown -count=1` Red 后 Green；`go test ./internal/api/mockruntime -count=1` PASS；`go test ./...` PASS；`sync-doc-index --check` zero drift。
- 文档证据：`mock-contract-suite/001-fixture-backed-mock-runtime` v1.3 新增并完成 4.5 remediation gate；新增 [BUG-0034](../bugs/BUG-0034.md)。

## 2 会话中的主要阻点/痛点

- Matcher 首次把 `AUTH_UNAUTHORIZED` 字符串关联到 backend-auth。
  - **证据**：`change-intake` matcher 高置信推荐 `backend-auth/001`，但实际失败文件、fixture owner 和历史 mock-runtime 报告都指向 `mock-contract-suite/001`。
  - **影响**：需要额外读取 fixture、mockruntime 和 mock-contract spec 来纠正 owner。
- 测试复制了 fixture response 语义。
  - **证据**：`missing-session` fixture 当前为 `404 PRACTICE_SESSION_NOT_FOUND`，测试仍硬编码 `401 AUTH_UNAUTHORIZED`。
  - **影响**：OpenAPI fixture 语义更新后，backend full-suite 被 stale test consumer 阻断。

## 3 根因归类

- Named scenario 测试把 response truth source 复制进 test table。
  - **类别**：spec-plan
- Matcher 对错误码字符串权重过高，容易把 mock contract drift 路由到 auth owner。
  - **类别**：skill

## 4 对流程资产的改进建议

- Mock runtime / mock transport 测试中，selector 行为可以硬编码，scenario response 语义必须从 fixture 读取。
  - **落点**：spec-plan
  - **优先级**：high
- `change-intake` matcher 后续可降低通用错误码对 owner 的权重，或提高文件路径 / package owner 对 routing 的权重。
  - **落点**：skill
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮最值得做：把“fixture-backed scenario tests read expected response from fixture”作为 mock-contract-suite 后续 review checklist 默认项。
- 可延后：优化 `change-intake` matcher 权重；本次已通过人工证据纠正 owner，未造成错误代码改动。
