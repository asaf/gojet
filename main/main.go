package main

import (
	"github.com/urfave/cli"
	"github.com/asaf/gojet/consts"
	"os"
	"fmt"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {
	app := cli.NewApp()
	app.Name = "gojet"
	app.Version = consts.Ver
	app.Usage = "command line utility"
	app.Flags = []cli.Flag{
	}
	app.Commands = []cli.Command{
		testCmds,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
	}
}
