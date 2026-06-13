# UX Funnel Downstream Implementation 交付复盘报告

> **日期**: 2026-06-12
> **审查人**: Claude (Fable 5)

## 1 复盘范围与成功证据

本次会话把 2026-06-12 设计层闭环（`product-scope` v2.1 决策 D-14~D-21）对齐到正式前后端实现，按 owner 逐个原地修订 spec/plan/checklist 后经 `/implement` → `/tdd` 落地。本次完成并全量验证通过的下游交付：

1. **frontend-shell（D-16 + D-21）** — spec 1.21→1.22、plan/checklist 001 1.16→1.17 新增 Phase 12：删除 `auth_reset` route 与 `AuthResetScreen`（route key + `/auth/reset` legacy path + SPA fallback 三层归一回 `auth_login`），登录页"忘记密码"改为静态帮助说明；设置页收敛为 `个人资料`/`隐私与数据` 双 tab，删除通知/订阅占位 tab，`登录与安全` 仅展示 `邮箱验证码 · 无密码`；默认主题与无效值 fallback 从 `warm` 改为 `ocean`。附带修复 pixel-parity 离线 golden 预览基础设施（`serve-pixel-parity.mjs` 用本地 react/react-dom 18.3.1 UMD + esbuild JSX 预编译替代 unpkg CDN，关闭 2026-06-12 复盘记录的 CDN 阻断阻点）。
2. **D-17 岗位推荐模块整体删除** — `backend-jobs-recommendations` spec 1.2→2.0 / plan 001 Phase 9 + `frontend-home-job-picks-and-parse` spec 1.12→2.0 / plan 002 Phase 7。跨 OpenAPI（jobmatch tag 12 operation + JobMatch fixtures + 12 tag/48 endpoint）、backend（`internal/jdmatch` 全包 + cmd/api wiring + privacy hook + `auth/identity.go`）、migration（000014 drop 5 表 + registry 行 + enum-sources）、shared（2 事件/2 job_type/schema）、config（prompts/rubrics/evals/ai-profiles + resolved-prompts）、frontend（`screens/jd_match` 全模块 + 一级导航四项收敛 + i18n + devMockClient + 10 场景目录）六层零残留删除。
3. **D-19 报告 CTA 单点收敛** — `frontend-report-dashboard` spec 1.2→1.3 / plan 001 Phase 6：`NextTab` 删除 `report-next-cta-a/b` 重复 CTA 改 footer 引导文案；`QuestionsTab` `加入本轮复练` 从 nav 改为 per-question 本地 `markedForReplay` toggle；CTA 唯一入口收敛到 Header。
4. **workspace pixel-parity auth gate 过时测试修复** — `goToWorkspace` 未 mock 认证 API，未登录点击渲染 `auth-route-gate` 而非 `workspace-empty`，是 frontend-shell Phase 10 auth gate 落地后未更新的过时测试；修复后全量 parity 转绿。

### 成功证据（跨层全绿）

- backend：`cd backend && go test ./... -count=1` → 57 包通过 / 0 失败。
- frontend：`pnpm --filter @easyinterview/frontend test` → 1077/1077；`typecheck`、`build` 通过。
- pixel-parity：`pnpm --filter @easyinterview/frontend test:pixel-parity` → 162 passed / 2 skipped / **0 failed**（删除前基线 22 failed）。
- 契约 / 工具链：`make codegen-check`（提交后 `git diff --exit-code` 清，exit=0）、`make lint`（含 inventory 48 op、validate-fixtures 48、mock-contract、prompt/rubric/config/ai-profile-coverage）、`make docs-check`、`bash migrations/lint.sh` 全过。
- 文档：`sync-doc-index --check` 零漂移；跨层 `rg -i "jobmatch|jd[-_]match"`（openapi/backend/shared/config/frontend）零残留（迁移文件 + 负向断言 + D-17 注释除外）。
- 提交：8 个 commit（frontend-shell 3 + jd-match 4 + report 1 + parity fix 1）。

## 2 会话中的主要阻点/痛点

1. **workspace parity 失败的早期误判**
   - **证据**：D-17 删除提交的 commit message 与工作日志先把 6 个 workspace pixel-parity 失败归为 "frontend-workspace-and-practice/001 的 D-14/D-18 待修订范围"；后经 Explore 深度盘点确认 D-14 枢纽化与 D-18 company_intel 嵌入在 `001-002`（completed）**已实现**，真实根因是 workspace 受 auth gate 保护、测试 `goToWorkspace` 未 mock 认证，是 auth gate 落地后的过时测试。
   - **影响**：一次归类返工 + commit/日志口径需后续会话纠正；若未深查会把"测试维护问题"长期挂在错误 owner 名下。

2. **删除型变更的多个计数型 gate 需同步且部分隐藏**
   - **证据**：D-17 删除触发 ≥5 个 hardcoded 计数 gate 失败：privacy matrix 35→30、offline eval 50→44、registry SnapshotSize/feature_key 13→11、openapi 60→48 endpoint + events schema 18→16、ai-profile-coverage 缺 profile；其中 `seed_baseline_prompt_rubric` 与 `prompt_lint` 的 seed-vs-yaml 校验只看 seed insert，不识别后续 drop migration 的 net state，需在两处 lint 显式增加 "retired by drop migration" 逻辑。
   - **影响**：删除一个模块需要在十余处分散位置同步数字与 gate 逻辑，逐个 `go test` / `make lint` 迭代定位，耗时显著。

