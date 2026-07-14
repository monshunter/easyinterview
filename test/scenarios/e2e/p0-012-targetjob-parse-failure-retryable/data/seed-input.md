# Seed Input

- TargetJob: imported through exact `{rawText,targetLanguage,resumeId}` with the text retained only in `target_jobs.raw_jd_text`.
- Error injections: `AI_PROVIDER_TIMEOUT`, `AI_OUTPUT_INVALID`, `AI_PROVIDER_SECRET_MISSING`, F3 unsupported/disabled.
- Privacy sentinels: raw JD body, provider secret, prompt/response wording, `Authorization:`.
