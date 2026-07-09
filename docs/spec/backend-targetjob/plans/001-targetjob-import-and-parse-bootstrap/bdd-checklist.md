# Backend TargetJob BDD Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-09

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.010 Text JD import 走完异步解析并可列表 / 详情 / 更新

- [x] 创建场景目录 `test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/`，并在 `test/scenarios/e2e/INDEX.md` 添加 `E2E.P0.010` 行（关联需求：backend-targetjob C-1/C-3/C-6/C-7/C-12）
- [x] 准备测试数据：干净用户邮箱 + cookie jar、合法 `manual_text` JD payload、固定 `targetLanguage`、`Idempotency-Key`、stub `target.import.default` profile（`APP_ENV=test`）、outbox 与 DB 清理脚本
- [x] 实现 setup / trigger / verify / cleanup：start auth → `POST /targets/import` import manual_text TargetJob → 等 drainer drain → `GET /targets` 验证列表出现该 job → `GET /targets/{id}` 验证 `analysis_status='ready'` + 草稿 requirements + provenance 字段 → `PATCH /targets/{id}` 验证合法 status / notes update 且不修改 `analysis_status`；同 `Idempotency-Key` 重复 import 验证返回同一 `targetJobId`，DB / outbox 不出现重复 row
- [x] 断言 outbox 中存在 `target.import.requested` 与 `target.parsed`，且 payload 不含 `raw_jd_text` / 完整 URL / prompt body
- [x] 执行并通过场景验证，记录 `.test-output/runs/.../E2E.P0.010/result.json` 证据
  <!-- verified: 2026-05-08 method=cmd-api-http run=targetjob-http-20260508 validBddEvidence=true -->
- [x] BUG-0146 C-16 regression: valid real-provider manual_text JD without company name parses ready, uses fallback company display, renders authenticated `/parse` without `JD 解析失败`, and records `ai_task_runs` jd_parse evidence
  <!-- verified: 2026-07-09 method=local-real-provider-browser targetJobId=019f44a1-b43e-754f-ba0b-3cd9ed11ce1f evidence=analysisStatus=ready companyName=未提供 title=AI应用技术负责人 browserRoute=route-parse aiTaskRun=success -->

## E2E.P0.011 URL JD import 守护与抓取

- [x] 创建场景目录 `test/scenarios/e2e/p0-011-targetjob-url-import-fetch-and-parse/`，并在 `test/scenarios/e2e/INDEX.md` 添加 `E2E.P0.011` 行（关联需求：backend-targetjob C-2/C-3/C-9）
- [x] 准备测试数据：本地 HTTPS fixture server（合规 JD HTML、超长 body、cross-origin redirect 进入私网、metadata 服务模拟）；非法 URL 集合（私网 IP、链路本地、`http` scheme、超长 body、redirect 进入私网）
- [x] 实现 setup / trigger / verify / cleanup：start auth → `POST /targets/import` import URL → drainer 抓取并解析 → `GET /targets/{id}` 验证 `target_job_sources.url` 为规范化 URL、`target_job_sources.snapshot_text` 为去密正文、`fetched_at` / `freshness_status='fresh'`；`target.parsed` 与 `source_refresh` 占位 job 写入
- [x] 断言非法目标：所有非法 URL 返回 B1 `TARGET_IMPORT_SOURCE_INVALID` 或 `TARGET_IMPORT_SOURCE_UNAVAILABLE`；事件 / metric label / log / audit 不含完整 URL / query 串 / 内网响应内容
- [x] 执行并通过场景验证，记录 `.test-output/runs/.../E2E.P0.011/result.json` 证据
  <!-- verified: 2026-05-08 method=cmd-api-http run=targetjob-http-20260508 validBddEvidence=true -->

## E2E.P0.012 Parse 失败 retryable / non-retryable

- [x] 创建场景目录 `test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/`，并在 `test/scenarios/e2e/INDEX.md` 添加 `E2E.P0.012` 行（关联需求：backend-targetjob C-4/C-5/C-10）
- [x] 准备测试数据：可注入失败的 stub provider（`AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID`）、F3 unsupported / disabled profile 配置 fixture、已存在的 manual_text TargetJob、cookie jar、`Idempotency-Key`
- [x] 实现 setup / trigger / verify / cleanup：start auth → 注入 `AI_PROVIDER_TIMEOUT` → import → drainer drain → 验证 `target.analysis.failed.retryable=true`、`GET /targets/{id}` 返回 404 且 `GET /targets` 不含失败 job；切换 stub 到 `AI_OUTPUT_INVALID` → 重新 import → 验证 `retryable=false` 与同样的不可见资产语义；切换 F3 `target.import.parse` 为 disabled / unsupported → 验证 import 启动 / drainer 阶段 fail-closed；切换 A3 缺 secret / config invalid → 验证 `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID`
- [x] 断言 error envelope / log / metric label / audit / outbox payload 不含 prompt body、response body、provider secret、`Authorization:` 等敏感模式；失败 TargetJob / source / raw JD 不作为可继续规划资产持久化，用户重试必须重新 import
- [x] 执行并通过场景验证，记录 `.test-output/runs/.../E2E.P0.012/result.json` 证据
  <!-- verified: 2026-05-08 method=cmd-api-http run=targetjob-http-20260508 validBddEvidence=true -->
- [x] Revision 2026-07-09 trigger covers parse-failure admission: failed imports emit `target.analysis.failed` but `GET /targets/{id}` is 404 and `GET /targets` excludes the failed job.
- [x] Revision 2026-07-09 verify covers failed TargetJob deletion / no dirty interview-list admission / no prompt-response or raw JD leakage.

## E2E.P0.013 Manual form import ready 直达

- [x] 创建场景目录 `test/scenarios/e2e/p0-013-targetjob-manual-form-ready/`，并在 `test/scenarios/e2e/INDEX.md` 添加 `E2E.P0.013` 行（关联需求：backend-targetjob C-3/C-6/C-9/C-11/C-13）
- [x] 准备测试数据：干净用户邮箱 + cookie jar、合法 `manual_form` payload（title / company / rawDescription）、固定 `targetLanguage`、`Idempotency-Key`、outbox 与 async_jobs 清理脚本；A3 / F3 可设置为不可用以证明该路径不依赖 AI parse
- [x] 实现 setup / trigger / verify / cleanup：start auth → `POST /targets/import` import manual_form TargetJob → 直接 `GET /targets/{id}` 验证 `analysis_status='ready'` + 至少 1 条 `must_have` draft requirement → `GET /targets` 验证列表可见；同 `Idempotency-Key` 重复 import 验证返回同一 `targetJobId`
- [x] 断言 202 响应保持 B2 `TargetJobWithJob` 形态，`job.jobType='target_import'` 且 `job.status='succeeded'`，但不存在待 drainer 消费的 `target_import` async job
- [x] 断言 outbox 不包含 `target.import.requested` / `target.parsed`，event / metric label / log / audit 不含 `raw_jd_text` / prompt body / response body
- [x] 执行并通过场景验证，记录 `.test-output/runs/.../E2E.P0.013/result.json` 证据
  <!-- verified: 2026-05-08 method=cmd-api-http run=targetjob-http-20260508 validBddEvidence=true -->
