# 02. API 定义（REST / JSON）

## 1. 概览

- 基础路径：`/api/v1`
- 鉴权：`Authorization: Bearer <access_token>`
- 数据格式：`application/json`
- 时间格式：RFC3339 UTC
- 长耗时任务：返回 `202 Accepted` + `Job`
- 命名规则：见 `00-shared-conventions.md`

本文件以 **接口设计与联调** 为目标，建议后续直接转成 OpenAPI 3.1。

---

## 2. 通用约定

### 2.1 Header

| Header | 必填 | 说明 |
|---|---|---|
| `Authorization` | 是 | 访问令牌 |
| `X-Request-ID` | 建议 | 请求链路追踪 |
| `Idempotency-Key` | 创建类接口建议 | 防重 |
| `traceparent` | 建议 | OpenTelemetry Trace 透传 |
| `Accept-Language` | 可选 | UI 语言 |

### 2.2 状态码

| 状态码 | 用途 |
|---|---|
| `200 OK` | 读取成功 |
| `201 Created` | 同步创建成功 |
| `202 Accepted` | 异步任务已接收 |
| `204 No Content` | 删除 / 空响应成功 |
| `400 Bad Request` | 参数非法 |
| `401 Unauthorized` | 未登录 / token 无效 |
| `403 Forbidden` | 无资源权限 |
| `404 Not Found` | 资源不存在 |
| `409 Conflict` | 状态冲突 / 幂等冲突 |
| `422 Unprocessable Entity` | 业务校验失败 |
| `429 Too Many Requests` | 频率限制 |
| `500 Internal Server Error` | 未分类错误 |

