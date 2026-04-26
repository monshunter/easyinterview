# 04. Metrics 与可观测性规范

## 1. 目标

可观测性不是“上线后补监控”，而是 easyinterview 产品可信度的一部分。P0 / P1 必须同时回答 4 类问题：

1. **产品是否跑通**  
   用户能否完成 JD 导入 -> 练习 -> 报告 -> 复练。
2. **体验是否稳定**  
   关键链路是否低延迟、低失败。
3. **AI 输出是否可信**  
   追问是否相关、报告是否空泛、评分是否异常。
4. **成本是否可控**  
   每次会话、每次报告、每个用户的模型成本是否失控。

---

## 2. 观测分层

| 层 | 工具 | 目标 |
|---|---|---|
| 产品分析 | PostHog / Segment / Warehouse | 漏斗、转化、留存、复练 |
| 服务指标 | Prometheus | 延迟、错误率、吞吐、资源占用 |
| 分布式追踪 | OpenTelemetry | 跨 API / Worker / AI 调用链路排障 |
| 日志 | Loki / ELK | 根因分析、审计辅助 |
| 错误追踪 | Sentry | 前端 / 后端异常聚合 |
| 质量评估 | 内部评估表 + Dashboard | 追问相关率、报告泛化率、异常高分 |

---

## 3. 通用指标命名规范

### 3.1 Prometheus 命名

- Counter：`*_total`
- Histogram：`*_duration_seconds`
- Gauge：`*_in_flight`, `*_queue_depth`
- 使用基础单位：
  - 时间：seconds
  - 大小：bytes
  - 金额：usd（如果必须）
- 不要用毫秒作为指标后缀，日志里可记 `latencyMs`

### 3.2 允许的 Prometheus labels

允许：

- `service`
- `route`
- `method`
- `status_code`
- `job_type`
- `task_type`
- `provider`
- `model_family`
- `language`
- `feature`
- `env`
- `result`

禁止：

- `user_id`
- `target_job_id`
- `session_id`
- `prompt_version`（过高基数，不作为指标 label，进入日志或 event）
- 原始 URL 全量 path
- 任意自由文本

### 3.3 事件埋点属性

产品分析事件中允许携带：

- `targetJobStatus`
- `analysisStatus`
- `practiceMode`
- `goal`
- `interviewerPersona`
- `language`
- `preparednessLevel`
- `mistakeStatus`
- `jobType`
- `reportStatus`
- `sourceType`

事件分析系统中如需 user 维度，使用平台自带匿名用户标识，不把真实邮箱作为事件属性。

---

## 4. 产品分析事件（Product Analytics）

## 4.1 关键漏斗

### P0 主漏斗

1. `target_import_requested`
2. `target_import_completed`
3. `practice_plan_created`
4. `practice_session_started`
5. `practice_session_completed`
6. `report_ready`
7. `report_viewed`
8. `mistake_retest_started`

### 真实复盘漏斗

1. `debrief_created`
2. `debrief_completed`
3. `mistake_created_from_debrief`
4. `practice_session_started_from_debrief`

### 简历定制漏斗

1. `resume_uploaded`
2. `resume_parse_ready`
3. `resume_tailor_requested`
4. `resume_tailor_ready`

## 4.2 事件定义

### `target_import_requested`

| 属性 | 类型 |
|---|---|
| `sourceType` | string |
| `targetLanguage` | string |
| `hasFile` | boolean |
| `hasUrl` | boolean |

### `target_import_completed`

| 属性 | 类型 |
|---|---|
| `sourceType` | string |
| `analysisStatus` | string |
| `requirementCount` | number |
| `coreThemeCount` | number |
| `durationMs` | number |

### `practice_plan_created`

| 属性 | 类型 |
|---|---|
| `goal` | string |
| `practiceMode` | string |
| `interviewerPersona` | string |
| `timeBudgetMinutes` | number |
| `questionBudget` | number |
| `language` | string |

### `practice_session_started`

| 属性 | 类型 |
|---|---|
| `practiceMode` | string |
| `goal` | string |
| `hintsEnabled` | boolean |
| `language` | string |

### `practice_turn_answer_submitted`

| 属性 | 类型 |
|---|---|
| `turnIndex` | number |
| `answerCharLength` | number |
| `followUpCountSoFar` | number |

### `practice_session_completed`

| 属性 | 类型 |
|---|---|
| `practiceMode` | string |
| `goal` | string |
| `turnCount` | number |
| `durationSeconds` | number |
| `hintsUsedCount` | number |

