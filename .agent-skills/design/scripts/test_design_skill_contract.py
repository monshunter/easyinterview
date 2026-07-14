"""Contract checks for the /design skill and its canonical BDD docs."""

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
SKILL_PATH = REPO_ROOT / ".agent-skills" / "design" / "SKILL.md"
SPEC_TEMPLATE_PATH = REPO_ROOT / "docs" / "spec" / "TEMPLATES.md"
INIT_SPEC_TEMPLATE_PATH = REPO_ROOT / ".agent-skills" / "init-docs" / "templates" / "spec-templates.md"


def _read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def test_design_skill_separates_domain_behavior_ids_from_real_e2e_ids():
    text = _read(SKILL_PATH)

    assert "stable domain Behavior ID such as `BDD.AUTH.001`" in text
    assert "may be verified by a code-level domain behavior test" in text
    assert "Allocate an ID such as `E2E.P0.001` only when" in text
    assert "drives the running product through real HTTP/UI" in text
    assert "domain Behavior IDs do not create scenario directories" in text


def test_design_skill_prohibits_ac_style_bdd_gate_ids_for_new_docs():
    text = _read(SKILL_PATH)

    assert "Generating BDD-Gate checklist items with `AC-*` references for new documents" in text


def test_design_skill_prohibits_hard_coverage_gates_in_test_plans():
    text = _read(SKILL_PATH)

    assert "Do not generate hard coverage-percentage gates in acceptance criteria or checklist items." in text
    assert "observational background rather than a completion, commit, or phase-exit condition" in text


def test_plan_docs_allow_behavior_or_real_e2e_ids_for_bdd_gate_examples():
    plan_template = _read(SPEC_TEMPLATE_PATH)
    init_template = _read(INIT_SPEC_TEMPLATE_PATH)

    assert "BDD-Gate: 验证 `${Behavior ID 或真实 E2E ID}` 通过" in plan_template
    assert "BDD-Gate: 验证 `${Behavior ID 或真实 E2E ID}` 通过" in init_template
    assert "BDD-Gate: 验证 AC-1, AC-2 通过" not in plan_template
    assert "BDD-Gate: 验证 AC-1, AC-2 通过" not in init_template


def test_design_skill_generates_bdd_plan_and_checklist_together():
    text = _read(SKILL_PATH)
    spec_template = _read(SPEC_TEMPLATE_PATH)

    assert "generate `bdd-plan.md` and `bdd-checklist.md`" in text
    assert "add `bddPlan` and `bddChecklist` to `context.yaml`" in text
    assert "bdd-checklist.md" in spec_template


def test_design_skill_requires_tdd_for_code_and_bdd_for_user_behavior():
    text = _read(SKILL_PATH)

    assert "Code plan requires TDD" in text
    assert "Feature plan requires BDD" in text
    assert "user-visible UI, API behavior, or business workflow" in text
    assert "BDD describes user-observable behavior" in text
    assert "Pure configuration defaults, internal contracts, tooling" in text
    assert "generate no BDD files" in text


def test_design_skill_never_wraps_code_tests_as_e2e():
    text = _read(SKILL_PATH)

    assert "do not create a `test/scenarios/e2e/` shell wrapper" in text
    assert "Go, Vitest, npm test, pytest, lint, source-contract, fixture, or build commands" in text
    assert "without mock transport or request interception replacing the backend" in text


def test_design_skill_requires_explicit_coverage_matrix_for_plan_and_tests():
    text = _read(SKILL_PATH)

    assert "Step 3.5: Build the Coverage Matrix" in text
    assert "Primary path" in text
    assert "Failure / recovery path" in text
    assert "Boundary condition" in text
    assert "Cross-layer contract" in text
    assert "Privacy / security / observability" in text
    assert "Regression / non-current-negative" in text
    assert "Every non-docs checklist item must name its verification source" in text
    assert "Test plans must include a coverage matrix" in text
    assert "BDD behavior selection must cover the primary user journey plus the highest-risk alternate or failure/recovery journey" in text


def test_design_skill_requires_branch_guard_before_doc_mutation():
    text = _read(SKILL_PATH)

    assert "Step 2.5: Branch Guard Before Document Mutation" in text
    assert "before invoking `/create-doc`" in text
    assert "fast-forward-only semantics" in text
    assert "Do not generate documents from a stale parent branch." in text
    assert "Never invoke `/create-doc`, create spec / plan directories, revise completed owner docs" in text
    assert "the Step 2.5 branch guard succeeds" in text


def test_design_skill_rejects_tool_name_branch_prefixes():
    text = _read(SKILL_PATH)

    assert "`design/{subject}`" in text
    assert "`spec-design/`" in text
    assert "never create new" in text
    assert "`codex/`, `claude/`, `gemini/`, `agent/`" in text
    assert "rename it to the semantic repository prefix before editing files" in text


def test_spec_templates_include_quality_gate_classification():
    plan_template = _read(SPEC_TEMPLATE_PATH)
    init_template = _read(INIT_SPEC_TEMPLATE_PATH)

    for text in (plan_template, init_template):
        assert "## 3 质量门禁分类" in text
        assert "Plan 类型" in text
        assert "TDD 策略" in text
        assert "BDD 策略" in text
        assert "替代验证 gate" in text


def test_spec_template_marks_acceptance_ids_as_descriptive_only():
    text = _read(SPEC_TEMPLATE_PATH)

    assert "`ID` 列是文档内的说明性编号" in text
    assert "它不是 BDD 场景编号" in text


def test_spec_templates_include_user_decision_section():
    spec_template = _read(SPEC_TEMPLATE_PATH)
    init_template = _read(INIT_SPEC_TEMPLATE_PATH)

    assert "## 3 用户决策 / 待确认事项" in spec_template
    assert "## 3 用户决策 / 待确认事项" in init_template
    assert "无待确认事项时可省略" in spec_template
    assert "无待确认事项时可省略" in init_template


def test_plan_templates_prohibit_hard_coverage_threshold_items():
    plan_template = _read(SPEC_TEMPLATE_PATH)
    init_template = _read(INIT_SPEC_TEMPLATE_PATH)

    assert "coverage >= N%" not in plan_template
    assert "覆盖率 ≥ N%" not in init_template
    assert "本计划列出的实现 / 测试项全部通过" in plan_template
    assert "本计划列出的实现 / 测试项全部通过" in init_template
