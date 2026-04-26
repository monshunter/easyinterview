# Repo Scaffold Bootstrap

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-04-26

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [repo-scaffold spec](../../spec.md) §3.1 锁定的 4 项决策落到仓库根：创建 7 个根容器目录、写入根 `Makefile`、`.editorconfig`、`.tool-versions`、`.gitignore`、`scripts/git-hooks/` 占位、`scripts/bootstrap.sh`，并通过本 plan 的 verification phase 在空环境跑通 `make help` / `make fmt` / `make lint` / `make test` / `make build`，证明占位 target 不阻塞其他 child subspec 的早期 import。

本 plan 是 `repo-scaffold` 唯一的 plan；后续如出现需要扩展（新增根目录或 Make target），递增 spec 与本 plan 版本，原地修订，不再开 sibling plan。

## 2 背景

[engineering-roadmap spec §5.1 / §5.7](../../../engineering-roadmap/spec.md#51-layer-a--foundation5-份全部-p0) 指出 A1 必须先于 A2 / A5 / B1 / B2 / B4 等所有依赖根目录的 child 落地，否则 W1 / W2 多个 child 会同时在不同根创建 `backend/`、`frontend/` 雪球。本 plan 的关键产出是「占位 target 全部 exit 0」，给后续 child 一个稳定的 fallback：业务实现还没接入时，CI / 本地命令也能跑通最小验证。

本 plan 不接入任何外部依赖（不下载 docker images、不安装 asdf 工具），全部产出限于仓库内文件与可在干净 shell 环境运行的 Makefile / shell script。

## 3 实施步骤

### Phase 1: 根目录与配置文件

#### 1.1 7 个根容器目录与占位 README

创建 `backend/`、`frontend/`、`openapi/`、`migrations/`、`scripts/`、`test/`、`deploy/` 共 7 个根目录，每个目录写入 `README.md`：1 行说明 + 1 行 owner subspec 链接（指向对应 child spec 占位行）。A1 只创建 `test/README.md`；如 `test/scenarios/` 已存在则必须保留，不在本 plan 中创建或初始化 scenarios 框架。

#### 1.2 根 `.editorconfig` / `.gitignore` / `.tool-versions`

- `.editorconfig`：UTF-8 / LF / 末行换行；`*.go` Tab=4，其余 Space=2。
- `.gitignore`：覆盖 Go / Node / Python / IDE / OS / build artifacts；不忽略 `docs/`、`AGENTS.md`、`README.md`。
- `.tool-versions`：声明字段名 `golang` / `nodejs` / `pnpm` / `python`，具体版本号在执行时按当前主流稳定版填入；版本号必须可被 asdf / mise 解析。

#### 1.3 根 `README.md`

≤80 行，结构固定：项目一段介绍 + `docs/` 索引指针 + `AGENTS.md` 入口 + 顶层 child spec 列表（指向 `docs/spec/INDEX.md`）。不重复 spec 内容。

### Phase 2: 顶层 Makefile 与 git hooks

#### 2.1 根 `Makefile`

锁定 10 个 phony target：`help` / `fmt` / `lint` / `test` / `build` / `dev-up` / `dev-down` / `codegen` / `migrate` / `install-hooks`。`help` 解析 `## ` 注释打印 target 列表；其余 target 占位实现按下列规则：

- `fmt` / `lint` / `test` / `build`：递归调用 `backend/Makefile` 与 `frontend/Makefile` 的同名 target；若子目录尚未提供该 target，打印 `TODO: implemented by <child>` 并 `exit 0`。
- `dev-up` / `dev-down`：打印 `TODO: implemented by A2 local-dev-stack` 并 `exit 0`。
- `codegen`：打印 `TODO: implemented by B2 openapi-v1-contract` 并 `exit 0`。
- `migrate`：打印 `TODO: implemented by B4 db-migrations-baseline` 并 `exit 0`。
- `install-hooks`：把 `scripts/git-hooks/pre-commit` 与 `scripts/git-hooks/commit-msg` 软链到 `.git/hooks/`。

#### 2.2 `scripts/git-hooks/` 占位

落地 `pre-commit`、`commit-msg` 两个 shell 脚本，初始内容为 `#!/usr/bin/env bash` + `exit 0` + `# extended by B1 / A5 in later plans`。`make install-hooks` 通过 `ln -sf` 安装。

#### 2.3 `scripts/bootstrap.sh`

打印当前 shell / OS / Go / Node / Python 版本与 `.tool-versions` 声明值的对照；不强制安装任何工具，只作为 onboarding 自检入口。

### Phase 3: Verification

#### 3.1 占位 target 自检

在干净仓库中依次跑 `make help` / `make fmt` / `make lint` / `make test` / `make build` / `make dev-up` / `make dev-down` / `make codegen` / `make migrate`，每个命令必须 `exit 0` 且无未定义变量错误。

#### 3.2 git hooks 安装自检

在 worktree 内执行 `make install-hooks`，确认 `.git/hooks/pre-commit` 与 `.git/hooks/commit-msg` 出现且为指向 `scripts/git-hooks/` 的符号链接；删除链接后再次跑 `make install-hooks` 必须幂等。

#### 3.3 文档一致性自检

运行共享 `context.yaml` validator，确认本 plan 的 `repo` target 解析到 `spec` / `plan` / `checklist` 三件套；运行 `/sync-doc-index --check`，确认 `docs/spec/INDEX.md` 与 `docs/spec/repo-scaffold/plans/INDEX.md` 对 Header 无 drift。父 roadmap Phase 2 的 spawn / index 收口已经由 `engineering-roadmap/001-decompose-subspecs` owner 完成，本 plan 不再修改父 owner 文档。

## 4 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) C-1 到 C-4 全部成立。
- 本 plan checklist 全部勾选；Phase 3 的 `make` 自检命令日志贴入工作日志。
- engineering-roadmap/001 Phase 2 的 spawn / index 收口已完成；本 plan 只提供 A1 仓库脚手架实现与验证证据，不重复修改父 roadmap checklist。

## 5 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 占位 target 在某个 shell 环境（dash / zsh）行为不一致导致 `exit 0` 不成立 | Phase 3.1 在 macOS zsh 与 Linux bash 各跑一次；若 Makefile 因 BSD/GNU `make` 差异失败，强制 `SHELL := /bin/bash` |
| 工具版本号写死后被 W1+ child 反复 bump | `.tool-versions` 只锁最低版本；任何 child 想升版必须递增本 spec D-2，避免散落在多份 child plan 中 |
| `scripts/git-hooks/` 占位被 B1 / A5 后续 plan 整段重写导致命名漂移 | 文件名（`pre-commit` / `commit-msg`）在本 spec D-4 锁定；后续 child 只能在文件内部追加规则，不得改文件名或新增同名 hook 到其它路径 |
| 根目录 README 占位被遗忘补全成空指针 | Phase 1.1 强制每个 README 必须包含 owner subspec 链接；任何根 README 缺链接由 sync-doc-index 报告 |
