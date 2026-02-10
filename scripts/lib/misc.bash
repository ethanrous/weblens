portable_sed() {
    if [[ "$(uname)" == "Darwin" ]]; then
        sed -i '' "$1" "$2"
    else
        sed -i "$1" "$2"
    fi
}

get_log_file() {
    log_prefix="$1"

    timestamp=$(date +"%Y-%m-%d_%H-%M-%S")
    log_filename="${log_prefix}-${timestamp}.log"

    log_path="$WEBLENS_ROOT/_build/logs/$log_filename"

    mkdir -p "$(dirname "$log_path")"
    touch "$log_path"
    ln -sf "$log_filename" $WEBLENS_ROOT/_build/logs/"${log_prefix}"-latest.log

    echo "$log_path"
}

export -f portable_sed
