# 核心工作流参考图对齐交付复盘报告

> **日期**: 2026-07-19
> **审查人**: Codex

**关联 Bug**: [BUG-0192](../bugs/BUG-0192.md)

**关联计划**:

- [Resume Create Flow](../spec/frontend-resume-workshop/plans/002-create-flow/plan.md)
- [Resume Listing and Readonly Detail](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)
- [Practice Text Event Loop](../spec/frontend-workspace-and-practice/plans/002-practice-text-event-loop/plan.md)
- [Report Screen and Generating Handoff](../spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/plan.md)

## 1 复盘范围与成功证据

- 本次交付按三张参考图重构上传简历、面试进行和面试报告三页的导航下方起点、desktop 内容面、响应式网格、卡片层级、按钮、SVG icon、字体与间距，同时保持既有 API、路由、消息、简历注册和报告事实语义。
- 用户追加要求已闭环：AI/用户角色改为方形标识；说明胶囊带 sparkle icon；Transcript 成为唯一滚动区；Composer 整体固定在会话卡底部；说明胶囊固定贴在输入框上方 8px。
- 三张追加参考图也已闭环：简历预览采用 `1512/1310/1150px` Header/背景板/纸张构图；报告列表采用 `1372px` 插画 Header、事实摘要卡和编号时间线；面试记录采用同宽三列 Context Strip，并让 AI/“我”共用消息卡片与头像轮廓。
- TDD 证据：Resume create 4 files / 21 tests；Practice 最终 4 files / 56 tests；报告页追加 2 files / 23 tests 的目标构图 RED/GREEN，并由根级 frontend 131 files / 1055 tests 完整覆盖。报告页 RED 明确拒绝缺失 Detail icon 与 1432px 目标构图，随后转 GREEN。
- 根回归：Python 615 tests / 4615 subtests、Go 全包、frontend 132 files / 1057 tests 全部通过；三页 owner 32 files / 242 tests、frontend typecheck 与 production build 通过。
- Chrome 证据：Resume 使用正式 real-mode frontend 完成 1916×821 / 390×844 containment；Practice 的 Transcript 在 desktop/mobile 实际滚动前后，input 坐标与 helper/input 8px gap 不变。报告页使用当前真实 backend ready report 完成 desktop 1920×964 与 exact mobile 390×844 full-page 验收：desktop 主体 1432px、四列 Context、两列 709px 内容卡、四个 46px Detail icon 和首屏完整 Overall 均闭合，两档 document overflow 均为 0。
- 文档证据：三组原 owner spec/plan/checklist/BDD/test 原地修订并恢复 `completed`；context、Header/INDEX、docs links 与 `git diff --check` 通过。

## 2 会话中的主要阻点/痛点

### 2.1 参考图一致不等于滚动行为一致

- **证据**：第一轮将说明胶囊放到 Transcript 最后一项，标准截图视觉接近参考图，但用户指出聊天增长后它会随记录移动；进一步要求输入框本身也必须始终贴底。
- **影响**：仅验收最终静态截图会漏掉内容增长后的相对定位，导致一次明确返工。

### 2.2 真实环境健康与完整视觉状态不是同一类证据

- **证据**：正式 real-mode frontend/backend 健康，但当前数据库没有可复用的 active session / ready report；Practice / Report 的完整视觉状态只能使用 repository fixture 验收。
- **影响**：若不显式区分，容易把 fixture 视觉验收错误描述成真实业务 E2E PASS。

### 2.3 通用 secrets scanner 被既有 Skill 文本阻断

- **证据**：`make lint` 唯一失败来自未改动的 `.agent-skills/change-intake/SKILL.md:120`，`generic-api-key` 规则把普通英文示例短语误判为密钥；其余 lint owner、Go lint 和 frontend lint 全部通过。
- **影响**：完整 lint 聚合无法给出绿色结果，需要额外拆分运行并解释现有误报。

### 2.4 报告页第一轮视觉 gate 允许“旧结构换皮”假绿

- **证据**：Phase 16 只锁定 1336px 宽度、背景、圆角和部分语义图标；用户再次指出报告页完全没有按目标稿改造，并用标框明确 Header CTA、单体 Context Strip 和贯穿各卡片的左侧语义图标轨。源码反查确认旧的四卡 Context 和无 Detail icon 结构仍在。
- **影响**：历史 focused/root/fixture 结果全部为绿，但没有回答目标图的组件结构问题，造成一次完整返工和用户二次纠正。

