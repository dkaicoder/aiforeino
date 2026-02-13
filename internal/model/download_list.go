package model

import (
	"log"
	"main/internal/database"
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

func (d *DownloadList) CrateTask() {
	result := database.MysqlDb.Create(&d)
	if result.Error != nil {
		log.Println("crateTask create error", result.Error)
	}
}

func (d *DownloadList) UpdateTask() {
	result := database.MysqlDb.Where("name = ?", d.Name).Updates(d)
	if result.Error != nil {
		log.Println("crateTask create error", result.Error)
	}
}
