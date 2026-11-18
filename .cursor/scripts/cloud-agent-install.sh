#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
go_mod="${repo_root}/api/go.mod"
go_root="/usr/local/go"
profile_snippet="/etc/profile.d/cursor-go.sh"

if [[ ! -f "$go_mod" ]]; then
  echo "api/go.mod not found at ${go_mod}" >&2
  exit 1
fi

go_version="$(awk '/^go / {print $2; exit}' "$go_mod")"
if [[ -z "$go_version" ]]; then
  echo "Could not read Go version from ${go_mod}" >&2
  exit 1
fi

install_go() {
  tmp="$(mktemp)"
  curl -fsSL "https://go.dev/dl/go${go_version}.linux-amd64.tar.gz" -o "$tmp"
  sudo rm -rf "$go_root"
  sudo tar -C /usr/local -xzf "$tmp"
  rm -f "$tmp"
}

if [[ ! -x "${go_root}/bin/go" ]]; then
  echo "Installing Go ${go_version} to ${go_root}"
  install_go
elif ! "${go_root}/bin/go" version | grep -q "go${go_version}"; then
  echo "Replacing $("${go_root}/bin/go" version | awk '{print $3}') with Go ${go_version}"
  install_go
else
  echo "Go ${go_version} already installed at ${go_root}"
fi

if [[ ! -f "$profile_snippet" ]] || ! grep -q '/usr/local/go/bin' "$profile_snippet"; then
  echo 'export PATH="/usr/local/go/bin:$PATH"' | sudo tee "$profile_snippet" >/dev/null
  sudo chmod 644 "$profile_snippet"
fi

export PATH="/usr/local/go/bin:$PATH"
"${go_root}/bin/go" version

bun_version="1.4.2"
bun_root="${HOME}/.bun"
bun_profile_snippet="/etc/profile.d/cursor-bun.sh"

install_bun() {
  curl -fsSL "https://bun.sh/install" | bash -s "bun-v${bun_version}"
}

if [[ ! -x "${bun_root}/bin/bun" ]]; then
  echo "Installing Bun ${bun_version} to ${bun_root}"
  install_bun
elif ! "${bun_root}/bin/bun" --version | grep -q "^${bun_version}"; then
  echo "Replacing Bun $("${bun_root}/bin/bun" --version) with Bun ${bun_version}"
  install_bun
else
  echo "Bun ${bun_version} already installed at ${bun_root}"
fi

if [[ ! -f "$bun_profile_snippet" ]] || ! grep -q "${bun_root}/bin" "$bun_profile_snippet"; then
  echo "export PATH=\"${bun_root}/bin:\$PATH\"" | sudo tee "$bun_profile_snippet" >/dev/null
  sudo chmod 644 "$bun_profile_snippet"
fi

export PATH="${bun_root}/bin:$PATH"
"${bun_root}/bin/bun" --version

cd "${repo_root}/frontend"
bun install --frozen-lockfile

# Warm the Go build cache so the first `make test` / `go build` / `go fix`
# starts from a hot cache instead of recompiling every vendored package.
echo "Warming Go build cache (vendored)"
cd "${repo_root}/api"
"${go_root}/bin/go" build -mod=vendor ./...
