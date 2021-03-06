package main

import (
	"github.com/urfave/cli"
	"fmt"
	"os"
	"io/ioutil"
	"github.com/pkg/errors"
	"github.com/asaf/gojet/model"
	"github.com/asaf/gojet/yaml"
	"github.com/asaf/gojet/cmds"
	"github.com/sirupsen/logrus"
)

var playbookRunCmd = cli.Command{
	Name:  "run",
	Usage: "run a playbook",
	Action: func(c *cli.Context) {
		l, err := logrus.ParseLevel(c.GlobalString("log"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "log level [%s] is invalid\n", l)
			return
		}

		logrus.SetLevel(l)

		pbookFname := c.String("file")
		if pbookFname == "" {
			fmt.Fprintf(os.Stderr, "playbook file is required\n")
			return
		}

		if code, err := testRun(pbookFname, c.String("vars"), c.Bool("env-vars")); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(2)
		} else {
			os.Exit(code)
		}
	},
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "file",
			Usage: "file to run",
		},
		cli.StringFlag{
			Name:  "vars",
			Usage: "file containing vars",
		},
		cli.BoolFlag{
			Name:  "env-vars",
			Usage: "bind env vars",
		},
	},
}

func testRun(pbookFname, varsFname string, withEnvVars bool) (int, error) {
	vars, err := loadVars(varsFname, withEnvVars)
	if err != nil {
		return 2, err
	}

	pbook, err := loadPlaybook(pbookFname)
	if err != nil {
		return 2, err
	}

	assertions, err := cmds.RunPlaybook(pbook, vars)
	if err != nil {
		return 2, errors.Wrap(err, "failed to run manifest")
	}

	Cyan.Printf("playbook :: playing %s\n", pbook.Name)
	errors := 0
	for _, as := range assertions {
		MagentaHi.Printf("stage :: %s\n", as.Name)
		errors += printAssertions(as)
	}

	if errors > 0 {
		return 1, nil
	}

	return 0, nil
}

func loadPlaybook(pbookFname string) (*model.Playbook, error) {
	f, err := ioutil.ReadFile(pbookFname)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open playbook")
	}

	var pbook *model.Playbook
	if err := yaml.Unmarshal(f, &pbook); err != nil {
		return nil, errors.Wrap(err, "failed to parse playbook")
	}

	return pbook, nil
}

func loadVars(varsFname string, withEnvVars bool) (model.Vars, error) {
	if varsFname == "" {
		return model.Vars{}, nil
	}

	f, err := ioutil.ReadFile(varsFname)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open vars file")
	}

	var vars model.Vars
	if err := yaml.Unmarshal(f, &vars); err != nil {
		return nil, errors.Wrap(err, "failed to parse vars file")
	}

	if err := vars.Resolve(withEnvVars); err != nil {
		return nil, errors.Wrap(err, "failed to resolve vars")
	}

	return vars, nil
}

func printAssertions(as *model.Assertions) int {
	errors := 0
	for _, a := range as.Assertions {
		if a.Actual != a.Expected {
			Red.Printf("[FAILED: %s] %s - expected [%v] actual [%v]\n", a.Msg, a.Kind, a.Expected, a.Actual)
			errors++
		} else {
			Green.Printf("[%s] %s \n", a.Msg, a.Kind)
		}
	}

	return errors
}
