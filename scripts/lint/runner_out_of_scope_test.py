from pathlib import Path

import runner_out_of_scope


def test_runner_out_of_scope_flags_out_of_scope_review_runner(tmp_path: Path) -> None:
    path = tmp_path / "main.go"
    path.write_text("runner := domainreview.NewRunner(opts)\n", encoding="utf-8")

    problems = runner_out_of_scope.scan_paths([path])

    assert len(problems) == 1
    assert "review.NewRunner" in problems[0]


def test_runner_out_of_scope_flags_background_mail_dispatcher(tmp_path: Path) -> None:
    path = tmp_path / "wire.go"
    path.write_text("d := auth.NewBackgroundMailDispatcher(opts)\n", encoding="utf-8")

    problems = runner_out_of_scope.scan_paths([path])

    assert any("NewBackgroundMailDispatcher" in p for p in problems)
    assert any("BackgroundMailDispatcher" in p for p in problems)


def test_runner_out_of_scope_flags_targetjob_drainer_instantiation(tmp_path: Path) -> None:
    path = tmp_path / "wire.go"
    path.write_text("drainer := targetjob.NewDrainer(targetjob.DrainerOptions{})\n", encoding="utf-8")

    problems = runner_out_of_scope.scan_paths([path])

    assert any("targetjob.NewDrainer" in p for p in problems)


def test_runner_out_of_scope_flags_removed_targetjob_contract(tmp_path: Path) -> None:
    targetjob = tmp_path / "backend" / "internal" / "targetjob"
    kernel = tmp_path / "backend" / "internal" / "runner"
    scenarios = tmp_path / "backend" / "cmd" / "api"
    targetjob.mkdir(parents=True)
    kernel.mkdir(parents=True)
    scenarios.mkdir(parents=True)
    (targetjob / "drainer.go").write_text("package targetjob\ntype ClaimedJob struct{}\n", encoding="utf-8")
    (kernel / "adapter_targetjob.go").write_text("package runner\nfunc FromTargetjobHandler() {}\n", encoding="utf-8")
    (scenarios / "resume_parse_drainer_scenario_test.go").write_text("package main\n", encoding="utf-8")

    problems = runner_out_of_scope.scan_removed_targetjob_contract(tmp_path)

    assert any("targetjob/drainer.go" in p for p in problems)
    assert any("runner/adapter_targetjob.go" in p for p in problems)
    assert any("drainer_scenario_test.go" in p for p in problems)
    assert any("duplicate targetjob async job type" in p for p in problems)


def test_runner_out_of_scope_accepts_native_kernel_contract(tmp_path: Path) -> None:
    handler = tmp_path / "backend" / "internal" / "targetjob" / "parse_executor.go"
    scenario = tmp_path / "backend" / "cmd" / "api" / "resume_parse_runner_scenario_test.go"
    handler.parent.mkdir(parents=True)
    scenario.parent.mkdir(parents=True)
    handler.write_text(
        "package targetjob\nfunc (p *ParseExecutor) Handle(ctx context.Context, job runner.ClaimedJob) runner.JobOutcome { return runner.JobOutcome{} }\n",
        encoding="utf-8",
    )
    scenario.write_text("package main\n", encoding="utf-8")

    assert runner_out_of_scope.scan_removed_targetjob_contract(tmp_path) == []


def test_runner_out_of_scope_flags_per_domain_drainer_fields(tmp_path: Path) -> None:
    path = tmp_path / "wire.go"
    path.write_text("_ = jdmatchRuntime.Drainer\n_ = reportRuntime.Runner\n", encoding="utf-8")

    problems = runner_out_of_scope.scan_paths([path])

    assert len(problems) == 2
    assert any("per-domain Drainer field" in p for p in problems)
    assert any("reportRuntime Runner/Reaper field" in p for p in problems)


def test_runner_out_of_scope_ignores_clean_production(tmp_path: Path) -> None:
    path = tmp_path / "clean.go"
    path.write_text(
        "kernel := runner.New(opts)\nkernel.Register(jobType, h)\n",
        encoding="utf-8",
    )

    assert runner_out_of_scope.scan_paths([path]) == []


def test_runner_out_of_scope_excludes_test_files() -> None:
    assert runner_out_of_scope.is_production_go(Path("backend/internal/runner/runtime.go"))
    assert not runner_out_of_scope.is_production_go(Path("backend/internal/runner/runtime_test.go"))
