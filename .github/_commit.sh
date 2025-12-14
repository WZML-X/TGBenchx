#!/usr/bin/env sh
set -e

git config user.email "41898282+github-actions[bot]@users.noreply.github.com"
git config user.name "github-actions[bot]"

git add src/ out/

if git diff --cached --quiet; then
  echo "No changes to commit"
  exit 0
fi

git commit -m "Update ${LIBRARY} results"

git push origin HEAD:master
