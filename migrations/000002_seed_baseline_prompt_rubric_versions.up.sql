-- F3 prompt-rubric-registry/001-baseline phase 4.4 seed migration.
-- Writes the 11 baseline feature_keys canonical multi coordinate into
-- prompt_versions and rubric_versions with template_hash matching the
-- on-disk config/prompts/<feature_key>/v0.1.0.{yaml,md} files.
-- Idempotent via ON CONFLICT DO NOTHING.

BEGIN;

INSERT INTO prompt_versions (id, feature_key, version, language, template_hash, template_body, is_active, created_at) VALUES
  ('9a63e69d-c434-5114-88c7-3a9060e2f06d', 'debrief.generate', 'v0.1.0', 'multi', '031dd8e6c4095a829dac6a747b0c8618a582fc0802480ed9743c1fea82b38c22', $body$You are a post-interview coach helping the candidate analyze a real interview.
Respond in the language indicated by `{{language}}` (default English).

Target role: {{targetTitle}}
Target summary: {{targetSummary}}
Recorded interview questions: {{questions}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): Completed real-interview debrief analysis.
- `$.questions` (required, array): Analyzed debrief questions preserving input order.
- `$.questions[]` (required, object): One analyzed interview question.
- `$.questions[].questionText` (required, string): Original or cleaned interview question text.
- `$.questions[].myAnswerSummary` (required, string): Candidate answer summary.
- `$.questions[].aiAnalysis` (required, string): Concise coaching analysis.
- `$.questions[].interviewerReaction` (optional, string): Interviewer reaction when supplied by the candidate.
- `$.riskItems` (required, array): Risks or follow-up concerns from the debrief.
- `$.riskItems[]` (required, object): One risk item.
- `$.riskItems[].label` (required, string): Risk label.
- `$.riskItems[].severity` (required, string enum(low, medium, high)): Risk severity.

Example complete JSON output:
```json
{
  "questions": [
    {
      "questionText": "How did you handle backpressure in the migration?",
      "myAnswerSummary": "Explained queue sizing and retry policy.",
      "aiAnalysis": "Good direction, but add numbers and rollback detail.",
      "interviewerReaction": "Asked for concrete failure metrics."
    }
  ],
  "riskItems": [
    {
      "label": "Thin rollback detail",
      "severity": "medium"
    }
  ]
}
```
<!-- output-schema-contract:end -->

Do not return timeline, lessons, follow_up_actions, nextRoundChecklist, or a
thank-you draft. Do not invent events the candidate did not describe.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('c496cd93-fe74-5ebd-8c3e-b972f63729f2', 'debrief.suggest_questions', 'v0.1.0', 'multi', '1418dfe290462412d642ff4ab6cf84fd9e502b0bfb593bcb0561582a0778ff5e', $body$You generate likely post-interview debrief questions from sanitized preparation context. Respond in the language indicated by `{{language}}` (default English).

Target role: {{role_title}}
Job summary: {{job_summary}}
Resume highlights: {{resume_highlights}}
Mock interview signals: {{mock_report_summary}}
Requested count: {{count}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): Likely post-interview debrief questions.
- `$.suggestions` (required, array): Suggested questions the candidate can answer from memory.
- `$.suggestions[]` (required, object): One suggested debrief question.
- `$.suggestions[].questionText` (required, string): Likely interview question.
- `$.suggestions[].whyLikelyAsked` (required, string): Why this question is likely or useful.
- `$.suggestions[].source` (required, string enum(jd, resume, mock_report, manual)): Context source that motivated the question.
- `$.suggestions[].stage` (optional, string): Optional interview stage or topic grouping.

Example complete JSON output:
```json
{
  "suggestions": [
    {
      "questionText": "Tell me about a time you improved reliability in a distributed system.",
      "whyLikelyAsked": "The JD emphasizes distributed systems and ownership of reliability.",
      "source": "jd",
      "stage": "onsite"
    }
  ]
}
```
<!-- output-schema-contract:end -->

Prefer concise questions the candidate can answer from memory. Do not include
raw resume or report prose beyond the generated question.
$body$, TRUE, '2026-05-16T00:00:00Z'),
  ('9bda6ff0-9fa2-5b18-8e98-243592fa1bf9', 'practice.session.first_question', 'v0.1.0', 'multi', '5275bb911bbf51ae57e04509cd64282aaed6ca75a8e4ed3e7fa660373f6565b8', $body$You are an experienced interviewer running a mock interview based on the
candidate's target job. Generate the first question for the session, anchored
in the role and the rubric the session will be scored against. Respond in the
language indicated by `{{language}}` (default English).

Role: {{role_title}} ({{seniority}})
Top required skills: {{top_skills}}
Rubric dimensions: {{rubric_dimensions}}
Practice goal: {{practice_goal}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): First mock-interview question generated from target job context.
- `$.questionText` (required, string): Question text shown to the candidate.
- `$.questionIntent` (required, string): Short intent label for why this question is asked.
- `$.focusDimension` (optional, string): Optional rubric dimension the question is designed to probe.
- `$.expectedSignals` (optional, array): Optional expected answer signals for later evaluator context.
- `$.expectedSignals[]` (required, string): One expected signal.
- `$.timeBudgetSeconds` (optional, integer): Optional suggested answer time budget in seconds.

Example complete JSON output:
```json
{
  "questionText": "Tell me about a time you improved reliability in a distributed system.",
  "questionIntent": "Probe ownership, tradeoffs, and evidence quality.",
  "focusDimension": "System design",
  "expectedSignals": [
    "Names constraints, tradeoffs, measured impact, and rollback plan."
  ],
  "timeBudgetSeconds": 180
}
```
<!-- output-schema-contract:end -->
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('ba817c2b-7771-5ab3-b44b-89881f703ed5', 'practice.session.follow_up', 'v0.1.0', 'multi', 'a486b559a2094bef15e430b906887cf3e74582f67a3ec252e0d9a71f9069b7a0', $body$You are continuing a mock interview. Based on the candidate's most recent
answer, propose exactly one follow-up question that probes deeper, addresses a
gap, or pivots to an uncovered rubric dimension. Respond in the language
indicated by `{{language}}` (default English).

Last question: {{last_question}}
Last answer: {{last_answer}}
Coverage so far: {{covered_dimensions}}
Remaining rubric dimensions: {{remaining_dimensions}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): Follow-up mock-interview question generated from the latest answer.
- `$.questionText` (required, string): Follow-up question text shown to the candidate.
- `$.questionIntent` (required, string): Short intent label for the follow-up question.
- `$.branchDimension` (optional, string): Optional rubric dimension or branch reason for the follow-up.
- `$.confidence` (optional, number): Optional confidence score for the follow-up choice.

Example complete JSON output:
```json
{
  "questionText": "Tell me about a time you improved reliability in a distributed system.",
  "questionIntent": "Probe ownership, tradeoffs, and evidence quality.",
  "branchDimension": "System design tradeoffs",
  "confidence": 0.82
}
```
<!-- output-schema-contract:end -->

Do not return more than one question.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('84078349-c25c-5fd6-84a4-09825145e468', 'practice.turn.lightweight_observe', 'v0.1.0', 'multi', '58e27c96325cfaecfa9a112864e4c8a34a799a2166e1b1df2e8979770c032059', $body$You are a real-time interview observer. The candidate is partway through an
answer; produce one short, neutral cue that the UI can surface without
interrupting the flow. Be concise (under 24 words) and never lead the
candidate to a specific answer. Respond in `{{language}}` (default English).

Question: {{question}}
Partial answer: {{partial_answer}}
Elapsed seconds: {{elapsed_seconds}}

Also produce `answerSummary`: one concise third-person summary of the
candidate's answer for downstream report assessment. Summarize concepts and
evidence; do not quote the answer verbatim.

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): Lightweight real-time interview observation cue.
- `$.cue` (required, string): Short neutral cue that may be surfaced to the candidate.
- `$.answerSummary` (optional, string): Optional concise third-person answer summary for report assessment.
- `$.severity` (optional, string enum(info, nudge, alert)): Optional cue urgency for downstream evaluation or UI treatment.
- `$.dimensionHint` (optional, string): Optional rubric dimension hinted by the cue.

Example complete JSON output:
```json
{
  "cue": "Clarify the tradeoff before moving to implementation details.",
  "answerSummary": "Candidate described implementation details but had not yet clarified the main tradeoff.",
  "severity": "nudge",
  "dimensionHint": "System design"
}
```
<!-- output-schema-contract:end -->
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('efa7e693-993b-5e72-8da7-fde07f80bc60', 'report.generate', 'v0.1.0', 'multi', '2e5fa63ccd84ff440d1aac65416977ca625e38615367a175beab13e90b0510eb', $body$You are an interview report writer. Produce a structured assessment from
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
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('4ad44434-3f9c-55d7-bea1-eacb554e10f6', 'report.question_assessment', 'v0.1.0', 'multi', 'd912a229470c84e7b8fcd46edec5b18c156913914e00632ffe8079c9ad5fb3e4', $body$You are an interview rubric judge. Score one answered turn from sanitized
session metadata and turn summaries; do not invent dimensions outside the
rubric. Respond in the language indicated by `{{language}}` (default English).

Session metadata: {{session_metadata}}
Turn summaries: {{turn_summaries}}
Question context: {{question_context}}
Answer summary: {{answer_summary}}
Rubric dimensions and score levels: {{rubric}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): Per-question rubric assessment.
- `$.dimension_results` (required, object): Map keyed by rubric dimension name. Each value contains score_level, status, confidence, and optional score.
- `$.overall_status` (required, string enum(needs_work, meets_bar, strong)): Overall dimension status.
- `$.confidence` (required, number): Overall confidence score.
- `$.strengths` (required, array): Strength observations for this answer.
- `$.strengths[]` (required, string): One strength.
- `$.gaps` (required, array): Gap observations for this answer.
- `$.gaps[]` (required, string): One gap.
- `$.recommended_framework` (required, string): Suggested answer framework for retry.
- `$.review_status` (required, string enum(open, queued_for_retry, resolved)): Question review status.

Example complete JSON output:
```json
{
  "dimension_results": {
    "system_design": {
      "score_level": "meets_bar",
      "status": "meets_bar",
      "confidence": 0.82,
      "score": 4.2
    }
  },
  "overall_status": "meets_bar",
  "confidence": 0.82,
  "strengths": [
    "Quantified backend reliability impact."
  ],
  "gaps": [
    "Needs deeper rollback and failure-mode analysis."
  ],
  "recommended_framework": "Use STAR with explicit constraints, tradeoffs, and measured outcome.",
  "review_status": "open"
}
```
<!-- output-schema-contract:end -->

Map `score_level` weak or developing to `status` `needs_work`, proficient to
`meets_bar`, and strong to `strong`. Set `review_status` to `open` unless the
turn clearly needs another retry, in which case use `queued_for_retry`; use
`resolved` only when the answer already closes a prior retry gap. Use
summarized observations only; do not request raw interview text or direct
quotes.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('410f16c3-3ea9-5327-a87a-027f039368b3', 'resume.parse', 'v0.1.0', 'multi', '8ee64c7dc89e0b8907c116766efeda4a20e3756fc165d92590cbe32396b362ed', $body$You are a resume parser. Extract structured experience from the supplied
resume text. Respond in the language indicated by `{{language}}` (default
English) regardless of the resume's source language.

Resume text:

{{resume_text}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): Structured resume summary parsed from supplied resume text.
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
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('2b8a3995-f76c-555d-bb47-1b85e7146613', 'resume.tailor.bullet_suggestions', 'v0.1.0', 'multi', '3214ba7fcaf6907fb74c4d0473dcc32ecee67a35805e751681b5b08f393a21e5', $body$You are a resume editor producing impact-driven bullet suggestions tailored to
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
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('9f94eeb0-f95a-5a00-b5a0-932330b5cf63', 'resume.tailor.gap_review', 'v0.1.0', 'multi', '9bd9e789fdb8447ca282b9e05834294bda4bff045d99d8b81896490ea6dea99b', $body$You are a resume coach reviewing alignment between a candidate's resume and a
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
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('3e4dae23-7bc3-56cb-868e-72e7c8a6c331', 'target.import.parse', 'v0.1.0', 'multi', 'd328a40d0ae3fbd39907e484c0dc230d30c8d76eb88b85c3c2699f7d739eb091', $body$You are an expert technical interviewer assistant. Extract the interview-ready
target job model from the following job description. Respond strictly in the
language identified by the `{{language}}` variable; if `{{language}}` is empty
or unknown, respond in English.

JD source URL (empty for non-URL imports): `{{jd_source_url}}`
JD raw text:

{{jd_text}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.
Produce a complete JSON value, not JSON Schema or an OpenAPI schema.

Output shape:
- `$` (required, object): Structured target job model extracted from a job description.
- `$.coreThemes` (required, array): Concise technical or domain themes from the role.
- `$.coreThemes[]` (required, string): One role theme.
- `$.interviewHypotheses` (required, array): Likely interview focus hypotheses grounded in the JD.
- `$.interviewHypotheses[]` (required, string): One interview hypothesis.
- `$.strengths` (required, array): Candidate-fit strengths that the JD would reward.
- `$.strengths[]` (required, string): One strength signal.
- `$.gaps` (required, array): Preparation gaps implied by the JD.
- `$.gaps[]` (required, string): One gap or preparation area.
- `$.riskSignals` (required, array): Risk or ambiguity signals in the JD.
- `$.riskSignals[]` (required, string): One risk signal.
- `$.requirements` (required, array): Interview-ready requirements used to build target job requirement records.
- `$.requirements[]` (required, object): One parsed requirement.
- `$.requirements[].kind` (required, string enum(must_have, nice_to_have, hidden_signal, interview_focus)): Requirement category.
- `$.requirements[].label` (required, string): Short requirement phrase.
- `$.requirements[].description` (optional, string): Optional explanation of why the requirement matters.
- `$.requirements[].evidenceLevel` (optional, string enum(explicit, inferred)): Whether the requirement was explicit or inferred.

Example complete JSON output:
```json
{
  "coreThemes": [
    "Distributed systems reliability"
  ],
  "interviewHypotheses": [
    "Interviewer may probe cache invalidation and rollback decisions."
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
$body$, TRUE, '2026-05-09T11:30:00Z')
ON CONFLICT (feature_key, version, language) DO NOTHING;

INSERT INTO rubric_versions (id, feature_key, version, language, schema_json, is_active, created_at) VALUES
  ('8921c937-9ab1-52cc-9502-718bfb0a5461', 'debrief.generate', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Generated question analyses preserve the candidate's recorded interview beats.", "name": "debrief_recall_completeness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Question analyses name the underlying interview signal rather than restating the answer.", "name": "debrief_lesson_specificity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Risk items are concrete, severity-calibrated, and actionable for the next preparation round.", "name": "debrief_action_quality", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "debrief.generate", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('a8d1acd0-39d8-55ec-8ad4-cf4c8ef39948', 'debrief.suggest_questions', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Suggested questions cover the likely interview stages and the user's available context.", "name": "debrief_recall_completeness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Rationales explain the signal behind each suggested question.", "name": "debrief_lesson_specificity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Output language matches the requested locale and uses stable source labels.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.25}], "feature_key": "debrief.suggest_questions", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-16T00:00:00Z'),
  ('0c570c46-b7b6-5ff8-9c4e-01591e59d3a0', 'practice.session.first_question', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Question or follow-up probes deeply into the candidate's reasoning rather than staying on the surface.", "name": "practice_depth", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Across the session, rubric dimensions are exercised without leaving large blind spots.", "name": "practice_dimension_coverage", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "practice.session.first_question", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('673b43cf-3cda-5eed-8fa6-2872157da379', 'practice.session.follow_up', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Follow-up question targets the candidate's actual gap or signal rather than recycling the prior turn.", "name": "followup_relevance", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Question or follow-up probes deeply into the candidate's reasoning rather than staying on the surface.", "name": "practice_depth", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "practice.session.follow_up", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('b827bc4a-b976-53f1-a79d-f0fe9087905f', 'practice.turn.lightweight_observe', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "The cue or generated artifact surfaces a high-signal moment rather than commentary.", "name": "practice_signal_strength", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Output is unambiguous and parseable without re-reading.", "name": "practice_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.2}], "feature_key": "practice.turn.lightweight_observe", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('b3fedd43-277e-59f3-be48-382c7e046891', 'report.generate', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Conclusions are anchored in sanitized turn summaries or recorded artifacts.", "name": "report_evidence", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Report cites concrete evidence and avoids hedging or generic language.", "name": "report_specificity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.25}, {"description": "Recommended next actions are specific, owned, and time-bounded.", "name": "report_action_quality", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.25}, {"description": "Final scores align with rubric levels and with the qualitative reasoning.", "name": "report_calibration", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.15}], "feature_key": "report.generate", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('12ab4687-4190-5792-b103-01ef7601b803', 'report.question_assessment', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Conclusions are anchored in sanitized turn summaries or recorded artifacts.", "name": "report_evidence", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Final scores align with rubric levels and with the qualitative reasoning.", "name": "report_calibration", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Per-dimension scores stay calibrated against the rubric levels and the cited evidence.", "name": "score_outlier", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "report.question_assessment", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('fdcb8879-51c7-538b-b51f-ad6c4c892ea7', 'resume.parse', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Resume content aligns with the target JD's required and preferred signals.", "name": "resume_match", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Bullets and section structure read cleanly without filler or jargon.", "name": "resume_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Edits preserve the candidate's stated facts and avoid embellishment.", "name": "resume_truthfulness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "resume.parse", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('b774fe9d-7d92-5f2f-8399-810d19ed8446', 'resume.tailor.bullet_suggestions', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Bullets read as outcomes rather than activities and quantify the impact when possible.", "name": "resume_impact", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Edits preserve the candidate's stated facts and avoid embellishment.", "name": "resume_truthfulness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Bullets and section structure read cleanly without filler or jargon.", "name": "resume_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "resume.tailor.bullet_suggestions", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('39db2d84-c578-5757-a790-b91a9b43c88d', 'resume.tailor.gap_review', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Resume content aligns with the target JD's required and preferred signals.", "name": "resume_match", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Edits preserve the candidate's stated facts and avoid embellishment.", "name": "resume_truthfulness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Bullets and section structure read cleanly without filler or jargon.", "name": "resume_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "resume.tailor.gap_review", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('dc1358dd-ad2a-58e4-a7bb-16b8fb730780', 'target.import.parse', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "All major JD fields (role, seniority, skills, responsibilities) are captured.", "name": "target_extraction_completeness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Captured fields reflect the JD without invention or paraphrase drift.", "name": "target_field_accuracy", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.2}], "feature_key": "target.import.parse", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z')
ON CONFLICT (feature_key, version, language) DO NOTHING;

COMMIT;
