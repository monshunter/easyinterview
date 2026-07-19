# Workspace and Interview Context BDD Plan

> **版本**: 1.34
> **状态**: completed
> **更新日期**: 2026-07-19

**关联 Plan**: [plan](./plan.md)

## 当前真实 E2E owner

| 场景 | Given | When | Then | 真实 E2E |
|---|---|---|---|---|
| 完成后的进度刷新 | real frontend/backend 已运行，用户已真实登录，TargetJob 已有轮次与 session | 通过真实 completion API 完成第一轮，并刷新 Home、Workspace 与 TargetJob 详情 | 用户在三处都看到第一轮已完成、第二轮为当前、第三轮待进行，并可打开同一 TargetJob 详情 | `E2E.P0.098` |

`E2E.P0.098` 不承接 JD import/parse、chat、session start、quick-start 或下一轮 plan 创建。

## Domain behavior

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.WORKSPACE.CONTEXT.001` | 用户打开 Workspace list/detail，后端 progress/plan 可能完整、终态或无效 | 选择 TargetJob、查看轮次或发起训练 | 页面只消费后端投影并保持 route、隐私、exact-plan reuse 与 fail-closed 约束 | `frontend/src/app/screens/workspace/WorkspaceScreen.test.tsx` + `hooks/useWorkspaceTargetJobs.test.tsx`，由根 `make test` 承接 |
| `BDD.WORKSPACE.DETAIL.002` | 用户打开已有或缺失绑定简历的 Workspace 详情 | 查看绑定简历、开始面试或打开面试报告 | 标题旁“绑定简历”只按 saved `resumeId` 打开对应 Resume 详情；Start/Reports 在标题下左对齐首行动作行；缺绑定只禁用 Start，不伪造绑定；页面无独立 binding/launch block 或页尾 Start | `frontend/src/app/screens/parse/ParseScreen.test.tsx` + `ParseResumeBinding.test.tsx` + `frontend/src/app/App.test.tsx`，由根 `make test` 承接代码行为 |
| `BDD.WORKSPACE.CARD.003` | 用户查看 Home 最近面试或 Workspace 面试规划卡片，TargetJob lifecycle status 为任意值且地点可能有值或缺失 | 卡片渲染共享主体 | 不显示 lifecycle status 文案/徽标；非空真实地点正常显示；缺失、空或空白地点不显示 `Location not set` 或空地点行；round rail 仍表达 backend progress | `frontend/src/app/screens/home/MockInterviewCard.test.tsx` domain behavior test，由根 `make test` 承接代码回归 |
| `BDD.PRACTICE.LAUNCH.004` | 用户从 Home recent、Workspace list/detail 或 Report replay/next-round 发起有效面试，opening LLM 请求保持 pending 或失败 | 点击启动并等待 | 立即显示统一全屏、可访问且阻断交互的诚实 indeterminate transition；无伪进度/opening；成功进入 `practice`，失败关闭 transition 并恢复原入口错误；auth redirect 不提前展示 | shared transition contract + `HomeRecentMocks.test.tsx`、`WorkspaceScreen.test.tsx`、`ParseResumeBinding.test.tsx`、`ReplayCta.test.tsx` domain behavior tests，由根 `make test` 承接 |
| `BDD.PRACTICE.GLOBAL_CHROME.005` | authenticated 用户已由 app bootstrap 取得账号/runtime context | 进入、使用或离开 Practice | 全局 App TopBar 始终位于独立 Practice Session Header 上方，导航/显示/设置入口可用；route 切换 `/me` 增量为 0；desktop/mobile 无横向溢出 | App/router/Practice/request-count/responsive tests，由根 `make test`；current-run Chrome desktop/mobile 作为 UI 证据，不冒充 E2E ID |
| `BDD.WORKSPACE.LIST.VISUAL.006` | authenticated 用户在 desktop/mobile 打开有 1~N 个 ready TargetJob 的 Workspace list | 浏览卡片、打开详情、删除或开始模拟面试 | TopBar 下方背景连续覆盖完整 viewport；内容层按参考稿显示标题区、双列宽卡/移动单列、公司/岗位/动态轮次、上次保存与独立动作，header CTA 右侧与第二列卡片右侧对齐；交互仍只调用既有 generated client/route，失败状态不丢卡且无横向溢出 | `WorkspaceScreen.test.tsx` + `WorkspaceVisual.test.ts` + `MockInterviewCard.test.tsx` domain/visual contract tests；current-run Chrome 仅作 UI 证据 |
| `BDD.PRACTICE.LAUNCH.VISUAL.007` | 用户从合法 Home、Workspace 或 Report 动作发起会话且 opening request 仍 pending | 查看启动反馈、等待成功或遭遇失败 | 保留全局 TopBar 的共享 brand transition 阻断背景交互并管理 focus/scroll；只表达真实 pending，成功进入 Practice，失败恢复原入口，desktop/mobile 无横向溢出 | `PracticeLaunchTransition.test.tsx` + caller tests + current-run Chrome manual acceptance；根 `make test` 承接代码层回归 |

`E2E.P0.098` 是 completion/progress refresh 的独立 suite handoff；quick-start/session start/next-round 不归入该 E2E。
