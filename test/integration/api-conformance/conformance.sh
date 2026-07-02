#!/usr/bin/env bash
set -eo pipefail

cd "${MINI_LAB_PATH}"/mini-lab

eval $(make --no-print-directory dev-env)

cd -

go test -v
