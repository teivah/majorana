package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func ReadFile(t *testing.T, filename string) string {
	content, err := os.ReadFile(filename)
	require.NoError(t, err)
	return string(content)
}
