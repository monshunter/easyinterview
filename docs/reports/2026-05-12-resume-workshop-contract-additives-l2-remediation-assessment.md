# Resume Workshop Contract Additives L2 Remediation 交付复盘报告

> **日期**: 2026-05-12
> **审查人**: Codex

## 1 复盘范围与成功证据

本次复盘覆盖 `plan-code-review` 后对三个 resume additive plan 的 L2 remediation：

- B2 OpenAPI 004：修复 `RegisterResumeRequest` / `ResumeAsset` 对 fileless resume intake 的 `fileObjectId` 必填漂移，补齐 paste / guided fixtures、generated Go/TS types 与 nullable codegen tests。
- B3 events 002：清除事件测试文件中的 retired `ResumeTailorMode` 字面量残留，使 checklist 3.1 的 negative search 当前可执行。
- Bug 建档：[BUG-0043](../bugs/BUG-0043.md) 记录 fileless intake / nullable codegen 漂移；B3 enum 主体沿用既有 [BUG-0042](../bugs/BUG-0042.md)。

已通过验证：

- `make codegen-openapi`
- `make lint-openapi`
- `make validate-fixtures`
- `make openapi-diff`
- `python3 -m pytest scripts/lint/openapi_inventory_test.py scripts/lint/validate_fixtures_test.py`
- `go test ./backend/cmd/codegen/openapi ./backend/internal/api/generated ./backend/internal/shared/events`
- `make lint-events`
- `pnpm --filter @easyinterview/frontend test src/lib/events/events.test.ts`
- `pnpm --filter @easyinterview/frontend typecheck`
- `git grep -nE 'ResumeTailorMode(Inline|Rewrite|Mirror)|"(inline|rewrite|mirror)"' -- shared/events.yaml shared/events/refs/ResumeTailorMode.json shared/events/baseline/events.v1.json shared/events/schemas/resume.tailor.completed.v1.json backend/internal/shared/events frontend/src/lib/events openapi/openapi.yaml openapi/fixtures` 返回 0 命中。

## 2 会话中的主要阻点/痛点

- **Fileless source path 没有成为 contract test 的一等断言**
  - **证据**：原 schema 和 generated DTO 仍要求 `fileObjectId`；`registerResume` 只有 upload 默认 fixture，`listResumes` 的 paste / guided items 仍带 file ID。
  - **影响**：Checklist 已完成但实际 paste / guided resume intake 被类型层阻断，需要补 OpenAPI、fixtures、generated artifacts 和 lint tests。

- **Primitive nullable 的 OpenAPI 表达方式与 diff gate 交互不直观**
  - **证据**：初始改成 `oneOf: [string, null]` 后，`make openapi-diff` 报 `composition-added` breaking；改为项目既有 `nullable: true` 后 diff gate 通过。
  - **影响**：修复路径需要返工一次，并暴露 OpenAPI plan/README 对 nullable schema style 的约定不够显式。

- **Generator nullable primitive 测试缺口**
  - **证据**：`nullable: true` 通过 schema validator 后，TS generated type 仍是 `fileObjectId?: string`；新增 generator red tests 后确认 `tsTypeFor` / `goTypeFor` 在 primitive switch 前未处理 nullable。
  - **影响**：fixture validator 与 generated consumer type 对同一 schema 的解释可能分裂，必须修 generator 而不只是改 schema。

## 3 根因归类

- Fileless source path 缺少语义 gate。
  - **类别**：spec-plan
  - OpenAPI 004 的 gate 更偏 operation coverage / fixture existence，缺少 `sourceType -> fileObjectId` 的 cross-field invariant。

- Nullable schema style 没有被写成执行规则。
  - **类别**：README / spec-plan
  - `openapi-diff` wrapper 已能处理 `nullable: true`，但 plan 没提示 primitive nullable 应优先使用该形态，避免 composition diff 噪声。

- Generator schema-expression matrix 不完整。
  - **类别**：test
  - Existing idempotency tests 能证明 codegen stable，不能证明 nullable / oneOf / enum / optional 等表达方式映射正确。

## 4 对流程资产的改进建议

- 在后续 OpenAPI additive plan/checklist 中增加 source discriminator cross-field gate。
  - **落点**：spec-plan
  - **优先级**：high
  - 示例：当 schema 新增 `sourceType` / `kind` / `mode` 时，必须有至少一个非默认 source fixture，并断言旧 required 字段不会阻断该 source。

- 在 `openapi/README.md` 或 OpenAPI plan 模板中写明 primitive nullable style。
  - **落点**：README
  - **优先级**：medium
  - 建议明确：primitive nullable 优先使用 `type: <primitive> + nullable: true`；只有真实 union 才使用 `oneOf`。

- 扩展 OpenAPI generator test matrix。
  - **落点**：test
  - **优先级**：medium
  - 当前已补 nullable string 单测；后续可追加 nullable enum、nullable `$ref`、nullable array/object、required nullable 与 optional nullable 的生成快照。

## 5 建议优先级与后续动作

最高价值后续动作：对 `openapi-v1-contract/004-resume-additive-coverage` 做一次小范围 `/plan-review --fix`，把 fileless source cross-field gate 和 nullable generator gate 固化回 plan/checklist，避免未来只靠本次新增测试隐式承接。

可延后动作：把 primitive nullable style 写入 `openapi/README.md`，并在下一轮 OpenAPI/codegen plan 统一扩展 generator schema-expression test matrix。
