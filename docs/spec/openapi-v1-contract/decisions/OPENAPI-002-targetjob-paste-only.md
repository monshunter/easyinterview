# OPENAPI-002 · TargetJob paste-only intake

> **ID**: OPENAPI-002
> **状态**: accepted
> **日期**: 2026-07-13
> **版本**: 1.2

## 1 背景

Home 当前同时暴露 JD 文本、URL、文件和手工表单四类导入形态，但用户已于 2026-07-13 明确要求首页只保留直接粘贴 JD 文本，并删除前后端的 JD 文件上传与 URL 导入能力。项目尚未上线，不需要保留旧 request union、旧 response provenance 字段或兼容分支。

现有 file 路径只登记上传对象，TargetJob parser 并未读取文件正文；URL 路径还引入远程抓取、SSRF 防护、内容可用性和刷新语义。继续维护四种 source 会让 UI、OpenAPI、fixture、generated DTO、backend runner 与场景资产同时承担无用户价值的分支，违背当前产品范围与最小完备实现原则。

## 2 决策

采用一次未上线 `v1.0.0 pre-release freeze correction`。所有 producer/consumer 同批迁移，不保留 source wrapper、兼容 alias、discriminator 或旧字段：

1. `POST /api/v1/targets/import`、operationId `importTargetJob`、`202 + TargetJobWithJob`、`Idempotency-Key` 语义保持不变；当前 inventory 仍为 37 operation / 10 tag。
2. `ImportTargetJobRequest` 改为 closed flattened object，且只包含 required `rawText`、`targetLanguage`、`resumeId`。`rawText` 必须同时声明 `minLength: 1` 与 `pattern: '\S'`，从 schema 层拒绝空字符串和纯空白文本。删除 `source`、`titleHint`、`companyNameHint`；标题、公司和结构化要求由服务端从 JD 文本解析，不接受第二套客户端事实。
3. 删除 `TargetJobImportSourceURL`、`TargetJobImportSourceManualText`、`TargetJobImportSourceFile`、`TargetJobImportSourceManualForm` 和 `TargetJobImportSource` schema。
4. 删除 `TargetJob.sourceType` 与 `TargetJob.sourceUrl`；TargetJob 当前只有 paste intake，不再向消费者暴露没有分支价值的来源 provenance。
5. 从 `UploadPresignRequest.purpose` 删除 `target_job_attachment`。通用 `createUploadPresign` operation 及 `resume` / `privacy_export` purpose 保留，不能因 TargetJob 收敛误删简历或隐私能力。
6. URL、JD 文件与 `manual_form` 不保留 route、request variant、fixture scenario、generated DTO、frontend consumer、backend branch 或 runtime refresh path。非法旧 shape 由 closed schema fail closed；不增加专用兼容错误码。
7. 从 `ApiErrorCode` 删除只服务 URL / 文件 source 的 `TARGET_IMPORT_SOURCE_INVALID` 与 `TARGET_IMPORT_SOURCE_UNAVAILABLE`；粘贴输入校验继续使用 `VALIDATION_FAILED`，异步导入失败继续使用 `TARGET_IMPORT_FAILED`。这两个 enum removals 各自计入 old-baseline breaking finding，不得静默过滤。

机器可执行的完整 finding oracle 为 [OPENAPI-002-targetjob-paste-only.expected-findings.json](./OPENAPI-002-targetjob-paste-only.expected-findings.json)。003 wrapper 必须从 merge-base 旧 baseline 比对 proposed OpenAPI，并对 `severity + JSON pointer + kind + before + after` 做顺序无关 exact-set 校验；缺 finding、多 finding、severity 漂移或 wildcard 授权均失败。`rawText` 是本次新建的 required property，其初始 `minLength` / `pattern` 约束必须归一到同一个 `required_property_added.after` 签名，不能额外产生第二个 `constraint_added` finding。wrapper RED 必须同时证明未归一化的额外 constraint finding 失败，以及 stale 15-finding oracle 对 actual 17 失败；GREEN 必须精确等于 17 breaking findings，其中包括两个 source-only `ApiErrorCode` enum removals。若底层 diff 仍独立报告未授权 finding，必须停止并修订 oracle/决策，禁止静默过滤或扩大 allowset。

