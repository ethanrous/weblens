portable_sed() {
    if [[ "$(uname)" == "Darwin" ]]; then
        sed -i '' "$1" "$2"
    else
        sed -i "$1" "$2"
    fi
}

get_log_file() {
    local log_prefix=""
    local sub_dir=""

    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--prefix")
            shift
            log_prefix="$1"
            ;;
        "--subdir")
            shift
            sub_dir="$1"
            ;;
        "-h" | "--help")
            echo "Usage: get_log_file [--prefix <log_prefix>] [--subdir <sub_dir>]"
            exit 0
            ;;
        esac
        shift
    done

    if [[ -z "$log_prefix" ]]; then
        echo "get_log_file requires a log_prefix argument. Aborting."
        return 1
    fi

    local timestamp
    timestamp=$(date +"%Y-%m-%d_%H-%M-%S")
    local log_filename="${log_prefix}-${timestamp}.log"

    local log_path="$WEBLENS_ROOT/_build/logs/"
    if [[ -n "$sub_dir" ]]; then
        log_path+="${sub_dir}/"
    fi
    log_path+="$log_filename"

    mkdir -p "$(dirname "$log_path")"
    touch "$log_path"
    # ln -sf "$log_filename" "$WEBLENS_ROOT/_build/logs/${log_prefix}-latest.log"

    echo "$log_path"
}

export -f portable_sed
