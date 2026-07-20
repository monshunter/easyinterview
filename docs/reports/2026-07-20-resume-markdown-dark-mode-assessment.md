# Resume Markdown Dark Mode 交付复盘报告

> **日期**: 2026-07-20
> **审查人**: Codex

## 1 复盘范围与成功证据

- 本次交付修复 [BUG-0195](../bugs/BUG-0195.md)：Markdown 简历继续使用 794px 白色连续纸张，但正文不再继承夜间应用壳的浅色 foreground。
- 原 owner 已在 [frontend-resume-workshop 001 Phase 28](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md) 原地修订；spec、UI design、plan/checklist 与 BDD 均增加 light/dark 可读性合同，没有建立平行 plan。
- TDD 证据：新增 CSS parity gate 先以 1/9 失败复现缺少局部 `color`，最小 CSS 修复后 9/9 通过。
- 回归证据：Resume Workshop 20 文件 / 125 tests、frontend typecheck/build、根 `make test` 621 tests / 4628 subtests 与所有 Go packages 均通过。
- 运行证据：frontend 定向重部署、共享依赖 4/4 健康；真实 Chrome light/dark 下 page/body/list/listitem/strong/code 均计算为 `rgb(34, 34, 34)`，白底为 `rgb(255, 255, 255)`，A4 宽度 794px、单页连续高度、无横向溢出、console warning/error 为 0。

## 2 会话中的主要阻点/痛点

- 截图中的“透明化”容易被理解为 opacity 或过渡动画问题。
  - **证据**：修复前真实 DOM 的 `opacity` 全部为 1；真正异常是正文计算色 `rgb(232, 237, 246)` 落在纯白纸张上。
  - **影响**：若只做视觉猜测，可能错误调整 opacity/animation，无法修复颜色继承。
- 既有 A4 回归只保护纸张几何，没有保护固定白色 surface 的前景色边界。
  - **证据**：原 CSS parity 检查宽度、比例和高度约束；新增 ink 断言在当前实现上唯一 RED。
  - **影响**：应用主题能力增强或 surface 从主题卡片改为固定白纸后，正文对比度可以静默回退。
- 主题切换截图存在短暂过渡帧，单张即时截图不能代替计算样式证据。
  - **证据**：首次夜间截图捕获过渡中间帧；等待过渡稳定后，最终截图和计算样式一致。
  - **影响**：如果只依据瞬时像素，容易把已修复页面误判为仍存在裁切或遮罩问题。

## 3 根因归类

- 固定白色文档 surface 缺少局部 foreground owner。
  - **类别**：spec-plan。
  - UI 合同锁定了白纸，却未明确白纸在 light/dark 两种应用模式下必须拥有独立墨水色；本次已在原 spec/plan 中补齐。
- CSS parity coverage matrix 缺少 theme alternate path。
  - **类别**：spec-plan。
  - A4 Phase 27 覆盖主路径和 responsive 边界，但未覆盖 dark-mode alternate；Phase 28 已用 source contrast gate 与真实 Chrome 补齐。
- 过渡帧截图属于一次性验收时序问题。
  - **类别**：无需仓库改动。
  - 当前 Chrome 验收已用稳定后截图和 computed style 双证据收口，无需新增通用 sleep 或截图框架。

## 4 对流程资产的改进建议

- 后续新增固定浅色/深色内容 surface 时，在 owner spec/plan 的 alternate-path coverage 中同时列出背景、正文、链接、rule 的主题隔离要求。
  - **落点**：对应 UI spec / plan。
  - **优先级**：high。
- CSS source gate 应验证 surface 自己声明的 foreground/background 组合及最低对比度，不要只检查根 token 或单个标题。
  - **落点**：对应 frontend owner test。
  - **优先级**：high。
- 浏览器验收主题切换时优先记录稳定后的 computed style，再用截图补充视觉结论；不把过渡帧截图单独作为 PASS/FAIL。
  - **落点**：无需仓库改动；沿用当前 Chrome 验收方法。
  - **优先级**：medium。

## 5 建议优先级与后续动作

- 本轮最高价值动作已经完成：Phase 28 将 dark-mode alternate path 和对比度 gate 固化到当前 Resume owner。
- 下一步建议在合并前执行 work-journal、docs/index/diff gate 并提交 `fix(resume): keep Markdown paper readable in dark mode (BUG-0195)`。
- 可延后处理：经用户确认后，把“固定 surface 同时 owning foreground/background”提炼到 `docs/bugs/PATTERNS.md`，供其他白色文档面复用。
