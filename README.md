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

  - `role` `(string: <required>)` - Name of the role.

  - `account` `(string)` - Name of Huawei Cloud account.

  - `user` `(string)` - Name of Huawei Cloud user.

  - `token_ttl` `(integer: 0 or string: "")` - The incremental lifetime for
    generated tokens. This current value of this will be referenced at renewal
    time.

  - `token_max_ttl` `(integer: 0 or string: "")` - The maximum lifetime for
    generated tokens. This current value of this will be referenced at renewal
    time.

  - `token_policies` `(array: [] or comma-delimited string: "")` - List of
    policies to encode onto generated tokens. Depending on the auth method, this
    list may be supplemented by user/group/other values.

  - `token_bound_cidrs` `(array: [] or comma-delimited string: "")` - List of
    CIDR blocks; if set, specifies blocks of IP addresses which can authenticate
    successfully, and ties the resulting token to these blocks as well.

  - `token_explicit_max_ttl` `(integer: 0 or string: "")` - If set, will encode
    an [explicit max TTL](/docs/concepts/tokens#token-time-to-live-periodic-tokens-and-explicit-max-ttls)
    onto the token. This is a hard cap even if `token_ttl` and `token_max_ttl`
    would otherwise allow a renewal.

  - `token_no_default_policy` `(bool: false)` - If set, the `default` policy will
    not be set on generated tokens; otherwise it will be added to the policies set
    in `token_policies`.

  - `token_num_uses` `(integer: 0)` - The maximum number of times a generated
    token may be used (within its lifetime); 0 means unlimited.

  - `token_period` `(integer: 0 or string: "")` - The
    [period](/docs/concepts/tokens#token-time-to-live-periodic-tokens-and-explicit-max-ttls),
    if any, to set on the token.

  - `token_type` `(string: "")` - The type of token that should be generated. Can
    be `service`, `batch`, or `default` to use the mount's tuned default (which
    unless changed will be `service` tokens). For token store roles, there are two
    additional possibilities: `default-service` and `default-batch` which specify
    the type to return unless the client requests a different type at generation
    time.

- Login to Vault.

  ```sh
  # It recommends saving token to a file(./token.txt), because token's length is very long.

  $ token=$(cat ./token.txt); vault write auth/auth-hw/login role=dev-role token=$token
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
