# 00. 全局统一约定（Shared Conventions）

> **状态**: historical-input
> **更新日期**: 2026-05-03
> **执行边界**: 本文件是旧产品 spec 阶段的历史技术输入，不是当前共享枚举、错误码、ID 或异步 job 的可执行真理源。当前实现以 `shared/conventions.yaml`、生成代码和 `docs/spec/shared-conventions-codified/spec.md` 为准；本文件中已删除的旧练习模式、错题本、成长中心等枚举不得恢复。

## 1. 目标

本文件定义前端 TypeScript、后端 Go、数据库 PostgreSQL、内部事件与日志字段之间的统一约定。所有其它文档默认继承本文件，除非显式覆盖。

---

## 2. 全局规则

### 2.1 命名规则

| 层 | 规则 | 示例 |
|---|---|---|
| JSON API 字段 | `camelCase` | `targetJobId`, `analysisStatus` |
| TypeScript 类型 / 接口 | `PascalCase` | `TargetJob`, `PracticeSession` |
| TypeScript 变量 / 函数 | `camelCase` | `fetchTargetJob`, `currentTurn` |
| Go 类型 | `PascalCase` | `TargetJob`, `PracticeMode` |
| Go 字段 JSON tag | `camelCase` | ``json:"targetJobId"`` |
| Go 包目录 | `lowercase` | `targetjob`, `practice`, `review` |
| PostgreSQL 表 / 列 | `snake_case` | `target_jobs`, `analysis_status` |
| 枚举值 | `lower_snake_case` | `core_interview`, `well_prepared` |
| 错误码 | `UPPER_SNAKE_CASE` | `TARGET_JOB_NOT_FOUND` |
| 内部事件名 | `dot.case` | `target.imported`, `report.generated` |
| 指标名 | `snake_case` | `practice_sessions_started_total` |

### 2.2 ID 规则

- **所有业务主键统一使用 UUIDv7**，在 API 中以字符串形式暴露。
- 前端、后端都不得假设 ID 可读或带业务语义。
- URL path 中统一使用 `{resourceId}` 风格，例如 `/targets/{targetId}`。
- 若前端需要临时对象 ID，使用 `tmp_` 前缀，仅在浏览器内有效，不可上送到正式业务表。

### 2.3 时间与时区

- API 时间字段统一为 **RFC3339 / ISO 8601 UTC**。
- 数据库存储统一使用 `timestamptz`。
- 前端显示时才根据用户时区本地化。
- 示例：`2026-04-23T13:45:12Z`

### 2.4 语言与地区

- 语言字段统一使用 **BCP 47** 标记：`en`, `zh-CN`, `en-SG`
- 地区字段可使用：
  - 城市 / 国家自由文本：`Singapore`
  - 或未来扩展为 ISO 3166 代码
- P0 / P1 先以自由文本为主，避免过度标准化阻塞产品迭代。

### 2.5 金额与成本

- 所有模型成本统一以 **美元微单位**保存：`costUsdMicros`
- 例：`$0.0235` 存为 `23500`
- 前端展示时再格式化成小数金额

---

## 3. API 包装规则

### 3.1 成功响应

- 读取型接口直接返回资源对象
- 写入型接口返回资源对象，或返回异步作业对象
- 列表接口统一使用 `items + pageInfo`

```json
{
  "items": [],
  "pageInfo": {
    "nextCursor": "eyJpZCI6IjAx...",
    "pageSize": 20,
    "hasMore": true
  }
}
```

### 3.2 错误响应

统一使用：

```json
{
  "error": {
    "code": "TARGET_JOB_NOT_FOUND",
    "message": "target job not found",
    "requestId": "req_01HV...",
    "retryable": false,
    "details": {
      "targetJobId": "0195..."
    }
  }
}
```

#### 错误码约束

