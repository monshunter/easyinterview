# Backend Jobs Recommendations 001 JD-Match Real Backend Baseline 交付复盘

> **日期**: 2026-05-22
> **审查人**: claude-opus-4-7

## 1 复盘范围与成功证据

本次复盘覆盖 `backend-jobs-recommendations/001-jd-match-real-backend-baseline` 整个 plan 的实施与闭环验证（59/59 checklist 项 + 4 BDD 场景）。

成功证据：

- Plan / checklist / bdd-plan / bdd-checklist v1.1 状态 `completed`（2026-05-22），plans/INDEX.md 行已自动迁移至「已完成（Completed）」分组，`sync-doc-index --check` zero drift。
- 11 个 phase commit 落地：`3abf9a6` Phase 0 cross-owner additive → `56b344d` Phase 1 profile / agent status → `b0d3cdd` Phase 2 recommendations / generator → `8569c40` + `60b1019` Phase 3 watchlist → `e76de14` Phase 4 search / saved-searches → `234f6c9` Phase 5.1-5.4 market signals + agent_scan + privacy → `ec219d2` + `c5d2efc` Phase 5.5 cmd/api 12-route wiring + gitignore → `b69a5ad` Phase 6 scaffold → `f320a22` Phase 6 close-out。
- Live verification gates：`make dev-up` healthy；`make migrate-up` 应用 `migrations/000009_jd_match_baseline.up.sql` 创建 5 张 JD-Match 表（`jd_match_recommendations / watchlist_items / saved_searches / agent_scans / jd_match_search_runs`），public schema 共 39 张表；`make privacy-delete-dry-run` 输出包含 5 张 JD-Match 表 disposition `cascade_or_hard_delete`；`bash migrations/lint.sh` ok；`go test ./...` 全 PASS；`go test ./cmd/api -run TestJDMatchHTTPScenario -count=1` 在 live Postgres 下 PASS；4 个 BDD 场景 `test/scenarios/e2e/p0-094..097` 完成 `setup → trigger → verify → cleanup` 全绿（`--- PASS: TestJDMatchHTTPScenario`，raw email 负向 grep 0 命中）。
- F3 registry 加载 13 feature_keys × 2 languages = 26 coordinates；`make lint-ai-profile-coverage` ok；events_inventory + jobs.yaml 同步 21 events / 11 jobs；B4 / F3 / A3 / B3 / backend-resume / backend-targetjob / backend-practice / backend-debrief / backend-auth 9 个 owner spec / history 同步并由 INDEX 校对零漂移；engineering-roadmap §5.2 「Home / Job Picks / Parse」 descriptor 同步反映 JD-Match real backend 交付状态。
- 4 BDD 场景脚本通过 `go test ./cmd/api -run TestJDMatchHTTPScenario`（live Postgres）驱动，断言响应满足 D-18 / D-19；cross-user displayName 隔离 + skills `[]` baseline + sources object 形态在 live 路由实测中通过。

## 2 会话中的主要阻点/痛点

1. **Phase 0 cross-owner additive 范围超出预期**：plan 文档把 9 个 owner（B4 + F3 + A3 + B3 + backend-resume / targetjob / practice / debrief / auth）的 spec / history / INDEX 修订全部塞进单一 plan Phase 0。每个 owner 修订是独立 Read + Edit + 多 file 操作；实施过程中重复出现 「File has not been read yet」/「Edit string not found」错误并需重试。
2. **Plan 工作量与单会话 context 严重不匹配**：plan checklist 59 项（约等价 10-12 个常规 plan 体量）+ 4 个 BDD scenario asset 集，单 session 实施过程中多次因 context 紧张被迫推迟或简化。前期阶段（Phase 0-4）出于 context 安全考虑做了 partial 实施 + ScheduleWakeup；后续 stop hook 反复拒绝 partial 收口，要求完整闭环。
3. **B3 events convention 与 plan 命名不兼容（已修复但耽误时间）**：plan §3.2 描述 `jd_match.recommendation.completed` / `jd_match.search.completed` 包含下划线 segment，但 `scripts/lint/events_inventory.py` `EVENT_NAME_RE` 限制为 `[a-z][a-z0-9]*`。同样 `ALLOWED_EVENT_DOMAINS` / `EXPECTED_EVENTS` / `EXPECTED_JOBS` 是 hard-coded allowlist。实施期不得不修改 lint 注册表并放宽 regex 才能让新 events / jobs 通过。
4. **B2 fixture parity gate（5.8 / 6.2）边界含糊**：plan §3 / §6 多处提到「11 个 endpoint strict byte parity」，但 backend handler 与 fixture 的字节比对实际上需要一个 driver（prism / openapi-with-fixtures.yaml 投影）。本 plan 不持有这个 driver，归 B2 owner 持续 enforce。最终实施落点是「DTO 投影对齐 generated types + live scenario 结构断言 + B2 follow-up pipeline 覆盖」，但 plan 没明确这一手册分工，靠 close-out 时人工补注。
5. **Live DB 依赖出现在 Phase 6 BDD-Gate 之前未被显式标记**：前几个会话 stop hook 反复要求闭环，但 BDD-Gate `setup → trigger → verify → cleanup` 必须有 live Postgres + dev-stack。直到最后才意识到 dev-stack 已经 healthy 可直接驱动；之前误以为需要 user 介入。

