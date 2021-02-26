package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

func createTimestamp() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

const CoinbaseProBaseurlKey = "COINBASE_PRO_BASEURL"
const CoinbaseProKeyKey = "COINBASE_PRO_KEY"
const CoinbaseProPassphraseKey = "COINBASE_PRO_PASSPHRASE"
const CoinbaseProSecretKey = "COINBASE_PRO_SECRET"

const UnsupportedHttpMethodErrorMessage = "supplied an unsupported or invalid http method"

var allowedHttpMethods = map[string]bool{ "GET":true, "POST":true, "DELETE":true }

type Client struct {
	baseUrl string
	key string
	passphrase string
	secret string
}

func New() (*Client, error) {
	baseUrl := os.Getenv(CoinbaseProBaseurlKey)
	key := os.Getenv(CoinbaseProKeyKey)
	passphrase := os.Getenv(CoinbaseProPassphraseKey)
	secret := os.Getenv(CoinbaseProSecretKey)

	if baseUrl == "" {
		return nil, errors.New("missing COINBASE_PRO_BASEURL")
	}

	if key == "" {
		return nil, errors.New("missing COINBASE_PRO_KEY")
	}

	if passphrase == "" {
		return nil, errors.New("missing COINBASE_PRO_PASSPHRASE")
	}

	if secret == "" {
		return nil, errors.New("missing COINBASE_PRO_SECRET")
	}

	return NewWithOptions(baseUrl, key, passphrase, secret)
}

func NewWithOptions(baseUrl, key, passphrase, secret string) (*Client, error) {
	client := Client{
		baseUrl: baseUrl,
		key: key,
		passphrase: passphrase,
		secret: secret,
	}

	return &client, nil
}

func allowedHttpMethod(httpMethod string) bool {
	_, found := allowedHttpMethods[httpMethod]
	return found
}

func (t *Client) buildRequest(httpMethod, requestPath string, requestData interface{}) (*http.Request, error) {
	if !allowedHttpMethod(httpMethod) {
		return &http.Request{}, errors.New(UnsupportedHttpMethodErrorMessage)
	}

	fullUrl := fmt.Sprintf("%s%s", t.baseUrl, requestPath)
	var requestBody = bytes.NewReader(make([]byte, 0))

	if requestData != nil {
		data, err := json.Marshal(requestData)
		if err != nil {
			return &http.Request{}, err
		}

		requestBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(httpMethod, fullUrl, requestBody)

	if err != nil {
		return &http.Request{}, err
	}

	return req, nil
}