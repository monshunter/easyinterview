"""Tests for shared plan-context validator and generator helpers."""

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


def _write_context_fixture(
    tmp_path: Path,
    discovery=None,
    target_discovery=None,
    metadata_overrides=None,
):
    docs_root = tmp_path / "docs"
    spec_dir = docs_root / "spec" / "demo"
    plan_dir = spec_dir / "plans" / "001-backend"
    plan_dir.mkdir(parents=True)
    spec_dir.mkdir(parents=True, exist_ok=True)

    (plan_dir / "plan.md").write_text(
        "# Plan\n\n"
        "> **版本**: 1.0\n"
        "> **状态**: active\n"
        "> **更新日期**: 2026-04-04\n",
        encoding="utf-8",
    )
    (plan_dir / "checklist.md").write_text(
        "# Checklist\n\n"
        "> **版本**: 1.0\n"
        "> **状态**: active\n"
        "> **更新日期**: 2026-04-04\n\n"
        "## Phase 1\n\n- [ ] 1.1 Item\n",
        encoding="utf-8",
    )
    (spec_dir / "spec.md").write_text("# Spec\n", encoding="utf-8")

    target = {
        "plan": "./plan.md",
        "checklist": "./checklist.md",
        "spec": "../../spec.md",
    }
    if target_discovery is not None:
        target["discovery"] = target_discovery

    metadata = {
        "subspec": "demo",
        "name": "001-backend",
        "sequence": 1,
        "supersedes": [],
        "specVersion": {"from": None, "to": 1.0},
    }
    if metadata_overrides:
        metadata.update(metadata_overrides)

    payload = {
        "apiVersion": "plancontext.agent.dev/v1alpha1",
        "kind": "PlanContext",
        "metadata": metadata,
        "spec": {
            "defaultTarget": "backend",
            "targets": {"backend": target},
        },
    }
    if discovery is not None:
        payload["spec"]["discovery"] = discovery

    context_path = plan_dir / "context.yaml"
    with open(context_path, "w", encoding="utf-8") as f:
        yaml.safe_dump(payload, f, sort_keys=False, allow_unicode=True)

    return docs_root, plan_dir, context_path


def test_validate_context_accepts_spec_centric_manifest_without_discovery(tmp_path):
    validator = _load_module(VALIDATE_PATH, "validate_context")
    docs_root, _, context_path = _write_context_fixture(tmp_path)

    result = validator.validate_context(
        context_path=str(context_path),
        docs_root=str(docs_root),
        target="backend",
    )

    assert result["target"] == "backend"
    assert any(item["role"] == "plan" for item in result["files"])


def test_validate_context_accepts_top_and_target_discovery(tmp_path):
    validator = _load_module(VALIDATE_PATH, "validate_context_with_discovery")
    docs_root, _, context_path = _write_context_fixture(
        tmp_path,
        discovery={
            "aliases": ["demo", "change-intake"],
            "keywords": ["issue intake"],
            "relatedBugs": ["BUG-0042"],
            "relatedSpecs": ["../../spec.md"],
        },
        target_discovery={
            "packages": ["internal/practice"],
            "uiRoutes": ["/login"],
            "apiNames": ["RefreshSession"],
        },
    )

    result = validator.validate_context(
        context_path=str(context_path),
        docs_root=str(docs_root),
        target="backend",
    )

    assert result["name"] == "001-backend"
    assert result["defaultTarget"] == "backend"
    assert result["discovery"]["aliases"] == ["demo", "change-intake"]
    assert result["targetDiscovery"]["packages"] == ["internal/practice"]
    assert result["targetDiscovery"]["uiRoutes"] == ["/login"]
    assert result["targetDiscovery"]["apiNames"] == ["RefreshSession"]


