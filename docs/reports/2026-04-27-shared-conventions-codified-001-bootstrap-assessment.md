# shared-conventions-codified/001-bootstrap 交付复盘报告

> **日期**: 2026-04-27
> **审查人**: Claude

## 1 复盘范围与成功证据

- 范围：B1 `shared-conventions-codified` 唯一 plan `001-bootstrap` 的 W0 跨语言共享真理源落地（plan 三件套：[plan](../spec/shared-conventions-codified/plans/001-bootstrap/plan.md) / [checklist](../spec/shared-conventions-codified/plans/001-bootstrap/checklist.md) / [context](../spec/shared-conventions-codified/plans/001-bootstrap/context.yaml)），交付 `shared/conventions.yaml` 真理源、Go 共享 module（`backend/internal/shared/{types,errors,idx}/`）、TS 共享 lib（`frontend/src/lib/{conventions,ids}/`）、跨语言 generator（`backend/cmd/codegen/conventions/`）、UUIDv7 + Idempotency-Key 双端工具、`UPPER_SNAKE_CASE` 错误码本地 lint gate（`scripts/lint/{conventions_yaml,error_codes}.py`）。
- 成功证据：
  - Checklist 11/11 全部勾选；plan + checklist Header `状态` 已切到 `completed`，`docs/spec/shared-conventions-codified/plans/INDEX.md` 通过 `sync-doc-index --fix-index` 自动迁组到 `## 2 已完成（Completed）`。
  - Phase 4.1：`make codegen-conventions` 跑两次后 `git status --short` 干净；删除 `backend/internal/shared/types/enums.go` 后再跑 generator 文件被完整还原。
  - Phase 4.2：`go test ./backend/internal/shared/... -count=1 -v` 18 用例全部通过（idx 6 + idempotency 7 + errors 5）；`go vet ./backend/...` 0 warning。
  - Phase 4.3：`pnpm --filter @easyinterview/frontend exec tsc --noEmit` exit 0；`vitest run` 3 test files / 24 tests 全部通过（ids 9 + idempotency 11 + enums 4）。
  - Phase 4.4：`docs/spec/INDEX.md` 中 B1 行匹配 spec Header `1.0 / active / 2026-04-26`；`sync-doc-index --check` 报告 `All documents are in sync. Zero drift detected.`；`engineering-roadmap/001-decompose-subspecs/checklist.md` 自 W0 spawn 起未被改动。
  - 治理 lint：`make lint` 输出 `OK: shared/conventions.yaml (14 enum types, 6 error codes)` + `OK: 6 Go constants, 6 TS entries; boundary clean`；4 项 black-box 注入（clean / Go 小写 / TS key/value 不一致 / 边界违规）均按预期分别 exit 0 / exit 1。
  - Git 流水：feature branch `feat/shared-conventions-codified-001-bootstrap-0426` 4 个 phase commit + 1 个 lifecycle close commit 通过 `git merge --ff-only` 全部 ff-merge 回 `dev`；外加 1 个 `dev` 上的预清理 commit `docs(governance): pin default parent branch to dev`。

## 2 会话中的主要阻点/痛点

- **plan 步骤顺序与依赖图不一致**
  - **证据**：plan §3 中 1.1（YAML）→ 1.2（generator at `backend/cmd/codegen/conventions/`）→ 2.1（`backend/go.mod`）。1.2 的 generator 必须能 `go run`，强依赖 2.1 的 module 初始化；按字面顺序执行 1.2 时 `go run` 会报 `cannot find main module`。会话内额外发出一段「natural ordering wrinkle」说明并把执行顺序调整为 1.1 → 2.1 → 1.2，Phase 1 commit 因此把 2.1 work 一并捎带。
  - **影响**：每个 phase commit 边界与 plan checklist 分组之间出现轻微错位；后续 reader 看 git log 时需要额外推理为何 Phase 1 commit 包含了 `backend/go.mod`。
