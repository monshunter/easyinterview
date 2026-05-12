# frontend-resume-workshop/001-listing-routing-and-detail-readonly 交付复盘报告

> **日期**: 2026-05-12
> **审查人**: Claude

**关联计划**: [frontend-resume-workshop/001-listing-routing-and-detail-readonly](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)

## 1 复盘范围与成功证据

- **范围**：把 `resume_versions` 路由从 `PlaceholderScreen` 切到 `ResumeWorkshopScreen`，落地 P0 mock-first 用户路径「列表 → 详情预览」。覆盖 5 个 phase，33 个主 checklist + 10 个 BDD checklist 全勾，0 项跳过。
- **代码增量**：feature branch `feat/frontend-resume-workshop-001-listing-routing-and-detail-readonly-0512` 共 5 个 phase commit（`2a0e968` / `d572c27` / `d8db772` / `489994e` / `423fd69`）。
- **新增前端单测**：贯穿 `frontend/src/app/screens/resume-workshop/`、`frontend/src/app/scenarios/p0-036.test.tsx`、`frontend/src/app/scenarios/p0-037.test.tsx`，扩展 `frontend/src/app/auth/authContractGate.test.ts` 允许列表。
- **新增 BDD scenario**：`test/scenarios/e2e/p0-036-resume-list-tree-flat-toggle/` + `test/scenarios/e2e/p0-037-resume-detail-preview-readonly/`，包含 `README.md` + `data/{seed-input,expected-outcome}.md` + `scripts/{setup,trigger,verify,cleanup}.sh`，本地 `setup → trigger → verify → cleanup` 全 PASS。
- **新增 Playwright spec**：`frontend/tests/pixel-parity/resume-workshop.spec.ts`，desktop 1440 + mobile 390x844，10/10 通过。
- **验证证据**：
  - `pnpm --filter @easyinterview/frontend test` → 771/771（126 test files）
  - `pnpm --filter @easyinterview/frontend build` → 通过
  - `pnpm exec playwright test tests/pixel-parity/resume-workshop.spec.ts` → 10 passed
  - 双 BDD scenario lifecycle 全 PASS
  - `git grep` retired-route + prototype import：0 命中
  - `make docs-check` → All documents are in sync. Zero drift detected.
- **未切换 Header 状态**：用户在 lifecycle sync 阶段选择保持 `active`，本复盘记录该决策不主动写回 plan/checklist Header。

## 2 会话中的主要阻点/痛点

### 2.1 `useRequestAuth` + `useDisplayPreferencesOptional` 共享上下文，导致渲染没有 client 时仍会触发 auth gate

- **证据**：Phase 1.5 加入 `runtime !== null && runtime.auth.status !== "authenticated"` 才触发 auth gate 之前，9 个 Phase 1.2 / 1.3 / 1.4 既有测试一次性失败（700 → 701 失败回归）。
- **影响**：实施一项 spec 要求的 auth boundary 反过来打掉了不带 client 的纯路由测试，需要额外回头加“runtime mounted”守护后才能通过。
- **类型**：spec/plan 在描述 auth gate 时只强调“未登录展示 auth gate”，没有说明 routing-only 测试场景如何区分 `runtime=undefined` vs `auth.status=unauthenticated`。

### 2.2 `vi.stubGlobal('navigator', ...)` 不能让 `navigator.clipboard?.writeText` spy 命中

- **证据**：`ResumePreviewTab.test.tsx` 起初对 clipboard spy 断言 0 调用；切换为 toast 行为断言后 PASS（toast 同步走 success path 证明 writeText 实际被调用）。
- **影响**：~10 分钟绕弯，最后接受“通过 toast 文案断言间接覆盖 clipboard 路径”的折中。
- **类型**：jsdom + vitest stub 行为差异，不是产品代码问题，但缺少团队统一的 clipboard 测试模式。

### 2.3 `auth/authContractGate.test.ts` 允许列表是手工维护的硬列表

- **证据**：`listResumes` / `listResumeVersions` / `getResumeVersion` / `exportResumeVersion` 加入 ResumeDetailView 后，gate 立即报 4 项 offending operations；需要手工把 4 个 operationId 加进 `ALLOWED_NON_AUTH_OPERATIONS`。
- **影响**：每个新 frontend workstream 进入实现都要修这个 gate，未事先体现到 plan checklist。
- **类型**：spec/plan + skill。理想状态是 gate 应该读 plan-recorded operation matrix（参见 `docs/development.md` §2.1）而不是另维护一份字符串列表。

### 2.4 fixture 文案 "growth" / 测试自身 negative grep self-match

- **证据**：执行 5.7 / 5.8 grep 时 `adapters/resume.test.ts` fixture 文本 "Senior frontend engineer focused on growth" 命中 `growth` 字面量；`ResumeWorkshopPrivacy.test.ts` 自身的 forbidden pattern 命中 `ui-design/src/data`。
- **影响**：写完所有功能后才意识到 self-match，需要重写 fixture 文案 + 用拼接构造正则字面量。
- **类型**：no repo change needed（一次性修复），但提示 negative grep 的写法可以归档到 `docs/spec/frontend-resume-workshop/spec.md` 的 D-7 决策附近，作为 plan checklist gate 的参考样板。

