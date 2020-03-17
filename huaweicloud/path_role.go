package huaweicloud

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/errwrap"
	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/helper/tokenutil"
	"github.com/hashicorp/vault/sdk/logical"
)

func pathRole(b *backend) *framework.Path {
	p := &framework.Path{
		Pattern: "role/" + framework.GenericNameRegex("role"),
		Fields: map[string]*framework.FieldSchema{
			"role": {
				Type:        framework.TypeLowerCaseString,
				Description: "The name of the role as it should appear in Vault.",
			},
			"identity": {
				Type:        framework.TypeString,
				Description: "The user identity to bind to this role.",
			},
		},
		ExistenceCheck: b.operationRoleExistenceCheck,
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.CreateOperation: b.operationRoleCreate,
			logical.UpdateOperation: b.operationRoleUpdate,
			logical.ReadOperation:   b.operationRoleRead,
			logical.DeleteOperation: b.operationRoleDelete,
		},
		HelpSynopsis:    pathRoleSyn,
		HelpDescription: pathRoleDesc,
	}

	tokenutil.AddTokenFields(p.Fields)
	return p
}

func pathListRoles(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "roles/?",
		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ListOperation: b.operationRoleList,
		},
		HelpSynopsis:    pathListRolesHelpSyn,
		HelpDescription: pathListRolesHelpDesc,
	}
}

func (b *backend) operationRoleExistenceCheck(ctx context.Context, req *logical.Request, data *framework.FieldData) (bool, error) {
	entry, err := readRole(ctx, req.Storage, data.Get("role").(string))
	if err != nil {
		return false, err
	}
	return entry != nil, nil
}

func (b *backend) operationRoleCreate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
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
	if role != nil {
		return nil, fmt.Errorf("role %s is already exist", roleName)
	}

	if _, ok := data.GetOk("identity"); !ok {
		return nil, errors.New("the identity is required to create a role")
	}

	role = &roleEntry{RoleName: roleName}
	return b.createUpdate(role, ctx, req, data)
}

func (b *backend) operationRoleUpdate(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
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
		return nil, fmt.Errorf("no role %s found to update", roleName)
	}

	return b.createUpdate(role, ctx, req, data)
}

func (b *backend) createUpdate(role *roleEntry, ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if i, ok := data.GetOk("identity"); ok {
		identity, err := newIdentity(i.(string))
		if err != nil {
			return nil, errwrap.Wrapf(fmt.Sprintf("unable to parse identity %s: {{err}}", i), err)
		}
		role.Identity = *identity
	}

	// Get tokenutil fields
	if err := role.ParseTokenFields(req, data); err != nil {
		return logical.ErrorResponse(err.Error()), logical.ErrInvalidRequest
	}

	if role.TokenMaxTTL > 0 && role.TokenTTL > role.TokenMaxTTL {
		return nil, errors.New("ttl exceeds max ttl")
	}

	entry, err := logical.StorageEntryJSON("role/"+role.RoleName, role)
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(ctx, entry); err != nil {
		return nil, err
	}

	if role.TokenTTL > b.System().MaxLeaseTTL() {
		resp := &logical.Response{}
		resp.AddWarning(fmt.Sprintf("ttl of %d exceeds the system max ttl of %d, the latter will be used during login",
			role.TokenTTL, b.System().MaxLeaseTTL()))
		return resp, nil
	}
	return nil, nil
}

func (b *backend) operationRoleRead(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	role, err := readRole(ctx, req.Storage, data.Get("role").(string))
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, nil
	}
	return &logical.Response{
		Data: role.ToResponseData(),
	}, nil
}

func (b *backend) operationRoleDelete(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	if err := req.Storage.Delete(ctx, "role/"+data.Get("role").(string)); err != nil {
		return nil, err
	}
	return nil, nil
}

func (b *backend) operationRoleList(ctx context.Context, req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	roleNames, err := req.Storage.List(ctx, "role/")
	if err != nil {
		return nil, err
	}
	return logical.ListResponse(roleNames), nil
}

func readRole(ctx context.Context, s logical.Storage, roleName string) (*roleEntry, error) {
	role, err := s.Get(ctx, "role/"+roleName)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, nil
	}
	result := &roleEntry{}
	if err := role.DecodeJSON(result); err != nil {
		return nil, err
	}

	return result, nil
}

const pathRoleSyn = `
Create a role and associate policies to it.
`

const pathRoleDesc = `
A precondition for login is that a role should be created in the backend.
The login endpoint takes in the role name against which the instance
should be validated. After authenticating the instance, the authorization
for the instance to access Vault's resources is determined by the policies
that are associated to the role though this endpoint.

Also, a 'max_ttl' can be configured in this endpoint that determines the maximum
duration for which a login can be renewed. Note that the 'max_ttl' has an upper
limit of the 'max_ttl' value on the backend's mount. The same applies to the 'ttl'.
`

const pathListRolesHelpSyn = `
Lists all the roles that are registered with Vault.
`

const pathListRolesHelpDesc = `
Roles will be listed by their respective role names.
`