- **plan §1.2 与 §2.2 在 `APIError` 归属上口径不一**
  - **证据**：plan §1.2 列出 generator 渲染 `http_dto.go (PageInfo / APIError)`，§2.2 又写「在 `backend/internal/shared/errors/` 中手写 `APIError struct` 基类与 `Wrap()` helper」。同一个 `APIError` 同时出现在「generator 输出」和「手写代码」两条所有权链上。
  - **影响**：在落地 §1.2 generator 与 §2.2 手写代码之间需要一段额外的设计判断（最终选定：Go side `APIError struct` 在 errors 包手写、TS side `ApiError interface` 通过 generator 写到 `conventions/errors.ts`）；判断结论留在 work-journal 「执行偏差记录」里，未回写到 plan / spec。
- **`13 类枚举` vs 实际 14 个生成类型的描述不一致**
  - **证据**：plan §1.1 与 §6 多处写「13 类枚举」，与 `B1 shared-conventions-codified §5` 13 个章节数对齐；但 §5.13「隐私请求类型 / 状态」一段同时列出两组并行枚举值（`export/delete` + `queued/processing/...`），最终 generator 必须输出 14 个独立 Go / TS 类型。
  - **影响**：generator + 校验脚本（`scripts/lint/conventions_yaml.py`）需要额外维护一条「13 source sections / 14 generated types」的映射注释，并在校验中显式断言 §5.1..§5.13 全覆盖。任何 reader（包括未来的 B2 / D1 / C 全域）都得理解这条隐含规则。
- **spec C-4 跨语言 idempotency 验收无对应 checklist 项**
  - **证据**：spec §6 C-4 要求「Go 与 TS 双端工具产出格式一致的 key」；plan §2.2 提到「`frontend/src/lib/conventions/idempotency.ts` 与 Go 端对偶」，但 plan §3 与 checklist 中的 2.4 文本只覆盖 TS 侧 `Idempotency-Key 24h TTL 工具`，没有任何 Go 侧 idempotency item。
  - **影响**：本次会话内主动补出 `backend/internal/shared/idx/idempotency.go` + 7 项单元测试以满足 C-4，并把这一偏差登记到 work-journal 「执行偏差记录」。后续若没有人读到该执行偏差，下一份引用 C-4 的 plan 仍可能再次踩到同一个跨语言对偶缺漏。
- **TS 工具链本应在 §2.3 / §2.4 之前就位，目前直到 4.3 才落齐**
  - **证据**：plan §2.3 描述 `frontend/package.json` 仅含 `build`/`lint`/`test` 占位脚本 + `uuid >=10`；plan §4.3 才要求 `tsc --noEmit` 与 vitest 用例。会话内 §2.4 实施后只能用 Python 结构化扫描代替真正的 TS 编译/测试 Red 阶段，TS 真正的编译错误（`noUncheckedIndexedAccess` 触发的 `parseIdempotencyKey` 类型问题）只能延后到 §4.3 才暴露并修复。
  - **影响**：TDD 在 §2.4 阶段缺一个真实可执行的 Red→Green 节拍；§4.3 一次性补齐 typescript / vitest devDep + tsconfig + 3 份 .test.ts，并发现一处类型窄化问题。如果 §2.3 直接落 typescript + tsconfig，§2.4 就能用 `tsc --noEmit` 作为 Red gate。

## 3 根因归类

- **plan 依赖顺序与渲染顺序耦合**：plan §3 把「先把真理源准备好」（1.x）放在「先把 module 初始化」（2.1）之前，但 1.2 的 generator 必须能 `go run`，会跨 1→2 触发 module 依赖。这是 plan 排序问题，不是一次性执行偏差。
  - **类别**：spec-plan
- **`APIError` 归属同时挂在 generator 与手写两条链路**：spec / plan 没有显式说明哪一侧是真理源。这是 plan / spec 文字层面的口径不一致。
  - **类别**：spec-plan
- **`13 类枚举` 描述与 generator 真实输出（14 类）数字不对齐**：plan / spec 直接复用上游文档的「13 类」表述，而上游 §5.13 实质包含 2 个并行枚举。这是 plan / spec 描述层面的精度问题，不是 generator 实现错误。
  - **类别**：spec-plan
- **C-4 验收要求与 plan checklist 双语对偶要求脱节**：spec §6 C-4 要求双语对偶，plan §2.4 与 checklist 仅覆盖 TS 侧。这是 plan checklist 漏项，不是 spec 不必要。
  - **类别**：spec-plan
