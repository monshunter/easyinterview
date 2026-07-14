# OPENAPI-005 · Resume list summary projection

> **ID**: OPENAPI-005
> **状态**: accepted
> **日期**: 2026-07-14
> **版本**: 1.1

## 1 背景

当前 `GET /api/v1/resumes` 的 `PaginatedResume.items` 直接复用完整 `Resume`，导致列表响应携带 `fileObjectId`、正文快照、解析摘要、结构化档案和 provenance 等只属于详情页的数据。Home 简历选择器、Resume Workshop 列表和其它列表消费者只需要稳定的身份、展示与可读性投影；让这些消费者下载并理解完整详情既扩大传输和隐私面，也迫使前端从详情字段重复推断“是否已有可读内容”。

项目尚未上线。用户于 2026-07-14 明确授权以 breaking correction 原地修订 v1.0.0 freeze：列表返回最小 closed summary，详情仍由既有 `getResume` 提供；全部消费者同批迁移，不保留兼容层。

## 2 决策

- 保持 `GET /api/v1/resumes`、operationId `listResumes`、`200 + PaginatedResume` 与 cursor/pageInfo 外层分页合同不变。
- `PaginatedResume.items` 从 `Resume` 改为 closed `ResumeSummary`。`ResumeSummary.additionalProperties=false`，required property set 精确为：
  - `id`: UUID；
  - `title`: string；
  - `displayName`: string；
  - `language`: BCP 47 string；
  - `sourceType`: `upload | paste`；
  - `parseStatus`: `TargetJobParseStatus`；
  - `summaryHeadline`: required nullable string，按 `parsed_summary.headline` → `parsed_summary.basics.headline` → `structured_profile.headline` → `structured_profile.basics.headline` 顺序取第一个 trim 后非空 string；均缺失、空白或非 string 时为 `null`；
  - `hasReadableContent`: required boolean；当且仅当 trim 后 `parsed_text_snapshot` 非空、trim 后 `original_text` 非空，或 `structured_profile` 是非空 object 时为 `true`。不得根据 `fileObjectId`、`sourceType` 或 `parseStatus` 猜测，前端也不得从详情字段重新推断；
  - `updatedAt`: RFC3339 date-time。
- `ResumeSummary` 禁止任何未列出的字段，尤其禁止 `fileObjectId`、`originalText`、`parsedTextSnapshot`、`parsedSummary`、`structuredProfile`、`createdAt`、`deletedAt`、`status` 与 provenance。
- 当前 register/create path 始终写入合法 `source_type`。若数据库存在 NULL 或非法 legacy row，list projection 必须按数据完整性错误 fail closed；不得根据 `fileObjectId` 补猜、伪造默认值或静默隐藏该行。本 correction 不为未上线历史形态新增兼容迁移。
- 保持 `GET /api/v1/resumes/{resumeId}`、operationId `getResume`、`200 + Resume` 全详情合同不变；详情页继续只从该 operation 读取详情字段。
- OpenAPI source、fixture/example、Go/TS codegen、backend store/service/handler、mock、Home/Resume Workshop/其它前端消费者必须同批迁移。不得增加旧 shape alias、可选详情字段、第二个列表 endpoint 或 frontend fallback fetch。

## 3 影响

| 边界 | 受影响的项 | Owner |
|------|-----------|-------|
| 契约 | `ResumeSummary`、`PaginatedResume.items`、schema inventory 与 generated types | openapi-v1-contract 001/003/004 |
| Fixtures / Mock | `Resumes/listResumes.json`、fixture validator、examples、Prism/mock parity | openapi-v1-contract 002 + mock-contract-suite |
| 后端 | list projection SQL/record/service mapper/handler；`getResume` detail mapper 保持 | backend-resume |
| 前端 | Home 简历选择器、Resume Workshop 列表及所有 `listResumes` consumers；详情继续消费 `getResume` | frontend-home + frontend-resume-workshop + shared client consumers |
| Regression | register/list、flat list/auth、detail read-only 分离 | backend/frontend owner tests + root `make test` |

## 4 迁移与回滚

- **迁移路径**：先记录本 accepted decision；003 Phase 9 从 merge-base 旧 baseline 对 proposed OpenAPI 生成并 exact-match machine oracle；001/002/004 与全部 backend/frontend/mock/scenario consumer 同批切换；全部 gate 通过后才允许 re-freeze。
- **放行条件**：37 operations/10 tags、list/get method/path/operationId/status、pagination envelope 与完整 `getResume` 保持；list item 只含九个 required 字段；backend 不读取详情列组装列表；frontend 不通过额外详情请求补列表；根 `make test` 与 contract/fixture/codegen gates 通过。
- **回滚**：任一 store projection、generated type、fixture/mock、consumer 或 BDD gate 未同批完成时整体回滚本 correction；不得恢复详情字段作为兼容可选项或由前端 N+1 请求补齐。
- **SemVer**：v1.0.0 尚未发布，本变更作为 accepted pre-release freeze correction 原地 re-freeze；发布后同类 response narrowing 必须使用 major version。

## 5 相关

- [openapi-v1-contract spec](../spec.md) D-37 / C-21
- [001-bootstrap](../plans/001-bootstrap/plan.md) Phase 16
- [002-fixtures-and-mock-source](../plans/002-fixtures-and-mock-source/plan.md) Phase 11
- [003-breaking-change-gate](../plans/003-breaking-change-gate/plan.md) Phase 9
- [004-resume-additive-coverage](../plans/004-resume-additive-coverage/plan.md) Phase 7

## 6 审计

| 项 | 内容 |
|----|------|
| 提议人 | product owner |
| Review | user explicitly authorized the pre-release breaking correction on 2026-07-14 |
| 实施分支 | `fix/request-dedup-resume-summary-routing-0714` |
| expected finding oracle | `decisions/OPENAPI-005-resume-list-summary.expected-findings.json`；仅由 003 Phase 9 在 proposed schema RED/GREEN 时生成，accepted decision 阶段不伪造该 JSON |
| baseline | `openapi/baseline/openapi-v1.0.0.yaml`；consumer gates 全绿后才允许 re-freeze |
| history | `2026-07-14 | 1.59 | OPENAPI-005 Resume list summary projection` |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-14 | 1.1 | 明确 summaryHeadline 与 hasReadableContent 的唯一持久化投影语义，禁止按来源或解析状态猜测。 | OPENAPI-005 |
| 2026-07-14 | 1.0 | 接受 list summary / detail full split、九字段 closed projection 与 no-compatibility 同批迁移。 | OPENAPI-005 |
