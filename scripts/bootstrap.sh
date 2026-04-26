#!/usr/bin/env bash
# Repo-scaffold A1 onboarding self-check.
# Print current shell / OS / Go / Node / Python versions next to the values
# declared in .tool-versions. This script never installs or modifies anything;
# it exists so a new contributor can run it once and see what to bring up
# manually (or via asdf / mise) to match the locked toolchain.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TOOL_VERSIONS_FILE="${ROOT_DIR}/.tool-versions"

declared_version() {
  local tool="$1"
  if [ -f "${TOOL_VERSIONS_FILE}" ]; then
    awk -v t="${tool}" '$1 == t { print $2; exit }' "${TOOL_VERSIONS_FILE}"
  fi
}

current_version() {
  local cmd="$1"
  local args="$2"
  if command -v "${cmd}" >/dev/null 2>&1; then
    "${cmd}" ${args} 2>&1 | head -n 1
  else
    echo "(not installed)"
  fi
}

print_row() {
  local label="$1"
  local declared="$2"
  local current="$3"
  printf "  %-10s declared=%-12s current=%s\n" "${label}" "${declared:-<unset>}" "${current}"
}

echo "easyinterview onboarding self-check"
echo
echo "Environment:"
echo "  shell      ${SHELL:-<unset>}"
echo "  os         $(uname -s) $(uname -r) $(uname -m)"
echo "  pwd        ${ROOT_DIR}"
if [ ! -f "${TOOL_VERSIONS_FILE}" ]; then
  echo
  echo "WARNING: ${TOOL_VERSIONS_FILE} not found; toolchain comparison skipped."
  exit 0
fi
echo "  tool-versions ${TOOL_VERSIONS_FILE}"
echo
echo "Toolchain:"
print_row "golang"  "$(declared_version golang)"  "$(current_version go 'version')"
print_row "nodejs"  "$(declared_version nodejs)"  "$(current_version node '--version')"
print_row "pnpm"    "$(declared_version pnpm)"    "$(current_version pnpm '--version')"
print_row "python"  "$(declared_version python)"  "$(current_version python3 '--version')"
echo
echo "Next steps:"
echo "  - install asdf or mise if you want declared versions automatically"
echo "  - run 'make help' to list available top-level targets"
echo "  - run 'make install-hooks' to enable repo git hooks"
