import subprocess
import sys
from pathlib import Path

import ui_demo_pruning as audit


SCRIPT = Path(__file__).resolve().with_name("ui_demo_pruning.py")
REPO_ROOT = SCRIPT.parents[2]


def write(path: Path, body: str) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(body, encoding="utf-8")


def finding_paths(report: audit.AuditReport) -> list[str]:
    return [finding.path for finding in report.findings]


def test_accepts_document_owned_ui_design_without_demo(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "docs" / "ui-design" / "README.md",
        "UI architecture and user-flow design lives in docs/ui-design/.\n",
    )
    write(
        repo / "frontend" / "README.md",
        "Implement the interaction contract documented in docs/ui-design/.\n",
    )

    report = audit.scan_repo(repo)

    assert not report.demo_directory_exists
    assert report.findings == []


def test_rejects_ui_demo_directory_even_when_empty(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    (repo / "ui-design").mkdir(parents=True)

    report = audit.scan_repo(repo)

    assert report.demo_directory_exists
    assert report.failed


def test_rejects_positive_active_demo_contract_but_allows_negative_guard(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    write(
        repo / "frontend" / "README.md",
        "Run test:pixel-parity against the golden preview before delivery.\n",
    )
    write(
        repo / "AGENTS.md",
        "正式 frontend 不再依赖 ui-design/ Demo 或 pixel parity。\n"
        "当前合同未引入 UI truth source 或 golden preview。\n",
    )

    report = audit.scan_repo(repo)

    assert finding_paths(report) == ["frontend/README.md"]
    assert report.findings[0].label == "pixel parity contract"


def test_ignores_historical_evidence_paths(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    for path in (
        repo / "docs" / "work-journal" / "2026-05-01.md",
        repo / "docs" / "bugs" / "BUG-0001.md",
        repo / "docs" / "reports" / "prototype-assessment.md",
        repo / "docs" / "spec" / "frontend-shell" / "history.md",
    ):
        write(path, "Historical ui-design/ golden preview and pixel parity evidence.\n")

    report = audit.scan_repo(repo)

    assert report.findings == []


def test_cli_reports_demo_and_active_reference_failures(tmp_path: Path) -> None:
    repo = tmp_path / "repo"
    (repo / "ui-design").mkdir(parents=True)
    write(repo / "Makefile", "test:\n\t@node ui-design/demo.test.mjs\n")

    result = subprocess.run(
        [sys.executable, str(SCRIPT), "--repo-root", str(repo)],
        check=False,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )

    assert result.returncode == 1
    assert "ui_demo_directory: present" in result.stderr
    assert "active_residuals (1)" in result.stderr
    assert "Makefile:2" in result.stderr


def test_makefile_exposes_ui_demo_pruning_gate() -> None:
    makefile = (REPO_ROOT / "Makefile").read_text(encoding="utf-8")

    assert "lint-ui-demo-pruning:" in makefile
    assert "ui_demo_pruning.py" in makefile
    assert "lint: " in makefile and "lint-ui-demo-pruning" in makefile
