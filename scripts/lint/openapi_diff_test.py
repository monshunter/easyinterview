"""Unit tests for scripts/lint/openapi_diff.py.

Coverage targets each rule in openapi-v1-contract spec §4.4:

    Prohibited (breaking):
        delete-endpoint, rename-path, change-method, delete-field,
        change-type, required-add, delete-enum-value.

    Allowed (additive):
        new-endpoint, new-tag, new-optional-field, new-enum-value-string,
        new-optional-query, new-example.

    Privacy-export whitelist (spec §3.1 D-12 / §4.4 P0 例外):
        501 → 202 transition is informational *iff* history.md got a new row;
        otherwise the wrapper escalates to breaking.
"""

from __future__ import annotations

import copy
import json
import os
import subprocess
import sys
import unittest
from pathlib import Path
from textwrap import dedent

import yaml

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))

import openapi_diff as od  # noqa: E402  (path setup above)


def _baseline_doc() -> dict:
    return yaml.safe_load(dedent(
        """
        openapi: 3.1.0
        info:
          title: t
          version: 1.0.0
        tags:
          - name: TargetJobs
          - name: PracticePlans
          - name: Privacy
        paths:
          /target-jobs:
            get:
              operationId: listTargetJobs
              parameters:
                - name: cursor
                  in: query
                  required: false
                  schema: { type: string }
              responses:
                "200":
                  content:
                    application/json:
                      schema:
                        $ref: '#/components/schemas/TargetJob'
          /practice-plans/{planId}:
            get:
              operationId: getPracticePlan
              responses:
                "200":
                  content:
                    application/json:
                      schema:
                        $ref: '#/components/schemas/PracticePlan'
          /privacy/exports:
            post:
              operationId: requestPrivacyExport
              responses:
                "501":
                  content:
                    application/json:
                      schema:
                        $ref: '#/components/schemas/ApiErrorResponse'
        components:
          schemas:
            TargetJob:
              type: object
              required: [id, title]
              properties:
                id: { type: string, format: uuid }
                title: { type: string }
                status:
                  type: string
                  enum: [draft, active, archived]
            PracticePlan:
              type: object
              properties:
                id: { type: string }
            PrivacyRequestWithJob:
              type: object
              properties:
                id: { type: string }
            ApiErrorResponse:
              type: object
              properties:
                error: { type: object }
        """
    ))


