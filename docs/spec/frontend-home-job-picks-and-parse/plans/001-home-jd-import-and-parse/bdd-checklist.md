# Home JD Import and Parse BDD Checklist

> **版本**: 2.32
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.HOME.JD.001` JD 导入与解析交接

- [x] Owner behavior tests 覆盖 import、parse、失败恢复、确认 handoff 与隐私。
- [x] 无 selectable Resume 或未显式选择时「立即面试」禁用且零 import；ready 或已有可读证据的未归档简历均可选择，只在显式选择后提交 exact request，不存在无简历训练分支。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] 当前无 JD import/parse 真实 E2E owner；P0.098 progress 场景不承接该行为。

## `BDD.HOME.JD.002` 共享 Workspace ready-detail 首行动作

- [x] 标题旁“绑定简历”只消费 saved `TargetJob.resumeId` 并打开对应 Resume 详情；缺失/无效绑定是异常状态，不链接、不伪造、不提供 rebind 或 fallback，Start/Reports/复练/下一轮全部 fail closed。
- [x] “立即面试 + 面试报告”在标题下左对齐首行动作行按序呈现；desktop 同排、mobile 同序换行，Start/Report 事实与错误边界不变。
- [x] 独立 launch/binding block、标题右侧 Report、页尾 Start 的 DOM/source 负向 gate 为零；根 `make test` 与独立 responsive/a11y gate 通过，不声明真实 E2E PASS。

## `BDD.HOME.JD.003` Home 视觉层级与响应式状态

- [x] Owner behavior tests 覆盖 Hero/subtitle/illustration、single intake card、runtime count、resume controls/CTA/privacy note、recent header/record/actions 与 loading/empty/error。
- [x] zh/en、light/dark、ocean/plum/customAccent、keyboard/focus、disabled/enabled 与 1~3 条动态 round rail 均由正式 component/style 断言覆盖。
- [x] Chrome `1916x821` 对照参考图并在 `390x844` 验证单列/no-overflow；截图与 console 结果仅作为人工视觉证据。
- [x] 根 `make test` 执行对应 Vitest；不新增或冒充真实 E2E。

## `BDD.HOME.PLAN.VISUAL.004` Workspace 详情目标构图

- [x] Visual/source tests 先锁定 `1250px` Header 右侧动作和基本信息、要求、隐性关注点、动态轮次四层卡面，并证明旧构图失败。
- [x] Parse/Workspace behavior tests 覆盖 saved/missing Resume、Start/Reports、合法/无效 progress、2~5 动态轮次、keyboard/touch 与请求/route 不变。
- [x] Chrome `1916x821` 对照参考稿并在 `390x844` 验证单列/no-overflow；截图与 console 结果只作 manual visual evidence，不新增真实 E2E。
- [x] 根 `make test`、typecheck/build、owner/docs/diff gates 通过后恢复 owner lifecycle。

## `BDD.HOME.JD.PARSE.VISUAL.005` JD 解析等待态

- [x] Owner tests 覆盖 shared job variant、四步 done/current/pending、真实 step label、ready replace、error/Back 和 internal-metadata negative。
- [x] Current-run desktop Chrome 对照参考稿验证 TopBar、JD/搜索插画、编号步骤轴和无横向溢出；mobile/reduced-motion 由 shared component contract 覆盖，不新增 E2E ID。（真实 step 1 与 step 4 均已捕获，最终到达 Workspace。）

## `BDD.HOME.JD.TEXTAREA.006` JD 输入区自动适配

- [x] Home layout/component tests 证明 212px 默认高度、100% 横向宽度、内部无纵向滚动，以及内容增长/删减时按最新 `scrollHeight` 增高/回缩。<!-- verified: 2026-07-20 method=focused-home-vitest evidence="All 9 Home files and 67 tests pass; HomeLayout directly proves 212px/100%/hidden-overflow plus controlled 420px growth and 224px shrink." -->
- [x] Chrome 在真实 desktop Home 验证空输入默认高度、粘贴多行长 JD 后完整可见、计数器/Resume/CTA 顺序稳定、`documentWidth=viewportWidth` 且 browser warning/error 为零；不新增 E2E ID。<!-- verified: 2026-07-20 method=chrome-extension-manual evidence="At 1916x821 the empty and shrunk textarea measured 212px; a 36-line JD grew to 993px with clientHeight=scrollHeight, width stayed 1346px inside a 1348px frame, documentWidth=viewportWidth=1916, Resume/CTA stayed ordered and console warnings/errors were zero. Evidence: .test-output/home-jd-textarea-acceptance/." -->
- [x] 根 `make test`、typecheck/build 与 owner/document gates 通过并记录证据。<!-- verified: 2026-07-20 method=full-regression evidence="Python 615/4615 subtests, Go all packages and frontend 134 files/1088 tests PASS; typecheck/build, redeploy, environment 4/4, owner context and doc/index/diff gates pass." -->
