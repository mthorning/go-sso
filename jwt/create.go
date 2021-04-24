package jwt

import (
	"encoding/json"
	"fmt"
	"github.com/mthorning/go-sso/config"
	"github.com/mthorning/go-sso/types"
	"github.com/mthorning/go-sso/utils"
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

func New(user types.User) string {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	u, err := uuid.NewV4()
	utils.CheckErr(err)

	payload := map[string]interface{}{
		"jti":   u.String(),
		"iat":   time.Now(),
		"name":  user.Name,
		"email": user.Email,
		"admin": user.Admin,
	}

	jsonHeader, err := json.Marshal(header)
	utils.CheckErr(err)
	h := encode(jsonHeader)

	jsonPayload, err := json.Marshal(payload)
	utils.CheckErr(err)
	p := encode(jsonPayload)

	s := createSignature(h, p)

	return fmt.Sprintf("%s.%s.%s", h, p, s)
}
