# P0.086 Expected Outcome

- Accept / reject hooks are bodyless (no `manualEditText` body) and carry IK header.
- Manual edit two-step: updateResumeVersion(structuredProfile.manualEdits[]) -> bodyless accept; saved-manual-pending alert appears when accept fails post-update.
- D-12: accept/reject never call updateResumeVersion; structuredProfile remains unchanged after pure accept.
- 422 validation alert in Edit Tab; 409 idempotency conflict alert; 404 cross-user generic alert; all routed via UpdateResumeVersionError / SuggestionDecisionError discriminated kinds.
- Privacy: structuredProfile content never appears in URL / localStorage / fetch transport log / toast text.
