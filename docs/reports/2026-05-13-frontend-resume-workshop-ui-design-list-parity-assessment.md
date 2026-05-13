# Frontend Resume Workshop UI-Design List Parity 交付复盘报告

> **日期**: 2026-05-13
> **审查人**: Codex

**关联计划**: [frontend-resume-workshop/001-listing-routing-and-detail-readonly](../spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)

## 1 复盘范围与成功证据

- **范围**：修复正式前端简历工坊列表页与 `ui-design/src/screen-resume-workshop.jsx` 当前页面在导航文案、header 主按钮、统计卡副文案、树形底稿操作、版本行结构、旧 toast locale key 和 BDD gate 上的差异。
- **Bug 记录**：[BUG-0050](../bugs/BUG-0050.md) 已建档。
- **成功证据**：
  - Focused red/green：`pnpm --filter @easyinterview/frontend test src/app/topbar/TopBar.test.tsx src/app/screens/resume-workshop/components/ResumeListView.test.tsx src/app/screens/resume-workshop/components/ResumeTreeView.test.tsx src/app/screens/resume-workshop/components/ResumeVersionRow.test.tsx` → 4 files / 19 tests passed。
  - Locale drift gate：`pnpm --filter @easyinterview/frontend test src/app/i18n/localeFiles.test.ts` → 1 file / 3 tests passed。
  - Resume workshop regression：`pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/ResumeWorkshopCssParity.test.ts src/app/scenarios/p0-036-resume-list-tree-flat-toggle.test.tsx src/app/screens/resume-workshop/ResumeWorkshopI18nA11y.test.tsx src/app/screens/resume-workshop/components/ResumeFlatView.test.tsx src/app/screens/resume-workshop/components/ResumeListAggregation.test.tsx` → 5 files / 21 tests passed。
  - Full frontend suite：`pnpm --filter @easyinterview/frontend test` → 127 files / 782 tests passed。
  - Build：`pnpm --filter @easyinterview/frontend build` → passed。
  - Pixel parity：`pnpm --filter @easyinterview/frontend test:pixel-parity` → 122 passed。
  - Whitespace gate：`git diff --check` → passed。

## 2 会话中的主要阻点/痛点

### 2.1 历史 parity gate 覆盖了可用性，但没有覆盖按钮层级与行结构

- **证据**：旧 P0.036 与组件测试能证明列表加载、树/平铺切换和打开详情，但没有断言 "新建简历" 是 header 主按钮，也没有断言 "选为底稿" 后才出现统一分叉入口。
- **影响**：正式前端保留了每个底稿卡片头部的旧 "基于这棵树新建版本" 按钮，用户对照 `ui-design` 时仍能看到功能按钮距离。

### 2.2 StatsStrip 和版本行缺少源级细节断言

- **证据**：旧测试没有覆盖 `2 / 3`、`3 个定制`、最高匹配来源、最近编辑版本说明，也没有覆盖 version row 的缩进、icon、tag、match badge 和 `打开 ->`。
- **影响**：页面数值能显示，但信息密度、可扫描性和原型视觉层级偏离明显。

### 2.3 TopBar 文案漂移没有被简历模块测试捕获

- **证据**：正式前端中文导航仍是 "简历版本"，而原型截图和源码已切到 "简历"；本次需要在 `TopBar.test.tsx` 加回归断言。
- **影响**：模块入口层面的命名漂移会让用户从第一屏就感知与原型不一致。

### 2.4 旧交互文案保留在 locale catalog 中

- **证据**：正式代码已经删除每行固定 "基于这棵树新建版本" toast 交互，但 `resumeWorkshop.tree.toastSelect` / `resumeWorkshop.tree.toastBranch` 仍保留在中英 locale 文件里。
- **影响**：旧口径虽然不再被组件引用，但仍可能被后续实现误用；本次用 `localeFiles.test.ts` 加负向断言后删除。

## 3 根因归类

| # | 根因 | 类别 |
|---|------|------|
| 1 | P0.036 / 组件测试只覆盖流程可用，没有覆盖当前 `ui-design` 的功能按钮层级和行结构 | spec-plan |
| 2 | `frontend-resume-workshop/001` 历史完成证据没有随着原型文案和按钮结构变化重新反查 | spec-plan |
| 3 | shell TopBar 文案与模块页面 parity 分属不同测试面，缺少跨层入口断言 | spec-plan |
| 4 | runtime locale catalog 未被纳入旧 UI 口径负向搜索，未使用 key 继续保留 | spec-plan |

## 4 对流程资产的改进建议

- **建议 A**：后续 UI parity plan/checklist 中，把 "功能按钮所属层级、选中后状态、行内 icon/tag/badge" 作为必检 DOM anchor，而不只检查页面截图和主流程。
  - **落点**：spec-plan
  - **优先级**：high
- **建议 B**：对导航入口文案与页面模块标题存在命名耦合的页面，在模块 focused tests 中引用至少一个 TopBar 文案断言或 shell contract gate。
  - **落点**：spec-plan
  - **优先级**：medium
- **建议 C**：UI parity L2 负向搜索应覆盖 i18n locale catalog 中的旧交互 key，防止未使用文案成为后续实现回流入口。
  - **落点**：spec-plan
  - **优先级**：medium
- **建议 D**：`/plan-code-review` 对用户截图反馈的 UI parity 修复应默认读取当前 `ui-design` 源码，并反向生成 red tests 覆盖可见差异，再开始样式修复。
  - **落点**：skill
  - **优先级**：medium

## 5 建议优先级与后续动作

- **最高优先级**：把建议 A 纳入后续 `frontend-resume-workshop` UI plan 的开题 gate，尤其是 create / branch / detail 编辑页，避免只做视觉 smoke。
- **下一步建议**：如果继续推进简历工坊，优先走 `/plan-code-review frontend-resume-workshop/001-listing-routing-and-detail-readonly frontend --fix` 做一次最终 L2 owner close-out；若用户要让 "新建简历" 和 "基于底稿分叉" 从占位导航进入真实流程，则应切到下一份 create / branch flow plan 后再 `/implement`。
