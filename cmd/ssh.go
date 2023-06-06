package cmd

import (
	"os"

	"github.com/anhk/sshp/app"
	"github.com/spf13/cobra"
)

var sshConfig = &app.SshConfig{}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "SSH command, for OpenSSH remote login client",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			os.Exit(-1)
		}
		app := &app.App{}
		app.Init()
		app.DoSSH(sshConfig.Parse(args[0]))
	},
}

func init() {
	sshCmd.Flags().IntVarP(&sshConfig.Port, "port", "p", 22, "Specifies the port to connect to on the remote host.")
	rootCmd.AddCommand(sshCmd)
}
