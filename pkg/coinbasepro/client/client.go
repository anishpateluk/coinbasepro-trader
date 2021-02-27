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

const CoinbaseProAccessKeyHeader = "CB-ACCESS-KEY"
const CoinbaseProAccessSignatureHeader = "CB-ACCESS-SIGN"
const CoinbaseProAccessTimestampHeader = "CB-ACCESS-TIMESTAMP"
const CoinbaseProAccessPassphraseHeader = "CB-ACCESS-PASSPHRASE"

const UnsupportedHttpMethodErrorMessage = "supplied an unsupported or invalid http method"

var allowedHttpMethods = map[string]bool{ "GET":true, "POST":true, "DELETE":true }

type Client struct {
	baseUrl string
	key string
	passphrase string
	secret string
	httpClient *http.Client
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
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	return &client, nil
}

func allowedHttpMethod(httpMethod string) bool {
	_, found := allowedHttpMethods[httpMethod]
	return found
}

func (t *Client) buildRequest(httpMethod, requestPath string, requestData interface{}) (req *http.Request, err error) {
	if !allowedHttpMethod(httpMethod) {
		return &http.Request{}, errors.New(UnsupportedHttpMethodErrorMessage)
	}

	fullUrl := fmt.Sprintf("%s%s", t.baseUrl, requestPath)
	var jsonBytes = make([]byte, 0)
	var requestBody = bytes.NewReader(jsonBytes)

	if requestData != nil {
		jsonBytes, err = json.Marshal(requestData)
		if err != nil {
			return &http.Request{}, err
		}

		requestBody = bytes.NewReader(jsonBytes)
	}

	req, err = http.NewRequest(httpMethod, fullUrl, requestBody)
	if err != nil {
		return &http.Request{}, err
	}

	timestamp := createTimestamp()
	signature, err := createSignature(t.secret, timestamp, httpMethod, requestPath, string(jsonBytes))
	if err != nil {
		return &http.Request{}, err
	}

	req.Header.Add(CoinbaseProAccessKeyHeader, t.key)
	req.Header.Add(CoinbaseProAccessPassphraseHeader, t.passphrase)
	req.Header.Add(CoinbaseProAccessTimestampHeader, timestamp)
	req.Header.Add(CoinbaseProAccessSignatureHeader, signature)

	return req, nil
}

func (t *Client) makeRequest(req *http.Request) (*http.Response, error) {
	res, err := t.httpClient.Do(req)
	return res, err
}