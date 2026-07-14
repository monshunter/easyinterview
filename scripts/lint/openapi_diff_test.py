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
import tempfile
import unittest
from pathlib import Path
from textwrap import dedent

import yaml

HERE = Path(__file__).resolve().parent
REPO_ROOT = HERE.parent.parent
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


def _openapi_002_docs(*, remove_source_errors: bool = True) -> tuple[dict, dict]:
    source_schemas = [
        "TargetJobImportSourceURL",
        "TargetJobImportSourceManualText",
        "TargetJobImportSourceFile",
        "TargetJobImportSourceManualForm",
        "TargetJobImportSource",
    ]
    target_required = [
        "id",
        "status",
        "analysisStatus",
        "title",
        "companyName",
        "targetLanguage",
        "sourceType",
        "requirements",
        "openQuestionIssueCount",
        "createdAt",
        "updatedAt",
    ]
    paths = {
        f"/fixture/{index}": {
            "get": {
                "operationId": f"fixtureOperation{index}",
                "responses": {"200": {"description": "ok"}},
            }
        }
        for index in range(35)
    }
    paths.update(
        {
            "/targets/import": {
                "post": {
                    "operationId": "importTargetJob",
                    "responses": {
                        "202": {
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/TargetJobWithJob"
                                    }
                                }
                            }
                        }
                    },
                }
            },
            "/uploads/presign": {
                "post": {
                    "operationId": "createUploadPresign",
                    "responses": {
                        "201": {
                            "content": {
                                "application/json": {
                                    "schema": {"$ref": "#/components/schemas/UploadPresign"}
                                }
                            }
                        }
                    },
                }
            },
        }
    )
    schemas = {name: {"type": "object"} for name in source_schemas}
    schemas.update(
        {
            "ImportTargetJobRequest": {
                "type": "object",
                "required": ["source", "targetLanguage", "resumeId"],
                "properties": {
                    "source": {"$ref": "#/components/schemas/TargetJobImportSource"},
                    "targetLanguage": {"type": "string"},
                    "resumeId": {"type": "string", "format": "uuid"},
                    "titleHint": {"type": "string"},
                    "companyNameHint": {"type": "string"},
                },
            },
            "TargetJob": {
                "type": "object",
                "required": target_required,
                "properties": {
                    "sourceType": {
                        "type": "string",
                        "enum": ["manual_text", "url", "file", "manual_form"],
                    },
                    "sourceUrl": {"oneOf": [{"type": "string"}, {"type": "null"}]},
                },
            },
            "UploadPresignRequest": {
                "type": "object",
                "properties": {
                    "purpose": {
                        "type": "string",
                        "enum": ["resume", "target_job_attachment", "privacy_export"],
                    }
                },
            },
            "ApiErrorCode": {
                "type": "string",
                "enum": [
                    "VALIDATION_FAILED",
                    "TARGET_IMPORT_FAILED",
                    "TARGET_IMPORT_SOURCE_INVALID",
                    "TARGET_IMPORT_SOURCE_UNAVAILABLE",
                ],
            },
        }
    )
    baseline = {
        "openapi": "3.1.0",
        "servers": [{"url": "/api/v1"}],
        "tags": [{"name": f"Tag{index}"} for index in range(10)],
        "paths": paths,
        "components": {"schemas": schemas},
    }
    current = copy.deepcopy(baseline)
    current_schemas = current["components"]["schemas"]
    for name in source_schemas:
        del current_schemas[name]
    request = current_schemas["ImportTargetJobRequest"]
    request["additionalProperties"] = False
    request["required"] = ["rawText", "targetLanguage", "resumeId"]
    del request["properties"]["source"]
    del request["properties"]["titleHint"]
    del request["properties"]["companyNameHint"]
    request["properties"]["rawText"] = {
        "type": "string",
        "minLength": 1,
        "pattern": r"\S",
    }
    target = current_schemas["TargetJob"]
    del target["properties"]["sourceType"]
    del target["properties"]["sourceUrl"]
    target["required"] = [value for value in target_required if value != "sourceType"]
    current_schemas["UploadPresignRequest"]["properties"]["purpose"]["enum"] = [
        "resume",
        "privacy_export",
    ]
    if remove_source_errors:
        current_schemas["ApiErrorCode"]["enum"] = [
            "VALIDATION_FAILED",
            "TARGET_IMPORT_FAILED",
        ]
    return baseline, current


def _d_35_docs() -> tuple[dict, dict]:
    base_message = {
        "type": "object",
        "required": ["id", "seqNo", "role", "content", "createdAt"],
        "properties": {
            "id": {"type": "string", "format": "uuid"},
            "seqNo": {"type": "integer", "format": "int32", "minimum": 1},
            "role": {"type": "string", "enum": ["user", "assistant"]},
            "content": {"type": "string"},
            "createdAt": {"type": "string", "format": "date-time"},
        },
    }
    paths = {
        f"/fixture/{index}": {
            "get": {
                "operationId": f"fixtureOperation{index}",
                "responses": {"200": {"description": "ok"}},
            }
        }
        for index in range(35)
    }
    paths.update(
        {
            "/practice/sessions/{sessionId}": {
                "get": {
                    "operationId": "getPracticeSession",
                    "responses": {
                        "200": {
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/PracticeSession"
                                    }
                                }
                            }
                        }
                    },
                }
            },
            "/practice/sessions/{sessionId}/messages": {
                "post": {
                    "operationId": "sendPracticeMessage",
                    "responses": {
                        "200": {
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/SendPracticeMessageResponse"
                                    }
                                }
                            }
                        }
                    },
                }
            },
        }
    )
    baseline = {
        "openapi": "3.1.0",
        "servers": [{"url": "/api/v1"}],
        "tags": [{"name": f"Tag{index}"} for index in range(10)],
        "paths": paths,
        "components": {
            "schemas": {
                "PracticeMessage": base_message,
                "PracticeSession": {
                    "type": "object",
                    "required": ["id", "messages"],
                    "properties": {
                        "id": {"type": "string", "format": "uuid"},
                        "messages": {
                            "type": "array",
                            "items": {"$ref": "#/components/schemas/PracticeMessage"},
                        },
                    },
                },
                "SendPracticeMessageResponse": {
                    "type": "object",
                    "required": [
                        "acknowledged",
                        "session",
                        "userMessage",
                        "assistantMessage",
                    ],
                    "properties": {
                        "acknowledged": {"type": "boolean"},
                        "session": {"$ref": "#/components/schemas/PracticeSession"},
                        "userMessage": {"$ref": "#/components/schemas/PracticeMessage"},
                        "assistantMessage": {"$ref": "#/components/schemas/PracticeMessage"},
                    },
                },
            }
        },
    }
    current = copy.deepcopy(baseline)
    schemas = current["components"]["schemas"]
    base_required = list(base_message["required"])
    schemas["PracticeMessage"] = {
        "oneOf": [
            {"$ref": "#/components/schemas/PracticeUserMessage"},
            {"$ref": "#/components/schemas/PracticeAssistantMessage"},
        ],
        "discriminator": {
            "propertyName": "role",
            "mapping": {
                "user": "#/components/schemas/PracticeUserMessage",
                "assistant": "#/components/schemas/PracticeAssistantMessage",
            },
        },
    }
    schemas["PracticeReplyStatus"] = {
        "type": "string",
        "enum": ["pending", "retryable_failed", "terminal_failed", "complete"],
    }
    schemas["PracticeUserMessage"] = {
        "type": "object",
        "additionalProperties": False,
        "required": base_required + ["clientMessageId", "replyStatus"],
        "properties": {
            **copy.deepcopy(base_message["properties"]),
            "role": {"type": "string", "enum": ["user"]},
            "clientMessageId": {"type": "string", "format": "uuid"},
            "replyStatus": {"$ref": "#/components/schemas/PracticeReplyStatus"},
        },
    }
    schemas["PracticeAssistantMessage"] = {
        "type": "object",
        "additionalProperties": False,
        "required": base_required,
        "properties": {
            **copy.deepcopy(base_message["properties"]),
            "role": {"type": "string", "enum": ["assistant"]},
        },
    }
    response = schemas["SendPracticeMessageResponse"]["properties"]
    response["userMessage"] = {"$ref": "#/components/schemas/PracticeUserMessage"}
    response["assistantMessage"] = {
        "$ref": "#/components/schemas/PracticeAssistantMessage"
    }
    return baseline, current


