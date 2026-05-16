# E2E.P0.071 Practice debrief start source question

> **场景 ID**: E2E.P0.071
> **自动化入口**: `cd backend && go test ./cmd/api -run TestE2EP0071PracticeDebriefStartUsesSourceQuestion -count=1`

验证 `startPracticeSession(goal='debrief')` 返回的 `currentTurn.questionText` 来自 source debrief 的第一条问题，且不调用 `practice.session.first_question` AI。
