# Bug 记录索引

> 本文件按模块组织所有 Bug 记录，便于快速检索和模式识别。

## Workspace

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0151](./BUG-0151.md) | archived TargetJob import jobs retried failure cleanup | medium | resolved | 2026-07-09 | `fix(targetjob): terminate archived target imports (BUG-0151)` |
| [BUG-0150](./BUG-0150.md) | archiveTargetJob handler was generated but not mounted in API mux | high | resolved | 2026-07-09 | `fix(targetjob): persist workspace archive delete (BUG-0150)` |
| [BUG-0149](./BUG-0149.md) | hidden signals stayed blank when JD parse omitted hidden_signal classification | medium | resolved | 2026-07-09 | `fix(frontend-home): derive hidden signals from JD parse (BUG-0149)` |
| [BUG-0147](./BUG-0147.md) | failed JD parse entered interview list and workspace kept route context | high | resolved | 2026-07-09 | `fix(workspace): enforce parse failure admission and route purity (BUG-0147)` |
| [BUG-0146](./BUG-0146.md) | target import parse rejected valid JD without company name | high | resolved | 2026-07-09 | `fix(targetjob): accept valid JD parses without company names (BUG-0146)` |
| [BUG-0144](./BUG-0144.md) | workspace plan cards dropped bound resume context | high | resolved | 2026-07-08 | `fix(workspace): close plan list UX and context regressions (BUG-0143, BUG-0144, BUG-0145)` |
| [BUG-0142](./BUG-0142.md) | targetjob failed detail returned 500 after profile schema pruning | high | resolved | 2026-07-08 | `fix(targetjob): align store with pruned profile schema (BUG-0142)` |
| [BUG-0028](./BUG-0028.md) | targetjob review exposed error envelope and parse runtime drift | medium | resolved | 2026-05-08 | `fix(backend-targetjob): align targetjob review contracts` |
| [BUG-0027](./BUG-0027.md) | targetjob L2 review exposed runtime gate and SSRF gaps | medium | resolved | 2026-05-08 | `fix(backend-targetjob): harden targetjob L2 runtime gates` |
| [BUG-0026](./BUG-0026.md) | targetjob L2 review exposed import parse contract drift | medium | resolved | 2026-05-08 | `fix(backend-targetjob): remediate import parse L2 findings` |
| [BUG-0025](./BUG-0025.md) | analysis failed redline rejected documented provider secret code | medium | resolved | 2026-05-08 | `feat(backend-targetjob): complete import parse bootstrap handoff` |
| [BUG-0024](./BUG-0024.md) | targetjob detail omitted parsed summary provenance | medium | resolved | 2026-05-08 | `feat(backend-targetjob): complete import parse bootstrap handoff` |

## Practice

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0162](./BUG-0162.md) | practice interviewer lost the real resume after parse output truncation | high | resolved | 2026-07-12 | `fix(interview): ground resumes and persist round progress (BUG-0162, BUG-0163)` |
| [BUG-0160](./BUG-0160.md) | conversation simplification review exposed lifecycle, retry, score, and scenario drift | high | resolved | 2026-07-12 | `fix(practice): close conversation review regressions (BUG-0160)` |
| [BUG-0159](./BUG-0159.md) | real conversation funnel retained PostgreSQL and observability drift | high | resolved | 2026-07-12 | `fix(practice): harden real conversation funnel (BUG-0159)` |
| [BUG-0158](./BUG-0158.md) | practice phone flow and session language drifted from the real interview contract | medium | resolved | 2026-07-11 | `fix(practice): align phone flow and session language (BUG-0158)` |
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
| [BUG-0186](./BUG-0186.md) | unanswered terminal prompt leaked into report assessment | high | resolved | 2026-07-18 | `fix(report): exclude unanswered terminal prompt (BUG-0186)` |
| [BUG-0182](./BUG-0182.md) | report retries repeatedly selected a terminal assistant message as evidence | high | resolved | 2026-07-16 | `fix(report): harden regeneration and local diagnostics (BUG-0182, BUG-0183)` |
| [BUG-0173](./BUG-0173.md) | malformed report conversation ID reached the UUID database query | medium | resolved | 2026-07-15 | `feat(report): integrate readonly conversation record (BUG-0173, BUG-0174)` |
| [BUG-0171](./BUG-0171.md) | report generation rejected valid context at a stale hardcoded byte limit | high | resolved | 2026-07-14 | `fix(runtime): centralize content size limits (BUG-0171)` |
| [BUG-0166](./BUG-0166.md) | unpaired report prompt example leaked unsupported facts into live output | high | resolved | 2026-07-13 | `fix(report): ground report semantics and reliability (BUG-0164, BUG-0165, BUG-0166)` |
| [BUG-0165](./BUG-0165.md) | empty report retry focus encoded as null and blocked ready dashboards | high | resolved | 2026-07-13 | `fix(report): ground report semantics and reliability (BUG-0164, BUG-0165, BUG-0166)` |
| [BUG-0064](./BUG-0064.md) | report replay handoff reused source sessions and pixel gate was false-green | high | resolved | 2026-05-16 | `fix(frontend-report): harden replay handoff and pixel gate (BUG-0064)` |
| [BUG-0063](./BUG-0063.md) | frontend report dashboard L2 review exposed route and parity drift | high | resolved | 2026-05-16 | `fix(frontend-report): remediate report dashboard L2 findings (BUG-0063)` |
| [BUG-0062](./BUG-0062.md) | report L2 review exposed prompt schema, persistence, and privacy drift | high | resolved | 2026-05-16 | `fix(backend-review): remediate report L2 findings (BUG-0062)` |
| [BUG-0061](./BUG-0061.md) | report runtime only mounted read routes | high | resolved | 2026-05-16 | `fix(backend-review): wire report runner runtime (BUG-0061)` |
| [BUG-0005](./BUG-0005.md) | report follow-up CTAs returned to setup instead of starting sessions | medium | resolved | 2026-05-02 | `fix(ui-design): start report follow-up sessions directly` |

