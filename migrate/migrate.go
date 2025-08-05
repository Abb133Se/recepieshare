package migrate

import (
	"log"

	"gorm.io/gorm"

	"github.com/Abb133Se/recepieshare/model"
)

func AutoMigration(db *gorm.DB) error {
	log.Println("Running auto migration...")
	return db.AutoMigrate(
		&model.User{},
		&model.Recipe{},
		&model.Ingredient{},
		&model.Comment{},
		&model.Favorite{},
		&model.Rating{},
		&model.Tag{},
		&model.Category{},
	)
}
