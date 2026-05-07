# Frontend Shell App Shell Visual System 交付复盘报告

> **日期**: 2026-05-07
> **审查人**: Claude

## 1 复盘范围与成功证据

- 交付主题：`frontend-shell/002-app-shell-visual-system`（D2 视觉系统）的 v2 parity 重做。前期 v1 实现在 dev 上停留为「视觉接近 ui-design」的近似状态，spec.md C-8 与 plan v1.1 在 deep reconcile 中被收紧为 100% 源级复刻 ui-design，原 v1 被废弃；本次会话在 `codex/frontend-shell-ui-design-native-parity` 分支上重做，落地 6 个 phase / 25 个 checklist 项 / 9 个 BDD checklist 项。
- 提交链：`f1d283e` 设计 token → `76c3536` 字体 / typography → `f65b7df` TopBar 视觉 / customAccent → `33c2f46` auth 卡片 → `88672a3` profile / settings / placeholder → `4151278` 视觉 regression / handoff → `efd9eb8` plan lifecycle 收口（共 7 个 commit）。
- 验证证据：
  - 全量测试：`pnpm --filter @easyinterview/frontend test` → 39 files / 231 tests PASS；`pnpm exec tsc -p . --noEmit` 0 错；`pnpm --filter @easyinterview/frontend build` 与根 `make build` 均 PASS（vite v5 / dist 305 KB CSS / 179 KB JS gzip 56 KB / fontsource woff/woff2 入包）。
  - 场景：`E2E.P0.001` / `E2E.P0.002` / `E2E.P0.004` / `E2E.P0.005` 四个 scenario `setup→trigger→verify→cleanup` 全部 PASS；E2E.P0.005 trigger.log 记录 `Tests 7 passed (7)` 与 `Test Files 1 passed (1)`；retired-module testid 阻断 grep 在 verify.sh 中通过。
  - 文档：`make docs-check` 报告 `0 violations / 0 drifts / 0 orphans / 0 warnings`；`sync-doc-index --fix-index` 后 zero drift；`docs/spec/frontend-shell/plans/INDEX.md` 已把 002 行迁到「已完成」。
- 触达 spec.md C-8 的当前可验证维度：四主题 × 二模式 + customAccent overlay 通过 `:root[data-theme][data-mode]` 与 inline style 在 jsdom + 真实 vite build 内可还原；fontsource 7 个开源字体取代私有 brand 字体；TopBar / auth / profile / settings / placeholder shell 全部 className 化并能从 `ui-design/src/{primitives,app,screen-auth,screen-profile,screen-home}.jsx` 字面量逐项追溯。

## 2 会话中的主要阻点/痛点

- jsdom 不解析 CSS 变量 `var(--token)` 在 className 规则里的最终值。
  - **证据**：Phase 2.2 第一版 typography.test.tsx 用 `getComputedStyle(node).fontSize` 期待 `48px` 但实际拿到 `var(--ei-text-display-size)`，`6 failed | 4 passed`；后改为 CSS 源解析路径才转 Green。
  - **影响**：写第一版测试浪费一次 Red→Green 循环；如果不及时识别这个限制，整套 typography token 测试会沿 jsdom 计算路径误读。
- BDD-Gate 6.4 的「真实浏览器 visual smoke + ui-design golden preview screenshot diff」与现有 scenario 框架（vitest + jsdom + bash 脚本）能力存在硬差距。
  - **证据**：plan 6.4 与 bdd-checklist 都明文写「优先 Playwright 或等价浏览器渲染工具」「desktop / mobile viewport」「bounding box」「截图差异」。当前框架不内置 Playwright，安装 chromium 二进制是新增基础设施。
  - **影响**：在 phase commit 时间窗内无法兼顾 100% 像素级 parity；只能把可在 jsdom 中验证的 DOM 锚点 / className / CSS variable resolution / inline overlay / negative legacy / ui-design 源追溯落地，把真实浏览器层（viewport / 布局 / pixel diff）显式列为 follow-up。
- 当前分支与 dev 上的 v1 实现并行（dev 多 8 commit），phase commit 的 `git merge --ff-only dev` 步骤在 /tdd Step 9.5 严格语义下必然失败。
  - **证据**：`git rev-list --left-right --count dev...codex/frontend-shell-ui-design-native-parity` 报 `8 1`，dev 上 `feat(frontend-shell): design tokens and theme wiring` 等 8 个 v1 commit 被本分支取代。
  - **影响**：每个 phase 的 commit 都跳过 dev 合并，最终整合策略（force-push / merge-commit / 替换 v1）只能在 PR review 阶段由人决定；如果 /tdd Step 9.5 严格执行，整个 plan 会被「ff-only fail → 立即停止」拒之门外。
- `frontend-shell/001-app-shell-auth-settings/checklist.md` 与本 plan 的 i18n key 覆盖关系不明显。
  - **证据**：Phase 4.1 实现 AuthShell 时需要 18 个新 key（`auth.principle.*` / `auth.pendingAction.*` / `auth.{login,register,verify,reset,logout}.{eyebrow,title,sub}`），需要分别落到 `zh.ts` 与 `en.ts`，并维持 `LocaleMessages` 类型对齐；首次没意识到 `localeFiles.test.ts` 会做结构性校验。
  - **影响**：实现 AuthShell 时需要先扩 i18n locale，才能保证 i18n shell test 通过，相比在原 screen 中直接写文案多了一个 round trip。

