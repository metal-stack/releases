#!/usr/bin/env bash
set -eo pipefail

# this script has the following tasks:
# - install general CI test deps
# - cleanup remainings of prior test runs
# - prepare a CI runner with a fresh version of mini-lab
#
# not intended to be run locally!

MINI_LAB_VERSION=${1:-master}
METAL_STACK_VERSION=${2:-$(git rev-parse --abbrev-ref HEAD)}
MINI_LAB_PATH=./mini-lab

if ! which yq; then
  sudo curl -Lo /usr/local/bin/yq https://github.com/mikefarah/yq/releases/download/3.3.2/yq_linux_amd64
  sudo chmod +x /usr/local/bin/yq
fi

if [ -d "${MINI_LAB_PATH}" ]; then
  # runners are dirty after some time. we should not just delete the mini-lab folder which contains
  # the vagrant directory. let's clean up before cloning the new folder.
  cd ${MINI_LAB_PATH}
  make cleanup
  /test/ci-cleanup.sh
  cd -
fi

rm -rf "${MINI_LAB_PATH}"
git clone $(yq r release.yaml 'projects.metal-stack.mini-lab.repository')

cd "${MINI_LAB_PATH}"

# we usually need to check out the latest mini-lab from master in CI because
# the tagged one could be incompatible with the devel state of
# this repository's release vector
git checkout "${MINI_LAB_VERSION}"

make cleanup
/test/ci-cleanup.sh

cd -
