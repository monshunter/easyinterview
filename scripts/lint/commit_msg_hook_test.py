import subprocess
import tempfile
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[2]
HOOK = REPO_ROOT / "scripts" / "git-hooks" / "commit-msg"


class CommitMsgHookTest(unittest.TestCase):
    def run_hook(self, message: str) -> subprocess.CompletedProcess[str]:
        with tempfile.NamedTemporaryFile("w", encoding="utf-8", delete=False) as handle:
            handle.write(message)
            path = Path(handle.name)
        try:
            return subprocess.run(
                [str(HOOK), str(path)],
                cwd=REPO_ROOT,
                text=True,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                check=False,
            )
        finally:
            path.unlink(missing_ok=True)

    def test_accepts_ascii_english_message(self) -> None:
        result = self.run_hook(
            "feat(work-journal): require english commit messages\n\n"
            "- Add work journal validation\n"
            "- Keep commit text ASCII-only\n"
        )

        self.assertEqual(result.returncode, 0, result.stderr)

    def test_rejects_non_ascii_subject(self) -> None:
        result = self.run_hook("feat(ai-provider): tts 隐私测试收口\n")

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("ASCII-only", result.stderr)

    def test_rejects_non_ascii_body(self) -> None:
        result = self.run_hook(
            "feat(work-journal): require english commit messages\n\n"
            "- 拒绝中文 body\n"
        )

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("ASCII-only", result.stderr)


if __name__ == "__main__":
    unittest.main()
