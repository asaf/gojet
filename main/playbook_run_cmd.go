package main

import (
	"github.com/urfave/cli"
	"fmt"
	"os"
	"io/ioutil"
	"github.com/pkg/errors"
	"github.com/ghodss/yaml"
	"github.com/asaf/gojet/model"
	"github.com/asaf/gojet/cmds"
	"bytes"
	"text/template"
)

var playbookRunCmd = cli.Command{
	Name:  "run",
	Usage: "run a playbook",
	Action: func(c *cli.Context) {
		pbookFname := c.String("file")
		if pbookFname == "" {
			fmt.Fprintf(os.Stderr, "playbook file is required\n")
			return
		}

		if err := testRun(pbookFname, c.String("vars")); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
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
	},
}

func testRun(pbookFname, varsFname string) error {
	vars, err := loadVars(varsFname)
	if err != nil {
		return err
	}

	pbook, err := loadPlaybook(pbookFname, vars)
	if err != nil {
		return err
	}

	assertions, err := cmds.RunPlaybook(pbook, vars)
	if err != nil {
		return errors.Wrap(err, "failed to run manifest")
	}

	Cyan.Printf("playing %s\n", pbook.Name)
	for st, as := range assertions {
		MagentaHi.Printf("stage %s\n", st)
		printAssertions(as)
	}

	return nil
}

func loadPlaybook(pbookFname string, vars model.Vars) (*model.Playbook, error) {
	f, err := ioutil.ReadFile(pbookFname)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open playbook")
	}

	tmpl := template.Must(template.New("").Parse(string(f)))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return nil, errors.Wrap(err, "playbook is invalid")
	}

	var man *model.Playbook
	if err := yaml.Unmarshal(buf.Bytes(), &man); err != nil {
		return nil, errors.Wrap(err, "failed to parse playbook")
	}

	return man, nil
}

func loadVars(varsFname string) (model.Vars, error) {
	f, err := ioutil.ReadFile(varsFname)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open vars file")
	}

	var vars model.Vars
	if err := yaml.Unmarshal(f, &vars); err != nil {
		return nil, errors.Wrap(err, "failed to parse vars file")
	}

	return vars, nil
}

func printAssertions(as *model.Assertions) {
	for _, a := range as.Assertions {
		if a.Actual != a.Expected {
			Red.Printf("[FAILED: %s] %s - expected [%v] actual [%v]\n", a.Msg, a.Kind, a.Expected, a.Actual)
		} else {
			Green.Printf("[SUCCESS: %s] %s \n", a.Msg, a.Kind)
		}
	}
}
