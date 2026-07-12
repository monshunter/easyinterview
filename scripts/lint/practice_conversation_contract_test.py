#!/usr/bin/env python3
"""Cross-layer contract tests for the simplified Practice conversation."""

from __future__ import annotations

import re
import unittest
from pathlib import Path

import yaml


ROOT = Path(__file__).resolve().parents[2]


class PracticeConversationContractTest(unittest.TestCase):
    def test_openapi_exposes_messages_without_question_state(self) -> None:
        openapi_text = (ROOT / "openapi/openapi.yaml").read_text(encoding="utf-8")
        document = yaml.safe_load(openapi_text)
        paths = document["paths"]
        schemas = document["components"]["schemas"]

        message_path = "/practice/sessions/{sessionId}/messages"
        self.assertEqual("sendPracticeMessage", paths[message_path]["post"]["operationId"])
        self.assertNotIn("/practice/sessions/{sessionId}/events", paths)
        self.assertNotIn("/practice/sessions/{sessionId}/events", openapi_text)

        for schema_name in ("PracticeMessage", "SendPracticeMessageRequest", "SendPracticeMessageResponse"):
            self.assertIn(schema_name, schemas)
        for schema_name in (
            "PracticeMode",
            "PracticeTurn",
            "PracticeSessionEventRequest",
            "SessionEventResult",
            "AssistantAction",
            "QuestionAssessment",
        ):
            self.assertNotIn(schema_name, schemas)

        self.assertNotIn("mode", schemas["PracticePlan"]["properties"])
        self.assertNotIn("questionBudget", schemas["PracticePlan"]["properties"])
        self.assertEqual(
            "#/components/schemas/PracticeMessage",
            schemas["PracticeSession"]["properties"]["messages"]["items"]["$ref"],
        )
        for field in ("hintsEnabled", "turnCount", "currentTurn"):
            self.assertNotIn(field, schemas["PracticeSession"]["properties"])

        report_properties = schemas["FeedbackReport"]["properties"]
        self.assertIn("dimensionAssessments", report_properties)
        self.assertIn("retryFocusCompetencyCodes", report_properties)
        self.assertNotIn("questionAssessments", report_properties)
        self.assertNotIn("retryFocusTurnIds", report_properties)

    def test_shared_conventions_have_eleven_current_enums(self) -> None:
        conventions = yaml.safe_load((ROOT / "shared/conventions.yaml").read_text(encoding="utf-8"))
        names = [entry["name"] for entry in conventions["enums"]]

        self.assertEqual(11, len(names))
        self.assertNotIn("PracticeMode", names)
        self.assertNotIn("QuestionReviewStatus", names)

    def test_internal_events_are_conversation_level(self) -> None:
        events = yaml.safe_load((ROOT / "shared/events.yaml").read_text(encoding="utf-8"))
        by_name = {entry["name"]: entry for entry in events["events"]}

        self.assertEqual(13, len(by_name))
        self.assertNotIn("practice.turn.completed", by_name)
        self.assertNotIn("mode", by_name["practice.session.started"]["requiredPayload"])
        self.assertNotIn("turnCount", by_name["practice.session.completed"]["requiredPayload"])
        self.assertNotIn("questionIssueCount", by_name["report.generated"]["requiredPayload"])

    def test_baseline_uses_messages_and_twenty_six_total_tables(self) -> None:
        migration = (ROOT / "migrations/000001_create_baseline.up.sql").read_text(encoding="utf-8").lower()
        table_names = re.findall(r"^create table ([a-z0-9_]+) ", migration, flags=re.MULTILINE)

        self.assertEqual(26, len(table_names))
        self.assertIn("practice_messages", table_names)
        self.assertNotIn("practice_turns", table_names)
        self.assertNotIn("question_assessments", table_names)
        for stale_column in (
            "question_budget",
            "hints_enabled",
            "turn_count",
            "question_text",
            "question_intent",
            "answer_text",
            "hint_text",
            "retry_focus_turn_ids",
        ):
            self.assertNotIn(stale_column, migration)

    def test_prompt_registry_has_six_conversation_level_coordinates(self) -> None:
        prompt_root = ROOT / "config/prompts"
        feature_keys = sorted(path.name for path in prompt_root.iterdir() if path.is_dir())

        self.assertEqual(
            [
                "practice.session.chat",
                "report.generate",
                "resume.parse",
                "resume.tailor.bullet_suggestions",
                "resume.tailor.gap_review",
                "target.import.parse",
            ],
            feature_keys,
        )


if __name__ == "__main__":
    unittest.main()
