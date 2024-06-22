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
	rootCmd.AddCommand(readCmd)

	writeCmd.PersistentFlags().StringVarP(&fileName, "file", "f", "data.txt", "write data from this file")
	writeBlockCmd.PersistentFlags().StringVarP(&fileName, "file", "f", "data.txt", "write data from this file")
}

var rootCmd = &cobra.Command{
	Use: "",
}

func Execute() {
	rootCmd.Execute()
}
