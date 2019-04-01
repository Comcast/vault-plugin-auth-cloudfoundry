Development Tools
=================

This directory provides various resources that may be useful for local development and testing of the authentication plugin.

# Docker Compose

The `docker-compose.yml` file provides an entrypoint for a [Docker Compose](https://docs.docker.com/compose/) environment. When run, it:
 * Starts Vault within a container
 * Builds the Cloud Foundry authentication plugin within a container
 * Builds the authentication plugin and initializes it with Vault

Docker and Docker Compose must be first installed.

Once started, Vault will be exposed at `http://127.0.0.1:8200`.

#### Usage

Start environment:

```shell
docker-compose up
```

Rebuild and start environment:

```shell
docker-compose up --force-recreate --build
```

Example login:

```shell
export VAULT_ADDR='http://127.0.0.1:8200'
vault login root
```

# Certs

A Makefile is provided to assist with creating _development_ certificates, similar to those that Cloud Foundry would use.

Openssl is required and must first be installed before creating certificates.

**Note**: These certificates are only suitable for development, not production environments.

Once run, it will create a `./certs` directory containing the following:

```
certs/
├── instance-invalid.csr
├── instance-invalid.key
├── instance-invalid.key.pem       # Instance private key
├── instance-invalid.pem           # Instance certificate (with invalid key usage)
├── instance.csr
├── instance.key
├── instance.key.pem               # Instance private key
├── instance.pem                   # Instance certificate
├── rootCA.key
├── rootCA.pem                     # Root certificate
└── rootCA.srl
```

#### Usage

```shell
make certs
```

