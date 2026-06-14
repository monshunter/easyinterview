# Seed Input

- Browser auth state: signed out.
- Runtime `/me` scenarios:
  - unauthenticated for Home and protected route redirect assertions.
  - hanging `/me` for protected route loading gate assertions.
- Protected frontend routes:
  - `parse`
  - `workspace`
  - `resume_versions`
  - `practice`
  - `generating`
  - `report`
  - `debrief`
  - `profile`
  - `settings`
- Backend API proof points:
  - session policy public/optional/protected classification.
  - target job, upload, resume, practice, profile, report, jobs, and jd-match route mount behind session middleware.
