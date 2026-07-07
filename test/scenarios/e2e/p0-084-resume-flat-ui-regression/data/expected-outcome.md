# P0.084 Expected Outcome

- Flat list/detail/create/rewrites/auth-gate specs pass.
- Rewrites accept-only modal can save overwrite and save-as-new via current
  flat handlers.
- Non-current operation tokens stay absent from runtime source:
  `ResumeBranchFlow`, `branchResumeVersion`, `seedStrategy`,
  `acceptResumeTailorSuggestion`, `rejectResumeTailorSuggestion`, and
  `updateResumeVersion`.
- Privacy: accepted rewrite content never appears in pendingAction params / URL
  / localStorage / fetch transport log.
