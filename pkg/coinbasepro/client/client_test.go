package client

import (
	"fmt"
	"os"
	"testing"
)


func TestNew(t *testing.T) {
	const testBaseUrl = "https://testbaseurl.com"
	const testKey = "testKey"
	const testPassphrase = "testPassphrase"
	const testSecret = "testSecret"

	t.Run("does not error when all environment variables exists and no params supplied ", func(t *testing.T) {
		os.Setenv("COINBASE_PRO_BASEURL", testBaseUrl)
		os.Setenv("COINBASE_PRO_KEY", testKey)
		os.Setenv("COINBASE_PRO_PASSPHRASE", testPassphrase)
		os.Setenv("COINBASE_PRO_SECRET", testSecret)

		_, err := New()

		if err != nil {
			t.Error("errored when all environment variables have values")
		}
	})

	noParamsMissingEnvVarCases := []string {
		"COINBASE_PRO_BASEURL",
		"COINBASE_PRO_KEY",
		"COINBASE_PRO_PASSPHRASE",
		"COINBASE_PRO_SECRET",
	}

	for i :=0; i < len(noParamsMissingEnvVarCases); i++ {
		envVar := noParamsMissingEnvVarCases[i]
		testCase := fmt.Sprintf("errors when no params supplied and %s environment variable is missing", envVar)
		t.Run(testCase, func(t *testing.T) {
			os.Setenv("COINBASE_PRO_BASEURL", testBaseUrl)
			os.Setenv("COINBASE_PRO_KEY", testKey)
			os.Setenv("COINBASE_PRO_PASSPHRASE", testPassphrase)
			os.Setenv("COINBASE_PRO_SECRET", testSecret)

			os.Setenv(envVar, "")

			_, err := New()

			if err == nil {
				t.Error("expected error, didn't get one")
			}
		})
	}
}