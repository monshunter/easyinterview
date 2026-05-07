# Frontend Shell UI-Design Pixel Parity Gate 交付复盘报告

> **日期**: 2026-05-08
> **审查人**: Claude

## 1 复盘范围与成功证据

- 交付主题：`frontend-shell/003-ui-design-pixel-parity-gate`，关闭 D2 retrospective（[2026-05-07](./2026-05-07-frontend-shell-app-shell-visual-system-assessment.md)）列出的最高优先级 follow-up — 把 D2 视觉系统的 desktop / mobile viewport bounding-box + screenshot 维度从 vitest+jsdom 范围抬到 Playwright + chromium 真实浏览器 gate。
- 提交链（在 `codex/frontend-shell-ui-design-native-parity` 分支）：
  - `7eecdc6` `docs(frontend-shell): derive D2 follow-up pixel parity gate plan` — spec v1.8 + plan v1.0 + checklist + bdd + context.yaml 派生。
  - `<feat>` `feat(frontend-shell): land ui-design pixel parity gate (Playwright)` — Playwright 基础设施 + 4 个 spec + 服务器 fixture + auth.css mobile fold + handoff README + 场景包。
  - `9b2897a` `docs(frontend-shell): close 003-ui-design-pixel-parity-gate plan lifecycle` — Header 状态切到 completed，INDEX 同步。
- 验证证据：
  - 单元测试：`pnpm --filter @easyinterview/frontend test` 40 files / 237 tests PASS；`pnpm exec tsc -p . --noEmit` 0 错。
  - 真实浏览器：`pnpm --filter @easyinterview/frontend test:pixel-parity` 42 项 PASS（4 spec × 2 chromium project，desktop 1440×900 + mobile 390×844）。
  - 场景：`E2E.P0.001 / 002 / 004 / 005 / 006` 五个 scenario `setup→trigger→verify→cleanup` 全 PASS；P0.006 trigger.log 含 `42 passed` + 0 failed + 4 spec markers + 双 project markers。
  - 构建：`pnpm --filter @easyinterview/frontend build` 与根 `make build` PASS。
  - 文档：`make docs-check` zero drift；`check_md_links` 双 OK；`docs/spec/frontend-shell/plans/INDEX.md` 把 003 行迁到「已完成」。
  - 负向：`grep -rE 'cypress|puppeteer|@emotion/|styled-components|tailwindcss|postcss-tailwind' frontend/` 命中只有 globalCss / fonts 测试的负向断言；`copernicus|styreneb` 命中只有 fonts.test.ts；retired-module testid 命中只在负向断言文件出现。
- 触达 spec.md C-9 的可执行验证维度：DOM 锚点（5 入口 testid + ei-* className）、computed style（height 58 / padding 32 / border 1px solid rgb(231,226,214) / font-family JetBrains Mono / display fontSize 48px / card padding 28px）、bounding box（TopBar fits viewport / 五入口不重叠 / display controls + user area 无重叠 / auth shell desktop 双列 + mobile 单列堆叠）、screenshot regression（home warm/light baseline 维护，dark / customAccent 通过 token + body bg + inline style 守住可见性）。

## 2 会话中的主要阻点/痛点

- 真实浏览器跨 ui-design golden preview 的像素 diff 在工程上不可达。
  - **证据**：`ui-design/index.html` 通过 unpkg.com 加载 React + Babel + Google Fonts；`frontend/dist` 通过 fontsource bundle 内嵌字体。同一字体名（Noto Serif SC / Inter / JetBrains Mono）在两套字体源下子像素渲染会偏移；强行做 RGB diff 会触发不可预测的 false positive。
  - **影响**：plan §4.1 原本写「frontend / ui-design 两份截图做对比断言两边 RGB 差异 ≤ 阈值」，实施时改为「frontend home baseline regression + 跨 ui-design DOM/computed style 对照 + dark/customAccent 状态可见性断言」，并在 plan §4 / scenario README §6 / 4.1 verified 注释里显式记录 deferred 原因。
- jsdom mobile 项目默认 webkit 引擎，与本仓库默认安装的 chromium 二进制不匹配。
  - **证据**：首次跑 `pnpm exec playwright test` 时 mobile project 因为 `devices["iPhone 14"]` 走 webkit 引擎，`browserType.launch: Executable doesn't exist at .../webkit_mac14_arm64_special-2251/pw_run.sh`。
  - **影响**：浪费一次 Red→Green 循环。修复方式是把 mobile project 也设为 chromium engine（仅改 viewport / deviceScaleFactor / hasTouch），这样安装步骤只需 `playwright install chromium`，不需要再多 100MB+ 的 webkit。
