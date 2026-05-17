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
| `saved` | "已保存 v1 主版本 · 进入简历工坊" | `resume_versions` (list) | — |
| `already_exists` | "已存在主版本 · 跳转查看" | `resume_versions?versionId=<masterId>&tab=preview` | — |
| `validation` | — | (no nav) | `resume-preview-confirm-inline-error` |
| `error` | — | (no nav) | generic `resumeWorkshop.create.errors.confirmFailed` |

## Idempotency

- `confirmResumeStructuredMaster` request carries `Idempotency-Key` matching `v1.<unix>.<uuidv7>`
- Same-asset replay reuses the same IK (cache hit)
- 422 + new attempt → fresh IK (cache cleared)

## Home / Workspace CTA

- Home CTA testid: `home-resume-create`
- Workspace MissingResumeState CTA: navigates to `resume_versions?flow=create`
- Unauthenticated route lands on `resume-workshop-auth-gate`; pendingAction params contain `{ flow: "create" }` and NOT raw text / structuredProfile / rawText

## Privacy

- structuredProfile JSON content does NOT appear in nav params
- localStorage / sessionStorage receive no setItem calls during confirm

## Trigger Log Assertions

- `Test Files +\d+ passed` matches
- Linked test files present in log
- Test names mentioning `409`, `422`, `replay`, `Home`, `Workspace` are exercised
