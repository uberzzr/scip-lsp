import os
import unittest
from unittest.mock import patch

from bsp_server.scip_sync_util.incremental import AGGREGATOR, WORK_DIR, index_file


class ScipIncrementalTest(unittest.TestCase):
    def setUp(self):
        self.test_cwd = "/test/cwd"
        self.test_file = "path/to/test_file.scala"
        self.test_manifest = "path/to/manifest.json"
        self.test_roots = "root1:root2"
        self.test_scip_file = os.path.join(
            self.test_cwd, WORK_DIR, "path_to_test_file_scala.scip"
        )

    @patch("bsp_server.util.utils.check")
    @patch("bsp_server.util.utils.safe_create")
    def test_index_file(self, mock_safe_create, mock_check):
        result = index_file(self.test_cwd, self.test_file, self.test_manifest)

        # Verify work directory is created
        mock_safe_create.assert_called_once_with(
            os.path.join(self.test_cwd, WORK_DIR), is_dir=True
        )

        # Verify command is called with correct arguments
        expected_args = [
            os.path.join(self.test_cwd, AGGREGATOR),
            "-m",
            self.test_manifest,
            "-f",
            self.test_file,
            "-o",
            self.test_scip_file,
        ]
        mock_check.assert_called_once_with(
            expected_args, cwd=os.path.join(self.test_cwd)
        )

        self.assertEqual(result, self.test_scip_file)


if __name__ == "__main__":
    unittest.main()
