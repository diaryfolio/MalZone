from __future__ import annotations

import json
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]


class PocContractTests(unittest.TestCase):
    def test_openapi_is_versioned_and_canary_only(self) -> None:
        contract = json.loads(
            (REPO_ROOT / "contracts/openapi/poc-v1alpha1.openapi.json").read_text(
                encoding="utf-8"
            )
        )
        self.assertEqual("3.1.0", contract["openapi"])
        self.assertEqual(
            "canary",
            contract["components"]["schemas"]["CanarySample"]["properties"]["kind"]["const"],
        )
        paths = contract["paths"]
        for path in (
            "/healthz",
            "/readyz",
            "/api/v1alpha1/analyses",
            "/api/v1alpha1/analyses/{name}",
            "/api/v1alpha1/analyses/{name}/cancel",
        ):
            self.assertIn(path, paths)

    def test_crd_and_chart_preserve_poc_boundaries(self) -> None:
        crd = (REPO_ROOT / "charts/malzone/crds/malzone.io_analyses.yaml").read_text(
            encoding="utf-8"
        )
        network_policy = (
            REPO_ROOT / "charts/malzone/templates/networkpolicy.yaml"
        ).read_text(encoding="utf-8")
        service_accounts = (
            REPO_ROOT / "charts/malzone/templates/serviceaccounts.yaml"
        ).read_text(encoding="utf-8")
        self.assertIn("enum: [canary]", crd)
        self.assertIn("maximum: 60", crd)
        self.assertIn("malzone-poc-runner-deny-all", network_policy)
        self.assertIn("automountServiceAccountToken: false", service_accounts)

    def test_windows_starter_is_halted_and_has_no_pod_network(self) -> None:
        starter = (
            REPO_ROOT
            / "examples/windows/windows11-enterprise-template.yaml.example"
        ).read_text(encoding="utf-8")
        self.assertIn("runStrategy: Halted", starter)
        self.assertIn("autoattachPodInterface: false", starter)
        self.assertIn("REPLACE_WRITABLE_GOLDEN_CLONE_PVC", starter)
        self.assertNotIn("http://", starter)
        self.assertNotIn("https://", starter)


if __name__ == "__main__":
    unittest.main()
