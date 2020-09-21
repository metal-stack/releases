#!/usr/bin/env bash
set -eo pipefail

# this script has the following tasks:
# - modify the mini-lab release to deploy either a given metal-stack version or the releases of the current branch of this repository

MINI_LAB_PATH=${1}
if [ -z "$MINI_LAB_PATH" ]; then
  echo "MINI_LAB_PATH needs to be passed as an argument"
fi

METAL_STACK_VERSION=${2:-$(git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD || git rev-parse --short HEAD)}

# use release vector of this repository
yq w -i "${MINI_LAB_PATH}/inventories/group_vars/all/images.yaml" 'metal_stack_release_version' "${METAL_STACK_VERSION}"

echo
echo "Modifying metal-stack release version in mini-lab to '${METAL_STACK_VERSION}':"
echo
echo "# ${MINI_LAB_PATH}/inventories/group_vars/all/images.yaml"
echo ---
cat "${MINI_LAB_PATH}/inventories/group_vars/all/images.yaml"
echo ...
echo
