from __future__ import annotations

import json
import re
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
SCHEMA_ROOT = REPO_ROOT / "contracts" / "schemas"
EXAMPLE_ROOT = REPO_ROOT / "contracts" / "software" / "examples"


class SoftwareContractTests(unittest.TestCase):
    def test_schemas_are_draft_2020_12_closed_objects(self) -> None:
        expected = {
            "software-package-version-v1alpha1.schema.json",
            "windows-image-recipe-v1alpha1.schema.json",
        }
        self.assertEqual(expected, {path.name for path in SCHEMA_ROOT.glob("*.json")})
        for path in SCHEMA_ROOT.glob("*.json"):
            schema = json.loads(path.read_text(encoding="utf-8"))
            self.assertEqual(
                "https://json-schema.org/draft/2020-12/schema", schema["$schema"]
            )
            self.assertFalse(schema["additionalProperties"], path)
            self.assertEqual("malzone.io/v1alpha1", schema["properties"]["apiVersion"]["const"])
            refs: list[str] = []

            def collect(value: object) -> None:
                if isinstance(value, dict):
                    if "$ref" in value:
                        refs.append(str(value["$ref"]))
                    for nested in value.values():
                        collect(nested)
                elif isinstance(value, list):
                    for nested in value:
                        collect(nested)

            collect(schema)
            for ref in refs:
                self.assertTrue(ref.startswith("#/$defs/"), (path, ref))
                self.assertIn(ref.removeprefix("#/$defs/"), schema["$defs"], (path, ref))

            exact = re.compile(schema["$defs"]["exactVersion"]["pattern"])
            self.assertIsNotNone(exact.fullmatch("1.2.3"))
            for forbidden in ("latest", "CURRENT", "Stable", ">=1.0", "1.*"):
                self.assertIsNone(exact.fullmatch(forbidden), (path, forbidden))

    def test_package_example_is_exact_hash_bound_and_non_secret(self) -> None:
        package = json.loads(
            (EXAMPLE_ROOT / "example-browser-1.0.0.json").read_text(encoding="utf-8")
        )
        self.assertEqual("SoftwarePackageVersion", package["kind"])
        self.assertEqual(64, len(package["spec"]["source"]["sha256"]))
        self.assertIsInstance(package["spec"]["installer"]["arguments"], list)
        serialized = json.dumps(package).lower()
        for forbidden in ("latest", "password", "api_key", "client_secret", "http://", "https://"):
            self.assertNotIn(forbidden, serialized)

    def test_recipe_pins_base_platform_and_exact_package_revisions(self) -> None:
        recipe = json.loads(
            (EXAMPLE_ROOT / "windows-11-browser-lab.json").read_text(encoding="utf-8")
        )
        self.assertEqual("WindowsImageRecipe", recipe["kind"])
        self.assertEqual(64, len(recipe["spec"]["baseImage"]["digest"]))
        self.assertEqual(64, len(recipe["spec"]["platformBundle"]["digest"]))
        self.assertEqual("offline-mirror-only", recipe["spec"]["buildPolicy"]["network"])
        for package in recipe["spec"]["packages"]:
            self.assertTrue(package["version"])
            self.assertNotEqual("latest", package["version"].lower())
            self.assertGreaterEqual(package["revision"], 1)


if __name__ == "__main__":
    unittest.main()
