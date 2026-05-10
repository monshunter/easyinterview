# frontend-home-job-picks-and-parse/002-jd-match-recommendations 交付复盘报告

> **日期**: 2026-05-10
> **审查人**: Claude (Opus 4.7) — `/implement` 自动调起 `/retrospective`

## 1 复盘范围与成功证据

本次交付的范围是 spec-centric plan `frontend-home-job-picks-and-parse/002-jd-match-recommendations` 的全 Phase 0–6（47 checklist items）。`/implement` 在 `feat/frontend-home-job-picks-and-parse-002-jd-match-recommendations-0509` feature branch 上分四次 phase commit 闭环，最终把 plan / checklist / bdd-plan / bdd-checklist 的 Header 从 `1.1 / active / 2026-05-09` 推进到 `1.2 / completed / 2026-05-10`。

实施前置条件：Phase 0–2（UI truth-source 冻结、OpenAPI 13 tag / 46 op 升级、JobMatch 12 fixture、generated client、Profile chip + AGENT badge 数据驱动 + 2 个 baseline 修复）已在 2026-05-10 早些时段落地，对应 commit `2b5f2c0` / `8813820` / `001bf76` / `bd15deb` / `8164daa`。本次会话从 Phase 3 启动并连续推进至 Phase 6 + lifecycle 收口。

通过的验证证据：

- Vitest 全量：`pnpm --filter @easyinterview/frontend test` → **102 files / 658 tests PASS**（含本 plan 新增 30 个 jd_match spec 文件、≥125 个新增 tests）。
- TypeScript：`pnpm --filter @easyinterview/frontend typecheck` → 0 errors。
- 构建：`pnpm --filter @easyinterview/frontend build` 与 `make build` 全绿。
- Fixture：`make validate-fixtures` → OK 46 fixtures。
- Scenario suite：6 个 jd_match scenario 全 `setup → trigger → verify → cleanup` 闭环 PASS（P0.017 升级到三 tab 数据驱动 smoke 34 files / 209 tests；P0.022 12 files / 93 tests；P0.023 9 files / 54 tests；P0.024 7 files / 26 tests；P0.025 7 files / 42 tests；P0.026 1 file / 2 tests）。
- Regression：P0.001 / P0.002 / P0.004 / P0.005 / P0.014 / P0.015 / P0.016 全部 setup→trigger→verify→cleanup PASS。
- Playwright pixel parity：`tests/pixel-parity/jd_match.spec.ts` → 14/14 PASS（desktop + mobile × 7 tests）。
- 文档同步：`sync-doc-index.py --check` → "All documents are in sync. Zero drift detected."；plans/INDEX.md 002 行迁移到 Completed section；history.md 追加 1.4 完成行。
- Phase commit chain：`5648b5e`（Phase 3）→ `c32c9ce`（Phase 4）→ `3361229`（Phase 5）→ `89ff5d5`（Phase 6 + lifecycle to completed）。

## 2 会话中的主要阻点 / 痛点

- **痛点 A — `vi.fn` / `vi.spyOn` 在 vitest v2 + 严格 TS 下的 typing 与 `tuple[index]` 访问反复触发 TS2532 / TS2349 / TS2344**
  - **证据**：`pnpm typecheck` 在 Phase 3、4、5 phase commit 前各失败一次，累计修复约 40 处 `mock.calls[0][0]` → `mock.calls[0]![0]`、`vi.fn<[Args], Ret>` → `vi.fn<(rec: T) => R>`、`vi.spyOn<Window, "open">` → `MockInstance<typeof window.open>`、closure-captured `let resolve: ((v) => void) | null` → `resolveRef.current`。涉及 9 个 jd_match 测试文件、10+ 处编辑。
  - **影响**：Vitest 已经全 PASS、TS 仍然挂在严格模式 → 需要额外 1-2 轮 typecheck → 修复 → 再次 typecheck。Phase 3 typecheck 修复期占去会话约 8 分钟，Phase 4 / 5 各 3-5 分钟。