def _openapi_004_docs() -> tuple[dict, dict]:
    report_path = "/targets/{targetJobId}/reports"
    paths = {
        f"/fixture/{index}": {
            "get": {
                "operationId": f"fixtureOperation{index}",
                "responses": {"200": {"description": "ok"}},
            }
        }
        for index in range(36)
    }
    paths[report_path] = {
        "get": {
            "operationId": "listTargetJobReports",
            "parameters": [
                {
                    "name": "targetJobId",
                    "in": "path",
                    "required": True,
                    "schema": {"type": "string", "format": "uuid"},
                },
                {
                    "name": "cursor",
                    "in": "query",
                    "required": False,
                    "schema": {"type": "string"},
                },
                {
                    "name": "pageSize",
                    "in": "query",
                    "required": False,
                    "schema": {
                        "type": "integer",
                        "format": "int32",
                        "minimum": 1,
                        "maximum": 100,
                    },
                },
                {"$ref": "#/components/parameters/XRequestID"},
                {"$ref": "#/components/parameters/Traceparent"},
                {"$ref": "#/components/parameters/AcceptLanguage"},
                {"$ref": "#/components/parameters/XClientVersion"},
            ],
            "responses": {
                "200": {
                    "content": {
                        "application/json": {
                            "schema": {
                                "$ref": "#/components/schemas/PaginatedFeedbackReport"
                            }
                        }
                    }
                },
                "default": {"description": "error"},
            },
        }
    }
    baseline = {
        "openapi": "3.1.0",
        "servers": [{"url": "/api/v1"}],
        "tags": [{"name": f"Tag{index}"} for index in range(10)],
        "paths": paths,
        "components": {
            "schemas": {
                "TargetJob": {
                    "type": "object",
                    "required": ["id"],
                    "properties": {
                        "id": {"type": "string", "format": "uuid"},
                        "latestReportId": {
                            "oneOf": [
                                {"type": "string", "format": "uuid"},
                                {"type": "null"},
                            ]
                        },
                    },
                },
                "PracticeRoundRef": {
                    "type": "object",
                    "required": ["roundId", "roundSequence"],
                    "properties": {
                        "roundId": {
                            "type": "string",
                            "pattern": "^round-[1-9][0-9]{0,9}-(hr|technical)$",
                        },
                        "roundSequence": {
                            "type": "integer",
                            "format": "int32",
                            "minimum": 1,
                            "maximum": 2147483647,
                        },
                    },
                },
                "PaginatedFeedbackReport": {
                    "type": "object",
                    "required": ["items"],
                    "properties": {
                        "items": {
                            "type": "array",
                            "items": {"$ref": "#/components/schemas/FeedbackReport"},
                        }
                    },
                },
                "FeedbackReport": {"type": "object"},
                "ApiErrorCode": {
                    "type": "string",
                    "enum": ["AI_PROVIDER_TIMEOUT"],
                },
                "ReportStatus": {
                    "type": "string",
                    "enum": ["queued", "generating", "ready", "failed"],
                },
            }
        },
    }
    current = copy.deepcopy(baseline)
    operation = current["paths"][report_path]["get"]
    operation["parameters"] = [
        parameter
        for parameter in operation["parameters"]
        if parameter.get("name") not in {"cursor", "pageSize"}
    ]
    operation["responses"]["200"]["content"]["application/json"]["schema"] = {
        "$ref": "#/components/schemas/TargetJobReportsOverview"
    }
    schemas = current["components"]["schemas"]
    del schemas["TargetJob"]["properties"]["latestReportId"]
    del schemas["PaginatedFeedbackReport"]
    schemas["PracticeRoundRef"]["additionalProperties"] = False
    schemas.update(
        {
            "TargetJobReportsOverview": {
                "type": "object",
                "required": ["targetJobId", "rounds"],
                "additionalProperties": False,
                "properties": {
                    "targetJobId": {"type": "string", "format": "uuid"},
                    "rounds": {
                        "type": "array",
                        "minItems": 2,
                        "maxItems": 5,
                        "items": {
                            "$ref": "#/components/schemas/TargetJobReportRoundOverview"
                        },
                    },
                },
            },
            "TargetJobReportRoundOverview": {
                "type": "object",
                "required": ["round", "currentReport", "latestAttempt"],
                "additionalProperties": False,
                "properties": {
                    "round": {"$ref": "#/components/schemas/PracticeRoundRef"},
                    "currentReport": {
                        "oneOf": [
                            {
                                "$ref": "#/components/schemas/TargetJobCurrentReportSummary"
                            },
                            {"type": "null"},
                        ]
                    },
                    "latestAttempt": {
                        "oneOf": [
                            {
                                "$ref": "#/components/schemas/TargetJobReportAttemptSummary"
                            },
                            {"type": "null"},
                        ]
                    },
                },
            },
            "TargetJobCurrentReportSummary": {
                "type": "object",
                "required": ["id", "generatedAt"],
                "additionalProperties": False,
                "properties": {
                    "id": {"type": "string", "format": "uuid"},
                    "generatedAt": {"type": "string", "format": "date-time"},
                },
            },
            "TargetJobReportAttemptSummary": {
                "type": "object",
                "required": ["id", "status", "errorCode", "createdAt"],
                "additionalProperties": False,
                "allOf": [
                    {
                        "if": {
                            "required": ["status"],
                            "properties": {"status": {"const": "failed"}},
                        },
                        "then": {
                            "properties": {
                                "errorCode": {
                                    "$ref": "#/components/schemas/ApiErrorCode"
                                }
                            }
                        },
                        "else": {"properties": {"errorCode": {"type": "null"}}},
                    }
                ],
                "properties": {
                    "id": {"type": "string", "format": "uuid"},
                    "status": {"$ref": "#/components/schemas/ReportStatus"},
                    "errorCode": {
                        "oneOf": [
                            {"$ref": "#/components/schemas/ApiErrorCode"},
                            {"type": "null"},
                        ]
                    },
                    "createdAt": {"type": "string", "format": "date-time"},
                },
            },
        }
    )
    return baseline, current