def test_validate_context_includes_bdd_checklist_role(tmp_path):
    validator = _load_module(VALIDATE_PATH, "validate_context_with_bdd_checklist")
    docs_root, plan_dir, context_path = _write_context_fixture(tmp_path)

    (plan_dir / "bdd-plan.md").write_text("# BDD Plan\n", encoding="utf-8")
    (plan_dir / "bdd-checklist.md").write_text(
        "# BDD Checklist\n\n## E2E.P0.001\n\n- [x] 创建场景目录\n",
        encoding="utf-8",
    )
    payload = yaml.safe_load(context_path.read_text(encoding="utf-8"))
    target = payload["spec"]["targets"]["backend"]
    target["bddPlan"] = "./bdd-plan.md"
    target["bddChecklist"] = "./bdd-checklist.md"
    context_path.write_text(
        yaml.safe_dump(payload, sort_keys=False, allow_unicode=True),
        encoding="utf-8",
    )

    result = validator.validate_context(
        context_path=str(context_path),
        docs_root=str(docs_root),
        target="backend",
    )

    roles = {item["role"] for item in result["files"]}
    assert "bdd-plan" in roles
    assert "bdd-checklist" in roles


def test_validate_context_rejects_wrong_api_version(tmp_path):
    validator = _load_module(VALIDATE_PATH, "validate_context_wrong_api_version")
    docs_root, _, context_path = _write_context_fixture(tmp_path)

    payload = yaml.safe_load(context_path.read_text(encoding="utf-8"))
    payload["apiVersion"] = "agent.interviewpilot.dev/v1alpha1"
    context_path.write_text(
        yaml.safe_dump(payload, sort_keys=False, allow_unicode=True),
        encoding="utf-8",
    )

    with pytest.raises(validator.ValidationError) as exc_info:
        validator.validate_context(
            context_path=str(context_path),
            docs_root=str(docs_root),
            target="backend",
        )

    assert exc_info.value.code == 2
    assert "plancontext.agent.dev/v1alpha1" in "\n".join(exc_info.value.lines)


def test_validate_context_rejects_commands_discovery(tmp_path):
    validator = _load_module(VALIDATE_PATH, "validate_context_commands_discovery")
    docs_root, _, context_path = _write_context_fixture(
        tmp_path,
        target_discovery={
            "packages": ["internal/practice"],
            "commands": ["make test-e2e"],
        },
    )

    with pytest.raises(validator.ValidationError) as exc_info:
        validator.validate_context(
            context_path=str(context_path),
            docs_root=str(docs_root),
            target="backend",
        )

    assert exc_info.value.code == 2
    assert "spec.targets.backend.discovery.commands is deprecated and must not be used" in "\n".join(
        exc_info.value.lines
    )


def test_validate_context_rejects_bad_discovery_shape(tmp_path):
    validator = _load_module(VALIDATE_PATH, "validate_context_invalid_discovery")
    docs_root, _, context_path = _write_context_fixture(
        tmp_path,
        discovery={"aliases": "demo"},
    )

    with pytest.raises(validator.ValidationError) as exc_info:
        validator.validate_context(
            context_path=str(context_path),
            docs_root=str(docs_root),
            target="backend",
        )

    assert exc_info.value.code == 2
    assert "spec.discovery.aliases must be a list of strings" in "\n".join(exc_info.value.lines)


def test_validate_context_rejects_missing_referenced_markdown(tmp_path):
    validator = _load_module(VALIDATE_PATH, "validate_context_missing_reference")
    docs_root, _, context_path = _write_context_fixture(tmp_path)

    payload = yaml.safe_load(context_path.read_text(encoding="utf-8"))
    payload["spec"]["targets"]["backend"]["references"] = ["./missing-reference.md"]
    context_path.write_text(
        yaml.safe_dump(payload, sort_keys=False, allow_unicode=True),
        encoding="utf-8",
    )

    with pytest.raises(validator.ValidationError) as exc_info:
        validator.validate_context(
            context_path=str(context_path),
            docs_root=str(docs_root),
            target="backend",
        )

    assert exc_info.value.code == 3
    assert "missing-reference.md" in "\n".join(exc_info.value.lines)


