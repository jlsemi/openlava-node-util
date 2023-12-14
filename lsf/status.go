package lsf

import (
	"bufio"
	"fmt"
	"jlsemi.com/openlava-utils/logs"
	"jlsemi.com/openlava-utils/util"
	"os"
	"regexp"
	"strings"
	"text/template"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

const (
	DEFAULT_MASTER_NODE_NAME = "manager"
	DEFAULT_CLIENT_NODE_NAME = "compute000"
	DEFAULT_SQLITE_DB_PATH   = "/tmp/openlava.db"
	DEFAULT_ALL_USER         = "all user"
)

var lsfLog = logs.GetLogger()

type LsfInfo struct {
	MasterNode  string
	ClientNode  string // 跳板机
	WorkerNodes []string
	QueueInfo   []*LsfQueueInfo // Queue相关信息
	Db          *gorm.DB
}

type LsfQueueInfo struct {
	QueueName string
	Users     string
	Hosts     []string
}

type LsfClusterOpenlavaConfig struct {
	HostName string
	HostType string
}

type LsbHostConfig struct {
	HostName string
	MaxNodes string
}

func GetQueueDetailInfo(queueName string) (*LsfQueueInfo, error) {
	// 获得当前queue的基本信息
	var (
		userGroup string
		hosts     []string
	)

	cmd := fmt.Sprintf("bqueues -l %s", queueName)
	rsp, err := util.ExecCommand("bash", []string{"-c", cmd})

	if err != nil {
		lsfLog.Errorf("execute failed, cmd: %s,  error: %v", cmd, err)
		return nil, err
	}

	// patterns
	hostContentPattern := regexp.MustCompile(`HOSTS:\s*(.*)`)
	userPattern := regexp.MustCompile(`USERS:\s*([^/\n]*)`)
	hostSplitPattern := regexp.MustCompile(`[^\s]*`)

	result := userPattern.FindAllStringSubmatch(rsp, -1)
	if len(result) == 0 {
		return nil, fmt.Errorf("GetQueueDetailInfo failed, usergroup not found")
	}

	userGroup = result[0][1]

	result = hostContentPattern.FindAllStringSubmatch(rsp, -1)
	if len(result) == 0 {
		return nil, fmt.Errorf("GetQueueInfo failed, hostCotent not found")
	}

	hosts = hostSplitPattern.FindAllString(result[0][1], -1)

	return &LsfQueueInfo{
		QueueName: queueName,
		Users:     userGroup,
		Hosts:     hosts,
	}, nil
}

func GetQueuesInfo() ([]*LsfQueueInfo, error) {
	queues := []*LsfQueueInfo{}

	cmd := fmt.Sprintf("bqueues -w | grep -v QUEUE_NAME | awk '{print $1}'")
	rsp, err := util.ExecCommand("bash", []string{"-c", cmd})

	if err != nil {
		lsfLog.Errorf("execute failed, cmd: %s, error: %v", cmd, err)
		return nil, err
	}

	reader := strings.NewReader(rsp)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		queuename := scanner.Text()
		result, err := GetQueueDetailInfo(queuename)

		if err != nil {
			return nil, err
		}

		queues = append(queues, result)
	}

	return queues, nil
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

func (info *LsfInfo) GenLsfQueueConfig(filepath string) error {
	var (
		f   *os.File
		err error
	)

	tmpl, err := template.New("lsb.queues").Parse(bqueueConfig)
	if err != nil {
		return err
	}

	if filepath == "" {
		f = os.Stdout
	} else {
		f, err = os.Create(filepath)
		if err != nil {
			return err
		}
	}

	err = tmpl.Execute(f, info.QueueInfo)
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

func (info *LsfInfo) AddHostToQueue(hostname string, queuename string) error {
	var (
		lsfQueue     []LsfQueue
		lsfQueueHost []LsfQueueHost
	)

	if hostname == info.ClientNode {
		lsfLog.Infof("clientNode %s cannot add to queue", hostname)
		return nil
	}

	// update queue info from lsf

	if result := info.Db.Where(&LsfQueue{QueueName: queuename}).Find(&lsfQueue); result.RowsAffected == 0 {
		return fmt.Errorf("queuename %v not exist", queuename)
	}

	if result := info.Db.Where(&LsfQueueHost{QueueName: queuename, HostName: hostname}).Find(&lsfQueueHost); result.RowsAffected == 0 {
		if err := info.Db.Create(&LsfQueueHost{
			QueueName: queuename,
			HostName:  hostname,
		}).Error; err != nil {
			lsfLog.Errorf(
				"Create QueueHost info failed, Queue name: %v, host name: %v,  error: %v",
				queuename,
				hostname,
				err,
			)
			return err
		}
	}

	lsfLog.Infof("Add %s to queue %s", hostname, queuename)
	return info.InitQueue()
}

func (info *LsfInfo) DelHostFromAllQueues(hostname string) error {
	if err := info.Db.Where("host_name = ?", hostname).Delete(&LsfQueueHost{}).Error; err != nil {
		return err
	}

	lsfLog.Infof("Delete %s from queues", hostname)
	return info.InitQueue()
}

func (info *LsfInfo) SyncQueue() error {

	var (
		queueNamesFromDbDict  = make(map[string]*LsfQueueInfo)
		queueNamesfromLsfDict = make(map[string]*LsfQueueInfo)
	)

	// get queues from db
	queueInfoFromDb, err := info.GetQueueInfo()
	if err != nil {
		return err
	}

	for _, result := range queueInfoFromDb {
		queueNamesFromDbDict[result.QueueName] = result
	}

	// get queues from bqueues
	queueInfoFromBqueues, err := GetQueuesInfo()
	if err != nil {
		return err
	}

	for _, result := range queueInfoFromBqueues {
		queueNamesfromLsfDict[result.QueueName] = result
	}

	// sync db info based on bqueues result
	for _, queueResult := range queueInfoFromBqueues {
		if _, ok := queueNamesFromDbDict[queueResult.QueueName]; !ok {
			err = info.AddQueue(queueResult.QueueName, queueResult.Users)
			if err != nil {
				return err
			}
		}

		for _, hostname := range queueResult.Hosts {
			err = info.AddDbHostQueue(hostname, queueResult.QueueName)
			if err != nil {
				return err
			}
		}
	}

	for _, queueResult := range queueInfoFromDb {
		hostsDict := map[string]string{}

		if _, ok := queueNamesfromLsfDict[queueResult.QueueName]; !ok {
			err = info.DelQueue(queueResult.QueueName)
			if err != nil {
				return err
			}

			lsfLog.Infof("del queue %s from db", queueResult.QueueName)
		}

		queueInfo, ok := queueNamesfromLsfDict[queueResult.QueueName]

		if ok {
			for _, hostname := range queueInfo.Hosts {
				hostsDict[hostname] = hostname
			}
		}

		for _, hostnameFromDb := range queueResult.Hosts {
			if _, ok := hostsDict[hostnameFromDb]; !ok {
				err := info.DelHostFromQueue(hostnameFromDb, queueResult.QueueName)
				if err != nil {
					return err
				}

				lsfLog.Infof("del host %s(%s) from db", hostnameFromDb, queueResult.QueueName)
			}
		}
	}

	queueInfo, err := info.GetQueueInfo()
	if err != nil {
		return err
	}

	info.QueueInfo = queueInfo
	return nil
}

func (info *LsfInfo) InitQueue() error {
	queueInfo, err := info.GetQueueInfo()
	if err != nil {
		return err
	}

	lsfLog.Debugf("InitQueue .... %v", queueInfo)

	if len(queueInfo) == 0 {
		// 数据库里没有数据，则直接使用bqueues的数据做数据源
		return info.SyncQueue()
	}

	info.QueueInfo = queueInfo
	return nil
}

func (info *LsfInfo) Init() error {
	hostsFromDb, err := info.GetHosts()

	if err != nil {
		return err
	}

	if len(hostsFromDb) > 0 {
		lsfLog.Infof("Init: Update hosts from db")
		info.UpdateWorkerNodes(hostsFromDb)
		return info.InitQueue()
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

	err = info.InitQueue()
	if err != nil {
		return err
	}

	lsfLog.Infof("Init: update hosts from bhost")
	return nil
}

func (info *LsfInfo) DelHostname(hostname string) error {

	err := info.DelHost(hostname)
	if err != nil {
		return err
	}

	hosts, err := info.GetHosts()
	if err != nil {
		return err
	}

	info.UpdateWorkerNodes(hosts)
	lsfLog.Infof("DelHost %s", hostname)
	return nil
}

func (info *LsfInfo) AddHostname(hostname string) error {

	if hostname == info.ClientNode {
		lsfLog.Infof("client node %v cannot add to hosts", hostname)
		return nil
	}

	// add host
	err := info.UpdateHosts([]string{hostname})
	if err != nil {
		return err
	}

	hosts, err := info.GetHosts()
	if err != nil {
		return err
	}

	info.UpdateWorkerNodes(hosts)
	lsfLog.Infof("AddHost %s", hostname)
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

	err = db.AutoMigrate(&LsfQueue{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&LsfQueueHost{})
	if err != nil {
		return nil, err
	}

	err = info.Init()
	if err != nil {
		return nil, err
	}

	return info, nil
}
