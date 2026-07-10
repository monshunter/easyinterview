import os
import subprocess
import tempfile
import textwrap
from pathlib import Path
import unittest


REPO_ROOT = Path(__file__).resolve().parents[2]
ACTIVE_REFERENCE_SUFFIXES = {
    ".go",
    ".json",
    ".md",
    ".mjs",
    ".py",
    ".sh",
    ".toml",
    ".ts",
    ".tsx",
    ".yaml",
    ".yml",
}
ACTIVE_REFERENCE_EXCLUDED_PARTS = {
    ".git",
    ".test-output",
    ".venv",
    "__pycache__",
    "coverage",
    "dist",
    "node_modules",
}
ACTIVE_REFERENCE_EXCLUDED_PREFIXES = (
    ("docs", "bugs"),
    ("docs", "reports"),
    ("docs", "work-journal"),
)


def active_reference_sources() -> dict[Path, str]:
    sources: dict[Path, str] = {}
    scripts_dir = REPO_ROOT / "scripts"
    for path in REPO_ROOT.rglob("*"):
        if path.is_symlink() or not path.is_file():
            continue
        relative = path.relative_to(REPO_ROOT)
        if any(part in ACTIVE_REFERENCE_EXCLUDED_PARTS for part in relative.parts):
            continue
        if any(relative.parts[: len(prefix)] == prefix for prefix in ACTIVE_REFERENCE_EXCLUDED_PREFIXES):
            continue
        if scripts_dir in path.parents and path.name.endswith("_test.py"):
            continue
        if path.suffix not in ACTIVE_REFERENCE_SUFFIXES and path.name != "Makefile":
            continue
        try:
            sources[path] = path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
    return sources


