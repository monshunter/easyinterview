# Expected Outcome

- `startPracticeSession` returns `201 + PracticeSession` with `status=running`, `turnCount=1`, and `currentTurn.turnIndex=1`.
- `getPracticeSession` returns the same running session with the same current turn.
- Store snapshot contains one `practice_turns` equivalent row and one `practice_session_events(seq_no=1, event_type=session_started)` equivalent row.
- Store snapshot contains one `practice.session.started` outbox payload with `planId`, `sessionId`, `targetJobId`, `goal`, `mode`, and `language`.
- The outbox payload excludes `questionText`, answer text, hint text, prompt body, response body, and provider secrets.
- The fake AIClient observes that `Complete` ran outside the store transaction window.
