package main

import (
	"github.com/urfave/cli/v2"
	"jlsemi.com/openlava-utils/logs"
	"jlsemi.com/openlava-utils/lsf"
	"os"
)

var utilLog = logs.GetLogger()

func NodeAddCommand(action func(ctx *cli.Context) error) *cli.Command {
	return &cli.Command{
		Name:   "add",
		Usage:  "add node to openlava, and generate new config",
		Action: action,
		Flags:  []cli.Flag{},
	}
}

func AddNode(ctx *cli.Context) error {
	lsfInfo, err := lsf.MakeLsfInfo()

	if err != nil {
		return err
	}

	err = lsfInfo.GenLsfClusterConfig()
	if err != nil {
		return err
	}

	err = lsfInfo.GenBhostsConfig()
	if err != nil {
		return err
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "nodecli"
	app.Usage = "openlava node utils"
	app.Commands = []*cli.Command{
		NodeAddCommand(AddNode),
	}

	if err := app.Run(os.Args); err != nil {
		utilLog.Fatal(err)
	}
}
