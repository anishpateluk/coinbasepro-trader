package coinbasepro

import (
	"testing"
)

func TestSigning(t *testing.T) {

	const secret = "YmFzZTY0c2VjcmV0"
	const timestamp = "1614191039"
	const httpMethod = "GET"
	const requestPath = "/test/testing"
	const jsonBody = "{\"foo\":\"bar\"}"

	const expectedSignature = "sVmyaEVrIxlhfPQMes8zDl/UCo1EEZpCIYrnFkutxXo="


	t.Run("returns expected signature", func(t *testing.T) {
		signature, err := createSignature(secret, timestamp, httpMethod, requestPath, jsonBody)
		if err != nil {
			t.Errorf("Error creating signature %s", err.Error())
			return
		}

		if signature != expectedSignature {
			t.Errorf("wanted %s got %s", signature, expectedSignature)
		}
	})
}