# Home 空规划隐藏最近模拟面试区块交付复盘

> **日期**: 2026-07-20
> **审查人**: Codex

**关联计划**: [001 Home + JD Import + Parse](../spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md)

## 1 复盘范围与成功证据

- 本次原地重开 `frontend-home-job-picks-and-parse/001-home-jd-import-and-parse` Phase 29：已认证用户成功加载但没有可展示面试规划时，Home 不再渲染“最近模拟面试”整个区块；loading、error 与 1～3 条记录保持既有行为。
- TDD 证据：新增 empty 行为断言在旧实现上形成唯一 RED；GREEN 后 `HomeRecentMocks.test.tsx` 14/14、Home owner 9 files / 68 tests 通过。
- 聚合证据：根 `make test` 通过 Python 615 tests / 4615 subtests、Go 全包与 frontend 136 files / 1113 tests；frontend typecheck、production build、context、Header/INDEX、docs links 与 diff gate 通过。
- `BDD.HOME.RECENT.EMPTY.007` 由 domain behavior test 承接，没有新增或冒充真实 E2E。

## 2 会话中的主要阻点/痛点

- `/change-intake` matcher 对“模拟面试 / mock”给出低置信度的 `mock-contract-suite` 候选，而真实 owner 是 Home/Parse 计划。
  - **证据**：matcher 推荐分数为 7 的 mock runtime；随后源码、UI 文案、工作日志和既有 Home recent owner 引用一致指向 `frontend-home-job-picks-and-parse/001-home-jd-import-and-parse`。
  - **影响**：增加一次 live repo search，但没有造成错误文档或代码写入。
- active → completed 的计划生命周期会让 plans INDEX 在执行中短暂出现预期 drift。
  - **证据**：`sync-doc-index --check` 只报告同一 owner 的 4 个 group drift；两次 deterministic fix 分别完成重开和收口投影。
  - **影响**：属于既有流程成本，没有阻断实现或产生人工判断风险。

## 3 根因归类

- matcher 的通用英文 token 对 mock tooling 权重偏高，但现有 `/change-intake` 已明确规定低置信度必须用 live plan/spec/code/Git 证据复核。
  - **类别**：无需仓库改动。
- 本次不是实现回归：旧 spec 与测试明确要求 empty state，用户提出的是新的用户可见行为，因此需要先原地修订 UI design、spec、plan/checklist 与 BDD。
  - **类别**：spec-plan；已在本次交付中完成修订。

## 4 对流程资产的改进建议

- 保持现有低置信度 live-search 规则，不因一次误排序修改 Skill；若后续再次出现“页面文案被 mock/tooling token 抢占 owner”的同类误配，再为 matcher 增加 UI label、route 与 `frontend/src/app/screens/*` 命中权重。
  - **落点**：`change-intake` matcher script
  - **优先级**：low
- 不新增 Bug 记录。本次行为与旧冻结合同冲突，属于经用户确认的 feature revision，不是代码偏离既有设计。
  - **落点**：无需仓库改动
  - **优先级**：low

## 5 建议优先级与后续动作

- 当前交付无需追加治理修订；最高价值的下一步是合并前审查本分支尚未进入 `main` 的整组 UI 对齐提交，确认 Phase 29 与前序 Home/Workspace/Resume 视觉改动作为同一分支历史可安全集成。
- matcher 权重优化暂缓，只有同类低置信度误配再次发生时再投入，避免为单次可恢复排序增加规则复杂度。
