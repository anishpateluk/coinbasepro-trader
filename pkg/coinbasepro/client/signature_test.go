package client

import (
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"testing"
)

func TestCreateSignature(t *testing.T) {

	const secret = "YmFzZTY0c2VjcmV0"
	const timestamp = "1614191039"
	const httpMethod = "GET"
	const requestPath = "/test/testing"
	const jsonBody = "{\"foo\":\"bar\"}"

	const expectedSignature = "sVmyaEVrIxlhfPQMes8zDl/UCo1EEZpCIYrnFkutxXo="


	t.Run("returns expected signature", func(t *testing.T) {
		signature, err := createSignature(secret, timestamp, httpMethod, requestPath, jsonBody)
		assert.Assert(t, is.Nil(err), "unexpected error creating signature", err)

		assert.Equal(t, signature, expectedSignature)
	})
}