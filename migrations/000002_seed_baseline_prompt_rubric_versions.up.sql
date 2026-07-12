-- F3 canonical six-feature prompt/rubric seed.
-- Generated mechanically from config/prompts and config/rubrics.
BEGIN;

INSERT INTO prompt_versions (id, feature_key, version, language, template_hash, template_body, is_active, created_at) VALUES
  ('40dd127e-afd8-5dcc-9697-3b0025b99702', 'practice.session.chat', 'v0.1.0', 'multi', 'e57c8e7b91772166af32f6cbc50332e9c7ae842ccc9fff8832895d44157b8d09', $body$You are conducting a realistic mock interview as a natural text conversation.
Use the confirmed target job, resume, interview round, practice goal, competency
focus, and ordered conversation history below. Respond in `{{language}}`.

Target job context: {{target_job_context}}
Resume context: {{resume_context}}
Interview round: {{interview_round}}
Practice goal: {{practice_goal}}
Competency focus: {{focus_competencies}}
Conversation history: {{conversation_history}}

Continue naturally with one useful interviewer message. Do not expose a question
number, total question count, question category, turn state, hint mode, scoring,
or internal reasoning. If the user asks for help, respond within the same normal
conversation instead of emitting a special hint action.

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): One ordinary assistant message in a continuous practice conversation.
- `$.messageText` (required, string): Assistant message shown to the user.

