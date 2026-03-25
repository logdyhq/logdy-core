#!/usr/bin/env bash

set -euo pipefail

VERSION="${1:-dev}"
OUT_DIR="${2:-bin}"
APP_NAME="logdy"
LDFLAGS="-X 'main.Version=${VERSION}'"

mkdir -p "${OUT_DIR}"

build_target() {
  local arch="$1"
  local goarm="${2:-}"
  local output_suffix="linux_${arch}"

  if [[ -n "${goarm}" ]]; then
    output_suffix="${output_suffix}v${goarm}"
  fi

  echo "Building ${APP_NAME}_${output_suffix}"
  GOOS=linux GOARCH="${arch}" CGO_ENABLED=0 GOARM="${goarm}" \
    go build -trimpath -ldflags "${LDFLAGS}" -o "${OUT_DIR}/${APP_NAME}_${output_suffix}" .
}

# Raspberry Pi Zero / 1 (ARMv6, 32-bit)
build_target arm 6

# Raspberry Pi 2 / 3 (ARMv7, 32-bit)
build_target arm 7

# Raspberry Pi 3 / 4 / 5 (ARM64, 64-bit OS)
build_target arm64

echo "Raspberry Pi binaries available in ${OUT_DIR}/"
