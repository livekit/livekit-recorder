package recorder

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChrome(t *testing.T) {
	r := &Recorder{}
	require.NoError(t, r.LaunchChrome("https://www.youtube.com/watch?v=v4I-19YAi2A", 1920, 1080))
}
