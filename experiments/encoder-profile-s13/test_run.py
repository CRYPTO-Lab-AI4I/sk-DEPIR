"""Regression for logger compatibility with real post-load driver metadata."""
import importlib.util,json,unittest
from pathlib import Path
HERE=Path(__file__).resolve().parent
spec=importlib.util.spec_from_file_location('profile_run',HERE/'run.py')
runner=importlib.util.module_from_spec(spec);spec.loader.exec_module(runner)
class MetadataLogging(unittest.TestCase):
    def test_post_load_code_does_not_require_discovery_samples(self):
        fixture=json.loads((HERE/'smoke-results.json').read_text())
        for check in fixture['checks']:
            code=check['code']
            self.assertNotIn('samples',code)
            self.assertIn('exact K='+str(code['k']),runner.describe_code(code))
            self.assertIn(str(code['reused_cache']),runner.describe_code(code))
if __name__=='__main__':unittest.main()
