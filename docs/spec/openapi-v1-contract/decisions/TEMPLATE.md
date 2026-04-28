# OPENAPI-NNN · &lt;short title&gt;

> **ID**: OPENAPI-NNN
> **状态**: draft
> **日期**: 2026-MM-DD

> **强制流程提示**：任何 breaking change（即 `make openapi-diff` 报 `severity: breaking`
> 且不在 [openapi/diff-config.yaml](../../../../openapi/diff-config.yaml) `whitelist` 内）
> 必须先有一份 `状态: accepted` 的 ADR，再修改 [openapi/openapi.yaml](../../../../openapi/openapi.yaml)，
> 再递增 [openapi/baseline/](../../../../openapi/baseline/) 与 [history.md](../history.md)。
> 顺序错位 = 治理事故。

## 1 背景

引用真理源（[spec.md](../spec.md) §章节、[02-api-definition.md](../../../../easyinterview-tech-docs/02-api-definition.md)、上游 ADR 等）说明触发本决策的契约 / 调用方 / 兼容性问题。明确「为什么现状不能继续」。

## 2 决策

一句话锁定结论，含具体技术名词与配置边界（受影响的 path / method / schema /
enum 值 / response status code）。例：

> 在 `POST /api/v1/foo` 中把 `bar` 字段的 type 从 `string` 改为 `integer`，
> 同步在 `Foo` schema 与 fixtures 中迁移；旧客户端按 §4 迁移路径升级。

## 3 影响

| 边界 | 受影响的项 | Owner |
|------|-----------|-------|
| 契约 | OpenAPI path / schema / enum / response | B2 |
| 后端 | generated DTO / handler interface | C 域 |
| 前端 | generated TS client / msw fixtures 消费 | D 域 |
| Mock | E1 mock-server fixtures | E1 |
| 文档 | spec / history / baseline / 02-api-definition | B2 + 文档 owner |

## 4 迁移与回滚

- **迁移路径**：调用方需要改动什么、什么时候改、是否需要灰度。包括前端
  msw、后端 handler、E1 mock-server 各自的承接窗口。
- **回滚条件**：哪些指标 / 日志触发回滚（错误率、调用方反馈、A5 deferred CI 信号），以及如何回滚（`openapi/openapi.yaml` revert、baseline 不递增）。
- **SemVer 升级**：参见 [openapi/baseline/README.md](../../../../openapi/baseline/README.md)
  阈值；本 ADR 是否触发 minor / major 升级，需在此显式声明。

## 5 相关

- [openapi-v1-contract spec.md](../spec.md) §<章节号>
- [openapi-v1-contract/history.md](../history.md)
- 上游 ADR：[ADR-Q?](../../engineering-roadmap/decisions/) （如有）
- 受影响的 child plan：<子 plan 链接>
- 上游业务文档：[02-api-definition.md](../../../../easyinterview-tech-docs/02-api-definition.md) §<章节>

## 6 审计

| 项 | 内容 |
|----|------|
| 提议人 | <github handle / role> |
| Review | <reviewer handles / 决策会议链接> |
| 实施 PR | <PR / commit 链接> |
| `make openapi-diff` 证据 | <粘贴 wrapper JSON 摘要：summary + 命中的 finding> |
| baseline 递增 | `openapi/baseline/openapi-v<MAJOR>.<MINOR>.<PATCH>.yaml` |
| history.md 递增 | `<日期> | <版本> | <一句话描述> | OPENAPI-NNN-<short>` |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-MM-DD | 1.0 | 初稿 | OPENAPI-NNN |
