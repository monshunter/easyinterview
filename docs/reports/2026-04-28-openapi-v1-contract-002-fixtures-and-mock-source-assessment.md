# openapi-v1-contract/002-fixtures-and-mock-source 交付复盘

> **日期**: 2026-04-28
> **审查人**: Claude

## 1 复盘范围与成功证据

- 范围：`docs/spec/openapi-v1-contract/plans/002-fixtures-and-mock-source` 全 4 phase 交付（plan / checklist 已切到 `completed`，dev 当前 HEAD = `bf51fb0`）。
  - Phase 1 — 落地 14 tag × 36 个 default fixture 与 `make validate-fixtures`。
  - Phase 2 — `make sync-fixtures-from-prototype` + `PROTOTYPE_MAPPING.md`，8 个 P0 闭环 endpoint 写入 `scenarios.prototype-baseline`。
  - Phase 3 — `make render-openapi-fixture-examples` 投影 + Prism 5 op smoke 字节级一致；同步删除 `openapi.yaml` 中 `requestPrivacyExport` 的 inline `example`，把 spec C-7 保护迁到 fixture 层。
  - Phase 4 — 全门禁复跑、`/sync-doc-index --check` zero drift、E1 W2 handoff 写入工作日志。
- 通过的可执行证据（均在本会话现场跑过）：
  - `make codegen-check` exit 0（lint-openapi + inventory + `git diff --exit-code` 全部干净）。
  - `make validate-fixtures` exit 0（spec C-6 / C-11 + 隐私 + UUIDv7 + 36 op 覆盖）。
  - `python3 -m unittest scripts.lint.validate_fixtures_test scripts.lint.validate_fixtures_cli_test scripts.codegen.sync_fixtures_from_prototype_test scripts.codegen.render_openapi_fixture_examples_test` 共 31 用例 OK。
  - `make sync-fixtures-from-prototype` + `make render-openapi-fixture-examples` 二次幂等，`git diff --exit-code -- openapi/fixtures openapi/.generated` 干净。
  - Prism 5.14.2 + Node v23.10.0 拉起 mock，`scripts/codegen/prism_fixture_smoke.py` 5 op byte-equal vs fixture（含 spec C-7 P0 `501 + error.code = "PRIVACY_EXPORT_NOT_AVAILABLE"`）。
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` zero drift；plans/INDEX.md 中本 plan 已迁到「已完成」组。
- 关联资产：[plan](../spec/openapi-v1-contract/plans/002-fixtures-and-mock-source/plan.md) / [checklist](../spec/openapi-v1-contract/plans/002-fixtures-and-mock-source/checklist.md) / [PROTOTYPE_MAPPING](../../openapi/fixtures/PROTOTYPE_MAPPING.md) / [openapi/fixtures/README](../../openapi/fixtures/README.md) / [render_openapi_fixture_examples.py](../../scripts/codegen/render_openapi_fixture_examples.py) / [prism_fixture_smoke.py](../../scripts/codegen/prism_fixture_smoke.py)。

## 2 会话中的主要阻点/痛点

- **`openapi.yaml` 残留 inline `example:` 与 fixture 真理源冲突**
  - **证据**：Phase 3.2 启动 Prism 后第一次 5 op smoke 中 `requestPrivacyExport` 返回 404 `Response for contentType: application/json and exampleKey: default does not exist`；翻 `openapi/.generated/openapi-with-fixtures.yaml` 才看到同一个 response 同时存在源 yaml 留下来的 `example:`（singular）与本 plan 投影的 `examples:`（plural），Prism 在两个 example 形态混合时挑不到 `default` 命名 example。
  - **影响**：触发一次跨 plan 的边界修正——删 openapi.yaml 的 inline example、改 `scripts/lint/openapi_inventory.py` 不再要求 example 字段、在 `validate_fixtures.py` 新增 `check_privacy_export_error_code`、跑 `make codegen-openapi` 重生成 backend / frontend artifacts；约 30 分钟的非主路径返工。
- **Prism 对非 2xx 命名 example 需要 `Prefer: code=<status>, example=default`**
  - **证据**：`requestPrivacyExport` 在 `Prefer: example=default`（不带 `code=501`）下返回 404 not_found；切到 `code=501, example=default` 立即返回 501 + 正确 body。原计划文本只写了 `Prefer: example=default`，没有提示非 200 的 code 子参数。
  - **影响**：Phase 3.2 多花约 5 分钟试错；后续 `prism_fixture_smoke.py` 与 fixtures README 同步补上 `code=<status>, example=default` 的写法。
- **`listExperienceCards` 的 prototype 数据节不在 `data.jsx`**
  - **证据**：plan 2.4 自检要求 8 个 P0 闭环 endpoint 全部 prototype-baseline 非空，但 `data.jsx` 没有 `experiences` 节；翻 UI 才发现 5 张经历卡 hardcode 在 `ui-design/src/screens-p1-depth.jsx#ExperienceLibraryScreen`。
  - **影响**：本会话补做了一次 UI prototype 数据搬家（把 5 张卡上提到 `data.jsx#experiences`），属于 plan 主路径外的判断；如果下次仍按「data.jsx 是唯一原型真理源」假设设计 plan，类似缺口会再触发一次副作用。
