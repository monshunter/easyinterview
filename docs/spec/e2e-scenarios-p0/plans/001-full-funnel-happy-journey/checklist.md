# 001 Full Funnel Happy Journey Checklist

> **版本**: 1.4
> **状态**: active
> **更新日期**: 2026-05-24

**关联计划**: [plan](./plan.md)

## Phase 0: 真后端环境与前置依赖验证

- [x] 0.1 确认 `make dev-up` postgres 可达 + `make migrate-up` 至最新；记录 `DATABASE_URL` 约定（验证来源：`make dev-doctor` / `make migrate-status` 输出）
  <!-- verified: 2026-05-24 command="make dev-doctor && DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' make migrate-up && DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' make migrate-status" evidence="dev-stack postgres/redis/minio OK; migrate status version=10 dirty=false" -->
- [x] 0.2 确认 `config.LoadCanonical(AppEnv:"test")` 加载成功且漏斗 AI 步骤（`resume.parse.default` / `target.import.default` / practice / `report.generate.default`）所需 profile/registry 在未配置 `AI_PROVIDER_*` 时可解析；实际 journey AI 由 scenario harness 注入确定性 stub / fixture client（验证来源：`cd backend && go test -v ./cmd/api -run 'TestE2EP0ConfigPreflight' -count=1`）
  <!-- verified: 2026-05-24 command="cd backend && go test -v ./cmd/api -run 'TestE2EP0ConfigPreflight' -count=1 && cd backend && go test -list 'TestE2EP0ConfigPreflight|TestE2EP0098' ./cmd/api" evidence="preflight executes 6 active chat profiles and now fails config.LoadCanonical errors instead of inheriting live-scenario skip behavior; TestE2EP0098 not implemented yet" -->
- [x] 0.3 按 §3.1 operation matrix grep 确认 9 行 operation matrix 真实挂载或具备显式备选状态（8 个主链必经 operation + `getJob` 备选轮询 / handler gate），非 mock-only（验证来源：`grep -rn "<operationId>" backend/internal/api/generated/` 命中 + handler 路径存在 / matrix 状态复核）
  <!-- verified: 2026-05-24 command="cd backend && go test -v ./cmd/api -run 'TestE2EP0(OperationMatrix|ConfigPreflight)' -count=1" evidence="TestE2EP0OperationMatrixPreflight asserts 9 matrix rows against generated AllRoutes, fixture files, cmd/api route wiring, and concrete handler declarations" -->
- [x] 0.4 设计 journey 前置 seed：通过 `registerResume` + `resume_parse` stub 产出 ready resume asset，不直接插入 ready 行；定义 cleanup 边界（验证来源：harness helper 设计评审）
  <!-- verified: 2026-05-24 command="cd backend && DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test -v ./cmd/api -run 'TestE2EP0FullFunnelReadyResumeSeedUsesRegisterResumeAndRunner' -count=1" evidence="TestE2EP0FullFunnelReadyResumeSeedUsesRegisterResumeAndRunner calls registerResume through the HTTP handler, processes resume_parse through the SQL runner with deterministic test AI, verifies ready asset/job/outbox, then deletes user/session/idempotency/resume/job/outbox seed data" -->

## Phase 1: API-level full-funnel journey（E2E.P0.098）

