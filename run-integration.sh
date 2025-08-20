#!/bin/bash
set -e

# run-tests.sh [ci|local]
#
# - ci:    Loads default + ci.override.env
# - local: Loads default + override.env

# Determine mode

mode="$1"
if [[ "$mode" != "ci" && "$mode" != "local" ]]; then
  echo "Must specify 'ci' or 'local' as first argument"
  exit 1
fi

# Determine which env files to read

ENV_FILES=("./default.env")

if [[ "$mode" == "ci" ]]; then
  echo "Loading CI overrides..."
  ENV_FILES+=("./ci.override.env")
else
  echo "Loading local overrides..."
  ENV_FILES+=("./override.env")
fi

# Load env files

for file in "${ENV_FILES[@]}"; do
  if [[ -f "$file" ]]; then
    echo "Loading env vars from $file"
    set -o allexport
    source "$file"
    set +o allexport
  else
    echo "Skipping missing optional env file: $file"
  fi
done

# Clone repos if CI

if [[ "$mode" == "ci" ]]; then
  echo "Cloning $USER_SERVICE_REPO branch $USER_SERVICE_BRANCH into $USER_SERVICE_PATH"
  git clone --branch "$USER_SERVICE_BRANCH" "$USER_SERVICE_REPO" "$USER_SERVICE_PATH"

  echo "Extracting the test private key for $USER_SERVICE_REPO"
  tar -xvzf "$USER_SERVICE_PATH/keys/keys.tar.gz" -C "$USER_SERVICE_PATH/keys"

  echo "Cloning $ROOM_SERVICE_REPO branch $ROOM_SERVICE_BRANCH into $ROOM_SERVICE_PATH"
  git clone --branch "$ROOM_SERVICE_BRANCH" "$ROOM_SERVICE_REPO" "$ROOM_SERVICE_PATH"
fi

# Run integration tests

docker compose -f ./compose.integration.yml up --build --abort-on-container-exit --exit-code-from test-runner
docker compose -f ./compose.integration.yml down