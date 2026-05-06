# Event and Outbox Contract Bootstrap

> **版本**: 1.7
> **状态**: completed
> **更新日期**: 2026-05-05

**关联 Checklist**: [checklist](./checklist.md)
**关联 Spec**: [spec](../../spec.md)

## 1 目标

把 [event-and-outbox-contract spec](../../spec.md) §3.1 锁定的 12 项决策（D-1..D-12）落到代码契约层：建立 `shared/events.yaml` 与 `shared/jobs.yaml` 双真理源、B3-owned 跨语言 generator（`backend/cmd/codegen/events`），生成 Go envelope / payload / jobType 常量（`backend/internal/shared/events/`、`backend/internal/shared/jobs/`）、TS envelope / payload / jobType 常量（`frontend/src/lib/events/`、`frontend/src/lib/jobs/`）以及当前 16 个事件的 JSON Schema 文件（`shared/events/schemas/<eventName>.v1.json`）；落地 `make codegen-events` 与 `make lint-events` 本地 drift gate；通过本 plan Phase 6/8 的本地命令证明 spec [§6 验收标准](../../spec.md#6-验收标准) C-1 / C-2 / C-6 / C-7 / C-10 全量成立，C-3 / C-4 / C-5 / C-8 / C-9 / C-11 由 handoff 文档串到 [B4 db-migrations-baseline](../../../db-migrations-baseline/spec.md)、[C8 backend-async-runtime](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)、[F1 observability-stack](../../../observability-stack/spec.md)、[F2 analytics-funnel](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 与各业务 C 域闭合。

本 plan 不实现 dispatcher 进程（归 C8）、不实现 Asynq handler 注册（归 C8）、不创建 `outbox_events` / `async_jobs` 表（归 B4，本 plan 仅锁定 column 名词典）、不实现业务 producer / consumer（归各 C 域）、不接入 PostHog / 前端埋点（归 F2）、不初始化 OTel SDK（归 F1）。后续如需新增事件 / 升级 `eventVersion` / 新增 canonical `job_type` / 调整 B2 API-facing subset，递增 [event-and-outbox-contract spec](../../spec.md) 版本与 [history](../../history.md)。已完成主题必须原地 reopen 本 plan 做 remediation，不开同主题 sibling plan。

## 2 背景

[engineering-roadmap §5.1](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 将 B3 保留为当前 active Contract spec；[engineering-roadmap §3.2 Q-2](../../../engineering-roadmap/spec.md#32-adr-q1q6-当前约束) 与 [ADR-Q2](../../../engineering-roadmap/decisions/ADR-Q2-async-orchestration.md) 把「公共 API / DB / event / metrics 中 `jobType` 沿用 snake_case，Asynq dotted name 由 C8 / B3 / B4 显式维护映射」锁为硬约束；[ADR-Q1](../../../engineering-roadmap/decisions/ADR-Q1-auth.md) 把 magic link 邮件派发约束到 internal-only `email_dispatch` canonical jobType，并红线掉 raw token / URL / email body / address 入库。本 plan 是这两条约束在契约层的 owner。

执行本 plan 前必须确认 [B1 shared-conventions-codified/001-bootstrap](../../../shared-conventions-codified/plans/001-bootstrap/plan.md) 已 `completed`：B1 generator 输出的 `backend/internal/shared/types/enums.go`、`frontend/src/lib/conventions/{enums,errors,pagination}.ts` 与 `shared/conventions.yaml` 是本 plan 复用的 enum 真理源；payload 字段类型（如 `TargetJobParseStatus` / `PracticeGoal` / `PracticeMode` / `ReadinessTier` / `QuestionReviewStatus` / `ReportStatus` / `InterviewerRole` / `PrivacyRequestType` / `PrivacyRequestStatus`）必须以 alias 引用，不得在本 plan 中复制。若 B1 未完成，先暂停本 plan。

每个 phase 是可独立验证的纵向切片：Phase 1 起来就能验证 `shared/events.yaml` 真理源；Phase 2 起来就能验证 `shared/jobs.yaml` 与 §3.1.2 API-facing subset 一致；Phase 3 起来就能 `make codegen-events` 双端生成；Phase 4 起来就能 `make lint-events` 拦截裸字面量与 dot.case 漂移；Phase 5 起来就能用 Go / TS 单元测试覆盖 envelope / trace 透传 / breaking-change 拦截 / `email_dispatch` 红线；Phase 6 收口 AC 验证 + handoff。

## 3 质量门禁分类

- **Plan 类型**: `tooling + code-internal + contract`。本 plan 修改 `shared/events.yaml` / `shared/jobs.yaml` 契约真理源、B3 codegen、Go / TS generated contract 包、JSON Schema baseline、lint / drift gate 与跨语言单元测试；不引入用户可感知 UI、HTTP API 行为、业务流程或端到端功能。
- **TDD 策略**: 必须通过 `/tdd --file docs/spec/event-and-outbox-contract/plans/001-bootstrap/checklist.md --references docs/spec/event-and-outbox-contract/plans/001-bootstrap/plan.md,docs/spec/event-and-outbox-contract/spec.md --phase-commit event-and-outbox-contract/001-bootstrap` 顺序执行。每个 checklist item 的 `验证:` 子句是 Red-Green-Refactor 断言来源；涉及真理源、generator、lint、baseline 或 generated output 的 item 必须先补 focused unit / lint / drift fixture 形成失败，再最小实现并复跑对应命令。
- **BDD 策略**: BDD 不适用。本 plan 不产生浏览器 UI、外部 API、用户业务流程或场景测试可观察行为，因此不创建 `bdd-plan.md` / `bdd-checklist.md`，主 checklist 不设置 `BDD-Gate:`。
- **替代验证 gate**: 使用内部契约 gate 代替 BDD：`python3 scripts/lint/events_inventory.py ...`、`make codegen-events`、`make lint-events`、Go generator / shared package tests、TS events / jobs tests、breaking / additive drift fixtures、本地 drift gate、`python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`。

## 4 实施步骤

### Phase 1: `shared/events.yaml` 真理源（envelope + 16 事件 payload）

#### 1.1 envelope schema 与 8 字段定义

在 `shared/events.yaml` 写入 envelope schema：按 [event-and-outbox-contract spec §3.1 D-1](../../spec.md#31-已锁定决策含-jobtype-映射表) 锁定的 8 个字段（`eventId` UUIDv7 / `eventName` dot.case / `eventVersion` int 起始 1 / `aggregateType` snake_case / `aggregateId` UUIDv7 / `occurredAt` RFC3339 / `producer` enum（`api` / `backend_async` / `dispatcher` / `review`）/ `traceId` optional, soft-required / `payload` polymorphic per eventName）写入 `envelope:` 节。`eventId` / `aggregateId` 字段类型必须通过 `$ref`-style alias 复用 [B1 shared/conventions.yaml](../../../shared-conventions-codified/spec.md#31-已锁定决策) 的 UUIDv7 校验语义；envelope 不重复声明 UUIDv7 正则。`producer` 字段在 yaml 中声明为 event-local enum，generator 输出常量到 `events.Producer*`。

#### 1.2 16 个事件 payload 字段清单

按 [spec §3.1.3](../../spec.md#313-16-个事件全集v1) 与 [spec §3.1.4](../../spec.md#314-v1-payload-schema-inventory) 逐行写入 16 个事件的 `events:` 列表。每个事件落地以下属性：`name`（dot.case，2 或 3 段，最后一段过去式）、`version: 1`、`producer`（来自 §3.1.3 producer 列）、`aggregateType`（snake_case）、`requiredPayload`（map，键为字段名，值为 `{type, source}`）、`optionalPayload`（v1 全部为空，留 schema slot 以便后续 additive 扩展）、`piiBoundary`（描述串，对应 §3.1.4 PII 列）。`requiredPayload` 中复用 B1 enum 的字段必须通过 `$ref`-style alias 引用，例如 `analysisStatus: { type: $ref:b1.TargetJobParseStatus }`；不得在 yaml 中复制 enum 字面量。

#### 1.3 event-local enum 与 B1 边界

`shared/events.yaml` 中 B1 未覆盖、但需要有界值的字段（`TargetImportSourceType` 取值 `url` / `text` / `file`，对应 `target.import.requested.sourceType`；`ResumeTailorMode` 取值 `inline` / `rewrite` / `mirror`，对应 `resume.tailor.completed.mode`；`SourceFreshnessStatus` 取值 `fresh` / `stale` / `failed`，对应 `source.refreshed.freshnessStatus`）必须在 yaml 顶层 `eventLocalEnums:` 节声明，并在字段描述里写明「不进入 B1 / B2 公共 enum，新增成员通过本 spec 修订」。Phase 4 lint 必须拒绝 `eventLocalEnums:` 中的成员被偷偷搬进 B1 `shared/conventions.yaml`：通过对比两份 yaml 的 enum 名称白名单，发现重名即 fail。

#### 1.4 事件清单完整性自检

写一段轻量校验脚本（可放 `scripts/lint/events_inventory.py` 或等价 `make` target 内联），断言：`events:` 列表长度 == 16；每个 `name` 唯一且匹配 dot.case 正则；最后一段动词来自 white-listed 过去式集合（`requested` / `parsed` / `started` / `completed` / `generated` / `failed` / `created` / `changed` / `refreshed`）；前缀 domain 来自固定 7 个（`target` / `practice` / `report` / `resume` / `debrief` / `source` / `privacy`）；§3.1.4 的 required field 集合与 yaml 完全一致；`producer` 与 §3.1.3 一致。任一漂移即 fail，确保 spec ↔ yaml 不对账漂移。

### Phase 2: `shared/jobs.yaml` 真理源（10 canonical jobType + `email_dispatch` 红线）

#### 2.1 10 个 canonical job_type 与 dotted name 映射

在 `shared/jobs.yaml` 写入 `jobs:` 列表，按 [spec §3.1.1](../../spec.md#311-dbbackend-runner-canonical-job_type--asynq-dotted-task-name-映射) 落地 10 项：`target_import` / `resume_parse` / `report_generate` / `resume_tailor` / `debrief_generate` / `source_refresh` / `embedding_upsert` / `privacy_export` / `privacy_delete` / `email_dispatch`。每项含 `canonical`（snake_case，与 `B4 db-migrations-baseline §5.9 async_jobs.job_type` 一致并扩展）、`asynqTask`（`<domain>.<action>` dotted name，与 ADR-Q2 §3.2 一致）、`apiFacing`（boolean，B2 OpenAPI subset 标记）、`triggerEvent`（来自 §3.1.1 触发事件列）、`ownerDomain`（C 域 ID）、`priority`（`critical` / `default` / `low`，按 ADR-Q2 §3.5 分类）。新增 canonical job 必须先递增本 spec / plan，再同步 [B4 async_jobs.job_type check constraint](../../../db-migrations-baseline/spec.md) 与 backend runner registry，本 plan 在 Phase 6 handoff 章节中明示该流程。

#### 2.2 B2 API-facing JobType subset 锁定

按 [spec §3.1.2](../../spec.md#312-b2-api-facing-jobtype-subset) 在 `shared/jobs.yaml` 顶层声明 `apiFacingSubset:` 列表，固定 7 项：`target_import` / `resume_parse` / `report_generate` / `resume_tailor` / `debrief_generate` / `privacy_export` / `privacy_delete`。`source_refresh` / `embedding_upsert` / `email_dispatch` 标记为 `apiFacing: false` 且不进入 subset；任一变更必须先额外修订 [B2 openapi-v1-contract spec](../../../openapi-v1-contract/spec.md)。Phase 4 lint 必须拒绝 subset 长度 != 7 或 subset 任一项 `apiFacing != true`，避免有人把 `email_dispatch` 偷偷加进 fixture / OpenAPI。

#### 2.3 `email_dispatch` payload 红线 schema

在 `shared/jobs.yaml` 中 `email_dispatch` 节增加 `payloadSchema:` 子节：仅允许 `authChallengeId` (uuidv7) / `userId` (uuidv7) / `templateKey` (controlled slug) / `locale` (BCP-47) / `deliverySecretRef` (opaque ref string, C1-owned) / `dedupeKey` (string)。同时声明 `redactedFields:` 列表（即 `rawMagicLinkToken` / `magicLinkUrl` / `recipientEmail` / `recipientEmailHash` / `emailBody` / `emailSubject` 等）作为「禁止字段」白名单；generator 把这份 redacted 列表写到 `backend/internal/shared/jobs/email_dispatch.go` 与 `frontend/src/lib/jobs/emailDispatch.ts` 的常量与 lint 钩子里。Phase 5 单元测试在 producer 层强制：构造包含任一 redacted 字段的 payload 必须在编码前被拒绝。

#### 2.4 dotted name typo 自检

写一段轻量校验脚本（同 1.4 风格），断言每条 `asynqTask` 与 `canonical` 之间的映射关系（`canonical = "target_import"` ↔ `asynqTask = "target.import"` 等）严格遵循 [spec §3.1.1](../../spec.md#311-dbbackend-runner-canonical-job_type--asynq-dotted-task-name-映射) 表格，避免 dotted name 笔误（如 `target_import` 写成 `target.imports`）。任一不匹配即 fail。

### Phase 3: B3 generator 与 Go / TS / JSON Schema 输出

#### 3.1 B3-owned generator 程序

在 `backend/cmd/codegen/events/main.go` 落地 B3 generator。**严格规则**：generator 是一个**独立的 Go 程序**，与 [B1 backend/cmd/codegen/conventions](../../../shared-conventions-codified/plans/001-bootstrap/plan.md#13-写入-generator) **不合并**进同一个 binary；本 plan 在 root `Makefile` 中声明 `codegen-events` target 调用 `go run ./backend/cmd/codegen/events`。该 generator 通过 import 引用 [B1 generated shared types](../../../shared-conventions-codified/spec.md#31-已锁定决策) 完成 alias 解析，但不修改 B1 generator 源码、不改写 B1 输出文件。

#### 3.2 Go 端输出

generator 渲染：`backend/internal/shared/events/envelope.go`（envelope struct，含 `EventID` / `EventName` / `EventVersion` / `AggregateType` / `AggregateID` / `OccurredAt time.Time` / `Producer Producer` / `TraceID *string` / `Payload json.RawMessage`，并提供 `MarshalJSON` / `UnmarshalJSON` 保证 wire-format 与 yaml 一致）、`backend/internal/shared/events/events.go`（16 个事件 `EventName*` 字符串常量 + 16 个 `<EventName>Payload` struct + JSON tag）、`backend/internal/shared/jobs/jobs.go`（10 个 canonical `JobType*` 常量 + 10 个 `AsynqTask*` 常量 + 7 项 API-facing subset slice + `email_dispatch` redaction policy 常量）。所有 enum-typed payload 字段通过 type alias 引用 B1（如 `Status enums.TargetJobParseStatus`）。

#### 3.3 TS 端输出

同一 generator 二进制额外渲染 TS：`frontend/src/lib/events/envelope.ts`（envelope discriminated union + `EventEnvelope<T>` generic）、`frontend/src/lib/events/events.ts`（16 个 `EVENT_NAME_*` 字面量常量 + 16 个 payload type alias + `EventNameToPayload` map）、`frontend/src/lib/jobs/jobs.ts`（10 个 `JOB_TYPE_*` 字面量 + 10 个 `ASYNQ_TASK_*` 字面量 + `API_FACING_JOB_TYPES` readonly tuple + `email_dispatch` redaction policy export）。B1 TS enum / type 通过 import alias 引用，不复制。

#### 3.4 JSON Schema 输出

generator 同步在 `shared/events/schemas/<eventName>.v1.json` 输出每个事件的 JSON Schema（Draft 2020-12），结构为「envelope schema + `payload` 字段引用 16 个事件各自的 payload schema」。`payload` 中复用 B1 enum 的字段由 B3 在生成时 inline 当前 B1 enum 值，或引用 B3-owned `shared/events/refs/<EnumName>.json` 桥接文件；本 plan 不要求 B1 先产出 `shared/conventions/json-schema/<EnumName>.json`。所有 16 个 schema 文件与 B3 refs 在重新生成时必须 idempotent（`git diff --exit-code` clean）。

#### 3.5 committed baseline manifests

在 `shared/events/baseline/events.v1.json` 与 `shared/jobs/baseline/jobs.v1.json` 写入当前 v1 freeze manifest。baseline 由 generator 从 `shared/events.yaml` / `shared/jobs.yaml` 生成并提交，不手写；`make lint-events` 比较当前真理源与 baseline，删除 eventName/jobType、修改 required 字段、修改字段类型、移除 enum 成员、把 internal-only job 标为 API-facing 都必须失败并提示需要 spec 修订 / eventVersion+1。

#### 3.5 generator idempotency 与 Make 入口

根 `Makefile` 在 [B1 codegen-conventions](../../../shared-conventions-codified/plans/001-bootstrap/plan.md#13-写入-generator) 与 [B2 codegen-openapi](../../../openapi-v1-contract/plans/001-bootstrap/plan.md#23-make-入口) 之后追加 `codegen-events` target；执行顺序为 `codegen-conventions` → `codegen-events` → `codegen-openapi`（B2 与 B3 互不依赖，但 B3 必须排在 B1 之后）。把 root `make codegen` 占位由 `codegen: codegen-conventions codegen-openapi` 升级为 `codegen: codegen-conventions codegen-events codegen-openapi`。`make help` 自动收录新 target。

#### 3.7 L2 remediation: B1 enum refs 不得在 B3 generator 内硬编码

B3 generator 生成 `shared/events/refs/<EnumName>.json` 时必须从 `shared/conventions.yaml` 读取 `$ref:b1.*` enum 当前值；B3 源码不得维护 B1 enum 字面量副本。若 `shared/events.yaml` 引用了 `shared/conventions.yaml` 中不存在的 B1 enum，generator 必须 fail-fast，避免 JSON Schema ref 静默过期。

### Phase 4: `make lint-events` 与本地 drift gate

#### 4.1 裸字面量 lint

落地 `scripts/lint/lint_events.py`（或等价 Go 工具）作为 `make lint-events` 实体：扫描 `backend/` 与 `frontend/` 下除 `backend/internal/shared/{events,jobs}` 与 `frontend/src/lib/{events,jobs}` 之外的所有源文件，拒绝出现 `"target.parsed"` / `"report.generated"` / `"report_generate"` / `"target.import"` 等 16 事件名 + 10 canonical jobType + 10 dotted task name 的裸字面量；命中即 fail，提示「请 import 常量包」。lint 必须以 dry-run 模式可在 `make lint-events` 中独立调用，不强依赖编辑器。

#### 4.2 16 事件 frozen list 强制

lint 同时校验 `backend/internal/shared/events/events.go` 与 `frontend/src/lib/events/events.ts` 中导出的事件名集合长度 == 16 且与 `shared/events.yaml` 一一对应；新增 / 删除 / 重命名事件必须先递增 spec 版本与 history。任何在 generated 文件之外手写的 `EventName*` 常量声明会被 lint 拒绝。

#### 4.3 B2 API-facing subset 不可变

lint 校验 `backend/internal/shared/jobs/jobs.go` 中 `APIFacingJobTypes` 长度 == 7 且与 `shared/jobs.yaml.apiFacingSubset` 一致；任一项 `apiFacing != true` 即 fail；命中说明有人在 jobs.yaml 中把 `email_dispatch` / `source_refresh` / `embedding_upsert` 偷偷标 true（参见 5 风险与应对的「subset 误扩」条目）。

#### 4.4 本地 drift gate

落地 `make codegen-events`（生成 + 不干别的）与组合 gate `make lint-events`（前述 4.1-4.3 全部跑）。本地完整 drift check 命令只检查 generator 真正写入的 artifact：`make codegen-events && make lint-events && git diff --exit-code -- shared/events.yaml shared/jobs.yaml backend/internal/shared/events/{envelope.go,events.go} backend/internal/shared/jobs/jobs.go frontend/src/lib/events/{envelope.ts,events.ts} frontend/src/lib/jobs/jobs.ts shared/events/{schemas,refs,baseline} shared/jobs/baseline`。手写 `*_test.*` 与 fixtures 由 `make lint-events` / Go / TS 单测覆盖，不作为 generated drift 路径。远端 CI required check 仅在 [A5 ci-pipeline-baseline](../../../ci-pipeline-baseline/spec.md) 触发条件成立后再接入；本 plan 不修改 A5 workflow，仅在 `Makefile` 注释中说明「当前是本地唯一强制 gate」。

#### 4.5 L2 remediation: 缺失 generated contract 文件必须 fail-fast

`make lint-events` 必须把 `backend/internal/shared/events/events.go`、`frontend/src/lib/events/events.ts`、`backend/internal/shared/jobs/jobs.go`、`frontend/src/lib/jobs/jobs.ts` 视为必需 generated contract 文件；任一文件缺失时应失败并提示运行 `make codegen-events`，不得跳过缺失路径导致 generated set 校验 fail-open。

### Phase 5: 单元测试（envelope / trace 透传 / breaking-change 拦截 / `email_dispatch` 红线）

#### 5.1 envelope round-trip 测试

在 `backend/internal/shared/events/envelope_test.go` 与 `frontend/src/lib/events/envelope.test.ts` 落地 round-trip 测试：构造 16 个事件中至少 3 个（覆盖不同 producer：`api` 的 `target.import.requested`、`backend_async` 的 `report.generated`、`review` 的 `debrief.completed`）的合法 envelope，序列化为 JSON 后再反序列化，断言所有字段（含 `eventId` UUIDv7、`occurredAt` RFC3339、`producer` enum、`payload` 类型）等值。Go 与 TS 两端必须使用同一份 fixture（可放 `shared/events/__fixtures__/`）确保 wire-format 跨语言一致。

#### 5.2 `traceId` soft-required 双码路测试

测试 `traceId` 既可能存在（producer 从 W3C `traceparent` / active span 派生）也可能缺失（producer 没有可用 trace context）：缺失分支 producer 必须仅落 warn log（而非 panic / error）并允许 dispatcher publish；存在分支 envelope 必须按字面量透传到 outbox row，consumer 端可解析回原值。两端各写至少 2 个 case：一个 `traceId == nil` warn-log + publish 通过，一个 `traceId == "00-..."` 透传无变换。F1 backend background runner span 重建逻辑由 [F1 observability-stack](../../../observability-stack/spec.md) 后续 plan 验收，本 plan 只覆盖契约层透传语义。

#### 5.3 breaking-change 拦截测试

在 generator 单元测试目录（如 `backend/cmd/codegen/events/breaking_test.go`）写若干 negative case fixture：拷贝 `shared/events.yaml` 到内存，分别变更：(a) 把 `report.generated` 的 `questionIssueCount` 类型从 `int` 改 `string`；(b) 删除 `target.parsed` 的 `requirementCount` required 字段；(c) 把 `report.generated` 重命名为 `report.generation_completed`（snake segment）；(d) 把 `target.import.requested` 的 `sourceType` enum 移除 `url` 成员。每个 fixture 必须让 lint / generator / `make lint-events` 在不同步骤失败，错误信息包含 `breaking change requires eventVersion + 1`。同样在 TS 端落地等价 negative test，以保证 generator 双端一致拦截。

#### 5.4 additive change 通过测试

在同一测试套件中加 positive case：给某个事件新增 optional payload 字段（v1 全部 required 完毕，再加 optional 是 additive 唯一允许的形态），断言 generator + lint 全部通过；并断言生成的 Go struct 中该字段为 pointer / TS 中该字段为 `?:` optional，与 v1 既有 required 字段互不冲突。该测试同时覆盖 [spec §4.2](../../spec.md#42-schema-约束)「v1 全部 required，后续 additive 只允许新增 optional」规则。

#### 5.5 `email_dispatch` 红线测试

在 `backend/internal/shared/jobs/email_dispatch_test.go` 与 `frontend/src/lib/jobs/emailDispatch.test.ts` 写「红线 payload 拒绝」测试：构造包含 `rawMagicLinkToken` / `magicLinkUrl` / `recipientEmail` / `emailBody` 任一字段的 payload，提交给本 plan 提供的 `BuildEmailDispatchPayload` helper，断言返回 error 或抛 exception；构造仅包含 `authChallengeId` / `userId` / `templateKey` / `locale` / `deliverySecretRef` / `dedupeKey` 的合法 payload，断言 helper 通过。同时写 lint case：在 fixture yaml 中尝试给 `email_dispatch.payloadSchema` 偷偷加 `recipientEmail` 字段，必须被 Phase 4 lint 立刻拒绝。

### Phase 6: Verification + handoff

#### 6.1 Spec C-1 / C-2 / C-6 / C-7 自检

`make codegen-events` 跑两次后 `git status` 必须 clean；删除 `backend/internal/shared/events/events.go` 与 `frontend/src/lib/events/events.ts` 与任一 `shared/events/schemas/*.json` 后再跑可还原（C-1）。生成产物中 16 个事件常量、payload struct / type 与 §3.1.4 字段清单逐字段一致（C-1 续）。`backend/internal/shared/jobs/jobs.go` 包含 10 个 `JobType*` 常量 + 10 个 `AsynqTask*` 常量；`source_refresh` / `embedding_upsert` / `email_dispatch` 三个常量在 Go / TS 两端被标记 internal-only 且不进入 `APIFacingJobTypes`（C-2）。临时变更 `report.generated` 的 `questionIssueCount` 类型并跑 `make lint-events`：lint 失败、提示 `breaking change requires eventVersion + 1`，revert 后恢复（C-6）。`jobs.AsynqTaskTargetImport == "target.import"`、`jobs.AsynqTaskPrivacyDelete == "privacy.delete"`、`jobs.AsynqTaskEmailDispatch == "email.dispatch"`（C-7）。

#### 6.2 Spec C-10 自检

按 [spec §6 C-10](../../spec.md#6-验收标准) 在仓库根 grep `target.import.requested` / `target_import_requested` / `report.generated` / `report_generated` 四种命名空间，断言：`target.import.requested` 仅出现在 `shared/events.yaml`、`backend/internal/shared/events/events.go`、`frontend/src/lib/events/events.ts`、`shared/events/schemas/target.import.requested.v1.json`、单元测试 fixture 与文档；`target_import_requested` snake_case 在本 plan 输出中**完全不出现**（属于 [F2 analytics-funnel](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec) 的产品分析事件命名空间，由 F2 spec 在自身 plan 注册）。即 16 个 internal eventName（dot.case）与 F2 产品分析事件名（snake_case）属于互不冲突的两个命名空间。

#### 6.3 handoff 文档与下游约束

把以下下游约束写入本 plan 末尾的 handoff 章节，并在 spec / plan 修订日志中链接；不新建独立 `handoff.md`，避免 handoff 约束漂在 plan 生命周期之外：

- [B4 db-migrations-baseline](../../../db-migrations-baseline/spec.md)：`outbox_events` 表必须落地以下 operational columns（B3 owns 列名，B4 owns DDL）：`publish_attempts integer not null default 0`、`next_attempt_at timestamptz not null default now()`、`locked_at timestamptz`、`last_error_code text`、`last_error_message text`，以及 `(publish_status, next_attempt_at, created_at)` 复合索引（dispatcher pending + due 查询走该索引）。`async_jobs.job_type` check constraint 必须包含 10 个 canonical 值（含 `email_dispatch`）。本 plan **不写 SQL**。
- [C8 backend-async-runtime](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)：dispatcher 必须实现 [spec §4.3](../../spec.md#43-outbox-协议约束) 的 `SELECT ... FOR UPDATE SKIP LOCKED` + `(publish_status='pending' and next_attempt_at <= now())` + `next_attempt_at asc, created_at asc` + 批量 ≤ 100；at-least-once 投递；指数退避 30s/2m/10m/1h/6h，max 5 attempts；C-3 / C-4 / C-5 / C-8 / C-11 由 C8 与各 C 域闭合。
- [F1 observability-stack](../../../observability-stack/spec.md)：必产 `outbox_events_pending` Gauge / `outbox_publish_duration_seconds` Histogram / `outbox_publish_failures_total` Counter；积压 > 100 触发告警（spec D-9 / C-9）。
- [F2 analytics-funnel](../../../engineering-roadmap/spec.md#51-当前已存在的-active-spec)：F2 PostHog 产品分析事件名严禁与本 plan 16 个 internal eventName 重名；命名空间分离由 C-10 守住。
- [C1 backend-auth](../../../engineering-roadmap/spec.md#52-当前-p0-实施-workstream-候选)：`email_dispatch` payload 红线（D-12）由 C1 enforce；C1 实现时必须通过本 plan 提供的 `BuildEmailDispatchPayload` helper 而非自行构造 raw map。

#### 6.4 文档与 INDEX 同步

本 plan checklist 全部勾选；Phase 6 关键命令日志贴入工作日志。本 plan 自身 checklist 与 Phase 6 验证完成后，把本 plan / checklist Header 从 `active` 切到 `completed`，并用 `/sync-doc-index --fix-index` 同步 [event-and-outbox-contract/plans/INDEX.md](../INDEX.md)；调用 `/sync-doc-index --check` 确认 [docs/spec/INDEX.md](../../../INDEX.md) 与 plans/INDEX 与 spec / plan Header 无 drift。不修改 [engineering-roadmap/001-decompose-subspecs](../../../engineering-roadmap/plans/001-decompose-subspecs/checklist.md) 已完成的 roadmap checklist；implementation 准入 gate 由后续 C 域 plan 在自身 verification phase 关闭，不重复登记父 roadmap。

### Phase 8: product-scope v1.2 event remediation

#### 8.1 Red: 事件 inventory 期望切到 16 项

先调整 `scripts/lint/events_inventory.py` / `scripts/lint/lint_events.py` 与 focused tests 的期望：事件数 18→16、domain 8→7，旧 `mistake.created` / `mistake.status.changed`、`mistakeCount`、`generatedMistakeCount` 仍存在时必须失败。

#### 8.2 Green: 更新事件真理源、baseline 与生成物

修订 `shared/events.yaml`、`shared/events/baseline/events.v1.json`、JSON Schema、Go / TS generated events：删除独立 `mistake` domain 事件；`report.generated.mistakeCount` 改为 `questionIssueCount`；`debrief.completed.generatedMistakeCount` 改为 `practiceFocusCount`；round-trip fixtures 改用保留事件覆盖 `api` / `worker` producer。

#### 8.3 Verify

运行 `make codegen-events`、`make lint-events`、Go / TS events tests；repo 搜索确认实现侧不再出现旧 mistake event、`MistakeStatus` event refs、`mistakeCount` 或 `generatedMistakeCount`。

## 5 验收标准

- spec [§6 验收标准](../../spec.md#6-验收标准) C-1 / C-2 / C-6 / C-7 / C-10 全量成立；C-3 / C-4 / C-5 / C-8 / C-9 / C-11 通过 Phase 6.3 handoff 文档串到下游 owner（B4 / C8 / F1 / 各 C 域），由各下游 plan 自身 verification phase 关闭。
- 本 plan checklist 全部勾选；Phase 6 关键命令日志贴入工作日志。
- B1 共享类型变更通过 alias / `$ref` 自动同步进 `shared/events.yaml` 与 generated 类型，不形成手写副本（spec §4.2）。
- `make codegen-events` 生成产物 idempotent；删除任意一个 generated 文件后再跑可还原；`make lint-events` 拦截 §4.4、§5.3 列举的 4 类 breaking change 与裸字面量。
- `email_dispatch` payload schema 不允许 `rawMagicLinkToken` / `magicLinkUrl` / `recipientEmail` / `emailBody` 等红线字段（D-12），单元测试覆盖 producer / lint 两端。
- B2 API-facing JobType subset 长度严格为 7（§3.1.2），`source_refresh` / `embedding_upsert` / `email_dispatch` 不进入 subset，lint 拒绝 subset 误扩。

## 6 风险与应对

| 风险 | 应对措施 |
|------|----------|
| 业务包（C 域 handler / consumer）出现裸字面量 `"target.parsed"` / `"report_generate"` 导致 dotted name typo 隐藏漂移 | Phase 4.1 `make lint-events` 扫描 `backend/` 与 `frontend/` 全量源（白名单仅 `backend/internal/shared/{events,jobs}` 与 `frontend/src/lib/{events,jobs}`），命中即 fail；本地 drift gate 是当前唯一强制 gate；C 域 PR 在本地需先跑 `make lint-events` 才能合并 |
| `shared/events.yaml` event-local enum（如 `TargetImportSourceType`）被偷偷搬进 [B1 shared/conventions.yaml](../../../shared-conventions-codified/spec.md) 形成跨 spec 漂移 | Phase 1.3 lint 维护两份 yaml 的 enum 名称白名单，发现重名 fail；任何提升必须先递增 B1 spec 版本与 history，再同步本 plan，禁止单边搬运 |
| `shared/jobs.yaml` 中 `asynqTask` dotted name 笔误（如 `target.imports` / `report.generates`）导致 C8 注册不到 task | Phase 2.4 写 mapping 校验脚本严格对 §3.1.1 表格；新增 canonical 必须先递增 spec 与 plan 再同步 yaml；generator 在编译期生成常量，业务包通过常量引用避免笔误 |
| `traceId` soft-required 在单元测试中误判：缺失时也产 error，破坏 dispatcher publish 路径 | Phase 5.2 强制双码路：缺失分支断言只 warn log + publish 通过，不返回 error；存在分支断言透传字面量；CI / 本地 lint 同时跑两类 test，避免单边过严或过松 |
| B2 API-facing JobType subset 被误扩（有人在 fixture 或 yaml 中把 `email_dispatch` / `source_refresh` 标 `apiFacing: true` 偷偷加进 OpenAPI） | Phase 2.2 + Phase 4.3 lint 双层拦：yaml 层 subset 长度 != 7 即 fail；generated `APIFacingJobTypes` 与 `JobType*Internal` 互斥校验；任一变更必须先递增 B2 spec 而非本 plan |
| Generator idempotency 被 IDE auto-format / import sorter 破坏，造成 `make codegen-events` 反复出现 diff | Phase 3 generator 输出固定 import 顺序、行尾、缩进、build tag；与 `.editorconfig` 对齐；首次 idempotent baseline 由本 plan 锁定；编辑器若自动改动，必须修 generator 模板而非放弃 gate |
| `email_dispatch` 红线被 C1 实现时绕过（手写 map[string]any 直接塞 `recipientEmail`） | Phase 2.3 + Phase 5.5 双层拦：generator 产生 `BuildEmailDispatchPayload` strongly-typed helper，C1 必须通过该 helper 构造；handoff 文档明示「禁止 raw map」；红线字段写入 `redactedFields` 常量供 C1 自检 |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-05-05 | 1.7 | A5 deep review remediation：将 B3 generated drift gate 收窄到 generator 实际输出文件，避免手写 tests / fixtures 被 `codegen-check` 误判为 generated drift。 | historical-spec-implementation-review |
| 2026-05-03 | 1.6 | 刷新计划主体口径：事件全集、generator、lint、round-trip 与 breaking-change 示例统一改为当前 16 事件；旧 `MistakeStatus` / mistake event 只保留在 Phase 8 remediation 说明中。 | product-scope v1.2 / event-and-outbox-contract v1.5 |
| 2026-05-03 | 1.5 | 原地 reopen，新增 Phase 8 remediation：按 product-scope v1.2 删除独立 mistake event domain，报告和复盘事件字段改为题目问题 / 复盘练习焦点计数。 | product-scope v1.2 / event-and-outbox-contract v1.5 |
| 2026-04-30 | 1.4 | 收口 L2 code-review remediation：B3 generator 改为从 `shared/conventions.yaml` 读取 B1 enum refs；`lint-events` 对缺失 generated contract 文件 fail-fast。 | plan-code-review remediation |
| 2026-04-30 | 1.3 | 完成 B3 event/job contract bootstrap implementation；收口 codegen / lint / drift / Go / TS contract tests 与 Phase 6 handoff verification。 | implementation close-out |
| 2026-04-30 | 1.2 | 补齐 TDD/BDD 质量门禁分类与 checklist 可执行验证断言；确认 BDD 不适用并以内部契约 gate 替代。 | implement gate remediation |
| 2026-04-29 | 1.1 | 收口 plan-review：新增 committed baseline manifests；JSON Schema enum refs 改为 B3-owned inline/refs；handoff 固定在 plan 章节内；修正 B4 db-migrations-baseline 相对链接。 | plan-review remediation |
