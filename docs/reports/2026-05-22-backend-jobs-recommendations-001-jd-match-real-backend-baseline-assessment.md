# Backend Jobs Recommendations 001 JD-Match L2 Assessment And Fix

> **日期**: 2026-05-22
> **审查人**: Codex
> **状态**: Fixed, Go
> **关联 Bug**: [BUG-0082](../bugs/BUG-0082.md), [BUG-0083](../bugs/BUG-0083.md)

## 1 结论

`/plan-code-review 001-jd-match-real-backend-baseline --fix` 初始复查结论为 **No-Go**，原因是 Phase 5.5-5.8 与 BDD-Gate 6.5-6.8 的证据不足，且主 checklist 的 PASS 注释与 BDD checklist / scenario INDEX / 实际测试体矛盾。

本次已完成同一 owner plan 内的修复与复验，当前结论为 **Go**：真实 A3/F3 adapter、search completed outbox、output_invalid、transactional privacy delete、BDD wrapper fail-marker hardening 和 post-reopen 文档证据均已闭环。

## 2 L2 Findings

| ID | 严重性 | 证据 | 结论 |
|----|--------|------|------|
| D-L2-001 | Blocking | `backend/cmd/api/jdmatch_live_scenario_test.go` 注释声明 `TestJDMatchHTTPScenario` 是 smoke，且不是完整 12-route + IK replay + drainer scenario；缺少 `DATABASE_URL` 时会 skip。 | 不能作为 5.6 / 6.1 / 6.5-6.8 的完整 BDD 证据。 |
| D-L2-002 | Blocking | `go test -list` 只列出 `TestJDMatchRoutesRegistered` / `TestBuildJDMatchRoutesReference` / `TestJDMatchUnauthorisedReturns401` / `TestJDMatchHTTPScenario`，没有 `TestBuildJDMatchRuntime` 或 `TestJDMatchAgentScanDrainerScenario`。 | 5.5 / 5.7 / 6.1 引用的 drainer/runtime gate 不存在。 |
| D-L2-003 | Blocking | `bdd-checklist.md` Header 原为 completed，但 E2E.P0.094-097 全部正文 checkbox 仍为 `[ ]`。 | 主 checklist 的 `bddChecklist=complete` 注释为 false-green。 |
| D-L2-004 | Blocking | `test/scenarios/e2e/INDEX.md` 中 E2E.P0.094-097 仍为 `scaffold` / `Pending`。 | 场景资产与执行状态不能支撑 Ready / automated / PASS。 |
| D-L2-005 | Blocking | 4 个 scenario `trigger.sh` 都只运行同一个 `TestJDMatchHTTPScenario`，`verify.sh` 只检查 `--- PASS` 与 raw email，未验证 package `ok`、no-op/skip、场景专属 DB/outbox/IK/fixture 断言。 | wrapper 证据不足，不能作为 BDD gate。 |

## 3 修复动作

- `plan.md` / `checklist.md` / `bdd-plan.md` / `bdd-checklist.md` 从 completed v1.1 退回 active v1.2。
- `checklist.md` 重新打开 5.5-5.8、6.1-6.2、6.5-6.8、6.12-6.13，并逐项写入 L2 reopened 证据。
- `context.yaml` `specVersion.to` 从 1.1 对齐到当前 spec v1.2。
- `plans/INDEX.md` 把 plan 001 从 Completed 移回 Active。
- `history.md` 追加 1.4 L2 reopen 行，保留 1.3 close-out attempt 的历史并由 1.4 明确 supersede。
- 后续 `/implement` 与本轮 `/plan-code-review --fix` 完成 runtime hardening：`main.go` 使用 `buildJDMatchAIAdapters(loader, db, targetJobRuntime.AI.Client)`，不再以 stub AI 启动生产 JD-Match；F3 prompt 要求 `jobMatchId` 并刷新 template_hash；search 成功发射 `jd_match.search.completed`，output_invalid 返回 `AI_OUTPUT_INVALID`；privacy delete 进入 cmd/api transaction rollback gate；4 个 scenario verify 脚本拒绝 `--- FAIL` / package `FAIL` 并要求 package-level `ok`。
- `history.md` 追加 1.5 completion 与 1.6 L2 hardening 行，plan / checklist / bdd-plan / bdd-checklist 保持 completed v1.2，plans/INDEX row 位于 Completed。

## 4 当前可接受证据

- `cd backend && go test ./...`：PASS。
- `cd backend && go test ./internal/jdmatch/...`：PASS。
- `cd backend && go test ./internal/ai/registry -count=1`：PASS（含 `TestBackendJDMatchF3Preflight` 与 prompt hash load）。
- `cd backend && DATABASE_URL='postgres://easyinterview:dev@localhost:5432/easyinterview?sslmode=disable' go test ./cmd/api -run '^(TestBuildJDMatchRuntimeWiresRoutesDrainerAndLifecycle|TestJDMatchRoutesRequireSessionOnAllRoutes|TestJDMatchHTTPScenario|TestJDMatchAgentScanDrainerScenario|TestJDMatchFixtureParity|TestJDMatchA3F3AdapterUsesRegistryProfilesForSearchAndRecommendation|TestParseJDMatchSearchIDsRejectsMissingJobMatchID|TestDeleteJDMatchDataForUserInTxCommitsOrderedDeletesAndAudit|TestDeleteJDMatchDataForUserInTxRollsBackOnDeleteFailure)$' -count=1 -v`：PASS，非 skip。
- E2E.P0.094 / P0.095 / P0.096 / P0.097 `setup.sh -> trigger.sh -> verify.sh`：全部 PASS。
- `rg -n 'LinkedIn|Boss|脉脉|拉勾|mistake|growth|drill|experiences|star\b' backend/internal/jdmatch`：无命中。

## 5 下一步

推荐下一步 owner skill：`/plan-code-review backend-jobs-recommendations/001-jd-match-real-backend-baseline`（不带 `--fix`），目标范围是最终只读复核本轮 L2 hardening 之后的代码、docs、scenario logs、bug/report/retrospective 收尾状态；若无新 finding，即可进入 `/work-journal` 提交。
