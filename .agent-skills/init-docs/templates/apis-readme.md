# API 定义规范

## 1 目录结构

```
docs/apis/
├── README.md                    # 本文件，规范说明
├── TEMPLATES.md                 # 模板资产
├── INDEX.md                     # 文档索引
├── ${service}-openapi.json      # OpenAPI 规格
├── ${subject}.schema.json       # JSON Schema
└── ${service}-api.json          # 自定义 API 定义
```

## 2 文件格式

所有 API 定义文件使用 JSON 格式，必须包含版本信息：

```json
{
  "version": "1.0.0",
  "info": {
    "title": "API 名称",
    "description": "API 描述"
  }
}
```

## 3 命名规范

| 类型 | 命名模式 | 示例 |
|------|----------|------|
| OpenAPI 规格 | `${service}-openapi.json` | `generator-openapi.json` |
| JSON Schema | `${subject}.schema.json` | `config.schema.json` |
| 接口定义 | `${service}-api.json` | `export-api.json` |

## 4 版本管理

使用语义化版本号（SemVer）：`MAJOR.MINOR.PATCH`

| 变更类型 | 版本号变化 | 示例 |
|----------|------------|------|
| 重大变更（不兼容） | MAJOR +1 | 1.0.0 → 2.0.0 |
| 新增功能（向后兼容） | MINOR +1 | 1.0.0 → 1.1.0 |
| Bug 修复 | PATCH +1 | 1.0.0 → 1.0.1 |

## 5 模板资产

- [API 模板](./TEMPLATES.md) — OpenAPI 与 JSON Schema 的最小示例

## 5.1 协作约定

协作前必须先阅读本目录 `README.md` 的命名与版本规则。
起草或修改正文时，必须参考同目录 `TEMPLATES.md`。
不得把 `README.md` 当作可复制模板。

## 6 检查清单

完成 API 定义前，确认以下事项：

- [ ] JSON 格式有效（可通过 JSON 验证工具检查）
- [ ] 包含版本信息
- [ ] 文件命名符合规范
- [ ] 已更新 `INDEX.md` 索引
