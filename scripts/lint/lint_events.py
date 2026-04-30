#!/usr/bin/env python3
"""Local lint and breaking-change gate for event/outbox contract assets."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any

import yaml


BREAKING = "breaking change requires eventVersion + 1"


def _by_event_name(data: dict[str, Any]) -> dict[str, dict[str, Any]]:
    return {
        event["name"]: event
        for event in data.get("events", [])
        if isinstance(event, dict) and isinstance(event.get("name"), str)
    }


def _by_job_type(data: dict[str, Any]) -> dict[str, dict[str, Any]]:
    return {
        job["canonical"]: job
        for job in data.get("jobs", [])
        if isinstance(job, dict) and isinstance(job.get("canonical"), str)
    }


def compare_events_baseline(current: dict[str, Any], baseline: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    current_events = _by_event_name(current)
    baseline_events = _by_event_name(baseline)

    for name in sorted(set(baseline_events) - set(current_events)):
        errors.append(f"{name}: deleted eventName is a breaking change; {BREAKING}")
    for name, baseline_event in baseline_events.items():
        current_event = current_events.get(name)
        if not current_event:
            continue
        baseline_payload = baseline_event.get("requiredPayload") or {}
        current_payload = current_event.get("requiredPayload") or {}
        for field in sorted(set(baseline_payload) - set(current_payload)):
            errors.append(f"{name}.{field}: deleted required payload field; {BREAKING}")
        for field, baseline_spec in baseline_payload.items():
            current_spec = current_payload.get(field)
            if not isinstance(current_spec, dict) or not isinstance(baseline_spec, dict):
                continue
            if current_spec.get("type") != baseline_spec.get("type"):
                errors.append(
                    f"{name}.{field}: required payload type changed "
                    f"{baseline_spec.get('type')!r} -> {current_spec.get('type')!r}; {BREAKING}"
                )
    return errors


def compare_jobs_baseline(current: dict[str, Any], baseline: dict[str, Any]) -> list[str]:
    errors: list[str] = []
    current_jobs = _by_job_type(current)
    baseline_jobs = _by_job_type(baseline)

    for canonical in sorted(set(baseline_jobs) - set(current_jobs)):
        errors.append(f"{canonical}: deleted jobType is a breaking change; {BREAKING}")
    for canonical, baseline_job in baseline_jobs.items():
        current_job = current_jobs.get(canonical)
        if not current_job:
            continue
        if current_job.get("apiFacing") != baseline_job.get("apiFacing"):
            errors.append(
                f"{canonical}.apiFacing changed "
                f"{baseline_job.get('apiFacing')!r} -> {current_job.get('apiFacing')!r}; {BREAKING}"
            )
        if current_job.get("asynqTask") != baseline_job.get("asynqTask"):
            errors.append(
                f"{canonical}.asynqTask changed "
                f"{baseline_job.get('asynqTask')!r} -> {current_job.get('asynqTask')!r}; {BREAKING}"
            )
    baseline_subset = baseline.get("apiFacingSubset") or []
    current_subset = current.get("apiFacingSubset") or []
    if current_subset != baseline_subset:
        errors.append(f"apiFacingSubset changed {baseline_subset!r} -> {current_subset!r}; {BREAKING}")
    return errors


def _load_yaml(path: Path) -> dict[str, Any]:
    data = yaml.safe_load(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise ValueError(f"{path} root must be a mapping")
    return data


def _load_json(path: Path) -> dict[str, Any]:
    data = json.loads(path.read_text(encoding="utf-8"))
    if not isinstance(data, dict):
        raise ValueError(f"{path} root must be a mapping")
    return data


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", default=str(Path(__file__).resolve().parents[2]))
    args = parser.parse_args()

    root = Path(args.repo_root).resolve()
    try:
        current_events = _load_yaml(root / "shared/events.yaml")
        current_jobs = _load_yaml(root / "shared/jobs.yaml")
        baseline_events = _load_json(root / "shared/events/baseline/events.v1.json")
        baseline_jobs = _load_json(root / "shared/jobs/baseline/jobs.v1.json")
    except (OSError, ValueError, yaml.YAMLError, json.JSONDecodeError) as exc:
        print(f"FAIL: {exc}", file=sys.stderr)
        return 2

    errors = compare_events_baseline(current_events, baseline_events)
    errors.extend(compare_jobs_baseline(current_jobs, baseline_jobs))
    if errors:
        for error in errors:
            print(f"FAIL: {error}", file=sys.stderr)
        return 1
    print("OK: event/job baselines match current truth sources")
    return 0


if __name__ == "__main__":
    sys.exit(main())
