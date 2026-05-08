# 001 Home JD Import and Parse L2 Remediation 交付复盘报告

> **日期**: 2026-05-08
> **审查人**: Codex

## 1 复盘范围与成功证据

本次复盘范围是 `frontend-home-job-picks-and-parse/001-home-jd-import-and-parse` 的 L2 code review 修复收口，覆盖 Home recent jobs、JD import pending action、Parse footer privacy、scenario wrapper、backend build/codegen gate 与 owner 文档状态。

已通过的成功证据：

- `pnpm --filter @easyinterview/frontend test`：52 files / 324 tests PASS
- `pnpm --filter @easyinterview/frontend typecheck`：PASS
- `pnpm --filter @easyinterview/frontend build`：PASS
- `pnpm --filter @easyinterview/frontend test:pixel-parity`：68/68 PASS
- `E2E.P0.014 / P0.015 / P0.016 / P0.017 / P0.001 / P0.002 / P0.004 / P0.005 / P0.006`：全部 setup→trigger→verify→cleanup PASS
- `make validate-fixtures`、`make codegen-check`、`make build`、`make docs-check`：PASS
- `cd backend && go test ./internal/targetjob ./internal/platform/secrets -count=1`：PASS
- `BUG-0024` 已建档；owner plan/checklist/BDD checklist 恢复 `completed`，INDEX zero drift。

## 2 会话中的主要阻点/痛点

### 2.1 历史 completed 状态掩盖了当前语义漂移

- **证据**：原 checklist/BDD 状态为 completed，但 L2 复核发现 recent jobs 排序截断、pending import route params 隐私、parse footer provider 文案、P0.006 wrapper 计数、codegen literal 等当前 gate 问题。
- **影响**：必须把 plan 原地恢复 active，重新执行实现、场景、构建、文档同步后再回 completed。

### 2.2 场景 wrapper 仍绑定旧 pixel parity 总数

- **证据**：P0.006 Playwright 本体 68/68 PASS，但 `verify.sh` 仍 grep `48 passed`，导致 wrapper FAIL。
- **影响**：本体和 wrapper 结果矛盾，必须修 scenario README / expected outcome / verify script 并重跑完整 P0.006。

### 2.3 Git ignore 误伤源码包

- **证据**：`backend/internal/platform/secrets` 补包后 `make build` 通过，但 `git check-ignore -v` 显示根 `.gitignore` 的 `secrets/` 规则忽略该源码目录。
- **影响**：如果不放行该路径，构建修复无法进入版本控制，后续 checkout 仍会失败。

### 2.4 generated contract gate 暴露非前端残留

- **证据**：`make codegen-check` initially failed on targetjob event/job naked literals；修复后 generated events/jobs/OpenAPI gate PASS。
- **影响**：plan-code-review 的修复范围从前端 UI 扩展到 backend targetjob contract lint，增加跨目录验证成本。

## 3 根因归类

- **spec/plan**：旧 completed checklist 的证据注释偏结构化 PASS，没有强制 current truth source 的 privacy、provider-neutral、generated-contract 语义复核。
- **README / scenario docs**：P0.006 scenario 文档与 wrapper 没有随新增 home/parse/jd_match parity spec 同步，导致测试总数硬编码漂移。
- **AGENTS.md / repo hygiene**：新增含 `secrets` 名称的源码目录时，没有显式要求 `git check-ignore -v` 检查。
- **no repo change needed**：React `act(...)` warning 仍是测试 harness 噪声；当前 gates 不把它作为 hard fail，本次未扩大范围处理。

## 4 对流程资产的改进建议

- **建议 1**：在 `/plan-code-review` 或 `test/scenarios/README.md` 中补充规则：scenario verify 不应只硬编码历史总数，新增 spec 时必须同时断言 spec marker。
  - **落点**：skill / README
  - **优先级**：high

- **建议 2**：在 AGENTS.md 或 repo hygiene 规则中增加检查：新增源码目录名命中 `secrets` / `config` / `env` 等常见 ignore token 时，必须运行 `git check-ignore -v`。
  - **落点**：AGENTS.md
  - **优先级**：medium

- **建议 3**：L2 code review 修复完成前，把 `make codegen-check` 列为涉及 backend/openapi/shared/generated artifact 时的默认补充 gate。
  - **落点**：plan-code-review skill
  - **优先级**：medium

## 5 建议优先级与后续动作

最高优先级是修 scenario verify 规则：这次 P0.006 已经说明“本体 PASS、wrapper FAIL”会直接拖慢收口，而且容易在新增 parity spec 时复发。

下一步建议由 `/plan-review --fix frontend-home-job-picks-and-parse/002-jd-match-recommendations` 开始，先把 P1 Job Picks 推荐计划的 operation matrix、ui-design parity gate、scenario verify marker 和 backend recommendation API 依赖补齐，再进入实现；备选路径是先用 `/plan-code-review frontend-shell/003-ui-design-pixel-parity-gate --fix` 回补 P0.006 wrapper 的通用模板规则。
