# Expected Outcome — E2E.P0.070

- retry plans preserve the source report's exact persisted round pair.
- next plans select the immediate existing canonical successor and current incomplete round.
- equal-duration and non-contiguous sequences do not change identity semantics.
- readback preserves source id plus `roundId/roundSequence`.
