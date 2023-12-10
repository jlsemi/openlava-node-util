package lsf

import (
	"bufio"
	"fmt"
	"jlsemi.com/openlava-utils/logs"
	"jlsemi.com/openlava-utils/util"
	"os"
	"strings"
	"text/template"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const (
	DEFAULT_MASTER_NODE_NAME = "manager"
	DEFAULT_CLIENT_NODE_NAME = "compute000"
	DEFAULT_SQLITE_DB_PATH   = "/tmp/openlava.db"
)

var lsfLog = logs.GetLogger()

type LsfInfo struct {
	MasterNode  string
	ClientNode  string // 跳板机
	WorkerNodes []string
	Db          *gorm.DB
}

type LsfClusterOpenlavaConfig struct {
	HostName string
	HostType string
}

type LsbHostConfig struct {
	HostName string
	MaxNodes string
}

func GetHosts() ([]string, error) {
	hosts := []string{}
	cmd := fmt.Sprintf("bhosts -w | grep -v HOST_NAME | awk '{print $1}'")
	rsp, err := util.ExecCommand("bash", []string{"-c", cmd})

	if err != nil {
		lsfLog.Errorf("execute failed, cmd: %s, error: %v", cmd, err)
		return hosts, err
	}

	reader := strings.NewReader(rsp)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		hostname := scanner.Text()
		hosts = append(hosts, hostname)
	}

	lsfLog.Infof("GetHosts from bhosts, get %d records", len(hosts))
	return hosts, nil
}

func (info *LsfInfo) GenBhostsConfig(filepath string) error {
	hosts := []*LsbHostConfig{}

	hosts = append(hosts, &LsbHostConfig{
		HostName: info.MasterNode,
		MaxNodes: "0",
	})

	for _, hostname := range info.WorkerNodes {
		hosts = append(hosts, &LsbHostConfig{
			HostName: hostname,
			MaxNodes: "!",
		})
	}

	tmpl, err := template.New("lsb.hosts").Parse(bhostConfig)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, hosts)
	if err != nil {
		return err
	}

	return nil
}

func (info *LsfInfo) GenLsfClusterConfig(filepath string) error {
	hosts := []*LsfClusterOpenlavaConfig{}

	hosts = append(hosts, &LsfClusterOpenlavaConfig{
		HostName: info.ClientNode,
		HostType: "0",
	})

	hosts = append(hosts, &LsfClusterOpenlavaConfig{
		HostName: info.MasterNode,
		HostType: "1",
	})

	for _, hostname := range info.WorkerNodes {
		hosts = append(hosts, &LsfClusterOpenlavaConfig{
			HostName: hostname,
			HostType: "1",
		})
	}

	tmpl, err := template.New("lsf.cluster.openlava").Parse(clusterConfig)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath)
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, hosts)
	if err != nil {
		return err
	}

	return nil
}

func (info *LsfInfo) UpdateWorkerNodes(hosts []string) {
	nodes := []string{}

	for _, hostname := range hosts {
		if hostname != info.MasterNode {
			nodes = append(nodes, hostname)
		}
	}

	info.WorkerNodes = nodes
}

func (info *LsfInfo) Init() error {
	hostsFromDb, err := info.GetHosts()

	if err != nil {
		return err
	}

	if len(hostsFromDb) > 0 {
		lsfLog.Infof("Init: Update hosts from db")
		info.UpdateWorkerNodes(hostsFromDb)
		return nil
	}

	hostsFromBhosts, err := GetHosts()
	if err != nil {
		return err
	}

	err = info.UpdateHosts(hostsFromBhosts)
	if err != nil {
		return err
	}

	info.UpdateWorkerNodes(hostsFromBhosts)
	return nil
}

func MakeLsfInfo() (*LsfInfo, error) {
	info := &LsfInfo{
		MasterNode: DEFAULT_MASTER_NODE_NAME,
		ClientNode: DEFAULT_CLIENT_NODE_NAME,
	}

	dbPath := DEFAULT_SQLITE_DB_PATH

	if os.Getenv("SQLITE_DB_PATH") != "" {
		dbPath = os.Getenv("SQLITE_DB_PATH")
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	info.Db = db

	err = db.AutoMigrate(&LsfHost{})

	if err != nil {
		return nil, err
	}

	err = info.Init()
	if err != nil {
		return nil, err
	}

	return info, nil
}
