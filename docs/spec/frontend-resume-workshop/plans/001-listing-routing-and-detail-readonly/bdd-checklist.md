# 001 BDD Checklist

> **版本**: 2.5
> **状态**: completed
> **更新日期**: 2026-07-14

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.036 flat list + auth boundary

- [x] 场景目录为 `test/scenarios/e2e/p0-036-resume-flat-list-auth-boundary/`，含 `README.md`、`data/seed-input.md`、`data/expected-outcome.md` 与四段脚本。
- [x] `trigger.sh` 执行 `src/app/scenarios/p0-036-resume-flat-list-auth-boundary.test.tsx`，覆盖未登录 no-fetch、已登录 flat table、row open navigation 和 out-of-scope route testid negative。
- [x] `verify.sh` 检查 trigger log 的 4 tests passed、测试文件 marker、out-of-scope testid negative 和 fallback-text negative。
- [x] 场景在 `test/scenarios/e2e/INDEX.md` 登记为 Ready / automated，描述当前 flat list contract。
- [x] P0.036 或 focused substitute gate 覆盖重复 upload/paste CTA absent、`archiveResume` 成功隐藏 row、失败保留 row 并显示错误。<!-- verified: 2026-07-07 method=focused-substitute tests=ResumeListView.test.tsx -->
- [x] P0.036 不再安装仅供自测的 prototype toast capture；正式 Resume Workshop source 对该 bridge 保持 zero reference。
  <!-- verified: 2026-07-10 method=orphan-resume-toast-removal evidence="Scoped source gate and literal search are clean; P0.036 setup/trigger/verify/cleanup passes 4/4, and Resume Workshop plus P0.036 passes 20 files/110 tests." -->

## E2E.P0.037 read-only source-format detail + removed actions + 404 fallback

- [x] 场景目录为 `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/`，含 `README.md`、`data/seed-input.md`、`data/expected-outcome.md` 与四段脚本。
- [x] `trigger.sh` 执行 `src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx`，覆盖 read-only source-format detail、pending PDF upload polling to source page stack、failed-with-snapshot PDF upload single-fetch、generic-name / raw-first-line name negative、out-of-scope rewrites tab ignored、removed action negative 和 not-found fallback。 <!-- verified: 2026-07-08 method=scenario scenario=E2E.P0.037 -->
- [x] `verify.sh` 检查 trigger log 的当前测试数 passed、测试文件 marker、out-of-scope testid negative、fallback-text negative、generic upload/paste name negative、raw-first-line name negative、pending upload polling marker、failed-with-snapshot single-fetch marker 和 error-code no-echo。 <!-- verified: 2026-07-07 method=scenario scenario=E2E.P0.037 -->
- [x] 场景在 `test/scenarios/e2e/INDEX.md` 登记为 Ready / automated，描述当前 read-only source-format detail contract。
- [x] P0.037 或 focused substitute gate 覆盖 pending 等待动画、ready Markdown DOM 渲染、failed-empty 失败态和 failed-with-snapshot single-fetch 终态。<!-- verified: 2026-07-07 method=focused-substitute tests=ResumeDetailView.test.tsx,ResumePreviewTab.test.tsx -->
- [x] P0.037 或 focused substitute gate 覆盖 upload PDF 渲染从上到下平铺的 PDF page stack，读取 `/resumes/{resumeId}/source`，不渲染 browser PDF viewer toolbar / native viewer shell；paste / Markdown / TXT 继续使用 Markdown engine。<!-- verified: 2026-07-08 method=focused-substitute tests=ResumePreviewTab.test.tsx,PdfPageStackPreview.test.tsx,pixel-parity/resume-workshop.spec.ts pdf page-stack -->
- [x] P0.037 或 focused substitute gate 覆盖 Markdown body card 不额外注入 `displayName` / header 名称 / summary / source metadata，并覆盖 PDF/Markdown 共用阅读背景板和 Markdown page surface。验证: `ResumePreviewTab.test.tsx`、`ResumeWorkshopCssParity.test.ts`、`frontend/tests/pixel-parity/resume-workshop.spec.ts`。<!-- verified: 2026-07-08 method=focused-substitute tests=ResumePreviewTab.test.tsx,ResumeWorkshopCssParity.test.ts,pixel-parity/resume-workshop.spec.ts desktop/mobile -->
- [x] P0.037 trigger 记录 stderr，verify 拒绝未被 `act(...)` 接管的 React update warning；failed-with-snapshot 单次请求用例 warning-free。
  <!-- verified: 2026-07-10 method=p0-037-async-test-lifecycle evidence="Final setup/trigger/verify/cleanup passes 6/6 with combined stdout/stderr evidence and zero unwrapped-update marker; full frontend passes 138 files/839 tests with zero React update warning." -->

## Phase 19 summary contract and request identity hardening

- [x] P0.036 seed/fixture 外层保持 `PaginatedResume`，仅把 `items` 改为 `ResumeSummary[]`；exact-key gate 只允许 `id,title,displayName,language,sourceType,parseStatus,summaryHeadline,hasReadableContent,updatedAt`，并断言所有锁定详情字段 absent。
- [x] P0.036 StrictMode 用例通过底层 transport spy 断言相同已认证 list request 恰好 1 次实际 transport；第一次 reject 后 retry 发起第 2 个新 transport并成功。
- [x] P0.037 ready detail 用例通过底层 transport spy 断言相同 `getResume(resumeId)` 初始 request 恰好 1 次；完整正文只能由 `getResume` fixture 提供，list summary 不含正文。
- [x] P0.037 pending 用例断言后续轮询只在前一个 request settle 后发生，failed/ready/已有正文停止轮询；第一次 reject 后 retry 发起新的 transport并成功。
- [x] 两个 verify 脚本检查 exact transport marker、retry marker、summary forbidden-field marker，并继续拒绝 no-op / skip / React unwrapped update warning。
- [x] 执行 P0.036 与 P0.037 `setup → trigger → verify → cleanup` 全 PASS，并把真实日志、transport 次数与 fixture key diff 记录到 checklist 证据。
  <!-- verified: 2026-07-14 evidence="Fresh P0.036/P0.037 wrappers PASS; list items expose exactly 9 summary fields, detail content comes only from getResume, transport/retry/poll markers are exact, and no no-op/skip/React warning is present." -->
