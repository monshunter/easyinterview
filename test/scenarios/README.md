# 场景测试框架

## 1 目标

本目录只承载操作真实运行环境的端到端场景测试约定。BDD 文档中的 domain behavior test 留在代码 owner，不因使用 Given/When/Then 就进入本目录。

当前仓库只维护一套本地场景契约。阶段差异通过场景编号和产品阶段表达，不通过多套环境拆分；默认场景编排只使用 shell / Python，外部依赖按需由 `make dev-up` 提供。明确要求完整容器部署时，使用 `make dev-container-up` 启动 frontend `http://127.0.0.1:10800` 与 backend `http://127.0.0.1:10801/api/v1`；这只是同一真实环境的可选部署形态。E2E 只接收针对真实运行环境的 HTTP API 调用，或针对真实运行前后端的浏览器 UI 操作；场景不得把包测试、源码检查或构建包装成 E2E。

当前标准套件：

| 套件 | 用途 | 默认执行方式 |
|------|------|--------------|
| `e2e` | 通过真实 API/UI 驱动运行中前后端的用户主链路与高风险链路 | automated / hybrid |

所有设计、计划、`BDD-Gate`、场景创建、环境操作与调查诊断，均以本目录文档为真理源。

## 2 基本原则

- 测试环境只保留一套本地 runner 契约；不要为普通 P0 场景默认引入 Kind / K8s / Helm。
- 场景编号必须使用行为导向 ID，格式为 `E2E.P{阶段}.{三位序号}`。
- checklist 引用真实 E2E 时使用已登记场景编号，不引用 `AC-*`；domain Behavior ID 留在代码 owner，不在本目录分配场景资产。
- 场景断言优先验证用户可见结果、关键证据与下一步行动建议。
- E2E 证据只有两类：调用正在运行的真实 backend HTTP API；或浏览器操作正在运行、使用真实 backend 的 frontend。名称中包含 `e2e`、使用 Playwright 或位于 `test/scenarios/e2e/`，都不能替代这项事实。
- `go test`、Vitest、`pnpm/npm test`、pytest、lint、source/contract test、codegen/drift check、build、直接 provider/eval CLI 和仅数据库断言均不是 E2E trigger/verify 证据，不得产生场景 PASS marker。前后端代码层单元测试作为整体全量回归，统一由根 Makefile 的 `make test` 承接；场景脚本不得选择性重编排或重复执行这些测试。
- Playwright 只有在相关用户动作真实访问 host-run frontend/backend 时才算 E2E。fixture client、mock/stub transport、dev-mock、component runner、静态截图基线，以及被 `route.fulfill` / `route.abort` 接管的请求均不算 E2E 证据；混合脚本必须从场景合同和 verify marker 中排除这些阶段。
- 配置默认值、合法/非法 override、跨字段约束和配置投影不单建 E2E 场景；它们只由 typed loader / validator owner 的一组契约测试负责。只有配置改变了真实用户流程时，E2E 才验证该流程的可见行为，不重复数值边界或配置 wiring。
- 执行者顺序默认是 AI Agent 先运行场景脚本、环境 preflight 和可自动化证据检查；需要真实浏览器观察或真实 provider 凭证时，再由人工或浏览器 Agent 接手补齐同一场景输出目录下的脱敏证据。
- `test/scenarios/` 新增场景工具只允许 shell / Python；需要账号、数据或环境准备时，应放在场景目录或 `_shared/` 下，不得新增正式 `backend/cmd` / Go helper 作为验收依赖。
- 不预设 Helm、外部 Git 平台或非当前项目组件名，环境契约必须由本仓库文档定义。
- 清理与污染控制属于场景契约的一部分；失败后必须优先检查环境污染。

## 3 目录结构

```text
test/scenarios/
├── README.md
├── _shared/
│   ├── README.md
│   └── scripts/
│       └── local-dev-runtime.sh
└── e2e/
    ├── README.md
    └── INDEX.md
```

当前共享脚本：

- `test/scenarios/_shared/scripts/local-dev-runtime.sh`
- `test/scenarios/env-setup.sh`
- `test/scenarios/env-status.sh`
- `test/scenarios/env-verify.sh`
- `test/scenarios/env-cleanup.sh`
- `test/scenarios/env-redeploy.sh`

需要人工参与的真实 provider / 浏览器 UAT 也必须登记为标准 `e2e` 场景目录，保留 `README.md`、`data/` 和四段脚本契约。此类 `hybrid` 场景的脚本负责环境准备、材料/配置/隐私 preflight、统一 result artifact 与 `MANUAL_REQUIRED` 状态；人工或浏览器 Agent 只补齐无法无密钥自动完成的真实操作证据。本地真实前后端联调只使用 `deploy/dev-stack/.env` 作为单一 env 来源，场景不得维护自己的独立 `.env` 副本。

## 3.1 共享环境生命周期

共享测试环境与本地前后端联调环境的生命周期由顶层 `test/scenarios/env-*.sh` 脚本管理，独立于任何具体场景目录：

