# 002 BDD Checklist

> **版本**: 1.1
> **状态**: completed
> **更新日期**: 2026-05-23

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.081 resume create flow upload / paste / guided happy + presign + register + parse polling + IK + retired negative + UI parity

- [x] 创建场景目录 `test/scenarios/e2e/p0-081-resume-create-flow-upload-paste-guided-happy/`，含 `README.md`（§6 baseline + §7 离线限制 + §8 mock-first 标注）+ `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 fixture variant：`Uploads/createUploadPresign.json default` + `Resumes/registerResume.json default / paste-text / guided-answers` + `Resumes/getResume.json default`；mock harness 在 fixture 未覆盖 `parseStatus` 多态时显式声明 deterministic stepping；用户未登录 / 已登录态 + lang fixture
- [x] 实现 `scripts/setup.sh`（A2 dev stack + Vite dev preview fixture-backed + 未登录态 + lang=zh-CN 注入）/ `scripts/trigger.sh`（未登录加载 → 登录恢复 → Upload tab → Paste tab → Guided tab 三路径完成 register + parse polling）/ `scripts/verify.sh`（断言 auth gate no-fetch + ≥ 20 testid + 三 sourceType payload shape 与 fixture 字节对齐 + Idempotency-Key header on presign / register + 无 IK on getResume + ParseFlow 7 step ticker DOM + lang 切换 + Accept-Language header + UI parity desktop/mobile DOM/computed/bounding/screenshot smoke + 隐私 grep + 旧入口 grep + prototype import grep）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-081-resume-create-flow-upload-paste-guided-happy/trigger.log` + verify 输出 + IK header spy + auth no-fetch request spy + UI parity desktop/mobile artifacts + axe a11y report + retired-testid grep 0 + prototype import grep 0 + mock harness deterministic stepping 显式声明
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.081 行（关联需求 `frontend-resume-workshop C-10, C-7, C-8, C-9`，状态 Ready，automated）

## E2E.P0.082 resume create flow parsing failure / timeout / cancel-and-return + retry

- [x] 创建场景目录 `test/scenarios/e2e/p0-082-resume-create-flow-parsing-failure-and-retry/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 fixture / mock variant：`Resumes/registerResume.json default` + `Resumes/getResume.json default`；mock harness 支持 attempt-aware `parseStatus` 序列模拟（`queued → generating → failed (errorCode='AI_TIMEOUT_RETRYABLE')` / `queued → generating × 8 → timeout`）；用户已登录 lang=zh-CN
- [x] 实现 `scripts/setup.sh`（A2 dev stack + fixture-backed + 登录态 + mock attempt-aware harness 启用）/ `scripts/trigger.sh`（Paste 提交 → ParseFlow failed → 重试解析成功 / Guided 提交 → ParseFlow generating × 8 → parseTimeout / 任一 ParseFlow 阶段点击取消并返回修改 → 验证输入保留）/ `scripts/verify.sh`（断言 failed-state DOM + errorCode 映射 + 重试 button 行为 + cancel-and-return state preservation + parseTimeout fallback + 隐私 DOM grep 不渲染 parsedTextSnapshot）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-082-resume-create-flow-parsing-failure-and-retry/trigger.log` + verify 输出 + mock harness attempt sequence dump + failed-state 截图 + cancel-state preservation DOM diff + 隐私 grep 0
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.082 行（关联需求 `frontend-resume-workshop C-10, C-8`，状态 Ready，automated）

## E2E.P0.083 resume create flow preview confirm + 409 fallback + 422 + Home / Workspace CTA handoff + auth pending action

- [x] 创建场景目录 `test/scenarios/e2e/p0-083-resume-create-flow-preview-confirm-and-cta-handoff/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [x] 准备 fixture variant：`Uploads/createUploadPresign.json default` + `Resumes/registerResume.json default` + `Resumes/getResume.json default` + `Resumes/confirmResumeStructuredMaster.json default / idempotency-replay / already-exists-409 / validation-422` + `Resumes/listResumeVersions.json master-only` + plan 001 `Resumes/listResumes.json default`；用户未登录 → 登录态 / lang=zh-CN
- [x] 实现 `scripts/setup.sh`（A2 dev stack + fixture-backed + 未登录态 + listResumeVersions master-only fixture 准备）/ `scripts/trigger.sh`（Home `1 分钟创建` 未登录 CTA → pendingAction → 登录恢复 → CreateFlow upload → register → parse ready → PreviewConfirm → confirm happy → list 验证；同 IK 二次 confirm replay；新 IK 409 → fallback nav；422 → inline；Workspace `WorkspaceMissingResumeState` CTA → pendingAction → 登录恢复 → CreateFlow）/ `scripts/verify.sh`（断言 confirmResumeStructuredMaster IK header + 三态错误映射 + listResumeVersions 路径 + nav 行为 + pendingAction 参数集合 + 隐私 grep + UI parity）/ `scripts/cleanup.sh`
- [x] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [x] 记录验证证据：`.test-output/e2e/p0-083-resume-create-flow-preview-confirm-and-cta-handoff/trigger.log` + verify 输出 + confirm IK header spy + 409 fallback nav trace + 422 inline 截图 + Home CTA pendingAction params dump + Workspace CTA pendingAction params dump + PreviewConfirm UI parity desktop/mobile artifacts + 隐私 grep 0
- [x] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.083 行（关联需求 `frontend-resume-workshop C-10, C-8, C-9`，状态 Ready，automated）
- [x] 2026-05-23 real-backend gate：P0.081-P0.083 trigger scripts now run `frontendOwners.realApiMode.test.ts`, and verify scripts reject missing real-mode marker / default backend base URL / test-file marker; focused real-mode Vitest PASS.
