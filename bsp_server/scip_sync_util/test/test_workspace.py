import os
import tempfile
import unittest
from dataclasses import asdict
from typing import Tuple
from unittest.mock import MagicMock, mock_open, patch

from bsp_server.scip_sync_util.workspace import (
    WORKSPACE_FILE_NAME,
    ScipMnemonics,
    ScipWorkspace,
    WorkspaceLinkType,
    add_to_workspace,
    create_workspace,
    get_manifest_for_file,
    populate_workspace,
    write_workspace,
)


class ScipWorkspaceTest(unittest.TestCase):
    def setUp(self):
        self.temp_dir = tempfile.mkdtemp()
        self.cwd = "path/to/cwd"
        self.source_list_path = os.path.join(self.temp_dir, self.cwd, "source_list.txt")
        self.another_source_list_path = os.path.join(
            self.temp_dir, self.cwd, "another_source_list.txt"
        )

        # createparent folder
        os.makedirs(os.path.dirname(self.source_list_path), exist_ok=True)

        # Create a temporary file for testing
        with open(self.source_list_path, "w") as f:
            f.write("file1.java\nfile2.java")

            # createparent folder
        os.makedirs(os.path.dirname(self.another_source_list_path), exist_ok=True)

        # Create a temporary file for testing
        with open(self.another_source_list_path, "w") as f:
            f.write("file3.java\nfile4.java")

    def test_add_file_to_workspace(self):
        workspace = ScipWorkspace()
        workspace.add_file("file1.java", 1, WorkspaceLinkType.BAZEL_TARGET)
        self.assertEqual(workspace.get_file("file1.java"), {"BAZEL_TARGET": "1"})

    def test_add_link_to_workspace(self):
        workspace = ScipWorkspace()
        link_id = workspace.add_link("//path/to:target", WorkspaceLinkType.BAZEL_TARGET)
        self.assertEqual(
            workspace.get_link(link_id, WorkspaceLinkType.BAZEL_TARGET),
            "//path/to:target",
        )

    def test_clear_workspace(self):
        workspace = ScipWorkspace()
        workspace.add_file("file1.java", 1, WorkspaceLinkType.BAZEL_TARGET)
        workspace.clear()
        self.assertEqual(workspace.files, {})
        self.assertEqual(workspace.links, {})
        self.assertEqual(workspace._last_link_id, 0)

    def test_get_link_for_existing_file(self):
        workspace = ScipWorkspace()
        workspace.add_file("file1.java", 1, WorkspaceLinkType.JAVA_MANIFEST)
        workspace.add_link("manifest1", WorkspaceLinkType.JAVA_MANIFEST)
        manifest = workspace.get_link("1", WorkspaceLinkType.JAVA_MANIFEST)
        self.assertEqual(manifest, "manifest1")

    def test_get_manifest_for_non_existing_file(self):
        workspace = ScipWorkspace()
        manifest = workspace.get_link("1", WorkspaceLinkType.JAVA_MANIFEST)
        self.assertIsNone(manifest)

    @patch("bsp_server.scip_sync_util.workspace.os.path.exists")
    @patch("bsp_server.util.utils.get_json")
    def test_get_manifest_for_existing_file(self, mock_get_json, mock_path_exists):
        mock_path_exists.return_value = True
        mock_get_json.return_value = {
            "files": {"file1.java": {"JAVA_MANIFEST": "1"}},
            "links": {"JAVA_MANIFEST": {"1": "manifest1"}},
        }
        manifest = get_manifest_for_file("file1.java", "/path/to/dest")
        self.assertEqual(manifest, "manifest1")

    @patch("bsp_server.scip_sync_util.workspace.os.path.exists")
    @patch("bsp_server.util.utils.get_json")
    def test_get_manifest_for_existing_file_with_roots(
        self, mock_get_json, mock_path_exists
    ):
        mock_path_exists.return_value = True
        mock_get_json.return_value = {
            "files": {
                "file1.java": {
                    "BAZEL_TARGET": "1",
                    "JAVA_MANIFEST": "2",
                    "JAVA_SOURCE_ROOTS": "3",
                }
            },
            "links": {
                "BAZEL_TARGET": {
                    "1": "//experimental/users/hshukla/scip/java-sample:src_main"
                },
                "JAVA_MANIFEST": {
                    "2": "bazel-out/k8-fastbuild/bin/experimental/users/hshukla/scip/java-sample/experimental/users/hshukla/scip/java-sample:src_main_manifest.jar"
                },
                "JAVA_SOURCE_ROOTS": {
                    "3": "experimental/users/hshukla/scip/java-sample/src/main/java"
                },
            },
            "_last_link_id": 3,
        }

        manifest = get_manifest_for_file("file1.java", "/path/to/dest")
        self.assertEqual(
            manifest,
            "bazel-out/k8-fastbuild/bin/experimental/users/hshukla/scip/java-sample/experimental/users/hshukla/scip/java-sample:src_main_manifest.jar",
        )

    @patch("bsp_server.scip_sync_util.workspace.os.path.exists")
    @patch("bsp_server.util.utils.get_json")
    def test_get_manifest_for_non_existing_file(self, mock_get_json, mock_path_exists):
        mock_path_exists.return_value = True
        mock_get_json.return_value = {
            "files": {"file1.java": {"JAVA_MANIFEST": "1"}},
            "links": {"JAVA_MANIFEST": {"1": "manifest1"}},
        }
        manifest = get_manifest_for_file("file2.java", "/path/to/dest")
        self.assertIsNone(manifest)

    @patch("bsp_server.scip_sync_util.workspace.os.path.exists")
    @patch("bsp_server.util.utils.get_string_lines")
    def test_add_to_workspace(self, mock_get_string_lines, mock_path_exists):
        mock_path_exists.return_value = True
        mock_get_string_lines.return_value = ["file1.java", "file2.java"]

        workspace = ScipWorkspace()
        target_mnemonics = {
            ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value: ["manifest1_options"],
            ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value: ["sources_list.txt"],
        }

        add_to_workspace("/cwd", workspace, "//path/to:target", target_mnemonics)

        self.assertEqual(
            workspace.get_file("file1.java"),
            {"BAZEL_TARGET": "1", "JAVA_MANIFEST": "2"},
        )
        self.assertEqual(
            workspace.get_file("file2.java"),
            {"BAZEL_TARGET": "1", "JAVA_MANIFEST": "2"},
        )
        self.assertEqual(
            workspace.get_link("1", WorkspaceLinkType.BAZEL_TARGET), "//path/to:target"
        )
        self.assertEqual(
            workspace.get_link("2", WorkspaceLinkType.JAVA_MANIFEST),
            "manifest1_options",
        )
        mock_get_string_lines.assert_called_once_with("/cwd/sources_list.txt")

    @patch("bsp_server.scip_sync_util.workspace.os.path.exists")
    @patch("bsp_server.scip_sync_util.workspace.add_files_for_target")
    def test_add_to_workspace_no_sources_list(
        self, mock_add_files_for_target, mock_path_exists
    ):
        mock_path_exists.return_value = False

        workspace = ScipWorkspace()
        target_mnemonics = {
            ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value: ["manifest1"],
            ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value: ["sources_list"],
        }

        add_to_workspace("/cwd", workspace, "//path/to:target", target_mnemonics)

        self.assertEqual(workspace.files, {})
        self.assertEqual(workspace.links, {})
        self.assertEqual(workspace._last_link_id, 0)
        mock_add_files_for_target.assert_not_called()

    @patch("bsp_server.scip_sync_util.workspace.os.path.exists")
    @patch("bsp_server.scip_sync_util.workspace.add_files_for_target")
    def test_add_to_workspace_none_sources_list(
        self, mock_add_files_for_target, mock_path_exists
    ):
        mock_path_exists.return_value = False

        workspace = ScipWorkspace()
        target_mnemonics = {
            ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value: ["manifest1"],
            ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value: [None],
        }

        add_to_workspace("/cwd", workspace, "//path/to:target", target_mnemonics)

        self.assertEqual(workspace.files, {})
        self.assertEqual(workspace.links, {})
        self.assertEqual(workspace._last_link_id, 0)
        mock_add_files_for_target.assert_not_called()

    @patch("bsp_server.util.utils.safe_create")
    @patch("bsp_server.util.utils.write_json")
    @patch("bsp_server.scip_sync_util.workspace.os.path.exists")
    def test_write_workspace_new(self, mock_exists, mock_write_json, mock_safe_create):
        # Setup
        workspace = ScipWorkspace()
        target_id = workspace.add_link("//target:one", WorkspaceLinkType.BAZEL_TARGET)
        manifest_id = workspace.add_link("manifest1", WorkspaceLinkType.JAVA_MANIFEST)
        workspace.add_file("file1.java", target_id, WorkspaceLinkType.BAZEL_TARGET)
        workspace.add_file("file1.java", manifest_id, WorkspaceLinkType.JAVA_MANIFEST)

        dest = "/path/to/dest"
        mock_exists.return_value = False

        # Call the function
        write_workspace(workspace, dest)

        # Verify
        mock_safe_create.assert_called_once_with(dest, is_dir=True)
        mock_exists.assert_not_called()  # Should not check if file exists when merge=False
        mock_write_json.assert_called_once()

        # Check that write_json was called with the correct arguments
        args, kwargs = mock_write_json.call_args
        self.assertEqual(args[0], asdict(workspace))
        self.assertEqual(args[1], os.path.join(dest, WORKSPACE_FILE_NAME))
        self.assertTrue(kwargs.get("pretty", False))

    @patch("bsp_server.scip_sync_util.workspace.add_to_workspace")
    @patch("bsp_server.scip_sync_util.workspace.create_workspace")
    def test_populate_workspace_with_regular_targets(
        self, mock_create_workspace, mock_add_to_workspace
    ):
        """Test populating workspace with regular targets"""
        # Setup
        mock_workspace = MagicMock(spec=ScipWorkspace)
        mock_create_workspace.return_value = mock_workspace

        # Target to output mapping
        target_to_output = {
            "//src/main/java/com/example:target1": {
                ScipMnemonics.INDEX_OUTPUT_MNEMONIC.value: ["index1"],
                ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value: ["manifest1"],
                ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value: [
                    self.source_list_path
                ],
            },
            "//src/main/java/com/example:target2": {
                ScipMnemonics.INDEX_OUTPUT_MNEMONIC.value: ["index2"],
                ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value: ["manifest2"],
                ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value: [
                    self.source_list_path
                ],
            },
        }

        # Call the function
        result = populate_workspace(self.cwd, target_to_output)

        # Verify
        mock_create_workspace.assert_called_once_with(self.cwd)
        self.assertEqual(result, mock_workspace)

        # Verify add_to_workspace was called for each target
        self.assertEqual(mock_add_to_workspace.call_count, 2)

        # Check the calls to add_to_workspace
        calls = [
            unittest.mock.call(
                self.cwd,
                mock_workspace,
                "//src/main/java/com/example:target1",
                target_to_output["//src/main/java/com/example:target1"],
            ),
            unittest.mock.call(
                self.cwd,
                mock_workspace,
                "//src/main/java/com/example:target2",
                target_to_output["//src/main/java/com/example:target2"],
            ),
        ]
        mock_add_to_workspace.assert_has_calls(calls, any_order=True)

    @patch("bsp_server.scip_sync_util.workspace.add_to_workspace")
    @patch("bsp_server.scip_sync_util.workspace.create_workspace")
    def test_populate_workspace_with_3rdparty_targets(
        self, mock_create_workspace, mock_add_to_workspace
    ):
        """Test populating workspace with 3rdparty targets (should be skipped)"""
        # Setup
        mock_workspace = MagicMock(spec=ScipWorkspace)
        mock_create_workspace.return_value = mock_workspace

        # Target to output mapping with 3rdparty targets
        target_to_output = {
            "//src/main/java/com/example:target1": {
                ScipMnemonics.INDEX_OUTPUT_MNEMONIC.value: ["index1"],
                ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value: ["manifest1"],
                ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value: [
                    self.source_list_path
                ],
            },
            "//3rdparty/java/guava:guava": {
                ScipMnemonics.INDEX_OUTPUT_MNEMONIC.value: ["index2"],
                ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value: ["manifest2"],
                ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value: [
                    self.source_list_path
                ],
            },
        }

        # Call the function
        result = populate_workspace(self.cwd, target_to_output)

        # Verify
        mock_create_workspace.assert_called_once_with(self.cwd)
        self.assertEqual(result, mock_workspace)

        # Verify add_to_workspace was called only for the non-3rdparty target
        mock_add_to_workspace.assert_called_once_with(
            self.cwd,
            mock_workspace,
            "//src/main/java/com/example:target1",
            target_to_output["//src/main/java/com/example:target1"],
        )

    @patch("bsp_server.scip_sync_util.workspace.add_to_workspace")
    @patch("bsp_server.scip_sync_util.workspace.create_workspace")
    def test_populate_workspace_with_empty_targets(
        self, mock_create_workspace, mock_add_to_workspace
    ):
        """Test populating workspace with empty targets"""
        # Setup
        mock_workspace = MagicMock(spec=ScipWorkspace)
        mock_create_workspace.return_value = mock_workspace

        # Empty target to output mapping
        target_to_output = {}

        # Call the function
        result = populate_workspace(self.cwd, target_to_output)

        # Verify
        mock_create_workspace.assert_called_once_with(self.cwd)
        self.assertEqual(result, mock_workspace)

        # Verify add_to_workspace was not called
        mock_add_to_workspace.assert_not_called()

    @patch("bsp_server.scip_sync_util.workspace.add_to_workspace")
    @patch("bsp_server.scip_sync_util.workspace.create_workspace")
    def test_populate_workspace_with_only_3rdparty_targets(
        self, mock_create_workspace, mock_add_to_workspace
    ):
        """Test populating workspace with only 3rdparty targets (all should be skipped)"""
        # Setup
        mock_workspace = MagicMock(spec=ScipWorkspace)
        mock_create_workspace.return_value = mock_workspace

        # Target to output mapping with only 3rdparty targets
        target_to_output = {
            "//3rdparty/java/guava:guava": {
                ScipMnemonics.INDEX_OUTPUT_MNEMONIC.value: ["index1"],
                ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value: ["manifest1"],
                ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value: [
                    self.source_list_path
                ],
            },
            "//3rdparty/java/jackson:jackson": {
                ScipMnemonics.INDEX_OUTPUT_MNEMONIC.value: ["index2"],
                ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value: ["manifest2"],
                ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value: [
                    self.source_list_path
                ],
            },
        }

        # Call the function
        result = populate_workspace(self.cwd, target_to_output)

        # Verify
        mock_create_workspace.assert_called_once_with(self.cwd)
        self.assertEqual(result, mock_workspace)

        # Verify add_to_workspace was not called for any target
        mock_add_to_workspace.assert_not_called()

    @patch("bsp_server.scip_sync_util.workspace.add_files_for_target")
    @patch("bsp_server.util.utils.get_string_lines")
    @patch("os.path.exists")
    def test_add_to_workspace_integration(
        self, mock_exists, mock_get_string_lines, mock_add_files_for_target
    ):
        """Test the integration between populate_workspace and add_to_workspace"""
        # Setup
        mock_exists.return_value = True
        # Set up get_string_lines to return different values for different calls
        mock_get_string_lines.side_effect = [
            ["file1.java", "file2.java"],
            ["file3.java", "file4.java"],  # Second call for source list
        ]

        # Create a real workspace
        workspace = create_workspace(self.cwd)

        # Target to output mapping
        target_to_output = {
            "//src/main/java/com/example:target1": {
                ScipMnemonics.INDEX_OUTPUT_MNEMONIC.value: ["index1"],
                ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value: [
                    "manifest1_options",
                    "manifest2_options",
                    "manifest3_options",
                    "some_other_template",
                ],
                ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value: [
                    self.source_list_path,
                    self.another_source_list_path,
                    None,
                ],
            }
        }

        # Call add_to_workspace directly
        add_to_workspace(
            self.cwd,
            workspace,
            "//src/main/java/com/example:target1",
            target_to_output["//src/main/java/com/example:target1"],
        )

        # Verify
        mock_exists.assert_has_calls(
            [
                unittest.mock.call(os.path.join(self.cwd, self.source_list_path)),
                unittest.mock.call(
                    os.path.join(self.cwd, self.another_source_list_path)
                ),
            ],
            any_order=True,
        )

        # Verify get_string_lines was called twice with the correct arguments
        expected_calls = [
            unittest.mock.call(os.path.join(self.cwd, self.source_list_path)),
        ]
        mock_get_string_lines.assert_has_calls(expected_calls, any_order=False)
        self.assertEqual(mock_get_string_lines.call_count, 2)

        # Verify add_files_for_target was called with the correct arguments
        mock_add_files_for_target.assert_has_calls(
            [
                unittest.mock.call(
                    workspace=workspace,
                    target="//src/main/java/com/example:target1",
                    files=[
                        "file1.java",
                        "file2.java",
                    ],  # Second return value from get_string_lines
                    manifest="manifest1_options",
                ),
                unittest.mock.call(
                    workspace=workspace,
                    target="//src/main/java/com/example:target1",
                    files=[
                        "file3.java",
                        "file4.java",
                    ],  # Second return value from get_string_lines
                    manifest="manifest2_options",
                ),
            ],
            any_order=True,
        )


if __name__ == "__main__":
    unittest.main()
