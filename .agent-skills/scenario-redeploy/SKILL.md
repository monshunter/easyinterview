---
name: scenario-redeploy
description: Rebuild and redeploy the repo components required by the local scenario environment. Read `test/scenarios/README.md` and the active suite README before commands. Prefer repo-tracked build and deploy scripts when present. Triggers on /scenario-redeploy.
---

# Scenario Redeploy Skill

Rebuild and redeploy the repo components required by the local scenario
environment used by scenario tests.

## Usage

```text
/scenario-redeploy
/scenario-redeploy controller
/scenario-redeploy cert-manager
```

## Workflow

1. Read `test/scenarios/README.md` and the active suite README.
2. Resolve the documented environment identity, repo-tracked redeploy entrypoint, and suite-local component mapping.
3. Prefer `test/scenarios/env-redeploy.sh [component]` when the framework or suite README documents it.
4. If no top-level redeploy script exists, fall back to the exact suite-documented repo-tracked script or Make target.
5. If the repo does not define a redeploy contract, stop and report the missing deployment contract instead of inventing commands.
6. Rebuild the requested components and apply the documented redeploy flow.
7. Run the documented smoke checks or `/scenario-env verify`.

## Rules

- Do not assume historical cluster types, Helm charts, namespaces, or component names.
- When only one component changed, prefer the narrowest redeploy that the repo documents.
- Treat `test/scenarios/env-redeploy.sh` as a framework-owned shim only. Concrete component names, aliases, and rebuild steps must come from the active suite `suite_env.sh` / README, not from the skill body.
