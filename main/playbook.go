package main

import (
	"github.com/urfave/cli"
)

var testCmds = cli.Command{
	Name:  "playbook",
	Usage: "Playbook Commands",
	Subcommands: []cli.Command{
		playbookRunCmd,
	},
}
