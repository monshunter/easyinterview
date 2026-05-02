# UI Design Context Selection Remediation 交付复盘报告

> **日期**: 2026-05-02
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 `ui-design` 与 `docs/ui-design` 的三类漂移：模拟面试规划页历史列表按当前规划 / 当前岗位过滤；复盘页移除 `收件箱` / `Inbox` 旧文案；复盘页 `目标岗位 / JD`、`关联模拟面试`、`绑定简历` 三个上下文动作改为本页选择弹窗。
- 新增 `ui-design/ui-design-contract.test.mjs`，覆盖当前规划历史取数、复盘上下文卡片不跨页跳转、当前 UI 源码不得暴露 `收件箱` / `Inbox` 文案。
- 通过验证：
  - `node --test ui-design/ui-design-contract.test.mjs`
  - `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`
  - `python3 scripts/lint/check_md_links.py docs/ui-design --ignore '**/TEMPLATES.md'`
  - `git diff --check`
  - `for f in ui-design/src/*.jsx; do npx --yes esbuild "$f" --loader:.jsx=jsx --format=iife --outfile=/tmp/$(basename "$f" .jsx).js >/dev/null || exit 1; done`

## 2 会话中的主要阻点/痛点

- `/change-intake` matcher 对 `ui-design` / `docs/ui-design` 查询给出了低置信且错误的 OpenAPI 计划候选。
  - **证据**：matcher 推荐 `openapi-v1-contract/002-fixtures-and-mock-source`，原因只命中 `packages=ui-design/src`，未识别用户显式给出的 `docs/ui-design`。
  - **影响**：需要人工回退到手动定位 UI 文档和源码，增加入口判断成本。
- 初始修复点集中在截图里的复盘页，随后全局检索才发现 `screens-p0-complete.jsx` 仍有 `收件箱` 文案。
  - **证据**：第一次契约测试只覆盖 debrief，后续 `rg "Inbox|收件箱" ui-design/src` 命中报告生成页。
  - **影响**：如果只按截图局部修复，废弃文案仍会从其他已加载页面回归。
- 全局 docs 链接检查当时存在与本次无关的既有断链。
  - **证据**：`python3 scripts/lint/check_md_links.py docs --ignore '**/TEMPLATES.md'` 报告顶层文档导航指向缺失的历史 UI 目录。
  - **影响**：当时只能用 `docs/ui-design` 子树链接检查作为相关范围证据；全局 docs gate 需要后续独立修复。

## 3 根因归类

- `change-intake` 对非 spec-centric 的 UI Design 文档缺少稳定识别路径。
  - **类别**：skill
- 废弃文案没有 repo 级或 UI 子树级契约测试，只依赖人工 grep。
  - **类别**：spec-plan
- `docs/README.md` 当时保留了历史 UI 目录链接，但当前仓库没有对应目录。
  - **类别**：README

## 4 对流程资产的改进建议

- 为 `/change-intake` matcher 增加 `docs/ui-design` 这类非 spec-centric 文档目录的识别能力，或在低置信且用户显式给出路径时优先报告该路径。
  - **落点**：skill
  - **优先级**：medium
- 将 UI 废弃文案、旧 route、跨页跳转禁用规则沉淀为可执行静态契约，而不是只写在文档里。
  - **落点**：spec-plan
  - **优先级**：high
- 独立修复 `docs/README.md` 的历史 UI 目录断链，恢复全局 `docs-check` 的可用性。
  - **落点**：README
  - **优先级**：medium

## 5 建议优先级与后续动作

- 最高价值：保留并扩展 `ui-design/ui-design-contract.test.mjs`，把后续 UI Design 决策都转成静态契约，避免再次靠截图人工发现漂移。
- 次高价值：修复 `docs/README.md` 的历史 UI 目录断链，让全局 docs 链接检查重新可作为收尾门禁。
- 可延后：增强 `/change-intake` matcher 对 `docs/ui-design` 的匹配权重，降低未来 UI 修订入口误路由。
