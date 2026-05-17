# 003 BDD Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-17

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.084 resume branch flow three seedStrategy + IK + 422 + cross-user + UI parity

- [ ] 创建场景目录 `test/scenarios/e2e/p0-084-resume-branch-flow-three-seed-strategies/`，含 `README.md`（§6 baseline + §7 离线限制 + §8 mock-first 标注）+ `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`Resumes/listResumes.json default` + `Resumes/listResumeVersions.json with-targeted-branches` + `Resumes/branchResumeVersion.json default / copy-master-sync / blank-sync / ai-select-202-with-job / idempotent-replay / validation-error-422`；用户未登录 / 已登录 lang fixture
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + Vite dev preview fixture-backed + 未登录态 + lang=zh-CN 注入）/ `scripts/trigger.sh`（未登录加载 → 登录恢复 → BranchFlow 三 seedStrategy 提交 + replay + 422 + cross-user + lang 切换）/ `scripts/verify.sh`（断言 auth gate no-fetch + ≥ 15 testid + 三 seed payload shape + IK header on branchResumeVersion + nav target 三态 + 422 inline + 404 parent/targetJob toast + lang 切换 + Accept-Language header + UI parity desktop/mobile DOM/computed/bounding/screenshot smoke + 隐私 grep + retired tailor mode grep + 旧入口 grep + prototype import grep）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-084-resume-branch-flow-three-seed-strategies/trigger.log` + verify 输出 + IK header spy + nav target trace + 422 inline 截图 + UI parity desktop/mobile artifacts + axe a11y report + retired-testid grep 0 + retired tailor mode grep 0 + prototype import grep 0
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.084 行（关联需求 `frontend-resume-workshop C-11, C-8, C-9`，状态 Ready，automated）

## E2E.P0.085 resume rewrites tab tailor run polling + 重新运行改写 + ready/failed/timeout

- [ ] 创建场景目录 `test/scenarios/e2e/p0-085-resume-rewrites-tab-tailor-run-polling/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`Resumes/branchResumeVersion.json ai-select-202-with-job` + `ResumeTailor/requestResumeTailor.json default` + `ResumeTailor/getResumeTailorRun.json default` + `Resumes/getResumeVersion.json targeted-with-suggestions`；mock harness 支持 attempt-aware tailor run status stepping；用户已登录 lang=zh-CN
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + fixture-backed + 已登录 + mock attempt-aware harness 启用）/ `scripts/trigger.sh`（ai_select branch 提交 → Rewrites Tab polling banner → ready → 重新运行改写 mode=gap_review → failed → 重试 → timeout / 切换 Edit / Rewrites tab → unmount 验证 cleanup）/ `scripts/verify.sh`（断言 polling banner 渲染 + getResumeTailorRun 无 IK header + requestResumeTailor 含 IK + suggestions[] 计数派生 + failed banner errorCode 映射 enum + timeout fallback + cleanup network sniff 0 后续 polling + UI parity + 隐私 grep + retired tailor mode grep）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-085-resume-rewrites-tab-tailor-run-polling/trigger.log` + verify 输出 + mock harness attempt sequence dump + IK header spy on requestResumeTailor + no-IK on getResumeTailorRun + failed banner 截图 + timeout banner 截图 + suggestions count 派生 + UI parity desktop/mobile artifacts + axe a11y report + cleanup network sniff log + 隐私 grep 0
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.085 行（关联需求 `frontend-resume-workshop C-11, C-8`，状态 Ready，automated）

## E2E.P0.086 resume suggestion accept / reject / manual edit + updateResumeVersion + 终态状态机

