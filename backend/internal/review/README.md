# Review service

The review service loads one completed practice conversation, orders its messages, and invokes `report.generate` once. It persists conversation-level readiness, dimension assessments, highlights, issues, next actions, competency focus codes, and generation provenance.

Reports do not contain per-question assessments, turn identifiers, or question-scoped retry state. Failures persist a typed failed report without exposing partial ready data.
