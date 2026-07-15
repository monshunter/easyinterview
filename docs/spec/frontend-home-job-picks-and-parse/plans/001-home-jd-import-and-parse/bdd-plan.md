# Home JD Import and Parse BDD Plan

> **版本**: 2.24
> **状态**: completed
> **更新日期**: 2026-07-15

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.HOME.JD.001` | 用户输入合法或非法 JD，且可能没有 selectable Resume、尚未显式选择或已选择 selectable Resume；import/parse 请求也可能失败 | 提交、等待、确认、重试或返回 | 没有显式 selectable Resume 时 CTA disabled 且零 import；ready 或已有可读证据的未归档简历均可选择，只有选择后才提交 exact request。后续 UI 使用 API 状态进入 Workspace 或显示可恢复失败；不重复创建事实、不泄露原文、不从浏览器存储伪造状态 | `frontend/src/app/screens/home/HomeImport.test.tsx` + `HomeScreen.test.tsx` + `HomeResumeSelection.test.tsx`，由根 `make test` 承接 |
| `BDD.HOME.JD.002` | ready Workspace detail 有已保存 Resume binding，或历史数据缺失/无效绑定 | 用户查看绑定简历、开始面试或打开报告 | 合法绑定只按 saved `resumeId` 打开对应详情；Start/Reports 位于左对齐首行动作行。缺绑属于异常状态且训练/报告后续全部 fail closed，不伪造、不 rebind、不 fallback；独立 launch/binding block 与页尾 Start 不存在 | `frontend/src/app/screens/parse/ParseScreen.test.tsx` + `ParseResumeBinding.test.tsx` + `frontend/src/app/App.test.tsx`，由根 `make test` 承接代码行为 |

当前没有覆盖 JD import、parse 或确认 handoff 的真实 API/UI E2E owner；进度刷新场景不承接这些行为。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