- AuthShell 的 `0.88fr 1.12fr` 固定 grid 在 mobile viewport (390px) 上把卡片推到视口右侧负坐标。
  - **证据**：`layout.spec.ts` mobile 项目断言 `card.right ≤ shell.right + 1` 在 D2 实现下失败（`shell.right=-331, card.right=-288`），意味着 AuthShell 在 mobile 下整体被推出视口。
  - **影响**：发现 D2 视觉系统的真实 mobile 体验有可见缺陷；当场补上 `@media (max-width: 768px)` 把 grid 折叠为单列，然后才让 layout.spec.ts mobile 通过。这是 jsdom 范围内不可能发现的问题，正是 Playwright gate 的价值所在。
- Playwright `toBeVisible()` 在 mobile viewport 下因为元素超出可视区域被判为 hidden。
  - **证据**：`screens.spec.ts` 用 `toBeVisible()` 检查 auth-login-email-form / password-stub / oauth-stub 在 mobile 下失败，因为 mobile 折叠后这些元素需要滚动才可见。
  - **影响**：把 `toBeVisible()` 改为 `toHaveCount(1)`（DOM 存在即可），保留 layout.spec.ts 中的 viewport / 滚动断言。

## 3 根因归类

- 跨 ui-design 像素 diff 的局限属于工程现实约束，不是流程缺陷。
  - **类别**：spec-plan（已就地修订 plan §4 + checklist 4.1 verify 注释 + scenario README §6）。无须仓库改动。
- mobile project webkit 默认引擎的浪费是 Playwright 默认配置 vs 本项目「单浏览器最小化安装」诉求之间的隐性张力。
  - **类别**：spec-plan（plan §4 Phase 1.2 与 checklist 1.2 已固化「mobile 也走 chromium engine」），可作为 frontend 视觉测试 plan 的默认模板。
- AuthShell 在 mobile 下溢出视口暴露了 D2 视觉系统对响应式断点的覆盖缺口。
  - **类别**：spec-plan（D2 plan 没有强制 mobile breakpoint；本次在 003 实施过程中临时补 `@media (max-width: 768px)`）。建议把「视觉系统 plan 必须在每个 shell-ish 卡片上声明 mobile breakpoint 行为」作为 `/design` Step 3.5 视觉类 row 的子项。
- Playwright `toBeVisible()` vs `toHaveCount(1)` 的差异是测试 API 的正常学习成本，不是流程缺陷。
  - **类别**：无需仓库改动（已在 spec 注释里说明选用 `toHaveCount(1)` 的理由）。

## 4 对流程资产的改进建议

- 把「视觉系统类 plan 必须在每个 shell 卡片上声明 mobile breakpoint」作为 `/design` Step 3.5 Coverage Matrix 的强制 row。
  - **落点**：`.agent-skills/design/SKILL.md` Step 3.5
  - **优先级**：medium
- 在 `/scenario-create` skill 的「视觉系统 / 浏览器渲染相关场景」分支里新增一条：mobile project 默认走 chromium engine + 自定义 viewport / hasTouch，避免 webkit 二进制冗余安装。
  - **落点**：`.agent-skills/scenario-create/SKILL.md`（D2 retrospective 已建议过新增视觉系统分支，本次提供具体配置 hint）
  - **优先级**：low
- 把 retrospective 中的「跨字体源像素 diff 不可达」结论写进 `frontend/README.md` §2.7 与 plan/scenario README，作为后续不再重复发现的 anchor（已在本次 commit 内固化）。
  - **落点**：spec-plan（已闭环）
  - **优先级**：done
- 在 D1 plan / spec 中加 mobile breakpoint 责任声明（`auth.css` / `screens.css` 必须显式 cover ≤768px）。
  - **落点**：spec-plan（D1 spec.md C-? 或 D2 spec.md C-8 修订）
  - **优先级**：low（本次 003 已就地修复，可待后续 frontend-shell spec 同步收口）

## 5 建议优先级与后续动作

- 下一轮最值得实施：
  - **决定本分支与 dev 上 D2 v1（8 commit）+ 003 follow-up 的整合策略**：用户在 PR review 阶段做 force-push / merge-commit / 替换 v1，把 codex 分支的 D2 v2 + D3 全部带到 dev；不要让两条分支长期并行。
  - **D2-D6 业务屏幕实现**：003 已经把视觉骨架的 jsdom + Playwright 双层 gate 都接入，D2-D6 owner 可以直接基于 `ei-screen-shell` / `ei-screen-card` 在已有 parity gate 守护下扩展业务内容。建议下一步派生 `frontend-home-job-picks-and-parse`（D2）。
- 可以延后处理：
  - `/design` Step 3.5 mobile breakpoint row（medium）：本次只在 AuthShell 暴露，体量小；后续如有更多视觉接入 plan 出现同类问题再固化。
  - `/scenario-create` mobile chromium engine hint（low）：靠 plan 模板传播即可，影响范围小。
  - 离线 vendor `ui-design/` CDN 资源（low）：当前 CI / 本地有外网；离线诉求出现再做。
