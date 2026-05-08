# 001 BDD Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-08

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.014 Home 默认渲染（empty + non-empty + 12+）

- [x] 创建场景目录 `test/scenarios/e2e/p0-014-home-default-render/`，含 `README.md`（§6 baseline + §7 离线限制）
- [x] 准备 fixture variant：`listTargetJobs.json` 至少 3 个 variant（empty / one-job / 12+jobs），按 `mock-contract-suite` 规则配置；通过 `make validate-fixtures`
- [x] 实现 `scripts/setup.sh`（预检 frontend dist + chromium + fixture variant 切换入口）/ `scripts/trigger.sh`（运行 home spec + 三 variant 切换）/ `scripts/verify.sh`（断言 hero/textarea/aux cards/empty state/12-card cap/`updatedAt desc` 排序/topbar 高亮/zh-en 切换/warm-dark-customAccent 切换）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：spec 调用栈 + variant 切换日志 + 截图（baseline + 当前）+ retired-testid grep 0 命中
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.014 行（关联需求 `frontend-home-job-picks-and-parse C-1, C-4`，状态 Ready，automated）

## E2E.P0.015 Paste/Upload/URL → Import → Parse 主路径

- [x] 创建场景目录 `test/scenarios/e2e/p0-015-jd-import-and-parse/`，含 `README.md`
- [x] 准备 fixture variant：`Uploads/createUploadPresign.json` 至少 2 个 variant（target_job_attachment 成功 / 4xx），`importTargetJob.json` 至少 4 个 variant（manual_text 成功 / file 成功 / url 成功 / 422 invalid source），`getTargetJob.json` 至少 3 个 variant（queued → processing → ready 序列模拟 / failed / requirements & hidden signals 富数据）
- [x] 实现 `scripts/setup.sh`（含三种 source variant 切换入口）/ `scripts/trigger.sh`（按 A/B/C 三条路径运行）/ `scripts/verify.sh`（断言 upload 路径先调 `createUploadPresign` 且 `purpose=target_job_attachment`、ImportTargetJobRequest discriminator 三种 type、side-effect 请求均带 `Idempotency-Key`、route 跳 parse、loading 4 步节奏 ≥600ms、preview 字段映射、summary/fitSummary provenance 可追溯、failed UI、JD raw text 0 命中、4xx inline 错误、前端 network/client spy 无 LLM/provider/prompt-registry 调用）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS（三条 source 路径 + failed variant 共 4 子用例）
- [x] 记录验证证据：mockTransport 调用日志 spy + 隐私反查日志 + 4xx 路径截图
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.015 行（关联需求 C-2, C-3, C-6，状态 Ready，automated）

## E2E.P0.016 Parse 编辑 + Confirm → workspace（含 auth pending action）

- [x] 创建场景目录 `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/`，含 `README.md`
- [x] 准备 fixture variant：`updateTargetJob.json` 至少 2 个 variant（成功 / 4xx）；signed-in / signed-out 两种状态切换入口
- [x] 实现 `scripts/setup.sh`（含 signed-in/out 切换）/ `scripts/trigger.sh`（按 A 已登录 / B 未登录 / C 通用 三子场景运行）/ `scripts/verify.sh`（断言 updateTargetJob body schema 仅含 supplied fields 且不含 hit toggle / summary / fitSummary / hidden signals、`Idempotency-Key` header 存在、route 跳 workspace 携带 5 字段 interviewContext、auth pending action 触发与登录恢复、Re-parse 只重新轮询 `getTargetJob` 不直连 LLM、Cancel 行为、隐私反查）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS（A/B/C 共 ≥3 子用例）
- [x] 记录验证证据：updateTargetJob request body 截取 + auth pending action 路径流 + interviewContext 字段集合断言
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.016 行（关联需求 C-5, C-7，状态 Ready，automated）

## E2E.P0.017 jd_match P1 Placeholder Smoke

- [x] 创建场景目录 `test/scenarios/e2e/p0-017-jd-match-placeholder/`，含 `README.md`
- [x] 准备：无需新 fixture（placeholder 不消费数据）；脚本入口校验 D1 generated client 不会被 jd_match 屏触发额外 API 调用
- [x] 实现 `scripts/setup.sh` / `scripts/trigger.sh`（通过 TopBar 与 home aux card 双入口进入 jd_match）/ `scripts/verify.sh`（断言 hero / profile chip / 三 tab / placeholder testid 全命中、TopBar `topbar-nav-jd_match` 高亮、旧业务 testid grep 0 命中、i18n zh/en 切换、warm-dark-customAccent 切换、mobile 不溢出、generated client 调用次数为 0 或仅限于 `getMe` 等已存在 D1 调用）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：retired-testid grep 0 命中日志 + generated client spy + mobile 截图
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.017 行（关联需求 C-8，状态 Ready，automated）

## 整体 Regression（Phase 6 收口）

- [x] D1+D2+D3 Regression 重跑：`E2E.P0.001 / 002 / 004 / 005 / 006` setup→trigger→verify→cleanup 全部 PASS（D2 视觉系统不被 home/parse/jd_match 改动破坏）
- [x] `pnpm --filter @easyinterview/frontend test` 全量 Vitest PASS（含本 plan 新增测试文件）
- [x] `pnpm --filter @easyinterview/frontend test:pixel-parity` 在 D2/D3 现有 21 spec × 2 viewport = 42 项基础上累加 home/parse/jd_match 新增 spec，总数全 PASS，并确认 parse loading footer 与 `ui-design` 源级结构一致但无前端 LLM/provider 请求
- [x] `pnpm --filter @easyinterview/frontend typecheck` + `pnpm --filter @easyinterview/frontend build` + `make build` 全 PASS
- [x] `make docs-check` zero drift；`/sync-doc-index --fix-index` post-fix zero drift；`check_md_links` 双 OK