Example complete JSON output:
```json
{
  "messageText": "example messageText"
}
```
<!-- output-schema-contract:end -->
$body$, TRUE, '2026-07-12T08:00:00Z'),
  ('aedcb7d9-a56c-5f34-9039-8cdc65828f53', 'report.generate', 'v0.1.0', 'multi', '2e5fa63ccd84ff440d1aac65416977ca625e38615367a175beab13e90b0510eb', $body$You are an interview report writer. Produce a structured assessment from
sanitized session metadata and turn summaries, anchored in the rubric. Respond
in the language indicated by `{{language}}` (default English).

Session metadata: {{session_metadata}}
Turn summaries: {{turn_summaries}}
Rubric dimensions and score levels: {{rubric_dimensions}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): Structured interview feedback report content.
- `$.summary` (required, string): Concise overall report summary.
- `$.dimension_scores` (required, array): Per-rubric dimension scores.
- `$.dimension_scores[]` (required, object): One dimension score.
- `$.dimension_scores[].name` (required, string): Rubric dimension name.
- `$.dimension_scores[].score` (required, number): Numeric score for the dimension.
- `$.dimension_scores[].reasoning` (required, string): Short reasoning for the score.
- `$.dimension_scores[].supporting_observations` (required, array): Summarized observations supporting the score.
- `$.dimension_scores[].supporting_observations[]` (required, string): One supporting observation.
- `$.highlights` (required, array): Positive evidence items.
- `$.highlights[]` (required, object): One evidence item.
- `$.highlights[].dimension` (required, string): Related rubric dimension.
- `$.highlights[].evidence` (required, string): Summarized evidence.
- `$.highlights[].confidence` (required, number): Confidence score for this evidence.
- `$.issues` (required, array): Improvement or risk evidence items.
- `$.issues[]` (required, object): One evidence item.
- `$.issues[].dimension` (required, string): Related rubric dimension.
- `$.issues[].evidence` (required, string): Summarized evidence.
- `$.issues[].confidence` (required, number): Confidence score for this evidence.
- `$.next_actions` (required, array): Recommended next actions for the candidate.
- `$.next_actions[]` (required, object): One next action.
- `$.next_actions[].type` (required, string): Action type.
- `$.next_actions[].label` (required, string): Action label shown to the candidate.
- `$.retry_focus_turn_ids` (required, array): Turn IDs recommended for retry focus.
- `$.retry_focus_turn_ids[]` (required, string): Practice turn ID.

Example complete JSON output:
```json
{
  "summary": "The candidate gave a structured answer with clear tradeoffs but should quantify impact.",
  "dimension_scores": [
    {
      "name": "System design",
      "score": 4.2,
      "reasoning": "Clear architecture tradeoffs, but limited quantified impact.",
      "supporting_observations": [
        "Used concrete operational examples from the session."
      ]
    }
  ],
  "highlights": [
    {
      "dimension": "System design",
      "evidence": "Explained queue backpressure and deployment tradeoffs.",
      "confidence": 0.82
    }
  ],
  "issues": [
    {
      "dimension": "Risk handling",
      "evidence": "Rollback plan was mentioned but not made concrete.",
      "confidence": 0.82
    }
  ],
  "next_actions": [
    {
      "type": "retry_round",
      "label": "Replay the system design follow-up"
    }
  ],
  "retry_focus_turn_ids": [
    "turn-3"
  ]
}
```
<!-- output-schema-contract:end -->

Use summarized observations only; do not request raw interview text or direct quotes.
$body$, TRUE, '2026-07-12T08:00:00Z'),
  ('a6924021-93db-552f-a853-08987d205429', 'resume.parse', 'v0.1.0', 'multi', '71f57bc206d0e983ff918d776d6ebb7c2ece1de1959a4238de91fdd5c612ed5d', $body$You are a resume parser. Extract structured experience from the supplied
resume text. Respond in the language indicated by `{{language}}` (default
English) regardless of the resume's source language.

Resume text:

{{resume_text}}

Generate `displayName` as a concise resume name for the UI. Use the candidate
name plus headline, role, or strongest technical positioning when available.
Never use "uploaded resume", "pasted resume", the file name, or a raw first-line
copy as `displayName`.

Generate `markdownText` as the complete resume body converted to Markdown for
UI rendering. Preserve the original writing order, section structure, wording,
bullets, and factual content. Do not summarize, rewrite, add, remove, or reorder
resume content; only normalize the representation to Markdown syntax.

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): Structured resume summary parsed from supplied resume text.
- `$.displayName` (required, string): Short meaningful resume name for UI display, derived from candidate name plus headline, role, or strongest technical positioning; never use uploaded/pasted resume, the file name, or a raw first-line copy.
- `$.markdownText` (required, string): Complete resume text converted to Markdown while preserving the source resume's writing order, section structure, wording, bullets, and factual content. Do not summarize, rewrite, add, remove, or reorder resume content.
- `$.basics` (required, object): Basic candidate identity and contact summary.
- `$.basics.name` (optional, string): Candidate name when present.
- `$.basics.headline` (optional, string): Candidate headline or target positioning.
- `$.basics.contact` (optional, string): Sanitized contact summary.
- `$.experiences` (required, array): Work experience entries.
- `$.experiences[]` (required, object): One work experience entry.
- `$.experiences[].company` (optional, string): Company name.
- `$.experiences[].title` (optional, string): Role title.
- `$.experiences[].start` (optional, string): Start date as stated or normalized.
- `$.experiences[].end` (optional, string): End date as stated or normalized.
- `$.experiences[].summary` (optional, string): Brief experience summary.
- `$.experiences[].bullets` (optional, array): Resume bullets for this experience.
- `$.experiences[].bullets[]` (required, string): One bullet.
- `$.projects` (required, array): Project entries.
- `$.projects[]` (required, object): One project entry.
- `$.projects[].name` (optional, string): Project name.
- `$.projects[].summary` (optional, string): Brief project summary.
- `$.projects[].technologies` (optional, array): Technologies or methods used in the project.
- `$.projects[].technologies[]` (required, string): One technology or method.
- `$.projects[].bullets` (optional, array): Project bullets or impact notes.
- `$.projects[].bullets[]` (required, string): One project bullet.
- `$.education` (required, array): Education entries.
- `$.education[]` (required, object): One education entry.
- `$.education[].school` (optional, string): School or institution name.
- `$.education[].degree` (optional, string): Degree or credential.
- `$.education[].field` (optional, string): Field of study when present.
- `$.education[].start` (optional, string): Start date as stated or normalized.
- `$.education[].end` (optional, string): End date as stated or normalized.
- `$.skills` (required, array): Skill keywords.
- `$.skills[]` (required, string): One skill.
- `$.languages` (required, array): Language proficiencies.
- `$.languages[]` (required, string): One language.

