#!/bin/bash

get_color_code() {
    local color_name="$1"
    case "$color_name" in
    "blue") printf '34m' ;;
    "green") printf '32m' ;;
    "orange") printf '38;2;255;165;0m' ;;
    *) printf "" ;; # Default to no color
    esac
}

show_as_subtask() {
    local task_name="$1"
    shift

    task_color="blue"
    verbose=false

    while [ "${1:-}" != "" ]; do
        case "$1" in
        "--color")
            shift
            task_color="$1"
            ;;
        "-v" | "--verbose")
            verbose=true
            ;;
        "--")
            break
            ;;
        esac
        shift
    done

    # Skip the -- separator
    if [[ "${1:-}" == "--" ]]; then
        shift
    fi

    local color_code
    color_code="$(get_color_code "$task_color")"
    local esc=$'\e'

    if [[ "$WEBLENS_VERBOSE" = "false" ]] && [[ "$verbose" = false ]]; then
        printf "\e[%s|-- %s\e[0m\n" "$color_code" "$task_name"
        local buf cmd_status
        buf=$("$@" 2>&1) && cmd_status=0 || cmd_status=$?
        if [[ $cmd_status -ne 0 ]]; then
            local err_prefix="${esc}[31m| ${esc}[0m"
            printf "\e[31m|-- %s failed (exit code %d):\e[0m\n" "$task_name" "$cmd_status"
            printf '%s\n' "${err_prefix}${buf//$'\n'/$'\n'${err_prefix}}"
            printf "\e[31m|--------------------------\e[0m\n\n"
        fi
        return "$cmd_status"
    fi

    printf "\e[%s|-- %s\e[0m\n" "$color_code" "$task_name"

    "$@" 2>&1 | sed "s/^/${esc}[${color_code}| ${esc}[0m/"
    local cmd_status=${PIPESTATUS[0]}

    printf "\e[%s|--------------------------\e[0m\n\n" "$color_code"
    return "$cmd_status"
}
export -f show_as_subtask
