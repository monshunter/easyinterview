# UI Design 文档规范

## 1 目录定位

`docs/ui-design/` 用于沉淀 EasyInterview 当前阶段已经收敛的 UI 信息架构、模块边界、用户流程和认证策略。

当前阶段的设计目标以运行时静态 UI 原型为准：

```text
ui-design/index.html
└─ ui-design/src/app.jsx
   ├─ TopBar 一级导航
   ├─ ROUTE_ALIASES 历史原型路由归一
   └─ screens 当前可渲染页面
```

当前 Git 跟踪的运行入口是 `ui-design/index.html`。

也就是说，本目录不再作为“未来可能目标”的独立草案存在；它必须和当前静态页面展示的运行时交互保持一致。静态页面调整后，`docs/ui-design/` 需要同步校对，避免文档与设计稿漂移。

`ui-design/canvas.html` 是设计画板总览。判断目标模块时，以 `src/app.jsx` 的一级导航、`ROUTE_ALIASES`、`normalizeRoute` 后的 `activeRouteName` 和实际 screen 返回内容为准；画板标签不能覆盖运行时交互，已移除的历史组件不得作为目标页面恢复。

## 2 目录关系

```text
docs/ui-design/
└─ 当前阶段目标 UI 架构、模块划分、用户流程、删除/降级范围
```

当前仓库只维护 `docs/ui-design/` 作为 UI 设计文档入口。历史 UI 盘点内容若需要保留，应迁入本目录并在 `INDEX.md` 中登记。

## 3 目录结构

```text
docs/ui-design/
├── README.md
├── TEMPLATES.md
├── INDEX.md
└── ${subject}.md
```

## 4 命名规范

| 类型 | 命名模式 | 示例 |
|------|----------|------|
| 总体架构 | `${subject}-architecture.md` | `ui-architecture.md` |
| 用户流程 | `${flow}-flow.md` | `user-flow.md` |
| 模块划分 | `module-${subject}.md` | `module-map.md` |
| 认证入口 | `auth-${subject}.md` | `auth-and-entry.md` |
| 范围裁剪 | `${subject}-scope.md` | `removed-modules-and-scope.md` |

## 5 文档格式

正文文档必须包含标准 Header：

```markdown
> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD
```

状态值沿用全局文档规范：`draft`、`active`、`completed`、`superseded`、`deprecated`。

## 6 必须记录的设计决策

目标 UI 文档必须显式记录：

- 当前顶部导航和用户菜单。
- 顶部主题色、暗色模式、语言切换和设置页字体预设等全局显示控制。
- 保留哪些用户任务和模块。
- 删除、降级或合并哪些历史页面。
- 首页、岗位推荐、模拟面试、简历、复盘之间的入口关系。
- 未登录用户在什么动作上触发登录。
- 登录、注册、验证、重置、退出登录的页面流。
- 页面框架图或流程图，优先使用文本图。
- 报告、复练当前轮、进入下一轮之间的边界。
- 简历原件、结构化主版本和岗位定制版本之间的关系。
- 用户画像与个人设置的边界。
- `ROUTE_ALIASES` 对旧 route 的归一关系，以及仍可直达但不属于当前目标入口的历史页面。
- 已清理或通过 `ROUTE_ALIASES` 归一的废弃 / 历史组件；`voice` 不保留 route alias，语音面试必须使用 `practice` 显式参数。

## 7 检查清单

- [ ] 文档包含标准 Header
- [ ] 文档与 `ui-design/index.html` 当前静态 UI 一致
- [ ] 目标模块边界清晰，不与已删除模块混用
- [ ] 用户流程包含默认入口、主要动作和返回路径
- [ ] 登录策略包含触发动作、拦截方式、取消路径和成功后恢复动作
- [ ] 全局显示控制不被误写成业务模块或认证门槛
- [ ] 已区分当前目标路由、历史原型归一路由、历史直达页面和废弃组件
- [ ] 已更新 `docs/ui-design/INDEX.md`
