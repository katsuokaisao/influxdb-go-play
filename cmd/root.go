package cmd

import (
	"github.com/spf13/cobra"
)

var (
	fileName string
)

func init() {
	rootCmd.AddCommand(writeCmd)
	rootCmd.AddCommand(writeBlockCmd)
	rootCmd.AddCommand(checkThreshold10MinutesAgoCmd)

	writeCmd.PersistentFlags().StringVarP(&fileName, "file", "f", "", "write data from this file")
	writeBlockCmd.PersistentFlags().StringVarP(&fileName, "file", "f", "", "write data from this file")
}

var rootCmd = &cobra.Command{
	Use: "",
}

func Execute() {
	rootCmd.Execute()
}
