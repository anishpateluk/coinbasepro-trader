package coinbasepro

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

const testBaseUrl = "https://testbaseurl.com"
const testKey = "testKey"
const testPassphrase = "testPassphrase"
const testSecret = "YmFzZTY0c2VjcmV0"

var clientOptionsMap = map[string]string {
	coinbaseProBaseurlKey:    testBaseUrl,
	coinbaseProKeyKey:        testKey,
	coinbaseProPassphraseKey: testPassphrase,
	coinbaseProSecretKey:     testSecret,
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

		_, err := NewClient()

		assert.Assert(t, is.Nil(err), "errored when all environment variables have values", err)
	})

	for envVar, _ := range clientOptionsMap {
		testCase := fmt.Sprintf("errors when %s environment variable is missing", envVar)
		t.Run(testCase, func(t *testing.T) {
			resetEnvVars()

			os.Setenv(envVar, "")

			_, err := NewClient()

			assert.Error(t, err, fmt.Sprintf("missing %s", envVar))
		})
	}

	t.Run("should initialise http.Client", func(t *testing.T) {
		resetEnvVars()

		client, err := NewClient()
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClient", err)

		assert.Assert(t, client.httpClient != nil)
		assert.Equal(t, client.httpClient.Timeout, 10 * time.Second)
	})
}

func TestNewWithOptions(t *testing.T) {
	t.Run("options should override environment variables", func(t *testing.T) {
		resetEnvVars()

		client, err := NewClientWithOptions("https://hello.com", "KlBsW", "Password1", "zzGfSK=")
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClientWithOptions", err)

		baseUrlEnvVar := clientOptionsMap[coinbaseProBaseurlKey]
		assert.Assert(t, client.baseUrl != baseUrlEnvVar, fmt.Sprintf("wanted %s, got %s", client.baseUrl, baseUrlEnvVar))

		keyEnvVar := clientOptionsMap[coinbaseProKeyKey]
		assert.Assert(t, client.key != keyEnvVar, "wanted %s got %s", client.key, keyEnvVar)

		passphraseEnvVar := clientOptionsMap[coinbaseProPassphraseKey]
		assert.Assert(t, client.passphrase != passphraseEnvVar, "wanted %s got %s", client.passphrase, passphraseEnvVar)

		secretEnvVar := clientOptionsMap[coinbaseProSecretKey]
		assert.Assert(t, client.secret != secretEnvVar,"wanted %s got %s", client.secret, secretEnvVar)
	})

	t.Run("should initialise http.Client", func(t *testing.T) {
		resetEnvVars()

		client, err := NewClientWithOptions("https://hello.com", "KlBsW", "Password1", "zzGfSK=")
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClientWithOptions", err)

		assert.Assert(t, client.httpClient != nil)
		assert.Equal(t, client.httpClient.Timeout, 10 * time.Second)
	})
}

func TestBuildRequest(t *testing.T) {
	resetEnvVars()

	t.Run("should error when unsupported httpMethod supplied", func(t *testing.T) {
		client, err := NewClient()
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClient", err)

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
		client, err := NewClient()
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClient", err)

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

		client, err := NewClient()
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClient", err)

		req, err := client.buildRequest("GET", "/test", nil)
		assert.Assert(t, is.Nil(err), "unexpected error from client.buildRequest", err)

		fullRequestUrl := fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.Path)
		assert.Equal(t, fullRequestUrl, expectedUrl)
	})

	t.Run("should build request with configured CoinbasePro Access http headers", func(t *testing.T) {
		expectedKeyHeader := clientOptionsMap[coinbaseProKeyKey]
		expectedPassphraseHeader := clientOptionsMap[coinbaseProPassphraseKey]

		client, err := NewClient()
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClient", err)

		req, err := client.buildRequest("GET", "/test", nil)
		assert.Assert(t, is.Nil(err), "unexpected error from client.buildRequest", err)

		keyHeader := req.Header.Get(coinbaseProAccessKeyHeader)
		signatureHeader := req.Header.Get(coinbaseProAccessSignatureHeader)
		timestampHeader := req.Header.Get(coinbaseProAccessTimestampHeader)
		passphraseHeader := req.Header.Get(coinbaseProAccessPassphraseHeader)
		assert.Equal(t, keyHeader, expectedKeyHeader)
		assert.Assert(t, signatureHeader != "" && len(signatureHeader) > 10 && !isStringNumeric(signatureHeader), "signature header error", signatureHeader)
		assert.Assert(t, timestampHeader != "" && len(timestampHeader) == 10 && isStringNumeric(timestampHeader), "timestamp header error", timestampHeader)
		assert.Equal(t, passphraseHeader, expectedPassphraseHeader)
	})

	t.Run("should build request with correct accept and content type headers", func(t *testing.T) {
		client, err := NewClient()
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClient", err)

		req, err := client.buildRequest("GET", "/test", nil)
		assert.Assert(t, is.Nil(err), "unexpected error from client.buildRequest", err)

		acceptsHeader := req.Header.Get(acceptHeaderKey)
		contentTypeHeader := req.Header.Get(contentTypeHeaderKey)
		assert.Equal(t, acceptsHeader, acceptHeaderValue)
		assert.Equal(t, contentTypeHeader, contentTypeHeaderValue)
	})
}

