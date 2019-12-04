#!/usr/bin/env bash

# Environment Flags
set -o errexit  # Exit when a command fails
set -o pipefail # Catch mysqldump fails
set -o nounset  # Exit when using undeclared variables

# Import
if [[ "${BOILERPLATE_CORE_IMPORTED:-}" == "true" ]]; then
  return
fi
BOILERPLATE_CORE_IMPORTED="true"

# Path
if [[ "${PATH}" != *"/usr/local/bin"* ]]; then
  export PATH="/usr/local/bin:${PATH}"
fi

if [[ "${PATH}" != *"${PWD}/.bin"* && -d "${PWD}/.bin" ]]; then
  export PATH="${PWD}/.bin:${PATH}"
fi

# Colors
if [[ "${BOILERPLATE_NO_COLOR:-}" != "true" ]]; then
  export FORCE_COLOR=1
  export TERM="xterm-256color"
fi

if [[ "${BOILERPLATE_NO_COLOR:-}" == "true" ]]; then

  export NO_COLOR=""
  export BOLD_COLOR=""
  export DIM_COLOR=""
  export UNDERLINED_COLOR=""
  export BLINK_COLOR=""
  export INVERTED_COLOR=""
  export HIDDEN_COLOR=""

  export DEFAULT_COLOR=""
  export BLACK_COLOR=""
  export RED_COLOR=""
  export GREEN_COLOR=""
  export YELLOW_COLOR=""
  export BLUE_COLOR=""
  export MAGENTA_COLOR=""
  export CYAN_COLOR=""
  export LIGHT_GRAY_COLOR=""
  export DARK_GRAY_COLOR=""
  export LIGHT_RED_COLOR=""
  export LIGHT_GREEN_COLOR=""
  export LIGHT_YELLO_COLOR=""
  export LIGHT_BLUE_COLOR=""
  export LIGHT_MAGENTA_COLOR=""
  export LIGHT_CYAN_COLOR=""
  export WHITE_COLOR=""

  export BG_DEFAULT_COLOR=""
  export BG_BLACK_COLOR=""
  export BG_RED_COLOR=""
  export BG_GREEN_COLOR=""
  export BG_YELLOW_COLOR=""
  export BG_BLUE_COLOR=""
  export BG_MAGENTA_COLOR=""
  export BG_CYAN_COLOR=""
  export BG_LIGHT_GRAY_COLOR=""
  export BG_DARK_GRAY_COLOR=""
  export BG_LIGHT_RED_COLOR=""
  export BG_LIGHT_GREEN_COLOR=""
  export BG_LIGHT_YELLOW_COLOR=""
  export BG_LIGHT_BLUE_COLOR=""
  export BG_LIGHT_MAGENTA_COLOR=""
  export BG_LIGHT_CYAN_COLOR=""
  export BG_WHITE_COLOR=""

else

  export NO_COLOR="\033[0m"
  export BOLD_COLOR="\033[1m"
  export DIM_COLOR="\033[2m"
  export UNDERLINED_COLOR="\033[4m"
  export BLINK_COLOR="\033[5m"
  export INVERTED_COLOR="\033[7m"
  export HIDDEN_COLOR="\033[8m"

  export DEFAULT_COLOR="\033[39m"
  export BLACK_COLOR="\033[30m"
  export RED_COLOR="\033[31m"
  export GREEN_COLOR="\033[32m"
  export YELLOW_COLOR="\033[33m"
  export BLUE_COLOR="\033[34m"
  export MAGENTA_COLOR="\033[35m"
  export CYAN_COLOR="\033[36m"
  export LIGHT_GRAY_COLOR="\033[37m"
  export DARK_GRAY_COLOR="\033[90m"
  export LIGHT_RED_COLOR="\033[91m"
  export LIGHT_GREEN_COLOR="\033[92m"
  export LIGHT_YELLO_COLOR="\033[93m"
  export LIGHT_BLUE_COLOR="\033[94m"
  export LIGHT_MAGENTA_COLOR="\033[95m"
  export LIGHT_CYAN_COLOR="\033[96m"
  export WHITE_COLOR="\033[97m"

  export BG_DEFAULT_COLOR="\033[49m"
  export BG_BLACK_COLOR="\033[40m"
  export BG_RED_COLOR="\033[41m"
  export BG_GREEN_COLOR="\033[42m"
  export BG_YELLOW_COLOR="\033[43m"
  export BG_BLUE_COLOR="\033[44m"
  export BG_MAGENTA_COLOR="\033[45m"
  export BG_CYAN_COLOR="\033[46m"
  export BG_LIGHT_GRAY_COLOR="\033[47m"
  export BG_DARK_GRAY_COLOR="\033[100m"
  export BG_LIGHT_RED_COLOR="\033[101m"
  export BG_LIGHT_GREEN_COLOR="\033[102m"
  export BG_LIGHT_YELLOW_COLOR="\033[103m"
  export BG_LIGHT_BLUE_COLOR="\033[104m"
  export BG_LIGHT_MAGENTA_COLOR="\033[105m"
  export BG_LIGHT_CYAN_COLOR="\033[106m"
  export BG_WHITE_COLOR="\033[107m"

