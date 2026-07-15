#!/usr/bin/env python3
"""Validate the redacted, exact-six full-page P0.099 screenshot manifest."""

from __future__ import annotations

import argparse
import hashlib
import json
import re
import shutil
import struct
import sys
import urllib.parse
import uuid
import zlib
from collections import Counter
from datetime import datetime
from pathlib import Path
from typing import Any


EXPECTED = {
    "report-zh-needs-practice-desktop.png": ("zh", "ready-needs-practice", "desktop", 1440, 1200),
    "report-zh-needs-practice-mobile.png": ("zh", "ready-needs-practice", "mobile", 390, 844),
    "report-en-well-prepared-desktop.png": ("en", "ready-well-prepared", "desktop", 1440, 1200),
    "report-en-well-prepared-mobile.png": ("en", "ready-well-prepared", "mobile", 390, 844),
    "report-generating-desktop.png": ("zh", "generating", "desktop", 1440, 1200),
    "report-generating-mobile.png": ("zh", "generating", "mobile", 390, 844),
}

PREPAREDNESS = {
    "ready-needs-practice": "needs_practice",
    "ready-well-prepared": "well_prepared",
    "generating": None,
}

FORBIDDEN_KEYS = {
    "answer",
    "answer_text",
    "auth_code",
    "cookie",
    "database_url",
    "email_code",
    "frozen_context",
    "jd_text",
    "prompt",
    "prompt_body",
    "provider_response",
    "raw_context",
    "response",
    "response_body",
    "resume_text",
    "session_cookie",
    "source_message_seq_nos",
    "transcript",
}

FORBIDDEN_VALUE_PATTERNS = (
    re.compile(r"(?i)\bei_session="),
    re.compile(r"(?i)\b(?:AI_PROVIDER_API_KEY|SESSION_COOKIE_SECRET|AUTH_CHALLENGE_TOKEN_PEPPER)\b"),
    re.compile(r"\bsk-[A-Za-z0-9_-]{12,}\b"),
    re.compile(r"(?i)auth/(?:email/)?verify\?token="),
    re.compile(r"(?i)postgres(?:ql)?://[^\s]+"),
)

PNG_METADATA_CHUNKS = {b"tEXt", b"zTXt", b"iTXt", b"eXIf", b"iCCP", b"tIME"}
FORBIDDEN_FILENAME_MARKERS = (
    "raw",
    "cookie",
    "browser-state",
    "candidate-output",
    "judge-output",
)
ALLOWED_TOP_LEVEL = {
    "cleanup.env",
    "conversation-navigation.json",
    "live-capture.json",
    "manual-visual-audit.json",
    "manifest.json",
    "result.json",
    "screenshots",
    "setup.env",
    "trigger.env",
    "trigger.log",
}
MAX_NAVIGATION_REQUEST_COUNT = 4
FORBIDDEN_TEXT_PATTERNS = (
    re.compile(
        rb'(?i)["\'](?:answer(?:_?text)?|auth_?code|cookie|email_?code|frozen_?context|jd_?text|prompt(?:_?body)?|provider_?response|raw_?context|response(?:_?body)?|resume_?text|session_?cookie(?:_?value)?|transcript)["\']\s*:'
    ),
    re.compile(rb"(?i)\bei_session="),
    re.compile(rb"(?i)\b(?:AI_PROVIDER_API_KEY|SESSION_COOKIE_SECRET|AUTH_CHALLENGE_TOKEN_PEPPER)\s*[:=]"),
    re.compile(rb"\bsk-[A-Za-z0-9_-]{12,}\b"),
    re.compile(rb"(?i)auth/(?:email/)?verify\?token="),
    re.compile(rb"(?i)\bDATABASE_URL\s*[:=]"),
    re.compile(rb"(?i)postgres(?:ql)?://[^\s]+"),
    re.compile(rb'''(?i)["']source_?message_?seq_?nos["']\s*:'''),
)

READY_VISUAL_CHECKS = {
    "report_page_visible",
    "expected_state_visible",
    "preparedness_visible",
    "dimension_and_evidence_content_visible",
    "action_region_visible",
    "action_labels_complete_without_clipping_or_ellipsis",
    "horizontal_overflow_absent",
}
GENERATING_VISUAL_CHECKS = {
    "report_page_visible",
    "expected_state_visible",
    "generating_indicator_visible",
    "ready_content_absent",
    "false_ready_claim_absent",
    "clipping_or_overlap_absent",
    "horizontal_overflow_absent",
}


class EvidenceError(ValueError):
    pass


def fail(message: str) -> None:
    raise EvidenceError(message)


def require_keys(value: dict[str, Any], expected: set[str], path: str) -> None:
    actual = set(value)
    if actual != expected:
        fail(f"{path} keys={sorted(actual)}, expected exactly {sorted(expected)}")


def scan_redline(value: Any, path: str = "$") -> None:
    if isinstance(value, dict):
        for key, child in value.items():
            normalized = key.lower()
            if normalized in FORBIDDEN_KEYS:
                fail(f"{path}.{key} is forbidden raw evidence")
            scan_redline(child, f"{path}.{key}")
        return
    if isinstance(value, list):
        for index, child in enumerate(value):
            scan_redline(child, f"{path}[{index}]")
        return
    if isinstance(value, str):
        for pattern in FORBIDDEN_VALUE_PATTERNS:
            if pattern.search(value):
                fail(f"{path} contains forbidden sensitive material")


