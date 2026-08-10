#!/bin/sh
set -eu

. "$(dirname "$0")/load-env.sh"
load_env_file

require_env() {
	name="$1"
	eval "value=\${$name:-}"
	if [ -z "$value" ]; then
		printf '%s\n' "Missing required env: $name" >&2
		exit 1
	fi
}

require_env ANDROID_KEYSTORE_FILE
require_env ANDROID_KEYSTORE_PASSWORD
require_env ANDROID_KEY_ALIAS
require_env ANDROID_KEY_PASSWORD

if [ -z "${VITE_API_BASE_URL:-}" ] && [ -n "${PUBLIC_API_BASE_URL:-}" ]; then
	export VITE_API_BASE_URL="$PUBLIC_API_BASE_URL"
fi

require_env VITE_API_BASE_URL

cd fe
yarn build
npx cap sync android
cd android
./gradlew assembleRelease
