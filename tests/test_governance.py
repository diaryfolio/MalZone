from __future__ import annotations

import unittest
from pathlib import Path

from scripts.check_design_sync import DesignSyncError, validate_changed_files


REPO_ROOT = Path(__file__).resolve().parents[1]

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

    def test_canonical_instructions_describe_automatic_gate(self) -> None:
        instructions = (REPO_ROOT / "CLAUDE.md").read_text(encoding="utf-8")
        for phrase in (
            "Mandatory design-sync workflow",
            "00-implementation-conformance.md",
            "changed-file design-sync gate",
            "pull-request description fields",
            "make design-check",
        ):
            self.assertIn(phrase, instructions)

    def test_documentation_only_change_does_not_require_companion_updates(self) -> None:
        validate_changed_files(["README.md", "docs/design/business/strategy.md"])

    def test_sensitive_change_requires_conformance_and_companion_design(self) -> None:
        with self.assertRaisesRegex(DesignSyncError, "conformance"):
            validate_changed_files(["controller/reconciler.go"])
        with self.assertRaisesRegex(DesignSyncError, "high-level design or ADR"):
            validate_changed_files([
                "controller/reconciler.go",
                "docs/design/high-level/00-implementation-conformance.md",
            ])
        validate_changed_files([
            "controller/reconciler.go",
            "docs/design/high-level/00-implementation-conformance.md",
            "docs/design/high-level/10-overall/02-runtime-topology-lifecycle.md",
        ])

    def test_contract_change_is_implementation_sensitive(self) -> None:
        with self.assertRaisesRegex(DesignSyncError, "conformance"):
            validate_changed_files(["contracts/schemas/event-v1.json"])


if __name__ == "__main__":
    unittest.main()
