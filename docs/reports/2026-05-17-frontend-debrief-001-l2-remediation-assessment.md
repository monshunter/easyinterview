# Frontend Debrief 001 L2 Remediation 交付复盘报告

> **日期**: 2026-05-17
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`frontend-debrief/001-debrief-screen-and-handoff` 的 L2 code review 修复、P0.069 Playwright parity 补齐、P0.065-P0.069 BDD 顺序链、Bug KB 与 plan/test/BDD 文档收口。
- 成功证据：
  - `pnpm --filter @easyinterview/frontend exec vitest run src/app/screens/debrief/DebriefScreen.test.tsx` 通过（6 tests）。
  - `pnpm --filter @easyinterview/frontend build` 通过。
  - `pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/debrief.spec.ts` 通过（11 passed / 1 skipped）。
  - `E2E.P0.065-P0.069` 按 `setup.sh -> trigger.sh -> verify.sh -> cleanup.sh` 顺序串行通过。
  - `BUG-0067` 已记录，`docs/bugs/PATTERNS.md` 新增 completed checklist / runner false-green 复发模式。
  - `make validate-fixtures`、`pnpm --filter @easyinterview/frontend typecheck`、`pnpm --filter @easyinterview/frontend lint`、`python3 -m pytest scripts/lint -q`、`make docs-check`、`git diff --check` 作为最终 gate 重新执行。

## 2 会话中的主要阻点/痛点

- Debrief route/context 恢复链缺失。
  - **证据**：新增红测显示 route / InterviewContext IDs 不会触发 `getTargetJob` / `getPracticeSession` / `getResumeVersion`，页面草稿不会自动水合。
  - **影响**：用户从 URL、auth pending action 或已存在 InterviewContext 进入 debrief 时，会落到缺上下文或推荐不触发状态。
- Submit / handoff payload 字段混用。
  - **证据**：English UI 下 `createDebrief` 仍发送 `zh-CN`；submit 401 没有 pending action；replay handoff 缺少 `language`。
  - **影响**：跨语言与未登录恢复路径不闭合，practice owner 无法稳定获得复盘会话语言。
- P0.069 pixel gate false-green。
  - **证据**：原 P0.069 trigger 只跑 i18n/privacy/devMock/legacy，没有调用 Playwright；verify 也不要求 Playwright marker。
  - **影响**：Phase 8.1-8.3 被标成 completed/deferred，但用户可见 UI parity 没有浏览器执行证据。
- Completed 文档证据漂移。
  - **证据**：`test-checklist.md` 与 `bdd-checklist.md` header 为 completed，但所有条目仍是 unchecked。
  - **影响**：后续 owner 无法区分真实完成、历史待补和本轮 L2 修复边界。

## 3 根因归类

- Route/context 恢复链属于 **spec-plan + frontend runtime** 问题：测试覆盖了直接 render，但没有覆盖 URL / InterviewContext / auth replay 的真实入口。
- Submit / handoff payload 属于 **frontend runtime contract** 问题：UI language、practice modality 和 mode 未在测试中分离断言。
- P0.069 false-green 属于 **spec-plan + scenario gate** 问题：verify 没有要求真实 runner marker 和 pass marker。
- Checklist 漂移属于 **spec-plan evidence drift**：completed 状态没有和 test checklist / BDD checklist 同批审计。

## 4 对流程资产的改进建议

- 在前端 L2 review 中固定检查 route params、InterviewContext 与 auth pending action replay 的同一恢复链。
  - **落点**：plan-code-review 执行习惯 / future plan gate
  - **优先级**：high
- 对所有 user-visible flow 的 payload 测试分离 `language`、`mode`、`modality`、`practiceGoal`。
  - **落点**：spec-plan
  - **优先级**：high
- 对 completed plan 先执行 unchecked/deferred/no-op 负向搜索。
  - **落点**：`docs/bugs/PATTERNS.md` 已新增模式 4
  - **优先级**：high
- Pixel parity scenario verify 必须要求 Playwright marker、目标 spec path 和 pass count。
  - **落点**：P0.069 trigger/verify 已修复；后续同类场景复用
  - **优先级**：high

## 5 建议优先级与后续动作

下一步建议先执行 `/work-journal` 并提交本次 `fix(frontend-debrief): close debrief L2 gaps (BUG-0067)`；随后进入 push / PR 回复阶段，PR 回复重点说明 route hydration、language/auth/handoff 修复、P0.069 Playwright gate 从 deferred 改为真实执行，以及 P0.065-P0.069 顺序场景证据。
