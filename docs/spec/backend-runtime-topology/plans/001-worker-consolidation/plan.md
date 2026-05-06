# Worker Consolidation

> **版本**: 1.0
> **状态**: completed
> **更新日期**: 2026-05-06

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

取消 P0 独立 worker 进程，把后台任务语义收敛到 backend 内部 runner，同时确保 config、events/jobs、observability、local dev 与受影响文档全部采用同一口径。

## 2 背景

`backend-auth` 实施中已经证明 magic link 邮件派发可以由 backend-internal goroutine 完成本地闭环。继续保留 `cmd/worker`、`WORKER_LISTEN_ADDR` 和 worker producer 会让尚未上线的产品承担不必要的进程、配置和观测复杂度。本计划把该架构选择落成一个独立 owner，避免散落在 backend-auth 或 historical roadmap 中。

## 3 质量门禁分类

- **Plan 类型**: `code-internal` + `contract` + `tooling` + `docs`。
- **TDD 策略**: 通过 `/implement backend-runtime-topology/001-worker-consolidation repo` -> `/tdd` 执行；每个代码/契约 checklist item 先写或调整 focused test / lint expectation，再实现最小变更；配置、event lint、codegen drift、Go/TS contract tests 是主要 Red-Green 证据。
- **BDD 策略**: 不适用：本计划不新增用户可见 UI、API 行为或端到端业务流程；它调整内部运行拓扑、配置字典、生成物和开发期 gate。
- **替代验证 gate**: `validate_context.py`、`sync-doc-index --check`、`make lint-config`、`make lint-events`、`make codegen-events-check`、focused Go/Python/TS tests、`make codegen-check`、`make docs-check`、worker/process 旧口径负向搜索、`git diff --check`。

## 4 实施步骤

### Phase 1: 文档与 owner 锁定

#### 1.1 创建 owner spec / plan

创建 `backend-runtime-topology` spec、history、plan、checklist、context 与 plans INDEX，锁定 P0 取消独立 worker 进程、后台任务收敛到 backend internal runner、开发期观测消费端不作为 gate。

#### 1.2 原地修订受影响 truth source

修订 engineering-roadmap、ADR-Q2/Q4/Q5/Q1、secrets-and-config、event-and-outbox-contract、observability-stack、local-dev-stack、backend-auth、ai-provider handoff 等 active 文档，把独立 worker 前置改为 backend internal runner / background executor 口径；历史记录可保留，但当前执行正文不得继续要求 worker 进程。

### Phase 2: Config and process cleanup

#### 2.1 删除 worker runtime config

从 `.env.example`、`deploy/dev-stack/.env.example`、`config/config.yaml`、`config/README.md`、A4 spec/plan/checklist、config bindings、validator 和 tests 中删除 `WORKER_LISTEN_ADDR` / `worker.listenAddr`。

#### 2.2 删除独立 worker entrypoint

删除 `backend/cmd/worker`，更新 Makefile、lint allowlist、env dict scanner、A3 README/config comments、migrations comments、dev-stack docs，使 `go build ./cmd/...` 不再构建 worker binary。

### Phase 3: Event producer and generated contract

#### 3.1 更新 B3 producer enum

把 `shared/events.yaml` envelope producer enum 和所有后台产出事件从 `worker` 改为 `backend_async`；更新 `scripts/lint/events_inventory.py` 与 tests。

#### 3.2 重生成 event artifacts

运行 `make codegen-events`，同步 Go/TS envelope producer、event schemas、baseline、fixtures 与 generated constants；更新 Go/TS event contract tests 的 fixture 期望。

### Phase 4: Observability and dev dependency boundary

#### 4.1 固化开发期观测不依赖消费端

修订 F1/A2 文档与相关 plan，使 metrics/logs 的生产、内存 registry 单测和 `/metrics` 文本出口成为 dev gate；Prometheus/Grafana/OTel/Loki 只作为生产或显式可选 profile，不得阻塞 `make dev-up`、TDD 或 BDD。

#### 4.2 负向搜索旧进程口径

对 active truth source 和 runtime code 执行 zero-reference 搜索：`WORKER_LISTEN_ADDR`、`worker.listenAddr`、`cmd/worker`、`build-worker`、独立 worker 进程默认前置等不得残留；允许历史 changelog / completed assessment 中作为历史证据出现。

### Phase 5: Verification and lifecycle

#### 5.1 执行 focused gates

运行 config、events、backend、frontend event contract 和 docs focused gates，确认直接受影响面通过。

#### 5.2 执行 aggregate gates

运行 `make lint`、`make test`、`make codegen-check`、`make docs-check`、`sync-doc-index --check`、`git diff --check`；失败必须修复后复跑。

## 5 验收标准

- P0 runtime docs 和 code 不再保留独立 worker 进程入口或配置。
- `job_type` / outbox / async queue 权重继续存在，但解释为 backend internal runner 消费的任务契约。
- B3 generated Go/TS/schema/baseline 与 `shared/events.yaml` 一致，producer 使用 `backend_async`。
- 开发期观测 gate 不依赖 Prometheus/Grafana/OTel/Loki 实例。
- 本计划 checklist、Header、INDEX、context 与验证证据一致。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 删除 worker 入口误删任务契约 | Phase 3 只改 producer/拓扑，不删除 `job_type` / outbox / retry |
| 历史文档负向搜索误伤 | 将 active truth source 与历史记录分开处理；历史 changelog 可保留 |
| codegen drift 漏同步 | `make codegen-events-check` 与 `make codegen-check` 作为强 gate |
| 观测消费端再次阻塞研发 | F1/A2 文档明确 consumer 只进生产或 opt-in profile |
