#!/usr/bin/env bash
# gitleaks second-layer secret scan (secrets-and-config spec D-6 / §6 C-8).
# Calls a locally installed `gitleaks` if available; otherwise prints an
# install hint and exits 0 so the lint chain does not block developers
# without the tool. The pre-commit-secrets.sh hook is the primary, mandatory
# gate — gitleaks supplements it with broader rules.
#
# Usage: scripts/lint/gitleaks.sh [repo-root]

set -euo pipefail

REPO_ROOT="${1:-$(pwd)}"

if ! command -v gitleaks >/dev/null 2>&1; then
  cat <<'EOF' >&2
gitleaks: not installed locally — skipping second-layer scan.
Recommended (per secrets-and-config spec §4.2): `brew install gitleaks` or
download from https://github.com/gitleaks/gitleaks/releases . The primary
pre-commit hook (scripts/git-hooks/pre-commit-secrets.sh) already runs.
Remote CI secret scan is wired by A5 ci-pipeline-baseline; A4 only runs
locally.
EOF
  exit 0
fi

cd "$REPO_ROOT"
# --no-banner keeps logs short for CI / Make output;
# --redact ensures matched secrets are masked in the report.
#
# Scan the git candidate set rather than the full working tree. Local runtime
# files such as deploy/dev-stack/.env are intentionally ignored by .gitignore
# and may contain real developer secrets; scanning them makes the lint target
# non-repeatable on configured dev machines.
SCAN_ROOT="$(mktemp -d)"
cleanup() {
  rm -rf "$SCAN_ROOT"
}
trap cleanup EXIT

while IFS= read -r -d '' path; do
  if [ ! -f "$path" ]; then
    continue
  fi
  mkdir -p "$SCAN_ROOT/$(dirname "$path")"
  cp -p "$path" "$SCAN_ROOT/$path"
done < <(git ls-files -z --cached --others --exclude-standard)

exec gitleaks detect --no-banner --no-git --redact --source "$SCAN_ROOT"