- [ ] 创建场景目录 `test/scenarios/e2e/p0-086-resume-suggestion-accept-reject-edit-and-update-version/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：`Resumes/getResumeVersion.json targeted-with-suggestions` + `Resumes/acceptResumeTailorSuggestion.json default / already-decided-409` + `Resumes/rejectResumeTailorSuggestion.json default / already-decided-409` + `Resumes/updateResumeVersion.json default / validation-error-422`；fixture scenario 名称与 error envelope 必须先由 backend-resume/002 Phase 8 收敛，若仍为 `conflict-409` + `TARGET_INVALID_STATE_TRANSITION`，本场景保持 blocked；用户 A 已登录 lang=zh-CN；用户 A 拥有 TARGETED 版本 `v1` + 5 pending suggestions；用户 B 已登录无版本
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + fixture-backed + 已登录 + 准备 v1 + 5 pending）/ `scripts/trigger.sh`（accept b1 + replay + 二次 accept new IK / b2 inline edit + Save manual edit / b3 reject + replay + 二次 reject new IK / Edit Tab headline+summary 保存 + replay + 422 / 不可编辑字段过滤 / 用户 B 调 accept v1.b1）/ `scripts/verify.sh`（断言 IK header on accept/reject/update + replay 行为 + 409 already-decided 映射 + 422 inline + manual edit 双路径选择正确 + accept 不自动 patch structured_profile DOM 断言 + 不可编辑字段 mapper 拦截 + cross-user 404 + 隐私 grep + retired tailor mode grep + 旧入口 grep + UI parity desktop/mobile）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-086-resume-suggestion-accept-reject-edit-and-update-version/trigger.log` + verify 输出 + IK header spy on 三类 op + 409 already-decided 截图 + 422 inline 截图 + accept 不改 structured_profile DOM diff + manual edit 路径分支 dump + 不可编辑字段 mapper 拦截日志 + cross-user 404 toast 截图 + UI parity artifacts + axe a11y report + 隐私 grep 0
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.086 行（关联需求 `frontend-resume-workshop C-11, C-8`，状态 Ready，automated）

## E2E.P0.087 resume detail exportPDF / copyText 一致性 + 三屏 UI parity + 旧入口与 retired tailor mode 负向

- [ ] 创建场景目录 `test/scenarios/e2e/p0-087-resume-detail-export-copy-consistency-and-parity/`，含 `README.md` + `data/seed-input.md` + `data/expected-outcome.md`
- [ ] 准备 fixture variant：plan 001 `exportResumeVersion.json p0-501-not-available` + plan 003 BranchFlow / RewritesTab / EditTab 主路径 fixture；用户已登录 lang=zh-CN；TARGETED 版本 `v1`
- [ ] 实现 `scripts/setup.sh`（A2 dev stack + fixture-backed + 已登录 + v1 ready）/ `scripts/trigger.sh`（desktop 1440px + mobile 390x844 viewport 切换；进入 BranchFlow / Rewrites Tab / Edit Tab 三屏；点击 Export PDF / 复制纯文本 在 Rewrites + Edit Tab 顶 header；执行 retired grep 与 prototype import grep）/ `scripts/verify.sh`（断言三屏 desktop/mobile DOM anchor + computed style + bounding box + 非空截图 + Export PDF Idempotency-Key header + 501 toast + copyText clipboard fallback + retired tailor mode grep 0 + 旧入口 grep 0 + prototype import grep 0 + axe a11y check + 隐私 grep 0）/ `scripts/cleanup.sh`
- [ ] 执行 `setup → trigger → verify → cleanup` 全 PASS
- [ ] 记录验证证据：`.test-output/e2e/p0-087-resume-detail-export-copy-consistency-and-parity/trigger.log` + verify 输出 + 三屏 desktop/mobile pixel parity artifacts + Export PDF IK header spy + 501 toast 截图 + copyText clipboard fallback 截图 + retired tailor mode grep 0 输出 + 旧入口 grep 0 输出 + prototype import grep 0 输出 + axe a11y report
- [ ] 在 `test/scenarios/e2e/INDEX.md` P0 表追加 P0.087 行（关联需求 `frontend-resume-workshop C-11, C-9, C-8`，状态 Ready，automated）
