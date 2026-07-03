#!/usr/bin/env bash
# Mirror release artifacts to the CDN origin bucket and refresh edge caches.
#
# Publishes to cdn.link-ai.tech/cli/:
#   cli/<version>/*           versioned binaries, checksums, skill archive
#   cli/install.sh            latest install script (fixed path)
#   cli/install.ps1           latest Windows install script (fixed path)
#   cli/install.md            agent install guide (fixed path)
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
timeout 120 curl -fsSL -o "$COSCLI" "$COSCLI_URL"
chmod +x "$COSCLI"

# Talk to COS with flags only (--init-skip skips the interactive config; -i/-k/-e
# are then required). This avoids any config-file schema mismatch across coscli
# versions. CDN_BUCKET must be the full bucket name incl. APPID (e.g.
# name-1250000000).
ENDPOINT="cos.${CDN_REGION}.myqcloud.com"
COS_FLAGS=(--init-skip=true -i "$CDN_SECRET_ID" -k "$CDN_SECRET_KEY" -e "$ENDPOINT")

# Wrap every coscli call: inject COS_FLAGS (auth + --init-skip) into EVERY
# invocation — without them coscli treats each call as a first run and drops
# into its interactive "Welcome to coscli!" config prompt, which then fails on
# the /dev/null stdin. Also redirect stdin from /dev/null (so it can never block
# on a prompt) and time-box it. The timeout is generous (15 min per call)
# because GitHub's overseas runners upload to COS in mainland China very slowly
# (~10-20 KB/s), so a few-MB artifact can legitimately take minutes.
cos() { timeout 900 "$COSCLI" "$@" "${COS_FLAGS[@]}" </dev/null; }

cos_put() {
  # cos_put <local-file> <remote-key>
  local src="$1" key="$2"
  log "uploading → ${key}"
  # Discard coscli's per-chunk progress spam; keep stderr for real errors.
  cos cp "$src" "cos://${CDN_BUCKET}/${key}" >/dev/null
  log "uploaded  → ${key}"
}

# Diagnostics: print coscli version + verify the bucket is reachable before
# uploading, so a failure surfaces a clear cause instead of a bare exit 1.
log "==> coscli version"
cos --version || true
log "==> Verifying bucket access (endpoint=${ENDPOINT}, bucket=${CDN_BUCKET})"
cos ls "cos://${CDN_BUCKET}/" >/dev/null || {
  echo "ERROR: cannot access bucket ${CDN_BUCKET} at ${ENDPOINT}." >&2
  echo "Check CDN_BUCKET (must include APPID, e.g. name-1250000000), CDN_REGION, and the key's COS permissions." >&2
  exit 1
}

# ── Upload fixed-path files FIRST ─────────────────────────────────────────────
#
# These are small (a few KB each) and are the entry points shared directly with
# users/agents (install scripts, install guide, skill bundle, version pointer).
# Upload them before the multi-MB platform binaries so they become available
# almost immediately instead of waiting out the slow cross-border binary upload.

log "==> Uploading fixed-path files"
cos_put "install.sh"      "${PREFIX}/install.sh"
cos_put "install.ps1"     "${PREFIX}/install.ps1"
cos_put "$DIST/latest.txt" "${PREFIX}/latest.txt"
# Agent install guide at a fixed URL so it can be shared with an agent directly:
#   https://<CDN_DOMAIN>/cli/install.md
cos_put "skills/linkai-cli/references/install.md" "${PREFIX}/install.md"
# Skill bundle at a fixed URL so it can be shared with an agent directly:
#   https://<CDN_DOMAIN>/cli/linkai-cli-skill.zip
[ -e "$DIST/linkai-cli-skill.zip" ] && cos_put "$DIST/linkai-cli-skill.zip" "${PREFIX}/linkai-cli-skill.zip"

# ── Upload versioned artifacts ────────────────────────────────────────────────
#
# The large platform binaries / packages. Uploaded last because they are slow
# and are only fetched by version-pinned installs, not the shared entry points.