### 2.5 面试记录角色样式被错误设计为非对称层级

- **证据**：初次追加改造让 AI 消息直接显示在主体背景中，只给“我”增加独立 surface，并让“我”的头像轮廓与 AI 不同；用户明确要求两种消息与头像轮廓保持一致。
- **影响**：共同组件没有形成共同视觉合同，导致一次可避免的细节返工；若只看页面整体截图，仍可能漏掉角色间的 computed-style 差异。

## 3 根因归类

- 固定 Composer / 滚动 Transcript 最初没有成为可执行 DOM/CSS/BBox 不变量，只描述了视觉形态。
  - **类别**：spec/plan
- Chrome 验收入口缺少对“real-mode health”“fixture visual state”“real API/UI E2E”三类证据的统一命名，容易产生边界混淆。
  - **类别**：README
- secrets scanner 对治理 Skill 的普通示例文本产生稳定误报，属于 lint owner 与 Skill 文本的契约漂移，不是本次 UI 实现缺陷。
  - **类别**：skill
- Chrome finalize 参数的两次格式尝试没有影响产品结果，属于一次性工具调用失误。
  - **类别**：无需仓库改动
- 报告页视觉合同把 max-width/rounded/overflow 当成主要完成证据，没有将参考图标框转换为共享 surface、divider、icon rail 和首屏 Overall 的结构性 RED。
  - **类别**：spec/plan
- 面试记录 visual gate 只验证用户消息 surface，没有对 assistant/user 的边框、圆角、padding、阴影和头像几何做成对称断言。
  - **类别**：spec/plan

## 4 对流程资产的改进建议

- 对所有聊天类页面，owner spec/plan 必须显式写出“唯一滚动区、固定区、DOM ownership、短/长内容、滚动前后 bbox 不变量”，并把这几项同时落到 source/component 与 Chrome gate。
  - **落点**：相关 UI spec / plan template
  - **优先级**：high
- 在场景或前端验证说明中统一定义三档证据标签：`real-mode health`、`formal frontend fixture visual acceptance`、`real API/UI E2E`，结果中必须逐档声明，不允许相互替代。
  - **落点**：`test/scenarios/README.md` 或 frontend verification README
  - **优先级**：medium
- 由 secrets-config / skills owner 单独处理 `.agent-skills/change-intake/SKILL.md:120` 的稳定误报：优先重写示例短语或增加精确、最小 allowlist，并保留真实 generic-key negative fixture。
  - **落点**：change-intake Skill + secrets lint owner
  - **优先级**：medium
- UI 参考图修订必须先列出可观察的组件所有权与几何关系，再以 source/component RED 拒绝旧结构；宽度、背景、圆角和无溢出只能是辅助 gate，不能单独作为完成条件。本轮已在原 report spec/plan Phase 17 固化该规则。
  - **落点**：相关 UI spec / plan
  - **优先级**：high
- 聊天记录的角色 visual contract 必须同时断言“共享基类 surface”和“仅身份 token 不同”：两类卡片的 border/radius/padding/shadow 相同，两类头像的 width/height/radius 相同。本轮已在 report owner Phase 18 原地固化。
  - **落点**：ReportConversation owner spec / plan / source contract test
  - **优先级**：high

## 5 建议优先级与后续动作

- 下一轮最高价值动作：把“参考图标框 → 组件所有权/几何关系 → source/component RED → real Chrome bbox”的闭环补入 UI plan 模板或前端验证规范，避免再次出现宽度和圆角已变、页面结构仍未改造的假绿。
- 同一优先级的后续项：把“滚动内容 + 固定操作区”的结构不变量补入聊天类 UI owner，覆盖短/长内容与滚动前后 bbox。
- 第二优先级：单独修复 secrets scanner 的现有治理文本误报，使根 `make lint` 恢复可直接作为聚合门禁使用。
- 可延后：为 fixture visual acceptance 增加更多长内容 fixture；当前通过压缩 desktop/mobile viewport 已真实产生 Transcript overflow 并验证了固定 Composer，不构成本次交付 blocker。
