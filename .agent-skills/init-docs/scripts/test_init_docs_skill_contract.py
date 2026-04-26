"""Contract checks for the /init-docs scaffold layout."""

from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[3]
SKILL_PATH = REPO_ROOT / ".agent-skills" / "init-docs" / "SKILL.md"
INIT_TEMPLATE_ROOT = REPO_ROOT / ".agent-skills" / "init-docs" / "templates"
DOC_ROOT = REPO_ROOT / "docs"
DOC_SUBDIRS = [
    "work-journal",
    "spec",
    "reports",
    "apis",
    "discuss",
    "bugs",
]
COLLAB_SENTENCE = "起草或修改正文时，必须参考同目录 `TEMPLATES.md`。"


def _read(path: Path) -> str:
    return path.read_text(encoding="utf-8")


def test_docs_subdirectories_have_separate_template_assets():
    for name in DOC_SUBDIRS:
        assert (DOC_ROOT / name / "TEMPLATES.md").exists(), name


def test_readmes_no_longer_inline_full_template_sections():
    banned_headings = {
        DOC_ROOT / "work-journal" / "README.md": "## 日志模板",
        DOC_ROOT / "spec" / "README.md": "## 4 文档模板",
        DOC_ROOT / "reports" / "README.md": "## 4 文档模板",
        DOC_ROOT / "apis" / "README.md": "## 5 文档模板",
        DOC_ROOT / "discuss" / "README.md": "## 4 文档模板",
        DOC_ROOT / "bugs" / "README.md": "## 6 模板",
    }

    for path, heading in banned_headings.items():
        assert heading not in _read(path), path.name


def test_readmes_explicitly_require_template_reference_for_collaboration():
    assert COLLAB_SENTENCE in _read(DOC_ROOT / "README.md")

    for name in DOC_SUBDIRS:
        assert COLLAB_SENTENCE in _read(DOC_ROOT / name / "README.md"), name
        assert COLLAB_SENTENCE in _read(INIT_TEMPLATE_ROOT / f"{name}-readme.md"), name


def test_init_docs_skill_declares_readme_templates_split():
    text = _read(SKILL_PATH)

    assert "README.md、`TEMPLATES.md` 和 INDEX.md" in text
    assert "README 只承载目录规范" in text
    assert "TEMPLATES.md" in text


def test_init_docs_template_resources_include_split_assets():
    expected = [
        "work-journal-templates.md",
        "spec-templates.md",
        "subspec-plans-readme.md",
        "subspec-plans-templates.md",
        "subspec-plans-index.md",
        "reports-templates.md",
        "apis-templates.md",
        "discuss-templates.md",
        "bugs-templates.md",
    ]

    for name in expected:
        assert (INIT_TEMPLATE_ROOT / name).exists(), name


def test_new_project_scaffold_omits_plan_projection_assets():
    text = _read(SKILL_PATH)

    assert not (DOC_ROOT / "plan").exists()
    assert (INIT_TEMPLATE_ROOT / "subspec-plans-readme.md").exists()
    assert (INIT_TEMPLATE_ROOT / "subspec-plans-templates.md").exists()
    assert (INIT_TEMPLATE_ROOT / "subspec-plans-index.md").exists()
    assert "top-level `docs/plan/`" in text


# --- Phase 7.1: test-framework scaffold ---

TEST_FRAMEWORK_TEMPLATES = [
    "scenarios-readme.md",
    "scenarios-shared-readme.md",
    "scenarios-e2e-readme.md",
    "scenarios-e2e-index.md",
]


def test_init_docs_declares_test_framework_option():
    """SKILL.md must declare test-framework as a scaffold option (C-15)."""
    text = _read(SKILL_PATH)
    assert "test-framework" in text, "SKILL.md must declare test-framework option"


def test_init_docs_test_framework_templates_exist():
    """All 4 test-framework template files must exist (C-15)."""
    for name in TEST_FRAMEWORK_TEMPLATES:
        assert (INIT_TEMPLATE_ROOT / name).exists(), \
            f"Template missing: {name}"


def test_plan_readmes_and_templates_default_to_in_place_revision():
    spec_readme = _read(DOC_ROOT / "spec" / "README.md")
    init_spec_readme = _read(INIT_TEMPLATE_ROOT / "spec-readme.md")

    expected_sentence = "同主题后续修订优先原地更新原 spec 和原 plan"

    assert expected_sentence in spec_readme
    assert expected_sentence in init_spec_readme


def test_plan_scaffold_prohibits_hard_coverage_threshold_gates():
    spec_template = _read(DOC_ROOT / "spec" / "TEMPLATES.md")
    init_spec_template = _read(INIT_TEMPLATE_ROOT / "spec-templates.md")

    banned_terms = ("coverage >= N%", "覆盖率 ≥ N%", "line coverage")

    assert all(term not in spec_template for term in banned_terms)
    assert all(term not in init_spec_template for term in banned_terms)


def test_bdd_checklist_templates_are_declared_in_spec_scaffolds():
    spec_template = _read(DOC_ROOT / "spec" / "TEMPLATES.md")
    init_spec_template = _read(INIT_TEMPLATE_ROOT / "spec-templates.md")
    plans_template = _read(INIT_TEMPLATE_ROOT / "subspec-plans-templates.md")

    for text in (spec_template, init_spec_template, plans_template):
        assert "bdd-plan.md" in text
        assert "bdd-checklist.md" in text
        assert "BDD-Gate" in text
    for text in (spec_template, init_spec_template):
        assert "主 `checklist.md` 只保留阶段级 `BDD-Gate`" in text
