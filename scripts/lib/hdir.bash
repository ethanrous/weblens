#!/bin/bash

is_hdir_running() {
    containerized=true
    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--containerized")
            shift
            containerized="$1"
            ;;
        *)
            "Unknown argument: $1"
            echo "Usage: is_hdir_running [--containerized true|false]"
            exit 1
            ;;
        esac
        shift
    done

    if [[ "$containerized" == true ]]; then
        if dockerc ps | grep weblens-hdir &>/dev/null; then
            return 0
        fi
    elif pgrep -f "uv run open.main.py" &>/dev/null; then
        return 0
    fi

    return 1
}
export -f is_hdir_running

launch_hdir() {
    containerized=true
    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--containerized")
            shift
            containerized="$1"
            ;;
        *)
            "Unknown argument: $1"
            echo "Usage: launch_hdir [--containerized true|false]"
            exit 1
            ;;
        esac
        shift
    done

    if [[ "$containerized" == true ]]; then
        if ! dockerc image ls | grep weblens_hdir &>/dev/null; then
            build_hdir
        fi

        dockerc run --rm -d --name weblens-hdir --publish 5500:5500 -v "${WEBLENS_ROOT}/_build/fs/core/cache/:/images" -v "${WEBLENS_ROOT}/_build/hdir/model-cache/:/root/.cache/huggingface" --network weblens-net ethrous/weblens_hdir
    else
        mkdir -p ${WEBLENS_ROOT}/_build/logs
        touch ${WEBLENS_ROOT}/_build/logs/hdir.log

        echo "Launching HDIR in development mode"
        (
            cd "${WEBLENS_ROOT}/hdir" || exit 1
            uv run open.main.py
        ) >"${WEBLENS_ROOT}/_build/logs/hdir.log" 2>&1 &

        echo "Waiting for HDIR server to become ready..."
        for i in $(seq 1 60); do
            if curl -s http://localhost:5500/health > /dev/null 2>&1; then
                echo "HDIR development server launched. Logs are being written to ${WEBLENS_ROOT}/_build/logs/hdir.log"
                return 0
            fi
            sleep 1
        done

        echo "Failed to launch HDIR development server within timeout. Check logs for details: ${WEBLENS_ROOT}/_build/logs/hdir.log"
        return 1
    fi
}
export -f launch_hdir

stop_hdir() {
    if is_hdir_running --containerized true; then
        dockerc stop weblens-hdir 2>/dev/null || true

        echo "Removing weblens-hdir container..."
    fi

    if is_hdir_running --containerized false; then
        pkill -f "uv run open.main.py" 2>/dev/null || true

        echo "Stopping HDIR development server..."
    fi
}
export -f stop_hdir
