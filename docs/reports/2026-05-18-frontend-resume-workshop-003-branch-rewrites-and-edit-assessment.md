# Frontend Resume Workshop 003 Branch Rewrites And Edit 交付复盘报告

> **日期**: 2026-05-18
> **审查人**: Claude

## 1 复盘范围与成功证据

- 交付范围：`docs/spec/frontend-resume-workshop/plans/003-branch-rewrites-and-edit` 全 7 phase（Phase 0 上游 preflight → Phase 7 BDD + parity 收口）；落地正式前端 `ResumeBranchFlow / ResumeRewritesTab / ResumeEditTab` 三屏 + 配套 hook + adapter，替换 plan 001/002 阶段的 `<NotImplementedPlaceholder>` 与 `<ComingSoonTab>` 占位。
- 通过证据：
  - 9 个 phase 提交按顺序落地（141e7abd / 176b5e33 / dcd3d6b6 / 7aa3e7b7 / 31bf8528 / 35f44d65 / 29c4aa50 / fbe80e09 / 0611d018），均通过 `LC_ALL=C perl` ASCII-only 校验。
  - `pnpm typecheck` 全程 PASS；`pnpm exec vitest run src/app/screens/resume-workshop/` 最终 39 文件 / 253 tests PASS（含新增 11+11+12+8+7+8+8+8 = 73 cases）。
  - 四个 BDD 场景 `setup → trigger → verify → cleanup` 全端到端 PASS：
    - `E2E.P0.084` 4 文件 / 42 tests
    - `E2E.P0.085` 3 文件 / 27 tests
    - `E2E.P0.086` 4 文件 / 34 tests
    - `E2E.P0.087` 5 文件 / 40 tests
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 报告 zero drift；`--fix-index` 把 plans/INDEX.md 行从 active 迁到 completed。
  - 三类 grep gate 在 `frontend/src/app/screens/resume-workshop/branch/` + `tabs/` 全 0 命中（retired modules / retired tailor mode / prototype runtime import）。
  - plan / checklist / bdd-plan / bdd-checklist Header 状态 active → completed，更新日期 2026-05-18。

## 2 会话中的主要阻点/痛点

- **Plan 003 §0.5 baseline grep 与 §7.10-7.12 closure scope 不一致**
  - **证据**：Phase 0.5 baseline 要求整个 `frontend/src/app/screens/resume-workshop/` retired tailor mode `(inline|rewrite|mirror)` 0 命中，但 plan 002 已在 `create/` 引入 `resume-preview-confirm-inline-error` testid 与 `inline validation error` 测试描述（共 6 行命中），与 §7.10-7.12 仅检 `branch/` + `tabs/` 的 scope 不一致。
  - **影响**：Phase 0 时增加判断成本，需在 evidence 注释里显式记录 false positive 性质，并把实际 enforcement 锚定到 §7.10-7.12 scope；如果以后另一个 plan 再用同一模式 baseline 整库 grep，类似 false positive 还会反复出现。

- **i18n 文案使用 `inline` / `by mistake` 与 retired-grep regex 产生新的 false positive**
  - **证据**：Phase 1-6 实施过程中我在多个测试描述、组件注释和 mapper docstring 中使用了英语 "inline alert" / "by mistake" 等自然语言，到 Phase 7.10-7.11 收口时被 grep 命中（4 处需改写）。
  - **影响**：实施阶段为了写出可读的英文测试名字与注释，第二次踩到与 baseline 同源的语义陷阱；最终改写为 `in-form alert` / `by accident` 并不损失语义，但说明 plan 文档若能在 §3 / §0.5 显式列出 "禁止在新代码 / 测试描述中使用 inline / mistake / rewrite / mirror 字眼"，可以省去这一轮往返。

- **Vitest fake timer 下 `useResumeTailorRunPolling` 轮询时序假设易过冲**
  - **证据**：首次提交的 `useResumeTailorRunPolling.test.tsx` 用 `initialDelayMs=100 / backoff=1.4`，`advance(150)` + `advance(200)` 第二步直接跨过 generating 进入 ready；同一文件 unmount 用例 `advance(20)` 也无意中触发第二次 tick，spy 调用 2 次而非 1 次。
  - **影响**：两个测试在第一次跑时 fail，需要把 backoff 改成常数 1.0 并选 50ms 步长后才稳定。这是经典的 fake-timer 时序假设错误，提示 plan 文档可以把 "exponential backoff 测试用常数 backoff 解锁断言" 写成显式约定。

