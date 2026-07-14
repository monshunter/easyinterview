# E2E.P0.099 Exact Six-Image Report Acceptance

> **Status**: Ready
> **Owner plan**: [e2e-scenarios-p0/001](../../../../docs/spec/e2e-scenarios-p0/plans/001-real-api-ui-journeys/plan.md)
> **Execution**: hybrid, serial only
> **Isolation**: shared host-run environment, synthetic account/data
> **Parallel-safe**: No

## Given / When / Then

- **Given** the shared v19 environment, real frontend/backend/provider, and three
  isolated current report resources: Chinese `needs_practice`, English
  `well_prepared`, and honest `generating`. The active backend was redeployed
  after setting `AI_DEBUG_PRINT_RAW_OUTPUT=false` in the local mode-`0600`
  `deploy/dev-stack/.env`.
- **When** a browser agent captures the exact desktop/mobile matrix with
  `fullPage: true`, records one provisional redacted `manifest.json` containing
  screenshot/ref/semantic-audit facts, and passes the current synthetic session
  cookie through temporary `P0_099_SESSION_COOKIE`.
- **Then** exactly six canonical PNGs exist, every row binds locale, state,
  viewport, fixture, report/session references, DB/API state, and a redacted
  fact-to-judgment-to-action audit; trigger independently captures the same
  three resources through authenticated live HTTP plus `read-only-postgres`,
  and a no-OCR review binds visible semantics to the exact six PNG digests.

This scenario no longer accepts the historical four practice/report images.
P0.099 is the minimal real report/generating visual acceptance required by
frontend-report-dashboard C-7; practice interaction remains a code-level gate
until a dedicated real API/UI journey is added.

## Exact screenshot matrix

| File | Locale / state | Viewport | Fixture label |
|------|----------------|----------|---------------|
| `report-zh-needs-practice-desktop.png` | zh / ready-needs-practice | 1440x1200 | `real-provider-ready-needs-practice-long` |
| `report-zh-needs-practice-mobile.png` | zh / ready-needs-practice | 390x844 | same report/session |
| `report-en-well-prepared-desktop.png` | en / ready-well-prepared | 1440x1200 | `real-provider-ready-well-prepared-long` |
| `report-en-well-prepared-mobile.png` | en / ready-well-prepared | 390x844 | same report/session |
| `report-generating-desktop.png` | zh / generating | 1440x1200 | `real-provider-generating-long` |
| `report-generating-mobile.png` | zh / generating | 390x844 | same report/session |

No seventh PNG, alternate filename, partial-page capture, or reused report ID
is accepted. Desktop/mobile rows for one state must share one report/session;
the three states must use isolated resources.

Every PNG is parsed as a complete chunk stream: signature, chunk boundaries,
CRC, one terminal `IEND`, no trailing bytes, and no textual/EXIF/profile/time
metadata. Each manifest row carries `screenshot_sha256`, which must equal the
current file digest. The validator reconstructs PNG filters and rejects blank
or near-solid images. For each ready mobile report it additionally samples the
last 844-pixel viewport and rejects a blank bottom region, so a top-only or
truncated capture cannot satisfy the full-page action-region gate. These pixel
checks prove file integrity and nonblank layout only; they do not infer report
semantics and are never treated as OCR.

## Browser-agent capture

Use a named session and never export browser state or cookies:

```bash
agent-browser --session p0-099-report open 'http://127.0.0.1:5173/report?reportId=<opaque-report-ref>'
agent-browser --session p0-099-report wait --load networkidle
agent-browser --session p0-099-report set viewport 1440 1200
agent-browser --session p0-099-report screenshot --full .test-output/e2e/p0-099-report-generating-live-ui/screenshots/<canonical-desktop-name>.png
agent-browser --session p0-099-report set viewport 390 844
agent-browser --session p0-099-report screenshot --full .test-output/e2e/p0-099-report-generating-live-ui/screenshots/<canonical-mobile-name>.png
```

Repeat for the three rows in the matrix. Capture `generating` immediately after
the real completion handoff while the API still returns `generating`; do not use
a fixture transport, route status override, or hidden mock server. Close the
named session after capture.

