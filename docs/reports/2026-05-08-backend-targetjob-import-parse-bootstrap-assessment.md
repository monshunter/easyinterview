# Backend TargetJob Import Parse Bootstrap 交付复盘报告

> **日期**: 2026-05-08
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：[backend-targetjob/001-targetjob-import-and-parse-bootstrap](../spec/backend-targetjob/plans/001-targetjob-import-and-parse-bootstrap/plan.md) 的 backend target Phase 6 收口，覆盖 TargetJob import / parse / failure / manual_form BDD handoff。
- 功能证据：新增并通过 `E2E.P0.010`、`E2E.P0.011`、`E2E.P0.012`、`E2E.P0.013` 四组场景资产，结果写入 `.test-output/runs/targetjob-20260508-001/e2e/E2E.P0.010..013/result.json`。
- 测试证据：`go test ./internal/targetjob ./internal/targetjob/urlfetch ./cmd/api -count=1` 通过；`make codegen-check`、`make validate-fixtures`、`make lint-events`、`make lint-config`、`make docs-check`、`git diff --check` 通过。
- 文档证据：plan / checklist / BDD 文档更新到 `v1.2 completed`，`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` zero drift。
- Bug 证据：建档 [BUG-0024](../bugs/BUG-0024.md) 与 [BUG-0025](../bugs/BUG-0025.md)，分别覆盖 generated response mapper 漏映射与 redline 误杀合法 error code。

## 2 会话中的主要阻点/痛点

- API mapper drift 只在 BDD red test 中暴露。
  - **证据**：`TestE2EP0010TextImportParseReady` red 失败于 `detail missing parsed summary provenance: <nil>`；store / pipeline 已写 summary，但 generated response mapper 没承接。
  - **影响**：如果没有端到端 response 断言，frontend handoff 会收到缺少 provenance 的 detail response。
- Privacy redline 与 documented error code 的边界不清。
  - **证据**：`AI_PROVIDER_SECRET_MISSING` 是合法 B1 error code，但 payload redline 因 `provider_secret` substring 拒绝写入 `target.analysis.failed`。
  - **影响**：retryable failure path 的 outbox evidence 会缺失，场景验证无法闭环。
- Active-scope 负向搜索的“0 命中”表述需要可执行 allowlist。
  - **证据**：直接对 `docs/spec/backend-targetjob` 搜索会命中 plan/checklist 自身列出的 forbidden token，以及 `active_scope_negative_test.go` 中的测试 token。
  - **影响**：执行者需要临时判断哪些命中是 gate 自身文字；这会降低 deep reconcile 的确定性。

## 3 根因归类

- API mapper drift。
  - **类别**：spec-plan
  - **根因**：plan 要求 summary provenance 可见，但 checklist 没把 generated API response round-trip 作为 mapper 层强制断言。
- Redline false positive。
  - **类别**：spec-plan
  - **根因**：privacy gate 没明确 structured enum / registry value 与自由文本 redline 的执行顺序。
- Negative-search allowlist 不明确。
  - **类别**：spec-plan
  - **根因**：负向搜索 gate 写了 forbidden token 清单，却没有给出可复制的命令、排除项和 self-reference 处理方式。

## 4 对流程资产的改进建议

- 在后续 backend API plan 的 Phase 6 handoff gate 中增加 “generated response mapper round-trip” 断言。
  - **落点**：spec-plan
  - **优先级**：high
- 在 privacy redline gate 中写明先校验 registry / enum，再扫描自由文本 payload；合法 documented code 不应被 substring redline 误杀。
  - **落点**：spec-plan
  - **优先级**：medium
- 为 active-scope negative search 提供标准命令和 allowlist 规则，排除 gate 自身文字与专用 negative-test fixture。
  - **落点**：spec-plan 或 `.agent-skills/plan-code-review`
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮最值得做：在 frontend-home-job-picks-and-parse 接入前，先跑一次 `/plan-code-review backend-targetjob/001-targetjob-import-and-parse-bootstrap frontend-handoff`，重点核对 generated client 消费字段与后端 response mapper 的 parity。
- 可随后处理：把 BUG-0024 / BUG-0025 的模式提炼进 `docs/bugs/PATTERNS.md`，但应按 `/bug-report` 要求由用户确认后再写。
- 可延后：把 negative-search allowlist 固化到 plan-review / plan-code-review shared helper，避免每个 plan 手写过滤规则。
