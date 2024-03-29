#!/usr/bin/env bash
set -eo pipefail

cd /mini-lab

pip install --root-user-action=ignore --upgrade pip junit-xml

ansible-playbook -i inventories/control-plane.yaml obtain_role_requirements.yaml
ansible-galaxy install --ignore-errors -r requirements.yaml

cd /integration

ansible-playbook -i /mini-lab/inventories/control-plane.yaml prepare.yaml

rm -f /output/*
export ANSIBLE_CONFIG=/mini-lab/ansible.cfg
export ANSIBLE_CALLBACKS_ENABLED=junit
export JUNIT_OUTPUT_DIR=/output
export JUNIT_FAIL_ON_IGNORE=true

ansible-playbook -i /mini-lab/inventories/control-plane.yaml integration.yaml -v
