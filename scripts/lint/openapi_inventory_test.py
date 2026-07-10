import copy
import unittest
from pathlib import Path

import yaml
import scripts.lint.openapi_inventory as inventory


class OpenAPIInventoryContractTest(unittest.TestCase):
    def test_product_scope_v21_inventory_includes_delete_me(self) -> None:
        # Current v1.0.0 pre-launch freeze after D-17 JD match removal and
        # D-20 flat resume contract.
        self.assertEqual(37, len(inventory.EXPECTED_OPERATIONS))
        self.assertIn(("Auth", "delete", "/me", "deleteMe"), inventory.EXPECTED_OPERATIONS)
        self.assertIn(("delete", "/me"), inventory.IK_REQUIRED)

    def test_complete_my_profile_operation_is_registered_without_idempotency_key(self) -> None:
        # Profile completion is guarded by the authenticated session and the
        # latest displayName/terms payload; it is not a generic idempotent op.
        self.assertIn(("Auth", "patch", "/me", "completeMyProfile"), inventory.EXPECTED_OPERATIONS)
        self.assertNotIn(("patch", "/me"), inventory.IK_REQUIRED)
        self.assertNotIn(("patch", "/me"), inventory.IK_FORBIDDEN)

    def test_list_practice_sessions_operation_is_registered_without_idempotency_key(self) -> None:
        # GET /practice/sessions surfaces target-job session history.
        # As a read-only operation it must not require Idempotency-Key.
        self.assertIn(
            ("PracticeSessions", "get", "/practice/sessions", "listPracticeSessions"),
            inventory.EXPECTED_OPERATIONS,
        )
        self.assertNotIn(("get", "/practice/sessions"), inventory.IK_REQUIRED)
        self.assertNotIn(("get", "/practice/sessions"), inventory.IK_FORBIDDEN)

    def test_practice_voice_turn_operation_is_registered_as_session_side_effect(self) -> None:
        path = "/practice/sessions/{sessionId}/voice-turns"

        self.assertIn(("PracticeSessions", "post", path, "createPracticeVoiceTurn"), inventory.EXPECTED_OPERATIONS)
        self.assertIn(("post", path), inventory.IK_REQUIRED)
        self.assertNotIn(("post", path), inventory.IK_FORBIDDEN)

        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        operation = data["paths"][path]["post"]
        self.assertEqual("createPracticeVoiceTurn", operation["operationId"])
        self.assertIn({"$ref": inventory.IDEMPOTENCY_REF}, operation["parameters"])

        request_ref = operation["requestBody"]["content"]["application/json"]["schema"]["$ref"]
        response_ref = operation["responses"]["200"]["content"]["application/json"]["schema"]["$ref"]
        self.assertEqual("#/components/schemas/CreatePracticeVoiceTurnRequest", request_ref)
        self.assertEqual("#/components/schemas/PracticeVoiceTurnResult", response_ref)

        schemas = data["components"]["schemas"]
        request = schemas["CreatePracticeVoiceTurnRequest"]
        self.assertEqual(
            ["clientVoiceTurnId", "turnId", "audio", "language", "practiceMode"],
            request["required"],
        )
        self.assertNotIn("manualTranscriptFallback", request["properties"])
        self.assertEqual(
            ["contentBase64", "contentType", "durationMs"],
            schemas["PracticeVoiceAudioInput"]["required"],
        )

        result = schemas["PracticeVoiceTurnResult"]
        self.assertEqual(
            [
                "voiceTurnId",
                "userTranscriptFinal",
                "assistantTextDraft",
                "ttsChunks",
                "providerMetaSummary",
                "session",
                "ttsError",
            ],
            result["required"],
        )
        self.assertEqual("#/components/schemas/PracticeVoiceTTSChunk", result["properties"]["ttsChunks"]["items"]["$ref"])
        self.assertEqual(
            [{"$ref": "#/components/schemas/PracticeVoiceTTSError"}, {"type": "null"}],
            result["properties"]["ttsError"]["oneOf"],
        )

    def test_d22_debrief_and_profile_operations_are_removed(self) -> None:
        self.assertNotIn("Profile", inventory.EXPECTED_TAGS)
        self.assertNotIn("Debriefs", inventory.EXPECTED_TAGS)
        removed_operation_ids = {
            "getMyProfile",
            "updateMyProfile",
            "listExperienceCards",
            "createExperienceCard",
            "updateExperienceCard",
            "createDebrief",
            "suggestDebriefQuestions",
            "getDebrief",
        }
        self.assertFalse(removed_operation_ids & {opid for *_rest, opid in inventory.EXPECTED_OPERATIONS})
        self.assertNotIn(("post", "/debriefs"), inventory.IK_REQUIRED)
        self.assertNotIn(("post", "/debriefs/question-suggestions"), inventory.IK_FORBIDDEN)

    def test_resume_workshop_additive_inventory_is_resumes_tag_only(self) -> None:
        resume_ops = {
            ("Resumes", "post", "/resumes", "registerResume"),
            ("Resumes", "get", "/resumes", "listResumes"),
            ("Resumes", "get", "/resumes/{resumeId}", "getResume"),
            ("Resumes", "patch", "/resumes/{resumeId}", "updateResume"),
            ("Resumes", "post", "/resumes/{resumeId}/duplicate", "duplicateResume"),
            ("Resumes", "post", "/resumes/{resumeId}/archive", "archiveResume"),
            ("Resumes", "post", "/resumes/{resumeId}/exports", "exportResume"),
        }

        for row in resume_ops:
            self.assertIn(row, inventory.EXPECTED_OPERATIONS)

        out_of_scope_ops = {
            "listResumeVersions",
            "getResumeVersion",
            "branchResumeVersion",
            "updateResumeVersion",
            "acceptResumeTailorSuggestion",
            "rejectResumeTailorSuggestion",
            "archiveResumeAsset",
            "exportResumeVersion",
        }
        self.assertFalse(out_of_scope_ops & {opid for *_rest, opid in inventory.EXPECTED_OPERATIONS})
        self.assertNotIn("ResumeVersions", inventory.EXPECTED_TAGS)
        self.assertIn("Resume", inventory.AI_PROVENANCE_SCHEMAS)
        self.assertEqual(
            {
                ("post", "/privacy/exports"): "PRIVACY_EXPORT_NOT_AVAILABLE",
                ("post", "/resumes/{resumeId}/exports"): "RESUME_EXPORT_NOT_AVAILABLE",
            },
            inventory.P0_501_ENDPOINTS,
        )

    def test_resume_workshop_side_effects_require_idempotency_key(self) -> None:
        for row in {
            ("post", "/resumes"),
            ("patch", "/resumes/{resumeId}"),
            ("post", "/resumes/{resumeId}/duplicate"),
            ("post", "/resumes/{resumeId}/archive"),
            ("post", "/resumes/{resumeId}/exports"),
        }:
            self.assertIn(row, inventory.IK_REQUIRED)

    def test_delete_me_contract_uses_idempotent_privacy_delete_job(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        operation = data["paths"]["/me"]["delete"]

        self.assertEqual("deleteMe", operation["operationId"])
        self.assertIn({"$ref": inventory.IDEMPOTENCY_REF}, operation["parameters"])
        self.assertIn("active", operation["description"])
        self.assertIn("privacy_delete", operation["description"])

        response = operation["responses"]["202"]
        schema = response["content"]["application/json"]["schema"]
        self.assertEqual("#/components/schemas/PrivacyRequestWithJob", schema["$ref"])

    def test_product_scope_semantic_invariants_are_enforced(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        errors: list[str] = []

        inventory.validate_product_scope_contract(data, errors)

        self.assertEqual([], errors)

    def test_practice_mode_contract_is_binary_and_not_found_errors_registered(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        schemas = data["components"]["schemas"]

        self.assertEqual(["assisted", "strict"], schemas["PracticeMode"]["enum"])
        self.assertIn("PRACTICE_PLAN_NOT_FOUND", schemas["ApiErrorCode"]["enum"])
        self.assertIn("PRACTICE_SESSION_NOT_FOUND", schemas["ApiErrorCode"]["enum"])

    def test_practice_turn_status_contract_exposes_runtime_states(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        schemas = data["components"]["schemas"]

        self.assertEqual(
            ["asked", "answered", "follow_up_requested", "assessed"],
            schemas["PracticeTurn"]["properties"]["status"]["enum"],
        )

    def test_practice_turn_status_owner_spec_and_baseline_are_in_sync(self) -> None:
        expected_statuses = ["asked", "answered", "follow_up_requested", "assessed"]
        baseline = yaml.safe_load(Path("openapi/baseline/openapi-v1.0.0.yaml").read_text(encoding="utf-8"))
        baseline_schemas = baseline["components"]["schemas"]
        owner_spec = Path("docs/spec/openapi-v1-contract/spec.md").read_text(encoding="utf-8")

        self.assertEqual(
            expected_statuses,
            baseline_schemas["PracticeTurn"]["properties"]["status"]["enum"],
        )
        self.assertNotIn(
            "turn_skipped",
            baseline_schemas["PracticeSessionEventRequest"]["properties"]["kind"]["enum"],
        )
        self.assertIn(
            "`PracticeTurn.status` wire enum 原地 rebase 为 4 值："
            "`asked` / `answered` / `follow_up_requested` / `assessed`",
            owner_spec,
        )
        self.assertNotIn("`skipped`", owner_spec)
        self.assertNotIn("turn_skipped", owner_spec)

    def test_practice_voice_event_kinds_extend_append_session_event_schema(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        kinds = data["components"]["schemas"]["PracticeSessionEventRequest"]["properties"]["kind"]["enum"]

        for kind in (
            "tts_chunk_started",
            "tts_chunk_played",
            "barge_in_detected",
            "assistant_context_committed",
        ):
            self.assertIn(kind, kinds)
        self.assertNotIn("turn_skipped", kinds)

    def test_practice_turn_status_generated_artifacts_are_in_sync(self) -> None:
        generated = yaml.safe_load(Path("backend/internal/api/generated/openapi.yaml").read_text(encoding="utf-8"))
        generated_status = generated["components"]["schemas"]["PracticeTurn"]["properties"]["status"]["enum"]
        ts_types = Path("frontend/src/api/generated/types.ts").read_text(encoding="utf-8")

        self.assertEqual(["asked", "answered", "follow_up_requested", "assessed"], generated_status)
        self.assertIn(
            'status: "asked" | "answered" | "follow_up_requested" | "assessed";',
            ts_types,
        )

    def test_resume_workshop_contract_uses_b1_vocabulary(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        schemas = data["components"]["schemas"]

        self.assertNotIn("ResumeVersionType", schemas)
        self.assertNotIn("ResumeSeedStrategy", schemas)
        self.assertNotIn("ResumeTailorSuggestionStatus", schemas)
        self.assertIn("RESUME_EXPORT_NOT_AVAILABLE", schemas["ApiErrorCode"]["enum"])

        resume = schemas["Resume"]
        reachable = inventory.reachable_schemas(schemas, ["Resume"])
        self.assertIn("GenerationProvenance", reachable)
        self.assertEqual(["upload", "paste"], resume["properties"]["sourceType"]["enum"])
        self.assertIn("structuredProfile", resume["properties"])

    def test_register_resume_contract_supports_fileless_sources(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        schemas = data["components"]["schemas"]

        register_resume = schemas["RegisterResumeRequest"]
        self.assertNotIn("fileObjectId", register_resume["required"])
        self.assertIn("sourceType", register_resume["properties"])
        self.assertIn("rawText", register_resume["properties"])
        self.assertNotIn("guidedAnswers", register_resume["properties"])
        self.assertEqual(["upload", "paste"], register_resume["properties"]["sourceType"]["enum"])
        self.assertEqual("string", register_resume["properties"]["fileObjectId"]["type"])
        self.assertTrue(register_resume["properties"]["fileObjectId"]["nullable"])

        resume = schemas["Resume"]
        self.assertNotIn("fileObjectId", resume["required"])
        self.assertEqual("string", resume["properties"]["fileObjectId"]["type"])
        self.assertTrue(resume["properties"]["fileObjectId"]["nullable"])

    def test_product_scope_semantic_invariants_reject_old_mistakes_scope(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        mutated = copy.deepcopy(data)
        mutated["tags"].append({"name": "Mistakes"})
        mutated["paths"]["/mistakes"] = {
            "get": {
                "tags": ["Mistakes"],
                "operationId": "listMistakes",
                "responses": {"default": {"$ref": inventory.APIERROR_REF}},
            }
        }
        mutated["components"]["schemas"]["TargetJob"]["properties"]["openMistakeCount"] = {"type": "integer"}
        errors: list[str] = []

        inventory.validate_product_scope_contract(mutated, errors)

        self.assertTrue(any("Mistakes" in err for err in errors), errors)
        self.assertTrue(any("/mistakes" in err for err in errors), errors)
        self.assertTrue(any("openMistakeCount" in err for err in errors), errors)

    def test_product_scope_semantic_invariants_reject_standalone_voice_scope(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        mutated = copy.deepcopy(data)
        mutated["tags"].append({"name": "Voice"})
        mutated["paths"]["/voice/sessions"] = {
            "post": {
                "tags": ["Voice"],
                "operationId": "startVoiceSession",
                "responses": {"default": {"$ref": inventory.APIERROR_REF}},
            }
        }
        errors: list[str] = []

        inventory.validate_product_scope_contract(mutated, errors)

        self.assertTrue(any("Voice" in err for err in errors), errors)
        self.assertTrue(any("/voice" in err for err in errors), errors)

    def test_product_scope_semantic_invariants_reject_out_of_scope_practice_values(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        mutated = copy.deepcopy(data)
        mutated["components"]["schemas"]["PracticeMode"]["enum"] = ["warmup", "core_interview", "single_drill"]
        mutated["components"]["schemas"]["ReportNextAction"]["properties"]["type"]["enum"] = ["drill", "review"]
        errors: list[str] = []

        inventory.validate_product_scope_contract(mutated, errors)

        self.assertTrue(any("PracticeMode" in err for err in errors), errors)
        self.assertTrue(any("ReportNextAction.type" in err for err in errors), errors)


if __name__ == "__main__":
    unittest.main()