- 错误码必须稳定、可程序化处理
- 不把内部实现细节暴露给前端
- 错误码命名建议：`<DOMAIN>_<ERROR>`
- 示例：
  - `AUTH_UNAUTHORIZED`
  - `TARGET_IMPORT_FAILED`
  - `PRACTICE_SESSION_CONFLICT`
  - `REPORT_NOT_READY`
  - `VALIDATION_FAILED`
  - `RATE_LIMITED`

### 3.3 分页

统一采用 **cursor pagination**：

- 请求参数：`pageSize`, `cursor`
- 响应字段：`pageInfo.nextCursor`, `pageInfo.hasMore`
- 默认 `pageSize = 20`
- 最大 `pageSize = 100`

### 3.4 幂等

以下接口要求支持 `Idempotency-Key`：

- `POST /uploads/presign`
- `POST /targets/import`
- `POST /practice/plans`
- `POST /practice/sessions`
- `POST /practice/sessions/{id}/complete`
- `POST /resume/tailor`
- `POST /debriefs`
- `POST /privacy/exports`
- `POST /privacy/deletions`

幂等 key 建议生存期：**24 小时**。

### 3.5 请求头

| Header | 必填 | 说明 |
|---|---|---|
| `Authorization` | 是 | `Bearer <access_token>` |
| `X-Request-ID` | 建议 | 前端生成或网关注入 |
| `Idempotency-Key` | 写接口建议 | 防止重复提交 |
| `traceparent` | 建议 | 透传分布式追踪 |
| `X-Client-Version` | 建议 | 前端版本 |
| `Accept-Language` | 可选 | UI 语言偏好 |

---

## 4. 异步任务约定

长耗时操作统一采用 **异步 Job 模式**：

- 目标岗位导入 / 解析
- 报告生成
- 简历定制
- Source 刷新
- 数据导出 / 删除

### 4.1 Job 对象

```json
{
  "id": "0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e",
  "jobType": "target_import",
  "status": "queued",
  "resourceType": "target_job",
  "resourceId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "errorCode": null,
  "createdAt": "2026-04-23T13:45:12Z",
  "updatedAt": "2026-04-23T13:45:12Z"
}
```

### 4.2 Job 状态

- `queued`
- `running`
- `succeeded`
- `failed`
- `cancelled`
- `dead`

---

## 5. 枚举目录

### 5.1 目标岗位状态

```text
draft
preparing
applied
interviewing
offer
rejected
archived
```

### 5.2 目标岗位解析状态

```text
queued
processing
ready
failed
```

### 5.3 练习模式

```text
warmup
core_interview
single_drill
counter_questions
debrief_replay
```

### 5.4 练习目标

```text
baseline
sprint
fix_mistake
debrief
```

### 5.5 面试角色

```text
generalist
hr
hiring_manager
technical_manager
peer
```

### 5.6 练习会话状态

```text
queued
running
waiting_user_input
completing
completed
failed
cancelled
```

### 5.7 报告状态

```text
queued
generating
ready
failed
```

### 5.8 准备度档位

```text
not_ready
needs_practice
basically_ready
well_prepared
```

### 5.9 维度状态

```text
strong
meets_bar
needs_work
```

### 5.10 置信度

```text
high
medium
low
```

### 5.11 错题状态

```text
open
improving
mastered
```

### 5.12 真实面试复盘状态

```text
draft
completed
```

### 5.13 隐私请求类型 / 状态

类型：

```text
export
delete
```

状态：

```text
queued
processing
completed
failed
cancelled
```

---

## 6. 前后端类型映射

### 6.1 TypeScript 参考定义

```ts
export type PracticeMode =
  | 'warmup'
  | 'core_interview'
  | 'single_drill'
  | 'counter_questions'
  | 'debrief_replay';

export type SessionStatus =
  | 'queued'
  | 'running'
  | 'waiting_user_input'
  | 'completing'
  | 'completed'
  | 'failed'
  | 'cancelled';

export interface PageInfo {
  nextCursor: string | null;
  pageSize: number;
  hasMore: boolean;
}

export interface ApiError {
  code: string;
  message: string;
  requestId: string;
  retryable: boolean;
  details?: Record<string, unknown>;
}
```

