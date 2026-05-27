# AI Raw Output Debug Switch 交付复盘报告

> **日期**: 2026-05-27
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付新增 `AI_DEBUG_PRINT_RAW_OUTPUT` / `ai.debugPrintRawOutput`；local dev/test 与本地真实联调默认开启，staging/prod 默认关闭。开启时 backend observability 仅对 LLM `Complete` 原始响应写入 stderr 的 `AI_RAW_OUTPUT_DEBUG_*` 调试区块。
- 持久化与常规观测边界保持不变：`ai_task_runs`、`audit_events`、metrics 和 structured log fields 仍只保留 hash、长度和允许字段。
- 已通过验证：
  - `go test ./internal/ai/aiclient/observability -count=1`
  - `go test ./internal/platform/config -count=1`
  - `go test ./cmd/api -run '^$' -count=0`
  - `python3 scripts/lint/env_dict.py --repo-root .`
  - `rg -n "AI_DEBUG_PRINT_RAW_OUTPUT|debugPrintRawOutput|AI_RAW_OUTPUT_DEBUG|WithRawOutputDebugWriter|rawOutputDebug" ...` 确认写入点集中在显式 debug writer、配置字典和文档说明。

## 2 会话中的主要阻点/痛点

- 真实模型 schema 失败时，原有审计只记录 `ResponseHash` / `ResponseCharLength`，无法还原具体非预期格式。
  - **证据**：`E2E.P0.100` 后的 `practice.turn.lightweight_observe` 失败只能定位到 `AI_OUTPUT_INVALID`，没有 raw response object key。
  - **影响**：无法判断是模型偶发格式漂移、prompt/schema 不一致，还是调用层缺少防护。
- env 字典有三方漂移门禁，新增一个调试变量需要同步 code binding、根 `.env.example`、spec §3.1.1 和 dev-stack 示例。
  - **证据**：`scripts/lint/env_dict.py` 会检查 `.env.example`、spec、代码声明和 provider registry。
  - **影响**：直接加 `os.Getenv` 或只改 dev-stack `.env` 会绕过 canonical config，并导致后续 lint 漂移。

## 3 根因归类

- AI 输出隐私边界此前只提供持久化脱敏，没有给本地调试留一个可审计、显式开启、生产默认关闭的 raw response 观察口。
  - **类别**：spec-plan
- local dev-stack README 已强调真实 provider 配置，但没有说明“调试 raw output”这类临时敏感输出应如何打开、在哪里出现、不得进入哪些持久化面。
  - **类别**：README

## 4 对流程资产的改进建议

- 在 AI provider / output schema 后续计划中把“raw output debug switch local 默认开启、staging/prod 默认关闭、仅 stderr、不得进入持久化审计”列为固定 gate。
  - **落点**：spec-plan
  - **优先级**：medium
- 若后续还有更多 schema 调试需求，优先扩展同一个 `AI_DEBUG_*` 调试边界，而不是为单个 E2E 场景新增独立 `.env` 或单独 runtime。
  - **落点**：README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一步优先复跑真实闭环场景时使用本地默认 `AI_DEBUG_PRINT_RAW_OUTPUT=true` 捕获 `practice.turn.lightweight_observe` 的真实返回格式；raw 日志仅保留在本机 `.test-output/`，不得提交到仓库。
- 后续如果仍出现 `AI_OUTPUT_INVALID`，应基于 raw 输出决定是修正 prompt/schema 契约，还是在 provider adapter 层增加更强的 response format / retry 策略。
