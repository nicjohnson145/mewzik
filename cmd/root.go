package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/deckarep/golang-set/v2"
	"github.com/jarxorg/wfs/osfs"
	"github.com/nicjohnson145/mewzik/config"
	"github.com/nicjohnson145/mewzik/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func Root() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mewzik [INPUTDIR]",
		Short: "Move and rename music",
		Args:  cobra.ExactArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// So we don't print usage messages on execution errors
			cmd.SilenceUsage = true
			// So we dont double report errors
			cmd.SilenceErrors = true
			return config.InitializeConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			outputPath := viper.GetString(config.OutputDir)
			if outputPath == "" {
				return fmt.Errorf("%v required", config.OutputDir)
			}

			proc := internal.NewProcessor(internal.ProcessorConfig{
				Logger:     config.InitLogger(),
				InputFS:    os.DirFS(args[0]),
				OutputFS:   osfs.New(outputPath),
				Extensions: mapset.NewSet(strings.Split(viper.GetString(config.Extensions), ",")...),
				Overrides: internal.MetadataOverride{
					Artist: viper.GetString(config.Artist),
					Album: viper.GetString(config.Album),
				},
			})

			return proc.Process()
		},
	}
	rootCmd.PersistentFlags().BoolP(config.Debug, "d", false, "Enable debug logging")

	rootCmd.Flags().StringP(config.OutputDir, "o", "", "The root of the output directory")
	rootCmd.Flags().StringP(config.Extensions, "e", config.DefaultExtensions, "Comma separated list of file extensions to process")

	rootCmd.Flags().StringP(config.Artist, "a", "", "Force file artist, disregarding file metadata")
	rootCmd.Flags().StringP(config.Album, "b", "", "Force file album, disregarding file metadata")

	rootCmd.AddCommand(
		versionCmd(),
	)

	return rootCmd
}
