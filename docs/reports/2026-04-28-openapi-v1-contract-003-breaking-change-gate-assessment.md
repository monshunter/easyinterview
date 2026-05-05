# openapi-v1-contract/003-breaking-change-gate 交付复盘

> **日期**: 2026-04-28
> **审查人**: Claude

## 1 复盘范围与成功证据

- 范围：`docs/spec/openapi-v1-contract/plans/003-breaking-change-gate` 全 4 phase 交付（plan / checklist 已切到 `completed`，dev 当前 HEAD = `51aed11`）。
  - Phase 1 — `openapi/baseline/openapi-v1.0.0.yaml` v1.0.0 freeze 快照（`info.description` 改写成 `BASELINE — DO NOT EDIT` 含义说明，其余按位拷贝自 `openapi/openapi.yaml`）+ `make openapi-diff` 接入根 `Makefile`（默认按 SemVer 取 `openapi/baseline/` 下最大版本，`BASELINE_VERSION=v1.0.0` 显式覆盖；错误版本号 fail-fast）。
  - Phase 2 — `openapi/diff-config.yaml` 锁定禁止集合 / additive 集合 / privacy export `501→202` 白名单；`scripts/lint/openapi_diff.py` wrapper 直接实现 spec §4.4 规则，启动时打印 `tool=wrapper-1.0.0`，并通过 `git show <ref>:history.md` 与工作树对比新增行数实现「同 PR `history.md` 增量」检查。
  - Phase 3 — `docs/spec/openapi-v1-contract/decisions/TEMPLATE.md` ADR 模板（`OPENAPI-NNN-<short>` + 9 字段固定）+ `openapi/baseline/README.md` SemVer 升级阈值默认值（v1.0.x patch / v1.x.0 minor 累积 ≥ 5 个新 endpoint / v2.0.0 major 任何 breaking）+ `history.md` Header 升 1.6 + 在 Header 下追加「修订规则」章节。
  - Phase 4 — Phase 4.1 三段构造自检 + Phase 4.2 一键 chained gate + Phase 4.3 工作日志声明 W2 implementation 准入 gate B2 部分闭合 + Phase 4.4 plan/checklist Header → completed、plans/INDEX.md 003 行迁入「已完成」组。
