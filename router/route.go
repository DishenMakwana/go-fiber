package router

import (
	"github.com/dishenmakwana/go-fiber/handler"
	"github.com/dishenmakwana/go-fiber/middleware"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	//Auth
	auth := api.Group("/auth")
	auth.Post("/login", middleware.ValidateUserLogin, handler.Login)

	v1 := api.Group("/user", middleware.Protected())

	// routes
	v1.Get("/", handler.GetAllUsers)
	v1.Get("/:id", handler.GetSingleUser)
	v1.Post("/", handler.CreateUser)
	v1.Put("/:id", handler.UpdateUser)
	v1.Delete("/:id", handler.DeleteUserByID)
}
