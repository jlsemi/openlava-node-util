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
	queuename string

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

	QueueNameFlag = &cli.StringFlag{
		Name:        "queuename",
		Usage:       "set queuename to operate",
		Destination: &queuename,
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
			QueueNameFlag,
		},
	}
}

func NodeDelCommand(action func(ctx *cli.Context) error) *cli.Command {
	return &cli.Command{
		Name:   "del",
		Usage:  "del node from openlava, and regenerate config",
		Action: action,
		Flags: []cli.Flag{
			ConfigDirFlag,
			HostNameFlag,
		},
	}
}

func ShowQueueInfoCommand(action func(ctx *cli.Context) error) *cli.Command {
	return &cli.Command{
		Name:   "queues",
		Usage:  "show all queue info",
		Action: action,
		Flags:  []cli.Flag{},
	}
}

func GenConfigCommand(action func(ctx *cli.Context) error) *cli.Command {
	return &cli.Command{
		Name:   "gen",
		Usage:  "generate config based on database",
		Action: action,
		Flags: []cli.Flag{
			ConfigDirFlag,
		},
	}
}

func SyncQueueInfoCommand(action func(ctx *cli.Context) error) *cli.Command {
	return &cli.Command{
		Name:   "sync_queue",
		Usage:  "sync queue info based on bqueues info",
		Action: action,
		Flags:  []cli.Flag{},
	}
}

func ShowQueueInfo(ctx *cli.Context) error {
	lsfInfo, err := lsf.MakeLsfInfo()

	if err != nil {
		return err
	}

	for _, info := range lsfInfo.QueueInfo {
		utilLog.Infof("QueueInfo: %v users: %v, hosts: %v", info.QueueName, info.Users, info.Hosts)
	}

	return nil
}

func GenConfig(ctx *cli.Context) error {
	lsfInfo, err := lsf.MakeLsfInfo()

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

	err = lsfInfo.GenLsfQueueConfig(fmt.Sprintf("%s/lsb.queues", configDir))
	if err != nil {
		return err
	}

	utilLog.Infof("GenConfig finished, config dir: %s", configDir)
	return nil
}

func DelNode(ctx *cli.Context) error {
	lsfInfo, err := lsf.MakeLsfInfo()

	if err != nil {
		return err
	}

	err = lsfInfo.DelHostname(hostname)
	if err != nil {
		return err
	}

	err = lsfInfo.DelHostFromAllQueues(hostname)
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

	err = lsfInfo.GenLsfQueueConfig(fmt.Sprintf("%s/lsb.queues", configDir))
	if err != nil {
		return err
	}

	return nil
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

	err = lsfInfo.AddHostToQueue(hostname, queuename)
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

	err = lsfInfo.GenLsfQueueConfig(fmt.Sprintf("%s/lsb.queues", configDir))
	if err != nil {
		return err
	}

	return nil
}

func SyncQueueInfo(ctx *cli.Context) error {
	lsfInfo, err := lsf.MakeLsfInfo()

	if err != nil {
		return err
	}

	return lsfInfo.SyncQueue()
}

func main() {
	app := cli.NewApp()
	app.Name = "nodecli"
	app.Usage = "openlava node utils"
	app.Commands = []*cli.Command{
		NodeAddCommand(AddNode),
		NodeDelCommand(DelNode),
		ShowQueueInfoCommand(ShowQueueInfo),
		SyncQueueInfoCommand(SyncQueueInfo),
		GenConfigCommand(GenConfig),
	}

	if err := app.Run(os.Args); err != nil {
		utilLog.Fatal(err)
	}
}
