# Repo Scaffold Bootstrap

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-10

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [repo-scaffold spec](../../spec.md) §3.1 锁定的 4 项决策落到仓库根：创建 9 个根容器目录、写入根 `Makefile`、`.editorconfig`、`.tool-versions`、`.gitignore`、`scripts/git-hooks/` 共享入口、`scripts/bootstrap.sh`，并通过本 plan 的 verification phase 在空环境跑通 `make help` / `make fmt` / `make lint` / `make test` / `make build`，证明根 target 不阻塞其他 child subspec 的早期 import。

2026-07-10 原地修订范围：当前 backend Go 代码已存在，`make fmt` 不再保留 child Makefile 空委托路径；根 target 直接执行 `gofmt`，并用 `gofmt -l` / `make fmt` / focused Go test 证明格式化入口真实可用。本修订不引入 frontend formatter 或新依赖。

2026-07-10 追加修订范围：当前 `dev-up` / `dev-down`、`codegen`、`migrate` 均已由对应 owner 提供真实入口；本 plan 删除旧空实现口径，改为记录根 target 委托关系和 dry-run 验证。
2026-07-10 三次追加修订范围：当前 git hook 入口已承接真实规则，`pre-commit` 委托 A4 secret scan，`commit-msg` 执行 ASCII-only message gate；本 plan 删除旧 hook 空实现表述，保留 `make install-hooks` 幂等安装合同。
2026-07-10 四次追加修订范围：已使用的 PDF module 最低要求 Go 1.24.1，而根工具链与 `go.work` 仍声明 1.24.0，且 `go.mod` 的直接依赖分类与 checksum 未 tidy；本 plan 联合 B1 将三处统一为当前仓库工具链 1.24.5，并把版本/tidy drift 接入根 lint 聚合，不改变业务依赖版本或运行行为。

本 plan 是 `repo-scaffold` 唯一的 plan；后续如出现需要扩展（新增根目录或 Make target），递增 spec 与本 plan 版本，原地修订，不再开 sibling plan。

## 2 背景

