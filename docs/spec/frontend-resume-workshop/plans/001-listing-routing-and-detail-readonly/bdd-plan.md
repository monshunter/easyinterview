# Resume Listing and Readonly Detail BDD Plan

> **版本**: 2.7
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.RESUME.READ.001` | 用户已有、缺失、处理中或读取失败的 resume | 打开、刷新或重试 Workshop list / readonly detail | UI 只从 API 渲染摘要与正文并保持 route、waiting/failure、删除和隐私边界；不从 fixture、URL 或浏览器存储伪造内容 | `frontend/src/app/screens/resume-workshop/ResumeWorkshopScreen.test.tsx` + `components/ResumeDetailView.test.tsx`，由根 `make test` 承接 |

当前没有覆盖 list/detail/删除的真实 API/UI E2E owner。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
