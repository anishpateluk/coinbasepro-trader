package client

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

var clientOptionsMap = map[string]string {
	CoinbaseProBaseurlKey:    "https://testbaseurl.com",
	CoinbaseProKeyKey:        "testKey",
	CoinbaseProPassphraseKey: "testPassphrase",
	CoinbaseProSecretKey:     "testSecret",
}

func resetEnvVars() {
	for key, value := range clientOptionsMap {
		os.Setenv(key, value)
	}
}

func TestNew(t *testing.T) {
	t.Run("does not error when all environment variables exist", func(t *testing.T) {
		resetEnvVars()

		_, err := New()

		if err != nil {
			t.Error("errored when all environment variables have values", err)
		}
	})

	for envVar, _ := range clientOptionsMap {
		testCase := fmt.Sprintf("errors when %s environment variable is missing", envVar)
		t.Run(testCase, func(t *testing.T) {
			resetEnvVars()

			os.Setenv(envVar, "")

			_, err := New()

			if err == nil {
				t.Error("expected error, didn't get one")
			}
		})
	}
}

func TestNewWithOptions(t *testing.T) {
	t.Run("options should override environment variables", func(t *testing.T) {
		resetEnvVars()

		client, err := NewWithOptions("https://hello.com", "KlBsW", "Password1", "zzGfSK=")

		if err != nil {
			t.Error("unexpected error", err)
		}

		baseUrlEnvVar := clientOptionsMap[CoinbaseProBaseurlKey]
		if client.baseUrl == baseUrlEnvVar {
			t.Errorf("wanted %s got %s", client.baseUrl, baseUrlEnvVar)
		}

		keyEnvVar := clientOptionsMap[CoinbaseProKeyKey]
		if client.key == keyEnvVar {
			t.Errorf("wanted %s got %s", client.key, keyEnvVar)
		}

		passphraseEnvVar := clientOptionsMap[CoinbaseProPassphraseKey]
		if client.passphrase == passphraseEnvVar {
			t.Errorf("wanted %s got %s", client.passphrase, passphraseEnvVar)
		}

		secretEnvVar := clientOptionsMap[CoinbaseProSecretKey]
		if client.secret == secretEnvVar {
			t.Errorf("wanted %s got %s", client.secret, secretEnvVar)
		}
	})
}

func TestBuildRequest(t *testing.T) {
	resetEnvVars()

	t.Run("should error when unsupported httpMethod supplied", func(t *testing.T) {
		client, err := New()

		if err != nil {
			t.Error("unexpected error", err)
		}

		for _, unsupportedHttpMethod := range []string {"get", "post", "not a http method" } {
			_, err = client.buildRequest(unsupportedHttpMethod, "/test", nil)

			if err == nil {
				t.Error("expected error when supplying an unsupported http method")
			}
		}

		for _, allowedHttpMethod := range []string { "GET", "POST", "DELETE" } {
			_, err = client.buildRequest(allowedHttpMethod, "/test", nil)

			if err != nil {
				t.Error("unexpected error when supplying supported http methods", err)
			}
		}
	})

	t.Run("should create expected request body", func(t *testing.T) {
		client, err := New()

		if err != nil {
			t.Error("unexpected error", err)
		}

		req, err := client.buildRequest("GET", "/test", nil)
		if err != nil {
			t.Error("unexpected error", err)
		}

		var requestBody []byte
		requestBodyCopyReader, err := req.GetBody()
		if err != nil {
			t.Error("unexpected error", err)
		}

		requestBodyLength, err := requestBodyCopyReader.Read(requestBody)
		if err != nil && requestBodyLength != 0 {
			t.Error("expected 0 length body and an EOF error which is valid for nil body content", err)
		}


		type testStruct struct {
			Foo string
		}

		requestData := testStruct{Foo: "bar"}

		req, err = client.buildRequest("GET", "/test", requestData)
		if err != nil {
			t.Error("unexpected error", err)
		}

		var requestBodyStruct = testStruct{}
		decoder := json.NewDecoder(req.Body)
		err = decoder.Decode(&requestBodyStruct)
		if err != nil {
			t.Error("unexpected error when decoding json", err)
		}

		if requestBodyStruct.Foo != requestData.Foo {
			t.Errorf("wanted %v, got %v", requestData, requestBodyStruct)
		}
	})

	t.Run("use configured base url", func(t *testing.T) {
		expectedUrl := "https://testbaseurl.com/test"

		client, err := New()
		if err != nil {
			t.Error("unexpected error", err)
		}

		req, err := client.buildRequest("GET", "/test", nil)
		if err != nil {
			t.Error("unexpected error building request", err)
		}

		fullRequestUrl := fmt.Sprintf("%s://%s%s", req.URL.Scheme, req.URL.Host, req.URL.Path)
		if fullRequestUrl != expectedUrl {
			t.Errorf("request built with unexpected url, wanted %s got %s", expectedUrl, fullRequestUrl)
		}
	})
}