## Redacted manifest

The browser-authored provisional manifest contains the top-level current
`scenario_id`, `run_id`, `capture_contract`, `privacy`, and exact six screenshot
rows. Each row contains its canonical file/locale/state/fixture/viewport,
`full_page: true`, opaque report/session refs, PNG SHA-256, and only
`evidence.content_audit`. That audit has exactly:

- reviewer-owned `fact_to_judgment_to_action`, `unsupported_count`,
  `irrelevant_advice_count`, and `causal_mismatch_count`;
- placeholder `item_verdict_count` and `action_label_audit`, which trigger
  replaces from the live public API projection.

The final trigger-bound manifest additionally contains only:

- current `run_id` and opaque report/session references;
- locale/state/fixture/viewport plus `full_page: true`;
- DB/API status and preparedness summaries;
- SHA-256 of frozen context, never the context itself;
- one canonical report-content SHA-256 shared by the current DB/API capture for
  each ready report; generating records `null` because no content exists yet;
- ready-report item count and zero unsupported/irrelevant/causal-mismatch
  counts; generating uses `not_applicable` and zero items;
- ready-report action-label language/unit/limit/counts proving `en <=24` words
  or `zh-CN <=64` Unicode code points without retaining duplicate label prose;
- privacy booleans proving cookie/raw frozen context were not written.

DB/API evidence comes from trigger's current-run trusted read-only DB/API
capture, never from a hand-entered or historical result. The browser-authored
provisional manifest deliberately has no `collection` / `db` / `api` object.
`capture_live_evidence.py --bind-manifest` atomically creates those objects and
replaces the two machine audit fields from the successful live capture while
preserving only the bounded semantic verdict fields. Every final row's
`evidence.collection` uses `trusted-current-run-db-api-capture` and binds the
same current `run_id`, opaque report reference, session reference, frozen
context digest, ready report-content digest, and `screenshot_sha256` that the
validator independently recomputes. The real screenshots only need to remain
within the language limit; E2E does not duplicate exact boundary-value
assertions for those labels. This acceptance evidence does not add a product
API field.

Manifest self-consistency is not live-binding evidence. On every trigger run,
`capture_live_evidence.py` writes a `p0-099-live-capture.v2`
`live-capture.json` projection for the exact three manifest resources. It joins
`feedback_reports` to `practice_sessions` through one read-only PostgreSQL
`SELECT`, then calls `GET /api/v1/reports/{reportId}`. Trigger performs both
captures immediately after setup/run identity checks and before the environment
readiness recheck, so an honest generating state cannot advance during the
runner and then be described from stale state.

The `read-only-postgres` projection proves the report/session relation, DB
status, preparedness, frozen-context digest, canonical report-content digest,
and microsecond UTC `session_created_at` / `report_created_at`. The validator
requires:

```text
setup_at < session_created_at <= report_created_at <= captured_at
```

It also requires exact DB/API equality for session, state, preparedness, and
ready-content digest. The persisted projection contains only opaque references,
timestamps, state, counts, and digests. Ready prose is canonically hashed in
memory. Before hashing, DB `highlights` / `issues` are projected through the
same public API shape, so internal `sourceMessageSeqNos` anchors participate in
neither the public digest nor the persisted artifact. Raw DB rows, frozen
context, API responses, anchors, and prose are never written.

Trigger takes `P0_099_DATABASE_URL` only as an optional process-local override;
otherwise it reads `DATABASE_URL` from the mode-`0600`
`deploy/dev-stack/.env`. It passes the value only in the capture subprocess
environment, never on the command line, and unsets its local copy immediately
after capture. The capture subprocess invokes `psql` with connection fields in
scoped `PG*` variables and sends the fixed `SELECT` over stdin.

This single trigger-time bind removes any bootstrap/backfill cycle: there is no
pre-capture command and no interval in which `generating` can be copied from an
older projection. A capture/bind failure leaves no trusted final manifest.

