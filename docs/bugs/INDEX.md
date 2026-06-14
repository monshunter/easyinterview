# Bug 记录索引

> 本文件按模块组织所有 Bug 记录，便于快速检索和模式识别。

## Workspace

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0028](./BUG-0028.md) | targetjob review exposed error envelope and parse runtime drift | medium | resolved | 2026-05-08 | `fix(backend-targetjob): align targetjob review contracts` |
| [BUG-0027](./BUG-0027.md) | targetjob L2 review exposed runtime gate and SSRF gaps | medium | resolved | 2026-05-08 | `fix(backend-targetjob): harden targetjob L2 runtime gates` |
| [BUG-0026](./BUG-0026.md) | targetjob L2 review exposed import parse contract drift | medium | resolved | 2026-05-08 | `fix(backend-targetjob): remediate import parse L2 findings` |
| [BUG-0025](./BUG-0025.md) | analysis failed redline rejected documented provider secret code | medium | resolved | 2026-05-08 | `feat(backend-targetjob): complete import parse bootstrap handoff` |
| [BUG-0024](./BUG-0024.md) | targetjob detail omitted parsed summary provenance | medium | resolved | 2026-05-08 | `feat(backend-targetjob): complete import parse bootstrap handoff` |

## Practice

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0103](./BUG-0103.md) | empty focus competency arrays broke practice plan creation | high | resolved | 2026-05-24 | `feat(e2e-scenarios): close full funnel journey (BUG-0103)` |
| [BUG-0067](./BUG-0067.md) | practice voice turn contract was not mounted as HTTP route | high | resolved | 2026-05-17 | `fix(practice-voice): wire voice turn route and E2E gate (BUG-0067)` |
| [BUG-0060](./BUG-0060.md) | backend-practice hint AI drifted from lightweight observe prompt contract | high | resolved | 2026-05-15 | `fix(backend-practice): align hint prompt contract` |
| [BUG-0059](./BUG-0059.md) | appendSessionEvent hint replay returned stored errors and hint snapshots incorrectly | high | resolved | 2026-05-15 | `fix(backend-practice): preserve append event replay snapshots` |
| [BUG-0058](./BUG-0058.md) | backend-practice hint replay leaked hint text and BDD gates under-asserted evidence | high | resolved | 2026-05-15 | `fix(backend-practice): remediate hint provenance L2 findings` |
| [BUG-0057](./BUG-0057.md) | frontend practice event loop L2 review exposed recovery and parity gate drift | high | resolved | 2026-05-14 | `fix(frontend-practice): remediate practice event loop L2 findings` |
| [BUG-0056](./BUG-0056.md) | backend-practice event replay allowed terminal duplicates and duplicate AI reservations | high | resolved | 2026-05-14 | `fix(backend-practice): reserve session events before AI` |
| [BUG-0055](./BUG-0055.md) | backend-practice event loop trusted client follow-up state | high | resolved | 2026-05-14 | `fix(backend-practice): harden event loop replay contracts` |
| [BUG-0054](./BUG-0054.md) | backend-practice event loop L2 review exposed payload, status, and BDD gate drift | high | resolved | 2026-05-13 | `fix(backend-practice): remediate event loop L2 findings` |
| [BUG-0053](./BUG-0053.md) | appendSessionEvent accepted stale turn submissions | medium | resolved | 2026-05-13 | `feat(backend-practice): complete event loop and completion` |
| [BUG-0033](./BUG-0033.md) | practice session L2 reviews exposed first-question and idempotency drift | high | resolved | 2026-05-10 | `fix(backend-practice): remediate session orchestration L2 findings` |
| [BUG-0004](./BUG-0004.md) | voice interview surface was removed while unifying practice routes | medium | resolved | 2026-05-02 | `fix(ui-design): restore voice interview surface in practice shell` |

## Review

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0064](./BUG-0064.md) | report replay handoff reused source sessions and pixel gate was false-green | high | resolved | 2026-05-16 | `fix(frontend-report): harden replay handoff and pixel gate (BUG-0064)` |
| [BUG-0063](./BUG-0063.md) | frontend report dashboard L2 review exposed route and parity drift | high | resolved | 2026-05-16 | `fix(frontend-report): remediate report dashboard L2 findings (BUG-0063)` |
| [BUG-0062](./BUG-0062.md) | report L2 review exposed prompt schema, persistence, and privacy drift | high | resolved | 2026-05-16 | `fix(backend-review): remediate report L2 findings (BUG-0062)` |
| [BUG-0061](./BUG-0061.md) | report runtime only mounted read routes | high | resolved | 2026-05-16 | `fix(backend-review): wire report runner runtime (BUG-0061)` |
| [BUG-0005](./BUG-0005.md) | report follow-up CTAs returned to setup instead of starting sessions | medium | resolved | 2026-05-02 | `fix(ui-design): start report follow-up sessions directly` |

