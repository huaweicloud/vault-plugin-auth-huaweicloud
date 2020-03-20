package huaweicloud

import (
	"fmt"
)

type identity struct {
	User    string `json:"user"`
	Account string `json:"account"`
}

func (i *identity) equal(other *identity) bool {
	return i.Account == other.Account && i.User == other.User
}

func (i *identity) toString() string {
	return fmt.Sprintf("%s:%s", i.Account, i.User)
}
