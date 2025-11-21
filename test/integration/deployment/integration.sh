#!/usr/bin/env bash
set -eo pipefail

cd /mini-lab

eval $(make dev-env)
export ANSIBLE_CONFIG=/mini-lab/ansible.cfg

ansible -m metalstack.base.metal_stack_release_vector localhost -i inventories/control-plane.yaml

mkdir -p ~/.ssh
cat << EOF >> ~/.ssh/config
Host *
    StrictHostKeyChecking no
    User root
    IdentityFile /mini-lab/files/ssh/id_rsa
    PubkeyAcceptedKeyTypes +ssh-rsa
EOF

cd -

# FIXME: Install correct version of metal_python
pip install --root-user-action=ignore --upgrade pip \
  coverage \
  pytest-rerunfailures \
  mock \
  metal_python \
  pytest \
  pytest-testinfra

# if you want to develop tests from within here, comment in the following line:
# bash

pytest -o cache_dir=/tmp/.pytest-cache --junitxml=/output/results_$(date "+%Y.%m.%d-%H.%M.%S").xml
