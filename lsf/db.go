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

func (info *LsfInfo) UpdateHosts(hosts []string) error {
	return info.Db.Transaction(func(tx *gorm.DB) error {
		for _, hostname := range hosts {
			if result := tx.Where(LsfHost{HostName: hostname}); result.RowsAffected == 0 {
				if err := tx.Create(&LsfHost{
					HostName: hostname,
				}).Error; err != nil {
					lsfLog.Errorf("UpdateHosts failed, failed to add %v, error: %v", hostname, err)
					return err
				}
			}
		}

		// do commit
		return nil
	})
}
