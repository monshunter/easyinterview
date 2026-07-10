# E2E.P0.037 — Resume Workshop detail: read-only resume body + 404 fallback

> Owner: [`frontend-resume-workshop/001-listing-routing-and-detail-readonly`](../../../../docs/spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)
> 关联需求: frontend-resume-workshop C-3, C-10, C-11
> 关联 BDD: [bdd-plan.md](../../../../docs/spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/bdd-plan.md) E2E.P0.037
> 状态: Ready · 执行方式: automated

## 1 范围

验证 Resume Workshop 详情页面 P0 路径：

- Flat resume detail 只渲染简历正文；不渲染 Preview / Rewrites / Edit tablist。
- Out-of-scope `?tab=rewrites` 参数被忽略，不能激活 `ResumeRewritesTab` 或任何 edit/rewrite surface。
- 详情页不暴露 Export PDF / Copy text / View original / original modal 等二次操作。
- 不存在的 resumeId 返回 404 → `NotFoundEmptyState` 渲染通用文案与返回列表 CTA，不直接回显 fixture `error.code`（如 `RESOURCE_NOT_FOUND`）。

## 2 触发流

- `setup.sh` 准备 `.test-output/e2e/p0-037-...`
- `trigger.sh` 在仓库根执行 `pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-037-resume-detail-preview-readonly.test.tsx`
- `verify.sh` 校验 6 tests passed + pending PDF upload 轮询到 source page stack + failed-with-snapshot PDF upload 单次请求 + out-of-scope testid 字面量 negative + 不出现 `TARGET_JOB_NOT_FOUND` 字符串。
- `cleanup.sh` 清理 setup 标记。

## 3 离线限制

vitest scenario test 全部本地执行，使用 in-tree fixtures，不依赖 Docker / Kind / 网络。