def test_validate_context_rejects_reference_outside_docs_boundary(tmp_path):
    validator = _load_module(VALIDATE_PATH, "validate_context_boundary_reference")
    docs_root, _, context_path = _write_context_fixture(tmp_path)

    payload = yaml.safe_load(context_path.read_text(encoding="utf-8"))
    payload["spec"]["targets"]["backend"]["references"] = ["../../../../../outside-docs.md"]
    context_path.write_text(
        yaml.safe_dump(payload, sort_keys=False, allow_unicode=True),
        encoding="utf-8",
    )

    with pytest.raises(validator.ValidationError) as exc_info:
        validator.validate_context(
            context_path=str(context_path),
            docs_root=str(docs_root),
            target="backend",
        )

    assert exc_info.value.code == 4
    error_text = "\n".join(exc_info.value.lines)
    assert "Path escapes docs/ boundary" in error_text
    assert "outside-docs.md" in error_text


def test_validate_context_includes_optional_branch_metadata(tmp_path):
    validator = _load_module(VALIDATE_PATH, "validate_context_with_branch_metadata")
    docs_root, _, context_path = _write_context_fixture(
        tmp_path,
        metadata_overrides={
            "baseBranch": "dev",
            "branch": "execution-automation-closure",
        },
    )

    result = validator.validate_context(
        context_path=str(context_path),
        docs_root=str(docs_root),
        target="backend",
    )

    assert result["baseBranch"] == "dev"
    assert result["branch"] == "execution-automation-closure"


def test_validate_context_rejects_non_string_branch_metadata(tmp_path):
    validator = _load_module(VALIDATE_PATH, "validate_context_invalid_branch_metadata")
    docs_root, _, context_path = _write_context_fixture(
        tmp_path,
        metadata_overrides={
            "baseBranch": ["dev"],
            "branch": 42,
        },
    )

    with pytest.raises(validator.ValidationError) as exc_info:
        validator.validate_context(
            context_path=str(context_path),
            docs_root=str(docs_root),
            target="backend",
        )

    assert exc_info.value.code == 2
    error_text = "\n".join(exc_info.value.lines)
    assert "metadata.baseBranch must be a string" in error_text
    assert "metadata.branch must be a string" in error_text


def test_generate_context_yaml_preserves_manual_discovery(tmp_path):
    generator = _load_module(GENERATE_PATH, "generate_context_yaml")
    docs_root, plan_dir, context_path = _write_context_fixture(
        tmp_path,
        discovery={
            "aliases": ["demo"],
            "keywords": ["manual keyword"],
            "customSignals": ["do-not-drop"],
        },
        target_discovery={
            "packages": ["internal/demo"],
            "custom": ["manual-target-signal"],
        },
    )

    config = generator.scan_directory_targets(
        plan_dir_path=str(plan_dir),
        dir_name=plan_dir.name,
        spec_dir=str(docs_root / "spec"),
        docs_root=str(docs_root),
    )
    config = generator.normalize_target_config(config)
    existing = generator.load_existing_manifest(str(context_path))
    config = generator.merge_preserved_discovery(config, existing)
    rendered = yaml.safe_load(generator.format_yaml("001-backend", config))

    assert rendered["spec"]["discovery"]["keywords"] == ["manual keyword"]
    assert rendered["spec"]["discovery"]["customSignals"] == ["do-not-drop"]
    assert rendered["spec"]["targets"]["backend"]["discovery"]["packages"] == ["internal/demo"]
    assert rendered["spec"]["targets"]["backend"]["discovery"]["custom"] == ["manual-target-signal"]
    assert rendered["metadata"]["subspec"] == "demo"
    assert rendered["metadata"]["name"] == "001-backend"