func TestMakeRequest(t *testing.T) {
	resetEnvVars()

	for _, httpStatus := range []int {http.StatusOK, http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError} {
		testCase := fmt.Sprintf("should return %v http response", httpStatus)
		t.Run(testCase, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.WriteHeader(httpStatus)
			}))
			defer ts.Close()

			client, err := NewClientWithOptions(ts.URL, testKey, testPassphrase, testSecret)
			assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClientWithOptions", err)

			req, err := client.buildRequest("GET", "/test", nil)
			assert.Assert(t, is.Nil(err), "unexpected error from client.buildRequest", err)

			res, err := client.sendRequest(req, 0)
			assert.Assert(t, is.Nil(err), "unexpected error from client.sendRequest", err)

			assert.Equal(t, res.StatusCode, httpStatus)
		})
	}

	t.Run("should retry specified number of times when rate limited by 429", func(t *testing.T) {
		requests := 0
		maxRetriesOn429 := 3
		ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requests += 1
			writer.WriteHeader(http.StatusTooManyRequests)
		}))
		defer ts.Close()

		client, err := NewClientWithOptions(ts.URL, testKey, testPassphrase, testSecret)
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClientWithOptions", err)

		req, err := client.buildRequest("GET", "/test", nil)
		assert.Assert(t, is.Nil(err), "unexpected error from client.buildRequest", err)

		_, err = client.sendRequest(req, maxRetriesOn429)
		assert.Assert(t, is.Nil(err), "unexpected error from client.sendRequest", err)

		assert.Equal(t, requests, maxRetriesOn429)
	})
}

func TestParseResponse(t *testing.T) {
	resetEnvVars()

	t.Run("should return nil, nil when no body in Ok http response", func(t *testing.T) {
		client, err := NewClient()
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClient", err)

		res := http.Response{
			StatusCode: http.StatusOK,
			Body: ioutil.NopCloser(bytes.NewReader(make([]byte, 0))),
		}

		obj, err := client.parseJsonResponse(&res, nil)
		assert.Assert(t, is.Nil(err), "unexpected error from client.parseJsonResponse", err)

		assert.Assert(t, is.Nil(obj), "parsed body not nil when expected to be")
	})

	t.Run("should return parsed response body from Ok http response", func(t *testing.T) {
		client, err := NewClient()
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClient", err)

		type testResult struct {
			Foo string `json:"foo"`
		}

		body := "{\"foo\":\"bar\"}"
		res := http.Response{
			StatusCode: http.StatusOK,
			Body: ioutil.NopCloser(bytes.NewBufferString(body)),
			ContentLength: int64(len(body)),
		}

		result := testResult{}
		obj, err := client.parseJsonResponse(&res, &result)
		assert.Assert(t, is.Nil(err), "unexpected error from client.parseJsonResponse", err)

		expected := &testResult{Foo: "bar"}
		assert.DeepEqual(t, obj, expected)
	})

	for _, httpStatus := range []int {http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusInternalServerError} {
		testCase := fmt.Sprintf("should return error with decoded message for %v http response", httpStatus)
		t.Run(testCase, func(t *testing.T) {
			apiErrorMessage := "i am an error message"

			client, err := NewClient()
			assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClient", err)

			body := fmt.Sprintf("{\"message\":\"%s\"}", apiErrorMessage)
			res := http.Response{
				StatusCode: httpStatus,
				Body: ioutil.NopCloser(bytes.NewBufferString(body)),
				ContentLength: int64(len(body)),
			}

			result, err := client.parseJsonResponse(&res, nil)

			assert.Assert(t, is.Nil(result), "expected nil result", result)
			assert.Error(t, err, fmt.Sprintf("%d - %s", httpStatus, apiErrorMessage))
		})
	}
}

