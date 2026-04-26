# 审查报告规范

## 1 目录结构

```
docs/reports/
├── README.md                              # 本文件，规范说明
├── TEMPLATES.md                           # 报告模板资产
├── INDEX.md                               # 文档索引
└── YYYY-MM-DD-${subject}-${type}.md       # 报告文档
```

## 2 文档元信息

报告文档使用简化的元信息：

```markdown
> **日期**: 2025-12-27
> **审查人**: Claude / 用户名
```

## 3 命名规范

报告文档使用日期前缀，便于按时间排序：

| 类型 | 命名模式 | 示例 |
|------|----------|------|
| 代码审查 | `YYYY-MM-DD-${subject}-review.md` | `2025-12-27-workspace-review.md` |
| 实施评估 / 交付复盘 | `YYYY-MM-DD-${subject}-assessment.md` | `2025-12-27-cache-assessment.md` |
| 完成验证 | `YYYY-MM-DD-${subject}-verification.md` | `2025-12-27-init-verification.md` |
| 方案评价 | `YYYY-MM-DD-${subject}-evaluation.md` | `2025-12-27-mock-interview-evaluation.md` |

## 4 模板资产

- [报告模板](./TEMPLATES.md) — 代码审查、交付复盘、完成验证的标准结构

## 4.1 协作约定

协作前必须先阅读本目录 `README.md` 的报告命名与索引规则。
起草或修改正文时，必须参考同目录 `TEMPLATES.md`。
不得把 `README.md` 当作可复制模板。

## 5 检查清单

完成报告文档前，确认以下事项：

- [ ] 文件命名包含日期前缀
- [ ] 包含元信息（日期、审查人）
- [ ] 已更新 `INDEX.md` 索引
- [ ] 如有关联计划，已添加链接
- [ ] 若为交付复盘，结论与建议均有明确证据支撑
