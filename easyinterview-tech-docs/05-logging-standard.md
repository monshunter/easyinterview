# 05. 日志规范（Logging Standard）

## 1. 目标

日志的目标不是“尽量多打”，而是：

1. 能快速排查核心链路故障
2. 能追踪 AI 任务的版本、耗时、成本和失败原因
3. 能支持安全审计与隐私请求
4. 不泄露用户敏感原文

---

## 2. 总体原则

### 2.1 统一结构化 JSON

后端所有应用日志统一输出单行 JSON，不允许生产环境混入自由文本日志。

### 2.2 一条日志只表达一个事实

- 一次请求开始
- 一次请求结束
- 一次 job 开始
- 一次 AI 调用完成
- 一次状态迁移
- 一次异常

不要在一条日志里塞过多上下文副本。

### 2.3 日志不是数据库

以下内容不应写入普通日志：

- 原始简历全文
- 原始 JD 全文
- 用户完整回答全文
- 完整 prompt 模板正文
- 完整模型输出正文
- 预签名 URL
- access token / refresh token

### 2.4 可追踪优先于可读性

每条日志都要能用以下上下文串起来：

- `requestId`
- `traceId`
- `jobId`
- `resourceType`
- `resourceId`

---

## 3. 日志分类

| 类别 | 说明 | 示例 |
|---|---|---|
| Access Log | HTTP 请求日志 | route、status、latency |
| Application Log | 业务状态迁移 / 关键动作 | report generated |
| Job Log | Worker 入队 / 出队 / 重试 | report_generate retry |
| AI Log | 模型调用、tokens、cost、fallback | followup_generate |
| Audit Log | 导出 / 删除 / 管理动作 | privacy delete requested |
| Security Log | 登录失败、越权、风控事件 | unauthorized access |
| Client Error | 前端异常上报 | React render error |

---

## 4. 必填字段

### 4.1 通用字段

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `ts` | string | 是 | RFC3339 UTC 时间 |
| `level` | string | 是 | `debug/info/warn/error` |
| `service` | string | 是 | `api`, `worker`, `web` |
| `env` | string | 是 | `dev`, `staging`, `prod` |
| `event` | string | 是 | 事件名称 |
| `message` | string | 是 | 简短说明 |
| `requestId` | string | 否 | HTTP 请求链路 ID |
| `traceId` | string | 否 | OpenTelemetry Trace ID |
| `spanId` | string | 否 | Span ID |
| `userId` | string | 否 | 仅在安全存储的应用日志里使用，不进指标 |
| `resourceType` | string | 否 | `target_job`, `practice_session` 等 |
| `resourceId` | string | 否 | 业务资源 ID |
| `version` | string | 是 | 服务版本 / git sha |

### 4.2 HTTP Access Log 额外字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `method` | string | HTTP 方法 |
| `route` | string | 归一化路由，如 `/api/v1/targets/{targetJobId}` |
| `statusCode` | number | HTTP 状态码 |
| `latencyMs` | number | 请求耗时 |
| `responseBytes` | number | 响应大小 |
| `clientVersion` | string | 前端版本 |
| `ipHash` | string | IP 的 hash |

### 4.3 Worker / Job Log 额外字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `jobId` | string | 异步 job ID |
| `jobType` | string | `target_import`, `report_generate` |
| `attempt` | number | 当前尝试次数 |
| `maxAttempts` | number | 最大重试次数 |
| `jobStatus` | string | `queued/running/succeeded/failed` |
| `queueLatencyMs` | number | 入队到执行延迟 |
| `durationMs` | number | 执行耗时 |

### 4.4 AI Log 额外字段

| 字段 | 类型 | 说明 |
|---|---|---|
| `aiTaskType` | string | `jd_parse`, `question_generate`, `report_generate` |
| `provider` | string | 模型供应商 |
| `modelId` | string | 完整模型标识 |
| `modelFamily` | string | 低基数分组，如 `gpt-x` |
| `promptVersion` | string | prompt 版本 |
| `rubricVersion` | string | rubric 版本 |
| `language` | string | 输出语言 |
| `inputTokens` | number | 输入 token |
| `outputTokens` | number | 输出 token |
| `costUsdMicros` | number | 成本 |
| `latencyMs` | number | AI 调用耗时 |
| `fallbackUsed` | boolean | 是否发生 fallback |
| `validationStatus` | string | `passed/auto_repaired/failed` |
| `errorCode` | string | 失败原因 |

---

## 5. 字段红线与脱敏规则

