package lsf

import (
	"gorm.io/gorm"
)

type LsfHost struct {
	gorm.Model
	HostName string
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
			lsfLog.Debugf("Check hostname %v exist: %v", hostname, result)
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
