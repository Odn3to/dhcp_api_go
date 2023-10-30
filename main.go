package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"dhcp-api-go/router"
	"dhcp-api-go/database"
)

// @title DHCP - API
// @version 1.0
// @description Gerencia e lida com as escritas para o serviço Kea
// @host 172.23.58.10:8005
// @BasePath /dhcp
// @schemes http https
func main() {

	database.ConectaNoBD()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
	}))

	router.Register(app)
	log.Fatal(app.Listen(fmt.Sprintf(":%s", "8005")))
}