### `report_ready`

| 属性 | 类型 |
|---|---|
| `preparednessLevel` | string |
| `mistakeCount` | number |
| `durationMs` | number |
| `language` | string |

### `report_viewed`

| 属性 | 类型 |
|---|---|
| `preparednessLevel` | string |
| `mistakeCount` | number |
| `viewDurationSeconds` | number |

### `mistake_retest_started`

| 属性 | 类型 |
|---|---|
| `competencyCode` | string |
| `previousStatus` | string |
| `mode` | string |

### `debrief_created`

| 属性 | 类型 |
|---|---|
| `roundType` | string |
| `interviewerRole` | string |
| `questionCount` | number |
| `language` | string |

### `resume_tailor_requested`

| 属性 | 类型 |
|---|---|
| `mode` | string |
| `language` | string |

---

## 5. 产品 KPI 定义

### 5.1 首次价值达成率（First Value Rate）

定义：

> 新用户在注册后 24 小时内，至少完成一次 `target_import_completed` 和一次 `practice_session_started`

公式：

```text
首次价值达成率 =
24 小时内完成岗位导入 + 开始练习的新用户数
/
24 小时内注册的新用户数
```

### 5.2 会话完成率

```text
practice_session_completed_total
/
practice_session_started_total
```

### 5.3 报告阅读完成率

可定义为：

- 报告页停留 >= 20 秒
- 或滚动 / 展开了至少 2 个逐题区块

### 5.4 7 日复练率

```text
7 天内发生第二次及以上 practice_session_completed 的用户数
/
7 天内至少发生第一次 practice_session_completed 的用户数
```

### 5.5 真实面试复盘回流率

```text
创建 debrief 后 7 天内再次开始 practice session 的用户数
/
创建 debrief 的用户数
```

---

## 6. 后端服务指标（Prometheus）

## 6.1 HTTP / API 指标

| 指标名 | 类型 | Labels | 说明 |
|---|---|---|---|
| `http_server_requests_total` | Counter | `service,route,method,status_code` | HTTP 请求总数 |
| `http_server_request_duration_seconds` | Histogram | `service,route,method` | HTTP 请求时长 |
| `http_server_in_flight_requests` | Gauge | `service,route` | 并发请求数 |
| `http_server_response_size_bytes` | Histogram | `service,route` | 响应大小 |

## 6.2 数据库指标

| 指标名 | 类型 | Labels | 说明 |
|---|---|---|---|
| `db_query_duration_seconds` | Histogram | `service,operation` | SQL 执行时长 |
| `db_queries_total` | Counter | `service,operation,result` | SQL 执行次数 |
| `db_pool_in_use_connections` | Gauge | `service` | 连接池占用 |
| `db_pool_idle_connections` | Gauge | `service` | 连接池空闲 |
| `db_pool_wait_count_total` | Counter | `service` | 连接等待总数 |

## 6.3 Redis / Queue 指标

| 指标名 | 类型 | Labels | 说明 |
|---|---|---|---|
| `async_jobs_enqueued_total` | Counter | `job_type` | 入队次数 |
| `async_jobs_processed_total` | Counter | `job_type,result` | 处理结果数 |
| `async_job_duration_seconds` | Histogram | `job_type` | 单 job 耗时 |
| `async_job_queue_depth` | Gauge | `job_type` | 队列积压 |
| `async_job_lag_seconds` | Gauge | `job_type` | 排队延迟 |

---

## 7. 业务域指标

## 7.1 目标岗位

| 指标名 | 类型 | Labels | 说明 |
|---|---|---|---|
| `target_import_requests_total` | Counter | `source_type` | 导入请求数 |
| `target_import_completed_total` | Counter | `source_type,result` | 导入完成数 |
| `target_import_duration_seconds` | Histogram | `source_type` | 导入完成耗时 |
| `target_analysis_failures_total` | Counter | `source_type,error_code` | 解析失败数 |

## 7.2 练习

| 指标名 | 类型 | Labels | 说明 |
|---|---|---|---|
| `practice_plans_created_total` | Counter | `goal,mode,persona` | 创建计划数 |
| `practice_sessions_started_total` | Counter | `goal,mode,language` | 开始会话数 |
| `practice_sessions_completed_total` | Counter | `goal,mode,language` | 完成会话数 |
| `practice_session_duration_seconds` | Histogram | `mode,language` | 会话时长 |
| `practice_turns_completed_total` | Counter | `mode,language` | 完成题目数 |
| `practice_hints_requested_total` | Counter | `mode` | 请求提示次数 |