## Materials

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0187](./BUG-0187.md) | resume parse polling flashed the generic loading state | medium | resolved | 2026-07-18 | `fix(frontend): stabilize resume parse polling (BUG-0187)` |
| [BUG-0140](./BUG-0140.md) | resume detail injected display name into Markdown body and split source preview surfaces | medium | resolved | 2026-07-08 | `fix(resume): close source preview and picker regressions (BUG-0139, BUG-0140, BUG-0141)` |
| [BUG-0139](./BUG-0139.md) | resume detail rendered Markdown as plain text and PDF AI failure kept non-Markdown snapshot | medium | resolved | 2026-07-07 | `fix(resume): close source preview and picker regressions (BUG-0139, BUG-0140, BUG-0141)` |
| [BUG-0138](./BUG-0138.md) | resume detail kept generated-name placeholder and repeated polling risk | high | resolved | 2026-07-07 | `fix(resume-workshop): generate names and stop detail polling (BUG-0138)` |
| [BUG-0137](./BUG-0137.md) | resume creation kept generic names and obsolete parse steps | medium | resolved | 2026-07-07 | `fix(resume-workshop): read full upload PDFs (BUG-0137)` |
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
| [BUG-0065](./BUG-0065.md) | debrief.generate prompt baseline used non-current output schema | high | resolved | 2026-05-16 | `feat(backend-debrief): close 001 debrief record and analysis baseline` |

## Eval

