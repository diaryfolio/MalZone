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
            "10-overall/05-software-catalog-image-composition.md",
            "10-overall/06-api-identity-observability-integrations.md",
            "10-overall/07-ai-automation-siem.md",
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
        roots = [
            path
            for path in REPO_ROOT.rglob("*.md")
            if "build" not in path.relative_to(REPO_ROOT).parts
        ]
        for document in roots:
            for target in MARKDOWN_LINK.findall(document.read_text(encoding="utf-8")):
                target = target.split("#", 1)[0]
                if not target or "://" in target or target.startswith("mailto:"):
                    continue
                resolved = (document.parent / target).resolve()
                self.assertTrue(resolved.exists(), f"{document}: missing {target}")

    def test_conformance_distinguishes_poc_from_production_runtime(self) -> None:
        conformance = (DESIGN_ROOT / "00-implementation-conformance.md").read_text(
            encoding="utf-8"
        )
        self.assertIn("There are no executable **production** runtime services", conformance)
        for capability in (
            "Harmless lifecycle POC API",
            "POC `Analysis` CRD and Job operator",
            "POC Helm packaging and runner isolation",
        ):
            row = next(line for line in conformance.splitlines() if capability in line)
            self.assertIn("| implemented |", row)
        for capability in (
            "Production Analysis REST/WebSocket API",
            "Production `Analysis` CRD and KubeVirt operator",
            "Disposable KubeVirt clone",
            "Windows collection agent",
        ):
            row = next(line for line in conformance.splitlines() if capability in line)
            self.assertIn("| designed |", row)

    def test_every_implemented_runtime_route_is_in_conformance(self) -> None:
        conformance = (DESIGN_ROOT / "00-implementation-conformance.md").read_text(
            encoding="utf-8"
        )
        patterns = (
            re.compile(r'@(?:app|router)\.(?:get|post|put|patch|delete)\(\s*["\']([^"\']+)'),
            re.compile(r'\b(?:app|router|r)\.(?:Get|Post|Put|Patch|Delete)\(\s*["\']([^"\']+)'),
            re.compile(
                r'HandleFunc\(\s*["\'](?:GET|POST|PUT|PATCH|DELETE)\s+([^"\']+)'
            ),
        )
        source_roots = ["api", "cmd", "internal", "controller"]
        for root_name in source_roots:
            root = REPO_ROOT / root_name
            if not root.exists():
                continue
            for source in root.rglob("*"):
                if source.suffix not in {".go", ".py", ".ts", ".tsx"}:
                    continue
                if source.name.endswith("_test.go"):
                    continue
                text = source.read_text(encoding="utf-8")
                for pattern in patterns:
                    for route in pattern.findall(text):
                        self.assertIn(route, conformance, f"{source}: undocumented {route}")


if __name__ == "__main__":
    unittest.main()
