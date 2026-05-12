# E2E.P0.037 — Resume Workshop detail: Preview tab + original modal + 404 fallback + export 501

> Owner: [`frontend-resume-workshop/001-listing-routing-and-detail-readonly`](../../../docs/spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)
> 关联需求: frontend-resume-workshop C-4, C-5, C-6, C-7, C-8
> 关联 BDD: [bdd-plan.md](../../../docs/spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/bdd-plan.md) E2E.P0.037
> 状态: Ready · 执行方式: automated

## 1 范围

验证 Resume Workshop 详情页面 P0 路径：

- MASTER 版本默认 active tab=preview，TARGETED 版本带 `?tab=rewrites` 时不被改写为 preview。
- Preview Tab 渲染 `buildResumePreview` 投影并暴露 Export PDF / Copy text / View original 三个按钮。
- View original 弹出 `role=dialog` + `aria-modal=true` modal，焦点落在关闭按钮，ESC 关闭。
- Export PDF 调用 `exportResumeVersion(versionId, { idempotencyKey })`，请求 header 携带 `Idempotency-Key` 且匹配 `v1.<unix>.<uuid>`，触发 P0 "PDF 导出能力即将开放" toast，不写 localStorage。
- 不存在的 versionId 返回 404 → `NotFoundEmptyState` 渲染通用文案与返回列表 CTA，不直接回显 fixture `error.code`（如 `TARGET_JOB_NOT_FOUND`）。

## 2 触发流

- `setup.sh` 准备 `.test-output/e2e/p0-037-...`
- `trigger.sh` 在仓库根执行 `pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx`
- `verify.sh` 校验 5 tests passed + retired testid 字面量 negative + 不出现 `TARGET_JOB_NOT_FOUND` 字符串。
- `cleanup.sh` 清理 setup 标记。

## 3 离线限制

vitest scenario test 全部本地执行，使用 in-tree fixtures，不依赖 Docker / Kind / 网络。
