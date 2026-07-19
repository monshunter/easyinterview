"""Tests for the minimal plan-context validator and generator contract."""

from __future__ import annotations

import importlib.util
from pathlib import Path

import pytest
import yaml


SHARED_DIR = Path(__file__).resolve().parent.parent / "shared" / "scripts"
VALIDATE_PATH = SHARED_DIR / "validate_context.py"
GENERATE_PATH = SHARED_DIR / "generate_context_yaml.py"


def _load_module(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


def _write_context_fixture(tmp_path: Path):
    docs_root = tmp_path / "docs"
    spec_dir = docs_root / "spec" / "demo"
    plan_dir = spec_dir / "plans" / "001-backend"
    plan_dir.mkdir(parents=True)
    (plan_dir / "plan.md").write_text("# Plan\n", encoding="utf-8")
    (plan_dir / "checklist.md").write_text("# Checklist\n", encoding="utf-8")
    (spec_dir / "spec.md").write_text("# Spec\n", encoding="utf-8")

    payload = {
        "apiVersion": "plancontext.agent.dev/v1alpha1",
        "kind": "PlanContext",
        "metadata": {"name": "001-backend"},
        "spec": {
            "defaultTarget": "backend",
            "targets": {
                "backend": {
                    "plan": "./plan.md",
                    "checklist": "./checklist.md",
                    "spec": "../../spec.md",
                }
            },
        },
    }
    context_path = plan_dir / "context.yaml"
    context_path.write_text(
        yaml.safe_dump(payload, sort_keys=False, allow_unicode=True),
        encoding="utf-8",
    )
    return docs_root, plan_dir, context_path


def _mutate_context(context_path: Path, mutator):
    payload = yaml.safe_load(context_path.read_text(encoding="utf-8"))
    mutator(payload)
    context_path.write_text(
        yaml.safe_dump(payload, sort_keys=False, allow_unicode=True),
        encoding="utf-8",
    )


def _validation_error(tmp_path: Path, mutator):
    validator = _load_module(VALIDATE_PATH, f"validator_{id(mutator)}")
    docs_root, _, context_path = _write_context_fixture(tmp_path)
    _mutate_context(context_path, mutator)
    with pytest.raises(validator.ValidationError) as exc_info:
        validator.validate_context(
            context_path=str(context_path),
            docs_root=str(docs_root),
            target="backend",
        )
    assert exc_info.value.code == 2
    return "\n".join(exc_info.value.lines)


def test_validate_context_accepts_minimal_manifest(tmp_path):
    validator = _load_module(VALIDATE_PATH, "validate_minimal_context")
    docs_root, _, context_path = _write_context_fixture(tmp_path)

    result = validator.validate_context(
        context_path=str(context_path),
        docs_root=str(docs_root),
        target="backend",
    )

    assert result["name"] == "001-backend"
    assert result["target"] == "backend"
    assert {item["role"] for item in result["files"]} == {
        "plan",
        "checklist",
        "spec",
    }
    assert "discovery" not in result
    assert "targetDiscovery" not in result
    assert "baseBranch" not in result
    assert "branch" not in result


@pytest.mark.parametrize(
    "field,value",
    [
        ("subspec", "demo"),
        ("sequence", 1),
        ("specVersion", {"from": None, "to": 1.0}),
        ("baseBranch", "main"),
        ("branch", "demo-branch"),
        ("custom", "value"),
    ],
)
def test_validate_context_rejects_extra_metadata(tmp_path, field, value):
    error_text = _validation_error(
        tmp_path,
        lambda payload: payload["metadata"].__setitem__(field, value),
    )
    assert f"metadata.{field} is not allowed" in error_text


def test_validate_context_rejects_top_level_discovery(tmp_path):
    error_text = _validation_error(
        tmp_path,
        lambda payload: payload["spec"].__setitem__("discovery", {"aliases": ["demo"]}),
    )
    assert "spec.discovery is not allowed" in error_text


@pytest.mark.parametrize(
    "field,value",
    [
        ("discovery", {"packages": ["internal/demo"]}),
        ("references", ["../../spec.md"]),
        ("custom", "value"),
    ],
)
def test_validate_context_rejects_extra_target_fields(tmp_path, field, value):
    error_text = _validation_error(
        tmp_path,
        lambda payload: payload["spec"]["targets"]["backend"].__setitem__(field, value),
    )
    assert f"spec.targets.backend.{field} is not allowed" in error_text


def test_validate_context_rejects_extra_spec_fields(tmp_path):
    error_text = _validation_error(
        tmp_path,
        lambda payload: payload["spec"].__setitem__("custom", "value"),
    )
    assert "spec.custom is not allowed" in error_text


def test_validate_context_includes_optional_link_roles(tmp_path):
    validator = _load_module(VALIDATE_PATH, "validate_optional_links")
    docs_root, plan_dir, context_path = _write_context_fixture(tmp_path)
    for filename in (
        "test-plan.md",
        "test-checklist.md",
        "bdd-plan.md",
        "bdd-checklist.md",
    ):
        (plan_dir / filename).write_text(f"# {filename}\n", encoding="utf-8")

    def add_optional_links(payload):
        target = payload["spec"]["targets"]["backend"]
        target.update(
            {
                "testPlan": "./test-plan.md",
                "testChecklist": "./test-checklist.md",
                "bddPlan": "./bdd-plan.md",
                "bddChecklist": "./bdd-checklist.md",
            }
        )

    _mutate_context(context_path, add_optional_links)
    result = validator.validate_context(
        context_path=str(context_path),
        docs_root=str(docs_root),
        target="backend",
    )
    assert {item["role"] for item in result["files"]} == {
        "plan",
        "checklist",
        "spec",
        "test-plan",
        "test-checklist",
        "bdd-plan",
        "bdd-checklist",
    }


def test_generate_context_yaml_emits_only_minimal_fields(tmp_path):
    generator = _load_module(GENERATE_PATH, "generate_minimal_context")
    docs_root, plan_dir, _ = _write_context_fixture(tmp_path)
    config = generator.scan_directory_targets(
        plan_dir_path=str(plan_dir),
        dir_name=plan_dir.name,
        spec_dir=str(docs_root / "spec"),
        docs_root=str(docs_root),
    )
    assert config is not None
    config["metadata"].update(
        {
            "subspec": "demo",
            "sequence": 1,
            "specVersion": {"from": None, "to": 1.0},
            "baseBranch": "main",
            "branch": "legacy-branch",
        }
    )
    config["discovery"] = {"aliases": ["demo"]}
    config["targets"]["backend"]["discovery"] = {"packages": ["internal/demo"]}
    config["targets"]["backend"]["references"] = ["../../spec.md"]

    rendered = yaml.safe_load(generator.format_yaml("001-backend", config))

    assert rendered["metadata"] == {"name": "001-backend"}
    assert set(rendered["spec"]) == {"defaultTarget", "targets"}
    assert set(rendered["spec"]["targets"]["backend"]) == {
        "plan",
        "checklist",
        "spec",
    }


def test_generate_context_yaml_is_deterministic(tmp_path):
    generator = _load_module(GENERATE_PATH, "generate_deterministic_context")
    docs_root, plan_dir, _ = _write_context_fixture(tmp_path)
    config = generator.scan_directory_targets(
        plan_dir_path=str(plan_dir),
        dir_name=plan_dir.name,
        spec_dir=str(docs_root / "spec"),
        docs_root=str(docs_root),
    )
    assert config is not None
    first = generator.format_yaml("001-backend", config)
    second = generator.format_yaml("001-backend", config)
    assert first == second


def test_generate_context_yaml_preserves_target_identity_and_allowed_links(tmp_path):
    generator = _load_module(GENERATE_PATH, "generate_preserved_target_identity")
    docs_root, plan_dir, context_path = _write_context_fixture(tmp_path)
    _mutate_context(
        context_path,
        lambda payload: payload["spec"]["targets"]["backend"].update(
            {
                "discovery": {"packages": ["internal/demo"]},
                "references": ["../../spec.md"],
            }
        ),
    )
    scanned = generator.scan_directory_targets(
        plan_dir_path=str(plan_dir),
        dir_name=plan_dir.name,
        spec_dir=str(docs_root / "spec"),
        docs_root=str(docs_root),
    )
    assert scanned is not None

    reconciled = generator.reconcile_existing_targets(scanned, str(context_path))
    rendered = yaml.safe_load(generator.format_yaml("001-backend", reconciled))

    assert rendered["spec"]["defaultTarget"] == "backend"
    assert set(rendered["spec"]["targets"]) == {"backend"}
    assert set(rendered["spec"]["targets"]["backend"]) == {
        "plan",
        "checklist",
        "spec",
    }