### 5.1 绝对禁止进入应用日志

- `rawJdText`
- `answerText`
- `resumeRawText`
- `thankYouDraft`
- `parsedSummary` 全量对象
- `promptTemplateBody`
- `modelRawResponse`
- 文件下载 URL / 上传 URL
- token / cookie / secret

### 5.2 允许记录摘要而非原文

| 内容 | 允许方式 |
|---|---|
| 回答文本 | `answerCharLength`, `answerLanguage`, `answerHash` |
| JD 文本 | `jdCharLength`, `jdHash` |
| prompt | `promptVersion`, `promptHash` |
| 模型输出 | `validationStatus`, `outputSchemaVersion` |
| 邮箱 | `emailHash` 或 mask |
| IP | `ipHash` |

### 5.3 hash 规则

建议使用：

- `sha256(value + salt)`
- salt 存在服务端安全配置中
- 前后端不要用不同规则，否则无法关联

---

## 6. 日志级别规范

### 6.1 `debug`

- 仅用于本地与开发环境
- 默认不在生产环境开启
- 不得包含敏感原文

### 6.2 `info`

用于正常业务流转的关键里程碑，例如：

- request completed
- session started
- job succeeded
- report generated
- privacy request created

### 6.3 `warn`

用于可恢复或需关注的异常，例如：

- AI fallback 发生
- 输出 auto-repair 后才通过
- 外部 source 过期
- 重试次数增加

### 6.4 `error`

用于真正失败，需要人工或告警关注：

- 关键 job 失败
- 数据不一致
- provider timeout 且 fallback 失败
- 资源越权
- 删除请求失败

---

## 7. 建议日志事件名

### 7.1 Access / API

- `http.request.started`
- `http.request.completed`
- `http.request.failed`

### 7.2 业务事件

- `target.import.requested`
- `target.import.completed`
- `practice.plan.created`
- `practice.session.started`
- `practice.event.accepted`
- `practice.session.completed`
- `report.generation.requested`
- `report.generated`
- `mistake.created`
- `mistake.status.changed`
- `resume.tailor.requested`
- `resume.tailor.completed`
- `debrief.created`
- `privacy.request.created`
- `privacy.request.completed`

### 7.3 AI 事件

- `ai.task.started`
- `ai.task.completed`
- `ai.task.failed`
- `ai.task.fallback`
- `ai.output.validation_failed`

### 7.4 安全 / 审计

- `auth.login.failed`
- `auth.token.invalid`
- `security.forbidden`
- `audit.export.created`
- `audit.delete.completed`

---

## 8. 示例日志

### 8.1 API Access Log

```json
{
  "ts": "2026-04-23T13:48:04Z",
  "level": "info",
  "service": "api",
  "env": "prod",
  "event": "http.request.completed",
  "message": "request completed",
  "requestId": "req_01J3M6A0W3MZ8T4C5M2YQ3N0FA",
  "traceId": "7f5a8c351f5f4d8bb8f34f4f6efc8f8a",
  "spanId": "84a8fbb14617a1d0",
  "userId": "0195f1f1-bd01-70d3-8c92-6b1d4100dba5",
  "method": "POST",
  "route": "/api/v1/practice/sessions/{sessionId}/events",
  "statusCode": 200,
  "latencyMs": 1842,
  "responseBytes": 734,
  "clientVersion": "web-0.7.3",
  "resourceType": "practice_session",
  "resourceId": "0195f309-7e62-7184-8cba-66f5db647061",
  "version": "api-1.14.2+9c13b2e"
}
```

### 8.2 Worker / Report Log

```json
{
  "ts": "2026-04-23T14:05:42Z",
  "level": "info",
  "service": "worker",
  "env": "prod",
  "event": "report.generated",
  "message": "feedback report generated",
  "jobId": "0195f341-e084-7e50-84d8-66e9707110eb",
  "jobType": "report_generate",
  "attempt": 1,
  "maxAttempts": 5,
  "jobStatus": "succeeded",
  "durationMs": 22163,
  "queueLatencyMs": 387,
  "userId": "0195f1f1-bd01-70d3-8c92-6b1d4100dba5",
  "resourceType": "feedback_report",
  "resourceId": "0195f33f-2d0d-72b2-b80d-20b0ee06f0d7",
  "version": "worker-1.14.2+9c13b2e"
}
```

### 8.3 AI Task Log