def _openapi_004_oracle() -> dict:
    report_pointer = "/paths/~1targets~1{targetJobId}~1reports/get"
    findings = [
        {
            "severity": "breaking",
            "path": f"{report_pointer}/parameters/cursor",
            "kind": "parameter_removed",
            "before": "query:optional",
            "after": "absent",
        },
        {
            "severity": "breaking",
            "path": f"{report_pointer}/parameters/pageSize",
            "kind": "parameter_removed",
            "before": "query:optional",
            "after": "absent",
        },
        {
            "severity": "breaking",
            "path": f"{report_pointer}/responses/200/content/application~1json/schema",
            "kind": "response_ref_changed",
            "before": "PaginatedFeedbackReport",
            "after": "TargetJobReportsOverview",
        },
        {
            "severity": "breaking",
            "path": "/components/schemas/TargetJob/properties/latestReportId",
            "kind": "property_removed",
            "before": "string|null",
            "after": "absent",
        },
        {
            "severity": "breaking",
            "path": "/components/schemas/PaginatedFeedbackReport",
            "kind": "schema_removed",
            "before": "present",
            "after": "absent",
        },
        {
            "severity": "breaking",
            "path": "/components/schemas/PracticeRoundRef/additionalProperties",
            "kind": "closed_object",
            "before": "unspecified",
            "after": False,
        },
        {
            "severity": "additive",
            "path": "/components/schemas/TargetJobReportsOverview",
            "kind": "schema_added_with_required_fields",
            "before": "absent",
            "after": "targetJobId,rounds",
        },
        {
            "severity": "additive",
            "path": "/components/schemas/TargetJobReportsOverview/additionalProperties",
            "kind": "closed_object",
            "before": "absent",
            "after": False,
        },
        {
            "severity": "additive",
            "path": "/components/schemas/TargetJobReportsOverview/properties/rounds",
            "kind": "canonical_array_bounds_added",
            "before": "absent",
            "after": "minItems=2,maxItems=5",
        },
        {
            "severity": "additive",
            "path": "/components/schemas/TargetJobReportRoundOverview",
            "kind": "schema_added_with_required_fields",
            "before": "absent",
            "after": "round,currentReport,latestAttempt",
        },
        {
            "severity": "additive",
            "path": "/components/schemas/TargetJobReportRoundOverview/additionalProperties",
            "kind": "closed_object",
            "before": "absent",
            "after": False,
        },
        {
            "severity": "additive",
            "path": "/components/schemas/TargetJobCurrentReportSummary",
            "kind": "schema_added_with_required_fields",
            "before": "absent",
            "after": "id,generatedAt",
        },
        {
            "severity": "additive",
            "path": "/components/schemas/TargetJobCurrentReportSummary/additionalProperties",
            "kind": "closed_object",
            "before": "absent",
            "after": False,
        },
        {
            "severity": "additive",
            "path": "/components/schemas/TargetJobReportAttemptSummary",
            "kind": "schema_added_with_required_fields",
            "before": "absent",
            "after": "id,status,errorCode,createdAt",
        },
        {
            "severity": "additive",
            "path": "/components/schemas/TargetJobReportAttemptSummary/additionalProperties",
            "kind": "closed_object",
            "before": "absent",
            "after": False,
        },
    ]
    return {
        "schemaVersion": 1,
        "decisionId": "OPENAPI-004",
        "authority": {
            "decision": "OPENAPI-004",
            "specDecision": "D-36",
            "historyVersion": "1.57",
            "productDecision": "R-A",
        },
        "invariants": {
            "inventory": {"operations": 37, "tags": 10},
            "listTargetJobReports": {
                "method": "GET",
                "path": "/api/v1/targets/{targetJobId}/reports",
                "operationId": "listTargetJobReports",
                "successStatus": 200,
                "baselineResponse": "PaginatedFeedbackReport",
                "response": "TargetJobReportsOverview",
            },
            "canonicalRounds": {"minItems": 2, "maxItems": 5},
            "compatibilityAliases": "forbidden",
        },
        "comparison": {
            "mode": "exact-set",
            "keyFields": ["severity", "path", "kind", "before", "after"],
            "orderSignificant": False,
            "missingFinding": "fail",
            "unexpectedFinding": "fail",
        },
        "findings": findings,
    }


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
        cur["tags"].append({"name": "Insights"})
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

    def test_operation_count_counts_operation_ids(self) -> None:
        self.assertEqual(od._operation_count(_baseline_doc()), 3)


class OpenAPI001OracleTests(unittest.TestCase):
    def test_exact_set_rejects_missing_unexpected_and_severity_drift(self) -> None:
        expected = [
            {
                "severity": "breaking",
                "path": "/components/schemas/FeedbackReport/properties/context",
                "kind": "required_property_added",
                "before": "absent",
                "after": "ReportContextSnapshot",
            }
        ]
        self.assertEqual([], od.compare_finding_sets(expected, copy.deepcopy(expected)))

        missing = od.compare_finding_sets(expected, [])
        self.assertTrue(any("missing" in error for error in missing), missing)

        unexpected_finding = {**expected[0], "path": "/unexpected"}
        unexpected = od.compare_finding_sets(expected, [*expected, unexpected_finding])
        self.assertTrue(any("unexpected" in error for error in unexpected), unexpected)

        drifted = [{**expected[0], "severity": "additive"}]
        severity = od.compare_finding_sets(expected, drifted)
        self.assertTrue(any("missing" in error for error in severity), severity)
        self.assertTrue(any("unexpected" in error for error in severity), severity)

    def test_decision_record_requires_matching_accepted_id(self) -> None:
        accepted = "# Decision\n\n> **ID**: OPENAPI-001\n> **状态**: accepted\n"
        self.assertEqual([], od.validate_decision_record(accepted, "OPENAPI-001"))
        self.assertTrue(od.validate_decision_record(accepted.replace("accepted", "draft"), "OPENAPI-001"))
        self.assertTrue(od.validate_decision_record(accepted, "OPENAPI-999"))

    def test_normalized_conditional_and_error_enum_findings(self) -> None:
        baseline = {
            "components": {
                "schemas": {
                    "CreatePracticePlanRequest": {
                        "type": "object",
                        "required": [
                            "targetJobId",
                            "goal",
                            "interviewerPersona",
                            "difficulty",
                            "language",
                            "timeBudgetMinutes",
                            "resumeId",
                        ],
                        "properties": {
                            "goal": {"type": "string"},
                            "targetJobId": {"type": "string"},
                        },
                    },
                    "ApiErrorCode": {"type": "string", "enum": ["REPORT_NOT_READY"]},
                }
            }
        }
        current = copy.deepcopy(baseline)
        request = current["components"]["schemas"]["CreatePracticePlanRequest"]
        request["required"] = ["goal"]
        request["additionalProperties"] = False
        request["properties"]["sourceReportId"] = {"type": "string", "format": "uuid"}
        request["oneOf"] = [
            {"properties": {"goal": {"const": "baseline"}}, "not": {"required": ["sourceReportId"]}},
            {
                "required": ["goal", "sourceReportId"],
                "properties": {"goal": {"enum": ["retry_current_round", "next_round"]}},
            },
        ]
        current["components"]["schemas"]["ApiErrorCode"]["enum"].append(
            "REPORT_CONTEXT_TOO_LARGE"
        )

        findings = od.normalize_openapi_001_findings(baseline, current)

        self.assertIn(
            {
                "severity": "breaking",
                "path": "/components/schemas/CreatePracticePlanRequest/oneOf",
                "kind": "conditional_contract_added",
                "before": "absent",
                "after": "baseline-required-fields-sourceReportId-forbidden|derived-retry-next-sourceReportId-required-nonnull-only",
            },
            findings,
        )


