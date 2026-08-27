from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
from pathlib import Path
from typing import Iterable


REPORT_FIELDS = (
    "Change Classification",
    "Design Docs Updated",
    "Contracts Updated",
    "Code/Deployment Areas Updated",
    "Architecture Delta",
    "Threat/Trust Boundary Delta",
    "Tests/Evidence",
    "Known Production Gaps",
    "Sync Status",
)
CONFORMANCE = "docs/design/high-level/00-implementation-conformance.md"


class DesignSyncError(ValueError):
    """The PR Design Sync Report is absent or inconsistent with its changed files."""


def _report_values(body: str) -> dict[str, str]:
    values: dict[str, str] = {}
    for field in REPORT_FIELDS:
        match = re.search(
            rf"^\s*-\s*{re.escape(field)}:\s*(.*?)\s*$",
            body,
            flags=re.MULTILINE,
        )
        if not match:
            raise DesignSyncError(f"missing Design Sync Report field: {field}")
        value = match.group(1).strip()
        if not value or (value.startswith("<") and value.endswith(">")):
            raise DesignSyncError(f"unresolved Design Sync Report field: {field}")
        values[field] = value
    return values


def validate_report(body: str, changed_files: Iterable[str]) -> None:
    files = {path.strip() for path in changed_files if path.strip()}
    values = _report_values(body)

    classification = values["Change Classification"].lower()
    if classification not in {"minor", "major"}:
        raise DesignSyncError("Change Classification must be minor or major")
    if values["Sync Status"].upper() != "PASS":
        raise DesignSyncError("Sync Status must be PASS before merge")

    if classification == "major":
        if CONFORMANCE not in files:
            raise DesignSyncError(
                "major changes must update the implementation conformance map"
            )
        companions = {
            path
            for path in files
            if path.startswith("docs/design/high-level/") and path != CONFORMANCE
        }
        decisions = {
            path
            for path in files
            if path.startswith("docs/design/decisions/") and not path.endswith("README.md")
        }
        if not companions and not decisions:
            raise DesignSyncError(
                "major changes must update an affected high-level design or ADR"
            )


def changed_files_against(base: str) -> list[str]:
    result = subprocess.run(
        ["git", "diff", "--name-only", f"{base}...HEAD"],
        check=True,
        text=True,
        capture_output=True,
    )
    return result.stdout.splitlines()


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--event", type=Path, required=True)
    parser.add_argument("--base", required=True)
    args = parser.parse_args()

    event = json.loads(args.event.read_text(encoding="utf-8"))
    body = (event.get("pull_request") or {}).get("body") or ""
    try:
        validate_report(body, changed_files_against(args.base))
    except (DesignSyncError, subprocess.CalledProcessError) as error:
        print(f"design-sync gate failed: {error}", file=sys.stderr)
        return 1
    print("design-sync gate passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

