package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(writeCmd)
	rootCmd.AddCommand(writeBlockCmd)
}

var rootCmd = &cobra.Command{
	Use: "",
}

func Execute() {
	rootCmd.Execute()
}