3. **codegen-events-check 的 `git diff --exit-code` 在删除提交前必然"失败"**
   - **证据**：`make codegen-check` 删除提交前报错（events.go/jobs.go 等生成物与 HEAD 有 diff），实为已正确 regenerate 但未提交的预期态；提交后 exit=0。
   - **影响**：中途容易误判为 codegen 缺陷而重复排查；需理解该 gate 是"工作树 vs HEAD"而非"yaml vs 生成物 drift"。

4. **pixel-parity golden 预览依赖外网 CDN**
   - **证据**：`ui-design/index.html` 通过 unpkg 拉 react/react-dom/@babel standalone，离线/CDN 阻断环境下 golden 侧空白渲染，topbar/screens golden 断言超时失败。
   - **影响**：本次以 `serve-pixel-parity.mjs` 自托管（本地 UMD + esbuild）闭环修复，但属上一份复盘已记录的同类阻点再次出现。

## 3 根因归类

1. **workspace 失败误判** — 根因：缺少"受 auth gate 保护的业务 route 在未 mock 认证的浏览器测试中渲染 auth-route-gate"这一显式核查项；归类 **spec-plan / README**（frontend-workspace 测试约定）+ 一次性执行判断不足（非纯流程缺陷，但归类核查可固化）。
2. **删除型计数 gate 分散** — 根因：模块删除缺少统一的"删除影响面 checklist"（计数 gate 清单 + drop-migration net-state lint 约定）；归类 **skill / README**（change-intake/implement 的删除型变更预读 + migrations/lint 文档）。
3. **codegen-events-check 语义** — 根因：gate 语义（工作树 vs HEAD）未在 development.md/Makefile 注释显式说明删除场景下的预期；归类 **README**（docs/development.md §codegen 注解）。
4. **golden CDN 依赖** — 根因：`ui-design/` 真理源以 CDN script 加载，pixel-parity 缺省离线自托管；归类 **README / skill**（ui-design 本地验证 README + parity harness 约定），与既往复盘建议一致，本次已落地基础设施修复。

## 4 对流程资产的改进建议

1. **删除型变更影响面 checklist**（落点：`change-intake` / `implement` skill 或 `AGENTS.md` §2.1.2 deep reconcile；优先级 **high**）：模块/契约删除时强制核查的统一清单——OpenAPI inventory/fixtures 计数、generated drift、seed-migration & prompt-lint 的 drop-migration net-state、privacy matrix table 计数、eval ≥N 基线、registry 字典计数、ai-profile-coverage、events schema 文件数、跨层零残留 `rg`、scenario INDEX 行。避免逐个 gate 试错。
2. **受保护业务 route 的浏览器测试约定**（落点：`frontend/tests/pixel-parity/README` 或 frontend-workspace spec 测试段；优先级 **high**）：明确"auth-gated route 的 pixel-parity helper 必须先 mock `/me` authenticated，否则渲染 auth-route-gate"，并把同类过时测试一次性核查纳入 frontend-shell auth gate 引入时的回归清单。
3. **codegen-events-check 语义注解**（落点：`docs/development.md` §codegen 或 Makefile target 注释；优先级 **medium**）：说明该 gate 是工作树 vs HEAD 的 `git diff --exit-code`，删除/契约变更提交前"失败"属预期，提交后即清。
4. **ui-design golden 离线自托管固化**（落点：`docs/ui-design/` 本地验证 README + `ui-design/README`；优先级 **medium**）：把本次 `serve-pixel-parity.mjs` 的本地 UMD + esbuild 自托管模式记为 pixel-parity 默认离线运行方式，避免每次复盘重复发现 CDN 阻点。

## 5 建议优先级与后续动作

### 下一轮最值得实施（按依赖顺序）

剩余三个设计决策围绕**简历资产模型**形成依赖链，D-20 为 foundational，必须最先实施以避免 D-14/D-15 返工：

1. **D-20 简历扁平化（foundational，最大跨层）** — owner：`frontend-resume-workshop` + `backend-resume`。`resumeVersionId` → `resumeId` 全层契约重命名（OpenAPI/backend/migration/frontend/practice/debrief）、删除版本树/主版本/岗位定制/轻量问答、改写建议仅"采纳"、采纳后预览选覆盖或另存。先做此项让后续 resume 绑定/派生口径统一。
2. **D-14 parse 单次确认漏斗** — owner：`frontend-home-job-picks-and-parse/001`（spec §10 已写，plan 待重开 Phase 7）。**无需后端契约变更**（createPracticePlan/startPracticeSession/updateTargetJob 已就位，"立即面试"复用 report D-19 同款 requestAuth→workspace auto-start relay）：轮次点选（`currentRoundIdx`）、绑定简历 pill + ResumePickerModal、底部「立即面试」/「仅保存规划」双按钮、删除旧单一「确认」。resume 真实 picker 列表与 D-20 flat 模型纠缠，建议 D-20 后实施。
3. **D-15 debrief 选一带二** — owner：`frontend-debrief`。**需先做 backend 契约新增**：`PracticeSession.resumeId`（backend-practice）+ `TargetJob.defaultResumeId`/`latestMockSessionId`（backend-targetjob），当前 generated types 均无；自动带入（选 session → JD+简历，选 JD → 默认简历+最近面试）依赖这些元数据，且与 D-20 resume 模型纠缠。建议 D-20 之后、契约 addendum 先行。

### 可延后

- 把本复盘 §4 的 4 项流程改进固化到对应 skill/README/development.md（high 两项建议在启动 D-20 前先落地"删除影响面 checklist"与"受保护 route 测试约定"，避免 D-20 大规模契约重命名时重复踩计数 gate 与过时测试坑）。
