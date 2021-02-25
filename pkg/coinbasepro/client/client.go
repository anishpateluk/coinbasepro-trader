package client

import (
	"errors"
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

//func Request(httpMethod, requestPath, payload, )  {
//
//}