#!/usr/bin/env python3
"""Generate an AltStore / SideStore source (altstore-source.json) for the iOS IPA.

App metadata (bundle ID, version, minimum iOS, privacy usage strings) is read from the
IPA itself, because AltStore rejects an install when the source disagrees with the app.
"""

from __future__ import annotations

import argparse
import json
import os
import plistlib
import zipfile
from datetime import datetime, timezone


def read_info_plist(ipa_path: str) -> dict:
    with zipfile.ZipFile(ipa_path) as ipa:
        for name in ipa.namelist():
            parts = name.split("/")
            # Payload/<App>.app/Info.plist (not nested bundles or frameworks)
            if len(parts) == 3 and parts[0] == "Payload" and parts[1].endswith(".app") and parts[2] == "Info.plist":
                return plistlib.loads(ipa.read(name))
    raise SystemExit(f"Info.plist not found in {ipa_path}")


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--ipa", required=True, help="Path to the built IPA")
    parser.add_argument("--download-url", required=True, help="Public URL of the IPA release asset")
    parser.add_argument("--repo", required=True, help="owner/name of the GitHub repository")
    parser.add_argument("--kind", required=True, help="nightly, tagged, ci-test or branch")
    parser.add_argument("--release-tag", required=True, help="Release tag the IPA is attached to")
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    info = read_info_plist(args.ipa)
    bundle_id = info["CFBundleIdentifier"]
    repo_url = f"https://github.com/{args.repo}"
    raw_base = f"https://raw.githubusercontent.com/{args.repo}/develop"

    stable = args.kind == "tagged"
    source_name = "Ikemen GO" if stable else f"Ikemen GO ({args.release_tag})"
    release_url = f"{repo_url}/releases/tag/{args.release_tag}"

    privacy = {
        key: value
        for key, value in info.items()
        if key.startswith("NS") and key.endswith("UsageDescription") and isinstance(value, str)
    }

    source = {
        "name": source_name,
        "identifier": "org.ikemen-engine.ikemen-go" + ("" if stable else f".{args.release_tag}"),
        "subtitle": "Open source fighting game engine",
        "website": repo_url,
        "iconURL": f"{raw_base}/external/icons/IkemenCylia_256.png",
        "tintColor": "#c0392b",
        "featuredApps": [bundle_id],
        "apps": [
            {
                "name": "Ikemen GO",
                "bundleIdentifier": bundle_id,
                "developerName": "Ikemen GO team",
                "subtitle": "Open source fighting game engine",
                "localizedDescription": (
                    "Ikemen GO is an open source fighting game engine that supports resources "
                    "from the M.U.G.E.N engine.\n\n"
                    + ("" if stable else "This is a development build and may be unstable.\n\n")
                    + f"Release notes: {release_url}"
                ),
                "iconURL": f"{raw_base}/external/icons/IkemenCylia_256.png",
                "tintColor": "#c0392b",
                "versions": [
                    {
                        "version": info["CFBundleShortVersionString"],
                        "buildVersion": str(info.get("CFBundleVersion", "")),
                        "date": datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ"),
                        "localizedDescription": f"See {release_url}",
                        "downloadURL": args.download_url,
                        "size": os.path.getsize(args.ipa),
                        "minOSVersion": info.get("MinimumOSVersion", ""),
                    }
                ],
                "appPermissions": {"entitlements": [], "privacy": privacy},
            }
        ],
        "news": [],
    }

    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(source, f, indent=2, ensure_ascii=False)
        f.write("\n")
    print(json.dumps(source["apps"][0]["versions"][0], indent=2))


if __name__ == "__main__":
    main()
