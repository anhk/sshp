package cmd

import (
	"fmt"
	"os"

	"github.com/anhk/sshp/app"
	"github.com/spf13/cobra"
)

var sshConfig = &app.Config{}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "ssh command, for OpenSSH remote login client",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintf(os.Stderr, "invalid argument for ssh: %v", args)
			os.Exit(-1)
		}
		//sshConfig.Target = args[0]

		app := &app.App{}
		app.Init()
		app.DoSSH(sshConfig.Parse(args[0]))
	},
}

func init() {

	sshCmd.Flags().IntVarP(&sshConfig.Port, "port", "p", 22, "ssh port")
	rootCmd.AddCommand(sshCmd)
}