- 通过的可执行证据（均在本会话现场跑过，HEAD 落点为 `51aed11`）：
  - `python3 -m unittest scripts.lint.openapi_diff_test -v` 21/21 OK，覆盖 spec §4.4 全部禁止集合（delete-endpoint / rename-path / change-method / delete-field / change-type / required-add 既有字段 / required-add 新字段 / 删除 enum value / required query 新增）+ 全部 additive 集合（new endpoint / new tag / new optional field / new string enum value / new optional query / new example）+ history-row 计数器 + 端到端 CLI（clean / 删字段 fail / optional pass / privacy 白名单 with-history pass / privacy 白名单 missing-history fail）。
  - Phase 4.1 在真实 `openapi/openapi.yaml` 上构造的三段自检（HEAD = Phase 1-3 commit `53a90a1` 作为 wrapper history-ref clean baseline）：删 `TargetJob.title` → exit 2、`{breaking:1}`、唯一 finding `field-deleted @ components.schemas.TargetJob.title`（spec C-4 ✅）；加 `PracticePlan.metadata`（optional） → exit 0、`{additive:1}`（spec C-5 ✅）；privacy `501→202` WITHOUT history.md 增量 → exit 2、`{breaking:1, informational:2}` 含 `history-not-incremented`（spec D-12 ✅）；同改动 + history.md 同 PR 加一行 → exit 0、`{informational:2}`，仅状态码 pair 命中 whitelist；revert 后 `make openapi-diff` summary 全 0。
  - Phase 4.2 一键 chained gate：`make codegen-check && make validate-fixtures && make openapi-diff` exit 0。`codegen-check` 输出 `openapi.yaml is valid` + `openapi inventory OK: 14 tags, 36 operations, ApiErrorResponse/IK/501/Provenance invariants enforced; B1 enums in sync.`；`validate-fixtures` 输出 `OK — 36 fixtures under openapi/fixtures`；`openapi-diff` 输出 `tool=wrapper-1.0.0 baseline=openapi/baseline/openapi-v1.0.0.yaml current=openapi/openapi.yaml` + `{summary: {breaking:0, additive:0, informational:0}, whitelistMatches: [], findings: []}`。
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check` 报告 `All documents are in sync. Zero drift detected.`（其他 W2/W3/P1 子 spec 的 `non-standard entry (待 spawn)` warnings 是预期 placeholder，不计入 003 close-out）。
  - 003 plan/checklist Header 状态由 `active` → `completed`；`plans/INDEX.md` 把 003 行迁入「已完成」组，与 001 / 002 同列。
- 关联资产：[plan](../spec/openapi-v1-contract/plans/003-breaking-change-gate/plan.md) / [checklist](../spec/openapi-v1-contract/plans/003-breaking-change-gate/checklist.md) / [openapi-diff wrapper](../../scripts/lint/openapi_diff.py) / [diff-config](../../openapi/diff-config.yaml) / [baseline README](../../openapi/baseline/README.md) / [ADR TEMPLATE](../spec/openapi-v1-contract/decisions/TEMPLATE.md) / [history.md 修订规则](../spec/openapi-v1-contract/history.md)。

## 2 会话中的主要阻点/痛点

- **Phase 4.1 第一次 replay 被 Phase 3.3 未提交的 `history.md` 编辑污染**
  - **证据**：Phase 1-3 全部完成但**未 commit** 时，我直接跑 Phase 4.1 #3a（privacy `501→202` 不带额外 history 增量），按规则期望 exit ≥1 + `history-not-incremented`。实际结果是 `make-exit=0`，summary `{informational: 2}`，**没有** `history-not-incremented` finding。原因：wrapper 默认 `--history-ref=HEAD`，对比的是 HEAD 的 history.md（7 行）vs 工作树（8 行：含 Phase 3.3 已写入但未提交的 1.6 row）；wrapper 看到「行数已增加」就认为白名单 gate 通过，直接降级为 informational。
  - **影响**：必须停下来先 `git commit` Phase 1-3（53a90a1），让 HEAD 正确包含 Phase 3.3 的 history.md 改动，再以新 HEAD 作为 history-ref 重跑 #3a。多花一次 commit 周期 + 一次 replay 周期；总计约 8 分钟。
- **`git checkout -- openapi/openapi.yaml docs/spec/openapi-v1-contract/history.md` 把 Phase 3.3 未提交的「修订规则」章节也回滚了**
  - **证据**：Phase 4.1 #3b 跑完后，我用一条 `git checkout -- <两个文件>` 试图回滚 self-check 的 TEMP 编辑。`openapi.yaml` 的 self-check 编辑是预期回滚，但同一条命令也把 Phase 3.3 在 history.md 写入的 Header 1.6、「修订规则」章节、1.6 row 一并撤销（因为它们都是 `history.md` 的 working-tree 改动，HEAD 还没有 commit 它们）。系统提醒提示文件被修改，回头读 history.md 确认全部丢失。
  - **影响**：必须重新手写 Phase 3.3 的 history.md 编辑（约 30 行），再走一次 commit。属于在前一痛点已经清楚「Phase 3 还没 commit」之后仍发生的二次副作用。
- **plan §1.2 把 OpenAPITools/openapi-diff CLI 写成默认实现，与 plan §2.3 / 风险表「wrapper 才是真理源」口径错位**
  - **证据**：plan §1.2 字面要求 `make openapi-diff` 「调用 `openapitools/openapi-diff` CLI（npm `openapi-diff`，或 java jar）；版本固定」；plan §2.3 又允许「如 `openapi-diff` 工具自身配置不足以覆盖 §2.1 / §2.2，落地 wrapper 脚本」并写明「最终退出码以 wrapper 为准」；plan §5 风险表也写「任何分歧以 spec §4.4 为准」。OpenAPITools/openapi-diff 没有可靠 npm 入口（npm `openapi-diff` 是已归档的 Atlassian 包，不是 OpenAPITools 实现），实战入口只剩 java jar；为了让本地 gate 不依赖外部网络下载，我直接在 wrapper 里实现 spec §4.4 规则，把外部工具留作未来可选项。
  - **影响**：实现路径与 plan §1.2 字面不完全一致，但与 §2.3 / §5 完全一致。判断成本约 5 分钟（确认 wrapper-only 路径不违反 plan）。spec §3.1 D-10 字面也是 OpenAPITools，所以同款不一致也存在于 spec 级。
- **plan §1.1 baseline marker 文本写「spec v1.3 锁定的 v1.0.0 freeze 快照」，但当前 spec.md Header 已是 v1.5**
  - **证据**：plan §1.1 给出 baseline `info.description` 的字面模板 `本文件是 openapi-v1-contract spec v1.3 锁定的 v1.0.0 freeze 快照。`；落地时实际仓库 spec.md 已经迭代到 v1.5（v1.4 为 001-bootstrap assessment remediation，v1.5 为 docs renderer 迁移；两次都不改 v1.0.0 contract surface）。我决定按 plan 字面写「spec v1.3」——因为 v1.3 才是把 36 endpoints / 14 tags / additive-only 规则锁死的版本，v1.4/v1.5 都是只改文案 / 治理的非契约修订。
  - **影响**：决定本身没改路径，但在 baseline marker 文本与 spec Header 之间引入了「为什么是 1.3 不是 1.5」的解释成本；如果后续读者看到 baseline 写 1.3 而 spec 已经 1.7+，需要再次推断「freeze 是哪一版」。本次没有触发返工。
- **history.md 在 Phase 3 / Phase 4 之间被反复编辑（修订规则 + 1.6 row + self-check 临时行 + revert）**
  - **证据**：本次会话对 `docs/spec/openapi-v1-contract/history.md` 的写操作发生 5 次：Phase 3.3 写入修订规则 + 1.6 row → Phase 4.1 self-check #3b 临时再加 1 行 → `git checkout` 回滚 → 再写 Phase 3.3 → 提交 53a90a1 → Phase 4.1 #3b 再次临时加行 → 再 `git checkout` 回滚（这次只剩 self-check 行因为 Phase 3.3 已经 committed 在 HEAD）。每次操作都是合法的，但「Phase 3 持久编辑 + Phase 4 self-check 临时编辑」共用同一个文件让 git checkout 边界很容易出事故（前一痛点已实证）。
  - **影响**：单次编辑都不长，但叠加起来造成至少一次返工。属于跟前两痛点同根因的体感放大。

## 3 根因归类

- **根因 A**：wrapper `--history-ref` 默认 `HEAD`，与「同 PR 增量」语义错位
  - **类别**：skill（`scripts/lint/openapi_diff.py`）+ spec-plan（plan 003 §2.2 / Phase 4.1）
  - **说明**：plan §2.2 的语义是「同 PR 增量」（i.e., 与父分支对比），但 wrapper 的实现是「与 HEAD 对比」。当当前会话还有未 commit 的 history.md 改动时，「与 HEAD 对比」会把这些改动当成增量，导致 self-check #3a 假阳性通过。这是 wrapper 设计上的轻微偏差，本会话靠手工先 commit Phase 1-3 解了；下次实施类似 plan 时若未先 commit，就会再次撞上。
- **根因 B**：history.md 同时承载「Phase 3 持久编辑」与「Phase 2/4 self-check 临时编辑」，git checkout 边界容易扩散
  - **类别**：spec-plan（plan 003 Phase 2.4 / Phase 4.1 编排）
  - **说明**：Phase 2.4 / Phase 4.1 self-check 的 #3 系列要求在 history.md 上构造临时增量，再 revert；同期 Phase 3.3 也在写 history.md 永久内容。当用 `git checkout` 撤销 self-check 临时编辑时，未 commit 的 Phase 3.3 改动也会一起被回滚。这是 plan 给两类编辑共用同一文件的隐性副作用。
- **根因 C**：plan §1.2（默认工具）与 plan §2.3 / §5（wrapper 才是真理源）口径不完全统一
  - **类别**：spec-plan（plan 003 §1.2）
  - **说明**：§1.2 把 OpenAPITools/openapi-diff CLI 写成 default，但实际 npm 没有可靠入口；wrapper 直接实现 §4.4 是更稳的本地 gate。spec §3.1 D-10 同款不一致。本次靠 §2.3 兜底落地，未来 reviewer 仍可能问「为什么没用 OpenAPITools 工具」。
- **根因 D**：plan §1.1 baseline marker 字面模板锁死 spec 版本号
  - **类别**：spec-plan（plan 003 §1.1）
  - **说明**：plan 字面给的 marker 模板写「spec v1.3 锁定的 v1.0.0 freeze 快照」，但 spec.md 在 plan 编写后又迭代到 v1.5（v1.4/v1.5 都是非契约修订）。落地时需要判断「写 v1.3 还是 v1.5」。本次靠语义判断（v1.3 是 contract lock 的版本）解决。
- **根因 E**：本会话单次 `git checkout -- <多文件>` 是一次性 execution mistake，不是 process defect
  - **类别**：无需仓库改动
  - **说明**：合理做法是 `git stash` 临时 self-check 编辑或单文件回滚；这次直接 `git checkout` 多文件是我的操作选择，不是 plan / skill 的缺陷。仅作记录。

## 4 对流程资产的改进建议

- **建议 1（medium）**：在 `scripts/lint/openapi_diff.py` 中把 `--history-ref` 的默认值从 `HEAD` 改成「当前分支与父分支（缺省 dev / main）的 merge-base」，并在 `openapi/diff-config.yaml` 的 `tooling` 段落新增 `historyDiffBase: dev` 配置项；保留 `--history-ref` 显式覆盖。理由：plan §2.2 字面写「同 PR 增量（通过 git diff 跨文件检查）」，merge-base 才是 PR 增量的正确锚点；HEAD 默认会把同会话 / 同分支但未 commit 的改动算作增量，触发本次假阳性。
  - **落点**：`scripts/lint/openapi_diff.py`（默认分支解析逻辑）+ `openapi/diff-config.yaml`（tooling.historyDiffBase）+ `openapi/baseline/README.md` Tooling 表（标注新默认值）
  - **优先级**：medium
- **建议 2（low）**：plan 003 / future plan 编写规范里加一句：当 Phase N self-check 与 Phase M 永久编辑触碰同一文件（典型如 history.md）时，self-check 应使用 `git stash` 而非 `git checkout` 撤销 working-tree 编辑，或显式要求「Phase 4.1 self-check 必须在 Phase 3.3 commit 之后再跑」。即使建议 1 落地，这一条仍是防御性的执行约束。
  - **落点**：spec-plan（`/design` skill 模板 + plan 0B4 table inventory.1 在下一次再同类计划时复用）
  - **优先级**：low
- **建议 3（low）**：把 plan §1.2 / spec §3.1 D-10 中「OpenAPITools/openapi-diff CLI 默认」改成「wrapper 直接实现 spec §4.4 规则；如未来引入 OpenAPITools / `oasdiff` 等外部 CLI，需在 `openapi/diff-config.yaml.tooling` 固定版本，且 wrapper 仍持有最终退出码」。本次靠 §2.3 兜底落地，但消除字面错位有助于后续 plan-review。
  - **落点**：spec-plan（spec §3.1 D-10 + plan 003 §1.2）
  - **优先级**：low（不影响行为，只消除文档歧义）
- **建议 4（low）**：plan §1.1 baseline marker 字面模板里的 `spec v1.3` 改成 `<当前 spec Header 版本>` 占位，并在 plan 旁批注「以落地时点 spec.md Header 版本为准」。或在 `/design` skill 模板里要求 plan 写 baseline marker 时不要硬编码 spec 版本。
  - **落点**：spec-plan（plan 003 §1.1，向 `/design` skill 模板回流）
  - **优先级**：low

## 5 建议优先级与后续动作

- **下一轮最值得做（medium）**：建议 1 — 把 wrapper history-ref 默认改成 merge-base，是真正减少未来同类假阳性 / 假阴性的根因修复，覆盖范围比文案修订大。可与未来一个 `make openapi-diff` 的小迭代一起做（如 `BASELINE_VERSION=v1.0.0` 之后再迭代 `HISTORY_REF=` 显式入口）。
- **同轮可一起做（low）**：建议 2 / 建议 3 / 建议 4 — 都是 plan / spec / skill 模板层面的文字调整，单次成本极小，可在下一次 `/plan-review` 或新建同类 plan 时一并处理；建议 2 哪怕建议 1 落地后仍值得保留作为「防御性执行约束」。
- **当前不做**：003 plan / checklist 已经全部 completed 且 merge 进 dev (`51aed11`)；本报告只产出建议，不直接修改任何 skill / spec / plan 资产。`make openapi-diff` 现有行为已被 21 单元测试 + Phase 4 实测 capture 锁住，不存在功能缺陷。
- **后续主路径**：B2 三个 child 全部 completed，[engineering-roadmap §5.7 W2 准入 gate](../spec/engineering-roadmap/spec.md#57-w2-implementation-准入-gate) 关于 B2 部分已闭合；C 全域与 D 全域 child 启动 W2 implementation 时直接消费 `backend/internal/api/generated/` 与 `frontend/src/api/generated/`，并由 `make codegen-check` / `make validate-fixtures` / `make openapi-diff` 三道本地 gate 拦截 drift。本 plan 不修改 [engineering-roadmap/001-decompose-subspecs 父 checklist](../spec/engineering-roadmap/plans/001-decompose-subspecs/checklist.md)，spec C-10 成立证据由 003 plan 持有。
