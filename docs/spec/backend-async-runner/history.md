# Backend Async Runner History

> **版本**: 1.16
> **状态**: active
> **更新日期**: 2026-07-13

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-13 | 1.16 | OPENAPI-002 runner contraction：消费 B3 2.15 7-job generated contract，删除 TargetJob refresh handler/registration/lease loop/low queue assignment，runtime 收敛为 6 handlers；独立 `source_records` persistence 保留。 | 001-internal-job-outbox-runner Phase 7 + event-and-outbox-contract/001 Phase 9 |
| 2026-07-13 | 1.15 | Supersede 1.12-1.14中的report durable/job max4口径：`GenerateReport`单次动作内持有initial+最多3次retry与10s/20s/40s等待，返回即销毁且新动作清零；`async_jobs.attempts/max_attempts`只作基础设施lease/finalize，不再编码产品重试。 | 001-internal-job-outbox-runner Phase 6 + backend-review/001 |
| 2026-07-13 | 1.14 | Phase 5闭环当前报告目标：分离business/infra退避，report job/provider上限4，kernel finalize全局generation fencing，report与`resume_tailor`直接事务fencing；删除review-store重复lease/reaper owner并用structure negative test锁定零回流。 | 001-internal-job-outbox-runner + backend-review/001 |
| 2026-07-13 | 1.13 | L2：report job显式max_attempts4；以claimed attempts作为lease generation，kernel finalize与当前report直接事务校验running+generation，阻止旧worker覆盖job/report副作用。 | 001-internal-job-outbox-runner + backend-review/001 |
| 2026-07-13 | 1.12 | 分离business async `[10s,20s,40s,80s]` cap80与outbox/infra `[30s,2m,10m,1h,6h]`；report durable max4只使用10s/20s/40s。 | 001-internal-job-outbox-runner + backend-review/001 |
| 2026-07-10 | 1.11 | 删除 targetjob test-only runtime、重复 async job 类型/SQL 与 handler adapter；五个业务 handler 和 cmd/api 场景直接使用 runner kernel。 | 001-internal-job-outbox-runner |
| 2026-07-10 | 1.10 | 技术债口径清理：active spec 与 completed plan 区分实施前 review runner 基线和当前 `runner.Runtime` + `review.GenerateHandler` owner 事实。 | 001-internal-job-outbox-runner |
| 2026-07-06 | 1.7 | 将 active spec 改为当前 runtime 合同表述：7 个可执行 handler、`privacy_export` contract-only、单一 kernel lifecycle、current handler owner map 和 generic out-of-scope runner gate。 | product-scope/001-core-loop-module-pruning Phase 6.30 |
| 2026-06-29 | 1.6 | 同步当前 runner job_type 事实：当前可执行 handler 集合收敛为 target/import、resume、report、privacy delete、email dispatch 等核心 job。 | product-scope/001-core-loop-module-pruning |
| 2026-05-22 | 1.4 | L2 review fix：补齐 scheduler 防饥饿、handler 返回后 fresh timestamp finalize、`report_generate` 失败走 kernel shared backoff 的验收口径与测试 gate。 | 001-internal-job-outbox-runner |
| 2026-05-22 | 1.3 | L2 completion audit 修正文档事实口径：将 spec / plan 中的旧 runner、旧 mail dispatcher 与 outbox dispatcher 缺失描述明确为 `001` 实施前基线，并补写当前 kernel + `OutboxDispatcher` 已接入的完成态事实，避免 active spec 继续表达旧实现为当前事实。 | 001-internal-job-outbox-runner |
| 2026-05-22 | 1.2 | L2 code review remediation：固化 `cmd/api` production bootstrap 必须挂接 `OutboxDispatcher` 的验收口径；补齐 owner BDD rerun 清单（含 `E2E.P0.011` / `053` / `093`）与 p0-033 live gate 重复运行证据。 | 001-internal-job-outbox-runner |
| 2026-05-22 | 1.1 | 合并 `main` 后按当前 B3/shared jobs 与 backend-jobs-recommendations 实现修订：将 `jd_match_agent_scan` 纳入当前可执行 runner 接管范围，明确 `jd_match_search` 仅为 future-async reserved，补齐 operation matrix、Phase 2/4 checklist、BDD rerun 与 context discovery。 | 001-internal-job-outbox-runner |
| 2026-05-21 | 1.0 | 初始创建：锁定 backend in-process 单一 runtime kernel、统一 lease/retry/dead-letter/reaper/shutdown 协议、落地 B3 outbox dispatcher 缺口、保留「不建独立 worker 进程」语义；P3 接管 `email_dispatch` 从 C1 进程内 channel 迁到 `async_jobs` 行。 | 001-internal-job-outbox-runner |
