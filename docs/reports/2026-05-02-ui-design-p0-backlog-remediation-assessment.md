# UI Design P0 Backlog Remediation 交付复盘报告

> **日期**: 2026-05-02
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖 `EasyInterview_UI_Revision_Backlog_v1.1.md` 中 P0 项对 `ui-design/` 静态 UI 的修订：全局 `InterviewContext`、报告 `sessionId` 守卫、报告复练 / 下一轮 payload 分离、Auth pending action 恢复、公司情报合规文案、空状态 / 失败状态，以及 `docs/ui-design/` 同步。

已通过的验证：

- `node --test ui-design/ui-design-contract.test.mjs`，11 项通过。
- `npx --yes esbuild ... --outdir=/tmp/easyinterview-ui-check --format=iife`，`ui-design/src/*.jsx` 解析通过。
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`，Header / INDEX zero drift。
- `python3 scripts/lint/check_md_links.py docs/ui-design`，通过。
- `git diff --check`，通过。
- `bash -n ui-design/run.sh`，脚本语法通过，仅有本机 locale warning。
- `bash ui-design/run.sh --no-open` 后，`curl -I /index.html` 与 `/canvas.html` 均返回 `200 OK`，随后已停止静态服务。

## 2 会话中的主要阻点/痛点

1. `/change-intake` 自动匹配没有命中 UI backlog / `ui-design` 主题，而是低置信度推荐了 OpenAPI 计划。
   - **证据**：matcher 输出 `confidence=low`，推荐 `openapi-v1-contract/001-bootstrap`，与用户明确给出的 `ui-design` 和 UI backlog 不匹配。
   - **影响**：需要人工改为手动定位 `docs/ui-design` 与 `ui-design/src` 运行时真理源，避免误改无关 spec plan。

2. P0 backlog 是外部未跟踪目录输入，且不是 spec-centric plan，当前仓库的计划匹配和 TDD 文档流没有自然 owner。
   - **证据**：`git status` 显示 `EasyInterview_Spec_and_P1_Docs_v1.1/` 为未跟踪输入；`docs/ui-design/README.md` 定义当前目标以静态 UI 运行时为准，而不是 spec plan。
   - **影响**：本次需要用 `ui-design/ui-design-contract.test.mjs` 承接 Red-Green-Refactor，而不能依赖 `/implement` 的 plan context。

3. 既有 UI 契约测试对部分源码形态断言较硬。
   - **证据**：加入 `InterviewContext` 后，旧断言仍要求 `getWorkspaceSessionHistory(lang, job, currentRound?.name)` 的旧签名，测试失败后需要同步为带 `interviewContext` 的新签名。
   - **影响**：测试维护产生一次小返工，但保留了“历史只属于当前规划”的行为约束。

## 3 根因归类

1. `change-intake` 的候选发现偏 spec-centric，对 `docs/ui-design` 和静态 UI backlog 这种运行时原型修订缺少直接路由规则。
   - **类别**：skill

2. `docs/ui-design` 已经声明运行时静态 UI 是目标真理源，但没有把外部 UI backlog 转为可发现的 plan/context。
   - **类别**：spec-plan

3. 旧契约测试把实现签名当作行为契约，说明部分测试还可以更偏行为结果而非调用形态。
   - **类别**：no repo change needed

## 4 对流程资产的改进建议

1. 在 `change-intake` 匹配规则中为 `ui-design`、`docs/ui-design`、`EasyInterview_UI_Revision_Backlog`、`InterviewContext`、`pendingAction`、`Report Dashboard(sessionId)` 增加优先候选归属。
   - **落点**：skill
   - **优先级**：high

2. 后续若继续按 UI backlog 修订，应考虑把 backlog 中 P0/P1 项投影成 `docs/ui-design` 下的轻量检查清单或 spec-centric plan context，便于 `/implement` 和 `/tdd` 接管。
   - **落点**：spec-plan
   - **优先级**：medium

3. UI 契约测试新增项应优先断言 route payload、缺省状态、可见文案和状态分支，减少对辅助函数签名的硬编码。
   - **落点**：no repo change needed
   - **优先级**：medium

## 5 建议优先级与后续动作

最高优先级是补强 `change-intake` 对静态 UI 主题的候选识别，避免后续 UI backlog 修订再次误投到 OpenAPI 或工程路线图计划。

如果 P1 backlog 也要继续落地，建议先把 P1 项整理为 `docs/ui-design` 可检查的任务清单，再进入代码修改；这能让 TDD 项与文档同步边界更清楚。
