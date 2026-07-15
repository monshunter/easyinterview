# PROTOTYPE_MAPPING

> **版本**: 1.8
> **状态**: active
> **更新日期**: 2026-07-15

把 [ui-design/src/data.jsx](../../ui-design/src/data.jsx) 的 mock 数据节映射到 OpenAPI v1 contract 的 P0 关键 operationId。`make sync-fixtures-from-prototype` 只读这张表 + data.jsx，把映射结果写入 §3 列出的 fixture 的 `scenarios.prototype-baseline` 节。该 scenario 是 spec §4.7 锁定的「ui 原型同源」入口；同步工具不会改写任何 fixture 的 `scenarios.default`。

## 1 同步原则

- **真理源单向**：data.jsx 是 prototype 的真理源；fixtures 是 OpenAPI 契约对外的 mock 真理源；同步工具单向把 data.jsx 折成 OpenAPI schema-valid 的 fixture body。
- **schema 校验 fail-fast**：写入前同步工具会跑一次 schema 校验，schema 不通过直接 fail 并打印 mapping 缺口（`Mapping gap: ...`）。不允许 sync 工具静默兜底或自动重命名。
- **缺失数据节直接跳过**：data.jsx 缺哪一节，对应 fixture 的 `prototype-baseline` 就不写入；不强制全量覆盖。但本表列出的 P0 关键 6 个 endpoint 必须存在数据节并写入 prototype-baseline。
- **id / 时间归一化**：data.jsx 中的 `tj-1` / `m1` / `今天 15:48` 等 prototype 风格在写入前会被归一化成 OpenAPI 契约要求的 UUIDv7 字面量与 RFC3339 UTC 时间，归一化方式由同步工具内部约定（见 §4）。
- **额外字段不可保留**：data.jsx 的展示字段（如 `statusTone` / `readinessLabel` 等）若 OpenAPI schema 不接受，同步工具必须丢弃，不写入 fixture。

## 2 主映射表（一对一 / 一对多）

| data.jsx 节 | OpenAPI operationId | 关系 | Tag | 说明 |
|-------------|---------------------|------|-----|------|
| `user` | `getMe` | 1:1 | Auth | `email` → `emailMasked`（脱敏）；`name` → `displayName`；`locale` → `uiLanguage` / `preferredPracticeLanguage`。 |
| `targetJobs[]` | `listTargetJobs` | 1:N | TargetJobs | 每个 `tj-N` 映射为 `TargetJob`：`title/company → companyName/locationText/language → targetLanguage`；`status` 取 OpenAPI enum 中最贴近的值；`statusTone/level/updatedAt` 等展示字段不入 fixture，来源 provenance 与 `latestReportId` 均不属于 TargetJob wire projection。 |
| `targetJobs[0]` + `jdSample` | `getTargetJob` | N:1 | TargetJobs | 取第一个 target job 的核心字段，再用 `jdSample.mustHave` / `jdSample.nice` 填 `requirements[]`，`jdSample.hidden` 折成 `summary.coreThemes`，`jdSample.rounds` 折成 2~5 条 `summary.interviewRounds[]`（含 `sequence/type/name/durationMinutes/focus`）。 |
| `sessionTranscript` | `getPracticeSession` | 1:1 | PracticeSessions | 按时间顺序映射为 `messages[]`；不生成题号、当前题或 turn 状态。 |
| `report` + `targetJobs[0]` | `getFeedbackReport` | N:1 | Reports | `report` 的 direct report 字段原样投影；`targetJobs[0]` 只用于生成稳定 `targetJobId`。同步器归一化冻结的 plan/resume/report/session UUID，并将公司名替换为通用 fixture 名。 |
| `reportConversation` | `getReportConversation` | 1:1 | Reports | 报告归属的只读会话记录：归一化 `reportId` 与冻结 context 的 plan/resume UUID；按 `sequence` 投影有序 `user` / `assistant` messages；只保留 `sequence/role/content/createdAt`，不投影 session 或消息内部 locator。 |

## 3 P0 闭环关键 endpoint 覆盖（plan 2.4 自检）

