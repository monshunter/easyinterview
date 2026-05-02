# UI Design Resume UX Revision 交付复盘报告

> **日期**: 2026-05-02
> **审查人**: Codex

## 1 复盘范围与成功证据

本次交付覆盖 `ui-design/src/screen-resume-workshop.jsx` 的简历模块 UX 修订，以及 `docs/ui-design/` 中对应目标文档同步。

已通过的验证：

- `node --test ui-design/ui-design-contract.test.mjs`，6 项通过。
- `npx esbuild` 逐个编译 `ui-design/src/*.jsx`，全部通过。
- `python3 scripts/lint/check_md_links.py docs/ui-design --ignore '**/TEMPLATES.md'`，通过。
- `python3 .agent-skills/sync-doc-index/scripts/sync-doc-index.py --check`，Header / INDEX zero drift。
- `docs/ui-design` 自定义 INDEX/Header 校验，13 行通过。
- `git diff --check`，通过。
- `bash -n ui-design/run.sh`，脚本语法通过，仅出现本机 locale warning。

## 2 会话中的主要阻点/痛点

1. `change-intake` 自动匹配没有命中本次 `ui-design` 简历模块上下文，而是推荐了通用 `engineering-roadmap/001-decompose-subspecs` 文档计划。
   - **证据**：matcher 返回 `confidence=medium`，推荐目标为 `docs/spec/engineering-roadmap/plans/001-decompose-subspecs/context.yaml`，与 `ui-design/src/screen-resume-workshop.jsx` 和 `docs/ui-design` 修订不匹配。
   - **影响**：需要人工判定归属，改为手动限定在静态 UI 与 `docs/ui-design` 文档范围。

2. 早前简历 IA 修订缺少“静态原型动作必须有可见结果”的契约约束。
   - **证据**：本次发现 `查看原件`、预览区导出/复制、分叉创建版本等按钮存在 no-op 或仅切换 tab 的行为。
   - **影响**：UI 看起来完整，但用户实际点击后反馈不足，降低简历模块的流程可信度。

## 3 根因归类

1. `change-intake` 检索粒度偏 spec-centric，对 `docs/ui-design` + `ui-design/src/*` 这类静态原型修订缺少更精确的候选归属。
   - **类别**：skill

2. UI 文档已有模块结构和边界，但没有把“按钮不得空动作、创建动作需体现结果、定制版本默认进入决策面”写成可验证约束。
   - **类别**：README / spec-plan

3. 原契约测试覆盖了移除项和跨页导航风险，但没有覆盖简历详情按钮、默认 tab 和原件预览这类用户体验关键路径。
   - **类别**：no repo change needed，本次已补契约测试。

## 4 对流程资产的改进建议

1. 在 `change-intake` 的匹配策略中补充 `docs/ui-design`、`ui-design/src`、`resume_versions`、`screen-resume-workshop` 等静态 UI 主题的直接归属规则。
   - **落点**：skill
   - **优先级**：medium

2. 在 `docs/ui-design/README.md` 的检查清单中补充一条：静态 UI 中可点击的业务动作必须有可见反馈、状态变化或明确禁用说明。
   - **落点**：README
   - **优先级**：high

3. 后续简历模块继续调整时，优先扩展 `ui-design/ui-design-contract.test.mjs`，把默认入口、来源追溯和导出/复制/创建反馈作为回归断言。
   - **落点**：no repo change needed
   - **优先级**：medium

## 5 建议优先级与后续动作

最高价值后续项是把“静态 UI 动作不得空转”加入 `docs/ui-design/README.md` 的常规检查清单，避免类似按钮反馈问题在其它模块复现。

`change-intake` 的匹配优化可以延后到下一次 skill 维护；本次已经通过手动归属完成交付，没有阻塞实现。
