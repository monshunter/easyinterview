# 项目文档

## 1 文档导航

| 目录 | 用途 | 入口 |
|------|------|------|
| [Spec 文档](./spec/) | Spec-centric 规格、计划、Checklist、context | [INDEX](./spec/INDEX.md) |
| [审查报告](./reports/) | Review、评估报告 | [INDEX](./reports/INDEX.md) |
| [讨论存档](./discuss/) | Agent 分析讨论 | [INDEX](./discuss/INDEX.md) |
| [UI Design 文档](./ui-design/) | 整理后的 UI 流程、模块划分与目标交互结构 | [INDEX](./ui-design/INDEX.md) |
| [API 定义](./apis/) | 接口定义（JSON） | [INDEX](./apis/INDEX.md) |
| [工作日志](./work-journal/) | 开发日志 | [INDEX](./work-journal/INDEX.md) |
| [Bug 知识库](./bugs/) | Bug 诊断记录与模式库 | [INDEX](./bugs/INDEX.md) |

## 2 文档规范速查

### 2.1 元信息格式

设计文档和计划文档必须在头部包含元信息：

```markdown
> **版本**: X.Y
> **状态**: draft | active | completed | superseded | deprecated
> **更新日期**: YYYY-MM-DD
```

**状态值说明**：

| 状态 | 含义 |
|------|------|
| `draft` | 草稿，尚未正式生效 |
| `active` | 生效中，当前有效版本 |
| `completed` | 已完成，作为历史交付记录保留 |
| `superseded` | 已被取代，需注明新文档路径 |
| `deprecated` | 已废弃，不再适用 |

Spec-centric plan 补充约定：

- 新 plan 默认采用串行 phase 格式，并位于 `docs/spec/<subspec>/plans/<NNN-plan>/`
- 每个 subspec 的计划索引位于 `docs/spec/<subspec>/plans/INDEX.md`
- `context.yaml` 是 plan 机器入口的唯一真理源
- spec / plan / checklist / context 模板示例统一放在 `docs/spec/TEMPLATES.md`

### 2.2 Markdown 层级规范

```markdown
# 文档标题      （仅用于主标题，每文档一个）
## 1 一级章节   （数字编号）
### 1.1 二级章节（层级编号）
#### 1.1.1 三级章节（层级编号）
```

### 2.3 目录规范文件

每个目录包含以下规范文件：

| 文件 | 用途 |
|------|------|
| `README.md` | 该目录的规范说明（命名、流程、检查清单） |
| `TEMPLATES.md` | 可复制模板与结构示例（按目录需要提供） |
| `INDEX.md` | 文档索引导航 |

### 2.4 协作约定

协作前必须先阅读目标目录 `README.md` 中的规则说明。
起草或修改正文时，必须参考同目录 `TEMPLATES.md`。
创建或修改 spec-centric plan 时，必须参考 `docs/spec/README.md` 与 `docs/spec/TEMPLATES.md`，不得在每个 `plans/` 目录重复 README 或模板文件。
不得把 `README.md` 当作可复制模板。

## 3 使用方式

1. 先读目标目录的 `README.md`，确认命名、状态、检查清单和索引要求。
2. 再参考对应模板文件起草文档正文；spec-centric plan 统一参考 `docs/spec/TEMPLATES.md`。
3. 文档落地后同步更新对应 `INDEX.md`。

## 4 项目目录映射

以下是本项目的代码目录分组，供 commit 拆分、日志分类等场景引用：

<!-- TODO: 根据项目实际目录结构填写映射表 -->
| 目录 | 分组 |
|------|------|
| `src/` | 核心代码 |
| `test/` | 测试 |
| `docs/` | 文档 |

## 5 关联文档

- [AGENTS.md](../AGENTS.md) — Agent 编码指令
- [工作日志规范](./work-journal/README.md) — 日志记录详细规范
