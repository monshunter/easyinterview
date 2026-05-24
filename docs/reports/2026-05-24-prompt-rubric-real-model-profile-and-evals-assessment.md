# Real Model Profile and Evals 交付复盘

> **日期**: 2026-05-24
> **审查人**: Claude

## 1 复盘范围与成功证据

范围：实施并验收 `prompt-rubric-registry/004-real-model-profile-and-evals`（F3 真实 LLM Judge、`judge.default` 激活、非 placeholder coverage 门禁、≥50 题离线评估集 + Promptfoo runner），经 `/implement` → `/tdd` 串行执行，feature 分支 `feat/prompt-rubric-registry-004-real-model-profile-and-evals-0524`，6 个 phase commit。

成功证据（全部实际运行）：

- `make test`：backend `go test ./...` 全绿 + frontend 1298 passed。
- `make lint`：golangci-lint `0 issues` + 全部本地 gate（含 `lint-ai-profile-coverage`、`lint-prompts-hardcode`、`lint-getenv-boundary`）OK。
- `make build`：backend `./cmd/...`（含新 `cmd/evalkit`）+ frontend bundle 成功。
- `make eval-offline`：drift-check（52 cases / single-source clean）+ `evalkit run`（offline 52，no network）+ Promptfoo runner `52 passed (100%) / 0 failed`。
- 接口/契约：`TestJudgeSignature` 断言 `Judge` 返回 `[]Score`（D-9 v2.8）；`LLMJudge` 逐维度 + 7 个 fail-close 子测；judge dispatch/adapter/bootstrap focused tests；`ai_profile_coverage` 11 单测（含 4 负向 placeholder/unsupported/missing）。
- 范围断言：active-scope zero-reference grep = 0 matches；`git diff` 证明仅 `judge.default` status 翻 active，13 个 chat 业务 profile status 未动。
- 生命周期：`validate_context` OK、`make docs-check` zero drift、`git diff --check` OK；plan/checklist 置 `completed`，plans INDEX 同步。

## 2 会话中的主要阻点/痛点

1. **基线预存红：spec 版本传播遗漏 Go 测试**。`backend/internal/ai/registry/backend_review_preflight_test.go` 硬编码断言 spec `版本: 2.7`，但前序 doc-only commit（`b2e64ab0` / `f793feef`）已把 spec 升到 v2.9 并同步了 README/Python lint，却未更新该 Go 测试断言，导致 registry 包在 `main` 上即为红。仅在本会话首次跑 `go test` 时才暴露。

2. **Promptfoo 原生依赖在 pnpm v10 下默认不构建**。`promptfoo@0.121.12` 依赖 `better-sqlite3` 原生 binding，pnpm v10 默认拦截 postinstall 构建脚本，`pnpm add` 后 `promptfoo --version` 直接崩溃（bindings 缺失）。需显式声明 `pnpm.onlyBuiltDependencies:[better-sqlite3]` 并 rebuild 才可用。

3. **新 CLI 入口撞上 A4 `os.Getenv` 边界 lint**。`cmd/evalkit` 初版直接读 `os.Getenv`（EVAL_LIVE / provider 配置 / secret），被 `lint-getenv-boundary`（secrets-and-config §4.1，仅允许 `platform/config`、`platform/secrets`、`cmd/api`、`cmd/migrate`）拦截，在 Phase 5 聚合 lint 才发现。

4. **Promptfoo↔Go 单一真理源桥接是真正的架构分叉**。plan §8.3 自身将其列为开放决策；exec-bridge / 预算结果 / JS 重实现三种取舍差异大，需先与用户对齐（已通过 AskUserQuestion 锁定 exec-bridge）才动手，避免大返工。

## 3 根因归类

