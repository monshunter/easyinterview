# UI-Design Pixel Parity Gate BDD Checklist

> **版本**: 1.6
> **状态**: completed
> **更新日期**: 2026-07-07

**关联 BDD Plan**: [bdd-plan](./bdd-plan.md)

## E2E.P0.006 real-browser ui-design pixel parity gate

- [x] Scenario assets exist under `test/scenarios/e2e/p0-006-ui-design-pixel-parity-gate/`
- [x] Setup verifies Chromium cache and `frontend/dist/index.html`
- [x] Trigger runs `pnpm --filter @easyinterview/frontend test:pixel-parity`
- [x] Verify requires passing summary, desktop/mobile project markers and all current 13 spec markers
- [x] Verify keeps out-of-scope route/module entries negative-only
- [x] Cleanup removes setup marker and preserves logs for evidence

## Closeout

- [x] `pnpm --filter @easyinterview/frontend build` passes
- [x] `pnpm --filter @easyinterview/frontend test:pixel-parity` passes
- [x] `validate_context.py --context docs/spec/frontend-shell/plans/003-ui-design-pixel-parity-gate/context.yaml --target frontend` passes
- [x] `sync-doc-index --check` passes
- [x] `make docs-check` passes