## Materials

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0123](./BUG-0123.md) | resume detail accepted rewrites missed flat profile and JD context | high | resolved | 2026-06-14 | `fix(resume): close detail rewrite review regressions (BUG-0123)` |
| [BUG-0081](./BUG-0081.md) | backend profile L2 review exposed privacy rollback and scenario evidence drift | high | resolved | 2026-05-21 | `fix(backend-profile): close profile L2 evidence drift (BUG-0081)` |
| [BUG-0076](./BUG-0076.md) | resume workshop real backend requests drifted from version contracts | high | resolved | 2026-05-18 | `fix(resume-workshop): close real backend gaps (BUG-0076, BUG-0077)` |

## Debrief

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0122](./BUG-0122.md) | debrief suggestions ignored sessionId in real backend context | medium | resolved | 2026-06-14 | `fix(debrief): wire session context for suggestions (BUG-0122)` |
| [BUG-0121](./BUG-0121.md) | debrief suggestions ignored resumeId in real backend context | medium | resolved | 2026-06-14 | `fix(debrief): wire resume context for suggestions (BUG-0121)` |
| [BUG-0078](./BUG-0078.md) | frontend debrief dev mock flow was stuck in analysis | high | resolved | 2026-05-18 | `fix(frontend-debrief): repair dev mock debrief flow (BUG-0078)` |
| [BUG-0071](./BUG-0071.md) | frontend debrief record modes drifted from ui-design parity | medium | resolved | 2026-05-17 | `fix(frontend-debrief): restore record mode ui parity (BUG-0071)` |
| [BUG-0069](./BUG-0069.md) | debrief real backend flows were blocked by mock-only contract drift | high | resolved | 2026-05-17 | `fix(frontend-debrief): repair real backend debrief flows (BUG-0068)` |
| [BUG-0068](./BUG-0068.md) | frontend debrief L2 review exposed route hydration and pixel gate drift | high | resolved | 2026-05-17 | `fix(frontend-debrief): close debrief L2 gaps (BUG-0067)` |
| [BUG-0065](./BUG-0065.md) | debrief.generate prompt baseline used retired output schema | high | resolved | 2026-05-16 | `feat(backend-debrief): close 001 debrief record and analysis baseline` |

## Eval

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0102](./BUG-0102.md) | Promptfoo eval output was written under config | medium | resolved | 2026-05-24 | `fix(prompt-rubric): keep promptfoo eval output in test output (BUG-0102)` |
| [BUG-0098](./BUG-0098.md) | prompt output schema review exposed jd_match posted drift and lint traceback | medium | resolved | 2026-05-24 | `fix(prompt-rubric): close output schema review findings (BUG-0098)` |
| [BUG-0097](./BUG-0097.md) | prompt registry seed migration missed jd_match feature keys | medium | resolved | 2026-05-24 | `fix(prompt-rubric): seed jd match prompt registry rows (BUG-0097)` |
| [BUG-0096](./BUG-0096.md) | prompt examples rendered as minimal placeholder JSON | medium | resolved | 2026-05-24 | `fix(prompt-rubric): render complete prompt example outputs (BUG-0096)` |
| [BUG-0095](./BUG-0095.md) | output schema validation accepted trailing model prose | medium | resolved | 2026-05-24 | `fix(prompt-rubric): reject trailing AI output schema content (BUG-0095)` |
| [BUG-0030](./BUG-0030.md) | prompt registry L2 review exposed provenance and no-op gate drift | medium | resolved | 2026-05-09 | `fix(prompt-rubric-registry): remediate provenance L2 findings` |
| [BUG-0006](./BUG-0006.md) | openai-compatible adapter assumed provider-specific model naming | medium | resolved | 2026-05-05 | `fix(historical-spec): deep reconcile existing plans` |

