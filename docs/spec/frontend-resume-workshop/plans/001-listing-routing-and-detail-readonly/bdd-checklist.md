# Resume Listing and Readonly Detail BDD Checklist

> **版本**: 2.14
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.RESUME.READ.001` 列表与只读详情

- [x] Owner behavior tests 覆盖 list/detail、waiting/failure、删除、route 与隐私。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 list/detail/delete 真实 E2E owner；不创建 wrapper 场景。

## `BDD.RESUME.LIST.002` 语义卡片列表

- [x] Owner behavior tests 覆盖语义 card/list、closed 摘要字段与 table/header/row 负向语义；参考稿几何由 `BDD.RESUME.LIST.VISUAL.003` owner。（验证：list/fixture/StrictMode tests PASS）
- [x] 卡片“打开”只导航 `resume_versions?resumeId=...`；删除成功隐藏、失败保留并显示可恢复错误；loading/empty/error/pagination 不回退。（验证：Resume Workshop 20 files / 118 tests PASS）
- [x] 根 `make test` 与独立 responsive/a11y gates 通过；该结果是代码层行为证据，不声明真实 E2E PASS。（验证：Chrome 1440/390 geometry+screenshot；根回归 PASS）

## `BDD.RESUME.LIST.VISUAL.003` 简历列表参考稿层级

- [x] 确认验证入口为 Resume list/CSS domain behavior tests 与 current-run Chrome UI acceptance，不创建 E2E wrapper。
- [x] 执行 owner tests，验证 desktop 双列等宽卡、mobile 满宽单列、与 Workspace 一致的 22px circled-plus create icon、文件 icon、名称/摘要、meta、删除与 footer 打开层级。（验证：create icon RED 后 focused 8 tests PASS；其余 Resume list/CSS owner tests PASS）
- [x] 执行 create/open/archive success/failure 与 loading/empty/error/pagination 回归，确认 route/generated client/closed DTO 不变。（验证：owner scope 24 files / 150 tests PASS）
- [x] 记录 1916×821 / 390×844 bbox、circled-plus 一致性、截图、keyboard、theme、console 与 no-overflow 证据。（验证：Resume/Workspace 图标均为 22×22、同一 viewBox/圆/十字路径/1.8 线宽；desktop 双列 690px，mobile 358px 单列，overflow 0；截图位于 `.test-output/list-ui-acceptance/`）

## `BDD.RESUME.DETAIL.VISUAL.004` 简历预览目标构图

- [x] Owner component/CSS tests 先 RED 后 GREEN，覆盖 Header 层级与 `1512/1310/1150px` desktop 内容面、背景板、白色纸张。<!-- verified: 2026-07-19 evidence="Resume detail component/CSS gates PASS within owner 32 files/242 tests" -->
- [x] PDF/Markdown renderer、无 header metadata 注入、无 actions/tabs/native viewer 与 mobile no-overflow 回归保持通过。<!-- verified: 2026-07-19 evidence="root frontend 132 files/1057 tests PASS; exact mobile overflowX=0" -->
- [x] Chrome skill 在真实 frontend/backend 上记录 desktop/mobile bbox、截图、主题、console 与 no-overflow；该 scoped UI evidence 不声明 E2E PASS。<!-- verified: 2026-07-19 evidence="desktop board/paper=1310/1150; mobile board/paper=358/332; screenshots captured and browser finalized" -->

### Phase 25 shared backdrop removal

- [x] Owner component/CSS tests 证明正文直接 `article` 无 presentation class、共享 preview-card CSS 缺席，Markdown page/PDF stack 正向 renderer 仍在。<!-- verified: 2026-07-20 evidence="Focused 14/14 and full Resume owner 118/118 PASS; source parity rejects both shared preview-card selectors while renderer tests retain Markdown page and PDF stack." -->
- [x] Current-run real Chrome detail smoke 证明正文不再出现共享浅色背景板，页面无横向溢出且 console 无 error；该 scoped UI evidence 不声明 E2E PASS。<!-- verified: 2026-07-20 evidence="Final real Markdown detail: legacy selector count 0; classless article is 1443px with transparent background, border none, radius 0, shadow none and padding 0; inner page is 1150px, overflowX 0, browser warning/error empty, screenshot captured." -->

## `BDD.RESUME.PARSE.VISUAL.005` 简历解析等待态

- [x] Component/polling tests 覆盖 resume variant、连续 pending 请求间 DOM 稳定、ready/failed 原子替换、返回动作和无伪进度。
- [x] Current-run desktop Chrome 连续采样证明等待构图不闪烁、不缩放、TopBar/返回可用且全宽无横向溢出；mobile 由共享 responsive contract 覆盖，不新增 E2E ID。（scene x=0、width=viewport=1920。）

## `BDD.RESUME.DELETE.CONFIRM.006` 简历删除二次确认

- [x] 确认验证入口为 `ResumeListView.test.tsx` 与 shared destructive-dialog domain behavior tests，不创建 E2E wrapper。
- [x] 执行首次点击零 `archiveResume`、取消/Escape/遮罩零副作用、初始焦点/focus trap/trigger focus restore 行为断言。
- [x] 执行确认单次提交、pending 关闭/重复提交锁定、失败保留卡片/弹窗与同 key retry、成功隐藏卡片行为断言。
- [x] 使用 Chrome skill 在真实 frontend/backend 截取简历删除确认框并记录 keyboard、console 与 no-overflow；该截图证据不声明 E2E PASS。<!-- verified: 2026-07-20 method=chrome-extension-manual evidence="1212x912 real frontend/backend screenshot; cancel initially focused; cancel closed dialog, preserved 李四 card and restored delete-trigger focus; horizontal overflow false; screenshot=.test-output/delete-confirmation-ui/resume-delete-confirmation.png" -->
