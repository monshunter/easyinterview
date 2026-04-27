# CI Pipeline Baseline Spec

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-04-27

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-layer-a--foundation5-份全部-p0) 把 A5 `ci-pipeline-baseline` 列为 Layer A · Foundation 的最后一份 child（依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md) 与 [A2 `local-dev-stack`](./../local-dev-stack/spec.md)）。它是 Wave 1 的 9 份契约 / 基础设施 spec 之一，决定了：

- 仓库每一份 PR / push 在 CI 中执行什么 job；
- 哪些 lint / test / codegen 在 CI 阶段强制（与 [B1 `shared-conventions-codified`](../shared-conventions-codified/spec.md) 与 [A4 `secrets-and-config`](./../secrets-and-config/spec.md) 提供的本地 lint 一致）；
- 镜像缓存策略（与 `test/scenarios/` 的 image-cache.sh 同源），不重复发明轮子；
- **本 spec 显式不做 deploy**：CD（staging / prod 部署）归 [E4 `release-gate-and-rollout`](../engineering-roadmap/spec.md#55-layer-e--integration4-份)，A5 只交付能合入的高质量产物（artifact）。

目标是：

1. **每一次 push / PR 都跑相同 gate**：lint → unit test → build → codegen drift check → contract diff（B2 接入后），通过即可合入。
2. **失败可解释**：每个 job 输出结构化 summary（GitHub Actions Job Summary）+ artifact（覆盖率、构建二进制、生成的 OpenAPI / TS Client diff）；定位时间从「翻 5 屏 log」缩到「一屏 summary」。
3. **缓存可控**：模块依赖（Go module / pnpm / pip）、镜像层（buildx）、外部测试镜像（image-cache）三层缓存策略统一；不污染主干分支。
4. **secrets 不出 CI**：CI runner 不持有生产 secret；任何 job 能用的 secret 必须在本 spec 登记（与 A4 字典对齐），新增由本 spec 修订流程控制。

本 spec 不实现 release / deploy（归 E4）、不实现 codegen 工具本身（归 [B2 `openapi-v1-contract`](./../openapi-v1-contract/spec.md) 与 B1）、不实现镜像构建脚本细节（归 A1 + A2）。

## 2 范围

### 2.1 In Scope

- **CI 平台**：GitHub Actions（与现有 git remote / workflows 落点一致）；workflow 落 `.github/workflows/*.yml`，由 A5 owner 维护。
- **Workflow 入口**：
  - `ci.yml`：每次 push / pull_request；包含 lint / test / build / codegen-drift / docs-check 5 个 job。
  - `nightly.yml`：每日 cron；包含完整测试套（含 `test/scenarios/` 场景测试，预留触发，实际执行由 `/scenario-run` 接入）+ 镜像缓存预热。
  - `dependabot.yml`：依赖更新策略（go modules / pnpm / GitHub Actions versions）。
- **必跑 job**：`lint-go` / `lint-ts` / `lint-config`（A4 owner）/ `lint-error-codes`（B1 owner）/ `unit-test-go` / `unit-test-ts` / `build-api` / `build-worker` / `build-frontend` / `codegen-drift-check` / `docs-check`（`/sync-doc-index --check` + 链接检查）。
- **缓存策略**：
  - `actions/setup-go` + module cache（key: `go-${{ hashFiles('**/go.sum') }}`）。
  - `pnpm` cache（key: `pnpm-${{ hashFiles('**/pnpm-lock.yaml') }}`）。
  - Docker buildx layer cache（registry cache or GHA cache）；镜像名空间在 A2 锁定。
- **artifact 输出**：
  - `coverage-go.html` / `coverage-ts.html`。
  - `bin/api`、`bin/worker`、`frontend/dist/` 构建产物（仅 main 分支保留 14 天）。
  - `openapi-diff.html`、`ts-client-diff.txt`（codegen drift 详情）。
- **CI 用 secret 字典**：`GITHUB_TOKEN`（默认）/ `GHCR_TOKEN`（pull 镜像）/ `CODECOV_TOKEN`（可选）；任何业务 secret（DB / Redis / OpenAI）禁止进 CI runner。
- **branch protection 建议**：`main` 强制 PR + ci-required；force push 禁止；同 [AGENTS.md / CLAUDE.md §7](../../../CLAUDE.md#7-git-分支策略) git 分支策略对齐。

### 2.2 Out of Scope

- 部署到 staging / prod（K8s manifest apply、Helm 升级、SLO 检查）：归 [E4](../engineering-roadmap/spec.md#55-layer-e--integration4-份)。
- 镜像 push 到 registry：本 spec 仅锁定 build；push / sign / SBOM 由 E4 在 W4/W5 接入。
- 性能 / 压测 / 漏洞扫描：归 E4 + F4（P1）。
- 场景测试集群拉起：归 `test/scenarios/` 与 [scenario-* skills](../../../.claude/skills/)；A5 仅在 nightly 触发入口。
- E2E 测试（[E2 `e2e-scenarios-p0`](../engineering-roadmap/spec.md#55-layer-e--integration4-份)）：W4 才接入；A5 在本 spec 阶段仅占位 job 名 `e2e`。
- 移动端 / 桌面端构建：当前不在范围。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | CI 平台 | GitHub Actions；workflow 落点 `.github/workflows/`；不引入第二平台（CircleCI / GitLab CI） | 后续脚本一致 |
| D-2 | trigger 矩阵 | `ci.yml`: push（dev / main）+ pull_request；`nightly.yml`: cron `0 18 * * *` UTC；`dependabot.yml`: 月度 | 不在每个 PR 上跑 nightly 重活 |
| D-3 | concurrency cancel | 同一 PR 的旧 run 自动 cancel（`concurrency: pr-${{ github.event.pull_request.number }}`） | 节省 CI 配额 |
| D-4 | matrix 测试 | Go 测试单 OS（`ubuntu-latest`），Node 测试单 OS；macOS / Windows 暂不入门 | 缩短 CI 时间 |
| D-5 | 必跑 job 列表 | 见 §2.1；任一新增由本 spec 修订；删除必须征得对应 owner | 防止 lint / test 漂移 |
| D-6 | codegen drift 校验 | OpenAPI codegen / 共享类型 generator（B1）必须在 CI `git diff --exit-code`；漂移即失败 | 与 B1 D-1 idempotent generator 一致 |
| D-7 | branch protection | `main` 必跑全部必跑 job；`dev` 必跑 lint + test + build；feature 分支不强制 | 保证主干始终绿 |
| D-8 | 不在 CI 注入业务 secret | DB / Redis / OpenAI / PostHog secrets 永不出现在 CI runner；测试默认走 stub / fixtures | 防止 CI 被攻击后泄漏 |
| D-9 | artifact 保留 | PR 分支 7 天；main 分支 14 天；CI 失败时强制保留以便定位 | 默认成本可控 |

### 3.2 待确认事项

- 是否在 PR 阶段强制运行 `test/scenarios/` 中标记 `parallel-safe` 的最小子集：默认不强制，由 nightly 承接；如发现高频回归，再升格。
- pnpm 版本：默认 `latest stable`；具体由 [B1](../shared-conventions-codified/spec.md#31-已锁定决策) D-3 选定。
- OS arm64 vs amd64：默认 amd64；arm64 镜像构建延后。

## 4 设计约束

### 4.1 流水线约束

- 每个 workflow 必须在 `runs-on` 行注释 owner（A5 / B1 / B2 / E4），便于跨 child 修订时定位 PR reviewer。
- 必跑 job 总耗时 P95 ≤ 8 分钟（在 A2 镜像缓存与 module cache 命中前提下）；超出由 owner 优化或拆 job。
- 任一 job 失败：必须在 GitHub Actions Job Summary 输出 5 行内的失败原因摘要，链向完整 log；不允许「自己去翻 log」。

### 4.2 安全与权限约束

- workflow `permissions:` 默认最小（`contents: read`）；需要 write 必须显式声明并由本 spec 修订登记。
- 不使用 `pull_request_target` trigger（避免 fork 注入）；如需对 fork 提供受限 CI，由本 spec 修订决定。
- 第三方 action 钉版本到 commit SHA（不用 `@main`）；新增第三方 action 必须在本 spec 中登记。

### 4.3 文档约束

- `.github/workflows/README.md` 维护当前 workflow 列表 + 必跑 job 矩阵 + 故障排查指引。
- 任何 job 改名 / 删除 / 必跑性变更：递增 spec 版本 + history。
- 与 [01-technical-architecture.md §13](../../../easyinterview-tech-docs/01-technical-architecture.md#13-性能预算建议) 的性能预算对齐：CI 每个 job 时间预算与生产 SLO 不直接挂钩，但 `nightly.yml` 中的场景测试结果应对 SLO baseline 起预警作用。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| `.github/workflows/*.yml` | A5 | workflow 入口与必跑 job 矩阵 |
| `make lint` / `make test` / `make build` 占位 | A1 | A5 在 CI 中调用 A1 锁定的 target |
| 错误码 lint / 共享类型 codegen drift | B1 | A5 接入 B1 提供的本地 lint / generator |
| OpenAPI codegen drift | B2 | A5 接入 B2 提供的 codegen pipeline |
| Config lint | A4 | A5 接入 `make lint-config` |
| 测试镜像缓存 | A2 + `test/scenarios/` | A5 调用既有 image-cache.sh，不重复实现 |
| Branch protection 配置 | A5 | 通过 `gh api` 同步（不依赖 web UI 手动设置） |
| 部署 CD | E4 | A5 不做 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | PR 触发 ci.yml | 仓库 main 已设 branch protection | 任意人提 PR | `ci.yml` 触发 5 个必跑 job；GH UI 显示 `ci/required` checks | A5 后续 001 |
| C-2 | lint 失败拦截 | 故意提交一个 `auth_unauthorized`（小写）错误码 | CI 触发 | `lint-error-codes` 失败；Job Summary 输出违规文件 + 行号 | A5 后续 001 + B1 接入 |
| C-3 | codegen drift 失败 | 故意修改 `openapi/openapi.yaml` 但不同步 generated client | CI | `codegen-drift-check` 失败；artifact `openapi-diff.html` 含 diff | A5 后续 001 + B2 接入 |
| C-4 | docs index drift | 故意修改 `docs/spec/<sub>/spec.md` Header 但不同步 INDEX | CI | `docs-check` 失败；Job Summary 含 `/sync-doc-index --check` 输出 | A5 后续 001 |
| C-5 | 缓存命中 | 二次 PR run（同 lockfile） | CI | Go module cache hit；pnpm cache hit；总耗时较冷启动减半（≥ 50%） | A5 后续 001 |
| C-6 | nightly 触发 | 时间到达 cron | `nightly.yml` 触发 | 完整测试套 + 镜像缓存预热全部跑完；失败时打开 GH issue（标签 `ci-nightly`） | A5 后续 001 |
| C-7 | concurrency cancel | 同 PR 短时间内 push 两次 | CI | 旧 run 被 cancel；新 run 正常完成 | A5 后续 001 |
| C-8 | secret 不泄漏 | grep workflow yaml 与 log | 全部 workflow | 不出现 `OPENAI_API_KEY` / `POSTHOG_PROJECT_API_KEY` 等业务 secret | A5 后续 001 |
| C-9 | artifact 可下载 | CI 跑完成功 / 失败 | GH UI | `coverage-go.html` / `coverage-ts.html` / `bin/api` / `bin/worker` / `frontend/dist/` artifact 可下载；保留时长按 D-9 | A5 后续 001 |
| C-10 | branch protection 同步 | 本 spec 接入 | 跑 `scripts/branch-protection-apply.sh` | `main` 强制 ci-required + 禁止 force push；通过 `gh api` 验证 | A5 后续 001 |

## 7 关联计划

A5 在本次 W1 spec 阶段不创建 impl plan（参见 [001-decompose-subspecs §3.1](../engineering-roadmap/plans/001-decompose-subspecs/plan.md#3-实施步骤)）。后续由 A5 自身的 `001-bootstrap`（W1 末或 W2 初）承接：

- 落地 `.github/workflows/{ci,nightly,dependabot}.yml` 与必跑 job 实现。
- 落地 `scripts/branch-protection-apply.sh`（通过 `gh api` 同步 `main` 保护规则）。
- 提供 `.github/workflows/README.md` 与故障排查指引。

后续如需扩展（性能压测、SBOM、签名）：递增 spec 版本，原地修订；不创建 sibling spec。
