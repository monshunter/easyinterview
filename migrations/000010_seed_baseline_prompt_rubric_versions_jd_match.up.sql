-- prompt-rubric-registry/002-output-schema-contract L2 remediation seed migration.
-- Adds jd_match prompt/rubric baselines that were introduced after the original F3 seed.
-- Idempotent via ON CONFLICT DO NOTHING.

BEGIN;

INSERT INTO prompt_versions (id, feature_key, version, language, template_hash, template_body, is_active, created_at) VALUES
  ('6721ff84-ae51-5bd2-9e4b-66be6412c10c', 'jd_match.recommendation', 'v0.1.0', 'en', '198dc876e70832bdc36d2dcf2c62d731eb6567198a0a192bf3903507ed84e9a1', $body$You are a career-discovery assistant. Generate JD-Match recommendations for the
authenticated candidate based on their structured profile and the internal jobs
pool. Always respond in English.

Candidate profile snapshot:

{{candidate_profile}}

Internal jobs pool (already filtered for relevance and freshness):

{{jobs_pool}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, array): Ranked JD-Match recommendations for a candidate.
- `$[]` (required, object): One JD-Match recommendation.
- `$[].jobMatchId` (required, string): Stable identifier from the internal jobs pool.
- `$[].title` (required, string): Job title.
- `$[].company` (required, string): Company name.
- `$[].location` (required, string): Job location.
- `$[].posted` (required, string): Human-readable freshness label.
- `$[].score` (required, integer): Candidate fit score from 0 to 100.
- `$[].fit` (required, object): Requirement-fit counts.
- `$[].fit.must` (required, integer): Matched must-have count.
- `$[].fit.total` (required, integer): Total must-have count.
- `$[].fit.plus` (required, integer): Matched plus/nice-to-have count.
- `$[].fit.totalPlus` (required, integer): Total plus/nice-to-have count.
- `$[].reasons` (required, array): Concise match reasons.
- `$[].reasons[]` (required, string): One match reason.
- `$[].risks` (required, array): Concise mismatch or stretch risks.
- `$[].risks[]` (required, string): One risk.
- `$[].highlights` (required, array): Skill or scope highlights.
- `$[].highlights[]` (required, string): One highlight.
- `$[].companyTag` (optional, string): Optional company tag.
- `$[].level` (optional, string enum(junior, mid, senior, staff, lead, principal)): Optional seniority level.
- `$[].comp` (optional, string): Optional compensation label.
- `$[].sourceUrl` (optional, string): Optional internal pool source URL.
- `$[].sourceLabel` (optional, string): Optional source label.
- `$[].networkNote` (optional, string): Optional aggregated non-PII network signal.
- `$[].similarInterviewers` (optional, integer): Optional count of similar interviewers.
- `$[].interviewHypotheses` (optional, array): Short interviewer-angle hypotheses.
- `$[].interviewHypotheses[]` (required, string): One hypothesis.

Example complete JSON output:
```json
[
  {
    "jobMatchId": "job-123",
    "title": "Senior Backend Engineer",
    "company": "Example Cloud",
    "location": "Remote US",
    "posted": "posted 2 days ago",
    "score": 86,
    "fit": {
      "must": 4,
      "total": 5,
      "plus": 2,
      "totalPlus": 3
    },
    "reasons": [
      "Recent backend platform work maps directly to the JD."
    ],
    "risks": [
      "Less evidence for frontend-heavy collaboration requirements."
    ],
    "highlights": [
      "Strong ownership of backend reliability work."
    ],
    "companyTag": "Growth-stage SaaS",
    "level": "senior",
    "comp": "$180k-$220k",
    "sourceUrl": "https://jobs.internal.example/job-123",
    "sourceLabel": "internal jobs pool",
    "networkNote": "3 prior interview reports mention similar backend platform scope.",
    "similarInterviewers": 3,
    "interviewHypotheses": [
      "Interviewer may probe cache invalidation and rollback decisions."
    ]
  }
]
```
<!-- output-schema-contract:end -->

Hard rules:

- Use only the supplied internal jobs pool entries; do not invent external
  links or fabricate companies the pool does not contain.
- Do not mention private individuals, raw recruiter contacts, or proprietary
  hiring data. `networkNote` must be aggregated and non-PII.
- Do not include free-form essays outside the JSON array.
- Preserve the language code in `{{language}}`; the JSON keys remain ASCII.
$body$, TRUE, '2026-05-21T00:00:00Z'),
  ('441f6098-4e17-5dbb-ad18-22bd618a9a28', 'jd_match.recommendation', 'v0.1.0', 'multi', 'eacff80357474fb25af1c1ecb6722449f7ce5688c9100b8e33149f6f15372b88', $body$You are a career-discovery assistant. Generate JD-Match recommendations for the
authenticated candidate based on their structured profile and the internal jobs
pool. Respond in the language indicated by `{{language}}` (default Chinese for
JD-Match) regardless of source language.

Candidate profile snapshot:

{{candidate_profile}}

Internal jobs pool (already filtered for relevance and freshness):

{{jobs_pool}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, array): Ranked JD-Match recommendations for a candidate.
- `$[]` (required, object): One JD-Match recommendation.
- `$[].jobMatchId` (required, string): Stable identifier from the internal jobs pool.
- `$[].title` (required, string): Job title.
- `$[].company` (required, string): Company name.
- `$[].location` (required, string): Job location.
- `$[].posted` (required, string): Human-readable freshness label.
- `$[].score` (required, integer): Candidate fit score from 0 to 100.
- `$[].fit` (required, object): Requirement-fit counts.
- `$[].fit.must` (required, integer): Matched must-have count.
- `$[].fit.total` (required, integer): Total must-have count.
- `$[].fit.plus` (required, integer): Matched plus/nice-to-have count.
- `$[].fit.totalPlus` (required, integer): Total plus/nice-to-have count.
- `$[].reasons` (required, array): Concise match reasons.
- `$[].reasons[]` (required, string): One match reason.
- `$[].risks` (required, array): Concise mismatch or stretch risks.
- `$[].risks[]` (required, string): One risk.
- `$[].highlights` (required, array): Skill or scope highlights.
- `$[].highlights[]` (required, string): One highlight.
- `$[].companyTag` (optional, string): Optional company tag.
- `$[].level` (optional, string enum(junior, mid, senior, staff, lead, principal)): Optional seniority level.
- `$[].comp` (optional, string): Optional compensation label.
- `$[].sourceUrl` (optional, string): Optional internal pool source URL.
- `$[].sourceLabel` (optional, string): Optional source label.
- `$[].networkNote` (optional, string): Optional aggregated non-PII network signal.
- `$[].similarInterviewers` (optional, integer): Optional count of similar interviewers.
- `$[].interviewHypotheses` (optional, array): Short interviewer-angle hypotheses.
- `$[].interviewHypotheses[]` (required, string): One hypothesis.

Example complete JSON output:
```json
[
  {
    "jobMatchId": "job-123",
    "title": "Senior Backend Engineer",
    "company": "Example Cloud",
    "location": "Remote US",
    "posted": "posted 2 days ago",
    "score": 86,
    "fit": {
      "must": 4,
      "total": 5,
      "plus": 2,
      "totalPlus": 3
    },
    "reasons": [
      "Recent backend platform work maps directly to the JD."
    ],
    "risks": [
      "Less evidence for frontend-heavy collaboration requirements."
    ],
    "highlights": [
      "Strong ownership of backend reliability work."
    ],
    "companyTag": "Growth-stage SaaS",
    "level": "senior",
    "comp": "$180k-$220k",
    "sourceUrl": "https://jobs.internal.example/job-123",
    "sourceLabel": "internal jobs pool",
    "networkNote": "3 prior interview reports mention similar backend platform scope.",
    "similarInterviewers": 3,
    "interviewHypotheses": [
      "Interviewer may probe cache invalidation and rollback decisions."
    ]
  }
]
```
<!-- output-schema-contract:end -->

Hard rules:

- Use only the supplied internal jobs pool entries; do not invent external
  links or fabricate companies the pool does not contain.
- Do not mention private individuals, raw recruiter contacts, or proprietary
  hiring data. `networkNote` must be aggregated and non-PII.
- Do not include free-form essays outside the JSON array.
- Preserve the language code in `{{language}}`; the JSON keys remain ASCII.
$body$, TRUE, '2026-05-21T00:00:00Z'),
  ('185e4bd1-2196-5314-8968-11a7cb0641cb', 'jd_match.search', 'v0.1.0', 'en', 'dd0904ec1e281a03c04c26c432cc92ad130665a1259e297c571440d520fa7288', $body$You are a job search ranker. Given a natural-language query, the candidate
profile, and the internal jobs pool, return the most relevant items from the
pool. Always respond in English.

User query:

{{query}}

Optional filters supplied by the caller:

{{filters}}

Candidate profile snapshot:

{{candidate_profile}}

Internal jobs pool:

{{jobs_pool}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, array): Ranked JD-Match search results from the internal jobs pool.
- `$[]` (required, object): One matched internal job reference.
- `$[].jobMatchId` (required, string): Stable identifier from the internal jobs pool.
- `$[].reason` (optional, string): Optional concise relevance reason for diagnostics.

Example complete JSON output:
```json
[
  {
    "jobMatchId": "job-123",
    "reason": "Adds measurable impact and ties the bullet to the target JD."
  }
]
```
<!-- output-schema-contract:end -->

Hard rules:

- Rank only items present in the supplied internal jobs pool; never inject
  external recruiting links or platform names.
- Do not echo the user query inside `reasons`, `risks`, or `highlights`.
- If the pool contains no relevant items for the query, return an empty JSON
  array. Do not invent items to satisfy the request.
- Preserve the language code in `{{language}}`; the JSON keys remain ASCII.
$body$, TRUE, '2026-05-21T00:00:00Z'),
  ('beda4d05-f944-5dc6-bb11-af2fdc1598d6', 'jd_match.search', 'v0.1.0', 'multi', 'd790bdf88262653107437bf084f66360496dd94788dcb1c141eefd710e478ca6', $body$You are a job search ranker. Given a natural-language query, the candidate
profile, and the internal jobs pool, return the most relevant items from the
pool. Respond in the language indicated by `{{language}}` (default Chinese for
JD-Match) regardless of source language.

User query:

{{query}}

Optional filters supplied by the caller:

{{filters}}

Candidate profile snapshot:

{{candidate_profile}}

Internal jobs pool:

{{jobs_pool}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, array): Ranked JD-Match search results from the internal jobs pool.
- `$[]` (required, object): One matched internal job reference.
- `$[].jobMatchId` (required, string): Stable identifier from the internal jobs pool.
- `$[].reason` (optional, string): Optional concise relevance reason for diagnostics.

Example complete JSON output:
```json
[
  {
    "jobMatchId": "job-123",
    "reason": "Adds measurable impact and ties the bullet to the target JD."
  }
]
```
<!-- output-schema-contract:end -->

Hard rules:

- Rank only items present in the supplied internal jobs pool; never inject
  external recruiting links or platform names.
- Do not echo the user query inside `reasons`, `risks`, or `highlights`.
- If the pool contains no relevant items for the query, return an empty JSON
  array. Do not invent items to satisfy the request.
- Preserve the language code in `{{language}}`; the JSON keys remain ASCII.
$body$, TRUE, '2026-05-21T00:00:00Z')
ON CONFLICT (feature_key, version, language) DO NOTHING;

