# E2E 场景套件说明

## 1 套件定位

`e2e` 是当前唯一活跃的场景套件。

它通过 repo-tracked 本地 runner，覆盖关键用户闭环与高风险链路。
阶段差异通过场景 ID 中的 `P0` / `P1` / `P2` / `P3` 表达，而不是通过多套环境拆分。

## 2 环境契约

- 环境类型：本地 runner（Go / Vitest / Playwright / browser smoke 等，按具体场景 README 声明）
- 环境模式：单一 repo-tracked 场景契约；外部依赖按需通过本地 dev stack 启动
- 不默认创建或要求 Kind / K8s / Helm；若未来 release owner 引入部署级场景，必须先修订本 README 和对应 owner plan
- 若提供 `test/scenarios/env-setup.sh` / `test/scenarios/env-cleanup.sh`，则它们是首选入口

## 3 场景设计要求

- 每个场景应验证一个可独立收口的用户行为切片
- 场景必须以真实用户目标组织，而不是按内部实现细节拆碎
- README 中必须明确 Given / When / Then
- 结果断言必须覆盖"用户得到了什么证据"与"用户接下来能做什么"

## 4 编号与索引

- 编号格式：`E2E.P{阶段}.{序号}`，例如 `E2E.P0.001`
- 目录格式：`p{阶段小写}-{序号}-<slug>`，例如 `p0-001-<slug>`
- `INDEX.md` 行格式见同目录 INDEX.md

## 5 污染控制

- 优先保证场景自有数据可清理
- 不把共享环境初始化逻辑塞进单个业务场景
- 若 cleanup 后仍残留资源，必须在结果中明确记录
