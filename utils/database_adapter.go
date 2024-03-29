package utils

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"io.librablock.go/models"
)

type DataBaseAdapter struct {
	url string
}

func NewDataBaseAdapter(url string) DataBaseAdapter {
	d := DataBaseAdapter{}
	d.url = url
	return d
}

func (database DataBaseAdapter) GetURL() string {
	return fmt.Sprintf("%s?charset=utf8mb4&parseTime=True&loc=Local", database.url)
}

func (database DataBaseAdapter) GetDB() *gorm.DB {
	db, err := gorm.Open("mysql", database.GetURL())

	if err != nil {
		panic("failed to connect database")
	}

	return db
}

func (database DataBaseAdapter) Migration() {
	db := database.GetDB()
	defer db.Close()

	db.AutoMigrate(&models.BlockModel{})
}

func (database DataBaseAdapter) GetLatestVersion() uint64 {
	db := database.GetDB()
	defer db.Close()
	result := models.BlockModel{}
	db.Order("version desc").First(&result)

	return result.Version
}

func (database DataBaseAdapter) GetVersion(id uint64) models.BlockModel {
	db := database.GetDB()
	defer db.Close()

	var result models.BlockModel
	db.Where("version = ?", id).First(&result)

	return result
}

func (database DataBaseAdapter) GetVersions(offset int, limit int) []models.BlockModel {
	db := database.GetDB()
	defer db.Close()
	if limit > 50 {
		limit = 50
	}

	var blocks []models.BlockModel
	db.Order("version desc").Offset(offset).Limit(limit).Find(&blocks)

	return blocks
}

func (database DataBaseAdapter) GetVersionsRefAddress(address string, offset int, limit int) []models.BlockModel {
	db := database.GetDB()
	defer db.Close()
	if limit > 50 {
		limit = 50
	}

	var blocks []models.BlockModel
	db.Where("source = ?", address).Or("destination = ?", address).Order("version desc").Offset(offset).Limit(limit).Find(&blocks)

	return blocks
}

func (database DataBaseAdapter) SaveBlock(model models.BlockModel) {
	db := database.GetDB()
	defer db.Close()

	db.Create(&model)
}
