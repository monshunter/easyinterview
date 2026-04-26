# 06. 内部事件契约（Event Contracts）

## 1. 目标

内部事件用于连接 API、Worker、报告生成、错题写入、source 刷新与审计，而不是为了“提前微服务化”。P0 / P1 推荐采用：

- **业务写库**
- **写 outbox**
- **由 Worker / Dispatcher 发布与消费**

这样既能保持实现简单，也能保证关键异步副作用不丢失。

---

## 2. 传输与一致性模型

### 2.1 推荐实现

- 业务事务提交时，同时写入 `outbox_events`
- 独立 dispatcher 轮询 `outbox_events`
- 将事件投递到：
  - Asynq job
  - 或内部 handler
- 发布成功后把 `publish_status` 标记为 `published`

### 2.2 语义

- **At-least-once**
- 消费方必须幂等
- 同一事件可能重复投递，必须根据 `eventId` 或业务主键去重

### 2.3 不采用的方式

P0 / P1 不建议直接依赖：

- 跨多服务实时事务
- 复杂 event choreography
- 没有 outbox 的“先写库后发消息”

---

## 3. 标准事件 Envelope

```json
{
  "eventId": "0195f4a8-5c7d-7de0-b248-e6bc7f2ca53b",
  "eventName": "report.generated",
  "eventVersion": 1,
  "aggregateType": "feedback_report",
  "aggregateId": "0195f33f-2d0d-72b2-b80d-20b0ee06f0d7",
  "occurredAt": "2026-04-23T14:05:42Z",
  "producer": "worker",
  "traceId": "7f5a8c351f5f4d8bb8f34f4f6efc8f8a",
  "payload": {}
}
```

### 3.1 字段说明

| 字段 | 必填 | 说明 |
|---|---|---|
| `eventId` | 是 | UUIDv7 |
| `eventName` | 是 | 事件名，`dot.case` |
| `eventVersion` | 是 | 从 1 开始递增 |
| `aggregateType` | 是 | 聚合根类型 |
| `aggregateId` | 是 | 聚合根 ID |
| `occurredAt` | 是 | 事件发生时间 |
| `producer` | 是 | 生产者模块 |
| `traceId` | 否 | 追踪链路 |
| `payload` | 是 | 事件负载 |

---

## 4. 事件目录

## 4.1 目标岗位相关

### `target.import.requested`

**生产者**：API  
**消费者**：Job Dispatcher / Worker

**触发时机**：用户调用 `POST /targets/import`

**Payload**

```json
{
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "userId": "0195f1f1-bd01-70d3-8c92-6b1d4100dba5",
  "sourceType": "url",
  "targetLanguage": "en"
}
```

### `target.parsed`

**生产者**：Worker  
**消费者**：

- Retrieval chunk upsert
- Analytics
- 可选 Source refresh handler

**触发时机**：JD 解析成功

**Payload**

```json
{
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "userId": "0195f1f1-bd01-70d3-8c92-6b1d4100dba5",
  "analysisStatus": "ready",
  "requirementCount": 7,
  "coreThemes": ["b2b_saas", "analytics", "leadership"]
}
```

### `target.analysis_failed`

**生产者**：Worker  
**消费者**：

- Analytics
- Alerting
- 可选重试调度器

**Payload**

```json
{
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "errorCode": "AI_OUTPUT_INVALID",
  "retryable": true
}
```

---

## 4.2 练习会话相关

### `practice.session.started`

**生产者**：API  
**消费者**：

- Analytics
- Session watchdog（可选）

**Payload**

```json
{
  "sessionId": "0195f309-7e62-7184-8cba-66f5db647061",
  "planId": "0195f307-076a-77d4-9135-c6ba39dcae95",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "goal": "sprint",
  "mode": "core_interview",
  "language": "en"
}
```

### `practice.turn.completed`

**生产者**：API  
**消费者**：

- Lightweight analytics
- Optional turn-level quality sampler

**触发时机**：某一题完成（不代表整轮会话结束）

**Payload**

```json
{
  "sessionId": "0195f309-7e62-7184-8cba-66f5db647061",
  "turnId": "0195f30b-9be1-71ae-81c0-9bf5e65d8f45",
  "turnIndex": 1,
  "questionIntent": "leadership_without_authority",
  "followUpCount": 1,
  "answerCharLength": 684
}
```

