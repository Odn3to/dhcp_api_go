package database

import (
    "gorm.io/gorm"
    "gorm.io/driver/sqlite"
)
var (
	DB  *gorm.DB
)

func ConectaNoBD() {
	db, err := gorm.Open(sqlite.Open("config_dhcp.db"), &gorm.Config{})
  	if err != nil {
    	panic("failed to connect database")
  	}

	DB = db
}

func GetDataBase() *gorm.DB {
	return DB
}