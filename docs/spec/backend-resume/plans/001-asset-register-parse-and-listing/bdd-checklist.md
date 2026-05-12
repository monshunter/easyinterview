# 001 BDD Checklist

> **版本**: 1.1
> **状态**: active
> **更新日期**: 2026-05-12

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.034 resume register + get + list

- [ ] 创建场景目录 `test/scenarios/e2e/p0-034-resume-register-and-list/`，含 `README.md`（§6 baseline + §7 离线限制）+ `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 B2 fixture scenario + 测试数据：`registerResume` `default` / `paste-text` / `guided-answers`、`getResume` `default` / `not-found`、`listResumes` `default` / `empty` / `paginated`；2 个测试用户（A / B）；用户 A 通过 backend-upload createUploadPresign (`purpose=resume`) + PUT 取得 fileObjectId；准备 25 个 resume_asset 用于 pagination（混合 sourceType）
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + backend-upload 拉起 + 用户登录 + presign upload + 批量 register）/ `scripts/trigger.sh`（依序触发 A1/A2/A3/B1/C1/C2/D1）/ `scripts/verify.sh`（断言 DB state + cross-user 404 + cursor 分页边界 + 精确 fixture scenario 字节比对 + validation 直接断言 + 隐私 grep + 旧口径 grep）/ `scripts/cleanup.sh`（清理 file_objects + resume_assets + 用户）
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-034-resume-register-and-list/trigger.log` + verify 输出 + DB state machine 轨迹 + register/get/list fixture byte diff 0 + validation error direct assertion + 隐私 grep 0 命中 + cross-user 404 验证 + cursor 序稳定性
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.034 行（关联需求 `backend-resume C-1, C-2, C-5, C-6, C-7, C-8`，状态 Ready，automated）

## E2E.P0.035 resume.parse async job lifecycle

- [ ] 创建场景目录 `test/scenarios/e2e/p0-035-resume-parse-async-job-lifecycle/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture + stub provider：A3 AIClient stub provider 3 response variant（success JSON / output_invalid / timeout）；backend internal runner 启动并注册 resume.parse handler；F3 `resume.parse` feature_key prompt / rubric / profile 就位
- [ ] 实现 `scripts/setup.sh`（runner + stub provider variant 注入 + 用户登录 + register 3 个 resume_asset 对应三 variant）/ `scripts/trigger.sh`（runner 拉队列 + 等待解析完成 / 失败 / 重试）/ `scripts/verify.sh`（断言 parse_status 状态机 + `parsed_text_snapshot` + Preview Confirm 前 `resume_versions` 不新增正式主版本 + ai_task_runs 多行 + outbox event 写入 + dispatcher publish + 隐私 grep + 旧口径 grep）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-035-resume-parse-async-job-lifecycle/trigger.log` + verify 输出 + DB parse_status 转换轨迹 + `resume_versions` count unchanged before Preview Confirm + ai_task_runs 行 dump + outbox_events 行 dump（PII grep 0 命中）+ stub provider call log
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.035 行（关联需求 `backend-resume C-3, C-4, C-13`，状态 Ready，automated）
