import unittest
from pathlib import Path

import yaml
import scripts.lint.openapi_inventory as inventory


class OpenAPIInventoryContractTest(unittest.TestCase):
    def test_v18_inventory_includes_delete_me(self) -> None:
        self.assertEqual(37, len(inventory.EXPECTED_OPERATIONS))
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


if __name__ == "__main__":
    unittest.main()