## Frontend

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0117](./BUG-0117.md) | auth verify recovery and skipped probe state regressed after unified login | medium | resolved | 2026-05-28 | `fix(frontend-shell): recover auth verify refresh failures (BUG-0117)` |
| [BUG-0116](./BUG-0116.md) | auth profile setup and scenario evidence drift escaped frontend-shell review | medium | resolved | 2026-05-28 | `fix(frontend-shell): close auth profile L2 gaps (BUG-0116)` |
| [BUG-0115](./BUG-0115.md) | unauthenticated home and interview routes mounted protected flows before login | high | resolved | 2026-05-28 | `fix(frontend-shell): gate protected routes before login (BUG-0115)` |
| [BUG-0112](./BUG-0112.md) | auth mail-link login stalled on empty 202 response and backend verify link | high | resolved | 2026-05-27 | `fix(auth): restore mail-link login flow (BUG-0112)` |
| [BUG-0101](./BUG-0101.md) | parse target switch kept stale preview | medium | resolved | 2026-05-24 | `fix(frontend-home): reset parse state on target switch (BUG-0101)` |
| [BUG-0100](./BUG-0100.md) | parse confirm handoff dropped workspace context fields | medium | resolved | 2026-05-24 | `fix(frontend-home): preserve parse confirm context (BUG-0100)` |
| [BUG-0099](./BUG-0099.md) | parse ready response skipped the ui-design loading demo | medium | resolved | 2026-05-24 | `fix(frontend-home): preserve parse loading demo (BUG-0099)` |
| [BUG-0089](./BUG-0089.md) | frontend owner plans missed real-backend handoff gates | medium | resolved | 2026-05-23 | `fix(frontend): close real backend handoff gate drift (BUG-0089)` |
| [BUG-0086](./BUG-0086.md) | home targetjob plan stayed fixture-only after real backend landed | medium | resolved | 2026-05-22 | `fix(frontend-home): close targetjob real backend gate drift (BUG-0086)` |
| [BUG-0085](./BUG-0085.md) | jd_match frontend plan stayed fixture-only after real backend landed | medium | resolved | 2026-05-22 | `fix(frontend-jd-match): close real backend gate drift (BUG-0085)` |
| [BUG-0079](./BUG-0079.md) | frontend shell popstate left unsafe URL markers visible | medium | resolved | 2026-05-18 | `fix(frontend-shell): scrub hostile popstate route privacy (BUG-0079)` |
| [BUG-0075](./BUG-0075.md) | resume branch rewrites L2 gates were false-green | medium | resolved | 2026-05-18 | `fix(frontend-resume-workshop): close branch rewrites L2 gates (BUG-0075)` |
| [BUG-0074](./BUG-0074.md) | resume create flow retry and L2 gates were false-green | medium | resolved | 2026-05-17 | `fix(frontend-resume-workshop): close create flow L2 gaps (BUG-0074)` |
| [BUG-0050](./BUG-0050.md) | resume workshop list drifted from ui-design tree controls | medium | resolved | 2026-05-13 | `fix(frontend-resume-workshop): restore ui-design list parity` |
| [BUG-0046](./BUG-0046.md) | resume workshop hid failed loads and original-source pending state | medium | resolved | 2026-05-12 | `fix(frontend-resume-workshop): harden resume workshop error states` |
| [BUG-0045](./BUG-0045.md) | resume workshop L2 review exposed original modal and CSS parity drift | medium | resolved | 2026-05-12 | `fix(frontend-resume-workshop): remediate listing detail L2 gaps` |
| [BUG-0041](./BUG-0041.md) | auth user menu browser parity missed mobile overflow | medium | resolved | 2026-05-11 | `fix(frontend-shell): close auth menu browser parity gap` |
| [BUG-0040](./BUG-0040.md) | workspace pixel gate depended on stale hydration and ignored baselines | medium | resolved | 2026-05-10 | `fix(frontend-shell): harden workspace pixel parity gate` |
| [BUG-0039](./BUG-0039.md) | frontend shell auth state and user menu parity drift | medium | resolved | 2026-05-10 | `fix(frontend-shell): restore auth state and user menu parity` |
| [BUG-0036](./BUG-0036.md) | Vite dev preview hit frontend port instead of fixture-backed API | high | resolved | 2026-05-10 | `fix(frontend-dev): default vite preview to fixture-backed API` |
| [BUG-0038](./BUG-0038.md) | jd_match search parity and pixel gate drift escaped L2 review | medium | resolved | 2026-05-10 | `fix(frontend-jd-match): restore search parity and clean pixel gate` |
| [BUG-0037](./BUG-0037.md) | jd_match L2 review exposed detail fetch and auth resume drift | medium | resolved | 2026-05-10 | `fix(frontend-jd-match): remediate jd match L2 findings` |
| [BUG-0032](./BUG-0032.md) | workspace follow-up review exposed synthetic id and fetch race drift | medium | resolved | 2026-05-09 | `fix(frontend-workspace): harden workspace review follow-up` |
| [BUG-0031](./BUG-0031.md) | workspace L2 review exposed route hydration and start flow drift | medium | resolved | 2026-05-09 | `fix(frontend-workspace): remediate workspace L2 findings` |
| [BUG-0029](./BUG-0029.md) | home JD import L2 review exposed privacy and gate drift | medium | resolved | 2026-05-08 | `fix(frontend-home): remediate jd import L2 findings` |
| [BUG-0021](./BUG-0021.md) | frontend shell TopBar drifted from ui-design source structure | medium | resolved | 2026-05-08 | `fix(frontend-shell): restore topbar ui-design source parity` |
| [BUG-0020](./BUG-0020.md) | frontend shell language switch was state-only | medium | resolved | 2026-05-07 | `fix(frontend-shell): restore app shell i18n` |
| [BUG-0019](./BUG-0019.md) | frontend shell review remediation missed build and auth edge gates | medium | resolved | 2026-05-07 | `fix(frontend-shell): harden app shell review remediation` |
| [BUG-0018](./BUG-0018.md) | frontend shell L2 review exposed route and auth wire drift | medium | resolved | 2026-05-07 | `fix(frontend-shell): remediate app shell L2 findings` |

