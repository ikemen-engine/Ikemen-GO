#!/usr/bin/env python3
"""Generate categorized release notes for a tag range.

GitHub's built-in release-note generator maps each commit in the range to the pull
request that contains that exact SHA. That fails on release branches, because every
backported commit is a cherry-pick with a new SHA that was never part of any PR, so
the generated changelog comes out empty.

This script resolves each commit to its pull request itself, following the
"(cherry picked from commit <sha>)" trailer that `git cherry-pick -x` records, so a
backported commit is credited to the original PR on develop. Entries are grouped
using the categories and exclusions declared in .github/release.yml, and each PR is
listed once no matter how many of its commits were backported.
"""

import argparse
import json
import os
import re
import subprocess
import sys
import urllib.error
import urllib.request

GRAPHQL_URL = "https://api.github.com/graphql"
# GitHub allows many aliased nodes per query; 100 keeps each request well inside
# the response-size and point budget.
BATCH_SIZE = 100
CHERRY_PICK_RE = re.compile(r"cherry picked from commit ([0-9a-f]{7,40})", re.I)


def log(message):
    print(message, file=sys.stderr)


def warn(message):
    # Surfaces in the Actions run summary as an annotation.
    print("::warning::" + message)


def run_git(args):
    result = subprocess.run(
        ["git"] + args, capture_output=True, text=True, check=True
    )
    return result.stdout


def read_config(path):
    """Read exclusion labels and categories from .github/release.yml.

    Parsed with a minimal reader rather than PyYAML to avoid adding a dependency to
    the release job. Only the small, fixed shape of that file is supported.
    """
    exclude = []
    categories = []
    if not os.path.exists(path):
        return exclude, categories

    section = None
    current = None
    for raw in open(path, encoding="utf-8"):
        line = raw.rstrip("\n")
        if not line.strip() or line.lstrip().startswith("#"):
            continue
        indent = len(line) - len(line.lstrip())
        text = line.strip()

        if indent <= 2 and text.rstrip(":") == "exclude":
            section, current = "exclude", None
            continue
        if indent <= 2 and text.rstrip(":") == "categories":
            section, current = "categories", None
            continue

        if section == "exclude":
            if text.startswith("- "):
                exclude.append(text[2:].strip().strip("'\""))
            continue

        if section == "categories":
            if text.startswith("- title:"):
                current = {"title": text.split(":", 1)[1].strip().strip("'\""),
                           "labels": []}
                categories.append(current)
            elif text.startswith("- ") and current is not None:
                current["labels"].append(text[2:].strip().strip("'\""))
    return exclude, categories


def commits_in_range(previous_tag, head):
    """Return [(sha, subject, lookup_sha, is_backport)] oldest-last.

    lookup_sha is the original develop commit for a cherry-pick, otherwise the
    commit's own SHA.
    """
    if previous_tag:
        rev_range = "%s..%s" % (previous_tag, head)
    else:
        rev_range = head

    raw = run_git(["log", "--format=%H%x1f%s%x1f%b%x1e", rev_range])
    commits = []
    for record in raw.split("\x1e"):
        record = record.strip("\n")
        if not record.strip():
            continue
        sha, subject, body = record.split("\x1f")
        match = CHERRY_PICK_RE.search(body)
        if match:
            commits.append((sha, subject, match.group(1), True))
        else:
            commits.append((sha, subject, sha, False))
    return commits


def graphql(token, query, variables=None):
    payload = json.dumps({"query": query, "variables": variables or {}}).encode()
    request = urllib.request.Request(
        GRAPHQL_URL,
        data=payload,
        headers={
            "Authorization": "Bearer " + token,
            "Accept": "application/vnd.github+json",
            "Content-Type": "application/json",
            "User-Agent": "ikemen-release-notes",
        },
    )
    try:
        with urllib.request.urlopen(request) as response:
            body = json.load(response)
    except urllib.error.HTTPError as error:
        log("GraphQL HTTP %s: %s" % (error.code, error.read()[:500]))
        raise
    if "errors" in body:
        raise RuntimeError("GraphQL errors: %s" % body["errors"][:3])
    return body["data"]


def resolve_pull_requests(token, owner, name, shas):
    """Map commit SHA -> PR metadata, batching lookups into few requests."""
    resolved = {}
    unique = list(dict.fromkeys(shas))

    for start in range(0, len(unique), BATCH_SIZE):
        batch = unique[start:start + BATCH_SIZE]
        fields = []
        for index, sha in enumerate(batch):
            fields.append(
                'c%d: object(oid: "%s") { ... on Commit { '
                "associatedPullRequests(first: 1) { nodes { "
                "number title url labels(first: 20) { nodes { name } } "
                "author { login } } } } }" % (index, sha)
            )
        query = "query { repository(owner: \"%s\", name: \"%s\") { %s } }" % (
            owner, name, " ".join(fields)
        )
        data = graphql(token, query)["repository"]

        for index, sha in enumerate(batch):
            node = data.get("c%d" % index)
            if not node:
                continue
            nodes = node.get("associatedPullRequests", {}).get("nodes") or []
            if not nodes:
                continue
            pull = nodes[0]
            author = (pull.get("author") or {}).get("login") or "unknown"
            resolved[sha] = {
                "number": pull["number"],
                "title": pull["title"],
                "url": pull["url"],
                "labels": [l["name"] for l in
                           (pull.get("labels", {}).get("nodes") or [])],
                "author": author,
            }
        log("resolved %d/%d commits" % (min(start + BATCH_SIZE, len(unique)),
                                        len(unique)))
    return resolved