Example complete JSON output:
```json
{
  "displayName": "Candidate A - Backend engineer",
  "markdownText": "# Candidate A\n\n## Experience\n- Reduced p95 latency by 32% by redesigning cache invalidation.\n\n## Skills\n- Go",
  "basics": {
    "name": "Candidate A",
    "headline": "Backend engineer focused on distributed systems",
    "contact": "email and phone redacted"
  },
  "experiences": [
    {
      "company": "Example Cloud",
      "title": "Senior Backend Engineer",
      "start": "2021",
      "end": "Present",
      "summary": "Owned high-throughput API reliability and platform migrations.",
      "bullets": [
        "Reduced p95 latency by 32% by redesigning cache invalidation."
      ]
    }
  ],
  "projects": [
    {
      "name": "Interview Prep Platform",
      "summary": "Built evidence-backed interview practice workflows.",
      "technologies": [
        "Go"
      ],
      "bullets": [
        "Reduced p95 latency by 32% by redesigning cache invalidation."
      ]
    }
  ],
  "education": [
    {
      "school": "Example University",
      "degree": "B.S. Computer Science",
      "field": "Computer Science",
      "start": "2021",
      "end": "Present"
    }
  ],
  "skills": [
    "Go"
  ],
  "languages": [
    "English - professional"
  ]
}
```
<!-- output-schema-contract:end -->

