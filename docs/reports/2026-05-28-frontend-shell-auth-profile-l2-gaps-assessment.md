# Frontend Shell Auth Profile L2 Gaps 交付复盘报告

> **日期**: 2026-05-28
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`plan-code-review frontend-shell/001-app-shell-auth-settings --fix` 的 L2 收口，重点修复单入口登录 request body、首次资料补全恢复门禁、P0.102 场景 wrapper 证据强度，并把 `docs/spec/frontend-shell/plans/001-app-shell-auth-settings/` 原地修订到 v1.15。
- 成功证据：
  - Red: `pnpm --filter @easyinterview/frontend test src/app/auth/AuthScreens.test.tsx` 先暴露 `{ email, returnTo }` body 和资料补全提前导航。
  - Green: `pnpm --filter @easyinterview/frontend typecheck`。
  - Green: `pnpm --filter @easyinterview/frontend test src/app/auth/AuthScreens.test.tsx src/app/auth/AuthVisual.test.tsx src/app/AppAuthDispatch.test.tsx`。
  - Green: P0.102 `setup.sh` / `trigger.sh` / `verify.sh` / `cleanup.sh`，`trigger.log` 包含 Vitest 24 tests、backend named `--- PASS` markers 与 package `ok` 行。
  - Green: `pnpm --filter @easyinterview/frontend build`、`bash -n` P0.102 scripts、`make docs-check`、`sync-doc-index --check`、`validate_context.py --target frontend`、`git diff --check`。
  - 收尾资产：新增 [BUG-0116](../bugs/BUG-0116.md)，计划 / checklist / plans INDEX 原地更新到 v1.15。

## 2 会话中的主要阻点/痛点

- 历史 auth 口径残留在测试与文档中。
  - **证据**：`AuthScreens.test.tsx` 原本断言 `startAuthEmailChallenge` 接收 `returnTo`；plan 的历史 Phase 6 / Phase 8 文案仍把 `email / returnTo` 写成当前行为。
  - **影响**：当前 Phase 9 email-only contract 被旧测试保护，L2 需要先改测试才能暴露真实漂移。
- 资料补全 callback 类型过宽。
  - **证据**：`AuthProfileSetupScreen` 原本接受 `Promise<void>`，无法判断后端刷新后的 `profileCompletionRequired` 是否已经清除。
  - **影响**：组件可以在资料仍未补全时恢复 pendingAction，违背首次资料补全硬门禁。
- 场景 wrapper 存在假绿风险。
  - **证据**：P0.102 verify 原本只检查 Go package 字符串；新增 forbidden marker 后发现 `grep` 对 `--- FAIL:` / `--- SKIP:` 缺少 `--`，会打印错误但脚本仍退出 0。
  - **影响**：场景结果可能显示 PASS，但未证明 focused backend tests 真正执行，也可能漏掉 shell option edge case。

## 3 根因归类

- 历史口径残留：
  - **类别**：spec-plan
  - Phase 9 改版后，历史 Phase 6 / Phase 8 的说明没有明确降级为历史兼容口径，导致 review 时需要额外 reconcile。
- callback 类型过宽：
  - **类别**：spec-plan
  - plan 要求“刷新 `/me` 并确认 false”，但 checklist 旧证据没有要求组件 test 覆盖 `profileCompletionRequired=true` 的负向路径。
- wrapper 假绿：
  - **类别**：README / skill
  - 场景框架文档要求 runner marker 和 pass marker，但现有场景模板 / 习惯没有强制 named PASS 与 `grep --` 细节。

## 4 对流程资产的改进建议

- 在 `scenario-create` 或场景 README 模板中增加 shell wrapper 约束：forbidden marker 固定字符串必须使用 `grep -Fq -- "$pattern"`，Go focused tests 如果作为场景证据应使用 `go test -v` 并由 verify 检查具体 `--- PASS:` 名称。
  - **落点**：skill / README
  - **优先级**：high
- 对 auth/profile 这类状态机计划，checklist 应要求至少一个“后端状态仍未达标时不得恢复业务动作”的负向 test，而不只写 happy path。
  - **落点**：spec-plan
  - **优先级**：medium
- 对已完成 plan 的历史阶段，若后续阶段取代了旧口径，历史段落应显式写明“不再作为当前验收口径”，并避免使用“当前”描述旧行为。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮最值得做：更新场景创建 / wrapper 指南，把 named PASS、runner marker、`grep --` 和 no-test/skip/fail 负向断言固化为新场景默认模板。
- 可以延后：对其他已完成 frontend-shell 历史阶段做一次只读扫描，找出仍用“当前”描述已废弃 auth/register/mail-link 口径的段落，避免下一次 L2 重复 reconcile。
