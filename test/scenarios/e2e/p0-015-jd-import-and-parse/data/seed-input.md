# Seed Input

- Import fixtures:
  - `openapi/fixtures/TargetJobs/importTargetJob.json`: manual text, file, URL success, and invalid source error.
  - `openapi/fixtures/Uploads/createUploadPresign.json`: target job attachment presign success and 4xx error.
  - `openapi/fixtures/TargetJobs/getTargetJob.json`: queued, processing, ready, and failed parse states.
- Frontend routes: Home import modal and Parse screen.
- Real backend overlay: `targetJob.realApiMode.test.ts` under `VITE_EI_API_MODE=real`.
- Privacy inputs: pasted JD text, source URL, upload metadata, and parse preview content must stay out of logs, URL, storage, and telemetry surfaces.
- AI boundary: this frontend import scenario does not call prompt registry, AIClient, provider keys, or LLM endpoints.