| ID | 标题 | 严重度 | 状态 | 发现日期 | 关联 Commit |
|----|------|--------|------|----------|-------------|
| [BUG-0145](./BUG-0145.md) | target import parse prompt omitted JD identity fields | medium | resolved | 2026-07-08 | `fix(workspace): close plan list UX and context regressions (BUG-0143, BUG-0144, BUG-0145)` |
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
| [BUG-0184](./BUG-0184.md) | interview plan cards duplicated an inactive lifecycle status and empty-location placeholder | low | resolved | 2026-07-17 | `fix(frontend): remove stale plan card metadata (BUG-0184)` |
| [BUG-0176](./BUG-0176.md) | settings review exposed fixture auth and failure-evidence gaps | medium | resolved | 2026-07-15 | `fix(settings): close review regressions (BUG-0176)` |
| [BUG-0170](./BUG-0170.md) | Duplicate safe reads and mixed command/detail route ownership | high | resolved | 2026-07-14 | `fix(core-flow): separate read paths and dedupe requests (BUG-0170)` |
| [BUG-0161](./BUG-0161.md) | practice timing and report progression ignored structured round context | high | resolved | 2026-07-12 | `fix(frontend): align round timing and progression (BUG-0161)` |
| [BUG-0154](./BUG-0154.md) | resume create prototype discarded the created detail handoff | medium | resolved | 2026-07-10 | `fix(ui-design): preserve created resume detail handoff (BUG-0154)` |
| [BUG-0152](./BUG-0152.md) | home and workspace quick-start dropped target context | medium | resolved | 2026-07-09 | `fix(frontend): preserve quick-start target context (BUG-0152)` |
| [BUG-0148](./BUG-0148.md) | round assumptions ignored structured LLM-derived interview rounds | medium | resolved | 2026-07-09 | `-` |
| [BUG-0143](./BUG-0143.md) | interview plan list cards lacked visual hierarchy and concise content | medium | resolved | 2026-07-08 | `fix(workspace): close plan list UX and context regressions (BUG-0143, BUG-0144, BUG-0145)` |
| [BUG-0141](./BUG-0141.md) | home resume picker hid readable existing resumes | medium | resolved | 2026-07-08 | `fix(resume): close source preview and picker regressions (BUG-0139, BUG-0140, BUG-0141)` |
| [BUG-0134](./BUG-0134.md) | home JD source controls over-separated from input card | medium | resolved | 2026-07-06 | `fix(frontend-home): integrate source controls into input card (BUG-0134)` |
| [BUG-0133](./BUG-0133.md) | home quick-start source and CTA layout drifted from planning flow | medium | resolved | 2026-07-06 | `fix(frontend-home): separate source and submit layout (BUG-0133)` |
| [BUG-0132](./BUG-0132.md) | home resume selector and recent mocks drifted from compact shortcut UX | medium | resolved | 2026-07-06 | `fix(frontend-home): use resume dropdown and cap recent mocks (BUG-0132)` |
| [BUG-0131](./BUG-0131.md) | home immediate interview skipped required resume pre-bind | medium | resolved | 2026-07-06 | `fix(frontend-home): prebind resume for immediate interview (BUG-0131)` |
| [BUG-0130](./BUG-0130.md) | parse launch allowed practice planning without a bound resume | high | resolved | 2026-06-30 | `fix(frontend-home): require explicit resume selection (BUG-0130)` |
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
| [BUG-0185](./BUG-0185.md) | resume tailor task run succeeded before business output was accepted | medium | resolved | 2026-07-17 | `fix(ai): close review findings (BUG-0185)` |
| [BUG-0183](./BUG-0183.md) | host-run SMTP port drift left the local Mailpit mailbox empty | medium | resolved | 2026-07-16 | `fix(report): harden regeneration and local diagnostics (BUG-0182, BUG-0183)` |
| [BUG-0180](./BUG-0180.md) | production email review exposed delivery lifecycle and PID ownership gaps | high | resolved | 2026-07-16 | `fix(platform): harden smtp lifecycle and pid ownership (BUG-0180)` |
| [BUG-0178](./BUG-0178.md) | host-run backend claimed full-container email jobs without the delivery secret | medium | resolved | 2026-07-16 | `feat(auth): add production smtp delivery (BUG-0178)` |
| [BUG-0155](./BUG-0155.md) | change-intake overcounted generic terms and ignored exact scenario owners | medium | resolved | 2026-07-10 | `fix(change-intake): prioritize exact owner evidence (BUG-0155)` |
| [BUG-0127](./BUG-0127.md) | host-run backend wildcard bind left resume page on stale 500 | medium | resolved | 2026-06-15 | `fix(local-dev-stack): bind host backend to loopback (BUG-0127)` |
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
| [BUG-0163](./BUG-0163.md) | completed practice rounds were not persisted as TargetJob progress | high | resolved | 2026-07-12 | `fix(interview): ground resumes and persist round progress (BUG-0162, BUG-0163)` |
| [BUG-0135](./BUG-0135.md) | migration privacy matrix omitted idempotency records | medium | resolved | 2026-07-06 | `fix(schema): cover idempotency in privacy matrix (BUG-0135)` |
| [BUG-0126](./BUG-0126.md) | resume flatten follow-ups exposed display, duplicate, rollback, and lint drift | medium | resolved | 2026-06-15 | `fix(review): close resume flatten follow-up gaps (BUG-0126)` |
| [BUG-0125](./BUG-0125.md) | resume archive persistence and P0.102 gate drift escaped review | medium | resolved | 2026-06-15 | `fix(review): persist resume archive and p0102 gate (BUG-0125)` |
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
| [BUG-0181](./BUG-0181.md) | Full staticcheck and UI pruning baselines blocked AI transport closeout | medium | resolved | 2026-07-16 | `refactor(ai): adopt openai-go and restore quality gates (BUG-0181)` |
| [BUG-0179](./BUG-0179.md) | owner checklist compression deleted activation markers and preflight accepted incidental mentions | medium | resolved | 2026-07-16 | `feat(auth): use Redis for email delivery secrets (BUG-0179)` |
| [BUG-0177](./BUG-0177.md) | OPENAPI-001 refreeze left its root regression test bound to the new baseline | medium | resolved | 2026-07-15 | `feat(report): align context grid and acceptance (BUG-0177)` |
| [BUG-0175](./BUG-0175.md) | E2E evidence redaction missed URL-encoded account email | medium | resolved | 2026-07-15 | `feat(settings): simplify account settings and harden evidence privacy (BUG-0175)` |
| [BUG-0174](./BUG-0174.md) | evidence privacy gate rejected benign PNG metadata | medium | resolved | 2026-07-15 | `feat(report): integrate readonly conversation record (BUG-0173, BUG-0174)` |
| [BUG-0172](./BUG-0172.md) | UI Demo pruning gate missed renamed browser contracts and owner drift | medium | resolved | 2026-07-15 | `fix(ui-design): close pruning review findings (BUG-0172)` |
| [BUG-0169](./BUG-0169.md) | Closeout contract gates drifted from current source ownership | medium | resolved | 2026-07-14 | `fix(interview): close turn UX and report navigation (BUG-0167, BUG-0168, BUG-0169)` |
| [BUG-0168](./BUG-0168.md) | P0.006 verifier treated business failed counters as Playwright failures | medium | resolved | 2026-07-14 | `fix(interview): close turn UX and report navigation (BUG-0167, BUG-0168, BUG-0169)` |
| [BUG-0167](./BUG-0167.md) | P0.058 report evidence fixture omitted required practice reply state | medium | resolved | 2026-07-14 | `fix(interview): close turn UX and report navigation (BUG-0167, BUG-0168, BUG-0169)` |
| [BUG-0164](./BUG-0164.md) | P0.058 verifier counted Go subtests as root tests | medium | resolved | 2026-07-12 | `fix(report): ground report semantics and reliability (BUG-0164, BUG-0165, BUG-0166)` |
| [BUG-0157](./BUG-0157.md) | P0.098 shadowed TargetJob fixture left a stale persistence assertion | medium | resolved | 2026-07-10 | `fix(test): align full-funnel target fixture (BUG-0157)` |
| [BUG-0156](./BUG-0156.md) | root test gate omitted Python tooling and skill contracts | medium | resolved | 2026-07-10 | `fix(test): aggregate Python contract tests (BUG-0156)` |
| [BUG-0153](./BUG-0153.md) | P0.037 request-count wait raced ready DOM commit | medium | resolved | 2026-07-10 | `fix(test): wait for P0.037 ready DOM (BUG-0153)` |
| [BUG-0136](./BUG-0136.md) | repo pruning cleanup review exposed lint and frontend guard drift | medium | resolved | 2026-07-07 | `fix(review): close repo pruning cleanup findings (BUG-0136)` |
| [BUG-0129](./BUG-0129.md) | core-loop pruning left stale profile and codegen test gates | medium | resolved | 2026-06-30 | `fix(core-loop): close pruning gate drift (BUG-0129)` |
| [BUG-0128](./BUG-0128.md) | core-loop pruning gates missed design canvas and privacy hook drift | medium | resolved | 2026-06-29 | `fix(core-loop): close pruning review drift (BUG-0128)` |
| [BUG-0124](./BUG-0124.md) | resume Phase 8 flat gate drifted on current profile fields | medium | resolved | 2026-06-14 | `fix(test): close resume phase8 flat gate drift (BUG-0124)` |
| [BUG-0120](./BUG-0120.md) | D-20 resume gates missed non-current fixture keys and stale scenario scripts | medium | resolved | 2026-06-14 | `fix(test): close resume d20 gate drift (BUG-0120)` |
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
| [BUG-0011](./BUG-0011.md) | mock contract gate ignored empty non-current fixture tag directories | medium | resolved | 2026-05-06 | `fix(mock-contract): reject non-current fixture tag dirs` |
| [BUG-0010](./BUG-0010.md) | mock contract runtime gate missed registry and stale route count | medium | resolved | 2026-05-05 | `fix(mock-contract): harden runtime drift gates` |
| [BUG-0003](./BUG-0003.md) | local quality gates skipped real backend and frontend execution | medium | resolved | 2026-04-30 | `fix(ci-pipeline): remediate local quality gates` |
