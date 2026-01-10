package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv(t *testing.T) {
	expectedEnv := Env{
		Username: "some-username",
		Password: "some-password",
	}

	t.Setenv("USERNAME", expectedEnv.Username)
	t.Setenv("PASSWORD", expectedEnv.Password)

	env, err := GetEnv()

	assert.NoError(t, err)
	assert.Equal(t, expectedEnv.Username, env.Username)
	assert.Equal(t, expectedEnv.Password, env.Password)
}

func TestGetEnvMissingUsername(t *testing.T) {
	expectedEnv := Env{
		Password: "some-password",
	}

	t.Setenv("PASSWORD", expectedEnv.Password)

	env, err := GetEnv()

	assert.Nil(t, env)
	assert.Equal(t, "username not found", err.Error())
}
