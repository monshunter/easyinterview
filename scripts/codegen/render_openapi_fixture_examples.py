#!/usr/bin/env python3
"""Project the `default` scenario of every fixture under `openapi/fixtures/`
into a derived `openapi/.generated/openapi-with-fixtures.yaml`, attaching the
fixture body as a named OpenAPI `default` example on the matching response.

Phase 3.1 owner per `002-fixtures-and-mock-source` plan §3 / spec C-9. The
projected file is the artefact Prism / docs-site renderers consume — the
hand-authored `openapi.yaml` itself never carries inline `examples`. The
projection is byte-stable across runs so `git diff --exit-code -- openapi/.generated/openapi-with-fixtures.yaml`
is the idempotency check called for in plan §3.1.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import sys
from collections import OrderedDict
from pathlib import Path
from typing import Iterable, List

import yaml

REPO_ROOT_DEFAULT = Path(__file__).resolve().parents[2]


def _yaml_load(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as f:
        return yaml.safe_load(f)


def _walk_fixtures(fixtures_root: Path) -> dict[str, dict]:
    """Map operationId → fixture JSON object."""
    out: dict[str, dict] = {}
    for tag_dir in sorted(p for p in fixtures_root.iterdir() if p.is_dir()):
        for fixture in sorted(p for p in tag_dir.iterdir() if p.suffix == ".json"):
            with fixture.open("r", encoding="utf-8") as f:
                out[fixture.stem] = json.load(f)
    return out


def _operation_ids(spec: dict) -> list[str]:
    out: list[str] = []
    for _path, methods in (spec.get("paths") or {}).items():
        if not isinstance(methods, dict):
            continue
        for _method, op in methods.items():
            if isinstance(op, dict) and isinstance(op.get("operationId"), str):
                out.append(op["operationId"])
    return sorted(out)


def _attach_example(spec: dict, opid: str, status: int, body) -> bool:
    """Find the operation `opid`, locate its response for `status` (or fall
    through to `default`), and attach `body` as a `default` named example.

    Returns True if attached.
    """
    for path, methods in (spec.get("paths") or {}).items():
        if not isinstance(methods, dict):
            continue
        for method, op in methods.items():
            if not isinstance(op, dict) or op.get("operationId") != opid:
                continue
            responses = op.setdefault("responses", {})
            target_key = str(status) if str(status) in responses else (
                "default" if "default" in responses else None
            )
            if target_key is None:
                return False
            response = responses[target_key]
            if not isinstance(response, dict):
                return False
            content = response.setdefault("content", {})
            json_block = content.setdefault("application/json", {})
            examples = json_block.setdefault("examples", {})
            examples["default"] = {
                "summary": f"Default fixture for {opid}",
                "value": body,
            }
            return True
    return False


# Force the generated YAML to honor insertion order rather than sorting keys
# alphabetically — this is what makes the output stable when fixtures and
# openapi.yaml change.
def _setup_ordered_yaml() -> None:
    def _represent_dict(dumper, data):
        return dumper.represent_mapping("tag:yaml.org,2002:map", data.items())

    yaml.add_representer(dict, _represent_dict, Dumper=yaml.SafeDumper)
    yaml.add_representer(OrderedDict, _represent_dict, Dumper=yaml.SafeDumper)


def render(repo_root: Path) -> Path:
    openapi_path = repo_root / "openapi" / "openapi.yaml"
    fixtures_root = repo_root / "openapi" / "fixtures"
    output_path = repo_root / "openapi" / ".generated" / "openapi-with-fixtures.yaml"

    spec = _yaml_load(openapi_path)
    fixtures = _walk_fixtures(fixtures_root)

    missing: List[str] = []
    for opid in _operation_ids(spec):
        if opid not in fixtures:
            missing.append(f"{opid}: missing fixture default example source")

    for opid, fixture in fixtures.items():
        default = (fixture.get("scenarios") or {}).get("default") or {}
        response = default.get("response") or {}
        status = response.get("status")
        if status is None:
            continue
        # No-body fixtures (logout 204 / startAuthEmailChallenge 202) still
        # need an example slot so consumers can confirm "no body" parity. We
        # write the literal fixture body — `null` or `{}` — as the example
        # value, synthesizing the `content/application-json` block when the
        # source openapi.yaml omitted it (the derived file is only consumed
        # by Prism / docs renderers, not by codegen, so the synthesized slot
        # is contained to this artefact).
        body = response.get("body", None)
        if not _attach_example(spec, opid, int(status), body):
            missing.append(f"{opid}: no response slot found for status {status}")

    if missing:
        for m in missing:
            print(f"render-openapi-fixture-examples: {m}", file=sys.stderr)
        raise RuntimeError(
            f"render-openapi-fixture-examples: failed to project {len(missing)} example(s)"
        )

    output_path.parent.mkdir(parents=True, exist_ok=True)
    with output_path.open("w", encoding="utf-8") as f:
        f.write(
            "# Generated by scripts/codegen/render_openapi_fixture_examples.py.\n"
            "# Do not edit by hand. Source of truth: openapi/openapi.yaml +\n"
            "# openapi/fixtures/<tag>/<operationId>.json (scenarios.default).\n"
            "# Re-run `make render-openapi-fixture-examples` to refresh.\n\n"
        )
        yaml.safe_dump(
            spec, f, sort_keys=False, allow_unicode=True, default_flow_style=False,
            width=120,
        )
    return output_path


def main(argv: Iterable[str]) -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--repo-root", type=Path, default=REPO_ROOT_DEFAULT)
    args = parser.parse_args(list(argv))
    repo_root = args.repo_root.resolve()
    _setup_ordered_yaml()
    try:
        out = render(repo_root)
    except RuntimeError as e:
        print(str(e), file=sys.stderr)
        return 1
    digest = hashlib.sha256(out.read_bytes()).hexdigest()
    print(f"render-openapi-fixture-examples: wrote {out} (sha256={digest[:12]}…)")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
