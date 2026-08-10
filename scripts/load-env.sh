load_env_file() {
	env_file="${ENV_FILE:-.env}"
	if [ ! -f "$env_file" ]; then
		return
	fi

	while IFS= read -r line || [ -n "$line" ]; do
		case "$line" in
			''|\#*) continue ;;
		esac

		key=${line%%=*}
		value=${line#*=}
		if [ -n "$key" ] && [ "$key" != "$line" ]; then
			export "$key=$value"
		fi
	done < "$env_file"
}
