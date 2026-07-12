# E2E.P0.046 message failure and recovery

Message submission uses `clientMessageId` for replay. Provider failure and empty resume context do not commit a fake assistant reply; empty context returns `VALIDATION_FAILED` with zero AI calls while the user reservation remains retryable. Exact replay and mismatch behavior remain deterministic.
