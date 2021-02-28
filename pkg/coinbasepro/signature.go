package coinbasepro

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

func createSignature(secret, timestamp, httpMethod, requestPath, jsonBody string) (string, error) {
	decodedSecret, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	hash := hmac.New(sha256.New, decodedSecret)
	signaturePayload := fmt.Sprintf("%s%s%s%s", timestamp, httpMethod, requestPath, jsonBody)
	_, err = hash.Write([]byte(signaturePayload))
	if err != nil {
		return "", err
	}

	signature := base64.StdEncoding.EncodeToString(hash.Sum(nil))
	return signature, nil
}