#!/usr/bin/env sh
set -eu
VERSION="${VERSION:-0.1.0}"
LDFLAGS="-s -w -X main.version=${VERSION}"
ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
cd "$ROOT"

mkdir -p bin dist
echo "构建当前平台 -> bin/ashow"
CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o bin/ashow ./cmd/ashow

build_one() {
  goos="$1"
  goarch="$2"
  out="$3"
  echo "交叉编译 ${goos}/${goarch} -> ${out}"
  CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" go build -ldflags "$LDFLAGS" -o "$out" ./cmd/ashow
}

build_one windows amd64 dist/ashow-windows-amd64.exe
build_one windows arm64 dist/ashow-windows-arm64.exe
build_one linux amd64 dist/ashow-linux-amd64
build_one linux arm64 dist/ashow-linux-arm64
build_one darwin amd64 dist/ashow-darwin-amd64
build_one darwin arm64 dist/ashow-darwin-arm64

echo "完成。输出目录: bin/ 与 dist/"
ls -l bin dist
