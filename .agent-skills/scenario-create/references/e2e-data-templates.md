# E2E Data Templates

本参考文件只描述当前仓库 `e2e` 套件常见的数据 stub 形态；套件编号、目录结构和执行约束仍以 `test/scenarios` README / INDEX 为准。

场景数据夹具应围绕单一共享环境中的真实用户闭环组织：
- JD 导入与目标岗位工作台
- 岗位定制练习会话
- 证据化报告与错题本
- 复练或简历改写建议
- 删除 / 导出等隐私敏感链路

## 推荐数据文件

默认生成以下两个 stub：

- `data/seed-input.md`
- `data/expected-outcome.md`

按需补充：

| 文件 | 适用场景 | 说明 |
|------|----------|------|
| `data/jd.txt` | JD 导入 | 原始 JD 文本 |
| `data/session-request.json` | 练习创建 | 输入参数、轮次、角色、语言 |
| `data/report-fragments.json` | 报告校验 | 关键观察、建议、准备度档位 |
| `data/privacy-request.json` | 删除/导出 | 用户请求参数 |
| `data/interaction-script.md` | hybrid 场景 | 关键交互顺序与人工边界 |

## 模板提示

### `data/seed-input.md`

```markdown
# Scenario Seed Input

## User Intent
- 用户当前目标

## Inputs
- JD / 面试题 / 复盘素材

## Preconditions
- 已存在的用户资料、目标岗位或过往会话
```

### `data/expected-outcome.md`

```markdown
# Expected Outcome

## User-visible result
- 页面 / API / 报告中应出现的关键结果

## System result
- 状态变化、异步任务完成、索引更新等

## Non-goals
- 本场景不校验的内容
```

## 约束

- 优先使用产品语义，而不是技术实现细节命名
- 期望结果优先写用户可见价值，再补充系统状态
- 若涉及评分，必须写清“证据 + 档位/置信度”，避免伪精确分数
- hybrid 场景必须明确人工动作边界与可复现证据
