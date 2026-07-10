# Expected Outcome

- `importTargetJob` returns `202 + TargetJobWithJob`.
- Job satisfies the terminal-state contract: `jobType=target_import`, `status=succeeded`.
- TargetJob is immediately `analysisStatus=ready` with a `must_have` draft requirement.
- No runner source row, async job payload, `target.import.requested`, or `target.parsed` event is generated.
- List and detail expose the ready manual-form TargetJob.