- **`AppRuntimeContext` 早期未 export 阻塞了 hook 单测**
  - **证据**：`useResumeBranchSubmit.test.tsx` / `useUpdateResumeVersion.test.tsx` / `useTailorSuggestionDecision.test.tsx` 等聚焦 hook 测试需要直接挂载 stub client 而不触发 `getRuntimeConfig / getMe` 真实请求。Phase 2 实施初始时 `AppRuntimeContext` 没有 export，hook 测试无法 mount。最终在 Phase 2 commit 内附带 export 该 context。
  - **影响**：增加一次小型 runtime/provider 重构；若 D1 frontend-shell README §2.3 把 "聚焦 hook 测试的运行时挂载方式" 显式列出，后续 owner 不需要遇到同类问题再决定 export。

- **生成器 ResumeWorkshopIcon API 与 ui-design 源不完全对齐（`color` prop、`arrow_left` vs `arrowLeft`）**
  - **证据**：Phase 1 把 ui-design `Icon name="arrow_left" color={T.ink3}` 直接迁移时被 `pnpm typecheck` 拒绝（icon 名是 camelCase、ResumeWorkshopIcon 没有 `color` prop）。需要改写为 `name="arrowLeft"` 并由父级容器写 `color` style 继承 stroke。
  - **影响**：实施阶段需要二次理解 ResumeWorkshopIcon 的工程口径与 ui-design `Icon` 的 prop 形态差异；这是 plan 001 时既定的迁移规则，但 plan 文档 / frontend README 没有显式记录“ResumeWorkshopIcon name 必须 camelCase / 不接受 color prop”。

## 3 根因归类

- **plan 文档的 grep gate scope 不一致**（plan-spec）
  - 类别：spec-plan
  - 性质：plan 003 §0.5 baseline grep 与 §7.10-7.12 closure grep scope 不同，未声明 baseline 与 closure scope 错位是“允许的设计差”还是“需要修订的内部不一致”。

- **plan 文档没有显式约束新代码 / 测试中的英语字眼**（plan-spec）
  - 类别：spec-plan
  - 性质：plan 003 §7.10-7.12 grep 把 `inline / rewrite / mirror / mistake / experiences / ...` 当作 retired 概念全文匹配，但实施中常用的英语短语包含这些词。

- **Vitest fake-timer + 指数退避测试时序假设**（无需仓库改动）
  - 类别：no repo change needed
  - 性质：一次性的测试设计调整，未涉及流程或 spec 缺陷。

- **AppRuntimeContext 默认未 export 影响 hook 测试 mount**（README）
  - 类别：README（`frontend/README.md` §2.3 runtime/mock transport 入口）
  - 性质：现在已 export 但说明缺失；后续 owner 仍需自行发现。

- **ResumeWorkshopIcon 迁移口径未显式记录**（README）
  - 类别：README（`frontend/README.md` §2.7 ui-design 原生迁移规则 / 或 plan 001 spec 中的 icon 迁移约束）
  - 性质：plan 001 已奠定该 helper，但 prop 形态与命名风格未显式写明，后续 owner 易踩同样问题。

## 4 对流程资产的改进建议

- **plan 003 retrospective 中显式记录“baseline scope vs closure scope 的差异约定”**
  - 落点：`docs/spec/frontend-resume-workshop/plans/003-branch-rewrites-and-edit/plan.md` 收口附录或同 spec 的 `history.md`。
  - 优先级：medium
  - 行动：补一段 "Phase 0 baseline 用整库 scope 是为了暴露 false positive，§7.10-7.12 closure 缩到 branch/ + tabs/ 是设计性差异"。