```json
{
  "ts": "2026-04-23T13:48:03Z",
  "level": "info",
  "service": "api",
  "env": "prod",
  "event": "ai.task.completed",
  "message": "follow-up generated",
  "requestId": "req_01J3M6A0W3MZ8T4C5M2YQ3N0FA",
  "traceId": "7f5a8c351f5f4d8bb8f34f4f6efc8f8a",
  "aiTaskType": "followup_generate",
  "provider": "providerA",
  "modelId": "providerA/gpt-x-2026-03",
  "modelFamily": "gpt-x",
  "promptVersion": "followup_v1.0.4",
  "rubricVersion": "rubric_turn_v1.0.2",
  "language": "en",
  "inputTokens": 1843,
  "outputTokens": 139,
  "costUsdMicros": 12800,
  "latencyMs": 1532,
  "fallbackUsed": false,
  "validationStatus": "passed",
  "resourceType": "practice_session",
  "resourceId": "0195f309-7e62-7184-8cba-66f5db647061",
  "version": "api-1.14.2+9c13b2e"
}
```

### 8.4 安全日志

```json
{
  "ts": "2026-04-23T14:15:02Z",
  "level": "warn",
  "service": "api",
  "env": "prod",
  "event": "security.forbidden",
  "message": "resource access denied",
  "requestId": "req_01J3M8FQY9Y63CZ0FK84Q5CHPB",
  "traceId": "bcaf7e3032e148bc90ce4c02d3184fb0",
  "userId": "0195f1f1-bd01-70d3-8c92-6b1d4100dba5",
  "resourceType": "feedback_report",
  "resourceId": "0195f33f-2d0d-72b2-b80d-20b0ee06f0d7",
  "errorCode": "AUTH_FORBIDDEN",
  "ipHash": "a7b5d8b5c1...",
  "version": "api-1.14.2+9c13b2e"
}
```

---

## 9. 前端日志 / 错误上报

前端不应把浏览器控制台 `console.log` 当成正式日志系统。生产环境建议：

- 业务埋点进入 analytics
- 运行时异常进入 Sentry
- 网络请求失败由 API 日志记录
- 前端仅上报必要的错误上下文：
  - route
  - requestId
  - clientVersion
  - browser info

禁止前端异常上报里携带：

- 输入框完整内容
- 简历原文
- 面试回答全文
- token / localStorage 中敏感信息

---

## 10. 采样与日志量控制

### 10.1 必须全量保留的日志

- `error`
- `warn`
- 隐私 / 审计事件
- 核心业务状态迁移（`report.generated`, `privacy.request.completed`）

### 10.2 可采样的日志

- 高频成功 access log
- 高频成功 AI task log
- 低价值 debug 日志

### 10.3 采样建议

- `http.request.completed`: 10% 采样（若已有 Prometheus）
- `ai.task.completed`: 100% 写入 DB `ai_task_runs`，应用日志可采样 20% 成功样本
- `ai.task.failed`: 100% 全量

---

## 11. 保留周期与访问权限

| 日志类型 | 建议保留 | 访问权限 |
|---|---|---|
| Access Log | 15~30 天 | 平台 / 后端 |
| Application Log | 30~90 天 | 平台 / 后端 |
| AI Error Log | 90 天 | 平台 / AI 工程 |
| Security Log | 180 天以上 | 安全 / 负责人 |
| Audit Log | 180 天以上 | 受限访问 |
| Frontend Error | 30~90 天 | 前端 / 平台 |

说明：

- 日志保留策略需与隐私政策一致
- 即使日志长期保留，也不得能重建用户完整原文

---

## 12. Go 落地建议

### 12.1 推荐库

- `zerolog`
- `zap`
- `slog`（如团队偏好标准库风格）

### 12.2 中间件职责

建议统一中间件注入：

- `requestId`
- `traceId`
- `userId`
- `clientVersion`
- `route`
- `latencyMs`

### 12.3 错误记录规则

记录错误时要同时包含：

- `errorCode`
- `errorClass`
- `retryable`
- `resourceType`
- `resourceId`

而不是只记一串 `error.message`。

---

## 13. 最低上线清单

- [ ] 所有 API route 有 access log
- [ ] 所有 Worker job 有开始 / 完成 / 失败日志
- [ ] 所有 AI 调用有 task 级日志
- [ ] 日志字段与本文件一致
- [ ] 日志中无原始 JD / 简历 / 回答明文
- [ ] 审计动作写入 `audit_events`
- [ ] 日志检索可按 `requestId`, `jobId`, `resourceId` 三种路径排查