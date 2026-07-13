# Review service

The review service consumes the immutable `report-context.v1` snapshot written by practice completion together with the terminal, sequence-ordered `practice_messages`. It does not join mutable TargetJob, Resume, or PracticePlan rows to rebuild report input. Missing snapshots, row-identity drift, session-language drift, message-count/last-sequence drift, and invalid message role/order fail closed.

Every queued, generating, ready, or failed read projects the same minimal public context from that frozen snapshot. Ready reports expose the direct summary, dimension code/label assessments, evidence, actions, report-local focus, and generation provenance without semantic reshaping. Internal `sourceMessageSeqNos` remain available for grounding validation but are intentionally omitted by the API mapper.

Reports do not contain per-question assessments, turn identifiers, or question-scoped retry state. Failures persist a typed failed report without exposing partial ready data.

Each `GenerateReport` invocation owns an in-memory retry session: one initial provider call plus at most three retries, separated by context-aware waits of 10, 20, and 40 seconds. The state is discarded when the invocation returns, so a later user action starts again from its initial call. `feedback_reports` stores no retry counter, and `async_jobs.attempts/max_attempts` remain infrastructure lease/finalization state rather than product retry state.