- **在 plan 类模板 / `/plan-review` skill 中加入 “禁止 retired 字面量出现在新代码 / 测试描述” 的预检 prompt**
  - 落点：`.claude/skills/plan-review` 或 plan 模板 §3 retired 字面量 baseline 段落，提示 owner 在动手前替换 `inline / mistake / ...` 为同义词。
  - 优先级：medium
  - 行动：在 `/plan-review` 的 retired-grep 章节附 "命中前请先在新代码 / 测试中使用 same-meaning English alternatives" 的提示。

- **`frontend/README.md` §2.3 显式说明聚焦 hook 测试的 runtime 挂载方式**
  - 落点：`frontend/README.md` §2.3 Runtime / Mock transport 入口。
  - 优先级：high（受益面：所有后续业务 hook owner）
  - 行动：补一段说明 “聚焦 hook 测试可通过 `AppRuntimeContext.Provider value={...}` 直接挂载 stub client，避免 `<AppRuntimeProvider>` 触发 `getRuntimeConfig / getMe` 实际请求”。

- **`frontend/README.md` §2.7 / plan 001 spec 列出 `ResumeWorkshopIcon` 命名与 prop 形态**
  - 落点：`frontend/README.md` §2.7 ui-design 原生迁移规则附录。
  - 优先级：medium
  - 行动：补一句 “Resume Workshop 内部 icon helper `ResumeWorkshopIcon` 使用 camelCase name 且 stroke 颜色继承父级 color，迁移 ui-design 的 `Icon name="arrow_left" color={...}` 时需双重改写”。

- **延后：Playwright pixel parity + axe-core a11y 套件接入**
  - 落点：plan 003 retrospective 提到的 follow-up plan（`frontend-resume-workshop/004-pixel-parity-and-axe-rollup` 或类似 plan id），由 owner 决定何时建立 chromium baseline。
  - 优先级：low（当前 Vitest DOM/style anchor 已经覆盖源级镜像；plan §3 D-5 明确允许 baseline 缺失时不阻塞 clean checkout）。
  - 行动：当 frontend-shell pixel-parity infra 升级到长期维护期时，再启 follow-up plan，把 `resume-workshop-branch-rewrites-edit.spec.ts` 与 axe-core 套件一次性接入。

## 5 建议优先级与后续动作

- **下一轮最值得实施**
  - 把 `frontend/README.md` §2.3 增补 “聚焦 hook 测试如何直接挂载 `AppRuntimeContext.Provider`”，因为 D2-D6 多个 owner（debrief、workspace、report）都在写 hook 测试，受益面最大。
  - 在 `/plan-review` skill 的 retired-grep 章节补充“命中前请先用同义词”的 prompt，避免下一个 plan 在 closure 时再次返工同样的 false positive。

- **可以延后**
  - Playwright pixel parity + axe-core 套件接入：等 baseline 维护机制定稿后再启 follow-up plan，本期 Vitest DOM/style anchor 已经承担 UI parity 真理。
  - plan 003 history 中显式记录“baseline vs closure scope 差异”：可在下一次 `plan-review --fix` 通过时一并补，独立维护成本不高。

- **不建议改动**
  - Vitest fake-timer 时序断言失败是一次性测试设计调整，已修正；不构成流程缺陷。

---

**关联资产**：
- 计划：[`docs/spec/frontend-resume-workshop/plans/003-branch-rewrites-and-edit/plan.md`](../spec/frontend-resume-workshop/plans/003-branch-rewrites-and-edit/plan.md)
- Checklist：[`checklist.md`](../spec/frontend-resume-workshop/plans/003-branch-rewrites-and-edit/checklist.md)
- BDD 资产：[`bdd-plan.md`](../spec/frontend-resume-workshop/plans/003-branch-rewrites-and-edit/bdd-plan.md) / [`bdd-checklist.md`](../spec/frontend-resume-workshop/plans/003-branch-rewrites-and-edit/bdd-checklist.md)
- 场景目录：`test/scenarios/e2e/p0-084-resume-branch-flow-three-seed-strategies` / `p0-085-...` / `p0-086-...` / `p0-087-...`
- 提交链：`141e7abd → 176b5e33 → dcd3d6b6 → 7aa3e7b7 → 31bf8528 → 35f44d65 → 29c4aa50 → fbe80e09 → 0611d018`
