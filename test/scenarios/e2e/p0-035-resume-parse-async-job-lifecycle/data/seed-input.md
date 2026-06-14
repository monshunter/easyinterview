# Seed Input

- User: `user-resume-parse`
- Resume asset variants:
  - `upload`: object-store backed file object read through backend-upload objectstore abstraction.
  - `paste`: raw text source stored in `resume_assets.original_text`.
  - retired `guided`: absent from current parse inputs; flatten migration drops
    `guided_answers` on the active schema.
- AI response variants:
  - success JSON with `basics`, `experiences`, `projects`, `education`, `skills`, and `languages`.
  - invalid output that fails schema validation.
  - timeout / fallback-exhausted failure path.
- Runtime dependencies:
  - `job_type=resume_parse`
  - dotted feature key `resume.parse`
  - A3 AIClient metadata for `AITaskRunTaskResumeParse`
  - F3 prompt/rubric/profile registry resolution