func TestExecuteRequest(t *testing.T) {
	resetEnvVars()

	t.Run("when expecting an empty response body, should execute request successfully", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client, err := NewClientWithOptions(ts.URL, testKey, testPassphrase, testSecret)
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClientWithOptions", err)

		res, err := client.executeRequest("GET", "/test", nil, nil, 0)
		assert.Assert(t, is.Nil(err), "unexpected error from client.executeRequest", err)

		assert.Equal(t, res, nil)
	})

	t.Run("when expecting a response body, should execute request successfully", func(t *testing.T) {
		type testResponseBody struct {
			Foo string `json:"foo"`
		}

		serverResponseBody := testResponseBody{Foo: "bar"}

		ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			responseBodyBytes, err := json.Marshal(serverResponseBody)
			assert.Assert(t, is.Nil(err), "unexpected error marshaling serverResponseBody", err)

			writer.WriteHeader(http.StatusOK)
			writer.Header().Add(contentTypeHeaderKey, contentTypeHeaderValue)
			writer.Write(responseBodyBytes)
		}))
		defer ts.Close()

		client, err := NewClientWithOptions(ts.URL, testKey, testPassphrase, testSecret)
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClientWithOptions", err)

		clientResponseBody := testResponseBody{}
		res, err := client.executeRequest("GET", "/test", nil, &clientResponseBody, 0)
		assert.Assert(t, is.Nil(err), "unexpected error from client.executeRequest", err)

		assert.DeepEqual(t, res, &serverResponseBody)
	})

	t.Run("when sending a request body, should execute request successfully", func(t *testing.T) {
		type testRequestBody struct {
			Foo string `json:"foo"`
		}

		clientRequestBody := testRequestBody{Foo: "bar"}

		ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer request.Body.Close()
			serverRequestBody := testRequestBody{}
			decoder := json.NewDecoder(request.Body)
			err := decoder.Decode(&serverRequestBody)
			assert.Assert(t, is.Nil(err), "unexpected error decoding clientRequestBody", err)

			assert.DeepEqual(t, clientRequestBody, serverRequestBody)

			writer.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client, err := NewClientWithOptions(ts.URL, testKey, testPassphrase, testSecret)
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClientWithOptions", err)

		res, err := client.executeRequest("GET", "/test", clientRequestBody, nil, 0)
		assert.Assert(t, is.Nil(err), "unexpected error from client.executeRequest", err)

		assert.DeepEqual(t, res, nil)
	})

	t.Run("should surface api error messages", func(t *testing.T) {
		serverResponseBody := ApiError{
			StatusCode: http.StatusTooManyRequests,
			Message: "Too Many Requests",
		}

		ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			responseBodyBytes, err := json.Marshal(serverResponseBody)
			assert.Assert(t, is.Nil(err), "unexpected error marshaling serverResponseBody", err)

			writer.WriteHeader(http.StatusTooManyRequests)
			writer.Header().Add(contentTypeHeaderKey, contentTypeHeaderValue)
			writer.Write(responseBodyBytes)
		}))
		defer ts.Close()

		client, err := NewClientWithOptions(ts.URL, testKey, testPassphrase, testSecret)
		assert.Assert(t, is.Nil(err), "unexpected error creating client using NewClientWithOptions", err)

		res, err := client.executeRequest("GET", "/test", nil, nil, 0)
		assert.DeepEqual(t, res, nil)

		assert.Error(t, err, fmt.Sprintf("%d - %s", serverResponseBody.StatusCode, serverResponseBody.Message))
	})
}