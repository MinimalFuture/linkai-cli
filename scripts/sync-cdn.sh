#!/usr/bin/env bash
# Mirror release artifacts to the CDN origin bucket and refresh edge caches.
#
# Publishes to cdn.link-ai.tech/cli/:
#   cli/<version>/*           versioned binaries, checksums, skill archive
#   cli/install.sh            latest install script (fixed path)
#   cli/install.ps1           latest Windows install script (fixed path)
#   cli/latest.txt            pointer to the current version (fixed path)
#
# The fixed-path files are cache-purged so users always get the newest copy.
#
# Required env (neutral names; configured as GitHub secrets):
#   RELEASE_VERSION   e.g. v1.2.3
#   CDN_SECRET_ID     object-storage access key id
#   CDN_SECRET_KEY    object-storage access key secret
#   CDN_BUCKET        origin bucket name
#   CDN_REGION        origin bucket region (e.g. ap-guangzhou)
#   CDN_DOMAIN        public CDN host (e.g. cdn.link-ai.tech)

set -euo pipefail

: "${RELEASE_VERSION:?RELEASE_VERSION is required}"
: "${CDN_SECRET_ID:?CDN_SECRET_ID is required}"
: "${CDN_SECRET_KEY:?CDN_SECRET_KEY is required}"
: "${CDN_BUCKET:?CDN_BUCKET is required}"
: "${CDN_REGION:?CDN_REGION is required}"
: "${CDN_DOMAIN:?CDN_DOMAIN is required}"

VERSION="${RELEASE_VERSION#v}"   # strip leading 'v' → path is cli/0.1.0/
PREFIX="cli"                     # path prefix under the CDN host
DIST="dist"                      # goreleaser output dir
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

log() { printf '  %s\n' "$*"; }

# ── Object-storage CLI (coscli) ───────────────────────────────────────────────

COSCLI="$WORK/coscli"
log "==> Installing object-storage CLI"
# coscli release assets are versioned (e.g. coscli-v1.0.8-linux-amd64), so there
# is no fixed "latest" filename. Resolve the linux-amd64 asset from the GitHub
# API and download it by its exact URL.
COSCLI_URL="$(
  curl -fsSL "https://api.github.com/repos/tencentyun/coscli/releases/latest" \
    | grep -oE '"browser_download_url": *"[^"]*linux-amd64"' \
    | head -1 | sed -E 's/.*"(https[^"]+)"/\1/'
)"
[ -n "$COSCLI_URL" ] || { echo "could not resolve coscli download URL" >&2; exit 1; }
log "coscli: $COSCLI_URL"
curl -fsSL -o "$COSCLI" "$COSCLI_URL"
chmod +x "$COSCLI"

CONF="$WORK/coscli.conf"
cat > "$CONF" <<EOF
cos:
  base:
    secretid: ${CDN_SECRET_ID}
    secretkey: ${CDN_SECRET_KEY}
    sessiontoken: ""
    protocol: https
  buckets:
    - name: ${CDN_BUCKET}
      alias: origin
      region: ${CDN_REGION}
      endpoint: cos.${CDN_REGION}.myqcloud.com
      ofs: false
EOF

cos_put() {
  # cos_put <local-file> <remote-key>
  # Use the bucket ALIAS (origin) defined in the config so coscli resolves the
  # endpoint automatically. Accessing by raw bucket name would additionally
  # require an -e endpoint flag on every call.
  local src="$1" key="$2"
  "$COSCLI" -c "$CONF" cp "$src" "cos://origin/${key}" >/dev/null
  log "uploaded → ${key}"
}

# ── Upload versioned artifacts ────────────────────────────────────────────────

log "==> Uploading versioned artifacts (${VERSION})"
for f in "$DIST"/*.tar.gz "$DIST"/*.zip "$DIST"/checksums.txt; do
  [ -e "$f" ] || continue
  cos_put "$f" "${PREFIX}/${VERSION}/$(basename "$f")"
done

# ── Upload fixed-path files ───────────────────────────────────────────────────

log "==> Uploading fixed-path files"
cos_put "install.sh"      "${PREFIX}/install.sh"
cos_put "install.ps1"     "${PREFIX}/install.ps1"
cos_put "$DIST/latest.txt" "${PREFIX}/latest.txt"

# ── Refresh CDN cache for fixed-path files ────────────────────────────────────
#
# Versioned paths are immutable, so only the fixed-path files need purging.
log "==> Refreshing CDN cache"
TCCLI="$WORK/tccli-venv"
python3 -m venv "$TCCLI"
# shellcheck disable=SC1091
. "$TCCLI/bin/activate"
pip install --quiet tccli

export TCCLI_SECRET_ID="$CDN_SECRET_ID"
export TCCLI_SECRET_KEY="$CDN_SECRET_KEY"

# CDN is a global service; purge all fixed-path URLs in one call.
if tccli cdn PurgeUrlsCache \
    --cli-unfold-argument \
    --region "$CDN_REGION" \
    --secretId "$CDN_SECRET_ID" \
    --secretKey "$CDN_SECRET_KEY" \
    --Urls "https://${CDN_DOMAIN}/${PREFIX}/install.sh" \
           "https://${CDN_DOMAIN}/${PREFIX}/install.ps1" \
           "https://${CDN_DOMAIN}/${PREFIX}/latest.txt" >/dev/null 2>&1; then
  log "purged install.sh / install.ps1 / latest.txt"
else
  log "(cache purge failed or unavailable — files will refresh on TTL expiry)"
fi

log "==> CDN sync complete"
