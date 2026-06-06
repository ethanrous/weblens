#!/bin/bash

is_embed_running() {
    containerized=true
    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--containerized")
            shift
            containerized="$1"
            ;;
        *)
            "Unknown argument: $1"
            echo "Usage: is_embed_running [--containerized true|false]"
            exit 1
            ;;
        esac
        shift
    done

    if [[ "$containerized" == true ]]; then
        if dockerc ps | grep weblens-embed &>/dev/null; then
            return 0
        fi
    elif pgrep -f "embed.*main.py" &>/dev/null; then
        return 0
    fi

    return 1
}
export -f is_embed_running

launch_embed() {
    containerized=true
    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--containerized")
            shift
            containerized="$1"
            ;;
        *)
            "Unknown argument: $1"
            echo "Usage: launch_embed [--containerized true|false]"
            exit 1
            ;;
        esac
        shift
    done

    if [[ "$containerized" == true ]]; then
        if ! dockerc image ls | grep weblens_embed &>/dev/null; then
            build_embed
        fi

        dockerc run --rm -d --name weblens-embed --publish 5500:5500 -v "${WEBLENS_ROOT}/_build/fs/core/cache/:/images" -v "${WEBLENS_ROOT}/_build/embed/model-cache/:/root/.cache/huggingface" --network weblens-net ghcr.io/ethanrous/weblens_embed
    else
        mkdir -p ${WEBLENS_ROOT}/_build/logs
        touch ${WEBLENS_ROOT}/_build/logs/embed.log

        echo "Launching embed service in development mode"
        (
            cd "${WEBLENS_ROOT}/embed" || exit 1
            uv sync --quiet
            uv run main.py
        ) >"${WEBLENS_ROOT}/_build/logs/embed.log" 2>&1 &

        echo "Waiting for embed server to become ready..."
        for i in $(seq 1 60); do
            if curl -s http://localhost:5500/health > /dev/null 2>&1; then
                echo "Embed development server launched. Logs are being written to ${WEBLENS_ROOT}/_build/logs/embed.log"
                return 0
            fi
            sleep 1
            printf "."
        done

        echo "\nFailed to launch embed development server within timeout. Check logs for details: ${WEBLENS_ROOT}/_build/logs/embed.log"
        return 1
    fi
}
export -f launch_embed

stop_embed() {
    if is_embed_running --containerized true; then
        dockerc stop weblens-embed 2>/dev/null || true

        echo "Removing weblens-embed container..."
    fi

    if is_embed_running --containerized false; then
        echo "Stopping embed development server..."

        pkill -f "embed.*main.py" 2>/dev/null || true
    fi
}
export -f stop_embed
