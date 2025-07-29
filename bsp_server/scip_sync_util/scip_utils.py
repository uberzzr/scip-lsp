import hashlib
import json
import multiprocessing
import os
import os.path
import re
import shutil
import tempfile
from concurrent.futures import ThreadPoolExecutor, as_completed
from functools import lru_cache

from bsp_server.scip_sync_util import scip_const
from bsp_server.util import utils


def parse_bazelproject(file_path: str) -> dict[str, list[str]]:
    """Parse a .bazelproject file and return a dictionary of the contents."""
    data = {}
    last_key = ""

    with open(file_path, "r") as file:
        for line in file:
            line = line.strip()
            # Skip empty lines and comments
            if not line or line.startswith("#"):
                continue

            key, value = _get_key_value(line)

            if key != "":
                data[key] = []
                last_key = key

            if value != "":
                data[last_key].append(value)

    return data


def _get_key_value(line: str) -> (str, str):
    line = line.strip()
    key = ""
    value = line

    if line.startswith("//"):
        return key, value

    if ":" in line:
        k, _, v = line.partition(":")
        key = k.strip()
        value = v.strip()

    return key, value


def get_containing_bazel_target(cwd: str, filepath: str, query_kinds: list[str]) -> str:
    from_target = "//" + filepath.rpartition("/src")[0] + "/..."
    query_string = (
        f'kind("{"|".join(query_kinds)}", rdeps("{from_target}", "{filepath}"))'
    )
    cmd = [
        "bazel",
        "query",
        query_string,
    ]
    return utils.output(cmd, cwd=cwd)


def old_copy_index(index_to_copy: set[str], dest: str) -> None:
    utils.safe_create(dest, is_dir=True)
    for index in index_to_copy:
        parts = index.split(os.path.sep + "bin" + os.path.sep)
        idx = parts[-1]
        index_filename = idx.replace("/", "_").replace("-", "_")

        dest_path = os.path.join(dest, index_filename)
        shutil.copy(index, dest_path)

        sha_filename = index_filename + ".sha256"
        with open(os.path.join(dest, sha_filename), "w") as f:
            f.write(generate_sha256(index) + "\n")


def generate_sha256(file_path: str) -> str:
    sha256 = hashlib.sha256()
    with open(file_path, "rb") as f:
        for byte_block in iter(lambda: f.read(4096), b""):
            sha256.update(byte_block)
    return sha256.hexdigest()


def get_sha256_for_file(file_path: str) -> str:
    try:
        with open(file_path, "r") as f:
            return f.read().strip()
    except FileNotFoundError:
        return None


def get_mnemonic_output(cwd, mnemonic, targets):
    """Verify mnemonic output using a query file."""
    union = " + ".join(f'"{target}"' for target in targets)
    query = f'mnemonic("{mnemonic}", {union})'

    # Create a temporary file for the query
    with tempfile.NamedTemporaryFile(
        mode="w", suffix=".txt", delete=False
    ) as query_file:
        query_file.write(query)
        query_file_path = query_file.name

    try:
        aquery_cmd = [
            "bazel",
            "aquery",
            "--query_file",
            query_file_path,
            "--aspects",
            scip_const.ASPECT_SCIP_INDEX,
            scip_const.ASPECT_OUTPUT_GROUPS,
            "--output=jsonproto",
            "--keep_going",
        ]
        action_out_json = utils.output(command=aquery_cmd, cwd=cwd)
        print(f"Processing action output...")
        return _get_all_outputs(action_out_json)
    except Exception as e:
        return {}