- **TS toolchain 落地节奏与 §2.x 红绿节拍不匹配**：plan 把 typescript / tsconfig / 测试 runner 全部放在 §4.3，使 §2.4 缺少 native Red gate；这是 plan 工序设计问题，没有触发实际 bug，但减弱了 TDD 节拍。
  - **类别**：spec-plan

> 备注：本次会话内的若干一次性偏差（gofmt 对齐导致的 substring 测试不匹配、`AllStatuss` 复数命名修正、`noUncheckedIndexedAccess` 在 destructuring 上的类型窄化、shell session cwd 偶发漂移）属于执行细节，不构成流程缺陷，不在本节列出。

## 4 对流程资产的改进建议

- 在 B1 / 后续 child spec 的 plan 模板中明确「步骤顺序必须是依赖序」的硬约束，并给出本次的反例：generator 落点早于 module init 时必须把 module init 提前。下一次写新 plan 时由作者负责自检。
  - **落点**：spec-plan（`docs/spec/shared-conventions-codified/spec.md` 与 `docs/spec/shared-conventions-codified/plans/001-bootstrap/plan.md`，可在 plan §3 顶部加一条 `## 3.0 步骤排序原则` 一句话规范）
  - **优先级**：medium
- 在 B1 spec / plan 中显式声明 `APIError` 在 Go 侧的归属（手写、errors 包），TS 侧由 generator 写到 `conventions/errors.ts`，避免下游 child（B2 / C 全域）再绕一圈做同样判断；同步在 §1.2 的描述里把 `http_dto.go` 的输出收敛到「PageInfo only」。
  - **落点**：spec-plan
  - **优先级**：medium
- 把「13 类枚举」描述统一改写为「13 个 §5 sections / 14 个生成类型」并在 spec §3 / plan §1.1 各加一条说明；同步在 `shared/conventions.yaml` 顶部注释（已添加）保持一致。这样未来引用 B1 的 plan / generator 测试不会再为同一段反差打补丁。
  - **落点**：spec-plan
  - **优先级**：low（已用 YAML 注释 + 校验脚本兜住，但文档统一仍有价值）
- 在 plan checklist 中增加一条 Go 侧 idempotency item（例如 `2.5 在 backend/internal/shared/idx/ 实现与 TS 端字节级一致的 Idempotency-Key 工具`），让 C-4 验收有显式落点；或者在 spec §6 C-4 注释一句「双端实现由 plan 2.4 的 Go/TS 配对项各自落地」。
  - **落点**：spec-plan
  - **优先级**：medium
- 把 typescript / tsconfig / vitest 的最小 devDep 落点上移到 plan §2.3 的「最小可运行 frontend workspace」目标里（即 §2.3 不仅声明 `package.json`，也安装 typescript + tsconfig 让 `tsc --noEmit` 能跑），从而 §2.4 的 TS 实现可以用 `tsc --noEmit` 与最小测试做 Red→Green。§4.3 仍保留为 verification 总跑一遍。
  - **落点**：spec-plan
  - **优先级**：low（不影响交付正确性，提升 TDD 节拍）

## 5 建议优先级与后续动作

- 下一轮最值得实施：**plan 排序原则 + APIError 归属说明 + checklist 跨语言对偶项**。这三条加起来覆盖了本次会话最大的设计判断成本，落点都在 B1 spec/plan 自身，改动幅度小，但能让后续引用 B1 的 9 份 W1 child spec / B2 OpenAPI / D1 frontend-shell 不再重新踩同一个判断。建议作为 W1 进 `/plan-review` 之前的 B1 spec 修订项打包提交。
- 次优：**`13 类 → 14 类型` 描述统一**。已经在 YAML 注释 + lint 脚本兜底，但文档层面的不一致会让未来 reader 反复对账。
- 可以延后：**TS 工具链上移到 §2.3**。改动需要重新决定 `package.json` 的 devDep 边界，且当前一次性补齐已经过 `tsc --noEmit` + 24 vitest 用例验证，没有遗留风险。
- 不建议改动：本次会话中的一次性执行偏差（gofmt 对齐、`Allxxxs` 复数命名、`noUncheckedIndexedAccess` 类型窄化、cwd 漂移）均为单点偏差，未构成系统性缺陷。
