WEBLENS_ROOT=$(git rev-parse --show-toplevel)
export WEBLENS_ROOT

if [[ -z "${WEBLENS_VERBOSE+x}" ]]; then
    echo "Use WEBLENS_VERBOSE=true to enable verbose output"
    export WEBLENS_VERBOSE=false
fi

LIB_DIR="$WEBLENS_ROOT/scripts/lib"

source "$LIB_DIR/meta.bash"
source "$LIB_DIR/build.bash"
source "$LIB_DIR/mongo.bash"
source "$LIB_DIR/docker-common.bash"
