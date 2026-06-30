# Frontend Parse Resume Binding 交付复盘报告

> **日期**: 2026-06-30
> **审查人**: Codex

## 1 复盘范围与成功证据

本次复盘范围是 `frontend-home-job-picks-and-parse/001-home-jd-import-and-parse` 的 Parse resume binding 修复：首页 JD 解析完成后，用户必须显式选择已有 ready 简历或进入简历创建，才能保存规划或启动模拟面试；系统不得默认选中最近简历，成功 handoff 不得再携带 `resume-unbound`。

已通过的成功证据：

- Red gate：新增 `ParseResumeBinding.test.tsx` 后，旧实现失败于 `listResumes` 未调用。
- Red follow-up：用户指出默认选中最新 ready 简历仍不满足强制选择后，扩展 `ParseResumeBinding.test.tsx`，旧实现失败于默认展示首份简历且 Save/Start 已启用。
- `CI=true COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend test src/app/screens/parse/ParseResumeBinding.test.tsx`：4 tests PASS。
- `CI=true COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend test src/app/screens/parse`：6 files / 32 tests PASS。
- `CI=true COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend build`：PASS（仅保留既有 Vite chunk size warning）。
- `CI=true COREPACK_ENABLE_DOWNLOAD_PROMPT=0 corepack pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/parse.spec.ts --grep "save plan navigates|start interview hands off"`：desktop/mobile 4 tests PASS，并输出 Save plan 与 Start interview 的 P0.016 markers。
- `test/scenarios/e2e/p0-016-parse-confirm-to-workspace/scripts/setup.sh` → `trigger.sh` → `verify.sh` → `cleanup.sh`：PASS。
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`：PASS。
- Bug 记录：[BUG-0130](../bugs/BUG-0130.md)。

## 2 会话中的主要阻点/痛点

### 2.1 历史 P0.016 成功口径把缺简历状态当成目标状态

- **证据**：旧 P0.016 围绕 Confirm → Workspace handoff 收口，历史 browser gate 曾断言 `workspace-missing-resume`，这与当前“启动前必须绑定简历”的产品语义冲突。
- **影响**：旧 gate 即使通过，也会继续允许 JD-only planning 作为成功路径。

### 2.2 owner context 需要从现象反查后才能稳定定位

- **证据**：问题表面是“模拟面试无法绑定简历”，但实际 owner 是首页 JD parse launch；需要同时读取 product/UI truth source、`frontend-home-job-picks-and-parse` spec/plan、ParseScreen 和 Workspace start-practice bridge 才能确认修复边界。
- **影响**：如果直接在 Workspace 或 Practice 后置补简历，会保留一次用户可见的二次确认，也无法阻止缺材料规划已经生成。

### 2.3 Playwright 初次执行受 stale dist 影响

- **证据**：直接运行 focused Playwright 时曾命中旧构建产物；重新执行 `pnpm --filter @easyinterview/frontend build` 后，同一 focused Playwright gate 通过。
- **影响**：浏览器 gate 必须依赖当前构建产物，scenario trigger 中需要保留 build 前置步骤。

### 2.4 “绑定简历”容易被误实现成隐式默认

- **证据**：Phase 7 首版把 ready 列表排序后默认选中最新简历，用户指出这仍然绕过了显式选择。
- **影响**：用户可能没有意识到哪份简历被用于面试规划，仍会形成用户材料确认不足的问题。

## 3 根因归类

- **spec-plan**：owner plan 历史 Phase 对 Confirm handoff 语义已经落后于当前流程，需要新增 Phase 7 并把 ready resume binding 写成成功条件。
- **spec-plan**：Phase 7 还必须把“显式选择”写成成功条件，不能只写“绑定 ready resume”，否则实现会倾向于自动选择最近项。
- **scenario scripts / README**：P0.016 wrapper 需要从“允许缺简历 workspace 状态”改为“拒绝 `resume-unbound` / `workspace-missing-resume` 成功 marker”。
- **no repo change needed**：stale dist 是执行顺序问题，本次已在 scenario trigger 保留 build 步骤，不需要额外治理文档修改。

## 4 对流程资产的改进建议

- 对任何会触发 AI 规划或练习会话的入口，owner plan 应明确列出必需上下文字段，并为缺字段状态写 negative gate。
  - **落点**：spec-plan
  - **优先级**：high

- 对候选人材料这类用户必须确认的输入，gate 应明确禁止自动默认值，并断言选择前主动作 disabled。
  - **落点**：spec-plan / frontend tests
  - **优先级**：high

- BDD wrapper 的 verify 脚本应同时检查正向成功 marker 和负向旧 marker，避免旧成功语义长期留在场景名称或日志里。
  - **落点**：scenario scripts / README
  - **优先级**：high

- 后续如继续调整 Parse launch 或 Workspace autoStart，建议把 Save plan 与 Start interview 的 browser fixture helper 抽出来，减少重复 mock route 成本。
  - **落点**：frontend test helper
  - **优先级**：medium

## 5 建议优先级与后续动作

最高优先级是对 Workspace autoStartPractice 消费侧做一次小范围 L2 review：确认它稳定消费 Parse 传入的真实 `resumeId`，并且不会在其他入口重新生成 `resume-unbound` 练习上下文。

第二优先级是审计同一核心链路的其他 AI 规划入口，查找是否仍存在“先生成规划、后补用户材料”的路径。

可以延后处理的是 parse Playwright mock helper 抽象；当前重复仍可控，等下一次新增 parse browser gate 时再抽取更稳妥。
