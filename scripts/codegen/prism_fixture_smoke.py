#!/usr/bin/env python3
"""Prism fixture-parity smoke for B2 002 fixture/example parity.

Hits the fixed operation matrix on a Prism mock server (assumed running on
http://127.0.0.1:4010 against `openapi/.generated/openapi-with-fixtures.yaml`)
with `Prefer: example=default`, then asserts the response body is byte-equal
to the matching fixture's `scenarios.default.response.body`.

This script is a hand-runnable smoke; it does not start Prism. To run:

    make render-openapi-fixture-examples
    npx @stoplight/prism-cli mock openapi/.generated/openapi-with-fixtures.yaml -p 4010 &
    python3 scripts/codegen/prism_fixture_smoke.py
"""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path
from typing import Iterable

REPO_ROOT_DEFAULT = Path(__file__).resolve().parents[2]


# (operationId, method, path, expected_status, fixture_relative_path)
SMOKE_MATRIX: tuple[tuple[str, str, str, int, str], ...] = (
    ("getMe", "GET", "/me", 200,
     "openapi/fixtures/Auth/getMe.json"),
    ("listResumes", "GET", "/resumes", 200,
     "openapi/fixtures/Resumes/listResumes.json"),
    ("getResume", "GET",
     "/resumes/01918fa0-0000-7000-8000-000000001000", 200,
     "openapi/fixtures/Resumes/getResume.json"),
    ("listTargetJobs", "GET", "/targets", 200,
     "openapi/fixtures/TargetJobs/listTargetJobs.json"),
    ("getTargetJob", "GET",
     "/targets/01918fa0-0000-7000-8000-000000002000", 200,
     "openapi/fixtures/TargetJobs/getTargetJob.json"),
    ("importTargetJob", "POST", "/targets/import", 202,
     "openapi/fixtures/TargetJobs/importTargetJob.json"),
    ("updateTargetJob", "PATCH",
     "/targets/01918fa0-0000-7000-8000-000000002000", 200,
     "openapi/fixtures/TargetJobs/updateTargetJob.json"),
    ("archiveTargetJob", "POST",
     "/targets/01918fa0-0000-7000-8000-000000002000/archive", 202,
     "openapi/fixtures/TargetJobs/archiveTargetJob.json"),
    ("getPracticeSession", "GET",
     "/practice/sessions/01918fa0-0050-7a00-8a00-000000000050", 200,
     "openapi/fixtures/PracticeSessions/getPracticeSession.json"),
    ("getFeedbackReport", "GET",
     "/reports/01918fa0-0070-7a00-8a00-000000000070", 200,
     "openapi/fixtures/Reports/getFeedbackReport.json"),
    ("getReportConversation", "GET",
     "/reports/01918fa0-0070-7000-8000-000000000070/conversation", 200,
     "openapi/fixtures/Reports/getReportConversation.json"),
    ("listTargetJobReports", "GET",
     "/targets/01918fa0-0000-7000-8000-000000002000/reports", 200,
     "openapi/fixtures/Reports/listTargetJobReports.json"),
    ("createPracticePlan", "POST", "/practice/plans", 201,
     "openapi/fixtures/PracticePlans/createPracticePlan.json"),
    ("requestPrivacyExport", "POST", "/privacy/exports", 501,
     "openapi/fixtures/Privacy/requestPrivacyExport.json"),
)


def _curl(
    method: str,
    url: str,
    prefer: str,
    request_body: dict | None = None,
) -> tuple[int, str]:
    cmd = [
        "/usr/bin/curl", "-s", "-w", "\nHTTP=%{http_code}\n",
        "-H", f"Prefer: {prefer}",
        "-H", "Cookie: ei_session=fake",
    ]
    if method != "GET":
        cmd += ["-X", method,
                "-H", "Idempotency-Key: 01918fa0-0001-7a00-8a00-aaaaaaaaaaaa"]
    if request_body is not None:
        cmd += [
            "-H",
            "Content-Type: application/json",
            "--data",
            json.dumps(request_body, ensure_ascii=False, separators=(",", ":")),
        ]
    cmd.append(url)
    res = subprocess.run(cmd, capture_output=True, text=True, check=False)
    text = res.stdout
    if "HTTP=" not in text:
        return -1, text
    body, _, status_line = text.rpartition("\nHTTP=")
    return int(status_line.strip().rstrip()), body.rstrip("\n")


def main(argv: Iterable[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", type=Path, default=REPO_ROOT_DEFAULT)
    parser.add_argument("--base-url", default="http://127.0.0.1:4010")
    args = parser.parse_args(list(argv))

    repo = args.repo_root.resolve()
    failures = 0
    for opid, method, path, expected, fixture_rel in SMOKE_MATRIX:
        prefer = f"code={expected}, example=default" if expected != 200 else "example=default"
        url = args.base_url + path
        with (repo / fixture_rel).open("r", encoding="utf-8") as f:
            fixture_default = json.load(f)["scenarios"]["default"]
        request_body = (fixture_default.get("request") or {}).get("body")
        got_status, body_text = _curl(method, url, prefer, request_body)
        try:
            prism_body = json.loads(body_text) if body_text.strip() else None
        except json.JSONDecodeError as e:
            print(f"FAIL {opid}: prism response is not JSON: {e}", file=sys.stderr)
            failures += 1
            continue
        fixture_body = fixture_default["response"]["body"]
        ok_status = got_status == expected
        ok_body = prism_body == fixture_body
        marker = "OK " if (ok_status and ok_body) else "FAIL"
        print(f"{marker} {opid} {method} {path} status={got_status}/{expected} body-equal={ok_body}")
        if not (ok_status and ok_body):
            failures += 1
            if not ok_body:
                print(
                    f"     fixture[:200]={json.dumps(fixture_body, ensure_ascii=False)[:200]}"
                )
                print(
                    f"     prism  [:200]={json.dumps(prism_body, ensure_ascii=False)[:200] if prism_body else None}"
                )

    if failures:
        print(f"prism-fixture-smoke: {failures} of {len(SMOKE_MATRIX)} checks failed", file=sys.stderr)
        return 1
    print(f"prism-fixture-smoke: OK — {len(SMOKE_MATRIX)} checks byte-equal")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