def paeth_predictor(left: int, up: int, upper_left: int) -> int:
    estimate = left + up - upper_left
    left_distance = abs(estimate - left)
    up_distance = abs(estimate - up)
    upper_left_distance = abs(estimate - upper_left)
    if left_distance <= up_distance and left_distance <= upper_left_distance:
        return left
    if up_distance <= upper_left_distance:
        return up
    return upper_left


def png_pixels(path: Path) -> tuple[int, int, int, list[bytes]]:
    try:
        data = path.read_bytes()
    except OSError as exc:
        fail(f"{path.name} cannot be read: {exc}")
    if not data.startswith(b"\x89PNG\r\n\x1a\n"):
        fail(f"{path.name} is not a valid PNG signature")

    offset = 8
    chunk_index = 0
    dimensions: tuple[int, int] | None = None
    channels: int | None = None
    idat_payloads: list[bytes] = []
    saw_iend = False
    while offset < len(data):
        if len(data) - offset < 12:
            fail(f"{path.name} has a truncated PNG chunk")
        length = struct.unpack(">I", data[offset : offset + 4])[0]
        kind = data[offset + 4 : offset + 8]
        payload_start = offset + 8
        payload_end = payload_start + length
        crc_end = payload_end + 4
        if crc_end > len(data):
            fail(f"{path.name} has a truncated PNG chunk payload")
        payload = data[payload_start:payload_end]
        expected_crc = struct.unpack(">I", data[payload_end:crc_end])[0]
        actual_crc = zlib.crc32(kind + payload) & 0xFFFFFFFF
        if actual_crc != expected_crc:
            fail(f"{path.name} PNG chunk {kind!r} has an invalid CRC")
        if kind in PNG_METADATA_CHUNKS:
            for pattern in FORBIDDEN_TEXT_PATTERNS:
                if pattern.search(payload):
                    fail(
                        f"{path.name} PNG metadata chunk "
                        f"{kind.decode('ascii', errors='replace')} contains forbidden sensitive material"
                    )
        if chunk_index == 0:
            if kind != b"IHDR" or length != 13:
                fail(f"{path.name} PNG must start with one 13-byte IHDR")
            width, height = struct.unpack(">II", payload[:8])
            if width <= 0 or height <= 0:
                fail(f"{path.name} PNG dimensions must be positive")
            bit_depth, color_type, compression, png_filter, interlace = payload[8:13]
            if bit_depth != 8 or color_type not in (2, 6):
                fail(f"{path.name} PNG must use supported 8-bit RGB or RGBA pixels")
            if compression != 0 or png_filter != 0 or interlace != 0:
                fail(f"{path.name} PNG uses unsupported compression/filter/interlace settings")
            dimensions = (width, height)
            channels = 3 if color_type == 2 else 4
        elif kind == b"IHDR":
            fail(f"{path.name} contains a duplicate IHDR")
        if kind == b"IDAT":
            idat_payloads.append(payload)
        if kind == b"IEND":
            if length != 0:
                fail(f"{path.name} IEND must have an empty payload")
            saw_iend = True
            if crc_end != len(data):
                fail(f"{path.name} contains trailing bytes after IEND")
            offset = crc_end
            break
        offset = crc_end
        chunk_index += 1

    if not saw_iend:
        fail(f"{path.name} is missing IEND")
    if not idat_payloads:
        fail(f"{path.name} is missing IDAT")
    if dimensions is None or channels is None:
        fail(f"{path.name} is missing IHDR")

    try:
        scanlines = zlib.decompress(b"".join(idat_payloads))
    except zlib.error as exc:
        fail(f"{path.name} IDAT is not a valid zlib stream: {exc}")
    width, height = dimensions
    row_length = 1 + width * channels
    expected_length = row_length * height
    if len(scanlines) != expected_length:
        fail(
            f"{path.name} decompressed scanline length={len(scanlines)}, "
            f"expected exactly {expected_length}"
        )
    rows: list[bytes] = []
    previous = bytes(width * channels)
    for row_index in range(height):
        row_start = row_index * row_length
        filter_type = scanlines[row_start]
        if filter_type > 4:
            fail(f"{path.name} scanline {row_index} has invalid filter type {filter_type}")
        encoded = scanlines[row_start + 1 : row_start + row_length]
        decoded = bytearray(len(encoded))
        for index, value in enumerate(encoded):
            left = decoded[index - channels] if index >= channels else 0
            up = previous[index]
            upper_left = previous[index - channels] if index >= channels else 0
            predictor = 0
            if filter_type == 1:
                predictor = left
            elif filter_type == 2:
                predictor = up
            elif filter_type == 3:
                predictor = (left + up) // 2
            elif filter_type == 4:
                predictor = paeth_predictor(left, up, upper_left)
            decoded[index] = (value + predictor) & 0xFF
        previous = bytes(decoded)
        rows.append(previous)
    return width, height, channels, rows


