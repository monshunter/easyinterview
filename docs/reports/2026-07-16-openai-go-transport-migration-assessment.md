# OpenAI-Go Transport Migration 交付复盘报告

> **日期**: 2026-07-16
> **审查人**: Codex

**关联计划**: [AIClient and Profile Bootstrap](../spec/ai-provider-and-model-routing/plans/001-aiclient-and-profile-bootstrap/plan.md)
**关联 Bug**: [BUG-0181](../bugs/BUG-0181.md)

## 1 复盘范围与成功证据

- 本次交付在不部署独立 AI gateway 的前提下，把 `openai_compatible` chat/stream/STT 与 `judge_compatible` chat 的自维护通用 wire 替换为 adapter-private `github.com/openai/openai-go/v3 v3.43.0`。
- `AIClient` public API、provider registry/profile、DeepSeek thinking、工具调用、JSON mode、error/meta/privacy、response cap 与跨 provider fallback 合同保持不变；SDK import 只存在于两个 compatible adapter 与共享 internal helper。
- RED/GREEN contract tests覆盖同 provider 最多两次 retry、profile timeout 总预算、stream cancellation/partial meta、单 SSE event 上限、非 streaming body cap 与 multipart STT。
- 最终验证：full `staticcheck ./...`、`go vet ./...`、根 `make test`、`make build`、AI terminology/profile/config lint、UI Demo/core-loop pruning、context/docs/index/mod/diff gates全部通过。根测试结果为 Python 567 tests / 4481 subtests、Go 全包、frontend 126 files / 1004 tests。
- 原 A3 plan/checklist 已完成 Phase 15 并恢复 `completed`，没有创建平行 plan、gateway service 或业务侧 SDK 依赖。

## 2 会话中的主要阻点/痛点

- 官方 SDK 的默认 retry 与流式读取语义不能只靠替换调用点推断。
  - **证据**：RED tests 证明旧 adapter 遇到首次 503 会直接失败，旧 stream cap 还错误地限制整个响应；GREEN 后才锁定初始请求加两次同 provider retry，以及 per-event 与 total-body 两套边界。
  - **影响**：如果缺少 adapter contract，迁移看似更轻量，却可能改变 timeout、流式错误顺序和内存上限。
- 迁移实现完成后，被 5 条既有基线 finding 阻塞收口。
  - **证据**：full staticcheck 报告 3 条 U1000 与 1 条 S1016；UI pruning 报告 1 条 active residual，且失败文件相对 `main` 均无迁移 diff。
  - **影响**：需要额外区分真正死代码、integration-only helper 与文档 owner 漂移，Phase 15.4 不能直接完成。
- `newTestKernel` 的默认静态分析结果与 integration build-tag consumer 不同。
  - **证据**：repo-wide search 找到唯一 consumer 位于 `//go:build integration` 文件；直接删除会破坏 integration 编译。
  - **影响**：若按 U1000 文本机械删除，会以“修复 lint”名义损坏真实集成测试。

## 3 根因归类

- SDK 行为边界需要项目合同映射，而不是仅替换 transport。
  - **类别**：spec/plan。当前 Phase 15 已补齐 retry、timeout、stream、STT、privacy 与 import boundary，已在本次会话内闭环。
- root `make test` 与 full lint/static-analysis 属于不同证据层，历史基线只在 owner closeout gate 中被同时执行。
  - **类别**：README / spec-plan。不是测试失败，而是本地质量入口对“全仓 staticcheck 是否为统一 gate”的表达不够集中。
- build-tag helper 未与唯一 consumer 共置。
  - **类别**：无需仓库改动。已通过最小代码移动修复，现有 Go build constraint 与 integration compile 足以防回归。
- UI pruning lint 正确拦截了后续文档重新引入的 truth-source 口径。
  - **类别**：无需仓库改动。gate 行为正确，不应增加 suppression。

## 4 对流程资产的改进建议

- 在 `ci-pipeline-baseline/001-local-quality-gates` 下一次原地修订时，明确 full `staticcheck ./...` 与 `golangci-lint run ./...` 的 owner 关系：要么由一个根入口稳定承接，要么禁止业务 plan 把未聚合的全仓检查当作隐含完成条件。
  - **落点**：spec-plan / Makefile owner
  - **优先级**：medium
- 保留 A3 的 SDK import boundary lint 与 mockserver contract tests，不把 `openai-go` 类型、retry 配置或 provider extras上移到 business package。
  - **落点**：ai-provider-and-model-routing spec/plan
  - **优先级**：high（已落实，后续只需防回退）
- 对 build-tag 专用测试 helper，优先与同 tag consumer 共置；只有多个 tag 组合共享时再建立带相同 constraint 的 helper 文件。
  - **落点**：backend testing convention / code review checklist
  - **优先级**：low

## 5 建议优先级与后续动作

- 最高价值动作是保持当前 A3 import-boundary 与 adapter contract 作为回归 gate；这直接保护业务零 SDK、timeout/retry 分层和隐私边界。
- 下一轮工程治理可由 `ci-pipeline-baseline/001-local-quality-gates` owner 审计 root lint 是否应统一承接 full staticcheck，避免不同 plan 各自附带全仓检查造成收口时才发现基线漂移。
- build-tag helper 共置与 UI pruning 无需立即新增治理文件；当前代码布局和现有零残留 lint 已足够，待出现第二次同类问题再上升为仓库规则。
