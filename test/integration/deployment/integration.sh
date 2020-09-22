#!/usr/bin/env bash
set -eo pipefail

cd /mini-lab
eval $(make dev-env)
cd -

# FIXME: metal_python must be installed in the proper version
pip install --upgrade pip nose coverage mock metal_python flaky

nosetests --with-flaky