## 7.3 报告 / 错题

| 指标名 | 类型 | Labels | 说明 |
|---|---|---|---|
| `reports_requested_total` | Counter | `language` | 报告生成请求 |
| `reports_ready_total` | Counter | `language,result` | 报告生成成功 / 失败 |
| `report_generation_duration_seconds` | Histogram | `language` | 报告生成耗时 |
| `mistakes_created_total` | Counter | `competency_code` | 新错题数 |
| `mistakes_mastered_total` | Counter | `competency_code` | 攻克错题数 |

## 7.4 简历 / 复盘

| 指标名 | 类型 | Labels | 说明 |
|---|---|---|---|
| `resume_parse_total` | Counter | `result` | 简历解析次数 |
| `resume_tailor_total` | Counter | `mode,result` | 简历定制次数 |
| `resume_tailor_duration_seconds` | Histogram | `mode` | 简历定制耗时 |
| `debriefs_created_total` | Counter | `round_type,language` | 复盘创建数 |
| `debrief_generation_duration_seconds` | Histogram | `round_type` | 复盘生成耗时 |

---

## 8. AI 调用指标

## 8.1 LLM / Embedding 指标

| 指标名 | 类型 | Labels | 说明 |
|---|---|---|---|
| `ai_task_runs_total` | Counter | `task_type,provider,model_family,result` | AI 任务总数 |
| `ai_task_latency_seconds` | Histogram | `task_type,provider,model_family` | AI 调用耗时 |
| `ai_task_input_tokens_total` | Counter | `task_type,provider,model_family` | 输入 token |
| `ai_task_output_tokens_total` | Counter | `task_type,provider,model_family` | 输出 token |
| `ai_task_cost_usd_total` | Counter | `task_type,provider,model_family` | 美元成本累计 |
| `ai_output_validation_failures_total` | Counter | `task_type,provider,model_family` | 结构化输出校验失败 |
| `ai_fallback_total` | Counter | `task_type,from_model,to_model` | fallback 次数 |

### 8.2 成本控制指标

建议至少有 3 个视角：

1. 每次会话成本
2. 每份报告成本
3. 每活跃用户日成本

可从 `ai_task_runs` 聚合得到：

- `avg_cost_per_practice_session_usd`
- `avg_cost_per_report_usd`
- `avg_cost_per_active_user_day_usd`

这些可在 Warehouse / BI 看板中展示，不强制直接上 Prometheus。

---

## 9. AI 质量指标

> 这部分不全靠 Prometheus 直接产出，很多来自日志聚合、离线评估或人工抽检。

## 9.1 线上可计算 / 半自动指标

| 指标 | 说明 | 来源 |
|---|---|---|
| 追问相关率 | 追问是否围绕上一答的缺口 | 日志抽样 + 评估服务 |
| 角色漂移率 | HR / Hiring Manager 角色口吻是否失真 | 规则 + 抽检 |
| 报告空泛率 | 报告是否缺少具体证据 | 评估服务 |
| 异常高分率 | 准备度明显偏高且无证据支撑 | 规则 + 抽检 |
| 语言混乱率 | 中英混杂 / 非目标语言输出 | 规则检测 |
| stale source rate | 引用过期 source 的比例 | source freshness 校验 |

## 9.2 推荐质量指标定义

### 追问相关率（Follow-up Relevance Rate）

```text
相关追问数 / 总追问数
```

判定方式：

- 初期：人工抽检 + LLM Judge
- 目标：>= 70%

### 报告可执行性评分

由人工或标注任务按 1~5 分打分，衡量：

- 是否指出具体问题
- 是否给出可执行建议
- 是否绑定题目证据

### 岗位贴合主观评分

用户在报告页可选打分：

- “这次练习是否围绕目标岗位展开？”
- 1~5 分

### 错题命中率

定义：

> 后续复练是否真正覆盖先前错题涉及的能力点

可从 `mistake_entries.competency_code` 与后续 `practice_plans.focus_competency_codes` 比较。

---

## 10. SLI / SLO

## 10.1 SLI

