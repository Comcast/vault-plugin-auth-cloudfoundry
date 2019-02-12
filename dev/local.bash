#!/usr/bin/env bash

# Enable job control
set -m
set -e

GO111MODULE=on go build \
	-o /tmp/vault/vault-plugin-auth-cloudfoundry

export VAULT_ADDR=http://localhost:8200

# Start Vault in background
VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200 vault server \
	-dev \
	-dev-root-token-id="root" \
	-config=dev/vault/config/local.hcl &

vault plugin register \
	-sha256="$(shasum -a 256 /tmp/vault/vault-plugin-auth-cloudfoundry |head -c64)" \
	-command=vault-plugin-auth-cloudfoundry \
	auth cf

vault auth enable \
	-path=cf \
	cf

# Resume Vault process in foreground
fg %1
