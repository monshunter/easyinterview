# Frontend Dev Mock Runtime 交付复盘报告

> **日期**: 2026-05-10
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 Vite frontend dev preview 默认打真实相对 API，导致请求落到 `localhost:5173/api/v1` 且页面不可见的问题；同步修复 `/change-intake` 在分支门禁前修改 `main` 的流程缺陷。
- 成功证据：`pnpm --filter @easyinterview/frontend test` 69 files / 439 tests PASS；`pnpm --filter @easyinterview/frontend typecheck` PASS；`pnpm --filter @easyinterview/frontend build` PASS；browser smoke 确认 dev preview 页面从 fixture-backed client 渲染且无 `/api/v1` 请求；`make docs-check` 和 `make lint-mock-contract` PASS；`/change-intake` contract tests 7 tests PASS。

## 2 会话中的主要阻点/痛点

- `change-intake` 在完成计划原地修订时先于 `/implement` 写文件。
  - **证据**：用户运行 `git status` 发现当前分支仍是 `main`，且已有 spec、frontend 和 README 未提交改动。
  - **影响**：默认父分支被 dirty worktree 污染，必须中断功能收尾先修流程规则。
- Mock runtime 已存在但未接入真实 Vite bootstrap。
  - **证据**：`main.tsx` 直接构造 `new EasyInterviewClient()`；generated client 默认 `/api/v1`，Vite config 无 proxy。
  - **影响**：组件和 injected-client 测试不能代表本地真实页面可见性。

## 3 根因归类

- `/change-intake` 缺少 branch guard。
  - **类别**：skill / AGENTS.md
- Mock-contract plan 未覆盖 app bootstrap dev-preview gate。
  - **类别**：spec-plan
- Frontend README 未明确说明 `5173` 相对 API 的行为。
  - **类别**：README

## 4 对流程资产的改进建议

- 已把 branch guard 写入 `/change-intake` 与 `AGENTS.md`，并用 contract test 固化。
  - **落点**：skill / AGENTS.md
  - **优先级**：high
- 后续涉及 frontend API wiring 的计划应保留 `main.tsx` bootstrap 或等价 dev smoke gate。
  - **落点**：spec-plan
  - **优先级**：high
- 若未来引入 MSW，应先评估是否替换现有 fixture-backed transport，而不是并行维护两套 mock truth source。
  - **落点**：spec-plan / README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：继续在当前 feature branch 上完成 `mock-contract-suite/001-fixture-backed-mock-runtime` 的收口提交，保持 `main` 不再承载 dirty worktree。
- 下一轮前端 owner 工作应优先接 `frontend-home-job-picks-and-parse/002-jd-match-recommendations`；现在 dev preview 已能无 backend 打开页面，适合恢复页面级开发和截图验证。
