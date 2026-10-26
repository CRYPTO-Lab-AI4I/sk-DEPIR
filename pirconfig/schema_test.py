"""Structural catalogue tests; run with jsonschema==4.25.1 (see README.md)."""
import copy
import json
from pathlib import Path
import unittest

from jsonschema import Draft202012Validator, ValidationError

ROOT = Path(__file__).resolve().parent


class SchemaTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        schema = json.loads((ROOT / "schema.json").read_text())
        Draft202012Validator.check_schema(schema)
        cls.validator = Draft202012Validator(schema)
        cls.fixture = json.loads((ROOT / "testdata/reference.json").read_text())

    def test_catalogues(self):
        self.validator.validate(self.fixture)
        self.validator.validate(json.loads((ROOT / "configs_catalogue.json").read_text()))

    def test_status_and_evidence(self):
        invalid = []
        doc = copy.deepcopy(self.fixture)
        del doc["configurations"][0]["code"]["polynomial"]
        invalid.append(doc)
        doc = copy.deepcopy(self.fixture)
        doc["configurations"][0]["capacity"]["status"] = "unknown"
        invalid.append(doc)
        doc = copy.deepcopy(self.fixture)
        doc["configurations"][0]["capacity"]["evidence"][0]["kind"] = "paper"
        invalid.append(doc)
        doc = copy.deepcopy(self.fixture)
        doc["configurations"][0]["capacity"]["evidence"][0]["kind"] = "monte_carlo"
        invalid.append(doc)
        doc = copy.deepcopy(self.fixture)
        doc["configurations"][0]["capacity"]["evidence"][0]["kind"] = "encoding"
        invalid.append(doc)
        doc = copy.deepcopy(self.fixture)
        del doc["configurations"][0]["query"]["dummy_positions"]
        invalid.append(doc)
        doc = copy.deepcopy(self.fixture)
        doc["extra"] = True
        invalid.append(doc)
        doc = copy.deepcopy(self.fixture)
        doc["configurations"][0]["code"]["polynomial"] = None
        invalid.append(doc)
        for i, doc in enumerate(invalid):
            with self.subTest(case=i), self.assertRaises(ValidationError):
                self.validator.validate(doc)

    def test_empirical_estimate(self):
        # Synthetic metadata exercises the format; these are not measured results.
        doc = copy.deepcopy(self.fixture)
        entry = doc["configurations"][0]
        entry["encoding_status"] = "candidate"
        entry["capacity"]["status"] = "estimated"
        entry["capacity"]["evidence"] = [{
            "kind": "monte_carlo", "reference": "synthetic/report.md",
            "script": "synthetic/run.py", "results": "synthetic/results.json",
            "method": "Synthetic test only",
            "monte_carlo": {"samples": 1000, "confidence": 0.95,
                            "lower_rate": 0.28, "upper_rate": 0.35},
        }]
        self.validator.validate(doc)
        entry["capacity"]["evidence"][0]["monte_carlo"]["confidence"] = 1
        with self.assertRaises(ValidationError):
            self.validator.validate(doc)


if __name__ == "__main__":
    unittest.main()
