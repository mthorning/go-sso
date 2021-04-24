package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/mthorning/go-sso/types"
	"github.com/mthorning/go-sso/utils"
	"github.com/nu7hatch/gouuid"
	"time"
)

var secret = "mysupersecretsecret"

func encode(part []byte) string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(part)
}

func sign(str string, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(str))
	return h.Sum(nil)
}

func Create(user types.User) string {
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

	jsonPayload, err := json.Marshal(payload)
	utils.CheckErr(err)

	token := fmt.Sprintf("%s.%s", encode(jsonHeader), encode(jsonPayload))
	signature := encode(sign(token, []byte(secret)))

	return fmt.Sprintf("%s.%s", token, signature)
}
