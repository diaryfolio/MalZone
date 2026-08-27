from __future__ import annotations

import argparse
import subprocess
import sys
from typing import Iterable


SYNCHRONIZED_PATH_PREFIXES = (
    "api/",
    "charts/",
    "cmd/",
    "config/",
    "contracts/",
    "controller/",
    "deploy/",
    "guest/",
    "helm/",
    "internal/",
    "operator/",
    "pkg/",
    "services/",
)
SYNCHRONIZED_ROOT_FILES = {
    "Dockerfile",
    "docker-compose.yml",
    "go.mod",
    "go.sum",
    "package-lock.json",
    "package.json",
    "pyproject.toml",
    "requirements.txt",
}
CONFORMANCE = "docs/design/high-level/00-implementation-conformance.md"


class DesignSyncError(ValueError):
    """Implementation-sensitive changes are not synchronized with design truth."""


def validate_changed_files(changed_files: Iterable[str]) -> None:
    files = {path.strip() for path in changed_files if path.strip()}
    requires_design_sync = any(
        path in SYNCHRONIZED_ROOT_FILES
        or path.startswith(SYNCHRONIZED_PATH_PREFIXES)
        for path in files
    )
    if not requires_design_sync:
        return

    if CONFORMANCE not in files:
        raise DesignSyncError(
            "implementation, deployment, or contract changes must update the "
            "implementation conformance map"
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
            "implementation, deployment, or contract changes must update an affected "
            "high-level design or ADR"
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
    parser.add_argument("--base", required=True)
    args = parser.parse_args()

    try:
        validate_changed_files(changed_files_against(args.base))
    except (DesignSyncError, subprocess.CalledProcessError) as error:
        print(f"design-sync gate failed: {error}", file=sys.stderr)
        return 1
    print("design-sync gate passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
