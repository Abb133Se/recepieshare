package main

import (
	"fmt"
	"log"

	_ "github.com/Abb133Se/recepieshare/docs"
	"github.com/Abb133Se/recepieshare/internal"
	"github.com/Abb133Se/recepieshare/middleware"
	"github.com/Abb133Se/recepieshare/migrate"
	"github.com/Abb133Se/recepieshare/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(cors.Default())
	r.Use(middleware.SiteVisitMiddleware())

	// for future use the specific methods and config needed
	// r.Use(cors.New(cors.Config{
	//     AllowOrigins:     []string{"http://localhost:3000"}, // your frontend URL
	//     AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	//     AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
	//     ExposeHeaders:    []string{"Content-Length"},
	//     AllowCredentials: true,
	//     MaxAge: 12 * time.Hour,
	// }))

	internal.InitLocalStorage("uploads")

	routes.AddRoutes(r)

	db, err1 := internal.GetGormInstance()
	if err1 != nil {
		fmt.Println("not connected")
	}

	if err := migrate.AutoMigration(db); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	err := r.Run(":3000")
	if err != nil {
		log.Fatalf("impossible to start server: %s", err)
	}
}
