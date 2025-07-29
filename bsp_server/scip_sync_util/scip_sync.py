import argparse
import os
import tempfile
from dataclasses import asdict, dataclass
from datetime import datetime
from pprint import pprint
from typing import Tuple

from bsp_server.bazel.execute_query import execute_query
from bsp_server.scip_sync_util import incremental, scip_utils
from bsp_server.scip_sync_util import workspace as scip_workspace
from bsp_server.scip_sync_util.mnemonics import ScipMnemonics
from bsp_server.scip_sync_util.scip_const import (
    ASPECT_OUTPUT_GROUPS,
    ASPECT_SCIP_INDEX,
    BAZEL,
    BUILD,
    DERIVE_TARGETS_FROM_DIRECTORIES,
    DIRECTORIES,
    JAVA_VERSION_FLAGS,
    SCIP_TOOLING_TARGET,
    SUPPORTED_RULES,
    TARGETS,
)
from bsp_server.util import utils

enable_scip_env = {"ENABLE_SCIP_INDEX_GEN": "true"}
SCIP_INDEX_DIR = ".scip"


@dataclass
class SyncStats:
    index_target_extraction_time_sec: float = 0
    index_sync_time_sec: float = 0
    total_duration_sec: float = 0
    copy_index_time_sec: float = 0
    total_scip_targets_identified: int = 0
    total_will_build_scip_target: int = 0
    passed_index_cnt: int = 0
    failed_index_cnt: int = 0


def main():
    parser = argparse.ArgumentParser(description="Sync scip index for given targets")
    parser.add_argument(
        "--cwd",
        type=str,
        required=True,
        help="current working dir where to execute bazel command",
    )

    parser.add_argument(
        "targets",
        type=str,
        nargs="*",
        help="targets to generate scip index",
    )

    parser.add_argument(
        "--filepath",
        type=str,
        help="files to generate scip index for, the target which includes this file is indexed",
    )

    parser.add_argument(
        "--depth",
        default=1,
        type=int,
        help="Dependency graph depth for the index generation",
    )

    sync_stats = SyncStats()

    index_start = datetime.now()

    args = parser.parse_args()
    targets = args.targets
    cwd = args.cwd
    filepath = args.filepath
    excludes = set()

    # Case 1: No initial files or targets - rewrite the workspace
    if not targets and not filepath:
        targets, excludes = fetch_targets_from_bazelproject(cwd)

    # Case 2: Filepath provided - check if it's in workspace
    if filepath:
        file_manifest = None
        try:
            # Remove cwd prefix from filepath for workspace lookup
            relative_filepath = filepath.replace(cwd + "/", "")
            file_manifest = scip_workspace.get_manifest_for_file(
                relative_filepath, os.path.join(cwd, SCIP_INDEX_DIR)
            )

            # If file is already in workspace, no need to update
            if file_manifest:
                file_index = incremental.index_file(
                    cwd, relative_filepath, file_manifest
                )
                scip_utils.old_copy_index(
                    {file_index}, os.path.join(cwd, SCIP_INDEX_DIR)
                )
                return
            else:
                print(f"File {relative_filepath} not found in workspace")
                return
        except Exception as e:
            print(f"Error checking file in workspace: {e}")
            return

        # File not in workspace or error occurred, add its target to the list
        file_target = scip_utils.get_containing_bazel_target(
            cwd, filepath, SUPPORTED_RULES
        )
        if file_target:
            targets.append(file_target)
        else:
            print(f"Could not find target for {filepath}")
            return

    if len(targets) == 0:
        print("No targets to sync ...")
        return

    # Get all deps
    print(f"Syncing deps for targets: {list(targets)}")
    buildable_scip_targets = get_dependency_graph(
        cwd=cwd, targets=targets, depth=args.depth, exclude_targets=excludes
    )

    if len(buildable_scip_targets) == 0:
        print("Found no targets to sync ...")
        return

    # Update stats
    sync_stats.total_scip_targets_identified = len(buildable_scip_targets)
    sync_stats.total_will_build_scip_target = len(buildable_scip_targets)
    sync_stats.index_target_extraction_time_sec = (
        datetime.now() - index_start
    ).total_seconds()

    # Generate SCIP indexes
    sync_scip(cwd, buildable_scip_targets, sync_stats)

    # Process target outputs and update workspace
    target_to_output = scip_utils.get_mnemonic_output(
        cwd,
        f"{ScipMnemonics.INDEX_OUTPUT_MNEMONIC.value}|{ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value}|{ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value}",
        buildable_scip_targets,
    )

    # Collect indexes and update workspace
    index_target_map = {}
    for target, target_mnemonics in target_to_output.items():
        # Collect indexes
        index_list = target_mnemonics.get(ScipMnemonics.INDEX_OUTPUT_MNEMONIC.value, [])
        for index in index_list:
            if os.path.exists(os.path.join(cwd, index)):
                index_target_map.setdefault(target, []).append(os.path.join(cwd, index))

    workspace = scip_workspace.populate_workspace(cwd, target_to_output)
    scip_workspace.write_workspace(workspace, os.path.join(cwd, SCIP_INDEX_DIR))

    # Calculate stats and copy indexes
    relevant_index = list(
        set(index_target_map.keys()).intersection(set(buildable_scip_targets))
    )
    sync_stats.passed_index_cnt = len(relevant_index)
    sync_stats.failed_index_cnt = (
        sync_stats.total_will_build_scip_target - sync_stats.passed_index_cnt
    )

    # Copy indexes to SCIP directory
    index_to_copy = set()
    for idx in relevant_index:
        index_to_copy.update(index_target_map[idx])

    start_copy_index = datetime.now()
    scip_utils.copy_index(index_to_copy, os.path.join(cwd, SCIP_INDEX_DIR))
    sync_stats.copy_index_time_sec = (datetime.now() - start_copy_index).total_seconds()

    print(f"--- Sync Stats ---")
    sync_stats.total_duration_sec = (datetime.now() - index_start).total_seconds()
    pprint(asdict(sync_stats), indent=2, sort_dicts=False)


