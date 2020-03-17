package huaweicloud

import (
	"github.com/hashicorp/vault/sdk/helper/tokenutil"
)

type roleEntry struct {
	tokenutil.TokenParams

	Identity identity `json:"identity"`
	RoleName string `json:"role_name"`
}

func (r *roleEntry) ToResponseData() map[string]interface{} {
	d := map[string]interface{}{
		"identity": r.Identity.toString(),
		"role_name": r.RoleName,
	}

	r.PopulateTokenData(d)

	return d
}
