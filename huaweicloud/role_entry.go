package huaweicloud

import (
	"github.com/hashicorp/vault/sdk/helper/tokenutil"
)

type roleEntry struct {
	tokenutil.TokenParams

	RoleName string   `json:"role_name"`
	Identity identity `json:"identity"`
}

func (r *roleEntry) ToResponseData() map[string]interface{} {
	d := map[string]interface{}{
		"role_name": r.RoleName,
		"identity":  r.Identity.toString(),
	}

	r.PopulateTokenData(d)

	return d
}
