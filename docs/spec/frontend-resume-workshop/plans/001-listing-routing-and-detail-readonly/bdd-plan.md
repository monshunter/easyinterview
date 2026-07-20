# Resume Listing and Readonly Detail BDD Plan

> **版本**: 2.15
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 Plan**: [plan](./plan.md)

## 行为合同

| Behavior ID | Given | When | Then | 验证入口 |
|-------------|-------|------|------|----------|
| `BDD.RESUME.READ.001` | 用户已有、缺失、处理中或读取失败的 resume | 打开、刷新或重试 Workshop list / readonly detail | UI 只从 API 渲染摘要与正文并保持 route、waiting/failure、删除和隐私边界；不从 fixture、URL 或浏览器存储伪造内容 | `frontend/src/app/screens/resume-workshop/ResumeWorkshopScreen.test.tsx` + `components/ResumeDetailView.test.tsx`，由根 `make test` 承接 |
| `BDD.RESUME.LIST.002` | 用户在 desktop 或 mobile 打开已有一份或多份简历的 Workshop 列表 | 浏览语义卡片、打开详情或删除一份简历 | 列表以语义 card/list 展示 closed 摘要；打开只携带 `resumeId`，删除失败保留原卡片，页面不出现 table/header/row 语义 | `frontend/src/app/screens/resume-workshop/components/ResumeListView.test.tsx` domain behavior tests，由根 `make test` 承接代码行为；几何由 `BDD.RESUME.LIST.VISUAL.003` owner |
| `BDD.RESUME.LIST.VISUAL.003` | 用户在 desktop/mobile 打开已有简历的 Workshop 列表 | 浏览参考稿标题区、卡片、meta 与 footer，或使用创建/打开/删除动作 | desktop 每行排列两张等宽卡，mobile 占满可用宽度并收敛为单列；创建入口使用与 Workspace 一致的 22px circled-plus；文件 icon、名称/摘要、来源/最近编辑、删除和“打开”层级一致；create/open/archive 仍使用既有 route/generated client，失败保留卡片且无横向溢出 | `ResumeListView.test.tsx` + `ResumeWorkshopCssParity.test.ts` domain behavior tests；current-run Chrome 仅作 UI 证据 |
| `BDD.RESUME.DETAIL.VISUAL.004` | 用户在 desktop/mobile 打开已有可读正文的简历详情 | 浏览 Header 与来源格式自适应正文 | desktop 显示共享左边界的 Back、蓝色 eyebrow、名称 kicker、主标题和 meta，并以约 `1512px` 内容面承载直接居中的 PDF 页面栈或 Markdown page surface；正文外层没有共享浅色背景板；mobile 同序满宽可读；PDF/Markdown renderer、只读行为和正文事实不变 | `ResumeDetailView.test.tsx` + `ResumePreviewTab.test.tsx` + `ResumeWorkshopCssParity.test.ts` domain behavior tests；current-run Chrome 仅作 UI 证据 |
| `BDD.RESUME.PARSE.VISUAL.005` | queued/processing Resume 尚无可读正文且后台轮询可能跨多个 request | 用户停留在详情等待、返回列表或等待 ready/failed | 页面持续显示共享 resume transition 且不闪现通用 loading；TopBar/返回可用，ready/failed 原子替换，desktop/mobile 无横向溢出 | `ResumeDetailView.test.tsx` + polling/CSS parity tests + current-run Chrome manual acceptance；根 `make test` 承接代码层回归 |
| `BDD.RESUME.DELETE.CONFIRM.006` | 用户在简历列表看到一份 active 简历，删除请求尚未发生 | 点击删除后取消，或明确确认并经历 pending/success/failure | 首次点击只打开可访问确认框且 archive 调用为 0；取消/Escape/遮罩关闭并恢复焦点；确认只发一个请求，pending 禁止关闭/重复提交；成功隐藏卡片，失败保留卡片与弹窗并同 key 重试 | `ResumeListView.test.tsx` + shared destructive-dialog component tests；current-run Chrome 仅作 UI 截图证据 |
| `BDD.RESUME.DETAIL.A4.007` | 用户在 desktop/mobile 打开 PDF 或 Markdown 简历详情 | 浏览来源格式自适应正文 | desktop PDF/Markdown page surface 共用 `794px` A4 纸宽；PDF 每页保持 `210:297`，Markdown 整份正文是一张不分页、无 A4 高度约束的连续长页面；mobile 在可用宽度内收敛且无横向溢出；PDF.js、Markdown DOM、只读边界与正文事实不变 | `ResumeWorkshopCssParity.test.ts` + `ResumePreviewTab.test.tsx` + `PdfPageStackPreview.test.tsx` domain behavior tests；current-run Chrome 仅作 UI 证据 |

当前没有覆盖 list/detail/删除的真实 API/UI E2E owner。代码层回归统一由根 `make test` 承接，不能作为 E2E 证据。
