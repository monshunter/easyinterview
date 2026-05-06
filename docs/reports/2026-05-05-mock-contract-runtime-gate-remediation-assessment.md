# Mock Contract Runtime Gate Remediation 交付复盘报告

> **日期**: 2026-05-05
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：按 `/plan-code-review mock-contract-suite --fix` 的 L2 findings，原地修复 `mock-contract-suite/001-fixture-backed-mock-runtime` Phase 4 gate 漏覆盖、Go generated route table 旧 36-row 口径和前端 mock scenario 测试缺口。
- 成功证据：`python3 -m unittest scripts.lint.makefile_dry_run_test scripts.mock_contract.fixture_registry_test -v`、`cd backend && go test ./cmd/codegen/openapi ./internal/api/mockruntime -count=1`、`pnpm --filter @easyinterview/frontend test src/api/mockTransport.test.ts`、`make lint-mock-contract`、context validation、`make docs-check`、`git diff --check` 均通过。
- Codegen 证据：未提交补丁状态下裸 `make codegen-check` 停在预期 generated diff；用临时 Git index 将当前补丁作为基线后，`make codegen-check` 通过，证明生成器幂等。
- Lifecycle：001 plan/checklist 已新增并完成 4.3 remediation，版本投影更新到 1.1，状态保持 `completed`；context discovery 已纳入 `scripts/mock_contract`。
- Bug linkage：本次 gate drift 已建档为 [BUG-0010](../bugs/BUG-0010.md)。

## 2 会话中的主要阻点/痛点

- registry helper 有 focused test，但没有进入 owner 聚合 gate。
  - **证据**：修复前 `make -n lint-mock-contract` 只包含 `validate-fixtures`、`lint-openapi` 与 `mock_runtime_boundary.py`；`scripts.mock_contract.fixture_registry_test` 未执行。
  - **影响**：Phase 1 registry metadata 可单独漂移，而 Phase 4 local gate 仍显示通过。
- generated 注释保留手写 operation count。
  - **证据**：修复前 `openapi/templates/go/server.tmpl` 和 `backend/internal/api/generated/server.gen.go` 写死 `36-row table`，当前 B2 truth source 为 34 operation。
  - **影响**：生成物文案与当前 contract 不一致，也暴露出模板未从 operation inventory 派生数量。
- 前端 mock transport 的 scenario 行为缺少回归断言。
  - **证据**：新增 named scenario / unknown scenario Vitest 后立即通过，说明实现已存在但测试未锁契约。
  - **影响**：未来重构 `createFixtureBackedFetch` 时可能破坏 fail-loudly contract 而不被 focused test 发现。
- dirty worktree 下 codegen-check 的 diff 语义需要明确记录。
  - **证据**：裸 `make codegen-check` 在当前未提交 generated diff 下返回非零；临时 index gate 通过。
  - **影响**：如果不记录，会把“预期生成物变更导致 drift gate 非零”误读成 codegen 非幂等。

## 3 根因归类

- Phase 4 gate 文案只列下游 lint，没有明确 owner helper test 必须进入聚合 target。
  - **类别**：spec-plan
- Go server 模板把 endpoint inventory 数量写成常量，缺少生成器测试约束。
  - **类别**：spec-plan
- 前端 scenario contract 依赖实现阅读而不是测试断言。
  - **类别**：spec-plan
- codegen-check 在未提交补丁中的 diff 语义是当前仓库 gate 设计事实，不需要改 gate。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 后续任何 plan 新增 owner-owned helper / registry / loader 时，Phase verification 必须写明“聚合 gate 直接执行对应 focused test”，并在 Makefile dry-run test 中锁住 wiring。
  - **落点**：spec-plan
  - **优先级**：high
- codegen 模板中的 inventory 数量、tag 数、operation 数应优先从 render data 派生；若必须写入文本，必须有 generator test 锁定当前 truth source。
  - **落点**：spec-plan
  - **优先级**：medium
- 对 mock scenario selection 这类跨前后端契约，测试应覆盖 default、named scenario 和 unknown scenario 三类，而不是只覆盖 happy-path default。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高价值下一步：继续推进 `frontend-shell/001-app-shell-auth-settings frontend` 时，直接把 `lint-mock-contract` 作为前置 gate，并消费 `createFixtureBackedFetch`；这样能验证 BUG-0010 后的 mock runtime handoff 不再漏 registry / scenario contract。
- 可延后动作：若后续反复遇到 generated diff 与未提交补丁的 gate 语义混淆，再考虑在 README 中补充临时 Git index 验证说明；当前只在本次 checklist / Bug / retrospective 中记录即可。
