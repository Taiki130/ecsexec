#!/bin/bash
set -euCo pipefail

cd "$(dirname "$0")/../.."

export LOG_LEVEL=debug
export RENOVATE_CONFIG_FILE=renovate.json5

renovate-config-validator
renovate --token="$(gh auth token)" \
    --require-config=ignored \
    --dry-run=full \
    "$(gh repo view --json nameWithOwner -q ".nameWithOwner")"