def visual_metrics(
    rows: list[bytes],
    width: int,
    channels: int,
    start_row: int = 0,
) -> tuple[int, float]:
    height = len(rows)
    if not 0 <= start_row < height:
        fail("visual content region is empty")
    x_step = max(1, width // 128)
    y_step = max(1, (height - start_row) // 128)
    colors: Counter[tuple[int, int, int]] = Counter()
    for row_index in range(start_row, height, y_step):
        row = rows[row_index]
        for x in range(0, width, x_step):
            offset = x * channels
            red, green, blue = row[offset : offset + 3]
            if channels == 4:
                alpha = row[offset + 3]
                red = (red * alpha + 255 * (255 - alpha)) // 255
                green = (green * alpha + 255 * (255 - alpha)) // 255
                blue = (blue * alpha + 255 * (255 - alpha)) // 255
            colors[(red >> 4, green >> 4, blue >> 4)] += 1
    samples = sum(colors.values())
    if samples == 0:
        fail("visual content sampling produced no pixels")
    return len(colors), max(colors.values()) / samples


def require_visual_content(
    path: Path,
    rows: list[bytes],
    width: int,
    channels: int,
    *,
    start_row: int = 0,
    label: str = "visual content",
    maximum_dominant_ratio: float = 0.998,
) -> None:
    distinct_colors, dominant_ratio = visual_metrics(rows, width, channels, start_row)
    if distinct_colors < 4 or dominant_ratio > maximum_dominant_ratio:
        fail(
            f"{path.name} {label} is blank or near-solid "
            f"(quantized_colors={distinct_colors}, dominant_ratio={dominant_ratio:.4f})"
        )


def remove_path(path: Path) -> None:
    if path.is_symlink() or path.is_file():
        path.unlink()
    elif path.is_dir():
        shutil.rmtree(path)
    else:
        path.unlink(missing_ok=True)


def sanitize_output(output_dir: Path, failed: bool, setup: bool = False) -> int:
    removed = 0
    if output_dir.is_symlink() or (output_dir.exists() and not output_dir.is_dir()):
        remove_path(output_dir)
        output_dir.mkdir(parents=True, mode=0o700)
        removed += 1
    if not output_dir.exists():
        return removed
    if setup:
        for path in sorted(output_dir.iterdir()):
            remove_path(path)
            removed += 1
        return removed

    for path in sorted(output_dir.iterdir()):
        if path.name not in ALLOWED_TOP_LEVEL:
            remove_path(path)
            removed += 1
    if failed:
        for path in (
            output_dir / "manifest.json",
            output_dir / "manual-visual-audit.json",
            output_dir / "independent-agent-audit.json",
        ):
            if path.exists():
                path.unlink()
                removed += 1
        screenshots = output_dir / "screenshots"
        if screenshots.exists():
            if screenshots.is_symlink() or not screenshots.is_dir():
                screenshots.unlink()
            else:
                shutil.rmtree(screenshots)
            removed += 1

    for path in sorted(output_dir.rglob("*")):
        if not path.is_file():
            continue
        normalized_name = path.name.lower()
        should_remove = any(marker in normalized_name for marker in FORBIDDEN_FILENAME_MARKERS)
        if not should_remove and path.suffix.lower() != ".png":
            try:
                body = path.read_bytes()
            except OSError:
                should_remove = True
            else:
                should_remove = any(pattern.search(body) for pattern in FORBIDDEN_TEXT_PATTERNS)
        if should_remove and path.exists():
            path.unlink()
            removed += 1
    return removed


def validate_ref(value: Any, path: str) -> str:
    if not isinstance(value, str) or not value or len(value) > 128 or re.search(r"\s", value):
        fail(f"{path} must be a non-empty opaque reference")
    return value


def validate_digest(value: Any, path: str) -> str:
    if not isinstance(value, str) or re.fullmatch(r"[0-9a-f]{64}", value) is None:
        fail(f"{path} must be a SHA-256 digest")
    return value


def validate_row(row: Any, output_dir: Path, expected_file: str, run_id: str) -> tuple[str, str, str, str]:
    if not isinstance(row, dict):
        fail(f"screenshot row for {expected_file} must be an object")
    require_keys(
        row,
        {
            "file",
            "locale",
            "state",
            "fixture",
            "viewport",
            "full_page",
            "report_ref",
            "session_ref",
            "screenshot_sha256",
            "evidence",
        },
        f"screenshot[{expected_file}]",
    )
    if row["file"] != expected_file:
        fail(f"screenshot row file={row['file']!r}, expected {expected_file!r}")

    locale, state, viewport_name, viewport_width, viewport_height = EXPECTED[expected_file]
    expected_fixture = f"real-provider-{state}-long"
    if (row["locale"], row["state"], row["fixture"], row["full_page"]) != (
        locale,
        state,
        expected_fixture,
        True,
    ):
        fail(f"{expected_file} locale/state/fixture/full_page contract mismatch")

    viewport = row["viewport"]
    if not isinstance(viewport, dict):
        fail(f"{expected_file}.viewport must be an object")
    require_keys(viewport, {"name", "width", "height"}, f"{expected_file}.viewport")
    if (viewport["name"], viewport["width"], viewport["height"]) != (
        viewport_name,
        viewport_width,
        viewport_height,
    ):
        fail(f"{expected_file} viewport contract mismatch")

    screenshot_path = output_dir / "screenshots" / expected_file
    if not screenshot_path.is_file() or screenshot_path.stat().st_size == 0:
        fail(f"missing screenshot {screenshot_path}")
    actual_width, actual_height, channels, pixel_rows = png_pixels(screenshot_path)
    if actual_width != viewport_width or actual_height < viewport_height:
        fail(
            f"{expected_file} PNG size={actual_width}x{actual_height}, "
            f"expected width={viewport_width} and full-page height>={viewport_height}"
        )
    require_visual_content(screenshot_path, pixel_rows, actual_width, channels)
    if state != "generating" and viewport_name == "mobile":
        require_visual_content(
            screenshot_path,
            pixel_rows,
            actual_width,
            channels,
            start_row=max(0, actual_height - viewport_height),
            label="ready mobile bottom region",
            maximum_dominant_ratio=0.985,
        )

    screenshot_digest = hashlib.sha256(screenshot_path.read_bytes()).hexdigest()
    if row["screenshot_sha256"] != screenshot_digest:
        fail(f"{expected_file} screenshot_sha256 does not match the PNG SHA-256")

    report_ref = validate_ref(row["report_ref"], f"{expected_file}.report_ref")
    session_ref = validate_ref(row["session_ref"], f"{expected_file}.session_ref")

    evidence = row["evidence"]
    if not isinstance(evidence, dict):
        fail(f"{expected_file}.evidence must be an object")
    require_keys(evidence, {"collection", "db", "api", "content_audit"}, f"{expected_file}.evidence")

    expected_status = "generating" if state == "generating" else "ready"
    expected_preparedness = PREPAREDNESS[state]
    db = evidence["db"]
    require_keys(
        db,
        {"status", "preparedness_level", "frozen_context_digest", "report_content_digest"},
        f"{expected_file}.evidence.db",
    )
    if db["status"] != expected_status or db["preparedness_level"] != expected_preparedness:
        fail(f"{expected_file} DB state does not match captured state")
    digest = validate_digest(db["frozen_context_digest"], f"{expected_file}.evidence.db.frozen_context_digest")
    report_content_digest = db["report_content_digest"]
    if state == "generating":
        if report_content_digest is not None:
            fail(f"{expected_file} generating capture must not claim a report content digest")
    else:
        report_content_digest = validate_digest(
            report_content_digest,
            f"{expected_file}.evidence.db.report_content_digest",
        )

    collection = evidence["collection"]
    if not isinstance(collection, dict):
        fail(f"{expected_file}.evidence.collection must be an object")
    require_keys(
        collection,
        {
            "run_id",
            "method",
            "report_ref",
            "session_ref",
            "frozen_context_digest",
            "report_content_digest",
            "screenshot_sha256",
        },
        f"{expected_file}.evidence.collection",
    )
    if collection != {
        "run_id": run_id,
        "method": "trusted-current-run-db-api-capture",
        "report_ref": report_ref,
        "session_ref": session_ref,
        "frozen_context_digest": digest,
        "report_content_digest": report_content_digest,
        "screenshot_sha256": screenshot_digest,
    }:
        fail(f"{expected_file} DB/API collection does not bind current run/report/session/context/screenshot")

    api = evidence["api"]
    require_keys(
        api,
        {"status", "preparedness_level", "report_content_digest", "source_message_seq_nos_exposed"},
        f"{expected_file}.evidence.api",
    )
    if (
        api["status"] != expected_status
        or api["preparedness_level"] != expected_preparedness
        or api["source_message_seq_nos_exposed"] is not False
    ):
        fail(f"{expected_file} API state/anchor redaction does not match captured state")
    if api["report_content_digest"] != report_content_digest:
        fail(f"{expected_file} DB/API/collection report content digests do not match")

    audit = evidence["content_audit"]
    require_keys(
        audit,
        {
            "fact_to_judgment_to_action",
            "item_verdict_count",
            "unsupported_count",
            "irrelevant_advice_count",
            "causal_mismatch_count",
            "action_label_audit",
        },
        f"{expected_file}.evidence.content_audit",
    )
    if state == "generating":
        if audit != {
            "fact_to_judgment_to_action": "not_applicable",
            "item_verdict_count": 0,
            "unsupported_count": 0,
            "irrelevant_advice_count": 0,
            "causal_mismatch_count": 0,
            "action_label_audit": {
                "language": "not_applicable",
                "unit": "not_applicable",
                "limit": 0,
                "counts": [],
            },
        }:
            fail(f"{expected_file} generating capture must not claim a content verdict")
    elif (
        audit["fact_to_judgment_to_action"] != "closed"
        or not isinstance(audit["item_verdict_count"], int)
        or audit["item_verdict_count"] < 1
        or any(audit[key] != 0 for key in ("unsupported_count", "irrelevant_advice_count", "causal_mismatch_count"))
    ):
        fail(f"{expected_file} ready content audit is not closed and redline-clean")
    if state != "generating":
        label_audit = audit["action_label_audit"]
        if not isinstance(label_audit, dict):
            fail(f"{expected_file} action_label_audit must be an object")
        require_keys(label_audit, {"language", "unit", "limit", "counts"}, f"{expected_file}.action_label_audit")
        expected_language = "zh-CN" if locale == "zh" else "en"
        expected_unit = "code_points" if locale == "zh" else "words"
        expected_limit = 64 if locale == "zh" else 24
        if (
            label_audit["language"] != expected_language
            or label_audit["unit"] != expected_unit
            or label_audit["limit"] != expected_limit
        ):
            fail(f"{expected_file} action label audit does not match locale")
        counts = label_audit["counts"]
        if (
            not isinstance(counts, list)
            or not 1 <= len(counts) <= 2
            or any(not isinstance(count, int) or isinstance(count, bool) or not 1 <= count <= expected_limit for count in counts)
        ):
            fail(f"{expected_file} action label counts exceed the user-facing limit")

    return state, report_ref, session_ref, digest


def validate_count(value: Any, path: str) -> int:
    if not isinstance(value, int) or isinstance(value, bool) or value < 0:
        fail(f"{path} must be a non-negative integer")
    return value


def validate_navigation_url(value: Any, expected_path: str, report_ref: str, path: str) -> None:
    if not isinstance(value, str):
        fail(f"{path} must be a relative reportId-only URL")
    parsed = urllib.parse.urlsplit(value)
    if (
        parsed.scheme
        or parsed.netloc
        or parsed.fragment
        or parsed.path != expected_path
        or urllib.parse.parse_qsl(parsed.query, keep_blank_values=True) != [("reportId", report_ref)]
    ):
        fail(f"{path} must be a relative reportId-only URL")


def validate_navigation_artifact(path: Path, run_id: str, ready_report_refs: set[str]) -> dict[str, str]:
    if not path.is_file():
        fail("missing bounded report conversation navigation artifact")
    try:
        navigation = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        fail(f"invalid report conversation navigation artifact: {exc}")
    if not isinstance(navigation, dict):
        fail("report conversation navigation artifact must be an object")
    scan_redline(navigation)
    require_keys(
        navigation,
        {
            "schema_version",
            "scenario_id",
            "run_id",
            "method",
            "report_ref",
            "urls",
            "request_audit",
            "privacy",
        },
        "conversation_navigation",
    )
    if (
        navigation["schema_version"] != "p0-099-conversation-navigation.v1"
        or navigation["scenario_id"] != "E2E.P0.099"
        or navigation["run_id"] != run_id
        or navigation["method"] != "real-browser-report-conversation-back"
    ):
        fail("report conversation navigation identity/run/method is invalid")
    report_ref = validate_ref(navigation["report_ref"], "conversation_navigation.report_ref")
    try:
        uuid.UUID(report_ref)
    except ValueError:
        fail("conversation_navigation.report_ref must be a UUID")
    if report_ref not in ready_report_refs:
        fail("conversation navigation must bind one current ready report")

    urls = navigation["urls"]
    if not isinstance(urls, dict):
        fail("conversation_navigation.urls must be an object")
    require_keys(urls, {"report", "conversation", "back"}, "conversation_navigation.urls")
    validate_navigation_url(urls["report"], "/report", report_ref, "conversation_navigation.urls.report")
    validate_navigation_url(
        urls["conversation"],
        "/report-conversation",
        report_ref,
        "conversation_navigation.urls.conversation",
    )
    validate_navigation_url(urls["back"], "/report", report_ref, "conversation_navigation.urls.back")

    request_audit = navigation["request_audit"]
    if not isinstance(request_audit, dict):
        fail("conversation_navigation.request_audit must be an object")
    require_keys(
        request_audit,
        {
            "report_get_path",
            "report_get_count",
            "conversation_get_path",
            "conversation_get_count",
            "public_session_list_request_count",
            "route_interception_used",
        },
        "conversation_navigation.request_audit",
    )
    if (
        request_audit["report_get_path"] != f"/api/v1/reports/{report_ref}"
        or request_audit["conversation_get_path"] != f"/api/v1/reports/{report_ref}/conversation"
    ):
        fail("conversation navigation request paths do not bind the reportId-only APIs")
    for key in ("report_get_count", "conversation_get_count"):
        count = validate_count(request_audit[key], f"conversation_navigation.request_audit.{key}")
        if not 1 <= count <= MAX_NAVIGATION_REQUEST_COUNT:
            fail(f"conversation_navigation.request_audit.{key} is outside the bounded browser observation")
    if request_audit["public_session_list_request_count"] != 0:
        fail("conversation navigation observed a forbidden public session-list request")
    if request_audit["route_interception_used"] is not False:
        fail("conversation navigation must not use route interception")

    if navigation["privacy"] != {
        "transcript_prose_written": False,
        "internal_locator_written": False,
        "browser_state_written": False,
    }:
        fail("conversation navigation must persist only bounded redacted facts")
    return {"report_ref": report_ref}


def parse_utc_timestamp(value: Any, path: str) -> datetime:
    if not isinstance(value, str) or not value.endswith("Z"):
        fail(f"{path} must be an RFC3339 UTC timestamp")
    try:
        parsed = datetime.fromisoformat(value[:-1] + "+00:00")
    except ValueError:
        fail(f"{path} must be an RFC3339 UTC timestamp")
    if parsed.utcoffset() is None or parsed.utcoffset().total_seconds() != 0:
        fail(f"{path} must be UTC")
    return parsed


def load_setup_boundary(output_dir: Path, run_id: str) -> datetime:
    path = output_dir / "setup.env"
    if not path.is_file():
        fail("missing setup.env current-run boundary")
    values: dict[str, str] = {}
    try:
        for line in path.read_text(encoding="utf-8").splitlines():
            if not line or "=" not in line:
                continue
            key, value = line.split("=", 1)
            values[key] = value
    except OSError as exc:
        fail(f"setup.env cannot be read: {exc}")
    if values.get("scenario") != "E2E.P0.099" or values.get("RUN_ID") != run_id:
        fail("setup.env scenario/run does not match current validation")
    return parse_utc_timestamp(values.get("setup_at"), "setup.env.setup_at")


def validate_conversation_capture(
    conversation: Any,
    expected: dict[str, dict[str, Any]],
    navigation: dict[str, str],
) -> None:
    if not isinstance(conversation, dict):
        fail("live conversation capture must be an object")
    require_keys(
        conversation,
        {"report_ref", "session_ref", "db", "api"},
        "live_capture.conversation",
    )
    report_ref = validate_ref(conversation["report_ref"], "live_capture.conversation.report_ref")
    if report_ref != navigation["report_ref"] or report_ref not in expected:
        fail("live conversation capture does not bind the current browser-selected report")
    binding = expected[report_ref]
    session_ref = validate_ref(conversation["session_ref"], "live_capture.conversation.session_ref")
    if session_ref != binding["session_ref"] or binding["status"] != "ready":
        fail("live conversation capture report/session/ready binding is invalid")

    db = conversation["db"]
    if not isinstance(db, dict):
        fail("live conversation DB projection must be an object")
    require_keys(
        db,
        {
            "report_status",
            "frozen_context_digest",
            "context_digest",
            "message_count",
            "strict_sequence_digest",
            "ordered_message_digest",
            "read_only",
            "ordered_by",
        },
        "live_capture.conversation.db",
    )
    if (
        db["report_status"] != "ready"
        or db["frozen_context_digest"] != binding["frozen_context_digest"]
        or db["read_only"] is not True
        or db["ordered_by"] != "seq_no ASC"
    ):
        fail("live conversation DB report/context/order binding is invalid")
    for key in ("frozen_context_digest", "context_digest", "strict_sequence_digest", "ordered_message_digest"):
        validate_digest(db[key], f"live_capture.conversation.db.{key}")
    message_count = validate_count(db["message_count"], "live_capture.conversation.db.message_count")
    if message_count < 1:
        fail("live conversation capture must bind at least one ordered message")

    api = conversation["api"]
    if not isinstance(api, dict):
        fail("live conversation API projection must be an object")
    require_keys(
        api,
        {
            "report_status",
            "context_digest",
            "message_count",
            "strict_sequence_digest",
            "ordered_message_digest",
            "authenticated",
            "internal_locator_exposed",
        },
        "live_capture.conversation.api",
    )
    if (
        api["report_status"] != db["report_status"]
        or api["context_digest"] != db["context_digest"]
        or api["message_count"] != message_count
        or api["strict_sequence_digest"] != db["strict_sequence_digest"]
        or api["ordered_message_digest"] != db["ordered_message_digest"]
        or api["authenticated"] is not True
        or api["internal_locator_exposed"] is not False
    ):
        fail("live conversation DB/API projection is not redacted and identical")
    for key in ("context_digest", "strict_sequence_digest", "ordered_message_digest"):
        validate_digest(api[key], f"live_capture.conversation.api.{key}")


def validate_live_capture(
    output_dir: Path,
    run_id: str,
    manifest_rows: list[dict[str, Any]],
    setup_at: datetime,
    navigation: dict[str, str],
) -> None:
    path = output_dir / "live-capture.json"
    if not path.is_file():
        fail("missing independent live capture artifact")
    try:
        capture = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        fail(f"invalid live capture artifact: {exc}")
    if not isinstance(capture, dict):
        fail("live capture artifact must be an object")
    scan_redline(capture)
    require_keys(
        capture,
        {
            "schema_version",
            "scenario_id",
            "run_id",
            "method",
            "captured_at",
            "result",
            "reason_code",
            "reports",
            "conversation",
            "privacy",
        },
        "live_capture",
    )
    if (
        capture["schema_version"] != "p0-099-live-capture.v3"
        or capture["scenario_id"] != "E2E.P0.099"
        or capture["run_id"] != run_id
        or capture["method"] != "authenticated-live-http+read-only-postgres"
        or capture["result"] != "PASS"
        or capture["reason_code"] != "captured"
    ):
        fail("live capture identity/result does not prove the current authenticated DB/API run")
    captured_at = parse_utc_timestamp(capture["captured_at"], "live_capture.captured_at")
    if captured_at <= setup_at:
        fail("live capture timestamp is not part of the current run")
    if capture["privacy"] != {
        "cookie_written": False,
        "database_url_written": False,
        "raw_api_written": False,
        "raw_db_written": False,
        "raw_frozen_context_written": False,
        "raw_conversation_content_written": False,
        "prose_written": False,
    }:
        fail("live DB/API capture privacy contract is not redacted")

    expected: dict[str, dict[str, Any]] = {}
    for row in manifest_rows:
        report_ref = row["report_ref"]
        binding = {
            "session_ref": row["session_ref"],
            "status": row["evidence"]["api"]["status"],
            "preparedness_level": row["evidence"]["api"]["preparedness_level"],
            "canonical_report_content_digest": row["evidence"]["api"]["report_content_digest"],
            "frozen_context_digest": row["evidence"]["db"]["frozen_context_digest"],
            "item_verdict_count": row["evidence"]["content_audit"]["item_verdict_count"],
            "action_label_audit": row["evidence"]["content_audit"]["action_label_audit"],
        }
        if report_ref in expected and expected[report_ref] != binding:
            fail(f"manifest rows disagree before live capture binding for {report_ref}")
        expected[report_ref] = binding

    reports = capture["reports"]
    if not isinstance(reports, list) or len(reports) != 3:
        fail("live capture must contain exactly three report projections")
    actual_refs = [row.get("report_ref") if isinstance(row, dict) else None for row in reports]
    if len(set(actual_refs)) != 3 or set(actual_refs) != set(expected):
        fail("live capture report references do not match the exact manifest resources")

    for index, report in enumerate(reports):
        if not isinstance(report, dict):
            fail(f"live capture report[{index}] must be an object")
        require_keys(
            report,
            {
                "report_ref",
                "session_ref",
                "status",
                "preparedness_level",
                "canonical_report_content_digest",
                "content_shape",
                "action_label_audit",
                "db",
            },
            f"live_capture.reports[{index}]",
        )
        report_ref = validate_ref(report["report_ref"], f"live_capture.reports[{index}].report_ref")
        binding = expected[report_ref]
        if (
            report["session_ref"] != binding["session_ref"]
            or report["status"] != binding["status"]
            or report["preparedness_level"] != binding["preparedness_level"]
            or report["canonical_report_content_digest"] != binding["canonical_report_content_digest"]
            or report["action_label_audit"] != binding["action_label_audit"]
        ):
            fail(f"live capture does not match manifest identity/state/content for {report_ref}")
        if report["canonical_report_content_digest"] is not None:
            validate_digest(
                report["canonical_report_content_digest"],
                f"live_capture.reports[{index}].canonical_report_content_digest",
            )

        db = report["db"]
        if not isinstance(db, dict):
            fail(f"live capture DB projection[{index}] must be an object")
        require_keys(
            db,
            {
                "status",
                "preparedness_level",
                "report_created_at",
                "session_created_at",
                "frozen_context_digest",
                "canonical_report_content_digest",
            },
            f"live_capture.reports[{index}].db",
        )
        report_created_at = parse_utc_timestamp(
            db["report_created_at"],
            f"live_capture.reports[{index}].db.report_created_at",
        )
        session_created_at = parse_utc_timestamp(
            db["session_created_at"],
            f"live_capture.reports[{index}].db.session_created_at",
        )
        if not (
            setup_at < session_created_at <= report_created_at <= captured_at
        ):
            fail(f"live capture DB projection for {report_ref} is not from the current run")
        if (
            db["status"] != binding["status"]
            or db["preparedness_level"] != binding["preparedness_level"]
            or db["frozen_context_digest"] != binding["frozen_context_digest"]
            or db["canonical_report_content_digest"]
            != binding["canonical_report_content_digest"]
        ):
            fail(f"live capture DB projection does not match manifest/API for {report_ref}")
        validate_digest(
            db["frozen_context_digest"],
            f"live_capture.reports[{index}].db.frozen_context_digest",
        )
        if db["canonical_report_content_digest"] is not None:
            validate_digest(
                db["canonical_report_content_digest"],
                f"live_capture.reports[{index}].db.canonical_report_content_digest",
            )

        shape = report["content_shape"]
        if not isinstance(shape, dict):
            fail(f"live_capture.reports[{index}].content_shape must be an object")
        require_keys(
            shape,
            {
                "dimension_assessment_count",
                "highlight_count",
                "issue_count",
                "next_action_count",
                "retry_focus_count",
            },
            f"live_capture.reports[{index}].content_shape",
        )
        dimensions = validate_count(shape["dimension_assessment_count"], f"live_capture.reports[{index}].dimension_assessment_count")
        highlights = validate_count(shape["highlight_count"], f"live_capture.reports[{index}].highlight_count")
        issues = validate_count(shape["issue_count"], f"live_capture.reports[{index}].issue_count")
        actions = validate_count(shape["next_action_count"], f"live_capture.reports[{index}].next_action_count")
        validate_count(shape["retry_focus_count"], f"live_capture.reports[{index}].retry_focus_count")
        if dimensions + highlights + issues != binding["item_verdict_count"]:
            fail(f"live capture content item count does not match manifest audit for {report_ref}")
        if actions != len(binding["action_label_audit"]["counts"]):
            fail(f"live capture action count does not match manifest audit for {report_ref}")

    validate_conversation_capture(capture["conversation"], expected, navigation)


def validate_manual_visual_audit(
    output_dir: Path,
    run_id: str,
    manifest_rows: list[dict[str, Any]],
) -> None:
    path = output_dir / "manual-visual-audit.json"
    if not path.is_file():
        fail("missing exact-six manual visual audit")
    try:
        audit = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        fail(f"invalid manual visual audit: {exc}")
    if not isinstance(audit, dict):
        fail("manual visual audit must be an object")
    scan_redline(audit)
    require_keys(
        audit,
        {
            "schema_version",
            "scenario_id",
            "run_id",
            "method",
            "result",
            "screenshots",
            "privacy",
        },
        "manual_visual_audit",
    )
    if (
        audit["schema_version"] != "p0-099-manual-visual-audit.v1"
        or audit["scenario_id"] != "E2E.P0.099"
        or audit["run_id"] != run_id
        or audit["method"] != "manual-image-review-no-ocr"
        or audit["result"] != "PASS"
    ):
        fail("manual visual audit identity/run/result is invalid")
    if audit["privacy"] != {
        "ocr_used": False,
        "prose_transcribed": False,
        "raw_content_written": False,
    }:
        fail("manual visual audit must use no OCR and persist no prose/raw content")

    rows = audit["screenshots"]
    if not isinstance(rows, list) or len(rows) != 6:
        fail("manual visual audit must contain exactly six screenshot rows")
    expected = {row["file"]: row for row in manifest_rows}
    files = [row.get("file") if isinstance(row, dict) else None for row in rows]
    if len(set(files)) != 6 or set(files) != set(expected):
        fail("manual visual audit must bind exactly six canonical screenshot files")
    for index, row in enumerate(rows):
        if not isinstance(row, dict):
            fail(f"manual visual audit row[{index}] must be an object")
        require_keys(row, {"file", "screenshot_sha256", "checks"}, f"manual_visual_audit.screenshots[{index}]")
        manifest_row = expected[row["file"]]
        if row["screenshot_sha256"] != manifest_row["screenshot_sha256"]:
            fail(f"manual visual audit SHA-256 does not bind current PNG {row['file']}")
        checks = row["checks"]
        if not isinstance(checks, dict):
            fail(f"manual visual audit checks for {row['file']} must be an object")
        expected_checks = (
            GENERATING_VISUAL_CHECKS
            if manifest_row["state"] == "generating"
            else READY_VISUAL_CHECKS
        )
        require_keys(checks, expected_checks, f"manual_visual_audit.screenshots[{index}].checks")
        if any(value is not True for value in checks.values()):
            fail(f"manual visual audit has an unaccepted check for {row['file']}")


def validate(output_dir: Path, run_id: str, automated_only: bool = False) -> None:
    setup_at = load_setup_boundary(output_dir, run_id)
    manifest_path = output_dir / "manifest.json"
    if not manifest_path.is_file():
        fail(f"missing current manifest {manifest_path}")
    try:
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as exc:
        fail(f"invalid manifest.json: {exc}")
    if not isinstance(manifest, dict):
        fail("manifest must be an object")
    scan_redline(manifest)
    require_keys(manifest, {"scenario_id", "run_id", "capture_contract", "screenshots", "privacy"}, "manifest")
    if manifest["scenario_id"] != "E2E.P0.099" or manifest["run_id"] != run_id:
        fail("manifest scenario_id/run_id does not match the current setup run")
    if manifest["capture_contract"] != "report-full-page-v1":
        fail("manifest capture_contract must be report-full-page-v1")

    privacy = manifest["privacy"]
    if privacy != {"redacted": True, "cookie_written": False, "raw_frozen_context_written": False}:
        fail("privacy must prove redaction with no cookie or raw frozen context on disk")

    rows = manifest["screenshots"]
    if not isinstance(rows, list) or len(rows) != 6:
        fail("manifest must contain exactly six screenshot rows")
    files = [row.get("file") if isinstance(row, dict) else None for row in rows]
    if len(set(files)) != 6 or set(files) != set(EXPECTED):
        fail("manifest must contain exactly six canonical screenshot filenames")

    screenshot_dir = output_dir / "screenshots"
    if screenshot_dir.is_symlink():
        fail("screenshots directory must not be a symlink")
    if not screenshot_dir.is_dir():
        fail("missing screenshots directory")
    screenshot_entries = list(screenshot_dir.iterdir())
    if any(path.is_symlink() for path in screenshot_entries):
        fail("canonical screenshots must not be symlinks")
    if (
        len(screenshot_entries) != 6
        or any(not path.is_file() or path.suffix != ".png" for path in screenshot_entries)
        or {path.name for path in screenshot_entries} != set(EXPECTED)
    ):
        fail("screenshots directory must contain exactly six canonical regular PNG files")

    grouped: dict[str, list[tuple[str, str, str]]] = {}
    for row in rows:
        state, report_ref, session_ref, digest = validate_row(row, output_dir, row["file"], run_id)
        grouped.setdefault(state, []).append((report_ref, session_ref, digest))
    if set(grouped) != {"ready-needs-practice", "ready-well-prepared", "generating"}:
        fail("manifest must cover needs-practice, well-prepared, and generating")
    for state, refs in grouped.items():
        if len(refs) != 2 or len(set(refs)) != 1:
            fail(f"{state} desktop/mobile rows must bind to the same report/session/context")
    if len({refs[0][0] for refs in grouped.values()}) != 3 or len({refs[0][1] for refs in grouped.values()}) != 3:
        fail("the three captured states must bind to isolated report/session resources")
    ready_report_refs = {
        refs[0][0]
        for state, refs in grouped.items()
        if state in {"ready-needs-practice", "ready-well-prepared"}
    }
    navigation = validate_navigation_artifact(
        output_dir / "conversation-navigation.json",
        run_id,
        ready_report_refs,
    )
    validate_live_capture(output_dir, run_id, rows, setup_at, navigation)
    if not automated_only:
        validate_manual_visual_audit(output_dir, run_id, rows)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--output-dir", type=Path)
    parser.add_argument("--run-id")
    parser.add_argument("--sanitize-output", type=Path)
    parser.add_argument("--failed", action="store_true")
    parser.add_argument("--setup", action="store_true")
    parser.add_argument("--automated-only", action="store_true")
    args = parser.parse_args()
    if args.sanitize_output is not None:
        removed = sanitize_output(args.sanitize_output, args.failed, args.setup)
        print(
            f"P0_099_PRIVACY_SANITIZE removed={removed} "
            f"failed={str(args.failed).lower()} setup={str(args.setup).lower()}"
        )
        return 0 if args.failed or args.setup or removed == 0 else 1
    if args.output_dir is None or not args.run_id:
        parser.error("--output-dir and --run-id are required for validation")
    try:
        validate(args.output_dir, args.run_id, args.automated_only)
    except EvidenceError as exc:
        print(f"P0.099 evidence invalid: {exc}", file=sys.stderr)
        return 1
    manual = "skipped" if args.automated_only else "bound"
    print(
        "P0_099_SIX_SCREENSHOT_PASS screenshots=6 states=3 "
        "live_binding=pass conversation_navigation=pass db_current_run=pass "
        f"manual_visual_audit={manual} privacy=redacted"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
