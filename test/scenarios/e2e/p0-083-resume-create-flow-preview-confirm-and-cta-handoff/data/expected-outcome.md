# P0.083 Expected Outcome

## PreviewConfirm DOM

- `resume-preview-confirm` testid renders after polling resolves ready
- Header anchors: `resume-preview-confirm-back-link`, `resume-preview-confirm-back-button`, `resume-preview-confirm-save-button`, `resume-preview-confirm-source`, `resume-preview-confirm-name`
- Sidebar anchors: `resume-preview-confirm-sidebar-what-saved`, `resume-preview-confirm-sidebar-parse-notes`
- Content anchor: `resume-preview-confirm-content`
- Inline error anchor: `resume-preview-confirm-inline-error` (only on 422)

## Confirm Outcome Matrix

| Outcome | Toast | Navigation | Inline Error |
|---------|-------|------------|--------------|
| `saved` | save success toast | `resume_versions` (list) | — |
| `validation` | — | (no nav) | `resume-preview-confirm-inline-error` |
| `error` | — | (no nav) | generic `resumeWorkshop.create.errors.confirmFailed` |

## Idempotency

- `updateResume` request carries `Idempotency-Key` matching `v1.<unix>.<uuidv7>`

## Home / Workspace CTA

- Home CTA testid: `home-resume-create`
- Workspace MissingResumeState CTA: navigates to `resume_versions?flow=create`
- Unauthenticated route lands on auth login pending-action state; pendingAction
  params contain `{ flow: "create" }` and NOT raw text / structuredProfile /
  rawText

## Privacy

- structuredProfile JSON content does NOT appear in nav params
- localStorage / sessionStorage receive no setItem calls during confirm

## Trigger Log Assertions

- `Test Files +\d+ passed` matches
- Linked test files present in log
- Test names mentioning `updateResume`, `422`, `guided`, `Home`, `Workspace`
  are exercised
