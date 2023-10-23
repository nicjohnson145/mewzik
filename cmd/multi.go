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

func multiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multi [INPUTDIR ...]",
		Short: "Move and rename music",
		Args:  cobra.MinimumNArgs(1),
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

			logger := config.InitLogger()
			outputFS := osfs.New(outputPath)
			extensions := mapset.NewSet(strings.Split(viper.GetString(config.Extensions), ",")...)

			for _, dir := range args {
				proc, err := internal.NewProcessor(internal.ProcessorConfig{
					Logger:     logger,
					InputFS:    os.DirFS(dir),
					OutputFS:   outputFS,
					Extensions: extensions,
				})
				if err != nil {
					return err
				}
				if err := proc.Process(); err != nil {
					return err
				}
			}

			return nil
		},
	}
	cmd.Flags().StringP(config.OutputDir, "o", "", "The root of the output directory")
	cmd.Flags().StringP(config.Extensions, "e", config.DefaultExtensions, "Comma separated list of file extensions to process")

	return cmd
}
