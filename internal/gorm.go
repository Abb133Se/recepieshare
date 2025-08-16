package internal

import (
	"log"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

func GetGormInstance() (*gorm.DB, error) {
	var err error
	once.Do(func() {
		dsn := "root:i1a3a7c7e@Se@tcp(127.0.0.1:3306)/recipes_db?charset=utf8mb4&parseTime=True&loc=Local"
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Println("failed to connect database:", err)
			return
		}

		sqlDB, _ := db.DB()
		sqlDB.SetMaxOpenConns(20)   // limit open connections
		sqlDB.SetMaxIdleConns(10)   // keep some idle
		sqlDB.SetConnMaxLifetime(0) // no forced recycling
	})
	return db, err
}