INSERT INTO rubric_versions (id, feature_key, version, language, schema_json, is_active, created_at) VALUES
  ('a66671bd-d616-5422-82bf-b1bde709d1c0', 'jd_match.recommendation', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "Recommendations align with the candidate's profile signals and the internal jobs pool entries actually exist.", "name": "relevance_to_profile", "score_levels": [{"description": "Off-topic suggestions, hallucinated entries, or roles unrelated to the candidate's profile.", "label": "weak", "threshold": 0.0}, {"description": "Partial alignment; a reviewer would still flag obvious mismatches in level, location, or function.", "label": "developing", "threshold": 0.4}, {"description": "Most recommendations match the candidate's level / function / location preferences without major gaps.", "label": "proficient", "threshold": 0.7}, {"description": "Recommendations feel curated; reviewers would highlight the calibration and risk callouts.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Risks and stretch markers are concrete, non-judgemental, and actionable for the candidate.", "name": "risk_clarity", "score_levels": [{"description": "Risks are missing, vague, or read as judgemental about the candidate.", "label": "weak", "threshold": 0.0}, {"description": "Some risks named but generic; reviewer would request more specifics.", "label": "developing", "threshold": 0.4}, {"description": "Risks are specific, sourced from the JD, and actionable.", "label": "proficient", "threshold": 0.7}, {"description": "Risks read like a thoughtful peer review; candidate knows exactly how to prepare.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "`interviewHypotheses` and `networkNote` give the candidate a clear next step without inventing PII.", "name": "actionability", "score_levels": [{"description": "Empty hypotheses, or notes that name private individuals or fabricated network data.", "label": "weak", "threshold": 0.0}, {"description": "Some hypotheses exist but feel generic; aggregated network notes are inconsistent.", "label": "developing", "threshold": 0.4}, {"description": "Most items have at least one concrete interview hypothesis grounded in the JD; notes stay aggregated.", "label": "proficient", "threshold": 0.7}, {"description": "Hypotheses read like a senior coach's prep plan; notes are aggregated, non-PII, and actionable.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "jd_match.recommendation", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-21T00:00:00Z'),
  ('47d442b8-940a-5a68-94a9-84df2818affa', 'jd_match.recommendation', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Recommendations align with the candidate's profile signals and the internal jobs pool entries actually exist.", "name": "relevance_to_profile", "score_levels": [{"description": "Off-topic suggestions, hallucinated entries, or roles unrelated to the candidate's profile.", "label": "weak", "threshold": 0.0}, {"description": "Partial alignment; a reviewer would still flag obvious mismatches in level, location, or function.", "label": "developing", "threshold": 0.4}, {"description": "Most recommendations match the candidate's level / function / location preferences without major gaps.", "label": "proficient", "threshold": 0.7}, {"description": "Recommendations feel curated; reviewers would highlight the calibration and risk callouts.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Risks and stretch markers are concrete, non-judgemental, and actionable for the candidate.", "name": "risk_clarity", "score_levels": [{"description": "Risks are missing, vague, or read as judgemental about the candidate.", "label": "weak", "threshold": 0.0}, {"description": "Some risks named but generic; reviewer would request more specifics.", "label": "developing", "threshold": 0.4}, {"description": "Risks are specific, sourced from the JD, and actionable.", "label": "proficient", "threshold": 0.7}, {"description": "Risks read like a thoughtful peer review; candidate knows exactly how to prepare.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "`interviewHypotheses` and `networkNote` give the candidate a clear next step without inventing PII.", "name": "actionability", "score_levels": [{"description": "Empty hypotheses, or notes that name private individuals or fabricated network data.", "label": "weak", "threshold": 0.0}, {"description": "Some hypotheses exist but feel generic; aggregated network notes are inconsistent.", "label": "developing", "threshold": 0.4}, {"description": "Most items have at least one concrete interview hypothesis grounded in the JD; notes stay aggregated.", "label": "proficient", "threshold": 0.7}, {"description": "Hypotheses read like a senior coach's prep plan; notes are aggregated, non-PII, and actionable.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "jd_match.recommendation", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-21T00:00:00Z'),
  ('ea52bb98-7da7-5202-9bd7-31a3512e4ab5', 'jd_match.search', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "Ranked items reflect the user query intent and stay within the supplied internal jobs pool.", "name": "query_alignment", "score_levels": [{"description": "Items unrelated to the query, or invented entries not present in the pool.", "label": "weak", "threshold": 0.0}, {"description": "Partial match; reviewer would re-order most items or remove obvious misfits.", "label": "developing", "threshold": 0.4}, {"description": "Most items match the query intent in level, function, or location.", "label": "proficient", "threshold": 0.7}, {"description": "Ranking matches what a senior recruiter would produce from the same pool.", "label": "strong", "threshold": 0.9}], "weight": 0.5}, {"description": "Results cover the relevant slice of the pool without collapsing onto a single company or seniority.", "name": "diversity", "score_levels": [{"description": "All results from the same company or one seniority level despite the pool offering more.", "label": "weak", "threshold": 0.0}, {"description": "Some diversity but key adjacent matches missing.", "label": "developing", "threshold": 0.4}, {"description": "Results cover the relevant companies and seniority bands.", "label": "proficient", "threshold": 0.7}, {"description": "Reviewer would call the result set a well-curated cross-section.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Output never echoes the user query, leaks PII, or links to external recruiting platforms.", "name": "privacy_compliance", "score_levels": [{"description": "User query repeated verbatim inside reasons, or external recruiter links included.", "label": "weak", "threshold": 0.0}, {"description": "Some compliance gaps; e.g. partial query echo or unaggregated network notes.", "label": "developing", "threshold": 0.4}, {"description": "No query echo, aggregated notes, no external recruiting links.", "label": "proficient", "threshold": 0.7}, {"description": "Output reads as if a privacy reviewer pre-sanitised every item.", "label": "strong", "threshold": 0.9}], "weight": 0.2}], "feature_key": "jd_match.search", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-21T00:00:00Z'),
  ('5af96307-70df-5dde-98d8-01e7b108e09e', 'jd_match.search', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Ranked items reflect the user query intent and stay within the supplied internal jobs pool.", "name": "query_alignment", "score_levels": [{"description": "Items unrelated to the query, or invented entries not present in the pool.", "label": "weak", "threshold": 0.0}, {"description": "Partial match; reviewer would re-order most items or remove obvious misfits.", "label": "developing", "threshold": 0.4}, {"description": "Most items match the query intent in level, function, or location.", "label": "proficient", "threshold": 0.7}, {"description": "Ranking matches what a senior recruiter would produce from the same pool.", "label": "strong", "threshold": 0.9}], "weight": 0.5}, {"description": "Results cover the relevant slice of the pool without collapsing onto a single company or seniority.", "name": "diversity", "score_levels": [{"description": "All results from the same company or one seniority level despite the pool offering more.", "label": "weak", "threshold": 0.0}, {"description": "Some diversity but key adjacent matches missing.", "label": "developing", "threshold": 0.4}, {"description": "Results cover the relevant companies and seniority bands.", "label": "proficient", "threshold": 0.7}, {"description": "Reviewer would call the result set a well-curated cross-section.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Output never echoes the user query, leaks PII, or links to external recruiting platforms.", "name": "privacy_compliance", "score_levels": [{"description": "User query repeated verbatim inside reasons, or external recruiter links included.", "label": "weak", "threshold": 0.0}, {"description": "Some compliance gaps; e.g. partial query echo or unaggregated network notes.", "label": "developing", "threshold": 0.4}, {"description": "No query echo, aggregated notes, no external recruiting links.", "label": "proficient", "threshold": 0.7}, {"description": "Output reads as if a privacy reviewer pre-sanitised every item.", "label": "strong", "threshold": 0.9}], "weight": 0.2}], "feature_key": "jd_match.search", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-21T00:00:00Z')
ON CONFLICT (feature_key, version, language) DO NOTHING;

COMMIT;
