package client

import "fmt"

type ApiError struct {
	StatusCode int
	Message string `json:"message"`
}

func (e ApiError) Error() string {
	return fmt.Sprintf("%d - %s", e.StatusCode, e.Message)
}