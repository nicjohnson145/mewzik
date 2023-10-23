package config

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	Debug = "debug"

	OutputDir  = "output-dir"
	Extensions = "extensions"

	Artist      = "artist"
	Album       = "album"
	AlbumArtist = "album-artist"
)

const (
	DefaultExtensions = ".mp3"
)

func InitializeConfig(cmd *cobra.Command) error {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	viper.SetDefault(Extensions, DefaultExtensions)

	return viper.BindPFlags(cmd.Flags())
}
