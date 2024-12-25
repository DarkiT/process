package utils

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// All returns a copy of strings representing the environment,
// in the form "key=value".
func All() []string {
	return os.Environ()
}

// Map returns a copy of strings representing the environment as a map.
func Map() map[string]string {
	return MapFromEnv(os.Environ())
}

// Set sets the value of the environment variable named by the `key`.
// It returns an error, if any.
func Set(key, value string) (err error) {
	err = os.Setenv(key, value)
	if err != nil {
		err = errors.New(fmt.Sprintf(`set environment key-value failed with key "%s", value "%s"`, key, value))
	}
	return
}

// SetMap sets the environment variables using map.
func SetMap(m map[string]string) (err error) {
	for k, v := range m {
		if err = Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

// MapFromEnv converts environment variables from slice to map.
func MapFromEnv(envs []string) map[string]string {
	m := make(map[string]string)
	i := 0
	for _, s := range envs {
		i = strings.IndexByte(s, '=')
		m[s[0:i]] = s[i+1:]
	}
	return m
}
