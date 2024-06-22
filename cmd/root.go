package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(writeCmd)
	rootCmd.AddCommand(writeBlockCmd)
	rootCmd.AddCommand(readCmd)
}

var rootCmd = &cobra.Command{
	Use: "",
}

func Execute() {
	rootCmd.Execute()
}
