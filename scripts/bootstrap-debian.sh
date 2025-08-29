#!/usr/bin/env bash
set -euo pipefail

# Bootstrap a Debian 12 host for Hermetica development & use
# - Installs prerequisites, Go 1.24.x, and pinned PD tools
# - Optionally copies binaries to /usr/local/bin

GO_VERSION=${GO_VERSION:-1.24.6}
INSTALL_SYSTEM_BIN=${INSTALL_SYSTEM_BIN:-0}

echo "[+] Updating apt and installing prerequisites"
sudo apt-get update -y
sudo apt-get install -y ca-certificates curl git build-essential pkg-config unzip \
  libpcap0.8 libpcap-dev

echo "[+] Installing Go ${GO_VERSION} (to /usr/local/go)"
if [ -d /usr/local/go ]; then sudo rm -rf /usr/local/go; fi
curl -fsSL https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz -o /tmp/go.tgz
sudo tar -C /usr/local -xzf /tmp/go.tgz
rm -f /tmp/go.tgz
export PATH=/usr/local/go/bin:${HOME}/go/bin:${PATH}

echo "[+] Installing ProjectDiscovery tools (pinned)"
go install github.com/projectdiscovery/subfinder/v2/cmd/subfinder@v2.8.0
go install github.com/projectdiscovery/dnsx/cmd/dnsx@v1.2.2
go install github.com/projectdiscovery/naabu/v2/cmd/naabu@v2.3.5
go install github.com/projectdiscovery/httpx/cmd/httpx@v1.7.1
go install github.com/projectdiscovery/katana/cmd/katana@v1.2.1
go install github.com/sensepost/gowitness@v3.0.5 || go install github.com/sensepost/gowitness@3.0.5

BIN_DIR="${GOBIN:-${HOME}/go/bin}"
echo "[+] Tools installed to ${BIN_DIR}"
ls -l ${BIN_DIR} | awk '{print $9}'

if [ "${INSTALL_SYSTEM_BIN}" = "1" ]; then
  echo "[+] Copying binaries to /usr/local/bin (requires sudo)"
  for t in subfinder dnsx naabu httpx katana gowitness; do
    if [ -x "${BIN_DIR}/$t" ]; then
      sudo cp "${BIN_DIR}/$t" /usr/local/bin/
    fi
  done
fi

echo "[+] Done. Add the following to your shell profile if needed:"
echo "    export PATH=\"/usr/local/go/bin:\$HOME/go/bin:\$PATH\""
echo "[+] Next: run ./bin/hermetica doctor --fix-paths --dry-run -c configs/hermetica.yaml"

