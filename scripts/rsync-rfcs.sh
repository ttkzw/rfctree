#!/bin/sh
set -eu
DEST_DIR="rfcs"

scripts_dir=$(dirname "$0")
cd "${scripts_dir}/.."
[ -d "${DEST_DIR}" ] || mkdir "${DEST_DIR}"
rsync -avz --delete rsync.rfc-editor.org::rfcs-json-only ${DEST_DIR}/
