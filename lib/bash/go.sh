#!/usr/bin/env bash

# Environment Flags
set -o errexit  # Exit when a command fails
set -o pipefail # Catch mysqldump fails
set -o nounset  # Exit when using undeclared variables

dependency "go"

function goVersion() {

  local GO_MOD
  GO_MOD="./src/inkr.com/go.mod"

  grep "^go " <"${GO_MOD}" | sed "s|^go[^0-9]*||"
}

function pumpAppVersion() {

  local GIT_BRANCH
  GIT_BRANCH="$(git branch | grep "\*" | cut -d ' ' -f2)"

  local APP_VERSION=""
  if [[ "${GIT_BRANCH}" == "release/"* ]]; then
    APP_VERSION="${GIT_BRANCH//release\//}"
  elif [[ "${GIT_BRANCH}" == "hotfix/"* ]]; then
    APP_VERSION="${GIT_BRANCH//hotfix\//}"
  else
    return 0
  fi

  while true; do
    local DOTS_COUNT
    DOTS_COUNT="$(tr -dc '.' <<<"${APP_VERSION}" | awk '{ print length; }')"
    if ((DOTS_COUNT >= 2)); then
      break
    else
      APP_VERSION="${APP_VERSION}.0"
    fi
  done

  local MODULE_PATH
  MODULE_PATH="./src/inkr.com/go.mod"
  MODULE_PATH="${MODULE_PATH%/*}"

  local VERSION_FILE="${MODULE_PATH}/version/version.go"

  local CURRENT_APP_VERSION
  CURRENT_APP_VERSION="$(
    grep "var Version =" <"${VERSION_FILE}" |
      sed \
        -e "s|^.*= \"||" \
        -e "s|\".*||"
  )"

  if [[ "${CURRENT_APP_VERSION}" == "${APP_VERSION}" ]]; then
    return 0
  fi

  echo -e "${OK_COLOR}==> Pumping app version from '${CURRENT_APP_VERSION}' to '${APP_VERSION}'..."

  (
    set -x
    sed -i.bak -e "s|var Version = \".*\"|var Version = \"${APP_VERSION}\"|" "${VERSION_FILE}"
  )
  rm -rf "${VERSION_FILE}.bak"

  (
    set -x
    git add "${VERSION_FILE}"
    git commit --message "[Boilerplate] Pump version to ${APP_VERSION}"
  )

  echo
}
