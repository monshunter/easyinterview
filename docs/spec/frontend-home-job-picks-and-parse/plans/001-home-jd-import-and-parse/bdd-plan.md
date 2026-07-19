# Home JD Import and Parse BDD Plan

> **版本**: 2.28
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.HOME.JD.001` | 用户输入合法或非法 JD，且可能没有 selectable Resume、尚未显式选择或已选择 selectable Resume；import/parse 请求也可能失败 | 提交、等待、确认、重试或返回 | 没有显式 selectable Resume 时 CTA disabled 且零 import；ready 或已有可读证据的未归档简历均可选择，只有选择后才提交 exact request。后续 UI 使用 API 状态进入 Workspace 或显示可恢复失败；不重复创建事实、不泄露原文、不从浏览器存储伪造状态 | `frontend/src/app/screens/home/HomeImport.test.tsx` + `HomeScreen.test.tsx` + `HomeResumeSelection.test.tsx`，由根 `make test` 承接 |
| `BDD.HOME.JD.002` | ready Workspace detail 有已保存 Resume binding，或历史数据缺失/无效绑定 | 用户查看绑定简历、开始面试或打开报告 | 合法绑定只按 saved `resumeId` 打开对应详情；Start/Reports 位于左对齐首行动作行。缺绑属于异常状态且训练/报告后续全部 fail closed，不伪造、不 rebind、不 fallback；独立 launch/binding block 与页尾 Start 不存在 | `frontend/src/app/screens/parse/ParseScreen.test.tsx` + `ParseResumeBinding.test.tsx` + `frontend/src/app/App.test.tsx`，由根 `make test` 承接代码行为 |
| `BDD.HOME.JD.003` | 用户在支持的语言、主题和 desktop/mobile viewport 打开 Home，数据可能 loading/empty/error/ready | 浏览 Hero、录入 JD、选择简历、查看或继续最近面试 | screenshot-aligned hierarchy 与状态均可读可操作；计数器来自真实 runtime limit；1~3 条 recent record 使用 API rounds/progress；mobile 单列无横溢；视觉重排不改变请求、路由、Resume gate 或隐私合同 | `HomeLayout.test.tsx` + `HomeRecentMocks.test.tsx` + TopBar visual tests；Chrome `1916x821` / `390x844` manual visual acceptance |
| `BDD.HOME.PLAN.VISUAL.004` | ready Workspace detail 有合法或异常的 Resume/progress 投影，并包含 2~5 个动态轮次 | 用户在 desktop/mobile 查看 Header、信息卡、要求、关注点与轮次并触发既有动作 | 目标构图保持 `1250px` 内容列、Header 右侧动作和四层响应式卡面；异常事实继续 fail closed，所有动态内容来自 TargetJob 且页面无横向溢出 | `ParsePlanVisual.test.ts` + Parse/Workspace detail behavior tests；Chrome `1916x821` / `390x844` manual visual acceptance |
| `BDD.HOME.JD.PARSE.VISUAL.005` | command-only Parse 对合法 TargetJob 轮询到 queued/processing、ready 或 failed | 用户查看四步等待、返回或等待 ready replace | TopBar/面试高亮保留；共享 job transition 展示四步视觉节奏但不伪装后端 percent/阶段事实；ready replace 到 Workspace detail，失败/返回可恢复，desktop/mobile 无横向溢出 | `ParseScreen.test.tsx` + shared scene/CSS tests + current-run Chrome manual acceptance；根 `make test` 承接代码层回归 |
| `BDD.HOME.JD.TEXTAREA.006` | 用户在 Home 打开唯一 JD textarea，内容可能为空、较短或为多行长文本 | 输入、粘贴或删除 JD 内容 | 默认可视高度至少 212px；长内容按实际 `scrollHeight` 增高并完整显示，删减后重新测量回缩但不低于默认高度；width 保持 100%，无内部纵向滚动或页面横向溢出，计数/Resume/CTA/request/route 不变 | `HomeLayout.test.tsx` + current-run Chrome desktop manual acceptance；根 `make test` 承接代码层回归 |

当前没有覆盖 JD import、parse 或确认 handoff 的真实 API/UI E2E owner；进度刷新场景不承接这些行为。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