Use ISO-style date strings; leave a field empty when the resume does not state it.
$body$, TRUE, '2026-07-12T08:00:00Z'),
  ('662c1078-d571-5464-a1f3-c008500013cb', 'resume.tailor.bullet_suggestions', 'v0.1.0', 'multi', '3214ba7fcaf6907fb74c4d0473dcc32ecee67a35805e751681b5b08f393a21e5', $body$You are a resume editor producing impact-driven bullet suggestions tailored to
a target JD. Each suggestion must keep facts truthful and cite the original
bullet for traceability. Respond in the language indicated by `{{language}}`
(default English).

Original bullet: {{original_bullet}}
Target context: {{jd_context}}
Tone: {{tone}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): Impact-driven resume bullet rewrite suggestions.
- `$.suggestions` (required, array): Canonical bullet suggestions persisted for review.
- `$.suggestions[]` (required, object): One bullet suggestion.
- `$.suggestions[].originalBullet` (required, string): Original source bullet for traceability.
- `$.suggestions[].suggestedBullet` (required, string): Truthful rewritten bullet tailored to the target JD.
- `$.suggestions[].reason` (required, string): Why the suggested bullet is stronger.

Example complete JSON output:
```json
{
  "suggestions": [
    {
      "originalBullet": "Worked on API reliability.",
      "suggestedBullet": "Improved API reliability by reducing incident rate 28% through retry-safe queue processing.",
      "reason": "Adds scope, measurable impact, and target-JD language."
    }
  ]
}
```
<!-- output-schema-contract:end -->

Provide at least three suggestions.
$body$, TRUE, '2026-07-12T08:00:00Z'),
  ('1026574f-5798-5eff-97ed-376de46f0398', 'resume.tailor.gap_review', 'v0.1.0', 'multi', '9bd9e789fdb8447ca282b9e05834294bda4bff045d99d8b81896490ea6dea99b', $body$You are a resume coach reviewing alignment between a candidate's resume and a
target JD. Respond in the language indicated by `{{language}}` (default
English).

Resume summary: {{resume_summary}}
JD summary: {{jd_summary}}
Target seniority: {{target_seniority}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): Resume-to-target-job gap review normalized for tailor run storage.
- `$.matchSummary` (required, object): Canonical match summary consumed by resume tailor parsing.
- `$.matchSummary.strengths` (required, array): Resume strengths to amplify for the target JD.
- `$.matchSummary.strengths[]` (required, string): One strength.
- `$.matchSummary.gaps` (required, array): Resume gaps to address for the target JD.
- `$.matchSummary.gaps[]` (required, string): One gap.

Example complete JSON output:
```json
{
  "matchSummary": {
    "strengths": [
      "Quantified backend reliability impact."
    ],
    "gaps": [
      "Needs deeper rollback and failure-mode analysis."
    ]
  }
}
```
<!-- output-schema-contract:end -->
$body$, TRUE, '2026-07-12T08:00:00Z'),
  ('45833fb3-09e6-541e-bb57-abb93a056493', 'target.import.parse', 'v0.1.0', 'multi', '979db6afdd08218d7593379f9b477952e4eddf743fedbafe1b46bf53144f2a2a', $body$You are an expert technical interviewer assistant. Extract the interview-ready
target job model from the following job description. Respond strictly in the
language identified by the `{{language}}` variable; if `{{language}}` is empty
or unknown, respond in English.
Always extract the canonical job title and company or hiring organization name
when they are present anywhere in the JD text, source URL, page heading, or
metadata-like lines. Do not leave them out when the JD includes them.
For `interviewRounds`, infer a reasonable interview plan from JD evidence,
role seniority, company or industry nature, team or business context,
role scope, hiring-process hints, and common interview practices for
similar roles. Emit 2 to 5 rounds. The number of rounds, round type, duration,
and focus must be specific to that inferred context. Do not emit a fixed
four-round template.
For `requirements`, always include at least one `hidden_signal` item with
`evidenceLevel` set to `inferred`. Hidden signals are unstated but likely
interview concerns inferred from JD omissions, company or industry nature,
business scenario, seniority, hiring-process wording, and risk or ambiguity
signals.

JD source URL (empty for non-URL imports): `{{jd_source_url}}`
JD raw text:

{{jd_text}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): Structured target job model extracted from a job description.
- `$.title` (required, string): Canonical job title or role name extracted from the JD.
- `$.companyName` (required, string): Canonical company or hiring organization name extracted from the JD.
- `$.coreThemes` (required, array): Concise technical or domain themes from the role.
- `$.coreThemes[]` (required, string): One role theme.
- `$.interviewRounds` (required, array): Likely 2 to 5 interview rounds inferred from JD evidence, role seniority, company or industry nature, team or business context, and hiring-process hints, including count, type, display name, duration, and focus.
- `$.interviewRounds[]` (required, object): One inferred interview round.
- `$.interviewRounds[].sequence` (required, integer): One-based round order inferred from the likely interview process.
- `$.interviewRounds[].type` (required, string enum(hr, technical, manager, cross_functional, culture, final, other)): Stable round category inferred from the JD.
- `$.interviewRounds[].name` (required, string): Candidate-facing round name in the response language.
- `$.interviewRounds[].durationMinutes` (required, integer): Estimated interview duration in minutes inferred from the round category and JD context.
- `$.interviewRounds[].focus` (required, string): Concise likely focus for this round, grounded in the JD and the inferred company, industry, seniority, and role context.
- `$.strengths` (required, array): Candidate-fit strengths that the JD would reward.
- `$.strengths[]` (required, string): One strength signal.
- `$.gaps` (required, array): Preparation gaps implied by the JD.
- `$.gaps[]` (required, string): One gap or preparation area.
- `$.riskSignals` (required, array): Risk or ambiguity signals in the JD; also source material for inferred hidden_signal requirements.
- `$.riskSignals[]` (required, string): One risk signal.
- `$.requirements` (required, array): Interview-ready requirements used to build target job requirement records. Must include at least one hidden_signal item with inferred evidence.
- `$.requirements[]` (required, object): One parsed requirement.
- `$.requirements[].kind` (required, string enum(must_have, nice_to_have, hidden_signal, interview_focus)): Requirement category.
- `$.requirements[].label` (required, string): Short requirement phrase.
- `$.requirements[].description` (optional, string): Optional explanation of why the requirement matters.
- `$.requirements[].evidenceLevel` (optional, string enum(explicit, inferred)): Whether the requirement was explicit or inferred.

Example complete JSON output:
```json
{
  "title": "Senior Backend Engineer",
  "companyName": "Acme",
  "coreThemes": [
    "Distributed systems reliability"
  ],
  "interviewRounds": [
    {
      "sequence": 1,
      "type": "technical",
      "name": "Technical system design",
      "durationMinutes": 60,
      "focus": "Probe distributed systems reliability and rollback decisions."
    },
    {
      "sequence": 2,
      "type": "manager",
      "name": "Hiring manager ownership interview",
      "durationMinutes": 45,
      "focus": "Assess ownership scope, incident judgment, and cross-team collaboration."
    }
  ],
  "strengths": [
    "Quantified backend reliability impact."
  ],
  "gaps": [
    "Needs deeper rollback and failure-mode analysis."
  ],
  "riskSignals": [
    "The JD asks for on-call ownership without naming team support."
  ],
  "requirements": [
    {
      "kind": "must_have",
      "label": "Design reliable distributed services",
      "description": "The JD explicitly calls for owning high-availability backend systems.",
      "evidenceLevel": "explicit"
    }
  ]
}
```
<!-- output-schema-contract:end -->

Do not include markdown fences in the JSON output.
$body$, TRUE, '2026-07-12T08:00:00Z')
ON CONFLICT (feature_key, version, language) DO NOTHING;

