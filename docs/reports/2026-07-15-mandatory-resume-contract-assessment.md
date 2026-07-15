# Mandatory Resume Contract 交付复盘报告

> **日期**: 2026-07-15
> **审查人**: Codex

**关联计划**: [Home JD Import and Parse](../spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md)

## 1 复盘范围与成功证据

本次交付在原 owner 中修订产品范围、Home/Resume/Workspace UI 设计与 `frontend-home-job-picks-and-parse/001` 的 spec、plan、checklist、BDD 和 context，将 selectable Resume 固化为当前及未来 JD import、模拟面试、报告与报告后动作的强制前置，并删除无简历/JD-only 降级承诺。

成功证据：

- Home/Parse focused Vitest 5 个文件、38 个测试通过，覆盖未选择简历时零 import、selectable 简历筛选、绑定读取和缺绑 fail-closed。
- 正式 Home 使用 `isSelectableInterviewResume`；其合同是未归档且 `parseStatus=ready` 或已有可读正文/结构化证据，`canSubmit` 与提交处理器均要求已显式选择。
- OpenAPI `ImportTargetJobRequest` 继续要求 `resumeId`，本次无 API 或代码变更。
- 旧承诺负向搜索为零；`validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check` 和 `make lint-core-loop-pruning-surface` 全部通过，pruning `real_residuals=0`。

## 2 会话中的主要阻点/痛点

- “强制简历”一度被误缩窄为“必须 `parseStatus=ready`”。
  - **证据**：正式 `isSelectableInterviewResume` 还允许已有 `parsedTextSnapshot`、`originalText` 或 structured profile 的非 ready 简历；`HomeResumeSelection.test.tsx` 明确覆盖 readable non-ready Resume。
  - **影响**：第一轮文档修订需要再次校正，若未反查代码，会制造比旧承诺更严格的新产品漂移。
- active 文档同时使用 `ready Resume`、`usable Resume` 和“可选简历”，但没有统一定义 selectable 语义。
  - **证据**：Home owner 与 Resume owner 原有文档分别使用 ready 和 readable-evidence 口径；正式筛选函数名已使用 selectable。
  - **影响**：后续设计评审容易把“必须有简历”与“必须等待完整解析成功”混为一谈。

## 3 根因归类

- 强制前置与可选择状态是两个不同维度：前者回答是否允许无简历训练，后者回答哪些已有简历具备足够证据。
  - **类别**：spec/plan
- 正式代码中的 `readyResumes` 状态变量仍承载 selectable 集合，命名弱化了 `isSelectableInterviewResume` 的真实合同。
  - **类别**：spec/plan；后续代码实施可在原 owner 中处理
- 本次 owner 发现、分支门禁、文档同步与验证流程正常工作。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 后续所有产品/UI/plan 文档统一使用 `selectable Resume`，首次出现时定义为“未归档且 ready 或已有可读正文/结构化证据”；“强制简历”只表达禁止空绑定。
  - **落点**：spec/plan
  - **优先级**：high
- 在原 Home owner 中将内部 `readyResumes` 等容易误导的变量/描述收敛为 selectable 语义，并保留 readable non-ready 回归测试；不改变业务行为。
  - **落点**：spec/plan，随后由 `/implement` 执行代码命名清理
  - **优先级**：medium
- 不新增通用治理规则或 sibling plan；D-30、D-18 和 Phase 24 已足以承接当前产品不变量与负向门禁。
  - **落点**：无需仓库改动
  - **优先级**：low

## 5 建议优先级与后续动作

下一步最高价值动作是先保持本次设计合同为唯一真理源；如继续降低实现语义歧义，使用 `/change-intake` 原地重开 `frontend-home-job-picks-and-parse/001`，仅清理 `readyResumes` 等误导命名并由 `/implement` + `/tdd` 保持 38 个现有行为断言不变。当前无须新增 E2E、API 兼容层或无简历场景。
