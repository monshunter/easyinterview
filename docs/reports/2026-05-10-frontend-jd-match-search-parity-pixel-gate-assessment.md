# Frontend JD Match Search Parity Pixel Gate 交付复盘报告

> **日期**: 2026-05-10
> **审查人**: Codex

## 1 复盘范围与成功证据

- 范围：`frontend-home-job-picks-and-parse/002-jd-match-recommendations` 的 L2 follow-up 修复，覆盖 SearchTab 源级 parity、jd_match pixel parity clean-checkout gate、P0.028 scenario verify 与 owner 文档证据同步。
- 成功证据：focused Red 先失败 3 项（缺 naturalLanguageHeading / 自然语言 label / company source），修复后 `SearchTab.test.tsx + jdMatchLocaleNamespaces.test.ts` 2 files / 23 tests PASS。
- 回归证据：`pnpm --filter @easyinterview/frontend test` 107 files / 674 tests PASS；`typecheck` PASS；`build` PASS；`jd_match.spec.ts` Playwright 20 passed；`E2E.P0.028` setup→trigger→verify→cleanup PASS；`make validate-fixtures` 和 `openapi_inventory.py` PASS。
- 交付记录：新增 [BUG-0038](../bugs/BUG-0038.md)，并将 spec / plan / checklist / bdd-plan / bdd-checklist 原地恢复为 active 后写入本次 remediation gate；验收通过后已切回 completed 并同步 INDEX。

## 2 会话中的主要阻点/痛点

- SearchTab review 最初只抓到实现漂移，随后才发现测试和 BDD verify 也固化了旧 4-source 弱口径。
  - **证据**：原 pixel spec 文案为 “four chip filters”，SearchTab 单测只断言 sources 容器存在。
  - **影响**：需要二次补强 Vitest、pixel spec 和 P0.028 verify，而不是单点修 UI。
- Pixel parity 初次复跑失败原因不是视觉差异，而是被 `.gitignore` 忽略的 snapshot baseline 不存在。
  - **证据**：Playwright 报告 “A snapshot doesn't exist”，且 `frontend/.gitignore` 忽略 `tests/pixel-parity/*-snapshots/`。
  - **影响**：历史 “常规运行 PASS” 依赖本地状态，不能作为 clean checkout 证据。
- Playwright 第二次复跑一度命中 stale 4173 server，页面仍使用旧 dist。
  - **证据**：新 testid 在 built source 中存在，但 Playwright 页面 element not found；清理 4173 后同命令 20/20 PASS。
  - **影响**：验证步骤需要显式处理 `reuseExistingServer` 带来的旧 bundle 风险。

## 3 根因归类

- SearchTab 非 loading 区域缺少可执行 parity anchor。
  - **类别**：spec-plan
- Screenshot baseline 被忽略但 checklist evidence 仍宣称常规 PASS。
  - **类别**：spec-plan / README
- Playwright `reuseExistingServer` 复用旧 dist 的风险未在 jd_match gate evidence 中说明。
  - **类别**：README / no repo change needed（本次通过验证步骤规避）

## 4 对流程资产的改进建议

- 在 plan-code-review 的 UI review checklist 中加入“逐段审 ui-design 源文件”的显式项，尤其覆盖非交互静态 row、label、icon、chips 和 counts。
  - **落点**：plan-code-review skill
  - **优先级**：high
- 在 pixel parity README 或 plan gate 中明确：若 snapshot baseline 被 `.gitignore` 忽略，常规 gate 只能使用 DOM/geometry/computed/screenshot-smoke；要使用 `toHaveScreenshot` 必须保证 baseline 可在 clean checkout / CI 获取。
  - **落点**：frontend pixel parity README / spec-plan（本轮已补入 `frontend/README.md`，owner plan gate 已同步）
  - **优先级**：high
- 在 Playwright parity rerun 操作说明中加入 stale server 检查：修改 frontend bundle 后，若 `reuseExistingServer=true`，先确认 4173 server 是否需要重启。
  - **落点**：frontend README 或 pixel parity gate 文档（本轮已补入 `frontend/README.md`）
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高优先级：把 “ignored snapshot 不可作为常规 PASS” 沉淀到 frontend pixel parity 公共文档，避免其它屏幕复用同一脆弱证据。本轮已完成，后续需在新增 pixel gate 时沿用。
- 下一优先级：更新 `/plan-code-review` 的 UI parity 审查清单，要求从 ui-design 源文件反向枚举 DOM/文案/icon/source row，而不是只看历史 tests。
- 可延后：为 Playwright server stale dist 添加 helper 或 npm script；当前可通过手工检查 4173 端口规避。
