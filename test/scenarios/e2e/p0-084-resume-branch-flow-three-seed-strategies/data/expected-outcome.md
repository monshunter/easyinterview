# P0.084 Expected Outcome

- Form renders with default focus=platform + seed=copy_master, submit disabled while name/target empty.
- copy_master / blank / ai_select dispatch the matching nav target with seedStrategy / Idempotency-Key on the wire.
- ai_select drops `tailorRunId` (from BranchResumeVersionAccepted.job.resourceId) into nav params for Phase 5 polling.
- IK replays on identical fingerprint; rotates on focus change; clears on 422.
- 422 path: `resume-branch-error` alert renders with localized validation copy; no nav; no row created.
- 404 cross-user: generic localized toast; pendingAction stays scrubbed.
- Privacy: form draft fields never appear in pendingAction params / URL / localStorage / fetch transport log.
