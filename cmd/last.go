package cmd

import (
	"github.com/anhk/sshp/app"
	"github.com/spf13/cobra"
)

var lastCount int = 10

var lastCmd = &cobra.Command{
	Use:   "last",
	Short: "last command, list the last `n` logged-in devices",
	Run: func(cmd *cobra.Command, args []string) {
		app := &app.App{}
		app.Init()
		app.ShowLastLoggedInDevices(lastCount)
	},
}

func init() {
	lastCmd.Flags().IntVarP(&lastCount, "number", "n", 10, "list `n` devices")
	rootCmd.AddCommand(lastCmd)
}
