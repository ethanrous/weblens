#!/bin/bash

# Single source of truth for the agno version: the Go binding pinned in go.mod.
# Keeping the native libagno.a in lockstep with the bindings avoids linking a
# library against mismatched binding declarations.
agno_version() {
    awk '/github.com\/ethanrous\/agno\/bindings\/go\/agno / {print $2; exit}' "${WEBLENS_ROOT}/go.mod"
}
export -f agno_version

# Build agno from a local path when WEBLENS_LOCAL_AGNO is
# set, otherwise fetch the pinned release.
setup_agno() {
    if [[ -n "${WEBLENS_LOCAL_AGNO:-}" ]]; then
        build_agno_local "$WEBLENS_LOCAL_AGNO"
    else
        fetch_agno
    fi
}
export -f setup_agno

# Downloads the pinned agno release lib if missing. Never builds from source.
fetch_agno() {
    local agno_lib_dir_path
    if [[ -n "${1:-}" ]]; then
        agno_lib_dir_path="$1"
    else
        agno_lib_dir_path="${WEBLENS_ROOT}/_build/lib"
    fi
    mkdir -p "$agno_lib_dir_path"

    lib_agno_path="$agno_lib_dir_path/libagno.a"

    local AGNO_VERSION
    AGNO_VERSION="$(agno_version)"
    if [[ -z "$AGNO_VERSION" ]]; then
        echo "Could not determine agno version from ${WEBLENS_ROOT}/go.mod" >&2
        return 1
    fi

    if [[ ! -e "$lib_agno_path" ]] || [[ ! -e "$agno_lib_dir_path/version.txt" ]] || [[ "$(cat "$agno_lib_dir_path/version.txt")" != "$AGNO_VERSION" ]]; then
        latest_release=$(gh api -H "Accept: application/vnd.github+json" -H "X-GitHub-Api-Version: 2026-03-10" '/repos/ethanrous/agno/releases?per_page=1' | jq -r '.[0].name')
        if [[ "$latest_release" != "$AGNO_VERSION" ]]; then
            orange=$(get_color_code "orange")
            printf "\e[%sUpdate available: Latest release of agno is $latest_release, but we are loading $AGNO_VERSION. Consider updating the agno bindings in go.mod to the latest release.\n\e[0m" "$orange"
        fi

        local arch=$(uname -m)
        local platform="macos"

        if [[ "$(uname)" == "Linux" ]]; then
            platform="linux"
        fi

        if [[ "$arch" == "aarch64" ]] || [[ "$arch" == "arm64" ]]; then
            arch_suffix="-aarch64-gpu"
        else
            arch_suffix="-x86_64-gpu"
        fi

        echo "Fetching agno release $AGNO_VERSION"

        local agno_url="https://github.com/ethanrous/agno/releases/download/${AGNO_VERSION}/libagno-${platform}${arch_suffix}.a"
        if ! curl -fsSL "$agno_url" -o "$lib_agno_path"; then
            echo "Failed to download agno release from $agno_url" >&2
            rm -f "$lib_agno_path"
            return 1
        fi

        if ! ranlib "$lib_agno_path"; then
            echo "ranlib failed on $lib_agno_path" >&2
            rm -f "$lib_agno_path"
            return 1
        fi

        # Only record the version after a verified-good download, so a failed
        # fetch is retried rather than cached as valid.
        echo "$AGNO_VERSION" > "$agno_lib_dir_path/version.txt"
    fi

    setup_agno_cgo
}
export -f fetch_agno

# Builds libagno.a from a local agno checkout and points the Go toolchain at
# its bindings via a generated workspace file. Rebuilds on every call (cargo
# is incremental, so this is fast when nothing changed).
build_agno_local() {
    # Canonicalize: a trailing slash would produce "//" in the go.work use
    # path, which go.work syntax treats as a comment start
    local agno_repo
    if ! agno_repo="$(cd "$1" 2>/dev/null && pwd)"; then
        echo "No such directory: $1" >&2
        return 1
    fi

    if [[ ! -f "$agno_repo/justfile" ]] || [[ ! -f "$agno_repo/bindings/go/agno/go.mod" ]]; then
        echo "Not an agno checkout (missing justfile or bindings/go/agno/go.mod): $agno_repo" >&2
        return 1
    fi
    if ! command -v just >/dev/null; then
        echo "'just' is required to build agno locally (https://just.systems)" >&2
        return 1
    fi

    local lib_dir="${WEBLENS_ROOT}/_build/lib/local"
    mkdir -p "$lib_dir"

    printf "Building agno from \e[34m%s\e[0m...\n" "$agno_repo"
    (cd "$agno_repo" && just build "$lib_dir/libagno.a")
    ranlib "$lib_dir/libagno.a"

    local go_version
    go_version=$(awk '/^go / {print $2; exit}' "$WEBLENS_ROOT/go.mod")

    local work_file="${WEBLENS_ROOT}/_build/agno.work"
    cat >"$work_file" <<EOF
go ${go_version}

use (
	${WEBLENS_ROOT}
	${WEBLENS_ROOT}/api
	${agno_repo}/bindings/go/agno
)
EOF

    export AGNO_LIB_DIR="$lib_dir"
    export GOWORK="$work_file"

    setup_agno_cgo
}
export -f build_agno_local

