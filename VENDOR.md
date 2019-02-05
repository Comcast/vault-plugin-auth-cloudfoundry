Vendor
======

This saga is what was done in order for the project's dependencies
to work with the new `go mod` system.

## Background

At the time of writting, the project's main.go is identical to
the [`mock-plugin`](https://github.com/hashicorp/vault/blob/v1.0.2/logical/plugin/mock/mock-plugin/main.go) from the Vault repo.

It imports the following:

* `github.com/hashicorp/vault/helper/pluginutil`
* `github.com/hashicorp/vault/logical`
* `github.com/hashicorp/vault/logical/plugin`
* `github.com/hashicorp/vault/logical/plugin/mock`

The Vault project, at `v1.0.2`, does not support the new `go mod` system. Instead, it uses `govendor` as seen in
[this](https://github.com/hashicorp/vault/blob/v1.0.2/scripts/update_deps.sh) script used by the project.


## Create go.mod file

##### 1

Checkout the Vault repository at the [`v1.0.2` / `37a1dc9c477c1c68c022d2084550f25bf20cac33`](https://github.com/hashicorp/vault/tree/37a1dc9c477c1c68c022d2084550f25bf20cac33) version.
This corresponds to the released `v1.0.2` Vault server, eg:

```shell
$ vault -version
Vault v1.0.2 ('37a1dc9c477c1c68c022d2084550f25bf20cac33')
```

Ths project does have a `./vendor` directory with a [`vendor.json`](https://github.com/hashicorp/vault/blob/v1.0.2/vendor/vendor.json) file.
This vendor file points to a repo that no longer exists on github.com (`github.com/tyrannosaurus-becks/aliyun-oss-go-sdk`), which prevents `go mod`
from using this file to create its own `go.mod` file. Delete the single entry in `./vendor/vendor.json` that references this repository.

##### 2

Initialize repo as `go mod` repo:

```shell
$ GO111MODULE=on go mod init
go: creating new go.mod: module github.com/hashicorp/vault
go: copying requirements from vendor/vendor.json
go: converting vendor/vendor.json: stat appengine/cloudsql@: unrecognized import path "appengine/cloudsql" (import path does not begin with hostname)
go: converting vendor/vendor.json: stat appengine_internal@: unrecognized import path "appengine_internal" (import path does not begin with hostname)
go: converting vendor/vendor.json: stat appengine@: unrecognized import path "appengine" (import path does not begin with hostname)
go: converting vendor/vendor.json: stat appengine_internal/base@: unrecognized import path "appengine_internal/base" (import path does not begin with hostname)
```

Copy contents of `require` block in the resulting `go.mod` file.

##### 3

Switch to this project. Initialize project as `go mod` project

```shell
GO111MODULE=on go mod init
```

Paste contents from step #2 into this project's `go.mod` file, in the `require` block.

Remove the following require statements:
* `github.com/coreos/etcd`
* `github.com/google/go-github`

Add the following require statements:
* `github.com/hashicorp/vault v1.0.2`

##### 4

Remove unused deps by running tidy

```shell
GO111MODULE=on go mod tidy
```

The `go.mod` file should now be correct.
