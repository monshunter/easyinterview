# Home Resume Picker 交付复盘报告

> **日期**: 2026-07-08
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：修复 Home `选择已有简历` 下拉框在 `listResumes` 已返回可读简历时仍显示空态并禁用的问题，并同步 Parse 页共用同一可选简历判断。关联 Bug：[BUG-0141](../bugs/BUG-0141.md)。
- 成功证据：
  - Red 阶段：`corepack pnpm --filter @easyinterview/frontend test src/app/screens/home/HomeResumeSelection.test.tsx` 失败，确认 `home-resume-select` 被 disabled，且 `home-resume-empty` 仍显示。
  - Focused green：`corepack pnpm --filter @easyinterview/frontend test src/app/interview-context/selectableResume.test.ts src/app/screens/home/HomeResumeSelection.test.tsx src/app/screens/parse/ParseResumeBinding.test.tsx` 通过。
  - Home/Parse 回归：`corepack pnpm --filter @easyinterview/frontend test src/app/screens/home src/app/screens/parse/ParseResumeBinding.test.tsx src/app/interview-context/selectableResume.test.ts` 通过，10 files / 66 tests。
  - Browser screenshot：`.test-output/screenshots/home-resume-picker-fixed-2026-07-08.png` 显示 Home 下拉已选中已有简历；Playwright 断言 `optionCount=3`、`emptyCount=0`、页面无 API error 文案。
  - 工程 gate：`corepack pnpm --filter @easyinterview/frontend build`、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`、`make docs-check`、`git diff --check` 通过。

## 2 会话中的主要阻点/痛点

- Home 和 Parse 各自复制了 `parseStatus === "ready"` 过滤规则。
  - **证据**：代码阅读确认 `HomeScreen.tsx` 与 `ParseScreen.tsx` 都有本地 `isSelectableResume`，且逻辑一致。
  - **影响**：修复 Home 时若不抽共享 helper，Parse 确认页仍可能在同一状态组合下显示空态。
- 既有测试只覆盖 ready 和 empty fixture，没有覆盖“非归档、有可读正文、parseStatus 非 ready”的组合。
  - **证据**：新增 Red test 构造 `failed` / `queued` / `processing` 三类可读简历后，旧实现仍禁用 select。
  - **影响**：真实或本地 AI 解析失败但已抽取正文时，用户看到简历列表有资产，Home 却无法选择。

## 3 根因归类

- 根因：Home/Parse 选择器把“可用于面试绑定”的判断简化为 `parseStatus === "ready"`，没有复用 Resume 详情中“已有可读正文即可展示”的当前合同。
  - **类别**：spec-plan。
- 根因：`frontend-resume-workshop/002-create-flow` 原计划只覆盖 Home create CTA handoff，没有把 Home existing-resume selector 作为同一 JD quick-start 入口的可执行 gate。
  - **类别**：spec-plan。

## 4 对流程资产的改进建议

- Home / Parse / Workspace 这类连续面试入口凡消费简历列表，都应复用 `isSelectableInterviewResume` 或等价共享 contract，不再复制状态枚举过滤。
  - **落点**：spec-plan
  - **优先级**：high
- 后续针对 Resume parse 状态的 UI 改动，应至少覆盖 ready、empty、failed-with-readable-body、queued-with-readable-body、queued-without-readable-body 五类 fixture。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高价值后续动作：在下一轮 workspace 简历 picker 或 practice handoff 调整前，先搜索是否还有本地 `parseStatus === "ready"` 简历过滤，统一改成共享 helper 或明确写出为什么该入口必须只接受 ready。
- 可延后动作：如果 backend 后续明确要求 practice plan 只能绑定 ready 简历，再把该约束提升到 OpenAPI / backend validation，并同步 Home/Parse 文案说明；不要让前端单独用隐式过滤表达这个限制。
