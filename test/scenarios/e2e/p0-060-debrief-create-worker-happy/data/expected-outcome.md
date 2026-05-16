# Expected Outcome

- `createDebrief` returns 202 with `DebriefWithJob.job.jobType=debrief_generate`
- worker happy path calls `UpdateDebriefCompleted`
- `debrief.completed` payload contains only ids and counts
- `ai_task_runs` contains a successful `debrief_generate` row
