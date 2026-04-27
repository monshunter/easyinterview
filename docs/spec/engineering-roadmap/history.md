# Engineering Roadmap History

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-04-27

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-04-27 | 1.4 | 修正 W1 Phase 3 gate 口径：parent plan 只完成 9 份 child spec 的 cross-spec review 与 spec-contract lock；A2/B2/F1/F3 的可执行 gate 交由各 child `001` plan 逐一验证，未通过前不得启动依赖它的 W2 implementation | 001-decompose-subspecs |
| 2026-04-26 | 1.3 | L2 code review remediation：ADR-Q3 从 PostHog Cloud 切换为自托管 PostHog 优先；补齐 async public `jobType` 与内部 Asynq handler 的命名边界；明确 Q-5 P0 导出延后是产品验收项的 W0 例外 | 001-decompose-subspecs |
| 2026-04-26 | 1.2 | W0 hard gate 6 项 ADR 全部 accepted（Q-1 自建 passwordless / Q-2 Asynq+Redis / Q-3 PostHog Cloud EU / Q-4 Kubernetes / Q-5 P0 仅删除 / Q-6 AIClient+Model Profile+外部 AI Gateway）；§3.2 表替换为锁定结论 + ADR 链接；§4.4 / §5.5 E4 / §6 C-1 / C-6 同步引用 ADR | 001-decompose-subspecs |
| 2026-04-26 | 1.1 | 补充 Q-6 AI 网关与模型路由 W0 决策项；A3 改为 provider-neutral `ai-gateway-and-model-routing`；明确 Higress 等 AI Gateway 作为独立部署组件而非业务 SDK | 001-decompose-subspecs |
| 2026-04-26 | 1.0 | 初始创建：定义 6 层 38 child subspec、6 wave 实施顺序、mock-first 集成策略、5 项 W0 hard gate 决策项 | 001-decompose-subspecs |
