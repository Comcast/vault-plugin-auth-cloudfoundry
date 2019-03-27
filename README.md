vault-plugin-auth-cloudfoundry
==============================

This is a standalone backend authentication plugin for use with [HashiCorp Vault](https://www.github.com/hashicorp/vault).
This plugin allows for applications running in Cloudfoundry using [Instance Identity](https://docs.cloudfoundry.org/adminguide/instance-identity.html) to authenticate with Vault.

## Background

#### CF

Cloud Foundry can be enabled to provide instance credentials to its apps ([Enabling Instance Identity](https://docs.cloudfoundry.org/adminguide/instance-identity.html)).
Once enabled, Cloud Foundry injects into each app's file system a unique x.509 certificate and RSA private key (referenced by `$CF_INSTANCE_CERT` and `$CF_INSTANCE_KEY` respectively).
This key-pair is updated regularly and automatically by Cloud Foundry, with a relatively short certificate TTL (eg 24 hours).

The public certificate contains fields identifying the app's:
 * CF Org GUID
 * CF Space GUID
 * CF Instance GUID

More information: [Using Instance Identity Credentials](https://docs.cloudfoundry.org/devguide/deploy-apps/instance-identity.html).

#### Vault

Vault provides a mechanism to create custom authentication plugins. These plugins typically authenticate users through use of cryptography and an established form of trust.
The plugin runs as a separate process on the same host as the main Vault process. Communication is via gRPC.

More information: [Vault: Building Plugin Backends](https://learn.hashicorp.com/vault/developer/plugin-backends).

A common difficulty when using Vault is how to bootstrap applications with the required credentials needed to authenticate. This can introduce complexity and security concerns
into CI/CD pipelines and/or other deployment mechanisms, since they then become responsible for distribution of credentials.

## Design

A Vault authentication plugin that would allow Cloud Foundry apps to authenticate using the Cloudy Foundry Instance Identity system. Apps would not need to be bootstrapped with
credentials. Instead, they would use the certificates provided by Cloud Foundry to authenticate themselves and gain access to Vault secrets.

![Auth Flow](/docs/vault-plugin-auth-cloudfoundry.png)

#### Authentication Flow

1) A Vault admin enables the Cloud Foundry Vault authentication plugin, and configures it to trust the Cloud Foundry certificate authority. This is a one-time configuration step.

2) The Cloud Foundy app generates a [JWT](https://jwt.io/introduction/) token. The token contains the app's public certificate using the [`x5c`](https://tools.ietf.org/html/rfc7515#section-4.1.6) field. The token
is signed using the app's private key. The JWT token is signed using the app's CF key pair.

3) The app makes an authentication request to Vault. Included in the request is the JWT token created in the previous step.

4) Vault recognizes the request as a Cloud Foundry authentication request and delegates to the plugin. The plugin processes the request and validates the following:
    * App's certificate was signed by the configured CA
    * App's certificate has not expired
    * App's certificate has proper key usages set
    * App's certificate contains the expected Cloud Foundry specific attributes (Org, Space, and Instance GUIDs)
    * The JWT token was signed using the app's public/private key pair

5) Once the plugin validates the provided JWT, it looks up the Vault policies to assign to the user. This can be done a number of ways.
In either case, the plugin is able to perform the lookup using its knowledge about the request:
    * The app includes a `role` in its request. A Vault admin can then create a mapping of roles to policies. The roles can be scoped to Cloud Foundry Orgs, Spaces, or Instances
    * A Vault admin creates a mapping of Cloudy Foundry Orgs, Spaces, and Instances to Vault policies

6) The plugin informs Vault that it approves or denies the token request. When approved, the plugin provides the Vault policies from the previous step and a sensible TTL, likely based on the certificate's TTL.

4) Vault then creates an appropriately scoped token and returns it to the app. The app now uses the token to read and write secrets.

# License

Licensed under the Apache License, Version 2.0: http://www.apache.org/licenses/LICENSE-2.0
