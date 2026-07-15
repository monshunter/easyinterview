# OPENAPI-001 · Grounded direct report semantics

> **ID**: OPENAPI-001
> **状态**: accepted
> **日期**: 2026-07-15
> **版本**: 1.7

## 1 背景

当前 `FeedbackReport` 让模型输出 numeric score，再由后端推导 readiness/status/confidence/action；报告读取又缺少完成时冻结的 resume/round/display context。真实数据已证明，这会把未回答追问解释成能力缺陷或把极少证据解释成高准备度，而且 report deep link 只能依赖可伪造 route identity。

项目尚未上线，用户已于 2026-07-12 明确确认“方案 A”：模型直接输出 summary 与 `dimension code + label + status + confidence`，后端只做冻结、校验、一次 repair 与持久化。继续保留旧 wire shape 或兼容层会增加歧义，不能解决真实性问题。

用户于 2026-07-13 进一步批准方案 A 的最终边界：`ReportNextAction.label` wire/schema `maxLength` 调整为 200 code points，仅作 malformed-output fuse；English `<=24` whitespace words / zh-CN `<=64` Unicode code points 是 semantic/UX gate，正式 UI 不得把 200 当成文案目标。针对 action label 的定向 repair 使用更保守的内部生成目标 English `<=18` words / zh-CN `<=52` code points，为 provider 分词与标点差异保留余量；18/52 不是新的 wire 或 UI 上限。

用户于 2026-07-15 进一步批准报告会话记录的产品边界：一份可查看会话记录只属于一份报告；公共会话列表没有入口价值，应删除 `listPracticeSessions`，并以报告定位的只读 `getReportConversation` 一对一替换。`feedback_reports.session_id` 的既有外键与唯一索引已经表达报告到会话的一对一绑定，无需新增关系表。

## 2 决策

采用一次未上线 `v1.0.0 pre-release freeze correction`，所有 producer/consumer 同批迁移且不保留兼容字段：

### 2.1 Authorized breaking findings（merge-base audit 必须 exact-match）

1. 删除 `DimensionAssessment.dimension`，新增 required `code` 与 `label`。
2. 删除 `ReportHighlight.dimension` / `ReportIssue.dimension`，新增 required `dimensionCode`。
3. 删除无引用 `DimensionResult` schema。
4. 将 `FeedbackReport.retryFocusCompetencyCodes` 重命名为 `retryFocusDimensionCodes`。
5. `CreatePracticePlanRequest` 保持顶层 `type: object` 与显式 superset properties，内部使用 closed conditional `oneOf`：baseline branch 要求 `goal=baseline` 与既有非 focus 参数并明确禁止 `sourceReportId`；derived branch 要求 `goal=retry_current_round|next_round` 与 non-null/nonblank UUID `sourceReportId`，且对象只允许这两个字段。删除 `focusCompetencyCodes`；derived 其余设置/focus 只由服务端投影。不得让 Go/TS codegen 降级为 `any`。

Conditional negative matrix is part of the accepted contract: baseline + any `sourceReportId` fails; each derived goal without, with null, blank or malformed `sourceReportId` fails; each derived goal with any baseline/settings/identity/focus extra fails; minimal `{goal,sourceReportId}` for both derived goals passes.
6. `FeedbackReport` 使用稳定 response envelope：metadata、nullable-until-state-resolved `summary/errorCode/preparednessLevel/provenance`、non-null `context` 与所有数组属性均 required。`ready` 分支要求 summary/preparednessLevel/provenance non-null 且 dimensionAssessments/nextActions non-empty；`failed` 分支要求 errorCode non-null；queued/generating/ready 的 errorCode 必须 null。非 ready payload 数组保持 empty。新增 `ReportContextSnapshot` required 字段（含 `hasNextRound`）。
7. `FeedbackReport`、`ReportContextSnapshot`、`DimensionAssessment`、`ReportHighlight`、`ReportIssue`、`ReportNextAction` 设置 `additionalProperties: false`。
8. 对 summary/code/label/evidence/action/focus arrays 增加当前 spec 锁定的 min/max/pattern；这是 producer contract tightening。

### 2.2 Additive surface

新增 `ReportContextSnapshot`，只暴露完成时冻结的最小投影：`sourcePlanId`、`targetJobTitle`、`targetJobCompany`、`resumeId`、`resumeDisplayName`、`roundId`、`roundSequence`、`roundName`、`roundType`、`language`、`hasNextRound`。不得暴露原始 JD、Resume 正文、structured profile、transcript、plan difficulty/persona/time budget 或内部 message anchors。

Derived plan 不从客户端重传上述 plan settings：retry 复用 source plan 的 persona/difficulty/language/current-round duration；next 复用 persona/difficulty/language，时长改取 frozen canonical successor；target/resume/round/focus 全由 source report 派生。`hasNextRound` 仅用于前端诚实禁用 CTA，服务端仍重新校验。

