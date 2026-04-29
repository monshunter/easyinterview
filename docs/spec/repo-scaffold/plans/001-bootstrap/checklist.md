# Repo Scaffold Bootstrap Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-04-29

**关联计划**: [plan](./plan.md)

## Phase 1: 根目录与配置文件

- [x] 1.1 创建 7 个初始根容器目录（`backend/` / `frontend/` / `openapi/` / `migrations/` / `scripts/` / `test/` / `deploy/`），每个目录补 `README.md`（1 行说明 + owner subspec 链接）；A1 只创建 `test/README.md`，如 `test/scenarios/` 已存在则保留且不初始化 scenarios 框架
- [x] 1.2 写入根 `.editorconfig`（UTF-8 / LF / 末行换行 / Go Tab=4 / 其余 Space=2）
- [x] 1.3 写入根 `.gitignore`（覆盖 Go / Node / Python / IDE / OS / build artifacts）
- [x] 1.4 写入根 `.tool-versions`（声明 `golang` / `nodejs` / `pnpm` / `python` 字段与版本号）
- [x] 1.5 写入根 `README.md`（≤80 行，含 docs / AGENTS / spec INDEX 入口指针）

## Phase 2: 顶层 Makefile 与 git hooks

- [x] 2.1 落地根 `Makefile`：10 个 phony target（`help` / `fmt` / `lint` / `test` / `build` / `dev-up` / `dev-down` / `codegen` / `migrate` / `install-hooks`），按 plan §3.2.1 规则实现占位
- [x] 2.2 创建 `scripts/git-hooks/pre-commit` 与 `scripts/git-hooks/commit-msg`（占位 `exit 0`，可执行权限）
- [x] 2.3 创建 `scripts/bootstrap.sh`（环境自检脚本，可执行权限）
- [x] 2.4 修复 `make install-hooks` 使用 Git 解析 hook 目录，确保普通 clone 与 linked worktree 均可安装 hook

## Phase 3: Verification

- [x] 3.1 在干净仓库跑 `make help` / `fmt` / `lint` / `test` / `build` / `dev-up` / `dev-down` / `codegen` / `migrate`，9 条命令全部 `exit 0`
- [x] 3.2 跑 `make install-hooks`，确认 `.git/hooks/pre-commit` 与 `commit-msg` 为指向 `scripts/git-hooks/` 的符号链接；重复执行幂等
- [x] 3.3 运行共享 `context.yaml` validator 确认 `repo` target 解析通过；运行 `/sync-doc-index --check` 确认 `docs/spec/INDEX.md` 与 `docs/spec/repo-scaffold/plans/INDEX.md` 对 Header 无 drift；不重复修改父 roadmap checklist

## Phase 4: v1.1 root-container remediation

- [x] 4.1 补齐 `config/README.md`，说明 A1 只锁根容器，A4 owns config schema / env 字典 / feature flags，A3/F3 消费 `config/ai-profiles/` 与 prompt/rubric 路径；README 不写 secret 示例值
- [x] 4.2 根 `README.md` 仓库结构表补入 `shared/` 与 `config/` 两行，9 个 A1 根容器均可从项目入口发现
- [x] 4.3 `context.yaml` package discovery 补入 `shared` 与 `config`，并同步 [plans/INDEX.md](../INDEX.md) 为 active 状态
- [x] 4.4 运行 context validator、`/sync-doc-index --check` 与 `git diff --check`，确认本次 artifact remediation 没有 Header / INDEX / whitespace drift
