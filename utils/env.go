package utils

import (
	"fmt"
	"os"
)

type Env struct {
	Username string
	Password string
}

func GetEnv() (*Env, error) {
	username, ok := os.LookupEnv("USERNAME")
	if !ok {
		return nil, fmt.Errorf("username not found")
	}

	password, ok := os.LookupEnv("PASSWORD")
	if !ok {
		return nil, fmt.Errorf("password not found")
	}

	return &Env{Username: username, Password: password}, nil
}