class OpenAPI002OracleTests(unittest.TestCase):
    @staticmethod
    def _oracle() -> dict:
        return json.loads(
            (
                REPO_ROOT
                / "docs/spec/openapi-v1-contract/decisions/OPENAPI-002-targetjob-paste-only.expected-findings.json"
            ).read_text(encoding="utf-8")
        )

    def _write_repo(self, repo: Path, *, remove_source_errors: bool) -> tuple[Path, Path, Path]:
        baseline, current = _openapi_002_docs(
            remove_source_errors=remove_source_errors
        )
        (repo / "openapi" / "baseline").mkdir(parents=True)
        decision_dir = repo / "docs" / "spec" / "openapi-v1-contract" / "decisions"
        decision_dir.mkdir(parents=True)
        baseline_path = repo / "openapi" / "baseline" / "openapi-v1.0.0.yaml"
        current_path = repo / "openapi" / "openapi.yaml"
        decision_path = decision_dir / "OPENAPI-002-targetjob-paste-only.md"
        oracle_path = decision_dir / "OPENAPI-002-targetjob-paste-only.expected-findings.json"
        baseline_path.write_text(yaml.safe_dump(baseline, sort_keys=False), encoding="utf-8")
        current_path.write_text(yaml.safe_dump(baseline, sort_keys=False), encoding="utf-8")
        decision_path.write_text(
            "# OPENAPI-002\n\n> **ID**: OPENAPI-002\n> **状态**: accepted\n",
            encoding="utf-8",
        )
        oracle_path.write_text(
            json.dumps(self._oracle(), ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )

        env = os.environ.copy()
        env.update(
            {
                "GIT_AUTHOR_NAME": "test",
                "GIT_AUTHOR_EMAIL": "test@example.com",
                "GIT_COMMITTER_NAME": "test",
                "GIT_COMMITTER_EMAIL": "test@example.com",
            }
        )
        subprocess.run(["git", "init", "-q", "-b", "main", str(repo)], check=True)
        subprocess.run(["git", "-C", str(repo), "add", "."], check=True)
        subprocess.run(
            ["git", "-C", str(repo), "commit", "-q", "-m", "old baseline"],
            check=True,
            env=env,
        )
        current_path.write_text(yaml.safe_dump(current, sort_keys=False), encoding="utf-8")
        return decision_path, oracle_path, current_path

    def _run_audit(
        self, repo: Path, decision_path: Path, oracle_path: Path, artifact_path: Path
    ) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [
                sys.executable,
                str(HERE / "openapi_diff.py"),
                "--repo-root",
                str(repo),
                "--decision-id",
                "OPENAPI-002",
                "--decision-record",
                str(decision_path),
                "--oracle",
                str(oracle_path),
                "--base-ref",
                "main",
                "--output",
                str(artifact_path),
            ],
            capture_output=True,
            text=True,
        )

    def test_raw_text_constraints_fold_into_one_required_property_finding(self) -> None:
        baseline, current = _openapi_002_docs()
        findings = od.normalize_openapi_002_findings(baseline, current)
        raw_text_findings = [
            finding
            for finding in findings
            if finding["path"]
            == "/components/schemas/ImportTargetJobRequest/properties/rawText"
        ]
        self.assertEqual(
            [
                {
                    "severity": "breaking",
                    "path": "/components/schemas/ImportTargetJobRequest/properties/rawText",
                    "kind": "required_property_added",
                    "before": "absent",
                    "after": r"string(minLength=1,pattern=\S)",
                }
            ],
            raw_text_findings,
        )
        self.assertEqual(17, len(findings))
        comparison_errors = od.compare_finding_sets(self._oracle()["findings"], findings)
        self.assertEqual([], comparison_errors)

    def test_report_pointer_removal_does_not_enter_openapi_002_manifest(self) -> None:
        baseline, current = _openapi_002_docs()
        baseline["components"]["schemas"]["TargetJob"]["properties"][
            "latestReportId"
        ] = {
            "oneOf": [
                {"type": "string", "format": "uuid"},
                {"type": "null"},
            ]
        }

        findings = od.normalize_openapi_002_findings(baseline, current)

        self.assertEqual(17, len(findings))
        self.assertEqual([], od.compare_finding_sets(self._oracle()["findings"], findings))

    def test_invariants_reject_inventory_operation_response_and_purpose_drift(self) -> None:
        _, current = _openapi_002_docs(remove_source_errors=False)
        invariants = self._oracle()["invariants"]
        self.assertEqual([], od.validate_openapi_002_invariants(current, invariants))

        drifted = copy.deepcopy(current)
        drifted["tags"].pop()
        del drifted["paths"]["/fixture/0"]
        drifted["paths"]["/targets/import"]["post"]["operationId"] = "renamedImport"
        drifted["paths"]["/uploads/presign"]["post"]["responses"]["201"][
            "content"
        ]["application/json"]["schema"]["$ref"] = "#/components/schemas/WrongResponse"
        drifted["components"]["schemas"]["UploadPresignRequest"]["properties"][
            "purpose"
        ]["enum"] = ["resume"]

        errors = od.validate_openapi_002_invariants(drifted, invariants)
        self.assertTrue(any("operations must equal 37" in error for error in errors), errors)
        self.assertTrue(any("tags must equal 10" in error for error in errors), errors)
        self.assertTrue(any("importTargetJob operationId" in error for error in errors), errors)
        self.assertTrue(any("createUploadPresign response" in error for error in errors), errors)
        self.assertTrue(any("remaining purposes" in error for error in errors), errors)

    def test_cli_exact_17_passes_and_writes_artifact(self) -> None:
        import tempfile

        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            decision_path, oracle_path, _ = self._write_repo(
                repo, remove_source_errors=True
            )
            artifact_path = repo / "audit.json"
            result = self._run_audit(repo, decision_path, oracle_path, artifact_path)
            self.assertEqual(0, result.returncode, result.stderr or result.stdout)
            payload = json.loads(result.stdout)
            self.assertEqual(17, payload["expectedFindingCount"])
            self.assertEqual(17, payload["findingCount"])
            self.assertEqual([], payload["errors"])
            self.assertEqual(payload, json.loads(artifact_path.read_text(encoding="utf-8")))

    def test_cli_stale_exact_15_vs_actual_17_fails_but_writes_artifact(self) -> None:
        import tempfile

        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            decision_path, oracle_path, _ = self._write_repo(
                repo, remove_source_errors=True
            )
            stale_oracle = self._oracle()
            stale_oracle["findings"] = [
                finding
                for finding in stale_oracle["findings"]
                if not (
                    finding["path"] == "/components/schemas/ApiErrorCode/enum"
                    and finding["before"]
                    in {
                        "TARGET_IMPORT_SOURCE_INVALID",
                        "TARGET_IMPORT_SOURCE_UNAVAILABLE",
                    }
                )
            ]
            oracle_path.write_text(
                json.dumps(stale_oracle, ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
            artifact_path = repo / "audit.json"
            result = self._run_audit(repo, decision_path, oracle_path, artifact_path)
            self.assertEqual(1, result.returncode, result.stderr or result.stdout)
            payload = json.loads(result.stdout)
            self.assertEqual(15, payload["expectedFindingCount"])
            self.assertEqual(17, payload["findingCount"])
            self.assertTrue(
                any(
                    "expected 15 findings but actual 17" in error
                    for error in payload["errors"]
                ),
                payload,
            )
            self.assertTrue(artifact_path.is_file(), payload)
            self.assertEqual(payload, json.loads(artifact_path.read_text(encoding="utf-8")))


class D35OracleTests(unittest.TestCase):
    @staticmethod
    def _oracle() -> dict:
        return {
            "schemaVersion": 1,
            "decisionId": "D-35",
            "authority": {
                "specDecision": "D-35",
                "historyVersion": "1.54",
                "productDecision": "方案 A",
            },
            "invariants": {
                "inventory": {"operations": 37, "tags": 10},
                "getPracticeSession": {
                    "method": "GET",
                    "path": "/api/v1/practice/sessions/{sessionId}",
                    "operationId": "getPracticeSession",
                    "successStatus": 200,
                    "response": "PracticeSession",
                },
                "sendPracticeMessage": {
                    "method": "POST",
                    "path": "/api/v1/practice/sessions/{sessionId}/messages",
                    "operationId": "sendPracticeMessage",
                    "successStatus": 200,
                    "response": "SendPracticeMessageResponse",
                },
                "retryEndpoint": "forbidden",
            },
            "comparison": {
                "mode": "exact-set",
                "keyFields": ["severity", "path", "kind", "before", "after"],
                "orderSignificant": False,
                "missingFinding": "fail",
                "unexpectedFinding": "fail",
            },
            "findings": [
                {
                    "severity": "breaking",
                    "path": "/components/schemas/PracticeMessage/oneOf",
                    "kind": "role_discriminated_union_added",
                    "before": "absent",
                    "after": "user=PracticeUserMessage,assistant=PracticeAssistantMessage;discriminator=role",
                },
                {
                    "severity": "breaking",
                    "path": "/components/schemas/PracticeUserMessage/properties/clientMessageId",
                    "kind": "required_property_added",
                    "before": "absent",
                    "after": "string",
                },
                {
                    "severity": "breaking",
                    "path": "/components/schemas/PracticeUserMessage/properties/replyStatus",
                    "kind": "required_property_added",
                    "before": "absent",
                    "after": "PracticeReplyStatus",
                },
                {
                    "severity": "breaking",
                    "path": "/components/schemas/PracticeUserMessage/required",
                    "kind": "required_set_changed",
                    "before": "id,seqNo,role,content,createdAt",
                    "after": "id,seqNo,role,content,createdAt,clientMessageId,replyStatus",
                },
                {
                    "severity": "breaking",
                    "path": "/components/schemas/PracticeUserMessage/additionalProperties",
                    "kind": "closed_object",
                    "before": "unspecified",
                    "after": False,
                },
                {
                    "severity": "breaking",
                    "path": "/components/schemas/PracticeAssistantMessage/additionalProperties",
                    "kind": "closed_object",
                    "before": "unspecified",
                    "after": False,
                },
                {
                    "severity": "breaking",
                    "path": "/components/schemas/SendPracticeMessageResponse/properties/userMessage",
                    "kind": "ref_changed",
                    "before": "PracticeMessage",
                    "after": "PracticeUserMessage",
                },
                {
                    "severity": "breaking",
                    "path": "/components/schemas/SendPracticeMessageResponse/properties/assistantMessage",
                    "kind": "ref_changed",
                    "before": "PracticeMessage",
                    "after": "PracticeAssistantMessage",
                },
                {
                    "severity": "additive",
                    "path": "/components/schemas/PracticeReplyStatus",
                    "kind": "schema_added",
                    "before": "absent",
                    "after": "enum(pending,retryable_failed,terminal_failed,complete)",
                },
                {
                    "severity": "additive",
                    "path": "/components/schemas/PracticeUserMessage",
                    "kind": "schema_added_with_required_fields",
                    "before": "absent",
                    "after": "id,seqNo,role,content,createdAt,clientMessageId,replyStatus",
                },
                {
                    "severity": "additive",
                    "path": "/components/schemas/PracticeAssistantMessage",
                    "kind": "schema_added_with_required_fields",
                    "before": "absent",
                    "after": "id,seqNo,role,content,createdAt",
                },
            ],
        }

    def _write_repo(self, repo: Path) -> tuple[Path, Path]:
        baseline, current = _d_35_docs()
        (repo / "openapi" / "baseline").mkdir(parents=True)
        authority_dir = repo / "docs" / "spec" / "openapi-v1-contract"
        decision_dir = authority_dir / "decisions"
        decision_dir.mkdir(parents=True)
        baseline_path = repo / "openapi" / "baseline" / "openapi-v1.0.0.yaml"
        current_path = repo / "openapi" / "openapi.yaml"
        oracle_path = (
            decision_dir
            / "D-35-practice-durable-recovery.expected-findings.json"
        )
        baseline_path.write_text(yaml.safe_dump(baseline, sort_keys=False), encoding="utf-8")
        current_path.write_text(yaml.safe_dump(baseline, sort_keys=False), encoding="utf-8")
        (authority_dir / "spec.md").write_text(
            "| D-35 | Practice message durable recovery（方案 A） |\n",
            encoding="utf-8",
        )
        (authority_dir / "history.md").write_text(
            "| 2026-07-13 | 1.54 | 按方案 A 增加 Practice durable reply status |\n",
            encoding="utf-8",
        )
        oracle_path.write_text(
            json.dumps(self._oracle(), ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )

        env = os.environ.copy()
        env.update(
            {
                "GIT_AUTHOR_NAME": "test",
                "GIT_AUTHOR_EMAIL": "test@example.com",
                "GIT_COMMITTER_NAME": "test",
                "GIT_COMMITTER_EMAIL": "test@example.com",
            }
        )
        subprocess.run(["git", "init", "-q", "-b", "main", str(repo)], check=True)
        subprocess.run(["git", "-C", str(repo), "add", "."], check=True)
        subprocess.run(
            ["git", "-C", str(repo), "commit", "-q", "-m", "old baseline"],
            check=True,
            env=env,
        )
        current_path.write_text(yaml.safe_dump(current, sort_keys=False), encoding="utf-8")
        return oracle_path, current_path

    def _run_audit(
        self, repo: Path, oracle_path: Path, artifact_path: Path
    ) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [
                sys.executable,
                str(HERE / "openapi_diff.py"),
                "--repo-root",
                str(repo),
                "--decision-id",
                "D-35",
                "--oracle",
                str(oracle_path),
                "--base-ref",
                "main",
                "--output",
                str(artifact_path),
            ],
            capture_output=True,
            text=True,
        )

    def test_normalizer_exactly_separates_eight_breaking_and_three_additive_findings(
        self,
    ) -> None:
        baseline, current = _d_35_docs()
        findings = od.normalize_d_35_findings(baseline, current)

        self.assertEqual([], od.compare_finding_sets(self._oracle()["findings"], findings))
        self.assertEqual(11, len(findings))
        self.assertEqual(8, sum(finding["severity"] == "breaking" for finding in findings))
        self.assertEqual(3, sum(finding["severity"] == "additive" for finding in findings))
        self.assertEqual([], od.validate_d_35_contract(current, baseline))
        self.assertEqual(
            [],
            od.validate_d_35_invariants(
                baseline, current, self._oracle()["invariants"]
            ),
        )

    def test_contract_and_invariants_reject_silent_drift_matrix(self) -> None:
        mutations = {
            "extra user property": lambda current: current["components"]["schemas"][
                "PracticeUserMessage"
            ]["properties"].update({"legacyCompat": {"type": "string"}}),
            "legacy union keywords": lambda current: current["components"]["schemas"][
                "PracticeMessage"
            ].update({"type": "object", "properties": {}, "required": []}),
            "assistant base field drift": lambda current: current["components"][
                "schemas"
            ]["PracticeAssistantMessage"]["properties"].update(
                {"content": {"type": "integer"}}
            ),
            "session projection bypass": lambda current: current["components"][
                "schemas"
            ]["PracticeSession"]["properties"]["messages"].update(
                {"items": {"$ref": "#/components/schemas/PracticeUserMessage"}}
            ),
            "send response required drift": lambda current: current["components"][
                "schemas"
            ]["SendPracticeMessageResponse"].update(
                {"required": ["session", "userMessage", "assistantMessage"]}
            ),
            "retry endpoint": lambda current: current["paths"].update(
                {
                    "/practice/sessions/{sessionId}/messages/retry": {
                        "post": {
                            "operationId": "retryPracticeMessage",
                            "responses": {"200": {"description": "unexpected"}},
                        }
                    }
                }
            ),
        }
        for name, mutate in mutations.items():
            with self.subTest(name=name):
                baseline, current = _d_35_docs()
                mutate(current)
                findings = od.normalize_d_35_findings(baseline, current)
                exact_errors = od.compare_finding_sets(
                    self._oracle()["findings"], findings
                )
                invariant_errors = od.validate_d_35_contract(current, baseline)
                invariant_errors.extend(
                    od.validate_d_35_invariants(
                        baseline, current, self._oracle()["invariants"]
                    )
                )
                self.assertTrue(exact_errors, name)
                self.assertTrue(invariant_errors, name)

    def test_cli_requires_authority_exact_manifest_and_preserves_artifact(self) -> None:
        import tempfile

        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            oracle_path, _ = self._write_repo(repo)
            artifact_path = repo / "audit.json"

            result = self._run_audit(repo, oracle_path, artifact_path)

            self.assertEqual(0, result.returncode, result.stderr or result.stdout)
            payload = json.loads(result.stdout)
            self.assertEqual(11, payload["expectedFindingCount"])
            self.assertEqual(11, payload["findingCount"])
            self.assertEqual({"breaking": 8, "additive": 3, "informational": 0}, payload["summary"])
            self.assertEqual([], payload["errors"])
            self.assertEqual(payload, json.loads(artifact_path.read_text(encoding="utf-8")))

            (repo / "docs/spec/openapi-v1-contract/spec.md").write_text(
                "| D-34 | unrelated |\n", encoding="utf-8"
            )
            failed = self._run_audit(repo, oracle_path, artifact_path)
            self.assertEqual(1, failed.returncode)
            self.assertTrue(
                any("spec D-35" in error for error in json.loads(failed.stdout)["errors"])
            )

    def test_cli_rejects_missing_extra_and_wildcard_authorization(self) -> None:
        import tempfile

        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            oracle_path, _ = self._write_repo(repo)
            oracle = self._oracle()
            oracle["findings"].pop()
            oracle["findings"][0]["path"] = "/components/schemas/*"
            oracle_path.write_text(
                json.dumps(oracle, ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
            artifact_path = repo / "audit.json"

            result = self._run_audit(repo, oracle_path, artifact_path)

            self.assertEqual(1, result.returncode)
            payload = json.loads(result.stdout)
            self.assertTrue(any("wildcard" in error for error in payload["errors"]), payload)
            self.assertTrue(any("expected 10 findings but actual 11" in error for error in payload["errors"]), payload)
            self.assertTrue(any("missing finding" in error for error in payload["errors"]), payload)
            self.assertTrue(any("unexpected finding" in error for error in payload["errors"]), payload)
            self.assertTrue(artifact_path.is_file(), payload)

    def test_cli_rejects_edited_baseline_and_simultaneous_refreeze(self) -> None:
        import tempfile

        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            oracle_path, current_path = self._write_repo(repo)
            baseline_path = repo / "openapi/baseline/openapi-v1.0.0.yaml"
            baseline_path.write_text(current_path.read_text(encoding="utf-8"), encoding="utf-8")
            artifact_path = repo / "audit.json"

            edited = self._run_audit(repo, oracle_path, artifact_path)

            self.assertEqual(1, edited.returncode)
            self.assertTrue(
                any(
                    "worktree baseline differs" in error
                    for error in json.loads(edited.stdout)["errors"]
                )
            )

            env = os.environ.copy()
            env.update(
                {
                    "GIT_AUTHOR_NAME": "test",
                    "GIT_AUTHOR_EMAIL": "test@example.com",
                    "GIT_COMMITTER_NAME": "test",
                    "GIT_COMMITTER_EMAIL": "test@example.com",
                }
            )
            subprocess.run(["git", "-C", str(repo), "add", "."], check=True)
            subprocess.run(
                ["git", "-C", str(repo), "commit", "-q", "-m", "simultaneous refreeze"],
                check=True,
                env=env,
            )
            zero = self._run_audit(repo, oracle_path, artifact_path)
            self.assertEqual(1, zero.returncode)
            zero_payload = json.loads(zero.stdout)
            self.assertEqual(0, zero_payload["findingCount"])
            self.assertTrue(
                any("expected 11 findings but actual 0" in error for error in zero_payload["errors"]),
                zero_payload,
            )

    def test_targetjob_mutations_do_not_enter_practice_manifest(self) -> None:
        baseline, current = _d_35_docs()
        baseline["components"]["schemas"]["TargetJob"] = {
            "type": "object",
            "properties": {"sourceUrl": {"type": "string"}},
        }
        current["components"]["schemas"]["TargetJob"] = {
            "type": "object",
            "properties": {},
        }

        findings = od.normalize_d_35_findings(baseline, current)

        self.assertEqual([], od.compare_finding_sets(self._oracle()["findings"], findings))


class OpenAPI005OracleTests(unittest.TestCase):
    @staticmethod
    def _preserved_audit() -> dict:
        return json.loads(
            (
                REPO_ROOT
                / "openapi/baseline/audits/OPENAPI-005-resume-list-summary.json"
            ).read_text(encoding="utf-8")
        )

    @classmethod
    def _old_baseline(cls) -> dict:
        audit = cls._preserved_audit()
        source_kind, source_ref, source_path = audit["baselineSource"].split(":", 2)
        if source_kind != "git":
            raise AssertionError("OPENAPI-005 audit baseline must be a git snapshot")
        baseline_text = od._git_show(REPO_ROOT, source_ref, REPO_ROOT / source_path)
        if baseline_text is None:
            raise AssertionError("OPENAPI-005 audit baseline snapshot must remain readable")
        if audit["baselineSha256"] != od._sha256_text(baseline_text):
            raise AssertionError("OPENAPI-005 audit baseline digest drifted")
        return yaml.safe_load(baseline_text)

    def test_normalizer_and_contract_lock_only_resume_list_detail_split(self) -> None:
        baseline = self._old_baseline()
        current = copy.deepcopy(baseline)
        schemas = current["components"]["schemas"]
        fields = {
            "id": {"type": "string", "format": "uuid"},
            "title": {"type": "string"},
            "displayName": {"type": "string"},
            "language": {"type": "string"},
            "sourceType": {"type": "string", "enum": ["upload", "paste"]},
            "parseStatus": {"$ref": "#/components/schemas/TargetJobParseStatus"},
            "summaryHeadline": {
                "oneOf": [{"type": "string"}, {"type": "null"}]
            },
            "hasReadableContent": {"type": "boolean"},
            "updatedAt": {"type": "string", "format": "date-time"},
        }
        schemas["ResumeSummary"] = {
            "type": "object",
            "additionalProperties": False,
            "required": list(fields),
            "properties": fields,
        }
        schemas["PaginatedResume"]["allOf"][1]["properties"]["items"]["items"] = {
            "$ref": "#/components/schemas/ResumeSummary"
        }

        findings = od.normalize_openapi_005_findings(baseline, current)
        expected_findings = {
            (
                "breaking",
                "/components/schemas/PaginatedResume/properties/items/items",
                "response_item_ref_changed",
                "Resume",
                "ResumeSummary",
            ),
            (
                "additive",
                "/components/schemas/ResumeSummary",
                "schema_added_with_required_fields",
                "absent",
                ",".join(fields),
            ),
            (
                "additive",
                "/components/schemas/ResumeSummary/additionalProperties",
                "closed_object",
                "absent",
                False,
            ),
        }
        expected_signatures = {
            "id": "string(format=uuid)",
            "title": "string",
            "displayName": "string",
            "language": "string",
            "sourceType": "enum(upload,paste)",
            "parseStatus": "TargetJobParseStatus",
            "summaryHeadline": "string|null",
            "hasReadableContent": "boolean",
            "updatedAt": "string(format=date-time)",
        }
        expected_findings.update(
            {
                (
                    "additive",
                    f"/components/schemas/ResumeSummary/properties/{name}",
                    "property_added",
                    "absent",
                    signature,
                )
                for name, signature in expected_signatures.items()
            }
        )
        self.assertEqual(
            expected_findings,
            {
                (
                    finding["severity"],
                    finding["path"],
                    finding["kind"],
                    finding["before"],
                    finding["after"],
                )
                for finding in findings
            },
        )
        self.assertEqual([], od.validate_openapi_005_contract(baseline, current))

        current["components"]["schemas"]["ResumeSummary"]["properties"][
            "fileObjectId"
        ] = {"type": "string", "format": "uuid"}
        errors = od.validate_openapi_005_contract(baseline, current)
        self.assertTrue(any("properties" in error for error in errors), errors)

    def test_repo_preserved_audit_replays_after_guarded_refreeze(self) -> None:
        decision = (
            REPO_ROOT
            / "docs/spec/openapi-v1-contract/decisions/OPENAPI-005-resume-list-summary.md"
        )
        oracle = decision.with_name(
            "OPENAPI-005-resume-list-summary.expected-findings.json"
        )
        self.assertTrue(oracle.is_file(), "OPENAPI-005 machine oracle must be generated")
        payload = self._preserved_audit()
        expected = json.loads(oracle.read_text(encoding="utf-8"))
        baseline = self._old_baseline()
        current_path = REPO_ROOT / payload["currentSource"]
        current_text = current_path.read_text(encoding="utf-8")
        current = yaml.safe_load(current_text)
        frozen_text = (
            REPO_ROOT / "openapi/baseline/openapi-v1.0.0.yaml"
        ).read_text(encoding="utf-8")

        self.assertEqual("OPENAPI-005", payload["decisionId"])
        self.assertEqual("exact-set", payload["mode"])
        self.assertEqual([], payload["errors"])
        self.assertEqual(payload["expectedFindingCount"], payload["findingCount"])
        self.assertEqual(12, payload["findingCount"])
        self.assertEqual(payload["currentSha256"], od._sha256_text(current_text))
        self.assertEqual(current_text, frozen_text)
        self.assertEqual(
            [],
            od.compare_finding_sets(
                expected["findings"], od.normalize_openapi_005_findings(baseline, current)
            ),
        )
        self.assertEqual(
            [], od.compare_finding_sets(payload["findings"], expected["findings"])
        )
        self.assertEqual([], od.validate_openapi_005_contract(baseline, current))
        self.assertEqual(
            [],
            od.validate_openapi_005_invariants(
                baseline, current, expected["invariants"]
            ),
        )


class OpenAPI004OracleTests(unittest.TestCase):
    @staticmethod
    def _oracle() -> dict:
        return _openapi_004_oracle()

    def _write_repo(self, repo: Path) -> tuple[Path, Path, Path]:
        baseline, current = _openapi_004_docs()
        (repo / "openapi" / "baseline").mkdir(parents=True)
        authority_dir = repo / "docs" / "spec" / "openapi-v1-contract"
        decision_dir = authority_dir / "decisions"
        decision_dir.mkdir(parents=True)
        baseline_path = repo / "openapi" / "baseline" / "openapi-v1.0.0.yaml"
        current_path = repo / "openapi" / "openapi.yaml"
        decision_path = decision_dir / "OPENAPI-004-targetjob-report-overview.md"
        oracle_path = (
            decision_dir
            / "OPENAPI-004-targetjob-report-overview.expected-findings.json"
        )
        baseline_path.write_text(
            yaml.safe_dump(baseline, sort_keys=False), encoding="utf-8"
        )
        current_path.write_text(
            yaml.safe_dump(baseline, sort_keys=False), encoding="utf-8"
        )
        decision_path.write_text(
            "# OPENAPI-004\n\n> **ID**: OPENAPI-004\n> **状态**: accepted\n\nR-A\n",
            encoding="utf-8",
        )
        (authority_dir / "spec.md").write_text(
            "| D-36 | accepted OPENAPI-004 TargetJob report overview R-A |\n",
            encoding="utf-8",
        )
        (authority_dir / "history.md").write_text(
            "| 2026-07-14 | 1.57 | accepted OPENAPI-004 TargetJob report overview |\n",
            encoding="utf-8",
        )
        oracle_path.write_text(
            json.dumps(self._oracle(), ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )

        env = os.environ.copy()
        env.update(
            {
                "GIT_AUTHOR_NAME": "test",
                "GIT_AUTHOR_EMAIL": "test@example.com",
                "GIT_COMMITTER_NAME": "test",
                "GIT_COMMITTER_EMAIL": "test@example.com",
            }
        )
        subprocess.run(["git", "init", "-q", "-b", "main", str(repo)], check=True)
        subprocess.run(["git", "-C", str(repo), "add", "."], check=True)
        subprocess.run(
            ["git", "-C", str(repo), "commit", "-q", "-m", "old baseline"],
            check=True,
            env=env,
        )
        current_path.write_text(
            yaml.safe_dump(current, sort_keys=False), encoding="utf-8"
        )
        return decision_path, oracle_path, current_path

    def _run_audit(
        self,
        repo: Path,
        decision_path: Path,
        oracle_path: Path,
        artifact_path: Path,
    ) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [
                sys.executable,
                str(HERE / "openapi_diff.py"),
                "--repo-root",
                str(repo),
                "--decision-id",
                "OPENAPI-004",
                "--decision-record",
                str(decision_path),
                "--oracle",
                str(oracle_path),
                "--base-ref",
                "main",
                "--output",
                str(artifact_path),
            ],
            capture_output=True,
            text=True,
        )

    def test_normalizer_exactly_isolates_six_breaking_and_nine_additive_findings(
        self,
    ) -> None:
        baseline, current = _openapi_004_docs()

        findings = od.normalize_openapi_004_findings(baseline, current)

        self.assertEqual([], od.compare_finding_sets(self._oracle()["findings"], findings))
        self.assertEqual(15, len(findings))
        self.assertEqual(6, sum(finding["severity"] == "breaking" for finding in findings))
        self.assertEqual(9, sum(finding["severity"] == "additive" for finding in findings))
        self.assertEqual([], od.validate_openapi_004_contract(baseline, current))
        self.assertEqual(
            [],
            od.validate_openapi_004_invariants(
                baseline, current, self._oracle()["invariants"]
            ),
        )

    def test_openapi_002_and_d_35_deltas_do_not_enter_report_manifest(self) -> None:
        baseline, current = _openapi_004_docs()
        baseline_schemas = baseline["components"]["schemas"]
        current_schemas = current["components"]["schemas"]
        baseline_schemas["ImportTargetJobRequest"] = {
            "type": "object",
            "properties": {"source": {"type": "string"}},
        }
        current_schemas["ImportTargetJobRequest"] = {
            "type": "object",
            "properties": {"rawText": {"type": "string"}},
        }
        baseline_schemas["PracticeMessage"] = {
            "type": "object",
            "properties": {"role": {"type": "string"}},
        }
        current_schemas["PracticeMessage"] = {
            "oneOf": [
                {"$ref": "#/components/schemas/PracticeUserMessage"},
                {"$ref": "#/components/schemas/PracticeAssistantMessage"},
            ]
        }

        findings = od.normalize_openapi_004_findings(baseline, current)

        self.assertEqual([], od.compare_finding_sets(self._oracle()["findings"], findings))

    def test_contract_and_invariants_reject_silent_drift_matrix(self) -> None:
        mutations = {
            "round bounds drift": lambda current: current["components"]["schemas"][
                "TargetJobReportsOverview"
            ]["properties"]["rounds"].update({"minItems": 1}),
            "round summary extra property": lambda current: current["components"][
                "schemas"
            ]["TargetJobReportRoundOverview"]["properties"].update(
                {"legacyReport": {"type": "string"}}
            ),
            "operation id drift": lambda current: current["paths"][
                "/targets/{targetJobId}/reports"
            ]["get"].update({"operationId": "listReports"}),
            "query alias": lambda current: current["paths"][
                "/targets/{targetJobId}/reports"
            ]["get"]["parameters"].append(
                {
                    "name": "offset",
                    "in": "query",
                    "required": False,
                    "schema": {"type": "integer"},
                }
            ),
        }
        for name, mutate in mutations.items():
            with self.subTest(name=name):
                baseline, current = _openapi_004_docs()
                mutate(current)
                exact_errors = od.compare_finding_sets(
                    self._oracle()["findings"],
                    od.normalize_openapi_004_findings(baseline, current),
                )
                contract_errors = od.validate_openapi_004_contract(baseline, current)
                invariant_errors = od.validate_openapi_004_invariants(
                    baseline, current, self._oracle()["invariants"]
                )
                self.assertTrue(exact_errors or contract_errors or invariant_errors, name)

    def test_cli_requires_authority_exact_manifest_and_is_deterministic(self) -> None:
        import tempfile

        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            decision_path, oracle_path, _ = self._write_repo(repo)
            artifact_path = repo / "audit.json"

            first = self._run_audit(
                repo, decision_path, oracle_path, artifact_path
            )
            first_artifact = artifact_path.read_bytes()
            second = self._run_audit(
                repo, decision_path, oracle_path, artifact_path
            )

            self.assertEqual(0, first.returncode, first.stderr or first.stdout)
            self.assertEqual(0, second.returncode, second.stderr or second.stdout)
            payload = json.loads(second.stdout)
            self.assertEqual(15, payload["expectedFindingCount"])
            self.assertEqual(15, payload["findingCount"])
            self.assertEqual(
                {"breaking": 6, "additive": 9, "informational": 0},
                payload["summary"],
            )
            self.assertEqual([], payload["errors"])
            self.assertEqual(first.stdout, second.stdout)
            self.assertEqual(first_artifact, artifact_path.read_bytes())

            decision_path.write_text(
                "# OPENAPI-004\n\n> **ID**: OPENAPI-004\n> **状态**: draft\n",
                encoding="utf-8",
            )
            (repo / "docs/spec/openapi-v1-contract/spec.md").write_text(
                "| D-35 | unrelated |\n", encoding="utf-8"
            )
            (repo / "docs/spec/openapi-v1-contract/history.md").write_text(
                "| 2026-07-13 | 1.56 | unrelated |\n", encoding="utf-8"
            )
            failed = self._run_audit(
                repo, decision_path, oracle_path, artifact_path
            )
            errors = json.loads(failed.stdout)["errors"]
            self.assertEqual(1, failed.returncode)
            self.assertTrue(any("status must be accepted" in error for error in errors), errors)
            self.assertTrue(any("spec D-36" in error for error in errors), errors)
            self.assertTrue(any("history 1.57" in error for error in errors), errors)

    def test_cli_rejects_wildcard_and_edited_worktree_baseline(self) -> None:
        import tempfile

        with tempfile.TemporaryDirectory() as tmp:
            repo = Path(tmp)
            decision_path, oracle_path, current_path = self._write_repo(repo)
            oracle = self._oracle()
            oracle["findings"].pop()
            oracle["findings"][0]["path"] = "/paths/*"
            oracle_path.write_text(
                json.dumps(oracle, ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
            artifact_path = repo / "audit.json"

            wildcard = self._run_audit(
                repo, decision_path, oracle_path, artifact_path
            )
            wildcard_errors = json.loads(wildcard.stdout)["errors"]
            self.assertEqual(1, wildcard.returncode)
            self.assertTrue(any("wildcard" in error for error in wildcard_errors), wildcard_errors)
            self.assertTrue(
                any("expected 14 findings but actual 15" in error for error in wildcard_errors),
                wildcard_errors,
            )

            oracle_path.write_text(
                json.dumps(self._oracle(), ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
            (repo / "openapi/baseline/openapi-v1.0.0.yaml").write_text(
                current_path.read_text(encoding="utf-8"), encoding="utf-8"
            )
            edited = self._run_audit(
                repo, decision_path, oracle_path, artifact_path
            )
            edited_errors = json.loads(edited.stdout)["errors"]
            self.assertEqual(1, edited.returncode)
            self.assertTrue(
                any("worktree baseline differs" in error for error in edited_errors),
                edited_errors,
            )


class D35PreservedAuditTests(unittest.TestCase):
    def test_repo_practice_audit_replays_from_old_baseline(self) -> None:
        audit = json.loads(
            (
                REPO_ROOT
                / "openapi/baseline/audits/D-35-practice-durable-recovery.json"
            ).read_text(encoding="utf-8")
        )
        oracle = json.loads(
            (
                REPO_ROOT
                / "docs/spec/openapi-v1-contract/decisions/D-35-practice-durable-recovery.expected-findings.json"
            ).read_text(encoding="utf-8")
        )
        source_kind, source_ref, source_path = audit["baselineSource"].split(":", 2)
        self.assertEqual("git", source_kind)
        baseline_text = od._git_show(REPO_ROOT, source_ref, REPO_ROOT / source_path)
        self.assertIsNotNone(baseline_text)
        current_text = (REPO_ROOT / audit["currentSource"]).read_text(encoding="utf-8")
        self.assertEqual(audit["baselineSha256"], od._sha256_text(baseline_text))
        self.assertEqual("openapi/openapi.yaml", audit["currentSource"])
        self.assertRegex(audit["currentSha256"], r"^[0-9a-f]{64}$")
        baseline = yaml.safe_load(baseline_text)
        current = yaml.safe_load(current_text)

        findings = od.normalize_d_35_findings(baseline, current)

        self.assertEqual([], audit["errors"])
        self.assertEqual([], od.compare_finding_sets(oracle["findings"], findings))
        self.assertEqual([], od.compare_finding_sets(audit["findings"], findings))
        self.assertEqual([], od.validate_d_35_contract(current, baseline))
        self.assertEqual(
            [],
            od.validate_d_35_invariants(
                baseline, current, oracle["invariants"]
            ),
        )
        self.assertEqual(11, len(findings))
        self.assertEqual(8, sum(finding["severity"] == "breaking" for finding in findings))
        self.assertEqual(3, sum(finding["severity"] == "additive" for finding in findings))


class OpenAPI004PreservedAuditTests(unittest.TestCase):
    def test_repo_report_overview_audit_replays_from_old_baseline(self) -> None:
        audit = json.loads(
            (
                REPO_ROOT
                / "openapi/baseline/audits/OPENAPI-004-targetjob-report-overview.json"
            ).read_text(encoding="utf-8")
        )
        oracle = json.loads(
            (
                REPO_ROOT
                / "docs/spec/openapi-v1-contract/decisions/OPENAPI-004-targetjob-report-overview.expected-findings.json"
            ).read_text(encoding="utf-8")
        )
        source_kind, source_ref, source_path = audit["baselineSource"].split(":", 2)
        self.assertEqual("git", source_kind)
        baseline_text = od._git_show(REPO_ROOT, source_ref, REPO_ROOT / source_path)
        self.assertIsNotNone(baseline_text)
        current_text = (REPO_ROOT / audit["currentSource"]).read_text(encoding="utf-8")
        self.assertEqual(audit["baselineSha256"], od._sha256_text(baseline_text))
        # The audit pins the proposal that introduced OPENAPI-004. Later accepted
        # additive changes legitimately change the live source without rewriting
        # this immutable historical record.
        self.assertRegex(audit["currentSha256"], r"^[0-9a-f]{64}$")
        baseline = yaml.safe_load(baseline_text)
        current = yaml.safe_load(current_text)

        findings = od.normalize_openapi_004_findings(baseline, current)

        decision_text = (
            REPO_ROOT
            / "docs/spec/openapi-v1-contract/decisions/OPENAPI-004-targetjob-report-overview.md"
        ).read_text(encoding="utf-8")
        self.assertEqual([], audit["errors"])
        self.assertEqual([], od.validate_decision_record(decision_text, "OPENAPI-004"))
        self.assertEqual([], od.validate_openapi_004_authority(REPO_ROOT, oracle))
        self.assertEqual([], od.compare_finding_sets(oracle["findings"], findings))
        self.assertEqual([], od.compare_finding_sets(audit["findings"], findings))
        self.assertEqual([], od.validate_openapi_004_contract(baseline, current))
        self.assertEqual(
            [],
            od.validate_openapi_004_invariants(
                baseline, current, oracle["invariants"]
            ),
        )
        self.assertEqual(15, len(findings))
        self.assertEqual(6, sum(finding["severity"] == "breaking" for finding in findings))
        self.assertEqual(9, sum(finding["severity"] == "additive" for finding in findings))


class OpenAPI002PreservedAuditTests(unittest.TestCase):
    def test_repo_openapi_002_preserved_audit_replays_from_old_baseline(self) -> None:
        audit = json.loads(
            (
                REPO_ROOT
                / "openapi/baseline/audits/OPENAPI-002-targetjob-paste-only.json"
            ).read_text(encoding="utf-8")
        )
        oracle = json.loads(
            (
                REPO_ROOT
                / "docs/spec/openapi-v1-contract/decisions/OPENAPI-002-targetjob-paste-only.expected-findings.json"
            ).read_text(encoding="utf-8")
        )
        source_kind, source_ref, source_path = audit["baselineSource"].split(":", 2)
        self.assertEqual("git", source_kind)
        baseline_text = od._git_show(REPO_ROOT, source_ref, REPO_ROOT / source_path)
        self.assertIsNotNone(baseline_text)
        self.assertEqual(audit["baselineSha256"], od._sha256_text(baseline_text))
        current_text = (REPO_ROOT / "openapi/openapi.yaml").read_text(encoding="utf-8")
        baseline = yaml.safe_load(baseline_text)
        current = yaml.safe_load(current_text)

        findings = od.normalize_openapi_002_findings(baseline, current)

        self.assertEqual([], audit["errors"])
        self.assertEqual([], od.compare_finding_sets(oracle["findings"], findings))
        self.assertEqual([], od.compare_finding_sets(audit["findings"], findings))
        self.assertEqual([], od.validate_openapi_002_invariants(current, oracle["invariants"]))
        self.assertEqual(17, len(findings))
        self.assertTrue(all(finding["severity"] == "breaking" for finding in findings))


class OpenAPI001PreservedAuditTests(unittest.TestCase):
    def test_repo_openapi_001_preserved_audit_exact_matches_machine_oracle(self) -> None:
        audit = json.loads(
            (
                REPO_ROOT
                / "openapi/baseline/audits/OPENAPI-001-report-direct-semantics.json"
            ).read_text(encoding="utf-8")
        )
        source_kind, source_ref, source_path = audit["baselineSource"].split(":", 2)
        self.assertEqual("git", source_kind)
        baseline_text = od._git_show(REPO_ROOT, source_ref, REPO_ROOT / source_path)
        self.assertIsNotNone(baseline_text)
        self.assertEqual(audit["baselineSha256"], od._sha256_text(baseline_text))
        self.assertRegex(audit["currentSha256"], r"^[0-9a-f]{64}$")
        oracle = json.loads(
            (
                REPO_ROOT
                / "docs/spec/openapi-v1-contract/decisions/OPENAPI-001-report-direct-semantics.expected-findings.json"
            ).read_text(encoding="utf-8")
        )

        # OPENAPI-001 predates later accepted contract revisions and its proposed
        # source is not a permanent snapshot. Preserve the signed exact finding
        # set instead of pretending the latest HEAD is that historical source.
        findings = audit["findings"]

        self.assertEqual([], audit["errors"])
        self.assertEqual([], od.compare_finding_sets(oracle["findings"], findings))
        self.assertEqual(36, len(findings))
        self.assertEqual(33, sum(finding["severity"] == "breaking" for finding in findings))
        self.assertEqual(3, sum(finding["severity"] == "additive" for finding in findings))
        self.assertIn(
            {
                "severity": "additive",
                "path": "/components/schemas/ApiErrorCode/enum",
                "kind": "enum_value_added",
                "before": "absent",
                "after": "REPORT_CONTEXT_TOO_LARGE",
            },
            findings,
        )


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
            self.assertEqual(payload["inventory"]["expectedOperations"], 37)
            self.assertEqual(payload["inventory"]["baselineOperations"], 3)
            self.assertEqual(payload["inventory"]["currentOperations"], 3)
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
