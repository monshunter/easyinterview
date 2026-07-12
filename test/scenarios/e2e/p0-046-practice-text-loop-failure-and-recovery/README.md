# E2E.P0.046 message failure and recovery

Message submission uses `clientMessageId` for replay. Provider failure does not commit a fake assistant reply; exact replay returns the original result and mismatched reuse fails with a typed conflict without duplicating messages.
