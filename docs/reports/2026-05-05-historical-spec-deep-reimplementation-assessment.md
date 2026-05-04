# Historical Spec Deep Reimplementation 交付复盘报告

> **日期**: 2026-05-05
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖 `historical-spec-implementation-review` 指定的现存 historical spec / plan 重新收口：先重新 Scope Inventory，再对 13 个真实可审目标逐项执行 artifact-level L1/L2 审查、必要修复和全局验证；`ai-gateway-and-model-routing/002-tools-streaming-and-stt` 仍按 draft-gated 排除，`historical-spec-implementation-review/001` 作为 orchestration plan 不作为业务实现目标。

成功证据包括：

- `make docs-check` PASS。
- `make codegen-check` PASS，B1/B2/B3 generated output 无 drift。
- `DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' make migrate-check` PASS。
- focused tests PASS：Python unittest 89、pytest 59、后端 A3/A4/B4 focused Go packages、前端 10 个契约 test files / 49 tests。
- `make lint` PASS；gitleaks 本机未安装时按 A4 既定策略跳过二层扫描。
- `make test` PASS；后端 Go packages 与前端 Vitest 全部通过。
- `make build` PASS；前端 build 仍是 D1 `frontend-shell` placeholder，当前按 Makefile 合同 exit 0。

## 2 会话中的主要阻点/痛点

- 历史 completion / PASS 标记会掩盖新版 product / UI spec 之后的语义漂移。
  - **证据**：用户明确指出“15/15 context PASS 都是新方案修订之前实施的”，要求忽略既有状态；本轮随后发现 OpenAPI fixture model provenance、B3 drift path、A4 runtime allowlist、A3 model-family 派生等旧 gate 未覆盖的问题。
  - **影响**：如果只按 checklist 状态或历史测试继续推进，会把旧产品/交互假设误判为仍然对齐。

- 早期 review 深度不足，容易把“小 diff”误当成“语义正确”。
  - **证据**：OpenAPI diff 不大，但用户质疑后才追加结构化 `openapi_inventory.py` product-scope guard，锁住 12 tag / 34 operation 之外的 current UI/product 语义。
  - **影响**：contract 数量不变时，独立 Mistakes/Growth/Voice/Drill、非 session-scoped report、旧 practice enum 等语义漂移可能不会被传统 diff gate 捕获。

- 生成物 drift gate 的路径边界曾过宽。
  - **证据**：`make codegen-check` 先失败于 B3 `codegen-events-check` 把手写 `frontend/src/lib/events/events.test.ts` 当成 generated output；修复后 Makefile 只 diff generator 实际输出文件，并新增 dry-run 回归测试。
  - **影响**：质量门禁会误伤手写测试，未来也可能让团队绕过 codegen-check。

- 本地 dev stack 的正确性必须靠真实 lifecycle，而不是只看 compose config。
  - **证据**：本轮实际执行了 idempotent `dev-up`、doctor、volume persistence、reset abort/force、redis-down failure、port conflict failure 和恢复。
  - **影响**：只跑 `docker compose config` 看不出 dry-run 安全、端口冲突诊断和数据卷边界。

## 3 根因归类

- Artifact-level review 规则没有一开始就强制执行。
  - **类别**：spec-plan / skill
  - **根因**：历史 `/plan-code-review` 执行习惯偏向 checklist 映射和现有 gate 复跑，未把新版 product/UI 不变量、旧口径负向搜索、生成物反查作为每个 target 的准入规则。

- 旧 gate 偏结构数量，语义覆盖不足。
  - **类别**：spec-plan
  - **根因**：B2/B3/B4/A4/A3 的既有 gate 能证明 shape 和 drift，但缺少对旧 route、旧 feature flag、旧 model/vendor 假设、旧 DB/event 字段的统一 negative guard。

- 报告/日志链接过早写成已完成。
  - **类别**：spec-plan
  - **根因**：`BUG-0006` 初始记录曾指向尚未创建的 work-journal，并提前写了 global pass；A5 docs-check 及时失败并迫使修正。

## 4 对流程资产的改进建议

- 把“diff 大小不是证据、历史 PASS 不算完成、每项必须 artifact-level 反查、每项必须旧口径负向搜索”固化进 `/plan-code-review` skill。
  - **落点**：skill
  - **优先级**：high

- 为 contract / migration / config / AI gateway plan 增加统一的 product-scope semantic lint checklist。
  - **落点**：spec-plan
  - **优先级**：high

- 让 codegen drift gate 明确声明 generated manifest，禁止用目录级 diff 混入手写测试和 fixtures。
  - **落点**：spec-plan / README
  - **优先级**：medium

- Scope Inventory 输出应把 draft-gated、docs-only orchestration、completed-but-reopened、true implementation target 分区显示。
  - **落点**：skill / shared script
  - **优先级**：medium

## 5 建议优先级与后续动作

最高优先级是把本轮台账 §1.1 的执行规章移入 `/plan-code-review` 和 future historical review plan 模板，避免下一轮再退回浅层核对。

其次是将本轮新增的语义 lint 思路推广到后续 P0 workstream：任何新 child plan 在进入实现前，必须先声明它依赖的 product/UI 不变量和旧口径负向搜索范围。
