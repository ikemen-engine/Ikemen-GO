#!/usr/bin/env bash
# Apply the Lua whitespace rules to every tracked .lua file, in place.
# Used by both the PR check (lint-lua-check.yml) and the post-push fixer
# (lint-code-style.yml), so the check and the fix can never disagree.
# Requires CodeFormat (EmmyLuaCodeStyle) on PATH; rules live in .editorconfig.
set -euo pipefail

git ls-files -z '*.lua' | while IFS= read -r -d '' f; do
	# Two passes: splitting `a; b` leaves the new line unindented until the
	# second pass. A third pass is a no-op.
	CodeFormat format -f "$f" -c .editorconfig -ow
	CodeFormat format -f "$f" -c .editorconfig -ow
done

# CodeFormat does not trim trailing whitespace at the end of comments.
git ls-files -z '*.lua' | xargs -0 -r sed -i 's/[[:blank:]]\+$//'
