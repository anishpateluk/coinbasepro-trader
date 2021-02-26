package client

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
)

var clientOptionsMap = map[string]string {
	CoinbaseProBaseurlKey:    "https://testbaseurl.com",
	CoinbaseProKeyKey:        "testKey",
	CoinbaseProPassphraseKey: "testPassphrase",
	CoinbaseProSecretKey:     "YmFzZTY0c2VjcmV0",
}

func resetEnvVars() {
	for key, value := range clientOptionsMap {
		os.Setenv(key, value)
	}
}

func isStringNumeric(str string) bool {
	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

func TestNew(t *testing.T) {
	t.Run("does not error when all environment variables exist", func(t *testing.T) {
		resetEnvVars()

		_, err := New()

		assert.Assert(t, is.Nil(err), "errored when all environment variables have values", err)
	})

	for envVar, _ := range clientOptionsMap {
		testCase := fmt.Sprintf("errors when %s environment variable is missing", envVar)
		t.Run(testCase, func(t *testing.T) {
			resetEnvVars()

			os.Setenv(envVar, "")

			_, err := New()

			assert.Error(t, err, fmt.Sprintf("missing %s", envVar))
		})
	}
}

func TestNewWithOptions(t *testing.T) {
	t.Run("options should override environment variables", func(t *testing.T) {
		resetEnvVars()

		client, err := NewWithOptions("https://hello.com", "KlBsW", "Password1", "zzGfSK=")
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewWithOptions", err)

		baseUrlEnvVar := clientOptionsMap[CoinbaseProBaseurlKey]
		assert.Assert(t, client.baseUrl != baseUrlEnvVar, fmt.Sprintf("wanted %s, got %s", client.baseUrl, baseUrlEnvVar))

		keyEnvVar := clientOptionsMap[CoinbaseProKeyKey]
		assert.Assert(t, client.key != keyEnvVar, "wanted %s got %s", client.key, keyEnvVar)

		passphraseEnvVar := clientOptionsMap[CoinbaseProPassphraseKey]
		assert.Assert(t, client.passphrase != passphraseEnvVar, "wanted %s got %s", client.passphrase, passphraseEnvVar)

		secretEnvVar := clientOptionsMap[CoinbaseProSecretKey]
		assert.Assert(t, client.secret != secretEnvVar,"wanted %s got %s", client.secret, secretEnvVar)
	})
}

func TestBuildRequest(t *testing.T) {
	resetEnvVars()

	t.Run("should error when unsupported httpMethod supplied", func(t *testing.T) {
		client, err := New()
		assert.Assert(t, is.Nil(err), "unexpected error creating client using New", err)

		for _, unsupportedHttpMethod := range []string {"get", "post", "not a http method" } {
			_, err = client.buildRequest(unsupportedHttpMethod, "/test", nil)
			assert.Error(t, err, UnsupportedHttpMethodErrorMessage, err)
		}

		for _, allowedHttpMethod := range []string { "GET", "POST", "DELETE" } {
			_, err = client.buildRequest(allowedHttpMethod, "/test", nil)

			assert.Assert(t, is.Nil(err), "unexpected error when supplying supported http methods", err)
		}
	})

	t.Run("should create expected request body", func(t *testing.T) {
		client, err := New()
		assert.Assert(t, is.Nil(err), "unexpected creating client using New", err)

		req, err := client.buildRequest("GET", "/test", nil)
		assert.Assert(t, is.Nil(err), "unexpected error from client.buildRequest", err)

		decoder := json.NewDecoder(req.Body)
		err = decoder.Decode(nil)
		assert.Error(t, err, "EOF", "expected EOF that represents an empty http request body", err)


		type testStruct struct {
			Foo string
		}

		requestData := testStruct{Foo: "bar"}

		req, err = client.buildRequest("GET", "/test", requestData)
		assert.Assert(t, is.Nil(err), "unexpected error from client.buildRequest", err)

		var decodedRequestBody = testStruct{}
		decoder = json.NewDecoder(req.Body)
		err = decoder.Decode(&decodedRequestBody)
		assert.Assert(t, is.Nil(err), "unexpected error decoding json", err)

		assert.DeepEqual(t, requestData, decodedRequestBody)
	})

	t.Run("should build request with configured base url and supplied path", func(t *testing.T) {
		expectedUrl := "https://testbaseurl.com/test"

		client, err := New()
		assert.Assert(t, is.Nil(err), "unexpected creating client using New", err)

		req, err := client.buildRequest("GET", "/test", nil)
		assert.Assert(t, is.Nil(err), "unexpected error from client.buildRequest", err)

		fullRequestUrl := fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.Path)
		assert.Equal(t, fullRequestUrl, expectedUrl)
	})

	t.Run("should build request with configured CoinbasePro Access http headers", func(t *testing.T) {
		expectedKeyHeader := clientOptionsMap[CoinbaseProKeyKey]
		expectedPassphraseHeader := clientOptionsMap[CoinbaseProPassphraseKey]

		client, err := New()
		assert.Assert(t, is.Nil(err), "unexpected creating client using New", err)

		req, err := client.buildRequest("GET", "/test", nil)
		assert.Assert(t, is.Nil(err), "unexpected error from client.buildRequest", err)

		keyHeader := req.Header.Get(CoinbaseProAccessKeyHeader)
		signatureHeader := req.Header.Get(CoinbaseProAccessSignatureHeader)
		timestampHeader := req.Header.Get(CoinbaseProAccessTimestampHeader)
		passphraseHeader := req.Header.Get(CoinbaseProAccessPassphraseHeader)
		assert.Equal(t, keyHeader, expectedKeyHeader)
		assert.Assert(t, signatureHeader != "" && len(signatureHeader) > 10 && !isStringNumeric(signatureHeader), "signature header error", signatureHeader)
		assert.Assert(t, timestampHeader != "" && len(timestampHeader) == 10 && isStringNumeric(timestampHeader), "timestamp header error", timestampHeader)
		assert.Equal(t, passphraseHeader, expectedPassphraseHeader)
	})
}