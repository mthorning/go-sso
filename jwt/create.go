package jwt

import (
	"encoding/json"
	"fmt"
	"github.com/mthorning/go-sso/config"
	"github.com/mthorning/go-sso/types"
	"github.com/nu7hatch/gouuid"
	"time"
)

type Config struct {
	Secret string `default:"devsecret"`
}

var Conf Config

func init() {
	config.SetConfig(&Conf)
}

func New(user types.User) (string, error) {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}

	payload := map[string]interface{}{
		"jti":   u.String(),
		"iat":   time.Now(),
		"name":  user.Name,
		"email": user.Email,
		"admin": user.Admin,
	}

	jsonHeader, err := json.Marshal(header)
	if err != nil {
		return "", err
	}

	h := encode(jsonHeader)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	p := encode(jsonPayload)

	s := createSignature(h, p)

	return fmt.Sprintf("%s.%s.%s", h, p, s), nil
}
