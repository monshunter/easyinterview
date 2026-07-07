# P0.084 Expected Outcome

- Flat list/detail/create/auth-gate specs pass.
- Resume detail remains read-only and does not expose Rewrites/Edit/export/copy
  or original-modal surfaces.
- Resume detail body shows original/parsed text before structured fallback.
- Non-current operation tokens stay absent from runtime source:
  `ResumeBranchFlow`, `branchResumeVersion`, `seedStrategy`,
  `acceptResumeTailorSuggestion`, `rejectResumeTailorSuggestion`, and
  `updateResumeVersion`.
- Privacy: resume body and structured fields never appear in pendingAction params
  / URL / localStorage / fetch transport log.
