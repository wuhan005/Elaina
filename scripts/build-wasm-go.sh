#!/bin/sh
set -eu

runtime_dir=${1:-${WASM_RUNTIME_DIR:-.wasm}}
mkdir -p "$runtime_dir"
runtime_dir=$(CDPATH= cd -- "$runtime_dir" && pwd)
project_dir=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$project_dir/tools/wasm-go"
GOTOOLCHAIN=go1.27.3 go run . "$runtime_dir/go"
