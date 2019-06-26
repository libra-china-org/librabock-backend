package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
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

func (database DataBaseAdapter) GetDB() (*gorm.DB) {
	db, err := gorm.Open("mysql", database.GetURL())

	if err != nil {
		panic("failed to connect database")
	}

	return db
}

func (database DataBaseAdapter) migration() {
	db := database.GetDB()
	defer db.Close()

	db.AutoMigrate(&BlockModel{})
}

func (database DataBaseAdapter) GetLatestVersion() uint64 {
	db := database.GetDB()
	defer db.Close()
	result := BlockModel{}
	db.Order("version desc").First(&result)

	return result.Version
}

func (database DataBaseAdapter) SaveBlock(model BlockModel) {
	db := database.GetDB()
	defer db.Close()

	db.Create(&model)
}