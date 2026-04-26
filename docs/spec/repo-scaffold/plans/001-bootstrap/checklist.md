# Repo Scaffold Bootstrap Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-26

**关联计划**: [plan](./plan.md)

## Phase 1: 根目录与配置文件

- [ ] 1.1 创建 7 个根容器目录（`backend/` / `frontend/` / `openapi/` / `migrations/` / `scripts/` / `test/` / `deploy/`），每个目录补 `README.md`（1 行说明 + owner subspec 链接）；保留现有 `test/scenarios/`
- [ ] 1.2 写入根 `.editorconfig`（UTF-8 / LF / 末行换行 / Go Tab=4 / 其余 Space=2）
- [ ] 1.3 写入根 `.gitignore`（覆盖 Go / Node / Python / IDE / OS / build artifacts）
- [ ] 1.4 写入根 `.tool-versions`（声明 `golang` / `nodejs` / `pnpm` / `python` 字段与版本号）
- [ ] 1.5 写入根 `README.md`（≤80 行，含 docs / AGENTS / spec INDEX 入口指针）

## Phase 2: 顶层 Makefile 与 git hooks

- [ ] 2.1 落地根 `Makefile`：10 个 phony target（`help` / `fmt` / `lint` / `test` / `build` / `dev-up` / `dev-down` / `codegen` / `migrate` / `install-hooks`），按 plan §3.2.1 规则实现占位
- [ ] 2.2 创建 `scripts/git-hooks/pre-commit` 与 `scripts/git-hooks/commit-msg`（占位 `exit 0`，可执行权限）
- [ ] 2.3 创建 `scripts/bootstrap.sh`（环境自检脚本，可执行权限）

## Phase 3: Verification

- [ ] 3.1 在干净仓库跑 `make help` / `fmt` / `lint` / `test` / `build` / `dev-up` / `dev-down` / `codegen` / `migrate`，9 条命令全部 `exit 0`
- [ ] 3.2 跑 `make install-hooks`，确认 `.git/hooks/pre-commit` 与 `commit-msg` 为指向 `scripts/git-hooks/` 的符号链接；重复执行幂等
- [ ] 3.3 更新 `docs/spec/INDEX.md` 中 `repo-scaffold` 行为真实链接 + 真实状态；同步 `engineering-roadmap/001-decompose-subspecs/checklist.md` Phase 2.1 勾选与备注