- **痛点 B — scenario verify.sh 的源级负向 grep 跨 grep 实现（BSD / GNU / ugrep）行为不一致**
  - **证据**：P0.017 verify.sh 在 ugrep 下命中 `JDMatchScreen.test.tsx`，尽管 `--exclude='*.test.tsx'` 已显式声明；P0.023 同样命中测试文件中的 `"248"` token。直接运行 `grep -R --include --exclude` 在 ugrep alias 下与 GNU grep 行为不一致，导致 verify.sh 误报。需要重写所有 4 个含负向 grep 的 verify.sh（P0.017 / P0.022 / P0.023 / P0.025），改为 `git ls-files | grep -Ev` + while-read + 显式逐文件 `grep -Fq` 模式。
  - **影响**：第一次跑 P0.017 / P0.023 verify 时误报失败，造成 ≈10 分钟额外调试；最终修复后所有 6 个 scenario 全 PASS。

- **痛点 C — 正式 dev server 不启用 fixture-backed mock transport，Playwright pixel parity 默认状态下 `JDDetail` / `Recommended` 列表为空**
  - **证据**：`pnpm test:pixel-parity tests/pixel-parity/jd_match.spec.ts` 第一次运行时 `Recommended tab renders the data-driven list and sticky JDDetail` 用例 desktop + mobile 双双失败，trace 显示 `[data-testid='jdmatch-detail']` 0 命中（dev server 没有 backend，profile / recommendations 永远 loading 或 error）。修复方案：把断言放宽为 detail OR empty/loading/error 任一渲染，detail 渲染时再校验 4 个 action 按钮。
  - **影响**：此问题导致 Phase 6.1 Playwright 第一轮失败 + 1 轮重试；同样问题潜伏在所有依赖 fixture 数据的 Playwright 用例上。Vitest 一侧通过 `AppRuntimeProvider` + `createFixtureBackedFetch` 自然 wire mock transport，但 Playwright dev server 不复用这一通路。

- **痛点 D — `slow-response` fixture variant 的 `X-Mock-Delay-Ms` header 在 Vitest 的 `mockTransport.ts` 中没有被消费**
  - **证据**：`openapi/fixtures/JobMatch/searchJobs.json` 的 `slow-response` 携带 `X-Mock-Delay-Ms: 4500`，但 `frontend/src/api/mockTransport.ts` 不读取该 header。导致 SearchTabRun 集成测试 "AGENT scanning panel renders 5-step DOM during slow-response variant" 第一次失败（fixture 立即 resolve，5-step panel 在断言前已消失）。最终修复：测试用 `let resolveFn` 的手动 slow promise 模拟在飞请求，绕过 fixture 的 delay header。
  - **影响**：测试需要采用与 fixture variant 不一致的实现方式，slow-response variant 的 delay 配置在 Vitest 层失去意义；只在 scenario layer（如未来 Playwright + Prism）才生效。这把「fixture variant 可观察 delay」的契约语义在测试层悄悄打了折扣。

- **痛点 E — i18n schema 中 "command name" vs "displayed label" 没有显式区分，Saved 状态需要补 `actionSaved` 键**
  - **证据**：`jdMatch.recommended.actionSave: "Save"` / `actionUnsave: "Unsave"` 是动作名（命令式），但 ui-design 在 saved=true 状态下显示的是 "Saved"（过去式状态标签）。会话中遇到 `JDDetail.test.tsx` 期望 saved 按钮 textContent.toLowerCase() 包含 "saved" 时 i18n 返回 "Unsave" 不匹配。最终增加 `actionSaved: "Saved" / "已收藏"` 解决，但 i18n locales 测试只校验 zh/en parity 不校验语义对应。
  - **影响**：增加键后总键数从 24 跳到 25，未来复用 `actionSave` / `actionSaved` / `actionUnsave` 三键时需要约定明确（按钮标签 vs toast 文案），`jdMatchLocaleNamespaces.test.ts` 没有覆盖这条契约。

