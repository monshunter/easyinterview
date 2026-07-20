# Workspace and Interview Context BDD Checklist

> **版本**: 1.36
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## 静态资产审计

- [x] `E2E.P0.098` 的 BDD 合同只描述真实登录、completion API、Home/Workspace/TargetJob progress refresh 与 TargetJob detail read。
- [x] JD import/parse、chat、session start、quick-start 与下一轮 plan 创建未归入该 E2E。
- [x] 其他 Workspace 用户行为明确为当前无真实 E2E owner 的行为合同。
- [x] 前后端代码层回归由仓库根 `make test` 独立承接，不作为 E2E 证据。

## `BDD.WORKSPACE.CONTEXT.001` Workspace 上下文与训练入口

- [x] Owner behavior tests 覆盖 list/detail、progress、route、exact-plan reuse、final/invalid 与 zero-call fail-closed。
- [x] 根 `make test` 已执行对应 Vitest；该结果不声明 `E2E.P0.098` PASS。

## `BDD.WORKSPACE.DETAIL.002` 详情开头信息层级

- [x] 标题旁“绑定简历”只使用 `TargetJob.resumeId` 导航对应 Resume 详情；缺失绑定不提供链接、不从 route/list/recent resume 兜底。
- [x] “立即面试”与“面试报告”在标题下首行动作行左对齐，desktop 同排、mobile 同序响应式换行；Start/Report route 与错误隔离保持。
- [x] 独立 Interview Launch/绑定简历 block、标题右侧 Report 与页尾 Start 的正式 DOM/source 负向 gate 为零；根 `make test` 和独立 responsive/a11y gates 通过。

## `BDD.WORKSPACE.CARD.003` 规划卡片可见元信息

- [x] 确认验证入口为 `MockInterviewCard.test.tsx` domain behavior test，不声明真实 E2E。
- [x] 执行 owner test，验证 lifecycle status 任意变化都不产生状态文案或徽标。
- [x] 执行 owner test，验证非空真实地点保留，缺失、空或空白地点不产生占位行。
- [x] 记录 focused domain behavior test 证据；仓库根 `make test` 由主 checklist Phase 29 post-pass 独立承接。（验证：2026-07-17 `MockInterviewCard.test.tsx` 13/13 PASS）

## 真实环境证据边界

本 checklist 只完成 owner 关联与静态资产审计；本轮未执行 `E2E.P0.098`，当前真实环境结果以场景 INDEX 的 `Ready` 为准，后续只由显式 `/scenario-run` 产生。

## `BDD.PRACTICE.LAUNCH.004` 会话启动等待反馈

- [x] 确认验证入口为四类正式 caller 的 domain behavior tests 和共享 transition contract，不声明真实 E2E。
- [x] deferred `startPracticeSession` 期间立即显示同一全屏 status/busy transition，背景动作被阻断且重复启动不产生第二次 side effect。
- [x] 成功只导航到现有 `practice` route；失败卸载 transition 并保留各 caller 原有错误；未登录 auth redirect 不提前显示。
- [x] title/body zh/en 完整，DOM 无百分比、伪阶段或 opening message；CSS reduced-motion gate 停用非必要循环动画。
- [x] 记录 focused domain behavior 证据；根 `make test` 与独立 desktop/mobile Chrome pending-state 证据由主 checklist 30.4 收口。（验证：2026-07-18，5 files / 45 tests PASS；frontend typecheck PASS；根 frontend 127 files / 1035 tests PASS；真实 LLM pending 桌面/移动端与成功导航 PASS）

## `BDD.PRACTICE.GLOBAL_CHROME.005` 会话页全局导航一致性

- [x] App/router tests 证明 Practice 同时显示共享 global TopBar 与独立 Practice Session Header，Generating 仍可隐藏 chrome。
- [x] 请求计数测试证明进入、离开和 Practice 内交互不追加 `getMe`。
- [x] desktop/mobile responsive/a11y 与 current-run Chrome 截图证明导航、设置齿轮、会话 CTA 可达且 document 无横向溢出。

## `BDD.WORKSPACE.LIST.VISUAL.006` 面试列表参考稿层级

- [x] 确认验证入口为 Workspace/Card domain behavior tests 与 current-run Chrome UI acceptance，不创建 E2E wrapper。
- [x] 执行 owner tests，验证全视口背景与居中内容分层、header CTA 与第二列卡片右边界对齐、desktop 双列宽卡、mobile 单列、公司/岗位/动态 rail/上次保存/footer 动作层级。（验证：背景与 CTA 对齐各 RED 1 项后，Workspace CSS/component 7 tests PASS；其余 Workspace/Card/CSS owner tests PASS）
- [x] 执行打开、删除成功/失败、启动 fail-closed 与 loading/empty/error 回归，确认 generated client 和 route 合同不变。（验证：owner scope 24 files / 150 tests PASS）
- [x] 记录 1916×821 / 2048×917 / 390×844 bbox、截图、keyboard、theme、console 与 no-overflow 证据。（验证：desktop 背景左右边界差值均为 0，CTA 与第二列卡片右边界差值 0px；mobile 358px 单列、document overflow 0；截图位于 `.test-output/list-ui-acceptance/`）
## `BDD.PRACTICE.LAUNCH.VISUAL.007` Practice 启动过渡构图

- [x] Shared/caller tests 覆盖 brand variant、portal blocking、focus/scroll lock、single-flight、成功与失败恢复、无伪百分比。
- [x] Current-run desktop Chrome 对照参考稿验证 TopBar、同心轨道、E 标识、标题/说明/indeterminate 线和无横向溢出；mobile 由共享 responsive contract 覆盖，不新增 E2E ID。（Workspace 与 Report caller 均完成真实启动并成功进入 Practice。）
- [x] Phase 34 remediation：真实 App root 层级下锁定完整背景 inert，证明可见 TopBar 的导航/设置/显示控制不能穿透 pending transition，卸载后属性恢复。<!-- verified: 2026-07-20 evidence="PracticeLaunchTransition 2/2 and root frontend 1115/1115 PASS." -->

## `BDD.WORKSPACE.DELETE.CONFIRM.008` 面试规划删除二次确认

- [x] 确认验证入口为 Workspace owner tests 与 shared destructive-dialog domain behavior tests，不创建 E2E wrapper。
- [x] 执行首次点击零 `archiveTargetJob`、卡片零导航、取消/Escape/遮罩零副作用、初始焦点/focus trap/trigger focus restore 行为断言。
- [x] 执行确认单次提交、pending 关闭/重复提交锁定、失败保留卡片/弹窗与同 key retry、成功隐藏卡片行为断言。
- [x] 使用 Chrome skill 在真实 frontend/backend 截取面试规划删除确认框并记录 keyboard、console 与 no-overflow；该截图证据不声明 E2E PASS。<!-- verified: 2026-07-20 method=chrome-extension-manual evidence="1212x912 real frontend/backend screenshot; cancel initially focused; cancel closed dialog, preserved interview-plan card and restored delete-trigger focus; console errors 0; horizontal overflow false; screenshot=.test-output/delete-confirmation-ui/workspace-delete-confirmation.png" -->
