# UI-Design Pixel Parity Gate Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-08

**关联计划**: [plan](./plan.md)

## Phase 1: Playwright 基础设施

- [ ] 1.1 引入 Playwright + chromium；验证: structural test 断言 `frontend/package.json` 包含 `@playwright/test` devDep 与 `test:pixel-parity` / `test:pixel-parity:install` script；focused test 断言运行 `pnpm exec playwright --version` 不抛错；缺失 chromium 时 `setup.sh` fail loudly（exit ≠ 0 + 可读提示），不允许 silent skip
- [ ] 1.2 配置 `frontend/playwright.config.ts`；验证: structural test 断言 config 声明 `testDir: ./tests/pixel-parity`、`projects` 包含 desktop (1440×900) + mobile (390×844) 两项、`webServer.command` 指向 `node ./scripts/serve-pixel-parity.mjs`、`webServer.url` 指向 `/health`、`expect.toHaveScreenshot` 默认阈值与 outputDir 设置；解析后 `defineConfig` 不抛 schema error
- [ ] 1.3 静态 server fixture；验证: focused test 断言 `frontend/scripts/serve-pixel-parity.mjs` 启动后 `/health` 返回 200、`/index.html` 与 `/ui-design/index.html` 都能 200 加载、缺失目录时 process exit ≠ 0 并打印明确缺失路径；server 仅依赖 node 内置模块（无第三方依赖）

## Phase 2: DOM + computed style parity

- [ ] 2.1 TopBar DOM + computed style；验证: `frontend/tests/pixel-parity/topbar.spec.ts` 在 desktop + mobile 两个 project 下断言 frontend dist 与 ui-design 加载后 5 个 `topbar-nav-*` testid 都存在、文本随 lang 一致、`getComputedStyle()` 读出的 height / padding / gap / border-bottom-width / background-color 在两边 1px / 1 hex 容差内；`aria-current` / `aria-pressed` 在两边等价
- [ ] 2.2 Auth / Profile / Settings / Placeholder DOM 锚点；验证: `frontend/tests/pixel-parity/screens.spec.ts` 加载 hash route `#auth_login` / `#profile` / `#settings` / `#company_intel` 后断言对应 D2 testid（`route-*` / `ei-auth-{shell,card}` / `ei-screen-{shell,card}`）在 frontend dist 渲染；ui-design 在同 hash 下渲染等价 DOM 结构（同源标题 / 卡片节奏 / 文案）；卡片 padding / border / border-radius computed 值一致

## Phase 3: Layout + bounding box parity

- [ ] 3.1 Desktop viewport bounding box；验证: `frontend/tests/pixel-parity/layout.spec.ts` 在 desktop project 上断言 `app-shell-topbar` `getBoundingClientRect()` 完全在 [0, 0, 1440, 58] 内；TopBar primary nav / display controls / user area / 五入口两两不重叠；auth login `ei-auth-card` 与 `ei-auth-side` 同行排列、`right(side) ≤ left(card)`；profile / settings shell 不溢出右侧
- [ ] 3.2 Mobile viewport 响应式；验证: `layout.spec.ts` 在 mobile project 上断言 TopBar `right ≤ 390`、五入口 testid 仍存在；auth shell 双列在 mobile 视口里折叠为单列时 `width(side) ≈ width(card)`；`route-auth_login` `bottom ≤ document.body.scrollHeight`

## Phase 4: Screenshot diff

- [ ] 4.1 默认 warm/light 截图基线；验证: `frontend/tests/pixel-parity/screenshot.spec.ts` 在两个 project 下加载 frontend home 与 ui-design home，关闭动画 + 等待 `document.fonts.ready`，调用 `expect(page).toHaveScreenshot('frontend-home-light.png')` 与 `'ui-design-home-light.png'`，并对两份截图做对比断言两边 RGB 差异 ≤ 阈值（desktop 2000、mobile 2500，可在 spec 内调整）；baseline 通过 `--update-snapshots` 维护，不入 git
- [ ] 4.2 Dark + customAccent 视觉差异对照；验证: `screenshot.spec.ts` 切到 dark 后 `--ei-color-bg-canvas` 解析为 `#16130e`、与 light baseline 截图差异像素数 > 5000；激活 customAccent (h=200, c=0.18) 后 TopBar swatch / accent 元素 computed `background` 含 `oklch(`、与 light baseline 截图差异像素数 > 1500（防 customAccent 静默失效）

## Phase 5: Scenario + handoff

- [ ] 5.1 派生 `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/`；验证: README.md 含 Playwright 安装步骤、`data/{seed-input,expected-outcome}.md` 描述输入与期望、`scripts/{setup,trigger,verify,cleanup}.sh` 形成完整契约：setup 预检 chromium + dist、trigger 跑 `pnpm --filter @easyinterview/frontend test:pixel-parity`、verify 断言 trigger.log 含 8 spec PASS（topbar / screens / layout / screenshot × desktop / mobile）+ 0 failed、cleanup 清理 setup marker；`test/scenarios/e2e/INDEX.md` 添加 P0.006 行
- [ ] 5.2 BDD-Gate: 验证 E2E.P0.006 通过；验证: 跑通 setup→trigger→verify→cleanup 完整链路；trigger.log 落在 `.test-output/e2e/p0-006-ui-design-pixel-parity-gate/trigger.log`；verify 阻断旧 entry / 旧文案回流的 grep 模式；BDD-checklist 同步勾选并写入证据
- [ ] 5.3 Handoff；验证: `frontend/README.md` §2.7 更新 pixel parity gate 入口、jsdom fast smoke 与 Playwright gate 分工、`--update-snapshots` baseline 重生成方式、E2E.P0.006 scenario 入口、chromium 安装步骤；`make docs-check` zero drift；负向搜索：`frontend/`、active spec/plan/checklist 不再有「Playwright follow-up 待派生」类语句

## Phase 6: Regression

- [ ] 6.1 D1 + D2 jsdom 行为 regression；验证: `pnpm --filter @easyinterview/frontend test` 全量通过（含 D1 + D2 + 新 jsdom 结构断言），`E2E.P0.001 / 002 / 004 / 005` setup→trigger→verify→cleanup 重跑全部通过
- [ ] 6.2 真实 build smoke；验证: `pnpm --filter @easyinterview/frontend build` 与根 `make build` 均通过；`frontend/dist/index.html` 存在且可被 serve-pixel-parity.mjs 正确托管
- [ ] 6.3 Active-scope 负向搜索；验证: `grep -R` `frontend/` + active 文档无遗留 retired-module testid 或文案；Playwright config / spec / scenario 中无私有 brand 字体名 / 旧设计参考；`@playwright/test` 是新增的唯一 visual-rendering 依赖，没有引入 cypress / puppeteer / @emotion / styled-components
