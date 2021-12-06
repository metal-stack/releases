#!/usr/bin/env bash
set -eo pipefail

cd /mini-lab

eval $(make dev-env)
export ANSIBLE_CONFIG=/mini-lab/ansible.cfg
ansible-playbook \
  -i inventories/control-plane.yaml \
  obtain_role_requirements.yaml
ansible-galaxy install --ignore-errors -r requirements.yaml

mkdir -p ~/.ssh
cat << EOF >> ~/.ssh/config
Host *
    StrictHostKeyChecking no
    User root
    IdentityFile /mini-lab/files/ssh/id_rsa
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

nosetests --with-flaky --no-flaky-report --with-xunit --xunit-file=/output/results_$(date "+%Y.%m.%d-%H.%M.%S").xml
