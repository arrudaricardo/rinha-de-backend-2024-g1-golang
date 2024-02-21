package main

import (
	"github.com/gofiber/fiber/v3"
	"rinha-de-backend-2024-q1-golang/api/route"
	"rinha-de-backend-2024-q1-golang/database"
)

func init() {
	database.ConnectDB()
}

func main() {
  defer database.ConnPool.Close()
	app := fiber.New()
	route.SetupRoutes(app)
	app.Listen(":8080")
}
