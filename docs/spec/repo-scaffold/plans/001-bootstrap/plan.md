# Repo Scaffold Bootstrap

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-04

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [repo-scaffold spec](../../spec.md) §3.1 锁定的 4 项决策落到仓库根：创建 9 个根容器目录、写入根 `Makefile`、`.editorconfig`、`.tool-versions`、`.gitignore`、`scripts/git-hooks/` 占位、`scripts/bootstrap.sh`，并通过本 plan 的 verification phase 在空环境跑通 `make help` / `make fmt` / `make lint` / `make test` / `make build`，证明占位 target 不阻塞其他 child subspec 的早期 import。

本 plan 是 `repo-scaffold` 唯一的 plan；后续如出现需要扩展（新增根目录或 Make target），递增 spec 与本 plan 版本，原地修订，不再开 sibling plan。

## 2 背景

[engineering-roadmap spec §5.1](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 A1 保留为当前 active Foundation spec；A1 必须先于 A2 / A5 / B1 / B2 / B4 等所有依赖根目录的 subject 落地，否则多个后续 subject 会同时在不同根创建 `backend/`、`frontend/` 雪球。本 plan 的关键产出是「占位 target 全部 exit 0」，给后续 subject 一个稳定的 fallback：业务实现还没接入时，CI / 本地命令也能跑通最小验证。

本 plan 不接入任何外部依赖（不下载 docker images、不安装 asdf 工具），全部产出限于仓库内文件与可在干净 shell 环境运行的 Makefile / shell script。

## 3 质量门禁分类

- **Plan 类型**: `tooling + repo-foundation + code-internal`。本 plan 修改仓库根容器目录、根 Makefile、git hook 占位、bootstrap shell 脚本和入口 README；不产生用户可见 UI、HTTP API 行为或业务 workflow。
- **TDD 策略**: 历史实现以 checklist 中 `make` / hook / context validator / sync-doc-index 自检项作为 Red-Green-Refactor 断言来源；重进本 plan 时必须通过 `/implement` -> `/tdd` 顺序执行，focused assertions 来源为根 Make target smoke、hook symlink idempotency、bootstrap script smoke 与 context validation。
- **BDD 策略**: BDD 不适用。本 plan 是仓库脚手架和本地工具入口，不引入用户可感知行为；后续 feature plan 维护自身 BDD gate。
- **替代验证 gate**: `make help`、`make fmt`、`make lint`、`make test`、`make build`、`make dev-up`、`make dev-down`、`make codegen`、`make migrate`、`make install-hooks`、context validation、`sync-doc-index --check`、`git diff --check`。

## 4 实施步骤

### Phase 1: 根目录与配置文件

#### 1.1 9 个根容器目录与占位 README

创建 `backend/`、`frontend/`、`openapi/`、`migrations/`、`scripts/`、`test/`、`deploy/`、`shared/`、`config/` 共 9 个根目录，每个目录写入 `README.md`：1 行说明 + 1 行 owner subspec 链接（指向对应 child spec 占位行）。A1 只创建 `test/README.md`；如 `test/scenarios/` 已存在则必须保留，不在本 plan 中创建或初始化 scenarios 框架。`shared/` 与 `config/` 只作为根容器：B1/B3 负责 `shared/` 下真理源，A3/A4/F3 负责 `config/` 下 AI profile、配置 schema 与 prompt/rubric 相关路径。

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

### Phase 4: v1.1 root-container remediation

#### 4.1 `config/` 根容器补齐

创建 `config/README.md`，说明 A1 只锁根容器，A4 owns `config/config.yaml`、`config/{dev,staging,prod}.yaml`、`config/feature-flags.yaml` 与 `.env.example` 字典，A3/F3 后续消费 `config/ai-profiles/` 与 prompt/rubric 路径。README 不写任何 secret 示例值。

#### 4.2 根 `README.md` 目录索引补齐

在根 `README.md` 的仓库结构表中补入 `shared/` 与 `config/` 两行，确保 A1 spec D-1 的 9 个根容器都能从项目入口发现。

#### 4.3 context / plans INDEX 同步

把本 plan / checklist / context 版本推进到 1.1，`context.yaml` 的 package discovery 补入 `shared` 与 `config`，并把 [plans/INDEX.md](../INDEX.md) 从 completed 行切回 active，直到 Phase 4 artifact remediation 验证通过。

## 5 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) C-1 到 C-5 全部成立；C-1 的根容器计数以 v1.1 的 9 个目录为准。
- 本 plan checklist 全部勾选；Phase 3 的 `make` 自检命令日志贴入工作日志。
- engineering-roadmap/001 的 roadmap rebaseline / index 收口已完成；本 plan 只提供 A1 仓库脚手架实现与验证证据，不重复修改父 roadmap checklist。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 占位 target 在某个 shell 环境（dash / zsh）行为不一致导致 `exit 0` 不成立 | Phase 3.1 在 macOS zsh 与 Linux bash 各跑一次；若 Makefile 因 BSD/GNU `make` 差异失败，强制 `SHELL := /bin/bash` |
| 工具版本号写死后被后续 subject 反复 bump | `.tool-versions` 只锁最低版本；任何 subject 想升版必须递增本 spec D-2，避免散落在多份 plan 中 |
| `scripts/git-hooks/` 占位被 B1 / A5 后续 plan 整段重写导致命名漂移 | 文件名（`pre-commit` / `commit-msg`）在本 spec D-4 锁定；后续 child 只能在文件内部追加规则，不得改文件名或新增同名 hook 到其它路径 |
| 根目录 README 占位被遗忘补全成空指针 | Phase 1.1 强制每个 README 必须包含 owner subspec 链接；任何根 README 缺链接由 sync-doc-index 报告 |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-05-04 | 1.2 | L1 plan-review remediation：补齐当前强制的质量门禁分类，并在 checklist 全部完成后将 plan lifecycle 收口为 completed。 | historical-spec-implementation-review/001 |
| 2026-04-29 | 1.1 | 原地 reopen A1 001-bootstrap，补齐 v1.1 spec 已锁定的 `shared/` / `config/` 根容器 artifact、根 README 索引与 context discovery；不创建 sibling plan。 | plan-review remediation |
