#!/bin/bash
set -euo pipefail

source ./scripts/lib/all.bash

lazy=false
filter=""
grep=""
headed=

while [ "${1:-}" != "" ]; do
    case "$1" in
    "--lazy")
        lazy=true
        ;;
    "--filter")
        shift
        filter="$1"
        ;;
    "--grep")
        shift
        grep="--grep '$1'"
        ;;
    "--headed")
        headed="--headed"
        export PW_WORKERS=1
        ;;
    "-v")
        export WEBLENS_VERBOSE=true
        ;;
    "-h" | "--help")
        echo "Usage: $0 [--no-lazy]"
        echo "  --no-lazy    Force rebuild everything"
        exit 0
        ;;
    esac
    shift
done

# Build Agno if needed
if ! does_agno_exist; then
    build_agno
else
    printf "Skipping Agno build (lazy mode)...\n"
fi

if ! is_mongo_running --stack-name "test-pw"; then
    show_as_subtask "Launching mongo..." "green" -- launch_mongo --stack-name "test-pw" --mongo-port 27020
else
    printf "MongoDB container is already running...\n"
fi

# Clean up any existing test databases in mongo
show_as_subtask "Dropping existing mongo DBs" "green" -- docker exec weblens-test-pw-mongo-mongod mongosh --eval 'db.adminCommand({"listDatabases": 1, filter: { "name": /^pw-/ }}).databases.forEach(d => db.getSiblingDB(d.name).dropDatabase())'

export VITE_DEBUG_BUILD=true
build_frontend "$lazy"

# Build Go binary
if [[ "$lazy" = false ]] || [[ ! -e "$WEBLENS_ROOT/_build/bin/weblens_debug" ]]; then
    show_as_subtask "Building Go binary..." --color "green" -- build_weblens_binary
else
    printf "Skipping Go binary build (lazy mode)...\n"
fi

rm -rf ./_build/playwright/report/coverage/ >/dev/null || true

# Install Playwright browsers if needed
pushd "${WEBLENS_ROOT}/weblens-vue/weblens-nuxt" >/dev/null

show_as_subtask "Installing Playwright browsers..." --color "green" -- pnpm exec playwright install chromium

PLAYWRIGHT_LOG_PATH=$(get_log_file --prefix "weblens-playwright-test" --subdir "playwright")

# Run Playwright tests
echo "Running Playwright tests... (logs will be saved to $PLAYWRIGHT_LOG_PATH)"
if [[ $headed ]]; then
    echo "Running in headed mode (browser UI will be visible)..."
fi

if ! show_as_subtask "Running Playwright tests..." -v --color "green" -- bash -c "set -o pipefail; NODE_OPTIONS=--max-old-space-size=8192 PW_WORKERS=${PW_WORKERS:-} pnpm exec playwright test \"${filter}\" ${grep} \"${headed:-}\" | tee \"$PLAYWRIGHT_LOG_PATH\""; then
    echo "Playwright tests failed. Check logs for details."
    popd >/dev/null

    echo "Playwright test logs saved to: .${PLAYWRIGHT_LOG_PATH#"${WEBLENS_ROOT}"}"

    exit 1
else
    echo "Playwright tests passed successfully."
    exit 0
fi
