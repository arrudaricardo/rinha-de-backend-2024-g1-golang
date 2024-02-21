package route

import "github.com/gofiber/fiber/v3"
import "rinha-de-backend-2024-q1-golang/api/controller"


func SetupRoutes(app *fiber.App) {

	app.Post("/clientes/:id/transacoes", controller.PostTransacoes)
	app.Get("/clientes/:id/extrato", controller.GetExtrato)

}