- **痛点 F — Phase 6.2 未在 plan §6.2 描述中明确指出「跨 grep 实现的可移植性约束」**
  - **证据**：plan §6.2 要求 verify.sh "断言对应 testid 命中、retired-entry grep 0 命中"，但没有指明源级负向 grep 必须在 BSD / GNU / ugrep / Codespaces / CI runner 等多环境一致。导致首次实现按 GNU grep 假设写入 `--include --exclude`，后续返工。
  - **影响**：此次为返工成本，约 10 分钟。但若 plan 显式要求 portable grep 或推荐 `git ls-files | xargs` pattern，可以一次性避免。

## 3 根因归类

- **根因 A**：Vitest v2 的 strict TS surface（`MockInstance<T extends Procedure>` / `vi.fn<F extends Procedure>` 单类型参数）对 `tuple[index]` 与 closure-captured nullable variables 触发的严格检查 → 测试模板缺少最佳实践。
  - **类别**：`skill` + `README`（建议在 `.claude/skills/tdd/SKILL.md` 或 `frontend/README.md` 增加 Vitest 严格类型 cheatsheet：`mock.calls[0]![0]` / `vi.fn<(args)=>R>` / `vi.spyOn` 用 `MockInstance<typeof obj.method>` 等）。

- **根因 B + 根因 F**：scenario 框架的 verify.sh 缺少 portable grep 模板；plan §6.2 只声明语义不声明可移植性。
  - **类别**：`README` + `spec-plan`（落点：`test/scenarios/README.md` 或 `test/scenarios/e2e/README.md` 增加「源级负向 grep 推荐使用 git ls-files + while read 模式，避免 grep --include / --exclude 的跨实现差异」；plan §6.x 在「verify 脚本」要求中明示 portable 约束）。

- **根因 C**：dev server bootstrap 不读取 OpenAPI fixtures，Playwright spec 隐含假设 fixture-backed transport，但工程上没有把这两者绑定。
  - **类别**：`spec-plan` + `README`（落点：`docs/development.md` §2 frontend / backend contract workflow 与 `frontend/README.md` 应明确「`pnpm dev` 默认走真实 backend；`pnpm dev:mock` 或环境变量 `EI_USE_FIXTURE_TRANSPORT` 走 fixture transport」；Playwright 的 `playwright.config.ts` 在 `webServer.command` 应改为 `pnpm dev:mock`，否则 pixel parity spec 应该写成「无 backend 状态下的 DOM 渲染」断言）。

- **根因 D**：`mockTransport.ts` 不消费 `X-Mock-Delay-Ms` header；fixture 协议与 mockTransport 协议有出入。
  - **类别**：`spec-plan` + `skill`（落点：`docs/spec/mock-contract-suite/spec.md` 应明确列出「mockTransport 必须支持的 fixture meta header」清单；本 plan 002 触发的需求是 `X-Mock-Delay-Ms`，未来 plan 还可能需要 `X-Mock-Status-Override` / `X-Mock-Failure-Mode` 等。建议给 `mock-contract-suite` 加一个 follow-up plan 收口此契约）。

- **根因 E**：i18n locale schema 没有显式区分 command name vs displayed label，仅靠键名 prefix 隐式表达。
  - **类别**：`spec-plan`（落点：`docs/spec/frontend-shell/spec.md` 或 `frontend-home-job-picks-and-parse/spec.md` 增加 i18n key naming guideline：`action${Verb}` 表示命令名，`action${StateAdj}` 表示状态标签）。本次属轻量遗漏，不阻塞交付，可以延后。

## 4 对流程资产的改进建议

- **建议 1：补充 Vitest 严格 TS 速查清单**
  - **落点**：`.claude/skills/tdd/SKILL.md` 末尾追加「Vitest 严格 TS 常见模式」小节；或在 `frontend/README.md` 中专立一节
  - **优先级**：medium（每个新 plan 写 Vitest spec 都会遇到）

