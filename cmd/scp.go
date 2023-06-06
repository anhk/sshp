package cmd

import (
	"os"

	"github.com/cihub/seelog"

	"github.com/anhk/sshp/app"
	"github.com/spf13/cobra"
)

var scpConfig = &app.ScpConfig{}

var scpCmd = &cobra.Command{
	Use:   "scp",
	Short: "SCP command, for OpenSSH secure file copy",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Help()
			os.Exit(-1)
		}
		seelog.Infof("%v -> %v", args[0], args[1])

		app := &app.App{}
		app.Init()
		app.DoSCP(scpConfig.Parse(args[0], args[1]))
	},
}

func init() {
	scpCmd.Flags().IntVarP(&scpConfig.Port, "port", "P", 22, "Specifies the port to connect to on the remote host.")
	scpCmd.Flags().BoolVarP(&scpConfig.Dir, "directory", "r", false, "Recursively copy entire directories")
	rootCmd.AddCommand(scpCmd)
}
