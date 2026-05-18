#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/../.." && pwd)"

default_migrations_dir="${repo_root}/backend/internal/migrations"
migrations_dir="${MIGRATIONS_DIR:-${default_migrations_dir}}"

if [[ -z "${DATABASE_URL:-}" && -f "${repo_root}/.env" ]]; then
  database_url_line="$(grep -E '^DATABASE_URL=' "${repo_root}/.env" | tail -n 1 || true)"
  if [[ -n "${database_url_line}" ]]; then
    DATABASE_URL="${database_url_line#DATABASE_URL=}"
    export DATABASE_URL
  fi
fi

if [[ ! -d "${migrations_dir}" ]]; then
  echo "Migration directory does not exist: ${migrations_dir}" >&2
  exit 1
fi

if ! command -v migrate >/dev/null 2>&1; then
  echo "migrate CLI is required." >&2
  echo "Install with: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest" >&2
  exit 1
fi

require_database_url() {
  if [[ -z "${DATABASE_URL:-}" ]]; then
    echo "DATABASE_URL is required." >&2
    echo "Set DATABASE_URL or add it to ${repo_root}/.env." >&2
    exit 1
  fi
}

run_migrate() {
  require_database_url
  migrate -path "${migrations_dir}" -database "${DATABASE_URL}" "$@"
}
