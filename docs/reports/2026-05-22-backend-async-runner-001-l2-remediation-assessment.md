# Backend Async Runner 001 L2 Remediation 交付复盘报告

> **日期**: 2026-05-22
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付覆盖 `backend-async-runner/001-internal-job-outbox-runner` 的 L2 code review 修复：`cmd/api` 生产启动链路接入 `OutboxDispatcher`，上传 presign 不再强制 MinIO 不支持的 SSE header，privacy 删除 runner 修复 Postgres JSON 参数类型，固定 audit ID 的 integration test 支持重复运行，并同步 owner spec / plan / checklist / BDD gate 与 BUG-0085。
- 已通过后端 focused gate：`go test ./cmd/api -run '^(TestMainRunnerKernelDrivesOutboxDispatcher|TestMain_SingleRuntimeShutdown)$' -count=1 -v`、`go test ./internal/runner/... ./internal/privacy/runner ./internal/upload/objectstore -count=1`、report HTTP BDD focused tests P0.041/P0.052/P0.053/P0.054/P0.055。
- 已通过 live / scenario gate：`make dev-doctor`，`p0-033-file-presign-register-roundtrip` rerun PASS，`async-runner-l2-20260522145606` 覆盖 P0.003/P0.010/P0.011/P0.012/P0.013/P0.034/P0.035/P0.060/P0.062/P0.077/P0.078/P0.080/P0.093/P0.094/P0.095/P0.096/P0.097。
- 已通过全量本地质量门禁：`go build ./...`、`go vet ./...`、`go test ./...`、`make lint-runner-legacy`、`validate_context.py --target backend`、`sync-doc-index.py --check`、`make docs-check`、`git diff --check`。

## 2 会话中的主要阻点/痛点

- 完成态计划没有证明生产启动调用方实际接入 outbox dispatcher。
  - **证据**：L2 反查发现 `backend/internal/runner` 层已有 dispatcher 能力，但 `backend/cmd/api/main.go` 的生产 runtime 只启动 worker lease loop，没有通过 `Runtime.SetOutboxDispatcher` 接入 outbox dispatch loop。
  - **影响**：历史 checklist 与单元 gate 形成 false-green，生产 API 进程无法真正消费 job completion outbox。
- Live BDD P0.033 首轮暴露多个非 mock 问题。
  - **证据**：P0.033 先后暴露本地环境缺参、MinIO 501 SSE header、privacy runner `could not determine data type of parameter $3`、固定 audit ID 重跑 duplicate key。
  - **影响**：需要多轮 runtime 修复和重跑，说明原 BDD 证据没有强制证明 live dependency、真实 provider 兼容性和 repeatability。
- 本地文件系统权限一度阻断命令执行。
  - **证据**：会话中曾出现 `Operation not permitted`，用户确认继续后权限恢复。
  - **影响**：造成短暂停顿；该问题属于本机环境状态，不应作为仓库代码缺陷处理。

## 3 根因归类

- 生产 caller 反查 gate 不够显式。
  - **类别**：skill / spec-plan
  - **说明**：生命周期类计划需要从声明的 runtime 能力反向追到生产入口、启动测试和 shutdown 语义，不能只验证 internal package 的局部行为。
- BDD live gate 对环境真实性和重复运行要求不够硬。
  - **类别**：scenario README / spec-plan
  - **说明**：P0.033 属于真实 object store、database、async runner 串联场景，gate 需要明确 no-skip、真实依赖、失败诊断和 repeated-run 约束。
- 本地权限阻塞不需要仓库改动。
  - **类别**：无需仓库改动
  - **说明**：权限恢复后同一命令链可继续执行，未发现需要修改 repo 脚本或文档的稳定缺陷。

## 4 对流程资产的改进建议

- 为 `/plan-code-review` 增加 lifecycle / async runner 类计划的生产入口反查规则。
  - **落点**：`.agent-skills/plan-code-review/SKILL.md`
  - **优先级**：high
  - **建议**：当计划声明 worker、dispatcher、outbox、scheduler、bootstrap、shutdown 等 runtime 能力时，必须反查 `cmd/api` 或实际生产 entrypoint，并要求至少一个 production-wiring test 覆盖启动链路。
- 将 live dependency BDD 的 repeatability 固化为场景 gate。
  - **落点**：`test/scenarios/README.md` 或对应 suite README / `/scenario-run`
  - **优先级**：medium
  - **建议**：依赖 Postgres、MinIO、Redis、Kind 或外部 provider 的场景必须记录环境预检、禁止静默 skip、失败诊断字段和重跑清理策略。
- 保留 owner plan 的 BDD matrix 区分脚本场景与 Go HTTP BDD。
  - **落点**：`docs/spec/backend-async-runner/plans/001-internal-job-outbox-runner/plan.md`
  - **优先级**：low
  - **建议**：继续使用 operation matrix 明确每个 operation 的 scenario 编号、consumer、handler、persistence 与 AI dependency，以便后续 L2 review 快速核验范围。

## 5 建议优先级与后续动作

- 下一轮最值得实施的是 `/plan-code-review` 的 production entrypoint reverse-audit 规则，因为它直接防止完成态计划遗漏真实启动链路。
- 可随后补充 scenario live dependency repeatability 规则，优先覆盖 P0.033 这类 object store + database + async runner 串联场景。
- 本次权限阻塞仅记录为环境事实，除非后续在同一仓库重复出现，否则不建议为它新增仓库规则。