### `practice.session.completed`

**生产者**：API  
**消费者**：

- Report generation job creator
- Analytics

**触发时机**：用户点击完成会话

**Payload**

```json
{
  "sessionId": "0195f309-7e62-7184-8cba-66f5db647061",
  "planId": "0195f307-076a-77d4-9135-c6ba39dcae95",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "turnCount": 6,
  "language": "en"
}
```

---

## 4.3 报告与错题相关

### `report.generation.requested`

**生产者**：API / Dispatcher  
**消费者**：Worker

**Payload**

```json
{
  "reportId": "0195f33f-2d0d-72b2-b80d-20b0ee06f0d7",
  "sessionId": "0195f309-7e62-7184-8cba-66f5db647061",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d"
}
```

### `report.generated`

**生产者**：Worker  
**消费者**：

- Mistake book updater
- Growth aggregator
- Analytics
- Notification (future)

**Payload**

```json
{
  "reportId": "0195f33f-2d0d-72b2-b80d-20b0ee06f0d7",
  "sessionId": "0195f309-7e62-7184-8cba-66f5db647061",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "preparednessLevel": "needs_practice",
  "mistakeCount": 3,
  "promptVersion": "report_v1.3.0",
  "rubricVersion": "rubric_core_v1.1.0",
  "modelId": "providerA/gpt-x"
}
```

### `report.generation_failed`

**生产者**：Worker  
**消费者**：

- Alerting
- Analytics
- Retry scheduler（可选）

**Payload**

```json
{
  "reportId": "0195f33f-2d0d-72b2-b80d-20b0ee06f0d7",
  "sessionId": "0195f309-7e62-7184-8cba-66f5db647061",
  "errorCode": "AI_PROVIDER_TIMEOUT",
  "retryable": true
}
```

### `mistake.created`

**生产者**：Worker / Review module  
**消费者**：

- Growth aggregator
- Analytics

**Payload**

```json
{
  "mistakeId": "0195f34e-8dcf-73b5-8b70-c7c8f032bb1e",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "sourceSessionId": "0195f309-7e62-7184-8cba-66f5db647061",
  "competencyCode": "leadership",
  "status": "open",
  "priority": 85
}
```

### `mistake.status_changed`

**生产者**：Review module  
**消费者**：

- Growth aggregator
- Analytics

**Payload**

```json
{
  "mistakeId": "0195f34e-8dcf-73b5-8b70-c7c8f032bb1e",
  "fromStatus": "improving",
  "toStatus": "mastered",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d"
}
```

---

## 4.4 简历与复盘相关

### `resume.parse.completed`

**生产者**：Worker  
**消费者**：

- Retrieval chunk upsert
- Analytics

**Payload**

```json
{
  "resumeAssetId": "0195f23a-e026-7135-a32c-e865a6feaa92",
  "userId": "0195f1f1-bd01-70d3-8c92-6b1d4100dba5",
  "parseStatus": "ready"
}
```

### `resume.tailor.completed`

**生产者**：Worker  
**消费者**：

- Analytics
- Notification（future）

**Payload**

```json
{
  "tailorRunId": "0195f37e-dd35-7f23-8f96-0208e6d34520",
  "resumeAssetId": "0195f23a-e026-7135-a32c-e865a6feaa92",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "mode": "bullet_suggestions",
  "status": "ready"
}
```

### `debrief.created`

**生产者**：API  
**消费者**：

- Debrief generation worker
- Analytics

**Payload**

```json
{
  "debriefId": "0195f39f-ec79-7d2b-a217-6264f344d6d1",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "roundType": "hiring_manager",
  "questionCount": 2
}
```

### `debrief.completed`

**生产者**：Worker  
**消费者**：

- Mistake extractor
- Growth aggregator
- Analytics

**Payload**

```json
{
  "debriefId": "0195f39f-ec79-7d2b-a217-6264f344d6d1",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "riskItemCount": 3,
  "generatedMistakeCount": 2
}
```

---

## 4.5 Source / Privacy 相关

### `source.refreshed`