INSERT INTO rubric_versions (id, feature_key, version, language, schema_json, is_active, created_at) VALUES
  ('2e8f5e6f-7efd-5565-89c6-fa81dac2f6fe', 'practice.session.chat', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "The reply follows the actual conversation and the confirmed job context.", "name": "followup_relevance", "score_levels": [{"description": "Ignores the conversation or asks an unrelated canned prompt.", "label": "weak", "threshold": 0.0}, {"description": "Partly follows context but misses a material signal.", "label": "developing", "threshold": 0.4}, {"description": "Continues the conversation with relevant depth.", "label": "proficient", "threshold": 0.7}, {"description": "Uses the strongest available signal to move the interview forward.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "The reply reads as a natural interviewer message without structural question metadata.", "name": "practice_depth", "score_levels": [{"description": "Exposes internal structure or sounds mechanical.", "label": "weak", "threshold": 0.0}, {"description": "Understandable but noticeably templated.", "label": "developing", "threshold": 0.4}, {"description": "Natural, concise, and easy to answer.", "label": "proficient", "threshold": 0.7}, {"description": "Natural and precisely calibrated to the conversation moment.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "The reply matches the requested language consistently.", "name": "language_consistency", "score_levels": [{"description": "Uses the wrong language.", "label": "weak", "threshold": 0.0}, {"description": "Contains avoidable mixed-language output.", "label": "developing", "threshold": 0.4}, {"description": "Uses the requested language consistently.", "label": "proficient", "threshold": 0.7}, {"description": "Uses fluent, role-appropriate language throughout.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "practice.session.chat", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-07-12T08:00:00Z'),
  ('25a17797-8e5e-5dbe-936a-681398f76c4d', 'report.generate', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Conclusions are anchored in sanitized turn summaries or recorded artifacts.", "name": "report_evidence", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Report cites concrete evidence and avoids hedging or generic language.", "name": "report_specificity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.25}, {"description": "Recommended next actions are specific, owned, and time-bounded.", "name": "report_action_quality", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.25}, {"description": "Final scores align with rubric levels and with the qualitative reasoning.", "name": "report_calibration", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.15}], "feature_key": "report.generate", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-07-12T08:00:00Z'),
  ('d69be0bd-dfac-593b-80c0-71950ee61cba', 'resume.parse', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Resume content aligns with the target JD's required and preferred signals.", "name": "resume_match", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Bullets and section structure read cleanly without filler or jargon.", "name": "resume_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Edits preserve the candidate's stated facts and avoid embellishment.", "name": "resume_truthfulness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "resume.parse", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-07-12T08:00:00Z'),
  ('fb3adfc3-7034-5ee4-bef6-3c488b2baab6', 'resume.tailor.bullet_suggestions', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Bullets read as outcomes rather than activities and quantify the impact when possible.", "name": "resume_impact", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Edits preserve the candidate's stated facts and avoid embellishment.", "name": "resume_truthfulness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Bullets and section structure read cleanly without filler or jargon.", "name": "resume_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "resume.tailor.bullet_suggestions", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-07-12T08:00:00Z'),
  ('d27465e1-667a-50c1-a8cf-0178bee3ed65', 'resume.tailor.gap_review', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Resume content aligns with the target JD's required and preferred signals.", "name": "resume_match", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Edits preserve the candidate's stated facts and avoid embellishment.", "name": "resume_truthfulness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Bullets and section structure read cleanly without filler or jargon.", "name": "resume_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "resume.tailor.gap_review", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-07-12T08:00:00Z'),
  ('755053bd-b633-5909-b5cf-ee57d7dcf28d', 'target.import.parse', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "All major JD fields (role, seniority, skills, responsibilities) are captured.", "name": "target_extraction_completeness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Captured fields reflect the JD without invention or paraphrase drift.", "name": "target_field_accuracy", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.2}], "feature_key": "target.import.parse", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-07-12T08:00:00Z')
ON CONFLICT (feature_key, version, language) DO NOTHING;

COMMIT;