## Platform

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0114](./BUG-0114.md) | auth registration treated display name as account identity boundary | high | resolved | 2026-05-27 | `fix(auth): enforce email identity semantics (BUG-0114)` |
| [BUG-0113](./BUG-0113.md) | scenario redeploy refreshed artifacts without restarting host-run services | high | resolved | 2026-05-27 | `fix(test): restart host-run services on redeploy (BUG-0113)` |
| [BUG-0107](./BUG-0107.md) | review follow-up exposed provider schema, report status, and privacy retry drift | medium | resolved | 2026-05-26 | `fix(backend): close review correctness follow-ups (BUG-0107)` |
| [BUG-0106](./BUG-0106.md) | privacy delete cleanup completed without removing UAT account identity | medium | resolved | 2026-05-26 | `fix(backend-async-runner): close privacy account identity cleanup (BUG-0106)` |
| [BUG-0091](./BUG-0091.md) | Postgres 18 dev volume layout made local dev stack unhealthy | medium | resolved | 2026-05-22 | `fix(local-dev-stack): repair Postgres 18 dev volume guard (BUG-0087)` |
| [BUG-0088](./BUG-0088.md) | async runner review exposed scheduler and report backoff drift | high | resolved | 2026-05-22 | `fix(backend-async-runner): harden scheduler and report backoff (BUG-0088)` |
| [BUG-0087](./BUG-0087.md) | async runner L2 review exposed outbox startup and live gate drift | high | resolved | 2026-05-22 | `fix(backend-async-runner): wire outbox startup and live gates (BUG-0085)` |
| [BUG-0080](./BUG-0080.md) | tool-name branch prefixes bypassed branch naming governance | medium | resolved | 2026-05-21 | `fix(governance): reject tool-name branch prefixes (BUG-0080)` |
| [BUG-0070](./BUG-0070.md) | async job polling and voice playback context were contract-only at runtime | high | resolved | 2026-05-17 | `fix(runtime): wire job polling and voice playback context (BUG-0070)` |
| [BUG-0049](./BUG-0049.md) | backend upload follow-up review exposed MinIO race and no-op live gate | medium | resolved | 2026-05-13 | `fix(backend-upload): close follow-up L2 gaps` |
| [BUG-0048](./BUG-0048.md) | backend upload L2 review exposed size privacy and live gate gaps | high | resolved | 2026-05-12 | `fix(backend-upload): harden upload privacy and live gates` |
| [BUG-0047](./BUG-0047.md) | upload presign runtime path was only partially wired | high | resolved | 2026-05-12 | `fix(backend-upload): wire presign runtime contracts` |
| [BUG-0035](./BUG-0035.md) | change-intake mutated main before branch guard | high | resolved | 2026-05-10 | `fix(change-intake): require branch guard before mutation` |
| [BUG-0022](./BUG-0022.md) | speech adapters ignored profile timeouts | medium | resolved | 2026-05-08 | `fix(ai-provider): enforce speech adapter timeouts` |
| [BUG-0017](./BUG-0017.md) | runtime topology lint missed structured producer and owner handoff forms | medium | resolved | 2026-05-07 | `fix(runtime): harden worker topology structured gate` |
| [BUG-0016](./BUG-0016.md) | runtime topology lint missed scripts and raw producer fields | medium | resolved | 2026-05-07 | `fix(runtime): harden worker topology scripts gate` |
| [BUG-0015](./BUG-0015.md) | runtime topology lint missed false-negative worker handoff forms | medium | resolved | 2026-05-06 | `fix(runtime): harden worker topology false-negative gate` |
| [BUG-0014](./BUG-0014.md) | backend runtime topology L2 review exposed worker drift gate gap | medium | resolved | 2026-05-06 | `fix(runtime): harden worker topology drift gate` |
| [BUG-0013](./BUG-0013.md) | backend auth L2 review exposed session contract drift | medium | resolved | 2026-05-06 | `fix(backend-auth): remediate passwordless L2 findings` |
| [BUG-0012](./BUG-0012.md) | AI client L2 review exposed tools, streaming, and observability drift | medium | resolved | 2026-05-06 | `fix(ai-provider): remediate tools streaming stt L2 findings` |
| [BUG-0009](./BUG-0009.md) | dev-stack profile catalog drift escaped lint gates | medium | resolved | 2026-05-05 | `fix(ai-provider): harden profile catalog drift gates` |
| [BUG-0008](./BUG-0008.md) | provider registry runtime bootstrap was only test-wired | medium | resolved | 2026-05-05 | `fix(ai-provider): wire provider registry runtime bootstrap` |
| [BUG-0007](./BUG-0007.md) | AI provider contract retained gateway terminology | medium | resolved | 2026-05-05 | `fix(ai-provider): remove gateway terminology from provider contract` |
| [BUG-0002](./BUG-0002.md) | secrets-config completed plan missed runtime binding drift | medium | resolved | 2026-04-30 | `fix(secrets-config): remediate L2 review findings` |

