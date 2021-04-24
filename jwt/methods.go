package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

func encode(part []byte) string {
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(part)
}

func createSignature(h, p string) string {
	hp := fmt.Sprintf("%s.%s", h, p)
	hash := hmac.New(sha256.New, []byte(Conf.Secret))
	hash.Write([]byte(hp))
	return encode(hash.Sum(nil))
}

func verifySignature(token string) bool {
	hps := strings.Split(token, ".")
	s := createSignature(hps[0], hps[1])
	return hps[2] == s
}
