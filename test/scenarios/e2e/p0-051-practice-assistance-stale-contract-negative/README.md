# E2E.P0.051 — Assistance stale-contract negative

> **Owner**: backend-practice/003-mode-policies-and-provenance
> **Execution**: focused Go test + current-scope lint
> **Status**: Ready

## Given / When / Then

Given a running continuous practice conversation, when the user asks for help in an ordinary message, the service sends it through `sendPracticeMessage` and `practice.session.chat`. No hint action, assistance mode, counter, feature flag or dedicated scenario/module is created.

The runner also executes the backend-practice stale-contract lint and rejects `no tests to run`.
