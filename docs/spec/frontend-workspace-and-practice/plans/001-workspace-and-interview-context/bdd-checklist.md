# Workspace and Interview Context BDD Checklist

> **版本**: 1.31
> **状态**: completed
> **更新日期**: 2026-07-19

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
