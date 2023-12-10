package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"jlsemi.com/openlava-utils/logs"
	"jlsemi.com/openlava-utils/lsf"
	"os"
)

var utilLog = logs.GetLogger()

var (
	configDir string
	hostname  string

	ConfigDirFlag = &cli.StringFlag{
		Name:        "config_dir",
		Usage:       "path to openlava config dir",
		Destination: &configDir,
		Required:    true,
	}

	HostNameFlag = &cli.StringFlag{
		Name:        "hostname",
		Usage:       "set hostname to operate",
		Destination: &hostname,
		Required:    true,
	}
)

func NodeAddCommand(action func(ctx *cli.Context) error) *cli.Command {
	return &cli.Command{
		Name:   "add",
		Usage:  "add node to openlava, and generate new config",
		Action: action,
		Flags: []cli.Flag{
			ConfigDirFlag,
			HostNameFlag,
		},
	}
}

func AddNode(ctx *cli.Context) error {
	lsfInfo, err := lsf.MakeLsfInfo()

	if err != nil {
		return err
	}

	err = lsfInfo.AddHostname(hostname)
	if err != nil {
		return err
	}

	err = lsfInfo.GenLsfClusterConfig(fmt.Sprintf("%s/lsf.cluster.openlava", configDir))
	if err != nil {
		return err
	}

	err = lsfInfo.GenBhostsConfig(fmt.Sprintf("%s/lsb.hosts", configDir))
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
