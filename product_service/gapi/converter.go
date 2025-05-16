package gapi

import (
	"os/user"
	"product/protos"
)

func convertOrder(user *user.User) *protos.Order {
	return &protos.Order{}
}
