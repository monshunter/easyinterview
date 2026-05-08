# 001-home-jd-import-and-parse 交付复盘报告

> **日期**: 2026-05-08
> **审查人**: OpenCode Agent

## 1 复盘范围与成功证据

**交付范围**: `frontend-home-job-picks-and-parse/001-home-jd-import-and-parse`（Home + Parse + jd_match 三屏端到端）

**通过证据**:

| Gate | 结果 |
|------|------|
| Vitest | 51/52 files PASS (319/320 tests)，1 pre-existing conventions-parity |
| Playwright pixel parity | 68/68 PASS（34 specs × 2 viewports） |
| Typecheck | 0 errors |
| Build | PASS |
| Scenario E2E (P0.014-017) | 全 PASS（setup→trigger→verify→cleanup） |
| Negative grep (7 patterns) | 0 命中 |
| Docs-check | zero drift |

## 2 会话中的主要阻点/痛点

### 2.1 authContractGate 阻塞 ParseScreen 客户端调用

- **证据**: `authContractGate.test.ts` 只允许 `startAuthEmailChallenge/verifyAuthEmailChallenge/getMe/logout/getRuntimeConfig` 5 个 operationId，ParseScreen 调用 `getTargetJob`/`updateTargetJob` 被拦截
- **影响**: 测试套件在 ParseFlow 文件进入后立即失败，需手动扩展允许集合才能继续

### 2.2 TypeScript boolean 作为 Record key 的索引类型错误

- **证据**: `HIT_CYCLE = { true: "partial", partial: false, false: true }` 的 Record 类型声明失败，`true`/`false` 不能作为 TS Record 的 key
- **影响**: 尝试 `Record<HitState, HitState>` 失败后，改用 `nextHit()` 函数式替代，增加一次 typecheck 循环

### 2.3 Vitest fake timer 下 mockImplementation 递归调用

- **证据**: `spy.mockImplementation` 内调用 `client.getTargetJob(_id)` 导致 spy 自递归，`advanceTimersByTimeAsync(0)` 后立即触发 2961 次调用
- **影响**: Phase 4 延迟约 10 分钟调试 fake timer + mock 交互问题

### 2.4 updateTargetJob 签名参数位置误判

- **证据**: `client.updateTargetJob(targetJobId, body, opts)` 的 spy.mock.calls[0] 结构为 `[id, body, opts]`，测试误取 `calls[0][0]` 作为 body
- **影响**: 一次测试失败后逐层排查 findTestid → fireEvent → spy assertions 链条

### 2.5 Scenario verify.sh 断言不可观测字符串

- **证据**: P0.016 verify.sh 使用 `grep -Fq "Idempotency-Key"` 检查测试日志，但 PASS 日志不输出字面断言字符串
- **影响**: 第四次 scenario 验证失败，需改写 verify.sh 为 `grep -q 'Tests.*passed'`

### 2.6 预存 type error 阻塞 build

- **证据**: HomeScreen.tsx `InterviewContext as Record<string,string>` 非法转换、MockInterviewCard.tsx `colors` possibly undefined
- **影响**: 虽非本次修改引入，但 `pnpm build` 链（tsc --noEmit && vite build）要求 typecheck clean，需在 Phase 6 顺带修复

## 3 根因归类

| 痛点 | 根因 | 类别 |
|------|------|------|
| 2.1 authContractGate 阻塞 | gate 的 ALLOWED_NON_AUTH_OPERATIONS 未随业务 plan 扩展，且 AGENTS.md 未明确要求 plan 预检 gate 文件 | AGENTS.md / spec-plan |
| 2.2 boolean Record key | TypeScript 类型系统限制，属常见编码摩擦 | 无需仓库改动 |
| 2.3 mock recursive loop | Vitest spy 与 mockTransport 两层 mock 叠加时的常见陷阱 | 无需仓库改动 |
| 2.4 参数位置误判 | generated client 方法签名未直观体现在测试中 | 无需仓库改动 |
| 2.5 verify.sh 断言不可观测 | scenario 模板未约定 verify.sh 最小断言模式 | skill / README |
| 2.6 预存 type error | Phase 2-3 交付时未强制 typecheck clean gate | AGENTS.md |

## 4 对流程资产的改进建议

### 4.1 authContractGate 自动发现

- **建议**: `/implement` Step 4.3 (Contract Preflight) 或 `/tdd` Step 1.5 应在发现 plan 涉及 `frontend/src/api/generated/client` 新 operationId 时，自动检查 `authContractGate.test.ts` 的 ALLOWED 集合是否覆盖
- **落点**: AGENTS.md §4.3（新增 sub-rule）或 /implement skill
- **优先级**: high — 每次新 plan 引入 client 调用都会触发

### 4.2 typecheck gate 强制前置

- **建议**: AGENTS.md §2.1 或 `frontend/README.md` 明确：任何 checklist 项完成前，`pnpm typecheck` 必须 0 错误（不仅是当前文件，而是全量）。Phase commit 前如存在预存 type error 应作为 blocking gate 先修复再继续
- **落点**: AGENTS.md
- **优先级**: medium — 避免预存错误阻塞后续 build

### 4.3 scenario verify.sh 最小断言模式

- **建议**: `test/scenarios/README.md` 或 scenario template 中明确：verify.sh 不应依赖测试日志中的字面断言字符串；应使用 PASS/FAIL marker（如 `grep -q 'Tests.*passed'`）或场景自有的 marker 文件
- **落点**: test/scenarios/README.md
- **优先级**: low — 仅在新增 scenario 时参考

## 5 建议优先级与后续动作

| 优先级 | 改进项 | 下一轮执行时机 |
|--------|--------|---------------|
| high | authContractGate 自动发现 | 下一个引入新 client operation 的 frontend plan 前实施 |
| medium | typecheck gate 强制前置 | 下一次 AGENTS.md 修订或 plan-review 时机合并 |
| low | scenario verify 断言模式 | 下一个新增 scenario 时自然采用 |

**无需立刻仓库变更的项**: boolean Record key、mock 递归、参数位置误判——均为常见 TypeScript/Vitest 编码摩擦，一次性的调试成本，不值得为此改动流程资产。