**生产者**：Worker  
**消费者**：

- Source cache updater
- Analytics

**Payload**

```json
{
  "sourceRecordId": "0195f411-fd6d-7a55-b5c7-658f7ae4cf28",
  "ownerType": "target_job",
  "ownerId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "freshnessStatus": "fresh"
}
```

### `privacy.request.created`

**生产者**：API  
**消费者**：

- Privacy worker
- Audit sink

**Payload**

```json
{
  "privacyRequestId": "0195f3d1-b947-7075-a44d-416da8b5c9e9",
  "userId": "0195f1f1-bd01-70d3-8c92-6b1d4100dba5",
  "requestType": "delete"
}
```

### `privacy.request.completed`

**生产者**：Worker  
**消费者**：

- Audit sink
- Notification（future）

**Payload**

```json
{
  "privacyRequestId": "0195f3d1-b947-7075-a44d-416da8b5c9e9",
  "userId": "0195f1f1-bd01-70d3-8c92-6b1d4100dba5",
  "requestType": "delete",
  "status": "completed"
}
```

---

## 5. 事件与 Job 的关系

不是所有事件都直接等于一个 Job，但以下场景通常会转成 Job：

| 事件 | 后续 Job |
|---|---|
| `target.import.requested` | `target_import` |
| `practice.session.completed` | `report_generate` |
| `debrief.created` | `debrief_generate` |
| `privacy.request.created` | `privacy_export` / `privacy_delete` |
| `target.parsed` | `embedding_upsert`（可选） |
| `resume.parse.completed` | `embedding_upsert`（可选） |

建议实现方式：

1. API 写业务表
2. API 写 outbox event
3. dispatcher / handler 决定是否创建 async job

这样避免 API handler 同时承担太多副作用。

---

## 6. 幂等规则

### 6.1 生产者幂等

- 每个 `outbox_event.id` 唯一
- 同一业务动作只生成一个主事件
- 重试写库时要防止重复事件

### 6.2 消费者幂等

消费方至少基于以下之一去重：

- `eventId`
- `aggregateType + aggregateId + eventName + eventVersion`
- `job_type + dedupe_key`

### 6.3 常见重复场景

- API 超时后前端重试
- dispatcher 发布后 ACK 丢失
- worker 成功但状态更新失败
- consumer 重启后重放消息

---

## 7. 事件版本化

### 7.1 非破坏性变更

允许：

- 新增可选字段
- 新增 payload 子字段
- 新增消费者

### 7.2 破坏性变更

必须：

- `eventVersion + 1`
- 旧版本与新版本并行消费一段时间
- 在代码中显式保留版本分支

### 7.3 命名规则

- 领域.动作
- 使用过去式表示已发生事实
- 示例：
  - `target.parsed`
  - `report.generated`
  - `mistake.created`

不要使用模糊命名：

- `something.updated`
- `entity.changed`
- `data.processed`

---

## 8. 失败与补偿

### 8.1 事件发布失败

- `outbox_events.publish_status = failed`
- dispatcher 重试
- 达到上限后告警并进入人工排查队列

### 8.2 消费失败

- job / consumer 写失败日志
- 可重试错误：按指数退避
- 不可重试错误：标记 dead letter / dead job

### 8.3 业务补偿例子

#### 场景：报告已写入，但错题写入失败

处理建议：

- 保持 `report.generated` 成功
- 额外发 `mistake.sync_failed`
- 后续补偿任务扫描 `written_to_mistake_book = true` 但无 mistake 记录的数据进行修复

---

## 9. 监控建议

对事件系统本身，至少监控：

- `outbox_events{publish_status="pending"}` 积压数
- 平均发布延迟
- 事件消费失败数
- dead job 数
- 同一事件重复消费数（若可统计）

---

## 10. 落地建议

P0 / P1 推荐的最小实现：

1. 用 PostgreSQL `outbox_events`
2. 用 worker dispatcher 扫描未发布事件
3. 将重任务转成 `async_jobs`
4. 在 consumer 内实现幂等写库
5. 所有关键事件写结构化日志
6. 所有关键事件进入 analytics 或指标聚合

这一套做法既能满足主链路稳定性，也不会在早期把团队拖入复杂消息中间件治理。