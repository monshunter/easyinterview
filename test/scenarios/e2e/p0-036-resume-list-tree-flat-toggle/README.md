# E2E.P0.036 — Resume Workshop list view: tree/flat toggle + StatsStrip + auth boundary

> Owner: [`frontend-resume-workshop/001-listing-routing-and-detail-readonly`](../../../docs/spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/plan.md)
> 关联需求: frontend-resume-workshop C-1, C-2, C-3, C-5, C-6, C-7, C-8, C-9
> 关联 BDD: [bdd-plan.md](../../../docs/spec/frontend-resume-workshop/plans/001-listing-routing-and-detail-readonly/bdd-plan.md) E2E.P0.036
> 状态: Ready · 执行方式: automated

## 1 范围

验证 Resume Workshop 列表视图的核心 P0 路径：

- 路由替换：`resume_versions` 走 `ResumeWorkshopScreen`，旧 `PlaceholderScreen` 不再出现。
- 未登录态展示 auth gate，0 触发 `listResumes` / `listResumeVersions` / `getResumeVersion` / `exportResumeVersion`。
- 登录态默认渲染 StatsStrip 4 项 + ViewSwitcher（tree active）+ 按 `resumeAssetId` 分组的 tree row；数量从 fixture body 派生，第二个无匹配 version 的 asset 显示 no-versions 占位。
- 「选为底稿」/「基于这棵树新建版本」按钮 click 触发 `eiToast` "即将开放"。
- ViewSwitcher 切到 flat 渲染 `ResumeFlatView` 行；fixture mock transport 不按 path 参数选 version scenario。
- DOM 中不出现 retired-route testid（`route-welcome` / `route-mistakes` / `route-drill` / `route-followup` / `route-onboarding` / `route-experiences` / `route-star` / `route-voice`）。

## 2 触发流

- `setup.sh` 准备 `.test-output/e2e/p0-036-...`
- `trigger.sh` 在仓库根执行 `pnpm --filter @easyinterview/frontend test src/app/scenarios/p0-036-resume-list-tree-flat-toggle.test.tsx`
- `verify.sh` 校验输出含 4 tests passed + 1 test file passed，并对 retired testid 字面量做 negative grep。
- `cleanup.sh` 清理 setup 标记。

## 3 离线限制

vitest scenario test 全部本地执行，使用 in-tree fixtures，不依赖 Docker / Kind / 网络。
