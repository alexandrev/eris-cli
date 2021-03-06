package commands

import (
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var quiet bool

var VerSion = &cobra.Command{
	Use:   "version",
	Short: "Display Eris's Platform Version.",
	Long:  `Display the current installed version of Eris.`,
	Run:   DisplayVersion,
}

func buildVerSionCommand() {
	addVerSionFlags()
}

func addVerSionFlags() {
	VerSion.Flags().BoolVarP(&quiet, "quiet", "q", false, "machine readable output")
}

func DisplayVersion(cmd *cobra.Command, args []string) {
	if !quiet {
		logger.Println("Eris CLI Version: " + VERSION)
	} else {
		logger.Println(VERSION)
	}
}
