# P0.082 Expected Outcome

## ParseFlow Failed State

- `resume-parse-failed-state` testid renders when polling reports parseStatus=failed
- errorCode mapped to enum:
  - `AI_TIMEOUT_RETRYABLE`
  - `PARSE_TIMEOUT`
- Two CTAs visible: `resume-parse-flow-retry` and `resume-parse-flow-back`

## Retry Path

- Clicking retry restarts polling without invoking a new `registerResume`
- The second poll cycle returns `ready` → ParseFlow transitions to PreviewConfirm

## Cancel-and-Return Path

- Clicking `resume-parse-flow-cancel` returns to stage="input"
- Original input is preserved:
  - paste raw text remains in the textarea
  - guided answers remain in the reducer state
  - pickedFile remains selected

## Timeout Path

- After `maxAttempts` polls without terminal status, snapshot transitions to failed with errorCode="PARSE_TIMEOUT"

## Privacy

- ParseFlow DOM does NOT render `parsedTextSnapshot` / `originalText` strings
- After ready transition, PreviewConfirm DOM renders structured fields but NOT private snapshot strings

## Trigger Log Assertions

- `Test Files +\d+ passed` matches
- Linked test files present in log
- `failed`, `cancel-and-return`, `timeout` strings present in test names