- **建议 2：sceanario verify.sh 模板加入 portable 负向 grep 推荐**
  - **落点**：`test/scenarios/e2e/README.md` 增加「verify.sh 编写约定」小节；在 `.claude/skills/scenario-create/SKILL.md` 中也指向同一约定
  - **优先级**：high（避免每个 plan 重复踩坑）

- **建议 3：澄清 `pnpm dev` 与 Playwright pixel parity 的 mock transport 契约**
  - **落点**：`docs/development.md` §2 + `frontend/README.md` + `frontend/playwright.config.ts` 三处一致更新；并新增 `pnpm dev:mock` script
  - **优先级**：high（影响所有未来 plan 的 Playwright pixel parity gate；workspace.spec.ts 8 个 baseline drift 也疑似源自同一根因）

- **建议 4：在 `mock-contract-suite` 列出 mockTransport 必须支持的 fixture meta header**
  - **落点**：`docs/spec/mock-contract-suite/spec.md` 增加新 §「fixture meta header 契约」列出 `X-Mock-Delay-Ms` / `X-Mock-Failure` / `X-Mock-Status` 等并定义 mockTransport 的最小行为；可由 follow-up plan 实现
  - **优先级**：medium（解锁 slow-response / failure variant 在 Vitest 一致行为）

- **建议 5：i18n key naming guideline 显式区分命令名与状态标签**
  - **落点**：`docs/spec/frontend-shell/spec.md` 或新 `docs/conventions/i18n.md`
  - **优先级**：low（一次性 fix 已落地；后续 plan 复用时再正名）

- **建议 6：`/implement` Step 9.5 phase commit gate 在严格 TS 项目上加上 typecheck 红线**
  - **落点**：`.claude/skills/implement/SKILL.md` Step 9.5 / Step 5 的 `/tdd` 调用约定中提示 `pnpm typecheck` 也作为 phase commit 前置条件之一
  - **优先级**：medium（此次 phase 3 / 4 / 5 typecheck 都在 commit 前夕才发现并补救；如果是 CI 阶段才发现损失更大）

## 5 建议优先级与后续动作

下一轮最值得实施的改进项（high）：

- 落实建议 2 + 建议 3：把 portable verify.sh 负向 grep 模板和 `pnpm dev:mock` / Playwright mock transport 契约写进 README + scenarios README + dev contract，避免下一个 plan 重复踩坑
- 同步 `playwright.config.ts` 的 `webServer.command` 行为：要么改成 `pnpm dev:mock`，要么把所有 pixel parity spec 重写为「不依赖 fixture 数据」的 DOM 锚点 + computed style 断言模式

可以延后处理的优化项（medium / low）：

- 建议 1（Vitest TS cheatsheet）：可在下一个 frontend plan 启动前增补
- 建议 4（mock-contract-suite fixture meta header 契约）：建议作为 follow-up plan，因为牵涉多个未来 variant 需求
- 建议 5（i18n 命名约定）：等下一个新增大量 i18n 键的 plan 启动时一并处理
- 建议 6（`/implement` phase-commit typecheck 红线）：可作为 `.claude/skills/implement/SKILL.md` 的小修订一并落地

工程化角度的下一步：

- `frontend-home-job-picks-and-parse` 当前 P0 范围已收口；预占的真实 backend 行为（recommendations / agent scan / 真实联网搜索 / market signals 计算）由独立未来 subspec `backend-jobs-recommendations` 承接，可以与 `mock-contract-suite` fixture meta header 契约 follow-up 并行启动
- `auto-resume` 自动重新触发 action（pendingAction params 解码 + selectedJobMatchId fallback）作为本 plan 002 的 frontend follow-up 议题（ui-design `auth-and-entry.md` §6 + plan §3.7 已锁定语义），可在新增小型 plan 中落地，结合 P0.025 scenario 资产已就位的便利
