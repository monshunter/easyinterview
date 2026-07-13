# Expected Outcome — E2E.P0.070

- empty retry preserves the frozen current round and stores a non-null empty focus array.
- issue-backed retry stores exact unique `retryFocusDimensionCodes`; readback preserves source and focus.
- start/send semantic focus contains only code, label and issue summaries, never anchors or raw source transcript.
- active registry resolution is prompt/rubric v0.2 with `registry.v1`, and the verified F3 semantic-focus release marker is consumed.
- next uses the immediate frozen canonical successor, successor duration and empty focus.
- Idempotency-Key mismatch runs no second create side effect and leaks no source id.
