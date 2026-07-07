# Worker Consolidation Checklist

> **版本**: 1.5
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: 文档与 owner 锁定

- [x] 1.1 创建 `backend-runtime-topology` owner spec / plan / checklist / context / plans INDEX；验证: context validator 通过，`docs/spec/INDEX.md` 与 plans INDEX 投影新 subject
  <!-- verified: 2026-05-06 command="python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-runtime-topology/plans/001-worker-consolidation/context.yaml --docs-root docs --target repo && python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check" -->
- [x] 1.2 原地修订受影响 active truth source；验证: engineering-roadmap、ADR-Q2/Q4/Q5/Q1、secrets-and-config、event-and-outbox-contract、observability-stack、local-dev-stack、backend-auth、ai-provider handoff 均改为 backend internal runner / app-produced metrics 口径
  <!-- verified: 2026-05-06 command="rg old worker/config/process terms across active specs + python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --fix-index" evidence="active specs/ADRs use backend internal runner/app-produced metrics; remaining exact non-current tokens are owner negative assertions, tests, completed plans, or history rows" -->

## Phase 2: Config and process cleanup

- [x] 2.1 删除独立后台进程 runtime config；验证: focused config test 断言独立后台进程 listen 配置不在 `DefaultEnvBindings` / validator / `.env.example` / `config/config.yaml`；`cd backend && go test ./internal/platform/config -run 'Test(DefaultEnvDictionaryOmitsWorkerListenAddr|ValidateProdAllSecretsPasses|ValidateProdRejectsDevDefaultDeploymentDependencies)' -count=1` 通过
  <!-- verified: 2026-05-06 command="cd backend && go test ./internal/platform/config -run 'Test(DefaultEnvDictionaryOmitsWorkerListenAddr|ValidateProdAllSecretsPasses|ValidateProdRejectsDevDefaultDeploymentDependencies)' -count=1" -->
- [x] 2.2 删除独立后台进程 entrypoint；验证: 独立后台进程目录不存在，`go build ./cmd/...` 不构建独立后台 binary，`make lint-getenv-boundary` allowlist 不包含独立后台 entrypoint，`python3 scripts/lint/env_dict.py --repo-root .` 通过
  <!-- verified: 2026-05-06 command="python3 scripts/lint/env_dict.py --repo-root . && make lint-getenv-boundary && cd backend && go build ./cmd/..." evidence="env_dict OK: 24 keys; build cmd/... OK; getenv boundary OK" -->

## Phase 3: Event producer and generated contract

- [x] 3.1 更新 B3 producer enum；验证: `python3 -m unittest scripts.lint.events_inventory_test` Red-Green 通过，`python3 scripts/lint/events_inventory.py shared/events.yaml shared/jobs.yaml` 通过
  <!-- verified: 2026-05-06 command="python3 -m unittest scripts.lint.events_inventory_test && python3 scripts/lint/events_inventory.py shared/events.yaml shared/jobs.yaml" -->
- [x] 3.2 重生成 event artifacts；验证: `make codegen-events-check` 通过，`cd backend && go test ./internal/shared/events -count=1` 与 `pnpm --filter @easyinterview/frontend test -- src/lib/events/envelope.test.ts` 通过
  <!-- verified: 2026-05-06 command="make codegen-events-check && cd backend && go test ./internal/shared/events -count=1 && pnpm --filter @easyinterview/frontend test -- src/lib/events/envelope.test.ts" note="event contract files staged before drift gate so generated artifacts are compared against the updated truth source" -->

## Phase 4: Observability and dev dependency boundary

- [x] 4.1 固化开发期观测不依赖消费端；验证: F1/A2 docs 明确 dev gate 只依赖内存 registry、应用 `/metrics` 文本出口和日志，不要求 Prometheus/Grafana/OTel/Loki 实例
  <!-- verified: 2026-05-06 command="rg -n 'Prometheus|Grafana|OTel|Loki|metrics' docs/spec/observability-stack/spec.md docs/spec/local-dev-stack/spec.md deploy/dev-stack/README.md deploy/dev-stack/docker-compose.yaml Makefile -S" evidence="A2 default dev stack excludes observability consumers; F1 treats them as production/optional consumers; /metrics/logs are the dev gate" -->
- [x] 4.2 负向搜索非当前进程口径；验证: active runtime/code/config/docs 对独立后台进程 env / entrypoint / build target / 默认前置无当前残留；tests、lint fixtures 与 owner negative assertions 只作为验证输入
  <!-- verified: 2026-05-06 command="rg -n 'WORKER_LISTEN_ADDR|worker\\.listenAddr|cmd/worker|build-worker|Asynq worker' .env.example config backend frontend shared scripts Makefile -S && rg -n 'WORKER_LISTEN_ADDR|worker\\.listenAddr|cmd/worker|build-worker|Asynq worker' docs/spec/*/spec.md docs/spec/engineering-roadmap/decisions/*.md docs/spec/*/history.md docs/spec/backend-runtime-topology/plans/001-worker-consolidation/*.md -S" evidence="runtime/code/config only has negative config test assertions; docs hits are new owner negative assertions, history, or explicit P0-not-retained text" -->
