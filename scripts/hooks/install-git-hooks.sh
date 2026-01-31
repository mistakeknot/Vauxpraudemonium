#!/usr/bin/env bash
set -euo pipefail

repo_root=$(git rev-parse --show-toplevel)
cd "$repo_root"

if [ ! -d .githooks ]; then
  echo "Missing .githooks directory; nothing to install."
  exit 1
fi

if [ -f .githooks/pre-commit ]; then
  chmod +x .githooks/pre-commit
fi

git config core.hooksPath .githooks

echo "Installed git hooks (core.hooksPath=.githooks)."
