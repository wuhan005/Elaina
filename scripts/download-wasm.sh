#!/bin/sh
set -eu

runtime_dir=${1:-${WASM_RUNTIME_DIR:-.wasm}}
mkdir -p "$runtime_dir"
runtime_dir=$(CDPATH= cd -- "$runtime_dir" && pwd)
download_dir=$(mktemp -d "$runtime_dir/.download-XXXXXX")
trap 'rm -rf "$download_dir"' EXIT HUP INT TERM

# Pin language artifacts and verify downloads before installing them.
checksum() {
    if command -v sha256sum >/dev/null 2>&1; then
        sha256sum "$1" | cut -d ' ' -f 1
    else
        shasum -a 256 "$1" | cut -d ' ' -f 1
    fi
}

download() {
    name=$1
    url=$2
    expected=$3
    if [ -f "$runtime_dir/$name" ] && [ "$(checksum "$runtime_dir/$name")" = "$expected" ]; then
        return
    fi
    curl --fail --location --silent --show-error --retry 3 "$url" -o "$download_dir/artifact"
    if [ "$(checksum "$download_dir/artifact")" != "$expected" ]; then
        echo "Checksum mismatch: $name" >&2
        exit 1
    fi
    mkdir -p "$(dirname "$runtime_dir/$name")"
    mv "$download_dir/artifact" "$runtime_dir/$name"
}

download python.wasm 'https://github.com/vmware-labs/webassembly-language-runtimes/releases/download/python/3.12.0%2B20231211-040d5a6/python-3.12.0.wasm' e5dc5a398b07b54ea8fdb503bf68fb583d533f10ec3f930963e02b9505f7a763
download php.wasm 'https://github.com/vmware-labs/webassembly-language-runtimes/releases/download/php/8.2.6%2B20230714-11be424/php-8.2.6-wasmedge.wasm' 5461eea8426378e2257f46c7eb1c734f607a89de9a921b72f8fc14c20ac3a33a
download javascript.wasm 'https://github.com/quickjs-ng/quickjs/releases/download/v0.17.0/qjs-wasi.wasm' 42a732a676ec2d93488c19411e0fad283bf72658fdad746f089914b523c783b1

# The no-EH Clang build runs in wazero without a native compiler process.
c_release=https://github.com/cppstudio-io/wasm-clang-runtime/releases/download/v0.1.0
c_source=https://raw.githubusercontent.com/cppstudio-io/wasm-clang-runtime/df1180d80184733c6a01599f76b92b4001d20f87
download c/clang.wasm "$c_release/clang22-noeh" e01346e33b31a3b3359a317450502bd4feb78cd6ca7fed08aca8601f9ff8834d
download c/wasm-ld.wasm "$c_release/lld22-noeh" 63c397c582e193aacd96fbdeb04b1bd32f47ae505f6dc4a569ed08a80ddfeff7
download c/sysroot.tar "$c_release/sysroot22.tar" d991c621cdf9b4640c2cd4aa21abb08491c2ae6ef9917cac5c6065553604bcec
download c/LICENSE "$c_source/LICENSE" cfc7749b96f63bd31c3c42b5c471bf756814053e847c10f3eb003417bc523d30
download c/LICENSE.llvm "$c_source/LICENSE.llvm" ebcd9bbf783a73d05c53ba4d586b8d5813dcdf3bbec50265860ccc885e606f47
download c/NOTICE "$c_source/NOTICE" b5d7bcfb23672141736fee1289c7cdc265c42a29bb47a153d37ec85a7a950a7f
download c/LICENSE.wasi-libc 'https://raw.githubusercontent.com/WebAssembly/wasi-libc/161b3195fc2558d2b1ba3eb9ffae3b2b47407623/LICENSE' 2711a8b5a5cdfef0e639f96c1aca12ae23d7d64a02d0507f1bdf14d2b27bbc3a
mkdir -p "$runtime_dir/c/sysroot" "$runtime_dir/c/include/sys"
tar --no-same-owner -xf "$runtime_dir/c/sysroot.tar" -C "$runtime_dir/c/sysroot"
project_dir=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cp "$project_dir/tools/wasm-c/network.c" "$runtime_dir/c/network.c"
cp "$project_dir/tools/wasm-c/include/netdb.h" "$runtime_dir/c/include/netdb.h"
cp "$project_dir/tools/wasm-c/include/sys/socket.h" "$runtime_dir/c/include/sys/socket.h"
cp "$project_dir/LICENSE" "$runtime_dir/c/LICENSE.elaina"
