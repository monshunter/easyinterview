# Frontend Parse Loading Browser Gate 交付复盘报告

> **日期**: 2026-05-24
> **审查人**: Codex

## 1 复盘范围与成功证据

本次复盘范围是 `E2E.P0.015` 的 ready-response parse loading 浏览器级 gate：在 `getTargetJob` 首次返回 `analysisStatus=ready` 时，场景验证必须能证明正式前端先展示 `ui-design` loading demo，不能直接进入 parsed preview。

已通过的成功证据：

- `pnpm --filter @easyinterview/frontend exec playwright test tests/pixel-parity/parse.spec.ts --grep "ready target job response keeps ui-design loading demo before preview"`：desktop/mobile 2 tests PASS，并输出 `E2E.P0.015 ready-response loading browser gate screenshotBytes=46630` 与 `screenshotBytes=169796`。
- `E2E.P0.015` setup → trigger → verify → cleanup：PASS；trigger 覆盖 real-mode gate、home/parse Vitest、frontend build 与 focused Playwright browser gate。
- `verify.sh` 已强制检查 Playwright spec、test title、`screenshotBytes` marker 与 pass count，避免只依赖宽泛 PASS。
- `validate_context.py`、`sync-doc-index --check`、`make docs-check`、`git diff --check`：PASS。
- 交付 commit：`89da7fd3 test(frontend-home): add parse loading browser gate (BUG-0099)`。

## 2 会话中的主要阻点/痛点

### 2.1 P0.015 原场景缺少浏览器级过渡态证据

- **证据**：修复 BUG-0099 后，P0.015 已覆盖 Vitest 状态机，但 scenario trigger/verify 没有截图或浏览器 DOM marker 证明 ready response 不会跳过 loading demo。
- **影响**：长期回归风险仍集中在真实路由和浏览器渲染层；只看 unit/Vitest 容易再次漏掉用户可见中间态。

### 2.2 Playwright gate 需要显式 frontend build

- **证据**：`tests/pixel-parity` 依赖 `frontend/dist` 静态产物，scenario trigger 必须在 Playwright 前执行 `pnpm --filter @easyinterview/frontend build`。
- **影响**：如果脚本只追加 Playwright 命令，可能在本地旧 dist 或缺 dist 情况下得到不可靠结果。

### 2.3 verify 需要检查状态特异 marker

- **证据**：本次 `verify.sh` 增加了 `ready target job response keeps ui-design loading demo before preview` 和 `E2E.P0.015 ready-response loading browser gate screenshotBytes=` 检查。
- **影响**：否则即使 Playwright PASS，也难以确认跑到的是 ready-response loading 过渡态，而不是同文件里其他 parse parity case。

## 3 根因归类

- **spec-plan**：owner checklist 已有 ready-response Vitest gate，但浏览器级 route/DOM/screenshot gate 没有在 P0.015 场景资产中固化。
- **scenario README / scripts**：历史 scenario verify 更偏向脚本存在和总 pass count，缺少“状态名 + 证据 artifact marker”的约束。
- **无需仓库改动**：`LC_ALL=C.UTF-8` locale warning 与既有 Home test `act(...)` warning 在本次验证中仍出现，但不影响新增 browser gate 结果；这两个 warning 不属于本次变更范围。

## 4 对流程资产的改进建议

- 对 frontend UI transition regressions，scenario verify 应检查能命名具体状态的 marker，并关联截图、DOM anchor 或 trace artifact。
  - **落点**：scenario / README
  - **优先级**：high

- 若后续 `frontend/tests/pixel-parity` 中继续出现 fixture-backed API mock，可在至少 2-3 个 spec 复用后提取共享 helper。
  - **落点**：frontend test helper
  - **优先级**：medium

- 依赖静态产物的 browser scenario 应在 trigger 中显式 build，避免使用残留 dist。
  - **落点**：scenario scripts
  - **优先级**：medium

## 5 建议优先级与后续动作

最高优先级是审计相邻的 P0.016 parse confirm → workspace handoff 场景，确认它是否同样缺少浏览器级 route/context evidence。推荐先由 `/scenario-run E2E.P0.016` 承接当前验证事实；若发现只覆盖 Vitest 或脚本级断言，再原地补 focused Playwright gate。

可以延后处理的是 pixel-parity mock helper 抽象：当前只有一个新增 gate，直接内联 fixture response helper 更便于控制变更范围；等重复模式稳定后再提取。
