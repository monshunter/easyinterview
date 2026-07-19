# UI Design 文档规范

## 1 目录定位

`docs/ui-design/` 用于沉淀 EasyInterview 当前阶段已经收敛的 UI 信息架构、模块边界、用户流程、交互约束、响应式要求和认证策略。

本目录是设计文档 owner，不是可运行前端、静态画板或组件源码。项目不维护与正式 `frontend/` 重复的 UI Demo；设计决策在本目录收敛后，由对应 frontend spec/plan 直接实施，并通过正式组件测试、响应式与可访问性断言、构建以及必要的真实业务场景验证。

`docs/ui-design/` 描述“用户看到什么、如何操作、状态如何流转以及模块边界是什么”；`frontend/` 负责“如何用生产代码实现并验证”。文档不得复制完整 JSX/CSS 实现，也不得要求与另一套可运行页面做源码或像素对照。

## 2 目录关系

```text
产品 / 模块 spec
        │
        ├── docs/ui-design/  UI 架构、流程、交互与响应式约束
        │
        └── frontend/        正式实现、组件合同与可执行测试
```

本目录只维护当前 UI 设计；范围外页面不保留独立说明文档。历史决策证据由 `docs/work-journal/`、`docs/bugs/`、`docs/reports/` 和各 subject `history.md` 承担。

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

设计文档优先使用文本结构、状态表、流程表和明确的不变量。需要表达视觉约束时，记录语义化 token、布局关系、viewport 行为和可访问性要求，不粘贴另一套前端实现。

## 6 必须记录的设计决策

目标 UI 文档必须显式记录：

- 当前顶部导航、登录入口和已登录账号设置入口。
- 顶部主题色、暗色模式、语言下拉等全局显示控制，以及固定产品字体边界。
- 当前保留的用户任务、模块和边界。
- 首页、面试、简历之间的入口关系。
- 未登录用户在什么动作上触发登录。
- 登录、注册、验证、退出登录的页面流；验证码重发与更换邮箱留在 `auth_verify` 内完成。
- 页面框架图或流程图，优先使用文本图。
- 报告、复练当前轮、进入下一轮之间的边界。
- 简历原始来源、解析文本快照和结构化内容之间的关系（平铺简历资产，无版本树）。
- 账号资料补全不得承担候选人画像产品语义；“设置”只保留账号级外观、账号与隐私能力。
- 范围外 route 输入的归一关系，以及范围外页面不得作为目标页面。
- `voice` 不保留独立 route；电话模式必须使用 `practice` 显式 `phone` 参数。
- loading、empty、error、disabled、responsive、keyboard 和 screen-reader 等用户可观察状态。

## 7 实施与验证合同

- 新页面或大幅 UI 修订先更新本目录对应文档和 owner spec/plan，再修改 `frontend/`。
- 正式前端应满足文档描述的架构、流程、交互、响应式和可访问性不变量，但不要求逐字转写 CSS/DOM。
- plan/checklist 必须为重要约束指定可执行证据，例如 component test、route test、responsive/browser smoke、fixture-backed state test 或真实 API/UI E2E。
- 浏览器截图可以用于人工审查或针对性回归，但不与第二套 Demo 做像素对照，也不作为唯一完成证据。
- 文档与正式前端不一致时，先判断是设计变化还是实现漂移：设计变化先修订文档，明显实现漂移由原 owner plan/test 修复。

## 8 检查清单

- [ ] 文档包含标准 Header
- [ ] 目标模块边界清晰，不与范围外模块混用
- [ ] 用户流程包含默认入口、主要动作、错误/空状态和返回路径
- [ ] 登录策略包含触发动作、拦截方式、取消路径和成功后接续动作
- [ ] 全局显示控制不被误写成业务模块或认证门槛
- [ ] 已区分当前目标 route、范围外归一输入和范围外组件
- [ ] 响应式与可访问性约束可由正式前端测试验证
- [ ] 未引入重复 UI Demo、源码复刻或双源像素对照要求
- [ ] 已更新 `docs/ui-design/INDEX.md`