class MakefileDryRunTest(unittest.TestCase):
    def test_dev_stack_dry_run_does_not_call_docker(self):
        with tempfile.TemporaryDirectory() as tmp:
            tmp_path = Path(tmp)
            calls = tmp_path / "docker.calls"
            fake_docker = tmp_path / "docker"
            fake_docker.write_text(
                textwrap.dedent(
                    """\
                    #!/usr/bin/env sh
                    echo "$@" >> "$DOCKER_CALL_LOG"
                    case "$*" in
                      *" ps "*minio-init*) echo "exited/0" ;;
                      *" ps "*) echo "running/healthy" ;;
                      *" info"*) echo "29.0.0" ;;
                    esac
                    exit 0
                    """
                ),
                encoding="utf-8",
            )
            fake_docker.chmod(0o755)

            env = os.environ.copy()
            env["PATH"] = f"{tmp_path}{os.pathsep}{env.get('PATH', '')}"
            env["DOCKER_CALL_LOG"] = str(calls)

            result = subprocess.run(
                ["make", "-n", "dev-up"],
                cwd=REPO_ROOT,
                env=env,
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                timeout=20,
            )

            self.assertEqual(
                result.returncode,
                0,
                msg=f"make -n dev-up failed\nstdout:\n{result.stdout}\nstderr:\n{result.stderr}",
            )
            self.assertFalse(
                calls.exists() and calls.read_text(encoding="utf-8").strip(),
                msg=(
                    "make -n dev-up must not invoke docker/compose; dry-run output was:\n"
                    f"{result.stdout}\n{result.stderr}"
                ),
            )

    def test_codegen_events_check_only_diffs_generated_artifacts(self):
        result = subprocess.run(
            ["make", "-n", "codegen-events-check"],
            cwd=REPO_ROOT,
            text=True,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            timeout=20,
        )

        self.assertEqual(
            result.returncode,
            0,
            msg=f"make -n codegen-events-check failed\nstdout:\n{result.stdout}\nstderr:\n{result.stderr}",
        )
        stdout = result.stdout
        generated_paths = [
            "backend/internal/shared/events/envelope.go",
            "backend/internal/shared/events/events.go",
            "backend/internal/shared/jobs/jobs.go",
            "frontend/src/lib/events/envelope.ts",
            "frontend/src/lib/events/events.ts",
            "frontend/src/lib/jobs/jobs.ts",
        ]
        for rel in generated_paths:
            self.assertIn(str(REPO_ROOT / rel), stdout)

        non_generated_paths = [
            "frontend/src/lib/events/events.test.ts",
            "frontend/src/lib/jobs/jobs.test.ts",
            "shared/events/__fixtures__",
        ]
        for rel in non_generated_paths:
            self.assertNotIn(str(REPO_ROOT / rel), stdout)
        self.assertNotIn(str(REPO_ROOT / "frontend/src/lib/events") + '"', stdout)
        self.assertNotIn(str(REPO_ROOT / "frontend/src/lib/jobs") + '"', stdout)

    def test_lint_wires_ai_profile_coverage_gate(self):
        makefile = (REPO_ROOT / "Makefile").read_text(encoding="utf-8")

        self.assertIn("lint-ai-profile-coverage", makefile)
        self.assertIn("scripts/lint/ai_profile_coverage.py", makefile)
        self.assertRegex(
            makefile,
            r"lint: .*lint-ai-profile-coverage",
            msg="top-level lint must run the A3/F3/Product-UI profile coverage gate",
        )

    def test_lint_wires_mock_contract_gate(self):
        makefile = (REPO_ROOT / "Makefile").read_text(encoding="utf-8")

        self.assertIn("lint-mock-contract", makefile)
        self.assertIn("scripts.mock_contract.fixture_registry_test", makefile)
        self.assertIn("scripts/lint/mock_runtime_boundary.py", makefile)
        self.assertRegex(
            makefile,
            r"lint: .*lint-mock-contract",
            msg="top-level lint must run the mock contract runtime boundary gate",
        )

    def test_lint_wires_core_loop_pruning_surface_gate(self):
        makefile = (REPO_ROOT / "Makefile").read_text(encoding="utf-8")

        self.assertIn("lint-core-loop-pruning-surface", makefile)
        self.assertIn("scripts/lint/core_loop_pruning_surface.py", makefile)
        self.assertRegex(
            makefile,
            r"lint: .*lint-core-loop-pruning-surface",
            msg="top-level lint must run the core-loop runtime/generated pruning surface gate",
        )

    def test_test_wires_all_declared_contract_suites_before_language_suites(self):
        makefile = (REPO_ROOT / "Makefile").read_text(encoding="utf-8")
        node_command = "node --test ui-design/ui-design-contract.test.mjs"
        python_command = "python3 -m pytest scripts .agent-skills -q"
        go_command = 'cd "$(ROOT_DIR)/backend" && go test ./...'
        frontend_command = "pnpm --filter @easyinterview/frontend test"

        self.assertIn(node_command, makefile)
        self.assertIn(python_command, makefile)
        self.assertLess(makefile.index(node_command), makefile.index(python_command))
        self.assertLess(makefile.index(python_command), makefile.index(go_command))
        self.assertLess(makefile.index(go_command), makefile.index(frontend_command))

        requirements = (REPO_ROOT / "requirements-dev.txt").read_text(encoding="utf-8")
        self.assertIn("pytest==9.0.3", requirements.splitlines())
        self.assertIn("PyYAML==6.0.3", requirements.splitlines())

    def test_production_scripts_have_active_references(self):
        scripts_dir = REPO_ROOT / "scripts"
        production_scripts = sorted(
            path
            for path in scripts_dir.rglob("*")
            if path.is_file()
            and (path.suffix in {".go", ".py", ".sh"} or path.parent.name == "git-hooks")
            and not path.name.endswith("_test.py")
        )
        sources = active_reference_sources()

        orphaned = []
        for script in production_scripts:
            if not any(path != script and script.name in source for path, source in sources.items()):
                orphaned.append(str(script.relative_to(REPO_ROOT)))

        self.assertEqual(
            orphaned,
            [],
            msg="production scripts need a current entry point, caller, or owner reference",
        )


if __name__ == "__main__":
    unittest.main()
