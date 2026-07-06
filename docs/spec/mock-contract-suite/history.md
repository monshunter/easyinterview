# Mock Contract Suite History

> **版本**: 1.11
> **状态**: active
> **更新日期**: 2026-07-06

## 1 修订记录

| 日期 | 版本 | 变更 | 关联计划 |
|------|------|------|----------|
| 2026-07-06 | 1.11 | 对齐 product-scope D-17 / D-20 / D-22 后的当前 B2 truth：mock fixture coverage 从历史 48/60 口径收敛为当前 10 tag / 35 operation；JobMatch、Profile、Debriefs 与简历版本/suggestion 旧 operation 不再作为正向 fixture coverage 或 mock response baseline。 | product-scope/001-core-loop-module-pruning Phase 6 |
| 2026-06-13 | 1.10 | product-scope D-17/D-20 后同步 JobMatch 删除与简历扁平化计数，保留 12 tag / 48 operation 过渡口径；该口径已在 v1.11 被 D-22 当前 truth 覆盖。 | product-scope/001-core-loop-module-pruning |
| 2026-05-28 | 1.9 | 对齐 B2 D-25 Auth single-entry profile completion：mock contract fixture coverage 从 59 operation 升到 60 operation，承接 `completeMyProfile` fixture、`UserContext.profileCompletionRequired` 和单入口邮箱验证码登录契约；§2.1 / §6 C-1 与 `openapi/fixtures/README.md` 计数已同步。 | backend-auth/001 Phase 8 + frontend-shell/001 Phase 9 |
| 2026-05-22 | 1.8 | 收窄 retired-token gate：继续拦截独立 `/voice` route / `Voice` tag，但允许 practice-voice-mvp 拥有的 `createPracticeVoiceTurn`、`/practice/sessions/{sessionId}/voice-turns` 与 `PracticeVoiceTurn*` generated artifacts；`lint-mock-contract` 与 repo-wide `make lint` 通过。 | 001-fixture-backed-mock-runtime Phase 6 |
| 2026-05-17 | 1.7 | B2 D-20/D-21/D-22 与 backend-resume D-23 additive 落地同步：mock contract fixture coverage 从 55 operation 升到 59 operation，承接 `suggestDebriefQuestions`、`listPracticeSessions`、`createPracticeVoiceTurn` 与 `confirmResumeStructuredMaster` fixtures；§2.1 / §6 C-1 与 `openapi/fixtures/README.md` 计数已同步。 | backend-resume/002-versions-tailor-runs-and-save-v1 Phase 1 |
| 2026-05-12 | 1.6 | B2 D-18 Resume Workshop additive 落地同步：mock contract fixture coverage 从 46 operation 升到 55 operation，承接 `Resumes` tag 新增 9 operation + 多 variant fixtures；§2.1 / §6 C-1 与 `openapi/fixtures/README.md` 计数已同步。 | openapi-v1-contract/004-resume-additive-coverage |
| 2026-05-11 | 1.5 | B2 D-18 Resume Workshop additive 升级声明阶段同步占位：§2.1 / §6 C-1 保留 46 operation 现状，追加 D-18 声明扩到 55 operation 的预告与落地路径引用（openapi-v1-contract/004-resume-additive-coverage）；本 spec 实际 inventory 数字升级跟随 B2 plan 004 落地后同步 1.5 → 1.6。 | openapi-v1-contract/004-resume-additive-coverage（声明阶段，docs-only） |
| 2026-05-10 | 1.4 | 合并 named scenario truth-source remediation 与 frontend Vite dev preview mock wiring 要求，固化后端 mockruntime 与前端 dev preview 两类 gate。 | 001-fixture-backed-mock-runtime |
| 2026-05-10 | 1.3 | 补充 frontend Vite dev preview 默认 fixture-backed mock wiring 要求，解决无真实 backend 时已开发页面不可见的问题。 | 001-fixture-backed-mock-runtime |
| 2026-05-06 | 1.2 | 对齐 backend-runtime-topology：mock runtime out-of-scope 从后台 worker 改为 backend internal runner，避免把独立 worker 当作默认前置。 | backend-runtime-topology/001-worker-consolidation |
| 2026-05-06 | 1.1 | 补充 fixture tag 目录级旧口径拦截要求，覆盖空目录和 Git 不跟踪目录残留。 | 001-fixture-backed-mock-runtime |
| 2026-05-05 | 1.0 | 从 engineering-roadmap S1 派生 fixture-backed mock runtime subject，作为前后端 mock runway owner。 | 001-fixture-backed-mock-runtime |
