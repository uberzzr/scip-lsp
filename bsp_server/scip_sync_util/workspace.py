import os
from dataclasses import asdict, dataclass, field
from enum import Enum
from typing import Union

from bsp_server.scip_sync_util.mnemonics import ScipMnemonics
from bsp_server.util import utils

WORKSPACE_FILE_NAME = "workspace.json"


class WorkspaceLinkType(Enum):
    JAVA_MANIFEST = "1"
    BAZEL_TARGET = "2"


@dataclass
class ScipWorkspace:
    # Single file could be related to multiple artifacts. To avoid duplication
    # we will keep only a reference to that artifact.
    # We will refer to it as "link".
    files: dict[str, dict[str, str]] = field(default_factory=dict)
    # This will maintain unique entries for each artifact.
    links: dict[str, dict[str, str]] = field(default_factory=dict)
    # This will be used to generate unique link id
    _last_link_id: int = 0

    def add_file(self, file: str, link_id: int, link_type: WorkspaceLinkType) -> None:
        if file not in self.files:
            self.files[file] = {}
        self.files[file].update({link_type.name: str(link_id)})

    def get_file(self, file: str):
        return self.files.get(file, None)

    def add_link(self, link: str, link_type: WorkspaceLinkType):
        self._last_link_id += 1
        # use str since parsing json will convert int to string
        id = str(self._last_link_id)
        type_links = self.links.get(link_type.name, {})
        type_links.update({id: link})
        self.links[link_type.name] = type_links
        return id

    def get_link(self, link_id: str, link_type: WorkspaceLinkType):
        return self.links.get(link_type.name, {}).get(link_id, None)

    def clear(self):
        self.files = {}
        self.links = {}
        self._last_link_id = 0


def workspace_to_dictionary(
    workspace: ScipWorkspace,
) -> dict[str, dict[str, dict[str, str]]]:
    workspace_dict = {}
    for file in workspace.files:
        target = None
        manifest = None
        for link_type, link_id in workspace.files[file].items():
            if link_type == WorkspaceLinkType.BAZEL_TARGET.name:
                target = workspace.get_link(link_id, WorkspaceLinkType.BAZEL_TARGET)
            if link_type == WorkspaceLinkType.JAVA_MANIFEST.name:
                manifest = workspace.get_link(link_id, WorkspaceLinkType.JAVA_MANIFEST)
        if target:
            workspace_dict.setdefault(target, {}).setdefault(
                ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value, set()
            ).add(file)
            if manifest:
                workspace_dict.setdefault(target, {}).setdefault(
                    ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value, manifest
                )
    return workspace_dict


def create_workspace(cwd: str) -> ScipWorkspace:
    return ScipWorkspace()


def populate_workspace(
    cwd: str, target_to_output: dict[str, dict[str, list[str]]]
) -> ScipWorkspace:
    workspace = create_workspace(cwd)
    for target, target_mnemonics in target_to_output.items():
        # Skip 3rd party targets for workspace
        if target.startswith("//3rdparty/") or target.startswith("bazel-out"):
            continue
        add_to_workspace(cwd, workspace, target, target_mnemonics)
    return workspace


def add_to_workspace(
    cwd: str,
    workspace: ScipWorkspace,
    target: str,
    target_mnemonics: dict[str, list[str]],
) -> None:
    manifests = target_mnemonics.get(
        ScipMnemonics.JAVA_TARGET_MANIFEST_MNEMONIC.value, [None]
    )
    # Filter the list to include only files that end with "_options"
    # bin and test bins are getting picked by mnemonic query, this should work for now
    manifests = [file for file in manifests if file and file.endswith("_options")]
    sources_lists = target_mnemonics.get(
        ScipMnemonics.UNPACKED_JAVA_SOURCES_MNEMONIC.value, [None]
    )
    for index, manifest in enumerate(manifests):
        # for given target we can have only 1 manifest and 1 sources list
        if index >= len(sources_lists):
            # safeguard, should not happen
            continue
        sources_list = sources_lists[index]
        if sources_list is None:
            continue
        if os.path.exists(os.path.join(cwd, sources_list)):
            lines = utils.get_string_lines(os.path.join(cwd, sources_list))
            add_files_for_target(
                workspace=workspace,
                target=target,
                files=lines,
                manifest=manifest,
            )


def add_files_for_target(
    workspace: ScipWorkspace, target: str, files: list[str], manifest: str
) -> None:
    target_id = workspace.add_link(target, WorkspaceLinkType.BAZEL_TARGET)
    manifest_id = workspace.add_link(manifest, WorkspaceLinkType.JAVA_MANIFEST)
    for file in files:
        workspace.add_file(file, target_id, WorkspaceLinkType.BAZEL_TARGET)
        workspace.add_file(file, manifest_id, WorkspaceLinkType.JAVA_MANIFEST)


def write_workspace(workspace: ScipWorkspace, dest: str) -> None:
    utils.safe_create(dest, is_dir=True)
    workspace_dict = asdict(workspace)
    utils.write_json(
        workspace_dict,
        os.path.join(dest, WORKSPACE_FILE_NAME),
        default_serializer=utils.set_to_list,
        pretty=True,
    )


def get_manifest_for_file(file: str, dest: str) -> Union[str, None]:
    if os.path.exists(os.path.join(dest, WORKSPACE_FILE_NAME)):
        json_obj = utils.get_json(os.path.join(dest, WORKSPACE_FILE_NAME))
        workspace = ScipWorkspace(**json_obj)
        # update workspace with new data
        file_links = workspace.get_file(file)
        if file_links:
            manifest_link = file_links.get(WorkspaceLinkType.JAVA_MANIFEST.name, None)
            # we are expecting only 1 manifest for a file
            return workspace.get_link(manifest_link, WorkspaceLinkType.JAVA_MANIFEST)
    return None
