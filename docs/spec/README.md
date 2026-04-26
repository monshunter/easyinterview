# 设计文档规范

## 1 目录结构

```
docs/spec/
├── README.md                    # 本文件，规范说明
├── TEMPLATES.md                 # 模板资产
├── INDEX.md                     # 文档索引
└── ${subject}-${type}.md        # 设计文档
```

## 2 文档元信息

所有设计文档必须在头部包含元信息：

```markdown
> **版本**: 1.0
> **状态**: active
> **更新日期**: 2025-12-27
```

**版本号说明**：
- **X（主版本）**：文档结构或目标发生重大变更时递增
- **Y（次版本）**：内容修订、补充细节时递增

**状态值说明**：

| 状态 | 含义 |
|------|------|
| `draft` | 草稿，尚未正式生效 |
| `active` | 生效中，当前有效版本 |
| `superseded` | 已被取代，需注明新文档路径 |
| `deprecated` | 已废弃，不再适用 |

## 3 命名规范

| 类型 | 命名模式 | 示例 |
|------|----------|------|
| 架构设计 | `${subject}-architecture.md` | `system-architecture.md` |
| 模块设计 | `${module}-design.md` | `font-renderer-design.md` |
| 数据结构 | `${subject}-schema.md` | `config-schema.md` |
| 接口设计 | `${subject}-interface.md` | `plugin-interface.md` |

## 4 模板资产

- [设计文档模板](./TEMPLATES.md) — 新文档的最小正文结构与验收标准写法
- 有未决产品或架构选择时，正文应显式包含 `用户决策 / 待确认事项`

## 4.1 协作约定

协作前必须先阅读本目录 `README.md` 的命名与状态规则。
起草或修改正文时，必须参考同目录 `TEMPLATES.md`。
不得把 `README.md` 当作可复制模板。

## 5 检查清单

完成设计文档前，确认以下事项：

- [ ] 文件命名符合规范
- [ ] 包含元信息头（版本、状态、更新日期）
- [ ] 已更新 `INDEX.md` 索引
- [ ] 版本号已递增（如为更新）
- [ ] 状态值正确设置
