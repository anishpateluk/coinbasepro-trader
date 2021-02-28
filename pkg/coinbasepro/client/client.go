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

const coinbaseProBaseurlKey = "COINBASE_PRO_BASEURL"
const coinbaseProKeyKey = "COINBASE_PRO_KEY"
const coinbaseProPassphraseKey = "COINBASE_PRO_PASSPHRASE"
const coinbaseProSecretKey = "COINBASE_PRO_SECRET"

const coinbaseProAccessKeyHeader = "CB-ACCESS-KEY"
const coinbaseProAccessSignatureHeader = "CB-ACCESS-SIGN"
const coinbaseProAccessTimestampHeader = "CB-ACCESS-TIMESTAMP"
const coinbaseProAccessPassphraseHeader = "CB-ACCESS-PASSPHRASE"

const acceptHeaderKey = "Accept"
const acceptHeaderValue = "application/json"
const contentTypeHeaderKey = "Content-Type"
const contentTypeHeaderValue = "Content-Type"

const UnsupportedHttpMethodErrorMessage = "supplied an unsupported or invalid http method"

const waitTimeOn429 = 300 * time.Millisecond

var allowedHttpMethods = map[string]bool{ "GET":true, "POST":true, "DELETE":true }

type Client struct {
	baseUrl string
	key string
	passphrase string
	secret string
	httpClient *http.Client
}

func NewClient() (*Client, error) {
	baseUrl := os.Getenv(coinbaseProBaseurlKey)
	key := os.Getenv(coinbaseProKeyKey)
	passphrase := os.Getenv(coinbaseProPassphraseKey)
	secret := os.Getenv(coinbaseProSecretKey)

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

	return NewClientWithOptions(baseUrl, key, passphrase, secret)
}

func NewClientWithOptions(baseUrl, key, passphrase, secret string) (*Client, error) {
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

	req.Header.Add(coinbaseProAccessKeyHeader, t.key)
	req.Header.Add(coinbaseProAccessPassphraseHeader, t.passphrase)
	req.Header.Add(coinbaseProAccessTimestampHeader, timestamp)
	req.Header.Add(coinbaseProAccessSignatureHeader, signature)

	req.Header.Add(contentTypeHeaderKey, contentTypeHeaderValue)
	req.Header.Add(acceptHeaderKey, acceptHeaderValue)

	return req, nil
}

func (t *Client) sendRequest(req *http.Request, maxRetriesOn429 int) (res *http.Response, err error) {
	if maxRetriesOn429 < 1 {
		maxRetriesOn429 = 1
	}

	for tries := 0; tries < maxRetriesOn429; tries++ {
		res, err = t.httpClient.Do(req)
		if err != nil {
			break
		}

		if res.StatusCode == http.StatusTooManyRequests {
			time.Sleep(waitTimeOn429)
			continue
		}

		break
	}

	return res, err
}

func (t *Client) parseJsonResponse(res *http.Response, result interface{}) (interface{}, error) {
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		apiError := ApiError{
			StatusCode: res.StatusCode,
		}
		decoder := json.NewDecoder(res.Body)
		if err := decoder.Decode(&apiError); err != nil {
			return nil, err
		}

		return nil, apiError
	}

	if res.ContentLength == 0 {
		return nil, nil
	}

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(result); err != nil {
		return nil, err
	}

	return result, nil
}

func (t *Client) executeRequest(httpMethod, requestPath string, requestBody interface{}, responseBody interface{}, maxRetiresOn429 int) (interface{}, error) {
	req, err := t.buildRequest(httpMethod, requestPath, requestBody)
	if err != nil {
		return nil, err
	}

	res, err := t.sendRequest(req, maxRetiresOn429)
	if err != nil {
		return nil, err
	}

	parsedResponse, err := t.parseJsonResponse(res, responseBody)
	if err != nil {
		return nil, err
	}

	return parsedResponse, nil
}