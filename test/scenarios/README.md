# 场景测试框架

## 1 目标

本目录承载 EasyInterview 的 BDD / 端到端场景测试约定。

当前仓库只维护一套本地场景契约。阶段差异通过场景编号、BDD 文档和产品阶段表达，不通过多套环境拆分；默认场景编排只使用 shell / Python，外部依赖按需由 `make dev-up` 提供。场景脚本可以调用已有产品 runner（例如既有包测试、Vitest、Playwright、browser smoke）作为被验证对象，但不得把场景专属依赖实现为新的 `backend/cmd` / Go helper 进程。

当前标准套件：

| 套件 | 用途 | 默认执行方式 |
|------|------|--------------|
| `e2e` | 围绕真实用户目标的主链路与高风险链路验证 | automated / hybrid |

Companion 目录：

| 目录 | 用途 | 约束 |
|------|------|------|
| `manual-uat` | 人工验收 runbook、账号/session 材料、输入材料与 checklist | 必须先有 `docs/spec/*/plans/*` owner plan；不是标准 runner 套件，不要求四段脚本契约，也不以 stub/mock 自动化证据冒充人工真实 provider UAT |

所有设计、计划、`BDD-Gate`、场景创建、环境操作与调查诊断，均以本目录文档为真理源。

## 2 基本原则

- 测试环境只保留一套本地 runner 契约；不要为普通 P0 场景默认引入 Kind / K8s / Helm。
- 场景编号必须使用行为导向 ID，例如 `E2E.P0.001`、`E2E.P1.003`。
- checklist 中的 `BDD-Gate` 只能引用场景编号，不引用 `AC-*`。
- 场景断言优先验证用户可见结果、关键证据与下一步行动建议。
- `test/scenarios/` 新增场景工具只允许 shell / Python；需要账号、数据或环境准备时，应放在场景目录或 `_shared/` 下，不得新增正式 `backend/cmd` / Go helper 作为验收依赖。
- 不预设 Helm、外部 Git 平台或历史项目组件名，环境契约必须由本仓库文档定义。
- 清理与污染控制属于场景契约的一部分；失败后必须优先检查环境污染。

## 3 目录结构

```text
test/scenarios/
├── README.md
├── _shared/
│   └── README.md
└── e2e/
    ├── README.md
    └── INDEX.md
```

约定但非强制的辅助路径：

- `test/scenarios/_shared/scripts/common.sh`
- `test/scenarios/_shared/scripts/image-cache.sh`
- `test/scenarios/env-setup.sh`
- `test/scenarios/env-status.sh`
- `test/scenarios/env-verify.sh`
- `test/scenarios/env-cleanup.sh`
- `test/scenarios/env-redeploy.sh`

如果这些脚本不存在，Agent 必须退回到 README 中定义的手工或 repo-tracked 命令，不能自行杜撰。

`manual-uat/` 只允许作为已登记 owner plan 的 companion 材料目录出现。新增或大幅修改人工验收材料前，必须先更新对应 spec/plan/checklist/BDD，并在 manual runbook 顶部链接 owner plan。

## 3.1 共享环境生命周期

共享测试环境与本地前后端联调环境的生命周期由顶层 `test/scenarios/env-*.sh` 脚本管理，独立于任何具体场景目录：

| 入口 | 作用 | 根 Makefile 等价入口 |
|------|------|----------------------|
| `test/scenarios/env-setup.sh` | 启动 dev-stack 外部依赖并可选执行 migrations | `make scenario-env-setup` |
| `test/scenarios/env-status.sh` | 输出当前 dev-doctor JSON 状态 | `make scenario-env-status` |
| `test/scenarios/env-verify.sh` | 验证共享环境 readiness | `make scenario-env-verify` |
| `test/scenarios/env-cleanup.sh` | 清理共享环境，默认保留命名卷 | `make scenario-env-cleanup` |
| `test/scenarios/env-redeploy.sh` | 按 `deps/backend/frontend/all` 刷新依赖或 build artifacts | `make scenario-env-redeploy TARGET=<target>` |

具体场景的 `scripts/setup.sh` / `trigger.sh` / `verify.sh` / `cleanup.sh` 只负责场景数据、runner 执行证据和场景自有清理，不得把共享环境 bootstrap 私有化，也不得引用另一个具体场景作为环境前提。开发者可以只执行 `/scenario-env setup` 或 `make scenario-env-setup` 构建环境，然后人工或由 Agent 运行目标场景；最新 manual UAT / 本地联调也遵循该边界，真实 backend/frontend 长驻进程仍按 runbook 在宿主机启动。

## 4 首次使用

1. 先读本文件。
2. 再读 `test/scenarios/e2e/README.md` 与 `test/scenarios/e2e/INDEX.md`。
3. 若存在镜像缓存脚本，先执行：

```bash
./test/scenarios/_shared/scripts/image-cache.sh pull
```

4. 按 README 建立或验证目标套件声明的本地 runner 与外部依赖；缺少明确脚本时不得自行杜撰 Kind / K8s 入口。

## 5 场景编号与目录命名

- 场景 ID：`E2E.P0.001`、`E2E.P1.004`
- 场景目录：`test/scenarios/e2e/p0-001-<slug>/`
- 目录 slug 应表达用户价值或业务动作，而非仅写技术术语。

## 6 场景契约

每个场景目录至少应包含：

- `README.md`
- `scripts/setup.sh`
- `scripts/cleanup.sh`
- `scripts/trigger.sh`
- `scripts/verify.sh`
- `data/seed-input.md`
- `data/expected-outcome.md`

## 7 运行输出

默认输出目录为 `.test-output/`。

`trigger.sh` 声称执行 Vitest、Playwright、pytest、Go test、lint 或其他 runner 时，必须把 runner 输出写入场景输出目录下的日志（通常是 `trigger.log`）。`verify.sh` 必须检查日志中的实际执行证据：命令/runner marker、目标测试文件或场景路径、以及 pass marker 或退出状态证据。禁止只检查测试文件、spec 文件、脚本或目录存在来代表 runner 已执行。

## 8 环境污染与恢复

当场景失败时，恢复顺序必须是：

1. 清理场景自身资源。
2. 定位并恢复受污染的共享组件。
3. 只有在前两者失败时才全量重建环境。
