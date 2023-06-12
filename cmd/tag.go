package cmd

import (
	"os"

	"github.com/anhk/sshp/app"
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "TAG command, tag <id> <tagName>",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Help()
			os.Exit(-1)
		}
		app := &app.App{}
		app.Init()
		app.DoTag(args[0], args[1])
	},
}

func init() {
	rootCmd.AddCommand(tagCmd)
}
