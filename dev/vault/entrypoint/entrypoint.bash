#!/usr/bin/env bash
set -e

sleep 1

# Initialize server, storing resulting JSON
vaultInit=$(vault operator init \
	-format=json \
	-key-shares=1 \
	-key-threshold=1)

# Extract the unseal key and root token from `operator init` output
unsealKey=$(echo "${vaultInit}" | jq -r ".unseal_keys_b64[0]")
rootToken=$(echo "${vaultInit}" | jq -r ".root_token")

# Normal "vault" commands will use this as their token
export VAULT_TOKEN="${rootToken}"

# Use unseal key to unseal server.
# This is required to make server usable.
vault operator unseal "${unsealKey}"

# Print root token to user
echo "================="
echo "Vault Server unsealed"
echo "ROOT TOKEN: \"${rootToken}\""
echo "UI: http://localhost:8200/ui"
echo "================="

# Register plugin(s)
pluginHash=$(sha256sum ${VAULT_PLUGIN_DIR}/mock-plugin | head -c 64)
vault write sys/plugins/catalog/mock-plugin \
	sha_256=${pluginHash} \
	command=mock-plugin

# Enable plugin(s)
vault secrets enable \
	-path=mock-plugin \
	-plugin-name=mock-plugin \
	plugin

# Keep process alive
sleep 1024d
