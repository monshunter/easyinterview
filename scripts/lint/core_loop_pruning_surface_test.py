import ast
from pathlib import Path
import subprocess
import sys

import core_loop_pruning_surface as audit
import runtime_topology as runtime_topology_lint


SCRIPT = Path(__file__).resolve().with_name("core_loop_pruning_surface.py")
RUNTIME_TOPOLOGY_SCRIPT = SCRIPT.with_name("runtime_topology.py")
REPO_ROOT = SCRIPT.parents[2]
PYTHON_IMPORT_ROOTS = (
    REPO_ROOT / ".agent-skills",
    REPO_ROOT / "scripts",
    REPO_ROOT / "test" / "scenarios",
)


def strict_lifecycle_tokens() -> tuple[str, ...]:
    return (
        "ret" "ired",
        "Ret" "ired",
        "de" "precated",
        "De" "precated",
        "退" "役",
    )


def obsolete_zh_status_tokens() -> tuple[str, ...]:
    return (
        "废" "弃",
    )


def old_scope_context_tokens() -> tuple[str, ...]:
    return (
        "\u65e7",
        "\u5386" "\u53f2",
    )


def write(path: Path, body: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(body, encoding="utf-8")


def bucket_paths(report: audit.AuditReport, bucket: str) -> list[str]:
    return [finding.path for finding in report.buckets[bucket]]


def unread_module_imports(path: Path) -> list[str]:
    tree = ast.parse(path.read_text(encoding="utf-8"), filename=str(path))
    loaded_names = {
        node.id
        for node in ast.walk(tree)
        if isinstance(node, ast.Name) and isinstance(node.ctx, ast.Load)
    }
    findings = []
    for node in tree.body:
        if isinstance(node, ast.Import):
            aliases = (
                (alias.asname or alias.name.split(".", 1)[0], alias.name)
                for alias in node.names
            )
        elif isinstance(node, ast.ImportFrom) and node.module != "__future__":
            aliases = (
                (alias.asname or alias.name, alias.name)
                for alias in node.names
                if alias.name != "*"
            )
        else:
            continue

        for bound_name, imported_name in aliases:
            if bound_name not in loaded_names:
                findings.append(f"{path.relative_to(REPO_ROOT)}:{node.lineno}:{imported_name}")
    return findings


def test_repo_owned_python_module_imports_are_consumed() -> None:
    missing_roots = [root for root in PYTHON_IMPORT_ROOTS if not root.is_dir()]
    assert missing_roots == []

    findings = []
    for root in PYTHON_IMPORT_ROOTS:
        for path in sorted(root.rglob("*.py")):
            if path.name == "__init__.py" or {".venv", "__pycache__"} & set(path.parts):
                continue
            findings.extend(unread_module_imports(path))

    assert findings == []


def test_frontend_runtime_config_has_single_generated_client_consumer() -> None:
    old_package = REPO_ROOT / "frontend" / "src" / "lib" / "runtime-config"
    assert not old_package.exists(), f"parallel runtime-config package remains: {old_package}"

    provider = (
        REPO_ROOT
        / "frontend"
        / "src"
        / "app"
        / "runtime"
        / "AppRuntimeProvider.tsx"
    ).read_text(encoding="utf-8")
    assert 'from "../../api/generated/client"' in provider
    assert 'from "../../api/generated/types"' in provider
    assert ".getRuntimeConfig(" in provider

    current_docs = (
        REPO_ROOT / "frontend" / "README.md",
        REPO_ROOT / "docs" / "spec" / "secrets-and-config" / "spec.md",
        REPO_ROOT
        / "docs"
        / "spec"
        / "secrets-and-config"
        / "plans"
        / "001-bootstrap"
        / "plan.md",
        REPO_ROOT
        / "docs"
        / "spec"
        / "secrets-and-config"
        / "plans"
        / "001-bootstrap"
        / "context.yaml",
        REPO_ROOT
        / "docs"
        / "spec"
        / "frontend-shell"
        / "plans"
        / "001-app-shell-auth-settings"
        / "context.yaml",
    )
    stale = [
        path.relative_to(REPO_ROOT).as_posix()
        for path in current_docs
        if "src/lib/runtime-config" in path.read_text(encoding="utf-8")
    ]
    assert stale == []


def test_frontend_openapi_codegen_omits_raw_spec_snapshot() -> None:
    generated = REPO_ROOT / "frontend" / "src" / "api" / "generated"
    assert (generated / "client.ts").is_file()
    assert (generated / "types.ts").is_file()
    assert not (generated / "spec.ts").exists()
    assert not (REPO_ROOT / "openapi" / "templates" / "ts" / "spec.tmpl").exists()

    renderer = (
        REPO_ROOT / "backend" / "cmd" / "codegen" / "openapi" / "render_ts.go"
    ).read_text(encoding="utf-8")
    forbidden = (
        "SpecYAMLLiteral",
        "tsTemplateStringLiteral",
        '"spec.tmpl"',
        '"spec.ts"',
    )
    assert [token for token in forbidden if token in renderer] == []


def test_lint_rule_sources_assemble_strict_lifecycle_terms_without_direct_literals() -> None:
    for script in (SCRIPT, RUNTIME_TOPOLOGY_SCRIPT):
        text = script.read_text(encoding="utf-8")
        leaked_tokens = [
            token
            for token in strict_lifecycle_tokens() + obsolete_zh_status_tokens() + old_scope_context_tokens()
            if token in text
        ]

        assert leaked_tokens == [], f"{script.name} leaks direct strict lifecycle tokens: {leaked_tokens}"


def test_negative_context_regexes_match_strict_lifecycle_terms() -> None:
    for token in strict_lifecycle_tokens():
        assert audit.NEGATIVE_CONTEXT_RE.search(token)
        assert runtime_topology_lint.OWNER_NEGATIVE_CONTEXT_RE.search(token)
    for token in obsolete_zh_status_tokens():
        assert audit.NEGATIVE_CONTEXT_RE.search(token)
    for token in old_scope_context_tokens():
        assert audit.NEGATIVE_CONTEXT_RE.search(token)
        assert runtime_topology_lint.OWNER_NEGATIVE_CONTEXT_RE.search(token)


def test_buckets_out_of_scope_surface_hits_by_allowed_context(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "migrations" / "000009_jd_match_baseline.up.sql",
        "CREATE TABLE jd_match_recommendations (id uuid);\n",
    )
    write(
        repo / "frontend" / "src" / "app" / "normalizeRoute.ts",
        'const aliases = { debrief: "home", profile: "home", jd_match: "home" };\n',
    )
    write(
        repo / "frontend" / "src" / "app" / "routeUrl.test.ts",
        'expect(normalizeRoute({ name: "debrief" })).toEqual({ name: "home" });\n',
    )
    write(
        repo / "backend" / "internal" / "api" / "debriefs" / "handler.go",
        "package debriefs\n",
    )

    report = audit.scan_repo(repo)

    assert bucket_paths(report, "migration_records") == [
        "migrations/000009_jd_match_baseline.up.sql"
    ]
    assert bucket_paths(report, "out_of_scope_normalization") == [
        "frontend/src/app/normalizeRoute.ts"
    ]
    assert bucket_paths(report, "negative_tests") == [
        "frontend/src/app/routeUrl.test.ts"
    ]
    assert bucket_paths(report, "real_residuals") == [
        "backend/internal/api/debriefs/handler.go"
    ]


def test_negative_context_words_do_not_allow_production_runtime_residuals(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "backend" / "internal" / "api" / "debriefs" / "handler.go",
        "package debriefs // delete this debrief handler before release\n",
    )

    report = audit.scan_repo(repo)

    assert bucket_paths(report, "real_residuals") == [
        "backend/internal/api/debriefs/handler.go"
    ]
    assert bucket_paths(report, "negative_tests") == []


def test_cli_fails_when_real_residuals_exist(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "config" / "prompts" / "jd_match.search" / "v0.1.0.yaml",
        "feature_key: jd_match.search\n",
    )

    result = subprocess.run(
        [sys.executable, str(SCRIPT), "--repo-root", str(repo)],
        check=False,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    assert result.returncode == 1
    assert "real_residuals (1)" in result.stderr
    assert "config/prompts/jd_match.search/v0.1.0.yaml" in result.stderr


def test_cli_passes_with_only_allowed_buckets(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "migrations" / "000014_drop_jd_match_module.up.sql",
        "DROP TABLE IF EXISTS jd_match_recommendations;\n",
    )
    write(
        repo / "frontend" / "src" / "app" / "routeUrl.ts",
        'const outOfScope = { "/debrief": "home", "/profile": "home" };\n',
    )
    write(
        repo / "scripts" / "lint" / "core_loop_pruning_surface_test.py",
        "assert 'candidate_profiles' not in sql\n",
    )

    result = subprocess.run(
        [sys.executable, str(SCRIPT), "--repo-root", str(repo)],
        check=False,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    assert result.returncode == 0, result.stderr
    assert "migration_records (1)" in result.stdout
    assert "out_of_scope_normalization (1)" in result.stdout
    assert "negative_tests (1)" in result.stdout
    assert "real_residuals (0)" in result.stdout


def test_model_profile_paths_are_not_candidate_profile_surface(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "backend" / "internal" / "ai" / "aiclient" / "profile" / "loader.go",
        'package profile\n\nconst env = "AI_MODEL_PROFILE_PATH"\n',
    )
    write(
        repo / "backend" / "internal" / "ai" / "aiclient" / "README.md",
        "Model Profile schema lives in [`profile/`](./profile/loader.go).\n",
    )

    report = audit.scan_repo(repo)

    assert report.buckets["real_residuals"] == []
    assert all(not findings for findings in report.buckets.values())


def test_frontend_contract_tests_are_negative_test_bucket(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "frontend" / "src" / "app" / "scope.test.ts",
        'assert.doesNotMatch(app, /nav\\("profile"\\)|nav\\("debrief"\\)/);\n',
    )

    report = audit.scan_repo(repo)

    assert bucket_paths(report, "negative_tests") == [
        "frontend/src/app/scope.test.ts"
    ]
    assert report.buckets["real_residuals"] == []
