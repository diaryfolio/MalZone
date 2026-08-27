from __future__ import annotations

import unittest
from pathlib import Path

from scripts.check_design_sync import DesignSyncError, validate_report


REPO_ROOT = Path(__file__).resolve().parents[1]


def report(classification: str = "minor") -> str:
    return f"""
## Design Sync Report
- Change Classification: {classification}
- Design Docs Updated: none; behavior unchanged
- Contracts Updated: none
- Code/Deployment Areas Updated: tests
- Architecture Delta: none
- Threat/Trust Boundary Delta: none
- Tests/Evidence: make design-check
- Known Production Gaps: none
- Sync Status: PASS
"""


class GovernanceTests(unittest.TestCase):
    def test_governance_pack_and_agent_instructions_exist(self) -> None:
        required = (
            "AGENTS.md",
            "CLAUDE.md",
            "CONTRIBUTING.md",
            "contracts/README.md",
            ".github/pull_request_template.md",
            ".github/workflows/design-sync.yml",
            "docs/prompts/governance/README.md",
            "docs/prompts/governance/major-change-policy.md",
            "docs/prompts/governance/design-sync-checklist.md",
            "docs/prompts/governance/design-sync-guardrail.prompt.md",
            "docs/prompts/governance/major-change-template.md",
        )
        for relative in required:
            self.assertTrue((REPO_ROOT / relative).is_file(), relative)

    def test_canonical_instructions_contain_required_report(self) -> None:
        instructions = (REPO_ROOT / "CLAUDE.md").read_text(encoding="utf-8")
        for phrase in (
            "Mandatory design-sync workflow",
            "00-implementation-conformance.md",
            "Threat/Trust Boundary Delta",
            "Sync Status: <PASS|FAIL>",
            "make design-check",
        ):
            self.assertIn(phrase, instructions)

    def test_minor_report_can_pass_without_design_diff(self) -> None:
        validate_report(report(), ["tests/test_example.py"])

    def test_major_report_requires_conformance_and_companion_design(self) -> None:
        with self.assertRaisesRegex(DesignSyncError, "conformance"):
            validate_report(report("major"), ["controller/reconciler.go"])
        with self.assertRaisesRegex(DesignSyncError, "high-level design or ADR"):
            validate_report(report("major"), [
                "controller/reconciler.go",
                "docs/design/high-level/00-implementation-conformance.md",
            ])
        validate_report(report("major"), [
            "controller/reconciler.go",
            "docs/design/high-level/00-implementation-conformance.md",
            "docs/design/high-level/10-overall/02-runtime-topology-lifecycle.md",
        ])

    def test_unresolved_or_failed_report_is_rejected(self) -> None:
        with self.assertRaises(DesignSyncError):
            validate_report(report().replace("minor", "<minor|major>", 1), [])
        with self.assertRaisesRegex(DesignSyncError, "PASS"):
            validate_report(report().replace("Sync Status: PASS", "Sync Status: FAIL"), [])


if __name__ == "__main__":
    unittest.main()
