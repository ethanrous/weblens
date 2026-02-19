#!/bin/bash
set -euo pipefail

source ./scripts/lib/all.bash

# Clean up any orphaned pw-worker docker stacks on exit
cleanup_pw_stacks() {
    echo "Cleaning up playwright worker stacks..."
    for stack in $(dockerc compose ls --format json 2>/dev/null | grep -o '"weblens-core-pw-worker-[0-9]*"' | tr -d '"' || true); do
        echo "Stopping orphaned stack: $stack"
        dockerc compose --project-name "$stack" down 2>/dev/null || true
    done
}
trap cleanup_pw_stacks EXIT

# Remove legacy shared playwright-test mongo stack if it exists (old test infra)
if dockerc ps -a --format '{{.Names}}' 2>/dev/null | grep -q 'weblens-playwright-test'; then
    echo "Removing legacy playwright-test mongo stack..."
    dockerc compose --project-name playwright-test down 2>/dev/null || true
fi

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

# ENABLE_SOURCEMAPS=true
build_frontend "$lazy"

# Build Go binary
if [[ "$lazy" = false ]] || [[ ! -e "$WEBLENS_ROOT/_build/bin/weblens_debug" ]]; then
    show_as_subtask "Building Go binary..." --color "green" -- build_weblens_binary
else
    printf "Skipping Go binary build (lazy mode)...\n"
fi

rm -f ./_build/playwright/report/coverage/index.html >/dev/null || true

# Install Playwright browsers if needed
pushd "${WEBLENS_ROOT}/weblens-vue/weblens-nuxt" >/dev/null

show_as_subtask "Installing Playwright browsers..." --color "green" -- pnpm exec playwright install chromium

PLAYWRIGHT_LOG_PATH=$(get_log_file "weblens-playwright-test")

# Run Playwright tests
echo "Running Playwright tests... (logs will be saved to $PLAYWRIGHT_LOG_PATH)"
if [[ $headed ]]; then
    echo "Running in headed mode (browser UI will be visible)..."
fi

if ! show_as_subtask "Running Playwright tests..." -v --color "green" -- bash -c "set -o pipefail; PW_WORKERS=${PW_WORKERS:-} pnpm exec playwright test \"${filter}\" ${grep} \"${headed:-}\" | tee \"$PLAYWRIGHT_LOG_PATH\""; then
    echo "Playwright tests failed. Check logs for details."
    popd >/dev/null

    echo "Playwright test logs saved to: .${PLAYWRIGHT_LOG_PATH#"${WEBLENS_ROOT}"}"

    exit 1
else
    echo "Playwright tests passed successfully."
    exit 0
fi
