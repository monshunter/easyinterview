---
name: scenario-redeploy
description: Rebuild and redeploy the repo components required by the local scenario or local frontend/backend integration environment. Read `test/scenarios/README.md` and the active suite README before commands. Prefer `test/scenarios/env-redeploy.sh [deps|backend|frontend|all]` and root `make scenario-env-redeploy TARGET=<target>` when present. Triggers on /scenario-redeploy.
---

# Scenario Redeploy Skill

Rebuild and redeploy the repo components required by the local scenario
environment used by scenario tests and local frontend/backend integration.

## Usage

```text
/scenario-redeploy
/scenario-redeploy deps
/scenario-redeploy backend
/scenario-redeploy frontend
/scenario-redeploy all
```

## Workflow

1. Read `test/scenarios/README.md` and the active suite README.
2. Resolve the documented environment identity, repo-tracked redeploy entrypoint,
   and component mapping. In this repo the supported top-level targets are
   `deps|backend|frontend|all`.
3. Prefer `test/scenarios/env-redeploy.sh [deps|backend|frontend|all]` when the
   framework or suite README documents it.
4. If the user explicitly wants Makefile integration, use
   `make scenario-env-redeploy TARGET=<target>`.
5. If no top-level redeploy script exists, fall back to the exact
   suite-documented repo-tracked script or Make target.
6. If the repo does not define a redeploy contract, stop and report the missing
   deployment contract instead of inventing commands.
7. Rebuild the requested components and apply the documented redeploy flow.
8. Run the documented smoke checks or `/scenario-env verify`.

## Current Host-Run Boundary

The current local topology is host-run: Docker Compose provides external
dependencies, backend/frontend processes are not deployed through a Kind, Helm,
or cluster rollout, and `test/scenarios/env-redeploy.sh backend|frontend|all`
must rebuild and restart the matching host-run backend/frontend process from
`deploy/dev-stack/.env`. After redeploy, report frontend/backend/Mailpit
addresses plus `.test-output/local-dev/{backend,frontend}.log` and PID files so
the developer can take over debugging. Keep secrets in ignored local files and
never print secret values.

An explicit full-container request is the documented exception: run
`make dev-container-up`, then `make dev-container-doctor`; the default frontend
and backend ports are `10800` and `10801`. Use `make dev-container-logs` for
diagnostics and `make dev-container-down` for a volume-preserving stop. Do not
route this explicit topology through the host-run `scenario-env-redeploy` path.

## Rules

- Do not assume cluster types, Helm charts, namespaces, or component
  names.
- When only one component changed, prefer the narrowest redeploy that the repo
  documents.
- Treat `test/scenarios/env-redeploy.sh` as a framework-owned shim. Concrete
  component names, aliases, and rebuild steps must come from the active suite
  README and repo scripts.
