package jwt

import (
	"errors"
	"github.com/mthorning/go-sso/types"
)

func Authenticate(token string) (types.User, error) {
	if !verifySignature(token) {
		return types.User{}, errors.New("Invalid signature")
	}
	return decodeUser(token)
}
