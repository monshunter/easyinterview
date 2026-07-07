# P0.086 Expected Outcome

- Accept-only Rewrites never call non-current accept/reject routes.
- Overwrite save uses `updateResume`; save-as-new uses `duplicateResume`.
- Non-current `updateResumeVersion` operation token is absent from runtime source.
- 422 validation alert appears in PreviewStage/Edit Tab save flows.
- Privacy: structuredProfile content never appears in URL / localStorage / fetch transport log / toast text.
