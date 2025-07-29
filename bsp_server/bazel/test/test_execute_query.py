import json
import os
import shutil
import tempfile
import unittest
from unittest.mock import ANY, patch

from bsp_server.bazel.execute_query import execute_query
from bsp_server.util import utils


class ExecuteQueryTest(unittest.TestCase):
    def setUp(self):
        self.temp_dir = tempfile.mkdtemp()
        self.query_output_file = os.path.join(self.temp_dir, "query_output_file")

    def tearDown(self):
        shutil.rmtree(self.temp_dir)

    @patch("bsp_server.util.utils.stream_output")
    def test_execute_query_success(self, m_stream_output):
        # Given
        targets = [
            "//my/test/target:src_main",
            "//my/test/target:test_main",
            "//my/other/test/target/...",
        ]

        # create the query result
        query_result = [
            {
                "type": "RULE",
                "rule": {
                    "name": "rule_name1",
                    "ruleClass": "java_library",
                    "location": "location1",
                    "attribute": [
                        {
                            "name": "srcs",
                            "type": "LABEL_LIST",
                            "stringListValue": ["src1A", "src1B"],
                        },
                        {
                            "name": "tags",
                            "type": "STRING_LIST",
                            "stringListValue": ["label1A", "label1B"],
                        },
                    ],
                    "ruleInput": ["rule_input1A", "rule_input1B"],
                },
            },
            {
                "type": "RULE",
                "rule": {
                    "name": "rule_name2",
                    "ruleClass": "java_binary",
                    "location": "location2",
                    "attribute": [
                        {
                            "name": "srcs",
                            "type": "LABEL_LIST",
                            "stringListValue": ["src2A", "src2B"],
                        },
                        {
                            "name": "tags",
                            "type": "STRING_LIST",
                            "stringListValue": ["label2A", "label2B"],
                        },
                    ],
                    "ruleInput": ["rule_input2A", "rule_input2B"],
                },
            },
        ]

        def m_stream_output_se(command, cwd, **_kwargs):
            self.assertEqual(
                [
                    "bazel",
                    "query",
                    "--output",
                    "streamed_jsonproto",
                    "--order_output=no",
                    "--query_file",
                    ANY,
                ],
                command,
            )
            self.assertEqual(cwd, self.temp_dir)

            # Verify
            self.assertEqual(
                utils.get_string_content(command[-1]),
                '"//my/test/target:src_main" + "//my/test/target:test_main" + "//my/other/test/target/..."',
            )

            return iter([(0, None, json.dumps(jp)) for jp in query_result])

        m_stream_output.side_effect = m_stream_output_se

        # When
        actual = execute_query(cwd=self.temp_dir, targets=targets)

        self.assertEqual(query_result, actual)


if __name__ == "__main__":
    unittest.main()
