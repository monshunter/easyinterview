# E2E.P0.069 Debrief Pixel Parity + Privacy + Legacy Negative

> **场景 ID**: E2E.P0.069
> **关联需求**: frontend-debrief C-15, C-16, C-17, C-18
> **隔离级别**: shared-cluster
> **parallel-safe**: No
> **自动化入口**: `scripts/setup.sh` -> `scripts/trigger.sh` -> `scripts/verify.sh` -> `scripts/cleanup.sh`

## Given

`DebriefScreen` end-to-end implementation is mounted; i18n debrief.*
namespace covers zh + en; debrief.css drives the theme + responsive
breakpoints via tokens.

## When

Run the Vitest assertions that exercise the parity-adjacent contracts
(i18n coverage, debrief privacy boundary, devMock fixture registry), build
the frontend dist, run `tests/pixel-parity/debrief.spec.ts` in Playwright
desktop/mobile projects, then execute the legacy negative grep gate over both
the implementation tree and the P0.065-069 scenario tree.

## Then

Vitest exits 0, the debrief Playwright parity gate passes, and the legacy
negative gate reports zero offending matches in the scoped paths.

## Pixel parity status

`scripts/trigger.sh` now runs the debrief-specific Playwright parity gate:
`pnpm --filter @easyinterview/frontend build` followed by
`pnpm --filter @easyinterview/frontend exec playwright test
tests/pixel-parity/debrief.spec.ts`. The gate asserts DOM anchors,
bounding boxes, responsive geometry, theme/customAccent computed values, and
non-empty screenshot smoke without requiring checked-in screenshot baselines.
