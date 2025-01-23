package internal

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func GetGormInstance() (db *gorm.DB, err error) {
	dsn := "root:i1a3a7c7e@Se@tcp(127.0.0.1:3306)/recipes_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return
}
