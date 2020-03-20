package huaweicloud

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/cidrutil"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/huaweicloud/golangsdk"
	huaweisdk "github.com/huaweicloud/golangsdk/openstack"
	"github.com/huaweicloud/golangsdk/openstack/identity/v3/tokens"
)

func pathLogin(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "login$",
		Fields: map[string]*framework.FieldSchema{
			"role": {
				Type: framework.TypeString,
				Description: `Name of the role against which the login is being attempted.
If a matching role is not found, login fails.`,
			},
			"token": {
				Type:        framework.TypeString,
				Description: "user's token against which to make the Huawei Cloud request.",
			},
		},
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathLoginUpdate,
		},
		HelpSynopsis:    pathLoginSyn,
		HelpDescription: pathLoginDesc,
	}
}

func (b *backend) pathLoginUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	token := ""
	if t, ok := data.GetOk("token"); ok {
		token = t.(string)
	} else {
		return nil, errors.New("missing token")
	}

	user, err := getTokenInfo(token)
	if err != nil {
		return nil, errwrap.Wrapf("error making upstream request: {{err}}", err)
	}

	roleName := ""
	if r, ok := data.GetOk("role"); ok {
		roleName = r.(string)
	} else {
		return nil, errors.New("missing role")
	}

	role, err := readRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("entry for role %s not found", roleName)
	}

	if len(role.TokenBoundCIDRs) > 0 {
		if req.Connection == nil {
			b.Logger().Warn("token bound CIDRs found but no connection information available for validation")

			return nil, logical.ErrPermissionDenied
		}

		if !cidrutil.RemoteAddrIsOk(req.Connection.RemoteAddr, role.TokenBoundCIDRs) {
			return nil, logical.ErrPermissionDenied
		}
	}

	identity := &identity{Account: user.Domain.Name, User: user.Name}
	if !identity.equal(&role.Identity) {
		return nil, errors.New("the caller's identity does not match the role's")
	}

	auth := &logical.Auth{
		Metadata: map[string]string{
			"account":   identity.Account,
			"user":      identity.User,
			"role_name": roleName,
		},
	}

	role.PopulateTokenAuth(auth)

	return &logical.Response{
		Auth: auth,
	}, nil
}

func (b *backend) pathLoginRenew(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	account := req.Auth.Metadata["account"]
	if account == "" {
		return nil, errors.New("unable to retrieve account from metadata during renewal")
	}
	user := req.Auth.Metadata["user"]
	if user == "" {
		return nil, errors.New("unable to retrieve user from metadata during renewal")
	}
	roleName := req.Auth.Metadata["role_name"]
	if roleName == "" {
		return nil, errors.New("unable to retrieve role name during renewal")
	}

	role, err := readRole(ctx, req.Storage, roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, fmt.Errorf("role entry(%s) is not found", roleName)
	}

	identity := &identity{Account: account, User: user}
	if !identity.equal(&role.Identity) {
		return nil, errors.New("the caller's identity does not match the role's")
	}

	resp := &logical.Response{Auth: req.Auth}
	resp.Auth.TTL = role.TokenTTL
	resp.Auth.MaxTTL = role.TokenMaxTTL
	resp.Auth.Period = role.TokenPeriod
	return resp, nil
}

func getTokenInfo(token string) (*tokens.User, error) {
	endpoint := "https://iam.myhwclouds.com:443/v3"
	client, err := huaweisdk.NewClient(endpoint)
	if err != nil {
		return nil, err
	}
	client.TokenID = token

	transport := cleanhttp.DefaultTransport()
	transport.TLSClientConfig = &tls.Config{}
	client.HTTPClient.Transport = transport

	v3Client, err := huaweisdk.NewIdentityV3(client, golangsdk.EndpointOpts{})
	if err != nil {
		return nil, err
	}

	return tokens.Get(v3Client, token).ExtractUser()
}

const pathLoginSyn = `
Authenticates a user with Vault.
`

const pathLoginDesc = `
Authenticate Huawei Cloud user using a token.

User token is authenticated by validating it on Huawei Cloud and then to see who signed the request.
`
