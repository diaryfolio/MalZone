from __future__ import annotations

import tempfile
import unittest
from pathlib import Path

from scripts.prepare_pages import prepare


class PagesTests(unittest.TestCase):
    def test_pages_source_contains_allow_list_and_render_metadata(self) -> None:
        with tempfile.TemporaryDirectory() as temporary:
            output = Path(temporary) / "site"
            prepare(output)
            expected = (
                "index.md",
                "README.md",
                "docs/design/high-level/design_01.md",
                "docs/design/high-level/10-overall/07-ai-automation-siem.md",
                "docs/design/business/business-value-and-market-strategy.md",
                "contracts/openapi/poc-v1alpha1.openapi.json",
                "_layouts/default.html",
                "assets/site.css",
                "assets/og.png",
            )
            for relative in expected:
                self.assertTrue((output / relative).is_file(), relative)
            self.assertTrue((output / "index.md").read_text(encoding="utf-8").startswith("---\n"))
            index = (output / "index.md").read_text(encoding="utf-8")
            self.assertNotIn(".md)", index)
            self.assertIn('href="docs/design/high-level/README.html"', index)
            self.assertIn("Keep malware analysis", index)
            self.assertIn('href="CLAUDE.html"', index)
            layout = (output / "_layouts/default.html").read_text(encoding="utf-8")
            self.assertIn('property="og:image"', layout)
            self.assertIn("/assets/og.png", layout)
            self.assertFalse((output / "docs/prompts/v1.md").exists())


if __name__ == "__main__":
    unittest.main()
