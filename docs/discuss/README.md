# 讨论存档规范

## 1 目录结构

```
docs/discuss/
├── README.md                   # 本文件，规范说明
├── TEMPLATES.md                # 模板资产
├── INDEX.md                    # 文档索引
├── ${subject}-by-${agent}.md   # Agent 分析文档
└── ${subject}-summary.md       # 汇总文档
```

## 2 Agent 名称标准化

| Agent | 标准名称 |
|-------|----------|
| Claude | `claude` |
| Codex | `codex` |
| Gemini | `gemini` |

## 3 命名规范

| 类型 | 命名模式 | 示例 |
|------|----------|------|
| Agent 分析 | `${subject}-by-${agent}.md` | `cache-strategy-by-claude.md` |
| 汇总文档 | `${subject}-summary.md` | `cache-strategy-summary.md` |

## 4 模板资产

- [讨论模板](./TEMPLATES.md) — Agent 分析文档与汇总文档的推荐结构

## 4.1 协作约定

协作前必须先阅读本目录 `README.md` 的命名与汇总规则。
起草或修改正文时，必须参考同目录 `TEMPLATES.md`。
不得把 `README.md` 当作可复制模板。

## 5 汇总规则

汇总各 Agent 建议时，必须遵守以下规则：

1. **标注来源**：每个决策点必须标注来源 Agent
2. **保留链接**：汇总文档必须链接到各 Agent 的原始分析
3. **明确决策**：必须明确最终采纳的方案

## 6 检查清单

完成讨论文档前，确认以下事项：

- [ ] Agent 名称使用标准化格式
- [ ] 汇总文档标注了决策来源
- [ ] 汇总文档链接了各 Agent 分析
- [ ] 已更新 `INDEX.md` 索引
