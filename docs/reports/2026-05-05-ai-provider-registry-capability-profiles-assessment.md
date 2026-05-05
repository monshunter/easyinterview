# AI Provider Registry Capability Profiles 交付复盘报告

> **日期**: 2026-05-05
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`ai-provider-and-model-routing/003-provider-registry-and-capability-profiles` 的 backend target，实现 Provider Registry、capability-scoped Model Profile、A3 AIClient 路由/fallback、A4/B1/F3 联动 gate，并完成 Phase 5 lifecycle closeout。
- 代码与文档状态：plan / checklist 已切到 `completed`，`docs/spec/ai-provider-and-model-routing/plans/INDEX.md` 已移入 Completed；`dev` 与 feature branch 均指向 `53ce537 feat(ai-provider): complete provider registry verification`。
- 成功证据：`make lint-config`、`make codegen-check`、`make docs-check`、`make lint`、`make test`、`make build`、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、context validation、`python3 scripts/lint/ai_provider_terminology.py --repo-root .`、`python3 -m pytest scripts/lint/ai_provider_terminology_test.py scripts/lint/env_dict_test.py scripts/lint/ai_profile_coverage_test.py scripts/lint/makefile_dry_run_test.py -q` 均通过。

## 2 会话中的主要阻点/痛点

- Worker entrypoint config test 漏迁移到 registry/profile 路径。
  - **证据**：Phase 5 `make test` 首次失败于 `backend/cmd/worker/main_test.go`，prod config validation 缺少 `AI_PROVIDER_REGISTRY_PATH`，测试仍断言 `ai.providerApiKey`。
  - **影响**：Phase 4 focused config tests 已通过，但全局 entrypoint 测试仍保留旧 key，说明 entrypoint-level config coverage 需要纳入最终迁移核对。
- Active ADR/spec 部署文案仍有旧 endpoint/key 主语。
  - **证据**：Phase 5 active-scope negative search 命中 ADR-Q4、ADR-Q6 与 A3 spec 中以单一 provider endpoint/key 描述部署注入的 active 文案。
  - **影响**：实现代码已切到 provider registry/profile/provider ref，但治理文档仍可能引导后续 owner 回到单一 endpoint 口径，需要在 closeout 前补充 doc reconcile。

## 3 根因归类

- Entry point config 漏迁移属于 `spec-plan` 根因：Phase 4 gate 覆盖 A4 bindings/validator 与 lint，但没有把 `cmd/api` / `cmd/worker` 这类组件入口配置测试列成显式核对项。
- ADR/spec active 文案滞后属于 `spec-plan` 根因：Phase 4 的 docs sync 项列出了 ADR-Q6、A3/A4/F3/roadmap，但 Phase 5 负向搜索才暴露 ADR-Q4 与部署影响范围文本仍需同步。
- 无需修改 `AGENTS.md` 或 skill：本轮既有 Phase 5 global gates 与 negative search 已成功拦截问题，流程缺口更适合固化到后续相关 plan 的 checklist gate。

## 4 对流程资产的改进建议

- 在后续 A3 / A4 config contract 变更 plan 中增加 entrypoint config gate。
  - **落点**：spec-plan
  - **优先级**：high
  - **建议**：当 env/config key 迁移时，checklist 应显式列出 `cmd/api`、`cmd/worker`、local deploy/dev-stack/scenario env 的启动配置测试或负向搜索，而不只覆盖 loader/bindings。
- 将部署 ADR 与 cross-spec active wording 纳入 Phase 4 docs sync 的固定搜索范围。
  - **落点**：spec-plan
  - **优先级**：medium
  - **建议**：涉及 provider、deploy、secrets、observability 等跨域契约时，Phase 4/5 应同时搜索 active ADR、engineering-roadmap spec、affected owner specs 与 plans INDEX，历史 revision rows 可作为例外但 active body 不应残留旧主语。

## 5 建议优先级与后续动作

- 下一轮最值得做的是在 `002-tools-streaming-and-stt` 或 C14 speech adapter plan 的 checklist 中先加入 entrypoint config gate 与 ADR-Q4/Q6 active wording gate，因为这些 workstream 会继续消费 provider registry/profile 契约。
- 可以延后的是把上述经验抽成 skill 级规则；当前已有 `/implement` + `/tdd` + Phase 5 negative search 能拦截此类问题，短期在 owner plan 中固化更直接。
