# Resume Phase 8 Flat Gate Drift 交付复盘报告

> **日期**: 2026-06-14
> **审查人**: Codex (GPT-5)

## 1 复盘范围与成功证据

本次交付范围是对 `frontend-resume-workshop/003-branch-rewrites-and-edit` 做 D-20 Phase 8 L2 收口：把 BUG-0123 暴露的 flat profile / route context 回归固化为剩余验收 gate，修正 P0.084 / P0.086 / P0.087 wrapper 对当前 `structuredProfile.experiences[]` 字段的误杀，并将 owner plan / BDD docs 从 historical branch/version 语义更新为当前 flat detail / Rewrites / Edit 语义。

已完成的交付项：

- `context.yaml` 改为 current flat discovery，移除旧 branch/version operation matrix，补齐 `updateResume`、`duplicateResume`、`requestResumeTailor`、`targetJobId`、flat profile save 等关键词。
- `plan.md` / `checklist.md` / `bdd-plan.md` / `bdd-checklist.md` 增加 Phase 8 current gate，覆盖 flat route、accept-only save modal、flat profile merge、route `targetJobId` rerun、Edit/export/copy non-regression、P0.084-P0.087 wrapper PASS。
- P0.084 / P0.086 / P0.087 verify retired grep 从 generic `experiences` 收紧为 `ExperiencesScreen|experiences-route`，继续拒绝旧独立模块但允许当前 flat profile field。
- `test/scenarios/e2e/INDEX.md` 和 P0.087 README 更新为 D-20 flat scenario 语义。
- 新增 [BUG-0124](../bugs/BUG-0124.md)，记录 scenario wrapper false-positive 与 Phase 8 gate drift。

通过证据：

- P0.084 `setup -> trigger -> verify -> cleanup` PASS；trigger log 显示 real-backend gate + 5 files / 46 tests passed。
- P0.085 `setup -> trigger -> verify -> cleanup` PASS；trigger log 显示 real-backend gate + 3 files / 30 tests passed。
- P0.086 `setup -> trigger -> verify -> cleanup` PASS；trigger log 显示 real-backend gate + 4 files / 41 tests passed。
- P0.087 `setup -> trigger -> verify -> cleanup` PASS；trigger log 显示 focused Vitest 5 files / 39 tests passed、frontend build PASS、Playwright 4 passed。
- `python3 .agent-skills/implement/shared/scripts/validate_context.py --context docs/spec/frontend-resume-workshop/plans/003-branch-rewrites-and-edit/context.yaml --docs-root docs --target frontend`
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
- `make docs-check`
- `pnpm --filter @easyinterview/frontend typecheck`
- `pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop` 通过，27 files / 161 tests PASS；保留既有 React `act(...)` warning。
- `make codegen-check`
- `git diff --check`

## 2 会话中的主要阻点/痛点

- **Retired grep 使用普通字段名导致 false-positive**
  - **证据**：P0.084 首次 verify 失败命中 `ResumeDetailView.tsx` 中合法的 `["sections", "experience", "experiences", "projects"]` allowlist。
  - **影响**：当前 D-20 flat profile field 被误判成旧 Experiences module，使 wrapper 不能同时证明“旧模块删除”和“当前 flat profile merge”。

- **Phase 8 current gate 没有承接 BUG-0123 的三类回归**
  - **证据**：原 Phase 8 checklist 只有 coarse-grained collapse / suggestion update version 两项，未明确 `experience` / `experiences` / `projects` merge、omitted `structuredProfile` fallback、route `targetJobId` rerun body。
  - **影响**：即使 BUG-0123 runtime 修复已经存在，owner plan 的剩余验收仍可能漏掉同类回归。

- **Context discovery 仍偏向 historical branch/version API**
  - **证据**：`context.yaml` 里保留 `branchResumeVersion`、accept/reject suggestion、`updateResumeVersion`、branch package 等 discovery anchors。
  - **影响**：后续 `/plan-code-review` 容易从旧 owner 语义开始审查，错过当前 D-20 flat route/detail/Rewrites/Edit 不变量。

## 3 根因归类

- **负向搜索没有按 current truth source 分类**
  - **类别**：spec-plan / test
  - `experiences` 在旧模块语境中是 shorthand，但在当前 D-20 profile 中是合法字段；wrapper 未将 retired route/component marker 与 current data field 分开。

- **Historical phase 与 current phase 没有在 owner plan 中硬分界**
  - **类别**：spec-plan
  - Phase 1-7 作为历史记录保留，但 Phase 8 没有足够明确地声明当前 owner gate，导致旧 PASS 和旧 BDD 文案继续影响收口判断。

## 4 对流程资产的改进建议

- **保留精确 retired identifier gate，避免普通 schema 字段进入 zero-reference 词表**
  - **落点**：对应 scenario `verify.sh` 与 owner checklist
  - **优先级**：high

- **类似 D-20 的 schema flatten 收口必须把 current-vs-historical 边界写进 plan 第一屏和 checklist phase 名称**
  - **落点**：spec-plan
  - **优先级**：high

- **后续 L2 review context 应优先列 current operationId / component / route context，再把 retired operation 放入 negative gate，而不是 discovery anchor**
  - **落点**：`context.yaml` / plan review checklist
  - **优先级**：medium

## 5 建议优先级与后续动作

下一步优先继续 `/plan-code-review frontend-resume-workshop/003-branch-rewrites-and-edit frontend --fix` 的提交收口：把本轮 BUG-0124、Phase 8 docs、scenario wrapper 和 reports/work-journal 一起提交，避免验证证据和 owner docs 再次分离。

备选路径是随后执行 `/plan-code-review frontend-resume-workshop/003-branch-rewrites-and-edit frontend` 的只读复核，专门检查 D-20 Phase 8 是否仍存在旧 branch/version literal 或 flat profile/context 漏项。
