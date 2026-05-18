#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${script_dir}/lib/migrate-common.sh"

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <migration_name>" >&2
  echo "Example: $0 create_users" >&2
  exit 1
fi

migrate create -ext sql -dir "${migrations_dir}" -seq "$1"
