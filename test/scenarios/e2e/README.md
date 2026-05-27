# E2E 场景套件说明

## 1 套件定位

`e2e` 是当前唯一活跃的场景套件。

它通过 repo-tracked 本地 runner 覆盖关键用户闭环与高风险链路。阶段差异通过场景 ID 中的 `P0` / `P1` / `P2` / `P3` 表达，而不是通过多套环境拆分。

## 2 环境契约

- 环境类型：本地 runner（Go / Vitest / Playwright / browser smoke 等，按具体场景 README 声明）
- 环境模式：单一 repo-tracked 场景契约；外部依赖按需通过 `make dev-up` 启动
- 不默认创建或要求 Kind / K8s / Helm；若未来 release owner 引入部署级场景，必须先修订本 README 和对应 owner plan
- 共享环境 lifecycle 首选顶层入口：`test/scenarios/env-setup.sh` / `test/scenarios/env-status.sh` / `test/scenarios/env-verify.sh` / `test/scenarios/env-cleanup.sh` / `test/scenarios/env-redeploy.sh`，或根 Makefile 等价入口 `make scenario-env-*`。这些入口独立于任何具体场景目录。
- 具体场景 `setup.sh` 只做场景数据准备和输出目录初始化，不得私有化共享环境 bootstrap，也不得把某个具体场景作为另一个场景的环境前置。

### 2.1 手动引导

当顶层 env 脚本不可用时，手动引导顺序是：`make dev-up` → `make dev-doctor` → 按目标场景 README 启动 repo-tracked runner。需要本地前后端联调或 manual UAT 时，先用 `test/scenarios/env-setup.sh` 构建共享依赖环境，再按 runbook 在宿主机启动 backend/frontend；真实 secret 只放在未跟踪本地文件中。

## 3 场景设计要求

- 每个场景应验证一个可独立收口的用户行为切片。
- 场景必须以真实用户目标组织，而不是按内部实现细节拆碎。
- README 中必须明确 Given / When / Then。
- 结果断言必须覆盖“用户得到了什么证据”与“用户接下来能做什么”。
- 当场景依赖 Vitest / Playwright / pytest / Go test / lint 等 runner 时，`verify.sh` 必须检查 runner 日志中的执行 marker、目标测试路径和 pass marker；不能只检查测试文件或脚本存在。

## 4 编号与索引

- 编号格式：`E2E.P{阶段}.{序号}`，例如 `E2E.P0.001`。
- 目录格式：`p{阶段小写}-{序号}-<slug>`，例如 `p0-001-default-home-shell`。
- `INDEX.md` 行格式为：场景 ID、关联需求、目录、描述、执行方式、状态。

## 5 污染控制

- 优先保证场景自有数据可清理。
- 不把共享环境初始化逻辑塞进单个业务场景。
- 若 cleanup 后仍残留资源，必须在结果中明确记录。
