# Seed Input

- TargetJob: `openapi/fixtures/TargetJobs/getTargetJob.json` ready state with saved `resumeId`, current practice plan, and canonical 2–5 `summary.interviewRounds[]`.
- Resumes: `openapi/fixtures/Resumes/listResumes.json` ready and empty variants.
- Practice: existing/create plan plus start-session fixtures for direct Start regression.
- Auth: authenticated and unauthenticated `getMe` variants.
- Hostile Parse URL: `/parse?targetJobId=<uuid>&section=reports`; `section=reports` must be discarded.
- Expected report handoff: `/reports?targetJobId=<same uuid>` with no report/status/round/section extras.
- Viewports: 1440x900 and 390x844 with fixed locale/time, loaded fonts and disabled motion for entry parity.
