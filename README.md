# vault-plugin-auth-cloudfoundry
Vault authentication plugin for Cloud Foundry &amp; Spring

## Local Development

Prereqs:
 * [Vault](https://www.vaultproject.io/downloads.html)
 * [Go toolchain](https://golang.org/doc/install) (minimum version of 1.11)

### Usage

* Register plugin
  ```shell
  # Register plugin as `cloudfoundry`
  vault plugin register \
  	-sha256="$(shasum -a 256 /tmp/vault/vault-plugin-auth-cloudfoundry |head -c64)" \
  	-command=vault-plugin-auth-cloudfoundry \
  	auth cloudfoundry
  ```

* Enable plugin
  ```shell
  # Enable at path `/cloudfoundry`
  vault auth enable -path=cloudfoundry/ cloudfoundry
  ```

* Configure root certificate authority
  ```shell
  vault write auth/cloudfoundry/config ca=@path/to/ca.crt
  ```

* Configure org to policy mapping (optional). Replace `org-uuid` with actual UUID.
  ```shell
  vault write auth/cloudfoundry/map/organizations/org-uuid value=$name-of-policy,$another-policy
  ```

* Configure space to policy mapping (optional). Replace `space-uuid` with actual UUID.
  ```shell
  vault write auth/cloudfoundry/map/spaces/space-uuid value=$name-of-policy,$another-policy
  ```

* Built JWT token. A function is included under cmd/cf-jwt but can easily be via a method of your choice

  ```shell
  go run main.go -policies ReadOnlyPolicy,ReadWritePolicy -location /tmp/jwt
  
  ```


* Login using CF instance certificate
  ```shell
  vault write auth/cloudfoundry/login jwt=@path/to/jwt
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

In executing the below, you may see an error similar to

```shell
Error registering plugin cloudfoundry: Put https://127.0.0.1:8200/v1/sys/plugins/catalog/auth/cloudfoundry: http: server gave HTTP response to HTTPS client
```

You may need to set VAULT_ADDR:
```shell
export VAULT_ADDR='http://127.0.0.1:8200'
```



Start Vault in dev mode using `local.hcl` config file:
```shell
VAULT_DEV_LISTEN_ADDRESS=0.0.0.0:8200 vault server -dev -dev-root-token-id="root" -config=dev/vault/config/local.hcl
```

In another terminal, register and enable the plugin:
```shell
# Register plugin as `cloudfoundry`
vault plugin register \
	-sha256="$(shasum -a 256 /tmp/vault/vault-plugin-auth-cloudfoundry |head -c64)" \
	-command=vault-plugin-auth-cloudfoundry \
	auth cloudfoundry
# Enable at path `/cloudfoundry`
vault auth enable -path=cloudfoundry cloudfoundry
```

You will need instance crt and key files from a CF instance and then you will need to set two environment variables:

```shell
# Example to get the keys from an instance
# cf ssh <my_app> -c "cat /etc/cf-instance-credentials/instance.key" >/Users/myname/cf_certs/instance.key
# cf ssh <my_app> -c "cat /etc/cf-instance-credentials/instance.crt" >/Users/myname/cf_certs/instance.crt
#

export CF_INSTANCE_KEY="<Local/Path/To/instance.key>"
export CF_INSTANCE_CERT="<Local/Path/To/instance.crt>"
```

Once this is done, you can then run cmd/cf-jwtmain.go

```shell
# This currently builds a JWT token and installs it in /tmp in a file called jwt. This will be more sophisticated later.
# go run main.go <policy> 
# where policy is a previously configured policy in Vault for the org/space for the instance.
To get a list of command line arguments: 
> go run main.go 
  -location string
        Directory and filename for JWT token (stdout is false) (default "/tmp/jwt")
  -policies string
        Policy or (comma-delimited) Policies to include in JWT token
  -stdout
        Print to STDOUT (default is false)
exit status 1

# To print JWT to stdout
> go run main.go -policies ReadOnlyAccess -stdout
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c

# Note: Example JWT-only.
```

Once you have a local JWT token, you can then issue a login command. The below presupposes a JWT stored in a file called "jwt" in /tmp

```shell
vault write auth/cloudfoundry/login jwt=@/tmp/jwt

# Result Example
Key                    Value
---                    -----
token                  s.IizX6LDtSyHxFRJkTYK3K693
token_accessor         IkpuyLTSSgzwbOfBS4mIvIbi
token_duration         2h23m19s
token_renewable        true
token_policies         ["default" "read-only"]
identity_policies      []
policies               ["default" "read-only"]
token_meta_cf_app      e89f1899-7904-4804-8020-323c533bdbc6
token_meta_cf_org      5155a5ba-08b9-439b-a12a-bfcff4e180cc
token_meta_cf_space    0738a7cb-ca84-4e1f-9529-e88ac1f814c4


```