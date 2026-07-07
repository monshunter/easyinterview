from pathlib import Path

import runner_non_current


def test_runner_non_current_flags_non_current_review_runner(tmp_path: Path) -> None:
    path = tmp_path / "main.go"
    path.write_text("runner := domainreview.NewRunner(opts)\n", encoding="utf-8")

    problems = runner_non_current.scan_paths([path])

    assert len(problems) == 1
    assert "review.NewRunner" in problems[0]


def test_runner_non_current_flags_background_mail_dispatcher(tmp_path: Path) -> None:
    path = tmp_path / "wire.go"
    path.write_text("d := auth.NewBackgroundMailDispatcher(opts)\n", encoding="utf-8")

    problems = runner_non_current.scan_paths([path])

    assert any("NewBackgroundMailDispatcher" in p for p in problems)
    assert any("BackgroundMailDispatcher" in p for p in problems)


def test_runner_non_current_flags_targetjob_drainer_instantiation(tmp_path: Path) -> None:
    path = tmp_path / "wire.go"
    path.write_text("drainer := targetjob.NewDrainer(targetjob.DrainerOptions{})\n", encoding="utf-8")

    problems = runner_non_current.scan_paths([path])

    assert any("targetjob.NewDrainer" in p for p in problems)


def test_runner_non_current_flags_per_domain_drainer_fields(tmp_path: Path) -> None:
    path = tmp_path / "wire.go"
    path.write_text("_ = jdmatchRuntime.Drainer\n_ = reportRuntime.Runner\n", encoding="utf-8")

    problems = runner_non_current.scan_paths([path])

    assert len(problems) == 2
    assert any("per-domain Drainer field" in p for p in problems)
    assert any("reportRuntime Runner/Reaper field" in p for p in problems)


def test_runner_non_current_ignores_clean_production(tmp_path: Path) -> None:
    path = tmp_path / "clean.go"
    path.write_text(
        "kernel := runner.New(opts)\nkernel.Register(jobType, runner.FromTargetjobHandler(h))\n",
        encoding="utf-8",
    )

    assert runner_non_current.scan_paths([path]) == []


def test_runner_non_current_excludes_test_files() -> None:
    assert runner_non_current.is_production_go(Path("backend/internal/runner/runtime.go"))
    assert not runner_non_current.is_production_go(Path("backend/internal/runner/runtime_test.go"))
