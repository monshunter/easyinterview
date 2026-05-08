# Backend TargetJob BDD Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-08

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.010 Text JD import 走完异步解析并可列表 / 详情 / 更新

- [ ] 创建场景目录 `test/scenarios/e2e/p0-010-targetjob-text-import-parse-ready/`，并在 `test/scenarios/e2e/INDEX.md` 添加 `E2E.P0.010` 行（关联需求：backend-targetjob C-1/C-3/C-6/C-7/C-12）
- [ ] 准备测试数据：干净用户邮箱 + cookie jar、合法 `manual_text` JD payload、固定 `targetLanguage`、`Idempotency-Key`、stub `target.import.default` profile（`APP_ENV=test`）、outbox 与 DB 清理脚本
- [ ] 实现 setup / trigger / verify / cleanup：start auth → `POST /targets/import` import manual_text TargetJob → 等 drainer drain → `GET /targets` 验证列表出现该 job → `GET /targets/{id}` 验证 `analysis_status='ready'` + 草稿 requirements + provenance 字段 → `PATCH /targets/{id}` 验证合法 status / notes update 且不修改 `analysis_status`；同 `Idempotency-Key` 重复 import 验证返回同一 `targetJobId`，DB / outbox 不出现重复 row
- [ ] 断言 outbox 中存在 `target.import.requested` 与 `target.parsed`，且 payload 不含 `raw_jd_text` / 完整 URL / prompt body
- [ ] 执行并通过场景验证，记录 `.test-output/runs/.../E2E.P0.010/result.json` 证据

## E2E.P0.011 URL JD import 守护与抓取

- [ ] 创建场景目录 `test/scenarios/e2e/p0-011-targetjob-url-import-fetch-and-parse/`，并在 `test/scenarios/e2e/INDEX.md` 添加 `E2E.P0.011` 行（关联需求：backend-targetjob C-2/C-3/C-9）
- [ ] 准备测试数据：本地 HTTPS fixture server（合规 JD HTML、超长 body、cross-origin redirect 进入私网、metadata 服务模拟）；非法 URL 集合（私网 IP、链路本地、`http` scheme、超长 body、redirect 进入私网）
- [ ] 实现 setup / trigger / verify / cleanup：start auth → `POST /targets/import` import URL → drainer 抓取并解析 → `GET /targets/{id}` 验证 `target_job_sources.url` 为规范化 URL、`target_job_sources.snapshot_text` 为去密正文、`fetched_at` / `freshness_status='fresh'`；`target.parsed` 与 `source_refresh` 占位 job 写入
- [ ] 断言非法目标：所有非法 URL 返回 B1 `TARGET_IMPORT_SOURCE_INVALID` 或 `TARGET_IMPORT_SOURCE_UNAVAILABLE`；事件 / metric label / log / audit 不含完整 URL / query 串 / 内网响应内容
- [ ] 执行并通过场景验证，记录 `.test-output/runs/.../E2E.P0.011/result.json` 证据

## E2E.P0.012 Parse 失败 retryable / non-retryable

- [ ] 创建场景目录 `test/scenarios/e2e/p0-012-targetjob-parse-failure-retryable/`，并在 `test/scenarios/e2e/INDEX.md` 添加 `E2E.P0.012` 行（关联需求：backend-targetjob C-4/C-5/C-10）
- [ ] 准备测试数据：可注入失败的 stub provider（`AI_PROVIDER_TIMEOUT` / `AI_OUTPUT_INVALID` / `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID`）、F3 unsupported / disabled profile 配置 fixture、已存在的 manual_text TargetJob、cookie jar、`Idempotency-Key`
- [ ] 实现 setup / trigger / verify / cleanup：start auth → 注入 `AI_PROVIDER_TIMEOUT` → import → drainer drain → 验证 `target.analysis.failed.retryable=true` 与 `analysis_status='failed'`；切换 stub 到 `AI_OUTPUT_INVALID` → 重新 import → 验证 `retryable=false`；切换 F3 `target.import.parse` 为 disabled / unsupported → 验证 import 启动 / drainer 阶段 fail-closed；切换 A3 缺 secret / config invalid → 验证 `AI_PROVIDER_SECRET_MISSING` / `AI_PROVIDER_CONFIG_INVALID`
- [ ] 断言 error envelope / log / metric label / audit / outbox payload 不含 prompt body、response body、provider secret、`Authorization:` 等敏感模式；失败 TargetJob 的 `target_job_sources` 行保留以便用户重试
- [ ] 执行并通过场景验证，记录 `.test-output/runs/.../E2E.P0.012/result.json` 证据

## E2E.P0.013 Manual form import ready 直达

- [ ] 创建场景目录 `test/scenarios/e2e/p0-013-targetjob-manual-form-ready/`，并在 `test/scenarios/e2e/INDEX.md` 添加 `E2E.P0.013` 行（关联需求：backend-targetjob C-3/C-6/C-9/C-11/C-13）
- [ ] 准备测试数据：干净用户邮箱 + cookie jar、合法 `manual_form` payload（title / company / rawDescription）、固定 `targetLanguage`、`Idempotency-Key`、outbox 与 async_jobs 清理脚本；A3 / F3 可设置为不可用以证明该路径不依赖 AI parse
- [ ] 实现 setup / trigger / verify / cleanup：start auth → `POST /targets/import` import manual_form TargetJob → 直接 `GET /targets/{id}` 验证 `analysis_status='ready'` + 至少 1 条 `must_have` draft requirement → `GET /targets` 验证列表可见；同 `Idempotency-Key` 重复 import 验证返回同一 `targetJobId`
- [ ] 断言 202 响应保持 B2 `TargetJobWithJob` 形态，`job.jobType='target_import'` 且 `job.status='succeeded'`，但不存在待 drainer 消费的 `target_import` async job
- [ ] 断言 outbox 不包含 `target.import.requested` / `target.parsed`，event / metric label / log / audit 不含 `raw_jd_text` / prompt body / response body
- [ ] 执行并通过场景验证，记录 `.test-output/runs/.../E2E.P0.013/result.json` 证据