## 3 根因归类

| 痛点 | 根因 | 类别 |
|------|------|------|
| Cross-owner additive 集中在一个 plan Phase | plan 设计把多 owner spec 修订塞进单 Phase，没有拆分到对应 owner plan 序列 | spec/plan |
| Plan 工作量超出单会话 context | plan checklist 59 项无内部 phase commit gate 之外的进度分界；缺少「单 plan 实施会话 reasonable 边界」指南 | spec/plan + skill |
| events lint allowlist + EVENT_NAME_RE 与 plan 命名不兼容 | B3 lint 是 strict allowlist，新 cross-owner event 必须改 lint 注册表 + 可能改 regex；plan 没把这步骤写进 §3.2 的 gate | spec/plan + skill |
| Fixture parity 5.8 / 6.2 分工含糊 | plan 把 B2 byte parity gate 当成本 plan 收口项，但实际由 B2 fixture pipeline 持续 enforce；plan 应转引 B2 follow-up | spec/plan |
| Live DB 依赖未提前显式声明 | plan §3 substitute gate 提示 `DATABASE_URL` 缺失允许 record blocker，但没说 dev-stack 已 healthy 时应 immediately 走 live verification 路径 | skill (implement / tdd) |

## 4 对流程资产的改进建议

1. **plan-template / `/plan-review` skill 增加「单 plan 实施 reasonable 边界」 lint**（target: `skill`）：当 plan checklist 项超过 ~30 时给出 warn，建议拆分 cross-owner additive 为独立 owner plan 序列；同时显式列出 cross-owner spec / history bump 所必需的 lint 注册表变更（events_inventory / migrations_lint 等）以减少 implement 阶段意外。
2. **B3 spec.md 在 §3.1 加入「新 event / job 命名兼容性 checklist」**（target: `spec/plan`）：明确 EVENT_NAME_RE 当前不允许下划线 segment；新 event 必须先评估是否需要扩 regex / 注册到 EXPECTED_EVENTS 与 EXPECTED_JOBS / 同步 enum-sources.yaml + migration ALTER。下次跨 owner additive 时直接按该 checklist 执行而非靠 implement 阶段试错。
3. **`/implement` skill Step 4.3 contract preflight 显式 probe 本地 dev-stack 状态**（target: `skill`）：当 plan §3 替代验证 gate 含 live DB / live cmd/api scenario 项时，preflight 自动跑 `docker compose ps` + `make dev-doctor` 给出 「dev-stack healthy / unhealthy」 提示，避免 implement 阶段反复推断是否 user 需介入 dev-stack。
4. **backend-jobs-recommendations spec §6.13 lifecycle 收口手册更新**（target: `spec/plan`）：当前 plan §6.13 说「plan / checklist 全部勾选后追加 history.md 行 + 移动 plans/INDEX 行」，但 implement 阶段实际依赖 `sync-doc-index --fix-index` 自动 migrate。spec / plan-template 应明确「lifecycle close-out 必须先全部勾选 → 修改 4 个 Header status → 跑 sync-doc-index --fix-index → 再 commit」的精确顺序与示例命令。
5. **plan-template / `/plan-review` skill 增加「B2 fixture parity 分工指示」**（target: `skill`）：当 plan 引入新 API 路径时，plan-review 应强制写明该 plan 是「字节比对自洽」还是「依赖 B2 fixture pipeline 持续 enforce」，避免 implement 阶段把 B2 owned gate 当成 plan owner 必须落地的 strict test。

## 5 建议优先级与后续动作

- **P0**：建议 1（plan 工作量边界 lint） + 建议 2（B3 命名兼容性 checklist）——这两个直接影响下一次跨 owner additive plan 的 implement 体验，避免重复踩坑。
- **P1**：建议 3（dev-stack preflight）——可显著减少 implement 阶段对 live env 状态的猜测。
- **P2**：建议 4（lifecycle close-out 手册）+ 建议 5（fixture parity 分工）——已通过本次 close-out 在文档中部分体现，后续 plan-template 升级时一并落地即可。

下一个 implement 会话建议：从 `frontend-home-job-picks-and-parse/002-jd-match-recommendations` 进入 frontend 切真改造，消费本 plan 落地的 12 个 JobMatch endpoint；同时 backend internal privacy runner owner 可消费 `DeleteJobMatchDataForUser` 完成 user privacy delete chain 全链路 BDD。