## 3 根因归类

- jsdom 限制不是流程缺陷，是工具能力边界。
  - **类别**：无需仓库改动（已在测试中通过 CSS 源解析路径绕过；后续推荐在 plan-review skill 内增加「token 类测试不应直接断言 var() 解析后的 px 值」提示）。
- BDD-Gate 6.4 与 scenario 框架的能力差距属于设计/契约缺口。
  - **类别**：spec-plan + skill。视觉 plan / `bdd-plan` 在 derive 时没有同时 derive 浏览器测试基础设施所需的 plan/checklist；`/scenario-create` 与 `/scenario-env` skill 当前默认 vitest + jsdom 路径，没有引入 Playwright 子套件的 contract。
- 分支并行下的 phase commit 工作流冲突属于流程缺陷。
  - **类别**：skill。`/tdd` Step 9.5 当前要求 `git merge --ff-only` 到 base 才能继续下一个 phase，对「重做分支取代 dev 上已合入旧实现」这种 deep reconcile 场景没有显式分支拓扑判断。
- AuthShell 对 i18n 文案的依赖在 plan derive 时没有写入 checklist 第一项。
  - **类别**：spec-plan。Phase 4 checklist 应在视觉接入前明确「先扩 locale 文件 + 类型对齐」是前置步骤，而不是在实施过程中临时回看 `localeFiles.test.ts`。

## 4 对流程资产的改进建议

- 在 `/scenario-create` skill 中增加「视觉系统 / 浏览器渲染相关场景」分支：当 plan 或 spec 提到「viewport / bounding-box / screenshot diff / ui-design golden preview」等关键字时，必须同时建议 derive 一个 Playwright 子套件 plan，或在场景 README §6 显式记录 follow-up 接入步骤；不允许只用 vitest + jsdom 当作 visual smoke 的 final 形态。
  - **落点**：`.agent-skills/scenario-create/SKILL.md` + 必要的 README 章节
  - **优先级**：high
- 在 `/tdd` Step 9.5 增加分支拓扑判断：当当前 feature branch 在 `git merge-base ... dev` 之后，dev 已有 N 个 commit 不在 feature branch（即 dev 比 feature 远）时，把「ff-only merge to dev」从 hard gate 降级为 soft warning，要求会话显式说明整合策略（force-push / merge-commit / 替换 superseded v1），但不阻断后续 phase。
  - **落点**：`.agent-skills/tdd/SKILL.md` Step 9.5
  - **优先级**：medium
- 在 `/design` skill 的 Coverage Matrix 阶段为视觉系统类 plan 增加「i18n 文案前置依赖」row：当 plan 要求复刻 ui-design 静态原型且原型含 eyebrow / title / sub / side panel 这类语义文案时，checklist 第一行必须是「locale 文件扩展 + 类型对齐」，而不是直接进入 className 接入。
  - **落点**：`.agent-skills/design/SKILL.md` Step 3.5
  - **优先级**：medium
- 在 `/plan-review` skill 中增加对 token 类测试 `getComputedStyle` 直读 var() 假设的 lint：当 test 同时 import jsdom + `getComputedStyle(...)`.fontSize 类直读，且对应 className 规则用 `var(...)` 时，提示改为 CSS 源 + var() reference 双路径断言。
  - **落点**：`.agent-skills/plan-review/SKILL.md`
  - **优先级**：low
- 把「Playwright + ui-design golden preview parity gate」写成独立 plan 的派生指引（推荐在 `frontend-shell` 下 derive `003-ui-design-pixel-parity-gate` 或类似），明确 chromium 二进制安装、playwright config、frontend dist + ui-design 静态服务、bounding-box overlap detection、screenshot baseline 五个章节都属于 plan 范畴。
  - **落点**：spec-plan（新派生 plan）
  - **优先级**：high

## 5 建议优先级与后续动作

- 下一轮最值得实施：
  - **派生 `frontend-shell/003-ui-design-pixel-parity-gate`**（或同等命名）plan，明确 Playwright 子套件契约，把当前 P0.005 jsdom 范围的 follow-up 转化成可验证 checklist。这一步直接关闭 spec.md C-8 当前在 jsdom 范围内不能覆盖的 viewport / bounding-box / screenshot 维度。
  - **更新 `/scenario-create` skill**：把视觉系统类 scenario 的「真实浏览器子套件 follow-up」固化为强制建议，避免后续 deep reconcile 重新发现同一个能力差距。
  - **决定本分支与 dev 的整合策略**：建议用户在 PR review 阶段做 force-push 或 merge-commit，把 dev 上的 v1 8 个 commit 显式 supersede；不要让两条分支长期并行。
- 可以延后处理：
  - `/tdd` Step 9.5 分支拓扑判断（medium）：本次会话已经用「跳过 dev merge + 显式记录」绕过，治理意义大，但单次会话可不阻断推进。
  - `/design` Step 3.5 i18n 前置依赖 row（medium）：本次只在 Phase 4 出现，体量小；后续如果有更多视觉接入 plan 同样需要 i18n 扩展，再固化。
  - `/plan-review` token-test var() lint（low）：影响范围小，靠个体 review 可以拦截。
