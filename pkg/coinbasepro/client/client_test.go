package client

import (
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
			t.Error("errored when all environment variables have values")
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
			t.Error("unexpected error")
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