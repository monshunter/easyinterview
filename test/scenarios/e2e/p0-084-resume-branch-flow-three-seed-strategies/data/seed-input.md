# P0.084 Seed Input

- `Resumes/listResumes.json` `default` — Asset `01918fa0-0000-7000-8000-000000001000`.
- `Resumes/listResumeVersions.json` `default` — `structured_master` version `0195f2d0-0001-7000-8000-000000000201` + targeted `...202`.
- `Resumes/branchResumeVersion.json` scenarios:
  - `default` / `copy-master-sync` (201 + ResumeVersion, seedStrategy=copy_master)
  - `blank-sync` (201 + ResumeVersion, seedStrategy=blank)
  - `ai-select-202-with-job` (202 + BranchResumeVersionAccepted, job=resume_tailor/queued, resourceId=tailorRunId)
  - `idempotent-replay` (201 + canonical replay)
  - `validation-error-422` (422 + VALIDATION_FAILED + details.field=displayName)
- User A authenticated; User B authenticated (cross-user); lang=zh-CN.
