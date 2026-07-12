# Expected outcome

The UI enters generating with plan/session/report context and no structured question or mode payload. Completion replay produces neither a second `session_completed` fact nor a second report job, so downstream round projection advances exactly once.
