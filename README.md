# vault-plugin-auth-cloudfoundry
Vault authentication plugin for Cloud Foundry &amp; Spring

## Local Development

Prereqs:
 * [Vault](https://www.vaultproject.io/downloads.html)
 * [Go toolchain](https://golang.org/doc/install) (minimum version of 1.11)

### Usage

* Register plugin
  ```shell
  # Register plugin as `cf`
  vault plugin register \
  	-sha256="$(shasum -a 256 /tmp/vault/vault-plugin-auth-cloudfoundry |head -c64)" \
  	-command=vault-plugin-auth-cloudfoundry \
  	auth cf
  ```

* Enable plugin
  ```shell
  # Enable at path `/cf`
  vault auth enable -path=cf cf
  ```

* Configure root certificate authority
  ```shell
  vault write auth/cf/config ca=@path/to/ca.crt
  ```

* Configure org to policy mapping (optional)
  ```shell
  vault write auth/cf/map/organizations/5c6e25f1-921b-40f4-a652-9e8501ca0a6b value=$name-of-policy,$another-policy
  ```

* Configure space to policy mapping (optional)
  ```shell
  vault write auth/cf/map/spaces/5c6e25f1-921b-40f4-a652-9e8501ca0a6b value=$name-of-policy,$another-policy
  ```

* Login using CF instance certificate
  ```shell
  vault write auth/cf/login cert=@path/to/instance.crt
  ```

### Local Development

The plugin can be built and used locally. Note: This assumes the Go toolchain is available.

Compile plugin and save binary in temporary directory (this can be changed as long as it
matches the plugin directory used by Vault server):
```shell
GO111MODULE=on go build -o /tmp/vault/vault-plugin-auth-cloudfoundry
```

#### Running Vault server

For local development, Vault can be started in `dev` mode. This mode, as the name suggests,
is useful for local development since the service starts unsealed and with a pre-determined root
token. Note: The `vault` binary should be downloaded and available in `$PATH`.

Start Vault in dev mode using `local.hcl` config file:
```shell
VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200 vault server -dev -dev-root-token-id="root" -config=dev/vault/config/local.hcl
```

In another terminal, register and enable the plugin:
```shell
# Register plugin as `cf`
vault plugin register \
	-sha256="$(shasum -a 256 /tmp/vault/vault-plugin-auth-cloudfoundry |head -c64)" \
	-command=vault-plugin-auth-cloudfoundry \
	auth cf
# Enable at path `/cf`
vault auth enable -path=cf cf
```
