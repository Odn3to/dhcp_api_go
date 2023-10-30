package router

import (
	"github.com/gofiber/fiber/v2"

	"dhcp-api-go/controllers"
)

func Register(app *fiber.App) {

	dhcp := app.Group("/dhcp")

	// Adicione o middleware às rotas
	dhcp.Use(controllers.TokenValidationMiddleware)

	dhcp.Get("/status", controllers.Status)
	dhcp.Get("/initConf", controllers.GetConfigDHCP)
	dhcp.Get("/interfaces", controllers.GetInterfaces)
	dhcp.Post("/conf", controllers.ConfiguradorDHCP)
	dhcp.Get("/data", controllers.GetCsv)
}