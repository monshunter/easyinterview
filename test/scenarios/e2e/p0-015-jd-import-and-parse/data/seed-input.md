# Seed Input

- Import fixtures:
  - `openapi/fixtures/TargetJobs/importTargetJob.json`: `default`, `paste-primary`, and `validation-blank-raw-text`.
  - `openapi/fixtures/TargetJobs/getTargetJob.json`: default parsed body from which focused tests derive queued, processing, ready, and failed parse states.
  - `openapi/fixtures/TargetJobs/listTargetJobs.json`: Home recent variants.
- Frontend routes: paste-only Home and Parse screen.
- Real backend overlay: `targetJob.realApiMode.test.ts` under `VITE_EI_API_MODE=real`.
- Auth continuation: one process-memory pending intent with an opaque route identifier; missing, expired, and duplicate consume cases fail closed.
- Privacy inputs: pasted JD text and parse preview content must stay out of logs, URL, browser storage, and telemetry surfaces.
- AI boundary: this frontend import scenario does not call prompt registry, AIClient, provider keys, or LLM endpoints.
- Browser viewports: desktop 1440×900 and mobile 390×844 for Home and Parse loading evidence.
