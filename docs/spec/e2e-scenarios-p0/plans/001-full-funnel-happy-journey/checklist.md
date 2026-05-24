# 001 Full Funnel Happy Journey Checklist

> **版本**: 1.0
> **状态**: active
> **更新日期**: 2026-05-24

**关联计划**: [plan](./plan.md)

## Phase 0: 真后端环境与前置依赖验证

- [ ] 0.1 确认 `make dev-up` postgres 可达 + `make migrate-up` 至最新；记录 `DATABASE_URL` 约定（验证来源：`make dev-doctor` / `make migrate-status` 输出）
- [ ] 0.2 确认 `config.LoadCanonical(AppEnv:"test")` 加载成功且漏斗 AI 步骤落 stub、未读 `AI_PROVIDER_*`（验证来源：`go test ./internal/ai/aiclient -run Config -count=1` + grep secret 边界）
- [ ] 0.3 按 §3.1 operation matrix grep 确认 8 个 operationId 真实挂载、非 mock-only（验证来源：`grep -rn "<operationId>" backend/internal/api/generated/` 命中 + handler 路径存在）
- [ ] 0.4 设计 journey 前置 seed（已认证 user + ready resume asset）与 cleanup 边界（验证来源：harness helper 设计评审）

## Phase 1: API-level full-funnel journey（E2E.P0.098）

- [ ] 1.1 编写 `backend/cmd/api/full_funnel_journey_scenario_test.go` harness（httptest + DATABASE_URL + LoadCanonical stub + 真实 stack，postgres 不可达 `t.Skip`）（验证来源：`TestE2EP0098` 初始 Red 可运行）
- [ ] 1.2 import → poll `getTargetJob` ready，断言 `target_import` 经真实 runner 完成、解析结果落库（验证来源：`TestE2EP0098` import 段断言）
- [ ] 1.3 `createPracticePlan(baseline)` → planId，断言 plan 落库并绑定 targetJob/resume（验证来源：`TestE2EP0098` plan 段断言）
- [ ] 1.4 `startPracticeSession` + `appendSessionEvent` 事件循环，断言 session/events 落库、outbox 仅一次（验证来源：`TestE2EP0098` session 段断言）
- [ ] 1.5 `completePracticeSession` → poll `getFeedbackReport` ready，断言 `report_generate` 经真实 runner 完成、nextActions 含 next_round（验证来源：`TestE2EP0098` report 段断言）
- [ ] 1.6 `createPracticePlan(next_round, sourceReportId)` → 派生 planId，断言关联 source report 且不同于首个 plan（验证来源：`TestE2EP0098` next_round 段断言）
- [ ] 1.7 start/complete/createPlan Idempotency-Key replay 无重复副作用（验证来源：`TestE2EP0098` 幂等段断言）
- [ ] 1.8 隐私红线 + legacy-negative 断言（验证来源：`TestE2EP0098` 隐私段断言 + `grep` 负向）
- [ ] 1.9 `TestE2EP0098` 全程转 Green：`cd backend && go test ./cmd/api -run 'TestE2EP0098' -count=1`（验证来源：Go test pass marker）

## Phase 2: Playwright full-stack journey（E2E.P0.099）

- [ ] 2.1 脚本拉起真后端进程（dev-stack postgres + stub AI）+ 前端 build/preview 指向真后端；seed user + resume asset（验证来源：setup.sh 启动 marker + health probe）
- [ ] 2.2 编写 `frontend/tests/e2e/full-funnel-journey.spec.ts` 驱动 UI 走完漏斗（导入→解析→workspace→practice→generating→report→next_round CTA）（验证来源：Playwright spec，初始 Red）
- [ ] 2.3 断言解析 loading 与 report generating 真实轮询 UI 在异步 job 推进下过渡到 ready（验证来源：Playwright 轮询断言）
- [ ] 2.4 断言 next_round CTA 触发 `createPracticePlan(next_round)` + `startPracticeSession` 且 nav query 含派生 planId / fresh sessionId（验证来源：Playwright handoff 断言 + network spy）
- [ ] 2.5 隐私（URL/storage/console）+ legacy 负向断言（验证来源：Playwright 隐私断言 + `grep` 负向）
- [ ] 2.6 `full-funnel-journey.spec.ts` 转 Green：`pnpm --filter @easyinterview/frontend exec playwright test tests/e2e/full-funnel-journey.spec.ts`（验证来源：Playwright pass marker）

## Phase 3: 场景登记与收口

- [ ] 3.1 创建 `p0-098-*` / `p0-099-*` 场景目录（README + data + 四段脚本）；`verify.sh` 检查 runner 日志真实执行证据并拒绝 no-op；登记 `test/scenarios/e2e/INDEX.md`（验证来源：脚本独立执行 + INDEX 行）
- [ ] 3.2 BDD-Gate: 验证 `E2E.P0.098` setup→trigger→verify→cleanup 通过并记录证据
- [ ] 3.3 BDD-Gate: 验证 `E2E.P0.099` setup→trigger→verify→cleanup 通过并记录证据
- [ ] 3.4 文档一致性：`validate_context.py` / `sync-doc-index --check` / `make docs-check` / `git diff --check` 通过（验证来源：各命令退出码）
- [ ] 3.5 operation matrix 终态与实现一致核对（验证来源：§3.1 matrix 逐行复核）
