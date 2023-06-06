package main

import (
	"github.com/anhk/sshp/cmd"
	"github.com/anhk/sshp/pkg/loginit"
)

func main() {
	loginit.Init()
	cmd.Execute()
}
