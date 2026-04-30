# A/B L2 Code Review Remediation 交付复盘报告

> **日期**: 2026-04-30
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `/plan-code-review` 发现的 A1-A5、B1-B4 实施漂移中的 4 个修复项：A3/B4 `ai_task_runs` row contract、A4 prod/staging config validation、A3 OpenAI-compatible 4xx error-code fallback、A5 `codegen-check` 纳入 B3 event/job drift gate。
- 原地修订并收口 3 个 owner plan：`ai-gateway-and-model-routing/001-aiclient-and-profile-bootstrap` v1.3、`secrets-and-config/001-bootstrap` v1.3、`ci-pipeline-baseline/001-local-quality-gates` v1.4；对应 checklist remediation item 全部勾选，plans INDEX 已恢复 Completed。
- 验证证据：`go test ./... -count=1`、`make test`、`make lint`、`make build`、`make docs-check`、`make codegen-check` 均通过；`make codegen-check` 已实际执行 B3 `codegen-events-check`。

## 2 会话中的主要阻点/痛点

- A3 `task_type` 语义冲突直到 B4 schema 对照时才显现。
  - **证据**：A3 原 `AITaskRunRow.TaskType` 写 `chat` / `embed`，B4 `ai_task_runs.task_type` check 只允许业务任务类型；修复后区分 Model Profile call kind 与 B4 business task type。
  - **影响**：如果真实 PG writer 接入，会导致 `ai_task_runs` 无法持久化，且旧 decorator 会吞掉 writer error。
- A4 validator 的正向测试不是 canonical prod 配置。
  - **证据**：旧 `TestValidateProdAllSecretsPasses` 用手写 partial binding；没有覆盖 database、redis、object storage、PostHog host、email provider 等 required/conditional keys。
  - **影响**：prod/staging 可能使用 `config/config.yaml` 中的本地 dev 默认值而通过启动校验。
- A5 聚合 gate 未随 B3 落地自动扩展。
  - **证据**：`make -n codegen-check` 修复前看不到 `codegen-events-check`；接入后立即发现 A3 新增裸 job literal，说明 B3 gate 原先确实不在标准本地门禁里。
  - **影响**：开发者只运行 A5 标准 gate 时会漏掉 event/job generated drift。

## 3 根因归类

- A3/B4 contract drift。
  - **类别**：spec/plan
  - A3 plan 写了“通过 DI 写入 `ai_task_runs`”，但没有把 B4 required columns、业务 task/resource context、writer error propagation 写成可测试契约。
- A4 prod validation drift。
  - **类别**：spec/plan + test
  - `secrets-and-config` 已有 canonical env dictionary，但 validator 正向测试仍绕过 canonical binding，导致 required key 覆盖面与真实 entrypoint 不一致。
- A5 aggregate gate drift。
  - **类别**：spec/plan
  - B3 子 gate 落地后没有同步进入 A5 aggregate checklist；只有单独运行 `make codegen-events-check` 才能发现问题。

## 4 对流程资产的改进建议

- 在 `/plan-code-review` 的 review checklist 中增加“跨 owner persistence row 必须对照 schema required columns 和 enum check”的显式检查。
  - **落点**：skill
  - **优先级**：medium
- 对涉及 runtime config 的 completed plan，L2 review 应优先检查正向测试是否使用 canonical loader / canonical binding，而不是手写局部 binding。
  - **落点**：skill
  - **优先级**：medium
- A5 后续新增 owner codegen/lint gate 时，应在子 owner close-out checklist 增加“是否需要接入 aggregate gate”的交接项。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高价值后续动作：更新 `/plan-code-review` 的 L2 审查清单，补上 schema required-column 对照和 canonical-binding 正向测试检查。
- 可延后：A5 计划模板或 close-out 检查中加入“新子 gate 是否接入聚合门禁”的提示，避免未来 B/C/D owner 落地后重复漏接。
