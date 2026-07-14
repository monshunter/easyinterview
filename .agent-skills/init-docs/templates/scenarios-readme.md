# 场景测试框架

## 1 目标

本目录只承载操作真实运行环境的端到端场景测试约定。BDD 文档中的 domain behavior test 留在代码 owner，不因使用 Given/When/Then 就进入本目录。

当前仓库只维护**一套**本地场景契约；阶段差异
通过场景编号和产品阶段来表达，而不是通过多套环境拆分。E2E 通过真实 HTTP API 或浏览器访问已运行 frontend，并让业务请求落到真实 backend；外部依赖按需由项目本地 dev stack 提供。

当前标准套件：

| 套件 | 用途 | 默认执行方式 |
|------|------|--------------|
| `e2e` | 通过真实 API/UI 驱动运行中前后端的用户主链路与高风险链路 | automated / hybrid |

所有设计、计划、`BDD-Gate`、场景创建、环境操作与调查诊断，均以本目录文档为真理源。

## 2 基本原则

- 测试环境只保留一套本地运行契约；不要为普通场景默认引入 Kind / K8s / Helm
- 只有真实 HTTP API / browser UI 流程才能分配 `E2E.P0.001`、`E2E.P1.003` 等场景编号
- `go test`、Vitest/npm test、pytest、lint、source-contract、fixture parity、build 和 package smoke 都是代码层 gate，不得出现在 E2E `trigger.sh` / `verify.sh` 或场景 PASS 证据中
- 浏览器 E2E 必须访问真实 frontend，且业务请求落到真实 backend；fixture transport、dev mock、jsdom 或 request interception/mock backend 不是 E2E
- domain Behavior ID 可以由代码层 behavior test 验证，不创建 E2E 目录；纯配置/内部/tooling BDD-N/A
- 根 `make test` 统一承接前后端全量单测回归，与 E2E 分层执行，不得嵌入 E2E 场景
- checklist 中的真实 E2E `BDD-Gate` 引用 E2E ID，不引用 `AC-*`
- 场景断言优先验证用户可见结果、关键证据与下一步行动建议
- 不预设 Helm、外部 Git 平台或未在本仓库声明的组件名，环境契约必须由本仓库文档定义
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
4. 按 README 建立或验证目标套件声明的已运行 frontend/backend 与外部依赖；缺少明确脚本时不得自行杜撰 Kind / K8s 入口

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

`trigger.sh` 必须发起真实 HTTP 请求或驱动真实浏览器 UI；`verify.sh` 必须校验真实响应、持久化结果或用户可见状态。任何脚本若只运行代码层测试、lint 或 build，应删除该 E2E 目录并把测试留在代码 owner。

## 7 运行输出

默认输出目录为 `.test-output/`。

## 8 环境污染与恢复

当场景失败时，恢复顺序必须是：

1. 清理场景自身资源
2. 定位并恢复受污染的共享组件
3. 只有在前两者失败时才全量重建环境