# Generate the scip index
def sync_scip(cwd: str, targets: set[str], stats: SyncStats) -> None:
    start = datetime.now()
    try:
        # Build tooling needed for the incremental flow
        cmd = [
            BAZEL,
            BUILD,
            SCIP_TOOLING_TARGET,
            "--keep_going",
        ] + JAVA_VERSION_FLAGS
        utils.output(cmd, cwd=cwd)
        targets_file = tempfile.NamedTemporaryFile(delete=False).name
        utils.write_list(targets, targets_file)
        cmd = [
            BAZEL,
            BUILD,
            "--target_pattern_file=" + targets_file,
            "--keep_going",
            "--aspects",
            ASPECT_SCIP_INDEX,
            ASPECT_OUTPUT_GROUPS,
        ] + JAVA_VERSION_FLAGS
        print("--- Initiating sync ----")
        utils.output(cmd, cwd=cwd)
    except Exception:
        # since we have many failing scip targets no need to panic
        pass
    finally:
        end = datetime.now()
        sync_time = (end - start).total_seconds()
        stats.index_sync_time_sec = sync_time
        print(f"--- Completed sync in {sync_time:.2f}s ---")


def get_dependency_graph(
    cwd: str,
    targets: set[str],
    depth: int,
    exclude_targets=None,
    query_kinds: list = SUPPORTED_RULES,
    query_rdeps: bool = False,
    query_deps: bool = True,
    query_rdeps_universe="//...",
) -> set[str]:
    if exclude_targets is None:
        exclude_targets = set()
    query_result = execute_query(
        cwd=cwd,
        targets=list(targets),
        query_kinds=query_kinds,
        query_deps=query_deps,
        query_rdeps=query_rdeps,
        query_depth=depth,
        query_rdeps_universe=query_rdeps_universe,
        soft_fail=True,
    )
    dep_graph = scip_utils.transform_bazel_query_results(query_result)
    masks = set()
    exclude_masks = set()
    sanitized_targets = set()
    # If target contains /... we need to replace it with specific target
    for target in targets:
        if target.endswith("/..."):
            masks.add("^" + target.replace("/...", ""))
        else:
            masks.add("^" + target + "$")

    for exclude in exclude_targets:
        if exclude.endswith("/..."):
            exclude_masks.add("^" + exclude.replace("/...", ""))
        else:
            exclude_masks.add("^" + exclude + "$")

    result = set()

    # in case of the rdeps we need to check dep, not target, target is the wanted result
    if query_rdeps:
        for key in dep_graph.keys():
            if scip_utils.filter_list_by_regex(
                set(dep_graph[key]["direct_deps"]), masks
            ) and not scip_utils.filter_list_by_regex(
                set(dep_graph[key]["direct_deps"]), exclude_masks
            ):
                result.add(key)
    else:
        sanitized_targets.update(
            scip_utils.filter_list_by_regex(set(dep_graph.keys()), masks)
        )
        for target in sanitized_targets:
            # run DFS on the dep_graph to get all the targets with limit depth, ignore depth if target is exported
            result.update(scip_utils.dfs(dep_graph, target, depth))
        result = result.difference(
            scip_utils.filter_list_by_regex(result, exclude_masks)
        )
    return result


def fetch_targets_from_bazelproject(cwd: str) -> Tuple[set[str], set[str]]:
    print("Reading targets from .bazelproject file...")

    bazel_project_file = os.path.join(cwd, ".ijwb", ".bazelproject")
    if not os.path.exists(bazel_project_file):
        print(".bazelproject file not found")
        return set(), set()

    targets = set()
    excludes = set()
    try:
        sections = scip_utils.parse_bazelproject(bazel_project_file)
        if TARGETS in sections:
            for t in sections[TARGETS]:
                if _should_exclude_path(t):
                    # remove - from start
                    excludes.add(t[1:])
                else:
                    targets.add(t)

        derive_from_dirs = (
            sections.get(DERIVE_TARGETS_FROM_DIRECTORIES, ["false"])[0].lower()
            == "true"
        )

        if derive_from_dirs and DIRECTORIES in sections:
            non_excluded_dirs = []
            excluded_dirs = []
            for d in sections[DIRECTORIES]:
                if _should_exclude_path(d):
                    # remove - from start
                    excluded_dirs.append(d[1:])
                else:
                    non_excluded_dirs.append(d)

            print(f"Loading targets from directories: {non_excluded_dirs}")
            if non_excluded_dirs:
                targets = targets | set(
                    convert_directories_to_targets(non_excluded_dirs)
                )
            if excluded_dirs:
                excludes = excludes | set(convert_directories_to_targets(excluded_dirs))

        return targets, excludes

    except Exception as e:
        print(f"Error fetching targets from bazelproject: {e}")
        return set(), set()


def _should_exclude_path(path: str) -> bool:
    return path.startswith("-")


def convert_directories_to_targets(directories: list[str]) -> list[str]:
    targets = []
    for directory in directories:
        directory = directory.strip().rstrip("/")
        # skip base directory
        if "." == directory:
            continue
        targets.append(f"//{directory}/...")
    return targets


if __name__ == "__main__":
    main()
