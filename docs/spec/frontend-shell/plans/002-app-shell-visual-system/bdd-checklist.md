# App Shell Visual System BDD Checklist

> **版本**: 2.5
> **状态**: completed
> **更新日期**: 2026-07-20

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## `BDD.SHELL.VISUAL.001` Shell 与显示偏好

- [x] Owner behavior tests 覆盖 shell 渲染、76px desktop chrome、TopBar 零主题菜单、圆形用户名首字符单一设置入口、Settings Appearance/Account/Privacy 状态、主题 draft/save/error、固定字体与业务状态隔离。
- [x] 根 `make test` 执行对应 Vitest；该结果是代码层行为证据，不是 E2E PASS。
- [x] Source/responsive/font gate 不包装成场景；真实设置路径仅引用 001 对 `E2E.P0.101` 的原地扩展。

## `BDD.SHELL.VISUAL.002` 跨页面有框操作按钮圆角一致性

- [x] Token/source behavior test 枚举正式有框 action consumer，并断言统一消费 `--ei-radius-control: 8px`；旧尖角 action 值零残留且无全局 `button` selector。<!-- verified: 28 CSS selectors + 10 inline recovery actions; focused source contract PASS -->
- [x] Affected component/root tests 保持默认、focus、disabled、pending、error-recovery 与点击行为；circular/pill、borderless link/back、card/input/status 例外不被误改。<!-- verified: focused 62/62 and root make test 615 passed/4615 subtests -->
- [x] Chrome desktop/mobile 验收 Settings 保存/退出/注销/dialog action 与 TopBar、Parse/Practice/Report/Resume 样本 computed radius、截图和 viewport containment；记录为人工 browser evidence，不声明 E2E。<!-- verified: Settings desktop/mobile and read-only cross-route recovery samples PASS; overflow=false; console warnings/errors=0 -->

## `BDD.SHELL.PRESET.003` Ocean / Plum / Forest 视觉主题

- [x] Owner tests 证明三预设 metadata、六组 theme-mode selector 和确认的 OKLCH accent/accent-soft matrix；Forest 可从 Custom 清晰回退，`warm`/unknown 不成为 active theme。（2026-07-20 focused owner tests 57/57 PASS。）
- [x] Current-run Chrome 在 desktop/mobile light/dark 验证四个一级选项、三预设 computed accent/soft、Save/Custom 层级与 viewport containment；记录为人工 browser evidence，不声明 E2E。（2026-07-20 E2E.P0.101 1440/390 real Chromium PASS。）
- [x] 账号保存与跨重登恢复只引用 001 `BDD.SHELL.SETTINGS.PRESET.007` / `E2E.P0.101`，本 owner 不创建重复真实场景。（2026-07-20 E2E.P0.101 trigger/verify/cleanup PASS。）
