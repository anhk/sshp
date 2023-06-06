package cmd

import (
	"github.com/anhk/sshp/app"
	"github.com/spf13/cobra"
)

var scpCmd = &cobra.Command{
	Use:   "scp",
	Short: "scp command, for OpenSSH secure file copy",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	config := app.Config{}
	scpCmd.Flags().IntVar(&config.Port, "P", 22, "scp port")
	rootCmd.AddCommand(scpCmd)
}