| SLI | 定义 |
|---|---|
| Target Import Success | `target_import_completed_total{result="success"} / target_import_requests_total` |
| Practice Start Success | 成功开始会话 / 开始会话请求 |
| Practice Turn Latency | 提交回答到返回下一步动作的延迟 |
| Report Ready Success | `reports_ready_total{result="success"} / reports_requested_total` |
| Report Ready Latency | 从 complete 到 report ready 的耗时 |
| Privacy Completion | 隐私请求在目标时间内完成的比例 |

## 10.2 推荐 SLO（P0 / P1）

| 目标 | SLO |
|---|---|
| JD 导入成功率 | >= 98% / 7d |
| 开始会话成功率 | >= 99% / 7d |
| 题间响应 P95 | <= 4s |
| 报告生成成功率 | >= 97% / 7d |
| 报告生成 P95 | <= 45s |
| 简历定制成功率 | >= 95% / 7d |
| 删除 / 导出完成时效 | 99% 在 24h 内完成 |

---

## 11. 告警策略

### 11.1 P1（高优先级，影响核心主链路）

- `practice_turn_latency_p95 > 6s` 持续 10 分钟
- `report_generation_success_rate < 95%` 持续 15 分钟
- `target_import_success_rate < 95%` 持续 15 分钟
- `async_job_queue_depth{job_type="report_generate"} > 100` 持续 10 分钟
- `ai_output_validation_failures_total` 5 分钟内激增 3 倍

### 11.2 P2（中优先级）

- `resume_tailor_duration_p95 > 60s`
- `stale_source_rate > 5%`
- `db_pool_wait_count_total` 短时激增
- `ai_fallback_total` 异常上升

### 11.3 P3（低优先级 / 工作时间处理）

- 单个 route 4xx 飙升
- 某语言下报告投诉率升高
- 成本日环比异常

---

## 12. Dashboard 建议

### 12.1 业务漏斗 Dashboard

展示：

- 新用户数
- 首次价值达成率
- 练习启动率
- 会话完成率
- 报告查看率
- 7 日复练率
- debrief 回流率

### 12.2 API / Session Health Dashboard

展示：

- API QPS
- route 级错误率
- practice turn latency P50 / P95 / P99
- 活跃会话数
- Redis 队列深度

### 12.3 Report Pipeline Dashboard

展示：

- report job enqueue rate
- ready success rate
- average generation duration
- failed by error_code
- schema validation failure

### 12.4 AI Cost & Quality Dashboard

展示：

- provider / model family 使用量
- token 消耗
- cost / day
- cost / session
- fallback rate
- 追问相关率
- 报告空泛率
- 异常高分率

### 12.5 Privacy & Compliance Dashboard

展示：

- 导出请求数
- 删除请求数
- 完成时长
- 失败原因分布
- 审计事件趋势

---

## 13. Trace 规范

### 13.1 Trace 传播

- 浏览器请求应尽量带上 `traceparent`
- API 进程为每个请求创建 span
- 异步 job 通过 job payload 透传 trace context
- AI 调用作为子 span 记录

### 13.2 核心 span 名称建议

- `HTTP POST /api/v1/targets/import`
- `JOB target_import`
- `AI jd_parse`
- `HTTP POST /api/v1/practice/sessions/{id}/events`
- `AI followup_generate`
- `JOB report_generate`
- `AI report_generate`

### 13.3 span attributes 建议

允许：

- `service.name`
- `http.route`
- `job.type`
- `ai.task_type`
- `ai.provider`
- `ai.model_family`
- `result`
- `error.code`

不建议：

- `user.id`
- `session.id`
- 原始 prompt / answer 文本

---

## 14. 数据质量与埋点治理

### 14.1 埋点版本化

每个事件必须至少有：

- `eventName`
- `eventVersion`
- `occurredAt`

### 14.2 埋点校验

上线前需验证：

- 是否有重复事件
- 是否字段缺失
- 是否枚举值漂移
- 是否包含敏感数据
- 是否前端、后端双发导致计数翻倍

### 14.3 指标回溯来源

关键指标应能回溯到：

- 原始 analytics event
- API / job metrics
- 业务表聚合
- 质量评估结果

避免只在某一个系统里看到“数字”，却无法解释来源。

---

## 15. 最低上线门槛

P0 首次外部上线前，以下必须齐备：

- [ ] API route 级 request / error / latency 指标
- [ ] queue depth 与 job duration 指标
- [ ] AI token / latency / cost 指标
- [ ] 主漏斗产品埋点
- [ ] 关键告警规则
- [ ] 1 个业务 dashboard
- [ ] 1 个 API / Worker dashboard
- [ ] trace 能串起至少一条 report_generate 链路