`FeedbackReport.summary`、`preparednessLevel` 与 `provenance` 在 queued/generating/failed 时为 null，在 ready 时必须 non-null；ready 的 dimensions/actions 必须 non-empty。`errorCode` 只允许 failed non-null，queued/generating/ready 必须 null。`context` 对所有当前 shape report 均存在。`reportId` 是 frontend 唯一 locator。

Generation/judge retry budget不扩展HTTP contract：`FeedbackReport`继续只暴露`queued/generating/ready/failed`及既有nullable payload/error/provenance；OpenAPI、fixtures与generated DTO不得新增`attemptCount`、`retryCount`、`reason`、`scope`或client retry endpoint。Frontend只按服务端状态诚实轮询；49次客户端窗口耗尽不等于服务端failed，也不授权客户端重新触发generation。

B1 同批新增 canonical non-retryable `REPORT_CONTEXT_TOO_LARGE`。B2 `ApiErrorCode` 只从 `shared/conventions.yaml` 同步该 string enum value；oracle 必须把它精确记录为 additive `enum_value_added`，不得放入 OPENAPI-001 breaking allowset。新 `ReportContextSnapshot` schema 及其自身 closure 也按 additive finding 记录；把该 schema 设为 `FeedbackReport.context` required 才是已授权 breaking finding。

### 2.3 A-200 action-label fuse correction

- `ReportNextAction.label` 为 `minLength=1,maxLength=200`。200不是正常文案目标、UX上限或可见性PASS依据。
- OPENAPI-001 old-baseline oracle对应 finding保持 `severity=breaking`、`path=/components/schemas/ReportNextAction/properties/label`、`kind=constraint_added`、`before=none`，`after`必须更新为`minLength=1,maxLength=200`；不得保留旧after值。
- Runtime/evalkit复用同一个产品完整validator并归一化schema+全部机械semantic violations：sole `nextActions[i].label` maxLength200和/或24/64走`action_labels`；定向generation内部目标18/52。其它任意schema、semantic或mixed violation走`whole_report`。每轮输出均完整复验；product generation与judge各自最多4调用，具体attempt/retry/reason/scope只属于内部持久化/manifest，不进入HTTP schema。
- A-200引入新current diff后，旧baseline clean PASS失效；必须重新生成preserved old-baseline audit、re-freeze并运行`make openapi-diff`。codegen-check需独立实际通过，不由本决策预先宣称。

机器可执行的完整 finding oracle 为 [OPENAPI-001-report-direct-semantics.expected-findings.json](./OPENAPI-001-report-direct-semantics.expected-findings.json)。ADR 正文是决策解释，003 wrapper 必须对该 JSON 的 `severity + JSON pointer + kind + before + after` 做顺序无关 exact-set 校验；缺 finding、多 finding 或 severity 漂移都失败，不靠人工合并概念条目。

### 2.4 Report-owned conversation locator correction

作为同一次未上线 v1.0.0 freeze correction 的扩展，执行一进一出替换，总量保持 37 operations / 10 tags：

1. 删除 `GET /api/v1/practice/sessions`、operationId `listPracticeSessions`、对应 query 与 `PaginatedPracticeSession`（确认无其它引用后）。删除默认 fixture、generated client/server method、mux/handler/service/store、mock registry、inventory 与 frontend import/consumer；不保留 deprecated operation、空列表、redirect、alias 或隐藏兼容 route。
2. 新增 `GET /api/v1/reports/{reportId}/conversation`、operationId `getReportConversation`。唯一用户 locator 是 `reportId`，服务端按 `reportId + current user` 读取 owned `feedback_reports`，再沿既有 `session_id` 读取 `practice_messages ORDER BY seq_no ASC`。
3. response 使用 closed `ReportConversation`，required 字段精确为 `reportId`、`reportStatus`、`context`、`messages`。`context` 复用冻结的 `ReportContextSnapshot`；每条 closed `ReportConversationMessage` 只包含 `sequence`、`role`、`content`、`createdAt`。不得暴露 `sessionId`、message ID、`clientMessageId`、`replyStatus`、`replyGeneration`、anchors 或其它内部 locator。
4. 只要 owned report row 存在，queued/generating/ready/failed 均可读取；missing/deleted/cross-user 统一 hidden 404。空 identity、未知 role、重复或非递增 sequence、额外 locator、报告/会话绑定不一致必须整份 fail closed，不得部分返回、猜测修复或重排。
5. endpoint 只读、无 AI、无写入、无 pagination/list、无新数据库对象。保留 `POST /practice/sessions startPracticeSession` 与 `GET /practice/sessions/{sessionId} getPracticeSession`，分别承担创建与 live-session recovery。

OPENAPI-001 v1.7 machine oracle 必须在实施 Phase 的 RED 后，以 merge-base old baseline 与 proposed OpenAPI 自动生成并 exact-match `severity + path + kind + before + after`；本次设计修订不得手写或预造 expected-findings JSON。只有新旧 operation、schemas 与相关 query/required/closure findings 的完整集合可授权，任何顺带 drift 都失败。

