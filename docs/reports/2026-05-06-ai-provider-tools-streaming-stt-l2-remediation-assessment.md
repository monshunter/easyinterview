# AI Provider Tools Streaming STT L2 Remediation 交付复盘报告

> **日期**: 2026-05-06
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 `$plan-code-review ai-provider-and-model-routing/002-tools-streaming-and-stt --fix` 发现的 L2 runtime drift，覆盖 stub tool replay、stream terminal meta canonical merge、observability ProfileResolver label enrichment。
- 关联记录：[BUG-0012](../bugs/BUG-0012.md)。
- 成功证据：
  - Red：stub provider focused test 修复前失败，证明带 tools payload 时 `ToolCalls` 为空。
  - Red：client stream focused tests 修复前失败，证明 final `done.Meta` 缺少 canonical profile/provider/model/capability fields。
  - Red：observability focused tests 修复前失败，证明可解析 profile 的 failure / stream labels 仍落到 `unknown`。
  - Green：`go test ./internal/ai/aiclient/providers/stub -count=1`、`go test ./internal/ai/aiclient -count=1`、`go test ./internal/ai/aiclient/observability -count=1` 通过。
  - Regression：`go test ./internal/ai/aiclient/... -count=1`、`make lint`、`make test`、`make codegen-check`、`make docs-check`、`git diff --check`、`sync-doc-index --check` 通过。
- 边界说明：本次复盘不宣称 plan 002 全量完成；`checklist.md` 的 6.3 本地部署 + Kind 场景 smoke 仍保持 blocked。

## 2 会话中的主要阻点/痛点

- 历史 gate 对工具调用只证明 payload 可传递，没有证明 stub 会输出结构化 tool calls。
  - **证据**：增强 `TestStubCompleteWithToolsIsDeterministic` 后先失败，返回的 `ToolCalls` 为空。
  - **影响**：依赖 stub 的后续工具链测试会误以为工具调用路径已可执行，实际只覆盖普通文本 completion。

- Stream terminal meta 的完整性没有在 client 统一出口校验。
  - **证据**：新增 partial done meta 测试后，`ModelProfileName`、`Provider`、`ModelID` 等字段缺失。
  - **影响**：流式完成或取消路径的审计、成本、观测数据可能缺失 canonical route 证据。

- Observability label enrichment 的配置存在，但失败与 stream 路径没有统一消费。
  - **证据**：pre-dispatch failure、stream done、stream error focused tests 修复前均无法命中 enriched labels。
  - **影响**：可解析 profile 的失败样本仍可能沉到 `unknown`，削弱告警和 profile 级排障能力。

- 计划仍依赖尚不存在的本地场景 smoke 资产做最终端到端验证。
  - **证据**：`checklist.md` 6.3 记录当前 scenario framework 只有 Planned 索引，缺少可执行 env setup / Kind 部署 / Ready AI tool-stream-STT 场景。
  - **影响**：内部契约已经闭环，但 provider 底座到业务流的 e2e 证据仍需后续 owner 补齐。

## 3 根因归类

- 根因：002 plan 原 gate 对 runtime semantic invariants 的描述不够细，缺少 stub output、client meta merge、metrics label 三个证据面。
  - **类别**：spec-plan

- 根因：observability tests 保留了“unknown label 可接受”的旧期望，没有随 ProfileResolver 语义一起升级。
  - **类别**：spec-plan / test

- 根因：场景测试框架与 AI provider 底座之间缺少 ready-state handoff，导致 6.3 只能保留为 blocked。
  - **类别**：spec-plan / README

## 4 对流程资产的改进建议

- AI provider 类 plan 的 L2 gate 应显式列出三类 runtime proof：provider substitute output、client canonical meta、observability label enrichment。
  - **落点**：相关 spec/plan gate；必要时补充 `plan-code-review` 的 AI provider review checklist
  - **优先级**：high

- 当引入 resolver / enricher 这类“配置存在但运行时可绕过”的对象时，测试必须覆盖 success、pre-dispatch failure、stream done、stream error 四个入口。
  - **落点**：plan checklist 与 focused unit tests
  - **优先级**：high

- 为 6.3 单独建立 scenario owner handoff：先确认 test/scenarios 的 env setup、Kind deploy、AI tool/stream/STT smoke case 是否属于当前 framework 范围，再决定补场景或降级为明确的 contract-only gate。
  - **落点**：scenario plan / AI provider plan handoff / `test/scenarios` README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：继续由 `ai-provider-and-model-routing/002-tools-streaming-and-stt` owner 收口 6.3，或拆给 scenario owner 建立可执行 smoke 资产；否则 plan 002 应保持 active，不应标记为全量 completed。
- 次优先级：把本次新增的 2.4、3.4、5.4 runtime proof 模式复用到后续 AI provider / profile routing plan-code-review，避免再次用结构性 PASS 代替运行时语义证据。
- 可延后：等 scenario framework 进入 Ready/Verified 后，再补真实 provider registry/profile/secret 组合的 privacy grep 与埋点端到端检查。
