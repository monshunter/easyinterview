# 场景测试框架

## 1 目标

本目录承载项目的 BDD / 端到端场景测试约定。

当前仓库只维护**一个** Kind 本地测试环境；阶段差异
通过场景编号、BDD 文档和产品阶段来表达，而不是通过多套环境拆分。

当前标准套件：

| 套件 | 用途 | 默认执行方式 |
|------|------|--------------|
| `e2e` | 围绕真实用户目标的主链路与高风险链路验证 | automated / hybrid |

所有设计、计划、`BDD-Gate`、场景创建、环境操作与调查诊断，均以本目录文档为真理源。

## 2 基本原则

- 测试环境只保留一个本地 Kind 集群
- 场景编号必须使用行为导向 ID，例如 `E2E.P0.001`、`E2E.P1.003`
- checklist 中的 `BDD-Gate` 只能引用场景编号，不引用 `AC-*`
- 场景断言优先验证用户可见结果、关键证据与下一步行动建议
- 不预设 Helm、外部 Git 平台或历史项目组件名，环境契约必须由本仓库文档定义
- 清理与污染控制属于场景契约的一部分；失败后必须优先检查环境污染

## 3 目录结构

```text
test/scenarios/
├── README.md
├── _shared/
│   └── README.md
├── e2e/
│   ├── README.md
│   └── INDEX.md
```

说明：

- `e2e/` 是当前唯一活跃的场景套件

约定但非强制的辅助路径：

- `test/scenarios/_shared/scripts/common.sh`
- `test/scenarios/_shared/scripts/image-cache.sh`
- `test/scenarios/env-setup.sh`
- `test/scenarios/env-cleanup.sh`

如果这些脚本不存在，Agent 必须退回到 README 中定义的手工或 repo-tracked 命令，不能自行杜撰。

## 4 首次使用

1. 先读本文件
2. 再读 `test/scenarios/e2e/README.md` 与 `test/scenarios/e2e/INDEX.md`
3. 若存在镜像缓存脚本，先执行：
   ```bash
   ./test/scenarios/_shared/scripts/image-cache.sh pull
   ```
4. 按 README 建立或验证单一 Kind 环境

## 5 场景编号与目录命名

- 场景 ID：`E2E.P0.001`、`E2E.P1.004`
- 场景目录：`test/scenarios/e2e/p0-001-<slug>/`
- 目录 slug 应表达用户价值或业务动作，而非仅写技术术语

## 6 场景契约

每个场景目录至少应包含：

- `README.md`
- `scripts/setup.sh`
- `scripts/cleanup.sh`
- `scripts/trigger.sh`、`scripts/verify.sh`（按需）
- `data/` 目录中的最小输入 / 期望输出夹具

## 7 运行输出

默认输出目录为 `.test-output/`。

## 8 环境污染与恢复

当场景失败时，恢复顺序必须是：

1. 清理场景自身资源
2. 定位并恢复受污染的共享组件
3. 只有在前两者失败时才全量重建环境
