package main

import (
	"github.com/anhk/sshp/cmd"
	"github.com/anhk/sshp/pkg/loginit"
	"github.com/cihub/seelog"
)

func main() {
	loginit.Init()
	defer seelog.Flush()
	cmd.Execute()
}
