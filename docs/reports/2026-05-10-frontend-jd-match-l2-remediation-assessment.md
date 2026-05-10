# Frontend JD Match L2 Remediation 交付复盘报告

> **日期**: 2026-05-10
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`/plan-code-review 002-jd-match-recommendations --fix` 的 L2 remediation，覆盖 `getJobRecommendation` 详情必调、`jd_match_action` 登录后 auto-resume、Search opaque pending payload、pixel parity gate、P0.027/P0.028/P0.030/P0.031 scenario verify 强化，以及 plan 002 文档生命周期恢复 completed。
- 关联计划：[002 JD Match Recommendations](../spec/frontend-home-job-picks-and-parse/plans/002-jd-match-recommendations/plan.md)。
- 关联 Bug：[BUG-0037](../bugs/BUG-0037.md)。
- 成功证据：`pnpm --filter @easyinterview/frontend test` 105 files / 663 tests PASS；`typecheck` PASS；`build` PASS；`make build` PASS；`make validate-fixtures` OK 46 fixtures；`pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/jd_match.spec.ts` 20 passed；P0.027/P0.028/P0.029/P0.030/P0.031/P0.017 scenario setup→trigger→verify→cleanup 均 PASS；`sync-doc-index --check` 与 `make docs-check` zero drift。

## 2 会话中的主要阻点/痛点

- L2 发现的核心语义没有被旧 gate 固化。
  - **证据**：原实现能通过历史 checklist，但缺少 `getJobRecommendation(jobMatchId)` detail fetch 红测和 auth auto-resume 一次性执行断言。
  - **影响**：需要新增 hook、状态恢复模块、组件接线、测试和 scenario verify，修复范围横跨代码、E2E 脚本和 plan 文档。
- BDD checklist 与实际已完成场景证据不同步。
  - **证据**：主 checklist 已显示 Phase 6 完成，但 `bdd-checklist.md` 多数项仍为 `[ ]`，需要重新补跑 P0.029/P0.017 并回写证据。
  - **影响**：L2 reviewer 必须额外判断历史 PASS 是否可复用，增加误判风险。
- Pixel parity 旧 gate 没有覆盖具体响应式布局不变量。
  - **证据**：补红测后发现 mobile Recommended grid 仍为 2 列，Market signals 缺可测 inner anchor，focused screenshot baseline 缺失。
  - **影响**：需要补 CSS、DOM anchor、route fixture mock、computed style 和截图断言。
- 当前 plan 文档仍残留旧“单一加载文案”口径。
  - **证据**：最终负向搜索命中 plan/bdd/checklist 标题和 i18n 说明中的旧文案；随后修正为 5 步 AGENT panel。
  - **影响**：如果不清理，后续执行者会在同一 plan 内看到互相矛盾的 Search loading 口径。

## 3 根因归类

- 语义 gate 不够强。
  - **类别**：spec-plan
  - 详情 fetch、pending action resume、opaque Search payload 这类跨层语义没有作为 plan 002 的必要验收 gate。
- L2 review 对 BDD checklist 的完成状态缺少强制反查。
  - **类别**：skill
  - 当前流程能发现问题，但没有自动提醒“主 checklist PASS 与 BDD checklist 未勾选”属于必须修复的文档漂移。
- Pixel parity gate 粒度不足。
  - **类别**：spec-plan
  - 早期 gate 偏 DOM 存在性，未把 grid column count、sticky/flow、theme computed style 和 focused screenshot 当成完成条件。
- 旧设计口径没有形成零残留检查项。
  - **类别**：spec-plan
  - plan 修订后没有同步加“单一加载文案”等旧短语负向搜索，导致正文残留。

## 4 对流程资产的改进建议

- 在 `plan-code-review` skill 增加 L2 closeout 检查：主 checklist 完成时，关联 BDD checklist 不得仍有 `[ ]`，除非每项有明确 skip reason。
  - **落点**：skill
  - **优先级**：high
- 在涉及 list/detail UI 的 plan gate 中固定三类断言：list operation 调用、detail operation 调用、detail-only fixture 字段渲染。
  - **落点**：spec-plan
  - **优先级**：high
- 在 auth pending action gate 中固定四类断言：pending params 最小化、登录后一次性恢复、重复触发防护、隐私字段负向搜索。
  - **落点**：spec-plan
  - **优先级**：high
- 将 BUG-0037 的检查模式候选沉淀到 `docs/bugs/PATTERNS.md`：L2 review 不能只看 route/DOM PASS，必须验证 operation-consumer 语义和 pending action resume 语义。
  - **落点**：README / bug pattern library
  - **优先级**：medium

## 5 建议优先级与后续动作

- 优先做：把 BUG-0037 的模式写入 `docs/bugs/PATTERNS.md`，并在下一次用户确认后更新，避免同类 L2 drift 再次只靠人工记忆发现。
- 其次做：在后续 plan-review / plan-code-review 中，把“旧设计口径负向搜索词”列入 owner plan gate，而不是只在执行末尾临时搜索。
- 可延后：清理现有测试中的 React `act(...)` warning；本轮所有断言通过，但这些 warning 会降低长输出的审查信噪比。
