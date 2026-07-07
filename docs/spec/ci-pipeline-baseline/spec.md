# Local Quality Gate and Deferred CI Spec

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-07-07

## 1 背景与目标

[engineering-roadmap spec §5.1](../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将原始 A5 `ci-pipeline-baseline` 保留为当前 active Foundation spec（依赖 [A1 `repo-scaffold`](../repo-scaffold/spec.md) 与 [A2 `local-dev-stack`](./../local-dev-stack/spec.md)）。该 subject 名称保留既有占位，但当前项目是个人单人开发者项目，P0 阶段不需要构建 GitHub Actions / GitLab CI 等远端 CI pipeline。

本 spec 在当前阶段只决定：

- 本地开发者手动执行哪些质量门禁命令；
- 哪些 lint / test / codegen / docs check 由本地 Make target 统一暴露；
- 什么时候才重新评估 CI pipeline；
- 如何避免为了未来多人协作提前引入 GitHub Actions、branch protection、artifact、CI secret、nightly job 等维护成本。

目标是：

1. **本地质量门禁先行**：P0 用 `make lint` / `make test` / `make build` / `make docs-check` / codegen drift check 等本地命令保证基本质量；不依赖远端 CI 才能开发。
2. **零 CI 运维负担**：当前不创建 `.github/workflows/*.yml`、不配置 required check、不开 nightly、不给 CI 注入任何业务 secret。
3. **保留未来切换点**：当项目进入多人协作、公开发布、付费用户或需要自动化 release gate 时，再在本 spec 原地修订，创建 CI implementation plan。
4. **与既有脚手架一致**：A1 提供顶层 Make target 占位，B1 / B2 / A4 等 owner 提供具体 lint / codegen / config check；A5 只负责把这些本地命令组织成一组可重复的质量门禁。

本 spec 不实现 release / deploy（归 E4）、不实现 codegen 工具本身（归 [B2 `openapi-v1-contract`](./../openapi-v1-contract/spec.md) 与 B1）、不实现镜像构建脚本细节（归 A1 + A2），也不在当前阶段实现任何远端 CI pipeline。

## 2 范围

### 2.1 In Scope

- **本地质量入口**：统一约定根 Make target：
  - `make lint`：聚合 Go / TS / error-code / config / metrics / log lint（按对应 owner 落地情况逐步接入）。
  - `make test`：聚合 Go / TS 单元测试；AI 单元测试默认走 stub / fixtures，不需要真实 AI provider secret。
  - `make build`：聚合 backend / frontend 构建；尚未落地的组件可先保留清晰占位输出。
  - `make docs-check`：执行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 与轻量链接检查（如 `python3 scripts/lint/check_md_links.py docs`）。
  - `make codegen-check`：执行已落地 generator 的 idempotency / drift check（B1、B2 按各自 plan 接入）。
- **本地输出契约**：每个 target 失败时必须输出 5 行内的人类可读摘要，并保留原始命令日志；不要求生成 HTML artifact。
- **secret 红线**：本地质量门禁不读取 `.env` 中的生产 secret；任何需要真实 provider 的本地部署验证归 A2/A3/A4，不归本地单测 gate。
- **未来 CI 触发条件**：记录何时需要从本地门禁升级为远端 CI。

### 2.2 Out of Scope

- GitHub Actions / GitLab CI / CircleCI workflow：当前不创建 `.github/workflows/{ci,nightly,dependabot}.yml`。
- PR / push required checks、branch protection 自动同步、`gh api` 脚本：单人阶段不做。
- CI artifact 输出（coverage HTML、构建产物上传、OpenAPI diff artifact）：单人阶段不做。
- nightly 定时任务、Dependabot 自动 PR、Codecov、GHCR push、Docker buildx 远端缓存：单人阶段不做。
- 部署到 staging / prod（K8s manifest apply、Helm 升级、SLO 检查）：归 [E4](../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)。
- 场景测试集群拉起：归 `test/scenarios/` 与 scenario skills；A5 不自动触发。
- 性能 / 压测 / 漏洞扫描：归 E4 + F4（P1 或 release 前）。

## 3 用户决策 / 待确认事项

### 3.1 已锁定决策

| ID | 决策 | 锁定值 | 影响 |
|----|------|--------|------|
| D-1 | 当前阶段是否构建 CI pipeline | 否。P0 单人开发阶段只要求本地质量门禁 | 不创建 workflow / required check / branch protection |
| D-2 | 质量门禁触发方式 | 开发者在本机手动执行 `make lint` / `make test` / `make build` / `make docs-check` / `make codegen-check`；pre-commit 可选，不作为当前必需 gate | 保持开发节奏轻量 |
| D-3 | 本地 gate owner | A5 只组织入口；B1/B2/A4/F1 等 owner 提供各自 lint / generator / check 实现 | 避免 A5 变成工具大杂烩 |
| D-4 | 业务 secret | DB / Redis / AI provider / PostHog secrets 不进入任何远端 runner；本地单测默认走 stub / fixtures | 防止过早引入 secret 管理复杂度 |
| D-5 | 远端 CI 升级触发条件 | 满足任一条件才重新评估：第二位长期贡献者加入、公开 release branch、付费用户上线、需要自动发版、回归频率高到本地门禁不足以控制 | CI 在需要时再建 |
| D-6 | 分支保护 | 当前不强制 branch protection；是否用 `dev/main` 线性提交记录由人工执行 | 单人项目避免流程噪声 |
| D-7 | artifact | 当前不上传 artifact；构建产物只保留在本地工作区 | 降低维护成本 |

### 3.2 待确认事项

- 远端 CI 的具体平台默认保留为 GitHub Actions，但只有触发 D-5 条件后才重新锁定。
- 若后续接入远端 CI，是否拆成 `002-remote-ci` plan，还是把本 spec 的首个实现 plan 直接升级为 CI plan：默认原地新增 plan，不改 subject 路径。

## 4 设计约束

### 4.1 本地门禁约束

- 所有本地 gate 必须可在仓库根执行，不要求开发者手动 `cd backend` / `cd frontend`。
- 任一 target 失败时必须返回非 0；跳过尚未落地组件时必须明确输出 `not implemented yet: <owner>`，不能假装通过。
- `make docs-check` 必须至少包含可执行的 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`；Header / INDEX drift 不能靠人工记忆或只写 slash skill 文本。
- `make codegen-check` 只能检查已经存在的 generator；B2 OpenAPI generator 未落地前不得制造失败 gate。

### 4.2 安全与权限约束

- 本地 gate 不读取生产 secret，不向网络上传源代码、coverage 或构建产物。
- AI 单元测试必须走 stub / fixtures；真实 AI provider smoke 属于 A2/A3/A4 本地部署验证，不属于 A5 本地单测 gate。
- 如果未来引入 CI，新增 workflow 前必须先修订本 spec，登记 runner secret 字典与权限边界。

### 4.3 文档约束

- `README.md` 或后续 `docs/development.md` 只记录本地命令，不声称项目已有 CI pipeline。
- 任何新增远端 CI job、required check、branch protection、artifact 或 secret：必须递增本 spec 版本 + history。
- A5 名称保留为 `ci-pipeline-baseline` 仅为避免目录 churn；正文真理源是“当前 deferred CI + local quality gate”。

## 5 模块边界

| 边界 | Owner | 说明 |
|------|-------|------|
| 根 `make lint` / `make test` / `make build` / `make docs-check` / `make codegen-check` 编排 | A5 + A1 | A1 提供 Makefile 结构，A5 约定本地质量入口 |
| 错误码 lint / 共享类型 codegen drift | B1 | A5 只聚合命令，不重写规则 |
| OpenAPI codegen drift | B2 | B2 generator 落地后再接入 `make codegen-check` |
| Config lint | A4 | A5 聚合 `make lint-config` 或等价入口 |
| Metrics / log lint | F1 | F1 helper 落地后再接入 `make lint` |
| 测试镜像缓存 | A2 + `test/scenarios/` | 手动场景测试需要时调用；A5 不自动预热 |
| 远端 CI / branch protection | Future A5 plan | 当前 deferred；触发 D-5 后再新增 plan |
| 部署 CD | E4 | A5 不做 |

## 6 验收标准

| ID | 场景 | Given | When | Then | 对应 Plan |
|----|------|-------|------|------|-----------|
| C-1 | 无远端 CI 文件 | 仓库处于 P0 单人开发阶段 | 检查 `.github/workflows/` | 不存在由 A5 创建的 `ci.yml` / `nightly.yml` / `dependabot.yml`；文档不声称 CI 已启用 | A5 后续 001（如需要） |
| C-2 | 本地 lint gate | 已落地 B1 lint 与后续 owner lint | `make lint` | 聚合已存在 lint；任一失败返回非 0；未落地 lint 明确标记 owner，不假通过 | A5 后续 001（如需要） + B1/A4/F1 |
| C-3 | 本地 test gate | Go / TS 测试已落地 | `make test` | 单元测试在本地运行；AI 单测走 stub / fixtures；不需要 AI provider secret | A5 后续 001（如需要） |
| C-4 | 本地 build gate | backend / frontend 构建入口存在 | `make build` | 已落地组件构建成功；未落地组件输出清晰占位 | A5 后续 001（如需要） |
| C-5 | docs gate | 任意 spec Header 与 INDEX 人为制造 drift | `make docs-check` 或直接执行 `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` | drift 被报告并返回非 0 | A5 后续 001（如需要） |
| C-6 | codegen drift gate | B1/B2 generator 已落地 | `make codegen-check` | 已接入 generator 重跑后无 diff；未落地 generator 不制造失败 | A5 后续 001（如需要） + B1/B2 |
| C-7 | CI deferred guard | 搜索仓库文档 | grep `ci.yml` / `branch protection` / `required check` | 当前文档把这些能力标记为 future / out of scope，不作为 P0 必需项 | 本次 spec 修订 |

## 7 关联计划

A5 当前已有 [001-local-quality-gates](./plans/001-local-quality-gates/plan.md) 作为本地质量门禁聚合 plan。若触发 D-5 需要远端 CI，再在本 spec 原地修订并创建 `002-remote-ci` 或等价新 plan；不得把远端 CI scope 塞回 001。

当前阶段只把 001 用于本地命令聚合，并继续约束其它 subject 不要把远端 CI pipeline 当成 P0 前置条件。