## 3 影响

| 边界 | 受影响的项 | Owner |
|------|-----------|-------|
| 契约 | TargetJob import/request/response schemas、upload purpose、baseline、generated DTO | openapi-v1-contract 001/003 |
| Fixtures | TargetJobs import/read fixtures、Uploads purpose scenarios、prototype mapping、Prism projection | openapi-v1-contract 002 |
| 后端 | import handler/service/store/runner、URL fetch/refresh、JD attachment 分支 | backend-targetjob/001 + backend-upload/001 |
| 前端 | Home JD intake、generated request type、URL/file/manual-form controls 与状态 | frontend-home-job-picks-and-parse/001 |
| Mock/consumer | fixture registry 与 paste-only consumer flow；删除旧 URL/file/manual-form 正向资产 | mock-contract-suite/001 + domain owners |
| Persistence/events | source provenance columns、attachment linkage、URL refresh event/job | database-migrations/001 + backend event owners |

## 4 迁移与回滚

- **迁移顺序**：先提交本 ADR 与 expected-findings oracle并锁定 merge-base，随后 snapshot 旧 `openapi-v1.0.0.yaml`；保持 worktree baseline 字节不变地更新 proposed `openapi/openapi.yaml`，再执行 old-baseline → proposed exact audit 并保存审计 artifact；之后同批更新 fixtures/prototype、Go/TS codegen、frontend/backend/mock/scenario consumers；全部通过后才原地 re-freeze baseline。
- **放行条件**：exact finding audit、37/10 inventory、paste-only request positive/negative schema tests、canonical `validation-blank-raw-text` 422 fixture、fixture validation、prototype sync 两次幂等、Prism byte parity、codegen check、downstream consumer tests、旧能力 positive/runtime zero-reference 与 current-baseline `make openapi-diff` 全部通过。Zero-reference 只审查当前正向 fixture、runtime、generated artifact 与 consumer 可达面；accepted ADR、机器 oracle 和显式 negative test/fixture declaration 可保留旧 token，但不得以整目录豁免掩盖正向引用。
- **回滚**：任一 consumer 未能同批迁移、finding 超出 oracle、旧分支仍可达或场景闭环失败时，整体回滚 OpenAPI/fixtures/codegen/frontend/backend/mock/scenario/baseline；不得单独恢复兼容字段。
- **SemVer**：baseline 尚未发布，因此保持 `v1.0.0` 并作为 accepted pre-release correction 原地 re-freeze；发布后同类变更必须使用 v2.0.0。

## 5 审计

| 项 | 内容 |
|----|------|
| 提议人 | product owner |
| Review | product owner 于 2026-07-13 明确要求 Home 只保留 JD 文本粘贴，并删除前后端 JD 文件上传与 URL 导入 |
| 实施分支 | `fix/interview-turn-ux-0713` |
| base-ref diff evidence | openapi-v1-contract/003 revision phase 在 baseline re-freeze 前保存 exact finding artifact |
| baseline | `openapi/baseline/openapi-v1.0.0.yaml` pre-release correction |

## 6 相关

- [openapi-v1-contract spec](../spec.md) D-16 / D-34 / C-18
- [001-bootstrap](../plans/001-bootstrap/plan.md)
- [002-fixtures-and-mock-source](../plans/002-fixtures-and-mock-source/plan.md)
- [003-breaking-change-gate](../plans/003-breaking-change-gate/plan.md)
- [mock-contract-suite](../../mock-contract-suite/spec.md)

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-14 | 1.2 | Correct the exact boundary from 15 to 17 findings by including both source-only ApiErrorCode removals; retain rawText constraint folding and stale-oracle RED coverage. | O-A / D-34 |
| 2026-07-13 | 1.1 | Fix freeze order, require non-whitespace rawText, lock the 15-finding normalization rule, and scope zero-reference to positive/runtime surfaces. | OPENAPI-002 L1 review |
| 2026-07-13 | 1.0 | Accept TargetJob paste-only pre-release correction and exact breaking finding boundary. | OPENAPI-002 |