def categorize(pulls, exclude, categories):
    """Bucket PRs into categories, honouring exclusions. Returns (buckets, skipped)."""
    buckets = {category["title"]: [] for category in categories}
    catchall = None
    for category in categories:
        if "*" in category["labels"]:
            catchall = category["title"]

    skipped = []
    for pull in pulls:
        labels = set(pull["labels"])
        if labels & set(exclude):
            continue

        placed = False
        for category in categories:
            if labels & set(category["labels"]):
                buckets[category["title"]].append(pull)
                placed = True
                break
        if not placed:
            if catchall:
                buckets[catchall].append(pull)
            else:
                skipped.append(pull)
    return buckets, skipped


def render(buckets, categories, repo, previous_tag, tag):
    lines = []
    for category in categories:
        entries = buckets.get(category["title"]) or []
        if not entries:
            continue
        lines.append("## " + category["title"])
        for pull in sorted(entries, key=lambda p: p["number"]):
            lines.append("* %s by @%s in %s" % (
                pull["title"], pull["author"], pull["url"]))
        lines.append("")

    if previous_tag and tag:
        lines.append("**Full Changelog**: https://github.com/%s/compare/%s...%s"
                     % (repo, previous_tag, tag))
        lines.append("")
    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo", required=True, help="owner/name")
    parser.add_argument("--previous-tag", default="")
    parser.add_argument("--tag", default="")
    parser.add_argument("--head", default="HEAD")
    parser.add_argument("--config", default=".github/release.yml")
    parser.add_argument("--output", required=True)
    parser.add_argument("--prefix", default="",
                        help="optional markdown file prepended to the notes")
    args = parser.parse_args()

    token = os.environ.get("GITHUB_TOKEN", "")
    if not token:
        log("GITHUB_TOKEN is required")
        return 1

    owner, name = args.repo.split("/", 1)
    exclude, categories = read_config(args.config)
    if not categories:
        log("no categories found in %s" % args.config)
        return 1

    commits = commits_in_range(args.previous_tag, args.head)
    log("%d commits in %s..%s" % (len(commits), args.previous_tag or "root",
                                  args.head))
    if not commits:
        warn("no commits found in range; release notes will be empty")

    backports = sum(1 for c in commits if c[3])
    log("%d of them are cherry-picks with a resolvable original commit" % backports)

    resolved = resolve_pull_requests(token, owner, name,
                                     [c[2] for c in commits])

    # Deduplicate by PR number: a PR contributing several commits is listed once.
    pulls = {}
    unresolved = []
    for sha, subject, lookup, is_backport in commits:
        pull = resolved.get(lookup)
        if not pull:
            unresolved.append((sha, subject, is_backport))
            continue
        pulls.setdefault(pull["number"], pull)

    log("%d unique pull requests" % len(pulls))

    buckets, skipped = categorize(list(pulls.values()), exclude, categories)

    # A cherry-pick whose original could not be traced is only worth a warning when
    # the commit would otherwise have appeared in the notes. Direct pushes to
    # develop (release chores, CI tweaks) have no PR by nature, and their
    # conventional-commit type is always one the config excludes, so they are noise.
    excluded_types = tuple(
        label.split(":", 1)[1].strip() + ":"
        for label in exclude if ":" in label
    )
    lost = [
        entry for entry in unresolved
        if entry[2] and not entry[1].startswith(excluded_types)
    ]
    for sha, subject, _ in lost:
        warn("backported commit %s could not be traced to a pull request: %s"
             % (sha[:8], subject))
    if unresolved:
        log("%d commits had no associated pull request (%d of them noteworthy)"
            % (len(unresolved), len(lost)))

    for pull in skipped:
        warn("PR #%d has no changelog category (labels: %s) and was omitted: %s"
             % (pull["number"], ", ".join(pull["labels"]) or "none",
                pull["title"]))

    body = render(buckets, categories, args.repo, args.previous_tag, args.tag)

    prefix = ""
    if args.prefix and os.path.exists(args.prefix):
        prefix = open(args.prefix, encoding="utf-8").read()
        if not prefix.endswith("\n"):
            prefix += "\n"

    with open(args.output, "w", encoding="utf-8") as handle:
        handle.write(prefix + body)

    log("wrote %s (%d bytes)" % (args.output, len(prefix + body)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
