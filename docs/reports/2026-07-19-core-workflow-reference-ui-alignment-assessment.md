# 核心工作流参考图对齐交付复盘报告

> **日期**: 2026-07-19
> **审查人**: Codex

**关联 Bug**: [BUG-0192](../bugs/BUG-0192.md)

**关联计划**:

- [Resume Create Flow](../spec/frontend-resume-workshop/plans/002-create-flow/plan.md)
- [Practice Text Event Loop](../spec/frontend-workspace-and-practice/plans/002-practice-text-event-loop/plan.md)
- [Report Screen and Generating Handoff](../spec/frontend-report-dashboard/plans/001-report-screen-and-generating-handoff/plan.md)

## 1 复盘范围与成功证据

- 本次交付按三张参考图重构上传简历、面试进行和面试报告三页的导航下方起点、desktop 内容面、响应式网格、卡片层级、按钮、SVG icon、字体与间距，同时保持既有 API、路由、消息、简历注册和报告事实语义。
- 用户追加要求已闭环：AI/用户角色改为方形标识；说明胶囊带 sparkle icon；Transcript 成为唯一滚动区；Composer 整体固定在会话卡底部；说明胶囊固定贴在输入框上方 8px。
- TDD 证据：Resume create 4 files / 21 tests；Practice 最终 4 files / 56 tests；Report 30 tests 全部通过。Practice helper ownership 的 RED 为 3 个预期失败，移动后转 GREEN。
- 根回归：Python 615 tests / 4615 subtests、Go 全包、frontend 131 files / 1054 tests 全部通过；frontend typecheck 与 production build 通过。
- Chrome 证据：Resume 使用正式 real-mode frontend 完成 1916×821 / 390×844 containment；Practice / Report 使用正式 frontend repository fixture 展示完整视觉状态。Practice 在 desktop/mobile 的 Transcript 实际滚动前后，input 坐标与 helper/input 8px gap 不变，document overflow 为 0。
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

## 3 根因归类

- 固定 Composer / 滚动 Transcript 最初没有成为可执行 DOM/CSS/BBox 不变量，只描述了视觉形态。
  - **类别**：spec/plan
- Chrome 验收入口缺少对“real-mode health”“fixture visual state”“real API/UI E2E”三类证据的统一命名，容易产生边界混淆。
  - **类别**：README
- secrets scanner 对治理 Skill 的普通示例文本产生稳定误报，属于 lint owner 与 Skill 文本的契约漂移，不是本次 UI 实现缺陷。
  - **类别**：skill
- Chrome finalize 参数的两次格式尝试没有影响产品结果，属于一次性工具调用失误。
  - **类别**：无需仓库改动

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

## 5 建议优先级与后续动作

- 下一轮最高价值动作：把“滚动内容 + 固定操作区”的结构不变量补入 UI plan 模板或前端验证规范，避免其他聊天、报告记录、长表单页面重复出现静态截图通过但动态体验漂移。
- 第二优先级：单独修复 secrets scanner 的现有治理文本误报，使根 `make lint` 恢复可直接作为聚合门禁使用。
- 可延后：为 fixture visual acceptance 增加更多长内容 fixture；当前通过压缩 desktop/mobile viewport 已真实产生 Transcript overflow 并验证了固定 Composer，不构成本次交付 blocker。
