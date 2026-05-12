# Resume Workshop Contract Additives 交付复盘报告

> **日期**: 2026-05-12
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖三个 spec-centric plan：

- [event-and-outbox-contract/002-resume-tailor-mode-drift-fix](../spec/event-and-outbox-contract/plans/002-resume-tailor-mode-drift-fix/plan.md)
- [db-migrations-baseline/002-resume-versions-additive](../spec/db-migrations-baseline/plans/002-resume-versions-additive/plan.md)
- [openapi-v1-contract/004-resume-additive-coverage](../spec/openapi-v1-contract/plans/004-resume-additive-coverage/plan.md)

交付结果已通过 4 个提交收口：`95e9d22 fix(events): align resume tailor mode contract`、`6310c81 feat(db-migrations): add resume version schema`、`510636d feat(openapi): add resume workshop contract coverage`、`cab6e9c docs(plans): close resume additive lifecycle`。三个 plan / checklist 均已切换为 `completed`，对应 plans/INDEX 与 work journal 已同步。

通过的验证证据：

- B3 event：`make lint-events`、`make codegen-events && make lint-events`、Go/TS focused event tests、旧 `inline` / `rewrite` / `mirror` 字面量负向搜索、`openapi_inventory.py`、`sync-doc-index --check`。
- B4 DB：`make migrate-up && make migrate-down && make migrate-up`、`make migrate-check`、`go test ./internal/migrations/...`、`python3 -m pytest scripts/lint/migrations_lint_test.py`、`migrations/lint.sh`、privacy deletion matrix / enum-source / DDL 负向检查。
- B2 OpenAPI：`make codegen-check`、`make validate-fixtures`、`make lint-openapi`、`make openapi-diff`、`make docs-check`、`python3 -m unittest scripts.lint.openapi_inventory_test scripts.lint.validate_fixtures_test scripts.lint.conventions_yaml_test`、`cd backend && go test ./cmd/codegen/openapi ./cmd/codegen/conventions ./internal/shared/types ./internal/shared/errors -count=1`、`pnpm --filter @easyinterview/frontend typecheck`。
- 提交后复查：`git status --short` 干净，latest commit message ASCII 校验通过，`make codegen-check` 与 `make docs-check` 通过。

## 2 会话中的主要阻点/痛点

- **痛点 A：三条 plan 之间存在真实前置依赖，必须先 B3/B4 后 B2**
  - **证据**：OpenAPI 004 checklist 的 cross-plan prerequisite 需要 B3 `ResumeTailorMode` 和 B4 `resume_versions` persistence 先落地；本会话最终按 `events → migrations → openapi → lifecycle` 拆成 4 个提交。
  - **影响**：这是正确依赖，不是实现错误；但多 plan 执行时如果不明确 owner/order，容易把 OpenAPI schema 建在未落地的 enum 或 migration 假设上。

- **痛点 B：生成器里存在硬编码 B1 enum 数量和 P0 export 例外描述**
  - **证据**：B1 conventions 从 14 个 enum 扩到 17 个 enum 后，需要修订 `backend/cmd/codegen/openapi/b1_sync.go` 与 `render_go.go`，把注释从固定数量改为动态来源；P0 export 例外也需要允许 privacy export 与 resume export 两个 endpoint。
  - **影响**：如果只改 YAML 和 generated artifacts，`make codegen-check` 会重新引入旧注释或造成 drift。

- **痛点 C：plan checklist 中的 swagger-cli 命令与仓库 Makefile 实际用法不一致**
  - **证据**：原 checklist 写法 `npx @apidevtools/swagger-cli@4.0.4 swagger-cli validate ...` 在本地表现为 invalid arguments；实际通过的命令是 `npx -p @apidevtools/swagger-cli@4.0.4 swagger-cli validate ...`，并已在 checklist 中同步为通过命令。
  - **影响**：同一个 gate 语义正确，但命令形态不准会造成无意义的失败排查。

- **痛点 D：Resume Workshop prototype fixture mapping 明确缺位，只能在本 plan 记录 N/A**
  - **证据**：`rg screen-resume-workshop|listResumes|Resume Workshop|resume openapi/fixtures/PROTOTYPE_MAPPING.md scripts/codegen/sync_fixtures_from_prototype.py` 0 命中；OpenAPI 004 只能按 plan 3.9 记录 mapping gap，而不是生成 `prototype-baseline`。
  - **影响**：当前 API fixture 可验证，但未来正式 Resume Workshop frontend owner 若要求 prototype-backed baseline，需要先补 mapping 规则。

## 3 根因归类

- **根因 A：跨 plan order 由 checklist 前置信号表达，但没有单独的 multi-plan owner lane**
  - **类别**：spec-plan

- **根因 B：generator templates 混入了会随 B1 truth source 改变的常量叙述**
  - **类别**：skill / spec-plan

- **根因 C：plan 中直接写第三方 CLI 命令时没有复用 Makefile 中已验证的命令形态**
  - **类别**：spec-plan / README

- **根因 D：mock fixture prototype mapping 只覆盖既有画板入口，未覆盖 Resume Workshop 新原型数据源**
  - **类别**：spec-plan / README

## 4 对流程资产的改进建议

- **建议 1：multi-plan `/implement` 会话在 context 解析后生成显式执行顺序摘要**
  - **落点**：`implement` skill 或对应 plan checklist 前置段
  - **优先级**：medium

- **建议 2：OpenAPI/codegen plan 增加 generator template hardcode audit**
  - **落点**：`openapi-v1-contract` 后续 plan 模板或 checklist gate
  - **优先级**：high

- **建议 3：第三方 CLI gate 优先引用 Makefile target 或 README 中已验证命令**
  - **落点**：`docs/spec/README.md` plan 编写约定，或 `openapi/README.md` gate 表
  - **优先级**：medium

- **建议 4：Resume Workshop frontend owner 启动前先决定是否补 `PROTOTYPE_MAPPING.md`**
  - **落点**：`frontend-workspace-and-practice/001` handoff 或未来 `frontend-resume-workshop` plan
  - **优先级**：medium

## 5 建议优先级与后续动作

最高价值的后续动作是启动 workspace/frontend owner 对 `listResumes` 的原地修订：`frontend-workspace-and-practice/001-workspace-and-interview-context` 已记录 OpenAPI 004 unblock 链接，下一步可把 disabled-list 改为 active-list，并决定 Resume Workshop prototype mapping 是否进入同一轮。

其次，建议在下一个 OpenAPI/codegen plan 前补一条 generator template hardcode audit gate，重点检查 enum 数量、error code 数量、P0 exception wording、operation count 这类会随 truth source 变化的叙述。