## 3 影响

| 边界 | 受影响的项 | Owner |
|------|-----------|-------|
| 契约 | OpenAPI report/request schemas、baseline、merge-base diff gate | openapi-v1-contract 001/003 |
| Fixtures | Reports / PracticePlans、删除 PracticeSessions list fixture、prototype sync、Prism projection | openapi-v1-contract 002 |
| 后端 | generated DTO、report mapper/store/generator、derived plan handler、report-owned conversation read model；删除 session-list runtime surface | backend-review/001 + backend-practice/001 |
| 前端 | generated TS、Generating/Report/ReportConversation、route/request negatives；删除 session-list consumer | frontend-report-dashboard/001 |
| Mock/downstream | report fixtures + generated consumers + list-operation zero-reference | registry、backend-review、backend-practice 与 frontend-report owners |

## 4 迁移与回滚

- **迁移顺序**：先提交本 ADR；B1 conventions gate 固化 canonical literal/retryability；报告会话扩展先由 prototype owner补齐静态页面；再运行旧 merge-base baseline → 新 OpenAPI audit；随后同批更新 schema、fixtures/prototype、Go/TS codegen、backend、frontend consumers并删除 session-list surface；最后原地 re-freeze `openapi-v1.0.0.yaml`。
- **放行条件**：旧 baseline findings 与 §2.1 exact-match；`make lint-openapi`、fixture validation、prototype sync 两次幂等、Prism smoke、codegen check、consumer tests 与 current-baseline `make openapi-diff` 全部通过。
- **回滚**：任一 consumer 未能同批迁移、finding 超出 §2.1/§2.4、privacy 泄漏或核心场景失败时，整体回滚 OpenAPI/fixtures/codegen/backend/frontend/baseline；不得单独保留兼容字段或恢复公共 session list 作为旁路。
- **SemVer**：baseline 尚未发布，因此本次保持 `v1.0.0` 并作为 accepted pre-release correction 原地 re-freeze；发布后同类变更必须使用 v2.0.0。

## 5 相关

- [openapi-v1-contract spec](../spec.md) D-21 / D-32 / §4.4
- [backend-review spec](../../backend-review/spec.md)
- [frontend-report-dashboard spec](../../frontend-report-dashboard/spec.md)
- [product-scope spec](../../product-scope/spec.md)

## 6 审计

| 项 | 内容 |
|----|------|
| 提议人 | backend-review orchestration owner |
| Review | product owner confirmed 方案 A on 2026-07-12, its 200/24/64 + repair-margin 18/52 boundary on 2026-07-13, and removal of public `listPracticeSessions` in favor of report-owned conversation on 2026-07-15 |
| 实施分支 | design owner `design/report-conversation-record`; implementation branch由 `/implement` 创建 |
| base-ref diff evidence | implementation Phase 5 records exact finding artifact before baseline re-freeze |
| baseline | `openapi/baseline/openapi-v1.0.0.yaml` pre-release correction |
| history | `2026-07-12 | 1.45 | OPENAPI-001 oracle severity/additive correction` |

## 7 修订记录

| 日期 | 版本 | 变更 | 关联 |
|------|------|------|------|
| 2026-07-15 | 1.7 | Extend the accepted pre-release correction with report-owned read-only conversation; replace public `listPracticeSessions` one-for-one, preserve 37/10 and existing report-session binding, and require generated exact oracle during implementation RED. | B2 001/002/003 + backend/frontend report owners |
| 2026-07-13 | 1.6 | Close FeedbackReport state machine：ready non-null summary/preparedness/provenance + non-empty dimensions/actions；failed-only non-null errorCode. | B2/F3/frontend-report |
| 2026-07-13 | 1.5 | Keep max4 generation/judge attempts internal：no attempt/retry/reason/scope HTTP fields or retry endpoint；status remains queued/generating/ready/failed. | backend-review/frontend-report-dashboard |
| 2026-07-13 | 1.4 | Clarify evalkit/runtime full-validator reuse：sole-label targeted repair；all other/mixed whole-report repair；one-budget full revalidation and second-invalid fail-close. | B2/F3 |
| 2026-07-13 | 1.3 | Finalize A：wire fuse200；semantic/UX 24 words / 64 code points；targeted repair internal margin18/52. | B2/F3/frontend-report |
| 2026-07-13 | 1.2 | Accept A-200：ReportNextAction.label fuse200；14/40 remains UX gate；expected finding after=minLength1,maxLength200. | B2/F3/frontend-report |
| 2026-07-12 | 1.1 | Make the oracle exact across severity and record `REPORT_CONTEXT_TOO_LARGE` enum widening as additive-only. | B1 001 + B2 003 |
| 2026-07-12 | 1.0 | Accept grounded direct report pre-release correction. | OPENAPI-001 |
