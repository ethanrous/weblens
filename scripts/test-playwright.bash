#!/bin/bash
set -euo pipefail

source ./scripts/lib/all.bash

MONGO_PORT=27020
MONGO_STACK_NAME="playwright-test"

lazy=true
filter=""
grep=""
headed=

while [ "${1:-}" != "" ]; do
    case "$1" in
    "--no-lazy")
        lazy=false
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
        ;;
    "-h" | "--help")
        echo "Usage: $0 [--no-lazy]"
        echo "  --no-lazy    Force rebuild everything"
        exit 0
        ;;
    esac
    shift
done

show_as_subtask "Launching MongoDB for Playwright tests..." "green" -- launch_mongo --stack-name "${MONGO_STACK_NAME}" --mongo-port "${MONGO_PORT}"

export WEBLENS_MONGODB_URI="mongodb://127.0.0.1:${MONGO_PORT}/?replicaSet=rs0&directConnection=true"
printf "Using MongoDB URI: %s\n" "$WEBLENS_MONGODB_URI"

# Build Agno if needed
if [[ "$lazy" = false ]] || ! does_agno_exist; then
    build_agno
else
    printf "Skipping Agno build (lazy mode)...\n"
fi

# ENABLE_SOURCEMAPS=true
build_frontend false

# Build Go binary
if [[ "$lazy" = false ]] || [[ ! -e "$WEBLENS_ROOT/_build/bin/weblens_debug" ]]; then
    show_as_subtask "Building Go binary..." "green" -- build_weblens_binary
else
    printf "Skipping Go binary build (lazy mode)...\n"
fi

rm -f ./_build/playwright/report/coverage/index.html >/dev/null || true

# Install Playwright browsers if needed
pushd "${WEBLENS_ROOT}/weblens-vue/weblens-nuxt" >/dev/null

show_as_subtask "Installing Playwright browsers..." "green" -- pnpm exec playwright install chromium

PLAYWRIGHT_LOG_PATH=$(get_log_file "weblens-playwright-test")

# Run Playwright tests
echo "Running Playwright tests... (logs will be saved to $PLAYWRIGHT_LOG_PATH)"
if [[ $headed ]]; then
    echo "Running in headed mode (browser UI will be visible)..."
fi

export WEBLENS_VERBOSE=true
if ! show_as_subtask "Running Playwright tests..." "green" -- bash -c "set -o pipefail; pnpm exec playwright test \"${filter}\" ${grep} \"${headed:-}\" | tee \"$PLAYWRIGHT_LOG_PATH\""; then
    echo "Playwright tests failed. Check logs for details."
    popd >/dev/null

    echo "Playwright test logs saved to: .${PLAYWRIGHT_LOG_PATH#"${WEBLENS_ROOT}"}"

    exit 1
else
    echo "Playwright tests passed successfully."
    exit 0
fi
