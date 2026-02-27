package model

import (
	"time"
)

type DownloadList struct {
	ID         int       `gorm:"column:id"`
	Name       string    `gorm:"column:name"`
	Type       string    `gorm:"column:type"`
	Path       string    `gorm:"column:path"`
	CreateTime time.Time `gorm:"column:create_time"`
	Status     int       `gorm:"column:status"`
}

func (d *DownloadList) TableName() string {
	return "download_list"
}
