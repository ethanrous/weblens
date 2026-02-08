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

    local task_color
    if [[ -z "${2+x}" ]]; then
        task_color="blue"
    else
        task_color="$2"
    fi

    local color_code
    color_code="$(get_color_code "$task_color")"
    printf "\e[%s|-- %s\e[0m\n" "$color_code" "$task_name"
    if [[ "$WEBLENS_VERBOSE" = "false" ]]; then
        cat - >/dev/null
        return
    fi

    local esc=$'\e'
    sed "s/^/${esc}[${color_code}| ${esc}[0m/"

    printf "\e[%s|--------------------------\e[0m\n\n" "$color_code"
}
export -f show_as_subtask
