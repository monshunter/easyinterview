# Frontend Report Dashboard Replay Pixel Gate Follow-up 交付复盘报告

> **日期**: 2026-05-16
> **审查人**: Codex

## 1 复盘范围与成功证据

本次复盘覆盖 `main` 相较 `checkpoint/013` 的 follow-up 修复：report replay / next-round CTA fresh-session handoff，以及 `E2E.P0.059` pixel parity scenario 的 false-green gate 修正。关联 Bug 记录为 [BUG-0064](../bugs/BUG-0064.md)。

成功证据：

- `pnpm --filter @easyinterview/frontend test src/app/auth/__tests__/pendingActionReplayPractice.test.ts src/app/screens/report/__tests__/ReplayCta.test.tsx`：2 files / 10 tests passed。
- `E2E.P0.057` setup → trigger → verify → cleanup 通过，验证 pendingAction 恢复到 workspace auto-start。
- `E2E.P0.059` setup → trigger → verify → cleanup 通过。
- P0.059 trigger 内执行 `pnpm --filter @easyinterview/frontend build`。
- P0.059 trigger 内执行 `pnpm --filter @easyinterview/frontend test:pixel-parity -- tests/pixel-parity/generating.spec.ts tests/pixel-parity/report.spec.ts`，Playwright `14 passed`。

## 2 会话中的主要阻点/痛点

- Report replay handoff 只验证离开 report，没有验证 fresh practice session 创建。
  - **证据**：红测把期待改为 `startPracticeSession` 调用后，当前实现调用次数为 0，并落到 `practice-session-lost`。
  - **影响**：已完成的 source session 会被上下文保留下来，核心复练入口不可用。
- P0.059 scenario wrapper 把 Playwright 作为“文件已存在”证据。
  - **证据**：旧 `trigger.sh` 不运行 `test:pixel-parity`，旧 `verify.sh` 只检查两个 spec 文件存在。
  - **影响**：scenario 可以在没有任何 visual parity runner 的情况下 false-green。
- 场景元数据存在复制残留。
  - **证据**：P0.057 / P0.058 / P0.059 `setup.sh` 曾写入或保留错误 scenario id；本轮统一修正为各自场景号。
  - **影响**：失败或证据归档时会误指向错误场景 ID。

## 3 根因归类

- Replay handoff 复用 source session：**类别** `spec-plan`。D-5 / BDD 文档原先描述 direct `practice` handoff，没有把 workspace owner 的 session creation 作为不可绕过的不变量。
- P0.059 false-green：**类别** `spec-plan`。场景 verify 缺少 runner marker 与 pass marker，只检查了静态资产。
- 场景元数据误标：**类别** `no repo change needed`。这是同目录脚本复制残留，已在本次修复中直接更正。

## 4 对流程资产的改进建议

- 在 report/dashboard 类 CTA gate 中固定断言“目标 owner 创建 fresh resource”，而不只断言 route name。
  - **落点**：spec-plan
  - **优先级**：high
- 对所有 scenario wrapper 禁止用“spec 文件存在”替代 runner 通过；verify 必须检查 trigger log 中的命令 marker、目标 spec path 和 pass marker。
  - **落点**：README / scenario-run skill
  - **优先级**：high
- 场景 setup metadata 应作为 verify 的轻量检查项，避免复制残留的 scenario id 进入证据目录。
  - **落点**：README
  - **优先级**：medium

## 5 建议优先级与后续动作

最高优先级是把 scenario wrapper 的执行证据规则沉淀到 `test/scenarios/README.md` 或 `scenario-run` skill：凡 trigger 声称覆盖 Playwright / Vitest / pytest，verify 必须检查实际 runner output，而不是只检查文件或脚本存在。其次，在 frontend-report-dashboard 后续计划里保留 workspace auto-start fresh-session invariant，避免 report owner 再次绕过 practice session creation。