## Schema

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0119](./BUG-0119.md) | resume flatten review left runtime and data retention regressions | high | resolved | 2026-06-14 | `fix(schema): close resume flatten review regressions (BUG-0119)` |
| [BUG-0084](./BUG-0084.md) | jd-match L2 follow-up missed privacy runner and agent-scan context wiring | high | resolved | 2026-05-22 | `fix(backend-jobs): close jd-match L2 follow-up (BUG-0084)` |
| [BUG-0083](./BUG-0083.md) | jd-match runtime bypassed A3/F3 and missed search/privacy contracts | high | resolved | 2026-05-22 | `fix(backend-jobs): harden jd-match runtime contracts (BUG-0083)` |
| [BUG-0077](./BUG-0077.md) | OpenAPI diff baseline lag hid the current additive contract change | medium | resolved | 2026-05-18 | `fix(resume-workshop): close real backend gaps (BUG-0076, BUG-0077)` |
| [BUG-0073](./BUG-0073.md) | resume tailor ready provenance was incomplete after DB roundtrip | medium | resolved | 2026-05-17 | `fix(backend-resume): persist tailor run provenance (BUG-0073)` |
| [BUG-0072](./BUG-0072.md) | practice voice fixture audio refs used mock-only scheme | medium | resolved | 2026-05-17 | `fix(practice-voice): align fixture audio refs (BUG-0072)` |
| [BUG-0052](./BUG-0052.md) | backend resume L2 review exposed validation and retry state drift | high | resolved | 2026-05-13 | `fix(backend-resume): remediate asset registration L2 findings` |
| [BUG-0051](./BUG-0051.md) | getResume not-found fixture used undocumented error code | medium | resolved | 2026-05-13 | `feat(backend-resume): close resume baseline verification` |
| [BUG-0044](./BUG-0044.md) | resume additive generated client and cleanup drift escaped contract gates | medium | resolved | 2026-05-12 | `fix(openapi): harden resume additive client contracts` |
| [BUG-0043](./BUG-0043.md) | resume fileless intake still required upload file object | high | resolved | 2026-05-12 | `fix(openapi): allow fileless resume intake contracts` |
| [BUG-0042](./BUG-0042.md) | resume tailor mode enum drifted across event consumers | medium | resolved | 2026-05-12 | `fix(events): align resume tailor mode contract` |
| [BUG-0023](./BUG-0023.md) | jobs_test referenced removed embedding_upsert constant after capability cleanup | medium | resolved | 2026-05-08 | `fix(event-outbox): drop stale embedding job reference in jobs_test` |
| [BUG-0001](./BUG-0001.md) | OpenAPI breaking-change gate missed composition diffs | medium | resolved | 2026-04-29 | `fix(openapi): tighten breaking-change gate composition diff` |

