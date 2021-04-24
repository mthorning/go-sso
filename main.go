package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
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

func checkErr(err error) {
	if err != nil {
		log.Fatal("Error", err)
	}
}

func main() {
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	payload := map[string]interface{}{
		"sub":   "1234567890",
		"name":  "Matt Thorning",
		"admin": true,
	}

	jsonHeader, err := json.Marshal(header)
	checkErr(err)

	jsonPayload, err := json.Marshal(payload)
	checkErr(err)

	token := fmt.Sprintf("%s.%s", encode(jsonHeader), encode(jsonPayload))
	signature := encode(sign(token, []byte(secret)))
	jwt := fmt.Sprintf("%s.%s", token, signature)

	fmt.Println(jwt)
}
