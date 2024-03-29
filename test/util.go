package test

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func ReadFile(t *testing.T, filename string) string {
	content, err := ioutil.ReadFile(filename)
	require.NoError(t, err)
	return string(content)
}
