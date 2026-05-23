# Repo Scaffold Spec

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-22

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将历史 A1 `repo-scaffold` 保留为当前 active Foundation spec（无上游依赖）。它是最早落地的基础 spec 之一（与历史 B1 `shared-conventions-codified` 并列），决定了：

- 后端、前端、OpenAPI 契约、DB migrations、共享真理源、配置、运维脚本在仓库中分别落在哪个目录；
- 顶层 `Makefile`、`.editorconfig`、`.tool-versions`、git hooks 提供给所有 child subspec 共享的根入口；
- 后续所有 child（A2–F4）在自己的 plan 中只能在已有的根目录内增减子目录或文件，不得另起平行根。

目标是：

1. **统一根 layout**：在任何后续 workstream 落地业务代码之前，先冻结仓库根目录的命名与边界，避免多个 subject 同时在不同位置创建 `backend/`、`frontend/` 雪球。
2. **提供最小可执行入口**：根 `Makefile` 必须能在空仓库环境下跑通 `help`、`fmt`、`lint`、`test` 等占位 target；具体实现可由后续 child（A2 / A5 / B2）在自己的 plan 中扩展，但根 target 名称在本 spec 锁定。
3. **统一工具链版本声明**：通过 `.tool-versions`（asdf-style）锁定 Go / Node / pnpm / Python 等关键工具的最低版本，避免本地与 CI 漂移。
4. **统一格式化与提交前检查**：`.editorconfig` 锁定缩进 / 换行 / 末行换行；git hooks（pre-commit / commit-msg）提供最小占位脚本，具体规则由 B1 / A5 在后续 plan 中加挂。

本 spec 不写业务代码、不部署服务，也不建立 monorepo 包管理工具（pnpm workspace / Go module 拓扑），那些都归对应 child（A2 `local-dev-stack` / A5 `ci-pipeline-baseline` / B1 `shared-conventions-codified` / B2 `openapi-v1-contract`）。

## 2 范围

### 2.1 In Scope

- 仓库根目录结构：`backend/`、`frontend/`、`openapi/`、`migrations/`、`scripts/`、`test/`、`deploy/`、`shared/`、`config/` 9 个顶层目录的语义边界与 README 占位；A1 只创建 `test/README.md`，不创建 `test/scenarios/` 场景测试框架；`shared/` 与 `config/` 只作为根容器，具体真理源内容由 B1 / A3 / A4 等 owner 增量落地。
- 顶层 `Makefile`：`help` / `fmt` / `lint` / `test` / `build` / `dev-up` 等 phony target 的命名与最小占位实现（占位实现可以仅打印 "TODO: implemented by <child>"）。
- 顶层配置文件：`.editorconfig`、`.tool-versions`、`.gitignore` 的最小内容与编辑约束。
- Git hooks：`scripts/git-hooks/` 目录的占位骨架与 `make install-hooks` target；具体 lint / commit-msg 规则由 B1 / A5 在后续 plan 加挂。
- 顶层 `README.md`：仓库 1 屏说明 + 各 child 索引指针；不重复 spec 内容，仅作为入口。
- `scripts/` 目录的最小 helper：仅包含 `bootstrap.sh`（一键打印当前环境与所需工具版本）的占位实现。

### 2.2 Out of Scope

