# Backend Resume Asset Register L2 Remediation 交付复盘报告

> **日期**: 2026-05-13
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-resume/001-asset-register-parse-and-listing` 的 L2 code review remediation，覆盖 register/list validation error mapping、`resume.parse` retryable failure 状态落库、failed retry state transition、P0.034/P0.035 scenario gate hardening。
- 关联 Bug：[BUG-0052](../bugs/BUG-0052.md)。
- 成功证据：`go test ./internal/resume/... ./cmd/api` PASS；P0.034 与 P0.035 均完成 `setup -> trigger -> verify -> cleanup`；`make docs-check`、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、`git diff --check` PASS。

## 2 会话中的主要阻点/痛点

- Validation / retry negative path 没有被原 P0 gate 锁住。
  - **证据**：review 阶段发现 `RegisterResume` / `ListResumes` 对 typed validation error 返回 500，`resume.parse` retryable timeout 未写 failed/error_code，但既有 handler fixture parity 和 cmd/api happy path 仍可通过。
  - **影响**：completed plan 需要重新打开 Phase 6 修复，增加 handler、job、store、cmd/api 和 scenario script 的补充验证。

- Parse 状态机需要同时审查业务状态和 async retry outcome。
  - **证据**：原实现用 retryable outcome 判断是否写 failed，导致用户可见 `parse_status` 停留在 processing；同时 store 只允许 queued -> processing，无法表达 failed retry。
  - **影响**：单看 async retry metadata 会误判完成，需要跨 `resume_assets`、`async_jobs` 和 drainer scenario 一起验证。

## 3 根因归类

- 缺少计划级 negative gate：原 P0.034/P0.035 的 scenario verify 没有 grep validation/retry test names，也没有 cmd/api negative scenario。
  - **类别**：spec-plan

- Handler error taxonomy 没有集中映射：register/list handler 各自 fallback 到 500，typed error 没有统一 HTTP status 边界。
  - **类别**：no repo change needed

- Retry state semantics 没有在 checklist 中写成可执行断言：计划要求没有 `failed_retryable`，但未把 "retryable failure 仍写 failed/error_code，再由 async_jobs 表达 retry" 固化成测试项。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 后续 backend-resume/002/003 plan 中，凡涉及用户可见状态列与 async retry metadata，应把两者分开列出，并要求 failed/retry/success 的 cmd-api 场景。
  - **落点**：spec-plan
  - **优先级**：high

- 对 completed plan 做 L2 remediation 时，若新增红灯测试，应同步加固对应 scenario `trigger.sh` / `verify.sh` 的测试名断言，避免只在临时 focused test 中变绿。
  - **落点**：spec-plan
  - **优先级**：medium

- 暂不修改 AGENTS.md 或 plan-code-review skill：本次 skill 已正确触发 deep review，缺口主要在该 plan 的原始 gate 粒度。
  - **落点**：no repo change needed
  - **优先级**：low

## 5 建议优先级与后续动作

- 下一轮最值得做的是审查 `backend-resume/002`：重点确认 Preview Confirm / version persistence 是否同样区分用户可见状态、async/outbox metadata 和 scenario negative gates。
- 可延后的是将这类 gate 模板抽象到 skill；先通过具体 backend-resume 后续 plan 验证模式是否复用稳定。
