# Engineering Roadmap · Decisions（ADR）

> 本目录承载 `engineering-roadmap` subject 的 foundational hard gate ADR 与后续重大架构决策。
>
> ADR 不是独立 subspec，所有 ADR 共享 `engineering-roadmap` 的 owner 与生命周期；通过 `engineering-roadmap/spec.md` §3.2 表回填最终结论后即视为锁定。

## 1 命名规范

- 文件名：`ADR-Q{n}-{kebab-case-topic}.md`，其中 `Q{n}` 与 `engineering-roadmap/spec.md` §3.2 表中的决策 ID 一一对应
- Header 字段顺序固定：`版本 / 状态 / 更新日期`，状态枚举只取 `draft`/`active`/`accepted`/`superseded`/`deprecated`
- 一个 ADR 只承担一个决策，多次修订原地更新版本号；推翻历史决策时新 ADR 显式 `supersedes` 旧 ADR

## 2 文档结构

每份 ADR 必须包含以下小节，缺一不可：

1. **背景** — 决策触发的业务/技术上下文，引用真理源（产品 spec / 技术文档 / 代码现状）
2. **选项与取舍** — 至少两个候选方案，列出 Pros / Cons / 适用条件
3. **决策** — 一句话锁定结论，含具体技术名词与配置边界
4. **影响范围** — 列出受影响的 child subspec / 共享契约 / 运维边界
5. **失效与修订条件** — 明确触发推翻或升级的具体阈值
6. **关联** — 引用对应 spec 章节、相关 ADR、依赖文档

## 3 索引

| ID | 决策项 | 文件 | 状态 | 锁定日期 |
|----|--------|------|------|----------|
| Q-1 | 认证方案 | [ADR-Q1-auth.md](./ADR-Q1-auth.md) | accepted | 2026-04-26 |
| Q-2 | 异步编排 | [ADR-Q2-async-orchestration.md](./ADR-Q2-async-orchestration.md) | accepted | 2026-04-26 |
| Q-3 | 分析平台 | [ADR-Q3-analytics-platform.md](./ADR-Q3-analytics-platform.md) | accepted | 2026-04-26 |
| Q-4 | 云部署目标 | [ADR-Q4-cloud-deploy-target.md](./ADR-Q4-cloud-deploy-target.md) | accepted | 2026-04-26 |
| Q-5 | 隐私节奏 | [ADR-Q5-privacy-cadence.md](./ADR-Q5-privacy-cadence.md) | accepted | 2026-04-26 |
| Q-6 | AI 网关与模型路由 | [ADR-Q6-ai-gateway-and-model-routing.md](./ADR-Q6-ai-gateway-and-model-routing.md) | accepted | 2026-04-26 |