| 痛点 | 根因 | 症状 vs 缺陷 | 归属 |
|------|------|--------------|------|
| 1 spec 版本传播遗漏 Go 测试 | doc-only spec 版本 bump 未跑下游 Go 测试，版本断言散落在多处（README/Python/Go）且无统一传播 gate | 流程缺陷（同类已在 README/Python 传播过，唯独 Go 漏） | `spec-plan` gate / `skill`（spec 版本 bump 的传播 checklist） |
| 2 Promptfoo 原生构建 | pnpm v10 build-script 审批模型 + 仓库 README 未记录 Node 原生依赖落地约定 | 环境缺陷（首次引入原生 npm 依赖） | `README`（根 / 依赖落地说明） |
| 3 getenv 边界撞线 | 新 CLI 入口未预读 A4 §4.1 env 边界；契约预读聚焦 frontend/backend 契约，未覆盖 secrets-and-config os.Getenv 边界 | 一次性执行疏漏，但属可预防的契约盲点 | `no repo change needed`（已就地用 platform/config 解决，无需改 A4） |
| 4 Promptfoo 桥接架构 | 多组件交互决策，plan 已标记开放 | 正常的架构决策点（已按 §4.1 咨询用户） | `no repo change needed` |

## 4 对流程资产的改进建议

1. **spec 版本 bump 传播 gate（spec-plan / skill）**：在 `/design` 或 spec 版本修订流程中，新增"版本字符串反向搜索"收尾步骤——bump `docs/spec/<subspec>/spec.md` 版本后，必须 grep 全仓（含 `backend/**/*_test.go`、`scripts/**`、README）旧版本号并清零，或显式跑受影响包的测试。本次 README/Python 已传播、Go 漏，说明传播靠人工记忆不可靠。

2. **原生 npm 依赖落地约定（README）**：在根 `README` / `docs/development.md` 记录"引入含原生 binding 的 pnpm 依赖时，必须在 `package.json` 声明 `pnpm.onlyBuiltDependencies` 并验证构建"，避免后续 owner 重复踩 pnpm v10 build-script 审批坑。

3. **契约预读清单补 secrets-and-config 边界（skill / AGENTS.md）**：`/implement` / `/tdd` 的前后端契约预读目前聚焦 development §2 与模块 README；当实施新增 `backend/cmd/*` 入口时，应同时预读 secrets-and-config §4.1 的 `os.Getenv` allowlist，避免在聚合 lint 阶段才发现边界违规。

4. **无需改 A4 allowlist**：evalkit 通过复用 `platform/config`+`platform/secrets` 并把 `EVAL_LIVE` 降级为 `--live` flag，已在不扩大 env 边界的前提下闭合，优于"把新 cmd 加进 allowlist"的跨 owner 改动；该模式可作为后续新 CLI 入口的范式。

## 5 建议优先级与后续动作

| 优先级 | 动作 | 目标资产 |
|--------|------|----------|
| P1 | spec 版本 bump 增加"旧版本号全仓 grep 清零 + 受影响包测试"收尾步骤 | `/design` skill 或 spec 修订 gate |
| P2 | 记录原生 npm 依赖 `onlyBuiltDependencies` 落地约定 | 根 README / `docs/development.md` |
| P3 | 契约预读清单补 secrets-and-config §4.1 os.Getenv 边界（命中 `cmd/*` 新入口时） | `/implement` / `/tdd` skill 或 AGENTS.md §2.1.3 |
| P3 | 将"复用 platform/config + flag 降级 env"沉淀为新 CLI 入口范式 | `docs/development.md` |

后续派生：`005-grayscale-and-quality-feedback`（PostHog 灰度分桶 + 报告页质量主观评分回流）由 spec §7 在本评估闭环验证后承接；EVAL_LIVE live 回归与录制 fixture 更新节奏（plan §8.3）建议在首次接入真实 judge provider 时确认。

> 备注：痛点 1（stale 版本断言）已在本会话 Phase 0 就地修复（commit `79dc0609`），其性质为前序 doc-only commit 的传播遗漏而非新缺陷，未单独建 `/bug-report`；若后续再现同类版本传播遗漏，建议按 `/bug-report` 立 PATTERN 记录。
