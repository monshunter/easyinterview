# Workspace 与 Resume 列表视觉对齐交付复盘报告

> **日期**: 2026-07-19
> **审查人**: Codex

**关联计划**: [Workspace + InterviewContext + Start Practice](../spec/frontend-workspace-and-practice/plans/001-workspace-and-interview-context/plan.md)；[Resume Listing Routing and Detail Readonly](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)

## 1 复盘范围与成功证据

本次交付把 query-free Workspace 面试规划列表和 Resume Workshop 简历列表按用户提供的桌面参考稿收敛到正式前端，并在同一轮反馈中完成三项补充校正：简历列表 desktop 每行两张等宽卡；Workspace 背景覆盖完整 viewport；Workspace 新建按钮与第二列卡片右侧对齐；Resume 新建按钮使用 Workspace 同款圆圈加号。

成功证据：

- TDD 证据：初始 Workspace 视觉合同 4 项预期失败；Resume 双列合同 2 项预期失败；Workspace 全宽背景和 CTA 右边界合同各 1 项预期失败；Resume create icon 以实际 `width=14` 预期失败。最终 owner scope 24 files / 151 tests PASS。
- 代码回归：frontend typecheck 和 production build PASS；仓库根 `make test` 为 615 tests / 4615 subtests PASS，Go 全包 PASS。
- Chrome 桌面证据：Workspace 在 1916×821 与 2048×917 下背景左右边界差值均为 0；1916×821 下 CTA 与第二列卡片 `right=1660`、差值 0px，两张卡均为 714×384。Resume 在 1916×821 下两张卡宽 690px、间距 28px。
- Chrome 移动端证据：390×844 下 Workspace 与 Resume 网格均为 358px 单列，document overflow 为 0。
- 跨页面视觉证据：Resume 与 Workspace 创建图标均为 22×22、`viewBox="0 0 24 24"`、`strokeWidth=1.8`、圆 `r=9` 和路径 `M12 8v8M8 12h8`。
- 工程门禁：两份 owner context validator、`sync-doc-index --check`、`make docs-check`、`git diff --check` 和 local-dev 4/4 dependency verify 均通过；截图为真实 JPEG，位于 `.test-output/list-ui-acceptance/`。

本轮没有创建或运行真实业务 E2E scenario；Chrome 结果只作为正式前端 UI 验收证据。归档按钮没有在浏览器中点击，以避免改变本地业务数据；成功/失败行为由 owner component tests 覆盖。

## 2 会话中的主要阻点/痛点

### 2.1 单条参考画面不足以推断列表列数

- **证据**：最初 Resume 参考画面只显露单条内容，后续用户明确补充 desktop 应为每行两张卡；对应 CSS 合同补充后旧单列规则出现 2 项预期失败。
- **影响**：列表几何发生一次返工，且必须重新执行 desktop/mobile 验收。

### 2.2 背景声明为 100vw 仍被限宽容器裁剪

- **证据**：旧 `.ei-workspace-plan-list` 同时持有 `max-width: 1508px`、`overflow: hidden` 和 100vw 伪元素；用户截图显示内容区右侧为空白带。拆分全宽 canvas 与居中 inner 后，Chrome 实测背景为 `x=0/right=viewportWidth`。
- **影响**：源码静态看到 100vw 容易产生“已全宽”的误判，必须用 live bbox 和最终截图才能发现裁剪边界。

### 2.3 相同创建动作的图标只做了局部还原

- **证据**：Workspace 使用 22px 圆圈加号，Resume 使用 14px 裸加号；用户再次指出跨页面不一致。补充断言后旧实现精确失败于 `width=14`。
- **影响**：单页本身可用，但产品级视觉语言仍不统一，需要一次额外实现、部署和 Chrome 对比。

### 2.4 一次校验命令使用了迁移前路径

- **证据**：首次 context gate 调用 `.agent-skills/implement/scripts/validate-context.py` 返回文件不存在；仓库反查后改用 `.agent-skills/implement/shared/scripts/validate_context.py`，两份 context 均 PASS。
- **影响**：只增加了一次短暂重跑，没有造成实现或文档误判。

## 3 根因归类

- Resume 列数未先显式化：根因属于 `spec/plan`。外部截图提供了风格和局部几何，但没有可靠表达多数据项时的网格规则；用户补充后已在 Resume spec/plan/UI owner 中固化双列合同。
- Workspace 背景断层：根因属于 `spec/plan` 与实现级视觉 gate。原合同关注内容区宽度和卡片 bbox，未把 canvas 的左右边界作为独立不变量。
- 创建图标漂移：根因属于 `spec/plan`。两个 sibling 页面分别实现相同语义动作，却没有把 SVG geometry 纳入跨页面一致性验收。
- validator 路径错误：归类为 `无需仓库改动`。当前 Skill 和仓库测试已经声明正确 shared 路径，本次为一次执行失误，未形成重复模式。

## 4 对流程资产的改进建议

- 在下一次截图对齐 plan 的视觉合同中先写出“viewport canvas、content container、grid、alignment anchor、sibling semantic action”五类明确坐标，不从单条 fixture 或截图可见项数量推断列表规则。
  - **落点**：`spec/plan`
  - **优先级**：high
- 对 full-bleed 页面把 `canvas.x=0`、`canvas.right=viewportWidth`、`scrollWidth=clientWidth` 作为同一 Chrome gate；仅检查 `width: 100vw` 源码不足以证明视觉覆盖。
  - **落点**：`spec/plan`
  - **优先级**：high
- 后续若第三个页面继续使用同一创建动作，评估提取共享 `CreateActionIcon`，并保留 SVG 参数一致性测试，避免 circle/path/size 再次分叉。
  - **落点**：frontend implementation follow-up
  - **优先级**：medium
- validator 路径错误无需新增规则；继续以 implement-owned Skill 当前命令和仓库反查为准。
  - **落点**：no repo change needed
  - **优先级**：low

## 5 建议优先级与后续动作

下一轮最值得优先执行的是：若继续对齐 Workspace/Resume 详情页，在编码前先把全屏背景、内容宽度、卡片列数、关键右边界和跨页面共用动作列成可量测合同，再以 1916×821 和 390×844 同时验收。当前两个列表页不需要再做结构重构。

可以延后评估共享 `CreateActionIcon`；当前仅两个 consumer，现有精确参数测试已经能阻止本轮同类漂移，不宜为了抽象而扩大本次交付。