- **`/tdd` Step 9.5 phase-commit hook 误拒**
  - **证据**：Phase 4 提交 `bf51fb0` 成功落到 feature 分支后，连写 `git checkout dev && git merge feat/... --ff-only && git checkout feat/...` 被 hook 阻断，理由为「a git commit was never created for Phase 4 docs changes」。但 `git log` 明确显示 commit 已在 feature 分支；把 `git merge` 拆成单独一条 Bash 后立即通过。
  - **影响**：Phase 4 close-out 多一次切分命令的等待与说明；hook 看到 `git checkout dev` 把当前分支换成 dev，再去检查「当前分支有没有这个 commit」——dev 当时确实还没合，于是 false-negative。
- **`make codegen-check` 在未提交 generated 改动时 dirty-tree 失败（与 001-bootstrap remediation 报告同因再现）**
  - **证据**：Phase 3 把 `openapi/openapi.yaml` 的 inline example 删掉之后，`make codegen-openapi` 把改动传播到 `backend/internal/api/generated/openapi.yaml` 与 `frontend/src/api/generated/spec.ts`；这些改动尚未提交时 `make codegen-check` 因 `git diff --exit-code` 退出 1。提交 Phase 3 commit 之后立即恢复。
  - **影响**：与 [`2026-04-28-openapi-v1-contract-001-bootstrap-remediation-assessment.md`](./2026-04-28-openapi-v1-contract-001-bootstrap-remediation-assessment.md) §2 同因；本会话不重复改 skill / Makefile，但再增一笔证据。

## 3 根因归类

- **根因 A**：B2 子 plan 之间「inline example 归属权」从 001 移交到 002 在文档里不显式
  - **类别**：spec-plan
  - **说明**：001-bootstrap 当时为了让 spec C-7 在 contract 层面可见，在 `openapi.yaml` 写了 inline `example:`；002 plan §3.1 只说「不得人工手写 OpenAPI examples」，但没明确指示「Phase 3.1 实施时要先把 001 留下的 inline example 清掉」。结果靠运行 Prism smoke 才发现，属于交付边界没在文档层面交代清楚。
- **根因 B**：plan §3.2 的 Prism 用法只覆盖 200 响应族
  - **类别**：spec-plan
  - **说明**：plan §3.2 的固定 5 op 中 `requestPrivacyExport` 是非 200 响应，但 Prefer 命令样例没分场景；导致实施时必须回看 Prism 文档自己补 `code=<status>` 子参数。
- **根因 C**：原型数据真理源在 `data.jsx` 与各 `screens-*.jsx` 间分裂
  - **类别**：spec-plan
  - **说明**：plan §2.1 假设「data.jsx 是 prototype 唯一真理源」，但 UI 现状把 ExperienceLibrary / ResumeVersions 等多页屏幕的 mock 数据 hardcode 在组件文件里。当 sync 工具按计划读 data.jsx 时，自然出现 mapping 缺口。
- **根因 D**：Step 9.5 的 phase-commit chained 命令被 hook false-negative
  - **类别**：skill
  - **说明**：`/tdd` SKILL.md 推荐 `git checkout <base> && git merge <feature> --ff-only && git checkout <feature>` 作为一条 Bash；本环境的 hook 在中间 `checkout` 之后看 dev 分支没有 commit，就拒了第三步的 merge。这是 skill 推荐写法与本仓库 hook 的兼容性问题，不是逻辑错误。
- **根因 E**：`make codegen-check` 在未提交的 generated 改动上不友好
  - **类别**：spec-plan / skill（与 001-bootstrap remediation 报告同根因）
  - **说明**：本会话再次撞到同一个限制，没有更多新证据。属于已知项，提示频次累积应触发改进。

