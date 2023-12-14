package lsf

import (
	"gorm.io/gorm"
)

type LsfHost struct {
	gorm.Model
	HostName string
}

type LsfQueue struct {
	gorm.Model
	QueueName string
	QueueUser string
}

type LsfQueueHost struct {
	gorm.Model
	QueueName string
	HostName  string
}

func (info *LsfInfo) GetQueueInfo() ([]*LsfQueueInfo, error) {
	var (
		queues     []LsfQueue
		queueHosts []LsfQueueHost
		queueInfo  []*LsfQueueInfo
	)

	if err := info.Db.Find(&queues).Error; err != nil {
		return nil, err
	}

	if err := info.Db.Find(&queueHosts).Error; err != nil {
		return nil, err
	}

	for _, queue := range queues {
		info := LsfQueueInfo{
			QueueName: queue.QueueName,
			Users:     queue.QueueUser,
		}

		hosts := []string{}

		for _, queueHost := range queueHosts {
			if queueHost.QueueName == queue.QueueName {
				hosts = append(hosts, queueHost.HostName)
			}
		}

		info.Hosts = hosts
		queueInfo = append(queueInfo, &info)
	}

	return queueInfo, nil
}

func (info *LsfInfo) AddDbHostQueue(hostname string, queuename string) error {
	var (
		lsfQueueHost []LsfQueueHost
	)

	if hostname == info.ClientNode {
		lsfLog.Infof("clientNode %s cannot add to queue", hostname)
		return nil
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
	return nil
}

func (info *LsfInfo) DelHostFromQueue(hostname string, queuename string) error {
	if err := info.Db.Where("host_name = ? AND queue_name = ?", hostname, queuename).Delete(&LsfQueueHost{}).Error; err != nil {
		return err
	}

	lsfLog.Infof("Delete %s from queue %s", hostname, queuename)
	return nil
}

func (info *LsfInfo) AddQueue(queuename string, queueuser string) error {
	var (
		lsfQueue []LsfQueue
	)

	if result := info.Db.Where(&LsfQueue{QueueName: queuename}).Find(&lsfQueue); result.RowsAffected == 0 {
		if err := info.Db.Create(&LsfQueue{
			QueueName: queuename,
			QueueUser: queueuser,
		}).Error; err != nil {
			lsfLog.Errorf("Create Queue info failed. Queue name: % , error: %v", queuename, err)
			return err
		}
	} else {
		lsfLog.Infof("Queue %s already exist", queuename)
	}

	lsfLog.Info("AddQueue to db: %s", queuename)
	return nil
}

func (info *LsfInfo) DelQueue(queuename string) error {
	if err := info.Db.Where("queue_name = ?", queuename).Delete(&LsfQueue{}).Error; err != nil {
		return err
	}

	lsfLog.Infof("Delete queue %s", queuename)
	return nil
}

func (info *LsfInfo) UpdateQueueInfo(queues []*LsfQueueInfo) error {
	var (
		lsfQueue    []LsfQueue
		lsfQueuHost []LsfQueueHost
	)

	for _, queueInfo := range queues {
		if result := info.Db.Where(&LsfQueue{QueueName: queueInfo.QueueName}).Find(&lsfQueue); result.RowsAffected == 0 {
			if err := info.Db.Create(&LsfQueue{
				QueueName: queueInfo.QueueName,
				QueueUser: queueInfo.Users,
			}).Error; err != nil {
				lsfLog.Errorf(
					"Create Queue info failed, Queue name: %v, error: %v",
					queueInfo.QueueName,
					err,
				)
				return err
			}
		}

		for _, hostname := range queueInfo.Hosts {
			if result := info.Db.Where(&LsfQueueHost{QueueName: queueInfo.QueueName, HostName: hostname}).Find(&lsfQueuHost); result.RowsAffected == 0 {
				if err := info.Db.Create(&LsfQueueHost{
					QueueName: queueInfo.QueueName,
					HostName:  hostname,
				}).Error; err != nil {
					lsfLog.Errorf(
						"Create Queue Host info failed, Queue name: %v, host name: %v, error: %v",
						queueInfo.QueueName,
						hostname,
						err,
					)
					return err
				}
			}
		}
	}

	return nil
}

func (info *LsfInfo) GetHosts() ([]string, error) {
	var (
		hosts     []LsfHost
		hostnames []string
	)
	if result := info.Db.Find(&hosts); result.Error != nil {
		return hostnames, result.Error
	}

	for _, hostInfo := range hosts {
		hostnames = append(hostnames, hostInfo.HostName)
	}

	return hostnames, nil
}

func (info *LsfInfo) DelHost(hostname string) error {
	if err := info.Db.Where("host_name = ?", hostname).Delete(&LsfHost{}).Error; err != nil {
		return err
	}

	return nil
}

func (info *LsfInfo) UpdateHosts(hosts []string) error {
	var lsfhosts []LsfHost

	for _, hostname := range hosts {
		if result := info.Db.Where(&LsfHost{HostName: hostname}).Find(&lsfhosts); result.RowsAffected == 0 {
			if err := info.Db.Create(&LsfHost{
				HostName: hostname,
			}).Error; err != nil {
				lsfLog.Errorf("UpdateHosts failed, failed to add %v, error: %v", hostname, err)
				return err
			}
		}
	}

	return nil
}
