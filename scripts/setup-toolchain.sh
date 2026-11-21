#!/bin/sh
# Installs the ARM64 cross-compilation toolchain on Debian 11 (bullseye) amd64.
#
# Debian 11 is used on purpose: its glibc and SDL2 versions are old enough for
# the binary to run on the TrimUI firmware (this is what other NextUI paks do).
# Used by the Dockerfile and by the GitHub Actions build job.
#
#   GO_VERSION   Go version to install (default: the "go" line of go.mod, or 1.26.0)
set -eu

GO_VERSION="${GO_VERSION:-}"
if [ -z "$GO_VERSION" ] && [ -f go.mod ]; then
    GO_VERSION="$(sed -n 's/^go \([0-9.]*\)$/\1/p' go.mod)"
fi
GO_VERSION="${GO_VERSION:-1.26.0}"

# Debian 11 is end-of-life: its packages now live on archive.debian.org.
cat >/etc/apt/sources.list <<SOURCES
deb http://archive.debian.org/debian bullseye main
deb http://archive.debian.org/debian bullseye-updates main
deb http://archive.debian.org/debian-security bullseye-security main
SOURCES
rm -f /etc/apt/sources.list.d/*

dpkg --add-architecture arm64
apt-get -o Acquire::Check-Valid-Until=false update
DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    ca-certificates curl make zip git pkg-config \
    crossbuild-essential-arm64 \
    libsdl2-dev:arm64 libsdl2-ttf-dev:arm64 libsdl2-image-dev:arm64 libsdl2-gfx-dev:arm64
rm -rf /var/lib/apt/lists/*

# my355 (Miyoo Flip, experimental): NextUI targets glibc 2.28 there
# (LoveRetro/gcc-arm-8.3-aarch64-my355), and the Debian 11 libSDL2_gfx
# needs glibc 2.29 (libm). The same SDL2_gfx 1.0.4 built by Debian 10
# needs glibc 2.17: the pak bundles it for my355 (resources/lib/my355).
GFX_DEB=libsdl2-gfx-1.0-0_1.0.4+dfsg-3_arm64.deb
GFX_DEB_SHA256=3f029efb72687afd92e532e6f0cd2775e418674c43f8f067dc7e9f3b08616efd
curl -fsSL -o "/tmp/$GFX_DEB" "http://archive.debian.org/debian/pool/main/libs/libsdl2-gfx/$GFX_DEB"
echo "$GFX_DEB_SHA256  /tmp/$GFX_DEB" | sha256sum -c -
dpkg-deb -x "/tmp/$GFX_DEB" /tmp/gfx-buster
mkdir -p /usr/lib/aarch64-linux-gnu/glibc-2.28
cp -L /tmp/gfx-buster/usr/lib/aarch64-linux-gnu/libSDL2_gfx-1.0.so.0 /usr/lib/aarch64-linux-gnu/glibc-2.28/
rm -rf "/tmp/$GFX_DEB" /tmp/gfx-buster

curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" | tar -C /usr/local -xz
ln -sf /usr/local/go/bin/go /usr/local/bin/go
ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt
go version
