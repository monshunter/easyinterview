# Settings Custom Theme Hierarchy 交付复盘报告

> **日期**: 2026-07-19
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 [BUG-0193](../bugs/BUG-0193.md)：Settings Appearance 的 Custom 面板覆盖一级主题选择器，导致用户无法回退到预定义主题；同时按用户补充参考把 hue/chroma 轨道改为可理解的彩色渐变。
- 关联 owner 为 [frontend-shell/001 Phase 19](../spec/frontend-shell/plans/001-app-shell-auth-settings/plan.md)。Phase 19 已完成，但 frontend-shell/001 因既有 Phase 15.3 auth locale Chrome 证据仍未完成而继续保持 `active`。
- RED 为 Settings behavior/visual focused `2 failed, 23 passed`；GREEN 为 `25/25`，并证明预定义主题回退保持零网络请求。
- `npm run typecheck`、production build、frontend redeploy、本地依赖 readiness 4/4 和根 `make test` 通过；根回归为 Python 615 tests / 4615 subtests、Go 全包和 frontend 全量。
- Chrome 在真实本地 Settings 页完成 `1440×900` 与 `390×844` 验收：一级三个按钮始终可见，Custom 面板严格位于下方，Custom -> Ocean -> Custom 可逆，hue/chroma 渐变与键盘 slider 联动正确，两档均无横向溢出或 browser error/warning。

## 2 会话中的主要阻点/痛点

- 现有 DOM 行为测试是绿色，但用户仍无法点击一级主题。
  - **证据**：源码和 jsdom 都能看到三个主题按钮；真实 CSS 却把 options 与 custom panel 分配到同一 `grid-row: 1 / span 2`，Chrome 中后者覆盖前者。
  - **影响**：如果只依赖节点存在、`aria-pressed` 与请求次数，视觉/可操作性缺陷会被错误收口。
- 上一轮 Settings 截图验收只覆盖总体横向功能卡，没有把 Custom 展开态作为独立状态验收。
  - **证据**：Phase 18 固定 Header、三张卡和 desktop/mobile no-overflow，但缺少 options-bottom 与 panel-top 的相对关系和 Custom -> preset 点击回路。
  - **影响**：collapsed/default 状态看起来正确，条件二级内容的覆盖问题直到用户实际切换后才暴露。
- 原生 range 灰色轨道虽然功能正确，但无法传达 hue/chroma 两个维度。
  - **证据**：用户补充全色彩参考；旧 CSS 只有 `accent-color`，轨道本身没有完整 hue 或当前 hue 下的 chroma 信息。
  - **影响**：用户需要试错才能理解拖动方向，视觉实现没有充分承接控件语义。

## 3 根因归类

- 条件二级内容与一级选择器共享 grid area，且缺少展开态结构/bbox gate，属于 `spec/plan`；本轮已在 frontend-shell spec、UI design、Phase 19 与 BDD 原地补齐。
- jsdom 不执行真实布局属于测试工具边界，类别为 `no repo change needed`；解决方式是把 source contract 与 current-run Chrome bbox/interaction 结合，而不是把更多视觉结论塞进 DOM 单测。
- 色条信息表达不足属于当前 UI design 细节缺口，类别为 `spec/plan`；本轮已将完整 hue 与 hue-aware chroma 写入 owner 合同和 executable CSS gate。

## 4 对流程资产的改进建议

- 对所有“一级选择 + 条件二级编辑”组件，在 owner plan 中固定常驻/挂载关系、正常文档流和可逆点击路径；展开态必须是 Chrome 验收状态之一。
  - **落点**：相关 UI owner 的 spec/plan/BDD
  - **优先级**：high
- 视觉验收遇到 grid/flex 条件内容时，component test 负责 DOM/状态，source test 负责禁止共享覆盖区域，Chrome 负责 bbox、可点击性和 overflow；三类证据不可互相替代。
  - **落点**：当前 plan 的 visual gate；后续同类 plan 复用
  - **优先级**：high
- 主题 range 继续保持 hue/chroma 两个维度，不为展示渐变重新引入 value、preview 或 reset 面板。
  - **落点**：frontend-shell spec/plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 本修复没有遗留实现或视图验收缺口；最高价值的下一步仍是继续 `/implement` frontend-shell/001 Phase 15.3，捕获 auth loading/error 的中英文真实 Chrome 状态并关闭当前 owner 唯一未完成项。
- 后续新增任何可展开设置控件时，优先复用本次“常驻一级 + 正常流二级 + preset 可逆”的 parent-owned 结构和三层验证方式。
- 不建议为色条增加数值或预览卡；当前渐变已经在不扩大信息层级的前提下传达两个调节维度。
