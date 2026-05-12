# Backend Upload 001 Security Privacy Hardening 交付复盘报告

> **日期**: 2026-05-12
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`backend-upload/001-file-objects-and-presign-baseline` review follow-up hardening，覆盖真实 upload size enforcement、presign idempotency TTL、runtime `privacy_delete` upload deleter wiring、DB hard delete + audit tombstone atomicity，以及 E2E.P0.033 live gate skip fail-fast。
- Bug 记录：新增 [BUG-0048](../bugs/BUG-0048.md)，记录 size limit bypass、privacy delete runtime gap、expired presign replay、audit tombstone loss 和 BDD skip false pass。
- Spec/plan 证据：`backend-upload` spec/history 升到 1.2；原 plan/checklist/BDD files 原地修订并在 verification 后恢复 `completed`，Phase 6 记录每项 focused gate。
- 通过证据：`go test ./backend/internal/upload/... ./backend/internal/privacy/runner ./backend/cmd/api -count=1`、`cd backend && go test ./...`、`python3 test/scenarios/e2e/p0-033-file-presign-register-roundtrip/scripts/script_contract_test.py`、`make docs-check`、`git diff --check`、`make test`、`make lint`、`make build` 均 PASS。
- 环境说明：本次没有伪造 live E2E.P0.033 PASS；脚本已改为缺少 `DATABASE_URL` / `OBJECT_STORAGE_*` 或出现 integration skip 时失败，真实 DB + MinIO evidence 需要在环境就绪后补跑。

## 2 会话中的主要阻点/痛点

- Presign `byteSize` validation 与上传完成事实之间缺少闭环。
  - **证据**：review 指出 MinIO path 丢弃 `byteSize`，register 只检查 `Exists`；修复前没有 `Stat` contract 或 size mismatch test。
  - **影响**：客户端可上传超过 `upload.maxBytes.*` 的对象，属于安全与资源治理缺口。
- Privacy delete 局部 deleter 没有进入真实 job runtime。
  - **证据**：`cmd/api` 只构建 targetjob drainer 的 target import/source refresh handlers；upload service 局限在 upload route。
  - **影响**：账号删除 API 创建的 job 不会清理 upload blob 和 `file_objects` 行。
- 通用 idempotency 默认 TTL 不适合 presign response。
  - **证据**：signed URL TTL 来自 `upload.presignTTLSeconds`，但 idempotency replay 仍使用 24h 默认值。
  - **影响**：客户端会收到已过期 `uploadUrl`，导致重试路径表面 201 但实际不可用。
- Audit tombstone 被放在 hard delete 之后。
  - **证据**：旧 `DeleteFileObjectsForUser` 先删 DB row，再插 audit tombstone。
  - **影响**：瞬时 audit 写入失败会永久丢失隐私删除 tombstone。
- BDD gate 仍把 skipped live checks 当作可接受输出。
  - **证据**：`go test` skip exit 0，旧 verify 只 grep test names。
  - **影响**：E2E.P0.033 可在没有 live DB / MinIO evidence 时被标记 Ready/PASS。

## 3 根因归类

- Upload completion contract 缺少 object store metadata invariant。
  - **类别**：spec-plan
- Privacy deletion acceptance 只验证 local deleter，没有验证 `DELETE /api/v1/me` job drainer owner path。
  - **类别**：spec-plan
- Idempotency middleware 的默认行为没有被 per-operation TTL 需求覆盖。
  - **类别**：spec-plan
- 隐私删除审计要求没有被明确成 DB transaction gate。
  - **类别**：spec-plan
- Scenario framework 没有自动把 skipped live integration 视为 BDD evidence failure。
  - **类别**：README / spec-plan

## 4 对流程资产的改进建议

- 后续 upload/file intake 类 plan 应把 “declared metadata equals stored object metadata” 写入 BDD 和 store/service focused tests。
  - **落点**：spec-plan
  - **优先级**：high
- 所有 privacy_delete owner plan 应要求 `cmd/api` runtime drainer handler registration test，而不仅是 owner deleter unit test。
  - **落点**：spec-plan
  - **优先级**：high
- 携带 expiring capability 的 API response 应显式审查 idempotency replay TTL，避免默认 24h 缓存过期能力。
  - **落点**：spec-plan
  - **优先级**：high
- 隐私删除中任何 required tombstone / audit marker 应与对应 DB deletion 同事务，或在失败时保留可重试实体。
  - **落点**：spec-plan
  - **优先级**：high
- `test/scenarios/README.md` 可补充通用规则：声明 live evidence 的场景必须 fail on skip，不能仅依赖 `go test` exit code。
  - **落点**：README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：在具备 live Postgres + MinIO env 后，按 `/scenario-run` 或手动 `setup -> trigger -> verify -> cleanup` 补跑 E2E.P0.033，产出真实 `DATABASE_URL` / `OBJECT_STORAGE_*` evidence。
- 下一步交付：执行 `/work-journal` 时使用 commit title `fix(backend-upload): harden upload privacy and live gates`，并在日志中引用 [BUG-0048](../bugs/BUG-0048.md)。
- 可延后：把 “fail on skip for live evidence” 提炼进 `test/scenarios/README.md` 或 scenario template，避免后续 live 场景重复实现脚本级 grep。
