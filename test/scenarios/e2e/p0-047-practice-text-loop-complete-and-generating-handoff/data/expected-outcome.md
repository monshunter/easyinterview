# Expected outcome

Opening-only and pending-reply conversations cannot invoke completion; the Finish control is natively disabled with a localized accessible reason. Direct backend attempts return `VALIDATION_FAILED` and write no completion/report side effects. The answered conversation atomically freezes `report-context.v1`, creates one report job, and hands only `reportId` to generating. Concurrent mutable context changes block until completion commits, and replay returns the identical snapshot without a second `session_completed` fact or report job.
