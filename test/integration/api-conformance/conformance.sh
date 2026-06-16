#!/usr/bin/env bash
set -eo pipefail

cd "${MINI_LAB_PATH}"/mini-lab

eval $(make --no-print-directory dev-env)

# TODO: can be removed after this was added to the mini-lab makefile
export METAL_APIV2_URL=http://v2.api.172.17.0.1.nip.io:8080

cd -

go test -v
