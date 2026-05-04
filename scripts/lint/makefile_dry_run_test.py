import os
import subprocess
import tempfile
import textwrap
from pathlib import Path
import unittest


REPO_ROOT = Path(__file__).resolve().parents[2]


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


if __name__ == "__main__":
    unittest.main()