## Test

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0124](./BUG-0124.md) | resume Phase 8 flat gate drifted on current profile fields | medium | resolved | 2026-06-14 | `fix(test): close resume phase8 flat gate drift (BUG-0124)` |
| [BUG-0120](./BUG-0120.md) | D-20 resume gates missed retired fixture keys and stale scenario scripts | medium | resolved | 2026-06-14 | `fix(test): close resume d20 gate drift (BUG-0120)` |
| [BUG-0118](./BUG-0118.md) | D-17/D-18/D-20 refactor left stale scenario and contract gates | medium | resolved | 2026-06-14 | `fix(test): close ux funnel refactor drift (BUG-0118)` |
| [BUG-0111](./BUG-0111.md) | scenario env review follow-ups exposed stale env and evidence gates | medium | resolved | 2026-05-27 | `fix(test): close scenario env review follow-ups (BUG-0111)` |
| [BUG-0110](./BUG-0110.md) | real-provider UAT bypassed the standard scenario runner | medium | resolved | 2026-05-27 | `fix(test): use dev-stack env for hybrid scenario (BUG-0110)` |
| [BUG-0109](./BUG-0109.md) | scenario environment lifecycle was coupled to individual scenario runners | medium | resolved | 2026-05-27 | `fix(test): decouple scenario env lifecycle (BUG-0109)` |
| [BUG-0108](./BUG-0108.md) | P0.050 task-run gate lagged answer summary observation | medium | resolved | 2026-05-26 | `fix(backend-practice): align p0050 task-run gate (BUG-0108)` |
| [BUG-0105](./BUG-0105.md) | manual full-funnel real-provider gate missed runtime-only blockers | high | resolved | 2026-05-26 | `fix(manual-uat): close real provider full funnel blockers (BUG-0105)` |
| [BUG-0104](./BUG-0104.md) | manual UAT account helper crossed into backend cmd before Mailpit boundary | medium | resolved | 2026-05-26 | `fix(local-dev-stack): add mailpit for local auth testing (BUG-0104)` |
| [BUG-0094](./BUG-0094.md) | gitleaks lint scanned ignored local env secrets | medium | resolved | 2026-05-23 | `feat(prompt-rubric): close output schema contract (BUG-0094)` |
| [BUG-0093](./BUG-0093.md) | Ready e2e scenarios were missing required data assets | medium | resolved | 2026-05-23 | `fix(test): add missing scenario data assets (BUG-0093)` |
| [BUG-0090](./BUG-0090.md) | frontend owner scenario wrappers widened scoped gates and ignored hash routes | medium | resolved | 2026-05-23 | `fix(test): close frontend owner scenario full-run gates (BUG-0090)` |
| [BUG-0092](./BUG-0092.md) | repo lint gates rejected current practice voice contracts and drifted follow-up evidence | medium | resolved | 2026-05-22 | `fix(repo-lint): align current contract gates (BUG-0088)` |
| [BUG-0082](./BUG-0082.md) | jd-match BDD closure treated smoke wrappers as completed gates | high | resolved | 2026-05-22 | `fix(backend-jobs): harden jd-match runtime contracts (BUG-0082, BUG-0083)` |
| [BUG-0066](./BUG-0066.md) | debrief scenario wrappers were missing or under-asserted | high | resolved | 2026-05-17 | `fix(test): close debrief scenario wrapper evidence (BUG-0066)` |
| [BUG-0034](./BUG-0034.md) | mockruntime named scenario test copied stale fixture expectation | medium | resolved | 2026-05-10 | `fix(mock-contract): align mockruntime named scenarios` |
| [BUG-0011](./BUG-0011.md) | mock contract gate ignored empty retired fixture tag directories | medium | resolved | 2026-05-06 | `fix(mock-contract): reject retired fixture tag dirs` |
| [BUG-0010](./BUG-0010.md) | mock contract runtime gate missed registry and stale route count | medium | resolved | 2026-05-05 | `fix(mock-contract): harden runtime drift gates` |
| [BUG-0003](./BUG-0003.md) | local quality gates skipped real backend and frontend execution | medium | resolved | 2026-04-30 | `fix(ci-pipeline): remediate local quality gates` |
