# Frontend Shell Language Dropdown 交付复盘报告

> **日期**: 2026-05-08
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付把 `ui-design` 与正式 `frontend` TopBar 语言控件从二选一 toggle 调整为可扩展 icon dropdown，并把正式前端语言选项抽到 `frontend/src/app/i18n/localeCatalog.ts` 的 `SUPPORTED_LOCALES`，后续新增语言只扩 locale 文件、catalog 元数据与测试，不改变 TopBar 控件结构。
- 二次修订后，TopBar 按 globe icon + 当前语言标签展示（如 `中文` / `English`），不再把多个候选语言拼在按钮上；locale 初始化优先级固化为用户显式选择（`localStorage["ei-lang"]`）> 浏览器 locale > English fallback。
- 已同步 `docs/ui-design/`、`frontend/README.md`、`frontend-shell` spec/history、002/003 plan 与 P0.004/P0.006 场景资产，删除“语言 toggle 是最终形态”的旧口径。
- 验证通过：`pnpm --filter @easyinterview/frontend test` 40 files / 240 tests PASS；focused language/TopBar suite 6 files / 33 tests PASS；`pnpm --filter @easyinterview/frontend test:pixel-parity tests/pixel-parity/topbar.spec.ts` 20 PASS；`pnpm --filter @easyinterview/frontend exec vite build` PASS；`make docs-check` zero drift；`git diff --check` clean。早前同轮已完成 `test:pixel-parity` 全量 48 PASS、`E2E.P0.004` 与 `E2E.P0.006` setup→trigger→verify→cleanup PASS；二次修订后用 focused P0.004 Vitest 与 TopBar parity 覆盖变更面。

## 2 会话中的主要阻点/痛点

- 初始实现仍把语言列表局部写在 TopBar 内，虽然控件形态已变成 dropdown，但不满足“以后会有很多种语言”的扩展要求。
  - **证据**：用户明确补充“以后会有很多种语言，所以必须为多语言设计”后，交付才新增 `localeCatalog.ts` 并把 TopBar 改为从 `SUPPORTED_LOCALES` 渲染。
  - **影响**：若只停在二语言局部数组，后续新增第三种语言会再次修改 TopBar 组件和测试，违反多语言元数据集中管理。
- 全量 `pnpm --filter @easyinterview/frontend typecheck` 被既有 jobs 导出漂移阻断。
  - **证据**：命令失败于 `src/lib/jobs/jobs.test.ts` 引用不存在的 `JOB_TYPE_EMBEDDING_UPSERT`。
  - **影响**：本次只能用 focused tests、全量 Vitest、Vite build、Playwright parity 和 scenario gates 证明本改动正确；typecheck 需要由 jobs owner 单独修复。
- P0.006 screenshot baseline 在本机首次运行时不存在，首轮 Playwright 写入 baseline 后失败，第二轮才 PASS。
  - **证据**：首轮 `test:pixel-parity` 报告 desktop/mobile `home-warm-light` snapshot missing；baseline 文件生成后重跑 48 PASS。
  - **影响**：这是 README 已记录的本地 baseline 维护成本，不是本次功能缺陷，但会增加首次验证时间。

## 3 根因归类

- 多语言扩展要求属于 `spec-plan`：原 spec 和 README 强调 locale 文件独立，但没有强制“语言控件选项必须来自 locale catalog，不得在 TopBar 写私有二语言数组”。
- typecheck 阻断属于 `no repo change needed`（本 session 范围内）：失败点在 jobs 旧契约，不在本次 TopBar/i18n 写入范围内。
- screenshot baseline 首跑成本属于 `README` 已覆盖的运行现实：P0.006 README 已说明 baseline 由本地维护并 gitignored，本次无需改流程。

## 4 对流程资产的改进建议

- 在未来 i18n 相关 plan/checklist 中显式加入“locale catalog gate”：新增语言必须更新 locale file、catalog metadata、TopBar dropdown option test、Accept-Language regression。
  - **落点**：spec-plan
  - **优先级**：medium
- 为 jobs 导出漂移单独开 owner 修复，恢复 `pnpm --filter @easyinterview/frontend typecheck` 作为可用全量 gate。
  - **落点**：spec-plan
  - **优先级**：high
- 保持 P0.006 baseline 首跑说明现状即可；不要把 gitignored screenshot baseline 纳入本次代码交付。
  - **落点**：no repo change needed
  - **优先级**：low

## 5 建议优先级与后续动作

- 最高优先级：修复 frontend jobs typecheck 漂移，让后续 UI/i18n 改动能重新使用全量 typecheck 作为硬门禁。
- 中优先级：下一个新增语言 workstream 从 `localeCatalog.ts` 开始，先补 locale metadata 和 locale 文件，再扩测试，不改 TopBar 控件结构。
- 低优先级：P0.006 screenshot baseline 继续保持本地生成 / gitignored 策略，只有当 CI 需要稳定 baseline artifact 时再升级流程。
