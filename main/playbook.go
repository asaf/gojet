package main

import (
	"github.com/urfave/cli"
)

var testCmds = cli.Command{
	Aliases: []string{"p", "pbook"},
	Name:    "playbook",
	Usage:   "Playbook Commands",
	Subcommands: []cli.Command{
		playbookRunCmd,
	},
}