def _get_all_outputs(json_data):
    """Get all outputs from the action json data."""
    data = json.loads(json_data)
    path_fragments = []
    if "pathFragments" in data:
        path_fragments = data["pathFragments"]
    artifacts = {}
    if "artifacts" in data:
        artifacts = data["artifacts"]
    actions = {}
    if "actions" in data:
        actions = data["actions"]
    targets = []
    if "targets" in data:
        targets = data["targets"]

    # Calculate thread pool size as half of available CPUs
    max_workers = get_thread_pool_size()

    # Create a manager for thread-safe shared objects
    manager = multiprocessing.Manager()

    # Pre-process path fragments to create a lookup dictionary for parent fragments
    # This avoids repeated searches in the process_fragment function
    parent_lookup = {}
    for fragment in path_fragments:
        parent_id = fragment.get("parentId")
        if parent_id:
            parent_lookup[fragment["id"]] = parent_id

    # Create a dictionary to map fragment IDs to their labels for quick lookup
    fragment_labels = {fragment["id"]: fragment["label"] for fragment in path_fragments}

    # Create a thread-safe dictionary to map pathFragmentId to the full path
    path_dict = manager.dict()

    with ThreadPoolExecutor(max_workers=max_workers) as executor:

        def process_fragment(fragment):
            """Process a single path fragment to build its full path."""
            fragment_id = fragment["id"]

            # Check if we've already processed this fragment
            if fragment_id in path_dict:
                return None

            # Build the full path by traversing parent IDs
            path_parts = []
            current_id = fragment_id

            # Collect all path parts by traversing up the parent chain
            while current_id:
                path_parts.append(fragment_labels[current_id])
                current_id = parent_lookup.get(current_id)

            # Combine path parts in reverse order (from root to leaf)
            full_path = "/".join(reversed(path_parts))

            return fragment_id, full_path

        # Submit all fragments for processing
        future_to_fragment = {
            executor.submit(process_fragment, fragment): fragment
            for fragment in path_fragments
        }

        # Collect results as they complete
        for future in as_completed(future_to_fragment):
            result = future.result()
            if result:  # Skip None results (already processed fragments)
                fragment_id, path = result
                path_dict[fragment_id] = path

    # Create a dictionary to map artifactId to pathFragmentId
    artifact_dict = {
        artifact["id"]: artifact["pathFragmentId"] for artifact in artifacts
    }

    # Optimize the target output dictionary creation
    # Group actions by target ID to reduce dictionary updates
    target_output_dict = {}
    for action in actions:
        target_id = action["targetId"]
        if target_id not in target_output_dict:
            target_output_dict[target_id] = {}
        if action["mnemonic"] not in target_output_dict[target_id]:
            target_output_dict[target_id][action["mnemonic"]] = []
        target_output_dict[target_id][action["mnemonic"]].extend(action["outputIds"])

    # Create a thread-safe dictionary for the final output
    target_output_paths = manager.dict()

    # Batch targets for processing to reduce thread overhead
    batch_size = max(1, len(targets) // (max_workers * 2))
    target_batches = [
        targets[i : i + batch_size] for i in range(0, len(targets), batch_size)
    ]

    def process_target_batch(target_batch):
        """Process a batch of targets to build their output paths."""
        batch_results = {}

        for target in target_batch:
            target_id = target["id"]
            target_label = target["label"]

            mnemonic_to_output_ids = target_output_dict.get(target_id, {})
            target_results = {}

            for mnemonic, output_ids in mnemonic_to_output_ids.items():
                # Use list comprehension with pre-filtering to improve performance
                valid_output_ids = [oid for oid in output_ids if oid in artifact_dict]
                output_paths = [
                    path_dict[artifact_dict[oid]] for oid in valid_output_ids
                ]

                if output_paths:  # Only add non-empty results
                    target_results[mnemonic] = output_paths

            if target_results:  # Only add targets with results
                batch_results[target_label] = target_results

        return batch_results

    # Process target batches in parallel
    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        batch_results = list(executor.map(process_target_batch, target_batches))

    # Merge all batch results into the final dictionary
    final_results = {}
    for batch_result in batch_results:
        for target_label, mnemonics in batch_result.items():
            if target_label not in final_results:
                final_results[target_label] = {}
            for mnemonic, paths in mnemonics.items():
                if mnemonic not in final_results[target_label]:
                    final_results[target_label][mnemonic] = []
                final_results[target_label][mnemonic].extend(paths)

    return final_results


@lru_cache(maxsize=1)
def get_thread_pool_size() -> int:
    """Cache the thread pool size calculation."""
    return max(1, multiprocessing.cpu_count() // 2)


def copy_index(index_to_copy: set[str], dest: str) -> None:
    utils.safe_create(dest, is_dir=True)

    def process_and_copy_scip_index(
        source_path: str, current_status: dict
    ) -> tuple[str, str]:
        try:
            # Process index info
            source_sha_path = source_path + scip_const.SHA256_FILE_SUFFIX
            relative_path = source_path.split(os.path.sep + "bin" + os.path.sep)[-1]
            scip_index_name = relative_path.replace("/", "_").replace("-", "_")
            new_sha = get_sha256_for_file(source_sha_path)

            # Check if copy needed
            if (
                scip_index_name in current_status
                and current_status[scip_index_name] == new_sha
            ):
                return (
                    scip_index_name,
                    scip_index_name + scip_const.SHA256_FILE_SUFFIX,
                )

            # Copy files if needed
            dest_index_path = os.path.join(dest, scip_index_name)
            shutil.copy(source_path, dest_index_path)
            shutil.copy(
                source_sha_path, dest_index_path + scip_const.SHA256_FILE_SUFFIX
            )
            return (scip_index_name, scip_index_name + scip_const.SHA256_FILE_SUFFIX)
        except Exception as e:
            print(f"Failed to process index {source_path}: {str(e)}")
            return None

    def get_current_status(filename: str) -> tuple[str, str]:
        if not filename.endswith(".scip"):
            return None, None
        sha = get_sha256_for_file(
            os.path.join(dest, filename + scip_const.SHA256_FILE_SUFFIX)
        )
        return (filename, sha) if sha else (None, None)

    with ThreadPoolExecutor(max_workers=get_thread_pool_size()) as executor:
        # Get current status
        current_status = dict(
            filter(
                None,
                executor.map(
                    get_current_status,
                    [f for f in os.listdir(dest) if f.endswith(".scip")],
                ),
            )
        )

        # Process and copy files
        copy_results = list(
            filter(
                None,
                executor.map(
                    lambda src: process_and_copy_scip_index(src, current_status),
                    index_to_copy,
                ),
            )
        )

        # Delete old files
        files_to_keep = {name for pair in copy_results for name in pair}
        files_to_delete = (
            set(os.listdir(dest)) - files_to_keep - {scip_const.WORKSPACE_FILE_NAME}
        )
        files_to_delete = {
            f
            for f in files_to_delete
            if not f.startswith(scip_const.JDK_SCIP_FILE_PREFIX)
        }

        if files_to_delete:
            list(
                executor.map(
                    lambda f: (
                        os.remove(os.path.join(dest, f))
                        if os.path.isfile(os.path.join(dest, f))
                        else shutil.rmtree(os.path.join(dest, f))
                    ),
                    files_to_delete,
                )
            )


def transform_bazel_query_results(qr: list[dict]) -> dict[str, dict[str, list[str]]]:
    """Transform the results of a bazel query into a dictionary of dependencies."""
    res = {}

    rules = set([])

    for target in qr:
        if target["type"] == "RULE":
            rules.add(target["rule"]["name"])

    for target in qr:
        if target["type"] != "RULE":
            continue

        rule = target["rule"]
        name = rule["name"]
        base_path = rule["name"].split(":")[0][2:]

        # add all rule inputs except external repository
        # to deps
        direct_deps = []

        for dep in rule.get("ruleInput", []):
            if dep.startswith("@"):
                continue
            if dep not in rules:
                continue

            direct_deps.append(dep)

        target_type = rule["ruleClass"]

        e_deps = []
        deps = []
        for attr in rule.get("attribute", []):
            if "stringListValue" not in attr:
                continue
            if attr["name"] == "deps":
                deps += [
                    dep
                    for dep in attr.get("stringListValue", [])
                    if not dep.startswith("@") and dep in rules
                ]

            if attr["name"] == "data":
                deps += [
                    data
                    for data in attr.get("stringListValue", [])
                    if not data.startswith("@") and data in rules
                ]

            if attr["name"] == "exports":
                e_deps = attr.get("stringListValue", [])

        info = {
            "base_path": base_path,
            "deps": list(set(deps)),
            "direct_deps": direct_deps,
            "exports": e_deps,
            "target_type": target_type,
        }

        res[name] = info

    return res


def filter_list_by_regex(list_to_filter: set[str], regex_set: set[str]) -> set[str]:
    """Filters a list based on regex patterns from another list."""

    filtered_list = set()
    for regex_pattern in regex_set:
        for item in list_to_filter:
            if re.search(regex_pattern, item):
                filtered_list.add(item)

    return filtered_list


def dfs(
    dep_graph: dict[str, dict[str, list[str]]], target: str, depth: int
) -> list[str]:
    """
    Run a depth-first search on a dependency graph. Returns a list of targets.
    Will stop at the specified depth. Depth is ignored if the target is exported.

    :param dep_graph : A dictionary of dependencies.
    :param target: The target to start the search from.
    :param depth: The depth to search to.
    :return: A list of targets.
    """
    result = []
    if target not in dep_graph:
        return result

    for exported_dep in dep_graph[target]["exports"]:
        result.append(exported_dep)
        result.extend(dfs(dep_graph, exported_dep, depth - 1))

    if depth < 0:
        return result

    result.extend([target])

    if depth == 0:
        return result

    for dep in dep_graph[target]["direct_deps"]:
        result.append(dep)
        result.extend(dfs(dep_graph, dep, depth - 1))

    return result
