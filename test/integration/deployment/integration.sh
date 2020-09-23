#!/usr/bin/env bash
set -eo pipefail

cd /mini-lab

eval $(make dev-env)
export ANSIBLE_CONFIG=/mini-lab/ansible.cfg
export ANSIBLE_VAGRANT_USE_CACHE=1
export ANSIBLE_VAGRANT_CACHE_FILE=/mini-lab/.ansible_vagrant_cache
export ANSIBLE_VAGRANT_CACHE_MAX_AGE=0
ansible-playbook \
  -i inventories/control-plane.yaml \
  obtain_role_requirements.yaml
ansible-galaxy install --ignore-errors -r requirements.yaml

mkdir -p ~/.ssh
cat << EOF >> ~/.ssh/config
Host *
    StrictHostKeyChecking no
EOF

cd -

# FIXME: Install correct version of metal_python
pip install \
  coverage \
  flaky \
  mock \
  metal_python \
  nose \
  testinfra

# if you want to develop tests from within here, comment in the following line:
# bash

nosetests --with-flaky
