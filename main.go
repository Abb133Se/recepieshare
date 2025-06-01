package main

import (
	"fmt"
	"log"

	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/migrate"
	"github.com/Abb133Se/recepieshare/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	routes.AddRoutes(r)

	db, err1 := internal.GetGormInstance()
	if err1 != nil {
		fmt.Println("not connected")
	}

	if err := migrate.AutoMigration(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("Database migrated successfully.")

	err := r.Run(":3000")
	if err != nil {
		log.Fatalf("impossible to start server: %s", err)
	}
}
