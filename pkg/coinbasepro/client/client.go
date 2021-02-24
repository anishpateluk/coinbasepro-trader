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

type Client struct {
	baseUrl string
	key string
	passphrase string
	secret string
}

func New(params ...string) (*Client, error) {
	baseUrl := os.Getenv("COINBASE_PRO_BASEURL")
	key := os.Getenv("COINBASE_PRO_KEY")
	passphrase := os.Getenv("COINBASE_PRO_PASSPHRASE")
	secret := os.Getenv("COINBASE_PRO_SECRET")

	if baseUrl == "" && len(params) < 1 {
		return nil, errors.New("missing COINBASE_PRO_BASEURL")
	}

	if key == "" && len(params) < 2 {
		return nil, errors.New("missing COINBASE_PRO_KEY")
	}

	if passphrase == "" && len(params) < 3 {
		return nil, errors.New("missing COINBASE_PRO_PASSPHRASE")
	}

	if secret == "" && len(params) < 4 {
		return nil, errors.New("missing COINBASE_PRO_SECRET")
	}


	//baseUrl := params[1:]
	//fmt.Sprintf("baseUrl %s\n", baseUrl)

	client := Client{}

	return &client, nil
}

//func Request(httpMethod, requestPath, payload, )  {
//
//}