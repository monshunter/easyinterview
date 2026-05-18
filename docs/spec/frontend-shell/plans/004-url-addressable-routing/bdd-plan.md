# URL-Addressable Routing BDD Plan

> **版本**: 1.2
> **状态**: completed
> **更新日期**: 2026-05-18

## 1 Scenario Map

| 场景 ID | 场景 | 覆盖 Phase | 对应验收 | Checklist Gate |
|---------|------|------------|----------|----------------|
| E2E.P0.088 | canonical path deep-link / reload / back-forward | Phase 2 + Phase 4 | C-11 | Phase 4.3 |
| E2E.P0.089 | auth pendingAction + URL privacy redline | Phase 3 | C-12 | Phase 3.3 |
| E2E.P0.090 | hash compatibility + legacy route negative regression | Phase 1 + Phase 4 | C-13 | Phase 4.4 |

## 2 Scenario Details

| 场景 ID | 场景 | Given | When | Then | 验证入口 |
|---------|------|-------|------|------|----------|
| E2E.P0.088 | Canonical path deep-link / reload / browser history | Frontend app built with Browser History router; fixture-backed mock runtime available; safe server-bound IDs exist for workspace, practice, generating, report, resume workshop and debrief paths | Open `/workspace?targetJobId=...&resumeVersionId=...&planId=...&autoStartPractice=1`, reload, navigate to `/practice?mode=voice&modality=voice&sessionId=...`, open `/generating?reportId=...&sessionId=...` and `/report?sessionId=...&reportId=...&reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT`, open `/resume-versions?versionId=...&tab=rewrites&tailorRunId=...`, open `/debrief?targetJobId=...&debriefId=...&debriefJobId=...`, then go back/forward | URL parses to the expected `Route` + params; TopBar active route and chrome hidden state match route catalog; InterviewContext hydrates from safe params; reload preserves resource context; back/forward does not double-push or lose params; current cross-owner handoff keys survive allowlist filtering; unknown/malformed query falls back without crashing | `test/scenarios/e2e/p0-088-url-addressable-routing-canonical/` |
| E2E.P0.089 | Auth pendingAction + URL privacy redline | User is unauthenticated; app opens URL-addressable workflow paths with safe IDs; test harness also seeds representative raw JD text, source URL, jd_match search query/label, resume text, guided answers, parsed summary, structured profile, suggestion text, question/answer text, debrief notes and AI prompt markers | User triggers auth-gated actions for workspace auto-start, report replay, home import, jd_match Recommended/Search pending action, resume workshop and debrief handoff, completes mock passwordless login, returns to the original route, and restores a hostile browser history entry carrying raw query / hash / `history.state` markers through `popstate` | Restored route and canonical URL contain only route name + safe IDs/hints; existing legal handoff params such as `autoStartPractice`, `practiceGoal`, `reportId`, `tailorRunId`, `pendingImportId`, `selectedJobMatchId`, `pendingJdMatchActionId`, `debriefId` and `debriefJobId` survive; jd_match query/label and raw payload markers have zero hits in URL/history/pendingAction/storage/logs; auth token/secret is not persisted in query, pendingAction, storage or history; login restore keeps the same target context; hostile popstate entries are immediately replaced with canonical safe URL and null `history.state` | `test/scenarios/e2e/p0-089-url-routing-auth-privacy/` |
| E2E.P0.090 | Hash compatibility + legacy route negative regression | Existing static preview / pixel parity inputs still use `#route=...`; old route aliases and standalone `voice` inputs are available as direct URLs / hash routes | Open representative hash URLs (`#route=home`, `#route=workspace&targetJobId=...`, `#route=practice&mode=voice&modality=voice`), open legacy/unknown paths and `#route=voice`, run retired-route grep | Hash routes still bootstrap through `normalizeRoute` and land on equivalent canonical path/route state; `voice` never materializes as standalone route; retired aliases normalize to retained routes or `home`; canonical output never emits retired paths; server fallback returns `index.html` for frontend paths and not for `/api/*` | `test/scenarios/e2e/p0-090-url-routing-hash-legacy-negative/` |

## 3 Regression References

| 场景 ID | 场景 | 复用目的 | 验证入口 |
|---------|------|----------|----------|
| E2E.P0.001 | 默认首页与五入口 Shell | 证明 canonical routing 未破坏默认 route、TopBar 五入口和旧入口负向约束 | `test/scenarios/e2e/p0-001-default-home-shell/` |
| E2E.P0.002 | 登录打断后恢复原业务动作 | 证明新的 pendingAction canonical restore 与 D1 auth gate 兼容 | `test/scenarios/e2e/p0-002-auth-pending-action-resume/` |
| E2E.P0.006 | 真实浏览器 ui-design pixel parity gate | 证明保留 `#route=` adapter 不破坏 pixel parity harness | `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/` |
| E2E.P0.087 | Resume Workshop latest active E2E | 作为编号连续性参照；本计划新增场景从 E2E.P0.088 开始 | `test/scenarios/e2e/p0-087-resume-detail-export-copy-consistency-and-parity/` |
