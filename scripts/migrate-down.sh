#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${script_dir}/lib/migrate-common.sh"

steps="${1:-1}"

if [[ ! "${steps}" =~ ^[0-9]+$ || "${steps}" -eq 0 ]]; then
  echo "Usage: $0 [positive_step_count]" >&2
  echo "Example: $0 1" >&2
  exit 1
fi

run_migrate down "${steps}"
