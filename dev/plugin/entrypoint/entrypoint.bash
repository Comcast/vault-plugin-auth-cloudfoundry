#!/usr/bin/env bash
set -e

# Copy build artificats to plugin directory. This is necessary
# since the plugin directory is a mounted volume
find ${BUILD_DIR} -executable -type f | while read p; do
	cp "${p}" ${VAULT_PLUGIN_DIR}/.
done

vault login "${VAULT_TOKEN}" >/dev/null 2>&1

if [ -z "${VAULT_PLUGIN_DIR}" ]; then
	echo "\$VAULT_PLUGIN_DIR is not defined"
	exit 1
fi
echo "Registering plugins in \"${VAULT_PLUGIN_DIR}\""

# Register plugin(s)
find ${VAULT_PLUGIN_DIR} -executable -type f | while read p; do
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
echo "Entering eternal slumber..."
sleep 1024d
