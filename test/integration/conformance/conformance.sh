#!/usr/bin/env bash
set -eo pipefail

cd /mini-lab

eval "$(make dev-env)"

cd -

go test -v