log "==> Uploading versioned artifacts (${VERSION})"
for f in "$DIST"/*.tar.gz "$DIST"/*.zip "$DIST"/checksums.txt; do
  [ -e "$f" ] || continue
  cos_put "$f" "${PREFIX}/${VERSION}/$(basename "$f")"
done

# ── Refresh CDN cache for fixed-path files ────────────────────────────────────
#
# Versioned paths are immutable, so only the fixed-path files need purging.
# Call the CDN PurgeUrlsCache API directly with curl + a TC3-HMAC-SHA256
# signature (openssl). This is a single fast HTTPS request — no Python/tccli to
# install. Best-effort: any failure just lets the files refresh on their TTL.
log "==> Refreshing CDN cache (best-effort)"

# hmac_sha256 <hex-or-string-key-mode> ... helpers
_sha256hex() { printf '%s' "$1" | openssl dgst -sha256 | sed 's/^.* //'; }
_hmac_key()  { printf '%s' "$2" | openssl dgst -sha256 -hmac "$1" | sed 's/^.* //'; }         # string key
_hmac_hex()  { printf '%s' "$3" | openssl dgst -sha256 -mac HMAC -macopt "hexkey:$2" | sed 's/^.* //'; } # hex key

purge_cache() {
  local host="cdn.tencentcloudapi.com" service="cdn" action="PurgeUrlsCache" version="2018-06-06"
  local ts date payload chash creq scope sts kDate kService kSigning sig auth resp
  ts="$(date +%s)"
  date="$(date -u -d "@$ts" +%Y-%m-%d 2>/dev/null || date -u -r "$ts" +%Y-%m-%d)"
  payload="{\"Urls\":[\"https://${CDN_DOMAIN}/${PREFIX}/install.sh\",\"https://${CDN_DOMAIN}/${PREFIX}/install.ps1\",\"https://${CDN_DOMAIN}/${PREFIX}/latest.txt\",\"https://${CDN_DOMAIN}/${PREFIX}/install.md\",\"https://${CDN_DOMAIN}/${PREFIX}/linkai-cli-skill.zip\"]}"

  chash="$(_sha256hex "$payload")"
  creq="POST
/

content-type:application/json; charset=utf-8
host:${host}
x-tc-action:$(printf '%s' "$action" | tr '[:upper:]' '[:lower:]')

content-type;host;x-tc-action
${chash}"
  scope="${date}/${service}/tc3_request"
  sts="TC3-HMAC-SHA256
${ts}
${scope}
$(_sha256hex "$creq")"

  kDate="$(_hmac_key "TC3${CDN_SECRET_KEY}" "$date")"
  kService="$(_hmac_hex '' "$kDate" "$service")"
  kSigning="$(_hmac_hex '' "$kService" 'tc3_request')"
  sig="$(_hmac_hex '' "$kSigning" "$sts")"
  auth="TC3-HMAC-SHA256 Credential=${CDN_SECRET_ID}/${scope}, SignedHeaders=content-type;host;x-tc-action, Signature=${sig}"

  resp="$(curl -fsS -m 30 -X POST "https://${host}/" \
    -H "Authorization: ${auth}" \
    -H "Content-Type: application/json; charset=utf-8" \
    -H "Host: ${host}" \
    -H "X-TC-Action: ${action}" \
    -H "X-TC-Timestamp: ${ts}" \
    -H "X-TC-Version: ${version}" \
    -d "$payload" 2>&1)" || { echo "purge request failed: $resp" >&2; return 1; }
  # A successful call has no top-level "Error"; anything else is a soft failure.
  case "$resp" in
    *'"Error"'*) echo "purge API error: $resp" >&2; return 1 ;;
  esac
  return 0
}
if purge_cache; then
  log "purged install.sh / install.ps1 / latest.txt / install.md / linkai-cli-skill.zip"
else
  log "(cache purge skipped/failed — files will refresh on TTL expiry)"
fi

log "==> CDN sync complete"
