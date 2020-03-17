package huaweicloud

import (
	"fmt"
	"strings"
)

type identity struct {
	Account string `json:"account"`
	User    string `json:"user"`
}

func (i *identity) equal(other *identity) bool {
	return i.Account == other.Account && i.User == other.User
}

func (i *identity) toString() string {
	return fmt.Sprintf("%s:%s", i.Account, i.User)
}

func newIdentity(i string) (*identity, error) {
	p := strings.Split(i, ":")
	if len(p) != 2 {
		return nil, fmt.Errorf("unrecognized identity: contains %d colon-separated fields, expected 2", len(p))
	}

	return &identity{Account: p[0], User: p[1]}, nil
}
