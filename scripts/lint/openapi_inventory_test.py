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
        response_ref = operation["responses"]["422"]["content"]["application/json"]["schema"]["$ref"]
        self.assertEqual("#/components/schemas/CreatePracticeVoiceTurnRequest", request_ref)
        self.assertEqual("#/components/schemas/ApiErrorResponse", response_ref)

        schemas = data["components"]["schemas"]
        request = schemas["CreatePracticeVoiceTurnRequest"]
        self.assertEqual(["clientVoiceTurnId", "audio", "language"], request["required"])
        self.assertNotIn("turnId", request["properties"])
        self.assertNotIn("practiceMode", request["properties"])
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

    def test_practice_message_contract_and_not_found_errors_registered(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        schemas = data["components"]["schemas"]

        self.assertNotIn("PracticeMode", schemas)
        self.assertNotIn("QuestionReviewStatus", schemas)
        self.assertEqual(["user", "assistant"], schemas["PracticeMessage"]["properties"]["role"]["enum"])
        self.assertIn("PRACTICE_PLAN_NOT_FOUND", schemas["ApiErrorCode"]["enum"])
        self.assertIn("PRACTICE_SESSION_NOT_FOUND", schemas["ApiErrorCode"]["enum"])

    def test_openapi_001_grounded_direct_report_contract(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        schemas = data["components"]["schemas"]

        self.assertNotIn("DimensionResult", schemas)
        self.assertIn("REPORT_CONTEXT_TOO_LARGE", schemas["ApiErrorCode"]["enum"])

        report = schemas["FeedbackReport"]
        self.assertFalse(report["additionalProperties"])
        self.assertEqual(
            [
                "id",
                "sessionId",
                "targetJobId",
                "status",
                "errorCode",
                "summary",
                "context",
                "preparednessLevel",
                "highlights",
                "issues",
                "nextActions",
                "dimensionAssessments",
                "retryFocusDimensionCodes",
                "provenance",
                "createdAt",
                "updatedAt",
            ],
            report["required"],
        )
        self.assertNotIn("retryFocusCompetencyCodes", report["properties"])
        self.assertEqual(1, report["properties"]["summary"]["oneOf"][0]["minLength"])
        self.assertEqual(360, report["properties"]["summary"]["oneOf"][0]["maxLength"])
        self.assertEqual(
            "#/components/schemas/ReportContextSnapshot",
            report["properties"]["context"]["$ref"],
        )
        self.assertEqual(6, report["properties"]["dimensionAssessments"]["maxItems"])
        ready_properties = report["allOf"][0]["then"]["properties"]
        self.assertEqual(1, ready_properties["dimensionAssessments"]["minItems"])
        self.assertEqual(1, ready_properties["nextActions"]["minItems"])
        self.assertEqual(
            "#/components/schemas/ReadinessTier",
            ready_properties["preparednessLevel"]["$ref"],
        )
        self.assertEqual(
            "#/components/schemas/GenerationProvenance",
            ready_properties["provenance"]["$ref"],
        )
        error_state = report["allOf"][1]
        self.assertEqual("failed", error_state["if"]["properties"]["status"]["const"])
        self.assertEqual(
            "#/components/schemas/ApiErrorCode",
            error_state["then"]["properties"]["errorCode"]["$ref"],
        )
        self.assertEqual("null", error_state["else"]["properties"]["errorCode"]["type"])
        non_ready_properties = report["allOf"][0]["else"]["properties"]
        self.assertEqual(0, non_ready_properties["dimensionAssessments"]["maxItems"])
        self.assertEqual("null", non_ready_properties["summary"]["type"])
        self.assertTrue(report["properties"]["retryFocusDimensionCodes"]["uniqueItems"])

        context = schemas["ReportContextSnapshot"]
        self.assertFalse(context["additionalProperties"])
        self.assertEqual(
            [
                "sourcePlanId",
                "targetJobTitle",
                "targetJobCompany",
                "resumeId",
                "resumeDisplayName",
                "roundId",
                "roundSequence",
                "roundName",
                "roundType",
                "language",
                "hasNextRound",
            ],
            context["required"],
        )

        dimension = schemas["DimensionAssessment"]
        self.assertEqual(["code", "label", "status", "confidence"], dimension["required"])
        self.assertNotIn("dimension", dimension["properties"])
        self.assertEqual(r"^[a-z][a-z0-9_]{1,63}$", dimension["properties"]["code"]["pattern"])
        self.assertFalse(dimension["additionalProperties"])

        for schema_name in ("ReportHighlight", "ReportIssue"):
            evidence = schemas[schema_name]
            self.assertNotIn("dimension", evidence["properties"])
            self.assertIn("dimensionCode", evidence["properties"])
            self.assertEqual(240, evidence["properties"]["evidence"]["maxLength"])
            self.assertFalse(evidence["additionalProperties"])

        self.assertEqual(200, schemas["ReportNextAction"]["properties"]["label"]["maxLength"])
        self.assertFalse(schemas["ReportNextAction"]["additionalProperties"])

        request = schemas["CreatePracticePlanRequest"]
        self.assertEqual("object", request["type"])
        self.assertEqual(["goal"], request["required"])
        self.assertFalse(request["additionalProperties"])
        self.assertNotIn("focusCompetencyCodes", request["properties"])
        self.assertFalse(request["properties"]["sourceReportId"].get("nullable", False))
        self.assertEqual(2, len(request["oneOf"]))

    def test_practice_session_exposes_ordered_messages_without_turn_state(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        schemas = data["components"]["schemas"]

        self.assertNotIn("PracticeTurn", schemas)
        self.assertNotIn("PracticeSessionEventRequest", schemas)
        self.assertEqual(
            "#/components/schemas/PracticeMessage",
            schemas["PracticeSession"]["properties"]["messages"]["items"]["$ref"],
        )

    def test_practice_message_owner_spec_and_baseline_are_in_sync(self) -> None:
        baseline = yaml.safe_load(Path("openapi/baseline/openapi-v1.0.0.yaml").read_text(encoding="utf-8"))
        baseline_schemas = baseline["components"]["schemas"]
        owner_spec = Path("docs/spec/openapi-v1-contract/spec.md").read_text(encoding="utf-8")

        self.assertIn("PracticeMessage", baseline_schemas)
        self.assertNotIn("PracticeTurn", baseline_schemas)
        self.assertNotIn("PracticeSessionEventRequest", baseline_schemas)
        self.assertIn("D-19 | Practice conversation pre-launch rebase", owner_spec)

    def test_practice_voice_request_has_no_turn_or_mode_contract(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        request = data["components"]["schemas"]["CreatePracticeVoiceTurnRequest"]
        self.assertNotIn("turnId", request["properties"])
        self.assertNotIn("practiceMode", request["properties"])

    def test_practice_message_generated_artifacts_are_in_sync(self) -> None:
        generated = yaml.safe_load(Path("backend/internal/api/generated/openapi.yaml").read_text(encoding="utf-8"))
        ts_types = Path("frontend/src/api/generated/types.ts").read_text(encoding="utf-8")

        self.assertIn("PracticeMessage", generated["components"]["schemas"])
        self.assertNotIn("PracticeTurn", generated["components"]["schemas"])
        self.assertIn("export interface PracticeMessage", ts_types)
        self.assertNotIn("export interface PracticeTurn", ts_types)

    def test_practice_round_identity_and_progress_contract_is_additive(self) -> None:
        for path in (
            Path("openapi/openapi.yaml"),
            Path("openapi/baseline/openapi-v1.0.0.yaml"),
        ):
            with self.subTest(path=path):
                schemas = yaml.safe_load(path.read_text(encoding="utf-8"))["components"]["schemas"]

                round_ref = schemas["PracticeRoundRef"]
                self.assertEqual(["roundId", "roundSequence"], round_ref["required"])
                self.assertNotIn("format", round_ref["properties"]["roundId"])
                self.assertEqual(
                    r"^round-[1-9][0-9]{0,9}-(hr|technical|manager|cross_functional|culture|final|other)$",
                    round_ref["properties"]["roundId"]["pattern"],
                )
                self.assertEqual(1, round_ref["properties"]["roundSequence"]["minimum"])
                self.assertEqual(2147483647, round_ref["properties"]["roundSequence"]["maximum"])

                progress = schemas["PracticeProgress"]
                self.assertEqual(["status", "completedRounds", "currentRound"], progress["required"])
                self.assertEqual(
                    ["not_started", "in_progress", "completed"],
                    progress["properties"]["status"]["enum"],
                )
                self.assertTrue(progress["properties"]["completedRounds"]["uniqueItems"])
                self.assertEqual(
                    "#/components/schemas/PracticeRoundRef",
                    progress["properties"]["completedRounds"]["items"]["$ref"],
                )
                self.assertEqual(
                    [
                        {"$ref": "#/components/schemas/PracticeRoundRef"},
                        {"type": "null"},
                    ],
                    progress["properties"]["currentRound"]["oneOf"],
                )

                request = schemas["CreatePracticePlanRequest"]
                self.assertNotIn("roundId", request["required"])
                self.assertNotIn("roundSequence", request["properties"])
                self.assertNotIn("format", request["properties"]["roundId"])
                self.assertEqual(round_ref["properties"]["roundId"]["pattern"], request["properties"]["roundId"]["pattern"])

                plan = schemas["PracticePlan"]
                self.assertNotIn("roundId", plan["required"])
                self.assertNotIn("roundSequence", plan["required"])
                self.assertEqual(["roundSequence"], plan["dependentRequired"]["roundId"])
                self.assertEqual(["roundId"], plan["dependentRequired"]["roundSequence"])
                self.assertIn({"type": "null"}, plan["properties"]["roundId"]["oneOf"])
                self.assertIn({"type": "null"}, plan["properties"]["roundSequence"]["oneOf"])
                plan_sequence = next(item for item in plan["properties"]["roundSequence"]["oneOf"] if item.get("type") == "integer")
                self.assertEqual(2147483647, plan_sequence["maximum"])

                target_job = schemas["TargetJob"]
                self.assertNotIn("practiceProgress", target_job["required"])
                self.assertEqual(
                    "#/components/schemas/PracticeProgress",
                    target_job["properties"]["practiceProgress"]["$ref"],
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
        mutated["components"]["schemas"]["PracticeMode"] = {
            "type": "string",
            "enum": ["assisted", "strict"],
        }
        mutated["components"]["schemas"]["ReportNextAction"]["properties"]["type"]["enum"] = ["drill", "review"]
        errors: list[str] = []

        inventory.validate_product_scope_contract(mutated, errors)

        self.assertTrue(any("PracticeMode" in err and "stale" in err for err in errors), errors)
        self.assertTrue(any("ReportNextAction.type" in err for err in errors), errors)


if __name__ == "__main__":
    unittest.main()