### 6.2 Go 参考定义

```go
package types

type PracticeMode string

const (
    PracticeModeWarmup          PracticeMode = "warmup"
    PracticeModeCoreInterview   PracticeMode = "core_interview"
    PracticeModeSingleDrill     PracticeMode = "single_drill"
    PracticeModeCounterQuestion PracticeMode = "counter_questions"
    PracticeModeDebriefReplay   PracticeMode = "debrief_replay"
)

type SessionStatus string

const (
    SessionStatusQueued           SessionStatus = "queued"
    SessionStatusRunning          SessionStatus = "running"
    SessionStatusWaitingUserInput SessionStatus = "waiting_user_input"
    SessionStatusCompleting       SessionStatus = "completing"
    SessionStatusCompleted        SessionStatus = "completed"
    SessionStatusFailed           SessionStatus = "failed"
    SessionStatusCancelled        SessionStatus = "cancelled"
)

type APIError struct {
    Code       string         `json:"code"`
    Message    string         `json:"message"`
    RequestID  string         `json:"requestId"`
    Retryable  bool           `json:"retryable"`
    Details    map[string]any `json:"details,omitempty"`
}
```

---

## 7. 版本化规则

### 7.1 API 版本

- URL 版本：`/api/v1`
- 非破坏性新增字段允许直接在 `v1` 内迭代
- 破坏性变更需要：
  - 新字段并行上线
  - 前后端兼容窗口
  - 废弃公告
  - 最终再切 `v2`

### 7.2 Prompt / Rubric / Model 版本

任何会影响输出解释性的 AI 结果都必须落以下字段：

- `promptVersion`
- `rubricVersion`
- `modelId`
- `provider`
- `language`
- `featureFlagSnapshot`

### 7.3 Event Schema 版本

事件 envelope 包含：

- `eventName`
- `eventVersion`
- `occurredAt`
- `eventId`

事件 payload 的破坏性变更不得直接覆盖旧版本，必须通过 `eventVersion` 升级。

---

## 8. 安全与敏感信息约束

### 8.1 不得写入普通日志的内容

- 原始简历全文
- 原始 JD 全文
- 用户完整回答全文
- 面试复盘中的敏感个人信息
- access token / refresh token / session secret
- 文件下载签名 URL

### 8.2 可以记录但必须脱敏的内容

- 用户邮箱：只允许 hash 或前 3 位 + 域名掩码
- IP 地址：只允许 hash 或网段
- 外部链接：允许记录域名，不记录敏感 querystring

### 8.3 允许进入数据库但不允许进入日志的内容

- 原始 JD 文本
- 原始回答文本
- 简历解析结果
- 复盘问题清单
- 报告全文

---

## 9. 契约生成与代码生成

为保持前后端一致性，建议采用：

1. **API 文档先行**
2. 生成 `OpenAPI 3.1`
3. 前端从 OpenAPI 生成：
   - TypeScript types
   - API client
   - 可选 Zod schema
4. 后端从 OpenAPI 生成：
   - Go DTO
   - 路由接口约束
   - 示例与 contract test

推荐工具（可替换）：

- Go：`oapi-codegen` / `ogen`
- TS：`openapi-typescript` / `orval`

---

## 10. 最小一致性检查清单

每次新功能进入开发前，都应先检查这 8 件事：

- [ ] JSON 字段是否使用 camelCase
- [ ] DB 列名是否使用 snake_case
- [ ] 枚举值是否复用已有定义
- [ ] 错误码是否可程序化处理
- [ ] 是否需要 `Idempotency-Key`
- [ ] 是否需要异步 Job
- [ ] 是否要记录 prompt / rubric / model 版本
- [ ] 日志中是否可能泄露敏感原文
