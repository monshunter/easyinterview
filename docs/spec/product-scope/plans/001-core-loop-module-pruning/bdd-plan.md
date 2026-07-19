# Core Loop Module Pruning BDD Plan

> **版本**: 1.8
> **状态**: completed
> **更新日期**: 2026-07-19

## 1 证据边界

本 BDD 描述用户可观察的核心范围，不虚构专属 E2E。现存 P0.098 只验
completion/progress，P0.099 只验 report/generating；二者都不承担整站模块
裁剪证明。代码级行为测试、范围 lint 与根 `make test` 是本计划的验证入口。

## 2 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| BDD.CORE.001 | 当前 Home、Workspace、Resume、Practice、Reports 导航 | 用户浏览各核心页面与可用操作 | 不出现 Growth、Mistakes、Drill、独立 Voice、Debrief 或 Profile 业务入口 | frontend behavior tests + scope lint + root `make test` |
| BDD.CORE.SETTINGS.001 | authenticated 用户看到全局 TopBar 与设置齿轮 | 用户进入“设置”并在 Appearance 预览/保存主题 | TopBar 只保留暗色/语言/设置入口；账号主题由 Settings 单次保存并在重登后恢复 | frontend settings/domain tests + `E2E.P0.101` 静态 handoff |

## 3 E2E 关系

只有真实 API/UI 流程才能登记到 `test/scenarios/e2e/`。本计划不新增场景，
也不把 mock、fixture、component test 或源码负向搜索映射成 P0 编号；需要新的
整站用户流程证据时，必须另行设计真实环境场景。
