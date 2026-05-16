# backend/internal/review

`review` owns the backend-review 001 report generation baseline:

- inline report runner for `async_jobs.job_type='report_generate'`
- generation orchestration for `report.generate` and `report.question_assessment`
- readiness tier, retry-focus, and next-action decisions
- user-scoped read models for `GET /api/v1/reports/{reportId}` and `GET /api/v1/targets/{targetJobId}/reports`

The HTTP layer lives in `backend/internal/api/reports`; persistence lives in
`backend/internal/store/review`. The current runner is intentionally inline
with the API runtime decision from backend-review D-13. A future shared async
runner may replace the polling shell, but the service/store contracts here
should remain the handoff boundary.

Report generation never persists raw `question_text`, `answer_text`, `hint_text`,
prompt bodies, response bodies, or provider secrets. Prompt/rubric provenance is
stored in `feedback_reports` and `ai_task_runs`, not metric labels.
