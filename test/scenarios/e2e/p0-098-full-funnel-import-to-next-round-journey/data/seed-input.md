# Seed Input

- User emails:
  - `full-funnel-journey@example.com`
  - `full-funnel-seed@example.com`
- Resume source:
  - registered through `POST /api/v1/resumes`
  - parsed by the real `resume_parse` runner with deterministic scenario AI
- JD source:
  - imported through `POST /api/v1/targets/import`
  - parsed by the real `target_import` runner with deterministic scenario AI
- Practice flow:
  - baseline practice plan
  - text session with one submitted answer
  - completed session producing a feedback report
  - next-round practice plan sourced from the ready report

No ready rows are inserted directly by the scenario.
