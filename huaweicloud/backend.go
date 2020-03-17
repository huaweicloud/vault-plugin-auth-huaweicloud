package huaweicloud

import (
	"context"

	"github.com/hashicorp/vault/sdk/framework"
	"github.com/hashicorp/vault/sdk/logical"
	"github.com/pkg/errors"
)

func Factory(ctx context.Context, conf *logical.BackendConfig) (logical.Backend, error) {
	b := newBackend()
	if err := b.Setup(ctx, conf); err != nil {
		return nil, errors.Wrapf(err, "failed to create factory")
	}
	return b, nil
}

func newBackend() *backend {
	b := &backend{}
	b.Backend = &framework.Backend{
		AuthRenew: b.pathLoginRenew,
		Help:      backendHelp,
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{
				"login",
			},
		},
		Paths: []*framework.Path{
			pathLogin(b),
			pathListRoles(b),
			pathRole(b),
		},
		BackendType: logical.TypeCredential,
	}
	return b
}

type backend struct {
	*framework.Backend
}

const backendHelp = `
That Huawei Cloud auth method allows entities to authenticate based on their
identity and pre-configured roles.
`
