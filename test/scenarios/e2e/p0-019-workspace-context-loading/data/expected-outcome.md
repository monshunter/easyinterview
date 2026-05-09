# Expected Outcome
- A: workspace-cta-start/workspace-binding-*/workspace-companyintel-* 全部命中
- B: workspace-empty-{eyebrow,title,desc,cta} 命中，CTA → home
- C: workspace-missing-resume-{eyebrow,title,desc,cta} 命中，CTA → resume_versions?flow=create
- D: archived → ready=false; not-found → error state
- Vitest Tests all passed
