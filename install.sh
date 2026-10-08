#!/bin/sh
# One-line installer for the opencode-console-notify plugin (Linux).
#
# Copies plugin/console-notify.js into the OpenCode plugins directory.
# Linux has no notification identity registry: notify-send addresses the
# app by name, so there is nothing to register (mirror of install.ps1,
# minus the AUMID half).
#
# One-line install from GitHub:
#
#     curl -fsSL https://raw.githubusercontent.com/IbatoLionDev/opencode-console-notify/main/install.sh | sh
#
# Local checkout (offline / development, supports all flags):
#
#     ./install.sh [--plugins-dir <path>] [--uninstall] [--upgrade]
#
# When the script runs from a checkout containing plugin/console-notify.js
# the local file is used and nothing is downloaded. Otherwise the plugin
# source is fetched from the URL above. --upgrade always downloads fresh
# bytes, skipping the local file. With --uninstall, the plugin file is
# removed instead and --upgrade is ignored.
#
# POSIX sh only (dash-safe): no bashisms, no arrays.

set -eu

PLUGINS_DIR="${HOME}/.config/opencode/plugins"
UNINSTALL=0
UPGRADE=0

usage() {
    echo "usage: install.sh [--plugins-dir <path>] [--uninstall] [--upgrade]" >&2
}

while [ $# -gt 0 ]; do
    case "$1" in
        --plugins-dir)
            [ $# -ge 2 ] || { usage; exit 2; }
            PLUGINS_DIR="$2"
            shift 2
            ;;
        --plugins-dir=*)
            PLUGINS_DIR="${1#--plugins-dir=}"
            shift
            ;;
        --uninstall)
            UNINSTALL=1
            shift
            ;;
        --upgrade)
            UPGRADE=1
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            usage
            exit 2
            ;;
    esac
done

if [ -z "${PLUGINS_DIR}" ]; then
    echo "PluginsDir must not be empty." >&2
    exit 2
fi

PLUGIN_NAME="console-notify.js"
BASE_URL="https://raw.githubusercontent.com/IbatoLionDev/opencode-console-notify/main"

# ------------------------------------------------------------------
# Uninstall
# ------------------------------------------------------------------
if [ "${UNINSTALL}" -eq 1 ]; then
    DEST="${PLUGINS_DIR}/${PLUGIN_NAME}"
    if [ -f "${DEST}" ]; then
        rm -f "${DEST}"
        echo "Removed plugin file: ${DEST}"
    else
        echo "Plugin file not present, nothing to remove: ${DEST}"
    fi
    echo "No identity marker for OpenCode.Notifier (no registry on Linux), nothing to remove"
    echo "Uninstall complete. Restart OpenCode to unload the plugin."
    exit 0
fi

# ------------------------------------------------------------------
# Install
# ------------------------------------------------------------------
BYTES_FILE=""
SOURCE_DESCRIPTION=""

# Local checkout / dev path: uses the file on disk, no network.
# --upgrade skips this branch so the install always uses fresh bytes.
# curl | sh has no script directory of its own: $0 is "sh".
case "$0" in
    *install.sh)
        SCRIPT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
        LOCAL_PLUGIN="${SCRIPT_DIR}/plugin/${PLUGIN_NAME}"
        if [ "${UPGRADE}" -eq 0 ] && [ -f "${LOCAL_PLUGIN}" ]; then
            BYTES_FILE="${LOCAL_PLUGIN}"
            SOURCE_DESCRIPTION="local file ${LOCAL_PLUGIN} (no download)"
        fi
        ;;
esac

if [ -z "${BYTES_FILE}" ]; then
    URL="${BASE_URL}/plugin/${PLUGIN_NAME}"
    TMP_DOWNLOAD="$(mktemp)"
    if command -v curl >/dev/null 2>&1; then
        if ! curl -fsSL --proto '=https' --user-agent 'opencode-console-notify/install.sh' -o "${TMP_DOWNLOAD}" "${URL}"; then
            rm -f "${TMP_DOWNLOAD}"
            echo "Failed to download the plugin source from ${URL}" >&2
            exit 1
        fi
    elif command -v wget >/dev/null 2>&1; then
        if ! wget -q -O "${TMP_DOWNLOAD}" --user-agent='opencode-console-notify/install.sh' "${URL}"; then
            rm -f "${TMP_DOWNLOAD}"
            echo "Failed to download the plugin source from ${URL}" >&2
            exit 1
        fi
    else
        rm -f "${TMP_DOWNLOAD}"
        echo "Need curl or wget to download the plugin source from ${URL}" >&2
        exit 1
    fi
    BYTES_FILE="${TMP_DOWNLOAD}"
    SOURCE_DESCRIPTION="downloaded from ${URL}"
fi

mkdir -p "${PLUGINS_DIR}"

DEST="${PLUGINS_DIR}/${PLUGIN_NAME}"
TEMP_FILE="$(mktemp "${PLUGINS_DIR}/${PLUGIN_NAME}.XXXXXX.tmp")"
# Best-effort cleanup: after a successful rename the temp path no longer
# exists and rm is a harmless no-op (same for the download temp file).
trap 'rm -f "${TEMP_FILE:-}" "${TMP_DOWNLOAD:-}"' EXIT
# Write to a temp file in the same directory, then rename into place,
# so a running OpenCode never sees a half-written plugin file.
cp "${BYTES_FILE}" "${TEMP_FILE}"
mv -f "${TEMP_FILE}" "${DEST}"

if command -v sha256sum >/dev/null 2>&1; then
    HASH="$(sha256sum "${DEST}" | cut -d' ' -f1)"
elif command -v shasum >/dev/null 2>&1; then
    HASH="$(shasum -a 256 "${DEST}" | cut -d' ' -f1)"
else
    HASH="(sha256sum unavailable)"
fi

echo "Installed: ${DEST}"
echo "SHA256:    ${HASH}"
echo "Source:    ${SOURCE_DESCRIPTION}"
echo "Identity:  app name OpenCode.Notifier (no registry on Linux)"
echo "Restart OpenCode to load the plugin."
