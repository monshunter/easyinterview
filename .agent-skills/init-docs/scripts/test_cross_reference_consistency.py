"""Cross-skill reference consistency tests (D-14, C-14, C-17)."""

from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent.parent.parent
SKILLS_ROOT = REPO_ROOT / ".agent-skills"

NON_CURRENT_SKILL_NAMES = [
    "/test-investigate",
    "/test-env",
    "/test-scenario",
    "/new-scenario",
]


# --- Phase 6.1: scenario framework de-branding ---


def test_scenario_readme_no_brand_in_title():
    """test/scenarios/README.md title must not contain InterviewPilot (C-14)."""
    readme = (REPO_ROOT / "test" / "scenarios" / "README.md").read_text(encoding="utf-8")
    first_line = readme.split("\n")[0]
    assert "InterviewPilot" not in first_line, \
        f"Scenario README title still branded: {first_line}"


def test_scenario_readme_no_brand_in_description():
    """test/scenarios/README.md §1 must not contain InterviewPilot (C-14)."""
    readme = (REPO_ROOT / "test" / "scenarios" / "README.md").read_text(encoding="utf-8")
    # Extract section 1 (between ## 1 and ## 2)
    start = readme.find("## 1")
    end = readme.find("## 2", start)
    section1 = readme[start:end] if start != -1 and end != -1 else ""
    assert "InterviewPilot" not in section1, \
        "Scenario README §1 still contains InterviewPilot"


def test_e2e_readme_no_brand_in_description():
    """test/scenarios/e2e/README.md §1 must not contain InterviewPilot (C-14)."""
    readme = (REPO_ROOT / "test" / "scenarios" / "e2e" / "README.md").read_text(encoding="utf-8")
    start = readme.find("## 1")
    end = readme.find("## 2", start)
    section1 = readme[start:end] if start != -1 and end != -1 else ""
    assert "InterviewPilot" not in section1, \
        "E2E README §1 still contains InterviewPilot"


# --- Phase 7.2: manual bootstrap commands ---


def test_e2e_readme_has_manual_bootstrap_section():
    """e2e/README.md §2 must contain manual bootstrap subsection (C-16)."""
    readme = (REPO_ROOT / "test" / "scenarios" / "e2e" / "README.md").read_text(encoding="utf-8")
    start = readme.find("## 2")
    end = readme.find("## 3", start) if readme.find("## 3", start) != -1 else len(readme)
    section2 = readme[start:end]
    assert "手动引导" in section2, \
        "e2e/README.md §2 must contain a 手动引导 subsection"


def test_scenario_env_has_no_script_fallback():
    """scenario-env/SKILL.md setup Step 3 must have a no-script fallback (C-16)."""
    text = (REPO_ROOT / ".agent-skills" / "scenario-env" / "SKILL.md").read_text(encoding="utf-8")
    assert "手动引导" in text or "manual bootstrap" in text.lower(), \
        "scenario-env SKILL.md must reference manual bootstrap fallback"


# --- Phase 8.3: non-current skill name scan ---


def test_all_skill_files_no_non_current_names():
    """No SKILL.md may reference non-current skill names (C-17, D-14)."""
    violations = []
    for skill_file in SKILLS_ROOT.rglob("SKILL.md"):
        text = skill_file.read_text(encoding="utf-8")
        for name in NON_CURRENT_SKILL_NAMES:
            if name in text:
                rel = skill_file.relative_to(REPO_ROOT)
                violations.append(f"{rel} references non-current '{name}'")
    assert not violations, "Non-current skill references found:\n" + "\n".join(violations)
