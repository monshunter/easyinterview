# easyinterview

围绕具体目标岗位、JD 与真实面试流程设计的 AI 面试训练产品。把「看到岗位 → 准备 → 模拟练习 → 证据化反馈 → 修正表达 → 复练 → 面后复盘」标准化、重复化、可视化，帮助高意图求职者在 24–72 小时内围绕一份真实 JD 完成高质量准备。

P0 主闭环：`JD 导入 → 目标面试规划 → 模拟面试 → 证据化报告 → 题目回顾 / 本轮复练 → 真实面试复盘`。

## 仓库结构

| 目录 | 说明 |
|------|------|
| [`backend/`](./backend/README.md) | Go 后端服务（HTTP API、领域模块、异步运行时） |
| [`frontend/`](./frontend/README.md) | TypeScript / React 前端工程（壳层与 P0 主屏幕） |
| [`openapi/`](./openapi/README.md) | OpenAPI v3 契约与 codegen 输入 |
| [`migrations/`](./migrations/README.md) | PostgreSQL schema 迁移脚本 |
| [`scripts/`](./scripts/README.md) | 跨语言运维脚本与 git hooks 占位 |
| [`test/`](./test/README.md) | 跨服务测试根目录；`test/scenarios/` 承载 BDD / E2E 场景契约与当前 Ready 场景脚本 |
| [`deploy/`](./deploy/README.md) | 本地 Docker Compose 依赖栈与后续部署资产根目录 |
| [`shared/`](./shared/README.md) | 跨语言共享真理源与生成输入（B1/B3 owner） |
| [`config/`](./config/README.md) | 应用配置、feature flags 与 AI profile 根容器（A3/A4/F3 owner） |
| [`docs/`](./docs/README.md) | 项目文档：spec、plan、报告、工作日志、Bug 知识库 |

## Agent 协作

- 根级 Agent 治理：[AGENTS.md](./AGENTS.md)（`CLAUDE.md` / `GEMINI.md` 指向同一文件）
- 共享技能：[`.agent-skills/`](./.agent-skills/)，由 `.claude/` / `.codex/` / `.gemini/` 软链接接入各 Agent 客户端

## 文档索引

- 文档导航：[docs/README.md](./docs/README.md)
- Spec 索引：[docs/spec/INDEX.md](./docs/spec/INDEX.md)
- 顶层规划：[engineering-roadmap](./docs/spec/engineering-roadmap/spec.md)（当前 active spec、P0 workstream 候选与 on-demand child 创建规则）
- 工作日志：[docs/work-journal/INDEX.md](./docs/work-journal/INDEX.md)

## 开发入口

详细开发与质量门禁说明：[docs/development.md](./docs/development.md)。

**5 个本地质量门禁**（A5 单人开发阶段，**无远端 CI pipeline**；远端 CI 升级触发条件见 [A5 spec D-5](./docs/spec/ci-pipeline-baseline/spec.md#31-已锁定决策)）：

| 入口 | 说明 |
|------|------|
| `make lint` | B1 conventions + A4 config + F1 observability (placeholder) + Go/TS lint |
| `make test` | 后端 Go + 前端 TS 单元测试；AI 测试走 stub/fixture，**不读取真实 secret** |
| `make build` | 后端 cmd 二进制 + 前端 bundle |
| `make docs-check` | sync-doc-index Header/INDEX drift + `docs/` 相对链接扫描 |
| `make codegen-check` | B1 conventions + B3 events/jobs + B2 OpenAPI generator drift gate（`git diff --exit-code`） |

其它入口：

| 入口 | 说明 |
|------|------|
| `make help` | 列出全部 phony target |
| `make fmt` | 代码格式化（递归至 `backend/Makefile` / `frontend/Makefile`） |
| `make dev-up` / `make dev-down` | 本地依赖编排，由 [`local-dev-stack`](./docs/spec/local-dev-stack/spec.md) 落地 |
| `make codegen` | 共享约定 / events jobs / OpenAPI 代码生成入口（B1 / B3 / B2 接管） |
| `make migrate` | DB schema 迁移入口（B4 接管） |
| `make install-hooks` | 把 `scripts/git-hooks/` 软链到 `.git/hooks/` |
| `scripts/bootstrap.sh` | 打印当前环境与 `.tool-versions` 声明值，作为开发环境自检 |

工具链版本由 [.tool-versions](./.tool-versions) 锁定（asdf / mise 兼容）。编辑器格式由 [.editorconfig](./.editorconfig) 锁定。

## License

源码与文档许可证待后续 release workstream 按 [engineering-roadmap S3](./docs/spec/engineering-roadmap/spec.md#64-s3--true-integration-and-release-gate) 确认。
