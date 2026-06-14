# Resume Detail Review Regressions 交付复盘报告

> **日期**: 2026-06-14
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付范围：修复 reviewer 指出的 `ResumeDetailView` 三条回归，包括 omitted `structuredProfile` 崩溃、accepted rewrites 未写入 flat profile bullets、tailor rerun 丢失 route `targetJobId`。
- 代码范围：`frontend/src/app/screens/resume-workshop/ResumeWorkshopScreen.tsx`、`frontend/src/app/screens/resume-workshop/components/ResumeDetailView.tsx`、`frontend/src/app/screens/resume-workshop/components/ResumeDetailView.test.tsx`。
- 成功证据：Red 阶段 focused test 先失败，暴露 3 个新增 regression；修复后 `ResumeDetailView.test.tsx` 12 tests PASS。
- 扩展验证：`ResumeWorkshopScreen.test.tsx` + `useRequestResumeTailor.test.tsx` 16 tests PASS；`pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop` 27 files / 161 tests PASS；`typecheck`、`build`、`git diff --check` PASS。
- 关联 Bug：[`BUG-0123`](../bugs/BUG-0123.md)。

## 2 会话中的主要阻点/痛点

- Reviewer 发现的问题集中在 D-20 flat resume reshape 后的 runtime data shape，但现有 regression test 仍偏向旧 `sections[]` happy path。
  - **证据**：原有 `ResumeDetailView.test.tsx` 只断言 `structuredProfile.sections[].bullets` overwrite；新增 flat `experience` / `experiences` / `projects` test 在 Red 阶段失败。
  - **影响**：用户可见成功 toast 与真实落盘内容不一致，属于高风险数据保存回归。
- Route context 在 root screen / auth gate 层已保留，但 detail rerun 子容器没有 body-level assertion。
  - **证据**：`parseResumeWorkshopParams` 和 `ResumeWorkshopScreen` root data attribute 已有 `targetJobId`，但 `requestResumeTailor` body Red 阶段缺少该字段。
  - **影响**：tailor rerun 从 JD-aware 退化为 generic suggestions，问题只靠路由/鉴权测试无法发现。
- Flat API omitted field 行为没有进入 UI fallback 测试。
  - **证据**：新增 omitted `structuredProfile` 场景 Red 阶段触发 `TypeError`。
  - **影响**：queued / failed / empty parse 这类真实响应可使 detail 页面崩溃。

## 3 根因归类

- D-20 plan Phase 8 迁移中，accepted rewrite 保存 helper 没有同步从 legacy `sections` 迁移到 flat profile shape。
  - **类别**：spec-plan
- Context-preservation 测试粒度停在 route/auth 层，没有覆盖 rerun request body。
  - **类别**：spec-plan
- omitted optional object 的 fallback 没有被当前 fixture matrix 覆盖。
  - **类别**：spec-plan

## 4 对流程资产的改进建议

- 在 `frontend-resume-workshop/003` Phase 8 checklist 中补充 flat profile save regression 明细：overwrite / duplicate 必须覆盖 `experience`、`experiences`、`projects`，不能只看 `sections`。
  - **落点**：spec-plan
  - **优先级**：high
- 对 `requestResumeTailor` 的 D-20 rerun gate 增加 route `targetJobId` body assertion，并明确无 `targetJobId` 时才允许 generic rerun。
  - **落点**：spec-plan
  - **优先级**：high
- 后续如修订 fixture matrix，增加 omitted `structuredProfile` flat resume scenario，避免 UI fallback 只在 handcrafted component test 中覆盖。
  - **落点**：spec-plan
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：把本次 regression tests 对应的不变量写回 `frontend-resume-workshop/003` Phase 8 gate，作为 D-20 收口完成条件。
- 次优先级：在 OpenAPI fixture 或 mock transport 场景中补一个 omitted `structuredProfile` detail response，扩大 fixture-backed UI gate 覆盖。