fi

export OK_COLOR="${GREEN_COLOR}"
export ERROR_COLOR="${RED_COLOR}"
export WARN_COLOR="${LIGHT_RED_COLOR}"

# Output
function logFormat() {

  local LOCAL_BOILERPLATE_LOG_TIME="${BOILERPLATE_LOG_TIME:-}"
  local LOCAL_BOILERPLATE_NO_EMPTY_LINE="${BOILERPLATE_NO_EMPTY_LINE:-}"

  SCRIPT_PATH=""
  for BASH_SOURCE_ITEM in "${BASH_SOURCE[@]}"; do
    if [[ "${BASH_SOURCE_ITEM}" != "${BASH_SOURCE[0]}" ]]; then
      SCRIPT_PATH="${SCRIPT_PATH}${NO_COLOR}[${BLUE_COLOR}${BASH_SOURCE_ITEM##*/}${NO_COLOR}]"
    fi
  done

  local LINE
  while IFS='' read -r LINE; do

    if [[ "${LOCAL_BOILERPLATE_NO_EMPTY_LINE}" == "true" && "${LINE}" == "" ]]; then
      continue
    fi

    if [[ -n "${1:-}" && "${1:-}" == "--error" ]]; then

      if [[ "${LINE}" == [+]*" "* ]]; then
        echo
        echo -e "${BOLD_COLOR} ${LINE} ${NO_COLOR}"
        echo
        continue
      fi

      echo -e "${WARN_COLOR}STDERR${NO_COLOR} ${LINE} ${NO_COLOR}"
      continue

    fi

    # Go: Add "src/inkr.com/" to file paths to bring back "Cmd + click to follow link"
    LINE="$(
      echo "${LINE}" |
        sed -E 's| inkr.com/| src/inkr.com/|g'
    )"

    local PREFIX=""

    if [[ "${LOCAL_BOILERPLATE_LOG_TIME}" == "true" ]]; then
      local TIME
      TIME="$(
        date +"%Y-%m-%d %H:%M:%S %Z"
      )"
      PREFIX="${PREFIX}${DARK_GRAY_COLOR}${TIME} "
    fi

    PREFIX="${PREFIX}${SCRIPT_PATH} "

    echo -e "${PREFIX}${NO_COLOR} ${LINE} ${NO_COLOR}"

  done
}

function registerLogger() {
  echo
  exec 3>&1
  exec > >(logFormat)
  exec 2> >(logFormat --error)
}

registerLogger

# Dependencies
function dependency() {

  local DEPENDENCY_NAME="${1:-}"

  if ! command -v "${DEPENDENCY_NAME}" >/dev/null; then

    echo "Dependency \"${DEPENDENCY_NAME}\" not found."

    case "${DEPENDENCY_NAME}" in
    aws)
      if command -v "brew" >/dev/null; then
        (
          set -x
          brew install "awscli"
        )
        echo
      else
        echo "No installation script support for \"${DEPENDENCY_NAME}\"." >&2
        return 1
      fi
      ;;
    envkey-source)
      (
        set -x
        curl -s "https://raw.githubusercontent.com/envkey/envkey-source/master/install.sh" | bash
      )
      ;;
    go)
      if command -v "brew" >/dev/null; then
        (
          set -x
          brew install "go"
        )
        echo
      else
        echo "No installation script support for \"${DEPENDENCY_NAME}\"." >&2
        return 1
      fi
      ;;
    jq)
      if command -v "brew" >/dev/null; then
        (
          set -x
          brew install "jq"
        )
        echo
      else
        echo "No installation script support for \"${DEPENDENCY_NAME}\"." >&2
        return 1
      fi
      ;;
    shellcheck)
      if command -v "brew" >/dev/null; then
        (
          set -x
          brew install "shellcheck"
        )
        echo
      fi
      ;;
    *)
      echo "No installation script support for \"${DEPENDENCY_NAME}\"." >&2
      return 1
      ;;
    esac

    if ! command -v "${DEPENDENCY_NAME}" >/dev/null; then
      echo "Dependency \"${DEPENDENCY_NAME}\" not found after installing." >&2
      return 1
    fi

  fi
}

dependency "jq"

# Project
function projectKey() {
  echo "myawesomeproject"
}

# Git Hooks
if command -v "git" >/dev/null && [[ -d "./.git" ]]; then
  git config "core.hooksPath" ".githooks"
fi

function loadEnvKey() {

  echo
  echo "Loading EnvKey..."

  # Check for dependency early
  dependency "envkey-source"

  if [[ -z "${ENVKEY:-}" && -f "./.env" ]]; then
    while IFS='' read -r LINE; do
      if [[ "${LINE}" == *"="* && "${LINE}" != "#"* ]]; then
        export "${LINE?}"
      fi
    done <"./.env"
  fi

  if [[ -n "${ENVKEY:-}" ]]; then
    eval "$(envkey-source "${ENVKEY}")"
  fi
}
