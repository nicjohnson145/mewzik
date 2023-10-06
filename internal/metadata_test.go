package internal

import (
	"testing"
	"github.com/stretchr/testify/require"
	"os"
)

func TestComputeDestinationPath(t *testing.T) {
	t.Run("smokes", func(t *testing.T) {
		root := os.DirFS("./testdata/sample-files/Anders - 669")
		got, err := computeDestinationPath(root, MetadataOverride{}, "02 - Diamonds.mp3")
		require.NoError(t, err)
		require.Equal(t, "Anders/669/02_Diamonds.mp3", got)
	})

	t.Run("overrides", func(t *testing.T) {
		root := os.DirFS("./testdata/sample-files/Anders - 669")
		md := MetadataOverride{
			Artist: "Foo Artist",
			Album: "Bar Album",
		}
		got, err := computeDestinationPath(root, md, "02 - Diamonds.mp3")
		require.NoError(t, err)
		require.Equal(t, "Foo_Artist/Bar_Album/02_Diamonds.mp3", got)
	})
}
