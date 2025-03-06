#!/bin/sh
# shellcheck shell=dash
#
# Licensed under the MIT license
# <LICENSE-MIT or https://opensource.org/licenses/MIT>, at your
# option. This file may not be copied, modified, or distributed
# except according to those terms.

set -u

APP_NAME="infragpt"
UV_VERSION="0.5.9"
INSTALLER_BASE_URL="${UV_INSTALLER_GITHUB_BASE_URL:-https://github.com}"
ARTIFACT_DOWNLOAD_URL="${INSTALLER_BASE_URL}/astral-sh/uv/releases/download/${UV_VERSION}"
PRINT_VERBOSE=${INSTALLER_PRINT_VERBOSE:-0}
PRINT_QUIET=${INSTALLER_PRINT_QUIET:-0}
NO_MODIFY_PATH=${INSTALLER_NO_MODIFY_PATH:-0}
INSTALL_UPDATER=1
UNMANAGED_INSTALL="${INFRAGPT_UNMANAGED_INSTALL:-}"
PYTHON_VERSION="python3.12"

if [ -n "${UNMANAGED_INSTALL}" ]; then
    NO_MODIFY_PATH=1
    INSTALL_UPDATER=0
fi

usage() {
    cat <<EOF
infragpt-installer.sh

The installer for InfraGPT

This script installs InfraGPT using the uv package manager.
It first downloads uv if needed and then uses it to install InfraGPT.

USAGE:
    infragpt-installer.sh [OPTIONS]

OPTIONS:
    -v, --verbose
            Enable verbose output

    -q, --quiet
            Disable progress output

    -p, --python VERSION
            Specify Python version to use (default: ${PYTHON_VERSION})

        --no-modify-path
            Don't configure the PATH environment variable

    -h, --help
            Print help information
EOF
}

main() {
    local _python_version="${PYTHON_VERSION}"

    for arg in "$@"; do
        case "$arg" in
            --help | -h)
                usage
                exit 0
                ;;
            --quiet | -q)
                PRINT_QUIET=1
                ;;
            --verbose | -v)
                PRINT_VERBOSE=1
                ;;
            --python=* | -p=*)
                _python_version="${arg#*=}"
                ;;
            --python | -p)
                if [ -n "${2:-}" ]; then
                    _python_version="$2"
                    shift
                else
                    err "option --python requires an argument"
                fi
                ;;
            --no-modify-path)
                NO_MODIFY_PATH=1
                ;;
            *)
                if [ "${arg%%--*}" = "" ]; then
                    err "unknown option $arg"
                fi
                ;;
        esac
    done

    # Check if uv is already installed
    if check_cmd uv; then
        say "uv is already installed, using existing installation"
        install_infragpt "${_python_version}"
    else
        say "uv not found, downloading and installing uv first"
        install_uv
        install_infragpt "${_python_version}"
    fi

    say ""
    say "InfraGPT installation completed successfully!"
    say "You can now run 'infragpt' to use it."
}

install_uv() {
    say "Downloading uv installer..."
    local _installer_url="${ARTIFACT_DOWNLOAD_URL}/uv-installer.sh"
    local _tmp_file
    _tmp_file="$(mktemp)"

    if ! downloader "$_installer_url" "$_tmp_file"; then
        say "Failed to download uv installer."
        say "Check your network connection and try again."
        exit 1
    fi

    say "Installing uv..."
    chmod +x "$_tmp_file"

    # Run the uv installer
    sh "$_tmp_file"
    local _result=$?

    # Clean up
    rm -f "$_tmp_file"

    if [ $_result -ne 0 ]; then
        err "Failed to install uv. Please check the error messages above."
    fi

    say "uv installed successfully!"
}

install_infragpt() {
    local _python_version="$1"

    say "Installing InfraGPT using uv..."
    say "Using Python version: ${_python_version}"

    if ! uv tool install --force --python "${_python_version}" infragpt@latest; then
        err "Failed to install InfraGPT. Please check the error messages above."
    fi

    say "InfraGPT installed successfully!"
}

say() {
    if [ "0" = "$PRINT_QUIET" ]; then
        echo "$1"
    fi
}

say_verbose() {
    if [ "1" = "$PRINT_VERBOSE" ]; then
        echo "$1"
    fi
}

err() {
    if [ "0" = "$PRINT_QUIET" ]; then
        local red
        local reset
        red=$(tput setaf 1 2>/dev/null || echo '')
        reset=$(tput sgr0 2>/dev/null || echo '')
        say "${red}ERROR${reset}: $1" >&2
    fi
    exit 1
}

check_cmd() {
    command -v "$1" > /dev/null 2>&1
    return $?
}

# This wraps curl or wget. Try curl first, if not installed,
# use wget instead.
downloader() {
    local _dld

    if check_cmd curl; then
        _dld=curl
    elif check_cmd wget; then
        _dld=wget
    else
        err "Need curl or wget to download files"
    fi

    if [ "$_dld" = curl ]; then
        curl -sSfL "$1" -o "$2"
    elif [ "$_dld" = wget ]; then
        wget "$1" -O "$2"
    else
        err "Unknown downloader"  # should not reach here
    fi
}

main "$@" || exit 1