class SpecRuleTests(unittest.TestCase):
    def setUp(self) -> None:
        self.base = _baseline_doc()

    # ---------- prohibited ---------------------------------------------------

    def test_delete_endpoint_is_breaking(self) -> None:
        cur = copy.deepcopy(self.base)
        del cur["paths"]["/practice-plans/{planId}"]
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(any(f["kind"] == "endpoint-removed" and f["severity"] == "breaking" for f in findings))

    def test_rename_path_is_breaking(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["paths"]["/practice-plans/{planUuid}"] = cur["paths"].pop("/practice-plans/{planId}")
        findings = od.diff_documents(self.base, cur)
        kinds = {f["kind"] for f in findings}
        self.assertIn("endpoint-removed", kinds)
        self.assertIn("endpoint-added", kinds)
        self.assertTrue(any(f["kind"] == "endpoint-removed" and f["severity"] == "breaking" for f in findings))

    def test_change_method_is_breaking(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["paths"]["/practice-plans/{planId}"]["put"] = cur["paths"]["/practice-plans/{planId}"].pop("get")
        findings = od.diff_documents(self.base, cur)
        kinds = [(f["kind"], f["severity"]) for f in findings]
        self.assertIn(("method-removed", "breaking"), kinds)
        self.assertIn(("method-added", "additive"), kinds)

    def test_delete_field_is_breaking(self) -> None:
        cur = copy.deepcopy(self.base)
        del cur["components"]["schemas"]["TargetJob"]["properties"]["title"]
        cur["components"]["schemas"]["TargetJob"]["required"] = ["id"]
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(any(f["kind"] == "field-deleted" and f["severity"] == "breaking" for f in findings))

    def test_change_type_is_breaking(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["components"]["schemas"]["TargetJob"]["properties"]["title"] = {"type": "integer"}
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(any(f["kind"] == "type-changed" and f["severity"] == "breaking" for f in findings))

    def test_oneof_branch_type_change_is_breaking(self) -> None:
        cur = copy.deepcopy(self.base)
        self.base["components"]["schemas"]["TargetJob"]["properties"]["nullableSessionId"] = {
            "oneOf": [
                {"type": "string", "format": "uuid"},
                {"type": "null"},
            ]
        }
        cur["components"]["schemas"]["TargetJob"]["properties"]["nullableSessionId"] = {
            "oneOf": [
                {"type": "integer"},
                {"type": "null"},
            ]
        }
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(
            any(
                f["kind"] == "type-changed"
                and f["severity"] == "breaking"
                and "nullableSessionId.oneOf[0]" in f["where"]
                for f in findings
            )
        )

    def test_allof_branch_ref_change_is_breaking(self) -> None:
        cur = copy.deepcopy(self.base)
        self.base["components"]["schemas"]["PaginatedTargetJob"] = {
            "allOf": [
                {"$ref": "#/components/schemas/PageInfo"},
                {
                    "type": "object",
                    "properties": {
                        "items": {
                            "type": "array",
                            "items": {"$ref": "#/components/schemas/TargetJob"},
                        }
                    },
                },
            ]
        }
        cur["components"]["schemas"]["PaginatedTargetJob"] = {
            "allOf": [
                {"$ref": "#/components/schemas/PageInfo"},
                {
                    "type": "object",
                    "properties": {
                        "items": {
                            "type": "array",
                            "items": {"$ref": "#/components/schemas/PracticePlan"},
                        }
                    },
                },
            ]
        }
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(
            any(
                f["kind"] == "ref-changed"
                and f["severity"] == "breaking"
                and "PaginatedTargetJob.allOf[1].items.items" in f["where"]
                for f in findings
            )
        )

    def test_required_add_on_existing_field_is_breaking(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["components"]["schemas"]["TargetJob"]["properties"]["status"] = {
            "type": "string",
            "enum": ["draft", "active", "archived"],
        }
        cur["components"]["schemas"]["TargetJob"]["required"] = ["id", "title", "status"]
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(
            any(f["kind"] == "field-promoted-required" and f["severity"] == "breaking" for f in findings)
        )

    def test_required_add_on_new_field_is_breaking(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["components"]["schemas"]["TargetJob"]["properties"]["mandatoryNew"] = {"type": "string"}
        cur["components"]["schemas"]["TargetJob"]["required"] = ["id", "title", "mandatoryNew"]
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(any(f["kind"] == "field-required-added" and f["severity"] == "breaking" for f in findings))

    def test_delete_enum_value_is_breaking(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["components"]["schemas"]["TargetJob"]["properties"]["status"]["enum"] = ["draft", "active"]
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(any(f["kind"] == "enum-value-removed" and f["severity"] == "breaking" for f in findings))

    def test_required_query_param_addition_is_breaking(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["paths"]["/target-jobs"]["get"]["parameters"].append({
            "name": "ownerId",
            "in": "query",
            "required": True,
            "schema": {"type": "string"},
        })
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(
            any(f["kind"] == "parameter-required-added-new" and f["severity"] == "breaking" for f in findings)
        )

    # ---------- allowed (additive) -------------------------------------------

    def test_new_endpoint_is_additive(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["paths"]["/target-jobs/new"] = {
            "get": {"operationId": "newOp", "responses": {"200": {}}}
        }
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(any(f["kind"] == "endpoint-added" and f["severity"] == "additive" for f in findings))
        self.assertFalse(any(f["severity"] == "breaking" for f in findings))

    def test_new_tag_is_additive(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["tags"].append({"name": "Growth"})
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(any(f["kind"] == "tag-added" and f["severity"] == "additive" for f in findings))

    def test_new_optional_field_is_additive(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["components"]["schemas"]["TargetJob"]["properties"]["metadata"] = {"type": "object"}
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(any(f["kind"] == "field-added" and f["severity"] == "additive" for f in findings))
        self.assertFalse(any(f["severity"] == "breaking" for f in findings))

    def test_new_string_enum_value_is_additive(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["components"]["schemas"]["TargetJob"]["properties"]["status"]["enum"] = [
            "draft", "active", "archived", "snoozed",
        ]
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(
            any(f["kind"] == "enum-value-added" and f["severity"] == "additive" for f in findings)
        )

    def test_new_optional_query_is_additive(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["paths"]["/target-jobs"]["get"]["parameters"].append({
            "name": "tagFilter",
            "in": "query",
            "required": False,
            "schema": {"type": "string"},
        })
        findings = od.diff_documents(self.base, cur)
        self.assertTrue(any(f["kind"] == "parameter-added" and f["severity"] == "additive" for f in findings))
        self.assertFalse(any(f["severity"] == "breaking" for f in findings))

    def test_new_example_is_not_breaking(self) -> None:
        cur = copy.deepcopy(self.base)
        cur["paths"]["/target-jobs"]["get"]["responses"]["200"]["content"]["application/json"]["examples"] = {
            "alpha": {"summary": "alpha", "value": {"id": "x"}},
        }
        findings = od.diff_documents(self.base, cur)
        self.assertFalse(any(f["severity"] == "breaking" for f in findings))


class HistoryGateTests(unittest.TestCase):
    def test_count_history_rows(self) -> None:
        text = dedent(
            """
            # History

            ## 1 修订记录

            | 日期 | 版本 | 变更 | 关联计划 |
            |------|------|------|----------|
            | 2026-04-28 | 1.0 | initial | - |
            | 2026-04-29 | 1.1 | add row | - |

            other content
            """
        )
        self.assertEqual(od._count_history_rows(text), 2)


class CLIWhitelistTests(unittest.TestCase):
    """End-to-end checks: build a tiny fake repo, run wrapper through CLI,
    confirm exit code + classifications match spec §4.4."""

    def _write_repo(self, tmp: Path) -> Path:
        repo = tmp
        (repo / "openapi" / "baseline").mkdir(parents=True)
        (repo / "docs" / "spec" / "openapi-v1-contract").mkdir(parents=True)

        baseline = _baseline_doc()
        with (repo / "openapi" / "baseline" / "openapi-v1.0.0.yaml").open("w") as f:
            yaml.safe_dump(baseline, f, sort_keys=False)
        with (repo / "openapi" / "openapi.yaml").open("w") as f:
            yaml.safe_dump(baseline, f, sort_keys=False)

        config_src = HERE.parent.parent / "openapi" / "diff-config.yaml"
        (repo / "openapi" / "diff-config.yaml").write_text(config_src.read_text())

        history_text = dedent(
            """\
            # OpenAPI v1 Contract History

            > **版本**: 1.0
            > **状态**: active
            > **更新日期**: 2026-04-28

            ## 1 修订记录

            | 日期 | 版本 | 变更 | 关联计划 |
            |------|------|------|----------|
            | 2026-04-28 | 1.0 | initial | - |
            """
        )
        history_path = repo / "docs" / "spec" / "openapi-v1-contract" / "history.md"
        history_path.write_text(history_text)

        # Initialize git so `git show HEAD:...history.md` works.
        env = os.environ.copy()
        env["GIT_AUTHOR_NAME"] = "test"
        env["GIT_AUTHOR_EMAIL"] = "test@example.com"
        env["GIT_COMMITTER_NAME"] = "test"
        env["GIT_COMMITTER_EMAIL"] = "test@example.com"
        subprocess.run(["git", "init", "-q", "-b", "main", str(repo)], check=True)
        subprocess.run(["git", "-C", str(repo), "add", "."], check=True)
        subprocess.run(
            ["git", "-C", str(repo), "commit", "-q", "-m", "init"],
            check=True,
            env=env,
        )
        return repo

    def _run(self, repo: Path, *extra_args: str) -> tuple[int, dict, str]:
        out = subprocess.run(
            [
                sys.executable,
                str(HERE / "openapi_diff.py"),
                "--repo-root",
                str(repo),
                "--fail-on-incompatible",
                *extra_args,
            ],
            capture_output=True,
            text=True,
        )
        try:
            payload = json.loads(out.stdout)
        except json.JSONDecodeError:
            payload = {"_raw": out.stdout}
        return out.returncode, payload, out.stderr

    def test_clean_baseline_passes(self) -> None:
        import tempfile
        with tempfile.TemporaryDirectory() as tmp:
            repo = self._write_repo(Path(tmp))
            rc, payload, stderr = self._run(repo)
            self.assertEqual(rc, 0, msg=payload)
            self.assertEqual(payload["summary"]["breaking"], 0)
            self.assertIn("wrapper-", stderr)

    def test_delete_field_fails_via_cli(self) -> None:
        import tempfile
        with tempfile.TemporaryDirectory() as tmp:
            repo = self._write_repo(Path(tmp))
            cur_path = repo / "openapi" / "openapi.yaml"
            doc = yaml.safe_load(cur_path.read_text())
            del doc["components"]["schemas"]["TargetJob"]["properties"]["title"]
            doc["components"]["schemas"]["TargetJob"]["required"] = ["id"]
            cur_path.write_text(yaml.safe_dump(doc, sort_keys=False))
            rc, payload, _ = self._run(repo)
            self.assertEqual(rc, 1)
            self.assertGreaterEqual(payload["summary"]["breaking"], 1)

    def test_optional_field_passes_via_cli(self) -> None:
        import tempfile
        with tempfile.TemporaryDirectory() as tmp:
            repo = self._write_repo(Path(tmp))
            cur_path = repo / "openapi" / "openapi.yaml"
            doc = yaml.safe_load(cur_path.read_text())
            doc["components"]["schemas"]["PracticePlan"]["properties"]["metadata"] = {"type": "object"}
            cur_path.write_text(yaml.safe_dump(doc, sort_keys=False))
            rc, payload, _ = self._run(repo)
            self.assertEqual(rc, 0, msg=payload)
            self.assertGreaterEqual(payload["summary"]["additive"], 1)

    def test_privacy_501_to_202_with_history_increment_passes(self) -> None:
        import tempfile
        with tempfile.TemporaryDirectory() as tmp:
            repo = self._write_repo(Path(tmp))
            cur_path = repo / "openapi" / "openapi.yaml"
            doc = yaml.safe_load(cur_path.read_text())
            doc["paths"]["/privacy/exports"]["post"]["responses"] = {
                "202": {
                    "content": {
                        "application/json": {
                            "schema": {"$ref": "#/components/schemas/PrivacyRequestWithJob"}
                        }
                    }
                }
            }
            cur_path.write_text(yaml.safe_dump(doc, sort_keys=False))
            history_path = repo / "docs" / "spec" / "openapi-v1-contract" / "history.md"
            history_text = history_path.read_text()
            history_text = history_text.replace(
                "| 2026-04-28 | 1.0 | initial | - |\n",
                "| 2026-04-28 | 1.0 | initial | - |\n| 2026-04-29 | 1.1 | privacy export 501→202 (whitelist additive) | - |\n",
            )
            history_path.write_text(history_text)
            rc, payload, _ = self._run(repo)
            self.assertEqual(rc, 0, msg=payload)
            self.assertEqual(payload["summary"]["breaking"], 0)
            kinds = {f["kind"]: f for f in payload["findings"]}
            self.assertIn("response-status-removed", kinds)
            self.assertIn("response-status-added", kinds)
            self.assertEqual(kinds["response-status-removed"]["severity"], "informational")
            self.assertEqual(kinds["response-status-added"]["severity"], "informational")

    def test_privacy_501_to_202_without_history_increment_fails(self) -> None:
        import tempfile
        with tempfile.TemporaryDirectory() as tmp:
            repo = self._write_repo(Path(tmp))
            cur_path = repo / "openapi" / "openapi.yaml"
            doc = yaml.safe_load(cur_path.read_text())
            doc["paths"]["/privacy/exports"]["post"]["responses"] = {
                "202": {
                    "content": {
                        "application/json": {
                            "schema": {"$ref": "#/components/schemas/PrivacyRequestWithJob"}
                        }
                    }
                }
            }
            cur_path.write_text(yaml.safe_dump(doc, sort_keys=False))
            rc, payload, _ = self._run(repo)
            self.assertEqual(rc, 1, msg=payload)
            kinds = {f["kind"] for f in payload["findings"]}
            self.assertIn("history-not-incremented", kinds)

    def test_privacy_501_to_202_committed_history_increment_passes_against_base_branch(self) -> None:
        import tempfile
        with tempfile.TemporaryDirectory() as tmp:
            repo = self._write_repo(Path(tmp))
            subprocess.run(["git", "-C", str(repo), "checkout", "-q", "-b", "feature/privacy-export"], check=True)

            cur_path = repo / "openapi" / "openapi.yaml"
            doc = yaml.safe_load(cur_path.read_text())
            doc["paths"]["/privacy/exports"]["post"]["responses"] = {
                "202": {
                    "content": {
                        "application/json": {
                            "schema": {"$ref": "#/components/schemas/PrivacyRequestWithJob"}
                        }
                    }
                }
            }
            cur_path.write_text(yaml.safe_dump(doc, sort_keys=False))

            history_path = repo / "docs" / "spec" / "openapi-v1-contract" / "history.md"
            history_text = history_path.read_text()
            history_text = history_text.replace(
                "| 2026-04-28 | 1.0 | initial | - |\n",
                "| 2026-04-28 | 1.0 | initial | - |\n| 2026-04-29 | 1.1 | privacy export 501→202 (whitelist additive) | - |\n",
            )
            history_path.write_text(history_text)

            env = os.environ.copy()
            env["GIT_AUTHOR_NAME"] = "test"
            env["GIT_AUTHOR_EMAIL"] = "test@example.com"
            env["GIT_COMMITTER_NAME"] = "test"
            env["GIT_COMMITTER_EMAIL"] = "test@example.com"
            subprocess.run(["git", "-C", str(repo), "add", "."], check=True)
            subprocess.run(
                ["git", "-C", str(repo), "commit", "-q", "-m", "privacy transition"],
                check=True,
                env=env,
            )

            rc, payload, _ = self._run(repo)
            self.assertEqual(rc, 0, msg=payload)
            self.assertEqual(payload["summary"]["breaking"], 0)
            self.assertEqual(payload["summary"]["informational"], 2)


if __name__ == "__main__":
    unittest.main()
