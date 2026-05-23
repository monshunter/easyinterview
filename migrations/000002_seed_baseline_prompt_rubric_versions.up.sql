-- F3 prompt-rubric-registry/001-baseline phase 4.4 seed migration.
-- Writes the 11 baseline feature_keys * 2 language coordinates into
-- prompt_versions and rubric_versions with template_hash matching the
-- on-disk config/prompts/<feature_key>/v0.1.0[.<lang>].{yaml,md} files.
-- Idempotent via ON CONFLICT DO NOTHING.

BEGIN;

INSERT INTO prompt_versions (id, feature_key, version, language, template_hash, template_body, is_active, created_at) VALUES
  ('e1ce8408-4111-52b3-bc43-f3658f842a85', 'debrief.generate', 'v0.1.0', 'en', '3418f03439d6239362d1c13c78ffbdacb6ef62de0c4051b2c652de6bfcdccd2f', $body$You are a post-interview coach helping the candidate analyze a real interview.
Respond in English.

Target role: {{targetTitle}}
Target summary: {{targetSummary}}
Recorded interview questions: {{questions}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

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

Example JSON:
```json
{
  "questions": [
    {
      "questionText": "string",
      "myAnswerSummary": "string",
      "aiAnalysis": "string"
    }
  ],
  "riskItems": [
    {
      "label": "string",
      "severity": "low"
    }
  ]
}
```
<!-- output-schema-contract:end -->

Do not return timeline, lessons, follow_up_actions, nextRoundChecklist, or a
thank-you draft. Do not invent events the candidate did not describe.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('9a63e69d-c434-5114-88c7-3a9060e2f06d', 'debrief.generate', 'v0.1.0', 'multi', '5c0cebc8bfd83de32ae53515a86b17a6aca44f86447b0c74e8051262e78be23f', $body$You are a post-interview coach helping the candidate analyze a real interview.
Respond in the language indicated by `{{language}}` (default English).

Target role: {{targetTitle}}
Target summary: {{targetSummary}}
Recorded interview questions: {{questions}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

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

Example JSON:
```json
{
  "questions": [
    {
      "questionText": "string",
      "myAnswerSummary": "string",
      "aiAnalysis": "string"
    }
  ],
  "riskItems": [
    {
      "label": "string",
      "severity": "low"
    }
  ]
}
```
<!-- output-schema-contract:end -->

Do not return timeline, lessons, follow_up_actions, nextRoundChecklist, or a
thank-you draft. Do not invent events the candidate did not describe.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('3fac0d74-abc1-5d91-983c-1439da5e1d8d', 'debrief.suggest_questions', 'v0.1.0', 'en', '140e8de90e5b2e9b61f50d62871d7a6e6a0f3382ac62399558af373d632ea53f', $body$You generate likely post-interview debrief questions from sanitized preparation context. Respond in English.

Target role: {{role_title}}
Job summary: {{job_summary}}
Resume highlights: {{resume_highlights}}
Mock interview signals: {{mock_report_summary}}
Requested count: {{count}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

Output shape:
- `$` (required, object): Likely post-interview debrief questions.
- `$.suggestions` (required, array): Suggested questions the candidate can answer from memory.
- `$.suggestions[]` (required, object): One suggested debrief question.
- `$.suggestions[].questionText` (required, string): Likely interview question.
- `$.suggestions[].whyLikelyAsked` (required, string): Why this question is likely or useful.
- `$.suggestions[].source` (required, string enum(jd, resume, mock_report, manual)): Context source that motivated the question.
- `$.suggestions[].stage` (optional, string): Optional interview stage or topic grouping.

Example JSON:
```json
{
  "suggestions": [
    {
      "questionText": "string",
      "whyLikelyAsked": "string",
      "source": "jd"
    }
  ]
}
```
<!-- output-schema-contract:end -->

Prefer concise questions the candidate can answer from memory. Do not include
raw resume or report prose beyond the generated question.
$body$, TRUE, '2026-05-16T00:00:00Z'),
  ('c496cd93-fe74-5ebd-8c3e-b972f63729f2', 'debrief.suggest_questions', 'v0.1.0', 'multi', '54c6f98481b4d0e153f9c62683ec1db2695c463b1e5a0b803e9da121f4885216', $body$You generate likely post-interview debrief questions from sanitized preparation context. Respond in the language indicated by `{{language}}` (default English).

Target role: {{role_title}}
Job summary: {{job_summary}}
Resume highlights: {{resume_highlights}}
Mock interview signals: {{mock_report_summary}}
Requested count: {{count}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

Output shape:
- `$` (required, object): Likely post-interview debrief questions.
- `$.suggestions` (required, array): Suggested questions the candidate can answer from memory.
- `$.suggestions[]` (required, object): One suggested debrief question.
- `$.suggestions[].questionText` (required, string): Likely interview question.
- `$.suggestions[].whyLikelyAsked` (required, string): Why this question is likely or useful.
- `$.suggestions[].source` (required, string enum(jd, resume, mock_report, manual)): Context source that motivated the question.
- `$.suggestions[].stage` (optional, string): Optional interview stage or topic grouping.

Example JSON:
```json
{
  "suggestions": [
    {
      "questionText": "string",
      "whyLikelyAsked": "string",
      "source": "jd"
    }
  ]
}
```
<!-- output-schema-contract:end -->

Prefer concise questions the candidate can answer from memory. Do not include
raw resume or report prose beyond the generated question.
$body$, TRUE, '2026-05-16T00:00:00Z'),
  ('07e46d0c-af3a-5e6f-98c6-7facc5989fd8', 'practice.session.first_question', 'v0.1.0', 'en', 'f4111d7d06d7f597ebb6560420591a816970c09814343b8f33173c881790b200', $body$You are an experienced interviewer running a mock interview based on the
candidate's target job. Generate the first question for the session, anchored
in the role and the rubric the session will be scored against. Respond in
English.

Role: {{role_title}} ({{seniority}})
Top required skills: {{top_skills}}
Rubric dimensions: {{rubric_dimensions}}
Practice goal: {{practice_goal}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

Output shape:
- `$` (required, object): First mock-interview question generated from target job context.
- `$.questionText` (required, string): Question text shown to the candidate.
- `$.questionIntent` (required, string): Short intent label for why this question is asked.
- `$.focusDimension` (optional, string): Optional rubric dimension the question is designed to probe.
- `$.expectedSignals` (optional, array): Optional expected answer signals for later evaluator context.
- `$.expectedSignals[]` (required, string): One expected signal.
- `$.timeBudgetSeconds` (optional, integer): Optional suggested answer time budget in seconds.

Example JSON:
```json
{
  "questionText": "string",
  "questionIntent": "string"
}
```
<!-- output-schema-contract:end -->
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('9bda6ff0-9fa2-5b18-8e98-243592fa1bf9', 'practice.session.first_question', 'v0.1.0', 'multi', '66d3f665bf082c330bd56f1b4330527aa46fe9b8629cfc9fd2b56e5c3bd3731b', $body$You are an experienced interviewer running a mock interview based on the
candidate's target job. Generate the first question for the session, anchored
in the role and the rubric the session will be scored against. Respond in the
language indicated by `{{language}}` (default English).

Role: {{role_title}} ({{seniority}})
Top required skills: {{top_skills}}
Rubric dimensions: {{rubric_dimensions}}
Practice goal: {{practice_goal}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

Output shape:
- `$` (required, object): First mock-interview question generated from target job context.
- `$.questionText` (required, string): Question text shown to the candidate.
- `$.questionIntent` (required, string): Short intent label for why this question is asked.
- `$.focusDimension` (optional, string): Optional rubric dimension the question is designed to probe.
- `$.expectedSignals` (optional, array): Optional expected answer signals for later evaluator context.
- `$.expectedSignals[]` (required, string): One expected signal.
- `$.timeBudgetSeconds` (optional, integer): Optional suggested answer time budget in seconds.

Example JSON:
```json
{
  "questionText": "string",
  "questionIntent": "string"
}
```
<!-- output-schema-contract:end -->
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('76a70941-aa36-56a8-b941-e7f36ee0e853', 'practice.session.follow_up', 'v0.1.0', 'en', '083563b6204c122893243f3b2e064642c3d8525aa6ef8e6c13c6e0183c1c626c', $body$You are continuing a mock interview. Based on the candidate's most recent
answer, propose exactly one follow-up question that probes deeper, addresses a
gap, or pivots to an uncovered rubric dimension. Respond in English.

Last question: {{last_question}}
Last answer: {{last_answer}}
Coverage so far: {{covered_dimensions}}
Remaining rubric dimensions: {{remaining_dimensions}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

Output shape:
- `$` (required, object): Follow-up mock-interview question generated from the latest answer.
- `$.questionText` (required, string): Follow-up question text shown to the candidate.
- `$.questionIntent` (required, string): Short intent label for the follow-up question.
- `$.branchDimension` (optional, string): Optional rubric dimension or branch reason for the follow-up.
- `$.confidence` (optional, number): Optional confidence score for the follow-up choice.

Example JSON:
```json
{
  "questionText": "string",
  "questionIntent": "string"
}
```
<!-- output-schema-contract:end -->

Do not return more than one question.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('ba817c2b-7771-5ab3-b44b-89881f703ed5', 'practice.session.follow_up', 'v0.1.0', 'multi', 'f393a9ee0a5f827fc4c2f129c9bc1a88a704247cd3b583db50090e3e200bc313', $body$You are continuing a mock interview. Based on the candidate's most recent
answer, propose exactly one follow-up question that probes deeper, addresses a
gap, or pivots to an uncovered rubric dimension. Respond in the language
indicated by `{{language}}` (default English).

Last question: {{last_question}}
Last answer: {{last_answer}}
Coverage so far: {{covered_dimensions}}
Remaining rubric dimensions: {{remaining_dimensions}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

Output shape:
- `$` (required, object): Follow-up mock-interview question generated from the latest answer.
- `$.questionText` (required, string): Follow-up question text shown to the candidate.
- `$.questionIntent` (required, string): Short intent label for the follow-up question.
- `$.branchDimension` (optional, string): Optional rubric dimension or branch reason for the follow-up.
- `$.confidence` (optional, number): Optional confidence score for the follow-up choice.

Example JSON:
```json
{
  "questionText": "string",
  "questionIntent": "string"
}
```
<!-- output-schema-contract:end -->

Do not return more than one question.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('05427dd7-cde4-544e-84c4-6356a514a6a2', 'practice.turn.lightweight_observe', 'v0.1.0', 'en', '10b0453b28d0402d4e82d5d499de20bd2b3cf982b4e0e5db9950f5dab1f59779', $body$You are a real-time interview observer. The candidate is partway through an
answer; produce one short, neutral cue that the UI can surface without
interrupting the flow. Be concise (under 24 words) and never lead the
candidate to a specific answer. Respond in English.

Question: {{question}}
Partial answer: {{partial_answer}}
Elapsed seconds: {{elapsed_seconds}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

Output shape:
- `$` (required, object): Lightweight real-time interview observation cue.
- `$.cue` (required, string): Short neutral cue that may be surfaced to the candidate.
- `$.severity` (optional, string enum(info, nudge, alert)): Optional cue urgency for downstream evaluation or UI treatment.
- `$.dimensionHint` (optional, string): Optional rubric dimension hinted by the cue.

Example JSON:
```json
{
  "cue": "string"
}
```
<!-- output-schema-contract:end -->
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('84078349-c25c-5fd6-84a4-09825145e468', 'practice.turn.lightweight_observe', 'v0.1.0', 'multi', '6ee6598036456841b094f2dcf223077fe11bd06eee629c140492ef8c49767c9b', $body$You are a real-time interview observer. The candidate is partway through an
answer; produce one short, neutral cue that the UI can surface without
interrupting the flow. Be concise (under 24 words) and never lead the
candidate to a specific answer. Respond in `{{language}}` (default English).

Question: {{question}}
Partial answer: {{partial_answer}}
Elapsed seconds: {{elapsed_seconds}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

Output shape:
- `$` (required, object): Lightweight real-time interview observation cue.
- `$.cue` (required, string): Short neutral cue that may be surfaced to the candidate.
- `$.severity` (optional, string enum(info, nudge, alert)): Optional cue urgency for downstream evaluation or UI treatment.
- `$.dimensionHint` (optional, string): Optional rubric dimension hinted by the cue.

Example JSON:
```json
{
  "cue": "string"
}
```
<!-- output-schema-contract:end -->
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('52d20f6c-0d9a-57ca-93fe-5f35b5f90ea3', 'report.generate', 'v0.1.0', 'en', '0cb4f9b4055dfa16df6f7ab1c1f172933eeffada206acfb470e7513951db77f5', $body$You are an interview report writer. Produce a structured assessment from
sanitized session metadata and turn summaries, anchored in the rubric. Respond
in English.

Session metadata: {{session_metadata}}
Turn summaries: {{turn_summaries}}
Rubric dimensions and score levels: {{rubric_dimensions}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

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

Example JSON:
```json
{
  "summary": "string",
  "dimension_scores": [
    {
      "name": "string",
      "score": 0.5,
      "reasoning": "string",
      "supporting_observations": [
        "string"
      ]
    }
  ],
  "highlights": [
    {
      "dimension": "string",
      "evidence": "string",
      "confidence": 0.5
    }
  ],
  "issues": [
    {
      "dimension": "string",
      "evidence": "string",
      "confidence": 0.5
    }
  ],
  "next_actions": [
    {
      "type": "string",
      "label": "string"
    }
  ],
  "retry_focus_turn_ids": [
    "string"
  ]
}
```
<!-- output-schema-contract:end -->

Use summarized observations only; do not request raw interview text or direct quotes.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('efa7e693-993b-5e72-8da7-fde07f80bc60', 'report.generate', 'v0.1.0', 'multi', '22f7a057a30f3f14760f3a0f27d80a153d53d21c2d6d85268a50328be58965d1', $body$You are an interview report writer. Produce a structured assessment from
sanitized session metadata and turn summaries, anchored in the rubric. Respond
in the language indicated by `{{language}}` (default English).

Session metadata: {{session_metadata}}
Turn summaries: {{turn_summaries}}
Rubric dimensions and score levels: {{rubric_dimensions}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

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

Example JSON:
```json
{
  "summary": "string",
  "dimension_scores": [
    {
      "name": "string",
      "score": 0.5,
      "reasoning": "string",
      "supporting_observations": [
        "string"
      ]
    }
  ],
  "highlights": [
    {
      "dimension": "string",
      "evidence": "string",
      "confidence": 0.5
    }
  ],
  "issues": [
    {
      "dimension": "string",
      "evidence": "string",
      "confidence": 0.5
    }
  ],
  "next_actions": [
    {
      "type": "string",
      "label": "string"
    }
  ],
  "retry_focus_turn_ids": [
    "string"
  ]
}
```
<!-- output-schema-contract:end -->

Use summarized observations only; do not request raw interview text or direct quotes.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('bd13ab0f-b8f9-59e7-9e91-ad3955ea70d5', 'report.question_assessment', 'v0.1.0', 'en', 'e7e944a97f5a4b036e444f152848be43f9476f2af43d1a18e913a272e11e740c', $body$You are an interview rubric judge. Score one answered turn from sanitized
session metadata and turn summaries; do not invent dimensions outside the
rubric. Respond in English.

Session metadata: {{session_metadata}}
Turn summaries: {{turn_summaries}}
Question context: {{question_context}}
Answer summary: {{answer_summary}}
Rubric dimensions and score levels: {{rubric}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

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
- `$.review_status` (required, string): Question review status.

Example JSON:
```json
{
  "dimension_results": {},
  "overall_status": "needs_work",
  "confidence": 0.5,
  "strengths": [
    "string"
  ],
  "gaps": [
    "string"
  ],
  "recommended_framework": "string",
  "review_status": "string"
}
```
<!-- output-schema-contract:end -->

Map `score_level` weak or developing to `status` `needs_work`, proficient to
`meets_bar`, and strong to `strong`. Use summarized observations only; do not
request raw interview text or direct quotes.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('4ad44434-3f9c-55d7-bea1-eacb554e10f6', 'report.question_assessment', 'v0.1.0', 'multi', 'b445463c17a8aa51644f402e3f0e1b109f6ce264013b3649543baed94ece02cb', $body$You are an interview rubric judge. Score one answered turn from sanitized
session metadata and turn summaries; do not invent dimensions outside the
rubric. Respond in the language indicated by `{{language}}` (default English).

Session metadata: {{session_metadata}}
Turn summaries: {{turn_summaries}}
Question context: {{question_context}}
Answer summary: {{answer_summary}}
Rubric dimensions and score levels: {{rubric}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

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
- `$.review_status` (required, string): Question review status.

Example JSON:
```json
{
  "dimension_results": {},
  "overall_status": "needs_work",
  "confidence": 0.5,
  "strengths": [
    "string"
  ],
  "gaps": [
    "string"
  ],
  "recommended_framework": "string",
  "review_status": "string"
}
```
<!-- output-schema-contract:end -->

Map `score_level` weak or developing to `status` `needs_work`, proficient to
`meets_bar`, and strong to `strong`. Use summarized observations only; do not
request raw interview text or direct quotes.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('df512597-b914-56e7-a745-7079dfd1af9c', 'resume.parse', 'v0.1.0', 'en', 'dc92ba565349ce27e591dd57726ec2ef210368864bee7fd16a036af9ee23c295', $body$You are a resume parser. Extract structured experience from the supplied
resume text. Respond in English regardless of the resume's source language.

Resume text:

{{resume_text}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

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
- `$.education` (required, array): Education entries.
- `$.education[]` (required, object): One education entry.
- `$.skills` (required, array): Skill keywords.
- `$.skills[]` (required, string): One skill.
- `$.languages` (required, array): Language proficiencies.
- `$.languages[]` (required, string): One language.

Example JSON:
```json
{
  "basics": {},
  "experiences": [
    {}
  ],
  "projects": [
    {}
  ],
  "education": [
    {}
  ],
  "skills": [
    "string"
  ],
  "languages": [
    "string"
  ]
}
```
<!-- output-schema-contract:end -->

Use ISO-style date strings; leave a field empty when the resume does not state it.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('410f16c3-3ea9-5327-a87a-027f039368b3', 'resume.parse', 'v0.1.0', 'multi', 'd27c45e39ca38c52a31a13b60cb37438e221323f6ecbd203f2d7590ea9917452', $body$You are a resume parser. Extract structured experience from the supplied
resume text. Respond in the language indicated by `{{language}}` (default
English) regardless of the resume's source language.

Resume text:

{{resume_text}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

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
- `$.education` (required, array): Education entries.
- `$.education[]` (required, object): One education entry.
- `$.skills` (required, array): Skill keywords.
- `$.skills[]` (required, string): One skill.
- `$.languages` (required, array): Language proficiencies.
- `$.languages[]` (required, string): One language.

Example JSON:
```json
{
  "basics": {},
  "experiences": [
    {}
  ],
  "projects": [
    {}
  ],
  "education": [
    {}
  ],
  "skills": [
    "string"
  ],
  "languages": [
    "string"
  ]
}
```
<!-- output-schema-contract:end -->

Use ISO-style date strings; leave a field empty when the resume does not state it.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('e4518389-f42f-5f88-9554-9ffacec0b7ab', 'resume.tailor.bullet_suggestions', 'v0.1.0', 'en', 'f63fefbe1edcea7ad24a0f8ed8d7495cf291972e260ee233699a2ab593156d6c', $body$You are a resume editor producing impact-driven bullet suggestions tailored to
a target JD. Each suggestion must keep facts truthful and cite the original
bullet for traceability. Respond in English.

Original bullet: {{original_bullet}}
Target context: {{jd_context}}
Tone: {{tone}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

Output shape:
- `$` (required, object): Impact-driven resume bullet rewrite suggestions.
- `$.suggestions` (required, array): Canonical bullet suggestions persisted for review.
- `$.suggestions[]` (required, object): One bullet suggestion.
- `$.suggestions[].originalBullet` (required, string): Original source bullet for traceability.
- `$.suggestions[].suggestedBullet` (required, string): Truthful rewritten bullet tailored to the target JD.
- `$.suggestions[].reason` (required, string): Why the suggested bullet is stronger.

Example JSON:
```json
{
  "suggestions": [
    {
      "originalBullet": "string",
      "suggestedBullet": "string",
      "reason": "string"
    }
  ]
}
```
<!-- output-schema-contract:end -->

Provide at least three suggestions.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('2b8a3995-f76c-555d-bb47-1b85e7146613', 'resume.tailor.bullet_suggestions', 'v0.1.0', 'multi', '95f17553d0d6b35a6f563de16afdb5fb61cc6d10c76fa02f00490d2eabe72afc', $body$You are a resume editor producing impact-driven bullet suggestions tailored to
a target JD. Each suggestion must keep facts truthful and cite the original
bullet for traceability. Respond in the language indicated by `{{language}}`
(default English).

Original bullet: {{original_bullet}}
Target context: {{jd_context}}
Tone: {{tone}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

Output shape:
- `$` (required, object): Impact-driven resume bullet rewrite suggestions.
- `$.suggestions` (required, array): Canonical bullet suggestions persisted for review.
- `$.suggestions[]` (required, object): One bullet suggestion.
- `$.suggestions[].originalBullet` (required, string): Original source bullet for traceability.
- `$.suggestions[].suggestedBullet` (required, string): Truthful rewritten bullet tailored to the target JD.
- `$.suggestions[].reason` (required, string): Why the suggested bullet is stronger.

Example JSON:
```json
{
  "suggestions": [
    {
      "originalBullet": "string",
      "suggestedBullet": "string",
      "reason": "string"
    }
  ]
}
```
<!-- output-schema-contract:end -->

Provide at least three suggestions.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('c59b2fa3-bb08-5dda-8ed6-591a445df1f2', 'resume.tailor.gap_review', 'v0.1.0', 'en', 'bfa219f5349915dbec7e578bf0258e6a5f28bbe247da4fe419706698a64c462c', $body$You are a resume coach reviewing alignment between a candidate's resume and a
target JD. Respond in English.

Resume summary: {{resume_summary}}
JD summary: {{jd_summary}}
Target seniority: {{target_seniority}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

Output shape:
- `$` (required, object): Resume-to-target-job gap review normalized for tailor run storage.
- `$.matchSummary` (required, object): Canonical match summary consumed by resume tailor parsing.
- `$.matchSummary.strengths` (required, array): Resume strengths to amplify for the target JD.
- `$.matchSummary.strengths[]` (required, string): One strength.
- `$.matchSummary.gaps` (required, array): Resume gaps to address for the target JD.
- `$.matchSummary.gaps[]` (required, string): One gap.

Example JSON:
```json
{
  "matchSummary": {
    "strengths": [
      "string"
    ],
    "gaps": [
      "string"
    ]
  }
}
```
<!-- output-schema-contract:end -->
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('9f94eeb0-f95a-5a00-b5a0-932330b5cf63', 'resume.tailor.gap_review', 'v0.1.0', 'multi', 'e24c6ddb575f3c33a6ec57596dce97614edc8eec5c59f6c3aee8748cb9203333', $body$You are a resume coach reviewing alignment between a candidate's resume and a
target JD. Respond in the language indicated by `{{language}}` (default
English).

Resume summary: {{resume_summary}}
JD summary: {{jd_summary}}
Target seniority: {{target_seniority}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

Output shape:
- `$` (required, object): Resume-to-target-job gap review normalized for tailor run storage.
- `$.matchSummary` (required, object): Canonical match summary consumed by resume tailor parsing.
- `$.matchSummary.strengths` (required, array): Resume strengths to amplify for the target JD.
- `$.matchSummary.strengths[]` (required, string): One strength.
- `$.matchSummary.gaps` (required, array): Resume gaps to address for the target JD.
- `$.matchSummary.gaps[]` (required, string): One gap.

Example JSON:
```json
{
  "matchSummary": {
    "strengths": [
      "string"
    ],
    "gaps": [
      "string"
    ]
  }
}
```
<!-- output-schema-contract:end -->
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('45115454-a2f6-5863-8962-2fafce569f01', 'target.import.parse', 'v0.1.0', 'en', 'e6534a387dace3083dce37a058e8d357d9720a68794f797e076956dad22f8560', $body$You are an expert technical interviewer assistant. Extract the interview-ready
target job model from the following job description. Respond in English.

JD source URL (empty for non-URL imports): `{{jd_source_url}}`
JD raw text:

{{jd_text}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

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

Example JSON:
```json
{
  "coreThemes": [
    "string"
  ],
  "interviewHypotheses": [
    "string"
  ],
  "strengths": [
    "string"
  ],
  "gaps": [
    "string"
  ],
  "riskSignals": [
    "string"
  ],
  "requirements": [
    {
      "kind": "must_have",
      "label": "string"
    }
  ]
}
```
<!-- output-schema-contract:end -->

Do not include markdown fences in the JSON output.
$body$, TRUE, '2026-05-09T11:30:00Z'),
  ('3e4dae23-7bc3-56cb-868e-72e7c8a6c331', 'target.import.parse', 'v0.1.0', 'multi', '72a6e9b17d4a59ff28c360dd71ec562822a5f0026485f07b3b10efd1613b3b8e', $body$You are an expert technical interviewer assistant. Extract the interview-ready
target job model from the following job description. Respond strictly in the
language identified by the `{{language}}` variable; if `{{language}}` is empty
or unknown, respond in English.

JD source URL (empty for non-URL imports): `{{jd_source_url}}`
JD raw text:

{{jd_text}}

<!-- output-schema-contract:start -->
Return strict JSON matching this schema-derived output contract.

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

Example JSON:
```json
{
  "coreThemes": [
    "string"
  ],
  "interviewHypotheses": [
    "string"
  ],
  "strengths": [
    "string"
  ],
  "gaps": [
    "string"
  ],
  "riskSignals": [
    "string"
  ],
  "requirements": [
    {
      "kind": "must_have",
      "label": "string"
    }
  ]
}
```
<!-- output-schema-contract:end -->

Do not include markdown fences in the JSON output.
$body$, TRUE, '2026-05-09T11:30:00Z')
ON CONFLICT (feature_key, version, language) DO NOTHING;

INSERT INTO rubric_versions (id, feature_key, version, language, schema_json, is_active, created_at) VALUES
  ('4f095285-46a6-5802-83f1-c7277c41959d', 'debrief.generate', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "Generated question analyses preserve the candidate's recorded interview beats.", "name": "debrief_recall_completeness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Question analyses name the underlying interview signal rather than restating the answer.", "name": "debrief_lesson_specificity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Risk items are concrete, severity-calibrated, and actionable for the next preparation round.", "name": "debrief_action_quality", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "debrief.generate", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('8921c937-9ab1-52cc-9502-718bfb0a5461', 'debrief.generate', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Generated question analyses preserve the candidate's recorded interview beats.", "name": "debrief_recall_completeness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Question analyses name the underlying interview signal rather than restating the answer.", "name": "debrief_lesson_specificity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Risk items are concrete, severity-calibrated, and actionable for the next preparation round.", "name": "debrief_action_quality", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "debrief.generate", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('3e230e42-3104-55cd-98a1-1e662637aa1f', 'debrief.suggest_questions', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "Suggested questions cover the likely interview stages and the user's available context.", "name": "debrief_recall_completeness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Rationales explain the signal behind each suggested question.", "name": "debrief_lesson_specificity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Output language matches the requested locale and uses stable source labels.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.25}], "feature_key": "debrief.suggest_questions", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-16T00:00:00Z'),
  ('a8d1acd0-39d8-55ec-8ad4-cf4c8ef39948', 'debrief.suggest_questions', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Suggested questions cover the likely interview stages and the user's available context.", "name": "debrief_recall_completeness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Rationales explain the signal behind each suggested question.", "name": "debrief_lesson_specificity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Output language matches the requested locale and uses stable source labels.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.25}], "feature_key": "debrief.suggest_questions", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-16T00:00:00Z'),
  ('48dec31c-156d-5df0-ab5e-47517effd2a5', 'practice.session.first_question', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "Question or follow-up probes deeply into the candidate's reasoning rather than staying on the surface.", "name": "practice_depth", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Across the session, rubric dimensions are exercised without leaving large blind spots.", "name": "practice_dimension_coverage", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "practice.session.first_question", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('0c570c46-b7b6-5ff8-9c4e-01591e59d3a0', 'practice.session.first_question', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Question or follow-up probes deeply into the candidate's reasoning rather than staying on the surface.", "name": "practice_depth", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Across the session, rubric dimensions are exercised without leaving large blind spots.", "name": "practice_dimension_coverage", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "practice.session.first_question", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('197588f1-b503-51eb-bbfe-30fd3d156162', 'practice.session.follow_up', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "Follow-up question targets the candidate's actual gap or signal rather than recycling the prior turn.", "name": "followup_relevance", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Question or follow-up probes deeply into the candidate's reasoning rather than staying on the surface.", "name": "practice_depth", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "practice.session.follow_up", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('673b43cf-3cda-5eed-8fa6-2872157da379', 'practice.session.follow_up', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Follow-up question targets the candidate's actual gap or signal rather than recycling the prior turn.", "name": "followup_relevance", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Question or follow-up probes deeply into the candidate's reasoning rather than staying on the surface.", "name": "practice_depth", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "practice.session.follow_up", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('6c0247f5-4006-5006-b6a6-225ea0bc08c8', 'practice.turn.lightweight_observe', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "The cue or generated artifact surfaces a high-signal moment rather than commentary.", "name": "practice_signal_strength", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Output is unambiguous and parseable without re-reading.", "name": "practice_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.2}], "feature_key": "practice.turn.lightweight_observe", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('b827bc4a-b976-53f1-a79d-f0fe9087905f', 'practice.turn.lightweight_observe', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "The cue or generated artifact surfaces a high-signal moment rather than commentary.", "name": "practice_signal_strength", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Output is unambiguous and parseable without re-reading.", "name": "practice_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.2}], "feature_key": "practice.turn.lightweight_observe", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('48186e3b-9e2e-5e00-b498-278c7bf82bed', 'report.generate', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "Conclusions are anchored in sanitized turn summaries or recorded artifacts.", "name": "report_evidence", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Report cites concrete evidence and avoids hedging or generic language.", "name": "report_specificity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.25}, {"description": "Recommended next actions are specific, owned, and time-bounded.", "name": "report_action_quality", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.25}, {"description": "Final scores align with rubric levels and with the qualitative reasoning.", "name": "report_calibration", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.15}], "feature_key": "report.generate", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('b3fedd43-277e-59f3-be48-382c7e046891', 'report.generate', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Conclusions are anchored in sanitized turn summaries or recorded artifacts.", "name": "report_evidence", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.35}, {"description": "Report cites concrete evidence and avoids hedging or generic language.", "name": "report_specificity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.25}, {"description": "Recommended next actions are specific, owned, and time-bounded.", "name": "report_action_quality", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.25}, {"description": "Final scores align with rubric levels and with the qualitative reasoning.", "name": "report_calibration", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.15}], "feature_key": "report.generate", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('8105d792-0d67-5734-a054-8583144357ff', 'report.question_assessment', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "Conclusions are anchored in sanitized turn summaries or recorded artifacts.", "name": "report_evidence", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Final scores align with rubric levels and with the qualitative reasoning.", "name": "report_calibration", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Per-dimension scores stay calibrated against the rubric levels and the cited evidence.", "name": "score_outlier", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "report.question_assessment", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('12ab4687-4190-5792-b103-01ef7601b803', 'report.question_assessment', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Conclusions are anchored in sanitized turn summaries or recorded artifacts.", "name": "report_evidence", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Final scores align with rubric levels and with the qualitative reasoning.", "name": "report_calibration", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Per-dimension scores stay calibrated against the rubric levels and the cited evidence.", "name": "score_outlier", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "report.question_assessment", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('fbe62b90-b5b9-5486-bd51-ef2d1f8e1180', 'resume.parse', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "Resume content aligns with the target JD's required and preferred signals.", "name": "resume_match", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Bullets and section structure read cleanly without filler or jargon.", "name": "resume_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Edits preserve the candidate's stated facts and avoid embellishment.", "name": "resume_truthfulness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "resume.parse", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('fdcb8879-51c7-538b-b51f-ad6c4c892ea7', 'resume.parse', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Resume content aligns with the target JD's required and preferred signals.", "name": "resume_match", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Bullets and section structure read cleanly without filler or jargon.", "name": "resume_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Edits preserve the candidate's stated facts and avoid embellishment.", "name": "resume_truthfulness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "resume.parse", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('5e8d9b79-dd1f-5ff6-b86d-3f1a227fae9f', 'resume.tailor.bullet_suggestions', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "Bullets read as outcomes rather than activities and quantify the impact when possible.", "name": "resume_impact", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Edits preserve the candidate's stated facts and avoid embellishment.", "name": "resume_truthfulness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Bullets and section structure read cleanly without filler or jargon.", "name": "resume_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "resume.tailor.bullet_suggestions", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('b774fe9d-7d92-5f2f-8399-810d19ed8446', 'resume.tailor.bullet_suggestions', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Bullets read as outcomes rather than activities and quantify the impact when possible.", "name": "resume_impact", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Edits preserve the candidate's stated facts and avoid embellishment.", "name": "resume_truthfulness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Bullets and section structure read cleanly without filler or jargon.", "name": "resume_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "resume.tailor.bullet_suggestions", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('23e8b273-a94e-535e-b181-8ba9bf9a0a6d', 'resume.tailor.gap_review', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "Resume content aligns with the target JD's required and preferred signals.", "name": "resume_match", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Edits preserve the candidate's stated facts and avoid embellishment.", "name": "resume_truthfulness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Bullets and section structure read cleanly without filler or jargon.", "name": "resume_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "resume.tailor.gap_review", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('39db2d84-c578-5757-a790-b91a9b43c88d', 'resume.tailor.gap_review', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "Resume content aligns with the target JD's required and preferred signals.", "name": "resume_match", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Edits preserve the candidate's stated facts and avoid embellishment.", "name": "resume_truthfulness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}, {"description": "Bullets and section structure read cleanly without filler or jargon.", "name": "resume_clarity", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.3}], "feature_key": "resume.tailor.gap_review", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('a44100a0-db1c-5400-8a0d-91a7bef88961', 'target.import.parse', 'v0.1.0', 'en', $schema${"dimensions": [{"description": "All major JD fields (role, seniority, skills, responsibilities) are captured.", "name": "target_extraction_completeness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Captured fields reflect the JD without invention or paraphrase drift.", "name": "target_field_accuracy", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.2}], "feature_key": "target.import.parse", "language": "en", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z'),
  ('dc1358dd-ad2a-58e4-a7bb-16b8fb730780', 'target.import.parse', 'v0.1.0', 'multi', $schema${"dimensions": [{"description": "All major JD fields (role, seniority, skills, responsibilities) are captured.", "name": "target_extraction_completeness", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Captured fields reflect the JD without invention or paraphrase drift.", "name": "target_field_accuracy", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.4}, {"description": "Output language matches the requested locale and uses consistent terminology.", "name": "language_consistency", "score_levels": [{"description": "Falls clearly short of the dimension; the artifact would not satisfy the user goal.", "label": "weak", "threshold": 0.0}, {"description": "Partially satisfies the dimension; still has obvious gaps a reviewer would call out.", "label": "developing", "threshold": 0.4}, {"description": "Meets the dimension at production-baseline quality with only minor refinements.", "label": "proficient", "threshold": 0.7}, {"description": "Exceeds the dimension's expectation; an experienced reviewer would highlight the work.", "label": "strong", "threshold": 0.9}], "weight": 0.2}], "feature_key": "target.import.parse", "language": "multi", "version": "v0.1.0"}$schema$, TRUE, '2026-05-09T11:30:00Z')
ON CONFLICT (feature_key, version, language) DO NOTHING;

COMMIT;
