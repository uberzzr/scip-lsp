import inspect
import os
import tempfile
import unittest
from unittest.mock import MagicMock, patch

from bsp_server.scip_sync_util import scip_utils
from bsp_server.scip_sync_util.scip_const import WORKSPACE_FILE_NAME


class TestScipUtils(unittest.TestCase):
    def setUp(self):
        self.cwd = tempfile.mkdtemp()
        self.bazel_project_file = self.resource_path("test.bazelproject")
        self.filepath = "path/to/src/file.java"
        self.query_kinds = ["java_library"]
        self.expected_from_target = "//path/to/..."
        self.expected_query_string = (
            'kind("java_library", rdeps("//path/to/...", "path/to/src/file.java"))'
        )
        self.expected_cmd = ["bazel", "query", self.expected_query_string]

    @staticmethod
    def resource_path(resource_name):
        resource_root = os.environ.get("RESOURCE_ROOT", "")
        # https://stackoverflow.com/questions/37546685/get-absolute-path-of-caller-file/37547263#37547263
        src_file = (inspect.stack()[1])[1]
        resources_dir = os.path.join(
            os.path.dirname(os.path.abspath(src_file)), resource_root, "resources"
        )
        return os.path.join(resources_dir, resource_name)

    def test_parse_bazel_project(self):
        actual_data = scip_utils.parse_bazelproject(self.bazel_project_file)

        expected_data = {
            "directories": ["search", "tooling/intellij"],
            "derive_targets_from_directories": ["false"],
            "targets": [
                "//search/service-grpc/...",
                "//experimental/users/hshukla/scip/java-sample:src_main",
            ],
            "additional_languages": ["scala"],
            "bazel_binary": ["/home/user/fievel/tools/bazel"],
            "test_sources": ["*src/integrationTest/*", "*src/test/*"],
        }

        self.assertDictEqual(actual_data, expected_data, "parse_bazelproject_mismatch")

    @patch("bsp_server.util.utils.output")
    def test_get_containing_bazel_target(self, mock_output):
        # Setup mock
        mock_output.return_value = "//path/to:target"

        # Call the function
        result = scip_utils.get_containing_bazel_target(
            self.cwd, self.filepath, self.query_kinds
        )

        # Verify the result
        self.assertEqual(result, "//path/to:target")

        # Verify output was called with correct arguments
        mock_output.assert_called_once_with(self.expected_cmd, cwd=self.cwd)

    @patch("bsp_server.util.utils.output")
    def test_get_containing_bazel_target_multiple_query_kinds(self, mock_output):
        # Setup mock
        mock_output.return_value = "//path/to:target"
        query_kinds = ["java_library", "java_binary"]
        expected_query_string = 'kind("java_library|java_binary", rdeps("//path/to/...", "path/to/src/file.java"))'
        expected_cmd = ["bazel", "query", expected_query_string]

        # Call the function
        result = scip_utils.get_containing_bazel_target(
            self.cwd, self.filepath, query_kinds
        )

        # Verify the result
        self.assertEqual(result, "//path/to:target")

        # Verify output was called with correct arguments
        mock_output.assert_called_once_with(expected_cmd, cwd=self.cwd)

    @patch("bsp_server.util.utils.safe_create")
    @patch("bsp_server.scip_sync_util.scip_utils.get_sha256_for_file")
    @patch("os.remove")
    @patch("shutil.copy")
    @patch("os.listdir")
    @patch("os.path.join")
    @patch("bsp_server.scip_sync_util.scip_utils.get_thread_pool_size")
    @patch("shutil.rmtree")
    @patch("os.path.isfile")
    def test_copy_index(
        self,
        mock_isfile,
        mock_rmtree,
        mock_thread_pool,
        mock_join,
        mock_listdir,
        mock_copy,
        mock_remove,
        mock_get_sha,
        mock_safe_create,
    ):
        # Setup test data
        index_to_copy = {
            "bazel-out/bin/path/to/index1.scip",
            "bazel-out/bin/path/to/index2.scip",
            "bazel-out/bin/path/to/existing_index.scip",
            "bazel-out/bin/path/to/failing_index.scip",
        }
        dest = "/path/to/dest"

        mock_thread_pool.return_value = 2
        mock_listdir.return_value = [
            "path_to_existing_index.scip",
            "path_to_existing_index.scip.sha256",
            "old_index.scip",
            "old_index.scip.sha256",
            "jdk_temurin_11.scip",
            "jdk_temurin_11.scip.sha256",
            WORKSPACE_FILE_NAME,
        ]
        mock_join.side_effect = lambda *args: "/".join(args)

        # Setup isfile mock to return True for existing files
        def isfile_side_effect(path):
            return any(filename in path for filename in mock_listdir.return_value)

        mock_isfile.side_effect = isfile_side_effect

        # Setup SHA returns
        def get_sha_side_effect(file_path):
            if "failing_index" in file_path:
                raise Exception("Simulated failure")
            if "existing_index" in file_path:
                return "same_sha"
            elif "path_to_existing_index" in file_path:
                return "same_sha"
            return "new_sha"

        mock_get_sha.side_effect = get_sha_side_effect

        # Setup copy to fail for failing_index
        def copy_side_effect(src, dst):
            if "failing_index" in src:
                raise Exception("Simulated copy failure")

        mock_copy.side_effect = copy_side_effect

        mock_rmtree.side_effect = lambda path: None

        # Call the function
        scip_utils.copy_index(index_to_copy, dest)

        mock_safe_create.assert_called_once_with(dest, is_dir=True)
        expected_copies = [
            (
                (
                    "bazel-out/bin/path/to/index1.scip",
                    "/path/to/dest/path_to_index1.scip",
                ),
                {},
            ),
            (
                (
                    "bazel-out/bin/path/to/index2.scip",
                    "/path/to/dest/path_to_index2.scip",
                ),
                {},
            ),
            (
                (
                    "bazel-out/bin/path/to/index1.scip.sha256",
                    "/path/to/dest/path_to_index1.scip.sha256",
                ),
                {},
            ),
            (
                (
                    "bazel-out/bin/path/to/index2.scip.sha256",
                    "/path/to/dest/path_to_index2.scip.sha256",
                ),
                {},
            ),
        ]

        # Verify copy calls - should only be 4 copies (2 new files + their SHA files)
        # Note: failing_index should not be counted in successful copies
        self.assertEqual(mock_copy.call_count, 4)
        mock_copy.assert_has_calls(expected_copies, any_order=True)

        expected_removes = [
            ((("/path/to/dest/old_index.scip"),), {}),
            ((("/path/to/dest/old_index.scip.sha256"),), {}),
        ]
        self.assertEqual(mock_remove.call_count, 2)
        mock_remove.assert_has_calls(expected_removes, any_order=True)

        # Verify JDK and workspace files were NOT removed
        for file_path in [
            "/path/to/dest/jdk_temurin_11.scip",
            "/path/to/dest/jdk_temurin_11.scip.sha256",
            f"/path/to/dest/{WORKSPACE_FILE_NAME}",
        ]:
            remove_call = (((file_path,),), {})
            self.assertNotIn(remove_call, mock_remove.call_args_list)

    @patch("builtins.open")
    @patch("shutil.copy")
    @patch("bsp_server.scip_sync_util.scip_utils.generate_sha256")
    def test_old_copy_index(self, m_generate_sha256, m_copy, m_open):
        m_generate_sha256.return_value = "someHashAbc"
        gen_scip = "/src/execroot/__main__/bazel-out/k8-fastbuild/bin/some_path/some_index.scip"
        index_to_copy = [gen_scip]

        scip_utils.old_copy_index(index_to_copy, os.path.join(self.cwd, "dest"))

        m_copy.assert_called_once_with(
            gen_scip,
            os.path.join(self.cwd, "dest", "some_path_some_index.scip"),
        )
        m_open.assert_called_once_with(
            os.path.join(self.cwd, "dest", "some_path_some_index.scip.sha256"),
            "w",
        )
        m_generate_sha256.assert_called_once_with(gen_scip)

    @patch("builtins.open")
    def test_get_sha256_for_file_success(self, mock_open):
        mock_file = mock_open.return_value.__enter__.return_value
        mock_file.read.return_value = "abc123def456  \n"

        result = scip_utils.get_sha256_for_file("/path/to/file.sha256")

        self.assertEqual(result, "abc123def456")
        mock_open.assert_called_once_with("/path/to/file.sha256", "r")

    @patch("builtins.open")
    def test_get_sha256_for_file_file_not_found(self, mock_open):
        mock_open.side_effect = FileNotFoundError()

        result = scip_utils.get_sha256_for_file("/path/to/nonexistent.sha256")

        self.assertIsNone(result)
        mock_open.assert_called_once_with("/path/to/nonexistent.sha256", "r")

    def test_transform_bazel_query_results_with_empty_query_result(self):
        qr = []
        result = scip_utils.transform_bazel_query_results(qr)
        self.assertEqual(result, {})

    def test_transform_bazel_query_results_with_rules(self):
        qr = [
            {
                "type": "RULE",
                "rule": {
                    "name": "//path/to:target",
                    "ruleClass": "java_library",
                    "attribute": [
                        {"name": "deps", "stringListValue": ["//path/to:dep1"]},
                        {"name": "exports", "stringListValue": ["//path/to:export"]},
                    ],
                    "ruleInput": ["//path/to:dep1", "//path/to:dep2"],
                },
            },
            {
                "type": "RULE",
                "rule": {
                    "name": "//path/to:dep1",
                    "ruleClass": "java_library",
                    "attribute": [
                        {"name": "data", "stringListValue": ["//path/to:data"]}
                    ],
                },
            },
            {
                "type": "RULE",
                "rule": {"name": "//path/to:export", "ruleClass": "java_library"},
            },
            {
                "type": "RULE",
                "rule": {"name": "//path/to:data", "ruleClass": "java_library"},
            },
        ]
        result = scip_utils.transform_bazel_query_results(qr)
        expected = {
            "//path/to:data": {
                "base_path": "path/to",
                "deps": [],
                "direct_deps": [],
                "exports": [],
                "target_type": "java_library",
            },
            "//path/to:dep1": {
                "base_path": "path/to",
                "deps": ["//path/to:data"],
                "direct_deps": [],
                "exports": [],
                "target_type": "java_library",
            },
            "//path/to:export": {
                "base_path": "path/to",
                "deps": [],
                "direct_deps": [],
                "exports": [],
                "target_type": "java_library",
            },
            "//path/to:target": {
                "base_path": "path/to",
                "deps": ["//path/to:dep1"],
                "direct_deps": ["//path/to:dep1"],
                "exports": ["//path/to:export"],
                "target_type": "java_library",
            },
        }
        self.assertEqual(result, expected)


if __name__ == "__main__":
    unittest.main()
