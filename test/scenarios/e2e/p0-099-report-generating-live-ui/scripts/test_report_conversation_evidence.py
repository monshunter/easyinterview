#!/usr/bin/env python3
"""Code-level contracts for P0.099's bounded conversation evidence.

These tests deliberately exercise only the scenario helper modules.  They are
not invoked by setup.sh, trigger.sh, or verify.sh and therefore are not E2E
evidence themselves.
"""

from __future__ import annotations

import importlib.util
import json
import struct
import sys
import tempfile
import unittest
import zlib
from copy import deepcopy
from pathlib import Path


SCRIPTS = Path(__file__).resolve().parent
REPORT_ID = "01918fa0-0070-7000-8000-000000000070"
SESSION_ID = "01918fa0-0080-7000-8000-000000000080"
RUN_ID = "e2e-p0-099-unit-test"


def load_script(name: str):
    path = SCRIPTS / name
    spec = importlib.util.spec_from_file_location(f"p0099_{path.stem}", path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {path}")
    module = importlib.util.module_from_spec(spec)
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module


capture = load_script("capture_live_evidence.py")
validator = load_script("validate_evidence.py")


def public_context() -> dict[str, object]:
    return {
        "sourcePlanId": "01918fa0-0040-7000-8000-000000000040",
        "targetJobTitle": "Platform Engineer",
        "targetJobCompany": "Acme",
        "resumeId": "01918fa0-0010-7000-8000-000000000010",
        "resumeDisplayName": "Platform resume",
        "roundId": "round-2-technical",
        "roundSequence": 2,
        "roundName": "Technical interview",
        "roundType": "technical",
        "language": "zh-CN",
        "hasNextRound": True,
    }


def api_conversation() -> dict[str, object]:
    return {
        "reportId": REPORT_ID,
        "reportStatus": "ready",
        "context": public_context(),
        "messages": [
            {
                "sequence": 2,
                "role": "user",
                "content": "private candidate sentence",
                "createdAt": "2026-07-15T08:00:12Z",
            },
            {
                "sequence": 5,
                "role": "assistant",
                "content": "private interviewer follow-up",
                "createdAt": "2026-07-15T08:00:23.123456Z",
            },
        ],
    }


def database_conversation() -> dict[str, object]:
    context = public_context()
    return {
        "report_ref": REPORT_ID,
        "session_ref": SESSION_ID,
        "status": "ready",
        "generation_context": {
            "plan": {"id": context["sourcePlanId"]},
            "targetJob": {
                "title": context["targetJobTitle"],
                "company": context["targetJobCompany"],
            },
            "resume": {
                "id": context["resumeId"],
                "displayName": context["resumeDisplayName"],
            },
            "round": {
                "id": context["roundId"],
                "sequence": context["roundSequence"],
                "name": context["roundName"],
                "type": context["roundType"],
            },
            "conversation": {
                "sessionId": SESSION_ID,
                "language": context["language"],
                "messageCount": 2,
                "lastMessageSeqNo": 5,
            },
            "hasNextRound": context["hasNextRound"],
        },
        "messages": [
            {
                "sequence": 2,
                "role": "user",
                "content": "private candidate sentence",
                "created_at": "2026-07-15T08:00:12.000000Z",
            },
            {
                "sequence": 5,
                "role": "assistant",
                "content": "private interviewer follow-up",
                "created_at": "2026-07-15T08:00:23.123456Z",
            },
        ],
    }


def navigation_artifact() -> dict[str, object]:
    return {
        "schema_version": "p0-099-conversation-navigation.v1",
        "scenario_id": "E2E.P0.099",
        "run_id": RUN_ID,
        "method": "real-browser-report-conversation-back",
        "report_ref": REPORT_ID,
        "urls": {
            "report": f"/report?reportId={REPORT_ID}",
            "conversation": f"/report-conversation?reportId={REPORT_ID}",
            "back": f"/report?reportId={REPORT_ID}",
        },
        "request_audit": {
            "report_get_path": f"/api/v1/reports/{REPORT_ID}",
            "report_get_count": 1,
            "conversation_get_path": f"/api/v1/reports/{REPORT_ID}/conversation",
            "conversation_get_count": 1,
            "public_session_list_request_count": 0,
            "route_interception_used": False,
        },
        "privacy": {
            "transcript_prose_written": False,
            "internal_locator_written": False,
            "browser_state_written": False,
        },
    }


def png_with_metadata(kind: bytes, payload: bytes) -> bytes:
    def chunk(chunk_kind: bytes, chunk_payload: bytes) -> bytes:
        checksum = zlib.crc32(chunk_kind + chunk_payload) & 0xFFFFFFFF
        return struct.pack(">I", len(chunk_payload)) + chunk_kind + chunk_payload + struct.pack(">I", checksum)

    header = struct.pack(">IIBBBBB", 1, 1, 8, 2, 0, 0, 0)
    pixels = zlib.compress(b"\x00\x20\x40\x60")
    return b"\x89PNG\r\n\x1a\n" + chunk(b"IHDR", header) + chunk(kind, payload) + chunk(b"IDAT", pixels) + chunk(b"IEND", b"")


class ReportConversationEvidenceTests(unittest.TestCase):
    def test_ready_visual_audit_requires_bottom_summary_after_actions(self) -> None:
        self.assertIn(
            "bottom_interview_summary_visible_after_actions",
            validator.READY_VISUAL_CHECKS,
        )

    def test_db_and_authenticated_api_projection_share_redacted_strict_digests(self) -> None:
        api_projection = capture.project_conversation(api_conversation(), REPORT_ID)
        database_projection = capture.project_database_conversation(database_conversation(), REPORT_ID)

        self.assertEqual(api_projection["report_status"], "ready")
        self.assertEqual(database_projection["report_status"], "ready")
        for key in (
            "context_digest",
            "message_count",
            "strict_sequence_digest",
            "ordered_message_digest",
        ):
            self.assertEqual(api_projection[key], database_projection[key])
        self.assertEqual(database_projection["session_ref"], SESSION_ID)
        self.assertNotIn("private candidate sentence", json.dumps(api_projection))
        self.assertNotIn("private interviewer follow-up", json.dumps(database_projection))

    def test_conversation_projection_fails_closed_for_locator_or_non_increasing_sequence(self) -> None:
        with_locator = api_conversation()
        with_locator["sessionId"] = SESSION_ID
        with self.assertRaises(capture.CaptureError):
            capture.project_conversation(with_locator, REPORT_ID)

        out_of_order = api_conversation()
        out_of_order["messages"] = list(reversed(out_of_order["messages"]))
        with self.assertRaises(capture.CaptureError):
            capture.project_conversation(out_of_order, REPORT_ID)

    def test_database_conversation_query_reads_messages_in_seq_no_order(self) -> None:
        query = capture.database_conversation_query(REPORT_ID)
        self.assertIn("from practice_messages pm", query)
        self.assertIn("order by pm.seq_no asc", query)
        self.assertIn("default_transaction_read_only", capture.postgres_environment("postgres://user@localhost/db")["PGOPTIONS"])

    def test_live_capture_binds_current_ready_report_without_conversation_prose(self) -> None:
        database_projection = capture.project_database_conversation(database_conversation(), REPORT_ID)
        api_projection = capture.project_conversation(api_conversation(), REPORT_ID)
        bound = capture.bind_conversation_capture(
            database_projection,
            api_projection,
            {
                "session_ref": SESSION_ID,
                "status": "ready",
                "frozen_context_digest": database_projection["frozen_context_digest"],
            },
        )
        expected = {
            REPORT_ID: {
                "session_ref": SESSION_ID,
                "status": "ready",
                "frozen_context_digest": database_projection["frozen_context_digest"],
            }
        }
        validator.validate_conversation_capture(bound, expected, {"report_ref": REPORT_ID})
        self.assertNotIn("private candidate sentence", json.dumps(bound))
        self.assertNotIn("private interviewer follow-up", json.dumps(bound))

        invalid = deepcopy(bound)
        invalid["db"]["ordered_by"] = "created_at ASC"
        with self.assertRaises(validator.EvidenceError):
            validator.validate_conversation_capture(invalid, expected, {"report_ref": REPORT_ID})

    def test_navigation_artifact_is_report_id_only_and_survives_privacy_sanitize(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            output_dir = Path(raw_dir)
            path = output_dir / "conversation-navigation.json"
            path.write_text(json.dumps(navigation_artifact()), encoding="utf-8")

            navigation = validator.validate_navigation_artifact(path, RUN_ID, {REPORT_ID})
            self.assertEqual(navigation["report_ref"], REPORT_ID)
            self.assertEqual(validator.sanitize_output(output_dir, failed=False), 0)
            self.assertTrue(path.exists())

            invalid = navigation_artifact()
            invalid["urls"] = dict(invalid["urls"])
            invalid["urls"]["conversation"] = f"/report-conversation?reportId={REPORT_ID}&sessionId={SESSION_ID}"
            path.write_text(json.dumps(invalid), encoding="utf-8")
            with self.assertRaises(validator.EvidenceError):
                validator.validate_navigation_artifact(path, RUN_ID, {REPORT_ID})

            invalid = navigation_artifact()
            invalid["request_audit"] = dict(invalid["request_audit"])
            invalid["request_audit"]["public_session_list_request_count"] = 1
            path.write_text(json.dumps(invalid), encoding="utf-8")
            with self.assertRaises(validator.EvidenceError):
                validator.validate_navigation_artifact(path, RUN_ID, {REPORT_ID})

    def test_png_metadata_gate_allows_technical_profile_but_rejects_session_material(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            profile_path = Path(raw_dir) / "profile.png"
            profile_payload = b"Display P3\x00\x00" + zlib.compress(b"synthetic color profile")
            profile_path.write_bytes(png_with_metadata(b"iCCP", profile_payload))
            width, height, channels, rows = validator.png_pixels(profile_path)
            self.assertEqual((width, height, channels, len(rows)), (1, 1, 3, 1))

            secret_path = Path(raw_dir) / "secret.png"
            secret_path.write_bytes(png_with_metadata(b"tEXt", b"Comment\x00ei_session=must-not-persist"))
            with self.assertRaises(validator.EvidenceError):
                validator.png_pixels(secret_path)

    def test_e2e_runners_do_not_call_this_code_level_test(self) -> None:
        for name in ("trigger.sh", "verify.sh"):
            source = (SCRIPTS / name).read_text(encoding="utf-8")
            self.assertNotIn("test_report_conversation_evidence", source)
            self.assertNotIn("unittest", source)


if __name__ == "__main__":
    unittest.main(verbosity=2)