`P0_099_SESSION_COOKIE` contains only the current synthetic session cookie
value, without the `ei_session=` name. It is read from the process environment,
used as an HTTP header in memory, and never printed, persisted, or passed as a
command-line argument. Missing/rejected cookie, missing `psql`, or an unavailable
live API/DB yields fail-closed `MANUAL_REQUIRED`; malformed capture evidence,
stale identity, state drift, or digest/count mismatch yields `FAIL`.

## Exact-six manual visual audit

After the automated PNG, current-run DB/API, privacy, and raw-debug gates pass,
trigger writes result `MANUAL_REQUIRED` with reason
`awaiting exact-six manual visual audit`. It retains the bounded evidence for
review and prints `P0_099_AUTOMATED_EVIDENCE_PASS`.

Review the six canonical PNGs directly in an image viewer. Do not use OCR or
copy visible prose. Write `manual-visual-audit.json` with exactly:

- `schema_version: p0-099-manual-visual-audit.v1`;
- `scenario_id: E2E.P0.099`, the current setup `run_id`,
  `method: manual-image-review-no-ocr`, and `result: PASS`;
- exactly six rows, each containing only its canonical `file`, the matching
  manifest `screenshot_sha256`, and the state-specific boolean `checks`;
- privacy object `{"ocr_used": false, "prose_transcribed": false,
  "raw_content_written": false}`.

For each ready row, the exact checks are all `true`:

```text
report_page_visible
expected_state_visible
preparedness_visible
dimension_and_evidence_content_visible
action_region_visible
action_labels_complete_without_clipping_or_ellipsis
horizontal_overflow_absent
```

For each generating row, the exact checks are all `true`:

```text
report_page_visible
expected_state_visible
generating_indicator_visible
ready_content_absent
false_ready_claim_absent
clipping_or_overlap_absent
horizontal_overflow_absent
```

No extra check, prose, OCR result, raw page content, or seventh row is allowed.
Run `verify.sh` after writing the artifact. It recomputes every PNG SHA-256,
validates the full automated contract again, binds the audit to the exact six
files, appends `P0_099_MANUAL_VISUAL_AUDIT_BOUND_PASS`, and changes
`result.json` to `PASS`.

Setup records the current backend log byte boundary. Trigger rejects any
`AI_RAW_OUTPUT_DEBUG_BEGIN` / `AI_RAW_OUTPUT_DEBUG_END` marker after that
boundary, proving the current run did not write raw model output to its log.

The executable schema is owned by `scripts/validate_evidence.py`. Do not add
`evidence.md`, cookie jars, browser state exports, raw API bodies, JD/resume/
transcript text, or provider prompt/response bodies to the output directory.
Setup removes every pre-existing top-level file, directory, and symlink without
following external symlink targets; later stages retain only the manifest,
canonical screenshots, and bounded setup/trigger/result/cleanup evidence.

## Execution

After the top-level scenario environment preflight:

```bash
bash test/scenarios/e2e/p0-099-report-generating-live-ui/scripts/setup.sh
# browser captures six images and writes provisional manifest.json for setup.env RUN_ID
read -r -s P0_099_SESSION_COOKIE && export P0_099_SESSION_COOKIE
# trigger captures DB/API and atomically binds all machine evidence into manifest.json
bash test/scenarios/e2e/p0-099-report-generating-live-ui/scripts/trigger.sh
unset P0_099_SESSION_COOKIE
# inspect the exact six PNGs without OCR, then write manual-visual-audit.json
bash test/scenarios/e2e/p0-099-report-generating-live-ui/scripts/verify.sh
bash test/scenarios/e2e/p0-099-report-generating-live-ui/scripts/cleanup.sh
```

Missing browser/live-capture prerequisites yield `MANUAL_REQUIRED` and remove
untrusted primary evidence. A completed automated gate retains its bounded six
images only while awaiting the exact manual audit. Present but malformed, stale,
non-redacted, non-full-page, non-visual, or non-exact evidence yields `FAIL` and
is deleted. Files containing forbidden raw keys or secret markers are deleted
regardless of filename. Full validator-confirmed `PASS` retains the redacted
evidence set.

Output stays under:

```text
.test-output/e2e/p0-099-report-generating-live-ui/
```
