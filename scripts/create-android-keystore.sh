#!/usr/bin/env bash
set -euo pipefail

keystore_file="${ANDROID_KEYSTORE_FILE:-fe/android/app/release.keystore}"
key_alias="${ANDROID_KEY_ALIAS:-minhquang-release}"

if ! command -v keytool >/dev/null 2>&1; then
	printf '%s\n' "keytool is required. Install or select a JDK before creating the keystore." >&2
	exit 1
fi

if [ -e "$keystore_file" ]; then
	printf '%s\n' "Keystore already exists: $keystore_file" >&2
	exit 1
fi

read_secret() {
	prompt="$1"
	value=""
	confirm=""

	while [ -z "$value" ]; do
		printf '%s' "$prompt" >&2
		read -r -s value
		printf '\n' >&2
	done

	printf '%s' "Confirm $prompt" >&2
	read -r -s confirm
	printf '\n' >&2

	if [ "$value" != "$confirm" ]; then
		printf '%s\n' "Passwords do not match." >&2
		exit 1
	fi

	printf '%s' "$value"
}

store_password="$(read_secret "Keystore password: ")"
key_password="$(read_secret "Key password: ")"

mkdir -p "$(dirname "$keystore_file")"

keytool -genkeypair \
	-v \
	-keystore "$keystore_file" \
	-storepass "$store_password" \
	-keypass "$key_password" \
	-alias "$key_alias" \
	-keyalg RSA \
	-keysize 2048 \
	-validity 10000 \
	-dname "CN=Minh Quang, OU=Mobile, O=Thien Vien Minh Quang, L=Ho Chi Minh, ST=Ho Chi Minh, C=VN"

printf '\n%s\n' "Keystore created: $keystore_file"
printf '%s\n' "Add these values to .env:"
printf 'ANDROID_KEYSTORE_FILE=%s\n' "${keystore_file#fe/android/app/}"
printf 'ANDROID_KEYSTORE_PASSWORD=%s\n' '<your-keystore-password>'
printf 'ANDROID_KEY_ALIAS=%s\n' "$key_alias"
printf 'ANDROID_KEY_PASSWORD=%s\n' '<your-key-password>'
