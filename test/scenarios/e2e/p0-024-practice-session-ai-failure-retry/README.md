# E2E.P0.024 practice opening failure and retry

An AI timeout fails the reserved session start without emitting a success event. Independently, an empty resume context fails closed as `VALIDATION_FAILED` before any AI call or opening message. A valid retry can persist exactly one opening assistant message.