- [x] 4.3 L2 remediation: 落地 `scripts/lint/runtime_topology.py` / `make lint-runtime-topology`，扫描 active code/doc handoff 中的 non-current standalone worker process 口径回流；同步清理 non-current code comments 与 completed owner plan/checklist 正文残留；验证: Red `python3 -m unittest scripts.lint.runtime_topology_test` 先失败，Green 后 `python3 -m unittest scripts.lint.runtime_topology_test`、`make lint-runtime-topology`、`make lint` 通过
  <!-- verified: 2026-05-06 command="python3 -m unittest scripts.lint.runtime_topology_test && make lint-runtime-topology && make lint" evidence="Red failed before runtime_topology.py existed; Green/unit lint/aggregate lint pass; active scan reports runtime_topology OK (274 active files scanned)" -->
- [x] 4.4 L2 remediation: 补强 `scripts/lint/runtime_topology.py` false-negative 覆盖，拦截 active handoff 中 non-current producer / listen-address / binding shorthand 与 async-runtime shorthand；同步修订 B3/A4/ADR-Q3 active handoff 文案；验证: Red `python3 -m unittest scripts.lint.runtime_topology_test` 先失败，Green 后 `python3 -m unittest scripts.lint.runtime_topology_test`、`make lint-runtime-topology`、`make lint` 通过
  <!-- verified: 2026-05-06 command="python3 -m unittest scripts.lint.runtime_topology_test && make lint-runtime-topology && make lint" evidence="Red failed on new false-negative fixtures; Green/unit lint/aggregate lint pass; active scan reports runtime_topology OK (274 active files scanned)" -->
- [x] 4.5 L2 remediation: 补强 `scripts/lint/runtime_topology.py` 对 `scripts/` tooling 面和 raw producer 字段形态的 false-negative 覆盖；验证: Red `python3 -m unittest scripts.lint.runtime_topology_test` 先失败，Green 后 `python3 -m unittest scripts.lint.runtime_topology_test`、`make lint-runtime-topology`、`make lint` 通过；负向 fixture 覆盖 `scripts/lint/env_dict.py` / `scripts/lint/getenv_boundary.go` 回流 `cmd/worker` / `WORKER_LISTEN_ADDR`，以及 `shared/events.yaml` / `shared/events/baseline/events.v1.json` 回流 `producer: worker` / `"producer": "worker"`
  <!-- verified: 2026-05-07 command="python3 -m unittest scripts.lint.runtime_topology_test && make lint-runtime-topology && make lint" evidence="Red initially missed scripts/raw producer fixtures; Green/unit lint/aggregate lint pass; active scan reports runtime_topology OK (299 active files scanned)" -->
- [x] 4.6 L2 remediation: 补强 `scripts/lint/runtime_topology.py` 对 YAML / JSON 结构化 producer 字段跨行 `worker` 回流的 false-negative 覆盖；验证: Red `python3 -m unittest scripts.lint.runtime_topology_test` 先失败，Green 后 `python3 -m unittest scripts.lint.runtime_topology_test`、`make lint-runtime-topology`、`make lint` 通过；负向 fixture 覆盖 `shared/events.yaml` 的 producer block / list 与 `shared/events/schemas/*.json` 的 producer enum / values 形态
  <!-- verified: 2026-05-07 command="python3 -m unittest scripts.lint.runtime_topology_test && make lint-runtime-topology && make lint" evidence="Red failed because multiline YAML/JSON producer worker fixtures were accepted; Green/unit lint/aggregate lint pass; structured producer scan rejects shared/events.yaml block/list and shared/events/schemas JSON enum values" -->
- [x] 4.7 L2 remediation: 补强 `scripts/lint/runtime_topology.py` 对 owner plan/checklist 当前 handoff 口径的 false-negative 覆盖；验证: Red `python3 -m unittest scripts.lint.runtime_topology_test` 先失败，Green 后 `python3 -m unittest scripts.lint.runtime_topology_test`、`make lint-runtime-topology`、`make lint` 通过；owner 负向断言、tests 与 lint 定义允许保留验证输入，但 current handoff / build / runtime / verification 入口不得使用 non-current worker 口径
  <!-- verified: 2026-05-07 command="python3 -m unittest scripts.lint.runtime_topology_test && make lint-runtime-topology && make lint" evidence="Red failed because owner plan/checklist current handoff fixtures were accepted; Green/unit lint/aggregate lint pass; owner plan/checklist are scanned with negative-evidence exceptions, runtime_topology OK reports 301 active files scanned" -->

## Phase 5: Verification and lifecycle

- [x] 5.1 focused gates 通过；验证: `make lint-config`、`make lint-events`、focused backend/frontend tests、context validator、`sync-doc-index --check` 通过
  <!-- verified: 2026-05-06 command="make lint-config && make lint-events && python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/backend-runtime-topology/plans/001-worker-consolidation/context.yaml --docs-root docs --target repo && python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check && cd backend && go test ./internal/platform/config ./internal/shared/events -count=1 && pnpm --filter @easyinterview/frontend test -- src/lib/events/envelope.test.ts" -->
- [x] 5.2 aggregate gates 通过；验证: `make lint`、`make test`、`make codegen-check`、`make docs-check`、`git diff --check` 通过
  <!-- verified: 2026-05-06 command="make lint && make test && make codegen-check && make docs-check && git diff --check && git diff --cached --check" -->
