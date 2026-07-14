import copy
import unittest
from pathlib import Path

import yaml
import scripts.lint.openapi_inventory as inventory
import scripts.lint.validate_fixtures as fixture_validator


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

    def test_targetjob_report_overview_is_closed_and_keeps_endpoint_inventory(self) -> None:
        decision = Path(
            "docs/spec/openapi-v1-contract/decisions/OPENAPI-004-targetjob-report-overview.md"
        ).read_text(encoding="utf-8")
        self.assertIn("> **状态**: accepted", decision)

        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        operation = data["paths"]["/targets/{targetJobId}/reports"]["get"]
        self.assertEqual("listTargetJobReports", operation["operationId"])
        self.assertEqual(
            ["targetJobId"],
            [parameter["name"] for parameter in operation["parameters"] if "name" in parameter],
        )
        self.assertEqual(
            [
                "#/components/parameters/XRequestID",
                "#/components/parameters/Traceparent",
                "#/components/parameters/AcceptLanguage",
                "#/components/parameters/XClientVersion",
            ],
            [parameter["$ref"] for parameter in operation["parameters"] if "$ref" in parameter],
        )
        self.assertEqual(
            "#/components/schemas/TargetJobReportsOverview",
            operation["responses"]["200"]["content"]["application/json"]["schema"]["$ref"],
        )

        schemas = data["components"]["schemas"]
        self.assertNotIn("PaginatedFeedbackReport", schemas)
        self.assertNotIn("latestReportId", schemas["TargetJob"]["properties"])
        self.assertIs(False, schemas["PracticeRoundRef"]["additionalProperties"])

        expected_shapes = {
            "TargetJobReportsOverview": (
                ["targetJobId", "rounds"],
                ["targetJobId", "rounds"],
            ),
            "TargetJobReportRoundOverview": (
                ["round", "currentReport", "latestAttempt"],
                ["round", "currentReport", "latestAttempt"],
            ),
            "TargetJobCurrentReportSummary": (
                ["id", "generatedAt"],
                ["id", "generatedAt"],
            ),
            "TargetJobReportAttemptSummary": (
                ["id", "status", "errorCode", "createdAt"],
                ["id", "status", "errorCode", "createdAt"],
            ),
        }
        for schema_name, (required, properties) in expected_shapes.items():
            with self.subTest(schema=schema_name):
                schema = schemas[schema_name]
                self.assertEqual("object", schema["type"])
                self.assertIs(False, schema["additionalProperties"])
                self.assertEqual(required, schema["required"])
                self.assertEqual(properties, list(schema["properties"]))

        round_overview = schemas["TargetJobReportRoundOverview"]
        self.assertEqual(
            "#/components/schemas/PracticeRoundRef",
            round_overview["properties"]["round"]["$ref"],
        )
        self.assertEqual(
            (2, 5),
            (
                schemas["TargetJobReportsOverview"]["properties"]["rounds"]["minItems"],
                schemas["TargetJobReportsOverview"]["properties"]["rounds"]["maxItems"],
            ),
        )
        self.assertEqual(
            [
                {"$ref": "#/components/schemas/TargetJobCurrentReportSummary"},
                {"type": "null"},
            ],
            round_overview["properties"]["currentReport"]["oneOf"],
        )
        self.assertEqual(
            [
                {"$ref": "#/components/schemas/TargetJobReportAttemptSummary"},
                {"type": "null"},
            ],
            round_overview["properties"]["latestAttempt"]["oneOf"],
        )

        attempt = schemas["TargetJobReportAttemptSummary"]
        self.assertEqual(
            "#/components/schemas/ReportStatus",
            attempt["properties"]["status"]["$ref"],
        )
        self.assertEqual(
            [
                {"$ref": "#/components/schemas/ApiErrorCode"},
                {"type": "null"},
            ],
            attempt["properties"]["errorCode"]["oneOf"],
        )

        valid = {
            "targetJobId": "01918fa0-0000-7000-8000-000000000001",
            "rounds": [
                {
                    "round": {
                        "roundId": "round-1-technical",
                        "roundSequence": 1,
                    },
                    "currentReport": {
                        "id": "01918fa0-0000-7000-8000-000000000010",
                        "generatedAt": "2026-07-14T01:02:03Z",
                    },
                    "latestAttempt": {
                        "id": "01918fa0-0000-7000-8000-000000000011",
                        "status": "failed",
                        "errorCode": "AI_PROVIDER_TIMEOUT",
                        "createdAt": "2026-07-14T01:03:03Z",
                    },
                },
                {
                    "round": {
                        "roundId": "round-2-manager",
                        "roundSequence": 2,
                    },
                    "currentReport": None,
                    "latestAttempt": None,
                },
            ],
        }
        errors: list[str] = []
        fixture_validator.schema_validate(
            valid,
            schemas["TargetJobReportsOverview"],
            root=data,
            path="response",
            errors=errors,
        )
        self.assertEqual([], errors)

        for status in ("queued", "generating", "ready"):
            with self.subTest(valid_status=status):
                body = copy.deepcopy(valid)
                body["rounds"][0]["latestAttempt"]["status"] = status
                body["rounds"][0]["latestAttempt"]["errorCode"] = None
                errors = []
                fixture_validator.schema_validate(
                    body,
                    schemas["TargetJobReportsOverview"],
                    root=data,
                    path="response",
                    errors=errors,
                )
                self.assertEqual([], errors)

        invalid_values = [
            {**valid, "pageInfo": {}},
            {
                **valid,
                "rounds": [{"round": valid["rounds"][0]["round"], "currentReport": None}],
            },
            {
                **valid,
                "rounds": [
                    {
                        **valid["rounds"][0],
                        "latestAttempt": {
                            **valid["rounds"][0]["latestAttempt"],
                            "status": "generating",
                        },
                    }
                ],
            },
            {
                **valid,
                "rounds": [
                    {
                        **valid["rounds"][0],
                        "latestAttempt": {
                            **valid["rounds"][0]["latestAttempt"],
                            "errorCode": None,
                        },
                    },
                    valid["rounds"][1],
                ],
            },
            {
                **valid,
                "rounds": [
                    {
                        **valid["rounds"][0],
                        "round": {
                            **valid["rounds"][0]["round"],
                            "unexpected": True,
                        },
                    }
                ],
            },
            {
                **valid,
                "rounds": [
                    {
                        **valid["rounds"][0],
                        "currentReport": {
                            **valid["rounds"][0]["currentReport"],
                            "provenance": {},
                        },
                    }
                ],
            },
        ]
        for index, body in enumerate(invalid_values):
            with self.subTest(invalid=index):
                errors = []
                fixture_validator.schema_validate(
                    body,
                    schemas["TargetJobReportsOverview"],
                    root=data,
                    path="response",
                    errors=errors,
                )
                self.assertTrue(errors, body)

        self.assertEqual(37, len(inventory.EXPECTED_OPERATIONS))
        self.assertEqual(10, len(inventory.EXPECTED_TAGS))

    def test_targetjob_report_overview_semantic_linter_rejects_legacy_and_open_shapes(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        errors: list[str] = []
        inventory.validate_targetjob_report_overview_contract(data, errors)
        self.assertEqual([], errors)

        mutated = copy.deepcopy(data)
        operation = mutated["paths"]["/targets/{targetJobId}/reports"]["get"]
        operation["parameters"].insert(
            1,
            {"name": "cursor", "in": "query", "schema": {"type": "string"}},
        )
        schemas = mutated["components"]["schemas"]
        schemas["PaginatedFeedbackReport"] = {"type": "object"}
        schemas["TargetJob"]["properties"]["latestReportId"] = {"type": "string"}
        schemas["PracticeRoundRef"].pop("additionalProperties")
        schemas["TargetJobReportsOverview"]["properties"]["rounds"].pop("minItems")
        schemas["TargetJobReportRoundOverview"]["properties"]["round"] = {
            "type": "object"
        }
        schemas["TargetJobReportRoundOverview"].pop("additionalProperties")
        schemas["TargetJobReportAttemptSummary"]["required"].remove("errorCode")
        errors = []
        inventory.validate_targetjob_report_overview_contract(mutated, errors)

        self.assertTrue(any("named parameters" in error for error in errors), errors)
        self.assertTrue(any("PaginatedFeedbackReport" in error for error in errors), errors)
        self.assertTrue(any("latestReportId" in error for error in errors), errors)
        self.assertTrue(any("PracticeRoundRef" in error for error in errors), errors)
        self.assertTrue(any("2..5" in error for error in errors), errors)
        self.assertTrue(any("round must reference" in error for error in errors), errors)
        self.assertTrue(any("additionalProperties" in error for error in errors), errors)
        self.assertTrue(any("errorCode" in error and "required" in error for error in errors), errors)

    def test_targetjob_report_overview_generated_artifacts_are_typed(self) -> None:
        generated = yaml.safe_load(
            Path("backend/internal/api/generated/openapi.yaml").read_text(encoding="utf-8")
        )
        ts_types = Path("frontend/src/api/generated/types.ts").read_text(encoding="utf-8")
        ts_client = Path("frontend/src/api/generated/client.ts").read_text(encoding="utf-8")
        go_types = Path("backend/internal/api/generated/types.gen.go").read_text(encoding="utf-8")
        go_server = Path("backend/internal/api/generated/server.gen.go").read_text(encoding="utf-8")

        operation = generated["paths"]["/targets/{targetJobId}/reports"]["get"]
        self.assertEqual(
            "#/components/schemas/TargetJobReportsOverview",
            operation["responses"]["200"]["content"]["application/json"]["schema"]["$ref"],
        )
        self.assertNotIn("PaginatedFeedbackReport", generated["components"]["schemas"])

        for type_name in (
            "TargetJobReportsOverview",
            "TargetJobReportRoundOverview",
            "TargetJobCurrentReportSummary",
            "TargetJobReportAttemptSummary",
        ):
            self.assertIn(f"export interface {type_name}", ts_types)
            self.assertNotIn(f"export type {type_name} = any", ts_types)
            self.assertIn(f"type {type_name} struct", go_types)

        self.assertNotIn("PaginatedFeedbackReport", ts_types)
        self.assertNotIn("PaginatedFeedbackReport", go_types)
        self.assertNotIn("latestReportId", ts_types)
        self.assertNotIn("LatestReportId", go_types)
        self.assertIn(
            "async listTargetJobReports(targetJobId: string, opts?: RequestOptions): Promise<Types.TargetJobReportsOverview>",
            ts_client,
        )
        self.assertIn(
            "ListTargetJobReports(w http.ResponseWriter, r *http.Request, targetJobId string)",
            go_server,
        )

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

        message = schemas["PracticeMessage"]
        self.assertEqual(
            [
                {"$ref": "#/components/schemas/PracticeUserMessage"},
                {"$ref": "#/components/schemas/PracticeAssistantMessage"},
            ],
            message["oneOf"],
        )
        self.assertEqual("role", message["discriminator"]["propertyName"])
        self.assertEqual(
            {
                "user": "#/components/schemas/PracticeUserMessage",
                "assistant": "#/components/schemas/PracticeAssistantMessage",
            },
            message["discriminator"]["mapping"],
        )

        user_message = schemas["PracticeUserMessage"]
        assistant_message = schemas["PracticeAssistantMessage"]
        self.assertFalse(user_message["additionalProperties"])
        self.assertFalse(assistant_message["additionalProperties"])
        self.assertEqual(["user"], user_message["properties"]["role"]["enum"])
        self.assertEqual(["assistant"], assistant_message["properties"]["role"]["enum"])
        self.assertTrue({"clientMessageId", "replyStatus"}.issubset(user_message["required"]))
        self.assertTrue({"clientMessageId", "replyStatus"}.issubset(user_message["properties"]))
        self.assertTrue({"clientMessageId", "replyStatus"}.isdisjoint(assistant_message["required"]))
        self.assertTrue({"clientMessageId", "replyStatus"}.isdisjoint(assistant_message["properties"]))
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

    def test_openapi_002_targetjob_intake_is_paste_only_and_operations_are_stable(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        schemas = data["components"]["schemas"]

        request = schemas["ImportTargetJobRequest"]
        self.assertEqual("object", request["type"])
        self.assertFalse(request["additionalProperties"])
        self.assertEqual(["rawText", "targetLanguage", "resumeId"], request["required"])
        self.assertEqual(
            {"rawText", "targetLanguage", "resumeId"},
            set(request["properties"]),
        )
        self.assertEqual("string", request["properties"]["rawText"]["type"])
        self.assertEqual(1, request["properties"]["rawText"]["minLength"])
        self.assertEqual(r"\S", request["properties"]["rawText"]["pattern"])

        for removed_schema in (
            "TargetJobImportSourceURL",
            "TargetJobImportSourceManualText",
            "TargetJobImportSourceFile",
            "TargetJobImportSourceManualForm",
            "TargetJobImportSource",
        ):
            self.assertNotIn(removed_schema, schemas)

        target_job = schemas["TargetJob"]
        self.assertNotIn("sourceType", target_job["required"])
        self.assertNotIn("sourceType", target_job["properties"])
        self.assertNotIn("sourceUrl", target_job["properties"])
        self.assertEqual(
            ["resume", "privacy_export"],
            schemas["UploadPresignRequest"]["properties"]["purpose"]["enum"],
        )

        import_operation = data["paths"]["/targets/import"]["post"]
        self.assertEqual("importTargetJob", import_operation["operationId"])
        self.assertEqual(
            "#/components/schemas/ImportTargetJobRequest",
            import_operation["requestBody"]["content"]["application/json"]["schema"]["$ref"],
        )
        self.assertEqual(
            "#/components/schemas/TargetJobWithJob",
            import_operation["responses"]["202"]["content"]["application/json"]["schema"]["$ref"],
        )

        presign_operation = data["paths"]["/uploads/presign"]["post"]
        self.assertEqual("createUploadPresign", presign_operation["operationId"])
        self.assertEqual(
            "#/components/schemas/UploadPresignRequest",
            presign_operation["requestBody"]["content"]["application/json"]["schema"]["$ref"],
        )
        self.assertEqual(
            "#/components/schemas/UploadPresign",
            presign_operation["responses"]["201"]["content"]["application/json"]["schema"]["$ref"],
        )
        self.assertEqual(37, len(inventory.EXPECTED_OPERATIONS))
        self.assertEqual(10, len(inventory.EXPECTED_TAGS))

    def test_targetjob_paste_only_semantic_linter_rejects_source_compatibility(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        errors: list[str] = []

        inventory.validate_targetjob_paste_only_contract(data, errors)

        self.assertEqual([], errors)

        mutated = copy.deepcopy(data)
        request = mutated["components"]["schemas"]["ImportTargetJobRequest"]
        request["properties"]["source"] = {"type": "object"}
        request["required"].append("source")
        mutated["components"]["schemas"]["TargetJob"]["properties"]["sourceUrl"] = {
            "type": "string",
        }
        mutated["components"]["schemas"]["UploadPresignRequest"]["properties"]["purpose"]["enum"].append(
            "target_job_attachment"
        )
        next(tag for tag in mutated["tags"] if tag["name"] == "Uploads")["description"] = (
            "Pre-signed object-storage URLs for resume / JD attachment uploads."
        )
        errors = []

        inventory.validate_targetjob_paste_only_contract(mutated, errors)

        self.assertTrue(any("ImportTargetJobRequest" in error and "properties" in error for error in errors), errors)
        self.assertTrue(any("TargetJob" in error and "sourceUrl" in error for error in errors), errors)
        self.assertTrue(any("UploadPresignRequest.purpose" in error for error in errors), errors)
        self.assertTrue(any("Uploads tag description" in error for error in errors), errors)

    def test_targetjob_paste_only_request_accepts_text_and_rejects_legacy_payloads(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        schema = data["components"]["schemas"]["ImportTargetJobRequest"]
        resume_id = "01918fa0-0001-7000-8000-000000000001"
        valid = {
            "rawText": "Senior backend engineer\nGo and PostgreSQL required.",
            "targetLanguage": "zh-CN",
            "resumeId": resume_id,
        }
        errors: list[str] = []
        fixture_validator.schema_validate(
            valid,
            schema,
            root=data,
            path="request",
            errors=errors,
        )
        self.assertEqual([], errors)

        invalid_payloads = [
            {**valid, "rawText": value}
            for value in ("", " ", "\t", "\n", " \t\n ")
        ] + [
            {
                "source": {"type": "manual_text", "rawText": valid["rawText"]},
                "targetLanguage": "zh-CN",
                "resumeId": resume_id,
            },
            {**valid, "source": {"type": "url", "url": "https://example.com/jd"}},
            {**valid, "source": {"type": "file", "fileObjectId": resume_id}},
            {
                **valid,
                "source": {
                    "type": "manual_form",
                    "title": "Backend engineer",
                    "rawDescription": "legacy",
                },
            },
            {**valid, "fileObjectId": resume_id},
            {**valid, "titleHint": "Backend engineer"},
            {**valid, "companyNameHint": "Example"},
            {**valid, "unexpected": True},
        ]
        for index, body in enumerate(invalid_payloads):
            with self.subTest(index=index, body=body):
                errors = []
                fixture_validator.schema_validate(
                    body,
                    schema,
                    root=data,
                    path="request",
                    errors=errors,
                )
                self.assertTrue(errors, body)

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
        go_types = Path("backend/internal/api/generated/types.gen.go").read_text(encoding="utf-8")

        generated_schemas = generated["components"]["schemas"]
        self.assertEqual(
            [
                {"$ref": "#/components/schemas/PracticeUserMessage"},
                {"$ref": "#/components/schemas/PracticeAssistantMessage"},
            ],
            generated_schemas["PracticeMessage"]["oneOf"],
        )
        self.assertNotIn("PracticeTurn", generated_schemas)
        self.assertIn(
            "export type PracticeMessage = PracticeUserMessage | PracticeAssistantMessage;",
            ts_types,
        )
        self.assertNotIn("export interface PracticeMessage", ts_types)
        self.assertNotIn("export interface PracticeTurn", ts_types)

        union_start = go_types.index("type PracticeMessage struct {")
        union_end = go_types.index("type PracticeSession struct {", union_start)
        go_union = go_types[union_start:union_end]
        self.assertIn("func (value PracticeMessage) MarshalJSON()", go_union)
        self.assertIn("func (value *PracticeMessage) UnmarshalJSON", go_union)
        self.assertIn("decoder.DisallowUnknownFields()", go_union)
        self.assertNotIn("any", go_union)
        self.assertNotIn("type PracticeMessage = any", go_types)

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
