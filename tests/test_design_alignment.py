from __future__ import annotations

import re
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
DESIGN_ROOT = REPO_ROOT / "docs" / "design" / "high-level"
MARKDOWN_LINK = re.compile(r"(?<!!)\[[^]]+\]\(([^)]+)\)")


class DesignAlignmentTests(unittest.TestCase):
    def test_required_canonical_documents_exist(self) -> None:
        required = {
            "README.md",
            "design_01.md",
            "00-implementation-conformance.md",
            "10-overall/01-objectives-principles.md",
            "10-overall/02-runtime-topology-lifecycle.md",
            "10-overall/03-contracts-data.md",
            "10-overall/04-components-technology.md",
            "20-deployment/01-kubernetes-kubevirt.md",
            "30-security/01-threat-model-zero-trust.md",
            "40-operations/01-day2-sre.md",
            "50-roadmap/01-delivery-roadmap.md",
        }
        self.assertEqual(required, {
            str(path.relative_to(DESIGN_ROOT))
            for path in DESIGN_ROOT.rglob("*.md")
        })

    def test_high_level_documents_have_balanced_mermaid_diagrams(self) -> None:
        for document in DESIGN_ROOT.rglob("*.md"):
            text = document.read_text(encoding="utf-8")
            self.assertIn("```mermaid", text, document)
            self.assertEqual(
                text.count("```mermaid"),
                len(re.findall(r"```mermaid\n.*?\n```", text, re.DOTALL)),
                document,
            )

    def test_relative_markdown_links_resolve(self) -> None:
        roots = [REPO_ROOT / "README.md", *REPO_ROOT.joinpath("docs").rglob("*.md")]
        for document in roots:
            for target in MARKDOWN_LINK.findall(document.read_text(encoding="utf-8")):
                target = target.split("#", 1)[0]
                if not target or "://" in target or target.startswith("mailto:"):
                    continue
                resolved = (document.parent / target).resolve()
                self.assertTrue(resolved.exists(), f"{document}: missing {target}")

    def test_conformance_does_not_claim_runtime_is_implemented(self) -> None:
        conformance = (DESIGN_ROOT / "00-implementation-conformance.md").read_text(
            encoding="utf-8"
        )
        self.assertIn("There are no executable runtime services", conformance)
        for capability in (
            "Analysis REST/WebSocket API",
            "`Analysis` CRD and operator",
            "Disposable KubeVirt clone",
            "Windows collection agent",
        ):
            row = next(line for line in conformance.splitlines() if capability in line)
            self.assertIn("| designed |", row)


if __name__ == "__main__":
    unittest.main()

