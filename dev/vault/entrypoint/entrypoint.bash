#!/usr/bin/env bash
set -e

sleep 1

vault login "${VAULT_TOKEN}" >/dev/null 2>&1

# Register plugin(s)
for p in ${VAULT_PLUGIN_DIR}/*; do
	# Ensure plugin is executable
	if [[ -x "${p}" ]]; then
		pluginHash=$(sha256sum "${p}" | head -c 64)
		pluginCommand=$(basename "${p}")
		pluginName=${pluginCommand#vault-plugin-auth-}

		echo "Enabling plugin: ${pluginCommand}"

		set -x

		# Register plugin
		# https://www.vaultproject.io/api/system/plugins-catalog.html#register-plugin
		vault plugin register \
			-sha256="${pluginHash}" \
			-command="${pluginCommand}" \
			auth "${pluginName}"

		# Enable plugin
		# https://www.vaultproject.io/api/system/auth.html#enable-auth-method
		vault auth enable "${pluginName}"

		set +x
	fi
done

# Keep process alive
sleep 1024d
