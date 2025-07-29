#!/bin/bash
set -e

# Bazel will be run in the workspace root of this repo.
SYNC_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &> /dev/null && pwd)"
WORKSPACE_ROOT="$(cd "$SYNC_SCRIPT_DIR" && git rev-parse --show-toplevel)"
(
    cd "$WORKSPACE_ROOT"
    bin/bazel build //bsp_server/scip_sync_util:scip_sync
)

# Sync script will be run from the user's current working directory.
"$WORKSPACE_ROOT/bazel-bin/bsp_server/scip_sync_util/scip_sync" "$@"