| 入口 | 作用 | 根 Makefile 等价入口 |
|------|------|----------------------|
| `test/scenarios/env-setup.sh` | 启动 dev-stack 外部依赖并可选执行 migrations | `make scenario-env-setup` |
| `test/scenarios/env-status.sh` | 输出当前 dev-doctor JSON 状态 | `make scenario-env-status` |
| `test/scenarios/env-verify.sh` | 验证共享环境 readiness | `make scenario-env-verify` |
| `test/scenarios/env-cleanup.sh` | 清理共享环境，默认保留命名卷 | `make scenario-env-cleanup` |
| `test/scenarios/env-redeploy.sh` | 按 `deps/backend/frontend/all` 刷新依赖或 build artifacts | `make scenario-env-redeploy TARGET=<target>` |
| `deploy/dev-stack` 的 `full-container` profile | 停止仓库 PID 文件管理的 host-run app，构建并启动依赖、migration、backend、frontend | `make dev-container-up`；停止用 `make dev-container-down` |
| `env-cleanup.sh --with-volumes` → `env-setup.sh --with-migrations` → `env-redeploy.sh all` → `env-verify.sh` | 清理数据、重跑迁移、重编译并重启 host-run backend/frontend，再验证 readiness | `make scenario-env-reset-redeploy` |

`env-setup.sh` / `env-status.sh` / `env-verify.sh` / `env-redeploy.sh` 必须给出开发者可接管的信息：frontend/backend/Mailpit/MinIO 地址、`.test-output/local-dev/{backend,frontend}.log`、PID 文件以及容器日志命令。当前 host-run 口径下，`env-redeploy.sh backend|frontend|all` 不是只做 build；它必须重新启动对应宿主机前后端进程，保证用户在浏览器里看到的服务已经加载当前代码和 `deploy/dev-stack/.env`。

`make scenario-env-reset-redeploy` 是显式清数据调试入口，会删除本地 named volumes。普通重启或仅重新加载当前代码时使用 `make scenario-env-redeploy TARGET=all`，不得把“重启”默认解释为 reset。

具体场景的 `scripts/setup.sh` / `trigger.sh` / `verify.sh` / `cleanup.sh` 只负责场景数据、runner 执行证据和场景自有清理，不得把共享环境 bootstrap 私有化，也不得引用另一个具体场景作为环境前提。默认仍由 `/scenario-env setup` 或 `make scenario-env-setup` 准备依赖、宿主机运行前后端；当请求明确指定全容器部署时，改用 `make dev-container-up`，该入口会先停止仓库 PID 文件管理的 host-run backend/frontend，避免两个 backend 竞争同一异步队列，再以 `10800` / `10801` 的真实 UI/API 入口完成同一场景验收。

## 4 首次使用

1. 先读本文件。
2. 再读 `test/scenarios/e2e/README.md` 与 `test/scenarios/e2e/INDEX.md`。
3. 按目标场景 README 准备并验证共享环境，再运行真实 HTTP 或真实浏览器场景。纯 Go / Vitest / pytest、源码契约、lint 或 build 目标不属于本场景框架；前后端全量单元回归运行根 `make test`，不嵌入四段场景脚本。Playwright 也必须先证明它连接的是真实 host-run 前后端。

## 5 场景编号与目录命名

- 场景 ID：`E2E.P{阶段}.{三位序号}`
- 场景目录：`test/scenarios/e2e/p{阶段}-{三位序号}-<slug>/`
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

## 7 运行输出与证据边界

默认输出目录为 `.test-output/`。

`trigger.sh` 必须把真实 HTTP 请求或真实浏览器 runner 输出写入场景输出目录（通常是 `trigger.log`）。`verify.sh` 必须检查实际执行证据：目标真实环境、HTTP 方法/状态或浏览器用户动作、业务结果，以及 runner 的成功状态。禁止只检查文件存在，也禁止使用 package test、source contract、lint、build 或数据库直查结果替代 E2E 证据。

若场景运行时使用完整 email、token、cookie 或其他敏感值，证据脱敏与 `verify.sh` 负向检查必须同时覆盖原文和 URL percent-encoded 等实际可进入 URL/日志的等价表示；只替换原文不能视为隐私 gate 通过。验证器应使用当前 run 的真实敏感输入生成检查值，避免仅匹配固定样例而漏掉变体。

`setup.sh` / `cleanup.sh` 可以为隔离目的写入或删除固定场景数据，环境 readiness 也可以作为 preflight；这些操作均不能独立产生 PASS。若一个 Playwright 流程同时包含真实请求和 route mock，只有未被拦截且实际到达真实服务的部分可以写入 Given / When / Then、expected outcome 和 verify marker。

Hybrid 场景若已经完成 AI Agent preflight 但缺少本地真实凭证、浏览器操作或人工观察证据，`verify.sh` 必须写出 `result=MANUAL_REQUIRED` 等价 JSON artifact 并退出 0；不得把它标记为 full PASS，也不得退化为框架 ERROR。补齐脱敏证据后，同一场景可再次运行并转为 PASS。

涉及 cookie、浏览器状态、provider 输出或原始业务内容的 hybrid 场景必须把输出保留策略做成可执行合同：setup 在写入当前 `run_id` 前删除输出目录内全部旧顶层文件、目录和符号链接且不得跟随外部链接；运行中只允许场景明确定义的有限文件名；PASS/FAIL cleanup 分别使用 allowlist，未知文件或敏感键即删除并使当前证据失败。不得只按少量已知文件名清理，因为历史 `state.json`、Playwright 目录或改名后的 cookie 文件同样属于污染。

## 8 环境污染与处理

当场景失败时，处理顺序必须是：

1. 清理场景自身资源。
2. 定位并重置受污染的共享组件。
3. 只有在前两者失败时才全量重建环境。
