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
(i18n coverage, debrief privacy boundary, devMock fixture registry) plus
the legacy negative grep gate over both the implementation tree and the
P0.065-069 scenario tree.

## Then

Vitest exits 0 and the legacy negative gate reports zero offending
matches in the scoped paths.

## Pixel parity status

Full Playwright pixel parity (`pnpm test:pixel-parity`) requires the
Playwright chromium runtime which is **not** installed in this scenario
runner. The frontend `test:pixel-parity` target stays available for
on-demand visual regression — see `frontend/tests/pixel-parity/` follow
up scoped to debrief when the install step lands. This scenario asserts
the structural + privacy + legacy gates that do not require a browser.
