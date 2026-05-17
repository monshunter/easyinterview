import copy
import unittest
from pathlib import Path

import yaml
import scripts.lint.openapi_inventory as inventory


class OpenAPIInventoryContractTest(unittest.TestCase):
    def test_product_scope_v12_inventory_includes_delete_me(self) -> None:
        # 56 + 1 for the frontend-debrief/001 Phase 0 cross-owner addendum
        # that adds `listPracticeSessions` to the inventory.
        self.assertEqual(57, len(inventory.EXPECTED_OPERATIONS))
        self.assertIn(("Auth", "delete", "/me", "deleteMe"), inventory.EXPECTED_OPERATIONS)
        self.assertIn(("delete", "/me"), inventory.IK_REQUIRED)

    def test_list_practice_sessions_operation_is_registered_without_idempotency_key(self) -> None:
        # frontend-debrief/001 Phase 0 cross-owner addendum: GET
        # /practice/sessions surfaces the Mock Session picker dataset.
        # As a read-only operation it must not require Idempotency-Key.
        self.assertIn(
            ("PracticeSessions", "get", "/practice/sessions", "listPracticeSessions"),
            inventory.EXPECTED_OPERATIONS,
        )
        self.assertNotIn(("get", "/practice/sessions"), inventory.IK_REQUIRED)
        self.assertNotIn(("get", "/practice/sessions"), inventory.IK_FORBIDDEN)

    def test_debrief_suggestions_operation_is_registered_without_idempotency_key(self) -> None:
        self.assertIn(
            ("Debriefs", "post", "/debriefs/question-suggestions", "suggestDebriefQuestions"),
            inventory.EXPECTED_OPERATIONS,
        )
        self.assertIn(("post", "/debriefs/question-suggestions"), inventory.IK_FORBIDDEN)
        self.assertNotIn(("post", "/debriefs/question-suggestions"), inventory.IK_REQUIRED)

    def test_resume_workshop_additive_inventory_is_resumes_tag_only(self) -> None:
        resume_ops = {
            ("Resumes", "get", "/resumes", "listResumes"),
            ("Resumes", "get", "/resumes/{resumeAssetId}/versions", "listResumeVersions"),
            ("Resumes", "get", "/resume-versions/{resumeVersionId}", "getResumeVersion"),
            ("Resumes", "post", "/resume-versions", "branchResumeVersion"),
            ("Resumes", "patch", "/resume-versions/{resumeVersionId}", "updateResumeVersion"),
            ("Resumes", "post", "/resume-versions/{resumeVersionId}/suggestions/{suggestionId}/accept", "acceptResumeTailorSuggestion"),
            ("Resumes", "post", "/resume-versions/{resumeVersionId}/suggestions/{suggestionId}/reject", "rejectResumeTailorSuggestion"),
            ("Resumes", "post", "/resumes/{resumeAssetId}/archive", "archiveResumeAsset"),
            ("Resumes", "post", "/resume-versions/{resumeVersionId}/exports", "exportResumeVersion"),
        }

        for row in resume_ops:
            self.assertIn(row, inventory.EXPECTED_OPERATIONS)

        self.assertNotIn("ResumeVersions", inventory.EXPECTED_TAGS)
        self.assertIn("ResumeVersion", inventory.AI_PROVENANCE_SCHEMAS)
        self.assertEqual(
            {
                ("post", "/privacy/exports"): "PRIVACY_EXPORT_NOT_AVAILABLE",
                ("post", "/resume-versions/{resumeVersionId}/exports"): "RESUME_EXPORT_NOT_AVAILABLE",
            },
            inventory.P0_501_ENDPOINTS,
        )

    def test_resume_workshop_side_effects_require_idempotency_key(self) -> None:
        for row in {
            ("post", "/resume-versions"),
            ("patch", "/resume-versions/{resumeVersionId}"),
            ("post", "/resume-versions/{resumeVersionId}/suggestions/{suggestionId}/accept"),
            ("post", "/resume-versions/{resumeVersionId}/suggestions/{suggestionId}/reject"),
            ("post", "/resumes/{resumeAssetId}/archive"),
            ("post", "/resume-versions/{resumeVersionId}/exports"),
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

    def test_p0_debrief_keeps_p1_followup_fields_optional_and_hidden(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        debrief = data["components"]["schemas"]["Debrief"]
        required = set(debrief["required"])

        self.assertNotIn("thankYouDraft", required)
        self.assertNotIn("nextRoundChecklist", required)

        props = debrief["properties"]
        for name in ("thankYouDraft", "nextRoundChecklist"):
            self.assertIn("P1 optional/hidden", props[name]["description"])

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
            ["asked", "answered", "follow_up_requested", "assessed", "skipped"],
            schemas["PracticeTurn"]["properties"]["status"]["enum"],
        )

    def test_practice_turn_status_generated_artifacts_are_in_sync(self) -> None:
        generated = yaml.safe_load(Path("backend/internal/api/generated/openapi.yaml").read_text(encoding="utf-8"))
        generated_status = generated["components"]["schemas"]["PracticeTurn"]["properties"]["status"]["enum"]
        ts_types = Path("frontend/src/api/generated/types.ts").read_text(encoding="utf-8")

        self.assertEqual(["asked", "answered", "follow_up_requested", "assessed", "skipped"], generated_status)
        self.assertIn(
            'status: "asked" | "answered" | "follow_up_requested" | "assessed" | "skipped";',
            ts_types,
        )

    def test_resume_workshop_contract_uses_b1_vocabulary(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        schemas = data["components"]["schemas"]

        self.assertEqual(["structured_master", "targeted"], schemas["ResumeVersionType"]["enum"])
        self.assertEqual(["copy_master", "blank", "ai_select"], schemas["ResumeSeedStrategy"]["enum"])
        self.assertEqual(["pending", "accepted", "rejected"], schemas["ResumeTailorSuggestionStatus"]["enum"])
        self.assertIn("RESUME_EXPORT_NOT_AVAILABLE", schemas["ApiErrorCode"]["enum"])

        resume_version = schemas["ResumeVersion"]
        reachable = inventory.reachable_schemas(schemas, ["ResumeVersion"])
        self.assertIn("GenerationProvenance", reachable)
        self.assertIn("provenance", resume_version["required"])

    def test_register_resume_contract_supports_fileless_sources(self) -> None:
        data = yaml.safe_load(Path("openapi/openapi.yaml").read_text(encoding="utf-8"))
        schemas = data["components"]["schemas"]

        register_resume = schemas["RegisterResumeRequest"]
        self.assertNotIn("fileObjectId", register_resume["required"])
        self.assertIn("sourceType", register_resume["properties"])
        self.assertIn("rawText", register_resume["properties"])
        self.assertIn("guidedAnswers", register_resume["properties"])
        self.assertEqual("string", register_resume["properties"]["fileObjectId"]["type"])
        self.assertTrue(register_resume["properties"]["fileObjectId"]["nullable"])

        resume_asset = schemas["ResumeAsset"]
        self.assertNotIn("fileObjectId", resume_asset["required"])
        self.assertEqual("string", resume_asset["properties"]["fileObjectId"]["type"])
        self.assertTrue(resume_asset["properties"]["fileObjectId"]["nullable"])

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

    def test_product_scope_semantic_invariants_reject_legacy_practice_values(self) -> None:
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
