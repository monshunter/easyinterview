# CI Pipeline Docs Anchor Gate 交付复盘报告

> **日期**: 2026-05-04
> **审查人**: Codex

## 1 复盘范围与成功证据

- 交付范围：原地修订 `ci-pipeline-baseline/001-local-quality-gates`，把 docs/spec Markdown heading fragment anchor 审计纳入 `make docs-check`。
- 成功证据：`python3 -m unittest scripts/lint/check_md_links_test.py` 15 tests PASS；`python3 scripts/lint/check_md_links.py docs/spec --ignore '**/TEMPLATES.md' --check-fragments` PASS；`make docs-check` PASS；`validate_context.py --context docs/spec/ci-pipeline-baseline/plans/001-local-quality-gates/context.yaml --target repo` PASS；`sync-doc-index --check` zero drift；`git diff --check` PASS。
- 生命周期证据：`plan.md` / `checklist.md` 已回到 `completed` v1.5，`docs/spec/ci-pipeline-baseline/plans/INDEX.md` 已同步。

## 2 会话中的主要阻点/痛点

- `make docs-check` 原本只检查文件存在，不检查 `#fragment` 是否命中真实 heading。
  - **证据**：新增 Red 测试前，`scan_directory(..., check_fragments=True)` 不存在；测试以 `unexpected keyword argument 'check_fragments'` 失败。
  - **影响**：历史实施审查只能依赖一次性 ad hoc heading anchor audit，无法在常规本地 gate 中持续防止漂移。
- 新 gate 首次运行发现 3 个真实旧文档坏锚点。
  - **证据**：`event-and-outbox-contract/plans/001-bootstrap/{plan.md,checklist.md}` 中 3 处链接使用 `jobtype`，真实 GitHub-style slug 保留 `job_type` 下划线。
  - **影响**：说明现有文档链接检查覆盖不完整，但该问题是明显文档 typo，按 `docs/bugs/README.md` 可不建独立 Bug 记录。
- GitHub-style slug 规则有细节成本。
  - **证据**：`↔` 这类符号应移除，`job_type` 下划线应保留；通过 `github-slugger@2.0.0` spot check 后补入测试 fixture。
  - **影响**：没有测试锁定时，轻量自写 slugger 容易误报或漏报。

## 3 根因归类

- docs-check coverage gap。
  - **类别**：spec-plan / tooling
  - **根因**：A5 001 最早只要求相对路径存在性，没有把 docs/spec fragment anchor drift 纳入 `make docs-check` 的固定契约。
- 旧链接 typo。
  - **类别**：no repo change needed
  - **根因**：人工写入 heading slug 时漏掉 `_`，此前没有自动 gate 捕捉。
- slugger 细节不显式。
  - **类别**：tooling
  - **根因**：仓库没有现成 Markdown heading slug contract tests，直到本次实现才覆盖中文、符号、多 hyphen 与重复 heading。

## 4 对流程资产的改进建议

- 保持 `check_md_links.py --check-fragments` 只作为显式开关，不改变默认相对链接检查语义。
  - **落点**：tooling
  - **优先级**：high
- 后续新增复杂 Markdown heading 写法时，优先扩展 `scripts/lint/check_md_links_test.py` fixture，而不是把 parser 扩成 Markdown 格式检查器。
  - **落点**：tooling
  - **优先级**：medium
- 不为本次旧锚点 typo 创建 `BUG-NNNN`，在 work journal 中记录即可。
  - **落点**：no repo change needed
  - **优先级**：low

## 5 建议优先级与后续动作

- 最高价值：把本次 gate 随提交固定下来，让后续 `make docs-check` 自动覆盖 Header/INDEX、相对链接、docs/spec fragment anchor 三类文档漂移。
- 可延后：如未来 docs heading 大量引入 HTML、emoji 或更复杂 inline Markdown，再评估是否引入独立 slugger 依赖；当前最小实现已有测试覆盖本仓库实际用例。