### 2.3 错误响应

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "invalid request body",
    "requestId": "req_01J3...",
    "retryable": false,
    "details": {
      "fields": [
        {
          "field": "source.type",
          "reason": "unsupported value"
        }
      ]
    }
  }
}
```

### 2.4 分页响应

```json
{
  "items": [],
  "pageInfo": {
    "nextCursor": null,
    "pageSize": 20,
    "hasMore": false
  }
}
```

---

## 3. 公共对象模型

### 3.1 Job

```json
{
  "id": "0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e",
  "jobType": "report_generate",
  "status": "queued",
  "resourceType": "feedback_report",
  "resourceId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "errorCode": null,
  "createdAt": "2026-04-23T13:45:12Z",
  "updatedAt": "2026-04-23T13:45:12Z"
}
```

B2 `openapi-v1-contract` v1.4 锁定 P0 API-facing 字面量：

- `jobType`: `target_import` / `resume_parse` / `report_generate` / `resume_tailor` / `debrief_generate` / `privacy_export` / `privacy_delete`
- `resourceType`: `target_job` / `feedback_report` / `resume_asset` / `resume_tailor_run` / `debrief` / `privacy_request`

DB / worker 内部可以保留非 API-facing job type（如 `source_refresh` / `embedding_upsert`），但它们不得出现在 v1.0.0 `GET /api/v1/jobs/{jobId}` response 中，除非 B2 spec additive 修订后追加枚举值。

### 3.2 TargetJob

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | string | UUIDv7 |
| `status` | enum | `draft / preparing / applied / interviewing / offer / rejected / archived` |
| `analysisStatus` | enum | `queued / processing / ready / failed` |
| `title` | string | 岗位名 |
| `companyName` | string | 公司名 |
| `locationText` | string \| null | 地点 |
| `targetLanguage` | string | 练习语言 |
| `sourceType` | enum | `manual_text / url / file / manual_form` |
| `sourceUrl` | string \| null | 来源链接 |
| `summary` | object \| null | 解析摘要 |
| `requirements` | array | 结构化要求 |
| `fitSummary` | object \| null | 用户与岗位的命中 / 缺口 |
| `latestReportId` | string \| null | 最近报告 |
| `openMistakeCount` | number | 未攻克错题数 |
| `createdAt` | string | 创建时间 |
| `updatedAt` | string | 更新时间 |

### 3.3 PracticePlan

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | string | UUIDv7 |
| `targetJobId` | string | 关联岗位 |
| `goal` | enum | `baseline / sprint / fix_mistake / debrief` |
| `mode` | enum | `warmup / core_interview / single_drill / counter_questions / debrief_replay` |
| `interviewerPersona` | enum | `generalist / hr / hiring_manager / technical_manager / peer` |
| `difficulty` | string | `easy / standard / stretch` |
| `language` | string | 练习语言 |
| `timeBudgetMinutes` | number | 预计时长 |
| `questionBudget` | number | 预计题数 |
| `status` | string | `draft / ready / archived` |
| `createdAt` | string | 创建时间 |

### 3.4 PracticeSession

```json
{
  "id": "0195f309-7e62-7184-8cba-66f5db647061",
  "planId": "0195f307-076a-77d4-9135-c6ba39dcae95",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "status": "waiting_user_input",
  "language": "en",
  "hintsEnabled": true,
  "turnCount": 1,
  "currentTurn": {
    "id": "0195f30b-9be1-71ae-81c0-9bf5e65d8f45",
    "turnIndex": 1,
    "questionText": "Tell me about a time you influenced a cross-functional team without authority.",
    "questionIntent": "leadership_without_authority",
    "status": "asked",
    "askedAt": "2026-04-23T13:47:18Z"
  },
  "createdAt": "2026-04-23T13:47:18Z",
  "updatedAt": "2026-04-23T13:47:18Z"
}
```

### 3.5 AssistantAction

```json
{
  "type": "ask_follow_up",
  "turnId": "0195f30b-9be1-71ae-81c0-9bf5e65d8f45",
  "questionText": "What measurable outcome did your intervention produce?",
  "hint": null,
  "sessionStatus": "waiting_user_input"
}
```

`type` 可取：

- `ask_question`
- `ask_follow_up`
- `show_hint`
- `session_wait`
- `session_completed`

### 3.6 FeedbackReport

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | string | UUIDv7 |
| `sessionId` | string | 关联会话 |
| `targetJobId` | string | 关联岗位 |
| `status` | enum | `queued / generating / ready / failed` |
| `preparednessLevel` | enum | `not_ready / needs_practice / basically_ready / well_prepared` |
| `highlights` | array | 亮点 |
| `issues` | array | 问题 |
| `nextActions` | array | 下一步建议 |
| `questionAssessments` | array | 逐题结果 |
| `mistakeIds` | array | 写入错题本的条目 |
| `promptVersion` | string | prompt 版本 |
| `rubricVersion` | string | rubric 版本 |
| `modelId` | string | 模型标识 |
| `createdAt` | string | 创建时间 |
| `updatedAt` | string | 更新时间 |

### 3.7 MistakeEntry

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | string | UUIDv7 |
| `targetJobId` | string | 来源岗位 |
| `sourceSessionId` | string \| null | 来源会话 |
| `sourceDebriefId` | string \| null | 来源复盘 |
| `competencyCode` | string | 能力点代码 |
| `questionText` | string | 原始问题 |
| `answerSummary` | string | 当时回答摘要 |
| `failureReasons` | array | 失守原因 |
| `recommendedFramework` | string | 推荐回答结构 |
| `status` | enum | `open / improving / mastered` |
| `priority` | number | 1-100，越大越优先 |
| `createdAt` | string | 创建时间 |
| `updatedAt` | string | 更新时间 |

### 3.8 Debrief

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | string | UUIDv7 |
| `targetJobId` | string | 关联岗位 |
| `status` | enum | `draft / completed` |
| `roundType` | string | `hr_screen / hiring_manager / behavioral / technical / culture / custom` |
| `interviewerRole` | string | 面试官角色 |
| `questions` | array | 真实问题与回答 |
| `riskItems` | array | 高风险项 |
| `nextRoundChecklist` | array | 下一轮清单 |
| `thankYouDraft` | string \| null | 感谢信草稿 |
| `createdAt` | string | 创建时间 |
| `updatedAt` | string | 更新时间 |

---

## 4. 鉴权与用户上下文

### `GET /api/v1/me`

返回当前用户最小上下文。

**Response 200**

```json
{
  "id": "0195f1f1-bd01-70d3-8c92-6b1d4100dba5",
  "emailMasked": "ali***@example.com",
  "displayName": "Alice",
  "uiLanguage": "zh-CN",
  "preferredPracticeLanguage": "en"
}
```

---

## 5. 上传与文件

### 5.1 `POST /api/v1/uploads/presign`

为 JD 文件、简历文件申请上传地址。

**Request**

```json
{
  "purpose": "resume",
  "fileName": "alice_resume.pdf",
  "contentType": "application/pdf",
  "byteSize": 281234
}
```

`purpose` 可取：

- `resume`
- `target_job_attachment`
- `privacy_export`

**Response 201**

```json
{
  "fileObjectId": "0195f220-e1e8-7537-83b6-e9fd38b21f3d",
  "uploadUrl": "https://storage.example.com/...",
  "method": "PUT",
  "headers": {
    "Content-Type": "application/pdf"
  },
  "expiresAt": "2026-04-23T14:15:12Z"
}
```

错误码：

- `UNSUPPORTED_FILE_TYPE`
- `FILE_TOO_LARGE`

---

## 6. 画像与经历卡

### 6.1 `GET /api/v1/profiles/me`

获取当前用户画像。

### 6.2 `PATCH /api/v1/profiles/me`

更新画像 Lite 信息。

**Request**

```json
{
  "headline": "B2B SaaS Product Manager",
  "yearsOfExperience": 6,
  "currentRole": "Senior Product Manager",
  "preferredPracticeLanguage": "en",
  "uiLanguage": "zh-CN",
  "region": "Singapore"
}
```

**Response 200**

返回更新后的 Profile。

### 6.3 `GET /api/v1/profiles/me/experience-cards`

分页获取经历卡。

### 6.4 `POST /api/v1/profiles/me/experience-cards`

创建经历卡。

**Request**

```json
{
  "title": "Led onboarding revamp",
  "companyName": "Acme",
  "situation": "Activation dropped after pricing change.",
  "task": "Improve first-week activation rate.",
  "action": "Redesigned onboarding and aligned sales/support.",
  "result": "Activation improved from 32% to 47%.",
  "skills": ["product_strategy", "cross_functional_leadership"],
  "language": "en"
}
```

### 6.5 `PATCH /api/v1/profiles/me/experience-cards/{cardId}`

更新经历卡。

---

## 7. 简历资产

### 7.1 `POST /api/v1/resumes`

注册已上传简历文件，触发解析。

**Request**

```json
{
  "fileObjectId": "0195f220-e1e8-7537-83b6-e9fd38b21f3d",
  "title": "General PM Resume",
  "language": "en"
}
```

**Response 202**

```json
{
  "resumeAssetId": "0195f23a-e026-7135-a32c-e865a6feaa92",
  "job": {
    "id": "0195f23b-6f32-7d44-a08e-0a52fbb5aa24",
    "jobType": "resume_parse",
    "status": "queued",
    "resourceType": "resume_asset",
    "resourceId": "0195f23a-e026-7135-a32c-e865a6feaa92",
    "errorCode": null,
    "createdAt": "2026-04-23T13:39:08Z",
    "updatedAt": "2026-04-23T13:39:08Z"
  }
}
```

### 7.2 `GET /api/v1/resumes/{resumeAssetId}`

获取简历资产及解析状态。

**Response 200**

```json
{
  "id": "0195f23a-e026-7135-a32c-e865a6feaa92",
  "title": "General PM Resume",
  "language": "en",
  "parseStatus": "ready",
  "fileObjectId": "0195f220-e1e8-7537-83b6-e9fd38b21f3d",
  "parsedSummary": {
    "headline": "Senior PM with 6 years in B2B SaaS",
    "skills": ["product_strategy", "analytics", "stakeholder_management"]
  },
  "createdAt": "2026-04-23T13:39:08Z",
  "updatedAt": "2026-04-23T13:39:25Z"
}
```

---

## 8. 目标岗位工作台

### 8.1 `POST /api/v1/targets/import`

支持从文本、链接或文件导入目标岗位。

**Request（URL 导入）**

```json
{
  "source": {
    "type": "url",
    "url": "https://jobs.example.com/pm-123"
  },
  "titleHint": "Senior Product Manager",
  "companyNameHint": "Acme",
  "targetLanguage": "en"
}
```

**Request（文本导入）**

```json
{
  "source": {
    "type": "manual_text",
    "rawText": "We are looking for a PM with B2B SaaS and analytics..."
  },
  "targetLanguage": "en"
}
```

**Request（文件导入）**

```json
{
  "source": {
    "type": "file",
    "fileObjectId": "0195f250-d8b2-7a41-a8fb-95cc4a10c6f8"
  },
  "targetLanguage": "en"
}
```

**Response 202**

```json
{
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "job": {
    "id": "0195f2d0-4a44-7fc2-8f77-1f9c4ce1ae9e",
    "jobType": "target_import",
    "status": "queued",
    "resourceType": "target_job",
    "resourceId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
    "errorCode": null,
    "createdAt": "2026-04-23T13:45:12Z",
    "updatedAt": "2026-04-23T13:45:12Z"
  }
}
```

错误码：

- `TARGET_SOURCE_INVALID`
- `TARGET_IMPORT_FAILED`
- `TARGET_DUPLICATE_IMPORT`

### 8.2 `GET /api/v1/targets`

查询岗位列表。

**Query**

- `status`
- `analysisStatus`
- `q`
- `cursor`
- `pageSize`

### 8.3 `GET /api/v1/targets/{targetJobId}`

获取目标岗位工作台。

**Response 200**

```json
{
  "id": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "status": "preparing",
  "analysisStatus": "ready",
  "title": "Senior Product Manager",
  "companyName": "Acme",
  "locationText": "Singapore",
  "targetLanguage": "en",
  "sourceType": "url",
  "sourceUrl": "https://jobs.example.com/pm-123",
  "summary": {
    "coreThemes": ["b2b_saas", "analytics", "cross_functional_leadership"],
    "interviewHypotheses": ["leadership", "stakeholder_alignment", "metrics_depth"]
  },
  "requirements": [
    {
      "id": "0195f2d4-5f1a-7849-8bfa-6342f64c891a",
      "kind": "must_have",
      "label": "5+ years in product management",
      "evidenceLevel": "explicit"
    }
  ],
  "fitSummary": {
    "strengths": ["B2B SaaS background", "strong analytics usage"],
    "gaps": ["limited marketplace experience"],
    "riskSignals": ["few examples on organizational influence"]
  },
  "latestReportId": "0195f33f-2d0d-72b2-b80d-20b0ee06f0d7",
  "openMistakeCount": 3,
  "createdAt": "2026-04-23T13:45:12Z",
  "updatedAt": "2026-04-23T13:46:08Z"
}
```

### 8.4 `PATCH /api/v1/targets/{targetJobId}`

更新岗位生命周期状态或元数据。

**Request**

```json
{
  "status": "interviewing",
  "locationText": "Singapore",
  "notes": "Phone screen scheduled next Tuesday."
}
```

---

## 9. 练习计划

### 9.1 `POST /api/v1/practice/plans`

创建练习计划。

**Request**

```json
{
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "goal": "sprint",
  "mode": "core_interview",
  "interviewerPersona": "hiring_manager",
  "difficulty": "standard",
  "language": "en",
  "questionBudget": 6,
  "timeBudgetMinutes": 20,
  "resumeAssetId": "0195f23a-e026-7135-a32c-e865a6feaa92",
  "focusCompetencyCodes": ["motivation", "leadership", "counter_questions"]
}
```

**Response 201**

返回 `PracticePlan`。

错误码：

- `TARGET_JOB_NOT_READY`
- `RESUME_ASSET_NOT_READY`

### 9.2 `GET /api/v1/practice/plans/{planId}`

获取计划详情。

---

## 10. 练习会话

### 10.1 `POST /api/v1/practice/sessions`

开始一轮会话。

**Request**

```json
{
  "planId": "0195f307-076a-77d4-9135-c6ba39dcae95",
  "hintsEnabled": true
}
```

**Response 201**

返回 `PracticeSession`，其中已包含首题。

### 10.2 `GET /api/v1/practice/sessions/{sessionId}`

获取会话最新状态，用于刷新 / 恢复。

### 10.3 `POST /api/v1/practice/sessions/{sessionId}/events`

追加会话事件，并返回下一步 AI 动作。

#### 支持的事件类型

| kind | payload |
|---|---|
| `answer_submitted` | `turnId`, `answerText` |
| `hint_requested` | `turnId` |
| `turn_skipped` | `turnId`, `reason` |
| `session_paused` | `reason` |
| `session_resumed` | 无 |

**Request（answer_submitted）**

```json
{
  "clientEventId": "evt_web_01J3M...",
  "kind": "answer_submitted",
  "occurredAt": "2026-04-23T13:48:02Z",
  "payload": {
    "turnId": "0195f30b-9be1-71ae-81c0-9bf5e65d8f45",
    "answerText": "In my last role I had to align sales, design and engineering..."
  }
}
```

**Response 200**

```json
{
  "acknowledged": true,
  "session": {
    "id": "0195f309-7e62-7184-8cba-66f5db647061",
    "status": "waiting_user_input",
    "turnCount": 1,
    "updatedAt": "2026-04-23T13:48:04Z"
  },
  "assistantAction": {
    "type": "ask_follow_up",
    "turnId": "0195f30b-9be1-71ae-81c0-9bf5e65d8f45",
    "questionText": "What measurable outcome did this cross-functional effort achieve?",
    "hint": null,
    "sessionStatus": "waiting_user_input"
  }
}
```

错误码：

- `PRACTICE_SESSION_NOT_ACTIVE`
- `PRACTICE_TURN_NOT_FOUND`
- `PRACTICE_EVENT_CONFLICT`
- `AI_PROVIDER_TIMEOUT`

### 10.4 `POST /api/v1/practice/sessions/{sessionId}/complete`

结束会话并触发报告生成。

**Request**

```json
{
  "clientCompletedAt": "2026-04-23T14:05:15Z"
}
```

**Response 202**

```json
{
  "reportId": "0195f33f-2d0d-72b2-b80d-20b0ee06f0d7",
  "job": {
    "id": "0195f341-e084-7e50-84d8-66e9707110eb",
    "jobType": "report_generate",
    "status": "queued",
    "resourceType": "feedback_report",
    "resourceId": "0195f33f-2d0d-72b2-b80d-20b0ee06f0d7",
    "errorCode": null,
    "createdAt": "2026-04-23T14:05:15Z",
    "updatedAt": "2026-04-23T14:05:15Z"
  }
}
```

---

## 11. 报告

### 11.1 `GET /api/v1/reports/{reportId}`

获取报告。若尚未完成，仍返回占位状态。

**Response 200（生成中）**

```json
{
  "id": "0195f33f-2d0d-72b2-b80d-20b0ee06f0d7",
  "status": "generating",
  "sessionId": "0195f309-7e62-7184-8cba-66f5db647061",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "createdAt": "2026-04-23T14:05:15Z",
  "updatedAt": "2026-04-23T14:05:18Z"
}
```

**Response 200（完成）**

```json
{
  "id": "0195f33f-2d0d-72b2-b80d-20b0ee06f0d7",
  "status": "ready",
  "sessionId": "0195f309-7e62-7184-8cba-66f5db647061",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "preparednessLevel": "needs_practice",
  "highlights": [
    {
      "dimension": "job_relevance",
      "evidence": "You consistently mapped product decisions back to activation metrics.",
      "confidence": "high"
    }
  ],
  "issues": [
    {
      "dimension": "content_specificity",
      "evidence": "Leadership examples lacked numeric outcome details in 2 answers.",
      "confidence": "high"
    }
  ],
  "nextActions": [
    {
      "type": "single_drill",
      "label": "Re-practice leadership with metrics and outcome framing"
    }
  ],
  "questionAssessments": [
    {
      "turnId": "0195f30b-9be1-71ae-81c0-9bf5e65d8f45",
      "questionIntent": "leadership_without_authority",
      "dimensionResults": {
        "contentSpecificity": { "status": "needs_work", "confidence": "high" },
        "structureClarity": { "status": "meets_bar", "confidence": "medium" },
        "jobRelevance": { "status": "strong", "confidence": "high" }
      },
      "writtenToMistakeBook": true
    }
  ],
  "mistakeIds": ["0195f34e-8dcf-73b5-8b70-c7c8f032bb1e"],
  "promptVersion": "report_v1.3.0",
  "rubricVersion": "rubric_core_v1.1.0",
  "modelId": "providerA/gpt-x",
  "createdAt": "2026-04-23T14:05:15Z",
  "updatedAt": "2026-04-23T14:05:42Z"
}
```

### 11.2 `GET /api/v1/targets/{targetJobId}/reports`

分页获取岗位下的报告列表。

---

## 12. 错题本

### 12.1 `GET /api/v1/mistakes`

查询错题列表。

**Query**

- `targetJobId`
- `status`
- `competencyCode`
- `cursor`
- `pageSize`

### 12.2 `POST /api/v1/mistakes/{mistakeId}/retest`

基于错题发起复练。接口会创建新的 PracticePlan。

**Request**

```json
{
  "mode": "single_drill",
  "language": "en",
  "timeBudgetMinutes": 8
}
```

**Response 201**

```json
{
  "plan": {
    "id": "0195f36f-1c78-7bde-a2b2-e8bc72b2d8ea",
    "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
    "goal": "fix_mistake",
    "mode": "single_drill",
    "interviewerPersona": "generalist",
    "difficulty": "standard",
    "language": "en",
    "timeBudgetMinutes": 8,
    "questionBudget": 2,
    "status": "ready",
    "createdAt": "2026-04-23T14:08:00Z"
  }
}
```

---

## 13. 简历定制

### 13.1 `POST /api/v1/resume/tailor`

为某份简历生成针对某岗位的改写建议。

**Request**

```json
{
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "resumeAssetId": "0195f23a-e026-7135-a32c-e865a6feaa92",
  "mode": "bullet_suggestions"
}
```

`mode` 可取：

- `gap_review`
- `bullet_suggestions`

**Response 202**

```json
{
  "tailorRunId": "0195f37e-dd35-7f23-8f96-0208e6d34520",
  "job": {
    "id": "0195f380-781d-72d0-a81e-1790f46d6ee1",
    "jobType": "resume_tailor",
    "status": "queued",
    "resourceType": "resume_tailor_run",
    "resourceId": "0195f37e-dd35-7f23-8f96-0208e6d34520",
    "errorCode": null,
    "createdAt": "2026-04-23T14:09:12Z",
    "updatedAt": "2026-04-23T14:09:12Z"
  }
}
```

### 13.2 `GET /api/v1/resume/tailor-runs/{tailorRunId}`

获取简历定制结果。

**Response 200**

```json
{
  "id": "0195f37e-dd35-7f23-8f96-0208e6d34520",
  "status": "ready",
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "resumeAssetId": "0195f23a-e026-7135-a32c-e865a6feaa92",
  "matchSummary": {
    "strengths": ["B2B SaaS and analytics are already visible."],
    "gaps": ["Outcome framing is weaker in leadership bullets."]
  },
  "suggestions": [
    {
      "originalBullet": "Worked with engineering and design on onboarding improvements.",
      "suggestedBullet": "Led a cross-functional onboarding revamp with engineering and design, increasing first-week activation from 32% to 47%.",
      "reason": "Adds measurable outcome and leadership signal."
    }
  ],
  "createdAt": "2026-04-23T14:09:12Z",
  "updatedAt": "2026-04-23T14:09:36Z"
}
```

---

## 14. 真实面试复盘

### 14.1 `POST /api/v1/debriefs`

创建真实面试复盘并触发增强总结。

**Request**

```json
{
  "targetJobId": "0195f2cf-67ef-7df4-a9f7-2fbd1135320d",
  "roundType": "hiring_manager",
  "interviewerRole": "hiring_manager",
  "language": "en",
  "questions": [
    {
      "questionText": "Why do you want to join our company?",
      "myAnswerSummary": "I focused on the product mission and growth stage.",
      "interviewerReaction": "Positive but asked for more specifics."
    },
    {
      "questionText": "Tell me about a failed launch.",
      "myAnswerSummary": "I explained the context but not the follow-up metrics.",
      "interviewerReaction": "Neutral"
    }
  ],
  "notes": "I felt weakest on leadership depth."
}
```

**Response 202**

```json
{
  "debriefId": "0195f39f-ec79-7d2b-a217-6264f344d6d1",
  "job": {
    "id": "0195f3a1-d55a-7138-83f6-f7e4a18fa62d",
    "jobType": "debrief_generate",
    "status": "queued",
    "resourceType": "debrief",
    "resourceId": "0195f39f-ec79-7d2b-a217-6264f344d6d1",
    "errorCode": null,
    "createdAt": "2026-04-23T14:11:24Z",
    "updatedAt": "2026-04-23T14:11:24Z"
  }
}
```

### 14.2 `GET /api/v1/debriefs/{debriefId}`

获取复盘结果。

---

## 15. 成长看板

### `GET /api/v1/growth/overview`

**Query**

- `window`: `7d / 30d / 90d`
- `targetJobId`（可选）

**Response 200**

```json
{
  "window": "30d",
  "summary": {
    "practiceSessionsCompleted": 8,
    "reportsReady": 8,
    "mistakesOpened": 11,
    "mistakesMastered": 4,
    "debriefCount": 2
  },
  "preparednessTrend": [
    { "date": "2026-04-05", "level": "needs_practice" },
    { "date": "2026-04-17", "level": "basically_ready" }
  ],
  "dimensionTrend": {
    "contentSpecificity": "improving",
    "jobRelevance": "strong",
    "followUpHandling": "needs_work"
  }
}
```

---

## 16. 异步 Job 查询

### `GET /api/v1/jobs/{jobId}`

返回 Job 当前状态。

**Response 200**

```json
{
  "id": "0195f341-e084-7e50-84d8-66e9707110eb",
  "jobType": "report_generate",
  "status": "running",
  "resourceType": "feedback_report",
  "resourceId": "0195f33f-2d0d-72b2-b80d-20b0ee06f0d7",
  "errorCode": null,
  "createdAt": "2026-04-23T14:05:15Z",
  "updatedAt": "2026-04-23T14:05:22Z"
}
```

---

## 17. 隐私与合规

### 17.1 `POST /api/v1/privacy/exports`

请求导出用户数据。

**Response 202**

```json
{
  "privacyRequestId": "0195f3d1-b947-7075-a44d-416da8b5c9e9",
  "job": {
    "id": "0195f3d2-6081-709c-baf7-82ef5d52e119",
    "jobType": "privacy_export",
    "status": "queued",
    "resourceType": "privacy_request",
    "resourceId": "0195f3d1-b947-7075-a44d-416da8b5c9e9",
    "errorCode": null,
    "createdAt": "2026-04-23T14:13:45Z",
    "updatedAt": "2026-04-23T14:13:45Z"
  }
}
```

### 17.2 `POST /api/v1/privacy/deletions`

请求删除用户数据。

**Response 202**

与导出一致，只是 `jobType = privacy_delete`。

### 17.3 `GET /api/v1/privacy/requests/{privacyRequestId}`

查询隐私请求状态与结果。

---

## 18. 事件与状态一致性要求

1. `POST /practice/sessions/{id}/events` 必须以 `clientEventId` 去重，防止重复提交回答。
2. `POST /practice/sessions/{id}/complete` 必须幂等，多次点击只能生成一个激活中的 report job。
3. `GET /reports/{id}` 在 `status != ready` 时也必须返回 200，而不是 404，便于前端轮询。
4. TargetJob 的 `status` 与 `analysisStatus` 分离：
   - `status` 是岗位生命周期
   - `analysisStatus` 是解析任务状态
5. 任何 AI 生成结果若 `schema validation` 失败：
   - API 不得返回半结构化脏数据
   - Worker 应把结果标为 `failed`，并附 `errorCode`

---

## 19. 推荐的 OpenAPI 拆分方式

建议拆为以下 tags：

- `Auth`
- `Uploads`
- `Profile`
- `Resumes`
- `TargetJobs`
- `PracticePlans`
- `PracticeSessions`
- `Reports`
- `Mistakes`
- `ResumeTailor`
- `Debriefs`
- `Growth`
- `Jobs`
- `Privacy`

并在生成时同步产出：

- TypeScript SDK
- Go DTO / Handler Interface
- Mock 数据与 contract test
