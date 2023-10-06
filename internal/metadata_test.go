package internal

import (
	"testing"
	"github.com/stretchr/testify/require"
	"os"
)

func TestComputeDestinationPath(t *testing.T) {
	t.Run("smokes", func(t *testing.T) {
		root := os.DirFS("./testdata/sample-files/Anders - 669")
		got, err := computeDestinationPath(root, "02 - Diamonds.mp3")
		require.NoError(t, err)
		require.Equal(t, "Anders/669/02_Diamonds.mp3", got)
	})
}
