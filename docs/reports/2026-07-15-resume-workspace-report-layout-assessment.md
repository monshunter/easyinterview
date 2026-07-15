# Resume、Workspace 与 Report 布局交付复盘报告

> **日期**: 2026-07-15
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次在原 owner 内实施三个相邻 UI 目标：Resume 列表响应式卡片网格、Workspace detail 标题旁绑定简历入口与首行动作行、Report ready `3/2/2/2/1` 与底部全宽“面试总评”。
- 关联计划：
  - [Resume listing owner](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)
  - [Workspace shared Parse owner](../spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md)
  - [Workspace acceptance owner](../spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md)
  - [Report dashboard owner](../spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/plan.md)
- TDD 证据包括 Resume focused RED 7 项失败后 GREEN、Workspace shared Parse RED/GREEN、Report 顶部三指标 RED 后 GREEN，以及 P0.099 visual-audit validator 的新增 RED/GREEN。
- 当前根 `make test` 通过：backend 551 tests / 4493 subtests，frontend 125 files / 993 tests；frontend lint、typecheck/build、四个 owner context、文档 Header/INDEX、Markdown links 与 `git diff --check` 均通过。
- Chrome 真实 host-run 验收覆盖 1440x1200 与 390x844：Resume 桌面 360px 固定列与 mobile 单列；Workspace saved resume 精确跳转、首行按钮顺序与旧 block 零残留；Report desktop 计数 `3/2/2/2/1`、总评 `grid-column: 1 / -1`、mobile 单列且无横溢。浏览器控制台错误为 0。
- Resume、Home/Parse 与 Workspace owner 已恢复 `completed`。Report Phase 12 代码、响应式与 BDD gate 已通过，但 P0.099 当前 exact-six/provider 场景未执行，因此 Report owner 保持 `active`，未复用历史 PASS。

## 2 会话中的主要阻点/痛点

- 根回归在 focused PASS 后发现 4 个 orphan locale key。
  - **证据**：首次 `make test` 的 `localeFiles.test.ts` 报出 `parse.resumeEmptyBody`、两个旧 table column key 与旧顶部 readiness key；删除后第二次根回归通过。
  - **影响**：如果只采用 focused owner tests，会错误保留已删除 UI 的中英文合同。
- Chrome full-page capture 会保留 sticky header 所在滚动位置，导致一张 report mobile 截图顶部覆盖，不能作为视觉 PASS 主证据。
  - **证据**：390px DOM/bbox/no-overflow 均通过，但该 PNG 顶部标题受 sticky header 覆盖；Resume/Workspace mobile 与 Report desktop 图片正常。
  - **影响**：几何事实与截图可用性可能被混为一谈；直接接受图片会形成虚假视觉闭环。
- 同一用户目标横跨四个 owner plan，但 Report 的 P0.099 exact-six/provider gate 比普通 Chrome UI 验收更重。
  - **证据**：前三个 owner 已无未完成项；Report checklist 仍明确要求当前六图、live API/DB 与 no-OCR audit，普通 Chrome 截图不能替代。
  - **影响**：若不区分 code/UI acceptance 与独立 provider E2E，容易错误恢复 Report `completed`，或反向阻塞已完成的 Resume/Workspace 生命周期。

## 3 根因归类

- orphan locale key 属于跨 owner 完整回归才能发现的残留；现有根门禁工作正常，类别为 **无需仓库改动**。
- full-page sticky capture 缺少明确的“截图前回到文档顶部并复核首屏”步骤，类别为 **README**。
- Report code/UI gate 与 P0.099 provider exact-six gate共处同一 Phase closeout，证据层级虽已写明，但执行状态仍容易被合并理解，类别为 **spec-plan**。
- 初次 focused command 的 `pnpm --dir` 参数顺序错误是一次性执行失误，未造成代码或证据误判，类别为 **无需仓库改动**。

## 4 对流程资产的改进建议

- 在 P0.099 README 的 browser capture 步骤中增加截图前 scroll-top、刷新 DOM/bbox、检查 sticky header 不覆盖标题的明确 preflight。
  - **落点**：`test/scenarios/e2e/p0-099-report-generating-live-ui/README.md`
  - **优先级**：high
- 在 Report Phase 12 checklist 中把“正式 frontend/Chrome layout acceptance 已通过”和“current provider exact-six scenario PASS”作为两个显式状态字段或分段 closeout，仍保持只有后者完成后才能恢复 plan `completed`。
  - **落点**：`frontend-report-dashboard/001` plan/checklist
  - **优先级**：medium
- 保留 locale reachability 在根 `make test` 中的现有位置，不复制到 E2E 或每个 focused owner gate。
  - **落点**：无需仓库改动
  - **优先级**：low

## 5 建议优先级与后续动作

- 下一步最高价值动作是用 `/scenario-run -i E2E.P0.099` 生成当前 exact-six/provider 证据，并在同一 run 中完成 no-OCR manual audit；通过后再完成 Report 12.5/12.6 与生命周期恢复。
- 若下一轮先改流程资产，优先补 P0.099 的 scroll-top/sticky-header capture preflight，避免重新生成六图后才发现图片不可验收。
- Resume、Home/Parse、Workspace 不需要新增 sibling plan；后续只在出现新回归时由原 owner 原地承接。
