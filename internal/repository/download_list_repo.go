package repository

import (
	"main/internal/model"

	"gorm.io/gorm"
)

type DownloadListRepository interface {
	CrateTask(download *model.DownloadList) error
	UpdateTask(name string, download *model.DownloadList) error
}

type DownloadListRepo struct {
	db *gorm.DB
}

func NewDownloadListRepo(db *gorm.DB) *DownloadListRepo {
	return &DownloadListRepo{db: db}
}

func (d *DownloadListRepo) CrateTask(download *model.DownloadList) error {
	result := d.db.Create(&download)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *DownloadListRepo) UpdateTask(name string, download *model.DownloadList) error {
	result := d.db.Where("name = ?", name).Updates(download)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
