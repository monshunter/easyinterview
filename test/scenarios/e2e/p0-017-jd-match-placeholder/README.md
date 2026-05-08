# E2E.P0.017 — JD Match P1 Placeholder Smoke

> **Scenario ID**: E2E.P0.017
> **Owner**: frontend-home-job-picks-and-parse/001-home-jd-import-and-parse
> **Status**: Ready
> **Execution**: automated

## Scope

Verifies the jd_match P1 placeholder shell renders correctly:
- Hero (label/title/sub) DOM anchors
- Profile snapshot chip (static, no real profile)
- Three tab labels (Recommended/Search/Watchlist)
- Placeholder content area with "Coming Soon" copy
- Old prototype business testids NOT present (negative search)
- Route reachable from both TopBar and Home aux card

## Fixtures

No new fixtures required — placeholder consumes no API data.

## Verification Points

- jdmatch-hero-label, jdmatch-hero-title, jdmatch-hero-sub present
- jdmatch-profile-chip, jdmatch-profile-chip-title present
- jdmatch-tab-recommended, jdmatch-tab-search, jdmatch-tab-watchlist present
- jdmatch-placeholder, jdmatch-placeholder-cta present
- Old prototype testids (jdmatch-card-*, jdmatch-saved-search-*, etc.) absent
- i18n zh/en switching
- Theme switching (warm/dark/customAccent)
- Mobile responsive (no overflow)
- TopBar topbar-nav-jd_match highlighted

## Scripts

- `scripts/setup.sh` — ensure frontend dist exists
- `scripts/trigger.sh` — navigate to jd_match via TopBar + home aux card
- `scripts/verify.sh` — assert DOM anchors, negative testid search, i18n
- `scripts/cleanup.sh` — reset state

## Offline Limitations

- No mock transport required
- Placeholder only; real recommendations data gated by plan 002-jd-match-recommendations