- [ ] 1.1 编写 `backend/cmd/api/full_funnel_journey_scenario_test.go` harness（httptest + DATABASE_URL + LoadCanonical + scenario stub/fixture AI + 真实 stack，postgres 不可达 `t.Skip`）（验证来源：`TestE2EP0098` 初始 Red 可运行）
- [ ] 1.2 import → poll `getTargetJob` ready，断言 `target_import` 经真实 runner 完成、解析结果落库（验证来源：`TestE2EP0098` import 段断言）
- [ ] 1.3 `createPracticePlan(baseline)` → planId，断言 plan 落库并绑定 targetJob/resume（验证来源：`TestE2EP0098` plan 段断言）
- [ ] 1.4 `startPracticeSession` + `appendSessionEvent` 事件循环，断言 session/events 落库、outbox 仅一次（验证来源：`TestE2EP0098` session 段断言）
- [ ] 1.5 `completePracticeSession` → poll `getFeedbackReport` ready，必要时用 `getJob` 作为 job 状态备选轮询 / handler gate，断言 `report_generate` 经真实 runner 完成、nextActions 含 next_round（验证来源：`TestE2EP0098` report 段断言）
- [ ] 1.6 `createPracticePlan(next_round, sourceReportId)` → 派生 planId，断言关联 source report 且不同于首个 plan（验证来源：`TestE2EP0098` next_round 段断言）
- [ ] 1.7 start/complete/createPlan Idempotency-Key replay 无重复副作用（验证来源：`TestE2EP0098` 幂等段断言）
- [ ] 1.8 隐私红线 + route-aware legacy-negative 断言（验证来源：`TestE2EP0098` 隐私段断言 + `verify.sh` 负向 grep；旧 route 覆盖 welcome/growth/plan/mistakes/drill/followup/experiences/star/onboarding/独立 voice，且不误伤合法 `startPracticeSession` / `createPracticePlan` / `resumeAssetId`）
- [ ] 1.9 `TestE2EP0098` 全程转 Green：`cd backend && go test -v ./cmd/api -run 'TestE2EP0098' -count=1`（验证来源：Go test pass marker）

## Phase 2: Playwright full-stack journey（E2E.P0.099）

- [ ] 2.1 脚本拉起真后端进程（dev-stack postgres + scenario stub/fixture AI）+ 前端 build/preview 通过 `VITE_EI_API_MODE=real` / `VITE_EI_API_BASE_URL=http://127.0.0.1:<backend-port>/api/v1` 指向真后端；seed user + resume asset（验证来源：setup.sh 启动 marker + health probe + frontend real-mode env marker）
- [ ] 2.2 编写 `frontend/tests/e2e/full-funnel-journey.spec.ts` + `frontend/playwright.e2e.config.ts`（`testDir: "./tests/e2e"`；`outputDir` 由 `EI_PLAYWRIGHT_OUTPUT_DIR` 指向 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/playwright`）驱动 UI 走完漏斗（导入→解析→workspace→practice→generating→report→next_round CTA）（验证来源：Playwright spec，初始 Red）
- [ ] 2.3 断言解析 loading 与 report generating 真实轮询 UI 在异步 job 推进下过渡到 ready（验证来源：Playwright 轮询断言）
- [ ] 2.4 断言 next_round CTA 触发 `createPracticePlan(next_round)` + `startPracticeSession` 且 nav query 含派生 planId / fresh sessionId（验证来源：Playwright handoff 断言 + network spy）
- [ ] 2.5 隐私（URL/storage/console）+ legacy 负向断言（验证来源：Playwright 隐私断言 + route-aware `grep` 负向 + frontend scope gate 或等价 scoped grep）
- [ ] 2.6 `full-funnel-journey.spec.ts` 转 Green：`EI_PLAYWRIGHT_OUTPUT_DIR="$REPO_ROOT/.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/playwright" pnpm --filter @easyinterview/frontend exec playwright test --config=playwright.e2e.config.ts tests/e2e/full-funnel-journey.spec.ts`（验证来源：Playwright pass marker，且不会被默认 `tests/pixel-parity` testDir 排除；trace/screenshot/video 等产物不写入 `frontend/.playwright-output` / `frontend/test-results`）

## Phase 3: 场景登记与收口

- [ ] 3.1 创建 `p0-098-*` / `p0-099-*` 场景目录（README + data + 四段脚本）；`verify.sh` 检查 runner 日志真实执行证据并拒绝 no-op，且确认 P0.099 Playwright 产物全部位于 `.test-output/e2e/p0-099-full-funnel-fullstack-ui-journey/`；执行 wrapper cleanup 后保留前置失败退出码；登记 `test/scenarios/e2e/INDEX.md`（验证来源：脚本独立执行 + INDEX 行）
- [ ] 3.2 BDD-Gate: 验证 `E2E.P0.098` setup→trigger→verify→cleanup 通过并记录证据
- [ ] 3.3 BDD-Gate: 验证 `E2E.P0.099` setup→trigger→verify→cleanup 通过并记录证据
- [ ] 3.4 文档一致性：`validate_context.py` / `sync-doc-index --check` / `make docs-check` / `git diff --check` 通过（验证来源：各命令退出码）
- [ ] 3.5 operation matrix 终态与实现一致核对（验证来源：§3.1 matrix 逐行复核）
