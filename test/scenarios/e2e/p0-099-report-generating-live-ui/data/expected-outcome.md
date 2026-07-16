# Expected Outcome

- Exactly six canonical full-page PNGs exist: zh needs-practice, en
  well-prepared, and honest generating, each at desktop and mobile viewport.
- Every `screenshot_sha256` matches a fully parsed, CRC-valid PNG with terminal
  `IEND`, no trailing bytes, no metadata-bearing chunk, and executable nonblank
  visual content; each ready mobile bottom viewport is also nonblank.
- Current-run manifest evidence binds run, report, session, frozen context,
  ready report content, and screenshot digests, while independent
  trigger-generated live HTTP and read-only PostgreSQL projections prove current
  identity/state/content rather than allowing manifest self-consistency alone.
- Setup, session, report, and capture microsecond timestamps prove
  `setup_at < session_created_at <= report_created_at <= captured_at` for all
  three report resources.
- The active backend has local raw I/O capture enabled at a resolved path
  outside current-run evidence; neither backend logs nor scenario artifacts
  contain captured request/response content.
- Ready screenshots show one four-peer target/round/frozen-resume/interview-record
  Context Strip, a visible frozen-resume copy link, responsive detail-pair
  alignment (desktop equal-height rows with internal whitespace; mobile
  natural-height single column), direct preparedness, dimension/evidence
  content, executable action labels, and the bottom full-width interview summary
  after the action region; action labels agree with current DB/API summaries and pass
  the same `en <=24 words` / `zh-CN <=64 Unicode code points` redacted count audit.
- Generating screenshots make no false claim that analysis is already ready.
- One current `manifest.json` maps each file to locale, state, fixture,
  viewport, report/session, DB/API state, and redacted content audit; its machine
  binding fields were atomically replaced by the trigger-time trusted capture.
- One current `live-capture.json` contains only report/session references,
  current-run timestamps, state, preparedness, content/action counts, frozen
  context and canonical content digests, and explicit no-secret/no-raw/no-prose
  privacy booleans.
- One `manual-visual-audit.json`, produced by direct image review without OCR,
  binds state-specific all-true visibility/completeness checks to the SHA-256 of
  each exact canonical PNG; only then can verify promote the result to `PASS`.
- No browser cookie/state, raw frozen context, prompt/response, JD, resume, or
  transcript prose is persisted in the scenario output.
- Package tests, source checks, builds, and static screenshot baselines are not
  accepted as E2E PASS evidence.
