# Frontend Shell I18n Remediation 交付复盘报告

> **日期**: 2026-05-07
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `frontend-shell/001-app-shell-auth-settings` 的 i18n remediation：原地修订 spec / plan / checklist / BDD，补齐 D1 shell `zh` / `en` 独立 locale 文件、浏览器 locale 初始化、English fallback、登录态不覆盖前端语言偏好、`Accept-Language` display hint、TopBar 语言下拉框契约与 E2E.P0.004 场景。
- 成功证据：
  - `pnpm --filter @easyinterview/frontend test src/app/i18n/i18nShell.test.tsx` 通过。
  - `pnpm --filter @easyinterview/frontend test src/app/i18n/localeFiles.test.ts src/app/i18n/i18nShell.test.tsx src/app/scenarios/p0-004-app-shell-language-switch.test.tsx` 通过。
  - `pnpm --filter @easyinterview/frontend test src/app/display/DisplayPreferencesProvider.test.tsx src/app/i18n/localeRuntime.test.tsx src/app/topbar/TopBar.test.tsx` 通过。
  - `pnpm --filter @easyinterview/frontend test src/app/topbar/TopBar.test.tsx src/app/auth/AuthScreens.test.tsx src/app/screens/ProfileScreen.test.tsx` 通过。
  - E2E.P0.004 `setup -> trigger -> verify -> cleanup` 通过，验证 English 文案和 `Accept-Language: en` 证据。
  - `pnpm --filter @easyinterview/frontend typecheck`、`pnpm --filter @easyinterview/frontend test`（30 files / 138 tests）、`pnpm --filter @easyinterview/frontend build`、`make docs-check`、context validator、`git diff --check` 均通过。
- 关联 Bug： [BUG-0020](../bugs/BUG-0020.md)。

## 2 会话中的主要阻点/痛点

- Display preference gate 只测到了 `lang` 状态，没有测用户可见文案。
  - **证据**：原 checklist 2.2 只要求“语言切换在登录前后保持稳定”；`TopBar.test.tsx` 只断言 select value 从 `zh` 到 `en`。
  - **影响**：正式前端看起来有语言控制，但实际没有 i18n 渲染层。

- 最近 frontend shell 实施没有反向审计静态原型的 `lang` 透传语义。
  - **证据**：`ui-design/src/app.jsx` 持有 `lang` 并传给 TopBar/auth/profile/settings；正式前端 TopBar 注释仍写“Labels are user-facing Chinese”。
  - **影响**：迁移时保留了控件形态，却丢掉了中英切换行为。

- Runtime/OpenAPI locale 契约与前端显示偏好的边界没有写清。
  - **证据**：初版 i18n remediation 把 `RuntimeConfig.defaultUiLanguage` 与 `/me.uiLanguage` 接成 UI 语言来源；review 发现这会在 `/me` 刷新期间覆盖前端设置。用户确认口径后，D-7 / Phase 2.5 / 2.8 改为浏览器初始化、English fallback、登录态不覆盖。
  - **影响**：如果不收紧边界，已登录用户的语言会被 runtime 或 `/me` 回写，甚至造成重复 locale 请求。

## 3 根因归类

- 根因 1：原 plan/checklist 对显示偏好类功能的验收语义不足。
  - **类别**：spec-plan
  - **说明**：测试只覆盖状态稳定，未覆盖“用户切换后看到了什么”。

- 根因 2：实现审查没有把 `ui-design/` 的语言切换作为必须迁移的不变量。
  - **类别**：spec-plan
  - **说明**：已有深度重校对规则要求反查 UI truth source，但 D1 的 checklist 没有把 i18n 细化为 gate。

- 根因 3：request locale hint 和 UI locale source 的职责边界混在一起。
  - **类别**：spec-plan
  - **说明**：`Accept-Language` 应由当前前端 UI locale 产生；runtime config 与 `/me` 可保留自己的字段，但不应反向决定 D1 shell 显示语言。

## 4 对流程资产的改进建议

- 在 frontend-shell plan 的 i18n remediation gate 中保留“浏览器默认 + 状态 + 文案 + header + 登录态不覆盖”五件套。
  - **落点**：spec-plan
  - **优先级**：high
  - **状态**：本次已落到 v1.6 2.4 / 2.5 / 2.6 / 2.7 / 2.8。

- 把 locale 文件拆分和语言下拉框从“实现细节”提升为 owner gate。
  - **落点**：spec-plan / frontend README
  - **优先级**：high
  - **状态**：本次已补入 spec D-7 / C-7、plan/checklist 2.7、E2E.P0.004 和 `frontend/README.md`。

- 后续 D2-D6 前端业务页面接入时，禁止直接硬编码新增用户可见静态文案；应扩展 D1 独立 locale 文件或在 owner plan 中声明 locale 扩展。
  - **落点**：spec-plan / frontend README
  - **优先级**：high

- 对“显示控制”类功能建立验收模式：控件存在、状态变化、用户可见结果、跨登录态稳定性都要有断言。
  - **落点**：spec-plan / test README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：后续前端 owner 在替换 `PlaceholderScreen` 时同步扩展独立 locale 文件，避免 D2-D6 页面再次分散硬编码中英文。
- 次优先级：新增语言前先扩展 `localeFiles.test.ts` 和 E2E 语言切换场景，避免 locale 文件存在但 key 不完整或 UI 控件退化。
- 可延后：将 BUG-0020 的模式沉淀到 `docs/bugs/PATTERNS.md`，作为“显示偏好只测状态、不测语义效果”的通用检查项。
