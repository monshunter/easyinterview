# Frontend Shell Auth and Settings BDD Checklist

> **版本**: 1.31
> **状态**: active
> **更新日期**: 2026-07-20

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

> 既有已勾选静态 handoff 是历史证据；下方 Settings 行为与 `E2E.P0.101` 原地扩展记录当前完成合同。

## `BDD.SHELL.AUTH.001` Shell auth 与安全恢复

- [x] Owner behavior tests 覆盖 auth loading、profile-incomplete、pending action、protected route、settings 与 logout recovery。
- [x] 根 `make test` 已执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。

## `E2E.P0.101` 静态 handoff

- [x] 仅关联真实 email-code、session 与 profile-completion 行为，不承接通用 shell/pending/settings。
- [x] 不引用已删除场景目录、wrapper 或历史 PASS。

以上旧条目仅为静态关联审计；真实运行状态只由 `e2e-scenarios-p0/001` 维护。本轮扩展后的当前 run `e2e-p0-101-20260715114513-19516` 已 PASS，静态 INDEX 生命周期状态仍按场景契约保持 `Ready`。

## `BDD.SHELL.SETTINGS.001` 单一入口与真实账号字段

- [x] TopBar/Settings domain tests 证明设置齿轮直达、无账号 dropdown/tab、runtime `displayName/email` 完整回填、`emailMasked` 零引用、零额外 `getMe` 和既有 logout 路由。
- [x] `E2E.P0.101` 静态资产原地扩展设置入口、真实账号字段和 logout；当前真实环境 run `e2e-p0-101-20260715114513-19516` PASS。
- [x] P0.101 code-level evidence gate 证明 raw-value equality 失败与 reporter 输出落盘都不会包含当前 run 邮箱原文或 URL-encoded 等价值。

## `BDD.SHELL.SETTINGS.002` 未登录保护

- [x] Domain tests 覆盖未登录直开 Settings、业务 screen 不挂载、登录/profile setup 后 safe pendingAction 恢复。

## `BDD.SHELL.SETTINGS.DELETE.001` 账号删除状态机

- [x] Component/backend contract tests 覆盖 dialog focus/Escape/取消零副作用、pending 禁止关闭/去重、recoverable 失败恢复/同 key 重试、`401` auth probe、`202` 后 `refreshAuth()` 重探测 `/me` 并提交 unauthenticated，以及 probe 网络错误的 honest auth error；不把该破坏性路径加入共享登录 E2E。
- [x] Default `createDevMockClient` regression 覆盖 verify/profile/deleteMe 后的 `getMe` 401，证明 fixture preview 与真实 backend 的删除后 auth transition 一致。

## `BDD.SHELL.AUTH.LOCALE.001` Auth route gate 本地化

- [ ] 中文 loading/error gate 的 eyebrow/title/body 无英文硬编码；英文切换使用同一 typed keys。
- [ ] Protected business screen/API 仍不在 auth probe 完成前挂载；focused 与根回归通过。
- [ ] Chrome extension automation skill 在 current-run 真实本地页面验证中文 gate 与英文切换；该证据不标记新的 E2E ID。

## `BDD.SHELL.SETTINGS.THEME.001` 账号级主题与请求预算

- [x] Ocean / Plum / Custom 一级选项与 Save 同属固定 primary row；Custom editor/error 位于其后，desktop 展开/收起不改变 Save 纵向位置，mobile 保持顺序且无横向溢出。（2026-07-20 SettingsScreen/ScreensVisual 26/26 + source contract PASS。）
- [x] current-run Chrome 在 desktop 量测 preset/custom Save bbox 差值不超过 1px，并在 390px mobile 验证 primary row、editor 后置与可操作性。（1440×900 delta=0px；390×844 documentWidth=viewportWidth；0 warning/error。）
- [x] Ocean / Plum / Custom 一级选择器在预定义与自定义状态下始终可访问；Custom 激活后 hue/chroma 二级编辑器只在一级下方进入正常文档流，不覆盖或替换一级；hue 完整光谱与当前 hue 的 chroma 渐变可见且不破坏 range 键盘/focus 语义。（2026-07-19 focused 25/25 + Chrome desktop/mobile PASS。）
- [x] 选择 Ocean / Plum 后二级编辑器隐藏且 custom accent 清除；切换过程零网络并保留既有 Save 请求预算。（2026-07-19 component + Chrome reversible switch PASS。）
- [x] current-run Chrome 在 desktop/mobile 验证 Custom -> Ocean/Plum 可逆切换、无遮挡与无横向溢出。（1440×900、390×844；documentWidth=viewportWidth；browser error/warning=0。）
- [x] Runtime/Settings tests 证明 bootstrap/auth recovery 的 `getMe` 提供确认主题，Settings/普通 route/Practice 切换均零额外 `/me`。（2026-07-19 frontend focused 45 tests PASS。）
- [x] Appearance tests 证明预定义主题/custom hue/chroma 本地预览零网络，Save 恰好一次 `updateMe`，成功直接更新 runtime/display 且无 follow-up GET。（2026-07-19 frontend focused 45 tests PASS。）
- [x] 失败、重试、离开未保存页面、迟到响应和 invalid server projection 均 fail closed，不污染最近一次确认主题或新的认证状态。（2026-07-19 新增 race/projection/atomicity tests 后 focused PASS。）
- [x] Backend/migration/OpenAPI tests 证明 profile-only/theme-only/combined 原子更新、约束和 `getMe` round-trip。
- [x] `E2E.P0.101` 原地覆盖真实主题保存、logout/relogin 恢复；current run `e2e-p0-101-20260719082610-75505` PASS（themeSavePatch=1、settingsMountedGetMe=0、settingsRouteSwitchGetMe=0、themeRelogin=plum、cleanup PASS）。

