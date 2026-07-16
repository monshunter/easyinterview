# Interview Plan Card Metadata 交付复盘报告

> **日期**: 2026-07-17
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付原地修订 `frontend-workspace-and-practice/001-workspace-and-interview-context`，移除 Home/Workspace 共享面试规划卡片中的两处 TargetJob lifecycle status，并在地点缺失、空或仅空白时省略地点行。
- 后端 `TargetJob.status`、OpenAPI、generated types、状态机、归档行为与 backend-persisted `practiceProgress` 均未改变。
- TDD 证据：新增行为测试先稳定复现状态可见和 `Location not set` 两项失败；Green 后 card tests 13/13、共享 Home/Workspace tests 41/41。
- 全量证据：仓库根 `make test` 通过，后端 584 tests / 4583 subtests、前端 126 files / 1027 tests；frontend typecheck、owner context、文档索引、docs-check、diff 与源代码残留搜索均通过。
- BDD 证据：`BDD.WORKSPACE.CARD.003` 由 `MockInterviewCard.test.tsx` domain behavior test 验证，不声明真实 E2E PASS。

## 2 会话中的主要阻点/痛点

- 卡片把同一个 `TargetJob.status` 以 company eyebrow 后缀和右侧 badge 重复展示。
  - **证据**：修复前组件的两个位置都消费 `statusLabel(job.status)`；新负向断言 Red 时仍找到状态文案。
  - **影响**：用户把长期默认的 `Draft` 误解为规划生成状态，并且卡片首屏信息密度被无效元信息占用。
- lifecycle status 数据合同与当前 UI 职责已经分离，但早期视觉合同仍要求“公司/状态 eyebrow”。
  - **证据**：当前训练进度只由 `practiceProgress` 和 round rail 表达；正式前端没有生产 `updateTargetJob` 状态维护入口。
  - **影响**：单看 API enum 容易误判该字段仍是当前页面的用户决策信息。
- 当前工作树已有大量其他未提交修改，且 UI design owner 文件存在重叠改动。
  - **证据**：首次分支门禁发现当前分支与计划 metadata 不同、工作树 dirty；用户明确授权在当前分支叠加后才继续。
  - **影响**：必须采用文件级、行级最小补丁，并在全量 gate 中区分本次修改与并行工作。

## 3 根因归类

- **spec-plan**：旧卡片合同把不可操作的 lifecycle status 继续列为可见信息，没有与当前“面试规划而非完整岗位生命周期管理”职责同步收敛。
- **spec-plan**：地点缺失行为未明确，组件测试反而把英文 fallback 固化成正向预期。
- **无需仓库改动**：当前脏工作树是本次会话的已知协作条件；分支门禁正确阻止了未授权切换，用户授权后通过外科式补丁完成交付，没有暴露新的治理缺口。

## 4 对流程资产的改进建议

- 后续修改共享业务卡片时，先在 owner spec 中区分“API 可用字段”和“当前页面可操作/可决策字段”。
  - **落点**：对应 feature 的 spec/plan。
  - **优先级**：medium。
- 可选字段的 UI 行为应同时定义有值与缺失值，不默认引入英文或技术型 fallback。
  - **落点**：对应 UI design 文档与 component behavior test。
  - **优先级**：medium。
- enum 元信息展示应使用非默认值做负向测试，防止 fixture 长期固定在首个 enum 时掩盖重复或误导展示。
  - **落点**：owner component tests。
  - **优先级**：low。

## 5 建议优先级与后续动作

- 本次 owner 文档、组件与测试已同步，不需要新增 sibling plan 或修改 OpenAPI。
- 下一步优先在当前分支后续整体 closeout 时，把 `BUG-0184` 与本次 Phase 29 文档/代码一起纳入同一语义 commit，避免 Bug 索引先于实际提交分离。
- 如后续产品确实引入投递状态管理，应先重新设计可操作入口和状态更新流程，再决定在哪个页面恢复 lifecycle status；不应仅恢复只读 badge。