# Cross-builds libagno.a for linux/<arch> from a local agno checkout and stages
# it next to the Go bindings for the container build to pick up. The lib must be
# built in a linux container: a host build on macos produces a darwin archive.
build_agno_for_image() {
    local agno_repo
    if ! agno_repo="$(cd "$1" 2>/dev/null && pwd)"; then
        echo "No such directory: $1" >&2
        return 1
    fi

    if [[ ! -f "$agno_repo/build/sh/build-agno.bash" ]] || [[ ! -f "$agno_repo/bindings/go/agno/go.mod" ]]; then
        echo "Not an agno checkout (missing build/sh/build-agno.bash or bindings/go/agno/go.mod): $agno_repo" >&2
        return 1
    fi

    local target_triple target_cpu cross_pkg cross_cc linker_var cc_var
    case "$2" in
    "amd64" | "x86_64")
        target_triple="x86_64-unknown-linux-gnu"
        target_cpu="x86_64"
        cross_pkg="gcc-x86-64-linux-gnu libc6-dev-amd64-cross"
        cross_cc="x86_64-linux-gnu-gcc"
        linker_var="CARGO_TARGET_X86_64_UNKNOWN_LINUX_GNU_LINKER"
        cc_var="CC_x86_64_unknown_linux_gnu"
        ;;
    "arm64" | "aarch64")
        target_triple="aarch64-unknown-linux-gnu"
        target_cpu="aarch64"
        cross_pkg="gcc-aarch64-linux-gnu libc6-dev-arm64-cross"
        cross_cc="aarch64-linux-gnu-gcc"
        linker_var="CARGO_TARGET_AARCH64_UNKNOWN_LINUX_GNU_LINKER"
        cc_var="CC_aarch64_unknown_linux_gnu"
        ;;
    *)
        echo "Unsupported architecture for a local agno build: $2" >&2
        return 1
        ;;
    esac

    local host_cpu
    case "$(uname -m)" in
    "arm64" | "aarch64")
        host_cpu="aarch64"
        ;;
    *)
        host_cpu="x86_64"
        ;;
    esac

    # The gpu kernels crate links a shared object for the target, so building for
    # a foreign arch needs that arch's toolchain. Baked into an image so apt only
    # runs on the first build.
    local builder_image="rust:1-bookworm"
    local cross_args=()
    if [[ "$host_cpu" != "$target_cpu" ]]; then
        builder_image="weblens-agno-builder:$target_cpu"
        if ! dockerc build -q -t "$builder_image" - >/dev/null <<EOF
FROM rust:1-bookworm
RUN apt-get update && apt-get install -y --no-install-recommends $cross_pkg && rm -rf /var/lib/apt/lists/*;
EOF
        then
            echo "Failed to build the $target_cpu cross toolchain image" >&2
            return 1
        fi

        cross_args=(-e "$linker_var=$cross_cc" -e "$cc_var=$cross_cc")
    fi

    local out_dir="${WEBLENS_ROOT}/_build/agno-image"
    rm -rf "$out_dir"
    mkdir -p "$out_dir"

    printf "Building agno for \e[34mlinux/%s\e[0m from \e[34m%s\e[0m...\n" "$2" "$agno_repo"

    # bookworm to match the glibc of the golang/debian stages in the image.
    # target/ is a volume so the linux build never collides with (or leaves
    # root-owned files in) the host's darwin build, and rustup/cargo are cached
    # because the toolchain pinned by rust-toolchain.toml is a ~1GB download.
    if ! dockerc run --rm \
        -v "$agno_repo:/src" \
        -v "$out_dir:/out" \
        -v weblens-agno-target:/src/target \
        -v weblens-agno-cargo:/usr/local/cargo/registry \
        -v weblens-agno-rustup:/usr/local/rustup \
        -e TARGET_TRIPLE="$target_triple" \
        ${cross_args[@]+"${cross_args[@]}"} \
        -w /src "$builder_image" \
        /src/build/sh/build-agno.bash /out/libagno.a; then
        echo "Failed to build agno for linux/$2 from $agno_repo" >&2
        return 1
    fi

    # The bindings must travel with the lib: linking a locally built library
    # against the released binding declarations is exactly the mismatch the
    # go.mod pin exists to prevent.
    mkdir -p "$out_dir/bindings/go"
    cp -R "$agno_repo/bindings/go/agno" "$out_dir/bindings/go/agno"
}
export -f build_agno_for_image

setup_agno_cgo() {
    echo "Setting up CGO flags for AGNO..."
    local lib_dir="${AGNO_LIB_DIR:-${WEBLENS_ROOT}/_build/lib}"
    export CGO_CFLAGS="-I${WEBLENS_ROOT}/agno/lib"
    export CGO_LDFLAGS="-L${lib_dir} -lagno -lstdc++ -lm"
}
export -f setup_agno_cgo