def test_generate_context_yaml_uses_shared_api_version(tmp_path):
    generator = _load_module(GENERATE_PATH, "generate_context_yaml_shared_api_version")
    docs_root, plan_dir, context_path = _write_context_fixture(tmp_path)

    config = generator.scan_directory_targets(
        plan_dir_path=str(plan_dir),
        dir_name=plan_dir.name,
        spec_dir=str(docs_root / "spec"),
        docs_root=str(docs_root),
    )
    config = generator.normalize_target_config(config)
    existing = generator.load_existing_manifest(str(context_path))
    config = generator.merge_preserved_discovery(config, existing)
    rendered = yaml.safe_load(generator.format_yaml("001-backend", config))

    assert rendered["apiVersion"] == "plancontext.agent.dev/v1alpha1"


def test_generate_context_yaml_drops_deprecated_commands_discovery(tmp_path):
    generator = _load_module(GENERATE_PATH, "generate_context_yaml_drop_commands")
    docs_root, plan_dir, context_path = _write_context_fixture(
        tmp_path,
        target_discovery={
            "packages": ["internal/demo"],
            "commands": ["make test-e2e"],
            "custom": ["manual-target-signal"],
        },
    )

    config = generator.scan_directory_targets(
        plan_dir_path=str(plan_dir),
        dir_name=plan_dir.name,
        spec_dir=str(docs_root / "spec"),
        docs_root=str(docs_root),
    )
    config = generator.normalize_target_config(config)
    existing = generator.load_existing_manifest(str(context_path))
    config = generator.merge_preserved_discovery(config, existing)
    rendered = yaml.safe_load(generator.format_yaml("001-backend", config))

    target_discovery = rendered["spec"]["targets"]["backend"]["discovery"]
    assert target_discovery["packages"] == ["internal/demo"]
    assert target_discovery["custom"] == ["manual-target-signal"]
    assert "commands" not in target_discovery


def test_generate_context_yaml_preserves_branch_metadata(tmp_path):
    generator = _load_module(GENERATE_PATH, "generate_context_yaml_branch_metadata")
    docs_root, plan_dir, context_path = _write_context_fixture(
        tmp_path,
        metadata_overrides={
            "baseBranch": "dev",
            "branch": "execution-automation-closure",
        },
    )

    config = generator.scan_directory_targets(
        plan_dir_path=str(plan_dir),
        dir_name=plan_dir.name,
        spec_dir=str(docs_root / "spec"),
        docs_root=str(docs_root),
    )
    config = generator.normalize_target_config(config)
    existing = generator.load_existing_manifest(str(context_path))
    config = generator.merge_preserved_discovery(config, existing)
    rendered = yaml.safe_load(generator.format_yaml("001-backend", config))

    assert rendered["metadata"]["baseBranch"] == "dev"
    assert rendered["metadata"]["branch"] == "execution-automation-closure"


def test_generate_context_yaml_promotes_bdd_plan_and_checklist(tmp_path):
    generator = _load_module(GENERATE_PATH, "generate_context_yaml_bdd_checklist")
    docs_root, plan_dir, context_path = _write_context_fixture(tmp_path)

    (plan_dir / "bdd-plan.md").write_text("# BDD Plan\n", encoding="utf-8")
    (plan_dir / "bdd-checklist.md").write_text("# BDD Checklist\n", encoding="utf-8")

    config = generator.scan_directory_targets(
        plan_dir_path=str(plan_dir),
        dir_name=plan_dir.name,
        spec_dir=str(docs_root / "spec"),
        docs_root=str(docs_root),
    )
    config = generator.normalize_target_config(config)
    existing = generator.load_existing_manifest(str(context_path))
    config = generator.merge_preserved_discovery(config, existing)
    rendered = yaml.safe_load(generator.format_yaml("001-backend", config))

    target = rendered["spec"]["targets"]["backend"]
    assert target["bddPlan"] == "./bdd-plan.md"
    assert target["bddChecklist"] == "./bdd-checklist.md"
