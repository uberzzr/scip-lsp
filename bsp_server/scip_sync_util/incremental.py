import os.path

from bsp_server.util import utils

# these are the paths to the binaries that we use to generate the scip index
# they will be generated during the build but for now we will assume they are present
AGGREGATOR = "bazel-bin/src/main/java/com/uber/scip/aggregator/aggregator_bin"
WORK_DIR = ".scip/tmp"


def index_file(cwd: str, file: str, manifest: str):
    file_name = file.replace("/", "_").replace(".", "_")
    scip_file_mutated = os.path.join(cwd, WORK_DIR, file_name + ".scip")
    utils.safe_create(os.path.join(cwd, WORK_DIR), is_dir=True)
    generation_args = [
        os.path.join(cwd, AGGREGATOR),
        "-m",
        manifest,
        "-f",
        file,
        "-o",
        scip_file_mutated,
    ]
    utils.check(generation_args, cwd=os.path.join(cwd))
    return scip_file_mutated
