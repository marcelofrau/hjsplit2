#!/usr/bin/env bash
set -euo pipefail

PLATFORM="${1:-all}"

# ── Config ──────────────────────────────────────────────────────────────────
declare -A CONFIG=()
if [ -f build.conf ]; then
  while IFS='=' read -r key value; do
    key="${key// /}"
    [ -z "$key" ] && continue
    [[ "$key" =~ ^# ]] && continue
    CONFIG["$key"]="$value"
  done < build.conf
fi

get_cfg() { echo "${CONFIG[$1]:-}"; }

# ── Version ─────────────────────────────────────────────────────────────────
if [ ! -f version.txt ]; then
  echo "Error: version.txt not found" >&2
  exit 1
fi
VERSION=$(tr -d ' \t\n\r' < version.txt)

OUTPUT_DIR="dist"
PACKAGE="."
LD_FLAGS="-s -w -X hjsplit2/internal/version.Version=${VERSION}"
ANY_OK=0

mkdir -p "$OUTPUT_DIR"

build() {
  local goos="$1" goarch="$2" suffix="$3"
  local name="hjsplit2-v${VERSION}-${goos}-${goarch}${suffix}"

  echo "==> Building: $name"
  if CGO_ENABLED=1 GOOS="$goos" GOARCH="$goarch" \
       go build -ldflags="$LD_FLAGS" -o "${OUTPUT_DIR}/${name}" "$PACKAGE" 2>&1; then
    size=$(du -h "${OUTPUT_DIR}/${name}" | cut -f1)
    echo "    OK: ${OUTPUT_DIR}/${name} ($size)"
    ANY_OK=1

    # UPX compression
    if [ "$(get_cfg USE_UPX)" = "true" ] && command -v upx &>/dev/null; then
      echo "    Compressing with UPX..."
      upx --best --no-color --no-progress "${OUTPUT_DIR}/${name}" 2>/dev/null || true
      comp=$(du -h "${OUTPUT_DIR}/${name}" | cut -f1)
      echo "    Compressed: $comp"
    fi
  else
    echo "    FAILED"
  fi
}

# ── Detect compilers ────────────────────────────────────────────────────────
HOST_OS=$(go env GOOS)
HOST_ARCH=$(go env GOARCH)
CC_TARGET=$(gcc -dumpmachine 2>/dev/null || echo "unknown")
echo "Host: $HOST_OS/$HOST_ARCH  CC: $CC_TARGET"

CC_386="${CONFIG[CC_386]:-}"
MINGW32_PATH="${CONFIG[MINGW32_PATH]:-}"

if [ -z "$CC_386" ] && [ -n "$MINGW32_PATH" ]; then
  candidate="${MINGW32_PATH}/bin/gcc.exe"
  [ -f "$candidate" ] && CC_386="$candidate"
fi
if [ -z "$CC_386" ]; then
  candidate=$(command -v i686-w64-mingw32-gcc 2>/dev/null || true)
  [ -n "$candidate" ] && CC_386="$candidate"
fi

echo ""

# ── Build ───────────────────────────────────────────────────────────────────
case "$PLATFORM" in
  native)
    build "$HOST_OS" "$HOST_ARCH" ""
    ;;
  win32)
    if [ -n "$CC_386" ]; then
      CC="$CC_386" build windows 386 ".exe"
    else
      echo "win32: configure CC_386 or MINGW32_PATH in build.conf"
      echo "  Or install: sudo apt install gcc-mingw-w64-i686"
    fi
    ;;
  win64)
    if echo "$CC_TARGET" | grep -qiE 'mingw32|windows'; then
      build windows amd64 ".exe"
    else
      echo "win64: install MinGW-w64"
      echo "  sudo apt install gcc-mingw-w64-x86-64"
    fi
    ;;
  linux)
    if [ "$HOST_OS" = "linux" ]; then
      if command -v gcc &>/dev/null; then
        build linux amd64 ""
      else
        echo "    FAILED: gcc not found (needed for CGO/Fyne)"
        echo "    Install: sudo apt install gcc"
        exit 1
      fi
    else
      echo "linux: build natively on Linux or use WSL"
    fi
    ;;
  all)
    echo "--- Building all platforms ---"
    build "$HOST_OS" "$HOST_ARCH" ""

    if echo "$CC_TARGET" | grep -qiE 'mingw32|windows'; then
      build windows amd64 ".exe"
    else
      echo "Skipping win64 (no MinGW)"
    fi

    if [ -n "$CC_386" ]; then
      CC="$CC_386" build windows 386 ".exe"
    else
      echo "Skipping win32 (configure CC_386 in build.conf)"
    fi

    if [ "$HOST_OS" = "linux" ]; then
      if command -v gcc &>/dev/null; then
        build linux amd64 ""
      else
        echo "    Skipping linux (gcc not found, install: sudo apt install gcc)"
      fi
    fi
    ;;
  *)
    echo "Usage: $0 [win32|win64|linux|all|native]"
    exit 1
    ;;
esac

if [ "$ANY_OK" -eq 0 ]; then
  echo ""
  echo "No builds succeeded." >&2
  exit 1
fi
