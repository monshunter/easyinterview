# Home Screenshot Alignment 交付复盘报告

> **日期**: 2026-07-19
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付原地修订并完成 [Home JD Import and Parse](../spec/frontend-home-job-picks-and-parse/plans/001-home-jd-import-and-parse/plan.md) Phase 25 与 [App Shell Visual System](../spec/frontend-shell/plans/002-app-shell-visual-system/plan.md) Phase 22：按参考图重做桌面 TopBar、Home Hero、单一 intake card、字节计数、Resume 选择、主按钮、隐私提示和最近模拟面试横向记录，同时保留 generated client、鉴权、幂等导入、Resume gate 与 practice handoff。

成功证据：

- Home / TopBar / i18n focused tests 102/102 PASS；`MockInterviewCard` keyboard regression 14/14 PASS。
- 根 `make test`：615 个工具测试（4615 subtests）、全部 backend package、frontend 127 files / 1044 tests PASS。
- frontend typecheck、lint 与 production build PASS；两个 owner `context.yaml` 校验通过。
- Chrome `1916x821` 下 intake card `y=242 / h=325`、recent record `y=679 / h=130`、页面无横向溢出；`390x844` 下 `scrollWidth=390`，light/dark 与 disabled/enabled 状态切换正常，console warning/error 为 0。
- Header / INDEX / docs link / diff / pruning gate 均通过，`real_residuals=0`。

## 2 会话中的主要阻点/痛点

- 首轮只完成 TopBar 时，页面主体仍与参考图存在显著差异。
  - **证据**：用户明确指出排版、按钮、Icon、字体等主体内容尚未对齐；随后 Home RED 以 10 个失败点覆盖旧 Hero、分散式表单、旧 card grid 和缺失的 responsive CSS。
  - **影响**：若把局部 chrome 通过当作页面完成，会产生错误收口和额外返工。
- focused Home 测试通过后，根回归仍发现 6 个旧 `home-hero-label` 消费者。
  - **证据**：第一次 `make test` 中 frontend 为 1038/1044，失败集中在 App 与 canonical/negative routing tests；改用当前 `home-hero-title` 路由锚点后第二次为 1044/1044。
  - **影响**：装饰性 testid 被跨模块当作页面身份，视觉删改会扩大无关测试维护面。
- Chrome 截图 API 返回 JPEG bytes，而初始证据文件使用 `.png` 扩展。
  - **证据**：`file` 显示 JFIF/JPEG；收口时已改为 `.jpg` 并复核 1916x821 与 390x844 尺寸。
  - **影响**：若只看扩展名，可能留下不可审计的伪 PNG 证据。

## 3 根因归类

- 页面级视觉任务缺少“主体区域逐块 + 目标 bbox + Chrome viewport”的完成定义时，容易被局部成功替代整体完成。
  - **类别**：spec-plan
- canonical route 回归依赖装饰性子节点，而不是稳定的 route root / 页面标题合同。
  - **类别**：spec-plan
- 截图内容类型由浏览器实现决定，不能从目标文件名推断。
  - **类别**：无需仓库改动

## 4 对流程资产的改进建议

- 后续页面截图对齐继续在 owner plan 中显式列出完整页面分区、参考 viewport、关键 bbox、字体/Icon/按钮状态、no-overflow 和 console gate；本次 Phase 25 / Phase 22 已按此方式固化。
  - **落点**：spec-plan
  - **优先级**：high
- 路由级回归统一依赖 `route-*` root 或稳定 heading，不再依赖 eyebrow、装饰插图等可能随视觉调整删除的节点。
  - **落点**：spec-plan
  - **优先级**：medium
- Chrome 证据写盘后始终执行 `file` 或 magic-byte 检查，并让扩展名与真实 MIME 一致。
  - **落点**：无需仓库改动
  - **优先级**：medium

## 5 建议优先级与后续动作

- 下一轮最高价值动作：继续沿用当前 Home Phase 25 的页面级 Chrome 验收方式调整“面试”页面，不复用 Home 的宽记录布局到 Workspace 固定卡片语义。
- 次优先动作：在后续视觉变更触及 App 路由测试时，顺手将残余页面身份断言收敛到 route root / heading；不为本次已修复的 6 个断言另建 sibling plan。
- 可延后项：若后续多次出现截图 MIME 不一致，再评估把 magic-byte 检查抽成共享证据脚本；单次问题暂不扩张治理资产。
