#!/usr/bin/env bash
set -eo pipefail

MINI_LAB_VERSION=${1:-master}
METAL_STACK_VERSION=${2:-$(git rev-parse --abbrev-ref HEAD)}

sudo curl -Lo /usr/local/bin/yq https://github.com/mikefarah/yq/releases/download/3.3.2/yq_linux_amd64
sudo chmod +x /usr/local/bin/yq

rm -rf mini-lab
git clone $(yq r release.yaml 'projects.metal-stack.mini-lab.repository')
cd mini-lab
# we need to check out the latest mini-lab from master in CI because the tagged one could
# be incompatible with the devel state of this repository's release vector
git checkout "${MINI_LAB_VERSION}"

make ci-prep

cd -
