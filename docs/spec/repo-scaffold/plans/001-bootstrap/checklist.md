# Repo Scaffold Bootstrap Checklist

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-10

**关联计划**: [plan](./plan.md)

## Phase 1: 根目录与配置文件

- [x] 1.1 创建 7 个初始根容器目录（`backend/` / `frontend/` / `openapi/` / `migrations/` / `scripts/` / `test/` / `deploy/`），每个目录补 `README.md`（1 行说明 + owner subspec 链接）；A1 只创建 `test/README.md`，如 `test/scenarios/` 已存在则保留且不初始化 scenarios 框架
- [x] 1.2 写入根 `.editorconfig`（UTF-8 / LF / 末行换行 / Go Tab=4 / 其余 Space=2）
- [x] 1.3 写入根 `.gitignore`（覆盖 Go / Node / Python / IDE / OS / build artifacts）
- [x] 1.4 写入根 `.tool-versions`（声明 `golang` / `nodejs` / `pnpm` / `python` 字段与版本号）
- [x] 1.5 写入根 `README.md`（≤80 行，含 docs / AGENTS / spec INDEX 入口指针）

## Phase 2: 顶层 Makefile 与 git hooks

- [x] 2.1 落地根 `Makefile`：10 个 phony target（`help` / `fmt` / `lint` / `test` / `build` / `dev-up` / `dev-down` / `codegen` / `migrate` / `install-hooks`），按 plan §2.1 规则实现根入口与当前 owner 委托
- [x] 2.2 创建并维护 `scripts/git-hooks/pre-commit` 与 `scripts/git-hooks/commit-msg` 共享入口；当前 `pre-commit` 委托 A4 secret scan，`commit-msg` 执行 ASCII-only message gate，文件均可执行
- [x] 2.3 创建 `scripts/bootstrap.sh`（环境自检脚本，可执行权限）
- [x] 2.4 修复 `make install-hooks` 使用 Git 解析 hook 目录，确保普通 clone 与 linked worktree 均可安装 hook

## Phase 3: Verification

- [x] 3.1 在干净仓库跑 `make help` / `fmt` / `lint` / `test` / `build`，并用 dry-run 验证 `dev-up` / `dev-down` / `codegen` / `migrate` 根委托，不保留空实现委托
- [x] 3.2 跑 `make install-hooks`，确认 `.git/hooks/pre-commit` 与 `commit-msg` 为指向 `scripts/git-hooks/` 的符号链接；重复执行幂等
- [x] 3.3 运行共享 `context.yaml` validator 确认 `repo` target 解析通过；运行 `/sync-doc-index --check` 确认 `docs/spec/INDEX.md` 与 `docs/spec/repo-scaffold/plans/INDEX.md` 对 Header 无 drift；不重复修改父 roadmap checklist

## Phase 4: v1.1 root-container remediation

- [x] 4.1 补齐 `config/README.md`，说明 A1 只锁根容器，A4 owns config schema / env 字典 / feature flags，A3/F3 消费 `config/ai-profiles/` 与 prompt/rubric 路径；README 不写 secret 示例值
- [x] 4.2 根 `README.md` 仓库结构表补入 `shared/` 与 `config/` 两行，9 个 A1 根容器均可从项目入口发现
- [x] 4.3 `context.yaml` package discovery 补入 `shared` 与 `config`，并同步 [plans/INDEX.md](../INDEX.md) 为 active 状态
- [x] 4.4 运行 context validator、`/sync-doc-index --check` 与 `git diff --check`，确认本次 artifact remediation 没有 Header / INDEX / whitespace drift

## Phase 5: root fmt target hardening

- [x] 5.1 删除根 `Makefile` 的 child Makefile 空委托路径：`fmt` 直接执行 backend Go `gofmt`，不新增 frontend formatter 或中转 Makefile；验证: `make -n fmt` 不再输出旧 fmt 空实现文本
- [x] 5.2 应用 `gofmt` 清理当前 backend Go 格式漂移；验证: `gofmt -l $(find backend -type f -name '*.go' | sort)` 无输出
- [x] 5.3 收口文档和执行 gate；验证: `make fmt`、focused Go tests、context validation、`sync-doc-index --check`、`make docs-check` 与 `git diff --check` 通过

## Phase 6: root target delegation wording cleanup

- [x] 6.1 A1 spec / plan / checklist 不再把 `dev-up` / `dev-down`、`codegen`、`migrate` 描述为空实现委托；当前口径为 dev-stack / codegen chain / backend migrate CLI 委托
- [x] 6.2 验证根 target dry-run 与文档 gate；验证: `make -n dev-up`、`make -n dev-down`、`make -n codegen`、`make -n migrate` 均展示真实委托；focused grep 当前 A1 owner 文档无旧空实现 target 口径；context validation、`sync-doc-index --check`、`make docs-check` 与 `git diff --check` 通过

## Phase 7: git hook wording cleanup

- [x] 7.1 A1 spec / plan / checklist 不再把 git hooks 描述为空实现；当前口径为 pre-commit secret scan + commit-msg ASCII gate 的共享入口
- [x] 7.2 验证 hook 文档 gate；验证: focused grep 当前 A1 owner 文档无旧 hook 空实现表述；context validation、`sync-doc-index --check`、`make docs-check` 与 `git diff --check` 通过
