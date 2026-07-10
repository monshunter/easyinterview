# UI Design 文档规范

## 1 目录定位

`docs/ui-design/` 用于沉淀 EasyInterview 当前阶段已经收敛的 UI 信息架构、模块边界、用户流程和认证策略。

当前阶段的设计目标以运行时静态 UI 原型为准：

```text
ui-design/index.html
└─ ui-design/src/app.jsx
   ├─ TopBar 一级导航
   ├─ ROUTE_ALIASES 范围外原型路由归一
   └─ screens 当前可渲染页面
```

当前 Git 跟踪的运行入口是 `ui-design/index.html`。

也就是说，本目录是当前运行时 UI 目标，不是“未来可能目标”的独立草案；它必须和当前静态页面展示的运行时交互保持一致。静态页面调整后，`docs/ui-design/` 需要同步校对，避免文档与设计稿漂移。

`ui-design/canvas.html` 是设计画板总览。判断目标模块时，以 `src/app.jsx` 的一级导航、`ROUTE_ALIASES`、`normalizeRoute` 后的 `activeRouteName` 和实际 screen 返回内容为准；画板标签不能覆盖运行时交互，范围外组件不得作为目标页面依据。

## 2 目录关系

```text
docs/ui-design/
└─ 当前阶段目标 UI 架构、模块划分、用户流程和模块边界
```

当前仓库只维护 `docs/ui-design/` 作为 UI 设计文档入口；范围外页面不保留独立说明文档。

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

## 5 文档格式

正文文档必须包含标准 Header：

```markdown
> **版本**: 1.0
> **状态**: active
> **更新日期**: YYYY-MM-DD
```

状态值沿用全局文档规范；UI 设计文档通常使用 `draft`、`active`、`completed`。

## 6 必须记录的设计决策

目标 UI 文档必须显式记录：

- 当前顶部导航和用户菜单。
- 顶部主题色、暗色模式、语言下拉和设置页字体预设等全局显示控制。
- 当前保留的用户任务、模块和边界。
- 首页、面试、简历之间的入口关系。
- 未登录用户在什么动作上触发登录。
- 登录、注册、验证、退出登录的页面流；验证码重发与更换邮箱留在 `auth_verify` 内完成。
- 页面框架图或流程图，优先使用文本图。
- 报告、复练当前轮、进入下一轮之间的边界。
- 简历原始来源、解析文本快照和结构化内容之间的关系（平铺简历资产，无版本树）。
- 账号资料补全不得承担候选人画像产品语义；设置与隐私只保留账号层能力。
- `ROUTE_ALIASES` 对范围外 route 的归一关系，以及范围外页面不得作为目标页面。
- 通过 `ROUTE_ALIASES` 归一的范围外组件；`voice` 不保留 route alias，电话模式必须使用 `practice` 显式 `phone` 参数。

## 7 检查清单

- [ ] 文档包含标准 Header
- [ ] 文档与 `ui-design/index.html` 当前静态 UI 一致
- [ ] 目标模块边界清晰，不与范围外模块混用
- [ ] 用户流程包含默认入口、主要动作和返回路径
- [ ] 登录策略包含触发动作、拦截方式、取消路径和成功后接续动作
- [ ] 全局显示控制不被误写成业务模块或认证门槛
- [ ] 已区分当前目标路由、范围外原型归一路由、范围外直达页面和范围外组件
- [ ] 已更新 `docs/ui-design/INDEX.md`