## 4 对流程资产的改进建议

- **建议 1（high）**：在 `docs/spec/openapi-v1-contract/plans/002-fixtures-and-mock-source/plan.md` 的 §3.1 起手处显式列出「先扫一次 `openapi.yaml`，把任何 hand-written `example:` / `examples:` 字段下放到 fixtures，再投影」。同时把同步整理 `scripts/lint/openapi_inventory.py`（不再要求 example 字段）写成 phase 子项，避免再发生「跑 Prism 时才发现冲突」。
  - **落点**：spec-plan（002-fixtures-and-mock-source plan + openapi-v1-contract spec §4.7 的 fixtures 单源声明可同步加一句「openapi.yaml 不持有 example」）
  - **优先级**：high
- **建议 2（medium）**：在 002 plan §3.2 / `openapi/fixtures/README.md` Prism smoke 命令矩阵里**按响应码族分组**写命令样例，特别是非 2xx 响应必须 `Prefer: code=<status>, example=default`。`scripts/codegen/prism_fixture_smoke.py` 已经按此规则实现，文档应同形态收口。
  - **落点**：spec-plan + README（plan §3.2 与 `openapi/fixtures/README.md` 的 Prism smoke 段落）
  - **优先级**：medium
- **建议 3（medium）**：在 002 plan §2.1 / `PROTOTYPE_MAPPING.md` 显式声明「若 prototype 真理源不在 `data.jsx`，sync 工具不允许跨文件抓取；要么把数据上提到 `data.jsx`（推荐），要么走 PROTOTYPE_MAPPING 的人工补节」。可同时让 UI 设计原型的目录级约定记录「mock 数据集中放 `data.jsx`」的工程约定。
  - **落点**：spec-plan（plan §2.1）+ 可选 UI 设计原型目录级 README
  - **优先级**：medium
- **建议 4（medium）**：调整 `/tdd` SKILL.md 中 Step 9.5 的样例命令——把 `git checkout <base> && git merge <feature> --ff-only && git checkout <feature>` 拆成 3 个独立 Bash 步骤；或在 `--auto` 模式里由 `/work-journal` 显式管理 merge，让本仓库的 phase-commit hook 看到与单步命令相同的现场。
  - **落点**：skill（`.agent-skills/tdd/SKILL.md`）
  - **优先级**：medium
- **建议 5（low）**：建议 4 与本会话第 5 个痛点同 reflect 的「`make codegen-check` 在未提交 generated 改动上的 dirty-tree 限制」已在 [`001-bootstrap-remediation-assessment.md`](./2026-04-28-openapi-v1-contract-001-bootstrap-remediation-assessment.md) §4 提过；本次只是累计第二笔证据，不重复出建议。等同类痛点再出现 1-2 次再触发流程改造（例如新增 `make codegen-openapi-verify` 仅做重生成 + lint，不做 `git diff --exit-code`）。
  - **落点**：spec-plan（openapi-v1-contract spec §4.5 / `Makefile` 设计）+ 可选 skill
  - **优先级**：low（保持已知项观察）

## 5 建议优先级与后续动作

- **下一轮最值得做（high）**：建议 1 — 在 002 plan / spec 中显式标注「openapi.yaml 不持有 example」并把 inline example 清理动作写进 Phase 3.1 子项；如果未来再有 plan 把 inline schema 元数据从 openapi.yaml 下放到 fixtures，可复用同一段落。
- **同轮可一起做（medium）**：建议 2 + 建议 3 + 建议 4 — 文字成本低、收益直接（Prism 多状态码用法 / prototype 真理源边界 / phase-commit 命令分步）；建议 4 还能减轻整个仓库的 phase-commit 摩擦。
- **延后处理（low）**：建议 5 — 等 `make codegen-check` dirty-tree 痛点累积到第 3 笔证据再正式触发 Makefile / skill 调整，避免过早抽象。
- **当前不做**：002 自身的 spec / plan / checklist 已经全部 completed，本报告只产出建议，不直接修改任何 skill / spec / plan 资产。
- **后续主路径**：[`openapi-v1-contract/003-breaking-change-gate`](../spec/openapi-v1-contract/plans/003-breaking-change-gate/plan.md) 仍处于 active，下一次 `/implement openapi-v1-contract/003-breaking-change-gate` 由 003 自己关闭 B2 freeze handoff。
