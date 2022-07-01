package guiver

import (
	"errors"
	"os"
)

type osClient struct{}

func (c *osClient) Get(key string) (string, error) {
	if value, isSet := os.LookupEnv(key); isSet {
		return value, nil
	}

	return "", errors.New("key not set")
}

func (c *osClient) Set(key string, value string) error {
	return os.Setenv(key, value)
}
