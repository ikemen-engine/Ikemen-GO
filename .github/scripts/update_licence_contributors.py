#!/usr/bin/env python3
"""Generate the LICENCE.txt copyright block from develop commit statistics."""

from __future__ import annotations

import json
import os
import re
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from collections import defaultdict
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, TypedDict


BASE_BRANCH = "develop"
MIN_COMMITS = 2
MIN_ADDITIONS = 100
MANUAL_CONTRIBUTORS = {"Suehiro": 2016}
EXCLUDED_LOGINS = {
    "lint-action",
    "ppitulaj",  # K4thos' second account.
}

API_VERSION = "2022-11-28"
STATS_RETRIES = 12
STATS_RETRY_SECONDS = 5


class Contributor(TypedDict):
    login: str
    commits: int
    additions: int
    first_year: int


def request_json(
    token: str,
    url: str,
    *,
    accepted_statuses: set[int] = {200},
) -> tuple[int, Any]:
    request = urllib.request.Request(
        url,
        headers={
            "Accept": "application/vnd.github+json",
            "Authorization": f"Bearer {token}",
            "User-Agent": "ikemen-go-licence-generator",
            "X-GitHub-Api-Version": API_VERSION,
        },
    )

    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            status = response.status
            if status not in accepted_statuses:
                raise RuntimeError(f"GitHub API returned unexpected HTTP {status}")
            if status in {202, 204}:
                return status, None
            return status, json.load(response)
    except urllib.error.HTTPError as exc:
        details = exc.read().decode("utf-8", errors="replace")
        raise RuntimeError(f"GitHub API returned HTTP {exc.code}: {details}") from exc
    except urllib.error.URLError as exc:
        raise RuntimeError(f"GitHub API request failed: {exc}") from exc


def fetch_contributor_stats(token: str, repository: str) -> list[dict[str, Any]]:
    url = f"https://api.github.com/repos/{repository}/stats/contributors"

    for attempt in range(STATS_RETRIES):
        status, data = request_json(token, url, accepted_statuses={200, 202, 204})
        if status == 200:
            return data
        if status == 204:
            return []

        if attempt + 1 < STATS_RETRIES:
            print("Contributor statistics are being generated; retrying...")
            time.sleep(STATS_RETRY_SECONDS)

    raise RuntimeError("GitHub did not finish generating contributor statistics")


def fetch_first_commit_years(token: str, repository: str) -> dict[str, int]:
    first_years: dict[str, int] = {}
    page = 1

    while True:
        query = urllib.parse.urlencode(
            {
                "sha": BASE_BRANCH,
                "per_page": 100,
                "page": page,
            }
        )
        _, commits = request_json(
            token,
            f"https://api.github.com/repos/{repository}/commits?{query}",
        )
        if not commits:
            break

        for commit in commits:
            author = commit.get("author")
            parents = commit.get("parents", [])
            if author is None or len(parents) > 1:
                continue

            login = author["login"]
            authored_at = commit["commit"]["author"]["date"]
            year = int(authored_at[:4])
            key = login.casefold()
            first_years[key] = min(first_years.get(key, year), year)

        if len(commits) < 100:
            break
        page += 1

    return first_years


def weekly_first_year(weeks: list[dict[str, int]]) -> int:
    active_weeks = [week["w"] for week in weeks if week["c"] > 0]
    if not active_weeks:
        raise RuntimeError("Contributor has no active commit weeks")
    return datetime.fromtimestamp(min(active_weeks), timezone.utc).year


def collect_contributors(
    stats: list[dict[str, Any]],
    first_years: dict[str, int],
) -> list[Contributor]:
    excluded = {login.casefold() for login in EXCLUDED_LOGINS}
    contributors: list[Contributor] = []

    for item in stats:
        author = item.get("author")
        if author is None:
            continue

        login = author["login"]
        key = login.casefold()
        if (
            author.get("type") == "Bot"
            or key.endswith("[bot]")
            or key in excluded
        ):
            continue

        weeks = item["weeks"]
        contributors.append(
            {
                "login": login,
                "commits": item["total"],
                "additions": sum(week["a"] for week in weeks),
                "first_year": first_years.get(key, weekly_first_year(weeks)),
            }
        )

    return contributors


def qualifies(contributor: Contributor) -> bool:
    return (
        contributor["commits"] >= MIN_COMMITS
        and contributor["additions"] >= MIN_ADDITIONS
    )


def generate_notice(
    contributors: list[Contributor],
) -> tuple[str, list[Contributor], list[Contributor]]:
    included = [contributor for contributor in contributors if qualifies(contributor)]
    skipped = [contributor for contributor in contributors if not qualifies(contributor)]
    names_by_year: dict[int, list[str]] = defaultdict(list)
    manual_logins = {login.casefold() for login in MANUAL_CONTRIBUTORS}
    for contributor in included:
        if contributor["login"].casefold() not in manual_logins:
            names_by_year[contributor["first_year"]].append(contributor["login"])

    lines = [
        f"Copyright (c) {year} {login}"
        for login, year in sorted(
            MANUAL_CONTRIBUTORS.items(),
            key=lambda item: (item[1], item[0].casefold()),
        )
    ]
    for year in sorted(names_by_year):
        names = sorted(set(names_by_year[year]), key=str.casefold)
        lines.append(f"Copyright (c) {year} {', '.join(names)}")

    current_year = datetime.now(timezone.utc).year
    lines.append(f"Copyright (c) 2016-{current_year} Ikemen GO contributors")
    return "\n".join(lines), included, skipped


def update_licence(notice: str) -> bool:
    path = Path("LICENCE.txt")
    text = path.read_text(encoding="utf-8")
    pattern = re.compile(
        r"\AMIT License\r?\n\r?\n"
        r"(?:Copyright \(c\)[^\r\n]*\r?\n)+"
        r"\r?\n(?=Permission is hereby granted)",
    )
    updated, replacements = pattern.subn(
        f"MIT License\n\n{notice}\n\n",
        text,
        count=1,
    )
    if replacements != 1:
        raise RuntimeError("Could not find the copyright block in LICENCE.txt")
    if updated == text:
        return False

    path.write_text(updated, encoding="utf-8", newline="\n")
    return True


def print_audit(included: list[Contributor], skipped: list[Contributor]) -> None:
    print(
        f"Policy: >= {MIN_COMMITS} non-merge commits to {BASE_BRANCH} "
        f"and >= {MIN_ADDITIONS} cumulative additions."
    )
    for heading, contributors in (
        ("Included", included),
        ("Covered only by the collective notice", skipped),
    ):
        print(f"\n{heading}:")
        for contributor in sorted(contributors, key=lambda item: item["login"].casefold()):
            print(
                f"  {contributor['login']}: {contributor['commits']} commits, "
                f"{contributor['additions']} additions, first contribution "
                f"{contributor['first_year']}"
            )


def main() -> int:
    token = os.environ.get("GITHUB_TOKEN")
    repository = os.environ.get("GITHUB_REPOSITORY", "ikemen-engine/Ikemen-GO")
    if not token:
        print("GITHUB_TOKEN is required", file=sys.stderr)
        return 2

    try:
        stats = fetch_contributor_stats(token, repository)
        first_years = fetch_first_commit_years(token, repository)
        contributors = collect_contributors(stats, first_years)
        notice, included, skipped = generate_notice(contributors)
        changed = update_licence(notice)
    except (RuntimeError, ValueError, KeyError, TypeError) as exc:
        print(exc, file=sys.stderr)
        return 1

    print_audit(included, skipped)
    print("\nLICENCE.txt updated." if changed else "\nLICENCE.txt is already current.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
