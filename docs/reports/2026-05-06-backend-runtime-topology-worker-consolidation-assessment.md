# Backend Runtime Topology Worker Consolidation 交付复盘报告

> **日期**: 2026-05-06
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-runtime-topology/001-worker-consolidation`，创建 P0 runtime topology owner spec/plan，删除独立 `backend/cmd/worker` 入口与 `WORKER_LISTEN_ADDR` / `worker.listenAddr` 配置，B3 producer 从 `worker` 改为 `backend_async`，并把开发期观测消费端从默认研发 gate 中移除。
- 成功证据：
  - `make lint-config`
  - `make lint-events`
  - `python3 -m unittest scripts.lint.events_inventory_test`
  - `cd backend && go test ./internal/platform/config ./internal/shared/events -count=1`
  - `pnpm --filter @easyinterview/frontend test -- src/lib/events/envelope.test.ts`
  - `make codegen-events-check`
  - `make lint`
  - `make test`
  - `make codegen-check`
  - `make docs-check`
  - `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-runtime-topology/plans/001-worker-consolidation/context.yaml --docs-root docs --target repo`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `git diff --check` 与 `git diff --cached --check`

## 2 会话中的主要阻点/痛点

- 初始执行顺序先进入代码 Red tests，随后被用户纠正为必须先创建 spec + plan 记录。
  - **证据**：本会话先新增 `validator_test.go` 与 `events_inventory_test.py` 的 Red 期望，用户随后明确指出“这个重构也是需要创建spec + plan 记录并跟进处理的”。
  - **影响**：架构拓扑调整本应先确立 owner spec/plan，再进入 TDD；中途切换增加了一次计划重排成本。
- `codegen-events-check` 对正在实施的真理源变更依赖 git index 状态。
  - **证据**：`make codegen-events-check` 首轮失败，输出的 diff 正是本次期望的 `worker` → `backend_async` 变更；暂存 B3 truth source 与生成物后同一 gate 通过。
  - **影响**：实现者需要理解该 gate 是“当前 index 与生成结果无漂移”，否则容易把期望变更误判成 codegen 失败。
- 文档 heading 重命名造成旧 fragment 失效。
  - **证据**：`make docs-check` 首轮失败，发现 B3 `DB/C8 canonical...` 与 A4 env 字典 `25 项` anchor 已因本次改名失效。
  - **影响**：虽被 docs gate 拦住并修复，但说明跨 completed plan 的旧链接会被 active spec heading 调整影响。

## 3 根因归类

- 架构级重构入口需要更明确的设计先行判定。
  - **类别**：spec-plan
  - 该任务改变 runtime topology、config dictionary、event producer contract 与 dev dependency boundary，不只是 backend-auth 的局部实现细节。
- codegen drift gate 的 index 语义没有在计划 gate 中显式说明。
  - **类别**：README
  - 当前 gate 本身可用，但计划中只写“运行 `make codegen-events-check`”，没有说明 intended generator diff 需要先进入 index 才能验证漂移。
- heading anchor 变化会影响 completed plan 文档链接。
  - **类别**：spec-plan
  - 本次重命名 active spec 小节是正确语义调整，但旧 completed plan 仍引用具体 fragment；需要在标题重命名时把 fragment 检查视为必跑 gate。

## 4 对流程资产的改进建议

- 后续涉及进程拓扑、配置字典、公共契约或开发依赖边界的实施入口，先创建或激活 owner spec/plan，再开始 TDD。
  - **落点**：spec-plan
  - **优先级**：high
- 在 B3 或 A5 codegen gate 文档中补一句：实施期真理源变更需要把 truth source 与生成物暂存到 index 后再运行 `make codegen-*-check`，或提供显式 `--against-working-tree` 检查模式。
  - **落点**：README
  - **优先级**：medium
- 在修改 active spec heading 时，把 `make docs-check` 放到同阶段必跑 gate，并优先修复 completed plan / history 中的 fragment 链接。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 优先：后续进入 `backend-async-runner` 设计前，直接复用本次 `backend-runtime-topology` owner spec 作为边界，不再新增独立 worker 进程前置。
- 次优先：为 B3/A5 codegen drift gate 补一段实施期 index 语义说明，降低后续 generator 变更的误判成本。
- 可延后：在下一次大范围 docs heading 变更时，评估是否给常被引用的小节保留稳定显式 anchor，减少 completed plan 链接 churn。