- `make dev-up` 真正拉起 Postgres / Redis / MinIO 等本地外部依赖：归 [A2 `local-dev-stack`](../engineering-roadmap/spec.md#51-当前已存在的-active-spec)。
- CI 管线（lint / test / build / codegen 工作流）：归 A5 `ci-pipeline-baseline`。
- monorepo 包管理（pnpm workspace 配置、Go module 拓扑）：归 B1 `shared-conventions-codified` 与 A2 `local-dev-stack` 协同。
- OpenAPI codegen 入口、fixtures 拆分：归 B2 `openapi-v1-contract`。
- 本地 Docker Compose 外部依赖栈：归 A2；staging / production / Helm / K8s 等部署资产只有在 release workstream 明确需要时才由 E4 原地设计。
- AI Gateway、secrets 管理：归 A3 / A4。
- `test/scenarios/` 场景测试框架目录与场景资产：归 E2 `e2e-scenarios-p0` 与 `/scenario-*` 流程。
- 业务代码与领域模块目录（`backend/internal/auth/...`、`frontend/src/features/...`）：仅锁定根级容器目录名，业务子目录由对应 child subspec 创建。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 顶层目录命名 | `backend/` / `frontend/` / `openapi/` / `migrations/` / `scripts/` / `test/` / `deploy/` / `shared/` / `config/`；由 engineering roadmap 与 B1/A3/A4 已锁定的共享真理源、配置根容器共同约束 | 所有 child 子目录必须落在这 9 个根容器之内 |
| D-2 | 工具链锁版 | `.tool-versions` 锁定 Go / Node / pnpm / Python 的最低版本，具体版本号由 001-bootstrap plan 在 codebase 实施时确定 | 本地与 CI 走同一套 asdf 兼容版本声明 |
| D-3 | Make target 命名 | `help` / `fmt` / `lint` / `test` / `build` / `dev-up` / `dev-down` / `codegen` / `migrate` / `install-hooks`；占位实现允许打印 "TODO" 由后续 child 接手 | 后续 child plan 不得新增同义 target，必须实现既定 target |
| D-4 | git hooks 落点 | `scripts/git-hooks/`，通过 `make install-hooks` 写入 `.git/hooks/`；不直接版本化 `.git/hooks/` | 兼容 worktree / clone 后再激活 |

### 3.2 待确认事项

- 是否引入 `go.work` 多 module 模式：A1 默认不创建 `go.work` 或 `go.mod`；如后续实现出现需要拆 module（例如把 `migrations/` 独立成 cmd），由 B1 `shared-conventions-codified` 修订后再落地。
- 顶层是否接入 `pre-commit` 框架（python-based）vs 纯 shell hooks：默认纯 shell，B1 / A5 接管时可重审。

## 4 设计约束

### 4.1 结构约束

- 顶层目录数量保持稳定，新增根目录必须先在本 spec §3.1 表中登记。
- 任何 child plan 不得创建 D-1 锁定的 9 个根容器之外的平行业务根目录。
- README 占位采用统一模板：1 行说明 + 1 行 owner subspec 链接，避免空目录。

### 4.2 工具链约束

- `.tool-versions` 由本 spec owner（A1）锁定字段名，具体版本号写入由 001-bootstrap plan 实施时落地。
- `Makefile` 必须自描述：`make help` 输出全部 target 与一行注释。
- 所有占位 target 必须以 exit 0 结束，禁止 `false` 直接失败，否则会阻塞其他 child 早期 import。

### 4.3 文档约束

- 根 `README.md` 长度 ≤ 80 行；详情指向 `docs/`、对应 child spec、`AGENTS.md`。
- `.editorconfig`、`.gitignore` 的修改必须经过本 spec 修订（递增版本 + history 记录）。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 根目录 layout | A1 | 锁定 9 个顶层容器目录 + 根级配置文件 |
| 后端 module 拓扑 | B1 | `go.mod` 名称、internal 包命名 |
| 前端包管理 | A2 + B1 | `package.json` / pnpm workspace |
| 共享真理源根容器 | A1 + B1/B3 | A1 锁 `shared/` 根；B1 owns `shared/conventions.yaml`，B3 owns `shared/events.yaml` / `shared/jobs.yaml` |
| 配置根容器 | A1 + A3/A4/F3 | A1 锁 `config/` 根；A4 owns config schema 与 feature flags，A3 消费 `config/ai-providers.yaml` / `config/ai-profiles.yaml`，F3 消费 plan-defined prompt/rubric 路径 |
| 本地依赖编排 | A2 | docker-compose、`make dev-up` 真正实现 |
| CI 管线 | A5 | lint/test/build/codegen 工作流 |
| OpenAPI 契约 | B2 | `openapi/` 内 fixtures、codegen 入口 |
| DB migrations | B4 | `migrations/` 内迁移文件与工具选型 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 根目录 spawn | 当前 worktree 尚未落地 A1 根容器目录（除 docs/、AGENTS.md、原型和输入资料外没有 backend/ frontend/ 等根目录） | 执行 001-bootstrap plan 全部 checklist | 9 个根容器目录、根 Makefile、`.editorconfig`、`.tool-versions`、`scripts/git-hooks/` 全部存在；`shared/README.md`、`config/README.md`、`test/README.md` 存在且不创建 `test/scenarios/`；`make help` 成功列出所有 target | 001-bootstrap |
| C-2 | 占位 target 不阻塞 | 根 Makefile 已落地 | 在空环境跑 `make fmt` / `make lint` / `make test` / `make build` | 全部 exit 0；缺失工具时打印 "TODO: implemented by <child>" 并以 0 退出 | 001-bootstrap |
| C-3 | git hooks 安装 | 根仓库 clone 后 | 执行 `make install-hooks` | `.git/hooks/pre-commit`、`.git/hooks/commit-msg` 链接到 `scripts/git-hooks/` 下文件；不修改其它 hook | 001-bootstrap |
| C-4 | 工具版本声明 | `.tool-versions` 已落地 | `asdf install`（或等价的 mise / nvm）按文件读取 | Go / Node / pnpm / Python 各能解析出锁定的最低版本 | 001-bootstrap |
| C-5 | 跨 subject 不冲突 | A1 完成后 | A2 / B1 / B2 / A5 等 active subject 各自 plan 进入实施 | 每个 subject 都在 A1 锁定的根目录内增量；不存在重命名根目录或新增平行业务根 | 后续 owner plan |

## 7 关联计划

- [001-bootstrap](./plans/001-bootstrap/plan.md)：在仓库根落地目录骨架、顶层 Makefile、配置文件、git hooks 占位与 `scripts/bootstrap.sh`。

本 spec v1.1 在 A1 001-bootstrap 已完成后把 `shared/` 与 `config/` 纳入根容器；后续若当前 worktree 仍缺 `config/README.md` 或相关 root README，应原地 reopen A1 001-bootstrap 做轻量 artifact remediation，不创建 sibling plan。

后续如需扩展（例如新增根目录或 Makefile target），递增 spec 版本并通过原地修订完成；不创建 sibling spec。
