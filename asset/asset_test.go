package asset

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultImage(t *testing.T) {
	require.NotNil(t, DefaultImage)

	header := make([]byte, 512)
	_, err := io.ReadFull(bytes.NewReader(DefaultImage), header)
	require.NoError(t, err)

	ct := http.DetectContentType(header)
	assert.Equal(t, "image/png", ct)
}
