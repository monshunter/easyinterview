# Resume Listing and Readonly Detail BDD Plan

> **版本**: 2.8
> **状态**: completed
> **更新日期**: 2026-07-15

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.RESUME.READ.001` | 用户已有、缺失、处理中或读取失败的 resume | 打开、刷新或重试 Workshop list / readonly detail | UI 只从 API 渲染摘要与正文并保持 route、waiting/failure、删除和隐私边界；不从 fixture、URL 或浏览器存储伪造内容 | `frontend/src/app/screens/resume-workshop/ResumeWorkshopScreen.test.tsx` + `components/ResumeDetailView.test.tsx`，由根 `make test` 承接 |
| `BDD.RESUME.LIST.002` | 用户在 desktop 或 mobile 打开已有一份或多份简历的 Workshop 列表 | 浏览卡片、打开详情或删除一份简历 | 列表以稳定宽度卡片展示 closed 摘要；desktop 多列左对齐、mobile 单列且无横向溢出；打开只携带 `resumeId`，删除失败保留原卡片，页面不出现 table/header/row 语义 | `frontend/src/app/screens/resume-workshop/components/ResumeListView.test.tsx` + `ResumeWorkshopCssParity.test.ts` + formal responsive owner assertions，由根 `make test` 承接代码行为 |

当前没有覆盖 list/detail/删除的真实 API/UI E2E owner。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
