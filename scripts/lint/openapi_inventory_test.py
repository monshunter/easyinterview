import copy
import unittest
from pathlib import Path

import yaml
import scripts.lint.openapi_inventory as inventory


class OpenAPIInventoryContractTest(unittest.TestCase):
    def test_product_scope_v12_inventory_includes_delete_me(self) -> None:
        self.assertEqual(34, len(inventory.EXPECTED_OPERATIONS))
        self.assertIn(("Auth", "delete", "/me", "deleteMe"), inventory.EXPECTED_OPERATIONS)
        self.assertIn(("delete", "/me"), inventory.IK_REQUIRED)

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
