# Vault Plugin: HuaweiCloud Auth Backend

This is a standalone backend plugin for use with [Hashicorp Vault](https://www.github.com/hashicorp/vault).
This plugin allows authentication to Vault using user's personal token of Huawei Cloud.

## Getting Started

This is a [Vault plugin](https://www.vaultproject.io/docs/internals/plugins.html)
and is meant to work with Vault. This guide assumes you have already installed Vault
and have a basic understanding of how Vault works.

Otherwise, first read this guide on how to [get started with Vault](https://www.vaultproject.io/intro/getting-started/install.html).

To learn specifically about how plugins work, see documentation on [Vault plugins](https://www.vaultproject.io/docs/internals/plugins.html).

## Security Model

This authentication model places Vault in the middle of a call between a client and Huawei Cloud's [api](https://support.huaweicloud.com/api-iam/iam_30_0004.html). Based on its response, it grants an access token based on pre-configured roles.

## Auth Flow

The basic mechanism of operation is per-role.

Roles are associated with a Huawei Cloud account and user. When logining to Vault, it matches the account and user name retrived from token with that of a pre-created role in Vault. It then checks what policies have been associated with the role, and grants a token accordingly.

## Usage

This guide assumes some familiarity with Vault and Vault's plugin
ecosystem. You must have a Vault server already running, unsealed, and
authenticated.

- Download and decompress the latest plugin binary from the Releases tab on
GitHub. Alternatively you can compile the plugin from source, if you're into
that kind a thing.

- Move the compiled plugin into Vault's configured `plugin_directory`.

  ```sh
  $ mv vault-plugin-auth-huaweicloud /etc/vault/plugins/
  ```

- Calculate the SHA256 of the plugin and register it in Vault's plugin catalog.
If you are downloading the pre-compiled binary, it is highly recommended that
you use the published checksums to verify integrity.

  ```sh
  $ export SHA256=$(shasum -a 256 "/etc/vault/plugins/vault-plugin-auth-huaweicloud" | cut -d' ' -f1)

  $ vault write sys/plugins/catalog/auth-hw \
      sha_256="${SHA256}" \
      command="vault-plugin-auth-huaweicloud"
  ```

- Mount the auth method.

  ```sh
  $ vault auth enable auth-hw
  ```

- Create role.

  ```sh
  $ vault write auth/auth-hw/role/dev-role \
      account="${account}" \
      user="user"
  ```

  - `access_token` - _(required)_ oauth access token for the Slack application.
    This comes from Slack when you install the application into your team.
    This is used to communicate with Slack's API on your application's behalf.

  - `teams` - _(required)_ comma-separated list of names or IDs of the teams
    (workspaces) for which to allow authentication. Slack is currently in the
    process of renaming "teams" to "workspaces", and it's confusing. We
    apologize. Team names and IDs are case sensitive.

  - `allow_bot_users` - _(default: false)_ allow bots to use their tokens to
    authenticate. By default, bots are not allowed to authenticate.

  - `allow_non_2fa` - _(default: true)_ allow users which do not have 2FA/MFA
    enabled on their Slack account to authenticate. By default, users must have
    2FA enabled on their Slack account to authenticate to Vault. Users must
    still be mapped to an appropriate policy to receive a token.

  - `allow_restricted_users` - _(default: false)_ allow multi-channel guests to
    authenticate. By default, restricted users will not be given a token, even
    if they are mapped to policies.

  - `allow_ultra_restricted_users` - _(default: false)_ allow single-channel
    guests to authenticate. By default, restricted users will not be given a
    token, even if they are mapped to policies.

  - `anyone_policies` - _(default: "")_ comma-separated list of policies to
    apply to everyone. If set, **any Slack member** will be able to authenticate
    to Vault and receive a token with these policies. By default, users must be
    a member of a group, usergroup, or mapped directly.

  Additionally, you can tune the TTLs:

  - `ttl` - minimum TTL for tokens created from this authentication.

  - `max_ttl` - maximum TTL for tokens created from this authentication.

- Login to Vault.
  Because the user's personal token of Huawei Cloud is very long, so it
  recommends to save the token in a file first, then pass it to vault.

  ```sh
  # save token to ./token.txt
  $ token=$(cat ./token.txt); vault write auth/hw/login role=dev-role token=$token
  ```

  The response will be a standard auth response with some token metadata:

  ```text
  Key                     Value
  ---                     -----
  token                   s.bmCw3arLhilGd0BWwOEEQ4X0
  token_accessor          6aN0hE5BPRNnnuv1uCD6BJAC
  token_duration          768h
  token_renewable         true
  token_policies          ["default"]
  identity_policies       []
  policies                ["default"]
  token_meta_account      account
  token_meta_role_name    dev-role
  token_meta_user         user
  ```
