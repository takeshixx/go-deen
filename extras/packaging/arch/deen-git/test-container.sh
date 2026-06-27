#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/../../../.." && pwd)"

container_bin="${CONTAINER_BIN:-container}"
image="${ARCH_IMAGE:-archlinux:latest}"
platform="${ARCH_PLATFORM:-linux/amd64}"

if ! command -v "${container_bin}" >/dev/null 2>&1; then
  printf 'error: %s not found\n' "${container_bin}" >&2
  exit 127
fi

"${container_bin}" run \
  --rm \
  --platform "${platform}" \
  -v "${repo_root}:/work" \
  "${image}" \
  bash -lc '
set -euo pipefail

# Apple container and some other lightweight runtimes do not support pacman
# download sandbox user switching. This is limited to the disposable container.
sed -i "/^\[options\]/a DisableSandbox" /etc/pacman.conf

pacman -Syu --noconfirm --needed \
  base-devel \
  git \
  go \
  pkgconf \
  libglvnd \
  libx11 \
  libxcursor \
  libxfixes \
  libxi \
  libxinerama \
  libxrandr

useradd -m builder
install -d -o builder -g builder /home/builder/pkg
cp -a /work/extras/packaging/arch/deen-git/. /home/builder/pkg/
chown -R builder:builder /home/builder/pkg

cd /home/builder/pkg
su builder -c "makepkg --syncdeps --noconfirm --needed"

pacman -U --noconfirm ./*.pkg.tar.*

LANG=C.UTF-8 deen -version
pacman -Q deen-git
pacman -Ql deen-git | grep -E "/usr/bin/deen|deen.desktop|LICENSE|README"
'