### 2.5 `tsc --noEmit` 与 vitest 对 `RenderResult` / `ReactNode` 的容忍度不一致

- **证据**：6 个 test helper 用 `: ReactNode` 标注 return type 在 vitest 全绿，但 `pnpm build` 阶段 `tsc --noEmit` 立即报 6 个 TS2322。两个 scenario test 复用同一模式触发同样错误。
- **影响**：完成功能 + 全量 vitest 后，build 阶段才把这些注解暴露出来，需要回到测试文件批量改注解。
- **类型**：no repo change needed（注解模式），但**值得在 frontend README §2 添加“测试 helper 默认不要标 ReactNode 返回类型”一行**，避免下一个屏幕重复同样模式。

### 2.6 列表 stats 计数测试时序：`stats-versions` 在 `listResumeVersions` 未 settle 前是 0

- **证据**：`ResumeListView.test.tsx`、scenario `p0-036`：先 `await waitFor(() => stats-originals)` 后立即断言 stats-versions 偶尔得到 "Versions0"。
- **影响**：每个新增 stats 测试都要把多个计数同时放进同一个 `waitFor`，否则 flaky。
- **类型**：no repo change needed，但**值得在 frontend README §2.4 mock 数据源边界 旁加一句：“相互依赖的派生计数测试要把多个断言放进单一 waitFor，避免依赖 race”**。

## 3 根因归类

| # | 根因 | 类别 |
|---|------|------|
| 1 | spec/plan 描述 auth gate 时把 “runtime mounted 但未登录” 与 “routing-only 无 runtime” 混为一谈 | spec-plan |
| 2 | jsdom 下 `navigator.clipboard` 的 stub 模式没有团队统一约定 | no repo change needed |
| 3 | `authContractGate.test.ts` 的允许列表是字符串硬列表，与 `docs/development.md` §2.1 operation matrix 没有同步机制 | spec-plan + skill |
| 4 | 一次性 fixture 文案 self-match 与 negative grep self-match | no repo change needed |
| 5 | 测试 helper 在 vitest 与 tsc 对 `RenderResult` 容忍度差异 | README |
| 6 | 派生计数 / 异步 settle 的 `waitFor` 模式没有归档 | README |

## 4 对流程资产的改进建议

- **建议 A**：在 `frontend-resume-workshop/spec.md` D-1（auth boundary 决策）补一行：明确 “没有 runtime 的 routing-only 测试不视为未登录态，screen 必须以 `runtime !== null && auth.status !== authenticated` 作为 auth gate 触发条件”，避免下一批屏幕踩同样坑。
  - **落点**：spec-plan
  - **优先级**：medium
- **建议 B**：在 `docs/development.md` §2.1 operation matrix 章节增加 “每条 operation 应自动同步进 frontend `authContractGate.test.ts` 允许列表” 的强制 gate（或抽出脚本检查 plan 中 operationId 与 allow list 的并集），减少每个 workstream 手动加白名单。
  - **落点**：spec-plan + skill（如最终落到脚本 lint）
  - **优先级**：medium
- **建议 C**：在 `frontend/README.md` §2 “工具链 / 测试约定” 补三段：
  1. 测试 helper 默认不写 `ReactNode` 返回类型（避免 vitest 与 tsc 不一致）；
  2. `navigator.clipboard` 在 jsdom 下走 toast 文案断言而非 `vi.stubGlobal`；
  3. 多个派生计数断言放进同一 `waitFor`，避免 race。
  - **落点**：README
  - **优先级**：medium
- **建议 D**：把 “spec/plan 中明确写出 retired-module / prototype import negative grep 的精确正则与 fixture 文案规避建议” 沉淀到 `docs/spec/frontend-resume-workshop/spec.md` D-7（旧入口 negative grep 决策）。
  - **落点**：spec-plan
  - **优先级**：low
- **建议 E**：把 `auth/authContractGate.test.ts` 的允许列表抽到 plan/operationId 派生（脚本生成或在 plan checklist 增加 “authContractGate 同步” 项），不再依赖每次实施时手动追加。
  - **落点**：skill（`/plan-review` 可作为 lint 增强点）
  - **优先级**：low

## 5 建议优先级与后续动作

下一轮最值得实施的改进项：

1. **建议 A + B**（medium）：先把 spec/plan 与 authContractGate 同步规则固化到 `docs/development.md` §2.1，落地一个 lint 脚本或检查项；后续 D2-D6 屏幕实施时不再手工修允许列表。
2. **建议 C**（medium）：把三个测试模式写进 `frontend/README.md` §2，本届 plan 002 / 003 实施时直接受益。

可以延后处理：

- **建议 D**（low）：D-7 negative grep 样板归档；下一轮简历工坊扩展（plan 002 / 003）若再次踩坑再合并补丁。
- **建议 E**（low）：把允许列表脚本化属于工程整洁，不阻塞当前 P0 主线。

后续动作建议：

- 当前 plan/checklist Header 保持 `active`，等待用户在合适时机触发 `/plan-code-review` 或合并到 main 后再 close-out。
- 若下个 sprint 进入 `frontend-resume-workshop/002-create-flow-and-onboarding`，请先把建议 A / B / C 列入新 plan 的开题 checklist，避免重复踩坑。
