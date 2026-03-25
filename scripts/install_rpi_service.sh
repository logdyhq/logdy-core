#!/usr/bin/env bash

set -euo pipefail

BIN_PATH="${1:-/usr/local/bin/logdy}"
SERVICE_PATH="/etc/systemd/system/logdy.service"
SERVICE_USER="${LOGDY_USER:-logdy}"
SERVICE_GROUP="${LOGDY_GROUP:-logdy}"

if [[ ! -x "${BIN_PATH}" ]]; then
  echo "Binary not found or not executable: ${BIN_PATH}"
  echo "Build/copy logdy first, then run this script with sudo."
  exit 1
fi

if ! id -u "${SERVICE_USER}" >/dev/null 2>&1; then
  useradd --system --home /var/lib/logdy --shell /usr/sbin/nologin "${SERVICE_USER}"
fi

if ! getent group "${SERVICE_GROUP}" >/dev/null 2>&1; then
  groupadd --system "${SERVICE_GROUP}"
fi

mkdir -p /var/lib/logdy
chown "${SERVICE_USER}:${SERVICE_GROUP}" /var/lib/logdy

tmp_service="$(mktemp)"
cp packaging/systemd/logdy.service "${tmp_service}"
sed -i "s/^User=.*/User=${SERVICE_USER}/" "${tmp_service}"
sed -i "s/^Group=.*/Group=${SERVICE_GROUP}/" "${tmp_service}"
sed -i "s|^ExecStart=.*|ExecStart=${BIN_PATH} --ui-ip 0.0.0.0 --port 8080 stdin|" "${tmp_service}"

install -m 0644 "${tmp_service}" "${SERVICE_PATH}"
rm -f "${tmp_service}"

systemctl daemon-reload
systemctl enable --now logdy.service
systemctl status --no-pager logdy.service
