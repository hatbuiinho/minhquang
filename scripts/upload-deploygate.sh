#!/bin/sh
set -eu

ENV_FILE="${ENV_FILE:-.env}"
APK_PATH="${DEPLOYGATE_APK_PATH:-fe/android/app/build/outputs/apk/release/app-release.apk}"

. "$(dirname "$0")/load-env.sh"

require_env() {
	name="$1"
	eval "value=\${$name:-}"
	if [ -z "$value" ]; then
		printf '%s\n' "Missing required env: $name" >&2
		exit 1
	fi
}

load_env_file

require_env DEPLOYGATE_API_TOKEN
require_env DEPLOYGATE_OWNER_NAME

if [ ! -f "$APK_PATH" ]; then
	printf '%s\n' "APK not found: $APK_PATH" >&2
	printf '%s\n' "Run: make android-release" >&2
	exit 1
fi

if printf '%s' "$APK_PATH" | grep -q 'unsigned'; then
	printf '%s\n' "Refusing to upload unsigned APK: $APK_PATH" >&2
	exit 1
fi

message="${DEPLOYGATE_MESSAGE:-Android release}"
release_note="${DEPLOYGATE_RELEASE_NOTE:-$message}"
api_url="https://deploygate.com/api/users/${DEPLOYGATE_OWNER_NAME}/apps"

set -- \
	--fail-with-body \
	--silent \
	--show-error \
	--url "$api_url" \
	-H "Authorization: Bearer ${DEPLOYGATE_API_TOKEN}" \
	-X POST \
	-F "file=@${APK_PATH}" \
	--form-string "message=${message}" \
	--form-string "release_note=${release_note}"

if [ -n "${DEPLOYGATE_DISTRIBUTION_KEY:-}" ]; then
	set -- "$@" --form-string "distribution_key=${DEPLOYGATE_DISTRIBUTION_KEY}"
elif [ -n "${DEPLOYGATE_DISTRIBUTION_NAME:-}" ]; then
	set -- "$@" --form-string "distribution_name=${DEPLOYGATE_DISTRIBUTION_NAME}"
fi

curl "$@"
printf '\n'