## `BDD.SHELL.PAGES.VISUAL.002` Auth 与 Settings 目标构图

- [x] Auth/Settings visual/component tests 覆盖 desktop 宽幅布局、mobile 单列、无溢出、装饰语义和退出堆叠操作，并先证明旧构图失败。
- [x] Auth/Settings behavior、locale、请求预算、主题/退出/删除状态机在视觉重排后保持通过；验证码页不出现无 producer 的倒计时或成功状态。
- [x] Chrome extension automation skill 在 current-run 真实页面验收登录、验证码、退出、设置 desktop/mobile，并验证真实中英文切换与 pendingAction 回跳；不新增 E2E ID。`BDD.SHELL.AUTH.LOCALE.001` 的 auth probe 中间 loading/error gate 仍独立保持未完成。

## `BDD.SHELL.TRANSITION.VISUAL.003` 共享异步等待态

- [x] Shared component tests 覆盖 brand/resume/report/job 四种 code-native SVG、状态语义、可选步骤/动作、无伪百分比与装饰隔离。（最终 focused 9 files / 89 tests PASS。）
- [x] TopBar/route tests 覆盖 Practice、Parse、Reports、Generating、报告详情/记录统一显示全局 chrome 并高亮“面试”。
- [x] Responsive/reduced-motion tests 与 current-run desktop Chrome 验证视觉层级、可操作返回、无横向溢出和非必要动画降级；mobile 由同一组件的 720px contract 覆盖，不新增 E2E ID。（1920px 四态真实截图，browser error/warning=0。）

## `BDD.SHELL.SETTINGS.ART.004` Settings Header 安全插画

- [x] Settings DOM/visual tests 固定账号窗口、头像资料、柱状图、锁、盾牌对勾和星芒的独立层级，并拒绝旧山形人物稀疏线稿。（2026-07-20 RED 后 focused GREEN。）
- [x] CSS/responsive/a11y tests 证明插画从当前主题 token 派生、保持 `aria-hidden`、desktop 不横溢且窄屏沿用隐藏规则。（ScreensVisual 14/14。）
- [x] Chrome extension automation skill 在 current-run Settings desktop 页面确认目标图形结构、Header/card 对齐和 browser error/warning 为零；不新增 E2E ID。（1264×964，SVG 360×200，7 layers，documentWidth=viewportWidth，Ocean/Plum/Custom 预览与刷新恢复通过。）

## `BDD.SHELL.TOPBAR.IDENTITY.005` 语言辨识与账号首字符

- [x] TopBar/App component tests 证明中文、拉丁、前导空白与空名称 initial 派生正确，settings 仍直达且不增加账号请求或菜单。（2026-07-20 focused TopBar/App/i18n 43/43 PASS。）
- [x] TopBar visual/responsive/a11y tests 证明语言按钮使用清晰 SVG chevron、独立可见底板、展开旋转、键盘/accessible name 和 mobile containment。（TopBarVisual 18/18 + typecheck PASS。）
- [x] Chrome extension automation skill 在 current-run 真实 desktop 页面验证中文用户名首字符、语言菜单开合、设置直达、无横向溢出和 browser error/warning 为零；不新增 E2E ID。（2026-07-20 Chrome：displayName `星期无` 对应设置标记 `星`；chevron 为 20×20 底板内 14×14 SVG，展开旋转 180°；设置按钮 42×42，`documentWidth=viewportWidth=1920`，跳转 `/settings`，0 warning/error。）

## `BDD.SHELL.BACK.COPY.006` 统一返回文案

- [x] Locale/source contract 证明所有正式二三级返回控件只消费 `common.back`，中文“返回”、英文“Back”，旧目标特定 action key 无消费者。<!-- verified: 2026-07-20 evidence="backNavigationCopy 16/16 PASS after an explicit all-failing RED." -->
- [x] Component navigation tests 证明标签统一后，各页面仍进入原 route/target，并保留 trusted-context、replace/push 与 fail-closed 行为。<!-- verified: 2026-07-20 evidence="Affected owner scope 65 files/495 tests PASS, including trusted report/generating Back and page route assertions." -->
- [x] Current-run Chrome 抽样验证报告生成及至少一个其它二三级页面的中文/英文返回标签、可操作目标、无横向溢出与零 browser warning/error。<!-- verified: 2026-07-20 evidence="Generating used 返回 at desktop/mobile with zero overflow; Report and Reports used Back after real language switch, Report Back reached /reports?targetJobId=<trusted-target>, Chinese was restored, and browser logs were empty." -->

## Phase 25 `BDD.SHELL.BACK.COPY.006` owner-specific exception

- [x] Locale/source contract 证明 shared Back 是默认值，Generating trusted Reports 专用 key 是唯一显式 owner 例外，Workspace fallback 与其它页面不恢复旧长文案。<!-- verified: 2026-07-20 evidence="Shared Back-copy suite passes 19/19 and both locale coverage suites pass 28/28." -->
- [x] Component navigation tests 证明专用标签只随既有 trusted destination 出现，route/target、replace/push 与 fail-closed 行为不变。<!-- verified: 2026-07-20 evidence="Generating waiting/error navigation suites retain the complete reports/workspace destination matrix while selecting copy from the same backDestination owner." -->
- [x] Current-run Chrome 验证 Generating 显示“返回面试报告”；当前 target-scoped ReportsScreen destination 由 component tests 承接，且无横向溢出或 browser error/warning。<!-- verified: 2026-07-20 evidence="Real authenticated Generating visibly rendered 返回面试报告 before automatic ready handoff; the transient state expired before a reliable live click, so the route claim is intentionally limited to code-level navigation evidence." -->