| operationId | 数据来源 | 映射状态 |
|-------------|----------|----------|
| `getMe` | `user` | ✅ |
| `listTargetJobs` | `targetJobs[]` | ✅ |
| `getTargetJob` | `targetJobs[0]` + `jdSample` | ✅ |
| `getPracticeSession` | `targetJobs[0]` + `sessionTranscript` | ✅ |
| `getFeedbackReport` | `report` + `targetJobs[0]` | ✅ |
| `getReportConversation` | `reportConversation` | ✅ |

## 4 归一化规则（同步工具内部约定）

- **id**: 同步工具用 `uuidv7_from_prototype("<section>:<prototype-id>")` 生成稳定 UUIDv7 字面量；同一 prototype id 多次跑产生相同 UUIDv7（保证幂等）。
- **datetime**: 所有时间字段统一固定为 prototype 的「现在」=`2026-04-28T13:45:12Z` 与固定锚点 `2026-04-28T12:00:00Z` / `2026-04-22T09:30:00Z`，不读取真实当前时间。
- **enum 翻译**:
    - `targetJobs[].status` 中文 → `TargetJobStatus`：`面试中→interviewing`、`准备中→preparing`、`草稿→draft`。
    - `targetJobs[].language` → `targetLanguage`：`中文→zh-CN`、`英文→en`。
    - TargetJob prototype projection 不读取或生成来源类型/URL；当前 JD intake 只有 paste-only 请求，response 不再暴露来源 provenance。
    - `report.preparednessLevel`、`dimensionAssessments[].status/confidence`、`highlights/issues[].dimensionCode` 与 `nextActions[].type` 必须已经使用当前 OpenAPI enum/字段，不做旧字段兼容翻译。
- **报告冻结上下文**: `report.context` 必须包含当前最小上下文；同步器只把 prototype plan/resume id 归一化为 UUIDv7，并把公司名替换为通用 fixture 名，不从其他展示字段重新推断轮次或结论。
- **报告派生焦点**: 仅投影 `retryFocusDimensionCodes`；`retryFocusCompetencyCodes`、`dimension`、question/turn 级报告字段不会被恢复。
- **报告记录 UI 镜像**: `data.jsx.reportOverview` 只镜像 `listTargetJobReports` 的 closed response shape：`targetJobId + rounds[].round/currentReport/latestAttempt`。UI 使用 `item.round.{roundId,roundSequence}`；不从 TargetJob 读取或重建 `latestReportId`。该 endpoint 的完整状态/失败矩阵由 hand-authored named scenarios 承担，不额外生成冗余 `prototype-baseline`。
- **报告会话记录**: `data.jsx.reportConversation` 单向生成 `getReportConversation.prototype-baseline`；其 `reportId`、冻结 context 与有序消息是闭合 read projection。消息 Markdown / GFM 文本原样保留（仅做 fixture 隐私文本清理），`sessionId`、`id`、`clientMessageId`、`replyStatus`、`replyGeneration` 与 `anchor` 等内部 locator 必须丢弃。该 prototype-baseline 的冻结上下文独立于 hand-authored 状态矩阵，后者仍共享自身 fixture context。
- **provenance**: AI schema 的 `provenance` 不依赖 data.jsx，由同步工具直接填入与 default scenario 同款的 `GenerationProvenance`（6 字段非空，非评分场景填 `not_applicable`）。
- **emailMasked**: `user.email` 走 `_masked_email("alice@example.com") → "ali***@example.com"`。
- **lossy 字段**: 任何 OpenAPI 不接受的展示字段（`statusTone` / `readinessLabel` / `t` / `qIdx` / 中文「2 小时前」等）必须丢弃，不能写入 fixture body。

## 5 修订规则

- 新增 data.jsx 数据节 / 新增 endpoint 时需要更新本表；同步工具读不到本表声明的节会报 `Mapping gap`。
- 翻译规则（§4）变化必须同步更新；同步工具内的 enum 翻译表与本文件保持一致。
- 同步工具内部 UUIDv7 派生算法不允许变更；变更会让所有 prototype-baseline id 漂移、破坏幂等。