[engineering-roadmap spec §5.1](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 A1 保留为当前 active Foundation spec；A1 必须先于 A2 / A5 / B1 / B2 / B4 等所有依赖根目录的 subject 落地，否则多个后续 subject 会同时在不同根创建 `backend/`、`frontend/` 雪球。本 plan 的关键产出是稳定的根入口命名与 owner 委托关系；当 child owner 已落地真实命令时，A1 不再保留空实现委托。

本 plan 不接入任何外部依赖（不下载 docker images、不安装 asdf 工具），全部产出限于仓库内文件与可在干净 shell 环境运行的 Makefile / shell script。

## 3 质量门禁分类

- **Plan 类型**: `tooling + repo-foundation + code-internal`。本 plan 修改仓库根容器目录、根 Makefile、git hook 入口、bootstrap shell 脚本和入口 README；不产生用户可见 UI、HTTP API 行为或业务 workflow。
- **TDD 策略**: 历史实现以 checklist 中 `make` / hook / context validator / sync-doc-index 自检项作为 Red-Green-Refactor 断言来源；重进本 plan 时必须通过 `/implement` -> `/tdd` 顺序执行，focused assertions 来源为根 Make target smoke、hook symlink idempotency、bootstrap script smoke 与 context validation。
- **BDD 策略**: BDD 不适用。本 plan 是仓库脚手架和本地工具入口，不引入用户可感知行为；后续 feature plan 维护自身 BDD gate。
- **替代验证 gate**: `make help`、`make fmt`、`gofmt -l`、`make lint-go-mod-tidy`、focused Go tests、`make lint`、`make test`、`make build`、`make -n dev-up`、`make -n dev-down`、`make -n codegen`、`make -n migrate`、`make install-hooks`、context validation、`sync-doc-index --check`、`git diff --check`。

## 4 实施步骤

### Phase 1: 根目录与配置文件

#### 1.1 9 个根容器目录与 README 入口

创建 `backend/`、`frontend/`、`openapi/`、`migrations/`、`scripts/`、`test/`、`deploy/`、`shared/`、`config/` 共 9 个根目录，每个目录写入 `README.md`：1 行说明 + 1 行 owner subspec 链接。A1 只创建 `test/README.md`；如 `test/scenarios/` 已存在则必须保留，不在本 plan 中创建或初始化 scenarios 框架。`shared/` 与 `config/` 只作为根容器：B1/B3 负责 `shared/` 下真理源，A3/A4/F3 负责 `config/` 下 AI profile、配置 schema 与 prompt/rubric 相关路径。

#### 1.2 根 `.editorconfig` / `.gitignore` / `.tool-versions`

- `.editorconfig`：UTF-8 / LF / 末行换行；`*.go` Tab=4，其余 Space=2。
- `.gitignore`：覆盖 Go / Node / Python / IDE / OS / build artifacts；不忽略 `docs/`、`AGENTS.md`、`README.md`。
- `.tool-versions`：声明字段名 `golang` / `nodejs` / `pnpm` / `python`，具体版本号在执行时按当前主流稳定版填入；版本号必须可被 asdf / mise 解析。

#### 1.3 根 `README.md`

≤80 行，结构固定：项目一段介绍 + `docs/` 索引指针 + `AGENTS.md` 入口 + 顶层 child spec 列表（指向 `docs/spec/INDEX.md`）。不重复 spec 内容。

### Phase 2: 顶层 Makefile 与 git hooks

#### 2.1 根 `Makefile`

锁定 10 个 phony target：`help` / `fmt` / `lint` / `test` / `build` / `dev-up` / `dev-down` / `codegen` / `migrate` / `install-hooks`。`help` 解析 `## ` 注释打印 target 列表；当前根实现按下列规则：

- `fmt`：根级直接执行 `gofmt`，格式化 `backend/**/*.go`；前端 formatter 未落地前不新增 frontend 格式化入口。
- `lint` / `test` / `build`：由 A5 聚合层直接调用已落地 backend/frontend/package gates；已落地 sub-target 失败必须穿透退出码。
- `dev-up` / `dev-down`：委托 `deploy/dev-stack` 的 `up` / `down` target；dev-stack 具体 compose 拓扑和健康检查归 A2。
- `codegen`：聚合 `codegen-conventions` / `codegen-events` / `codegen-openapi`，执行 B1/B3/B2 的真实生成链。
- `migrate`：进入 backend migrate CLI help；`migrate-up` / `migrate-down` / `migrate-status` / `migrate-create` / `migrate-check` 由 B4 维护。
- `install-hooks`：把 `scripts/git-hooks/pre-commit` 与 `scripts/git-hooks/commit-msg` 软链到 `.git/hooks/`。

#### 2.2 `scripts/git-hooks/` shared entries

落地并维护 `pre-commit`、`commit-msg` 两个 shell 脚本。`pre-commit` 作为共享入口委托 A4 `pre-commit-secrets.sh`，`commit-msg` 执行英文 / ASCII-only commit message gate；后续 owner 只能在既有入口内追加规则，不得新增平行 hook 路径。`make install-hooks` 通过 `ln -sf` 安装。

#### 2.3 `scripts/bootstrap.sh`

打印当前 shell / OS / Go / Node / Python 版本与 `.tool-versions` 声明值的对照；不强制安装任何工具，只作为 onboarding 自检入口。

### Phase 3: Verification

#### 3.1 根 target 委托自检

在干净仓库中依次跑 `make help` / `make fmt` / `make lint` / `make test` / `make build`；对会启动外部依赖或改写生成物的入口使用 `make -n dev-up` / `make -n dev-down` / `make -n codegen` / `make -n migrate` 验证根委托关系，确认不再输出空实现文本。

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

### Phase 5: root fmt target hardening

#### 5.1 Replace child Makefile fallback

删除根 `Makefile` 中只服务于早期 scaffold 的 `recurse_target` fallback，将 `fmt` target 改为 `find backend -name '*.go' -print0 | xargs -0 gofmt -w`。不新增 frontend formatter，不创建 `backend/Makefile` / `frontend/Makefile` 作为中转层。

#### 5.2 Apply gofmt to existing drift

运行 `gofmt -w` 清理当前 backend Go 文件格式漂移，并用 `gofmt -l` 验证格式化后没有剩余 Go 文件输出。

#### 5.3 Verification and docs sync

执行 `make fmt`、focused Go tests、context validation、`sync-doc-index --check`、`make docs-check` 与 `git diff --check`；focused grep 确认 `recurse_target` 与旧 fmt 空实现文本不再出现在执行面。

### Phase 6: root target delegation wording cleanup

#### 6.1 Make target delegation contract

更新 A1 spec / plan / checklist，使 `dev-up` / `dev-down`、`codegen`、`migrate` 反映当前根 `Makefile` 委托到 dev-stack、真实 codegen 链与 backend migrate CLI 的事实，不再描述为空实现委托。

#### 6.2 Verification

执行 `make -n dev-up`、`make -n dev-down`、`make -n codegen`、`make -n migrate` 作为无副作用委托证明；focused grep 确认当前 A1 owner 文档不再把这些 target 描述为空实现委托；执行 context validation、`sync-doc-index --check`、`make docs-check` 与 `git diff --check`。

### Phase 7: git hook wording cleanup

#### 7.1 Current hook entry contract

更新 A1 spec / plan / checklist，使 `scripts/git-hooks/pre-commit` 与 `scripts/git-hooks/commit-msg` 反映当前真实规则：pre-commit 通过 A4 secret scan，commit-msg 执行 ASCII-only message gate；A1 继续只锁文件名、安装位置和共享入口约束。

#### 7.2 Verification

执行 hook focused grep、context validation、`sync-doc-index --check`、`make docs-check` 与 `git diff --check`，确认当前 A1 owner 文档不再把 git hooks 描述为空实现。

### Phase 8: Go toolchain and module metadata convergence

#### 8.1 Single Go version contract

把 `.tool-versions`、根 `go.work` 与 `backend/go.mod` 的 Go 版本统一为当前仓库实际使用的 `1.24.5`。`go.mod` 不增加第二个 `toolchain` directive；直接 import 的 modules 由标准 tidy 分类，不改 dependency version。

#### 8.2 Tidy drift gate and verification

增加 `lint-go-mod-tidy` 子 target 并纳入根 `lint` 聚合；先在当前 module metadata 上证明 RED，再运行 `go mod tidy -go=1.24.5 -compat=1.24`。验证 focused gate、`go test ./...`、`go build ./cmd/...`、bootstrap version output、context、docs/diff 与 pruning gates。

## 5 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) C-1 到 C-5 全部成立；C-1 的根容器计数以 v1.1 的 9 个目录为准。
- 本 plan checklist 全部勾选；Phase 3 的 `make` 自检命令日志贴入工作日志。
- engineering-roadmap/001 的 roadmap rebaseline / index 收口已完成；本 plan 只提供 A1 仓库脚手架实现与验证证据，不重复修改父 roadmap checklist。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 根 target 委托在某个 shell 环境（dash / zsh）行为不一致导致真实失败被隐藏或误报 | Phase 3.1 / Phase 6.2 使用 dry-run + focused grep 验证委托命令；若 Makefile 因 BSD/GNU `make` 差异失败，强制 `SHELL := /bin/bash` |
| 工具版本号写死后被后续 subject 反复 bump | `.tool-versions` 只锁最低版本；任何 subject 想升版必须递增本 spec D-2，避免散落在多份 plan 中 |
| `scripts/git-hooks/` 共享入口被后续 plan 整段重写导致命名漂移 | 文件名（`pre-commit` / `commit-msg`）在本 spec D-4 锁定；后续 child 只能在文件内部追加规则，不得改文件名或新增同名 hook 到其它路径 |
| 根目录 README 入口被遗忘补全成空指针 | Phase 1.1 强制每个 README 必须包含 owner subspec 链接；任何根 README 缺链接由 sync-doc-index 报告 |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-10 | 1.6 | Align the Go toolchain and module directive at 1.24.5, tidy direct dependencies and checksums, and add a root tidy drift lint gate. | tech-debt pruning |
| 2026-07-10 | 1.5 | 收敛 git hook 当前事实：pre-commit 已委托 A4 secret scan，commit-msg 已执行 ASCII-only message gate，不再描述为空实现。 | tech-debt pruning |
| 2026-07-10 | 1.4 | 将 `dev-up` / `dev-down`、`codegen`、`migrate` 从空实现口径收敛为当前根 target 委托关系，并用 dry-run 验证。 | tech-debt pruning |
| 2026-07-10 | 1.3 | 根 `make fmt` 改为真实 `gofmt` 入口，删除 child Makefile 空委托路径，并清理现有 Go 格式漂移。 | tech-debt pruning |
| 2026-05-04 | 1.2 | L1 plan-review remediation：补齐当前强制的质量门禁分类，并在 checklist 全部完成后将 plan lifecycle 收口为 completed。 | docs-only L1 remediation |
| 2026-04-29 | 1.1 | 原地 reopen A1 001-bootstrap，补齐 v1.1 spec 已锁定的 `shared/` / `config/` 根容器 artifact、根 README 索引与 context discovery；不创建 sibling plan。 | plan-review remediation |
