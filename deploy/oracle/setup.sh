#!/usr/bin/env bash
# Bootstraps the calculator on a fresh Ubuntu VM (Oracle Cloud Always Free, ARM or x86).
# Safe to re-run: it updates the code and redeploys.
#
#   ./setup.sh <git-repo-url> [domain]
#
#   domain  optional; enables HTTPS via Caddy + Let's Encrypt (DNS must point to this VM).
#           Without it the app is served over HTTP on the VM's public IP.
#
# Env: BRANCH (default: main), APP_DIR (default: ~/calculator)
set -euo pipefail

REPO_URL=${1:?usage: setup.sh <git-repo-url> [domain]}
SITE_ADDRESS=${2:-:80}
BRANCH=${BRANCH:-main}
APP_DIR=${APP_DIR:-$HOME/calculator}

echo "==> Installing Docker (if needed)"
if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sudo sh
  sudo usermod -aG docker "$USER"
fi

echo "==> Opening ports 80/443 in the VM firewall"
# Oracle's Ubuntu images ship iptables rules that REJECT everything except SSH.
# This is in addition to the VCN Security List ingress rules in the console.
open_port() {
  local proto=$1 port=$2
  sudo iptables -C INPUT -p "$proto" --dport "$port" -m conntrack --ctstate NEW -j ACCEPT 2>/dev/null ||
    sudo iptables -I INPUT -p "$proto" --dport "$port" -m conntrack --ctstate NEW -j ACCEPT
}
open_port tcp 80
open_port tcp 443
open_port udp 443
if command -v netfilter-persistent >/dev/null 2>&1; then
  sudo netfilter-persistent save
fi

echo "==> Fetching ${REPO_URL} (${BRANCH})"
if [ -d "$APP_DIR/.git" ]; then
  git -C "$APP_DIR" fetch --quiet origin "$BRANCH"
  git -C "$APP_DIR" checkout --quiet "$BRANCH"
  git -C "$APP_DIR" pull --quiet --ff-only origin "$BRANCH"
else
  git clone --quiet --branch "$BRANCH" "$REPO_URL" "$APP_DIR"
fi
cd "$APP_DIR"

echo "==> Building and starting (SITE_ADDRESS=${SITE_ADDRESS})"
echo "SITE_ADDRESS=${SITE_ADDRESS}" >.env
sudo docker compose -f compose.yaml -f deploy/oracle/compose.prod.yaml up -d --build --remove-orphans
sudo docker image prune -f >/dev/null

echo "==> Waiting for the API"
for _ in $(seq 1 30); do
  if curl -fsS http://localhost/api/health >/dev/null 2>&1; then
    break
  fi
  sleep 2
done
curl -fsS http://localhost/api/health && echo

PUBLIC_IP=$(curl -fsS https://ifconfig.me 2>/dev/null || echo "<public-ip>")
if [ "$SITE_ADDRESS" = ":80" ]; then
  echo "==> Done: http://${PUBLIC_IP}  (Swagger: http://${PUBLIC_IP}/api/docs)"
else
  echo "==> Done: https://${SITE_ADDRESS}  (Swagger: https://${SITE_ADDRESS}/api/docs)"
fi
