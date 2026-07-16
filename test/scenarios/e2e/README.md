# E2E 场景套件说明

## 1 套件定位

`e2e` 是当前唯一活跃的场景套件。

它通过真实运行环境中的 HTTP API 或浏览器 UI 操作覆盖关键用户闭环与高风险链路。阶段差异通过场景 ID 中的 `P0` / `P1` / `P2` / `P3` 表达，而不是通过多套环境拆分。

## 2 环境契约

- 环境类型：默认是 host-run frontend/backend 上的真实 HTTP API 或浏览器 UI 场景；明确要求全容器部署时，使用同一场景契约下的 `full-container` profile；可按 README 声明 automated / hybrid
- 环境模式：单一 repo-tracked 场景契约；外部依赖按需通过 `make dev-up` 启动
- 不默认创建或要求 Kind / K8s / Helm；若未来 release owner 引入部署级场景，必须先修订本 README 和对应 owner plan
- 共享环境 lifecycle 首选顶层入口：`test/scenarios/env-setup.sh` / `test/scenarios/env-status.sh` / `test/scenarios/env-verify.sh` / `test/scenarios/env-cleanup.sh` / `test/scenarios/env-redeploy.sh`，或根 Makefile 等价入口 `make scenario-env-*`。这些入口独立于任何具体场景目录。
- 创建、验证或重建共享环境后，顶层入口必须输出开发者可接管的服务地址和调试入口；`env-redeploy.sh backend|frontend|all` 必须重新启动当前 host-run backend/frontend，而不只刷新构建产物。
- 需要清理数据并从当前代码重新调试时，使用 `make scenario-env-reset-redeploy`；该入口组合 `env-cleanup.sh --with-volumes`、`env-setup.sh --with-migrations`、`env-redeploy.sh all`、`env-verify.sh`。普通重启不清数据，使用 `make scenario-env-redeploy TARGET=all`。
- 明确要求全容器验收时，使用 `make dev-container-up` 构建并启动真实 frontend/backend，默认入口与 host-run 统一为 `http://127.0.0.1:10900` 与 `http://127.0.0.1:10901/api/v1`；健康检查、日志与停止分别使用 `dev-container-doctor`、`dev-container-logs`、`dev-container-down`。
- 具体场景 `setup.sh` 只做场景数据准备和输出目录初始化，不得私有化共享环境 bootstrap，也不得把某个具体场景作为另一个场景的环境前置。
- 运行场景前必须先准备并验证共享环境；不同场景可在各自 README 中声明额外本地配置，例如真实 provider secret、host-run backend/frontend 进程或浏览器操作证据。
- `hybrid` 场景仍是标准 `e2e` 场景：AI Agent 先运行 setup/trigger/verify/cleanup，若缺少真实凭证或人工观察，结果应为 `MANUAL_REQUIRED`，后续由人工或浏览器 Agent 在同一输出目录补齐证据后重跑。

### 2.1 手动引导

当顶层 env 脚本不可用时，默认手动引导顺序是：`make dev-up` → `make dev-doctor` → 按目标场景 README 启动 repo-tracked runner。明确指定全容器验收时，直接使用 `make dev-container-up` → `make dev-container-doctor`，并在 `10900` 上操作真实 UI、确认业务请求到达 `10901` backend；真实 secret 只放在未跟踪本地文件中。

## 3 场景设计要求

- 每个场景应验证一个可独立收口的用户行为切片。
- 场景必须以真实用户目标组织，而不是按内部实现细节拆碎。
- README 中必须明确 Given / When / Then。
- 结果断言必须覆盖“用户得到了什么证据”与“用户接下来能做什么”。
- E2E PASS 只允许来自正在运行的真实 backend HTTP API，或连接真实 backend 的 host-run frontend 浏览器操作。
- `go test`、Vitest、`pnpm/npm test`、pytest、lint、source/contract test、codegen/drift check、build、直接 provider/eval CLI 和仅数据库断言不得作为 E2E trigger/verify 阶段或 PASS marker。前后端代码层单元测试作为整体全量回归，统一运行根 Makefile 的 `make test`；E2E 场景不得挑选、拼装或重复执行其中一部分。
- Playwright 不是自动豁免。fixture/mock/stub transport、dev-mock、component runner、静态截图基线，以及被 route interception 接管的请求都不是 E2E 证据；混合流程只能保留确实到达真实前后端的行为和断言。
- 配置默认值、override、validator 和 wiring 不创建场景。只有配置改变真实用户流程时，场景才验证一次用户可见行为，不复制配置数值边界。

## 4 编号与索引

- 编号格式：`E2E.P{阶段}.{三位序号}`。
- 目录格式：`p{阶段小写}-{三位序号}-<slug>`。
- `INDEX.md` 行格式为：场景 ID、关联需求、目录、描述、执行方式、状态。

## 5 污染控制

- 优先保证场景自有数据可清理。
- 不把共享环境初始化逻辑塞进单个业务场景。
- 若 cleanup 后仍有未清资源，必须在结果中明确记录。
